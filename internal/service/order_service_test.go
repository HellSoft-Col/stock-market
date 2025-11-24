package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/HellSoft-Col/stock-market/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

// Mock repositories and services
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetByClOrdID(ctx context.Context, clOrdID string) (*domain.Order, error) {
	args := m.Called(ctx, clOrdID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) UpdateToFilled(
	ctx context.Context,
	session mongo.SessionContext,
	clOrdID, fillID string,
	filledQty int,
) error {
	args := m.Called(ctx, session, clOrdID, fillID, filledQty)
	return args.Error(0)
}

func (m *MockOrderRepository) UpdateToPartiallyFilled(
	ctx context.Context,
	session mongo.SessionContext,
	clOrdID, fillID string,
	filledQty int,
) error {
	args := m.Called(ctx, session, clOrdID, fillID, filledQty)
	return args.Error(0)
}

func (m *MockOrderRepository) GetPendingByProductAndSide(
	ctx context.Context,
	product, side string,
) ([]*domain.Order, error) {
	args := m.Called(ctx, product, side)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) GetPendingOrders(ctx context.Context) ([]*domain.Order, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) Cancel(ctx context.Context, clOrdID string) error {
	args := m.Called(ctx, clOrdID)
	return args.Error(0)
}

type MockMarketService struct {
	mock.Mock
}

func (m *MockMarketService) ProcessOrder(order *domain.Order, clientConn domain.ClientConnection) {
	m.Called(order, clientConn)
}

func (m *MockMarketService) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockMarketService) Stop() error {
	args := m.Called()
	return args.Error(0)
}

type MockBroadcaster struct {
	mock.Mock
}

func (m *MockBroadcaster) RegisterClient(teamName string, conn domain.ClientConnection) {
	m.Called(teamName, conn)
}

func (m *MockBroadcaster) UnregisterClient(teamName string) {
	m.Called(teamName)
}

func (m *MockBroadcaster) SendToClient(teamName string, msg interface{}) error {
	args := m.Called(teamName, msg)
	return args.Error(0)
}

func (m *MockBroadcaster) BroadcastToAll(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockBroadcaster) GetConnectedClients() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]string)
}

