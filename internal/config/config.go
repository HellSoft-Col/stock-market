package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     ServerConfig     `yaml:"server"`
	MongoDB    MongoDBConfig    `yaml:"mongodb"`
	Market     MarketConfig     `yaml:"market"`
	Tournament TournamentConfig `yaml:"tournament"`
	Logging    LoggingConfig    `yaml:"logging"`
	Security   SecurityConfig   `yaml:"security"`
}

type ServerConfig struct {
	Protocol       string        `yaml:"protocol"`
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	MaxConnections int           `yaml:"maxConnections"`
	ReadTimeout    time.Duration `yaml:"readTimeout"`
	WriteTimeout   time.Duration `yaml:"writeTimeout"`
	MaxMessageSize int           `yaml:"maxMessageSize"`
}

type MongoDBConfig struct {
	URI      string        `yaml:"uri"`
	Database string        `yaml:"database"`
	Timeout  time.Duration `yaml:"timeout"`
}

type MarketConfig struct {
	TickerInterval     time.Duration `yaml:"tickerInterval"`
	OfferTimeout       time.Duration `yaml:"offerTimeout"`
	OrderBookDepth     int           `yaml:"orderBookDepth"`
	EnablePartialFills bool          `yaml:"enablePartialFills"`
	TransactionRetries int           `yaml:"transactionRetries"`
}

type TournamentConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Duration time.Duration `yaml:"duration"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

type SecurityConfig struct {
	RateLimitOrdersPerMin     int `yaml:"rateLimitOrdersPerMin"`
	RateLimitConnectionsPerIP int `yaml:"rateLimitConnectionsPerIP"`
	MaxMessageRetries         int `yaml:"maxMessageRetries"`
}

func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables if set
	if mongoURI := os.Getenv("MONGODB_URI"); mongoURI != "" {
		config.MongoDB.URI = mongoURI
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.Logging.Level = logLevel
	}

	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}

	return &config, nil
}
