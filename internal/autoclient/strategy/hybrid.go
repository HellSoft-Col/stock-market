package strategy

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/autoclient/market"
	"github.com/HellSoft-Col/stock-market/internal/autoclient/production"
	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/rs/zerolog/log"
)

// HybridStrategy is an intelligent trader that combines ALL strategies:
// - Production: Produces profitable products using recipes
// - Market Making: Provides liquidity with bid/ask spreads
// - Arbitrage: Exploits price differences and production margins
// - Momentum: Follows trends and captures price movements
// - Inventory Management: Optimizes what to buy, sell, and hold
type HybridStrategy struct {
	name   string
	apiKey string

	// Configuration
	productionInterval time.Duration
	quoteInterval      time.Duration
	tradeInterval      time.Duration

	// Production config
	minProductionMargin float64 // Min profit margin to produce
	maxProductionRuns   int     // Max concurrent production runs

	// Market making config
	spread       float64 // Bid-ask spread
	quoteSize    int     // Size per quote
	maxInventory int     // Max inventory per product

	// Arbitrage config
	minArbitrageSpread float64 // Min spread for arbitrage

	// Momentum config
	momentumThreshold float64 // Price change threshold to trigger
	maxPosition       int     // Max position size per product

	// Risk management
	maxOrderSize int
	maxDailyLoss float64

	// Production system
	calculator    *production.ProductionCalculator
	recipeManager *production.RecipeManager
	role          *production.Role

	// State tracking
	lastProduction    time.Time
	lastQuote         time.Time
	lastTrade         time.Time
	productionQueue   map[string]int // Product -> quantity queued
	priceHistory      map[string][]PricePoint
	eventHistory      *EventHistory
	teamName          string
	dailyPnL          float64
	startOfDayBalance float64

	// Strategy weights (dynamically adjusted)
	weights StrategyWeights

	// Health
	health StrategyHealth
}

// StrategyWeights controls priority of each sub-strategy
type StrategyWeights struct {
	Production   float64 // 0-1
	MarketMaking float64
	Arbitrage    float64
	Momentum     float64
}

// NewHybridStrategy creates a new hybrid intelligent trader
func NewHybridStrategy(name string) *HybridStrategy {
	return &HybridStrategy{
		name:            name,
		calculator:      production.NewProductionCalculator(),
		productionQueue: make(map[string]int),
		priceHistory:    make(map[string][]PricePoint),
		eventHistory:    NewEventHistory(100), // Track last 100 events
		weights: StrategyWeights{
			Production:   0.30, // 30% focus on production
			MarketMaking: 0.25, // 25% focus on market making
			Arbitrage:    0.25, // 25% focus on arbitrage
			Momentum:     0.20, // 20% focus on momentum
		},
		health: StrategyHealth{
			Status:     HealthStatusHealthy,
			LastUpdate: time.Now(),
		},
	}
}

// Name returns the strategy name
func (s *HybridStrategy) Name() string {
	return s.name
}

// Initialize initializes the strategy with configuration
func (s *HybridStrategy) Initialize(config map[string]interface{}) error {
	// Read API key for funny messages
	s.apiKey = GetConfigString(config, "apiKey", "")

	// Production settings
	s.productionInterval = GetConfigDuration(config, "productionInterval", 30*time.Second)
	s.minProductionMargin = GetConfigFloat(config, "minProductionMargin", 0.10) // 10%
	s.maxProductionRuns = GetConfigInt(config, "maxProductionRuns", 5)

	// Market making settings
	s.quoteInterval = GetConfigDuration(config, "quoteInterval", 5*time.Second)
	s.spread = GetConfigFloat(config, "spread", 0.02) // 2%
	s.quoteSize = GetConfigInt(config, "quoteSize", 50)
	s.maxInventory = GetConfigInt(config, "maxInventory", 500)

	// Arbitrage settings
	s.minArbitrageSpread = GetConfigFloat(config, "minArbitrageSpread", 0.05) // 5%
	s.tradeInterval = GetConfigDuration(config, "tradeInterval", 3*time.Second)

	// Momentum settings
	s.momentumThreshold = GetConfigFloat(config, "momentumThreshold", 0.03) // 3%
	s.maxPosition = GetConfigInt(config, "maxPosition", 300)

	// Risk management
	s.maxOrderSize = GetConfigInt(config, "maxOrderSize", 200)
	s.maxDailyLoss = GetConfigFloat(config, "maxDailyLoss", 1000.0)

	// Initialize message generator with API key
	InitMessageGenerator(s.name, s.apiKey)

	log.Info().
		Str("strategy", s.name).
		Dur("productionInterval", s.productionInterval).
		Dur("quoteInterval", s.quoteInterval).
		Float64("spread", s.spread).
		Msg("Hybrid strategy initialized")

	return nil
}

