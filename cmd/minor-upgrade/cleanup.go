package main

import (
	"github.com/shishir9159/kapetanios/utils"
	"go.uber.org/zap"
	"os"
	"os/exec"
)

// TODO: remove the previous version packages
//  make sure to check if the previous version exists
//  and current kubernetes version

// TODO: check kubelet status and view the service logs with journalctl -xeu kubelet

func restartKubelet(c Controller) error {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Fatal("Failed to create chroot on /host",
			zap.Error(err))
	}

	// refactor into two
	cmd := exec.Command("/bin/bash", "-c", "systemctl daemon-reload && systemctl restart kubelet")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err = cmd.Run()

	if err != nil {
		c.log.Error("Failed to restart kubelet",
			zap.Error(err))
		return err
	}

	if err = changedRoot(); err != nil {
		c.log.Fatal("Failed to exit from the updated root",
			zap.Error(err))
	}

	return err
}
