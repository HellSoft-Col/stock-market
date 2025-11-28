# Recipe Fix Guide

## Problem
Teams in the database may not have the correct `species` field set, which prevents them from getting the correct production recipes based on the official 12 species table.

## Solution
We've created a Go command-line tool that will:
1. Update each team's `species` field based on their team name
2. Generate and assign the correct recipes (1 basic + 2 premium) for each species
3. Save all changes to the database

## The 12 Species Recipes

| Species | Basic (Free) | Premium 1 (+30%) | Premium 2 (+30%) |
|---------|--------------|------------------|------------------|
| **Avocultores** | PALTA-OIL | GUACA (5 FOSFO + 3 PITA) | SEBO (8 NUCREM) |
| **Monjes de Fosforescencia** | FOSFO | GUACA (5 PALTA-OIL + 3 PITA) | NUCREM (6 SEBO) |
| **Cosechadores de Pita** | PITA | SEBO (8 NUCREM) | CASCAR-ALLOY (10 FOSFO) |
| **Herreros Cósmicos** | CASCAR-ALLOY | QUANTUM-PULP (7 PALTA-OIL) | SKIN-WRAP (12 ASTRO-BUTTER) |
| **Extractores** | QUANTUM-PULP | NUCREM (6 SEBO) | FOSFO (9 SKIN-WRAP) |
| **Tejemanteles** | SKIN-WRAP | PITA (8 CASCAR-ALLOY) | ASTRO-BUTTER (10 GUACA) |
| **Cremeros Astrales** | ASTRO-BUTTER | CASCAR-ALLOY (10 FOSFO) | PALTA-OIL (7 QUANTUM-PULP) |
| **Mineros del Sebo** | SEBO | ASTRO-BUTTER (10 GUACA) | GUACA (5 PALTA-OIL + 3 PITA) |
| **Núcleo Cremero** | NUCREM | SKIN-WRAP (12 ASTRO-BUTTER) | QUANTUM-PULP (7 PALTA-OIL) |
| **Destiladores** | GUACA | PALTA-OIL (7 QUANTUM-PULP) | FOSFO (9 SKIN-WRAP) |
| **Cartógrafos** | GUACA | NUCREM (6 SEBO) | PITA (8 CASCAR-ALLOY) |
| **Someliers Andorianos** | PALTA-OIL | SEBO (8 NUCREM) | CASCAR-ALLOY (10 FOSFO) |

## How to Use

### Option 1: Dry Run (Recommended First)

Run the tool in dry-run mode to see what would be changed WITHOUT making actual changes:

```bash
go run cmd/update-recipes/main.go --dry-run
```

This will show you:
- Which teams will be updated
- What their new species will be  
- What recipes they'll get
- No changes will be made to the database

### Option 2: Apply the Fix

Once you've verified the dry-run output looks correct, apply the changes:

```bash
go run cmd/update-recipes/main.go
```

Or with a specific config file:

```bash
go run cmd/update-recipes/main.go --config config.production.yaml
```

### Option 3: Use the Admin Dashboard

After running the Go tool to fix species, you can also use the admin dashboard:

1. Log into the admin dashboard
2. Click the **"Update All Recipes"** button in Admin Controls
3. This will regenerate recipes for all teams based on their (now correct) species

## Team Name to Species Mapping

The tool uses the following mapping based on `automated-clients.yaml`:

| Team Name | Species |
|-----------|---------|
| Alquimistas de Palta | Avocultores |
| Arpistas de Pita-Pita | Cosechadores de Pita |
| Avocultores del Hueso Cósmico | Avocultores |
| Cartógrafos de Fosfolima | Cartógrafos |
| Cosechadores de Semillas | Cosechadores de Pita |
| Forjadores Holográficos | Herreros Cósmicos |
| Ingenieros Holo-Aguacate | Núcleo Cremero |
| Mensajeros del Núcleo | Núcleo Cremero |
| Monjes del Guacamole Estelar | Destiladores |
| Orfebres de Cáscara | Herreros Cósmicos |
| Someliers de Aceite | Someliers Andorianos |

## Expected Output

When you run the tool, you should see output like:

```
✅ Team updated successfully team="Alquimistas de Palta" species="Avocultores" basicProduct="PALTA-OIL" recipeCount=3
✅ Team updated successfully team="Arpistas de Pita-Pita" species="Cosechadores de Pita" basicProduct="PITA" recipeCount=3
...
📝 Update Summary updated=11 skipped=1 errors=0 total=12
🎉 Teams updated successfully!
✅ All teams now have correct species and recipes
```

## Verification

After running the fix, verify in the admin dashboard:

1. Go to **Team Management** tab
2. Check that each team has the correct **Species** value
3. The **Species** column should match the mapping above

## Troubleshooting

### Error: "Team not found in species mapping"
- This means there's a team in the database that's not in our mapping
- Add the team name and its species to the `teamSpeciesMapping` in `cmd/update-recipes/main.go`
- Or remove/rename the team in the database

### Error: "Unknown species"
- The species value doesn't match one of the 12 official species
- Check the species mapping and ensure it's spelled correctly

### No recipes showing up for teams
- Make sure you ran the update-recipes tool first to set correct species
- Then click "Update All Recipes" in the admin dashboard
- Check the browser console and server logs for any errors

## Files Modified

- `cmd/update-recipes/main.go` - New CLI tool to fix species and recipes
- `internal/domain/interfaces.go` - Added `UpdateSpecies` method to TeamRepository
- `internal/repository/mongodb/team_repository.go` - Implemented `UpdateSpecies` method

## Next Steps

After fixing recipes, teams should be able to:
1. Produce their basic product for free
2. Produce 2 premium products (requires ingredients from other teams)
3. Trade with other teams to get ingredients they need
4. Earn +30% production bonus on premium products
