package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// Repository interfaces

type TeamRepository interface {
	GetByAPIKey(ctx context.Context, apiKey string) (*Team, error)
	GetByTeamName(ctx context.Context, teamName string) (*Team, error)
	UpdateLastLogin(ctx context.Context, teamName string) error
	UpdateInventory(ctx context.Context, teamName string, inventory map[string]int) error
	UpdateBalance(ctx context.Context, teamName string, balance float64) error
	UpdateBalanceBy(ctx context.Context, teamName string, deltaBalance float64) error
	UpdateInitialBalance(ctx context.Context, teamName string, initialBalance float64) error
	UpdateMembers(ctx context.Context, teamName string, members string) error
	UpdateSpecies(ctx context.Context, teamName string, species string) error
	UpdateRecipes(ctx context.Context, teamName string, recipes map[string]Recipe) error
	Create(ctx context.Context, team *Team) error
	GetAll(ctx context.Context) ([]*Team, error)
	GetTeamsWithInventory(ctx context.Context, product string, minQuantity int) ([]*Team, error)
}

type InventoryRepository interface {
	RecordTransaction(ctx context.Context, session mongo.SessionContext, transaction *InventoryTransaction) error
	GetTeamInventory(ctx context.Context, teamName string) (map[string]int, error)
	GetTeamTransactions(ctx context.Context, teamName string, since time.Time) ([]*InventoryTransaction, error)
}

type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	GetByClOrdID(ctx context.Context, clOrdID string) (*Order, error)
	UpdateToFilled(ctx context.Context, session mongo.SessionContext, clOrdID, fillID string, filledQty int) error
	UpdateToPartiallyFilled(
		ctx context.Context,
		session mongo.SessionContext,
		clOrdID, fillID string,
		filledQty int,
	) error
	GetPendingByProductAndSide(ctx context.Context, product, side string) ([]*Order, error)
	GetPendingOrders(ctx context.Context) ([]*Order, error)
	GetHistoricalOrders(ctx context.Context, limit int) ([]*Order, error)
	Cancel(ctx context.Context, clOrdID string) error
}

type FillRepository interface {
	Create(ctx context.Context, session mongo.SessionContext, fill *Fill) error
	GetByTeamSince(ctx context.Context, teamName string, since time.Time) ([]*Fill, error)
	GetRecentSellersByProduct(ctx context.Context, product string, since time.Time) ([]string, error)
	GetAll(ctx context.Context) ([]*Fill, error)
	DeleteAll(ctx context.Context) error
}

type MarketStateRepository interface {
	UpdateLastTrade(ctx context.Context, product string, price float64, quantity int) error
	UpdateBestPrices(ctx context.Context, product string, bestBid, bestAsk *float64) error
	GetByProduct(ctx context.Context, product string) (*MarketState, error)
	GetAll(ctx context.Context) ([]*MarketState, error)
	Upsert(ctx context.Context, marketState *MarketState) error
}

// In-memory order book for Market Engine
type OrderBookRepository interface {
	AddOrder(product, side string, order *Order)
	RemoveOrder(product, side, clOrdID string) bool
	GetBestBid(product string) *Order
	GetBestAsk(product string) *Order
	GetBuyOrders(product string) []*Order
	GetSellOrders(product string) []*Order
	GetRecentSellers(product string) []string
	Clear() // For testing
	LoadFromDatabase(ctx context.Context, orderRepo OrderRepository) error
}

// Service interfaces

type AuthService interface {
	ValidateToken(ctx context.Context, token string) (*Team, error)
}

type OrderService interface {
	ProcessOrder(ctx context.Context, teamName string, orderMsg *OrderMessage) error
}

type MarketService interface {
	ProcessOrder(order *Order, clientConn ClientConnection)
	Start(ctx context.Context) error
	Stop() error
}

type ProductionService interface {
	ProcessProduction(ctx context.Context, teamName string, prodMsg *ProductionUpdateMessage) error
}

type InventoryService interface {
	UpdateInventory(
		ctx context.Context,
		teamName string,
		product string,
		change int,
		reason string,
		orderID string,
		fillID string,
	) error
	GetTeamInventory(ctx context.Context, teamName string) (map[string]int, error)
	CanSell(ctx context.Context, teamName string, product string, quantity int) (bool, error)
}

type ResyncService interface {
	GenerateEventDelta(ctx context.Context, teamName string, since time.Time) (*EventDeltaMessage, error)
}

type PerformanceService interface {
	GenerateTeamReport(ctx context.Context, teamName string, since time.Time) (*PerformanceReportMessage, error)
	GenerateGlobalReport(ctx context.Context, since time.Time) (*GlobalPerformanceReportMessage, error)
	BroadcastGlobalReport(ctx context.Context, since time.Time) error
	SendTeamReport(ctx context.Context, teamName string, since time.Time) error
}

type DebugModeService interface {
	IsEnabled() bool
	SetEnabled(ctx context.Context, enabled bool, updatedBy string) error
	ValidateDebugRequest(debugMode string) error
}

// Transport interfaces
type ClientConnection interface {
	SendMessage(msg any) error
	GetTeamName() string
	Close() error
}

type Broadcaster interface {
	RegisterClient(teamName string, conn ClientConnection)
	UnregisterClient(teamName string)
	SendToClient(teamName string, msg any) error
	BroadcastToAll(msg any) error
	GetConnectedClients() []string
}

// Database interface
type Database interface {
	GetClient() *mongo.Client
	GetDatabase() *mongo.Database
	WithTransaction(ctx context.Context, fn func(mongo.SessionContext) (any, error)) (any, error)
	Close(ctx context.Context) error
	Ping(ctx context.Context) error
}
