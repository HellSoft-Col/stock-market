package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/config"
	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/HellSoft-Col/stock-market/internal/service"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebSocketServer struct {
	config           *config.Config
	server           *http.Server
	clients          map[string]*WebSocketClientHandler   // addr -> client
	teamClients      map[string][]*WebSocketClientHandler // teamName -> clients
	clientsMu        sync.RWMutex
	shutdown         chan struct{}
	wg               sync.WaitGroup
	router           *MessageRouter
	debugModeService domain.DebugModeService
}

func NewWebSocketServer(
	cfg *config.Config,
	router *MessageRouter,
	debugModeService domain.DebugModeService,
) *WebSocketServer {
	return &WebSocketServer{
		config:           cfg,
		clients:          make(map[string]*WebSocketClientHandler),
		teamClients:      make(map[string][]*WebSocketClientHandler),
		shutdown:         make(chan struct{}),
		router:           router,
		debugModeService: debugModeService,
	}
}

func (s *WebSocketServer) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/ws", s.handleWebSocket)

	// Admin API endpoints
	mux.HandleFunc("/admin/api/debug-mode", s.handleDebugModeAPI)

	mux.HandleFunc("/", s.serveStaticFiles)

	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
	}

	log.Info().
		Str("address", addr).
		Int("maxConnections", s.config.Server.MaxConnections).
		Msg("WebSocket server started")

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("WebSocket server error")
		}
	}()

	return nil
}

// Admin API Endpoints

func (s *WebSocketServer) handleDebugModeAPI(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "GET":
		s.handleGetDebugMode(w, r)
	case "POST":
		s.handleSetDebugMode(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *WebSocketServer) handleGetDebugMode(w http.ResponseWriter, r *http.Request) {
	enabled := s.debugModeService.IsEnabled()

	response := map[string]interface{}{
		"enabled": enabled,
	}

	json.NewEncoder(w).Encode(response)
}

func (s *WebSocketServer) handleSetDebugMode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled   bool   `json:"enabled"`
		UpdatedBy string `json:"updatedBy"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate updatedBy is provided
	if req.UpdatedBy == "" {
		http.Error(w, "updatedBy field is required", http.StatusBadRequest)
		return
	}

	// Update debug mode
	ctx := r.Context()
	if err := s.debugModeService.SetEnabled(ctx, req.Enabled, req.UpdatedBy); err != nil {
		log.Error().Err(err).Msg("Failed to set debug mode")
		http.Error(w, "Failed to update debug mode", http.StatusInternalServerError)
		return
	}

	// Broadcast notification to all connected clients
	notification := map[string]interface{}{
		"type": "SYSTEM_NOTIFICATION",
		"message": fmt.Sprintf(
			"Debug mode %s by %s",
			map[bool]string{true: "ENABLED", false: "DISABLED"}[req.Enabled],
			req.UpdatedBy,
		),
		"debugMode": req.Enabled,
	}
	s.router.broadcaster.BroadcastToAll(notification)

	response := map[string]interface{}{
		"success":   true,
		"debugMode": req.Enabled,
		"message": fmt.Sprintf(
			"Debug mode %s successfully",
			map[bool]string{true: "enabled", false: "disabled"}[req.Enabled],
		),
	}

	json.NewEncoder(w).Encode(response)
}

func (s *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade connection to WebSocket")
		return
	}

	clientAddr := r.RemoteAddr
	userAgent := r.Header.Get("User-Agent")
	log.Info().
		Str("clientAddr", clientAddr).
		Str("userAgent", userAgent).
		Msg("New WebSocket client connected")

	client := NewWebSocketClientHandler(conn, s.config, s, s.router, userAgent)

	s.clientsMu.Lock()
	s.clients[clientAddr] = client
	s.clientsMu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer func() {
			if err := conn.Close(); err != nil {
				log.Error().Err(err).Msg("Failed to close WebSocket connection")
			}
		}()

		client.Handle()

		s.clientsMu.Lock()
		delete(s.clients, clientAddr)
		if client.teamName != "" {
			delete(s.clients, client.teamName)
		}
		s.clientsMu.Unlock()

		// Remove session from auth service
		if authSvc, ok := s.router.authService.(*service.AuthService); ok && clientAddr != "" {
			authSvc.RemoveSession(clientAddr)
		}

		log.Info().
			Str("clientAddr", clientAddr).
			Str("teamName", client.teamName).
			Msg("WebSocket client disconnected")
	}()
}

func (s *WebSocketServer) serveStaticFiles(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.ServeFile(w, r, "web/index.html")
		return
	}

	http.ServeFile(w, r, "web"+r.URL.Path)
}

func (s *WebSocketServer) RegisterClientByTeam(teamName string, client *WebSocketClientHandler) {
	s.clientsMu.Lock()
	s.clients[teamName] = client
	s.clientsMu.Unlock()

	log.Debug().
		Str("teamName", teamName).
		Msg("WebSocket client registered by team name")
}

func (s *WebSocketServer) SendToClient(teamName string, message interface{}) error {
	s.clientsMu.RLock()
	client, exists := s.clients[teamName]
	s.clientsMu.RUnlock()

	if !exists {
		return fmt.Errorf("client not found: %s", teamName)
	}

	return client.SendMessage(message)
}

func (s *WebSocketServer) BroadcastToAll(message interface{}) error {
	s.clientsMu.RLock()
	clients := make([]*WebSocketClientHandler, 0, len(s.clients))
	for _, client := range s.clients {
		if client.teamName != "" {
			clients = append(clients, client)
		}
	}
	s.clientsMu.RUnlock()

	var lastErr error
	for _, client := range clients {
		if err := client.SendMessage(message); err != nil {
			log.Warn().
				Str("teamName", client.teamName).
				Err(err).
				Msg("Failed to broadcast message to WebSocket client")
			lastErr = err
		}
	}

	return lastErr
}

func (s *WebSocketServer) GetConnectedClients() []string {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	teams := make([]string, 0, len(s.clients))
	for teamName, client := range s.clients {
		if client.teamName != "" && teamName == client.teamName {
			teams = append(teams, teamName)
		}
	}

	return teams
}

func (s *WebSocketServer) GetDetailedSessions() []*domain.SessionInfo {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	sessions := make([]*domain.SessionInfo, 0, len(s.clients))
	for addr, client := range s.clients {
		teamName := client.teamName
		if teamName == "" {
			teamName = "Anonymous"
		}

		// Simple heuristic: if addr contains pattern suggesting web browser
		clientType := "Java Client"
		if strings.Contains(addr, ":") {
			// This is a simplification - web browsers typically have different connection patterns
			clientType = "Web Browser"
		}

		session := &domain.SessionInfo{
			TeamName:      teamName,
			RemoteAddr:    addr,
			UserAgent:     "", // Will enhance later
			ClientType:    clientType,
			ConnectedAt:   time.Now().Add(-time.Hour).Format(time.RFC3339), // Placeholder
			LastActivity:  time.Now().Format(time.RFC3339),
			Authenticated: client.teamName != "",
		}
		sessions = append(sessions, session)
	}

	return sessions
}

func (s *WebSocketServer) Stop() error {
	log.Info().Msg("Stopping WebSocket server")

	close(s.shutdown)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if s.server != nil {
		if err := s.server.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("Error during server shutdown")
		}
	}

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("WebSocket server stopped gracefully")
	case <-time.After(15 * time.Second):
		log.Warn().Msg("WebSocket server stop timeout")
	}

	return nil
}
