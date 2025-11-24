# 🤖 Automated Trading Client

Multi-client automated trading system for the Andorian Avocado Exchange with intelligent production algorithms and diverse trading strategies.

## ✅ Features

- **Production Algorithm**: Recursive calculation matching the student guide exactly
- **4 Trading Strategies**:
  - **Auto-Producer**: Intelligent production cycle (basic → sell → buy ingredients → premium)
  - **Market Maker**: Continuous liquidity provision with limit orders
  - **Random Trader**: Unpredictable behavior for testing student algorithms
  - **Liquidity Provider**: Fast offer acceptance for market functionality
- **Multi-Client Support**: Run multiple bots simultaneously with different strategies
- **Auto-Reconnection**: Handles disconnections gracefully
- **Resync Support**: Recovers missed events after crashes
- **Statistics**: Real-time P&L, order counts, and performance metrics

## 🚀 Quick Start

### 1. Build

```bash
cd /path/to/stock-market
go build -o bin/automated-client ./cmd/automated-client/
```

### 2. Configure

Edit `automated-clients.yaml` to configure your bots:

```yaml
server:
  host: "localhost"
  port: 9000

clients:
  - name: "auto-producer-1"
    token: "TK-PROD-001"
    species: "avocultores"
    strategy: "auto_producer"
    enabled: true
    config:
      productionInterval: "60s"
      basicProduct: "PALTA-OIL"
      premiumProduct: "GUACA"
      autoSellBasic: true
```

### 3. Run

```bash
# Basic usage
./bin/automated-client

# With custom config
./bin/automated-client --config my-config.yaml

# With statistics reporting
./bin/automated-client --stats

# With debug logging
./bin/automated-client --log-level debug
```

## 📊 Usage

### Command Line Options

```
--config string
    Path to configuration file (default "automated-clients.yaml")

--log-level string
    Log level: debug, info, warn, error (default "info")

--stats
    Show statistics every 30 seconds
```

### Example Output

```
2024-11-24T08:55:00Z INF 🤖 Starting Automated Trading Clients config=automated-clients.yaml
2024-11-24T08:55:00Z INF Configuration loaded enabledClients=4 server=localhost:9000
2024-11-24T08:55:00Z INF Starting client name=auto-producer-1 species=avocultores strategy=auto_producer
2024-11-24T08:55:01Z INF Connected to server server=localhost:9000 session=auto-producer-1
2024-11-24T08:55:01Z INF Login successful balance=10000 session=auto-producer-1 team=AutoBot1
2024-11-24T08:55:02Z INF Production system initialized branches=2 maxDepth=4 session=auto-producer-1
2024-11-24T08:55:02Z INF ✅ All clients started. Press Ctrl+C to stop.
```

### With Statistics

```bash
./bin/automated-client --stats
```

Output every 30 seconds:

```
2024-11-24T08:56:00Z INF 📊 Client Statistics connected=4 total=4
2024-11-24T08:56:00Z INF   └─ client=auto-producer-1 status="✅ Connected" balance=10338 fills=2 orders=2 pnl=3.38
2024-11-24T08:56:00Z INF   └─ client=market-maker-1 status="✅ Connected" balance=10125 fills=5 orders=12 pnl=1.25
2024-11-24T08:56:00Z INF   └─ client=random-trader-1 status="✅ Connected" balance=9987 fills=3 orders=8 pnl=-0.13
2024-11-24T08:56:00Z INF   └─ client=liquidity-provider-1 status="✅ Connected" balance=10056 fills=1 orders=0 pnl=0.56
```

## 🎯 Trading Strategies

### Auto-Producer

**Purpose**: Automated production and trading cycle

**Strategy**:
1. Try premium production (if have ingredients) → hold for good price
2. Fallback to basic production (free) → sell immediately
3. Use capital to buy ingredients from students
4. Repeat cycle

**Configuration**:
```yaml
strategy: "auto_producer"
config:
  productionInterval: "60s"        # How often to produce
  basicProduct: "PALTA-OIL"       # Product for basic production
  premiumProduct: "GUACA"         # Product for premium production
  autoSellBasic: true             # Sell basic immediately
  minProfitMargin: 0.05           # 5% minimum profit for premium sales
```

### Market Maker

**Purpose**: Provide continuous liquidity with limit orders

**Strategy**: Place buy/sell limit orders with configurable spread

**Configuration**:
```yaml
strategy: "market_maker"
config:
  spread: 0.02                    # 2% spread
  quoteSize: 100                  # Order size
  maxInventory: 1000              # Max inventory per product
  updateFrequency: "5s"           # How often to update quotes
  products:                       # Products to make markets in
    - "PALTA-OIL"
    - "FOSFO"
    - "GUACA"
```

### Random Trader

**Purpose**: Create market chaos for testing

**Strategy**: Random trades at random intervals

**Configuration**:
```yaml
strategy: "random_trader"
config:
  minInterval: "10s"              # Minimum wait between trades
  maxInterval: "30s"              # Maximum wait between trades
  orderSizeMin: 50                # Minimum order size
  orderSizeMax: 200               # Maximum order size
```

