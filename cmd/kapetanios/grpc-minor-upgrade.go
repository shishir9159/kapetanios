package main

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/shishir9159/kapetanios/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

// server is used to implement proto.MinorUpgradeServer.
type minorUpgradeServer struct {
	pb.MinorUpgradeServer
}

// ClusterHealthChecking implements proto.Upgrade
func (s *minorUpgradeServer) ClusterHealthChecking(_ context.Context, in *pb.PrerequisitesMinorUpgrade) (*pb.UpgradeResponse, error) {

	var proceedNextStep, skipRetryCurrentStep, TerminateApplication = false, false, false

	log.Printf("Received etcd status: %t", in.GetEtcdStatus())
	log.Printf("Received storage availability: %v", in.GetStorageAvailability())
	log.Printf("Received error: %v", in.GetErr())

	if in.GetEtcdStatus() && in.GetStorageAvailability() >= 50 {
		proceedNextStep = true
		skipRetryCurrentStep = true
	}

	if in.GetErr() != "" {

	}

	return &pb.UpgradeResponse{
		ProceedNextStep:      false,
		SkipRetryCurrentStep: false,
		TerminateApplication: false,
	}, nil
}

// UpgradeVersionSelection implements proto.Upgrade
func (s *minorUpgradeServer) UpgradeVersionSelection(_ context.Context, in *pb.AvailableVersions) (*pb.ClusterUpgradeResponse, error) {
	log.Printf("Received etcd status: %s", in.GetVersion())
	log.Printf("Received error: %v", in.GetErr())
	return &pb.ClusterUpgradeResponse{
		ProceedNextStep:      false,
		SkipRetryCurrentStep: false,
		TerminateApplication: false,
		CertificateRenewal:   false,
		Version:              "",
	}, nil
}

// ClusterCompatibility implements proto.Upgrade
func (s *minorUpgradeServer) ClusterCompatibility(_ context.Context, in *pb.UpgradeCompatibility) (*pb.UpgradeResponse, error) {
	log.Printf("Received etcd status: %t", in.GetOsCompatibility())
	log.Printf("Received storage availability: %v", in.GetDiff())
	log.Printf("Received error: %v", in.GetErr())
	return &pb.UpgradeResponse{
		ProceedNextStep:      false,
		SkipRetryCurrentStep: false,
		TerminateApplication: false,
	}, nil
}

// ClusterComponentUpgrade implements proto.Upgrade
func (s *minorUpgradeServer) ClusterComponentUpgrade(_ context.Context, in *pb.ComponentUpgradeStatus) (*pb.UpgradeResponse, error) {
	log.Printf("Received etcd status: %t", in.GetComponentUpgradeSuccess())
	log.Printf("Received storage availability: %v", in.GetComponent())
	log.Printf("Received log: %v", in.GetLog())
	log.Printf("Received error: %v", in.GetErr())
	return &pb.UpgradeResponse{
		ProceedNextStep:      false,
		SkipRetryCurrentStep: false,
		TerminateApplication: false,
	}, nil
}

// ClusterUpgradePlan implements proto.Upgrade
func (s *minorUpgradeServer) ClusterUpgradePlan(_ context.Context, in *pb.UpgradePlan) (*pb.UpgradeResponse, error) {
	log.Printf("Received etcd status: %s", in.GetCurrentClusterVersion())
	log.Printf("Received log: %v", in.GetLog())
	log.Printf("Received error: %v", in.GetErr())
	return &pb.UpgradeResponse{
		ProceedNextStep:      false,
		SkipRetryCurrentStep: false,
		TerminateApplication: false,
	}, nil
}

// ClusterUpgrade implements proto.Upgrade
func (s *minorUpgradeServer) ClusterUpgrade(_ context.Context, in *pb.UpgradeStatus) (*pb.UpgradeResponse, error) {
	log.Printf("Received etcd status: %t", in.GetUpgradeSuccess())
	log.Printf("Received storage availability: %v", in.GetLog())
	log.Printf("Received error: %v", in.GetErr())
	return &pb.UpgradeResponse{
		ProceedNextStep:      false,
		SkipRetryCurrentStep: false,
		TerminateApplication: false,
	}, nil
}

// ClusterComponentRestart implements proto.Upgrade
func (s *minorUpgradeServer) ClusterComponentRestart(_ context.Context, in *pb.ComponentRestartStatus) (*pb.UpgradeResponse, error) {
	log.Printf("Received etcd status: %t", in.GetComponentRestartSuccess())
	log.Printf("Received storage availability: %v", in.GetComponent())
	log.Printf("Received storage availability: %v", in.GetLog())
	log.Printf("Received error: %v", in.GetErr())
	return &pb.UpgradeResponse{
		ProceedNextStep:      false,
		SkipRetryCurrentStep: false,
		TerminateApplication: false,
	}, nil
}

func MinorUpgradeGrpc(log *zap.Logger, ch chan<- *grpc.Server) {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Error("failed to listen",
			zap.Error(err))
	}

	s := grpc.NewServer()

	// in dev mode
	reflection.Register(s)
	pb.RegisterMinorUpgradeServer(s, &minorUpgradeServer{})

	log.Info("upgrade sever listening")

	ch <- s

	if er := s.Serve(lis); er != nil {
		log.Error("failed to serve", zap.Error(er))
	}

	log.Info("Shutting down gRPC server...")
}
