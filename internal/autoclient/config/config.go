package config

import (
	"fmt"
	"os"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/autoclient/production"
	"gopkg.in/yaml.v3"
)

// Config is the root configuration structure
type Config struct {
	Server       ServerConfig                  `yaml:"server"`
	DeepSeek     DeepSeekConfig                `yaml:"deepseek"`
	Global       GlobalConfig                  `yaml:"global"`
	SpeciesRoles map[string]*production.Role   `yaml:"species_roles"`
	Recipes      map[string]*production.Recipe `yaml:"recipes"`
	Clients      []ClientConfig                `yaml:"clients"`
}

// ServerConfig contains server connection settings
type ServerConfig struct {
	Host                 string        `yaml:"host"`
	Port                 int           `yaml:"port"`
	UseSSL               bool          `yaml:"useSSL"`
	ReconnectInterval    time.Duration `yaml:"reconnectInterval"`
	MaxReconnectAttempts int           `yaml:"maxReconnectAttempts"`
}

// DeepSeekConfig contains DeepSeek API settings
type DeepSeekConfig struct {
	APIKey     string        `yaml:"apiKey"`
	Endpoint   string        `yaml:"endpoint"`
	Timeout    time.Duration `yaml:"timeout"`
	MaxRetries int           `yaml:"maxRetries"`
}

// GlobalConfig contains global risk limits and trading settings
type GlobalConfig struct {
	RiskLimits           RiskLimits    `yaml:"riskLimits"`
	TradingPace          time.Duration `yaml:"tradingPace"`          // How often to execute strategy (e.g., 5s, 10s)
	MinTimeBetweenOrders time.Duration `yaml:"minTimeBetweenOrders"` // Minimum delay between sending orders (e.g., 500ms, 1s)
}

// RiskLimits defines trading risk constraints
type RiskLimits struct {
	MaxOrderSize    int     `yaml:"maxOrderSize"`
	MaxPositionSize int     `yaml:"maxPositionSize"`
	MaxDailyLoss    float64 `yaml:"maxDailyLoss"`
}

// ClientConfig represents an automated client configuration
type ClientConfig struct {
	Name                 string                 `yaml:"name"`
	Token                string                 `yaml:"token"`
	Species              string                 `yaml:"species"`
	Strategy             string                 `yaml:"strategy"`
	Enabled              bool                   `yaml:"enabled"`
	TradingPace          *time.Duration         `yaml:"tradingPace,omitempty"`          // Optional: Override global trading pace
	MinTimeBetweenOrders *time.Duration         `yaml:"minTimeBetweenOrders,omitempty"` // Optional: Override global min time between orders
	Config               map[string]interface{} `yaml:"config"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables in the YAML content
	expanded := os.ExpandEnv(string(data))

	var config Config
	if err := yaml.Unmarshal([]byte(expanded), &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Host == "" {
		return fmt.Errorf("server host is required")
	}

	if c.Server.Port == 0 {
		return fmt.Errorf("server port is required")
	}

	if len(c.Clients) == 0 {
		return fmt.Errorf("at least one client configuration is required")
	}

	for i, client := range c.Clients {
		if client.Token == "" {
			return fmt.Errorf("client %d: token is required", i)
		}

		if client.Strategy == "" {
			return fmt.Errorf("client %d: strategy is required", i)
		}

		if client.Species != "" {
			if _, exists := c.SpeciesRoles[client.Species]; !exists {
				return fmt.Errorf("client %d: unknown species '%s'", i, client.Species)
			}
		}
	}

	return nil
}

// GetEnabledClients returns only enabled client configurations
func (c *Config) GetEnabledClients() []ClientConfig {
	enabled := make([]ClientConfig, 0, len(c.Clients))
	for _, client := range c.Clients {
		if client.Enabled {
			enabled = append(enabled, client)
		}
	}
	return enabled
}

// GetRole returns the production role for a species
func (c *Config) GetRole(species string) (*production.Role, error) {
	role, exists := c.SpeciesRoles[species]
	if !exists {
		return nil, fmt.Errorf("role not found for species: %s", species)
	}
	return role, nil
}

// GetTradingPace returns the effective trading pace for a client
// If client has specific trading pace, use it; otherwise use global
func (c *Config) GetTradingPace(clientCfg ClientConfig) time.Duration {
	if clientCfg.TradingPace != nil {
		return *clientCfg.TradingPace
	}
	if c.Global.TradingPace > 0 {
		return c.Global.TradingPace
	}
	// Default to 1 second if not configured
	return 1 * time.Second
}

// GetMinTimeBetweenOrders returns the effective minimum time between orders for a client
// If client has specific setting, use it; otherwise use global
func (c *Config) GetMinTimeBetweenOrders(clientCfg ClientConfig) time.Duration {
	if clientCfg.MinTimeBetweenOrders != nil {
		return *clientCfg.MinTimeBetweenOrders
	}
	if c.Global.MinTimeBetweenOrders > 0 {
		return c.Global.MinTimeBetweenOrders
	}
	// Default to no delay (0) if not configured
	return 0
}
