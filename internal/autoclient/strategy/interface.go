package strategy

import (
	"context"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/autoclient/market"
	"github.com/HellSoft-Col/stock-market/internal/domain"
)

// Strategy defines the interface for trading strategies
type Strategy interface {
	// Name returns the strategy name
	Name() string

	// Initialize initializes the strategy with configuration
	Initialize(config map[string]interface{}) error

	// OnLogin is called when connected and logged in
	OnLogin(ctx context.Context, loginInfo *domain.LoginOKMessage) error

	// OnTicker is called when ticker updates arrive
	OnTicker(ctx context.Context, ticker *domain.TickerMessage) error

	// OnFill is called when a fill notification arrives
	OnFill(ctx context.Context, fill *domain.FillMessage) error

	// OnOffer is called when an offer request arrives
	OnOffer(ctx context.Context, offer *domain.OfferMessage) (*OfferResponse, error)

	// OnInventoryUpdate is called when inventory changes
	OnInventoryUpdate(ctx context.Context, inventory map[string]int) error

	// OnBalanceUpdate is called when balance changes
	OnBalanceUpdate(ctx context.Context, balance float64) error

	// OnOrderBookUpdate is called when orderbook updates arrive
	OnOrderBookUpdate(ctx context.Context, orderbook *domain.OrderBookUpdateMessage) error

	// Execute is called periodically to generate trading actions
	Execute(ctx context.Context, state *market.MarketState) ([]*Action, error)

	// Health returns the strategy's current health status
	Health() StrategyHealth
}

// Action represents a trading action to execute
type Action struct {
	Type        ActionType
	Order       *domain.OrderMessage
	Cancel      *domain.CancelMessage
	AcceptOffer *domain.AcceptOfferMessage
	Production  *domain.ProductionUpdateMessage
}

// ActionType defines the type of action
type ActionType string

const (
	ActionTypeOrder       ActionType = "ORDER"
	ActionTypeCancel      ActionType = "CANCEL"
	ActionTypeAcceptOffer ActionType = "ACCEPT_OFFER"
	ActionTypeProduction  ActionType = "PRODUCTION"
)

// OfferResponse represents a response to an offer
type OfferResponse struct {
	Accept          bool
	QuantityOffered int
	PriceOffered    float64
	Reason          string
}

// StrategyHealth represents the health status of a strategy
type StrategyHealth struct {
	Status     HealthStatus
	Message    string
	PnL        float64
	OpenOrders int
	ErrorCount int
	LastUpdate time.Time
}

// HealthStatus defines health statuses
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)
