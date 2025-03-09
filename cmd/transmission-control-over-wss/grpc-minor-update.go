package main

import (
	"context"
	"flag"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/shishir9159/kapetanios/internal/wss"
	pb "github.com/shishir9159/kapetanios/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

// TODO: rpc stream
// TODO: error handle

// minorUpgradeServer is used to implement proto.MinorUpgradeServer.
type minorUpdateServer struct {
	log            *zap.Logger
	upgradeConfig  *upgradeConfig
	connectionPool *wss.ConnectionPool
	pb.MinorUpdateServer
}

type prerequisitesCheck struct {
	RepoName       string `json:"repoName"`
	RepoBaseURL    string `json:"repoBaseUrl"`
	CurrentVersion string `json:"currentVersion"`
	Log            string `json:"log"`
	Err            string `json:"err"`
}

type availableVersion struct {
	Distro              string `json:"distro"`
	CurrentK8sVersion   string `json:"currentK8sVersion"`
	AvailableK8sVersion string `json:"availableK8sVersion"`
	Log                 string `json:"log"`
	Err                 string `json:"err"`
}

// TODO: Create an interface instead of generic that
//  takes the structs write them in the connection
//  and wait for reading
// TODO: state id ---------

// PrerequisitesCheck implements proto.MinorUpdateServer
func (s *minorUpdateServer) PrerequisitesCheck(_ context.Context, in *pb.PrerequisitesMinorUpdate) (*pb.UpgradeResponse, error) {

	var retryCurrentStep, proceedNextStep, terminateApplication = false, false, false

	nodeHealth := prerequisitesCheck{
		RepoName:       in.GetRepoName(),
		RepoBaseURL:    in.GetRepoBaseURL(),
		CurrentVersion: in.GetCurrentVersion(),
		Log:            in.GetLog(),
		Err:            in.GetErr(),
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	payload, err := json.Marshal(&nodeHealth)
	if err != nil {

	}

	s.connectionPool.BroadcastMessage(payload)
	response, err := s.connectionPool.ReadMessages()
	if err != nil {
		// return nil, err
	}

	// debugging
	// TODO: make a separate function
	//  make an specific field for each response
	//  dedicated for the decision --

	// TODO:
	//  onPromptError: you already know the answer
	switch response {
	case "retry":
		retryCurrentStep = true
		break
	case "next step":
		proceedNextStep = true
		break
	case "terminate application":
		terminateApplication = true
		break
	default:
		s.connectionPool.BroadcastMessage(payload)
		response, err = s.connectionPool.ReadMessages()
		if err != nil {
			s.log.Error("error reporting cluster health",
				zap.Error(err))
		}
		s.log.Error("unknown response from frontend",
			zap.String("response", response))
	}

	s.log.Info("received cluster health status",
		zap.String("repo name", in.GetRepoName()),
		zap.String("repo baseURL", in.GetRepoBaseURL()),
		zap.String("received storage availability", in.GetCurrentVersion()),
		zap.String("received error", in.GetErr()))

	return &pb.UpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		RetryCurrentStep:     retryCurrentStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// UpdateVersionSelection implements proto.MinorUpdateServer
func (s *minorUpdateServer) UpdateVersionSelection(_ context.Context, in *pb.AvailableVersions) (*pb.ClusterUpgradeResponse, error) {

	var retryCurrentStep, proceedNextStep, terminateApplication = false, false, false

	versions := availableVersion{
		Distro:              in.GetDistro(),
		CurrentK8sVersion:   in.GetCurrentVersion(),
		AvailableK8sVersion: in.GetVersion(),
		Log:                 in.Get,
		Err:                 in.GetErr(),
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	payload, err := json.Marshal(&versions)
	if err != nil {

	}

	s.connectionPool.BroadcastMessage(payload)
	response, err := s.connectionPool.ReadMessages()
	if err != nil {

		//return nil, err
	}

	switch response {
	case "retry":
		retryCurrentStep = true
		break
	case "next step":
		proceedNextStep = true
		break
	case "terminate application":
		terminateApplication = true
		break
	default:
		s.connectionPool.BroadcastMessage(payload)
		response, err = s.connectionPool.ReadMessages()
		if err != nil {
			s.log.Error("error reporting cluster health",
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
		RetryCurrentStep:     retryCurrentStep,
		TerminateApplication: terminateApplication,
		CertificateRenewal:   certificateRenewal,
		Version:              "",
	}, nil
}

func MinorUpdateGrpc(log *zap.Logger, pool *wss.ConnectionPool, ch chan<- *grpc.Server) {

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Error("failed to listen",
			zap.Error(err))
	}

	s := grpc.NewServer()
	server := minorUpgradeServer{
		connectionPool: pool,
		log:            log,
	}

	// in dev mode
	reflection.Register(s)
	pb.RegisterMinorUpgradeServer(s, &server)

	server.log.Info("Upgrade server listening")

	ch <- s
	if er := s.Serve(lis); er != nil {
		server.log.Error("failed to serve",
			zap.Error(er))
	}

	server.log.Info("Shutting down gRPC server...")
}
