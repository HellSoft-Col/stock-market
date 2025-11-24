package production

import (
	"testing"
)

func TestRecipeManager(t *testing.T) {
	recipes := map[string]*Recipe{
		"PALTA-OIL": {
			Product:      "PALTA-OIL",
			Ingredients:  nil,
			PremiumBonus: 1.0,
		},
		"GUACA": {
			Product: "GUACA",
			Ingredients: map[string]int{
				"FOSFO": 5,
				"PITA":  3,
			},
			PremiumBonus: 1.30,
		},
	}

	rm := NewRecipeManager(recipes)

	t.Run("GetRecipe", func(t *testing.T) {
		recipe, err := rm.GetRecipe("GUACA")
		if err != nil {
			t.Fatalf("Expected to find recipe, got error: %v", err)
		}

		if recipe.Product != "GUACA" {
			t.Errorf("Expected product GUACA, got %s", recipe.Product)
		}

		if recipe.PremiumBonus != 1.30 {
			t.Errorf("Expected bonus 1.30, got %f", recipe.PremiumBonus)
		}
	})

	t.Run("GetRecipe_NotFound", func(t *testing.T) {
		_, err := rm.GetRecipe("INVALID")
		if err == nil {
			t.Error("Expected error for non-existent recipe")
		}
	})

	t.Run("IsBasicProduction", func(t *testing.T) {
		if !rm.IsBasicProduction("PALTA-OIL") {
			t.Error("Expected PALTA-OIL to be basic production")
		}

		if rm.IsBasicProduction("GUACA") {
			t.Error("Expected GUACA to not be basic production")
		}

		if rm.IsBasicProduction("INVALID") {
			t.Error("Expected false for non-existent product")
		}
	})

	t.Run("CanProducePremium", func(t *testing.T) {
		inventory := map[string]int{
			"FOSFO": 10,
			"PITA":  5,
		}

		if !rm.CanProducePremium("GUACA", inventory) {
			t.Error("Expected to be able to produce GUACA")
		}

		// Insufficient ingredients
		insufficientInv := map[string]int{
			"FOSFO": 2,
			"PITA":  1,
		}

		if rm.CanProducePremium("GUACA", insufficientInv) {
			t.Error("Expected not to be able to produce GUACA with insufficient ingredients")
		}

		// Missing ingredient
		missingInv := map[string]int{
			"FOSFO": 10,
		}

		if rm.CanProducePremium("GUACA", missingInv) {
			t.Error("Expected not to be able to produce GUACA with missing ingredients")
		}
	})

	t.Run("ConsumeIngredients", func(t *testing.T) {
		inventory := map[string]int{
			"FOSFO": 10,
			"PITA":  5,
		}

		err := rm.ConsumeIngredients("GUACA", inventory)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if inventory["FOSFO"] != 5 {
			t.Errorf("Expected FOSFO to be 5, got %d", inventory["FOSFO"])
		}

		if inventory["PITA"] != 2 {
			t.Errorf("Expected PITA to be 2, got %d", inventory["PITA"])
		}
	})

	t.Run("ConsumeIngredients_Insufficient", func(t *testing.T) {
		inventory := map[string]int{
			"FOSFO": 2,
			"PITA":  1,
		}

		err := rm.ConsumeIngredients("GUACA", inventory)
		if err == nil {
			t.Error("Expected error when consuming with insufficient ingredients")
		}
	})

	t.Run("ConsumeIngredients_Basic", func(t *testing.T) {
		inventory := map[string]int{}

		err := rm.ConsumeIngredients("PALTA-OIL", inventory)
		if err != nil {
			t.Errorf("Expected no error for basic production, got: %v", err)
		}
	})

	t.Run("GetRequiredIngredients", func(t *testing.T) {
		ingredients, err := rm.GetRequiredIngredients("GUACA")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if ingredients["FOSFO"] != 5 {
			t.Errorf("Expected FOSFO: 5, got %d", ingredients["FOSFO"])
		}

		if ingredients["PITA"] != 3 {
			t.Errorf("Expected PITA: 3, got %d", ingredients["PITA"])
		}

		// Basic production
		basicIngr, err := rm.GetRequiredIngredients("PALTA-OIL")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(basicIngr) != 0 {
			t.Error("Expected empty ingredients for basic production")
		}
	})

	t.Run("GetMissingIngredients", func(t *testing.T) {
		inventory := map[string]int{
			"FOSFO": 2,
			"PITA":  1,
		}

		missing, err := rm.GetMissingIngredients("GUACA", inventory)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if missing["FOSFO"] != 3 {
			t.Errorf("Expected missing FOSFO: 3, got %d", missing["FOSFO"])
		}

		if missing["PITA"] != 2 {
			t.Errorf("Expected missing PITA: 2, got %d", missing["PITA"])
		}

		// With sufficient inventory
		sufficientInv := map[string]int{
			"FOSFO": 10,
			"PITA":  10,
		}

		missing, err = rm.GetMissingIngredients("GUACA", sufficientInv)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(missing) != 0 {
			t.Error("Expected no missing ingredients")
		}
	})
}
