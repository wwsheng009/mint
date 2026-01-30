// Package observation provides intelligent analysis for DevTools.
package observation

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// TestHotspotDetector tests the hotspot detector.
func TestHotspotDetector(t *testing.T) {
	cfg := DefaultHotspotConfig()
	// Set lower threshold for testing
	cfg.SlowFrameThreshold = time.Millisecond

	detector := NewHotspotDetector(cfg)
	if detector.IsEnabled() {
		t.Error("New detector should be disabled")
	}

	detector.Enable()
	if !detector.IsEnabled() {
		t.Error("Enable() should enable detector")
	}

	// Process some frames
	for i := 0; i < 10; i++ {
		entry := &devtools.FrameEntry{
			FrameID:    devtools.FrameID(i),
			Duration:   2 * time.Millisecond,
			LayoutTime: time.Millisecond,
			PaintTime:  time.Millisecond,
			StartTime:  time.Now(),
		}
		detector.ProcessFrame(entry)
	}

	stats := detector.GetStats()
	if stats.TotalFrames != 10 {
		t.Errorf("Expected 10 frames, got %d", stats.TotalFrames)
	}

	hotspots := detector.GetHotspots()
	if len(hotspots) == 0 {
		t.Error("Expected at least one hotspot")
	}

	detector.Reset()
	stats = detector.GetStats()
	if stats.TotalFrames != 0 {
		t.Error("Reset() should clear stats")
	}
}

// TestHotspotDetector_ComponentTime tests component-level hotspot tracking.
func TestHotspotDetector_ComponentTime(t *testing.T) {
	detector := NewHotspotDetector(DefaultHotspotConfig())
	detector.Enable()

	nodeID := devtools.NodeID("component-123")

	// Record some fast times
	for i := 0; i < 5; i++ {
		detector.RecordComponentTime(nodeID, time.Millisecond)
	}

	// Record a slow time
	detector.RecordComponentTime(nodeID, 20*time.Millisecond)

	hotspot := detector.GetHotspot(nodeID)
	if hotspot == nil {
		t.Fatal("Expected hotspot for component")
	}

	if hotspot.FrameCount != 6 {
		t.Errorf("Expected 6 frames, got %d", hotspot.FrameCount)
	}

	if hotspot.Severity == HotspotSeverityNone {
		t.Error("Expected non-none severity for slow component")
	}
}

// TestWasteDetector tests the waste detector.
func TestWasteDetector(t *testing.T) {
	detector := NewWasteDetector(DefaultWasteConfig())
	detector.Enable()

	nodeID := devtools.NodeID("component-456")

	// First render - not waste
	detector.ProcessLayout(nodeID, 1, 0x1000)

	// Same state - waste
	detector.ProcessLayout(nodeID, 1, 0x1000)
	detector.ProcessLayout(nodeID, 1, 0x1000)

	// Changed state - not waste
	detector.ProcessLayout(nodeID, 2, 0x1001)

	report := detector.GetWasteReport(nodeID)
	if report == nil {
		t.Fatal("Expected waste report for component")
	}

	if report.TotalRenders != 3 {
		t.Errorf("Expected 3 renders (first render doesn't create report), got %d", report.TotalRenders)
	}

	// Should have some wasted renders
	if report.WastedRenders == 0 {
		t.Error("Expected some wasted renders")
	}
}

// TestWasteDetector_Severity tests waste severity classification.
func TestWasteDetector_Severity(t *testing.T) {
	detector := NewWasteDetector(DefaultWasteConfig())
	detector.Enable()

	nodeID := devtools.NodeID("component-789")

	// Create 50% waste rate
	for i := 0; i < 10; i++ {
		detector.ProcessLayout(nodeID, uint32(i), uint64(i)) // Change
		detector.ProcessLayout(nodeID, uint32(i), uint64(i)) // No change (waste)
	}

	report := detector.GetWasteReport(nodeID)
	if report == nil {
		t.Fatal("Expected waste report")
	}

	if report.WasteRate < 40 || report.WasteRate > 60 {
		t.Errorf("Expected waste rate around 50%%, got %.1f%%", report.WasteRate)
	}

	if report.Severity == WasteNone {
		t.Error("Expected non-none severity for 50% waste")
	}
}

