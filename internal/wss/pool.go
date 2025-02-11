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
	conn   *websocket.Conn
	buffer []string   // Store client inputs
	mu     sync.Mutex // Protect inputs from concurrent access
}

type Pool struct {
	clients    map[*Client]bool
	mutex      sync.RWMutex // Protect clients map
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte // Broadcast messages to all clients (optional)
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
				client.conn.Close()
				fmt.Println("Client unregistered")
			}
			p.mutex.Unlock()

		case message := <-p.broadcast:
			p.mutex.RLock()
			for client := range p.clients {
				err := client.conn.WriteMessage(websocket.TextMessage, message)
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
	defer c.conn.Close() // Close connection on exit

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			pool.RemoveClient(c) // Remove client from pool on error
			break
		}

		c.mu.Lock()
		c.buffer = append(c.buffer, string(message)) // Store input
		c.mu.Unlock()

		fmt.Printf("Received: %s from client\n", message)

		// Example: Echo back the message (optional)
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Println("Error writing message:", err)
			pool.RemoveClient(c)
			break
		}
	}
}

func (c *Client) GetInputs() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buffer // Return a copy to avoid data race if the slice is modified elsewhere
}

var pool = NewPool()

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade:", err)
		return
	}
	defer func(conn *websocket.Conn) {
		er := conn.Close()
		if er != nil {

		}
	}(conn)

	client := &Client{conn: conn, buffer: make([]string, 0)}
	pool.AddClient(client)
	defer pool.RemoveClient(client)

	go client.ReadInputs()
	receivedInputs := client.GetInputs()
	fmt.Println("Client Inputs:", receivedInputs)

	for {
		er := conn.WriteMessage(websocket.TextMessage, []byte("Server message"))
		if er != nil {
			log.Println("Error writing message:", er)
			break
		}

	}

}
