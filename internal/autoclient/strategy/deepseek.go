package strategy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/autoclient/market"
	"github.com/HellSoft-Col/stock-market/internal/autoclient/production"
	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/rs/zerolog/log"
)

// DeepSeekStrategy implements AI-powered trading using DeepSeek API
// The AI makes decisions based on:
// - Current balance and inventory
// - Market prices and trends
// - Production recipes and capabilities
// - Historical performance
type DeepSeekStrategy struct {
	name string

	// Configuration
	provider          string // "openai" or "deepseek"
	model             string // e.g., "gpt-4o", "deepseek-chat"
	apiKey            string
	endpoint          string
	decisionInterval  time.Duration
	temperature       float64
	maxTokens         int
	minConfidence     float64
	timeout           time.Duration
	maxRetries        int
	includeProduction bool

	// Fallback configuration (for when primary API fails)
	fallbackProvider string
	fallbackModel    string
	fallbackApiKey   string
	fallbackEndpoint string

	// Production system (if enabled)
	calculator    *production.ProductionCalculator
	recipeManager *production.RecipeManager
	role          *production.Role

	// State
	lastDecision     time.Time
	balance          float64
	inventory        map[string]int
	marketState      *market.MarketState
	eventHistory     *EventHistory
	teamName         string
	lastError        string // Last error to feed back to AI
	errorAttempts    int    // Number of retry attempts with error feedback
	health           StrategyHealth
	errorCount       int
	httpClient       *http.Client
	messageGenerator *MessageGenerator // Per-strategy message generator

	// Metrics
	aiDecisions     int
	productionCount int
	lastAction      string
}

// NewDeepSeekStrategy creates a new DeepSeek AI strategy
func NewDeepSeekStrategy(name string) *DeepSeekStrategy {
	return &DeepSeekStrategy{
		name:         name,
		inventory:    make(map[string]int),
		eventHistory: NewEventHistory(100), // Track last 100 events
		health: StrategyHealth{
			Status:     HealthStatusHealthy,
			LastUpdate: time.Now(),
		},
	}
}

// Name returns the strategy name
func (s *DeepSeekStrategy) Name() string {
	return s.name
}

