# Team to Species Mapping - Recipe Fix Documentation

## Overview
This document maps the current team names (from `automated-clients.yaml`) to their correct species from the official 12 species table.

## The 12 Species and Their Recipes

Each species can produce:
- **1 BASIC product** (free, no ingredients required)
- **2 PREMIUM products** (require ingredients from other species, +30% production bonus)

| # | Species | Basic (Free) | Premium 1 (+30%) | Premium 2 (+30%) |
|---|---------|--------------|------------------|------------------|
| 1 | **Avocultores** | PALTA-OIL | GUACA (5 FOSFO + 3 PITA) | SEBO (8 NUCREM) |
| 2 | **Monjes de Fosforescencia** | FOSFO | GUACA (5 PALTA-OIL + 3 PITA) | NUCREM (6 SEBO) |
| 3 | **Cosechadores de Pita** | PITA | SEBO (8 NUCREM) | CASCAR-ALLOY (10 FOSFO) |
| 4 | **Herreros Cósmicos** | CASCAR-ALLOY | QUANTUM-PULP (7 PALTA-OIL) | SKIN-WRAP (12 ASTRO-BUTTER) |
| 5 | **Extractores** | QUANTUM-PULP | NUCREM (6 SEBO) | FOSFO (9 SKIN-WRAP) |
| 6 | **Tejemanteles** | SKIN-WRAP | PITA (8 CASCAR-ALLOY) | ASTRO-BUTTER (10 GUACA) |
| 7 | **Cremeros Astrales** | ASTRO-BUTTER | CASCAR-ALLOY (10 FOSFO) | PALTA-OIL (7 QUANTUM-PULP) |
| 8 | **Mineros del Sebo** | SEBO | ASTRO-BUTTER (10 GUACA) | GUACA (5 PALTA-OIL + 3 PITA) |
| 9 | **Núcleo Cremero** | NUCREM | SKIN-WRAP (12 ASTRO-BUTTER) | QUANTUM-PULP (7 PALTA-OIL) |
| 10 | **Destiladores** | GUACA | PALTA-OIL (7 QUANTUM-PULP) | FOSFO (9 SKIN-WRAP) |
| 11 | **Cartógrafos** | GUACA | NUCREM (6 SEBO) | PITA (8 CASCAR-ALLOY) |
| 12 | **Someliers Andorianos** | PALTA-OIL | SEBO (8 NUCREM) | CASCAR-ALLOY (10 FOSFO) |

## Current Team Name to Species Mapping

Based on `automated-clients.yaml`, here's how the current team names map to species:

| Team Name | Species | Basic Product | Authorized Products |
|-----------|---------|---------------|---------------------|
| **Alquimistas de Palta** | Avocultores | PALTA-OIL | PALTA-OIL, GUACA, SEBO |
| **Arpistas de Pita-Pita** | Cosechadores de Pita | PITA | PITA, SEBO, CASCAR-ALLOY |
| **Avocultores del Hueso Cósmico** | Avocultores | PALTA-OIL | PALTA-OIL, GUACA, SEBO |
| **Cartógrafos de Fosfolima** | Cartógrafos | GUACA | GUACA, NUCREM, PITA |
| **Cosechadores de Semillas** | Cosechadores de Pita | PITA | PITA, SEBO, CASCAR-ALLOY |
| **Forjadores Holográficos** | Herreros Cósmicos | CASCAR-ALLOY | CASCAR-ALLOY, QUANTUM-PULP, SKIN-WRAP |
| **Ingenieros Holo-Aguacate** | Núcleo Cremero | NUCREM | NUCREM, SKIN-WRAP, QUANTUM-PULP |
| **Mensajeros del Núcleo** | Núcleo Cremero | NUCREM | NUCREM, SKIN-WRAP, QUANTUM-PULP |
| **Monjes del Guacamole Estelar** | Destiladores | GUACA | GUACA, PALTA-OIL, FOSFO |
| **Orfebres de Cáscara** | Herreros Cósmicos | CASCAR-ALLOY | CASCAR-ALLOY, QUANTUM-PULP, SKIN-WRAP |
| **Someliers de Aceite** | Someliers Andorianos | PALTA-OIL | PALTA-OIL, SEBO, CASCAR-ALLOY |

## Important Note: Species Assignment Strategy

Looking at the current teams, I notice:
- Some teams are named after alchemy/production (e.g., "Alquimistas de Palta" = Avocultores)
- The mapping is based on the **authorized products** in the YAML file
- Each team should have exactly 3 authorized products that match their species recipes

### Missing Species in Current Teams

The following species are **NOT** represented in `automated-clients.yaml`:
- ❌ Monjes de Fosforescencia (FOSFO basic)
- ❌ Extractores (QUANTUM-PULP basic)
- ❌ Tejemanteles (SKIN-WRAP basic)
- ❌ Cremeros Astrales (ASTRO-BUTTER basic)
- ❌ Mineros del Sebo (SEBO basic)

This means only **7 out of 12** species are currently represented! For a complete market, you should add teams for the missing species.

## How to Fix Recipes

### Method 1: Update Team Species in Database (Recommended)

Run the Python script to update team species:

```bash
cd scripts
python fix-team-species.py
```

This will:
1. Update the `species` field for each team in MongoDB
2. Ensure correct species assignments based on authorized products

### Method 2: Use Admin Dashboard

1. Log into the admin dashboard
2. Click the **"Update All Recipes"** button
3. This will rebuild recipes for all teams based on their species

### Method 3: Manual Database Update

Connect to MongoDB and run:

```javascript
// Example: Update "Alquimistas de Palta" to Avocultores species
db.teams.updateOne(
  { teamName: "Alquimistas de Palta" },
  { $set: { species: "Avocultores" } }
)

// Then call UPDATE_ALL_RECIPES from admin dashboard
```

## Recipe Loading Flow

1. **Team Creation** → `species` field is set
2. **Login** → Server loads team's `recipes` from database
3. **Recipe Update** → Admin clicks "Update All Recipes"
   - Server calls `buildRecipesForSpecies(team.Species, basicProduct)`
   - Creates 3 recipes (1 basic + 2 premium) based on species
   - Saves to database
4. **Production** → Team can only produce what's in their `recipes` map

## Verification

After fixing, verify by checking the admin dashboard:

1. Go to **Team Management** tab
2. Edit any team
3. Check the **Species** field matches the table above
4. Click "Update All Recipes" to regenerate recipes based on species

## Server Code Locations

- Recipe building logic: `internal/transport/message_router.go:buildRecipesForSpecies()`
- Species mapping: `internal/transport/message_router.go:getBasicProductForSpecies()`
- Recipe update handler: `internal/transport/message_router.go:handleUpdateAllRecipes()`
- Admin HTML: `web/admin.html:updateAllRecipes()`

## Notes

- The server code already has **correct** recipe definitions
- The issue is teams may not have the correct `species` field set
- Once species are fixed, clicking "Update All Recipes" will load all correct recipes
- All recipes follow the official 12 species table exactly
