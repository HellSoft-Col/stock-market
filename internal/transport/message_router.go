package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/HellSoft-Col/stock-market/internal/market"
	"github.com/HellSoft-Col/stock-market/internal/service"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type MessageRouter struct {
	authService        domain.AuthService
	orderService       domain.OrderService
	broadcaster        domain.Broadcaster
	marketService      domain.MarketService
	resyncService      domain.ResyncService
	productionService  domain.ProductionService
	performanceService domain.PerformanceService
	rateLimiter        *service.RateLimiter
	orderRepo          domain.OrderRepository
	orderBook          domain.OrderBookRepository
	fillRepo           domain.FillRepository
}

func NewMessageRouter(
	authService domain.AuthService,
	orderService domain.OrderService,
	broadcaster domain.Broadcaster,
	marketService domain.MarketService,
	resyncService domain.ResyncService,
	productionService domain.ProductionService,
	performanceService domain.PerformanceService,
	rateLimiter *service.RateLimiter,
	orderRepo domain.OrderRepository,
	orderBook domain.OrderBookRepository,
	fillRepo domain.FillRepository,
) *MessageRouter {
	return &MessageRouter{
		authService:        authService,
		orderService:       orderService,
		broadcaster:        broadcaster,
		marketService:      marketService,
		resyncService:      resyncService,
		productionService:  productionService,
		performanceService: performanceService,
		rateLimiter:        rateLimiter,
		orderRepo:          orderRepo,
		orderBook:          orderBook,
		fillRepo:           fillRepo,
	}
}

func (r *MessageRouter) RouteMessage(ctx context.Context, rawMessage string, client MessageClient) error {
	// Parse base message to get type
	var baseMsg domain.BaseMessage
	if err := json.Unmarshal([]byte(rawMessage), &baseMsg); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	log.Debug().
		Str("type", baseMsg.Type).
		Str("clientAddr", client.GetRemoteAddr()).
		Str("teamName", client.GetTeamName()).
		Msg("Routing message")

	switch baseMsg.Type {
	case "LOGIN":
		return r.handleLogin(ctx, rawMessage, client)
	case "ORDER":
		return r.handleOrder(ctx, rawMessage, client)
	case "PRODUCTION_UPDATE":
		return r.handleProductionUpdate(ctx, rawMessage, client)
	case "ACCEPT_OFFER":
		return r.handleAcceptOffer(ctx, rawMessage, client)
	case "RESYNC":
		return r.handleResync(ctx, rawMessage, client)
	case "CANCEL":
		return r.handleCancelOrder(ctx, rawMessage, client)
	case "REQUEST_ALL_ORDERS":
		return r.handleRequestAllOrders(ctx, client)
	case "REQUEST_HISTORICAL_ORDERS":
		return r.handleRequestHistoricalOrders(ctx, rawMessage, client)
	case "REQUEST_PERFORMANCE_REPORT":
		return r.handleRequestPerformanceReport(ctx, rawMessage, client)
	case "REQUEST_CONNECTED_SESSIONS":
		return r.handleRequestConnectedSessions(ctx, client)
	case "PING":
		return r.handlePing(ctx, client)
	case "ADMIN_CANCEL_ALL_ORDERS":
		return r.handleAdminCancelAllOrders(ctx, client)
	case "ADMIN_CANCEL_ORDER":
		return r.handleAdminCancelOrder(ctx, rawMessage, client)
	case "ADMIN_BROADCAST":
		return r.handleAdminBroadcast(ctx, rawMessage, client)
	case "ADMIN_CREATE_ORDER":
		return r.handleAdminCreateOrder(ctx, rawMessage, client)
	case "EXPORT_DATA":
		return r.handleExportData(ctx, rawMessage, client)
	case "GET_AVAILABLE_TEAMS":
		return r.handleGetAvailableTeams(ctx, client)
	case "GET_TEAM_ACTIVITY":
		return r.handleGetTeamActivity(ctx, rawMessage, client)
	case "GET_ALL_TEAMS":
		return r.handleGetAllTeams(ctx, client)
	case "UPDATE_TEAM":
		return r.handleUpdateTeam(ctx, rawMessage, client)
	case "UPDATE_TEAM_MEMBERS":
		return r.handleUpdateTeamMembers(ctx, rawMessage, client)
	case "RESET_TEAM_BALANCE":
		return r.handleResetTeamBalance(ctx, rawMessage, client)
	case "RESET_TEAM_INVENTORY":
		return r.handleResetTeamInventory(ctx, rawMessage, client)
	case "RESET_TEAM_PRODUCTION":
		return r.handleResetTeamProduction(ctx, rawMessage, client)
	case "RESET_TOURNAMENT_CONFIG":
		return r.handleResetTournamentConfig(ctx, rawMessage, client)
	case "UPDATE_ALL_RECIPES":
		return r.handleUpdateAllRecipes(ctx, client)
	case "SDK_EMULATOR":
		return r.handleSDKEmulator(ctx, rawMessage, client)
	default:
		return r.handleEcho(ctx, baseMsg, client)
	}
}

func (r *MessageRouter) handleLogin(ctx context.Context, rawMessage string, client MessageClient) error {
	// Parse message
	var loginMsg domain.LoginMessage
	if err := json.Unmarshal([]byte(rawMessage), &loginMsg); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid LOGIN message format", "")
	}

	// Validate token
	if loginMsg.Token == "" {
		return r.sendError(client, domain.ErrAuthFailed, "Token is required", "")
	}

	// Authenticate team
	team, err := r.authService.ValidateToken(ctx, loginMsg.Token)
	if err != nil {
		return r.sendError(client, domain.ErrAuthFailed, "Invalid token", "")
	}

	// Create session for this client
	userAgent := ""
	if wsClient, ok := client.(*WebSocketClientHandler); ok {
		userAgent = wsClient.GetUserAgent()
	}
	authService, ok := r.authService.(*service.AuthService)
	if ok {
		authService.CreateSession(team.TeamName, loginMsg.Token, client.GetRemoteAddr(), userAgent)
	}

	// Set client team name and register
	client.SetTeamName(team.TeamName)
	client.RegisterWithServer(team.TeamName)
	r.broadcaster.RegisterClient(team.TeamName, client)

	// Ensure team has correct recipes - if missing or empty, rebuild them
	if team.Recipes == nil || len(team.Recipes) == 0 {
		log.Warn().
			Str("teamName", team.TeamName).
			Str("species", team.Species).
			Msg("Team has no recipes - rebuilding from species")

		// Get basic product from authorized products
		basicProduct := ""
		if len(team.AuthorizedProducts) > 0 {
			basicProduct = team.AuthorizedProducts[0]
		}

		// Build recipes
		team.Recipes = buildRecipesForSpecies(team.Species, basicProduct)

		// Update in database
		if authSvc, ok := r.authService.(*service.AuthService); ok {
			if err := authSvc.UpdateRecipes(ctx, team.TeamName, team.Recipes); err != nil {
				log.Error().
					Err(err).
					Str("teamName", team.TeamName).
					Msg("Failed to update team recipes in database")
			} else {
				log.Info().
					Str("teamName", team.TeamName).
					Msg("Team recipes updated in database during login")
			}
		}
	}

	// Get team inventory
	inventory := team.Inventory
	if inventory == nil {
		inventory = make(map[string]int)
	}

	// Ensure role has default energy values if missing or zero
	role := team.Role
	if role.BaseEnergy == 0 {
		role.BaseEnergy = 3.0 // Default base energy
		log.Warn().
			Str("teamName", team.TeamName).
			Msg("Team missing baseEnergy, using default value of 3.0")
	}
	if role.LevelEnergy == 0 {
		role.LevelEnergy = 2.0 // Default level energy
		log.Warn().
			Str("teamName", team.TeamName).
			Msg("Team missing levelEnergy, using default value of 2.0")
	}

	// Generate and send LOGIN_OK response
	loginOKMsg := &domain.LoginOKMessage{
		Type:               "LOGIN_OK",
		Team:               team.TeamName,
		Species:            team.Species,
		InitialBalance:     team.InitialBalance,
		CurrentBalance:     team.CurrentBalance,
		Inventory:          inventory,
		AuthorizedProducts: team.AuthorizedProducts,
		Recipes:            team.Recipes,
		Role:               role,
		ServerTime:         time.Now().Format(time.RFC3339),
	}

	// Debug: Log the role values being sent
	log.Info().
		Str("teamName", team.TeamName).
		Str("species", team.Species).
		Str("clientAddr", client.GetRemoteAddr()).
		Int("roleBranches", role.Branches).
		Int("roleMaxDepth", role.MaxDepth).
		Float64("roleDecay", role.Decay).
		Float64("roleBudget", role.Budget).
		Float64("roleBaseEnergy", role.BaseEnergy).
		Float64("roleLevelEnergy", role.LevelEnergy).
		Msg("Team logged in successfully")

	// Debug: Log the full JSON being sent
	if jsonBytes, err := json.Marshal(loginOKMsg); err == nil {
		log.Debug().
			Str("teamName", team.TeamName).
			Str("loginJSON", string(jsonBytes)).
			Msg("LOGIN_OK JSON payload")
	}

	return client.SendMessage(loginOKMsg)
}

