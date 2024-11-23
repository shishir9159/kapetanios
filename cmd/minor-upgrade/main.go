package main

import (
	"context"
	"flag"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
	//"github.com/rs/zerolog/log"
)

var (
	maxAttempts = 3
	addr        = flag.String("addr", "kapetanios.default.svc.cluster.local:50051", "the address to connect to")
)

type Controller struct {
	ctx context.Context
	log *zap.Logger
}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
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
		c.log.Error("did not connect",
			zap.Error(err))
	}
	//grpc.WithDisableServiceConfig()
	defer func(conn *grpc.ClientConn) {
		er := conn.Close()
		if er != nil {
			c.log.Error("failed to close the grpc connection",
				zap.Error(er))
		}
	}(conn)

	for i := 0; i < maxAttempts; i++ {
		skip, er := Prerequisites(c, conn)
		if er != nil {
			c.log.Error("failed to fetch minor versions for kubernetes version upgrade",
				zap.Int("attempt", i),
				zap.Error(er))
		}

		if skip {
			break
		}
	}

	var version string

	for i := 0; i < maxAttempts; i++ {
		var skip bool
		skip, version, err = availableVersions(c, conn)

		if err != nil {
			c.log.Error("failed to fetch minor versions for kubernetes version upgrade",
				zap.Int("attempt", i),
				zap.Error(err))
		}

		if skip {
			break
		}
	}

	if err != nil {
		c.log.Error("failed to fetch minor versions for kubernetes version upgrade",
			zap.Error(err))
	}

	// todo: include in the testing
	testing := false
	version = "1.26.2-1.1"

	if testing {
		//version = kubernetesVersion
	}

	// TODO: refactor
	//   plan := "v1.26.6"
	//   clusterVersion should be separated from package(+) or components(-) version?

	var diff string
	for i := 0; i < maxAttempts; i++ {
		var skip bool
		skip, diff, err = compatibility(c, "v1.26.2", conn)
		if err != nil {
			c.log.Error("failed to get diff",
				zap.Int("attempt", i),
				zap.Error(err))
		}

		if skip {
			break
		}
	}

	c.log.Info("diff for upgrade plan",
		zap.String("diff", diff))

	for i := 0; i < maxAttempts; i++ {
		skip, er := k8sComponentsUpgrade(c, "kubeadm", version, conn)
		if er != nil {
			c.log.Error("failed to get upgrade kubeadm",
				zap.Int("attempt", i),
				zap.Error(er))
		}

		if skip {
			break
		}
	}

	var plan string
	for i := 0; i < maxAttempts; i++ {
		var skip bool
		skip, plan, err = upgradePlan(c, conn)
		if err != nil {
			c.log.Error("failed to get upgrade plan",
				zap.Int("attempt", i),
				zap.Error(err))
		}

		if skip {
			break
		}
	}

	fmt.Println(plan)

	for i := 0; i < maxAttempts; i++ {
		skip, er := clusterUpgrade(c, version, conn)
		if er != nil {
			c.log.Error("failed to get upgrade plan",
				zap.Int("attempt", i),
				zap.Error(er))
		}

		if skip {
			break
		}
	}

	for i := 0; i < maxAttempts; i++ {
		skip, er := k8sComponentsUpgrade(c, "kubelet", version, conn)
		if er != nil {
			c.log.Error("failed to get upgrade plan",
				zap.Int("attempt", i),
				zap.Error(er))
		}

		if skip {
			break
		}
	}

	for i := 0; i < maxAttempts; i++ {
		skip, er := restartComponent(c, "kubelet", conn)
		if er != nil {
			c.log.Error("failed to restart kubelet",
				zap.Int("attempt", i),
				zap.Error(er))
		}

		if skip {
			break
		}
	}

	// TODO: etcd node restart based
	//  on condition

	for i := 0; i < maxAttempts; i++ {
		skip, er := k8sComponentsUpgrade(c, "kubectl", version, conn)
		if er != nil {
			c.log.Error("failed to get upgrade plan",
				zap.Int("attempt", i),
				zap.Error(er))
		}

		if skip {
			break
		}
	}

	// TODO: sanityChecking & finalizer
}
