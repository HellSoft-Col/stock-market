#!/usr/bin/env python3
"""
Final Trading Test - With All Corrections
========================================

Test with corrected message formats including clOrdID for orders.
"""

import asyncio
import json
import random
import time
import logging
import sys
import uuid
from datetime import datetime, timedelta

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler(f'final_test_{datetime.now().strftime("%Y%m%d_%H%M%S")}.log'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)

class FinalTestClient:
    """Test client with all corrections including clOrdID"""
    
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
        self.successes = []
        self.message_queue = asyncio.Queue()
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
    
    async def message_listener(self):
        """Separate coroutine to listen for messages"""
        while self.running and self.websocket:
            try:
                message = await self.websocket.recv()
                data = json.loads(message)
                await self.message_queue.put(data)
                
            except Exception as e:
                if self.running:
                    self.logger.error(f"Message listener error: {e}")
                break
    
    async def wait_for_response(self, expected_types: list, timeout: float = 5.0) -> dict:
        """Wait for a specific type of response"""
        start_time = time.time()
        
        while time.time() - start_time < timeout:
            try:
                message = await asyncio.wait_for(self.message_queue.get(), timeout=1.0)
                
                if message.get("type") in expected_types:
                    return message
                elif message.get("type") == "TICKER":
                    self.logger.debug(f"Ignoring ticker: {message.get('product', 'unknown')}")
                    continue
                else:
                    return message
                    
            except asyncio.TimeoutError:
                continue
        
        return {"type": "TIMEOUT", "message": "No response received"}
    
    def generate_order_id(self) -> str:
        """Generate unique order ID"""
        timestamp = int(time.time() * 1000)
        return f"ORD-{self.token[:6]}-{timestamp}"
    
    async def send_corrected_order(self, side: str, product: str, quantity: int, price: float):
        """Send order with clOrdID included"""
        order_id = self.generate_order_id()
        
        order_message = {
            "type": "ORDER",
            "clOrdID": order_id,
            "side": side,
            "product": product,
            "quantity": quantity,
            "price": round(price, 2),
            "mode": "LIMIT"
        }
        
        self.logger.info(f"Sending corrected order: {side} {quantity} {product} @ ${price} (ID: {order_id})")
        await self.websocket.send(json.dumps(order_message))
        self.orders_sent += 1
        
        response = await self.wait_for_response(["ORDER_SUCCESS", "ORDER_OK", "FILL", "ERROR"])
        
        if response.get("type") in ["ORDER_SUCCESS", "ORDER_OK", "FILL"]:
            self.orders_success += 1
            self.successes.append(f"Order successful: {response}")
            self.logger.info(f"✓ Corrected order successful: {response}")
            return True
        else:
            self.errors.append(f"Corrected order failed: {response}")
            self.logger.warning(f"✗ Corrected order failed: {response}")
            return False
    
    async def send_alternative_order(self, side: str, product: str, quantity: int, price: float):
        """Send order with nested data structure + clOrdID"""
        order_id = self.generate_order_id()
        
        order_message = {
            "type": "ORDER",
            "data": {
                "clOrdID": order_id,
                "side": side,
                "product": product,
                "quantity": quantity,
                "price": round(price, 2),
                "mode": "LIMIT"
            }
        }
        
        self.logger.info(f"Sending alternative order: {side} {quantity} {product} @ ${price} (ID: {order_id})")
        await self.websocket.send(json.dumps(order_message))
        self.orders_sent += 1
        
        response = await self.wait_for_response(["ORDER_SUCCESS", "ORDER_OK", "FILL", "ERROR"])
        
        if response.get("type") in ["ORDER_SUCCESS", "ORDER_OK", "FILL"]:
            self.orders_success += 1
            self.successes.append(f"Alternative order successful: {response}")
            self.logger.info(f"✓ Alternative order successful: {response}")
            return True
        else:
            self.errors.append(f"Alternative order failed: {response}")
            self.logger.warning(f"✗ Alternative order failed: {response}")
            return False
    
    async def try_ping(self):
        """Test basic ping to see if it works"""
        ping_message = {"type": "PING"}
        
        self.logger.info("Sending PING")
        await self.websocket.send(json.dumps(ping_message))
        
        response = await self.wait_for_response(["PONG", "PING_OK"])
        
        if response.get("type") in ["PONG", "PING_OK"]:
            self.successes.append(f"Ping successful: {response}")
            self.logger.info(f"✓ Ping successful: {response}")
            return True
        else:
            self.errors.append(f"Ping failed: {response}")
            self.logger.warning(f"✗ Ping failed: {response}")
            return False
    
    async def comprehensive_test(self, duration_seconds: int):
        """Run comprehensive test with all formats"""
        self.running = True
        
        # Start message listener
        listener_task = asyncio.create_task(self.message_listener())
        
        self.logger.info(f"Starting {duration_seconds}-second comprehensive test")
        
        try:
            # Test 1: Basic ping
            await self.try_ping()
            await asyncio.sleep(2)
            
            # Test 2: Corrected order format
            await self.send_corrected_order("BUY", "FOSFO", 1, 10.00)
            await asyncio.sleep(3)
            
            # Test 3: Alternative order format
            await self.send_alternative_order("SELL", "PITA", 1, 18.00)
            await asyncio.sleep(3)
            
            # Test 4: More corrected orders
            await self.send_corrected_order("BUY", "GUACA", 1, 35.00)
            await asyncio.sleep(3)
            
            # Test 5: Market order
            order_id = self.generate_order_id()
            market_order = {
                "type": "ORDER",
                "clOrdID": order_id,
                "side": "SELL",
                "product": "FOSFO",
                "quantity": 1,
                "mode": "MARKET"  # No price for market orders
            }
            
            self.logger.info(f"Sending market order: SELL 1 FOSFO (ID: {order_id})")
            await self.websocket.send(json.dumps(market_order))
            self.orders_sent += 1
            
            response = await self.wait_for_response(["ORDER_SUCCESS", "ORDER_OK", "FILL", "ERROR"])
            
            if response.get("type") in ["ORDER_SUCCESS", "ORDER_OK", "FILL"]:
                self.orders_success += 1
                self.successes.append(f"Market order successful: {response}")
                self.logger.info(f"✓ Market order successful: {response}")
            else:
                self.errors.append(f"Market order failed: {response}")
                self.logger.warning(f"✗ Market order failed: {response}")
            
        except Exception as e:
            self.logger.error(f"Error in comprehensive test: {e}")
        finally:
            self.running = False
            listener_task.cancel()
        
        self.logger.info("Comprehensive test completed")
    
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
            'errors': self.errors,
            'successes': self.successes
        }

