package main

import (
	"errors"
	pb "github.com/shishir9159/kapetanios/proto"
	"go.uber.org/zap"
	"os"
)

func Prerequisites(c Controller, connection pb.MinorUpgradeClient) error {

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

	rpc, err := connection.ClusterHealthChecking(c.ctx,
		&pb.PrerequisitesMinorUpgrade{
			EtcdStatus:          false,
			StorageAvailability: 0,
			Err:                 "",
		})

	if err != nil {
		c.log.Error("could not send status update: ", zap.Error(err))
	}

	c.log.Info("Backup Status",
		zap.Bool("next step", rpc.GetProceedNextStep()),
		zap.Bool("retry", rpc.GetSkipRetryCurrentStep()),
		zap.Bool("terminate application", rpc.GetTerminateApplication()))

	return nil
}
