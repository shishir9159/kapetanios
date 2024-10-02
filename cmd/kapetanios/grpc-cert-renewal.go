package main

import (
	"context"

	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/shishir9159/kapetanios/proto"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

// server is used to implement proto.RenewalClient.
type server struct {
	pb.RenewalServer
}

// StatusUpdate implements proto.Renewal
func (s *server) StatusUpdate(_ context.Context, in *pb.CreateRequest) (*pb.CreateResponse, error) {
	log.Printf("Received: %v", in.GetBackupSuccess())
	return &pb.CreateResponse{NextStep: true}, nil
}

func CertGrpc() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterRenewalServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func main() {
	CertGrpc()
}
