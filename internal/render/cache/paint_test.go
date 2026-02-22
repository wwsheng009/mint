package cache

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// PaintCache Core Tests
// =============================================================================

func TestPaintCache_New(t *testing.T) {
	cache := NewPaintCache()

	if cache == nil {
		t.Fatal("NewPaintCache() should not return nil")
	}
	if cache.GetEntryCount() != 0 {
		t.Errorf("expected 0 entries, got %d", cache.GetEntryCount())
	}
	stats := cache.Stats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Errorf("expected no stats, got %v", stats)
	}
}

func TestPaintCache_PutAndGet(t *testing.T) {
	cache := NewPaintCache()
	boxID := "test-box"

	// Create test content
	content := [][]cacheCell{
		{
			{cluster: "H", style: style.NewStyle(), valid: true},
			{cluster: "i", style: style.NewStyle(), valid: true},
		},
		{
			{cluster: "!", style: style.NewStyle(), valid: true},
		},
	}

	// Put
	cache.Put(boxID, content, 2, 2, 1)

	// Get
	retContent, width, height, found := cache.Get(boxID, 1)
	if !found {
		t.Fatal("expected to find cached content")
	}
	if width != 2 || height != 2 {
		t.Errorf("expected 2x2, got %dx%d", width, height)
	}
	if len(retContent) != 2 || len(retContent[0]) != 2 {
		t.Error("content dimensions mismatch")
	}

	stats := cache.Stats()
	if stats.Hits != 1 {
		t.Errorf("expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 0 {
		t.Errorf("expected 0 misses, got %d", stats.Misses)
	}
}

func TestPaintCache_Get_NotFound(t *testing.T) {
	cache := NewPaintCache()

	_, _, _, found := cache.Get("nonexistent", 1)
	if found {
		t.Error("should not find nonexistent entry")
	}

	stats := cache.Stats()
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}
}

func TestPaintCache_VersionMismatch(t *testing.T) {
	cache := NewPaintCache()
	boxID := "versioned-box"

	content := [][]cacheCell{
		{{cluster: "A", style: style.Style{}, valid: true}},
	}

	cache.Put(boxID, content, 1, 1, 1)

	// Get with wrong version should miss
	_, _, _, found := cache.Get(boxID, 2)
	if found {
		t.Error("should not find with version mismatch")
	}

	stats := cache.Stats()
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}
}

// =============================================================================
// Invalidation Tests
// =============================================================================

func TestPaintCache_Invalidate(t *testing.T) {
	cache := NewPaintCache()
	boxID := "invalidatable-box"

	content := [][]cacheCell{
		{{cluster: "X", style: style.Style{}, valid: true}},
	}

	cache.Put(boxID, content, 1, 1, 1)

	// Invalidate
	cache.Invalidate(boxID)

	// Get should miss
	_, _, _, found := cache.Get(boxID, 1)
	if found {
		t.Error("should not find invalidated entry")
	}
}

func TestPaintCache_InvalidateAll(t *testing.T) {
	cache := NewPaintCache()

	// Put multiple entries
	for i := 0; i < 5; i++ {
		content := [][]cacheCell{
			{{cluster: string(rune('A' + i)), style: style.Style{}, valid: true}},
		}
		cache.Put(string(rune('0'+i)), content, 1, 1, 1)
	}

	// Invalidate all
	cache.InvalidateAll()

	// All should miss
	for i := 0; i < 5; i++ {
		_, _, _, found := cache.Get(string(rune('0'+i)), 1)
		if found {
			t.Error("invalidated entries should not be found")
		}
	}
	// Misses should be 5
	stats := cache.Stats()
	if stats.Misses != 5 {
		t.Errorf("expected 5 misses, got %d", stats.Misses)
	}
}

func TestPaintCache_Remove(t *testing.T) {
	cache := NewPaintCache()
	boxID := "removable-box"

	content := [][]cacheCell{
		{{cluster: "R", style: style.Style{}, valid: true}},
	}

	cache.Put(boxID, content, 1, 1, 1)

	// Remove
	cache.Remove(boxID)

	if cache.GetEntryCount() != 0 {
		t.Errorf("expected 0 entries after remove, got %d", cache.GetEntryCount())
	}

	_, _, _, found := cache.Get(boxID, 1)
	if found {
		t.Error("removed entry should not be found")
	}
}

