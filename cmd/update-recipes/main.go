package main

import (
	"context"
	"flag"

	"github.com/HellSoft-Col/stock-market/internal/config"
	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/HellSoft-Col/stock-market/internal/repository/mongodb"
	"github.com/HellSoft-Col/stock-market/pkg/logger"
	"github.com/rs/zerolog/log"
)

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

func getBasicProductForSpecies(species string) string {
	basicProducts := map[string]string{
		"Avocultores":              "PALTA-OIL",
		"Monjes de Fosforescencia": "FOSFO",
		"Cosechadores de Pita":     "PITA",
		"Herreros Cósmicos":        "CASCAR-ALLOY",
		"Extractores":              "QUANTUM-PULP",
		"Tejemanteles":             "SKIN-WRAP",
		"Cremeros Astrales":        "ASTRO-BUTTER",
		"Mineros del Sebo":         "SEBO",
		"Núcleo Cremero":           "NUCREM",
		"Destiladores":             "GUACA",
		"Cartógrafos":              "GUACA",
		"Someliers Andorianos":     "PALTA-OIL",
	}

	return basicProducts[species]
}

func main() {
	configFile := flag.String("config", "config.yaml", "Path to configuration file")
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
	defer func() {
		if err := db.Close(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to close database connection")
		}
	}()

	log.Info().Msg("Connected to database successfully")

	// Create team repository
	teamRepo := mongodb.NewTeamRepository(db.GetDatabase())

	// Get all teams
	teams, err := teamRepo.GetAll(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get teams")
	}

	log.Info().Int("count", len(teams)).Msg("Found teams")

	updatedCount := 0
	errorCount := 0

	for _, team := range teams {
		log.Info().
			Str("teamName", team.TeamName).
			Str("species", team.Species).
			Msg("Processing team")

		basicProduct := getBasicProductForSpecies(team.Species)
		if basicProduct == "" {
			log.Warn().
				Str("teamName", team.TeamName).
				Str("species", team.Species).
				Msg("Unknown species - skipping")
			errorCount++
			continue
		}

		// Build correct recipes
		correctRecipes := buildRecipesForSpecies(team.Species, basicProduct)

		// Update team recipes
		if err := teamRepo.UpdateRecipes(ctx, team.TeamName, correctRecipes); err != nil {
			log.Error().
				Err(err).
				Str("teamName", team.TeamName).
				Msg("Failed to update team recipes")
			errorCount++
			continue
		}

		log.Info().
			Str("teamName", team.TeamName).
			Interface("recipes", correctRecipes).
			Msg("Updated team recipes")
		updatedCount++
	}

	log.Info().
		Int("total", len(teams)).
		Int("updated", updatedCount).
		Int("errors", errorCount).
		Msg("Recipe update complete")

	if errorCount > 0 {
		log.Fatal().Int("errors", errorCount).Msg("Some teams failed to update")
	}

	log.Info().Msg("🎉 All teams updated successfully!")
}