func TestOrderService_ProcessOrder(t *testing.T) {
	limitPrice := 10.5

	tests := []struct {
		name          string
		teamName      string
		orderMsg      *domain.OrderMessage
		setupMocks    func(*MockOrderRepository, *MockMarketService, *MockBroadcaster)
		expectedError string
	}{
		{
			name:     "successful MARKET BUY order",
			teamName: "team1",
			orderMsg: &domain.OrderMessage{
				ClOrdID: "ORD-001",
				Side:    "BUY",
				Mode:    "MARKET",
				Product: "FOSFO",
				Qty:     10,
			},
			setupMocks: func(repo *MockOrderRepository, market *MockMarketService, broadcaster *MockBroadcaster) {
				repo.On("GetByClOrdID", mock.Anything, "ORD-001").Return(nil, domain.ErrOrderNotFound)
				repo.On("Create", mock.Anything, mock.Anything).Return(nil)
				broadcaster.On("SendToClient", "team1", mock.Anything).Return(nil)
				market.On("ProcessOrder", mock.Anything, mock.Anything).Return()
			},
			expectedError: "",
		},
		{
			name:     "successful LIMIT SELL order",
			teamName: "team2",
			orderMsg: &domain.OrderMessage{
				ClOrdID:    "ORD-002",
				Side:       "SELL",
				Mode:       "LIMIT",
				Product:    "GUACA",
				Qty:        5,
				LimitPrice: &limitPrice,
			},
			setupMocks: func(repo *MockOrderRepository, market *MockMarketService, broadcaster *MockBroadcaster) {
				repo.On("GetByClOrdID", mock.Anything, "ORD-002").Return(nil, domain.ErrOrderNotFound)
				repo.On("Create", mock.Anything, mock.Anything).Return(nil)
				broadcaster.On("SendToClient", "team2", mock.Anything).Return(nil)
				market.On("ProcessOrder", mock.Anything, mock.Anything).Return()
			},
			expectedError: "",
		},
		{
			name:     "nil order message",
			teamName: "team1",
			orderMsg: nil,
			setupMocks: func(repo *MockOrderRepository, market *MockMarketService, broadcaster *MockBroadcaster) {
			},
			expectedError: "order message is nil",
		},
		{
			name:     "empty team name",
			teamName: "",
			orderMsg: &domain.OrderMessage{
				ClOrdID: "ORD-003",
				Side:    "BUY",
				Mode:    "MARKET",
				Product: "FOSFO",
				Qty:     10,
			},
			setupMocks: func(repo *MockOrderRepository, market *MockMarketService, broadcaster *MockBroadcaster) {
			},
			expectedError: "team name is required",
		},
		{
			name:     "duplicate order ID",
			teamName: "team1",
			orderMsg: &domain.OrderMessage{
				ClOrdID: "ORD-001",
				Side:    "BUY",
				Mode:    "MARKET",
				Product: "FOSFO",
				Qty:     10,
			},
			setupMocks: func(repo *MockOrderRepository, market *MockMarketService, broadcaster *MockBroadcaster) {
				existingOrder := &domain.Order{ClOrdID: "ORD-001"}
				repo.On("GetByClOrdID", mock.Anything, "ORD-001").Return(existingOrder, nil)
			},
			expectedError: "duplicate order ID: ORD-001",
		},
		{
			name:     "invalid product",
			teamName: "team1",
			orderMsg: &domain.OrderMessage{
				ClOrdID: "ORD-004",
				Side:    "BUY",
				Mode:    "MARKET",
				Product: "INVALID",
				Qty:     10,
			},
			setupMocks: func(repo *MockOrderRepository, market *MockMarketService, broadcaster *MockBroadcaster) {
			},
			expectedError: "invalid product: INVALID",
		},
		{
			name:     "invalid side",
			teamName: "team1",
			orderMsg: &domain.OrderMessage{
				ClOrdID: "ORD-005",
				Side:    "INVALID",
				Mode:    "MARKET",
				Product: "FOSFO",
				Qty:     10,
			},
			setupMocks: func(repo *MockOrderRepository, market *MockMarketService, broadcaster *MockBroadcaster) {
			},
			expectedError: "side must be BUY or SELL",
		},
		{
			name:     "zero quantity",
			teamName: "team1",
			orderMsg: &domain.OrderMessage{
				ClOrdID: "ORD-006",
				Side:    "BUY",
				Mode:    "MARKET",
				Product: "FOSFO",
				Qty:     0,
			},
			setupMocks: func(repo *MockOrderRepository, market *MockMarketService, broadcaster *MockBroadcaster) {
			},
			expectedError: "quantity must be positive",
		},
		{
			name:     "LIMIT order without price",
			teamName: "team1",
			orderMsg: &domain.OrderMessage{
				ClOrdID: "ORD-007",
				Side:    "BUY",
				Mode:    "LIMIT",
				Product: "FOSFO",
				Qty:     10,
			},
			setupMocks: func(repo *MockOrderRepository, market *MockMarketService, broadcaster *MockBroadcaster) {
			},
			expectedError: "LIMIT orders must have a positive limitPrice",
		},
		{
			name:     "database create error",
			teamName: "team1",
			orderMsg: &domain.OrderMessage{
				ClOrdID: "ORD-008",
				Side:    "BUY",
				Mode:    "MARKET",
				Product: "FOSFO",
				Qty:     10,
			},
			setupMocks: func(repo *MockOrderRepository, market *MockMarketService, broadcaster *MockBroadcaster) {
				repo.On("GetByClOrdID", mock.Anything, "ORD-008").Return(nil, domain.ErrOrderNotFound)
				repo.On("Create", mock.Anything, mock.Anything).Return(fmt.Errorf("database error"))
			},
			expectedError: "failed to save order: database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := new(MockOrderRepository)
			mockMarket := new(MockMarketService)
			mockBroadcaster := new(MockBroadcaster)

			tt.setupMocks(mockRepo, mockMarket, mockBroadcaster)

			// Create service
			svc := service.NewOrderService(mockRepo, mockMarket, mockBroadcaster)

			// Execute
			err := svc.ProcessOrder(context.Background(), tt.teamName, tt.orderMsg)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
			mockMarket.AssertExpectations(t)
			mockBroadcaster.AssertExpectations(t)
		})
	}
}

func TestOrderService_NilService(t *testing.T) {
	var svc *service.OrderService

	err := svc.ProcessOrder(context.Background(), "team1", &domain.OrderMessage{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "order service is nil")
}

func TestOrderService_NilRepository(t *testing.T) {
	mockMarket := new(MockMarketService)
	mockBroadcaster := new(MockBroadcaster)

	svc := service.NewOrderService(nil, mockMarket, mockBroadcaster)

	err := svc.ProcessOrder(context.Background(), "team1", &domain.OrderMessage{
		ClOrdID: "ORD-001",
		Side:    "BUY",
		Mode:    "MARKET",
		Product: "FOSFO",
		Qty:     10,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "order repository unavailable")
}

func TestOrderService_NilMarketService(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	mockBroadcaster := new(MockBroadcaster)

	svc := service.NewOrderService(mockRepo, nil, mockBroadcaster)

	err := svc.ProcessOrder(context.Background(), "team1", &domain.OrderMessage{
		ClOrdID: "ORD-001",
		Side:    "BUY",
		Mode:    "MARKET",
		Product: "FOSFO",
		Qty:     10,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "market service unavailable")
}
