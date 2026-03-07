package intent

import (
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// Performance Monitoring Middleware
// =============================================================================

// PerformanceMiddleware tracks intent execution times.
type PerformanceMiddleware struct {
	mu     sync.RWMutex
	stats  map[string]*IntentPerformanceStats

	// Threshold triggers warning if duration exceeds this
	Threshold time.Duration

	// WarningFunc is called when threshold exceeded
	WarningFunc func(intentType string, duration time.Duration)
}

// IntentPerformanceStats stores performance statistics for an intent type.
type IntentPerformanceStats struct {
	Count       int
	TotalTime   time.Duration
	MinTime     time.Duration
	MaxTime     time.Duration
	ErrorCount  int
}

// NewPerformanceMiddleware creates a new performance middleware.
func NewPerformanceMiddleware(threshold time.Duration) *PerformanceMiddleware {
	return &PerformanceMiddleware{
		stats:     make(map[string]*IntentPerformanceStats),
		Threshold: threshold,
		WarningFunc: func(intentType string, duration time.Duration) {
			fmt.Printf("[PerformanceWarning] %s took %v (threshold: %v)\n",
				intentType, duration, threshold)
		},
	}
}

// BeforeEmit is a no-op.
func (m *PerformanceMiddleware) BeforeEmit(ctx *MiddlewareContext) error {
	// No-op needed
	return nil
}

// AfterEmit records performance statistics.
func (m *PerformanceMiddleware) AfterEmit(ctx *MiddlewareContext, result *EmitResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	intentType := ctx.Intent.IntentType()
	stats, ok := m.stats[intentType]
	if !ok {
		stats = &IntentPerformanceStats{
			MinTime: result.Duration,
			MaxTime: result.Duration,
		}
		m.stats[intentType] = stats
	}

	// Update statistics
	stats.Count++
	stats.TotalTime += result.Duration
	if result.Duration < stats.MinTime {
		stats.MinTime = result.Duration
	}
	if result.Duration > stats.MaxTime {
		stats.MaxTime = result.Duration
	}
	if result.Error != nil {
		stats.ErrorCount++
	}

	// Check threshold
	if m.Threshold > 0 && result.Duration > m.Threshold && m.WarningFunc != nil {
		m.WarningFunc(intentType, result.Duration)
	}
}

// GetStats returns performance statistics for an intent type.
func (m *PerformanceMiddleware) GetStats(intentType string) *IntentPerformanceStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if stats, ok := m.stats[intentType]; ok {
		// Return a copy
		return &IntentPerformanceStats{
			Count:      stats.Count,
			TotalTime:  stats.TotalTime,
			MinTime:    stats.MinTime,
			MaxTime:    stats.MaxTime,
			ErrorCount: stats.ErrorCount,
		}
	}
	return nil
}

// GetAllStats returns all performance statistics.
func (m *PerformanceMiddleware) GetAllStats() map[string]*IntentPerformanceStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*IntentPerformanceStats, len(m.stats))
	for k, v := range m.stats {
		result[k] = &IntentPerformanceStats{
			Count:      v.Count,
			TotalTime:  v.TotalTime,
			MinTime:    v.MinTime,
			MaxTime:    v.MaxTime,
			ErrorCount: v.ErrorCount,
		}
	}
	return result
}

// Reset clears all statistics.
func (m *PerformanceMiddleware) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats = make(map[string]*IntentPerformanceStats)
}

// GenerateReport generates a performance report.
func (m *PerformanceMiddleware) GenerateReport() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	report := "=== Intent Performance Report ===\n"
	report += fmt.Sprintf("Intent Types: %d\n\n", len(m.stats))

	for intentType, stats := range m.stats {
		avgTime := time.Duration(int64(stats.TotalTime) / int64(stats.Count))
		report += fmt.Sprintf(
			"Type: %s\n"+
				"  Count: %d\n"+
				"  Total: %v\n"+
				"  Average: %v\n"+
				"  Min: %v\n"+
				"  Max: %v\n"+
				"  Errors: %d\n\n",
			intentType, stats.Count, stats.TotalTime,
			avgTime, stats.MinTime, stats.MaxTime, stats.ErrorCount,
		)
	}

	return report
}

// =============================================================================
// Timing Middleware (Simple Timer)
// =============================================================================

// TimingMiddleware measures and logs intent execution times.
type TimingMiddleware struct {
	// Enable enables timing.
	Enable bool
}

// NewTimingMiddleware creates a new timing middleware.
func NewTimingMiddleware(enable bool) *TimingMiddleware {
	return &TimingMiddleware{Enable: enable}
}

// BeforeEmit is a no-op.
func (m *TimingMiddleware) BeforeEmit(ctx *MiddlewareContext) error {
	return nil
}

// AfterEmit logs execution time.
func (m *TimingMiddleware) AfterEmit(ctx *MiddlewareContext, result *EmitResult) {
	if !m.Enable {
		return
	}

	fmt.Printf("[Timing] %s took %v\n",
		ctx.Intent.IntentType(), result.Duration)
}
