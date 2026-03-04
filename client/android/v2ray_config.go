//go:build android

package android

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
// This should be called with the app's files directory path from Android Context
func (c *Client) SetV2RayConfigPath(filesDir string) {
	c.v2rayConfigPath = filepath.Join(filesDir, "v2ray-config.json")
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

// getV2RayConfigPath returns the default config path for Android
func (c *Client) getV2RayConfigPath() string {
	// If custom path is set, use it
	if c.v2rayConfigPath != "" {
		return c.v2rayConfigPath
	}

	// Fallback to a default path (should be set via SetV2RayConfigPath from Android app)
	// Using /data/local/tmp as fallback for testing, but production should use app files dir
	return filepath.Join("/data/local/tmp", "v2ray-config.json")
}
