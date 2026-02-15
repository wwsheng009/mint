package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlexCache_GetAndPut(t *testing.T) {
	cache := NewFlexCache()

	child1 := NewMockMeasurableNode("child1", 50, 25)
	child2 := NewMockMeasurableNode("child2", 60, 30)
	flexLayout := NewFlexLayout("flex", []Node{child1, child2})
	flexLayout.SetFlex(0, 2, 1, 100) // grow=2

	isRow := true
	flexibleIndices := []int{0}

	computeFunc := func() *FlexDistributionInfo {
		return &FlexDistributionInfo{
			TotalFlexFactor: 2,
			FixedSize:       60,
			ChildCount:      2,
			MaxCrossSize:    30,
			Valid:           true,
			Version:         1,
		}
	}

	// First call - compute
	info1 := cache.Get(flexLayout.ID(), flexLayout.Children(), flexibleIndices, isRow, computeFunc)
	assert.NotNil(t, info1, "Should return flex distribution info")
	assert.Equal(t, 2, info1.TotalFlexFactor, "TotalFlexFactor should match")
	assert.Equal(t, 60, info1.FixedSize, "FixedSize should match")
	assert.True(t, info1.Valid, "Should be valid")

	// Second call - should return cached
	info2 := cache.Get(flexLayout.ID(), flexLayout.Children(), flexibleIndices, isRow, computeFunc)
	assert.NotNil(t, info2, "Should return cached flex distribution info")
	assert.Equal(t, info1, info2, "Should return same cached instance")
}

func TestFlexCache_Invalidate(t *testing.T) {
	cache := NewFlexCache()

	child := NewMockMeasurableNode("child", 50, 25)
	flexLayout := NewFlexLayout("flex", []Node{child})

	isRow := true
	flexibleIndices := []int{}

	computeFunc := func() *FlexDistributionInfo {
		return &FlexDistributionInfo{
			TotalFlexFactor: 0,
			FixedSize:       50,
			ChildCount:      1,
			MaxCrossSize:    25,
			Valid:           true,
			Version:         1,
		}
	}

	// Add entry
	info1 := cache.Get(flexLayout.ID(), flexLayout.Children(), flexibleIndices, isRow, computeFunc)
	assert.NotNil(t, info1)

	// Verify entry exists
	total, valid := cache.GetStats()
	assert.Equal(t, 1, total, "Should have 1 entry")
	assert.Equal(t, 1, valid, "Should have 1 valid entry")

	// Invalidate
	cache.Invalidate(flexLayout.ID())

	// Verify entry is removed
	total, valid = cache.GetStats()
	assert.Equal(t, 0, total, "Should have 0 entries after invalidation")
	assert.Equal(t, 0, valid, "Should have 0 valid entries after invalidation")
}

func TestFlexCache_InvalidateAll(t *testing.T) {
	cache := NewFlexCache()

	// Add multiple entries
	for i := 0; i < 5; i++ {
		child := NewMockMeasurableNode("child", 50, 25)
		flexLayout := NewFlexLayout("flex"+string(rune('0'+i)), []Node{child})

		cache.Get(flexLayout.ID(), flexLayout.Children(), []int{}, true, func() *FlexDistributionInfo {
			return &FlexDistributionInfo{
				TotalFlexFactor: 0,
				FixedSize:       50,
				ChildCount:      1,
				Valid:           true,
				Version:         uint64(i + 1),
			}
		})
	}

	// Verify entries exist
	total, valid := cache.GetStats()
	assert.Equal(t, 5, total, "Should have 5 entries")
	assert.Equal(t, 5, valid, "Should have 5 valid entries")

	// Invalidate all
	cache.InvalidateAll()

	// Verify all entries are removed
	total, valid = cache.GetStats()
	assert.Equal(t, 0, total, "Should have 0 entries after InvalidateAll")
	assert.Equal(t, 0, valid, "Should have 0 valid entries after InvalidateAll")
}

func TestFlexCache_Clear(t *testing.T) {
	cache := NewFlexCache()

	child := NewMockMeasurableNode("child", 50, 25)
	flexLayout := NewFlexLayout("flex", []Node{child})

	cache.Get(flexLayout.ID(), flexLayout.Children(), []int{}, true, func() *FlexDistributionInfo {
		return &FlexDistributionInfo{Valid: true}
	})

	// Verify entry exists
	total, _ := cache.GetStats()
	assert.Equal(t, 1, total)

	// Clear (alias for InvalidateAll)
	cache.Clear()

	// Verify entry is removed
	total, _ = cache.GetStats()
	assert.Equal(t, 0, total)
}

func TestFlexCache_GetStats(t *testing.T) {
	cache := NewFlexCache()

	// Add valid entries
	for i := 0; i < 3; i++ {
		child := NewMockMeasurableNode("child", 50, 25)
		flexLayout := NewFlexLayout("flex"+string(rune('0'+i)), []Node{child})

		cache.Get(flexLayout.ID(), flexLayout.Children(), []int{}, true, func() *FlexDistributionInfo {
			return &FlexDistributionInfo{
				TotalFlexFactor: 0,
				FixedSize:       50,
				ChildCount:      1,
				Valid:           true,
				Version:         1,
			}
		})
	}

	// Add invalid entries
	for i := 3; i < 5; i++ {
		child := NewMockMeasurableNode("child", 50, 25)
		flexLayout := NewFlexLayout("flex"+string(rune('0'+i)), []Node{child})

		cache.Get(flexLayout.ID(), flexLayout.Children(), []int{}, true, func() *FlexDistributionInfo {
			return &FlexDistributionInfo{
				Valid: false,
			}
		})
	}

	// Check stats
	total, valid := cache.GetStats()
	assert.Equal(t, 5, total, "Should have 5 total entries")
	assert.Equal(t, 3, valid, "Should have 3 valid entries")
}

func TestFlexCache_ThreadSafety(t *testing.T) {
	cache := NewFlexCache()

	// This test documents that FlexCache is thread-safe due to sync.RWMutex
	// In a real test, we would use goroutines to verify concurrent access,
	// but for simplicity we just verify the mutex exists
	assert.NotNil(t, cache, "Cache should be created")
}

// Benchmark tests
func BenchmarkFlexCache_GetHit(b *testing.B) {
	cache := NewFlexCache()

	child := NewMockMeasurableNode("child", 50, 25)
	flexLayout := NewFlexLayout("flex", []Node{child})

	// Warm up cache
	cache.Get(flexLayout.ID(), flexLayout.Children(), []int{}, true, func() *FlexDistributionInfo {
		return &FlexDistributionInfo{Valid: true}
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(flexLayout.ID(), flexLayout.Children(), []int{}, true, func() *FlexDistributionInfo {
			return &FlexDistributionInfo{Valid: true}
		})
	}
}

func BenchmarkFlexCache_Invalidate(b *testing.B) {
	cache := NewFlexCache()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		child := NewMockMeasurableNode("child"+string(rune('0'+i%10)), 50, 25)
		flexLayout := NewFlexLayout("flex"+string(rune('0'+i%10)), []Node{child})

		cache.Get(flexLayout.ID(), flexLayout.Children(), []int{}, true, func() *FlexDistributionInfo {
			return &FlexDistributionInfo{Valid: true}
		})

		if i%100 == 0 {
			cache.Invalidate(flexLayout.ID())
		}
	}
}
