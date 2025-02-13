package wss

//
//import (
//	"fmt"
//	"github.com/gorilla/websocket"
//	"log"
//	"net/http"
//	"sync"
//)
//
//var upgrader = websocket.Upgrader{HandshakeTimeout: 0,
//	ReadBufferSize:  0,
//	WriteBufferSize: 0,
//	WriteBufferPool: nil,
//	Subprotocols:    nil,
//	Error:           nil,
//	CheckOrigin: func(r *http.Request) bool {
//		return true
//	},
//	EnableCompression: false,
//}
//
//type Client struct {
//	Conn *websocket.Conn `json:"conn"`
//	Mu   sync.Mutex      `json:"mu"`
//}
//
//type Pool struct {
//	Clients    map[*Client]bool `json:"clients"`
//	Buffer     []string         `json:"buffer"`
//	Mutex      sync.RWMutex     `json:"mutex"`
//	Register   chan *Client     `json:"register"`
//	Unregister chan *Client     `json:"unregister"`
//	Broadcast  chan []byte      `json:"broadcast"`
//}
//
//func NewPool() *Pool {
//	return &Pool{
//		Clients:    make(map[*Client]bool),
//		Register:   make(chan *Client),
//		Unregister: make(chan *Client),
//		Broadcast:  make(chan []byte),
//	}
//}
//
//func (pool *Pool) Run() {
//	for {
//		select {
//		case client := <-pool.Register:
//			//pool.Mutex.Lock()
//			pool.Clients[client] = true
//			//pool.Mutex.Unlock()
//			fmt.Println("client registered")
//
//		case client := <-pool.Unregister:
//			//pool.Mutex.Lock()
//			if _, ok := pool.Clients[client]; ok {
//				delete(pool.Clients, client)
//				err := client.Conn.Close()
//				if err != nil {
//					return
//				}
//				fmt.Println("client unregistered")
//			}
//			//pool.Mutex.Unlock()
//
//		case message := <-pool.Broadcast:
//			//pool.Mutex.RLock()
//			for client := range pool.Clients {
//				err := client.Conn.WriteMessage(websocket.TextMessage, message)
//				if err != nil {
//					log.Println("error writing message:", err)
//					pool.Unregister <- client // Unregister client on error
//				}
//			}
//			//pool.Mutex.RUnlock()
//		}
//	}
//}
//
//func (pool *Pool) AddClient(client *Client) {
//	pool.Register <- client
//}
//
//func (pool *Pool) RemoveClient(client *Client) {
//	pool.Unregister <- client
//}
//
//func (pool *Pool) BroadcastMessage(message []byte) {
//	pool.Broadcast <- message
//}
//
//func (c *Client) ReadInputs(pool *Pool) {
//	defer func(Conn *websocket.Conn) {
//		err := Conn.Close()
//		if err != nil {
//
//		}
//	}(c.Conn)
//
//	for {
//		_, message, err := c.Conn.ReadMessage()
//		if err != nil {
//			log.Println("Error reading message:", err)
//			// client is removed from the pool on error
//			pool.RemoveClient(c)
//			break
//		}
//
//		c.Mu.Lock()
//		c.Buffer = append(c.Buffer, string(message)) // Store input
//		c.Mu.Unlock()
//
//		fmt.Printf("Received: %s from client\n", message)
//
//		if err = c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
//			log.Println("Error writing message:", err)
//			pool.RemoveClient(c)
//			break
//		}
//	}
//}
//
//func (c *Client) GetInputs() []string {
//	//c.Mu.Lock()
//	//defer c.Mu.Unlock()
//	return c.Buffer // Return a copy to avoid data race if the slice is modified elsewhere
//}