func TestPaintCache_Clear(t *testing.T) {
	cache := NewPaintCache()

	// Add entries
	for i := 0; i < 3; i++ {
		content := [][]cacheCell{
			{{cluster: string(rune('A' + i)), style: style.Style{}, valid: true}},
		}
		cache.Put(string(rune('0'+i)), content, 1, 1, 1)
	}

	// Clear
	cache.Clear()

	if cache.GetEntryCount() != 0 {
		t.Errorf("expected 0 entries after clear, got %d", cache.GetEntryCount())
	}

	stats := cache.Stats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Errorf("expected cleared stats, got %v", stats)
	}
}

// =============================================================================
// Resize Tests
// =============================================================================

func TestPaintCache_Resize(t *testing.T) {
	cache := NewPaintCache()

	// Add more than limit
	for i := 0; i < 10; i++ {
		content := [][]cacheCell{
			{{cluster: string(rune('A' + i)), style: style.Style{}, valid: true}},
		}
		cache.Put(string(rune('0'+i)), content, 1, 1, 1)
	}

	// Resize to 5
	cache.Resize(5)

	if cache.GetEntryCount() > 5 {
		t.Errorf("expected at most 5 entries, got %d", cache.GetEntryCount())
	}
}

func TestPaintCache_Resize_Noop(t *testing.T) {
	cache := NewPaintCache()

	// Add 3 entries
	for i := 0; i < 3; i++ {
		content := [][]cacheCell{
			{{cluster: string(rune('A' + i)), style: style.Style{}, valid: true}},
		}
		cache.Put(string(rune('0'+i)), content, 1, 1, 1)
	}

	// Resize to 10 (larger than current)
	cache.Resize(10)

	if cache.GetEntryCount() != 3 {
		t.Errorf("expected 3 entries, got %d", cache.GetEntryCount())
	}
}

// =============================================================================
// CacheStats Tests
// =============================================================================

func TestCacheStats_HitRate(t *testing.T) {
	stats := CacheStats{Hits: 80, Misses: 20}

	rate := stats.HitRate()
	if rate != 80.0 {
		t.Errorf("expected 80.0%%, got %.1f%%", rate)
	}

	stats = CacheStats{Hits: 0, Misses: 0}
	rate = stats.HitRate()
	if rate != 0.0 {
		t.Errorf("expected 0.0%%, got %.1f%%", rate)
	}
}

func TestCacheStats_String(t *testing.T) {
	stats := CacheStats{Hits: 100, Misses: 50}

	str := stats.String()
	if str == "" {
		t.Error("String() should return non-empty")
	}
}

// =============================================================================
// PaintingContext Tests
// =============================================================================

func TestPaintingContext_New(t *testing.T) {
	buffer := paint.NewBuffer(10, 10)
	pc := NewPaintingContext(NewPaintCache(), buffer, 1)

	if pc == nil {
		t.Fatal("NewPaintingContext() should not return nil")
	}
	if pc.version != 1 {
		t.Errorf("expected version 1, got %d", pc.version)
	}
}

func TestPaintingContext_TryPaintFromCache_Basic(t *testing.T) {
	cache := NewPaintCache()
	buffer := paint.NewBuffer(10, 10)
	pc := NewPaintingContext(cache, buffer, 1)

	boxID := "cache-test-box"

	// Pre-populate cache
	content := [][]cacheCell{
		{
			{cluster: "T", style: style.NewStyle().Foreground(style.Color("red")), valid: true},
			{cluster: "e", style: style.NewStyle().Foreground(style.Color("red")), valid: true},
			{cluster: "s", style: style.NewStyle().Foreground(style.Color("red")), valid: true},
			{cluster: "t", style: style.NewStyle().Foreground(style.Color("red")), valid: true},
		},
	}
	cache.Put(boxID, content, 4, 1, 1)

	// Try to paint from cache
	painted := pc.TryPaintFromCache(buffer, boxID, 0, 0)
	if !painted {
		t.Fatal("should have painted from cache")
	}

	// Verify content was painted
	ch := buffer.Cells[0][0].Cluster
	if ch != "T" {
		t.Errorf("expected 'T', got %s", ch)
	}
}

func TestPaintingContext_TryPaintFromCache_Miss(t *testing.T) {
	cache := NewPaintCache()
	buffer := paint.NewBuffer(10, 10)
	pc := NewPaintingContext(cache, buffer, 1)

	painted := pc.TryPaintFromCache(buffer, "nonexistent", 0, 0)
	if painted {
		t.Error("should not paint from cache (miss)")
	}
}

