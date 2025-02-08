package main

import (
	"bytes"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"google.golang.org/grpc"
	"os"
	"os/exec"
)

// TODO: remove the previous version packages
//  make sure to check if the previous version exists
//  and current kubernetes version

// TODO: check kubelet status and view the service logs with journalctl -xeu kubelet

func restartComponent(c Controller, component string, conn *grpc.ClientConn) (bool, error) {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to create chroot on /host")
	}

	// refactor into two
	cmd := exec.Command("/bin/bash", "-c", "systemctl daemon-reload && systemctl restart "+component)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	err = cmd.Run()

	if err != nil {
		c.log.Error().Err(err).
			Msg("failed to restart kubelet")
	}

	if err = changedRoot(); err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to exit from the updated root")
	}

	conn.ResetConnectBackoff()
	connection := pb.NewMinorUpgradeClient(conn)

	rpc, err := connection.ClusterComponentRestart(c.ctx,
		&pb.ComponentRestartStatus{
			ComponentRestartSuccess: true,
			Component:               component,
			Log:                     "",
			Err:                     "",
		})

	if err != nil {
		c.log.Error().Err(err).
			Msg("could not send status update")
		// TODO: retry for communication
		return false, err
	}

	c.log.Info().
		Bool("next step", rpc.GetProceedNextStep()).
		Bool("terminate application", rpc.GetTerminateApplication()).
		Msg("backup status")

	if rpc.GetTerminateApplication() {
		os.Exit(0)
	}

	return rpc.GetProceedNextStep(), nil
}
