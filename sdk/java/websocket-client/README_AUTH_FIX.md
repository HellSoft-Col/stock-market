# Authentication Fix for SDK

## Problem
Previously, when you called `conectar()`, the method would return immediately after establishing the WebSocket connection, but **before** receiving the LOGIN_OK response from the server. If you tried to send orders or production updates immediately after connecting, you would get an "Not authenticated" error because the authentication state transition happens asynchronously.

## Solution
The SDK now provides three ways to handle authentication:

### Option 1: conectarYEsperarLogin() - Blocking (Recommended for Simple Cases)
```java
ConectorBolsa connector = new ConectorBolsa();
connector.addListener(new MyEventListener());

// This blocks until LOGIN_OK is received
LoginOKMessage loginInfo = connector.conectarYEsperarLogin(
    "wss://trading.hellsoft.tech/ws", 
    "your-token"
);

// Now safe to send orders immediately
connector.enviarOrden(order);
```

### Option 2: conectar() + esperarLogin() - Flexible Blocking
```java
ConectorBolsa connector = new ConectorBolsa();
connector.addListener(new MyEventListener());

// Connect (non-blocking)
connector.conectar("wss://trading.hellsoft.tech/ws", "your-token");

// Do other initialization work here...

// Wait for authentication to complete
LoginOKMessage loginInfo = connector.esperarLogin();

// Or wait with timeout
try {
    LoginOKMessage loginInfo = connector.esperarLogin(5, TimeUnit.SECONDS);
} catch (TimeoutException e) {
    log.error("Authentication timed out");
}

// Now safe to send orders
connector.enviarOrden(order);
```

### Option 3: Event Listener - Async (Recommended for Production)
```java
ConectorBolsa connector = new ConectorBolsa();

connector.addListener(new EventListener() {
    @Override
    public void onLoginOk(LoginOKMessage message) {
        log.info("Authenticated as team: {}", message.getTeam());
        
        // Now safe to send orders
        startTrading();
    }
    
    @Override
    public void onError(ErrorMessage message) {
        if (message.getCode() == ErrorCode.AUTH_FAILED) {
            log.error("Authentication failed: {}", message.getReason());
        }
    }
    
    // ... implement other methods
});

// Connect (non-blocking)
connector.conectar("wss://trading.hellsoft.tech/ws", "your-token");

// Your application continues running...
```

## Complete Working Example
```java
import tech.hellsoft.trading.*;
import tech.hellsoft.trading.dto.client.*;
import tech.hellsoft.trading.dto.server.*;
import tech.hellsoft.trading.enums.*;

public class TradingBotExample {
    public static void main(String[] args) throws Exception {
        ConectorBolsa connector = new ConectorBolsa();
        
        // Add event listener for receiving market data
        connector.addListener(new EventListener() {
            @Override
            public void onLoginOk(LoginOKMessage message) {
                System.out.println("✅ Authenticated as: " + message.getTeam());
                System.out.println("Initial balance: $" + message.getInitialBalance());
            }
            
            @Override
            public void onFill(FillMessage message) {
                System.out.println("✅ Order filled: " + message.getClOrdID());
            }
            
            @Override
            public void onTicker(TickerMessage message) {
                System.out.println("📊 Price update: " + message.getProduct() + 
                                   " = $" + message.getPrice());
            }
            
            @Override
            public void onError(ErrorMessage message) {
                System.err.println("❌ Error: " + message.getReason());
            }
            
            // ... implement other required methods
        });
        
        // Connect and wait for authentication
        System.out.println("Connecting to server...");
        LoginOKMessage loginInfo = connector.conectarYEsperarLogin(
            "wss://trading.hellsoft.tech/ws",
            System.getenv("TRADING_TOKEN")
        );
        
        System.out.println("✅ Connected! Authorized products: " + 
                           loginInfo.getAuthorizedProducts());
        
        // Now safe to send orders
        OrderMessage order = OrderMessage.builder()
            .clOrdID("order-" + System.currentTimeMillis())
            .side(OrderSide.BUY)
            .product(Product.GUACA)
            .qty(10)
            .mode(OrderMode.MARKET)
            .build();
        
        System.out.println("Sending order...");
        connector.enviarOrden(order);
        
        // Send production update
        ProductionUpdateMessage production = ProductionUpdateMessage.builder()
            .product(Product.GUACA)
            .quantity(5)
            .build();
        
        System.out.println("Sending production update...");
        connector.enviarActualizacionProduccion(production);
        
        // Keep running to receive updates
        Thread.sleep(10000);
        
        // Clean shutdown
        connector.shutdown();
    }
}
```

## Technical Details

### What Changed Internally?
1. **Added `CompletableFuture<LoginOKMessage> loginFuture`**: A future that completes when LOGIN_OK is received
2. **New methods**: `conectarYEsperarLogin()`, `esperarLogin()`, and `esperarLogin(timeout, unit)`
3. **Auto-complete**: The future automatically completes when `onLoginOk()` is called
4. **Error handling**: If AUTH_FAILED error is received, the future completes exceptionally

### Why Was This Needed?
The WebSocket protocol is inherently asynchronous:
1. `conectar()` establishes the TCP/WebSocket connection (sync)
2. SDK automatically sends LOGIN message (sync)
3. Server processes login and sends LOGIN_OK (async!)
4. SDK receives LOGIN_OK and updates state to AUTHENTICATED (async!)

The gap between steps 2 and 4 could cause "Not authenticated" errors if you sent messages too quickly.

### Thread Safety
- All waiting methods use `CompletableFuture`, which is thread-safe
- The `loginFuture` is marked `volatile` for visibility across threads
- Virtual threads are used throughout, so blocking is cheap

## Migration Guide

### Before (Would Fail)
```java
connector.conectar("wss://server.com/ws", "token");
connector.enviarOrden(order); // ❌ Not authenticated!
```

### After (Option 1 - Simplest)
```java
connector.conectarYEsperarLogin("wss://server.com/ws", "token");
connector.enviarOrden(order); // ✅ Works!
```

### After (Option 2 - Flexible)
```java
connector.conectar("wss://server.com/ws", "token");
connector.esperarLogin(); // Wait here
connector.enviarOrden(order); // ✅ Works!
```

### After (Option 3 - Production Ready)
```java
connector.addListener(new EventListener() {
    @Override
    public void onLoginOk(LoginOKMessage msg) {
        connector.enviarOrden(order); // ✅ Works!
    }
    // ...
});
connector.conectar("wss://server.com/ws", "token");
```
