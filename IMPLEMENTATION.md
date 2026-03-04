# NetBird → Netmie V2Ray Integration

## Implementation Status

This implementation provides the foundation for integrating V2Ray capabilities into Netmie (formerly NetBird).

### Completed Components

#### Phase 1: Basic Infrastructure ✓
- [x] Renamed command from `netbird` to `netmie` in root.go
- [x] Added xray-core dependency to go.mod
- [x] Created directory structure:
  - `client/internal/v2ray/` - V2Ray core module
  - `client/internal/coordinator/` - Coexistence coordinator
  - `assets/` - Configuration and geo files

#### Phase 2: V2Ray Core Module ✓
- [x] **config.go** - Configuration management with JSON parsing
- [x] **tun.go** - TUN interface management (cross-platform)
- [x] **tun_linux.go** - Linux TUN implementation
- [x] **tun_darwin.go** - macOS TUN implementation
- [x] **tun_windows.go** - Windows TUN placeholder
- [x] **router.go** - Routing table management (ID: 0x1BD1)
- [x] **engine.go** - V2Ray engine lifecycle management
- [x] **xray_wrapper.go** - xray-core integration wrapper (placeholder)
- [x] **geosite.go** - GeoSite data management
- [x] **geoip.go** - GeoIP data management

#### Phase 3: CLI Commands ✓
- [x] **vup.go** - Start V2Ray connection
- [x] **vdown.go** - Stop V2Ray connection
- [x] **vstatus.go** - Query V2Ray status
- [x] Extended **daemon.proto** with V2Ray RPC methods:
  - `VUp(VUpRequest) returns (VUpResponse)`
  - `VDown(VDownRequest) returns (VDownResponse)`
  - `VStatus(VStatusRequest) returns (VStatusResponse)`

#### Phase 4: Coordinator ✓
- [x] **manager.go** - Coexistence manager for NetBird + V2Ray
- [x] **route_arbiter.go** - Routing table allocation (NetBird: 0x1BD0, V2Ray: 0x1BD1)
- [x] **dns_arbiter.go** - DNS port allocation (NetBird: 53, V2Ray: 5353)

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                      Netmie Client                           │
│  ┌──────────────────┐         ┌──────────────────────────┐ │
│  │ NetBird Module   │         │ V2Ray Module             │ │
│  │ - wt0 interface  │         │ - v2ray0 interface       │ │
│  │ - WireGuard      │         │ - xray-core              │ │
│  │ - Route Table    │         │ - Route Table            │ │
│  │   0x1BD0         │         │   0x1BD1                 │ │
│  └──────────────────┘         └──────────────────────────┘ │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ Coordinator                                           │  │
│  │ - Resource allocation                                 │  │
│  │ - Conflict resolution                                 │  │
│  │ - Coexistence management                              │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Key Design Decisions

1. **Modular Architecture**: V2Ray is a separate module with minimal coupling to NetBird core
2. **Independent Routing Tables**:
   - NetBird: 0x1BD0
   - V2Ray: 0x1BD1
3. **DNS Port Separation**:
   - NetBird: 53
   - V2Ray: 5353
4. **Coordinator Pattern**: Centralized resource management and conflict resolution
5. **Platform Abstraction**: Platform-specific implementations for TUN and routing

### Usage

```bash
# Start NetBird connection
netmie up

# Start V2Ray connection
netmie vup

# Check V2Ray status
netmie vstatus

# Stop V2Ray connection
netmie vdown

# Stop NetBird connection
netmie down
```

### Configuration

V2Ray configuration should be placed at:
- Linux/macOS: `~/.netmie/v2ray-config.json`
- Windows: `%PROGRAMDATA%\Netmie\v2ray-config.json`

Sample configuration is available in `assets/v2ray-config-sample.json`.

### Next Steps

#### Immediate (Required for Basic Functionality)
1. **Implement xray-core integration** in `xray_wrapper.go`:
   - Convert Config to xray-core format
   - Create and manage xray instance
   - Handle lifecycle events

2. **Implement daemon server handlers**:
   - Add VUp/VDown/VStatus handlers to `client/server/server.go`
   - Integrate with V2Ray engine
   - Connect to coordinator

3. **Generate protobuf code**:
   ```bash
   protoc --go_out=. --go-grpc_out=. client/proto/daemon.proto
   ```

4. **Download Geo files**:
   - geosite.dat from https://github.com/v2fly/domain-list-community
   - geoip.dat from https://github.com/v2fly/geoip

#### Phase 5: Configuration Management
- Implement Management Service V2Ray API
- Add configuration sync from server
- Support hot reload of configuration

#### Phase 6: Mobile Support
- Implement VPN manager for Android/iOS
- Add mutual exclusion mode
- Integrate with mobile TUN interfaces

#### Phase 7: Testing & Optimization
- Unit tests for all modules
- Integration tests for coexistence
- Performance optimization
- Platform compatibility testing

### Dependencies

- **xray-core v1.8.7**: Core V2Ray implementation
- **vishvananda/netlink**: Linux network interface management
- **sirupsen/logrus**: Logging
- **spf13/cobra**: CLI framework
- **grpc**: RPC communication

### Platform Support

| Platform | TUN | Routing | Status |
|----------|-----|---------|--------|
| Linux    | ✓   | ✓       | Implemented |
| macOS    | ✓   | ✓       | Implemented |
| Windows  | ○   | ○       | Placeholder |
| Android  | -   | -       | Planned |
| iOS      | -   | -       | Planned |

Legend: ✓ Implemented, ○ Placeholder, - Not started

### Known Limitations

1. **xray-core integration is placeholder**: Actual xray instance creation not implemented
2. **Windows support incomplete**: TUN and routing need implementation
3. **Mobile platforms not started**: Android/iOS support pending
4. **No configuration sync**: Management service integration pending
5. **Geo files not bundled**: Need to be downloaded separately

### Contributing

When implementing remaining features:
1. Follow the modular architecture pattern
2. Add platform-specific implementations in separate files
3. Use the coordinator for resource management
4. Add comprehensive error handling and logging
5. Write tests for new functionality

### License

Same as NetBird project.