func (r *MessageRouter) handleOrder(ctx context.Context, rawMessage string, client MessageClient) error {
	// Check authentication
	if client == nil || client.GetTeamName() == "" {
		return r.sendError(client, domain.ErrAuthFailed, "Must login first", "")
	}

	// Check rate limit
	if !r.rateLimiter.Allow(client.GetTeamName()) {
		return r.sendError(client, domain.ErrRateLimitExceeded, "Too many orders per minute", "")
	}

	// Parse message
	var orderMsg domain.OrderMessage
	if err := json.Unmarshal([]byte(rawMessage), &orderMsg); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid ORDER message format", "")
	}

	// Validate order message fields
	if orderMsg.ClOrdID == "" {
		return r.sendError(client, domain.ErrInvalidMessage, "Order ID (clOrdID) is required", "")
	}
	if orderMsg.Product == "" {
		return r.sendError(client, domain.ErrInvalidProduct, "Product is required", orderMsg.ClOrdID)
	}
	if orderMsg.Side != "BUY" && orderMsg.Side != "SELL" {
		return r.sendError(client, domain.ErrInvalidMessage, "Side must be BUY or SELL", orderMsg.ClOrdID)
	}
	if orderMsg.Qty <= 0 {
		return r.sendError(client, domain.ErrInvalidQuantity, "Quantity must be positive", orderMsg.ClOrdID)
	}
	if orderMsg.Mode != "MARKET" && orderMsg.Mode != "LIMIT" {
		return r.sendError(client, domain.ErrInvalidMessage, "Mode must be MARKET or LIMIT", orderMsg.ClOrdID)
	}
	if orderMsg.Mode == "LIMIT" && (orderMsg.LimitPrice == nil || *orderMsg.LimitPrice <= 0) {
		return r.sendError(
			client,
			domain.ErrInvalidMessage,
			"Limit price must be positive for LIMIT orders",
			orderMsg.ClOrdID,
		)
	}

	// Process order
	if err := r.orderService.ProcessOrder(ctx, client.GetTeamName(), &orderMsg); err != nil {
		log.Warn().
			Str("teamName", client.GetTeamName()).
			Str("clOrdID", orderMsg.ClOrdID).
			Err(err).
			Msg("Order processing failed")

		return r.sendError(client, r.getOrderErrorCode(err, orderMsg.ClOrdID), err.Error(), orderMsg.ClOrdID)
	}

	log.Info().
		Str("teamName", client.GetTeamName()).
		Str("clOrdID", orderMsg.ClOrdID).
		Str("side", orderMsg.Side).
		Str("product", orderMsg.Product).
		Int("qty", orderMsg.Qty).
		Msg("Order accepted for processing")

	return nil
}

func (r *MessageRouter) getOrderErrorCode(err error, clOrdID string) string {
	switch {
	case err.Error() == "duplicate order ID: "+clOrdID:
		return domain.ErrDuplicateOrderID
	case contains(err.Error(), "invalid product"):
		return domain.ErrInvalidProduct
	case contains(err.Error(), "quantity"):
		return domain.ErrInvalidQuantity
	default:
		return domain.ErrInvalidOrder
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (r *MessageRouter) handleProductionUpdate(ctx context.Context, rawMessage string, client MessageClient) error {
	// Check authentication
	if client == nil || client.GetTeamName() == "" {
		return r.sendError(client, domain.ErrAuthFailed, "Must login first", "")
	}

	// Parse message
	var prodMsg domain.ProductionUpdateMessage
	if err := json.Unmarshal([]byte(rawMessage), &prodMsg); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid PRODUCTION_UPDATE message format", "")
	}

	// Validate production message
	if prodMsg.Product == "" {
		return r.sendError(client, domain.ErrInvalidProduct, "Product is required", "")
	}
	if prodMsg.Quantity <= 0 {
		return r.sendError(client, domain.ErrInvalidQuantity, "Quantity must be positive", "")
	}

	// Process production update
	if err := r.productionService.ProcessProduction(ctx, client.GetTeamName(), &prodMsg); err != nil {
		log.Warn().
			Str("teamName", client.GetTeamName()).
			Str("product", prodMsg.Product).
			Err(err).
			Msg("Production update failed")

		return r.sendError(client, r.getProductionErrorCode(err), err.Error(), "")
	}

	log.Info().
		Str("teamName", client.GetTeamName()).
		Str("product", prodMsg.Product).
		Int("quantity", prodMsg.Quantity).
		Msg("Production update processed")

	return nil
}

func (r *MessageRouter) getProductionErrorCode(err error) string {
	switch {
	case contains(err.Error(), "not authorized"):
		return domain.ErrUnauthorizedProduction
	case contains(err.Error(), "invalid product"):
		return domain.ErrInvalidProduct
	case contains(err.Error(), "quantity"):
		return domain.ErrInvalidQuantity
	default:
		return domain.ErrInvalidMessage
	}
}

func (r *MessageRouter) handleAcceptOffer(ctx context.Context, rawMessage string, client MessageClient) error {
	// Check authentication
	if client == nil || client.GetTeamName() == "" {
		return r.sendError(client, domain.ErrAuthFailed, "Must login first", "")
	}

	// Parse message
	var acceptMsg domain.AcceptOfferMessage
	if err := json.Unmarshal([]byte(rawMessage), &acceptMsg); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid ACCEPT_OFFER message format", "")
	}

	// Validate accept offer message
	if acceptMsg.OfferID == "" {
		return r.sendError(client, domain.ErrInvalidMessage, "Offer ID is required", "")
	}

	// Get market engine and offer generator
	marketEngine, ok := r.marketService.(*market.MarketEngine)
	if !ok || marketEngine == nil || marketEngine.OfferGenerator == nil {
		return r.sendError(client, domain.ErrServiceUnavailable, "Offer handling service not available", "")
	}

	// Process offer acceptance
	if err := marketEngine.OfferGenerator.HandleAcceptOffer(&acceptMsg, client.GetTeamName()); err != nil {
		log.Warn().
			Str("teamName", client.GetTeamName()).
			Str("offerID", acceptMsg.OfferID).
			Err(err).
			Msg("Failed to handle offer acceptance")

		return r.sendError(client, r.getOfferErrorCode(err, acceptMsg.OfferID), err.Error(), "")
	}

	log.Info().
		Str("teamName", client.GetTeamName()).
		Str("offerID", acceptMsg.OfferID).
		Bool("accept", acceptMsg.Accept).
		Msg("Offer response processed")

	return nil
}

func (r *MessageRouter) getOfferErrorCode(err error, offerID string) string {
	if err.Error() == "offer not found or expired: "+offerID {
		return domain.ErrOfferExpired
	}
	return domain.ErrOfferExpired
}

func (r *MessageRouter) handleResync(ctx context.Context, rawMessage string, client MessageClient) error {
	// Check if client is authenticated
	if client == nil || client.GetTeamName() == "" {
		return r.sendError(client, domain.ErrAuthFailed, "Must login first", "")
	}

	var resyncMsg domain.ResyncMessage
	if err := json.Unmarshal([]byte(rawMessage), &resyncMsg); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid RESYNC message format", "")
	}

	// Parse the lastSync timestamp
	var since time.Time
	var err error
	if resyncMsg.LastSync != "" {
		since, err = time.Parse(time.RFC3339, resyncMsg.LastSync)
		if err != nil {
			return r.sendError(client, domain.ErrInvalidMessage, "Invalid lastSync timestamp format", "")
		}
	} else {
		// If no lastSync provided, return events from 24 hours ago
		since = time.Now().Add(-24 * time.Hour)
	}

	// Generate event delta
	eventDelta, err := r.resyncService.GenerateEventDelta(ctx, client.GetTeamName(), since)
	if err != nil {
		log.Error().
			Str("teamName", client.GetTeamName()).
			Err(err).
			Msg("Failed to generate event delta")
		return r.sendError(client, domain.ErrServiceUnavailable, "Failed to generate resync data", "")
	}

	log.Info().
		Str("teamName", client.GetTeamName()).
		Time("since", since).
		Int("eventCount", len(eventDelta.Events)).
		Msg("Resync completed")

	return client.SendMessage(eventDelta)
}

func (r *MessageRouter) handleCancelOrder(ctx context.Context, rawMessage string, client MessageClient) error {
	if client == nil || client.GetTeamName() == "" {
		return r.sendError(client, domain.ErrAuthFailed, "Must login first", "")
	}

	var cancelMsg domain.CancelMessage
	if err := json.Unmarshal([]byte(rawMessage), &cancelMsg); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid CANCEL message format", "")
	}

	if cancelMsg.ClOrdID == "" {
		return r.sendError(client, domain.ErrInvalidMessage, "ClOrdID is required", "")
	}

	// Check if orderRepo is available
	if r.orderRepo == nil {
		log.Error().Msg("OrderRepository is nil")
		return r.sendError(client, domain.ErrServiceUnavailable, "Order service unavailable", cancelMsg.ClOrdID)
	}

	// Get the order to verify ownership
	order, err := r.orderRepo.GetByClOrdID(ctx, cancelMsg.ClOrdID)
	if err != nil || order == nil {
		log.Warn().
			Str("clOrdID", cancelMsg.ClOrdID).
			Str("teamName", client.GetTeamName()).
			Msg("Order not found for cancellation")
		return r.sendError(client, domain.ErrInvalidOrder, "Order not found", cancelMsg.ClOrdID)
	}

	// Verify the order belongs to this team
	if order.TeamName != client.GetTeamName() {
		log.Warn().
			Str("clOrdID", cancelMsg.ClOrdID).
			Str("orderTeam", order.TeamName).
			Str("requestingTeam", client.GetTeamName()).
			Msg("Team attempted to cancel another team's order")
		return r.sendError(client, domain.ErrAuthFailed, "Cannot cancel another team's order", cancelMsg.ClOrdID)
	}

	// Check if order is already filled or cancelled
	if order.Status == "FILLED" {
		return r.sendError(client, domain.ErrInvalidOrder, "Cannot cancel filled order", cancelMsg.ClOrdID)
	}
	if order.Status == "CANCELLED" {
		return r.sendError(client, domain.ErrInvalidOrder, "Order already cancelled", cancelMsg.ClOrdID)
	}

	// Cancel the order
	if err := r.orderRepo.Cancel(ctx, cancelMsg.ClOrdID); err != nil {
		log.Error().
			Err(err).
			Str("clOrdID", cancelMsg.ClOrdID).
			Msg("Failed to cancel order")
		return r.sendError(client, domain.ErrServiceUnavailable, "Failed to cancel order", cancelMsg.ClOrdID)
	}

	// Remove from order book if it's there
	r.orderBook.RemoveOrder(order.Product, order.Side, cancelMsg.ClOrdID)

	// Send acknowledgment
	ack := &domain.OrderAckMessage{
		Type:       "ORDER_ACK",
		ClOrdID:    cancelMsg.ClOrdID,
		Status:     "CANCELLED",
		ServerTime: time.Now().Format(time.RFC3339),
	}

	log.Info().
		Str("clOrdID", cancelMsg.ClOrdID).
		Str("teamName", client.GetTeamName()).
		Str("product", order.Product).
		Msg("Order cancelled successfully")

	return client.SendMessage(ack)
}

