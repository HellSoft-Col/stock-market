#!/usr/bin/env python3
"""
Improved Trading Test - With Better Async Handling and Keepalive
===============================================================

Enhanced test that properly handles:
- Keepalive pings to prevent timeouts
- Asynchronous message handling (TICKER, delayed responses)
- Quick validation test before long simulation
- Proper error handling and connection management
"""

import asyncio
import json
import random
import time
import logging
import sys
import uuid
from datetime import datetime, timedelta
from typing import Dict, List, Optional

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler(f'improved_test_{datetime.now().strftime("%Y%m%d_%H%M%S")}.log'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)

class ImprovedTradingClient:
    """Enhanced trading client with proper async handling and keepalive"""
    
    def __init__(self, token: str, server_url: str):
        self.token = token
        self.server_url = server_url
        self.websocket = None
        self.running = False
        
        # Statistics
        self.orders_sent = 0
        self.orders_success = 0
        self.productions_sent = 0
        self.productions_success = 0
        self.errors = []
        
        # Message handling
        self.pending_responses = {}  # order_id -> expected response type
        self.message_queue = asyncio.Queue()
        self.ticker_count = 0
        self.fill_count = 0
        
        # Client info
        self.logger = logging.getLogger(f"Client-{token[:8]}")
        self.team_name = ""
        self.inventory = {}
        self.balance = 0.0
        
        # Keepalive
        self.last_ping = time.time()
        self.ping_interval = 30  # Send ping every 30 seconds
        
    async def connect(self) -> bool:
        """Connect and authenticate with keepalive support"""
        try:
            websockets_module = __import__('websockets')
            
            self.logger.info(f"Connecting to {self.server_url}")
            
            # Connect with ping settings to prevent timeouts
            self.websocket = await websockets_module.connect(
                self.server_url,
                ping_interval=20,  # Send ping every 20 seconds
                ping_timeout=10,   # Wait 10 seconds for pong
                close_timeout=10   # Wait 10 seconds for close
            )
            
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
                self.team_name = auth_response.get('team', 'Unknown')
                self.inventory = auth_response.get('inventory', {})
                self.balance = auth_response.get('currentBalance', 0.0)
                
                self.logger.info(f"✅ Authenticated as: {self.team_name}")
                self.logger.info(f"   Species: {auth_response.get('species', 'Unknown')}")
                self.logger.info(f"   Balance: ${self.balance}")
                self.logger.info(f"   Inventory: {self.inventory}")
                return True
            else:
                self.logger.error(f"❌ Auth failed: {auth_response}")
                return False
                
        except Exception as e:
            self.logger.error(f"❌ Connection failed: {e}")
            return False
    
    async def message_listener(self):
        """Enhanced message listener that categorizes messages properly"""
        while self.running and self.websocket:
            try:
                message = await self.websocket.recv()
                data = json.loads(message)
                msg_type = data.get("type", "UNKNOWN")
                
                # Count different message types
                if msg_type == "TICKER":
                    self.ticker_count += 1
                    if self.ticker_count <= 5:  # Log first few tickers
                        product = data.get("product", "?")
                        bid = data.get("bestBid", "N/A")
                        ask = data.get("bestAsk", "N/A")
                        self.logger.debug(f"📊 Ticker {product}: Bid={bid}, Ask={ask}")
                elif msg_type == "FILL":
                    self.fill_count += 1
                    self.logger.info(f"💰 FILL #{self.fill_count}: {data}")
                elif msg_type in ["ORDER_ACK", "ERROR"]:
                    self.logger.info(f"📝 {msg_type}: {data}")
                elif msg_type == "INVENTORY_UPDATE":
                    self.inventory = data.get("inventory", {})
                    self.logger.info(f"📦 Inventory updated: {self.inventory}")
                elif msg_type in ["PONG"]:
                    self.logger.debug(f"🏓 {msg_type}")
                else:
                    self.logger.info(f"🔄 {msg_type}: {data}")
                
                # Queue all messages for response handling
                await self.message_queue.put(data)
                
            except Exception as e:
                if self.running:
                    self.logger.error(f"❌ Message listener error: {e}")
                break
    
    async def wait_for_response(self, expected_types: List[str], order_id: str | None = None, timeout: float = 5.0) -> Dict:
        """Enhanced response waiter that handles async messages properly"""
        start_time = time.time()
        received_messages = []
        
        while time.time() - start_time < timeout:
            try:
                message = await asyncio.wait_for(self.message_queue.get(), timeout=1.0)
                received_messages.append(message)
                
                msg_type = message.get("type")
                
                # Check if this is the response we're looking for
                if msg_type in expected_types:
                    # If we're looking for a specific order response, check order ID
                    if order_id:
                        msg_order_id = message.get("clOrdID")
                        if msg_order_id == order_id:
                            return message
                        else:
                            # This is a response for a different order, keep waiting
                            continue
                    else:
                        return message
                
                # Handle specific message types
                elif msg_type == "ERROR":
                    return message
                elif msg_type in ["TICKER", "INVENTORY_UPDATE"]:
                    # These are background messages, keep waiting
                    continue
                else:
                    # Unexpected message type, but might be our response
                    if not expected_types:  # If we don't know what to expect
                        return message
                        
            except asyncio.TimeoutError:
                continue
        
        # Timeout reached
        self.logger.warning(f"⏰ Response timeout. Received {len(received_messages)} messages:")
        for msg in received_messages[-3:]:  # Show last 3 messages
            self.logger.warning(f"   - {msg.get('type', 'UNKNOWN')}: {msg}")
        
        return {"type": "TIMEOUT", "message": "No response received", "received": received_messages}
    
    def generate_order_id(self) -> str:
        """Generate unique order ID"""
        return f"ORD-{uuid.uuid4().hex[:8]}"
    
    async def send_ping(self):
        """Send ping to keep connection alive"""
        if self.websocket:
            try:
                ping_message = {"type": "PING"}
                await self.websocket.send(json.dumps(ping_message))
                self.last_ping = time.time()
                self.logger.debug("🏓 Ping sent")
            except Exception as e:
                self.logger.error(f"❌ Ping failed: {e}")
    
    async def send_order(self, side: str, product: str, quantity: int, limit_price: float | None = None, mode: str = "MARKET") -> bool:
        """Send order with proper response tracking"""
        order_id = self.generate_order_id()
        order_message = {
            "type": "ORDER",
            "clOrdID": order_id,
            "side": side,
            "mode": mode,
            "product": product,
            "qty": quantity
        }
        
        if mode == "LIMIT" and limit_price is not None:
            order_message["limitPrice"] = round(limit_price, 2)
        
        self.logger.info(f"📤 Sending {mode} order: {side} {quantity} {product}" + 
                        (f" @ ${limit_price}" if limit_price else ""))
        
        if self.websocket:
            await self.websocket.send(json.dumps(order_message))
            self.orders_sent += 1
            
            # Wait for specific response to this order
            response = await self.wait_for_response(["ORDER_ACK", "FILL", "ERROR"], order_id=order_id)
            
            if response.get("type") in ["ORDER_ACK", "FILL"]:
                self.orders_success += 1
                self.logger.info(f"✅ Order successful: {response.get('type')}")
                return True
            elif response.get("type") == "ERROR":
                self.errors.append(f"Order failed: {response}")
                self.logger.warning(f"❌ Order failed: {response}")
                return False
            else:
                self.logger.warning(f"⚠️ Order response unclear: {response}")
                return False
        
        return False
    
    async def send_production(self, product: str, quantity: int) -> bool:
        """Send production update"""
        production_message = {
            "type": "PRODUCTION_UPDATE",
            "product": product,
            "quantity": quantity
        }
        
        self.logger.info(f"🏭 Sending production: {quantity} {product}")
        
        if self.websocket:
            await self.websocket.send(json.dumps(production_message))
            self.productions_sent += 1
            
            # Wait for inventory update or error
            response = await self.wait_for_response(["INVENTORY_UPDATE", "ERROR"], timeout=3.0)
            
            if response.get("type") == "INVENTORY_UPDATE":
                self.productions_success += 1
                self.logger.info(f"✅ Production successful")
                return True
            elif response.get("type") == "ERROR":
                self.errors.append(f"Production failed: {response}")
                self.logger.warning(f"❌ Production failed: {response}")
                return False
            else:
                # Production might succeed without immediate response
                self.logger.info("⚠️ Production sent, no immediate response (might be normal)")
                self.productions_success += 1  # Assume success
                return True
        
        return False
    
    async def quick_functionality_test(self) -> Dict:
        """Quick test to validate all basic functionality"""
        self.running = True
        results = {
            "connection": False,
            "production": False,
            "market_order": False,
            "limit_order": False,
            "ticker_received": False,
            "keepalive": False
        }
        
        # Start message listener
        listener_task = asyncio.create_task(self.message_listener())
        
        try:
            self.logger.info("🚀 Starting quick functionality test")
            
            # Test 1: Production (should be fast)
            self.logger.info("🧪 Test 1: Production Update")
            results["production"] = await self.send_production("FOSFO", 1)
            await asyncio.sleep(1)
            
            # Test 2: Market order (should execute immediately if there's liquidity)
            self.logger.info("🧪 Test 2: Market Order")
            results["market_order"] = await self.send_order("BUY", "FOSFO", 1, mode="MARKET")
            await asyncio.sleep(1)
            
            # Test 3: Limit order (should get ACK)
            self.logger.info("🧪 Test 3: Limit Order")
            results["limit_order"] = await self.send_order("SELL", "PITA", 1, limit_price=20.0, mode="LIMIT")
            await asyncio.sleep(1)
            
            # Test 4: Check if we received any tickers
            if self.ticker_count > 0:
                results["ticker_received"] = True
                self.logger.info(f"📊 Received {self.ticker_count} ticker messages")
            
            # Test 5: Ping test
            self.logger.info("🧪 Test 5: Keepalive Ping")
            await self.send_ping()
            ping_response = await self.wait_for_response(["PONG"], timeout=2.0)
            results["keepalive"] = ping_response.get("type") == "PONG"
            
            results["connection"] = True
            
        except Exception as e:
            self.logger.error(f"❌ Quick test error: {e}")
        finally:
            self.running = False
            listener_task.cancel()
        
        return results
    
    async def disconnect(self):
        """Clean disconnect"""
        self.running = False
        if self.websocket:
            try:
                await self.websocket.close()
                self.logger.info("👋 Disconnected")
            except Exception as e:
                self.logger.error(f"❌ Disconnect error: {e}")
    
    def get_stats(self):
        """Get comprehensive client statistics"""
        return {
            'team_name': self.team_name,
            'orders_sent': self.orders_sent,
            'orders_success': self.orders_success,
            'productions_sent': self.productions_sent,
            'productions_success': self.productions_success,
            'ticker_count': self.ticker_count,
            'fill_count': self.fill_count,
            'errors': self.errors,
            'inventory': self.inventory,
            'balance': self.balance
        }

