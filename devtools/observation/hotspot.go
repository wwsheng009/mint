// Package observation provides intelligent analysis for DevTools.
//
// This file implements the HotspotDetector for identifying performance bottlenecks.
package observation

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// HotspotSeverity indicates the severity level of a hotspot.
type HotspotSeverity int

const (
	HotspotSeverityNone     HotspotSeverity = iota
	HotspotSeverityWarning                 // > 16.67ms
	HotspotSeverityCritical                // > 33.33ms
)

// String returns the string representation of the severity.
func (s HotspotSeverity) String() string {
	switch s {
	case HotspotSeverityWarning:
		return "Warning"
	case HotspotSeverityCritical:
		return "Critical"
	default:
		return "None"
	}
}

// FrameMetrics captures performance metrics for a single frame.
type FrameMetrics struct {
	FrameID    devtools.FrameID
	Duration   time.Duration
	LayoutTime time.Duration
	PaintTime  time.Duration
	SlowNodes  []devtools.NodeID
	Timestamp  time.Time
}

// ComponentHotspot tracks hotspot statistics for a single component.
type ComponentHotspot struct {
	NodeID         devtools.NodeID
	FrameCount     uint64
	TotalTime      time.Duration
	AvgTime        time.Duration
	MaxTime        time.Duration
	LastSlowTime   time.Time
	Severity       HotspotSeverity
	LastUpdated    time.Time

	// EWMA (Exponentially Weighted Moving Average) for smoothing
	ewmaTime       float64
	ewmaAlpha      float64  // Smoothing factor (0.0-1.0)
}

// updateEWMATime updates the EWMA time with a new measurement.
func (ch *ComponentHotspot) updateEWMATime(duration time.Duration) {
	dur := float64(duration.Nanoseconds())
	if ch.ewmaTime == 0 {
		ch.ewmaTime = dur
	} else {
		ch.ewmaTime = ch.ewmaAlpha*dur + (1-ch.ewmaAlpha)*ch.ewmaTime
	}
	ch.AvgTime = time.Duration(int64(ch.ewmaTime))
}

// HotspotDetector identifies performance bottlenecks in components.
type HotspotDetector struct {
	mu     sync.RWMutex
	enabled atomic.Bool

	// Frame metrics ring buffer
	frameBuffer *ringBuffer[FrameMetrics]
	bufferSize  int

	// Component hotspot tracking
	componentStats map[devtools.NodeID]*ComponentHotspot

	// Configuration
	slowFrameThreshold     time.Duration // Default: 16.67ms (60fps)
	hotComponentThreshold  time.Duration // Default: 5ms
	criticalThreshold      time.Duration // Default: 33.33ms
	ewmaAlpha              float64        // Default: 0.2

	// Statistics
	totalFrames        atomic.Uint64
	slowFrames         atomic.Uint64
	criticalFrames     atomic.Uint64

	// Callback
	onHotspotDetected func(*ComponentHotspot)
}

// HotspotConfig configures the HotspotDetector.
type HotspotConfig struct {
	SlowFrameThreshold    time.Duration
	HotComponentThreshold time.Duration
	CriticalThreshold     time.Duration
	BufferSize           int
	EWMAAlpha            float64
}

// DefaultHotspotConfig returns the default configuration.
func DefaultHotspotConfig() *HotspotConfig {
	return &HotspotConfig{
		SlowFrameThreshold:    time.Duration(16667000), // 16.67ms
		HotComponentThreshold: time.Duration(5000000),  // 5ms
		CriticalThreshold:     time.Duration(33333000), // 33.33ms
		BufferSize:           128,
		EWMAAlpha:            0.2,
	}
}

// NewHotspotDetector creates a new hotspot detector.
func NewHotspotDetector(cfg *HotspotConfig) *HotspotDetector {
	if cfg == nil {
		cfg = DefaultHotspotConfig()
	}

	return &HotspotDetector{
		frameBuffer:            newRingBuffer[FrameMetrics](cfg.BufferSize),
		bufferSize:             cfg.BufferSize,
		componentStats:         make(map[devtools.NodeID]*ComponentHotspot),
		slowFrameThreshold:     cfg.SlowFrameThreshold,
		hotComponentThreshold:  cfg.HotComponentThreshold,
		criticalThreshold:      cfg.CriticalThreshold,
		ewmaAlpha:              cfg.EWMAAlpha,
	}
}

// Enable enables the hotspot detector.
func (hd *HotspotDetector) Enable() {
	hd.enabled.Store(true)
}

