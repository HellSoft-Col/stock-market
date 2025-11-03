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
	if client.GetTeamName() == "" {
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
	if client.GetTeamName() == "" {
		return r.sendError(client, domain.ErrAuthFailed, "Must login first", "")
	}

	// Parse message
	var prodMsg domain.ProductionUpdateMessage
	if err := json.Unmarshal([]byte(rawMessage), &prodMsg); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid PRODUCTION_UPDATE message format", "")
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
	if client.GetTeamName() == "" {
		return r.sendError(client, domain.ErrAuthFailed, "Must login first", "")
	}

	// Parse message
	var acceptMsg domain.AcceptOfferMessage
	if err := json.Unmarshal([]byte(rawMessage), &acceptMsg); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid ACCEPT_OFFER message format", "")
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
	if client.GetTeamName() == "" {
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
	if client.GetTeamName() == "" {
		return r.sendError(client, domain.ErrAuthFailed, "Must login first", "")
	}

	var cancelMsg domain.CancelMessage
	if err := json.Unmarshal([]byte(rawMessage), &cancelMsg); err != nil {
		return r.sendError(client, domain.ErrInvalidMessage, "Invalid CANCEL message format", "")
	}

	if cancelMsg.ClOrdID == "" {
		return r.sendError(client, domain.ErrInvalidMessage, "ClOrdID is required", "")
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
	if client.GetTeamName() == "" {
		return r.sendError(client, domain.ErrAuthFailed, "Must login first", "")
	}

	orders, err := r.orderRepo.GetPendingOrders(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get pending orders")
		return r.sendError(client, domain.ErrServiceUnavailable, "Failed to get orders", "")
	}

	orderSummaries := make([]*domain.OrderSummary, 0, len(orders))
	for _, order := range orders {
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
	if client.GetTeamName() == "" {
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
	if client.GetTeamName() == "" {
		return r.sendError(client, domain.ErrAuthFailed, "Must login first", "")
	}

	// Get detailed session information from auth service
	authService, ok := r.authService.(*service.AuthService)
	if !ok {
		// Fallback to simple broadcaster info
		return r.handleRequestConnectedSessionsFallback(ctx, client)
	}

	activeSessions := authService.GetActiveSessions()

	sessions := make([]*domain.SessionInfo, 0)
	totalConnections := 0

	for _, teamSessions := range activeSessions {
		for _, session := range teamSessions {
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
	if client.GetTeamName() == "" {
		return r.sendError(client, domain.ErrAuthFailed, "Must login first", "")
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
	errorMsg := &domain.ErrorMessage{
		Type:      "ERROR",
		Code:      code,
		Reason:    reason,
		ClOrdID:   clOrdID,
		Timestamp: time.Now(),
	}

	return client.SendMessage(errorMsg)
}
