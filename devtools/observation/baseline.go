// Package observation provides intelligent analysis for DevTools.
//
// This file implements the BaselineComparator for performance regression detection.
package observation

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// TrendDirection indicates the direction of a metric change.
type TrendDirection int

const (
	TrendNone TrendDirection = iota
	TrendImproved
	TrendDegraded
	TrendUnchanged
)

// String returns the string representation of the direction.
func (t TrendDirection) String() string {
	switch t {
	case TrendImproved:
		return "Improved"
	case TrendDegraded:
		return "Degraded"
	case TrendUnchanged:
		return "Unchanged"
	default:
		return "None"
	}
}

// FrameSnapshot captures the state of frames at a point in time.
type FrameSnapshot struct {
	Name         string             `json:"name"`
	Timestamp    time.Time          `json:"timestamp"`
	Duration     time.Duration      `json:"duration"`

	FrameCount   int                `json:"frame_count"`
	AvgDuration  time.Duration      `json:"avg_duration"`
	P95Duration  time.Duration      `json:"p95_duration"`
	P99Duration  time.Duration      `json:"p99_duration"`

	TotalLayouts int                `json:"total_layouts"`
	TotalPaints  int                `json:"total_paints"`
	TotalEvents  int                `json:"total_events"`

	// Component stats
	Hotspots     []ComponentHotspot `json:"hotspots"`
	WasteReports []WasteReport      `json:"waste_reports"`
}

// ComparisonResult compares a metric against its baseline.
type ComparisonResult struct {
	Metric        string         `json:"metric"`
	Current       float64        `json:"current"`
	Baseline      float64        `json:"baseline"`
	ChangePercent float64        `json:"change_percent"`
	Direction     TrendDirection `json:"direction"`
	Significance  bool           `json:"significance"` // Statistically significant
}

// BaselineComparison contains multiple comparison results.
type BaselineComparison struct {
	BaselineName  string              `json:"baseline_name"`
	BaselineTime  time.Time           `json:"baseline_time"`
	CurrentTime   time.Time           `json:"current_time"`
	Comparisons   []ComparisonResult  `json:"comparisons"`
	OverallTrend  TrendDirection      `json:"overall_trend"`
	Regressions   []string            `json:"regressions"` // Metrics that degraded
}

// BaselineComparator compares current performance against historical baselines.
type BaselineComparator struct {
	mu     sync.RWMutex
	enabled atomic.Bool

	// Stored baselines
	baselines  map[string]*FrameSnapshot

	// Stored snapshots for comparison
	snapshots  map[time.Time]*FrameSnapshot

	// Configuration
	significanceThreshold float64 // % change considered significant
	maxSnapshots          int
}

// BaselineConfig configures the BaselineComparator.
type BaselineConfig struct {
	SignificanceThreshold float64 // % change for significance
	MaxSnapshots          int     // Max snapshots to keep
}

// DefaultBaselineConfig returns the default configuration.
func DefaultBaselineConfig() *BaselineConfig {
	return &BaselineConfig{
		SignificanceThreshold: 10.0, // 10% change
		MaxSnapshots:          10,
	}
}

// NewBaselineComparator creates a new baseline comparator.
func NewBaselineComparator(cfg *BaselineConfig) *BaselineComparator {
	if cfg == nil {
		cfg = DefaultBaselineConfig()
	}

	return &BaselineComparator{
		baselines:            make(map[string]*FrameSnapshot),
		snapshots:            make(map[time.Time]*FrameSnapshot),
		significanceThreshold: cfg.SignificanceThreshold,
		maxSnapshots:          cfg.MaxSnapshots,
	}
}

// Enable enables the baseline comparator.
func (bc *BaselineComparator) Enable() {
	bc.enabled.Store(true)
}

// Disable disables the baseline comparator.
func (bc *BaselineComparator) Disable() {
	bc.enabled.Store(false)
}

// IsEnabled returns whether the comparator is enabled.
func (bc *BaselineComparator) IsEnabled() bool {
	return bc.enabled.Load()
}

