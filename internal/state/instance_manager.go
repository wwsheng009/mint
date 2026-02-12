package state

import (
	"sync"
	"time"
)

// InstanceManager manages component instances across renders.
// It is responsible for:
// - Creating new instances for unseen VNodes
// - Reusing existing instances when keys match
// - Cleaning up instances that are no longer needed
// - Enforcing max instance limits to prevent memory leaks
type InstanceManager struct {
	mu            sync.RWMutex
	instances     map[string]ComponentInstance // key -> instance
	instanceOrder []string                      // Order of instance creation (for LRU)
	lastAccess    map[string]time.Time          // Last access time for each instance
	maxInstances  int                            // Maximum number of instances to keep
}

// NewInstanceManager creates a new instance manager
func NewInstanceManager() *InstanceManager {
	return &InstanceManager{
		instances:     make(map[string]ComponentInstance),
		instanceOrder: make([]string, 0),
		lastAccess:    make(map[string]time.Time),
		maxInstances:  1000, // Default limit
	}
}

// SetMaxInstances sets the maximum number of instances to keep
// When the limit is exceeded, the least recently used instances are cleaned up
func (m *InstanceManager) SetMaxInstances(max int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.maxInstances = max
	m.cleanupLRU()
}

// GetOrCreate finds an existing instance by key or creates a new one.
// The creator function is only called if no existing instance is found.
//
// This is the core of the instance matching algorithm:
// 1. Check if an instance with the key exists
// 2. If yes, reuse it (update props if changed)
// 3. If no, call creator to make a new instance
// 4. Update last access time
// 5. Enforce max instance limit
func (m *InstanceManager) GetOrCreate(key string, creator func() ComponentInstance) ComponentInstance {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Normalize empty key to a default
	if key == "" {
		key = "__default__"
	}

	// Try to find existing instance
	if inst, exists := m.instances[key]; exists {
		// Update last access time
		m.lastAccess[key] = time.Now()
		// Move to end of order (most recently used)
		m.moveToEnd(key)
		return inst
	}

	// Create new instance
	inst := creator()

	// Store the instance
	m.instances[key] = inst
	m.instanceOrder = append(m.instanceOrder, key)
	m.lastAccess[key] = time.Now()

	// Call OnMount for new instances
	inst.OnMount()

	// Enforce instance limit
	m.cleanupLRU()

	return inst
}

// Get retrieves an instance by key without creating a new one
// Returns nil if the instance doesn't exist
func (m *InstanceManager) Get(key string) ComponentInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if key == "" {
		key = "__default__"
	}

	return m.instances[key]
}

// Remove removes an instance by key
// Returns the removed instance or nil if not found
func (m *InstanceManager) Remove(key string) ComponentInstance {
	m.mu.Lock()
	defer m.mu.Unlock()

	if key == "" {
		key = "__default__"
	}

	inst, exists := m.instances[key]
	if !exists {
		return nil
	}

	// Call OnUnmount before removing
	inst.OnUnmount()

	// Remove from maps
	delete(m.instances, key)
	delete(m.lastAccess, key)

	// Remove from order slice
	for i, k := range m.instanceOrder {
		if k == key {
			m.instanceOrder = append(m.instanceOrder[:i], m.instanceOrder[i+1:]...)
			break
		}
	}

	return inst
}

// Update updates an instance's props if they have changed
// Returns true if the props were different and the instance was updated
func (m *InstanceManager) Update(key string, newProps map[string]interface{}) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if key == "" {
		key = "__default__"
	}

	inst, exists := m.instances[key]
	if !exists {
		return false
	}

	return inst.SetProps(newProps)
}

// Cleanup removes all instances that are not in the active keys set
// This should be called after each render to clean up unused instances
func (m *InstanceManager) Cleanup(activeKeys []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a set of active keys
	activeSet := make(map[string]bool)
	for _, key := range activeKeys {
		if key == "" {
			key = "__default__"
		}
		activeSet[key] = true
	}

	// Find keys to remove
	var toRemove []string
	for key := range m.instances {
		if !activeSet[key] {
			toRemove = append(toRemove, key)
		}
	}

	// Remove unused instances
	for _, key := range toRemove {
		inst := m.instances[key]
		inst.OnUnmount()

		delete(m.instances, key)
		delete(m.lastAccess, key)

		// Remove from order
		for i, k := range m.instanceOrder {
			if k == key {
				m.instanceOrder = append(m.instanceOrder[:i], m.instanceOrder[i+1:]...)
				break
			}
		}
	}
}

// Clear removes all instances
func (m *InstanceManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Call OnUnmount for all instances
	for _, inst := range m.instances {
		inst.OnUnmount()
	}

	m.instances = make(map[string]ComponentInstance)
	m.instanceOrder = make([]string, 0)
	m.lastAccess = make(map[string]time.Time)
}

// Count returns the number of instances
func (m *InstanceManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.instances)
}

// cleanupLRU removes least recently used instances when limit is exceeded
func (m *InstanceManager) cleanupLRU() {
	for len(m.instances) > m.maxInstances && len(m.instanceOrder) > 0 {
		// Remove oldest (first in order)
		oldestKey := m.instanceOrder[0]

		// Unmount the instance
		if inst, exists := m.instances[oldestKey]; exists {
			inst.OnUnmount()
		}

		// Remove from maps
		delete(m.instances, oldestKey)
		delete(m.lastAccess, oldestKey)

		// Remove from order
		m.instanceOrder = m.instanceOrder[1:]
	}
}

// moveToEnd moves a key to the end of the instance order (most recently used)
func (m *InstanceManager) moveToEnd(key string) {
	for i, k := range m.instanceOrder {
		if k == key {
			// Remove from current position
			m.instanceOrder = append(m.instanceOrder[:i], m.instanceOrder[i+1:]...)
			// Add to end
			m.instanceOrder = append(m.instanceOrder, key)
			break
		}
	}
}

// GetDirtyInstances returns all instances marked as dirty
// These instances need to be re-rendered
func (m *InstanceManager) GetDirtyInstances() []ComponentInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	dirty := make([]ComponentInstance, 0)
	for _, inst := range m.instances {
		if inst.IsDirty() {
			dirty = append(dirty, inst)
		}
	}
	return dirty
}

// MarkAllDirty marks all instances as needing re-render
func (m *InstanceManager) MarkAllDirty() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, inst := range m.instances {
		inst.MarkDirty()
	}
}

// GetAllInstances returns all instances managed by this InstanceManager
// This is used by HitMap enrichment to connect ComponentInstances with layout nodes
func (m *InstanceManager) GetAllInstances() map[string]ComponentInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid concurrent modification
	result := make(map[string]ComponentInstance, len(m.instances))
	for key, inst := range m.instances {
		result[key] = inst
	}
	return result
}
