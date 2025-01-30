package main

import (
	"bufio"
	"errors"
	pb "github.com/shishir9159/kapetanios/proto"
	"github.com/shishir9159/kapetanios/utils"
	"google.golang.org/grpc"
	"os"
	"strings"
)

func getDistro(c Controller) (string, error) {

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

	//defer func() {
	//	if er := fi.Close(); er != nil {
	//		c.log.Fatal().Err(er).
	//			Msg("failed to close /etc/os-release")
	//	}
	//}()
	//
	//// try double buffering the chunks
	//buf := make([]byte, 1024)
	//for {
	//	// read a chunk
	//	n, er := fi.Read(buf)
	//	if er != nil && er != io.EOF {
	//		c.log.Error().Err(er).
	//			Msg("failed to read chunk by chunk from /etc/os-release")
	//	}
	//
	//	if substring(buf[:n], 0) == "ID" {}
	//
	//	if n == 0 {
	//		break
	//	}
	//}

	var distro string

	scanner := bufio.NewScanner(fi)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			// Trim leading 'ID=' and remove quotes if present
			idValue := strings.TrimPrefix(line, "ID=")
			idValue = strings.Trim(idValue, `"`)
			distro = idValue
			c.log.Info().
				Str("distro name", idValue).
				Msg("~distro~")
		}
	}

	if er := scanner.Err(); er != nil {
		c.log.Error().Err(er).
			Msg("failed to read the file /etc/os-release")
	}

	if err = changedRoot(); err != nil {
		c.log.Fatal().Err(err).
			Msg("failed to exit from the updated root")
	}

	os.Exit(0)

	return distro, err
}

func Prerequisites(c Controller, conn *grpc.ClientConn) (bool, string, error) {

	// TODO: how to know the current node is etcd with clientSet?
	//  	- etcd cluster from the cm
	//		- how to get the hostname and ip address
	//  check if external etcd running if it's an etcd node

	distro, err := getDistro(c)

	etcdNode := os.Getenv("ETCD_NODE")
	if etcdNode == "false" {
		return false, distro, errors.New("ETCD_NODE environment variable set false")
	} else if etcdNode == "true" {
		return false, distro, errors.New("ETCD_NODE environment variable set to be True")
	}

	connection := pb.NewMinorUpgradeClient(conn)

	rpc, err := connection.ClusterHealthChecking(c.ctx,
		&pb.PrerequisitesMinorUpgrade{
			EtcdStatus: true,
			// TODO: refactor
			StorageAvailability: 50,
			Err:                 "",
		})

	if err != nil {
		c.log.Error().Err(err).
			Msg("could not send status update")
		return false, distro, err
	}

	c.log.Info().
		Bool("next step", rpc.GetProceedNextStep()).
		Bool("terminate application", rpc.GetProceedNextStep()).
		Msg("prerequisite step response")

	return rpc.GetProceedNextStep(), distro, nil
}
