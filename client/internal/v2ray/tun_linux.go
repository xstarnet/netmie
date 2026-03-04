package v2ray

import (
	"fmt"
	"net"
	"os/exec"

	"github.com/vishvananda/netlink"
)

// createPlatform creates TUN interface on Linux
func (tm *TunManager) createPlatform() error {
	return tm.createLinux()
}

// destroyPlatform destroys TUN interface on Linux
func (tm *TunManager) destroyPlatform() error {
	return tm.destroyLinux()
}

// createLinux creates TUN interface on Linux
func (tm *TunManager) createLinux() error {
	// Create TUN interface using ip tuntap
	cmd := exec.Command("ip", "tuntap", "add", "dev", tm.config.Name, "mode", "tun")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create TUN interface: %w", err)
	}

	// Get the link
	link, err := netlink.LinkByName(tm.config.Name)
	if err != nil {
		return fmt.Errorf("failed to get link: %w", err)
	}

	// Set MTU
	if err := netlink.LinkSetMTU(link, tm.config.MTU); err != nil {
		return fmt.Errorf("failed to set MTU: %w", err)
	}

	// Parse and add IP address
	ip, ipNet, err := parseIP(tm.config.IP)
	if err != nil {
		return err
	}

	addr := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   ip,
			Mask: ipNet.Mask,
		},
	}

	if err := netlink.AddrAdd(link, addr); err != nil {
		return fmt.Errorf("failed to add IP address: %w", err)
	}

	// Bring interface up
	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("failed to bring interface up: %w", err)
	}

	tm.logTunCreated()
	return nil
}

// destroyLinux destroys TUN interface on Linux
func (tm *TunManager) destroyLinux() error {
	// Delete TUN interface
	cmd := exec.Command("ip", "link", "delete", tm.config.Name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete TUN interface: %w", err)
	}

	tm.logTunDestroyed()
	return nil
}
