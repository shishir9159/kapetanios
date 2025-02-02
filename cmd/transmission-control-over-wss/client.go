package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

func main() {
	serverAddr := "ws://localhost:8080/ws"
	conn, _, err := websocket.DefaultDialer.Dial(serverAddr, nil)
	if err != nil {
		log.Fatal("Error connecting to WebSocket server:", err)
	}
	defer conn.Close()

	steps := []string{"step 1", "step 2", "step 3", "step 4", "step 5"}

	for _, step := range steps {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(step)); err != nil {
			log.Println("Error writing message:", err)
			return
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			return
		}
		fmt.Println("Received:", string(msg))

		time.Sleep(1 * time.Second)
	}
}
