package inspector

import (
	"fmt"
	"runtime"
	"time"
)

// PerformanceMetrics tracks rendering and memory performance
type PerformanceMetrics struct {
	// Rendering metrics
	FrameCount      int64         // Total frames rendered
	TotalRenderTime time.Duration // Cumulative render time
	LastRenderTime  time.Duration // Last frame render time
	AvgRenderTime   time.Duration // Average render time per frame
	FPS             float64       // Current frames per second

	// Memory metrics
	LastHeapAlloc   uint64        // Last heap allocation
	LastHeapSys     uint64        // Last system heap memory
	LastHeapObjects uint64        // Last number of heap objects
	HeapGrowth      uint64        // Heap growth since last check
	NumGC           uint32        // Number of garbage collections
	LastGCTime      time.Duration // Last GC duration

	// Timing
	StartTime       time.Time     // When monitoring started
	LastUpdateTime  time.Time     // Last update time
}

// PerformanceAnalyzer monitors and analyzes performance
type PerformanceAnalyzer struct {
	metrics         PerformanceMetrics
	enabled         bool
	history         []PerformanceSnapshot // History of snapshots
	maxHistory      int                    // Maximum history to keep
	frameTimes      []time.Duration        // Recent frame times for FPS calc
	maxFrameTimes   int                    // Max frame times to keep
}

// PerformanceSnapshot represents metrics at a point in time
type PerformanceSnapshot struct {
	Timestamp       time.Time
	RenderTime      time.Duration
	HeapAlloc       uint64
	HeapSys         uint64
	HeapObjects     uint64
	NumGC           uint32
}

// NewPerformanceAnalyzer creates a new performance analyzer
func NewPerformanceAnalyzer() *PerformanceAnalyzer {
	return &PerformanceAnalyzer{
		enabled:       false,
		maxHistory:    100,
		maxFrameTimes: 60, // Keep 60 frames for FPS calculation (1 second at 60fps)
		frameTimes:    make([]time.Duration, 0, 60),
		history:       make([]PerformanceSnapshot, 0, 100),
		metrics: PerformanceMetrics{
			StartTime: time.Now(),
		},
	}
}

// Enable starts performance monitoring
func (pa *PerformanceAnalyzer) Enable() {
	pa.enabled = true
	pa.metrics.StartTime = time.Now()
	pa.metrics.LastUpdateTime = time.Now()
}

// Disable stops performance monitoring
func (pa *PerformanceAnalyzer) Disable() {
	pa.enabled = false
}

// IsEnabled returns whether monitoring is enabled
func (pa *PerformanceAnalyzer) IsEnabled() bool {
	return pa.enabled
}

// StartFrame marks the start of a frame render
func (pa *PerformanceAnalyzer) StartFrame() {
	if !pa.enabled {
		return
	}
	pa.metrics.LastUpdateTime = time.Now()
}

// EndFrame marks the end of a frame render
func (pa *PerformanceAnalyzer) EndFrame() {
	if !pa.enabled {
		return
	}

	renderTime := time.Since(pa.metrics.LastUpdateTime)

	// Update metrics
	pa.metrics.FrameCount++
	pa.metrics.TotalRenderTime += renderTime
	pa.metrics.LastRenderTime = renderTime
	pa.metrics.AvgRenderTime = time.Duration(
		int64(pa.metrics.TotalRenderTime) / pa.metrics.FrameCount,
	)

	// Track frame times for FPS calculation
	pa.frameTimes = append(pa.frameTimes, renderTime)
	if len(pa.frameTimes) > pa.maxFrameTimes {
		pa.frameTimes = pa.frameTimes[1:]
	}

	// Calculate FPS
	pa.updateFPS()

	// Take snapshot
	pa.takeSnapshot()
}

// updateFPS calculates current FPS based on recent frame times
func (pa *PerformanceAnalyzer) updateFPS() {
	if len(pa.frameTimes) == 0 {
		return
	}

	// Sum recent frame times
	var total time.Duration
	for _, ft := range pa.frameTimes {
		total += ft
	}

	// FPS = 1 second / average frame time
	avgFrameTime := total / time.Duration(len(pa.frameTimes))
	if avgFrameTime > 0 {
		pa.metrics.FPS = 1.0 / avgFrameTime.Seconds()
	}
}

