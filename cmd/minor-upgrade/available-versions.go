package main

import (
	"bytes"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"go.uber.org/zap"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

func availableVersions(c Controller, connection pb.MinorUpgradeClient) ([]string, error) {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Error("Failed to create chroot on /host",
			zap.Error(err))
		return nil, err
	}

	cmd := exec.Command("/bin/bash", "-c", "apt update -y")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()

	time.Sleep(4 * time.Second)
	if err != nil {
		c.log.Error("Failed to update vm",
			zap.Error(err))
		return nil, err
	}

	// TODO: detect redhat, and run: yum list --showduplicates kubeadm --disableexcludes=kubernetes

	cmd = exec.Command("/bin/bash", "-c", "apt-cache madison kubeadm | awk '{ print $3 }'")

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	//wait.PollUntilContextTimeout()

	err = cmd.Run()
	// output delimiter is " | "
	// extract second and the third column

	// TODO: return output

	if err != nil {
		c.log.Error("Failed to list available versions",
			zap.Error(err))
		return nil, err
	}

	outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	c.log.Info("outString and errString",
		zap.String("outStr", outStr),
		zap.String("errStr", errStr))

	availableVersionSlice := strings.Split(outStr, "\n")

	// TODO:
	//  sort them based on the delimiter "." and "-'

	if err = changedRoot(); err != nil {
		c.log.Fatal("Failed to exit from the updated root",
			zap.Error(err))
	}

	rpc, err := connection.UpgradeVersionSelection(c.ctx,
		&pb.AvailableVersions{
			Version: nil,
			Err:     "",
		})

	if err != nil {
		c.log.Error("could not send status update: ", zap.Error(err))
	}

	c.log.Info("available versions",
		zap.Bool("next step", rpc.GetProceedNextStep()),
		zap.Bool("retry", rpc.GetSkipRetryCurrentStep()),
		zap.Bool("terminate application", rpc.GetTerminateApplication()),
		zap.Bool("certificate renewal", rpc.GetCertificateRenewal()),
		zap.String("version", rpc.Version))

	return availableVersionSlice, nil
}
