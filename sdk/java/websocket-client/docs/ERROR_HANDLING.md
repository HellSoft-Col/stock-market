# Error Handling Best Practices - Java SDK

## Table of Contents
1. [Overview](#overview)
2. [Error Types](#error-types)
3. [Implementing onError](#implementing-onerror)
4. [Common Error Codes](#common-error-codes)
5. [Best Practices](#best-practices)
6. [Recovery Strategies](#recovery-strategies)
7. [Examples](#examples)

---

## Overview

Proper error handling is crucial for building robust trading bots. The Stock Market SDK provides comprehensive error reporting through the `onError()` callback method. This guide shows you how to handle errors effectively.

---

## Error Types

### 1. Server Errors
Errors returned by the server when your trading bot violates business rules or constraints.

### 2. Connection Errors  
Network-related errors that occur when connection is lost or unstable.

### 3. Validation Errors
Errors caught by the SDK before sending messages to the server (e.g., null values, invalid formats).

---

## Implementing onError

The `onError()` method in `EventListener` is called whenever an error occurs:

```java
@Override
public void onError(ErrorMessage message) {
    // Extract error details
    ErrorCode code = message.getCode();
    String reason = message.getReason();
    String clOrdID = message.getClOrdID();  // May be null
    
    // Log the error
    log.error("Error [{}]: {} (Order: {})", code, reason, clOrdID);
    
    // Handle based on error code
    switch (code) {
        case INSUFFICIENT_BALANCE:
            handleInsufficientBalance(message);
            break;
        case UNAUTHORIZED_PRODUCT:
            handleUnauthorizedProduct(message);
            break;
        case INVALID_ORDER:
            handleInvalidOrder(message);
            break;
        case OFFER_EXPIRED:
            handleOfferExpired(message);
            break;
        case DUPLICATE_ORDER_ID:
            handleDuplicateOrderId(message);
            break;
        default:
            handleUnknownError(message);
    }
}
```

---

## Common Error Codes

### `INSUFFICIENT_BALANCE`
**Cause:** Attempting to buy more than your available balance allows.

**Example:**
```
Balance: $1,000
Order: BUY 100 GUACA @ $15.00
Cost: $1,500
Result: ERROR - INSUFFICIENT_BALANCE
```

**How to Fix:**
```java
private void handleInsufficientBalance(ErrorMessage error) {
    log.warn("Insufficient balance for order: {}", error.getClOrdID());
    
    // Cancel pending large orders
    cancelLargeOrders();
    
    // Adjust strategy to smaller quantities
    maxOrderSize = Math.min(maxOrderSize, calculateMaxAffordableQty());
    
    // Wait for fills to increase balance
    pauseTrading(5_000); // 5 seconds
}
```

---

### `UNAUTHORIZED_PRODUCT`
**Cause:** Attempting to trade a product your team is not authorized for.

**Example:**
```
Authorized: [FOSFO, PITA]
Order: BUY 10 GUACA
Result: ERROR - UNAUTHORIZED_PRODUCT
```

**How to Fix:**
```java
private void handleUnauthorizedProduct(ErrorMessage error) {
    String product = extractProductFromError(error);
    
    // Add to blacklist
    unauthorizedProducts.add(product);
    
    // Update strategy to only trade authorized products
    updateTradingStrategy();
    
    log.warn("Product {} not authorized, added to blacklist", product);
}
```

---

### `INVALID_ORDER`
**Cause:** Order format is incorrect or missing required fields.

**Common Issues:**
- Negative or zero quantity
- Invalid price (negative or zero for LIMIT orders)
- Missing required fields
- Invalid enum values

**Example:**
```java
// ❌ BAD - Will cause INVALID_ORDER
OrderMessage bad = OrderMessage.builder()
    .clOrdID(orderIdGen.next())
    .side(OrderSide.BUY)
    .product(Product.GUACA)
    .qty(0)  // ❌ Zero quantity
    .mode(OrderMode.LIMIT)
    .limitPrice(-5.0)  // ❌ Negative price
    .build();

// ✅ GOOD
OrderMessage good = OrderMessage.builder()
    .clOrdID(orderIdGen.next())
    .side(OrderSide.BUY)
    .product(Product.GUACA)
    .qty(10)  // ✅ Positive quantity
    .mode(OrderMode.LIMIT)
    .limitPrice(15.50)  // ✅ Positive price
    .build();
```

**How to Fix:**
```java
private void handleInvalidOrder(ErrorMessage error) {
    log.error("Invalid order format: {}", error.getReason());
    
    // Review order creation logic
    // Add validation before sending
    // Fix the bug in your code
}

// Add validation helper
private boolean isValidOrder(OrderMessage order) {
    if (order.getQty() <= 0) return false;
    if (order.getMode() == OrderMode.LIMIT && order.getLimitPrice() <= 0) return false;
    if (order.getClOrdID() == null || order.getClOrdID().isEmpty()) return false;
    return true;
}
```

---

### `OFFER_EXPIRED`
**Cause:** Trying to accept an offer that has already expired.

**How to Fix:**
```java
private void handleOfferExpired(ErrorMessage error) {
    String offerId = extractOfferIdFromError(error);
    
    // Remove from pending offers
    pendingOffers.remove(offerId);
    
    // Don't retry - offer is gone
    log.debug("Offer {} expired, removed from queue", offerId);
}

@Override
public void onOffer(OfferMessage offer) {
    // Calculate if we can respond in time
    Integer expiresIn = offer.getExpiresIn();  // milliseconds
    
    if (expiresIn != null && expiresIn < 1000) {
        // Less than 1 second - might be too tight
        log.warn("Offer expires very soon: {}ms", expiresIn);
    }
    
    // Respond quickly
    if (shouldAcceptOffer(offer)) {
        acceptOfferImmediately(offer);
    }
}
```

---

### `DUPLICATE_ORDER_ID`
**Cause:** Sending an order with an ID that was already used.

**This should NEVER happen if you use `OrderIdGenerator`!**

**How to Fix:**
```java
// ❌ WRONG - Can cause duplicates
.clOrdID("ORDER-" + System.currentTimeMillis())

// ✅ CORRECT - Thread-safe, guaranteed unique
private final OrderIdGenerator orderIdGen = new OrderIdGenerator("TEAM-A");

OrderMessage order = OrderMessage.builder()
    .clOrdID(orderIdGen.next())  // ✅ Always unique
    ...
    .build();
```

**If it still happens:**
```java
private void handleDuplicateOrderId(ErrorMessage error) {
    log.error("CRITICAL: Duplicate order ID detected: {}", error.getClOrdID());
    
    // This indicates a bug in your code!
    // Check if you're:
    // 1. Using OrderIdGenerator correctly
    // 2. Not creating multiple generators with same prefix
    // 3. Not manually creating order IDs
    
    // Emergency: Reset counter (only in development!)
    if (isDevelopmentMode()) {
        orderIdGen.reset();
    }
}
```

---

### `AUTH_FAILED`
**Cause:** Invalid token or authentication issue.

**How to Fix:**
```java
@Override
public void onError(ErrorMessage error) {
    if (error.getCode() == ErrorCode.AUTH_FAILED) {
        log.error("Authentication failed: {}", error.getReason());
        
        // Stop trading immediately
        stopAllTrading();
        
        // Disconnect
        connector.desconectar();
        
        // Alert operator
        sendAlert("Bot authentication failed - manual intervention required");
        
        // Don't retry - requires manual token fix
    }
}
```

---

### `RATE_LIMIT_EXCEEDED`
**Cause:** Sending too many requests too quickly.

**How to Fix:**
```java
private final Semaphore rateLimiter = new Semaphore(10); // Max 10 concurrent

private void handleRateLimitExceeded(ErrorMessage error) {
    log.warn("Rate limit exceeded, backing off");
    
    // Exponential backoff
    int backoffMs = 1000 * (int) Math.pow(2, rateLimitRetries);
    rateLimitRetries++;
    
    try {
        Thread.sleep(backoffMs);
    } catch (InterruptedException e) {
        Thread.currentThread().interrupt();
    }
    
    // Reset on success
    if (rateLimitRetries > 0) {
        rateLimitRetries = Math.max(0, rateLimitRetries - 1);
    }
}

// Better: Add rate limiting proactively
private void sendOrderWithRateLimit(OrderMessage order) {
    try {
        rateLimiter.acquire();
        connector.enviarOrden(order);
    } catch (InterruptedException e) {
        Thread.currentThread().interrupt();
    } finally {
        // Release after 100ms
        scheduler.schedule(() -> rateLimiter.release(), 100, TimeUnit.MILLISECONDS);
    }
}
```

---

## Best Practices

### 1. Always Implement onError
```java
// ❌ BAD - No error handling
@Override
public void onError(ErrorMessage message) {
    // Empty - errors ignored!
}

// ✅ GOOD - Comprehensive error handling
@Override
public void onError(ErrorMessage message) {
    logError(message);
    updateMetrics(message);
    handleByErrorCode(message);
    alertIfCritical(message);
}
```

### 2. Log Errors with Context
```java
// ❌ BAD - No context
log.error("Error occurred");

// ✅ GOOD - Full context
log.error("Order error [code={}, clOrdID={}, reason={}]: {}", 
    message.getCode(),
    message.getClOrdID(),
    message.getReason(),
    message.getTimestamp());
```

### 3. Track Error Metrics
```java
private final Map<ErrorCode, Integer> errorCounts = new ConcurrentHashMap<>();

@Override
public void onError(ErrorMessage message) {
    // Track error frequency
    errorCounts.merge(message.getCode(), 1, Integer::sum);
    
    // Alert if too many errors
    if (errorCounts.values().stream().mapToInt(Integer::intValue).sum() > 100) {
        log.error("Too many errors! Stopping bot.");
        emergency Stop();
    }
}
```

### 4. Implement Circuit Breaker
```java
private int consecutiveErrors = 0;
private static final int MAX_CONSECUTIVE_ERRORS = 5;

@Override
public void onError(ErrorMessage message) {
    consecutiveErrors++;
    
    if (consecutiveErrors >= MAX_CONSECUTIVE_ERRORS) {
        log.error("Circuit breaker triggered after {} errors", consecutiveErrors);
        pauseTrading();
        consecutiveErrors = 0;
    }
}

// Reset on success
@Override
public void onFill(FillMessage message) {
    consecutiveErrors = 0;  // Reset on successful trade
}
```

### 5. Don't Retry Unrecoverable Errors
```java
private boolean isRecoverable(ErrorCode code) {
    return switch (code) {
        case RATE_LIMIT_EXCEEDED -> true;  // Can retry after backoff
        case OFFER_EXPIRED -> false;  // Offer is gone, can't retry
        case AUTH_FAILED -> false;  // Requires manual intervention
        case INSUFFICIENT_BALANCE -> false;  // Need to wait for fills
        case INVALID_ORDER -> false;  // Bug in code, can't auto-fix
        default -> false;
    };
}

@Override
public void onError(ErrorMessage message) {
    if (isRecoverable(message.getCode())) {
        retryWithBackoff(message);
    } else {
        handlePermanentFailure(message);
    }
}
```

---

## Recovery Strategies

### Strategy 1: Exponential Backoff
```java
private void retryWithExponentialBackoff(Runnable action, int attempt) {
    int delayMs = (int) (1000 * Math.pow(2, attempt));
    scheduler.schedule(action, delayMs, TimeUnit.MILLISECONDS);
}
```

### Strategy 2: Circuit Breaker with Auto-Reset
```java
private void initCircuitBreaker() {
    // Auto-reset every 60 seconds
    scheduler.scheduleAtFixedRate(() -> {
        if (consecutiveErrors > 0) {
            log.info("Circuit breaker auto-reset");
            consecutiveErrors = 0;
            resumeTrading();
        }
    }, 60, 60, TimeUnit.SECONDS);
}
```

### Strategy 3: Fallback Strategy
```java
@Override
public void onError(ErrorMessage message) {
    if (message.getCode() == ErrorCode.INSUFFICIENT_BALANCE) {
        // Fallback: Switch to smaller orders
        currentStrategy = new ConservativeStrategy(smallerOrderSize);
    }
}
```

---

## Complete Example: Production-Ready Error Handler

```java
public class RobustTradingBot implements EventListener {
    private final OrderIdGenerator orderIdGen = new OrderIdGenerator("TEAM-A");
    private final Map<ErrorCode, Integer> errorCounts = new ConcurrentHashMap<>();
    private final Set<String> unauthorizedProducts = ConcurrentHashMap.newKeySet();
    private int consecutiveErrors = 0;
    private boolean isPaused = false;
    
    @Override
    public void onError(ErrorMessage message) {
        // 1. Log with full context
        log.error("Trading error [code={}, orderId={}, reason={}]", 
            message.getCode(), 
            message.getClOrdID(), 
            message.getReason());
        
        // 2. Track metrics
        errorCounts.merge(message.getCode(), 1, Integer::sum);
        consecutiveErrors++;
        
        // 3. Circuit breaker
        if (consecutiveErrors >= 5) {
            log.error("Circuit breaker triggered!");
            pauseTrading();
            return;
        }
        
        // 4. Handle by error code
        switch (message.getCode()) {
            case INSUFFICIENT_BALANCE:
                handleInsufficientBalance(message);
                break;
            case UNAUTHORIZED_PRODUCT:
                String product = extractProduct(message.getReason());
                unauthorizedProducts.add(product);
                log.warn("Blacklisted product: {}", product);
                break;
            case RATE_LIMIT_EXCEEDED:
                handleRateLimit();
                break;
            case OFFER_EXPIRED:
                // Don't retry - offer is gone
                break;
            case INVALID_ORDER:
                // Bug in code - alert developer
                sendDeveloperAlert(message);
                break;
            case AUTH_FAILED:
                emergencyStop("Authentication failed");
                break;
            default:
                log.error("Unhandled error code: {}", message.getCode());
        }
    }
    
    @Override
    public void onFill(FillMessage message) {
        // Reset error counter on success
        consecutiveErrors = 0;
    }
    
    private void pauseTrading() {
        isPaused = true;
        log.warn("Trading paused");
        
        // Auto-resume after 60 seconds
        scheduler.schedule(() -> {
            isPaused = false;
            consecutiveErrors = 0;
            log.info("Trading resumed");
        }, 60, TimeUnit.SECONDS);
    }
    
    private void handleInsufficientBalance(ErrorMessage error) {
        // Reduce order sizes
        maxOrderSize = Math.min(maxOrderSize, 5);
        log.info("Reduced max order size to {}", maxOrderSize);
    }
    
    private void handleRateLimit() {
        // Exponential backoff
        int backoffMs = 1000 * (int) Math.pow(2, rateLimitRetries++);
        try {
            Thread.sleep(Math.min(backoffMs, 30000)); // Max 30s
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }
}
```

---

## Debugging Tips

### Enable Debug Logging
```java
// In your logging configuration
logging.level.tech.hellsoft.trading=DEBUG
```

### Add Error Counts to Dashboard
```java
private void printErrorSummary() {
    log.info("=== Error Summary ===");
    errorCounts.forEach((code, count) -> 
        log.info("{}: {} occurrences", code, count));
}
```

### Test Error Scenarios
```java
// Use the SDK Emulator to test error handling
// web/index.html > SDK Emulator > onError
```

---

## Summary Checklist

- ✅ Implement `onError()` with comprehensive handling
- ✅ Use `OrderIdGenerator` to avoid DUPLICATE_ORDER_ID
- ✅ Validate orders before sending
- ✅ Log errors with full context
- ✅ Track error metrics
- ✅ Implement circuit breaker
- ✅ Don't retry unrecoverable errors
- ✅ Use exponential backoff for rate limits
- ✅ Handle AUTH_FAILED with emergency stop
- ✅ Reset error counters on success

**Remember:** Good error handling is the difference between a bot that crashes and one that runs reliably 24/7!
