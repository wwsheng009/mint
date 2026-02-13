package instance

// InstanceRegistry manages component instances with dual indexing
// - NodeID index (uint64): Primary lookup using stable runtime identifier
// - Key index (string): Legacy fallback for migration period
//
// This allows gradual migration from key-based to NodeID-based identity
type InstanceRegistry struct {
	instancesByID map[uint64]*Instance // NEW: NodeID index (primary)
	instancesByKey map[string]*Instance // KEEP: backward compatibility during migration
}

// NewInstanceRegistry creates a new instance registry
func NewInstanceRegistry() *InstanceRegistry {
	return &InstanceRegistry{
		instancesByID: make(map[uint64]*Instance),
		instancesByKey: make(map[string]*Instance),
	}
}

// RegisterInstance registers instance with both NodeID and key
// The instance is stored under both indexes for lookup
func (r *InstanceRegistry) RegisterInstance(instance *Instance, key string, nodeID uint64) {
	// Index by NodeID (primary)
	if nodeID != 0 {
		r.instancesByID[nodeID] = instance
		// Update the instance's NodeID field
		instance.NodeID = nodeID
	}

	// Index by key (legacy support, may be empty)
	if key != "" {
		r.instancesByKey[key] = instance
	}
}

// GetInstanceByID gets instance by NodeID (NEW: primary method)
// This is the preferred lookup method in Phase 4+
func (r *InstanceRegistry) GetInstanceByID(nodeID uint64) *Instance {
	return r.instancesByID[nodeID]
}

// GetInstanceByKey gets instance by key (KEPT: fallback during migration)
// This is maintained for backward compatibility
func (r *InstanceRegistry) GetInstanceByKey(key string) *Instance {
	return r.instancesByKey[key]
}

// UnregisterInstance removes instance from both indexes
func (r *InstanceRegistry) UnregisterInstance(instance *Instance) {
	if instance == nil {
		return
	}

	// Remove from NodeID index
	if instance.NodeID != 0 {
		delete(r.instancesByID, instance.NodeID)
	}

	// Remove from key index
	if instance.ID != "" {
		delete(r.instancesByKey, instance.ID)
	}
}

// Clear removes all instances from registry
func (r *InstanceRegistry) Clear() {
	r.instancesByID = make(map[uint64]*Instance)
	r.instancesByKey = make(map[string]*Instance)
}

// Count returns total number of instances registered
func (r *InstanceRegistry) Count() int {
	return len(r.instancesByID)
}

// GetAllInstances returns all instances by NodeID
func (r *InstanceRegistry) GetAllInstances() map[uint64]*Instance {
	result := make(map[uint64]*Instance, len(r.instancesByID))
	for id, inst := range r.instancesByID {
		result[id] = inst
	}
	return result
}
