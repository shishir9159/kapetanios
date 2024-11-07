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

// ClusterHealthChecking implements proto.MinorUpgradeServer
func (s *minorUpgradeServer) ClusterHealthChecking(_ context.Context, in *pb.PrerequisitesMinorUpgrade) (*pb.UpgradeResponse, error) {

	var proceedNextStep, terminateApplication = false, false

	if in.GetEtcdStatus() && in.GetStorageAvailability() >= 50 {
		proceedNextStep = true
	}

	if in.GetErr() != "" {

	}

	log.Printf("Received etcd status: %t", in.GetEtcdStatus())
	log.Printf("Received storage availability: %v", in.GetStorageAvailability())
	log.Printf("Received error: %v", in.GetErr())

	return &pb.UpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// UpgradeVersionSelection implements proto.MinorUpgradeServer
func (s *minorUpgradeServer) UpgradeVersionSelection(_ context.Context, in *pb.AvailableVersions) (*pb.ClusterUpgradeResponse, error) {
	var proceedNextStep, terminateApplication = false, false

	if in.GetErr() != "" {
		proceedNextStep = true
	}

	log.Printf("Received k8s component version: %s", in.GetVersion())
	log.Printf("Received error: %v", in.GetErr())

	return &pb.ClusterUpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		TerminateApplication: false,
		CertificateRenewal:   terminateApplication,
		Version:              "",
	}, nil
}

// ClusterCompatibility implements proto.Upgrade
func (s *minorUpgradeServer) ClusterCompatibility(_ context.Context, in *pb.UpgradeCompatibility) (*pb.UpgradeResponse, error) {

	var proceedNextStep, terminateApplication = false, false

	if in.GetErr() != "" {
		proceedNextStep = true
	}

	log.Printf("Received os compatibility: %t", in.GetOsCompatibility())
	log.Printf("Received kubernetes version diff: %v", in.GetDiff())
	log.Printf("Received error: %v", in.GetErr())

	return &pb.UpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// ClusterComponentUpgrade implements proto.Upgrade
func (s *minorUpgradeServer) ClusterComponentUpgrade(_ context.Context, in *pb.ComponentUpgradeStatus) (*pb.UpgradeResponse, error) {

	var proceedNextStep, terminateApplication = false, false

	if in.GetComponentUpgradeSuccess() && in.GetErr() != "" {
		proceedNextStep = true
	}

	log.Printf("Received component upgrade status: %t", in.GetComponentUpgradeSuccess())
	log.Printf("for the k8s component: %v", in.GetComponent())
	log.Printf("Received log: %v", in.GetLog())
	log.Printf("Received error: %v", in.GetErr())

	return &pb.UpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// ClusterUpgradePlan implements proto.Upgrade
func (s *minorUpgradeServer) ClusterUpgradePlan(_ context.Context, in *pb.UpgradePlan) (*pb.UpgradeResponse, error) {

	var proceedNextStep, terminateApplication = false, false

	if in.GetErr() != "" {
		proceedNextStep = true
	}

	log.Printf("Received current cluster version: %s", in.GetCurrentClusterVersion())
	log.Printf("Received log: %v", in.GetLog())
	log.Printf("Received error: %v", in.GetErr())

	return &pb.UpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// ClusterUpgrade implements proto.Upgrade
func (s *minorUpgradeServer) ClusterUpgrade(_ context.Context, in *pb.UpgradeStatus) (*pb.UpgradeResponse, error) {

	var proceedNextStep, terminateApplication = false, false

	if in.GetErr() != "" {
		proceedNextStep = true
	}

	log.Printf("Received cluster upgrade status: %t", in.GetUpgradeSuccess())
	log.Printf("Received log: %v", in.GetLog())
	log.Printf("Received error: %v", in.GetErr())

	return &pb.UpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// ClusterComponentRestart implements proto.Upgrade
func (s *minorUpgradeServer) ClusterComponentRestart(_ context.Context, in *pb.ComponentRestartStatus) (*pb.UpgradeResponse, error) {

	var proceedNextStep, terminateApplication = false, false

	if in.GetErr() != "" {
		proceedNextStep = true
	}

	log.Printf("Received component upgraded status: %t", in.GetComponentRestartSuccess())
	log.Printf("for the component: %v", in.GetComponent())
	log.Printf("Received log: %v", in.GetLog())
	log.Printf("Received error: %v", in.GetErr())

	return &pb.UpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		TerminateApplication: terminateApplication,
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

	log.Info("upgrade server listening")

	ch <- s
	if er := s.Serve(lis); er != nil {
		log.Error("failed to serve", zap.Error(er))
	}

	log.Info("Shutting down gRPC server...")
}
