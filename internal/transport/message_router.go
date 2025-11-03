package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/HellSoft-Col/stock-market/internal/market"
	"github.com/HellSoft-Col/stock-market/internal/service"
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
	case "REQUEST_ORDER_BOOK":
		return r.handleRequestOrderBook(ctx, rawMessage, client)
	case "REQUEST_CONNECTED_SESSIONS":
		return r.handleRequestConnectedSessions(ctx, client)
	case "REQUEST_PERFORMANCE_REPORT":
		return r.handleRequestPerformanceReport(ctx, rawMessage, client)
	case "PING":
		return r.handlePing(ctx, client)
	case "ADMIN_CANCEL_ALL_ORDERS":
		return r.handleAdminCancelAllOrders(ctx, client)
	case "ADMIN_BROADCAST":
		return r.handleAdminBroadcast(ctx, rawMessage, client)
	case "ADMIN_CREATE_ORDER":
		return r.handleAdminCreateOrder(ctx, rawMessage, client)
	case "EXPORT_DATA":
		return r.handleExportData(ctx, rawMessage, client)
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

	// Get team inventory
	inventory := team.Inventory
	if inventory == nil {
		inventory = make(map[string]int)
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
		Role:               team.Role,
		ServerTime:         time.Now().Format(time.RFC3339),
	}

	log.Info().
		Str("teamName", team.TeamName).
		Str("species", team.Species).
		Str("clientAddr", client.GetRemoteAddr()).
		Msg("Team logged in successfully")

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
		return r.sendError(client, domain.ErrInvalidMessage, "Limit price must be positive for LIMIT orders", orderMsg.ClOrdID)
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

func (r *MessageRouter) handleRequestPerformanceReport(ctx context.Context, rawMessage string, client MessageClient) error {
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