async def run_comprehensive_test(tokens: list, server_url: str, duration: int = 30):
    """Run comprehensive test with corrected formats"""
    logger.info(f"Starting {duration}-second comprehensive test with {len(tokens)} clients")
    
    clients = []
    
    # Connect clients
    for i, token in enumerate(tokens):
        client = FinalTestClient(token, server_url)
        if await client.connect():
            clients.append(client)
            logger.info(f"Client {i+1} connected successfully")
            await asyncio.sleep(1)
        else:
            logger.error(f"Failed to connect client {token[:8]}")
    
    if not clients:
        logger.error("No clients connected - test failed")
        return
    
    logger.info(f"Successfully connected {len(clients)} clients")
    
    # Run tests for each client
    for i, client in enumerate(clients):
        logger.info(f"Testing client {i+1} with corrected formats...")
        await client.comprehensive_test(duration // len(clients))
        await asyncio.sleep(2)
    
    # Generate final analysis
    logger.info("="*60)
    logger.info("COMPREHENSIVE TEST FINAL REPORT")
    logger.info("="*60)
    
    total_orders_sent = sum(client.orders_sent for client in clients)
    total_orders_success = sum(client.orders_success for client in clients)
    all_errors = []
    all_successes = []
    
    for client in clients:
        all_errors.extend(client.errors)
        all_successes.extend(client.successes)
    
    logger.info(f"Connected Clients: {len(clients)}")
    logger.info(f"Orders Sent: {total_orders_sent}")
    logger.info(f"Orders Successful: {total_orders_success}")
    logger.info(f"Order Success Rate: {total_orders_success/total_orders_sent*100:.1f}%" if total_orders_sent > 0 else "N/A")
    logger.info(f"Total Errors: {len(all_errors)}")
    logger.info(f"Total Successes: {len(all_successes)}")
    
    if all_successes:
        logger.info("\n✅ SUCCESSFUL OPERATIONS:")
        logger.info("-" * 35)
        for success in all_successes[:3]:
            logger.info(f"  ✓ {success}")
    
    if all_errors:
        logger.info("\n❌ REMAINING ERRORS:")
        logger.info("-" * 25)
        for error in all_errors[:3]:
            logger.info(f"  ✗ {error}")
    
    logger.info("\nPER-CLIENT RESULTS:")
    logger.info("-" * 25)
    for i, client in enumerate(clients):
        stats = client.get_stats()
        logger.info(f"Client {i+1} ({client.token[:8]}):")
        logger.info(f"  Orders: {stats['orders_success']}/{stats['orders_sent']}")
        logger.info(f"  Successes: {len(stats['successes'])}")
        logger.info(f"  Errors: {len(stats['errors'])}")
    
    # Disconnect all clients
    for client in clients:
        await client.disconnect()
    
    logger.info("="*60)
    
    # Final recommendation
    if total_orders_success > 0:
        logger.info("🎉 SUCCESS: Orders are working with clOrdID!")
        logger.info("✅ READY FOR FULL 15-MINUTE SIMULATION")
        return True
    else:
        logger.warning("⚠ ISSUE: Still need to debug order format")
        return False

async def main():
    """Main entry point"""
    tokens = [
        "TK-09jKZrvn0NF11v99j10vT4Fx",  # Alquimistas de Palta
        "TK-NVUoEHwzH1BRcgcyyDdhx2a4",  # Arpistas de Pita-Pita
    ]
    
    server_url = "wss://trading.hellsoft.tech/ws"
    
    logger.info("Starting Comprehensive Final Test")
    logger.info(f"Server: {server_url}")
    logger.info(f"Testing with {len(tokens)} clients")
    
    try:
        success = await run_comprehensive_test(tokens, server_url, duration=30)
        
        if success:
            logger.info("\n🚀 READY TO RUN 15-MINUTE SIMULATION!")
        else:
            logger.info("\n🔧 NEED MORE DEBUGGING BEFORE FULL SIMULATION")
            
    except KeyboardInterrupt:
        logger.info("Test interrupted by user")
    except Exception as e:
        logger.error(f"Test failed: {e}")

if __name__ == "__main__":
    asyncio.run(main())