// OnLogin is called when connected and logged in
func (s *HybridStrategy) OnLogin(ctx context.Context, loginInfo *domain.LoginOKMessage) error {
	s.startOfDayBalance = loginInfo.CurrentBalance
	s.teamName = loginInfo.Team

	// Initialize message generator with team name (no AI for hybrid, use templates)
	InitMessageGenerator(s.teamName, "")

	// Initialize production system
	s.role = &production.Role{
		Branches:    loginInfo.Role.Branches,
		MaxDepth:    loginInfo.Role.MaxDepth,
		Decay:       loginInfo.Role.Decay,
		BaseEnergy:  loginInfo.Role.BaseEnergy,
		LevelEnergy: loginInfo.Role.LevelEnergy,
	}

	// Convert recipes
	recipes := make(map[string]*production.Recipe)
	for product, recipe := range loginInfo.Recipes {
		recipes[product] = &production.Recipe{
			Product:      product,
			Ingredients:  recipe.Ingredients,
			PremiumBonus: recipe.PremiumBonus,
		}
	}
	s.recipeManager = production.NewRecipeManager(recipes)

	log.Info().
		Str("strategy", s.name).
		Str("team", loginInfo.Team).
		Int("recipes", len(recipes)).
		Float64("balance", loginInfo.CurrentBalance).
		Msg("Hybrid strategy logged in")

	return nil
}

// OnTicker is called when ticker updates arrive
func (s *HybridStrategy) OnTicker(ctx context.Context, ticker *domain.TickerMessage) error {
	// Track price history for momentum detection
	if ticker.Mid != nil {
		s.addPricePoint(ticker.Product, *ticker.Mid)
	}
	return nil
}

// OnFill is called when a fill notification arrives
func (s *HybridStrategy) OnFill(ctx context.Context, fill *domain.FillMessage) error {
	// Track fill event
	pnl := 0.0
	if fill.Side == "BUY" {
		pnl = -fill.FillPrice * float64(fill.FillQty)
	} else {
		pnl = fill.FillPrice * float64(fill.FillQty)
	}

	s.eventHistory.AddEvent(TradingEvent{
		Timestamp: time.Now(),
		Type:      "FILL",
		Action:    fill.Side,
		Product:   fill.Product,
		Quantity:  fill.FillQty,
		Price:     fill.FillPrice,
		PnL:       pnl,
		Message:   fill.CounterpartyMessage,
	})

	log.Info().
		Str("strategy", s.name).
		Str("product", fill.Product).
		Str("side", fill.Side).
		Int("qty", fill.FillQty).
		Float64("price", fill.FillPrice).
		Str("counterparty", fill.Counterparty).
		Msg("✅ Fill received")
	return nil
}

// OnOffer is called when an offer request arrives
func (s *HybridStrategy) OnOffer(ctx context.Context, offer *domain.OfferMessage) (*OfferResponse, error) {
	// Smart offer response: accept if profitable
	// This is part of the liquidity provision aspect

	// Quick profitability check
	productionCost := s.estimateProductionCost(offer.Product)
	if productionCost > 0 && offer.MaxPrice > productionCost*1.2 {
		// Accept if price is 20% above our cost
		return &OfferResponse{
			Accept: true,
			Reason: "Profitable offer accepted",
		}, nil
	}

	return &OfferResponse{
		Accept: false,
		Reason: "Not profitable",
	}, nil
}

// OnInventoryUpdate is called when inventory changes
func (s *HybridStrategy) OnInventoryUpdate(ctx context.Context, inventory map[string]int) error {
	return nil
}

// OnBalanceUpdate is called when balance changes
func (s *HybridStrategy) OnBalanceUpdate(ctx context.Context, balance float64) error {
	// Update daily P&L
	s.dailyPnL = balance - s.startOfDayBalance

	// Risk check: stop trading if losses exceed limit
	if s.dailyPnL < -s.maxDailyLoss {
		s.health.Status = HealthStatusDegraded
		s.health.Message = fmt.Sprintf("Daily loss limit reached: $%.2f", s.dailyPnL)
		log.Warn().
			Str("strategy", s.name).
			Float64("dailyPnL", s.dailyPnL).
			Msg("Daily loss limit reached, going defensive")

		// Adjust weights to be more conservative
		s.weights.Production = 0.50
		s.weights.MarketMaking = 0.30
		s.weights.Arbitrage = 0.20
		s.weights.Momentum = 0.0
	}

	return nil
}

