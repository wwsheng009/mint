// Package observation provides intelligent analysis for DevTools.
//
// This file implements the insights system for the observation layer.
package observation

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// ConfidenceLevel indicates the confidence level of an insight.
type ConfidenceLevel int

const (
	ConfidenceLow ConfidenceLevel = iota    // < 0.5
	ConfidenceMedium                        // 0.5-0.7
	ConfidenceHigh                          // 0.7-0.9
	ConfidenceVeryHigh                      // > 0.9
)

// String returns the string representation of the confidence level.
func (c ConfidenceLevel) String() string {
	switch c {
	case ConfidenceVeryHigh:
		return "Very High"
	case ConfidenceHigh:
		return "High"
	case ConfidenceMedium:
		return "Medium"
	default:
		return "Low"
	}
}

// Severity indicates the severity of an insight.
type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityCritical
)

// String returns the string representation of the severity.
func (s Severity) String() string {
	switch s {
	case SeverityCritical:
		return "Critical"
	case SeverityWarning:
		return "Warning"
	default:
		return "Info"
	}
}

// InsightType indicates the type of insight.
type InsightType int

const (
	InsightHotspot InsightType = iota
	InsightWaste
	InsightJitter
	InsightAnomaly
	InsightRegression
)

// String returns the string representation of the insight type.
func (i InsightType) String() string {
	switch i {
	case InsightHotspot:
		return "Hotspot"
	case InsightWaste:
		return "Waste"
	case InsightJitter:
		return "Jitter"
	case InsightAnomaly:
		return "Anomaly"
	case InsightRegression:
		return "Regression"
	default:
		return "Unknown"
	}
}

// Evidence represents a piece of evidence supporting an insight.
type Evidence struct {
	Source   string  `json:"source"`
	Metric   string  `json:"metric"`
	Value    float64 `json:"value"`
	Expected float64 `json:"expected"`
}

// Suggestion represents an actionable suggestion.
type Suggestion struct {
	Priority       int    `json:"priority"`
	Action         string `json:"action"`
	Reason         string `json:"reason"`
	ExpectedImpact string `json:"expected_impact"`
}

