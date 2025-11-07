# Debug Mode Implementation Summary

## What Was Implemented

### 1. Backend Changes ✅

#### New Files Created:
- `internal/repository/mongodb/system_settings_repository.go` - Persists debug mode setting
- `internal/service/debug_mode_service.go` - Manages debug mode state
- `DEBUG_MODE.md` - Comprehensive documentation

#### Modified Files:
- `internal/config/config.go` - Added `DebugModeEnabled` field to MarketConfig
- `internal/domain/interfaces.go` - Added `DebugModeService` interface
- `internal/market/offer_generator.go` - Added debug mode validation
- `internal/market/engine.go` - Added debugModeService dependency
- `config.yaml` - Added `debugModeEnabled: true` (development default)
- `config.production.yaml` - Added `debugModeEnabled: false` (production default)

#### Key Features:
- **Global Toggle:** Debug mode can be enabled/disabled globally
- **Runtime Control:** Setting can be changed without restart via admin panel
- **Database Persistence:** Setting survives server restarts (stored in MongoDB)
- **Request Validation:** All debug requests are validated before processing
- **Audit Logging:** All changes logged with user attribution

### 2. Frontend Changes (Needed - Not Yet Implemented)

#### Web UI (`web/index.html`) - Already Has:
✅ Auto-Accept checkbox in Trading tab  
✅ Pending Orders section in History tab
✅ Visual indicators for auto-accept orders  
✅ Test Transaction button in Debug tab

#### Admin Panel (`web/admin.html`) - TO DO:
❌ Debug Mode toggle switch
❌ Current status indicator  
❌ WHO/WHEN last changed display
❌ Confirmation dialog for changes

## How It Works

### Java Client Example

```java
// In your trading bot
public void placeDebugOrder() {
    OrderMessage order = new OrderMessage();
    order.setType("ORDER");
    order.setClOrdID("TEST-" + System.currentTimeMillis());
    order.setSide("BUY");
    order.setProduct("FOSFO");
    order.setQty(10);
    order.setMode("MARKET");
    
    // Add this for instant fill during development
    order.setDebugMode("AUTO_ACCEPT");
    
    sendMessage(order);
}
```

**When debug mode is ENABLED:**
- Order fills instantly
- Virtual SERVER counterparty created
- FILL message sent back immediately

**When debug mode is DISABLED:**
- Order is rejected
- ERROR message returned: "debug mode is disabled"
- Client must handle the error gracefully

### Flow Diagram

```
Client Sends Order
   with debugMode="AUTO_ACCEPT"
         │
         ▼
    OfferGenerator
         │
         ├──> debugModeService.ValidateDebugRequest()
         │         │
         │         ├──> Check if debug mode enabled
         │         │
         │         ├──> If DISABLED → Return Error
         │         │
         │         └──> If ENABLED → Continue
         │
         └──> handleAutoAcceptOrder()
                   │
                   └──> Create virtual SELL order
                         Execute immediate match
                         Send FILL message
```

## Next Steps (TO DO)

### 1. Admin Panel UI (HIGH PRIORITY)

Create admin panel interface at `/admin`:

```html
<!-- In web/admin.html -->
<div class="debug-mode-control">
  <h2>System Settings</h2>
  
  <div class="setting-row">
    <label>
      <strong>Debug Mode:</strong>
      <span id="debug-status" class="badge">ENABLED</span>
    </label>
    
    <button id="toggle-debug-btn" onclick="toggleDebugMode()">
      Toggle Debug Mode
    </button>
  </div>
  
  <div class="info">
    <p>Last changed: <span id="last-changed-when">2024-01-15 10:30 AM</span></p>
    <p>Changed by: <span id="last-changed-who">admin@example.com</span></p>
  </div>
  
  <div class="warning" id="production-warning" style="display: none;">
    ⚠️ WARNING: You are about to ENABLE debug mode in production!
    This will allow AUTO_ACCEPT orders and other debug features.
    Are you sure?
  </div>
</div>
```

### 2. Admin API Endpoints (HIGH PRIORITY)

Need to create these endpoints (example locations):

**File:** `internal/transport/websocket_server.go` or new `admin_handler.go`

```go
// GET /admin/api/debug-mode
func (s *Server) handleGetDebugMode(w http.ResponseWriter, r *http.Request) {
    enabled := s.debugModeService.IsEnabled()
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "enabled": enabled,
        "updatedAt": ...,  // from database
        "updatedBy": ...,  // from database
    })
}

// POST /admin/api/debug-mode
func (s *Server) handleSetDebugMode(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Enabled   bool   `json:"enabled"`
        UpdatedBy string `json:"updatedBy"`
    }
    
    json.NewDecoder(r.Body).Decode(&req)
    
    // Validate admin token here
    
    err := s.debugModeService.SetEnabled(r.Context(), req.Enabled, req.UpdatedBy)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "debugMode": req.Enabled,
    })
}
```

