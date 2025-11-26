package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/HellSoft-Col/stock-market/internal/domain"
)

type FillRepository struct {
	collection *mongo.Collection
}

func NewFillRepository(db *mongo.Database) *FillRepository {
	return &FillRepository{
		collection: db.Collection("fills"),
	}
}

func (r *FillRepository) Create(ctx context.Context, session mongo.SessionContext, fill *domain.Fill) error {
	fill.ExecutedAt = time.Now()

	_, err := r.collection.InsertOne(session, fill)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("fill ID already exists: %w", err)
		}
		return fmt.Errorf("failed to create fill: %w", err)
	}

	return nil
}

func (r *FillRepository) GetByTeamSince(ctx context.Context, teamName string, since time.Time) ([]*domain.Fill, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"buyer": teamName},
			{"seller": teamName},
		},
		"executedAt": bson.M{"$gt": since},
	}

	opts := options.Find().SetSort(bson.M{"executedAt": 1})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get fills by team since: %w", err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to close cursor")
		}
	}()

	var fills []*domain.Fill
	for cursor.Next(ctx) {
		var fill domain.Fill
		if err := cursor.Decode(&fill); err != nil {
			return nil, fmt.Errorf("failed to decode fill: %w", err)
		}
		fills = append(fills, &fill)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return fills, nil
}

func (r *FillRepository) GetRecentSellersByProduct(
	ctx context.Context,
	product string,
	since time.Time,
) ([]string, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"product":    product,
				"executedAt": bson.M{"$gt": since},
			},
		},
		{
			"$group": bson.M{
				"_id": "$seller",
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent sellers: %w", err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to close cursor")
		}
	}()

	var sellers []string
	for cursor.Next(ctx) {
		var result struct {
			ID string `bson:"_id"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode seller: %w", err)
		}
		sellers = append(sellers, result.ID)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return sellers, nil
}

func (r *FillRepository) GetAll(ctx context.Context) ([]*domain.Fill, error) {
	opts := options.Find().SetSort(bson.M{"executedAt": -1})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get all fills: %w", err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to close cursor")
		}
	}()

	var fills []*domain.Fill
	for cursor.Next(ctx) {
		var fill domain.Fill
		if err := cursor.Decode(&fill); err != nil {
			return nil, fmt.Errorf("failed to decode fill: %w", err)
		}
		fills = append(fills, &fill)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return fills, nil
}

func (r *FillRepository) DeleteAll(ctx context.Context) error {
	result, err := r.collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to delete all fills: %w", err)
	}

	log.Info().
		Int64("deletedCount", result.DeletedCount).
		Msg("All fills deleted from database")

	return nil
}

var _ domain.FillRepository = (*FillRepository)(nil)
