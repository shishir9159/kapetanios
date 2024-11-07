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

// expirationServer is used to implement proto.ValidityServer.
type expirationServer struct {
	pb.ValidityServer
}

// ClusterHealthChecking implements proto.ValidityServer
func (s *expirationServer) ClusterHealthChecking(_ context.Context, in *pb.PrerequisitesExpiration) (*pb.CertificateValidityResponse, error) {

	log.Printf("Received backup sucess: %v", in.GetEtcdStatus())
	log.Printf("Received disk pressure status: %v", in.GetDiskPressure())
	log.Printf("Received error: %v", in.GetErr())

	return &pb.CertificateValidityResponse{
		Received: true,
	}, nil
}

// ExpirationInfo implements proto.ValidityServer
func (s *expirationServer) ExpirationInfo(_ context.Context, in *pb.Expiration) (*pb.CertificateValidityResponse, error) {

	if in.GetValidCertificate() {

	}

	for certificate, index := range in.GetCertificates() {
		log.Println(certificate, " index: ", index)
	}

	for caAuthority, index := range in.GetCertificateAuthorities() {
		log.Println(caAuthority, " index: ", index)
	}

	return &pb.CertificateValidityResponse{
		Received: true,
	}, nil
}

func ExpirationGrpc(log *zap.Logger, ch chan<- *grpc.Server) {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Error("failed to listen", zap.Error(err))
	}

	s := grpc.NewServer()

	// in dev mode
	reflection.Register(s)
	pb.RegisterValidityServer(s, &expirationServer{})

	log.Info("cert validity server listening")

	ch <- s
	if er := s.Serve(lis); er != nil {
		log.Error("failed to serve", zap.Error(er))
	}

	log.Info("Shutting down gRPC server...")
}
