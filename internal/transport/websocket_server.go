package transport

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"github.com/yourusername/avocado-exchange-server/internal/config"
	"github.com/yourusername/avocado-exchange-server/internal/domain"
	"github.com/yourusername/avocado-exchange-server/internal/service"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebSocketServer struct {
	config      *config.Config
	server      *http.Server
	clients     map[string]*WebSocketClientHandler   // addr -> client
	teamClients map[string][]*WebSocketClientHandler // teamName -> clients
	clientsMu   sync.RWMutex
	shutdown    chan struct{}
	wg          sync.WaitGroup
	router      *MessageRouter
}

func NewWebSocketServer(cfg *config.Config, router *MessageRouter) *WebSocketServer {
	return &WebSocketServer{
		config:      cfg,
		clients:     make(map[string]*WebSocketClientHandler),
		teamClients: make(map[string][]*WebSocketClientHandler),
		shutdown:    make(chan struct{}),
		router:      router,
	}
}

func (s *WebSocketServer) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/ws", s.handleWebSocket)

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

func (s *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	s.clientsMu.RLock()
	clientCount := len(s.clients)
	s.clientsMu.RUnlock()

	if clientCount >= s.config.Server.MaxConnections {
		log.Warn().
			Int("currentConnections", clientCount).
			Int("maxConnections", s.config.Server.MaxConnections).
			Msg("Connection limit reached, rejecting client")
		http.Error(w, "Too many connections", http.StatusTooManyRequests)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade connection to WebSocket")
		return
	}

	clientAddr := r.RemoteAddr
	log.Info().
		Str("clientAddr", clientAddr).
		Msg("New WebSocket client connected")

	client := NewWebSocketClientHandler(conn, s.config, s, s.router)

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
