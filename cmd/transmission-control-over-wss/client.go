package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

func client() {
	serverAddr := "ws://localhost:8080/ws"
	conn, _, err := websocket.DefaultDialer.Dial(serverAddr, nil)
	if err != nil {
		log.Fatal("Error connecting to WebSocket server:", err)
	}
	defer func(conn *websocket.Conn) {
		er := conn.Close()
		if er != nil {
			log.Println("Error closing WebSocket connection:", er)
		}
	}(conn)

	steps := []string{"step 1", "step 2", "step 3", "step 4", "step 5"}

	for _, step := range steps {
		if er := conn.WriteMessage(websocket.TextMessage, []byte(step)); er != nil {
			log.Println("Error writing message:", er)
			return
		}

		_, msg, er := conn.ReadMessage()
		if er != nil {
			log.Println("Error reading message:", er)
			return
		}
		fmt.Println("Received:", string(msg))

		time.Sleep(1 * time.Second)
	}
}
