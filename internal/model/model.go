package model

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	v1 "k8s.io/api/core/v1"
	"kapetanios/internal/orchestration"
	"kapetanios/internal/wss"
	"os"
	"sync"
	"time"
)

type Cluster struct {
	Nodes    v1.NodeList `nodes:"type:json"`
	Validity time.Time   `expiration-date:"type:json"`
}

type NodeInfo struct {
	OS                string
	KernelVersion     string
	KubernetesVersion string
}

type State interface {
	SendMessage(upgradeState *UpgradeState, log *zap.Logger, pool *wss.ConnectionPool) (bool, bool, bool)
}

func (u *UpgradeConfigSelection) SendMessage(upgradeState *UpgradeState, log *zap.Logger, pool *wss.ConnectionPool) (bool, bool, bool) {

	upgradeProgression := &UpgradeProgression{
		StateName: "upgrade config selection",
		State:     u,
	}

	payload, err := json.Marshal(upgradeProgression)
	if err != nil {
		log.Error("failed to marshal upgrade progression",
			zap.Error(err))
	}

	for {
		log.Info("broadcasting payload for available package version")
		pool.BroadcastMessage(payload)

		response, er := pool.ReadMessages()
		if er != nil {
			log.Info("failed to read payload for available package version",
				zap.Error(er))
		}

		selections := SelectedConfig{}
		er = json.Unmarshal([]byte(response), &selections)

		if er != nil || selections.K8sVersion == nil {
			log.Error("failed to unmarshal selected upgrade config",
				zap.Error(er))
			continue
		}

		upgradeState.SelectedConfig = &selections
		break
	}

	// TODO:
	//  HANDLE THE ERROR
	//  error: rpmdb: BDB0113 Thread/process 1380430/139742582999936 failed: BDB1507 Thread died in Berkeley DB library
	//  error: db5 error(-30973) from dbenv->failchk: BDB0087 DB_RUNRECOVERY: Fatal error, run database recovery
	//  error: cannot open Packages index using db5 -  (-30973)
	//  error: cannot open Packages database in /var/lib/rpm
	//  Error: Error: rpmdb open failed

	return false, false, false
}

type UpgradeProgression struct {
	StateName string `json:"stateName"`
	State     State  `json:"state"`
}

// todo: should I keep track record if nodes were already
//  tainted before draining
// TODO: set a specific label
//  if you want to avoid upgrade

type K8sVersion struct {
	DistroPrettyName string `json:"distroPrettyName"`
	TargetK8sVersion string `json:"targetK8sVersion"`
}

type SelectedConfig struct {
	DrainNodes bool `json:"drainNodes"`
	// TODO:
	//  Allow different repo for the same OS
	PromptOnError      bool         `json:"promptOnError"`
	CertificateRenewal bool         `json:"certificateRenewal"`
	K8sVersion         []K8sVersion `json:"k8sVersion"`
}

type K8sPackageInfo struct {
	Distro            string   `json:"distro"`
	DistroVersion     string   `json:"distroVersion"`
	DistroPrettyName  string   `json:"distroPrettyName"`
	RepoName          string   `json:"repoName"`
	RepoBaseURL       string   `json:"repoBaseUrl"`
	AvailableVersions []string `json:"availableVersions"`
}

type NodesStatus struct {
	NodesUpgraded            []string `json:"nodesUpgraded"`
	NodesToBeUpgraded        []string `json:"nodesToBeUpgraded"`
	NodesFailedToBeUpgraded  []string `json:"nodesFailedToBeUpgraded"`
	NodesSkippedToBeUpgraded []string `json:"nodesSkippedToBeUpgraded"`
}

type ProgressDetail struct {
	StepCount uint64 `json:"stepCount"`
	Log       string `json:"log"`
	Err       string `json:"err"`
}

type UpgradeState struct {
	SelectedConfig  *SelectedConfig  `json:"selectedConfig"`
	NodesStatus     *NodesStatus     `json:"nodesStatus"`
	Nodes           []NodeInfo       `json:"nodes"`
	K8sPackages     []K8sPackageInfo `json:"k8sPackages"`
	ProgressDetails []ProgressDetail `json:"progressDetails"`
}

type DBMS interface {
	InitConnection() error
	//	ReadConfig
}

type HandlerLock struct {
	Mu          sync.Mutex
	HandlerName string
}

type Nefario struct {
	Client      *orchestration.Client
	Ctx         context.Context
	Cancel      context.CancelFunc
	Log         *zap.Logger
	HandlerLock HandlerLock
	// configMap as DBMS is a conspiracy theory!!!
	DBMS      DBMS
	Namespace string
}

// TODO:
//  reverse

type UpgradeConfigSelection struct {
	K8sPackagesInfo []K8sPackageInfo `json:"k8sPackagesInfo"`
}

func NewNefario() *Nefario {

	cfg := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "message",
			LevelKey:     "level",
			EncodeLevel:  zapcore.CapitalLevelEncoder,
			TimeKey:      "time",
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	logger := zap.Must(cfg.Build())
	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {
			logger.Info("error syncing logger before application terminates",
				zap.Error(er))
		}
	}(logger)

	Client, err := orchestration.NewClient()
	if err != nil {
		logger.Error("error creating orchestration client",
			zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Nefario{
		Client:    Client,
		Ctx:       ctx,
		Cancel:    cancel,
		Namespace: os.Getenv("KAPETANIOS_NAMESPACE"),
		Log:       logger,
	}
}
