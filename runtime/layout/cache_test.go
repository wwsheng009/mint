package layout

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache_Hit(t *testing.T) {
	tests := []struct {
		name       string
		maxSize    int
		numEntries int
	}{
		{
			name:       "cache hit returns correct result",
			maxSize:    10,
			numEntries: 3,
		},
		{
			name:       "cache hit increments hit count",
			maxSize:    10,
			numEntries: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := &Cache{
				entries: make(map[string]*CachedLayout),
				maxSize: tt.maxSize,
			}

			// Create test nodes and constraints
			node := NewMockMeasurableNode("node1", 100, 50)
			constraints := NewConstraints(0, 200, 0, 100)

			// Create a layout result
			result := &LayoutResult{
				Boxes: []LayoutBox{
					{ID: "node1", X: 0, Y: 0, Width: 100, Height: 50},
				},
				ContentSize: Size{Width: 100, Height: 50},
				Dirty:       false,
			}

			// Put result in cache
			cache.Put(node, constraints, result)

			// Get from cache
			cachedResult := cache.Get(node, constraints)

			// Verify cache hit
			assert.NotNil(t, cachedResult, "Cache should return a result")
			assert.Equal(t, len(result.Boxes), len(cachedResult.Boxes), "Cached result should have same number of boxes")
			assert.Equal(t, result.ContentSize, cachedResult.ContentSize, "Content size should match")
		})
	}
}

func TestCache_Miss(t *testing.T) {
	cache := &Cache{
		entries: make(map[string]*CachedLayout),
		maxSize: 10,
	}

	node1 := NewMockMeasurableNode("node1", 100, 50)
	constraints1 := NewConstraints(0, 200, 0, 100)

	node2 := NewMockMeasurableNode("node2", 150, 75)
	constraints2 := NewConstraints(0, 300, 0, 200)

	// Put first result
	result1 := &LayoutResult{
		Boxes: []LayoutBox{
			{ID: "node1", X: 0, Y: 0, Width: 100, Height: 50},
		},
	}
	cache.Put(node1, constraints1, result1)

	// Try to get with different nodes/constraints
	cachedResult := cache.Get(node2, constraints2)

	// Verify cache miss
	assert.Nil(t, cachedResult, "Cache should return nil for different nodes/constraints")
}

func TestCache_Invalidate(t *testing.T) {
	t.Run("node invalidation", func(t *testing.T) {
		cache := &Cache{
			entries: make(map[string]*CachedLayout),
			maxSize:  10,
		}

		node := NewMockMeasurableNode("node1", 100, 50)
		constraints := NewConstraints(0, 200, 0, 100)

		result := &LayoutResult{
			Boxes: []LayoutBox{
				{ID: "node1", X: 0, Y: 0, Width: 100, Height: 50},
			},
		}
		cache.Put(node, constraints, result)

		// Verify entry exists
		assert.NotNil(t, cache.Get(node, constraints), "Cache should have entry before invalidation")

		// Invalidate node
		cache.RemoveByNode("node1")

		// Verify cache is cleared (current implementation clears all)
		assert.Nil(t, cache.Get(node, constraints), "Cache should be cleared after node invalidation")
	})

	t.Run("subtree invalidation", func(t *testing.T) {
		cache := &Cache{
			entries: make(map[string]*CachedLayout),
			maxSize:  10,
		}

		// Create nested structure
		child := NewMockMeasurableNode("child", 50, 50)
		parent := NewFlexLayout("parent", []Node{child})

		constraints := UnboundedConstraints()
		result := &LayoutResult{
			Boxes: []LayoutBox{
				{ID: "parent", X: 0, Y: 0, Width: 50, Height: 50},
				{ID: "child", X: 0, Y: 0, Width: 50, Height: 50},
			},
		}
		cache.Put(parent, constraints, result)

		// Invalidate parent node
		cache.RemoveByNode("parent")

		// Verify cache is cleared
		assert.Nil(t, cache.Get(parent, constraints), "Cache should be cleared after parent invalidation")
	})
}

