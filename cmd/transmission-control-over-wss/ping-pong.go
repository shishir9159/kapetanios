package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		response := processMessage(string(msg))
		if err := conn.WriteMessage(websocket.TextMessage, []byte(response)); err != nil {
			fmt.Println("Error writing message:", err)
			break
		}
	}
}

func processMessage(msg string) string {
	switch msg {
	case "step 1":
		return "Response for step 1"
	case "step 2":
		return "Response for step 2"
	case "step 3":
		return "Response for step 3"
	case "step 4":
		return "Response for step 4"
	case "step 5":
		return "Response for step 5"
	default:
		return "Unknown step"
	}
}

func main() {
	http.HandleFunc("/ws", handleConnection)
	fmt.Println("WebSocket server started on :8080")
	http.ListenAndServe(":8080", nil)
}
