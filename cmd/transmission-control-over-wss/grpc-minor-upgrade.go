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
	"log"
	"net"
	"strings"
)

// minorUpgradeServer is used to implement proto.MinorUpgradeServer.
type minorUpgradeServer struct {
	log     *zap.Logger
	clients map[*websocket.Conn]bool
	pb.MinorUpgradeServer
}

// TODO:
//  retry current step

type clusterHealth struct {
	// todo: whose responsibility is EtcdStatus bool?
	EtcdStatus          bool   `json:"etcdStatus"`
	StorageAvailability uint64 `json:"storageAvailability"`
	Err                 string `json:"err"`
}

type availableVersion struct {
	CurrentVersion   string   `json:"currentVersion"`
	AvailableVersion []string `json:"availableVersion"`
	Err              string   `json:"err"`
}

type clusterCompatability struct {
	OSCompatability bool   `json:"osCompatability"`
	Diff            string `json:"diff"`
	Err             string `json:"err"`
}

type clusterComponentUpgrade struct {
	ComponentUpgradeSuccess bool   `json:"componentUpgradeSuccess"`
	Component               string `json:"component"`
	Log                     string `json:"log"`
	Err                     string `json:"err"`
}

type clusterUpgradePlan struct {
	ClusterVersion string `json:"clusterVersion"`
	Log            string `json:"log"`
	Err            string `json:"err"`
}

type clusterUpgradeSuccess struct {
	UpgradeSuccess bool   `json:"upgradeSuccess"`
	Log            string `json:"log"`
	Err            string `json:"err"`
}

type componentRestartSuccess struct {
	ComponentRestartSuccess bool   `json:"componentRestartSuccess"`
	Component               string `json:"component"`
	Log                     string `json:"log"`
	Err                     string `json:"err"`
}

// TODO: Create an interface instead of generic that
//  takes the structs write them in the connection
//  and wait for reading
// TODO: state id ---------

func readMessage(ctx context.Context, conn *websocket.Conn, messageChan chan string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Error reading from %s: %v", conn.RemoteAddr().String(), err)
				return
			}

			if msgType != websocket.TextMessage {
				log.Printf("unexpected message type: %v", msgType)
			}

			log.Printf("Received from %s: %s", conn.RemoteAddr().String(), string(msg))

			// Send first message and return
			select {
			case messageChan <- string(msg):
			default:
			}
			return
		}
	}
}

func writeMessage[T any](value T, clients map[*websocket.Conn]bool) (string, error) {

	// Create a context with cancel to stop all Goroutines
	ctx, cancel := context.WithCancel(context.Background())

	// Channel to receive the first message
	messageChan := make(chan string, 1)

	for conn := range clients {
		if err := conn.WriteJSON(value); err != nil {
			continue
		}

		// Start reading messages for all clients
		go readMessage(ctx, conn, messageChan)
	}

	// Wait for the first message
	message := <-messageChan

	// Stop all reading Goroutines
	cancel()

	return strings.TrimSpace(string(message)), nil
}

// ClusterHealthChecking implements proto.MinorUpgradeServer
func (s *minorUpgradeServer) ClusterHealthChecking(_ context.Context, in *pb.PrerequisitesMinorUpgrade) (*pb.UpgradeResponse, error) {

	var proceedNextStep, terminateApplication = false, false

	nodeHealth := clusterHealth{
		EtcdStatus:          in.GetEtcdStatus(),
		StorageAvailability: in.GetStorageAvailability(),
		Err:                 in.GetErr(),
	}

	response, err := writeMessage(nodeHealth, s.clients)
	//response, err := ClusterHealthReport(nodeHealth, s.conn)
	if err != nil {
		s.log.Error("error reporting cluster health",
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
		response, err = writeMessage(nodeHealth, s.clients)
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

	versions := availableVersion{
		CurrentVersion:   "",
		AvailableVersion: in.GetVersion(),
		Err:              in.GetErr(),
	}

	response, err := writeMessage(versions, s.clients)
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
		response, err = writeMessage(versions, s.clients)
		if err != nil {
			s.log.Error("Error reporting cluster health",
				zap.Error(err))
		}
		s.log.Error("unknown response from frontend",
			zap.String("response", response))
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

	compatability := clusterCompatability{
		OSCompatability: in.GetOsCompatibility(),
		Diff:            in.GetDiff(),
		Err:             in.GetErr(),
	}

	response, err := writeMessage(compatability, s.clients)
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
		response, err = writeMessage(compatability, s.clients)
		if err != nil {
			s.log.Error("Error reporting cluster health",
				zap.Error(err))
		}
		s.log.Error("unknown response from frontend",
			zap.String("response", response))
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

	componentUpgrade := clusterComponentUpgrade{
		ComponentUpgradeSuccess: in.ComponentUpgradeSuccess,
		Component:               in.GetComponent(),
		Log:                     in.GetLog(),
		Err:                     in.GetErr(),
	}

	response, err := writeMessage(componentUpgrade, s.clients)
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
		response, err = writeMessage(componentUpgrade, s.clients)
		if err != nil {
			s.log.Error("Error reporting cluster health",
				zap.Error(err))
		}
		s.log.Error("unknown response from frontend",
			zap.String("response", response))
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

	upgradePlan := clusterUpgradePlan{
		ClusterVersion: in.GetCurrentClusterVersion(),
		Log:            in.GetLog(),
		Err:            in.GetErr(),
	}

	response, err := writeMessage(upgradePlan, s.clients)
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
		response, err = writeMessage(upgradePlan, s.clients)
		if err != nil {
			s.log.Error("Error reporting cluster health",
				zap.Error(err))
		}
		s.log.Error("unknown response from frontend",
			zap.String("response", response))
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

	upgradeSuccess := clusterUpgradeSuccess{
		UpgradeSuccess: in.GetUpgradeSuccess(),
		Log:            in.GetLog(),
		Err:            in.GetErr(),
	}

	response, err := writeMessage(upgradeSuccess, s.clients)
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
		response, err = writeMessage(upgradeSuccess, s.clients)
		if err != nil {
			s.log.Error("Error reporting cluster health",
				zap.Error(err))
		}
		s.log.Error("unknown response from frontend",
			zap.String("response", response))
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

	restartSuccess := componentRestartSuccess{
		ComponentRestartSuccess: false,
		Component:               "",
		Log:                     "",
		Err:                     "",
	}

	response, err := writeMessage(restartSuccess, s.clients)
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
		response, err = writeMessage(restartSuccess, s.clients)
		if err != nil {
			s.log.Error("Error reporting cluster health",
				zap.Error(err))
		}
		s.log.Error("unknown response from frontend",
			zap.String("response", response))
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

func MinorUpgradeGrpc(zlog *zap.Logger, clients map[*websocket.Conn]bool, ch chan<- *grpc.Server) {

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		zlog.Error("failed to listen",
			zap.Error(err))
	}

	s := grpc.NewServer()
	server := minorUpgradeServer{
		clients: clients,
		log:     zlog,
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
