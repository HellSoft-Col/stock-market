package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/yourusername/avocado-exchange-server/internal/config"
	"github.com/yourusername/avocado-exchange-server/internal/market"
	"github.com/yourusername/avocado-exchange-server/internal/repository/memory"
	"github.com/yourusername/avocado-exchange-server/internal/repository/mongodb"
	"github.com/yourusername/avocado-exchange-server/internal/service"
	"github.com/yourusername/avocado-exchange-server/internal/transport"
	"github.com/yourusername/avocado-exchange-server/pkg/logger"
)

func main() {
	var configFile = flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize logger
	logger.InitLogger(cfg.Logging.Level, cfg.Logging.Format)

	log.Info().
		Str("version", "1.0.0").
		Str("protocol", cfg.Server.Protocol).
		Int("port", cfg.Server.Port).
		Msg("Starting Intergalactic Avocado Stock Exchange Server")

	// Connect to database
	db := mongodb.NewDatabase(&cfg.MongoDB)
	ctx := context.Background()

	if err := db.Connect(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := db.Close(ctx); err != nil {
			log.Error().Err(err).Msg("Error closing database connection")
		}
	}()

	// Test database connection
	if err := db.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping database")
	}

	log.Info().Msg("Database connection established")

	// Create repositories
	teamRepo := mongodb.NewTeamRepository(db.GetDatabase())
	orderRepo := mongodb.NewOrderRepository(db.GetDatabase())
	fillRepo := mongodb.NewFillRepository(db.GetDatabase())
	marketStateRepo := mongodb.NewMarketStateRepository(db.GetDatabase())
	orderBookRepo := memory.NewOrderBookRepository()

	// Create broadcaster
	broadcaster := transport.NewBroadcaster()

	// Create market engine
	marketEngine := market.NewMarketEngine(cfg, db, orderRepo, fillRepo, marketStateRepo, orderBookRepo, broadcaster)

	// Create rate limiter
	rateLimitConfig := service.RateLimitConfig{
		OrdersPerMin: cfg.Security.RateLimitOrdersPerMin,
		WindowSize:   time.Minute,
	}
	rateLimiter := service.NewRateLimiter(rateLimitConfig)

	// Create services
	authService := service.NewAuthService(teamRepo)
	orderService := service.NewOrderService(orderRepo, marketEngine)
	resyncService := service.NewResyncService(fillRepo)
	productionService := service.NewProductionService(teamRepo)

	// Create message router
	router := transport.NewMessageRouter(authService, orderService, broadcaster, marketEngine, resyncService, productionService, rateLimiter)

	// Create ticker service
	tickerService := market.NewTickerService(cfg, marketStateRepo, orderBookRepo, broadcaster)

	// Start market engine
	if err := marketEngine.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to start market engine")
	}
	defer func() {
		if err := marketEngine.Stop(); err != nil {
			log.Error().Err(err).Msg("Error stopping market engine")
		}
	}()

	// Start ticker service
	if err := tickerService.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start ticker service")
	}
	defer func() {
		if err := tickerService.Stop(); err != nil {
			log.Error().Err(err).Msg("Error stopping ticker service")
		}
	}()

	// Create and start WebSocket server
	server := transport.NewWebSocketServer(cfg, router)
	if err := server.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start WebSocket server")
	}

	// Wait for shutdown signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Info().Msg("Shutdown signal received")

	// Graceful shutdown
	if err := server.Stop(); err != nil {
		log.Error().Err(err).Msg("Error during server shutdown")
	}

	log.Info().Msg("Server shutdown complete")
}
