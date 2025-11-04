#!/bin/bash
# Script to check for latest versions of Maven dependencies

echo "🔍 Checking for latest dependency versions..."
echo ""

check_version() {
    local group=$1
    local artifact=$2
    local current=$3
    
    echo -n "Checking $artifact... "
    
    # Query Maven Central API
    latest=$(curl -s "https://search.maven.org/solrsearch/select?q=g:$group+AND+a:$artifact&rows=1&wt=json" | \
             grep -o '"latestVersion":"[^"]*"' | \
             cut -d'"' -f4)
    
    if [ -z "$latest" ]; then
        echo "❌ Failed to fetch"
        return
    fi
    
    if [ "$current" = "$latest" ]; then
        echo "✅ $current (latest)"
    else
        echo "⚠️  $current → $latest (update available)"
    fi
}

echo "📦 Main Dependencies:"
check_version "com.google.code.gson" "gson" "2.13.1"
check_version "org.projectlombok" "lombok" "1.18.40"
check_version "org.slf4j" "slf4j-api" "2.0.16"

echo ""
echo "🧪 Test Dependencies:"
check_version "org.junit.jupiter" "junit-jupiter" "5.11.4"
check_version "org.mockito" "mockito-core" "5.18.0"

echo ""
echo "✨ Check complete!"
