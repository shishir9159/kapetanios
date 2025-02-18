package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/shishir9159/kapetanios/internal/orchestration"
	"github.com/shishir9159/kapetanios/internal/wss"
	"go.uber.org/zap"
	"html/template"
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

type upgradeProgression struct {
	CurrentStep           uint8  `yaml:"currentStep"`
	MinorUpgradeNamespace string `yaml:"minorUpgradeNamespace"`
}

type MinorityReport struct {
	certificateRenewal bool   `yaml:"certificateRenewal"`
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
	pool        *wss.ConnectionPool
}

func readJSONConfig(c Controller) (MinorityReport, error) {

	configMapName := "kapetanios"

	configMap, er := c.client.Clientset().CoreV1().ConfigMaps(c.namespace).Get(context.Background(), configMapName, metav1.GetOptions{})
	if er != nil {
		c.log.Error("error fetching the configMap",
			zap.Error(er))
		return MinorityReport{}, er
	}

	report := MinorityReport{
		nodesUpgraded:     configMap.Data["NODES_UPGRADED"],
		NodesToBeUpgraded: configMap.Data["UBUNTU_K8S_VERSION"],
		UbuntuK8sVersion:  configMap.Data["REDHAT8_K8S_VERSION"],
		Redhat8K8sVersion: configMap.Data["REDHAT9_K8S_VERSION"],
		Redhat9K8sVersion: configMap.Data["NODES_TO_BE_UPGRADED"],
	}

	return report, nil
}

func writeConfig(c Controller, report MinorityReport) error {

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

func readyz(isReady *atomic.Value) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if isReady == nil || !isReady.Load().(bool) {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (server *Server) minorUpgrade(w http.ResponseWriter, r *http.Request) {
	var Json = jsoniter.ConfigFastest
	decoder := Json.NewDecoder(r.Body)
	var t upgradeProgression
	err := decoder.Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		panic(err)
	}

	log.Println(t)
}

// TODO:  URGENT FIX - IT WORKS ONLY ONCE
// TODO:  URGENT FIX - NEW CLIENT IS ENTERING
func (server *Server) minorUpdateUpgrade(w http.ResponseWriter, r *http.Request) {

	// TODO:
	//  prepare the global minority report

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

	if server.mu.TryLock() == false {
		ctx, _ := context.WithCancel(server.pool.ReadCtx)
		go server.pool.ReadMessageFromConn(ctx, client)
		//go server.pool.ReadMessageFromConn(ctx)
		// TODO: use the context
		// todo: channel
		time.Sleep(480 * time.Second)
		return
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	// TODO: race condition - readCtx can be cancelled

	// todo --------- the process needs to be auto started ---------
	//  ------------- and server.initialized must be true ----------
	Client, _ := orchestration.NewClient()
	logger := zap.Must(zap.NewProduction())
	defer func(logger *zap.Logger) {
		er := logger.Sync()
		if er != nil {
			logger.Info("error syncing logger before application terminates",
				zap.Error(er))
		}
	}(logger)

	namespace := os.Getenv("KAPETANIOS_NAMESPACE")

	c := Controller{
		log:       logger,
		client:    Client,
		namespace: namespace,
		ctx:       context.Background(),
	}
	// todo ---------

	minorityReport, err := readJSONConfig(c)
	if err != nil {
		// TODO: no restart mode or draining
		log.Println("error reading config map:", err)
		//c.log.Error("could not read config map",
		//	zap.Error(err))
	}

	MinorUpgrade(server.pool, minorityReport)
}

func StartServer(ctx context.Context) {

	pool := wss.NewPool()
	// TODO: remove all the clients when the job ends
	go pool.Run()

	server := Server{
		ctx:  ctx,
		pool: pool,
	}

	http.HandleFunc("/minor-upgrade", server.minorUpdateUpgrade)
	// TODO: work on this api
	http.HandleFunc("/upgrade", server.minorUpgrade)
	http.HandleFunc("/livez", livez)
	http.HandleFunc("/", home)

	fmt.Println("WebSocket server started on :80")

	er := http.ListenAndServe(":80", nil)
	if er != nil {
		panic(er)
	}
}

func main() {

	//report, err	 := readJSONConfig()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//if len(report.NodesToBeUpgraded) != 0 {
	//
	//}

	// todo: resume connections after server restarts
	//  Prerequisites(minorUpgradeNamespace)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	StartServer(ctx)
}

func home(w http.ResponseWriter, r *http.Request) {
	err := homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
	if err != nil {
		return
	}
}

func runAway(ws *websocket.Conn) {
	_, _, err := ws.ReadMessage()
	var ce *websocket.CloseError
	if errors.As(err, &ce) {
		switch ce.Code {
		case websocket.CloseNormalClosure,
			websocket.CloseGoingAway,
			websocket.CloseNoStatusReceived:
			//todo: s.env.Statusf("Web socket closed by client: %s", err)
			return
		}
	}
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {

    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;

    var print = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        output.appendChild(d);
        output.scroll(0, output.scrollHeight);
    };

    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };

    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };

    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };

});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output" style="max-height: 70vh;overflow-y: scroll;"></div>
</td></tr></table>
</body>
</html>
`))