// TestJitterDetector tests the jitter detector.
func TestJitterDetector(t *testing.T) {
	cfg := DefaultJitterConfig()
	cfg.MinSamples = 5 // Lower for testing

	detector := NewJitterDetector(cfg)
	detector.Enable()

	// Add consistent frames
	for i := 0; i < 10; i++ {
		entry := &devtools.FrameEntry{
			FrameID:   devtools.FrameID(i),
			Duration:  16 * time.Millisecond,
			StartTime: time.Now(),
		}
		detector.ProcessFrame(entry)
	}

	report := detector.GetReport()
	if report == nil {
		t.Fatal("Expected report")
	}

	if report.SampleCount < 10 {
		t.Errorf("Expected at least 10 samples, got %d", report.SampleCount)
	}

	// Low jitter for consistent frames
	if report.Severity == JitterHigh {
		t.Error("Expected low jitter for consistent frames")
	}
}

// TestJitterDetector_HighJitter tests high jitter detection.
func TestJitterDetector_HighJitter(t *testing.T) {
	cfg := DefaultJitterConfig()
	cfg.MinSamples = 5

	detector := NewJitterDetector(cfg)
	detector.Enable()

	// Add highly variable frames
	durations := []time.Duration{
		5 * time.Millisecond,
		30 * time.Millisecond,
		10 * time.Millisecond,
		25 * time.Millisecond,
		5 * time.Millisecond,
		30 * time.Millisecond,
		10 * time.Millisecond,
		25 * time.Millisecond,
	}

	for i, d := range durations {
		entry := &devtools.FrameEntry{
			FrameID:   devtools.FrameID(i),
			Duration:  d,
			StartTime: time.Now(),
		}
		detector.ProcessFrame(entry)
	}

	report := detector.GetReport()
	if report == nil {
		t.Fatal("Expected report")
	}

	// Should detect jitter
	if report.CurrentJitter < 0.3 {
		t.Errorf("Expected high CV (>0.3), got %.2f", report.CurrentJitter)
	}
}

// TestBehaviorProfiler tests the behavior profiler.
func TestBehaviorProfiler(t *testing.T) {
	profiler := NewBehaviorProfiler(DefaultProfilerConfig())
	profiler.Enable()

	nodeID := devtools.NodeID("component-1001")

	// Record some updates
	for i := 0; i < 25; i++ {
		duration := time.Duration(5+i%10) * time.Millisecond
		profiler.RecordComponentUpdate(nodeID, duration, "click")
	}

	profile := profiler.GetProfile(nodeID)
	if profile == nil {
		t.Fatal("Expected profile for component")
	}

	if profile.SampleCount != 25 {
		t.Errorf("Expected 25 samples, got %d", profile.SampleCount)
	}

	// Should have baseline after min samples
	if profile.Baseline == nil {
		t.Error("Expected baseline after min samples")
	}
}

// TestBehaviorProfiler_Anomaly tests anomaly detection.
func TestBehaviorProfiler_Anomaly(t *testing.T) {
	cfg := DefaultProfilerConfig()
	cfg.MinSamples = 5
	cfg.AnomalySigma = 2.0  // Lower threshold for testing

	profiler := NewBehaviorProfiler(cfg)
	profiler.Enable()

	nodeID := devtools.NodeID("component-1002")

	// Build baseline with varying times (20 samples to trigger baseline update)
	// Vary between 8-12ms to create variance
	for i := 0; i < 20; i++ {
		duration := time.Duration(8+(i%5)) * time.Millisecond  // 8-12ms
		profiler.RecordComponentUpdate(nodeID, duration, "update")
	}

	profile := profiler.GetProfile(nodeID)
	if profile == nil {
		t.Fatal("Expected profile to exist")
	}

	if profile.Baseline == nil {
		t.Fatal("Expected baseline to be created after 20 samples")
	}

	// Send very anomalous time (10x slower)
	profiler.RecordComponentUpdate(nodeID, 100*time.Millisecond, "update")

	// Check profile again for anomaly
	profile = profiler.GetProfile(nodeID)
	if profile == nil {
		t.Fatal("Expected profile to exist")
	}

	if profile.AnomalyCount == 0 {
		t.Errorf("Expected anomaly detection for very slow render (baseline mean: %.2f ns, stdDev: %.2f ns, duration: 100ms)",
			profile.Baseline.Mean, profile.Baseline.StdDev)
	}
}

