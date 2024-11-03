package main

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/shishir9159/kapetanios/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
	//"github.com/rs/zerolog/log"
)

var (
	addr = flag.String("addr", "kapetanios.default.svc.cluster.local:50051", "the address to connect to")
)

type Controller struct {
	ctx context.Context
	log *zap.Logger
}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Println(err)
	}

	// TODO:
	//  replace zap with zeroLog

	c := Controller{
		ctx: ctx,
		log: logger,
	}

	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {
			c.log.Error("failed to close logger",
				zap.Error(er))
		}
	}(logger)

	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		c.log.Error("did not connect", zap.Error(err))
	}
	//grpc.WithDisableServiceConfig()
	defer func(conn *grpc.ClientConn) {
		er := conn.Close()
		if er != nil {
			c.log.Error("failed to close the grpc connection",
				zap.Error(er))
		}
	}(conn)

	connection := pb.NewRenewalClient(conn)

	err = Prerequisites()
	if err != nil {

	}

	availableVersionList, err := availableVersions(c.log)

	if len(availableVersionList) == 0 {
		c.log.Fatal("no available versions for minor upgrade",
			zap.Error(err))
	}

	if err != nil {
		c.log.Error("failed to fetch minor versions for kubernetes version upgrade",
			zap.Error(err))
	}

	// todo: include in the testing
	testing := false
	latest := false
	version := "1.26.6-1.1"

	if testing {
		//version = kubernetesVersion
	}

	if latest {
	}

	// TODO:
	//  if available version fails or works,
	//  check if that matches with upgradePlane
	//  if the latest is selected

	// TODO: refactor
	plan := "v1.26.6"

	diff, err := compatibility(c.log, plan)
	if err != nil {
		c.log.Error("failed to get diff",
			zap.Error(err))
	}

	c.log.Info("diff for upgrade plan",
		zap.String("diff", diff))

	kubeadmUpgrade, err := k8sComponentsUpgrade(c.log, "kubeadm", version)
	if err != nil {
		c.log.Error("failed to get upgrade kubeadm",
			zap.Error(err))
	}

	if !kubeadmUpgrade {

	}

	plain, err := upgradePlan(c.log)
	if err != nil {
		c.log.Error("failed to get upgrade plan",
			zap.Error(err))
	}

	fmt.Println(plain)

	k8sUpgrade, err := clusterUpgrade(c.log, version)
	if err != nil {
		c.log.Error("failed to get upgrade plan",
			zap.Error(err))
	}

	if !k8sUpgrade {

	}

	kubeletUpgrade, err := k8sComponentsUpgrade(c.log, "kubelet", version)
	if err != nil {
		c.log.Error("failed to get upgrade plan",
			zap.Error(err))
	}

	if !kubeletUpgrade {

	}

	err = restartComponent(c, "kubelet")
	if err != nil {
		c.log.Error("failed to restart kubelet",
			zap.Error(err))
	}

	//etcdNode := os.Getenv("ETCD_NODE")
	//if etcdNode == "true" {
	//	err = restartComponent(c, "etcd")
	//	if err != nil {
	//		c.log.Error("failed to restart etcd",
	//			zap.Error(err))
	//	}
	//}

	_, err = k8sComponentsUpgrade(c.log, "kubectl", version)
	if err != nil {
		c.log.Error("failed to get upgrade plan",
			zap.Error(err))
	}
}
