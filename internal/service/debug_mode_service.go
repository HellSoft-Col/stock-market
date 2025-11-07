package service

import (
	"context"
	"errors"
	"sync"

	"github.com/HellSoft-Col/stock-market/internal/config"
	"github.com/HellSoft-Col/stock-market/internal/repository/mongodb"
	"github.com/rs/zerolog/log"
)

var (
	ErrDebugModeDisabled = errors.New("debug mode is disabled - this operation is not allowed in production mode")
)

// DebugModeService manages the global debug mode setting
type DebugModeService struct {
	config       *config.Config
	settingsRepo *mongodb.SystemSettingsRepository
	enabled      bool
	mu           sync.RWMutex
}

func NewDebugModeService(cfg *config.Config, settingsRepo *mongodb.SystemSettingsRepository) *DebugModeService {
	service := &DebugModeService{
		config:       cfg,
		settingsRepo: settingsRepo,
		enabled:      cfg.Market.DebugModeEnabled, // Start with config default
	}

	// Load from database if available
	ctx := context.Background()
	if dbValue, err := settingsRepo.GetDebugMode(ctx); err == nil {
		service.enabled = dbValue
		log.Info().Bool("debugMode", dbValue).Msg("Loaded debug mode setting from database")
	} else {
		log.Warn().Err(err).Msg("Failed to load debug mode from database, using config default")
	}

	return service
}

// IsEnabled returns whether debug mode is currently enabled
func (s *DebugModeService) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// SetEnabled updates the debug mode setting
func (s *DebugModeService) SetEnabled(ctx context.Context, enabled bool, updatedBy string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Save to database
	if err := s.settingsRepo.SetDebugMode(ctx, enabled, updatedBy); err != nil {
		return err
	}

	s.enabled = enabled
	log.Info().
		Bool("debugMode", enabled).
		Str("updatedBy", updatedBy).
		Msg("Debug mode setting updated")

	return nil
}

// ValidateDebugRequest checks if a debug operation is allowed
func (s *DebugModeService) ValidateDebugRequest(debugMode string) error {
	if debugMode == "" {
		return nil // Not a debug request
	}

	if !s.IsEnabled() {
		return ErrDebugModeDisabled
	}

	return nil
}
