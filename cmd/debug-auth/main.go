package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/yourusername/avocado-exchange-server/internal/config"
	"github.com/yourusername/avocado-exchange-server/internal/repository/mongodb"
	"github.com/yourusername/avocado-exchange-server/internal/service"
	"github.com/yourusername/avocado-exchange-server/pkg/logger"
)

func main() {
	var configFile = flag.String("config", "config.yaml", "Path to configuration file")
	var token = flag.String("token", "", "Token to test")
	flag.Parse()

	if *token == "" {
		log.Fatal().Msg("Token is required. Use -token flag")
	}

	// Initialize logger
	logger.InitLogger("info", "console")

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Connect to database
	db := mongodb.NewDatabase(&cfg.MongoDB)
	ctx := context.Background()

	if err := db.Connect(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close(ctx)

	log.Info().Msg("Connected to database successfully")

	// Create repositories and services
	teamRepo := mongodb.NewTeamRepository(db.GetDatabase())
	authService := service.NewAuthService(teamRepo)

	fmt.Printf("\n🔍 Testing authentication for token: %s\n", *token)
	fmt.Println(strings.Repeat("=", 60))

	// Test direct repository lookup
	fmt.Printf("\n1. Testing direct repository lookup...\n")
	team, err := teamRepo.GetByAPIKey(ctx, *token)
	if err != nil {
		fmt.Printf("❌ Repository lookup failed: %v\n", err)
	} else {
		fmt.Printf("✅ Repository lookup successful!\n")
		fmt.Printf("   Team Name: %s\n", team.TeamName)
		fmt.Printf("   Species: %s\n", team.Species)
		fmt.Printf("   API Key: %s\n", team.APIKey)
	}

	// Test auth service
	fmt.Printf("\n2. Testing auth service...\n")
	team, err = authService.ValidateToken(ctx, *token)
	if err != nil {
		fmt.Printf("❌ Auth service failed: %v\n", err)
	} else {
		fmt.Printf("✅ Auth service successful!\n")
		fmt.Printf("   Team Name: %s\n", team.TeamName)
		fmt.Printf("   Species: %s\n", team.Species)
	}

	// List all teams for debugging
	fmt.Printf("\n3. Listing all teams in database...\n")
	teams, err := teamRepo.GetAll(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to get teams: %v\n", err)
	} else {
		fmt.Printf("Found %d teams:\n", len(teams))
		for i, t := range teams {
			fmt.Printf("   %d. %s (Token: %s, Species: %s)\n",
				i+1, t.TeamName, t.APIKey, t.Species)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🔍 Debug complete!")
}
