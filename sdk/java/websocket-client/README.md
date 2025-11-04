# Stock Market Trading SDK - Java 25

WebSocket-based client SDK for connecting to the Stock Market trading server.

## Features

- ✅ Java 25 with virtual threads for scalability
- ✅ WebSocket communication with built-in `java.net.http.WebSocket`
- ✅ JSON serialization with Gson
- ✅ Automatic heartbeat/ping-pong management
- ✅ Sequential message processing
- ✅ Thread-safe listener notifications
- ✅ Comprehensive error handling
- ✅ Clean callback-based API

## Installation

### Gradle

```kotlin
dependencies {
    implementation("tech.hellsoft.trading:websocket-client:1.0.0-SNAPSHOT")
}
```

### Maven

```xml
<dependency>
    <groupId>tech.hellsoft.trading</groupId>
    <artifactId>websocket-client</artifactId>
    <version>1.0.0-SNAPSHOT</version>
</dependency>
```

## Quick Start

```java
import tech.hellsoft.trading.*;
import tech.hellsoft.trading.dto.client.*;
import tech.hellsoft.trading.dto.server.*;
import tech.hellsoft.trading.enums.*;

public class Example implements EventListener {
    public static void main(String[] args) throws Exception {
        // Create connector with default config
        ConectorBolsa connector = new ConectorBolsa();
        
        // Add listener
        connector.addListener(new Example());
        
        // Connect and authenticate
        connector.conectar("localhost", 8080, "your-token-here");
        
        // Send a buy order
        OrderMessage order = OrderMessage.builder()
            .clOrdID("order-001")
            .side(OrderSide.BUY)
            .mode(OrderMode.LIMIT)
            .product(Product.GUACA)
            .qty(10)
            .limitPrice(100.0)
            .build();
        
        connector.enviarOrden(order);
        
        // Keep running
        Thread.sleep(60000);
        
        // Disconnect
        connector.desconectar();
        connector.shutdown();
    }
    
    @Override
    public void onLoginOk(LoginOKMessage message) {
        System.out.println("Logged in as: " + message.getTeam());
        System.out.println("Balance: " + message.getCurrentBalance());
    }
    
    @Override
    public void onFill(FillMessage message) {
        System.out.println("Order filled: " + message.getClOrdID() +
            " - " + message.getFillQty() + " @ " + message.getFillPrice());
    }
    
    @Override
    public void onTicker(TickerMessage message) {
        System.out.println(message.getProduct() + " - " +
            "Bid: " + message.getBestBid() + " Ask: " + message.getBestAsk());
    }
    
    @Override
    public void onOffer(OfferMessage message) {
        System.out.println("Offer received: " + message.getOfferId());
    }
    
    @Override
    public void onError(ErrorMessage message) {
        System.err.println("Error: " + message.getCode() + " - " + message.getReason());
    }
    
    @Override
    public void onOrderAck(OrderAckMessage message) {
        System.out.println("Order acknowledged: " + message.getClOrdID());
    }
    
    @Override
    public void onInventoryUpdate(InventoryUpdateMessage message) {
        System.out.println("Inventory updated: " + message.getInventory());
    }
    
    @Override
    public void onBalanceUpdate(BalanceUpdateMessage message) {
        System.out.println("Balance updated: " + message.getBalance());
    }
    
    @Override
    public void onEventDelta(EventDeltaMessage message) {
        System.out.println("Event delta received");
    }
    
    @Override
    public void onBroadcast(BroadcastNotificationMessage message) {
        System.out.println("Broadcast: " + message.getMessage());
    }
    
    @Override
    public void onConnectionLost(Throwable error) {
        System.err.println("Connection lost: " + error.getMessage());
    }
}
```

## Configuration

Customize SDK behavior with `ConectorConfig`:

```java
ConectorConfig config = ConectorConfig.builder()
    .heartbeatInterval(Duration.ofSeconds(30))
    .connectionTimeout(Duration.ofSeconds(10))
    .autoReconnect(true)
    .maxReconnectAttempts(5)
    .build();

ConectorBolsa connector = new ConectorBolsa(config);
```

## API Reference

### ConectorBolsa Methods

#### Connection Management

```java
void conectar(String host, int port, String token) throws ConexionFallidaException
void desconectar()
void shutdown()
```

#### Message Sending

```java
void enviarOrden(OrderMessage order)
void enviarCancelacion(String clOrdID)
void enviarActualizacionProduccion(ProductionUpdateMessage update)
void enviarRespuestaOferta(AcceptOfferMessage response)
```

#### Listeners

```java
void addListener(EventListener listener)
void removeListener(EventListener listener)
```

#### State

```java
ConnectionState getState()
```

### EventListener Interface

All callback methods run on virtual threads and must not block:

```java
void onLoginOk(LoginOKMessage message)
void onFill(FillMessage message)
void onTicker(TickerMessage message)
void onOffer(OfferMessage message)
void onError(ErrorMessage message)
void onOrderAck(OrderAckMessage message)
void onInventoryUpdate(InventoryUpdateMessage message)
void onBalanceUpdate(BalanceUpdateMessage message)
void onEventDelta(EventDeltaMessage message)
void onBroadcast(BroadcastNotificationMessage message)
void onConnectionLost(Throwable error)
```

## Building Orders

### Market Order

```java
OrderMessage order = OrderMessage.builder()
    .clOrdID("order-123")
    .side(OrderSide.BUY)
    .mode(OrderMode.MARKET)
    .product(Product.FOSFO)
    .qty(50)
    .build();
```

### Limit Order

```java
OrderMessage order = OrderMessage.builder()
    .clOrdID("order-456")
    .side(OrderSide.SELL)
    .mode(OrderMode.LIMIT)
    .product(Product.PALTA_OIL)
    .qty(25)
    .limitPrice(150.0)
    .message("Optional message")
    .build();
```

## Products

Available products:

- `Product.GUACA` - GUACA
- `Product.SEBO` - SEBO
- `Product.PALTA_OIL` - PALTA-OIL
- `Product.CASCAR_ALLOY` - CASCAR-ALLOY
- `Product.FOSFO` - FOSFO

## Error Handling

Handle errors via the `onError` callback:

```java
@Override
public void onError(ErrorMessage message) {
    switch (message.getCode()) {
        case AUTH_FAILED:
            // Handle authentication failure
            break;
        case INSUFFICIENT_INVENTORY:
            // Handle insufficient inventory
            break;
        case INVALID_ORDER:
            // Handle invalid order
            break;
        default:
            log.error("Error: {}", message.getReason());
    }
}
```

## Thread Safety

- All public methods are thread-safe
- Listener callbacks execute on virtual threads
- Messages are processed sequentially
- Sending messages is synchronized

## Build Requirements

- Java 25
- Gradle 8.5+

## Build

```bash
./gradlew build
```

## Test

```bash
./gradlew test
```

## License

MIT
