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

// server is used to implement proto.RenewalServer.
type server struct {
	pb.RenewalServer
}

// ClusterHealthChecking implements proto.RenewalServer
func (s *server) ClusterHealthChecking(_ context.Context, in *pb.PrerequisitesRenewal) (*pb.RenewalResponse, error) {

	proceedNextStep, terminateApplication := false, false

	if in.GetEtcdStatus() && in.GetKubeDirFreeSpace() >= 50 && in.GetKubeDirFreeSpace() >= 50 {
		proceedNextStep = true
	}

	log.Printf("Received backup sucess: %v", in.GetEtcdStatus())
	log.Printf("Received renewal sucess: %v", in.GetExternallyManagedCerts())
	log.Printf("Received restart sucess: %v", in.GetKubeDirFreeSpace())
	log.Printf("Received retry attempt: %s", in.GetLocalAPIEndpoint())
	log.Printf("Received error: %v", in.GetErr())

	return &pb.RenewalResponse{
		ProceedNextStep:      proceedNextStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// BackupUpdate implements proto.RenewalServer
func (s *server) BackupUpdate(_ context.Context, in *pb.BackupStatus) (*pb.RenewalResponse, error) {

	proceedNextStep, terminateApplication := false, false

	if in.GetEtcdBackupSuccess() && in.GetKubeConfigBackupSuccess() && in.GetFileChecklistValidation() {
		proceedNextStep = true
	}

	if in.GetErr() != "" {
		//	TODO: interaction and decide if to retry or terminate application
		proceedNextStep = false
	}

	log.Printf("Received backup sucess: %v", in.GetEtcdBackupSuccess())
	log.Printf("Received renewal sucess: %v", in.GetKubeConfigBackupSuccess())
	log.Printf("Received restart sucess: %v", in.GetFileChecklistValidation())
	log.Printf("Received error: %v", in.GetErr())

	return &pb.RenewalResponse{
		ProceedNextStep:      proceedNextStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// RenewalUpdate implements proto.RenewalServer
func (s *server) RenewalUpdate(_ context.Context, in *pb.RenewalStatus) (*pb.RenewalResponse, error) {

	proceedNextStep, terminateApplication := false, false

	if in.GetRenewalSuccess() && in.GetKubeConfigBackup() && in.GetFileChecklistValidation() {
		proceedNextStep = true
	}

	log.Printf("Received renewal sucess: %v", in.GetRenewalSuccess())
	log.Printf("Received restart sucess: %v", in.GetKubeConfigBackup())
	log.Printf("Received retry attempt: %d", in.GetFileChecklistValidation())
	log.Printf("Received error: %v", in.GetErr())

	return &pb.RenewalResponse{
		ProceedNextStep:      proceedNextStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// RestartUpdate implements proto.RenewalServer
func (s *server) RestartUpdate(_ context.Context, in *pb.RestartStatus) (*pb.RenewalFinalizer, error) {

	gracefullyShutDown, retryRestartingComponents := false, false

	if in.GetEtcdRestart() && in.GetKubeletRestart() {
		gracefullyShutDown = true
	}

	if in.GetErr() != "" {

	}

	log.Printf("Received backup sucess: %v", in.GetEtcdRestart())
	log.Printf("Received renewal sucess: %v", in.GetKubeletRestart())
	log.Printf("Received renewal sucess: %v", in.GetEtcdError())
	log.Printf("Received retry attempt: %s", in.GetKubeletError())

	// error occurring at the command execution
	log.Printf("Received error: %v", in.GetErr())

	return &pb.RenewalFinalizer{
		GracefullyShutDown:        gracefullyShutDown,
		RetryRestartingComponents: retryRestartingComponents,
		OverrideUserKubeConfig:    true, //TODO: prompt
	}, nil
}

func CertGrpc(log *zap.Logger, ch chan<- *grpc.Server) {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Error("failed to listen", zap.Error(err))
	}

	s := grpc.NewServer()

	// in dev mode
	reflection.Register(s)
	pb.RegisterRenewalServer(s, &server{})

	log.Info("cert renewal sever listening")

	ch <- s
	if er := s.Serve(lis); er != nil {
		log.Error("failed to serve", zap.Error(er))
	}

	log.Info("Shutting down gRPC server...")
}
