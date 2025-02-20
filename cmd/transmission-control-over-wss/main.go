package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"github.com/shishir9159/kapetanios/internal/wss"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
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
	namespace string
	log       *zap.Logger
	ctx       context.Context
	client    *orchestration.Client
}

// TODO: update -- should it be query to one vm or all the vm?

type upgrade struct {
	nefario Nefario
	mu      sync.Mutex
	upgraded chan bool
	pool    *wss.ConnectionPool
}

type upgradeProgression struct {
	CurrentStep           uint8  `yaml:"currentStep"`
	MinorUpgradeNamespace string `yaml:"minorUpgradeNamespace"`
}

// todo: should i keep track record if nodes were already
//  tainted before draining

type upgradeReport struct {
	certificateRenewal bool   `yaml:"certificateRenewal"`
	drainNodes         bool   `yaml:"drainNodes"`
	nodesUpgraded      string `yaml:"nodesUpgraded"`
	NodesToBeUpgraded  string `yaml:"nodesToBeUpgraded"`
	UbuntuK8sVersion   string `yaml:"ubuntuK8sVersion"` // currently only works with 24.02
	Redhat8K8sVersion  string `yaml:"redhat8K8sVersion"`
	Redhat9K8sVersion  string `yaml:"redhat9K8sVersion"`
}

type Server struct {
	ctx         context.Context
	waitChannel chan bool
	mu          sync.Mutex
}

