// Package v2 provides pattern detection and analysis for DevTools.
//
// This file implements the pattern detector for identifying behavioral patterns.
package v2

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// PatternDetector detects behavioral patterns in component updates.
type PatternDetector struct {
	mu     sync.RWMutex
	enabled atomic.Bool

	config *PatternDetectorConfig

	// Event history for pattern detection (fixed window)
	eventHistory   *EventRingBuffer
	historySize    int

	// Detected patterns
	patterns       map[devtools.NodeID][]*DetectedPattern
	patternCount   atomic.Uint64

	// Per-pattern state
	oscillationState map[devtools.NodeID]*OscillationState
	sameFieldState   map[devtools.NodeID]*SameFieldState
	cascadeState     map[devtools.NodeID]*CascadeState

	// Callbacks
	onPatternDetected func(*DetectedPattern)
}

// EventRingBuffer is a fixed-size ring buffer for events.
type EventRingBuffer struct {
	events []PatternEvent
	size   int
	writePos uint32
	count  uint32
	mu     sync.RWMutex
}

// NewEventRingBuffer creates a new event ring buffer.
func NewEventRingBuffer(size int) *EventRingBuffer {
	return &EventRingBuffer{
		events: make([]PatternEvent, size),
		size:   size,
	}
}

// Push adds an event to the buffer.
func (erb *EventRingBuffer) Push(event PatternEvent) {
	erb.mu.Lock()
	erb.events[erb.writePos%uint32(erb.size)] = event
	erb.writePos++
	if erb.count < uint32(erb.size) {
		erb.count++
	}
	erb.mu.Unlock()
}

// GetLastN returns the last N events for a node.
func (erb *EventRingBuffer) GetLastN(nodeID devtools.NodeID, n int) []PatternEvent {
	erb.mu.RLock()
	defer erb.mu.RUnlock()

	if n <= 0 || erb.count == 0 {
		return nil
	}

	if n > int(erb.count) {
		n = int(erb.count)
	}

	result := make([]PatternEvent, 0, n)

	// Scan backwards from most recent
	for i := int(erb.writePos) - 1; i >= 0 && len(result) < n; i-- {
		idx := i % erb.size
		event := erb.events[idx]
		if event.NodeID == nodeID {
			result = append(result, event)
		}
	}

	return result
}

// GetAll returns all events.
func (erb *EventRingBuffer) GetAll() []PatternEvent {
	erb.mu.RLock()
	defer erb.mu.RUnlock()

	result := make([]PatternEvent, erb.count)
	start := int(erb.writePos) - int(erb.count)
	if start < 0 {
		// Wrapped around
		firstPart := erb.events[start+erb.size:]
		copy(result, firstPart)
		copy(result[len(firstPart):], erb.events[:int(erb.writePos)])
	} else {
		copy(result, erb.events[start:int(erb.writePos)])
	}

	return result
}

// NewPatternDetector creates a new pattern detector.
func NewPatternDetector(cfg *PatternDetectorConfig) *PatternDetector {
	if cfg == nil {
		cfg = DefaultPatternDetectorConfig()
	}

	return &PatternDetector{
		config:          cfg,
		eventHistory:    NewEventRingBuffer(1000),
		historySize:     1000,
		patterns:        make(map[devtools.NodeID][]*DetectedPattern),
		oscillationState: make(map[devtools.NodeID]*OscillationState),
		sameFieldState:   make(map[devtools.NodeID]*SameFieldState),
		cascadeState:     make(map[devtools.NodeID]*CascadeState),
	}
}

// Enable enables the pattern detector.
func (pd *PatternDetector) Enable() {
	pd.enabled.Store(true)
}

// Disable disables the pattern detector.
func (pd *PatternDetector) Disable() {
	pd.enabled.Store(false)
}

// IsEnabled returns whether the detector is enabled.
func (pd *PatternDetector) IsEnabled() bool {
	return pd.enabled.Load()
}

