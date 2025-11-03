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

type MarketStateRepository struct {
	collection *mongo.Collection
}

func NewMarketStateRepository(db *mongo.Database) *MarketStateRepository {
	return &MarketStateRepository{
		collection: db.Collection("market_state"),
	}
}

func (r *MarketStateRepository) UpdateLastTrade(ctx context.Context, product string, price float64, quantity int) error {
	filter := bson.M{"product": product}
	update := bson.M{
		"$set": bson.M{
			"lastTradePrice": price,
			"updatedAt":      time.Now(),
		},
		"$inc": bson.M{
			"volume24h": quantity,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to update last trade: %w", err)
	}

	return nil
}

func (r *MarketStateRepository) UpdateBestPrices(ctx context.Context, product string, bestBid, bestAsk *float64) error {
	filter := bson.M{"product": product}
	update := bson.M{
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	if bestBid != nil {
		update["$set"].(bson.M)["bestBid"] = *bestBid
	}

	if bestAsk != nil {
		update["$set"].(bson.M)["bestAsk"] = *bestAsk
	}

	// Calculate mid price if both bid and ask are available
	if bestBid != nil && bestAsk != nil {
		mid := (*bestBid + *bestAsk) / 2
		update["$set"].(bson.M)["mid"] = mid
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to update best prices: %w", err)
	}

	return nil
}

func (r *MarketStateRepository) GetByProduct(ctx context.Context, product string) (*domain.MarketState, error) {
	var marketState domain.MarketState
	err := r.collection.FindOne(ctx, bson.M{"product": product}).Decode(&marketState)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Return default market state if not found
			return &domain.MarketState{
				Product:   product,
				Volume24h: 0,
				UpdatedAt: time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get market state: %w", err)
	}
	return &marketState, nil
}

func (r *MarketStateRepository) GetAll(ctx context.Context) ([]*domain.MarketState, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to get all market states: %w", err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to close cursor")
		}
	}()

	var states []*domain.MarketState
	for cursor.Next(ctx) {
		var state domain.MarketState
		if err := cursor.Decode(&state); err != nil {
			return nil, fmt.Errorf("failed to decode market state: %w", err)
		}
		states = append(states, &state)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return states, nil
}

func (r *MarketStateRepository) Upsert(ctx context.Context, marketState *domain.MarketState) error {
	marketState.UpdatedAt = time.Now()

	filter := bson.M{"product": marketState.Product}
	opts := options.Replace().SetUpsert(true)

	_, err := r.collection.ReplaceOne(ctx, filter, marketState, opts)
	if err != nil {
		return fmt.Errorf("failed to upsert market state: %w", err)
	}

	return nil
}

var _ domain.MarketStateRepository = (*MarketStateRepository)(nil)
