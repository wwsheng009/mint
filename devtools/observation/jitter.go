// Package observation provides intelligent analysis for DevTools.
//
// This file implements the JitterDetector for detecting frame time variance.
package observation

import (
	"math"
	"sync"
	"sync/atomic"

	"github.com/wwsheng009/mint/devtools"
)

// JitterSeverity indicates the severity level of jitter.
type JitterSeverity int

const (
	JitterNone JitterSeverity = iota
	JitterLow     // CV < 0.2
	JitterMedium  // CV 0.2-0.5
	JitterHigh    // CV > 0.5
)

// String returns the string representation of the severity.
func (j JitterSeverity) String() string {
	switch j {
	case JitterLow:
		return "Low"
	case JitterMedium:
		return "Medium"
	case JitterHigh:
		return "High"
	default:
		return "None"
	}
}

// JitterReport contains jitter analysis results.
type JitterReport struct {
	CurrentJitter  float64              // Current jitter value (CV)
	MeanFrameTime  float64              // Mean frame time in ms
	StdDev         float64              // Standard deviation in ms
	MinFrameTime   float64              // Minimum frame time in ms
	MaxFrameTime   float64              // Maximum frame time in ms
	JitterFrames   []devtools.FrameID   // Frames with high jitter
	AnomalyFrames  []devtools.FrameID   // Statistical anomalies
	Severity       JitterSeverity
	SampleCount    int
	LastUpdated    int64 // Unix nanoseconds
}

// JitterDetector detects frame time variance and instability.
type JitterDetector struct {
	mu     sync.RWMutex
	enabled atomic.Bool

	// Frame time samples (duration in nanoseconds)
	frameTimes []float64
	windowSize int

	// Statistics (running calculation)
	count        int
	mean         float64
	M2           float64  // Sum of squared differences from mean (for Welford's algorithm)
	minTime      float64
	maxTime      float64

	// Configuration
	jitterThreshold   float64 // Jitter threshold (CV multiplier)
	anomalySigma      float64 // Anomaly detection (standard deviations)
	minSamples        int     // Minimum samples before reporting

	// Callback
	onJitterDetected  func(*JitterReport)
}

// JitterConfig configures the JitterDetector.
type JitterConfig struct {
	WindowSize      int     // Number of frames to analyze
	JitterThreshold float64 // CV threshold for warning
	AnomalySigma    float64 // Sigma for anomaly detection
	MinSamples      int     // Minimum samples required
}

// DefaultJitterConfig returns the default configuration.
func DefaultJitterConfig() *JitterConfig {
	return &JitterConfig{
		WindowSize:      60,              // 1 second at 60fps
		JitterThreshold: 0.3,             // CV of 0.3 for warning
		AnomalySigma:    2.0,             // 2 sigma
		MinSamples:      10,              // Need at least 10 samples
	}
}

// NewJitterDetector creates a new jitter detector.
func NewJitterDetector(cfg *JitterConfig) *JitterDetector {
	if cfg == nil {
		cfg = DefaultJitterConfig()
	}

	return &JitterDetector{
		frameTimes:      make([]float64, cfg.WindowSize),
		windowSize:      cfg.WindowSize,
		jitterThreshold: cfg.JitterThreshold,
		anomalySigma:    cfg.AnomalySigma,
		minSamples:      cfg.MinSamples,
		minTime:         math.MaxFloat64,
		maxTime:         0,
	}
}

// Enable enables the jitter detector.
func (jd *JitterDetector) Enable() {
	jd.enabled.Store(true)
}

// Disable disables the jitter detector.
func (jd *JitterDetector) Disable() {
	jd.enabled.Store(false)
}

// IsEnabled returns whether the detector is enabled.
func (jd *JitterDetector) IsEnabled() bool {
	return jd.enabled.Load()
}

// OnJitterDetected sets a callback for jitter detection events.
func (jd *JitterDetector) OnJitterDetected(fn func(*JitterReport)) {
	jd.mu.Lock()
	defer jd.mu.Unlock()
	jd.onJitterDetected = fn
}

// ProcessFrame processes a frame and updates jitter statistics.
func (jd *JitterDetector) ProcessFrame(entry *devtools.FrameEntry) {
	if !jd.enabled.Load() {
		return
	}

	duration := float64(entry.Duration.Nanoseconds())
	frameID := entry.FrameID

	jd.mu.Lock()
	defer jd.mu.Unlock()

	// Update min/max
	if duration < jd.minTime {
		jd.minTime = duration
	}
	if duration > jd.maxTime {
		jd.maxTime = duration
	}

	// Add to circular buffer
	pos := jd.count % jd.windowSize
	jd.frameTimes[pos] = duration
	jd.count++

	// Update running statistics using Welford's algorithm
	jd.count++
	delta := duration - jd.mean
	jd.mean += delta / float64(jd.count)
	delta2 := duration - jd.mean
	jd.M2 += delta * delta2

	// Only report if we have enough samples
	if jd.count < jd.minSamples {
		return
	}

	// Generate report
	report := jd.generateReport(frameID)

	// Trigger callback if significant jitter
	if report.Severity != JitterNone && jd.onJitterDetected != nil {
		jd.onJitterDetected(report)
	}
}

