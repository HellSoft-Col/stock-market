package production

import (
	"testing"
)

func TestProductionCalculator_Avocultores(t *testing.T) {
	// Avocultores role from the guide
	role := &Role{
		Branches:    2,
		MaxDepth:    4,
		Decay:       0.7651,
		BaseEnergy:  3.0,
		LevelEnergy: 2.0,
	}

	calc := NewProductionCalculator()
	units := calc.CalculateUnits(role)

	// Expected calculation from guide:
	// Level 0: (3.0 + 2.0×0) × (0.7651^0 × 2^0) = 3.0 × 1.0 = 3
	// Level 1: (3.0 + 2.0×1) × (0.7651^1 × 2^1) = 5.0 × 1.530 = 8
	// Level 2: (3.0 + 2.0×2) × (0.7651^2 × 2^2) = 7.0 × 2.344 = 16
	// Level 3: (3.0 + 2.0×3) × (0.7651^3 × 2^3) = 9.0 × 3.599 = 32
	// Level 4: (3.0 + 2.0×4) × (0.7651^4 × 2^4) = 11.0 × 5.521 = 61
	// Total should be around 120-121

	if units < 118 || units > 123 {
		t.Errorf("Expected units around 120-121, got %d", units)
	}

	t.Logf("Avocultores basic production: %d units", units)
}

func TestProductionCalculator_PremiumBonus(t *testing.T) {
	calc := NewProductionCalculator()

	baseUnits := 13
	bonus := 1.30 // +30%

	premiumUnits := calc.ApplyPremiumBonus(baseUnits, bonus)

	// 13 × 1.30 = 16.9 ≈ 17
	expected := 17

	if premiumUnits != expected {
		t.Errorf("Expected %d premium units, got %d", expected, premiumUnits)
	}

	t.Logf("Basic: %d units → Premium (+30%%): %d units", baseUnits, premiumUnits)
}

func TestProductionCalculator_DifferentSpecies(t *testing.T) {
	testCases := []struct {
		name     string
		role     *Role
		minUnits int
		maxUnits int
	}{
		{
			name: "Monjes",
			role: &Role{
				Branches:    2,
				MaxDepth:    3,
				Decay:       0.8,
				BaseEnergy:  4.0,
				LevelEnergy: 1.5,
			},
			minUnits: 60,
			maxUnits: 70,
		},
		{
			name: "Cosechadores",
			role: &Role{
				Branches:    3,
				MaxDepth:    2,
				Decay:       0.9,
				BaseEnergy:  2.0,
				LevelEnergy: 3.0,
			},
			minUnits: 20,
			maxUnits: 100,
		},
	}

	calc := NewProductionCalculator()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			units := calc.CalculateUnits(tc.role)

			if units < tc.minUnits || units > tc.maxUnits {
				t.Errorf("%s: Expected units between %d-%d, got %d",
					tc.name, tc.minUnits, tc.maxUnits, units)
			}

			t.Logf("%s production: %d units", tc.name, units)
		})
	}
}

func TestProductionCalculator_ZeroDepth(t *testing.T) {
	role := &Role{
		Branches:    2,
		MaxDepth:    0,
		Decay:       0.7,
		BaseEnergy:  5.0,
		LevelEnergy: 2.0,
	}

	calc := NewProductionCalculator()
	units := calc.CalculateUnits(role)

	// With MaxDepth=0, only level 0 contributes
	// (5.0 + 2.0×0) × (0.7^0 × 2^0) = 5.0 × 1.0 = 5
	expected := 5

	if units != expected {
		t.Errorf("Expected %d units, got %d", expected, units)
	}
}
