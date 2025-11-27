#!/usr/bin/env python3
"""
Update all team recipes in MongoDB to match the correct species recipes.
This fixes the issue where teams don't have proper premium production recipes.
"""

from pymongo import MongoClient
import sys


# Recipe definitions matching the correct table
RECIPES_BY_SPECIES = {
    "Avocultores": {
        "basic": "PALTA-OIL",
        "premium": {
            "GUACA": {
                "type": "PREMIUM",
                "ingredients": {"FOSFO": 5, "PITA": 3},
                "premiumBonus": 1.3,
            },
            "SEBO": {
                "type": "PREMIUM",
                "ingredients": {"NUCREM": 8},
                "premiumBonus": 1.3,
            },
        },
    },
    "Monjes de Fosforescencia": {
        "basic": "FOSFO",
        "premium": {
            "GUACA": {
                "type": "PREMIUM",
                "ingredients": {"PALTA-OIL": 5, "PITA": 3},
                "premiumBonus": 1.3,
            },
            "NUCREM": {
                "type": "PREMIUM",
                "ingredients": {"SEBO": 6},
                "premiumBonus": 1.3,
            },
        },
    },
    "Cosechadores de Pita": {
        "basic": "PITA",
        "premium": {
            "SEBO": {
                "type": "PREMIUM",
                "ingredients": {"NUCREM": 8},
                "premiumBonus": 1.3,
            },
            "CASCAR-ALLOY": {
                "type": "PREMIUM",
                "ingredients": {"FOSFO": 10},
                "premiumBonus": 1.3,
            },
        },
    },
    "Herreros Cósmicos": {
        "basic": "CASCAR-ALLOY",
        "premium": {
            "QUANTUM-PULP": {
                "type": "PREMIUM",
                "ingredients": {"PALTA-OIL": 7},
                "premiumBonus": 1.3,
            },
            "SKIN-WRAP": {
                "type": "PREMIUM",
                "ingredients": {"ASTRO-BUTTER": 12},
                "premiumBonus": 1.3,
            },
        },
    },
    "Extractores": {
        "basic": "QUANTUM-PULP",
        "premium": {
            "NUCREM": {
                "type": "PREMIUM",
                "ingredients": {"SEBO": 6},
                "premiumBonus": 1.3,
            },
            "FOSFO": {
                "type": "PREMIUM",
                "ingredients": {"SKIN-WRAP": 9},
                "premiumBonus": 1.3,
            },
        },
    },
    "Tejemanteles": {
        "basic": "SKIN-WRAP",
        "premium": {
            "PITA": {
                "type": "PREMIUM",
                "ingredients": {"CASCAR-ALLOY": 8},
                "premiumBonus": 1.3,
            },
            "ASTRO-BUTTER": {
                "type": "PREMIUM",
                "ingredients": {"GUACA": 10},
                "premiumBonus": 1.3,
            },
        },
    },
    "Cremeros Astrales": {
        "basic": "ASTRO-BUTTER",
        "premium": {
            "CASCAR-ALLOY": {
                "type": "PREMIUM",
                "ingredients": {"FOSFO": 10},
                "premiumBonus": 1.3,
            },
            "PALTA-OIL": {
                "type": "PREMIUM",
                "ingredients": {"QUANTUM-PULP": 7},
                "premiumBonus": 1.3,
            },
        },
    },
    "Mineros del Sebo": {
        "basic": "SEBO",
        "premium": {
            "ASTRO-BUTTER": {
                "type": "PREMIUM",
                "ingredients": {"GUACA": 10},
                "premiumBonus": 1.3,
            },
            "GUACA": {
                "type": "PREMIUM",
                "ingredients": {"PALTA-OIL": 5, "PITA": 3},
                "premiumBonus": 1.3,
            },
        },
    },
    "Núcleo Cremero": {
        "basic": "NUCREM",
        "premium": {
            "SKIN-WRAP": {
                "type": "PREMIUM",
                "ingredients": {"ASTRO-BUTTER": 12},
                "premiumBonus": 1.3,
            },
            "QUANTUM-PULP": {
                "type": "PREMIUM",
                "ingredients": {"PALTA-OIL": 7},
                "premiumBonus": 1.3,
            },
        },
    },
    "Destiladores": {
        "basic": "GUACA",
        "premium": {
            "PALTA-OIL": {
                "type": "PREMIUM",
                "ingredients": {"QUANTUM-PULP": 7},
                "premiumBonus": 1.3,
            },
            "FOSFO": {
                "type": "PREMIUM",
                "ingredients": {"SKIN-WRAP": 9},
                "premiumBonus": 1.3,
            },
        },
    },
    "Cartógrafos": {
        "basic": "GUACA",
        "premium": {
            "NUCREM": {
                "type": "PREMIUM",
                "ingredients": {"SEBO": 6},
                "premiumBonus": 1.3,
            },
            "PITA": {
                "type": "PREMIUM",
                "ingredients": {"CASCAR-ALLOY": 8},
                "premiumBonus": 1.3,
            },
        },
    },
    "Someliers Andorianos": {
        "basic": "PALTA-OIL",
        "premium": {
            "SEBO": {
                "type": "PREMIUM",
                "ingredients": {"NUCREM": 8},
                "premiumBonus": 1.3,
            },
            "CASCAR-ALLOY": {
                "type": "PREMIUM",
                "ingredients": {"FOSFO": 10},
                "premiumBonus": 1.3,
            },
        },
    },
}