// OnPatternDetected sets a callback for pattern detection events.
func (pd *PatternDetector) OnPatternDetected(fn func(*DetectedPattern)) {
	pd.mu.Lock()
	defer pd.mu.Unlock()
	pd.onPatternDetected = fn
}

// RecordEvent records an event for pattern detection.
func (pd *PatternDetector) RecordEvent(nodeID devtools.NodeID, fieldType string, fieldValue interface{}) {
	if !pd.enabled.Load() {
		return
	}

	event := PatternEvent{
		NodeID:    nodeID,
		FieldType: fieldType,
		FieldValue: fieldValue,
		Timestamp: time.Now(),
	}

	pd.eventHistory.Push(event)

	// Run pattern detectors
	pd.detectOscillations(nodeID)
	pd.detectSameField(nodeID, fieldType)
	pd.detectHighFrequency(nodeID)
}

// RecordFrameEvent records a frame-level event.
func (pd *PatternDetector) RecordFrameEvent(frameID devtools.FrameID, nodeID devtools.NodeID, mutationType string) {
	if !pd.enabled.Load() {
		return
	}

	event := PatternEvent{
		NodeID:    nodeID,
		FieldType: mutationType,
		Timestamp: time.Now(),
		FrameID:   frameID,
	}

	pd.eventHistory.Push(event)
}

// DetectPatterns runs all pattern detectors.
func (pd *PatternDetector) DetectPatterns() []*DetectedPattern {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	var allPatterns []*DetectedPattern

	for _, patterns := range pd.patterns {
		allPatterns = append(allPatterns, patterns...)
	}

	return allPatterns
}

// GetPatterns returns detected patterns for a node.
func (pd *PatternDetector) GetPatterns(nodeID devtools.NodeID) []*DetectedPattern {
	pd.mu.RLock()
	defer pd.mu.RUnlock()

	if patterns, exists := pd.patterns[nodeID]; exists {
		result := make([]*DetectedPattern, len(patterns))
		copy(result, patterns)
		return result
	}
	return nil
}

// GetAllPatterns returns all detected patterns.
func (pd *PatternDetector) GetAllPatterns() map[devtools.NodeID][]*DetectedPattern {
	pd.mu.RLock()
	defer pd.mu.RUnlock()

	result := make(map[devtools.NodeID][]*DetectedPattern)
	for nodeID, patterns := range pd.patterns {
		result[nodeID] = append(result[nodeID], patterns...)
	}
	return result
}

// addPattern adds a detected pattern.
func (pd *PatternDetector) addPattern(pattern *DetectedPattern) {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	pd.patterns[pattern.NodeID] = append(pd.patterns[pattern.NodeID], pattern)
	pd.patternCount.Add(1)

	if pd.onPatternDetected != nil {
		pd.onPatternDetected(pattern)
	}
}

// GetStats returns pattern detection statistics.
func (pd *PatternDetector) GetStats() PatternStats {
	pd.mu.RLock()
	defer pd.mu.RUnlock()

	stats := PatternStats{
		TotalPatterns:  pd.patternCount.Load(),
		ActivePatterns: len(pd.patterns),
	}

	for _, patterns := range pd.patterns {
		for _, p := range patterns {
			stats.AvgConfidence += p.Confidence
			switch p.Type {
			case PatternOscillation:
				stats.Oscillations++
			case PatternSameField:
				stats.SameFields++
			case PatternCascadeBurst:
				stats.CascadeBursts++
			case PatternLayoutRevert:
				stats.LayoutReverts++
			case PatternHighFrequency:
				stats.HighFrequency++
			case PatternBurst:
				stats.Bursts++
			}
		}
	}

	if stats.TotalPatterns > 0 {
		stats.AvgConfidence /= float64(stats.TotalPatterns)
	}

	return stats
}

