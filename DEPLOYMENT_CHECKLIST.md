# Deployment Checklist - Debug Mode System

## 🚀 Pre-Deployment (Done)
- ✅ Code committed and pushed to main
- ✅ Commit: `aefe19f` - "Add global debug mode system with runtime toggle and admin panel controls"
- ✅ 15 files changed, 1586 insertions(+), 50 deletions(-)

## ⏱️ During Deployment (Wait ~10 minutes)

### Expected Build Process
1. GitHub Actions or CI/CD pulls latest code
2. Go builds the server binary
3. Docker container is built (if using containers)
4. Server is deployed to production
5. Server restarts with new code

### What to Watch For
```bash
# Expected log output on server startup:
{"level":"info","message":"Starting Intergalactic Avocado Stock Exchange Server"}
{"level":"info","message":"Database connection established"}
{"level":"info","message":"Debug mode service initialized","debugModeEnabled":false}
{"level":"info","message":"Loaded debug mode setting from database","debugMode":false}
{"level":"info","message":"Market engine started"}
{"level":"info","message":"WebSocket server started"}
```

## ✅ Post-Deployment Testing

### 1. Verify Server is Running
```bash
# Check server health
curl https://your-production-url.com/

# Should return 200 OK
```

### 2. Check Debug Mode Status via API
```bash
# Get current debug mode status
curl https://your-production-url.com/admin/api/debug-mode

# Expected response:
# {"enabled":false}
```

### 3. Test Admin Panel
1. Open: `https://your-production-url.com/admin.html`
2. Login with admin token
3. Look for "Debug Mode Control" section at the top
4. Status should show: **"REAL MODE"** in green ✅
5. Button should say: **"Enable Debug Mode"** in yellow

### 4. Test Java Client (Debug Disabled)
```java
// This should be REJECTED since debug mode is disabled
OrderMessage order = new OrderMessage();
order.setDebugMode("AUTO_ACCEPT");
client.sendMessage(order);

// Expected: ERROR message
// {
//   "type": "ERROR",
//   "error": "debug mode is disabled - this operation is not allowed in production mode"
// }
```

### 5. Test Web UI
1. Open: `https://your-production-url.com/`
2. Connect and login
3. Go to Trading tab
4. **The "Auto-Accept" checkbox should be visible**
   - (Future enhancement: hide when disabled)
5. Check the checkbox and place an order
6. **Expected:** Order is rejected with error

### 6. Test Debug Mode Toggle (OPTIONAL - Use with Caution)
1. In admin panel, click **"Enable Debug Mode"**
2. Confirm the warning dialog
3. Enter your name/email (e.g., "admin@example.com")
4. Status should change to **"DEBUG MODE"** in yellow
5. Try Java client AUTO_ACCEPT order again
6. **Should work now** - order fills instantly
7. **IMPORTANT:** Toggle back to disabled!
8. Click **"Disable Debug Mode"**
9. Verify status is back to **"REAL MODE"**

### 7. Verify Database Persistence
```javascript
// In MongoDB shell or client
use avocado_exchange_prod

// Check the system_settings collection
db.system_settings.findOne({_id: "debugModeEnabled"})

// Expected output:
{
  "_id": "debugModeEnabled",
  "value": false,
  "updatedAt": ISODate("2024-..."),
  "updatedBy": "admin@example.com"  // or empty if never changed
}
```

### 8. Check Server Logs
Look for these log entries:
```json
// On startup
{"level":"info","debugModeEnabled":false,"message":"Debug mode service initialized"}

// If someone toggles debug mode
{"level":"info","debugMode":true,"updatedBy":"admin@example.com","message":"Debug mode setting updated"}

// If debug request is rejected
{"level":"warn","clOrdID":"TEST-123","debugMode":"AUTO_ACCEPT","message":"Debug mode request rejected - debug mode is disabled"}
```

## 🐛 Troubleshooting

### Issue: Admin panel doesn't show debug mode section
**Fix:**
- Hard refresh browser (Ctrl+F5 or Cmd+Shift+R)
- Check browser console for JavaScript errors
- Verify admin.html was deployed correctly

### Issue: API returns 404 for /admin/api/debug-mode
**Fix:**
- Check server logs for startup errors
- Verify WebSocketServer routes were registered
- Check that server is running the new binary

### Issue: Debug mode is enabled in production
**Fix:**
1. Immediately toggle OFF via admin panel
2. Check `config.production.yaml` has `debugModeEnabled: false`
3. Verify correct config file is being used
4. Check database setting: `db.system_settings.findOne({_id: "debugModeEnabled"})`

### Issue: AUTO_ACCEPT orders still work when disabled
**Fix:**
1. Check server logs - should see "Debug mode request rejected"
2. Verify OfferGenerator is validating requests
3. Check debugModeService.IsEnabled() returns false
4. Restart server if needed

### Issue: Toggle doesn't persist after server restart
**Fix:**
- Check MongoDB connection
- Verify system_settings collection exists
- Check collection write permissions
- Review server logs for database errors

## 📊 Monitoring (After Deployment)

### Metrics to Watch
- [ ] Number of rejected debug requests (should be > 0 if clients are trying)
- [ ] Debug mode toggle events (should be rare/zero in prod)
- [ ] ERROR logs related to debug mode
- [ ] Server uptime and restarts

### Alerts to Set Up
- Alert if debug mode is enabled for > 30 minutes
- Alert on multiple failed toggle attempts
- Alert on spike in rejected debug requests

## 🎉 Success Criteria

- ✅ Server deployed and running
- ✅ Debug mode status API returns `{"enabled":false}`
- ✅ Admin panel shows "REAL MODE" in green
- ✅ AUTO_ACCEPT orders are rejected
- ✅ Toggle button works (if tested)
- ✅ Database persists the setting
- ✅ No errors in server logs

## 📝 Notes for Future

### When to Enable Debug Mode in Production
**Valid reasons:**
- Troubleshooting a critical issue
- Testing a new feature deployment
- Running controlled experiments
- Developer demo/presentation

**Always:**
1. Notify the team before enabling
2. Document reason in audit trail
3. Monitor closely while enabled
4. Disable immediately after done
5. Maximum 30 minutes enabled

### Regular Maintenance
- Weekly: Review debug mode change history
- Monthly: Check for abandoned debug toggles
- Quarterly: Review who has admin access

---

## Current Time: Start Timer Now! ⏰

**Deployment started at:** ___________  
**Expected completion:** ___________ (+ 10 minutes)  
**Actual completion:** ___________

While waiting, you can:
- ☕ Get coffee
- 📖 Review DEBUG_MODE.md documentation
- 🧪 Prepare test data for Java client
- 📝 Plan next features

**Next actions after deployment:**
1. Run through this checklist
2. Test the admin panel toggle
3. Verify Java client behavior
4. Check server logs
5. Celebrate! 🎉
