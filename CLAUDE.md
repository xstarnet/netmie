# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Netmie is a P2P VPN solution based on NetBird with integrated V2Ray proxy functionality. It allows users to:
- Connect to NetBird P2P VPN networks (WireGuard-based)
- Run V2Ray proxy locally (SOCKS5/HTTP endpoints)
- Manage both services through a unified CLI and daemon

The project is written in Go and uses gRPC for CLI-to-daemon communication.

## Build & Development Commands

### Building
```bash
# Build the client binary
go build -o ./netmie ./client

# Install and restart daemon (development)
# This script now automatically stops and removes existing service before reinstalling
./dev-install.sh
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./client/cmd/...

# Run a single test
go test -run TestName ./path/to/package
```

### Linting
```bash
# Lint only changed files (fast, for pre-push)
make lint

# Lint entire codebase (slow, matches CI)
make lint-all

# Install linter locally
make lint-install

# Setup git hooks for automatic pre-push linting
make setup-hooks
```

### Cleanup
```bash
# Clean all processes and residual files
./cleanup.sh
```

## Architecture

### High-Level Design

The system has three main components:

1. **CLI Client** (`client/cmd/`) - Command-line interface for user interaction
2. **Daemon Server** (`client/server/`) - Background service managing both NetBird and V2Ray
3. **V2Ray Engine** (`client/internal/v2ray/`) - V2Ray proxy management

### Communication Flow

```
CLI Commands (vup, vdown, vstatus, vconfig)
    ↓
gRPC over Unix Socket (/var/run/netbird.sock)
    ↓
Daemon Server (runs as root)
    ↓
V2Ray Engine / NetBird Engine
```

### Key Modules

**CLI Commands** (`client/cmd/`)
- `vup.go` - Start V2Ray proxy
- `vdown.go` - Stop V2Ray proxy
- `vstatus.go` - Check V2Ray status
- `vconfig.go` - Download and decrypt V2Ray configuration from HTTP endpoint
- `up.go`, `down.go`, `status.go` - NetBird equivalents
- `root.go` - Root command setup and gRPC connection utilities

**Daemon Server** (`client/server/server.go`)
- Implements gRPC service methods including V2Ray RPC handlers
- `VUp()` - Starts V2Ray engine
- `VDown()` - Stops V2Ray engine
- `VStatus()` - Returns V2Ray status and config version
- `VConfig()` - Updates V2Ray configuration from file path
- Manages both NetBird and V2Ray engines with separate mutexes

**V2Ray Engine** (`client/internal/v2ray/`)
- `engine.go` - Core engine with lifecycle management (Start, Stop, GetStatus)
- `config.go` - Configuration management (load, validate, update)
- `xray_wrapper.go` - Wraps v2ray-core library
- `router.go` - Route configuration
- `tun.go`, `tun_linux.go`, `tun_darwin.go`, `tun_windows.go` - TUN interface support
- `geoip.go`, `geosite.go` - Geo-based routing data

### gRPC Protocol

The daemon exposes a `DaemonService` with V2Ray-specific RPC methods:
- `VUp(VUpRequest) -> VUpResponse` - Start V2Ray
- `VDown(VDownRequest) -> VDownResponse` - Stop V2Ray
- `VStatus(VStatusRequest) -> VStatusResponse` - Get status and config version
- `VConfig(VConfigRequest) -> VConfigResponse` - Update configuration

Protocol definitions are in `client/proto/daemon.proto`.

## Configuration

### V2Ray Configuration
- Default path: `~/.netmie/v2ray-config.json` (user mode) or `/etc/netmie/v2ray-config.json` (daemon mode)
- Format: Standard V2Ray JSON configuration
- Supports inbounds (HTTP) and outbounds (VMess, VLESS, etc.)
- Configuration is loaded by daemon and validated before starting
- Daemon logs to `/var/log/netbird/client.log`

### Daemon Socket
- Unix socket path: `/var/run/netbird.sock`
- Used for CLI-to-daemon communication
- Requires root access for daemon

## Important Implementation Details

### V2Ray Lifecycle
1. CLI sends `VUp` RPC to daemon
2. Daemon acquires `v2rayMutex` lock
3. Daemon loads config from `/etc/netmie/v2ray-config.json` (or `~/.netmie/v2ray-config.json` in user mode)
4. Daemon creates and starts V2Ray engine
5. Engine wraps v2ray-core and manages lifecycle

### Configuration Updates
- `VConfig` RPC accepts absolute file path to temporary config
- Daemon reads and validates the configuration
- Daemon writes validated config to `/etc/netmie/v2ray-config.json`
- If V2Ray is running, it's restarted with new config
- Config version is tracked for status reporting

### Configuration Structure
- Use `json.RawMessage` for `Settings`, `TLS`, `TCP`, `WS` fields in config structs
- This preserves exact JSON structure when marshaling/unmarshaling
- Prevents type conversion issues when passing config to v2ray-core
- See `client/internal/v2ray/config.go` for struct definitions

