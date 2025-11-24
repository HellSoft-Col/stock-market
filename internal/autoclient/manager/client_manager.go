package manager

import (
	"context"
	"fmt"
	"sync"

	"github.com/HellSoft-Col/stock-market/internal/autoclient/config"
	"github.com/HellSoft-Col/stock-market/internal/autoclient/strategy"
	"github.com/rs/zerolog/log"
)

// ClientManager manages multiple trading sessions
type ClientManager struct {
	config          *config.Config
	strategyFactory *strategy.StrategyFactory
	sessions        map[string]*TradingSession
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewClientManager creates a new client manager
func NewClientManager(cfg *config.Config) *ClientManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &ClientManager{
		config:          cfg,
		strategyFactory: strategy.NewStrategyFactory(),
		sessions:        make(map[string]*TradingSession),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// Start starts all enabled clients
func (cm *ClientManager) Start() error {
	log.Info().Msg("Starting client manager")

	// Get enabled clients
	enabledClients := cm.config.GetEnabledClients()

	if len(enabledClients) == 0 {
		return fmt.Errorf("no enabled clients found in configuration")
	}

	log.Info().
		Int("count", len(enabledClients)).
		Msg("Starting trading clients")

	// Start each client
	for _, clientCfg := range enabledClients {
		if err := cm.startClient(clientCfg); err != nil {
			log.Error().
				Err(err).
				Str("client", clientCfg.Name).
				Msg("Failed to start client")
			continue
		}
	}

	log.Info().
		Int("active", len(cm.sessions)).
		Msg("Client manager started")

	return nil
}

// startClient starts a single client session
func (cm *ClientManager) startClient(clientCfg config.ClientConfig) error {
	log.Info().
		Str("name", clientCfg.Name).
		Str("strategy", clientCfg.Strategy).
		Str("species", clientCfg.Species).
		Msg("Starting client")

	// Create strategy
	strat, err := cm.strategyFactory.Create(clientCfg.Strategy, clientCfg.Name)
	if err != nil {
		return fmt.Errorf("failed to create strategy: %w", err)
	}

	// Create session
	session := NewTradingSession(
		clientCfg.Name,
		clientCfg,
		cm.config.Server.Host,
		cm.config.Server.Port,
		cm.config.Server.ReconnectInterval,
		cm.config.Server.MaxReconnectAttempts,
		strat,
	)

	// Store session
	cm.mu.Lock()
	cm.sessions[clientCfg.Name] = session
	cm.mu.Unlock()

	// Start session
	if err := session.Start(); err != nil {
		cm.mu.Lock()
		delete(cm.sessions, clientCfg.Name)
		cm.mu.Unlock()
		return fmt.Errorf("failed to start session: %w", err)
	}

	log.Info().
		Str("name", clientCfg.Name).
		Msg("Client started successfully")

	return nil
}

// Stop stops all clients
func (cm *ClientManager) Stop() error {
	log.Info().Msg("Stopping client manager")

	cm.cancel()

	// Stop all sessions
	cm.mu.Lock()
	sessions := make([]*TradingSession, 0, len(cm.sessions))
	for _, session := range cm.sessions {
		sessions = append(sessions, session)
	}
	cm.mu.Unlock()

	// Stop sessions in parallel
	var wg sync.WaitGroup
	for _, session := range sessions {
		wg.Add(1)
		go func(s *TradingSession) {
			defer wg.Done()
			if err := s.Stop(); err != nil {
				log.Error().
					Err(err).
					Str("session", s.id).
					Msg("Error stopping session")
			}
		}(session)
	}

	wg.Wait()

	log.Info().Msg("Client manager stopped")

	return nil
}

// GetSession returns a session by name
func (cm *ClientManager) GetSession(name string) (*TradingSession, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	session, exists := cm.sessions[name]
	return session, exists
}

// GetAllSessions returns all sessions
func (cm *ClientManager) GetAllSessions() []*TradingSession {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	sessions := make([]*TradingSession, 0, len(cm.sessions))
	for _, session := range cm.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// GetStats returns statistics for all sessions
func (cm *ClientManager) GetStats() []map[string]interface{} {
	sessions := cm.GetAllSessions()

	stats := make([]map[string]interface{}, len(sessions))
	for i, session := range sessions {
		stats[i] = session.GetStats()
	}

	return stats
}

// GetConnectedCount returns the number of connected sessions
func (cm *ClientManager) GetConnectedCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	count := 0
	for _, session := range cm.sessions {
		if session.IsConnected() {
			count++
		}
	}

	return count
}
