package strategy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// MessageGenerator generates funny trading messages using DeepSeek or fallback templates
type MessageGenerator struct {
	apiKey      string
	endpoint    string
	httpClient  *http.Client
	teamName    string
	useAI       bool
	rng         *rand.Rand
	mu          sync.Mutex
	lastAICall  time.Time
	minInterval time.Duration
}

// NewMessageGenerator creates a new message generator
func NewMessageGenerator(teamName, apiKey string) *MessageGenerator {
	return &MessageGenerator{
		apiKey:      apiKey,
		endpoint:    "https://api.deepseek.com/v1/chat/completions",
		teamName:    teamName,
		useAI:       apiKey != "",
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
		minInterval: 2 * time.Second, // Rate limit: max 1 message per 2 seconds
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GenerateOrderMessage generates a funny message for an order
func (mg *MessageGenerator) GenerateOrderMessage(action, product string, quantity int, price float64) string {
	mg.mu.Lock()
	defer mg.mu.Unlock()

	// Use AI if available and not rate limited
	if mg.useAI && time.Since(mg.lastAICall) >= mg.minInterval {
		aiMsg := mg.generateAIMessage(action, product, quantity, price)
		if aiMsg != "" {
			mg.lastAICall = time.Now()
			return aiMsg
		}
	}

	// Fallback to template-based messages
	return mg.generateTemplateMessage(action, product, quantity, price)
}

// generateAIMessage calls DeepSeek to generate a creative message
func (mg *MessageGenerator) generateAIMessage(action, product string, quantity int, price float64) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	prompt := fmt.Sprintf(`Genera un mensaje CORTO (máximo 60 caracteres) en español para una orden de trading.

Contexto:
- Equipo: %s
- Acción: %s
- Producto: %s (aguacates espaciales)
- Cantidad: %d
- Precio: $%.2f

Requisitos:
1. Máximo 60 caracteres (IMPORTANTE)
2. Usar jerga de "31 minutos" (Tulio, Bodoque, Chile, etc)
3. Usar jerga andoriana (intergaláctica, espacial)
4. Referencias a aguacates
5. Incluir 1-2 emojis relevantes
6. Tono divertido pero profesional
7. NO usar hashtags

Ejemplos:
- "🥑 Tulio recomienda comprar! -Los Someliers"
- "💫 Aguacates al infinito! -Bodoque Team"
- "🚀 Palta power activado -Andorianos"

Responde SOLO con el mensaje, sin explicaciones.`, mg.teamName, action, product, quantity, price)

	request := DeepSeekRequest{
		Model: "deepseek-chat",
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		Temperature: 0.9, // High creativity
		MaxTokens:   50,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return ""
	}

	req, err := http.NewRequestWithContext(ctx, "POST", mg.endpoint, bytes.NewBuffer(body))
	if err != nil {
		return ""
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+mg.apiKey)

	resp, err := mg.httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var deepseekResp DeepSeekResponse
	if err := json.Unmarshal(respBody, &deepseekResp); err != nil {
		return ""
	}

	if len(deepseekResp.Choices) == 0 {
		return ""
	}

	message := strings.TrimSpace(deepseekResp.Choices[0].Message.Content)
	message = strings.Trim(message, `"'`)

	// Ensure it's not too long
	if len(message) > 60 {
		message = message[:57] + "..."
	}

	log.Debug().
		Str("message", message).
		Msg("Generated AI message")

	return message
}

// generateTemplateMessage generates a message using templates
func (mg *MessageGenerator) generateTemplateMessage(action, product string, quantity int, price float64) string {
	templates := mg.getTemplates(action, product)
	if len(templates) == 0 {
		return fmt.Sprintf("-%s", mg.teamName)
	}

	template := templates[mg.rng.Intn(len(templates))]
	message := fmt.Sprintf(template, mg.teamName)

	// Ensure it's not too long
	if len(message) > 60 {
		message = message[:57] + "..."
	}

	return message
}

// getTemplates returns message templates based on action and product
func (mg *MessageGenerator) getTemplates(action, product string) []string {
	isBuy := action == "BUY" || action == "COMPRAR"

	// Common templates for all products
	common := []string{
		// 31 Minutos references
		"🎭 Tulio aprueba esto! -%s",
		"🐰 Bodoque al poder! -%s",
		"📺 31 minutos de aguacates -%s",
		"🎪 Calcetín recomienda! -%s",
		"🦁 Mico el Micófono dice: GO! -%s",

		// Andoriano/Space references
		"🚀 Propulsores al máximo -%s",
		"💫 Energía intergaláctica -%s",
		"🌌 Comercio espacial activo -%s",
		"⚡ Hiperespacio activado -%s",
		"🛸 Nave lista para comercio -%s",

		// Aguacate references
		"🥑 Palta power! -%s",
		"🥑💚 Aguacate supremo -%s",
		"🥑✨ Oro verde espacial -%s",
		"🥑🚀 Guacamole galáctico -%s",
		"🥑⭐ Palta premium -%s",
	}

	if isBuy {
		buy := []string{
			"💰 Comprando como Tulio! -%s",
			"🛒 Bodoque va de shopping -%s",
			"🥑 Acumulando aguacates -%s",
			"💸 Invirtiendo en el futuro -%s",
			"🎯 Target adquirido -%s",
			"🌟 A por el oro verde! -%s",
			"🚀 Subiendo a bordo -%s",
			"💎 Diamantes verdes! -%s",
			"🎪 Comprando el show -%s",
			"🥑🔥 Palta caliente! -%s",
			"💫 Energía compradora -%s",
			"🛸 Cargando nave -%s",
			"⚡ Compra relámpago -%s",
			"🎭 Tulio dice: COMPRA! -%s",
			"🐰 Bodoque acumula -%s",
		}
		return append(common, buy...)
	} else {
		sell := []string{
			"💵 Vendiendo con estilo -%s",
			"📈 Profit time! -%s",
			"🥑💰 Cosechando ganancias -%s",
			"✨ Liquidando posición -%s",
			"🎯 Objetivo cumplido -%s",
			"🌟 Momento de brillar -%s",
			"🚀 Despegando hacia profit -%s",
			"💎 Materializando valor -%s",
			"🎪 El show continúa -%s",
			"🥑📊 Exportando aguacates -%s",
			"💫 Energía vendedora -%s",
			"🛸 Descargando carga -%s",
			"⚡ Venta estelar -%s",
			"🎭 Tulio vende bien! -%s",
			"🐰 Bodoque hace caja -%s",
		}
		return append(common, sell...)
	}
}

// Event represents a trading event for context
type TradingEvent struct {
	Timestamp time.Time
	Type      string // "FILL", "ORDER", "PRODUCTION", "OFFER"
	Action    string // "BUY", "SELL", "PRODUCE", "ACCEPT", "REJECT"
	Product   string
	Quantity  int
	Price     float64
	PnL       float64
	Message   string
}

// EventHistory tracks recent trading events for context
type EventHistory struct {
	events    []TradingEvent
	maxEvents int
	mu        sync.RWMutex
}

// NewEventHistory creates a new event history tracker
func NewEventHistory(maxEvents int) *EventHistory {
	return &EventHistory{
		events:    make([]TradingEvent, 0, maxEvents),
		maxEvents: maxEvents,
	}
}

// AddEvent adds an event to the history
func (eh *EventHistory) AddEvent(event TradingEvent) {
	eh.mu.Lock()
	defer eh.mu.Unlock()

	eh.events = append(eh.events, event)

	// Keep only recent events
	if len(eh.events) > eh.maxEvents {
		eh.events = eh.events[len(eh.events)-eh.maxEvents:]
	}
}

// GetRecent returns recent events
func (eh *EventHistory) GetRecent(count int) []TradingEvent {
	eh.mu.RLock()
	defer eh.mu.RUnlock()

	if count > len(eh.events) {
		count = len(eh.events)
	}

	if count == 0 {
		return []TradingEvent{}
	}

	// Return last 'count' events
	return eh.events[len(eh.events)-count:]
}

// GetSummary returns a text summary of recent events
func (eh *EventHistory) GetSummary() string {
	events := eh.GetRecent(20)
	if len(events) == 0 {
		return "No recent events"
	}

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("\nRECENT ACTIVITY (%d events):\n", len(events)))

	// Group by type
	fills := 0
	orders := 0
	productions := 0
	totalPnL := 0.0

	for _, event := range events {
		switch event.Type {
		case "FILL":
			fills++
			totalPnL += event.PnL
		case "ORDER":
			orders++
		case "PRODUCTION":
			productions++
		}
	}

	summary.WriteString(fmt.Sprintf("  Orders Sent: %d\n", orders))
	summary.WriteString(fmt.Sprintf("  Fills Received: %d\n", fills))
	summary.WriteString(fmt.Sprintf("  Productions: %d\n", productions))
	summary.WriteString(fmt.Sprintf("  Current Session P&L: $%.2f\n", totalPnL))

	// Show last 5 events
	summary.WriteString("\nLast 5 Events:\n")
	start := len(events) - 5
	if start < 0 {
		start = 0
	}
	for _, event := range events[start:] {
		elapsed := time.Since(event.Timestamp)
		summary.WriteString(fmt.Sprintf("  [%s ago] %s %s: %s x%d @ $%.2f\n",
			formatDuration(elapsed),
			event.Type,
			event.Action,
			event.Product,
			event.Quantity,
			event.Price))
	}

	return summary.String()
}

// formatDuration formats a duration in human-readable form
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh", int(d.Hours()))
}
