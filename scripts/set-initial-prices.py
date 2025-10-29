#!/usr/bin/env python3
"""
Initial Price Setter for Intergalactic Avocado Stock Exchange
Sets initial base prices for all trading products in the database.

Usage:
    python scripts/set-initial-prices.py [--mongodb-uri URI] [--database NAME]
"""

import argparse
import json
import os
from datetime import datetime

try:
    from pymongo import MongoClient
    PYMONGO_AVAILABLE = True
except ImportError:
    PYMONGO_AVAILABLE = False
    print("⚠️  pymongo not installed - database operations unavailable")
    print("   Install with: pip install pymongo")
    print("   Preview mode still available")
    MongoClient = None

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
        },
        
        # Extended Products (if needed)
        "NUCREM": {
            "base_price": 32.40,
            "description": "Nuclear Cream - High-energy avocado cream for space travel"
        },
        "CASCAR-ALLOY": {
            "base_price": 67.80,
            "description": "Cascara Alloy - Processed avocado shells for construction"
        },
        "GTRON": {
            "base_price": 89.25,
            "description": "Graviton Particles - Quantum avocado energy particles"
        }
    }

def connect_to_mongodb(uri, database_name):
    """Connect to MongoDB and return the database."""
    if not PYMONGO_AVAILABLE:
        print("❌ pymongo not available - cannot connect to database")
        return None
        
    try:
        client = MongoClient(uri)
        db = client[database_name]
        
        # Test connection
        db.admin.command('ping')
        print(f"✅ Connected to MongoDB: {database_name}")
        return db
    except Exception as e:
        print(f"❌ Failed to connect to MongoDB: {e}")
        return None

def set_market_state_prices(db, prices):
    """Set initial prices in the market_state collection."""
    market_state_collection = db.market_state
    
    updated_count = 0
    created_count = 0
    
    for product, data in prices.items():
        base_price = data["base_price"]
        description = data["description"]
        
        # Create market state document
        market_state = {
            "product": product,
            "bestBid": base_price * 0.98,  # 2% below base price
            "bestAsk": base_price * 1.02,  # 2% above base price  
            "mid": base_price,
            "lastTradePrice": base_price,
            "volume24h": 0,
            "lastUpdated": datetime.utcnow(),
            "description": description,
            "initialBasePrice": base_price
        }
        
        # Upsert the document
        result = market_state_collection.update_one(
            {"product": product},
            {"$set": market_state},
            upsert=True
        )
        
        if result.upserted_id:
            created_count += 1
            print(f"✅ Created market state for {product}: ${base_price:.2f}")
        else:
            updated_count += 1
            print(f"🔄 Updated market state for {product}: ${base_price:.2f}")
    
    return created_count, updated_count

def create_price_reference_document(db, prices):
    """Create a reference document with all initial prices."""
    price_reference = {
        "_id": "initial_prices_reference",
        "created_at": datetime.utcnow(),
        "version": "1.0",
        "description": "Initial base prices for Intergalactic Avocado Stock Exchange",
        "prices": prices,
        "pricing_notes": [
            "Prices are set to encourage trading activity",
            "Market forces will adjust prices through bid/ask spreads",
            "Offer generator uses +10% premium above mid prices",
            "Base prices can be adjusted by administrators"
        ]
    }
    
    # Store in a metadata collection
    metadata_collection = db.price_metadata
    result = metadata_collection.update_one(
        {"_id": "initial_prices_reference"},
        {"$set": price_reference},
        upsert=True
    )
    
    if result.upserted_id:
        print("✅ Created price reference document")
    else:
        print("🔄 Updated price reference document")