// TestBaselineComparator tests the baseline comparator.
func TestBaselineComparator(t *testing.T) {
	comparator := NewBaselineComparator(DefaultBaselineConfig())
	comparator.Enable()

	// Create baseline frames
	baselineFrames := make([]*devtools.FrameEntry, 10)
	for i := 0; i < 10; i++ {
		baselineFrames[i] = &devtools.FrameEntry{
			FrameID:   devtools.FrameID(i),
			Duration:  16 * time.Millisecond,
			StartTime: time.Now().Add(-time.Minute),
		}
	}

	err := comparator.CreateSnapshot("baseline1", baselineFrames, nil, nil)
	if err != nil {
		t.Errorf("CreateSnapshot failed: %v", err)
	}

	// Create current frames (slower)
	currentFrames := make([]*devtools.FrameEntry, 10)
	for i := 0; i < 10; i++ {
		currentFrames[i] = &devtools.FrameEntry{
			FrameID:   devtools.FrameID(i + 100),
			Duration:  25 * time.Millisecond, // Slower
			StartTime: time.Now(),
		}
	}

	comparison := comparator.CompareWithBaseline("baseline1", currentFrames)
	if comparison == nil {
		t.Fatal("Expected comparison result")
	}

	if comparison.OverallTrend != TrendDegraded {
		t.Errorf("Expected degraded trend, got %s", comparison.OverallTrend)
	}
}

// TestObservationLayer tests the observation layer coordinator.
func TestObservationLayer(t *testing.T) {
	layer := NewObservationLayer(DefaultLayerConfig())

	if layer.IsEnabled() {
		t.Error("New layer should be disabled")
	}

	layer.Enable(LevelEnhanced)
	if !layer.IsEnabled() {
		t.Error("Enable() should enable layer")
	}

	if layer.GetLevel() != LevelEnhanced {
		t.Errorf("Expected LevelEnhanced, got %s", layer.GetLevel())
	}

	// Check that appropriate detectors are enabled
	if !layer.hotspot.IsEnabled() {
		t.Error("Hotspot detector should be enabled at LevelEnhanced")
	}

	if !layer.waste.IsEnabled() {
		t.Error("Waste detector should be enabled at LevelEnhanced")
	}

	if layer.profiler.IsEnabled() {
		t.Error("Profiler should NOT be enabled at LevelEnhanced")
	}

	layer.Disable()
	if layer.IsEnabled() {
		t.Error("Disable() should disable layer")
	}
}

// TestObservationLayer_Insights tests insight generation.
func TestObservationLayer_Insights(t *testing.T) {
	layer := NewObservationLayer(DefaultLayerConfig())
	layer.Enable(LevelBasic)

	// Generate some hotspot activity
	detector := layer.hotspot
	nodeID := devtools.NodeID("component-999")

	for i := 0; i < 10; i++ {
		detector.RecordComponentTime(nodeID, 20*time.Millisecond)
	}

	// Get insights
	insights := layer.GetInsights()
	if len(insights) == 0 {
		t.Error("Expected insights to be generated")
	}

	// Check summary
	summary := layer.GetInsightSummary()
	if summary.TotalInsights == 0 {
		t.Error("Expected non-zero insight count in summary")
	}
}

