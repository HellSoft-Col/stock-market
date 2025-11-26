# SDK Emulator Guide

## Overview

The SDK Emulator is a powerful debugging tool integrated into the web UI that allows you to test all Java SDK EventListener callbacks without needing multiple teams or complex setups. This is perfect for:

- Testing how your Java trading bot handles different market scenarios
- Debugging edge cases (errors, offer expiration, etc.)
- Simulating fills and inventory changes
- Testing balance updates and error handling

## How It Works

1. **Web UI Connection**: You connect to the trading server using the web UI (typically with your team credentials)
2. **Java SDK Connection**: Your Java SDK client connects to the same server with a different team name
3. **Message Routing**: The web UI sends specially formatted messages to the server, which routes them to your Java SDK client
4. **Event Callbacks**: Your Java SDK's EventListener methods are triggered as if the events came from real market activity

## Setup Instructions

### 1. Start the Trading Server

```bash
cd /path/to/stock-market
go run cmd/server/main.go
```

### 2. Connect Your Java SDK Client

Create a simple test client that logs all events:

```java
import tech.hellsoft.trading.*;
import tech.hellsoft.trading.dto.server.*;

public class TestClient implements EventListener {
    public static void main(String[] args) throws Exception {
        ConectorBolsa connector = new ConectorBolsa();
        connector.addListener(new TestClient());
        
        // Connect to the server
        connector.conectar("ws://localhost:8080/ws", "your-team-token");
        
        System.out.println("Client connected as: YOUR-TEAM-NAME");
        System.out.println("Use this team name in the SDK Emulator web UI");
        
        // Keep running
        Thread.sleep(Long.MAX_VALUE);
    }
    
    @Override
    public void onOffer(OfferMessage message) {
        System.out.println("✅ onOffer triggered!");
        System.out.println("   Buyer: " + message.getBuyer());
        System.out.println("   Product: " + message.getProduct());
        System.out.println("   Quantity: " + message.getQuantityRequested());
        System.out.println("   Max Price: $" + message.getMaxPrice());
    }
    
    @Override
    public void onFill(FillMessage message) {
        System.out.println("✅ onFill triggered!");
        System.out.println("   Side: " + message.getSide());
        System.out.println("   Product: " + message.getProduct());
        System.out.println("   Quantity: " + message.getFillQty());
        System.out.println("   Price: $" + message.getFillPrice());
    }
    
    @Override
    public void onError(ErrorMessage message) {
        System.out.println("✅ onError triggered!");
        System.out.println("   Code: " + message.getCode());
        System.out.println("   Reason: " + message.getReason());
    }
    
    @Override
    public void onInventoryUpdate(InventoryUpdateMessage message) {
        System.out.println("✅ onInventoryUpdate triggered!");
        System.out.println("   New Inventory: " + message.getInventory());
    }
    
    @Override
    public void onBalanceUpdate(BalanceUpdateMessage message) {
        System.out.println("✅ onBalanceUpdate triggered!");
        System.out.println("   New Balance: $" + message.getNewBalance());
    }
    
    @Override
    public void onBroadcast(BroadcastNotificationMessage message) {
        System.out.println("✅ onBroadcast triggered!");
        System.out.println("   From: " + message.getSender());
        System.out.println("   Message: " + message.getMessage());
    }
    
    @Override
    public void onOrderAck(OrderAckMessage message) {
        System.out.println("✅ onOrderAck triggered!");
        System.out.println("   Order ID: " + message.getClOrdID());
        System.out.println("   Status: " + message.getStatus());
    }
    
    @Override
    public void onTicker(TickerMessage message) {
        System.out.println("✅ onTicker triggered!");
        System.out.println("   Product: " + message.getProduct());
        System.out.println("   Price: $" + message.getPrice());
    }
    
    @Override
    public void onEventDelta(EventDeltaMessage message) {
        System.out.println("✅ onEventDelta triggered!");
        System.out.println("   Events: " + message.getEvents().size());
    }
    
    @Override
    public void onLoginOk(LoginOKMessage message) {
        System.out.println("✅ Login successful!");
        System.out.println("   Team: " + message.getTeam());
    }
    
    @Override
    public void onConnectionLost(Throwable error) {
        System.out.println("❌ Connection lost: " + error.getMessage());
    }
    
    @Override
    public void onGlobalPerformanceReport(GlobalPerformanceReportMessage message) {
        System.out.println("✅ onGlobalPerformanceReport triggered!");
    }
}
```

### 3. Open the Web UI

1. Navigate to `http://localhost:8080/` in your browser
2. Click "Connect" to connect to the WebSocket server
3. Enter your team token and click "Login"
4. Navigate to the **"SDK Emulator"** tab

### 4. Configure the Target Team

