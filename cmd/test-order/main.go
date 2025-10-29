package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run main.go <API_KEY> <SIDE> <PRODUCT> [QUANTITY]")
		fmt.Println("Example: go run main.go TK-ANDROMEDA-2025-AVOCULTORES BUY FOSFO 10")
		fmt.Println("Sides: BUY, SELL")
		fmt.Println("Products: GUACA, SEBO, PALTA-OIL, FOSFO, NUCREM, CASCAR-ALLOY, GTRON, H-GUACA, PITA")
		os.Exit(1)
	}

	apiKey := os.Args[1]
	side := os.Args[2]
	product := os.Args[3]
	quantity := 10

	if len(os.Args) > 4 {
		if _, err := fmt.Sscanf(os.Args[4], "%d", &quantity); err != nil {
			fmt.Printf("Invalid quantity: %v\n", err)
			os.Exit(1)
		}
	}

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

	fmt.Printf("Connected to WebSocket server\n")

	// Send LOGIN message
	loginMessage := map[string]any{
		"type":  "LOGIN",
		"token": apiKey,
	}

	if err := conn.WriteJSON(loginMessage); err != nil {
		fmt.Printf("Failed to send LOGIN: %v\n", err)
		return
	}
	fmt.Printf("Sent LOGIN\n")

	// Read LOGIN response
	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		fmt.Printf("Failed to set read deadline: %v\n", err)
		return
	}
	var loginResponse map[string]any
	if err := conn.ReadJSON(&loginResponse); err != nil {
		fmt.Printf("Failed to read LOGIN response: %v\n", err)
		return
	}

	if responseType, ok := loginResponse["type"].(string); !ok || responseType != "LOGIN_OK" {
		fmt.Printf("LOGIN failed: %+v\n", loginResponse)
		return
	}

	teamName := loginResponse["team"].(string)
	fmt.Printf("✅ Logged in as: %s\n", teamName)

	// Generate unique order ID
	orderID := fmt.Sprintf("ORD-%s-%d-%s", teamName, time.Now().Unix(), uuid.New().String()[:8])

	// Send ORDER message
	orderMessage := map[string]any{
		"type":    "ORDER",
		"clOrdID": orderID,
		"side":    side,
		"mode":    "MARKET",
		"product": product,
		"qty":     quantity,
		"message": fmt.Sprintf("Test %s order from %s", side, teamName),
	}

	if err := conn.WriteJSON(orderMessage); err != nil {
		fmt.Printf("Failed to send ORDER: %v\n", err)
		return
	}

	fmt.Printf("📦 Sent %s order: %s %d %s (ID: %s)\n", side, side, quantity, product, orderID)

	// Listen for responses (FILL messages or errors)
	fmt.Printf("🔄 Waiting for responses...\n")
	if err := conn.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
		fmt.Printf("Failed to set read deadline: %v\n", err)
		return
	}

	responseCount := 0
	for responseCount < 5 { // Limit responses
		var response map[string]any
		if err := conn.ReadJSON(&response); err != nil {
			fmt.Printf("Failed to read response: %v\n", err)
			break
		}

		responseType, _ := response["type"].(string)
		fmt.Printf("\n📨 Received %s message:\n", responseType)

		prettyResponse, _ := json.MarshalIndent(response, "", "  ")
		fmt.Printf("%s\n", string(prettyResponse))

		if responseType == "FILL" {
			fmt.Printf("🎉 TRADE EXECUTED!\n")
			break
		} else if responseType == "ERROR" {
			fmt.Printf("❌ ORDER REJECTED!\n")
			break
		}

		responseCount++

		// Reset deadline for next message
		if err := conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
			fmt.Printf("Failed to set read deadline: %v\n", err)
			break
		}
	}

	if responseCount == 0 {
		fmt.Printf("⏰ No response received (order may be pending in book)\n")
	}
}
