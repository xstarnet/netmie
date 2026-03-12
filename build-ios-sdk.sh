#!/bin/bash

# Netmie iOS SDK Build Script
# This script compiles the Netmie SDK (NetBird fork with v2ray support) as an iOS xcframework
#
# CRITICAL NOTES:
# 1. Must use pushd/popd instead of cd - the cd command has issues with shell hooks
# 2. Must set CGO_ENABLED=0 for gomobile to work correctly
# 3. Must ensure gomobile and gobind are installed from golang.org/x/mobile (not sagernet)
# 4. The package has //go:build ios tags, so gomobile handles the build constraints

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Netmie iOS SDK Build Script ===${NC}"
echo ""

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
echo "Script directory: $SCRIPT_DIR"

# Configuration
GOMOBILE_VERSION="v0.0.0-20251113184115-a159579294ab"
OUTPUT_NAME="NetmieSDK.xcframework"
BUNDLE_ID="io.netmie.framework"
VERSION="0.1.0"
PACKAGE_PATH="./client/ios/NetBirdSDK"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

echo "Go version: $(go version)"
echo ""

# Step 1: Install gomobile and gobind
echo -e "${YELLOW}Step 1: Installing gomobile and gobind...${NC}"
go install golang.org/x/mobile/cmd/gomobile@${GOMOBILE_VERSION}
go install golang.org/x/mobile/cmd/gobind@${GOMOBILE_VERSION}

GOPATH=$(go env GOPATH)
GOMOBILE_BIN="$GOPATH/bin/gomobile"

if [ ! -f "$GOMOBILE_BIN" ]; then
    echo -e "${RED}Error: gomobile not found at $GOMOBILE_BIN${NC}"
    exit 1
fi

echo -e "${GREEN}✓ gomobile installed at $GOMOBILE_BIN${NC}"
echo ""

# Step 2: Initialize gomobile
echo -e "${YELLOW}Step 2: Initializing gomobile...${NC}"
$GOMOBILE_BIN init
echo -e "${GREEN}✓ gomobile initialized${NC}"
echo ""

# Step 3: Clean previous build
if [ -d "$SCRIPT_DIR/$OUTPUT_NAME" ]; then
    echo -e "${YELLOW}Step 3: Cleaning previous build...${NC}"
    rm -rf "$SCRIPT_DIR/$OUTPUT_NAME"
    echo -e "${GREEN}✓ Previous build cleaned${NC}"
else
    echo -e "${YELLOW}Step 3: No previous build to clean${NC}"
fi
echo ""

# Step 4: Build the xcframework
echo -e "${YELLOW}Step 4: Building iOS xcframework...${NC}"
echo "This may take several minutes..."
echo ""

# CRITICAL: Use pushd instead of cd due to shell hook issues
# CRITICAL: Set CGO_ENABLED=0 for gomobile to work
pushd "$SCRIPT_DIR" > /dev/null

CGO_ENABLED=0 \
PATH="$PATH:$GOPATH/bin" \
$GOMOBILE_BIN bind \
    -target=ios \
    -bundleid="$BUNDLE_ID" \
    -ldflags="-X github.com/netbirdio/netbird/version.version=$VERSION" \
    -o "./$OUTPUT_NAME" \
    "$PACKAGE_PATH"

BUILD_EXIT_CODE=$?
popd > /dev/null

if [ $BUILD_EXIT_CODE -ne 0 ]; then
    echo -e "${RED}Error: Build failed with exit code $BUILD_EXIT_CODE${NC}"
    exit $BUILD_EXIT_CODE
fi

echo ""
echo -e "${GREEN}✓ Build completed successfully!${NC}"
echo ""

# Step 5: Verify the output
if [ -d "$SCRIPT_DIR/$OUTPUT_NAME" ]; then
    echo -e "${GREEN}=== Build Summary ===${NC}"
    echo "Output: $SCRIPT_DIR/$OUTPUT_NAME"
    echo ""
    echo "Contents:"
    ls -lh "$SCRIPT_DIR/$OUTPUT_NAME"
    echo ""

    # Check for required architectures
    if [ -d "$SCRIPT_DIR/$OUTPUT_NAME/ios-arm64" ]; then
        echo -e "${GREEN}✓ ios-arm64 (device) architecture present${NC}"
    else
        echo -e "${RED}✗ ios-arm64 architecture missing${NC}"
    fi

    if [ -d "$SCRIPT_DIR/$OUTPUT_NAME/ios-arm64_x86_64-simulator" ]; then
        echo -e "${GREEN}✓ ios-arm64_x86_64-simulator architecture present${NC}"
    else
        echo -e "${RED}✗ Simulator architecture missing${NC}"
    fi

    echo ""
    echo -e "${GREEN}=== Next Steps ===${NC}"
    echo "1. Copy the framework to your Xcode project:"
    echo "   cp -R $SCRIPT_DIR/$OUTPUT_NAME /path/to/your/project/Frameworks/"
    echo ""
    echo "2. In Xcode:"
    echo "   - Add $OUTPUT_NAME to your target"
    echo "   - Set 'Embed & Sign' for the main app target"
    echo "   - Set 'Do Not Embed' for extension targets"
    echo ""
else
    echo -e "${RED}Error: Output framework not found at $SCRIPT_DIR/$OUTPUT_NAME${NC}"
    exit 1
fi

echo -e "${GREEN}=== Build Complete ===${NC}"
