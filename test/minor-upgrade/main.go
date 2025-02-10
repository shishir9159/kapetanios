package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type clusterHealth struct {
	// todo: whose responsibility is etcdStatus bool?
	EtcdStatus          bool   `json:"etcdStatus"`
	StorageAvailability uint64 `json:"storageAvailability"`
	Err                 string `json:"err"`
}

func main() {
	url := "ws://kapetanios.com:80/minor-upgrade"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer func(conn *websocket.Conn) {
		er := conn.Close()
		if er != nil {

		}
	}(conn)

	fmt.Println("Connected to WebSocket server")

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		var c clusterHealth
		if err := json.Unmarshal(message, &c); err != nil {
			log.Printf("Error parsing JSON: %v", err)
			continue
		}

		fmt.Printf("Received: Type=%t, Data=%d\n", c.EtcdStatus, c.StorageAvailability)
	}
}
