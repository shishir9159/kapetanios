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
	//if cm shows updated nodes to a certain value
	//	 and desired kubernetesVersion exists on the cm,
	//   then, call the minor upgrade

	// TODO: throw error no master nodes found

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

	configMap, er := c.client.Clientset().CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if er != nil {
		c.log.Error("error fetching the configMap",
			zap.Error(er))
	}

	targetedVersion := configMap.Data["TargetedVersion"]
	nodesToBeUpgraded := configMap.Data["NodesToBeUpgraded"]

	if targetedVersion != "" && nodesToBeUpgraded != "" {
		LastDance(namespace)
		configMap.Data["TargetedVersion"] = ""
		configMap.Data["NodesToBeUpgraded"] = ""
	}
}
