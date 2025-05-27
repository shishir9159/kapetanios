package main

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	WriteBufferPool:  nil,
	HandshakeTimeout: 0,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Error:             nil,
	Subprotocols:      nil,
	EnableCompression: false,
}

type testConnection struct {
	pool *ConnectionPool
	log  *zap.Logger
}

type Client struct {
	send chan []byte     // Buffered channel of outbound messages: need to implement
	Conn *websocket.Conn `json:"conn"`
}

type ConnectionPool struct {
	broadcast   chan []byte
	register    chan *Client
	unregister  chan *Client
	mutex       sync.RWMutex
	cancel      context.CancelFunc
	readCancel  context.CancelFunc
	Ctx         context.Context  `json:"context"`
	Clients     map[*Client]bool `json:"clients"`
	ReadCtx     context.Context  `json:"readContext"`
	MessageChan chan string      `json:"messageChan"`
}

func NewPool() *ConnectionPool {
	ctx, cancel := context.WithCancel(context.Background())
	readCtx, readCancel := context.WithCancel(ctx)

	return &ConnectionPool{
		Ctx:         ctx,
		cancel:      cancel,
		ReadCtx:     readCtx,
		readCancel:  readCancel,
		Clients:     make(map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan []byte),
		MessageChan: make(chan string, 1),
	}
}

func (pool *ConnectionPool) Run(log *zap.Logger) {

	for {

		select {
		case client := <-pool.register:
			pool.Clients[client] = true

		case client := <-pool.unregister:
			if _, ok := pool.Clients[client]; ok {
				delete(pool.Clients, client)
				log.Info("client unregistered",
					zap.Int("runtime.NumGoroutine ", runtime.NumGoroutine()),
					zap.String("client address", client.Conn.RemoteAddr().String()),
					zap.Int("remaining clients", len(pool.Clients)))

				if client.Conn != nil {
					err := client.Conn.Close()
					if err != nil {
						log.Error("close connection failed",
							zap.String("client address", client.Conn.RemoteAddr().String()),
							zap.Error(err))
					}
				}
			}

		case message := <-pool.broadcast:
			for client := range pool.Clients {
				if client.Conn == nil {
					pool.unregister <- client
					log.Error("client connection is nil",
						zap.String("message body:", string(message)),
						zap.Int("remaining clients", len(pool.Clients)),
						zap.String("client address", client.Conn.RemoteAddr().String()))

				} else {

					err := client.Conn.WriteMessage(websocket.TextMessage, message)
					if err != nil {
						log.Error("error writing message",
							zap.String("message body:", string(message)),
							zap.Int("remaining clients", len(pool.Clients)),
							//zap.Int("closing error", websocket.CloseError.Error(err))),
							zap.String("client address", client.Conn.RemoteAddr().String()),
							zap.Error(err))
					}
				}
			}
		}
	}
}

func (pool *ConnectionPool) AddClient(client *Client) {
	pool.register <- client
}

func (pool *ConnectionPool) RemoveClient(client *Client) {
	pool.unregister <- client
}

func (pool *ConnectionPool) CancelReadContext() {
	pool.readCancel()
	pool.ReadCtx, pool.readCancel = context.WithCancel(pool.Ctx)
}

func (pool *ConnectionPool) BroadcastMessage(message []byte) {
	pool.broadcast <- message
}

func (pool *ConnectionPool) ReadMessages() (string, error) {

	for i := 0; ; i++ {
		if len(pool.Clients) == 0 {

			time.Sleep(3 * time.Second)
			log.Println("no clients in the pool to read message from",
				zap.Int("retry count", i))

			continue
		}

		break
	}

	ctx, _ := context.WithCancel(pool.ReadCtx)

	for client := range pool.Clients {
		if pool.Clients[client] {
			go func() {
				err := pool.ReadMessageFromConn(ctx, client)
				if err != nil {
					log.Println("error reading message from connection",
						zap.String("client address", client.Conn.RemoteAddr().String()),
						zap.Error(err))
				}
			}()
		}
	}

	message := <-pool.MessageChan
	pool.CancelReadContext()

	return strings.TrimSpace(message), nil
}

func (pool *ConnectionPool) ReadMessageFromConn(ctx context.Context, client *Client) error {

	pool.Clients[client] = false

	log.Printf("reading message from connection with client address being %s", client.Conn.RemoteAddr().String())

	for {

		select {
		case <-ctx.Done():
			pool.Clients[client] = true
			log.Println("context done", client.Conn.RemoteAddr().String())
			return nil

		default:

			fmt.Println("default entry")

			err := client.Conn.SetReadDeadline(time.Now().Add(10 * time.Second))
			if err != nil {
				log.Println("error setting read deadline for client",
					zap.Error(err))
				return err
			}

			msgType, msg, err := client.Conn.ReadMessage()
			if err != nil || msgType == websocket.CloseMessage {
				if websocket.IsCloseError(err) {
					fmt.Println("websocket connection close error", client.Conn.RemoteAddr().String())
				}
				fmt.Sprintf("error reading from %s for messagetype %d: %v - message: %s",
					client.Conn.RemoteAddr().String(), msgType, err, string(msg))
				pool.Clients[client] = true
				pool.RemoveClient(client)
				return err
			}

			if msgType != websocket.TextMessage {
				fmt.Sprintf("unexpected message type: %v", msgType)
			}

			fmt.Sprintf("received from %s: %s",
				client.Conn.RemoteAddr().String(), string(msg))

			pool.MessageChan <- string(msg)
			pool.Clients[client] = true

			return nil
		}
	}
}

func (t *testConnection) testAddClient(w http.ResponseWriter, r *http.Request) {

	t.log.Info("testAddClient called")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		t.log.Info("connection upgrade error:",
			zap.Error(err))

		return
	}

	client := Client{
		Conn: conn,
	}

	//for {
	//
	//}

	t.pool.AddClient(&client)
	messages, err := t.pool.ReadMessages()
	if err != nil {
		t.log.Info("error reading messages from pool",
			zap.Error(err))
		return
	}

	t.pool.BroadcastMessage([]byte(messages))

	t.log.Info("messages read from pool",
		zap.String("messages", messages),
		zap.Int("number of go routines", runtime.NumGoroutine()))

}

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

	pool := NewPool()
	go pool.Run(logger)

	t := testConnection{
		pool: pool,
		log:  logger,
	}

	http.Handle("/test", http.HandlerFunc(t.testAddClient))
	t.log.Info("http server started",
		zap.String("address", ":80"),
		zap.String("mode", "debug"),
		zap.String("log level", "debug"),
		zap.String("log format", "json"),
		zap.String("log output", "stdout"),
		zap.String("log error output", "stderr"),
		zap.String("log encoding", "json"),
		zap.String("log level encoder", "capital"),
		zap.String("log time encoder", "iso8601"))
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Println(err)
		return
	}
}
