package main

import (
	"context"
	"flag"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	. "kapetanios/internal/model"
	"kapetanios/internal/wss"
	pb "kapetanios/proto"
	"net"
	"os"
	"sync"
)

// TODO:
//  what if the update fails?
//  throw error or try another vm

// minorUpgradeServer is used to implement proto.MinorUpgradeServer.
type minorUpdateServer struct {
	mu             *sync.Mutex
	log            *zap.Logger
	state          *UpgradeState
	connectionPool *wss.ConnectionPool
	pb.MinorUpdateServer
}

// PackageInfoUpdate implements proto.MinorUpdateServer
func (s *minorUpdateServer) PackageInfoUpdate(_ context.Context, in *pb.PackageInfo) (*pb.UpdateFinalizer, error) {

	// TODO:
	var retryCurrentStep, responseReceived = false, true

	packageInfo := K8sPackageInfo{
		Distro:            in.GetDistro(),
		DistroPrettyName:  in.GetDistroPrettyName(),
		RepoName:          in.GetRepoName(),
		RepoBaseURL:       in.GetRepoBaseURL(),
		AvailableVersions: in.GetVersion(),
	}

	//	TODO:
	//   WRITE MUTEX for append
	s.state.K8sPackages = append(s.state.K8sPackages, packageInfo)

	// TODO: take prompt from user,
	//  only if an error occurs

	if os.Getenv("TESTING") == "true" {
		//TODO: response = currentVersion
		// skip broadcast
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

func MinorUpdateGrpc(log *zap.Logger, pool *wss.ConnectionPool, state *UpgradeState, ch chan<- *grpc.Server) {

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Error("failed to listen",
			zap.Error(err))
	}

	s := grpc.NewServer()
	server := minorUpdateServer{
		connectionPool: pool,
		state:          state,
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
