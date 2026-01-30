// Package client provides enhanced TUI debug panel for DevTools.
//
// This file implements an enhanced debug panel that connects to
// real DevTools data for visualization.
package client

import (
	"fmt"
	"strings"
	"sync"

	"github.com/wwsheng009/mint/devtools"
)

// EnhancedDebugPanel provides a TUI-based debug panel with real data.
type EnhancedDebugPanel struct {
	mu            sync.RWMutex
	enabled       bool
	devtools      *devtools.DevTools

	// View configuration
	showTimeline  bool
	showCausal    bool
	showSnapshots bool
	showReplay    bool
	showStats     bool

	// Current selections
	selectedFrame devtools.FrameID
	selectedNode  string

	// Visualization
	visualizer   *Visualizer

	// Dimensions
	width  int
	height int

	// Cached data
	timeline     *devtools.FrameTimeline
	causalGraph  *devtools.CausalGraph
}

// NewEnhancedDebugPanel creates a new enhanced debug panel.
func NewEnhancedDebugPanel(dt *devtools.DevTools) *EnhancedDebugPanel {
	width, height := 80, 24

	return &EnhancedDebugPanel{
		devtools:     dt,
		visualizer:   NewVisualizer(width, height),
		showTimeline: true,
		showCausal:   true,
		showStats:    true,
		width:        width,
		height:       height,
		selectedFrame: 0,
	}
}

// Enable enables the debug panel.
func (p *EnhancedDebugPanel) Enable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = true
	p.timeline = devtools.NewFrameTimeline()
	p.timeline.Enable()
}

// Disable disables the debug panel.
func (p *EnhancedDebugPanel) Disable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = false
	if p.timeline != nil {
		p.timeline.Disable()
	}
}

// SetSize sets the panel dimensions.
func (p *EnhancedDebugPanel) SetSize(width, height int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.width = width
	p.height = height
	p.visualizer.width = width
	p.visualizer.height = height
}

// Update updates the panel with current DevTools data.
func (p *EnhancedDebugPanel) Update() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.enabled {
		return
	}

	// Update timeline if needed
	// In real implementation, this would sync with DevTools data
}

// Render renders the debug panel to a string.
func (p *EnhancedDebugPanel) Render() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var builder strings.Builder

	// Header
	builder.WriteString(p.renderHeader())

	// Main content
	builder.WriteString(p.renderContent())

	// Footer
	builder.WriteString(p.renderFooter())

	return builder.String()
}

// renderHeader renders the panel header with stats.
func (p *EnhancedDebugPanel) renderHeader() string {
	var builder strings.Builder

	builder.WriteString("┌")
	builder.WriteString(strings.Repeat("─", p.width-2))
	builder.WriteString("┐\n")

	// Title line with current frame info
	title := fmt.Sprintf(" DevTools [Frame %d] ", p.selectedFrame)
	padding := p.width - len(title) - 2
	if padding < 0 {
		padding = 0
	}

	builder.WriteString("│")
	builder.WriteString(title)
	builder.WriteString(strings.Repeat(" ", padding))
	builder.WriteString("│\n")

	builder.WriteString("├")
	builder.WriteString(strings.Repeat("─", p.width-2))
	builder.WriteString("┤\n")

	return builder.String()
}

// renderContent renders the main content area.
func (p *EnhancedDebugPanel) renderContent() string {
	var builder strings.Builder

	contentHeight := p.height - 4 // Account for header/footer

	if p.showTimeline && p.timeline != nil {
		builder.WriteString(p.renderTimelineView(contentHeight / 2))
	}

	if p.showCausal {
		builder.WriteString(p.renderCausalView(contentHeight / 3))
	}

	if p.showStats {
		builder.WriteString(p.renderStatsView())
	}

	return builder.String()
}

