package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/HellSoft-Col/stock-market/internal/config"
	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/HellSoft-Col/stock-market/internal/repository/mongodb"
	"github.com/HellSoft-Col/stock-market/pkg/logger"
	"github.com/rs/zerolog/log"
)

// Team name to species mapping (from automated-clients.yaml + current teams)
var teamSpeciesMapping = map[string]string{
	// From automated-clients.yaml - current teams
	"Alquimistas de Palta":          "Avocultores",
	"Arpistas de Pita-Pita":         "Cosechadores de Pita",
	"Avocultores del Hueso Cósmico": "Avocultores",
	"Cartógrafos de Fosfolima":      "Cartógrafos",
	"Cosechadores de Semillas":      "Cosechadores de Pita",
	"Forjadores Holográficos":       "Herreros Cósmicos",
	"Ingenieros Holo-Aguacate":      "Núcleo Cremero",
	"Mensajeros del Núcleo":         "Núcleo Cremero",
	"Mineros de Guacatrones":        "Destiladores", // GUACA specialty (Guacatrones = Guaca)
	"Monjes del Guacamole Estelar":  "Destiladores", // GUACA specialty
	"Orfebres de Cáscara":           "Herreros Cósmicos",
	"Someliers de Aceite":           "Someliers Andorianos",

	// Additional standard team names from admin.html seedTeams
	"Avocultores de Paltalima": "Avocultores",
	"Monjes de Fosforescencia": "Monjes de Fosforescencia",
	"Cosechadores de Pita":     "Cosechadores de Pita",
	"Herreros Cósmicos":        "Herreros Cósmicos",
	"Extractores Cuánticos":    "Extractores",
	"Tejemanteles Estelares":   "Tejemanteles",
	"Cremeros Astrales":        "Cremeros Astrales",
	"Mineros del Sebo":         "Mineros del Sebo",
	"Núcleo Cremero":           "Núcleo Cremero",
	"Destiladores de Guaca":    "Destiladores",
	"Someliers Andorianos":     "Someliers Andorianos",
}

