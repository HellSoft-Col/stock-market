# Trading Strategy Guide

## Overview

This guide explains the 6 trading strategies available in the automated trading client system. Each strategy has different characteristics, risk profiles, and use cases.

## Strategy Comparison

| Strategy | Complexity | Risk | Capital Req | Best For |
|----------|------------|------|-------------|----------|
| Auto Producer | Medium | Low | Medium | Producers with recipes |
| Market Maker | Medium | Medium | High | Providing liquidity |
| Random Trader | Low | High | Low | Market chaos/testing |
| Liquidity Provider | Low | Medium | Medium | Fast execution |
| Momentum Trader | High | High | Medium | Trend following |
| Arbitrage | High | Medium | High | Spread capture |

## Strategy Details

### 1. Auto Producer (`auto_producer`)

**Purpose**: Automated production and trading cycle for species with production capabilities.

**How It Works**:
1. Produce basic product (free, no ingredients)
2. Sell basic product for capital
3. Buy ingredients from market
4. Produce premium product (with ingredients)
5. Sell premium at higher price
6. Repeat cycle

**Configuration**:
```yaml
strategy: "auto_producer"
config:
  productionInterval: "60s"      # Time between productions
  basicProduct: "PALTA-OIL"      # Product with no ingredients
  premiumProduct: "GUACA"        # Product requiring ingredients
  autoSellBasic: true            # Auto-sell basic products
  minProfitMargin: 0.05          # 5% minimum profit margin
```

**Best For**:
- Avocultores, Monjes, Cosechadores (any producer)
- Stable, predictable income
- Learning production mechanics

**Risk Level**: Low
- No market timing risk
- Energy management required
- Ingredient availability dependency

**Expected Returns**: 5-15% per cycle

### 2. Market Maker (`market_maker`)

**Purpose**: Provide continuous liquidity by posting bid/ask quotes.

**How It Works**:
1. Monitor market prices
2. Post buy orders below market
3. Post sell orders above market
4. Profit from bid-ask spread
5. Manage inventory risk

**Configuration**:
```yaml
strategy: "market_maker"
config:
  spread: 0.02                   # 2% spread (1% each side)
  quoteSize: 100                 # Size of each quote
  maxInventory: 1000             # Max inventory per product
  updateFrequency: "5s"          # Quote update frequency
  products:
    - "PALTA-OIL"
    - "FOSFO"
    - "GUACA"
```

**Best For**:
- High-volume markets
- Stable prices
- Generating consistent small profits

**Risk Level**: Medium
- Inventory risk (holding products)
- Adverse selection risk
- Capital intensive

**Expected Returns**: 1-3% per trade, high frequency

### 3. Random Trader (`random_trader`)

**Purpose**: Create market chaos and test system resilience.

**How It Works**:
1. Wait random interval
2. Choose random product
3. Random buy/sell direction
4. Random order size
5. Repeat indefinitely

**Configuration**:
```yaml
strategy: "random_trader"
config:
  minInterval: "10s"             # Min time between orders
  maxInterval: "30s"             # Max time between orders
  orderSizeMin: 50               # Min order size
  orderSizeMax: 200              # Max order size
```

**Best For**:
- Testing systems
- Creating market volatility
- Filling classroom environments

**Risk Level**: High
- No strategy logic
- Purely random
- Educational tool

**Expected Returns**: Random (likely negative)

### 4. Liquidity Provider (`liquidity_provider`)

**Purpose**: Quickly accept market offers to facilitate trades.

**How It Works**:
1. Monitor incoming offers
2. Evaluate offer quality
3. Accept profitable offers fast
4. Build inventory/capital

**Configuration**:
```yaml
strategy: "liquidity_provider"
config:
  fillRate: 0.8                  # Accept 80% of good offers
  responseTime: "500ms"          # Response speed
  priceImprovement: 0.005        # 0.5% better than market
  targetProducts:
    - "FOSFO"
    - "PITA"
    - "NUCREM"
```

**Best For**:
- High-activity markets
- Quick profits
- Offer-based trading

**Risk Level**: Medium
- Fast decision making
- Inventory accumulation
- Requires good pricing

**Expected Returns**: 0.5-2% per offer

### 5. Momentum Trader (`momentum_trader`)

**Purpose**: Follow price trends - buy rising, sell falling.

**How It Works**:
1. Track price history (lookback period)
2. Calculate momentum (% change)
3. Buy if momentum > threshold (uptrend)
4. Sell if momentum < -threshold (downtrend)
5. Use stop loss and take profit

**Configuration**:
```yaml
strategy: "momentum_trader"
config:
  products:
    - "PALTA-OIL"
    - "GUACA"
  lookbackPeriod: "5m"           # How far back to analyze
  momentumThreshold: 0.03        # 3% change triggers trade
  maxPosition: 500               # Max inventory per product
  positionSize: 100              # Trade size
  updateFrequency: "10s"         # Analysis frequency
  stopLoss: 0.10                 # 10% stop loss
  takeProfit: 0.15               # 15% take profit
```

**How Momentum is Calculated**:
```
momentum = (current_price - old_price) / old_price

If momentum > 0.03:  BUY  (price rising 3%+)
If momentum < -0.03: SELL (price falling 3%+)
```

**Best For**:
- Trending markets
- Volatile products
- Directional betting

**Risk Level**: High
- Trend reversal risk
- Stop loss protection needed
- Can chase bad moves

**Expected Returns**: -10% to +25% (high variance)

### 6. Arbitrage (`arbitrage`)

**Purpose**: Exploit price inefficiencies and spreads.

**How It Works**:
1. **Spread Arbitrage**: 
   - Buy at bid, sell at ask
   - Capture bid-ask spread
   
