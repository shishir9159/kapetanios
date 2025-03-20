package main

import (
	"context"
	"flag"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"github.com/shishir9159/kapetanios/internal/wss"
	//_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"os"
	"sync"
)

var (
	port = flag.Int("port", 50051, "the server port")
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:   0,
	WriteBufferSize:  0,
	WriteBufferPool:  nil,
	HandshakeTimeout: 0,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Error:             nil,
	Subprotocols:      nil,
	EnableCompression: false,
}

type nodeInfo struct {
	os                string
	kernelVersion     string
	kubernetesVersion string
}

type Nefario struct {
	client    *orchestration.Client
	ctx       context.Context
	cancel    context.CancelFunc
	log       *zap.Logger
	mu        sync.Mutex
	namespace string
}

type MinorUpgrade struct {
	mu           sync.Mutex
	finalizer    chan bool
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

// todo: should I keep track record if nodes were already
//  tainted before draining TODO: set a specific label
//  if you want to avoid upgrade

type K8sVersion struct {
	DistroPrettyName string `json:"distroPrettyName"`
	TargetK8sVersion string `json:"targetK8sVersion"`
}

type selectedConfig struct {
	DrainNodes         bool         `json:"drainNodes"`
	PromptOnError      bool         `json:"promptOnError"`
	CertificateRenewal bool         `json:"certificateRenewal"`
	K8sVersion         []K8sVersion `json:"k8sVersion"`
}

// only RHEL8, RHEL9, Ubuntu 24.02 supported for now
type k8sPackageInfo struct {
	Distro            string   `json:"distro"`
	DistroVersion     string   `json:"distroVersion"`
	DistroPrettyName  string   `json:"distroPrettyName"`
	RepoName          string   `json:"repoName"`
	RepoBaseURL       string   `json:"repoBaseUrl"`
	AvailableVersions []string `json:"availableVersions"`
}

type nodesStatus struct {
	NodesUpgraded           []string `json:"nodesUpgraded"`
	NodesToBeUpgraded       []string `json:"nodesToBeUpgraded"`
	NodesFailedToBeUpgraded []string `json:"nodesFailedToBeUpgraded"`
}

type progressDetail struct {
	StepCount uint64 `json:"stepCount"`
	Log       string `json:"log"`
	Err       string `json:"err"`
}

type UpgradeState struct {
	SelectedConfig  *selectedConfig  `json:"selectedConfig"`
	NodeStatus      *nodesStatus     `json:"nodesStatus"`
	Nodes           []nodeInfo       `json:"nodes"`
	K8sPackages     []k8sPackageInfo `json:"k8sPackages"`
	ProgressDetails []progressDetail `json:"progressDetails"`
}

func readConfig(nefario *Nefario) (UpgradeState, error) {

	configMapName := "kapetanios"

	configMap, err := nefario.client.Clientset().CoreV1().ConfigMaps(nefario.namespace).
		Get(nefario.ctx, configMapName, metav1.GetOptions{})
	if err != nil {
		nefario.log.Error("error fetching the configMap",
			zap.Error(err))
		return UpgradeState{}, err
	}

	state := UpgradeState{}

	var json = jsoniter.ConfigFastest

	err = json.Unmarshal([]byte(configMap.Data["UPGRADE_STATE"]), &state)
	if err != nil {
		nefario.log.Error("error unmarshaling the configMap")
	}

	if state.NodeStatus == nil {
		state.NodeStatus = &nodesStatus{
			NodesToBeUpgraded: []string{},
		}
	}

	return state, nil
}

func writeConfig(nefario *Nefario, state UpgradeState) error {

	configMapName := "kapetanios"

	configMap, er := nefario.client.Clientset().CoreV1().ConfigMaps(nefario.namespace).
		Get(context.Background(), configMapName, metav1.GetOptions{})
	if er != nil {
		nefario.log.Error("error fetching the configMap",
			zap.Error(er))
	}

	var json = jsoniter.ConfigFastest

	payload, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		nefario.log.Error("error marshaling the configMap")
	}

	configMap.Data["UPGRADE_STATE"] = string(payload)

	_, er = nefario.client.Clientset().CoreV1().ConfigMaps(nefario.namespace).
		Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if er != nil {
		nefario.log.Error("error updating configMap",
			zap.Error(er))
	}

	return nil
}

