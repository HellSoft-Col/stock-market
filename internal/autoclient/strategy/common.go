package strategy

import (
	"fmt"
	"sync"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/google/uuid"
)

// GenerateOrderID generates a unique order ID
func GenerateOrderID() string {
	return fmt.Sprintf("ORD-%s", uuid.New().String()[:8])
}

// CreateBuyOrder creates a market buy order
func CreateBuyOrder(product string, quantity int, message string) *domain.OrderMessage {
	// Use provided message or generate one
	if message == "" {
		message = GetMessageGenerator().GenerateOrderMessage("BUY", product, quantity, 0)
	}
	return &domain.OrderMessage{
		Type:    "ORDER",
		ClOrdID: GenerateOrderID(),
		Side:    "BUY",
		Mode:    "MARKET",
		Product: product,
		Qty:     quantity,
		Message: message,
	}
}

// CreateSellOrder creates a market sell order
func CreateSellOrder(product string, quantity int, message string) *domain.OrderMessage {
	// Use provided message or generate one
	if message == "" {
		message = GetMessageGenerator().GenerateOrderMessage("SELL", product, quantity, 0)
	}
	return &domain.OrderMessage{
		Type:    "ORDER",
		ClOrdID: GenerateOrderID(),
		Side:    "SELL",
		Mode:    "MARKET",
		Product: product,
		Qty:     quantity,
		Message: message,
	}
}

// CreateLimitBuyOrder creates a limit buy order
func CreateLimitBuyOrder(product string, quantity int, price float64, message string) *domain.OrderMessage {
	// Use provided message or generate one
	if message == "" {
		message = GetMessageGenerator().GenerateOrderMessage("BUY", product, quantity, price)
	}
	return &domain.OrderMessage{
		Type:       "ORDER",
		ClOrdID:    GenerateOrderID(),
		Side:       "BUY",
		Mode:       "LIMIT",
		Product:    product,
		Qty:        quantity,
		LimitPrice: &price,
		Message:    message,
	}
}

// CreateLimitSellOrder creates a limit sell order
func CreateLimitSellOrder(product string, quantity int, price float64, message string) *domain.OrderMessage {
	// Use provided message or generate one
	if message == "" {
		message = GetMessageGenerator().GenerateOrderMessage("SELL", product, quantity, price)
	}
	return &domain.OrderMessage{
		Type:       "ORDER",
		ClOrdID:    GenerateOrderID(),
		Side:       "SELL",
		Mode:       "LIMIT",
		Product:    product,
		Qty:        quantity,
		LimitPrice: &price,
		Message:    message,
	}
}

// CreateProduction creates a production update message
func CreateProduction(product string, quantity int) *domain.ProductionUpdateMessage {
	return &domain.ProductionUpdateMessage{
		Type:     "PRODUCTION",
		Product:  product,
		Quantity: quantity,
	}
}

// CreateAcceptOffer creates an accept offer message
func CreateAcceptOffer(offerID string, quantity int, price float64) *domain.AcceptOfferMessage {
	return &domain.AcceptOfferMessage{
		Type:            "ACCEPT_OFFER",
		OfferID:         offerID,
		Accept:          true,
		QuantityOffered: quantity,
		PriceOffered:    price,
	}
}

// GetConfigString extracts a string from config map
func GetConfigString(config map[string]interface{}, key string, defaultValue string) string {
	if val, ok := config[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

// GetConfigInt extracts an int from config map
func GetConfigInt(config map[string]interface{}, key string, defaultValue int) int {
	if val, ok := config[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return defaultValue
}

// GetConfigFloat extracts a float64 from config map
func GetConfigFloat(config map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := config[key]; ok {
		if floatVal, ok := val.(float64); ok {
			return floatVal
		}
	}
	return defaultValue
}

// GetConfigBool extracts a bool from config map
func GetConfigBool(config map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := config[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}

// GetConfigDuration extracts a duration from config map
func GetConfigDuration(config map[string]interface{}, key string, defaultValue time.Duration) time.Duration {
	if val, ok := config[key]; ok {
		if strVal, ok := val.(string); ok {
			if dur, err := time.ParseDuration(strVal); err == nil {
				return dur
			}
		}
	}
	return defaultValue
}

// GetConfigStringSlice extracts a string slice from config map
func GetConfigStringSlice(config map[string]interface{}, key string) []string {
	if val, ok := config[key]; ok {
		if slice, ok := val.([]interface{}); ok {
			result := make([]string, 0, len(slice))
			for _, item := range slice {
				if strVal, ok := item.(string); ok {
					result = append(result, strVal)
				}
			}
			return result
		}
	}
	return []string{}
}

// Global message generator instance
var (
	globalMessageGenerator *MessageGenerator
	messageGeneratorOnce   sync.Once
)

// InitMessageGenerator initializes the global message generator
func InitMessageGenerator(teamName, apiKey string) {
	messageGeneratorOnce.Do(func() {
		globalMessageGenerator = NewMessageGenerator(teamName, apiKey)
	})
}

// GetMessageGenerator returns the global message generator
func GetMessageGenerator() *MessageGenerator {
	if globalMessageGenerator == nil {
		// Fallback with no AI
		globalMessageGenerator = NewMessageGenerator("Team", "")
	}
	return globalMessageGenerator
}
