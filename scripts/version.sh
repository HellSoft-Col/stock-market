#!/bin/bash

# Version management script for Java WebSocket Client

set -e

VERSION_FILE="sdk/java/websocket-client/VERSION"

# Function to display current version
show_version() {
    if [ -f "$VERSION_FILE" ]; then
        echo "Current version: $(cat $VERSION_FILE)"
    else
        echo "No VERSION file found. Default would be: 1.0.0-SNAPSHOT"
    fi
}

# Function to set version
set_version() {
    local new_version=$1
    if [ -z "$new_version" ]; then
        echo "Error: Version number required"
        echo "Usage: $0 set <version>"
        exit 1
    fi
    
    # Validate version format (basic semver)
    if [[ ! $new_version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "Error: Version must be in format X.Y.Z"
        exit 1
    fi
    
    echo "$new_version" > "$VERSION_FILE"
    echo "Version set to: $new_version"
}

# Function to increment version
increment_version() {
    local part=$1
    
    if [ ! -f "$VERSION_FILE" ]; then
        echo "1.0.0" > "$VERSION_FILE"
    fi
    
    local current=$(cat "$VERSION_FILE")
    local major=$(echo $current | cut -d. -f1)
    local minor=$(echo $current | cut -d. -f2)
    local patch=$(echo $current | cut -d. -f3)
    
    case $part in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        patch|*)
            patch=$((patch + 1))
            ;;
    esac
    
    local new_version="$major.$minor.$patch"
    echo "$new_version" > "$VERSION_FILE"
    echo "Version incremented to: $new_version"
}

# Main script logic
case "${1:-}" in
    show)
        show_version
        ;;
    set)
        set_version "$2"
        ;;
    increment|inc)
        increment_version "${2:-patch}"
        ;;
    *)
        echo "Usage: $0 {show|set <version>|increment [major|minor|patch]}"
        echo ""
        echo "Examples:"
        echo "  $0 show                    # Show current version"
        echo "  $0 set 1.2.3              # Set specific version"
        echo "  $0 increment patch        # Increment patch version (default)"
        echo "  $0 increment minor        # Increment minor version"
        echo "  $0 increment major        # Increment major version"
        exit 1
        ;;
esac