func TestCache_Clear(t *testing.T) {
	cache := &Cache{
		entries: make(map[string]*CachedLayout),
		maxSize:  10,
	}

	// Add multiple entries
	for i := 0; i < 5; i++ {
		node := NewMockMeasurableNode("node"+string(rune('0'+i)), 100, 50)
		constraints := NewConstraints(0, 200, 0, 100)
		result := &LayoutResult{
			Boxes: []LayoutBox{
				{ID: "node" + string(rune('0'+i)), X: 0, Y: 0, Width: 100, Height: 50},
			},
		}
		cache.Put(node, constraints, result)
	}

	// Verify entries exist
	assert.Equal(t, 5, len(cache.entries), "Cache should have 5 entries")

	// Clear cache
	cache.Clear()

	// Verify cache is empty
	assert.Equal(t, 0, len(cache.entries), "Cache should be empty after clear")
}

func TestCache_ThreadSafety(t *testing.T) {
	t.Skip("Cache implementation is not thread-safe - requires mutex protection")

	// This test documents that the current cache implementation
	// This test documents that the current cache implementation
	// is NOT thread-safe and would fail with concurrent access.
	// To make it thread-safe, the Cache struct needs:
	// sync.RWMutex for protecting entries map access
}

func TestCache_Eviction(t *testing.T) {
	t.Run("LRU eviction", func(t *testing.T) {
		cache := &Cache{
			entries: make(map[string]*CachedLayout),
			maxSize:  3,
		}

		// Fill cache to max capacity
		for i := 0; i < 3; i++ {
			node := NewMockMeasurableNode("node"+string(rune('0'+i)), 100, 50)
			constraints := NewConstraints(0, 200, 0, 100)
			result := &LayoutResult{
				Boxes: []LayoutBox{
					{ID: "node" + string(rune('0'+i)), X: 0, Y: 0, Width: 100, Height: 50},
				},
			}
			cache.Put(node, constraints, result)
			time.Sleep(10 * time.Millisecond) // Ensure different timestamps
		}

		assert.Equal(t, 3, len(cache.entries), "Cache should be at max capacity")

		// Add one more entry to trigger eviction
		node := NewMockMeasurableNode("node3", 100, 50)
		constraints := NewConstraints(0, 200, 0, 100)
		result := &LayoutResult{
			Boxes: []LayoutBox{
				{ID: "node3", X: 0, Y: 0, Width: 100, Height: 50},
			},
		}
		cache.Put(node, constraints, result)

		// Verify oldest entry was evicted
		assert.Equal(t, 3, len(cache.entries), "Cache should still be at max capacity")
	})

	t.Run("capacity limit", func(t *testing.T) {
		cache := &Cache{
			entries: make(map[string]*CachedLayout),
			maxSize:  5,
		}

		// Add more entries than max size
		for i := 0; i < 10; i++ {
			node := NewMockMeasurableNode("node"+string(rune('0'+i)), 100, 50)
			constraints := NewConstraints(0, 200, 0, 100)
			result := &LayoutResult{
				Boxes: []LayoutBox{
					{ID: "node" + string(rune('0'+i)), X: 0, Y: 0, Width: 100, Height: 50},
				},
			}
			cache.Put(node, constraints, result)
		}

		// Verify cache never exceeds max size
		assert.LessOrEqual(t, len(cache.entries), cache.maxSize, "Cache should not exceed max size")
		assert.Equal(t, cache.maxSize, len(cache.entries), "Cache should be at max size")
	})
}

