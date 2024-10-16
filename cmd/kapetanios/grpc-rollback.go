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
// port = flag.Int("port", 50051, "The server port")
)

// server is used to implement proto.RenewalClient.
type rollbackServer struct {
	pb.RollbackServer
}

// StatusUpdate implements proto.Renewal
func (s *rollbackServer) StatusUpdate(_ context.Context, in *pb.CreateRollbackRequest) (*pb.CreateRollbackResponse, error) {
	log.Printf("received prerequisite check status: %v", in.GetPrerequisiteCheckSuccess())
	log.Printf("received renewal sucess: %v", in.GetRollbackSuccess())
	log.Printf("received restart sucess: %v", in.GetRestartSuccess())
	log.Printf("received mumber of retry attempt: %d", in.GetRetryAttempt())
	log.Printf("received log: %v", in.GetLog())
	log.Printf("Received error: %v", in.GetErr())

	return &pb.CreateRollbackResponse{
		ProceedNextStep:      true,
		SkipRetryCurrentStep: true,
		TerminateApplication: true,
	}, nil
}

func RollbackGrpc(log *zap.Logger) {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Error("failed to listen", zap.Error(err))
	}
	s := grpc.NewServer()

	// in dev mode
	reflection.Register(s)
	pb.RegisterRollbackServer(s, &rollbackServer{})

	log.Info("rollback sever listening")
	if er := s.Serve(lis); er != nil {

		log.Error("failed to serve", zap.Error(er))
	}
}
