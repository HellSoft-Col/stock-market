package mongodb

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/yourusername/avocado-exchange-server/internal/config"
	"github.com/yourusername/avocado-exchange-server/internal/domain"
)

type Database struct {
	client   *mongo.Client
	database *mongo.Database
	config   *config.MongoDBConfig
}

func NewDatabase(cfg *config.MongoDBConfig) *Database {
	return &Database{
		config: cfg,
	}
}

func (db *Database) Connect(ctx context.Context) error {
	log.Info().
		Str("uri", db.config.URI).
		Str("database", db.config.Database).
		Msg("Connecting to MongoDB")

	// Set client options
	clientOptions := options.Client().ApplyURI(db.config.URI)

	// Set timeouts
	clientOptions.SetServerSelectionTimeout(db.config.Timeout)
	clientOptions.SetConnectTimeout(db.config.Timeout)

	// Set connection pool settings
	clientOptions.SetMaxPoolSize(100)
	clientOptions.SetMinPoolSize(10)

	// Create client
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		client.Disconnect(ctx)
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db.client = client
	db.database = client.Database(db.config.Database)

	log.Info().Msg("Successfully connected to MongoDB")

	// Create indexes
	if err := db.createIndexes(ctx); err != nil {
		log.Warn().Err(err).Msg("Failed to create some indexes")
	}

	return nil
}

func (db *Database) createIndexes(ctx context.Context) error {
	log.Info().Msg("Creating MongoDB indexes")

	// Teams collection indexes
	teamsCollection := db.database.Collection("teams")
	teamsIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{"apiKey", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{"teamName", 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	if _, err := teamsCollection.Indexes().CreateMany(ctx, teamsIndexes); err != nil {
		log.Error().Err(err).Msg("Failed to create teams indexes")
		return err
	}

	// Orders collection indexes
	ordersCollection := db.database.Collection("orders")
	ordersIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{"clOrdID", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{"teamName", 1}, {"createdAt", -1}},
		},
		{
			Keys: bson.D{{"status", 1}, {"product", 1}, {"side", 1}},
		},
		{
			Keys: bson.D{{"status", 1}, {"expiresAt", 1}},
		},
	}
	if _, err := ordersCollection.Indexes().CreateMany(ctx, ordersIndexes); err != nil {
		log.Error().Err(err).Msg("Failed to create orders indexes")
		return err
	}

	// Fills collection indexes
	fillsCollection := db.database.Collection("fills")
	fillsIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{"fillID", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{"buyer", 1}, {"executedAt", -1}},
		},
		{
			Keys: bson.D{{"seller", 1}, {"executedAt", -1}},
		},
		{
			Keys: bson.D{{"product", 1}, {"executedAt", -1}},
		},
	}
	if _, err := fillsCollection.Indexes().CreateMany(ctx, fillsIndexes); err != nil {
		log.Error().Err(err).Msg("Failed to create fills indexes")
		return err
	}

	// Market state collection indexes
	marketStateCollection := db.database.Collection("market_state")
	marketStateIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{"product", 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	if _, err := marketStateCollection.Indexes().CreateMany(ctx, marketStateIndexes); err != nil {
		log.Error().Err(err).Msg("Failed to create market_state indexes")
		return err
	}

	log.Info().Msg("MongoDB indexes created successfully")
	return nil
}

func (db *Database) GetClient() *mongo.Client {
	return db.client
}

func (db *Database) GetDatabase() *mongo.Database {
	return db.database
}

func (db *Database) WithTransaction(ctx context.Context, fn func(mongo.SessionContext) (any, error)) (any, error) {
	session, err := db.client.StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	result, err := session.WithTransaction(ctx, fn)
	if err != nil {
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	return result, nil
}

func (db *Database) Ping(ctx context.Context) error {
	return db.client.Ping(ctx, readpref.Primary())
}

func (db *Database) Close(ctx context.Context) error {
	if db.client != nil {
		log.Info().Msg("Closing MongoDB connection")
		return db.client.Disconnect(ctx)
	}
	return nil
}

// Verify the Database implements the interface
var _ domain.Database = (*Database)(nil)
