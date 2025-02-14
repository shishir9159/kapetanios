package main

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/shishir9159/kapetanios/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

// rollbackServer is used to implement proto.RollbackServer.
type rollbackServer struct {
	pb.RollbackServer
	log *zap.Logger
}

// Prerequisites implements proto.RollbackServer
func (s *rollbackServer) Prerequisites(_ context.Context, in *pb.PrerequisitesRollback) (*pb.RollbackResponse, error) {

	s.log.Info("received cluster prerequisites for rollback",
		zap.Bool("backup exists", in.GetBackupExists()),
		zap.Bool("space availability", in.GetSpaceAvailability()),
		zap.String("error", in.GetErr()))

	return &pb.RollbackResponse{
		ProceedNextStep:      false,
		TerminateApplication: false,
	}, nil
}

// RollbackUpdate implements proto.RollbackServer
func (s *rollbackServer) RollbackUpdate(_ context.Context, in *pb.RollbackStatus) (*pb.RollbackResponse, error) {

	s.log.Info("received cluster certificates rollback status",
		zap.Bool("prerequisites check status", in.GetPrerequisitesCheckSuccess()), // TODO: move the responsibility to previous section
		zap.Bool("renewal success", in.GetRollbackSuccess()),
		zap.Bool("restart success", in.GetRestartSuccess()),
		zap.Uint32("attempt count", in.GetRetryAttempt()), // TODO: might be unnecessary
		zap.String("rollback log", in.GetLog()),
		zap.String("error", in.GetErr()))

	return &pb.RollbackResponse{
		ProceedNextStep:      false,
		TerminateApplication: false,
	}, nil
}

func (s *rollbackServer) Restarts(_ context.Context, in *pb.RollbackRestartStatus) (*pb.RollbackFinalizer, error) {

	s.log.Info("received successful component restart status",
		zap.Bool("etcd restart success", in.GetEtcdRestart()),
		zap.Bool("kubelet restart success", in.GetKubeletRestart()),
		zap.String("etcd error", in.GetEtcdError()),
		zap.String("kubelet error", in.GetKubeletError()),
		// error occurring at the command execution
		zap.String("received error", in.GetErr()))

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
	//reflection.register(s)
	pb.RegisterRollbackServer(s, &rollbackServer{})

	log.Info("rollback sever listening")
	if er := s.Serve(lis); er != nil {

		log.Error("failed to serve",
			zap.Error(er))
	}
}
