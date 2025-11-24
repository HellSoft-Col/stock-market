package strategy

import (
	"context"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/autoclient/market"
	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/rs/zerolog/log"
)

// MarketMakerStrategy provides continuous liquidity by placing limit orders
type MarketMakerStrategy struct {
	name string

	// Configuration
	spread          float64
	quoteSize       int
	maxInventory    int
	updateFrequency time.Duration
	products        []string

	// State
	lastUpdate time.Time
	health     StrategyHealth
}

// NewMarketMakerStrategy creates a new market maker strategy
func NewMarketMakerStrategy(name string) *MarketMakerStrategy {
	return &MarketMakerStrategy{
		name: name,
		health: StrategyHealth{
			Status:     HealthStatusHealthy,
			LastUpdate: time.Now(),
		},
	}
}

// Name returns the strategy name
func (s *MarketMakerStrategy) Name() string {
	return s.name
}

// Initialize initializes the strategy with configuration
func (s *MarketMakerStrategy) Initialize(config map[string]interface{}) error {
	s.spread = GetConfigFloat(config, "spread", 0.02) // 2% default
	s.quoteSize = GetConfigInt(config, "quoteSize", 100)
	s.maxInventory = GetConfigInt(config, "maxInventory", 1000)
	s.updateFrequency = GetConfigDuration(config, "updateFrequency", 5*time.Second)
	s.products = GetConfigStringSlice(config, "products")

	log.Info().
		Str("strategy", s.name).
		Float64("spread", s.spread).
		Int("quoteSize", s.quoteSize).
		Strs("products", s.products).
		Msg("Market maker strategy initialized")

	return nil
}

// OnLogin is called when connected and logged in
func (s *MarketMakerStrategy) OnLogin(ctx context.Context, loginInfo *domain.LoginOKMessage) error {
	log.Info().
		Str("strategy", s.name).
		Str("team", loginInfo.Team).
		Msg("Market maker logged in")
	return nil
}

// OnTicker is called when ticker updates arrive
func (s *MarketMakerStrategy) OnTicker(ctx context.Context, ticker *domain.TickerMessage) error {
	return nil
}

// OnFill is called when a fill notification arrives
func (s *MarketMakerStrategy) OnFill(ctx context.Context, fill *domain.FillMessage) error {
	log.Info().
		Str("strategy", s.name).
		Str("side", fill.Side).
		Str("product", fill.Product).
		Int("qty", fill.FillQty).
		Float64("price", fill.FillPrice).
		Msg("Market maker fill")
	return nil
}

// OnOffer is called when an offer request arrives
func (s *MarketMakerStrategy) OnOffer(ctx context.Context, offer *domain.OfferMessage) (*OfferResponse, error) {
	// Market maker doesn't respond to offers (uses limit orders)
	return &OfferResponse{
		Accept: false,
		Reason: "Market maker uses limit orders",
	}, nil
}

// OnInventoryUpdate is called when inventory changes
func (s *MarketMakerStrategy) OnInventoryUpdate(ctx context.Context, inventory map[string]int) error {
	return nil
}

// OnBalanceUpdate is called when balance changes
func (s *MarketMakerStrategy) OnBalanceUpdate(ctx context.Context, balance float64) error {
	return nil
}

// OnOrderBookUpdate is called when orderbook updates arrive
func (s *MarketMakerStrategy) OnOrderBookUpdate(ctx context.Context, orderbook *domain.OrderBookUpdateMessage) error {
	return nil
}

// Execute is called periodically to generate trading actions
func (s *MarketMakerStrategy) Execute(ctx context.Context, state *market.MarketState) ([]*Action, error) {
	// Check if it's time to update quotes
	if time.Since(s.lastUpdate) < s.updateFrequency {
		return nil, nil
	}

	actions := []*Action{}

	// Place limit orders for each product
	for _, product := range s.products {
		price := state.GetPrice(product)
		if price == nil {
			continue
		}

		midPrice := *price
		inventory := state.GetInventoryQuantity(product)

		// Calculate bid/ask prices based on spread
		bidPrice := midPrice * (1 - s.spread/2)
		askPrice := midPrice * (1 + s.spread/2)

		// Place buy order if inventory not too high
		if inventory < s.maxInventory {
			actions = append(actions, &Action{
				Type:  ActionTypeOrder,
				Order: CreateLimitBuyOrder(product, s.quoteSize, bidPrice, "MM buy"),
			})
		}

		// Place sell order if have inventory
		if inventory > 0 {
			sellQty := s.quoteSize
			if sellQty > inventory {
				sellQty = inventory
			}
			actions = append(actions, &Action{
				Type:  ActionTypeOrder,
				Order: CreateLimitSellOrder(product, sellQty, askPrice, "MM sell"),
			})
		}
	}

	s.lastUpdate = time.Now()
	s.health.LastUpdate = time.Now()

	if len(actions) > 0 {
		log.Info().
			Str("strategy", s.name).
			Int("actions", len(actions)).
			Msg("Market maker placing quotes")
	}

	return actions, nil
}

// Health returns the strategy's current health status
func (s *MarketMakerStrategy) Health() StrategyHealth {
	return s.health
}
