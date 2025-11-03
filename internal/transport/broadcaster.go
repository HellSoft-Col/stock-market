package transport

import (
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/HellSoft-Col/stock-market/internal/domain"
)

type Broadcaster struct {
	clients   map[string][]domain.ClientConnection // teamName -> list of connections
	clientsMu sync.RWMutex
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[string][]domain.ClientConnection),
	}
}

func (b *Broadcaster) RegisterClient(teamName string, conn domain.ClientConnection) {
	b.clientsMu.Lock()
	defer b.clientsMu.Unlock()

	// Add connection to the team's connection list
	b.clients[teamName] = append(b.clients[teamName], conn)

	totalConnections := 0
	for _, connections := range b.clients {
		totalConnections += len(connections)
	}

	log.Debug().
		Str("teamName", teamName).
		Int("teamConnections", len(b.clients[teamName])).
		Int("totalConnections", totalConnections).
		Msg("Client registered with broadcaster")
}

func (b *Broadcaster) UnregisterClient(teamName string) {
	b.clientsMu.Lock()
	defer b.clientsMu.Unlock()

	delete(b.clients, teamName)

	totalConnections := 0
	for _, connections := range b.clients {
		totalConnections += len(connections)
	}

	log.Debug().
		Str("teamName", teamName).
		Int("totalConnections", totalConnections).
		Msg("All clients unregistered for team")
}

func (b *Broadcaster) UnregisterSpecificClient(teamName string, conn domain.ClientConnection) {
	b.clientsMu.Lock()
	defer b.clientsMu.Unlock()

	connections := b.clients[teamName]
	for i, client := range connections {
		if client == conn {
			// Remove this specific connection
			b.clients[teamName] = append(connections[:i], connections[i+1:]...)
			break
		}
	}

	// If no connections left for this team, remove the team entry
	if len(b.clients[teamName]) == 0 {
		delete(b.clients, teamName)
	}

	totalConnections := 0
	for _, connections := range b.clients {
		totalConnections += len(connections)
	}

	log.Debug().
		Str("teamName", teamName).
		Int("teamConnections", len(b.clients[teamName])).
		Int("totalConnections", totalConnections).
		Msg("Specific client unregistered")
}

func (b *Broadcaster) SendToClient(teamName string, msg any) error {
	b.clientsMu.RLock()
	connections, exists := b.clients[teamName]
	b.clientsMu.RUnlock()

	if !exists || len(connections) == 0 {
		log.Warn().
			Str("teamName", teamName).
			Msg("Attempted to send message to unregistered team")
		return nil // Don't error for missing clients
	}

	var lastErr error
	deadConnections := make([]domain.ClientConnection, 0)

	// Send to all connections for this team
	for _, client := range connections {
		if err := client.SendMessage(msg); err != nil {
			log.Warn().
				Str("teamName", teamName).
				Err(err).
				Msg("Failed to send message to one client connection")
			lastErr = err
			deadConnections = append(deadConnections, client)
		}
	}

	// Remove dead connections
	for _, deadConn := range deadConnections {
		b.UnregisterSpecificClient(teamName, deadConn)
	}

	log.Debug().
		Str("teamName", teamName).
		Int("connections", len(connections)).
		Int("failed", len(deadConnections)).
		Msg("Message sent to team clients")

	return lastErr
}

func (b *Broadcaster) BroadcastToAll(msg any) error {
	b.clientsMu.RLock()
	allConnections := make(map[string][]domain.ClientConnection)
	for teamName, connections := range b.clients {
		// Copy the slice to avoid race conditions
		teamConnections := make([]domain.ClientConnection, len(connections))
		copy(teamConnections, connections)
		allConnections[teamName] = teamConnections
	}
	b.clientsMu.RUnlock()

	var lastErr error
	deadConnections := make(map[string][]domain.ClientConnection)
	totalSent := 0
	totalFailed := 0

	for teamName, connections := range allConnections {
		for _, client := range connections {
			if err := client.SendMessage(msg); err != nil {
				log.Warn().
					Str("teamName", teamName).
					Err(err).
					Msg("Failed to broadcast message to client")
				lastErr = err
				deadConnections[teamName] = append(deadConnections[teamName], client)
				totalFailed++
			} else {
				totalSent++
			}
		}
	}

	// Remove dead connections
	for teamName, deadConns := range deadConnections {
		for _, deadConn := range deadConns {
			b.UnregisterSpecificClient(teamName, deadConn)
		}
	}

	if totalFailed > 0 {
		log.Info().
			Int("sent", totalSent).
			Int("failed", totalFailed).
			Msg("Broadcast completed with some failures")
	} else {
		log.Debug().
			Int("sent", totalSent).
			Msg("Broadcast completed successfully")
	}

	return lastErr
}

func (b *Broadcaster) GetConnectedClients() []string {
	b.clientsMu.RLock()
	defer b.clientsMu.RUnlock()

	var teams []string
	for teamName, connections := range b.clients {
		// Add team name for each connection (to show multiple connections)
		for range connections {
			teams = append(teams, teamName)
		}
	}

	return teams
}

var _ domain.Broadcaster = (*Broadcaster)(nil)
