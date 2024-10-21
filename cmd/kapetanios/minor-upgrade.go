package main

import (
	"context"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func Drain(node string) error {

	return nil
}

func removeTaint(node *corev1.Node) {

	if node.Spec.Taints != nil {
		//	TODO: what if there are multiple taints

	}

	taint := []corev1.Taint{
		{
			Key:    "minor-upgrade-running",
			Value:  "processing",
			Effect: "NoSchedule",
		},
	}

	node.Spec.Taints = taint
}

func taint(node *corev1.Node) {

	if node.Spec.Taints != nil {
		//	TODO: what if there are multiple taints

	}

	node.Spec.Taints = []corev1.Taint{
		{
			Key:    "minor-upgrade-running",
			Value:  "processing",
			Effect: "NoSchedule",
		},
	}
}

// be careful about the different version across
// the nodes

// TODO: for testing purposes, try the current version

// TODO: run only in the second node

func MinorUpgrade(namespace string) {

	logger, err := zap.NewProduction()
	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {
			logger.Fatal("error syncing logger before application terminates", zap.Error(err))
		}
	}(logger)

	// TODO:
	//  refactor
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

	roleName := "minor-upgrade"

	renewalMinionManager := orchestration.NewMinions(client)

	nodes, err := c.client.Clientset().CoreV1().Nodes().List(context.Background(), metav1.ListOptions{LabelSelector: ""})

	// TODO: sort with control-plane role, error no master nodes found

	if err != nil {
		c.log.Error("error listing nodes",
			zap.Error(err))
	}

	if len(nodes.Items) == 0 {
		c.log.Error("no nodes found",
			zap.Error(err))
		//	return err or call grpc
	}

	for index, node := range nodes.Items {

		// namespace should only be included after the consideration for the existing
		// service account, cluster role binding
		descriptor := renewalMinionManager.MinionBlueprint("quay.io/klovercloud/minor-upgrade", roleName, node.Name)

		// TODO: instead of pod monitoring for creation, monitor for successful restarts
		//  er = RestartByLabel(c, map[string]string{"tier": "control-plane"}, node.Name)
		//  if er != nil {
		//  	c.log.Error("error restarting pods for certificate renewal",
		//	    	zap.Error(er))
		//	  //retry logic
		//	 // return er
		//	 break
		//  }

		// TODO: drain add node selector or something,
		//   add the same thing on the necessary pods(except for ds)

		//  TODO: after the pod is scheduled
		//   must first drain the node
		//   if failed, must be tainted again to
		//   schedule nodes

		descriptor.Spec.Tolerations = []corev1.Toleration{
			{
				Key:               "minor-upgrade-running",
				Operator:          "",
				Value:             "processing",
				Effect:            "",
				TolerationSeconds: &[]int64{3}[0],
			},
		}

		err = Drain("")
		if err != nil {
			c.log.Error("failed to drain node",
				zap.String("node name:", ""),
				zap.Error(err))
		}

		taint(&node)

		// TODO:
		//  if the pod doesn't schedule, check for taint
		//  check for all pod related event with informer

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
	}

	err = RestartRemainingComponents(c, "default")
	if err != nil {
		c.log.Error("error restarting renewal components", zap.Error(err))
	}

	//step 3. no need to Restart pods to adopt with the upgrade

	//

	// TODO: monitor the pod restart after upgrade
	//  All containers are restarted after upgrade, because the container spec hash value is changed
	//		just monitor the NODES before creating minion, no need to restart
	//RestartByLabel(c, map[string]string{"tier": "control-plane"}, node.Name)
}

//  ToDo:
//   update the information in the configMaps
//   specially about k8s version
