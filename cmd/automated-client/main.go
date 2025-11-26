package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/autoclient/config"
	"github.com/HellSoft-Col/stock-market/internal/autoclient/manager"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	configFile = flag.String("config", "automated-clients.yaml", "Path to configuration file")
	logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	showStats  = flag.Bool("stats", false, "Show statistics periodically")
)

func main() {
	flag.Parse()

	// Load .env file if it exists (ignore errors if not found)
	_ = godotenv.Load()

	// Setup logging
	setupLogging(*logLevel)

	log.Info().
		Str("config", *configFile).
		Msg("🤖 Starting Automated Trading Clients")

	// Load configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to load configuration")
	}

	log.Info().
		Int("enabledClients", len(cfg.GetEnabledClients())).
		Str("server", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)).
		Msg("Configuration loaded")

	// Create client manager
	clientManager := manager.NewClientManager(cfg)

	// Start all clients
	if err := clientManager.Start(); err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to start clients")
	}

	// Start statistics reporter if enabled
	if *showStats {
		go statsReporter(clientManager)
	}

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	log.Info().Msg("✅ All clients started. Press Ctrl+C to stop.")

	<-sigCh

	log.Info().Msg("Shutdown signal received, stopping clients...")

	// Stop all clients
	if err := clientManager.Stop(); err != nil {
		log.Error().
			Err(err).
			Msg("Error during shutdown")
	}

	log.Info().Msg("👋 Shutdown complete")
}

// setupLogging configures the logger
func setupLogging(level string) {
	// Parse log level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	// Configure zerolog
	zerolog.SetGlobalLevel(logLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	})

	log.Info().
		Str("level", logLevel.String()).
		Msg("Logging configured")
}

// statsReporter periodically reports statistics
func statsReporter(cm *manager.ClientManager) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := cm.GetStats()
		connectedCount := cm.GetConnectedCount()

		log.Info().
			Int("total", len(stats)).
			Int("connected", connectedCount).
			Msg("📊 Client Statistics")

		for _, sessionStats := range stats {
			name := sessionStats["id"].(string)
			connected := sessionStats["connected"].(bool)
			authenticated := sessionStats["authenticated"].(bool)

			status := "❌ Disconnected"
			if connected && authenticated {
				status = "✅ Connected"
			}

			agentStats, ok := sessionStats["agent"].(map[string]interface{})
			if ok {
				pnl := agentStats["pnl"].(float64)
				balance := agentStats["balance"].(float64)
				ordersSent := agentStats["ordersSent"].(int)
				fillsReceived := agentStats["fillsReceived"].(int)

				log.Info().
					Str("client", name).
					Str("status", status).
					Float64("pnl", pnl).
					Float64("balance", balance).
					Int("orders", ordersSent).
					Int("fills", fillsReceived).
					Msg("  └─")
			} else {
				log.Info().
					Str("client", name).
					Str("status", status).
					Msg("  └─")
			}
		}
	}
}