// Disable disables the hotspot detector.
func (hd *HotspotDetector) Disable() {
	hd.enabled.Store(false)
}

// IsEnabled returns whether the detector is enabled.
func (hd *HotspotDetector) IsEnabled() bool {
	return hd.enabled.Load()
}

// OnHotspotDetected sets a callback for hotspot detection events.
func (hd *HotspotDetector) OnHotspotDetected(fn func(*ComponentHotspot)) {
	hd.mu.Lock()
	defer hd.mu.Unlock()
	hd.onHotspotDetected = fn
}

// ProcessFrame processes a frame and detects hotspots.
func (hd *HotspotDetector) ProcessFrame(entry *devtools.FrameEntry) {
	if !hd.enabled.Load() {
		return
	}

	// Create frame metrics
	metrics := &FrameMetrics{
		FrameID:    entry.FrameID,
		Duration:   entry.Duration,
		LayoutTime: entry.LayoutTime,
		PaintTime:  entry.PaintTime,
		Timestamp:  time.Now(),
	}

	hd.totalFrames.Add(1)

	// Check for slow frame
	isSlow := entry.Duration > hd.slowFrameThreshold
	isCritical := entry.Duration > hd.criticalThreshold

	if isSlow {
		hd.slowFrames.Add(1)
	}
	if isCritical {
		hd.criticalFrames.Add(1)
	}

	// Store in ring buffer
	hd.frameBuffer.Push(*metrics)

	// Analyze component-level data
	// In real implementation, this would extract component timings from the frame
	// For now, we track frame-level as component-level
	hd.analyzeFrameMetrics(metrics, isSlow, isCritical)
}

// analyzeFrameMetrics analyzes frame metrics and updates component stats.
func (hd *HotspotDetector) analyzeFrameMetrics(metrics *FrameMetrics, isSlow, isCritical bool) {
	hd.mu.Lock()
	defer hd.mu.Unlock()

	// For now, use frame ID as a proxy for component analysis
	// In real implementation, you'd extract actual component timings

	// Create/update hotspot entry for the frame
	nodeID := devtools.NodeID(fmt.Sprintf("frame-%d", metrics.FrameID)) // Placeholder
	hotspot, exists := hd.componentStats[nodeID]

	if !exists {
		hotspot = &ComponentHotspot{
			NodeID:    nodeID,
			ewmaAlpha: hd.ewmaAlpha,
		}
		hd.componentStats[nodeID] = hotspot
	}

	// Update statistics
	hotspot.FrameCount++
	hotspot.TotalTime += metrics.Duration
	hotspot.updateEWMATime(metrics.Duration)

	if metrics.Duration > hotspot.MaxTime {
		hotspot.MaxTime = metrics.Duration
	}

	// Update severity and last slow time
	if isCritical {
		hotspot.Severity = HotspotSeverityCritical
		hotspot.LastSlowTime = time.Now()
	} else if isSlow {
		if hotspot.Severity != HotspotSeverityCritical {
			hotspot.Severity = HotspotSeverityWarning
		}
		hotspot.LastSlowTime = time.Now()
	}

	hotspot.LastUpdated = time.Now()

	// Trigger callback if significantly slow
	if isSlow && hd.onHotspotDetected != nil {
		hd.onHotspotDetected(hotspot)
	}
}

// RecordComponentTime records a specific component's rendering time.
func (hd *HotspotDetector) RecordComponentTime(nodeID devtools.NodeID, duration time.Duration) {
	if !hd.enabled.Load() {
		return
	}

	hd.mu.Lock()
	defer hd.mu.Unlock()

	hotspot, exists := hd.componentStats[nodeID]
	if !exists {
		hotspot = &ComponentHotspot{
			NodeID:    nodeID,
			ewmaAlpha: hd.ewmaAlpha,
		}
		hd.componentStats[nodeID] = hotspot
	}

	hotspot.FrameCount++
	hotspot.TotalTime += duration
	hotspot.updateEWMATime(duration)

	if duration > hotspot.MaxTime {
		hotspot.MaxTime = duration
	}

	// Check against threshold
	if duration > hd.hotComponentThreshold {
		if duration > hd.criticalThreshold {
			hotspot.Severity = HotspotSeverityCritical
		} else if hotspot.Severity != HotspotSeverityCritical {
			hotspot.Severity = HotspotSeverityWarning
		}
		hotspot.LastSlowTime = time.Now()
	}

	hotspot.LastUpdated = time.Now()

	// Trigger callback
	if duration > hd.hotComponentThreshold && hd.onHotspotDetected != nil {
		hd.onHotspotDetected(hotspot)
	}
}