// Reset clears all detected patterns.
func (pd *PatternDetector) Reset() {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	pd.patterns = make(map[devtools.NodeID][]*DetectedPattern)
	pd.patternCount.Store(0)
	pd.oscillationState = make(map[devtools.NodeID]*OscillationState)
	pd.sameFieldState = make(map[devtools.NodeID]*SameFieldState)
	pd.cascadeState = make(map[devtools.NodeID]*CascadeState)
}

// CleanupExpired removes expired patterns.
func (pd *PatternDetector) CleanupExpired(timeout time.Duration) {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	now := time.Now()
	for nodeID, patterns := range pd.patterns {
		active := make([]*DetectedPattern, 0)
		for _, p := range patterns {
			if now.Sub(p.EndTime) < timeout {
				active = append(active, p)
			}
		}
		if len(active) == 0 {
			delete(pd.patterns, nodeID)
		} else {
			pd.patterns[nodeID] = active
		}
	}
}

// OscillationState tracks state for oscillation detection.
type OscillationState struct {
	LastValue    interface{}
	Changes      []interface{}
	FirstChange  time.Time
}

// SameFieldState tracks state for same-field detection.
type SameFieldState struct {
	ChangeCount  int
	FirstChange  time.Time
	LastChange   time.Time
}

// CascadeState tracks state for cascade detection.
type CascadeState struct {
	AffectedNodes []devtools.NodeID
	StartTime     time.Time
}

// detectOscillations detects A→B→A→B oscillation patterns.
func (pd *PatternDetector) detectOscillations(nodeID devtools.NodeID) {
	events := pd.eventHistory.GetLastN(nodeID, pd.config.OscillationMinCycles*2)
	if len(events) < pd.config.OscillationMinCycles*2 {
		return
	}

	// Check for alternating values
	state, exists := pd.oscillationState[nodeID]
	if !exists {
		state = &OscillationState{
			Changes: make([]interface{}, 0, 10),
		}
		pd.oscillationState[nodeID] = state
	}

	// Track value changes
	for _, event := range events {
		if event.FieldValue == state.LastValue {
			continue
		}

		// Check if this value was seen before (potential oscillation)
		for _, prevValue := range state.Changes {
			if event.FieldValue == prevValue {
				// Found oscillation
				confidence := pd.calculateOscillationConfidence(state.Changes)
				pattern := &DetectedPattern{
					ID:         fmt.Sprintf("oscillation-%s-%d", nodeID, time.Now().UnixNano()),
					Type:       PatternOscillation,
					NodeID:     nodeID,
					Confidence: confidence,
					Severity:   pd.severityFromConfidence(confidence),
					StartTime:  state.FirstChange,
					EndTime:    event.Timestamp,
					Evidence: []PatternEvidence{
						{
							Timestamp:   event.Timestamp,
							Description: fmt.Sprintf("Value oscillated between %v and %v", prevValue, event.FieldValue),
						},
					},
				}
				pd.addPattern(pattern)
				return
			}
		}

		state.Changes = append(state.Changes, event.FieldValue)
		state.LastValue = event.FieldValue
		if len(state.Changes) == 1 {
			state.FirstChange = event.Timestamp
		}
	}
}

