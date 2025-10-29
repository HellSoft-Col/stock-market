package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/yourusername/avocado-exchange-server/internal/domain"
)

type OrderRepository struct {
	collection *mongo.Collection
}

func NewOrderRepository(db *mongo.Database) *OrderRepository {
	return &OrderRepository{
		collection: db.Collection("orders"),
	}
}

func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
	order.CreatedAt = time.Now()
	order.Status = "PENDING"
	order.FilledQty = 0

	_, err := r.collection.InsertOne(ctx, order)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("order ID already exists: %w", err)
		}
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

func (r *OrderRepository) GetByClOrdID(ctx context.Context, clOrdID string) (*domain.Order, error) {
	var order domain.Order
	err := r.collection.FindOne(ctx, bson.M{"clOrdID": clOrdID}).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order by clOrdID: %w", err)
	}
	return &order, nil
}

func (r *OrderRepository) UpdateToFilled(ctx context.Context, session mongo.SessionContext, clOrdID, fillID string, filledQty int) error {
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":    "FILLED",
			"filledBy":  fillID,
			"filledQty": filledQty,
			"filledAt":  now,
		},
	}

	result, err := r.collection.UpdateOne(session, bson.M{"clOrdID": clOrdID}, update)
	if err != nil {
		return fmt.Errorf("failed to update order to filled: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrOrderNotFound
	}

	return nil
}

func (r *OrderRepository) UpdateToPartiallyFilled(ctx context.Context, session mongo.SessionContext, clOrdID, fillID string, filledQty int) error {
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":   "PARTIALLY_FILLED",
			"filledBy": fillID,
			"filledAt": now,
		},
		"$inc": bson.M{
			"filledQty": filledQty,
		},
	}

	result, err := r.collection.UpdateOne(session, bson.M{"clOrdID": clOrdID}, update)
	if err != nil {
		return fmt.Errorf("failed to update order to partially filled: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrOrderNotFound
	}

	return nil
}

func (r *OrderRepository) GetPendingByProductAndSide(ctx context.Context, product, side string) ([]*domain.Order, error) {
	filter := bson.M{
		"product": product,
		"side":    side,
		"status":  bson.M{"$in": []string{"PENDING", "PARTIALLY_FILLED"}},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending orders: %w", err)
	}
	defer cursor.Close(ctx)

	var orders []*domain.Order
	for cursor.Next(ctx) {
		var order domain.Order
		if err := cursor.Decode(&order); err != nil {
			return nil, fmt.Errorf("failed to decode order: %w", err)
		}
		orders = append(orders, &order)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return orders, nil
}

func (r *OrderRepository) GetPendingOrders(ctx context.Context) ([]*domain.Order, error) {
	filter := bson.M{
		"status": bson.M{"$in": []string{"PENDING", "PARTIALLY_FILLED"}},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending orders: %w", err)
	}
	defer cursor.Close(ctx)

	var orders []*domain.Order
	for cursor.Next(ctx) {
		var order domain.Order
		if err := cursor.Decode(&order); err != nil {
			return nil, fmt.Errorf("failed to decode order: %w", err)
		}
		orders = append(orders, &order)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return orders, nil
}

func (r *OrderRepository) Cancel(ctx context.Context, clOrdID string) error {
	update := bson.M{
		"$set": bson.M{
			"status": "CANCELLED",
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"clOrdID": clOrdID}, update)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrOrderNotFound
	}

	return nil
}

var _ domain.OrderRepository = (*OrderRepository)(nil)
