package wss

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
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
	Conn   *websocket.Conn `json:"conn"`
	Buffer []string        `json:"buffer"`
	Mu     sync.Mutex      `json:"mu"`
}

type Pool struct {
	clients    map[*Client]bool `json:"clients"`
	mutex      sync.RWMutex     `json:"mutex"`
	register   chan *Client     `json:"register"`
	unregister chan *Client     `json:"unregister"`
	broadcast  chan []byte      `json:"broadcast"`
}

func NewPool() *Pool {
	return &Pool{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
	}
}

func (p *Pool) Run() {
	for {
		select {
		case client := <-p.register:
			p.mutex.Lock()
			p.clients[client] = true
			p.mutex.Unlock()
			fmt.Println("Client registered")

		case client := <-p.unregister:
			p.mutex.Lock()
			if _, ok := p.clients[client]; ok {
				delete(p.clients, client)
				err := client.Conn.Close()
				if err != nil {
					return
				}
				fmt.Println("Client unregistered")
			}
			p.mutex.Unlock()

		case message := <-p.broadcast:
			p.mutex.RLock()
			for client := range p.clients {
				err := client.Conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Println("Error writing message:", err)
					p.unregister <- client // Unregister client on error
				}
			}
			p.mutex.RUnlock()
		}
	}
}

func (p *Pool) AddClient(client *Client) {
	p.register <- client
}

func (p *Pool) RemoveClient(client *Client) {
	p.unregister <- client
}

func (p *Pool) Broadcast(message []byte) {
	p.broadcast <- message
}

func (c *Client) ReadInputs() {
	defer c.Conn.Close() // Close connection on exit

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			pool.RemoveClient(c) // Remove client from pool on error
			break
		}

		c.mu.Lock()
		c.Buffer = append(c.Buffer, string(message)) // Store input
		c.mu.Unlock()

		fmt.Printf("Received: %s from client\n", message)

		// Example: Echo back the message (optional)
		if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Println("Error writing message:", err)
			pool.RemoveClient(c)
			break
		}
	}
}

func (c *Client) GetInputs() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Buffer // Return a copy to avoid data race if the slice is modified elsewhere
}
