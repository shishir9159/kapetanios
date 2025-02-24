package wss

import (
	"context"
	"fmt"
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
			//pool.Mutex.Lock()
			pool.Clients[client] = true
			//pool.Mutex.Unlock()
			fmt.Println("client registered")

		case client := <-pool.unregister:
			//pool.Mutex.Lock()
			if _, ok := pool.Clients[client]; ok {
				log.Info("client unregistered",
					zap.String("client address", client.Conn.RemoteAddr().String()),
					zap.Int("remaining clients", len(pool.Clients)))
				delete(pool.Clients, client)
				err := client.Conn.Close()
				if err != nil {
					log.Error("close connection failed",
						zap.Error(err))
				}
			}
			//pool.Mutex.Unlock()

		case message := <-pool.broadcast:
			// TODO: concurrent map reading writing failure
			//pool.Mutex.RLock()
			for client := range pool.Clients {
				// maybe not only writeJson will work
				//err := client.Conn.WriteJSON(message)
				err := client.Conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Error("error writing message:",
						zap.String("message body:", string(message)),
						zap.Int("remaining clients", len(pool.Clients)),
						//zap.Int("closing error", websocket.CloseError.Error(err))),
						zap.Error(err))
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
	pool.ReadCtx, pool.readCancel = context.WithCancel(pool.Ctx)
}

func (pool *ConnectionPool) BroadcastMessage(message []byte) {
	pool.broadcast <- message
}

func (pool *ConnectionPool) ReadMessages() (string, error) {

	for {
		if len(pool.Clients) == 0 {
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	ctx, _ := context.WithCancel(pool.ReadCtx)

	for client := range pool.Clients {
		if pool.Clients[client] {
			go pool.ReadMessageFromConn(ctx, client)
		}
	}

	message := <-pool.MessageChan
	pool.CancelReadContext()

	return strings.TrimSpace(message), nil
}

func (pool *ConnectionPool) ReadMessageFromConn(ctx context.Context, client *Client) {
	pool.Clients[client] = false
	for {
		select {
		case <-ctx.Done():
			pool.Clients[client] = true
			log.Println()
			return
		default:
			msgType, msg, err := client.Conn.ReadMessage()
			if err != nil || msgType == websocket.CloseMessage {
				log.Printf("error reading from %s for messagetype %d: %v",
					client.Conn.RemoteAddr().String(), msgType, err)
				pool.RemoveClient(client)
				return
			}

			if msgType != websocket.TextMessage {
				log.Printf("unexpected message type: %v",
					msgType)
			}
			log.Printf("received from %s: %s",
				client.Conn.RemoteAddr().String(), string(msg))

			select {
			case pool.MessageChan <- string(msg):
			default:
				
			}

			pool.Clients[client] = true
			return
		}
	}
}
