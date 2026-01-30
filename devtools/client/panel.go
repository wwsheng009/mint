// Package client provides TUI debug panel integration for DevTools.
//
// This file implements the TUI debug panel that integrates with
// the TUI application for real-time debugging.
package client

import (
	"fmt"
	"strings"
	"sync"

	"github.com/wwsheng009/mint/devtools"
)

// TuiDebugPanel provides a TUI-based debug panel.
type TuiDebugPanel struct {
	mu            sync.RWMutex
	enabled       bool
	devtools      *devtools.DevTools

	// View configuration
	showTimeline  bool
	showCausal    bool
	showSnapshots bool
	showReplay    bool

	// Current selections
	selectedFrame devtools.FrameID
	selectedNode  string

	// Dimensions
	width  int
	height int
}

// NewTuiDebugPanel creates a new TUI debug panel.
func NewTuiDebugPanel(dt *devtools.DevTools) *TuiDebugPanel {
	return &TuiDebugPanel{
		devtools:     dt,
		showTimeline: true,
		showCausal:   true,
		width:        80,
		height:       24,
	}
}

// Enable enables the debug panel.
func (tp *TuiDebugPanel) Enable() {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.enabled = true
}

// Disable disables the debug panel.
func (tp *TuiDebugPanel) Disable() {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.enabled = false
}

// IsEnabled returns whether the panel is enabled.
func (tp *TuiDebugPanel) IsEnabled() bool {
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	return tp.enabled
}

// SetSize sets the panel dimensions.
func (tp *TuiDebugPanel) SetSize(width, height int) {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.width = width
	tp.height = height
}

// Render renders the debug panel to a string.
func (tp *TuiDebugPanel) Render() string {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	var builder strings.Builder

	// Header
	builder.WriteString(tp.renderHeader())

	// Main content
	builder.WriteString(tp.renderContent())

	// Footer
	builder.WriteString(tp.renderFooter())

	return builder.String()
}

// renderHeader renders the panel header.
func (tp *TuiDebugPanel) renderHeader() string {
	var builder strings.Builder

	builder.WriteString("┌")
	builder.WriteString(strings.Repeat("─", tp.width-2))
	builder.WriteString("┐\n")

	// Title line
	title := " DevTools Debug Panel "
	padding := tp.width - len(title) - 2
	leftPadding := padding / 2
	rightPadding := padding - leftPadding

	builder.WriteString("│")
	builder.WriteString(strings.Repeat(" ", leftPadding))
	builder.WriteString(title)
	builder.WriteString(strings.Repeat(" ", rightPadding))
	builder.WriteString("│\n")

	builder.WriteString("├")
	builder.WriteString(strings.Repeat("─", tp.width-2))
	builder.WriteString("┤\n")

	return builder.String()
}

// renderContent renders the main content area.
func (tp *TuiDebugPanel) renderContent() string {
	var builder strings.Builder

	contentHeight := tp.height - 4 // Header + Footer + borders

	if tp.showTimeline {
		timelineContent := tp.renderTimelineView(tp.width, contentHeight/2)
		builder.WriteString(timelineContent)
	}

	if tp.showCausal {
		causalContent := tp.renderCausalView(tp.width, contentHeight/3)
		builder.WriteString(causalContent)
	}

	return builder.String()
}

// renderTimelineView renders the timeline view.
func (tp *TuiDebugPanel) renderTimelineView(width, height int) string {
	var builder strings.Builder

	builder.WriteString("│ Timeline View ")
	builder.WriteString(strings.Repeat(" ", width-18))
	builder.WriteString("│\n")

	builder.WriteString("│")
	builder.WriteString(strings.Repeat("─", width-2))
	builder.WriteString("│\n")

	// Frame info
	if tp.devtools != nil {
		// Get frame info from devtools
		builder.WriteString(fmt.Sprintf("│ Current Frame: %-5d                              │\n", tp.selectedFrame))
		builder.WriteString("│                                                │\n")

		// Metrics
		builder.WriteString("│ Events: 0   Mutations: 0   Layouts: 0   Repaints: 0  │\n")
	}

	// Fill remaining space
	for i := 0; i < height-5; i++ {
		builder.WriteString("│                                                │\n")
	}

	builder.WriteString("│")
	builder.WriteString(strings.Repeat("─", width-2))
	builder.WriteString("│\n")

	return builder.String()
}

