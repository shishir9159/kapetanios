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

// TODO: error handle per action

// minorUpgradeServer is used to implement proto.MinorUpgradeServer.
type minorUpgradeServer struct {
	log            *zap.Logger
	minorUpgrade   *MinorUpgrade
	state          *UpgradeState
	connectionPool *wss.ConnectionPool
	pb.MinorUpgradeServer
}

type progressionState struct {
	SelectedConfig  *selectedConfig   `json:"selectedConfig"`
	NodeStatus      *nodesStatus      `json:"nodesStatus"`
	ProgressDetails []progressDetails `json:"progressDetails"`
}

func (progression *progressionState) SendMessage(upgradeState *UpgradeState, log *zap.Logger, pool *wss.ConnectionPool) (bool, bool, bool) {

	var retryCurrentStep, proceedNextStep, terminateApplication = false, false, false

	upgradeProgression := &UpgradeProgression{
		StateName: "upgrade progression",
		State:     progression,
	}

	json := jsoniter.ConfigFastest
	payload, err := json.Marshal(upgradeProgression)
	if err != nil {
		log.Error("failed to marshal upgradeProgression",
			zap.Error(err))
	}

	pool.BroadcastMessage(payload)

	response, err := pool.ReadMessages()

	if err != nil {
		// return nil, err
	}

	// TODO:
	//  if prompt on error

	// TODO:
	//  save the last payload???

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
		pool.BroadcastMessage(payload)
		response, err = pool.ReadMessages()
		if err != nil {
			log.Error("error reporting cluster health",
				zap.Error(err))
		}
		log.Error("unknown response from frontend",
			zap.String("response", response))
	}

	selections := selectedConfig{}
	err = json.Unmarshal([]byte(response), &selections)
	if err != nil {

	}

	upgradeState.SelectedConfig = &selections

	return retryCurrentStep, proceedNextStep, terminateApplication
}

