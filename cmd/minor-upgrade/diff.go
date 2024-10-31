package main

import (
	"bytes"
	"github.com/shishir9159/kapetanios/utils"
	"go.uber.org/zap"
	"io"
	"os"
	"os/exec"
)

func Diff(log *zap.Logger, version string) (string, error) {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		log.Error("Failed to create chroot on /host",
			zap.Error(err))
		return "", err
	}

	//var stdoutBuf, stderrBuf bytes.Buffer

	// upgrade plan to list available upgrade options
	// --config is not necessary as it is saved in the cm

	var stdoutBuf, stderrBuf bytes.Buffer

	//kubeadm upgrade diff to see the changes
	cmd := exec.Command("/bin/bash", "-c", "kubeadm upgrade diff "+version+" --config /etc/kubernetes/kubeadm-config.yaml")
	//cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	//cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err = cmd.Run()

	// TODO:
	//  try combinedOutput and revert back later
	//out, err := cmd.CombinedOutput()
	if err != nil {
		log.Info("cmd.Run() failed with",
			zap.Error(err))
	}

	// list all the node names
	// and sort the list from the smallest worker node by resources
	// if it works successfully in the worker nodes, work on the master nodes
	//	--certificate-renewal=false
	//	kubeadm upgrade node (name) [version] --dry-run

	diff, _ := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())

	if err = changedRoot(); err != nil {
		log.Fatal("Failed to exit from the updated root",
			zap.Error(err))
	}

	return diff, nil
}