2. **Production Arbitrage**:
   - Calculate production cost
   - Compare to market price
   - Produce if profitable
   
3. **Cross-Product Arbitrage**:
   - Compare ingredient costs
   - Calculate product value
   - Trade discrepancies

**Configuration**:
```yaml
strategy: "arbitrage"
config:
  minSpread: 0.05                # 5% minimum spread
  positionSize: 50               # Trade size
  maxPositions: 5                # Concurrent positions
  checkFrequency: "5s"           # Opportunity scan rate
```

**Arbitrage Examples**:

Example 1 - Spread Arbitrage:
```
FOSFO: Bid=100, Ask=110
Spread = (110-100)/100 = 10%

If spread > minSpread (5%):
  Buy at 100, Sell at 110
  Profit = 10 per unit
```

Example 2 - Production Arbitrage:
```
GUACA recipe requires:
  FOSFO: 5 units @ 100 = 500
  PITA: 3 units @ 50 = 150
  Total cost = 650

GUACA market price = 800

Profit = 800 - 650 = 150 (23%)
Action: Buy ingredients, produce, sell
```

**Best For**:
- Inefficient markets
- Multiple products
- Mathematical traders

**Risk Level**: Medium
- Execution risk (prices move)
- Capital intensive
- Competition from others

**Expected Returns**: 2-8% per opportunity

## Strategy Selection Guide

### For Beginners

Start with:
1. **Auto Producer** - Safe, predictable
2. **Liquidity Provider** - Simple logic
3. **Random Trader** - Observe market dynamics

### For Intermediate

Add:
4. **Market Maker** - Learn market making
5. **Momentum Trader** - Understand trends

### For Advanced

Complete with:
6. **Arbitrage** - Complex but profitable

## Mixed Strategy Deployment

Recommended bot mix for balanced market:

```yaml
# Production Base (30%)
- 3x Auto Producer (different species)

# Liquidity Layer (40%)
- 2x Market Maker (tight spreads)
- 2x Liquidity Provider (fast fills)

# Trading Layer (30%)
- 2x Momentum Trader (trend followers)
- 1x Arbitrage (opportunistic)
- 1x Random Trader (chaos agent)
```

## Performance Tuning

### Auto Producer

Optimize:
- `productionInterval`: Match energy regeneration
- `minProfitMargin`: Balance speed vs profit
- Ingredient buy logic: Time purchases wisely

### Market Maker

Optimize:
- `spread`: Wider = more profit, less fills
- `quoteSize`: Match market depth
- `maxInventory`: Control risk exposure

### Momentum Trader

Optimize:
- `lookbackPeriod`: Shorter = more reactive
- `momentumThreshold`: Higher = fewer trades
- `stopLoss`: Protect downside
- `takeProfit`: Lock in gains

### Arbitrage

Optimize:
- `minSpread`: Higher = fewer but better opportunities
- `checkFrequency`: Balance speed vs CPU
- Product selection: Focus on liquid products

## Common Pitfalls

### Auto Producer
- ❌ Producing without checking ingredient prices
- ✅ Calculate profitability before premium production

### Market Maker
- ❌ Posting quotes without checking competition
- ✅ Adjust spread based on market conditions

### Momentum Trader
- ❌ Chasing every small move
- ✅ Wait for clear momentum signals
- ❌ No stop loss
- ✅ Always use stop loss

### Arbitrage
- ❌ Assuming prices won't move
- ✅ Execute both legs quickly
- ❌ Ignoring fees
- ✅ Calculate net profit after fees

## Backtesting Strategies

Before deploying:

1. **Paper Trading**: Run in test mode first
2. **Small Size**: Start with minimum position sizes
3. **Monitor**: Watch closely for 1 hour
4. **Analyze**: Review P&L, fill rate, errors
5. **Tune**: Adjust parameters based on results
6. **Scale**: Gradually increase size

## Strategy Metrics

Track these KPIs for each strategy:

### Common Metrics
- Win rate (%)
- Average profit per trade
- Maximum drawdown
- Sharpe ratio
- Total P&L

### Strategy-Specific

**Auto Producer**:
- Productions per hour
- Basic vs premium ratio
- Ingredient cost efficiency

**Market Maker**:
- Spread captured
- Inventory turnover
- Quote hit rate

**Momentum Trader**:
- Trend accuracy
- Stop loss frequency
- Average hold time

**Arbitrage**:
- Opportunities found
- Execution speed
- Spread captured

## Advanced Techniques

### Dynamic Parameter Adjustment

Adapt to market conditions:

```go
// Pseudo-code
if market.volatility > HIGH {
    momentum.stopLoss = 0.15  // Wider stop loss
    market_maker.spread = 0.03  // Wider spread
}

if market.volume < LOW {
    arbitrage.minSpread = 0.10  // Higher threshold
    liquidity_provider.fillRate = 0.5  // More selective
}
```

### Multi-Strategy Agents

Combine strategies:

```yaml
# Agent that produces AND trades
- Production: auto_producer (background)
- Trading: momentum_trader (opportunistic)
```

### Adaptive Strategies

Learn from history:

1. Track success rate by time of day
2. Adjust aggressiveness accordingly
3. Reduce activity during losses
4. Increase during winning streaks

## Conclusion

Each strategy serves different purposes:

- **Auto Producer**: Steady income, production-focused
- **Market Maker**: Liquidity provision, spread capture
- **Random Trader**: Market dynamics, testing
- **Liquidity Provider**: Fast execution, offer response
- **Momentum Trader**: Trend following, directional
- **Arbitrage**: Efficiency hunting, mathematical

Best results come from:
1. Understanding each strategy deeply
2. Matching strategy to market conditions
3. Proper risk management
4. Continuous monitoring and adjustment

Start simple, scale gradually, and always monitor performance!