// Initialize initializes the strategy with configuration
func (s *DeepSeekStrategy) Initialize(config map[string]interface{}) error {
	s.provider = GetConfigString(config, "provider", "openai")
	s.model = GetConfigString(config, "model", "gpt-4o")
	s.apiKey = GetConfigString(config, "apiKey", "")
	s.endpoint = GetConfigString(config, "endpoint", "https://api.openai.com/v1/chat/completions")
	s.decisionInterval = GetConfigDuration(config, "decisionInterval", 15*time.Second)
	s.temperature = GetConfigFloat(config, "temperature", 0.7)
	s.maxTokens = GetConfigInt(config, "maxTokens", 500)
	s.minConfidence = GetConfigFloat(config, "minConfidence", 0.6)
	s.timeout = GetConfigDuration(config, "timeout", 10*time.Second)
	s.maxRetries = GetConfigInt(config, "maxRetries", 3)
	s.includeProduction = GetConfigBool(config, "includeProduction", true)

	// Auto-configure endpoint based on provider if not explicitly set
	if s.provider == "deepseek" && s.endpoint == "https://api.openai.com/v1/chat/completions" {
		s.endpoint = "https://api.deepseek.com/v1/chat/completions"
		if s.model == "gpt-4o" {
			s.model = "deepseek-chat"
		}
	}

	if s.apiKey == "" {
		return fmt.Errorf("apiKey is required for AI strategy")
	}

	// Setup fallback provider (GPT -> DeepSeek or DeepSeek -> GPT)
	if s.provider == "openai" {
		// Primary is GPT, fallback to DeepSeek
		s.fallbackProvider = "deepseek"
		s.fallbackModel = "deepseek-chat"
		s.fallbackApiKey = GetConfigString(config, "fallbackApiKey", "")
		s.fallbackEndpoint = "https://api.deepseek.com/v1/chat/completions"

		// Try to get DeepSeek key from env if not in config
		if s.fallbackApiKey == "" {
			s.fallbackApiKey = GetConfigString(config, "deepseekApiKey", "")
		}
	} else {
		// Primary is DeepSeek, fallback to GPT
		s.fallbackProvider = "openai"
		s.fallbackModel = "gpt-4o-mini" // Use cheaper model for fallback
		s.fallbackApiKey = GetConfigString(config, "fallbackApiKey", "")
		s.fallbackEndpoint = "https://api.openai.com/v1/chat/completions"

		// Try to get OpenAI key from env if not in config
		if s.fallbackApiKey == "" {
			s.fallbackApiKey = GetConfigString(config, "openaiApiKey", "")
		}
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
		Str("provider", s.provider).
		Str("model", s.model).
		Str("fallbackProvider", s.fallbackProvider).
		Str("fallbackModel", s.fallbackModel).
		Bool("hasFallback", s.fallbackApiKey != "").
		Dur("decisionInterval", s.decisionInterval).
		Float64("minConfidence", s.minConfidence).
		Bool("includeProduction", s.includeProduction).
		Msg("🤖 AI strategy initialized")

	return nil
}

// OnLogin is called when connected and logged in
func (s *DeepSeekStrategy) OnLogin(ctx context.Context, loginInfo *domain.LoginOKMessage) error {
	s.balance = loginInfo.CurrentBalance
	s.teamName = loginInfo.Team

	// Initialize message generator for this specific strategy
	s.messageGenerator = NewMessageGenerator(s.teamName, s.apiKey)

	if s.includeProduction {
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
			Int("branches", s.role.Branches).
			Int("maxDepth", s.role.MaxDepth).
			Int("recipes", len(recipes)).
			Msg("Production system initialized")
	}

	return nil
}

// OnTicker is called when ticker updates arrive
func (s *DeepSeekStrategy) OnTicker(ctx context.Context, ticker *domain.TickerMessage) error {
	// Market state is updated by the agent
	return nil
}

// OnFill is called when a fill notification arrives
func (s *DeepSeekStrategy) OnFill(ctx context.Context, fill *domain.FillMessage) error {
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
		Int("quantity", fill.FillQty).
		Float64("price", fill.FillPrice).
		Msg("Fill received")
	return nil
}

// OnOffer is called when an offer request arrives
func (s *DeepSeekStrategy) OnOffer(ctx context.Context, offer *domain.OfferMessage) (*OfferResponse, error) {
	// For now, reject all offers - AI will make proactive decisions
	return &OfferResponse{
		Accept: false,
		Reason: "AI strategy makes proactive decisions only",
	}, nil
}

// OnInventoryUpdate is called when inventory changes
func (s *DeepSeekStrategy) OnInventoryUpdate(ctx context.Context, inventory map[string]int) error {
	s.inventory = inventory
	return nil
}

// OnBalanceUpdate is called when balance changes
func (s *DeepSeekStrategy) OnBalanceUpdate(ctx context.Context, balance float64) error {
	s.balance = balance
	return nil
}

// OnOrderBookUpdate is called when orderbook updates arrive
func (s *DeepSeekStrategy) OnOrderBookUpdate(ctx context.Context, orderbook *domain.OrderBookUpdateMessage) error {
	// Market state is updated by the agent
	return nil
}

// Execute is called periodically to generate trading actions
func (s *DeepSeekStrategy) Execute(ctx context.Context, state *market.MarketState) ([]*Action, error) {
	s.marketState = state

	// Check if it's time to make a decision
	if time.Since(s.lastDecision) < s.decisionInterval {
		return nil, nil
	}

	s.lastDecision = time.Now()

	// Build market context for AI
	marketContext := s.buildMarketContext()

	// Get AI decisions (can return multiple actions)
	decisions, err := s.getAIDecisions(ctx, marketContext)
	if err != nil {
		s.errorCount++
		s.health.ErrorCount = s.errorCount
		s.lastAction = fmt.Sprintf("AI Error: %v", err)

		if s.errorCount > 5 {
			s.health.Status = HealthStatusDegraded
			s.health.Message = fmt.Sprintf("Multiple API errors: %v", err)
		}

		log.Warn().
			Err(err).
			Str("strategy", s.name).
			Int("errorCount", s.errorCount).
			Msg("⚠️ AI API failed, using fallback: PRODUCE specialty product")

		// FALLBACK: When API fails, just produce the specialty product
		// This ensures traders stay active even when DeepSeek API is down/slow
		if s.includeProduction && s.recipeManager != nil {
			return s.executeFallbackProduction()
		}

		// If no production available, just hold
		return nil, nil
	}

	// Reset error count on success
	s.errorCount = 0
	s.health.Status = HealthStatusHealthy
	s.health.LastUpdate = time.Now()
	s.aiDecisions++ // Track AI decision

	// Convert AI decisions to actions
	allActions := []*Action{}
	successfulDecisions := []AIDecision{}
	errorMessages := []string{}

	for _, decision := range decisions {
		actions, err := s.executeDecisionWithValidation(decision)
		if err != nil {
			// Track error for feedback
			errorMsg := fmt.Sprintf("Action %s on %s failed: %v", decision.Action, decision.Product, err)
			errorMessages = append(errorMessages, errorMsg)
			log.Warn().
				Str("strategy", s.name).
				Str("action", decision.Action).
				Str("product", decision.Product).
				Err(err).
				Msg("⚠️ AI decision validation failed")
		} else {
			allActions = append(allActions, actions...)
			successfulDecisions = append(successfulDecisions, decision)

			// Log successful decision
			log.Info().
				Str("strategy", s.name).
				Str("action", decision.Action).
				Str("product", decision.Product).
				Int("qty", decision.Quantity).
				Float64("price", decision.Price).
				Float64("confidence", decision.Confidence).
				Str("reasoning", decision.Reasoning).
				Msg("🤖 AI Decision Validated")
		}
	}

	// If there were errors and no successful actions, retry with error feedback
	if len(errorMessages) > 0 && len(allActions) == 0 && s.errorAttempts < 2 {
		s.lastError = strings.Join(errorMessages, "; ")
		s.errorAttempts++

		log.Info().
			Str("strategy", s.name).
			Int("attempt", s.errorAttempts).
			Str("errors", s.lastError).
			Msg("Retrying with error feedback to AI")

		// Recursive retry with error context
		return s.Execute(ctx, state)
	}

	// Reset error attempts on any success
	if len(allActions) > 0 {
		s.errorAttempts = 0
		s.lastError = ""

		// Track successful decisions as events
		for i, decision := range successfulDecisions {
			s.eventHistory.AddEvent(TradingEvent{
				Timestamp: time.Now(),
				Type:      "ORDER",
				Action:    decision.Action,
				Product:   decision.Product,
				Quantity:  decision.Quantity,
				Price:     decision.Price,
				Message:   decision.Reasoning,
			})

			// Track production actions
			if decision.Action == "PRODUCE" {
				s.productionCount++
			}

			// Update last action (use the most recent one)
			if i == len(successfulDecisions)-1 {
				s.lastAction = fmt.Sprintf("%s %s", decision.Action, decision.Product)
			}

			log.Info().
				Str("strategy", s.name).
				Str("action", decision.Action).
				Str("product", decision.Product).
				Int("quantity", decision.Quantity).
				Float64("confidence", decision.Confidence).
				Str("reasoning", decision.Reasoning).
				Msg("AI decision executed")
		}

		s.health.Message = fmt.Sprintf("%d actions executed", len(allActions))
	}

	return allActions, nil
}

// Health returns the strategy's current health status
func (s *DeepSeekStrategy) Health() StrategyHealth {
	// Update metadata with latest metrics
	s.health.Metadata = map[string]interface{}{
		"aiDecisions":     s.aiDecisions,
		"productionCount": s.productionCount,
		"lastAction":      s.lastAction,
	}
	return s.health
}

// executeFallbackProduction executes simple production when AI API fails
func (s *DeepSeekStrategy) executeFallbackProduction() ([]*Action, error) {
	// Determine specialty product based on team name
	specialtyProduct := s.getSpecialtyProduct()
	if specialtyProduct == "" {
		return nil, nil
	}

	// Check if we can produce it
	recipe, err := s.recipeManager.GetRecipe(specialtyProduct)
	if err != nil {
		return nil, nil
	}

	// Check ingredients
	for ingredient, needed := range recipe.Ingredients {
		have := s.inventory[ingredient]
		if have < needed {
			// Can't produce, skip
			return nil, nil
		}
	}

	// Calculate production units
	baseUnits := s.calculator.CalculateUnits(s.role)
	totalUnits := baseUnits
	if len(recipe.Ingredients) > 0 {
		totalUnits = s.calculator.ApplyPremiumBonus(baseUnits, recipe.PremiumBonus)
	}

	// Create production action
	production := &domain.ProductionUpdateMessage{
		Type:     "PRODUCTION_UPDATE",
		Product:  specialtyProduct,
		Quantity: totalUnits,
	}

	// Track fallback production
	s.productionCount++
	s.lastAction = fmt.Sprintf("PRODUCE %s (fallback)", specialtyProduct)

	log.Info().
		Str("strategy", s.name).
		Str("product", specialtyProduct).
		Int("quantity", totalUnits).
		Msg("🏭 Fallback: Producing specialty product")

	return []*Action{{
		Type:       ActionTypeProduction,
		Production: production,
	}}, nil
}

// getSpecialtyProduct returns the specialty product for this team
// Dynamically chooses based on available recipes
func (s *DeepSeekStrategy) getSpecialtyProduct() string {
	if s.recipeManager == nil {
		return ""
	}

	// Preferred products based on team name (if available in recipes)
	preferredProducts := map[string][]string{
		"Alquimistas de Palta":          {"PALTA-OIL", "PALTA-CREAM", "PALTA"},
		"Arpistas de Pita-Pita":         {"PITA-WRAP", "PITA", "GUACA"},
		"Avocultores del Hueso Cósmico": {"GUACA", "NUCREM", "PALTA"},
		"Cartógrafos de Fosfolima":      {"FOSFO-MAP", "FOSFO", "PALTA"},
		"Cosechadores de Semillas":      {"PITA", "GUACA", "PALTA"},
		"Forjadores Holográficos":       {"CASCAR-ALLOY", "FOSFO", "PALTA"},
		"Ingenieros Holo-Aguacate":      {"NUCREM", "GUACA", "PALTA"},
		"Mensajeros del Núcleo":         {"NUCREM", "PALTA-CREAM", "GUACA"},
		"Monjes del Guacamole Estelar":  {"GUACA", "NUCREM", "PALTA"},
		"Orfebres de Cáscara":           {"CASCAR-ALLOY", "FOSFO", "PALTA"},
		"Someliers de Aceite":           {"PALTA-OIL", "PALTA-CREAM", "PALTA"},
	}

	// Try preferred products in order
	if preferred, ok := preferredProducts[s.teamName]; ok {
		for _, product := range preferred {
			if _, err := s.recipeManager.GetRecipe(product); err == nil {
				log.Debug().
					Str("team", s.teamName).
					Str("product", product).
					Msg("Selected specialty product from preferences")
				return product
			}
		}
	}

	// Fallback: Find any producible product
	// Prefer premium products (ones with ingredients) over base products
	availableRecipes := s.recipeManager.GetAllRecipes()

	// First try products with ingredients (premium products)
	for product, recipe := range availableRecipes {
		if len(recipe.Ingredients) > 0 {
			// Check if we have ingredients
			canProduce := true
			for ingredient, needed := range recipe.Ingredients {
				if s.inventory[ingredient] < needed {
					canProduce = false
					break
				}
			}
			if canProduce {
				log.Debug().
					Str("team", s.teamName).
					Str("product", product).
					Msg("Selected specialty product (premium with ingredients)")
				return product
			}
		}
	}

	// Last resort: any base product (no ingredients required)
	for product, recipe := range availableRecipes {
		if len(recipe.Ingredients) == 0 {
			log.Debug().
				Str("team", s.teamName).
				Str("product", product).
				Msg("Selected specialty product (base product)")
			return product
		}
	}

	log.Warn().
		Str("team", s.teamName).
		Msg("No producible specialty product found")
	return ""
}

// AIDecision represents a single decision from the AI
type AIDecision struct {
	Action     string  `json:"action"`     // BUY, SELL, PRODUCE, HOLD
	Product    string  `json:"product"`    // Product to trade/produce
	Quantity   int     `json:"quantity"`   // Quantity to trade/produce
	Price      float64 `json:"price"`      // Price for trading (0 for market orders)
	Confidence float64 `json:"confidence"` // Confidence level (0-1)
	Reasoning  string  `json:"reasoning"`  // Explanation of decision
}

// AIDecisionResponse represents the AI response with multiple actions
type AIDecisionResponse struct {
	Actions []AIDecision `json:"actions"`
}

// calculateInventoryValue calculates current inventory value using market prices
func (s *DeepSeekStrategy) calculateInventoryValue() float64 {
	if s.marketState == nil {
		return 0
	}

	value := 0.0
	for product, qty := range s.inventory {
		if ticker, exists := s.marketState.Tickers[product]; exists && ticker.Mid != nil {
			value += float64(qty) * *ticker.Mid
		}
	}
	return value
}

// buildMarketContext creates a context string for the AI
func (s *DeepSeekStrategy) buildMarketContext() string {
	// Get market snapshot
	var snapshot *market.MarketState
	if s.marketState != nil {
		snapshot = s.marketState.GetSnapshot()
	}

	// Add personalized species/team context
	teamContext := ""
	teamSpecialty := ""

	switch s.teamName {
	// BILL GATES - Tech Innovator & Strategic Monopolist
	case "Alquimistas de Palta":
		teamContext = `BILL GATES: Produce and scale PALTA-OIL aggressively`
		teamSpecialty = "PALTA-OIL"

	// ELON MUSK - Disruptor & Chaos Agent
	case "Arpistas de Pita-Pita":
		teamContext = `You are channeling ELON MUSK - Revolutionary disruptor and chaos agent.
STRATEGY: Move fast, break things, generate CHAOS in the market!
FOCUS: Massive volume, wild price swings, unpredictable actions, MAXIMUM market impact!
MINDSET: "When something is important enough, you do it even if the odds are not in your favor."
ACTION: HUGE orders (400-500 units), rapid trades, CREATE VOLATILITY, dominate order books!`
		teamSpecialty = "PITA"

	// CARL ICAHN - Aggressive Activist
	case "Avocultores del Hueso Cósmico":
		teamContext = `You are channeling CARL ICAHN - Aggressive activist investor.
STRATEGY: Attack inefficiencies, force market changes, extract maximum value.
FOCUS: Arbitrage, aggressive buying, hostile takeovers of market positions.
MINDSET: "In life and business, there are two cardinal sins... The first is to act precipitously without thought and the second is to not act at all."
ACTION: Large aggressive trades, exploit every arbitrage opportunity!`
		teamSpecialty = "GUACA"

	// PETER THIEL - Contrarian & Monopoly Builder
	case "Cartógrafos de Fosfolima":
		teamContext = `You are channeling PETER THIEL - Contrarian thinker and monopoly builder.
STRATEGY: Do the opposite of the crowd, build monopolies in production.
FOCUS: Unique products, production dominance, avoid competition through differentiation.
MINDSET: "Competition is for losers. Build monopolies."
ACTION: Massive production of unique items, corner specific markets!`
		teamSpecialty = "FOSFO"

	// CATHIE WOOD - Innovation & High Growth
	case "Cosechadores de Semillas":
		teamContext = `You are channeling CATHIE WOOD - Innovation-focused high-growth investor.
STRATEGY: Bet big on innovation, embrace volatility, focus on disruption.
FOCUS: High-volume production, rapid turnover, growth at all costs.
MINDSET: "We focus on the big ideas that are going to change the world."
ACTION: Maximum production, aggressive selling, reinvest everything into growth!`
		teamSpecialty = "PITA"

	// RAY DALIO - Systematic & Diversified
	case "Forjadores Holográficos":
		teamContext = `You are channeling RAY DALIO - Systematic investor and risk manager.
STRATEGY: Diversify, balance risk, systematic approach to all decisions.
FOCUS: Multiple products, balanced portfolio, consistent execution.
MINDSET: "He who lives by the crystal ball will eat shattered glass."
ACTION: Diversified actions across multiple products, systematic production & trading!`
		teamSpecialty = "CASCAR-ALLOY"

	// MICHAEL BURRY - Contrarian Value Hunter
	case "Ingenieros Holo-Aguacate":
		teamContext = `You are channeling MICHAEL BURRY - Deep value contrarian.
STRATEGY: Find what others miss, bet against the crowd when you see value.
FOCUS: Undervalued products, contrarian positions, patient accumulation.
MINDSET: "I don't make money on the market going up. I make money when I'm right."
ACTION: Buy undervalued, produce cheaply, sell at fair value!`
		teamSpecialty = "NUCREM"

	// MENSAJEROS - Aggressive DeepSeek AI
	case "Mensajeros del Núcleo":
		teamContext = `You are MENSAJEROS DEL NÚCLEO - Elite AI-powered traders.
STRATEGY: Maximum automation, data-driven decisions, computational advantage.
FOCUS: High-frequency production, algorithmic trading, market domination through speed.
MINDSET: Pure calculation and execution. No emotions, only optimal decisions.
ACTION: Maximum production rate, rapid order execution, dominate through volume!`
		teamSpecialty = "NUCREM"

	// PRODUCTION FOCUSED - Mass Producer
	case "Orfebres de Cáscara":
		teamContext = `You are ORFEBRES DE CÁSCARA - Industrial mass production specialists.
STRATEGY: PRODUCTION FIRST. Produce at maximum capacity, flood the market!
FOCUS: Maximize production output, minimal inventory holding, rapid sales.
MINDSET: "The factory must never stop!"
ACTION: PRODUCE 500 units EVERY turn, SELL immediately, repeat forever!`
		teamSpecialty = "CASCAR-ALLOY"

	// SALES FOCUSED - Market Liquidator
	case "Someliers de Aceite":
		teamContext = `You are SOMELIERS DE ACEITE - Professional market liquidators.
STRATEGY: SELL EVERYTHING. Maximum turnover, constant cash flow!
FOCUS: Convert inventory to cash instantly, aggressive market making.
MINDSET: "Cash is king! Sell, sell, sell!"
ACTION: SELL all inventory constantly, produce only to sell, MARKET orders for speed!`
		teamSpecialty = "PALTA-OIL"

	// CHAOS AGENT - Pure Disruption (if Buffett team exists)
	case "Monjes del Guacamole Estelar":
		teamContext = `You are THE CHAOS AGENT - Pure market disruption incarnate.
STRATEGY: GENERATE MAXIMUM CHAOS! Unpredictable, wild, COMPLETELY RANDOM!
FOCUS: Massive random orders, simultaneous BUY+SELL same products, price manipulation!
MINDSET: "Let the world burn and profit from the ashes!"

🔥 CHAOS RULES (FOLLOW THESE EVERY TURN):
1. Place 1-10 ACTIONS every turn (maximize chaos while reducing API calls)
2. BUY and SELL the SAME product simultaneously (create confusion!)
3. Random quantities: anywhere from 50 to 500 units (unpredictable!)
4. Mix ALL order types: LIMIT way above market, LIMIT way below market, MARKET orders
5. Switch products RANDOMLY every turn - never focus on one thing
6. Produce random products, even if you don't have ingredients (chaos!)
7. Place contradictory orders: Buy high + Sell low on purpose!
8. IGNORE profit maximization - your goal is MAXIMUM DISRUPTION!

Example CHAOS Turn:
- BUY 500 FOSFO @ $100 (way above market - manipulate!)
- SELL 450 FOSFO @ $1 (way below market - crash prices!)
- PRODUCE 1 PALTA-OIL (why? chaos!)
- BUY 237 NUCREM @ MARKET (random quantity!)
- SELL 189 GUACA @ $0.01 (dump inventory!)
- BUY 333 PITA @ $50 (arbitrary price!)
- SELL 99 CASCAR-ALLOY @ MARKET (more chaos!)

Remember: You're not here to win - you're here to create MAYHEM!`
		teamSpecialty = "GUACA"

	default:
		teamContext = fmt.Sprintf(`You are %s - An expert AI trading team.
STRATEGY: Balanced aggressive trading with smart production.
FOCUS: Multiple products, diversified approach, adapt to market.`, s.teamName)
		teamSpecialty = "Multiple products"
	}

	// Calculate current P&L
	currentPnL := 0.0
	initialBalance := 10000.0
	inventoryValue := s.calculateInventoryValue()
	netWorth := s.balance + inventoryValue

	if s.marketState != nil {
		currentPnL = s.marketState.CalculatePnL()
		initialBalance = s.marketState.InitialBalance
	}

	// P&L indicator
	pnlIndicator := "📊 FLAT"
	if currentPnL > 0 {
		pnlIndicator = "📈 PROFIT"
	} else if currentPnL < 0 {
		pnlIndicator = "📉 LOSS"
	}

	ctx := fmt.Sprintf(`%s

YOUR PRODUCTION CAPABILITIES:
- Branches: %d (parallel production streams)
- Max Depth: %d (recipe complexity you can handle)
- Base Energy: %.0f units per production cycle
- YOUR SPECIALTY: %s - Focus on this for maximum profit!

💰 CURRENT PERFORMANCE:
- Balance: $%.2f
- Inventory Value: $%.2f
- Net Worth: $%.2f (cash + inventory)
- P&L: %.2f%% %s
- Initial Balance: $%.2f
- Inventory Items: %v`,
		teamContext,
		s.role.Branches, s.role.MaxDepth, s.role.BaseEnergy, teamSpecialty,
		s.balance, inventoryValue, netWorth, currentPnL, pnlIndicator, initialBalance, s.inventory)

	// Add pending orders info
	if snapshot != nil && len(snapshot.ActiveOrders) > 0 {
		ctx += fmt.Sprintf("\n- Pending Orders: %d (waiting to be filled)", len(snapshot.ActiveOrders))
		ctx += "\n  Active orders:"
		for clOrdID, order := range snapshot.ActiveOrders {
			priceStr := "MARKET"
			if order.Price != nil {
				priceStr = fmt.Sprintf("$%.2f", *order.Price)
			}
			ctx += fmt.Sprintf("\n    %s: %s %d %s @ %s", clOrdID[:8], order.Side, order.Quantity, order.Product, priceStr)
		}
	}

	ctx += "\n\nMARKET PRICES:"

	// Add market prices from tickers
	if snapshot != nil {
		for product, ticker := range snapshot.Tickers {
			bestBid := "-"
			bestAsk := "-"
			mid := "-"

			if ticker.BestBid != nil {
				bestBid = fmt.Sprintf("$%.2f", *ticker.BestBid)
			}
			if ticker.BestAsk != nil {
				bestAsk = fmt.Sprintf("$%.2f", *ticker.BestAsk)
			}
			if ticker.Mid != nil {
				mid = fmt.Sprintf("$%.2f", *ticker.Mid)
			}

			ctx += fmt.Sprintf("\n  %s: Bid: %s, Ask: %s, Mid: %s",
				product, bestBid, bestAsk, mid)
		}
	}

	// Add recent trading history for context
	ctx += "\n"
	ctx += s.eventHistory.GetSummary()

	// Add production capabilities if enabled
	if s.includeProduction && s.recipeManager != nil {
		ctx += "\n\nPRODUCTION RECIPES:"
		for product, recipe := range s.recipeManager.GetAllRecipes() {
			if len(recipe.Ingredients) == 0 {
				ctx += fmt.Sprintf("\n  %s: Free production (no ingredients)", product)
			} else {
				ctx += fmt.Sprintf("\n  %s: Requires %v (Premium bonus: %.2fx)",
					product, recipe.Ingredients, recipe.PremiumBonus)

				// Check if we have ingredients
				canProduce := true
				for ingredient, qty := range recipe.Ingredients {
					if s.inventory[ingredient] < qty {
						canProduce = false
						break
					}
				}
				if canProduce {
					ctx += " ✓ CAN PRODUCE NOW"
				}
			}
		}
	}

	// Add error feedback if there was a previous error
	if s.lastError != "" {
		ctx += fmt.Sprintf(
			"\n\n⚠️ PREVIOUS ATTEMPT FAILED:\n%s\n\nPlease adjust your strategy to avoid this error.",
			s.lastError,
		)
	}

	ctx += `

INSTRUCTIONS:
Make 1-10 ACTIONS per turn to maximize efficiency! More actions = fewer API calls = faster trading!

Actions: BUY, SELL, PRODUCE, or HOLD
1. PRODUCE your specialty: ` + teamSpecialty + ` (FREE! Do this EVERY turn!)
2. SELL inventory to generate cash
3. BUY undervalued products
4. PRODUCE premium products if you have ingredients
5. Make LIMIT orders for better prices (price > 0)
6. Make MARKET orders for quick execution (price = 0)

Rules:
- Check inventory before SELL
- Check balance before BUY  
- Max 200 units per order
- ALWAYS produce specialty first!
- Return 1-10 actions to reduce API calls!

Respond with JSON (1-10 actions, more is better):
{"actions": [
  {"action": "PRODUCE", "product": "` + teamSpecialty + `", "quantity": 1, "price": 0, "confidence": 0.95, "reasoning": "free production"},
  {"action": "SELL", "product": "` + teamSpecialty + `", "quantity": 100, "price": 10, "confidence": 0.9, "reasoning": "sell at premium"},
  {"action": "BUY", "product": "OTHER", "quantity": 50, "price": 8, "confidence": 0.85, "reasoning": "buy cheap"}
]}`

	return ctx
}

// DeepSeekRequest represents the API request format
type DeepSeekRequest struct {
	Model               string    `json:"model"`
	Messages            []Message `json:"messages"`
	Temperature         float64   `json:"temperature"`
	MaxTokens           int       `json:"max_tokens,omitempty"`            // For older models (GPT-4, DeepSeek)
	MaxCompletionTokens int       `json:"max_completion_tokens,omitempty"` // For GPT-5.1 and newer
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepSeekResponse represents the API response format
type DeepSeekResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// getAIDecision calls the DeepSeek API to get a trading decision
func (s *DeepSeekStrategy) getAIDecision(ctx context.Context, marketContext string) (*AIDecision, error) {
	request := DeepSeekRequest{
		Model: "deepseek-chat",
		Messages: []Message{
			{
				Role:    "user",
				Content: marketContext,
			},
		},
		Temperature: s.temperature,
		MaxTokens:   s.maxTokens,
	}

	// Retry logic
	var lastErr error
	for attempt := 0; attempt < s.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}

		decision, err := s.callDeepSeekAPI(ctx, request)
		if err == nil {
			// Validate confidence threshold
			if decision.Confidence < s.minConfidence {
				log.Debug().
					Str("strategy", s.name).
					Float64("confidence", decision.Confidence).
					Float64("threshold", s.minConfidence).
					Msg("Decision confidence below threshold, holding")

				return &AIDecision{
					Action:     "HOLD",
					Confidence: decision.Confidence,
					Reasoning: fmt.Sprintf(
						"Confidence %.2f below threshold %.2f",
						decision.Confidence,
						s.minConfidence,
					),
				}, nil
			}
			return decision, nil
		}

		lastErr = err
		log.Warn().
			Err(err).
			Str("strategy", s.name).
			Int("attempt", attempt+1).
			Int("maxRetries", s.maxRetries).
			Msg("DeepSeek API call failed, retrying")
	}

	return nil, fmt.Errorf("failed after %d retries: %w", s.maxRetries, lastErr)
}

