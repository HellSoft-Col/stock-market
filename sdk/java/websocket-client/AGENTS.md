# AGENTS.MD - Stock Market Java SDK Development Guide

## Project Overview

Building a Java 25 WebSocket-based client SDK for connecting to the Stock Market Server. The SDK handles low-level communication, message serialization/deserialization, connection management, and threading, allowing client applications to focus on business logic.

**Package**: `tech.hellsoft.trading`

## Core Technology Stack

- **Java 25** - Latest language features
- **Virtual Threads** - For all concurrent operations
- **Built-in WebSocket Client** - `java.net.http.WebSocket`
- **Gson** - JSON serialization (Google's library)
- **Lombok** - Reduce boilerplate
- **JUnit 5** - Testing framework

## Project Structure
```
src/main/java/tech/hellsoft/trading/
├── ConectorBolsa.java           (Main SDK class)
├── EventListener.java           (Callback interface)
├── enums/
│   ├── MessageType.java
│   ├── OrderSide.java
│   ├── OrderMode.java
│   ├── Product.java
│   ├── ErrorCode.java
│   ├── OrderStatus.java
│   └── RecipeType.java
├── dto/
│   ├── client/                  (Client → Server)
│   │   ├── LoginMessage.java
│   │   ├── OrderMessage.java
│   │   ├── ProductionUpdateMessage.java
│   │   ├── AcceptOfferMessage.java
│   │   ├── ResyncMessage.java
│   │   └── CancelMessage.java
│   └── server/                  (Server → Client)
│       ├── LoginOKMessage.java
│       ├── FillMessage.java
│       ├── TickerMessage.java
│       ├── OfferMessage.java
│       ├── ErrorMessage.java
│       ├── OrderAckMessage.java
│       ├── InventoryUpdateMessage.java
│       ├── BalanceUpdateMessage.java
│       ├── EventDeltaMessage.java
│       └── BroadcastNotificationMessage.java
├── exception/
│   ├── ConexionFallidaException.java
│   └── ValidationException.java
└── internal/                    (NOT exported)
    ├── WebSocketHandler.java
    ├── MessageRouter.java
    └── JsonSerializer.java
```

## Build Configuration

### build.gradle.kts
```kotlin
plugins {
    `java-library`
    `maven-publish`
}

group = "tech.hellsoft.trading"
version = "1.0.0-SNAPSHOT"

java {
    toolchain {
        languageVersion = JavaLanguageVersion.of(25)
    }
    withSourcesJar()
    withJavadocJar()
}

repositories {
    mavenCentral()
}

dependencies {
    // JSON processing
    implementation("com.google.code.gson:gson:2.11.0")
    
    // Lombok
    compileOnly("org.projectlombok:lombok:1.18.30")
    annotationProcessor("org.projectlombok:lombok:1.18.30")
    
    // Logging
    implementation("org.slf4j:slf4j-api:2.0.9")
    
    // Testing
    testImplementation(platform("org.junit:junit-bom:5.10.0"))
    testImplementation("org.junit.jupiter:junit-jupiter")
    testImplementation("org.mockito:mockito-core:5.8.0")
    testCompileOnly("org.projectlombok:lombok:1.18.30")
    testAnnotationProcessor("org.projectlombok:lombok:1.18.30")
}

tasks.test {
    useJUnitPlatform()
}
```

### module-info.java
```java
module tech.hellsoft.trading {
    // Export public API only
    exports tech.hellsoft.trading;
    exports tech.hellsoft.trading.enums;
    exports tech.hellsoft.trading.dto.client;
    exports tech.hellsoft.trading.dto.server;
    exports tech.hellsoft.trading.exception;
    
    // DO NOT export internal package
    // exports tech.hellsoft.trading.internal; // ❌
    
    requires java.net.http;
    requires com.google.gson;
    requires static lombok;
    requires org.slf4j;
}
```

## Critical Code Style Rules

### 1. NO ELSE STATEMENTS - Always Use Guard Clauses
```java
// ✅ CORRECT: Guard clauses
public void enviarOrden(OrderMessage orden) {
    if (!isLoggedIn) {
        throw new IllegalStateException("Not logged in");
    }
    
    if (orden == null) {
        throw new IllegalArgumentException("Order cannot be null");
    }
    
    if (orden.getQty() <= 0) {
        throw new ValidationException("Quantity must be positive");
    }
    
    sendMessage(orden);
}

// ❌ WRONG: Nested if-else
public void enviarOrden(OrderMessage orden) {
    if (isLoggedIn) {
        if (orden != null) {
            if (orden.getQty() > 0) {
                sendMessage(orden);
            } else {
                throw new ValidationException("Quantity must be positive");
            }
        } else {
            throw new IllegalArgumentException("Order cannot be null");
        }
    } else {
        throw new IllegalStateException("Not logged in");
    }
}
```

### 2. Prefer Functional Programming Methods
```java
// ✅ CORRECT: Use streams and functional methods
public List<Product> getAuthorizedProducts(LoginOKMessage loginOk) {
    return loginOk.getAuthorizedProducts().stream()
        .filter(Product::isAvailable)
        .sorted(Comparator.comparing(Product::getValue))
        .collect(Collectors.toList());
}

// ✅ CORRECT: Use forEach for side effects
private void notifyAllListeners(Consumer<EventListener> action) {
    listeners.forEach(listener ->
        executor.execute(() -> safeInvoke(listener, action))
    );
}

// ✅ CORRECT: Use Optional instead of null checks
public Optional<Double> getMarketPrice(Product product) {
    return Optional.ofNullable(prices.get(product));
}

// ❌ WRONG: Imperative loops
public List<Product> getAuthorizedProducts(LoginOKMessage loginOk) {
    List<Product> result = new ArrayList<>();
    for (Product p : loginOk.getAuthorizedProducts()) {
        if (p.isAvailable()) {
            result.add(p);
        }
    }
    return result;
}
```

### 3. Use Lombok Extensively
```java
// ✅ DTOs with Lombok
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class OrderMessage {
    private MessageType type;
    private String clOrdID;
    private OrderSide side;
    private OrderMode mode;
    private Product product;
    private Integer qty;
    private Double limitPrice;
    private String expiresAt;
    private String message;
}

// ✅ Exceptions with Lombok
@Getter
public class ConexionFallidaException extends Exception {
    private final String host;
    private final int port;
    
    public ConexionFallidaException(String message, String host, int port) {
        super(message);
        this.host = host;
        this.port = port;
    }
}

// ✅ Logging with Lombok
@Slf4j
public class ConectorBolsa {
    public void conectar(String host, int port) {
        log.info("Connecting to {}:{}", host, port);
        // ...
    }
}
```

## Enumeration Patterns

All enumerations MUST support JSON serialization with custom string values:
```java
public enum MessageType {
    LOGIN("LOGIN"),
    ORDER("ORDER"),
    FILL("FILL"),
    TICKER("TICKER");
    
    private final String value;
    
    MessageType(String value) {
        this.value = value;
    }
    
    @JsonValue
    public String toJson() {
        return value;
    }
    
    @JsonCreator
    public static MessageType fromJson(String value) {
        return Arrays.stream(values())
            .filter(type -> type.value.equals(value))
            .findFirst()
            .orElseThrow(() -> new IllegalArgumentException("Unknown type: " + value));
    }
}
```

**Special case - Product enum with hyphens:**
```java
public enum Product {
    GUACA("GUACA"),
    SEBO("SEBO"),
    PALTA_OIL("PALTA-OIL"),      // Underscore in Java, hyphen in JSON
    CASCAR_ALLOY("CASCAR-ALLOY"); // Underscore in Java, hyphen in JSON
    
    private final String value;
    
    Product(String value) {
        this.value = value;
    }
    
    @JsonValue
    public String getValue() {
        return value;
    }
    
    @JsonCreator
    public static Product fromJson(String value) {
        return Arrays.stream(values())
            .filter(p -> p.value.equals(value))
            .findFirst()
            .orElseThrow(() -> new IllegalArgumentException("Unknown product: " + value));
    }
}
```

## Gson Configuration

### Singleton Gson Instance with Custom Settings
```java
@Getter
public class JsonSerializer {
    private static final Gson GSON = new GsonBuilder()
        .setFieldNamingPolicy(FieldNamingPolicy.LOWER_CASE_WITH_UNDERSCORES)
        .registerTypeAdapter(Instant.class, new InstantAdapter())
        .setPrettyPrinting() // For debugging
        .create();
    
    public static String toJson(Object obj) {
        return GSON.toJson(obj);
    }
    
    public static <T> T fromJson(String json, Class<T> clazz) {
        return GSON.fromJson(json, clazz);
    }
    
    // For generic types
    public static <T> T fromJson(String json, Type type) {
        return GSON.fromJson(json, type);
    }
}

// Custom adapter for Instant
class InstantAdapter extends TypeAdapter<Instant> {
    @Override
    public void write(JsonWriter out, Instant value) throws IOException {
        if (value == null) {
            out.nullValue();
            return;
        }
        out.value(value.toString());
    }
    
    @Override
    public Instant read(JsonReader in) throws IOException {
        if (in.peek() == JsonToken.NULL) {
            in.nextNull();
            return null;
        }
        return Instant.parse(in.nextString());
    }
}
```

## Threading Architecture

### Virtual Threads for All Concurrent Operations
```java
@Slf4j
public class ConectorBolsa {
    private final ExecutorService callbackExecutor = 
        Executors.newVirtualThreadPerTaskExecutor();
    
    private final List<EventListener> listeners = new CopyOnWriteArrayList<>();
    
    private void notifyFill(FillMessage fill) {
        listeners.forEach(listener ->
            callbackExecutor.execute(() -> {
                try {
                    listener.onFill(fill);
                } catch (Exception e) {
                    log.error("Listener error in onFill", e);
                }
            })
        );
    }
}
```

### WebSocket Message Reception
```java
private class WebSocketHandler implements WebSocket.Listener {
    private final StringBuilder messageBuffer = new StringBuilder();
    
    @Override
    public CompletionStage<?> onText(WebSocket webSocket, 
                                      CharSequence data, 
                                      boolean last) {
        messageBuffer.append(data);
        
        if (last) {
            String json = messageBuffer.toString();
            messageBuffer.setLength(0);
            
            // Process on virtual thread
            callbackExecutor.execute(() -> processMessage(json));
        }
        
        webSocket.request(1);
        return null;
    }
    
    @Override
    public void onError(WebSocket webSocket, Throwable error) {
        log.error("WebSocket error", error);
        notifyConnectionLost(error);
    }
}
```

## Message Routing Pattern
```java
private void processMessage(String json) {
    try {
        // First parse to determine type
        JsonObject obj = JsonParser.parseString(json).getAsJsonObject();
        String typeStr = obj.get("type").getAsString();
        MessageType type = MessageType.fromJson(typeStr);
        
        // Route based on type using switch expression
        switch (type) {
            case LOGIN_OK -> {
                LoginOKMessage msg = JsonSerializer.fromJson(json, LoginOKMessage.class);
                listeners.forEach(l -> l.onLoginOk(msg));
            }
            case FILL -> {
                FillMessage msg = JsonSerializer.fromJson(json, FillMessage.class);
                listeners.forEach(l -> l.onFill(msg));
            }
            case TICKER -> {
                TickerMessage msg = JsonSerializer.fromJson(json, TickerMessage.class);
                listeners.forEach(l -> l.onTicker(msg));
            }
            case ERROR -> {
                ErrorMessage msg = JsonSerializer.fromJson(json, ErrorMessage.class);
                listeners.forEach(l -> l.onError(msg));
            }
            case OFFER -> {
                OfferMessage msg = JsonSerializer.fromJson(json, OfferMessage.class);
                listeners.forEach(l -> l.onOffer(msg));
            }
            case BROADCAST_NOTIFICATION -> {
                BroadcastNotificationMessage msg = 
                    JsonSerializer.fromJson(json, BroadcastNotificationMessage.class);
                listeners.forEach(l -> l.onBroadcast(msg));
            }
            default -> log.warn("Unknown message type: {}", type);
        }
    } catch (Exception e) {
        log.error("Failed to process message: {}", json, e);
    }
}
```

## Validation Patterns

### Client-Side Validation Before Sending
```java
public void enviarOrden(OrderMessage orden) {
    if (!isLoggedIn) {
        throw new IllegalStateException("Must be logged in to send orders");
    }
    
    validateOrder(orden);
    sendMessage(orden);
}

private void validateOrder(OrderMessage orden) {
    if (orden.getClOrdID() == null || orden.getClOrdID().isBlank()) {
        throw new ValidationException("clOrdID is required");
    }
    
    if (orden.getSide() == null) {
        throw new ValidationException("side is required");
    }
    
    if (orden.getMode() == null) {
        throw new ValidationException("mode is required");
    }
    
    if (orden.getProduct() == null) {
        throw new ValidationException("product is required");
    }
    
    if (orden.getQty() == null || orden.getQty() <= 0) {
        throw new ValidationException("quantity must be positive");
    }
    
    if (orden.getMode() == OrderMode.LIMIT) {
        if (orden.getLimitPrice() == null || orden.getLimitPrice() <= 0) {
            throw new ValidationException("LIMIT orders require positive limitPrice");
        }
    }
    
    if (orden.getMessage() != null && orden.getMessage().length() > 200) {
        throw new ValidationException("message exceeds 200 characters");
    }
}
```

## Thread-Safe Message Sending
```java
private final Semaphore sendSemaphore = new Semaphore(1);
private volatile WebSocket webSocket;

private void sendMessage(Object message) {
    if (webSocket == null || webSocket.isOutputClosed()) {
        throw new IllegalStateException("WebSocket not connected");
    }
    
    try {
        sendSemaphore.acquire();
        try {
            String json = JsonSerializer.toJson(message);
            log.debug("Sending: {}", json);
            webSocket.sendText(json, true).get();
        } finally {
            sendSemaphore.release();
        }
    } catch (InterruptedException e) {
        Thread.currentThread().interrupt();
        throw new RuntimeException("Send interrupted", e);
    } catch (Exception e) {
        throw new RuntimeException("Failed to send message", e);
    }
}
```

## Error Handling Strategy

### Structured Error Information
```java
@Getter
public enum ErrorCode {
    AUTH_FAILED("AUTH_FAILED", Severity.FATAL),
    INVALID_ORDER("INVALID_ORDER", Severity.ERROR),
    INSUFFICIENT_INVENTORY("INSUFFICIENT_INVENTORY", Severity.ERROR),
    RATE_LIMIT_EXCEEDED("RATE_LIMIT_EXCEEDED", Severity.WARNING),
    OFFER_EXPIRED("OFFER_EXPIRED", Severity.INFO);
    
    private final String value;
    private final Severity severity;
    
    ErrorCode(String value, Severity severity) {
        this.value = value;
        this.severity = severity;
    }
    
    public enum Severity {
        FATAL,      // Must terminate
        ERROR,      // Operation failed
        WARNING,    // Degraded
        INFO        // Informational
    }
}
```

## Functional Callback Patterns
```java
// ✅ Use functional interfaces for cleaner callbacks
private void notifyListeners(Consumer<EventListener> action) {
    listeners.stream()
        .forEach(listener ->
            callbackExecutor.execute(() -> safeInvoke(listener, action))
        );
}

private void safeInvoke(EventListener listener, Consumer<EventListener> action) {
    try {
        action.accept(listener);
    } catch (Exception e) {
        log.error("Listener callback failed", e);
    }
}

// Usage examples
private void notifyFill(FillMessage fill) {
    notifyListeners(listener -> listener.onFill(fill));
}

private void notifyTicker(TickerMessage ticker) {
    notifyListeners(listener -> listener.onTicker(ticker));
}

private void notifyError(ErrorMessage error) {
    notifyListeners(listener -> listener.onError(error));
}
```

## Immutability and Concurrency
```java
// ✅ Use immutable collections where possible
public class LoginOKMessage {
    private final String team;
    private final double initialBalance;
    private final Map<Product, Integer> inventory;
    private final List<Product> authorizedProducts;
    
    // Return unmodifiable views
    public Map<Product, Integer> getInventory() {
        return Collections.unmodifiableMap(inventory);
    }
    
    public List<Product> getAuthorizedProducts() {
        return Collections.unmodifiableList(authorizedProducts);
    }
}

// ✅ Use concurrent collections for mutable state
private final Map<Product, Double> currentPrices = new ConcurrentHashMap<>();
private final List<EventListener> listeners = new CopyOnWriteArrayList<>();
```

## Testing Best Practices

### Unit Tests with Mockito
```java
@Slf4j
class ConectorBolsaTest {
    
    @Test
    void shouldValidateOrderBeforeSending() {
        ConectorBolsa connector = new ConectorBolsa();
        
        OrderMessage invalidOrder = OrderMessage.builder()
            .type(MessageType.ORDER)
            .side(OrderSide.BUY)
            .qty(0) // Invalid!
            .build();
        
        assertThrows(ValidationException.class, 
            () -> connector.enviarOrden(invalidOrder));
    }
    
    @Test
    void shouldNotifyListenersOnFill() throws Exception {
        ConectorBolsa connector = new ConectorBolsa();
        CountDownLatch latch = new CountDownLatch(1);
        
        EventListener listener = new EventListener() {
            @Override
            public void onFill(FillMessage fill) {
                assertEquals(10, fill.getFillQty());
                assertEquals(Product.FOSFO, fill.getProduct());
                latch.countDown();
            }
            
            // Implement other required methods...
        };
        
        connector.addListener(listener);
        
        // Simulate fill message
        FillMessage fill = FillMessage.builder()
            .type(MessageType.FILL)
            .fillQty(10)
            .product(Product.FOSFO)
            .build();
        
        connector.simulateFill(fill); // Test method
        
        assertTrue(latch.await(1, TimeUnit.SECONDS));
    }
}
```

## Common Patterns Summary

### Guard Clauses (NO ELSE)
```java
// Always validate at the beginning, return/throw early
if (invalid) {
    throw exception;
}

if (anotherInvalid) {
    throw exception;
}

// Main logic here
```

### Functional Methods
```java
// Use streams, map, filter, forEach instead of loops
list.stream()
    .filter(predicate)
    .map(function)
    .collect(Collectors.toList());
```

### Virtual Threads
```java
// Always use virtual threads for concurrency
ExecutorService executor = Executors.newVirtualThreadPerTaskExecutor();
```

### Lombok
```java
// Use @Data, @Builder, @Getter, @Slf4j, @AllArgsConstructor, @NoArgsConstructor
@Data
@Builder
@Slf4j
public class MyClass { }
```

### Enums with JSON
```java
// Always use @JsonValue and @JsonCreator for enum serialization
@JsonValue
public String toJson() { return value; }

@JsonCreator
public static MyEnum fromJson(String value) { ... }
```

## Critical Reminders

- ☐ NO else statements - use guard clauses
- ☐ Prefer functional programming (streams, lambdas, Optional)
- ☐ Use virtual threads for ALL concurrency
- ☐ Use Lombok to reduce boilerplate
- ☐ All enums must support JSON serialization
- ☐ Package is `tech.hellsoft.trading`
- ☐ Only export public API in module-info.java
- ☐ Use Gson for JSON with custom configuration
- ☐ Thread-safe message sending with Semaphore
- ☐ Notify listeners on virtual threads
- ☐ Validate early, fail fast
- ☐ Use `@Slf4j` for logging
- ☐ Immutable DTOs with @Builder
- ☐ CopyOnWriteArrayList for listeners
- ☐ ConcurrentHashMap for shared state

## Documentation Standards

Every public API method must have Javadoc:
```java
/**
 * Sends a buy or sell order to the market.
 * 
 * <p>The order is validated before sending. If validation fails, a
 * {@link ValidationException} is thrown. The order response arrives
 * asynchronously via {@link EventListener#onFill(FillMessage)}.
 * 
 * @param orden the order to send (must not be null)
 * @throws IllegalStateException if not logged in
 * @throws ValidationException if order validation fails
 * @see EventListener#onFill(FillMessage)
 */
public void enviarOrden(OrderMessage orden) {
    // ...
}
```

---

This guide provides the essential patterns and practices for building the Stock Market SDK with modern Java 25 features, following functional programming principles and clean code standards.