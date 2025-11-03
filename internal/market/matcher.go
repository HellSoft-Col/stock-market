package market

import (
	"context"
	"fmt"

	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/rs/zerolog/log"
)

type MatchResult struct {
	Matched       bool
	BuyOrder      *domain.Order
	SellOrder     *domain.Order
	TradePrice    float64
	TradeQty      int
	FillID        string // Set during execution
	GenerateOffer bool
}

type Matcher struct {
	orderBook        domain.OrderBookRepository
	offerGenerator   *OfferGenerator
	inventoryService domain.InventoryService
	teamRepo         domain.TeamRepository
}

func NewMatcher(orderBook domain.OrderBookRepository, inventoryService domain.InventoryService, teamRepo domain.TeamRepository) *Matcher {
	return &Matcher{
		orderBook:        orderBook,
		inventoryService: inventoryService,
		teamRepo:         teamRepo,
	}
}

func (m *Matcher) SetOfferGenerator(offerGenerator *OfferGenerator) {
	m.offerGenerator = offerGenerator
}

func (m *Matcher) ProcessOrder(order *domain.Order) (*MatchResult, error) {
	switch order.Side {
	case "BUY":
		return m.processBuyOrder(order)
	case "SELL":
		return m.processSellOrder(order)
	default:
		return nil, fmt.Errorf("invalid side: %s", order.Side)
	}
}

func (m *Matcher) processBuyOrder(buyOrder *domain.Order) (*MatchResult, error) {
	// Validate buyer has sufficient balance before processing
	if m.teamRepo != nil {
		canBuy, err := m.canBuy(buyOrder)
		if err != nil {
			log.Warn().
				Str("buyTeam", buyOrder.TeamName).
				Str("product", buyOrder.Product).
				Err(err).
				Msg("Failed to check buyer balance")
			return nil, fmt.Errorf("balance validation failed: %w", err)
		}
		if !canBuy {
			logEvent := log.Info().
				Str("buyTeam", buyOrder.TeamName).
				Str("product", buyOrder.Product)
			if buyOrder.Price == nil {
				logEvent = logEvent.Str("price", "MARKET")
			}
			if buyOrder.Price != nil {
				logEvent = logEvent.Float64("price", *buyOrder.Price)
			}
			logEvent.Int("qty", buyOrder.Quantity).
				Msg("Buyer has insufficient balance")
			return nil, fmt.Errorf("insufficient balance for buy order")
		}
	}

	// First, try immediate matching with existing SELL orders
	sellOrders := m.orderBook.GetSellOrders(buyOrder.Product)

	for _, sellOrder := range sellOrders {
		if !m.canMatch(buyOrder, sellOrder) {
			continue
		}

		// Check if seller still has inventory
		if m.inventoryService != nil {
			canSell, err := m.inventoryService.CanSell(context.Background(), sellOrder.TeamName, sellOrder.Product, sellOrder.Quantity-sellOrder.FilledQty)
			if err != nil {
				log.Warn().
					Str("sellTeam", sellOrder.TeamName).
					Str("product", sellOrder.Product).
					Err(err).
					Msg("Failed to check seller inventory")
				continue
			}
			if !canSell {
				log.Debug().
					Str("sellTeam", sellOrder.TeamName).
					Str("product", sellOrder.Product).
					Int("needed", sellOrder.Quantity-sellOrder.FilledQty).
					Msg("Seller no longer has sufficient inventory")
				continue
			}
		}

		return m.createMatchResult(buyOrder, sellOrder), nil
	}

	// No immediate match - add to order book and broadcast offer to teams with inventory
	m.orderBook.AddOrder(buyOrder.Product, buyOrder.Side, buyOrder)
	m.broadcastOfferToEligibleSellers(buyOrder)

	return &MatchResult{
		Matched:       false,
		GenerateOffer: true,
	}, nil
}

func (m *Matcher) createMatchResult(buyOrder, sellOrder *domain.Order) *MatchResult {
	// Calculate trade quantity (smaller of remaining quantities)
	buyRemainingQty := buyOrder.Quantity - buyOrder.FilledQty
	sellRemainingQty := sellOrder.Quantity - sellOrder.FilledQty
	tradeQty := min(buyRemainingQty, sellRemainingQty)

	// Determine trade price
	tradePrice := m.calculateTradePrice(buyOrder, sellOrder)

	log.Debug().
		Str("buyClOrdID", buyOrder.ClOrdID).
		Str("sellClOrdID", sellOrder.ClOrdID).
		Int("tradeQty", tradeQty).
		Float64("tradePrice", tradePrice).
		Msg("Orders matched")

	return &MatchResult{
		Matched:    true,
		BuyOrder:   buyOrder,
		SellOrder:  sellOrder,
		TradePrice: tradePrice,
		TradeQty:   tradeQty,
	}
}

