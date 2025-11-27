#!/bin/bash

echo "🔍 Deployment Monitor - Checking production status..."
echo ""

# Check current production version
echo "📦 Current Production Version:"
PROD_VERSION=$(curl -s https://trading.hellsoft.tech | sed -n '1627p' | grep -oE "const [a-zA-Z0-9_]+ = '[^']+'" | head -1)
echo "  $PROD_VERSION"
echo ""

# Check if version constants are valid JavaScript
echo "🔬 JavaScript Validation:"
INVALID_VAR=$(curl -s https://trading.hellsoft.tech | sed -n '1627,1629p' | grep -E "^[[:space:]]*const [0-9]")
if [ -n "$INVALID_VAR" ]; then
    echo "  ❌ BROKEN - Invalid variable names found"
    echo "  $INVALID_VAR"
else
    echo "  ✅ VALID - No syntax errors detected"
fi
echo ""

# Check latest commit in repo
echo "📋 Expected Version:"
LATEST_COMMIT=$(git log --oneline -1 | awk '{print $1}')
echo "  Commit: $LATEST_COMMIT ($(git log -1 --format=%s))"
echo ""

# Check GitHub Actions status
echo "🚀 GitHub Actions Status:"
BUILD_STATUS=$(curl -s https://api.github.com/repos/HellSoft-Col/stock-market/actions/runs?per_page=1 | grep -E '"status"|"conclusion"' | head -2 | tr -d ' ",')
echo "  $BUILD_STATUS"
echo ""

# Instructions
if echo "$PROD_VERSION" | grep -q "{{VERSION}}"; then
    echo "⏳ Deployment not yet applied - placeholders still present"
    echo "   Wait a few more minutes and run this script again"
elif echo "$PROD_VERSION" | grep -qE "const [0-9a-f]{7} ="; then
    if [ -z "$INVALID_VAR" ]; then
        echo "✅ DEPLOYMENT SUCCESSFUL!"
        echo "   Production is running latest version with valid JavaScript"
        echo ""
        echo "🎯 Next steps:"
        echo "   1. Hard refresh browser: Ctrl+Shift+R (or Cmd+Shift+R on Mac)"
        echo "   2. Open browser console (F12)"
        echo "   3. Click 'Connect' button"
        echo "   4. Verify no JavaScript errors"
    else
        echo "⚠️  OLD BROKEN VERSION STILL DEPLOYED"
        echo "   Wait for current build to complete (~5 more minutes)"
    fi
else
    echo "🔄 Deployment in progress..."
    echo "   Run this script again in 2-3 minutes"
fi
