package main

import (
	"context"
	"flag"
	"github.com/gorilla/websocket"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"github.com/shishir9159/kapetanios/internal/wss"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
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
}

// should it be foreman?

type Nefario struct {
	client    *orchestration.Client
	ctx       context.Context
	cancel    context.CancelFunc
	log       *zap.Logger
	mu        sync.Mutex
	namespace string
}

// TODO: update -- should it be query to one vm or all the vm?

type Upgrade struct {
	nefario  *Nefario
	mu       sync.Mutex
	upgraded chan bool
	pool     *wss.ConnectionPool
	config   *upgradeConfig
}

type upgradeProgression struct {
	CurrentStep           uint8  `yaml:"currentStep"`
	MinorUpgradeNamespace string `yaml:"minorUpgradeNamespace"`
}

// todo: should I keep track record if nodes were already
//  tainted before draining

type upgradeConfig struct {
	certificateRenewal bool   `yaml:"certificateRenewal"`
	drainNodes         bool   `yaml:"drainNodes"`
	nodesUpgraded      string `yaml:"nodesUpgraded"`
	NodesToBeUpgraded  string `yaml:"nodesToBeUpgraded"`
	UbuntuK8sVersion   string `yaml:"ubuntuK8sVersion"` // currently only works with 24.02
	Redhat8K8sVersion  string `yaml:"redhat8K8sVersion"`
	Redhat9K8sVersion  string `yaml:"redhat9K8sVersion"`
}

func readConfig(nefario *Nefario) (upgradeConfig, error) {

	configMapName := "kapetanios"

	configMap, er := nefario.client.Clientset().CoreV1().ConfigMaps(nefario.namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
	if er != nil {
		nefario.log.Error("error fetching the configMap",
			zap.Error(er))
		return upgradeConfig{}, er
	}

	report := upgradeConfig{
		certificateRenewal: false,
		drainNodes:         false,
		//drainNodes:        bool(configMap.Data["DRAIN_NODES"]),
		nodesUpgraded:     configMap.Data["NODES_UPGRADED"],
		NodesToBeUpgraded: configMap.Data["NODES_TO_BE_UPGRADED"],
		UbuntuK8sVersion:  configMap.Data["UBUNTU_K8S_VERSION"],
		Redhat8K8sVersion: configMap.Data["REDHAT8_K8S_VERSION"],
		Redhat9K8sVersion: configMap.Data["REDHAT9_K8S_VERSION"],
	}

	nefario.log.Info("reading config",
		zap.String("nodes upgraded", configMap.Data["NODES_UPGRADED"]),
		zap.String("ubuntu k8s version", configMap.Data["UBUNTU_K8S_VERSION"]),
		zap.String("redhat k8s version", configMap.Data["REDHAT8_K8S_VERSION"]),
		zap.String("redhat 9 k8s version", configMap.Data["REDHAT9_K8S_VERSION"]),
		zap.String("nodes to be upgraded", configMap.Data["NODES_TO_BE_UPGRADED"]))

	nefario.log.Info("updated config",
		zap.String("nodes upgraded", report.nodesUpgraded),
		zap.String("ubuntu k8s version", report.UbuntuK8sVersion),
		zap.String("redhat k8s version", report.Redhat8K8sVersion),
		zap.String("redhat 9 k8s version", report.Redhat9K8sVersion),
		zap.String("nodes to be upgraded", report.NodesToBeUpgraded))

	return report, nil
}

func writeConfig(nefario *Nefario, report upgradeConfig) error {

	configMapName := "kapetanios"

	configMap, er := nefario.client.Clientset().CoreV1().ConfigMaps(nefario.namespace).
		Get(context.Background(), configMapName, metav1.GetOptions{})
	if er != nil {
		nefario.log.Error("error fetching the configMap",
			zap.Error(er))
	}

	// todo: check if value initialized
	//  default values

	configMap.Data["NODES_UPGRADED"] = report.nodesUpgraded
	configMap.Data["UBUNTU_K8S_VERSION"] = report.UbuntuK8sVersion
	configMap.Data["REDHAT8_K8S_VERSION"] = report.Redhat8K8sVersion
	configMap.Data["REDHAT9_K8S_VERSION"] = report.Redhat9K8sVersion
	configMap.Data["NODES_TO_BE_UPGRADED"] = report.NodesToBeUpgraded

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
		_, er := w.Write([]byte("exceeds maximum number of concurrent connections!\n quit older running tabs\n"))
		if er != nil {
			upgrade.nefario.log.Info("error writing concurrent connections warning:",
				zap.Error(er))
		}

		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		upgrade.nefario.log.Info("connection upgrade error:",
			zap.Error(err))
		return
	}

	client := &wss.Client{
		Conn: conn,
	}

	upgrade.pool.AddClient(client)
	defer upgrade.pool.RemoveClient(client)

	// todo: add on prerequisite
	if upgrade.nefario.mu.TryLock() {
		// only allow one api would run at a time
		defer upgrade.nefario.mu.Unlock()
	}

	if upgrade.mu.TryLock() {
		//defer upgrade.mu.Unlock()

		// TODO: race condition - readCtx can be cancelled

		upgrade.nefario.log.Info("registered client: ",
			zap.String("client address: ", client.Conn.RemoteAddr().String()))

		config, err := readConfig(upgrade.nefario)
		if err != nil {
			// TODO: no restart mode or draining
			log.Println("error reading config map:", err)
			//c.log.Error("could not read config map",
			//	zap.Error(err))
		}

		upgrade.config = &config
		upgrade.upgraded = make(chan bool, 1)
		upgrade.MinorUpgrade()

		return
	}

	ctx, _ := context.WithCancel(upgrade.pool.ReadCtx)
	go upgrade.pool.ReadMessageFromConn(ctx, client)
	// TODO: use the context
	// todo: broken pipe error

	<-upgrade.upgraded
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

	if len(upgrade.config.NodesToBeUpgraded) != 0 {
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
