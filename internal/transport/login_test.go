package transport

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/domain"
)

// TestLoginOKMessageJSONSerialization verifies that TeamRole energy fields
// are correctly serialized to JSON with camelCase field names
func TestLoginOKMessageJSONSerialization(t *testing.T) {
	// Create a team with energy values (simulating database data)
	team := &domain.Team{
		TeamName:           "TestTeam",
		Species:            "Premium",
		InitialBalance:     100000.0,
		CurrentBalance:     100000.0,
		Inventory:          map[string]int{"FOSFO": 10},
		AuthorizedProducts: []string{"FOSFO", "PITA"},
		Recipes: map[string]domain.Recipe{
			"FOSFO": {
				Type:         "BASIC",
				Ingredients:  map[string]int{},
				PremiumBonus: 1.0,
			},
		},
		Role: domain.TeamRole{
			Branches:    2,
			MaxDepth:    4,
			Decay:       0.75,
			Budget:      25.0,
			BaseEnergy:  3.0,
			LevelEnergy: 2.0,
		},
	}

	// Create LOGIN_OK message exactly as the server does
	loginOKMsg := &domain.LoginOKMessage{
		Type:               "LOGIN_OK",
		Team:               team.TeamName,
		Species:            team.Species,
		InitialBalance:     team.InitialBalance,
		CurrentBalance:     team.CurrentBalance,
		Inventory:          team.Inventory,
		AuthorizedProducts: team.AuthorizedProducts,
		Recipes:            team.Recipes,
		Role:               team.Role,
		ServerTime:         time.Now().Format(time.RFC3339),
	}

	// Marshal to JSON (this is what gets sent over the wire)
	jsonBytes, err := json.Marshal(loginOKMsg)
	if err != nil {
		t.Fatalf("Failed to marshal LOGIN_OK message: %v", err)
	}

	jsonString := string(jsonBytes)
	t.Logf("Generated JSON: %s", jsonString)

	// Verify the JSON contains the energy fields with camelCase names
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &jsonMap); err != nil {
		t.Fatalf("Failed to unmarshal JSON back to map: %v", err)
	}

	// Check that role exists
	role, ok := jsonMap["role"].(map[string]interface{})
	if !ok {
		t.Fatal("Role field missing or not a map in JSON")
	}

	// Verify all role fields exist with correct names
	testCases := []struct {
		fieldName     string
		expectedValue float64
	}{
		{"branches", 2.0},
		{"maxDepth", 4.0},
		{"decay", 0.75},
		{"budget", 25.0},
		{"baseEnergy", 3.0},  // Critical field
		{"levelEnergy", 2.0}, // Critical field
	}

	for _, tc := range testCases {
		value, exists := role[tc.fieldName]
		if !exists {
			t.Errorf("Field '%s' missing in role JSON", tc.fieldName)
			continue
		}

		floatValue, ok := value.(float64)
		if !ok {
			t.Errorf("Field '%s' is not a float64, got %T", tc.fieldName, value)
			continue
		}

		if floatValue != tc.expectedValue {
			t.Errorf("Field '%s' has wrong value: got %.2f, want %.2f",
				tc.fieldName, floatValue, tc.expectedValue)
		} else {
			t.Logf("✓ Field '%s' correctly serialized as %.2f", tc.fieldName, floatValue)
		}
	}

	// CRITICAL: Verify that baseEnergy and levelEnergy are NOT null
	if role["baseEnergy"] == nil {
		t.Error("baseEnergy is null in JSON - THIS IS THE BUG!")
	}
	if role["levelEnergy"] == nil {
		t.Error("levelEnergy is null in JSON - THIS IS THE BUG!")
	}

	// Verify camelCase naming (not snake_case)
	if _, exists := role["base_energy"]; exists {
		t.Error("Found 'base_energy' (snake_case) - should be 'baseEnergy' (camelCase)")
	}
	if _, exists := role["level_energy"]; exists {
		t.Error("Found 'level_energy' (snake_case) - should be 'levelEnergy' (camelCase)")
	}
}

// TestLoginOKMessageWithDefaultEnergy tests the default energy value logic
func TestLoginOKMessageWithDefaultEnergy(t *testing.T) {
	// Create a team with ZERO energy values (simulating old database records)
	team := &domain.Team{
		TeamName:           "TestTeam",
		Species:            "Premium",
		InitialBalance:     100000.0,
		CurrentBalance:     100000.0,
		Inventory:          map[string]int{},
		AuthorizedProducts: []string{"FOSFO"},
		Recipes:            map[string]domain.Recipe{},
		Role: domain.TeamRole{
			Branches:    2,
			MaxDepth:    4,
			Decay:       0.75,
			Budget:      25.0,
			BaseEnergy:  0.0, // Zero value
			LevelEnergy: 0.0, // Zero value
		},
	}

	// Apply the same default logic as in handleLogin
	role := team.Role
	if role.BaseEnergy == 0 {
		role.BaseEnergy = 3.0
	}
	if role.LevelEnergy == 0 {
		role.LevelEnergy = 2.0
	}

	// Create LOGIN_OK message
	loginOKMsg := &domain.LoginOKMessage{
		Type:               "LOGIN_OK",
		Team:               team.TeamName,
		Species:            team.Species,
		InitialBalance:     team.InitialBalance,
		CurrentBalance:     team.CurrentBalance,
		Inventory:          team.Inventory,
		AuthorizedProducts: team.AuthorizedProducts,
		Recipes:            team.Recipes,
		Role:               role,
		ServerTime:         time.Now().Format(time.RFC3339),
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(loginOKMsg)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Verify energy fields have default values
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &jsonMap); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	roleMap, ok := jsonMap["role"].(map[string]interface{})
	if !ok {
		t.Fatal("Role field missing")
	}

	baseEnergy := roleMap["baseEnergy"].(float64)
	levelEnergy := roleMap["levelEnergy"].(float64)

	if baseEnergy != 3.0 {
		t.Errorf("Default baseEnergy not applied: got %.2f, want 3.0", baseEnergy)
	} else {
		t.Log("✓ Default baseEnergy correctly applied: 3.0")
	}

	if levelEnergy != 2.0 {
		t.Errorf("Default levelEnergy not applied: got %.2f, want 2.0", levelEnergy)
	} else {
		t.Log("✓ Default levelEnergy correctly applied: 2.0")
	}
}

// TestTeamRoleStructTags verifies the JSON tags are correct on the struct
func TestTeamRoleStructTags(t *testing.T) {
	role := domain.TeamRole{
		Branches:    2,
		MaxDepth:    4,
		Decay:       0.75,
		Budget:      25.0,
		BaseEnergy:  3.0,
		LevelEnergy: 2.0,
	}

	jsonBytes, err := json.Marshal(role)
	if err != nil {
		t.Fatalf("Failed to marshal TeamRole: %v", err)
	}

	jsonString := string(jsonBytes)
	t.Logf("TeamRole JSON: %s", jsonString)

	// Parse back to verify field names
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify camelCase field names exist
	requiredFields := []string{"branches", "maxDepth", "decay", "budget", "baseEnergy", "levelEnergy"}
	for _, field := range requiredFields {
		if _, exists := result[field]; !exists {
			t.Errorf("Required field '%s' missing in JSON", field)
		} else {
			t.Logf("✓ Field '%s' present in JSON", field)
		}
	}

	// Ensure snake_case fields don't exist
	forbiddenFields := []string{"base_energy", "level_energy", "max_depth"}
	for _, field := range forbiddenFields {
		if _, exists := result[field]; exists {
			t.Errorf("Forbidden snake_case field '%s' found in JSON", field)
		}
	}
}
