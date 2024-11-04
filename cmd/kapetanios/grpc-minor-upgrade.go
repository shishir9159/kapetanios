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

// server is used to implement proto.MinorUpgradeServer.
type minorUpgradeServer struct {
	pb.MinorUpgradeServer
}

// ClusterHealthChecking implements proto.Upgrade
func (s *minorUpgradeServer) ClusterHealthChecking(_ context.Context, in *pb.PrerequisitesMinorUpgrade) (*pb.UpgradeResponse, error) {
	log.Printf("Received etcd status: %d", in.GetEtcdStatus())
	log.Printf("Received storage availability: %v", in.GetStorageAvailability())
	log.Printf("Received error: %v", in.GetErr())
	return &pb.UpgradeResponse{
		ProceedNextStep:      false,
		SkipRetryCurrentStep: false,
		TerminateApplication: false,
	}, nil
}

// UpgradeVersionSelection implements proto.Upgrade
func (s *minorUpgradeServer) UpgradeVersionSelection(_ context.Context, in *pb.AvailableVersions) (*pb.ClusterUpgradeResponse, error) {
	log.Printf("Received etcd status: %d", in.GetVersion())
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
	log.Printf("Received etcd status: %d", in.GetOsCompatibility())
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
	log.Printf("Received etcd status: %d", in.GetComponentUpgradeSuccess())
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
	log.Printf("Received etcd status: %d", in.GetCurrentClusterVersion())
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
	log.Printf("Received etcd status: %d", in.GetUpgradeSuccess())
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
	log.Printf("Received etcd status: %d", in.GetComponentRestartSuccess())
	log.Printf("Received storage availability: %v", in.GetComponent())
	log.Printf("Received storage availability: %v", in.GetLog())
	log.Printf("Received error: %v", in.GetErr())
	return &pb.UpgradeResponse{
		ProceedNextStep:      false,
		SkipRetryCurrentStep: false,
		TerminateApplication: false,
	}, nil
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
	pb.RegisterMinorUpgradeServer(s, &minorUpgradeServer{})

	log.Info("upgrade sever listening")
	if er := s.Serve(lis); er != nil {

		log.Error("failed to serve", zap.Error(er))
	}
}