func (r *MessageRouter) handleEcho(ctx context.Context, baseMsg domain.BaseMessage, client MessageClient) error {
	// Echo back for testing unknown message types
	response := map[string]any{
		"type":          "ECHO",
		"original":      baseMsg,
		"time":          time.Now().Format(time.RFC3339),
		"authenticated": client.GetTeamName() != "",
	}

	return client.SendMessage(response)
}

func (r *MessageRouter) handlePing(ctx context.Context, client MessageClient) error {
	response := map[string]any{
		"type":      "PONG",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	return client.SendMessage(response)
}

func (r *MessageRouter) handleRequestAllOrders(ctx context.Context, client MessageClient) error {
	if client == nil || client.GetTeamName() == "" {
		return r.sendError(client, domain.ErrAuthFailed, "Must login first", "")
	}

	if r.orderRepo == nil {
		log.Error().Msg("OrderRepository is nil")
		return r.sendError(client, domain.ErrServiceUnavailable, "Order service unavailable", "")
	}

	orders, err := r.orderRepo.GetPendingOrders(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get pending orders")
		return r.sendError(client, domain.ErrServiceUnavailable, "Failed to get orders", "")
	}

	teamName := client.GetTeamName()
	isAdmin := teamName == "admin"
	orderSummaries := make([]*domain.OrderSummary, 0)
	for _, order := range orders {
		if order == nil {
			continue
		}
		if !isAdmin && order.TeamName != teamName {
			continue
		}
		summary := &domain.OrderSummary{
			ClOrdID:   order.ClOrdID,
			TeamName:  order.TeamName,
			Side:      order.Side,
			Mode:      order.Mode,
			Product:   order.Product,
			Quantity:  order.Quantity,
			Price:     order.Price,
			FilledQty: order.FilledQty,
			Message:   order.Message,
			CreatedAt: order.CreatedAt.Format(time.RFC3339),
		}
		orderSummaries = append(orderSummaries, summary)
	}

	response := &domain.AllOrdersMessage{
		Type:       "ALL_ORDERS",
		Orders:     orderSummaries,
		ServerTime: time.Now().Format(time.RFC3339),
	}

	return client.SendMessage(response)
}

func (r *MessageRouter) handleRequestHistoricalOrders(ctx context.Context, rawMessage string, client MessageClient) error {
	// Only admin can view historical orders
	if client == nil || client.GetTeamName() != "admin" {
		return r.sendError(client, domain.ErrAuthFailed, "Admin access required", "")
	}

	if r.orderRepo == nil {
		log.Error().Msg("OrderRepository is nil")
		return r.sendError(client, domain.ErrServiceUnavailable, "Order service unavailable", "")
	}

	// Parse message to get limit
	var requestMsg struct {
		Type  string `json:"type"`
		Limit int    `json:"limit"`
	}
	if err := json.Unmarshal([]byte(rawMessage), &requestMsg); err != nil {
		log.Warn().Err(err).Msg("Failed to parse historical orders request")
		requestMsg.Limit = 100 // Default limit
	}

	// Get historical orders from repository
	orders, err := r.orderRepo.GetHistoricalOrders(ctx, requestMsg.Limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get historical orders")
		return r.sendError(client, domain.ErrServiceUnavailable, "Failed to get historical orders", "")
	}

	// Convert to summaries
	orderSummaries := make([]*domain.OrderSummary, 0, len(orders))
	for _, order := range orders {
		if order == nil {
			continue
		}
		summary := &domain.OrderSummary{
			ClOrdID:   order.ClOrdID,
			TeamName:  order.TeamName,
			Side:      order.Side,
			Mode:      order.Mode,
			Product:   order.Product,
			Quantity:  order.Quantity,
			Price:     order.Price,
			FilledQty: order.FilledQty,
			Status:    order.Status,
			Message:   order.Message,
			CreatedAt: order.CreatedAt.Format(time.RFC3339),
		}
		if !order.UpdatedAt.IsZero() {
			summary.UpdatedAt = order.UpdatedAt.Format(time.RFC3339)
		}
		orderSummaries = append(orderSummaries, summary)
	}

	response := &domain.HistoricalOrdersMessage{
		Type:       "HISTORICAL_ORDERS",
		Orders:     orderSummaries,
		Count:      len(orderSummaries),
		ServerTime: time.Now().Format(time.RFC3339),
	}

	return client.SendMessage(response)
}

func (r *MessageRouter) handleRequestOrderBook(ctx context.Context, rawMessage string, client MessageClient) error {
	if client == nil || client.GetTeamName() == "" {
		return r.sendError(client, domain.ErrAuthFailed, "Must login first", "")
	}

	var request struct {
		Type    string `json:"type"`
		Product string `json:"product"`
	}

	if err := json.Unmarshal([]byte(rawMessage), &request); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid REQUEST_ORDER_BOOK message format", "")
	}

	if request.Product == "" {
		return r.sendError(client, domain.ErrInvalidMessage, "Product is required", "")
	}

	if r.orderBook == nil {
		log.Error().Msg("OrderBook is nil")
		return r.sendError(client, domain.ErrServiceUnavailable, "Order book service unavailable", "")
	}

	buyOrders := r.orderBook.GetBuyOrders(request.Product)
	sellOrders := r.orderBook.GetSellOrders(request.Product)

	buySummaries := make([]*domain.OrderSummary, 0, len(buyOrders))
	for _, order := range buyOrders {
		summary := &domain.OrderSummary{
			ClOrdID:   order.ClOrdID,
			TeamName:  order.TeamName,
			Side:      order.Side,
			Mode:      order.Mode,
			Product:   order.Product,
			Quantity:  order.Quantity,
			Price:     order.Price,
			FilledQty: order.FilledQty,
			Message:   order.Message,
			CreatedAt: order.CreatedAt.Format(time.RFC3339),
		}
		buySummaries = append(buySummaries, summary)
	}

	sellSummaries := make([]*domain.OrderSummary, 0, len(sellOrders))
	for _, order := range sellOrders {
		summary := &domain.OrderSummary{
			ClOrdID:   order.ClOrdID,
			TeamName:  order.TeamName,
			Side:      order.Side,
			Mode:      order.Mode,
			Product:   order.Product,
			Quantity:  order.Quantity,
			Price:     order.Price,
			FilledQty: order.FilledQty,
			Message:   order.Message,
			CreatedAt: order.CreatedAt.Format(time.RFC3339),
		}
		sellSummaries = append(sellSummaries, summary)
	}

	response := &domain.OrderBookUpdateMessage{
		Type:       "ORDER_BOOK_UPDATE",
		Product:    request.Product,
		BuyOrders:  buySummaries,
		SellOrders: sellSummaries,
		ServerTime: time.Now().Format(time.RFC3339),
	}

	return client.SendMessage(response)
}

func (r *MessageRouter) handleRequestConnectedSessions(ctx context.Context, client MessageClient) error {
	if client == nil || client.GetTeamName() == "" {
		return r.sendError(client, domain.ErrAuthFailed, "Must login first", "")
	}

	if r.authService == nil {
		log.Error().Msg("AuthService is nil")
		return r.sendError(client, domain.ErrServiceUnavailable, "Session service unavailable", "")
	}

	// Get detailed session information from auth service
	authService, ok := r.authService.(*service.AuthService)
	if !ok {
		// Fallback to simple broadcaster info
		return r.handleRequestConnectedSessionsFallback(ctx, client)
	}

	activeSessions := authService.GetActiveSessions()
	if activeSessions == nil {
		activeSessions = make(map[string][]*service.Session)
	}

	sessions := make([]*domain.SessionInfo, 0)
	totalConnections := 0

	for _, teamSessions := range activeSessions {
		if teamSessions == nil {
			continue
		}
		for _, session := range teamSessions {
			if session == nil {
				continue
			}
			sessionInfo := &domain.SessionInfo{
				TeamName:      session.TeamName,
				RemoteAddr:    session.RemoteAddr,
				UserAgent:     session.UserAgent,
				ClientType:    session.ClientType,
				ConnectedAt:   session.ConnectedAt.Format(time.RFC3339),
				LastActivity:  session.LastActivity.Format(time.RFC3339),
				Authenticated: true,
			}
			sessions = append(sessions, sessionInfo)
			totalConnections++
		}
	}

	response := &domain.ConnectedSessionsMessage{
		Type:       "CONNECTED_SESSIONS",
		Sessions:   sessions,
		ServerTime: time.Now().Format(time.RFC3339),
	}

	// Also send server stats
	statsResponse := &domain.ServerStatsMessage{
		Type: "SERVER_STATS",
		Stats: map[string]interface{}{
			"totalConnections": totalConnections,
			"uniqueTeams":      len(activeSessions),
			"totalOrders":      0,         // We could track this
			"totalFills":       0,         // We could track this
			"uptime":           "Unknown", // We could track this
		},
		ServerTime: time.Now().Format(time.RFC3339),
	}

	// Send both messages
	if err := client.SendMessage(response); err != nil {
		return err
	}

	return client.SendMessage(statsResponse)
}

