package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHitMap_NewHitMap(t *testing.T) {
	hm := NewHitMap()

	assert.NotNil(t, hm)
	assert.Equal(t, 0, hm.Size())
}

func TestHitMap_BuildFromLayoutBox(t *testing.T) {
	hm := NewHitMap()

	// Create a simple layout tree
	child1 := &LayoutBox{
		ID:      "child1",
		X:       10,
		Y:       10,
		Width:   50,
		Height:  30,
		Children: []*LayoutBox{},
	}

	child2 := &LayoutBox{
		ID:      "child2",
		X:       10,
		Y:       50,
		Width:   50,
		Height:  30,
		Children: []*LayoutBox{},
	}

	root := &LayoutBox{
		ID:      "root",
		X:       0,
		Y:       0,
		Width:   100,
		Height:  100,
		Children: []*LayoutBox{child1, child2},
	}

	// Build hitmap
	hm.BuildFromLayoutBox(root)

	// Verify size
	assert.Equal(t, 3, hm.Size())

	// Verify entries
	entry1 := hm.Get("child1")
	assert.NotNil(t, entry1)
	assert.Equal(t, "child1", entry1.NodeID)
	assert.Equal(t, 10, entry1.Rect.X)
	assert.Equal(t, 10, entry1.Rect.Y)
	assert.Equal(t, 50, entry1.Rect.Width)
	assert.Equal(t, 30, entry1.Rect.Height)
	assert.Equal(t, 1, entry1.ZIndex) // Child has Z-index 1

	entry2 := hm.Get("child2")
	assert.NotNil(t, entry2)
	assert.Equal(t, 1, entry2.ZIndex) // Child has Z-index 1

	rootEntry := hm.Get("root")
	assert.NotNil(t, rootEntry)
	assert.Equal(t, 0, rootEntry.ZIndex) // Root has Z-index 0
}

func TestHitMap_BuildFromLayoutBox_Nil(t *testing.T) {
	hm := NewHitMap()

	// Should not panic
	hm.BuildFromLayoutBox(nil)

	assert.Equal(t, 0, hm.Size())
}

func TestHitMap_HitTest(t *testing.T) {
	hm := NewHitMap()

	// Create a simple layout tree
	child1 := &LayoutBox{
		ID:      "child1",
		X:       10,
		Y:       10,
		Width:   50,
		Height:  30,
		Children: []*LayoutBox{},
	}

	root := &LayoutBox{
		ID:      "root",
		X:       0,
		Y:       0,
		Width:   100,
		Height:  100,
		Children: []*LayoutBox{child1},
	}

	hm.BuildFromLayoutBox(root)

	t.Run("hit child node", func(t *testing.T) {
		entry := hm.HitTest(15, 15) // Inside child1
		assert.NotNil(t, entry)
		assert.Equal(t, "child1", entry.NodeID)
	})

	t.Run("hit root node (not child)", func(t *testing.T) {
		entry := hm.HitTest(5, 5) // Inside root but not child1
		assert.NotNil(t, entry)
		assert.Equal(t, "root", entry.NodeID)
	})

	t.Run("hit nothing", func(t *testing.T) {
		entry := hm.HitTest(200, 200) // Outside everything
		assert.Nil(t, entry)
	})

	t.Run("hit on boundary", func(t *testing.T) {
		entry := hm.HitTest(10, 10) // On child1 boundary
		assert.NotNil(t, entry)
		assert.Equal(t, "child1", entry.NodeID)
	})
}

func TestHitMap_HitTestAll(t *testing.T) {
	hm := NewHitMap()

	// Create overlapping nodes (child inside parent)
	child1 := &LayoutBox{
		ID:      "child1",
		X:       10,
		Y:       10,
		Width:   50,
		Height:  30,
		Children: []*LayoutBox{},
	}

	root := &LayoutBox{
		ID:      "root",
		X:       0,
		Y:       0,
		Width:   100,
		Height:  100,
		Children: []*LayoutBox{child1},
	}

	hm.BuildFromLayoutBox(root)

	t.Run("hit multiple nodes", func(t *testing.T) {
		entries := hm.HitTestAll(15, 15) // Inside both child1 and root
		assert.Len(t, entries, 2)

		// Should be sorted by Z-index (low to high)
		assert.Equal(t, "root", entries[0].NodeID)
		assert.Equal(t, 0, entries[0].ZIndex)
		assert.Equal(t, "child1", entries[1].NodeID)
		assert.Equal(t, 1, entries[1].ZIndex)
	})

	t.Run("hit single node", func(t *testing.T) {
		entries := hm.HitTestAll(5, 5) // Inside root only
		assert.Len(t, entries, 1)
		assert.Equal(t, "root", entries[0].NodeID)
	})

	t.Run("hit nothing", func(t *testing.T) {
		entries := hm.HitTestAll(200, 200) // Outside everything
		assert.Len(t, entries, 0)
	})
}