func TestPaintingContext_UpdateCache(t *testing.T) {
	cache := NewPaintCache()
	buffer := paint.NewBuffer(10, 10)
	pc := NewPaintingContext(cache, buffer, 1)

	// Paint some content to buffer
	testStyle := style.NewStyle().Foreground(style.Color("blue"))
	for i := 0; i < 5; i++ {
		buffer.Cells[0][i] = paint.Cell{
			Cluster: string(rune('A' + i)),
			Style:   testStyle,
		}
	}

	// Update cache
	boxID := "update-test-box"
	bounds := layout.Rect{X: 0, Y: 0, Width: 5, Height: 1}
	pc.UpdateCache(boxID, bounds, buffer)

	// Verify cache has the entry
	if cache.GetEntryCount() != 1 {
		t.Fatalf("expected 1 entry, got %d", cache.GetEntryCount())
	}

	// Retrieve and verify
	content, width, height, found := cache.Get(boxID, 1)
	if !found {
		t.Fatal("should find entry")
	}
	if width != 5 || height != 1 {
		t.Errorf("expected 5x1, got %dx%d", width, height)
	}
	if content[0][0].cluster != "A" {
		t.Errorf("expected 'A', got %s", content[0][0].cluster)
	}
}

func TestPaintingContext_InvalidateCache(t *testing.T) {
	cache := NewPaintCache()
	buffer := paint.NewBuffer(10, 10)
	pc := NewPaintingContext(cache, buffer, 1)

	// Add entry to cache
	content := [][]cacheCell{
		{{cluster: "I", style: style.Style{}, valid: true}},
	}
	cache.Put("invalidatable", content, 1, 1, 1)

	pc.InvalidateCache("invalidatable")

	// Should miss
	_, _, _, found := cache.Get("invalidatable", 1)
	if found {
		t.Error("should miss after invalidate")
	}
}

func TestPaintingContext_GetDirtyRects(t *testing.T) {
	oldBuffer := paint.NewBuffer(5, 5)
	newBuffer := paint.NewBuffer(5, 5)
	pc := NewPaintingContext(nil, oldBuffer, 1)
	pc.UpdateBufferCopy(oldBuffer)

	// Add some different cells
	newBuffer.Cells[1][1] = paint.Cell{Cluster: "X", Style: style.NewStyle()}
	newBuffer.Cells[1][2] = paint.Cell{Cluster: "Y", Style: style.NewStyle()}

	dirtyRects := pc.GetDirtyRects(newBuffer)
	if len(dirtyRects) != 1 {
		t.Errorf("expected 1 dirty rect, got %d", len(dirtyRects))
	}
	if dirtyRects[0].X != 1 || dirtyRects[0].Y != 1 {
		t.Errorf("unexpected dirty rect: %v", dirtyRects[0])
	}
}

func TestPaintingContext_UpdateBufferCopy(t *testing.T) {
	buffer := paint.NewBuffer(10, 10)
	pc := NewPaintingContext(nil, nil, 1)
	pc.UpdateBufferCopy(buffer)

	if pc.bufferCopy == nil {
		t.Error("bufferCopy should not be nil after update")
	}
}

func TestPaintingContext_GetStats(t *testing.T) {
	cache := NewPaintCache()
	buffer := paint.NewBuffer(10, 10)
	pc := NewPaintingContext(cache, buffer, 1)

	stats := pc.GetStats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Errorf("expected empty stats, got %v", stats)
	}
}

// =============================================================================
// Nil Handling Tests
// =============================================================================

func TestPaintCache_Nil_Handling(t *testing.T) {
	var cache *PaintCache = nil

	// Should not panic
	cache.Put("test", nil, 0, 0, 0)
	cache.Get("test", 0)
	cache.Invalidate("test")
	cache.InvalidateAll()
	cache.Remove("test")
	cache.Clear()
	cache.Resize(10)
	cache.Stats()
	cache.GetEntryCount()
}

func TestPaintingContext_Nil_Handling(t *testing.T) {
	var pc *PaintingContext = nil
	buffer := paint.NewBuffer(10, 10)

	// Should not panic
	pc.TryPaintFromCache(buffer, "test", 0, 0)
	pc.UpdateCache("test", layout.Rect{}, buffer)
	pc.InvalidateCache("test")
	pc.GetDirtyRects(buffer)
	pc.UpdateBufferCopy(buffer)
	pc.GetStats()
}
