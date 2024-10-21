package main

import (
	"context"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"time"
)

func Drain(node string) error {

	return nil
}

func Taint(node string) error {

	return nil
}

// be careful about the different version across
// the nodes

// TODO: for testing purposes, try the current version

func MinorUpgrade() {

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

	matchLabels := map[string]string{}
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
		//	return err or call grpc
	}

	for index, node := range nodes.Items {

		// namespace should only be included after the consideration for the existing
		// service account, cluster role binding
		descriptor := renewalMinionManager.MinionBlueprint("quay.io/klovercloud/certs-renewal", roleName, node.Name)

		// kubectl get event --namespace default --field-selector involvedObject.name=minions
		// how many pods this logic need to be in the orchestration too
		minion, er := c.client.Clientset().CoreV1().Pods(namespace).Create(context.Background(), descriptor, metav1.CreateOptions{})
		if er != nil {
			c.log.Error("Cert Renewal pod creation failed: ",
				zap.Int("index", index),
				zap.Error(er))

			//return er
			return
		}

		c.log.Info("Cert Renewal pod created",
			zap.Int("index", index),
			zap.String("pod_name", minion.Name))

		// todo: wait for request for restart from the minions
		time.Sleep(5 * time.Second)

		er = RestartByLabel(c, map[string]string{"tier": "control-plane"}, node.Name)
		if er != nil {
			c.log.Error("error restarting pods for certificate renewal",
				zap.Error(er))

			//retry logic
			//return er
			break
		}
	}

	err = RestartRemainingComponents(c, "default")
	if err != nil {
		c.log.Error("error restarting renewal components", zap.Error(err))
	}

	// TODO: drain add node selector or something,
	//   add the same thing on the necessary pods(except for ds)

	//  TODO: after the pod is scheduled
	//   must first drain the node
	//   if failed, must be tainted again to
	//   schedule nodes

	err = Drain("")
	if err != nil {
		c.log.Error("failed to drain node",
			zap.String("node name:", ""),
			zap.Error(err))
	}

	err = Taint("")
	if err != nil {
		c.log.Error("failed to taint node",
			zap.String("node name:", ""),
			zap.Error(err))
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
