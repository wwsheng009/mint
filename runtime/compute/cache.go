// Package compute provides caching for layout calculation results
package compute

import (
	"fmt"
	"sync"

	"github.com/wwsheng009/mint/runtime"
)

// LayoutCache caches layout measurement results
// Strategy: Only cache leaf nodes (nodes without children) to avoid
// the complexity of caching entire subtrees
type LayoutCache struct {
	mu    sync.RWMutex
	cache map[LayoutCacheKey]LayoutCacheEntry

	// Statistics
	hits   int
	misses int
}

// LayoutCacheKey is the key for cached layout results
type LayoutCacheKey struct {
	VNodeType   string
	VNodeKey    string
	Constraints runtime.BoxConstraints
	PropsHash   uint64
}

// LayoutCacheEntry is a cached layout result
type LayoutCacheEntry struct {
	Box     runtime.Box
	Size    runtime.Size
	Hash    uint64
	IsLeaf  bool  // True if this is a leaf node (no children)
	VNodeID string // For debugging
}

// NewLayoutCache creates a new layout cache
func NewLayoutCache() *LayoutCache {
	return &LayoutCache{
		cache: make(map[LayoutCacheKey]LayoutCacheEntry),
	}
}

// Get retrieves a cached layout result
func (c *LayoutCache) Get(key LayoutCacheKey) (LayoutCacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.cache == nil {
		return LayoutCacheEntry{}, false
	}
	entry, ok := c.cache[key]
	if ok {
		c.hits++
	} else {
		c.misses++
	}
	return entry, ok
}

// Set stores a layout result in the cache
func (c *LayoutCache) Set(key LayoutCacheKey, entry LayoutCacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache == nil {
		c.cache = make(map[LayoutCacheKey]LayoutCacheEntry)
	}
	c.cache[key] = entry
}

// Clear clears all cached entries
func (c *LayoutCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[LayoutCacheKey]LayoutCacheEntry)
	c.hits = 0
	c.misses = 0
}

// Size returns the number of cached entries
func (c *LayoutCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache)
}

// Invalidate removes entries that match a predicate
func (c *LayoutCache) Invalidate(predicate func(LayoutCacheKey) bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache == nil {
		return
	}
	for key := range c.cache {
		if predicate(key) {
			delete(c.cache, key)
		}
	}
}

// InvalidateByType removes all entries for a specific VNode type
func (c *LayoutCache) InvalidateByType(vNodeType string) {
	c.Invalidate(func(key LayoutCacheKey) bool {
		return key.VNodeType == vNodeType
	})
}

// InvalidateByKey removes all entries for a specific VNode key
func (c *LayoutCache) InvalidateByKey(vnodeKey string) {
	c.Invalidate(func(key LayoutCacheKey) bool {
		return key.VNodeKey == vnodeKey
	})
}

// Stats returns cache statistics
func (c *LayoutCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return CacheStats{
		Size:    len(c.cache),
		Hits:    c.hits,
		Misses:  c.misses,
		HitRate: hitRate,
	}
}

// CacheStats represents cache statistics
type CacheStats struct {
	Size    int
	Hits    int
	Misses  int
	HitRate float64
}

// String returns a string representation of the stats
func (s CacheStats) String() string {
	return fmt.Sprintf("Cache[size=%d, hits=%d, misses=%d, hit_rate=%.2f%%]",
		s.Size, s.Hits, s.Misses, s.HitRate*100)
}

// ResetStats resets hit/miss counters without clearing the cache
func (c *LayoutCache) ResetStats() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.hits = 0
	c.misses = 0
}