// OnOrderBookUpdate is called when orderbook updates arrive
func (s *HybridStrategy) OnOrderBookUpdate(ctx context.Context, orderbook *domain.OrderBookUpdateMessage) error {
	return nil
}

// Execute is called periodically to generate trading actions
func (s *HybridStrategy) Execute(ctx context.Context, state *market.MarketState) ([]*Action, error) {
	actions := []*Action{}

	// Get market snapshot
	snapshot := state.GetSnapshot()

	// 1. PRODUCTION LOGIC - Produce profitable products
	if time.Since(s.lastProduction) >= s.productionInterval {
		productionActions := s.executeProduction(snapshot)
		actions = append(actions, productionActions...)
		s.lastProduction = time.Now()
	}

	// 2. ARBITRAGE LOGIC - Find and exploit price differences
	arbitrageActions := s.executeArbitrage(snapshot)
	actions = append(actions, arbitrageActions...)

	// 3. MARKET MAKING LOGIC - Provide liquidity
	if time.Since(s.lastQuote) >= s.quoteInterval {
		marketMakingActions := s.executeMarketMaking(snapshot)
		actions = append(actions, marketMakingActions...)
		s.lastQuote = time.Now()
	}

	// 4. MOMENTUM LOGIC - Follow trends
	if time.Since(s.lastTrade) >= s.tradeInterval {
		momentumActions := s.executeMomentum(snapshot)
		actions = append(actions, momentumActions...)
		s.lastTrade = time.Now()
	}

	// 5. INVENTORY MANAGEMENT - Sell excess, buy needs
	inventoryActions := s.executeInventoryManagement(snapshot)
	actions = append(actions, inventoryActions...)

	// Update health
	s.health.LastUpdate = time.Now()
	s.health.PnL = s.dailyPnL

	if len(actions) > 0 {
		// Count action types
		orderCount := 0
		productionCount := 0
		for _, action := range actions {
			switch action.Type {
			case ActionTypeOrder:
				orderCount++
			case ActionTypeProduction:
				productionCount++
			}
		}

		log.Info().
			Str("strategy", s.name).
			Int("total", len(actions)).
			Int("orders", orderCount).
			Int("production", productionCount).
			Msg("🎯 Hybrid strategy generating actions")
	}

	return actions, nil
}