// detectSameField detects same-field-multiple-times pattern.
func (pd *PatternDetector) detectSameField(nodeID devtools.NodeID, fieldType string) {
	events := pd.eventHistory.GetLastN(nodeID, pd.config.SameFieldMinCount)
	if len(events) < pd.config.SameFieldMinCount {
		return
	}

	// Count consecutive same-field changes
	state, exists := pd.sameFieldState[nodeID]
	now := time.Now()

	if !exists {
		state = &SameFieldState{
			FirstChange: now,
		}
		pd.sameFieldState[nodeID] = state
	}

	// Reset if too much time has passed
	if !state.LastChange.IsZero() && now.Sub(state.LastChange) > pd.config.SameFieldMaxWindow {
		state.ChangeCount = 0
		state.FirstChange = now
	}

	// Count same-field events in window
	count := 0
	windowStart := now.Add(-pd.config.SameFieldMaxWindow)
	for _, event := range events {
		if event.FieldType == fieldType && event.Timestamp.After(windowStart) {
			count++
		}
	}

	if count >= pd.config.SameFieldMinCount {
		confidence := float64(count) / float64(pd.config.SameFieldMinCount)
		if confidence > 1.0 {
			confidence = 1.0
		}

		pattern := &DetectedPattern{
			ID:         fmt.Sprintf("samefield-%s-%s-%d", nodeID, fieldType, time.Now().UnixNano()),
			Type:       PatternSameField,
			NodeID:     nodeID,
			Confidence: confidence,
			Severity:   pd.severityFromConfidence(confidence),
			StartTime:  state.FirstChange,
			EndTime:    now,
			Duration:   now.Sub(state.FirstChange),
			Evidence: []PatternEvidence{
				{
					Timestamp:   now,
					Description: fmt.Sprintf("Field '%s' changed %d times in %v", fieldType, count, now.Sub(state.FirstChange)),
				},
			},
			Metadata: map[string]interface{}{
				"field":     fieldType,
				"count":     count,
				"window":    pd.config.SameFieldMaxWindow,
			},
		}
		pd.addPattern(pattern)

		// Reset state
		state.ChangeCount = 0
		state.FirstChange = time.Time{}
	}

	state.LastChange = now
}

// detectHighFrequency detects high-frequency update patterns.
func (pd *PatternDetector) detectHighFrequency(nodeID devtools.NodeID) {
	events := pd.eventHistory.GetLastN(nodeID, 100) // Get more events
	if len(events) < 10 {
		return
	}

	// Calculate update rate
	now := time.Now()
	windowStart := now.Add(-pd.config.HighFreqMinDuration)

	recentCount := 0
	var oldestRecent time.Time
	for _, event := range events {
		if event.Timestamp.After(windowStart) {
			recentCount++
			if oldestRecent.IsZero() || event.Timestamp.Before(oldestRecent) {
				oldestRecent = event.Timestamp
			}
		}
	}

	if recentCount == 0 {
		return
	}

	duration := now.Sub(oldestRecent)
	if duration == 0 {
		return
	}

	rate := float64(recentCount) / duration.Seconds()

	if rate > pd.config.HighFreqThreshold {
		confidence := rate / pd.config.HighFreqThreshold
		if confidence > 1.0 {
			confidence = 1.0
		}

		pattern := &DetectedPattern{
			ID:         fmt.Sprintf("highfreq-%s-%d", nodeID, time.Now().UnixNano()),
			Type:       PatternHighFrequency,
			NodeID:     nodeID,
			Confidence: confidence,
			Severity:   pd.severityFromConfidence(confidence),
			StartTime:  oldestRecent,
			EndTime:    now,
			Duration:   duration,
			Evidence: []PatternEvidence{
				{
					Timestamp:   now,
					Description: fmt.Sprintf("%.1f updates/sec (threshold: %.1f)", rate, pd.config.HighFreqThreshold),
					Data: map[string]interface{}{
						"rate":     rate,
						"count":    recentCount,
						"duration": duration,
					},
				},
			},
		}
		pd.addPattern(pattern)
	}
}

// calculateOscillationConfidence calculates confidence for oscillation pattern.
func (pd *PatternDetector) calculateOscillationConfidence(changes []interface{}) float64 {
	// More cycles = higher confidence
	cycles := len(changes) / 2
	if cycles >= pd.config.OscillationMinCycles {
		return 0.8 + float64(cycles-pd.config.OscillationMinCycles)*0.05
	}
	return 0.5
}

// severityFromConfidence converts confidence to severity.
func (pd *PatternDetector) severityFromConfidence(confidence float64) PatternSeverity {
	switch {
	case confidence >= 0.8:
		return PatternSeverityHigh
	case confidence >= 0.6:
		return PatternSeverityMedium
	case confidence >= 0.4:
		return PatternSeverityLow
	default:
		return PatternSeverityInfo
	}
}