// takeSnapshot captures current performance state
func (pa *PerformanceAnalyzer) takeSnapshot() {
	// Read memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Calculate heap growth
	if pa.metrics.LastHeapAlloc > 0 {
		pa.metrics.HeapGrowth = memStats.HeapAlloc - pa.metrics.LastHeapAlloc
	} else {
		pa.metrics.HeapGrowth = 0
	}

	// Update current memory metrics
	pa.metrics.LastHeapAlloc = memStats.HeapAlloc
	pa.metrics.LastHeapSys = memStats.HeapSys
	pa.metrics.LastHeapObjects = memStats.HeapObjects
	pa.metrics.NumGC = memStats.NumGC
	pa.metrics.LastGCTime = time.Duration(memStats.PauseTotalNs)

	// Create snapshot
	snapshot := PerformanceSnapshot{
		Timestamp:   time.Now(),
		RenderTime:  pa.metrics.LastRenderTime,
		HeapAlloc:   memStats.HeapAlloc,
		HeapSys:     memStats.HeapSys,
		HeapObjects: memStats.HeapObjects,
		NumGC:       memStats.NumGC,
	}

	pa.history = append(pa.history, snapshot)
	if len(pa.history) > pa.maxHistory {
		pa.history = pa.history[1:]
	}
}

// GetMetrics returns current performance metrics
func (pa *PerformanceAnalyzer) GetMetrics() PerformanceMetrics {
	return pa.metrics
}

// GetHistory returns performance history
func (pa *PerformanceAnalyzer) GetHistory() []PerformanceSnapshot {
	return pa.history
}

// FormatMetrics formats metrics as text
func (pa *PerformanceAnalyzer) FormatMetrics() string {
	if pa.metrics.FrameCount == 0 {
		return "No performance data available"
	}

	var lines []string
	lines = append(lines, "┌─ Performance Metrics ─────────────────────────┐")

	// Rendering
	lines = append(lines, "│ Rendering:                                      │")
	lines = append(lines, fmt.Sprintf("│   Frames: %-35d │", pa.metrics.FrameCount))
	lines = append(lines, fmt.Sprintf("│   FPS: %-37.1f │", pa.metrics.FPS))
	lines = append(lines, fmt.Sprintf("│   Last Render: %-30s │",
		formatDuration(pa.metrics.LastRenderTime)))
	lines = append(lines, fmt.Sprintf("│   Avg Render: %-30s │",
		formatDuration(pa.metrics.AvgRenderTime)))

	// Memory
	lines = append(lines, "│ Memory:                                         │")
	lines = append(lines, fmt.Sprintf("│   Heap Alloc: %-29s │",
		formatBytes(pa.metrics.LastHeapAlloc)))
	lines = append(lines, fmt.Sprintf("│   Heap Sys: %-31s │",
		formatBytes(pa.metrics.LastHeapSys)))
	lines = append(lines, fmt.Sprintf("│   Heap Objects: %-27d │",
		pa.metrics.LastHeapObjects))
	lines = append(lines, fmt.Sprintf("│   GC Count: %-30d │",
		pa.metrics.NumGC))

	lines = append(lines, "└────────────────────────────────────────────────┘")

	return joinLines(lines)
}

// FormatCompact formats metrics in a single line
func (pa *PerformanceAnalyzer) FormatCompact() string {
	if pa.metrics.FrameCount == 0 {
		return "No data"
	}

	return fmt.Sprintf("FPS: %.1f | Render: %s | Mem: %s | GC: %d",
		pa.metrics.FPS,
		formatDuration(pa.metrics.LastRenderTime),
		formatBytes(pa.metrics.LastHeapAlloc),
		pa.metrics.NumGC)
}

// Reset clears all metrics
func (pa *PerformanceAnalyzer) Reset() {
	pa.metrics = PerformanceMetrics{
		StartTime: time.Now(),
	}
	pa.history = make([]PerformanceSnapshot, 0, pa.maxHistory)
	pa.frameTimes = make([]time.Duration, 0, pa.maxFrameTimes)
}

// Helper functions

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%d ns", d.Nanoseconds())
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.1f µs", float64(d.Microseconds())/10.0)
	} else if d < time.Second {
		return fmt.Sprintf("%.2f ms", float64(d.Milliseconds())/10.0/100.0)
	} else {
		return fmt.Sprintf("%.2f s", d.Seconds())
	}
}

// formatBytes formats bytes in a human-readable way
func formatBytes(b uint64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case b < KB:
		return fmt.Sprintf("%d B", b)
	case b < MB:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(KB))
	case b < GB:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(MB))
	default:
		return fmt.Sprintf("%.2f GB", float64(b)/float64(GB))
	}
}

// joinLines joins lines with newlines
func joinLines(lines []string) string {
	result := ""
	for _, line := range lines {
		result += line + "\n"
	}
	return result[:len(result)-1] // Remove trailing newline
}
