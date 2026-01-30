// Package v1 provides pure statistics collection for DevTools.
//
// This file implements the metrics collector with pure statistics (no judgments).
package v1

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// MetricsSnapshot represents a snapshot of collected metrics.
type MetricsSnapshot struct {
	Timestamp       time.Time
	TotalFrames     uint64
	TotalMutations  uint64
	TotalLayouts    uint64
	TotalRepaints   uint64
	TotalEvents     uint64
	ComponentCount  int
	ActiveDuration  time.Duration
}

// ComponentMetrics represents metrics for a single component (pure statistics).
type ComponentMetrics struct {
	NodeID         devtools.NodeID
	MutationCount  uint64
	LayoutCount    uint64
	RepaintCount   uint64
	FirstSeen      time.Time
	LastSeen       time.Time
	TotalDuration  time.Duration
	AvgDuration    time.Duration
	MaxDuration    time.Duration
	MinDuration    time.Duration
}

// MetricsCollector collects metrics without making judgments.
type MetricsCollector struct {
	mu    sync.RWMutex
	level *LevelController

	// Level 1+: Atomic counters (zero lock overhead)
	totalFrames    atomic.Uint64
	totalMutations atomic.Uint64
	totalLayouts   atomic.Uint64
	totalRepaints  atomic.Uint64
	totalEvents    atomic.Uint64

	// Level 1+: Component-level statistics
	componentStats map[devtools.NodeID]*ComponentMetrics

	// Timing
	startTime      time.Time
	lastUpdateTime time.Time
}

// NewMetricsCollector creates a new metrics collector.
func NewMetricsCollector(level *LevelController) *MetricsCollector {
	return &MetricsCollector{
		level:         level,
		componentStats: make(map[devtools.NodeID]*ComponentMetrics),
		startTime:     time.Now(),
		lastUpdateTime: time.Now(),
	}
}

// RecordFrame records a frame (very fast - atomic only).
func (mc *MetricsCollector) RecordFrame(entry *devtools.FrameEntry) {
	if !mc.level.ShouldCollectBasicStats() {
		return
	}

	mc.totalFrames.Add(1)

	// Track component-level stats at enhanced level
	if mc.level.ShouldCollectEnhancedStats() && entry != nil {
		mc.updateComponentStats(entry)
	}

	mc.lastUpdateTime = time.Now()
}

// RecordMutation records a mutation.
func (mc *MetricsCollector) RecordMutation(nodeID devtools.NodeID) {
	if !mc.level.ShouldCollectBasicStats() {
		return
	}

	mc.totalMutations.Add(1)

	// Update component stats
	if mc.level.ShouldCollectEnhancedStats() {
		mc.mu.Lock()
		stat := mc.getOrCreateStat(nodeID)
		stat.MutationCount++
		stat.LastSeen = time.Now()
		mc.mu.Unlock()
	}
}

// RecordLayout records a layout operation.
func (mc *MetricsCollector) RecordLayout(nodeID devtools.NodeID) {
	if !mc.level.ShouldCollectBasicStats() {
		return
	}

	mc.totalLayouts.Add(1)

	// Update component stats
	if mc.level.ShouldCollectEnhancedStats() {
		mc.mu.Lock()
		stat := mc.getOrCreateStat(nodeID)
		stat.LayoutCount++
		stat.LastSeen = time.Now()
		mc.mu.Unlock()
	}
}

// RecordRepaint records a repaint operation.
func (mc *MetricsCollector) RecordRepaint(nodeID devtools.NodeID) {
	if !mc.level.ShouldCollectBasicStats() {
		return
	}

	mc.totalRepaints.Add(1)

	// Update component stats
	if mc.level.ShouldCollectEnhancedStats() {
		mc.mu.Lock()
		stat := mc.getOrCreateStat(nodeID)
		stat.RepaintCount++
		stat.LastSeen = time.Now()
		mc.mu.Unlock()
	}
}

// RecordEvent records an event.
func (mc *MetricsCollector) RecordEvent(eventType string) {
	if !mc.level.ShouldCollectBasicStats() {
		return
	}

	mc.totalEvents.Add(1)
	mc.lastUpdateTime = time.Now()
}

// updateComponentStats updates component statistics from a frame entry.
func (mc *MetricsCollector) updateComponentStats(entry *devtools.FrameEntry) {
	// In a real implementation, this would extract component-level timings
	// For now, we track at the frame level
}

// getOrCreateStat gets or creates a component stat entry.
func (mc *MetricsCollector) getOrCreateStat(nodeID devtools.NodeID) *ComponentMetrics {
	stat, exists := mc.componentStats[nodeID]
	if !exists {
		stat = &ComponentMetrics{
			NodeID:    nodeID,
			FirstSeen: time.Now(),
			LastSeen:  time.Now(),
		}
		mc.componentStats[nodeID] = stat
	}
	return stat
}

// GetSnapshot returns a snapshot of current metrics.
func (mc *MetricsCollector) GetSnapshot() *MetricsSnapshot {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return &MetricsSnapshot{
		Timestamp:      time.Now(),
		TotalFrames:    mc.totalFrames.Load(),
		TotalMutations: mc.totalMutations.Load(),
		TotalLayouts:   mc.totalLayouts.Load(),
		TotalRepaints:  mc.totalRepaints.Load(),
		TotalEvents:    mc.totalEvents.Load(),
		ComponentCount: len(mc.componentStats),
		ActiveDuration: time.Since(mc.startTime),
	}
}

// GetComponentMetrics returns metrics for a specific component.
func (mc *MetricsCollector) GetComponentMetrics(nodeID devtools.NodeID) *ComponentMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if stat, exists := mc.componentStats[nodeID]; exists {
		// Return a copy
		copy := *stat
		return &copy
	}
	return nil
}

// GetAllComponentMetrics returns all component metrics.
func (mc *MetricsCollector) GetAllComponentMetrics() []*ComponentMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make([]*ComponentMetrics, 0, len(mc.componentStats))
	for _, stat := range mc.componentStats {
		copy := *stat
		result = append(result, &copy)
	}
	return result
}

// Reset clears all collected metrics.
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.totalFrames.Store(0)
	mc.totalMutations.Store(0)
	mc.totalLayouts.Store(0)
	mc.totalRepaints.Store(0)
	mc.totalEvents.Store(0)

	mc.componentStats = make(map[devtools.NodeID]*ComponentMetrics)
	mc.startTime = time.Now()
	mc.lastUpdateTime = time.Now()
}
