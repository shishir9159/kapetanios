package main

import (
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"go.uber.org/zap"
	"os/exec"
	"time"
)

// Refactor to new internal library Restart
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

func Restart(c Controller, connection pb.RollbackClient) error {

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

	rpc, err := connection.Restarts(c.ctx,
		&pb.RollbackRestartStatus{
			EtcdRestart:    false,
			KubeletRestart: false,
			EtcdError:      "",
			KubeletError:   "",
			Err:            "",
		})

	if err != nil {
		c.log.Error("could not send status update: ", zap.Error(err))
	}

	c.log.Info("Status Update",
		zap.Bool("response received", rpc.GetResponseReceived()))

	return err
}
