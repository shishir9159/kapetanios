package wss

import (
	"context"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Client struct {
	send chan []byte     // Buffered channel of outbound messages: need to implement
	Conn *websocket.Conn `json:"conn"`
}

type ConnectionPool struct {
	Log         *zap.Logger
	broadcast   chan []byte
	register    chan *Client
	unregister  chan *Client
	mu          sync.RWMutex
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

func (pool *ConnectionPool) Run() {

	for {
		select {
		case client := <-pool.register:
			pool.Clients[client] = true

		case client := <-pool.unregister:
			pool.mu.Lock()
			if _, ok := pool.Clients[client]; ok {
				pool.Log.Info("client unregistered",
					zap.String("client address", client.Conn.RemoteAddr().String()),
					zap.Int("remaining clients", len(pool.Clients)))
				delete(pool.Clients, client)

				if client.Conn != nil {
					err := client.Conn.Close()
					if err != nil {
						pool.Log.Error("close connection failed",
							zap.String("client address", client.Conn.RemoteAddr().String()),
							zap.Error(err))
					}
				}
			}
			pool.mu.Unlock()

		case message := <-pool.broadcast:
			for client := range pool.Clients {
				if client.Conn == nil {
					pool.unregister <- client
					pool.Log.Error("client connection is nil",
						zap.String("message body:", string(message)),
						zap.Int("remaining clients", len(pool.Clients)),
						zap.String("client address", client.Conn.RemoteAddr().String()))

				} else {

					err := client.Conn.WriteMessage(websocket.TextMessage, message)
					if err != nil {
						pool.Log.Error("error writing message",
							zap.String("message body:", string(message)),
							zap.Int("remaining clients", len(pool.Clients)),
							//zap.Int("closing error", websocket.CloseError.Error(err))),
							zap.String("client address", client.Conn.RemoteAddr().String()),
							zap.Error(err))
						// TODO: concurrent map reading writing failure
						// pool.mu.RLock()
						// TODO:
						//  pool.unregister <- client
						// pool.Mutex.RUnlock()
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
		if len(pool.Clients) != 0 {
			break
		}

		time.Sleep(3 * time.Second)
		pool.Log.Info("no clients in the pool to read from",
			zap.Int("retry count", i))
	}

	ctx, _ := context.WithCancel(pool.ReadCtx)

	for client := range pool.Clients {
		if pool.Clients[client] {
			go func() {
				err := pool.ReadMessageFromConn(ctx, client)
				if err != nil {
					pool.Log.Error("error reading message from connection",
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

// pub sub and a single fifo wouldn't make this easier
// as unsubscribing to listening to message would still
// be challenging
// todo:
//  Circular Queue or Ring Buffer Requirements:
//  - FIFO
//  - runtime.NumGoroutine()
//  - bounded at compile time
//  - single reader, multiple writers
//  - sync pool to reuse the heap allocations
//  - provide methods that blocked until it could pop
//  - maybe a long lived read connection is not bad if managed
//  - provide methods that could pop if possible or return !ok
//  - set limit on queue element size and insertion rate per element
//  - how should I read? circular? pub sub pattern when an input happens?

func (pool *ConnectionPool) ReadMessageFromConn(ctx context.Context, client *Client) error {

	pool.Clients[client] = false
	pool.Log.Info("reading message from connection with client address being",
		zap.String("client address", client.Conn.RemoteAddr().String()),
		zap.Int("number of go routines", runtime.NumGoroutine()),
		zap.Int("remaining clients", len(pool.Clients)))

	for {

		select {
		case <-ctx.Done():
			pool.Clients[client] = true
			pool.Log.Info("context done",
				zap.String("client address", client.Conn.RemoteAddr().String()))
			return nil

		default:
			msgType, msg, err := client.Conn.ReadMessage()
			if err != nil || msgType == websocket.CloseMessage {
				if websocket.IsCloseError(err) {
					pool.Log.Info("websocket connection close error",
						zap.String("client address", client.Conn.RemoteAddr().String()))
				}

				pool.Log.Info("error reading",
					zap.String("client address", client.Conn.RemoteAddr().String()),
					zap.Int("message type", msgType),
					zap.ByteString("message", msg),
					zap.Error(err))
				pool.Clients[client] = true
				pool.RemoveClient(client)
				return err
			}

			if msgType != websocket.TextMessage {
				pool.Log.Info("unexpected message type",
					zap.Int("message type", msgType))
			}

			pool.Log.Info("received from",
				zap.String("client address", client.Conn.RemoteAddr().String()),
				zap.ByteString("message", msg))

			pool.MessageChan <- string(msg)
			pool.Clients[client] = true

			return nil
		}
	}
}
