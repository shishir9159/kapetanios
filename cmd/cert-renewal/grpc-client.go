package main

import (
	"context"
	"flag"
	pb "github.com/shishir9159/kapetanios/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

const (
	defaultName = "cert-renewal"
)

var (
	addr = flag.String("addr", "kapetanios.default.svc.cluster.local:50051", "the address to connect to")
	//addr = flag.String("addr", "dns:[//10.96.0.1/]kapetanios.default.svc.cluster.local[:50051]", "the address to connect to")
	name = flag.String("name", defaultName, "gRPC test")
)

func GrpcClient(log *zap.Logger) {

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

	c := pb.NewRenewalClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	rpc, err := c.StatusUpdate(ctx,
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
}
