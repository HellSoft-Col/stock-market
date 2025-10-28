package main

import (
	"context"
	"flag"
	"math/rand"

	"github.com/rs/zerolog/log"
	"github.com/yourusername/avocado-exchange-server/internal/config"
	"github.com/yourusername/avocado-exchange-server/internal/domain"
	"github.com/yourusername/avocado-exchange-server/internal/repository/mongodb"
	"github.com/yourusername/avocado-exchange-server/pkg/logger"
)

func main() {
	var configFile = flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

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

	// Create team repository
	teamRepo := mongodb.NewTeamRepository(db.GetDatabase())

	// Sample teams
	teams := []*domain.Team{
		{
			APIKey:             "TK-ANDROMEDA-2025-AVOCULTORES",
			TeamName:           "EquipoAndromeda",
			Species:            "Avocultores del Hueso Cósmico",
			InitialBalance:     10000.0,
			AuthorizedProducts: []string{"PALTA-OIL", "GUACA", "SEBO"},
			Recipes: map[string]domain.Recipe{
				"PALTA-OIL": {
					Type:         "BASIC",
					Ingredients:  map[string]int{},
					PremiumBonus: 1.0,
				},
				"GUACA": {
					Type:         "PREMIUM",
					Ingredients:  map[string]int{"FOSFO": 5, "PITA": 3},
					PremiumBonus: 1.30,
				},
				"SEBO": {
					Type:         "PREMIUM",
					Ingredients:  map[string]int{"NUCREM": 8},
					PremiumBonus: 1.30,
				},
			},
			Role: domain.TeamRole{
				Branches:    2,
				MaxDepth:    4,
				Decay:       0.7651 + (rand.Float64()-0.5)*0.1, // ±5% variation
				Budget:      24.83 + (rand.Float64()-0.5)*2.48, // ±5% variation
				BaseEnergy:  3.0,
				LevelEnergy: 2.0,
			},
		},
		{
			APIKey:             "TK-ORION-2025-MONJES",
			TeamName:           "EquipoOrion",
			Species:            "Monjes del Aguacate Sagrado",
			InitialBalance:     10000.0,
			AuthorizedProducts: []string{"FOSFO", "NUCREM", "GTRON"},
			Recipes: map[string]domain.Recipe{
				"FOSFO": {
					Type:         "BASIC",
					Ingredients:  map[string]int{},
					PremiumBonus: 1.0,
				},
				"NUCREM": {
					Type:         "BASIC",
					Ingredients:  map[string]int{},
					PremiumBonus: 1.0,
				},
				"GTRON": {
					Type:         "PREMIUM",
					Ingredients:  map[string]int{"FOSFO": 12, "NUCREM": 6},
					PremiumBonus: 1.45,
				},
			},
			Role: domain.TeamRole{
				Branches:    3,
				MaxDepth:    3,
				Decay:       0.8200 + (rand.Float64()-0.5)*0.082, // ±5% variation
				Budget:      22.50 + (rand.Float64()-0.5)*2.25,   // ±5% variation
				BaseEnergy:  3.0,
				LevelEnergy: 2.0,
			},
		},
		{
			APIKey:             "TK-VEGA-2025-ALQUIMISTAS",
			TeamName:           "EquipoVega",
			Species:            "Alquimistas de la Pulpa",
			InitialBalance:     10000.0,
			AuthorizedProducts: []string{"PITA", "CASCAR-ALLOY", "H-GUACA"},
			Recipes: map[string]domain.Recipe{
				"PITA": {
					Type:         "BASIC",
					Ingredients:  map[string]int{},
					PremiumBonus: 1.0,
				},
				"CASCAR-ALLOY": {
					Type:         "PREMIUM",
					Ingredients:  map[string]int{"PITA": 20},
					PremiumBonus: 1.25,
				},
				"H-GUACA": {
					Type:         "ULTRA",
					Ingredients:  map[string]int{"GUACA": 3, "GTRON": 2, "CASCAR-ALLOY": 1},
					PremiumBonus: 2.00,
				},
			},
			Role: domain.TeamRole{
				Branches:    1,
				MaxDepth:    6,
				Decay:       0.6800 + (rand.Float64()-0.5)*0.068, // ±5% variation
				Budget:      28.90 + (rand.Float64()-0.5)*2.89,   // ±5% variation
				BaseEnergy:  3.0,
				LevelEnergy: 2.0,
			},
		},
	}

	log.Info().Int("count", len(teams)).Msg("Seeding teams")

	for _, team := range teams {
		if err := teamRepo.Create(ctx, team); err != nil {
			log.Error().
				Str("teamName", team.TeamName).
				Err(err).
				Msg("Failed to create team")
		} else {
			log.Info().
				Str("teamName", team.TeamName).
				Str("apiKey", team.APIKey).
				Str("species", team.Species).
				Msg("Team created successfully")
		}
	}

	log.Info().Msg("Team seeding completed")
}
