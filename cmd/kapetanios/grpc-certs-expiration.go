package main

import (
	"context"
	"flag"
	"fmt"
	. "kapetanios/internal/model"
	pb "kapetanios/proto"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// expirationServer is used to implement proto.ValidityServer.
type expirationServer struct {
	pb.ValidityServer
	log                  *zap.Logger
	nodesCertsExpiration *[]NodeCertsExpiration
}

// ClusterHealthChecking implements proto.ValidityServer
func (s *expirationServer) ClusterHealthChecking(_ context.Context, in *pb.PrerequisitesExpiration) (*pb.CertificateValidityResponse, error) {

	s.log.Info("ClusterHealthChecking called",
		zap.Bool("etcd status", in.GetEtcdStatus()),
		zap.Bool("disk pressure", in.GetDiskPressure()),
		zap.String("received error", in.GetErr()))

	return &pb.CertificateValidityResponse{
		ResponseReceived: false,
	}, nil
}

// ExpirationInfo implements proto.ValidityServer
func (s *expirationServer) ExpirationInfo(_ context.Context, in *pb.Expiration) (*pb.CertificateValidityResponse, error) {

	var externallyManagedCerts []string
	var nodeCertsExpiration NodeCertsExpiration

	if in.GetValidCertificate() {

	}

	nodeCertsExpiration.NodeName = in.GetNodeName()
	nodeCertsExpiration.CertificateExpiration = []CertsExpiration{}

	s.log.Info("node name: " + nodeCertsExpiration.NodeName)
	for _, certificate := range in.GetCertificates() {

		// todo: sanity checking
		if certificate.ExternallyManaged == "yes" {
			externallyManagedCerts = append(externallyManagedCerts, certificate.Name)
		}

		var certificateExpiration CertsExpiration

		certificateExpiration.Name = certificate.Name
		certificateExpiration.Expires = certificate.Expires
		certificateExpiration.ResidualTime = certificate.ResidualTime
		certificateExpiration.CertificateAuthority = certificate.CertificateAuthority
		certificateExpiration.ExternallyManaged = certificate.ExternallyManaged

		s.log.Debug("certificate information",
			zap.String("certificate name", certificate.Name),
			zap.String("certificate expires", certificate.Expires),
			zap.String("certificate residual time", certificate.ResidualTime),
			zap.String("certificate authority", certificate.CertificateAuthority),
			zap.String("externally managed", certificate.ExternallyManaged))

		nodeCertsExpiration.CertificateExpiration = append(nodeCertsExpiration.CertificateExpiration, certificateExpiration)
	}

	for _, caAuthority := range in.GetCertificateAuthorities() {

		var certificateAuthority CertsAuthority

		certificateAuthority.Name = caAuthority.Name
		certificateAuthority.Expires = caAuthority.Expires
		certificateAuthority.ResidualTime = caAuthority.ResidualTime
		certificateAuthority.ExternallyManaged = caAuthority.ExternallyManaged

		s.log.Debug("certificate authority information",
			zap.String("certificate authority name", caAuthority.Name),
			zap.String("certificate authority expires", caAuthority.Expires),
			zap.String("certificate authority residual time", caAuthority.ResidualTime),
			zap.String("certificate authority externally managed", caAuthority.ExternallyManaged))

		nodeCertsExpiration.CertificatesAuthority = append(nodeCertsExpiration.CertificatesAuthority, certificateAuthority)
	}

	if len(in.GetMissingCertificates()) != 0 {

		nodeCertsExpiration.MissingCertificate = in.GetMissingCertificates()

		s.log.Info("following certificates are missing",
			zap.Strings("missing certificates", nodeCertsExpiration.MissingCertificate))
	}

	if len(externallyManagedCerts) != 0 {

		s.log.Info("following certificates are externally managed")

		for _, cert := range externallyManagedCerts {
			s.log.Info("certificate: " + cert)
		}

		s.log.Info(`suggestions at current scenario is to remove the node and rejoin by following steps:
		step 1. cordon, drain, delete: kubectl drain <node-name> --ignore-daemonsets --delete-local-data;
		step 2. kubectl delete node <node-name>\n
		step 3. kubeadm token create --print-join-command --config /etc/kubernetes/kubeadm/kubeadm-config.yaml\n
		step 4. kubeadm init phase upload-certs --upload-certs --config /etc/kubernetes/kubeadm/kubeadm-config.yaml\n
		step 5. kubeadm join <master-node>:6443 --token <23-characters-long-token>
                --discovery-token-ca-cert-hash sha256:<64-characters-long-token>
				--control-plane --certificate-key<64-characters-long-certificate-from-the-output-of-step-3>
				--apiserver-advertise-address <new-master-node-ip> --v=14`)
	}

	*s.nodesCertsExpiration = append(*s.nodesCertsExpiration, nodeCertsExpiration)

	return &pb.CertificateValidityResponse{
		ResponseReceived: false,
	}, nil
}

func ExpirationGrpc(zlog *zap.Logger, ch chan<- *grpc.Server, nodesCertsExpiration *[]NodeCertsExpiration) {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		zlog.Error("failed to listen",
			zap.Error(err))
	}

	s := grpc.NewServer()
	server := expirationServer{
		log:                  zlog,
		nodesCertsExpiration: nodesCertsExpiration,
	}

	// in dev mode
	// reflection.register(s)
	pb.RegisterValidityServer(s, &server)

	zlog.Info("cert validity server listening")

	ch <- s
	if er := s.Serve(lis); er != nil {
		zlog.Error("failed to serve",
			zap.Error(er))
	}

	zlog.Info("Shutting down gRPC server...")
}
