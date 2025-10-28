package client

import (
	"fmt"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketClient struct {
	conn *websocket.Conn
	url  string
}

func NewWebSocketClient(host string) *WebSocketClient {
	return &WebSocketClient{
		url: fmt.Sprintf("ws://%s/ws", host),
	}
}

func (c *WebSocketClient) Connect() error {
	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	return nil
}

func (c *WebSocketClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *WebSocketClient) SendMessage(message interface{}) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}
	return c.conn.WriteJSON(message)
}

func (c *WebSocketClient) ReadMessage(v interface{}) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}
	return c.conn.ReadJSON(v)
}

func (c *WebSocketClient) SetReadDeadline(t time.Time) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}
	return c.conn.SetReadDeadline(t)
}

func (c *WebSocketClient) Login(token string) (map[string]any, error) {
	loginMessage := map[string]any{
		"type":  "LOGIN",
		"token": token,
	}

	if err := c.SendMessage(loginMessage); err != nil {
		return nil, fmt.Errorf("failed to send login: %w", err)
	}

	c.SetReadDeadline(time.Now().Add(5 * time.Second))

	var response map[string]any
	if err := c.ReadMessage(&response); err != nil {
		return nil, fmt.Errorf("failed to read login response: %w", err)
	}

	if responseType, ok := response["type"].(string); !ok || responseType != "LOGIN_OK" {
		return response, fmt.Errorf("login failed: %+v", response)
	}

	return response, nil
}
