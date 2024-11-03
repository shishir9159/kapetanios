package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
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
	backupCount            = 7
	overRideUserKubeConfig = 0
	name                   = flag.String("name", defaultName, "gRPC test")
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

	//zap.ReplaceGlobals(logger)

	// replace zap with zerolog

	c := Controller{
		ctx: context.Background(),
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
		log.Error("did not connect", zap.Error(err))
	}
	//grpc.WithDisableServiceConfig()
	defer func(conn *grpc.ClientConn) {
		er := conn.Close()
		if er != nil {
			log.Error("failed to close the grpc connection",
				zap.Error(er))
		}
	}(conn)

	connection := pb.NewRenewalClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	rpc, err := connection.StatusUpdate(ctx,
		&pb.CreateRequest{
			BackupSuccess:  true,
			RenewalSuccess: true,
			RetryAttempt:   0,
			RestartSuccess: true,
			Log:            "",
			Err:            "",
		})

	if err != nil {
		log.Error("could not send status update: ", zap.Error(err))
	}

	log.Info("Status Update", zap.Bool("next step", rpc.GetProceedNextStep()), zap.Bool("retry", rpc.GetSkipRetryCurrentStep()))

	err = PrerequisitesForCertRenewal(c.log)
	if err != nil {
		c.log.Error("failed to get cluster health status",
			zap.Error(err))
	}

	//	step 1. Backup directories
	err = BackupCertificatesKubeConfigs(c, backupCount)
	if err != nil {
		c.log.Error("failed to backup certificates and kubeConfigs",
			zap.Error(err))
	}

	//	step 2. Kubeadm certs renew all
	err = Renew(c)
	if err != nil {
		c.log.Error("failed to renew certificates and kubeConfigs",
			zap.Error(err))
	}

	GrpcClient(c.log)

	//step 3. Restarting pods to work with the updated certificates
	err = Restart(c)
	if err != nil {
		c.log.Error("failed to restart kubernetes components after certificate renewal",
			zap.Error(err))
	}

	if overRideUserKubeConfig != 0 {
		err = Copy("/etc/kubernetes/admin.conf", "/root/.kube/config")
		if err != nil {
			c.log.Error("failed to pass kubernetes admin privilege to the root user",
				zap.Error(err))
		}
	}
}
