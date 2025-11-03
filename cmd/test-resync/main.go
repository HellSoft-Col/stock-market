package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/client"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <API_KEY> [LAST_SYNC_TIME]")
		fmt.Println("Example: go run main.go TK-ANDROMEDA-2025-AVOCULTORES")
		fmt.Println("Example: go run main.go TK-ANDROMEDA-2025-AVOCULTORES 2025-01-28T12:00:00Z")
		os.Exit(1)
	}

	apiKey := os.Args[1]
	var lastSync string
	if len(os.Args) > 2 {
		lastSync = os.Args[2]
	}

	// Connect to WebSocket server
	wsClient := client.NewWebSocketClient("localhost:9001")
	if err := wsClient.Connect(); err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := wsClient.Close(); err != nil {
			fmt.Printf("Failed to close connection: %v\n", err)
		}
	}()

	fmt.Printf("Connected to WebSocket server\n")

	// Login
	loginResponse, err := wsClient.Login(apiKey)
	if err != nil {
		fmt.Printf("LOGIN failed: %v\n", err)
		return
	}

	teamName := loginResponse["team"].(string)
	fmt.Printf("✅ Logged in as: %s\n", teamName)

	// Send RESYNC message
	resyncMessage := map[string]any{
		"type": "RESYNC",
	}

	if lastSync != "" {
		resyncMessage["lastSync"] = lastSync
		fmt.Printf("🔄 Requesting resync since: %s\n", lastSync)
	} else {
		fmt.Printf("🔄 Requesting full resync (last 24 hours)\n")
	}

	if err := wsClient.SendMessage(resyncMessage); err != nil {
		fmt.Printf("Failed to send RESYNC: %v\n", err)
		return
	}

	// Listen for EVENT_DELTA response
	fmt.Printf("🔄 Waiting for EVENT_DELTA response...\n")
	if err := wsClient.SetReadDeadline(time.Now().Add(15 * time.Second)); err != nil {
		fmt.Printf("Failed to set read deadline: %v\n", err)
		return
	}

	var response map[string]any
	if err := wsClient.ReadMessage(&response); err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		return
	}

	responseType, _ := response["type"].(string)
	fmt.Printf("\n📨 Received %s message:\n", responseType)

	prettyResponse, _ := json.MarshalIndent(response, "", "  ")
	fmt.Printf("%s\n", string(prettyResponse))

	switch responseType {
	case "EVENT_DELTA":
		if events, ok := response["events"].([]interface{}); ok {
			fmt.Printf("\n✅ RESYNC SUCCESSFUL!\n")
			fmt.Printf("📊 Received %d events\n", len(events))

			if len(events) > 0 {
				fmt.Printf("\nEvent details:\n")
				for i, event := range events {
					eventData, _ := json.MarshalIndent(event, "", "  ")
					fmt.Printf("Event %d:\n%s\n\n", i+1, string(eventData))
				}
			} else {
				fmt.Printf("No events found since last sync\n")
			}
		}
	case "ERROR":
		fmt.Printf("❌ RESYNC FAILED!\n")
		if reason, ok := response["reason"].(string); ok {
			fmt.Printf("Reason: %s\n", reason)
		}
	}
}
