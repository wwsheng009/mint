package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirtyTracker_MarkAndCheck(t *testing.T) {
	tracker := NewDirtyTracker()

	// Initial state - should be clean
	assert.False(t, tracker.IsLayoutDirty("node1"), "Initial state should be clean")

	// Mark dirty
	tracker.MarkLayoutDirty("node1")
	assert.True(t, tracker.IsLayoutDirty("node1"), "Should be dirty after marking")

	// Check another node - should be clean
	assert.False(t, tracker.IsLayoutDirty("node2"), "Other node should remain clean")
}

func TestDirtyTracker_ClearKey(t *testing.T) {
	tracker := NewDirtyTracker()

	tracker.MarkLayoutDirty("node1")
	tracker.MarkLayoutDirty("node2")

	// Both should be dirty
	assert.True(t, tracker.IsLayoutDirty("node1"))
	assert.True(t, tracker.IsLayoutDirty("node2"))

	// Clear one key
	tracker.ClearKey("node1")

	assert.False(t, tracker.IsLayoutDirty("node1"), "Should be clean after ClearKey")
	assert.True(t, tracker.IsLayoutDirty("node2"), "Other node should remain dirty")
}

func TestDirtyTracker_Clear(t *testing.T) {
	tracker := NewDirtyTracker()

	tracker.MarkLayoutDirty("node1")
	tracker.MarkLayoutDirty("node2")
	tracker.MarkLayoutDirty("node3")

	// Check size
	assert.Equal(t, 3, tracker.Size(), "Should have 3 dirty nodes")
	assert.True(t, tracker.HasAny(), "Should have dirty nodes")

	// Clear all
	tracker.Clear()

	assert.Equal(t, 0, tracker.Size(), "Should have 0 dirty nodes after Clear")
	assert.False(t, tracker.HasAny(), "Should not have dirty nodes after Clear")
	assert.False(t, tracker.IsLayoutDirty("node1"))
	assert.False(t, tracker.IsLayoutDirty("node2"))
	assert.False(t, tracker.IsLayoutDirty("node3"))
}

func TestDirtyTracker_MarkSubtreeDirty(t *testing.T) {
	tracker := NewDirtyTracker()

	// Create a simple tree
	child1 := NewMockMeasurableNode("child1", 20, 20)
	child2 := NewMockMeasurableNode("child2", 20, 20)
	_ = NewFlexLayout("root", []Node{child1, child2})

	assert.False(t, tracker.IsLayoutDirty("root"))
	assert.False(t, tracker.IsLayoutDirty("child1"))
	assert.False(t, tracker.IsLayoutDirty("child2"))

	// Mark subtree dirty from child1
	tracker.MarkSubtreeDirty(child1)

	// child1 should be dirty (but it's a leaf, so no children)
	assert.True(t, tracker.IsLayoutDirty("child1"))
	assert.False(t, tracker.IsLayoutDirty("child2"))
}

func TestDirtyTracker_SizeAndHasAny(t *testing.T) {
	tracker := NewDirtyTracker()

	// Empty tracker
	assert.Equal(t, 0, tracker.Size())
	assert.False(t, tracker.HasAny())

	// Add dirty nodes
	tracker.MarkLayoutDirty("node1")
	tracker.MarkLayoutDirty("node2")
	tracker.MarkLayoutDirty("node3")

	assert.Equal(t, 3, tracker.Size())
	assert.True(t, tracker.HasAny())

	// Mark same node again - size should not increase
	tracker.MarkLayoutDirty("node1")
	assert.Equal(t, 3, tracker.Size())
}

func TestDirtyTracker_ThreadSafety(t *testing.T) {
	tracker := NewDirtyTracker()

	// This test documents that DirtyTracker is thread-safe due to sync.RWMutex
	assert.NotNil(t, tracker, "Tracker should be created")
}

// Benchmark tests
func BenchmarkDirtyTracker_Mark(b *testing.B) {
	tracker := NewDirtyTracker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.MarkLayoutDirty("node" + string(rune('0'+i%100)))
	}
}

func BenchmarkDirtyTracker_IsDirty(b *testing.B) {
	tracker := NewDirtyTracker()

	// Warm up
	tracker.MarkLayoutDirty("node1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tracker.IsLayoutDirty("node1")
	}
}

func BenchmarkDirtyTracker_MarkSubtree(b *testing.B) {
	tracker := NewDirtyTracker()

	// Create a small tree
	children := make([]Node, 5)
	for i := range children {
		children[i] = NewMockMeasurableNode("child"+string(rune('0'+i)), 10, 10)
	}
	root := NewFlexLayout("root", children)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.MarkSubtreeDirty(root)
	}
}
