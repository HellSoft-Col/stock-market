package market

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/HellSoft-Col/stock-market/internal/config"
	"github.com/HellSoft-Col/stock-market/internal/domain"
)

type ActiveOffer struct {
	OfferMsg  *domain.OfferMessage
	BuyOrder  *domain.Order
	ExpiresAt time.Time
}

type OfferGenerator struct {
	config       *config.Config
	fillRepo     domain.FillRepository
	marketRepo   domain.MarketStateRepository
	broadcaster  domain.Broadcaster
	marketEngine *MarketEngine

	activeOffers map[string]*ActiveOffer
	mu           sync.RWMutex

	cleanup  *time.Ticker
	shutdown chan struct{}
	wg       sync.WaitGroup
	running  bool
}

func NewOfferGenerator(
	cfg *config.Config,
	fillRepo domain.FillRepository,
	marketRepo domain.MarketStateRepository,
	broadcaster domain.Broadcaster,
	marketEngine *MarketEngine,
) *OfferGenerator {
	return &OfferGenerator{
		config:       cfg,
		fillRepo:     fillRepo,
		marketRepo:   marketRepo,
		broadcaster:  broadcaster,
		marketEngine: marketEngine,
		activeOffers: make(map[string]*ActiveOffer),
		shutdown:     make(chan struct{}),
	}
}

func (og *OfferGenerator) Start() error {
	og.mu.Lock()
	defer og.mu.Unlock()

	if og.running {
		return nil
	}

	log.Info().Msg("Starting offer generator")
	og.running = true

	// Start cleanup ticker
	og.cleanup = time.NewTicker(100 * time.Millisecond)
	og.wg.Add(1)
	go og.cleanupLoop()

	return nil
}

func (og *OfferGenerator) Stop() error {
	og.mu.Lock()
	defer og.mu.Unlock()

	if !og.running {
		return nil
	}

	log.Info().Msg("Stopping offer generator")
	close(og.shutdown)
	og.running = false

	if og.cleanup != nil {
		og.cleanup.Stop()
	}

	og.wg.Wait()
	log.Info().Msg("Offer generator stopped")
	return nil
}

func (og *OfferGenerator) cleanupLoop() {
	defer og.wg.Done()

	for {
		select {
		case <-og.shutdown:
			return
		case <-og.cleanup.C:
			og.cleanupExpiredOffers()
		}
	}
}

func (og *OfferGenerator) cleanupExpiredOffers() {
	og.mu.Lock()
	defer og.mu.Unlock()

	now := time.Now()
	expiredOffers := make([]string, 0)

	for offerID, offer := range og.activeOffers {
		if now.After(offer.ExpiresAt) {
			expiredOffers = append(expiredOffers, offerID)
		}
	}

	for _, offerID := range expiredOffers {
		delete(og.activeOffers, offerID)
	}

	if len(expiredOffers) > 0 {
		log.Debug().
			Int("expiredOffers", len(expiredOffers)).
			Msg("Cleaned up expired offers")
	}
}

func (og *OfferGenerator) GenerateOffer(buyOrder *domain.Order) error {
	// Find teams that recently sold this product
	recentSellers, err := og.fillRepo.GetRecentSellersByProduct(
		context.Background(),
		buyOrder.Product,
		time.Now().Add(-1*time.Hour), // Last hour
	)
	if err != nil {
		log.Warn().
			Err(err).
			Str("product", buyOrder.Product).
			Msg("Failed to get recent sellers")
		return err
	}

	if len(recentSellers) == 0 {
		log.Debug().
			Str("product", buyOrder.Product).
			Msg("No recent sellers found for offer generation")
		return nil
	}

	// Calculate offer price (10% above current mid price)
	marketState, err := og.marketRepo.GetByProduct(context.Background(), buyOrder.Product)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get market state for offer price")
		return err
	}

	var offerPrice float64
	if marketState.Mid != nil {
		offerPrice = *marketState.Mid * 1.10 // 10% premium
	} else {
		offerPrice = 10.0 // Default price
	}

	// Generate unique offer ID
	offerID := fmt.Sprintf("off-%d-%s", time.Now().Unix(), uuid.New().String()[:8])

	// Determine expiration (configurable or default)
	var expiresIn *int
	var expiresAt time.Time

	if og.config.Market.OfferTimeout > 0 {
		timeoutMs := int(og.config.Market.OfferTimeout.Milliseconds())
		expiresIn = &timeoutMs
		expiresAt = time.Now().Add(og.config.Market.OfferTimeout)
	} else {
		// No expiration
		expiresAt = time.Now().Add(24 * time.Hour) // Far future
	}

	// Create offer message
	offerMsg := &domain.OfferMessage{
		Type:              "OFFER",
		OfferID:           offerID,
		Buyer:             buyOrder.TeamName,
		Product:           buyOrder.Product,
		QuantityRequested: buyOrder.Quantity,
		MaxPrice:          offerPrice,
		ExpiresIn:         expiresIn,
		Timestamp:         time.Now(),
	}

	// Store in memory
	og.mu.Lock()
	og.activeOffers[offerID] = &ActiveOffer{
		OfferMsg:  offerMsg,
		BuyOrder:  buyOrder,
		ExpiresAt: expiresAt,
	}
	og.mu.Unlock()

	// Send to potential sellers
	sentCount := 0
	for _, sellerTeam := range recentSellers {
		if sellerTeam != buyOrder.TeamName { // Don't send to self
			err := og.broadcaster.SendToClient(sellerTeam, offerMsg)
			if err != nil {
				log.Debug().
					Str("team", sellerTeam).
					Str("offerID", offerID).
					Err(err).
					Msg("Failed to send offer to team")
			} else {
				sentCount++
				log.Debug().
					Str("team", sellerTeam).
					Str("offerID", offerID).
					Str("product", buyOrder.Product).
					Int("qty", buyOrder.Quantity).
					Float64("maxPrice", offerPrice).
					Msg("Offer sent to potential seller")
			}
		}
	}

	log.Info().
		Str("offerID", offerID).
		Str("product", buyOrder.Product).
		Int("potentialSellers", len(recentSellers)).
		Int("sentCount", sentCount).
		Msg("Offer generated and sent")

	return nil
}

