package main

import (
	"bytes"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"go.uber.org/zap"
	"io"
	"os"
	"os/exec"
)

func compatibility(c Controller, version string, connection pb.MinorUpgradeClient) (string, error) {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Error("Failed to create chroot on /host",
			zap.Error(err))
		return "", err
	}

	//var stdoutBuf, stderrBuf bytes.Buffer

	// upgrade plan to list available upgrade options
	// --config is not necessary as it is saved in the cm

	var stdoutBuf, stderrBuf bytes.Buffer

	//kubeadm upgrade diff to see the changes
	cmd := exec.Command("/bin/bash", "-c", "kubeadm upgrade diff "+version+" --config /etc/kubernetes/kubeadm-config.yaml")

	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err = cmd.Run()

	if err != nil {
		c.log.Info("cmd.Run() failed with",
			zap.Error(err))
	}

	// TODO: OS compatibility
	diff, _ := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())

	//var dryRunStdoutBuf, dryRunStderrBuf bytes.Buffer
	//
	////kubeadm upgrade diff to see the changes
	//cmd = exec.Command("/bin/bash", "-c", "kubeadm upgrade diff "+version+" --config /etc/kubernetes/kubeadm-config.yaml")
	//
	//cmd.Stdout = io.MultiWriter(os.Stdout, &dryRunStdoutBuf)
	//cmd.Stderr = io.MultiWriter(os.Stderr, &dryRunStderrBuf)
	//
	//err = cmd.Run()
	//
	//if err != nil {
	//	c.log.Info("cmd.Run() failed with",
	//		zap.Error(err))
	//}

	if err = changedRoot(); err != nil {
		c.log.Fatal("Failed to exit from the updated root",
			zap.Error(err))
	}

	rpc, err := connection.ClusterCompatibility(c.ctx,
		&pb.UpgradeCompatibility{
			OsCompatibility: true,
			Diff:            diff,
			Err:             "",
		})

	//	TODO: kubeadm upgrade node (name) [version] --dry-run

	if err != nil {
		c.log.Error("could not send status update: ", zap.Error(err))
		return "", err
	}

	c.log.Info("Backup Status",
		zap.Bool("next step", rpc.GetProceedNextStep()),
		zap.Bool("retry", rpc.GetSkipRetryCurrentStep()),
		zap.Bool("terminate application", rpc.GetTerminateApplication()))

	return diff, nil
}
