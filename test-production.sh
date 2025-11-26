#!/bin/bash

echo "=================================="
echo "Testing Production Deployment"
echo "=================================="
echo ""

# Test 1: Check if site is accessible
echo "1. Testing site accessibility..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" https://trading.hellsoft.tech)
if [ "$HTTP_CODE" = "200" ]; then
    echo "   ✅ Site is accessible (HTTP $HTTP_CODE)"
else
    echo "   ❌ Site returned HTTP $HTTP_CODE"
fi
echo ""

# Test 2: Check for JavaScript syntax errors
echo "2. Checking for JavaScript syntax errors..."
SYNTAX_ERROR=$(curl -s https://trading.hellsoft.tech | grep -c "Unexpected token")
if [ "$SYNTAX_ERROR" -eq 0 ]; then
    echo "   ✅ No syntax errors detected"
else
    echo "   ❌ Syntax errors found: $SYNTAX_ERROR"
fi
echo ""

# Test 3: Check if toggleConnection function exists
echo "3. Checking if toggleConnection is defined..."
HAS_TOGGLE=$(curl -s https://trading.hellsoft.tech | grep -c "function toggleConnection")
if [ "$HAS_TOGGLE" -gt 0 ]; then
    echo "   ✅ toggleConnection function found"
else
    echo "   ❌ toggleConnection function not found"
fi
echo ""

# Test 4: Check script closing brackets
echo "4. Checking script structure..."
curl -s https://trading.hellsoft.tech > /tmp/prod_index.html
OPEN_SCRIPT=$(grep -c "<script>" /tmp/prod_index.html)
CLOSE_SCRIPT=$(grep -c "</script>" /tmp/prod_index.html)
echo "   Opening <script> tags: $OPEN_SCRIPT"
echo "   Closing </script> tags: $CLOSE_SCRIPT"
if [ "$OPEN_SCRIPT" -eq "$CLOSE_SCRIPT" ]; then
    echo "   ✅ Script tags balanced"
else
    echo "   ❌ Script tags unbalanced"
fi
echo ""

# Test 5: Check last modified date
echo "5. Checking deployment timestamp..."
LAST_MODIFIED=$(curl -s -I https://trading.hellsoft.tech | grep -i "last-modified" | cut -d' ' -f2-)
echo "   Last Modified: $LAST_MODIFIED"
echo ""

echo "=================================="
echo "Test Summary"
echo "=================================="
echo ""
echo "If all tests pass:"
echo "  → Open https://trading.hellsoft.tech in your browser"
echo "  → Press Ctrl+Shift+R (or Cmd+Shift+R on Mac) to hard refresh"
echo "  → Click the Connect button"
echo "  → It should work without errors!"
echo ""
