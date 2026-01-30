// DevTools Interactive Panel - Using Mint TUI Framework
package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/wwsheng009/mint/devtools"
	"github.com/wwsheng009/mint/devtools/client"
	"github.com/wwsheng009/mint/devtools/observation"
	v1 "github.com/wwsheng009/mint/devtools/observation/v1"
	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// Global State
// =============================================================================

var (
	debugState struct {
		sync.RWMutex
		currentView   View
		selectedFrame int
		selectedTab   int
		frameCount    int
		mutationCount int
		lastEvent     string
		eventLog      []string
	}
)

type View int

const (
	ViewTimeline View = iota
	ViewCausal
	ViewStats
	ViewPatterns
)

func init() {
	debugState.currentView = ViewTimeline
	debugState.selectedTab = 0
	debugState.selectedFrame = 0
	debugState.frameCount = 0
	debugState.mutationCount = 0
	debugState.eventLog = make([]string, 0, 8)
}

func logEvent(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	debugState.Lock()
	debugState.lastEvent = msg
	debugState.eventLog = append(debugState.eventLog, msg)
	if len(debugState.eventLog) > 6 {
		debugState.eventLog = debugState.eventLog[1:]
	}
	debugState.Unlock()
	// fmt.Printf("[DEBUG] %s\n", msg)
}

// =============================================================================
// DevTools Panel Component
// =============================================================================

type DevToolsPanel struct {
	*component.BaseComponent
	*component.StateHolder

	devtools    *devtools.DevTools
	panel       *client.TuiDebugPanel
	observation *observation.Layer
	tabs        []*Tab
	width       int
	height      int
	lastUpdate  time.Time
}

func NewDevToolsPanel(id string, width, height int) *DevToolsPanel {
	dt := devtools.New()
	dt.Enable()

	p := client.NewTuiDebugPanel(dt)
	p.Enable()
	p.SetSize(width, height)

	cfg := observation.DefaultConfig()
	cfg.InitialLevel = v1.LevelAdvanced
	obs := observation.NewLayer(cfg)
	obs.LinkComponents()
	obs.Enable(v1.LevelAdvanced)

	panel := &DevToolsPanel{
		BaseComponent: component.NewBaseComponent(id),
		StateHolder:   component.NewStateHolder(),
		devtools:      dt,
		panel:         p,
		observation:   obs,
		tabs:          make([]*Tab, 4),
		width:         width,
		height:        height,
		lastUpdate:    time.Now(),
	}

	// Create tabs
	tabNames := []string{"Timeline", "Causal", "Stats", "Patterns"}
	tabX := 2
	for i, name := range tabNames {
		panel.tabs[i] = NewTab(fmt.Sprintf("tab-%d", i), name, tabX, 2, i)
		tabX += len(name) + 4
	}

	return panel
}

// Measure returns the preferred size
func (p *DevToolsPanel) Measure(maxWidth, maxHeight int) (width, height int) {
	return p.width, p.height
}

// Paint renders the panel
func (p *DevToolsPanel) Paint(ctx component.PaintContext, buf *paint.Buffer) {
	x := ctx.X
	y := ctx.Y
	width := ctx.AvailableWidth
	height := ctx.AvailableHeight

	// Update simulation data on each paint
	p.updateSimulation()

	// Clear area
	emptyStyle := style.Style{}
	for py := y; py < y+height; py++ {
		for px := x; px < x+width; px++ {
			buf.SetCell(px, py, ' ', emptyStyle)
		}
	}

	// Header
	headerStyle := style.Style{}.Foreground(style.Cyan).Bold(true)
	buf.SetString(x+2, y, " DevTools Interactive Panel ", headerStyle)

	// Border
	borderStyle := style.Style{}.Foreground(style.BrightBlack)
	for i := x; i < x+width; i++ {
		buf.SetCell(i, y+1, '─', borderStyle)
	}

	// Status
	debugState.RLock()
	status := fmt.Sprintf("Frame: %d | Mutations: %d", debugState.frameCount, debugState.mutationCount)
	debugState.RUnlock()
	buf.SetString(x+width-len(status)-2, y, status, style.Style{}.Foreground(style.White))

	// Tabs
	for _, tab := range p.tabs {
		tab.Paint(buf)
	}

	// Content area
	contentY := y + 4
	contentHeight := height - 6

	debugState.RLock()
	currentView := debugState.currentView
	selectedFrame := debugState.selectedFrame
	debugState.RUnlock()

	switch currentView {
	case ViewTimeline:
		p.paintTimelineView(buf, x, contentY, width, contentHeight, selectedFrame)
	case ViewCausal:
		p.paintCausalView(buf, x, contentY, width, contentHeight, selectedFrame)
	case ViewStats:
		p.paintStatsView(buf, x, contentY, width, contentHeight)
	case ViewPatterns:
		p.paintPatternsView(buf, x, contentY, width, contentHeight)
	}

	// Footer
	footerY := y + height - 3
	for i := x; i < x+width; i++ {
		buf.SetCell(i, footerY, '─', borderStyle)
	}

	// Help
	helpStyle := style.Style{}.Foreground(style.Yellow)
	help := " [Tab] Switch  [←/→] Navigate  [Q]uit "
	helpX := x + (width-len(help))/2
	buf.SetString(helpX, footerY+1, help, helpStyle)

	// Last event
	debugState.RLock()
	lastEvent := debugState.lastEvent
	debugState.RUnlock()
	if lastEvent != "" && len(lastEvent) < width-4 {
		eventStyle := style.Style{}.Foreground(style.Magenta)
		buf.SetString(x+2, footerY+1, lastEvent, eventStyle)
	}
}

