package main

import (
	"bytes"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"os/exec"
)

func compatibility(c Controller, version string, conn *grpc.ClientConn) (bool, string, error) {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Error("Failed to create chroot on /host",
			zap.Error(err))
		return false, "", err
	}

	// upgrade plan to list available upgrade options
	// --config is not necessary as it is saved in the cm

	//kubeadm upgrade diff to see the changes
	cmd := exec.Command("/bin/bash", "-c", "kubeadm upgrade diff "+version+" --config /etc/kubernetes/kubeadm/kubeadm-config.yaml")

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	// TODO: CONTAINER-RUNTIME compatibility

	err = cmd.Run()

	if err != nil {
		c.log.Info("cmd.Run() failed with",
			zap.Error(err))
		// TODO: return the error to the server : return false, "", err
	}

	// TODO: OS compatibility
	diff, _ := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())

	if err = changedRoot(); err != nil {
		c.log.Fatal("Failed to exit from the updated root",
			zap.Error(err))
	}

	conn.ResetConnectBackoff()
	connection := pb.NewMinorUpgradeClient(conn)

	rpc, err := connection.ClusterCompatibility(c.ctx,
		&pb.UpgradeCompatibility{
			OsCompatibility: true,
			Diff:            diff,
			Err:             "", // TODO: check for nil pointer and return
		})

	if err != nil {
		c.log.Error("could not send status update: ", zap.Error(err))
		return false, "", err
	}

	c.log.Info("upgrade diff",
		zap.Bool("proceed to the next step", rpc.GetProceedNextStep()),
		zap.Bool("terminate application", rpc.GetTerminateApplication()))

	return rpc.GetProceedNextStep(), diff, nil
}
