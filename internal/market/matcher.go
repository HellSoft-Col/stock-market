package market

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/yourusername/avocado-exchange-server/internal/domain"
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
	orderBook      domain.OrderBookRepository
	offerGenerator *OfferGenerator
}

func NewMatcher(orderBook domain.OrderBookRepository) *Matcher {
	return &Matcher{
		orderBook: orderBook,
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
	// Get all SELL orders for this product, sorted by price ASC (lowest first)
	sellOrders := m.orderBook.GetSellOrders(buyOrder.Product)

	// Try to find a match
	for _, sellOrder := range sellOrders {
		if !m.canMatch(buyOrder, sellOrder) {
			continue
		}

		return m.createMatchResult(buyOrder, sellOrder), nil
	}

	// No match found - add to order book and generate offer
	m.orderBook.AddOrder(buyOrder.Product, buyOrder.Side, buyOrder)
	m.generateOfferAsync(buyOrder)

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

func (m *Matcher) processSellOrder(sellOrder *domain.Order) (*MatchResult, error) {
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
		var tradePrice float64
		if buyOrder.Price != nil {
			tradePrice = *buyOrder.Price
		} else if sellOrder.Price != nil {
			tradePrice = *sellOrder.Price
		} else {
			// Both are market orders - use mid price or default
			marketState, _ := m.getMarketPrice(sellOrder.Product)
			if marketState != nil && marketState.Mid != nil {
				tradePrice = *marketState.Mid
			} else {
				tradePrice = 10.0 // Default price
			}
		}

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

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
