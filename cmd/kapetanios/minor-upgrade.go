package main

import (
	"context"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func drain(node corev1.Node) error {

	return nil
}

func removeTaint(node *corev1.Node) {

	taints := node.Spec.Taints

	if len(taints) == 0 {
		return
	}

	taintToRemove := corev1.Taint{
		Key:    "minor-upgrade-running",
		Value:  "processing",
		Effect: corev1.TaintEffectNoSchedule,
	}

	newTaints := []corev1.Taint{taintToRemove}

	for _, taint := range taints {
		if taint.MatchTaint(&taintToRemove) {
			continue
		}

		newTaints = append(newTaints, taint)
	}

	node.Spec.Taints = newTaints
}

func addTaint(node *corev1.Node) {

	taints := node.Spec.Taints

	// TODO: declare as a struct maybe?
	taintToAdd := corev1.Taint{
		Key:    "minor-upgrade-running",
		Value:  "processing",
		Effect: corev1.TaintEffectNoSchedule,
	}

	newTaints := []corev1.Taint{taintToAdd}

	if len(taints) != 0 {
		for _, taint := range taints {
			if taint.MatchTaint(&taintToAdd) {
				return
			}

			newTaints = append(newTaints, taint)
		}

		return
	}

	node.Spec.Taints = newTaints
}

func uncordon(node *corev1.Node) error {

	return nil
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
			//	return err or call grpc
			zap.Error(err))
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
				Operator:          "Equal",
				Value:             "processing",
				Effect:            "NoSchedule",
				TolerationSeconds: &[]int64{3}[0],
			},
		}

		err = drain(node)
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