### Liquidity Provider

**Purpose**: Quickly fill student orders

**Strategy**: Accept offers with price improvement

**Configuration**:
```yaml
strategy: "liquidity_provider"
config:
  fillRate: 0.8                   # 80% acceptance rate
  responseTime: "500ms"           # Response time to offers
  priceImprovement: 0.005         # 0.5% better than market
  targetProducts:                 # Products to provide liquidity for
    - "FOSFO"
    - "PITA"
    - "NUCREM"
```

## 🔧 Configuration

### Server Configuration

```yaml
server:
  host: "localhost"               # Server hostname
  port: 9000                      # Server port
  reconnectInterval: 5s           # Wait between reconnection attempts
  maxReconnectAttempts: 10        # Max reconnection attempts
```

### Species Roles

Configure production parameters for all 12 species:

```yaml
species_roles:
  avocultores:
    branches: 2
    maxDepth: 4
    decay: 0.7651
    baseEnergy: 3.0
    levelEnergy: 2.0
  monjes:
    branches: 2
    maxDepth: 3
    decay: 0.8
    baseEnergy: 4.0
    levelEnergy: 1.5
  # ... more species
```

### Recipes

Define production recipes:

```yaml
recipes:
  GUACA:
    product: "GUACA"
    ingredients:
      FOSFO: 5
      PITA: 3
    premiumBonus: 1.30             # +30% bonus
  
  PALTA-OIL:
    product: "PALTA-OIL"
    ingredients: null              # Basic production (no ingredients)
    premiumBonus: 1.0
```

## 🧪 Testing

### Run Production Tests

```bash
go test -v ./internal/autoclient/production/...
```

### Test with Single Client

```yaml
clients:
  - name: "test-bot"
    token: "TK-TEST-001"
    species: "avocultores"
    strategy: "auto_producer"
    enabled: true
```

### Dry Run (No Real Orders)

Modify strategy to log actions without sending:

```go
// In Execute() method
log.Info().Msg("Would send order...") // Instead of actually sending
return nil, nil
```

## 📈 Production Algorithm

The recursive production algorithm matches the student guide:

```
Energy(level) = baseEnergy + levelEnergy × level
Factor(level) = decay^level × branches^level
Units(level) = Energy(level) × Factor(level)
Total = Σ Units(level) for level = 0 to maxDepth
```

**Example (Avocultores)**:
- Basic Production: **119 units**
- Premium Production (+30%): **155 units**

All tests passing ✅

## 🔄 Crash Recovery

The client supports automatic recovery:

1. **Auto-Reconnection**: Reconnects after network failures
2. **Resync**: Recovers missed events using `RESYNC` message
3. **State Preservation**: Maintains inventory and balance state

## 📊 Statistics

When running with `--stats`, you get:

- Connection status per client
- P&L percentage
- Current balance
- Orders sent / Fills received
- Error counts
- Strategy health

## 🛠️ Development

### Project Structure

```
cmd/automated-client/
  main.go                         # Entry point

internal/autoclient/
  production/
    calculator.go                 # Recursive production algorithm
    recipe.go                     # Recipe management
  config/
    config.go                     # Configuration loader
  market/
    state.go                      # Market state tracker
  strategy/
    interface.go                  # Strategy interface
    auto_producer.go              # Auto-producer strategy
    market_maker.go               # Market maker strategy
    random_trader.go              # Random trader strategy
    liquidity_provider.go         # Liquidity provider strategy
  agent/
    trading_agent.go              # Trading agent
  manager/
    session.go                    # Session management
    client_manager.go             # Multi-client orchestration
```

### Adding New Strategies

1. Implement the `Strategy` interface
2. Register in `strategy/registry.go`
3. Add to configuration file

Example:

```go
type MyStrategy struct {
    // ... fields
}

func (s *MyStrategy) Execute(ctx context.Context, state *market.MarketState) ([]*Action, error) {
    // Your trading logic here
    return actions, nil
}

// Register in registry.go
factory.Register("my_strategy", func(name string) Strategy {
    return NewMyStrategy(name)
})
```

## 🎓 For Students

This system creates realistic market conditions for testing your trading algorithms:

- **Active Market**: Always someone to trade with
- **Diverse Behaviors**: Multiple strategy types
- **Interdependence**: Bots need your products (creates demand)
- **Liquidity**: Fast order fulfillment
- **Pressure Testing**: Random and aggressive traders challenge your algorithms

## 🐛 Troubleshooting

### Connection Failed

```
ERROR Failed to connect: connection refused
```

**Solution**: Check server is running on correct host/port

### Login Failed

```
ERROR Login error: INVALID_TOKEN
```

**Solution**: Verify token in configuration matches server database

### Strategy Error

```
ERROR Strategy execution error
```

**Solution**: Check strategy configuration parameters and logs with `--log-level debug`

## 📝 License

Part of the Andorian Avocado Exchange project.

## 🤝 Contributing

This is an educational project. Improvements welcome!

---

**Built with Go 1.25** | **~2,500 lines of production code** | **All tests passing** ✅