func (og *OfferGenerator) GenerateTargetedOffer(buyOrder *domain.Order, eligibleTeams []*domain.Team) error {
	// Handle debug modes
	if buyOrder.DebugMode == "AUTO_ACCEPT" {
		return og.handleAutoAcceptOrder(buyOrder)
	}

	// Calculate offer price (10% above current mid price)
	marketState, err := og.marketRepo.GetByProduct(context.Background(), buyOrder.Product)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get market state for offer price")
		return err
	}

	var offerPrice float64
	if marketState.Mid != nil {
		offerPrice = *marketState.Mid * 1.10 // 10% premium
	} else {
		offerPrice = 10.0 // Default price
	}

	// Generate unique offer ID
	offerID := fmt.Sprintf("off-%d-%s", time.Now().Unix(), uuid.New().String()[:8])

	// Determine expiration (configurable or default)
	var expiresIn *int
	var expiresAt time.Time

	if og.config.Market.OfferTimeout > 0 {
		timeoutMs := int(og.config.Market.OfferTimeout.Milliseconds())
		expiresIn = &timeoutMs
		expiresAt = time.Now().Add(og.config.Market.OfferTimeout)
	} else {
		// No expiration
		expiresAt = time.Now().Add(24 * time.Hour) // Far future
	}

	// Create offer message
	offerMsg := &domain.OfferMessage{
		Type:              "OFFER",
		OfferID:           offerID,
		Buyer:             buyOrder.TeamName,
		Product:           buyOrder.Product,
		QuantityRequested: buyOrder.Quantity - buyOrder.FilledQty,
		MaxPrice:          offerPrice,
		ExpiresIn:         expiresIn,
		Timestamp:         time.Now(),
	}

	// Store in memory
	og.mu.Lock()
	og.activeOffers[offerID] = &ActiveOffer{
		OfferMsg:  offerMsg,
		BuyOrder:  buyOrder,
		ExpiresAt: expiresAt,
	}
	og.mu.Unlock()

	// Send to eligible teams only
	sentCount := 0
	for _, team := range eligibleTeams {
		if team.TeamName != buyOrder.TeamName { // Don't send to self
			err := og.broadcaster.SendToClient(team.TeamName, offerMsg)
			if err != nil {
				log.Debug().
					Str("team", team.TeamName).
					Str("offerID", offerID).
					Err(err).
					Msg("Failed to send targeted offer to team")
			} else {
				sentCount++
				log.Debug().
					Str("team", team.TeamName).
					Str("offerID", offerID).
					Str("product", buyOrder.Product).
					Int("teamInventory", team.Inventory[buyOrder.Product]).
					Int("qty", buyOrder.Quantity-buyOrder.FilledQty).
					Float64("maxPrice", offerPrice).
					Msg("Targeted offer sent to team with inventory")
			}
		}
	}

	log.Info().
		Str("offerID", offerID).
		Str("product", buyOrder.Product).
		Int("eligibleTeams", len(eligibleTeams)).
		Int("sentCount", sentCount).
		Msg("Targeted offer generated and sent to teams with inventory")

	return nil
}

