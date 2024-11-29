package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"time"
	//"github.com/rs/zerolog/log"
)

var (
	maxAttempts = 3
	addr        = flag.String("addr", "kapetanios.default.svc.cluster.local:50051", "the address to connect to")
)

type Controller struct {
	ctx context.Context
	log zerolog.Logger
}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	logger := zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Caller().Logger()

	// TODO:
	//  replace zap with zeroLog

	c := Controller{
		ctx: ctx,
		log: logger,
	}

	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		c.log.Error().Err(err).
			Msg("couldn't connect to the kapetanios")

	}

	//grpc.WithDisableServiceConfig()
	defer func(conn *grpc.ClientConn) {
		er := conn.Close()
		if er != nil {
			c.log.Error().Err(er).
				Msg("failed to close the grpc connection")
		}
	}(conn)

	for i := 0; i < maxAttempts; i++ {
		skip, er := Prerequisites(c, conn)
		if er != nil {
			c.log.Error().Err(er).
				Int("attempt", i).
				Msg("failed to check prerequisites for cluster upgrade")
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
			c.log.Error().Err(err).
				Int("attempt", i).
				Msg("failed to fetch minor versions for the kubernetes upgrade")
		}

		if skip {
			break
		}
	}

	// todo: include in the testing
	testing := false
	version = "1.26.5-1.1"

	if testing {
		//version = kubernetesVersion
	}

	// TODO: refactor
	//   plan := "v1.26.6"
	//   clusterVersion should be separated from package(+) or components(-) version?

	var diff string
	for i := 0; i < maxAttempts; i++ {
		var skip bool
		skip, diff, err = compatibility(c, "v1.26.5", conn)
		if err != nil {
			c.log.Error().Err(err).
				Int("attempt", i).
				Msg("failed to diff")
		}

		if skip {
			break
		}
	}

	c.log.Info().
		Str("version", version).
		Str("diff", diff).
		Msg("diff for upgrade plan")

	for i := 0; i < maxAttempts; i++ {
		skip, er := k8sComponentsUpgrade(c, "kubeadm", version, conn)
		if er != nil {
			c.log.Error().Err(er).
				Int("attempt", i).
				Msg("failed to upgrade kubeadm")
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
			c.log.Error().Err(err).
				Int("attempt", i).
				Msg("failed to get upgrade plan")
		}

		if skip {
			break
		}
	}

	fmt.Println(plan)

	for i := 0; i < maxAttempts; i++ {
		skip, er := clusterUpgrade(c, version, conn)
		if er != nil {
			c.log.Error().Err(er).
				Int("attempt", i).
				Msg("failed to upgrade cluster")
		}

		if skip {
			break
		}
	}

	for i := 0; i < maxAttempts; i++ {
		skip, er := k8sComponentsUpgrade(c, "kubelet", version, conn)
		if er != nil {
			c.log.Error().Err(er).
				Int("attempt", i).
				Msg("failed to upgrade kubelet")
		}

		if skip {
			break
		}
	}

	for i := 0; i < maxAttempts; i++ {
		skip, er := restartComponent(c, "kubelet", conn)
		if er != nil {
			c.log.Error().Err(er).
				Int("attempt", i).
				Msg("failed to restart kubelet")
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
			c.log.Error().Err(er).
				Int("attempt", i).
				Msg("failed to upgrade kubectl")
		}

		if skip {
			break
		}
	}

	// TODO: sanityChecking & finalizer
}
