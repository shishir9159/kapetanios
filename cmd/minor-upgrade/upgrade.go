package main

import (
	"github.com/shishir9159/kapetanios/utils"
	"go.uber.org/zap"
	"os/exec"
	"time"
)

func Upgrade(log *zap.Logger, version string) (bool, error) {

	// The NodeRestriction admission plugin
	// prevents the kubelet from setting or modifying labels
	// with a node-restriction.kubernetes.io/ prefix.
	// But, the plugin is needed to be inserted to kubelet configuration
	// and will bring extra complexity.

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		log.Error("Failed to create chroot on /host",
			zap.Error(err))
		return false, err
	}

	// Copy Recursively

	cmd := exec.Command("/bin/bash", "-c", "apt-mark unhold kubeadm && apt-get update && apt-get install -y kubeadm="+version+" && apt-mark hold kubeadm")
	err = cmd.Run()

	time.Sleep(4 * time.Second)
	if err != nil {
		log.Error("Failed to install kubeadm",
			zap.Error(err))
		return false, err
	}

	if err = changedRoot(); err != nil {
		log.Fatal("Failed to exit from the updated root",
			zap.Error(err))
	}

	// TODO:
	//  Check for kubernetes repo if no version is found
	//  disableexclude

	return false, nil
}
