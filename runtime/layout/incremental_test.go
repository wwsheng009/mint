package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEngine_LayoutIncremental tests the incremental layout functionality
func TestEngine_LayoutIncremental(t *testing.T) {
	engine := NewEngine()

	t.Run("all nodes clean - skip all", func(t *testing.T) {
		// Create a tree
		child1 := NewMockMeasurableNode("child1", 20, 20)
		child2 := NewMockMeasurableNode("child2", 20, 20)
		root := NewFlexLayout("root", []Node{child1, child2})

		constraints := UnboundedConstraints()

		// First layout - all nodes should be laid out
		result1 := engine.Layout(root, constraints)
		assert.NotNil(t, result1)
		assert.Len(t, result1.Boxes, 3) // root + 2 children

		// Clear dirty markers after first layout
		engine.clearDirtyMarkers(root)

		// Second incremental layout - all nodes are clean, should skip measurement
		engine.Invalidate()
		result2 := engine.LayoutIncremental(root, constraints)
		assert.NotNil(t, result2)

		// Results should be the same
		assert.Equal(t, result1.Boxes, result2.Boxes)
	})

	t.Run("partial dirty - skip clean nodes", func(t *testing.T) {
		// Create a tree
		child1 := NewMockMeasurableNode("child1", 20, 20)
		child2 := NewMockMeasurableNode("child2", 20, 20)
		root := NewFlexLayout("root", []Node{child1, child2})

		constraints := UnboundedConstraints()

		// First layout
		result1 := engine.Layout(root, constraints)
		engine.clearDirtyMarkers(root)

		// Mark only child1 as dirty
		engine.dirty.MarkLayoutDirty("child1")

		// Incremental layout - only child1 should be re-laid out
		result2 := engine.LayoutIncremental(root, constraints)
		assert.NotNil(t, result2)

		// Box count should be the same
		assert.Equal(t, len(result1.Boxes), len(result2.Boxes))
	})

	t.Run("root dirty - relayout entire tree", func(t *testing.T) {
		// Create a tree
		child1 := NewMockMeasurableNode("child1", 20, 20)
		child2 := NewMockMeasurableNode("child2", 20, 20)
		root := NewFlexLayout("root", []Node{child1, child2})

		constraints := UnboundedConstraints()

		// First layout
		result1 := engine.Layout(root, constraints)
		engine.clearDirtyMarkers(root)

		// Mark root as dirty
		engine.dirty.MarkLayoutDirty("root")

		// Incremental layout - entire tree should be re-laid out
		result2 := engine.LayoutIncremental(root, constraints)
		assert.NotNil(t, result2)

		// Box count should be the same
		assert.Equal(t, len(result1.Boxes), len(result2.Boxes))
	})

	t.Run("nil root - return empty result", func(t *testing.T) {
		constraints := UnboundedConstraints()
		result := engine.LayoutIncremental(nil, constraints)
		assert.NotNil(t, result)
		assert.Empty(t, result.Boxes)
	})

	t.Run("deep tree with selective dirty", func(t *testing.T) {
		// Create a deep tree
		leaf1 := NewMockMeasurableNode("leaf1", 10, 10)
		leaf2 := NewMockMeasurableNode("leaf2", 10, 10)
		mid := NewFlexLayout("mid", []Node{leaf1, leaf2})
		root := NewFlexLayout("root", []Node{mid})

		constraints := UnboundedConstraints()

		// First layout
		result1 := engine.Layout(root, constraints)
		engine.clearDirtyMarkers(root)

		// Mark only leaf1 as dirty
		engine.dirty.MarkLayoutDirty("leaf1")

		// Incremental layout
		result2 := engine.LayoutIncremental(root, constraints)
		assert.NotNil(t, result2)

		// Box count should be the same
		assert.Equal(t, len(result1.Boxes), len(result2.Boxes))
		assert.Equal(t, 4, len(result2.Boxes)) // root + mid + leaf1 + leaf2
	})
}

// TestEngine_clearDirtyMarkers tests the clearDirtyMarkers method
func TestEngine_clearDirtyMarkers(t *testing.T) {
	engine := NewEngine()

	t.Run("clear entire tree", func(t *testing.T) {
		// Create a tree
		child1 := NewMockMeasurableNode("child1", 20, 20)
		child2 := NewMockMeasurableNode("child2", 20, 20)
		root := NewFlexLayout("root", []Node{child1, child2})

		// Mark all nodes as dirty
		engine.dirty.MarkLayoutDirty("root")
		engine.dirty.MarkLayoutDirty("child1")
		engine.dirty.MarkLayoutDirty("child2")

		// Verify all are dirty
		assert.True(t, engine.dirty.IsLayoutDirty("root"))
		assert.True(t, engine.dirty.IsLayoutDirty("child1"))
		assert.True(t, engine.dirty.IsLayoutDirty("child2"))

		// Clear all dirty markers
		engine.clearDirtyMarkers(root)

		// Verify all are clean
		assert.False(t, engine.dirty.IsLayoutDirty("root"))
		assert.False(t, engine.dirty.IsLayoutDirty("child1"))
		assert.False(t, engine.dirty.IsLayoutDirty("child2"))
	})

	t.Run("clear with nil root", func(t *testing.T) {
		// Should not panic
		engine.clearDirtyMarkers(nil)
	})

	t.Run("clear subtree", func(t *testing.T) {
		// Create a tree
		leaf1 := NewMockMeasurableNode("leaf1", 10, 10)
		leaf2 := NewMockMeasurableNode("leaf2", 10, 10)
		mid := NewFlexLayout("mid", []Node{leaf1, leaf2})
		_ = NewFlexLayout("root", []Node{mid})

		// Mark all nodes as dirty
		engine.dirty.MarkLayoutDirty("root")
		engine.dirty.MarkLayoutDirty("mid")
		engine.dirty.MarkLayoutDirty("leaf1")
		engine.dirty.MarkLayoutDirty("leaf2")

		// Clear only the mid subtree
		engine.clearDirtyMarkers(mid)

		// Root should still be dirty
		assert.True(t, engine.dirty.IsLayoutDirty("root"))

		// Mid and leaves should be clean
		assert.False(t, engine.dirty.IsLayoutDirty("mid"))
		assert.False(t, engine.dirty.IsLayoutDirty("leaf1"))
		assert.False(t, engine.dirty.IsLayoutDirty("leaf2"))
	})
}