// executeProduction handles production decisions
func (s *HybridStrategy) executeProduction(state *market.MarketState) []*Action {
	actions := []*Action{}

	// Find most profitable production opportunity
	type ProductionOpp struct {
		Product        string
		Margin         float64
		Units          int
		MarketPrice    float64
		ProductionCost float64
	}

	opportunities := []ProductionOpp{}

	for product, recipe := range s.recipeManager.GetAllRecipes() {
		// Can we produce this?
		if !s.recipeManager.CanProducePremium(product, state.Inventory) {
			// Check if it's a basic product (no ingredients)
			if len(recipe.Ingredients) > 0 {
				continue
			}
		}

		// Calculate production cost
		productionCost := s.calculateRecipeProductionCost(state, recipe)

		// Get market price (or use default)
		price := state.GetPrice(product)
		var marketPrice float64
		if price == nil {
			marketPrice = getDefaultPrice(product)
		} else {
			marketPrice = *price
		}

		// Calculate profit margin
		var margin float64
		if productionCost == 0 {
			// Free production (basic products) - always produce!
			margin = 999.0 // Very high margin to prioritize
		} else {
			margin = (marketPrice - productionCost) / productionCost
		}

		// Produce if profitable OR if it's free (basic product)
		if margin > s.minProductionMargin || productionCost == 0 {
			// Calculate units
			baseUnits := s.calculator.CalculateUnits(s.role)
			units := baseUnits
			if len(recipe.Ingredients) > 0 {
				units = s.calculator.ApplyPremiumBonus(baseUnits, recipe.PremiumBonus)
			}

			opportunities = append(opportunities, ProductionOpp{
				Product:        product,
				Margin:         margin,
				Units:          units,
				MarketPrice:    marketPrice,
				ProductionCost: productionCost,
			})
		}
	}

	// Sort by margin (highest first)
	if len(opportunities) == 0 {
		return actions
	}

	// Find best opportunity
	bestOpp := opportunities[0]
	for _, opp := range opportunities {
		if opp.Margin > bestOpp.Margin {
			bestOpp = opp
		}
	}

	// Produce the most profitable product
	log.Info().
		Str("strategy", s.name).
		Str("product", bestOpp.Product).
		Float64("margin", bestOpp.Margin*100).
		Int("units", bestOpp.Units).
		Float64("marketPrice", bestOpp.MarketPrice).
		Float64("cost", bestOpp.ProductionCost).
		Msg("🏭 Producing profitable product")

	actions = append(actions, &Action{
		Type: ActionTypeProduction,
		Production: &domain.ProductionUpdateMessage{
			Type:     "PRODUCTION_UPDATE",
			Product:  bestOpp.Product,
			Quantity: bestOpp.Units,
		},
	})

	// After producing, sell aggressively for liquidity
	if bestOpp.ProductionCost == 0 {
		// Free production - sell at MARKET for instant liquidity
		actions = append(actions, &Action{
			Type:  ActionTypeOrder,
			Order: CreateSellOrder(bestOpp.Product, bestOpp.Units, "MARKET sell basic production"),
		})
	} else {
		// Premium production - sell at LIMIT for better price
		sellPrice := bestOpp.MarketPrice * 1.05 // 5% above market
		actions = append(actions, &Action{
			Type:  ActionTypeOrder,
			Order: CreateLimitSellOrder(bestOpp.Product, bestOpp.Units, sellPrice, "LIMIT sell premium production"),
		})
	}

	return actions
}

// executeArbitrage finds arbitrage opportunities
func (s *HybridStrategy) executeArbitrage(state *market.MarketState) []*Action {
	actions := []*Action{}

	// Look for wide spreads
	for product, ticker := range state.Tickers {
		if ticker.BestBid == nil || ticker.BestAsk == nil {
			continue
		}

		bid := *ticker.BestBid
		ask := *ticker.BestAsk

		if bid == 0 || ask == 0 {
			continue
		}

		spread := (ask - bid) / bid

		if spread > s.minArbitrageSpread {
			log.Info().
				Str("product", product).
				Float64("spread", spread*100).
				Msg("💎 Arbitrage opportunity detected")

			inventory := state.GetInventoryQuantity(product)

			// Buy at bid
			if state.Balance > bid*float64(s.quoteSize) {
				actions = append(actions, &Action{
					Type:  ActionTypeOrder,
					Order: CreateLimitBuyOrder(product, s.quoteSize, bid, "Arb buy"),
				})
			}

			// Sell at ask (if have inventory)
			if inventory >= s.quoteSize {
				actions = append(actions, &Action{
					Type:  ActionTypeOrder,
					Order: CreateLimitSellOrder(product, s.quoteSize, ask, "Arb sell"),
				})
			}
		}
	}

	// Recipe-based arbitrage: produce if cheaper than market
	for product, recipe := range s.recipeManager.GetAllRecipes() {
		productionCost := s.calculateRecipeProductionCost(state, recipe)
		if productionCost == 0 {
			continue
		}

		price := state.GetPrice(product)
		if price == nil {
			continue
		}
		marketPrice := *price

		margin := (marketPrice - productionCost) / productionCost

		if margin > s.minArbitrageSpread {
			// Check if we can produce
			if s.recipeManager.CanProducePremium(product, state.Inventory) || len(recipe.Ingredients) == 0 {
				log.Info().
					Str("product", product).
					Float64("margin", margin*100).
					Float64("marketPrice", marketPrice).
					Float64("cost", productionCost).
					Msg("🎯 Recipe arbitrage detected")

				// Buy ingredients if needed
				for ingredient, qty := range recipe.Ingredients {
					if state.GetInventoryQuantity(ingredient) < qty {
						ingredientPrice := state.GetPrice(ingredient)
						if ingredientPrice != nil {
							actions = append(actions, &Action{
								Type:  ActionTypeOrder,
								Order: CreateBuyOrder(ingredient, qty, "Buy ingredient"),
							})
						}
					}
				}
			}
		}
	}

	return actions
}