func (p *DevToolsPanel) paintTimelineView(buf *paint.Buffer, x, y, width, height int, selectedFrame int) {
	titleStyle := style.Style{}.Foreground(style.Cyan).Bold(true)
	labelStyle := style.Style{}.Foreground(style.Green)
	valueStyle := style.Style{}.Foreground(style.White)

	buf.SetString(x+2, y, "Timeline View (Last 30 Frames)", titleStyle)

	// Frame chart
	chartY := y + 2
	debugState.RLock()
	frameCount := debugState.frameCount
	debugState.RUnlock()

	chartWidth := width - 4
	if chartWidth > 60 {
		chartWidth = 60
	}

	for i := 0; i < chartWidth; i++ {
		frameNum := frameCount - chartWidth + i
		cellStyle := style.Style{}.Foreground(style.Blue)

		if frameNum < 0 {
			cellStyle = style.Style{}.Foreground(style.BrightBlack)
		} else if frameNum == selectedFrame {
			cellStyle = style.Style{}.Foreground(style.Magenta).Bold(true)
		}

		buf.SetCell(x+2+i, chartY, '█', cellStyle)
	}

	// Details
	detailY := chartY + 2
	buf.SetString(x+4, detailY, "Selected Frame:", labelStyle)
	buf.SetString(x+20, detailY, fmt.Sprintf("%d", selectedFrame), valueStyle)

	buf.SetString(x+4, detailY+1, "Events:", labelStyle)
	buf.SetString(x+20, detailY+1, fmt.Sprintf("%d", selectedFrame%5), valueStyle)

	buf.SetString(x+4, detailY+2, "Mutations:", labelStyle)
	buf.SetString(x+20, detailY+2, fmt.Sprintf("%d", selectedFrame%3+1), valueStyle)

	helpStyle := style.Style{}.Foreground(style.Yellow)
	buf.SetString(x+4, detailY+4, "Use ←/→ to navigate frames", helpStyle)
}

func (p *DevToolsPanel) paintCausalView(buf *paint.Buffer, x, y, width, height, selectedFrame int) {
	titleStyle := style.Style{}.Foreground(style.Cyan).Bold(true)
	chainStyle := style.Style{}.Foreground(style.Green)
	exampleStyle := style.Style{}.Foreground(style.White)

	buf.SetString(x+2, y, "Causal Chain View", titleStyle)

	// Causal diagram
	chainY := y + 2
	chain := "[Events] → [Mutations] → [Layout] → [Repaint]"
	buf.SetString(x+4, chainY, chain, chainStyle)

	// Example
	exampleY := chainY + 2
	examples := []string{
		"keypress (event)",
		"  → setState (mutation)",
		"    → calculateLayout (layout)",
		"      → drawScreen (repaint)",
	}

	for i, line := range examples {
		lineStyle := exampleStyle
		if i == 0 {
			lineStyle = style.Style{}.Foreground(style.Cyan)
		}
		buf.SetString(x+6, exampleY+i, line, lineStyle)
	}

	// Analysis
	infoY := exampleY + 6
	buf.SetString(x+4, infoY, "Current Frame Analysis:", titleStyle)
	buf.SetString(x+4, infoY+1, fmt.Sprintf("Frame %d triggered by: user_input", selectedFrame), exampleStyle)
	buf.SetString(x+4, infoY+2, "Affected components: 3", exampleStyle)
}

