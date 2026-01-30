// Package observation provides the main observation layer for DevTools.
//
// This file coordinates V1 (pure statistics) and V2 (pattern detection + insights).
package observation

import (
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
	"github.com/wwsheng009/mint/devtools/observation/v1"
	"github.com/wwsheng009/mint/devtools/observation/v2"
)

// Layer is the main observation layer coordinating V1 and V2.
type Layer struct {
	mu    sync.RWMutex
	level *v1.LevelController

	// V1: Pure statistics (no judgments)
	metrics  *v1.MetricsCollector
	stats    *v1.StatsAnalyzer
	series   *v1.TimeSeriesStore

	// V2: Pattern detection + insights (with judgments)
	patterns  *v2.PatternDetector
	confidence *v2.ConfidenceModel
	insights  *v2.InsightsGenerator
}

// Config configures the observation layer.
type Config struct {
	InitialLevel      v1.Level
	TimeSeriesWindow  int
	SeriesTTL         time.Duration
	PatternTTL        time.Duration
	InsightTTL        time.Duration
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		InitialLevel:     v1.LevelBasic,
		TimeSeriesWindow: 30,
		SeriesTTL:        10 * time.Minute,
		PatternTTL:       5 * time.Minute,
		InsightTTL:       5 * time.Minute,
	}
}

// NewLayer creates a new observation layer.
func NewLayer(cfg *Config) *Layer {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	level := v1.NewLevelController(cfg.InitialLevel)

	return &Layer{
		level:     level,
		metrics:   v1.NewMetricsCollector(level),
		stats:     v1.NewStatsAnalyzer(level, nil), // Will be linked after creation
		series:    v1.NewTimeSeriesStore(level, cfg.TimeSeriesWindow, cfg.SeriesTTL),
		patterns:   v2.NewPatternDetector(nil),
		confidence: v2.NewConfidenceModel(),
		insights:   v2.NewInsightsGenerator(nil, nil),
	}
}

// LinkComponents links components after creation.
func (ol *Layer) LinkComponents() {
	ol.mu.Lock()
	defer ol.mu.Unlock()

	// Link stats analyzer with metrics collector
	ol.stats = v1.NewStatsAnalyzer(ol.level, ol.metrics)

	// Link insights generator
	ol.insights = v2.NewInsightsGenerator(ol.confidence, ol.patterns)
}

// SetLevel sets the observation level.
func (ol *Layer) SetLevel(level v1.Level) {
	ol.level.SetLevel(level)
}

// GetLevel returns the current observation level.
func (ol *Layer) GetLevel() v1.Level {
	return ol.level.GetLevel()
}

// Enable enables the observation layer.
func (ol *Layer) Enable(level v1.Level) {
	ol.mu.Lock()
	defer ol.mu.Unlock()

	ol.level.SetLevel(level)

	// Enable V2 features based on level
	if level >= v1.LevelEnhanced {
		ol.patterns.Enable()
		ol.insights.Enable()
	}
}

// Disable disables the observation layer.
func (ol *Layer) Disable() {
	ol.mu.Lock()
	defer ol.mu.Unlock()

	ol.level.SetLevel(v1.LevelNone)
	ol.patterns.Disable()
	ol.insights.Disable()
}

// ProcessFrame processes a frame through V1 statistics.
func (ol *Layer) ProcessFrame(entry *devtools.FrameEntry) {
	if !ol.level.IsEnabled() {
		return
	}

	// V1: Record metrics
	ol.metrics.RecordFrame(entry)

	// V2: Record events for pattern detection
	if ol.level.ShouldCollectEnhancedStats() {
		for i := 0; i < entry.EventCount; i++ {
			ol.patterns.RecordFrameEvent(entry.FrameID, devtools.NodeID(i), "event")
		}
	}
}

// RecordMutation records a mutation event.
func (ol *Layer) RecordMutation(nodeID devtools.NodeID, fieldType string, fieldValue interface{}) {
	if !ol.level.IsEnabled() {
		return
	}

	// V1: Count
	ol.metrics.RecordMutation(nodeID)

	// V2: Pattern detection
	if ol.level.ShouldCollectEnhancedStats() {
		ol.patterns.RecordEvent(nodeID, fieldType, fieldValue)
		ol.series.AddPoint(nodeID, float64(time.Now().UnixNano()))
	}
}

