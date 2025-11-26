# Automated Trading Client Testing Guide

## Prerequisites

Before testing the automated trading client, ensure you have:

1. **MongoDB Running** - The server requires MongoDB replica set
   ```bash
   # Start MongoDB replica set (if using Docker)
   docker-compose up -d mongodb
   
   # Or start local MongoDB
   mongod --replSet rs0 --port 27017
   ```

2. **Server Binary** - Build or use existing server binary
   ```bash
   # Build the server
   go build -o bin/server cmd/exchange-server/main.go
   ```

3. **Team Tokens** - Create test team accounts in MongoDB
   ```bash
   # Use the seed script or create manually
   go run cmd/seed/main.go
   
   # Or use Python script
   python scripts/generate-team-tokens.py
   ```

## Test Scenario 1: Basic Functionality Test

### Step 1: Start the Server

```bash
# Start the exchange server
./bin/server
```

Expected output:
```
INFO  Starting Andorian Avocado Exchange Server
INFO  MongoDB connected
INFO  Server listening on :9000
```

### Step 2: Configure Test Clients

Edit `automated-clients.yaml` to enable a single test client:

```yaml
clients:
  - name: "test-producer-1"
    token: "YOUR_TEST_TOKEN_HERE"
    species: "avocultores"
    strategy: "auto_producer"
    enabled: true
    config:
      productionInterval: "30s"
      basicProduct: "PALTA-OIL"
      premiumProduct: "GUACA"
      autoSellBasic: true
      minProfitMargin: 0.05
```

### Step 3: Start the Automated Client

```bash
# Run with statistics enabled
./bin/automated-client --stats --log-level debug
```

Expected output:
```
INFO  🤖 Starting Automated Trading Clients
INFO  Configuration loaded enabledClients=1 server=localhost:9000
INFO  ✅ All clients started. Press Ctrl+C to stop.
INFO  📊 Client Statistics total=1 connected=1
INFO    └─ client=test-producer-1 status=✅ Connected pnl=0 balance=10000 orders=0 fills=0
```

### Step 4: Monitor Activity

Watch for these events in the logs:
- ✅ Connection established
- 🔐 Authentication successful
- 🏭 Production started
- 📦 Orders placed
- 💰 Fills received
- 📊 Statistics updates

## Test Scenario 2: Multiple Strategy Test

Enable multiple clients with different strategies:

```bash
# Edit automated-clients.yaml to enable:
# - auto_producer (Avocultores)
# - market_maker (Herreros)
# - random_trader (Cosechadores)
# - liquidity_provider (Extractores)

./bin/automated-client --stats
```

Expected behavior:
- **Auto Producer**: Produces PALTA-OIL, sells it, buys ingredients, produces GUACA
- **Market Maker**: Posts bid/ask orders continuously
- **Random Trader**: Places random orders at irregular intervals
- **Liquidity Provider**: Accepts offers from the market quickly

## Test Scenario 3: Stress Test

Test system resilience under load:

```bash
# Enable 10+ clients simultaneously
# Monitor for:
# - Connection stability
# - Order rate limits
# - Memory usage
# - CPU usage

# Run with debug logging
./bin/automated-client --stats --log-level debug

# In another terminal, monitor system resources
top -pid $(pgrep automated-client)
```

## Test Scenario 4: Reconnection Test

Test auto-reconnection logic:

1. Start automated client
2. Stop the server (Ctrl+C)
3. Observe reconnection attempts in logs:
   ```
   WARN  Connection lost, attempting reconnection...
   INFO  Reconnection attempt 1/10
   ```
4. Restart the server
5. Verify clients reconnect automatically:
   ```
   INFO  Reconnection successful
   INFO  Re-authenticating...
   INFO  ✅ Connected
   ```

## Test Scenario 5: Resync Test

Test crash recovery using RESYNC:

