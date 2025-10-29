package transport

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"github.com/yourusername/avocado-exchange-server/internal/config"
	"github.com/yourusername/avocado-exchange-server/internal/domain"
)

type WebSocketClientHandler struct {
	conn     *websocket.Conn
	config   *config.Config
	server   *WebSocketServer
	router   *MessageRouter
	teamName string
}

func NewWebSocketClientHandler(conn *websocket.Conn, config *config.Config, server *WebSocketServer, router *MessageRouter) *WebSocketClientHandler {
	return &WebSocketClientHandler{
		conn:   conn,
		config: config,
		server: server,
		router: router,
	}
}

func (c *WebSocketClientHandler) Handle() {
	if c.config.Server.ReadTimeout > 0 {
		_ = c.conn.SetReadDeadline(time.Now().Add(c.config.Server.ReadTimeout))
	}

	c.conn.SetPongHandler(func(string) error {
		if c.config.Server.ReadTimeout > 0 {
			if err := c.conn.SetReadDeadline(time.Now().Add(c.config.Server.ReadTimeout)); err != nil {
				log.Error().Err(err).Msg("Failed to set read deadline in pong handler")
			}
		}
		return nil
	})

	for {
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().
					Str("clientAddr", c.conn.RemoteAddr().String()).
					Str("teamName", c.teamName).
					Err(err).
					Msg("WebSocket read error")
			}
			break
		}

		if messageType != websocket.TextMessage {
			continue
		}

		line := strings.TrimSpace(string(message))
		if line == "" {
			continue
		}

		if c.config.Server.ReadTimeout > 0 {
			if err := c.conn.SetReadDeadline(time.Now().Add(c.config.Server.ReadTimeout)); err != nil {
				log.Error().Err(err).Msg("Failed to set read deadline")
				break
			}
		}

		if err := c.handleMessage(line); err != nil {
			log.Error().
				Str("clientAddr", c.conn.RemoteAddr().String()).
				Str("teamName", c.teamName).
				Str("message", line).
				Err(err).
				Msg("Error handling WebSocket message")

			errorMsg := &domain.ErrorMessage{
				Type:      "ERROR",
				Code:      domain.ErrInvalidMessage,
				Reason:    fmt.Sprintf("Message processing failed: %v", err),
				Timestamp: time.Now(),
			}
			_ = c.SendMessage(errorMsg)
		}
	}
}

func (c *WebSocketClientHandler) handleMessage(rawMessage string) error {
	log.Debug().
		Str("clientAddr", c.conn.RemoteAddr().String()).
		Str("teamName", c.teamName).
		Str("message", rawMessage).
		Msg("Received WebSocket message")

	ctx := context.Background()
	return c.router.RouteMessage(ctx, rawMessage, c)
}

func (c *WebSocketClientHandler) SendMessage(message any) error {
	if c.config.Server.WriteTimeout > 0 {
		_ = c.conn.SetWriteDeadline(time.Now().Add(c.config.Server.WriteTimeout))
	}

	if err := c.conn.WriteJSON(message); err != nil {
		return fmt.Errorf("failed to send WebSocket message: %w", err)
	}

	log.Debug().
		Str("clientAddr", c.conn.RemoteAddr().String()).
		Str("teamName", c.teamName).
		Interface("message", message).
		Msg("Sent WebSocket message")

	return nil
}

func (c *WebSocketClientHandler) GetTeamName() string {
	return c.teamName
}

func (c *WebSocketClientHandler) SetTeamName(teamName string) {
	c.teamName = teamName
}

func (c *WebSocketClientHandler) GetRemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func (c *WebSocketClientHandler) RegisterWithServer(teamName string) {
	c.server.RegisterClientByTeam(teamName, c)
}

func (c *WebSocketClientHandler) Close() error {
	return c.conn.Close()
}

func (c *WebSocketClientHandler) Cleanup() {
	// This should be called when the client disconnects
	// to clean up session tracking
}
