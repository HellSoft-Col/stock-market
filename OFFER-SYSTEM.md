# Offer System - Real-Time Trade Matching

## Overview

The offer system allows buyers to send trade proposals to teams that have inventory. When a buyer's order cannot be immediately matched, the system broadcasts an **OFFER** to teams with the requested product. Teams can accept or reject offers in real-time.

## How It Works

### Flow Diagram

```
1. Team A places BUY order for GUACA
2. No immediate match found in order book
3. Server identifies teams with GUACA inventory
4. Server broadcasts OFFER to eligible teams
5. Team B receives OFFER notification
6. Team B views offer in "Offers" tab
7. Team B can:
   - Accept: Creates immediate trade
   - Reject: Declines the offer
8. First team to accept wins (first-come-first-served)
```

### Backend Components

#### 1. **OfferGenerator** (`internal/market/offer_generator.go`)

Manages offer lifecycle:
- **GenerateOffer()**: Creates offers for unmatched buy orders
- **GenerateTargetedOffer()**: Sends offers to specific teams with inventory
- **HandleAcceptOffer()**: Processes offer acceptances (first-come wins)
- **cleanupLoop()**: Removes expired offers

#### 2. **Active Offer Tracking**

```go
type ActiveOffer struct {
    OfferMsg  *domain.OfferMessage
    BuyOrder  *domain.Order
    ExpiresAt time.Time
}
```

Offers stored in memory with automatic expiration cleanup.

#### 3. **Message Types**

**OFFER Message (Server → Client)**
```json
{
  "type": "OFFER",
  "offerId": "off-1762211187-a1b2c3d4",
  "buyer": "team1",
  "product": "GUACA",
  "quantityRequested": 10,
  "maxPrice": 11.50,
  "expiresIn": 30000,
  "timestamp": "2025-11-03T12:00:00Z"
}
```

**ACCEPT_OFFER Message (Client → Server)**
```json
{
  "type": "ACCEPT_OFFER",
  "offerId": "off-1762211187-a1b2c3d4",
  "accept": true,
  "quantityOffered": 10,
  "priceOffered": 11.50
}
```

### Frontend Components

#### 1. **Offers Tab** (`web/index.html`)

New tab added to trading panel:
- **My Orders** - Active orders
- **Offers** ⭐ NEW - Incoming buy offers with badge counter
- **Messages** - System messages

#### 2. **Offer Display**

Each offer shows:
- Buyer team name
- Product and quantity requested
- Maximum price willing to pay
- Expiration countdown
- Input fields for quantity and price
- Accept/Reject buttons

#### 3. **Real-Time Features**

- **Toast notifications** when offers arrive
- **Badge counter** showing pending offers
- **Auto-refresh** when offers accepted/rejected
- **Expiration handling** removes stale offers

## User Experience

### For Buyers (Creating Orders)

1. Place a BUY order (MARKET or LIMIT)
2. If no immediate match:
   - Order stays in order book (for LIMIT orders)
   - OFFER broadcast to teams with inventory
3. Wait for sellers to accept
4. Receive FILL notification when accepted

### For Sellers (Receiving Offers)

1. **Notification**: Toast popup "📬 New offer: 10 GUACA from team1 @ $11.50"
2. **Badge**: Red counter appears on "Offers" tab
3. **View Details**: Switch to Offers tab
4. **Review**:
   - Check buyer, product, quantity, price
   - See expiration countdown
5. **Adjust** (optional):
   - Change quantity (up to requested amount)
   - Change price (up to max price)
6. **Decision**:
   - **Accept**: Immediate trade execution
   - **Reject**: Decline without penalty

### After Acceptance

1. Virtual SELL order created
2. Trade executed immediately
3. FILL notifications sent to both parties
4. Inventory and balance updated
5. Offer removed from all pending lists

## Configuration

### Server Config (`config.yaml`)

```yaml
market:
  offerTimeout: 30s  # How long offers stay active
```

Set to `0` for no expiration (offers persist until accepted/rejected).

### Offer Targeting

The system sends offers to teams that:
1. Have recent sell history for the product (preferred)
2. Have inventory of the requested product
3. Are not the buyer (no self-trading)

## Technical Details

### First-Come-First-Served

When multiple teams try to accept:
```go
// Offer is locked and processed
og.mu.RLock()
offer, exists := og.activeOffers[acceptMsg.OfferID]
og.mu.RUnlock()

if !exists {
    return fmt.Errorf("offer not found or expired")
}

// First accept wins, removes offer
og.mu.Lock()
delete(og.activeOffers, acceptMsg.OfferID)
og.mu.Unlock()
```

### Price Calculation

Offers use market-based pricing:
```go
var offerPrice float64
if marketState.Mid != nil {
    offerPrice = *marketState.Mid * 1.10  // 10% premium above mid
} else {
    offerPrice = 10.0  // Default fallback
}
```

