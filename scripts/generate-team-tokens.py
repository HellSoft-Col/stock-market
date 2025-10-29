#!/usr/bin/env python3
"""
Team Token Generator for Intergalactic Avocado Stock Exchange
Creates fun team names and secure tokens for the trading competition.

Usage:
    python scripts/generate-team-tokens.py [number_of_teams]
"""

import random
import secrets
import string
import argparse
import json
from datetime import datetime

def get_original_teams():
    """Get the 12 original team names and species from 31 Minutos universe."""
    return [
        {
            "name": "Avocultores del Hueso Cósmico",
            "species": "Básico",
            "specialty": "PALTA-OIL",
            "recipe": "5 FOSFO + 3 PITA"
        },
        {
            "name": "Monjes del Guacamole Estelar", 
            "species": "Premium",
            "specialty": "FOSFO",
            "recipe": "5 FOSFO + 3 PITA"
        },
        {
            "name": "Cosechadores de Semillas",
            "species": "Premium",
            "specialty": "PITA",
            "recipe": "8 NUCREM"
        },
        {
            "name": "Mineros de Guacatrones",
            "species": "Premium",
            "specialty": "H-GUACA",
            "recipe": "12 PALTA-OIL + 5 CASCAR-ALLOY"
        },
        {
            "name": "Someliers de Aceite",
            "species": "Premium",
            "specialty": "PALTA-OIL",
            "recipe": "4 SEBO"
        },
        {
            "name": "Orfebres de Cáscara",
            "species": "Premium",
            "specialty": "FOSFO",
            "recipe": "10 GTRON + 6 FOSFO"
        },
        {
            "name": "Ingenieros Holo-Aguacate",
            "species": "Premium",
            "specialty": "H-GUACA",
            "recipe": "12 PALTA-OIL + 5 CASCAR-ALLOY"
        },
        {
            "name": "Arpistas de Pita-Pita",
            "species": "Premium",
            "specialty": "PITA",
            "recipe": "8 NUCREM"
        },
        {
            "name": "Cartógrafos de Fosfolima",
            "species": "Premium",
            "specialty": "FOSFO",
            "recipe": "4 SEBO"
        },
        {
            "name": "Mensajeros del Núcleo",
            "species": "Premium",
            "specialty": "NUCREM",
            "recipe": "8 NUCREM"
        },
        {
            "name": "Alquimistas de Palta",
            "species": "Premium",
            "specialty": "PALTA-OIL",
            "recipe": "5 FOSFO + 3 PITA"
        },
        {
            "name": "Forjadores Holográficos",
            "species": "Premium",
            "specialty": "H-GUACA",
            "recipe": "10 GTRON + 6 FOSFO"
        }
    ]

def generate_team_names():
    """Get the original 12 team names."""
    teams = get_original_teams()
    return [team["name"] for team in teams]

def generate_token():
    """Generate a secure token with TK- prefix."""
    # Generate 24 random characters (uppercase, lowercase, numbers)
    token_chars = string.ascii_uppercase + string.ascii_lowercase + string.digits
    token_suffix = ''.join(secrets.choice(token_chars) for _ in range(24))
    return f"TK-{token_suffix}"

def get_team_species(team_name):
    """Get the species for a specific team name."""
    teams = get_original_teams()
    for team in teams:
        if team["name"] == team_name:
            return team["species"]
    return "Premium"  # Default fallback

def generate_teams(num_teams):
    """Generate team data with names, tokens, and species."""
    original_teams = get_original_teams()
    
    if num_teams > len(original_teams):
        print(f"Warning: Requested {num_teams} teams but only {len(original_teams)} original teams available.")
        print(f"Generating {len(original_teams)} teams instead.")
        num_teams = len(original_teams)
    
    teams = []
    
    # Shuffle the original teams for variety
    shuffled_teams = original_teams.copy()
    random.shuffle(shuffled_teams)
    
    for i in range(num_teams):
        team_data = shuffled_teams[i]
        
        team = {
            "teamName": team_data["name"],
            "token": generate_token(),
            "species": team_data["species"],
            "specialty": team_data["specialty"],
            "recipe": team_data["recipe"],
            "initialBalance": 100000,  # $100,000 starting balance
            "authorizedProducts": ["FOSFO", "PITA", "PALTA-OIL", "GUACA", "SEBO", "H-GUACA", "NUCREM", "CASCAR-ALLOY", "GTRON"]
        }
        teams.append(team)
    
    return teams