// Insight represents an analyzed observation with confidence and suggestions.
type Insight struct {
	ID          string          `json:"id"`
	Type        InsightType     `json:"type"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Confidence  ConfidenceLevel `json:"confidence"`
	Evidence    []Evidence      `json:"evidence"`
	Suggestions []Suggestion    `json:"suggestions"`
	Severity    Severity        `json:"severity"`
	ComponentID string          `json:"component_id,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

// InsightSummary provides a summary of all insights.
type InsightSummary struct {
	TotalInsights   int
	CriticalCount   int
	WarningCount    int
	InfoCount       int
	ByType          map[InsightType]int
	ByComponent     map[string]int
}

// ObservationLevel represents the observation level (0-3).
type ObservationLevel int

const (
	LevelDisabled ObservationLevel = iota
	LevelBasic     // HotspotDetector only
	LevelEnhanced  // HotspotDetector + WasteDetector
	LevelAdvanced  // All detectors + BehaviorProfiler
	LevelComplete  // All detectors + Profiler + BaselineComparator
)

// String returns the string representation of the observation level.
func (l ObservationLevel) String() string {
	switch l {
	case LevelBasic:
		return "Basic"
	case LevelEnhanced:
		return "Enhanced"
	case LevelAdvanced:
		return "Advanced"
	case LevelComplete:
		return "Complete"
	default:
		return "Disabled"
	}
}

// ObservationLayer coordinates all observation detectors.
type ObservationLayer struct {
	mu     sync.RWMutex
	enabled atomic.Bool
	level   atomic.Value // ObservationLevel

	// Detectors
	hotspot  *HotspotDetector
	waste    *WasteDetector
	jitter   *JitterDetector
	profiler *BehaviorProfiler
	baseline *BaselineComparator

	// Insights
	insights     []Insight
	insightsMu   sync.RWMutex
	maxInsights  int

	// Configuration
	config *LayerConfig
}

// LayerConfig configures the observation layer.
type LayerConfig struct {
	HotspotConfig  *HotspotConfig
	WasteConfig    *WasteConfig
	JitterConfig   *JitterConfig
	ProfilerConfig *ProfilerConfig
	BaselineConfig *BaselineConfig
	MaxInsights    int
}

// DefaultLayerConfig returns the default configuration.
func DefaultLayerConfig() *LayerConfig {
	return &LayerConfig{
		HotspotConfig:  DefaultHotspotConfig(),
		WasteConfig:    DefaultWasteConfig(),
		JitterConfig:   DefaultJitterConfig(),
		ProfilerConfig: DefaultProfilerConfig(),
		BaselineConfig: DefaultBaselineConfig(),
		MaxInsights:    100,
	}
}

// NewObservationLayer creates a new observation layer.
func NewObservationLayer(cfg *LayerConfig) *ObservationLayer {
	if cfg == nil {
		cfg = DefaultLayerConfig()
	}

	layer := &ObservationLayer{
		hotspot:     NewHotspotDetector(cfg.HotspotConfig),
		waste:       NewWasteDetector(cfg.WasteConfig),
		jitter:      NewJitterDetector(cfg.JitterConfig),
		profiler:    NewBehaviorProfiler(cfg.ProfilerConfig),
		baseline:    NewBaselineComparator(cfg.BaselineConfig),
		insights:    make([]Insight, 0, cfg.MaxInsights),
		maxInsights: cfg.MaxInsights,
		config:      cfg,
	}

	layer.level.Store(LevelDisabled)

	// Set up callbacks
	layer.setupCallbacks()

	return layer
}

// setupCallbacks sets up detector callbacks for insight generation.
func (ol *ObservationLayer) setupCallbacks() {
	ol.hotspot.OnHotspotDetected(func(hotspot *ComponentHotspot) {
		ol.generateHotspotInsight(hotspot)
	})

	ol.waste.OnWasteDetected(func(report *WasteReport) {
		ol.generateWasteInsight(report)
	})

	ol.jitter.OnJitterDetected(func(report *JitterReport) {
		ol.generateJitterInsight(report)
	})

	ol.profiler.OnAnomalyDetected(func(nodeID devtools.NodeID, duration time.Duration, baseline *PerformanceBaseline) {
		ol.generateAnomalyInsight(nodeID, duration, baseline)
	})
}

// Enable enables the observation layer at the specified level.
func (ol *ObservationLayer) Enable(level ObservationLevel) {
	ol.enabled.Store(true)
	ol.level.Store(level)

	// Enable detectors based on level
	switch level {
	case LevelBasic:
		ol.hotspot.Enable()
	case LevelEnhanced:
		ol.hotspot.Enable()
		ol.waste.Enable()
	case LevelAdvanced:
		ol.hotspot.Enable()
		ol.waste.Enable()
		ol.jitter.Enable()
		ol.profiler.Enable()
	case LevelComplete:
		ol.hotspot.Enable()
		ol.waste.Enable()
		ol.jitter.Enable()
		ol.profiler.Enable()
		ol.baseline.Enable()
	}
}

// Disable disables the observation layer.
func (ol *ObservationLayer) Disable() {
	ol.enabled.Store(false)
	ol.level.Store(LevelDisabled)

	ol.hotspot.Disable()
	ol.waste.Disable()
	ol.jitter.Disable()
	ol.profiler.Disable()
	ol.baseline.Disable()
}

// IsEnabled returns whether the layer is enabled.
func (ol *ObservationLayer) IsEnabled() bool {
	return ol.enabled.Load()
}

// GetLevel returns the current observation level.
func (ol *ObservationLayer) GetLevel() ObservationLevel {
	return ol.level.Load().(ObservationLevel)
}

// ProcessFrame processes a frame through all enabled detectors.
func (ol *ObservationLayer) ProcessFrame(entry interface{}) {
	if !ol.enabled.Load() {
		return
	}

	// Type assertion for different frame types
	// This is a simplified version - real implementation would handle proper types
}

// GetHotspots returns current hotspots.
func (ol *ObservationLayer) GetHotspots() []*ComponentHotspot {
	return ol.hotspot.GetHotspots()
}

// GetWasteReports returns current waste reports.
func (ol *ObservationLayer) GetWasteReports() []*WasteReport {
	return ol.waste.GetWasteReports()
}

// GetJitterReport returns current jitter report.
func (ol *ObservationLayer) GetJitterReport() *JitterReport {
	return ol.jitter.GetReport()
}

// GetProfiles returns all behavior profiles.
func (ol *ObservationLayer) GetProfiles() []*BehaviorProfile {
	return ol.profiler.GetProfiles()
}

// GetInsights returns all generated insights.
func (ol *ObservationLayer) GetInsights() []Insight {
	ol.insightsMu.RLock()
	defer ol.insightsMu.RUnlock()

	result := make([]Insight, len(ol.insights))
	copy(result, ol.insights)
	return result
}

// GetInsightSummary returns a summary of all insights.
func (ol *ObservationLayer) GetInsightSummary() *InsightSummary {
	ol.insightsMu.RLock()
	defer ol.insightsMu.RUnlock()

	summary := &InsightSummary{
		TotalInsights: len(ol.insights),
		ByType:        make(map[InsightType]int),
		ByComponent:   make(map[string]int),
	}

	for _, insight := range ol.insights {
		summary.ByType[insight.Type]++

		switch insight.Severity {
		case SeverityCritical:
			summary.CriticalCount++
		case SeverityWarning:
			summary.WarningCount++
		default:
			summary.InfoCount++
		}

		if insight.ComponentID != "" {
			summary.ByComponent[insight.ComponentID]++
		}
	}

	return summary
}

// addInsight adds an insight to the collection.
func (ol *ObservationLayer) addInsight(insight Insight) {
	ol.insightsMu.Lock()
	defer ol.insightsMu.Unlock()

	// Check for duplicates and update instead
	for i, existing := range ol.insights {
		if existing.Type == insight.Type && existing.ComponentID == insight.ComponentID {
			// Update existing insight
			ol.insights[i] = insight
			return
		}
	}

	// Add new insight
	ol.insights = append(ol.insights, insight)

	// Prune old insights if too many
	if len(ol.insights) > ol.maxInsights {
		ol.insights = ol.insights[len(ol.insights)-ol.maxInsights:]
	}
}

// generateHotspotInsight generates an insight from a hotspot detection.
func (ol *ObservationLayer) generateHotspotInsight(hotspot *ComponentHotspot) {
	severity := SeverityInfo
	if hotspot.Severity == HotspotSeverityCritical {
		severity = SeverityCritical
	} else if hotspot.Severity == HotspotSeverityWarning {
		severity = SeverityWarning
	}

	// Calculate confidence based on sample count
	confidence := ConfidenceLow
	if hotspot.FrameCount > 100 {
		confidence = ConfidenceVeryHigh
	} else if hotspot.FrameCount > 50 {
		confidence = ConfidenceHigh
	} else if hotspot.FrameCount > 20 {
		confidence = ConfidenceMedium
	}

	suggestions := []Suggestion{
		{
			Priority:       1,
			Action:         "Review component rendering logic",
			Reason:         "Component is consistently slow",
			ExpectedImpact: "Improved frame rate",
		},
	}

	if hotspot.Severity == HotspotSeverityCritical {
		suggestions = append(suggestions, Suggestion{
			Priority:       0,
			Action:         "Consider virtualization or lazy loading",
			Reason:         "Component is critically slow",
			ExpectedImpact: "Significant performance improvement",
		})
	}

	insight := Insight{
		ID:          fmt.Sprintf("hotspot-%s", hotspot.NodeID),
		Type:        InsightHotspot,
		Title:       fmt.Sprintf("Performance hotspot detected in component %s", hotspot.NodeID),
		Description: fmt.Sprintf("Component averaging %v per render (max: %v)", hotspot.AvgTime, hotspot.MaxTime),
		Confidence:  confidence,
		Evidence: []Evidence{
			{Source: "HotspotDetector", Metric: "AvgTime", Value: float64(hotspot.AvgTime.Nanoseconds()), Expected: float64(5000000)},
			{Source: "HotspotDetector", Metric: "FrameCount", Value: float64(hotspot.FrameCount), Expected: 0},
		},
		Suggestions: suggestions,
		Severity:    severity,
		ComponentID: string(hotspot.NodeID),
		CreatedAt:   time.Now(),
	}

	ol.addInsight(insight)
}

// generateWasteInsight generates an insight from a waste detection.
func (ol *ObservationLayer) generateWasteInsight(report *WasteReport) {
	severity := SeverityInfo
	if report.Severity == WasteHigh {
		severity = SeverityCritical
	} else if report.Severity == WasteMedium {
		severity = SeverityWarning
	}

	confidence := ConfidenceMedium
	if report.TotalRenders > 50 {
		confidence = ConfidenceHigh
	}

	insight := Insight{
		ID:          fmt.Sprintf("waste-%s", report.NodeID),
		Type:        InsightWaste,
		Title:       fmt.Sprintf("Wasted renders detected in component %s", report.NodeID),
		Description: fmt.Sprintf("%.1f%% of renders are wasted (no actual changes)", report.WasteRate),
		Confidence:  confidence,
		Evidence: []Evidence{
			{Source: "WasteDetector", Metric: "WasteRate", Value: report.WasteRate, Expected: 10},
			{Source: "WasteDetector", Metric: "TotalRenders", Value: float64(report.TotalRenders), Expected: 0},
		},
		Suggestions: []Suggestion{
			{
				Priority:       1,
				Action:         "Implement shouldComponentUpdate or equivalent",
				Reason:         "Component is rendering without changes",
				ExpectedImpact: "Reduced CPU usage",
			},
			{
				Priority:       2,
				Action:         "Memoize component props",
				Reason:         "Prevent unnecessary re-renders",
				ExpectedImpact: "Improved render efficiency",
			},
		},
		Severity:    severity,
		ComponentID: string(report.NodeID),
		CreatedAt:   time.Now(),
	}

	ol.addInsight(insight)
}

// generateJitterInsight generates an insight from a jitter detection.
func (ol *ObservationLayer) generateJitterInsight(report *JitterReport) {
	severity := SeverityInfo
	if report.Severity == JitterHigh {
		severity = SeverityWarning
	}

	confidence := ConfidenceMedium
	if report.SampleCount > 100 {
		confidence = ConfidenceHigh
	}

	insight := Insight{
		ID:          "jitter-frame-time",
		Type:        InsightJitter,
		Title:       "Frame time jitter detected",
		Description: fmt.Sprintf("Frame time variance (CV: %.2f) indicates unstable performance", report.CurrentJitter),
		Confidence:  confidence,
		Evidence: []Evidence{
			{Source: "JitterDetector", Metric: "CV", Value: report.CurrentJitter, Expected: 0.2},
			{Source: "JitterDetector", Metric: "MeanFrameTime", Value: report.MeanFrameTime, Expected: 16.67},
		},
		Suggestions: []Suggestion{
			{
				Priority:       1,
				Action:         "Investigate inconsistent render times",
				Reason:         "High jitter affects perceived smoothness",
				ExpectedImpact: "Smoother animations",
			},
		},
		Severity:  severity,
		CreatedAt: time.Now(),
	}

	ol.addInsight(insight)
}

// generateAnomalyInsight generates an insight from an anomaly detection.
func (ol *ObservationLayer) generateAnomalyInsight(nodeID devtools.NodeID, duration time.Duration, baseline *PerformanceBaseline) {
	severity := SeverityWarning
	if duration > baseline.P99 {
		severity = SeverityCritical
	}

	insight := Insight{
		ID:          fmt.Sprintf("anomaly-%s", nodeID),
		Type:        InsightAnomaly,
		Title:       fmt.Sprintf("Performance anomaly detected in component %s", nodeID),
		Description: fmt.Sprintf("Render time (%v) significantly exceeds baseline (%v)", duration, baseline.P95),
		Confidence:  ConfidenceHigh,
		Evidence: []Evidence{
			{Source: "BehaviorProfiler", Metric: "Duration", Value: float64(duration.Nanoseconds()), Expected: float64(baseline.P95.Nanoseconds())},
		},
		Suggestions: []Suggestion{
			{
				Priority:       1,
				Action:         "Check for one-time expensive operations",
				Reason:         "Anomalous render time detected",
				ExpectedImpact: "Stable performance",
			},
		},
		Severity:    severity,
		ComponentID: string(nodeID),
		CreatedAt:   time.Now(),
	}

	ol.addInsight(insight)
}

// Reset clears all observation data.
func (ol *ObservationLayer) Reset() {
	ol.hotspot.Reset()
	ol.waste.Reset()
	ol.jitter.Reset()
	ol.profiler.Reset()
	ol.baseline.Reset()

	ol.insightsMu.Lock()
	ol.insights = make([]Insight, 0, ol.maxInsights)
	ol.insightsMu.Unlock()
}

// GetStats returns combined statistics from all detectors.
func (ol *ObservationLayer) GetStats() *LayerStats {
	return &LayerStats{
		HotspotStats:  ol.hotspot.GetStats(),
		WasteStats:    ol.waste.GetStats(),
		JitterStats:   ol.jitter.GetStats(),
		ProfilerStats: ol.profiler.GetStats(),
		InsightSummary: ol.GetInsightSummary(),
		Level:         ol.GetLevel(),
		Enabled:       ol.IsEnabled(),
	}
}

// LayerStats contains statistics from the entire observation layer.
type LayerStats struct {
	HotspotStats  HotspotStats
	WasteStats    WasteStats
	JitterStats   JitterStats
	ProfilerStats ProfilerStats
	InsightSummary *InsightSummary
	Level         ObservationLevel
	Enabled       bool
}
