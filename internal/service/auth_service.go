package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/rs/zerolog/log"
)

type Session struct {
	TeamName     string
	Token        string
	ConnectedAt  time.Time
	LastActivity time.Time
	RemoteAddr   string
	UserAgent    string
	ClientType   string
}

type AuthService struct {
	teamRepo           domain.TeamRepository
	activeSessions     map[string][]*Session // teamName -> sessions
	sessionsByAddr     map[string]*Session   // remoteAddr -> session
	sessionsMu         sync.RWMutex
	maxSessionsPerTeam int
}

func NewAuthService(teamRepo domain.TeamRepository) *AuthService {
	return &AuthService{
		teamRepo:           teamRepo,
		activeSessions:     make(map[string][]*Session),
		sessionsByAddr:     make(map[string]*Session),
		maxSessionsPerTeam: 5, // Allow up to 5 concurrent sessions per team
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

func (s *AuthService) CreateSession(teamName, token, remoteAddr, userAgent string) *Session {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()

	// Determine client type based on user agent
	clientType := "Unknown"
	if strings.Contains(userAgent, "Mozilla") || strings.Contains(userAgent, "Chrome") ||
		strings.Contains(userAgent, "Safari") {
		clientType = "Web Browser"
	} else if userAgent == "" {
		clientType = "Java/Native Client"
	}

	session := &Session{
		TeamName:     teamName,
		Token:        token,
		ConnectedAt:  time.Now(),
		LastActivity: time.Now(),
		RemoteAddr:   remoteAddr,
		UserAgent:    userAgent,
		ClientType:   clientType,
	}

	// Check if team has too many sessions
	existingSessions := s.activeSessions[teamName]
	if len(existingSessions) >= s.maxSessionsPerTeam {
		// Remove oldest session
		oldestSession := existingSessions[0]
		s.removeSessionInternal(oldestSession)
		log.Warn().
			Str("teamName", teamName).
			Int("maxSessions", s.maxSessionsPerTeam).
			Str("removedAddr", oldestSession.RemoteAddr).
			Msg("Removed oldest session due to limit")
	}

	// Add new session
	s.activeSessions[teamName] = append(s.activeSessions[teamName], session)
	s.sessionsByAddr[remoteAddr] = session

	log.Info().
		Str("teamName", teamName).
		Str("remoteAddr", remoteAddr).
		Str("clientType", clientType).
		Int("teamSessions", len(s.activeSessions[teamName])).
		Msg("Session created")

	return session
}

func (s *AuthService) RemoveSession(remoteAddr string) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()

	session := s.sessionsByAddr[remoteAddr]
	if session != nil {
		s.removeSessionInternal(session)
	}
}

func (s *AuthService) removeSessionInternal(session *Session) {
	// Remove from sessions by addr
	delete(s.sessionsByAddr, session.RemoteAddr)

	// Remove from team sessions
	teamSessions := s.activeSessions[session.TeamName]
	for i, sess := range teamSessions {
		if sess == session {
			s.activeSessions[session.TeamName] = append(teamSessions[:i], teamSessions[i+1:]...)
			break
		}
	}

	// Remove team entry if no sessions left
	if len(s.activeSessions[session.TeamName]) == 0 {
		delete(s.activeSessions, session.TeamName)
	}

	log.Info().
		Str("teamName", session.TeamName).
		Str("remoteAddr", session.RemoteAddr).
		Msg("Session removed")
}

func (s *AuthService) UpdateSessionActivity(remoteAddr string) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()

	if session, exists := s.sessionsByAddr[remoteAddr]; exists {
		session.LastActivity = time.Now()
	}
}

func (s *AuthService) GetActiveSessions() map[string][]*Session {
	s.sessionsMu.RLock()
	defer s.sessionsMu.RUnlock()

	// Create a deep copy to avoid race conditions
	result := make(map[string][]*Session)
	for teamName, sessions := range s.activeSessions {
		sessionsCopy := make([]*Session, len(sessions))
		copy(sessionsCopy, sessions)
		result[teamName] = sessionsCopy
	}

	return result
}

func (s *AuthService) GetTeamSessionCount(teamName string) int {
	s.sessionsMu.RLock()
	defer s.sessionsMu.RUnlock()

	return len(s.activeSessions[teamName])
}

var _ domain.AuthService = (*AuthService)(nil)

func (s *AuthService) GetAllTeams(ctx context.Context) ([]*domain.Team, error) {
	if s.teamRepo == nil {
		return nil, fmt.Errorf("team repository is nil")
	}

	return s.teamRepo.GetAll(ctx)
}

func (s *AuthService) UpdateTeam(
	ctx context.Context,
	teamName string,
	balance float64,
	inventory map[string]int,
) error {
	if s.teamRepo == nil {
		return fmt.Errorf("team repository is nil")
	}

	if err := s.teamRepo.UpdateBalance(ctx, teamName, balance); err != nil {
		return err
	}

	if err := s.teamRepo.UpdateInventory(ctx, teamName, inventory); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) ResetTeamBalance(ctx context.Context, teamName string) error {
	if s.teamRepo == nil {
		return fmt.Errorf("team repository is nil")
	}

	team, err := s.teamRepo.GetByTeamName(ctx, teamName)
	if err != nil {
		return err
	}

	return s.teamRepo.UpdateBalance(ctx, teamName, team.InitialBalance)
}

func (s *AuthService) ResetTeamInventory(ctx context.Context, teamName string) error {
	if s.teamRepo == nil {
		return fmt.Errorf("team repository is nil")
	}

	// Reset to initial inventory (empty or default values)
	inventory := make(map[string]int)
	products := []string{"FOSFO", "GUACA", "SEBO", "PALTA-OIL", "PITA", "H-GUACA"}
	for _, product := range products {
		inventory[product] = 0
	}

	return s.teamRepo.UpdateInventory(ctx, teamName, inventory)
}

func (s *AuthService) UpdateInitialBalance(ctx context.Context, teamName string, balance float64) error {
	if s.teamRepo == nil {
		return fmt.Errorf("team repository is nil")
	}

	return s.teamRepo.UpdateInitialBalance(ctx, teamName, balance)
}

func (s *AuthService) UpdateTeamMembers(ctx context.Context, teamName string, members string) error {
	if s.teamRepo == nil {
		return fmt.Errorf("team repository is nil")
	}

	return s.teamRepo.UpdateMembers(ctx, teamName, members)
}

func (s *AuthService) UpdateRecipes(ctx context.Context, teamName string, recipes map[string]domain.Recipe) error {
	if s.teamRepo == nil {
		return fmt.Errorf("team repository is nil")
	}

	return s.teamRepo.UpdateRecipes(ctx, teamName, recipes)
}
