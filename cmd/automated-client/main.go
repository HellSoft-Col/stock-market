package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/autoclient/config"
	"github.com/HellSoft-Col/stock-market/internal/autoclient/dashboard"
	"github.com/HellSoft-Col/stock-market/internal/autoclient/manager"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	configFile = flag.String("config", "automated-clients.yaml", "Path to configuration file")
	logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	showStats  = flag.Bool("stats", false, "Show statistics periodically (disables TUI)")
	noTUI      = flag.Bool("no-tui", false, "Disable TUI dashboard (logs to console)")
	logFile    = flag.String("log-file", "trading.log", "Log file path (when using TUI)")
)

func main() {
	flag.Parse()

	// Load .env file if it exists (ignore errors if not found)
	envErr := godotenv.Load()

	// Setup logging
	setupLogging(*logLevel)

	if envErr == nil {
		log.Info().Msg("✅ .env file loaded successfully")
	} else {
		log.Warn().Err(envErr).Msg("⚠️ .env file not found (using environment variables)")
	}

	// Verify API keys are set
	deepseekKey := os.Getenv("DEEPSEEK_API_KEY")
	openaiKey := os.Getenv("OPENAI_API_KEY")

	if deepseekKey != "" {
		log.Info().
			Str("keyPrefix", deepseekKey[:min(10, len(deepseekKey))]+"...").
			Int("keyLength", len(deepseekKey)).
			Msg("✅ DEEPSEEK_API_KEY loaded")
	}

	if openaiKey != "" {
		log.Info().
			Str("keyPrefix", openaiKey[:min(15, len(openaiKey))]+"...").
			Int("keyLength", len(openaiKey)).
			Msg("✅ OPENAI_API_KEY loaded")
	}

	if deepseekKey == "" && openaiKey == "" {
		log.Fatal().Msg("❌ No AI API keys found! Please set OPENAI_API_KEY or DEEPSEEK_API_KEY in .env file")
	}

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

	// TUI is DEFAULT, disable with --no-tui or --stats
	if *showStats || *noTUI {
		// Console mode with stats
		if *showStats {
			go statsReporter(clientManager)
		}

		// Wait for interrupt signal
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

		log.Info().Msg("✅ All clients started. Press Ctrl+C to stop.")

		<-sigCh
	} else {
		// DEFAULT: Run TUI dashboard with logs to file
		setupFileLogging(*logFile, *logLevel)
		runDashboard(clientManager)
	}

	log.Info().Msg("Shutdown signal received, stopping clients...")

	// Stop all clients
	if err := clientManager.Stop(); err != nil {
		log.Error().
			Err(err).
			Msg("Error during shutdown")
	}

	log.Info().Msg("👋 Shutdown complete")
}

// setupLogging configures the logger for console output
func setupLogging(level string) {
	// Parse log level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	// Configure zerolog with colors
	zerolog.SetGlobalLevel(logLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    false, // Enable colors
	})

	log.Info().
		Str("level", logLevel.String()).
		Msg("Logging configured")
}

// setupFileLogging configures logging to a file (for TUI mode)
func setupFileLogging(filename, level string) {
	// Parse log level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	// Open log file
	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		os.Exit(1)
	}

	// Configure zerolog to write to file
	zerolog.SetGlobalLevel(logLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        logFile,
		TimeFormat: time.RFC3339,
		NoColor:    true, // No colors in file
	})
}

// runDashboard starts the fancy TUI dashboard
func runDashboard(cm *manager.ClientManager) {
	// Create dashboard
	dash := dashboard.NewDashboard()
	collector := dashboard.NewMetricsCollector(dash)

	// Create context for metrics updater
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start metrics updater
	go metricsUpdater(ctx, cm, collector)

	// Run the TUI
	p := tea.NewProgram(dash, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Error().Err(err).Msg("Error running dashboard")
	}
}

// metricsUpdater continuously updates dashboard metrics from client manager
func metricsUpdater(ctx context.Context, cm *manager.ClientManager, collector *dashboard.MetricsCollector) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := cm.GetStats()

			for _, sessionStats := range stats {
				teamName := sessionStats["id"].(string)
				connected := sessionStats["connected"].(bool)
				authenticated := sessionStats["authenticated"].(bool)

				// Initialize trader if not exists
				agentStats, ok := sessionStats["agent"].(map[string]interface{})
				if !ok {
					continue
				}

				strategyName := "unknown"
				if strategy, ok := sessionStats["strategy"].(string); ok {
					strategyName = strategy
				}

				// Extract metrics
				pnl := agentStats["pnl"].(float64)
				balance := agentStats["balance"].(float64)
				ordersSent := agentStats["ordersSent"].(int)
				fillsReceived := agentStats["fillsReceived"].(int)

				// Calculate inventory value and net worth
				inventoryValue := 0.0
				if invValue, ok := agentStats["inventoryValue"].(float64); ok {
					inventoryValue = invValue
				}

				netWorth := balance + inventoryValue

				// Determine status with granular states
				status := "waiting"
				if connected && authenticated {
					status = "active"
				} else if connected && !authenticated {
					status = "connected" // Connected but not authenticated yet
				} else if !connected {
					status = "connecting" // Trying to connect
				}

				// Get active orders count
				activeOrders := 0
				if active, ok := agentStats["activeOrders"].(int); ok {
					activeOrders = active
				}

				// Get last action
				lastAction := ""
				if action, ok := agentStats["lastAction"].(string); ok {
					lastAction = action
				}

				// Get AI decisions count
				aiDecisions := 0
				if ai, ok := agentStats["aiDecisions"].(int); ok {
					aiDecisions = ai
				}

				// Get production count
				productionCount := 0
				if prod, ok := agentStats["productionCount"].(int); ok {
					productionCount = prod
				}

				// Update metrics
				metrics := &dashboard.TraderMetrics{
					TeamName:        teamName,
					Strategy:        strategyName,
					Balance:         balance,
					InventoryValue:  inventoryValue,
					NetWorth:        netWorth,
					PnL:             0,
					PnLPercent:      pnl,
					OrdersPlaced:    ordersSent,
					FillsReceived:   fillsReceived,
					ActiveOrders:    activeOrders,
					LastAction:      lastAction,
					LastActionTime:  time.Now(),
					Status:          status,
					AIDecisions:     aiDecisions,
					ProductionCount: productionCount,
					Connected:       connected,
					Authenticated:   authenticated,
				}

				collector.UpdateAll(teamName, metrics)
			}
		}
	}
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

				// Format P&L with emoji
				pnlEmoji := "📊"
				if pnl > 0 {
					pnlEmoji = "📈"
				} else if pnl < 0 {
					pnlEmoji = "📉"
				}

				log.Info().
					Str("client", name).
					Str("status", status).
					Str("pnl%", fmt.Sprintf("%s %.2f%%", pnlEmoji, pnl)).
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
