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

	// TODO:
	//  replace zap with zeroLog

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

	// TODO:
	err = Prerequisites()
	if err != nil {

	}

	availableVersion, err := availableVersions(c.log)

	if len(availableVersion) == 0 {
		c.log.Fatal("no available versions for minor upgrade",
			zap.Error(err))
	}

	if err != nil {
		c.log.Error("failed to fetch minor versions for kubernetes version upgrade",
			zap.Error(err))
	}

	// todo: include in the testing
	current := true
	latest := false
	version := "1.26.5-1.1"

	if current {
		//version = kubernetesVersion
	}

	if latest {
	}

	// TODO:
	//  if available version fails or works,
	//  check if that matches with upgradePlane
	//  if the latest is selected

	kubeadmUpgrade, err := k8sComponentsUpgrade(c.log, "kubeadm", version)
	if err != nil {
		c.log.Error("failed to get upgrade kubeadm",
			zap.Error(err))
	}

	if !kubeadmUpgrade {

	}

	plan, err := upgradePlan(c.log)
	if err != nil {
		c.log.Error("failed to get upgrade plan",
			zap.Error(err))
	}

	diff, err := Diff(c.log, plan)
	if err != nil {
		c.log.Error("failed to get diff",
			zap.Error(err))
	}

	c.log.Info("diff for upgrade plan",
		zap.String("diff", diff))

	k8sUpgrade, err := clusterUpgrade(c.log, version)
	if err != nil {
		c.log.Error("failed to get upgrade plan",
			zap.Error(err))
	}

	if !k8sUpgrade {

	}

	kubeletUpgrade, err := k8sComponentsUpgrade(c.log, "kubelet", version)
	if err != nil {
		c.log.Error("failed to get upgrade plan",
			zap.Error(err))
	}

	if kubeletUpgrade {
		// TODO: sudo systemctl daemon-reload
		//  sudo systemctl restart kubelet
	}

	kubectlUpgrade, err := k8sComponentsUpgrade(c.log, "kubectl", version)
	if err != nil {
		c.log.Error("failed to get upgrade plan",
			zap.Error(err))
	}

	if kubectlUpgrade {
	}

	// TODO:
	//  --certificate-renewal=false

	GrpcClient(c.log)
}