1. Start automated client and let it trade
2. Force kill the client (kill -9)
3. Restart the client
4. Verify it uses RESYNC to recover state:
   ```
   INFO  Requesting state resync...
   INFO  RESYNC response received
   INFO  Market state restored
   ```

## Verification Checklist

### Connection & Authentication
- [ ] Client connects to server successfully
- [ ] Authentication with team token works
- [ ] Role and recipe data received from server
- [ ] Graceful shutdown on Ctrl+C

### Production System
- [ ] Basic production calculates correct units
- [ ] Premium production requires correct ingredients
- [ ] Production bonus applied correctly (30%)
- [ ] Energy consumption tracked properly

### Trading Strategies
- [ ] **Auto Producer**: Completes full production cycle
- [ ] **Market Maker**: Posts continuous two-sided quotes
- [ ] **Random Trader**: Creates market volatility
- [ ] **Liquidity Provider**: Accepts offers rapidly

### Order Management
- [ ] Orders sent with valid format
- [ ] Fill messages received and processed
- [ ] Inventory updated on fills
- [ ] Balance updated correctly

### Error Handling
- [ ] Handles invalid orders gracefully
- [ ] Survives network disconnections
- [ ] Rate limiting respected
- [ ] Insufficient inventory handled

### Performance
- [ ] Low CPU usage (<5% per client)
- [ ] Low memory usage (<50MB per client)
- [ ] No goroutine leaks
- [ ] Fast message processing (<10ms)

## Performance Benchmarks

Target metrics per client:
- **CPU Usage**: < 5%
- **Memory Usage**: < 50 MB
- **Message Latency**: < 10ms avg
- **Orders Per Second**: 5-10 (market maker)
- **Reconnection Time**: < 5 seconds

## Debugging Tips

### Enable Debug Logging
```bash
./bin/automated-client --log-level debug
```

### Monitor Specific Client
```bash
./bin/automated-client --log-level debug 2>&1 | grep "client-name"
```

### Check MongoDB State
```bash
mongosh avocado_exchange
db.teams.find({name: "test-producer-1"})
db.orders.find({teamId: "..."})
db.fills.find({teamId: "..."})
```

### Server-Side Logs
Check server logs for errors:
```bash
tail -f logs/server.log
```

## Known Issues & Limitations

1. **MongoDB Required**: Server requires MongoDB with replica set
2. **Team Tokens**: Must be pre-created in database
3. **Rate Limits**: Default 100 orders/min per team
4. **Network Timeout**: 60s read timeout
5. **Max Clients**: Tested up to 100 concurrent clients

## Next Steps After Testing

1. ✅ Verify all test scenarios pass
2. Document any issues found
3. Tune configuration parameters
4. Add more sophisticated strategies
5. Implement performance monitoring
6. Create integration tests
7. Deploy to production environment

## Quick Test Command

For a quick smoke test:

```bash
# Terminal 1: Start server
./bin/server

# Terminal 2: Start one test client
./bin/automated-client --stats --log-level info

# Terminal 3: Monitor logs
tail -f logs/automated-client.log
```

## Troubleshooting

### "Connection refused"
- Verify server is running on port 9000
- Check firewall settings
- Verify host/port in config

### "Authentication failed"
- Verify team token exists in database
- Check token format in config
- Ensure team has correct permissions

### "Rate limit exceeded"
- Reduce order frequency in strategy config
- Check server rate limit settings
- Monitor orders per minute

### "Insufficient inventory"
- Check production is running
- Verify ingredient availability
- Review buy order logic

## Test Data

Sample team tokens for testing (create these in MongoDB):
```
TK-PROD-001 - Avocultores (Auto Producer)
TK-MM-001   - Herreros (Market Maker)
TK-RAND-001 - Cosechadores (Random Trader)
TK-LP-001   - Extractores (Liquidity Provider)
```

Each team should start with:
- Balance: 10,000 credits
- Energy: Full (based on species)
- Inventory: Empty (or basic starter items)