In the "Target Configuration" section:
1. Enter the team name of your Java SDK client in the "Target Team Name" field
2. This is the team name that was shown when your Java client connected

### 5. Test Event Callbacks

Now you can test each EventListener callback:

#### Testing onOffer

1. Fill in the offer details:
   - Buyer Team: `TEAM-BOT-01` (simulated buyer)
   - Product: `GUACA`
   - Quantity: `10`
   - Max Price: `25.00`
   - Expires In: `30000` (ms)

2. Click "Send OFFER to Client"

3. Your Java SDK client should print:
   ```
   ✅ onOffer triggered!
      Buyer: TEAM-BOT-01
      Product: GUACA
      Quantity: 10
      Max Price: $25.0
   ```

#### Testing onFill

1. Configure the fill details
2. Click "Send FILL to Client"
3. Observe the callback in your Java client

#### Testing onError

1. Select an error code (e.g., `INSUFFICIENT_BALANCE`)
2. Enter an error message
3. Click "Send ERROR to Client"
4. Your Java client's `onError` method will be called

## Available Events

The SDK Emulator supports all Java SDK EventListener callbacks:

| Event | Method | Description |
|-------|--------|-------------|
| OFFER | `onOffer()` | Simulate offers from other teams |
| FILL | `onFill()` | Simulate order fills/executions |
| ERROR | `onError()` | Test error handling |
| INVENTORY_UPDATE | `onInventoryUpdate()` | Update inventory levels |
| BALANCE_UPDATE | `onBalanceUpdate()` | Change account balance |
| BROADCAST_NOTIFICATION | `onBroadcast()` | Server-wide announcements |
| ORDER_ACK | `onOrderAck()` | Order acknowledgments |
| TICKER | `onTicker()` | Market price updates |
| EVENT_DELTA | `onEventDelta()` | Resync events |

## Common Use Cases

### 1. Testing Offer Handling

Simulate offers from multiple teams to test your bot's decision-making:

```java
@Override
public void onOffer(OfferMessage message) {
    // Your bot decides whether to accept
    if (message.getMaxPrice() >= myTargetPrice) {
        AcceptOfferMessage response = AcceptOfferMessage.builder()
            .offerId(message.getOfferId())
            .quantity(message.getQuantityRequested())
            .price(message.getMaxPrice())
            .build();
        connector.enviarRespuestaOferta(response);
    }
}
```

### 2. Testing Error Recovery

Test how your bot handles errors:

1. Send `INSUFFICIENT_BALANCE` error
2. Send `UNAUTHORIZED_PRODUCT` error
3. Send `RATE_LIMIT_EXCEEDED` error

Verify your bot handles each appropriately.

### 3. Testing Inventory Management

Simulate production or trades affecting inventory:

1. Update inventory to low levels
2. Verify your bot places buy orders
3. Update inventory to high levels
4. Verify your bot places sell orders

## Tips

1. **Use the Emulation Log**: The emulation log at the bottom shows all sent messages
2. **Target Team Validation**: Make sure your Java client is connected before sending events
3. **Message Timing**: Some events (like offers) have expiration times - test quickly!
4. **Combine Events**: Test sequences like: Fill → Inventory Update → Balance Update

## Troubleshooting

### "Please login first" error
- Make sure you're logged in with the web UI

### Java client doesn't receive messages
- Verify the target team name matches exactly
- Check that your Java client is still connected
- Look for errors in the server logs

### Messages not showing in Java client
- Ensure your EventListener methods are implemented
- Check for exceptions in your callback methods
- Verify the Java SDK is not filtering messages

## Technical Details

### Message Flow

```
Web UI (You) 
    ↓
    SDK_EMULATOR message
    ↓
Server (MessageRouter)
    ↓
    Validates & Routes
    ↓
Java SDK Client (Target Team)
    ↓
    EventListener.onXXX() called
```

### SDK_EMULATOR Message Format

```json
{
  "type": "SDK_EMULATOR",
  "targetTeam": "YOUR-JAVA-CLIENT-TEAM",
  "messageType": "OFFER",
  "messagePayload": {
    "offerId": "OFFER-123",
    "buyer": "TEAM-BOT-01",
    "product": "GUACA",
    "quantityRequested": 10,
    "maxPrice": 25.0,
    "expiresIn": 30000
  }
}
```

The server unwraps this and sends just the payload with the messageType to the target team.

## Security Note

The SDK Emulator is a debugging tool. In production:
- Consider restricting who can send SDK_EMULATOR messages
- Add authentication/authorization checks
- Log all emulator usage for audit purposes

## Next Steps

1. Implement automated test suites using the emulator
2. Create scenario scripts for complex market conditions
3. Use for training new developers on SDK usage
4. Build integration tests that verify bot behavior

---

**Happy Testing!** 🚀

If you encounter any issues or have suggestions for improvement, please file an issue in the repository.
