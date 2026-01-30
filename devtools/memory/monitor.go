// Package memory provides memory optimization utilities for DevTools.
package memory

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// =============================================================================
// Memory Monitor
// =============================================================================

// Monitor tracks memory usage and provides alerts.
type Monitor struct {
	mu                sync.RWMutex
	enabled           bool
	alertThreshold    float64 // Percentage (0-1) for alerts
	warningThreshold  float64 // Percentage (0-1) for warnings
	criticalThreshold float64 // Percentage (0-1) for critical alerts
	sampleInterval    time.Duration
	lastSample        MemorySnapshot
	history           []MemorySnapshot
	maxHistory        int
	alertCallback     func(MemoryAlert)
	stopChan          chan struct{}
}

// MemorySnapshot represents a point-in-time memory snapshot.
type MemorySnapshot struct {
	Timestamp       time.Time `json:"timestamp"`
	AllocMB         float64   `json:"alloc_mb"`
	TotalAllocMB    float64   `json:"total_alloc_mb"`
	SysMB           float64   `json:"sys_mb"`
	HeapAllocMB     float64   `json:"heap_alloc_mb"`
	HeapSysMB       float64   `json:"heap_sys_mb"`
	HeapIdleMB      float64   `json:"heap_idle_mb"`
	HeapInUseMB     float64   `json:"heap_inuse_mb"`
	HeapReleasedMB  float64   `json:"heap_released_mb"`
	HeapObjects     uint64    `json:"heap_objects"`
	StackInUseMB    float64   `json:"stack_inuse_mb"`
	StackMB         float64   `json:"stack_mb"`
	NumGC           uint32    `json:"num_gc"`
	NumGoroutine    int       `json:"num_goroutine"`
	NextGC          uint64    `json:"next_gc"`
	LastGC          time.Time `json:"last_gc,omitempty"`
	PauseTotalNs    uint64    `json:"pause_total_ns"`
}

// MemoryAlert represents a memory alert.
type MemoryAlert struct {
	Level       AlertLevel `json:"level"`
	Timestamp   time.Time  `json:"timestamp"`
	Snapshot    MemorySnapshot `json:"snapshot"`
	Message     string     `json:"message"`
	UsagePercent float64   `json:"usage_percent"`
}

// AlertLevel represents the severity of an alert.
type AlertLevel int

const (
	AlertInfo AlertLevel = iota
	AlertWarning
	AlertCritical
)

func (al AlertLevel) String() string {
	switch al {
	case AlertCritical:
		return "critical"
	case AlertWarning:
		return "warning"
	default:
		return "info"
	}
}

// NewMonitor creates a new memory monitor.
func NewMonitor() *Monitor {
	return &Monitor{
		enabled:           true,
		alertThreshold:    0.9,  // 90%
		warningThreshold:  0.75, // 75%
		criticalThreshold: 0.95, // 95%
		sampleInterval:    5 * time.Second,
		history:           make([]MemorySnapshot, 0, 100),
		maxHistory:        100,
		stopChan:          make(chan struct{}),
	}
}

// Enable enables memory monitoring.
func (m *Monitor) Enable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = true
}

// Disable disables memory monitoring.
func (m *Monitor) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = false
}

// IsEnabled returns true if monitoring is enabled.
func (m *Monitor) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

// SetThresholds sets the alert thresholds.
func (m *Monitor) SetThresholds(warning, alert, critical float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.warningThreshold = warning
	m.alertThreshold = alert
	m.criticalThreshold = critical
}

// SetSampleInterval sets the sampling interval.
func (m *Monitor) SetSampleInterval(interval time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sampleInterval = interval
}

// SetAlertCallback sets the callback for alerts.
func (m *Monitor) SetAlertCallback(callback func(MemoryAlert)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alertCallback = callback
}

// =============================================================================
// Memory Sampling
// =============================================================================