// ClusterHealthChecking implements proto.MinorUpgradeServer
func (s *minorUpgradeServer) ClusterHealthChecking(_ context.Context, in *pb.PrerequisitesMinorUpgrade) (*pb.UpgradeResponse, error) {

	var progressDetail progressDetails
	progressDetail.StepCount = 1
	progressDetail.Log = in.GetLog()
	progressDetail.Err = in.GetErr()

	// override by stepCount
	s.state.ProgressDetails = append(s.state.ProgressDetails, progressDetail)

	progress := progressionState{
		SelectedConfig:  s.state.SelectedConfig,
		NodeStatus:      s.state.NodeStatus,
		ProgressDetails: s.state.ProgressDetails,
	}

	retryCurrentStep, proceedNextStep, terminateApplication := progress.SendMessage(s.state, s.log, s.connectionPool)

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

	var progressDetail progressDetails
	progressDetail.StepCount = 2
	progressDetail.Log = ""
	progressDetail.Err = in.GetErr()
	if progressDetail.Err == "" {
		progressDetail.Log = "completed"
	}

	s.state.ProgressDetails = append(s.state.ProgressDetails, progressDetail)

	progress := progressionState{
		SelectedConfig:  s.state.SelectedConfig,
		NodeStatus:      s.state.NodeStatus,
		ProgressDetails: s.state.ProgressDetails,
	}

	if in.GetConsistentVersioning() == false {
		// TODO:
		//	 handle error
	}

	retryCurrentStep, proceedNextStep, terminateApplication := progress.SendMessage(s.state, s.log, s.connectionPool)

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

// ClusterCompatibility implements proto.Upgrade
func (s *minorUpgradeServer) ClusterCompatibility(_ context.Context, in *pb.UpgradeCompatibility) (*pb.UpgradeResponse, error) {

	var progressDetail progressDetails
	progressDetail.StepCount = 3
	progressDetail.Log = in.GetDiff()
	progressDetail.Err = in.GetErr()

	s.state.ProgressDetails = append(s.state.ProgressDetails, progressDetail)

	progress := progressionState{
		SelectedConfig:  s.state.SelectedConfig,
		NodeStatus:      s.state.NodeStatus,
		ProgressDetails: s.state.ProgressDetails,
	}

	if in.GetOsCompatibility() == false {
		// TODO:
		//	 handle error
		//   wrap error with the state
		//   and create prompt for the user
	}

	retryCurrentStep, proceedNextStep, terminateApplication := progress.SendMessage(s.state, s.log, s.connectionPool)

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

// ClusterComponentUpgrade implements proto.Upgrade
func (s *minorUpgradeServer) ClusterComponentUpgrade(_ context.Context, in *pb.ComponentUpgradeStatus) (*pb.UpgradeResponse, error) {

	var progressDetail progressDetails
	progressDetail.StepCount = 4
	progressDetail.Log = in.GetLog()
	progressDetail.Err = in.GetErr()

	s.state.ProgressDetails = append(s.state.ProgressDetails, progressDetail)

	progress := progressionState{
		SelectedConfig:  s.state.SelectedConfig,
		NodeStatus:      s.state.NodeStatus,
		ProgressDetails: s.state.ProgressDetails,
	}

	if in.GetComponentUpgradeSuccess() == false {
		// TODO:
		//	 handle error
		//   in.GetComponent()
	}

	retryCurrentStep, proceedNextStep, terminateApplication := progress.SendMessage(s.state, s.log, s.connectionPool)

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

// ClusterUpgradePlan implements proto.Upgrade
func (s *minorUpgradeServer) ClusterUpgradePlan(_ context.Context, in *pb.UpgradePlan) (*pb.UpgradeResponse, error) {

	var progressDetail progressDetails
	progressDetail.StepCount = 5
	progressDetail.Log = in.GetPlan()
	progressDetail.Err = in.GetErr()

	s.state.ProgressDetails = append(s.state.ProgressDetails, progressDetail)

	progress := progressionState{
		SelectedConfig:  s.state.SelectedConfig,
		NodeStatus:      s.state.NodeStatus,
		ProgressDetails: s.state.ProgressDetails,
	}

	retryCurrentStep, proceedNextStep, terminateApplication := progress.SendMessage(s.state, s.log, s.connectionPool)

	s.log.Info("received cluster Upgrade plan",
		zap.String("received log", in.GetPlan()),
		zap.String("error", in.GetErr()))

	return &pb.UpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		RetryCurrentStep:     retryCurrentStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// ClusterUpgrade implements proto.Upgrade
func (s *minorUpgradeServer) ClusterUpgrade(_ context.Context, in *pb.UpgradeStatus) (*pb.UpgradeResponse, error) {

	var progressDetail progressDetails
	progressDetail.StepCount = 6
	progressDetail.Log = in.GetLog()
	progressDetail.Err = in.GetErr()

	s.state.ProgressDetails = append(s.state.ProgressDetails, progressDetail)

	progress := progressionState{
		SelectedConfig:  s.state.SelectedConfig,
		NodeStatus:      s.state.NodeStatus,
		ProgressDetails: s.state.ProgressDetails,
	}

	if !in.GetUpgradeSuccess() {
		// TODO:
		//	 handle error
	}

	retryCurrentStep, proceedNextStep, terminateApplication := progress.SendMessage(s.state, s.log, s.connectionPool)

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

// ClusterComponentRestart implements proto.Upgrade
func (s *minorUpgradeServer) ClusterComponentRestart(_ context.Context, in *pb.ComponentRestartStatus) (*pb.UpgradeResponse, error) {

	var progressDetail progressDetails
	progressDetail.StepCount = 7
	progressDetail.Log = in.GetLog()
	progressDetail.Err = in.GetErr()

	s.state.ProgressDetails = append(s.state.ProgressDetails, progressDetail)

	progress := progressionState{
		SelectedConfig:  s.state.SelectedConfig,
		NodeStatus:      s.state.NodeStatus,
		ProgressDetails: s.state.ProgressDetails,
	}

	if in.GetComponentRestartSuccess() == false {
		// TODO:
		//	 handle error
		//	 wrap error with in.GetComponent()s
	}

	retryCurrentStep, proceedNextStep, terminateApplication := progress.SendMessage(s.state, s.log, s.connectionPool)

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
