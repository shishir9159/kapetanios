package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	//"github.com/rs/zerolog/log"
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

	err = Prerequisites()
	if err != nil {

	}

	availableVersion, err := AvailableVersions()

	if len(availableVersion) == 0 {
		c.log.Fatal("no available versions for minor upgrade",
			zap.Error(err))
	}

	latest := false
	version := ""

	if latest {

	}

	// TODO:
	//  if available version fails or works,
	//  check if that matches with upgradePlane
	//  if the latest is selected

	upgradePlan, err := UpgradePlan(c.log, version)
	if err != nil {
		c.log.Error("failed to get upgrade plan",
			zap.Error(err))
	}

	if err != nil {
		c.log.Error("failed to fetch minor versions for kubernetes version upgrade",
			zap.Error(err))
	}

	diff, err := Diff(c.log, upgradePlan)
	if err != nil {
		c.log.Error("failed to get diff",
			zap.Error(err))
	}

	err = Upgade()
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
