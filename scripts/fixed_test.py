#!/usr/bin/env python3
"""
Fixed Trading Test - With Correct Message Formats
================================================

Test with corrected message formats based on error analysis.
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
        logging.FileHandler(f'fixed_test_{datetime.now().strftime("%Y%m%d_%H%M%S")}.log'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)

class FixedTestClient:
    """Test client with corrected message formats and proper WebSocket handling"""
    
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
        """Separate coroutine to listen for messages and avoid recv conflicts"""
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
        """Wait for a specific type of response from the message queue"""
        start_time = time.time()
        
        while time.time() - start_time < timeout:
            try:
                # Check if there's a message in the queue
                message = await asyncio.wait_for(self.message_queue.get(), timeout=1.0)
                
                if message.get("type") in expected_types:
                    return message
                elif message.get("type") == "TICKER":
                    # Ignore ticker messages, they're just market updates
                    self.logger.debug(f"Ignoring ticker: {message}")
                    continue
                else:
                    # This might be our response
                    return message
                    
            except asyncio.TimeoutError:
                continue
        
        return {"type": "TIMEOUT", "message": "No response received"}
    
    async def send_order_v1(self, side: str, product: str, quantity: int, price: float):
        """Send order with format v1 (current format)"""
        order_message = {
            "type": "ORDER",
            "side": side,
            "product": product,
            "quantity": quantity,
            "price": round(price, 2),
            "mode": "LIMIT"
        }
        
        self.logger.info(f"Sending order v1: {side} {quantity} {product} @ ${price}")
        await self.websocket.send(json.dumps(order_message))
        self.orders_sent += 1
        
        response = await self.wait_for_response(["ORDER_SUCCESS", "ORDER_OK", "ERROR"])
        
        if response.get("type") in ["ORDER_SUCCESS", "ORDER_OK"]:
            self.orders_success += 1
            self.logger.info(f"✓ Order v1 successful: {response}")
            return True
        else:
            self.errors.append(f"Order v1 failed: {response}")
            self.logger.warning(f"✗ Order v1 failed: {response}")
            return False
    
    async def send_order_v2(self, side: str, product: str, quantity: int, price: float):
        """Send order with format v2 (nested data)"""
        order_message = {
            "type": "ORDER",
            "data": {
                "side": side,
                "product": product,
                "quantity": quantity,
                "price": round(price, 2),
                "mode": "LIMIT"
            }
        }
        
        self.logger.info(f"Sending order v2: {side} {quantity} {product} @ ${price}")
        await self.websocket.send(json.dumps(order_message))
        self.orders_sent += 1
        
        response = await self.wait_for_response(["ORDER_SUCCESS", "ORDER_OK", "ERROR"])
        
        if response.get("type") in ["ORDER_SUCCESS", "ORDER_OK"]:
            self.orders_success += 1
            self.logger.info(f"✓ Order v2 successful: {response}")
            return True
        else:
            self.errors.append(f"Order v2 failed: {response}")
            self.logger.warning(f"✗ Order v2 failed: {response}")
            return False
    
    async def send_production_v1(self, product: str, quantity: int):
        """Send production with format v1"""
        production_message = {
            "type": "PRODUCTION",
            "product": product,
            "quantity": quantity
        }
        
        self.logger.info(f"Sending production v1: {quantity} {product}")
        await self.websocket.send(json.dumps(production_message))
        self.productions_sent += 1
        
        response = await self.wait_for_response(["PRODUCTION_SUCCESS", "PRODUCTION_OK", "ERROR"])
        
        if response.get("type") in ["PRODUCTION_SUCCESS", "PRODUCTION_OK"]:
            self.productions_success += 1
            self.logger.info(f"✓ Production v1 successful: {response}")
            return True
        else:
            self.errors.append(f"Production v1 failed: {response}")
            self.logger.warning(f"✗ Production v1 failed: {response}")
            return False
    
    async def send_production_v2(self, product: str, quantity: int):
        """Send production with format v2"""
        production_message = {
            "type": "PRODUCTION",
            "data": {
                "product": product,
                "quantity": quantity
            }
        }
        
        self.logger.info(f"Sending production v2: {quantity} {product}")
        await self.websocket.send(json.dumps(production_message))
        self.productions_sent += 1
        
        response = await self.wait_for_response(["PRODUCTION_SUCCESS", "PRODUCTION_OK", "ERROR"])
        
        if response.get("type") in ["PRODUCTION_SUCCESS", "PRODUCTION_OK"]:
            self.productions_success += 1
            self.logger.info(f"✓ Production v2 successful: {response}")
            return True
        else:
            self.errors.append(f"Production v2 failed: {response}")
            self.logger.warning(f"✗ Production v2 failed: {response}")
            return False
    
    async def test_message_formats(self, duration_seconds: int):
        """Test different message formats"""
        self.running = True
        
        # Start message listener
        listener_task = asyncio.create_task(self.message_listener())
        
        self.logger.info(f"Starting {duration_seconds}-second format test")
        
        try:
            # Test production formats
            await self.send_production_v1("FOSFO", 3)
            await asyncio.sleep(2)
            
            await self.send_production_v2("FOSFO", 5)
            await asyncio.sleep(2)
            
            # Test order formats
            await self.send_order_v1("BUY", "FOSFO", 1, 10.00)
            await asyncio.sleep(2)
            
            await self.send_order_v2("SELL", "PITA", 2, 18.00)
            await asyncio.sleep(2)
            
            # Test more orders
            await self.send_order_v1("BUY", "GUACA", 1, 35.00)
            
        except Exception as e:
            self.logger.error(f"Error in format test: {e}")
        finally:
            self.running = False
            listener_task.cancel()
        
        self.logger.info("Format test completed")
    
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

async def run_format_test(tokens: list, server_url: str, duration: int = 30):
    """Run test with corrected message formats"""
    logger.info(f"Starting {duration}-second format test with {len(tokens)} clients")
    
    clients = []
    
    # Connect clients one by one to avoid overwhelming
    for i, token in enumerate(tokens):
        client = FixedTestClient(token, server_url)
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
    
    # Run tests sequentially to avoid message conflicts
    for i, client in enumerate(clients):
        logger.info(f"Testing client {i+1} message formats...")
        await client.test_message_formats(duration // len(clients))
        await asyncio.sleep(2)  # Pause between clients
    
    # Generate analysis report
    logger.info("="*50)
    logger.info("MESSAGE FORMAT TEST REPORT")
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
        for error in all_errors[:3]:  # Show first 3 errors
            logger.info(f"  - {error}")
    
    # Disconnect all clients
    for client in clients:
        await client.disconnect()
    
    logger.info("="*50)
    
    # Final assessment
    if total_orders_success > 0 or total_productions_success > 0:
        logger.info("✓ SUCCESS: Found working message formats!")
        if total_orders_success > 0:
            logger.info("  - Orders are working")
        if total_productions_success > 0:
            logger.info("  - Productions are working")
        logger.info("✓ Ready for full 15-minute simulation")
    else:
        logger.warning("⚠ STILL ISSUES: Need to investigate server API more")

async def main():
    """Main entry point"""
    tokens = [
        "TK-09jKZrvn0NF11v99j10vT4Fx",  # Alquimistas de Palta
        "TK-NVUoEHwzH1BRcgcyyDdhx2a4",  # Arpistas de Pita-Pita
    ]
    
    server_url = "wss://trading.hellsoft.tech/ws"
    
    logger.info("Starting Message Format Test")
    logger.info(f"Server: {server_url}")
    logger.info(f"Testing with {len(tokens)} clients")
    
    try:
        await run_format_test(tokens, server_url, duration=30)
    except KeyboardInterrupt:
        logger.info("Test interrupted by user")
    except Exception as e:
        logger.error(f"Test failed: {e}")

if __name__ == "__main__":
    asyncio.run(main())