// TestObservationLayer_AllLevels tests all observation levels.
func TestObservationLayer_AllLevels(t *testing.T) {
	layer := NewObservationLayer(DefaultLayerConfig())

	levels := []ObservationLevel{
		LevelBasic,
		LevelEnhanced,
		LevelAdvanced,
		LevelComplete,
	}

	for _, level := range levels {
		layer.Enable(level)

		if layer.GetLevel() != level {
			t.Errorf("Expected %s, got %s", level, layer.GetLevel())
		}

		// Verify appropriate detectors are enabled
		switch level {
		case LevelBasic:
			if !layer.hotspot.IsEnabled() {
				t.Errorf("Hotspot should be enabled at %s", level)
			}
		case LevelEnhanced:
			if !layer.hotspot.IsEnabled() || !layer.waste.IsEnabled() {
				t.Errorf("Hotspot and Waste should be enabled at %s", level)
			}
		case LevelAdvanced:
			if !layer.hotspot.IsEnabled() || !layer.waste.IsEnabled() ||
				!layer.jitter.IsEnabled() || !layer.profiler.IsEnabled() {
				t.Errorf("All except Baseline should be enabled at %s", level)
			}
		case LevelComplete:
			if !layer.hotspot.IsEnabled() || !layer.waste.IsEnabled() ||
				!layer.jitter.IsEnabled() || !layer.profiler.IsEnabled() ||
				!layer.baseline.IsEnabled() {
				t.Errorf("All detectors should be enabled at %s", level)
			}
		}
	}
}

// TestRingBuffer tests the ring buffer implementation.
func TestRingBuffer(t *testing.T) {
	rb := newRingBuffer[int](5)

	// Test basic push
	rb.Push(1)
	rb.Push(2)
	rb.Push(3)

	items := rb.GetLastN(3)
	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}

	// Test wraparound
	rb.Push(4)
	rb.Push(5)
	rb.Push(6) // Should overwrite 1

	items = rb.GetLastN(5)
	if len(items) != 5 {
		t.Errorf("Expected 5 items, got %d", len(items))
	}

	// Check values - should be [2,3,4,5,6]
	expected := []int{2, 3, 4, 5, 6}
	for i, v := range items {
		if v != expected[i] {
			t.Errorf("Item %d: expected %d, got %d", i, expected[i], v)
		}
	}

	// Test GetLastN with limit
	items = rb.GetLastN(2)
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}
}

// BenchmarkHotspotDetector_ProcessFrame benchmarks frame processing.
func BenchmarkHotspotDetector_ProcessFrame(b *testing.B) {
	detector := NewHotspotDetector(DefaultHotspotConfig())
	detector.Enable()

	entry := &devtools.FrameEntry{
		FrameID:    1,
		Duration:   16 * time.Millisecond,
		LayoutTime: 10 * time.Millisecond,
		PaintTime:  6 * time.Millisecond,
		StartTime:  time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry.FrameID = devtools.FrameID(i)
		detector.ProcessFrame(entry)
	}
}

// BenchmarkWasteDetector_ProcessLayout benchmarks layout processing.
func BenchmarkWasteDetector_ProcessLayout(b *testing.B) {
	detector := NewWasteDetector(DefaultWasteConfig())
	detector.Enable()

	nodeID := devtools.NodeID("bench-1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.ProcessLayout(nodeID, uint32(i), uint64(i))
	}
}

// BenchmarkObservationLayer_ProcessFrame benchmarks layer frame processing.
func BenchmarkObservationLayer_ProcessFrame(b *testing.B) {
	layer := NewObservationLayer(DefaultLayerConfig())
	layer.Enable(LevelComplete)

	entry := &devtools.FrameEntry{
		FrameID:    1,
		Duration:   16 * time.Millisecond,
		LayoutTime: 10 * time.Millisecond,
		PaintTime:  6 * time.Millisecond,
		StartTime:  time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry.FrameID = devtools.FrameID(i)
		layer.hotspot.ProcessFrame(entry)
		layer.waste.ProcessFrame(entry)
		layer.jitter.ProcessFrame(entry)
	}
}
