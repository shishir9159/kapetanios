package wss

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
	"sync"
)

var upgrader = websocket.Upgrader{HandshakeTimeout: 0,
	ReadBufferSize:  0,
	WriteBufferSize: 0,
	WriteBufferPool: nil,
	Subprotocols:    nil,
	Error:           nil,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	EnableCompression: false,
}

type Client struct {
	Conn *websocket.Conn `json:"conn"`
	// TODO: check if this lock is necessary
	Mu sync.Mutex `json:"mu"`
}

type ConnectionPool struct {
	broadcast   chan []byte
	register    chan *Client
	unregister  chan *Client
	mutex       sync.RWMutex
	cancel      context.CancelFunc
	readCancel  context.CancelFunc
	MessageChan chan string      `json:"messageChan"`
	Ctx         context.Context  `json:"context"`
	Clients     map[*Client]bool `json:"clients"`
	ReadCtx     context.Context  `json:"readContext"`
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
			//pool.Mutex.Lock()
			pool.Clients[client] = true
			//pool.Mutex.Unlock()
			fmt.Println("client registered")

		case client := <-pool.unregister:
			//pool.Mutex.Lock()
			if _, ok := pool.Clients[client]; ok {
				delete(pool.Clients, client)
				err := client.Conn.Close()
				if err != nil {
					fmt.Println("error closing connections", err)
					return
				}
				fmt.Println("client unregistered", len(pool.Clients))
			}
			//pool.Mutex.Unlock()

		case message := <-pool.broadcast:
			//pool.Mutex.RLock()
			for client := range pool.Clients {
				// maybe not only writeJson will work
				//err := client.Conn.WriteJSON(message)
				err := client.Conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Println("error writing message:", err, message)
					pool.unregister <- client
				}
			}
			//pool.Mutex.RUnlock()
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
}

func (pool *ConnectionPool) BroadcastMessage(message []byte) {
	pool.broadcast <- message
}

func (pool *ConnectionPool) ReadMessages() (string, error) {

	// todo: derive this child context from the inherited context
	//  Create a context with cancel to stop all Goroutines

	ctx, _ := context.WithCancel(pool.ReadCtx)

	for client := range pool.Clients {
		if pool.Clients[client] {
			go pool.ReadMessageFromConn(ctx, client)
		}
	}

	message := <-pool.MessageChan
	pool.CancelReadContext()

	// TODO: after cancellation, reinitialize the Read Context
	pool.ReadCtx, pool.readCancel = context.WithCancel(pool.Ctx)

	return strings.TrimSpace(message), nil
}

func (pool *ConnectionPool) ReadMessageFromConn(ctx context.Context, client *Client) {
	pool.Clients[client] = false
	for {
		select {
		case <-ctx.Done():
			pool.Clients[client] = true
			return
		default:
			msgType, msg, err := client.Conn.ReadMessage()
			if err != nil || msgType == websocket.CloseMessage {
				log.Printf("error reading from %s for messagetype %d: %v", client.Conn.RemoteAddr().String(), msgType, err)
				pool.RemoveClient(client)
				return
			}

			if msgType != websocket.TextMessage {
				log.Printf("unexpected message type: %v", msgType)
			}
			log.Printf("received from %s: %s", client.Conn.RemoteAddr().String(), string(msg))

			select {
			case pool.MessageChan <- string(msg):
			default:
			}

			pool.Clients[client] = true
			return
		}
	}
}
