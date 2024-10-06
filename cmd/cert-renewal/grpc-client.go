package main

import (
	"context"
	"flag"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/shishir9159/kapetanios/proto"
)

const (
	defaultName = "cert-renewal"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
	name = flag.String("name", defaultName, "gRPC test")
)

func GrpcClient() {

	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
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
		log.Fatalf("could not send status update: %v", err)
	}
	log.Printf("Status Update: %t", r.GetNextStep())
}