// Sample takes a memory snapshot.
func (m *Monitor) Sample() MemorySnapshot {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	snapshot := MemorySnapshot{
		Timestamp:       time.Now(),
		AllocMB:         bToMb(memStats.Alloc),
		TotalAllocMB:    bToMb(memStats.TotalAlloc),
		SysMB:           bToMb(memStats.Sys),
		HeapAllocMB:     bToMb(memStats.HeapAlloc),
		HeapSysMB:       bToMb(memStats.HeapSys),
		HeapIdleMB:      bToMb(memStats.HeapIdle),
		HeapInUseMB:     bToMb(memStats.HeapInuse),
		HeapReleasedMB:  bToMb(memStats.HeapReleased),
		HeapObjects:     memStats.HeapObjects,
		StackInUseMB:    bToMb(memStats.StackInuse),
		StackMB:         bToMb(memStats.StackSys),
		NumGC:           memStats.NumGC,
		NumGoroutine:    runtime.NumGoroutine(),
		NextGC:          memStats.NextGC,
		PauseTotalNs:    memStats.PauseTotalNs,
	}

	if memStats.LastGC > 0 {
		snapshot.LastGC = time.Unix(0, int64(memStats.LastGC))
	}

	m.mu.Lock()
	m.lastSample = snapshot

	// Add to history
	m.history = append(m.history, snapshot)
	if len(m.history) > m.maxHistory {
		m.history = m.history[1:]
	}
	m.mu.Unlock()

	return snapshot
}

// GetLastSample returns the most recent sample.
func (m *Monitor) GetLastSample() MemorySnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastSample
}

// GetHistory returns the sample history.
func (m *Monitor) GetHistory() []MemorySnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]MemorySnapshot, len(m.history))
	copy(result, m.history)
	return result
}

// GetStats returns current memory statistics.
func (m *Monitor) GetStats() *MemoryStats {
	snapshot := m.Sample()
	m.mu.RLock()
	defer m.mu.RUnlock()

	return &MemoryStats{
		Current:        snapshot,
		WarningThreshold: m.warningThreshold,
		AlertThreshold:   m.alertThreshold,
		CriticalThreshold: m.criticalThreshold,
		HistoryCount:    len(m.history),
	}
}

// =============================================================================
// Memory Pressure
// =============================================================================

// GetMemoryPressure returns the current memory pressure (0-1).
// Uses a heuristic based on heap usage and GC frequency.
func (m *Monitor) GetMemoryPressure() float64 {
	snapshot := m.Sample()

	// Estimate memory limit (64MB for typical terminal apps)
	// In reality, this varies by system
	const estimatedLimitMB = 64

	// Primary indicator: heap usage
	pressure := snapshot.HeapSysMB / estimatedLimitMB

	// Secondary indicator: GC frequency
	// Frequent GC suggests memory pressure
	if snapshot.NumGC > 0 {
		// More GCs = more pressure
		gcPressure := float64(snapshot.NumGC) / 100.0
		if gcPressure > 0.3 {
			pressure = max(pressure, gcPressure)
		}
	}

	// Clamp to 0-1
	if pressure > 1 {
		pressure = 1
	}
	if pressure < 0 {
		pressure = 0
	}

	return pressure
}

// IsMemoryPressureHigh returns true if memory pressure is above the threshold.
func (m *Monitor) IsMemoryPressureHigh(threshold float64) bool {
	return m.GetMemoryPressure() > threshold
}

// =============================================================================
// Continuous Monitoring
// =============================================================================

