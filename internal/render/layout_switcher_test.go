package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wwsheng009/mint/runtime"
)

// =============================================================================
// LayoutEngineType Tests
// =============================================================================

func TestLayoutEngineType_String(t *testing.T) {
	tests := []struct {
		engineType LayoutEngineType
		expected   string
	}{
		{LayoutEngineCompute, "compute"},
		{LayoutEngineNew, "layout"},
		{LayoutEngineBoth, "both"},
		{LayoutEngineType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.engineType.String())
		})
	}
}

func TestParseLayoutEngineType(t *testing.T) {
	tests := []struct {
		input    string
		expected LayoutEngineType
	}{
		{"compute", LayoutEngineCompute},
		{"stable", LayoutEngineCompute},
		{"old", LayoutEngineCompute},
		{"layout", LayoutEngineNew},
		{"new", LayoutEngineNew},
		{"experimental", LayoutEngineNew},
		{"both", LayoutEngineBoth},
		{"parallel", LayoutEngineBoth},
		{"compare", LayoutEngineBoth},
		{"unknown", LayoutEngineCompute}, // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, ParseLayoutEngineType(tt.input))
		})
	}
}

// =============================================================================
// LayoutSwitcher Tests
// =============================================================================

func TestNewLayoutSwitcher(t *testing.T) {
	switcher := NewLayoutSwitcher()
	assert.NotNil(t, switcher)
	assert.Equal(t, LayoutEngineCompute, switcher.GetEngineType())
	assert.NotNil(t, switcher.computeEngine)
	assert.NotNil(t, switcher.newEngine)
}

func TestLayoutSwitcher_SetEngineType(t *testing.T) {
	switcher := NewLayoutSwitcher()

	// Default is compute
	assert.Equal(t, LayoutEngineCompute, switcher.GetEngineType())

	// Switch to new
	switcher.SetEngineType(LayoutEngineNew)
	assert.Equal(t, LayoutEngineNew, switcher.GetEngineType())

	// Switch to both
	switcher.SetEngineType(LayoutEngineBoth)
	assert.Equal(t, LayoutEngineBoth, switcher.GetEngineType())

	// Switch back to compute
	switcher.SetEngineType(LayoutEngineCompute)
	assert.Equal(t, LayoutEngineCompute, switcher.GetEngineType())
}

func TestLayoutSwitcher_SetCompareResults(t *testing.T) {
	switcher := NewLayoutSwitcher()

	// Default is false
	switcher.mu.RLock()
	compare := switcher.compareResults
	switcher.mu.RUnlock()
	assert.False(t, compare)

	// Enable comparison
	switcher.SetCompareResults(true)
	switcher.mu.RLock()
	compare = switcher.compareResults
	switcher.mu.RUnlock()
	assert.True(t, compare)
}

func TestLayoutSwitcher_SetTolerance(t *testing.T) {
	switcher := NewLayoutSwitcher()

	// Default is 1.0
	switcher.mu.RLock()
	tolerance := switcher.tolerancePercent
	switcher.mu.RUnlock()
	assert.Equal(t, 1.0, tolerance)

	// Set new tolerance
	switcher.SetTolerance(5.0)
	switcher.mu.RLock()
	tolerance = switcher.tolerancePercent
	switcher.mu.RUnlock()
	assert.Equal(t, 5.0, tolerance)
}

func TestLayoutSwitcher_ClearCache(t *testing.T) {
	switcher := NewLayoutSwitcher()
	// Should not panic
	switcher.ClearCache()
}

func TestLayoutSwitcher_SetDebug(t *testing.T) {
	switcher := NewLayoutSwitcher()
	// Should not panic
	switcher.SetDebug(true)
	switcher.SetDebug(false)
}

// =============================================================================
// ParallelRenderingPipeline Tests
// =============================================================================

func TestNewParallelRenderingPipeline(t *testing.T) {
	pipeline := NewParallelRenderingPipeline()
	assert.NotNil(t, pipeline)
	assert.NotNil(t, pipeline.switcher)
	assert.NotNil(t, pipeline.paintEngine)
	assert.Equal(t, LayoutEngineCompute, pipeline.GetLayoutEngineType())
}

func TestParallelRenderingPipeline_SetLayoutEngineType(t *testing.T) {
	pipeline := NewParallelRenderingPipeline()

	// Default is compute
	assert.Equal(t, LayoutEngineCompute, pipeline.GetLayoutEngineType())

	// Switch to new
	pipeline.SetLayoutEngineType(LayoutEngineNew)
	assert.Equal(t, LayoutEngineNew, pipeline.GetLayoutEngineType())

	// Switch to both
	pipeline.SetLayoutEngineType(LayoutEngineBoth)
	assert.Equal(t, LayoutEngineBoth, pipeline.GetLayoutEngineType())
}

func TestParallelRenderingPipeline_Render_NilVNode(t *testing.T) {
	pipeline := NewParallelRenderingPipeline()
	err := pipeline.Render(nil, nil, runtime.BoxConstraints{}, nil)
	assert.NoError(t, err)
}

// =============================================================================
// CacheStats Tests
// =============================================================================

func TestComputeEngineAdapter_GetStats(t *testing.T) {
	adapter := NewComputeEngineAdapter()
	stats := adapter.GetStats()
	assert.Equal(t, 0, stats.Hits)
	assert.Equal(t, 0, stats.Misses)
}

func TestNewLayoutEngineAdapter_GetStats(t *testing.T) {
	adapter := NewNewLayoutEngineAdapter()
	stats := adapter.GetStats()
	assert.Equal(t, 0, stats.Hits)
	assert.Equal(t, 0, stats.Misses)
}

// =============================================================================
// SwitcherStats Tests
// =============================================================================

func TestSwitcherStats(t *testing.T) {
	stats := SwitcherStats{}

	// Initial values
	stats.mu.RLock()
	assert.Equal(t, int64(0), stats.TotalRenders)
	stats.mu.RUnlock()

	// Increment
	stats.mu.Lock()
	stats.TotalRenders++
	stats.ComputeRenders++
	stats.mu.Unlock()

	stats.mu.RLock()
	assert.Equal(t, int64(1), stats.TotalRenders)
	assert.Equal(t, int64(1), stats.ComputeRenders)
	stats.mu.RUnlock()
}

// =============================================================================
// BenchmarkResult Tests
// =============================================================================

func TestBenchmarkResult(t *testing.T) {
	result := &BenchmarkResult{
		EngineType:     LayoutEngineCompute,
		LayoutDuration: 1000,
		PaintDuration:  500,
		TotalDuration:  1500,
		CacheHits:      10,
		CacheMisses:    2,
	}

	assert.Equal(t, LayoutEngineCompute, result.EngineType)
	assert.Equal(t, int64(1000), result.LayoutDuration.Nanoseconds())
	assert.NoError(t, result.Error)
}

// =============================================================================
// Helper function tests
// =============================================================================

func TestAbs(t *testing.T) {
	assert.Equal(t, 5, abs(5))
	assert.Equal(t, 5, abs(-5))
	assert.Equal(t, 0, abs(0))
}