// renderTimelineView renders the timeline view with real data.
func (p *EnhancedDebugPanel) renderTimelineView(height int) string {
	var builder strings.Builder

	if p.timeline == nil {
		return "│ Timeline not initialized\n"
	}

	builder.WriteString("│ Timeline View (last 30 frames)\n")
	builder.WriteString("│ ")
	builder.WriteString(strings.Repeat("─", p.width-4))
	builder.WriteString("\n")

	// Get frames
	frames := p.timeline.GetLastNFrames(30)
	if len(frames) == 0 {
		return "│ No frames recorded yet\n"
	}

	// Render timeline chart
	builder.WriteString(p.visualizer.RenderPerformanceChart(frames))
	builder.WriteString("\n│")

	// Render frame list
	builder.WriteString("\n│ Frame Details:\n")

	// Show last 5 frames
	start := len(frames) - 5
	if start < 0 {
		start = 0
	}

	for i := start; i < len(frames); i++ {
		frame := frames[i]
		prefix := " "
		if frame.FrameID == p.selectedFrame {
			prefix = "►"
		}

		builder.WriteString(fmt.Sprintf("│ %s[%d] %v ",
			prefix, frame.FrameID, frame.Duration))

		// Metrics
		if frame.EventCount > 0 || frame.MutationCount > 0 {
			builder.WriteString(fmt.Sprintf("(E:%d M:%d)",
				frame.EventCount, frame.MutationCount))
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

// renderCausalView renders the causal graph view.
func (p *EnhancedDebugPanel) renderCausalView(height int) string {
	var builder strings.Builder

	builder.WriteString("│ Causal Graph View\n")
	builder.WriteString("│ ")
	builder.WriteString(strings.Repeat("─", p.width-4))
	builder.WriteString("\n")

	// Get summary from current frame
	if p.timeline != nil {
		frames := p.timeline.GetAllFrames()
		if len(frames) > 0 && p.selectedFrame < devtools.FrameID(len(frames)) {
			frame := frames[int(p.selectedFrame)]
			builder.WriteString(p.visualizer.RenderFrameDetails(frame))
		}
	} else {
		builder.WriteString("│ Timeline not initialized\n")
	}

	return builder.String()
}

// renderStatsView renders statistics view.
func (p *EnhancedDebugPanel) renderStatsView() string {
	var builder strings.Builder

	builder.WriteString("│ Statistics\n")
	builder.WriteString("│ ")
	builder.WriteString(strings.Repeat("─", p.width-4))
	builder.WriteString("\n")

	// Get EventBus stats
	stats := p.devtools.GetEventBus().GetStats()
	builder.WriteString(fmt.Sprintf("│ Events Sent: %d\n", stats.EventsSent.Load()))
	builder.WriteString(fmt.Sprintf("│ Events Dropped: %d\n", stats.EventsDropped.Load()))
	builder.WriteString(fmt.Sprintf("│ Backpressure Drops: %d\n", stats.BackpressureDrops.Load()))
	builder.WriteString(fmt.Sprintf("│ Buffer Usage: %d/%d\n",
		stats.CurrentBufferLen.Load(),
		p.timeline.GetCapacity(),
	))

	return builder.String()
}

// renderFooter renders the panel footer with help.
func (p *EnhancedDebugPanel) renderFooter() string {
	var builder strings.Builder

	builder.WriteString("│\n")
	builder.WriteString("│ Keys: [←/→]Frame [T]imeline [C]ausal [S]tats [Q]uit\n")

	builder.WriteString("└")
	builder.WriteString(strings.Repeat("─", p.width-2))
	builder.WriteString("┘")

	return builder.String()
}

// HandleInput handles keyboard input for the debug panel.
func (p *EnhancedDebugPanel) HandleInput(key rune) bool {
	switch key {
	case 'q', 'Q':
		return false // Exit
	case 't', 'T':
		p.mu.Lock()
		p.showTimeline = !p.showTimeline
		p.mu.Unlock()
	case 'c', 'C':
		p.mu.Lock()
		p.showCausal = !p.showCausal
		p.mu.Unlock()
	case 's', 'S':
		p.mu.Lock()
		p.showStats = !p.showStats
		p.mu.Unlock()
	case '←':
		p.mu.Lock()
		if p.selectedFrame > 0 {
			p.selectedFrame--
		}
		p.mu.Unlock()
	case '→':
		p.mu.Lock()
		p.selectedFrame++
		p.mu.Unlock()
	}

	return true
}

// GetSelectedFrame returns the currently selected frame.
func (p *EnhancedDebugPanel) GetSelectedFrame() devtools.FrameID {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.selectedFrame
}

// SetSelectedFrame sets the selected frame.
func (p *EnhancedDebugPanel) SetSelectedFrame(frameID devtools.FrameID) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.timeline != nil {
		maxFrame := p.timeline.GetFrameCount()
		if maxFrame > 0 && int(frameID) < maxFrame {
			p.selectedFrame = frameID
		}
	}
}

// GetState returns the current panel state.
func (p *EnhancedDebugPanel) GetState() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]interface{}{
		"Enabled":       p.enabled,
		"ShowTimeline":  p.showTimeline,
		"ShowCausal":    p.showCausal,
		"ShowStats":     p.showStats,
		"SelectedFrame": p.selectedFrame,
		"Width":         p.width,
		"Height":        p.height,
	}
}

// ToggleTimeline toggles the timeline view.
func (p *EnhancedDebugPanel) ToggleTimeline() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.showTimeline = !p.showTimeline
}

// ToggleCausal toggles the causal graph view.
func (p *EnhancedDebugPanel) ToggleCausal() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.showCausal = !p.showCausal
}

// ToggleStats toggles the statistics view.
func (p *EnhancedDebugPanel) ToggleStats() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.showStats = !p.showStats
}

// Inspect inspects a component and returns detailed information.
func (p *EnhancedDebugPanel) Inspect(nodeID string) map[string]interface{} {
	// This would interface with the runtime to get component details
	return map[string]interface{}{
		"NodeID":     nodeID,
		"Type":       "Component",
		"FrameID":    p.selectedFrame,
		"Visible":     true,
		"Properties": map[string]interface{}{
			"text":     "example",
			"enabled":  true,
		},
	}
}
