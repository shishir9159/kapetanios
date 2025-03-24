package wss

import (
	"context"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"log"
	"strings"
	"sync"
	"time"
)

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
			// pool.Mutex.Lock()
			if _, ok := pool.Clients[client]; ok {
				log.Info("client unregistered",
					zap.String("client address", client.Conn.RemoteAddr().String()),
					zap.Int("remaining clients", len(pool.Clients)))
				delete(pool.Clients, client)

				if client.Conn != nil {
					err := client.Conn.Close()
					if err != nil {
						log.Error("close connection failed",
							zap.String("client address", client.Conn.RemoteAddr().String()),
							zap.Error(err))
					}
				}
			}
			// pool.Mutex.Unlock()

		case message := <-pool.broadcast:
			// TODO: concurrent map reading writing failure
			// pool.Mutex.RLock()
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
						// TODO:
						//  pool.unregister <- client
					}
				}
			}
			// pool.Mutex.RUnlock()
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

	log.Printf("reading message from connection with client address being %s", client.Conn.RemoteAddr().String())

	for {

		select {
		case <-ctx.Done():
			pool.Clients[client] = true
			log.Println("context done",
				zap.String("client address", client.Conn.RemoteAddr().String()))
			return nil

		default:
			// TODO: exponentially back off till 15 seconds
			err := client.Conn.SetReadDeadline(time.Now().Add(400 * time.Second))
			if err != nil {
				log.Println("error setting read deadline for client",
					zap.Error(err))
				// TODO:
				//  pool.RemoveClient(client)
				return err
			}

			msgType, msg, err := client.Conn.ReadMessage()
			if err != nil || msgType == websocket.CloseMessage {
				if websocket.IsCloseError(err) {
					log.Println("websocket connection close error",
						zap.String("client address", client.Conn.RemoteAddr().String()))
				}
				log.Printf("error reading from %s for messagetype %d: %v - message: %s",
					client.Conn.RemoteAddr().String(), msgType, err, string(msg))
				pool.Clients[client] = true
				pool.RemoveClient(client)
				return err
			}

			if msgType != websocket.TextMessage {
				log.Printf("unexpected message type: %v",
					msgType)
			}

			log.Printf("received from %s: %s",
				client.Conn.RemoteAddr().String(), string(msg))

			pool.MessageChan <- string(msg)
			pool.Clients[client] = true

			return nil
		}
	}
}
