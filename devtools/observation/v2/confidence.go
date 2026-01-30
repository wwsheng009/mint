// Package v2 provides pattern detection and analysis for DevTools.
//
// This file implements the confidence model with 5-signal scoring.
package v2

import (
	"math"
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// ConfidenceLevel represents the confidence level.
type ConfidenceLevel int

const (
	ConfidenceNone ConfidenceLevel = iota // < 0.4
	ConfidenceLow                           // 0.4-0.6
	ConfidenceMedium                         // 0.6-0.8
	ConfidenceHigh                           // 0.8-0.9
	ConfidenceVeryHigh                       // > 0.9
)

// String returns the string representation.
func (cl ConfidenceLevel) String() string {
	switch cl {
	case ConfidenceVeryHigh:
		return "Very High"
	case ConfidenceHigh:
		return "High"
	case ConfidenceMedium:
		return "Medium"
	case ConfidenceLow:
		return "Low"
	default:
		return "None"
	}
}

// FromFloat64 converts a float64 confidence to ConfidenceLevel.
func FromFloat64(c float64) ConfidenceLevel {
	switch {
	case c < 0.4:
		return ConfidenceNone
	case c < 0.6:
		return ConfidenceLow
	case c < 0.8:
		return ConfidenceMedium
	case c < 0.9:
		return ConfidenceHigh
	default:
		return ConfidenceVeryHigh
	}
}

// SignalScores represents the 5 types of signal scores.
type SignalScores struct {
	Statistical float64 // 0.0-1.0: How anomalous is the value statistically
	Pattern     float64 // 0.0-1.0: How much does it look like a bug pattern
	Causal      float64 // 0.0-1.0: Does it actually cause performance issues
	Context     float64 // 0.0-1.0: Is the context appropriate (penalty)
	Historical  float64 // 0.0-1.0: Is this a persistent/recurring issue
}

// ConfidenceWeights represents the weights for each signal type.
type ConfidenceWeights struct {
	Statistical float64 // Default 0.25
	Pattern     float64 // Default 0.20
	Causal      float64 // Default 0.30 (highest weight)
	Context     float64 // Default -0.10 (penalty)
	Historical  float64 // Default 0.15
}

// DefaultWeights returns the default confidence weights.
func DefaultWeights() *ConfidenceWeights {
	return &ConfidenceWeights{
		Statistical: 0.25,
		Pattern:     0.20,
		Causal:      0.30,
		Context:     -0.10,
		Historical:  0.15,
	}
}

// ConfidenceModel calculates confidence based on multiple signal types.
type ConfidenceModel struct {
	mu      sync.RWMutex
	weights *ConfidenceWeights

	// Historical data for statistical analysis
	valueHistory map[devtools.NodeID][]float64
	historySize int
}

// NewConfidenceModel creates a new confidence model.
func NewConfidenceModel() *ConfidenceModel {
	return &ConfidenceModel{
		weights:      DefaultWeights(),
		valueHistory: make(map[devtools.NodeID][]float64),
		historySize:  100,
	}
}

// SetWeights sets custom confidence weights.
func (cm *ConfidenceModel) SetWeights(weights *ConfidenceWeights) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.weights = weights
}

// GetWeights returns the current weights.
func (cm *ConfidenceModel) GetWeights() *ConfidenceWeights {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.weights
}

// Calculate calculates overall confidence from signal scores.
func (cm *ConfidenceModel) Calculate(scores *SignalScores) float64 {
	w := cm.GetWeights()

	confidence :=
		w.Statistical*scores.Statistical +
		w.Pattern*scores.Pattern +
		w.Causal*scores.Causal +
		w.Historical*scores.Historical

	// Context is a penalty (negative weight)
	if scores.Context > 0 {
		confidence -= w.Context * scores.Context
	}

	// Clamp to [0, 1]
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	return confidence
}

// CalculateStatisticalScore calculates statistical confidence (percentile rank).
func (cm *ConfidenceModel) CalculateStatisticalScore(nodeID devtools.NodeID, value float64) float64 {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	history, exists := cm.valueHistory[nodeID]
	if !exists || len(history) == 0 {
		return 0.5 // No data, neutral confidence
	}

	// Calculate percentile rank
	lessThanCount := 0
	for _, v := range history {
		if v < value {
			lessThanCount++
		}
	}

	percentile := float64(lessThanCount) / float64(len(history))
	return percentile
}

// CalculatePatternScore calculates pattern confidence based on detected patterns.
func (cm *ConfidenceModel) CalculatePatternScore(patterns []*DetectedPattern) float64 {
	if len(patterns) == 0 {
		return 0
	}

	// Pattern bonuses from V2 design
	patternBonuses := map[PatternType]float64{
		PatternOscillation: 0.30,
		PatternSameField:   0.20,
		PatternLayoutRevert: 0.30,
		PatternCascadeBurst: 0.15,
	}

	maxScore := 0.0
	for _, p := range patterns {
		bonus := patternBonuses[p.Type]
		if bonus > maxScore {
			maxScore = bonus
		}
	}

	return maxScore
}

