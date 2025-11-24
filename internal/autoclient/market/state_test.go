package market

import (
	"testing"

	"github.com/HellSoft-Col/stock-market/internal/domain"
)

func TestNewMarketState(t *testing.T) {
	state := NewMarketState("test-team")

	if state.TeamName != "test-team" {
		t.Errorf("Expected team name 'test-team', got '%s'", state.TeamName)
	}

	if state.Inventory == nil {
		t.Error("Expected inventory to be initialized")
	}

	if state.Tickers == nil {
		t.Error("Expected tickers to be initialized")
	}
}

func TestUpdateBalance(t *testing.T) {
	state := NewMarketState("test-team")

	state.UpdateBalance(10000.0)
	if state.Balance != 10000.0 {
		t.Errorf("Expected balance 10000.0, got %f", state.Balance)
	}

	state.UpdateBalance(12500.50)
	if state.Balance != 12500.50 {
		t.Errorf("Expected balance 12500.50, got %f", state.Balance)
	}
}

func TestInventoryOperations(t *testing.T) {
	state := NewMarketState("test-team")

	// Add inventory
	state.AddInventory("PALTA-OIL", 100)
	if qty := state.GetInventoryQuantity("PALTA-OIL"); qty != 100 {
		t.Errorf("Expected 100, got %d", qty)
	}

	// Add more
	state.AddInventory("PALTA-OIL", 50)
	if qty := state.GetInventoryQuantity("PALTA-OIL"); qty != 150 {
		t.Errorf("Expected 150, got %d", qty)
	}

	// Remove inventory
	success := state.RemoveInventory("PALTA-OIL", 50)
	if !success {
		t.Error("Expected removal to succeed")
	}
	if qty := state.GetInventoryQuantity("PALTA-OIL"); qty != 100 {
		t.Errorf("Expected 100, got %d", qty)
	}

	// Try to remove more than available
	success = state.RemoveInventory("PALTA-OIL", 200)
	if success {
		t.Error("Expected removal to fail when insufficient inventory")
	}
}

func TestCalculatePnL(t *testing.T) {
	state := NewMarketState("test-team")
	state.InitialBalance = 10000.0
	state.Balance = 12000.0

	// Add inventory
	state.AddInventory("PALTA-OIL", 100)

	// Add ticker for pricing
	mid := 25.0
	state.UpdateTicker(&domain.TickerMessage{
		Product: "PALTA-OIL",
		Mid:     &mid,
	})

	// Calculate P&L
	// Net worth = 12000 + (100 * 25) = 12000 + 2500 = 14500
	// P&L% = ((14500 - 10000) / 10000) * 100 = 45%
	pnl := state.CalculatePnL()
	expected := 45.0

	if pnl < expected-0.01 || pnl > expected+0.01 {
		t.Errorf("Expected P&L around %.2f%%, got %.2f%%", expected, pnl)
	}
}

func TestHasSufficientBalance(t *testing.T) {
	state := NewMarketState("test-team")
	state.Balance = 1000.0

	if !state.HasSufficientBalance(500.0) {
		t.Error("Expected sufficient balance for 500")
	}

	if !state.HasSufficientBalance(1000.0) {
		t.Error("Expected sufficient balance for 1000")
	}

	if state.HasSufficientBalance(1500.0) {
		t.Error("Expected insufficient balance for 1500")
	}
}

func TestHasSufficientInventory(t *testing.T) {
	state := NewMarketState("test-team")
	state.AddInventory("FOSFO", 50)

	if !state.HasSufficientInventory("FOSFO", 25) {
		t.Error("Expected sufficient inventory for 25")
	}

	if !state.HasSufficientInventory("FOSFO", 50) {
		t.Error("Expected sufficient inventory for 50")
	}

	if state.HasSufficientInventory("FOSFO", 75) {
		t.Error("Expected insufficient inventory for 75")
	}

	if state.HasSufficientInventory("PITA", 10) {
		t.Error("Expected insufficient inventory for non-existent product")
	}
}

func TestGetSnapshot(t *testing.T) {
	state := NewMarketState("test-team")
	state.Balance = 10000.0
	state.AddInventory("PALTA-OIL", 100)

	snapshot := state.GetSnapshot()

	// Verify snapshot data
	if snapshot.Balance != state.Balance {
		t.Error("Snapshot balance mismatch")
	}

	if snapshot.Inventory["PALTA-OIL"] != 100 {
		t.Error("Snapshot inventory mismatch")
	}

	// Modify original state
	state.AddInventory("PALTA-OIL", 50)

	// Snapshot should be unchanged
	if snapshot.Inventory["PALTA-OIL"] != 100 {
		t.Error("Snapshot should be independent of original state")
	}
}

func TestGetPrice(t *testing.T) {
	state := NewMarketState("test-team")

	// No ticker yet
	price := state.GetPrice("PALTA-OIL")
	if price != nil {
		t.Error("Expected nil price for product without ticker")
	}

	// Add ticker
	mid := 26.0
	state.UpdateTicker(&domain.TickerMessage{
		Product: "PALTA-OIL",
		Mid:     &mid,
	})

	price = state.GetPrice("PALTA-OIL")
	if price == nil {
		t.Fatal("Expected price to be available")
	}

	if *price != 26.0 {
		t.Errorf("Expected price 26.0, got %f", *price)
	}
}

func TestAddFill(t *testing.T) {
	state := NewMarketState("test-team")

	// Add fills
	for i := 0; i < 5; i++ {
		state.AddFill(&domain.FillMessage{
			Product: "PALTA-OIL",
			FillQty: 10,
		})
	}

	if len(state.RecentFills) != 5 {
		t.Errorf("Expected 5 fills, got %d", len(state.RecentFills))
	}

	// Add more fills (should keep only last 100)
	for i := 0; i < 100; i++ {
		state.AddFill(&domain.FillMessage{
			Product: "FOSFO",
			FillQty: 5,
		})
	}

	if len(state.RecentFills) != 100 {
		t.Errorf("Expected max 100 fills, got %d", len(state.RecentFills))
	}

	// Most recent should be FOSFO
	if state.RecentFills[len(state.RecentFills)-1].Product != "FOSFO" {
		t.Error("Expected most recent fill to be FOSFO")
	}
}

func TestOfferManagement(t *testing.T) {
	state := NewMarketState("test-team")

	offer := &domain.OfferMessage{
		OfferID: "OFFER-123",
		Product: "PALTA-OIL",
	}

	// Add offer
	state.AddOffer(offer)
	if _, exists := state.PendingOffers["OFFER-123"]; !exists {
		t.Error("Expected offer to be added")
	}

	// Remove offer
	state.RemoveOffer("OFFER-123")
	if _, exists := state.PendingOffers["OFFER-123"]; exists {
		t.Error("Expected offer to be removed")
	}
}
