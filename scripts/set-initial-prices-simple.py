#!/usr/bin/env python3
"""
Initial Price Preview for Intergalactic Avocado Stock Exchange
Shows initial base prices for all trading products.

Usage:
    python scripts/set-initial-prices-simple.py
"""

import json
from datetime import datetime

def get_initial_prices():
    """Define initial base prices for all products."""
    return {
        # Basic Avocado Products  
        "FOSFO": {
            "base_price": 12.50,
            "description": "Fosforo Aguacate - Premium quality phosphorus-enriched avocados"
        },
        "PITA": {
            "base_price": 8.75,
            "description": "Pita Avocados - Standard trading grade avocados"
        },
        "PALTA-OIL": {
            "base_price": 25.00,
            "description": "Palta Oil - Refined avocado oil for intergalactic cooking"
        },
        "GUACA": {
            "base_price": 15.30,
            "description": "Guacamole Base - Ready-to-prepare guacamole ingredients"
        },
        "SEBO": {
            "base_price": 18.90,
            "description": "Sebo Avocado Fat - Industrial grade avocado fat"
        },
        "H-GUACA": {
            "base_price": 45.60,
            "description": "Hydrogen Guacamole - Advanced processed guacamole with extended shelf life"
        }
    }

def validate_pricing_logic(prices):
    """Validate that pricing logic makes sense."""
    print("🥑 Intergalactic Avocado Stock Exchange - Initial Prices")
    print("=" * 70)
    print("\n📊 Price Analysis:")
    print("=" * 70)
    
    sorted_products = sorted(prices.items(), key=lambda x: x[1]["base_price"])
    
    print(f"{'Product':<12} | {'Base':<8} | {'Bid':<8} | {'Ask':<8} | {'Offer':<8} | Description")
    print("-" * 70)
    
    for product, data in sorted_products:
        base = data["base_price"]
        bid = base * 0.98
        ask = base * 1.02
        offer_price = base * 1.10
        desc = data["description"][:35] + "..." if len(data["description"]) > 35 else data["description"]
        
        print(f"{product:<12} | ${base:>6.2f} | ${bid:>6.2f} | ${ask:>6.2f} | ${offer_price:>6.2f} | {desc}")
    
    print("=" * 70)
    print("\n📈 Pricing Logic:")
    print("• Bid Price: Base Price - 2% (what buyers offer)")
    print("• Ask Price: Base Price + 2% (what sellers want)")  
    print("• Mid Price: Base Price (market equilibrium)")
    print("• Offer Price: Mid Price + 10% (server-generated offers)")
    
    return sorted_products

def export_prices_json(prices, filename):
    """Export prices to JSON file for database seeding."""
    # Create MongoDB-compatible documents
    market_state_docs = []
    
    for product, data in prices.items():
        base_price = data["base_price"]
        
        doc = {
            "product": product,
            "bestBid": round(base_price * 0.98, 2),
            "bestAsk": round(base_price * 1.02, 2),
            "mid": base_price,
            "lastTradePrice": base_price,
            "volume24h": 0,
            "lastUpdated": datetime.utcnow().isoformat(),
            "description": data["description"],
            "initialBasePrice": base_price
        }
        market_state_docs.append(doc)
    
    export_data = {
        "exported_at": datetime.now().isoformat(),
        "description": "Initial market state documents for MongoDB",
        "collection": "market_state",
        "documents": market_state_docs,
        "insert_command": "db.market_state.insertMany(" + json.dumps(market_state_docs, indent=2) + ")"
    }
    
    with open(filename, 'w') as f:
        json.dump(export_data, f, indent=2)
    
    print(f"\n📄 Market state documents exported to: {filename}")
    print(f"   Use this file to manually seed the database or with a MongoDB client")

def main():
    prices = get_initial_prices()
    
    # Show pricing analysis
    sorted_products = validate_pricing_logic(prices)
    
    # Export JSON for database seeding
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    json_filename = f"output/initial_prices_{timestamp}.json"
    export_prices_json(prices, json_filename)
    
    print(f"\n🎯 Next steps:")
    print(f"   1. Review the prices above")
    print(f"   2. Seed the database using the exported JSON:")
    print(f"      - Option A: Import via MongoDB Compass or CLI")
    print(f"      - Option B: Use the Go seeding tool")
    print(f"      - Option C: Install pymongo and use the full script")
    print(f"   3. Start the trading server")
    print(f"   4. Generate team tokens: python scripts/generate-team-tokens.py 10")
    
    print(f"\n💰 Price Summary:")
    total_value = sum(data["base_price"] for data in prices.values())
    avg_price = total_value / len(prices)
    print(f"   • Products: {len(prices)}")
    print(f"   • Average price: ${avg_price:.2f}")
    print(f"   • Price range: ${min(data['base_price'] for data in prices.values()):.2f} - ${max(data['base_price'] for data in prices.values()):.2f}")
    
    print(f"\n🚀 Ready for intergalactic avocado trading!")

if __name__ == "__main__":
    main()