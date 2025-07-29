package main

import (
	"context"
	"kapetanios/internal/orchestration"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
)

type Controller struct {
	mu        sync.Mutex
	client    *orchestration.Client
	namespace string
	ctx       context.Context
	log       *zap.Logger
}

func Cert(namespace string) {

	logger := zap.Must(zap.NewProduction())
	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {
			logger.Info("error syncing logger before application terminates",
				zap.Error(er))
		}
	}(logger)

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

	// TODO:
	//  Controller Definition need to be moved with the
	//  initial Setup and making sure there exists only one
	//  :refactor Controller name

	InitialSetup(c)

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
		//	return erring or call grpc
	}

	ch := make(chan *grpc.Server, 1)
	go CertGrpc(c.log, ch)

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

	// TODO: refactor this part to orchestrator by decomposing
	for index, node := range nodes.Items {

		// namespace should only be included after the consideration for the existing
		// service account, cluster role binding
		descriptor := renewalMinionManager.MinionBlueprint("laplaciandemon/certs-renewal", roleName, node.Name)

		// TODO:
		//  optional backoff if the ip or hostname doesn't match

		value := "False"
		if node.Labels["assigned-node-role-etcd.kubernetes.io"] == "true" {
			value = "True"
		}

		externalETCDNodeEnv := corev1.EnvVar{
			Name:  "ETCD_NODE",
			Value: value,
		}

		descriptor.Spec.Containers[0].Env = append(descriptor.Spec.Containers[0].Env, externalETCDNodeEnv)

		// kubectl get event --namespace default --field-selector involvedObject.name=minions
		// how many pods to monitor should be sent to the orchestration too
		minion, er := c.client.Clientset().CoreV1().Pods(namespace).Create(context.Background(), descriptor, metav1.CreateOptions{})
		if er != nil {
			c.log.Error("Cert Renewal pod creation failed: ",
				zap.Int("index", index),
				zap.Error(er))

			// todo:
			//  graceful stop
			//return er
			return
		}

		c.log.Info("Cert Renewal pod created",
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
						//  possibly send the event with error about the pod
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

		er = RestartByLabel(c, map[string]string{"tier": "control-plane"}, node.Name)
		if er != nil {
			c.log.Error("error restarting pods for certificate renewal",
				zap.Error(er))

			// TODO: abort
			//  and send error message
			break
		}
	}

	(<-ch).GracefulStop()

	err = RestartRemainingComponents(c, "default")
	if err != nil {
		c.log.Error("error restarting renewal components",
			zap.Error(err))
	}
}
