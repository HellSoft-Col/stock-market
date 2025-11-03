# Admin Dashboard Guide

## Overview
The admin dashboard provides comprehensive monitoring and control over the stock exchange server with a distinctive red theme.

## Access

### Login
1. Navigate to: `http://localhost:9000`
2. Enter admin token (any token that returns `team: "admin"` from LOGIN_OK)
3. Automatically redirected to `/admin.html`

### Security
- Only users with `teamName === "admin"` can access
- Regular users are automatically redirected back to main page
- Admin token stored in localStorage for session persistence

## Dashboard Sections

### 📊 Statistics Overview (Top Cards)
- **Active Users**: Real-time count of connected sessions
- **Active Orders**: Number of pending orders in the system
- **Total Volume**: Cumulative trading volume ($ amount)
- **Total Trades**: Number of executed trades

### 👥 Connected Users Panel
- Shows all connected sessions grouped by team
- Displays client type (Web Browser / Java/Native Client)
- Shows remote address for each connection
- Refresh button for manual updates
- Auto-refreshes every 5 seconds

### 📈 Team Performance Panel
- P&L (Profit & Loss) for each team
- Color-coded: Green for profit, Red for loss
- Total number of trades per team
- Sortable by performance
- Real-time updates from performance reports

### 📋 All Active Orders Table
- Complete order book view
- Columns: Team, Side, Product, Quantity, Price, Mode
- Color-coded sides: Green for BUY, Red for SELL
- Sortable and filterable
- Shows MARKET orders separately from LIMIT orders

### 🎮 Admin Controls
Four main action buttons:

1. **Create Order**
   - Place orders on behalf of any team
   - Select team from dropdown (populated with connected teams)
   - Choose side (BUY/SELL)
   - Select product (FOSFO, GUACA, SEBO, PALTA-OIL, PITA)
   - Set quantity
   - Choose mode (MARKET/LIMIT)
   - Set limit price if applicable
   - Orders prefixed with "ADMIN-" in clOrdID

2. **Cancel All Orders**
   - Cancels ALL pending orders in the system
   - Requires confirmation dialog
   - Use with caution!

3. **Broadcast Message**
   - Send announcement to all connected users
   - Appears as toast notification
   - Useful for maintenance warnings

4. **Export Data**
   - Download system data
   - Includes orders, trades, performance metrics
   - CSV/JSON format options

### 📊 Market Overview Panel
- Real-time ticker data
- Best bid/ask prices
- Mid prices for all products
- Volume indicators
- Market state visualization

## Features

### Auto-Refresh
- Automatically refreshes all data every 5 seconds
- Includes: sessions, orders, performance, market data
- Manual refresh buttons available on each panel

### Real-Time Updates
- WebSocket connection for live data
- Status indicator shows connection state
- Auto-reconnect on disconnection
- Toast notifications for important events

### Responsive Design
- Works on desktop, tablet, mobile
- Grid layout adapts to screen size
- Scrollable sections for long lists
- Modal dialogs for actions

## Color Scheme

### Admin Red Theme
- Primary: `#dc2626` (red-600)
- Hover: `#991b1b` (red-900)
- Background gradient: red-600 to red-900
- Accent colors maintain red palette
- Contrasts with green theme of regular UI

### Status Colors
- **Connected**: Green dot
- **Disconnected**: Red dot
- **BUY orders**: Green badge
- **SELL orders**: Red badge
- **Profit**: Green text
- **Loss**: Red text

## WebSocket Messages

### Sent by Admin
- `LOGIN` - Authenticate as admin
- `REQUEST_CONNECTED_SESSIONS` - Get all users
- `REQUEST_ALL_ORDERS` - Get all orders
- `REQUEST_PERFORMANCE_REPORT` (scope: "global") - Get all team P&L
- `ORDER` - Create admin order

### Received by Admin
- `LOGIN_OK` - Authentication success
- `CONNECTED_SESSIONS` - List of all sessions
- `ALL_ORDERS` - Complete order book
- `GLOBAL_PERFORMANCE_REPORT` - All team performance
- `TICKER` - Market data updates
- `ERROR` - Error messages

## Usage Examples

### Monitor Trading Activity
1. Watch "Active Orders" panel for new orders
2. Check "Team Performance" for P&L changes
3. Monitor "Total Trades" stat for activity

### Create Emergency Order
1. Click "Create Order" button
2. Select affected team
3. Choose appropriate side/product/quantity
4. Submit to execute immediately

### Handle Server Maintenance
1. Click "Broadcast Message"
2. Enter: "Server maintenance in 5 minutes"
3. All users receive notification
4. Use "Cancel All Orders" if needed
5. Users can reconnect after maintenance

### Export Trading Data
1. Click "Export Data"
2. Select date range
3. Choose format (CSV/JSON)
4. Download for analysis

## Keyboard Shortcuts
- `Ctrl/Cmd + R` - Refresh all panels
- `Esc` - Close modal dialogs
- `Ctrl/Cmd + L` - Logout

## Troubleshooting

### Admin Can't Login
- Verify token returns `team: "admin"` in LOGIN_OK
- Check console for errors
- Verify WebSocket connection
- Clear localStorage and retry

### Data Not Updating
- Check connection status indicator
- Click manual refresh buttons
- Check browser console for errors
- Verify server is running

### Performance Issues
- Reduce auto-refresh interval
- Close unused browser tabs
- Check server logs for errors
- Monitor MongoDB performance

## Best Practices

1. **Monitor Regularly**: Check dashboard during active trading hours
2. **Use Admin Orders Sparingly**: Only for testing or emergencies
3. **Broadcast Before Actions**: Warn users before canceling orders
4. **Export Data Frequently**: Regular backups of trading data
5. **Keep Connection Stable**: Watch for disconnection alerts

## Development

### Customization
- Edit `/web/admin.html` for UI changes
- Modify WebSocket handlers for new features
- Add new panels by extending grid layout
- Customize colors in Tailwind config

### Adding Features
1. Add button to Admin Controls section
2. Implement handler function
3. Add WebSocket message type
4. Update backend if needed
5. Test thoroughly before deployment

## Security Considerations

- Admin dashboard has full system access
- Protect admin tokens carefully
- Consider IP whitelisting for admin access
- Log all admin actions for audit trail
- Implement rate limiting for admin operations
- Use HTTPS in production

## Support

For issues or feature requests:
- Check server logs: `/tmp/server-admin.log`
- Review browser console for errors
- Contact system administrator
- Report bugs via issue tracker