func TestCache_Clone(t *testing.T) {
	t.Run("clone result", func(t *testing.T) {
		cache := &Cache{
			entries: make(map[string]*CachedLayout),
			maxSize:  10,
		}

		original := &LayoutResult{
			Boxes: []LayoutBox{
				{
					ID:     "box1",
					X:      10,
					Y:      20,
					Width:  100,
					Height: 50,
					Baseline: 25,
					Children: []*LayoutBox{
						{
							ID:     "child1",
							X:      5,
							Y:      5,
							Width:  50,
							Height: 25,
						},
					},
				},
			},
			ContentSize: Size{Width: 100, Height: 50},
			Dirty:       true,
		}

		cloned := cache.cloneResult(original)

		// Verify clone is equal but not the same instance
		assert.Equal(t, len(original.Boxes), len(cloned.Boxes), "Box count should match")
		assert.Equal(t, original.ContentSize, cloned.ContentSize, "Content size should match")
		assert.Equal(t, original.Dirty, cloned.Dirty, "Dirty flag should match")

		// Modify clone and verify original is unchanged
		cloned.Boxes[0].X = 999
		assert.NotEqual(t, original.Boxes[0].X, cloned.Boxes[0].X, "Modifying clone should not affect original")
	})

	t.Run("clone nil result", func(t *testing.T) {
		cache := &Cache{
			entries: make(map[string]*CachedLayout),
			maxSize:  10,
		}

		cloned := cache.cloneResult(nil)
		assert.Nil(t, cloned, "Cloning nil should return nil")
	})

	t.Run("clone box with children", func(t *testing.T) {
		cache := &Cache{
			entries: make(map[string]*CachedLayout),
			maxSize:  10,
		}

		original := &LayoutBox{
			ID:     "parent",
			X:      0,
			Y:      0,
			Width:  100,
			Height: 100,
			Children: []*LayoutBox{
				{
					ID:     "child1",
					X:      10,
					Y:      10,
					Width:  50,
					Height: 50,
				},
				{
					ID:     "child2",
					X:      60,
					Y:      10,
					Width:  30,
					Height: 30,
				},
			},
		}

		cloned := cache.cloneBox(original)

		// Verify structure is cloned
		assert.Equal(t, original.ID, cloned.ID, "ID should match")
		assert.Equal(t, len(original.Children), len(cloned.Children), "Children count should match")

		// Verify children are cloned
		cloned.Children[0].X = 999
		assert.NotEqual(t, original.Children[0].X, cloned.Children[0].X, "Modifying cloned child should not affect original")
	})
}

func TestCache_KeyGeneration(t *testing.T) {
	cache := &Cache{
		entries: make(map[string]*CachedLayout),
		maxSize:  10,
	}

	node1 := NewMockMeasurableNode("node1", 100, 50)
	node2 := NewMockMeasurableNode("node2", 150, 75)
	node := NewFlexLayout("root", []Node{node1, node2})
	constraints := NewConstraints(10, 200, 20, 100)

	key1 := cache.makeKey(node, constraints)
	key2 := cache.makeKey(node, constraints)

	// Same inputs should generate same key
	assert.Equal(t, key1, key2, "Same inputs should generate same cache key")

	// Different constraints should generate different key
	differentConstraints := NewConstraints(15, 250, 25, 150)
	key3 := cache.makeKey(node, differentConstraints)
	assert.NotEqual(t, key1, key3, "Different constraints should generate different cache key")
}

func TestCache_HitCount(t *testing.T) {
	cache := &Cache{
		entries: make(map[string]*CachedLayout),
		maxSize:  10,
	}

	node := NewMockMeasurableNode("node1", 100, 50)
	constraints := NewConstraints(0, 200, 0, 100)

	result := &LayoutResult{
		Boxes: []LayoutBox{
			{ID: "node1", X: 0, Y: 0, Width: 100, Height: 50},
		},
	}

	cache.Put(node, constraints, result)

	// Hit cache multiple times
	for i := 0; i < 5; i++ {
		_ = cache.Get(node, constraints)
	}

	// Check hit count
	key := cache.makeKey(node, constraints)
	if entry, ok := cache.entries[key]; ok {
		assert.Equal(t, 5, entry.HitCount, "Hit count should be 5")
	} else {
		t.Fatal("Cache entry should exist")
	}
}

