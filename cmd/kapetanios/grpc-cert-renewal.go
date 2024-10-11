package main

import (
	"context"
	"flag"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc/reflection"
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
	log.Printf("Received backup sucess: %v", in.GetBackupSuccess())
	log.Printf("Received renewal sucess: %v", in.GetRenewalSuccess())
	log.Printf("Received restart sucess: %v", in.GetRestartSuccess())
	log.Printf("Received log: %v", in.GetLog())
	log.Printf("Received error: %v", in.GetErr())
	return &pb.CreateResponse{ProceedNextStep: true, RetryCurrentStep: false}, nil
}

func CertGrpc(l *zap.Logger) {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		l.Error("failed to listen", zap.Error(err))
	}
	s := grpc.NewServer()

	// in dev mode
	reflection.Register(s)
	pb.RegisterRenewalServer(s, &server{})

	l.Info("sever listening")
	if er := s.Serve(lis); er != nil {
		l.Error("failed to serve", zap.Error(er))
	}
}