// Start begins continuous monitoring.
func (m *Monitor) Start() {
	ticker := time.NewTicker(m.sampleInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				if m.IsEnabled() {
					m.checkAndAlert()
				}
			case <-m.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop stops continuous monitoring.
func (m *Monitor) Stop() {
	close(m.stopChan)
}

// checkAndAlert checks memory and triggers alerts if needed.
func (m *Monitor) checkAndAlert() {
	snapshot := m.Sample()
	pressure := m.GetMemoryPressure()

	m.mu.RLock()
	callback := m.alertCallback
	warningThresh := m.warningThreshold
	alertThresh := m.alertThreshold
	criticalThresh := m.criticalThreshold
	m.mu.RUnlock()

	if callback == nil {
		return
	}

	if pressure >= criticalThresh {
		callback(MemoryAlert{
			Level:        AlertCritical,
			Timestamp:    time.Now(),
			Snapshot:     snapshot,
			Message:      fmt.Sprintf("Critical memory pressure: %.1f%%", pressure*100),
			UsagePercent: pressure,
		})
	} else if pressure >= alertThresh {
		callback(MemoryAlert{
			Level:        AlertWarning,
			Timestamp:    time.Now(),
			Snapshot:     snapshot,
			Message:      fmt.Sprintf("High memory pressure: %.1f%%", pressure*100),
			UsagePercent: pressure,
		})
	} else if pressure >= warningThresh {
		callback(MemoryAlert{
			Level:        AlertInfo,
			Timestamp:    time.Now(),
			Snapshot:     snapshot,
			Message:      fmt.Sprintf("Memory usage elevated: %.1f%%", pressure*100),
			UsagePercent: pressure,
		})
	}
}

// =============================================================================
// Memory Stats
// =============================================================================

// MemoryStats provides comprehensive memory statistics.
type MemoryStats struct {
	Current           MemorySnapshot `json:"current"`
	WarningThreshold  float64        `json:"warning_threshold"`
	AlertThreshold    float64        `json:"alert_threshold"`
	CriticalThreshold float64        `json:"critical_threshold"`
	HistoryCount      int            `json:"history_count"`
	AverageUsageMB    float64        `json:"average_usage_mb"`
	PeakUsageMB       float64        `json:"peak_usage_mb"`
	Trend             Trend          `json:"trend"`
}

// Trend represents the memory usage trend.
type Trend string

const (
	TrendStable   Trend = "stable"
	TrendRising   Trend = "rising"
	TrendFalling  Trend = "falling"
	TrendUnknown  Trend = "unknown"
)

// GetDetailedStats returns detailed memory statistics.
func (m *Monitor) GetDetailedStats() *MemoryStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &MemoryStats{
		Current:           m.lastSample,
		WarningThreshold:  m.warningThreshold,
		AlertThreshold:    m.alertThreshold,
		CriticalThreshold: m.criticalThreshold,
		HistoryCount:      len(m.history),
	}

	if len(m.history) > 0 {
		// Calculate average
		sum := 0.0
		peak := 0.0
		for _, snap := range m.history {
			sum += snap.HeapAllocMB
			if snap.HeapAllocMB > peak {
				peak = snap.HeapAllocMB
			}
		}
		stats.AverageUsageMB = sum / float64(len(m.history))
		stats.PeakUsageMB = peak

		// Determine trend
		stats.Trend = m.calculateTrend()
	}

	return stats
}

// calculateTrend determines the memory usage trend.
func (m *Monitor) calculateTrend() Trend {
	if len(m.history) < 3 {
		return TrendUnknown
	}

	// Compare recent samples to older samples
	recentCount := len(m.history) / 3
	if recentCount < 2 {
		recentCount = 2
	}

	// Average of recent samples
	recentSum := 0.0
	for i := len(m.history) - recentCount; i < len(m.history); i++ {
		recentSum += m.history[i].HeapAllocMB
	}
	recentAvg := recentSum / float64(recentCount)

	// Average of older samples
	olderSum := 0.0
	olderCount := len(m.history) - recentCount
	for i := 0; i < olderCount; i++ {
		olderSum += m.history[i].HeapAllocMB
	}
	olderAvg := olderSum / float64(olderCount)

	// Compare
	diff := recentAvg - olderAvg
	threshold := olderAvg * 0.1 // 10% change threshold

	if diff > threshold {
		return TrendRising
	} else if diff < -threshold {
		return TrendFalling
	}
	return TrendStable
}

// =============================================================================
// Memory Profiling
// =============================================================================

// StartHeapProfile starts heap profiling.
func (m *Monitor) StartHeapProfile(duration time.Duration) ([]byte, error) {
	// This would integrate with runtime/pprof
	// For now, return a summary
	snapshot := m.Sample()

	summary := map[string]interface{}{
		"timestamp":      snapshot.Timestamp,
		"heap_alloc_mb":  snapshot.HeapAllocMB,
		"heap_sys_mb":    snapshot.HeapSysMB,
		"heap_objects":   snapshot.HeapObjects,
		"num_goroutine":  snapshot.NumGoroutine,
		"num_gc":         snapshot.NumGC,
		"pause_total_ms": snapshot.PauseTotalNs / 1000000,
	}

	return json.MarshalIndent(summary, "", "  ")
}

// =============================================================================
// Utilities
// =============================================================================

func bToMb(b uint64) float64 {
	return float64(b) / 1024 / 1024
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
