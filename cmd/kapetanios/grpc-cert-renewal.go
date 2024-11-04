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

// server is used to implement proto.RenewalClient.
type server struct {
	pb.RenewalServer
}

// ClusterHealthChecking implements proto.Renewal
func (s *server) ClusterHealthChecking(_ context.Context, in *pb.PrerequisitesRenewal) (*pb.CreateResponse, error) {
	log.Printf("Received backup sucess: %v", in.GetEtcdStatus())
	log.Printf("Received renewal sucess: %v", in.GetExternallyManagedCerts())
	log.Printf("Received restart sucess: %v", in.GetKubeDirFreeSpace())
	log.Printf("Received retry attempt: %s", in.GetLocalAPIEndpoint())
	log.Printf("Received error: %v", in.GetErr())

	return &pb.CreateResponse{
		ProceedNextStep:      true,
		SkipRetryCurrentStep: true,
	}, nil
}

// BackupUpdate implements proto.Renewal
func (s *server) BackupUpdate(_ context.Context, in *pb.BackupStatus) (*pb.CreateResponse, error) {
	log.Printf("Received backup sucess: %v", in.GetEtcdBackup())
	log.Printf("Received renewal sucess: %v", in.GetKubeConfigBackup())
	log.Printf("Received restart sucess: %v", in.GetFileChecklistValidation())
	log.Printf("Received error: %v", in.GetErr())

	return &pb.CreateResponse{
		ProceedNextStep:      true,
		SkipRetryCurrentStep: true,
	}, nil
}

// RenewalUpdate implements proto.Renewal
func (s *server) RenewalUpdate(_ context.Context, in *pb.RenewalStatus) (*pb.CreateResponse, error) {

	log.Printf("Received renewal sucess: %v", in.GetRenewalSuccess())
	log.Printf("Received restart sucess: %v", in.GetKubeConfigBackup())
	log.Printf("Received retry attempt: %d", in.GetFileChecklistValidation())
	log.Printf("Received error: %v", in.GetErr())

	return &pb.CreateResponse{
		ProceedNextStep:      true,
		SkipRetryCurrentStep: true,
	}, nil
}

// RestartUpdate implements proto.Renewal
func (s *server) RestartUpdate(_ context.Context, in *pb.RestartStatus) (*pb.CreateResponse, error) {
	log.Printf("Received backup sucess: %v", in.GetEtcdRestart())
	log.Printf("Received renewal sucess: %v", in.GetKubeletRestart())
	log.Printf("Received restart sucess: %v", in.GetEtcdRestart())
	log.Printf("Received retry attempt: %s", in.GetKubeletError())
	log.Printf("Received error: %v", in.GetErr())

	return &pb.CreateResponse{
		ProceedNextStep:      true,
		SkipRetryCurrentStep: true,
	}, nil
}

func CertGrpc(log *zap.Logger) {
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
	if er := s.Serve(lis); er != nil {

		log.Error("failed to serve", zap.Error(er))
	}
}
