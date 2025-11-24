package production

import "math"

// Role represents a species' production parameters for the recursive algorithm
type Role struct {
	Branches    int     `json:"branches"    yaml:"branches"`
	MaxDepth    int     `json:"maxDepth"    yaml:"maxDepth"`
	Decay       float64 `json:"decay"       yaml:"decay"`
	BaseEnergy  float64 `json:"baseEnergy"  yaml:"baseEnergy"`
	LevelEnergy float64 `json:"levelEnergy" yaml:"levelEnergy"`
}

// ProductionCalculator implements the recursive production algorithm
type ProductionCalculator struct{}

// NewProductionCalculator creates a new production calculator
func NewProductionCalculator() *ProductionCalculator {
	return &ProductionCalculator{}
}

// CalculateUnits calculates the total units produced using the recursive algorithm
// Formula:
//
//	Energy(level) = baseEnergy + levelEnergy × level
//	Factor(level) = decay^level × branches^level
//	Units(level) = Energy(level) × Factor(level)
//	Total = Σ Units(level) for level = 0 to maxDepth
func (pc *ProductionCalculator) CalculateUnits(role *Role) int {
	return pc.calculateRecursive(0, role)
}

// calculateRecursive is the recursive function that sums contributions from each level
func (pc *ProductionCalculator) calculateRecursive(level int, role *Role) int {
	// BASE CASE: Maximum depth reached
	if level > role.MaxDepth {
		return 0
	}

	// Calculate energy at this level
	energy := role.BaseEnergy + role.LevelEnergy*float64(level)

	// Calculate multiplier factor
	decay := math.Pow(role.Decay, float64(level))
	branches := math.Pow(float64(role.Branches), float64(level))
	factor := decay * branches

	// Contribution from this level
	contribution := int(math.Round(energy * factor))

	// RECURSIVE CASE: Add contributions from lower levels
	return contribution + pc.calculateRecursive(level+1, role)
}

// ApplyPremiumBonus applies the premium production bonus (typically +30%)
func (pc *ProductionCalculator) ApplyPremiumBonus(baseUnits int, bonus float64) int {
	return int(math.Round(float64(baseUnits) * bonus))
}