def build_recipes_for_species(species):
    """Build the complete recipe map for a species."""
    if species not in RECIPES_BY_SPECIES:
        print(f"⚠️  Warning: Unknown species {species}")
        return {}

    species_data = RECIPES_BY_SPECIES[species]
    recipes = {}

    # Add basic recipe
    basic_product = species_data["basic"]
    recipes[basic_product] = {"type": "BASIC", "ingredients": {}, "premiumBonus": 1.0}

    # Add premium recipes
    for product, recipe in species_data["premium"].items():
        recipes[product] = recipe

    return recipes


def main():
    # Connect to MongoDB
    print("🔌 Connecting to MongoDB...")
    client = MongoClient(
        "mongodb://localhost:27017,localhost:27018,localhost:27019/?replicaSet=rs0"
    )
    db = client["avocado_exchange"]
    teams_collection = db["teams"]

    # Get all teams
    teams = list(teams_collection.find({}))
    print(f"📊 Found {len(teams)} teams")

    updated_count = 0
    error_count = 0

    for team in teams:
        team_name = team.get("teamName", "Unknown")
        species = team.get("species", "Unknown")

        print(f"\n🔄 Processing {team_name} ({species})...")

        # Build correct recipes for this species
        correct_recipes = build_recipes_for_species(species)

        if not correct_recipes:
            print(f"   ⚠️  Skipping {team_name} - unknown species")
            error_count += 1
            continue

        # Update the team's recipes
        try:
            result = teams_collection.update_one(
                {"_id": team["_id"]}, {"$set": {"recipes": correct_recipes}}
            )

            if result.modified_count > 0:
                print(f"   ✅ Updated recipes for {team_name}")
                print(
                    f"      Basic: {[k for k, v in correct_recipes.items() if v['type'] == 'BASIC']}"
                )
                print(
                    f"      Premium: {[k for k, v in correct_recipes.items() if v['type'] == 'PREMIUM']}"
                )
                updated_count += 1
            else:
                print(f"   ℹ️  No changes needed for {team_name}")
        except Exception as e:
            print(f"   ❌ Error updating {team_name}: {e}")
            error_count += 1

    print(f"\n{'=' * 60}")
    print(f"📈 Summary:")
    print(f"   Total teams: {len(teams)}")
    print(f"   Updated: {updated_count}")
    print(f"   Errors: {error_count}")
    print(f"   Unchanged: {len(teams) - updated_count - error_count}")
    print(f"{'=' * 60}")

    if error_count == 0:
        print("🎉 All teams updated successfully!")
    else:
        print(f"⚠️  {error_count} teams had errors")
        sys.exit(1)


if __name__ == "__main__":
    main()