func readConfig(c Nefario) (upgradeReport, error) {

	configMapName := "kapetanios"

	configMap, er := c.client.Clientset().CoreV1().ConfigMaps(c.namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
	if er != nil {
		c.log.Error("error fetching the configMap",
			zap.Error(er))
		return upgradeReport{}, er
	}

	report := upgradeReport{
		//certificateRenewal: false,
		//drainNodes:        bool(configMap.Data["DRAIN_NODES"]),
		drainNodes:        false,
		nodesUpgraded:     configMap.Data["NODES_UPGRADED"],
		NodesToBeUpgraded: configMap.Data["UBUNTU_K8S_VERSION"],
		UbuntuK8sVersion:  configMap.Data["REDHAT8_K8S_VERSION"],
		Redhat8K8sVersion: configMap.Data["REDHAT9_K8S_VERSION"],
		Redhat9K8sVersion: configMap.Data["NODES_TO_BE_UPGRADED"],
	}

	return report, nil
}

func writeConfig(c Nefario, report upgradeReport) error {

	configMapName := "kapetanios"

	configMap, er := c.client.Clientset().CoreV1().ConfigMaps(c.namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
	if er != nil {
		c.log.Error("error fetching the configMap",
			zap.Error(er))
	}

	// todo: check if value initialized
	//  default values

	configMap.Data["NODES_UPGRADED"] = report.nodesUpgraded
	configMap.Data["UBUNTU_K8S_VERSION"] = report.UbuntuK8sVersion
	configMap.Data["REDHAT8_K8S_VERSION"] = report.Redhat8K8sVersion
	configMap.Data["REDHAT9_K8S_VERSION"] = report.Redhat9K8sVersion
	configMap.Data["NODES_TO_BE_UPGRADED"] = report.NodesToBeUpgraded

	_, er = c.client.Clientset().CoreV1().ConfigMaps(c.namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if er != nil {
		c.log.Error("error updating configMap",
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

func (upgrade *upgrade) minorUpgrade(w http.ResponseWriter, r *http.Request) {
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
		log.Println("upgrade:", err)
		return
	}
	defer func(conn *websocket.Conn) {
		er := conn.Close()
		if er != nil {
			log.Println("error closing connection:", er, conn.RemoteAddr())
		}
	}(conn)

	client := &wss.Client{
		Conn: conn,
	}

	upgrade.pool.AddClient(client)
	defer upgrade.pool.RemoveClient(client)

	if upgrade.mu.TryLock() {
		defer upgrade.mu.Unlock()

		// TODO: race condition - readCtx can be cancelled


		upgrade.nefario.log.Info("registered client: ",
			zap.String("client address: ", client.Conn.RemoteAddr().String()))

		minorityReport, err := readConfig(upgrade.nefario)
		if err != nil {
			// TODO: no restart mode or draining
			log.Println("error reading config map:", err)
			//c.log.Error("could not read config map",
			//	zap.Error(err))
		}

		upgrade.upgraded = make(chan bool)
		upgrade.MinorUpgrade(minorityReport)

		return
	}

	ctx, _ := context.WithCancel(upgrade.pool.ReadCtx)
	go upgrade.pool.ReadMessageFromConn(ctx, client)
	// TODO: use the context
	// todo: broken pipe error

	<-upgrade.upgraded
}

func (nefario *Nefario) stop(w http.ResponseWriter, r *http.Request) {

	if nefario.mu.TryLock() {
		
	}

	// cleanup

	nefario.log.Info("stopping all minions and ongoing process")
	// all connection closed
	// stop channel to stop all the connections

	nefario.ctx = context.Background()
}

// TODO: lifetime - cleanup
func (server *Server) minorUpdateUpgrade(w http.ResponseWriter, r *http.Request) {

	if len(server.pool.Clients) > 5 {
		_, er := w.Write([]byte("exceeds maximum number of concurrent connections!\n quit older running tabs\n"))
		if er != nil {
			log.Println("error writing concurrent connections warning:", er)
		}
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer func(conn *websocket.Conn) {
		er := conn.Close()
		if er != nil {
			log.Println("error closing connection:", er, conn.RemoteAddr())
		}
	}(conn)

	client := &wss.Client{
		Conn: conn,
	}
	server.pool.AddClient(client)
	defer server.pool.RemoveClient(client)

	if server.mu.TryLock() {
		defer server.mu.Unlock()

		// TODO: race condition - readCtx can be cancelled

		// todo --------- the process needs to be auto started ---------
		//  ------------- and server.initialized must be true ----------
		Client, _ := orchestration.NewClient()
		cfg := zap.Config{
			Encoding:         "json",
			Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
			EncoderConfig: zapcore.EncoderConfig{
				MessageKey: "message",

				LevelKey:    "level",
				EncodeLevel: zapcore.CapitalLevelEncoder,

				TimeKey:    "time",
				EncodeTime: zapcore.ISO8601TimeEncoder,

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

		namespace := os.Getenv("KAPETANIOS_NAMESPACE")

		c := Nefario{
			log:       logger,
			client:    Client,
			namespace: namespace,
			ctx:       context.Background(),
		}
		// todo ---------

		c.log.Info("client registered: ",
			zap.String("client address: ", client.Conn.RemoteAddr().String()))

		minorityReport, err := readConfig(c)
		if err != nil {
			// TODO: no restart mode or draining
			log.Println("error reading config map:", err)
			//c.log.Error("could not read config map",
			//	zap.Error(err))
		}

		c.log.Info("entered minor upgrade")
		MinorUpgrade(server.pool, minorityReport)

		c.log.Info("minor upgrade completed")
		return
	}

	ctx, _ := context.WithCancel(server.pool.ReadCtx)
	log.Println("using read context", ctx)
	go server.pool.ReadMessageFromConn(ctx, client)
	// TODO: use the context
	// todo: channel
	// todo: broken pipe error
	time.Sleep(540 * time.Second)
}

func StartServer(ctx context.Context) {

	pool := wss.NewPool()
	// TODO: remove all the clients when the job ends
	go pool.Run()

	server := Server{
		ctx: ctx,
	}

	upgrade nefario

	http.HandleFunc("/minor-upgrade", server.minorUpdateUpgrade)
	// TODO: work on this api
	http.HandleFunc("/upgrade", server.minorUpgrade)
	http.HandleFunc("/livez", livez)

	fmt.Println("WebSocket server started on :80")

	er := http.ListenAndServe(":80", nil)
	if er != nil {
		panic(er)
	}
}

func main() {

	Client, _ := orchestration.NewClient()
	cfg := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey:    "time",
			EncodeTime: zapcore.ISO8601TimeEncoder,

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

	ctx, cancel := context.WithCancel(context.Background())

	nefario := Nefario{
		log:       logger,
		client:    Client,
		namespace: os.Getenv("KAPETANIOS_NAMESPACE"),
		ctx:       context.Background(),
	}


	report, err := readConfig(nefario)
	if err != nil {
		log.Fatal(err)
	}
	if len(report.NodesToBeUpgraded) != 0 {

	}

	// todo: resume connections after server restarts
	//  Prerequisites(minorUpgradeNamespace)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	StartServer(ctx)
}
