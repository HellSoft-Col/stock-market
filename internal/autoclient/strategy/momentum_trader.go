package strategy

import (
	"context"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/autoclient/market"
	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/rs/zerolog/log"
)

// MomentumStrategy implements a momentum-based trading strategy
// Buys when prices are rising, sells when prices are falling
type MomentumStrategy struct {
	name string

	// Configuration
	products          []string
	lookbackPeriod    time.Duration
	momentumThreshold float64
	maxPosition       int
	positionSize      int
	updateFrequency   time.Duration
	stopLoss          float64
	takeProfit        float64

	// State
	priceHistory     map[string][]PricePoint
	positionSizes    map[string]int
	entryPrices      map[string]float64
	lastExecute      time.Time
	maxHistoryLength int
	health           StrategyHealth
}

// PricePoint represents a price at a specific time
type PricePoint struct {
	Price     float64
	Timestamp time.Time
}

// NewMomentumStrategy creates a new momentum trader
func NewMomentumStrategy(name string) *MomentumStrategy {
	return &MomentumStrategy{
		name:             name,
		priceHistory:     make(map[string][]PricePoint),
		positionSizes:    make(map[string]int),
		entryPrices:      make(map[string]float64),
		maxHistoryLength: 100,
		lastExecute:      time.Now(),
		health: StrategyHealth{
			Status:     HealthStatusHealthy,
			LastUpdate: time.Now(),
		},
	}
}

// Name returns the strategy name
func (s *MomentumStrategy) Name() string {
	return s.name
}

// Initialize initializes the strategy with configuration
func (s *MomentumStrategy) Initialize(config map[string]interface{}) error {
	s.products = GetConfigStringSlice(config, "products")
	s.lookbackPeriod = GetConfigDuration(config, "lookbackPeriod", 5*time.Minute)
	s.momentumThreshold = GetConfigFloat(config, "momentumThreshold", 0.03) // 3% default
	s.maxPosition = GetConfigInt(config, "maxPosition", 500)
	s.positionSize = GetConfigInt(config, "positionSize", 100)
	s.updateFrequency = GetConfigDuration(config, "updateFrequency", 10*time.Second)
	s.stopLoss = GetConfigFloat(config, "stopLoss", 0.10)     // 10% stop loss
	s.takeProfit = GetConfigFloat(config, "takeProfit", 0.15) // 15% take profit

	// Initialize price history for tracked products
	for _, product := range s.products {
		s.priceHistory[product] = make([]PricePoint, 0, s.maxHistoryLength)
		s.positionSizes[product] = 0
		s.entryPrices[product] = 0
	}

	log.Info().
		Str("strategy", s.name).
		Strs("products", s.products).
		Dur("lookback", s.lookbackPeriod).
		Float64("threshold", s.momentumThreshold).
		Msg("Momentum strategy initialized")

	return nil
}

// OnLogin is called when connected and logged in
func (s *MomentumStrategy) OnLogin(ctx context.Context, loginInfo *domain.LoginOKMessage) error {
	log.Info().
		Str("strategy", s.name).
		Str("team", loginInfo.Team).
		Msg("Momentum trader logged in")
	return nil
}

// OnTicker updates price history when ticker messages arrive
func (s *MomentumStrategy) OnTicker(ctx context.Context, ticker *domain.TickerMessage) error {
	// Only track products we're interested in
	if _, tracked := s.priceHistory[ticker.Product]; !tracked {
		return nil
	}

	// Use mid price if available, otherwise average of bid/ask
	var midPrice float64
	if ticker.Mid != nil && *ticker.Mid > 0 {
		midPrice = *ticker.Mid
	} else if ticker.BestBid != nil && ticker.BestAsk != nil {
		midPrice = (*ticker.BestBid + *ticker.BestAsk) / 2
	} else {
		return nil // No price data
	}

	// Add new price point
	s.priceHistory[ticker.Product] = append(s.priceHistory[ticker.Product], PricePoint{
		Price:     midPrice,
		Timestamp: time.Now(),
	})

	// Limit history size
	if len(s.priceHistory[ticker.Product]) > s.maxHistoryLength {
		s.priceHistory[ticker.Product] = s.priceHistory[ticker.Product][1:]
	}

	log.Debug().
		Str("product", ticker.Product).
		Float64("price", midPrice).
		Int("historySize", len(s.priceHistory[ticker.Product])).
		Msg("Price updated")

	return nil
}