// RecordLayout records a layout operation.
func (ol *Layer) RecordLayout(nodeID devtools.NodeID) {
	if !ol.level.IsEnabled() {
		return
	}

	ol.metrics.RecordLayout(nodeID)
}

// RecordRepaint records a repaint operation.
func (ol *Layer) RecordRepaint(nodeID devtools.NodeID) {
	if !ol.level.IsEnabled() {
		return
	}

	ol.metrics.RecordRepaint(nodeID)
}

// GetMetrics returns the current metrics snapshot (V1).
func (ol *Layer) GetMetrics() *v1.MetricsSnapshot {
	return ol.metrics.GetSnapshot()
}

// GetComponentMetrics returns metrics for a specific component (V1).
func (ol *Layer) GetComponentMetrics(nodeID devtools.NodeID) *v1.ComponentMetrics {
	return ol.metrics.GetComponentMetrics(nodeID)
}

// GetTopN returns the top N components by metric type (V1).
func (ol *Layer) GetTopN(metric v1.MetricType, n int) []*v1.ComponentRank {
	return ol.stats.GetTopN(metric, n)
}

// GetDistribution returns distribution statistics for a metric (V1).
func (ol *Layer) GetDistribution(metric v1.MetricType) *v1.Distribution {
	return ol.stats.GetDistribution(metric)
}

// GetTimeSeries returns time series data for a component (V1).
func (ol *Layer) GetTimeSeries(nodeID devtools.NodeID) []v1.DataPoint {
	return ol.series.GetTimeSeries(nodeID)
}

// GetPatterns returns detected patterns for a component (V2).
func (ol *Layer) GetPatterns(nodeID devtools.NodeID) []*v2.DetectedPattern {
	return ol.patterns.GetPatterns(nodeID)
}

// GetAllPatterns returns all detected patterns (V2).
func (ol *Layer) GetAllPatterns() map[devtools.NodeID][]*v2.DetectedPattern {
	return ol.patterns.GetAllPatterns()
}

// GetInsights returns all generated insights (V2).
func (ol *Layer) GetInsights() []*v2.Insight {
	return ol.insights.GetAllInsights()
}

// GetHighConfidenceInsights returns insights above a confidence threshold.
func (ol *Layer) GetHighConfidenceInsights(threshold float64) []*v2.Insight {
	return ol.insights.GetHighConfidenceInsights(threshold)
}

// GeneratePatternInsights generates insights from detected patterns (V2).
func (ol *Layer) GeneratePatternInsights() []*v2.Insight {
	return ol.insights.GeneratePatternInsights()
}

// CalculateSignalConfidence calculates confidence for a signal (V2).
func (ol *Layer) CalculateSignalConfidence(signal *v2.Signal) float64 {
	return ol.confidence.CalculateSignalConfidence(signal)
}

// GetStats returns combined statistics from V1 and V2.
type LayerStats struct {
	// V1 Stats
	Metrics    *v1.MetricsSnapshot
	Level      v1.Level
	Overhead   float64

	// V2 Stats
	PatternStats   v2.PatternStats
	InsightStats   v2.InsightStats
}

// GetStats returns statistics from the observation layer.
func (ol *Layer) GetStats() *LayerStats {
	ol.mu.RLock()
	defer ol.mu.RUnlock()

	return &LayerStats{
		Metrics:      ol.metrics.GetSnapshot(),
		Level:        ol.level.GetLevel(),
		Overhead:     ol.level.ExpectedOverhead(),
		PatternStats: ol.patterns.GetStats(),
		InsightStats: ol.insights.GetStats(),
	}
}

// CleanupExpired removes expired patterns and insights.
func (ol *Layer) CleanupExpired() {
	ol.patterns.CleanupExpired()
	ol.insights.CleanupExpired()
	ol.series.CleanupExpired()
}

// Reset clears all observation data.
func (ol *Layer) Reset() {
	ol.mu.Lock()
	defer ol.mu.Unlock()

	ol.metrics.Reset()
	ol.patterns.Reset()
	ol.confidence.Reset()
	ol.insights.Reset()
	ol.series.Clear()
}
