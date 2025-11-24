package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/autoclient/agent"
	"github.com/HellSoft-Col/stock-market/internal/autoclient/config"
	"github.com/HellSoft-Col/stock-market/internal/autoclient/strategy"
	"github.com/HellSoft-Col/stock-market/internal/client"
	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/rs/zerolog/log"
)

// TradingSession represents a single automated trading session
type TradingSession struct {
	id       string
	config   config.ClientConfig
	client   *client.WebSocketClient
	agent    *agent.TradingAgent
	strategy strategy.Strategy

	// Connection state
	connected     bool
	authenticated bool
	lastSync      time.Time

	// Context and control
	ctx    context.Context
	cancel context.CancelFunc
	stopCh chan struct{}
	wg     sync.WaitGroup
	mu     sync.RWMutex

	// Server config
	serverHost string
	serverPort int

	// Reconnection
	reconnectInterval    time.Duration
	maxReconnectAttempts int
	reconnectAttempts    int
}

// NewTradingSession creates a new trading session
func NewTradingSession(
	id string,
	clientConfig config.ClientConfig,
	serverHost string,
	serverPort int,
	reconnectInterval time.Duration,
	maxReconnectAttempts int,
	strat strategy.Strategy,
) *TradingSession {
	ctx, cancel := context.WithCancel(context.Background())

	return &TradingSession{
		id:                   id,
		config:               clientConfig,
		serverHost:           serverHost,
		serverPort:           serverPort,
		reconnectInterval:    reconnectInterval,
		maxReconnectAttempts: maxReconnectAttempts,
		strategy:             strat,
		ctx:                  ctx,
		cancel:               cancel,
		stopCh:               make(chan struct{}),
	}
}

// Start starts the trading session
func (s *TradingSession) Start() error {
	log.Info().
		Str("session", s.id).
		Str("strategy", s.config.Strategy).
		Msg("Starting trading session")

	// Initialize strategy
	if err := s.strategy.Initialize(s.config.Config); err != nil {
		return fmt.Errorf("failed to initialize strategy: %w", err)
	}

	// Start connection loop
	s.wg.Add(1)
	go s.connectionLoop()

	return nil
}

// Stop stops the trading session
func (s *TradingSession) Stop() error {
	log.Info().
		Str("session", s.id).
		Msg("Stopping trading session")

	s.cancel()
	close(s.stopCh)

	if s.agent != nil {
		s.agent.Stop()
	}

	if s.client != nil {
		s.client.Close()
	}

	s.wg.Wait()

	return nil
}

// connectionLoop manages connection and reconnection
func (s *TradingSession) connectionLoop() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.stopCh:
			return
		default:
			if err := s.connectAndRun(); err != nil {
				log.Error().
					Err(err).
					Str("session", s.id).
					Msg("Connection error")

				s.mu.Lock()
				s.connected = false
				s.authenticated = false
				s.reconnectAttempts++
				s.mu.Unlock()

				if s.reconnectAttempts >= s.maxReconnectAttempts {
					log.Error().
						Str("session", s.id).
						Int("attempts", s.reconnectAttempts).
						Msg("Max reconnect attempts reached, giving up")
					return
				}

				log.Info().
					Str("session", s.id).
					Int("attempt", s.reconnectAttempts).
					Dur("wait", s.reconnectInterval).
					Msg("Reconnecting...")

				select {
				case <-time.After(s.reconnectInterval):
					continue
				case <-s.ctx.Done():
					return
				case <-s.stopCh:
					return
				}
			}
		}
	}
}