### Error Handling
- Daemon returns gRPC errors with descriptive messages
- CLI displays user-friendly error messages
- Daemon logs to `/var/log/netbird/client.log` (not journalctl)
- Common errors:
  - "cannot unmarshal string into...port" - Port type mismatch, fixed in `mergeInbounds()`
  - "failed to load v2ray config" - Invalid JSON or v2ray-core compatibility issue
  - "config file not found" - Check `/etc/netmie/v2ray-config.json` exists

## Testing Patterns

- Unit tests use `_test.go` suffix
- Mock implementations for testing (check `mock.go` files)
- Integration tests in `client/cmd/` for CLI commands
- Use `testutil_test.go` for common test utilities

## Linting Configuration

The project uses golangci-lint with specific rules in `.golangci.yaml`:
- Enabled linters: errcheck, gosec, govet, staticcheck, revive, etc.
- Exclusions for generated code and specific paths
- Pre-push hook runs `make lint` on changed files

## Common Development Tasks

### Adding a New V2Ray Command
1. Create new file in `client/cmd/` (e.g., `vnewcmd.go`)
2. Define cobra command with gRPC call to daemon
3. Add RPC method to `client/proto/daemon.proto`
4. Implement handler in `client/server/server.go`
5. Run `make lint` before pushing

### V2Ray Configuration Management (vconfig)

The `vconfig` command downloads and decrypts V2Ray configuration from HTTP endpoints using AES-256-GCM encryption.

**Setup:**
1. Generate encryption key (one-time):
   ```bash
   go run ./tools/generate-key/main.go
   ```
   This outputs a base64-encoded 32-byte key.

2. Copy the key to `client/cmd/encryption_key.txt`:
   ```bash
   cp client/cmd/encryption_key.txt.example client/cmd/encryption_key.txt
   # Edit encryption_key.txt and paste the generated key
   ```

3. Rebuild the client binary (key is embedded at compile time):
   ```bash
   go build -o ./netmie ./client
   ```

4. Share the same key with your server for encryption.

**Client usage:**
```bash
# Download and decrypt config from HTTP endpoint
netmie vconfig https://config-server.com/v2ray-config
```

**How it works:**
1. CLI downloads encrypted config from HTTP URL (expects JSON response with `{code: 20000, data: {data: "base64", encrypted: true}}`)
2. Decrypts using embedded AES-256 key (from `encryption_key.txt`)
3. Merges server-provided outbounds with client-generated inbounds (HTTP proxy on specified port)
4. **Important:** Fixes port type conversion - server may return port as string, client converts to int
5. Sends merged config to daemon via gRPC
6. Daemon validates and stores config at `/etc/netmie/v2ray-config.json`

**Security notes:**
- Encryption key is embedded in binary at compile time (base64-encoded)
- `encryption_key.txt` is in `.gitignore` and NOT committed to git
- Only clients with correct key can decrypt config
- AES-256-GCM provides both confidentiality and authenticity
- Each encryption uses random nonce (no two ciphertexts are identical)
- If key is compromised, regenerate key, rebuild binary, and redeploy

**Server response format:**
The server must return JSON in this format:
```json
{
  "code": 20000,
  "data": {
    "data": "base64-encoded-encrypted-config",
    "encrypted": true
  }
}
```

The decrypted config should contain outbounds with VMess/VLESS servers. Example entry format:
```json
{
  "ps": "server-name",
  "port": "17893",  // Can be string or int, client will convert
  "id": "uuid",
  "aid": 0,
  "net": "tcp",
  "type": "none",
  "tls": "none",
  "add": "server-ip"
}
```

**Common issues:**
- Port type mismatch: If v2ray-core reports "cannot unmarshal string into...port of type uint16", the server is returning port as string. The `mergeInbounds()` function in `vconfig.go` handles this conversion automatically.
- Config structure: Use `json.RawMessage` for nested settings to preserve exact JSON structure when passing to v2ray-core.

### Debugging V2Ray Issues
```bash
# Check daemon status
sudo systemctl status netbird

# View daemon logs (daemon logs to file, not journalctl)
sudo tail -f /var/log/netbird/client.log

# Filter V2Ray-related logs
sudo grep -i v2ray /var/log/netbird/client.log

# Check V2Ray status via CLI
netmie vstatus

# Verify configuration
cat /etc/netmie/v2ray-config.json

# Check port availability
netstat -tlnp | grep 10808
```

### Modifying V2Ray Engine
- Changes to `client/internal/v2ray/engine.go` affect all V2Ray operations
- Ensure thread-safety with mutex locks
- Test with both proxy and TUN modes
- Update status tracking in `EngineStatus` enum if needed

## Dependencies

- Go 1.25+
- v2ray-core v5.46.0 (imported as dependency)
- gRPC and Protocol Buffers
- Linux: iproute2, iptables/nftables
- macOS/Windows: No additional system dependencies

## Platform-Specific Code

- `client/firewall/` - Platform-specific firewall rules (Linux iptables/nftables, Windows USP filter)
- `client/internal/v2ray/tun_*.go` - TUN interface implementations per platform
- `client/cmd/debug_*.go` - Platform-specific debug commands
- `client/server/state_*.go` - Platform-specific state management
