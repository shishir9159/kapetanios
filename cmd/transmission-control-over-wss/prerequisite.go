package main

import (
	"context"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// if the labels exists
// if the labels match with the expectations
// number of control-plane pods and nodes running
// if the nodes match with the expectations

func Prerequisites(namespace string) {
	// if cm shows updated nodes to a certain value
	// and desired kubernetesVersion exists on the cm,
	// then, call the minor upgrade

	// TODO: throw error no master nodes found
	logger := zap.Must(zap.NewProduction())
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			logger.Info("error syncing logger before application terminates",
				zap.Error(err))
		}
	}(logger)

	// TODO:
	//  refactor
	client, err := orchestration.NewClient()

	// TODO: add namespace in the controller itself
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

	configMapName := "kapetanios"

	configMap, er := c.client.Clientset().CoreV1().ConfigMaps(namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
	if er != nil {
		c.log.Error("error fetching the configMap",
			zap.Error(er))
	}

	c.log.Info("after fetching configmap")

	// TODO: to be foolproof check if the number of nodes the same
	//  if that is the case, the first node consideration need to be taken

	targetedVersion := configMap.Data["TARGETED_K8S_VERSION"]
	nodesToBeUpgraded := configMap.Data["NODES_TO_BE_UPGRADED"]
	// todo: upgradedNodes := configMap.Data["UPGRADED_NODES"]

	if targetedVersion != "" && nodesToBeUpgraded != "" {
		LastDance(c, nodesToBeUpgraded, namespace)
		configMap.Data["TARGETED_K8S_VERSION"] = ""
		configMap.Data["NODES_TO_BE_UPGRADED"] = ""
		// todo: upgradedNodes := configMap.Data["UPGRADED_NODES"]

		_, er = c.client.Clientset().CoreV1().ConfigMaps(namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
		if er != nil {
			c.log.Error("error updating configMap",
				zap.Error(er))
		}
	}
}
