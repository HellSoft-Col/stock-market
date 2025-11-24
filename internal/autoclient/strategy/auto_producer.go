package strategy

import (
	"context"
	"fmt"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/autoclient/market"
	"github.com/HellSoft-Col/stock-market/internal/autoclient/production"
	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/rs/zerolog/log"
)

// AutoProducerStrategy implements automated production and trading
// Strategy:
// 1. Try premium production first (if have ingredients) → hold for good price
// 2. Fallback to basic production (free) → sell immediately for capital
// 3. Use capital to buy ingredients from students
// 4. Repeat cycle to maximize profit
type AutoProducerStrategy struct {
	name string

	// Configuration
	productionInterval time.Duration
	basicProduct       string
	premiumProduct     string
	autoSellBasic      bool
	minProfitMargin    float64

	// Production system
	calculator    *production.ProductionCalculator
	recipeManager *production.RecipeManager
	role          *production.Role

	// State
	lastProduction time.Time
	health         StrategyHealth
}

// NewAutoProducerStrategy creates a new auto-producer strategy
func NewAutoProducerStrategy(name string) *AutoProducerStrategy {
	return &AutoProducerStrategy{
		name:       name,
		calculator: production.NewProductionCalculator(),
		health: StrategyHealth{
			Status:     HealthStatusHealthy,
			LastUpdate: time.Now(),
		},
	}
}

// Name returns the strategy name
func (s *AutoProducerStrategy) Name() string {
	return s.name
}

// Initialize initializes the strategy with configuration
func (s *AutoProducerStrategy) Initialize(config map[string]interface{}) error {
	s.productionInterval = GetConfigDuration(config, "productionInterval", 60*time.Second)
	s.basicProduct = GetConfigString(config, "basicProduct", "")
	s.premiumProduct = GetConfigString(config, "premiumProduct", "")
	s.autoSellBasic = GetConfigBool(config, "autoSellBasic", true)
	s.minProfitMargin = GetConfigFloat(config, "minProfitMargin", 0.05)

	if s.basicProduct == "" {
		return fmt.Errorf("basicProduct is required")
	}

	if s.premiumProduct == "" {
		return fmt.Errorf("premiumProduct is required")
	}

	log.Info().
		Str("strategy", s.name).
		Str("basicProduct", s.basicProduct).
		Str("premiumProduct", s.premiumProduct).
		Dur("interval", s.productionInterval).
		Msg("Auto-producer strategy initialized")

	return nil
}

// OnLogin is called when connected and logged in
func (s *AutoProducerStrategy) OnLogin(ctx context.Context, loginInfo *domain.LoginOKMessage) error {
	// Initialize production system with role and recipes
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
		Int("branches", s.role.Branches).
		Int("maxDepth", s.role.MaxDepth).
		Msg("Production system initialized")

	return nil
}

// OnTicker is called when ticker updates arrive
func (s *AutoProducerStrategy) OnTicker(ctx context.Context, ticker *domain.TickerMessage) error {
	// No action needed for auto-producer on ticker updates
	return nil
}

// OnFill is called when a fill notification arrives
func (s *AutoProducerStrategy) OnFill(ctx context.Context, fill *domain.FillMessage) error {
	log.Info().
		Str("strategy", s.name).
		Str("side", fill.Side).
		Str("product", fill.Product).
		Int("qty", fill.FillQty).
		Float64("price", fill.FillPrice).
		Msg("Fill received")
	return nil
}

// OnOffer is called when an offer request arrives
func (s *AutoProducerStrategy) OnOffer(ctx context.Context, offer *domain.OfferMessage) (*OfferResponse, error) {
	// Auto-producer doesn't respond to offers (focuses on production)
	return &OfferResponse{
		Accept: false,
		Reason: "Auto-producer strategy does not accept offers",
	}, nil
}

// OnInventoryUpdate is called when inventory changes
func (s *AutoProducerStrategy) OnInventoryUpdate(ctx context.Context, inventory map[string]int) error {
	log.Debug().
		Str("strategy", s.name).
		Interface("inventory", inventory).
		Msg("Inventory updated")
	return nil
}

// OnBalanceUpdate is called when balance changes
func (s *AutoProducerStrategy) OnBalanceUpdate(ctx context.Context, balance float64) error {
	log.Debug().
		Str("strategy", s.name).
		Float64("balance", balance).
		Msg("Balance updated")
	return nil
}