// CreateSnapshot creates a named baseline snapshot.
func (bc *BaselineComparator) CreateSnapshot(name string, frames []*devtools.FrameEntry, hotspots []ComponentHotspot, wasteReports []WasteReport) error {
	if !bc.enabled.Load() {
		return nil
	}

	snapshot := bc.buildSnapshot(name, frames, hotspots, wasteReports)

	bc.mu.Lock()
	defer bc.mu.Unlock()

	bc.baselines[name] = snapshot

	return nil
}

// CreateCheckpoint creates an unnamed checkpoint for later comparison.
func (bc *BaselineComparator) CreateCheckpoint(frames []*devtools.FrameEntry, hotspots []ComponentHotspot, wasteReports []WasteReport) {
	if !bc.enabled.Load() {
		return
	}

	snapshot := bc.buildSnapshot("", frames, hotspots, wasteReports)

	bc.mu.Lock()
	defer bc.mu.Unlock()

	bc.snapshots[snapshot.Timestamp] = snapshot

	// Prune old snapshots if too many
	if len(bc.snapshots) > bc.maxSnapshots {
		bc.pruneSnapshots()
	}
}

// buildSnapshot builds a snapshot from current data.
func (bc *BaselineComparator) buildSnapshot(name string, frames []*devtools.FrameEntry, hotspots []ComponentHotspot, wasteReports []WasteReport) *FrameSnapshot {
	if len(frames) == 0 {
		return &FrameSnapshot{
			Name:      name,
			Timestamp: time.Now(),
		}
	}

	// Calculate statistics
	durations := make([]time.Duration, len(frames))
	totalLayouts := 0
	totalPaints := 0

	for i, f := range frames {
		durations[i] = f.Duration
		totalLayouts += f.LayoutCount
		totalPaints += f.RepaintCount
	}

	// Simple sort for percentiles (bubble sort for small datasets)
	for i := 0; i < len(durations)-1; i++ {
		for j := 0; j < len(durations)-i-1; j++ {
			if durations[j] > durations[j+1] {
				durations[j], durations[j+1] = durations[j+1], durations[j]
			}
		}
	}

	n := len(durations)
	sum := time.Duration(0)
	for _, d := range durations {
		sum += d
	}

	return &FrameSnapshot{
		Name:         name,
		Timestamp:    time.Now(),
		Duration:     time.Since(frames[0].StartTime),
		FrameCount:   n,
		AvgDuration:  sum / time.Duration(n),
		P95Duration:  durations[n*95/100],
		P99Duration:  durations[n*99/100],
		TotalLayouts: totalLayouts,
		TotalPaints:  totalPaints,
		Hotspots:     hotspots,
		WasteReports: wasteReports,
	}
}

// pruneSnapshots removes the oldest snapshots.
func (bc *BaselineComparator) pruneSnapshots() {
	// Find and remove oldest snapshot
	var oldestTime time.Time
	for t := range bc.snapshots {
		if oldestTime.IsZero() || t.Before(oldestTime) {
			oldestTime = t
		}
	}
	if !oldestTime.IsZero() {
		delete(bc.snapshots, oldestTime)
	}
}

// CompareWithBaseline compares current state against a named baseline.
func (bc *BaselineComparator) CompareWithBaseline(name string, frames []*devtools.FrameEntry) *BaselineComparison {
	bc.mu.RLock()
	baseline, exists := bc.baselines[name]
	bc.mu.RUnlock()

	if !exists || baseline == nil {
		return nil
	}

	current := bc.buildSnapshot("", frames, nil, nil)
	return bc.compareSnapshots(baseline, current)
}

// CompareWithRecent compares current state against the most recent checkpoint.
func (bc *BaselineComparator) CompareWithRecent(frames []*devtools.FrameEntry) *BaselineComparison {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// Find most recent snapshot
	var recent *FrameSnapshot
	var recentTime time.Time

	for t, s := range bc.snapshots {
		if recent == nil || t.After(recentTime) {
			recent = s
			recentTime = t
		}
	}

	if recent == nil {
		return nil
	}

	current := bc.buildSnapshot("", frames, nil, nil)
	return bc.compareSnapshots(recent, current)
}

