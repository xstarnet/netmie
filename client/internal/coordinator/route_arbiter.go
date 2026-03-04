package coordinator

import (
	"fmt"
	"sync"
)

const (
	// NetBirdRouteTableID is the routing table ID for NetBird
	NetBirdRouteTableID = 0x1BD0
	// V2RayRouteTableID is the routing table ID for V2Ray
	V2RayRouteTableID = 0x1BD1
)

// RouteArbiter manages routing table allocation
type RouteArbiter struct {
	mu         sync.RWMutex
	allocation map[EngineType]int
}

// NewRouteArbiter creates a new route arbiter
func NewRouteArbiter() *RouteArbiter {
	return &RouteArbiter{
		allocation: make(map[EngineType]int),
	}
}

// Allocate allocates a routing table ID for an engine
func (ra *RouteArbiter) Allocate(engineType EngineType) error {
	ra.mu.Lock()
	defer ra.mu.Unlock()

	// Check if already allocated
	if _, exists := ra.allocation[engineType]; exists {
		return fmt.Errorf("routing table already allocated for %s", engineType)
	}

	// Assign table ID based on engine type
	var tableID int
	switch engineType {
	case EngineNetBird:
		tableID = NetBirdRouteTableID
	case EngineV2Ray:
		tableID = V2RayRouteTableID
	default:
		return fmt.Errorf("unknown engine type: %s", engineType)
	}

	ra.allocation[engineType] = tableID
	return nil
}

// Release releases the routing table for an engine
func (ra *RouteArbiter) Release(engineType EngineType) error {
	ra.mu.Lock()
	defer ra.mu.Unlock()

	if _, exists := ra.allocation[engineType]; !exists {
		return fmt.Errorf("no routing table allocated for %s", engineType)
	}

	delete(ra.allocation, engineType)
	return nil
}

// GetTableID returns the routing table ID for an engine
func (ra *RouteArbiter) GetTableID(engineType EngineType) (int, error) {
	ra.mu.RLock()
	defer ra.mu.RUnlock()

	tableID, exists := ra.allocation[engineType]
	if !exists {
		return 0, fmt.Errorf("no routing table allocated for %s", engineType)
	}

	return tableID, nil
}
