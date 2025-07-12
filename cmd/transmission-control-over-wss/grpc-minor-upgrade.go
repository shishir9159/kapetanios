package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	. "kapetanios/internal/model"
	"kapetanios/internal/wss"
	pb "kapetanios/proto"
	"net"
)

// minorUpgradeServer is used to implement proto.MinorUpgradeServer
type minorUpgradeServer struct {
	log            *zap.Logger
	minorUpgrade   *MinorUpgrade
	state          *UpgradeState
	connectionPool *wss.ConnectionPool
	pb.MinorUpgradeServer
}

type progressionState struct {
	SelectedConfig  *SelectedConfig  `json:"selectedConfig"`
	NodeStatus      *NodesStatus     `json:"nodesStatus"`
	ProgressDetails []ProgressDetail `json:"progressDetails"`
}

type upgradeFinalizer struct {
	NodeStatus *NodesStatus `json:"nodesStatus"`
}

func (progression *progressionState) SendMessage(upgradeState *UpgradeState, log *zap.Logger, pool *wss.ConnectionPool) (bool, bool, bool) {

	var retryCurrentStep, proceedNextStep, terminateApplication = false, false, false

	progression.SelectedConfig = upgradeState.SelectedConfig
	progression.NodeStatus = upgradeState.NodesStatus
	progression.ProgressDetails = upgradeState.ProgressDetails

	upgradeProgression := &UpgradeProgression{
		StateName: "upgrade progression",
		State:     progression,
	}

	payload, err := json.Marshal(upgradeProgression)
	if err != nil {
		log.Error("failed to marshal upgradeProgression",
			zap.Error(err))
	}

	pool.BroadcastMessage(payload)

	lastStep := len(progression.ProgressDetails) - 1

	var response string
	if progression.SelectedConfig.PromptOnError == false || progression.ProgressDetails[lastStep].Err != "" {

		// TODO:
		//  save the last payload in sync.Pool???
		response, err = pool.ReadMessages()
		if err != nil {
			log.Error("failed to read progression response",
				zap.Error(err))
		}
	} else {
		response = "proceedNextStep"
	}

	switch response {
	case "retryCurrentStep":
		retryCurrentStep = true
		break
	case "proceedNextStep":
		proceedNextStep = true
		break
	case "abortMinorUpgrade":
		// TODO:
		// CLean the running pods
		// cancel the main context
		break
	case "terminateApplication":
		upgradeState.NodesStatus.NodesFailedToBeUpgraded =
			append(upgradeState.NodesStatus.NodesFailedToBeUpgraded, upgradeState.NodesStatus.NodesToBeUpgraded[0])
		upgradeState.NodesStatus.NodesToBeUpgraded = upgradeState.NodesStatus.NodesToBeUpgraded[1:]
		terminateApplication = true
		break
	default:
		pool.BroadcastMessage(payload)
		response, err = pool.ReadMessages()
		if err != nil {
			log.Error("error reporting cluster health",
				zap.Error(err))
		}

		if response != "repeat" {
			log.Error("unknown response from frontend",
				zap.String("response", response))
		}
	}

	return retryCurrentStep, proceedNextStep, terminateApplication
}

func appendUniqueStep(progressDetails []ProgressDetail, stepDetail ProgressDetail) []ProgressDetail {
	for _, progress := range progressDetails {
		if progress.StepCount == stepDetail.StepCount {
			progress = stepDetail
			return progressDetails
		}
	}

	return append(progressDetails, stepDetail)
}

