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
	"os"
	"slices"
	"strings"
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

type PackageInfo struct {
	Distro              string   `json:"distro"`
	DistroVersion       string   `json:"distroVersion"`
	RepoName            string   `json:"repoName"`
	RepoBaseURL         string   `json:"repoBaseUrl"`
	CurrentK8sVersion   string   `json:"currentK8sVersion"`
	AvailableK8sVersion []string `json:"availableK8sVersion"`
	Log                 string   `json:"log"`
	Err                 string   `json:"err"`
}

type packageInfoForVersionSelection struct {
	State        string         `json:"state"` // TODO: (default value)
	ClusterState *upgradeConfig `json:"clusterState"`
}

// TODO: Create an interface instead of generic that
//  takes the structs write them in the connection
//  and wait for reading

// PackageInfoUpdate implements proto.MinorUpdateServer
func (s *minorUpdateServer) PackageInfoUpdate(_ context.Context, in *pb.PackageInfo) (*pb.UpdateFinalizer, error) {

	var retryCurrentStep, responseReceived = false, false

	packageInfo := &PackageInfo{
		Distro:              in.GetDistro(),
		DistroVersion:       in.GetDistroVersion(), // TODO
		RepoName:            in.GetRepoName(),
		RepoBaseURL:         in.GetRepoBaseURL(),
		CurrentK8sVersion:   in.GetCurrentVersion(),
		AvailableK8sVersion: in.GetVersion(),
		Log:                 in.GetLog(),
		Err:                 in.GetErr(),
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	payload, err := json.Marshal(packageInfo)
	if err != nil {

	}

	// TODO:

	if strings.Contains(packageInfo.Distro, "Red Hat Enterprise Linux 8.") {
		s.upgradeConfig.Redhat8K8sRepoName = packageInfo.RepoName
		s.upgradeConfig.Redhat8K8sRepoBaseURL = packageInfo.RepoBaseURL
	} else if strings.Contains(packageInfo.Distro, "Red Hat Enterprise Linux 9.") {
		s.upgradeConfig.Redhat9K8sRepoName = packageInfo.RepoName
		s.upgradeConfig.Redhat9K8sRepoBaseURL = packageInfo.RepoBaseURL
	}

	minorUpdate := &MinorUpdate{
		Step:  "0",
		State: s.upgradeConfig,
	}

	if strings.Contains(packageInfo.Distro, "Red Hat Enterprise Linux 8.") {
		minorUpdate.State.Redhat8K8sRepoName = packageInfo.RepoName
		minorUpdate.State.Redhat8K8sRepoBaseURL = packageInfo.RepoBaseURL
		minorUpdate.State.Rhel8AvailableK8sVersion = packageInfo.AvailableK8sVersion
	} else if strings.Contains(packageInfo.Distro, "Red Hat Enterprise Linux 9.") {
		minorUpdate.State.Redhat9K8sRepoName = packageInfo.RepoName
		minorUpdate.State.Redhat9K8sRepoBaseURL = packageInfo.RepoBaseURL
		minorUpdate.State.Rhel9AvailableK8sVersion = packageInfo.AvailableK8sVersion
	}

	payload, err = json.Marshal(packageInfo)
	if err != nil {
	}

	// TODO: take prompt from user,
	//  only if an error occurs

	//response, err := s.connectionPool.ReadMessages()
	//if err != nil {
	//	// return nil, err
	//}
	//
	//// TODO:
	////  onPromptError: you already know the answer
	//switch response {
	//case "retry":
	//	retryCurrentStep = true
	//	break
	//case "next step":
	//	proceedNextStep = true
	//	break
	//case "terminate application":
	//	terminateApplication = true
	//	break
	//default:
	//	s.connectionPool.BroadcastMessage(payload)
	//	response, err = s.connectionPool.ReadMessages()
	//	if err != nil {
	//		s.log.Error("error reporting cluster health",
	//			zap.Error(err))
	//	}
	//	s.log.Error("unknown response from frontend",
	//		zap.String("response", response))
	//}

	s.connectionPool.BroadcastMessage(payload)
	response, err := s.connectionPool.ReadMessages()
	if err != nil {
		//return nil, err
	}

	if os.Getenv("DEBUG") == "true" {
		//TODO: response = currentVersion
		// skip broadcast
	}

	switch response {
	case "retry":
		retryCurrentStep = true
		break
	default:
		if slices.Contains(packageInfo.AvailableK8sVersion, response) {
			if strings.Contains(packageInfo.Distro, "Red Hat Enterprise Linux 8.") {
				s.upgradeConfig.Rhel8TargetK8sVersion = response
				s.upgradeConfig.Redhat8K8sRepoName = packageInfo.RepoName
				s.upgradeConfig.Redhat8K8sRepoBaseURL = packageInfo.RepoBaseURL
			} else if strings.Contains(packageInfo.Distro, "Red Hat Enterprise Linux 9.") {
				s.upgradeConfig.Rhel9TargetK8sVersion = response
				s.upgradeConfig.Redhat9K8sRepoName = packageInfo.RepoName
				s.upgradeConfig.Redhat9K8sRepoBaseURL = packageInfo.RepoBaseURL
			}

			responseReceived = true
			break
		}
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

	return &pb.UpdateFinalizer{
		ResponseReceived:     responseReceived,
		RetryCurrentStep:     retryCurrentStep,
		TerminateApplication: false,
	}, nil
}

func MinorUpdateGrpc(log *zap.Logger, pool *wss.ConnectionPool, config *upgradeConfig, ch chan<- *grpc.Server) {

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Error("failed to listen",
			zap.Error(err))
	}

	s := grpc.NewServer()
	server := minorUpdateServer{
		connectionPool: pool,
		upgradeConfig:  config,
		log:            log,
	}

	// in dev mode
	reflection.Register(s)
	pb.RegisterMinorUpdateServer(s, &server)

	server.log.Info("update server listening")

	ch <- s
	if er := s.Serve(lis); er != nil {
		server.log.Error("failed to serve",
			zap.Error(er))
	}

	server.log.Info("Shutting down update gRPC server...")
}
