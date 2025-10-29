package market

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/yourusername/avocado-exchange-server/internal/config"
	"github.com/yourusername/avocado-exchange-server/internal/domain"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderCommand struct {
	Order  *domain.Order
	Client domain.ClientConnection // Will be nil until Phase 7
}

type MarketEngine struct {
	config           *config.Config
	db               domain.Database
	orderRepo        domain.OrderRepository
	fillRepo         domain.FillRepository
	marketRepo       domain.MarketStateRepository
	orderBook        domain.OrderBookRepository
	matcher          *Matcher
	broadcaster      domain.Broadcaster
	OfferGenerator   *OfferGenerator
	inventoryService domain.InventoryService
	teamRepo         domain.TeamRepository

	orderChan chan OrderCommand
	shutdown  chan struct{}
	wg        sync.WaitGroup
	running   bool
	mu        sync.RWMutex
}

func NewMarketEngine(
	cfg *config.Config,
	db domain.Database,
	orderRepo domain.OrderRepository,
	fillRepo domain.FillRepository,
	marketRepo domain.MarketStateRepository,
	orderBook domain.OrderBookRepository,
	broadcaster domain.Broadcaster,
	inventoryService domain.InventoryService,
	teamRepo domain.TeamRepository,
) *MarketEngine {
	return &MarketEngine{
		config:           cfg,
		db:               db,
		orderRepo:        orderRepo,
		fillRepo:         fillRepo,
		marketRepo:       marketRepo,
		orderBook:        orderBook,
		broadcaster:      broadcaster,
		inventoryService: inventoryService,
		teamRepo:         teamRepo,
		matcher:          NewMatcher(orderBook, inventoryService, teamRepo),
		orderChan:        make(chan OrderCommand, 1000), // Buffered channel
		shutdown:         make(chan struct{}),
	}
}

func (m *MarketEngine) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("market engine already running")
	}

	log.Info().Msg("Starting market engine")

	// Create and start offer generator
	m.OfferGenerator = NewOfferGenerator(m.config, m.fillRepo, m.marketRepo, m.broadcaster, m)
	if err := m.OfferGenerator.Start(); err != nil {
		return fmt.Errorf("failed to start offer generator: %w", err)
	}

	// Set offer generator in matcher
	m.matcher.SetOfferGenerator(m.OfferGenerator)

	// Load pending orders from database into order book
	if err := m.orderBook.LoadFromDatabase(ctx, m.orderRepo); err != nil {
		return fmt.Errorf("failed to load order book from database: %w", err)
	}

	m.running = true

	// Start market engine goroutine
	m.wg.Add(1)
	go m.run()

	log.Info().Msg("Market engine started")
	return nil
}

func (m *MarketEngine) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	log.Info().Msg("Stopping market engine")

	close(m.shutdown)
	m.running = false

	// Stop offer generator
	if m.OfferGenerator != nil {
		if err := m.OfferGenerator.Stop(); err != nil {
			log.Warn().Err(err).Msg("Error stopping offer generator")
		}
	}

	return nil
}

func (m *MarketEngine) ProcessOrder(order *domain.Order, client domain.ClientConnection) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.running {
		log.Warn().
			Str("clOrdID", order.ClOrdID).
			Msg("Market engine not running, order ignored")
		return
	}

	// Send order to processing channel
	select {
	case m.orderChan <- OrderCommand{Order: order, Client: client}:
		log.Debug().
			Str("clOrdID", order.ClOrdID).
			Str("teamName", order.TeamName).
			Msg("Order queued for processing")
	default:
		log.Error().
			Str("clOrdID", order.ClOrdID).
			Msg("Order channel full, order dropped")
	}
}

func (m *MarketEngine) run() {
	defer m.wg.Done()

	log.Info().Msg("Market engine processing loop started")

	for {
		select {
		case <-m.shutdown:
			log.Info().Msg("Market engine processing loop stopping")
			return

		case cmd := <-m.orderChan:
			m.processOrderCommand(cmd)
		}
	}
}