// ClusterHealthChecking implements proto.MinorUpgradeServer
func (s *minorUpgradeServer) ClusterHealthChecking(_ context.Context, in *pb.PrerequisitesMinorUpgrade) (*pb.UpgradeResponse, error) {

	var retryCurrentStep, proceedNextStep, terminateApplication = false, false, false

	var stepDetail ProgressDetail
	stepDetail.StepCount = 1
	stepDetail.Log = in.GetLog()
	stepDetail.Err = in.GetErr()
	if stepDetail.Err == "" {
		stepDetail.Log = "completed"
	}

	s.state.ProgressDetails = appendUniqueStep(s.state.ProgressDetails, stepDetail)

	progress := progressionState{}
	retryCurrentStep, proceedNextStep, terminateApplication = progress.SendMessage(s.state, s.log, s.connectionPool)

	// TODO:
	//  handle error if etcd is not healthy
	//  or storage is or storage is not available
	//  storage - at least 1GB is empty or 10% ? empty

	s.log.Info("received cluster health status",
		zap.Bool("etcd status", in.GetEtcdStatus()),
		zap.Uint64("received storage availability", in.GetStorageAvailability()),
		zap.String("received error", in.GetErr()))

	return &pb.UpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		RetryCurrentStep:     retryCurrentStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// PackageAvailabilityCheck implements proto.MinorUpgradeServer
func (s *minorUpgradeServer) PackageAvailabilityCheck(_ context.Context, in *pb.VersionCheck) (*pb.ClusterUpgradeResponse, error) {

	var retryCurrentStep, proceedNextStep, terminateApplication = false, false, false

	var stepDetail ProgressDetail
	stepDetail.StepCount = 2
	stepDetail.Log = ""
	stepDetail.Err = in.GetErr()

	if stepDetail.Err == "" {
		stepDetail.Log = "completed"
	}

	s.state.ProgressDetails = appendUniqueStep(s.state.ProgressDetails, stepDetail)

	progress := progressionState{}

	if in.GetConsistentVersioning() == false {
		// TODO:
		//	 handle error
	}

	// TODO:
	//  proper diagnosis

	retryCurrentStep, proceedNextStep, terminateApplication = progress.SendMessage(s.state, s.log, s.connectionPool)

	s.log.Info("consistent versioning",
		zap.Bool("k8s component version are matched", in.GetConsistentVersioning()),
		zap.String("error", in.GetErr()))

	return &pb.ClusterUpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		RetryCurrentStep:     retryCurrentStep,
		TerminateApplication: terminateApplication,
		CertificateRenewal:   s.state.SelectedConfig.CertificateRenewal,
		Version:              "",
	}, nil
}

