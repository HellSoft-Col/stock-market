package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test_login.go <API_KEY>")
		fmt.Println("Example: go run test_login.go TK-ANDROMEDA-2025-AVOCULTORES")
		os.Exit(1)
	}

	apiKey := os.Args[1]

	// Connect to WebSocket server
	u := url.URL{Scheme: "ws", Host: "localhost:9001", Path: "/ws"}
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

	fmt.Printf("Connected to WebSocket server, testing LOGIN with: %s\n", apiKey)

	// Send LOGIN message
	loginMessage := map[string]any{
		"type":  "LOGIN",
		"token": apiKey,
		"tz":    "America/Bogota",
	}

	if err := conn.WriteJSON(loginMessage); err != nil {
		fmt.Printf("Failed to send message: %v\n", err)
		return
	}

	fmt.Printf("Sent LOGIN message\n")

	// Read response with timeout
	if err := conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		fmt.Printf("Failed to set read deadline: %v\n", err)
		return
	}

	var response map[string]any
	if err := conn.ReadJSON(&response); err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		return
	}

	// Pretty print the response
	prettyResponse, _ := json.MarshalIndent(response, "", "  ")
	fmt.Printf("Received response:\n%s\n", string(prettyResponse))

	// Check if login was successful
	responseType, ok := response["type"].(string)
	if !ok {
		fmt.Printf("\n⚠️ Invalid response format\n")
		return
	}

	switch responseType {
	case "LOGIN_OK":
		fmt.Printf("\n✅ LOGIN SUCCESSFUL!\n")
		if team, ok := response["team"].(string); ok {
			fmt.Printf("Team: %s\n", team)
		}
		if species, ok := response["species"].(string); ok {
			fmt.Printf("Species: %s\n", species)
		}
	case "ERROR":
		fmt.Printf("\n❌ LOGIN FAILED!\n")
		if reason, ok := response["reason"].(string); ok {
			fmt.Printf("Reason: %s\n", reason)
		}
	}
}
