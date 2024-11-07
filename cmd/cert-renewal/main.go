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
)

const (
	defaultName = "cert-renewal"
)

var ()

var (
	maxAttempts            = 3
	backupCount            = 7
	overRideUserKubeConfig = 0
	//addr = flag.String("addr", "dns:[//10.96.0.1/]kapetanios.default.svc.cluster.local[:50051]", "the address to connect to")
	addr = flag.String("addr", "kapetanios.default.svc.cluster.local:50051", "the address to connect to")
)

type Controller struct {
	ctx context.Context
	log *zap.Logger
}

func main() {

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Println(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	//zap.ReplaceGlobals(logger)

	// replace zap with zerolog

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

	for i := 0; i < maxAttempts; i++ {
		skip, er := PrerequisitesForCertRenewal(c, connection)
		if er != nil {
			c.log.Error("failed to get cluster health status",
				zap.Error(er))
		}

		if skip {
			break
		}
	}

	//	step 1. Backup directories
	for i := 0; i < maxAttempts; i++ {
		skip, er := BackupCertificatesKubeConfigs(c, backupCount, connection)
		if er != nil {
			c.log.Error("failed to backup certificates and kubeConfigs",
				zap.Error(er))
		}

		if skip {
			break
		}
	}

	//	step 2. Kubeadm certs renew all
	for i := 0; i < maxAttempts; i++ {
		skip, er := Renew(c, connection)
		if er != nil {
			c.log.Error("failed to renew certificates and kubeConfigs",
				zap.Error(er))
		}

		if skip {
			break
		}
	}

	//step 3. Restarting pods to work with the updated certificates
	for i := 0; i < maxAttempts; i++ {
		skip, er := Restart(c, connection)
		if er != nil {
			c.log.Error("failed to restart kubernetes components after certificate renewal",
				zap.Error(er))
		}

		if skip {
			break
		}
	}

	if overRideUserKubeConfig != 0 {
		err = Copy("/etc/kubernetes/admin.conf", "/root/.kube/config")
		if err != nil {
			c.log.Error("failed to pass kubernetes admin privilege to the root user",
				zap.Error(err))
		}
	}
}