// connectAndRun connects to server and runs the message loop
func (s *TradingSession) connectAndRun() error {
	// Create WebSocket client
	s.client = client.NewWebSocketClient(fmt.Sprintf("%s:%d", s.serverHost, s.serverPort))

	// Connect
	if err := s.client.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	log.Info().
		Str("session", s.id).
		Str("server", fmt.Sprintf("%s:%d", s.serverHost, s.serverPort)).
		Msg("Connected to server")

	s.mu.Lock()
	s.connected = true
	s.mu.Unlock()

	// Create agent
	s.agent = agent.NewTradingAgent(s.id, s.client, s.strategy)

	// Login
	if err := s.login(); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Start agent
	if err := s.agent.Start(s.ctx); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	// Reset reconnect attempts on successful connection
	s.mu.Lock()
	s.reconnectAttempts = 0
	s.mu.Unlock()

	// Run message loop
	return s.messageLoop()
}

// login performs the login handshake
func (s *TradingSession) login() error {
	log.Info().
		Str("session", s.id).
		Msg("Logging in...")

	loginMsg := &domain.LoginMessage{
		Type:  "LOGIN",
		Token: s.config.Token,
		TZ:    "UTC",
	}

	if err := s.client.SendMessage(loginMsg); err != nil {
		return fmt.Errorf("failed to send login: %w", err)
	}

	// Wait for LOGIN_OK
	if err := s.client.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return fmt.Errorf("failed to set read deadline: %w", err)
	}

	var response map[string]interface{}
	if err := s.client.ReadMessage(&response); err != nil {
		return fmt.Errorf("failed to read login response: %w", err)
	}

	// Check response type
	msgType, ok := response["type"].(string)
	if !ok {
		return fmt.Errorf("invalid login response: missing type")
	}

	if msgType == "ERROR" {
		reason := response["reason"].(string)
		return fmt.Errorf("login error: %s", reason)
	}

	if msgType != "LOGIN_OK" {
		return fmt.Errorf("unexpected response type: %s", msgType)
	}

	// Parse LOGIN_OK message
	loginOK, err := s.parseLoginOK(response)
	if err != nil {
		return fmt.Errorf("failed to parse LOGIN_OK: %w", err)
	}

	// Update session state
	s.mu.Lock()
	s.authenticated = true
	s.lastSync = time.Now()
	s.mu.Unlock()

	// Notify agent
	if err := s.agent.HandleLoginOK(loginOK); err != nil {
		return fmt.Errorf("agent login failed: %w", err)
	}

	log.Info().
		Str("session", s.id).
		Str("team", loginOK.Team).
		Str("species", loginOK.Species).
		Float64("balance", loginOK.CurrentBalance).
		Msg("Login successful")

	// Trigger resync if this is a reconnection (not first connection)
	s.mu.RLock()
	isReconnection := s.reconnectAttempts > 0 || !s.lastSync.IsZero()
	s.mu.RUnlock()

	if isReconnection {
		log.Info().
			Str("session", s.id).
			Msg("Reconnection detected, triggering resync")
		if err := s.Resync(); err != nil {
			log.Warn().
				Err(err).
				Msg("Resync failed but continuing")
		}
	}

	return nil
}

// parseLoginOK parses LOGIN_OK message
func (s *TradingSession) parseLoginOK(data map[string]interface{}) (*domain.LoginOKMessage, error) {
	// Re-marshal and unmarshal to convert to proper types
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var loginOK domain.LoginOKMessage
	if err := json.Unmarshal(jsonData, &loginOK); err != nil {
		return nil, err
	}

	return &loginOK, nil
}

// messageLoop reads and processes messages from the server
func (s *TradingSession) messageLoop() error {
	log.Info().
		Str("session", s.id).
		Msg("Starting message loop")

	// Clear read deadline for message loop
	if err := s.client.SetReadDeadline(time.Time{}); err != nil {
		return fmt.Errorf("failed to clear read deadline: %w", err)
	}

	for {
		select {
		case <-s.ctx.Done():
			return nil
		case <-s.stopCh:
			return nil
		default:
			var message map[string]interface{}
			if err := s.client.ReadMessage(&message); err != nil {
				return fmt.Errorf("failed to read message: %w", err)
			}

			if err := s.handleMessage(message); err != nil {
				log.Error().
					Err(err).
					Str("session", s.id).
					Interface("message", message).
					Msg("Failed to handle message")
			}
		}
	}
}

