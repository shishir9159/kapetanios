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
	Mu   sync.Mutex      `json:"mu"`
}

type ConnectionPool struct {
	Context    context.Context  `json:"context"`
	Clients    map[*Client]bool `json:"clients"`
	Buffer     []string         `json:"buffer"`
	Mutex      sync.RWMutex     `json:"mutex"`
	Register   chan *Client     `json:"register"`
	Unregister chan *Client     `json:"unregister"`
	Broadcast  chan []byte      `json:"broadcast"`
}

func NewPool() *ConnectionPool {
	return &ConnectionPool{
		Clients:    make(map[*Client]bool),
		Buffer:     nil,
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan []byte),
	}
}

func (pool *ConnectionPool) Run() {
	for {
		select {
		case client := <-pool.Register:
			//pool.Mutex.Lock()
			pool.Clients[client] = true
			//pool.Mutex.Unlock()
			fmt.Println("client registered")

		case client := <-pool.Unregister:
			//pool.Mutex.Lock()
			if _, ok := pool.Clients[client]; ok {
				delete(pool.Clients, client)
				err := client.Conn.Close()
				if err != nil {
					return
				}
				fmt.Println("client unregistered")
			}
			//pool.Mutex.Unlock()

		case message := <-pool.Broadcast:
			//pool.Mutex.RLock()
			for client := range pool.Clients {
				// maybe only writeJson will work
				err := client.Conn.WriteJSON( message)
				if err != nil {
					log.Println("error writing message:", err)
					pool.Unregister <- client // Unregister client on error
				}
			}
			//pool.Mutex.RUnlock()
		}
	}
}

func (pool *ConnectionPool) AddClient(client *Client) {
	pool.Register <- client
}

func (pool *ConnectionPool) RemoveClient(client *Client) {
	pool.Unregister <- client
}

func (pool *ConnectionPool) BroadcastMessage(message []byte) {
	pool.Broadcast <- message
}

func (pool *ConnectionPool) ReadMessages() {
	defer func(Conn *websocket.Conn) {
		err := Conn.Close()
		if err != nil {

		}
	}(c.Conn)

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			// client is removed from the pool on error
			pool.RemoveClient(c)
			break
		}

		c.Mu.Lock()
		c.Buffer = append(c.Buffer, string(message)) // Store input
		c.Mu.Unlock()

		fmt.Printf("received: %s from client\n", message)

		if err = c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Println("error writing message:", err)
			pool.RemoveClient(c)
			break
		}
	}
}

func (c *Client) GetInputs() []string {
	//c.Mu.Lock()
	//defer c.Mu.Unlock()
	return c.Buffer // Return a copy to avoid data race if the slice is modified elsewhere
}


func readMessage(ctx context.Context, conn *websocket.Conn, messageChan chan string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Error reading from %s: %v", conn.RemoteAddr().String(), err)
				return
			}

			if msgType != websocket.TextMessage {
				log.Printf("unexpected message type: %v", msgType)
			}

			log.Printf("Received from %s: %s", conn.RemoteAddr().String(), string(msg))

			select {
			case messageChan <- string(msg):
			default:
			}
			return
		}
	}
}

func writeMessage[T any](value T, clients map[*websocket.Conn]bool) (string, error) {

	// Create a context with cancel to stop all Goroutines
	ctx, cancel := context.WithCancel(context.Background())

	if ctx.Deadline()

	// Channel to receive the first message
		messageChan := make(chan string, 1)

	for conn := range clients {
		if err := conn.WriteJSON(value); err != nil {
			cancel()
			return "", err
		}

		// Start reading messages for all clients
		go readMessage(ctx, conn, messageChan)
	}

	// Wait for the first message
	message := <-messageChan

	// Stop all reading Goroutines
	cancel()

	return strings.TrimSpace(string(message)), nil
}
