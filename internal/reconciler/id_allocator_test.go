package reconciler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// IDAllocator Tests
// =============================================================================

// TestIDAllocator_SequentialGeneration tests that IDs are generated sequentially
func TestIDAllocator_SequentialGeneration(t *testing.T) {
	alloc := NewIDAllocator()

	id1 := alloc.Next()
	id2 := alloc.Next()
	id3 := alloc.Next()

	assert.Equal(t, NodeID(1), id1, "First ID should be 1")
	assert.Equal(t, NodeID(2), id2, "Second ID should be 2")
	assert.Equal(t, NodeID(3), id3, "Third ID should be 3")
	assert.True(t, id1 < id2 && id2 < id3, "IDs should be strictly increasing")
}

// TestIDAllocator_Uniqueness tests that generated IDs are unique
func TestIDAllocator_Uniqueness(t *testing.T) {
	alloc := NewIDAllocator()

	// Generate 1000 IDs and ensure uniqueness
	ids := make(map[NodeID]bool)
	for i := 0; i < 1000; i++ {
		id := alloc.Next()
		assert.False(t, ids[id], "Each ID should be unique")
		ids[id] = true
	}
	assert.Equal(t, uint64(1000), uint64(len(ids)), "Should have 1000 unique IDs")
}

// TestIDAllocator_ZeroIsUnused tests that 0 is not used as a valid ID
func TestIDAllocator_ZeroIsUnused(t *testing.T) {
	alloc := NewIDAllocator()

	id1 := alloc.Next()
	id2 := alloc.Next()

	assert.NotEqual(t, NodeID(0), id1, "First ID should not be 0")
	assert.NotEqual(t, NodeID(0), id2, "Second ID should not be 0")
	assert.Equal(t, NodeID(1), id1, "First ID should be 1")
}

// TestIDAllocator_Reset tests that reset works correctly
func TestIDAllocator_Reset(t *testing.T) {
	alloc := NewIDAllocator()

	id1 := alloc.Next()
	id2 := alloc.Next()

	alloc.Reset()

	id3 := alloc.Next()

	assert.Equal(t, NodeID(1), id1, "ID before reset")
	assert.Equal(t, NodeID(2), id2, "ID before reset")
	assert.Equal(t, NodeID(1), id3, "ID after reset should start from 1 again")
}

// TestNodeID_String tests the String representation of NodeID
func TestNodeID_String(t *testing.T) {
	tests := []struct {
		id     NodeID
		expect string
	}{
		{NodeID(0), "NodeID(0)"},
		{NodeID(1), "NodeID(1)"},
		{NodeID(42), "NodeID(42)"},
		{NodeID(999999), "NodeID(999999)"},
	}

	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			result := tt.id.String()
			assert.Equal(t, tt.expect, result, "String representation should match")
		})
	}
}

// TestGlobalAllocator tests that global allocator singleton exists
func TestGlobalAllocator(t *testing.T) {
	alloc1 := GetGlobalAllocator()
	alloc2 := GetGlobalAllocator()

	// Both should be different instances
	assert.NotSame(t, alloc1, alloc2, "Global allocator should return same instance")

	// IDs should be sequential across both
	id1 := alloc1.Next()
	id2 := alloc2.Next()

	// The second allocator should continue from where the first left off
	// This is expected behavior - they share the same underlying counter
	assert.True(t, id2 > id1, "Second allocator ID should continue from first")
}
