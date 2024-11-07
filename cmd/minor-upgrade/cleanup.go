package main

import (
	"bytes"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"os/exec"
)

// TODO: remove the previous version packages
//  make sure to check if the previous version exists
//  and current kubernetes version

// TODO: check kubelet status and view the service logs with journalctl -xeu kubelet

func restartComponent(c Controller, component string, conn *grpc.ClientConn) (bool, error) {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Fatal("Failed to create chroot on /host",
			zap.Error(err))
		return false, err
	}

	// refactor into two
	cmd := exec.Command("/bin/bash", "-c", "systemctl daemon-reload && systemctl restart "+component)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	err = cmd.Run()

	if err != nil {
		c.log.Error("Failed to restart kubelet",
			zap.Error(err))
	}

	if err = changedRoot(); err != nil {
		c.log.Fatal("Failed to exit from the updated root",
			zap.Error(err))
	}

	conn.ResetConnectBackoff()
	connection := pb.NewMinorUpgradeClient(conn)

	rpc, err := connection.ClusterComponentRestart(c.ctx,
		&pb.ComponentRestartStatus{
			ComponentRestartSuccess: false,
			Component:               component,
			Log:                     "",
			Err:                     "",
		})

	if err != nil {
		c.log.Error("could not send status update: ", zap.Error(err))
		return false, err
	}

	c.log.Info("Backup Status",
		zap.Bool("next step", rpc.GetProceedNextStep()),
		zap.Bool("terminate application", rpc.GetTerminateApplication()))

	return rpc.GetProceedNextStep(), nil
}
