package production

import "fmt"

// Recipe represents a production recipe
type Recipe struct {
	Product      string         `json:"product"      yaml:"product"`
	Ingredients  map[string]int `json:"ingredients"  yaml:"ingredients"`  // nil for basic production
	PremiumBonus float64        `json:"premiumBonus" yaml:"premiumBonus"` // typically 1.30 for +30%
}

// RecipeManager manages recipes and validates ingredient availability
type RecipeManager struct {
	recipes map[string]*Recipe
}

// NewRecipeManager creates a new recipe manager
func NewRecipeManager(recipes map[string]*Recipe) *RecipeManager {
	return &RecipeManager{
		recipes: recipes,
	}
}

// GetRecipe returns the recipe for a product
func (rm *RecipeManager) GetRecipe(product string) (*Recipe, error) {
	recipe, exists := rm.recipes[product]
	if !exists {
		return nil, fmt.Errorf("recipe not found: %s", product)
	}
	return recipe, nil
}

// IsBasicProduction checks if a product can be produced without ingredients
func (rm *RecipeManager) IsBasicProduction(product string) bool {
	recipe, exists := rm.recipes[product]
	if !exists {
		return false
	}
	return len(recipe.Ingredients) == 0
}

// CanProducePremium checks if all ingredients are available for premium production
func (rm *RecipeManager) CanProducePremium(product string, inventory map[string]int) bool {
	recipe, exists := rm.recipes[product]
	if !exists || recipe.Ingredients == nil {
		return false
	}

	// Check all ingredients are available
	for ingredient, required := range recipe.Ingredients {
		available := inventory[ingredient]
		if available < required {
			return false
		}
	}

	return true
}

// ConsumeIngredients deducts ingredients from inventory
func (rm *RecipeManager) ConsumeIngredients(product string, inventory map[string]int) error {
	recipe, exists := rm.recipes[product]
	if !exists {
		return fmt.Errorf("recipe not found: %s", product)
	}

	if recipe.Ingredients == nil {
		return nil // Basic production, no ingredients to consume
	}

	// Verify ingredients before consuming (safety check)
	for ingredient, required := range recipe.Ingredients {
		available := inventory[ingredient]
		if available < required {
			return fmt.Errorf("insufficient %s: have %d, need %d", ingredient, available, required)
		}
	}

	// Deduct ingredients
	for ingredient, required := range recipe.Ingredients {
		inventory[ingredient] -= required
	}

	return nil
}

// GetRequiredIngredients returns the ingredients needed for a product
func (rm *RecipeManager) GetRequiredIngredients(product string) (map[string]int, error) {
	recipe, exists := rm.recipes[product]
	if !exists {
		return nil, fmt.Errorf("recipe not found: %s", product)
	}

	if len(recipe.Ingredients) == 0 {
		return map[string]int{}, nil
	}

	// Return a copy to prevent modification
	ingredients := make(map[string]int, len(recipe.Ingredients))
	for k, v := range recipe.Ingredients {
		ingredients[k] = v
	}

	return ingredients, nil
}

// GetMissingIngredients returns what ingredients are missing for production
func (rm *RecipeManager) GetMissingIngredients(product string, inventory map[string]int) (map[string]int, error) {
	recipe, exists := rm.recipes[product]
	if !exists {
		return nil, fmt.Errorf("recipe not found: %s", product)
	}

	if len(recipe.Ingredients) == 0 {
		return map[string]int{}, nil
	}

	missing := make(map[string]int)
	for ingredient, required := range recipe.Ingredients {
		available := inventory[ingredient]
		if available < required {
			missing[ingredient] = required - available
		}
	}

	return missing, nil
}

// GetAllRecipes returns all recipes
func (rm *RecipeManager) GetAllRecipes() map[string]*Recipe {
	return rm.recipes
}
