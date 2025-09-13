#!/bin/bash

# MultiView Monitor Release Packaging Script
# Creates release archives for distribution

set -e

APP_NAME="multiview-monitor"
VERSION="1.0.0"
BUILD_DIR="dist"
RELEASE_DIR="releases"

echo "Packaging MultiView Monitor v$VERSION for release"

# Clean and create release directory
rm -rf $RELEASE_DIR
mkdir -p $RELEASE_DIR

# Package function
package() {
    local os=$1
    local arch=$2
    local ext=$3
    
    local binary_name="${APP_NAME}_${VERSION}_${os}_${arch}${ext}"
    local package_name="${APP_NAME}_${VERSION}_${os}_${arch}"
    
    if [ ! -f "$BUILD_DIR/$binary_name" ]; then
        echo "Binary not found: $BUILD_DIR/$binary_name"
        return 1
    fi
    
    echo "Packaging for $os/$arch..."
    
    # Create temporary directory
    local temp_dir=$(mktemp -d)
    local package_dir="$temp_dir/$package_name"
    mkdir -p "$package_dir"
    
    # Copy binary
    cp "$BUILD_DIR/$binary_name" "$package_dir/${APP_NAME}${ext}"
    
    # Copy documentation
    cp README.md "$package_dir/"
    cp CLAUDE.md "$package_dir/"
    
    # Copy example configs
    mkdir -p "$package_dir/examples"
    cp examples/*.yaml "$package_dir/examples/" 2>/dev/null || true
    cp configs/multiview-monitor.yaml "$package_dir/examples/" 2>/dev/null || true
    
    # Create install script for Unix systems
    if [ "$os" != "windows" ]; then
        cat > "$package_dir/install.sh" << 'EOF'
#!/bin/bash
echo "Installing MultiView Monitor..."
sudo cp multiview-monitor /usr/local/bin/
sudo chmod +x /usr/local/bin/multiview-monitor
echo "Installation complete! Run: multiview-monitor --help"
EOF
        chmod +x "$package_dir/install.sh"
    fi
    
    # Create archive
    if [ "$os" = "windows" ]; then
        # ZIP for Windows
        (cd "$temp_dir" && zip -r "$package_name.zip" "$package_name")
        mv "$temp_dir/$package_name.zip" "$RELEASE_DIR/"
        echo "✓ Created: $RELEASE_DIR/$package_name.zip"
    else
        # TAR.GZ for Unix systems
        (cd "$temp_dir" && tar -czf "$package_name.tar.gz" "$package_name")
        mv "$temp_dir/$package_name.tar.gz" "$RELEASE_DIR/"
        echo "✓ Created: $RELEASE_DIR/$package_name.tar.gz"
    fi
    
    # Cleanup
    rm -rf "$temp_dir"
}

# Check if binaries exist
if [ ! -d "$BUILD_DIR" ] || [ -z "$(ls -A $BUILD_DIR)" ]; then
    echo "No binaries found in $BUILD_DIR. Run ./build.sh first."
    exit 1
fi

# Package all platforms
package "linux" "amd64" ""
package "linux" "arm64" ""
package "darwin" "amd64" ""
package "darwin" "arm64" ""
package "windows" "amd64" ".exe"

echo ""
echo "Release packaging completed!"
echo "Archives created in $RELEASE_DIR:"
ls -la $RELEASE_DIR

echo ""
echo "Archive sizes:"
du -h $RELEASE_DIR/*

echo ""
echo "To test an archive:"
echo "  cd /tmp"
echo "  tar -xzf $(pwd)/$RELEASE_DIR/${APP_NAME}_${VERSION}_linux_amd64.tar.gz"
echo "  cd ${APP_NAME}_${VERSION}_linux_amd64"
echo "  ./multiview-monitor --help"