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

	// Initialize message generator for funny order messages
	InitMessageGenerator(s.name, "")

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
	// Check if this is one of our products
	isOurProduct := false
	for _, product := range s.products {
		if product == offer.Product {
			isOurProduct = true
			break
		}
	}

	if !isOurProduct {
		return &OfferResponse{
			Accept: false,
			Reason: "Not a product we make markets in",
		}, nil
	}

	// Accept offers to rebalance inventory when we're short
	// Note: We can't check current inventory here since we don't track it in strategy
	// But we can accept offers opportunistically to provide liquidity

	// Accept 50% of offers to provide two-sided liquidity
	// In production, would check inventory levels
	accept := true // For now, accept all offers for our products

	if accept {
		log.Info().
			Str("strategy", s.name).
			Str("product", offer.Product).
			Int("qty", offer.QuantityRequested).
			Float64("price", offer.MaxPrice).
			Msg("Market maker accepting offer for liquidity")

		return &OfferResponse{
			Accept:          true,
			QuantityOffered: offer.QuantityRequested,
			PriceOffered:    offer.MaxPrice,
			Reason:          "Providing liquidity",
		}, nil
	}

	return &OfferResponse{
		Accept: false,
		Reason: "Inventory level adequate",
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
		var midPrice float64

		if price == nil {
			// No market price yet - use a reasonable starting price
			midPrice = getDefaultPrice(product)
			log.Debug().
				Str("strategy", s.name).
				Str("product", product).
				Float64("defaultPrice", midPrice).
				Msg("Using default price (no market data)")
		} else {
			midPrice = *price
		}

		inventory := state.GetInventoryQuantity(product)

		// Calculate bid/ask prices with tighter spread for more liquidity
		bidPrice := midPrice * (1 - s.spread/2)
		askPrice := midPrice * (1 + s.spread/2)

		// Place two-sided quotes to create liquid market
		// BUY side: Always place to provide bids
		if inventory < s.maxInventory {
			actions = append(actions, &Action{
				Type:  ActionTypeOrder,
				Order: CreateLimitBuyOrder(product, s.quoteSize, bidPrice, "MM BID"),
			})

			log.Debug().
				Str("strategy", s.name).
				Str("product", product).
				Float64("bidPrice", bidPrice).
				Int("bidQty", s.quoteSize).
				Msg("Placing BID quote")
		}

		// SELL side: Only place if we have inventory
		// Can't sell what we don't have!
		if inventory > 0 {
			sellQty := s.quoteSize
			if inventory < sellQty {
				sellQty = inventory
			}

			actions = append(actions, &Action{
				Type:  ActionTypeOrder,
				Order: CreateLimitSellOrder(product, sellQty, askPrice, "MM ASK"),
			})

			log.Debug().
				Str("strategy", s.name).
				Str("product", product).
				Float64("askPrice", askPrice).
				Int("askQty", sellQty).
				Int("inventory", inventory).
				Msg("Placing ASK quote")
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
