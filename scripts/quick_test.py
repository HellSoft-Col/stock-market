#!/usr/bin/env python3
"""
Quick Trading Test - 20 Second Analysis
======================================

Short test to identify and fix issues before running full simulation.
"""

import asyncio
import json
import random
import time
import logging
import sys
from datetime import datetime, timedelta

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler(f'quick_test_{datetime.now().strftime("%Y%m%d_%H%M%S")}.log'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)

class QuickTestClient:
    """Simple test client with proper error handling"""
    
    def __init__(self, token: str, server_url: str):
        self.token = token
        self.server_url = server_url
        self.websocket = None
        self.running = False
        self.orders_sent = 0
        self.orders_success = 0
        self.productions_sent = 0
        self.productions_success = 0
        self.errors = []
        self.logger = logging.getLogger(f"Client-{token[:8]}")
        
    async def connect(self) -> bool:
        """Connect and authenticate"""
        try:
            websockets_module = __import__('websockets')
            
            self.logger.info(f"Connecting to {self.server_url}")
            self.websocket = await websockets_module.connect(self.server_url)
            
            # Authenticate
            auth_message = {
                "type": "LOGIN",
                "token": self.token
            }
            await self.websocket.send(json.dumps(auth_message))
            
            # Wait for auth response
            response = await asyncio.wait_for(self.websocket.recv(), timeout=10.0)
            auth_response = json.loads(response)
            
            if auth_response.get("type") == "LOGIN_OK":
                team_name = auth_response.get('team', 'Unknown')
                self.logger.info(f"✓ Authenticated as: {team_name}")
                return True
            else:
                self.logger.error(f"✗ Auth failed: {auth_response}")
                return False
                
        except Exception as e:
            self.logger.error(f"✗ Connection failed: {e}")
            return False
    
    async def send_order(self, side: str, product: str, quantity: int, price: float):
        """Send a single order with proper error handling"""
        try:
            # Try different order formats to see which one works
            order_message = {
                "type": "ORDER",
                "side": side,
                "product": product,
                "quantity": quantity,
                "price": round(price, 2),
                "mode": "LIMIT"
            }
            
            self.logger.info(f"Sending order: {side} {quantity} {product} @ ${price}")
            await self.websocket.send(json.dumps(order_message))
            self.orders_sent += 1
            
            # Wait for response (but don't block other operations)
            try:
                response = await asyncio.wait_for(self.websocket.recv(), timeout=5.0)
                response_data = json.loads(response)
                
                if response_data.get("type") in ["ORDER_SUCCESS", "ORDER_OK"]:
                    self.orders_success += 1
                    self.logger.info(f"✓ Order successful: {response_data}")
                else:
                    self.errors.append(f"Order failed: {response_data}")
                    self.logger.warning(f"✗ Order failed: {response_data}")
                    
            except asyncio.TimeoutError:
                self.logger.warning("⚠ Order response timeout")
                
        except Exception as e:
            self.errors.append(f"Order send error: {e}")
            self.logger.error(f"✗ Order send error: {e}")
    
    async def send_production(self, product: str, quantity: int):
        """Send a production request"""
        try:
            # Try different production formats
            production_message = {
                "type": "PRODUCTION",
                "product": product,
                "quantity": quantity
            }
            
            self.logger.info(f"Sending production: {quantity} {product}")
            await self.websocket.send(json.dumps(production_message))
            self.productions_sent += 1
            
            try:
                response = await asyncio.wait_for(self.websocket.recv(), timeout=5.0)
                response_data = json.loads(response)
                
                if response_data.get("type") in ["PRODUCTION_SUCCESS", "PRODUCTION_OK"]:
                    self.productions_success += 1
                    self.logger.info(f"✓ Production successful: {response_data}")
                else:
                    self.errors.append(f"Production failed: {response_data}")
                    self.logger.warning(f"✗ Production failed: {response_data}")
                    
            except asyncio.TimeoutError:
                self.logger.warning("⚠ Production response timeout")
                
        except Exception as e:
            self.errors.append(f"Production send error: {e}")
            self.logger.error(f"✗ Production send error: {e}")
    
    async def test_sequence(self, duration_seconds: int):
        """Run a short test sequence"""
        self.running = True
        end_time = datetime.now() + timedelta(seconds=duration_seconds)
        
        self.logger.info(f"Starting {duration_seconds}-second test sequence")
        
        action_count = 0
        while datetime.now() < end_time and self.running and action_count < 5:
            try:
                if action_count == 0:
                    # First, try a simple production
                    await self.send_production("FOSFO", 5)
                elif action_count == 1:
                    # Then try a buy order
                    await self.send_order("BUY", "FOSFO", 2, 10.50)
                elif action_count == 2:
                    # Try a sell order
                    await self.send_order("SELL", "PITA", 1, 18.00)
                elif action_count == 3:
                    # Another production
                    await self.send_production("PITA", 3)
                elif action_count == 4:
                    # Final order
                    await self.send_order("BUY", "GUACA", 1, 35.00)
                
                action_count += 1
                await asyncio.sleep(3)  # Wait between actions
                
            except Exception as e:
                self.logger.error(f"Error in test sequence: {e}")
                break
        
        self.logger.info(f"Test sequence completed. Actions: {action_count}")
    
    async def disconnect(self):
        """Clean disconnect"""
        self.running = False
        if self.websocket:
            try:
                await self.websocket.close()
                self.logger.info("Disconnected")
            except Exception as e:
                self.logger.error(f"Disconnect error: {e}")
    
    def get_stats(self):
        """Get client statistics"""
        return {
            'orders_sent': self.orders_sent,
            'orders_success': self.orders_success,
            'productions_sent': self.productions_sent,
            'productions_success': self.productions_success,
            'errors': self.errors
        }

