package reconciler

import "fmt"

// =============================================================================
// ID Allocator - Runtime NodeID Generation
// =============================================================================
// This module provides unique runtime NodeID allocation for Fiber nodes.
//
// The NodeID is a stable identifier assigned during Fiber mount and
// preserved during Fiber clone operations. This provides a reliable
// runtime identity independent of VNode keys and paths.
//
// See: docs/render/fiber/IDENTITY_REFACTORING_PLAN.md

// =============================================================================
// NodeID Type
// =============================================================================

// NodeID is a unique runtime identifier for a Fiber node
// Unlike string keys, NodeID is stable across layer reorganization
// and provides O(1) map lookup performance
type NodeID uint64

// String returns string representation of NodeID for debugging
func (n NodeID) String() string {
	if n == 0 {
		return "NodeID(0)"
	}
	return fmt.Sprintf("NodeID(%d)", n)
}

// =============================================================================
// IDAllocator
// =============================================================================

// IDAllocator generates sequential unique NodeID values
//
// NodeIDs are allocated:
// - When creating new Fiber nodes (mount)
// - Preserved when cloning existing Fiber nodes (update)
//
// This ensures:
// - Stable identity across re-renders
// - Layer reorganization does not affect IDs
// - No dependency on path strings
type IDAllocator struct {
	next uint64
}

// NewIDAllocator creates a new ID allocator starting from 1
func NewIDAllocator() *IDAllocator {
	return &IDAllocator{
		next: 0, // First call to Next() will return 1
	}
}

// Next returns the next unique NodeID
// The returned value is guaranteed to be unique for this allocator
func (a *IDAllocator) Next() NodeID {
	a.next++
	return NodeID(a.next)
}

// Reset resets the allocator to initial state
// Use with caution - this can lead to duplicate IDs if existing fibers remain
func (a *IDAllocator) Reset() {
	a.next = 0
}

// =============================================================================
// Global Allocator (initialized by reconciler)
// =============================================================================

// globalIDAllocator is the shared allocator for all Fiber nodes
// It is initialized by the reconciler package
var globalIDAllocator = NewIDAllocator()

// =============================================================================
// Helper Functions
// =============================================================================

// GetGlobalAllocator returns the global ID allocator
func GetGlobalAllocator() *IDAllocator {
	return globalIDAllocator
}

// GenerateNodeID is a convenience function to get a new NodeID
func GenerateNodeID() NodeID {
	return globalIDAllocator.Next()
}