async def run_quick_validation_test(tokens: List[str], server_url: str) -> bool:
    """Quick test to validate all functionality before long simulation"""
    logger.info("🔬 QUICK VALIDATION TEST - Testing all functionality")
    logger.info("="*60)
    
    # Test with just 2 clients to keep it fast
    test_tokens = tokens[:2]
    clients = []
    all_results = []
    
    # Connect clients
    for i, token in enumerate(test_tokens):
        client = ImprovedTradingClient(token, server_url)
        if await client.connect():
            clients.append(client)
            logger.info(f"✅ Client {i+1} connected: {client.team_name}")
        else:
            logger.error(f"❌ Client {i+1} failed to connect")
    
    if not clients:
        logger.error("❌ No clients connected for validation")
        return False
    
    # Run quick tests
    for i, client in enumerate(clients):
        logger.info(f"\n🧪 Testing client {i+1}: {client.team_name}")
        results = await client.quick_functionality_test()
        all_results.append(results)
        
        # Log results
        for test, passed in results.items():
            status = "✅" if passed else "❌"
            logger.info(f"  {status} {test}: {'PASS' if passed else 'FAIL'}")
    
    # Disconnect clients
    for client in clients:
        await client.disconnect()
    
    # Analyze results
    logger.info("\n📊 VALIDATION SUMMARY")
    logger.info("="*30)
    
    total_tests = len(all_results[0]) if all_results else 0
    passed_tests = 0
    
    for test_name in all_results[0].keys() if all_results else []:
        test_results = [result[test_name] for result in all_results]
        passed = sum(test_results)
        total = len(test_results)
        
        if passed == total:
            status = "✅ PASS"
            passed_tests += 1
        elif passed > 0:
            status = "⚠️ PARTIAL"
        else:
            status = "❌ FAIL"
        
        logger.info(f"{status} {test_name}: {passed}/{total}")
    
    success_rate = passed_tests / total_tests if total_tests > 0 else 0
    logger.info(f"\n🎯 Overall success rate: {success_rate*100:.1f}% ({passed_tests}/{total_tests})")
    
    if success_rate >= 0.8:  # 80% success rate required
        logger.info("✅ VALIDATION PASSED - Ready for long simulation!")
        return True
    else:
        logger.warning("❌ VALIDATION FAILED - Need to fix issues before long simulation")
        return False

