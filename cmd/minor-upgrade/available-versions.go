package main

import (
	"bytes"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"google.golang.org/grpc"
	"os/exec"
	"strings"
)

func availableVersions(c Controller, conn *grpc.ClientConn) (bool, string, error) {

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

	var repoSearch string
	if c.distro == "rhel" {
		repoSearch = "yum list --showduplicates *kubeadm --disableexcludes=kubernetes | grep .x86_64 | awk '{ print $2 }'"
	} else if c.distro == "ubuntu" {
		repoSearch = "apt-cache madison kubeadm | awk '{ print $3 }'"
	}

	cmd = exec.Command("/bin/bash", "-c", repoSearch)
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	err = cmd.Run()
	// output delimiter is " | "
	// extract second and the third column

	if err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to list available versions")
		// TODO: refactor this to send the error : return false, "", err
	}

	c.log.Debug().
		Str("stdout", stdoutBuf.String()).
		Str("stderr", stderrBuf.String()).
		Msg("successfully fetched repository updates")

	//outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	//c.log.Info().
	//	Str("out", outStr).
	//	Str("err", errStr).
	//	Msg("outString & errString")

	availableVersionSlice := strings.Split(stdoutBuf.String(), "\n")

	if len(availableVersionSlice) == 0 {
		c.log.Error().Err(err).
			Msg("no available versions found for minor upgrade")
		// todo: return false, "", err
	}

	// TODO:
	//  sort them based on the delimiter "." and "-' + give a score by adding them ups with positional
	//  values

	if err = changedRoot(); err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to exit from the updated root")
	}

	conn.ResetConnectBackoff()
	connection := pb.NewMinorUpgradeClient(conn)

	rpc, err := connection.UpgradeVersionSelection(c.ctx,
		&pb.AvailableVersions{
			Version: availableVersionSlice,
			Err:     "",
		})

	if err != nil {
		c.log.Error().Err(err).
			Msg("could not send status update")
		// TODO: retry for communication
		return false, "", err
	}

	c.log.Info().
		Bool("proceed to next step", rpc.GetProceedNextStep()).
		Bool("terminate application", rpc.GetTerminateApplication()).
		Bool("certificate renewal", rpc.GetCertificateRenewal()).
		Str("fetch the version to upgrade", rpc.GetVersion()).
		Msg("available versions")

	// TODO: VERSION COALESCING - FROM NUMERIC VALUES

	return rpc.GetProceedNextStep(), rpc.GetVersion(), nil
}
