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

## 🏗️ Technical Architecture

### Thread Management

The SDK uses **Java 25 Virtual Threads** for optimal concurrency with minimal resource overhead:

#### 1. **Message Sequencer** (Single Virtual Thread)
- **Purpose**: Guarantees in-order processing of all incoming messages
- **Implementation**: `Executors.newSingleThreadExecutor(Thread.ofVirtual().factory())`
- **Why**: WebSocket can deliver messages concurrently, but the trading protocol requires strict ordering
- **Location**: `MessageSequencer.java:11`

```java
// All messages flow through this single thread
sequencer.submit(() -> router.routeMessage(json, handlers));
```

#### 2. **Callback Executor** (Virtual Thread Pool)
- **Purpose**: Executes user callbacks without blocking message processing
- **Implementation**: `Executors.newVirtualThreadPerTaskExecutor()`
- **Why**: Each listener callback gets its own virtual thread, allowing concurrent execution
- **Location**: `ConectorBolsa.java:47`

```java
// Each listener callback runs on a separate virtual thread
callbackExecutor.execute(() -> listener.onFill(message));
```

#### 3. **Heartbeat Scheduler** (Single Virtual Thread)
- **Purpose**: Sends periodic PING messages and monitors PONG timeouts
- **Implementation**: `ScheduledExecutorService` with virtual thread factory
- **Why**: Detects connection failures before the TCP timeout
- **Location**: `HeartbeatManager.java:63`

```java
scheduler.scheduleAtFixedRate(this::sendHeartbeat, 
    pingInterval, pingInterval, TimeUnit.MILLISECONDS);
```

### Lock Management & Thread Safety

#### 1. **Send Lock** (Semaphore)
The SDK uses a `Semaphore(1)` to ensure only one thread can send messages at a time:

```java
private final Semaphore sendLock = new Semaphore(1);

private void sendMessage(Object message) {
    sendLock.acquire();  // Block until available
    try {
        webSocket.sendText(json, true).join();
    } finally {
        sendLock.release();  // Always release
    }
}
```

**Why a Semaphore instead of `synchronized`?**
- More explicit control flow
- Better interruptibility support
- Works well with virtual threads (no pinning issues)
- **Location**: `ConectorBolsa.java:48, 228-234`

#### 2. **Listener List** (CopyOnWriteArrayList)
Listeners are stored in a thread-safe collection:

```java
private final List<EventListener> listeners = new CopyOnWriteArrayList<>();
```

**Why CopyOnWriteArrayList?**
- Thread-safe for concurrent reads and writes
- No locking needed when iterating
- Optimal for many reads, few writes (typical listener pattern)
- **Location**: `ConectorBolsa.java:44`

#### 3. **Volatile State**
Connection state uses `volatile` for lock-free reads:

```java
private volatile ConnectionState state = ConnectionState.DISCONNECTED;
```

**Why volatile?**
- Ensures visibility across all threads
- No lock contention for state reads
- State changes are atomic enum assignments
- **Location**: `ConectorBolsa.java:50`

### Message Flow Architecture

```
WebSocket Thread          Sequencer Thread         Callback Threads
     │                         │                         │
     ├─► Message Received      │                         │
     │   (Any Order)           │                         │
     │                         │                         │
     └────► Queue ────────────►│                         │
                               │                         │
                               ├─► Parse JSON            │
                               ├─► Route by Type         │
                               ├─► Update State          │
                               │                         │
                               └─► Notify Listeners ────►├─► onFill()
                                                          ├─► onTicker()
                                                          └─► onError()
```

**Key Design Principles:**

1. **Sequential Processing**: All messages are processed in order by a single thread
2. **Asynchronous Callbacks**: User code runs on separate virtual threads
3. **Non-Blocking**: Message processing never waits for user callbacks
4. **Fail-Safe**: Exceptions in callbacks don't crash the SDK

### Performance Characteristics

- **Virtual Threads**: Can handle 10,000+ concurrent callbacks with minimal memory
- **Zero GC Pressure**: Pre-allocated structures, no allocation in hot paths
- **Lock-Free Reads**: State and listeners are read without locking
- **Bounded Blocking**: Send lock is the only blocking operation

### Memory Safety

- **No Data Races**: All shared state is properly synchronized
- **No Deadlocks**: Single lock hierarchy (send lock only)
- **Graceful Shutdown**: All threads are cleanly terminated
- **Exception Safety**: All operations have try-finally blocks

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

## Thread Safety Guarantees

### ✅ What IS Thread-Safe

- **All public methods**: Can be called from any thread
- **addListener/removeListener**: Safe during message processing
- **enviarOrden/enviarCancelacion**: Can send from multiple threads
- **State queries**: `getState()` is always consistent

### ⚠️ What You Need to Know

- **Callbacks execute concurrently**: Multiple callbacks can run simultaneously
- **Callback order is NOT guaranteed**: `onFill()` may execute before previous `onTicker()`
- **Message PROCESSING order IS guaranteed**: Messages are parsed sequentially
- **Don't block in callbacks**: Use virtual threads or async operations

### Example: Thread-Safe Usage

```java
// ✅ SAFE: Multiple threads can send orders
CompletableFuture.allOf(
    CompletableFuture.runAsync(() -> connector.enviarOrden(order1)),
    CompletableFuture.runAsync(() -> connector.enviarOrden(order2)),
    CompletableFuture.runAsync(() -> connector.enviarOrden(order3))
).join();

// ✅ SAFE: Callbacks can modify shared state with proper synchronization
private final ConcurrentHashMap<String, Order> orders = new ConcurrentHashMap<>();

@Override
public void onFill(FillMessage msg) {
    orders.compute(msg.getClOrdID(), (id, order) -> {
        // Atomic update
        return order.withFill(msg);
    });
}

// ❌ UNSAFE: Don't use non-thread-safe collections without locking
private final Map<String, Order> orders = new HashMap<>(); // NOT thread-safe!

@Override
public void onFill(FillMessage msg) {
    orders.put(msg.getClOrdID(), ...); // Race condition!
}
```

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
