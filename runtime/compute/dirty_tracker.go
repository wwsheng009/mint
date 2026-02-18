// Package compute provides dirty tracking for incremental layout
package compute

import (
	"sync"
)

// DirtyTracker tracks which layout nodes need to be recalculated
type DirtyTracker struct {
	mu    sync.RWMutex
	dirty map[string]bool
}

// NewDirtyTracker creates a new dirty tracker
func NewDirtyTracker() *DirtyTracker {
	return &DirtyTracker{
		dirty: make(map[string]bool),
	}
}

// MarkLayoutDirty marks a VNode (by key) as needing layout
func (t *DirtyTracker) MarkLayoutDirty(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.dirty == nil {
		t.dirty = make(map[string]bool)
	}
	t.dirty[key] = true
}

// NeedLayout checks if a VNode (by key) needs layout
func (t *DirtyTracker) NeedLayout(key string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.dirty == nil {
		return false
	}
	return t.dirty[key]
}

// NeedLayoutBox checks if a ComputedBox needs layout
func (t *DirtyTracker) NeedLayoutBox(box *ComputedBox) bool {
	if box == nil {
		return false
	}
	if box.LayoutDirty {
		return true
	}
	// Check by DiffKey (Fiber-first)
	if box.DiffKey != "" {
		return t.NeedLayout(box.DiffKey)
	}
	// Fallback: Check by VNode key (deprecated, will be removed)
	if box.VNode != nil {
		key := box.VNode.Key()
		if key != "" {
			return t.NeedLayout(key)
		}
	}
	return false
}

// Clear clears all dirty flags
func (t *DirtyTracker) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.dirty = nil
}

// ClearKey clears the dirty flag for a specific key
func (t *DirtyTracker) ClearKey(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.dirty != nil {
		delete(t.dirty, key)
	}
}

// DirtyKeys returns all keys marked as dirty
func (t *DirtyTracker) DirtyKeys() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.dirty == nil {
		return nil
	}
	keys := make([]string, 0, len(t.dirty))
	for key := range t.dirty {
		keys = append(keys, key)
	}
	return keys
}

// MarkSubtreeDirty marks a subtree as needing layout
func (t *DirtyTracker) MarkSubtreeDirty(box *ComputedBox) {
	if box == nil {
		return
	}

	// Mark this box
	box.MarkDirty()

	// Mark all descendants
	t.markDescendantsDirty(box)
}

func (t *DirtyTracker) markDescendantsDirty(box *ComputedBox) {
	box.LayoutDirty = true
	// Use DiffKey (Fiber-first)
	if box.DiffKey != "" {
		t.MarkLayoutDirty(box.DiffKey)
	} else if box.VNode != nil {
		// Fallback: VNode.Key() (deprecated)
		if key := box.VNode.Key(); key != "" {
			t.MarkLayoutDirty(key)
		}
	}
	for _, child := range box.Children {
		t.markDescendantsDirty(child)
	}
}
