package main

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	// Connect to WebSocket server
	u := url.URL{Scheme: "ws", Host: "localhost:9001", Path: "/ws"}
	fmt.Printf("Connecting to %s\n", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("Failed to close connection: %v\n", err)
		}
	}()

	fmt.Println("Connected to WebSocket server")

	// Send test message
	testMessage := map[string]interface{}{
		"type":    "LOGIN",
		"token":   "TK-TEST-123",
		"message": "Hello from test client",
	}

	if err := conn.WriteJSON(testMessage); err != nil {
		fmt.Printf("Failed to send message: %v\n", err)
		return
	}

	fmt.Printf("Sent: %+v\n", testMessage)

	// Set up interrupt handler
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Read responses
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			var response map[string]interface{}
			err := conn.ReadJSON(&response)
			if err != nil {
				fmt.Printf("Read error: %v\n", err)
				return
			}
			fmt.Printf("Received: %+v\n", response)
		}
	}()

	// Keep connection alive and handle graceful shutdown
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				fmt.Printf("Ping error: %v\n", err)
				return
			}
		case <-interrupt:
			fmt.Println("Interrupt received, closing connection...")
			err := conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			)
			if err != nil {
				fmt.Printf("Error sending close message: %v\n", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
