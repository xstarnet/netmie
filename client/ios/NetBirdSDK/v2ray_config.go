//go:build ios

package NetBirdSDK

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SetV2RayConfig sets the V2Ray configuration
func (c *Client) SetV2RayConfig(configJSON string) error {
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return fmt.Errorf("invalid JSON config: %w", err)
	}

	configPath := c.getV2RayConfigPath()

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, []byte(configJSON), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	c.v2rayConfigPath = configPath
	return nil
}

// SetV2RayConfigPath sets a custom V2Ray configuration path
// This should be called with the app's container directory or App Group container path
func (c *Client) SetV2RayConfigPath(containerDir string) {
	c.v2rayConfigPath = filepath.Join(containerDir, "v2ray-config.json")
}

// GetV2RayConfig gets the V2Ray configuration
func (c *Client) GetV2RayConfig() (string, error) {
	if c.v2rayConfigPath == "" {
		c.v2rayConfigPath = c.getV2RayConfigPath()
	}

	data, err := os.ReadFile(c.v2rayConfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to read config: %w", err)
	}

	return string(data), nil
}

// getV2RayConfigPath returns the default config path for iOS
func (c *Client) getV2RayConfigPath() string {
	// If custom path is set, use it
	if c.v2rayConfigPath != "" {
		return c.v2rayConfigPath
	}

	// Fallback to Documents directory
	// Production apps should use SetV2RayConfigPath with App Group container
	if homeDir := os.Getenv("HOME"); homeDir != "" {
		return filepath.Join(homeDir, "Documents", "v2ray-config.json")
	}

	// Last resort fallback
	return filepath.Join("/tmp", "v2ray-config.json")
}
