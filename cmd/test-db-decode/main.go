package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/HellSoft-Col/stock-market/internal/config"
	"github.com/HellSoft-Col/stock-market/internal/repository/mongodb"
	"github.com/HellSoft-Col/stock-market/pkg/logger"
	"github.com/rs/zerolog/log"
)

func main() {
	var configFile = flag.String("config", "config.yaml", "Path to configuration file")
	var apiKey = flag.String("apikey", "", "API key to test")
	flag.Parse()

	if *apiKey == "" {
		fmt.Println("Usage: go run main.go --apikey <API_KEY>")
		fmt.Println("Example: go run main.go --apikey TK-09jKZrvn0NF11v99j10vT4Fx")
		return
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
	defer func() {
		if err := db.Close(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to close database connection")
		}
	}()

	fmt.Printf("✅ Connected to MongoDB\n\n")

	// Create team repository
	teamRepo := mongodb.NewTeamRepository(db.GetDatabase())

	// Get team by API key
	fmt.Printf("🔍 Fetching team with API key: %s\n\n", *apiKey)
	team, err := teamRepo.GetByAPIKey(ctx, *apiKey)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("✅ Team found: %s\n\n", team.TeamName)

	// Print raw team structure
	fmt.Println("=== Raw Team Structure ===")
	teamJSON, _ := json.MarshalIndent(team, "", "  ")
	fmt.Printf("%s\n\n", teamJSON)

	// Check role fields specifically
	fmt.Println("=== Role Fields ===")
	fmt.Printf("Branches: %d\n", team.Role.Branches)
	fmt.Printf("MaxDepth: %d\n", team.Role.MaxDepth)
	fmt.Printf("Decay: %.4f\n", team.Role.Decay)
	fmt.Printf("Budget: %.2f\n", team.Role.Budget)
	fmt.Printf("BaseEnergy: %.2f\n", team.Role.BaseEnergy)
	fmt.Printf("LevelEnergy: %.2f\n\n", team.Role.LevelEnergy)

	// Check if energy values are zero
	if team.Role.BaseEnergy == 0 {
		fmt.Println("⚠️  WARNING: BaseEnergy is 0!")
	} else {
		fmt.Println("✅ BaseEnergy has a value")
	}

	if team.Role.LevelEnergy == 0 {
		fmt.Println("⚠️  WARNING: LevelEnergy is 0!")
	} else {
		fmt.Println("✅ LevelEnergy has a value")
	}
}
