package main

import (
	"go.uber.org/zap"
	"os/exec"
	"syscall"
	"time"
)

func restartService(c Controller, component string) error {

	// edit this part

	err := syscall.Chroot("/host")
	if err != nil {
		c.log.Error("Failed to create chroot on /host",
			zap.String("component", component),
			zap.Error(err))
		return err
	}

	cmd := exec.Command("/bin/bash", "-c", "systemctl restart "+component)

	err = cmd.Run()

	time.Sleep(10 * time.Second)
	if err != nil {
		c.log.Error("Failed to restart service",
			zap.String("component", component),
			zap.Error(err))
		return err
	}

	return nil
}

func Restart(c Controller) error {

	err := restartService(c, "etcd")
	if err != nil {
		c.log.Error("Error restarting etcd: ",
			zap.Error(err))
	}

	err = restartService(c, "kubelet")
	if err != nil {
		c.log.Error("Error restarting kubelet: ",
			zap.Error(err))
	}

	return nil
}
