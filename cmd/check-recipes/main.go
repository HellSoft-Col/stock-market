package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/config"
	"github.com/HellSoft-Col/stock-market/internal/repository/mongodb"
)

func main() {
	// Load configuration - try local config first, fallback to main config
	configFile := "config.local.yaml"
	cfg, err := config.Load(configFile)
	if err != nil {
		configFile = "config.yaml"
		cfg, err = config.Load(configFile)
	}
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to MongoDB
	db := mongodb.NewDatabase(&cfg.MongoDB)
	ctx := context.Background()

	if err := db.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := db.Close(ctx); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// Test database connection
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create team repository
	teamRepo := mongodb.NewTeamRepository(db.GetDatabase())

	// Get Avocultores team
	team, err := teamRepo.GetByTeamName(ctx, "Avocultores")
	if err != nil {
		log.Fatalf("Failed to get Avocultores team: %v", err)
	}

	fmt.Println("========================================")
	fmt.Println("Team: Avocultores")
	fmt.Println("========================================")
	fmt.Printf("Species: %s\n", team.Species)
	fmt.Printf("Token: %s\n", team.APIKey)
	fmt.Printf("Authorized Products: %v\n", team.AuthorizedProducts)
	fmt.Println()

	fmt.Println("Recipes:")
	fmt.Println("--------")

	if team.Recipes == nil || len(team.Recipes) == 0 {
		fmt.Println("❌ NO RECIPES FOUND!")
		fmt.Println("This is the problem - recipes are not stored in MongoDB")
		os.Exit(1)
	}

	// Pretty print recipes as JSON
	recipesJSON, err := json.MarshalIndent(team.Recipes, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal recipes: %v", err)
	}
	fmt.Println(string(recipesJSON))
	fmt.Println()

	// Check each expected recipe
	fmt.Println("Recipe Validation:")
	fmt.Println("------------------")

	// Check PALTA-OIL (basic product)
	if recipe, ok := team.Recipes["PALTA-OIL"]; ok {
		fmt.Printf("✓ PALTA-OIL: Type=%s, Ingredients=%v, Bonus=%.1f\n",
			recipe.Type, recipe.Ingredients, recipe.PremiumBonus)
		if recipe.Type != "BASIC" {
			fmt.Println("  ⚠️  WARNING: Should be BASIC type!")
		}
		if len(recipe.Ingredients) != 0 {
			fmt.Println("  ⚠️  WARNING: Should have no ingredients!")
		}
	} else {
		fmt.Println("❌ PALTA-OIL recipe missing!")
	}

	// Check GUACA (premium product)
	if recipe, ok := team.Recipes["GUACA"]; ok {
		fmt.Printf("✓ GUACA: Type=%s, Ingredients=%v, Bonus=%.1f\n",
			recipe.Type, recipe.Ingredients, recipe.PremiumBonus)
		if recipe.Type != "PREMIUM" {
			fmt.Println("  ❌ ERROR: Should be PREMIUM type!")
		}
		fosfo, hasFosfo := recipe.Ingredients["FOSFO"]
		pita, hasPita := recipe.Ingredients["PITA"]
		if !hasFosfo || fosfo != 5 {
			fmt.Printf("  ❌ ERROR: Should require 5 FOSFO (got %d)!\n", fosfo)
		}
		if !hasPita || pita != 3 {
			fmt.Printf("  ❌ ERROR: Should require 3 PITA (got %d)!\n", pita)
		}
		if hasFosfo && hasPita && fosfo == 5 && pita == 3 {
			fmt.Println("  ✓ Ingredients are correct!")
		}
	} else {
		fmt.Println("❌ GUACA recipe missing!")
	}

	// Check SEBO (premium product)
	if recipe, ok := team.Recipes["SEBO"]; ok {
		fmt.Printf("✓ SEBO: Type=%s, Ingredients=%v, Bonus=%.1f\n",
			recipe.Type, recipe.Ingredients, recipe.PremiumBonus)
		if recipe.Type != "PREMIUM" {
			fmt.Println("  ❌ ERROR: Should be PREMIUM type!")
		}
		nucrem, hasNucrem := recipe.Ingredients["NUCREM"]
		if !hasNucrem || nucrem != 8 {
			fmt.Printf("  ❌ ERROR: Should require 8 NUCREM (got %d)!\n", nucrem)
		} else {
			fmt.Println("  ✓ Ingredients are correct!")
		}
	} else {
		fmt.Println("❌ SEBO recipe missing!")
	}

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("Inventory:")
	fmt.Println("========================================")
	if team.Inventory == nil || len(team.Inventory) == 0 {
		fmt.Println("(empty)")
	} else {
		inventoryJSON, _ := json.MarshalIndent(team.Inventory, "", "  ")
		fmt.Println(string(inventoryJSON))
	}

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("CONCLUSION:")
	fmt.Println("========================================")

	if len(team.Recipes) == 3 {
		_, hasBasic := team.Recipes["PALTA-OIL"]
		guaca, hasGuaca := team.Recipes["GUACA"]
		sebo, hasSebo := team.Recipes["SEBO"]

		if hasBasic && hasGuaca && hasSebo &&
			guaca.Type == "PREMIUM" && sebo.Type == "PREMIUM" &&
			guaca.Ingredients["FOSFO"] == 5 && guaca.Ingredients["PITA"] == 3 &&
			sebo.Ingredients["NUCREM"] == 8 {
			fmt.Println("✓ All recipes are correctly configured in MongoDB!")
			fmt.Println()
			fmt.Println("The issue must be in:")
			fmt.Println("1. Recipe loading from MongoDB during production")
			fmt.Println("2. Recipe lookup logic in production_service.go")
			fmt.Println("3. Type comparison (PREMIUM vs premium case sensitivity?)")
			fmt.Println()
			fmt.Println("Next steps:")
			fmt.Println("- Start the server and check logs when producing GUACA")
			fmt.Println("- Look for 'Team recipes loaded' and 'Recipe lookup result' logs")
			os.Exit(0)
		}
	}

	fmt.Println("❌ Recipes are NOT properly configured!")
	fmt.Println()
	fmt.Println("Solution: Run 'Update All Recipes' from admin UI or use:")
	fmt.Println("  go run cmd/update-recipes/main.go")
	os.Exit(1)
}