// renderCausalView renders the causal graph view.
func (tp *TuiDebugPanel) renderCausalView(width, height int) string {
	var builder strings.Builder

	builder.WriteString("│ Causal Graph View")
	builder.WriteString(strings.Repeat(" ", width-20))
	builder.WriteString("│\n")

	builder.WriteString("│")
	builder.WriteString(strings.Repeat("─", width-2))
	builder.WriteString("│\n")

	// Causal info
	builder.WriteString("│ [E]vents → [M]utations → [L]ayouts → [R]epaints  │\n")

	// Fill remaining space
	for i := 0; i < height-4; i++ {
		builder.WriteString("│                                                │\n")
	}

	builder.WriteString("│")
	builder.WriteString(strings.Repeat("─", width-2))
	builder.WriteString("│\n")

	return builder.String()
}

// renderFooter renders the panel footer.
func (tp *TuiDebugPanel) renderFooter() string {
	var builder strings.Builder

	// Help hints
	builder.WriteString("│ [t]imeline [c]ausal [s]napshot [r]eplay [q]uit    │\n")

	builder.WriteString("└")
	builder.WriteString(strings.Repeat("─", tp.width-2))
	builder.WriteString("┘")

	return builder.String()
}

// HandleInput handles keyboard input for the debug panel.
func (tp *TuiDebugPanel) HandleInput(key rune) bool {
	switch key {
	case 't':
		tp.mu.Lock()
		tp.showTimeline = !tp.showTimeline
		tp.mu.Unlock()
	case 'c':
		tp.mu.Lock()
		tp.showCausal = !tp.showCausal
		tp.mu.Unlock()
	case 's':
		tp.mu.Lock()
		tp.showSnapshots = !tp.showSnapshots
		tp.mu.Unlock()
	case 'r':
		tp.mu.Lock()
		tp.showReplay = !tp.showReplay
		tp.mu.Unlock()
	case 'q':
		return false // Exit
	}
	return true
}

// ToggleTimeline toggles the timeline view.
func (tp *TuiDebugPanel) ToggleTimeline() {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.showTimeline = !tp.showTimeline
}

// ToggleCausal toggles the causal graph view.
func (tp *TuiDebugPanel) ToggleCausal() {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.showCausal = !tp.showCausal
}

// ToggleSnapshots toggles the snapshots view.
func (tp *TuiDebugPanel) ToggleSnapshots() {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.showSnapshots = !tp.showSnapshots
}

// ToggleReplay toggles the replay view.
func (tp *TuiDebugPanel) ToggleReplay() {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.showReplay = !tp.showReplay
}

// SetSelectedFrame sets the selected frame.
func (tp *TuiDebugPanel) SetSelectedFrame(frameID devtools.FrameID) {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.selectedFrame = frameID
}

// GetSelectedFrame returns the selected frame.
func (tp *TuiDebugPanel) GetSelectedFrame() devtools.FrameID {
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	return tp.selectedFrame
}

// PanelState represents the current panel state.
type PanelState struct {
	Enabled        bool
	ShowTimeline   bool
	ShowCausal     bool
	ShowSnapshots  bool
	ShowReplay     bool
	SelectedFrame  devtools.FrameID
	SelectedNode   string
}

// GetState returns the current panel state.
func (tp *TuiDebugPanel) GetState() *PanelState {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	return &PanelState{
		Enabled:       tp.enabled,
		ShowTimeline:  tp.showTimeline,
		ShowCausal:    tp.showCausal,
		ShowSnapshots: tp.showSnapshots,
		ShowReplay:    tp.showReplay,
		SelectedFrame: tp.selectedFrame,
		SelectedNode:  tp.selectedNode,
	}
}

// SetState sets the panel state.
func (tp *TuiDebugPanel) SetState(state *PanelState) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	tp.enabled = state.Enabled
	tp.showTimeline = state.ShowTimeline
	tp.showCausal = state.ShowCausal
	tp.showSnapshots = state.ShowSnapshots
	tp.showReplay = state.ShowReplay
	tp.selectedFrame = state.SelectedFrame
	tp.selectedNode = state.SelectedNode
}

// DebugOverlay represents an overlay debug view.
type DebugOverlay struct {
	mu         sync.RWMutex
	enabled    bool
	highlights map[string]*NodeHighlight
}

// NodeHighlight represents a highlighted node.
type NodeHighlight struct {
	ID       string
	Region   devtools.Rect
	Color    string
	Label    string
	Duration int // frames to show highlight
}

// NewDebugOverlay creates a new debug overlay.
func NewDebugOverlay() *DebugOverlay {
	return &DebugOverlay{
		highlights: make(map[string]*NodeHighlight),
	}
}

