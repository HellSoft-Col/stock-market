package strategy

import (
	"context"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/autoclient/market"
	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/rs/zerolog/log"
)

// ArbitrageStrategy looks for price discrepancies and exploits them
// In this exchange, arbitrage opportunities might exist between:
// 1. Production cost vs market price
// 2. Buy/sell spread inefficiencies
// 3. Ingredient prices vs product prices
type ArbitrageStrategy struct {
	name string

	// Configuration
	minSpread      float64       // Minimum spread to consider (e.g., 0.02 = 2%)
	positionSize   int           // Size of each arbitrage trade
	maxPositions   int           // Max concurrent positions
	checkFrequency time.Duration // How often to check for opportunities

	// State
	lastExecute time.Time
	health      StrategyHealth
}

// NewArbitrageStrategy creates a new arbitrage strategy
func NewArbitrageStrategy(name string) *ArbitrageStrategy {
	return &ArbitrageStrategy{
		name:        name,
		lastExecute: time.Now(),
		health: StrategyHealth{
			Status:     HealthStatusHealthy,
			LastUpdate: time.Now(),
		},
	}
}

// Name returns the strategy name
func (s *ArbitrageStrategy) Name() string {
	return s.name
}

// Initialize initializes the strategy with configuration
func (s *ArbitrageStrategy) Initialize(config map[string]interface{}) error {
	s.minSpread = GetConfigFloat(config, "minSpread", 0.05) // 5% minimum spread
	s.positionSize = GetConfigInt(config, "positionSize", 50)
	s.maxPositions = GetConfigInt(config, "maxPositions", 5)
	s.checkFrequency = GetConfigDuration(config, "checkFrequency", 5*time.Second)

	log.Info().
		Str("strategy", s.name).
		Float64("minSpread", s.minSpread).
		Int("positionSize", s.positionSize).
		Msg("Arbitrage strategy initialized")

	return nil
}

// OnLogin is called when connected and logged in
func (s *ArbitrageStrategy) OnLogin(ctx context.Context, loginInfo *domain.LoginOKMessage) error {
	log.Info().
		Str("strategy", s.name).
		Str("team", loginInfo.Team).
		Msg("Arbitrage trader logged in")
	return nil
}

// OnTicker is called when ticker updates arrive
func (s *ArbitrageStrategy) OnTicker(ctx context.Context, ticker *domain.TickerMessage) error {
	return nil
}

// OnFill is called when a fill notification arrives
func (s *ArbitrageStrategy) OnFill(ctx context.Context, fill *domain.FillMessage) error {
	log.Info().
		Str("strategy", s.name).
		Str("product", fill.Product).
		Str("side", fill.Side).
		Int("qty", fill.FillQty).
		Float64("price", fill.FillPrice).
		Msg("📊 Arbitrage fill received")
	return nil
}

// OnOffer is called when an offer request arrives
func (s *ArbitrageStrategy) OnOffer(ctx context.Context, offer *domain.OfferMessage) (*OfferResponse, error) {
	// Arbitrage: Accept offers that have good spreads
	// Calculate if accepting this offer creates a profit opportunity

	// For simplicity, accept offers with good prices
	// In a real implementation, check against current market prices

	return &OfferResponse{
		Accept: false,
		Reason: "Analyzing spread",
	}, nil
}

// OnInventoryUpdate is called when inventory changes
func (s *ArbitrageStrategy) OnInventoryUpdate(ctx context.Context, inventory map[string]int) error {
	return nil
}

// OnBalanceUpdate is called when balance changes
func (s *ArbitrageStrategy) OnBalanceUpdate(ctx context.Context, balance float64) error {
	return nil
}

// OnOrderBookUpdate is called when orderbook updates arrive
func (s *ArbitrageStrategy) OnOrderBookUpdate(ctx context.Context, orderbook *domain.OrderBookUpdateMessage) error {
	return nil
}

