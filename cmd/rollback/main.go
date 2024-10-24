package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
)

type Controller struct {
	ctx context.Context
	log *zap.Logger
}

func main() {

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Println(err)
	}

	//zap.ReplaceGlobals(logger)

	// replace zap with zeroLog

	c := Controller{
		ctx: context.Background(),
		log: logger,
	}

	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {
			c.log.Error("failed to close logger",
				zap.Error(er))
		}
	}(logger)

	// err = PrerequisitesForRollback(c.log)
	// if err != nil {
	//  	c.log.Error("the cluster didn't meet the condition for rollback",
	//	 	zap.Error(err))
	// }

	//	step 1. replacing the last generated configs with
	//  the latest backups would cause the cert renewal rollback

	err = Rollback()
	if err != nil {
		c.log.Error("failed to renew certificates and kubeConfigs",
			zap.Error(err))
	}

	GrpcClient(c.log)

	//step 3. Restarting pods to work with the updated certificates
	err = Restart(c)
	if err != nil {
		c.log.Error("failed to restart kubernetes components after certificate renewal",
			zap.Error(err))
	}
}
