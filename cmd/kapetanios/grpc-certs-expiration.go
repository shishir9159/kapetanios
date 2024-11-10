package main

import (
	"context"
	"flag"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc/reflection"
	"net"

	pb "github.com/shishir9159/kapetanios/proto"
	"google.golang.org/grpc"
)

// expirationServer is used to implement proto.ValidityServer.
type expirationServer struct {
	pb.ValidityServer
	log *zap.Logger
}

// ClusterHealthChecking implements proto.ValidityServer
func (s *expirationServer) ClusterHealthChecking(_ context.Context, in *pb.PrerequisitesExpiration) (*pb.CertificateValidityResponse, error) {

	s.log.Info("ClusterHealthChecking called",
		zap.Bool("etcd status", in.GetEtcdStatus()),
		zap.Bool("disk pressure", in.GetDiskPressure()),
		zap.String("received error", in.GetErr()))

	return &pb.CertificateValidityResponse{
		Received: true,
	}, nil
}

// ExpirationInfo implements proto.ValidityServer
func (s *expirationServer) ExpirationInfo(_ context.Context, in *pb.Expiration) (*pb.CertificateValidityResponse, error) {

	if in.GetValidCertificate() {

	}

	for _, certificate := range in.GetCertificates() {
		s.log.Info("certificate name: " + certificate.Name +
			"certificate expires: " + certificate.Expires +
			"certificate residual time: " + certificate.ResidualTime +
			"certificate authority: " + certificate.CertificateAuthority +
			"externally managed: " + certificate.ExternallyManaged)
	}

	for _, caAuthority := range in.GetCertificateAuthorities() {
		s.log.Info("certificate authority name: " + caAuthority.Name +
			"certificate expires: " + caAuthority.Expires +
			"certificate residual time: " + caAuthority.ResidualTime +
			"externally managed: " + caAuthority.ExternallyManaged)
	}

	return &pb.CertificateValidityResponse{
		Received: true,
	}, nil
}

func ExpirationGrpc(zlog *zap.Logger, ch chan<- *grpc.Server) {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		zlog.Error("failed to listen", zap.Error(err))
	}

	s := grpc.NewServer()

	// in dev mode
	reflection.Register(s)
	pb.RegisterValidityServer(s, &expirationServer{
		log: zlog,
	})

	zlog.Info("cert validity server listening")

	ch <- s
	if er := s.Serve(lis); er != nil {
		zlog.Error("failed to serve", zap.Error(er))
	}

	zlog.Info("Shutting down gRPC server...")
}
