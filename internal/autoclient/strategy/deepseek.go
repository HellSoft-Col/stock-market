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
	apiKey            string
	endpoint          string
	decisionInterval  time.Duration
	temperature       float64
	maxTokens         int
	minConfidence     float64
	timeout           time.Duration
	maxRetries        int
	includeProduction bool

	// Production system (if enabled)
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
	lastError     string // Last error to feed back to AI
	errorAttempts int    // Number of retry attempts with error feedback
	health        StrategyHealth
	errorCount    int
	httpClient    *http.Client
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
	s.apiKey = GetConfigString(config, "apiKey", "")
	s.endpoint = GetConfigString(config, "endpoint", "https://api.deepseek.com/v1/chat/completions")
	s.decisionInterval = GetConfigDuration(config, "decisionInterval", 15*time.Second)
	s.temperature = GetConfigFloat(config, "temperature", 0.7)
	s.maxTokens = GetConfigInt(config, "maxTokens", 500)
	s.minConfidence = GetConfigFloat(config, "minConfidence", 0.6)
	s.timeout = GetConfigDuration(config, "timeout", 10*time.Second)
	s.maxRetries = GetConfigInt(config, "maxRetries", 3)
	s.includeProduction = GetConfigBool(config, "includeProduction", true)

	if s.apiKey == "" {
		return fmt.Errorf("apiKey is required for DeepSeek strategy")
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
		Float64("minConfidence", s.minConfidence).
		Bool("includeProduction", s.includeProduction).
		Msg("DeepSeek AI strategy initialized")

	return nil
}

// OnLogin is called when connected and logged in
func (s *DeepSeekStrategy) OnLogin(ctx context.Context, loginInfo *domain.LoginOKMessage) error {
	s.balance = loginInfo.CurrentBalance
	s.teamName = loginInfo.Team

	// Initialize message generator with team name
	InitMessageGenerator(s.teamName, s.apiKey)

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

		if s.errorCount > 5 {
			s.health.Status = HealthStatusDegraded
			s.health.Message = fmt.Sprintf("Multiple API errors: %v", err)
		}

		log.Error().
			Err(err).
			Str("strategy", s.name).
			Int("errorCount", s.errorCount).
			Msg("Failed to get AI decisions")

		return nil, err
	}

	// Reset error count on success
	s.errorCount = 0
	s.health.Status = HealthStatusHealthy
	s.health.LastUpdate = time.Now()

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
				Msg("AI decision validation failed")
		} else {
			allActions = append(allActions, actions...)
			successfulDecisions = append(successfulDecisions, decision)
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
	return s.health
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
	case "Mensajeros del Núcleo":
		teamContext = `You are the MENSAJEROS DEL NÚCLEO - Elite NUCREM miners and cosmic messengers.
SPECIES TRAITS: Efficient, strategic, technologically advanced. Masters of NUCREM extraction and trading.`
		teamSpecialty = "NUCREM"

	case "Monjes del Guacamole Estelar":
		teamContext = `You are the MONJES DEL GUACAMOLE ESTELAR - Warren Buffett disciples of the avocado arts.
SPECIES TRAITS: Patient value investors, quality-focused, long-term thinking, philosophical traders.`
		teamSpecialty = "GUACA (premium)"

	default:
		teamContext = fmt.Sprintf(`You are %s - An expert trading team on the Andorian Avocado Exchange.`, s.teamName)
		teamSpecialty = "Multiple products"
	}

	ctx := fmt.Sprintf(`%s

YOUR PRODUCTION CAPABILITIES:
- Branches: %d (parallel production streams)
- Max Depth: %d (recipe complexity you can handle)
- Base Energy: %.0f units per production cycle
- YOUR SPECIALTY: %s - Focus on this for maximum profit!

CURRENT STATUS:
- Balance: $%.2f
- Inventory: %v`,
		teamContext,
		s.role.Branches, s.role.MaxDepth, s.role.BaseEnergy, teamSpecialty,
		s.balance, s.inventory)

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
You should make 2-4 DIFFERENT decisions per turn to maximize profit and market activity!

Available actions:
1. BUY products (MARKET with price=0 OR LIMIT with specific price)
2. SELL inventory products (MARKET or LIMIT)
3. PRODUCE products (free or with ingredients)
4. HOLD on specific products (wait for better price)

🎯 DECISION STRATEGY - Make MULTIPLE actions per turn:
✓ Example Turn: PRODUCE ` + teamSpecialty + ` (free) → SELL 50 ` + teamSpecialty + ` (LIMIT @ $X) → BUY 30 PITA (LIMIT @ $Y)
✓ Diversify: Don't just produce OR sell - do BOTH plus strategic buying!
✓ BE AGGRESSIVE: Your specialty product (` + teamSpecialty + `) costs ZERO - produce and sell it EVERY turn!