// GetHotspots returns all detected hotspots sorted by severity.
func (hd *HotspotDetector) GetHotspots() []*ComponentHotspot {
	hd.mu.RLock()
	defer hd.mu.RUnlock()

	result := make([]*ComponentHotspot, 0, len(hd.componentStats))
	for _, h := range hd.componentStats {
		if h.Severity != HotspotSeverityNone {
			result = append(result, h)
		}
	}

	// Sort by severity (critical first) then by avg time
	// Simple bubble sort for small slices
	for i := 0; i < len(result)-1; i++ {
		for j := 0; j < len(result)-i-1; j++ {
			if result[j].Severity < result[j+1].Severity ||
				(result[j].Severity == result[j+1].Severity && result[j].AvgTime < result[j+1].AvgTime) {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}

	return result
}

// GetHotspot returns hotspot info for a specific component.
func (hd *HotspotDetector) GetHotspot(nodeID devtools.NodeID) *ComponentHotspot {
	hd.mu.RLock()
	defer hd.mu.RUnlock()

	return hd.componentStats[nodeID]
}

// GetFrameMetrics returns recent frame metrics.
func (hd *HotspotDetector) GetFrameMetrics(n int) []FrameMetrics {
	return hd.frameBuffer.GetLastN(n)
}

// GetStats returns detector statistics.
func (hd *HotspotDetector) GetStats() HotspotStats {
	return HotspotStats{
		TotalFrames:    hd.totalFrames.Load(),
		SlowFrames:     hd.slowFrames.Load(),
		CriticalFrames: hd.criticalFrames.Load(),
		TotalComponents: uint64(len(hd.componentStats)),
		HotComponentCount: hd.countHotComponents(),
	}
}

func (hd *HotspotDetector) countHotComponents() uint64 {
	hd.mu.RLock()
	defer hd.mu.RUnlock()

	var count uint64
	for _, h := range hd.componentStats {
		if h.Severity != HotspotSeverityNone {
			count++
		}
	}
	return count
}

// Reset clears all hotspot data.
func (hd *HotspotDetector) Reset() {
	hd.mu.Lock()
	defer hd.mu.Unlock()

	hd.componentStats = make(map[devtools.NodeID]*ComponentHotspot)
	hd.frameBuffer.Clear()
	hd.totalFrames.Store(0)
	hd.slowFrames.Store(0)
	hd.criticalFrames.Store(0)
}

// HotspotStats contains statistics about hotspot detection.
type HotspotStats struct {
	TotalFrames      uint64
	SlowFrames       uint64
	CriticalFrames   uint64
	TotalComponents  uint64
	HotComponentCount uint64
}

// SlowFrameRate returns the percentage of slow frames.
func (s HotspotStats) SlowFrameRate() float64 {
	if s.TotalFrames == 0 {
		return 0
	}
	return float64(s.SlowFrames) / float64(s.TotalFrames) * 100
}

// ringBuffer is a simple ring buffer implementation.
type ringBuffer[T any] struct {
	buffer []T
	size   int
	writePos uint32
	count   uint32
	mu      sync.RWMutex
}

func newRingBuffer[T any](size int) *ringBuffer[T] {
	return &ringBuffer[T]{
		buffer: make([]T, size),
		size:   size,
	}
}

func (rb *ringBuffer[T]) Push(item T) {
	rb.mu.Lock()
	pos := rb.writePos % uint32(rb.size)
	rb.buffer[pos] = item
	rb.writePos++
	if rb.count < uint32(rb.size) {
		rb.count++
	}
	rb.mu.Unlock()
}

func (rb *ringBuffer[T]) GetLastN(n int) []T {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if n <= 0 || rb.count == 0 {
		return nil
	}

	if n > int(rb.count) {
		n = int(rb.count)
	}

	result := make([]T, n)

	if rb.count < uint32(rb.size) {
		// No wraparound yet, just copy from beginning
		copy(result, rb.buffer[:rb.count])
		return result
	}

	// Handle wraparound
	pos := int(rb.writePos) % rb.size
	start := pos - n
	if start < 0 {
		// Wrapped around: need to copy from two segments
		firstLen := -start
		copy(result, rb.buffer[rb.size+start:])
		copy(result[firstLen:], rb.buffer[:pos])
	} else {
		copy(result, rb.buffer[start:pos])
	}

	return result
}

func (rb *ringBuffer[T]) Clear() {
	rb.mu.Lock()
	rb.writePos = 0
	rb.count = 0
	rb.mu.Unlock()
}
