package strategy

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/autoclient/market"
	"github.com/HellSoft-Col/stock-market/internal/autoclient/production"
	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/rs/zerolog/log"
)

// BuffettStrategy implements Warren Buffett-style value investing using AI
// Philosophy:
// - "Be fearful when others are greedy, greedy when others are fearful"
// - Long-term value over short-term gains
// - Quality over quantity
// - Patience and discipline
// - Focus on fundamentals (production costs vs market price)
// - Buy undervalued assets and hold
type BuffettStrategy struct {
	name string

	// Configuration
	apiKey            string
	endpoint          string
	decisionInterval  time.Duration
	temperature       float64
	maxTokens         int
	minConfidence     float64
	timeout           time.Duration
	maxRetries        int
	includeProduction bool

	// Buffett-specific parameters
	minValueMargin   float64 // Minimum margin of safety (default 30%)
	maxPriceEarnings float64 // Max P/E ratio to consider
	longTermHorizon  time.Duration
	qualityThreshold float64 // Quality score threshold

	// Production system
	calculator    *production.ProductionCalculator
	recipeManager *production.RecipeManager
	role          *production.Role

	// State
	lastDecision  time.Time
	balance       float64
	inventory     map[string]int
	marketState   *market.MarketState
	eventHistory  *EventHistory
	teamName      string
	lastError     string
	errorAttempts int

	// Buffett-specific tracking
	intrinsicValues map[string]float64  // Product -> calculated intrinsic value
	positions       map[string]Position // Long-term positions

	health     StrategyHealth
	errorCount int
	httpClient *http.Client
}

// Position represents a long-term holding
type Position struct {
	Product        string
	Quantity       int
	AvgCostBasis   float64
	AcquiredAt     time.Time
	IntrinsicValue float64
	Reasoning      string
}

// NewBuffettStrategy creates a new Warren Buffett-style AI strategy
func NewBuffettStrategy(name string) *BuffettStrategy {
	return &BuffettStrategy{
		name:            name,
		inventory:       make(map[string]int),
		intrinsicValues: make(map[string]float64),
		positions:       make(map[string]Position),
		eventHistory:    NewEventHistory(100),
		health: StrategyHealth{
			Status:     HealthStatusHealthy,
			LastUpdate: time.Now(),
		},
	}
}

// Name returns the strategy name
func (s *BuffettStrategy) Name() string {
	return s.name
}

// Initialize initializes the strategy with configuration
func (s *BuffettStrategy) Initialize(config map[string]interface{}) error {
	s.apiKey = GetConfigString(config, "apiKey", "")
	s.endpoint = GetConfigString(config, "endpoint", "https://api.deepseek.com/v1/chat/completions")
	s.decisionInterval = GetConfigDuration(config, "decisionInterval", 30*time.Second)
	s.temperature = GetConfigFloat(config, "temperature", 0.5) // Lower for conservative decisions
	s.maxTokens = GetConfigInt(config, "maxTokens", 800)
	s.minConfidence = GetConfigFloat(config, "minConfidence", 0.7) // Higher threshold
	s.timeout = GetConfigDuration(config, "timeout", 10*time.Second)
	s.maxRetries = GetConfigInt(config, "maxRetries", 3)
	s.includeProduction = GetConfigBool(config, "includeProduction", true)

	// Buffett-specific parameters
	s.minValueMargin = GetConfigFloat(config, "minValueMargin", 0.30)     // 30% margin of safety
	s.maxPriceEarnings = GetConfigFloat(config, "maxPriceEarnings", 15.0) // Conservative P/E
	s.longTermHorizon = GetConfigDuration(config, "longTermHorizon", 24*time.Hour)
	s.qualityThreshold = GetConfigFloat(config, "qualityThreshold", 0.8)

	if s.apiKey == "" {
		return fmt.Errorf("apiKey is required for Buffett strategy")
	}

	// Initialize HTTP client
	s.httpClient = &http.Client{
		Timeout: s.timeout,
	}

	// Initialize production calculator if enabled
	if s.includeProduction {
		s.calculator = production.NewProductionCalculator()
	}

	log.Info().
		Str("strategy", s.name).
		Dur("decisionInterval", s.decisionInterval).
		Float64("minValueMargin", s.minValueMargin).
		Msg("Warren Buffett strategy initialized")

	return nil
}

