package v2ray

import (
	"context"
	"fmt"
	"net"
	"sync"

	log "github.com/sirupsen/logrus"
)

// EngineStatus represents the V2Ray engine status
type EngineStatus int

const (
	StatusStopped EngineStatus = iota
	StatusStarting
	StatusRunning
	StatusStopping
)

func (s EngineStatus) String() string {
	switch s {
	case StatusStopped:
		return "stopped"
	case StatusStarting:
		return "starting"
	case StatusRunning:
		return "running"
	case StatusStopping:
		return "stopping"
	default:
		return "unknown"
	}
}

// EngineMode represents the V2Ray engine mode
type EngineMode int

const (
	// ModeProxy runs V2Ray as a local proxy (SOCKS5/HTTP)
	ModeProxy EngineMode = iota
	// ModeTUN runs V2Ray with TUN interface (system-wide VPN)
	ModeTUN
)

// Engine represents the V2Ray engine
type Engine struct {
	mu            sync.RWMutex
	status        EngineStatus
	mode          EngineMode
	configManager *ConfigManager
	tunManager    *TunManager
	routeManager  *RouteManager
	xrayWrapper   *XrayWrapper
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewEngine creates a new V2Ray engine in proxy mode
func NewEngine(configPath string) *Engine {
	return &Engine{
		status:        StatusStopped,
		mode:          ModeProxy,
		configManager: NewConfigManager(configPath),
	}
}

// NewEngineWithMode creates a new V2Ray engine with specified mode
func NewEngineWithMode(configPath string, mode EngineMode) *Engine {
	return &Engine{
		status:        StatusStopped,
		mode:          mode,
		configManager: NewConfigManager(configPath),
	}
}

// Start starts the V2Ray engine
func (e *Engine) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status != StatusStopped {
		return fmt.Errorf("engine is already running or starting")
	}

	e.status = StatusStarting
	log.Infof("Starting V2Ray engine in %s mode...", e.getModeString())

	// Create context
	e.ctx, e.cancel = context.WithCancel(ctx)

	// Load configuration
	if err := e.configManager.Load(); err != nil {
		e.status = StatusStopped
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if ports are available
	if err := e.checkPortsAvailable(); err != nil {
		e.status = StatusStopped
		return fmt.Errorf("port check failed: %w", err)
	}

	// Only create TUN and routes in TUN mode
	if e.mode == ModeTUN {
		// Create TUN interface
		e.tunManager = NewTunManager(nil)
		if err := e.tunManager.Create(); err != nil {
			e.status = StatusStopped
			return fmt.Errorf("failed to create TUN interface: %w", err)
		}

		// Initialize route manager
		e.routeManager = NewRouteManager(&RouteConfig{
			TableID:       V2RayRouteTableID,
			InterfaceName: e.tunManager.GetName(),
		})

		// Setup routes
		if err := e.routeManager.Setup(); err != nil {
			_ = e.tunManager.Destroy()
			e.status = StatusStopped
			return fmt.Errorf("failed to setup routes: %w", err)
		}
	}

	// Start xray-core
	e.xrayWrapper = NewXrayWrapper(e.configManager.Get())
	if err := e.xrayWrapper.Start(e.ctx); err != nil {
		if e.mode == ModeTUN {
			_ = e.routeManager.Cleanup()
			_ = e.tunManager.Destroy()
		}
		e.status = StatusStopped
		return fmt.Errorf("failed to start xray-core: %w", err)
	}

	e.status = StatusRunning
	log.Infof("V2Ray engine started successfully in %s mode", e.getModeString())
	return nil
}

func (e *Engine) getModeString() string {
	if e.mode == ModeTUN {
		return "TUN"
	}
	return "Proxy"
}

// checkPortsAvailable checks if all required ports are available
func (e *Engine) checkPortsAvailable() error {
	config := e.configManager.Get()
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	// Check inbound ports
	for _, inbound := range config.Inbounds {
		if inbound.Port == 0 {
			continue
		}

		// Try to listen on the port
		addr := fmt.Sprintf(":%d", inbound.Port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("port %d is already in use (protocol: %s)", inbound.Port, inbound.Protocol)
		}
		listener.Close()
	}

	return nil
}

// Stop stops the V2Ray engine
func (e *Engine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status != StatusRunning {
		return fmt.Errorf("engine is not running")
	}

	e.status = StatusStopping
	log.Info("Stopping V2Ray engine...")

	// Cancel context
	if e.cancel != nil {
		e.cancel()
	}

	// Stop xray-core
	if e.xrayWrapper != nil {
		if err := e.xrayWrapper.Stop(); err != nil {
			log.Warnf("Failed to stop xray-core: %v", err)
		}
	}

	// Cleanup routes and TUN only in TUN mode
	if e.mode == ModeTUN {
		// Cleanup routes
		if e.routeManager != nil {
			if err := e.routeManager.Cleanup(); err != nil {
				log.Warnf("Failed to cleanup routes: %v", err)
			}
		}

		// Destroy TUN interface
		if e.tunManager != nil {
			if err := e.tunManager.Destroy(); err != nil {
				log.Warnf("Failed to destroy TUN interface: %v", err)
			}
		}
	}

	e.status = StatusStopped
	log.Info("V2Ray engine stopped")
	return nil
}

// GetStatus returns the current engine status
func (e *Engine) GetStatus() EngineStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.status
}

// GetStatusInfo returns detailed status information
func (e *Engine) GetStatusInfo() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	info := map[string]interface{}{
		"status": e.status.String(),
	}

	if e.tunManager != nil {
		info["interface"] = e.tunManager.GetName()
		info["ip"] = e.tunManager.GetIP()
	}

	if e.configManager != nil {
		info["config_version"] = e.configManager.GetVersion()
	}

	return info
}