// Species to basic product mapping (from the 12 species table)
var speciesToBasicProduct = map[string]string{
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
	// NOTE: Only using EXISTING products: GUACA, SEBO, PALTA-OIL, FOSFO, NUCREM, CASCAR-ALLOY, PITA
	// Products NOT available: QUANTUM-PULP, SKIN-WRAP, ASTRO-BUTTER
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
		// ORIGINAL: QUANTUM-PULP (7 PALTA-OIL), SKIN-WRAP (12 ASTRO-BUTTER)
		// MODIFIED: Using only existing products
		recipes["GUACA"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"PALTA-OIL": 7},
			PremiumBonus: 1.3,
		}
		recipes["NUCREM"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"SEBO": 6},
			PremiumBonus: 1.3,
		}

	case "Extractores":
		// ORIGINAL: NUCREM (6 SEBO), FOSFO (9 SKIN-WRAP)
		// MODIFIED: Using only existing products
		recipes["NUCREM"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"SEBO": 6},
			PremiumBonus: 1.3,
		}
		recipes["FOSFO"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"PALTA-OIL": 7},
			PremiumBonus: 1.3,
		}

	case "Tejemanteles":
		// ORIGINAL: PITA (8 CASCAR-ALLOY), ASTRO-BUTTER (10 GUACA)
		// MODIFIED: Using only existing products
		recipes["PITA"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"CASCAR-ALLOY": 8},
			PremiumBonus: 1.3,
		}
		recipes["GUACA"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"FOSFO": 5, "PITA": 3},
			PremiumBonus: 1.3,
		}

	case "Cremeros Astrales":
		// ORIGINAL: CASCAR-ALLOY (10 FOSFO), PALTA-OIL (7 QUANTUM-PULP)
		// MODIFIED: Using only existing products
		recipes["CASCAR-ALLOY"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"FOSFO": 10},
			PremiumBonus: 1.3,
		}
		recipes["PALTA-OIL"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"SEBO": 6},
			PremiumBonus: 1.3,
		}

	case "Mineros del Sebo":
		// ORIGINAL: ASTRO-BUTTER (10 GUACA), GUACA (5 PALTA-OIL + 3 PITA)
		// MODIFIED: Using only existing products
		recipes["NUCREM"] = domain.Recipe{
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
		// ORIGINAL: SKIN-WRAP (12 ASTRO-BUTTER), QUANTUM-PULP (7 PALTA-OIL)
		// MODIFIED: Using only existing products
		recipes["PITA"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"GUACA": 10},
			PremiumBonus: 1.3,
		}
		recipes["FOSFO"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"PALTA-OIL": 7},
			PremiumBonus: 1.3,
		}

	case "Destiladores":
		// ORIGINAL: PALTA-OIL (7 QUANTUM-PULP), FOSFO (9 SKIN-WRAP)
		// MODIFIED: Using only existing products
		recipes["PALTA-OIL"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"SEBO": 7},
			PremiumBonus: 1.3,
		}
		recipes["FOSFO"] = domain.Recipe{
			Type:         "PREMIUM",
			Ingredients:  map[string]int{"NUCREM": 6},
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
	dryRun := flag.Bool("dry-run", false, "Show what would be updated without making changes")
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

	log.Info().
		Int("teamCount", len(teams)).
		Msg("🔧 Starting team species and recipe update")

	updatedCount := 0
	skippedCount := 0
	errorCount := 0

	for _, team := range teams {
		// Skip admin team
		if team.TeamName == "admin" {
			skippedCount++
			continue
		}

		// Get correct species for this team
		correctSpecies, found := teamSpeciesMapping[team.TeamName]
		if !found {
			log.Warn().
				Str("team", team.TeamName).
				Str("currentSpecies", team.Species).
				Msg("⚠️  Team not found in species mapping - skipping")
			errorCount++
			continue
		}

		// Get basic product for this species
		basicProduct := speciesToBasicProduct[correctSpecies]
		if basicProduct == "" {
			log.Error().
				Str("team", team.TeamName).
				Str("species", correctSpecies).
				Msg("❌ Unknown species - cannot determine basic product")
			errorCount++
			continue
		}

		// Build recipes for this species
		recipes := buildRecipesForSpecies(correctSpecies, basicProduct)

		if *dryRun {
			log.Info().
				Str("team", team.TeamName).
				Str("currentSpecies", team.Species).
				Str("newSpecies", correctSpecies).
				Str("basicProduct", basicProduct).
				Int("recipeCount", len(recipes)).
				Msg("🔍 [DRY RUN] Would update team")
			updatedCount++
			continue
		}

		// Update species
		if team.Species != correctSpecies {
			if err := teamRepo.UpdateSpecies(ctx, team.TeamName, correctSpecies); err != nil {
				log.Error().
					Err(err).
					Str("team", team.TeamName).
					Msg("Failed to update species")
				errorCount++
				continue
			}
		}

		// Update recipes
		if err := teamRepo.UpdateRecipes(ctx, team.TeamName, recipes); err != nil {
			log.Error().
				Err(err).
				Str("team", team.TeamName).
				Msg("Failed to update recipes")
			errorCount++
			continue
		}

		log.Info().
			Str("team", team.TeamName).
			Str("species", correctSpecies).
			Str("basicProduct", basicProduct).
			Int("recipeCount", len(recipes)).
			Msg("✅ Team updated successfully")
		updatedCount++
	}

	// Summary
	fmt.Println()
	log.Info().Msg("=" + string(make([]byte, 60)))
	log.Info().
		Int("updated", updatedCount).
		Int("skipped", skippedCount).
		Int("errors", errorCount).
		Int("total", len(teams)).
		Msg("📝 Update Summary")
	log.Info().Msg("=" + string(make([]byte, 60)))

	if *dryRun {
		log.Info().Msg("🔍 This was a DRY RUN - no changes were made")
		log.Info().Msg("💡 Run without --dry-run to apply changes")
	} else if updatedCount > 0 {
		log.Info().Msg("🎉 Teams updated successfully!")
		log.Info().Msg("✅ All teams now have correct species and recipes")
	}

	if errorCount > 0 {
		log.Warn().
			Int("errors", errorCount).
			Msg("⚠️  Some teams had errors - check logs above")
	}
}