// OnFill handles fill notifications (update position tracking)
func (s *MomentumStrategy) OnFill(ctx context.Context, fill *domain.FillMessage) error {
	if fill.Side == "BUY" {
		oldSize := s.positionSizes[fill.Product]
		newSize := oldSize + fill.FillQty

		// Update average entry price
		if oldSize > 0 {
			oldAvg := s.entryPrices[fill.Product]
			s.entryPrices[fill.Product] = (oldAvg*float64(oldSize) + fill.FillPrice*float64(fill.FillQty)) / float64(
				newSize,
			)
		} else {
			s.entryPrices[fill.Product] = fill.FillPrice
		}

		s.positionSizes[fill.Product] = newSize
	} else {
		s.positionSizes[fill.Product] -= fill.FillQty
		if s.positionSizes[fill.Product] <= 0 {
			s.positionSizes[fill.Product] = 0
			s.entryPrices[fill.Product] = 0
		}
	}

	log.Info().
		Str("strategy", s.name).
		Str("product", fill.Product).
		Str("side", fill.Side).
		Int("size", fill.FillQty).
		Float64("price", fill.FillPrice).
		Int("position", s.positionSizes[fill.Product]).
		Msg("📊 Fill received")

	return nil
}

// OnOffer is called when an offer request arrives
func (s *MomentumStrategy) OnOffer(ctx context.Context, offer *domain.OfferMessage) (*OfferResponse, error) {
	// Momentum trader doesn't respond to offers (uses market orders)
	return &OfferResponse{
		Accept: false,
		Reason: "Momentum trader uses market orders",
	}, nil
}

// OnInventoryUpdate is called when inventory changes
func (s *MomentumStrategy) OnInventoryUpdate(ctx context.Context, inventory map[string]int) error {
	return nil
}

// OnBalanceUpdate is called when balance changes
func (s *MomentumStrategy) OnBalanceUpdate(ctx context.Context, balance float64) error {
	return nil
}

// OnOrderBookUpdate is called when orderbook updates arrive
func (s *MomentumStrategy) OnOrderBookUpdate(ctx context.Context, orderbook *domain.OrderBookUpdateMessage) error {
	return nil
}

// Execute is called periodically to generate trading actions
func (s *MomentumStrategy) Execute(ctx context.Context, state *market.MarketState) ([]*Action, error) {
	// Check if enough time has passed since last execution
	if time.Since(s.lastExecute) < s.updateFrequency {
		return nil, nil
	}
	s.lastExecute = time.Now()

	actions := make([]*Action, 0)

	for _, product := range s.products {
		momentum := s.calculateMomentum(product)
		if momentum == 0 {
			continue // Not enough data
		}

		currentInventory := state.GetInventoryQuantity(product)
		currentPrice := s.getCurrentPrice(product)

		if currentPrice == 0 {
			continue // No price data
		}

		// Check stop loss and take profit conditions
		exitActions := s.checkExitConditions(product, currentInventory, currentPrice, state)
		actions = append(actions, exitActions...)

		// Determine if we should enter a position
		if momentum > s.momentumThreshold {
			// Positive momentum - buy signal
			buyActions := s.executeBuySignal(product, currentInventory, currentPrice, momentum, state)
			actions = append(actions, buyActions...)
		} else if momentum < -s.momentumThreshold {
			// Negative momentum - sell signal
			sellActions := s.executeSellSignal(product, currentInventory, currentPrice, momentum)
			actions = append(actions, sellActions...)
		}
	}

	return actions, nil
}

// calculateMomentum computes the rate of price change
func (s *MomentumStrategy) calculateMomentum(product string) float64 {
	history := s.priceHistory[product]
	if len(history) < 2 {
		return 0
	}

	// Clean old data points beyond lookback period
	cutoffTime := time.Now().Add(-s.lookbackPeriod)
	validHistory := make([]PricePoint, 0)
	for _, point := range history {
		if point.Timestamp.After(cutoffTime) {
			validHistory = append(validHistory, point)
		}
	}
	s.priceHistory[product] = validHistory

	if len(validHistory) < 2 {
		return 0
	}

	// Calculate momentum as percentage change
	oldestPrice := validHistory[0].Price
	newestPrice := validHistory[len(validHistory)-1].Price

	if oldestPrice == 0 {
		return 0
	}

	momentum := (newestPrice - oldestPrice) / oldestPrice

	log.Debug().
		Str("product", product).
		Float64("momentum", momentum*100).
		Float64("oldPrice", oldestPrice).
		Float64("newPrice", newestPrice).
		Int("dataPoints", len(validHistory)).
		Msg("Momentum calculated")

	return momentum
}