func (m *MarketEngine) processOrderCommand(cmd OrderCommand) {
	order := cmd.Order

	log.Debug().
		Str("clOrdID", order.ClOrdID).
		Str("teamName", order.TeamName).
		Str("side", order.Side).
		Str("product", order.Product).
		Int("qty", order.Quantity).
		Msg("Processing order")

	// Check if order has expired
	if order.ExpiresAt != nil && time.Now().After(*order.ExpiresAt) {
		log.Info().
			Str("clOrdID", order.ClOrdID).
			Msg("Order expired, cancelling")

		if err := m.orderRepo.Cancel(context.Background(), order.ClOrdID); err != nil {
			log.Error().Str("clOrdID", order.ClOrdID).Err(err).Msg("Failed to cancel expired order")
		}
		return
	}

	// Try to match the order
	matchResult, err := m.matcher.ProcessOrder(order)
	if err != nil {
		log.Error().
			Str("clOrdID", order.ClOrdID).
			Err(err).
			Msg("Error matching order")
		return
	}

	if matchResult.Matched {
		// Execute the trade
		if err := m.executeTrade(matchResult); err != nil {
			log.Error().
				Str("buyerClOrdID", matchResult.BuyOrder.ClOrdID).
				Str("sellerClOrdID", matchResult.SellOrder.ClOrdID).
				Err(err).
				Msg("Failed to execute trade")
		}
	} else {
		// No match found, order stays in book
		log.Debug().
			Str("clOrdID", order.ClOrdID).
			Msg("No match found, order added to book")
	}
}

func (m *MarketEngine) executeTrade(result *MatchResult) error {
	buyOrder := result.BuyOrder
	sellOrder := result.SellOrder

	// Determine if this is a full or partial fill
	fillQty := result.TradeQty
	isFullFillBuy := fillQty == (buyOrder.Quantity - buyOrder.FilledQty)
	isFullFillSell := fillQty == (sellOrder.Quantity - sellOrder.FilledQty)

	// Start transaction with retry logic
	var err error
	for attempt := 0; attempt < m.config.Market.TransactionRetries; attempt++ {
		err = m.executeTradeTransaction(buyOrder, sellOrder, result)
		if err == nil {
			break
		}

		log.Warn().
			Int("attempt", attempt+1).
			Int("maxRetries", m.config.Market.TransactionRetries).
			Str("buyerClOrdID", buyOrder.ClOrdID).
			Str("sellerClOrdID", sellOrder.ClOrdID).
			Err(err).
			Msg("Trade transaction failed, retrying")

		time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
	}

	if err != nil {
		log.Error().
			Str("buyerClOrdID", buyOrder.ClOrdID).
			Str("sellerClOrdID", sellOrder.ClOrdID).
			Err(err).
			Msg("Trade transaction failed after all retries")
		return err
	}

	// Remove filled orders from order book
	if isFullFillBuy {
		m.orderBook.RemoveOrder(buyOrder.Product, buyOrder.Side, buyOrder.ClOrdID)
	}
	if isFullFillSell {
		m.orderBook.RemoveOrder(sellOrder.Product, sellOrder.Side, sellOrder.ClOrdID)
	}

	// Update market state
	if err := m.marketRepo.UpdateLastTrade(context.Background(), buyOrder.Product, result.TradePrice, fillQty); err != nil {
		log.Warn().Err(err).Msg("Failed to update market state")
	}

	// Broadcast market state update
	m.broadcastMarketStateUpdate(buyOrder.Product)

	log.Info().
		Str("fillID", result.FillID).
		Str("buyer", buyOrder.TeamName).
		Str("seller", sellOrder.TeamName).
		Str("product", buyOrder.Product).
		Int("qty", fillQty).
		Float64("price", result.TradePrice).
		Bool("fullFillBuy", isFullFillBuy).
		Bool("fullFillSell", isFullFillSell).
		Msg("Trade executed successfully")

	// Broadcast FILL messages to both parties
	m.broadcastFill(result, buyOrder, sellOrder)

	// Broadcast balance and inventory updates
	m.broadcastBalanceUpdate(buyOrder.TeamName)
	m.broadcastBalanceUpdate(sellOrder.TeamName)
	m.broadcastInventoryUpdate(buyOrder.TeamName)
	m.broadcastInventoryUpdate(sellOrder.TeamName)

	return nil
}

