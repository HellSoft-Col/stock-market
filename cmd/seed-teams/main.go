package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"

	"github.com/HellSoft-Col/stock-market/internal/config"
	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/HellSoft-Col/stock-market/internal/repository/mongodb"
	"github.com/HellSoft-Col/stock-market/pkg/logger"
	"github.com/rs/zerolog/log"
)

type TeamData struct {
	TeamName           string   `json:"teamName"`
	Token              string   `json:"token"`
	Species            string   `json:"species"`
	Specialty          string   `json:"specialty"`
	Recipe             string   `json:"recipe"`
	InitialBalance     int      `json:"initialBalance"`
	AuthorizedProducts []string `json:"authorizedProducts"`
}

type TeamsFile struct {
	GeneratedAt string     `json:"generated_at"`
	TotalTeams  int        `json:"total_teams"`
	Teams       []TeamData `json:"teams"`
}

func parseRecipe(recipeStr string) map[string]int {
	// Simple recipe parser for formats like "5 FOSFO + 3 PITA" or "8 NUCREM"
	ingredients := make(map[string]int)

	// For now, return empty map - recipes can be added later as the system evolves
	// The current trading system focuses on direct trading rather than crafting
	return ingredients
}

func createDomainTeam(teamData TeamData) *domain.Team {
	// Convert the generated team data to domain.Team format
	return &domain.Team{
		APIKey:             teamData.Token,
		TeamName:           teamData.TeamName,
		Species:            teamData.Species,
		InitialBalance:     float64(teamData.InitialBalance),
		AuthorizedProducts: teamData.AuthorizedProducts,
		Recipes:            buildRecipesForSpecies(teamData.Species, teamData.Specialty),
		Role: domain.TeamRole{
			Branches:    2,
			MaxDepth:    4,
			Decay:       0.75,
			Budget:      25.0,
			BaseEnergy:  3.0,
			LevelEnergy: 2.0,
		},
	}
}

// buildRecipesForSpecies creates all recipes (basic + premium) for a species
func buildRecipesForSpecies(species string, basicProduct string) map[string]domain.Recipe {
	recipes := make(map[string]domain.Recipe)

	// Add basic recipe (free production)
	recipes[basicProduct] = domain.Recipe{
		Type:         "BASIC",
		Ingredients:  map[string]int{},
		PremiumBonus: 1.0,
	}

	// Add premium recipes based on species (30% bonus)
	switch species {
	case "Avocultores":
		recipes["GUACA"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"FOSFO": 5, "PITA": 3},
			PremiumBonus: 1.3,
		}
		recipes["SEBO"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"NUCREM": 8},
			PremiumBonus: 1.3,
		}

	case "Monjes de Fosforescencia":
		recipes["GUACA"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"PALTA-OIL": 5, "PITA": 3},
			PremiumBonus: 1.3,
		}
		recipes["NUCREM"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"SEBO": 6},
			PremiumBonus: 1.3,
		}

	case "Cosechadores de Pita":
		recipes["SEBO"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"NUCREM": 8},
			PremiumBonus: 1.3,
		}
		recipes["CASCAR-ALLOY"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"FOSFO": 10},
			PremiumBonus: 1.3,
		}

	case "Herreros Cósmicos":
		recipes["QUANTUM-PULP"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"PALTA-OIL": 7},
			PremiumBonus: 1.3,
		}
		recipes["SKIN-WRAP"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"ASTRO-BUTTER": 12},
			PremiumBonus: 1.3,
		}

	case "Extractores":
		recipes["NUCREM"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"SEBO": 6},
			PremiumBonus: 1.3,
		}
		recipes["FOSFO"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"SKIN-WRAP": 9},
			PremiumBonus: 1.3,
		}

	case "Tejemanteles":
		recipes["PITA"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"CASCAR-ALLOY": 8},
			PremiumBonus: 1.3,
		}
		recipes["ASTRO-BUTTER"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"GUACA": 10},
			PremiumBonus: 1.3,
		}

	case "Cremeros Astrales":
		recipes["CASCAR-ALLOY"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"FOSFO": 10},
			PremiumBonus: 1.3,
		}
		recipes["PALTA-OIL"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"QUANTUM-PULP": 7},
			PremiumBonus: 1.3,
		}

	case "Mineros del Sebo":
		recipes["ASTRO-BUTTER"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"GUACA": 10},
			PremiumBonus: 1.3,
		}
		recipes["GUACA"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"PALTA-OIL": 5, "PITA": 3},
			PremiumBonus: 1.3,
		}

	case "Núcleo Cremero":
		recipes["SKIN-WRAP"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"ASTRO-BUTTER": 12},
			PremiumBonus: 1.3,
		}
		recipes["QUANTUM-PULP"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"PALTA-OIL": 7},
			PremiumBonus: 1.3,
		}

	case "Destiladores":
		recipes["PALTA-OIL"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"QUANTUM-PULP": 7},
			PremiumBonus: 1.3,
		}
		recipes["FOSFO"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"SKIN-WRAP": 9},
			PremiumBonus: 1.3,
		}

	case "Cartógrafos":
		recipes["NUCREM"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"SEBO": 6},
			PremiumBonus: 1.3,
		}
		recipes["PITA"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"CASCAR-ALLOY": 8},
			PremiumBonus: 1.3,
		}

	case "Someliers Andorianos":
		recipes["SEBO"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"NUCREM": 8},
			PremiumBonus: 1.3,
		}
		recipes["CASCAR-ALLOY"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"FOSFO": 10},
			PremiumBonus: 1.3,
		}
	}

	return recipes
}

