#!/usr/bin/env python3
"""
Simplified Trading Simulation Script
===================================

A simplified version of the trading simulation that handles import issues
and focuses on core functionality.
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

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler(f'simulation_{datetime.now().strftime("%Y%m%d_%H%M%S")}.log'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)

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
    
    def __init__(self, token: str, server_url: str):
        self.token = token
        self.server_url = server_url
        self.websocket = None
        self.running = False
        self.orders_placed = 0
        self.orders_filled = 0
        self.productions = 0
        self.response_times = []
        self.logger = logging.getLogger(f"Client-{token[:8]}")
        
    async def connect(self) -> bool:
        """Connect to the trading server"""
        try:
            # Dynamic import to avoid static analysis issues
            websockets_module = __import__('websockets')
            
            self.logger.info(f"Connecting to {self.server_url}")
            self.websocket = await websockets_module.connect(self.server_url)
            
            # Authenticate
            auth_message = {
                "type": "LOGIN",
                "token": self.token
            }
            await self.websocket.send(json.dumps(auth_message))
            
            # Wait for authentication response
            response = await asyncio.wait_for(self.websocket.recv(), timeout=10.0)
            auth_response = json.loads(response)
            
            if auth_response.get("type") == "LOGIN_OK":
                self.logger.info(f"Successfully authenticated: {auth_response.get('team')}")
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
            self.logger.info("Disconnected from server")
    
    async def send_message(self, message: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """Send message to server and measure response time"""
        if not self.websocket:
            return None
            
        try:
            start_time = time.time()
            await self.websocket.send(json.dumps(message))
            
            # Wait for response with timeout
            response_raw = await asyncio.wait_for(self.websocket.recv(), timeout=15.0)
            response_time = time.time() - start_time
            
            response = json.loads(response_raw)
            self.response_times.append(response_time)
            
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
                "message": f"Simulation order from {self.token[:8]}"
            }
        }
        
        response = await self.send_message(order_message)
        self.orders_placed += 1
        
        if response and response.get("type") in ["ORDER_SUCCESS", "ORDER_OK"]:
            self.logger.info(f"Order placed: {side} {quantity} {product} @ ${price}")
            return True
        else:
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
        
        if response and response.get("type") in ["PRODUCTION_SUCCESS", "PRODUCTION_OK"]:
            self.productions += 1
            self.logger.info(f"Production completed: {quantity} {product}")
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
    
    async def trading_session(self, duration_minutes: int):
        """Perform trading for specified duration"""
        end_time = datetime.now() + timedelta(minutes=duration_minutes)
        action_count = 0
        
        self.logger.info(f"Starting {duration_minutes}-minute trading session")
        self.running = True
        
        while datetime.now() < end_time and self.running:
            try:
                # Random action selection
                action = random.choices(
                    ['buy', 'sell', 'production', 'wait'],
                    weights=[35, 35, 20, 10]
                )[0]
                
                if action == 'production':
                    # Production simulation
                    product = random.choice(self.PRODUCTS[:3])  # Focus on basic products
                    quantity = random.randint(5, 25)
                    await self.simulate_production(product, quantity)
                    
                elif action in ['buy', 'sell']:
                    # Generate realistic trading
                    product = random.choice(self.PRODUCTS)
                    quantity = random.randint(1, 8)
                    
                    # Market trend simulation
                    market_trend = random.uniform(-0.3, 0.3)
                    price = self.generate_realistic_price(product, market_trend)
                    
                    # Slightly competitive pricing
                    if action == 'buy':
                        price *= random.uniform(0.99, 1.02)
                    else:
                        price *= random.uniform(0.98, 1.01)
                    
                    await self.place_order(action.upper(), product, quantity, price)
                    action_count += 1
                
                # Variable delay for realistic simulation
                await asyncio.sleep(random.uniform(1.0, 4.0))
                
            except Exception as e:
                self.logger.error(f"Error in trading session: {e}")
                await asyncio.sleep(2.0)
        
        self.logger.info(f"Trading session completed. Actions performed: {action_count}")
    
    async def listen_for_messages(self):
        """Listen for incoming messages from server"""
        while self.running and self.websocket:
            try:
                message = await asyncio.wait_for(self.websocket.recv(), timeout=2.0)
                data = json.loads(message)
                
                # Handle different message types
                if data.get("type") == "FILL":
                    self.orders_filled += 1
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
    
    def __init__(self, tokens: List[str], server_url: str):
        self.tokens = tokens
        self.server_url = server_url
        self.clients: List[TradingClient] = []
        self.start_time = None
        self.logger = logging.getLogger("Simulation")
        
    async def initialize_clients(self) -> bool:
        """Initialize and connect all clients"""
        self.logger.info(f"Initializing {len(self.tokens)} trading clients")
        
        for token in self.tokens:
            client = TradingClient(token, self.server_url)
            success = await client.connect()
            
            if success:
                self.clients.append(client)
                self.logger.info(f"Client {token[:8]} connected successfully")
            else:
                self.logger.error(f"Failed to connect client {token[:8]}")
        
        connected_count = len(self.clients)
        self.logger.info(f"Successfully connected {connected_count}/{len(self.tokens)} clients")
        return connected_count > 0
    
    async def run_simulation(self, duration_minutes: int = 15):
        """Run the complete trading simulation"""
        self.start_time = datetime.now()
        end_time = self.start_time + timedelta(minutes=duration_minutes)
        
        self.logger.info(f"Starting {duration_minutes}-minute trading simulation with {len(self.clients)} clients")
        self.logger.info(f"Simulation will run from {self.start_time} to {end_time}")
        
        # Create tasks for all clients
        tasks = []
        
        # Start trading sessions for all clients
        for client in self.clients:
            trading_task = asyncio.create_task(client.trading_session(duration_minutes))
            listen_task = asyncio.create_task(client.listen_for_messages())
            tasks.extend([trading_task, listen_task])
        
        try:
            # Wait for all tasks to complete or timeout
            await asyncio.wait_for(
                asyncio.gather(*tasks, return_exceptions=True),
                timeout=duration_minutes * 60 + 60  # Add 1 minute buffer
            )
        except asyncio.TimeoutError:
            self.logger.warning("Simulation timed out, stopping all clients")
        except Exception as e:
            self.logger.error(f"Simulation error: {e}")
        finally:
            # Stop all clients
            for client in self.clients:
                client.running = False
            
            # Cancel remaining tasks
            for task in tasks:
                if not task.done():
                    task.cancel()
            
            # Wait a bit for graceful shutdown
            await asyncio.sleep(2)
            
            # Disconnect all clients
            await self.shutdown_clients()
            
            # Generate final report
            self.generate_final_report()
    
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
        
        if not self.clients:
            self.logger.error("No clients were connected - simulation failed")
            return
        
        total_orders = sum(client.orders_placed for client in self.clients)
        total_fills = sum(client.orders_filled for client in self.clients)
        total_productions = sum(client.productions for client in self.clients)
        
        # Calculate average response times
        all_response_times = []
        for client in self.clients:
            all_response_times.extend(client.response_times)
        
        avg_response_time = sum(all_response_times) / len(all_response_times) if all_response_times else 0
        
        duration = datetime.now() - self.start_time if self.start_time else timedelta(0)
        
        self.logger.info(f"Simulation Duration: {duration}")
        self.logger.info(f"Connected Clients: {len(self.clients)}")
        self.logger.info(f"Total Orders Placed: {total_orders}")
        self.logger.info(f"Total Orders Filled: {total_fills}")
        self.logger.info(f"Fill Rate: {total_fills/total_orders*100:.1f}%" if total_orders > 0 else "N/A")
        self.logger.info(f"Total Productions: {total_productions}")
        self.logger.info(f"Average Response Time: {avg_response_time:.3f}s")
        
        if all_response_times:
            self.logger.info(f"Min Response Time: {min(all_response_times):.3f}s")
            self.logger.info(f"Max Response Time: {max(all_response_times):.3f}s")
        
        self.logger.info("\nPER-CLIENT STATISTICS:")
        self.logger.info("-" * 40)
        
        for client in self.clients:
            self.logger.info(f"Client {client.token[:8]}:")
            self.logger.info(f"  Orders Placed: {client.orders_placed}")
            self.logger.info(f"  Orders Filled: {client.orders_filled}")
            self.logger.info(f"  Productions: {client.productions}")
            if client.response_times:
                avg_time = sum(client.response_times) / len(client.response_times)
                self.logger.info(f"  Avg Response: {avg_time:.3f}s")
        
        self.logger.info("="*60)

async def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(description='Simplified Trading Simulation')
    parser.add_argument('--tokens', type=str, required=True,
                       help='Comma-separated list of team tokens')
    parser.add_argument('--server', type=str, default='wss://trading.hellsoft.tech/ws',
                       help='WebSocket server URL')
    parser.add_argument('--duration', type=int, default=15,
                       help='Simulation duration in minutes')
    parser.add_argument('--verbose', '-v', action='store_true',
                       help='Enable verbose logging')
    
    args = parser.parse_args()
    
    if args.verbose:
        logging.getLogger().setLevel(logging.DEBUG)
    
    # Check for websockets dependency
    try:
        import websockets
        logger.info("✓ websockets library found")
    except ImportError:
        logger.error("✗ websockets library not found")
        logger.error("Please install with: pip install websockets")
        return 1
    
    # Parse tokens
    tokens = [token.strip() for token in args.tokens.split(',')]
    
    if len(tokens) < 2:
        logger.error("At least 2 tokens are required for meaningful simulation")
        return 1
    
    logger.info(f"Starting simulation with {len(tokens)} clients")
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