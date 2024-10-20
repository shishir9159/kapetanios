package main

import (
	"context"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"go.uber.org/zap"
)

// be careful about the different version across
// the nodes

func MinorUpgrade() {

	logger, err := zap.NewProduction()

	if err != nil {
		panic(err)
	}

	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {

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

	// TODO: drain add node selector or something,
	//   add the same thing on the necessary pods(except for ds)
	//

	//  TODO: after the pod is scheduled
	//   must first drain the node
	//   if failed, must be tainted again to
	//   schedule nodes

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
