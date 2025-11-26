#!/bin/bash

# Get version information
COMMIT_HASH=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
COMMIT_SHORT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
COMMIT_MSG=$(git log -1 --pretty=%B 2>/dev/null | head -n1 || echo "unknown")
BUILD_DATE=$(date -u +"%Y-%m-%d %H:%M:%S UTC")

echo "Injecting version info into HTML files..."
echo "  Commit: $COMMIT_SHORT ($COMMIT_HASH)"
echo "  Message: $COMMIT_MSG"
echo "  Date: $BUILD_DATE"

# Replace placeholders in index.html
sed -i.bak \
    -e "s/BUILD_VERSION/$COMMIT_SHORT/g" \
    -e "s/BUILD_COMMIT/$COMMIT_HASH/g" \
    -e "s/BUILD_DATE/$BUILD_DATE/g" \
    web/index.html

# Also update the version display placeholder
sed -i.bak \
    -e "s/<span id=\"version-info\" class=\"font-mono\">BUILD_VERSION<\/span>/<span id=\"version-info\" class=\"font-mono\" title=\"$COMMIT_MSG\">$COMMIT_SHORT<\/span>/g" \
    web/index.html

# Clean up backup files
rm -f web/index.html.bak

echo "✅ Version info injected successfully"
echo "   Short hash: $COMMIT_SHORT will be displayed in UI"
