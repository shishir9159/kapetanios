package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

func wsclient() {
	serverAddr := "ws://localhost:8080/ws"
	conn, _, err := websocket.DefaultDialer.Dial(serverAddr, nil)
	if err != nil {
		log.Fatal("error connecting to webSocket server:", err)
	}
	defer func(conn *websocket.Conn) {
		er := conn.Close()
		if er != nil {
			log.Println("error closing webSocket connection:", er)
		}
	}(conn)

	steps := []string{"step 1", "step 2", "step 3", "step 4", "step 5"}

	for _, step := range steps {
		if er := conn.WriteMessage(websocket.TextMessage, []byte(step)); er != nil {
			log.Println("error writing message:", er)
			return
		}

		_, msg, er := conn.ReadMessage()
		if er != nil {
			log.Println("error reading message:", er)
			return
		}
		fmt.Println("received:", string(msg))

		time.Sleep(1 * time.Second)
	}
}