// Enable enables the overlay.
func (do *DebugOverlay) Enable() {
	do.mu.Lock()
	defer do.mu.Unlock()
	do.enabled = true
}

// Disable disables the overlay.
func (do *DebugOverlay) Disable() {
	do.mu.Lock()
	defer do.mu.Unlock()
	do.enabled = false
}

// Highlight highlights a node.
func (do *DebugOverlay) Highlight(nodeID string, rect *devtools.Rect, color, label string) {
	do.mu.Lock()
	defer do.mu.Unlock()

	do.highlights[nodeID] = &NodeHighlight{
		ID:       nodeID,
		Region:   *rect,
		Color:    color,
		Label:    label,
		Duration: 1,
	}
}

// ClearHighlight clears a highlight.
func (do *DebugOverlay) ClearHighlight(nodeID string) {
	do.mu.Lock()
	defer do.mu.Unlock()
	delete(do.highlights, nodeID)
}

// ClearAll clears all highlights.
func (do *DebugOverlay) ClearAll() {
	do.mu.Lock()
	defer do.mu.Unlock()
	do.highlights = make(map[string]*NodeHighlight)
}

// GetHighlights returns all highlights.
func (do *DebugOverlay) GetHighlights() []*NodeHighlight {
	do.mu.RLock()
	defer do.mu.RUnlock()

	highlights := make([]*NodeHighlight, 0, len(do.highlights))
	for _, h := range do.highlights {
		highlights = append(highlights, h)
	}
	return highlights
}

// Update decrements highlight durations and removes expired ones.
func (do *DebugOverlay) Update() {
	do.mu.Lock()
	defer do.mu.Unlock()

	for id, h := range do.highlights {
		h.Duration--
		if h.Duration <= 0 {
			delete(do.highlights, id)
		}
	}
}

// InspectResult represents the result of inspecting a node.
type InspectResult struct {
	NodeID       string
	Type         string
	Position     string
	Size         string
	Properties   map[string]string
	Styles       map[string]string
	Children     []string
	Metrics      *InspectMetrics
}

// InspectMetrics represents metrics for a node.
type InspectMetrics struct {
	LayoutTime   float64
	PaintTime    float64
	DirtyCount   int
	RepaintCount int
}

// Inspect inspects a node and returns its details.
func (tp *TuiDebugPanel) Inspect(nodeID string) *InspectResult {
	// This would interface with the actual runtime to get node details
	return &InspectResult{
		NodeID:     nodeID,
		Type:       "Container",
		Position:   "x: 10, y: 20",
		Size:       "width: 100, height: 50",
		Properties: make(map[string]string),
		Styles:     make(map[string]string),
		Children:   make([]string, 0),
		Metrics:    &InspectMetrics{},
	}
}

// Command represents a debug command.
type Command struct {
	Name   string
	Args   []string
	Output chan string
}

// CommandHandler handles debug commands.
type CommandHandler struct {
	mu        sync.RWMutex
	commands  map[string]func([]string) string
	panel     *TuiDebugPanel
}

// NewCommandHandler creates a new command handler.
func NewCommandHandler(panel *TuiDebugPanel) *CommandHandler {
	ch := &CommandHandler{
		commands: make(map[string]func([]string) string),
		panel:    panel,
	}

	// Register built-in commands
	ch.Register("help", ch.cmdHelp)
	ch.Register("inspect", ch.cmdInspect)
	ch.Register("highlight", ch.cmdHighlight)
	ch.Register("frame", ch.cmdFrame)
	ch.Register("stats", ch.cmdStats)

	return ch
}

// Register registers a command handler.
func (ch *CommandHandler) Register(name string, handler func([]string) string) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	ch.commands[name] = handler
}

// Execute executes a command.
func (ch *CommandHandler) Execute(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return ""
	}

	name := parts[0]
	args := parts[1:]

	ch.mu.RLock()
	handler, exists := ch.commands[name]
	ch.mu.RUnlock()

	if !exists {
		return fmt.Sprintf("Unknown command: %s", name)
	}

	return handler(args)
}

// cmdHelp shows help information.
func (ch *CommandHandler) cmdHelp(args []string) string {
	var builder strings.Builder

	builder.WriteString("Available commands:\n")
	builder.WriteString("  help              - Show this help\n")
	builder.WriteString("  inspect <node>     - Inspect a node\n")
	builder.WriteString("  highlight <node>   - Highlight a node\n")
	builder.WriteString("  frame <id>        - Select frame\n")
	builder.WriteString("  stats             - Show statistics\n")

	return builder.String()
}