func (m *MarketEngine) executeTradeTransaction(buyOrder, sellOrder *domain.Order, result *MatchResult) error {
	_, err := m.db.WithTransaction(context.Background(), func(sc mongo.SessionContext) (any, error) {
		// Generate unique fill ID
		fillID := fmt.Sprintf("FILL-%d-%s", time.Now().Unix(), uuid.New().String()[:8])
		result.FillID = fillID

		fillQty := result.TradeQty

		// Update buy order
		if fillQty == (buyOrder.Quantity - buyOrder.FilledQty) {
			// Full fill
			err := m.orderRepo.UpdateToFilled(context.Background(), sc, buyOrder.ClOrdID, fillID, fillQty)
			if err != nil {
				return nil, fmt.Errorf("failed to update buy order to filled: %w", err)
			}
		} else {
			// Partial fill
			err := m.orderRepo.UpdateToPartiallyFilled(context.Background(), sc, buyOrder.ClOrdID, fillID, fillQty)
			if err != nil {
				return nil, fmt.Errorf("failed to update buy order to partially filled: %w", err)
			}
		}

		// Update sell order
		if fillQty == (sellOrder.Quantity - sellOrder.FilledQty) {
			// Full fill
			err := m.orderRepo.UpdateToFilled(context.Background(), sc, sellOrder.ClOrdID, fillID, fillQty)
			if err != nil {
				return nil, fmt.Errorf("failed to update sell order to filled: %w", err)
			}
		} else {
			// Partial fill
			err := m.orderRepo.UpdateToPartiallyFilled(context.Background(), sc, sellOrder.ClOrdID, fillID, fillQty)
			if err != nil {
				return nil, fmt.Errorf("failed to update sell order to partially filled: %w", err)
			}
		}

		// Create fill record
		fill := &domain.Fill{
			FillID:        fillID,
			BuyerClOrdID:  buyOrder.ClOrdID,
			SellerClOrdID: sellOrder.ClOrdID,
			Buyer:         buyOrder.TeamName,
			Seller:        sellOrder.TeamName,
			Product:       buyOrder.Product,
			Quantity:      fillQty,
			Price:         result.TradePrice,
			BuyerMessage:  buyOrder.Message,
			SellerMessage: sellOrder.Message,
		}

		err := m.fillRepo.Create(context.Background(), sc, fill)
		if err != nil {
			return nil, fmt.Errorf("failed to create fill record: %w", err)
		}

		// Update inventories if service is available
		if m.inventoryService != nil {
			// Buyer gains inventory
			if err := m.inventoryService.UpdateInventory(context.Background(), buyOrder.TeamName, buyOrder.Product, fillQty, "TRADE_BUY", buyOrder.ClOrdID, fillID); err != nil {
				log.Warn().
					Str("fillID", fillID).
					Str("buyer", buyOrder.TeamName).
					Err(err).
					Msg("Failed to update buyer inventory, continuing with trade")
			}

			// Seller loses inventory
			if err := m.inventoryService.UpdateInventory(context.Background(), sellOrder.TeamName, sellOrder.Product, -fillQty, "TRADE_SELL", sellOrder.ClOrdID, fillID); err != nil {
				log.Warn().
					Str("fillID", fillID).
					Str("seller", sellOrder.TeamName).
					Err(err).
					Msg("Failed to update seller inventory, continuing with trade")
			}
		}

		// Update team balances
		totalCost := result.TradePrice * float64(fillQty)

		// Buyer loses balance
		if err := m.teamRepo.UpdateBalanceBy(context.Background(), buyOrder.TeamName, -totalCost); err != nil {
			return nil, fmt.Errorf("failed to update buyer balance: %w", err)
		}

		// Seller gains balance
		if err := m.teamRepo.UpdateBalanceBy(context.Background(), sellOrder.TeamName, totalCost); err != nil {
			return nil, fmt.Errorf("failed to update seller balance: %w", err)
		}

		return fill, nil
	})
	return err
}

func (m *MarketEngine) broadcastFill(result *MatchResult, buyOrder, sellOrder *domain.Order) {
	fillQty := result.TradeQty
	fillPrice := result.TradePrice
	serverTime := time.Now().Format(time.RFC3339)

	// Calculate remaining quantities for partial fills
	buyRemainingQty := (buyOrder.Quantity - buyOrder.FilledQty) - fillQty
	sellRemainingQty := (sellOrder.Quantity - sellOrder.FilledQty) - fillQty

	// Create FILL message for buyer
	buyerFillMsg := &domain.FillMessage{
		Type:                "FILL",
		ClOrdID:             buyOrder.ClOrdID,
		FillQty:             fillQty,
		FillPrice:           fillPrice,
		Side:                "BUY",
		Product:             buyOrder.Product,
		Counterparty:        sellOrder.TeamName,
		CounterpartyMessage: sellOrder.Message,
		ServerTime:          serverTime,
		RemainingQty:        buyRemainingQty,
		TotalQty:            buyOrder.Quantity,
	}

	// Create FILL message for seller
	sellerFillMsg := &domain.FillMessage{
		Type:                "FILL",
		ClOrdID:             sellOrder.ClOrdID,
		FillQty:             fillQty,
		FillPrice:           fillPrice,
		Side:                "SELL",
		Product:             sellOrder.Product,
		Counterparty:        buyOrder.TeamName,
		CounterpartyMessage: buyOrder.Message,
		ServerTime:          serverTime,
		RemainingQty:        sellRemainingQty,
		TotalQty:            sellOrder.Quantity,
	}

	// Send FILL messages
	if err := m.broadcaster.SendToClient(buyOrder.TeamName, buyerFillMsg); err != nil {
		log.Warn().
			Str("teamName", buyOrder.TeamName).
			Str("fillID", result.FillID).
			Err(err).
			Msg("Failed to send FILL to buyer")
	} else {
		log.Debug().
			Str("teamName", buyOrder.TeamName).
			Str("fillID", result.FillID).
			Msg("FILL sent to buyer")
	}

	if err := m.broadcaster.SendToClient(sellOrder.TeamName, sellerFillMsg); err != nil {
		log.Warn().
			Str("teamName", sellOrder.TeamName).
			Str("fillID", result.FillID).
			Err(err).
			Msg("Failed to send FILL to seller")
	} else {
		log.Debug().
			Str("teamName", sellOrder.TeamName).
			Str("fillID", result.FillID).
			Msg("FILL sent to seller")
	}
}

