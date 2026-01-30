// Package v1 provides pure statistics collection for DevTools.
//
// This file implements statistical analysis (TopN, percentiles) without judgments.
package v1

import (
	"sort"
	"sync"

	"github.com/wwsheng009/mint/devtools"
)

// MetricType represents the type of metric to analyze.
type MetricType int

const (
	MetricMutations MetricType = iota
	MetricLayouts
	MetricRepaints
	MetricEvents
)

// ComponentRank represents a component ranking.
type ComponentRank struct {
	NodeID   devtools.NodeID
	Value    uint64
	Percentile float64 // Percentile rank (0-100)
}

// PercentileValue represents a percentile value.
type PercentileValue struct {
	Percentile float64 // 0.0-1.0
	Value      uint64
}

// StatsAnalyzer provides statistical analysis (pure statistics, no judgments).
type StatsAnalyzer struct {
	mu     sync.RWMutex
	level  *LevelController
	metrics *MetricsCollector

	// Caching
	topCache       map[MetricType][]*ComponentRank
	topCacheSize   int
	dirty          bool
}

// NewStatsAnalyzer creates a new stats analyzer.
func NewStatsAnalyzer(level *LevelController, metrics *MetricsCollector) *StatsAnalyzer {
	return &StatsAnalyzer{
		level:       level,
		metrics:     metrics,
		topCache:    make(map[MetricType][]*ComponentRank),
		topCacheSize: 10,
		dirty:       true,
	}
}

// SetTopN sets the top N cache size.
func (sa *StatsAnalyzer) SetTopN(n int) {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	sa.topCacheSize = n
	sa.dirty = true
}

// GetTopN returns the top N components by metric type.
func (sa *StatsAnalyzer) GetTopN(metric MetricType, n int) []*ComponentRank {
	if !sa.level.ShouldCollectAdvancedStats() {
		return nil
	}

	sa.mu.Lock()
	defer sa.mu.Unlock()

	// Use cache if available and valid
	if !sa.dirty {
		if cached, exists := sa.topCache[metric]; exists {
			if len(cached) >= n {
				return cached[:n]
			}
		}
	}

	// Build rankings
	components := sa.metrics.GetAllComponentMetrics()
	ranks := make([]*ComponentRank, 0, len(components))

	for _, comp := range components {
		var value uint64
		switch metric {
		case MetricMutations:
			value = comp.MutationCount
		case MetricLayouts:
			value = comp.LayoutCount
		case MetricRepaints:
			value = comp.RepaintCount
		}

		if value > 0 {
			ranks = append(ranks, &ComponentRank{
				NodeID: comp.NodeID,
				Value:  value,
			})
		}
	}

	// Sort by value (descending)
	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i].Value > ranks[j].Value
	})

	// Calculate percentiles
	if len(ranks) > 0 {
		maxVal := ranks[0].Value
		for _, rank := range ranks {
			if maxVal > 0 {
				rank.Percentile = float64(rank.Value) / float64(maxVal) * 100
			}
		}
	}

	// Limit to top N
	if n > 0 && len(ranks) > n {
		ranks = ranks[:n]
	}

	// Update cache
	sa.topCache[metric] = ranks
	sa.dirty = false

	return ranks
}

// GetPercentiles calculates percentile values for a metric across all components.
func (sa *StatsAnalyzer) GetPercentiles(metric MetricType, percentiles []float64) []PercentileValue {
	if !sa.level.ShouldCollectAdvancedStats() {
		return nil
	}

	components := sa.metrics.GetAllComponentMetrics()
	if len(components) == 0 {
		return nil
	}

	// Extract values
	values := make([]uint64, 0, len(components))
	for _, comp := range components {
		var value uint64
		switch metric {
		case MetricMutations:
			value = comp.MutationCount
		case MetricLayouts:
			value = comp.LayoutCount
		case MetricRepaints:
			value = comp.RepaintCount
		}
		if value > 0 {
			values = append(values, value)
		}
	}

	if len(values) == 0 {
		return nil
	}

	// Sort
	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})

	// Calculate percentiles
	result := make([]PercentileValue, 0, len(percentiles))
	n := len(values)

	for _, p := range percentiles {
		if p < 0 || p > 1 {
			continue
		}
		idx := int(float64(n-1) * p)
		if idx < 0 {
			idx = 0
		}
		if idx >= n {
			idx = n - 1
		}
		result = append(result, PercentileValue{
			Percentile: p,
			Value:      values[idx],
		})
	}

	return result
}

// GetDistribution returns distribution statistics for a metric.
type Distribution struct {
	Count     int
	Min       uint64
	Max       uint64
	Mean      float64
	Median    uint64
	P90       uint64
	P95       uint64
	P99       uint64
	StdDev    float64
}

// GetDistribution calculates distribution statistics.
func (sa *StatsAnalyzer) GetDistribution(metric MetricType) *Distribution {
	if !sa.level.ShouldCollectAdvancedStats() {
		return nil
	}

	components := sa.metrics.GetAllComponentMetrics()
	if len(components) == 0 {
		return nil
	}

	// Extract values
	values := make([]uint64, 0, len(components))
	for _, comp := range components {
		var value uint64
		switch metric {
		case MetricMutations:
			value = comp.MutationCount
		case MetricLayouts:
			value = comp.LayoutCount
		case MetricRepaints:
			value = comp.RepaintCount
		}
		values = append(values, value)
	}

	if len(values) == 0 {
		return nil
	}

	// Sort
	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})

	n := len(values)
	min := values[0]
	max := values[n-1]

	// Calculate mean
	sum := uint64(0)
	for _, v := range values {
		sum += v
	}
	mean := float64(sum) / float64(n)

	// Calculate standard deviation
	variance := 0.0
	for _, v := range values {
		diff := float64(v) - mean
		variance += diff * diff
	}
	variance /= float64(n)
	stdDev := variance

	// Get percentiles
	percentiles := sa.GetPercentiles(metric, []float64{0.5, 0.90, 0.95, 0.99})
	median := uint64(0)
	p90 := uint64(0)
	p95 := uint64(0)
	p99 := uint64(0)

	if len(percentiles) >= 4 {
		median = percentiles[0].Value
		p90 = percentiles[1].Value
		p95 = percentiles[2].Value
		p99 = percentiles[3].Value
	}

	return &Distribution{
		Count:  n,
		Min:    min,
		Max:    max,
		Mean:   mean,
		Median: median,
		P90:    p90,
		P95:    p95,
		P99:    p99,
		StdDev: stdDev,
	}
}

// InvalidateCache invalidates the cache.
func (sa *StatsAnalyzer) InvalidateCache() {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	sa.dirty = true
}
