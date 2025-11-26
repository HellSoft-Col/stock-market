# Deployment Guide - Automated Trading Clients

## Overview

This guide covers deploying the automated trading client system for the Andorian Avocado Exchange in various environments.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Configuration](#configuration)
4. [Deployment Scenarios](#deployment-scenarios)
5. [Monitoring](#monitoring)
6. [Troubleshooting](#troubleshooting)
7. [Best Practices](#best-practices)

## Prerequisites

### System Requirements

- **Go**: 1.21 or later
- **MongoDB**: 4.4+ with replica set configured
- **Memory**: 50MB per bot (minimum)
- **CPU**: 1-2% per bot (idle), 5-10% (active trading)
- **Network**: Low latency connection to exchange server

### Required Access

- Team tokens for each bot account
- Exchange server host/port
- MongoDB connection string (for server)

## Quick Start

### 1. Build the Binary

```bash
cd /path/to/stock-market
go build -o bin/automated-client cmd/automated-client/main.go
```

### 2. Create Configuration

```bash
cp automated-clients.yaml.example automated-clients.yaml
# Edit with your settings
vim automated-clients.yaml
```

### 3. Set Up Team Tokens

```bash
# Use the seed script to create bot accounts
go run cmd/seed-teams/main.go --config bot-teams.yaml

# Or manually in MongoDB
mongosh avocado_exchange
db.teams.insertOne({
  name: "bot-producer-1",
  token: "TK-PROD-001",
  species: "avocultores",
  balance: 10000,
  ...
})
```

### 4. Run the Client

```bash
./bin/automated-client --config automated-clients.yaml --stats
```

## Configuration

### Server Connection

```yaml
server:
  host: "localhost"      # Exchange server host
  port: 9000             # Exchange server port
  reconnectInterval: 5s  # Time between reconnection attempts
  maxReconnectAttempts: 10
```

### Global Limits

```yaml
global:
  riskLimits:
    maxOrderSize: 1000
    maxPositionSize: 5000
    maxDailyLoss: 10000.0
```

### Bot Configuration

```yaml
clients:
  - name: "bot-name"
    token: "TEAM_TOKEN"      # From MongoDB
    species: "avocultores"   # Must match token
    strategy: "auto_producer"
    enabled: true
    config:
      # Strategy-specific config
      productionInterval: "60s"
      basicProduct: "PALTA-OIL"
      premiumProduct: "GUACA"
```

## Deployment Scenarios

### Scenario 1: Local Development

For testing and development:

```bash
# Terminal 1: Start MongoDB
mongod --replSet rs0

# Terminal 2: Start exchange server
./bin/server

# Terminal 3: Start bots with debug logging
./bin/automated-client --log-level debug --stats
```

### Scenario 2: Single Server Production

Deploy all components on one server:

```bash
# Install as systemd service
sudo cp automated-client.service /etc/systemd/system/
sudo systemctl enable automated-client
sudo systemctl start automated-client

# Check status
sudo systemctl status automated-client
sudo journalctl -u automated-client -f
```

Example systemd service file:

```ini
[Unit]
Description=Andorian Trading Bots
After=network.target mongodb.service

[Service]
Type=simple
User=trading
WorkingDirectory=/opt/trading-bots
ExecStart=/opt/trading-bots/bin/automated-client --config /etc/trading-bots/config.yaml
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

### Scenario 3: Docker Deployment

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o automated-client cmd/automated-client/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/automated-client /usr/local/bin/
COPY automated-clients.yaml /etc/trading/
CMD ["automated-client", "--config", "/etc/trading/automated-clients.yaml"]
```

```bash
# Build and run
docker build -t trading-bots .
docker run -d --name bots \
  -v /path/to/config.yaml:/etc/trading/automated-clients.yaml \
  trading-bots
```

### Scenario 4: Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: trading-bots
spec:
  replicas: 1
  selector:
    matchLabels:
      app: trading-bots
  template:
    metadata:
      labels:
        app: trading-bots
    spec:
      containers:
      - name: bots
        image: trading-bots:latest
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        volumeMounts:
        - name: config
          mountPath: /etc/trading
      volumes:
      - name: config
        configMap:
          name: bot-config
```

### Scenario 5: Multi-Region Deployment

For geographic distribution:

```yaml
# Region 1: US-East
bots:
  - bot-prod-us1
  - bot-mm-us1

# Region 2: EU-West  
bots:
  - bot-prod-eu1
  - bot-mm-eu1
```

Use load balancing and regional exchange servers.

## Monitoring

### Built-in Statistics

Enable with `--stats` flag:

```bash
./bin/automated-client --stats
```

Output every 30 seconds:
```
INFO Client Statistics total=6 connected=6
INFO   └─ client=auto-producer-1 status=✅ Connected pnl=125.50 balance=10125.50 orders=45 fills=42
INFO   └─ client=market-maker-1 status=✅ Connected pnl=87.25 balance=10087.25 orders=234 fills=210
...
```

### Metrics Endpoint (Future)

```go
// Expose metrics on HTTP
http.Handle("/metrics", promhttp.Handler())
```

### Key Metrics to Monitor

1. **Connection Health**
   - Connected clients
   - Reconnection attempts
   - Last heartbeat time

2. **Trading Performance**
   - Orders sent vs filled
   - Fill rate (%)
   - Average fill time
   - P&L per bot

3. **System Health**
   - Memory usage
   - CPU usage
   - Goroutine count
   - Error rate

4. **Production Metrics** (for producers)
   - Productions per hour
   - Basic vs premium ratio
   - Ingredient consumption
   - Energy utilization

### Logging

Configure log levels:

```bash
# Production: info level
./bin/automated-client --log-level info

# Debug: debug level
./bin/automated-client --log-level debug

# Quiet: warn level
./bin/automated-client --log-level warn
```

Log to file:

```bash
./bin/automated-client --log-level info 2>&1 | tee -a logs/bots.log
```

## Troubleshooting

### Connection Issues

**Problem**: "Connection refused"

**Solutions**:
```bash
# Check server is running
netstat -an | grep 9000

# Test connectivity
telnet localhost 9000

# Check firewall
sudo ufw status
```

**Problem**: "Authentication failed"

**Solutions**:
```bash
# Verify token in MongoDB
mongosh avocado_exchange
db.teams.findOne({token: "TK-PROD-001"})

# Check token matches species
# Ensure token is not expired
```

### Performance Issues

**Problem**: High CPU usage

**Solutions**:
- Reduce `updateFrequency` in strategies
- Lower `checkFrequency` for arbitrage
- Reduce number of concurrent bots
- Increase `productionInterval` for producers

**Problem**: High memory usage

**Solutions**:
- Limit `maxHistoryLength` in momentum trader
- Reduce `lookbackPeriod`
- Check for goroutine leaks: `go tool pprof`

### Trading Issues

**Problem**: Orders rejected

**Solutions**:
- Check rate limits (default: 100 orders/min)
- Verify sufficient balance
- Check inventory availability
- Review order parameters

**Problem**: No fills

**Solutions**:
- Check market has liquidity
- Review order prices
- Verify product exists
- Check if using correct order type (MARKET vs LIMIT)

### Production Issues

**Problem**: Production fails

**Solutions**:
- Verify energy available
- Check ingredients in inventory
- Confirm recipe configuration
- Check production cooldown

## Best Practices

### Configuration

1. **Start Small**: Begin with 1-2 bots, scale gradually
2. **Test First**: Run in test environment before production
3. **Monitor Closely**: Watch metrics for first 24 hours
4. **Set Limits**: Configure risk limits appropriately
5. **Version Control**: Keep configs in git

### Operations

1. **Graceful Shutdown**: Use `Ctrl+C`, not `kill -9`
2. **Regular Backups**: Backup MongoDB regularly
3. **Log Rotation**: Implement log rotation
4. **Health Checks**: Monitor connection status
5. **Alerts**: Set up alerts for errors

### Security

1. **Secure Tokens**: Never commit tokens to git
2. **Environment Variables**: Use env vars for sensitive data
3. **Network Security**: Use firewall rules
4. **Access Control**: Limit who can deploy
5. **Audit Logs**: Keep deployment logs

### Strategy Selection

1. **Diversify**: Mix different strategy types
2. **Balance**: Auto-producers + market makers + traders
3. **Risk Management**: Set appropriate position limits
4. **Monitoring**: Track strategy performance
5. **Adjust**: Tune parameters based on results

### Scaling

For large deployments (50+ bots):

1. **Split by Strategy**:
   ```bash
   # File 1: producers.yaml (10 bots)
   # File 2: market-makers.yaml (20 bots)
   # File 3: traders.yaml (30 bots)
   
   ./bin/automated-client --config producers.yaml &
   ./bin/automated-client --config market-makers.yaml &
   ./bin/automated-client --config traders.yaml &
   ```

2. **Use Multiple Servers**: Distribute load across machines

3. **Database Optimization**: Index frequently queried fields

4. **Connection Pooling**: Reuse connections efficiently

## Example Configurations

### Conservative Setup (Classroom)

```yaml
# 3 bots, low activity, safe for teaching
clients:
  - name: "demo-producer"
    strategy: "auto_producer"
    enabled: true
    
  - name: "demo-mm"
    strategy: "market_maker"
    enabled: true
    
  - name: "demo-random"
    strategy: "random_trader"
    enabled: true
```

### Aggressive Setup (Simulation)

```yaml
# 10 bots, high activity, maximum market dynamics
clients:
  - 3x auto_producer
  - 2x market_maker
  - 2x momentum_trader
  - 1x arbitrage
  - 2x random_trader
```

### Production Setup (Tournament)

```yaml
# 20 bots, balanced, competitive
clients:
  - 6x auto_producer (mixed species)
  - 4x market_maker (tight spreads)
  - 4x momentum_trader (aggressive)
  - 2x arbitrage (opportunistic)
  - 4x liquidity_provider (fast execution)
```

## Migration Guide

### From Manual Trading to Bots

1. **Backup**: Export current team data
2. **Test**: Run bots in parallel initially
3. **Monitor**: Compare bot vs manual performance
4. **Gradual**: Replace manual traders one by one
5. **Validate**: Ensure P&L matches expectations

### Upgrading Bot Version

1. **Read Changelog**: Review breaking changes
2. **Test**: Run in dev environment first
3. **Backup Config**: Save current configuration
4. **Rolling Update**: Update one bot at a time
5. **Rollback Plan**: Keep previous version binary

## Support

For issues:
1. Check logs: `tail -f logs/automated-client.log`
2. Review docs: `docs/AUTOMATED_CLIENT.md`
3. Test connection: `docs/TESTING_GUIDE.md`
4. Check configuration: `automated-clients.yaml`

## Summary Checklist

Before deploying:
- [ ] MongoDB running and accessible
- [ ] Exchange server running
- [ ] Team tokens created
- [ ] Configuration file valid
- [ ] Binary built successfully
- [ ] Test run completed
- [ ] Monitoring enabled
- [ ] Logging configured
- [ ] Alerts set up
- [ ] Backup plan ready
