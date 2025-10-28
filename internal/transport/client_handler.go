package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/yourusername/avocado-exchange-server/internal/config"
	"github.com/yourusername/avocado-exchange-server/internal/domain"
)

type ClientHandler struct {
	conn     net.Conn
	config   *config.Config
	server   *TCPServer
	router   *MessageRouter
	teamName string
	scanner  *bufio.Scanner
	encoder  *json.Encoder
}

func NewClientHandler(conn net.Conn, config *config.Config, server *TCPServer, router *MessageRouter) *ClientHandler {
	return &ClientHandler{
		conn:    conn,
		config:  config,
		server:  server,
		router:  router,
		scanner: bufio.NewScanner(conn),
		encoder: json.NewEncoder(conn),
	}
}

func (c *ClientHandler) Handle() {
	// Set read timeout
	if c.config.Server.ReadTimeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(c.config.Server.ReadTimeout))
	}

	for c.scanner.Scan() {
		line := strings.TrimSpace(c.scanner.Text())
		if line == "" {
			continue
		}

		// Reset read timeout for each message
		if c.config.Server.ReadTimeout > 0 {
			c.conn.SetReadDeadline(time.Now().Add(c.config.Server.ReadTimeout))
		}

		if err := c.handleMessage(line); err != nil {
			log.Error().
				Str("clientAddr", c.conn.RemoteAddr().String()).
				Str("teamName", c.teamName).
				Str("message", line).
				Err(err).
				Msg("Error handling message")

			// Send error response
			errorMsg := &domain.ErrorMessage{
				Type:      "ERROR",
				Code:      domain.ErrInvalidMessage,
				Reason:    fmt.Sprintf("Message processing failed: %v", err),
				Timestamp: time.Now(),
			}
			_ = c.SendMessage(errorMsg)
		}
	}

	if err := c.scanner.Err(); err != nil {
		log.Error().
			Str("clientAddr", c.conn.RemoteAddr().String()).
			Str("teamName", c.teamName).
			Err(err).
			Msg("Scanner error")
	}
}

func (c *ClientHandler) handleMessage(rawMessage string) error {
	log.Debug().
		Str("clientAddr", c.conn.RemoteAddr().String()).
		Str("teamName", c.teamName).
		Str("message", rawMessage).
		Msg("Received message")

	// Route message through message router
	ctx := context.Background()
	return c.router.RouteMessage(ctx, rawMessage, c)
}

func (c *ClientHandler) SendMessage(message any) error {
	// Set write timeout
	if c.config.Server.WriteTimeout > 0 {
		c.conn.SetWriteDeadline(time.Now().Add(c.config.Server.WriteTimeout))
	}

	if err := c.encoder.Encode(message); err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}

	log.Debug().
		Str("clientAddr", c.conn.RemoteAddr().String()).
		Str("teamName", c.teamName).
		Interface("message", message).
		Msg("Sent message")

	return nil
}

func (c *ClientHandler) GetTeamName() string {
	return c.teamName
}

func (c *ClientHandler) SetTeamName(teamName string) {
	c.teamName = teamName
}

func (c *ClientHandler) GetRemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func (c *ClientHandler) RegisterWithServer(teamName string) {
	c.server.RegisterClientByTeam(teamName, c)
}

func (c *ClientHandler) Close() error {
	return c.conn.Close()
}
