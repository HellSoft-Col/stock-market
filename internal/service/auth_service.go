package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/yourusername/avocado-exchange-server/internal/domain"
)

type AuthService struct {
	teamRepo domain.TeamRepository
}

func NewAuthService(teamRepo domain.TeamRepository) *AuthService {
	return &AuthService{
		teamRepo: teamRepo,
	}
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (*domain.Team, error) {
	// Validate token format
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	// Basic token format validation (should start with TK-)
	if !strings.HasPrefix(token, "TK-") {
		return nil, fmt.Errorf("invalid token format")
	}

	// Look up team by API key
	team, err := s.teamRepo.GetByAPIKey(ctx, token)
	if err == domain.ErrTeamNotFound {
		log.Warn().
			Str("token", token).
			Msg("Authentication failed - team not found")
		return nil, fmt.Errorf("invalid token")
	}

	if err != nil {
		log.Error().
			Str("token", token).
			Err(err).
			Msg("Database error during authentication")
		return nil, fmt.Errorf("authentication service error")
	}

	// Update last login timestamp
	if err := s.teamRepo.UpdateLastLogin(ctx, team.TeamName); err != nil {
		log.Warn().
			Str("teamName", team.TeamName).
			Err(err).
			Msg("Failed to update last login timestamp")
		// Don't fail authentication for this
	}

	log.Info().
		Str("teamName", team.TeamName).
		Str("species", team.Species).
		Msg("Team authenticated successfully")

	return team, nil
}

// Verify the service implements the interface
var _ domain.AuthService = (*AuthService)(nil)
