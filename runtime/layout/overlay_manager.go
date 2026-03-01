package layout

import (
	"sync"
)

// OverlayManager manages a stack of floating layers (portals).
// Portals allow components to mount their children to a different location
// in the layout tree (e.g., Modal/Tooltip mounting to a portal root).
//
// Typical usage:
//   1. App creates a PortalRoot (usually at top level)
//   2. Portal component pushes itself to the stack with portalRootTarget
//   3. Layout engine queries OverlayManager for active portal nodes
//   4. Portal nodes are laid out independently from their parent tree
type OverlayManager struct {
	mu    sync.RWMutex
	stack []*OverlayEntry  // Stack of overlay entries (ordered by priority)

	// Map of portal ID to entry for quick lookup
	entries map[string]*OverlayEntry

	// Cache flag - true if stack needs rebuild
	dirty bool
}

// OverlayEntry represents a single portal in the overlay stack
type OverlayEntry struct {
	// Unique identifier for this portal
	ID string

	// The layout box representing the portal container
	// This is the node that will be rendered in the overlay layer
	Box *LayoutBox

	// The ID of the target fiber where this portal should be mounted
	// This is the PortalRoot target specified by the Portal component
	PortalRootID string

	// Priority for Z-ordering (higher = closer to top)
	Priority int

	// Whether this entry is active (not removed/hidden)
	Active bool
}

// NewOverlayManager creates a new overlay manager
func NewOverlayManager() *OverlayManager {
	return &OverlayManager{
		stack:   make([]*OverlayEntry, 0),
		entries: make(map[string]*OverlayEntry),
		dirty:   false,
	}
}

// Push adds a new portal to the overlay stack
// The portal will be laid out and rendered independently from its parent tree
func (o *OverlayManager) Push(id string, box *LayoutBox, portalRootID string, priority int) {
	o.mu.Lock()
	defer o.mu.Unlock()

	entry := &OverlayEntry{
		ID:           id,
		Box:          box,
		PortalRootID: portalRootID,
		Priority:     priority,
		Active:       true,
	}

	// Remove existing entry with same ID if exists
	if existing, ok := o.entries[id]; ok {
		existing.Active = false
	}

	o.entries[id] = entry
	o.dirty = true
}

// Pop removes and returns the top portal from the stack
// Returns nil if no active portal exists
func (o *OverlayManager) Pop() *OverlayEntry {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Build stack if dirty
	if o.dirty {
		o.rebuildStack()
	}

	// Rebuild again if stack is empty but entries exist
	if len(o.stack) == 0 && len(o.entries) > 0 {
		o.rebuildStack()
	}

	if len(o.stack) == 0 {
		return nil
	}

	// Find highest priority active entry
	var highest *OverlayEntry
	highestIdx := -1

	for i, entry := range o.stack {
		if entry.Active {
			if highest == nil || entry.Priority > highest.Priority {
				highest = entry
				highestIdx = i
			}
		}
	}

	if highest == nil {
		return nil
	}

	// Deactivate and remove
	highest.Active = false
	delete(o.entries, highest.ID)

	// Remove from stack
	o.stack = append(o.stack[:highestIdx], o.stack[highestIdx+1:]...)
	o.dirty = true

	return highest
}

// Top returns the top (highest priority) portal from the stack
// Returns nil if no active portal exists
func (o *OverlayManager) Top() *OverlayEntry {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.dirty {
		o.rebuildStack()
	}

	for i := len(o.stack) - 1; i >= 0; i-- {
		if o.stack[i].Active {
			return o.stack[i]
		}
	}

	return nil
}

// GetAll returns all active portal entries in the stack
// Returned in ascending priority order (lower priority first, higher priority last)
func (o *OverlayManager) GetAll() []*OverlayEntry {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.dirty {
		o.rebuildStack()
	}

	// Filter active entries
	active := make([]*OverlayEntry, 0, len(o.stack))
	for _, entry := range o.stack {
		if entry.Active {
			active = append(active, entry)
		}
	}

	return active
}

// GetByID returns a specific portal entry by ID
// Returns nil if not found or inactive
func (o *OverlayManager) GetByID(id string) *OverlayEntry {
	o.mu.Lock()
	defer o.mu.Unlock()

	if entry, ok := o.entries[id]; ok && entry.Active {
		return entry
	}
	return nil
}

// Remove deactivates and removes a specific portal from the stack
func (o *OverlayManager) Remove(id string) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if entry, ok := o.entries[id]; ok {
		entry.Active = false
		delete(o.entries, id)
		o.dirty = true
	}
}

// Clear removes all portals from the stack
func (o *OverlayManager) Clear() {
	o.mu.Lock()
	defer o.mu.Unlock()

	for _, entry := range o.entries {
		entry.Active = false
	}

	o.entries = make(map[string]*OverlayEntry)
	o.stack = make([]*OverlayEntry, 0)
	o.dirty = false
}

// Size returns the number of active portals in the stack
func (o *OverlayManager) Size() int {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.dirty {
		o.rebuildStack()
	}

	count := 0
	for _, entry := range o.stack {
		if entry.Active {
			count++
		}
	}
	return count
}

// rebuildStack rebuilds the stack with active entries sorted by priority
// Must be called with write lock held
func (o *OverlayManager) rebuildStack() {
	if !o.dirty {
		return
	}

	// Collect active entries
	active := make([]*OverlayEntry, 0, len(o.entries))
	for _, entry := range o.entries {
		if entry.Active {
			active = append(active, entry)
		}
	}

	// Sort by priority (ascending: lower priority first, higher priority last)
	// Equal priority maintains insertion order due to stable sort
	for i := 0; i < len(active); i++ {
		for j := i + 1; j < len(active); j++ {
			if active[i].Priority > active[j].Priority {
				active[i], active[j] = active[j], active[i]
			}
		}
	}

	o.stack = active
	o.dirty = false
}

