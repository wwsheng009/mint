// Package v2 provides pattern detection and analysis for DevTools.
//
// This file implements the insights system using the confidence model.
package v2

import (
	"fmt"
	"sync"
	"time"
)

// InsightType represents the type of insight.
type InsightType int

const (
	InsightPattern InsightType = iota
	InsightHotspot
	InsightWaste
	InsightJitter
	InsightAnomaly
	InsightRegression
)

// String returns the string representation.
func (it InsightType) String() string {
	switch it {
	case InsightPattern:
		return "Pattern"
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

// InsightSeverity represents the severity of an insight.
type InsightSeverity int

const (
	SeverityInfo InsightSeverity = iota
	SeverityWarning
	SeverityCritical
)

// String returns the string representation.
func (is InsightSeverity) String() string {
	switch is {
	case SeverityCritical:
		return "Critical"
	case SeverityWarning:
		return "Warning"
	default:
		return "Info"
	}
}

// Suggestion represents an actionable suggestion.
type Suggestion struct {
	Priority       int     // 0 = highest
	Action         string
	Reason         string
	ExpectedImpact string
	Confidence     float64 // 0.0-1.0
}

// Evidence represents a piece of evidence.
type Evidence struct {
	Source   string
	Metric   string
	Value    float64
	Expected float64
}

// Insight represents an analyzed observation with confidence.
type Insight struct {
	ID          string
	Type        InsightType
	Title       string
	Description string
	Confidence  float64       // 0.0-1.0
	Level       ConfidenceLevel
	Severity    InsightSeverity
	Evidence    []Evidence
	Suggestions []Suggestion
	ComponentID string
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

// IsExpired returns true if the insight has expired.
func (i *Insight) IsExpired() bool {
	return !i.ExpiresAt.IsZero() && time.Now().After(i.ExpiresAt)
}

// InsightsGenerator generates insights using the confidence model.
type InsightsGenerator struct {
	mu           sync.RWMutex
	enabled      bool
	confidence   *ConfidenceModel
	patterns     *PatternDetector
	ttl          time.Duration
	maxInsights   int

	insights     []*Insight
	insightIndex map[string]int // ID -> index
}

// NewInsightsGenerator creates a new insights generator.
func NewInsightsGenerator(confidence *ConfidenceModel, patterns *PatternDetector) *InsightsGenerator {
	return &InsightsGenerator{
		confidence:    confidence,
		patterns:      patterns,
		ttl:           5 * time.Minute,
		maxInsights:   100,
		insights:      make([]*Insight, 0, 100),
		insightIndex:  make(map[string]int),
	}
}

// Enable enables the insights generator.
func (ig *InsightsGenerator) Enable() {
	ig.mu.Lock()
	defer ig.mu.Unlock()
	ig.enabled = true
}

// Disable disables the insights generator.
func (ig *InsightsGenerator) Disable() {
	ig.mu.Lock()
	defer ig.mu.Unlock()
	ig.enabled = false
}

// SetTTL sets the time-to-live for insights.
func (ig *InsightsGenerator) SetTTL(ttl time.Duration) {
	ig.mu.Lock()
	defer ig.mu.Unlock()
	ig.ttl = ttl
}

// GeneratePatternInsights generates insights from detected patterns.
func (ig *InsightsGenerator) GeneratePatternInsights() []*Insight {
	ig.mu.Lock()
	defer ig.mu.Unlock()

	if !ig.enabled {
		return nil
	}

	allPatterns := ig.patterns.GetAllPatterns()
	insights := make([]*Insight, 0)

	for _, patterns := range allPatterns {
		for _, pattern := range patterns {
			// Only generate insights for high-confidence patterns
			if pattern.Confidence < 0.4 {
				continue
			}

			insight := ig.createPatternInsight(pattern)
			insights = append(insights, insight)
		}
	}

	return insights
}

// createPatternInsight creates an insight from a pattern.
func (ig *InsightsGenerator) createPatternInsight(pattern *DetectedPattern) *Insight {
	suggestions := ig.generateSuggestions(pattern)

	insight := &Insight{
		ID:          fmt.Sprintf("insight-pattern-%s-%d", pattern.NodeID, time.Now().UnixNano()),
		Type:        InsightPattern,
		Title:       fmt.Sprintf("%s pattern detected on %s", pattern.Type, pattern.NodeID),
		Description: ig.patternDescription(pattern),
		Confidence:  pattern.Confidence,
		Level:       FromFloat64(pattern.Confidence),
		Severity:    ig.severityFromPatternSeverity(pattern.Severity),
		Evidence:    ig.patternEvidence(pattern),
		Suggestions: suggestions,
		ComponentID: string(pattern.NodeID),
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(ig.ttl),
	}

	return insight
}

// patternDescription generates a description for a pattern.
func (ig *InsightsGenerator) patternDescription(pattern *DetectedPattern) string {
	switch pattern.Type {
	case PatternOscillation:
		return fmt.Sprintf("Component %s shows oscillating behavior (value changes back and forth)", pattern.NodeID)
	case PatternSameField:
		return fmt.Sprintf("Field '%s' on component %s changed multiple times rapidly", pattern.Metadata["field"], pattern.NodeID)
	case PatternHighFrequency:
		return fmt.Sprintf("Component %s is updating at high frequency", pattern.NodeID)
	case PatternBurst:
		return fmt.Sprintf("Component %s shows burst update pattern", pattern.NodeID)
	case PatternCascadeBurst:
		return fmt.Sprintf("Component %s triggered cascading updates", pattern.NodeID)
	case PatternLayoutRevert:
		return fmt.Sprintf("Component %s had layout immediately reverted", pattern.NodeID)
	default:
		return fmt.Sprintf("Pattern detected on component %s", pattern.NodeID)
	}
}

// patternEvidence generates evidence from a pattern.
func (ig *InsightsGenerator) patternEvidence(pattern *DetectedPattern) []Evidence {
	evidence := make([]Evidence, 0, len(pattern.Evidence))

	for _, e := range pattern.Evidence {
		ev := Evidence{
			Source: "PatternDetector",
		}

		// Extract metric from description if possible
		if len(e.Data) > 0 {
			if rate, ok := e.Data["rate"].(float64); ok {
				ev.Metric = "update_rate"
				ev.Value = rate
				ev.Expected = 60.0 // 60 Hz threshold
			}
			if count, ok := e.Data["count"].(int); ok {
				ev.Metric = "change_count"
				ev.Value = float64(count)
				ev.Expected = 5.0
			}
		}

		evidence = append(evidence, ev)
	}

	return evidence
}

// severityFromPatternSeverity converts pattern severity to insight severity.
func (ig *InsightsGenerator) severityFromPatternSeverity(ps PatternSeverity) InsightSeverity {
	switch ps {
	case PatternSeverityHigh:
		return SeverityCritical
	case PatternSeverityMedium:
		return SeverityWarning
	case PatternSeverityLow:
		return SeverityWarning
	default:
		return SeverityInfo
	}
}

// generateSuggestions generates suggestions for a pattern.
func (ig *InsightsGenerator) generateSuggestions(pattern *DetectedPattern) []Suggestion {
	suggestions := make([]Suggestion, 0)

	switch pattern.Type {
	case PatternOscillation:
		suggestions = append(suggestions, Suggestion{
			Priority:       0,
			Action:         "Review state update logic for oscillating conditions",
			Reason:         "Oscillating values indicate conditional logic may be unstable",
			ExpectedImpact: "Stable component behavior",
			Confidence:     pattern.Confidence,
		})

	case PatternSameField:
		fieldName := ""
		if f, ok := pattern.Metadata["field"].(string); ok {
			fieldName = f
		}
		suggestions = append(suggestions, Suggestion{
			Priority:       0,
			Action:         fmt.Sprintf("Coalesce multiple updates to field '%s'", fieldName),
			Reason:         "Multiple rapid changes to the same field can be batched",
			ExpectedImpact: "Reduced render count",
			Confidence:     pattern.Confidence,
		})

	case PatternHighFrequency:
		suggestions = append(suggestions, Suggestion{
			Priority:       0,
			Action:         "Consider debouncing or throttling updates",
			Reason:         "High frequency updates may overwhelm the render pipeline",
			ExpectedImpact: "Reduced CPU usage",
			Confidence:     pattern.Confidence,
		})

	case PatternBurst:
		suggestions = append(suggestions, Suggestion{
			Priority:       1,
			Action:         "Review batch update mechanisms",
			Reason:         "Burst patterns indicate opportunities for batching",
			ExpectedImpact: "Smoother frame times",
			Confidence:     pattern.Confidence,
		})

	case PatternLayoutRevert:
		suggestions = append(suggestions, Suggestion{
			Priority:       0,
			Action:         "Review layout calculation logic",
			Reason:         "Layout was immediately reverted, indicating unnecessary calculation",
			ExpectedImpact: "Reduced layout thrashing",
			Confidence:     pattern.Confidence,
		})
	}

	return suggestions
}

// AddInsight adds an insight to the generator.
func (ig *InsightsGenerator) AddInsight(insight *Insight) {
	ig.mu.Lock()
	defer ig.mu.Unlock()

	// Check if insight already exists
	if idx, exists := ig.insightIndex[insight.ID]; exists {
		ig.insights[idx] = insight
		return
	}

	// Add new insight
	ig.insights = append(ig.insights, insight)
	ig.insightIndex[insight.ID] = len(ig.insights) - 1

	// Prune if too many
	if len(ig.insights) > ig.maxInsights {
		// Remove oldest
		oldest := ig.insights[0]
		delete(ig.insightIndex, oldest.ID)
		ig.insights = ig.insights[1:]

		// Rebuild index
		ig.insightIndex = make(map[string]int)
		for i, ins := range ig.insights {
			ig.insightIndex[ins.ID] = i
		}
	}
}

// GetAllInsights returns all insights.
func (ig *InsightsGenerator) GetAllInsights() []*Insight {
	ig.mu.RLock()
	defer ig.mu.RUnlock()

	// Filter expired insights
	result := make([]*Insight, 0)
	now := time.Now()

	for _, insight := range ig.insights {
		if insight.ExpiresAt.IsZero() || now.Before(insight.ExpiresAt) {
			result = append(result, insight)
		}
	}

	return result
}

// GetInsightsByType returns insights filtered by type.
func (ig *InsightsGenerator) GetInsightsByType(insightType InsightType) []*Insight {
	ig.mu.RLock()
	defer ig.mu.RUnlock()

	result := make([]*Insight, 0)
	now := time.Now()

	for _, insight := range ig.insights {
		if insight.Type == insightType && (insight.ExpiresAt.IsZero() || now.Before(insight.ExpiresAt)) {
			result = append(result, insight)
		}
	}

	return result
}

// GetHighConfidenceInsights returns insights above a confidence threshold.
func (ig *InsightsGenerator) GetHighConfidenceInsights(threshold float64) []*Insight {
	ig.mu.RLock()
	defer ig.mu.RUnlock()

	result := make([]*Insight, 0)
	now := time.Now()

	for _, insight := range ig.insights {
		if insight.Confidence >= threshold && (insight.ExpiresAt.IsZero() || now.Before(insight.ExpiresAt)) {
			result = append(result, insight)
		}
	}

	return result
}

// CleanupExpired removes expired insights.
func (ig *InsightsGenerator) CleanupExpired() {
	ig.mu.Lock()
	defer ig.mu.Unlock()

	now := time.Now()
	active := make([]*Insight, 0)
	ig.insightIndex = make(map[string]int)

	for _, insight := range ig.insights {
		if insight.ExpiresAt.IsZero() || now.Before(insight.ExpiresAt) {
			active = append(active, insight)
			ig.insightIndex[insight.ID] = len(active) - 1
		}
	}

	ig.insights = active
}

// Reset clears all insights.
func (ig *InsightsGenerator) Reset() {
	ig.mu.Lock()
	defer ig.mu.Unlock()

	ig.insights = make([]*Insight, 0, ig.maxInsights)
	ig.insightIndex = make(map[string]int)
}

// GetStats returns insight statistics.
type InsightStats struct {
	TotalInsights      int
	ActiveInsights     int
	HighConfidenceCount int
	ByType             map[InsightType]int
	BySeverity         map[InsightSeverity]int
	AvgConfidence      float64
}

// GetStats returns statistics about generated insights.
func (ig *InsightsGenerator) GetStats() InsightStats {
	ig.mu.RLock()
	defer ig.mu.RUnlock()

	now := time.Now()
	stats := InsightStats{
		ByType:     make(map[InsightType]int),
		BySeverity: make(map[InsightSeverity]int),
	}

	totalConfidence := 0.0

	for _, insight := range ig.insights {
		if insight.ExpiresAt.IsZero() || now.Before(insight.ExpiresAt) {
			stats.ActiveInsights++
			totalConfidence += insight.Confidence
			stats.ByType[insight.Type]++
			stats.BySeverity[insight.Severity]++
			if insight.Confidence >= 0.8 {
				stats.HighConfidenceCount++
			}
		}
		stats.TotalInsights++
	}

	if stats.ActiveInsights > 0 {
		stats.AvgConfidence = totalConfidence / float64(stats.ActiveInsights)
	}

	return stats
}
