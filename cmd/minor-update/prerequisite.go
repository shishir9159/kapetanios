package main

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/rs/zerolog"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"google.golang.org/grpc"
	"os"
	"os/exec"
	"strings"
)

func getDistro(c Controller) (string, string, error) {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to create chroot on /host")
	}

	fi, err := os.Open("/etc/os-release")
	if err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to open /etc/os-release")
	}

	var distro string
	var distroVersion string

	scanner := bufio.NewScanner(fi)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			idValue := strings.TrimPrefix(line, "PRETTY_NAME=")
			idValue = strings.Trim(idValue, `"`)
			distro = idValue
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			versionValue := strings.TrimPrefix(line, "VERSION_ID=")
			versionValue = strings.Trim(versionValue, `"`)
			distroVersion = versionValue
		}

		c.log.Debug().
			Str("distro version", distroVersion).
			Str("distro name", distro).
			Msg("~distro~")
	}

	if er := scanner.Err(); er != nil {
		c.log.Error().Err(er).
			Msg("failed to read the file /etc/os-release")
	}

	if err = changedRoot(); err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to exit from the updated root")
	}

	return distro, distroVersion, err
}

func getRepoInfo(distro string, log *zerolog.Logger) (string, string, string, error) {

	if distro == "" {
		return "", "", "", errors.New("distro is required")
	}

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		log.Fatal().Err(err).
			Msg("failed to create chroot on /host")
	}

	var updateCommand string

	if strings.Contains(distro, "Red Hat Enterprise Linux") {
		updateCommand = "yum update -y"
	} else if distro == "ubuntu" {
		updateCommand = "apt update -y"
	}

	// TODO: re test on ubuntu and debian
	cmd := exec.Command("/bin/bash", "-c", updateCommand)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	err = cmd.Run()
	if err != nil {
		log.Error().Err(err).
			Msg("failed to fetch repository updates")
	}

	log.Debug().
		Str("stdout", stdoutBuf.String()).
		Str("stderr", stderrBuf.String()).
		Msg("successfully fetched repository updates")

	stdoutBuf.Reset()
	stderrBuf.Reset()

	var currentVersionCMD string
	if strings.Contains(distro, "Red Hat Enterprise Linux") {
		currentVersionCMD = "rpm -qi kubectl | awk '{ print $3 }' | tail -n +2 | head -n 1"
	}

	cmd = exec.Command("/bin/bash", "-c", currentVersionCMD)
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	err = cmd.Run()

	if err != nil {
		log.Fatal().Err(err).
			Msg("failed to list available versions")
		// TODO: refactor this to send the error : return false, "", err
	}

	currentVersion, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	log.Info().
		Str("stdOut", currentVersion).
		Str("stdErr", errStr).
		Msg("successfully fetched repository updates")

	stdoutBuf.Reset()
	stderrBuf.Reset()

	var repoNameCmd string
	if strings.Contains(distro, "Red Hat Enterprise Linux") {
		repoNameCmd = "yum info kubectl --installed --quiet | grep 'From repo' | awk '{ print $4 }'"
	}
	// TODO:
	//  UBUNTU + Debian
	//else if distro == "ubuntu" {
	//	repoNameCmd = "apt-cache madison kubeadm | awk '{ print $3 }'"
	//}

	cmd = exec.Command("/bin/bash", "-c", repoNameCmd)
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	err = cmd.Run()

	if err != nil {
		log.Fatal().Err(err).
			Msg("failed to list available versions")
		// TODO: refactor this to send the error : return false, "", err
	}

	repoName, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	log.Info().
		Str("stdOut", repoName).
		Str("stdErr", errStr).
		Msg("successfully fetched repository updates")

	stdoutBuf.Reset()
	stderrBuf.Reset()

	var repoBaseURLCMD string
	if strings.Contains(distro, "Red Hat Enterprise Linux") {
		repoBaseURLCMD = "yum repolist --repo " + repoName + " -v enabled --quiet | grep Repo-baseurl | awk '{ print $3}' | tail -n 1"
	}

	cmd = exec.Command("/bin/bash", "-c", repoBaseURLCMD)
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	err = cmd.Run()

	if err != nil {
		// should be fatal
		log.Error().Err(err).
			Msg("failed to list available versions")
	}

	repoBaseURL, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	log.Info().
		Str("stdOut", repoBaseURL).
		Str("stdErr", errStr).
		Msg("successfully fetched repository updates")

	if err = changedRoot(); err != nil {
		log.Fatal().Err(err).
			Msg("failed to exit from the updated root")
	}

	return currentVersion, repoName, repoBaseURL, nil
}

func Prerequisites(c Controller, conn *grpc.ClientConn) (bool, string, error) {

	// TODO: how to know the current node is etcd with clientSet?
	//  	- etcd cluster from the cm
	//		- how to get the hostname and ip address
	//  check if external etcd running if it's an etcd node

	distro, distroVersion, err := getDistro(c)

	currentVersion, repoName, repoBaseURL, err := getRepoInfo(distro, &c.log)

	etcdNode := os.Getenv("ETCD_NODE")
	if etcdNode == "false" {
		return false, distro, errors.New("ETCD_NODE environment variable set false")
	} else if etcdNode == "true" {
		return false, distro, errors.New("ETCD_NODE environment variable set to be True")
	}

	connection := pb.NewMinorUpdateClient(conn)

	rpc, err := connection.PrerequisitesCheck(c.ctx,
		&pb.PrerequisitesMinorUpdate{
			Distro:         distro,
			DistroVersion:  distroVersion,
			RepoName:       repoName,
			RepoBaseURL:    repoBaseURL,
			CurrentVersion: strings.TrimSpace(currentVersion),
			Log:            "",
			Err:            "",
		})

	if err != nil {
		c.log.Error().Err(err).
			Msg("could not send status update")
		return false, distro, err
	}

	c.log.Info().
		Bool("next step", rpc.GetProceedNextStep()).
		Bool("retry current step", rpc.GetRetryCurrentStep()).
		Bool("terminate application", rpc.GetTerminateApplication()).
		Msg("prerequisite step response")

	if rpc.GetTerminateApplication() {
		os.Exit(0)
	} else if rpc.GetRetryCurrentStep() {
		return false, "", nil
	}

	return rpc.GetProceedNextStep(), distro, nil
}