// OnLogin is called when connected and logged in
func (s *BuffettStrategy) OnLogin(ctx context.Context, loginInfo *domain.LoginOKMessage) error {
	s.balance = loginInfo.CurrentBalance
	s.teamName = loginInfo.Team

	// Initialize message generator
	InitMessageGenerator(s.teamName, s.apiKey)

	if s.includeProduction {
		s.role = &production.Role{
			Branches:    loginInfo.Role.Branches,
			MaxDepth:    loginInfo.Role.MaxDepth,
			Decay:       loginInfo.Role.Decay,
			BaseEnergy:  loginInfo.Role.BaseEnergy,
			LevelEnergy: loginInfo.Role.LevelEnergy,
		}

		recipes := make(map[string]*production.Recipe)
		for product, recipe := range loginInfo.Recipes {
			recipes[product] = &production.Recipe{
				Product:      product,
				Ingredients:  recipe.Ingredients,
				PremiumBonus: recipe.PremiumBonus,
			}
		}
		s.recipeManager = production.NewRecipeManager(recipes)

		// Calculate intrinsic values based on production costs
		s.calculateIntrinsicValues()

		log.Info().
			Str("strategy", s.name).
			Int("recipes", len(recipes)).
			Msg("Production-based valuation initialized")
	}

	return nil
}

// OnTicker is called when ticker updates arrive
// NOTE: Agent already updates market state, this is for strategy-specific tracking
func (s *BuffettStrategy) OnTicker(ctx context.Context, ticker *domain.TickerMessage) error {
	// Recalculate intrinsic values periodically
	if time.Since(s.lastDecision) > 5*time.Minute {
		s.calculateIntrinsicValues()
	}
	return nil
}

