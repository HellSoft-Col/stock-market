package transport

import (
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/yourusername/avocado-exchange-server/internal/domain"
)

type Broadcaster struct {
	clients   map[string]domain.ClientConnection
	clientsMu sync.RWMutex
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[string]domain.ClientConnection),
	}
}

func (b *Broadcaster) RegisterClient(teamName string, conn domain.ClientConnection) {
	b.clientsMu.Lock()
	defer b.clientsMu.Unlock()

	b.clients[teamName] = conn

	log.Debug().
		Str("teamName", teamName).
		Int("totalClients", len(b.clients)).
		Msg("Client registered with broadcaster")
}

func (b *Broadcaster) UnregisterClient(teamName string) {
	b.clientsMu.Lock()
	defer b.clientsMu.Unlock()

	delete(b.clients, teamName)

	log.Debug().
		Str("teamName", teamName).
		Int("totalClients", len(b.clients)).
		Msg("Client unregistered from broadcaster")
}

func (b *Broadcaster) SendToClient(teamName string, msg any) error {
	b.clientsMu.RLock()
	client, exists := b.clients[teamName]
	b.clientsMu.RUnlock()

	if !exists {
		log.Warn().
			Str("teamName", teamName).
			Msg("Attempted to send message to unregistered client")
		return nil // Don't error for missing clients
	}

	if err := client.SendMessage(msg); err != nil {
		log.Warn().
			Str("teamName", teamName).
			Err(err).
			Msg("Failed to send message to client")

		// Remove dead client
		b.UnregisterClient(teamName)
		return err
	}

	log.Debug().
		Str("teamName", teamName).
		Msg("Message sent to client")

	return nil
}

func (b *Broadcaster) BroadcastToAll(msg any) error {
	b.clientsMu.RLock()
	clients := make(map[string]domain.ClientConnection)
	for teamName, client := range b.clients {
		clients[teamName] = client
	}
	b.clientsMu.RUnlock()

	var lastErr error
	deadClients := make([]string, 0)

	for teamName, client := range clients {
		if err := client.SendMessage(msg); err != nil {
			log.Warn().
				Str("teamName", teamName).
				Err(err).
				Msg("Failed to broadcast message to client")
			lastErr = err
			deadClients = append(deadClients, teamName)
		}
	}

	// Remove dead clients
	for _, teamName := range deadClients {
		b.UnregisterClient(teamName)
	}

	if len(deadClients) > 0 {
		log.Info().
			Int("deadClients", len(deadClients)).
			Int("totalClients", len(clients)).
			Msg("Broadcast completed with some failures")
	} else {
		log.Debug().
			Int("totalClients", len(clients)).
			Msg("Broadcast completed successfully")
	}

	return lastErr
}

func (b *Broadcaster) GetConnectedClients() []string {
	b.clientsMu.RLock()
	defer b.clientsMu.RUnlock()

	teams := make([]string, 0, len(b.clients))
	for teamName := range b.clients {
		teams = append(teams, teamName)
	}

	return teams
}

var _ domain.Broadcaster = (*Broadcaster)(nil)