// OnOrderBookUpdate is called when orderbook updates arrive
func (s *AutoProducerStrategy) OnOrderBookUpdate(ctx context.Context, orderbook *domain.OrderBookUpdateMessage) error {
	// No action needed for auto-producer
	return nil
}

// Execute is called periodically to generate trading actions
func (s *AutoProducerStrategy) Execute(ctx context.Context, state *market.MarketState) ([]*Action, error) {
	// Check if it's time to produce
	if time.Since(s.lastProduction) < s.productionInterval {
		return nil, nil
	}

	actions := []*Action{}

	// Get current inventory snapshot
	inventory := state.GetSnapshot().Inventory

	// Strategy 1: Try premium production first (more profitable)
	if s.recipeManager.CanProducePremium(s.premiumProduct, inventory) {
		log.Info().
			Str("strategy", s.name).
			Str("product", s.premiumProduct).
			Msg("Attempting premium production")

		// Calculate premium units
		baseUnits := s.calculator.CalculateUnits(s.role)
		recipe, _ := s.recipeManager.GetRecipe(s.premiumProduct)
		premiumUnits := s.calculator.ApplyPremiumBonus(baseUnits, recipe.PremiumBonus)

		// Consume ingredients (locally, server will validate)
		tempInventory := make(map[string]int)
		for k, v := range inventory {
			tempInventory[k] = v
		}
		if err := s.recipeManager.ConsumeIngredients(s.premiumProduct, tempInventory); err != nil {
			log.Error().Err(err).Msg("Failed to consume ingredients")
		} else {
			// Update local state
			state.UpdateInventory(tempInventory)
			state.AddInventory(s.premiumProduct, premiumUnits)

			// Notify server about production
			actions = append(actions, &Action{
				Type:       ActionTypeProduction,
				Production: CreateProduction(s.premiumProduct, premiumUnits),
			})

			log.Info().
				Str("strategy", s.name).
				Str("product", s.premiumProduct).
				Int("units", premiumUnits).
				Interface("consumed", recipe.Ingredients).
				Msg("Premium production completed")

			// Optionally sell premium if price is excellent
			if price := state.GetPrice(s.premiumProduct); price != nil {
				recipe, _ := s.recipeManager.GetRecipe(s.premiumProduct)
				if recipe.Ingredients != nil {
					// Calculate cost of ingredients
					ingredientCost := 0.0
					for ing, qty := range recipe.Ingredients {
						if ingPrice := state.GetPrice(ing); ingPrice != nil {
							ingredientCost += *ingPrice * float64(qty)
						}
					}

					// Sell if profit margin is good
					profitPerUnit := *price - (ingredientCost / float64(premiumUnits))
					margin := profitPerUnit / *price
					if margin >= s.minProfitMargin {
						actions = append(actions, &Action{
							Type:  ActionTypeOrder,
							Order: CreateSellOrder(s.premiumProduct, premiumUnits, "Premium auto-sell"),
						})

						log.Info().
							Str("strategy", s.name).
							Str("product", s.premiumProduct).
							Int("units", premiumUnits).
							Float64("price", *price).
							Float64("margin", margin).
							Msg("Selling premium production")
					}
				}
			}

			s.lastProduction = time.Now()
			return actions, nil
		}
	}

	// Strategy 2: Produce basic and sell immediately for capital
	log.Info().
		Str("strategy", s.name).
		Str("product", s.basicProduct).
		Msg("Attempting basic production")

	baseUnits := s.calculator.CalculateUnits(s.role)

	// Update local inventory
	state.AddInventory(s.basicProduct, baseUnits)

	// Notify server about production
	actions = append(actions, &Action{
		Type:       ActionTypeProduction,
		Production: CreateProduction(s.basicProduct, baseUnits),
	})

	log.Info().
		Str("strategy", s.name).
		Str("product", s.basicProduct).
		Int("units", baseUnits).
		Msg("Basic production completed")

	// Auto-sell basic to generate capital for buying ingredients
	if s.autoSellBasic && baseUnits > 0 {
		actions = append(actions, &Action{
			Type:  ActionTypeOrder,
			Order: CreateSellOrder(s.basicProduct, baseUnits, "Basic auto-sell for capital"),
		})

		log.Info().
			Str("strategy", s.name).
			Str("product", s.basicProduct).
			Int("units", baseUnits).
			Msg("Selling basic production immediately")
	}

	s.lastProduction = time.Now()
	s.health.LastUpdate = time.Now()

	return actions, nil
}

// Health returns the strategy's current health status
func (s *AutoProducerStrategy) Health() StrategyHealth {
	return s.health
}