async def run_quick_test(tokens: list, server_url: str, duration: int = 20):
    """Run quick test with multiple clients"""
    logger.info(f"Starting {duration}-second quick test with {len(tokens)} clients")
    
    clients = []
    
    # Connect all clients
    for token in tokens:
        client = QuickTestClient(token, server_url)
        if await client.connect():
            clients.append(client)
        else:
            logger.error(f"Failed to connect client {token[:8]}")
    
    if not clients:
        logger.error("No clients connected - test failed")
        return
    
    logger.info(f"Successfully connected {len(clients)} clients")
    
    # Run test sequences in parallel
    tasks = []
    for client in clients:
        task = asyncio.create_task(client.test_sequence(duration))
        tasks.append(task)
    
    # Wait for all tests to complete
    await asyncio.gather(*tasks, return_exceptions=True)
    
    # Generate analysis report
    logger.info("="*50)
    logger.info("QUICK TEST ANALYSIS REPORT")
    logger.info("="*50)
    
    total_orders_sent = sum(client.orders_sent for client in clients)
    total_orders_success = sum(client.orders_success for client in clients)
    total_productions_sent = sum(client.productions_sent for client in clients)
    total_productions_success = sum(client.productions_success for client in clients)
    all_errors = []
    
    for client in clients:
        all_errors.extend(client.errors)
    
    logger.info(f"Connected Clients: {len(clients)}")
    logger.info(f"Orders Sent: {total_orders_sent}")
    logger.info(f"Orders Successful: {total_orders_success}")
    logger.info(f"Order Success Rate: {total_orders_success/total_orders_sent*100:.1f}%" if total_orders_sent > 0 else "N/A")
    logger.info(f"Productions Sent: {total_productions_sent}")
    logger.info(f"Productions Successful: {total_productions_success}")
    logger.info(f"Production Success Rate: {total_productions_success/total_productions_sent*100:.1f}%" if total_productions_sent > 0 else "N/A")
    logger.info(f"Total Errors: {len(all_errors)}")
    
    if all_errors:
        logger.info("\nERROR ANALYSIS:")
        logger.info("-" * 30)
        error_types = {}
        for error in all_errors:
            error_key = error.split(':')[0] if ':' in error else error
            error_types[error_key] = error_types.get(error_key, 0) + 1
        
        for error_type, count in error_types.items():
            logger.info(f"{error_type}: {count} occurrences")
        
        logger.info("\nSample Errors:")
        for error in all_errors[:5]:  # Show first 5 errors
            logger.info(f"  - {error}")
    
    logger.info("\nPER-CLIENT STATS:")
    logger.info("-" * 20)
    for i, client in enumerate(clients):
        stats = client.get_stats()
        logger.info(f"Client {i+1} ({client.token[:8]}):")
        logger.info(f"  Orders: {stats['orders_success']}/{stats['orders_sent']}")
        logger.info(f"  Productions: {stats['productions_success']}/{stats['productions_sent']}")
        logger.info(f"  Errors: {len(stats['errors'])}")
    
    # Disconnect all clients
    for client in clients:
        await client.disconnect()
    
    logger.info("="*50)
    
    # Recommendations based on results
    if total_orders_sent == 0 and total_productions_sent == 0:
        logger.warning("⚠ ISSUE: No messages were sent - check connection/authentication")
    elif total_orders_success == 0 and total_productions_success == 0:
        logger.warning("⚠ ISSUE: All messages failed - check message format")
    elif len(all_errors) > len(clients) * 2:
        logger.warning("⚠ ISSUE: High error rate - need to fix message formats")
    else:
        logger.info("✓ ANALYSIS: Test looks good, ready for longer simulation")

async def main():
    """Main entry point for quick test"""
    # Use the tokens from the previous simulation
    tokens = [
        "TK-09jKZrvn0NF11v99j10vT4Fx",  # Alquimistas de Palta
        "TK-NVUoEHwzH1BRcgcyyDdhx2a4",  # Arpistas de Pita-Pita
        "TK-XqnoG2blE3DFmApa75iexwvC",  # Avocultores del Hueso Cósmico
    ]
    
    server_url = "wss://trading.hellsoft.tech/ws"
    
    logger.info("Starting Quick Test Analysis")
    logger.info(f"Server: {server_url}")
    logger.info(f"Tokens: {len(tokens)} clients")
    
    try:
        await run_quick_test(tokens, server_url, duration=20)
    except KeyboardInterrupt:
        logger.info("Test interrupted by user")
    except Exception as e:
        logger.error(f"Test failed: {e}")

if __name__ == "__main__":
    asyncio.run(main())