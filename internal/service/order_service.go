package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/HellSoft-Col/stock-market/internal/domain"
)

var validProducts = map[string]bool{
	"GUACA":        true,
	"SEBO":         true,
	"PALTA-OIL":    true,
	"FOSFO":        true,
	"NUCREM":       true,
	"CASCAR-ALLOY": true,
	"GTRON":        true,
	"H-GUACA":      true,
	"PITA":         true,
}

type OrderService struct {
	orderRepo   domain.OrderRepository
	marketSvc   domain.MarketService
	broadcaster domain.Broadcaster
}

func NewOrderService(orderRepo domain.OrderRepository, marketSvc domain.MarketService, broadcaster domain.Broadcaster) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		marketSvc:   marketSvc,
		broadcaster: broadcaster,
	}
}

func (s *OrderService) ProcessOrder(ctx context.Context, teamName string, orderMsg *domain.OrderMessage) error {
	// Validate order message
	if err := s.validateOrder(orderMsg); err != nil {
		return err
	}

	// Check for duplicate order ID
	if _, err := s.orderRepo.GetByClOrdID(ctx, orderMsg.ClOrdID); err == nil {
		return fmt.Errorf("duplicate order ID: %s", orderMsg.ClOrdID)
	} else if err != domain.ErrOrderNotFound {
		return fmt.Errorf("failed to check order ID: %w", err)
	}

	// Set default price for MARKET orders
	var price *float64
	if orderMsg.Mode == "LIMIT" && orderMsg.LimitPrice != nil {
		price = orderMsg.LimitPrice
	}

	// Parse expiration time if provided
	var expiresAt *time.Time
	if orderMsg.ExpiresAt != nil && *orderMsg.ExpiresAt != "" {
		if parsedTime, err := time.Parse(time.RFC3339, *orderMsg.ExpiresAt); err == nil {
			expiresAt = &parsedTime
		}
	}

	// Create order entity
	order := &domain.Order{
		ClOrdID:    orderMsg.ClOrdID,
		TeamName:   teamName,
		Side:       orderMsg.Side,
		Mode:       orderMsg.Mode,
		Product:    orderMsg.Product,
		Quantity:   orderMsg.Qty,
		Price:      price,
		LimitPrice: orderMsg.LimitPrice,
		Message:    orderMsg.Message,
		DebugMode:  orderMsg.DebugMode,
		ExpiresAt:  expiresAt,
	}

	// Save order to database as PENDING
	if err := s.orderRepo.Create(ctx, order); err != nil {
		log.Error().
			Str("clOrdID", order.ClOrdID).
			Str("teamName", teamName).
			Err(err).
			Msg("Failed to create order")
		return fmt.Errorf("failed to save order: %w", err)
	}

	log.Info().
		Str("clOrdID", order.ClOrdID).
		Str("teamName", teamName).
		Str("side", order.Side).
		Str("mode", order.Mode).
		Str("product", order.Product).
		Int("qty", order.Quantity).
		Msg("Order created successfully")

	// Send acknowledgment
	ackMsg := &domain.OrderAckMessage{
		Type:       "ORDER_ACK",
		ClOrdID:    order.ClOrdID,
		Status:     "PENDING",
		ServerTime: time.Now().Format(time.RFC3339),
	}

	if s.broadcaster != nil {
		if err := s.broadcaster.SendToClient(teamName, ackMsg); err != nil {
			log.Warn().
				Str("clOrdID", order.ClOrdID).
				Str("teamName", teamName).
				Err(err).
				Msg("Failed to send order acknowledgment")
		}
	}

	// Send order to market engine
	s.marketSvc.ProcessOrder(order, nil) // Client connection will be added in Phase 7

	return nil
}

func (s *OrderService) validateOrder(orderMsg *domain.OrderMessage) error {
	// Required fields
	if orderMsg.ClOrdID == "" {
		return fmt.Errorf("clOrdID is required")
	}
	if orderMsg.Side != "BUY" && orderMsg.Side != "SELL" {
		return fmt.Errorf("side must be BUY or SELL")
	}
	if orderMsg.Mode != "MARKET" && orderMsg.Mode != "LIMIT" {
		return fmt.Errorf("mode must be MARKET or LIMIT")
	}
	if orderMsg.Product == "" {
		return fmt.Errorf("product is required")
	}
	if orderMsg.Qty <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	// Validate product
	if !validProducts[orderMsg.Product] {
		return fmt.Errorf("invalid product: %s", orderMsg.Product)
	}

	// LIMIT orders must have a price
	if orderMsg.Mode == "LIMIT" && (orderMsg.LimitPrice == nil || *orderMsg.LimitPrice <= 0) {
		return fmt.Errorf("LIMIT orders must have a positive limitPrice")
	}

	// Message length limit
	if len(orderMsg.Message) > 200 {
		return fmt.Errorf("message too long (max 200 characters)")
	}

	return nil
}

// Generate unique order ID for testing
func GenerateOrderID(teamName string) string {
	timestamp := time.Now().Unix()
	shortUUID := uuid.New().String()[:8]
	return fmt.Sprintf("ORD-%s-%d-%s", teamName, timestamp, shortUUID)
}

var _ domain.OrderService = (*OrderService)(nil)
