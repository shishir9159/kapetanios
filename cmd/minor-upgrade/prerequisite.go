package main

import (
	"errors"
	pb "github.com/shishir9159/kapetanios/proto"
	"os"
)

func Prerequisites(c Controller, connection pb.UpgradeClient) error {

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

	return nil
}
