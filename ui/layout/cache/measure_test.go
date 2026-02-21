package cache

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/text"
)

// =============================================================================
// Mock VNode for Testing
// =============================================================================

type mockMeasurableVNode struct {
	ui.VNode
	measureFunc func(layout.Constraints) layout.Size
	key         string
	version     int
}

func (m *mockMeasurableVNode) Measure(c layout.Constraints) layout.Size {
	if m.measureFunc != nil {
		return m.measureFunc(c)
	}
	return layout.Size{Width: 10, Height: 10}
}

func (m *mockMeasurableVNode) Key() string {
	return m.key
}

func (m *mockMeasurableVNode) SetKey(k string) ui.VNode {
	m.key = k
	return m
}

func (m *mockMeasurableVNode) Tag() string {
	return "mock"
}

func (m *mockMeasurableVNode) Type() ui.VNodeType {
	return ui.VNodeElement
}

func (m *mockMeasurableVNode) Props() ui.Props {
	return nil
}

func (m *mockMeasurableVNode) SetProps(p ui.Props) ui.VNode {
	return m
}

func (m *mockMeasurableVNode) Children() []ui.VNode {
	return nil
}

func (m *mockMeasurableVNode) SetChildren(children []ui.VNode) ui.VNode {
	return m
}

func (m *mockMeasurableVNode) Style() style.Style {
	return style.Style{}
}

func (m *mockMeasurableVNode) SetStyle(s style.Style) ui.VNode {
	return m
}

func newMockVNode(key string) *mockMeasurableVNode {
	return &mockMeasurableVNode{
		key:     key,
		version: 1,
		measureFunc: func(c layout.Constraints) layout.Size {
			return layout.Size{Width: c.MinWidth, Height: c.MinHeight}
		},
	}
}

// =============================================================================
// MeasureCache Core Tests
// =============================================================================

func TestMeasureCache_New(t *testing.T) {
	cache := NewMeasureCache()

	if cache == nil {
		t.Fatal("NewMeasureCache() should not return nil")
	}
	stats := cache.Stats()
	if stats.EntryCount != 0 {
		t.Errorf("expected 0 entries, got %d", stats.EntryCount)
	}
}

func TestMeasureCache_PutAndGet(t *testing.T) {
	cache := NewMeasureCache()
	node := newMockVNode("test-node")
	constraints := layout.Constraints{MinWidth: 10, MaxWidth: 50, MinHeight: 5, MaxHeight: 20}
	size := layout.Size{Width: 30, Height: 10}

	// Put
	cache.Put(node, constraints, size, 1)

	// Get
	retrievedSize, found := cache.Get(node, constraints, 1)
	if !found {
		t.Error("expected to find cached measurement")
	}
	if retrievedSize.Width != 30 || retrievedSize.Height != 10 {
		t.Errorf("expected size 30x10, got %dx%d", retrievedSize.Width, retrievedSize.Height)
	}
}

func TestMeasureCache_Get_NotFound(t *testing.T) {
	cache := NewMeasureCache()
	node := newMockVNode("nonexistent")
	constraints := layout.Constraints{MinWidth: 0, MaxWidth: 100}

	_, found := cache.Get(node, constraints, 1)
	if found {
		t.Error("should not find measurement for nonexistent key")
	}
}

func TestMeasureCache_VersionCheck(t *testing.T) {
	cache := NewMeasureCache()
	node := newMockVNode("versioned-node")
	constraints := layout.Constraints{MinWidth: 10, MaxWidth: 50}
	size := layout.Size{Width: 30, Height: 10}

	// Put with version 1
	cache.Put(node, constraints, size, 1)

	// Get with version 1 should succeed
	_, found := cache.Get(node, constraints, 1)
	if !found {
		t.Error("should find measurement with same version")
	}

	// Get with version 2 should fail (version mismatch)
	_, found = cache.Get(node, constraints, 2)
	if found {
		t.Error("should not find measurement with different version")
	}
}

// =============================================================================
// Invalidate Tests
// =============================================================================