// compareSnapshots performs the actual comparison.
func (bc *BaselineComparator) compareSnapshots(baseline, current *FrameSnapshot) *BaselineComparison {
	comparison := &BaselineComparison{
		BaselineName: baseline.Name,
		BaselineTime: baseline.Timestamp,
		CurrentTime:  current.Timestamp,
		Comparisons:  make([]ComparisonResult, 0),
		Regressions:  make([]string, 0),
	}

	// Compare metrics
	comparisons := []struct {
		name                     string
		current                  float64
		baseline                 float64
		lowerIsBetter           bool
	}{
		{"AvgFrameTime", float64(current.AvgDuration.Microseconds()), float64(baseline.AvgDuration.Microseconds()), true},
		{"P95FrameTime", float64(current.P95Duration.Microseconds()), float64(baseline.P95Duration.Microseconds()), true},
		{"P99FrameTime", float64(current.P99Duration.Microseconds()), float64(baseline.P99Duration.Microseconds()), true},
		{"FramesPerSecond", float64(baseline.FrameCount) / baseline.Duration.Seconds(), float64(current.FrameCount) / current.Duration.Seconds(), false},
	}

	regressionCount := 0
	improvementCount := 0

	for _, c := range comparisons {
		result := bc.compareMetric(c.name, c.current, c.baseline, c.lowerIsBetter)
		comparison.Comparisons = append(comparison.Comparisons, result)

		if result.Significance {
			if result.Direction == TrendDegraded {
				comparison.Regressions = append(comparison.Regressions, c.name)
				regressionCount++
			} else if result.Direction == TrendImproved {
				improvementCount++
			}
		}
	}

	// Determine overall trend
	if regressionCount > improvementCount {
		comparison.OverallTrend = TrendDegraded
	} else if improvementCount > regressionCount {
		comparison.OverallTrend = TrendImproved
	} else {
		comparison.OverallTrend = TrendUnchanged
	}

	return comparison
}

// compareMetric compares a single metric against baseline.
func (bc *BaselineComparator) compareMetric(name string, current, baseline float64, lowerIsBetter bool) ComparisonResult {
	changePercent := 0.0
	if baseline != 0 {
		changePercent = ((current - baseline) / baseline) * 100
	}

	direction := TrendUnchanged
	significance := abs(changePercent) >= bc.significanceThreshold

	if significance {
		if lowerIsBetter {
			if current > baseline {
				direction = TrendDegraded
			} else {
				direction = TrendImproved
			}
		} else {
			if current > baseline {
				direction = TrendImproved
			} else {
				direction = TrendDegraded
			}
		}
	}

	return ComparisonResult{
		Metric:        name,
		Current:       current,
		Baseline:      baseline,
		ChangePercent: changePercent,
		Direction:     direction,
		Significance:  significance,
	}
}

// GetBaseline retrieves a named baseline.
func (bc *BaselineComparator) GetBaseline(name string) *FrameSnapshot {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return bc.baselines[name]
}

// ListBaselines returns all baseline names.
func (bc *BaselineComparator) ListBaselines() []string {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	names := make([]string, 0, len(bc.baselines))
	for name := range bc.baselines {
		names = append(names, name)
	}
	return names
}

// DeleteBaseline removes a named baseline.
func (bc *BaselineComparator) DeleteBaseline(name string) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	delete(bc.baselines, name)
}

// ExportBaseline exports a baseline as JSON.
func (bc *BaselineComparator) ExportBaseline(name string) ([]byte, error) {
	baseline := bc.GetBaseline(name)
	if baseline == nil {
		return nil, nil
	}

	return json.MarshalIndent(baseline, "", "  ")
}

// ImportBaseline imports a baseline from JSON.
func (bc *BaselineComparator) ImportBaseline(data []byte) error {
	var snapshot FrameSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return err
	}

	bc.mu.Lock()
	defer bc.mu.Unlock()

	bc.baselines[snapshot.Name] = &snapshot
	return nil
}

// Reset clears all baseline data.
func (bc *BaselineComparator) Reset() {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	bc.baselines = make(map[string]*FrameSnapshot)
	bc.snapshots = make(map[time.Time]*FrameSnapshot)
}

// Helper function
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
