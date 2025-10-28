package transport

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/yourusername/avocado-exchange-server/internal/config"
)

type TCPServer struct {
	config    *config.Config
	listener  net.Listener
	clients   map[string]*ClientHandler
	clientsMu sync.RWMutex
	shutdown  chan struct{}
	wg        sync.WaitGroup
	router    *MessageRouter
}

func NewTCPServer(cfg *config.Config, router *MessageRouter) *TCPServer {
	return &TCPServer{
		config:   cfg,
		clients:  make(map[string]*ClientHandler),
		shutdown: make(chan struct{}),
		router:   router,
	}
}

func (s *TCPServer) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.listener = listener

	log.Info().
		Str("address", addr).
		Int("maxConnections", s.config.Server.MaxConnections).
		Msg("TCP server started")

	// Start accepting connections
	s.wg.Add(1)
	go s.acceptConnections()

	return nil
}

func (s *TCPServer) acceptConnections() {
	defer s.wg.Done()

	for {
		select {
		case <-s.shutdown:
			return
		default:
		}

		// Set accept deadline to check for shutdown
		if deadline, ok := s.listener.(*net.TCPListener); ok {
			deadline.SetDeadline(time.Now().Add(1 * time.Second))
		}

		conn, err := s.listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // Check shutdown signal again
			}
			log.Error().Err(err).Msg("Failed to accept connection")
			continue
		}

		// Check connection limit
		s.clientsMu.RLock()
		clientCount := len(s.clients)
		s.clientsMu.RUnlock()

		if clientCount >= s.config.Server.MaxConnections {
			log.Warn().
				Int("currentConnections", clientCount).
				Int("maxConnections", s.config.Server.MaxConnections).
				Msg("Connection limit reached, rejecting client")
			conn.Close()
			continue
		}

		// Handle new client
		s.wg.Add(1)
		go s.handleClient(conn)
	}
}

func (s *TCPServer) handleClient(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	log.Info().
		Str("clientAddr", clientAddr).
		Msg("New client connected")

	// Create client handler
	client := NewClientHandler(conn, s.config, s, s.router)

	// Register client temporarily with address
	s.clientsMu.Lock()
	s.clients[clientAddr] = client
	s.clientsMu.Unlock()

	// Handle client messages
	client.Handle()

	// Cleanup on disconnect
	s.clientsMu.Lock()
	delete(s.clients, clientAddr)
	if client.teamName != "" {
		delete(s.clients, client.teamName)
	}
	s.clientsMu.Unlock()

	log.Info().
		Str("clientAddr", clientAddr).
		Str("teamName", client.teamName).
		Msg("Client disconnected")
}

func (s *TCPServer) RegisterClientByTeam(teamName string, client *ClientHandler) {
	s.clientsMu.Lock()
	s.clients[teamName] = client
	s.clientsMu.Unlock()

	log.Debug().
		Str("teamName", teamName).
		Msg("Client registered by team name")
}

func (s *TCPServer) SendToClient(teamName string, message interface{}) error {
	s.clientsMu.RLock()
	client, exists := s.clients[teamName]
	s.clientsMu.RUnlock()

	if !exists {
		return fmt.Errorf("client not found: %s", teamName)
	}

	return client.SendMessage(message)
}

func (s *TCPServer) BroadcastToAll(message interface{}) error {
	s.clientsMu.RLock()
	clients := make([]*ClientHandler, 0, len(s.clients))
	for _, client := range s.clients {
		if client.teamName != "" { // Only authenticated clients
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
				Msg("Failed to broadcast message to client")
			lastErr = err
		}
	}

	return lastErr
}

func (s *TCPServer) GetConnectedClients() []string {
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

func (s *TCPServer) Stop() error {
	log.Info().Msg("Stopping TCP server")

	close(s.shutdown)

	if s.listener != nil {
		s.listener.Close()
	}

	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	// Wait with timeout
	select {
	case <-done:
		log.Info().Msg("TCP server stopped gracefully")
	case <-time.After(10 * time.Second):
		log.Warn().Msg("TCP server stop timeout")
	}

	return nil
}
