// Package memory provides memory optimization utilities for DevTools.
package memory

import (
	"math"
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// =============================================================================
// Sampling Strategy
// =============================================================================

// SamplingStrategy determines when to collect detailed data.
type SamplingStrategy interface {
	// ShouldSample returns true if data should be collected for this frame.
	ShouldSample(frameID devtools.FrameID, context *SamplingContext) bool

	// GetSampleRate returns the current sampling rate (0-1).
	GetSampleRate() float64

	// Reset resets the strategy state.
	Reset()
}

// SamplingContext provides context for sampling decisions.
type SamplingContext struct {
	FrameRate       float64
	MutationCount   int
	LayoutCount     int
	RepaintCount    int
	PatternDetected bool
	MemoryPressure  float64 // 0-1, 1 = high pressure
	UserInteracting bool
}

// =============================================================================
// Adaptive Sampling Strategy
// =============================================================================

// AdaptiveStrategy adjusts sampling based on system conditions.
type AdaptiveStrategy struct {
	mu               sync.RWMutex
	minRate          float64
	maxRate          float64
	currentRate      float64
	targetRate       float64
	memoryThreshold  float64
	activityWindow   *ActivityWindow
	lastAdjustment   time.Time
	adjustInterval   time.Duration
	enabled          bool
}

// NewAdaptiveStrategy creates a new adaptive sampling strategy.
func NewAdaptiveStrategy(minRate, maxRate float64) *AdaptiveStrategy {
	return &AdaptiveStrategy{
		minRate:         minRate,
		maxRate:         maxRate,
		currentRate:     maxRate, // Start with maximum sampling
		targetRate:      maxRate,
		memoryThreshold: 0.8,     // 80% memory usage triggers reduction
		activityWindow:  NewActivityWindow(60),
		adjustInterval:  1 * time.Second,
		lastAdjustment:  time.Now(),
		enabled:         true,
	}
}

// ShouldSample determines if the current frame should be sampled.
func (s *AdaptiveStrategy) ShouldSample(frameID devtools.FrameID, context *SamplingContext) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.enabled {
		return false
	}

	// Adjust rate periodically
	if time.Since(s.lastAdjustment) > s.adjustInterval {
		s.adjustRate(context)
		s.lastAdjustment = time.Now()
	}

	// Record activity
	s.activityWindow.Record(context)

	// Always sample if user is interacting
	if context.UserInteracting {
		return true
	}

	// Always sample if pattern detected
	if context.PatternDetected {
		return true
	}

	// Sample based on current rate
	// Use frame ID for deterministic sampling
	samplePoint := float64(frameID%100) / 100.0
	return samplePoint < s.currentRate
}

// adjustRate adjusts the sampling rate based on conditions.
func (s *AdaptiveStrategy) adjustRate(context *SamplingContext) {
	// High memory pressure -> reduce sampling
	if context.MemoryPressure > s.memoryThreshold {
		s.targetRate = s.minRate
		return
	}

	// High activity -> increase sampling
	avgActivity := s.activityWindow.GetAverage()
	if avgActivity > 0.5 { // More than 50% activity
		s.targetRate = s.maxRate
		return
	}

	// Low activity -> moderate sampling
	s.targetRate = (s.minRate + s.maxRate) / 2

	// Gradually adjust current rate toward target
	rateDiff := s.targetRate - s.currentRate
	s.currentRate += rateDiff * 0.1 // 10% adjustment per iteration

	// Clamp to range
	s.currentRate = math.Max(s.minRate, math.Min(s.maxRate, s.currentRate))
}

// GetSampleRate returns the current sampling rate.
func (s *AdaptiveStrategy) GetSampleRate() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentRate
}

// SetTargetRate sets the target sampling rate.
func (s *AdaptiveStrategy) SetTargetRate(rate float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.targetRate = math.Max(s.minRate, math.Min(s.maxRate, rate))
}

// Reset resets the strategy.
func (s *AdaptiveStrategy) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.currentRate = s.maxRate
	s.targetRate = s.maxRate
	s.activityWindow.Reset()
	s.lastAdjustment = time.Now()
}

// Enable enables the strategy.
func (s *AdaptiveStrategy) Enable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = true
}

// Disable disables the strategy.
func (s *AdaptiveStrategy) Disable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = false
}

// =============================================================================
// Fixed Rate Sampling Strategy
// =============================================================================

// FixedRateStrategy samples at a constant rate.
type FixedRateStrategy struct {
	mu     sync.RWMutex
	rate   float64
	enabled bool
}

// NewFixedRateStrategy creates a fixed rate sampling strategy.
func NewFixedRateStrategy(rate float64) *FixedRateStrategy {
	return &FixedRateStrategy{
		rate:   math.Max(0, math.Min(1, rate)),
		enabled: true,
	}
}

// ShouldSample determines if the current frame should be sampled.
func (s *FixedRateStrategy) ShouldSample(frameID devtools.FrameID, context *SamplingContext) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.enabled {
		return false
	}

	samplePoint := float64(frameID%100) / 100.0
	return samplePoint < s.rate
}

// GetSampleRate returns the sampling rate.
func (s *FixedRateStrategy) GetSampleRate() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rate
}

// SetRate sets the sampling rate.
func (s *FixedRateStrategy) SetRate(rate float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rate = math.Max(0, math.Min(1, rate))
}

// Reset is a no-op for fixed rate strategy.
func (s *FixedRateStrategy) Reset() {}

// =============================================================================
// Priority-Based Sampling Strategy
// =============================================================================

// PriorityStrategy samples based on frame priority.
type PriorityStrategy struct {
	mu          sync.RWMutex
	threshold   int  // Priority threshold (0-100)
	enabled     bool
}

