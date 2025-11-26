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

// GlobalConfig contains global risk limits
type GlobalConfig struct {
	RiskLimits RiskLimits `yaml:"riskLimits"`
}

// RiskLimits defines trading risk constraints
type RiskLimits struct {
	MaxOrderSize    int     `yaml:"maxOrderSize"`
	MaxPositionSize int     `yaml:"maxPositionSize"`
	MaxDailyLoss    float64 `yaml:"maxDailyLoss"`
}

// ClientConfig represents an automated client configuration
type ClientConfig struct {
	Name     string                 `yaml:"name"`
	Token    string                 `yaml:"token"`
	Species  string                 `yaml:"species"`
	Strategy string                 `yaml:"strategy"`
	Enabled  bool                   `yaml:"enabled"`
	Config   map[string]interface{} `yaml:"config"`
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
