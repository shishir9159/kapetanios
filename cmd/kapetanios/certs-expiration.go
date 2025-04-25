package main

import (
	"context"
	"fmt"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	"time"
)

type NodeCertsExpiration struct {
	nodeName              string
	missingCertificate    []string
	certificateExpiration []CertsExpiration
	certificatesAuthority []CertsAuthority
}

type CertsExpiration struct {
	Name                 string
	Expires              string
	ResidualTime         string
	CertificateAuthority string
	ExternallyManaged    string
}

type CertsAuthority struct {
	Name              string
	Expires           string
	ResidualTime      string
	ExternallyManaged string
}

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

		watcher, err := c.client.Clientset().CoreV1().Pods(namespace).Watch(context.TODO(), metav1.ListOptions{
			LabelSelector: listOptions.LabelSelector,
		})

		defer watcher.Stop()

		if err != nil {
			c.log.Error("watcher for certs expiration pod failed: ",
				zap.Error(err))
		}

		if watcher == nil {
			c.log.Error("could not create watcher",
				zap.Error(fmt.Errorf("watcher is nil")))
			//	TODO: avoid outLoop and pass after a timer
		}

		if watcher == nil {
			c.log.Info("watcher is nil, sleeping for 10 seconds")
			time.Sleep(10 * time.Second)
		} else {
		outerLoop:
			// TODO: handle possible empty channel return
			for event := range watcher.ResultChan() {
				pod, ok := event.Object.(*corev1.Pod)
				if !ok {
					c.log.Error("watcher returned unexpected type",
						zap.Reflect("event", event),
						zap.Reflect("object", event.Object))
					continue
				}

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

		c.log.Info("Cert Expiration pod created",
			zap.Int("index", index),
			zap.String("pod_name", minion.Name))
		// todo: wait for request for restart from the minions
	}

	(<-ch).GracefulStop()

	for _, nodeCertsExpiration := range nodesCertsExpiration {

		c.log.Info("certs expiration info for node:",
			zap.String("node name", nodeCertsExpiration.nodeName),
			zap.Strings("missing certificate list", nodeCertsExpiration.missingCertificate))

		for _, certificate := range nodeCertsExpiration.certificateExpiration {
			c.log.Info("Certs Expiration Info",
				zap.String("name", certificate.Name),
				zap.String("expires", certificate.Expires),
				zap.String("residual time", certificate.ResidualTime),
				zap.String("certificate authority", certificate.CertificateAuthority),
				zap.String("externally managed", certificate.ExternallyManaged))
		}

		for _, caAuthority := range nodeCertsExpiration.certificatesAuthority {
			c.log.Info("Certs Authority Info",
				zap.String("name", caAuthority.Name),
				zap.String("expires", caAuthority.Expires),
				zap.String("residual time", caAuthority.ResidualTime),
				zap.String("externally managed", caAuthority.ExternallyManaged))
		}
	}
}