func (r *MessageRouter) handleRequestConnectedSessionsFallback(ctx context.Context, client MessageClient) error {
	connectedClients := r.broadcaster.GetConnectedClients()

	// Group clients by team name and count multiple connections
	teamConnections := make(map[string]int)
	for _, teamName := range connectedClients {
		teamConnections[teamName]++
	}

	sessions := make([]*domain.SessionInfo, 0, len(teamConnections))
	for teamName, count := range teamConnections {
		// Determine likely client type based on team connections
		clientType := "Java Client"
		if count > 1 {
			clientType = "Multiple Clients" // Web + Java
		}

		session := &domain.SessionInfo{
			TeamName:      teamName,
			RemoteAddr:    fmt.Sprintf("%d connections", count),
			ClientType:    clientType,
			ConnectedAt:   time.Now().Add(-time.Minute * 5).Format(time.RFC3339),
			LastActivity:  time.Now().Format(time.RFC3339),
			Authenticated: teamName != "" && teamName != "Anonymous",
		}
		sessions = append(sessions, session)
	}

	response := &domain.ConnectedSessionsMessage{
		Type:       "CONNECTED_SESSIONS",
		Sessions:   sessions,
		ServerTime: time.Now().Format(time.RFC3339),
	}

	return client.SendMessage(response)
}

func (r *MessageRouter) handleRequestPerformanceReport(
	ctx context.Context,
	rawMessage string,
	client MessageClient,
) error {
	if client == nil || client.GetTeamName() == "" {
		return r.sendError(client, domain.ErrAuthFailed, "Must login first", "")
	}

	if r.performanceService == nil {
		log.Error().Msg("PerformanceService is nil")
		return r.sendError(client, domain.ErrServiceUnavailable, "Performance service unavailable", "")
	}

	var request struct {
		Type      string `json:"type"`
		Scope     string `json:"scope"`     // "team", "global"
		TeamName  string `json:"teamName"`  // For team reports, optional if requesting own team
		StartTime string `json:"startTime"` // RFC3339 format, optional
	}

	if err := json.Unmarshal([]byte(rawMessage), &request); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid REQUEST_PERFORMANCE_REPORT message format", "")
	}

	if request.Scope == "" {
		request.Scope = "team" // Default to team scope
	}

	// Parse start time or default to 24 hours ago
	var since time.Time
	if request.StartTime != "" {
		var err error
		since, err = time.Parse(time.RFC3339, request.StartTime)
		if err != nil {
			return r.sendError(client, domain.ErrInvalidMessage, "Invalid startTime format", "")
		}
	} else {
		since = time.Now().Add(-24 * time.Hour)
	}

	switch request.Scope {
	case "global":
		if err := r.performanceService.BroadcastGlobalReport(ctx, since); err != nil {
			log.Error().Err(err).Msg("Failed to broadcast global performance report")
			return r.sendError(client, domain.ErrServiceUnavailable, "Failed to generate global report", "")
		}
		return nil

	case "team":
		teamName := request.TeamName
		if teamName == "" {
			teamName = client.GetTeamName()
		}

		if err := r.performanceService.SendTeamReport(ctx, teamName, since); err != nil {
			log.Error().
				Str("teamName", teamName).
				Err(err).
				Msg("Failed to send team performance report")
			return r.sendError(client, domain.ErrServiceUnavailable, "Failed to generate team report", "")
		}
		return nil

	default:
		return r.sendError(client, domain.ErrInvalidMessage, "Scope must be 'team' or 'global'", "")
	}
}

func (r *MessageRouter) sendError(client MessageClient, code, reason, clOrdID string) error {
	if client == nil {
		log.Error().
			Str("code", code).
			Str("reason", reason).
			Msg("Cannot send error: client is nil")
		return fmt.Errorf("client is nil")
	}

	errorMsg := &domain.ErrorMessage{
		Type:      "ERROR",
		Code:      code,
		Reason:    reason,
		ClOrdID:   clOrdID,
		Timestamp: time.Now(),
	}

	return client.SendMessage(errorMsg)
}

func (r *MessageRouter) handleAdminCancelAllOrders(ctx context.Context, client MessageClient) error {
	// Check if client is admin
	if client == nil || client.GetTeamName() != "admin" {
		return r.sendError(client, domain.ErrAuthFailed, "Admin access required", "")
	}

	if r.orderRepo == nil {
		log.Error().Msg("OrderRepository is nil")
		return r.sendError(client, domain.ErrServiceUnavailable, "Order service unavailable", "")
	}

	// Get all pending orders
	orders, err := r.orderRepo.GetPendingOrders(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get pending orders for cancellation")
		return r.sendError(client, domain.ErrServiceUnavailable, "Failed to get orders", "")
	}

	// Cancel all orders
	cancelledCount := 0
	for _, order := range orders {
		if order == nil {
			continue
		}
		if err := r.orderRepo.Cancel(ctx, order.ClOrdID); err != nil {
			log.Warn().
				Err(err).
				Str("clOrdID", order.ClOrdID).
				Msg("Failed to cancel order")
			continue
		}
		// Remove from order book
		if r.orderBook != nil {
			r.orderBook.RemoveOrder(order.Product, order.Side, order.ClOrdID)
		}
		cancelledCount++
	}

	log.Info().
		Str("admin", client.GetTeamName()).
		Int("cancelledCount", cancelledCount).
		Msg("Admin cancelled all orders")

	response := &domain.AdminActionResponse{
		Type:       "ADMIN_ACTION_RESPONSE",
		Action:     "CANCEL_ALL_ORDERS",
		Success:    true,
		Message:    fmt.Sprintf("Successfully cancelled %d orders", cancelledCount),
		Count:      cancelledCount,
		ServerTime: time.Now().Format(time.RFC3339),
	}

	return client.SendMessage(response)
}

func (r *MessageRouter) handleAdminCancelOrder(ctx context.Context, rawMessage string, client MessageClient) error {
	// Check if client is admin
	if client == nil || client.GetTeamName() != "admin" {
		return r.sendError(client, domain.ErrAuthFailed, "Admin access required", "")
	}

	// Parse the cancel order message
	var cancelMsg struct {
		Type     string `json:"type"`
		ClOrdID  string `json:"clOrdID"`
		TeamName string `json:"teamName"`
	}
	if err := json.Unmarshal([]byte(rawMessage), &cancelMsg); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid cancel order message format", "")
	}

	if cancelMsg.ClOrdID == "" {
		return r.sendError(client, domain.ErrInvalidMessage, "Order ID is required", "")
	}

	if r.orderRepo == nil {
		log.Error().Msg("OrderRepository is nil")
		return r.sendError(client, domain.ErrServiceUnavailable, "Order service unavailable", "")
	}

	// Get the order to verify it exists and get its details
	order, err := r.orderRepo.GetByClOrdID(ctx, cancelMsg.ClOrdID)
	if err != nil {
		log.Warn().
			Err(err).
			Str("clOrdID", cancelMsg.ClOrdID).
			Msg("Order not found for cancellation")
		return r.sendError(client, "ORDER_NOT_FOUND", "Order not found", cancelMsg.ClOrdID)
	}

	// Cancel the order
	if err := r.orderRepo.Cancel(ctx, cancelMsg.ClOrdID); err != nil {
		log.Error().
			Err(err).
			Str("clOrdID", cancelMsg.ClOrdID).
			Msg("Failed to cancel order")

		response := &domain.AdminActionResponse{
			Type:       "ADMIN_ACTION_RESPONSE",
			Action:     "CANCEL_ORDER",
			Success:    false,
			Message:    fmt.Sprintf("Failed to cancel order: %v", err),
			ServerTime: time.Now().Format(time.RFC3339),
		}
		return client.SendMessage(response)
	}

	// Remove from order book
	if r.orderBook != nil && order != nil {
		r.orderBook.RemoveOrder(order.Product, order.Side, order.ClOrdID)
	}

	log.Info().
		Str("admin", client.GetTeamName()).
		Str("clOrdID", cancelMsg.ClOrdID).
		Str("teamName", cancelMsg.TeamName).
		Msg("Admin cancelled individual order")

	// Send success response to admin
	response := &domain.AdminActionResponse{
		Type:       "ADMIN_ACTION_RESPONSE",
		Action:     "CANCEL_ORDER",
		Success:    true,
		Message:    fmt.Sprintf("Successfully cancelled order %s", cancelMsg.ClOrdID),
		Count:      1,
		ServerTime: time.Now().Format(time.RFC3339),
	}

	return client.SendMessage(response)
}

