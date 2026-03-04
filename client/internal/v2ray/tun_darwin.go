package v2ray

import (
	"fmt"
	"os/exec"
)

// createPlatform creates TUN interface on macOS
func (tm *TunManager) createPlatform() error {
	return tm.createDarwin()
}

// destroyPlatform destroys TUN interface on macOS
func (tm *TunManager) destroyPlatform() error {
	return tm.destroyDarwin()
}

// createDarwin creates TUN interface on macOS
func (tm *TunManager) createDarwin() error {
	// On macOS, we need to use utun interfaces
	// For now, use a simple approach with ifconfig
	cmd := exec.Command("ifconfig", tm.config.Name, "create")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create TUN interface: %w", err)
	}

	// Set IP address
	cmd = exec.Command("ifconfig", tm.config.Name, tm.config.IP, tm.config.IP)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set IP address: %w", err)
	}

	// Set MTU
	cmd = exec.Command("ifconfig", tm.config.Name, "mtu", fmt.Sprintf("%d", tm.config.MTU))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set MTU: %w", err)
	}

	// Bring interface up
	cmd = exec.Command("ifconfig", tm.config.Name, "up")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring interface up: %w", err)
	}

	tm.logTunCreated()
	return nil
}

// destroyDarwin destroys TUN interface on macOS
func (tm *TunManager) destroyDarwin() error {
	cmd := exec.Command("ifconfig", tm.config.Name, "destroy")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to destroy TUN interface: %w", err)
	}

	tm.logTunDestroyed()
	return nil
}
