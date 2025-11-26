# Test Configuration Summary

## 12 Teams with Diverse Strategies

| # | Team Name | Token | Strategy | Primary Product | Config Highlights |
|---|-----------|-------|----------|-----------------|-------------------|
| 1 | **Alquimistas de Palta** | TK-09jK... | auto_producer | PALTA-OIL | 45s production, 8% margin |
| 2 | **Arpistas de Pita-Pita** | TK-NVUO... | market_maker | PITA | 3% spread, 80 size, 4s updates |
| 3 | **Avocultores del Hueso Cósmico** | TK-Xqno... | hybrid | PALTA-OIL | All-in-one, 40s production |
| 4 | **Cartógrafos de Fosfolima** | TK-egak... | momentum_trader | FOSFO | 4m lookback, 2.5% threshold |
| 5 | **Cosechadores de Semillas** | TK-QJ3a... | arbitrage | PITA | 4% spread, 4s checks |
| 6 | **Forjadores Holográficos** | TK-B8Yd... | random_trader | H-GUACA | 8-25s chaos, 40-180 size |
| 7 | **Ingenieros Holo-Aguacate** | TK-SO4U... | liquidity_provider | H-GUACA | 85% fill, 400ms response |
| 8 | **Mensajeros del Núcleo** | TK-XKAG... | deepseek (AI) | NUCREM | AI-powered, 20s decisions |
| 9 | **Mineros de Guacatrones** | TK-easO... | auto_producer | H-GUACA | Aggressive 35s, 6% margin |
| 10 | **Monjes del Guacamole Estelar** | TK-3h1J... | buffett (AI) | FOSFO | Value investing, 35% margin |
| 11 | **Orfebres de Cáscara** | TK-6fAQ... | market_maker | FOSFO | Conservative 4% spread |
| 12 | **Someliers de Aceite** | TK-TUkY... | hybrid | PALTA-OIL | Production focus, 30s cycles |

## Strategy Distribution

- **auto_producer**: 2 teams (Alquimistas, Mineros)
- **market_maker**: 2 teams (Arpistas, Orfebres)
- **hybrid**: 2 teams (Avocultores, Someliers)
- **momentum_trader**: 1 team (Cartógrafos)
- **arbitrage**: 1 team (Cosechadores)
- **random_trader**: 1 team (Forjadores)
- **liquidity_provider**: 1 team (Ingenieros)
- **deepseek (AI)**: 1 team (Mensajeros)
- **buffett (AI)**: 1 team (Monjes)

## How to Run Test

```bash
# 1. Ensure .env has your DeepSeek API key
cat .env | grep DEEPSEEK_API_KEY

# 2. Run automated client
./bin/automated-client -config automated-clients.yaml

# 3. Monitor in another terminal
./bin/trading-cli status --watch
./bin/trading-cli pnl --detailed
```

## Expected Behavior

Each team will trade according to its strategy:
- **Producers** will generate products and sell them
- **Market makers** will provide liquidity with continuous quotes
- **Hybrid** teams will do everything simultaneously
- **AI teams** will make intelligent decisions every 20-35 seconds
- **Random trader** will create market chaos
- **Arbitrage** will exploit price differences

## Files Created

- ✅ `.env` - Contains real API key (git ignored)
- ✅ `.env.sample` - Template for others (git tracked)
- ✅ `automated-clients.yaml` - 12 team configuration (git ignored)
- ✅ `.gitignore` - Protection rules

All sensitive data is protected! 🔒
