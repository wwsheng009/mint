package inspector

import (
	"testing"
	"time"
)

// TestNewPerformanceAnalyzer tests creating a new performance analyzer
func TestNewPerformanceAnalyzer(t *testing.T) {
	pa := NewPerformanceAnalyzer()

	if pa == nil {
		t.Fatal("Expected non-nil PerformanceAnalyzer")
	}

	if pa.enabled {
		t.Error("New analyzer should be disabled by default")
	}

	if pa.maxHistory != 100 {
		t.Errorf("Expected maxHistory 100, got %d", pa.maxHistory)
	}

	if pa.maxFrameTimes != 60 {
		t.Errorf("Expected maxFrameTimes 60, got %d", pa.maxFrameTimes)
	}
}

// TestEnableDisable_Performance tests enabling and disabling the analyzer
func TestEnableDisable_Performance(t *testing.T) {
	pa := NewPerformanceAnalyzer()

	// Test enable
	pa.Enable()
	if !pa.enabled {
		t.Error("Should be enabled after Enable()")
	}

	// Test disable
	pa.Disable()
	if pa.enabled {
		t.Error("Should be disabled after Disable()")
	}
}

// TestIsEnabled tests checking enabled state
func TestIsEnabled(t *testing.T) {
	pa := NewPerformanceAnalyzer()

	if pa.IsEnabled() {
		t.Error("New analyzer should not be enabled")
	}

	pa.Enable()
	if !pa.IsEnabled() {
		t.Error("Should be enabled after Enable()")
	}

	pa.Disable()
	if pa.IsEnabled() {
		t.Error("Should be disabled after Disable()")
	}
}

// TestStartEndFrame tests frame timing
func TestStartEndFrame(t *testing.T) {
	pa := NewPerformanceAnalyzer()
	pa.Enable()

	pa.StartFrame()
	time.Sleep(10 * time.Millisecond)
	pa.EndFrame()

	metrics := pa.GetMetrics()

	if metrics.FrameCount != 1 {
		t.Errorf("Expected frame count 1, got %d", metrics.FrameCount)
	}

	if metrics.LastRenderTime < 10*time.Millisecond {
		t.Errorf("Expected render time >= 10ms, got %v", metrics.LastRenderTime)
	}
}

// TestMultipleFrames tests multiple frame timing
func TestMultipleFrames(t *testing.T) {
	pa := NewPerformanceAnalyzer()
	pa.Enable()

	// Simulate 10 frames
	for i := 0; i < 10; i++ {
		pa.StartFrame()
		time.Sleep(1 * time.Millisecond)
		pa.EndFrame()
	}

	metrics := pa.GetMetrics()

	if metrics.FrameCount != 10 {
		t.Errorf("Expected frame count 10, got %d", metrics.FrameCount)
	}

	if metrics.AvgRenderTime == 0 {
		t.Error("Average render time should not be zero")
	}

	if metrics.FPS == 0 {
		t.Error("FPS should be calculated")
	}
}

// TestFPSCalculation tests FPS calculation
func TestFPSCalculation(t *testing.T) {
	pa := NewPerformanceAnalyzer()
	pa.Enable()

	// Simulate frames at ~60 FPS (16.67ms per frame)
	for i := 0; i < 60; i++ {
		pa.StartFrame()
		time.Sleep(16 * time.Millisecond)
		pa.EndFrame()
	}

	metrics := pa.GetMetrics()

	// FPS should stay in a reasonable range.
	// On Windows and under full-suite load, Sleep(16ms) often lands closer to 20ms+,
	// so allow a wider lower bound while still catching obviously broken calculations.
	if metrics.FPS < 40 || metrics.FPS > 70 {
		t.Errorf("Expected FPS around 60, got %.2f", metrics.FPS)
	}
}

// TestMemoryMetrics tests memory metric collection
func TestMemoryMetrics(t *testing.T) {
	pa := NewPerformanceAnalyzer()
	pa.Enable()

	pa.StartFrame()
	pa.EndFrame()

	metrics := pa.GetMetrics()

	if metrics.LastHeapAlloc == 0 {
		t.Error("Heap alloc should be collected")
	}

	if metrics.LastHeapSys == 0 {
		t.Error("Heap sys should be collected")
	}

	if metrics.NumGC == 0 {
		// GC might not have run yet, so we'll skip this check
		t.Log("GC hasn't run yet")
	}
}

// TestHistory tests performance history tracking
func TestHistory(t *testing.T) {
	pa := NewPerformanceAnalyzer()
	pa.Enable()

	// Generate some frames
	for i := 0; i < 5; i++ {
		pa.StartFrame()
		time.Sleep(1 * time.Millisecond)
		pa.EndFrame()
	}

	history := pa.GetHistory()

	if len(history) != 5 {
		t.Errorf("Expected 5 history entries, got %d", len(history))
	}

	// Check first snapshot
	if history[0].Timestamp.IsZero() {
		t.Error("Snapshot should have timestamp")
	}
}

