# Java SDK Specification
## Bolsa Interestelar de Aguacates Andorianos

**Version**: 1.0  
**Server Compatibility**: v2.5+  
**Last Updated**: January 15, 2025

---

## Table of Contents

1. [Overview](#overview)
2. [SDK Responsibilities](#sdk-responsibilities)
3. [Enumerations](#enumerations)
4. [Connection Protocol](#connection-protocol)
5. [Message Types Reference](#message-types-reference)
6. [Server Error Codes](#server-error-codes)
7. [Data Structures (DTOs)](#data-structures-dtos)
8. [SDK Interface](#sdk-interface)
9. [Threading Model](#threading-model)
10. [Validation Rules](#validation-rules)
11. [Exception Handling](#exception-handling)

---

## Overview

The Java SDK provides a **WebSocket-based client** for connecting to the Stock Market Server. It handles low-level communication, message serialization/deserialization, connection management, and threading, allowing client applications to focus on business logic.

### What the SDK Does

✅ **Manages WebSocket connection** (connect, disconnect, reconnect)  
✅ **Handles message serialization/deserialization** (JSON ↔ Java objects)  
✅ **Thread-safe message sending** (concurrent order submission)  
✅ **Asynchronous message reception** (callback-based event handling)  
✅ **Validates message format** before sending  
✅ **Provides structured error information** from server

### What the SDK Does NOT Do

❌ **Business logic** (production algorithms, P&L calculation)  
❌ **State management** (inventory tracking, balance tracking)  
❌ **Persistence** (snapshots, configuration files)  
❌ **User interface** (console commands, displays)  
❌ **Trading strategy** (when to buy/sell, pricing decisions)

---

## SDK Responsibilities

### 1. Connection Management

The SDK establishes and maintains a WebSocket connection to the server:

- **Initial connection**: `conectar(host, port)` establishes TCP connection
- **Authentication**: `login(apiKey, listener)` authenticates and receives initial state
- **Heartbeat**: Automatic ping/pong to keep connection alive
- **Disconnection handling**: Notifies client via `onConnectionLost()` callback
- **Clean shutdown**: `desconectar()` closes connection gracefully

### 2. Message Protocol

The SDK handles the JSON-based message protocol:

- **Serialization**: Converts Java objects (e.g., `OrderMessage`) to JSON
- **Deserialization**: Converts JSON strings to Java objects (e.g., `FillMessage`)
- **Type discrimination**: Routes messages based on `type` field
- **Validation**: Ensures required fields are present before sending

### 3. Threading

The SDK manages concurrent access to the WebSocket:

- **Send thread safety**: Uses semaphore to prevent simultaneous sends
- **Receive thread**: Dedicated thread listens for incoming messages
- **Callback execution**: Executes client callbacks on receiver thread
- **Non-blocking sends**: Returns immediately after queuing message

### 4. Error Propagation

The SDK delivers server errors to the client:

- **Error parsing**: Extracts error code, reason, and context
- **Callback notification**: Invokes `onError()` with structured error object
- **Context preservation**: Includes related `clOrdID` when applicable

---

## Enumerations

The SDK uses enumerations to represent string constants in the protocol. This provides type safety and prevents typos.

### MessageType

**Purpose**: Discriminates message types in the JSON protocol

**Values**:

| Enum Value | String Value | Direction | Description |
|------------|--------------|-----------|-------------|
| `LOGIN` | `"LOGIN"` | Client → Server | Authentication request |
| `LOGIN_OK` | `"LOGIN_OK"` | Server → Client | Authentication success |
| `ORDER` | `"ORDER"` | Client → Server | Buy/sell order submission |
| `ORDER_ACK` | `"ORDER_ACK"` | Server → Client | Order receipt acknowledgment |
| `FILL` | `"FILL"` | Server → Client | Order execution confirmation |
| `TICKER` | `"TICKER"` | Server → Client | Market price update |
| `OFFER` | `"OFFER"` | Server → Client | Direct purchase offer |
| `ACCEPT_OFFER` | `"ACCEPT_OFFER"` | Client → Server | Accept/reject offer |
| `PRODUCTION_UPDATE` | `"PRODUCTION_UPDATE"` | Client → Server | Production completion |
| `INVENTORY_UPDATE` | `"INVENTORY_UPDATE"` | Server → Client | Inventory synchronization |
| `BALANCE_UPDATE` | `"BALANCE_UPDATE"` | Server → Client | Balance synchronization |
| `RESYNC` | `"RESYNC"` | Client → Server | Request missed events |
| `EVENT_DELTA` | `"EVENT_DELTA"` | Server → Client | Batch of missed events |
| `CANCEL` | `"CANCEL"` | Client → Server | Cancel order request |
| `ERROR` | `"ERROR"` | Server → Client | Error notification |
| `BROADCAST_NOTIFICATION` | `"BROADCAST_NOTIFICATION"` | Server → Client | Admin broadcast message |

**Java Implementation**:
```java
public enum MessageType {
    LOGIN("LOGIN"),
    LOGIN_OK("LOGIN_OK"),
    ORDER("ORDER"),
    ORDER_ACK("ORDER_ACK"),
    FILL("FILL"),
    TICKER("TICKER"),
    OFFER("OFFER"),
    ACCEPT_OFFER("ACCEPT_OFFER"),
    PRODUCTION_UPDATE("PRODUCTION_UPDATE"),
    INVENTORY_UPDATE("INVENTORY_UPDATE"),
    BALANCE_UPDATE("BALANCE_UPDATE"),
    RESYNC("RESYNC"),
    EVENT_DELTA("EVENT_DELTA"),
    CANCEL("CANCEL"),
    ERROR("ERROR"),
    BROADCAST_NOTIFICATION("BROADCAST_NOTIFICATION");
    
    private final String value;
    
    MessageType(String value) {
        this.value = value;
    }
    
    public String getValue() {
        return value;
    }
    
    @JsonValue
    public String toJson() {
        return value;
    }
    
    @JsonCreator
    public static MessageType fromJson(String value) {
        for (MessageType type : values()) {
            if (type.value.equals(value)) {
                return type;
            }
        }
        throw new IllegalArgumentException("Unknown message type: " + value);
    }
}
```

---

### OrderSide

**Purpose**: Specifies order direction (buy or sell)

**Values**:

| Enum Value | String Value | Description |
|------------|--------------|-------------|
| `BUY` | `"BUY"` | Purchase order - spend cash to acquire products |
| `SELL` | `"SELL"` | Sale order - sell products to gain cash |

**Usage**:
- Set in `OrderMessage.side` when submitting orders
- Returned in `FillMessage.side` when order executes

**Java Implementation**:
```java
public enum OrderSide {
    BUY("BUY"),
    SELL("SELL");
    
    private final String value;
    
    OrderSide(String value) {
        this.value = value;
    }
    
    public String getValue() {
        return value;
    }
    
    @JsonValue
    public String toJson() {
        return value;
    }
    
    @JsonCreator
    public static OrderSide fromJson(String value) {
        for (OrderSide side : values()) {
            if (side.value.equals(value)) {
                return side;
            }
        }
        throw new IllegalArgumentException("Unknown order side: " + value);
    }
}
```

---

### OrderMode

**Purpose**: Specifies order execution strategy

**Values**:

| Enum Value | String Value | Description | Guarantees |
|------------|--------------|-------------|------------|
| `MARKET` | `"MARKET"` | Execute immediately at best available price | Guarantees execution, not price |
| `LIMIT` | `"LIMIT"` | Execute only at specified price or better | Guarantees price, not execution |

**Usage**:
- Set in `OrderMessage.mode` when submitting orders
- `LIMIT` orders require `OrderMessage.limitPrice` to be set

**Java Implementation**:
```java
public enum OrderMode {
    MARKET("MARKET"),
    LIMIT("LIMIT");
    
    private final String value;
    
    OrderMode(String value) {
        this.value = value;
    }
    
    public String getValue() {
        return value;
    }
    
    @JsonValue
    public String toJson() {
        return value;
    }
    
    @JsonCreator
    public static OrderMode fromJson(String value) {
        for (OrderMode mode : values()) {
            if (mode.value.equals(value)) {
                return mode;
            }
        }
        throw new IllegalArgumentException("Unknown order mode: " + value);
    }
}
```

---

### Product

**Purpose**: Valid product codes in the market

**Values**:

| Enum Value | String Value | Description |
|------------|--------------|-------------|
| `GUACA` | `"GUACA"` | Guacamole (avocado-based product) |
| `SEBO` | `"SEBO"` | Tallow/fat product |
| `PALTA_OIL` | `"PALTA-OIL"` | Avocado oil |
| `FOSFO` | `"FOSFO"` | Phosphorescent material |
| `NUCREM` | `"NUCREM"` | Nuclear cream |
| `CASCAR_ALLOY` | `"CASCAR-ALLOY"` | Shell alloy |
| `PITA` | `"PITA"` | Pita bread/fiber |

**Usage**:
- Set in `OrderMessage.product`, `ProductionUpdateMessage.product`
- Returned in `FillMessage.product`, `TickerMessage.product`, `OfferMessage.product`
- Keys in `LoginOKMessage.inventory`, `LoginOKMessage.recipes`

**Java Implementation**:
```java
public enum Product {
    GUACA("GUACA"),
    SEBO("SEBO"),
    PALTA_OIL("PALTA-OIL"),
    FOSFO("FOSFO"),
    NUCREM("NUCREM"),
    CASCAR_ALLOY("CASCAR-ALLOY"),
    PITA("PITA");
    
    private final String value;
    
    Product(String value) {
        this.value = value;
    }
    
    public String getValue() {
        return value;
    }
    
    @JsonValue
    public String toJson() {
        return value;
    }
    
    @JsonCreator
    public static Product fromJson(String value) {
        for (Product product : values()) {
            if (product.value.equals(value)) {
                return product;
            }
        }
        throw new IllegalArgumentException("Unknown product: " + value);
    }
    
    public static Set<String> getAllValues() {
        return Arrays.stream(values())
            .map(Product::getValue)
            .collect(Collectors.toSet());
    }
}
```

---

### ErrorCode

**Purpose**: Categorizes server errors

**Values**:

| Enum Value | String Value | Severity | Description |
|------------|--------------|----------|-------------|
| `AUTH_FAILED` | `"AUTH_FAILED"` | Fatal | Authentication token is invalid |
| `INVALID_ORDER` | `"INVALID_ORDER"` | Error | Order validation failed |
| `INVALID_PRODUCT` | `"INVALID_PRODUCT"` | Error | Product code does not exist |
| `INVALID_QUANTITY` | `"INVALID_QUANTITY"` | Error | Quantity out of valid range |
| `DUPLICATE_ORDER_ID` | `"DUPLICATE_ORDER_ID"` | Error | Order ID already used |
| `UNAUTHORIZED_PRODUCTION` | `"UNAUTHORIZED_PRODUCTION"` | Error | Team cannot produce this product |
| `OFFER_EXPIRED` | `"OFFER_EXPIRED"` | Info | Offer no longer valid |
| `RATE_LIMIT_EXCEEDED` | `"RATE_LIMIT_EXCEEDED"` | Warning | Too many requests per second |
| `SERVICE_UNAVAILABLE` | `"SERVICE_UNAVAILABLE"` | Transient | Server temporarily unavailable |
| `INSUFFICIENT_INVENTORY` | `"INSUFFICIENT_INVENTORY"` | Error | Not enough inventory to sell |
| `INVALID_MESSAGE` | `"INVALID_MESSAGE"` | Error | Message parsing failed |

**Usage**:
- Returned in `ErrorMessage.code`
- Used in client error handling logic

**Java Implementation**:
```java
public enum ErrorCode {
    AUTH_FAILED("AUTH_FAILED", Severity.FATAL),
    INVALID_ORDER("INVALID_ORDER", Severity.ERROR),
    INVALID_PRODUCT("INVALID_PRODUCT", Severity.ERROR),
    INVALID_QUANTITY("INVALID_QUANTITY", Severity.ERROR),
    DUPLICATE_ORDER_ID("DUPLICATE_ORDER_ID", Severity.ERROR),
    UNAUTHORIZED_PRODUCTION("UNAUTHORIZED_PRODUCTION", Severity.ERROR),
    OFFER_EXPIRED("OFFER_EXPIRED", Severity.INFO),
    RATE_LIMIT_EXCEEDED("RATE_LIMIT_EXCEEDED", Severity.WARNING),
    SERVICE_UNAVAILABLE("SERVICE_UNAVAILABLE", Severity.TRANSIENT),
    INSUFFICIENT_INVENTORY("INSUFFICIENT_INVENTORY", Severity.ERROR),
    INVALID_MESSAGE("INVALID_MESSAGE", Severity.ERROR);
    
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
    
    @JsonValue
    public String toJson() {
        return value;
    }
    
    @JsonCreator
    public static ErrorCode fromJson(String value) {
        for (ErrorCode code : values()) {
            if (code.value.equals(value)) {
                return code;
            }
        }
        // Return unknown if not recognized (for forward compatibility)
        return null;
    }
    
    public enum Severity {
        FATAL,      // Must terminate application
        ERROR,      // Operation failed, can continue
        WARNING,    // Degraded performance, can continue
        INFO,       // Informational, no action needed
        TRANSIENT   // Temporary issue, retry recommended
    }
}
```

---

### OrderStatus

**Purpose**: Indicates order lifecycle state

**Values**:

| Enum Value | String Value | Description |
|------------|--------------|-------------|
| `PENDING` | `"PENDING"` | Order received, awaiting match |
| `FILLED` | `"FILLED"` | Order completely executed |
| `PARTIALLY_FILLED` | `"PARTIALLY_FILLED"` | Order partially executed |
| `CANCELLED` | `"CANCELLED"` | Order cancelled before full execution |

**Usage**:
- Returned in `OrderAckMessage.status`

**Java Implementation**:
```java
public enum OrderStatus {
    PENDING("PENDING"),
    FILLED("FILLED"),
    PARTIALLY_FILLED("PARTIALLY_FILLED"),
    CANCELLED("CANCELLED");
    
    private final String value;
    
    OrderStatus(String value) {
        this.value = value;
    }
    
    public String getValue() {
        return value;
    }
    
    @JsonValue
    public String toJson() {
        return value;
    }
    
    @JsonCreator
    public static OrderStatus fromJson(String value) {
        for (OrderStatus status : values()) {
            if (status.value.equals(value)) {
                return status;
            }
        }
        throw new IllegalArgumentException("Unknown order status: " + value);
    }
}
```

---

### RecipeType

**Purpose**: Categorizes production recipes

**Values**:

| Enum Value | String Value | Description |
|------------|--------------|-------------|
| `BASIC` | `"BASIC"` | No ingredients required, lower output |
| `PREMIUM` | `"PREMIUM"` | Requires ingredients, higher output (+30%) |

**Usage**:
- Returned in `Recipe.type` within `LoginOKMessage`

**Java Implementation**:
```java
public enum RecipeType {
    BASIC("BASIC"),
    PREMIUM("PREMIUM");
    
    private final String value;
    
    RecipeType(String value) {
        this.value = value;
    }
    
    public String getValue() {
        return value;
    }
    
    @JsonValue
    public String toJson() {
        return value;
    }
    
    @JsonCreator
    public static RecipeType fromJson(String value) {
        for (RecipeType type : values()) {
            if (type.value.equals(value)) {
                return type;
            }
        }
        throw new IllegalArgumentException("Unknown recipe type: " + value);
    }
}
```

---

## Connection Protocol

### Connection Lifecycle

```
1. Client calls conectar("localhost", 9000)
   ↓
2. SDK establishes WebSocket connection to ws://localhost:9000/ws
   ↓
3. Connection opens → SDK sets internal connected=true
   ↓
4. Client calls login("TK-TEAM-2025", listener)
   ↓
5. SDK sends LOGIN message
   ↓
6. Server responds with LOGIN_OK or ERROR
   ↓
7. SDK invokes listener.onLoginOk() or listener.onError()
   ↓
8. Client now ready to send orders, receive events
   ↓
9. SDK continuously receives messages, invokes callbacks
   ↓
10. On disconnect: SDK invokes listener.onConnectionLost()
```

### WebSocket Details

**Protocol**: WebSocket (RFC 6455)  
**Encoding**: UTF-8 text frames (JSON)  
**URL Format**: `ws://hostname:port/ws`  
**Message Format**: One JSON object per WebSocket text frame  
**Heartbeat**: Server sends PING every 30s, client responds with PONG

### Authentication Flow

```
Client → Server: {"type": "LOGIN", "token": "TK-...", "tz": "America/Bogota"}
Server → Client: {"type": "LOGIN_OK", "team": "...", "initialBalance": 10000, ...}
                 OR
Server → Client: {"type": "ERROR", "code": "AUTH_FAILED", "reason": "Invalid token"}
```

**Authentication Token Format**: `TK-{CLUSTER}-{YEAR}-{SPECIES}`  
Example: `TK-ANDROMEDA-2025-AVOCULTORES`

**Session Management**: Server allows up to 5 concurrent sessions per team. When limit is exceeded, oldest session is terminated.

---

## Message Types Reference

### Message Structure

All messages are JSON objects with a `type` field that identifies the message category.

**Common Fields Across All Messages:**

| Field | Type | Description |
|-------|------|-------------|
| `type` | String | Message discriminator (e.g., "LOGIN", "ORDER", "FILL") |
| `serverTime` | String | ISO-8601 timestamp (server → client messages only) |

### Client → Server Messages

| Type | Purpose | Triggers |
|------|---------|----------|
| `LOGIN` | Authenticate with server | Client calls `login()` |
| `ORDER` | Submit buy/sell order | Client calls `enviarOrden()` |
| `PRODUCTION_UPDATE` | Report production completion | Client calls `enviarProduccion()` |
| `ACCEPT_OFFER` | Accept/reject direct offer | Client calls `aceptarOferta()` |
| `RESYNC` | Request missed events | Client calls `resync()` |
| `CANCEL` | Cancel pending order | Client calls `cancelarOrden()` |

### Server → Client Messages

| Type | Purpose | Callback Invoked |
|------|---------|------------------|
| `LOGIN_OK` | Confirm authentication | `onLoginOk()` |
| `FILL` | Confirm order execution | `onFill()` |
| `TICKER` | Market price update (every 5s) | `onTicker()` |
| `OFFER` | Direct purchase offer | `onOffer()` |
| `ERROR` | Validation error or failure | `onError()` |
| `ORDER_ACK` | Acknowledge order receipt | (Optional callback) |
| `INVENTORY_UPDATE` | Inventory synchronization | (Optional callback) |
| `BALANCE_UPDATE` | Balance synchronization | (Optional callback) |
| `EVENT_DELTA` | Batch of missed fills | `onFill()` for each event |
| `BROADCAST_NOTIFICATION` | Admin broadcast announcement | `onBroadcast()` |

---

## Server Error Codes

When the server rejects an operation, it sends an `ERROR` message with a specific error code.

### Error Response Format

```json
{
  "type": "ERROR",
  "code": "ERROR_CODE",
  "reason": "Human-readable description",
  "clOrdID": "ORD-123",
  "timestamp": "2025-01-15T10:30:00Z"
}
```

### Error Code Catalog

#### Authentication Errors

**`AUTH_FAILED`**  
**Meaning**: Token validation failed  
**Common Causes**:
- Token does not exist in database
- Token format is incorrect (must start with "TK-")
- Token is empty or null

**Client Action**: Verify token in configuration, ensure team exists in system

---

#### Order Validation Errors

**`INVALID_ORDER`**  
**Meaning**: Order message failed validation  
**Common Causes**:
- Missing required fields (clOrdID, side, product, qty)
- Invalid field values

**Client Action**: Check all required fields are present and valid

---

**`INVALID_PRODUCT`**  
**Meaning**: Product code does not exist  
**Valid Products**: `GUACA`, `SEBO`, `PALTA-OIL`, `FOSFO`, `NUCREM`, `CASCAR-ALLOY`, `PITA`  
**Common Causes**:
- Typo in product name
- Case sensitivity (must be uppercase)

**Client Action**: Validate product against the valid products list

---

**`INVALID_QUANTITY`**  
**Meaning**: Quantity is out of valid range  
**Valid Range**: qty > 0  
**Common Causes**:
- Negative quantity
- Zero quantity

**Client Action**: Ensure quantity is a positive integer

---

**`DUPLICATE_ORDER_ID`**  
**Meaning**: clOrdID has already been used  
**Common Causes**:
- Reusing the same order ID
- Poor ID generation strategy

**Client Action**: Generate unique IDs (UUID or timestamp-based recommended)

---

**`INSUFFICIENT_INVENTORY`**  
**Meaning**: Team does not have enough inventory to sell  
**Common Causes**:
- Local inventory tracking is out of sync
- Race condition (sold inventory between check and send)

**Client Action**: Trigger resync or manual inventory verification

---

#### Production Errors

**`UNAUTHORIZED_PRODUCTION`**  
**Meaning**: Team cannot produce this product  
**Common Causes**:
- Attempting to produce product not in `authorizedProducts` list
- Product belongs to another species

**Client Action**: Check `LoginOKMessage.authorizedProducts` list

---

#### Offer Errors

**`OFFER_EXPIRED`**  
**Meaning**: Offer is no longer valid  
**Common Causes**:
- Took too long to accept (timeout)
- Buyer canceled the offer
- Another seller already accepted

**Client Action**: Accept offers more quickly or handle gracefully

---

#### Rate Limiting Errors

**`RATE_LIMIT_EXCEEDED`**  
**Meaning**: Sending requests too fast  
**Limit**: 10 requests per second per team  
**Common Causes**:
- Tight loop sending orders without delay
- Multiple threads sending concurrently

**Client Action**: Implement rate limiting (minimum 100ms between requests)

---

#### System Errors

**`SERVICE_UNAVAILABLE`**  
**Meaning**: Server cannot process request  
**Common Causes**:
- Database connection issues
- High server load
- Maintenance mode

**Client Action**: Retry with exponential backoff

---

**`INVALID_MESSAGE`**  
**Meaning**: Message could not be parsed  
**Common Causes**:
- Malformed JSON
- Missing `type` field
- Invalid JSON syntax

**Client Action**: Validate JSON before sending

---

## Data Structures (DTOs)

### Client → Server DTOs

#### LoginMessage

**Purpose**: Authenticate with the server using API token

**Fields**:

| Field | Type | Required | Description | Validation |
|-------|------|----------|-------------|------------|
| `type` | MessageType | Yes | Always `MessageType.LOGIN` | Hardcoded enum constant |
| `token` | String | Yes | API authentication key | Must start with "TK-" |
| `tz` | String | No | Client timezone | IANA timezone ID (e.g., "America/Bogota") |

**Example**:
```json
{
  "type": "LOGIN",
  "token": "TK-ANDROMEDA-2025-AVOCULTORES",
  "tz": "America/Bogota"
}
```

**SDK Validation**: Token must not be null/empty and must start with "TK-"

---

#### OrderMessage

**Purpose**: Submit a buy or sell order to the market

**Fields**:

| Field | Type | Required | Description | Validation |
|-------|------|----------|-------------|------------|
| `type` | MessageType | Yes | Always `MessageType.ORDER` | Hardcoded enum constant |
| `clOrdID` | String | Yes | Client order ID (must be unique) | Not null/empty, unique per session |
| `side` | OrderSide | Yes | Order direction | Must be `OrderSide.BUY` or `OrderSide.SELL` |
| `mode` | OrderMode | Yes | Execution mode | Must be `OrderMode.MARKET` or `OrderMode.LIMIT` |
| `product` | Product | Yes | Product code | Must be valid Product enum value |
| `qty` | Integer | Yes | Quantity to trade | Must be > 0 |
| `limitPrice` | Double | Conditional | Price limit (required if mode=LIMIT) | Must be > 0 if mode=LIMIT |
| `expiresAt` | String | No | Order expiration | ISO-8601 timestamp |
| `message` | String | No | Message for counterparty | Max 200 characters |
| `debugMode` | String | No | Debug flag | "AUTO_ACCEPT", "TEAM_ONLY", or null |

**Examples**:

Market Order (BUY):
```json
{
  "type": "ORDER",
  "clOrdID": "ORD-AVOCULTORES-1705334400-a1b2c3d4",
  "side": "BUY",
  "mode": "MARKET",
  "product": "FOSFO",
  "qty": 10,
  "message": "Necesito para producción premium"
}
```

Limit Order (SELL):
```json
{
  "type": "ORDER",
  "clOrdID": "ORD-AVOCULTORES-1705334401-e5f6g7h8",
  "side": "SELL",
  "mode": "LIMIT",
  "product": "PALTA-OIL",
  "qty": 25,
  "limitPrice": 28.50,
  "expiresAt": "2025-01-15T15:00:00Z",
  "message": "Premium quality oil"
}
```

**SDK Validation**:
- clOrdID: Not null/empty
- side: Valid OrderSide enum value
- mode: Valid OrderMode enum value
- product: Valid Product enum value
- qty: > 0
- limitPrice: Required and > 0 if mode=OrderMode.LIMIT
- message: ≤ 200 characters if present

**Java Example**:
```java
OrderMessage order = new OrderMessage();
order.setType(MessageType.ORDER);
order.setClOrdID("ORD-AVOCULTORES-1705334400-a1b2c3d4");
order.setSide(OrderSide.BUY);
order.setMode(OrderMode.MARKET);
order.setProduct(Product.FOSFO);
order.setQty(10);
order.setMessage("Necesito para producción premium");
```

---

#### ProductionUpdateMessage

**Purpose**: Notify server of production completion

**Fields**:

| Field | Type | Required | Description | Validation |
|-------|------|----------|-------------|------------|
| `type` | MessageType | Yes | Always `MessageType.PRODUCTION_UPDATE` | Hardcoded enum constant |
| `product` | Product | Yes | Product produced | Valid Product enum value |
| `quantity` | Integer | Yes | Quantity produced | Must be > 0 |

**Example**:
```json
{
  "type": "PRODUCTION_UPDATE",
  "product": "GUACA",
  "quantity": 17
}
```

**SDK Validation**:
- product: Valid Product enum value
- quantity: > 0

**Java Example**:
```java
ProductionUpdateMessage production = new ProductionUpdateMessage();
production.setType(MessageType.PRODUCTION_UPDATE);
production.setProduct(Product.GUACA);
production.setQuantity(17);
```

**Note**: The SDK does NOT validate if the team is authorized to produce this product or if the quantity calculation is correct. That is the client's responsibility.

---

#### AcceptOfferMessage

**Purpose**: Accept or reject a direct offer from another team

**Fields**:

| Field | Type | Required | Description | Validation |
|-------|------|----------|-------------|------------|
| `type` | MessageType | Yes | Always `MessageType.ACCEPT_OFFER` | Hardcoded enum constant |
| `offerId` | String | Yes | Offer ID from OfferMessage | Not null/empty |
| `accept` | Boolean | Yes | true=accept, false=reject | Must be boolean |
| `quantityOffered` | Integer | Conditional | Quantity to sell (if accept=true) | > 0 if accept=true |
| `priceOffered` | Double | Conditional | Price per unit (if accept=true) | > 0 if accept=true |

**Examples**:

Accept Offer:
```json
{
  "type": "ACCEPT_OFFER",
  "offerId": "OFFER-1705334567890",
  "accept": true,
  "quantityOffered": 15,
  "priceOffered": 23.00
}
```

Reject Offer:
```json
{
  "type": "ACCEPT_OFFER",
  "offerId": "OFFER-1705334567890",
  "accept": false
}
```

**SDK Validation**:
- offerId: Not null/empty
- accept: Must be boolean
- If accept=true: quantityOffered > 0 and priceOffered > 0

---

#### ResyncMessage

**Purpose**: Request missed events after reconnection

**Fields**:

| Field | Type | Required | Description | Validation |
|-------|------|----------|-------------|------------|
| `type` | MessageType | Yes | Always `MessageType.RESYNC` | Hardcoded enum constant |
| `lastSync` | String | Yes | Last known event timestamp | ISO-8601 format |

**Example**:
```json
{
  "type": "RESYNC",
  "lastSync": "2025-01-15T14:32:45Z"
}
```

**SDK Validation**:
- lastSync: Not null/empty, valid ISO-8601 format

---

#### CancelMessage

**Purpose**: Cancel a pending order

**Fields**:

| Field | Type | Required | Description | Validation |
|-------|------|----------|-------------|------------|
| `type` | MessageType | Yes | Always `MessageType.CANCEL` | Hardcoded enum constant |
| `clOrdID` | String | Yes | Order ID to cancel | Not null/empty |

**Example**:
```json
{
  "type": "CANCEL",
  "clOrdID": "ORD-AVOCULTORES-1705334400-a1b2c3d4"
}
```

**SDK Validation**:
- clOrdID: Not null/empty

---

### Server → Client DTOs

#### LoginOKMessage

**Purpose**: Confirm successful authentication and provide initial state

**Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `type` | MessageType | Always `MessageType.LOGIN_OK` |
| `team` | String | Team name |
| `species` | String | Species name (e.g., "Avocultores") |
| `initialBalance` | Double | Starting balance (for P&L calculation) |
| `currentBalance` | Double | Current cash balance |
| `inventory` | Map<Product, Integer> | Current inventory {product: quantity} |
| `authorizedProducts` | List<Product> | Products this team can produce |
| `recipes` | Map<Product, Recipe> | Production recipes {product: recipe} |
| `role` | TeamRole | Recursive algorithm parameters |
| `serverTime` | String | Server timestamp (ISO-8601) |

**Nested Structure - Recipe**:

| Field | Type | Description |
|-------|------|-------------|
| `type` | RecipeType | `RecipeType.BASIC` or `RecipeType.PREMIUM` |
| `ingredients` | Map<Product, Integer> | null for basic, {product: qty} for premium |
| `premiumBonus` | Double | Production multiplier (typically 1.30) |

**Nested Structure - TeamRole**:

| Field | Type | Description |
|-------|------|-------------|
| `branches` | Integer | Branching factor for recursive algorithm |
| `maxDepth` | Integer | Maximum recursion depth |
| `decay` | Double | Decay rate per level |
| `budget` | Double | Budget (not used in production calculation) |
| `baseEnergy` | Double | Base energy at level 0 |
| `levelEnergy` | Double | Energy increment per level |

**Example**:
```json
{
  "type": "LOGIN_OK",
  "team": "EquipoAndromeda",
  "species": "Avocultores",
  "initialBalance": 10000.00,
  "currentBalance": 10000.00,
  "inventory": {
    "PALTA-OIL": 0,
    "FOSFO": 0,
    "GUACA": 0
  },
  "authorizedProducts": ["PALTA-OIL", "GUACA", "SEBO"],
  "recipes": {
    "PALTA-OIL": {
      "type": "BASIC",
      "ingredients": null,
      "premiumBonus": 1.0
    },
    "GUACA": {
      "type": "PREMIUM",
      "ingredients": {
        "FOSFO": 5,
        "PITA": 3
      },
      "premiumBonus": 1.30
    }
  },
  "role": {
    "branches": 2,
    "maxDepth": 4,
    "decay": 0.7651,
    "budget": 10000.0,
    "baseEnergy": 3.0,
    "levelEnergy": 2.0
  },
  "serverTime": "2025-01-15T10:00:00Z"
}
```

**Client Responsibility**: Store this data for:
- Calculating production quantities (using `role` parameters)
- Validating production authorization (using `authorizedProducts`)
- Tracking initial balance for P&L calculation

---

#### FillMessage

**Purpose**: Confirm order execution (trade completed)

**Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `type` | MessageType | Always `MessageType.FILL` |
| `clOrdID` | String | Your order ID that was filled |
| `fillQty` | Integer | Quantity filled in this execution |
| `fillPrice` | Double | Price per unit |
| `side` | OrderSide | `OrderSide.BUY` or `OrderSide.SELL` |
| `product` | Product | Product traded |
| `counterparty` | String | Other team's name |
| `counterpartyMessage` | String | Message from other team |
| `serverTime` | String | Execution timestamp (ISO-8601) |
| `remainingQty` | Integer | (Optional) Quantity still unfilled (partial fills) |
| `totalQty` | Integer | (Optional) Total order quantity (partial fills) |

**Example**:
```json
{
  "type": "FILL",
  "clOrdID": "ORD-AVOCULTORES-1705334400-a1b2c3d4",
  "fillQty": 10,
  "fillPrice": 18.25,
  "side": "BUY",
  "product": "FOSFO",
  "counterparty": "EquipoMonjes",
  "counterpartyMessage": "Fresh from the mines!",
  "serverTime": "2025-01-15T10:05:23Z"
}
```

**Client Responsibility**: Update local state:
- If `side=OrderSide.BUY`: Deduct cash (fillQty × fillPrice), add to inventory
- If `side=OrderSide.SELL`: Add cash (fillQty × fillPrice), remove from inventory

**Java Example**:
```java
@Override
public void onFill(FillMessage fill) {
    if (fill.getSide() == OrderSide.BUY) {
        double cost = fill.getFillQty() * fill.getFillPrice();
        state.setBalance(state.getBalance() - cost);
        state.addInventory(fill.getProduct(), fill.getFillQty());
    } else if (fill.getSide() == OrderSide.SELL) {
        double revenue = fill.getFillQty() * fill.getFillPrice();
        state.setBalance(state.getBalance() + revenue);
        state.removeInventory(fill.getProduct(), fill.getFillQty());
    }
}
```

---

#### TickerMessage

**Purpose**: Periodic market price updates (sent every 5 seconds)

**Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `type` | MessageType | Always `MessageType.TICKER` |
| `product` | Product | Product code |
| `bestBid` | Double | Highest buy price (can be null if no buyers) |
| `bestAsk` | Double | Lowest sell price (can be null if no sellers) |
| `mid` | Double | (bestBid + bestAsk) / 2 (can be null) |
| `volume24h` | Integer | Trading volume in last 24 hours |
| `serverTime` | String | Timestamp (ISO-8601) |

**Example**:
```json
{
  "type": "TICKER",
  "product": "FOSFO",
  "bestBid": 17.80,
  "bestAsk": 18.20,
  "mid": 18.00,
  "volume24h": 1520,
  "serverTime": "2025-01-15T10:05:00Z"
}
```

**Client Responsibility**: Store `mid` price for:
- P&L calculation (valuing inventory)
- Estimating order costs before submitting
- Market trend analysis

**Note**: `bestBid`, `bestAsk`, and `mid` can be null if there are no active orders on that side of the market.

---

#### OfferMessage

**Purpose**: Direct purchase offer from another team

**Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `type` | MessageType | Always `MessageType.OFFER` |
| `offerId` | String | Unique offer identifier |
| `buyer` | String | Team making the offer |
| `product` | Product | Product requested |
| `quantityRequested` | Integer | Quantity wanted |
| `maxPrice` | Double | Maximum price per unit willing to pay |
| `expiresIn` | Integer | Milliseconds until expiration (can be null) |
| `timestamp` | String | Offer creation time (ISO-8601) |

**Example**:
```json
{
  "type": "OFFER",
  "offerId": "OFFER-1705334567890",
  "buyer": "EquipoBeta",
  "product": "PITA",
  "quantityRequested": 15,
  "maxPrice": 23.00,
  "expiresIn": 30000,
  "timestamp": "2025-01-15T10:10:00Z"
}
```

**Client Responsibility**: Decide whether to accept:
- Check if you have sufficient inventory
- Compare `maxPrice` to current market price
- Accept via `aceptarOferta()` if desirable

**Note**: If `expiresIn` is null, the offer has no automatic expiration (though the buyer can cancel at any time).

---

#### ErrorMessage

**Purpose**: Notify client of validation errors or operation failures

**Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `type` | MessageType | Always `MessageType.ERROR` |
| `code` | ErrorCode | Error code enum (see Error Codes section) |
| `reason` | String | Human-readable explanation |
| `clOrdID` | String | (Optional) Related order ID if applicable |
| `timestamp` | String | Error timestamp (ISO-8601) |

**Example**:
```json
{
  "type": "ERROR",
  "code": "INSUFFICIENT_INVENTORY",
  "reason": "Cannot sell 50 PITA - only 25 available",
  "clOrdID": "ORD-AVOCULTORES-1705334450-x9y8z7",
  "timestamp": "2025-01-15T10:12:30Z"
}
```

**SDK Responsibility**: Parse and deliver to client via `onError()` callback

---

#### OrderAckMessage

**Purpose**: Acknowledge order receipt (NOT execution)

**Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `type` | MessageType | Always `MessageType.ORDER_ACK` |
| `clOrdID` | String | Order ID acknowledged |
| `status` | OrderStatus | Order status enum value |
| `serverTime` | String | Acknowledgment timestamp (ISO-8601) |

**Example**:
```json
{
  "type": "ORDER_ACK",
  "clOrdID": "ORD-AVOCULTORES-1705334400-a1b2c3d4",
  "status": "PENDING",
  "serverTime": "2025-01-15T10:05:15Z"
}
```

**Important**: `ORDER_ACK` with status "PENDING" means the server received your order and is processing it. It does NOT mean the order was executed. Wait for `FILL` message for execution confirmation.

---

#### InventoryUpdateMessage

**Purpose**: Server-initiated inventory synchronization (sent after production)

**Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `type` | MessageType | Always `MessageType.INVENTORY_UPDATE` |
| `inventory` | Map<Product, Integer> | Complete inventory snapshot {product: quantity} |
| `serverTime` | String | Update timestamp (ISO-8601) |

**Example**:
```json
{
  "type": "INVENTORY_UPDATE",
  "inventory": {
    "PALTA-OIL": 13,
    "FOSFO": 5,
    "PITA": 3,
    "GUACA": 0
  },
  "serverTime": "2025-01-15T10:15:00Z"
}
```

**Client Responsibility**: Update local inventory state to match server

---

#### BalanceUpdateMessage

**Purpose**: Server-initiated balance synchronization

**Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `type` | MessageType | Always `MessageType.BALANCE_UPDATE` |
| `balance` | Double | Current cash balance |
| `serverTime` | String | Update timestamp (ISO-8601) |

**Example**:
```json
{
  "type": "BALANCE_UPDATE",
  "balance": 12450.75,
  "serverTime": "2025-01-15T10:15:00Z"
}
```

**Client Responsibility**: Update local balance state to match server

---

#### EventDeltaMessage

**Purpose**: Batch of missed events after resync

**Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `type` | MessageType | Always `MessageType.EVENT_DELTA` |
| `events` | List<FillMessage> | Array of missed fill events |
| `serverTime` | String | Resync timestamp (ISO-8601) |

**Example**:
```json
{
  "type": "EVENT_DELTA",
  "events": [
    {
      "type": "FILL",
      "clOrdID": "ORD-AVOCULTORES-1705334400-a1b2c3d4",
      "fillQty": 10,
      "fillPrice": 18.00,
      "side": "SELL",
      "product": "FOSFO",
      "counterparty": "EquipoGamma",
      "counterpartyMessage": "Thanks!",
      "serverTime": "2025-01-15T14:33:12Z"
    },
    {
      "type": "FILL",
      "clOrdID": "ORD-AVOCULTORES-1705334401-b2c3d4e5",
      "fillQty": 5,
      "fillPrice": 22.00,
      "side": "BUY",
      "product": "PITA",
      "counterparty": "EquipoDelta",
      "counterpartyMessage": "Fresh harvest",
      "serverTime": "2025-01-15T14:34:01Z"
    }
  ],
  "serverTime": "2025-01-15T14:35:00Z"
}
```

**SDK Responsibility**: For each fill in `events`, invoke `onFill()` callback

---

#### BroadcastNotificationMessage

**Purpose**: Admin broadcast message to all connected clients

**Fields**:

| Field | Type | Description |
|-------|------|-------------|
| `type` | MessageType | Always `MessageType.BROADCAST_NOTIFICATION` |
| `message` | String | Announcement text from admin |
| `sender` | String | Always "admin" |
| `serverTime` | String | Broadcast timestamp (ISO-8601) |

**Example**:
```json
{
  "type": "BROADCAST_NOTIFICATION",
  "message": "Tournament starts in 5 minutes! Good luck traders!",
  "sender": "admin",
  "serverTime": "2025-01-15T14:55:00Z"
}
```

**Client Responsibility**: Display announcement to user

**Use Cases**:
- Tournament start/end announcements
- Server maintenance warnings
- Important market updates
- Rule changes or clarifications

**Java Example**:
```java
@Override
public void onBroadcast(BroadcastNotificationMessage broadcast) {
    System.out.println("\n" + "=".repeat(50));
    System.out.println("📢 ADMIN ANNOUNCEMENT");
    System.out.println("=".repeat(50));
    System.out.println(broadcast.getMessage());
    System.out.println("=".repeat(50) + "\n");
}
```

---

## SDK Interface

### ConectorBolsa Class

The main SDK class that handles all server communication.

#### Constructor

```java
public ConectorBolsa()
```

Creates a new connector instance. Does not establish connection.

---

#### conectar()

```java
public void conectar(String host, int port) throws ConexionFallidaException
```

**Purpose**: Establishes WebSocket connection to the server

**Parameters**:
- `host` - Server hostname or IP address (e.g., "localhost", "192.168.1.100")
- `port` - Server port (typically 9000)

**Throws**:
- `ConexionFallidaException` - If connection fails (network error, timeout, server unreachable)

**Behavior**:
1. Creates WebSocket URL: `ws://{host}:{port}/ws`
2. Establishes TCP connection
3. Performs WebSocket handshake
4. Blocks for up to 5 seconds waiting for connection
5. Sets internal `connected` flag to true
6. Returns control to caller

**Thread Safety**: This method is thread-safe

**Example**:
```java
ConectorBolsa connector = new ConectorBolsa();
try {
    connector.conectar("localhost", 9000);
    System.out.println("Connected successfully");
} catch (ConexionFallidaException e) {
    System.err.println("Connection failed: " + e.getMessage());
}
```

---

#### login()

```java
public void login(String apiKey, EventListener listener)
```

**Purpose**: Authenticates with the server and registers event callbacks

**Parameters**:
- `apiKey` - Team authentication token (format: "TK-...")
- `listener` - Callback handler implementing EventListener interface

**Throws**:
- `IllegalArgumentException` - If apiKey is null, empty, or doesn't start with "TK-"
- `IllegalStateException` - If not connected to server

**Behavior**:
1. Validates apiKey format
2. Stores listener reference for future callbacks
3. Sends LOGIN message to server
4. Returns immediately (response arrives via callback)

**Response Handling**:
- Success: Server sends LOGIN_OK → SDK calls `listener.onLoginOk()`
- Failure: Server sends ERROR → SDK calls `listener.onError()`

**Thread Safety**: This method is thread-safe

**Example**:
```java
connector.login("TK-ANDROMEDA-2025-AVOCULTORES", myClientInstance);
// Response comes via callback:
// - onLoginOk() if successful
// - onError() if authentication fails
```

---

#### enviarOrden()

```java
public void enviarOrden(OrderMessage orden) throws IllegalStateException
```

**Purpose**: Sends a buy or sell order to the market

**Parameters**:
- `orden` - Order details (clOrdID, side, product, qty, etc.)

**Throws**:
- `IllegalStateException` - If not logged in (must call login() first)

**Behavior**:
1. Validates that login has completed
2. Serializes order to JSON
3. Sends via WebSocket
4. Returns immediately

**Response Handling**:
- Immediate: Server sends ORDER_ACK → (optional callback)
- Delayed: Server sends FILL → SDK calls `listener.onFill()`
- Error: Server sends ERROR → SDK calls `listener.onError()`

**Thread Safety**: This method is thread-safe (uses internal semaphore)

**Example**:
```java
OrderMessage order = new OrderMessage();
order.setType(MessageType.ORDER);
order.setClOrdID(generateOrderID());
order.setSide(OrderSide.BUY);
order.setMode(OrderMode.MARKET);
order.setProduct(Product.FOSFO);
order.setQty(10);
order.setMessage("For premium production");

connector.enviarOrden(order);
// Response comes via callback:
// - onFill() when order executes
// - onError() if validation fails
```

---

#### enviarProduccion()

```java
public void enviarProduccion(String producto, int cantidad) throws IllegalArgumentException
```

**Purpose**: Notifies server of production completion

**Parameters**:
- `producto` - Product produced (e.g., "GUACA")
- `cantidad` - Quantity produced (must be > 0)

**Throws**:
- `IllegalArgumentException` - If cantidad <= 0

**Behavior**:
1. Validates quantity is positive
2. Creates PRODUCTION_UPDATE message
3. Sends via WebSocket
4. Returns immediately

**Response Handling**:
- Success: Server sends INVENTORY_UPDATE → (optional callback)
- Error: Server sends ERROR → SDK calls `listener.onError()`

**Thread Safety**: This method is thread-safe

**Important**: The SDK does NOT validate:
- Whether the team is authorized to produce this product
- Whether the quantity calculation is correct
- Whether ingredients were consumed

These are client responsibilities.

**Example**:
```java
connector.enviarProduccion(Product.GUACA, 17);
// Response comes via callback:
// - (optional) INVENTORY_UPDATE
// - onError() if not authorized
```

---

#### aceptarOferta()

```java
public void aceptarOferta(String offerId, int cantidad, double precio)
```

**Purpose**: Accepts a direct offer from another team

**Parameters**:
- `offerId` - Offer ID from OfferMessage
- `cantidad` - Quantity to sell
- `precio` - Price per unit

**Behavior**:
1. Creates ACCEPT_OFFER message with accept=true
2. Sends via WebSocket
3. Returns immediately

**Response Handling**:
- Success: Server sends FILL → SDK calls `listener.onFill()`
- Error: Server sends ERROR → SDK calls `listener.onError()` (e.g., OFFER_EXPIRED)

**Thread Safety**: This method is thread-safe

**Example**:
```java
@Override
public void onOffer(OfferMessage offer) {
    // Decide to accept
    connector.aceptarOferta(offer.getOfferId(), 
                           offer.getQuantityRequested(), 
                           offer.getMaxPrice());
}
```

---

#### rechazarOferta()

```java
public void rechazarOferta(String offerId)
```

**Purpose**: Explicitly rejects a direct offer

**Parameters**:
- `offerId` - Offer ID from OfferMessage

**Behavior**:
1. Creates ACCEPT_OFFER message with accept=false
2. Sends via WebSocket
3. Returns immediately

**Note**: Rejecting is optional. Offers expire automatically if ignored.

**Thread Safety**: This method is thread-safe

---

#### resync()

```java
public void resync(Instant ultimaSincronizacion)
```

**Purpose**: Requests missed events after reconnection

**Parameters**:
- `ultimaSincronizacion` - Timestamp of last known event

**Behavior**:
1. Converts Instant to ISO-8601 string
2. Creates RESYNC message
3. Sends via WebSocket
4. Returns immediately

**Response Handling**:
- Server sends EVENT_DELTA → SDK calls `listener.onFill()` for each event

**Thread Safety**: This method is thread-safe

**Example**:
```java
// After reconnecting
Instant lastEventTime = Instant.parse("2025-01-15T14:32:45Z");
connector.resync(lastEventTime);
// Response comes via callback:
// - onFill() for each missed event
```

---

#### cancelarOrden()

```java
public void cancelarOrden(String clOrdID)
```

**Purpose**: Cancels a pending order

**Parameters**:
- `clOrdID` - Order ID to cancel

**Behavior**:
1. Creates CANCEL message
2. Sends via WebSocket
3. Returns immediately

**Response Handling**:
- Success: Order is removed from order book (no specific confirmation)
- Error: Server sends ERROR if order doesn't exist or already filled

**Thread Safety**: This method is thread-safe

---

#### desconectar()

```java
public void desconectar()
```

**Purpose**: Closes WebSocket connection gracefully

**Behavior**:
1. Sends WebSocket close frame
2. Waits for server acknowledgment
3. Closes TCP socket
4. Sets internal `connected` flag to false

**Thread Safety**: This method is thread-safe

**Example**:
```java
// At program shutdown
connector.desconectar();
System.out.println("Disconnected from server");
```

---

#### isConnected()

```java
public boolean isConnected()
```

**Purpose**: Checks current connection status

**Returns**: true if WebSocket is connected, false otherwise

**Thread Safety**: This method is thread-safe

---

### EventListener Interface

Interface that client applications must implement to receive server events.

#### onLoginOk()

```java
void onLoginOk(LoginOKMessage message)
```

**Purpose**: Called when authentication succeeds

**Parameters**:
- `message` - Login confirmation with initial state

**When Called**: After successful login

**Thread Context**: Executed on SDK receiver thread

**Client Responsibilities**:
- Store initial balance for P&L calculation
- Store authorized products list
- Store recipes for production validation
- Store role parameters for production algorithm

**Example**:
```java
@Override
public void onLoginOk(LoginOKMessage msg) {
    state.setInitialBalance(msg.getInitialBalance());
    state.setCurrentBalance(msg.getCurrentBalance());
    state.setInventory(msg.getInventory());
    state.setAuthorizedProducts(msg.getAuthorizedProducts());
    state.setRecipes(msg.getRecipes());
    state.setRole(msg.getRole());
    
    System.out.println("✅ Logged in as " + msg.getTeam());
    System.out.println("💰 Balance: $" + msg.getCurrentBalance());
}
```

---

#### onFill()

```java
void onFill(FillMessage fill)
```

**Purpose**: Called when an order is executed

**Parameters**:
- `fill` - Execution details (price, quantity, counterparty)

**When Called**: 
- After market order matches
- After limit order is triggered
- During resync (for each missed fill)

**Thread Context**: Executed on SDK receiver thread

**Client Responsibilities**:
- Update balance (add/subtract cash based on side)
- Update inventory (add/subtract products based on side)
- Display trade confirmation to user

**Example**:
```java
@Override
public void onFill(FillMessage fill) {
    if (fill.getSide() == OrderSide.BUY) {
        double cost = fill.getFillQty() * fill.getFillPrice();
        state.setBalance(state.getBalance() - cost);
        
        int currentQty = state.getInventory().getOrDefault(fill.getProduct(), 0);
        state.getInventory().put(fill.getProduct(), currentQty + fill.getFillQty());
        
        System.out.printf("💰 BOUGHT: %d %s @ $%.2f = -$%.2f%n", 
            fill.getFillQty(), fill.getProduct().getValue(), fill.getFillPrice(), cost);
    } else if (fill.getSide() == OrderSide.SELL) {
        double revenue = fill.getFillQty() * fill.getFillPrice();
        state.setBalance(state.getBalance() + revenue);
        
        int currentQty = state.getInventory().get(fill.getProduct());
        state.getInventory().put(fill.getProduct(), currentQty - fill.getFillQty());
        
        System.out.printf("💵 SOLD: %d %s @ $%.2f = +$%.2f%n", 
            fill.getFillQty(), fill.getProduct().getValue(), fill.getFillPrice(), revenue);
    }
    
    if (fill.getCounterpartyMessage() != null) {
        System.out.println("   💬 \"" + fill.getCounterpartyMessage() + "\"");
    }
}
```

---

#### onTicker()

```java
void onTicker(TickerMessage ticker)
```

**Purpose**: Called every 5 seconds with market price updates

**Parameters**:
- `ticker` - Current market prices for a product

**When Called**: Periodically (every 5 seconds per product)

**Thread Context**: Executed on SDK receiver thread

**Client Responsibilities**:
- Update price cache for P&L calculation
- Optionally display market status
- Use for trading strategy decisions

**Example**:
```java
@Override
public void onTicker(TickerMessage ticker) {
    if (ticker.getMid() != null) {
        state.getCurrentPrices().put(ticker.getProduct(), ticker.getMid());
    }
    
    // Optional: Display market activity
    // System.out.printf("📊 %s: bid=%.2f ask=%.2f mid=%.2f vol=%d%n",
    //     ticker.getProduct(), ticker.getBestBid(), ticker.getBestAsk(), 
    //     ticker.getMid(), ticker.getVolume24h());
}
```

---

#### onOffer()

```java
void onOffer(OfferMessage offer)
```

**Purpose**: Called when another team makes a direct purchase offer

**Parameters**:
- `offer` - Offer details (product, quantity, price, expiration)

**When Called**: When another team sends a direct offer

**Thread Context**: Executed on SDK receiver thread

**Client Responsibilities**:
- Decide whether to accept or reject
- Check inventory availability
- Compare offer price to market price
- Call `aceptarOferta()` or `rechazarOferta()` if desired

**Example**:
```java
@Override
public void onOffer(OfferMessage offer) {
    System.out.printf("📬 OFFER from %s: %d %s @ max $%.2f%n",
        offer.getBuyer(), offer.getQuantityRequested(), 
        offer.getProduct(), offer.getMaxPrice());
    
    // Check inventory
    int available = state.getInventory().getOrDefault(offer.getProduct(), 0);
    if (available < offer.getQuantityRequested()) {
        System.out.println("   ⚠️ Insufficient inventory");
        return;
    }
    
    // Check price
    double marketPrice = state.getCurrentPrices().getOrDefault(offer.getProduct(), 0.0);
    if (offer.getMaxPrice() >= marketPrice * 0.95) { // Accept if ≥ 95% of market
        connector.aceptarOferta(offer.getOfferId(), 
                               offer.getQuantityRequested(), 
                               offer.getMaxPrice());
        System.out.println("   ✅ Accepted offer");
    } else {
        System.out.println("   ❌ Price too low");
    }
}
```

---

#### onError()

```java
void onError(ErrorMessage error)
```

**Purpose**: Called when server rejects an operation

**Parameters**:
- `error` - Error details (code, reason, context)

**When Called**: After validation failure or operation rejection

**Thread Context**: Executed on SDK receiver thread

**Client Responsibilities**:
- Handle different error codes appropriately
- Display error to user
- Trigger corrective actions (e.g., resync on inventory mismatch)

**Example**:
```java
@Override
public void onError(ErrorMessage error) {
    switch (error.getCode()) {
        case AUTH_FAILED:
            System.err.println("❌ Authentication failed: " + error.getReason());
            System.exit(1);
            break;
            
        case INSUFFICIENT_INVENTORY:
            System.err.println("🐛 Inventory mismatch: " + error.getReason());
            // Trigger resync
            connector.resync(Instant.now().minusSeconds(60));
            break;
            
        case OFFER_EXPIRED:
            System.out.println("ℹ️ Offer expired");
            break;
            
        case RATE_LIMIT_EXCEEDED:
            System.err.println("⏱️ Rate limit exceeded, pausing...");
            try { Thread.sleep(1000); } catch (InterruptedException e) {}
            break;
            
        default:
            System.err.println("❌ Error [" + error.getCode() + "]: " + error.getReason());
    }
}
```

---

#### onConnectionLost()

```java
void onConnectionLost(Exception exception)
```

**Purpose**: Called when WebSocket connection is lost

**Parameters**:
- `exception` - Exception that caused the disconnection

**When Called**: 
- Network failure
- Server shutdown
- Manual disconnect
- Timeout

**Thread Context**: Executed on SDK receiver thread

**Client Responsibilities**:
- Notify user of disconnection
- Attempt reconnection if desired
- Save state before attempting recovery

**Example**:
```java
@Override
public void onConnectionLost(Exception e) {
    System.err.println("⚠️ Connection lost: " + e.getMessage());
    
    // Save current state
    saveSnapshot();
    
    // Attempt reconnection
    try {
        Thread.sleep(5000); // Wait 5 seconds
        connector.conectar("localhost", 9000);
        connector.login(apiKey, this);
        
        // Resync after reconnect
        connector.resync(lastEventTimestamp);
        
    } catch (Exception reconnectError) {
        System.err.println("❌ Reconnection failed: " + reconnectError.getMessage());
    }
}
```

---

#### onBroadcast()

```java
void onBroadcast(BroadcastNotificationMessage broadcast)
```

**Purpose**: Called when server admin sends a broadcast announcement

**Parameters**:
- `broadcast` - Announcement details (message, sender, timestamp)

**When Called**: When admin sends system-wide announcement

**Thread Context**: Executed on SDK receiver thread

**Client Responsibilities**:
- Display announcement prominently to user
- Log announcement for later reference
- Optionally pause trading or alert user if critical

**Example**:
```java
@Override
public void onBroadcast(BroadcastNotificationMessage broadcast) {
    // Display prominently
    System.out.println("\n" + "=".repeat(60));
    System.out.println("📢 ADMIN ANNOUNCEMENT @ " + broadcast.getServerTime());
    System.out.println("=".repeat(60));
    System.out.println(broadcast.getMessage());
    System.out.println("=".repeat(60) + "\n");
    
    // Log to file for reference
    logBroadcast(broadcast);
    
    // Optional: Play sound or show popup for important announcements
    if (broadcast.getMessage().toLowerCase().contains("tournament")) {
        playNotificationSound();
    }
}
```

**Common Broadcast Messages**:
- `"Tournament starts in 5 minutes!"`
- `"Trading will pause in 1 minute for system maintenance"`
- `"All orders have been cancelled - market reset complete"`
- `"Tournament ends in 30 seconds - start liquidating inventory!"`

---

## Threading Model

### SDK Threading Architecture

The SDK uses two threads:

1. **Main Thread**: Where client code runs (calls SDK methods)
2. **Receiver Thread**: WebSocket message receiver (invokes callbacks)

### Thread Responsibilities

**Main Thread**:
- Calls SDK methods: `conectar()`, `login()`, `enviarOrden()`, etc.
- Client business logic
- User interface (console, GUI)

**Receiver Thread**:
- Listens for incoming WebSocket messages
- Parses JSON to Java objects
- Invokes callback methods: `onLoginOk()`, `onFill()`, `onTicker()`, etc.

### Thread Safety Guarantees

**SDK Methods (Main → SDK)**:
- All public methods are thread-safe
- Uses internal semaphore to prevent simultaneous sends
- Safe to call from multiple threads

**Callbacks (SDK → Client)**:
- All callbacks execute on receiver thread
- Callbacks for same message type are sequential (never concurrent)
- Callbacks for different message types may interleave

### Client Threading Considerations

**What Clients Must Do**:

1. **Keep Callbacks Fast**: Don't block receiver thread with long operations
   ```java
   @Override
   public void onFill(FillMessage fill) {
       // ✅ Good: Quick state update
       state.updateBalance(fill);
       
       // ❌ Bad: Long computation
       // performComplexAnalysis(); // Blocks other callbacks
       
       // ✅ Good: Offload to another thread
       executor.submit(() -> performComplexAnalysis());
   }
   ```

2. **Synchronize Shared State**: If main thread reads state that callbacks modify
   ```java
   public class EstadoCliente {
       private final Object lock = new Object();
       private double balance;
       
       public void setBalance(double balance) {
           synchronized (lock) {
               this.balance = balance;
           }
       }
       
       public double getBalance() {
           synchronized (lock) {
               return balance;
           }
       }
   }
   ```

3. **Use Concurrent Collections**: For maps/lists modified by callbacks
   ```java
   private Map<String, Integer> inventory = new ConcurrentHashMap<>();
   private Map<String, Double> prices = new ConcurrentHashMap<>();
   ```

**What Clients Should NOT Do**:
- ❌ Call `Thread.sleep()` in callbacks
- ❌ Perform I/O operations in callbacks (file writes, database queries)
- ❌ Show blocking dialogs/prompts in callbacks
- ❌ Assume callbacks execute on main thread

---

## Validation Rules

The SDK performs **format validation** before sending messages. Business logic validation is the client's responsibility.

### Order Validation (SDK Layer)

**Required Fields**:
- `clOrdID`: Not null, not empty
- `side`: Valid OrderSide enum value (BUY or SELL)
- `mode`: Valid OrderMode enum value (MARKET or LIMIT)
- `product`: Valid Product enum value
- `qty`: Greater than 0

**Conditional Fields**:
- `limitPrice`: Required if mode=OrderMode.LIMIT, must be > 0

**Optional Fields**:
- `message`: If present, ≤ 200 characters
- `expiresAt`: If present, valid ISO-8601 format

**What SDK Does NOT Validate**:
- Whether team has sufficient balance (client should validate)
- Whether team has sufficient inventory for sell orders (client should validate)
- Whether price is reasonable

**Java Example with Enums**:
```java
// Validate order before sending
public void validateOrder(OrderMessage order) throws ValidationException {
    if (order.getClOrdID() == null || order.getClOrdID().isEmpty()) {
        throw new ValidationException("clOrdID is required");
    }
    
    if (order.getSide() == null) {
        throw new ValidationException("side is required");
    }
    
    if (order.getMode() == null) {
        throw new ValidationException("mode is required");
    }
    
    if (order.getProduct() == null) {
        throw new ValidationException("product is required");
    }
    
    if (order.getQty() <= 0) {
        throw new ValidationException("quantity must be positive");
    }
    
    if (order.getMode() == OrderMode.LIMIT && 
        (order.getLimitPrice() == null || order.getLimitPrice() <= 0)) {
        throw new ValidationException("LIMIT orders require positive limitPrice");
    }
    
    if (order.getMessage() != null && order.getMessage().length() > 200) {
        throw new ValidationException("message exceeds 200 characters");
    }
}
```

### Production Validation (SDK Layer)

**Required Fields**:
- `product`: Valid Product enum value
- `quantity`: Greater than 0

**What SDK Does NOT Validate**:
- Whether team is authorized to produce this product
- Whether quantity calculation is correct
- Whether ingredients are available

**Java Example with Enums**:
```java
// Client validates authorization before producing
public void producir(Product product, boolean premium) 
        throws ProductoNoAutorizadoException {
    
    // Client validates (NOT SDK)
    if (!state.getAuthorizedProducts().contains(product)) {
        throw new ProductoNoAutorizadoException(product, state.getAuthorizedProducts());
    }
    
    // Calculate quantity (client responsibility)
    int quantity = calculateProductionQuantity(product, premium);
    
    // SDK just sends the message
    connector.enviarProduccion(product, quantity);
}
```

### Product Enum Usage

The SDK uses the `Product` enum to represent all valid products:

```java
// All valid products are enum constants
Product.GUACA
Product.SEBO
Product.PALTA_OIL  // Note: underscore in Java, hyphen in JSON
Product.FOSFO
Product.NUCREM
Product.CASCAR_ALLOY
Product.PITA

// Get all product string values
Set<String> allProducts = Product.getAllValues();
// Returns: ["GUACA", "SEBO", "PALTA-OIL", "FOSFO", "NUCREM", "CASCAR-ALLOY", "PITA"]

// Parse from JSON string
Product product = Product.fromJson("PALTA-OIL");  // Returns Product.PALTA_OIL

// Convert to JSON string
String jsonValue = Product.FOSFO.getValue();  // Returns "FOSFO"
```

**Benefits of using enums**:
- Compile-time type safety (impossible to use invalid product)
- IDE autocomplete support
- No typo errors
- Centralized product catalog

---

## Exception Handling

### SDK Exceptions

Exceptions thrown by SDK methods (synchronous errors):

#### ConexionFallidaException

**Thrown By**: `conectar()`  
**Meaning**: Failed to establish initial connection  
**Common Causes**:
- Server is not running
- Incorrect host/port
- Network unreachable
- Firewall blocking connection

**Example**:
```java
try {
    connector.conectar("localhost", 9000);
} catch (ConexionFallidaException e) {
    System.err.println("Cannot connect: " + e.getMessage());
    // Check server status, verify network
}
```

---

#### IllegalStateException

**Thrown By**: `enviarOrden()`, `enviarProduccion()`, etc.  
**Meaning**: Operation called before login  
**Common Causes**:
- Forgot to call `login()` first
- Login failed but code continued

**Example**:
```java
try {
    connector.enviarOrden(order);
} catch (IllegalStateException e) {
    System.err.println("Not logged in: " + e.getMessage());
    // Ensure login() was called and succeeded
}
```

---

#### IllegalArgumentException

**Thrown By**: `login()`, `enviarProduccion()`  
**Meaning**: Invalid parameter value  
**Common Causes**:
- API key doesn't start with "TK-"
- Quantity is ≤ 0

**Example**:
```java
try {
    connector.login("INVALID-TOKEN", listener);
} catch (IllegalArgumentException e) {
    System.err.println("Invalid API key format: " + e.getMessage());
}
```

---

### Server Errors (Asynchronous)

Errors sent by server arrive via `onError()` callback:

```java
@Override
public void onError(ErrorMessage error) {
    // Handle based on error code using enum
    switch (error.getCode()) {
        case AUTH_FAILED:
            // Fatal: cannot proceed
            System.err.println("❌ Authentication failed: " + error.getReason());
            System.exit(1);
            break;
            
        case INVALID_PRODUCT:
        case INVALID_QUANTITY:
        case INVALID_ORDER:
            // Bug: client validation failed
            System.err.println("🐛 Validation bug: " + error.getReason());
            break;
            
        case INSUFFICIENT_INVENTORY:
            // Sync issue: trigger resync
            System.err.println("⚠️ Inventory mismatch: " + error.getReason());
            connector.resync(Instant.now().minusSeconds(60));
            break;
            
        case OFFER_EXPIRED:
            // Expected: timing issue, no action needed
            System.out.println("ℹ️ Offer expired");
            break;
            
        case RATE_LIMIT_EXCEEDED:
            // Slow down
            System.err.println("⏱️ Rate limit exceeded, pausing...");
            try { Thread.sleep(1000); } catch (InterruptedException e) {}
            break;
            
        case SERVICE_UNAVAILABLE:
            // Retry with backoff
            System.err.println("⚠️ Service unavailable, will retry");
            break;
            
        default:
            // Unknown error: log and continue
            System.err.println("❌ Unknown error: " + error.getReason());
    }
}
```

---

## Appendix A: Order ID Generation

Order IDs must be unique per session. Recommended approach:

```java
public static String generateOrderID(String teamName) {
    long timestamp = System.currentTimeMillis() / 1000;
    String uuid = UUID.randomUUID().toString().substring(0, 8);
    return String.format("ORD-%s-%d-%s", teamName, timestamp, uuid);
}
```

**Example Output**: `ORD-AVOCULTORES-1705334400-a1b2c3d4`

**Alternative**: Use UUID only
```java
public static String generateOrderID() {
    return "ORD-" + UUID.randomUUID().toString();
}
```

---

## Appendix B: Maven Dependencies

Required dependencies for SDK implementation:

```xml
<dependencies>
    <!-- WebSocket Client -->
    <dependency>
        <groupId>org.java-websocket</groupId>
        <artifactId>Java-WebSocket</artifactId>
        <version>1.5.3</version>
    </dependency>
    
    <!-- JSON Serialization -->
    <dependency>
        <groupId>com.fasterxml.jackson.core</groupId>
        <artifactId>jackson-databind</artifactId>
        <version>2.15.0</version>
    </dependency>
    
    <!-- Optional: For ISO-8601 date handling -->
    <dependency>
        <groupId>com.fasterxml.jackson.datatype</groupId>
        <artifactId>jackson-datatype-jsr310</artifactId>
        <version>2.15.0</version>
    </dependency>
</dependencies>
```

---

## Appendix C: Connection Troubleshooting

| Symptom | Possible Cause | Solution |
|---------|----------------|----------|
| `ConexionFallidaException` during `conectar()` | Server not running | Start server, verify with `netstat -an \| grep 9000` |
| `AUTH_FAILED` after `login()` | Invalid token | Check token format, verify team exists in database |
| No callbacks invoked | Listener not registered | Ensure `login()` was called with listener |
| `IllegalStateException` on `enviarOrden()` | Not logged in | Call `login()` first, wait for `onLoginOk()` |
| Connection drops randomly | Network unstable | Implement reconnection logic in `onConnectionLost()` |
| `RATE_LIMIT_EXCEEDED` errors | Sending too fast | Add 100ms delay between orders |

---

## Appendix D: Quick Reference

### Message Type Summary

| Client → Server | Server → Client |
|-----------------|-----------------|
| LOGIN | LOGIN_OK |
| ORDER | FILL |
| PRODUCTION_UPDATE | TICKER |
| ACCEPT_OFFER | OFFER |
| RESYNC | ERROR |
| CANCEL | ORDER_ACK |
| | INVENTORY_UPDATE |
| | BALANCE_UPDATE |
| | EVENT_DELTA |
| | BROADCAST_NOTIFICATION |

### Error Code Quick Reference

| Code | Severity | Action |
|------|----------|--------|
| AUTH_FAILED | Fatal | Exit program |
| INVALID_ORDER | Bug | Fix validation |
| INVALID_PRODUCT | Bug | Fix validation |
| INVALID_QUANTITY | Bug | Fix validation |
| DUPLICATE_ORDER_ID | Bug | Fix ID generation |
| UNAUTHORIZED_PRODUCTION | Validation | Check authorized products |
| OFFER_EXPIRED | Info | Accept faster |
| RATE_LIMIT_EXCEEDED | Warning | Slow down |
| SERVICE_UNAVAILABLE | Transient | Retry with backoff |
| INSUFFICIENT_INVENTORY | Sync | Trigger resync |
| INVALID_MESSAGE | Bug | Fix JSON |

---

**End of SDK Specification**