func (r *MessageRouter) handleAdminBroadcast(ctx context.Context, rawMessage string, client MessageClient) error {
	// Check if client is admin
	if client == nil || client.GetTeamName() != "admin" {
		return r.sendError(client, domain.ErrAuthFailed, "Admin access required", "")
	}

	var broadcastMsg domain.AdminBroadcastMessage
	if err := json.Unmarshal([]byte(rawMessage), &broadcastMsg); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid ADMIN_BROADCAST message format", "")
	}

	if broadcastMsg.Message == "" {
		return r.sendError(client, domain.ErrInvalidMessage, "Message is required", "")
	}

	if r.broadcaster == nil {
		log.Error().Msg("Broadcaster is nil")
		return r.sendError(client, domain.ErrServiceUnavailable, "Broadcast service unavailable", "")
	}

	// Create broadcast notification
	notification := &domain.BroadcastNotificationMessage{
		Type:       "BROADCAST_NOTIFICATION",
		Message:    broadcastMsg.Message,
		Sender:     "admin",
		ServerTime: time.Now().Format(time.RFC3339),
	}

	// Broadcast to all connected clients
	if err := r.broadcaster.BroadcastToAll(notification); err != nil {
		log.Error().Err(err).Msg("Failed to broadcast message")
		return r.sendError(client, domain.ErrServiceUnavailable, "Failed to broadcast message", "")
	}

	log.Info().
		Str("admin", client.GetTeamName()).
		Str("message", broadcastMsg.Message).
		Msg("Admin broadcast message")

	response := &domain.AdminActionResponse{
		Type:       "ADMIN_ACTION_RESPONSE",
		Action:     "BROADCAST",
		Success:    true,
		Message:    "Message broadcast to all users",
		ServerTime: time.Now().Format(time.RFC3339),
	}

	return client.SendMessage(response)
}

func (r *MessageRouter) handleSDKEmulator(ctx context.Context, rawMessage string, client MessageClient) error {
	// Parse the emulator message
	var emulatorMsg domain.SDKEmulatorMessage
	if err := json.Unmarshal([]byte(rawMessage), &emulatorMsg); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid SDK_EMULATOR message format", "")
	}

	if emulatorMsg.TargetTeam == "" {
		return r.sendError(client, domain.ErrInvalidMessage, "targetTeam is required", "")
	}

	if emulatorMsg.MessageType == "" {
		return r.sendError(client, domain.ErrInvalidMessage, "messageType is required", "")
	}

	if r.broadcaster == nil {
		log.Error().Msg("Broadcaster is nil")
		return r.sendError(client, domain.ErrServiceUnavailable, "Broadcast service unavailable", "")
	}

	// Special handling for OFFER emulation - create a real backing buy order
	if emulatorMsg.MessageType == "OFFER" {
		return r.handleSDKEmulatorOffer(ctx, emulatorMsg, client)
	}

	// Build the message to send - add the type field to the payload
	messageToSend := make(map[string]interface{})
	messageToSend["type"] = emulatorMsg.MessageType
	for k, v := range emulatorMsg.MessagePayload {
		messageToSend[k] = v
	}

	// Send the emulated message to the target team
	if err := r.broadcaster.SendToClient(emulatorMsg.TargetTeam, messageToSend); err != nil {
		log.Error().
			Err(err).
			Str("targetTeam", emulatorMsg.TargetTeam).
			Str("messageType", emulatorMsg.MessageType).
			Msg("Failed to send emulated message")
		return r.sendError(client, domain.ErrServiceUnavailable, fmt.Sprintf("Failed to send message to %s", emulatorMsg.TargetTeam), "")
	}

	log.Info().
		Str("sender", client.GetTeamName()).
		Str("targetTeam", emulatorMsg.TargetTeam).
		Str("messageType", emulatorMsg.MessageType).
		Msg("SDK Emulator message sent")

	// Send confirmation back to the emulator
	response := map[string]interface{}{
		"type":        "SDK_EMULATOR_ACK",
		"success":     true,
		"targetTeam":  emulatorMsg.TargetTeam,
		"messageType": emulatorMsg.MessageType,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	return client.SendMessage(response)
}

