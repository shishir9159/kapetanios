package main

import (
	"context"
	"flag"
	"github.com/rs/zerolog"
	pb "github.com/shishir9159/kapetanios/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"time"
)

var (
	maxAttempts = 3
	backupCount = 7
	//addr = flag.String("addr", "dns:[//10.96.0.1/]kapetanios.default.svc.cluster.local[:50051]", "the address to connect to")
	addr = flag.String("addr", "kapetanios.default.svc.cluster.local:50051", "the address to connect to")
	//"name": [{"service": "grpc.examples.echo.Echo"}],
	retryPolicy = `{
		"methodConfig": [{
		  "retryPolicy": {
			  "MaxAttempts": 4,
			  "InitialBackoff": ".01s",
			  "MaxBackoff": ".01s",
			  "BackoffMultiplier": 1.0,
			  "RetryableStatusCodes": [ "UNAVAILABLE" ]
		  }
		}]}`
)

type Controller struct {
	ctx context.Context
	log zerolog.Logger
}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	logger := zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Logger()

	c := Controller{
		ctx: ctx,
		log: logger,
	}

	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.NewClient(
		*addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(retryPolicy))

	if err != nil {
		c.log.Error().Err(err).
			Msg("could not connect to kapetanios")
	}

	//grpc.WithDisableServiceConfig()
	defer func(conn *grpc.ClientConn) {
		er := conn.Close()
		if er != nil {
			c.log.Error().Err(er).
				Msg("could not close the grpc connection to kapetanios")
		}
	}(conn)

	connection := pb.NewRenewalClient(conn)

	for i := 0; i < maxAttempts; i++ {
		skip, er := PrerequisitesForCertRenewal(c, connection)
		if er != nil {
			c.log.Error().Err(er).
				Msg("failed to get cluster health status")
		}

		if skip {
			break
		}
	}

	//	step 1. Backup directories
	for i := 0; i < maxAttempts; i++ {
		skip, er := BackupCertificatesKubeConfigs(c, backupCount, connection)
		if er != nil {
			c.log.Error().Err(er).
				Msg("failed to backup certificates and kubeConfigs")

		}

		if skip {
			break
		}
	}

	//	step 2. Kubeadm certs renew all
	for i := 0; i < maxAttempts; i++ {
		skip, er := Renew(c, connection)
		if er != nil {
			c.log.Error().Err(er).
				Msg("failed to renew the certificates and kubeConfigs")
		}

		if skip {
			break
		}
	}

	var overrideUserKubeConfig bool

	//step 3. Restarting pods to work with the updated certificates
	for i := 0; i < maxAttempts; i++ {
		var skip bool
		skip, overrideUserKubeConfig, err = Restart(c, connection)
		if err != nil {
			c.log.Error().Err(err).
				Msg("failed to restart the certificates and kubeConfigs")
		}

		if !skip {
			break
		}
	}

	if overrideUserKubeConfig {

		// TODO: syscall + mkdir -p folder

		//err = Copy("/etc/kubernetes/admin.conf", "/root/.kube/config")
		//if err != nil {
		//	c.log.Error("failed to pass kubernetes admin privilege to the root user",
		//		zap.Error(err))
		//}

	}
}
