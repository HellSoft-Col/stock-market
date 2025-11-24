package service

import (
	"context"
	"fmt"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
)

type InventoryService struct {
	teamRepo      domain.TeamRepository
	inventoryRepo domain.InventoryRepository
	db            domain.Database
}

func NewInventoryService(
	teamRepo domain.TeamRepository,
	inventoryRepo domain.InventoryRepository,
	db domain.Database,
) *InventoryService {
	return &InventoryService{
		teamRepo:      teamRepo,
		inventoryRepo: inventoryRepo,
		db:            db,
	}
}

func (s *InventoryService) UpdateInventory(
	ctx context.Context,
	teamName string,
	product string,
	change int,
	reason string,
	orderID string,
	fillID string,
) error {
	_, err := s.db.WithTransaction(ctx, func(sc mongo.SessionContext) (any, error) {
		// Get current team data
		team, err := s.teamRepo.GetByTeamName(ctx, teamName)
		if err != nil {
			return nil, fmt.Errorf("failed to get team: %w", err)
		}

		// Initialize inventory if nil
		if team.Inventory == nil {
			team.Inventory = make(map[string]int)
		}

		// Calculate new quantity
		currentQty := team.Inventory[product]
		newQty := currentQty + change

		// Validate the change
		if newQty < 0 {
			return nil, fmt.Errorf("insufficient inventory: have %d, trying to change by %d", currentQty, change)
		}

		// Update inventory
		team.Inventory[product] = newQty
		team.LastInventoryUpdate = time.Now()

		// Save updated inventory
		if err := s.teamRepo.UpdateInventory(ctx, teamName, team.Inventory); err != nil {
			return nil, fmt.Errorf("failed to update inventory: %w", err)
		}

		// Record the transaction
		transaction := &domain.InventoryTransaction{
			TeamName:  teamName,
			Product:   product,
			Change:    change,
			Reason:    reason,
			OrderID:   orderID,
			FillID:    fillID,
			Timestamp: time.Now(),
		}

		if err := s.inventoryRepo.RecordTransaction(ctx, sc, transaction); err != nil {
			return nil, fmt.Errorf("failed to record inventory transaction: %w", err)
		}

		log.Info().
			Str("teamName", teamName).
			Str("product", product).
			Int("change", change).
			Int("newQty", newQty).
			Str("reason", reason).
			Msg("Inventory updated")

		return nil, nil
	})

	return err
}

func (s *InventoryService) GetTeamInventory(ctx context.Context, teamName string) (map[string]int, error) {
	team, err := s.teamRepo.GetByTeamName(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	if team.Inventory == nil {
		return make(map[string]int), nil
	}

	return team.Inventory, nil
}

func (s *InventoryService) CanSell(ctx context.Context, teamName string, product string, quantity int) (bool, error) {
	inventory, err := s.GetTeamInventory(ctx, teamName)
	if err != nil {
		return false, err
	}

	available := inventory[product]
	return available >= quantity, nil
}

// Initialize inventory for a new team
func (s *InventoryService) InitializeTeamInventory(
	ctx context.Context,
	teamName string,
	initialInventory map[string]int,
) error {
	for product, quantity := range initialInventory {
		if err := s.UpdateInventory(ctx, teamName, product, quantity, "INITIAL", "", ""); err != nil {
			return fmt.Errorf("failed to initialize inventory for %s: %w", product, err)
		}
	}
	return nil
}

// Get teams that have enough inventory to sell a product
func (s *InventoryService) GetEligibleSellers(
	ctx context.Context,
	product string,
	minQuantity int,
) ([]*domain.Team, error) {
	return s.teamRepo.GetTeamsWithInventory(ctx, product, minQuantity)
}

var _ domain.InventoryService = (*InventoryService)(nil)