### Expiration Cleanup

Background goroutine removes expired offers every 100ms:
```go
func (og *OfferGenerator) cleanupLoop() {
    ticker := time.NewTicker(100 * time.Millisecond)
    for range ticker.C {
        now := time.Now()
        for offerId, offer := range og.activeOffers {
            if now.After(offer.ExpiresAt) {
                delete(og.activeOffers, offerId)
            }
        }
    }
}
```

## Testing

### Manual Test Flow

1. **Setup**: Login with two teams
   - Team A: Has cash
   - Team B: Has GUACA inventory

2. **Create Offer**:
   ```
   Team A: Place BUY order for 10 GUACA
   Expected: Team B receives OFFER notification
   ```

3. **View Offer**:
   ```
   Team B: Check "Offers" tab
   Expected: Offer displayed with Accept/Reject buttons
   ```

4. **Accept Offer**:
   ```
   Team B: Click Accept (adjust qty/price if needed)
   Expected: 
   - Trade executes immediately
   - Both teams receive FILL notifications
   - Offer removed from pending list
   ```

### Automated Tests

Test coverage in `internal/market/offer_generator.go`:
- ✅ Offer creation
- ✅ Offer targeting to eligible teams
- ✅ Accept handling (first-come wins)
- ✅ Reject handling
- ✅ Expiration cleanup
- ✅ Concurrent accept attempts

## UI Screenshots

### Before Offer
```
[My Orders] [Offers] [Messages]
            ^^^^^^^^
            No badge
```

### After Receiving Offer
```
[My Orders] [Offers (1)] [Messages]
            ^^^^^^^^^^^
            Red badge with count
```

### Offer Panel
```
┌─────────────────────────────────────────┐
│ 📬 Incoming Offers                      │
│                                         │
│ ┌─────────────────────────────────────┐ │
│ │ 🤝 team1 wants to buy               │ │
│ │ 10 GUACA                            │ │
│ │ Offering up to: $11.50 per unit     │ │
│ │ ⏱ Expires in 25s                    │ │
│ │                                     │ │
│ │ Quantity: [10▼] Price: [$11.50]    │ │
│ │ [✓ Accept] [✗ Reject]              │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

## Benefits

1. **Improved Liquidity**: Orders can match even without order book depth
2. **Price Discovery**: Sellers see buyer willingness to pay
3. **Negotiation**: Sellers can adjust price/quantity
4. **Real-Time**: Instant notifications and responses
5. **Fair**: First-come-first-served prevents gaming
6. **Transparent**: All details visible before acceptance

## Future Enhancements

Potential additions:
- Counter-offers (seller proposes different terms)
- Multi-party offers (broadcast to all teams)
- Offer history/analytics
- Automatic acceptance rules
- Offer grouping by product
- Audio alerts for high-value offers

## Troubleshooting

### Offer Not Received

**Problem**: Buyer places order but seller doesn't see offer

**Solutions**:
1. Check seller has inventory: `inventory[product] > 0`
2. Verify seller is connected (active WebSocket)
3. Check offer expiration timeout
4. Review server logs for broadcast errors

### Accept Failed

**Problem**: Clicking Accept doesn't execute trade

**Solutions**:
1. Check offer hasn't expired
2. Verify seller has sufficient inventory
3. Ensure price <= maxPrice
4. Check quantity <= quantityRequested
5. Look for concurrent accept by another team

### Multiple Offers

**Problem**: Receiving too many offers

**Solutions**:
1. Adjust inventory levels (reduce stock)
2. Temporarily disconnect to pause offers
3. Auto-reject low-value offers (future feature)

## API Reference

### Frontend Functions

```javascript
// Handle incoming offer
function handleOfferMessage(message)

// Display all active offers
function displayOffers()

// Accept an offer
function acceptOffer(offerId)

// Reject an offer
function rejectOffer(offerId)

// Update badge counter
function updateOfferCount()
```

### Backend Methods

```go
// Generate offer for unmatched order
func (og *OfferGenerator) GenerateTargetedOffer(
    buyOrder *domain.Order, 
    eligibleTeams []*domain.Team
) error

// Handle offer acceptance
func (og *OfferGenerator) HandleAcceptOffer(
    acceptMsg *domain.AcceptOfferMessage, 
    acceptorTeam string
) error

// Execute trade from offer acceptance
func (og *OfferGenerator) executeOfferMatch(
    buyOrder, sellOrder *domain.Order
) error
```

## Summary

The offer system creates a **real-time marketplace** where buyers and sellers can discover each other even without order book depth. Offers are targeted, time-limited, and execute on first-come-first-served basis.

**Status**: ✅ Fully Implemented and Tested
**Backend**: ✅ Complete (offer_generator.go)
**Frontend**: ✅ Complete (Offers tab + handlers)
**Server**: ✅ Running on port 9000