func (m *Matcher) calculateTradePrice(buyOrder, sellOrder *domain.Order) float64 {
	// Seller's price wins if available
	if sellOrder.Price != nil {
		return *sellOrder.Price
	}

	// Buyer's price if seller is market order
	if buyOrder.Price != nil {
		return *buyOrder.Price
	}

	// Both market orders - use mid price or default
	marketState, _ := m.getMarketPrice(buyOrder.Product)
	if marketState != nil && marketState.Mid != nil {
		return *marketState.Mid
	}

	return 10.0 // Default price
}

func (m *Matcher) generateOfferAsync(buyOrder *domain.Order) {
	if m.offerGenerator == nil {
		return
	}

	go func() {
		if err := m.offerGenerator.GenerateOffer(buyOrder); err != nil {
			log.Warn().
				Str("clOrdID", buyOrder.ClOrdID).
				Err(err).
				Msg("Failed to generate offer")
		}
	}()
}

func (m *Matcher) broadcastOfferToEligibleSellers(buyOrder *domain.Order) {
	if m.offerGenerator == nil || m.inventoryService == nil || m.teamRepo == nil {
		// Fallback to old method if services not available
		m.generateOfferAsync(buyOrder)
		return
	}

	go func() {
		ctx := context.Background()

		// Get teams with inventory for this product
		neededQty := buyOrder.Quantity - buyOrder.FilledQty
		eligibleTeams, err := m.teamRepo.GetTeamsWithInventory(ctx, buyOrder.Product, neededQty)
		if err != nil {
			log.Warn().
				Str("clOrdID", buyOrder.ClOrdID).
				Str("product", buyOrder.Product).
				Err(err).
				Msg("Failed to get eligible sellers, using fallback offer generation")
			m.generateOfferAsync(buyOrder)
			return
		}

		if len(eligibleTeams) == 0 {
			log.Info().
				Str("clOrdID", buyOrder.ClOrdID).
				Str("product", buyOrder.Product).
				Int("neededQty", neededQty).
				Msg("No teams have sufficient inventory for offer")
			return
		}

		log.Info().
			Str("clOrdID", buyOrder.ClOrdID).
			Str("product", buyOrder.Product).
			Int("neededQty", neededQty).
			Int("eligibleTeams", len(eligibleTeams)).
			Msg("Broadcasting offer to eligible sellers")

		// Generate targeted offer to eligible teams
		if err := m.offerGenerator.GenerateTargetedOffer(buyOrder, eligibleTeams); err != nil {
			log.Warn().
				Str("clOrdID", buyOrder.ClOrdID).
				Err(err).
				Msg("Failed to generate targeted offer")
		}
	}()
}

func (m *Matcher) canBuy(buyOrder *domain.Order) (bool, error) {
	team, err := m.teamRepo.GetByTeamName(context.Background(), buyOrder.TeamName)
	if err != nil {
		return false, fmt.Errorf("failed to get team: %w", err)
	}

	// Calculate required cost for this order
	var requiredCost float64
	if buyOrder.Price != nil {
		// LIMIT order: use limit price
		requiredCost = *buyOrder.Price * float64(buyOrder.Quantity)
		return team.CurrentBalance >= requiredCost, nil
	}

	// MARKET order: estimate using current market price or default
	marketState, _ := m.getMarketPrice(buyOrder.Product)
	estimatedPrice := 15.0 // Conservative default price for market orders
	if marketState != nil && marketState.Mid != nil {
		estimatedPrice = *marketState.Mid
	}
	requiredCost = estimatedPrice * float64(buyOrder.Quantity)

	return team.CurrentBalance >= requiredCost, nil
}

