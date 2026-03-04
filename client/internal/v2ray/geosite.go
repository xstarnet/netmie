package v2ray

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// GeoSiteManager manages GeoSite data
type GeoSiteManager struct {
	dataPath string
	data     []byte
}

// NewGeoSiteManager creates a new GeoSite manager
func NewGeoSiteManager(dataPath string) *GeoSiteManager {
	return &GeoSiteManager{
		dataPath: dataPath,
	}
}

// Load loads GeoSite data from file
func (gsm *GeoSiteManager) Load() error {
	if _, err := os.Stat(gsm.dataPath); os.IsNotExist(err) {
		return fmt.Errorf("geosite.dat not found at %s", gsm.dataPath)
	}

	data, err := os.ReadFile(gsm.dataPath)
	if err != nil {
		return fmt.Errorf("failed to read geosite.dat: %w", err)
	}

	gsm.data = data
	log.Infof("GeoSite data loaded from %s (%d bytes)", gsm.dataPath, len(data))
	return nil
}

// GetPath returns the path to geosite.dat
func (gsm *GeoSiteManager) GetPath() string {
	return gsm.dataPath
}

// DefaultGeoSitePath returns the default path for geosite.dat
func DefaultGeoSitePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "/etc/netmie/geosite.dat"
	}
	return filepath.Join(homeDir, ".netmie", "geosite.dat")
}
