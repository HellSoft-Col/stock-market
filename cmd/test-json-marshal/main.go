package main

import (
	"encoding/json"
	"fmt"

	"github.com/HellSoft-Col/stock-market/internal/domain"
)

func main() {
	// Create a test team role with energy values
	role := domain.TeamRole{
		Branches:    2,
		MaxDepth:    4,
		Decay:       0.75,
		Budget:      25.0,
		BaseEnergy:  3.0,
		LevelEnergy: 2.0,
	}

	// Test 1: Marshal the role directly
	fmt.Println("=== Test 1: Marshal TeamRole directly ===")
	roleJSON, err := json.MarshalIndent(role, "", "  ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("%s\n\n", roleJSON)
	}

	// Test 2: Marshal in a LoginOKMessage
	fmt.Println("=== Test 2: Marshal TeamRole in LoginOKMessage ===")
	loginMsg := domain.LoginOKMessage{
		Type:               "LOGIN_OK",
		Team:               "TestTeam",
		Species:            "TestSpecies",
		InitialBalance:     10000.0,
		CurrentBalance:     10000.0,
		Inventory:          map[string]int{"FOSFO": 10},
		AuthorizedProducts: []string{"FOSFO", "PITA"},
		Recipes: map[string]domain.Recipe{
			"FOSFO": {
				Type:         "BASIC",
				Ingredients:  map[string]int{},
				PremiumBonus: 1.0,
			},
		},
		Role:       role,
		ServerTime: "2024-01-01T00:00:00Z",
	}

	loginJSON, err := json.MarshalIndent(loginMsg, "", "  ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("%s\n\n", loginJSON)
	}

	// Test 3: Check if energy values are present
	fmt.Println("=== Test 3: Verify energy values in JSON ===")
	var result map[string]any
	if err := json.Unmarshal(loginJSON, &result); err != nil {
		fmt.Printf("Error unmarshaling: %v\n", err)
		return
	}

	if role, ok := result["role"].(map[string]any); ok {
		fmt.Printf("✅ Role field exists\n")
		fmt.Printf("  baseEnergy: %v (type: %T)\n", role["baseEnergy"], role["baseEnergy"])
		fmt.Printf("  levelEnergy: %v (type: %T)\n", role["levelEnergy"], role["levelEnergy"])

		if role["baseEnergy"] == nil {
			fmt.Printf("⚠️  WARNING: baseEnergy is nil!\n")
		}
		if role["levelEnergy"] == nil {
			fmt.Printf("⚠️  WARNING: levelEnergy is nil!\n")
		}
	} else {
		fmt.Printf("❌ Role field missing or invalid\n")
	}
}
