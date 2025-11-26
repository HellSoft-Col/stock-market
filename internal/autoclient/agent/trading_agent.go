package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/autoclient/market"
	"github.com/HellSoft-Col/stock-market/internal/autoclient/strategy"
	"github.com/HellSoft-Col/stock-market/internal/client"
	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/rs/zerolog/log"
)

// PendingOrder tracks orders waiting for fills
type PendingOrder struct {
	OrderID  string
	ClOrdID  string
	Product  string
	Side     string
	Quantity int
	Price    float64
	SentTime time.Time
}

// TradingAgent coordinates strategies and executes trades
type TradingAgent struct {
	name     string
	client   *client.WebSocketClient
	strategy strategy.Strategy
	state    *market.MarketState

	// Execution
	executionInterval time.Duration
	stopCh            chan struct{}
	wg                sync.WaitGroup

	// Order tracking
	pendingOrders map[string]*PendingOrder // Track orders by ClOrdID
	orderTimeout  time.Duration            // Configurable timeout
	fillTimes     []time.Duration          // Track fill latencies
	avgFillTime   time.Duration            // Average fill time

	// Stats
	ordersSent    int
	fillsReceived int
	errorCount    int
	timeoutCount  int
	mu            sync.RWMutex
}

// NewTradingAgent creates a new trading agent
func NewTradingAgent(name string, client *client.WebSocketClient, strat strategy.Strategy) *TradingAgent {
	return &TradingAgent{
		name:              name,
		client:            client,
		strategy:          strat,
		state:             market.NewMarketState(name),
		executionInterval: 1 * time.Second, // Execute strategy every second
		orderTimeout:      5 * time.Minute, // Default 5 minute timeout (configurable)
		pendingOrders:     make(map[string]*PendingOrder),
		fillTimes:         make([]time.Duration, 0, 100),
		stopCh:            make(chan struct{}),
	}
}

// SetOrderTimeout sets the order timeout duration
func (a *TradingAgent) SetOrderTimeout(timeout time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.orderTimeout = timeout
	log.Info().
		Str("agent", a.name).
		Dur("timeout", timeout).
		Msg("Order timeout configured")
}

// Start starts the trading agent
func (a *TradingAgent) Start(ctx context.Context) error {
	log.Info().
		Str("agent", a.name).
		Str("strategy", a.strategy.Name()).
		Msg("Starting trading agent")

	// Start execution loop
	a.wg.Add(1)
	go a.executionLoop(ctx)

	// Start timeout checker loop
	a.wg.Add(1)
	go a.timeoutCheckerLoop(ctx)

	return nil
}

// Stop stops the trading agent
func (a *TradingAgent) Stop() error {
	log.Info().
		Str("agent", a.name).
		Msg("Stopping trading agent")

	close(a.stopCh)
	a.wg.Wait()

	return nil
}

// executionLoop periodically executes the strategy
func (a *TradingAgent) executionLoop(ctx context.Context) {
	defer a.wg.Done()

	ticker := time.NewTicker(a.executionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		case <-ticker.C:
			if err := a.executeStrategy(ctx); err != nil {
				log.Error().
					Err(err).
					Str("agent", a.name).
					Msg("Strategy execution error")
				a.incrementErrorCount()
			}
		}
	}
}

// timeoutCheckerLoop periodically checks for timed out orders
func (a *TradingAgent) timeoutCheckerLoop(ctx context.Context) {
	defer a.wg.Done()

	// Check every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		case <-ticker.C:
			a.checkOrderTimeouts()
		}
	}
}

// checkOrderTimeouts checks for orders that have exceeded timeout
func (a *TradingAgent) checkOrderTimeouts() {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	timedOut := make([]string, 0)

	for clOrdID, order := range a.pendingOrders {
		age := now.Sub(order.SentTime)
		if age > a.orderTimeout {
			log.Warn().
				Str("agent", a.name).
				Str("clOrdID", clOrdID).
				Str("product", order.Product).
				Str("side", order.Side).
				Dur("age", age).
				Dur("timeout", a.orderTimeout).
				Msg("⏱️ Order timeout - cancelling")

			timedOut = append(timedOut, clOrdID)
			a.timeoutCount++
		}
	}

	// Remove timed out orders from pending
	// Note: In a real implementation, send CANCEL messages
	for _, clOrdID := range timedOut {
		delete(a.pendingOrders, clOrdID)

		// Optionally send cancel message
		cancelMsg := &domain.CancelMessage{
			Type:    "CANCEL",
			ClOrdID: clOrdID,
		}
		if err := a.client.SendMessage(cancelMsg); err != nil {
			log.Error().
				Err(err).
				Str("clOrdID", clOrdID).
				Msg("Failed to send cancel for timed out order")
		}
	}
}