func (m *MarketEngine) broadcastMarketStateUpdate(product string) {
	if m.broadcaster == nil {
		return
	}

	// Get current market state
	marketState, err := m.marketRepo.GetByProduct(context.Background(), product)
	if err != nil {
		log.Warn().Str("product", product).Err(err).Msg("Failed to get market state for broadcast")
		return
	}

	// Get best bid and ask from order book
	bestBid := m.orderBook.GetBestBid(product)
	bestAsk := m.orderBook.GetBestAsk(product)

	// Update market state with current order book data
	var bestBidPrice, bestAskPrice, midPrice *float64
	if bestBid != nil && bestBid.Price != nil {
		bestBidPrice = bestBid.Price
	}
	if bestAsk != nil && bestAsk.Price != nil {
		bestAskPrice = bestAsk.Price
	}
	if bestBidPrice != nil && bestAskPrice != nil {
		mid := (*bestBidPrice + *bestAskPrice) / 2.0
		midPrice = &mid
	}

	// Create ticker message
	tickerMsg := &domain.TickerMessage{
		Type:       "TICKER",
		Product:    product,
		BestBid:    bestBidPrice,
		BestAsk:    bestAskPrice,
		Mid:        midPrice,
		Volume24h:  marketState.Volume24h,
		ServerTime: time.Now().Format(time.RFC3339),
	}

	// Broadcast to all clients
	if err := m.broadcaster.BroadcastToAll(tickerMsg); err != nil {
		log.Warn().Str("product", product).Err(err).Msg("Failed to broadcast ticker update")
	}

	// Update market state repository with new best prices
	if err := m.marketRepo.UpdateBestPrices(context.Background(), product, bestBidPrice, bestAskPrice); err != nil {
		log.Warn().Str("product", product).Err(err).Msg("Failed to update best prices in market state")
	}

	log.Debug().
		Str("product", product).
		Interface("bestBid", bestBidPrice).
		Interface("bestAsk", bestAskPrice).
		Interface("mid", midPrice).
		Msg("Market state broadcasted")
}

func (m *MarketEngine) broadcastBalanceUpdate(teamName string) {
	if m.broadcaster == nil || m.teamRepo == nil {
		return
	}

	// Get current team balance
	team, err := m.teamRepo.GetByTeamName(context.Background(), teamName)
	if err != nil {
		log.Warn().Str("teamName", teamName).Err(err).Msg("Failed to get team for balance broadcast")
		return
	}

	// Create balance update message
	balanceMsg := &domain.BalanceUpdateMessage{
		Type:       "BALANCE_UPDATE",
		Balance:    team.CurrentBalance,
		ServerTime: time.Now().Format(time.RFC3339),
	}

	// Send to specific team
	if err := m.broadcaster.SendToClient(teamName, balanceMsg); err != nil {
		log.Warn().Str("teamName", teamName).Err(err).Msg("Failed to broadcast balance update")
	}

	log.Debug().
		Str("teamName", teamName).
		Float64("balance", team.CurrentBalance).
		Msg("Balance update broadcasted")
}

func (m *MarketEngine) broadcastInventoryUpdate(teamName string) {
	if m.broadcaster == nil || m.teamRepo == nil {
		return
	}

	// Get current team inventory
	team, err := m.teamRepo.GetByTeamName(context.Background(), teamName)
	if err != nil {
		log.Warn().Str("teamName", teamName).Err(err).Msg("Failed to get team for inventory broadcast")
		return
	}

	inventory := team.Inventory
	if inventory == nil {
		inventory = make(map[string]int)
	}

	// Create inventory update message
	inventoryMsg := &domain.InventoryUpdateMessage{
		Type:       "INVENTORY_UPDATE",
		Inventory:  inventory,
		ServerTime: time.Now().Format(time.RFC3339),
	}

	// Send to specific team
	if err := m.broadcaster.SendToClient(teamName, inventoryMsg); err != nil {
		log.Warn().Str("teamName", teamName).Err(err).Msg("Failed to broadcast inventory update")
	}

	log.Debug().
		Str("teamName", teamName).
		Interface("inventory", inventory).
		Msg("Inventory update broadcasted")
}

var _ domain.MarketService = (*MarketEngine)(nil)