// executeMarketMaking provides liquidity
func (s *HybridStrategy) executeMarketMaking(state *market.MarketState) []*Action {
	actions := []*Action{}

	// Get all products we know about (from recipes or tickers)
	products := make(map[string]bool)
	for product := range state.Tickers {
		products[product] = true
	}
	for product := range s.recipeManager.GetAllRecipes() {
		products[product] = true
	}

	for product := range products {
		// Get or estimate mid price
		var midPrice float64
		ticker := state.Tickers[product]
		if ticker != nil && ticker.Mid != nil {
			midPrice = *ticker.Mid
		} else {
			midPrice = getDefaultPrice(product)
		}

		inventory := state.GetInventoryQuantity(product)

		// Calculate bid/ask
		bidPrice := midPrice * (1 - s.spread/2)
		askPrice := midPrice * (1 + s.spread/2)

		// ALWAYS place buy quotes to provide liquidity (if we have balance)
		if state.Balance > bidPrice*float64(s.quoteSize) {
			// Vary buy size: 50-100% of quoteSize
			buyQty := s.quoteSize/2 + int(time.Now().UnixNano()%int64(s.quoteSize/2))

			// Primary buy order
			actions = append(actions, &Action{
				Type:  ActionTypeOrder,
				Order: CreateLimitBuyOrder(product, buyQty, bidPrice, "MM buy liquidity"),
			})

			// Add SECOND deeper buy order for extra liquidity (if we have funds)
			if state.Balance > bidPrice*float64(s.quoteSize)*2 {
				deeperBidPrice := bidPrice * 0.98 // 2% below regular bid
				deeperBuyQty := buyQty / 2
				if deeperBuyQty > 10 {
					actions = append(actions, &Action{
						Type:  ActionTypeOrder,
						Order: CreateLimitBuyOrder(product, deeperBuyQty, deeperBidPrice, "MM deeper buy"),
					})
				}
			}
		}

		// Place sell quote if have ANY inventory (not just > quoteSize)
		if inventory > 10 {
			sellQty := s.quoteSize
			if sellQty > inventory {
				sellQty = inventory / 2 // Sell half
				if sellQty < 10 {
					sellQty = 10
				}
			}
			actions = append(actions, &Action{
				Type:  ActionTypeOrder,
				Order: CreateLimitSellOrder(product, sellQty, askPrice, "MM sell"),
			})
		}
	}

	return actions
}

// executeMomentum follows price trends
func (s *HybridStrategy) executeMomentum(state *market.MarketState) []*Action {
	actions := []*Action{}

	for product := range state.Tickers {
		momentum := s.calculateMomentum(product)
		if math.Abs(momentum) < s.momentumThreshold {
			continue
		}

		price := state.GetPrice(product)
		if price == nil {
			continue
		}
		currentPrice := *price

		inventory := state.GetInventoryQuantity(product)

		// Positive momentum: buy
		if momentum > 0 && inventory < s.maxPosition {
			log.Info().
				Str("product", product).
				Float64("momentum", momentum*100).
				Msg("📈 Positive momentum, buying")

			buyQty := s.quoteSize
			if state.Balance > currentPrice*float64(buyQty) {
				actions = append(actions, &Action{
					Type:  ActionTypeOrder,
					Order: CreateBuyOrder(product, buyQty, "Momentum buy"),
				})
			}
		}

		// Negative momentum: sell
		if momentum < 0 && inventory > 0 {
			log.Info().
				Str("product", product).
				Float64("momentum", momentum*100).
				Msg("📉 Negative momentum, selling")

			sellQty := s.quoteSize
			if sellQty > inventory {
				sellQty = inventory
			}
			actions = append(actions, &Action{
				Type:  ActionTypeOrder,
				Order: CreateSellOrder(product, sellQty, "Momentum sell"),
			})
		}
	}

	return actions
}

