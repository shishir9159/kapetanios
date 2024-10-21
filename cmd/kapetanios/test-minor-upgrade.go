package main

import (
	"context"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"time"
)

func TestMinorUpgrade(namespace string) {

	logger, err := zap.NewProduction()
	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {
			logger.Fatal("error syncing logger before application terminates", zap.Error(err))
		}
	}(logger)

	client, err := orchestration.NewClient()

	c := Controller{
		client: client,
		ctx:    context.Background(),
		log:    logger,
	}

	if err != nil {
		c.log.Error("error creating kubernetes client",
			zap.Error(err))
	}

	roleName := "node-2-for-test"

	// TODO: add the label to node 2
	matchLabels := map[string]string{"assigned-node-role.kubernetes.io": roleName}

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
		c.log.Error("no nodes found",
			//	return err or call grpc
			zap.Error(err))
	}

	for index, node := range nodes.Items {

		// namespace should only be included after the consideration for the existing
		// service account, cluster role binding
		descriptor := renewalMinionManager.MinionBlueprint("quay.io/klovercloud/minor-upgrade", roleName, node.Name)

		descriptor.Spec.Tolerations = []corev1.Toleration{
			{
				Key:               "minor-upgrade-running",
				Operator:          "Equal",
				Value:             "processing",
				Effect:            "NoSchedule",
				TolerationSeconds: &[]int64{3}[0],
			},
		}

		err = drain(c, node)
		if err != nil {
			c.log.Error("failed to drain node",
				zap.String("node name:", node.Name),
				zap.Error(err))
		}

		addTaint(&node)

		err = uncordon(&node)
		if err != nil {
			c.log.Error("failed to uncordon node",
				zap.String("node name:", node.Name),
				zap.Error(err))
		}

		// TODO:
		//  if the pod doesn't schedule, check for taint
		//  check for all pod related event with informer

		// TODO: monitor the node status with watch

		// TODO: monitor the pod restart after upgrade
		//  All containers are restarted after upgrade, because the container spec hash value is changed
		//		just monitor the NODES before creating minion, no need to restart
		//RestartByLabel(c, map[string]string{"tier": "control-plane"}, node.Name)

		minion, er := c.client.Clientset().CoreV1().Pods(namespace).Create(context.Background(), descriptor, metav1.CreateOptions{})
		if er != nil {
			c.log.Error("minor upgrade pod creation failed: ",
				zap.Int("index", index),
				zap.Error(er))

			//return er
			return
		}

		c.log.Info("minor upgrade pod created",
			zap.Int("index", index),
			zap.String("pod_name", minion.Name))

		// todo: wait for request for restart from the minions
		time.Sleep(5 * time.Second)

		removeTaint(&node)
	}
}

//  ToDo:
//   update the information in the configMaps
//   specially about k8s version
