package coordinator

import (
	"fmt"
	"sync"
)

const (
	// NetBirdDNSPort is the DNS port for NetBird
	NetBirdDNSPort = 53
	// V2RayDNSPort is the DNS port for V2Ray
	V2RayDNSPort = 5353
)

// DNSArbiter manages DNS port allocation
type DNSArbiter struct {
	mu         sync.RWMutex
	allocation map[EngineType]int
}

// NewDNSArbiter creates a new DNS arbiter
func NewDNSArbiter() *DNSArbiter {
	return &DNSArbiter{
		allocation: make(map[EngineType]int),
	}
}

// Allocate allocates a DNS port for an engine
func (da *DNSArbiter) Allocate(engineType EngineType) error {
	da.mu.Lock()
	defer da.mu.Unlock()

	// Check if already allocated
	if _, exists := da.allocation[engineType]; exists {
		return fmt.Errorf("DNS port already allocated for %s", engineType)
	}

	// Assign port based on engine type
	var port int
	switch engineType {
	case EngineNetBird:
		port = NetBirdDNSPort
	case EngineV2Ray:
		port = V2RayDNSPort
	default:
		return fmt.Errorf("unknown engine type: %s", engineType)
	}

	da.allocation[engineType] = port
	return nil
}

// Release releases the DNS port for an engine
func (da *DNSArbiter) Release(engineType EngineType) error {
	da.mu.Lock()
	defer da.mu.Unlock()

	if _, exists := da.allocation[engineType]; !exists {
		return fmt.Errorf("no DNS port allocated for %s", engineType)
	}

	delete(da.allocation, engineType)
	return nil
}

// GetPort returns the DNS port for an engine
func (da *DNSArbiter) GetPort(engineType EngineType) (int, error) {
	da.mu.RLock()
	defer da.mu.RUnlock()

	port, exists := da.allocation[engineType]
	if !exists {
		return 0, fmt.Errorf("no DNS port allocated for %s", engineType)
	}

	return port, nil
}
