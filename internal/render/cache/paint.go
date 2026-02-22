package cache

import (
	"fmt"
	"sync"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// Paint Cache
// =============================================================================

// PaintCache caches the rendered content of PaintableBox instances.
// This avoids redundant painting operations for nodes that haven't changed.
type PaintCache struct {
	mu    sync.RWMutex
	cache map[string]*paintedEntry
	stats CacheStats
}

// paintedEntry represents a cached painted box.
type paintedEntry struct {
	boxID    string // Node ID or unique identifier
	content  [][]cacheCell // Cached cell matrix
	width    int
	height   int
	version  int  // Version of the node when cached
	tainted  bool // Whether the cache is tainted (needs repaint)
	hitCount int  // Number of times this entry was used
}

// cacheCell represents a single cached cell.
type cacheCell struct {
	cluster string
	style   style.Style
	valid   bool
}

// NewPaintCache creates a new paint cache.
func NewPaintCache() *PaintCache {
	return &PaintCache{
		cache: make(map[string]*paintedEntry),
	}
}

// Get retrieves cached content for a PaintableBox.
// Returns (content, width, height, found).
func (pc *PaintCache) Get(boxID string, version int) ([][]cacheCell, int, int, bool) {
	if pc == nil {
		return nil, 0, 0, false
	}

	pc.mu.RLock()
	defer pc.mu.RUnlock()

	entry, exists := pc.cache[boxID]
	if !exists {
		pc.stats.Misses++
		return nil, 0, 0, false
	}

	// Check if cache is tainted or version mismatch
	if entry.tainted || entry.version != version {
		pc.stats.Misses++
		return nil, 0, 0, false
	}

	entry.hitCount++
	pc.stats.Hits++
	return entry.content, entry.width, entry.height, true
}

// Put stores cached content for a PaintableBox.
func (pc *PaintCache) Put(boxID string, content [][]cacheCell, width, height, version int) {
	if pc == nil {
		return
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.cache[boxID] = &paintedEntry{
		boxID:   boxID,
		content: content,
		width:   width,
		height:  height,
		version: version,
		tainted: false,
	}
}

// Invalidate marks a cached entry as tainted.
func (pc *PaintCache) Invalidate(boxID string) {
	if pc == nil {
		return
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	if entry, exists := pc.cache[boxID]; exists {
		entry.tainted = true
	}
}

// InvalidateAll invalidates all cached entries.
func (pc *PaintCache) InvalidateAll() {
	if pc == nil {
		return
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	for _, entry := range pc.cache {
		entry.tainted = true
	}
}

// Remove removes a cached entry.
func (pc *PaintCache) Remove(boxID string) {
	if pc == nil {
		return
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	delete(pc.cache, boxID)
}

// Clear removes all cached entries.
func (pc *PaintCache) Clear() {
	if pc == nil {
		return
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.cache = make(map[string]*paintedEntry)
	pc.stats = CacheStats{}
}

// Resize limits the cache size to the maximum number of entries.
func (pc *PaintCache) Resize(maxEntries int) {
	if pc == nil {
		return
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	if len(pc.cache) <= maxEntries {
		return
	}

	// Evict least recently used (lowest hit count)
	pc.evictLRU(len(pc.cache) - maxEntries)
}

// evictLRU evicts the specified number of least recently used entries.
func (pc *PaintCache) evictLRU(count int) {
	if count <= 0 {
		return
	}

	for i := 0; i < count && len(pc.cache) > 0; i++ {
		var minHits int
		var minKey string
		minHits = 1<<31 - 1 // Max int

		for key, entry := range pc.cache {
			if entry.hitCount < minHits {
				minHits = entry.hitCount
				minKey = key
			}
		}

		if minKey != "" {
			delete(pc.cache, minKey)
		}
	}
}

// Stats returns cache statistics.
func (pc *PaintCache) Stats() CacheStats {
	if pc == nil {
		return CacheStats{}
	}

	pc.mu.RLock()
	defer pc.mu.RUnlock()

	return pc.stats
}

// GetEntryCount returns the number of cached entries.
func (pc *PaintCache) GetEntryCount() int {
	if pc == nil {
		return 0
	}

	pc.mu.RLock()
	defer pc.mu.RUnlock()

	return len(pc.cache)
}

// =============================================================================
// Cache Stats
// =============================================================================

// CacheStats contains statistics about the paint cache.
type CacheStats struct {
	Hits   int // Number of cache hits
	Misses int // Number of cache misses
}

// HitRate returns the cache hit rate as a percentage.
func (s CacheStats) HitRate() float64 {
	if s.Hits+s.Misses == 0 {
		return 0.0
	}
	return float64(s.Hits) / float64(s.Hits+s.Misses) * 100.0
}

// String returns a string representation of the stats.
func (s CacheStats) String() string {
	return fmt.Sprintf("hits=%d misses=%d hit_rate=%.1f%%", s.Hits, s.Misses, s.HitRate())
}

// =============================================================================
// Painting Context - Helper for cache-aware painting
// =============================================================================

// PaintingContext provides context for cache-aware painting.
type PaintingContext struct {
	cache      *PaintCache
	bufferCopy *paint.Buffer // Copy of previous buffer for comparison
	version    int
}

// NewPaintingContext creates a new painting context.
func NewPaintingContext(cache *PaintCache, buffer *paint.Buffer, version int) *PaintingContext {
	pc := &PaintingContext{
		cache:   cache,
		version: version,
	}

	// Create buffer copy for dirty rect calculation
	if buffer != nil {
		pc.bufferCopy = pc.cloneBuffer(buffer)
	}

	return pc
}

// cloneBuffer creates a deep copy of a buffer.
func (pc *PaintingContext) cloneBuffer(buffer *paint.Buffer) *paint.Buffer {
	if buffer == nil {
		return nil
	}
	copied := paint.NewBuffer(buffer.Width, buffer.Height)
	for y := 0; y < buffer.Height; y++ {
		for x := 0; x < buffer.Width; x++ {
			copied.Cells[y][x] = buffer.Cells[y][x]
		}
	}
	return copied
}

// TryPaintFromCache attempts to paint from cache.
// Returns true if painted from cache, false otherwise.
func (pc *PaintingContext) TryPaintFromCache(buffer *paint.Buffer, boxID string, x, y int) bool {
	if pc == nil || pc.cache == nil || buffer == nil {
		return false
	}

	content, width, height, found := pc.cache.Get(boxID, pc.version)
	if !found {
		return false
	}

	// Copy cached content to buffer
	pc.paintContent(buffer, content, x, y, width, height)
	return true
}

// paintContent paints cached content to buffer.
func (pc *PaintingContext) paintContent(
	buffer *paint.Buffer,
	content [][]cacheCell,
	x, y, width, height int,
) {
	if buffer == nil || content == nil {
		return
	}

	for dy := 0; dy < height && y+dy < buffer.Height; dy++ {
		if len(content) <= dy {
			break
		}
		for dx := 0; dx < width && x+dx < buffer.Width; dx++ {
			if len(content[dy]) <= dx {
				break
			}
			cell := content[dy][dx]
			if cell.valid && x+dx >= 0 && y+dy >= 0 {
				buffer.Cells[y+dy][x+dx] = paint.Cell{
					Cluster: cell.cluster,
					Style:   cell.style,
				}
			}
		}
	}
}

// UpdateCache updates the cache with new painted content.
func (pc *PaintingContext) UpdateCache(boxID string, bounds layout.Rect, buffer *paint.Buffer) {
	if pc == nil || pc.cache == nil || buffer == nil {
		return
	}

	// Extract content from buffer
	content := make([][]cacheCell, bounds.Height)
	for dy := 0; dy < bounds.Height; dy++ {
		content[dy] = make([]cacheCell, bounds.Width)
		for dx := 0; dx < bounds.Width; dx++ {
			if bounds.X+dx >= 0 && bounds.Y+dy >= 0 &&
				bounds.X+dx < buffer.Width && bounds.Y+dy < buffer.Height {
				cell := buffer.Cells[bounds.Y+dy][bounds.X+dx]
				content[dy][dx] = cacheCell{
					cluster: cell.Cluster,
					style:   cell.Style,
					valid:   true,
				}
			}
		}
	}

	pc.cache.Put(boxID, content, bounds.Width, bounds.Height, pc.version)
}

// InvalidateCache invalidates a cached entry.
func (pc *PaintingContext) InvalidateCache(boxID string) {
	if pc == nil || pc.cache == nil {
		return
	}
	pc.cache.Invalidate(boxID)
}

// GetDirtyRects calculates dirty rectangles by comparing current and previous buffer.
func (pc *PaintingContext) GetDirtyRects(currentBuffer *paint.Buffer) []layout.Rect {
	if pc == nil || pc.bufferCopy == nil || currentBuffer == nil {
		return nil
	}

	rects := make([]layout.Rect, 0)
	detected := make([][]bool, currentBuffer.Height)
	for y := 0; y < currentBuffer.Height; y++ {
		detected[y] = make([]bool, currentBuffer.Width)
	}

	// Grid-based dirty detection (merge adjacent cells)
	for y := 0; y < min(currentBuffer.Height, pc.bufferCopy.Height); y++ {
		for x := 0; x < min(currentBuffer.Width, pc.bufferCopy.Width); x++ {
			if detected[y][x] {
				continue
			}

			current := currentBuffer.Cells[y][x]
			previous := pc.bufferCopy.Cells[y][x]

			if current.Cluster != previous.Cluster || current.Style != previous.Style {
				// Found changed cell, try to expand rectangle
				maxX := x
				for xx := x + 1; xx < currentBuffer.Width; xx++ {
					cy := currentBuffer.Cells[y][xx]
					py := pc.bufferCopy.Cells[y][xx]
					if cy.Cluster == py.Cluster && cy.Style == py.Style {
						break
					}
					maxX = xx
				}

				maxY := y
				for yy := y + 1; yy < currentBuffer.Height; yy++ {
					rowChanged := false
					for xx := x; xx <= maxX; xx++ {
						cy := currentBuffer.Cells[yy][xx]
						py := pc.bufferCopy.Cells[yy][xx]
						if cy.Cluster != py.Cluster || cy.Style != py.Style {
							rowChanged = true
							break
						}
					}
					if !rowChanged {
						break
					}
					maxY = yy
				}

				rects = append(rects, layout.Rect{
					X:      x,
					Y:      y,
					Width:  maxX - x + 1,
					Height: maxY - y + 1,
				})

				// Mark cells as detected
				for yy := y; yy <= maxY && yy < currentBuffer.Height; yy++ {
					for xx := x; xx <= maxX && xx < currentBuffer.Width; xx++ {
						detected[yy][xx] = true
					}
				}
			}
		}
	}

	return rects
}

// UpdateBufferCopy updates the stored buffer copy.
func (pc *PaintingContext) UpdateBufferCopy(buffer *paint.Buffer) {
	if pc == nil {
		return
	}
	pc.bufferCopy = pc.cloneBuffer(buffer)
}

// GetStats returns cache statistics.
func (pc *PaintingContext) GetStats() CacheStats {
	if pc == nil || pc.cache == nil {
		return CacheStats{}
	}
	return pc.cache.Stats()
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