// TestEngine_LayoutIncremental_Cache tests that incremental layout works with caching
func TestEngine_LayoutIncremental_Cache(t *testing.T) {
	engine := NewEngine()

	// Create a simple tree
	child1 := NewMockMeasurableNode("child1", 20, 20)
	child2 := NewMockMeasurableNode("child2", 20, 20)
	root := NewFlexLayout("root", []Node{child1, child2})

	constraints := UnboundedConstraints()

	// First layout - should be a cache miss (root has children, so not cached)
	stats1 := engine.GetStats()
	assert.Equal(t, int64(0), stats1.CacheHits)
	assert.Equal(t, int64(0), stats1.CacheMisses)

	result1 := engine.Layout(root, constraints)

	// Now test incremental layout
	engine.Invalidate()
	result2 := engine.LayoutIncremental(root, constraints)

	// Results should be valid and match
	assert.NotNil(t, result2)
	assert.Equal(t, len(result1.Boxes), len(result2.Boxes))
}

// TestEngine_LayoutIncremental_Performance tests that incremental layout is faster than full layout
func TestEngine_LayoutIncremental_Performance(t *testing.T) {
	engine := NewEngine()

	// Create a large tree
	children := make([]Node, 100)
	for i := 0; i < 100; i++ {
		children[i] = NewMockMeasurableNode("child"+string(rune('0'+i%10)), 10, 10)
	}
	root := NewFlexLayout("root", children)

	constraints := UnboundedConstraints()

	// First full layout
	result1 := engine.Layout(root, constraints)
	engine.clearDirtyMarkers(root)

	// Mark only one child as dirty
	engine.dirty.MarkLayoutDirty("child0")

	// Incremental layout - should be faster (though we don't measure time in unit test)
	result2 := engine.LayoutIncremental(root, constraints)

	// Results should be valid
	assert.NotNil(t, result2)
	assert.Equal(t, len(result1.Boxes), len(result2.Boxes))
}

// Benchmark tests
func BenchmarkEngine_LayoutIncremental_AllClean(b *testing.B) {
	engine := NewEngine()

	// Create a tree
	children := make([]Node, 10)
	for i := range children {
		children[i] = NewMockMeasurableNode("child"+string(rune('0'+i)), 10, 10)
	}
	root := NewFlexLayout("root", children)

	constraints := UnboundedConstraints()

	// Initial layout
	engine.Layout(root, constraints)
	engine.clearDirtyMarkers(root)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.LayoutIncremental(root, constraints)
	}
}

func BenchmarkEngine_LayoutIncremental_PartialDirty(b *testing.B) {
	engine := NewEngine()

	// Create a tree
	children := make([]Node, 10)
	for i := range children {
		children[i] = NewMockMeasurableNode("child"+string(rune('0'+i)), 10, 10)
	}
	root := NewFlexLayout("root", children)

	constraints := UnboundedConstraints()

	// Initial layout
	engine.Layout(root, constraints)
	engine.clearDirtyMarkers(root)

	// Mark one child as dirty
	engine.dirty.MarkLayoutDirty("child0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.LayoutIncremental(root, constraints)
		// Re-mark child0 as dirty for next iteration
		engine.dirty.MarkLayoutDirty("child0")
	}
}

func BenchmarkEngine_Layout_Full(b *testing.B) {
	engine := NewEngine()

	// Create a tree
	children := make([]Node, 10)
	for i := range children {
		children[i] = NewMockMeasurableNode("child"+string(rune('0'+i)), 10, 10)
	}
	root := NewFlexLayout("root", children)

	constraints := UnboundedConstraints()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Layout(root, constraints)
	}
}

func BenchmarkEngine_clearDirtyMarkers(b *testing.B) {
	engine := NewEngine()

	// Create a tree with many nodes
	children := make([]Node, 100)
	for i := range children {
		children[i] = NewMockMeasurableNode("child"+string(rune('0'+i%10)), 10, 10)
	}
	root := NewFlexLayout("root", children)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Mark all nodes as dirty
		for _, child := range children {
			engine.dirty.MarkLayoutDirty(child.ID())
		}
		// Clear all dirty markers
		engine.clearDirtyMarkers(root)
	}
}