// generateReport generates a jitter report from current statistics.
func (jd *JitterDetector) generateReport(currentFrame devtools.FrameID) *JitterReport {
	// Calculate variance and standard deviation
	variance := jd.M2 / float64(jd.count)
	stdDev := math.Sqrt(variance)

	// Calculate coefficient of variation (CV)
	var cv float64
	if jd.mean > 0 {
		cv = stdDev / jd.mean
	}

	// Determine severity
	var severity JitterSeverity
	switch {
	case cv < 0.2:
		severity = JitterNone
	case cv < 0.5:
		severity = JitterMedium
	default:
		severity = JitterHigh
	}

	// Find jitter frames (frames that deviate significantly from mean)
	jitterFrames := jd.findJitterFrames(jd.mean, stdDev)

	// Find anomaly frames (statistical outliers)
	anomalyFrames := jd.findAnomalyFrames(jd.mean, stdDev)

	return &JitterReport{
		CurrentJitter: cv,
		MeanFrameTime: jd.mean / 1e6, // Convert to ms
		StdDev:        stdDev / 1e6,
		MinFrameTime:  jd.minTime / 1e6,
		MaxFrameTime:  jd.maxTime / 1e6,
		JitterFrames:  jitterFrames,
		AnomalyFrames: anomalyFrames,
		Severity:      severity,
		SampleCount:   jd.count,
	}
}

// findJitterFrames finds frames with high jitter.
func (jd *JitterDetector) findJitterFrames(mean, stdDev float64) []devtools.FrameID {
	// Frames beyond 1 standard deviation
	var frames []devtools.FrameID
	threshold := mean + stdDev

	sampleCount := jd.count
	if sampleCount > jd.windowSize {
		sampleCount = jd.windowSize
	}

	for i := 0; i < sampleCount; i++ {
		if jd.frameTimes[i] > threshold {
			// Frame IDs would need to be tracked separately
			// For now, return empty slice
		}
	}

	return frames
}

// findAnomalyFrames finds statistical anomalies.
func (jd *JitterDetector) findAnomalyFrames(mean, stdDev float64) []devtools.FrameID {
	// Frames beyond anomalySigma standard deviations
	var frames []devtools.FrameID
	threshold := mean + (jd.anomalySigma * stdDev)

	sampleCount := jd.count
	if sampleCount > jd.windowSize {
		sampleCount = jd.windowSize
	}

	for i := 0; i < sampleCount; i++ {
		if jd.frameTimes[i] > threshold {
			// Frame IDs would need to be tracked separately
		}
	}

	return frames
}

// GetReport returns the current jitter report.
func (jd *JitterDetector) GetReport() *JitterReport {
	jd.mu.Lock()
	defer jd.mu.Unlock()

	if jd.count < jd.minSamples {
		return &JitterReport{
			Severity:    JitterNone,
			SampleCount: jd.count,
		}
	}

	return jd.generateReport(0)
}

// GetStats returns detector statistics.
func (jd *JitterDetector) GetStats() JitterStats {
	jd.mu.RLock()
	defer jd.mu.RUnlock()

	var variance float64
	if jd.count > 0 {
		variance = jd.M2 / float64(jd.count)
	}
	stdDev := math.Sqrt(variance)

	var cv float64
	if jd.mean > 0 {
		cv = stdDev / jd.mean
	}

	return JitterStats{
		SampleCount:    jd.count,
		MeanDurationMs: jd.mean / 1e6,
		StdDevMs:       stdDev / 1e6,
		CV:             cv,
		MinDurationMs:  jd.minTime / 1e6,
		MaxDurationMs:  jd.maxTime / 1e6,
	}
}

// Reset clears all jitter data.
func (jd *JitterDetector) Reset() {
	jd.mu.Lock()
	defer jd.mu.Unlock()

	jd.frameTimes = make([]float64, jd.windowSize)
	jd.count = 0
	jd.mean = 0
	jd.M2 = 0
	jd.minTime = math.MaxFloat64
	jd.maxTime = 0
}

// JitterStats contains statistics about jitter detection.
type JitterStats struct {
	SampleCount    int
	MeanDurationMs float64
	StdDevMs       float64
	CV             float64 // Coefficient of Variation
	MinDurationMs  float64
	MaxDurationMs  float64
}

// GetSeverity returns the severity based on CV.
func (s JitterStats) GetSeverity() JitterSeverity {
	switch {
	case s.CV < 0.2:
		return JitterNone
	case s.CV < 0.5:
		return JitterMedium
	default:
		return JitterHigh
	}
}
