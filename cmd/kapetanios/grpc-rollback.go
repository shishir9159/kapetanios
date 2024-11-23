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

// rollbackServer is used to implement proto.RollbackServer.
type rollbackServer struct {
	pb.RollbackServer
}

// Prerequisites implements proto.RollbackServer
func (s *rollbackServer) Prerequisites(_ context.Context, in *pb.PrerequisitesRollback) (*pb.RollbackResponse, error) {
	log.Printf("received prerequisite check status: %v", in.GetBackupExists())
	log.Printf("received renewal sucess: %v", in.GetSpaceAvailability())
	log.Printf("Received error: %v", in.GetErr())

	return &pb.RollbackResponse{
		ProceedNextStep:      false,
		TerminateApplication: false,
	}, nil
}

// RollbackUpdate implements proto.RollbackServer
func (s *rollbackServer) RollbackUpdate(_ context.Context, in *pb.RollbackStatus) (*pb.RollbackResponse, error) {
	log.Printf("received prerequisite check status: %v", in.GetPrerequisitesCheckSuccess())
	log.Printf("received renewal sucess: %v", in.GetRollbackSuccess())
	log.Printf("received prerequisite check status: %v", in.GetRestartSuccess())
	log.Printf("received renewal sucess: %v", in.GetRetryAttempt())
	log.Printf("received prerequisite check status: %v", in.GetLog())
	log.Printf("Received error: %v", in.GetErr())

	return &pb.RollbackResponse{
		ProceedNextStep:      false,
		TerminateApplication: false,
	}, nil
}

func (s *rollbackServer) Restarts(_ context.Context, in *pb.RollbackRestartStatus) (*pb.RollbackFinalizer, error) {
	log.Printf("Received backup sucess: %v", in.GetEtcdRestart())
	log.Printf("Received renewal sucess: %v", in.GetKubeletRestart())
	log.Printf("Received renewal sucess: %v", in.GetEtcdError())
	log.Printf("Received retry attempt: %s", in.GetKubeletError())
	// error occurring at the command execution
	log.Printf("Received error: %v", in.GetErr())

	return &pb.RollbackFinalizer{
		ResponseReceived: true,
	}, nil
}

func RollbackGrpc(log *zap.Logger) {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Error("failed to listen",
			zap.Error(err))
	}
	s := grpc.NewServer()

	// in dev mode
	//reflection.Register(s)
	pb.RegisterRollbackServer(s, &rollbackServer{})

	log.Info("rollback sever listening")
	if er := s.Serve(lis); er != nil {

		log.Error("failed to serve",
			zap.Error(er))
	}
}