def validate_pricing_logic(prices):
    """Validate that pricing logic makes sense."""
    print("\n📊 Price Analysis:")
    print("=" * 60)
    
    sorted_products = sorted(prices.items(), key=lambda x: x[1]["base_price"])
    
    for product, data in sorted_products:
        base = data["base_price"]
        bid = base * 0.98
        ask = base * 1.02
        offer_price = base * 1.10
        
        print(f"{product:<12} | Base: ${base:>6.2f} | Bid: ${bid:>6.2f} | Ask: ${ask:>6.2f} | Offer: ${offer_price:>6.2f}")
    
    print("=" * 60)
    print("\n📈 Pricing Logic:")
    print("• Bid Price: Base Price - 2% (what buyers offer)")
    print("• Ask Price: Base Price + 2% (what sellers want)")  
    print("• Mid Price: Base Price (market equilibrium)")
    print("• Offer Price: Mid Price + 10% (server-generated offers)")

def export_prices_json(prices, filename):
    """Export prices to JSON file for backup/reference."""
    export_data = {
        "exported_at": datetime.now().isoformat(),
        "description": "Initial base prices for Intergalactic Avocado Stock Exchange",
        "prices": prices
    }
    
    with open(filename, 'w') as f:
        json.dump(export_data, f, indent=2)
    
    print(f"📄 Prices exported to: {filename}")

def main():
    parser = argparse.ArgumentParser(
        description="Set initial base prices for trading products",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
    # Use environment variables for connection
    python scripts/set-initial-prices.py
    
    # Specify connection details
    python scripts/set-initial-prices.py --mongodb-uri "mongodb://localhost:27017" --database "avocado_exchange"
    
    # Preview prices without setting them
    python scripts/set-initial-prices.py --preview-only
        """
    )
    
    parser.add_argument('--mongodb-uri', 
                       default=os.getenv('MONGODB_URI', 'mongodb://localhost:27017'),
                       help='MongoDB connection URI (default: MONGODB_URI env var or localhost)')
    parser.add_argument('--database',
                       default=os.getenv('DATABASE_NAME', 'avocado_exchange_prod'),
                       help='Database name (default: DATABASE_NAME env var or avocado_exchange_prod)')
    parser.add_argument('--preview-only', action='store_true',
                       help='Preview prices without setting them in database')
    parser.add_argument('--export-json', 
                       help='Export prices to JSON file')
    
    args = parser.parse_args()
    
    # Get initial prices
    prices = get_initial_prices()
    
    print("🥑 Intergalactic Avocado Stock Exchange - Initial Price Setter")
    print("=" * 65)
    
    # Validate and show pricing logic
    validate_pricing_logic(prices)
    
    # Export to JSON if requested
    if args.export_json:
        export_prices_json(prices, args.export_json)
    
    # Preview mode - don't connect to database
    if args.preview_only:
        print("\n👁️  PREVIEW MODE - No database changes made")
        print(f"Total products: {len(prices)}")
        return
    
    # Check if pymongo is available
    if not PYMONGO_AVAILABLE:
        print("\n❌ Cannot connect to database - pymongo not installed")
        print("Install with: pip install pymongo")
        print("Use --preview-only flag to see pricing without database connection")
        return
    
    # Connect to database
    print(f"\n🔗 Connecting to database...")
    print(f"URI: {args.mongodb_uri}")
    print(f"Database: {args.database}")
    
    db = connect_to_mongodb(args.mongodb_uri, args.database)
    if not db:
        return
    
    # Set prices in database
    print(f"\n💰 Setting initial prices...")
    created, updated = set_market_state_prices(db, prices)
    
    # Create reference document
    create_price_reference_document(db, prices)
    
    # Summary
    print(f"\n✅ Initial pricing setup complete!")
    print(f"   • Products created: {created}")
    print(f"   • Products updated: {updated}")
    print(f"   • Total products: {len(prices)}")
    
    print(f"\n🎯 Next steps:")
    print(f"   1. Start the trading server")
    print(f"   2. Monitor price movements through ticker service")
    print(f"   3. Adjust prices if needed using admin tools")
    
    print(f"\n🚀 Ready for intergalactic avocado trading!")

if __name__ == "__main__":
    main()