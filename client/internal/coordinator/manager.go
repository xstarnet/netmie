package coordinator

import (
	"context"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
)

// EngineType represents the type of VPN engine
type EngineType int

const (
	EngineNetBird EngineType = iota
	EngineV2Ray
)

func (e EngineType) String() string {
	switch e {
	case EngineNetBird:
		return "NetBird"
	case EngineV2Ray:
		return "V2Ray"
	default:
		return "Unknown"
	}
}

// EngineState represents the state of an engine
type EngineState int

const (
	StateStopped EngineState = iota
	StateRunning
)

// Manager coordinates multiple VPN engines
type Manager struct {
	mu           sync.RWMutex
	engines      map[EngineType]EngineState
	routeArbiter *RouteArbiter
	dnsArbiter   *DNSArbiter
	allowCoexist bool
}

// NewManager creates a new coordinator manager
func NewManager(allowCoexist bool) *Manager {
	return &Manager{
		engines: map[EngineType]EngineState{
			EngineNetBird: StateStopped,
			EngineV2Ray:   StateStopped,
		},
		routeArbiter: NewRouteArbiter(),
		dnsArbiter:   NewDNSArbiter(),
		allowCoexist: allowCoexist,
	}
}

// RequestStart requests to start an engine
func (m *Manager) RequestStart(ctx context.Context, engineType EngineType) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if engine is already running
	if m.engines[engineType] == StateRunning {
		return fmt.Errorf("%s engine is already running", engineType)
	}

	// Check coexistence policy
	if !m.allowCoexist {
		// Check if any other engine is running
		for et, state := range m.engines {
			if et != engineType && state == StateRunning {
				return fmt.Errorf("cannot start %s: %s is already running (coexistence not allowed)", engineType, et)
			}
		}
	}

	// Allocate resources
	if err := m.allocateResources(engineType); err != nil {
		return fmt.Errorf("failed to allocate resources: %w", err)
	}

	m.engines[engineType] = StateRunning
	log.Infof("Coordinator: %s engine started", engineType)
	return nil
}

// RequestStop requests to stop an engine
func (m *Manager) RequestStop(ctx context.Context, engineType EngineType) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if engine is running
	if m.engines[engineType] != StateRunning {
		return fmt.Errorf("%s engine is not running", engineType)
	}

	// Release resources
	if err := m.releaseResources(engineType); err != nil {
		log.Warnf("Failed to release resources for %s: %v", engineType, err)
	}

	m.engines[engineType] = StateStopped
	log.Infof("Coordinator: %s engine stopped", engineType)
	return nil
}

// GetState returns the state of an engine
func (m *Manager) GetState(engineType EngineType) EngineState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.engines[engineType]
}

// GetAllStates returns the states of all engines
func (m *Manager) GetAllStates() map[EngineType]EngineState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	states := make(map[EngineType]EngineState)
	for et, state := range m.engines {
		states[et] = state
	}
	return states
}

// allocateResources allocates resources for an engine
func (m *Manager) allocateResources(engineType EngineType) error {
	// Allocate routing table
	if err := m.routeArbiter.Allocate(engineType); err != nil {
		return fmt.Errorf("failed to allocate routing table: %w", err)
	}

	// Allocate DNS port
	if err := m.dnsArbiter.Allocate(engineType); err != nil {
		_ = m.routeArbiter.Release(engineType)
		return fmt.Errorf("failed to allocate DNS port: %w", err)
	}

	return nil
}

// releaseResources releases resources for an engine
func (m *Manager) releaseResources(engineType EngineType) error {
	if err := m.routeArbiter.Release(engineType); err != nil {
		log.Warnf("Failed to release routing table for %s: %v", engineType, err)
	}

	if err := m.dnsArbiter.Release(engineType); err != nil {
		log.Warnf("Failed to release DNS port for %s: %v", engineType, err)
	}

	return nil
}
