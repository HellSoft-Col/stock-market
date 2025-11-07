# Debug Mode System

## Overview

The stock market platform includes a comprehensive debug mode system that allows testing and development features to be toggled on/off globally. This is essential for running the platform in production while maintaining development capabilities.

## Configuration

### Config Files

#### Development (`config.yaml`)
```yaml
market:
  debugModeEnabled: true  # Debug features enabled for development/testing
```

#### Production (`config.production.yaml`)
```yaml
market:
  debugModeEnabled: false  # Debug features disabled in production
```

### Runtime Toggle

The debug mode can be toggled at runtime through the admin panel or API without restarting the server. The setting is persisted in MongoDB's `system_settings` collection.

## Debug Features

### 1. Auto-Accept Orders

**Description:** Orders with `debugMode: "AUTO_ACCEPT"` are instantly matched by the server without requiring a counterparty.

**Use Cases:**
- Testing order flow
- Development of trading algorithms
- Demonstration without real market participants

**Example (Java Client):**
```java
OrderMessage order = new OrderMessage();
order.setType("ORDER");
order.setClOrdID("TEST-" + System.currentTimeMillis());
order.setSide("BUY");
order.setProduct("FOSFO");
order.setQty(10);
order.setMode("MARKET");
order.setDebugMode("AUTO_ACCEPT");  // This triggers instant fill

client.sendMessage(order);
```

**Example (Web UI):**
- Go to Trading tab
- Check the "Auto-Accept" checkbox
- Place an order
- Order fills instantly

**Example (JavaScript):**
```javascript
const orderMessage = {
    type: "ORDER",
    clOrdID: `TEST-${Date.now()}`,
    side: "BUY",
    product: "FOSFO",
    qty: 10,
    mode: "MARKET",
    debugMode: "AUTO_ACCEPT"  // Instant fill
};

socket.send(JSON.stringify(orderMessage));
```

### 2. Error Injection

**Description:** Simulate various error conditions for testing error handling.

**Available Error Types:**
- `INSUFFICIENT_BALANCE` - Test balance validation
- `UNAUTHORIZED_PRODUCT` - Test product authorization
- `DISCONNECT_CLIENT` - Test disconnection handling
- `OFFER_EXPIRED` - Test offer expiration

**Example:**
```javascript
const errorMessage = {
    type: "DEBUG_ERROR",
    errorType: "INSUFFICIENT_BALANCE"
};
socket.send(JSON.stringify(errorMessage));
```

### 3. Production Testing

Test production algorithms without affecting real inventory.

## When Debug Mode is Disabled

When `debugModeEnabled: false`, the following happens:

1. **Orders with debugMode rejected:**
   ```
   Error: "debug mode is disabled - this operation is not allowed in production mode"
   ```

2. **Web UI changes:**
   - Auto-Accept checkbox is hidden/disabled
   - Debug tab features are grayed out
   - Visual indicator shows "REAL MODE" in UI

3. **Java clients get errors:**
   ```
   {
     "type": "ERROR",
     "error": "DEBUG_MODE_DISABLED",
     "message": "Debug mode is disabled in production"
   }
   ```

## Admin Panel Control

### Toggle Debug Mode

**Endpoint:** `POST /admin/api/debug-mode`

**Request:**
```json
{
  "enabled": false,
  "updatedBy": "admin@example.com"
}
```

**Response:**
```json
{
  "success": true,
  "debugMode": false,
  "message": "Debug mode disabled successfully"
}
```

### Check Current Status

**Endpoint:** `GET /admin/api/debug-mode`

**Response:**
```json
{
  "enabled": false,
  "updatedAt": "2024-01-15T10:30:00Z",
  "updatedBy": "admin@example.com"
}
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     Client Request                       │
│        (Order with debugMode: "AUTO_ACCEPT")            │
└───────────────────────┬─────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│              OfferGenerator.GenerateTargetedOffer        │
│                                                          │
│  1. Check: debugModeService.ValidateDebugRequest()     │
│  2. If disabled → Return Error                          │
│  3. If enabled → Process AUTO_ACCEPT                    │
└───────────────────────┬─────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│                 DebugModeService                        │
│                                                          │
│  • Loads setting from MongoDB on startup               │
│  • Caches in memory with mutex                         │
│  • Validates all debug requests                        │
│  • Updates persist to database                         │
└─────────────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│                      MongoDB                            │
│                 system_settings                         │
│                                                          │
│  {                                                      │
│    "_id": "debugModeEnabled",                          │
│    "value": true,                                      │
│    "updatedAt": "2024-01-15T10:30:00Z",              │
│    "updatedBy": "admin@example.com"                   │
│  }                                                      │
└─────────────────────────────────────────────────────────┘
```

## Best Practices

### Development Environment
1. **Enable debug mode** in `config.yaml`
2. Use AUTO_ACCEPT for rapid testing
3. Test error scenarios with error injection
4. Monitor logs for debug operations

### Staging Environment
1. **Enable debug mode** for final testing
2. Test toggle functionality
3. Verify production mode blocks debug features
4. Test admin panel controls

### Production Environment
1. **Disable debug mode** in `config.production.yaml`
2. Monitor for rejected debug requests
3. Only enable temporarily for troubleshooting
4. Always disable after debugging session
5. Log who enabled/disabled debug mode

## Monitoring

### Logs to Watch

```json
{
  "level": "info",
  "message": "Debug mode setting updated",
  "debugMode": false,
  "updatedBy": "admin@example.com",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

```json
{
  "level": "warn",
  "message": "Debug mode request rejected - debug mode is disabled",
  "clOrdID": "TEST-12345",
  "debugMode": "AUTO_ACCEPT"
}
```

### Metrics to Track

- Number of debug mode toggles per day
- Number of rejected debug requests
- Number of AUTO_ACCEPT orders processed
- Audit trail of who changed debug mode

## Security Considerations

1. **Admin-only access:** Only admin users can toggle debug mode
2. **Audit logging:** All changes are logged with user attribution
3. **Production safeguard:** config.production.yaml defaults to disabled
4. **Runtime validation:** Every debug request is validated
5. **No backdoors:** No way to bypass debug mode check

## Troubleshooting

### "Debug mode is disabled" errors in development

**Symptom:** Getting errors when trying to use AUTO_ACCEPT

**Solution:**
1. Check `config.yaml` has `debugModeEnabled: true`
2. Check database `system_settings` collection
3. Restart server after config change
4. Verify logs show "Loaded debug mode setting from database"

### Debug features still work in production

**Symptom:** Debug features working when they shouldn't

**Solution:**
1. Verify `config.production.yaml` has `debugModeEnabled: false`
2. Check admin panel didn't enable it
3. Check database setting
4. Review access logs for admin panel usage

### Cannot toggle debug mode from admin panel

**Symptom:** Admin API returns errors

**Solution:**
1. Check authentication tokens
2. Verify MongoDB connection
3. Check system_settings collection permissions
4. Review server logs for detailed error messages
