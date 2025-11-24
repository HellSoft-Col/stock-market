# How Teams Can Fill Orders - Complete Guide

## Investigation Date
November 24, 2025

## Executive Summary
Teams can fill orders from other teams through **THREE MECHANISMS**:
1. **Placing SELL orders** that match existing BUY orders (instant match)
2. **Accepting OFFERS** from buyers (interactive negotiation)
3. **Automated responses** using trading bots/strategies

---

## MECHANISM 1: Direct SELL Orders (Immediate Matching)

### How It Works

When Team A places a **BUY order**, the system:
1. Tries to match with existing SELL orders immediately
2. If no match found → adds BUY order to order book + generates offers

When Team B places a **SELL order**, the system:
1. **Checks Team B has inventory** (cannot sell what you don't have)
2. Scans existing BUY orders in the order book
3. If prices match → **INSTANT TRADE** executed
4. If no match → SELL order added to order book (waits for buyers)

### Example Flow

```
TIME | ACTION | ORDER BOOK
-----|--------|------------
10:00 | Team A places: BUY 10 GUACA @ $12.00 (LIMIT)
      | → No SELL orders available
      | → Added to order book
      |
      | ORDER BOOK:
      | BUY: Team A - 10 GUACA @ $12.00
      | SELL: (empty)

10:05 | Team B places: SELL 10 GUACA @ $11.00 (LIMIT)
      | → System checks: Team B has 10 GUACA? YES ✅
      | → System scans BUY orders
      | → Found: Team A willing to pay $12.00
      | → Price check: $11.00 <= $12.00? YES ✅
      | 
      | ⚡ INSTANT MATCH EXECUTED
      | Trade Price: $11.00 (seller's price wins)
      | Trade Qty: 10
      |
      | ATOMIC TRANSACTION:
      | 1. Team A: GUACA +10, Balance -$110
      | 2. Team B: GUACA -10, Balance +$110
      | 3. Both orders marked FILLED
      | 4. Removed from order book
      |
      | ORDER BOOK:
      | BUY: (empty)
      | SELL: (empty)
```

### Code Reference

**File**: `internal/market/matcher.go` (line 264-329)

```go
func (m *Matcher) processSellOrder(sellOrder *domain.Order) (*MatchResult, error) {
    // 1. VALIDATE INVENTORY FIRST
    if m.inventoryService != nil {
        canSell, err := m.inventoryService.CanSell(
            context.Background(),
            sellOrder.TeamName,
            sellOrder.Product,
            sellOrder.Quantity,
        )
        if !canSell {
            return nil, fmt.Errorf("insufficient inventory for sell order")
        }
    }
    
    // 2. SCAN EXISTING BUY ORDERS
    buyOrders := m.orderBook.GetBuyOrders(sellOrder.Product)
    
    for _, buyOrder := range buyOrders {
        // 3. CHECK IF CAN MATCH
        if !m.canMatch(buyOrder, sellOrder) {
            continue
        }
        
        // 4. CREATE MATCH RESULT
        return &MatchResult{
            Matched:    true,
            BuyOrder:   buyOrder,
            SellOrder:  sellOrder,
            TradePrice: tradePrice,
            TradeQty:   tradeQty,
        }, nil
    }
    
    // 5. NO MATCH - ADD TO BOOK
    m.orderBook.AddOrder(sellOrder.Product, sellOrder.Side, sellOrder)
    return &MatchResult{Matched: false}, nil
}
```

### Matching Rules

**File**: `internal/market/matcher.go` (line 331-344)

```go
func (m *Matcher) canMatch(buyOrder, sellOrder *domain.Order) bool {
    // Rule 1: Same team cannot trade with itself
    if buyOrder.TeamName == sellOrder.TeamName {
        return false
    }
    
    // Rule 2: LIMIT vs LIMIT - buy price must be >= sell price
    if buyOrder.Mode == "LIMIT" && sellOrder.Mode == "LIMIT" {
        return *buyOrder.LimitPrice >= *sellOrder.LimitPrice
    }
    
    // Rule 3: All other combinations match
    // - MARKET vs LIMIT
    // - MARKET vs MARKET
    return true
}
```

### UI: How to Place SELL Order

**Web Interface**:
1. Go to "Trading" tab
2. Select "SELL" side
3. Choose product (e.g., GUACA)
4. Enter quantity
5. Choose mode:
   - **MARKET**: Sells at best available buy price
   - **LIMIT**: Sets minimum price you'll accept
6. Click "Submit Order"
7. If there's a matching BUY order → **Instant trade!**

**Message Format**:
```json
{
  "type": "ORDER",
  "clOrdID": "ORD-TeamB-1732492800-abc123",
  "side": "SELL",
  "mode": "LIMIT",
  "product": "GUACA",
  "qty": 10,
  "limitPrice": 11.00,
  "message": "💰 Fresh avocados!"
}
```

---

## MECHANISM 2: Accepting OFFERS (Interactive System)

### What are OFFERS?

When a BUY order **cannot match immediately**, the system generates an **OFFER** message and sends it to teams that have the required product in inventory.

**OFFERS are targeted** - only teams with sufficient inventory receive them!

### The Offer Flow

```
┌─────────────────────────────────────────────┐
│ 1. Team A places BUY order                  │
│    - BUY 15 GUACA @ MARKET                  │
└─────────────┬───────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────┐
│ 2. System checks for immediate match        │
│    - No SELL orders in book                 │
│    - Order added to book                    │
└─────────────┬───────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────┐
│ 3. System queries inventory                 │
│    "Who has >= 15 GUACA?"                   │
│                                             │
│    Results:                                 │
│    ✅ Team B: 20 GUACA                       │
│    ✅ Team C: 30 GUACA                       │
│    ❌ Team D: 5 GUACA (not enough)          │
│    ❌ Team E: 0 GUACA                        │
└─────────────┬───────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────┐
│ 4. Generate OFFER message                   │
│    - OfferID: "off-1732492800-xyz789"       │
│    - Buyer: Team A                          │
│    - Product: GUACA                         │
│    - Quantity: 15                           │
│    - MaxPrice: $11.00 (10% above mid)       │
│    - Expires: 30 seconds                    │
└─────────────┬───────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────┐
│ 5. Broadcast OFFER to eligible teams        │
│                                             │
│    Team B receives: 📬 OFFER                 │
│    Team C receives: 📬 OFFER                 │
│    Team D: (no offer - insufficient inv)    │
│    Team E: (no offer - no inventory)        │
└─────────────┬───────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────┐
│ 6. Teams can ACCEPT or REJECT               │
│                                             │
│    Team B's Options:                        │
│    - Accept: Sell 15 GUACA @ $11.00         │
│    - Accept: Sell 10 GUACA @ $10.50         │
│    - Reject: No thanks                      │
│                                             │
│    Team C's Options:                        │
│    - Accept: Sell 15 GUACA @ $11.00         │
│    - Reject: Waiting for higher price       │
└─────────────┬───────────────────────────────┘
              │
              ▼ (Team B accepts first)
┌─────────────────────────────────────────────┐
│ 7. ACCEPT_OFFER received                    │
│    - OfferID: "off-1732492800-xyz789"       │
│    - Acceptor: Team B                       │
│    - Quantity: 15 GUACA                     │
│    - Price: $10.80 (negotiated)             │
└─────────────┬───────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────┐
│ 8. Create Virtual SELL Order                │
│    - ClOrdID: "VIRT-SELL-1732492815-abc"    │
│    - Team: Team B                           │
│    - Side: SELL                             │
│    - Quantity: 15 GUACA                     │
│    - Price: $10.80                          │
└─────────────┬───────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────┐
│ 9. Execute Trade (Same as Mechanism 1)      │
│    - Match BUY (Team A) with SELL (Team B)  │
│    - Trade Price: $10.80                    │
│    - Update inventories & balances          │
│    - Send FILL messages                     │
│    - Offer deleted from active offers       │
└─────────────────────────────────────────────┘
```

### Code: Offer Generation

**File**: `internal/market/offer_generator.go` (line 242-342)

```go
func (og *OfferGenerator) GenerateTargetedOffer(
    buyOrder *domain.Order,
    eligibleTeams []*domain.Team,
) error {
    // 1. Calculate offer price (10% above mid)
    marketState, _ := og.marketRepo.GetByProduct(context.Background(), buyOrder.Product)
    var offerPrice float64
    if marketState.Mid != nil {
        offerPrice = *marketState.Mid * 1.10  // 10% premium
    } else {
        offerPrice = 10.0
    }
    
    // 2. Generate unique offer ID
    offerID := fmt.Sprintf("off-%d-%s", time.Now().Unix(), uuid.New().String()[:8])
    
    // 3. Set expiration
    var expiresIn *int
    var expiresAt time.Time
    if og.config.Market.OfferTimeout > 0 {
        timeoutMs := int(og.config.Market.OfferTimeout.Milliseconds())
        expiresIn = &timeoutMs
        expiresAt = time.Now().Add(og.config.Market.OfferTimeout)
    }
    
    // 4. Create offer message
    offerMsg := &domain.OfferMessage{
        Type:              "OFFER",
        OfferID:           offerID,
        Buyer:             buyOrder.TeamName,
        Product:           buyOrder.Product,
        QuantityRequested: buyOrder.Quantity - buyOrder.FilledQty,
        MaxPrice:          offerPrice,
        ExpiresIn:         expiresIn,
        Timestamp:         time.Now(),
    }
    
    // 5. Store in memory
    og.activeOffers[offerID] = &ActiveOffer{
        OfferMsg:  offerMsg,
        BuyOrder:  buyOrder,
        ExpiresAt: expiresAt,
    }
    
    // 6. Send to eligible teams ONLY
    for _, team := range eligibleTeams {
        if team.TeamName != buyOrder.TeamName {  // Don't send to buyer
            og.broadcaster.SendToClient(team.TeamName, offerMsg)
            
            log.Debug().
                Str("team", team.TeamName).
                Str("offerID", offerID).
                Int("teamInventory", team.Inventory[buyOrder.Product]).
                Msg("Targeted offer sent to team with inventory")
        }
    }
}
```

### Code: Accepting Offer

**File**: `internal/market/offer_generator.go` (line 345-411)

```go
func (og *OfferGenerator) HandleAcceptOffer(
    acceptMsg *domain.AcceptOfferMessage,
    acceptorTeam string,
) error {
    // 1. Find the offer
    offer, exists := og.activeOffers[acceptMsg.OfferID]
    if !exists {
        return fmt.Errorf("offer not found or expired: %s", acceptMsg.OfferID)
    }
    
    // 2. Check expiration
    if time.Now().After(offer.ExpiresAt) {
        delete(og.activeOffers, acceptMsg.OfferID)
        return fmt.Errorf("offer expired: %s", acceptMsg.OfferID)
    }
    
    // 3. Handle rejection
    if !acceptMsg.Accept {
        log.Debug().Str("offerID", acceptMsg.OfferID).Msg("Offer declined")
        return nil
    }
    
    // 4. Validate quantities and prices
    if acceptMsg.QuantityOffered <= 0 || acceptMsg.PriceOffered <= 0 {
        return fmt.Errorf("invalid quantity or price in offer acceptance")
    }
    
    // 5. Create virtual SELL order
    virtualSellOrder := &domain.Order{
        ClOrdID:   fmt.Sprintf("VIRT-SELL-%d-%s", time.Now().UnixNano(), uuid.New().String()[:8]),
        TeamName:  acceptorTeam,
        Side:      "SELL",
        Mode:      "MARKET",
        Product:   offer.BuyOrder.Product,
        Quantity:  acceptMsg.QuantityOffered,
        Price:     &acceptMsg.PriceOffered,
        Message:   fmt.Sprintf("Accepting offer %s", acceptMsg.OfferID),
        CreatedAt: time.Now(),
        Status:    "PENDING",
        FilledQty: 0,
    }
    
    // 6. Execute immediate match (uses same transaction as normal trades)
    err := og.executeOfferMatch(offer.BuyOrder, virtualSellOrder)
    if err != nil {
        return fmt.Errorf("failed to execute offer acceptance: %w", err)
    }
    
    // 7. Clean up
    delete(og.activeOffers, acceptMsg.OfferID)
    
    log.Info().
        Str("offerID", acceptMsg.OfferID).
        Str("buyer", offer.BuyOrder.TeamName).
        Str("seller", acceptorTeam).
        Str("product", offer.BuyOrder.Product).
        Int("qty", acceptMsg.QuantityOffered).
        Float64("price", acceptMsg.PriceOffered).
        Msg("Offer accepted and trade executed")
    
    return nil
}
```

### UI: How to Accept OFFER

**Web Interface**:

1. **Receive Offer Notification**:
   - Toast notification: "📬 New offer: 15 GUACA from Team A @ $11.00"
   - Badge appears on "Offers" tab with count

2. **Go to Offers Tab**:
   - See pending offers with details:
     - Buyer name
     - Product & quantity
     - Max price offered
     - Expiration countdown

3. **Negotiate**:
   - Adjust quantity (can sell less than requested)
   - Adjust price (cannot exceed max price)
   - Example:
     - Offer: 15 GUACA @ $11.00
     - You counter: 10 GUACA @ $10.50

4. **Accept or Reject**:
   - Click **"Accept"** → Trade executes immediately
   - Click **"Reject"** → Offer removed from your list

**Message Format (ACCEPT)**:
```json
{
  "type": "ACCEPT_OFFER",
  "offerId": "off-1732492800-xyz789",
  "accept": true,
  "quantityOffered": 10,
  "priceOffered": 10.50
}
```

**Message Format (REJECT)**:
```json
{
  "type": "ACCEPT_OFFER",
  "offerId": "off-1732492800-xyz789",
  "accept": false,
  "quantityOffered": 0,
  "priceOffered": 0
}
```

### Offer Display Code

**File**: `web/index.html` (line 1932-2001)

```javascript
function displayOffers() {
    const offersList = document.getElementById('offers-list');
    
    if (!activeOffers || activeOffers.length === 0) {
        offersList.innerHTML = `
            <div class="flex items-center justify-center h-full text-gray-500">
                <i class="fas fa-info-circle mr-2"></i>No pending offers
            </div>
        `;
        return;
    }
    
    offersList.innerHTML = activeOffers.map(offer => {
        const expiresIn = offer.expiresIn 
            ? `Expires in ${Math.floor(offer.expiresIn / 1000)}s` 
            : 'No expiry';
        
        return `
            <div class="p-4 bg-white rounded-lg border-2 border-yellow-400 shadow-lg">
                <div class="flex items-start justify-between mb-3">
                    <div class="flex-1">
                        <div class="flex items-center mb-2">
                            <i class="fas fa-handshake text-yellow-500 mr-2"></i>
                            <span class="font-semibold">
                                ${offer.buyer} wants to buy
                            </span>
                        </div>
                        <div class="text-lg font-bold text-primary-600">
                            ${offer.quantityRequested} ${offer.product}
                        </div>
                        <div class="text-sm text-gray-600">
                            Offering up to: <span class="font-semibold text-green-600">
                                $${offer.maxPrice.toFixed(2)}
                            </span> per unit
                        </div>
                        <div class="text-xs text-gray-500 mt-1">
                            <i class="fas fa-clock mr-1"></i>${expiresIn}
                        </div>
                    </div>
                </div>
                
                <!-- Negotiation inputs -->
                <div class="grid grid-cols-2 gap-2 mb-2">
                    <div>
                        <label class="block text-xs text-gray-600 mb-1">Quantity to sell</label>
                        <input type="number" id="qty-${offer.offerId}" 
                               value="${offer.quantityRequested}" 
                               min="1" 
                               max="${offer.quantityRequested}"
                               class="w-full px-2 py-1 text-sm border rounded">
                    </div>
                    <div>
                        <label class="block text-xs text-gray-600 mb-1">Your price</label>
                        <input type="number" id="price-${offer.offerId}" 
                               value="${offer.maxPrice.toFixed(2)}" 
                               step="0.01"
                               max="${offer.maxPrice}"
                               class="w-full px-2 py-1 text-sm border rounded">
                    </div>
                </div>
                
                <!-- Accept/Reject buttons -->
                <div class="flex space-x-2">
                    <button onclick="acceptOffer('${offer.offerId}')" 
                            class="flex-1 bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded-lg text-sm font-medium">
                        <i class="fas fa-check mr-1"></i>Accept
                    </button>
                    <button onclick="rejectOffer('${offer.offerId}')" 
                            class="flex-1 bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded-lg text-sm font-medium">
                        <i class="fas fa-times mr-1"></i>Reject
                    </button>
                </div>
            </div>
        `;
    }).join('');
}
```

---

## MECHANISM 3: Automated Bot Responses

### Strategy-Based Acceptance

Automated trading bots can respond to offers programmatically based on their strategy.

**File**: `internal/autoclient/strategy/common.go` (line 96-104)

```go
// CreateAcceptOffer creates an accept offer message
func CreateAcceptOffer(offerID string, quantity int, price float64) *domain.AcceptOfferMessage {
    return &domain.AcceptOfferMessage{
        Type:            "ACCEPT_OFFER",
        OfferID:         offerID,
        Accept:          true,
        QuantityOffered: quantity,
        PriceOffered:    price,
    }
}
```

### Example: Arbitrage Strategy

**File**: `internal/autoclient/strategy/arbitrage.go` (line 88-127)

```go
func (s *ArbitrageStrategy) HandleOffer(offer *domain.OfferMessage) *TradingAction {
    // Check if we have inventory
    inventory := s.state.GetInventory(offer.Product)
    if inventory < offer.QuantityRequested {
        return s.rejectOffer(offer.OfferID, "Insufficient inventory")
    }
    
    // Calculate if accepting this offer creates a profit opportunity
    currentMid := s.state.GetMidPrice(offer.Product)
    if currentMid == nil {
        return s.rejectOffer(offer.OfferID, "No market price available")
    }
    
    // Accept if price is above our threshold
    profitMargin := (offer.MaxPrice - *currentMid) / *currentMid
    if profitMargin > s.profitThreshold {
        return &TradingAction{
            Type: ActionTypeAcceptOffer,
            AcceptOffer: CreateAcceptOffer(
                offer.OfferID,
                offer.QuantityRequested,
                offer.MaxPrice,
            ),
            Reason: fmt.Sprintf("Profitable offer: %.2f%% margin", profitMargin*100),
        }
    }
    
    return s.rejectOffer(offer.OfferID, "Price below threshold")
}
```

### Example: Liquidity Provider Strategy

**File**: `internal/autoclient/strategy/liquidity_provider.go` (line 120-145)

```go
func (s *LiquidityProviderStrategy) HandleOffer(offer *domain.OfferMessage) *TradingAction {
    // Liquidity providers accept offers probabilistically
    inventory := s.state.GetInventory(offer.Product)
    
    if inventory < offer.QuantityRequested {
        return rejectOffer(offer.OfferID, "Insufficient inventory")
    }
    
    // Random acceptance based on fillRate
    if rand.Float64() < s.fillRate {
        // Accept offer at slightly below max price to ensure fill
        acceptPrice := offer.MaxPrice * 0.98
        
        return &TradingAction{
            Type: ActionTypeAcceptOffer,
            AcceptOffer: CreateAcceptOffer(
                offer.OfferID,
                offer.QuantityRequested,
                acceptPrice,
            ),
            Reason: "Liquidity provider accepting offer",
        }
    }
    
    return rejectOffer(offer.OfferID, "Random rejection")
}
```

---

## COMPARISON OF MECHANISMS

| Aspect | Direct SELL Order | Accept OFFER | Automated Bot |
|--------|------------------|--------------|---------------|
| **Speed** | Instant if match exists | Depends on seller response | Milliseconds |
| **Negotiation** | No - price is fixed | Yes - can adjust qty/price | Strategy-based |
| **Initiative** | Seller decides when to sell | Buyer initiates, seller responds | Bot decides |
| **Visibility** | Public order book | Private offer to eligible teams | Automated |
| **Best For** | Known market conditions | Responding to demand | High-frequency trading |
| **Price Discovery** | Transparent | Negotiated | Algorithm-driven |

---

## INVENTORY VALIDATION (CRITICAL!)

### Before Matching

**File**: `internal/market/matcher.go` (line 266-289)

```go
func (m *Matcher) processSellOrder(sellOrder *domain.Order) (*MatchResult, error) {
    // VALIDATE INVENTORY FIRST
    if m.inventoryService != nil {
        canSell, err := m.inventoryService.CanSell(
            context.Background(),
            sellOrder.TeamName,
            sellOrder.Product,
            sellOrder.Quantity,
        )
        if !canSell {
            log.Info().
                Str("sellTeam", sellOrder.TeamName).
                Str("product", sellOrder.Product).
                Int("qty", sellOrder.Quantity).
                Msg("Seller has insufficient inventory")
            return nil, fmt.Errorf("insufficient inventory for sell order")
        }
    }
    // ... continue with matching
}
```

### During Transaction

**File**: `internal/market/engine.go` (line 380-389)

```go
// Inside executeTradeTransaction (MongoDB transaction)

// Seller loses inventory (unless SERVER)
if sellOrder.TeamName != "SERVER" {
    if err := m.inventoryService.UpdateInventory(
        context.Background(),
        sellOrder.TeamName,
        sellOrder.Product,
        -fillQty,  // Negative: removing from inventory
        "TRADE_SELL",
        sellOrder.ClOrdID,
        fillID
    ); err != nil {
        // Transaction will rollback if this fails
        return nil, fmt.Errorf("failed to update seller inventory: %w", err)
    }
}
```

**Result**: Cannot sell products you don't have!

---

## TARGETED OFFER SYSTEM

### Who Receives Offers?

**File**: `internal/market/matcher.go` (line 190-237)

```go
func (m *Matcher) broadcastOfferToEligibleSellers(buyOrder *domain.Order) {
    go func() {
        // Get teams with inventory for this product
        neededQty := buyOrder.Quantity - buyOrder.FilledQty
        eligibleTeams, err := m.teamRepo.GetTeamsWithInventory(
            ctx,
            buyOrder.Product,
            neededQty
        )
        
        if len(eligibleTeams) == 0 {
            log.Info().
                Str("product", buyOrder.Product).
                Int("neededQty", neededQty).
                Msg("No teams have sufficient inventory for offer")
            return
        }
        
        log.Info().
            Int("eligibleTeams", len(eligibleTeams)).
            Msg("Broadcasting offer to eligible sellers")
        
        // Generate targeted offer to eligible teams
        og.GenerateTargetedOffer(buyOrder, eligibleTeams)
    }()
}
```

### Example Query Result

```
BUY Order: 15 GUACA

Database Query: GetTeamsWithInventory("GUACA", 15)

Results:
┌──────────┬───────────┬──────────┐
│ Team     │ Inventory │ Eligible │
├──────────┼───────────┼──────────┤
│ Team A   │ 0 GUACA   │ ❌ NO    │
│ Team B   │ 20 GUACA  │ ✅ YES   │
│ Team C   │ 30 GUACA  │ ✅ YES   │
│ Team D   │ 5 GUACA   │ ❌ NO    │
│ Team E   │ 100 GUACA │ ✅ YES   │
└──────────┴───────────┴──────────┘

OFFER sent to: Team B, Team C, Team E
OFFER NOT sent to: Team A (buyer), Team D (insufficient)
```

**This is SMART**: 
- Reduces noise (teams only see relevant offers)
- Increases fill rate (only capable sellers notified)
- Preserves privacy (others don't see you can't fulfill)

---

## COMPLETE FLOW DIAGRAM

```
╔═══════════════════════════════════════════════════════════════╗
║                     HOW TEAMS FILL ORDERS                     ║
╚═══════════════════════════════════════════════════════════════╝

┌─────────────────────────────────────────────────────────────┐
│ SCENARIO 1: Direct Matching (Fastest)                      │
└─────────────────────────────────────────────────────────────┘

Team A                     Order Book                  Team B
  │                                                       │
  │ 1. BUY 10 GUACA @ $12                                │
  ├──────────────────────────►                           │
  │                        BUY: A-10@$12                 │
  │                        SELL: (empty)                 │
  │                                                       │
  │                                                       │
  │                                                2. SELL 10 GUACA @ $11
  │                           ◄───────────────────────────┤
  │                                                       │
  │                        ⚡ MATCH!                      │
  │◄──────────── FILL: 10@$11 ─────────────────────────►│
  │   +10 GUACA, -$110                    -10 GUACA, +$110
  │                                                       │

┌─────────────────────────────────────────────────────────────┐
│ SCENARIO 2: Offer System (Interactive)                     │
└─────────────────────────────────────────────────────────────┘

Team A              Market Engine         Team B    Team C
  │                        │                 │         │
  │ 1. BUY 15 GUACA @ MKT  │                 │         │
  ├───────────────────────►│                 │         │
  │                        │ 2. Check inventory        │
  │                        ├─────────► Query DB        │
  │                        │     B: 20 ✅              │
  │                        │     C: 30 ✅              │
  │                        │                 │         │
  │                        │ 3. Generate OFFER         │
  │                        ├────────────────►│         │
  │                        │    📬 OFFER     │         │
  │                        ├────────────────────────►  │
  │                        │              📬 OFFER     │
  │                        │                 │         │
  │                        │ 4. ACCEPT_OFFER │         │
  │                        │◄────────────────┤         │
  │                        │  15@$10.80      │         │
  │                        │                 │         │
  │                        │ 5. Execute Trade│         │
  │◄──────── FILL: 15@$10.80 ───────────────┤         │
  │  +15 GUACA, -$162                -15 GUACA, +$162  │
  │                                                     │

┌─────────────────────────────────────────────────────────────┐
│ SCENARIO 3: Automated Bot (Milliseconds)                   │
└─────────────────────────────────────────────────────────────┘

Team A              Market Engine         BotTeam (Strategy)
  │                        │                     │
  │ 1. BUY 20 GUACA @ MKT  │                     │
  ├───────────────────────►│                     │
  │                        │ 2. Generate OFFER   │
  │                        ├────────────────────►│
  │                        │      📬 OFFER       │
  │                        │                     │
  │                        │        ⚡ Bot analyzes in 10ms
  │                        │        - Check inventory: ✅
  │                        │        - Check profit: ✅
  │                        │        - Decision: ACCEPT
  │                        │                     │
  │                        │ 3. ACCEPT_OFFER     │
  │                        │◄────────────────────┤
  │                        │    (automated)      │
  │                        │                     │
  │◄──────── FILL: 20@$11.00 ───────────────────┤
  │  +20 GUACA, -$220            -20 GUACA, +$220
```

---

## SUMMARY

### ✅ Teams Can Fill Orders By:

1. **Placing SELL orders** that match existing BUY orders
   - Instant execution
   - Public order book
   - Price transparency

2. **Accepting OFFERS** from buyers
   - Targeted to teams with inventory
   - Negotiable quantity and price
   - Interactive and educational

3. **Using automated bots** with strategies
   - Millisecond response times
   - Strategy-based decisions
   - High-frequency trading

### 🛡️ Safety Mechanisms:

- ✅ Inventory checked BEFORE matching
- ✅ Inventory updated in ATOMIC transaction
- ✅ Cannot sell more than you have
- ✅ Cannot accept offers you can't fulfill
- ✅ Offers only sent to capable sellers
- ✅ All trades are atomic (all-or-nothing)

### 📊 Educational Value:

1. **Supply and Demand**: Students see who needs what
2. **Price Discovery**: Negotiation teaches market dynamics
3. **Inventory Management**: Must produce to trade
4. **Strategic Thinking**: When to accept vs. wait
5. **Real-time Feedback**: Instant results from decisions

---

**Investigation completed by**: AI Assistant  
**Date**: November 24, 2025  
**Status**: COMPLETE AND VERIFIED ✅
