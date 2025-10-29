package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/yourusername/avocado-exchange-server/internal/domain"
)

type TeamRepository struct {
	collection *mongo.Collection
}

func NewTeamRepository(db *mongo.Database) *TeamRepository {
	return &TeamRepository{
		collection: db.Collection("teams"),
	}
}

func (r *TeamRepository) GetByAPIKey(ctx context.Context, apiKey string) (*domain.Team, error) {
	var team domain.Team
	err := r.collection.FindOne(ctx, bson.M{"apiKey": apiKey}).Decode(&team)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to get team by API key: %w", err)
	}
	return &team, nil
}

func (r *TeamRepository) GetByTeamName(ctx context.Context, teamName string) (*domain.Team, error) {
	var team domain.Team
	err := r.collection.FindOne(ctx, bson.M{"teamName": teamName}).Decode(&team)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to get team by name: %w", err)
	}
	return &team, nil
}

func (r *TeamRepository) UpdateLastLogin(ctx context.Context, teamName string) error {
	update := bson.M{
		"$set": bson.M{
			"lastLogin": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"teamName": teamName}, update)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrTeamNotFound
	}

	return nil
}

func (r *TeamRepository) Create(ctx context.Context, team *domain.Team) error {
	team.CreatedAt = time.Now()
	team.LastLogin = time.Now()

	_, err := r.collection.InsertOne(ctx, team)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("team already exists: %w", err)
		}
		return fmt.Errorf("failed to create team: %w", err)
	}

	return nil
}

func (r *TeamRepository) GetAll(ctx context.Context) ([]*domain.Team, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %w", err)
	}
	defer cursor.Close(ctx)

	var teams []*domain.Team
	for cursor.Next(ctx) {
		var team domain.Team
		if err := cursor.Decode(&team); err != nil {
			return nil, fmt.Errorf("failed to decode team: %w", err)
		}
		teams = append(teams, &team)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return teams, nil
}

func (r *TeamRepository) UpdateInventory(ctx context.Context, teamName string, inventory map[string]int) error {
	update := bson.M{
		"$set": bson.M{
			"inventory":           inventory,
			"lastInventoryUpdate": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"teamName": teamName}, update)
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrTeamNotFound
	}

	return nil
}

func (r *TeamRepository) UpdateBalance(ctx context.Context, teamName string, balance float64) error {
	update := bson.M{
		"$set": bson.M{
			"currentBalance": balance,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"teamName": teamName}, update)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrTeamNotFound
	}

	return nil
}

func (r *TeamRepository) GetTeamsWithInventory(ctx context.Context, product string, minQuantity int) ([]*domain.Team, error) {
	filter := bson.M{
		fmt.Sprintf("inventory.%s", product): bson.M{"$gte": minQuantity},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams with inventory: %w", err)
	}
	defer cursor.Close(ctx)

	var teams []*domain.Team
	for cursor.Next(ctx) {
		var team domain.Team
		if err := cursor.Decode(&team); err != nil {
			return nil, fmt.Errorf("failed to decode team: %w", err)
		}
		teams = append(teams, &team)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return teams, nil
}

var _ domain.TeamRepository = (*TeamRepository)(nil)
