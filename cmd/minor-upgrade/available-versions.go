package main

import (
	"bytes"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"os/exec"
	"strings"
)

func availableVersions(c Controller, conn *grpc.ClientConn) (bool, string, error) {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Error("Failed to create chroot on /host",
			zap.Error(err))
		return false, "", err
	}

	cmd := exec.Command("/bin/bash", "-c", "apt update -y")

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	err = cmd.Run()

	if err != nil {
		c.log.Error("Failed to update vm",
			zap.Error(err))
	}

	// TODO: detect redhat, and run: yum list --showduplicates kubeadm --disableexcludes=kubernetes

	cmd = exec.Command("/bin/bash", "-c", "apt-cache madison kubeadm | awk '{ print $3 }'")

	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	//wait.PollUntilContextTimeout()

	err = cmd.Run()
	// output delimiter is " | "
	// extract second and the third column

	if err != nil {
		c.log.Error("Failed to list available versions",
			zap.Error(err))
		// TODO: refactor this to send the error : return false, "", err
	}

	outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	c.log.Info("outString and errString",
		zap.String("outStr", outStr),
		zap.String("errStr", errStr))

	availableVersionSlice := strings.Split(outStr, "\n")

	if len(availableVersionSlice) == 0 {
		c.log.Error("no available versions for minor upgrade",
			zap.Error(err))
		// todo: panic??? return false, "", err
	}

	// TODO:
	//  sort them based on the delimiter "." and "-' + give a score by adding them ups with positional
	//  values

	if err = changedRoot(); err != nil {
		c.log.Fatal("Failed to exit from the updated root",
			zap.Error(err))
	}

	conn.ResetConnectBackoff()
	connection := pb.NewMinorUpgradeClient(conn)

	rpc, err := connection.UpgradeVersionSelection(c.ctx,
		&pb.AvailableVersions{
			Version: availableVersionSlice,
			Err:     "",
		})

	if err != nil {
		c.log.Error("could not send status update: ", zap.Error(err))
		return false, "", err
	}

	c.log.Info("available versions",
		zap.Bool("proceed to next step", rpc.GetProceedNextStep()),
		zap.Bool("terminate application", rpc.GetTerminateApplication()),
		zap.Bool("certificate renewal", rpc.GetCertificateRenewal()),
		zap.String("fetch the version to upgrade", rpc.GetVersion()))

	return rpc.GetProceedNextStep(), rpc.GetVersion(), nil
}
