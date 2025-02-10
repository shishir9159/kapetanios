package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
)

var (
	minorUpgradeNamespace = "default"
	port                  = flag.Int("port", 50051, "The server port")
	addr                  = flag.String("addr", "kapetanios.default.svc.cluster.local:80", "http service address")
	Clients               = make(map[int]*websocket.Conn)
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

type Server struct {
	currentStep   uint8
	clients       map[*websocket.Conn]bool
	handleMessage func(message []byte) // New message handler
}

func (server *Server) echo(w http.ResponseWriter, r *http.Request) {

	if len(Clients) == 0 {
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}

		Clients[0] = connection

		defer func(connection *websocket.Conn) {
			delete(Clients, 0)
			//delete(server.clients, connection)
			err = connection.Close()
			if err != nil {
				log.Println("error closing connection:", err)
			}
		}(connection)
	}

	//server.clients[connection] = true // Save the connection using it as a key

	for {
		mt, message, er := Clients[0].ReadMessage()

		if er != nil || mt == websocket.CloseMessage {
			log.Println("read error:", er)
			break // Exit the loop if the client tries to close the connection or the connection is interrupted
		}

		log.Printf("recv: %s", message)

		server.WriteMessage(message)

		er = Clients[0].WriteMessage(mt, message)
		if er != nil {
			log.Println("write:", er)
			break
		}

		go server.handleMessage(message)
	}
}

func (server *Server) handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("error upgrading connection:", err)
		return
	}

	defer func(conn *websocket.Conn) {
		er := conn.Close()
		if er != nil {
			fmt.Println("error closing connection:", er)
		}
	}(conn)

	for {
		msgType, msg, er := conn.ReadMessage()
		fmt.Println("read:", msgType, msg, er)
		if er != nil {
			fmt.Println("error reading message:", er, msgType)
			break
		}

		response := processMessage(string(msg))

		if er = conn.WriteMessage(websocket.TextMessage, []byte(response)); er != nil {
			fmt.Println("error writing message:", er)
			break
		}
	}
}

func processMessage(msg string) string {
	switch msg {
	case "step 1":
		return "response for step 1"
	case "step 2":
		return "response for step 2"
	case "step 3":
		return "response for step 3"
	case "step 4":
		return "response for step 4"
	case "step 5":
		return "response for step 5"
	default:
		return "unknown step"
	}
}

func (server *Server) minorUpgrade(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("error upgrading connection:", err)
		return
	}

	defer func(conn *websocket.Conn) {
		er := conn.Close()
		if er != nil {
			fmt.Println("error closing connection:", er)
		}
	}(conn)

	MinorUpgradeFirstRun(minorUpgradeNamespace, conn)
}

// func StartServer(handleMessage func(message []byte)) *Server {

func StartServer(handleMessage func(message []byte)) {

	server := Server{
		0,
		make(map[*websocket.Conn]bool),
		handleMessage,
	}

	http.HandleFunc("/minor-upgrade", server.minorUpgrade)
	http.HandleFunc("/ws", server.handleConnection)
	http.HandleFunc("/echo", server.echo)
	http.HandleFunc("/", home)

	fmt.Println("WebSocket server started on :80")

	//go func() {
	er := http.ListenAndServe(":80", nil)
	if er != nil {
		panic(er)
	}
	//}()

	//return &server
}

func main() {

	// todo: resume connection
	//  Prerequisites(minorUpgradeNamespace)

	StartServer(messageHandler)
}

func messageHandler(message []byte) {
	fmt.Println(string(message))
}

func (server *Server) WriteMessage(message []byte) {
	for index := range Clients {
		err := Clients[index].WriteMessage(websocket.TextMessage, message)
		if err != nil {
			return
		}
	}
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