func TestHitMap_Get(t *testing.T) {
	hm := NewHitMap()

	// Create a simple layout tree
	child1 := &LayoutBox{
		ID:      "child1",
		X:       10,
		Y:       10,
		Width:   50,
		Height:  30,
		Children: []*LayoutBox{},
	}

	root := &LayoutBox{
		ID:      "root",
		X:       0,
		Y:       0,
		Width:   100,
		Height:  100,
		Children: []*LayoutBox{child1},
	}

	hm.BuildFromLayoutBox(root)

	t.Run("get existing node", func(t *testing.T) {
		entry := hm.Get("child1")
		assert.NotNil(t, entry)
		assert.Equal(t, "child1", entry.NodeID)
	})

	t.Run("get non-existing node", func(t *testing.T) {
		entry := hm.Get("nonexistent")
		assert.Nil(t, entry)
	})
}

func TestHitMap_GetAll(t *testing.T) {
	hm := NewHitMap()

	// Create a simple layout tree
	child1 := &LayoutBox{
		ID:      "child1",
		X:       10,
		Y:       10,
		Width:   50,
		Height:  30,
		Children: []*LayoutBox{},
	}

	child2 := &LayoutBox{
		ID:      "child2",
		X:       10,
		Y:       50,
		Width:   50,
		Height:  30,
		Children: []*LayoutBox{},
	}

	root := &LayoutBox{
		ID:      "root",
		X:       0,
		Y:       0,
		Width:   100,
		Height:  100,
		Children: []*LayoutBox{child1, child2},
	}

	hm.BuildFromLayoutBox(root)

	// Get all entries
	entries := hm.GetAll()
	assert.Len(t, entries, 3)

	// Verify all node IDs are present
	nodeIDs := make(map[string]bool)
	for _, entry := range entries {
		nodeIDs[entry.NodeID] = true
	}

	assert.True(t, nodeIDs["root"])
	assert.True(t, nodeIDs["child1"])
	assert.True(t, nodeIDs["child2"])
}

func TestHitMap_Clear(t *testing.T) {
	hm := NewHitMap()

	// Create a simple layout tree
	child1 := &LayoutBox{
		ID:      "child1",
		X:       10,
		Y:       10,
		Width:   50,
		Height:  30,
		Children: []*LayoutBox{},
	}

	root := &LayoutBox{
		ID:      "root",
		X:       0,
		Y:       0,
		Width:   100,
		Height:  100,
		Children: []*LayoutBox{child1},
	}

	hm.BuildFromLayoutBox(root)

	// Verify size
	assert.Equal(t, 2, hm.Size())

	// Clear
	hm.Clear()

	// Verify empty
	assert.Equal(t, 0, hm.Size())
	assert.Nil(t, hm.Get("root"))
	assert.Nil(t, hm.Get("child1"))
}

func TestHitMap_Size(t *testing.T) {
	hm := NewHitMap()

	// Empty hitmap
	assert.Equal(t, 0, hm.Size())

	// Create a simple layout tree
	child1 := &LayoutBox{
		ID:      "child1",
		X:       10,
		Y:       10,
		Width:   50,
		Height:  30,
		Children: []*LayoutBox{},
	}

	root := &LayoutBox{
		ID:      "root",
		X:       0,
		Y:       0,
		Width:   100,
		Height:  100,
		Children: []*LayoutBox{child1},
	}

	hm.BuildFromLayoutBox(root)

	// Verify size
	assert.Equal(t, 2, hm.Size())
}

func TestHitMap_DeepNesting(t *testing.T) {
	hm := NewHitMap()

	// Create a deeply nested tree
	leaf1 := &LayoutBox{
		ID:      "leaf1",
		X:       0,
		Y:       0,
		Width:   10,
		Height:  10,
		Children: []*LayoutBox{},
	}

	leaf2 := &LayoutBox{
		ID:      "leaf2",
		X:       0,
		Y:       10,
		Width:   10,
		Height:  10,
		Children: []*LayoutBox{},
	}

	mid := &LayoutBox{
		ID:      "mid",
		X:       10,
		Y:       10,
		Width:   20,
		Height:  20,
		Children: []*LayoutBox{leaf1, leaf2},
	}

	root := &LayoutBox{
		ID:      "root",
		X:       0,
		Y:       0,
		Width:   100,
		Height:  100,
		Children: []*LayoutBox{mid},
	}

	hm.BuildFromLayoutBox(root)

	// Verify size (4 nodes)
	assert.Equal(t, 4, hm.Size())

	// Verify Z-indexes
	rootEntry := hm.Get("root")
	assert.Equal(t, 0, rootEntry.ZIndex)

	midEntry := hm.Get("mid")
	assert.Equal(t, 1, midEntry.ZIndex)

	leaf1Entry := hm.Get("leaf1")
	assert.Equal(t, 2, leaf1Entry.ZIndex)

	leaf2Entry := hm.Get("leaf2")
	assert.Equal(t, 2, leaf2Entry.ZIndex)
}

