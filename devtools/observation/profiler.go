// Package observation provides intelligent analysis for DevTools.
//
// This file implements the BehaviorProfiler for component behavior profiling.
package observation

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// Frequency indicates how frequently a component updates.
type Frequency int

const (
	FrequencyLow Frequency = iota    // < 10 updates/sec
	FrequencyMedium                   // 10-60 updates/sec
	FrequencyHigh                     // > 60 updates/sec
)

// String returns the string representation of the frequency.
func (f Frequency) String() string {
	switch f {
	case FrequencyHigh:
		return "High"
	case FrequencyMedium:
		return "Medium"
	default:
		return "Low"
	}
}

// PerformanceBaseline represents a performance baseline for a component.
type PerformanceBaseline struct {
	Mean      float64       // Mean duration in nanoseconds
	StdDev    float64       // Standard deviation
	Min       time.Duration // Minimum duration
	Max       time.Duration // Maximum duration
	P50       time.Duration // 50th percentile
	P95       time.Duration // 95th percentile
	P99       time.Duration // 99th percentile
	SampleCount int         // Number of samples
	UpdatedAt time.Time
}

// BehaviorProfile represents the behavior profile of a component.
type BehaviorProfile struct {
	NodeID          devtools.NodeID
	SampleCount     uint64

	// Update frequency
	UpdateFrequency Frequency
	UpdatesPerSec   float64

	// Trigger patterns
	TriggerEvents   []string          // Event types that trigger updates
	TriggeredBy     map[string]int    // Count of triggers by event type

	// Performance characteristics
	Durations       []time.Duration   // Sample durations (circular buffer)
	DurationPos     int
	DurationsCap    int

	// Baseline
	Baseline        *PerformanceBaseline

	// Anomaly detection
	AnomalyCount    int
	LastAnomalyTime time.Time
	AnomalyThreshold float64 // Sigma threshold for anomaly

	// State
	CreatedAt       time.Time
	LastUpdated     time.Time

	mu              sync.RWMutex
}

// addDuration adds a duration sample to the profile.
func (bp *BehaviorProfile) addDuration(d time.Duration) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	if cap(bp.Durations) == 0 {
		bp.Durations = make([]time.Duration, 0, bp.DurationsCap)
	}

	if len(bp.Durations) < bp.DurationsCap {
		bp.Durations = append(bp.Durations, d)
	} else {
		bp.Durations[bp.DurationPos] = d
		bp.DurationPos = (bp.DurationPos + 1) % bp.DurationsCap
	}

	bp.SampleCount++
	bp.LastUpdated = time.Now()
}

// updateBaseline recalculates the performance baseline.
func (bp *BehaviorProfile) updateBaseline() {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	if len(bp.Durations) == 0 {
		return
	}

	// Calculate statistics
	samples := make([]float64, len(bp.Durations))
	for i, d := range bp.Durations {
		samples[i] = float64(d.Nanoseconds())
	}

	// Sort for percentiles
	// Simple bubble sort for small datasets
	for i := 0; i < len(samples)-1; i++ {
		for j := 0; j < len(samples)-i-1; j++ {
			if samples[j] > samples[j+1] {
				samples[j], samples[j+1] = samples[j+1], samples[j]
			}
		}
	}

	n := len(samples)
	mean := mean(samples)
	variance := variance(samples, mean)

	bp.Baseline = &PerformanceBaseline{
		Mean:       mean,
		StdDev:     math.Sqrt(variance),
		Min:        bp.Durations[0],
		Max:        bp.Durations[n-1],
		P50:        time.Duration(int64(samples[n*50/100])),
		P95:        time.Duration(int64(samples[n*95/100])),
		P99:        time.Duration(int64(samples[n*99/100])),
		SampleCount: n,
		UpdatedAt:   time.Now(),
	}
}

// isAnomaly checks if a duration is anomalous based on baseline.
func (bp *BehaviorProfile) isAnomaly(d time.Duration) bool {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	if bp.Baseline == nil || bp.Baseline.StdDev == 0 {
		return false
	}

	duration := float64(d.Nanoseconds())
	zScore := math.Abs(duration - bp.Baseline.Mean) / bp.Baseline.StdDev
	return zScore > bp.AnomalyThreshold
}

// recordTrigger records that an update was triggered by a specific event.
func (bp *BehaviorProfile) recordTrigger(eventType string) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	if bp.TriggeredBy == nil {
		bp.TriggeredBy = make(map[string]int)
	}
	bp.TriggeredBy[eventType]++

	// Update trigger events list
	found := false
	for _, e := range bp.TriggerEvents {
		if e == eventType {
			found = true
			break
		}
	}
	if !found {
		bp.TriggerEvents = append(bp.TriggerEvents, eventType)
	}
}

// BehaviorProfiler creates and maintains behavior profiles for components.
type BehaviorProfiler struct {
	mu     sync.RWMutex
	enabled atomic.Bool

	// Component profiles
	profiles map[devtools.NodeID]*BehaviorProfile

	// Configuration
	sampleCapacity int           // Max duration samples per component
	anomalySigma   float64       // Sigma threshold for anomaly
	minSamples     int           // Minimum samples before baseline

	// Statistics
	totalProfiles  atomic.Uint64
	activeProfiles atomic.Uint64

	// Callback
	onAnomalyDetected func(nodeID devtools.NodeID, duration time.Duration, baseline *PerformanceBaseline)
}

// ProfilerConfig configures the BehaviorProfiler.
type ProfilerConfig struct {
	SampleCapacity int
	AnomalySigma   float64
	MinSamples     int
}