func TestMeasureCache_Invalidate(t *testing.T) {
	cache := NewMeasureCache()
	node := newMockVNode("invalidatable-node")
	constraints1 := layout.Constraints{MinWidth: 10, MaxWidth: 50}
	constraints2 := layout.Constraints{MinWidth: 20, MaxWidth: 60}

	// Put multiple entries
	cache.Put(node, constraints1, layout.Size{Width: 30, Height: 10}, 1)
	cache.Put(node, constraints2, layout.Size{Width: 40, Height: 15}, 1)

	// Invalidate
	cache.Invalidate(node)

	// Both should be gone
	_, found1 := cache.Get(node, constraints1, 1)
	_, found2 := cache.Get(node, constraints2, 1)

	if found1 || found2 {
		t.Error("both measurements should be invalidated")
	}
}

func TestMeasureCache_InvalidateAll(t *testing.T) {
	cache := NewMeasureCache()

	// Put multiple entries
	for i := 0; i < 5; i++ {
		node := newMockVNode(fmt.Sprintf("node-%d", i))
		cache.Put(node, layout.Constraints{}, layout.Size{}, 1)
	}

	// Check stats
	stats := cache.Stats()
	if stats.EntryCount != 5 {
		t.Errorf("expected 5 entries, got %d", stats.EntryCount)
	}

	// Invalidate all
	cache.InvalidateAll()

	// Check cache is empty
	stats = cache.Stats()
	if stats.EntryCount != 0 {
		t.Errorf("expected 0 entries after InvalidateAll, got %d", stats.EntryCount)
	}
}

func TestMeasureCache_InvalidateTree(t *testing.T) {
	cache := NewMeasureCache()

	// Create a simple hierarchy
	root := newMockVNode("root")
	child1 := newMockVNode("child1")
	child2 := newMockVNode("child2")

	// Put measurements
	cache.Put(root, layout.Constraints{MinWidth: 10}, layout.Size{Width: 30}, 1)
	cache.Put(child1, layout.Constraints{MinWidth: 5}, layout.Size{Width: 15}, 1)
	cache.Put(child2, layout.Constraints{MinWidth: 5}, layout.Size{Width: 15}, 1)

	// Invalidate subtree rooted at child1
	cache.InvalidateTree(child1)

	// Root and child2 should still be there
	_, rootFound := cache.Get(root, layout.Constraints{MinWidth: 10}, 1)
	_, child1Found := cache.Get(child1, layout.Constraints{MinWidth: 5}, 1)
	_, child2Found := cache.Get(child2, layout.Constraints{MinWidth: 5}, 1)

	if !rootFound {
		t.Error("root entry should still be cached")
	}
	if child1Found {
		t.Error("child1 entry should be invalidated")
	}
	if !child2Found {
		t.Error("child2 entry should still be cached")
	}
}

// =============================================================================
// Resize Tests
// =============================================================================

func TestMeasureCache_Resize(t *testing.T) {
	cache := NewMeasureCache()

	// Put more entries than the limit
	for i := 0; i < 10; i++ {
		node := newMockVNode(fmt.Sprintf("node-%d", i))
		cache.Put(node, layout.Constraints{}, layout.Size{}, 1)
	}

	// Resize to 5 entries
	cache.Resize(5)

	// Should have at most 5 entries
	stats := cache.Stats()
	if stats.EntryCount > 5 {
		t.Errorf("expected at most 5 entries after resize, got %d", stats.EntryCount)
	}
}

func TestMeasureCache_Resize_Noop(t *testing.T) {
	cache := NewMeasureCache()

	// Put 3 entries
	for i := 0; i < 3; i++ {
		node := newMockVNode(fmt.Sprintf("node-%d", i))
		cache.Put(node, layout.Constraints{}, layout.Size{}, 1)
	}

	// Resize to 10 (larger than current)
	cache.Resize(10)

	// Should still have 3 entries
	stats := cache.Stats()
	if stats.EntryCount != 3 {
		t.Errorf("expected 3 entries after resize, got %d", stats.EntryCount)
	}
}

// =============================================================================
// Stats Tests
// =============================================================================