func (m *Matcher) processSellOrder(sellOrder *domain.Order) (*MatchResult, error) {
	// Validate seller has sufficient inventory before processing
	if m.inventoryService != nil {
		canSell, err := m.inventoryService.CanSell(context.Background(), sellOrder.TeamName, sellOrder.Product, sellOrder.Quantity)
		if err != nil {
			log.Warn().
				Str("sellTeam", sellOrder.TeamName).
				Str("product", sellOrder.Product).
				Err(err).
				Msg("Failed to check seller inventory")
			return nil, fmt.Errorf("inventory validation failed: %w", err)
		}
		if !canSell {
			log.Info().
				Str("sellTeam", sellOrder.TeamName).
				Str("product", sellOrder.Product).
				Int("qty", sellOrder.Quantity).
				Msg("Seller has insufficient inventory")
			return nil, fmt.Errorf("insufficient inventory for sell order")
		}
	}
	// Get all BUY orders for this product, sorted by price DESC (highest first)
	buyOrders := m.orderBook.GetBuyOrders(sellOrder.Product)

	for _, buyOrder := range buyOrders {
		// Check if this buy order can match
		if !m.canMatch(buyOrder, sellOrder) {
			continue
		}

		// Calculate trade quantity (smaller of remaining quantities)
		buyRemainingQty := buyOrder.Quantity - buyOrder.FilledQty
		sellRemainingQty := sellOrder.Quantity - sellOrder.FilledQty
		tradeQty := min(buyRemainingQty, sellRemainingQty)

		// Determine trade price (buyer's price wins)
		tradePrice := m.determineTradePrice(buyOrder, sellOrder)

		log.Debug().
			Str("buyClOrdID", buyOrder.ClOrdID).
			Str("sellClOrdID", sellOrder.ClOrdID).
			Int("tradeQty", tradeQty).
			Float64("tradePrice", tradePrice).
			Msg("Orders matched")

		return &MatchResult{
			Matched:    true,
			BuyOrder:   buyOrder,
			SellOrder:  sellOrder,
			TradePrice: tradePrice,
			TradeQty:   tradeQty,
		}, nil
	}

	// No match found, add to order book
	m.orderBook.AddOrder(sellOrder.Product, sellOrder.Side, sellOrder)

	return &MatchResult{
		Matched: false,
	}, nil
}

func (m *Matcher) canMatch(buyOrder, sellOrder *domain.Order) bool {
	// Same team can't trade with itself
	if buyOrder.TeamName == sellOrder.TeamName {
		return false
	}

	// LIMIT vs LIMIT: buy price must be >= sell price
	if buyOrder.Mode == "LIMIT" && sellOrder.Mode == "LIMIT" {
		return *buyOrder.LimitPrice >= *sellOrder.LimitPrice
	}

	// All other combinations match (MARKET vs LIMIT, MARKET vs MARKET)
	return true
}

func (m *Matcher) getMarketPrice(product string) (*domain.MarketState, error) {
	// This would normally call marketStateRepo, but we'll simplify for now
	// Will be properly implemented when we add market state tracking
	return nil, nil
}

/*
SIMPLIFIED ORDER MATCHING ALGORITHM DOCUMENTATION

The order matching system now follows these educational principles:

1. SELL ORDER PROCESSING:
   - First checks if there are existing BUY orders that can match immediately
   - If no immediate match, order goes to order book for future matching

2. BUY ORDER PROCESSING:
   - First checks if there are existing SELL orders that can match immediately
   - If no immediate match, broadcasts an OFFER to all teams with inventory
   - Only teams with sufficient inventory receive the offer

3. OFFER SYSTEM:
   - BUY orders create OFFERS that are sent to teams with the required product
   - Teams can approve or deny offers through the web interface
   - Offers have expiration times for urgency
   - Students can see and interact with offers in real-time

4. INVENTORY VALIDATION:
   - Teams must have sufficient inventory to receive sell offers
   - Production updates inventory automatically
   - Trades update inventory (buyer gains, seller loses)

5. EDUCATIONAL BENEFITS:
   - Students see exactly who can fulfill their orders
   - Clear cause-and-effect between production and trading opportunities
   - Interactive approval/denial teaches negotiation concepts
   - Real-time inventory tracking shows resource management

For debugging: Check team inventories, monitor offer broadcasts, and verify trades update inventories correctly.
*/

func (m *Matcher) determineTradePrice(buyOrder, sellOrder *domain.Order) float64 {
	if buyOrder.Price != nil {
		return *buyOrder.Price
	}
	if sellOrder.Price != nil {
		return *sellOrder.Price
	}

	// Both are market orders - use mid price or default
	marketState, _ := m.getMarketPrice(sellOrder.Product)
	if marketState != nil && marketState.Mid != nil {
		return *marketState.Mid
	}
	return 10.0 // Default price
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
