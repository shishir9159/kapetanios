package main

import (
	"errors"
	pb "github.com/shishir9159/kapetanios/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"os"
)

func Prerequisites(c Controller, conn *grpc.ClientConn) error {

	// TODO: how to know the current node is etcd with clientSet?
	//  	- etcd cluster from the cm
	//		- how to get the hostname and ip address
	//  check if external etcd running if it's an etcd node

	etcdNode := os.Getenv("ETCD_NODE")
	if etcdNode == "false" {
		return errors.New("ETCD_NODE environment variable set false")
	} else if etcdNode == "true" {
		return errors.New("ETCD_NODE environment variable set to be True")
	}

	connection := pb.NewMinorUpgradeClient(conn)

	rpc, err := connection.ClusterHealthChecking(c.ctx,
		&pb.PrerequisitesMinorUpgrade{
			EtcdStatus:          true,
			StorageAvailability: 50,
			Err:                 "",
		})

	if err != nil {
		c.log.Error("could not send status update: ", zap.Error(err))
		return err
	}

	c.log.Info("prerequisite step response",
		zap.Bool("next step", rpc.GetProceedNextStep()),
		zap.Bool("retry", rpc.GetSkipRetryCurrentStep()),
		zap.Bool("terminate application", rpc.GetTerminateApplication()))

	return nil
}