func (r *MessageRouter) handleSDKEmulatorOffer(ctx context.Context, emulatorMsg domain.SDKEmulatorMessage, client MessageClient) error {
	// Extract offer details from the payload
	buyer, _ := emulatorMsg.MessagePayload["buyer"].(string)
	product, _ := emulatorMsg.MessagePayload["product"].(string)
	qtyFloat, _ := emulatorMsg.MessagePayload["quantityRequested"].(float64)
	maxPriceFloat, _ := emulatorMsg.MessagePayload["maxPrice"].(float64)
	expiresInFloat, _ := emulatorMsg.MessagePayload["expiresIn"].(float64)

	qty := int(qtyFloat)
	maxPrice := maxPriceFloat
	expiresIn := int(expiresInFloat)

	if buyer == "" || product == "" || qty <= 0 || maxPrice <= 0 {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid offer payload - buyer, product, quantityRequested, and maxPrice are required", "")
	}

	// In debug mode, we don't validate team inventory for SDK emulator
	// The target team (receiver) should have inventory to accept, but we let them try
	// This is purely for SDK testing purposes

	// Create a debug buy order WITHOUT validating the buyer team
	// This is a sandbox order - the buyer team doesn't need to exist or have balance
	timestamp := time.Now().Unix()
	clOrdID := fmt.Sprintf("SDK-EMU-BUY-%s-%d", buyer, timestamp)

	// Create the order directly in the database (bypass validation)
	order := &domain.Order{
		ClOrdID:   clOrdID,
		TeamName:  buyer, // This can be a fake team name
		Side:      "BUY",
		Mode:      "LIMIT",
		Product:   product,
		Quantity:  qty,
		Price:     &maxPrice,
		Message:   "SDK Emulator debug order - for testing only",
		Status:    "PENDING",
		FilledQty: 0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save the debug order to the database
	if err := r.orderRepo.Create(ctx, order); err != nil {
		log.Error().
			Err(err).
			Str("clOrdID", clOrdID).
			Str("buyer", buyer).
			Msg("Failed to create SDK emulator debug order")
		return r.sendError(client, domain.ErrServiceUnavailable, "Failed to create debug order", clOrdID)
	}

	// Generate unique offer ID
	offerID := fmt.Sprintf("SDK-EMU-OFFER-%d-%s", timestamp, uuid.New().String()[:8])

	// Create the OFFER message to send to the target team
	offerMessage := &domain.OfferMessage{
		Type:              "OFFER",
		OfferID:           offerID,
		Buyer:             buyer,
		Product:           product,
		QuantityRequested: qty,
		MaxPrice:          maxPrice,
		Timestamp:         time.Now(),
	}

	// Add expiration if provided
	if expiresIn > 0 {
		offerMessage.ExpiresIn = &expiresIn
	}

	// Send the offer directly to the target team
	if err := r.broadcaster.SendToClient(emulatorMsg.TargetTeam, offerMessage); err != nil {
		log.Error().
			Err(err).
			Str("targetTeam", emulatorMsg.TargetTeam).
			Str("offerID", offerID).
			Msg("Failed to send SDK emulator offer to target team")
		return r.sendError(client, domain.ErrServiceUnavailable, fmt.Sprintf("Failed to send offer to %s", emulatorMsg.TargetTeam), "")
	}

	// Register the offer in the offer generator so it can be accepted
	// We need to link this offer to the buy order we created
	marketEngine, ok := r.marketService.(*market.MarketEngine)
	if !ok || marketEngine == nil || marketEngine.OfferGenerator == nil {
		log.Error().Msg("Failed to access offer generator for SDK emulator")
		return r.sendError(client, domain.ErrServiceUnavailable, "Market service unavailable", "")
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(24 * time.Hour) // Default 24 hours
	if expiresIn > 0 {
		expiresAt = time.Now().Add(time.Duration(expiresIn) * time.Millisecond)
	}

	// Store the offer in the active offers map so it can be accepted
	activeOffer := &market.ActiveOffer{
		OfferMsg:  offerMessage,
		BuyOrder:  order,
		ExpiresAt: expiresAt,
	}

	// Register the offer with the offer generator
	marketEngine.OfferGenerator.RegisterOffer(offerID, activeOffer)

	log.Info().
		Str("offerID", offerID).
		Str("clOrdID", clOrdID).
		Time("expiresAt", expiresAt).
		Msg("SDK Emulator offer registered - can be accepted for real trade")

	log.Info().
		Str("emulator", client.GetTeamName()).
		Str("buyer", buyer).
		Str("targetTeam", emulatorMsg.TargetTeam).
		Str("product", product).
		Int("qty", qty).
		Float64("maxPrice", maxPrice).
		Str("clOrdID", clOrdID).
		Str("offerID", offerID).
		Msg("SDK Emulator created debug buy order and sent offer to target team")

	// Send confirmation back to the emulator
	response := map[string]interface{}{
		"type":        "SDK_EMULATOR_ACK",
		"success":     true,
		"targetTeam":  emulatorMsg.TargetTeam,
		"messageType": "OFFER",
		"clOrdID":     clOrdID,
		"offerID":     offerID,
		"message":     fmt.Sprintf("Debug buy order created as '%s' (sandbox team). Offer sent to '%s'. Make sure %s has inventory to accept!", buyer, emulatorMsg.TargetTeam, emulatorMsg.TargetTeam),
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	return client.SendMessage(response)
}

func (r *MessageRouter) handleAdminCreateOrder(ctx context.Context, rawMessage string, client MessageClient) error {
	// Check if client is admin
	if client == nil || client.GetTeamName() != "admin" {
		return r.sendError(client, domain.ErrAuthFailed, "Admin access required", "")
	}

	var adminOrderMsg domain.AdminCreateOrderMessage
	if err := json.Unmarshal([]byte(rawMessage), &adminOrderMsg); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid ADMIN_CREATE_ORDER message format", "")
	}

	// Validate required fields
	if adminOrderMsg.TeamName == "" {
		return r.sendError(client, domain.ErrInvalidMessage, "TeamName is required", "")
	}

	// Generate order ID
	timestamp := time.Now().Unix()
	clOrdID := fmt.Sprintf("ADMIN-%s-%d", adminOrderMsg.TeamName, timestamp)

	// Create order message
	orderMsg := &domain.OrderMessage{
		Type:       "ORDER",
		ClOrdID:    clOrdID,
		Side:       adminOrderMsg.Side,
		Mode:       adminOrderMsg.Mode,
		Product:    adminOrderMsg.Product,
		Qty:        adminOrderMsg.Qty,
		LimitPrice: adminOrderMsg.LimitPrice,
		Message:    fmt.Sprintf("Admin order: %s", adminOrderMsg.Message),
	}

	// Process order using order service
	if err := r.orderService.ProcessOrder(ctx, adminOrderMsg.TeamName, orderMsg); err != nil {
		log.Warn().
			Str("admin", client.GetTeamName()).
			Str("targetTeam", adminOrderMsg.TeamName).
			Err(err).
			Msg("Admin order creation failed")
		return r.sendError(client, domain.ErrInvalidOrder, err.Error(), clOrdID)
	}

	log.Info().
		Str("admin", client.GetTeamName()).
		Str("targetTeam", adminOrderMsg.TeamName).
		Str("clOrdID", clOrdID).
		Msg("Admin created order")

	ackMsg := &domain.OrderAckMessage{
		Type:       "ORDER_ACK",
		ClOrdID:    clOrdID,
		Status:     "PENDING",
		ServerTime: time.Now().Format(time.RFC3339),
	}
	if r.broadcaster != nil {
		_ = r.broadcaster.SendToClient("admin", ackMsg)
	}

	response := &domain.AdminActionResponse{
		Type:       "ADMIN_ACTION_RESPONSE",
		Action:     "CREATE_ORDER",
		Success:    true,
		Message:    fmt.Sprintf("Order %s created for team %s", clOrdID, adminOrderMsg.TeamName),
		ServerTime: time.Now().Format(time.RFC3339),
	}

	return client.SendMessage(response)
}

func (r *MessageRouter) handleExportData(ctx context.Context, rawMessage string, client MessageClient) error {
	// Check if client is admin
	if client == nil || client.GetTeamName() != "admin" {
		return r.sendError(client, domain.ErrAuthFailed, "Admin access required", "")
	}

	var exportMsg domain.ExportDataMessage
	if err := json.Unmarshal([]byte(rawMessage), &exportMsg); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid EXPORT_DATA message format", "")
	}

	dataType := exportMsg.DataType
	if dataType == "" {
		dataType = "all"
	}

	var data interface{}
	var count int

	switch dataType {
	case "orders", "all":
		if r.orderRepo == nil {
			return r.sendError(client, domain.ErrServiceUnavailable, "Order repository unavailable", "")
		}
		orders, err := r.orderRepo.GetPendingOrders(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to export orders")
			return r.sendError(client, domain.ErrServiceUnavailable, "Failed to export orders", "")
		}
		data = orders
		count = len(orders)

	case "sessions":
		authService, ok := r.authService.(*service.AuthService)
		if !ok {
			return r.sendError(client, domain.ErrServiceUnavailable, "Auth service unavailable", "")
		}
		sessions := authService.GetActiveSessions()
		data = sessions
		count = 0
		for _, teamSessions := range sessions {
			count += len(teamSessions)
		}

	default:
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid dataType", "")
	}

	log.Info().
		Str("admin", client.GetTeamName()).
		Str("dataType", dataType).
		Int("count", count).
		Msg("Admin exported data")

	response := &domain.ExportDataResponse{
		Type:       "EXPORT_DATA_RESPONSE",
		DataType:   dataType,
		Data:       data,
		Count:      count,
		ServerTime: time.Now().Format(time.RFC3339),
	}

	return client.SendMessage(response)
}

func (r *MessageRouter) handleGetAvailableTeams(ctx context.Context, client MessageClient) error {
	if client == nil || client.GetTeamName() != "admin" {
		return r.sendError(client, domain.ErrAuthFailed, "Admin access required", "")
	}

	if r.orderRepo == nil {
		return r.sendError(client, domain.ErrServiceUnavailable, "Order repository unavailable", "")
	}

	authService, ok := r.authService.(*service.AuthService)
	if !ok {
		return r.sendError(client, domain.ErrServiceUnavailable, "Auth service unavailable", "")
	}

	activeSessions := authService.GetActiveSessions()
	connectedTeams := make(map[string]bool)
	for teamName := range activeSessions {
		connectedTeams[teamName] = true
	}

	teams := make([]domain.TeamInfo, 0)

	allOrders, err := r.orderRepo.GetPendingOrders(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get orders for team info")
	}

	orderCounts := make(map[string]int)
	for _, order := range allOrders {
		if order != nil {
			orderCounts[order.TeamName]++
		}
	}

	for teamName := range connectedTeams {
		teams = append(teams, domain.TeamInfo{
			TeamName:     teamName,
			Connected:    true,
			ActiveOrders: orderCounts[teamName],
		})
	}

	response := &domain.AvailableTeamsResponse{
		Type:       "AVAILABLE_TEAMS",
		Teams:      teams,
		Count:      len(teams),
		ServerTime: time.Now().Format(time.RFC3339),
	}

	log.Info().
		Str("admin", client.GetTeamName()).
		Int("teamCount", len(teams)).
		Msg("Admin requested available teams")

	return client.SendMessage(response)
}

func (r *MessageRouter) handleGetTeamActivity(ctx context.Context, rawMessage string, client MessageClient) error {
	if client == nil || client.GetTeamName() != "admin" {
		return r.sendError(client, domain.ErrAuthFailed, "Admin access required", "")
	}

	var activityMsg domain.GetTeamActivityMessage
	if err := json.Unmarshal([]byte(rawMessage), &activityMsg); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid GET_TEAM_ACTIVITY message format", "")
	}

	teamName := activityMsg.TeamName
	if teamName == "" {
		return r.sendError(client, domain.ErrInvalidMessage, "TeamName is required", "")
	}

	activities := make([]domain.ActivityRecord, 0)

	if r.orderRepo != nil {
		orders, err := r.orderRepo.GetPendingOrders(ctx)
		if err == nil {
			for _, order := range orders {
				if order != nil && order.TeamName == teamName {
					price := 0.0
					if order.Price != nil {
						price = *order.Price
					}
					activities = append(activities, domain.ActivityRecord{
						Timestamp: order.CreatedAt.Format(time.RFC3339),
						Action:    "ORDER",
						Details:   fmt.Sprintf("%s %s %d %s", order.Mode, order.Side, order.Quantity, order.Product),
						Product:   order.Product,
						Quantity:  order.Quantity,
						Price:     price,
					})
				}
			}
		}
	}

	response := &domain.TeamActivityResponse{
		Type:       "TEAM_ACTIVITY",
		TeamName:   teamName,
		Activities: activities,
		Count:      len(activities),
		ServerTime: time.Now().Format(time.RFC3339),
	}

	log.Info().
		Str("admin", client.GetTeamName()).
		Str("targetTeam", teamName).
		Int("activityCount", len(activities)).
		Msg("Admin requested team activity")

	return client.SendMessage(response)
}

func (r *MessageRouter) handleGetAllTeams(ctx context.Context, client MessageClient) error {
	if client == nil || client.GetTeamName() != "admin" {
		return r.sendError(client, domain.ErrAuthFailed, "Admin access required", "")
	}

	teamRepo, ok := r.authService.(*service.AuthService)
	if !ok {
		return r.sendError(client, domain.ErrServiceUnavailable, "Team service unavailable", "")
	}

	allTeams, err := teamRepo.GetAllTeams(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all teams")
		return r.sendError(client, domain.ErrServiceUnavailable, "Failed to get teams", "")
	}

	activeSessions := teamRepo.GetActiveSessions()
	teamsData := make([]*domain.TeamData, 0, len(allTeams))

	for _, team := range allTeams {
		if team == nil {
			continue
		}
		_, connected := activeSessions[team.TeamName]
		teamsData = append(teamsData, &domain.TeamData{
			TeamName:           team.TeamName,
			Species:            team.Species,
			Members:            team.Members,
			InitialBalance:     team.InitialBalance,
			CurrentBalance:     team.CurrentBalance,
			Inventory:          team.Inventory,
			AuthorizedProducts: team.AuthorizedProducts,
			Connected:          connected,
		})
	}

	response := &domain.AllTeamsResponse{
		Type:       "ALL_TEAMS",
		Teams:      teamsData,
		Count:      len(teamsData),
		ServerTime: time.Now().Format(time.RFC3339),
	}

	log.Info().
		Str("admin", client.GetTeamName()).
		Int("teamCount", len(teamsData)).
		Msg("Admin requested all teams")

	return client.SendMessage(response)
}

