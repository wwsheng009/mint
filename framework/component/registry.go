package component

import (
	"sync"
)

// Registry maintains NodeID → Updater mappings for direct Msg routing.
//
// Registry allows the Pump to route MouseMsg directly to components by TargetID,
// bypassing the need for hierarchical event traversal.
//
// Usage:
//   registry := NewRegistry()
//   registry.Register("button1", buttonComponent)
//   updater := registry.Lookup("button1")
//   if updater != nil {
//       updater.Update(mouseMsg)
//   }
type Registry struct {
	mu         sync.RWMutex
	components map[string]Updater
}

// NewRegistry creates a new component registry.
func NewRegistry() *Registry {
	return &Registry{
		components: make(map[string]Updater),
	}
}

// Register registers a component with its NodeID.
//
// If a component with the same ID already exists, it will be replaced.
func (r *Registry) Register(id string, component Updater) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.components[id] = component
}

// Lookup looks up a component by NodeID.
//
// Returns nil if the component is not found.
func (r *Registry) Lookup(id string) Updater {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.components[id]
}

// Unregister removes a component from the registry.
func (r *Registry) Unregister(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.components, id)
}

// Clear removes all components from the registry.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.components = make(map[string]Updater)
}

// Size returns the number of registered components.
func (r *Registry) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.components)
}

// List returns all registered component IDs.
//
// The returned slice is a copy and is safe to modify.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.components))
	for id := range r.components {
		ids = append(ids, id)
	}
	return ids
}