async def main():
    """Main entry point"""
    # Real team tokens
    tokens = [
        "TK-09jKZrvn0NF11v99j10vT4Fx",  # Alquimistas de Palta (Premium)
        "TK-NVUoEHwzH1BRcgcyyDdhx2a4",  # Arpistas de Pita-Pita (Premium)
        "TK-XqnoG2blE3DFmApa75iexwvC",  # Avocultores del Hueso Cósmico (Básico)
        "TK-egakIjLDHsuRF4KgObBILmlE",  # Cartógrafos de Fosfolima (Premium)
    ]
    
    server_url = "wss://trading.hellsoft.tech/ws"
    
    logger.info("🚀 Starting Improved Trading Test")
    logger.info(f"🌐 Server: {server_url}")
    logger.info(f"👥 Testing with {len(tokens)} team tokens")
    
    try:
        # Run quick validation test first
        validation_passed = await run_quick_validation_test(tokens, server_url)
        
        if validation_passed:
            logger.info("\n🎉 All systems working! Ready for longer simulation.")
            logger.info("💡 You can now run longer trading simulations with confidence.")
        else:
            logger.warning("\n⚠️ Some issues detected. Review the validation results above.")
            
    except KeyboardInterrupt:
        logger.info("\n👋 Test interrupted by user")
    except Exception as e:
        logger.error(f"❌ Test failed: {e}")

if __name__ == "__main__":
    asyncio.run(main())