package strategy

import (
	"context"
	"math/rand"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/autoclient/market"
	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/rs/zerolog/log"
)

// RandomTraderStrategy creates market chaos with unpredictable trades
type RandomTraderStrategy struct {
	name string

	// Configuration
	minInterval  time.Duration
	maxInterval  time.Duration
	orderSizeMin int
	orderSizeMax int

	// State
	lastTrade time.Time
	nextTrade time.Time
	rng       *rand.Rand
	health    StrategyHealth
}

// NewRandomTraderStrategy creates a new random trader strategy
func NewRandomTraderStrategy(name string) *RandomTraderStrategy {
	return &RandomTraderStrategy{
		name: name,
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
		health: StrategyHealth{
			Status:     HealthStatusHealthy,
			LastUpdate: time.Now(),
		},
	}
}

// Name returns the strategy name
func (s *RandomTraderStrategy) Name() string {
	return s.name
}

// Initialize initializes the strategy with configuration
func (s *RandomTraderStrategy) Initialize(config map[string]interface{}) error {
	s.minInterval = GetConfigDuration(config, "minInterval", 10*time.Second)
	s.maxInterval = GetConfigDuration(config, "maxInterval", 30*time.Second)
	s.orderSizeMin = GetConfigInt(config, "orderSizeMin", 50)
	s.orderSizeMax = GetConfigInt(config, "orderSizeMax", 200)

	// Schedule first trade
	s.scheduleNextTrade()

	log.Info().
		Str("strategy", s.name).
		Dur("minInterval", s.minInterval).
		Dur("maxInterval", s.maxInterval).
		Msg("Random trader strategy initialized")

	return nil
}

// scheduleNextTrade schedules the next random trade
func (s *RandomTraderStrategy) scheduleNextTrade() {
	intervalRange := s.maxInterval - s.minInterval
	randomDuration := s.minInterval + time.Duration(s.rng.Int63n(int64(intervalRange)))
	s.nextTrade = time.Now().Add(randomDuration)

	log.Debug().
		Str("strategy", s.name).
		Time("nextTrade", s.nextTrade).
		Dur("wait", randomDuration).
		Msg("Next random trade scheduled")
}

// OnLogin is called when connected and logged in
func (s *RandomTraderStrategy) OnLogin(ctx context.Context, loginInfo *domain.LoginOKMessage) error {
	log.Info().
		Str("strategy", s.name).
		Str("team", loginInfo.Team).
		Msg("Random trader logged in")
	return nil
}

// OnTicker is called when ticker updates arrive
func (s *RandomTraderStrategy) OnTicker(ctx context.Context, ticker *domain.TickerMessage) error {
	return nil
}

// OnFill is called when a fill notification arrives
func (s *RandomTraderStrategy) OnFill(ctx context.Context, fill *domain.FillMessage) error {
	log.Info().
		Str("strategy", s.name).
		Str("side", fill.Side).
		Str("product", fill.Product).
		Int("qty", fill.FillQty).
		Msg("Random trade executed")
	return nil
}

// OnOffer is called when an offer request arrives
func (s *RandomTraderStrategy) OnOffer(ctx context.Context, offer *domain.OfferMessage) (*OfferResponse, error) {
	// Randomly accept or reject (50/50)
	accept := s.rng.Float64() < 0.5

	return &OfferResponse{
		Accept:          accept,
		QuantityOffered: offer.QuantityRequested,
		PriceOffered:    offer.MaxPrice,
		Reason:          "Random decision",
	}, nil
}

// OnInventoryUpdate is called when inventory changes
func (s *RandomTraderStrategy) OnInventoryUpdate(ctx context.Context, inventory map[string]int) error {
	return nil
}

// OnBalanceUpdate is called when balance changes
func (s *RandomTraderStrategy) OnBalanceUpdate(ctx context.Context, balance float64) error {
	return nil
}

// OnOrderBookUpdate is called when orderbook updates arrive
func (s *RandomTraderStrategy) OnOrderBookUpdate(ctx context.Context, orderbook *domain.OrderBookUpdateMessage) error {
	return nil
}

// Execute is called periodically to generate trading actions
func (s *RandomTraderStrategy) Execute(ctx context.Context, state *market.MarketState) ([]*Action, error) {
	// Check if it's time for a random trade
	if time.Now().Before(s.nextTrade) {
		return nil, nil
	}

	snapshot := state.GetSnapshot()

	// Get available products (ones with prices)
	availableProducts := []string{}
	for product := range snapshot.Tickers {
		availableProducts = append(availableProducts, product)
	}

	if len(availableProducts) == 0 {
		s.scheduleNextTrade()
		return nil, nil
	}

	// Pick a random product
	product := availableProducts[s.rng.Intn(len(availableProducts))]

	// Random quantity
	quantity := s.orderSizeMin + s.rng.Intn(s.orderSizeMax-s.orderSizeMin+1)

	// Random side (buy or sell)
	isBuy := s.rng.Float64() < 0.5

	var action *Action

	if isBuy {
		// Random buy
		action = &Action{
			Type:  ActionTypeOrder,
			Order: CreateBuyOrder(product, quantity, "Random buy"),
		}

		log.Info().
			Str("strategy", s.name).
			Str("product", product).
			Int("qty", quantity).
			Msg("Random BUY order")
	} else {
		// Random sell (only if have inventory)
		inventory := snapshot.Inventory[product]
		if inventory > 0 {
			sellQty := quantity
			if sellQty > inventory {
				sellQty = inventory
			}

			action = &Action{
				Type:  ActionTypeOrder,
				Order: CreateSellOrder(product, sellQty, "Random sell"),
			}

			log.Info().
				Str("strategy", s.name).
				Str("product", product).
				Int("qty", sellQty).
				Msg("Random SELL order")
		}
	}

	s.lastTrade = time.Now()
	s.scheduleNextTrade()
	s.health.LastUpdate = time.Now()

	if action != nil {
		return []*Action{action}, nil
	}

	return nil, nil
}

// Health returns the strategy's current health status
func (s *RandomTraderStrategy) Health() StrategyHealth {
	return s.health
}