// handleMessage routes messages to appropriate handlers
func (s *TradingSession) handleMessage(data map[string]interface{}) error {
	msgType, ok := data["type"].(string)
	if !ok {
		return fmt.Errorf("message missing type field")
	}

	// Re-marshal for proper type conversion
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	switch msgType {
	case "FILL":
		var fill domain.FillMessage
		if err := json.Unmarshal(jsonData, &fill); err != nil {
			return err
		}
		return s.agent.HandleFill(&fill)

	case "TICKER":
		var ticker domain.TickerMessage
		if err := json.Unmarshal(jsonData, &ticker); err != nil {
			return err
		}
		return s.agent.HandleTicker(&ticker)

	case "OFFER":
		var offer domain.OfferMessage
		if err := json.Unmarshal(jsonData, &offer); err != nil {
			return err
		}
		return s.agent.HandleOffer(&offer)

	case "INVENTORY_UPDATE":
		var invUpdate domain.InventoryUpdateMessage
		if err := json.Unmarshal(jsonData, &invUpdate); err != nil {
			return err
		}
		return s.agent.HandleInventoryUpdate(invUpdate.Inventory)

	case "BALANCE_UPDATE":
		var balUpdate domain.BalanceUpdateMessage
		if err := json.Unmarshal(jsonData, &balUpdate); err != nil {
			return err
		}
		return s.agent.HandleBalanceUpdate(balUpdate.Balance)

	case "ORDER_BOOK_UPDATE":
		var obUpdate domain.OrderBookUpdateMessage
		if err := json.Unmarshal(jsonData, &obUpdate); err != nil {
			return err
		}
		return s.agent.HandleOrderBookUpdate(&obUpdate)

	case "ORDER_ACK":
		log.Debug().
			Str("session", s.id).
			Interface("ack", data).
			Msg("Order acknowledged")
		return nil

	case "ERROR":
		var errMsg domain.ErrorMessage
		if err := json.Unmarshal(jsonData, &errMsg); err != nil {
			return err
		}
		s.agent.HandleError(&errMsg)
		return nil

	case "EVENT_DELTA":
		// Handle event delta (array of fills)
		var eventDelta domain.EventDeltaMessage
		if err := json.Unmarshal(jsonData, &eventDelta); err != nil {
			return err
		}
		for _, fill := range eventDelta.Events {
			if err := s.agent.HandleFill(&fill); err != nil {
				log.Error().Err(err).Msg("Failed to handle event delta fill")
			}
		}
		return nil

	default:
		log.Debug().
			Str("session", s.id).
			Str("type", msgType).
			Msg("Unhandled message type")
		return nil
	}
}

// Resync performs resync operation to recover missed events
func (s *TradingSession) Resync() error {
	s.mu.RLock()
	lastSync := s.lastSync
	s.mu.RUnlock()

	log.Info().
		Str("session", s.id).
		Time("lastSync", lastSync).
		Msg("Performing resync")

	resyncMsg := &domain.ResyncMessage{
		Type:     "RESYNC",
		LastSync: lastSync.Format(time.RFC3339),
	}

	if err := s.client.SendMessage(resyncMsg); err != nil {
		return fmt.Errorf("failed to send resync: %w", err)
	}

	// Server will respond with EVENT_DELTA containing missed events
	log.Info().
		Str("session", s.id).
		Msg("Resync request sent")

	return nil
}

// GetStats returns session statistics
func (s *TradingSession) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]interface{}{
		"id":                s.id,
		"connected":         s.connected,
		"authenticated":     s.authenticated,
		"lastSync":          s.lastSync,
		"reconnectAttempts": s.reconnectAttempts,
	}

	if s.agent != nil {
		stats["agent"] = s.agent.GetStats()
	}

	return stats
}

// IsConnected returns connection status
func (s *TradingSession) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connected && s.authenticated
}