// ClusterCompatibility implements proto.MinorUpgradeServer
func (s *minorUpgradeServer) ClusterCompatibility(_ context.Context, in *pb.UpgradeCompatibility) (*pb.UpgradeResponse, error) {

	var retryCurrentStep, proceedNextStep, terminateApplication = false, false, false

	var stepDetail ProgressDetail
	stepDetail.StepCount = 3
	stepDetail.Log = in.GetDiff()
	stepDetail.Err = in.GetErr()

	s.state.ProgressDetails = appendUniqueStep(s.state.ProgressDetails, stepDetail)

	progress := progressionState{}

	if in.GetOsCompatibility() == false {
		// TODO:
		//	 handle error
		//   wrap error with the state
		//   and create prompt for the user
	}

	// TODO:
	//  proper diagnosis
	retryCurrentStep, proceedNextStep, terminateApplication = progress.SendMessage(s.state, s.log, s.connectionPool)

	s.log.Info("received cluster compatibility report",
		zap.Bool("os compatibility", in.GetOsCompatibility()),
		zap.String("kubernetes version diff", in.GetDiff()),
		zap.String("error", in.GetErr()))

	return &pb.UpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		RetryCurrentStep:     retryCurrentStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// ClusterComponentUpgrade implements proto.MinorUpgradeServer
func (s *minorUpgradeServer) ClusterComponentUpgrade(_ context.Context, in *pb.ComponentUpgradeStatus) (*pb.UpgradeResponse, error) {

	var retryCurrentStep, proceedNextStep, terminateApplication = false, false, false

	var stepDetail ProgressDetail
	stepDetail.StepCount = 4
	stepDetail.Log = in.GetLog()
	stepDetail.Err = in.GetErr()

	s.state.ProgressDetails = appendUniqueStep(s.state.ProgressDetails, stepDetail)

	progress := progressionState{}

	if in.GetComponentUpgradeSuccess() == false {
		// TODO:
		//	 handle error
		//   in.GetComponent()
	}

	// TODO:
	//  proper diagnosis
	retryCurrentStep, proceedNextStep, terminateApplication = progress.SendMessage(s.state, s.log, s.connectionPool)

	s.log.Info("received cluster component Upgrade status",
		zap.Bool("component successful Upgrade", in.GetComponentUpgradeSuccess()),
		zap.String("component", in.GetComponent()),
		zap.String("log", in.GetLog()),
		zap.String("error", in.GetErr()))

	return &pb.UpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		RetryCurrentStep:     retryCurrentStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// ClusterUpgradePlan implements proto.MinorUpgradeServer
func (s *minorUpgradeServer) ClusterUpgradePlan(_ context.Context, in *pb.UpgradePlan) (*pb.UpgradeResponse, error) {

	var retryCurrentStep, proceedNextStep, terminateApplication = false, false, false

	var stepDetail ProgressDetail
	stepDetail.StepCount = 5
	stepDetail.Log = in.GetPlan()
	stepDetail.Err = in.GetErr()

	s.state.ProgressDetails = appendUniqueStep(s.state.ProgressDetails, stepDetail)

	progress := progressionState{}

	// TODO:
	//  proper diagnosis

	retryCurrentStep, proceedNextStep, terminateApplication = progress.SendMessage(s.state, s.log, s.connectionPool)

	s.log.Info("received cluster Upgrade plan",
		zap.String("received plan", in.GetPlan()),
		zap.String("error", in.GetErr()))

	return &pb.UpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		RetryCurrentStep:     retryCurrentStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// ClusterUpgrade implements proto.MinorUpgradeServer
func (s *minorUpgradeServer) ClusterUpgrade(_ context.Context, in *pb.UpgradeStatus) (*pb.UpgradeResponse, error) {

	var retryCurrentStep, proceedNextStep, terminateApplication = false, false, false

	var stepDetail ProgressDetail
	stepDetail.StepCount = 6
	stepDetail.Log = in.GetLog()
	stepDetail.Err = in.GetErr()

	s.state.ProgressDetails = appendUniqueStep(s.state.ProgressDetails, stepDetail)

	progress := progressionState{}

	if !in.GetUpgradeSuccess() {
		// TODO:
		//	 handle error
	}

	// TODO:
	//  proper diagnosis

	retryCurrentStep, proceedNextStep, terminateApplication = progress.SendMessage(s.state, s.log, s.connectionPool)

	s.log.Info("received cluster Upgrade plan",
		zap.Bool("cluster Upgrade status", in.GetUpgradeSuccess()),
		zap.String("unsuccessful component", in.GetUnsuccessfulComponent()),
		zap.String("log", in.GetLog()),
		zap.String("error", in.GetErr()))

	return &pb.UpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		RetryCurrentStep:     retryCurrentStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// ClusterComponentRestart implements proto.MinorUpgradeServer
func (s *minorUpgradeServer) ClusterComponentRestart(_ context.Context, in *pb.ComponentRestartStatus) (*pb.UpgradeResponse, error) {

	var retryCurrentStep, proceedNextStep, terminateApplication = false, false, false

	var stepDetail ProgressDetail
	stepDetail.StepCount = 7
	stepDetail.Log = in.GetLog()
	stepDetail.Err = in.GetErr()

	s.state.ProgressDetails = appendUniqueStep(s.state.ProgressDetails, stepDetail)

	progress := progressionState{}

	if in.GetComponentRestartSuccess() == false {
		// TODO:
		//	 handle error
		//	 wrap error with in.GetComponent()s
	}

	// TODO:
	//  proper diagnosis

	retryCurrentStep, proceedNextStep, terminateApplication = progress.SendMessage(s.state, s.log, s.connectionPool)

	s.log.Info("received cluster component restart status",
		zap.Bool("component restart success", in.GetComponentRestartSuccess()),
		zap.String("component", in.GetComponent()),
		zap.String("log", in.GetLog()),
		zap.String("error", in.GetErr()))

	return &pb.UpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		RetryCurrentStep:     retryCurrentStep,
		TerminateApplication: terminateApplication,
	}, nil
}

func MinorUpgradeGrpc(log *zap.Logger, pool *wss.ConnectionPool, state *UpgradeState, ch chan<- *grpc.Server) {

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Error("failed to listen",
			zap.Error(err))
	}

	s := grpc.NewServer()
	server := minorUpgradeServer{
		connectionPool: pool,
		state:          state,
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
