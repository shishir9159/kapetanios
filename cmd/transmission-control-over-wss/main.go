package main

import (
	"context"
	"flag"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"github.com/shishir9159/kapetanios/internal/wss"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"os"
	"sync"
)

var (
	port = flag.Int("port", 50051, "The server port")
	addr = flag.String("addr", "kapetanios.default.svc.cluster.local:80", "http service address")
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

type Upgrade struct {
	nefario  *Nefario
	mu       sync.Mutex
	upgraded chan bool
	pool     *wss.ConnectionPool
	config   *MinorUpgrade
	nodeInfo nodeInfo
}

type State interface {
	ReadMessage() (bool, bool, bool)
}

type Stat interface {
	placeHolderForNow()
}

type UpgradeProgression struct {
	ID        uint32       `json:"id"`
	StateName string       `json:"stateName"`
	Stat      Stat         `json:"stat"`
	State     *MinorUpdate `json:"state"`
}

// TODO:
//  https://web.archive.org/web/20160818125551/https://joneisen.tumblr.com/post/53695478114/golang-and-default-values

type MinorUpdate struct {
	State *MinorUpgrade `json:"state"`
}

// todo: should I keep track record if nodes were already
//  tainted before draining

type K8sVersion struct {
	DistroPrettyName string `json:"distroPrettyName"`
	TargetK8sVersion string `json:"targetK8sVersion"`
}

type selectedConfig struct {
	DrainNodes         bool         `json:"drainNodes"`
	PromptOnError      bool         `json:"promptOnLock"`
	CertificateRenewal bool         `json:"certificateRenewal"`
	K8sVersion         []K8sVersion `json:"k8sVersion"`
}

type k8sPackageInfo struct {
	Distro string `json:"distro"`
	DistroVersion string `json:"distroVersion"`
	DistroPrettyName string `json:"distroPrettyName"`
	RepoName string `json:"repoName"`
	RepoVersion string `json:"repoVersion"`
	AvailableVersions []string `json:"availableVersions"`

}

packageInfo := &PackageInfo{
Distro:              in.GetDistro(),
DistroVersion:       in.GetDistroVersion(), // TODO
RepoName:            in.GetRepoName(),
RepoBaseURL:         in.GetRepoBaseURL(),
CurrentK8sVersion:   in.GetCurrentVersion(),
AvailableK8sVersion: in.GetVersion(),

type MinorUpgrade struct {
	SelectedConfig         *selectedConfig `json:"selectedConfig"`
	NodesUpgraded          string          `json:"nodesUpgraded"`
	NodesToBeUpgraded      string          `json:"nodesToBeUpgraded"`
	UbuntuK8sVersion       string          `json:"ubuntuK8sVersion"` // currently only works with 24.02
	UbuntuK8sRepoName      string          `json:"ubuntuK8sRepoName"`
	UbuntuK8sRepoBaseURL   string          `json:"ubuntuK8sRepoBaseURL"`
	Rhel8K8sVersion        string          `json:"rhel8K8sVersion"`
	Rhel8K8sRepoName       string          `json:"rhel8K8sRepoName"`
	Rhel8K8sRepoBaseURL    string          `json:"rhel8K8sRepoBaseURL"`
	Rhel9K8sVersion        string          `json:"rhel9K8sVersion"`
	Rhel9K8sRepoName       string          `json:"rhel9K8sRepoName"`
	Rhel9K8sRepoBaseURL    string          `json:"rhel9K8sRepoBaseURL"`
	UbuntuTargetK8sVersion string          `json:"ubuntuTargetK8sVersion"`
	Rhel8TargetK8sVersion  string          `json:"rhel8TargetK8sVersion"`
	Rhel9TargetK8sVersion  string          `json:"rhel9TargetK8sVersion"`
	// TODO: store these too for we might restart
	UbuntuAvailableK8sVersion []string `json:"ubuntuAvailableK8sVersion"`
	Rhel8AvailableK8sVersion  []string `json:"rhel8AvailableK8sVersion"`
	Rhel9AvailableK8sVersion  []string `json:"rhel9AvailableK8sVersion"`
}

func readConfig(nefario *Nefario) (MinorUpgrade, error) {

	configMapName := "kapetanios"

	configMap, er := nefario.client.Clientset().CoreV1().ConfigMaps(nefario.namespace).
		Get(nefario.ctx, configMapName, metav1.GetOptions{})
	if er != nil {
		nefario.log.Error("error fetching the configMap",
			zap.Error(er))
		return MinorUpgrade{}, er
	}

	config := MinorUpgrade{
		SelectedConfig: &selectedConfig{
			DrainNodes:         false,
			PromptOnError:      false,
			CertificateRenewal: false,
			K8sVersion:         nil,
		},
		NodesUpgraded:          configMap.Data["NODES_UPGRADED"],
		NodesToBeUpgraded:      configMap.Data["NODES_TO_BE_UPGRADED"],
		UbuntuK8sVersion:       configMap.Data["UBUNTU_K8S_VERSION"],
		Rhel8K8sVersion:        configMap.Data["REDHAT8_K8S_VERSION"],
		Rhel9K8sVersion:        configMap.Data["REDHAT9_K8S_VERSION"],
		UbuntuTargetK8sVersion: configMap.Data["UBUNTU_TARGET_K8S_VERSION"],
		Rhel8TargetK8sVersion:  configMap.Data["RHEL8_TARGET_K8S_VERSION"],
		Rhel9TargetK8sVersion:  configMap.Data["RHEL9_TARGET_K8S_VERSION"],
	}

	return config, nil
}

func writeConfig(nefario *Nefario, report MinorUpgrade) error {

	configMapName := "kapetanios"

	configMap, er := nefario.client.Clientset().CoreV1().ConfigMaps(nefario.namespace).
		Get(context.Background(), configMapName, metav1.GetOptions{})
	if er != nil {
		nefario.log.Error("error fetching the configMap",
			zap.Error(er))
	}

	// TODO: duct taped for now
	//  make a function that uses reflect to put the values
	//  or use a json on the configMap

	configMap.Data["NODES_UPGRADED"] = report.NodesUpgraded
	configMap.Data["NODES_TO_BE_UPGRADED"] = report.NodesToBeUpgraded
	configMap.Data["UBUNTU_K8S_VERSION"] = report.UbuntuK8sVersion
	configMap.Data["REDHAT8_K8S_VERSION"] = report.Rhel8K8sVersion
	configMap.Data["REDHAT9_K8S_VERSION"] = report.Rhel9K8sVersion
	configMap.Data["UBUNTU_TARGET_K8S_VERSION"] = report.UbuntuTargetK8sVersion
	configMap.Data["RHEL8_TARGET_K8S_VERSION"] = report.Rhel8TargetK8sVersion
	configMap.Data["RHEL9_TARGET_K8S_VERSION"] = report.Rhel9TargetK8sVersion

	_, er = nefario.client.Clientset().CoreV1().ConfigMaps(nefario.namespace).
		Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if er != nil {
		nefario.log.Error("error updating configMap",
			zap.Error(er))
	}

	nefario.log.Info("writing config",
		zap.String("nodes upgraded", report.NodesUpgraded),
		zap.String("nodes to be upgraded", report.NodesToBeUpgraded))

	return nil
}

// healthz is a liveness probe.
func livez(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("shuttle", "launched")
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

// TODO: lifetime - cleanup

func (nefario *Nefario) report(w http.ResponseWriter, _ *http.Request) {
	config, err := readConfig(nefario)
	if err != nil {
		return
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	payload, err := json.Marshal(&config)
	if err != nil {
		panic(err)
	}

	w.Write(payload)
}

func (upgrade *Upgrade) minorUpgrade(w http.ResponseWriter, r *http.Request) {

	//var Json = jsoniter.ConfigFastest
	//decoder := Json.NewDecoder(r.Body)
	//var t upgradeProgression
	//err := decoder.Decode(&t)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusBadRequest)
	//	panic(err)
	//}
	//
	//log.Println(t)

	if len(upgrade.pool.Clients) > 5 {
		io.WriteString(w, "maximum number of concurrent connection exceeded!")

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

	// todo: add on prerequisite
	//  if upgrade.nefario.mu.TryLock() {
	//	 // only allow one api would run at a time
	//	 defer upgrade.nefario.mu.Unlock()
	//}

	if upgrade.mu.TryLock() {
		//defer upgrade.mu.Unlock()

		// TODO: race condition - readCtx can be cancelled

		upgrade.nefario.log.Info("registered client: ",
			zap.String("client address: ", client.Conn.RemoteAddr().String()))

		config, er := readConfig(upgrade.nefario)
		if er != nil {
			upgrade.nefario.log.Info("error reading config",
				zap.Error(er))
		}

		upgrade.config = &config
		upgrade.upgraded = make(chan bool, 1)
		upgrade.MinorUpgrade()

		return
	}

	ctx, _ := context.WithCancel(upgrade.pool.ReadCtx)
	go upgrade.pool.ReadMessageFromConn(ctx, client)

	<-upgrade.upgraded

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

	upgrade := Upgrade{
		pool:    pool,
		config:  &config,
		nefario: &nefario,
	}
	nefario.log.Debug("nodes to be upgraded",
		zap.String("nodesToBeUpgraded", upgrade.config.NodesToBeUpgraded))

	// TODO:
	//  read the configMap on how to
	//  upgrade using selected Config

	if upgrade.config.NodesToBeUpgraded != "" {
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
