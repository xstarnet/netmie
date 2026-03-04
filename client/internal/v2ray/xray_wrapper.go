package v2ray

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/v2fly/v2ray-core/v5"
	_ "github.com/v2fly/v2ray-core/v5/main/distro/all"
)

// XrayWrapper wraps v2ray-core functionality
type XrayWrapper struct {
	mu       sync.RWMutex
	config   *Config
	instance *core.Instance
	running  bool
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewXrayWrapper creates a new v2ray wrapper
func NewXrayWrapper(config *Config) *XrayWrapper {
	return &XrayWrapper{
		config:  config,
		running: false,
	}
}

// Start starts v2ray-core
func (xw *XrayWrapper) Start(ctx context.Context) error {
	xw.mu.Lock()
	defer xw.mu.Unlock()

	if xw.running {
		return fmt.Errorf("v2ray is already running")
	}

	log.Info("Starting v2ray-core...")

	// Convert our config to JSON bytes
	configBytes, err := json.Marshal(xw.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Debug: print the actual config being sent to v2ray
	log.Debugf("V2Ray config JSON: %s", string(configBytes))

	// Create v2ray instance from JSON config
	// LoadConfig accepts io.Reader
	reader := bytes.NewReader(configBytes)
	v2rayConfig, err := core.LoadConfig("json", reader)
	if err != nil {
		return fmt.Errorf("failed to load v2ray config: %w", err)
	}

	// Create v2ray instance
	instance, err := core.New(v2rayConfig)
	if err != nil {
		return fmt.Errorf("failed to create v2ray instance: %w", err)
	}

	// Create context for v2ray
	xw.ctx, xw.cancel = context.WithCancel(ctx)

	// Start v2ray instance
	if err := instance.Start(); err != nil {
		return fmt.Errorf("failed to start v2ray: %w", err)
	}

	xw.instance = instance
	xw.running = true
	log.Info("v2ray-core started successfully")
	return nil
}

// Stop stops v2ray-core
func (xw *XrayWrapper) Stop() error {
	xw.mu.Lock()
	defer xw.mu.Unlock()

	if !xw.running {
		return fmt.Errorf("v2ray is not running")
	}

	log.Info("Stopping v2ray-core...")

	// Cancel context
	if xw.cancel != nil {
		xw.cancel()
	}

	// Close v2ray instance
	if xw.instance != nil {
		if err := xw.instance.Close(); err != nil {
			log.Warnf("Error closing v2ray instance: %v", err)
		}
		xw.instance = nil
	}

	xw.running = false
	log.Info("v2ray-core stopped successfully")
	return nil
}

// IsRunning returns whether v2ray is running
func (xw *XrayWrapper) IsRunning() bool {
	xw.mu.RLock()
	defer xw.mu.RUnlock()
	return xw.running
}
