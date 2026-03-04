package v2ray

import (
	"fmt"
)

// createPlatform creates TUN interface on Windows
func (tm *TunManager) createPlatform() error {
	return tm.createWindows()
}

// destroyPlatform destroys TUN interface on Windows
func (tm *TunManager) destroyPlatform() error {
	return tm.destroyWindows()
}

// createWindows creates TUN interface on Windows
func (tm *TunManager) createWindows() error {
	// Windows TUN creation requires wintun driver
	// This is a placeholder - actual implementation would use wintun
	return fmt.Errorf("Windows TUN creation not yet implemented")
}

// destroyWindows destroys TUN interface on Windows
func (tm *TunManager) destroyWindows() error {
	return fmt.Errorf("Windows TUN destruction not yet implemented")
}