// executeStrategy executes the trading strategy
func (a *TradingAgent) executeStrategy(ctx context.Context) error {
	// Get actions from strategy
	actions, err := a.strategy.Execute(ctx, a.state)
	if err != nil {
		return fmt.Errorf("strategy execute failed: %w", err)
	}

	// Execute each action
	for _, action := range actions {
		if err := a.executeAction(action); err != nil {
			log.Error().
				Err(err).
				Str("agent", a.name).
				Str("actionType", string(action.Type)).
				Msg("Action execution failed")
			continue
		}
	}

	return nil
}

// executeAction executes a single trading action
func (a *TradingAgent) executeAction(action *strategy.Action) error {
	switch action.Type {
	case strategy.ActionTypeOrder:
		return a.sendOrder(action.Order)
	case strategy.ActionTypeCancel:
		return a.sendCancel(action.Cancel)
	case strategy.ActionTypeAcceptOffer:
		return a.sendAcceptOffer(action.AcceptOffer)
	case strategy.ActionTypeProduction:
		return a.sendProduction(action.Production)
	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// sendOrder sends an order to the server
func (a *TradingAgent) sendOrder(order *domain.OrderMessage) error {
	a.mu.Lock()
	a.ordersSent++

	// Track pending order
	price := 0.0
	if order.LimitPrice != nil {
		price = *order.LimitPrice
	}
	a.pendingOrders[order.ClOrdID] = &PendingOrder{
		OrderID:  "", // Will be set when we get ORDER_ACK
		ClOrdID:  order.ClOrdID,
		Product:  order.Product,
		Side:     order.Side,
		Quantity: order.Qty,
		Price:    price,
		SentTime: time.Now(),
	}
	a.mu.Unlock()

	// Add to market state for strategy visibility
	pricePtr := &price
	if price == 0 {
		pricePtr = nil
	}
	a.state.AddActiveOrder(order.ClOrdID, &domain.OrderSummary{
		ClOrdID:  order.ClOrdID,
		Product:  order.Product,
		Side:     order.Side,
		Mode:     order.Mode,
		Quantity: order.Qty,
		Price:    pricePtr,
	})

	if err := a.client.SendMessage(order); err != nil {
		// Remove from pending if send fails
		a.mu.Lock()
		delete(a.pendingOrders, order.ClOrdID)
		a.mu.Unlock()
		a.state.RemoveActiveOrder(order.ClOrdID)
		return fmt.Errorf("failed to send order: %w", err)
	}

	log.Info().
		Str("agent", a.name).
		Str("clOrdID", order.ClOrdID).
		Str("side", order.Side).
		Str("product", order.Product).
		Int("qty", order.Qty).
		Msg("📤 Order sent")

	return nil
}

// sendCancel sends a cancel request
func (a *TradingAgent) sendCancel(cancel *domain.CancelMessage) error {
	if err := a.client.SendMessage(cancel); err != nil {
		return fmt.Errorf("failed to send cancel: %w", err)
	}

	log.Info().
		Str("agent", a.name).
		Str("clOrdID", cancel.ClOrdID).
		Msg("Cancel sent")

	return nil
}

// sendAcceptOffer sends an accept offer message
func (a *TradingAgent) sendAcceptOffer(accept *domain.AcceptOfferMessage) error {
	if err := a.client.SendMessage(accept); err != nil {
		return fmt.Errorf("failed to send accept offer: %w", err)
	}

	log.Info().
		Str("agent", a.name).
		Str("offerID", accept.OfferID).
		Bool("accepting", accept.Accept).
		Int("qty", accept.QuantityOffered).
		Float64("price", accept.PriceOffered).
		Msg("🤝 ACCEPTING OFFER: We're filling another team's order!")

	return nil
}

// sendProduction sends a production update
func (a *TradingAgent) sendProduction(production *domain.ProductionUpdateMessage) error {
	if err := a.client.SendMessage(production); err != nil {
		return fmt.Errorf("failed to send production: %w", err)
	}

	log.Info().
		Str("agent", a.name).
		Str("product", production.Product).
		Int("qty", production.Quantity).
		Msg("Production sent")

	return nil
}

// HandleLoginOK handles login success
func (a *TradingAgent) HandleLoginOK(loginOK *domain.LoginOKMessage) error {
	// Update state
	a.state.TeamName = loginOK.Team
	a.state.Species = loginOK.Species
	a.state.Balance = loginOK.CurrentBalance
	a.state.InitialBalance = loginOK.CurrentBalance
	a.state.Inventory = loginOK.Inventory
	a.state.Recipes = loginOK.Recipes
	a.state.Role = loginOK.Role

	// Initialize strategy
	return a.strategy.OnLogin(context.Background(), loginOK)
}

// HandleFill handles fill messages
func (a *TradingAgent) HandleFill(fill *domain.FillMessage) error {
	a.mu.Lock()
	a.fillsReceived++

	// Check if we have pending order and calculate fill time
	if pending, exists := a.pendingOrders[fill.ClOrdID]; exists {
		fillTime := time.Since(pending.SentTime)

		// Track fill time
		a.fillTimes = append(a.fillTimes, fillTime)
		if len(a.fillTimes) > 100 {
			a.fillTimes = a.fillTimes[1:]
		}

		// Calculate average
		a.calculateAvgFillTime()

		log.Info().
			Str("agent", a.name).
			Str("clOrdID", fill.ClOrdID).
			Str("product", fill.Product).
			Str("side", fill.Side).
			Int("qty", fill.FillQty).
			Float64("price", fill.FillPrice).
			Dur("fillTime", fillTime).
			Dur("avgFillTime", a.avgFillTime).
			Str("counterparty", fill.Counterparty).
			Str("counterpartyMessage", fill.CounterpartyMessage).
			Msg("✅ FILL: Our order accepted by another team!")

		// Remove from pending
		delete(a.pendingOrders, fill.ClOrdID)

		// Remove from market state
		a.mu.Unlock()
		a.state.RemoveActiveOrder(fill.ClOrdID)
		a.mu.Lock()
	} else {
		log.Debug().
			Str("clOrdID", fill.ClOrdID).
			Msg("Fill received for unknown order (might be old or from resync)")
	}

	a.mu.Unlock()

	// Update state based on fill
	switch fill.Side {
	case "BUY":
		// Bought: add to inventory, reduce balance
		cost := fill.FillPrice * float64(fill.FillQty)
		a.state.UpdateBalance(a.state.Balance - cost)
		a.state.AddInventory(fill.Product, fill.FillQty)
	case "SELL":
		// Sold: remove from inventory, increase balance
		revenue := fill.FillPrice * float64(fill.FillQty)
		a.state.UpdateBalance(a.state.Balance + revenue)
		a.state.RemoveInventory(fill.Product, fill.FillQty)
	}

	// Add to history
	a.state.AddFill(fill)

	// Notify strategy
	return a.strategy.OnFill(context.Background(), fill)
}

// calculateAvgFillTime calculates average fill time (must be called with lock held)
func (a *TradingAgent) calculateAvgFillTime() {
	if len(a.fillTimes) == 0 {
		a.avgFillTime = 0
		return
	}

	var total time.Duration
	for _, t := range a.fillTimes {
		total += t
	}
	a.avgFillTime = total / time.Duration(len(a.fillTimes))
}

// HandleTicker handles ticker updates
func (a *TradingAgent) HandleTicker(ticker *domain.TickerMessage) error {
	a.state.UpdateTicker(ticker)
	return a.strategy.OnTicker(context.Background(), ticker)
}

// HandleOffer handles offer messages
func (a *TradingAgent) HandleOffer(offer *domain.OfferMessage) error {
	a.state.AddOffer(offer)

	// Get response from strategy
	response, err := a.strategy.OnOffer(context.Background(), offer)
	if err != nil {
		return err
	}

	// If strategy accepts, send accept message
	if response.Accept {
		acceptMsg := strategy.CreateAcceptOffer(
			offer.OfferID,
			response.QuantityOffered,
			response.PriceOffered,
		)
		return a.sendAcceptOffer(acceptMsg)
	}

	return nil
}

// HandleInventoryUpdate handles inventory updates
func (a *TradingAgent) HandleInventoryUpdate(inventory map[string]int) error {
	a.state.UpdateInventory(inventory)
	return a.strategy.OnInventoryUpdate(context.Background(), inventory)
}

// HandleBalanceUpdate handles balance updates
func (a *TradingAgent) HandleBalanceUpdate(balance float64) error {
	a.state.UpdateBalance(balance)
	return a.strategy.OnBalanceUpdate(context.Background(), balance)
}

// HandleOrderBookUpdate handles orderbook updates
func (a *TradingAgent) HandleOrderBookUpdate(orderbook *domain.OrderBookUpdateMessage) error {
	// TODO: Update state with orderbook
	return a.strategy.OnOrderBookUpdate(context.Background(), orderbook)
}

// HandleError handles error messages
func (a *TradingAgent) HandleError(errMsg *domain.ErrorMessage) {
	log.Error().
		Str("agent", a.name).
		Str("message", errMsg.Reason).
		Msg("Server error received")

	a.incrementErrorCount()
}

// incrementErrorCount safely increments the error counter
func (a *TradingAgent) incrementErrorCount() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.errorCount++
}

// GetStats returns agent statistics
func (a *TradingAgent) GetStats() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	pnl := a.state.CalculatePnL()

	return map[string]interface{}{
		"ordersSent":     a.ordersSent,
		"fillsReceived":  a.fillsReceived,
		"errorCount":     a.errorCount,
		"timeoutCount":   a.timeoutCount,
		"pendingOrders":  len(a.pendingOrders),
		"avgFillTime":    a.avgFillTime.String(),
		"pnl":            pnl,
		"balance":        a.state.Balance,
		"strategyHealth": a.strategy.Health(),
	}
}