func main() {
	configFile := flag.String("config", "config.yaml", "Path to configuration file")
	teamsFile := flag.String("teams", "", "Path to teams JSON file")
	flag.Parse()

	if *teamsFile == "" {
		log.Fatal().Msg("Teams file is required. Use -teams flag to specify the JSON file")
	}

	// Initialize logger
	logger.InitLogger("info", "console")

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Read teams file
	log.Info().Str("file", *teamsFile).Msg("Reading teams from file")
	teamsData, err := os.ReadFile(*teamsFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read teams file")
	}

	// Parse teams JSON
	var teamsFile_parsed TeamsFile
	if err := json.Unmarshal(teamsData, &teamsFile_parsed); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse teams JSON")
	}

	log.Info().
		Int("totalTeams", teamsFile_parsed.TotalTeams).
		Str("generatedAt", teamsFile_parsed.GeneratedAt).
		Msg("Teams file loaded")

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

	log.Info().Msg("Connected to database successfully")

	// Create team repository
	teamRepo := mongodb.NewTeamRepository(db.GetDatabase())

	// Convert and seed teams
	log.Info().Int("count", len(teamsFile_parsed.Teams)).Msg("Seeding teams")

	successCount := 0
	errorCount := 0

	for _, teamData := range teamsFile_parsed.Teams {
		domainTeam := createDomainTeam(teamData)

		if err := teamRepo.Create(ctx, domainTeam); err != nil {
			log.Error().
				Str("teamName", teamData.TeamName).
				Str("token", teamData.Token).
				Err(err).
				Msg("Failed to create team")
			errorCount++
		} else {
			log.Info().
				Str("teamName", teamData.TeamName).
				Str("token", teamData.Token).
				Str("species", teamData.Species).
				Str("specialty", teamData.Specialty).
				Msg("Team created successfully")
			successCount++
		}
	}

	log.Info().
		Int("successful", successCount).
		Int("errors", errorCount).
		Int("total", len(teamsFile_parsed.Teams)).
		Msg("Team seeding completed")

	if errorCount > 0 {
		log.Warn().Int("errors", errorCount).Msg("Some teams failed to be created")
	} else {
		log.Info().Msg("🎉 All teams seeded successfully!")
	}
}
