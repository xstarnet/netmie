package v2ray

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	log "github.com/sirupsen/logrus"
)

// Config represents V2Ray configuration
type Config struct {
	Inbounds  []Inbound  `json:"inbounds"`
	Outbounds []Outbound `json:"outbounds"`
	Routing   *Routing   `json:"routing,omitempty"`
	DNS       *DNS       `json:"dns,omitempty"`
}

type Inbound struct {
	Tag      string                 `json:"tag"`
	Protocol string                 `json:"protocol"`
	Port     int                    `json:"port,omitempty"`
	Listen   string                 `json:"listen,omitempty"`
	Settings map[string]interface{} `json:"settings,omitempty"`
}

type Outbound struct {
	Tag            string                 `json:"tag,omitempty"`
	Protocol       string                 `json:"protocol"`
	Settings       map[string]interface{} `json:"settings,omitempty"`
	StreamSettings *StreamSettings        `json:"streamSettings,omitempty"`
}

type StreamSettings struct {
	Network  string                 `json:"network,omitempty"`
	Security string                 `json:"security,omitempty"`
	TLS      map[string]interface{} `json:"tlsSettings,omitempty"`
	TCP      map[string]interface{} `json:"tcpSettings,omitempty"`
	WS       map[string]interface{} `json:"wsSettings,omitempty"`
}

type Routing struct {
	DomainStrategy string `json:"domainStrategy,omitempty"`
	Rules          []Rule `json:"rules,omitempty"`
}

type Rule struct {
	Type        string   `json:"type"`
	Domain      []string `json:"domain,omitempty"`
	IP          []string `json:"ip,omitempty"`
	OutboundTag string   `json:"outboundTag"`
}

type DNS struct {
	Servers []string `json:"servers"`
}

// ConfigManager manages V2Ray configuration
type ConfigManager struct {
	mu         sync.RWMutex
	config     *Config
	configPath string
	version    string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(configPath string) *ConfigManager {
	return &ConfigManager{
		configPath: configPath,
	}
}

// Load loads configuration from file
func (cm *ConfigManager) Load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	cm.config = &config
	log.Infof("V2Ray config loaded from %s", cm.configPath)
	return nil
}

// Save saves configuration to file
func (cm *ConfigManager) Save() error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.config == nil {
		return fmt.Errorf("no config to save")
	}

	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(cm.configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(cm.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	log.Infof("V2Ray config saved to %s", cm.configPath)
	return nil
}

// Get returns the current configuration
func (cm *ConfigManager) Get() *Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config
}

// Set sets a new configuration
func (cm *ConfigManager) Set(config *Config) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config = config
}

// GetVersion returns the configuration version
func (cm *ConfigManager) GetVersion() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.version
}

// SetVersion sets the configuration version
func (cm *ConfigManager) SetVersion(version string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.version = version
}

// DefaultConfigPath returns the default configuration file path
func DefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "/etc/netmie/v2ray-config.json"
	}
	return filepath.Join(homeDir, ".netmie", "v2ray-config.json")
}
