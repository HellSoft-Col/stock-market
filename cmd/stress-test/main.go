package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/client"
)

type TestResult struct {
	ClientID      int
	Success       bool
	Error         error
	OrdersSent    int
	FillsReceived int
	Duration      time.Duration
}

var products = []string{"FOSFO", "PITA", "GUACA", "PALTA-OIL", "SEBO"}
var sides = []string{"BUY", "SELL"}
var testTokens = []string{
	"TK-ANDROMEDA-2025-AVOCULTORES",
	"TK-ORION-2025-MONJES",
	"TK-VEGA-2025-ALQUIMISTAS",
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <NUM_CLIENTS> [DURATION_SECONDS]")
		fmt.Println("Example: go run main.go 50 30")
		os.Exit(1)
	}

	var numClients int
	durationSeconds := 30

	if _, err := fmt.Sscanf(os.Args[1], "%d", &numClients); err != nil {
		fmt.Printf("Invalid number of clients: %v\n", err)
		os.Exit(1)
	}
	if len(os.Args) > 2 {
		if _, err := fmt.Sscanf(os.Args[2], "%d", &durationSeconds); err != nil {
			fmt.Printf("Invalid duration: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("🧪 Starting stress test with %d clients for %d seconds\n", numClients, durationSeconds)
	fmt.Printf("🎯 Target server: ws://localhost:9001/ws\n")
	fmt.Printf("⏰ Test duration: %d seconds\n", durationSeconds)
	fmt.Println("=====================================")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(durationSeconds)*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	results := make(chan TestResult, numClients)

	// Start all clients concurrently
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			result := runClient(ctx, clientID)
			results <- result
		}(i)

		// Stagger client starts slightly to avoid overwhelming the server
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for all clients to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and analyze results
	successCount := 0
	totalOrders := 0
	totalFills := 0

	fmt.Printf("📊 Collecting results...\n")

	for result := range results {

		if result.Success {
			successCount++
			fmt.Printf("✅ Client %d: %d orders, %d fills, %.2fs\n",
				result.ClientID, result.OrdersSent, result.FillsReceived, result.Duration.Seconds())
		} else {
			fmt.Printf("❌ Client %d failed: %v\n", result.ClientID, result.Error)
		}

		totalOrders += result.OrdersSent
		totalFills += result.FillsReceived
	}

	// Print summary
	fmt.Println("\n📈 STRESS TEST SUMMARY")
	fmt.Println("=====================")
	fmt.Printf("Total clients: %d\n", numClients)
	fmt.Printf("Successful clients: %d (%.1f%%)\n", successCount, float64(successCount)/float64(numClients)*100)
	fmt.Printf("Failed clients: %d\n", numClients-successCount)
	fmt.Printf("Total orders sent: %d\n", totalOrders)
	fmt.Printf("Total fills received: %d\n", totalFills)
	fmt.Printf("Fill rate: %.1f%%\n", float64(totalFills)/float64(totalOrders)*100)
	fmt.Printf("Orders per second: %.1f\n", float64(totalOrders)/float64(durationSeconds))

	if successCount == numClients {
		fmt.Println("\n🎉 ALL TESTS PASSED!")
	} else {
		fmt.Printf("\n⚠️  %d clients failed\n", numClients-successCount)
	}
}

func runClient(ctx context.Context, clientID int) TestResult {
	start := time.Now()
	result := TestResult{
		ClientID: clientID,
		Success:  false,
	}

	// Connect to server
	wsClient := client.NewWebSocketClient("localhost:9001")
	if err := wsClient.Connect(); err != nil {
		result.Error = fmt.Errorf("connect failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}
	defer func() {
		_ = wsClient.Close() // Error is expected in stress test
	}()

	// Login with random token
	token := testTokens[clientID%len(testTokens)]
	_, err := wsClient.Login(token)
	if err != nil {
		result.Error = fmt.Errorf("login failed: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Generate random orders until context is cancelled
	orderCount := 0
	fillCount := 0

	ticker := time.NewTicker(time.Duration(rand.Intn(500)+100) * time.Millisecond) // 100-600ms between orders
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			result.Success = true
			result.OrdersSent = orderCount
			result.FillsReceived = fillCount
			result.Duration = time.Since(start)
			return result

		case <-ticker.C:
			// Send random order
			if err := sendRandomOrder(wsClient, clientID, orderCount); err != nil {
				result.Error = fmt.Errorf("order %d failed: %w", orderCount, err)
				result.Duration = time.Since(start)
				return result
			}
			orderCount++

			// Try to read any response (non-blocking)
			_ = wsClient.SetReadDeadline(time.Now().Add(10 * time.Millisecond)) // Ignore deadline errors in stress test
			var response map[string]any
			if err := wsClient.ReadMessage(&response); err == nil {
				if responseType, ok := response["type"].(string); ok && responseType == "FILL" {
					fillCount++
				}
			}

			// Reset ticker for next order (vary the timing)
			ticker.Reset(time.Duration(rand.Intn(500)+100) * time.Millisecond)
		}
	}
}

func sendRandomOrder(wsClient *client.WebSocketClient, clientID, orderNum int) error {
	product := products[rand.Intn(len(products))]
	side := sides[rand.Intn(len(sides))]
	qty := rand.Intn(10) + 1

	orderID := fmt.Sprintf("ORD-Client%d-%d-%d", clientID, time.Now().Unix(), orderNum)

	orderMsg := map[string]any{
		"type":    "ORDER",
		"clOrdID": orderID,
		"side":    side,
		"mode":    "MARKET",
		"product": product,
		"qty":     qty,
		"message": fmt.Sprintf("Stress test order from client %d", clientID),
	}

	return wsClient.SendMessage(orderMsg)
}