// NewPriorityStrategy creates a priority-based sampling strategy.
func NewPriorityStrategy(threshold int) *PriorityStrategy {
	return &PriorityStrategy{
		threshold: threshold,
		enabled:   true,
	}
}

// ShouldSample determines if the current frame should be sampled.
func (s *PriorityStrategy) ShouldSample(frameID devtools.FrameID, context *SamplingContext) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.enabled {
		return false
	}

	// Calculate priority
	priority := s.calculatePriority(context)
	return priority >= s.threshold
}

// calculatePriority calculates frame priority (0-100).
func (s *PriorityStrategy) calculatePriority(context *SamplingContext) int {
	priority := 0

	// User interaction = high priority
	if context.UserInteracting {
		priority += 50
	}

	// Pattern detected = high priority
	if context.PatternDetected {
		priority += 30
	}

	// High mutation count = medium priority
	if context.MutationCount > 10 {
		priority += 15
	} else if context.MutationCount > 5 {
		priority += 10
	} else if context.MutationCount > 0 {
		priority += 5
	}

	// High layout count = medium priority
	if context.LayoutCount > 5 {
		priority += 10
	}

	// Low frame rate = higher priority (need to investigate)
	if context.FrameRate < 30 {
		priority += 20
	} else if context.FrameRate < 50 {
		priority += 10
	}

	return priority
}

// GetSampleRate returns the current sampling rate (approximated).
func (s *PriorityStrategy) GetSampleRate() float64 {
	// For priority-based, return the threshold as a rate
	s.mu.RLock()
	defer s.mu.RUnlock()
	return float64(s.threshold) / 100.0
}

// SetThreshold sets the priority threshold.
func (s *PriorityStrategy) SetThreshold(threshold int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.threshold = threshold
}

// Reset is a no-op for priority strategy.
func (s *PriorityStrategy) Reset() {}

// =============================================================================
// Activity Window
// =============================================================================

// ActivityWindow tracks recent activity levels.
type ActivityWindow struct {
	buffer   []float64
	capacity int
	index    int
	full     bool
	mu       sync.RWMutex
}

// NewActivityWindow creates a new activity window.
func NewActivityWindow(capacity int) *ActivityWindow {
	return &ActivityWindow{
		buffer:   make([]float64, capacity),
		capacity: capacity,
		index:    0,
		full:     false,
	}
}

// Record records an activity level (0-1).
func (aw *ActivityWindow) Record(context *SamplingContext) {
	aw.mu.Lock()
	defer aw.mu.Unlock()

	// Calculate activity level
	activity := 0.0

	if context.MutationCount > 0 {
		activity += 0.3
	}
	if context.LayoutCount > 0 {
		activity += 0.3
	}
	if context.RepaintCount > 0 {
		activity += 0.2
	}
	if context.UserInteracting {
		activity += 0.2
	}

	activity = math.Min(1.0, activity)

	aw.buffer[aw.index] = activity
	aw.index = (aw.index + 1) % aw.capacity

	if aw.index == 0 {
		aw.full = true
	}
}

// GetAverage returns the average activity level.
func (aw *ActivityWindow) GetAverage() float64 {
	aw.mu.RLock()
	defer aw.mu.RUnlock()

	count := aw.capacity
	if !aw.full {
		count = aw.index
	}

	if count == 0 {
		return 0
	}

	sum := 0.0
	for i := 0; i < count; i++ {
		sum += aw.buffer[i]
	}

	return sum / float64(count)
}

// Reset clears the activity window.
func (aw *ActivityWindow) Reset() {
	aw.mu.Lock()
	defer aw.mu.Unlock()

	aw.buffer = make([]float64, aw.capacity)
	aw.index = 0
	aw.full = false
}

// =============================================================================
// Sampling Manager
// =============================================================================

// SamplingManager manages sampling strategies.
type SamplingManager struct {
	mu        sync.RWMutex
	strategy  SamplingStrategy
	decisions map[devtools.FrameID]bool // Cache of decisions
	stats     SamplingStats
}

// SamplingStats tracks sampling statistics.
type SamplingStats struct {
	TotalFrames    int `json:"total_frames"`
	SampledFrames  int `json:"sampled_frames"`
	SkippedFrames  int `json:"skipped_frames"`
	SampleRate     float64 `json:"sample_rate"`
}

// NewSamplingManager creates a new sampling manager.
func NewSamplingManager(strategy SamplingStrategy) *SamplingManager {
	return &SamplingManager{
		strategy:  strategy,
		decisions: make(map[devtools.FrameID]bool),
	}
}

// ShouldSample determines if the current frame should be sampled.
func (sm *SamplingManager) ShouldSample(frameID devtools.FrameID, context *SamplingContext) bool {
	shouldSample := sm.strategy.ShouldSample(frameID, context)

	sm.mu.Lock()
	sm.decisions[frameID] = shouldSample
	sm.stats.TotalFrames++
	if shouldSample {
		sm.stats.SampledFrames++
	} else {
		sm.stats.SkippedFrames++
	}
	if sm.stats.TotalFrames > 0 {
		sm.stats.SampleRate = float64(sm.stats.SampledFrames) / float64(sm.stats.TotalFrames)
	}
	sm.mu.Unlock()

	return shouldSample
}

// SetStrategy changes the sampling strategy.
func (sm *SamplingManager) SetStrategy(strategy SamplingStrategy) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.strategy = strategy
}

// GetStats returns sampling statistics.
func (sm *SamplingManager) GetStats() SamplingStats {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.stats
}

// Reset clears all statistics and decisions.
func (sm *SamplingManager) Reset() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.decisions = make(map[devtools.FrameID]bool)
	sm.stats = SamplingStats{}
	sm.strategy.Reset()
}
