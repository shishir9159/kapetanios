package main

import (
	"errors"
	"os"
)

func Prerequisites() error {

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

	// TODO: kernel version compatibility

	// TODO:
	//  Check which versions are available to upgrade to and
	//  validate whether your current cluster is upgradeable

	// space usage
	// new kubectl, kubeadm, kubelet versions and all the other

	return nil
}
