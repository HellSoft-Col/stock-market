# Authentication Fix - Transparent Internal Waiting

## Problem
When calling `conectar()` followed immediately by `enviarOrden()` or `enviarActualizacionProduccion()`, the SDK threw:
```
IllegalStateException: "Not authenticated"
```

This happened because authentication is asynchronous - `conectar()` returns after establishing the WebSocket connection, but before the LOGIN_OK response arrives.

## Solution
**The SDK now automatically waits for authentication internally** in all send methods. No changes needed to your code!

### How it Works
1. `conectar()` initializes a `CompletableFuture` that completes when LOGIN_OK arrives
2. All send methods (`enviarOrden`, `enviarActualizacionProduccion`, etc.) call an internal `waitForAuthentication()` method
3. This method checks if already authenticated - if yes, returns immediately
4. If not yet authenticated, it waits for the `CompletableFuture` to complete (with timeout)
5. Once authenticated, the message is sent

### Usage - No Changes Required!
```java
ConectorBolsa connector = new ConectorBolsa();
connector.addListener(new MyEventListener());

// Connect
connector.conectar("wss://trading.hellsoft.tech/ws", "your-token");

// Send immediately - SDK waits internally for authentication
connector.enviarOrden(order); // ✅ Works!
connector.enviarActualizacionProduccion(update); // ✅ Works!
```

## What Changed

### Modified Methods
1. **`enviarOrden(OrderMessage)`** - Now calls `waitForAuthentication()` before sending
2. **`enviarCancelacion(String)`** - Now calls `waitForAuthentication()` before sending
3. **`enviarActualizacionProduccion(ProductionUpdateMessage)`** - Now calls `waitForAuthentication()` before sending
4. **`enviarRespuestaOferta(AcceptOfferMessage)`** - Now calls `waitForAuthentication()` before sending

### New Internal Method
- **`waitForAuthentication()`** - Private method that blocks until authentication completes or times out

### Technical Details
- Uses `CompletableFuture<LoginOKMessage>` internally
- Timeout based on `config.getConnectionTimeout()`
- Thread-safe with `volatile` field
- Completes exceptionally on AUTH_FAILED errors
- Null/argument validation happens **before** waiting (fail-fast)

## Benefits
1. **Zero API Changes** - Existing code works without modifications
2. **Simple** - No extra public methods to remember
3. **Transparent** - Authentication handling is invisible to users
4. **Safe** - Proper timeout handling prevents hanging
5. **Fast** - If already authenticated, no waiting occurs

## Error Handling
If authentication fails or times out, send methods throw:
```java
IllegalStateException: "Authentication failed" 
// or
IllegalStateException: "Authentication timed out"
```

## Files Modified
- `src/main/java/tech/hellsoft/trading/ConectorBolsa.java`
- `src/test/java/tech/hellsoft/trading/ConectorBolsaTest.java` (test expectations updated)

## Testing
- ✅ All 417 tests pass
- ✅ Build successful
- ✅ No breaking changes