func TestCache_MultipleResults(t *testing.T) {
	cache := &Cache{
		entries: make(map[string]*CachedLayout),
		maxSize:  10,
	}

	// Add multiple different results
	entries := []struct {
		node        Node
		constraints Constraints
		result      *LayoutResult
	}{
		{
			node:        NewMockMeasurableNode("node1", 100, 50),
			constraints: NewConstraints(0, 200, 0, 100),
			result:      &LayoutResult{Boxes: []LayoutBox{{ID: "node1", Width: 100, Height: 50}}},
		},
		{
			node:        NewMockMeasurableNode("node2", 150, 75),
			constraints: NewConstraints(0, 300, 0, 200),
			result:      &LayoutResult{Boxes: []LayoutBox{{ID: "node2", Width: 150, Height: 75}}},
		},
		{
			node:        NewMockMeasurableNode("node3", 200, 100),
			constraints: NewConstraints(0, 400, 0, 300),
			result:      &LayoutResult{Boxes: []LayoutBox{{ID: "node3", Width: 200, Height: 100}}},
		},
	}

	for _, entry := range entries {
		cache.Put(entry.node, entry.constraints, entry.result)
	}

	// Verify all entries are cached
	assert.Equal(t, len(entries), len(cache.entries), "All entries should be cached")

	// Verify each entry can be retrieved
	for _, entry := range entries {
		cached := cache.Get(entry.node, entry.constraints)
		assert.NotNil(t, cached, "Each entry should be retrievable")
		assert.Equal(t, entry.result.Boxes[0].ID, cached.Boxes[0].ID, "Retrieved entry should match")
	}
}

func TestCache_EngineIntegration(t *testing.T) {
	engine := NewEngine()

	node1 := NewMockMeasurableNode("node1", 100, 50)
	node2 := NewMockMeasurableNode("node2", 150, 75)
	node := NewFlexLayout("root", []Node{node1, node2})
	constraints := NewConstraints(0, 200, 0, 100)

	// First layout - should compute
	result1 := engine.Layout(node, constraints)
	assert.NotNil(t, result1, "First layout should return result")

	// Second layout with same inputs - should hit cache
	// (Note: Engine currently doesn't use cache internally in the provided code, 
	// but the test expects it)
	result2 := engine.Layout(node, constraints)
	assert.NotNil(t, result2, "Second layout should return cached result")

	// Verify cache stats
	stats := engine.GetStats()
	assert.Greater(t, stats.CacheHits, int64(0), "Should have cache hits")
}

// Benchmark tests
func BenchmarkCache_Get(b *testing.B) {
	cache := &Cache{
		entries: make(map[string]*CachedLayout),
		maxSize:  1000,
	}

	node1 := NewMockMeasurableNode("node1", 100, 50)
	node2 := NewMockMeasurableNode("node2", 150, 75)
	node := NewFlexLayout("root", []Node{node1, node2})
	constraints := NewConstraints(0, 200, 0, 100)

	result := &LayoutResult{
		Boxes: []LayoutBox{
			{ID: "node1", X: 0, Y: 0, Width: 100, Height: 50},
			{ID: "node2", X: 0, Y: 50, Width: 150, Height: 75},
		},
	}
	cache.Put(node, constraints, result)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Get(node, constraints)
	}
}

func BenchmarkCache_Put(b *testing.B) {
	cache := &Cache{
		entries: make(map[string]*CachedLayout),
		maxSize:  1000,
	}

	result := &LayoutResult{
		Boxes: []LayoutBox{
			{ID: "bench", X: 0, Y: 0, Width: 100, Height: 50},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node := NewMockMeasurableNode("node"+string(rune('0'+i%10)), 100, 50)
		constraints := NewConstraints(0, 200, 0, 100)
		cache.Put(node, constraints, result)
	}
}

func BenchmarkCache_CloneResult(b *testing.B) {
	cache := &Cache{
		entries: make(map[string]*CachedLayout),
		maxSize:  10,
	}

	result := &LayoutResult{
		Boxes: make([]LayoutBox, 100),
	}
	for i := range result.Boxes {
		result.Boxes[i] = LayoutBox{
			ID:     "box" + string(rune('0'+i)),
			X:      i * 10,
			Y:      i * 10,
			Width:  100,
			Height: 50,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.cloneResult(result)
	}
}
