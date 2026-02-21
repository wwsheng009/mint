package cache

import (
	"fmt"
	"sync"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Measure Cache
// =============================================================================

// MeasureCache stores measurement results to avoid redundant calculations.
// It is used during layout to cache the measured size of VNodes under specific constraints.
type MeasureCache struct {
	mu    sync.RWMutex
	cache map[string]*cacheEntry
}

// cacheEntry holds the cached measurement and its metadata.
type cacheEntry struct {
	size      layout.Size
	version   int    // Version of the VNode at time of measurement
	timestamp int64  // For LRU eviction (optional)
	hitCount  int    // Number of times this entry was accessed
}

// NewMeasureCache creates a new empty measure cache.
func NewMeasureCache() *MeasureCache {
	return &MeasureCache{
		cache: make(map[string]*cacheEntry),
	}
}

// Get retrieves a cached measurement for the given VNode and constraints.
// Returns (size, found).
func (mc *MeasureCache) Get(vnode ui.VNode, constraints layout.Constraints, version int) (layout.Size, bool) {
	if mc == nil {
		return layout.Size{}, false
	}

	key := mc.key(vnode, constraints)

	mc.mu.RLock()
	entry, exists := mc.cache[key]
	mc.mu.RUnlock()

	if !exists {
		return layout.Size{}, false
	}

	// Check version to ensure cache is still valid
	if entry.version != version {
		mc.mu.Lock()
		delete(mc.cache, key)
		mc.mu.Unlock()
		return layout.Size{}, false
	}

	// Update hit count
	mc.mu.Lock()
	entry.hitCount++
	mc.mu.Unlock()

	return entry.size, true
}

// Put stores a measurement for the given VNode and constraints.
func (mc *MeasureCache) Put(vnode ui.VNode, constraints layout.Constraints, size layout.Size, version int) {
	if mc == nil {
		return
	}

	key := mc.key(vnode, constraints)

	mc.mu.Lock()
	if entry, exists := mc.cache[key]; exists {
		// Update existing entry
		entry.size = size
		entry.version = version
		entry.hitCount = entry.hitCount + 1
	} else {
		// Create new entry
		mc.cache[key] = &cacheEntry{
			size:    size,
			version: version,
		}
	}
	mc.mu.Unlock()
}

// Invalidate removes all cached measurements for the given VNode.
// Call this when a VNode is modified or removed.
func (mc *MeasureCache) Invalidate(vnode ui.VNode) {
	if mc == nil {
		return
	}

	nodeKey := mc.nodeKey(vnode)

	mc.mu.Lock()
	for key := range mc.cache {
		if nodeMatchesKey(nodeKey, key) {
			delete(mc.cache, key)
		}
	}
	mc.mu.Unlock()
}

// InvalidateAll clears the entire cache.
// Call this when the layout tree is significantly changed.
func (mc *MeasureCache) InvalidateAll() {
	if mc == nil {
		return
	}

	mc.mu.Lock()
	mc.cache = make(map[string]*cacheEntry)
	mc.mu.Unlock()
}

// InvalidateTree invalidates cache entries for a subtree rooted at the given VNode.
func (mc *MeasureCache) InvalidateTree(root ui.VNode) {
	if mc == nil {
		return
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	var queue []ui.VNode
	queue = append(queue, root)

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		// Invalidate this node's cache entries
		nodeKey := mc.nodeKey(node)
		for key := range mc.cache {
			if nodeMatchesKey(nodeKey, key) {
				delete(mc.cache, key)
			}
		}

		// Add children to queue
		children := node.Children()
		for _, child := range children {
			queue = append(queue, child)
		}
	}
}

// Stats returns statistics about the cache.
func (mc *MeasureCache) Stats() CacheStats {
	if mc == nil {
		return CacheStats{}
	}

	mc.mu.RLock()
	defer mc.mu.RUnlock()

	var totalHits int
	for _, entry := range mc.cache {
		totalHits += entry.hitCount
	}

	return CacheStats{
		EntryCount: len(mc.cache),
		TotalHits:  totalHits,
	}
}

// Resize limits the cache size to the maximum number of entries.
// If the cache exceeds the limit, least recently used entries are evicted.
func (mc *MeasureCache) Resize(maxEntries int) {
	if mc == nil {
		return
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	if len(mc.cache) <= maxEntries {
		return
	}

	// Simple eviction strategy: remove entries with lowest hit count
	for len(mc.cache) > maxEntries {
		var minKey string
		minHits := int(^uint(0) >> 1) // Max int

		for key, entry := range mc.cache {
			if entry.hitCount < minHits {
				minHits = entry.hitCount
				minKey = key
			}
		}

		if minKey != "" {
			delete(mc.cache, minKey)
		} else {
			break
		}
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// key generates a cache key for a VNode with specific constraints.
func (mc *MeasureCache) key(vnode ui.VNode, constraints layout.Constraints) string {
	return fmt.Sprintf("%s|%s",
		mc.nodeKey(vnode),
		mc.constraintsKey(constraints))
}

// nodeKey generates a unique identifier for a VNode.
func (mc *MeasureCache) nodeKey(vnode ui.VNode) string {
	if vnode == nil {
		return "nil"
	}
	key := vnode.Key()
	if key == "" {
		// Fallback to type and tag if no key
		key = fmt.Sprintf("%s-%s", vnode.Type(), vnode.Tag())
	}
	return key
}

// constraintsKey generates a string representation of constraints.
func (mc *MeasureCache) constraintsKey(c layout.Constraints) string {
	return fmt.Sprintf("%d,%d,%d,%d",
		c.MinWidth, c.MaxWidth,
		c.MinHeight, c.MaxHeight)
}

// nodeMatchesKey checks if a node key matches a cache entry key.
func nodeMatchesKey(nodeKey, cacheKey string) bool {
	// Cache key format is: "nodeKey|constraintsKey"
	// We check if the cache key starts with the node key followed by '|'
	prefix := nodeKey + "|"
	return len(cacheKey) >= len(prefix) && cacheKey[:len(prefix)] == prefix
}

// =============================================================================
// Cache Stats
// =============================================================================

// CacheStats contains statistics about the measure cache.
type CacheStats struct {
	EntryCount int
	TotalHits  int
}

// String returns a string representation of the stats.
func (s CacheStats) String() string {
	hitRate := 0.0
	if s.EntryCount > 0 {
		hitRate = float64(s.TotalHits) / float64(s.EntryCount)
	}
	return fmt.Sprintf("entries=%d hits=%d hit_rate=%.2f",
		s.EntryCount, s.TotalHits, hitRate)
}

// =============================================================================
// Measurable Interface
// =============================================================================

// Measurable is an interface for VNodes that can measure themselves.
type Measurable interface {
	// Measure returns the preferred size of the node given the constraints.
	Measure(layout.Constraints) layout.Size
}

// MeasureWithCache measures a VNode with caching support.
// If the measurement is cached, it returns the cached value.
// Otherwise, it measures and caches the result.
func MeasureWithCache(
	cache *MeasureCache,
	vnode ui.VNode,
	constraints layout.Constraints,
	version int,
) layout.Size {
	// Try to get from cache
	if cache != nil {
		if size, found := cache.Get(vnode, constraints, version); found {
			return size
		}
	}

	// Measure the node
	var size layout.Size
	if measurable, ok := vnode.(Measurable); ok {
		size = measurable.Measure(constraints)
	} else {
		// Default size if not measurable
		size = layout.Size{Width: 0, Height: 0}

		// Try to get size from props
		props := vnode.Props()
		if props != nil {
			if w, ok := props["width"].(int); ok {
				size.Width = w
			}
			if h, ok := props["height"].(int); ok {
				size.Height = h
			}
		}
	}

	// Cache the result
	if cache != nil {
		cache.Put(vnode, constraints, size, version)
	}

	return size
}
