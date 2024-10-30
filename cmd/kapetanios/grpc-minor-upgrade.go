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

//var (
//	port = flag.Int("port", 50051, "The server port")
//)

// server is used to implement proto.RenewalClient.
type minorUpgradeServer struct {
	pb.UpgradeServer
}

// StatusUpdate implements proto.Upgrade
func (s *minorUpgradeServer) StatusUpdate(_ context.Context, in *pb.CreateUpgradeRequest) (*pb.CreateUpgradeResponse, error) {
	log.Printf("Received restart sucess: %v", in.GetRestartSuccess())
	log.Printf("Received retry attempt: %d", in.GetRetryAttempt())
	log.Printf("Received log: %v", in.GetLog())
	log.Printf("Received error: %v", in.GetErr())
	return &pb.CreateUpgradeResponse{}, nil
}

func MinorUpgradeGrpc(log *zap.Logger) {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Error("failed to listen", zap.Error(err))
	}
	s := grpc.NewServer()

	// in dev mode
	reflection.Register(s)
	//pb.RegisterRenewalServer(s, &minorUpgradeServer{})

	log.Info("upgrade sever listening")
	if er := s.Serve(lis); er != nil {

		log.Error("failed to serve", zap.Error(er))
	}
}
