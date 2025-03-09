package main

import (
	"bytes"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"google.golang.org/grpc"
	"os/exec"
	"strings"
)

func availableVersions(c Controller, conn *grpc.ClientConn) (bool, error) {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to create chroot on /host")
	}

	var updateCommand string

	if c.distro == "rhel" {
		updateCommand = "yum update -y"
	} else if c.distro == "ubuntu" {
		updateCommand = "apt update -y"
	}

	// TODO: re test on ubuntu and debian
	cmd := exec.Command("/bin/bash", "-c", updateCommand)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	err = cmd.Run()
	if err != nil {
		c.log.Error().Err(err).
			Msg("failed to fetch repository updates")
	}

	c.log.Debug().
		Str("stdout", stdoutBuf.String()).
		Str("stderr", stderrBuf.String()).
		Msg("successfully fetched repository updates")

	var packageSearch string
	if c.distro == "rhel" {
		packageSearch = "yum list --showduplicates *kubeadm --disableexcludes=kubernetes --quiet | grep .x86_64 | awk '{ print $2 }'"
	} else if c.distro == "ubuntu" {
		packageSearch = "apt-cache madison kubeadm | awk '{ print $3 }'"
	}

	cmd = exec.Command("/bin/bash", "-c", packageSearch)
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	err = cmd.Run()

	if err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to list available versions")
		// TODO: refactor this to send the error : return false, "", err
	}

	c.log.Debug().
		Str("stdout", stdoutBuf.String()).
		Str("stderr", stderrBuf.String()).
		Msg("successfully fetched repository updates")

	availableVersionSlice := strings.Split(stdoutBuf.String(), "\n")

	if len(availableVersionSlice) == 0 {
		c.log.Error().Err(err).
			Msg("no available versions found for minor upgrade")
		// todo: return false, "", err
	}

	if err = changedRoot(); err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to exit from the updated root")
	}

	conn.ResetConnectBackoff()
	connection := pb.NewMinorUpdateClient(conn)

	rpc, err := connection.UpdateVersionSelection(c.ctx,
		&pb.AvailableVersions{
			Version: availableVersionSlice,
			Err:     "",
		})

	if err != nil {
		c.log.Error().Err(err).
			Msg("could not send status update")
		return false, err
	}

	c.log.Info().
		Bool("retry current step", rpc.GetRetryCurrentStep()).
		Bool("response received", rpc.GetResponseReceived()).
		Msg("available versions")

	if rpc.GetRetryCurrentStep() {
		return false, nil
	}

	return true, nil
}
