// Package client provides visualization components for DevTools.
//
// This file implements visual renderers for different DevTools views.
package client

import (
	"fmt"
	"strings"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// Visualizer renders visual representations of DevTools data.
type Visualizer struct {
	width  int
	height int
	colors  *ColorScheme
}

// ColorScheme defines colors for visualization.
type ColorScheme struct {
	Header     string
	Border     string
	Selected   string
	Highlight  string
	Timeline   string
	Event      string
	Mutation   string
	Layout     string
	Repaint    string
	Hotspot    string
	Normal     string
}

// DefaultColorScheme returns the default color scheme.
func DefaultColorScheme() *ColorScheme {
	return &ColorScheme{
		Header:    "cyan",
		Border:    "dim",
		Selected:  "reverse",
		Highlight: "yellow",
		Timeline:  "blue",
		Event:     "green",
		Mutation:  "magenta",
		Layout:    "blue",
		Repaint:   "red",
		Hotspot:   "red",
		Normal:    "white",
	}
}

// NewVisualizer creates a new visualizer.
func NewVisualizer(width, height int) *Visualizer {
	return &Visualizer{
		width:  width,
		height: height,
		colors:  DefaultColorScheme(),
	}
}

// RenderTimeline renders the frame timeline as a visual chart.
func (v *Visualizer) RenderTimeline(timeline *devtools.FrameTimeline, selectedFrame devtools.FrameID) string {
	var builder strings.Builder

	frames := timeline.GetAllFrames()
	if len(frames) == 0 {
		return "│ No frames recorded yet\n"
	}

	// Timeline header
	builder.WriteString("│ Frame Timeline:\n")
	builder.WriteString("│ ")

	// Calculate bar width
	maxWidth := v.width - 10
	barWidth := maxWidth / len(frames)
	if barWidth < 3 {
		barWidth = 3
	}

	// Render each frame
	for i, frame := range frames {
		// Select color based on performance
		color := v.getPerformanceColor(frame.Duration)

		// Highlight selected frame
		if frame.FrameID == selectedFrame {
			builder.WriteString("[" + color + "]" + fmt.Sprintf("%4d", frame.FrameID))
		} else {
			builder.WriteString(" " + color + fmt.Sprintf("%4d", frame.FrameID))
		}

		// Show performance indicator
		indicator := v.getPerformanceIndicator(frame.Duration)
		builder.WriteString(indicator)

		// New line every 10 frames
		if (i + 1) % 10 == 0 && i < len(frames)-1 {
			builder.WriteString("\n│ ")
		}
	}
	builder.WriteString("\n")

	return builder.String()
}

// RenderCausalGraph renders the causal graph as a tree.
func (v *Visualizer) RenderCausalGraph(graph *devtools.CausalGraph) string {
	var builder strings.Builder

	if graph == nil {
		return "│ No causal graph available\n"
	}

	summary := graph.GetFrameSummary()
	builder.WriteString(fmt.Sprintf("│ Frame %d Summary:\n", summary.FrameID))
	builder.WriteString(fmt.Sprintf("│   Events: %d, Mutations: %d, Layouts: %d, Repaints: %d\n",
		summary.EventCount,
		summary.MutationCount,
		summary.LayoutCount,
		summary.RepaintCount))

	builder.WriteString("│\n│ Causal Chain:\n")
	builder.WriteString("│ [E] Event → [M] Mutation → [L] Layout → [R] Repaint\n")

	return builder.String()
}

// RenderFrameDetails renders detailed information about a frame.
func (v *Visualizer) RenderFrameDetails(frame *devtools.FrameEntry) string {
	var builder strings.Builder

	if frame == nil {
		return "│ No frame selected\n"
	}

	builder.WriteString(fmt.Sprintf("│ Frame %d Details:\n", frame.FrameID))
	builder.WriteString("│ ───────────────\n")
	builder.WriteString(fmt.Sprintf("│ Start Time:  %s\n", formatTime(frame.StartTime)))
	builder.WriteString(fmt.Sprintf("│ End Time:    %s\n", formatTime(frame.EndTime)))
	builder.WriteString(fmt.Sprintf("│ Duration:    %v\n", frame.Duration))
	builder.WriteString("│\n")
	builder.WriteString(fmt.Sprintf("│ Event Count:   %d\n", frame.EventCount))
	builder.WriteString(fmt.Sprintf("│ Layout Count:  %d\n", frame.LayoutCount))
	builder.WriteString(fmt.Sprintf("│ Repaint Count: %d\n", frame.RepaintCount))
	builder.WriteString(fmt.Sprintf("│ Edge Count:    %d\n", frame.EdgeCount))

	// Performance breakdown
	builder.WriteString("│\n│ Performance:\n")
	builder.WriteString(fmt.Sprintf("│   Layout:  %v\n", frame.LayoutTime))
	builder.WriteString(fmt.Sprintf("│   Paint:   %v\n", frame.PaintTime))
	builder.WriteString(fmt.Sprintf("│   Total:   %v\n", frame.TotalTime))

	return builder.String()
}

// RenderPerformanceChart renders a performance bar chart.
func (v *Visualizer) RenderPerformanceChart(frames []*devtools.FrameEntry) string {
	var builder strings.Builder

	if len(frames) == 0 {
		return "│ No performance data\n"
	}

	builder.WriteString("│ Performance Chart (last 30 frames):\n")
	builder.WriteString("│ ─────────────────────────────────────\n")

	// Show last 30 frames
	start := len(frames) - 30
	if start < 0 {
		start = 0
	}

	maxDuration := time.Duration(0)
	for _, f := range frames[start:] {
		if f.Duration > maxDuration {
			maxDuration = f.Duration
		}
	}

	for _, frame := range frames[start:] {
		// Calculate bar width
		width := float64(frame.Duration) / float64(maxDuration) * 40
		if width < 1 {
			width = 1
		}

		// Color based on performance
		color := v.getPerformanceColor(frame.Duration)

		// Render bar
		builder.WriteString(fmt.Sprintf("│ [%s]", color))
		builder.WriteString(strings.Repeat("■", int(width)))
		builder.WriteString(" " + fmt.Sprintf("%5v", frame.Duration))

		// Frame number
		builder.WriteString(fmt.Sprintf(" #%4d\n", frame.FrameID))
	}

	return builder.String()
}

// RenderStats renders statistics.
func (v *Visualizer) RenderStats(stats interface{}) string {
	var builder strings.Builder

	builder.WriteString("│ Statistics:\n")
	builder.WriteString("│ ───────────\n")

	// Convert stats to map[string]interface{} if possible
	if m, ok := stats.(map[string]interface{}); ok {
		for key, value := range m {
			builder.WriteString(fmt.Sprintf("│ %s: %v\n", key, value))
		}
	}

	return builder.String()
}

// RenderComponentTree renders the component tree.
func (v *Visualizer) RenderComponentTree(nodes []devtools.NodeID, selected devtools.NodeID) string {
	var builder strings.Builder

	builder.WriteString("│ Component Tree:\n")
	builder.WriteString("│ ───────────────\n")

	for _, nodeID := range nodes {
		prefix := "  "
		if nodeID == selected {
			prefix = "► "
		}

		builder.WriteString(fmt.Sprintf("│ %s[%s]\n", prefix, nodeID))
	}

	return builder.String()
}

// RenderHelp renders help information.
func (v *Visualizer) RenderHelp() string {
	var builder strings.Builder

	builder.WriteString("│ DevTools Help:\n")
	builder.WriteString("│ ─────────────\n")
	builder.WriteString("│ Keyboard Shortcuts:\n")
	builder.WriteString("│   Ctrl+D   - Toggle debug panel\n")
	builder.WriteString("│   Ctrl+T   - Toggle timeline view\n")
	builder.WriteString("│   Ctrl+C   - Toggle causal graph view\n")
	builder.WriteString("│   F1       - Show this help\n")
	builder.WriteString("│\n")
	builder.WriteString("│ Commands:\n")
	builder.WriteString("│   inspect <node>   - Inspect component\n")
	builder.WriteString("│   highlight <node>  - Highlight component\n")
	builder.WriteString("│   stats            - Show statistics\n")
	builder.WriteString("│   help             - Show commands\n")

	return builder.String()
}

// Helper functions

func (v *Visualizer) getPerformanceColor(duration time.Duration) string {
	// Color based on frame duration (assuming 60fps = 16.67ms)
	if duration < 10*time.Millisecond {
		return "green" // Good
	}
	if duration < 20*time.Millisecond {
		return "yellow" // Warning
	}
	return "red" // Slow
}

func (v *Visualizer) getPerformanceIndicator(duration time.Duration) string {
	// Visual indicator for performance
	if duration < 10*time.Millisecond {
		return "✓" // Good
	}
	if duration < 20*time.Millisecond {
		return "!" // Warning
	}
	return "✗" // Slow
}

func formatTime(t time.Time) string {
	return t.Format("15:04:05.000")
}

// FormatDuration formats a duration for display.
func FormatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%d ns", d.Nanoseconds())
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%.1f µs", float64(d.Nanoseconds())/1000)
	}
	if d < time.Second {
		return fmt.Sprintf("%.1f ms", float64(d.Microseconds())/1000)
	}
	return fmt.Sprintf("%.2f s", d.Seconds())
}
