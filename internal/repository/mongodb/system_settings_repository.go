package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SystemSettings struct {
	Key       string    `bson:"_id"       json:"key"`
	Value     bool      `bson:"value"     json:"value"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
	UpdatedBy string    `bson:"updatedBy" json:"updatedBy"`
}

type SystemSettingsRepository struct {
	collection *mongo.Collection
}

func NewSystemSettingsRepository(db *mongo.Database) *SystemSettingsRepository {
	return &SystemSettingsRepository{
		collection: db.Collection("system_settings"),
	}
}

func (r *SystemSettingsRepository) GetDebugMode(ctx context.Context) (bool, error) {
	var settings SystemSettings
	err := r.collection.FindOne(ctx, bson.M{"_id": "debugModeEnabled"}).Decode(&settings)
	if err == mongo.ErrNoDocuments {
		// Default to true if not set
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return settings.Value, nil
}

func (r *SystemSettingsRepository) SetDebugMode(ctx context.Context, enabled bool, updatedBy string) error {
	settings := SystemSettings{
		Key:       "debugModeEnabled",
		Value:     enabled,
		UpdatedAt: time.Now(),
		UpdatedBy: updatedBy,
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": "debugModeEnabled"},
		bson.M{"$set": settings},
		opts,
	)
	return err
}