// OnFill is called when a fill notification arrives
func (s *BuffettStrategy) OnFill(ctx context.Context, fill *domain.FillMessage) error {
	// Track fill event
	pnl := 0.0
	if fill.Side == "BUY" {
		pnl = -fill.FillPrice * float64(fill.FillQty)

		// Update position tracking for buys
		pos, exists := s.positions[fill.Product]
		if exists {
			// Update average cost basis
			totalCost := (pos.AvgCostBasis * float64(pos.Quantity)) + (fill.FillPrice * float64(fill.FillQty))
			pos.Quantity += fill.FillQty
			pos.AvgCostBasis = totalCost / float64(pos.Quantity)
			s.positions[fill.Product] = pos
		} else {
			// New position
			s.positions[fill.Product] = Position{
				Product:        fill.Product,
				Quantity:       fill.FillQty,
				AvgCostBasis:   fill.FillPrice,
				AcquiredAt:     time.Now(),
				IntrinsicValue: s.intrinsicValues[fill.Product],
				Reasoning:      "Long-term value investment",
			}
		}
	} else {
		pnl = fill.FillPrice * float64(fill.FillQty)

		// Update position for sells
		if pos, exists := s.positions[fill.Product]; exists {
			pos.Quantity -= fill.FillQty
			if pos.Quantity <= 0 {
				delete(s.positions, fill.Product)
			} else {
				s.positions[fill.Product] = pos
			}
		}
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
		Int("quantity", fill.FillQty).
		Float64("price", fill.FillPrice).
		Msg("Fill received")
	return nil
}

// OnOffer is called when an offer request arrives
func (s *BuffettStrategy) OnOffer(ctx context.Context, offer *domain.OfferMessage) (*OfferResponse, error) {
	// Buffett evaluates offers based on intrinsic value
	intrinsicValue, exists := s.intrinsicValues[offer.Product]
	if !exists {
		return &OfferResponse{
			Accept: false,
			Reason: "No valuation available",
		}, nil
	}

	// Accept if price is below intrinsic value with margin of safety
	marginOfSafety := (intrinsicValue - offer.MaxPrice) / intrinsicValue
	if marginOfSafety >= s.minValueMargin {
		return &OfferResponse{
			Accept: true,
			Reason: fmt.Sprintf("Excellent value: %.0f%% below intrinsic value", marginOfSafety*100),
		}, nil
	}

	return &OfferResponse{
		Accept: false,
		Reason: "Price above intrinsic value",
	}, nil
}

// OnInventoryUpdate is called when inventory changes
func (s *BuffettStrategy) OnInventoryUpdate(ctx context.Context, inventory map[string]int) error {
	s.inventory = inventory
	return nil
}

// OnBalanceUpdate is called when balance changes
func (s *BuffettStrategy) OnBalanceUpdate(ctx context.Context, balance float64) error {
	s.balance = balance
	return nil
}

// OnOrderBookUpdate is called when orderbook updates arrive
func (s *BuffettStrategy) OnOrderBookUpdate(ctx context.Context, orderbook *domain.OrderBookUpdateMessage) error {
	return nil
}

// Execute is called periodically to generate trading actions
func (s *BuffettStrategy) Execute(ctx context.Context, state *market.MarketState) ([]*Action, error) {
	s.marketState = state

	// Check if it's time to make a decision (longer interval for patient investing)
	if time.Since(s.lastDecision) < s.decisionInterval {
		return nil, nil
	}

	s.lastDecision = time.Now()

	// Build Warren Buffett-style market context
	marketContext := s.buildBuffettContext()

	// Get AI decisions with Buffett philosophy
	decisions, err := s.getAIDecisions(ctx, marketContext)
	if err != nil {
		s.errorCount++
		log.Error().
			Err(err).
			Str("strategy", s.name).
			Msg("Failed to get AI decisions")
		return nil, err
	}

	// Execute decisions with validation
	allActions := []*Action{}
	successfulDecisions := []AIDecision{}
	errorMessages := []string{}

	for _, decision := range decisions {
		actions, err := s.executeDecisionWithValidation(decision)
		if err != nil {
			errorMsg := fmt.Sprintf("Action %s on %s failed: %v", decision.Action, decision.Product, err)
			errorMessages = append(errorMessages, errorMsg)
		} else {
			allActions = append(allActions, actions...)
			successfulDecisions = append(successfulDecisions, decision)
		}
	}

	// Retry with error feedback if needed
	if len(errorMessages) > 0 && len(allActions) == 0 && s.errorAttempts < 2 {
		s.lastError = strings.Join(errorMessages, "; ")
		s.errorAttempts++
		return s.Execute(ctx, state)
	}

	// Track successful decisions
	if len(allActions) > 0 {
		s.errorAttempts = 0
		s.lastError = ""

		for _, decision := range successfulDecisions {
			s.eventHistory.AddEvent(TradingEvent{
				Timestamp: time.Now(),
				Type:      "ORDER",
				Action:    decision.Action,
				Product:   decision.Product,
				Quantity:  decision.Quantity,
				Price:     decision.Price,
				Message:   decision.Reasoning,
			})
		}
	}

	return allActions, nil
}

// Health returns the strategy's current health status
func (s *BuffettStrategy) Health() StrategyHealth {
	return s.health
}

// calculateIntrinsicValues calculates intrinsic value for each product
func (s *BuffettStrategy) calculateIntrinsicValues() {
	if s.recipeManager == nil || s.marketState == nil {
		return
	}

	snapshot := s.marketState.GetSnapshot()

	for product, recipe := range s.recipeManager.GetAllRecipes() {
		// Calculate production cost
		productionCost := 0.0
		if len(recipe.Ingredients) == 0 {
			// Free production - intrinsic value is market-based
			if price := snapshot.GetPrice(product); price != nil {
				productionCost = *price * 0.5 // Conservative estimate
			}
		} else {
			// Calculate ingredient costs
			for ingredient, qty := range recipe.Ingredients {
				if price := snapshot.GetPrice(ingredient); price != nil {
					productionCost += *price * float64(qty)
				}
			}
		}

		// Calculate output value with premium bonus
		baseUnits := s.calculator.CalculateUnits(s.role)
		outputUnits := float64(baseUnits)
		if len(recipe.Ingredients) > 0 {
			outputUnits = float64(s.calculator.ApplyPremiumBonus(baseUnits, recipe.PremiumBonus))
		}

		// Intrinsic value per unit = production cost / output units
		intrinsicValue := productionCost / outputUnits

		// Add margin for quality and certainty
		intrinsicValue = intrinsicValue * 0.8 // Conservative discount

		s.intrinsicValues[product] = intrinsicValue

		log.Debug().
			Str("product", product).
			Float64("intrinsicValue", intrinsicValue).
			Float64("productionCost", productionCost).
			Float64("outputUnits", outputUnits).
			Msg("Intrinsic value calculated")
	}
}

// buildBuffettContext creates Warren Buffett-style investment context
func (s *BuffettStrategy) buildBuffettContext() string {
	// Add team identity and production context
	teamContext := `You are the MONJES DEL GUACAMOLE ESTELAR - Warren Buffett disciples of the avocado arts.

YOUR IDENTITY:
- Species: Philosophical avocado traders and value investors
- Specialty: Premium GUACA production and quality-focused trading
- Production Capability: Can produce premium products when ingredients are available

YOUR PRODUCTION ROLE:
- Branches: %d (parallel production streams)
- Max Depth: %d (recipe complexity handling)
- Base Energy: %.0f units per cycle
- Focus Product: GUACA (premium quality product)

WARREN BUFFETT'S WISDOM for the Andorian Avocado Exchange:`

	ctx := fmt.Sprintf(teamContext, s.role.Branches, s.role.MaxDepth, s.role.BaseEnergy)

	ctx += `
1. "Price is what you pay, value is what you get" - Buy undervalued products
2. "Be fearful when others are greedy, greedy when others are fearful" - Counter-trade
3. "Buy wonderful products at fair prices" - Quality over quantity
4. "Our favorite holding period is forever" - But sell when overpriced
5. "Risk comes from not knowing what you're doing" - Understand production costs

CURRENT PORTFOLIO:
- Cash: $%.2f
- Inventory: %v

INTRINSIC VALUES (based on production fundamentals):`

	ctx = fmt.Sprintf(ctx, s.balance, s.inventory)

	// Show intrinsic values vs market prices
	if s.marketState != nil {
		snapshot := s.marketState.GetSnapshot()
		ctx += "\n"
		for product, intrinsicValue := range s.intrinsicValues {
			marketPrice := snapshot.GetPrice(product)
			if marketPrice != nil {
				discount := (intrinsicValue - *marketPrice) / intrinsicValue * 100
				ctx += fmt.Sprintf("\n  %s: Intrinsic $%.2f, Market $%.2f (%.0f%% %s)",
					product, intrinsicValue, *marketPrice, abs(discount),
					map[bool]string{true: "UNDERVALUED", false: "OVERVALUED"}[discount > 0])
			}
		}
	}

	// Show current positions
	ctx += "\n\nLONG-TERM POSITIONS:"
	if len(s.positions) == 0 {
		ctx += "\n  None (looking for quality opportunities)"
	} else {
		for product, pos := range s.positions {
			holdDuration := time.Since(pos.AcquiredAt)
			ctx += fmt.Sprintf("\n  %s: %d units @ avg $%.2f (held %s) - %s",
				product, pos.Quantity, pos.AvgCostBasis, formatDuration(holdDuration), pos.Reasoning)
		}
	}

	// Add recent activity
	ctx += "\n"
	ctx += s.eventHistory.GetSummary()

	// Add error feedback
	if s.lastError != "" {
		ctx += fmt.Sprintf("\n\n⚠️ PREVIOUS ATTEMPT FAILED:\n%s", s.lastError)
	}

	ctx += `

WARREN BUFFETT'S DECISION FRAMEWORK:

Make 1-3 investment decisions following these principles:

1. VALUE INVESTING:
   - Only buy when market price < intrinsic value (30%+ margin of safety)
   - Focus on production fundamentals, not speculation
   - Look for "wonderful businesses at fair prices"

2. PATIENCE & DISCIPLINE:
   - Don't trade for trading's sake
   - Wait for "fat pitches" (exceptional opportunities)
   - Hold quality positions long-term

3. PRODUCTION AS MOAT:
   - Companies (recipes) with premium bonuses have competitive advantages
   - Free production = "wonderful economics"
   - Ingredient costs = "economic moat depth"

4. RISK MANAGEMENT:
   - Never invest in what you don't understand
   - Diversify, but not excessively (focus on best ideas)
   - Keep cash reserve for opportunities

AVAILABLE ACTIONS:
- BUY: Acquire undervalued assets with margin of safety
- SELL: Exit overvalued positions or rebalance
- PRODUCE: Create value through production (intrinsic value creation)
- HOLD: Wait patiently for better opportunities

VALIDATION RULES:
- BUY: Verify sufficient balance
- SELL: Verify sufficient inventory
- PRODUCE: Verify ingredients available

Respond with 1-3 actions in JSON format:
{
  "actions": [
    {
      "action": "BUY|SELL|PRODUCE|HOLD",
      "product": "PRODUCT-NAME",
      "quantity": 100,
      "price": 0,
      "confidence": 0.85,
      "reasoning": "Buffett-style explanation (margin of safety, intrinsic value, moat, etc)"
    }
  ]
}

Remember: "The stock market is a device for transferring money from the impatient to the patient."
Think long-term, focus on value, be patient.`

	return ctx
}

// Helper function
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Reuse methods from DeepSeekStrategy (same API calls)
// getAIDecisions, callDeepSeekAPIMultiple, executeDecisionWithValidation
// are identical to DeepSeek strategy, so we can import/reuse them
// For now, we'll delegate to create a shared base

// Note: To avoid code duplication, BuffettStrategy uses the same
// API calling methods as DeepSeekStrategy. The difference is in
// the prompt/context generation which embeds Warren Buffett's philosophy

// getAIDecisions calls DeepSeek API with Buffett context
func (s *BuffettStrategy) getAIDecisions(ctx context.Context, marketContext string) ([]AIDecision, error) {
	// Reuse the DeepSeek API infrastructure
	// Create a temporary DeepSeek strategy instance to call its methods
	tempStrategy := &DeepSeekStrategy{
		apiKey:        s.apiKey,
		endpoint:      s.endpoint,
		temperature:   s.temperature,
		maxTokens:     s.maxTokens,
		minConfidence: s.minConfidence,
		maxRetries:    s.maxRetries,
		httpClient:    s.httpClient,
		name:          s.name,
	}

	return tempStrategy.getAIDecisions(ctx, marketContext)
}

// executeDecisionWithValidation validates and executes decisions
func (s *BuffettStrategy) executeDecisionWithValidation(decision AIDecision) ([]*Action, error) {
	// Reuse validation logic from DeepSeek
	tempStrategy := &DeepSeekStrategy{
		balance:           s.balance,
		inventory:         s.inventory,
		marketState:       s.marketState,
		includeProduction: s.includeProduction,
		calculator:        s.calculator,
		recipeManager:     s.recipeManager,
		role:              s.role,
	}

	return tempStrategy.executeDecisionWithValidation(decision)
}