// executeInventoryManagement optimizes inventory
func (s *HybridStrategy) executeInventoryManagement(state *market.MarketState) []*Action {
	actions := []*Action{}

	// Actively manage inventory - sell excess but not too aggressively
	for product, qty := range state.Inventory {
		// Sell at 1/3 capacity (less aggressive selling, more balanced)
		threshold := s.maxInventory / 3
		if threshold < 30 {
			threshold = 30 // Lower minimum threshold
		}

		if qty > threshold {
			price := state.GetPrice(product)
			var useMarket bool
			var sellPrice float64

			if price == nil {
				// No market price, use default and LIMIT order
				sellPrice = getDefaultPrice(product) * 1.03
				useMarket = false
			} else {
				// Use mix: 50% MARKET, 50% LIMIT
				useMarket = (time.Now().UnixNano() % 2) == 0
				sellPrice = *price * 1.02 // 2% above market
			}

			excess := qty - threshold
			sellQty := excess
			if sellQty > s.maxOrderSize {
				sellQty = s.maxOrderSize
			}
			// Sell at least 30 units to keep things moving
			if sellQty < 30 && qty >= 30 {
				sellQty = 30
			}

			log.Info().
				Str("product", product).
				Int("inventory", qty).
				Int("sellQty", sellQty).
				Bool("market", useMarket).
				Msg("🧹 Actively managing inventory")

			if useMarket {
				actions = append(actions, &Action{
					Type:  ActionTypeOrder,
					Order: CreateSellOrder(product, sellQty, "Clear inventory MARKET"),
				})
			} else {
				actions = append(actions, &Action{
					Type:  ActionTypeOrder,
					Order: CreateLimitSellOrder(product, sellQty, sellPrice, "Clear inventory LIMIT"),
				})
			}
		}
	}

	// Buy ingredients we're low on
	for product, recipe := range s.recipeManager.GetAllRecipes() {
		// Skip if not profitable to produce
		productionCost := s.calculateRecipeProductionCost(state, recipe)
		if productionCost == 0 {
			continue
		}

		price := state.GetPrice(product)
		if price == nil {
			continue
		}

		margin := (*price - productionCost) / productionCost
		if margin < s.minProductionMargin {
			continue
		}

		// Check if we need ingredients
		for ingredient, needed := range recipe.Ingredients {
			have := state.GetInventoryQuantity(ingredient)
			if have < needed {
				ingredientPrice := state.GetPrice(ingredient)
				if ingredientPrice != nil && state.Balance > *ingredientPrice*float64(needed) {
					log.Info().
						Str("ingredient", ingredient).
						Int("needed", needed).
						Int("have", have).
						Msg("🛒 Buying needed ingredient")

					actions = append(actions, &Action{
						Type:  ActionTypeOrder,
						Order: CreateBuyOrder(ingredient, needed-have, "Buy ingredient"),
					})
				}
			}
		}
	}

	return actions
}

// Helper methods

func (s *HybridStrategy) addPricePoint(product string, price float64) {
	if s.priceHistory[product] == nil {
		s.priceHistory[product] = []PricePoint{}
	}

	s.priceHistory[product] = append(s.priceHistory[product], PricePoint{
		Price:     price,
		Timestamp: time.Now(),
	})

	// Keep only last 100 points
	if len(s.priceHistory[product]) > 100 {
		s.priceHistory[product] = s.priceHistory[product][len(s.priceHistory[product])-100:]
	}
}

func (s *HybridStrategy) calculateMomentum(product string) float64 {
	history := s.priceHistory[product]
	if len(history) < 10 {
		return 0
	}

	// Compare recent average to older average
	recentAvg := 0.0
	oldAvg := 0.0

	recentCount := len(history) / 3
	oldCount := len(history) / 3

	for i := len(history) - recentCount; i < len(history); i++ {
		recentAvg += history[i].Price
	}
	recentAvg /= float64(recentCount)

	for i := 0; i < oldCount; i++ {
		oldAvg += history[i].Price
	}
	oldAvg /= float64(oldCount)

	if oldAvg == 0 {
		return 0
	}

	return (recentAvg - oldAvg) / oldAvg
}

func (s *HybridStrategy) calculateRecipeProductionCost(state *market.MarketState, recipe *production.Recipe) float64 {
	if len(recipe.Ingredients) == 0 {
		return 0 // Free production
	}

	cost := 0.0
	for ingredient, qty := range recipe.Ingredients {
		price := state.GetPrice(ingredient)
		if price == nil {
			return 0 // Can't calculate
		}
		cost += *price * float64(qty)
	}

	return cost
}

func (s *HybridStrategy) estimateProductionCost(product string) float64 {
	recipe, err := s.recipeManager.GetRecipe(product)
	if err != nil {
		return 0
	}

	// Rough estimate without market state
	if len(recipe.Ingredients) == 0 {
		return 0
	}

	// Return a high estimate if we don't have market data
	return 999999.0
}

// Health returns the strategy's current health status
func (s *HybridStrategy) Health() StrategyHealth {
	return s.health
}
