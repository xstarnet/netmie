package v2ray

import (
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
)

const (
	// DefaultV2RayInterface is the default TUN interface name for V2Ray
	DefaultV2RayInterface = "v2ray0"
	// DefaultV2RayIP is the default IP address for V2Ray TUN interface
	DefaultV2RayIP = "10.233.0.1/24"
	// DefaultV2RayMTU is the default MTU for V2Ray TUN interface
	DefaultV2RayMTU = 1500
)

// TunConfig represents TUN interface configuration
type TunConfig struct {
	Name string
	IP   string
	MTU  int
}

// TunManager manages V2Ray TUN interface
type TunManager struct {
	config *TunConfig
	fd     int
}

// NewTunManager creates a new TUN manager
func NewTunManager(config *TunConfig) *TunManager {
	if config == nil {
		config = &TunConfig{
			Name: DefaultV2RayInterface,
			IP:   DefaultV2RayIP,
			MTU:  DefaultV2RayMTU,
		}
	}
	return &TunManager{
		config: config,
		fd:     -1,
	}
}

// Create creates and configures the TUN interface
func (tm *TunManager) Create() error {
	return tm.createPlatform()
}

// Destroy destroys the TUN interface
func (tm *TunManager) Destroy() error {
	return tm.destroyPlatform()
}

// GetFD returns the TUN file descriptor
func (tm *TunManager) GetFD() int {
	return tm.fd
}

// GetName returns the TUN interface name
func (tm *TunManager) GetName() string {
	return tm.config.Name
}

// GetIP returns the TUN interface IP address
func (tm *TunManager) GetIP() string {
	return tm.config.IP
}

// parseIP parses IP address with CIDR notation
func parseIP(ipStr string) (net.IP, *net.IPNet, error) {
	ip, ipNet, err := net.ParseCIDR(ipStr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid IP address: %w", err)
	}
	return ip, ipNet, nil
}

// logTunCreated logs TUN interface creation
func (tm *TunManager) logTunCreated() {
	log.Infof("V2Ray TUN interface created: %s (%s, MTU: %d)",
		tm.config.Name, tm.config.IP, tm.config.MTU)
}

// logTunDestroyed logs TUN interface destruction
func (tm *TunManager) logTunDestroyed() {
	log.Infof("V2Ray TUN interface destroyed: %s", tm.config.Name)
}
