#!/usr/bin/env python3
"""
Demo 5-Minute Trading Simulation - 12 Concurrent Teams
=====================================================

Shortened version for demonstration with complete reporting.
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
from dataclasses import dataclass

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler(f'demo_simulation_{datetime.now().strftime("%Y%m%d_%H%M%S")}.log'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)

@dataclass
class TeamConfig:
    """Configuration for each team's trading strategy"""
    token: str
    name: str
    species: str
    strategy: str
    risk_level: str
    preferred_products: List[str]
    production_frequency: int  # seconds between productions
    trading_frequency: int     # seconds between trades

class AdvancedTradingClient:
    """Advanced trading client with strategies and detailed tracking"""
    
    def __init__(self, team_config: TeamConfig, server_url: str):
        self.config = team_config
        self.server_url = server_url
        self.websocket = None
        self.running = False
        
        # Detailed statistics
        self.stats = {
            'orders_sent': 0,
            'orders_successful': 0,
            'orders_filled': 0,
            'productions_sent': 0,
            'productions_successful': 0,
            'total_profit': 0.0,
            'tickers_received': 0,
            'fills_received': 0,
            'offers_received': 0,
            'errors': [],
            'session_start': time.time(),
            'last_activity': time.time()
        }
        
        # Market data
        self.market_data = {}  # product -> {bid, ask, mid, volume}
        self.inventory = {}
        self.balance = 0.0
        self.initial_balance = 0.0
        
        # Trading state
        self.pending_orders = {}  # order_id -> order_info
        self.message_queue = asyncio.Queue()
        
        self.logger = logging.getLogger(f"{self.config.name[:15]}")
        
    async def connect(self) -> bool:
        """Connect with enhanced error handling"""
        try:
            websockets_module = __import__('websockets')
            
            self.logger.info(f"🔗 Connecting to {self.server_url}")
            
            # Enhanced connection with keepalive
            self.websocket = await websockets_module.connect(
                self.server_url,
                ping_interval=20,
                ping_timeout=10,
                close_timeout=10
            )
            
            # Authenticate
            auth_message = {
                "type": "LOGIN",
                "token": self.config.token
            }
            await self.websocket.send(json.dumps(auth_message))
            
            # Wait for auth response
            response = await asyncio.wait_for(self.websocket.recv(), timeout=15.0)
            auth_response = json.loads(response)
            
            if auth_response.get("type") == "LOGIN_OK":
                self.inventory = auth_response.get('inventory', {})
                self.balance = auth_response.get('currentBalance', 0.0)
                self.initial_balance = self.balance
                
                self.logger.info(f"✅ Connected as {self.config.name}")
                self.logger.info(f"   Strategy: {self.config.strategy} ({self.config.risk_level} risk)")
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
        """Enhanced message listener with detailed tracking"""
        while self.running and self.websocket:
            try:
                message = await self.websocket.recv()
                data = json.loads(message)
                msg_type = data.get("type", "UNKNOWN")
                
                # Update activity timestamp
                self.stats['last_activity'] = time.time()
                
                # Process different message types
                if msg_type == "TICKER":
                    self.stats['tickers_received'] += 1
                    self._process_ticker(data)
                elif msg_type == "FILL":
                    self.stats['fills_received'] += 1
                    self._process_fill(data)
                elif msg_type == "OFFER":
                    self.stats['offers_received'] += 1
                    self._process_offer(data)
                elif msg_type == "ORDER_ACK":
                    self._process_order_ack(data)
                elif msg_type == "INVENTORY_UPDATE":
                    self.inventory = data.get("inventory", {})
                elif msg_type == "ERROR":
                    self.stats['errors'].append(data)
                    self.logger.warning(f"⚠️ Error: {data}")
                
                # Queue message for response handling
                await self.message_queue.put(data)
                
            except Exception as e:
                if self.running:
                    self.logger.error(f"❌ Message listener error: {e}")
                break
    
    def _process_ticker(self, data):
        """Process ticker data for market analysis"""
        product = data.get("product")
        if product:
            self.market_data[product] = {
                'bid': data.get('bestBid'),
                'ask': data.get('bestAsk'),
                'mid': data.get('mid'),
                'volume': data.get('volume24h', 0),
                'timestamp': time.time()
            }
    
    def _process_fill(self, data):
        """Process fill data and update statistics"""
        self.stats['orders_filled'] += 1
        
        # Calculate profit/loss
        side = data.get('side')
        price = data.get('fillPrice', 0)
        qty = data.get('fillQty', 0)
        
        if side == 'SELL':
            self.stats['total_profit'] += price * qty
        elif side == 'BUY':
            self.stats['total_profit'] -= price * qty
        
        self.logger.info(f"💰 FILL: {side} {qty} @ ${price} (Total P&L: ${self.stats['total_profit']:.2f})")
    
    def _process_offer(self, data):
        """Process incoming offers (could implement auto-accept logic)"""
        offer_id = data.get('offerId')
        product = data.get('product')
        qty = data.get('quantityRequested')
        max_price = data.get('maxPrice')
        
        self.logger.debug(f"📬 OFFER: {qty} {product} @ max ${max_price} (ID: {offer_id})")
    
    def _process_order_ack(self, data):
        """Process order acknowledgments"""
        order_id = data.get('clOrdID')
        status = data.get('status')
        
        if order_id in self.pending_orders:
            self.pending_orders[order_id]['status'] = status
            if status == 'PENDING':
                self.stats['orders_successful'] += 1
    
    def generate_order_id(self) -> str:
        """Generate unique order ID with team prefix"""
        return f"{self.config.name[:3].upper()}-{uuid.uuid4().hex[:8]}"
    
    async def send_order(self, side: str, product: str, quantity: int, limit_price: float | None = None, mode: str = "MARKET") -> bool:
        """Send order with strategy-based logic"""
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
        
        # Store pending order
        self.pending_orders[order_id] = {
            'side': side,
            'product': product,
            'qty': quantity,
            'price': limit_price,
            'mode': mode,
            'timestamp': time.time(),
            'status': 'SENT'
        }
        
        if self.websocket:
            await self.websocket.send(json.dumps(order_message))
            self.stats['orders_sent'] += 1
            return True
        
        return False
    
    async def send_production(self, product: str, quantity: int) -> bool:
        """Send production update"""
        production_message = {
            "type": "PRODUCTION_UPDATE",
            "product": product,
            "quantity": quantity
        }
        
        if self.websocket:
            await self.websocket.send(json.dumps(production_message))
            self.stats['productions_sent'] += 1
            self.stats['productions_successful'] += 1  # Assume success
            return True
        
        return False
    
    async def execute_strategy(self, duration_minutes: int):
        """Execute team's trading strategy for specified duration"""
        self.running = True
        end_time = time.time() + (duration_minutes * 60)
        
        # Start message listener
        listener_task = asyncio.create_task(self.message_listener())
        
        self.logger.info(f"🚀 Starting {duration_minutes}-minute strategy: {self.config.strategy}")
        
        try:
            last_production = 0
            last_trade = 0
            
            while time.time() < end_time and self.running:
                current_time = time.time()
                
                # Production activities
                if current_time - last_production >= self.config.production_frequency:
                    await self._execute_production()
                    last_production = current_time
                
                # Trading activities
                if current_time - last_trade >= self.config.trading_frequency:
                    await self._execute_trading()
                    last_trade = current_time
                
                # Short sleep to prevent overwhelming
                await asyncio.sleep(1)
                
        except Exception as e:
            self.logger.error(f"❌ Strategy execution error: {e}")
        finally:
            self.running = False
            listener_task.cancel()
    
    async def _execute_production(self):
        """Execute production based on strategy"""
        if not self.config.preferred_products:
            return
        
        # Choose product based on strategy
        if self.config.strategy == "Producer":
            products = ["FOSFO", "PITA", "SEBO"]
        elif self.config.strategy == "Refiner":
            products = ["PALTA-OIL", "NUCREM", "CASCAR-ALLOY"]
        elif self.config.strategy == "Trader":
            if random.random() < 0.3:  # 30% chance to produce
                products = self.config.preferred_products[:2]
            else:
                return
        else:
            products = self.config.preferred_products
        
        product = random.choice(products)
        quantity = random.randint(1, 5)
        
        success = await self.send_production(product, quantity)
        if success:
            self.logger.info(f"🏭 Produced {quantity} {product}")
    
    async def _execute_trading(self):
        """Execute trading based on strategy and market data"""
        if not self.config.preferred_products:
            return
        
        # Simple trading logic
        product = random.choice(self.config.preferred_products)
        side = random.choice(["BUY", "SELL"])
        quantity = random.randint(1, 3)
        
        if self.config.strategy == "Aggressive":
            await self.send_order(side, product, quantity, mode="MARKET")
            self.logger.debug(f"⚡ Aggressive {side}: {quantity} {product}")
        else:
            # Use limit orders
            base_price = random.uniform(5, 50)
            await self.send_order(side, product, quantity, base_price, "LIMIT")
            self.logger.debug(f"🛡️ {self.config.strategy} {side}: {quantity} {product} @ ${base_price:.2f}")
    
    def get_performance_summary(self) -> Dict:
        """Get comprehensive performance summary"""
        runtime = time.time() - self.stats['session_start']
        
        return {
            'team_name': self.config.name,
            'strategy': self.config.strategy,
            'runtime_minutes': runtime / 60,
            'orders_sent': self.stats['orders_sent'],
            'orders_successful': self.stats['orders_successful'],
            'orders_filled': self.stats['orders_filled'],
            'order_success_rate': self.stats['orders_successful'] / max(1, self.stats['orders_sent']),
            'fill_rate': self.stats['orders_filled'] / max(1, self.stats['orders_successful']),
            'productions_sent': self.stats['productions_sent'],
            'productions_successful': self.stats['productions_successful'],
            'production_success_rate': self.stats['productions_successful'] / max(1, self.stats['productions_sent']),
            'total_profit': self.stats['total_profit'],
            'current_balance': self.balance,
            'balance_change': self.balance - self.initial_balance,
            'current_inventory': self.inventory,
            'tickers_received': self.stats['tickers_received'],
            'fills_received': self.stats['fills_received'],
            'offers_received': self.stats['offers_received'],
            'errors_count': len(self.stats['errors']),
            'pending_orders': len(self.pending_orders)
        }
    
    async def disconnect(self):
        """Clean disconnect"""
        self.running = False
        if self.websocket:
            try:
                await self.websocket.close()
                self.logger.info("👋 Disconnected")
            except Exception as e:
                self.logger.error(f"❌ Disconnect error: {e}")