def save_teams_json(teams, filename):
    """Save teams to JSON file for database seeding."""
    output = {
        "generated_at": datetime.now().isoformat(),
        "total_teams": len(teams),
        "teams": teams
    }
    
    with open(filename, 'w', encoding='utf-8') as f:
        json.dump(output, f, indent=2, ensure_ascii=False)
    
    print(f"✅ Teams saved to {filename}")

def save_teams_csv(teams, filename):
    """Save teams to CSV file for easy distribution."""
    import csv
    
    with open(filename, 'w', newline='', encoding='utf-8') as f:
        writer = csv.writer(f)
        writer.writerow(['Team Name', 'Token', 'Species', 'Specialty', 'Recipe', 'Initial Balance'])
        
        for team in teams:
            writer.writerow([
                team['teamName'],
                team['token'], 
                team['species'],
                team['specialty'],
                team['recipe'],
                f"${team['initialBalance']:,}"
            ])
    
    print(f"✅ Teams saved to {filename}")

def print_teams_table(teams):
    """Print teams in a formatted table."""
    print("\n" + "="*130)
    print("🥑 INTERGALACTIC AVOCADO STOCK EXCHANGE - TEAM TOKENS 🚀")
    print("="*130)
    print(f"{'#':<3} {'Team Name':<35} {'Token':<30} {'Species':<10} {'Specialty':<12} {'Balance':<10}")
    print("-"*130)
    
    for i, team in enumerate(teams, 1):
        print(f"{i:<3} {team['teamName']:<35} {team['token']:<30} {team['species']:<10} {team['specialty']:<12} ${team['initialBalance']:,}")
    
    print("-"*130)
    print(f"Total Teams: {len(teams)}")
    print("="*130)

def main():
    parser = argparse.ArgumentParser(
        description="Generate team tokens for Intergalactic Avocado Stock Exchange",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
    python scripts/generate-team-tokens.py 20           # Generate 20 teams
    python scripts/generate-team-tokens.py 30 --seed 42 # Generate 30 teams with specific seed
    python scripts/generate-team-tokens.py 25 --json-only # Generate 25 teams, JSON output only
        """
    )
    
    parser.add_argument('num_teams', type=int, nargs='?', default=25,
                       help='Number of teams to generate (default: 25)')
    parser.add_argument('--seed', type=int, help='Random seed for reproducible results')
    parser.add_argument('--json-only', action='store_true', 
                       help='Only generate JSON file (skip CSV and table output)')
    parser.add_argument('--output-dir', default='./output', 
                       help='Output directory for generated files (default: ./output)')
    
    args = parser.parse_args()
    
    # Set random seed if provided
    if args.seed:
        random.seed(args.seed)
        print(f"🎲 Using random seed: {args.seed}")
    
    # Create output directory
    import os
    os.makedirs(args.output_dir, exist_ok=True)
    
    # Generate teams
    print(f"🚀 Generating {args.num_teams} teams...")
    teams = generate_teams(args.num_teams)
    
    # Sort teams by name for better organization
    teams.sort(key=lambda x: x['teamName'])
    
    # Generate timestamp for filenames
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    
    # Save files
    json_filename = f"{args.output_dir}/teams_{timestamp}.json"
    save_teams_json(teams, json_filename)
    
    if not args.json_only:
        csv_filename = f"{args.output_dir}/teams_{timestamp}.csv" 
        save_teams_csv(teams, csv_filename)
        print_teams_table(teams)
        
        print(f"\n📁 Files generated:")
        print(f"   • JSON (for database): {json_filename}")
        print(f"   • CSV (for distribution): {csv_filename}")
    else:
        print(f"\n📁 File generated: {json_filename}")
    
    print(f"\n🎯 Next steps:")
    print(f"   1. Review the generated teams above")
    print(f"   2. Use the JSON file to seed the database:")
    print(f"      go run cmd/seed/main.go -teams {json_filename}")
    print(f"   3. Distribute tokens from the CSV file to teams")
    
    print(f"\n🥑 Ready for intergalactic avocado trading! 🚀")

if __name__ == "__main__":
    main()