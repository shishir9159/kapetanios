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
type minorUpgradeServer struct {
	log            *zap.Logger
	upgradeConfig  *upgradeConfig
	connectionPool *wss.ConnectionPool
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

// TODO:
//  pass a generic???
//  receive a generic too???

func actionResponse(payload []byte, log *zap.Logger, pool *wss.ConnectionPool) (bool, bool, bool, error) {

	pool.BroadcastMessage(payload)
	response, err := pool.ReadMessages()
	if err != nil {
		log.Error("Error reading response",
			zap.Error(err))
		// return false, false, false, err
	}

	switch response {
	case "retry":
		return true, false, false, nil
	case "next step":
		return false, true, true, nil
	case "terminate application":
		return false, false, true, nil
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

	// TODO: this should not happen as there is no
	//  break statement in the switch
	return false, false, false, nil
}

// TODO: Create an interface instead of generic that
//  takes the structs write them in the connection
//  and wait for reading
// TODO: state id ---------

// ClusterHealthChecking implements proto.MinorUpgradeServer
func (s *minorUpgradeServer) ClusterHealthChecking(_ context.Context, in *pb.PrerequisitesMinorUpgrade) (*pb.UpgradeResponse, error) {

	var retryCurrentStep, proceedNextStep, terminateApplication = false, false, false

	nodeHealth := clusterHealth{
		EtcdStatus:          in.GetEtcdStatus(),
		StorageAvailability: in.GetStorageAvailability(),
		Err:                 in.GetErr(),
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
		zap.Bool("etcd status", in.GetEtcdStatus()),
		zap.Uint64("received storage availability", in.GetStorageAvailability()),
		zap.String("received error", in.GetErr()))

	return &pb.UpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		RetryCurrentStep:     retryCurrentStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// UpgradeVersionSelection implements proto.MinorUpgradeServer
func (s *minorUpgradeServer) UpgradeVersionSelection(_ context.Context, in *pb.AvailableVersions) (*pb.ClusterUpgradeResponse, error) {

	var retryCurrentStep, proceedNextStep, terminateApplication = false, false, false

	versions := availableVersion{
		CurrentVersion:   "",
		AvailableVersion: in.GetVersion(),
		Err:              in.GetErr(),
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

// ClusterCompatibility implements proto.Upgrade
func (s *minorUpgradeServer) ClusterCompatibility(_ context.Context, in *pb.UpgradeCompatibility) (*pb.UpgradeResponse, error) {

	var retryCurrentStep, proceedNextStep, terminateApplication = false, false, false

	compatability := clusterCompatability{
		OSCompatability: in.GetOsCompatibility(),
		Diff:            in.GetDiff(),
		Err:             in.GetErr(),
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	payload, err := json.Marshal(&compatability)
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

	var retryCurrentStep, proceedNextStep, terminateApplication = false, false, false

	componentUpgrade := clusterComponentUpgrade{
		ComponentUpgradeSuccess: in.ComponentUpgradeSuccess,
		Component:               in.GetComponent(),
		Log:                     in.GetLog(),
		Err:                     in.GetErr(),
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	payload, err := json.Marshal(&componentUpgrade)
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

	var retryCurrentStep, proceedNextStep, terminateApplication = false, false, false

	upgradePlan := clusterUpgradePlan{
		ClusterVersion: in.GetCurrentClusterVersion(),
		Log:            in.GetLog(),
		Err:            in.GetErr(),
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	payload, err := json.Marshal(&upgradePlan)
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

	s.log.Info("received cluster Upgrade plan",
		zap.String("cluster version", in.GetCurrentClusterVersion()),
		zap.String("received log", in.GetLog()),
		zap.String("error", in.GetErr()))

	return &pb.UpgradeResponse{
		ProceedNextStep:      proceedNextStep,
		RetryCurrentStep:     retryCurrentStep,
		TerminateApplication: terminateApplication,
	}, nil
}

// ClusterUpgrade implements proto.Upgrade
func (s *minorUpgradeServer) ClusterUpgrade(_ context.Context, in *pb.UpgradeStatus) (*pb.UpgradeResponse, error) {

	var retryCurrentStep, proceedNextStep, terminateApplication = false, false, false

	upgradeSuccess := clusterUpgradeSuccess{
		UpgradeSuccess: in.GetUpgradeSuccess(),
		Log:            in.GetLog(),
		Err:            in.GetErr(),
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	payload, err := json.Marshal(&upgradeSuccess)
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

	s.log.Info("received cluster Upgrade plan",
		zap.Bool("cluster Upgrade status", in.GetUpgradeSuccess()),
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

	var retryCurrentStep, proceedNextStep, terminateApplication = false, false, false

	restartSuccess := componentRestartSuccess{
		ComponentRestartSuccess: false,
		Component:               "",
		Log:                     "",
		Err:                     "",
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	payload, err := json.Marshal(&restartSuccess)
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

func MinorUpgradeGrpc(log *zap.Logger, pool *wss.ConnectionPool, ch chan<- *grpc.Server) {

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
