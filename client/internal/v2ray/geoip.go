package v2ray

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// GeoIPManager manages GeoIP data
type GeoIPManager struct {
	dataPath string
	data     []byte
}

// NewGeoIPManager creates a new GeoIP manager
func NewGeoIPManager(dataPath string) *GeoIPManager {
	return &GeoIPManager{
		dataPath: dataPath,
	}
}

// Load loads GeoIP data from file
func (gim *GeoIPManager) Load() error {
	if _, err := os.Stat(gim.dataPath); os.IsNotExist(err) {
		return fmt.Errorf("geoip.dat not found at %s", gim.dataPath)
	}

	data, err := os.ReadFile(gim.dataPath)
	if err != nil {
		return fmt.Errorf("failed to read geoip.dat: %w", err)
	}

	gim.data = data
	log.Infof("GeoIP data loaded from %s (%d bytes)", gim.dataPath, len(data))
	return nil
}

// GetPath returns the path to geoip.dat
func (gim *GeoIPManager) GetPath() string {
	return gim.dataPath
}

// DefaultGeoIPPath returns the default path for geoip.dat
func DefaultGeoIPPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "/etc/netmie/geoip.dat"
	}
	return filepath.Join(homeDir, ".netmie", "geoip.dat")
}
