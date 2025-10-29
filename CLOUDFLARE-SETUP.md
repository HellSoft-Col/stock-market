# Cloudflare HTTPS Setup (FREE)

This guide shows how to set up free HTTPS for your trading application using Cloudflare.

## Why Cloudflare?

- ✅ **Free SSL/TLS certificates** (auto-renewed)
- ✅ **Always Use HTTPS** redirect
- ✅ **WebSocket support** for real-time trading
- ✅ **Global CDN** for faster load times
- ✅ **DDoS protection**
- ✅ **HTTP/2 and HTTP/3** support

## Setup Steps

### 1. Add Domain to Cloudflare

1. Go to [cloudflare.com](https://cloudflare.com)
2. Create a free account
3. Click "Add a Site"
4. Enter your domain: `hellsoft.tech`
5. Choose the **Free plan**
6. Cloudflare will scan your existing DNS records

### 2. Update Nameservers

1. Copy the Cloudflare nameservers (e.g., `jane.ns.cloudflare.com`)
2. Go to your domain registrar (where you bought the domain)
3. Update nameservers to use Cloudflare's
4. Wait for DNS propagation (can take up to 24 hours)

### 3. Configure DNS Record

After your container deploys, you'll get an IP address. Add this DNS record:

```
Type: A
Name: trading
Content: [Your Azure Container IP from deployment]
Proxy status: Proxied (orange cloud icon) ✅
TTL: Auto
```

### 4. Configure SSL/TLS Settings

In Cloudflare Dashboard → SSL/TLS:

- **Overview → SSL/TLS encryption mode**: `Flexible`
  - This allows HTTPS on the client side, HTTP to your container
- **Edge Certificates → Always Use HTTPS**: `On` ✅
- **Edge Certificates → Minimum TLS Version**: `TLS 1.2`

### 5. Optional: Page Rules (Free plan includes 3)

Create a page rule for WebSocket support:
- **URL pattern**: `trading.hellsoft.tech/ws*`
- **Settings**: 
  - Cache Level: Bypass
  - Disable Apps
  - Disable Performance

## Testing Your Setup

1. **HTTP Redirect**: Visit `http://trading.hellsoft.tech` - should redirect to HTTPS
2. **HTTPS Access**: Visit `https://trading.hellsoft.tech` - should load your app
3. **SSL Certificate**: Click the lock icon - should show Cloudflare certificate
4. **WebSocket**: Test trading functionality - should work over WSS

## Troubleshooting

### SSL/TLS Issues
- Ensure **Flexible** mode (not Full or Strict)
- Check **Always Use HTTPS** is enabled
- Wait for certificate provisioning (can take 15 minutes)

### WebSocket Issues
- Verify page rule for `/ws*` path
- Check that Cloudflare supports WebSockets (it does on free plan)
- Test direct container access first: `http://[container-fqdn]:8080`

### DNS Issues
- Verify A record points to correct container IP
- Ensure **Proxied** (orange cloud) is enabled
- Check DNS propagation: `dig trading.hellsoft.tech`

## Architecture

```
User Browser
    ↓ HTTPS/WSS
Cloudflare (Free SSL)
    ↓ HTTP/WS  
Azure Container Instance (Port 8080)
    ↓
Your Trading Application
```

## Benefits of This Setup

1. **Zero SSL costs** - Cloudflare handles all certificates
2. **Automatic renewals** - No certificate expiration worries
3. **Performance boost** - Global CDN caches static assets
4. **Security** - DDoS protection and Web Application Firewall
5. **Analytics** - Free traffic analytics in Cloudflare dashboard

## Next Steps

After setup is complete:
1. Test all functionality
2. Monitor in Cloudflare Analytics
3. Consider upgrading to Pro plan for additional features (optional)