func (r *MessageRouter) handleUpdateTeam(ctx context.Context, rawMessage string, client MessageClient) error {
	if client == nil || client.GetTeamName() != "admin" {
		return r.sendError(client, domain.ErrAuthFailed, "Admin access required", "")
	}

	var updateMsg domain.UpdateTeamMessage
	if err := json.Unmarshal([]byte(rawMessage), &updateMsg); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid UPDATE_TEAM message format", "")
	}

	if updateMsg.TeamName == "" {
		return r.sendError(client, domain.ErrInvalidMessage, "Team name is required", "")
	}

	teamRepo, ok := r.authService.(*service.AuthService)
	if !ok {
		return r.sendError(client, domain.ErrServiceUnavailable, "Team service unavailable", "")
	}

	err := teamRepo.UpdateTeam(ctx, updateMsg.TeamName, updateMsg.Balance, updateMsg.Inventory)
	if err != nil {
		log.Error().Err(err).Str("team", updateMsg.TeamName).Msg("Failed to update team")
		return r.sendError(client, domain.ErrServiceUnavailable, "Failed to update team", "")
	}

	response := &domain.TeamUpdatedResponse{
		Type:       "TEAM_UPDATED",
		Success:    true,
		Message:    fmt.Sprintf("Team %s updated successfully", updateMsg.TeamName),
		ServerTime: time.Now().Format(time.RFC3339),
	}

	log.Info().
		Str("admin", client.GetTeamName()).
		Str("targetTeam", updateMsg.TeamName).
		Float64("balance", updateMsg.Balance).
		Msg("Admin updated team")

	return client.SendMessage(response)
}

func (r *MessageRouter) handleUpdateTeamMembers(ctx context.Context, rawMessage string, client MessageClient) error {
	if client == nil || client.GetTeamName() != "admin" {
		return r.sendError(client, domain.ErrAuthFailed, "Admin access required", "")
	}

	var updateMsg domain.UpdateTeamMembersMessage
	if err := json.Unmarshal([]byte(rawMessage), &updateMsg); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid UPDATE_TEAM_MEMBERS message format", "")
	}

	if updateMsg.TeamName == "" {
		return r.sendError(client, domain.ErrInvalidMessage, "Team name is required", "")
	}

	// Get team repository
	teamRepo, ok := r.authService.(*service.AuthService)
	if !ok {
		return r.sendError(client, domain.ErrServiceUnavailable, "Team service unavailable", "")
	}

	// Access the underlying repository to update members
	if err := teamRepo.UpdateTeamMembers(ctx, updateMsg.TeamName, updateMsg.Members); err != nil {
		log.Error().Err(err).Str("team", updateMsg.TeamName).Msg("Failed to update team members")
		return r.sendError(client, domain.ErrServiceUnavailable, "Failed to update team members", "")
	}

	response := &domain.TeamUpdatedResponse{
		Type:       "TEAM_MEMBERS_UPDATED",
		Success:    true,
		Message:    fmt.Sprintf("Team %s members updated successfully", updateMsg.TeamName),
		ServerTime: time.Now().Format(time.RFC3339),
	}

	log.Info().
		Str("admin", client.GetTeamName()).
		Str("targetTeam", updateMsg.TeamName).
		Str("members", updateMsg.Members).
		Msg("Admin updated team members")

	return client.SendMessage(response)
}

func (r *MessageRouter) handleResetTeamBalance(ctx context.Context, rawMessage string, client MessageClient) error {
	if client == nil || client.GetTeamName() != "admin" {
		return r.sendError(client, domain.ErrAuthFailed, "Admin access required", "")
	}

	var resetMsg domain.ResetTeamBalanceMessage
	if err := json.Unmarshal([]byte(rawMessage), &resetMsg); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid RESET_TEAM_BALANCE message format", "")
	}

	if resetMsg.TeamName == "" {
		return r.sendError(client, domain.ErrInvalidMessage, "Team name is required", "")
	}

	teamRepo, ok := r.authService.(*service.AuthService)
	if !ok {
		return r.sendError(client, domain.ErrServiceUnavailable, "Team service unavailable", "")
	}

	err := teamRepo.ResetTeamBalance(ctx, resetMsg.TeamName)
	if err != nil {
		log.Error().Err(err).Str("team", resetMsg.TeamName).Msg("Failed to reset team balance")
		return r.sendError(client, domain.ErrServiceUnavailable, "Failed to reset balance", "")
	}

	response := &domain.TeamUpdatedResponse{
		Type:       "TEAM_UPDATED",
		Success:    true,
		Message:    fmt.Sprintf("Team %s balance reset successfully", resetMsg.TeamName),
		ServerTime: time.Now().Format(time.RFC3339),
	}

	log.Info().
		Str("admin", client.GetTeamName()).
		Str("targetTeam", resetMsg.TeamName).
		Msg("Admin reset team balance")

	return client.SendMessage(response)
}

func (r *MessageRouter) handleResetTeamInventory(ctx context.Context, rawMessage string, client MessageClient) error {
	if client == nil || client.GetTeamName() != "admin" {
		return r.sendError(client, domain.ErrAuthFailed, "Admin access required", "")
	}

	var resetMsg domain.ResetTeamInventoryMessage
	if err := json.Unmarshal([]byte(rawMessage), &resetMsg); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid RESET_TEAM_INVENTORY message format", "")
	}

	if resetMsg.TeamName == "" {
		return r.sendError(client, domain.ErrInvalidMessage, "Team name is required", "")
	}

	teamRepo, ok := r.authService.(*service.AuthService)
	if !ok {
		return r.sendError(client, domain.ErrServiceUnavailable, "Team service unavailable", "")
	}

	err := teamRepo.ResetTeamInventory(ctx, resetMsg.TeamName)
	if err != nil {
		log.Error().Err(err).Str("team", resetMsg.TeamName).Msg("Failed to reset team inventory")
		return r.sendError(client, domain.ErrServiceUnavailable, "Failed to reset inventory", "")
	}

	response := &domain.TeamUpdatedResponse{
		Type:       "TEAM_UPDATED",
		Success:    true,
		Message:    fmt.Sprintf("Team %s inventory reset successfully", resetMsg.TeamName),
		ServerTime: time.Now().Format(time.RFC3339),
	}

	log.Info().
		Str("admin", client.GetTeamName()).
		Str("targetTeam", resetMsg.TeamName).
		Msg("Admin reset team inventory")

	return client.SendMessage(response)
}

func (r *MessageRouter) handleResetTeamProduction(ctx context.Context, rawMessage string, client MessageClient) error {
	if client == nil || client.GetTeamName() != "admin" {
		return r.sendError(client, domain.ErrAuthFailed, "Admin access required", "")
	}

	var resetMsg domain.ResetTeamProductionMessage
	if err := json.Unmarshal([]byte(rawMessage), &resetMsg); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid RESET_TEAM_PRODUCTION message format", "")
	}

	if resetMsg.TeamName == "" {
		return r.sendError(client, domain.ErrInvalidMessage, "Team name is required", "")
	}

	response := &domain.TeamUpdatedResponse{
		Type:       "TEAM_UPDATED",
		Success:    true,
		Message:    "Production reset is not yet implemented",
		ServerTime: time.Now().Format(time.RFC3339),
	}

	log.Info().
		Str("admin", client.GetTeamName()).
		Str("targetTeam", resetMsg.TeamName).
		Msg("Admin requested production reset (not implemented)")

	return client.SendMessage(response)
}