func (p *DevToolsPanel) paintStatsView(buf *paint.Buffer, x, y, width, height int) {
	titleStyle := style.Style{}.Foreground(style.Cyan).Bold(true)
	labelStyle := style.Style{}.Foreground(style.Green)
	valueStyle := style.Style{}.Foreground(style.White)

	buf.SetString(x+2, y, "Statistics Summary", titleStyle)

	statsY := y + 2
	metrics := p.observation.GetMetrics()

	buf.SetString(x+4, statsY, "Total Frames:", labelStyle)
	buf.SetString(x+20, statsY, fmt.Sprintf("%d", metrics.TotalFrames), valueStyle)

	buf.SetString(x+4, statsY+1, "Total Mutations:", labelStyle)
	buf.SetString(x+20, statsY+1, fmt.Sprintf("%d", metrics.TotalMutations), valueStyle)

	buf.SetString(x+4, statsY+2, "Total Layouts:", labelStyle)
	buf.SetString(x+20, statsY+2, fmt.Sprintf("%d", metrics.TotalLayouts), valueStyle)

	// Top components
	topY := statsY + 4
	buf.SetString(x+4, topY, "Top Components:", titleStyle)

	topN := p.observation.GetTopN(v1.MetricMutations, 5)
	for i, rank := range topN {
		line := fmt.Sprintf("%d. %s: %d mutations", i+1, rank.NodeID, rank.Value)
		buf.SetString(x+6, topY+1+i, line, valueStyle)
	}

	if len(topN) == 0 {
		buf.SetString(x+6, topY+1, "No components yet.", style.Style{}.Foreground(style.BrightBlack))
	}
}

func (p *DevToolsPanel) paintPatternsView(buf *paint.Buffer, x, y, width, height int) {
	titleStyle := style.Style{}.Foreground(style.Cyan).Bold(true)
	patternStyle := style.Style{}.Foreground(style.Yellow)
	infoStyle := style.Style{}.Foreground(style.White)

	buf.SetString(x+2, y, "Detected Patterns", titleStyle)

	patternsY := y + 2
	allPatterns := p.observation.GetAllPatterns()

	if len(allPatterns) == 0 {
		buf.SetString(x+4, patternsY, "No patterns detected yet.", style.Style{}.Foreground(style.BrightBlack))
		buf.SetString(x+4, patternsY+1, "Generate more activity to see patterns.", style.Style{}.Foreground(style.BrightBlack))
		return
	}

	// Display patterns
	count := 0
	for nodeID, pats := range allPatterns {
		if count >= 8 {
			break
		}

		buf.SetString(x+4, patternsY+count, fmt.Sprintf("%s:", nodeID), titleStyle)

		for _, pat := range pats {
			if count >= 10 {
				break
			}
			line := fmt.Sprintf("  • %s (%d%%)", pat.Type, int(pat.Confidence*100))
			buf.SetString(x+4, patternsY+count+1, line, patternStyle)
			count++
		}
		count++
	}

	// Legend
	legendY := patternsY + count + 2
	buf.SetString(x+4, legendY, "Pattern Types:", style.Style{}.Foreground(style.Green))

	legends := []string{
		"  • Oscillation - A→B→A→B value changes",
		"  • SameField - Rapid same-field updates",
		"  • HighFrequency - >60 updates/sec",
	}
	for i, line := range legends {
		buf.SetString(x+4, legendY+1+i, line, infoStyle)
	}
}

func (p *DevToolsPanel) SetTabFocus(tabIndex int) {
	debugState.Lock()
	defer debugState.Unlock()

	for i, tab := range p.tabs {
		tab.SetFocus(i == tabIndex)
	}

	debugState.selectedTab = tabIndex
	switch tabIndex {
	case 0:
		debugState.currentView = ViewTimeline
	case 1:
		debugState.currentView = ViewCausal
	case 2:
		debugState.currentView = ViewStats
	case 3:
		debugState.currentView = ViewPatterns
	}
	p.MarkDirty()
}

func (p *DevToolsPanel) HandleEvent(ev event.Event) bool {
	switch ev.Type() {
	case event.EventKeyPress:
		if keyEv, ok := ev.(*event.KeyEvent); ok {
			return p.handleKeyPress(keyEv)
		}
	}
	return false
}

