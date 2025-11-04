# Java SDK Implementation Plan - Bolsa Interestelar de Aguacates Andorianos

**Complete Implementation Guide**  
**Package**: `tech.hellsoft.trading`  
**Language**: Java 25  
**Version**: 1.0.0

---

## Table of Contents

1. [Project Overview](#project-overview)
2. [Technology Stack](#technology-stack)
3. [Project Structure](#project-structure)
4. [Phase 1: Foundation & Enumerations](#phase-1-foundation--enumerations)
5. [Phase 2: Configuration System](#phase-2-configuration-system)
6. [Phase 3: Exception Classes](#phase-3-exception-classes)
7. [Phase 4: Data Transfer Objects](#phase-4-data-transfer-objects)
8. [Phase 5: Internal Infrastructure - State Management](#phase-5-internal-infrastructure---state-management)
9. [Phase 6: Internal Infrastructure - Message Processing](#phase-6-internal-infrastructure---message-processing)
10. [Phase 7: Connection Management](#phase-7-connection-management)
11. [Phase 8: Public API](#phase-8-public-api)
12. [Phase 9: Testing](#phase-9-testing)
13. [Phase 10: Documentation](#phase-10-documentation)
14. [Appendices](#appendices)

---

## Project Overview

This SDK provides a complete WebSocket-based client for the Bolsa Interestelar de Aguacates Andorianos trading system. It handles low-level communication, message serialization/deserialization, connection management, and threading, allowing client applications to focus on business logic.

### Key Features

✅ **Automatic connection management** with heartbeat  
✅ **Auto-reconnect** with exponential backoff  
✅ **Thread-safe message processing** with sequencing  
✅ **State mutation locking** for callback safety  
✅ **Virtual threads** for all concurrent operations  
✅ **Type-safe enumerations** with JSON support  
✅ **Flexible configuration** system  
✅ **Clean, elegant API** (no manual connection loops)

### What the SDK Does

- Manages WebSocket connection lifecycle
- Handles message serialization/deserialization
- Thread-safe message sending
- Asynchronous message reception with callbacks
- Validates message format before sending
- Provides structured error information
- Sequential message processing by type
- Automatic heartbeat pings
- Automatic reconnection with resync

### What the SDK Does NOT Do

- Business logic (trading algorithms, P&L calculation)
- State management (inventory tracking, balance tracking)
- Persistence (snapshots, configuration files)
- User interface (console commands, displays)
- Trading strategy (when to buy/sell, pricing decisions)

---

## Technology Stack

- **Java 25** - Latest language features
- **Virtual Threads** - For all concurrent operations
- **Built-in WebSocket Client** - `java.net.http.WebSocket`
- **Gson** - JSON serialization (Google's library)
- **Lombok** - Reduce boilerplate
- **JUnit 5** - Testing framework
- **SLF4J** - Logging facade

---

## Project Structure

```
src/main/java/tech/hellsoft/trading/
├── ConectorBolsa.java           (Main SDK class)
├── EventListener.java           (Callback interface)
├── config/
│   └── ConectorConfig.java
├── enums/
│   ├── MessageType.java
│   ├── OrderSide.java
│   ├── OrderMode.java
│   ├── Product.java
│   ├── ErrorCode.java
│   ├── OrderStatus.java
│   ├── RecipeType.java
│   └── ConnectionState.java
├── dto/
│   ├── client/                  (Client → Server)
│   │   ├── LoginMessage.java
│   │   ├── OrderMessage.java
│   │   ├── ProductionUpdateMessage.java
│   │   ├── AcceptOfferMessage.java
│   │   ├── ResyncMessage.java
│   │   ├── CancelMessage.java
│   │   └── PingMessage.java
│   └── server/                  (Server → Client)
│       ├── Recipe.java
│       ├── TeamRole.java
│       ├── LoginOKMessage.java
│       ├── FillMessage.java
│       ├── TickerMessage.java
│       ├── OfferMessage.java
│       ├── ErrorMessage.java
│       ├── OrderAckMessage.java
│       ├── InventoryUpdateMessage.java
│       ├── BalanceUpdateMessage.java
│       ├── EventDeltaMessage.java
│       ├── BroadcastNotificationMessage.java
│       └── PongMessage.java
├── exception/
│   ├── ConexionFallidaException.java
│   ├── ValidationException.java
│   └── StateLockException.java
└── internal/                    (NOT exported)
    ├── connection/
    │   ├── WebSocketHandler.java
    │   ├── ConnectionManager.java
    │   └── HeartbeatManager.java
    ├── routing/
    │   ├── MessageRouter.java
    │   ├── StateLocker.java
    │   └── MessageSequencer.java
    └── serialization/
        └── JsonSerializer.java
```

---

## Phase 1: Foundation & Enumerations

### Step 1.1: Build Configuration

**File**: `build.gradle.kts`

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
    implementation("com.google.code.gson:gson:2.11.0")
    compileOnly("org.projectlombok:lombok:1.18.30")
    annotationProcessor("org.projectlombok:lombok:1.18.30")
    implementation("org.slf4j:slf4j-api:2.0.9")
    
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

### Step 1.2: Module Configuration

**File**: `src/main/java/module-info.java`

```java
module tech.hellsoft.trading {
    exports tech.hellsoft.trading;
    exports tech.hellsoft.trading.enums;
    exports tech.hellsoft.trading.dto.client;
    exports tech.hellsoft.trading.dto.server;
    exports tech.hellsoft.trading.exception;
    exports tech.hellsoft.trading.config;
    
    requires java.net.http;
    requires com.google.gson;
    requires static lombok;
    requires org.slf4j;
}
```

### Step 1.3: Enumerations

#### MessageType Enum

**File**: `src/main/java/tech/hellsoft/trading/enums/MessageType.java`

```java
package tech.hellsoft.trading.enums;

import com.google.gson.annotations.JsonCreator;
import com.google.gson.annotations.SerializedName;

import java.util.Arrays;

public enum MessageType {
    @SerializedName("LOGIN") LOGIN("LOGIN"),
    @SerializedName("LOGIN_OK") LOGIN_OK("LOGIN_OK"),
    @SerializedName("ORDER") ORDER("ORDER"),
    @SerializedName("ORDER_ACK") ORDER_ACK("ORDER_ACK"),
    @SerializedName("FILL") FILL("FILL"),
    @SerializedName("TICKER") TICKER("TICKER"),
    @SerializedName("OFFER") OFFER("OFFER"),
    @SerializedName("ACCEPT_OFFER") ACCEPT_OFFER("ACCEPT_OFFER"),
    @SerializedName("PRODUCTION_UPDATE") PRODUCTION_UPDATE("PRODUCTION_UPDATE"),
    @SerializedName("INVENTORY_UPDATE") INVENTORY_UPDATE("INVENTORY_UPDATE"),
    @SerializedName("BALANCE_UPDATE") BALANCE_UPDATE("BALANCE_UPDATE"),
    @SerializedName("RESYNC") RESYNC("RESYNC"),
    @SerializedName("EVENT_DELTA") EVENT_DELTA("EVENT_DELTA"),
    @SerializedName("CANCEL") CANCEL("CANCEL"),
    @SerializedName("ERROR") ERROR("ERROR"),
    @SerializedName("BROADCAST_NOTIFICATION") BROADCAST_NOTIFICATION("BROADCAST_NOTIFICATION"),
    @SerializedName("PING") PING("PING"),
    @SerializedName("PONG") PONG("PONG");
    
    private final String value;
    
    MessageType(String value) {
        this.value = value;
    }
    
    public String getValue() {
        return value;
    }
    
    @JsonCreator
    public static MessageType fromJson(String value) {
        return Arrays.stream(values())
            .filter(type -> type.value.equals(value))
            .findFirst()
            .orElseThrow(() -> new IllegalArgumentException("Unknown message type: " + value));
    }
}
```

#### OrderSide Enum

**File**: `src/main/java/tech/hellsoft/trading/enums/OrderSide.java`

```java
package tech.hellsoft.trading.enums;

import com.google.gson.annotations.JsonCreator;
import com.google.gson.annotations.SerializedName;

import java.util.Arrays;

public enum OrderSide {
    @SerializedName("BUY") BUY("BUY"),
    @SerializedName("SELL") SELL("SELL");
    
    private final String value;
    
    OrderSide(String value) {
        this.value = value;
    }
    
    public String getValue() {
        return value;
    }
    
    @JsonCreator
    public static OrderSide fromJson(String value) {
        return Arrays.stream(values())
            .filter(side -> side.value.equals(value))
            .findFirst()
            .orElseThrow(() -> new IllegalArgumentException("Unknown order side: " + value));
    }
}
```

#### OrderMode Enum

**File**: `src/main/java/tech/hellsoft/trading/enums/OrderMode.java`

```java
package tech.hellsoft.trading.enums;

import com.google.gson.annotations.JsonCreator;
import com.google.gson.annotations.SerializedName;

import java.util.Arrays;

public enum OrderMode {
    @SerializedName("MARKET") MARKET("MARKET"),
    @SerializedName("LIMIT") LIMIT("LIMIT");
    
    private final String value;
    
    OrderMode(String value) {
        this.value = value;
    }
    
    public String getValue() {
        return value;
    }
    
    @JsonCreator
    public static OrderMode fromJson(String value) {
        return Arrays.stream(values())
            .filter(mode -> mode.value.equals(value))
            .findFirst()
            .orElseThrow(() -> new IllegalArgumentException("Unknown order mode: " + value));
    }
}
```

#### Product Enum

**File**: `src/main/java/tech/hellsoft/trading/enums/Product.java`

```java
package tech.hellsoft.trading.enums;

import com.google.gson.annotations.JsonCreator;
import com.google.gson.annotations.SerializedName;

import java.util.Arrays;
import java.util.Set;
import java.util.stream.Collectors;

public enum Product {
    @SerializedName("GUACA") GUACA("GUACA"),
    @SerializedName("SEBO") SEBO("SEBO"),
    @SerializedName("PALTA-OIL") PALTA_OIL("PALTA-OIL"),
    @SerializedName("FOSFO") FOSFO("FOSFO"),
    @SerializedName("NUCREM") NUCREM("NUCREM"),
    @SerializedName("CASCAR-ALLOY") CASCAR_ALLOY("CASCAR-ALLOY"),
    @SerializedName("PITA") PITA("PITA");
    
    private final String value;
    
    Product(String value) {
        this.value = value;
    }
    
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
    
    public static Set<String> getAllValues() {
        return Arrays.stream(values())
            .map(Product::getValue)
            .collect(Collectors.toSet());
    }
}
```

#### ErrorCode Enum

**File**: `src/main/java/tech/hellsoft/trading/enums/ErrorCode.java`

```java
package tech.hellsoft.trading.enums;

import com.google.gson.annotations.JsonCreator;
import com.google.gson.annotations.SerializedName;

import java.util.Arrays;

public enum ErrorCode {
    @SerializedName("AUTH_FAILED") AUTH_FAILED("AUTH_FAILED", Severity.FATAL),
    @SerializedName("INVALID_ORDER") INVALID_ORDER("INVALID_ORDER", Severity.ERROR),
    @SerializedName("INVALID_PRODUCT") INVALID_PRODUCT("INVALID_PRODUCT", Severity.ERROR),
    @SerializedName("INVALID_QUANTITY") INVALID_QUANTITY("INVALID_QUANTITY", Severity.ERROR),
    @SerializedName("DUPLICATE_ORDER_ID") DUPLICATE_ORDER_ID("DUPLICATE_ORDER_ID", Severity.ERROR),
    @SerializedName("UNAUTHORIZED_PRODUCTION") UNAUTHORIZED_PRODUCTION("UNAUTHORIZED_PRODUCTION", Severity.ERROR),
    @SerializedName("OFFER_EXPIRED") OFFER_EXPIRED("OFFER_EXPIRED", Severity.INFO),
    @SerializedName("RATE_LIMIT_EXCEEDED") RATE_LIMIT_EXCEEDED("RATE_LIMIT_EXCEEDED", Severity.WARNING),
    @SerializedName("SERVICE_UNAVAILABLE") SERVICE_UNAVAILABLE("SERVICE_UNAVAILABLE", Severity.TRANSIENT),
    @SerializedName("INSUFFICIENT_INVENTORY") INSUFFICIENT_INVENTORY("INSUFFICIENT_INVENTORY", Severity.ERROR),
    @SerializedName("INVALID_MESSAGE") INVALID_MESSAGE("INVALID_MESSAGE", Severity.ERROR);
    
    private final String value;
    private final Severity severity;
    
    ErrorCode(String value, Severity severity) {
        this.value = value;
        this.severity = severity;
    }
    
    public String getValue() {
        return value;
    }
    
    public Severity getSeverity() {
        return severity;
    }
    
    @JsonCreator
    public static ErrorCode fromJson(String value) {
        return Arrays.stream(values())
            .filter(code -> code.value.equals(value))
            .findFirst()
            .orElse(null);
    }
    
    public enum Severity {
        FATAL,      // Must terminate application
        ERROR,      // Operation failed, can continue
        WARNING,    // Degraded performance
        INFO,       // Informational
        TRANSIENT   // Temporary, retry recommended
    }
}
```

#### OrderStatus Enum

**File**: `src/main/java/tech/hellsoft/trading/enums/OrderStatus.java`

```java
package tech.hellsoft.trading.enums;

import com.google.gson.annotations.JsonCreator;
import com.google.gson.annotations.SerializedName;

import java.util.Arrays;

public enum OrderStatus {
    @SerializedName("PENDING") PENDING("PENDING"),
    @SerializedName("FILLED") FILLED("FILLED"),
    @SerializedName("PARTIALLY_FILLED") PARTIALLY_FILLED("PARTIALLY_FILLED"),
    @SerializedName("CANCELLED") CANCELLED("CANCELLED");
    
    private final String value;
    
    OrderStatus(String value) {
        this.value = value;
    }
    
    public String getValue() {
        return value;
    }
    
    @JsonCreator
    public static OrderStatus fromJson(String value) {
        return Arrays.stream(values())
            .filter(status -> status.value.equals(value))
            .findFirst()
            .orElseThrow(() -> new IllegalArgumentException("Unknown order status: " + value));
    }
}
```

#### RecipeType Enum

**File**: `src/main/java/tech/hellsoft/trading/enums/RecipeType.java`

```java
package tech.hellsoft.trading.enums;

import com.google.gson.annotations.JsonCreator;
import com.google.gson.annotations.SerializedName;

import java.util.Arrays;

public enum RecipeType {
    @SerializedName("BASIC") BASIC("BASIC"),
    @SerializedName("PREMIUM") PREMIUM("PREMIUM");
    
    private final String value;
    
    RecipeType(String value) {
        this.value = value;
    }
    
    public String getValue() {
        return value;
    }
    
    @JsonCreator
    public static RecipeType fromJson(String value) {
        return Arrays.stream(values())
            .filter(type -> type.value.equals(value))
            .findFirst()
            .orElseThrow(() -> new IllegalArgumentException("Unknown recipe type: " + value));
    }
}
```

#### ConnectionState Enum

**File**: `src/main/java/tech/hellsoft/trading/enums/ConnectionState.java`

```java
package tech.hellsoft.trading.enums;

public enum ConnectionState {
    DISCONNECTED,    // Initial state or after disconnect
    CONNECTING,      // Connection in progress
    CONNECTED,       // WebSocket connected but not authenticated
    AUTHENTICATED,   // Logged in and ready
    RECONNECTING,    // Attempting to reconnect after loss
    CLOSED           // Permanently closed
}
```

---

## Phase 2: Configuration System

### ConectorConfig

**File**: `src/main/java/tech/hellsoft/trading/config/ConectorConfig.java`

```java
package tech.hellsoft.trading.config;

import lombok.Builder;
import lombok.Data;
import lombok.extern.slf4j.Slf4j;

import java.time.Duration;

/**
 * Configuration for ConectorBolsa SDK behavior.
 * 
 * <p>Use the builder pattern to customize SDK behavior:
 * <pre>{@code
 * ConectorConfig config = ConectorConfig.builder()
 *     .heartbeatInterval(Duration.ofSeconds(30))
 *     .connectionTimeout(Duration.ofSeconds(10))
 *     .autoReconnect(true)
 *     .maxReconnectAttempts(5)
 *     .enableMessageSequencing(true)
 *     .enableStateLocking(true)
 *     .build();
 * }</pre>
 */
@Data
@Builder
@Slf4j
public class ConectorConfig {
    
    /**
     * Interval between heartbeat pings to keep connection alive.
     * Default: 30 seconds
     */
    @Builder.Default
    private Duration heartbeatInterval = Duration.ofSeconds(30);
    
    /**
     * Timeout for initial connection establishment.
     * Default: 10 seconds
     */
    @Builder.Default
    private Duration connectionTimeout = Duration.ofSeconds(10);
    
    /**
     * Enable automatic reconnection on connection loss.
     * Default: true
     */
    @Builder.Default
    private boolean autoReconnect = true;
    
    /**
     * Maximum number of reconnection attempts before giving up.
     * Default: 5 attempts
     * Set to -1 for infinite attempts
     */
    @Builder.Default
    private int maxReconnectAttempts = 5;
    
    /**
     * Initial delay before first reconnection attempt.
     * Default: 1 second
     */
    @Builder.Default
    private Duration reconnectInitialDelay = Duration.ofSeconds(1);
    
    /**
     * Maximum delay between reconnection attempts (exponential backoff cap).
     * Default: 30 seconds
     */
    @Builder.Default
    private Duration reconnectMaxDelay = Duration.ofSeconds(30);
    
    /**
     * Multiplier for exponential backoff between reconnection attempts.
     * Default: 2.0 (doubles delay each time)
     */
    @Builder.Default
    private double reconnectBackoffMultiplier = 2.0;
    
    /**
     * Enable message sequencing to ensure callbacks of the same type
     * are processed sequentially. When enabled:
     * - Multiple OFFER messages process one at a time
     * - Multiple FILL messages process one at a time
     * - But different types can process concurrently
     * 
     * Default: true (recommended for state safety)
     */
    @Builder.Default
    private boolean enableMessageSequencing = true;
    
    /**
     * Timeout for message sequencing queue.
     * Default: 30 seconds
     */
    @Builder.Default
    private Duration messageSequencingTimeout = Duration.ofSeconds(30);
    
    /**
     * Enable state mutation locking for thread-safe callback handling.
     * Default: true
     */
    @Builder.Default
    private boolean enableStateLocking = true;
    
    /**
     * Timeout for acquiring state mutation locks.
     * Default: 5 seconds
     */
    @Builder.Default
    private Duration stateLockTimeout = Duration.ofSeconds(5);
    
    /**
     * Enable automatic resync on reconnection.
     * Default: true
     */
    @Builder.Default
    private boolean autoResyncOnReconnect = true;
    
    /**
     * How far back to resync events after reconnection.
     * Default: 5 minutes
     */
    @Builder.Default
    private Duration resyncLookback = Duration.ofMinutes(5);
    
    /**
     * Creates a default configuration instance.
     */
    public static ConectorConfig defaultConfig() {
        return ConectorConfig.builder().build();
    }
    
    /**
     * Validates configuration values.
     * @throws IllegalArgumentException if configuration is invalid
     */
    public void validate() {
        if (heartbeatInterval.isNegative() || heartbeatInterval.isZero()) {
            throw new IllegalArgumentException("Heartbeat interval must be positive");
        }
        
        if (connectionTimeout.isNegative() || connectionTimeout.isZero()) {
            throw new IllegalArgumentException("Connection timeout must be positive");
        }
        
        if (maxReconnectAttempts < -1 || maxReconnectAttempts == 0) {
            throw new IllegalArgumentException("Max reconnect attempts must be -1 or positive");
        }
        
        if (reconnectInitialDelay.isNegative()) {
            throw new IllegalArgumentException("Reconnect initial delay cannot be negative");
        }
        
        if (reconnectMaxDelay.compareTo(reconnectInitialDelay) < 0) {
            throw new IllegalArgumentException("Reconnect max delay must be >= initial delay");
        }
        
        if (reconnectBackoffMultiplier < 1.0) {
            throw new IllegalArgumentException("Reconnect backoff multiplier must be >= 1.0");
        }
        
        if (stateLockTimeout.isNegative() || stateLockTimeout.isZero()) {
            throw new IllegalArgumentException("State lock timeout must be positive");
        }
        
        if (messageSequencingTimeout.isNegative() || messageSequencingTimeout.isZero()) {
            throw new IllegalArgumentException("Message sequencing timeout must be positive");
        }
        
        if (resyncLookback.isNegative()) {
            throw new IllegalArgumentException("Resync lookback cannot be negative");
        }
        
        log.debug("Configuration validated successfully");
    }
}
```

---

## Phase 3: Exception Classes

### ConexionFallidaException

**File**: `src/main/java/tech/hellsoft/trading/exception/ConexionFallidaException.java`

```java
package tech.hellsoft.trading.exception;

import lombok.Getter;

@Getter
public class ConexionFallidaException extends Exception {
    private final String host;
    private final int port;
    
    public ConexionFallidaException(String message, String host, int port) {
        super(message);
        this.host = host;
        this.port = port;
    }
    
    public ConexionFallidaException(String message, String host, int port, Throwable cause) {
        super(message, cause);
        this.host = host;
        this.port = port;
    }
}
```

### ValidationException

**File**: `src/main/java/tech/hellsoft/trading/exception/ValidationException.java`

```java
package tech.hellsoft.trading.exception;

public class ValidationException extends RuntimeException {
    public ValidationException(String message) {
        super(message);
    }
    
    public ValidationException(String message, Throwable cause) {
        super(message, cause);
    }
}
```

### StateLockException

**File**: `src/main/java/tech/hellsoft/trading/exception/StateLockException.java`

```java
package tech.hellsoft.trading.exception;

import lombok.Getter;

/**
 * Thrown when a state mutation lock cannot be acquired within the timeout period.
 */
@Getter
public class StateLockException extends RuntimeException {
    private final String actionName;
    private final long timeoutMillis;
    
    public StateLockException(String actionName, long timeoutMillis) {
        super(String.format("Could not acquire lock for action '%s' within %d ms", 
            actionName, timeoutMillis));
        this.actionName = actionName;
        this.timeoutMillis = timeoutMillis;
    }
}
```

---

## Phase 4: Data Transfer Objects

### Nested DTOs

#### Recipe

**File**: `src/main/java/tech/hellsoft/trading/dto/server/Recipe.java`

```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.Product;
import tech.hellsoft.trading.enums.RecipeType;

import java.util.Map;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class Recipe {
    private RecipeType type;
    private Map<Product, Integer> ingredients;
    private Double premiumBonus;
}
```

#### TeamRole

**File**: `src/main/java/tech/hellsoft/trading/dto/server/TeamRole.java`

```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class TeamRole {
    private Integer branches;
    private Integer maxDepth;
    private Double decay;
    private Double budget;
    private Double baseEnergy;
    private Double levelEnergy;
}
```

### Client DTOs

#### LoginMessage

**File**: `src/main/java/tech/hellsoft/trading/dto/client/LoginMessage.java`

```java
package tech.hellsoft.trading.dto.client;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class LoginMessage {
    private MessageType type;
    private String token;
    private String tz;
}
```

#### OrderMessage

**File**: `src/main/java/tech/hellsoft/trading/dto/client/OrderMessage.java`

```java
package tech.hellsoft.trading.dto.client;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.OrderMode;
import tech.hellsoft.trading.enums.OrderSide;
import tech.hellsoft.trading.enums.Product;

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
    private String debugMode;
}
```

#### ProductionUpdateMessage

**File**: `src/main/java/tech/hellsoft/trading/dto/client/ProductionUpdateMessage.java`

```java
package tech.hellsoft.trading.dto.client;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.Product;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ProductionUpdateMessage {
    private MessageType type;
    private Product product;
    private Integer quantity;
}
```

#### AcceptOfferMessage

**File**: `src/main/java/tech/hellsoft/trading/dto/client/AcceptOfferMessage.java`

```java
package tech.hellsoft.trading.dto.client;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class AcceptOfferMessage {
    private MessageType type;
    private String offerId;
    private Boolean accept;
    private Integer quantityOffered;
    private Double priceOffered;
}
```

#### ResyncMessage

**File**: `src/main/java/tech/hellsoft/trading/dto/client/ResyncMessage.java`

```java
package tech.hellsoft.trading.dto.client;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ResyncMessage {
    private MessageType type;
    private String lastSync;
}
```

#### CancelMessage

**File**: `src/main/java/tech/hellsoft/trading/dto/client/CancelMessage.java`

```java
package tech.hellsoft.trading.dto.client;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CancelMessage {
    private MessageType type;
    private String clOrdID;
}
```

#### PingMessage

**File**: `src/main/java/tech/hellsoft/trading/dto/client/PingMessage.java`

```java
package tech.hellsoft.trading.dto.client;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class PingMessage {
    private MessageType type;
    private String timestamp;
}
```

### Server DTOs

#### LoginOKMessage

**File**: `src/main/java/tech/hellsoft/trading/dto/server/LoginOKMessage.java`

```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.Product;

import java.util.Collections;
import java.util.List;
import java.util.Map;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class LoginOKMessage {
    private MessageType type;
    private String team;
    private String species;
    private Double initialBalance;
    private Double currentBalance;
    private Map<Product, Integer> inventory;
    private List<Product> authorizedProducts;
    private Map<Product, Recipe> recipes;
    private TeamRole role;
    private String serverTime;
    
    public Map<Product, Integer> getInventory() {
        return inventory == null ? Map.of() : Collections.unmodifiableMap(inventory);
    }
    
    public List<Product> getAuthorizedProducts() {
        return authorizedProducts == null ? List.of() : Collections.unmodifiableList(authorizedProducts);
    }
    
    public Map<Product, Recipe> getRecipes() {
        return recipes == null ? Map.of() : Collections.unmodifiableMap(recipes);
    }
}
```

#### FillMessage

**File**: `src/main/java/tech/hellsoft/trading/dto/server/FillMessage.java`

```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.OrderSide;
import tech.hellsoft.trading.enums.Product;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class FillMessage {
    private MessageType type;
    private String clOrdID;
    private Integer fillQty;
    private Double fillPrice;
    private OrderSide side;
    private Product product;
    private String counterparty;
    private String counterpartyMessage;
    private String serverTime;
    private Integer remainingQty;
    private Integer totalQty;
}
```

#### TickerMessage

**File**: `src/main/java/tech/hellsoft/trading/dto/server/TickerMessage.java`

```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.Product;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class TickerMessage {
    private MessageType type;
    private Product product;
    private Double bestBid;
    private Double bestAsk;
    private Double mid;
    private Integer volume24h;
    private String serverTime;
}
```

#### OfferMessage

**File**: `src/main/java/tech/hellsoft/trading/dto/server/OfferMessage.java`

```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.Product;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class OfferMessage {
    private MessageType type;
    private String offerId;
    private String buyer;
    private Product product;
    private Integer quantityRequested;
    private Double maxPrice;
    private Integer expiresIn;
    private String timestamp;
}
```

#### ErrorMessage

**File**: `src/main/java/tech/hellsoft/trading/dto/server/ErrorMessage.java`

```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.ErrorCode;
import tech.hellsoft.trading.enums.MessageType;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ErrorMessage {
    private MessageType type;
    private ErrorCode code;
    private String reason;
    private String clOrdID;
    private String timestamp;
}
```

#### OrderAckMessage

**File**: `src/main/java/tech/hellsoft/trading/dto/server/OrderAckMessage.java`

```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.OrderStatus;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class OrderAckMessage {
    private MessageType type;
    private String clOrdID;
    private OrderStatus status;
    private String serverTime;
}
```

#### InventoryUpdateMessage

**File**: `src/main/java/tech/hellsoft/trading/dto/server/InventoryUpdateMessage.java`

```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.enums.Product;

import java.util.Map;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class InventoryUpdateMessage {
    private MessageType type;
    private Map<Product, Integer> inventory;
    private String serverTime;
}
```

#### BalanceUpdateMessage

**File**: `src/main/java/tech/hellsoft/trading/dto/server/BalanceUpdateMessage.java`

```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class BalanceUpdateMessage {
    private MessageType type;
    private Double balance;
    private String serverTime;
}
```

#### EventDeltaMessage

**File**: `src/main/java/tech/hellsoft/trading/dto/server/EventDeltaMessage.java`

```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;

import java.util.List;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class EventDeltaMessage {
    private MessageType type;
    private List<FillMessage> events;
    private String serverTime;
}
```

#### BroadcastNotificationMessage

**File**: `src/main/java/tech/hellsoft/trading/dto/server/BroadcastNotificationMessage.java`

```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class BroadcastNotificationMessage {
    private MessageType type;
    private String message;
    private String sender;
    private String serverTime;
}
```

#### PongMessage

**File**: `src/main/java/tech/hellsoft/trading/dto/server/PongMessage.java`

```java
package tech.hellsoft.trading.dto.server;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import tech.hellsoft.trading.enums.MessageType;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class PongMessage {
    private MessageType type;
    private String timestamp;
}
```

---

## Phase 5: Internal Infrastructure - State Management

### StateLocker

**File**: `src/main/java/tech/hellsoft/trading/internal/routing/StateLocker.java`

```java
package tech.hellsoft.trading.internal.routing;

import lombok.extern.slf4j.Slf4j;
import tech.hellsoft.trading.config.ConectorConfig;
import tech.hellsoft.trading.exception.StateLockException;

import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.locks.Lock;
import java.util.concurrent.locks.ReentrantLock;
import java.util.function.Consumer;

/**
 * Thread-safe state mutation locking system.
 */
@Slf4j
public class StateLocker {
    
    private final ConcurrentHashMap<String, Lock> locks = new ConcurrentHashMap<>();
    private final ConectorConfig config;
    
    public StateLocker(ConectorConfig config) {
        this.config = config;
        
        initializeLock("fill");
        initializeLock("inventory_update");
        initializeLock("balance_update");
        initializeLock("production");
        initializeLock("offer_accept");
    }
    
    public void executeWithLock(String actionName, Runnable action) {
        if (!config.isEnableStateLocking()) {
            action.run();
            return;
        }
        
        Lock lock = locks.computeIfAbsent(actionName, k -> new ReentrantLock());
        
        try {
            boolean acquired = lock.tryLock(
                config.getStateLockTimeout().toMillis(),
                TimeUnit.MILLISECONDS
            );
            
            if (!acquired) {
                throw new StateLockException(actionName, config.getStateLockTimeout().toMillis());
            }
            
            try {
                log.trace("Acquired lock for action: {}", actionName);
                action.run();
            } finally {
                lock.unlock();
                log.trace("Released lock for action: {}", actionName);
            }
            
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            throw new RuntimeException("Interrupted while waiting for lock: " + actionName, e);
        }
    }
    
    public <T> void executeWithLock(String actionName, T data, Consumer<T> action) {
        executeWithLock(actionName, () -> action.accept(data));
    }
    
    private void initializeLock(String actionName) {
        locks.putIfAbsent(actionName, new ReentrantLock());
        log.debug("Initialized lock for action: {}", actionName);
    }
    
    public void clearLocks() {
        locks.clear();
        log.debug("Cleared all locks");
    }
}
```

### MessageSequencer

**File**: `src/main/java/tech/hellsoft/trading/internal/routing/MessageSequencer.java`

```java
package tech.hellsoft.trading.internal.routing;

import lombok.extern.slf4j.Slf4j;
import tech.hellsoft.trading.config.ConectorConfig;
import tech.hellsoft.trading.enums.MessageType;

import java.util.concurrent.*;

/**
 * Ensures sequential processing of messages by type.
 * 
 * <p>Each MessageType gets its own single-threaded executor, ensuring that:
 * <ul>
 *   <li>Multiple OFFER messages are processed one at a time</li>
 *   <li>Multiple FILL messages are processed one at a time</li>
 *   <li>Multiple TICKER messages are processed one at a time</li>
 *   <li>BUT different message types can process concurrently</li>
 * </ul>
 */
@Slf4j
public class MessageSequencer {
    
    private final ConcurrentHashMap<MessageType, ExecutorService> executorsByType;
    private final ConectorConfig config;
    private final boolean enabled;
    
    public MessageSequencer(ConectorConfig config) {
        this.config = config;
        this.enabled = config.isEnableMessageSequencing();
        this.executorsByType = new ConcurrentHashMap<>();
        
        if (enabled) {
            initializeExecutor(MessageType.FILL);
            initializeExecutor(MessageType.TICKER);
            initializeExecutor(MessageType.OFFER);
            initializeExecutor(MessageType.ERROR);
            initializeExecutor(MessageType.LOGIN_OK);
            initializeExecutor(MessageType.BROADCAST_NOTIFICATION);
            initializeExecutor(MessageType.EVENT_DELTA);
            
            log.info("Message sequencing enabled - callbacks will be processed sequentially by type");
        } else {
            log.info("Message sequencing disabled - callbacks may process concurrently");
        }
    }
    
    public void execute(MessageType messageType, Runnable callback) {
        if (!enabled) {
            callback.run();
            return;
        }
        
        ExecutorService executor = executorsByType.computeIfAbsent(
            messageType,
            this::createExecutorForType
        );
        
        executor.execute(() -> {
            try {
                log.trace("Processing {} message sequentially", messageType);
                callback.run();
            } catch (Exception e) {
                log.error("Error processing {} callback", messageType, e);
            }
        });
    }
    
    public boolean execute(MessageType messageType, Runnable callback, 
                          long timeout, TimeUnit unit) {
        if (!enabled) {
            callback.run();
            return true;
        }
        
        ExecutorService executor = executorsByType.computeIfAbsent(
            messageType,
            this::createExecutorForType
        );
        
        CompletableFuture<Void> future = CompletableFuture.runAsync(() -> {
            try {
                log.trace("Processing {} message sequentially", messageType);
                callback.run();
            } catch (Exception e) {
                log.error("Error processing {} callback", messageType, e);
                throw new CompletionException(e);
            }
        }, executor);
        
        try {
            future.get(timeout, unit);
            return true;
        } catch (TimeoutException e) {
            log.warn("Timeout processing {} message after {} {}", 
                messageType, timeout, unit);
            return false;
        } catch (Exception e) {
            log.error("Error processing {} message", messageType, e);
            return false;
        }
    }
    
    public int getQueueSize(MessageType messageType) {
        if (!enabled) {
            return 0;
        }
        
        ExecutorService executor = executorsByType.get(messageType);
        if (executor instanceof ThreadPoolExecutor tpe) {
            return tpe.getQueue().size();
        }
        return 0;
    }
    
    public void shutdown() {
        if (!enabled) {
            return;
        }
        
        log.info("Shutting down message sequencers...");
        
        executorsByType.values().forEach(executor -> {
            executor.shutdown();
            try {
                if (!executor.awaitTermination(2, TimeUnit.SECONDS)) {
                    executor.shutdownNow();
                }
            } catch (InterruptedException e) {
                executor.shutdownNow();
                Thread.currentThread().interrupt();
            }
        });
        
        executorsByType.clear();
        log.info("Message sequencers shut down");
    }
    
    private ExecutorService createExecutorForType(MessageType type) {
        log.debug("Creating sequential executor for message type: {}", type);
        
        return Executors.newSingleThreadExecutor(r -> {
            Thread t = Thread.ofVirtual().unstarted(r);
            t.setName("msg-seq-" + type.getValue().toLowerCase());
            t.setDaemon(true);
            return t;
        });
    }
    
    private void initializeExecutor(MessageType type) {
        executorsByType.putIfAbsent(type, createExecutorForType(type));
    }
}
```

---

## Phase 6: Internal Infrastructure - Message Processing

### JsonSerializer

**File**: `src/main/java/tech/hellsoft/trading/internal/serialization/JsonSerializer.java`

```java
package tech.hellsoft.trading.internal.serialization;

import com.google.gson.*;
import com.google.gson.stream.JsonReader;
import com.google.gson.stream.JsonToken;
import com.google.gson.stream.JsonWriter;
import lombok.extern.slf4j.Slf4j;

import java.io.IOException;
import java.lang.reflect.Type;
import java.time.Instant;

@Slf4j
public class JsonSerializer {
    
    private static final Gson GSON = new GsonBuilder()
        .registerTypeAdapter(Instant.class, new InstantAdapter())
        .setPrettyPrinting()
        .create();
    
    public static String toJson(Object obj) {
        if (obj == null) {
            throw new IllegalArgumentException("Cannot serialize null object");
        }
        return GSON.toJson(obj);
    }
    
    public static <T> T fromJson(String json, Class<T> clazz) {
        if (json == null || json.isBlank()) {
            throw new IllegalArgumentException("Cannot deserialize null or empty JSON");
        }
        
        if (clazz == null) {
            throw new IllegalArgumentException("Target class cannot be null");
        }
        
        return GSON.fromJson(json, clazz);
    }
    
    public static <T> T fromJson(String json, Type type) {
        if (json == null || json.isBlank()) {
            throw new IllegalArgumentException("Cannot deserialize null or empty JSON");
        }
        
        if (type == null) {
            throw new IllegalArgumentException("Target type cannot be null");
        }
        
        return GSON.fromJson(json, type);
    }
    
    private static class InstantAdapter extends TypeAdapter<Instant> {
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
            
            String timestamp = in.nextString();
            return Instant.parse(timestamp);
        }
    }
}
```

### MessageRouter

**File**: `src/main/java/tech/hellsoft/trading/internal/routing/MessageRouter.java`

```java
package tech.hellsoft.trading.internal.routing;

import com.google.gson.JsonObject;
import com.google.gson.JsonParser;
import lombok.extern.slf4j.Slf4j;
import tech.hellsoft.trading.EventListener;
import tech.hellsoft.trading.dto.server.*;
import tech.hellsoft.trading.enums.MessageType;
import tech.hellsoft.trading.internal.serialization.JsonSerializer;

import java.util.List;
import java.util.function.Consumer;

@Slf4j
public class MessageRouter {
    
    private final List<EventListener> listeners;
    private final StateLocker stateLocker;
    private final MessageSequencer messageSequencer;
    private final Runnable onLoginOkCallback;
    
    public MessageRouter(List<EventListener> listeners, 
                        StateLocker stateLocker,
                        MessageSequencer messageSequencer,
                        Runnable onLoginOkCallback) {
        this.listeners = listeners;
        this.stateLocker = stateLocker;
        this.messageSequencer = messageSequencer;
        this.onLoginOkCallback = onLoginOkCallback;
    }
    
    public void processMessage(String json) {
        if (json == null || json.isBlank()) {
            log.warn("Received null or empty message");
            return;
        }
        
        try {
            JsonObject obj = JsonParser.parseString(json).getAsJsonObject();
            
            if (!obj.has("type")) {
                log.error("Message missing 'type' field: {}", json);
                return;
            }
            
            String typeStr = obj.get("type").getAsString();
            MessageType type = MessageType.fromJson(typeStr);
            
            routeMessage(type, json);
            
        } catch (Exception e) {
            log.error("Failed to process message: {}", json, e);
        }
    }
    
    private void routeMessage(MessageType type, String json) {
        switch (type) {
            case LOGIN_OK -> {
                LoginOKMessage msg = JsonSerializer.fromJson(json, LoginOKMessage.class);
                messageSequencer.execute(MessageType.LOGIN_OK, () -> {
                    notifyListeners(listener -> listener.onLoginOk(msg));
                    if (onLoginOkCallback != null) {
                        onLoginOkCallback.run();
                    }
                });
            }
            
            case FILL -> {
                FillMessage msg = JsonSerializer.fromJson(json, FillMessage.class);
                messageSequencer.execute(MessageType.FILL, () -> {
                    stateLocker.executeWithLock("fill", msg, fill -> 
                        notifyListeners(listener -> listener.onFill(fill))
                    );
                });
            }
            
            case TICKER -> {
                TickerMessage msg = JsonSerializer.fromJson(json, TickerMessage.class);
                messageSequencer.execute(MessageType.TICKER, () -> {
                    notifyListeners(listener -> listener.onTicker(msg));
                });
            }
            
            case OFFER -> {
                OfferMessage msg = JsonSerializer.fromJson(json, OfferMessage.class);
                messageSequencer.execute(MessageType.OFFER, () -> {
                    notifyListeners(listener -> listener.onOffer(msg));
                });
            }
            
            case ERROR -> {
                ErrorMessage msg = JsonSerializer.fromJson(json, ErrorMessage.class);
                messageSequencer.execute(MessageType.ERROR, () -> {
                    notifyListeners(listener -> listener.onError(msg));
                });
            }
            
            case ORDER_ACK -> {
                OrderAckMessage msg = JsonSerializer.fromJson(json, OrderAckMessage.class);
                messageSequencer.execute(MessageType.ORDER_ACK, () -> {
                    log.debug("Order acknowledged: {}", msg.getClOrdID());
                });
            }
            
            case INVENTORY_UPDATE -> {
                InventoryUpdateMessage msg = JsonSerializer.fromJson(json, InventoryUpdateMessage.class);
                messageSequencer.execute(MessageType.INVENTORY_UPDATE, () -> {
                    stateLocker.executeWithLock("inventory_update", msg, update -> 
                        log.debug("Inventory updated: {}", update.getInventory())
                    );
                });
            }
            
            case BALANCE_UPDATE -> {
                BalanceUpdateMessage msg = JsonSerializer.fromJson(json, BalanceUpdateMessage.class);
                messageSequencer.execute(MessageType.BALANCE_UPDATE, () -> {
                    stateLocker.executeWithLock("balance_update", msg, update -> 
                        log.debug("Balance updated: {}", update.getBalance())
                    );
                });
            }
            
            case EVENT_DELTA -> {
                EventDeltaMessage msg = JsonSerializer.fromJson(json, EventDeltaMessage.class);
                messageSequencer.execute(MessageType.EVENT_DELTA, () -> {
                    msg.getEvents().forEach(fill -> 
                        stateLocker.executeWithLock("fill", fill, f -> 
                            notifyListeners(listener -> listener.onFill(f))
                        )
                    );
                });
            }
            
            case BROADCAST_NOTIFICATION -> {
                BroadcastNotificationMessage msg = JsonSerializer.fromJson(json, BroadcastNotificationMessage.class);
                messageSequencer.execute(MessageType.BROADCAST_NOTIFICATION, () -> {
                    notifyListeners(listener -> listener.onBroadcast(msg));
                });
            }
            
            case PONG -> {
                PongMessage msg = JsonSerializer.fromJson(json, PongMessage.class);
                log.trace("Received PONG: {}", msg.getTimestamp());
            }
            
            default -> log.warn("Unknown or unhandled message type: {}", type);
        }
    }
    
    private void notifyListeners(Consumer<EventListener> action) {
        listeners.forEach(listener -> {
            try {
                action.accept(listener);
            } catch (Exception e) {
                log.error("Listener callback failed", e);
            }
        });
    }
}
```

---

## Phase 7: Connection Management

### HeartbeatManager

**File**: `src/main/java/tech/hellsoft/trading/internal/connection/HeartbeatManager.java`

```java
package tech.hellsoft.trading.internal.connection;

import lombok.extern.slf4j.Slf4j;
import tech.hellsoft.trading.config.ConectorConfig;
import tech.hellsoft.trading.dto.client.PingMessage;
import tech.hellsoft.trading.enums.MessageType;

import java.time.Instant;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.ScheduledFuture;
import java.util.concurrent.TimeUnit;
import java.util.function.Consumer;

@Slf4j
public class HeartbeatManager {
    
    private final ConectorConfig config;
    private final Consumer<PingMessage> sendPingCallback;
    private final ScheduledExecutorService scheduler;
    
    private ScheduledFuture<?> heartbeatTask;
    private volatile boolean running = false;
    
    public HeartbeatManager(ConectorConfig config, Consumer<PingMessage> sendPingCallback) {
        this.config = config;
        this.sendPingCallback = sendPingCallback;
        this.scheduler = Executors.newSingleThreadScheduledExecutor(r -> {
            Thread t = Thread.ofVirtual().unstarted(r);
            t.setName("heartbeat-manager");
            t.setDaemon(true);
            return t;
        });
    }
    
    public void start() {
        if (running) {
            log.warn("Heartbeat already running");
            return;
        }
        
        long intervalMillis = config.getHeartbeatInterval().toMillis();
        
        heartbeatTask = scheduler.scheduleAtFixedRate(
            this::sendPing,
            intervalMillis,
            intervalMillis,
            TimeUnit.MILLISECONDS
        );
        
        running = true;
        log.info("Heartbeat started (interval: {} ms)", intervalMillis);
    }
    
    public void stop() {
        if (!running) {
            return;
        }
        
        if (heartbeatTask != null) {
            heartbeatTask.cancel(false);
            heartbeatTask = null;
        }
        
        running = false;
        log.info("Heartbeat stopped");
    }
    
    public void shutdown() {
        stop();
        scheduler.shutdown();
        try {
            if (!scheduler.awaitTermination(1, TimeUnit.SECONDS)) {
                scheduler.shutdownNow();
            }
        } catch (InterruptedException e) {
            scheduler.shutdownNow();
            Thread.currentThread().interrupt();
        }
        log.info("Heartbeat manager shut down");
    }
    
    public boolean isRunning() {
        return running;
    }
    
    private void sendPing() {
        try {
            PingMessage ping = PingMessage.builder()
                .type(MessageType.PING)
                .timestamp(Instant.now().toString())
                .build();
            
            sendPingCallback.accept(ping);
            log.trace("PING sent");
            
        } catch (Exception e) {
            log.error("Failed to send PING", e);
        }
    }
}
```

### ConnectionManager

**File**: `src/main/java/tech/hellsoft/trading/internal/connection/ConnectionManager.java`

```java
package tech.hellsoft.trading.internal.connection;

import lombok.extern.slf4j.Slf4j;
import tech.hellsoft.trading.config.ConectorConfig;
import tech.hellsoft.trading.enums.ConnectionState;
import tech.hellsoft.trading.exception.ConexionFallidaException;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.WebSocket;
import java.time.Instant;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.function.Consumer;

@Slf4j
public class ConnectionManager {
    
    private final ConectorConfig config;
    private final WebSocket.Listener websocketListener;
    private final Consumer<ConnectionState> stateChangeCallback;
    private final Runnable onReconnectedCallback;
    
    private volatile WebSocket webSocket;
    private volatile ConnectionState state = ConnectionState.DISCONNECTED;
    private volatile String host;
    private volatile int port;
    
    private final AtomicInteger reconnectAttempts = new AtomicInteger(0);
    private final ScheduledExecutorService reconnectScheduler;
    
    private Instant lastSyncTime = Instant.now();
    
    public ConnectionManager(ConectorConfig config,
                            WebSocket.Listener websocketListener,
                            Consumer<ConnectionState> stateChangeCallback,
                            Runnable onReconnectedCallback) {
        this.config = config;
        this.websocketListener = websocketListener;
        this.stateChangeCallback = stateChangeCallback;
        this.onReconnectedCallback = onReconnectedCallback;
        
        this.reconnectScheduler = Executors.newSingleThreadScheduledExecutor(r -> {
            Thread t = Thread.ofVirtual().unstarted(r);
            t.setName("reconnect-scheduler");
            t.setDaemon(true);
            return t;
        });
    }
    
    public void connect(String host, int port) throws ConexionFallidaException {
        if (state == ConnectionState.CONNECTED || state == ConnectionState.AUTHENTICATED) {
            log.warn("Already connected");
            return;
        }
        
        this.host = host;
        this.port = port;
        
        setState(ConnectionState.CONNECTING);
        
        try {
            URI uri = URI.create(String.format("ws://%s:%d/ws", host, port));
            log.info("Connecting to {}", uri);
            
            HttpClient client = HttpClient.newHttpClient();
            
            CompletableFuture<WebSocket> wsFuture = client.newWebSocketBuilder()
                .buildAsync(uri, websocketListener);
            
            webSocket = wsFuture.get(
                config.getConnectionTimeout().toMillis(),
                TimeUnit.MILLISECONDS
            );
            
            setState(ConnectionState.CONNECTED);
            reconnectAttempts.set(0);
            
            log.info("Connected successfully to {}:{}", host, port);
            
        } catch (Exception e) {
            setState(ConnectionState.DISCONNECTED);
            throw new ConexionFallidaException("Connection failed: " + e.getMessage(), host, port, e);
        }
    }
    
    public void setAuthenticated() {
        if (state == ConnectionState.CONNECTED) {
            setState(ConnectionState.AUTHENTICATED);
            lastSyncTime = Instant.now();
            log.info("Connection authenticated");
        }
    }
    
    public void handleConnectionLoss(Throwable error) {
        log.warn("Connection lost: {}", error.getMessage());
        
        setState(ConnectionState.DISCONNECTED);
        webSocket = null;
        
        if (config.isAutoReconnect() && shouldAttemptReconnect()) {
            scheduleReconnect();
        } else {
            setState(ConnectionState.CLOSED);
            log.info("Auto-reconnect disabled or max attempts reached");
        }
    }
    
    public void disconnect() {
        if (webSocket != null && !webSocket.isOutputClosed()) {
            log.info("Disconnecting gracefully");
            webSocket.sendClose(WebSocket.NORMAL_CLOSURE, "Client disconnect");
            webSocket = null;
        }
        
        setState(ConnectionState.CLOSED);
    }
    
    public void shutdown() {
        disconnect();
        
        reconnectScheduler.shutdown();
        try {
            if (!reconnectScheduler.awaitTermination(1, TimeUnit.SECONDS)) {
                reconnectScheduler.shutdownNow();
            }
        } catch (InterruptedException e) {
            reconnectScheduler.shutdownNow();
            Thread.currentThread().interrupt();
        }
        
        log.info("Connection manager shut down");
    }
    
    public ConnectionState getState() {
        return state;
    }
    
    public boolean isConnected() {
        return state == ConnectionState.CONNECTED || state == ConnectionState.AUTHENTICATED;
    }
    
    public boolean isAuthenticated() {
        return state == ConnectionState.AUTHENTICATED;
    }
    
    public WebSocket getWebSocket() {
        return webSocket;
    }
    
    public Instant getLastSyncTime() {
        return lastSyncTime;
    }
    
    public void updateLastSyncTime() {
        this.lastSyncTime = Instant.now();
    }
    
    private void setState(ConnectionState newState) {
        if (this.state != newState) {
            log.debug("Connection state: {} -> {}", this.state, newState);
            this.state = newState;
            
            if (stateChangeCallback != null) {
                stateChangeCallback.accept(newState);
            }
        }
    }
    
    private boolean shouldAttemptReconnect() {
        int maxAttempts = config.getMaxReconnectAttempts();
        int currentAttempts = reconnectAttempts.get();
        
        return maxAttempts == -1 || currentAttempts < maxAttempts;
    }
    
    private void scheduleReconnect() {
        int attempt = reconnectAttempts.incrementAndGet();
        
        long delayMillis = calculateReconnectDelay(attempt);
        
        log.info("Scheduling reconnect attempt {} in {} ms", attempt, delayMillis);
        setState(ConnectionState.RECONNECTING);
        
        reconnectScheduler.schedule(() -> {
            try {
                log.info("Reconnect attempt {}", attempt);
                connect(host, port);
                
                if (onReconnectedCallback != null) {
                    onReconnectedCallback.run();
                }
                
            } catch (ConexionFallidaException e) {
                log.error("Reconnect attempt {} failed", attempt, e);
                handleConnectionLoss(e);
            }
        }, delayMillis, TimeUnit.MILLISECONDS);
    }
    
    private long calculateReconnectDelay(int attempt) {
        long initialDelay = config.getReconnectInitialDelay().toMillis();
        long maxDelay = config.getReconnectMaxDelay().toMillis();
        double multiplier = config.getReconnectBackoffMultiplier();
        
        double delay = initialDelay * Math.pow(multiplier, attempt - 1);
        
        return Math.min((long) delay, maxDelay);
    }
}
```

### WebSocketHandler

**File**: `src/main/java/tech/hellsoft/trading/internal/connection/WebSocketHandler.java`

```java
package tech.hellsoft.trading.internal.connection;

import lombok.extern.slf4j.Slf4j;
import tech.hellsoft.trading.internal.routing.MessageRouter;

import java.net.http.WebSocket;
import java.util.concurrent.CompletionStage;
import java.util.concurrent.ExecutorService;
import java.util.function.Consumer;

@Slf4j
public class WebSocketHandler implements WebSocket.Listener {
    
    private final StringBuilder messageBuffer = new StringBuilder();
    private final MessageRouter router;
    private final ExecutorService callbackExecutor;
    private final Consumer<Throwable> onErrorCallback;
    
    public WebSocketHandler(MessageRouter router,
                           ExecutorService callbackExecutor,
                           Consumer<Throwable> onErrorCallback) {
        this.router = router;
        this.callbackExecutor = callbackExecutor;
        this.onErrorCallback = onErrorCallback;
    }
    
    @Override
    public CompletionStage<?> onText(WebSocket webSocket, CharSequence data, boolean last) {
        messageBuffer.append(data);
        
        if (last) {
            String json = messageBuffer.toString();
            messageBuffer.setLength(0);
            
            callbackExecutor.execute(() -> router.processMessage(json));
        }
        
        webSocket.request(1);
        return null;
    }
    
    @Override
    public void onError(WebSocket webSocket, Throwable error) {
        log.error("WebSocket error", error);
        if (onErrorCallback != null) {
            onErrorCallback.accept(error);
        }
    }
    
    @Override
    public CompletionStage<?> onClose(WebSocket webSocket, int statusCode, String reason) {
        log.info("WebSocket closed: {} - {}", statusCode, reason);
        if (onErrorCallback != null) {
            onErrorCallback.accept(new RuntimeException("Connection closed: " + reason));
        }
        return null;
    }
    
    @Override
    public void onOpen(WebSocket webSocket) {
        log.debug("WebSocket opened");
        webSocket.request(1);
    }
}
```

---

## Phase 8: Public API

### EventListener Interface

**File**: `src/main/java/tech/hellsoft/trading/EventListener.java`

```java
package tech.hellsoft.trading;

import tech.hellsoft.trading.dto.server.*;

/**
 * Callback interface for receiving server events.
 * 
 * <p><b>Thread Safety:</b> All callbacks are invoked on SDK virtual threads.
 * State-mutating callbacks (onFill, etc.) are protected by automatic locking
 * and sequential processing to prevent race conditions.
 */
public interface EventListener {
    
    /**
     * Called when authentication succeeds.
     */
    void onLoginOk(LoginOKMessage message);
    
    /**
     * Called when an order is executed (trade completed).
     * 
     * <p><b>Thread Safety:</b> This callback is automatically protected by
     * state locking and sequential processing.
     */
    void onFill(FillMessage fill);
    
    /**
     * Called every 5 seconds with market price updates.
     */
    void onTicker(TickerMessage ticker);
    
    /**
     * Called when another team makes a direct purchase offer.
     */
    void onOffer(OfferMessage offer);
    
    /**
     * Called when server rejects an operation or validation fails.
     */
    void onError(ErrorMessage error);
    
    /**
     * Called when WebSocket connection is lost.
     * 
     * <p><b>Note:</b> If auto-reconnect is enabled, SDK will automatically
     * attempt to reconnect.
     */
    void onConnectionLost(Exception exception);
    
    /**
     * Called when server admin sends a broadcast announcement.
     */
    void onBroadcast(BroadcastNotificationMessage broadcast);
}
```

### ConectorBolsa Main Class

**File**: `src/main/java/tech/hellsoft/trading/ConectorBolsa.java`

```java
package tech.hellsoft.trading;

import lombok.extern.slf4j.Slf4j;
import tech.hellsoft.trading.config.ConectorConfig;
import tech.hellsoft.trading.dto.client.*;
import tech.hellsoft.trading.enums.*;
import tech.hellsoft.trading.exception.*;
import tech.hellsoft.trading.internal.connection.*;
import tech.hellsoft.trading.internal.routing.*;
import tech.hellsoft.trading.internal.serialization.JsonSerializer;

import java.time.Instant;
import java.util.List;
import java.util.concurrent.*;

/**
 * Main SDK class for WebSocket-based trading client.
 * 
 * <p><b>Basic Usage:</b>
 * <pre>{@code
 * ConectorBolsa connector = new ConectorBolsa();
 * connector.conectar("localhost", 9000);
 * connector.login("TK-TEAM-2025", myEventListener);
 * connector.enviarOrden(order);
 * connector.desconectar();
 * }</pre>
 */
@Slf4j
public class ConectorBolsa {
    
    private final ConectorConfig config;
    private final ExecutorService callbackExecutor;
    private final Semaphore sendSemaphore = new Semaphore(1);
    private final List<EventListener> listeners = new CopyOnWriteArrayList<>();
    
    private final StateLocker stateLocker;
    private final MessageSequencer messageSequencer;
    private ConnectionManager connectionManager;
    private HeartbeatManager heartbeatManager;
    private MessageRouter messageRouter;
    private WebSocketHandler websocketHandler;
    
    private volatile String currentApiKey;
    
    public ConectorBolsa() {
        this(ConectorConfig.defaultConfig());
    }
    
    public ConectorBolsa(ConectorConfig config) {
        this.config = config;
        config.validate();
        
        this.callbackExecutor = Executors.newVirtualThreadPerTaskExecutor();
        this.stateLocker = new StateLocker(config);
        this.messageSequencer = new MessageSequencer(config);
        
        log.info("ConectorBolsa initialized");
    }
    
    public void conectar(String host, int port) throws ConexionFallidaException {
        if (host == null || host.isBlank()) {
            throw new ConexionFallidaException("Host cannot be null or empty", host, port);
        }
        
        if (port <= 0 || port > 65535) {
            throw new ConexionFallidaException("Invalid port number", host, port);
        }
        
        messageRouter = new MessageRouter(
            listeners,
            stateLocker,
            messageSequencer,
            this::onLoginOkInternal
        );
        
        websocketHandler = new WebSocketHandler(
            messageRouter,
            callbackExecutor,
            this::handleConnectionError
        );
        
        connectionManager = new ConnectionManager(
            config,
            websocketHandler,
            this::handleStateChange,
            this::handleReconnection
        );
        
        connectionManager.connect(host, port);
        
        heartbeatManager = new HeartbeatManager(config, this::sendPingInternal);
        heartbeatManager.start();
    }
    
    public void login(String apiKey, EventListener listener) {
        if (!connectionManager.isConnected()) {
            throw new IllegalStateException("Not connected to server. Call conectar() first.");
        }
        
        validateApiKey(apiKey);
        
        if (listener == null) {
            throw new IllegalArgumentException("Listener cannot be null");
        }
        
        this.currentApiKey = apiKey;
        listeners.add(listener);
        
        LoginMessage login = LoginMessage.builder()
            .type(MessageType.LOGIN)
            .token(apiKey)
            .tz(java.util.TimeZone.getDefault().getID())
            .build();
        
        sendMessage(login);
        log.info("Login message sent");
    }
    
    public void enviarOrden(OrderMessage orden) {
        if (!connectionManager.isAuthenticated()) {
            throw new IllegalStateException("Must be authenticated to send orders");
        }
        
        if (orden == null) {
            throw new IllegalArgumentException("Order cannot be null");
        }
        
        validateOrder(orden);
        
        if (orden.getType() == null) {
            orden.setType(MessageType.ORDER);
        }
        
        sendMessage(orden);
        log.info("Order sent: {}", orden.getClOrdID());
    }
    
    public void enviarProduccion(Product product, int cantidad) {
        if (!connectionManager.isAuthenticated()) {
            throw new IllegalStateException("Must be authenticated to send production updates");
        }
        
        if (product == null) {
            throw new IllegalArgumentException("Product cannot be null");
        }
        
        if (cantidad <= 0) {
            throw new IllegalArgumentException("Quantity must be positive");
        }
        
        ProductionUpdateMessage production = ProductionUpdateMessage.builder()
            .type(MessageType.PRODUCTION_UPDATE)
            .product(product)
            .quantity(cantidad)
            .build();
        
        sendMessage(production);
        log.info("Production sent: {} x{}", product.getValue(), cantidad);
    }
    
    public void aceptarOferta(String offerId, int cantidad, double precio) {
        if (!connectionManager.isAuthenticated()) {
            throw new IllegalStateException("Must be authenticated to accept offers");
        }
        
        if (offerId == null || offerId.isBlank()) {
            throw new IllegalArgumentException("Offer ID cannot be null or empty");
        }
        
        if (cantidad <= 0) {
            throw new IllegalArgumentException("Quantity must be positive");
        }
        
        if (precio <= 0) {
            throw new IllegalArgumentException("Price must be positive");
        }
        
        AcceptOfferMessage accept = AcceptOfferMessage.builder()
            .type(MessageType.ACCEPT_OFFER)
            .offerId(offerId)
            .accept(true)
            .quantityOffered(cantidad)
            .priceOffered(precio)
            .build();
        
        sendMessage(accept);
        log.info("Offer accepted: {}", offerId);
    }
    
    public void rechazarOferta(String offerId) {
        if (!connectionManager.isAuthenticated()) {
            throw new IllegalStateException("Must be authenticated to reject offers");
        }
        
        if (offerId == null || offerId.isBlank()) {
            throw new IllegalArgumentException("Offer ID cannot be null or empty");
        }
        
        AcceptOfferMessage reject = AcceptOfferMessage.builder()
            .type(MessageType.ACCEPT_OFFER)
            .offerId(offerId)
            .accept(false)
            .build();
        
        sendMessage(reject);
        log.info("Offer rejected: {}", offerId);
    }
    
    public void resync(Instant ultimaSincronizacion) {
        if (!connectionManager.isAuthenticated()) {
            throw new IllegalStateException("Must be authenticated to resync");
        }
        
        if (ultimaSincronizacion == null) {
            throw new IllegalArgumentException("Last sync timestamp cannot be null");
        }
        
        ResyncMessage resync = ResyncMessage.builder()
            .type(MessageType.RESYNC)
            .lastSync(ultimaSincronizacion.toString())
            .build();
        
        sendMessage(resync);
        log.info("Resync requested from {}", ultimaSincronizacion);
    }
    
    public void cancelarOrden(String clOrdID) {
        if (!connectionManager.isAuthenticated()) {
            throw new IllegalStateException("Must be authenticated to cancel orders");
        }
        
        if (clOrdID == null || clOrdID.isBlank()) {
            throw new IllegalArgumentException("Order ID cannot be null or empty");
        }
        
        CancelMessage cancel = CancelMessage.builder()
            .type(MessageType.CANCEL)
            .clOrdID(clOrdID)
            .build();
        
        sendMessage(cancel);
        log.info("Order cancellation sent: {}", clOrdID);
    }
    
    public void desconectar() {
        log.info("Disconnecting...");
        
        if (heartbeatManager != null) {
            heartbeatManager.shutdown();
        }
        
        if (messageSequencer != null) {
            messageSequencer.shutdown();
        }
        
        if (connectionManager != null) {
            connectionManager.shutdown();
        }
        
        callbackExecutor.shutdown();
        try {
            if (!callbackExecutor.awaitTermination(2, TimeUnit.SECONDS)) {
                callbackExecutor.shutdownNow();
            }
        } catch (InterruptedException e) {
            callbackExecutor.shutdownNow();
            Thread.currentThread().interrupt();
        }
        
        log.info("Disconnected");
    }
    
    public ConnectionState getConnectionState() {
        return connectionManager != null ? connectionManager.getState() : ConnectionState.DISCONNECTED;
    }
    
    public boolean isAuthenticated() {
        return connectionManager != null && connectionManager.isAuthenticated();
    }
    
    public ConectorConfig getConfig() {
        return config;
    }
    
    // ===== INTERNAL METHODS =====
    
    private void sendMessage(Object message) {
        if (connectionManager == null || !connectionManager.isConnected()) {
            throw new IllegalStateException("Not connected to server");
        }
        
        try {
            sendSemaphore.acquire();
            try {
                String json = JsonSerializer.toJson(message);
                log.debug("Sending: {}", json);
                connectionManager.getWebSocket().sendText(json, true).join();
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
    
    private void sendPingInternal(PingMessage ping) {
        try {
            sendMessage(ping);
        } catch (Exception e) {
            log.error("Failed to send PING", e);
        }
    }
    
    private void onLoginOkInternal() {
        connectionManager.setAuthenticated();
        connectionManager.updateLastSyncTime();
    }
    
    private void handleConnectionError(Throwable error) {
        if (connectionManager != null) {
            connectionManager.handleConnectionLoss(error);
        }
        
        listeners.forEach(listener -> {
            try {
                listener.onConnectionLost(new Exception(error));
            } catch (Exception e) {
                log.error("Listener error in onConnectionLost", e);
            }
        });
    }
    
    private void handleStateChange(ConnectionState newState) {
        log.info("Connection state changed to: {}", newState);
        
        if (newState == ConnectionState.DISCONNECTED || newState == ConnectionState.CLOSED) {
            if (heartbeatManager != null && heartbeatManager.isRunning()) {
                heartbeatManager.stop();
            }
        }
        
        if (newState == ConnectionState.CONNECTED || newState == ConnectionState.AUTHENTICATED) {
            if (heartbeatManager != null && !heartbeatManager.isRunning()) {
                heartbeatManager.start();
            }
        }
    }
    
    private void handleReconnection() {
        if (currentApiKey != null) {
            LoginMessage login = LoginMessage.builder()
                .type(MessageType.LOGIN)
                .token(currentApiKey)
                .tz(java.util.TimeZone.getDefault().getID())
                .build();
            
            sendMessage(login);
            log.info("Re-authentication sent after reconnection");
            
            if (config.isAutoResyncOnReconnect()) {
                Instant resyncFrom = connectionManager.getLastSyncTime()
                    .minus(config.getResyncLookback());
                
                ResyncMessage resync = ResyncMessage.builder()
                    .type(MessageType.RESYNC)
                    .lastSync(resyncFrom.toString())
                    .build();
                
                sendMessage(resync);
                log.info("Auto-resync sent from {}", resyncFrom);
            }
        }
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
    
    private void validateApiKey(String apiKey) {
        if (apiKey == null || apiKey.isBlank()) {
            throw new IllegalArgumentException("API key cannot be null or empty");
        }
        
        if (!apiKey.startsWith("TK-")) {
            throw new IllegalArgumentException("API key must start with 'TK-'");
        }
    }
}
```

---

## Phase 9: Testing

### Test Structure

Create the following test files:

1. **EnumSerializationTest.java** - Test JSON serialization for all enums
2. **DtoSerializationTest.java** - Test DTO builders and JSON round-trips
3. **ValidationTest.java** - Test order and message validation
4. **StateLockTest.java** - Test state locking mechanism
5. **MessageSequencerTest.java** - Test sequential message processing
6. **HeartbeatTest.java** - Test heartbeat manager
7. **ConnectionManagerTest.java** - Test auto-reconnect logic
8. **IntegrationTest.java** - End-to-end integration tests

---

## Phase 10: Documentation

### README.md Example

```markdown
# Trading SDK - Bolsa Interestelar de Aguacates Andorianos

WebSocket-based Java SDK for trading in the Bolsa Interestelar.

## Features

✅ Automatic connection management with heartbeat
✅ Auto-reconnect with exponential backoff
✅ Thread-safe message processing with sequencing
✅ State mutation locking for callback safety
✅ Type-safe enumerations
✅ Flexible configuration system
✅ Clean, elegant API

## Quick Start

```java
// Create connector with default config
ConectorBolsa connector = new ConectorBolsa();

// Connect and login
connector.conectar("localhost", 9000);
connector.login("TK-TEAM-2025", new MyEventListener());

// Send orders
OrderMessage order = OrderMessage.builder()
    .type(MessageType.ORDER)
    .clOrdID(generateOrderID())
    .side(OrderSide.BUY)
    .mode(OrderMode.MARKET)
    .product(Product.FOSFO)
    .qty(10)
    .build();

connector.enviarOrden(order);

// Cleanup
connector.desconectar();
```

## Configuration

```java
ConectorConfig config = ConectorConfig.builder()
    .heartbeatInterval(Duration.ofSeconds(30))
    .autoReconnect(true)
    .enableMessageSequencing(true)
    .enableStateLocking(true)
    .build();

ConectorBolsa connector = new ConectorBolsa(config);
```

## License

See LICENSE file for details.
```

---

## Appendices

### Appendix A: Coding Standards

**Critical Rules:**
- ❌ NO else statements - use guard clauses
- ✅ Functional programming patterns (streams, lambdas, Optional)
- ✅ Virtual threads for all concurrency
- ✅ Lombok to minimize boilerplate
- ✅ Guard clauses for validation

### Appendix B: Message Sequencing

Messages of the same type are processed sequentially:
- Multiple OFFER messages: one at a time
- Multiple FILL messages: one at a time
- Different types: concurrent processing allowed

### Appendix C: State Locking

State-mutating operations (FILL, INVENTORY_UPDATE, BALANCE_UPDATE) are protected by automatic locking to prevent race conditions.

### Appendix D: Connection States

```
DISCONNECTED → CONNECTING → CONNECTED → AUTHENTICATED
↓             ↓
RECONNECTING ← ←  ← (on loss)
↓
CLOSED (permanent)
```

---

**End of Implementation Plan**