// CalculateCausalScore calculates causal confidence based on impact.
func (cm *ConfidenceModel) CalculateCausalScore(layoutImpact, repaintImpact int, causedFrameSpike bool) float64 {
	score := 0.0

	// Layout impact
	if layoutImpact > 0 {
		score += 0.3
	}

	// Repaint impact
	if repaintImpact > 0 {
		score += 0.3
	}

	// Caused frame spike
	if causedFrameSpike {
		score += 0.4
	}

	return score
}

// CalculateContextPenalty calculates context penalty (reduces confidence).
func (cm *ConfidenceModel) CalculateContextPenalty(scenario string) float64 {
	// Context penalties from V2 design
	contextPenalties := map[string]float64{
		"animation":     0.30,
		"user_input":    0.20,
		"loading_state": 0.20,
		"drag_drop":     0.25,
	}

	if penalty, exists := contextPenalties[scenario]; exists {
		return penalty
	}
	return 0
}

// CalculateHistoricalScore calculates historical confidence.
func (cm *ConfidenceModel) CalculateHistoricalScore(anomalyDuration time.Duration, isRegression bool, firstTime bool) float64 {
	score := 0.0

	// Persistent anomaly
	if anomalyDuration > 5*time.Minute {
		score += 0.4
	}

	// Regression from baseline
	if isRegression {
		score += 0.3
	}

	// First time anomaly (reduce confidence)
	if firstTime {
		score -= 0.2
	}

	if score < 0 {
		score = 0
	}
	return score
}

// RecordValue records a value for statistical analysis.
func (cm *ConfidenceModel) RecordValue(nodeID devtools.NodeID, value float64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	history := cm.valueHistory[nodeID]
	history = append(history, value)

	// Keep fixed window
	if len(history) > cm.historySize {
		history = history[len(history)-cm.historySize:]
	}

	cm.valueHistory[nodeID] = history
}

// GetValueHistory returns the value history for a node.
func (cm *ConfidenceModel) GetValueHistory(nodeID devtools.NodeID) []float64 {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if history, exists := cm.valueHistory[nodeID]; exists {
		result := make([]float64, len(history))
		copy(result, history)
		return result
	}
	return nil
}

// GetDistributionStats returns distribution statistics for a node.
type DistributionStats struct {
	Count   int
	Mean    float64
	StdDev  float64
	Min     float64
	Max     float64
	P50     float64
	P95     float64
	P99     float64
}

// GetDistributionStats calculates distribution statistics.
func (cm *ConfidenceModel) GetDistributionStats(nodeID devtools.NodeID) *DistributionStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	history, exists := cm.valueHistory[nodeID]
	if !exists || len(history) == 0 {
		return nil
	}

	n := len(history)
	stats := &DistributionStats{Count: n}

	// Calculate min, max, sum
	minVal := history[0]
	maxVal := history[0]
	sum := 0.0

	for _, v := range history {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
		sum += v
	}

	stats.Min = minVal
	stats.Max = maxVal
	stats.Mean = sum / float64(n)

	// Calculate variance and standard deviation
	variance := 0.0
	for _, v := range history {
		diff := v - stats.Mean
		variance += diff * diff
	}
	variance /= float64(n)
	stats.StdDev = math.Sqrt(variance)

	// Calculate percentiles (simple sorting for small datasets)
	sorted := make([]float64, n)
	copy(sorted, history)

	// Simple bubble sort
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	stats.P50 = sorted[n*50/100]
	stats.P95 = sorted[n*95/100]
	stats.P99 = sorted[n*99/100]

	return stats
}

// Reset clears all historical data.
func (cm *ConfidenceModel) Reset() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.valueHistory = make(map[devtools.NodeID][]float64)
}

// Signal represents a signal for confidence calculation.
type Signal struct {
	NodeID            devtools.NodeID
	Value             float64
	Patterns          []*DetectedPattern
	LayoutImpact      int
	RepaintImpact     int
	CausedFrameSpike  bool
	Scenario          string
	AnomalyDuration   time.Duration
	IsRegression      bool
	IsFirstTime       bool
}

// CalculateSignalConfidence calculates confidence for a signal.
func (cm *ConfidenceModel) CalculateSignalConfidence(signal *Signal) float64 {
	scores := &SignalScores{
		Statistical: cm.CalculateStatisticalScore(signal.NodeID, signal.Value),
		Pattern:     cm.CalculatePatternScore(signal.Patterns),
		Causal:      cm.CalculateCausalScore(signal.LayoutImpact, signal.RepaintImpact, signal.CausedFrameSpike),
		Context:     cm.CalculateContextPenalty(signal.Scenario),
		Historical:  cm.CalculateHistoricalScore(signal.AnomalyDuration, signal.IsRegression, signal.IsFirstTime),
	}

	return cm.Calculate(scores)
}
