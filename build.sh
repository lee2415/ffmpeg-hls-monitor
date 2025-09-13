#!/bin/bash

# MultiView Monitor Cross-Platform Build Script
# Builds static binaries for multiple platforms

set -e

APP_NAME="multiview-monitor"
VERSION="1.0.0"
BUILD_DIR="dist"
LDFLAGS="-w -s -X main.version=$VERSION"

echo "Building MultiView Monitor v$VERSION"

# Clean previous builds
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

# Build function
build() {
    local os=$1
    local arch=$2
    local ext=$3
    
    echo "Building for $os/$arch..."
    
    local output="$BUILD_DIR/${APP_NAME}_${VERSION}_${os}_${arch}${ext}"
    
    env CGO_ENABLED=0 GOOS=$os GOARCH=$arch \
        go build -a -ldflags "$LDFLAGS -extldflags '-static'" \
        -o "$output" ./cmd/monitor
    
    if [ $? -eq 0 ]; then
        echo "✓ Built: $output"
        ls -lh "$output"
    else
        echo "✗ Failed to build for $os/$arch"
        exit 1
    fi
}

# Build for multiple platforms
echo "Building for Linux..."
build "linux" "amd64" ""
build "linux" "arm64" ""

echo "Building for macOS..."
build "darwin" "amd64" ""
build "darwin" "arm64" ""

echo "Building for Windows..."
build "windows" "amd64" ".exe"

echo ""
echo "Build completed! Binaries created in $BUILD_DIR:"
ls -la $BUILD_DIR

echo ""
echo "Binary sizes:"
du -h $BUILD_DIR/*

echo ""
echo "To test a binary:"
echo "  ./dist/${APP_NAME}_${VERSION}_linux_amd64 --help"
echo ""
echo "To create release archives:"
echo "  ./package.sh"