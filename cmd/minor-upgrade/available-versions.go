package main

import (
	"github.com/shishir9159/kapetanios/utils"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"time"
)

func availableVersions(log *zap.Logger) ([]string, error) {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		log.Error("Failed to create chroot on /host",
			zap.Error(err))
		return nil, err
	}

	cmd := exec.Command("/bin/bash", "-c", "apt update -y")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()

	time.Sleep(4 * time.Second)
	if err != nil {
		log.Error("Failed to update vm",
			zap.Error(err))
		return nil, err
	}

	// TODO: detect redhat, and run: yum list --showduplicates kubeadm --disableexcludes=kubernetes

	cmd.Args = append([]string{cmd.Path}, "-c apt-cache madison kubeadm | awk '{ print $3 }'")
	//wait.PollUntilContextTimeout()

	err = cmd.Run()
	// output delimiter is " | "
	// extract second and the third column

	// TODO: return output

	if err != nil {
		log.Error("Failed to list available versions",
			zap.Error(err))
		return nil, err
	}

	if err = changedRoot(); err != nil {
		log.Fatal("Failed to exit from the updated root",
			zap.Error(err))
	}

	// TODO:
	//  Check for kubernetes repo if no version is found
	//  disableexclude

	return nil, nil
}
