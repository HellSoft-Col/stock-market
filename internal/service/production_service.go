package service

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/yourusername/avocado-exchange-server/internal/domain"
)

type ProductionService struct {
	teamRepo domain.TeamRepository
}

func NewProductionService(teamRepo domain.TeamRepository) *ProductionService {
	return &ProductionService{
		teamRepo: teamRepo,
	}
}

func (s *ProductionService) ProcessProduction(ctx context.Context, teamName string, prodMsg *domain.ProductionUpdateMessage) error {
	// Validate message
	if err := s.validateProduction(prodMsg); err != nil {
		return err
	}

	// Get team to check authorized products
	team, err := s.teamRepo.GetByTeamName(ctx, teamName)
	if err != nil {
		log.Warn().
			Str("teamName", teamName).
			Err(err).
			Msg("Could not verify team for production - allowing production")
	}

	// Check if team is authorized to produce this product
	if team != nil {
		if err := s.validateAuthorization(team, prodMsg.Product); err != nil {
			return err
		}
	}

	// Log production event (server doesn't track inventory)
	log.Info().
		Str("teamName", teamName).
		Str("product", prodMsg.Product).
		Int("quantity", prodMsg.Quantity).
		Msg("Production update received")

	return nil
}

func (s *ProductionService) validateProduction(prodMsg *domain.ProductionUpdateMessage) error {
	if prodMsg.Product == "" {
		return fmt.Errorf("product is required")
	}

	if prodMsg.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	// Validate product exists
	validProducts := map[string]bool{
		"GUACA":        true,
		"SEBO":         true,
		"PALTA-OIL":    true,
		"FOSFO":        true,
		"NUCREM":       true,
		"CASCAR-ALLOY": true,
		"GTRON":        true,
		"H-GUACA":      true,
		"PITA":         true,
	}

	if !validProducts[prodMsg.Product] {
		return fmt.Errorf("invalid product: %s", prodMsg.Product)
	}

	return nil
}

func (s *ProductionService) validateAuthorization(team *domain.Team, product string) error {
	// Check if product is in team's authorized products
	for _, authorizedProduct := range team.AuthorizedProducts {
		if authorizedProduct == product {
			return nil
		}
	}

	return fmt.Errorf("team not authorized to produce %s", product)
}

// Verify the service implements the interface
var _ domain.ProductionService = (*ProductionService)(nil)