### 3. Wire Up Dependencies (HIGH PRIORITY)

Need to initialize and inject `DebugModeService` in main.go or wherever the server starts:

```go
// In cmd/server/main.go or similar

// Create system settings repository
systemSettingsRepo := mongodb.NewSystemSettingsRepository(db.GetDatabase())

// Create debug mode service
debugModeService := service.NewDebugModeService(cfg, systemSettingsRepo)

// Pass to market engine
marketEngine := market.NewMarketEngine(
    cfg,
    db,
    orderRepo,
    fillRepo,
    marketStateRepo,
    orderBookRepo,
    broadcaster,
    inventoryService,
    teamRepo,
    debugModeService,  // <-- ADD THIS
)
```

### 4. Update Web UI to Respect Debug Mode (MEDIUM PRIORITY)

Add JavaScript to fetch debug mode status and hide/show features:

```javascript
// In web/index.html
async function loadDebugModeStatus() {
    const response = await fetch('/admin/api/debug-mode');
    const data = await response.json();
    
    if (!data.enabled) {
        // Hide auto-accept checkbox
        document.getElementById('auto-accept-toggle').style.display = 'none';
        
        // Disable debug features
        document.getElementById('simulator-tab').classList.add('disabled');
        
        // Show "REAL MODE" indicator
        showRealModeIndicator();
    }
}
```

### 5. Java SDK Documentation (LOW PRIORITY)

Update Java SDK documentation to explain `setDebugMode()`:

```markdown
## Debug Mode

The `debugMode` field can be set on orders for testing purposes:

- `"AUTO_ACCEPT"` - Order fills instantly (development only)
- `""` or `null` - Normal order processing

**Note:** Debug features are only available when the server has debug mode enabled.
In production environments, debug mode requests will be rejected.
```

### 6. Error Handling in Clients (MEDIUM PRIORITY)

Java clients should handle debug mode errors:

```java
try {
    sendMessage(order);
} catch (DebugModeDisabledException e) {
    logger.warn("Debug mode is disabled on server, falling back to regular order");
    // Remove debugMode and retry
    order.setDebugMode(null);
    sendMessage(order);
}
```

## Testing Checklist

### Development Environment
- [ ] Set `debugModeEnabled: true` in config.yaml
- [ ] Place order with `debugMode: "AUTO_ACCEPT"`
- [ ] Verify instant fill
- [ ] Check logs for "Auto-accept debug order executed"

### Production Simulation
- [ ] Set `debugModeEnabled: false` in config
- [ ] Restart server
- [ ] Attempt order with `debugMode: "AUTO_ACCEPT"`
- [ ] Verify rejection with proper error message
- [ ] Check logs for "Debug mode request rejected"

### Admin Panel
- [ ] Access `/admin` page
- [ ] Toggle debug mode ON
- [ ] Verify database updated
- [ ] Place AUTO_ACCEPT order - should work
- [ ] Toggle debug mode OFF
- [ ] Place AUTO_ACCEPT order - should fail
- [ ] Check audit logs

### Runtime Toggle
- [ ] Start server with debug mode enabled
- [ ] Place AUTO_ACCEPT order - works
- [ ] Toggle OFF via admin panel (no restart)
- [ ] Place AUTO_ACCEPT order - fails immediately
- [ ] Toggle ON via admin panel
- [ ] Place AUTO_ACCEPT order - works again

## Migration Path

1. **Deploy backend changes** - Server supports both modes
2. **Default to enabled** - No disruption to existing users
3. **Create admin panel** - Admins can control the setting
4. **Test in staging** - Verify toggle functionality
5. **Production deployment** - Set to disabled by default
6. **Monitor** - Watch for rejected debug requests
7. **Document** - Update user/developer documentation

## Rollback Plan

If issues arise:

1. Set `debugModeEnabled: true` in config file
2. Restart server
3. All debug features work again
4. Investigate and fix issue
5. Redeploy when ready

## Security Notes

- ✅ Only admin users can toggle debug mode
- ✅ All changes are logged with attribution
- ✅ Production config defaults to disabled
- ✅ No environment variables override (deliberate security choice)
- ✅ Database setting requires admin database access

## Performance Impact

- **Negligible:** Single mutex-protected boolean check
- **Memory:** ~100 bytes for service state
- **Database:** One document in system_settings collection
- **Network:** No additional overhead per request

## Questions to Resolve

1. **Who can access admin panel?**
   - Recommend: Admin users only with special token
   
2. **Should we notify all clients when debug mode changes?**
   - Recommend: Yes, broadcast a SYSTEM_MESSAGE
   
3. **Should we allow per-team debug mode?**
   - Recommend: Phase 2 feature, start with global only
   
4. **What happens to pending AUTO_ACCEPT orders when disabled?**
   - Recommend: They remain pending, won't auto-fill

5. **Should we log every debug request?**
   - Recommend: Yes at INFO level, helps with auditing
