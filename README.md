# 🥑 Intergalactic Avocado Stock Exchange Server

[![Build Container](https://github.com/HellSoft-Col/stock-market/actions/workflows/build-container.yml/badge.svg)](https://github.com/HellSoft-Col/stock-market/actions/workflows/build-container.yml)
[![Publish Java SDK](https://github.com/HellSoft-Col/stock-market/actions/workflows/publish-java-sdk.yml/badge.svg)](https://github.com/HellSoft-Col/stock-market/actions/workflows/publish-java-sdk.yml)
[![Security Scan](https://github.com/HellSoft-Col/stock-market/actions/workflows/security-scan.yml/badge.svg)](https://github.com/HellSoft-Col/stock-market/actions/workflows/security-scan.yml)

A real-time trading system for educational tournaments where university students compete in a virtual stock exchange. Now featuring WebSocket connectivity, a modern web UI for real-time market visualization, and official SDKs! 🚀

## 📦 Official SDKs

### Java SDK ☕

Connect to the exchange using our official Java 25 SDK with WebSocket support:

```java
import tech.hellsoft.trading.*;
import tech.hellsoft.trading.enums.*;

// Create connector
ConectorBolsa connector = new ConectorBolsa();

// Add listener
connector.addListener(new EventListener() {
    @Override
    public void onLoginOk(LoginOKMessage msg) {
        System.out.println("✅ Logged in as: " + msg.getTeam());
    }
    
    @Override
    public void onFill(FillMessage msg) {
        System.out.println("💰 Fill: " + msg.getFillQty() + " @ " + msg.getFillPrice());
    }
    // ... other callbacks
});

// Connect and authenticate
connector.conectar("localhost", 9000, "your-token-here");

// Place an order
OrderMessage order = OrderMessage.builder()
    .clOrdID("order-001")
    .side(OrderSide.BUY)
    .mode(OrderMode.MARKET)
    .product(Product.GUACA)
    .qty(10)
    .build();
    
connector.enviarOrden(order);
```

#### Installation

**Gradle:**
```kotlin
repositories {
    maven {
        url = uri("https://maven.pkg.github.com/HellSoft-Col/stock-market")
        credentials {
            username = project.findProperty("gpr.user") as String? ?: System.getenv("GITHUB_ACTOR")
            password = project.findProperty("gpr.key") as String? ?: System.getenv("GITHUB_TOKEN")
        }
    }
}

dependencies {
    implementation("tech.hellsoft.trading:websocket-client:1.0.0-SNAPSHOT")
}
```

**Maven:**
```xml
<repositories>
    <repository>
        <id>github</id>
        <url>https://maven.pkg.github.com/HellSoft-Col/stock-market</url>
    </repository>
</repositories>

<dependency>
    <groupId>tech.hellsoft.trading</groupId>
    <artifactId>websocket-client</artifactId>
    <version>1.0.0-SNAPSHOT</version>
</dependency>
```

📚 **[Full Java SDK Documentation →](sdk/java/websocket-client/README.md)**

**Features:**
- ☕ Java 25 with Virtual Threads for high concurrency
- 🔒 Thread-safe with lock-free reads and single send semaphore
- ⚡ Sequential message processing guarantees ordering
- 🔄 Automatic heartbeat and connection management
- 🧵 Asynchronous callbacks on virtual threads

> **Note:** This is currently the only official SDK. Python and JavaScript SDKs are not planned.

## Quick Start

### 1. Build the server 🔨
```bash
go build -o exchange-server ./cmd/server
```

### 2. Run the server 🚀
```bash
./exchange-server
```

### 3. Access the Web UI 🌐
Open your browser and navigate to: **http://localhost:9000**

The web interface provides:
- Real-time market data visualization
- Interactive trading interface  
- Live order book display
- Team authentication
- Message monitoring

### 4. Set up MongoDB (for local development) 🗄️
```bash
# Start MongoDB replica set with Docker
docker-compose up -d mongodb1 mongodb2 mongodb3 mongo-init

# Wait for replica set initialization
sleep 15
```

### 5. Seed test data 🌱
```bash
./seed-teams
```

### 6. Test the exchange 🧪

#### Using Web Interface
1. Open http://localhost:9000 in your browser
2. Enter a team token: `TK-ANDROMEDA-2025-AVOCULTORES`
3. Click "Connect" then "Login"
4. Place orders using the trading interface
5. Watch real-time market data updates

#### Using Command Line Tools
```bash
# Test authentication
./test-login TK-ANDROMEDA-2025-AVOCULTORES

# Test placing orders
./test-order TK-ANDROMEDA-2025-AVOCULTORES BUY FOSFO 10
./test-order TK-ORION-2025-MONJES SELL FOSFO 5

# Test basic connectivity  
./simple-client
```

### 7. Test complete trading flow 📊
```bash
# Terminal 1: Start a BUY order (will wait in order book and generate OFFER)
./test-order TK-ANDROMEDA-2025-AVOCULTORES BUY FOSFO 10

# Terminal 2: Start a SELL order (should match and both get FILL messages)
./test-order TK-ORION-2025-MONJES SELL FOSFO 10

# Terminal 3: Test resync after trades
./test-resync TK-ANDROMEDA-2025-AVOCULTORES

# Terminal 4: Watch real-time TICKER messages
./simple-client
```

### 8. Advanced testing 🔬
```bash
# Test production updates
./test-production TK-ANDROMEDA-2025-AVOCULTORES GUACA 16  # Should succeed (authorized)
./test-production TK-ANDROMEDA-2025-AVOCULTORES FOSFO 10  # Should fail (not authorized)

# Test rate limiting (send many orders quickly)
for i in {1..105}; do ./test-order TK-ANDROMEDA-2025-AVOCULTORES BUY FOSFO 1; done

# Test offer acceptance flow
./test-order TK-ANDROMEDA-2025-AVOCULTORES BUY H-GUACA 1  # Should generate OFFER

# Test partial fills
./test-order TK-ANDROMEDA-2025-AVOCULTORES BUY FOSFO 15
./test-order TK-ORION-2025-MONJES SELL FOSFO 10  # Partial fill
./test-order TK-VEGA-2025-ALQUIMISTAS SELL FOSFO 5  # Complete the order
```

#### Expected LOGIN_OK Response:
```json
{
  "type": "LOGIN_OK",
  "team": "EquipoAndromeda", 
  "species": "Avocultores del Hueso Cósmico",
  "initialBalance": 10000.0,
  "authorizedProducts": ["PALTA-OIL", "GUACA", "SEBO"],
  "recipes": {
    "GUACA": {
      "type": "PREMIUM",
      "ingredients": {"FOSFO": 5, "PITA": 3},
      "premiumBonus": 1.30
    }
  },
  "role": {
    "branches": 2,
    "maxDepth": 4,
    "decay": 0.7651,
    "budget": 24.83
  },
  "serverTime": "2025-01-28T12:00:00Z"
}
```

#### Example FILL Response:
```json
{
  "type": "FILL",
  "clOrdID": "ORD-EquipoAndromeda-1738012345-a1b2c3d4",
  "fillQty": 10,
  "fillPrice": 7.35,
  "side": "BUY",
  "product": "FOSFO",
  "counterparty": "EquipoOrion",
  "counterpartyMessage": "Pure phospholime from nebulas ⭐",
  "serverTime": "2025-01-28T12:05:30Z",
  "remainingQty": 0,
  "totalQty": 10
}
```

#### Example TICKER Broadcast:
```json
{
  "type": "TICKER",
  "product": "FOSFO",
  "bestBid": 7.20,
  "bestAsk": 7.50,
  "mid": 7.35,
  "volume24h": 450,
  "serverTime": "2025-01-28T12:05:00Z"
}
```

#### Example OFFER Message:
```json
{
  "type": "OFFER",
  "offerId": "off-1738012350-a1b2c3d4",
  "buyer": "EquipoAndromeda",
  "product": "H-GUACA",
  "quantityRequested": 1,
  "maxPrice": 66.00,
  "expiresIn": 5000,
  "timestamp": "2025-01-28T12:05:23Z"
}
```

#### Example PRODUCTION_UPDATE Request:
```json
{
  "type": "PRODUCTION_UPDATE",
  "product": "GUACA",
  "quantity": 16
}
```

#### Example Rate Limit ERROR:
```json
{
  "type": "ERROR",
  "code": "RATE_LIMIT_EXCEEDED",
  "reason": "Too many orders per minute",
  "timestamp": "2025-01-28T12:05:30Z"
}
```

#### Example ORDER Request:
```json
{
  "type": "ORDER",
  "clOrdID": "ORD-EquipoAndromeda-1738012345-a1b2c3d4",
  "side": "BUY",
  "mode": "MARKET", 
  "product": "FOSFO",
  "qty": 10,
  "message": "Need for GUACA production 🥑"
}
```

#### Expected Trading Flow:
1. **Order Validation**: Product exists, quantity > 0, unique clOrdID
2. **Database Persistence**: Order saved with status=PENDING
3. **Market Engine**: Order sent to matching engine via channel
4. **Order Matching**: Check opposite side order book for matches
5. **Trade Execution**: If matched, atomic transaction updates both orders + creates fill
6. **Fill Broadcast**: FILL messages sent to both buyer and seller (Phase 7)

## ✨ Key Features

### 🔐 Authentication & Security
- Token-based authentication with team validation
- Per-team rate limiting (100 orders/min default)
- Production authorization validation
- Comprehensive error handling with specific error codes

### 📊 Trading Engine
- Real-time order matching with price-time priority
- Support for MARKET and LIMIT orders
- Partial fill execution
- Self-trade prevention
- Order lifecycle management (PENDING → FILLED/PARTIALLY_FILLED)

### 💾 Data Persistence
- MongoDB with replica set support
- Atomic transactions for trade execution
- Order book persistence for crash recovery
- Complete fill history tracking
- Auto-retry logic on transaction conflicts

### 📡 Real-Time Communication
- WebSocket protocol for browser compatibility
- FILL message broadcasting to both parties
- TICKER broadcasts every 5 seconds with market data
- OFFER generation for improved liquidity
- EVENT_DELTA for crash recovery and resync

### 🌐 Web Interface
- Real-time market data visualization
- Interactive trading interface
- Live order book display
- Message monitoring and error display
- Static file serving at http://localhost:9000

### 🎯 Market Features
- Best bid/ask calculations
- Mid-price discovery
- Volume tracking
- Offer acceptance flow
- Configurable offer expiration

### 🏗️ Architecture Overview

```
Web UI ──HTTP──► WebSocket Server ──► MongoDB Replica Set
               │        │                    │
    Go Clients ─WebSocket─       ▼          ▼
               │        [Static Files] [Repositories]
               ▼              │              │
        [Client Handler] ──────              ▼
               │                       [Order Book]
               ▼                            │
        [Message Router]                     ▼
               │                      [Market Engine]
               ▼                            │
        [Order Service]                      ▼
               │                      [Transaction]
               ▼                            │
         FILL Broadcast ◄───────────────────▼
```

### 🔄 Message Flow

1. **Client Connection**: WebSocket handshake and authentication
2. **Order Submission**: Validation → Persistence → Market Engine
3. **Matching**: Price-time priority algorithm with atomic execution
4. **Broadcasting**: FILL messages to both parties, TICKER to all clients
5. **State Sync**: RESYNC support for crash recovery

## ⚙️ Configuration

The server uses `config.yaml` for configuration:

```yaml
server:
  protocol: "websocket"  # Changed from "tcp"
  host: "0.0.0.0"
  port: 9000
  maxConnections: 100
  readTimeout: 30s
  writeTimeout: 10s
```

## 📚 Dependencies

- Go 1.25+
- MongoDB 7.0+ with replica set (for transactions)
- Docker & Docker Compose (for MongoDB setup)

### Go Modules
- github.com/rs/zerolog (structured logging)
- gopkg.in/yaml.v3 (configuration)
- go.mongodb.org/mongo-driver (database)
- github.com/google/uuid (ID generation)
- **github.com/gorilla/websocket (WebSocket support)** ✨

## 🗄️ Database Features

### Collections
- `teams` - Team authentication and configuration
- `orders` - Order lifecycle tracking (PENDING → FILLED)
- `fills` - Immutable trade records
- `market_state` - Real-time market data

### Indexes
- Unique indexes on apiKey, teamName, clOrdID, fillID
- Compound indexes for order matching and querying
- Performance indexes on executedAt timestamps

### Transactions
- Atomic order filling with MongoDB sessions
- Order status updates + fill creation in single transaction
- Auto-retry on transaction conflicts

## 🐳 Docker Deployment

```bash
# Build and run everything
docker-compose up --build

# Run just the database
docker-compose up -d mongodb1 mongodb2 mongodb3 mongo-init

# View logs
docker-compose logs -f exchange-server
```