// healthz is a liveness probe.
func livez(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("shuttle", "launched")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

//func readyz(isReady *atomic.Value) http.HandlerFunc {
//	return func(w http.ResponseWriter, _ *http.Request) {
//		if isReady == nil || !isReady.Load().(bool) {
//			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
//			return
//		}
//		w.WriteHeader(http.StatusOK)
//	}
//}

func (upgrade *MinorUpgrade) minorUpgrade(w http.ResponseWriter, r *http.Request) {

	if len(upgrade.pool.Clients) > 5 {
		_, err := io.WriteString(w, "maximum number of concurrent connection exceeded!")
		if err != nil {

		}

		upgrade.nefario.log.Info("exceeds maximum number of concurrent connections!\n quit older running tabs\n")
		// TODO: -- wrap it in one of the error state(possible other: certificate renewal) --
		_, er := w.Write([]byte("exceeds maximum number of concurrent connections!\n quit older running tabs\n"))
		if er != nil {
			upgrade.nefario.log.Info("error writing concurrent connections warning:",
				zap.Error(er))
		}

		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// TODO: wrap it in error state
		upgrade.nefario.log.Info("connection upgrade error:",
			zap.Error(err))
		return
	}

	client := &wss.Client{
		Conn: conn,
	}

	upgrade.pool.AddClient(client)

	upgrade.nefario.log.Info("registered client: ",
		zap.String("client address: ", client.Conn.RemoteAddr().String()))

	// todo: add on prerequisite
	//  if upgrade.nefario.mu.TryLock() {
	//	 // only allow one api would run at a time
	//	 defer upgrade.nefario.mu.Unlock()
	//}

	if upgrade.mu.TryLock() {
		//defer upgrade.mu.Unlock()

		// TODO: race condition - readCtx can be cancelled

		config, er := readConfig(upgrade.nefario)
		if er != nil {
			upgrade.nefario.log.Info("error reading config",
				zap.Error(er))
		}

		upgrade.upgradeState = &config
		upgrade.finalizer = make(chan bool, 1)
		upgrade.MinorUpgrade()

		return
	}

	ctx, _ := context.WithCancel(upgrade.pool.ReadCtx)
	go upgrade.pool.ReadMessageFromConn(ctx, client)

	<-upgrade.finalizer

	// can fail
	upgrade.pool.CancelReadContext()

	// TODO: should there be any mutex for removing clients?
	for client = range upgrade.pool.Clients {
		upgrade.pool.RemoveClient(client)
	}

}

//func (nefario *Nefario) stop(w http.ResponseWriter, r *http.Request) {
//
//	if nefario.mu.TryLock() {
//
//	}
//
//	// cleanup
//
//	nefario.log.Info("stopping all minions and ongoing process")
//	// all connection closed
//	// stop channel to stop all the connections
//
//	nefario.ctx = context.Background()
//}

func main() {

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

	nefario := Nefario{
		client:    Client,
		ctx:       ctx,
		cancel:    cancel,
		namespace: os.Getenv("KAPETANIOS_NAMESPACE"),
		log:       logger,
	}

	// TODO: refactor read write,
	//  be conservative what you send
	config, err := readConfig(&nefario)
	if err != nil {
		nefario.log.Info("failed to read config map: ",
			zap.Error(err))
	}

	pool := wss.NewPool()
	// TODO: remove all the clients when the job ends
	go pool.Run(nefario.log)

	upgrade := MinorUpgrade{
		pool:         pool,
		upgradeState: &config,
		nefario:      &nefario,
	}
	nefario.log.Debug("nodes to be upgraded",
		zap.Strings("nodesToBeUpgraded", upgrade.upgradeState.NodeStatus.NodesToBeUpgraded))

	if len(upgrade.upgradeState.NodeStatus.NodesToBeUpgraded) != 0 {
		er := Prerequisites(&upgrade)
		if er != nil {
			upgrade.nefario.log.Info("error writing prerequisites warning:",
				zap.Error(er))
		}
	}

	http.HandleFunc("/minor-upgrade", upgrade.minorUpgrade)
	http.HandleFunc("/livez", livez)

	upgrade.nefario.log.Info("starting kapetanios server on :80")
	err = http.ListenAndServe(":80", nil)
	if err != nil {
		upgrade.nefario.log.Error("error starting kapetanios server",
			zap.Error(err))
		panic(err)
	}
}
