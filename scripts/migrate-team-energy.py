#!/usr/bin/env python3
"""
Migration script to add BaseEnergy and LevelEnergy fields to existing teams.

Usage:
    python migrate-team-energy.py <mongodb_url>

Example:
    python migrate-team-energy.py mongodb://localhost:27017/stockmarket
    python migrate-team-energy.py "mongodb+srv://user:pass@cluster.mongodb.net/stockmarket"
"""

import sys
from pymongo import MongoClient
from pymongo.errors import ConnectionFailure, OperationFailure

# Default energy values (matching the seed script)
DEFAULT_BASE_ENERGY = 3.0
DEFAULT_LEVEL_ENERGY = 2.0


def migrate_teams(mongodb_url):
    """Migrate teams to add energy fields if missing."""

    print(f"🔗 Connecting to MongoDB...")
    print(
        f"   URL: {mongodb_url.split('@')[-1] if '@' in mongodb_url else mongodb_url}"
    )

    try:
        # Connect to MongoDB
        client = MongoClient(mongodb_url, serverSelectionTimeoutMS=5000)

        # Test connection
        client.admin.command("ping")
        print("✅ Connected successfully!\n")

        # Get database name from URL
        db_name = mongodb_url.split("/")[-1].split("?")[0]
        db = client[db_name]
        teams_collection = db["teams"]

        # Find teams without energy fields or with zero values
        print("🔍 Searching for teams that need migration...")

        # Query for teams where role.baseEnergy or role.levelEnergy is missing or zero
        query = {
            "$or": [
                {"role.baseEnergy": {"$exists": False}},
                {"role.levelEnergy": {"$exists": False}},
                {"role.baseEnergy": 0},
                {"role.levelEnergy": 0},
            ]
        }

        teams_to_migrate = list(teams_collection.find(query))

        if not teams_to_migrate:
            print("✅ No teams need migration. All teams have energy values!")
            return 0

        print(f"📊 Found {len(teams_to_migrate)} team(s) that need migration:\n")

        for team in teams_to_migrate:
            team_name = team.get("teamName", "Unknown")
            current_base = team.get("role", {}).get("baseEnergy", "missing")
            current_level = team.get("role", {}).get("levelEnergy", "missing")
            print(f"   • {team_name}")
            print(f"     Current baseEnergy: {current_base}")
            print(f"     Current levelEnergy: {current_level}")

        print("\n" + "=" * 60)
        print("⚠️  This will update the following fields:")
        print(f"   • role.baseEnergy = {DEFAULT_BASE_ENERGY}")
        print(f"   • role.levelEnergy = {DEFAULT_LEVEL_ENERGY}")
        print("=" * 60 + "\n")

        response = (
            input("Do you want to proceed with the migration? (yes/no): ")
            .strip()
            .lower()
        )

        if response not in ["yes", "y"]:
            print("❌ Migration cancelled.")
            return 1

        print("\n🔄 Starting migration...\n")

        # Perform the migration
        updated_count = 0
        failed_count = 0

        for team in teams_to_migrate:
            team_name = team.get("teamName", "Unknown")

            try:
                # Update the team with energy values
                result = teams_collection.update_one(
                    {"_id": team["_id"]},
                    {
                        "$set": {
                            "role.baseEnergy": DEFAULT_BASE_ENERGY,
                            "role.levelEnergy": DEFAULT_LEVEL_ENERGY,
                        }
                    },
                )

                if result.modified_count > 0:
                    print(f"   ✅ Updated: {team_name}")
                    updated_count += 1
                else:
                    print(f"   ⚠️  No changes needed: {team_name}")

            except Exception as e:
                print(f"   ❌ Failed to update {team_name}: {e}")
                failed_count += 1

        # Summary
        print("\n" + "=" * 60)
        print("📊 Migration Summary:")
        print(f"   • Teams updated: {updated_count}")
        print(f"   • Teams failed: {failed_count}")
        print(f"   • Total processed: {len(teams_to_migrate)}")
        print("=" * 60 + "\n")

        if failed_count == 0:
            print("✅ Migration completed successfully!")
            return 0
        else:
            print(f"⚠️  Migration completed with {failed_count} error(s).")
            return 1

    except ConnectionFailure as e:
        print(f"❌ Failed to connect to MongoDB: {e}")
        print("\nPlease check:")
        print("  1. The MongoDB URL is correct")
        print("  2. MongoDB is running and accessible")
        print("  3. Network connectivity is working")
        return 1

    except OperationFailure as e:
        print(f"❌ MongoDB operation failed: {e}")
        print("\nPlease check:")
        print("  1. Authentication credentials are correct")
        print("  2. User has sufficient permissions")
        return 1

    except Exception as e:
        print(f"❌ Unexpected error: {e}")
        return 1

    finally:
        if "client" in locals():
            client.close()
            print("\n🔌 Disconnected from MongoDB.")


def main():
    """Main entry point."""

    print("=" * 60)
    print("  Team Energy Migration Script")
    print("  Adds BaseEnergy and LevelEnergy to existing teams")
    print("=" * 60 + "\n")

    if len(sys.argv) != 2:
        print("❌ Error: MongoDB URL is required\n")
        print("Usage:")
        print(f"  python {sys.argv[0]} <mongodb_url>\n")
        print("Examples:")
        print(f"  python {sys.argv[0]} mongodb://localhost:27017/stockmarket")
        print(
            f"  python {sys.argv[0]} 'mongodb+srv://user:pass@cluster.mongodb.net/stockmarket'\n"
        )
        return 1

    mongodb_url = sys.argv[1]

    # Validate URL format
    if not mongodb_url.startswith(("mongodb://", "mongodb+srv://")):
        print("❌ Error: Invalid MongoDB URL format")
        print("   URL must start with 'mongodb://' or 'mongodb+srv://'\n")
        return 1

    return migrate_teams(mongodb_url)


if __name__ == "__main__":
    sys.exit(main())