func (p *DevToolsPanel) handleKeyPress(keyEv *event.KeyEvent) bool {
	// Check for special keys first
	switch keyEv.Special {
	case event.KeyEscape:
		// Signal quit by returning a special event
		return false // Let app handle quit
	case event.KeyTab:
		newTab := (debugState.selectedTab + 1) % 4
		p.SetTabFocus(newTab)
		logEvent("Switched to tab %d", newTab)
		return true
	case event.KeyLeft:
		debugState.Lock()
		if debugState.selectedFrame > 0 {
			debugState.selectedFrame--
		}
		debugState.Unlock()
		p.MarkDirty()
		return true
	case event.KeyRight:
		debugState.Lock()
		debugState.selectedFrame++
		debugState.Unlock()
		p.MarkDirty()
		return true
	}

	// Check for regular keys
	switch keyEv.Key.Rune {
	case 'q', 'Q':
		return false // Let app handle quit
	case '1', '2', '3', '4':
		tab := int(keyEv.Key.Rune - '1')
		p.SetTabFocus(tab)
		logEvent("Switched to tab %d", tab)
		return true
	case 't', 'T':
		p.SetTabFocus(0)
		return true
	case 'c', 'C':
		p.SetTabFocus(1)
		return true
	case 's', 'S':
		p.SetTabFocus(2)
		return true
	case 'p', 'P':
		p.SetTabFocus(3)
		return true
	}

	return false
}

func (p *DevToolsPanel) updateSimulation() {
	// Throttle updates to once per second (for better visualization)
	now := time.Now()
	if now.Sub(p.lastUpdate) < time.Second {
		return
	}
	p.lastUpdate = now

	// Simulate activity
	debugState.Lock()
	debugState.frameCount++
	p.devtools.BeginFrame()
	p.devtools.RecordEvent("sim", "node-1", "bubble", nil)
	p.devtools.EndFrame()

	// Simulate mutations
	frameCount := debugState.frameCount
	if frameCount%3 == 0 {
		nodeID := devtools.NodeID(fmt.Sprintf("node-%d", frameCount%5))
		p.observation.RecordMutation(nodeID, "state_update", frameCount%3)
		debugState.mutationCount++
	}
	debugState.Unlock()

	// Update panel
	p.panel.SetSelectedFrame(devtools.FrameID(debugState.selectedFrame))
	p.MarkDirty()
}

func (p *DevToolsPanel) Cleanup() {
	p.panel.Disable()
	p.observation.Reset()
	p.devtools.Disable()
	_ = p.devtools.Shutdown()
}

// =============================================================================
// Tab Component
// =============================================================================

type Tab struct {
	*component.BaseComponent
	text   string
	x, y   int
	width  int
	index  int
	focused bool
}

func NewTab(id, text string, x, y, index int) *Tab {
	return &Tab{
		BaseComponent: component.NewBaseComponent(id),
		text:          text,
		x:             x,
		y:             y,
		width:         len(text) + 4,
		index:         index,
		focused:       false,
	}
}

func (t *Tab) Paint(buf *paint.Buffer) {
	tabStyle := style.Style{}.Foreground(style.BrightBlack)
	if t.focused {
		tabStyle = style.Style{}.Foreground(style.Magenta).Bold(true)
	}

	text := " " + t.text + " "
	if t.focused {
		text = "[" + t.text + "]"
	}

	buf.SetString(t.x, t.y, text, tabStyle)
}

func (t *Tab) SetFocus(focus bool) {
	t.focused = focus
}

// =============================================================================
// Main
// =============================================================================

func main() {
	defer func() {
		fmt.Print("\x1b[?25h") // Show cursor
		fmt.Print("\x1b[0m")  // Reset style
		fmt.Print("\x1b[H")   // Move cursor to top-left
		fmt.Println()
	}()

	const (
		width  = 80
		height = 24
	)

	root := NewDevToolsPanel("devtools-panel", width, height)
	app := framework.NewApp()
	app.SetRoot(root)

	// Subscribe to keyboard events for quit
	app.OnEvent(event.EventKeyPress, event.EventHandlerFunc(func(ev event.Event) bool {
		if keyEv, ok := ev.(*event.KeyEvent); ok {
			if keyEv.Special == event.KeyEscape || keyEv.Key.Rune == 'q' || keyEv.Key.Rune == 'Q' {
				root.Cleanup()
				app.Quit()
				return true
			}
		}
		return false
	}))

	// Print intro
	fmt.Println("=== DevTools Interactive Panel ===")
	fmt.Println()
	fmt.Println("Controls:")
	fmt.Println("  [Tab]     - Switch between views")
	fmt.Println("  [1-4]     - Direct tab selection")
	fmt.Println("  [←/→]     - Navigate frames (Timeline)")
	fmt.Println("  [Q]       - Quit")
	fmt.Println()
	fmt.Println("Views:")
	fmt.Println("  1. Timeline - Frame timeline and navigation")
	fmt.Println("  2. Causal   - Causal chain visualization")
	fmt.Println("  3. Stats    - Component statistics")
	fmt.Println("  4. Patterns - Detected behavioral patterns")
	fmt.Println()

	logEvent("DevTools panel started...")

	// Run app
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	fmt.Println("\nDevTools panel exited. Goodbye!")
}
