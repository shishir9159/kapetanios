package main

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	. "kapetanios/internal/model"
	"kapetanios/internal/orchestration"
	"time"
)

func Expiration(namespace string) {

	logger := zap.Must(zap.NewProduction())
	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {
			logger.Info("error syncing logger before application terminates",
				zap.Error(er))
		}
	}(logger)

	// TODO:
	//  refactor
	client, err := orchestration.NewClient()

	c := Controller{
		client:    client,
		ctx:       context.Background(),
		log:       logger,
		namespace: namespace,
	}

	if err != nil {
		c.log.Error("error creating kubernetes client",
			zap.Error(err))
	}

	roleName := "certs"
	matchLabels := map[string]string{"assigned-node-role-certs.kubernetes.io": roleName}

	labelSelector := metav1.LabelSelector{MatchLabels: matchLabels}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	renewalMinionManager := orchestration.NewMinions(client)

	nodes, err := c.client.Clientset().CoreV1().Nodes().List(context.Background(), listOptions)
	if err != nil {
		c.log.Error("error listing nodes",
			zap.Error(err))
	}

	if nodes == nil || len(nodes.Items) == 0 {
		c.log.Error("no master nodes found",
			zap.Error(err))
		// TODO:
		//  push to error message
		return
	}

	nodesCertsExpiration := make([]NodeCertsExpiration, 0)
	ch := make(chan *grpc.Server, 1)
	go ExpirationGrpc(c.log, ch, &nodesCertsExpiration)

	labelSelector = metav1.LabelSelector{
		MatchLabels: map[string]string{"app": "certs"},
	}

	listOptions = metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	watcher, err := c.client.Clientset().CoreV1().Pods(namespace).Watch(context.TODO(), metav1.ListOptions{
		LabelSelector: listOptions.LabelSelector,
	})

	defer watcher.Stop()

	if err != nil {
		c.log.Error("watcher for certs expiration pod failed: ",
			zap.Error(err))
	}

	// TODO: refactor this part inside the orchestrator

	for index, node := range nodes.Items {
		descriptor := renewalMinionManager.MinionBlueprint("laplaciandemon/certs-expiration", roleName, node.Name)
		minion, er := c.client.Clientset().CoreV1().Pods(namespace).Create(context.Background(), descriptor, metav1.CreateOptions{})
		if er != nil {
			c.log.Error("Cert Expiration pod creation failed: ",
				zap.Int("index", index),
				zap.Error(er))

			// todo:
			//  graceful stop

			return // er // defer gracefulStop
		}

		c.log.Info("Cert Expiration pod created",
			zap.Int("index", index),
			zap.String("pod_name", minion.Name))

		if watcher == nil {
			c.log.Info("watcher is nil, sleeping for 10 seconds")
			time.Sleep(10 * time.Second)
		} else {
		outerLoop:
			for event := range watcher.ResultChan() {
				pod, ok := event.Object.(*corev1.Pod)
				if !ok {
					c.log.Error("watcher returned unexpected type",
						zap.Reflect("event", event),
						zap.Reflect("object", event.Object))
					continue
				}

				// TODO:
				//  evicted

				switch event.Type {
				case watch.Modified:
					if pod.Status.Phase == corev1.PodSucceeded {
						c.log.Info("certs expiration pod has completed successfully!",
							zap.String("pod name", pod.Name),
							zap.String("pod namespace", pod.Namespace))
						break outerLoop
					} else if pod.Status.Phase == corev1.PodFailed {
						c.log.Info("certs expiration pod has failed!",
							zap.String("pod name", pod.Name),
							zap.String("pod namespace", pod.Namespace))
						// TODO:
						//  handle pod failure
						break outerLoop

					}
				case watch.Deleted:
					c.log.Info("certs expiration pod was deleted!",
						zap.String("pod name", pod.Name),
						zap.String("pod namespace", pod.Namespace))
					break outerLoop
				}
			}
		}
	}

	(<-ch).GracefulStop()

	for _, nodeCertsExpiration := range nodesCertsExpiration {

		c.log.Info("certs expiration info for node:",
			zap.String("node name", nodeCertsExpiration.NodeName),
			zap.Strings("missing certificate list", nodeCertsExpiration.MissingCertificate))

		for _, certificate := range nodeCertsExpiration.CertificateExpiration {
			c.log.Info("Certs Expiration Info",
				zap.String("name", certificate.Name),
				zap.String("expires", certificate.Expires),
				zap.String("residual time", certificate.ResidualTime),
				zap.String("certificate authority", certificate.CertificateAuthority),
				zap.String("externally managed", certificate.ExternallyManaged))
		}

		for _, caAuthority := range nodeCertsExpiration.CertificatesAuthority {
			c.log.Info("Certs Authority Info",
				zap.String("name", caAuthority.Name),
				zap.String("expires", caAuthority.Expires),
				zap.String("residual time", caAuthority.ResidualTime),
				zap.String("externally managed", caAuthority.ExternallyManaged))

		}
	}
}
