// Package v2 provides pattern detection and analysis for DevTools.
//
// This file implements pattern type definitions for the pattern detector.
package v2

import (
	"fmt"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// PatternType represents the type of detected pattern.
type PatternType int

const (
	PatternOscillation PatternType = iota // A→B→A→B oscillation
	PatternSameField                       // Same field changed multiple times
	PatternCascadeBurst                    // Cascading burst updates
	PatternLayoutRevert                    // Layout immediately reverted
	PatternHighFrequency                   // High frequency updates
	PatternBurst                           // Burst of updates
)

// String returns the string representation of the pattern type.
func (pt PatternType) String() string {
	switch pt {
	case PatternOscillation:
		return "Oscillation"
	case PatternSameField:
		return "SameField"
	case PatternCascadeBurst:
		return "CascadeBurst"
	case PatternLayoutRevert:
		return "LayoutRevert"
	case PatternHighFrequency:
		return "HighFrequency"
	case PatternBurst:
		return "Burst"
	default:
		return "Unknown"
	}
}

// PatternSeverity represents the severity of a pattern.
type PatternSeverity int

const (
	PatternSeverityInfo PatternSeverity = iota
	PatternSeverityLow
	PatternSeverityMedium
	PatternSeverityHigh
)

// String returns the string representation of the severity.
func (ps PatternSeverity) String() string {
	switch ps {
	case PatternSeverityHigh:
		return "High"
	case PatternSeverityMedium:
		return "Medium"
	case PatternSeverityLow:
		return "Low"
	default:
		return "Info"
	}
}

// PatternEvidence represents evidence supporting a pattern detection.
type PatternEvidence struct {
	Timestamp   time.Time
	Description string
	Data        map[string]interface{}
}

// DetectedPattern represents a detected behavioral pattern.
type DetectedPattern struct {
	ID         string
	Type       PatternType
	NodeID     devtools.NodeID
	Confidence float64 // 0.0 - 1.0
	Severity   PatternSeverity
	StartTime  time.Time
	EndTime    time.Time
	Duration   time.Duration
	Evidence   []PatternEvidence
	Metadata   map[string]interface{}
}

// String returns a string representation of the pattern.
func (dp *DetectedPattern) String() string {
	return fmt.Sprintf("[%s] %s on %s (confidence: %.2f)",
		dp.Severity, dp.Type, dp.NodeID, dp.Confidence)
}

// IsActive returns true if the pattern is still active (recent).
func (dp *DetectedPattern) IsActive(timeout time.Duration) bool {
	return time.Since(dp.EndTime) < timeout
}

// PatternEvent represents a single event that could be part of a pattern.
type PatternEvent struct {
	NodeID     devtools.NodeID
	FieldType  string // For same-field detection
	FieldValue interface{}
	Timestamp  time.Time
	FrameID    devtools.FrameID
}

// PatternDetectorConfig configures the pattern detector.
type PatternDetectorConfig struct {
	// Oscillation detection
	OscillationMinCycles int
	OscillationMaxWindow time.Duration

	// Same field detection
	SameFieldMinCount   int
	SameFieldMaxWindow  time.Duration

	// Cascade burst detection
	CascadeMinNodes     int
	CascadeMaxDelay     time.Duration

	// Layout revert detection
	LayoutRevertMaxDelay time.Duration

	// High frequency detection
	HighFreqThreshold    float64 // updates per second
	HighFreqMinDuration  time.Duration

	// Burst detection
	BurstMultiplier      float64 // vs average rate
	BurstMinCount        int
	BurstWindow          time.Duration
}

// DefaultPatternDetectorConfig returns the default configuration.
func DefaultPatternDetectorConfig() *PatternDetectorConfig {
	return &PatternDetectorConfig{
		OscillationMinCycles:  3,
		OscillationMaxWindow:  500 * time.Millisecond,

		SameFieldMinCount:  5,
		SameFieldMaxWindow: 200 * time.Millisecond,

		CascadeMinNodes: 3,
		CascadeMaxDelay: 100 * time.Millisecond,

		LayoutRevertMaxDelay: 50 * time.Millisecond,

		HighFreqThreshold:   60.0, // > 60 updates/sec
		HighFreqMinDuration: 1 * time.Second,

		BurstMultiplier: 5.0, // 5x average rate
		BurstMinCount:   10,
		BurstWindow:     100 * time.Millisecond,
	}
}

// PatternStats contains statistics about pattern detection.
type PatternStats struct {
	TotalPatterns    uint64
	Oscillations     uint64
	SameFields       uint64
	CascadeBursts    uint64
	LayoutReverts    uint64
	HighFrequency    uint64
	Bursts           uint64
	ActivePatterns   int
	AvgConfidence    float64
}
