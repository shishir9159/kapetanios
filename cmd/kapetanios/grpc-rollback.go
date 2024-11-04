package main

import (
	"context"
	"flag"
	"fmt"
	"go.uber.org/zap"
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

// Prerequisite implements proto.Renewal
func (s *rollbackServer) Prerequisite(_ context.Context, in *pb.PrerequisitesRollback) (*pb.CreateResponse, error) {
	log.Printf("received prerequisite check status: %v", in.GetBackupExists())
	log.Printf("received renewal sucess: %v", in.GetBackupExists())
	log.Printf("Received error: %v", in.GetErr())

	return &pb.CreateResponse{
		ProceedNextStep:      true,
		SkipRetryCurrentStep: true,
		TerminateApplication: true,
	}, nil
}

// RollbackUpdate implements proto.Renewal
func (s *rollbackServer) RollbackUpdate(_ context.Context, in *pb.RollbackStatus) (*pb.CreateResponse, error) {
	log.Printf("received prerequisite check status: %v", in.GetPrerequisitesCheckSuccess())
	log.Printf("received renewal sucess: %v", in.GetRollbackSuccess())
	log.Printf("received prerequisite check status: %v", in.GetRestartSuccess())
	log.Printf("received renewal sucess: %v", in.GetRetryAttempt())
	log.Printf("received prerequisite check status: %v", in.GetLog())
	log.Printf("Received error: %v", in.GetErr())

	return &pb.CreateResponse{
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
	//reflection.Register(s)
	pb.RegisterRollbackServer(s, &rollbackServer{})

	log.Info("rollback sever listening")
	if er := s.Serve(lis); er != nil {

		log.Error("failed to serve", zap.Error(er))
	}
}
