package main

import (
	"context"
	"fmt"
	"kapetanios/internal/orchestration"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
)

func Rollback(namespace string) {

	logger := zap.Must(zap.NewProduction())
	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {
			logger.Info("error syncing logger before application terminates",
				zap.Error(er))
		}
	}(logger)

	// refactor
	client, err := orchestration.NewClient()

	c := Controller{
		client: client,
		ctx:    context.Background(),
		log:    logger,
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err != nil {
		c.log.Error("error creating kubernetes client",
			zap.Error(err))
	}

	//InitialSetup(c)
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

	if len(nodes.Items) == 0 {
		c.log.Error("no master nodes found",
			zap.Error(err))
	}

	ch := make(chan *grpc.Server, 1)
	go RollbackGrpc(c.log, ch)

	// TODO: refactor this part to orchestrator

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

	for index, node := range nodes.Items {

		// namespace should only be included after the consideration for the existing
		// service account, cluster role binding
		descriptor := renewalMinionManager.MinionBlueprint("laplaciandemon/certs-rollback", roleName, node.Name)

		// TODO: if etcd exists in the node (check by IP)
		//  add environment value to restart etcd or not

		// kubectl get event --namespace default --field-selector involvedObject.name=minions
		// how many pods this logic needs to be in the orchestration too
		minion, er := c.client.Clientset().CoreV1().Pods(namespace).Create(context.Background(), descriptor, metav1.CreateOptions{})
		if er != nil {
			c.log.Error("rollback pod creation failed: ",
				zap.Int("index", index),
				zap.Error(er))

			//return er
			return
		}

		c.log.Info("rollback pod created",
			zap.Int("index", index),
			zap.String("pod_name", minion.Name))

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

		err = RestartByLabel(c, map[string]string{"tier": "control-plane"}, node.Name)
		if err != nil {
			c.log.Error("error restarting pods for certificate renewal",
				zap.Error(er))

			//retry logic
			//return er
			//push error
			break
		}
	}

	(<-ch).GracefulStop()

	c.log.Debug("entering restartRemainingComponent")

	err = RestartRemainingComponents(c, "default")
	if err != nil {
		c.log.Error("error restarting renewal components",
			zap.Error(err))
	}

	c.log.Info("entering grpc server")
}
