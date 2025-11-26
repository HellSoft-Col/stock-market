package strategy

import (
	"context"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/autoclient/market"
	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/rs/zerolog/log"
)

// LiquidityProviderStrategy quickly fills student orders to ensure market functionality
type LiquidityProviderStrategy struct {
	name string

	// Configuration
	fillRate         float64  // Probability of accepting offers (0.0-1.0)
	priceImprovement float64  // How much better than market (e.g., 0.005 = 0.5%)
	targetProducts   []string // Products to provide liquidity for
	responseTime     time.Duration

	// State
	health StrategyHealth
}

// NewLiquidityProviderStrategy creates a new liquidity provider strategy
func NewLiquidityProviderStrategy(name string) *LiquidityProviderStrategy {
	return &LiquidityProviderStrategy{
		name: name,
		health: StrategyHealth{
			Status:     HealthStatusHealthy,
			LastUpdate: time.Now(),
		},
	}
}

// Name returns the strategy name
func (s *LiquidityProviderStrategy) Name() string {
	return s.name
}

// Initialize initializes the strategy with configuration
func (s *LiquidityProviderStrategy) Initialize(config map[string]interface{}) error {
	s.fillRate = GetConfigFloat(config, "fillRate", 0.8)                   // 80% fill rate default
	s.priceImprovement = GetConfigFloat(config, "priceImprovement", 0.005) // 0.5% improvement
	s.responseTime = GetConfigDuration(config, "responseTime", 500*time.Millisecond)
	s.targetProducts = GetConfigStringSlice(config, "targetProducts")

	log.Info().
		Str("strategy", s.name).
		Float64("fillRate", s.fillRate).
		Float64("priceImprovement", s.priceImprovement).
		Strs("targetProducts", s.targetProducts).
		Msg("Liquidity provider strategy initialized")

	return nil
}

// OnLogin is called when connected and logged in
func (s *LiquidityProviderStrategy) OnLogin(ctx context.Context, loginInfo *domain.LoginOKMessage) error {
	// Initialize message generator for funny order messages
	InitMessageGenerator(loginInfo.Team, "")

	log.Info().
		Str("strategy", s.name).
		Str("team", loginInfo.Team).
		Msg("Liquidity provider logged in")
	return nil
}

// OnTicker is called when ticker updates arrive
func (s *LiquidityProviderStrategy) OnTicker(ctx context.Context, ticker *domain.TickerMessage) error {
	return nil
}

// OnFill is called when a fill notification arrives
func (s *LiquidityProviderStrategy) OnFill(ctx context.Context, fill *domain.FillMessage) error {
	log.Info().
		Str("strategy", s.name).
		Str("side", fill.Side).
		Str("product", fill.Product).
		Int("qty", fill.FillQty).
		Float64("price", fill.FillPrice).
		Msg("Liquidity provided")
	return nil
}

// OnOffer is called when an offer request arrives
func (s *LiquidityProviderStrategy) OnOffer(ctx context.Context, offer *domain.OfferMessage) (*OfferResponse, error) {
	// Check if we provide liquidity for this product
	if len(s.targetProducts) > 0 {
		found := false
		for _, product := range s.targetProducts {
			if product == offer.Product || product == "ALL" {
				found = true
				break
			}
		}
		if !found {
			return &OfferResponse{
				Accept: false,
				Reason: "Not providing liquidity for this product",
			}, nil
		}
	}

	// Accept based on fill rate
	if s.fillRate < 1.0 {
		// Simulate acceptance probability
		// In real implementation, could use randomness
		// For now, accept if fillRate >= 0.5
		if s.fillRate < 0.5 {
			return &OfferResponse{
				Accept: false,
				Reason: "Fill rate threshold not met",
			}, nil
		}
	}

	// Calculate improved price
	improvedPrice := offer.MaxPrice * (1 - s.priceImprovement)

	log.Info().
		Str("strategy", s.name).
		Str("product", offer.Product).
		Int("qty", offer.QuantityRequested).
		Float64("maxPrice", offer.MaxPrice).
		Float64("improvedPrice", improvedPrice).
		Msg("Accepting offer with price improvement")

	return &OfferResponse{
		Accept:          true,
		QuantityOffered: offer.QuantityRequested,
		PriceOffered:    improvedPrice,
		Reason:          "Liquidity provision",
	}, nil
}

// OnInventoryUpdate is called when inventory changes
func (s *LiquidityProviderStrategy) OnInventoryUpdate(ctx context.Context, inventory map[string]int) error {
	return nil
}

// OnBalanceUpdate is called when balance changes
func (s *LiquidityProviderStrategy) OnBalanceUpdate(ctx context.Context, balance float64) error {
	return nil
}

// OnOrderBookUpdate is called when orderbook updates arrive
func (s *LiquidityProviderStrategy) OnOrderBookUpdate(
	ctx context.Context,
	orderbook *domain.OrderBookUpdateMessage,
) error {
	return nil
}

// Execute is called periodically to generate trading actions
func (s *LiquidityProviderStrategy) Execute(ctx context.Context, state *market.MarketState) ([]*Action, error) {
	// Liquidity provider primarily responds to offers, not proactive trading
	// Could add market orders here if needed

	s.health.LastUpdate = time.Now()
	return nil, nil
}

// Health returns the strategy's current health status
func (s *LiquidityProviderStrategy) Health() StrategyHealth {
	return s.health
}
