#!/usr/bin/env python3
"""
Corrected Trading Test - With Proper Server Message Formats
==========================================================

Test script with corrected message formats based on server implementation analysis.
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
        logging.FileHandler(f'corrected_test_{datetime.now().strftime("%Y%m%d_%H%M%S")}.log'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)

class CorrectedTestClient:
    """Test client with proper server message formats"""
    
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
        self.message_queue = asyncio.Queue()
        self.logger = logging.getLogger(f"Client-{token[:8]}")
        self.team_name = ""
        
    async def connect(self) -> bool:
        """Connect and authenticate"""
        try:
            websockets_module = __import__('websockets')
            
            self.logger.info(f"Connecting to {self.server_url}")
            self.websocket = await websockets_module.connect(self.server_url)
            
            # Authenticate using correct LOGIN format
            auth_message = {
                "type": "LOGIN",
                "token": self.token
            }
            await self.websocket.send(json.dumps(auth_message))
            
            # Wait for auth response
            response = await asyncio.wait_for(self.websocket.recv(), timeout=10.0)
            auth_response = json.loads(response)
            
            if auth_response.get("type") == "LOGIN_OK":
                self.team_name = auth_response.get('team', 'Unknown')
                self.logger.info(f"✓ Authenticated as: {self.team_name}")
                self.logger.info(f"  Species: {auth_response.get('species', 'Unknown')}")
                self.logger.info(f"  Balance: ${auth_response.get('currentBalance', 0)}")
                self.logger.info(f"  Inventory: {auth_response.get('inventory', {})}")
                return True
            else:
                self.logger.error(f"✗ Auth failed: {auth_response}")
                return False
                
        except Exception as e:
            self.logger.error(f"✗ Connection failed: {e}")
            return False
    
    async def message_listener(self):
        """Listen for incoming messages"""
        while self.running and self.websocket:
            try:
                message = await self.websocket.recv()
                data = json.loads(message)
                await self.message_queue.put(data)
                
                # Log important messages
                msg_type = data.get("type", "UNKNOWN")
                if msg_type in ["FILL", "ERROR", "ORDER_ACK"]:
                    self.logger.info(f"Received {msg_type}: {data}")
                elif msg_type == "TICKER":
                    # Log tickers more briefly
                    product = data.get("product", "?")
                    bid = data.get("bestBid", "N/A")
                    ask = data.get("bestAsk", "N/A")
                    self.logger.debug(f"Ticker {product}: Bid={bid}, Ask={ask}")
                
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
                
                msg_type = message.get("type")
                if msg_type in expected_types:
                    return message
                elif msg_type == "ERROR":
                    return message
                elif msg_type in ["TICKER", "INVENTORY_UPDATE"]:
                    # Skip these and keep waiting
                    continue
                else:
                    # This might be our response
                    return message
                    
            except asyncio.TimeoutError:
                continue
        
        return {"type": "TIMEOUT", "message": "No response received"}
    
    def generate_order_id(self) -> str:
        """Generate a unique order ID"""
        return f"ORD-{uuid.uuid4().hex[:8]}"
    
    async def send_order(self, side: str, product: str, quantity: int, limit_price: float | None = None, mode: str = "MARKET"):
        """Send order with correct format"""
        order_message = {
            "type": "ORDER",
            "clOrdID": self.generate_order_id(),
            "side": side,
            "mode": mode,
            "product": product,
            "qty": quantity
        }
        
        # Add limit price for LIMIT orders
        if mode == "LIMIT" and limit_price is not None:
            order_message["limitPrice"] = round(limit_price, 2)
        
        self.logger.info(f"Sending {mode} order: {side} {quantity} {product}" + 
                        (f" @ ${limit_price}" if limit_price else ""))
        if self.websocket:
            await self.websocket.send(json.dumps(order_message))
        self.orders_sent += 1
        
        # Wait for response
        response = await self.wait_for_response(["ORDER_ACK", "FILL", "ERROR"])
        
        if response.get("type") in ["ORDER_ACK", "FILL"]:
            self.orders_success += 1
            self.logger.info(f"✓ Order successful: {response}")
            return True
        elif response.get("type") == "ERROR":
            self.errors.append(f"Order failed: {response}")
            self.logger.warning(f"✗ Order failed: {response}")
            return False
        else:
            self.errors.append(f"Order timeout: {response}")
            self.logger.warning(f"⚠ Order response timeout: {response}")
            return False
    
    async def send_production(self, product: str, quantity: int):
        """Send production update with correct format"""
        production_message = {
            "type": "PRODUCTION_UPDATE",
            "product": product,
            "quantity": quantity
        }
        
        self.logger.info(f"Sending production update: {quantity} {product}")
        if self.websocket:
            await self.websocket.send(json.dumps(production_message))
        self.productions_sent += 1
        
        # Wait for response (production might not always send immediate response)
        response = await self.wait_for_response(["PRODUCTION_OK", "INVENTORY_UPDATE", "ERROR"], timeout=3.0)
        
        if response.get("type") in ["PRODUCTION_OK", "INVENTORY_UPDATE"]:
            self.productions_success += 1
            self.logger.info(f"✓ Production successful: {response}")
            return True
        elif response.get("type") == "ERROR":
            self.errors.append(f"Production failed: {response}")
            self.logger.warning(f"✗ Production failed: {response}")
            return False
        elif response.get("type") == "TIMEOUT":
            # Production might succeed without immediate response
            self.logger.info("⚠ Production sent, no immediate response (might be normal)")
            return True
        else:
            self.logger.info(f"⚠ Production sent, unexpected response: {response}")
            return True
    
    async def test_basic_functionality(self, duration_seconds: int = 30):
        """Test basic trading functionality"""
        self.running = True
        
        # Start message listener
        listener_task = asyncio.create_task(self.message_listener())
        
        self.logger.info(f"Starting {duration_seconds}-second functionality test")
        
        try:
            # Test 1: Production update
            self.logger.info("=== Test 1: Production Update ===")
            await self.send_production("FOSFO", 5)
            await asyncio.sleep(2)
            
            # Test 2: Market buy order
            self.logger.info("=== Test 2: Market Buy Order ===")
            await self.send_order("BUY", "FOSFO", 1, mode="MARKET")
            await asyncio.sleep(2)
            
            # Test 3: Limit sell order
            self.logger.info("=== Test 3: Limit Sell Order ===")
            await self.send_order("SELL", "PITA", 2, limit_price=18.50, mode="LIMIT")
            await asyncio.sleep(2)
            
            # Test 4: Another production
            self.logger.info("=== Test 4: Another Production ===")
            await self.send_production("PITA", 3)
            await asyncio.sleep(2)
            
            # Test 5: Limit buy order
            self.logger.info("=== Test 5: Limit Buy Order ===")
            await self.send_order("BUY", "GUACA", 1, limit_price=35.00, mode="LIMIT")
            await asyncio.sleep(2)
            
            # Test 6: Market sell order
            self.logger.info("=== Test 6: Market Sell Order ===")
            await self.send_order("SELL", "FOSFO", 1, mode="MARKET")
            
            # Wait a bit to see if we get any delayed responses
            await asyncio.sleep(5)
            
        except Exception as e:
            self.logger.error(f"Error in functionality test: {e}")
        finally:
            self.running = False
            listener_task.cancel()
        
        self.logger.info("Functionality test completed")
    
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
            'team_name': self.team_name,
            'orders_sent': self.orders_sent,
            'orders_success': self.orders_success,
            'productions_sent': self.productions_sent,
            'productions_success': self.productions_success,
            'errors': self.errors
        }

async def run_corrected_test(tokens: list, server_url: str, duration: int = 30):
    """Run corrected test with proper message formats"""
    logger.info(f"Starting {duration}-second corrected test with {len(tokens)} clients")
    
    clients = []
    
    # Connect clients sequentially to avoid overwhelming server
    for i, token in enumerate(tokens):
        client = CorrectedTestClient(token, server_url)
        if await client.connect():
            clients.append(client)
            logger.info(f"Client {i+1} connected successfully")
            await asyncio.sleep(1)  # Small delay between connections
        else:
            logger.error(f"Failed to connect client {token[:8]}")
    
    if not clients:
        logger.error("No clients connected - test failed")
        return
    
    logger.info(f"Successfully connected {len(clients)} clients")
    
    # Run tests sequentially to avoid conflicts
    for i, client in enumerate(clients):
        logger.info(f"Testing client {i+1} ({client.team_name})...")
        await client.test_basic_functionality(duration // len(clients))
        if i < len(clients) - 1:  # Don't wait after the last client
            await asyncio.sleep(3)
    
    # Generate final report
    logger.info("="*60)
    logger.info("CORRECTED TEST RESULTS")
    logger.info("="*60)
    
    total_orders_sent = sum(client.orders_sent for client in clients)
    total_orders_success = sum(client.orders_success for client in clients)
    total_productions_sent = sum(client.productions_sent for client in clients)
    total_productions_success = sum(client.productions_success for client in clients)
    all_errors = []
    
    for client in clients:
        all_errors.extend(client.errors)
    
    logger.info(f"Total Clients Connected: {len(clients)}")
    logger.info(f"Orders Sent: {total_orders_sent}")
    logger.info(f"Orders Successful: {total_orders_success}")
    logger.info(f"Order Success Rate: {total_orders_success/total_orders_sent*100:.1f}%" if total_orders_sent > 0 else "N/A")
    logger.info(f"Productions Sent: {total_productions_sent}")
    logger.info(f"Productions Successful: {total_productions_success}")
    logger.info(f"Production Success Rate: {total_productions_success/total_productions_sent*100:.1f}%" if total_productions_sent > 0 else "N/A")
    logger.info(f"Total Errors: {len(all_errors)}")
    
    logger.info("\nPER-CLIENT RESULTS:")
    logger.info("-" * 40)
    for i, client in enumerate(clients):
        stats = client.get_stats()
        logger.info(f"Client {i+1} ({stats['team_name']}):")
        logger.info(f"  Orders: {stats['orders_success']}/{stats['orders_sent']}")
        logger.info(f"  Productions: {stats['productions_success']}/{stats['productions_sent']}")
        logger.info(f"  Errors: {len(stats['errors'])}")
    
    if all_errors:
        logger.info("\nERROR DETAILS:")
        logger.info("-" * 20)
        for error in all_errors[:5]:  # Show first 5 errors
            logger.info(f"  - {error}")
    
    # Disconnect all clients
    for client in clients:
        await client.disconnect()
    
    logger.info("="*60)
    
    # Assessment
    if total_orders_success > 0 and total_productions_success > 0:
        logger.info("✅ SUCCESS: Server communication is working correctly!")
        logger.info("   Both orders and productions are functioning.")
        logger.info("   Ready for full trading simulation.")
    elif total_orders_success > 0:
        logger.info("⚠️  PARTIAL SUCCESS: Orders work, but productions may have issues")
    elif total_productions_success > 0:
        logger.info("⚠️  PARTIAL SUCCESS: Productions work, but orders may have issues")
    else:
        logger.warning("❌ ISSUES REMAIN: Need to investigate further")

async def main():
    """Main entry point with real team tokens"""
    # Real team tokens provided by user
    tokens = [
        "TK-09jKZrvn0NF11v99j10vT4Fx",  # Alquimistas de Palta (Premium)
        "TK-NVUoEHwzH1BRcgcyyDdhx2a4",  # Arpistas de Pita-Pita (Premium)
        "TK-XqnoG2blE3DFmApa75iexwvC",  # Avocultores del Hueso Cósmico (Básico)
        "TK-egakIjLDHsuRF4KgObBILmlE",  # Cartógrafos de Fosfolima (Premium)
    ]
    
    server_url = "wss://trading.hellsoft.tech/ws"
    
    logger.info("Starting Corrected Trading Test")
    logger.info(f"Server: {server_url}")
    logger.info(f"Testing with {len(tokens)} team tokens")
    
    try:
        await run_corrected_test(tokens, server_url, duration=45)
    except KeyboardInterrupt:
        logger.info("Test interrupted by user")
    except Exception as e:
        logger.error(f"Test failed: {e}")

if __name__ == "__main__":
    asyncio.run(main())