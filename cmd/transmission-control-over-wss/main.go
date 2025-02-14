package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/shishir9159/kapetanios/internal/wss"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	port = flag.Int("port", 50051, "The server port")
	addr = flag.String("addr", "kapetanios.default.svc.cluster.local:80", "http service address")
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 0,
	ReadBufferSize:   0,
	WriteBufferSize:  0,
	WriteBufferPool:  nil,
	Subprotocols:     nil,
	Error:            nil,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	EnableCompression: false,
}

type MinorityReport struct {
	CurrentStep           uint8    `yaml:"currentStep"`
	UpgradedNodes         []string `yaml:"upgradedNodes"`
	NodesToBeUpgraded     []string `yaml:"nodesToBeUpgraded"`
	RedhatK8sVersion      string   `yaml:"redhatK8sVersion"`
	UbuntuK8sVersion      string   `yaml:"ubuntuK8sVersion"`
	MinorUpgradeNamespace string   `yaml:"minorUpgradeNamespace"`
}

type Server struct {
	ctx         context.Context
	initialized bool
	pool        *wss.ConnectionPool
}

func readJSONConfig() (MinorityReport, error) {

	if _, err := os.Stat("/etc/report/upgrade.json"); err != nil {
		return MinorityReport{}, err
	}

	data, err := os.ReadFile("/etc/report/upgrade.json")
	if err != nil {
		return MinorityReport{}, err
	}

	report := MinorityReport{}
	var json = jsoniter.ConfigFastest
	if err = json.Unmarshal(data, &report); err != nil {
		return MinorityReport{}, err
	}

	return report, nil
}

func writeJSONConfig(report MinorityReport) error {

	var json = jsoniter.ConfigFastest
	reportJson, err := json.Marshal(report)
	if err != nil {
		return err
	}

	if err = os.WriteFile("/etc/report/upgrade.json", reportJson, 0644); err != nil {
		return err
	}

	return nil
}

func (server *Server) minorUpgrade(w http.ResponseWriter, r *http.Request) {

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

	if server.initialized {
		ctx, _ := context.WithCancel(server.pool.ReadCtx)
		go server.pool.ReadMessageFromConn(ctx, client)
		//go server.pool.ReadMessageFromConn(ctx)
		// TODO: use the context
		time.Sleep(480 * time.Second)
		return
	}

	// TODO: race condtion - readctx can be cancelled

	minorityReport := MinorityReport{
		CurrentStep:           0,
		MinorUpgradeNamespace: "default",
	}

	MinorUpgrade(&minorityReport, server.pool)
	server.initialized = true
}

func StartServer(ctx context.Context) {

	pool := wss.NewPool()
	// TODO: remove all the clients when the job ends
	go pool.Run()

	server := Server{
		ctx:         ctx,
		initialized: false,
		pool:        pool,
	}

	http.HandleFunc("/minor-upgrade", server.minorUpgrade)
	http.HandleFunc("/", home)

	fmt.Println("WebSocket server started on :80")

	er := http.ListenAndServe(":80", nil)
	if er != nil {
		panic(er)
	}
}

func main() {

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
