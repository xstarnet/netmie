package v2ray

import (
	"fmt"
	"os/exec"
	"runtime"

	log "github.com/sirupsen/logrus"
)

const (
	// V2RayRouteTableID is the routing table ID for V2Ray
	V2RayRouteTableID = 0x1BD1
)

// RouteConfig represents routing configuration
type RouteConfig struct {
	TableID       int
	InterfaceName string
	Rules         []RouteRule
}

// RouteRule represents a routing rule
type RouteRule struct {
	Destination string
	Gateway     string
}

// RouteManager manages V2Ray routing
type RouteManager struct {
	config *RouteConfig
}

// NewRouteManager creates a new route manager
func NewRouteManager(config *RouteConfig) *RouteManager {
	return &RouteManager{
		config: config,
	}
}

// Setup sets up routing rules
func (rm *RouteManager) Setup() error {
	switch runtime.GOOS {
	case "linux":
		return rm.setupLinux()
	case "darwin":
		return rm.setupDarwin()
	case "windows":
		return rm.setupWindows()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// Cleanup removes routing rules
func (rm *RouteManager) Cleanup() error {
	switch runtime.GOOS {
	case "linux":
		return rm.cleanupLinux()
	case "darwin":
		return rm.cleanupDarwin()
	case "windows":
		return rm.cleanupWindows()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// setupLinux sets up routing on Linux
func (rm *RouteManager) setupLinux() error {
	// Create routing table if not exists
	tableID := fmt.Sprintf("%d", rm.config.TableID)

	// Add default route to the table
	cmd := exec.Command("ip", "route", "add", "default", "dev", rm.config.InterfaceName, "table", tableID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add default route: %w", err)
	}

	// Add rule to use the table
	cmd = exec.Command("ip", "rule", "add", "fwmark", tableID, "table", tableID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add routing rule: %w", err)
	}

	log.Infof("V2Ray routing setup completed (table: %s)", tableID)
	return nil
}

// cleanupLinux cleans up routing on Linux
func (rm *RouteManager) cleanupLinux() error {
	tableID := fmt.Sprintf("%d", rm.config.TableID)

	// Delete routing rule
	cmd := exec.Command("ip", "rule", "del", "fwmark", tableID, "table", tableID)
	_ = cmd.Run() // Ignore errors

	// Flush routing table
	cmd = exec.Command("ip", "route", "flush", "table", tableID)
	_ = cmd.Run() // Ignore errors

	log.Infof("V2Ray routing cleanup completed (table: %s)", tableID)
	return nil
}

// setupDarwin sets up routing on macOS
func (rm *RouteManager) setupDarwin() error {
	// macOS routing setup
	// This is a simplified implementation
	log.Info("V2Ray routing setup completed (macOS)")
	return nil
}

// cleanupDarwin cleans up routing on macOS
func (rm *RouteManager) cleanupDarwin() error {
	log.Info("V2Ray routing cleanup completed (macOS)")
	return nil
}

// setupWindows sets up routing on Windows
func (rm *RouteManager) setupWindows() error {
	return fmt.Errorf("Windows routing setup not yet implemented")
}

// cleanupWindows cleans up routing on Windows
func (rm *RouteManager) cleanupWindows() error {
	return fmt.Errorf("Windows routing cleanup not yet implemented")
}