📊 ORDER TYPE STRATEGY (60% LIMIT, 40% MARKET):
- LIMIT orders (PREFERRED): Better profit margins, strategic positioning
  * LIMIT BUY: Set 3-7% BELOW mid price (e.g., Mid $10 → Buy @ $9.30-$9.70)
  * LIMIT SELL: Set 3-7% ABOVE mid price (e.g., Mid $10 → Sell @ $10.30-$10.70)
- MARKET orders: When you need guaranteed execution
  * Use when: Inventory full, urgent sale, or grabbing good deals

💰 PROFIT MAXIMIZATION TACTICS:
1. FREE PRODUCTION ADVANTAGE: Your specialty (` + teamSpecialty + `) has ZERO cost!
   → Produce it EVERY turn and sell at ANY positive price = pure profit!
   
2. SPREAD CAPTURE: Place both buy and sell orders on same product
   → Example: Buy FOSFO @ $7.50, Sell FOSFO @ $8.50 = $1 profit per unit
   
3. INVENTORY CHURN: Don't hoard - constantly produce and sell
   → High turnover = more profit opportunities
   
4. INGREDIENT ARBITRAGE: Buy cheap ingredients, produce premium, sell high
   → Calculate: Premium sell price > (ingredient costs + 10% margin)?

5. MARKET MAKING: Help provide liquidity while earning spreads
   → Place orders on multiple products simultaneously

⚠️ VALIDATION RULES:
- For SELL: Check you have enough inventory BEFORE selling
- For BUY: Check balance is sufficient (quantity × price ≤ current balance)
- For PRODUCE: Check you have required ingredients BEFORE producing
- Quantity: Reasonable positive integers (typically 10-200 units)
- Price: Set to 0 for MARKET orders, or specify price for LIMIT orders
- NEVER use quantities over 500 - orders will fail validation
- Always verify (quantity × estimated_price) <= your balance

Respond ONLY with valid JSON in this exact format:
{
  "actions": [
    {
      "action": "BUY|SELL|PRODUCE|HOLD",
      "product": "PRODUCT-NAME",
      "quantity": 100,
      "price": 0,
      "confidence": 0.75,
      "reasoning": "Brief explanation of this action"
    }
  ]
}

Examples:
1. LIMIT buy below market:
   {"actions": [{"action": "BUY", "product": "FOSFO", "quantity": 50, "price": 7.5, "confidence": 0.8, "reasoning": "Limit buy 5% below mid price"}]}

2. MARKET order (immediate):
   {"actions": [{"action": "SELL", "product": "PALTA-OIL", "quantity": 100, "price": 0, "confidence": 0.9, "reasoning": "Quick liquidation needed"}]}

3. Multiple actions (produce + sell):
   {"actions": [
     {"action": "PRODUCE", "product": "PALTA-OIL", "quantity": 1, "price": 0, "confidence": 0.95, "reasoning": "Free production"},
     {"action": "SELL", "product": "PALTA-OIL", "quantity": 80, "price": 10.5, "confidence": 0.85, "reasoning": "LIMIT sell 5% above market"}
   ]}

4. Hold and wait:
   {"actions": [{"action": "HOLD", "product": "", "quantity": 0, "price": 0, "confidence": 0.5, "reasoning": "Waiting for better prices"}]}

Remember: Set price=0 for MARKET orders (instant), or specify price for LIMIT orders (better control). Confidence should be 0-1.`

	return ctx
}

// DeepSeekRequest represents the API request format
type DeepSeekRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
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

// getAIDecisions calls the DeepSeek API to get multiple trading decisions
func (s *DeepSeekStrategy) getAIDecisions(ctx context.Context, marketContext string) ([]AIDecision, error) {
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

		decisions, err := s.callDeepSeekAPIMultiple(ctx, request)
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
		log.Warn().
			Err(err).
			Str("strategy", s.name).
			Int("attempt", attempt+1).
			Int("maxRetries", s.maxRetries).
			Msg("DeepSeek API call failed, retrying")
	}

	return nil, fmt.Errorf("failed after %d retries: %w", s.maxRetries, lastErr)
}

// callDeepSeekAPIMultiple makes the actual API call and parses multiple actions
func (s *DeepSeekStrategy) callDeepSeekAPIMultiple(ctx context.Context, request DeepSeekRequest) ([]AIDecision, error) {
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
		if decision.Price > 0 {
			order = CreateLimitBuyOrder(decision.Product, decision.Quantity, decision.Price, "")
		} else {
			order = CreateBuyOrder(decision.Product, decision.Quantity, "")
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
		if decision.Price > 0 {
			order = CreateLimitSellOrder(decision.Product, decision.Quantity, decision.Price, "")
		} else {
			order = CreateSellOrder(decision.Product, decision.Quantity, "")
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