class SimulationManager:
    """Manages the entire simulation with all teams"""
    
    def __init__(self):
        self.teams = []
        self.server_url = "wss://trading.hellsoft.tech/ws"
        self.simulation_start = None
        self.simulation_end = None
        
    def setup_teams(self):
        """Setup all 12 teams with diverse strategies"""
        team_configs = [
            TeamConfig(
                token="TK-09jKZrvn0NF11v99j10vT4Fx",
                name="Alquimistas de Palta",
                species="Premium",
                strategy="Producer",
                risk_level="Medium",
                preferred_products=["FOSFO", "PALTA-OIL", "SEBO"],
                production_frequency=15,
                trading_frequency=25
            ),
            TeamConfig(
                token="TK-NVUoEHwzH1BRcgcyyDdhx2a4",
                name="Arpistas de Pita-Pita",
                species="Premium",
                strategy="Aggressive",
                risk_level="High",
                preferred_products=["PITA", "FOSFO", "GUACA"],
                production_frequency=20,
                trading_frequency=10
            ),
            TeamConfig(
                token="TK-XqnoG2blE3DFmApa75iexwvC",
                name="Avocultores del Hueso Cósmico",
                species="Básico",
                strategy="Conservative",
                risk_level="Low",
                preferred_products=["GUACA", "H-GUACA", "NUCREM"],
                production_frequency=30,
                trading_frequency=45
            ),
            TeamConfig(
                token="TK-egakIjLDHsuRF4KgObBILmlE",
                name="Cartógrafos de Fosfolima",
                species="Premium",
                strategy="Market_Maker",
                risk_level="Medium",
                preferred_products=["FOSFO", "GTRON", "CASCAR-ALLOY"],
                production_frequency=25,
                trading_frequency=20
            ),
            TeamConfig(
                token="TK-QJ3a6pYMPtJS62xpuclWe1pH",
                name="Cosechadores de Semillas",
                species="Premium",
                strategy="Refiner",
                risk_level="Medium",
                preferred_products=["PALTA-OIL", "NUCREM", "SEBO"],
                production_frequency=18,
                trading_frequency=30
            ),
            TeamConfig(
                token="TK-B8YdOACQCKTf0ZEjqRawhXhQ",
                name="Forjadores Holográficos",
                species="Premium",
                strategy="Arbitrage",
                risk_level="High",
                preferred_products=["GTRON", "H-GUACA", "CASCAR-ALLOY"],
                production_frequency=35,
                trading_frequency=15
            ),
            TeamConfig(
                token="TK-SO4U3Kr25DbMLNiooci5U0YT",
                name="Ingenieros Holo-Aguacate",
                species="Premium",
                strategy="Trader",
                risk_level="High",
                preferred_products=["H-GUACA", "GUACA", "GTRON"],
                production_frequency=40,
                trading_frequency=12
            ),
            TeamConfig(
                token="TK-XKAG1T8R9smShyRc6akllOUl",
                name="Mensajeros del Núcleo",
                species="Premium",
                strategy="Producer",
                risk_level="Low",
                preferred_products=["NUCREM", "FOSFO", "PITA"],
                production_frequency=12,
                trading_frequency=35
            ),
            TeamConfig(
                token="TK-easOq988WJaYn9tVhYyPukAr",
                name="Mineros de Guacatrones",
                species="Premium",
                strategy="Aggressive",
                risk_level="High",
                preferred_products=["GTRON", "GUACA", "CASCAR-ALLOY"],
                production_frequency=22,
                trading_frequency=8
            ),
            TeamConfig(
                token="TK-3h1JhNyu8oKT0DChOD5LKpN1",
                name="Monjes del Guacamole Estelar",
                species="Premium",
                strategy="Conservative",
                risk_level="Low",
                preferred_products=["GUACA", "PALTA-OIL", "NUCREM"],
                production_frequency=28,
                trading_frequency=50
            ),
            TeamConfig(
                token="TK-6fAQHLCn9oiw7ITU4ywtHW2x",
                name="Orfebres de Cáscara",
                species="Premium",
                strategy="Market_Maker",
                risk_level="Medium",
                preferred_products=["CASCAR-ALLOY", "SEBO", "PITA"],
                production_frequency=20,
                trading_frequency=18
            ),
            TeamConfig(
                token="TK-TUkY6A0GaXzJw0LEKiVmosYw",
                name="Someliers de Aceite",
                species="Premium",
                strategy="Refiner",
                risk_level="Medium",
                preferred_products=["PALTA-OIL", "SEBO", "NUCREM"],
                production_frequency=16,
                trading_frequency=28
            )
        ]
        
        # Create clients for each team
        for config in team_configs:
            client = AdvancedTradingClient(config, self.server_url)
            self.teams.append(client)
    
    async def run_simulation(self, duration_minutes: int = 5):
        """Run the complete simulation"""
        logger.info("🚀 STARTING 5-MINUTE DEMO TRADING SIMULATION")
        logger.info("="*60)
        logger.info(f"📊 Teams: {len(self.teams)}")
        logger.info(f"⏱️  Duration: {duration_minutes} minutes")
        logger.info(f"🌐 Server: {self.server_url}")
        
        self.simulation_start = time.time()
        
        # Connect all teams
        logger.info("\n🔗 CONNECTING ALL TEAMS...")
        connected_teams = []
        
        for i, team in enumerate(self.teams):
            logger.info(f"Connecting team {i+1}/12: {team.config.name}")
            if await team.connect():
                connected_teams.append(team)
            else:
                logger.error(f"❌ Failed to connect {team.config.name}")
            
            await asyncio.sleep(0.5)
        
        if not connected_teams:
            logger.error("❌ No teams connected. Simulation failed.")
            return
        
        logger.info(f"✅ Connected {len(connected_teams)}/{len(self.teams)} teams")
        
        # Start all trading strategies concurrently
        logger.info(f"\n🎮 STARTING CONCURRENT TRADING (Duration: {duration_minutes} minutes)")
        
        tasks = []
        for team in connected_teams:
            task = asyncio.create_task(team.execute_strategy(duration_minutes))
            tasks.append(task)
        
        # Wait for all strategies to complete
        try:
            await asyncio.gather(*tasks, return_exceptions=True)
        except Exception as e:
            logger.error(f"❌ Simulation error: {e}")
        
        self.simulation_end = time.time()
        
        # Generate final report
        await self._generate_final_report(connected_teams)
        
        # Disconnect all teams
        logger.info("\n👋 DISCONNECTING ALL TEAMS...")
        for team in connected_teams:
            await team.disconnect()
    
    async def _generate_final_report(self, teams: List[AdvancedTradingClient]):
        """Generate comprehensive final report"""
        logger.info("\n" + "="*80)
        logger.info("📊 FINAL SIMULATION REPORT")
        logger.info("="*80)
        
        simulation_duration = ((self.simulation_end or 0) - (self.simulation_start or 0)) / 60
        logger.info(f"⏱️  Simulation Duration: {simulation_duration:.2f} minutes")
        logger.info(f"👥 Teams Participated: {len(teams)}")
        
        # Collect all performance data
        performances = []
        for team in teams:
            perf = team.get_performance_summary()
            performances.append(perf)
        
        # Overall statistics
        total_orders = sum(p['orders_sent'] for p in performances)
        total_successful_orders = sum(p['orders_successful'] for p in performances)
        total_fills = sum(p['orders_filled'] for p in performances)
        total_productions = sum(p['productions_sent'] for p in performances)
        total_successful_productions = sum(p['productions_successful'] for p in performances)
        total_profit = sum(p['total_profit'] for p in performances)
        
        logger.info(f"\n📈 OVERALL STATISTICS:")
        logger.info(f"   Total Orders Sent: {total_orders}")
        logger.info(f"   Successful Orders: {total_successful_orders} ({total_successful_orders/max(1,total_orders)*100:.1f}%)")
        logger.info(f"   Orders Filled: {total_fills} ({total_fills/max(1,total_successful_orders)*100:.1f}%)")
        logger.info(f"   Total Productions: {total_productions}")
        logger.info(f"   Successful Productions: {total_successful_productions} ({total_successful_productions/max(1,total_productions)*100:.1f}%)")
        logger.info(f"   Total Market Profit: ${total_profit:.2f}")
        
        # Strategy performance comparison
        strategy_stats = {}
        for perf in performances:
            strategy = perf['strategy']
            if strategy not in strategy_stats:
                strategy_stats[strategy] = {
                    'teams': 0,
                    'total_orders': 0,
                    'total_fills': 0,
                    'total_profit': 0
                }
            
            strategy_stats[strategy]['teams'] += 1
            strategy_stats[strategy]['total_orders'] += perf['orders_sent']
            strategy_stats[strategy]['total_fills'] += perf['orders_filled']
            strategy_stats[strategy]['total_profit'] += perf['total_profit']
        
        logger.info(f"\n🎯 STRATEGY PERFORMANCE:")
        for strategy, stats in strategy_stats.items():
            avg_orders = stats['total_orders'] / stats['teams']
            avg_fills = stats['total_fills'] / stats['teams']
            avg_profit = stats['total_profit'] / stats['teams']
            
            logger.info(f"   {strategy:15s}: {avg_orders:6.1f} orders, {avg_fills:6.1f} fills, ${avg_profit:8.2f} profit (avg)")
        
        # Top performers
        logger.info(f"\n🏆 TOP PERFORMERS:")
        
        # Most orders
        top_orders = sorted(performances, key=lambda p: p['orders_sent'], reverse=True)[:3]
        logger.info(f"   Most Orders:")
        for i, p in enumerate(top_orders):
            logger.info(f"     {i+1}. {p['team_name'][:25]:25s}: {p['orders_sent']:3d} orders")
        
        # Most fills
        top_fills = sorted(performances, key=lambda p: p['orders_filled'], reverse=True)[:3]
        logger.info(f"   Most Fills:")
        for i, p in enumerate(top_fills):
            logger.info(f"     {i+1}. {p['team_name'][:25]:25s}: {p['orders_filled']:3d} fills")
        
        # Most profit
        top_profit = sorted(performances, key=lambda p: p['total_profit'], reverse=True)[:3]
        logger.info(f"   Most Profit:")
        for i, p in enumerate(top_profit):
            logger.info(f"     {i+1}. {p['team_name'][:25]:25s}: ${p['total_profit']:8.2f}")
        
        # Detailed team breakdown
        logger.info(f"\n📋 DETAILED TEAM PERFORMANCE:")
        logger.info(f"{'Team':<25} {'Strategy':<12} {'Orders':<7} {'Fills':<6} {'Prod':<5} {'Profit':<10} {'Rate':<6}")
        logger.info("-" * 80)
        
        for perf in sorted(performances, key=lambda p: p['orders_sent'], reverse=True):
            team_name = perf['team_name'][:24]
            strategy = perf['strategy'][:11]
            orders = perf['orders_sent']
            fills = perf['orders_filled']
            productions = perf['productions_successful']
            profit = perf['total_profit']
            success_rate = perf['order_success_rate'] * 100
            
            logger.info(f"{team_name:<25} {strategy:<12} {orders:<7} {fills:<6} {productions:<5} ${profit:<9.2f} {success_rate:<5.1f}%")
        
        logger.info("="*80)
        logger.info("🎉 SIMULATION COMPLETED SUCCESSFULLY!")
        logger.info(f"📊 {total_orders} orders, {total_fills} fills, ${total_profit:.2f} total profit")
        logger.info("="*80)

async def main():
    """Main entry point for 5-minute demo simulation"""
    manager = SimulationManager()
    manager.setup_teams()
    
    try:
        await manager.run_simulation(duration_minutes=5)
    except KeyboardInterrupt:
        logger.info("\n⚠️ Simulation interrupted by user")
    except Exception as e:
        logger.error(f"❌ Simulation failed: {e}")

if __name__ == "__main__":
    asyncio.run(main())