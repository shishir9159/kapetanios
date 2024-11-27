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
type renewalServer struct {
	pb.RenewalServer
	log *zap.Logger
}

// ClusterHealthChecking implements proto.RenewalServer
func (s *renewalServer) ClusterHealthChecking(_ context.Context, in *pb.PrerequisitesRenewal) (*pb.RenewalResponse, error) {

	proceedNextStep, terminateApplication := false, false

	if in.GetEtcdStatus() && in.GetKubeDirFreeSpace() >= 50 && in.GetKubeDirFreeSpace() >= 50 {
		proceedNextStep = true
	}

	s.log.Info("received cluster health report",
		zap.Bool("etcd status", in.GetEtcdStatus()),
		zap.Bool("certs externally managed status", in.GetExternallyManagedCerts()),
		zap.Int64("available space in kubernetes config directory", in.GetKubeDirFreeSpace()),
		zap.String("local api endpoint", in.GetLocalAPIEndpoint()),
		zap.String("error", in.GetErr()))

	return &pb.RenewalResponse{
		ProceedNextStep:      proceedNextStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// BackupUpdate implements proto.RenewalServer
func (s *renewalServer) BackupUpdate(_ context.Context, in *pb.BackupStatus) (*pb.RenewalResponse, error) {

	proceedNextStep, terminateApplication := false, false

	if in.GetEtcdBackupSuccess() && in.GetKubeConfigBackupSuccess() && in.GetFileChecklistValidation() {
		proceedNextStep = true
	}

	if in.GetErr() != "" {
		//	TODO: interaction and decide if to retry or terminate application
		proceedNextStep = false
	}

	s.log.Info("received cluster backup status",
		zap.Bool("etcd backup success", in.GetEtcdBackupSuccess()),
		zap.Bool("k8s config backup success", in.GetKubeConfigBackupSuccess()),
		zap.Bool("backup files verified", in.GetFileChecklistValidation()),
		zap.String("backup error", in.GetErr()))

	return &pb.RenewalResponse{
		ProceedNextStep:      proceedNextStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// RenewalUpdate implements proto.RenewalServer
func (s *renewalServer) RenewalUpdate(_ context.Context, in *pb.RenewalStatus) (*pb.RenewalResponse, error) {

	proceedNextStep, terminateApplication := false, false

	if in.GetRenewalSuccess() {
		proceedNextStep = true
	}

	//Todo:
	// updated expiration date
	s.log.Info("received renewal status",
		zap.Bool("renewal success", in.GetRenewalSuccess()),
		zap.String("renewal log", in.GetRenewalLog()),
		zap.String("renewal error", in.GetRenewalLog()),
		zap.String("application log", in.GetLog()),
		zap.String("application error", in.GetErr()))

	return &pb.RenewalResponse{
		ProceedNextStep:      proceedNextStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// RestartUpdate implements proto.RenewalServer
func (s *renewalServer) RestartUpdate(_ context.Context, in *pb.RestartStatus) (*pb.RenewalFinalizer, error) {

	retryComponentsRestart := false

	if in.GetEtcdRestart() && in.GetKubeletRestart() {

	}

	if in.GetErr() != "" {

	}

	s.log.Info("received component successful restart status",
		zap.Bool("etcd restart success", in.GetEtcdRestart()),
		zap.Bool("kubelet restart success", in.GetKubeletRestart()),
		zap.String("etcd logs after restart", in.GetEtcdLog()),
		zap.String("kubelet logs after restart", in.GetKubeletLog()),
		zap.String("etcd errors after restart", in.GetEtcdError()),
		zap.String("kubelet errors after restart", in.GetKubeletError()),
		zap.String("application log", in.GetLog()),
		zap.String("application error", in.GetErr()))

	// error occurring at the command execution
	log.Printf("Received error: %v", in.GetErr())

	return &pb.RenewalFinalizer{
		ResponseReceived:       true,
		RetryCurrentStep:       retryComponentsRestart,
		OverrideUserKubeConfig: true, //TODO: prompt or initial input
	}, nil
}

func CertGrpc(log *zap.Logger, ch chan<- *grpc.Server) {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Error("failed to listen",
			zap.Error(err))
	}

	s := grpc.NewServer()

	// in dev mode
	reflection.Register(s)
	pb.RegisterRenewalServer(s, &renewalServer{})

	log.Info("cert renewal sever listening")

	ch <- s
	if er := s.Serve(lis); er != nil {
		log.Error("failed to serve",
			zap.Error(er))
	}

	log.Info("Shutting down gRPC server...")
}
