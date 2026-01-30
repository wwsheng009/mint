// Package v1 provides pure statistics collection for DevTools.
//
// This file implements time series storage with fixed window (no unbounded growth).
package v1

import (
	"sort"
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// DataPoint represents a single data point in time series.
type DataPoint struct {
	FrameID   devtools.FrameID
	Timestamp time.Time
	Value     float64
}

// TimeSeries represents a fixed-size time series for a component.
type TimeSeries struct {
	NodeID    devtools.NodeID
	Points    []DataPoint
	writePos  int
	count     int
	capacity  int
	lastSeen  time.Time
}

// NewTimeSeries creates a new time series.
func NewTimeSeries(nodeID devtools.NodeID, capacity int) *TimeSeries {
	return &TimeSeries{
		NodeID:   nodeID,
		Points:   make([]DataPoint, capacity),
		capacity: capacity,
	}
}

// Add adds a data point to the time series.
func (ts *TimeSeries) Add(point DataPoint) {
	ts.Points[ts.writePos] = point
	ts.writePos = (ts.writePos + 1) % ts.capacity
	if ts.count < ts.capacity {
		ts.count++
	}
	ts.lastSeen = time.Now()
}

// GetLastN returns the last N data points.
func (ts *TimeSeries) GetLastN(n int) []DataPoint {
	if n <= 0 || ts.count == 0 {
		return nil
	}

	if n > ts.count {
		n = ts.count
	}

	result := make([]DataPoint, n)

	// Handle wraparound
	pos := ts.writePos - n
	if pos < 0 {
		// Wrapped around
		firstPart := ts.capacity + pos
		copy(result, ts.Points[firstPart:])
		copy(result[ts.capacity-firstPart:], ts.Points[:ts.writePos])
	} else {
		copy(result, ts.Points[pos:ts.writePos])
	}

	return result
}

// GetAll returns all data points.
func (ts *TimeSeries) GetAll() []DataPoint {
	return ts.GetLastN(ts.count)
}

// Percentile calculates the percentile value.
func (ts *TimeSeries) Percentile(p float64) float64 {
	if ts.count == 0 {
		return 0
	}

	points := ts.GetAll()
	if len(points) == 0 {
		return 0
	}

	// Sort by value
	sorted := make([]float64, len(points))
	for i, pt := range points {
		sorted[i] = pt.Value
	}
	sort.Float64s(sorted)

	idx := int(float64(len(sorted)-1) * p)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}

	return sorted[idx]
}

// Clear clears the time series.
func (ts *TimeSeries) Clear() {
	ts.writePos = 0
	ts.count = 0
	ts.lastSeen = time.Time{}
}

// TimeSeriesStore manages multiple time series with fixed window.
type TimeSeriesStore struct {
	mu         sync.RWMutex
	level      *LevelController
	windowSize int
	ttl        time.Duration
	series     map[devtools.NodeID]*TimeSeries
}

// NewTimeSeriesStore creates a new time series store.
func NewTimeSeriesStore(level *LevelController, windowSize int, ttl time.Duration) *TimeSeriesStore {
	return &TimeSeriesStore{
		level:      level,
		windowSize: windowSize,
		ttl:        ttl,
		series:     make(map[devtools.NodeID]*TimeSeries),
	}
}

// AddPoint adds a data point for a component.
func (tss *TimeSeriesStore) AddPoint(nodeID devtools.NodeID, value float64) {
	if !tss.level.ShouldCollectEnhancedStats() {
		return
	}

	tss.mu.Lock()
	defer tss.mu.Unlock()

	ts := tss.getOrCreate(nodeID)
	ts.Add(DataPoint{
		Timestamp: time.Now(),
		Value:     value,
	})
}

// AddFramePoint adds a data point with frame ID.
func (tss *TimeSeriesStore) AddFramePoint(nodeID devtools.NodeID, frameID devtools.FrameID, value float64) {
	if !tss.level.ShouldCollectEnhancedStats() {
		return
	}

	tss.mu.Lock()
	defer tss.mu.Unlock()

	ts := tss.getOrCreate(nodeID)
	ts.Add(DataPoint{
		FrameID:   frameID,
		Timestamp: time.Now(),
		Value:     value,
	})
}

// GetTimeSeries returns the time series for a component.
func (tss *TimeSeriesStore) GetTimeSeries(nodeID devtools.NodeID) []DataPoint {
	tss.mu.RLock()
	defer tss.mu.RUnlock()

	if ts, exists := tss.series[nodeID]; exists {
		return ts.GetAll()
	}
	return nil
}

// GetLastN returns the last N data points for a component.
func (tss *TimeSeriesStore) GetLastN(nodeID devtools.NodeID, n int) []DataPoint {
	tss.mu.RLock()
	defer tss.mu.RUnlock()

	if ts, exists := tss.series[nodeID]; exists {
		return ts.GetLastN(n)
	}
	return nil
}

// GetPercentile returns the percentile value for a component.
func (tss *TimeSeriesStore) GetPercentile(nodeID devtools.NodeID, percentile float64) float64 {
	tss.mu.RLock()
	defer tss.mu.RUnlock()

	if ts, exists := tss.series[nodeID]; exists {
		return ts.Percentile(percentile)
	}
	return 0
}

// getOrCreate gets or creates a time series for a node.
func (tss *TimeSeriesStore) getOrCreate(nodeID devtools.NodeID) *TimeSeries {
	ts, exists := tss.series[nodeID]
	if !exists {
		ts = NewTimeSeries(nodeID, tss.windowSize)
		tss.series[nodeID] = ts
	}
	return ts
}

// CleanupExpired removes expired time series.
func (tss *TimeSeriesStore) CleanupExpired() {
	tss.mu.Lock()
	defer tss.mu.Unlock()

	now := time.Now()
	for nodeID, ts := range tss.series {
		if now.Sub(ts.lastSeen) > tss.ttl {
			delete(tss.series, nodeID)
		}
	}
}

// GetAll returns all time series.
func (tss *TimeSeriesStore) GetAll() map[devtools.NodeID][]DataPoint {
	tss.mu.RLock()
	defer tss.mu.RUnlock()

	result := make(map[devtools.NodeID][]DataPoint)
	for nodeID, ts := range tss.series {
		result[nodeID] = ts.GetAll()
	}
	return result
}

// Clear clears all time series.
func (tss *TimeSeriesStore) Clear() {
	tss.mu.Lock()
	defer tss.mu.Unlock()

	tss.series = make(map[devtools.NodeID]*TimeSeries)
}

// Count returns the number of time series.
func (tss *TimeSeriesStore) Count() int {
	tss.mu.RLock()
	defer tss.mu.RUnlock()

	return len(tss.series)
}