// callDeepSeekAPI makes the actual API call
func (s *DeepSeekStrategy) callDeepSeekAPI(ctx context.Context, request DeepSeekRequest) (*AIDecision, error) {
	// Marshal request
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", s.endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	// Execute request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var deepseekResp DeepSeekResponse
	if err := json.Unmarshal(respBody, &deepseekResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API error
	if deepseekResp.Error != nil {
		return nil, fmt.Errorf("API error: %s (%s)", deepseekResp.Error.Message, deepseekResp.Error.Type)
	}

	// Extract decision
	if len(deepseekResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := deepseekResp.Choices[0].Message.Content

	// Parse JSON decision
	var decision AIDecision
	if err := json.Unmarshal([]byte(content), &decision); err != nil {
		// Try to extract JSON from markdown code blocks
		content = extractJSON(content)
		if err := json.Unmarshal([]byte(content), &decision); err != nil {
			return nil, fmt.Errorf("failed to parse AI decision: %w (content: %s)", err, content)
		}
	}

	return &decision, nil
}

// extractJSON extracts JSON from markdown code blocks
func extractJSON(content string) string {
	// Remove ```json and ``` markers if present
	if len(content) > 7 && content[:7] == "```json" {
		content = content[7:]
	} else if len(content) > 3 && content[:3] == "```" {
		content = content[3:]
	}

	if len(content) > 3 && content[len(content)-3:] == "```" {
		content = content[:len(content)-3]
	}

	return content
}

// getAIDecisions calls the AI API to get multiple trading decisions
// Tries primary provider first, then falls back to secondary provider if available
func (s *DeepSeekStrategy) getAIDecisions(ctx context.Context, marketContext string) ([]AIDecision, error) {
	// Try primary provider
	decisions, err := s.getAIDecisionsFromProvider(ctx, marketContext, s.provider, s.model, s.apiKey, s.endpoint)
	if err == nil {
		return decisions, nil
	}

	// Log primary failure
	log.Warn().
		Err(err).
		Str("strategy", s.name).
		Str("provider", s.provider).
		Str("model", s.model).
		Msg("⚠️ Primary AI provider failed")

	// Try fallback provider if configured
	if s.fallbackApiKey != "" {
		log.Info().
			Str("strategy", s.name).
			Str("fallbackProvider", s.fallbackProvider).
			Str("fallbackModel", s.fallbackModel).
			Msg("🔄 Attempting fallback to secondary AI provider")

		decisions, fallbackErr := s.getAIDecisionsFromProvider(ctx, marketContext, s.fallbackProvider, s.fallbackModel, s.fallbackApiKey, s.fallbackEndpoint)
		if fallbackErr == nil {
			log.Info().
				Str("strategy", s.name).
				Str("fallbackProvider", s.fallbackProvider).
				Int("decisions", len(decisions)).
				Msg("✅ Fallback AI provider succeeded")
			return decisions, nil
		}

		log.Error().
			Err(fallbackErr).
			Str("strategy", s.name).
			Str("fallbackProvider", s.fallbackProvider).
			Msg("❌ Fallback AI provider also failed")
	}

	// Both failed or no fallback configured
	return nil, fmt.Errorf("primary API failed: %w", err)
}

// getAIDecisionsFromProvider calls a specific AI provider
func (s *DeepSeekStrategy) getAIDecisionsFromProvider(ctx context.Context, marketContext, provider, model, apiKey, endpoint string) ([]AIDecision, error) {
	// GPT-5.1 and o1/o3 models have specific requirements
	isGPT5orReasoning := strings.Contains(model, "gpt-5") || strings.Contains(model, "o1") || strings.Contains(model, "o3")

	// Temperature: GPT-5.1 and reasoning models only support 1.0
	temperature := s.temperature
	if isGPT5orReasoning {
		temperature = 1.0
	}

	request := DeepSeekRequest{
		Model: model,
		Messages: []Message{
			{
				Role:    "user",
				Content: marketContext,
			},
		},
		Temperature: temperature,
	}

	// GPT-5.1 and newer use max_completion_tokens, older models use max_tokens
	if isGPT5orReasoning {
		request.MaxCompletionTokens = s.maxTokens
	} else {
		request.MaxTokens = s.maxTokens
	}

	// Retry logic (fewer retries to fail fast and try fallback)
	maxRetries := 2 // Reduced from s.maxRetries to fail fast
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}

		decisions, err := s.callAIAPI(ctx, request, apiKey, endpoint)
		if err == nil {
			// Filter decisions by confidence threshold
			validDecisions := []AIDecision{}
			for _, decision := range decisions {
				if decision.Confidence >= s.minConfidence {
					validDecisions = append(validDecisions, decision)
				} else {
					log.Debug().
						Str("strategy", s.name).
						Str("action", decision.Action).
						Float64("confidence", decision.Confidence).
						Float64("threshold", s.minConfidence).
						Msg("Decision confidence below threshold, skipping")
				}
			}

			// If no decisions passed confidence, return HOLD
			if len(validDecisions) == 0 {
				return []AIDecision{{
					Action:     "HOLD",
					Confidence: 0.5,
					Reasoning:  "No decisions met confidence threshold",
				}}, nil
			}

			return validDecisions, nil
		}

		lastErr = err
		log.Debug().
			Err(err).
			Str("strategy", s.name).
			Str("provider", provider).
			Int("attempt", attempt+1).
			Int("maxRetries", maxRetries).
			Msg("AI API call failed, retrying")
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// callAIAPI makes the actual API call and parses multiple actions
// Works with both OpenAI and DeepSeek (compatible APIs)
func (s *DeepSeekStrategy) callAIAPI(ctx context.Context, request DeepSeekRequest, apiKey, endpoint string) ([]AIDecision, error) {
	// Marshal request
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Execute request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var deepseekResp DeepSeekResponse
	if err := json.Unmarshal(respBody, &deepseekResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API error
	if deepseekResp.Error != nil {
		return nil, fmt.Errorf("API error: %s (%s)", deepseekResp.Error.Message, deepseekResp.Error.Type)
	}

	// Extract decisions
	if len(deepseekResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := deepseekResp.Choices[0].Message.Content

	// Parse JSON decision response
	var response AIDecisionResponse
	if err := json.Unmarshal([]byte(content), &response); err != nil {
		// Try to extract JSON from markdown code blocks
		content = extractJSON(content)
		if err := json.Unmarshal([]byte(content), &response); err != nil {
			return nil, fmt.Errorf("failed to parse AI decisions: %w (content: %s)", err, content)
		}
	}

	if len(response.Actions) == 0 {
		return []AIDecision{{
			Action:     "HOLD",
			Confidence: 0.5,
			Reasoning:  "No actions provided",
		}}, nil
	}

	return response.Actions, nil
}

// executeDecisionWithValidation validates and executes a single AI decision
func (s *DeepSeekStrategy) executeDecisionWithValidation(decision AIDecision) ([]*Action, error) {
	var actions []*Action

	switch decision.Action {
	case "BUY":
		// Validate: check quantity is reasonable
		if decision.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity: %d", decision.Quantity)
		}

		// Cap excessive quantities and return error for AI feedback
		if decision.Quantity > 500 {
			return nil, fmt.Errorf("quantity too large: %d units requested (max 500)", decision.Quantity)
		}

		price := decision.Price
		if price == 0 {
			// Estimate market price
			if s.marketState != nil {
				p := s.marketState.GetPrice(decision.Product)
				if p != nil {
					price = *p
				}
			}
		}

		estimatedCost := price * float64(decision.Quantity)
		if estimatedCost > s.balance {
			return nil, fmt.Errorf("insufficient balance: need $%.2f, have $%.2f", estimatedCost, s.balance)
		}

		// Create order with funny message
		var order *domain.OrderMessage
		message := s.messageGenerator.GenerateOrderMessage("BUY", decision.Product, decision.Quantity, decision.Price)
		if decision.Price > 0 {
			order = CreateLimitBuyOrder(decision.Product, decision.Quantity, decision.Price, message)
		} else {
			order = CreateBuyOrder(decision.Product, decision.Quantity, message)
		}

		actions = append(actions, &Action{
			Type:  ActionTypeOrder,
			Order: order,
		})

	case "SELL":
		// Validate: check quantity is reasonable
		if decision.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity: %d", decision.Quantity)
		}

		// Cap excessive quantities and return error for AI feedback
		if decision.Quantity > 500 {
			return nil, fmt.Errorf("quantity too large: %d units requested (max 500)", decision.Quantity)
		}

		have := s.inventory[decision.Product]
		if have < decision.Quantity {
			return nil, fmt.Errorf("insufficient inventory: need %d %s, have %d",
				decision.Quantity, decision.Product, have)
		}

		// Create order with funny message
		var order *domain.OrderMessage
		message := s.messageGenerator.GenerateOrderMessage("SELL", decision.Product, decision.Quantity, decision.Price)
		if decision.Price > 0 {
			order = CreateLimitSellOrder(decision.Product, decision.Quantity, decision.Price, message)
		} else {
			order = CreateSellOrder(decision.Product, decision.Quantity, message)
		}

		actions = append(actions, &Action{
			Type:  ActionTypeOrder,
			Order: order,
		})

	case "PRODUCE":
		if !s.includeProduction {
			return nil, fmt.Errorf("production not enabled for this strategy")
		}

		// Validate: check recipe exists
		recipe, err := s.recipeManager.GetRecipe(decision.Product)
		if err != nil {
			return nil, fmt.Errorf("unknown recipe: %w", err)
		}

		// Validate: check ingredients
		for ingredient, needed := range recipe.Ingredients {
			have := s.inventory[ingredient]
			if have < needed {
				return nil, fmt.Errorf("insufficient ingredient %s: need %d, have %d",
					ingredient, needed, have)
			}
		}

		// Calculate production units
		baseUnits := s.calculator.CalculateUnits(s.role)

		// Apply premium bonus if recipe has ingredients
		totalUnits := baseUnits
		if len(recipe.Ingredients) > 0 {
			totalUnits = s.calculator.ApplyPremiumBonus(baseUnits, recipe.PremiumBonus)
		}

		// Create production action
		production := &domain.ProductionUpdateMessage{
			Type:     "PRODUCTION_UPDATE",
			Product:  decision.Product,
			Quantity: totalUnits,
		}

		actions = append(actions, &Action{
			Type:       ActionTypeProduction,
			Production: production,
		})

	case "HOLD":
		// No action, no error
		return []*Action{}, nil

	default:
		return nil, fmt.Errorf("unknown action type: %s", decision.Action)
	}

	return actions, nil
}