func (og *OfferGenerator) HandleAcceptOffer(acceptMsg *domain.AcceptOfferMessage, acceptorTeam string) error {
	og.mu.RLock()
	offer, exists := og.activeOffers[acceptMsg.OfferID]
	og.mu.RUnlock()

	if !exists {
		return fmt.Errorf("offer not found or expired: %s", acceptMsg.OfferID)
	}

	// Check expiration
	if time.Now().After(offer.ExpiresAt) {
		og.mu.Lock()
		delete(og.activeOffers, acceptMsg.OfferID)
		og.mu.Unlock()
		return fmt.Errorf("offer expired: %s", acceptMsg.OfferID)
	}

	if !acceptMsg.Accept {
		log.Debug().
			Str("offerID", acceptMsg.OfferID).
			Str("team", acceptorTeam).
			Msg("Offer declined")
		return nil
	}

	// Validate offer acceptance
	if acceptMsg.QuantityOffered <= 0 || acceptMsg.PriceOffered <= 0 {
		return fmt.Errorf("invalid quantity or price in offer acceptance")
	}

	// Create virtual SELL order from acceptor
	virtualSellOrder := &domain.Order{
		ClOrdID:   fmt.Sprintf("VIRT-SELL-%d-%s", time.Now().UnixNano(), uuid.New().String()[:8]),
		TeamName:  acceptorTeam,
		Side:      "SELL",
		Mode:      "MARKET",
		Product:   offer.BuyOrder.Product,
		Quantity:  acceptMsg.QuantityOffered,
		Price:     &acceptMsg.PriceOffered,
		Message:   fmt.Sprintf("Accepting offer %s", acceptMsg.OfferID),
		CreatedAt: time.Now(),
		Status:    "PENDING",
		FilledQty: 0,
	}

	// Execute immediate match
	err := og.executeOfferMatch(offer.BuyOrder, virtualSellOrder)
	if err != nil {
		return fmt.Errorf("failed to execute offer acceptance: %w", err)
	}

	// Clean up
	og.mu.Lock()
	delete(og.activeOffers, acceptMsg.OfferID)
	og.mu.Unlock()

	log.Info().
		Str("offerID", acceptMsg.OfferID).
		Str("buyer", offer.BuyOrder.TeamName).
		Str("seller", acceptorTeam).
		Str("product", offer.BuyOrder.Product).
		Int("qty", acceptMsg.QuantityOffered).
		Float64("price", acceptMsg.PriceOffered).
		Msg("Offer accepted and trade executed")

	return nil
}

func (og *OfferGenerator) executeOfferMatch(buyOrder, sellOrder *domain.Order) error {
	// Calculate trade quantity and price
	tradeQty := minInt(buyOrder.Quantity-buyOrder.FilledQty, sellOrder.Quantity)
	tradePrice := *sellOrder.Price // Seller's price wins

	result := &MatchResult{
		Matched:    true,
		BuyOrder:   buyOrder,
		SellOrder:  sellOrder,
		TradePrice: tradePrice,
		TradeQty:   tradeQty,
	}

	// Execute the trade using market engine's transaction logic
	return og.marketEngine.executeTradeTransaction(buyOrder, sellOrder, result)
}

func (og *OfferGenerator) handleAutoAcceptOrder(buyOrder *domain.Order) error {
	// For AUTO_ACCEPT debug mode, create a virtual sell order with server-generated price
	marketState, err := og.marketRepo.GetByProduct(context.Background(), buyOrder.Product)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get market state for auto-accept")
		return err
	}

	var price float64
	if marketState.Mid != nil {
		price = *marketState.Mid
	} else {
		price = 10.0 // Default price
	}

	virtualSellOrder := &domain.Order{
		ClOrdID:   fmt.Sprintf("AUTO-SELL-%d-%s", time.Now().UnixNano(), uuid.New().String()[:8]),
		TeamName:  "SERVER",
		Side:      "SELL",
		Mode:      "LIMIT",
		Product:   buyOrder.Product,
		Quantity:  buyOrder.Quantity - buyOrder.FilledQty,
		Price:     &price,
		Message:   "Auto-accept debug order",
		CreatedAt: time.Now(),
		Status:    "PENDING",
		FilledQty: 0,
	}

	// Execute immediate match
	err = og.executeOfferMatch(buyOrder, virtualSellOrder)
	if err != nil {
		return fmt.Errorf("failed to execute auto-accept order: %w", err)
	}

	log.Info().
		Str("buyClOrdID", buyOrder.ClOrdID).
		Str("product", buyOrder.Product).
		Int("qty", virtualSellOrder.Quantity).
		Float64("price", price).
		Msg("Auto-accept debug order executed")

	return nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