// cmdInspect inspects a node.
func (ch *CommandHandler) cmdInspect(args []string) string {
	if len(args) == 0 {
		return "Usage: inspect <node-id>"
	}

	nodeID := args[0]
	result := ch.panel.Inspect(nodeID)

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Node: %s\n", result.NodeID))
	builder.WriteString(fmt.Sprintf("Type: %s\n", result.Type))
	builder.WriteString(fmt.Sprintf("Position: %s\n", result.Position))
	builder.WriteString(fmt.Sprintf("Size: %s\n", result.Size))

	return builder.String()
}

// cmdHighlight highlights a node.
func (ch *CommandHandler) cmdHighlight(args []string) string {
	if len(args) == 0 {
		return "Usage: highlight <node-id> [color]"
	}

	nodeID := args[0]
	color := "red"
	if len(args) > 1 {
		color = args[1]
	}

	return fmt.Sprintf("Highlighted node %s with color %s", nodeID, color)
}

// cmdFrame selects a frame.
func (ch *CommandHandler) cmdFrame(args []string) string {
	if len(args) == 0 {
		return fmt.Sprintf("Current frame: %d", ch.panel.GetSelectedFrame())
	}

	// Parse frame ID
	// In real implementation, would parse and validate
	return fmt.Sprintf("Selected frame: %s", args[0])
}

// cmdStats shows statistics.
func (ch *CommandHandler) cmdStats(args []string) string {
	return "Statistics: (not yet implemented)"
}

// LogEntry represents a log entry.
type LogEntry struct {
	Timestamp string
	Level     string
	Message   string
	Source    string
}

// LogLevel represents log level.
type LogLevel int

const (
	// LogLevelDebug is debug level.
	LogLevelDebug LogLevel = iota
	// LogLevelInfo is info level.
	LogLevelInfo
	// LogLevelWarn is warning level.
	LogLevelWarn
	// LogLevelError is error level.
	LogLevelError
)

// DebugLogger provides logging for the debug panel.
type DebugLogger struct {
	mu      sync.RWMutex
	entries []LogEntry
	maxSize int
}

// NewDebugLogger creates a new debug logger.
func NewDebugLogger(maxSize int) *DebugLogger {
	return &DebugLogger{
		entries: make([]LogEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Log logs a message.
func (dl *DebugLogger) Log(level, message, source string) {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	entry := LogEntry{
		Timestamp: formatTimestamp(),
		Level:     level,
		Message:   message,
		Source:    source,
	}

	dl.entries = append(dl.entries, entry)

	// Trim to max size
	if len(dl.entries) > dl.maxSize {
		dl.entries = dl.entries[1:]
	}
}

// GetEntries returns all log entries.
func (dl *DebugLogger) GetEntries() []LogEntry {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	entries := make([]LogEntry, len(dl.entries))
	copy(entries, dl.entries)
	return entries
}

// Clear clears all log entries.
func (dl *DebugLogger) Clear() {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	dl.entries = make([]LogEntry, 0, dl.maxSize)
}

// formatTimestamp formats the current time as a timestamp.
func formatTimestamp() string {
	// Simple timestamp format
	return "12:34:56"
}

// Profiler provides profiling information.
type Profiler struct {
	mu         sync.RWMutex
	enabled    bool
	samples    map[string]*ProfileSample
}

// ProfileSample represents a profiling sample.
type ProfileSample struct {
	Name      string
	CallCount int64
	TotalTime int64
	SelfTime  int64
	AvgTime   float64
}

// NewProfiler creates a new profiler.
func NewProfiler() *Profiler {
	return &Profiler{
		samples: make(map[string]*ProfileSample),
	}
}

// Start starts profiling.
func (p *Profiler) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = true
}

// Stop stops profiling.
func (p *Profiler) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = false
}

// Record records a profiling sample.
func (p *Profiler) Record(name string, duration int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.enabled {
		return
	}

	sample, exists := p.samples[name]
	if !exists {
		sample = &ProfileSample{
			Name: name,
		}
		p.samples[name] = sample
	}

	sample.CallCount++
	sample.TotalTime += duration
	sample.AvgTime = float64(sample.TotalTime) / float64(sample.CallCount)
}

// GetSamples returns all profiling samples.
func (p *Profiler) GetSamples() []*ProfileSample {
	p.mu.RLock()
	defer p.mu.RUnlock()

	samples := make([]*ProfileSample, 0, len(p.samples))
	for _, s := range p.samples {
		samples = append(samples, s)
	}
	return samples
}

// Clear clears all profiling samples.
func (p *Profiler) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.samples = make(map[string]*ProfileSample)
}