// TestHistoryLimit tests history size limit
func TestHistoryLimit(t *testing.T) {
	pa := NewPerformanceAnalyzer()
	pa.maxHistory = 10 // Reduce limit for testing
	pa.history = make([]PerformanceSnapshot, 0, 10)
	pa.Enable()

	// Generate more frames than history limit
	for i := 0; i < 20; i++ {
		pa.StartFrame()
		time.Sleep(1 * time.Millisecond)
		pa.EndFrame()
	}

	history := pa.GetHistory()

	if len(history) > 10 {
		t.Errorf("History should be limited to 10, got %d", len(history))
	}
}

// TestReset tests resetting metrics
func TestReset(t *testing.T) {
	pa := NewPerformanceAnalyzer()
	pa.Enable()

	// Generate some data
	for i := 0; i < 5; i++ {
		pa.StartFrame()
		time.Sleep(1 * time.Millisecond)
		pa.EndFrame()
	}

	// Reset
	pa.Reset()

	metrics := pa.GetMetrics()

	if metrics.FrameCount != 0 {
		t.Errorf("Frame count should be 0 after reset, got %d", metrics.FrameCount)
	}

	if len(pa.history) != 0 {
		t.Errorf("History should be empty after reset, got %d entries", len(pa.history))
	}
}

// TestFormatMetrics tests metrics formatting
func TestFormatMetrics(t *testing.T) {
	pa := NewPerformanceAnalyzer()
	pa.Enable()

	pa.StartFrame()
	time.Sleep(5 * time.Millisecond)
	pa.EndFrame()

	output := pa.FormatMetrics()

	if output == "" {
		t.Error("Output should not be empty")
	}

	requiredStrings := []string{
		"Performance Metrics",
		"Frames:",
		"FPS:",
		"Memory:",
		"Heap Alloc:",
	}

	for _, s := range requiredStrings {
		if !contains(output, s) {
			t.Errorf("Output should contain '%s'", s)
		}
	}
}

// TestFormatMetrics_NoData tests formatting with no data
func TestFormatMetrics_NoData(t *testing.T) {
	pa := NewPerformanceAnalyzer()

	output := pa.FormatMetrics()

	if output != "No performance data available" {
		t.Errorf("Expected 'No performance data available', got '%s'", output)
	}
}

// TestFormatCompact_Performance tests compact formatting
func TestFormatCompact_Performance(t *testing.T) {
	pa := NewPerformanceAnalyzer()
	pa.Enable()

	pa.StartFrame()
	time.Sleep(1 * time.Millisecond)
	pa.EndFrame()

	output := pa.FormatCompact()

	if output == "" {
		t.Error("Output should not be empty")
	}

	requiredParts := []string{
		"FPS:",
		"Render:",
		"Mem:",
		"GC:",
	}

	for _, part := range requiredParts {
		if !contains(output, part) {
			t.Errorf("Output should contain '%s'", part)
		}
	}
}

// TestFormatCompact_NoData_Performance tests compact formatting with no data
func TestFormatCompact_NoData_Performance(t *testing.T) {
	pa := NewPerformanceAnalyzer()

	output := pa.FormatCompact()

	if output != "No data" {
		t.Errorf("Expected 'No data', got '%s'", output)
	}
}

// TestDisabledAnalyzer tests that disabled analyzer doesn't collect data
func TestDisabledAnalyzer(t *testing.T) {
	pa := NewPerformanceAnalyzer()
	// Don't enable

	pa.StartFrame()
	time.Sleep(1 * time.Millisecond)
	pa.EndFrame()

	metrics := pa.GetMetrics()

	if metrics.FrameCount != 0 {
		t.Errorf("Disabled analyzer should not track frames, got count %d", metrics.FrameCount)
	}
}

// TestFormatDuration tests duration formatting
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Duration
		contains string
	}{
		{"Nanoseconds", 500 * time.Nanosecond, "ns"},
		{"Microseconds", 500 * time.Microsecond, "µs"},
		{"Milliseconds", 500 * time.Millisecond, "ms"},
		{"Seconds", 2 * time.Second, "s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.input)
			if !contains(result, tt.contains) {
				t.Errorf("Expected '%s' in output, got '%s'", tt.contains, result)
			}
		})
	}
}

// TestFormatBytes tests byte formatting
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		contains string
	}{
		{"Bytes", 500, "B"},
		{"KB", 1024, "KB"},
		{"MB", 1024 * 1024, "MB"},
		{"GB", 1024 * 1024 * 1024, "GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBytes(tt.input)
			if !contains(result, tt.contains) {
				t.Errorf("Expected '%s' in output, got '%s'", tt.contains, result)
			}
		})
	}
}
