package model

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/v2/bson"
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
	OS                string `json:"os" bson:"os"`
	KernelVersion     string `json:"kernelVersion" bson:"kernelVersion"`
	KubernetesVersion string `json:"kubernetesVersion" bson:"kubernetesVersion"`
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
	PromptOnError      bool         `json:"promptOnError" bson:"promptOnError"`
	CertificateRenewal bool         `json:"certificateRenewal" bson:"certificateRenewal"`
	K8sVersion         []K8sVersion `json:"k8sVersion" bson:"k8sVersion"`
}

type K8sPackageInfo struct {
	Distro            string   `json:"distro" bson:"distro"`
	DistroVersion     string   `json:"distroVersion" bson:"distroVersion"`
	DistroPrettyName  string   `json:"distroPrettyName" bson:"distroPrettyName"`
	RepoName          string   `json:"repoName" bson:"repoName"`
	RepoBaseURL       string   `json:"repoBaseUrl" bson:"repoBaseUrl"`
	AvailableVersions []string `json:"availableVersions" bson:"availableVersions"`
}

type NodesStatus struct {
	NodesUpgraded            []string `json:"nodesUpgraded" bson:"nodesUpgraded"`
	NodesToBeUpgraded        []string `json:"nodesToBeUpgraded" bson:"nodesToBeUpgraded"`
	NodesFailedToBeUpgraded  []string `json:"nodesFailedToBeUpgraded" bson:"nodesFailedToBeUpgraded"`
	NodesSkippedToBeUpgraded []string `json:"nodesSkippedToBeUpgraded" bson:"nodesSkippedToBeUpgraded"`
}

type ProgressDetail struct {
	StepCount uint64 `json:"stepCount" bson:"stepCount"`
	Log       string `json:"log" bson:"log"`
	Err       string `json:"err" bson:"err"`
}

type NodeMinorUpgradeHistory struct {
	Name           string           `bson:"name"`
	LastUpdated    time.Time        `bson:"lastUpdated"`
	UpdatedVersion string           `bson:"updatedVersion"`
	ProgressDetail []ProgressDetail `bson:"progressDetail"`
}

type ClusterMinorUpgradeHistory struct {
	ID                       bson.ObjectID             `bson:"_id"`
	NodesMinorUpgradeHistory []NodeMinorUpgradeHistory `json:"nodeMinorUpgradeHistory" bson:"nodesMinorUpgradeHistory"`
}

type UpgradeState struct {
	SelectedConfig  *SelectedConfig  `json:"selectedConfig" bson:"selectedConfig"`
	NodesStatus     *NodesStatus     `json:"nodesStatus" bson:"nodesStatus"`
	Nodes           []NodeInfo       `json:"nodes" bson:"nodes"`
	K8sPackages     []K8sPackageInfo `json:"k8sPackages" bson:"k8sPackages"`
	ProgressDetails []ProgressDetail `json:"progressDetails" bson:"progressDetails"`
}

type KubernetesCertificatesExpiration struct {
	LastUpdated          time.Time
	NodesCertsExpiration []NodeCertsExpiration
}

type NodeCertsExpiration struct {
	NodeName              string
	MissingCertificate    []string
	CertificateExpiration []CertsExpiration
	CertificatesAuthority []CertsAuthority
}

type CertsExpiration struct {
	Name                 string
	Expires              string
	ResidualTime         string
	CertificateAuthority string
	ExternallyManaged    string
}

type CertsAuthority struct {
	Name              string
	Expires           string
	ResidualTime      string
	ExternallyManaged string
}

type DBMS interface {
	InitConnection() error
	ReadConfig() (UpgradeState, error)
	WriteConfig(UpgradeState) error
	MinorUpgradeReport() error
	MinorUpgradeReportUpdate(NodeMinorUpgradeHistory) error
	LastMinorUpgradeReport() (ClusterMinorUpgradeHistory, error)
	CertificateReportUpdate(KubernetesCertificatesExpiration) error
	LastCertificateReport() (KubernetesCertificatesExpiration, error)
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
	DBMS        DBMS
	Namespace   string
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
