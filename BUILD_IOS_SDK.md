# Netmie iOS SDK Build Guide

## Overview

This document explains how to build the Netmie iOS SDK (NetBird fork with v2ray support) as an xcframework for use in iOS applications.

## Quick Start

```bash
cd /path/to/netmie
./build-ios-sdk.sh
```

The script will:
1. Install gomobile and gobind
2. Initialize gomobile
3. Build the xcframework
4. Verify the output

## Critical Issues & Solutions

### Issue 1: `cd` Command Doesn't Work

**Problem**: Using `cd` in bash scripts fails due to shell hook conflicts (gvm, nvm, etc.)

**Solution**: Use `pushd` and `popd` instead:
```bash
# ❌ WRONG - cd doesn't work
cd /path/to/netmie
gomobile bind ...

# ✅ CORRECT - use pushd/popd
pushd /path/to/netmie
gomobile bind ...
popd
```

### Issue 2: "no exported names in the package"

**Problem**: gomobile reports no exported symbols in the package

**Solution**: Set `CGO_ENABLED=0`:
```bash
# ❌ WRONG - missing CGO_ENABLED
gomobile bind -target=ios ...

# ✅ CORRECT - set CGO_ENABLED=0
CGO_ENABLED=0 gomobile bind -target=ios ...
```

### Issue 3: Wrong gomobile Version

**Problem**: Using sagernet's gomobile fork instead of the official one

**Solution**: Install from `golang.org/x/mobile`:
```bash
# ✅ CORRECT - use official gomobile
go install golang.org/x/mobile/cmd/gomobile@v0.0.0-20251113184115-a159579294ab
go install golang.org/x/mobile/cmd/gobind@v0.0.0-20251113184115-a159579294ab
```

### Issue 4: Build Tags Not Recognized

**Problem**: The package uses `//go:build ios` tags, causing issues when analyzing

**Solution**: gomobile automatically handles build tags - just ensure you're in the correct directory and using relative paths:
```bash
# ✅ CORRECT - relative path from project root
gomobile bind -target=ios ./client/ios/NetBirdSDK
```

## Manual Build Steps

If you need to build manually without the script:

```bash
# 1. Navigate to netmie directory
cd /path/to/netmie

# 2. Install gomobile (if not already installed)
go install golang.org/x/mobile/cmd/gomobile@v0.0.0-20251113184115-a159579294ab
go install golang.org/x/mobile/cmd/gobind@v0.0.0-20251113184115-a159579294ab

# 3. Initialize gomobile
gomobile init

# 4. Build the framework (MUST use pushd, not cd!)
pushd /path/to/netmie
CGO_ENABLED=0 \
PATH=$PATH:$(go env GOPATH)/bin \
gomobile bind \
    -target=ios \
    -bundleid=io.netmie.framework \
    -ldflags="-X github.com/netbirdio/netbird/version.version=0.1.0" \
    -o ./NetmieSDK.xcframework \
    ./client/ios/NetBirdSDK
popd
```

## Verification

After building, verify the output:

```bash
ls -lh NetmieSDK.xcframework/
```

You should see:
- `Info.plist`
- `ios-arm64/` (device architecture)
- `ios-arm64_x86_64-simulator/` (simulator architecture)

## Integration with Xcode

1. Copy the framework to your project:
   ```bash
   cp -R NetmieSDK.xcframework /path/to/YourApp/Frameworks/
   ```

2. In Xcode:
   - Select your target → General → Frameworks, Libraries, and Embedded Content
   - Click "+" and add `NetmieSDK.xcframework`
   - For the main app target: Set to "Embed & Sign"
   - For extension targets: Set to "Do Not Embed"

3. Update imports in Swift code:
   ```swift
   import NetmieSDK  // Changed from NetBirdSDK
   ```

## Troubleshooting

### Build fails with "unable to import bind"

This usually means gomobile wasn't initialized properly:
```bash
gomobile init
```

### Build fails with "no exported names"

Make sure you're setting `CGO_ENABLED=0` and using the correct directory.

### Framework is missing architectures

Ensure you're using `-target=ios` (not `-target=ios/arm64` or similar). The `ios` target automatically builds for both device and simulator.

## CI/CD Integration

For GitHub Actions or other CI systems, see `.github/workflows/mobile-build-validation.yml` for the reference implementation.

Key points:
- Use `macos-latest` runner for iOS builds
- Install Go from `go.mod` version
- Set `CGO_ENABLED=0`
- Use relative paths

## Version Information

- Go version: 1.24+ (specified in go.mod)
- gomobile version: v0.0.0-20251113184115-a159579294ab
- Target platforms: iOS (arm64 device + arm64/x86_64 simulator)

## References

- Official gomobile: https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile
- NetBird iOS client: https://github.com/netbirdio/netbird/tree/main/client/ios
- Netmie (fork): https://github.com/yourusername/netmie
