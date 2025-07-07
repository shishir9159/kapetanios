package model

import (
	"context"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"github.com/shishir9159/kapetanios/internal/wss"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	v1 "k8s.io/api/core/v1"
	"os"
	"sync"
	"time"
)

type Cluster struct {
	Nodes    v1.NodeList `nodes:"type:json"`
	Validity time.Time   `expiration-date:"type:json"`
}

type nodeInfo struct {
	os                string
	kernelVersion     string
	kubernetesVersion string
}

type MinorUpgrade struct {
	mu           sync.Mutex
	finalizer    chan bool
	// reverse the dependency
	nefario      *Nefario
	upgradeState *UpgradeState
	pool         *wss.ConnectionPool
}

type State interface {
	SendMessage(upgradeState *UpgradeState, log *zap.Logger, pool *wss.ConnectionPool) (bool, bool, bool)
}

type UpgradeProgression struct {
	StateName string `json:"stateName"`
	State     State  `json:"state"`
}

type K8sVersion struct {
	DistroPrettyName string `json:"distroPrettyName"`
	TargetK8sVersion string `json:"targetK8sVersion"`
}

type selectedConfig struct {
	DrainNodes bool `json:"drainNodes"`
	// TODO:
	//  Allow different repo for the same OS
	PromptOnError      bool         `json:"promptOnError"`
	CertificateRenewal bool         `json:"certificateRenewal"`
	K8sVersion         []K8sVersion `json:"k8sVersion"`
}

type k8sPackageInfo struct {
	Distro            string   `json:"distro"`
	DistroVersion     string   `json:"distroVersion"`
	DistroPrettyName  string   `json:"distroPrettyName"`
	RepoName          string   `json:"repoName"`
	RepoBaseURL       string   `json:"repoBaseUrl"`
	AvailableVersions []string `json:"availableVersions"`
}

type nodesStatus struct {
	NodesUpgraded            []string `json:"nodesUpgraded"`
	NodesToBeUpgraded        []string `json:"nodesToBeUpgraded"`
	NodesFailedToBeUpgraded  []string `json:"nodesFailedToBeUpgraded"`
	NodesSkippedToBeUpgraded []string `json:"nodesSkippedToBeUpgraded"`
}

type progressDetail struct {
	StepCount uint64 `json:"stepCount"`
	Log       string `json:"log"`
	Err       string `json:"err"`
}

type UpgradeState struct {
	SelectedConfig  *selectedConfig  `json:"selectedConfig"`
	NodesStatus     *nodesStatus     `json:"nodesStatus"`
	Nodes           []nodeInfo       `json:"nodes"`
	K8sPackages     []k8sPackageInfo `json:"k8sPackages"`
	ProgressDetails []progressDetail `json:"progressDetails"`
}


type DBMS interface {
	InitConnection() error
}

type Nefario struct {
	Client *orchestration.Client
	Ctx    context.Context
	Cancel context.CancelFunc
	Log    *zap.Logger
	Mu     sync.Mutex
	// cyclic imports
	DBMS      DBMS
	Namespace string
}

// TODO:
//  reverse

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


type upgradeConfigSelection struct {
	K8sPackagesInfo []k8sPackageInfo `json:"k8sPackagesInfo"`