// Execute is called periodically to generate trading actions
func (s *ArbitrageStrategy) Execute(ctx context.Context, state *market.MarketState) ([]*Action, error) {
	// Check if enough time has passed since last execution
	if time.Since(s.lastExecute) < s.checkFrequency {
		return nil, nil
	}
	s.lastExecute = time.Now()

	actions := make([]*Action, 0)

	// Look for arbitrage opportunities in the market
	// Get snapshot to avoid holding locks
	snapshot := state.GetSnapshot()

	for product, ticker := range snapshot.Tickers {
		if ticker.BestBid == nil || ticker.BestAsk == nil {
			continue
		}

		bestBid := *ticker.BestBid
		bestAsk := *ticker.BestAsk

		if bestBid == 0 || bestAsk == 0 {
			continue
		}

		// Calculate spread
		spread := (bestAsk - bestBid) / bestBid

		// Check if spread is wide enough for arbitrage
		if spread > s.minSpread {
			log.Info().
				Str("product", product).
				Float64("spread", spread*100).
				Float64("bid", bestBid).
				Float64("ask", bestAsk).
				Msg("🔍 Arbitrage opportunity detected")

			// Place buy at bid and sell at ask (market making with profit)
			// This is a simple form of arbitrage - in reality, you'd need to check:
			// 1. Production costs vs market prices
			// 2. Cross-product arbitrage (ingredients vs products)
			// 3. Temporal arbitrage (price movements)

			// For now, we implement simple spread capture
			currentInventory := snapshot.GetInventoryQuantity(product)

			// Buy low
			if snapshot.HasSufficientBalance(bestBid * float64(s.positionSize)) {
				actions = append(actions, &Action{
					Type:  ActionTypeOrder,
					Order: CreateLimitBuyOrder(product, s.positionSize, bestBid, "Arbitrage buy"),
				})
			}

			// Sell high (if we have inventory)
			if currentInventory >= s.positionSize {
				actions = append(actions, &Action{
					Type:  ActionTypeOrder,
					Order: CreateLimitSellOrder(product, s.positionSize, bestAsk, "Arbitrage sell"),
				})
			}
		}
	}

	// Recipe-based arbitrage: Check if producing is cheaper than buying
	for recipeName, recipe := range snapshot.Recipes {
		productionCost := s.calculateProductionCost(snapshot, recipe)
		if productionCost == 0 {
			continue // Can't produce or missing price data
		}

		// Get market price for the product (recipeName is the product name)
		marketPrice := s.getMarketPrice(snapshot, recipeName)
		if marketPrice == 0 {
			continue
		}

		profitMargin := (marketPrice - productionCost) / productionCost

		if profitMargin > s.minSpread {
			log.Info().
				Str("recipe", recipeName).
				Str("product", recipeName).
				Float64("productionCost", productionCost).
				Float64("marketPrice", marketPrice).
				Float64("profitMargin", profitMargin*100).
				Msg("🏭 Production arbitrage opportunity")

			// In a full implementation:
			// 1. Buy ingredients at market price
			// 2. Produce the product
			// 3. Sell at market price
			// For now, just log the opportunity
		}
	}

	return actions, nil
}

// calculateProductionCost calculates the cost to produce a product
func (s *ArbitrageStrategy) calculateProductionCost(state *market.MarketState, recipe domain.Recipe) float64 {
	if len(recipe.Ingredients) == 0 {
		// Basic production - no ingredient cost
		return 0
	}

	totalCost := 0.0
	for ingredient, qty := range recipe.Ingredients {
		price := s.getMarketPrice(state, ingredient)
		if price == 0 {
			return 0 // Can't calculate without prices
		}
		totalCost += price * float64(qty)
	}

	return totalCost
}

// getMarketPrice gets the mid price for a product
func (s *ArbitrageStrategy) getMarketPrice(state *market.MarketState, product string) float64 {
	ticker, exists := state.Tickers[product]
	if !exists {
		return 0
	}

	if ticker.Mid != nil && *ticker.Mid > 0 {
		return *ticker.Mid
	}

	if ticker.BestBid != nil && ticker.BestAsk != nil {
		return (*ticker.BestBid + *ticker.BestAsk) / 2
	}

	return 0
}

// Health returns the strategy's current health status
func (s *ArbitrageStrategy) Health() StrategyHealth {
	s.health.LastUpdate = time.Now()
	return s.health
}
