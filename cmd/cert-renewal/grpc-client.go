package main

import (
	"context"
	"flag"
	"go.uber.org/zap"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/shishir9159/kapetanios/proto"
)

const (
	defaultName = "cert-renewal"
)

var (
	addr = flag.String("addr", "kapetanios.com:50051", "the address to connect to")
	//addr = flag.String("addr", "kapetanios.default.svc.cluster.local:50051", "the address to connect to")
	name = flag.String("name", defaultName, "gRPC test")
)

func GrpcClient(log *zap.Logger) {

	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("did not connect", zap.Error(err))
	}
	defer func(conn *grpc.ClientConn) {
		er := conn.Close()
		if er != nil {

		}
	}(conn)
	c := pb.NewRenewalClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.StatusUpdate(ctx, &pb.CreateRequest{BackupSuccess: true, RenewalSuccess: true, RestartSuccess: true})
	if err != nil {
		log.Error("could not send status update: ", zap.Error(err))
	}
	log.Error("Status Update", zap.Bool("next step", r.GetNextStep()))
}