func (r *MessageRouter) handleResetTournamentConfig(
	ctx context.Context,
	rawMessage string,
	client MessageClient,
) error {
	if client == nil || client.GetTeamName() != "admin" {
		return r.sendError(client, domain.ErrAuthFailed, "Admin access required", "")
	}

	var tournamentMsg domain.ResetTournamentConfigMessage
	if err := json.Unmarshal([]byte(rawMessage), &tournamentMsg); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid RESET_TOURNAMENT_CONFIG message format", "")
	}

	if tournamentMsg.Balance < 0 {
		return r.sendError(client, domain.ErrInvalidMessage, "Balance cannot be negative", "")
	}

	if len(tournamentMsg.TeamConfigs) == 0 {
		return r.sendError(client, domain.ErrInvalidMessage, "No team configurations provided", "")
	}

	teamRepo, ok := r.authService.(*service.AuthService)
	if !ok {
		return r.sendError(client, domain.ErrServiceUnavailable, "Team service unavailable", "")
	}

	// Delete all fill history to start fresh
	if r.fillRepo != nil {
		if err := r.fillRepo.DeleteAll(ctx); err != nil {
			log.Warn().Err(err).Msg("Failed to delete fills during tournament reset")
		} else {
			log.Info().Msg("All fill history cleared for tournament reset")
		}
	}

	// Cancel all active orders
	ordersCanceled := 0
	if r.orderRepo != nil {
		allOrders, err := r.orderRepo.GetPendingOrders(ctx)
		if err == nil {
			for _, order := range allOrders {
				if order != nil {
					if err := r.orderRepo.Cancel(ctx, order.ClOrdID); err != nil {
						log.Warn().
							Err(err).
							Str("clOrdID", order.ClOrdID).
							Msg("Failed to cancel order during tournament reset")
					} else {
						ordersCanceled++
						// Remove from order book
						if r.orderBook != nil {
							r.orderBook.RemoveOrder(order.Product, order.Side, order.ClOrdID)
						}
					}
				}
			}
		}
	}

	// Reset each team
	teamsReset := 0
	for _, teamConfig := range tournamentMsg.TeamConfigs {
		if teamConfig.TeamName == "" || teamConfig.TeamName == "admin" {
			continue
		}

		// CRITICAL: Set InitialBalance FIRST to ensure it matches the reset balance
		// This ensures ROI starts at 0% for all teams
		if err := teamRepo.UpdateInitialBalance(ctx, teamConfig.TeamName, tournamentMsg.Balance); err != nil {
			log.Warn().Err(err).Str("team", teamConfig.TeamName).Msg("Failed to update initial balance")
			continue
		}

		// Reset inventory to zero first
		if err := teamRepo.ResetTeamInventory(ctx, teamConfig.TeamName); err != nil {
			log.Warn().Err(err).Str("team", teamConfig.TeamName).Msg("Failed to reset inventory")
			continue
		}

		// Set inventory from config if provided
		if len(teamConfig.Inventory) > 0 {
			if err := teamRepo.UpdateTeam(ctx, teamConfig.TeamName, tournamentMsg.Balance, teamConfig.Inventory); err != nil {
				log.Warn().Err(err).Str("team", teamConfig.TeamName).Msg("Failed to set team inventory")
				continue
			}
		} else {
			// Update balance with empty inventory
			if err := teamRepo.UpdateTeam(ctx, teamConfig.TeamName, tournamentMsg.Balance, nil); err != nil {
				log.Warn().Err(err).Str("team", teamConfig.TeamName).Msg("Failed to set tournament balance")
				continue
			}
		}

		teamsReset++

		log.Info().
			Str("team", teamConfig.TeamName).
			Interface("inventory", teamConfig.Inventory).
			Float64("initialBalance", tournamentMsg.Balance).
			Float64("currentBalance", tournamentMsg.Balance).
			Msg("Team reset for tournament - ROI should be 0%")
	}

	response := &domain.TournamentResetCompleteResponse{
		Type:           "TOURNAMENT_RESET_COMPLETE",
		Success:        true,
		TeamsReset:     teamsReset,
		OrdersCanceled: ordersCanceled,
		Message: fmt.Sprintf(
			"Tournament reset: %d teams configured, %d orders cancelled",
			teamsReset,
			ordersCanceled,
		),
		ServerTime: time.Now().Format(time.RFC3339),
	}

	log.Info().
		Str("admin", client.GetTeamName()).
		Int("teamsReset", teamsReset).
		Int("ordersCanceled", ordersCanceled).
		Float64("balance", tournamentMsg.Balance).
		Msg("Tournament configuration reset completed")

	// Broadcast updated balance and inventory to all teams
	// This ensures clients see the reset immediately
	for _, teamConfig := range tournamentMsg.TeamConfigs {
		if teamConfig.TeamName == "" || teamConfig.TeamName == "admin" {
			continue
		}

		// Send balance update
		balanceMsg := &domain.BalanceUpdateMessage{
			Type:       "BALANCE_UPDATE",
			Balance:    tournamentMsg.Balance,
			ServerTime: time.Now().Format(time.RFC3339),
		}
		if err := r.broadcaster.SendToClient(teamConfig.TeamName, balanceMsg); err != nil {
			log.Debug().Err(err).Str("team", teamConfig.TeamName).Msg("Failed to broadcast balance update after tournament reset")
		}

		// Send inventory update
		inventoryMsg := &domain.InventoryUpdateMessage{
			Type:       "INVENTORY_UPDATE",
			Inventory:  teamConfig.Inventory,
			ServerTime: time.Now().Format(time.RFC3339),
		}
		if err := r.broadcaster.SendToClient(teamConfig.TeamName, inventoryMsg); err != nil {
			log.Debug().Err(err).Str("team", teamConfig.TeamName).Msg("Failed to broadcast inventory update after tournament reset")
		}
	}

	return client.SendMessage(response)
}

func (r *MessageRouter) handleUpdateAllRecipes(
	ctx context.Context,
	client MessageClient,
) error {
	if client == nil || client.GetTeamName() != "admin" {
		return r.sendError(client, domain.ErrAuthFailed, "Admin access required", "")
	}

	teamRepo, ok := r.authService.(*service.AuthService)
	if !ok {
		return r.sendError(client, domain.ErrServiceUnavailable, "Team service unavailable", "")
	}

	// Get all teams
	teams, err := teamRepo.GetAllTeams(ctx)
	if err != nil {
		return r.sendError(client, domain.ErrServiceUnavailable, "Failed to get teams", "")
	}

	// Update recipes for each team based on species
	teamsUpdated := 0
	for _, team := range teams {
		if team.TeamName == "admin" {
			continue
		}

		// Get the basic product (first authorized product)
		basicProduct := ""
		if len(team.AuthorizedProducts) > 0 {
			basicProduct = team.AuthorizedProducts[0]
		} else {
			log.Warn().Str("team", team.TeamName).Msg("Team has no authorized products, skipping")
			continue
		}

		// Build recipes based on species
		recipes := buildRecipesForSpecies(team.Species, basicProduct)

		// Update team recipes
		if err := teamRepo.UpdateRecipes(ctx, team.TeamName, recipes); err != nil {
			log.Warn().
				Err(err).
				Str("team", team.TeamName).
				Msg("Failed to update team recipes")
			continue
		}

		teamsUpdated++
		log.Info().
			Str("team", team.TeamName).
			Str("species", team.Species).
			Int("recipeCount", len(recipes)).
			Msg("Team recipes updated")
	}

	// Send response
	response := map[string]interface{}{
		"type":         "RECIPES_UPDATED",
		"success":      true,
		"teamsUpdated": teamsUpdated,
		"message":      fmt.Sprintf("Updated recipes for %d teams", teamsUpdated),
		"serverTime":   time.Now().Format(time.RFC3339),
	}

	log.Info().
		Str("admin", client.GetTeamName()).
		Int("teamsUpdated", teamsUpdated).
		Msg("All team recipes updated")

	return client.SendMessage(response)
}

// buildRecipesForSpecies creates all recipes (basic + premium) for a species
func buildRecipesForSpecies(species string, basicProduct string) map[string]domain.Recipe {
	recipes := make(map[string]domain.Recipe)

	// Add basic recipe (free production)
	recipes[basicProduct] = domain.Recipe{
		Type:         "BASIC",
		Ingredients:  map[string]int{},
		PremiumBonus: 1.0,
	}

	// Add premium recipes based on species (30% bonus)
	switch species {
	case "Avocultores":
		recipes["GUACA"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"FOSFO": 5, "PITA": 3},
			PremiumBonus: 1.3,
		}
		recipes["SEBO"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"NUCREM": 8},
			PremiumBonus: 1.3,
		}

	case "Monjes de Fosforescencia":
		recipes["GUACA"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"PALTA-OIL": 5, "PITA": 3},
			PremiumBonus: 1.3,
		}
		recipes["NUCREM"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"SEBO": 6},
			PremiumBonus: 1.3,
		}

	case "Cosechadores de Pita":
		recipes["SEBO"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"NUCREM": 8},
			PremiumBonus: 1.3,
		}
		recipes["CASCAR-ALLOY"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"FOSFO": 10},
			PremiumBonus: 1.3,
		}

	case "Herreros Cósmicos":
		recipes["QUANTUM-PULP"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"PALTA-OIL": 7},
			PremiumBonus: 1.3,
		}
		recipes["SKIN-WRAP"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"ASTRO-BUTTER": 12},
			PremiumBonus: 1.3,
		}

	case "Extractores":
		recipes["NUCREM"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"SEBO": 6},
			PremiumBonus: 1.3,
		}
		recipes["FOSFO"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"SKIN-WRAP": 9},
			PremiumBonus: 1.3,
		}

	case "Tejemanteles":
		recipes["PITA"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"CASCAR-ALLOY": 8},
			PremiumBonus: 1.3,
		}
		recipes["ASTRO-BUTTER"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"GUACA": 10},
			PremiumBonus: 1.3,
		}

	case "Cremeros Astrales":
		recipes["CASCAR-ALLOY"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"FOSFO": 10},
			PremiumBonus: 1.3,
		}
		recipes["PALTA-OIL"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"QUANTUM-PULP": 7},
			PremiumBonus: 1.3,
		}

	case "Mineros del Sebo":
		recipes["ASTRO-BUTTER"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"GUACA": 10},
			PremiumBonus: 1.3,
		}
		recipes["GUACA"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"PALTA-OIL": 5, "PITA": 3},
			PremiumBonus: 1.3,
		}

	case "Núcleo Cremero":
		recipes["SKIN-WRAP"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"ASTRO-BUTTER": 12},
			PremiumBonus: 1.3,
		}
		recipes["QUANTUM-PULP"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"PALTA-OIL": 7},
			PremiumBonus: 1.3,
		}

	case "Destiladores":
		recipes["PALTA-OIL"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"QUANTUM-PULP": 7},
			PremiumBonus: 1.3,
		}
		recipes["FOSFO"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"SKIN-WRAP": 9},
			PremiumBonus: 1.3,
		}

	case "Cartógrafos":
		recipes["NUCREM"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"SEBO": 6},
			PremiumBonus: 1.3,
		}
		recipes["PITA"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"CASCAR-ALLOY": 8},
			PremiumBonus: 1.3,
		}

	case "Someliers Andorianos":
		recipes["SEBO"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"NUCREM": 8},
			PremiumBonus: 1.3,
		}
		recipes["CASCAR-ALLOY"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"FOSFO": 10},
			PremiumBonus: 1.3,
		}
	}

	return recipes
}
