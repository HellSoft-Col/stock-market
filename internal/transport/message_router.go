package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/yourusername/avocado-exchange-server/internal/domain"
	"github.com/yourusername/avocado-exchange-server/internal/market"
	"github.com/yourusername/avocado-exchange-server/internal/service"
)

type MessageRouter struct {
	authService       domain.AuthService
	orderService      domain.OrderService
	broadcaster       domain.Broadcaster
	marketService     domain.MarketService
	resyncService     domain.ResyncService
	productionService domain.ProductionService
	rateLimiter       *service.RateLimiter
}

func NewMessageRouter(
	authService domain.AuthService,
	orderService domain.OrderService,
	broadcaster domain.Broadcaster,
	marketService domain.MarketService,
	resyncService domain.ResyncService,
	productionService domain.ProductionService,
	rateLimiter *service.RateLimiter,
) *MessageRouter {
	return &MessageRouter{
		authService:       authService,
		orderService:      orderService,
		broadcaster:       broadcaster,
		marketService:     marketService,
		resyncService:     resyncService,
		productionService: productionService,
		rateLimiter:       rateLimiter,
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
	default:
		// For unimplemented messages, echo back for now
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

	// Set client team name and register
	client.SetTeamName(team.TeamName)
	client.RegisterWithServer(team.TeamName)
	r.broadcaster.RegisterClient(team.TeamName, client)

	// Generate and send LOGIN_OK response
	loginOKMsg := &domain.LoginOKMessage{
		Type:               "LOGIN_OK",
		Team:               team.TeamName,
		Species:            team.Species,
		InitialBalance:     team.InitialBalance,
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
