package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/yourusername/avocado-exchange-server/internal/client"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <API_KEY> <PRODUCT> [QUANTITY]")
		fmt.Println("Example: go run main.go TK-ANDROMEDA-2025-AVOCULTORES GUACA 16")
		fmt.Println("Products: GUACA, SEBO, PALTA-OIL, FOSFO, NUCREM, CASCAR-ALLOY, GTRON, H-GUACA, PITA")
		os.Exit(1)
	}

	apiKey := os.Args[1]
	product := os.Args[2]
	quantity := 1

	if len(os.Args) > 3 {
		fmt.Sscanf(os.Args[3], "%d", &quantity)
	}

	// Connect to WebSocket server
	wsClient := client.NewWebSocketClient("localhost:9001")
	if err := wsClient.Connect(); err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer wsClient.Close()

	fmt.Printf("Connected to WebSocket server\n")

	// Login
	loginResponse, err := wsClient.Login(apiKey)
	if err != nil {
		fmt.Printf("LOGIN failed: %v\n", err)
		return
	}

	teamName := loginResponse["team"].(string)
	fmt.Printf("✅ Logged in as: %s\n", teamName)

	// Send PRODUCTION_UPDATE message
	productionMessage := map[string]any{
		"type":     "PRODUCTION_UPDATE",
		"product":  product,
		"quantity": quantity,
	}

	if err := wsClient.SendMessage(productionMessage); err != nil {
		fmt.Printf("Failed to send PRODUCTION_UPDATE: %v\n", err)
		return
	}

	fmt.Printf("🏭 Sent production update: %s +%d\n", product, quantity)

	// Listen for response
	fmt.Printf("🔄 Waiting for response...\n")
	wsClient.SetReadDeadline(time.Now().Add(10 * time.Second))

	var response map[string]any
	if err := wsClient.ReadMessage(&response); err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		return
	}

	responseType, _ := response["type"].(string)
	fmt.Printf("\n📨 Received %s message:\n", responseType)

	prettyResponse, _ := json.MarshalIndent(response, "", "  ")
	fmt.Printf("%s\n", string(prettyResponse))

	if responseType == "ERROR" {
		fmt.Printf("❌ PRODUCTION UPDATE FAILED!\n")
		if reason, ok := response["reason"].(string); ok {
			fmt.Printf("Reason: %s\n", reason)
		}
	} else {
		fmt.Printf("✅ PRODUCTION UPDATE SUCCESSFUL!\n")
	}
}