// checkExitConditions checks stop loss and take profit
func (s *MomentumStrategy) checkExitConditions(
	product string,
	currentInventory int,
	currentPrice float64,
	state *market.MarketState,
) []*Action {
	actions := make([]*Action, 0)

	if currentInventory <= 0 {
		return actions // No position to exit
	}

	entryPrice := s.entryPrices[product]
	if entryPrice == 0 {
		return actions
	}

	priceChange := (currentPrice - entryPrice) / entryPrice

	// Stop loss triggered
	if priceChange < -s.stopLoss {
		log.Warn().
			Str("product", product).
			Float64("loss", priceChange*100).
			Msg("Stop loss triggered - selling position")

		actions = append(actions, &Action{
			Type:  ActionTypeOrder,
			Order: CreateSellOrder(product, currentInventory, "Stop loss"),
		})
		return actions
	}

	// Take profit triggered
	if priceChange > s.takeProfit {
		log.Info().
			Str("product", product).
			Float64("profit", priceChange*100).
			Msg("Take profit triggered - selling position")

		actions = append(actions, &Action{
			Type:  ActionTypeOrder,
			Order: CreateSellOrder(product, currentInventory, "Take profit"),
		})
	}

	return actions
}

// executeBuySignal executes buy orders based on momentum
func (s *MomentumStrategy) executeBuySignal(
	product string,
	currentInventory int,
	currentPrice float64,
	momentum float64,
	state *market.MarketState,
) []*Action {
	actions := make([]*Action, 0)

	if currentInventory >= s.maxPosition {
		log.Debug().
			Str("product", product).
			Int("inventory", currentInventory).
			Msg("Max position reached, skipping buy")
		return actions
	}

	// Calculate position size (don't exceed max)
	positionSize := s.positionSize
	if currentInventory+positionSize > s.maxPosition {
		positionSize = s.maxPosition - currentInventory
	}

	if positionSize <= 0 {
		return actions
	}

	// Check if we have enough balance
	cost := currentPrice * float64(positionSize)
	if !state.HasSufficientBalance(cost) {
		log.Debug().
			Str("product", product).
			Float64("cost", cost).
			Float64("balance", state.Balance).
			Msg("Insufficient balance for buy order")
		return actions
	}

	log.Info().
		Str("product", product).
		Float64("momentum", momentum*100).
		Int("size", positionSize).
		Float64("price", currentPrice).
		Msg("🚀 Momentum buy signal")

	actions = append(actions, &Action{
		Type:  ActionTypeOrder,
		Order: CreateBuyOrder(product, positionSize, "Momentum buy"),
	})

	return actions
}

// executeSellSignal executes sell orders based on momentum
func (s *MomentumStrategy) executeSellSignal(
	product string,
	currentInventory int,
	currentPrice float64,
	momentum float64,
) []*Action {
	actions := make([]*Action, 0)

	if currentInventory <= 0 {
		log.Debug().
			Str("product", product).
			Msg("No inventory to sell")
		return actions
	}

	// Sell entire position or position size, whichever is smaller
	sellSize := s.positionSize
	if currentInventory < sellSize {
		sellSize = currentInventory
	}

	log.Info().
		Str("product", product).
		Float64("momentum", momentum*100).
		Int("size", sellSize).
		Float64("price", currentPrice).
		Msg("📉 Momentum sell signal")

	actions = append(actions, &Action{
		Type:  ActionTypeOrder,
		Order: CreateSellOrder(product, sellSize, "Momentum sell"),
	})

	return actions
}

// getCurrentPrice returns the latest price for a product
func (s *MomentumStrategy) getCurrentPrice(product string) float64 {
	history := s.priceHistory[product]
	if len(history) == 0 {
		return 0
	}
	return history[len(history)-1].Price
}

// Health returns the strategy's current health status
func (s *MomentumStrategy) Health() StrategyHealth {
	s.health.LastUpdate = time.Now()

	// Calculate total position value for PnL
	totalPnL := 0.0
	for product, size := range s.positionSizes {
		if size > 0 {
			entryPrice := s.entryPrices[product]
			currentPrice := s.getCurrentPrice(product)
			if currentPrice > 0 && entryPrice > 0 {
				totalPnL += (currentPrice - entryPrice) * float64(size)
			}
		}
	}

	s.health.PnL = totalPnL

	return s.health
}
