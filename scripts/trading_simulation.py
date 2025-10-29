#!/usr/bin/env python3
"""
Multi-Client Trading Simulation Script
=====================================

This script simulates multiple trading clients connecting to the server
with different tokens, performing burst production, and executing buy/sell
orders to test server performance and order validation.

Features:
- Multiple concurrent WebSocket connections
- Burst production simulation
- Realistic buy/sell order generation
- Order validation and timing analysis
- Comprehensive logging and monitoring
- 15-minute automated trading session

Usage:
    python3 trading_simulation.py --config simulation_config.json
    python3 trading_simulation.py --tokens TK-1001,TK-1002,TK-1003
"""

import asyncio
import json
import random
import time
import logging
import argparse
import sys
from datetime import datetime, timedelta
from typing import List, Dict, Any, Optional
from dataclasses import dataclass, field
import statistics

try:
    import websockets
except ImportError:
    print("websockets library not found. Install with: pip install websockets")
    sys.exit(1)

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler(f'trading_simulation_{datetime.now().strftime("%Y%m%d_%H%M%S")}.log'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)

@dataclass
class OrderStats:
    """Statistics for order execution"""
    total_orders: int = 0
    successful_orders: int = 0
    failed_orders: int = 0
    avg_response_time: float = 0.0
    first_fills: int = 0
    order_times: List[float] = field(default_factory=list)

@dataclass
class ClientState:
    """State tracking for each trading client"""
    token: str
    balance: float = 10000.0
    inventory: Dict[str, int] = field(default_factory=lambda: {
        'FOSFO': 50, 'PITA': 30, 'PALTA-OIL': 20,
        'GUACA': 15, 'SEBO': 40, 'H-GUACA': 10
    })
    orders_placed: int = 0
    orders_filled: int = 0
    production_count: int = 0
    last_production: Optional[datetime] = None
    connection_time: Optional[datetime] = None
    is_connected: bool = False
    stats: OrderStats = field(default_factory=OrderStats)

