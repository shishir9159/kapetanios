package main

import (
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"go.uber.org/zap"
	"os/exec"
	"time"
)

func restartService(c Controller, component string) error {

	cmd := exec.Command("/bin/bash", "-c", "systemctl restart "+component)

	err := cmd.Run()

	time.Sleep(4 * time.Second)
	if err != nil {
		c.log.Error("Failed to restart service",
			zap.String("component", component),
			zap.Error(err))
		return err
	}

	return nil
}

func Restart(c Controller, connection pb.RenewalClient) error {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Fatal("Failed to create chroot on /host",
			zap.Error(err))
	}

	err = restartService(c, "etcd")
	if err != nil {
		c.log.Error("Error restarting etcd: ",
			zap.Error(err))
	}

	err = restartService(c, "kubelet")
	if err != nil {
		c.log.Error("Error restarting kubelet: ",
			zap.Error(err))
	}

	if err = changedRoot(); err != nil {
		c.log.Fatal("Failed to exit from the updated root",
			zap.Error(err))
	}

	rpc, err := connection.RestartUpdate(c.ctx,
		&pb.RenewalStatus{
			RenewalSuccess:          false,
			KubeConfigBackup:        false,
			FileChecklistValidation: false,
			Err:                     "",
		})

	if err != nil {
		c.log.Error("could not send status update: ", zap.Error(err))
	}

	c.log.Info("Backup Status",
		zap.Bool("next step", rpc.GetProceedNextStep()),
		zap.Bool("retry", rpc.GetSkipRetryCurrentStep()),
		zap.Bool("terminate application", rpc.GetTerminateApplication()))

	return err
}