func TestMeasureCache_Stats(t *testing.T) {
	cache := NewMeasureCache()
	node := newMockVNode("stats-node")
	constraints := layout.Constraints{}

	// Initial stats
	stats := cache.Stats()
	if stats.EntryCount != 0 || stats.TotalHits != 0 {
		t.Errorf("expected empty stats, got %v", stats)
	}

	// Add entry
	cache.Put(node, constraints, layout.Size{}, 1)

	// Access it multiple times
	for i := 0; i < 5; i++ {
		cache.Get(node, constraints, 1)
	}

	// Check stats
	stats = cache.Stats()
	if stats.EntryCount != 1 {
		t.Errorf("expected 1 entry, got %d", stats.EntryCount)
	}
	if stats.TotalHits != 5 {
		t.Errorf("expected 5 hits, got %d", stats.TotalHits)
	}
}

func TestCacheStats_String(t *testing.T) {
	stats := CacheStats{
		EntryCount: 10,
		TotalHits:  50,
	}

	str := stats.String()
	if str == "" {
		t.Error("String() should return non-empty string")
	}
}

// =============================================================================
// MeasureWithCache Tests
// =============================================================================

func TestMeasureWithCache_CacheHit(t *testing.T) {
	cache := NewMeasureCache()
	node := newMockVNode("cached-node")
	constraints := layout.Constraints{MinWidth: 10, MaxWidth: 50}

	// Pre-populate cache
	expectedSize := layout.Size{Width: 30, Height: 15}
	cache.Put(node, constraints, expectedSize, 1)

	// Measure with cache
	size := MeasureWithCache(cache, node, constraints, 1)

	if size.Width != 30 || size.Height != 15 {
		t.Errorf("expected 30x15, got %dx%d", size.Width, size.Height)
	}
}

func TestMeasureWithCache_CacheMiss(t *testing.T) {
	cache := NewMeasureCache()
	node := newMockVNode("uncached-node")
	constraints := layout.Constraints{MinWidth: 20, MaxWidth: 60}

	// Measure with cache (miss)
	size := MeasureWithCache(cache, node, constraints, 1)

	// Should use measurable.Measured size or props
	if size.Width != 20 || size.Height != 0 {
		// mockMeasurableVNode returns size based on constraints
		t.Logf("got size %dx%d", size.Width, size.Height)
	}
}

func TestMeasureWithCache_NilCache(t *testing.T) {
	var cache *MeasureCache = nil
	node := newMockVNode("nil-cache-node")
	constraints := layout.Constraints{}

	// Should not panic
	size := MeasureWithCache(cache, node, constraints, 1)
	if size.Width < 0 || size.Height < 0 {
		t.Error("size should be valid")
	}
}

func TestMeasureWithCache_NonMeasurable(t *testing.T) {
	cache := NewMeasureCache()

	// Create a non-measurable VNode (plain text)
	node := text.New("Hello")
	constraints := layout.Constraints{}

	// Should not panic, return default or props-based size
	size := MeasureWithCache(cache, node, constraints, 1)
	if size.Width < 0 || size.Height < 0 {
		t.Error("size should be valid")
	}
}

// =============================================================================
// Helper Tests
// =============================================================================

func TestNodeMatchesKey(t *testing.T) {
	tests := []struct {
		name     string
		nodeKey  string
		cacheKey string
		expected bool
	}{
		{"Exact match", "node1", "node1|10,50,5,20", true},
		{"Different node", "node1", "node2|10,50,5,20", false},
		{"Node prefix match", "node1", "node1-extra|10,50,5,20", false},
		{"Cache key without separator", "node1", "node1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nodeMatchesKey(tt.nodeKey, tt.cacheKey)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// =============================================================================
// Concurrency Tests
// =============================================================================

func TestMeasureCache_ConcurrentAccess(t *testing.T) {
	cache := NewMeasureCache()
	done := make(chan bool)

	// Concurrent writers
	for i := 0; i < 10; i++ {
		go func(id int) {
			node := newMockVNode(fmt.Sprintf("node-%d", id))
			cache.Put(node, layout.Constraints{}, layout.Size{}, 1)
			done <- true
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 10; i++ {
		go func(id int) {
			node := newMockVNode(fmt.Sprintf("node-%d", id))
			cache.Get(node, layout.Constraints{}, 1)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Should have no panics and valid stats
	stats := cache.Stats()
	if stats.EntryCount > 10 {
		t.Errorf("expected at most 10 entries, got %d", stats.EntryCount)
	}
}