class TradingClient:
    """Individual trading client with WebSocket connection"""
    
    PRODUCTS = ['FOSFO', 'PITA', 'PALTA-OIL', 'GUACA', 'SEBO', 'H-GUACA']
    PRICE_RANGES = {
        'FOSFO': (8.0, 15.0),
        'PITA': (12.0, 22.0),
        'PALTA-OIL': (20.0, 35.0),
        'GUACA': (28.0, 45.0),
        'SEBO': (5.0, 12.0),
        'H-GUACA': (40.0, 60.0)
    }
    
    def __init__(self, token: str, server_url: str = "ws://localhost:8080"):
        self.token = token
        self.server_url = server_url
        self.state = ClientState(token=token)
        self.websocket = None
        self.running = False
        self.message_queue = asyncio.Queue()
        self.logger = logging.getLogger(f"Client-{token}")
        
    async def connect(self) -> bool:
        """Connect to the trading server"""
        try:
            self.logger.info(f"Connecting to {self.server_url}")
            self.websocket = await websockets.connect(self.server_url)
            self.state.connection_time = datetime.now()
            
            # Authenticate
            auth_message = {
                "type": "AUTH",
                "data": {"token": self.token}
            }
            await self.websocket.send(json.dumps(auth_message))
            
            # Wait for authentication response
            response = await asyncio.wait_for(self.websocket.recv(), timeout=5.0)
            auth_response = json.loads(response)
            
            if auth_response.get("type") == "AUTH_SUCCESS":
                self.state.is_connected = True
                self.logger.info(f"Successfully authenticated with token {self.token}")
                return True
            else:
                self.logger.error(f"Authentication failed: {auth_response}")
                return False
                
        except Exception as e:
            self.logger.error(f"Connection failed: {e}")
            return False
    
    async def disconnect(self):
        """Disconnect from the server"""
        self.running = False
        if self.websocket:
            await self.websocket.close()
            self.state.is_connected = False
            self.logger.info("Disconnected from server")
    
    async def send_message(self, message: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """Send message to server and measure response time"""
        if not self.websocket:
            return None
            
        try:
            start_time = time.time()
            await self.websocket.send(json.dumps(message))
            
            # Wait for response with timeout
            response_raw = await asyncio.wait_for(self.websocket.recv(), timeout=10.0)
            response_time = time.time() - start_time
            
            response = json.loads(response_raw)
            self.state.stats.order_times.append(response_time)
            
            self.logger.debug(f"Message sent: {message['type']}, Response time: {response_time:.3f}s")
            return response
            
        except asyncio.TimeoutError:
            self.logger.warning(f"Timeout waiting for response to {message['type']}")
            return None
        except Exception as e:
            self.logger.error(f"Error sending message: {e}")
            return None
    
    async def place_order(self, side: str, product: str, quantity: int, 
                         price: float, mode: str = "LIMIT") -> bool:
        """Place a buy or sell order"""
        order_message = {
            "type": "ORDER",
            "data": {
                "side": side,
                "product": product,
                "quantity": quantity,
                "price": round(price, 2),
                "mode": mode,
                "message": f"Simulation order from {self.token}"
            }
        }
        
        response = await self.send_message(order_message)
        self.state.stats.total_orders += 1
        
        if response and response.get("type") == "ORDER_SUCCESS":
            self.state.stats.successful_orders += 1
            self.state.orders_placed += 1
            self.logger.info(f"Order placed: {side} {quantity} {product} @ ${price}")
            return True
        else:
            self.state.stats.failed_orders += 1
            self.logger.warning(f"Order failed: {response}")
            return False
    
    async def simulate_production(self, product: str, quantity: int) -> bool:
        """Simulate production of a product"""
        production_message = {
            "type": "PRODUCTION",
            "data": {
                "product": product,
                "quantity": quantity
            }
        }
        
        response = await self.send_message(production_message)
        
        if response and response.get("type") == "PRODUCTION_SUCCESS":
            self.state.production_count += 1
            self.state.last_production = datetime.now()
            self.logger.info(f"Production completed: {quantity} {product}")
            
            # Update inventory
            if product in self.state.inventory:
                self.state.inventory[product] += quantity
            else:
                self.state.inventory[product] = quantity
            return True
        else:
            self.logger.warning(f"Production failed: {response}")
            return False
    
    def generate_realistic_price(self, product: str, market_trend: float = 0.0) -> float:
        """Generate realistic price based on product and market conditions"""
        min_price, max_price = self.PRICE_RANGES[product]
        base_price = random.uniform(min_price, max_price)
        
        # Apply market trend
        trend_factor = 1.0 + (market_trend * 0.1)
        
        # Add some randomness
        volatility = random.uniform(0.95, 1.05)
        
        return round(base_price * trend_factor * volatility, 2)
    
    async def burst_trading_session(self, duration_minutes: int = 5):
        """Perform burst trading for specified duration"""
        end_time = datetime.now() + timedelta(minutes=duration_minutes)
        burst_count = 0
        
        self.logger.info(f"Starting {duration_minutes}-minute burst trading session")
        
        while datetime.now() < end_time and self.running:
            try:
                # Random action selection
                action = random.choices(
                    ['buy', 'sell', 'production', 'wait'],
                    weights=[30, 30, 20, 20]
                )[0]
                
                if action == 'production':
                    # Burst production
                    product = random.choice(self.PRODUCTS[:3])  # Focus on basic products
                    quantity = random.randint(5, 20)
                    await self.simulate_production(product, quantity)
                    
                elif action in ['buy', 'sell']:
                    # Generate realistic trading
                    product = random.choice(self.PRODUCTS)
                    quantity = random.randint(1, 10)
                    
                    # Market trend simulation
                    market_trend = random.uniform(-0.5, 0.5)
                    price = self.generate_realistic_price(product, market_trend)
                    
                    # Slight price adjustment for competitive orders
                    if action == 'buy':
                        price *= random.uniform(0.98, 1.02)  # Slightly competitive buy
                    else:
                        price *= random.uniform(0.98, 1.02)  # Slightly competitive sell
                    
                    await self.place_order(action.upper(), product, quantity, price)
                    burst_count += 1
                
                # Variable delay for realistic simulation
                await asyncio.sleep(random.uniform(0.5, 3.0))
                
            except Exception as e:
                self.logger.error(f"Error in burst session: {e}")
                await asyncio.sleep(1.0)
        
        self.logger.info(f"Burst trading session completed. Actions performed: {burst_count}")
    
    async def listen_for_messages(self):
        """Listen for incoming messages from server"""
        while self.running and self.websocket:
            try:
                message = await asyncio.wait_for(self.websocket.recv(), timeout=1.0)
                data = json.loads(message)
                
                # Handle different message types
                if data.get("type") == "FILL":
                    self.state.orders_filled += 1
                    self.state.stats.first_fills += 1
                    self.logger.info(f"Order filled: {data}")
                elif data.get("type") == "TICKER":
                    self.logger.debug(f"Market update: {data}")
                elif data.get("type") == "ERROR":
                    self.logger.warning(f"Server error: {data}")
                    
            except asyncio.TimeoutError:
                continue
            except Exception as e:
                self.logger.error(f"Error listening for messages: {e}")
                break

class TradingSimulation:
    """Main simulation coordinator"""
    
    def __init__(self, tokens: List[str], server_url: str = "ws://localhost:8080"):
        self.tokens = tokens
        self.server_url = server_url
        self.clients: List[TradingClient] = []
        self.start_time = None
        self.end_time = None
        self.logger = logging.getLogger("Simulation")
        
    async def initialize_clients(self) -> bool:
        """Initialize and connect all clients"""
        self.logger.info(f"Initializing {len(self.tokens)} trading clients")
        
        for token in self.tokens:
            client = TradingClient(token, self.server_url)
            success = await client.connect()
            
            if success:
                self.clients.append(client)
                self.logger.info(f"Client {token} connected successfully")
            else:
                self.logger.error(f"Failed to connect client {token}")
        
        connected_count = len(self.clients)
        self.logger.info(f"Successfully connected {connected_count}/{len(self.tokens)} clients")
        return connected_count > 0
    
    async def run_simulation(self, duration_minutes: int = 15):
        """Run the complete 15-minute trading simulation"""
        self.start_time = datetime.now()
        self.end_time = self.start_time + timedelta(minutes=duration_minutes)
        
        self.logger.info(f"Starting {duration_minutes}-minute trading simulation with {len(self.clients)} clients")
        self.logger.info(f"Simulation will run from {self.start_time} to {self.end_time}")
        
        # Start all clients
        for client in self.clients:
            client.running = True
        
        # Phase 1: Initial burst production (2 minutes)
        self.logger.info("=== PHASE 1: BURST PRODUCTION ===")
        await self.run_burst_production_phase(2)
        
        # Phase 2: Mixed trading activity (10 minutes)
        self.logger.info("=== PHASE 2: MIXED TRADING ===")
        await self.run_mixed_trading_phase(10)
        
        # Phase 3: Final competitive trading (3 minutes)
        self.logger.info("=== PHASE 3: COMPETITIVE TRADING ===")
        await self.run_competitive_trading_phase(3)
        
        # Stop all clients
        await self.shutdown_clients()
        
        # Generate final report
        self.generate_final_report()
    
    async def run_burst_production_phase(self, duration_minutes: int):
        """Phase 1: Burst production to build inventory"""
        tasks = []
        for client in self.clients:
            task = asyncio.create_task(self.client_burst_production(client, duration_minutes))
            tasks.append(task)
        
        # Also start message listeners
        for client in self.clients:
            task = asyncio.create_task(client.listen_for_messages())
            tasks.append(task)
        
        await asyncio.sleep(duration_minutes * 60)
        
        # Cancel tasks
        for task in tasks:
            task.cancel()
        
        await asyncio.gather(*tasks, return_exceptions=True)
    
    async def client_burst_production(self, client: TradingClient, duration_minutes: int):
        """Individual client burst production"""
        end_time = datetime.now() + timedelta(minutes=duration_minutes)
        
        while datetime.now() < end_time and client.running:
            try:
                # Focus on basic products for production
                product = random.choice(['FOSFO', 'PITA', 'PALTA-OIL'])
                quantity = random.randint(10, 30)
                
                await client.simulate_production(product, quantity)
                await asyncio.sleep(random.uniform(5, 15))
                
            except Exception as e:
                client.logger.error(f"Error in burst production: {e}")
                await asyncio.sleep(1)
    
    async def run_mixed_trading_phase(self, duration_minutes: int):
        """Phase 2: Mixed production and trading"""
        tasks = []
        for client in self.clients:
            task = asyncio.create_task(client.burst_trading_session(duration_minutes))
            tasks.append(task)
        
        await asyncio.gather(*tasks, return_exceptions=True)
    
    async def run_competitive_trading_phase(self, duration_minutes: int):
        """Phase 3: Competitive trading with aggressive pricing"""
        end_time = datetime.now() + timedelta(minutes=duration_minutes)
        
        self.logger.info("Starting competitive trading phase - testing order priority")
        
        # Create competitive scenarios
        tasks = []
        for i, client in enumerate(self.clients):
            task = asyncio.create_task(
                self.competitive_client_trading(client, duration_minutes, i)
            )
            tasks.append(task)
        
        await asyncio.gather(*tasks, return_exceptions=True)
    
    async def competitive_client_trading(self, client: TradingClient, duration_minutes: int, client_index: int):
        """Individual client competitive trading"""
        end_time = datetime.now() + timedelta(minutes=duration_minutes)
        
        while datetime.now() < end_time and client.running:
            try:
                # Create competitive scenarios
                product = random.choice(client.PRODUCTS)
                quantity = random.randint(1, 5)
                
                # Slightly different pricing strategies per client
                base_price = client.generate_realistic_price(product)
                
                if client_index % 2 == 0:
                    # Aggressive buyer
                    price = base_price * random.uniform(1.01, 1.05)
                    await client.place_order("BUY", product, quantity, price)
                else:
                    # Aggressive seller
                    price = base_price * random.uniform(0.95, 0.99)
                    await client.place_order("SELL", product, quantity, price)
                
                await asyncio.sleep(random.uniform(2, 8))
                
            except Exception as e:
                client.logger.error(f"Error in competitive trading: {e}")
                await asyncio.sleep(1)
    
    async def shutdown_clients(self):
        """Gracefully shutdown all clients"""
        self.logger.info("Shutting down all clients...")
        
        tasks = []
        for client in self.clients:
            task = asyncio.create_task(client.disconnect())
            tasks.append(task)
        
        await asyncio.gather(*tasks, return_exceptions=True)
        self.logger.info("All clients disconnected")
    
    def generate_final_report(self):
        """Generate comprehensive simulation report"""
        self.logger.info("="*60)
        self.logger.info("TRADING SIMULATION FINAL REPORT")
        self.logger.info("="*60)
        
        total_orders = sum(client.state.stats.total_orders for client in self.clients)
        total_successful = sum(client.state.stats.successful_orders for client in self.clients)
        total_failed = sum(client.state.stats.failed_orders for client in self.clients)
        total_production = sum(client.state.production_count for client in self.clients)
        total_fills = sum(client.state.orders_filled for client in self.clients)
        
        # Calculate average response times
        all_response_times = []
        for client in self.clients:
            all_response_times.extend(client.state.stats.order_times)
        
        avg_response_time = statistics.mean(all_response_times) if all_response_times else 0
        
        duration = self.end_time - self.start_time if self.end_time and self.start_time else timedelta(0)
        self.logger.info(f"Simulation Duration: {duration}")
        self.logger.info(f"Connected Clients: {len(self.clients)}")
        self.logger.info(f"Total Orders Placed: {total_orders}")
        self.logger.info(f"Successful Orders: {total_successful} ({total_successful/total_orders*100:.1f}%)")
        self.logger.info(f"Failed Orders: {total_failed} ({total_failed/total_orders*100:.1f}%)")
        self.logger.info(f"Total Productions: {total_production}")
        self.logger.info(f"Total Fills: {total_fills}")
        self.logger.info(f"Average Response Time: {avg_response_time:.3f}s")
        
        if all_response_times:
            self.logger.info(f"Min Response Time: {min(all_response_times):.3f}s")
            self.logger.info(f"Max Response Time: {max(all_response_times):.3f}s")
        
        self.logger.info("\nPER-CLIENT STATISTICS:")
        self.logger.info("-" * 40)
        
        for client in self.clients:
            stats = client.state.stats
            self.logger.info(f"Client {client.token}:")
            self.logger.info(f"  Orders: {stats.successful_orders}/{stats.total_orders}")
            self.logger.info(f"  Productions: {client.state.production_count}")
            self.logger.info(f"  Fills: {client.state.orders_filled}")
            if stats.order_times:
                avg_time = statistics.mean(stats.order_times)
                self.logger.info(f"  Avg Response: {avg_time:.3f}s")
        
        self.logger.info("="*60)

async def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(description='Trading Server Simulation Script')
    parser.add_argument('--tokens', type=str, required=True,
                       help='Comma-separated list of team tokens (e.g., TK-1001,TK-1002,TK-1003)')
    parser.add_argument('--server', type=str, default='ws://localhost:8080',
                       help='WebSocket server URL')
    parser.add_argument('--duration', type=int, default=15,
                       help='Simulation duration in minutes')
    parser.add_argument('--verbose', '-v', action='store_true',
                       help='Enable verbose logging')
    
    args = parser.parse_args()
    
    if args.verbose:
        logging.getLogger().setLevel(logging.DEBUG)
    
    # Parse tokens
    tokens = [token.strip() for token in args.tokens.split(',')]
    
    if len(tokens) < 2:
        logger.error("At least 2 tokens are required for meaningful simulation")
        return 1
    
    logger.info(f"Starting simulation with tokens: {tokens}")
    logger.info(f"Server: {args.server}")
    logger.info(f"Duration: {args.duration} minutes")
    
    # Create and run simulation
    simulation = TradingSimulation(tokens, args.server)
    
    try:
        # Initialize clients
        if not await simulation.initialize_clients():
            logger.error("Failed to initialize clients")
            return 1
        
        # Run simulation
        await simulation.run_simulation(args.duration)
        
        logger.info("Simulation completed successfully!")
        return 0
        
    except KeyboardInterrupt:
        logger.info("Simulation interrupted by user")
        await simulation.shutdown_clients()
        return 1
    except Exception as e:
        logger.error(f"Simulation failed: {e}")
        await simulation.shutdown_clients()
        return 1

if __name__ == "__main__":
    exit_code = asyncio.run(main())
    sys.exit(exit_code)