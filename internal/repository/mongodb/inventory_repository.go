package mongodb

import (
	"context"
	"time"

	"github.com/yourusername/avocado-exchange-server/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type InventoryRepository struct {
	collection *mongo.Collection
}

func NewInventoryRepository(db *mongo.Database) *InventoryRepository {
	return &InventoryRepository{
		collection: db.Collection("inventory_transactions"),
	}
}

func (r *InventoryRepository) RecordTransaction(ctx context.Context, session mongo.SessionContext, transaction *domain.InventoryTransaction) error {
	transaction.Timestamp = time.Now()
	_, err := r.collection.InsertOne(session, transaction)
	return err
}

func (r *InventoryRepository) GetTeamInventory(ctx context.Context, teamName string) (map[string]int, error) {
	// This aggregates all transactions to get current inventory
	pipeline := []bson.M{
		{"$match": bson.M{"teamName": teamName}},
		{"$group": bson.M{
			"_id":   "$product",
			"total": bson.M{"$sum": "$change"},
		}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	inventory := make(map[string]int)
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Total int    `bson:"total"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		if result.Total > 0 {
			inventory[result.ID] = result.Total
		}
	}

	return inventory, cursor.Err()
}

func (r *InventoryRepository) GetTeamTransactions(ctx context.Context, teamName string, since time.Time) ([]*domain.InventoryTransaction, error) {
	filter := bson.M{
		"teamName":  teamName,
		"timestamp": bson.M{"$gte": since},
	}

	opts := options.Find().SetSort(bson.M{"timestamp": -1})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []*domain.InventoryTransaction
	if err := cursor.All(ctx, &transactions); err != nil {
		return nil, err
	}

	return transactions, nil
}

var _ domain.InventoryRepository = (*InventoryRepository)(nil)
