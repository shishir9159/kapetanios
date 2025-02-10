package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	pb "github.com/shishir9159/kapetanios/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"strings"
)

//	jsoniter "github.com/json-iterator/go"

// minorUpgradeServer is used to implement proto.MinorUpgradeServer.
type minorUpgradeServer struct {
	log  *zap.Logger
	conn *websocket.Conn
	pb.MinorUpgradeServer
}

//var json = jsoniter.ConfigFastest

type clusterHealth struct {
	// todo: whose responsibility is etcdStatus bool?
	etcdStatus          bool   `json:"etcdStatus"`
	storageAvailability uint64 `json:"storageAvailability"`
	err                 string `json:"err"`
}

// TODO: state id ---------
func ClusterHealthReport(nodeHealth clusterHealth, conn *websocket.Conn) (string, error) {

	// todo: create a function payload, expected decision
	if err := conn.WriteJSON(nodeHealth); err != nil {
		return "", err
	}

	msgType, msg, err := conn.ReadMessage()
	if err != nil {
		return "", err
	}

	if msgType != websocket.BinaryMessage {
		return "", fmt.Errorf("unexpected message type: %v", msgType)
	}

	return strings.TrimSpace(string(msg)), nil
}

// ClusterHealthChecking implements proto.MinorUpgradeServer
func (s *minorUpgradeServer) ClusterHealthChecking(_ context.Context, in *pb.PrerequisitesMinorUpgrade) (*pb.UpgradeResponse, error) {

	var proceedNextStep, terminateApplication = false, false

	nodeHealth := clusterHealth{
		etcdStatus:          in.GetEtcdStatus(),
		storageAvailability: in.GetStorageAvailability(),
		err:                 in.GetErr(),
	}

	response, err := ClusterHealthReport(nodeHealth, s.conn)
	if err != nil {
		s.log.Error("Error reporting cluster health",
			zap.Error(err))
	}

	switch response {
	case "next step":
		proceedNextStep = true
		break
	case "terminate application":
		terminateApplication = true
		break
	default:
		response, err = ClusterHealthReport(nodeHealth, s.conn)
		if err != nil {
			s.log.Error("Error reporting cluster health",
				zap.Error(err))
		}
		s.log.Error("unknown response from frontend",
			zap.String("response", response))
	}

	s.log.Info("received cluster health status",
		zap.Bool("etcd status", in.GetEtcdStatus()),
		zap.Uint64("received storage availability", in.GetStorageAvailability()),
		zap.String("received error", in.GetErr()))

	return &pb.UpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// UpgradeVersionSelection implements proto.MinorUpgradeServer
func (s *minorUpgradeServer) UpgradeVersionSelection(_ context.Context, in *pb.AvailableVersions) (*pb.ClusterUpgradeResponse, error) {

	var proceedNextStep, terminateApplication = false, false

	proceedNextStep = true
	if in.GetErr() != "" {

	}

	s.log.Info("available version list",
		zap.Strings("k8s component version", in.GetVersion()),
		zap.String("error", in.GetErr()))

	return &pb.ClusterUpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		TerminateApplication: terminateApplication,
		CertificateRenewal:   certificateRenewal,
		Version:              "",
	}, nil
}

// ClusterCompatibility implements proto.Upgrade
func (s *minorUpgradeServer) ClusterCompatibility(_ context.Context, in *pb.UpgradeCompatibility) (*pb.UpgradeResponse, error) {

	var proceedNextStep, terminateApplication = false, false

	proceedNextStep = true
	if in.GetErr() != "" {

	}

	s.log.Info("received cluster compatibility report",
		zap.Bool("os compatibility", in.GetOsCompatibility()),
		zap.String("kubernetes version diff", in.GetDiff()),
		zap.String("error", in.GetErr()))

	return &pb.UpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// ClusterComponentUpgrade implements proto.Upgrade
func (s *minorUpgradeServer) ClusterComponentUpgrade(_ context.Context, in *pb.ComponentUpgradeStatus) (*pb.UpgradeResponse, error) {

	var proceedNextStep, terminateApplication = false, false

	if in.GetComponentUpgradeSuccess() {
		proceedNextStep = true
	}

	if in.GetErr() != "" {

	}

	s.log.Info("received cluster component upgrade status",
		zap.Bool("component successful upgrade", in.GetComponentUpgradeSuccess()),
		zap.String("component", in.GetComponent()),
		zap.String("log", in.GetLog()),
		zap.String("error", in.GetErr()))

	return &pb.UpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// ClusterUpgradePlan implements proto.Upgrade
func (s *minorUpgradeServer) ClusterUpgradePlan(_ context.Context, in *pb.UpgradePlan) (*pb.UpgradeResponse, error) {

	var proceedNextStep, terminateApplication = false, false

	proceedNextStep = true
	if in.GetErr() != "" {

	}

	s.log.Info("received cluster upgrade plan",
		zap.String("cluster version", in.GetCurrentClusterVersion()),
		zap.String("received log", in.GetLog()),
		zap.String("error", in.GetErr()))

	return &pb.UpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// ClusterUpgrade implements proto.Upgrade
func (s *minorUpgradeServer) ClusterUpgrade(_ context.Context, in *pb.UpgradeStatus) (*pb.UpgradeResponse, error) {

	var proceedNextStep, terminateApplication = false, false

	proceedNextStep = true
	if in.GetErr() != "" {

	}

	s.log.Info("received cluster upgrade plan",
		zap.Bool("cluster upgrade status", in.GetUpgradeSuccess()),
		zap.String("log", in.GetLog()),
		zap.String("error", in.GetErr()))

	return &pb.UpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// ClusterComponentRestart implements proto.Upgrade
func (s *minorUpgradeServer) ClusterComponentRestart(_ context.Context, in *pb.ComponentRestartStatus) (*pb.UpgradeResponse, error) {

	var proceedNextStep, terminateApplication = false, false

	proceedNextStep = true
	if in.GetErr() != "" {

	}

	s.log.Info("received cluster component restart status",
		zap.Bool("component restart success", in.GetComponentRestartSuccess()),
		zap.String("component", in.GetComponent()),
		zap.String("log", in.GetLog()),
		zap.String("error", in.GetErr()))

	return &pb.UpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		TerminateApplication: terminateApplication,
	}, nil
}

func MinorUpgradeGrpc(zlog *zap.Logger, conn *websocket.Conn, ch chan<- *grpc.Server) {

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		zlog.Error("failed to listen",
			zap.Error(err))
	}

	s := grpc.NewServer()
	server := minorUpgradeServer{
		conn: conn,
		log:  zlog,
	}

	// in dev mode
	reflection.Register(s)
	pb.RegisterMinorUpgradeServer(s, &server)

	server.log.Info("upgrade server listening")

	ch <- s
	if er := s.Serve(lis); er != nil {
		server.log.Error("failed to serve",
			zap.Error(er))
	}

	server.log.Info("Shutting down gRPC server...")
}
