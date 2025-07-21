package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "kapetanios/internal/model"
	. "kapetanios/internal/model/mongo"
	"kapetanios/internal/wss"
	"net/http"
	"os"
	"sync"
)

var (
	port           = flag.Int("port", 50051, "the server port")
	gzipWriterPool = sync.Pool{
		New: func() any {
			return gzip.NewWriter(nil)
		},
	}
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

// todo: should I keep track record if nodes were already
//  tainted before draining
// TODO: set a specific label
//  if you want to avoid upgrade

// TODO:
func readConfig(nefario *Nefario) (UpgradeState, error) {

	configMapName := "kapetanios"

	configMap, err := nefario.Client.Clientset().CoreV1().ConfigMaps(nefario.Namespace).
		Get(nefario.Ctx, configMapName, metav1.GetOptions{})
	if err != nil {
		nefario.Log.Error("error fetching the configMap",
			zap.Error(err))
		return UpgradeState{}, err
	}

	state := UpgradeState{}

	err = json.Unmarshal([]byte(configMap.Data["UPGRADE_STATE"]), &state)
	if err != nil {
		nefario.Log.Error("error unmarshalling the configMap",
			zap.Error(err))
	}

	if state.NodesStatus == nil {
		state.NodesStatus = &NodesStatus{
			NodesToBeUpgraded: []string{},
		}
	}

	return state, nil
}

func writeConfig(nefario *Nefario, state UpgradeState) error {

	configMapName := "kapetanios"

	configMap, er := nefario.Client.Clientset().CoreV1().ConfigMaps(nefario.Namespace).
		Get(context.Background(), configMapName, metav1.GetOptions{})
	if er != nil {
		nefario.Log.Error("error fetching the configMap",
			zap.Error(er))
	}

	payload, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		nefario.Log.Error("error marshaling the configMap",
			zap.Error(err))
	}

	configMap.Data["UPGRADE_STATE"] = string(payload)

	_, er = nefario.Client.Clientset().CoreV1().ConfigMaps(nefario.Namespace).
		Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if er != nil {
		nefario.Log.Error("error updating configMap",
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
//
//		w.WriteHeader(http.StatusOK)
//	}
//}

type MinorUpgrade struct {
	mu        sync.Mutex
	finalizer chan bool
	// reverse the dependency
	nefario      *Nefario
	upgradeState *UpgradeState
	pool         *wss.ConnectionPool
}

func (upgrade *MinorUpgrade) report(w http.ResponseWriter, _ *http.Request) {
	report, err := upgrade.nefario.DBMS.LastMinorUpgradeReport()
	if err != nil {
		upgrade.nefario.Log.Error("could not fetch last upgrade report",
			zap.Error(err))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err = json.NewEncoder(w).Encode(report.NodesMinorUpgradeHistory); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (upgrade *MinorUpgrade) minorUpgradev2(w http.ResponseWriter, r *http.Request) {

	if len(upgrade.pool.Clients) > 5 {

		_, err := io.WriteString(w, "maximum number of concurrent connection exceeded!")
		if err != nil {
			upgrade.nefario.Log.Error("error writing max connection exceeding error",
				zap.Error(err))
		}

		upgrade.nefario.Log.Info("exceeds maximum number of concurrent connections!\n quit older running tabs\n")
		// TODO: -- wrap it in one of the error state(possible other: certificate renewal) --
		_, er := w.Write([]byte("exceeds maximum number of concurrent connections!\n quit older running tabs\n"))
		if er != nil {
			upgrade.nefario.Log.Info("error writing concurrent connections warning:",
				zap.Error(er))
		}

		return
	}

	query := r.URL.Query()
	id, exists := query["upgrade"]

	if id == "initialize" && upgrade.nefario.HandlerLock.Mu {

		if exists && len(id) > 0 {
			fmt.Fprintf(w, "Query param 'id' exists with value: %s\n", id[0])
		} else {
			fmt.Fprintln(w, "Query param 'id' does not exist.")
		}
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// TODO: wrap it in error state
		upgrade.nefario.Log.Info("connection upgrade error:",
			zap.Error(err))

		return
	}

	client := &wss.Client{
		Conn: conn,
	}

	for c := range upgrade.pool.Clients {
		if client.Conn == c.Conn {
			upgrade.nefario.Log.Info("removing old connections",
				zap.String("client", client.Conn.RemoteAddr().String()))
			closeMessage := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "closing connection")
			if er := c.Conn.WriteMessage(websocket.CloseMessage, closeMessage); er != nil {
				// handle error
			}

			er := c.Conn.Close()
			if er != nil {
				upgrade.nefario.Log.Info("error closing connection",
					zap.Error(er))
				return
			}
			upgrade.pool.RemoveClient(c)
		}
	}

	upgrade.pool.AddClient(client)
	upgrade.nefario.Log.Info("registered client: ",
		zap.String("client address: ", client.Conn.RemoteAddr().String()))

	// todo: add on prerequisite
	//  if upgrade.nefario.mu.TryLock() {
	//	 // only allow one api would run at a time
	//	 defer upgrade.nefario.mu.Unlock()
	//}

	if upgrade.mu.TryLock() {
		//defer upgrade.mu.Unlock()

		// TODO: race condition - readCtx can be cancelled

		config, er := upgrade.nefario.DBMS.ReadConfig()
		if er != nil {
			upgrade.nefario.Log.Info("failed to read mongo collection: ",
				zap.Error(er))
		}

		// TODO:
		//  derive from main context
		// r.WithContext(context.WithValue(r.Context(), "upgradeState", upgrade.upgradeState))

		// TODO:
		//config, er := readConfig(upgrade.nefario)
		//if er != nil {
		//	upgrade.nefario.Log.Info("error reading config",
		//		zap.Error(er))
		//}

		upgrade.upgradeState = &config
		if upgrade.upgradeState.NodesStatus == nil {
			upgrade.upgradeState.NodesStatus = &NodesStatus{
				NodesUpgraded:            []string{},
				NodesToBeUpgraded:        []string{},
				NodesFailedToBeUpgraded:  []string{},
				NodesSkippedToBeUpgraded: []string{},
			}
		}

		upgrade.finalizer = make(chan bool, 1)

		err := upgrade.nefario.DBMS.MinorUpgradeReport()
		if err != nil {
			upgrade.nefario.Log.Info("failed to report minor upgrade",
				zap.Error(err))
			//TODO:
			// handle gracefully
			// return
		}
		upgrade.MinorUpgrade()

		return
	}

	ctx, _ := context.WithCancel(upgrade.pool.ReadCtx)
	go func() {

		upgrade.nefario.Log.Info("reading message from connection",
			zap.String("client address: ", client.Conn.RemoteAddr().String()))

		er := upgrade.pool.ReadMessageFromConn(ctx, client)
		if er != nil {
			upgrade.nefario.Log.Info("error reading message from connection",
				zap.Error(er))
		}
	}()

	<-upgrade.finalizer

	// can fail
	upgrade.pool.CancelReadContext()

	// TODO: should there be any mutex for removing clients?
	for client = range upgrade.pool.Clients {
		upgrade.pool.RemoveClient(client)
	}

}

func (upgrade *MinorUpgrade) minorUpgrade(w http.ResponseWriter, r *http.Request) {

	if len(upgrade.pool.Clients) > 5 {

		_, err := io.WriteString(w, "maximum number of concurrent connection exceeded!")
		if err != nil {
			upgrade.nefario.Log.Error("error writing max connection exceeding error",
				zap.Error(err))
		}

		upgrade.nefario.Log.Info("exceeds maximum number of concurrent connections!\n quit older running tabs\n")
		// TODO: -- wrap it in one of the error state(possible other: certificate renewal) --
		_, er := w.Write([]byte("exceeds maximum number of concurrent connections!\n quit older running tabs\n"))
		if er != nil {
			upgrade.nefario.Log.Info("error writing concurrent connections warning:",
				zap.Error(er))
		}

		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// TODO: wrap it in error state
		upgrade.nefario.Log.Info("connection upgrade error:",
			zap.Error(err))

		return
	}

	client := &wss.Client{
		Conn: conn,
	}

	for c := range upgrade.pool.Clients {
		if client.Conn == c.Conn {
			upgrade.nefario.Log.Info("removing old connections",
				zap.String("client", client.Conn.RemoteAddr().String()))
			closeMessage := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "closing connection")
			if er := c.Conn.WriteMessage(websocket.CloseMessage, closeMessage); er != nil {
				// handle error
			}

			er := c.Conn.Close()
			if er != nil {
				upgrade.nefario.Log.Info("error closing connection",
					zap.Error(er))
				return
			}
			upgrade.pool.RemoveClient(c)
		}
	}

	upgrade.pool.AddClient(client)
	upgrade.nefario.Log.Info("registered client: ",
		zap.String("client address: ", client.Conn.RemoteAddr().String()))

	// todo: add on prerequisite
	//  if upgrade.nefario.mu.TryLock() {
	//	 // only allow one api would run at a time
	//	 defer upgrade.nefario.mu.Unlock()
	//}

	if upgrade.mu.TryLock() {
		//defer upgrade.mu.Unlock()

		// TODO: race condition - readCtx can be cancelled

		config, er := upgrade.nefario.DBMS.ReadConfig()
		if er != nil {
			upgrade.nefario.Log.Info("failed to read mongo collection: ",
				zap.Error(er))
		}

		// TODO:
		//  derive from main context
		// r.WithContext(context.WithValue(r.Context(), "upgradeState", upgrade.upgradeState))

		// TODO:
		//config, er := readConfig(upgrade.nefario)
		//if er != nil {
		//	upgrade.nefario.Log.Info("error reading config",
		//		zap.Error(er))
		//}

		upgrade.upgradeState = &config
		if upgrade.upgradeState.NodesStatus == nil {
			upgrade.upgradeState.NodesStatus = &NodesStatus{
				NodesUpgraded:            []string{},
				NodesToBeUpgraded:        []string{},
				NodesFailedToBeUpgraded:  []string{},
				NodesSkippedToBeUpgraded: []string{},
			}
		}

		upgrade.finalizer = make(chan bool, 1)

		err := upgrade.nefario.DBMS.MinorUpgradeReport()
		if err != nil {
			upgrade.nefario.Log.Info("failed to report minor upgrade",
				zap.Error(err))
			//TODO:
			// handle gracefully
			// return
		}
		upgrade.MinorUpgrade()

		return
	}

	ctx, _ := context.WithCancel(upgrade.pool.ReadCtx)
	go func() {

		upgrade.nefario.Log.Info("reading message from connection",
			zap.String("client address: ", client.Conn.RemoteAddr().String()))

		er := upgrade.pool.ReadMessageFromConn(ctx, client)
		if er != nil {
			upgrade.nefario.Log.Info("error reading message from connection",
				zap.Error(er))
		}
	}()

	<-upgrade.finalizer

	// can fail
	upgrade.pool.CancelReadContext()

	// TODO: should there be any mutex for removing clients?
	for client = range upgrade.pool.Clients {
		upgrade.pool.RemoveClient(client)
	}

}

func main() {

	nefario := NewNefario()

	// TODO:
	if os.Getenv("MONGO") == "true" {
		m := Mongo{
			DSN:        os.Getenv("MONGO_URI"),
			DB:         "lighthouse",
			Collection: "minorUpgrade",
			Log:        nefario.Log,
		}

		m.DSN = "mongodb://root:5uZLvWUrvJ@kcprod-mongodb-svc.lighthouse.svc.cluster.local:27017"
		err := m.InitConnection()
		if err != nil {
			return
		}

		nefario.DBMS = &m
	}

	config, err := nefario.DBMS.ReadConfig()
	if err != nil {
		nefario.Log.Info("failed to read mongo collection: ",
			zap.Error(err))
	}

	// TODO:
	//  config, err := readConfig(nefario)
	//  if err != nil {
	//	  nefario.Log.Info("failed to read config map: ",
	//  		zap.Error(err))
	//  }

	pool := wss.NewPool()
	pool.Log = nefario.Log
	go pool.Run()

	upgrade := MinorUpgrade{
		pool:         pool,
		upgradeState: &config,
		nefario:      nefario,
	}

	if upgrade.upgradeState.NodesStatus == nil {
		upgrade.upgradeState.NodesStatus = &NodesStatus{
			NodesUpgraded:            []string{},
			NodesToBeUpgraded:        []string{},
			NodesFailedToBeUpgraded:  []string{},
			NodesSkippedToBeUpgraded: []string{},
		}
	}

	if len(upgrade.upgradeState.NodesStatus.NodesToBeUpgraded) != 0 {
		er := continueMinorUpgrade(&upgrade)
		if er != nil {
			upgrade.nefario.Log.Info("error to continue trying minor upgrade:",
				zap.Error(er))
		}
	}

	err = prerequisites(&upgrade)
	if err != nil {
		upgrade.nefario.Log.Info("error to check prerequisites:",
			zap.Error(err))
	}

	http.HandleFunc("/minor-upgrade-v2", upgrade.minorUpgradev2)
	http.HandleFunc("/minor-upgrade", upgrade.minorUpgrade)
	http.HandleFunc("/report", upgrade.report)
	http.HandleFunc("/livez", livez)

	upgrade.nefario.Log.Info("starting kapetanios server on :80")
	defer handlePanic()
	err = http.ListenAndServe(":80", nil)
	if err != nil {
		upgrade.nefario.Log.Error("error starting kapetanios server",
			zap.Error(err))
		panic(err)
	}
}

func handlePanic() {
	fmt.Println("handling panic")
	if r := recover(); r != nil {
		// TODO:
		//  store timestamp in a configMap list
		fmt.Println("Recovered from panic:", r)
	}
}
