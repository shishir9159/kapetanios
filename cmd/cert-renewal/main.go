package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
)

var (
	backupCount            = 7
	overRideUserKubeConfig = 0
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

	// replace zap with zerolog

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

	err = PrerequisitesForCertRenewal(c.log)
	if err != nil {
		c.log.Error("failed to get cluster health status",
			zap.Error(err))
	}

	//	step 1. Backup directories
	err = BackupCertificatesKubeConfigs(c, backupCount)
	if err != nil {
		c.log.Error("failed to backup certificates and kubeConfigs",
			zap.Error(err))
	}

	//	step 2. Kubeadm certs renew all
	err = Renew(c)
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

	if overRideUserKubeConfig != 0 {
		err = Copy("/etc/kubernetes/admin.conf", "/root/.kube/config")
		if err != nil {
			c.log.Error("failed to pass kubernetes admin privilege to the root user",
				zap.Error(err))
		}
	}
}
