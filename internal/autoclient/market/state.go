package market

import (
	"sync"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/domain"
)

// MarketState represents the current state of the market and portfolio
type MarketState struct {
	mu sync.RWMutex

	// Identity
	TeamName string
	Species  string

	// Portfolio
	Balance        float64
	InitialBalance float64
	Inventory      map[string]int

	// Market Data
	Tickers       map[string]*domain.TickerMessage
	OrderBooks    map[string]*OrderBook
	ActiveOrders  map[string]*domain.OrderSummary
	RecentFills   []*domain.FillMessage
	PendingOffers map[string]*domain.OfferMessage

	// Metadata
	LastUpdate time.Time
	Recipes    map[string]domain.Recipe
	Role       domain.TeamRole
}

// OrderBook represents an orderbook for a product
type OrderBook struct {
	Product    string
	BuyOrders  []*domain.OrderSummary
	SellOrders []*domain.OrderSummary
	BestBid    *float64
	BestAsk    *float64
	Spread     float64
	UpdatedAt  time.Time
}

// NewMarketState creates a new market state
func NewMarketState(teamName string) *MarketState {
	return &MarketState{
		TeamName:      teamName,
		Inventory:     make(map[string]int),
		Tickers:       make(map[string]*domain.TickerMessage),
		OrderBooks:    make(map[string]*OrderBook),
		ActiveOrders:  make(map[string]*domain.OrderSummary),
		RecentFills:   make([]*domain.FillMessage, 0),
		PendingOffers: make(map[string]*domain.OfferMessage),
		Recipes:       make(map[string]domain.Recipe),
		LastUpdate:    time.Now(),
	}
}

// UpdateBalance updates the cash balance
func (ms *MarketState) UpdateBalance(balance float64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.Balance = balance
	ms.LastUpdate = time.Now()
}

// UpdateInventory updates the entire inventory
func (ms *MarketState) UpdateInventory(inventory map[string]int) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.Inventory = inventory
	ms.LastUpdate = time.Now()
}

// AddInventory adds quantity to a product
func (ms *MarketState) AddInventory(product string, quantity int) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.Inventory[product] += quantity
	ms.LastUpdate = time.Now()
}

// RemoveInventory removes quantity from a product
func (ms *MarketState) RemoveInventory(product string, quantity int) bool {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	current := ms.Inventory[product]
	if current < quantity {
		return false
	}

	ms.Inventory[product] -= quantity
	if ms.Inventory[product] == 0 {
		delete(ms.Inventory, product)
	}

	ms.LastUpdate = time.Now()
	return true
}

// GetInventoryQuantity returns the quantity of a product in inventory
func (ms *MarketState) GetInventoryQuantity(product string) int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.Inventory[product]
}

// UpdateTicker updates the ticker for a product
func (ms *MarketState) UpdateTicker(ticker *domain.TickerMessage) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.Tickers[ticker.Product] = ticker
	ms.LastUpdate = time.Now()
}

// GetPrice returns the mid price for a product
func (ms *MarketState) GetPrice(product string) *float64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	ticker, exists := ms.Tickers[product]
	if !exists {
		return nil
	}

	return ticker.Mid
}

// AddFill records a trade execution
func (ms *MarketState) AddFill(fill *domain.FillMessage) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.RecentFills = append(ms.RecentFills, fill)

	// Keep only last 100 fills
	if len(ms.RecentFills) > 100 {
		ms.RecentFills = ms.RecentFills[len(ms.RecentFills)-100:]
	}

	ms.LastUpdate = time.Now()
}

// AddOffer records a pending offer
func (ms *MarketState) AddOffer(offer *domain.OfferMessage) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.PendingOffers[offer.OfferID] = offer
	ms.LastUpdate = time.Now()
}

// RemoveOffer removes an offer
func (ms *MarketState) RemoveOffer(offerID string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.PendingOffers, offerID)
	ms.LastUpdate = time.Now()
}

// CalculatePnL calculates profit and loss percentage
func (ms *MarketState) CalculatePnL() float64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.InitialBalance == 0 {
		return 0
	}

	// Calculate inventory value
	inventoryValue := 0.0
	for product, quantity := range ms.Inventory {
		ticker, exists := ms.Tickers[product]
		if exists && ticker.Mid != nil {
			inventoryValue += float64(quantity) * *ticker.Mid
		}
	}

	// Net worth = cash + inventory value
	netWorth := ms.Balance + inventoryValue

	// P&L% = ((current - initial) / initial) * 100
	return ((netWorth - ms.InitialBalance) / ms.InitialBalance) * 100.0
}

// GetSnapshot returns a read-only snapshot of the market state
func (ms *MarketState) GetSnapshot() *MarketState {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	// Create a shallow copy
	snapshot := &MarketState{
		TeamName:       ms.TeamName,
		Species:        ms.Species,
		Balance:        ms.Balance,
		InitialBalance: ms.InitialBalance,
		LastUpdate:     ms.LastUpdate,
		Role:           ms.Role,
	}

	// Deep copy maps
	snapshot.Inventory = make(map[string]int, len(ms.Inventory))
	for k, v := range ms.Inventory {
		snapshot.Inventory[k] = v
	}

	snapshot.Tickers = make(map[string]*domain.TickerMessage, len(ms.Tickers))
	for k, v := range ms.Tickers {
		snapshot.Tickers[k] = v
	}

	snapshot.OrderBooks = make(map[string]*OrderBook, len(ms.OrderBooks))
	for k, v := range ms.OrderBooks {
		snapshot.OrderBooks[k] = v
	}

	snapshot.ActiveOrders = make(map[string]*domain.OrderSummary, len(ms.ActiveOrders))
	for k, v := range ms.ActiveOrders {
		snapshot.ActiveOrders[k] = v
	}

	snapshot.RecentFills = make([]*domain.FillMessage, len(ms.RecentFills))
	copy(snapshot.RecentFills, ms.RecentFills)

	snapshot.PendingOffers = make(map[string]*domain.OfferMessage, len(ms.PendingOffers))
	for k, v := range ms.PendingOffers {
		snapshot.PendingOffers[k] = v
	}

	snapshot.Recipes = make(map[string]domain.Recipe, len(ms.Recipes))
	for k, v := range ms.Recipes {
		snapshot.Recipes[k] = v
	}

	return snapshot
}

// HasSufficientBalance checks if there's enough balance for a purchase
func (ms *MarketState) HasSufficientBalance(amount float64) bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.Balance >= amount
}

// HasSufficientInventory checks if there's enough inventory for a sale
func (ms *MarketState) HasSufficientInventory(product string, quantity int) bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.Inventory[product] >= quantity
}