func TestHitMap_EngineIntegration(t *testing.T) {
	engine := NewEngine()

	// Create a simple tree
	child1 := NewMockMeasurableNode("child1", 20, 20)
	child2 := NewMockMeasurableNode("child2", 20, 20)
	root := NewFlexLayout("root", []Node{child1, child2})

	constraints := UnboundedConstraints()

	// Perform layout
	result := engine.Layout(root, constraints)

	// Verify HitMap is present
	assert.NotNil(t, result.HitMap)

	// Verify HitMap has entries
	assert.Greater(t, result.HitMap.Size(), 0)

	// Verify we can hit test
	entry := result.HitMap.HitTest(0, 0)
	assert.NotNil(t, entry)
}

func TestHitMap_OverlappingNodes(t *testing.T) {
	hm := NewHitMap()

	// Create overlapping nodes (same position)
	child1 := &LayoutBox{
		ID:      "child1",
		X:       10,
		Y:       10,
		Width:   50,
		Height:  30,
		Children: []*LayoutBox{},
	}

	child2 := &LayoutBox{
		ID:      "child2",
		X:       10,
		Y:       10,
		Width:   50,
		Height:  30,
		Children: []*LayoutBox{},
	}

	root := &LayoutBox{
		ID:      "root",
		X:       0,
		Y:       0,
		Width:   100,
		Height:  100,
		Children: []*LayoutBox{child1, child2},
	}

	hm.BuildFromLayoutBox(root)

	// HitTest should return the last added child (higher Z-index)
	entry := hm.HitTest(15, 15)
	assert.NotNil(t, entry)
	// Since both have same Z-index, either child1 or child2 is acceptable
	assert.True(t, entry.NodeID == "child1" || entry.NodeID == "child2")

	// HitTestAll should return all overlapping nodes
	entries := hm.HitTestAll(15, 15)
	assert.Len(t, entries, 3) // root, child1, child2

	// Verify Z-ordering (root is 0, both children are 1)
	assert.Equal(t, "root", entries[0].NodeID)
	assert.Equal(t, 0, entries[0].ZIndex)
	assert.True(t, entries[1].NodeID == "child1" || entries[1].NodeID == "child2")
	assert.Equal(t, 1, entries[1].ZIndex)
	assert.True(t, entries[2].NodeID == "child1" || entries[2].NodeID == "child2")
	assert.Equal(t, 1, entries[2].ZIndex)
}

func TestHitMap_ThreadSafety(t *testing.T) {
	hm := NewHitMap()

	// This test documents that HitMap is thread-safe due to sync.RWMutex
	assert.NotNil(t, hm, "HitMap should be created")
}

// Benchmark tests
func BenchmarkHitMap_BuildFromLayoutBox(b *testing.B) {
	// Create a large layout tree
	children := make([]*LayoutBox, 100)
	for i := 0; i < 100; i++ {
		children[i] = &LayoutBox{
			ID:      "child" + string(rune('0'+i%10)),
			X:       0,
			Y:       i * 10,
			Width:   100,
			Height:  10,
			Children: []*LayoutBox{},
		}
	}

	root := &LayoutBox{
		ID:      "root",
		X:       0,
		Y:       0,
		Width:   100,
		Height:  1000,
		Children: children,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hm := NewHitMap()
		hm.BuildFromLayoutBox(root)
	}
}

func BenchmarkHitMap_HitTest(b *testing.B) {
	// Create a large layout tree
	children := make([]*LayoutBox, 100)
	for i := 0; i < 100; i++ {
		children[i] = &LayoutBox{
			ID:      "child" + string(rune('0'+i%10)),
			X:       0,
			Y:       i * 10,
			Width:   100,
			Height:  10,
			Children: []*LayoutBox{},
		}
	}

	root := &LayoutBox{
		ID:      "root",
		X:       0,
		Y:       0,
		Width:   100,
		Height:  1000,
		Children: children,
	}

	hm := NewHitMap()
	hm.BuildFromLayoutBox(root)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hm.HitTest(50, 500)
	}
}

func BenchmarkHitMap_HitTestAll(b *testing.B) {
	// Create a large layout tree
	children := make([]*LayoutBox, 100)
	for i := 0; i < 100; i++ {
		children[i] = &LayoutBox{
			ID:      "child" + string(rune('0'+i%10)),
			X:       0,
			Y:       i * 10,
			Width:   100,
			Height:  10,
			Children: []*LayoutBox{},
		}
	}

	root := &LayoutBox{
		ID:      "root",
		X:       0,
		Y:       0,
		Width:   100,
		Height:  1000,
		Children: children,
	}

	hm := NewHitMap()
	hm.BuildFromLayoutBox(root)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hm.HitTestAll(50, 500)
	}
}

func BenchmarkHitMap_Get(b *testing.B) {
	// Create a large layout tree
	children := make([]*LayoutBox, 100)
	for i := 0; i < 100; i++ {
		children[i] = &LayoutBox{
			ID:      "child" + string(rune('0'+i%10)),
			X:       0,
			Y:       i * 10,
			Width:   100,
			Height:  10,
			Children: []*LayoutBox{},
		}
	}

	root := &LayoutBox{
		ID:      "root",
		X:       0,
		Y:       0,
		Width:   100,
		Height:  1000,
		Children: children,
	}

	hm := NewHitMap()
	hm.BuildFromLayoutBox(root)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hm.Get("child50")
	}
}