// DefaultProfilerConfig returns the default configuration.
func DefaultProfilerConfig() *ProfilerConfig {
	return &ProfilerConfig{
		SampleCapacity: 100,
		AnomalySigma:   2.5,  // 2.5 sigma
		MinSamples:     20,
	}
}

// NewBehaviorProfiler creates a new behavior profiler.
func NewBehaviorProfiler(cfg *ProfilerConfig) *BehaviorProfiler {
	if cfg == nil {
		cfg = DefaultProfilerConfig()
	}

	return &BehaviorProfiler{
		profiles:       make(map[devtools.NodeID]*BehaviorProfile),
		sampleCapacity: cfg.SampleCapacity,
		anomalySigma:   cfg.AnomalySigma,
		minSamples:     cfg.MinSamples,
	}
}

// Enable enables the profiler.
func (bp *BehaviorProfiler) Enable() {
	bp.enabled.Store(true)
}

// Disable disables the profiler.
func (bp *BehaviorProfiler) Disable() {
	bp.enabled.Store(false)
}

// IsEnabled returns whether the profiler is enabled.
func (bp *BehaviorProfiler) IsEnabled() bool {
	return bp.enabled.Load()
}

// OnAnomalyDetected sets a callback for anomaly detection events.
func (bp *BehaviorProfiler) OnAnomalyDetected(fn func(devtools.NodeID, time.Duration, *PerformanceBaseline)) {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.onAnomalyDetected = fn
}

// RecordComponentUpdate records a component update for profiling.
func (bp *BehaviorProfiler) RecordComponentUpdate(nodeID devtools.NodeID, duration time.Duration, triggerEvent string) {
	if !bp.enabled.Load() {
		return
	}

	bp.mu.Lock()
	profile, exists := bp.profiles[nodeID]
	if !exists {
		profile = &BehaviorProfile{
			NodeID:          nodeID,
			DurationsCap:    bp.sampleCapacity,
			AnomalyThreshold: bp.anomalySigma,
			CreatedAt:       time.Now(),
			LastUpdated:     time.Now(),
		}
		bp.profiles[nodeID] = profile
		bp.totalProfiles.Add(1)
	}
	bp.mu.Unlock()

	// Add duration sample
	profile.addDuration(duration)

	// Record trigger
	if triggerEvent != "" {
		profile.recordTrigger(triggerEvent)
	}

	// Update baseline if we have enough samples
	if profile.SampleCount >= uint64(bp.minSamples) && profile.SampleCount%10 == 0 {
		profile.updateBaseline()
		bp.activeProfiles.Add(1)
	}

	// Check for anomaly
	if profile.Baseline != nil && profile.isAnomaly(duration) {
		profile.AnomalyCount++
		profile.LastAnomalyTime = time.Now()

		if bp.onAnomalyDetected != nil {
			bp.onAnomalyDetected(nodeID, duration, profile.Baseline)
		}
	}
}

// GetProfile returns the behavior profile for a component.
func (bp *BehaviorProfiler) GetProfile(nodeID devtools.NodeID) *BehaviorProfile {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	return bp.profiles[nodeID]
}

// GetProfiles returns all behavior profiles.
func (bp *BehaviorProfiler) GetProfiles() []*BehaviorProfile {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	result := make([]*BehaviorProfile, 0, len(bp.profiles))
	for _, p := range bp.profiles {
		result = append(result, p)
	}

	return result
}

// UpdateFrequency calculates update frequency for a profile.
func (bp *BehaviorProfiler) UpdateFrequency(nodeID devtools.NodeID) Frequency {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	profile := bp.profiles[nodeID]
	if profile == nil || profile.SampleCount == 0 {
		return FrequencyLow
	}

	// Calculate updates per second
	elapsed := time.Since(profile.CreatedAt).Seconds()
	if elapsed <= 0 {
		elapsed = 1
	}
	ups := float64(profile.SampleCount) / elapsed

	profile.mu.Lock()
	profile.UpdatesPerSec = ups

	// Classify frequency
	var freq Frequency
	switch {
	case ups > 60:
		freq = FrequencyHigh
	case ups >= 10:
		freq = FrequencyMedium
	default:
		freq = FrequencyLow
	}
	profile.UpdateFrequency = freq
	profile.mu.Unlock()

	return freq
}

// GetStats returns profiler statistics.
func (bp *BehaviorProfiler) GetStats() ProfilerStats {
	bp.mu.RLock()
	totalProfiles := len(bp.profiles)
	activeCount := 0
	totalAnomalies := 0

	for _, p := range bp.profiles {
		if p.Baseline != nil {
			activeCount++
		}
		totalAnomalies += p.AnomalyCount
	}
	bp.mu.RUnlock()

	return ProfilerStats{
		TotalProfiles:    uint64(totalProfiles),
		ActiveProfiles:   uint64(activeCount),
		TotalAnomalies:   uint64(totalAnomalies),
	}
}

// Reset clears all profile data.
func (bp *BehaviorProfiler) Reset() {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.profiles = make(map[devtools.NodeID]*BehaviorProfile)
	bp.totalProfiles.Store(0)
	bp.activeProfiles.Store(0)
}

// ProfilerStats contains statistics about behavior profiling.
type ProfilerStats struct {
	TotalProfiles  uint64
	ActiveProfiles uint64
	TotalAnomalies uint64
}

// Helper functions for statistics

func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func variance(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		diff := v - mean
		sum += diff * diff
	}
	return sum / float64(len(values))
}
