package main

import (
	"bufio"
	"errors"
	pb "kapetanios/proto"
	"kapetanios/utils"
	"os"
	"strings"

	"google.golang.org/grpc"
)

func getDistro(c Controller) (string, string) {

	changedRoot, err := utils.ChangeRoot("/host")
	if err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to create chroot on /host")
	}

	errorString := ""

	fi, err := os.Open("/etc/os-release")
	if err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to open /etc/os-release")

		errorString = "error: " + err.Error() + "\n"
	}

	var distroPrettyName string

	scanner := bufio.NewScanner(fi)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			prettyName := strings.TrimPrefix(line, "PRETTY_NAME=")
			prettyName = strings.Trim(prettyName, `"`)
			distroPrettyName = prettyName
			c.log.Info().
				Str("distro pretty name", distroPrettyName).
				Msg("~distro~")
		}
	}

	if er := scanner.Err(); er != nil {
		c.log.Error().Err(er).
			Msg("failed to read the file /etc/os-release")

		errorString = "============================"
		errorString = "error: " + er.Error() + "\n"
	}

	if err = changedRoot(); err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to exit from the updated root")
	}

	return distroPrettyName, errorString
}

func Prerequisites(c Controller, conn *grpc.ClientConn) (bool, string, error) {

	// TODO: how to know the current node is etcd with clientSet?
	//  	- etcd cluster from the cm
	//		- how to get the hostname and ip address
	//  check if external etcd running if it's an etcd node

	distro, errorString := getDistro(c)
	if errorString != "" {
		c.log.Error().
			Str("error", errorString).
			Msg("failed to get distro pretty name")
	}

	etcdNode := os.Getenv("ETCD_NODE")
	if etcdNode == "false" {
		return false, distro, errors.New("ETCD_NODE environment variable set false")
	} else if etcdNode == "true" {
		return false, distro, errors.New("ETCD_NODE environment variable set to be True")
	}

	connection := pb.NewMinorUpgradeClient(conn)

	rpc, err := connection.ClusterHealthChecking(c.ctx,
		&pb.PrerequisitesMinorUpgrade{
			// TODO:
			EtcdStatus:          true,
			StorageAvailability: 50,
			Log:                 "",
			Err:                 errorString,
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
