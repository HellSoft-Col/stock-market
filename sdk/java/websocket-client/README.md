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

The SDK is published to **GitHub Packages** automatically on every push to `main`.

### Option 1: GitHub Packages (Recommended)

#### Step 1: Create GitHub Personal Access Token
1. Go to GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Select scope: `read:packages`
4. Copy the token (you'll need it in Step 2)

#### Step 2: Configure Gradle

Add to your `build.gradle.kts`:

```kotlin
repositories {
    mavenCentral()
    maven {
        url = uri("https://maven.pkg.github.com/HellSoft-Col/stock-market") // Replace HellSoft-Col/stock-market
        credentials {
            username = project.findProperty("gpr.user") as String? ?: System.getenv("GITHUB_ACTOR")
            password = project.findProperty("gpr.token") as String? ?: System.getenv("GITHUB_TOKEN")
        }
    }
}

dependencies {
    implementation("tech.hellsoft.trading:websocket-client:1.0.0-SNAPSHOT")
}
```

Create `gradle.properties` in your project root (or `~/.gradle/gradle.properties`):
```properties
gpr.user=YOUR_GITHUB_USERNAME
gpr.token=YOUR_GITHUB_TOKEN
```

#### Step 3: Configure Maven (Alternative)

Add to your `pom.xml`:

```xml
<repositories>
    <repository>
        <id>github</id>
        <url>https://maven.pkg.github.com/HellSoft-Col/stock-market</url> <!-- Replace HellSoft-Col/stock-market -->
    </repository>
</repositories>

<dependencies>
    <dependency>
        <groupId>tech.hellsoft.trading</groupId>
        <artifactId>websocket-client</artifactId>
        <version>1.0.0-SNAPSHOT</version>
    </dependency>
</dependencies>
```

Add to `~/.m2/settings.xml`:
```xml
<settings>
    <servers>
        <server>
            <id>github</id>
            <username>YOUR_GITHUB_USERNAME</username>
            <password>YOUR_GITHUB_TOKEN</password>
        </server>
    </servers>
</settings>
```

### Option 2: Download JAR Directly

1. Go to the [GitHub Actions](../../actions/workflows/publish-java-sdk.yml) page
2. Click on the latest successful build
3. Download the `package` artifact
4. Add the JAR to your project:

```kotlin
dependencies {
    implementation(files("libs/websocket-client-1.0.0-SNAPSHOT.jar"))
}
```

### Option 3: Build from Source

```bash
git clone <repository-url>
cd sdk/java/websocket-client
./gradlew build
# JAR will be in: build/libs/websocket-client-1.0.0-SNAPSHOT.jar
```

## Quick Start

```java
import tech.hellsoft.trading.*;
import tech.hellsoft.trading.dto.client.*;
import tech.hellsoft.trading.dto.server.*;
import tech.hellsoft.trading.enums.*;

public class Example implements EventListener {
    public static void main(String[] args) throws Exception {
        // 1️⃣ Create connector with default config
        // This initializes 3 thread pools but does NOT start them yet:
        //   - Message Sequencer (single virtual thread) - CREATED but idle
        //   - Callback Executor (virtual thread pool) - CREATED but idle  
        //   - Heartbeat not created yet (waits for connection)
        ConectorBolsa connector = new ConectorBolsa();
        
        // 2️⃣ Add listener (runs on main thread)
        // Listeners are stored in CopyOnWriteArrayList for thread-safe access
        connector.addListener(new Example());
        
        // 3️⃣ Connect and authenticate
        // This triggers:
        //   - WebSocket connection establishment (platform thread)
        //   - Heartbeat manager starts (scheduled virtual thread)
        //   - LOGIN message sent through send semaphore
        //   - Message sequencer starts processing incoming messages
        connector.conectar("localhost", 8080, "your-token-here");
        // From this point:
        //   - WebSocket thread receives all messages
        //   - Sequencer thread processes them IN ORDER
        //   - Callbacks execute on separate virtual threads
        
        // 4️⃣ Send a buy order (can be called from any thread)
        // Flow:
        //   1. Main thread calls enviarOrden()
        //   2. Acquires send semaphore (blocks if another send is in progress)
        //   3. Serializes to JSON
        //   4. Sends via WebSocket
        //   5. Releases send semaphore
        OrderMessage order = OrderMessage.builder()
            .clOrdID("order-001")
            .side(OrderSide.BUY)
            .mode(OrderMode.LIMIT)
            .product(Product.GUACA)
            .qty(10)
            .limitPrice(100.0)
            .build();
        
        connector.enviarOrden(order);
        // Main thread is now free - order is sent asynchronously
        
        // 5️⃣ Keep running to receive messages
        // Meanwhile, in parallel:
        //   - WebSocket thread: Receives messages continuously
        //   - Sequencer thread: Processes messages one-by-one
        //   - Callback threads: Execute onFill(), onTicker(), etc.
        //   - Heartbeat thread: Sends PING every 30 seconds
        Thread.sleep(60000);
        
        // 6️⃣ Disconnect and shutdown (runs on main thread)
        // This triggers:
        //   - Heartbeat stops (scheduled thread terminates)
        //   - WebSocket closes gracefully
        //   - Sequencer stops accepting new messages
        //   - Callback executor stops (waits for running callbacks)
        connector.desconectar();
        connector.shutdown();
        // All threads are now terminated cleanly
    }
    
    // 🧵 All callbacks below run on SEPARATE virtual threads
    // Each callback execution gets its own virtual thread from the callback executor
    // Multiple callbacks can run concurrently (e.g., onFill and onTicker at same time)
    
    @Override
    public void onLoginOk(LoginOKMessage message) {
        // This runs on: Virtual Thread #1 (from callback executor)
        // Triggered by: Sequencer thread after processing LOGIN_OK message
        System.out.println("Logged in as: " + message.getTeam());
        System.out.println("Balance: " + message.getCurrentBalance());
        // While this runs, sequencer continues processing next messages
    }
    
    @Override
    public void onFill(FillMessage message) {
        // This runs on: Virtual Thread #2 (from callback executor)
        // Can run concurrently with onTicker() or other callbacks
        System.out.println("Order filled: " + message.getClOrdID() +
            " - " + message.getFillQty() + " @ " + message.getFillPrice());
        // ⚠️ If you modify shared state here, use thread-safe collections!
    }
    
    @Override
    public void onTicker(TickerMessage message) {
        // This runs on: Virtual Thread #3 (from callback executor)
        // May execute while onFill() is still running
        System.out.println(message.getProduct() + " - " +
            "Bid: " + message.getBestBid() + " Ask: " + message.getBestAsk());
    }
    
    @Override
    public void onOffer(OfferMessage message) {
        // This runs on: Virtual Thread #4 (from callback executor)
        System.out.println("Offer received: " + message.getOfferId());
    }
    
    @Override
    public void onError(ErrorMessage message) {
        // This runs on: Virtual Thread #5 (from callback executor)
        System.err.println("Error: " + message.getCode() + " - " + message.getReason());
    }
    
    @Override
    public void onOrderAck(OrderAckMessage message) {
        // This runs on: Virtual Thread #6 (from callback executor)
        System.out.println("Order acknowledged: " + message.getClOrdID());
    }
    
    @Override
    public void onInventoryUpdate(InventoryUpdateMessage message) {
        // This runs on: Virtual Thread #7 (from callback executor)
        System.out.println("Inventory updated: " + message.getInventory());
    }
    
    @Override
    public void onBalanceUpdate(BalanceUpdateMessage message) {
        // This runs on: Virtual Thread #8 (from callback executor)
        System.out.println("Balance updated: " + message.getBalance());
    }
    
    @Override
    public void onEventDelta(EventDeltaMessage message) {
        // This runs on: Virtual Thread #9 (from callback executor)
        System.out.println("Event delta received");
    }
    
    @Override
    public void onBroadcast(BroadcastNotificationMessage message) {
        // This runs on: Virtual Thread #10 (from callback executor)
        System.out.println("Broadcast: " + message.getMessage());
    }
    
    @Override
    public void onConnectionLost(Throwable error) {
        // This runs on: Virtual Thread #11 (from callback executor)
        // Triggered by: WebSocket error or heartbeat timeout
        System.err.println("Connection lost: " + error.getMessage());
    }
}
```

### 🧵 Thread Execution Flow Explained

When you run the example above, here's what happens behind the scenes:

```
Time    Thread              Action
────────────────────────────────────────────────────────────────────
T0      Main Thread         Creates ConectorBolsa
                           → Initializes MessageSequencer (idle)
                           → Initializes CallbackExecutor (idle)
                           → No threads running yet

T1      Main Thread         Calls addListener()
                           → Adds to CopyOnWriteArrayList
                           → Still no threads running

T2      Main Thread         Calls conectar()
                           → Starts WebSocket connection
                           → Starts HeartbeatManager
        Heartbeat Thread    Starts scheduling PING messages
        WebSocket Thread    Connects and waits for messages
        Sequencer Thread    Starts and waits for messages

T3      Main Thread         Calls enviarOrden()
                           → Acquires semaphore
                           → Sends via WebSocket
                           → Releases semaphore
                           → Returns immediately

T4      WebSocket Thread    Receives FILL message
                           → Passes to MessageSequencer
        Sequencer Thread    Deserializes JSON → FillMessage
                           → Submits callback to executor
                           → Continues to next message
        Virtual Thread #1   Executes onFill() callback
                           → Your code runs here

T5      WebSocket Thread    Receives TICKER message
                           → Passes to MessageSequencer
        Sequencer Thread    Deserializes JSON → TickerMessage
                           → Submits callback to executor
        Virtual Thread #2   Executes onTicker() callback
                           → Runs concurrently with #1!

T6      Heartbeat Thread    Sends PING message
                           → Acquires send semaphore
                           → Sends via WebSocket
                           → Releases semaphore

T7      Main Thread         Calls desconectar()
                           → Stops heartbeat
                           → Closes WebSocket
        Main Thread         Calls shutdown()
                           → Stops sequencer
                           → Stops callback executor
                           → Waits for threads to finish
                           → All threads terminated ✅
```

### 🔑 Key Takeaways

1. **Initialization is cheap**: Creating `ConectorBolsa()` just allocates structures
2. **Connection activates threads**: `conectar()` starts the thread pools
3. **Sending is thread-safe**: Multiple threads can call `enviarOrden()` safely
4. **Messages are ordered**: Sequencer guarantees FILL #1 is processed before FILL #2
5. **Callbacks are concurrent**: `onFill()` and `onTicker()` can run simultaneously
6. **Shutdown is clean**: All threads terminate gracefully

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
