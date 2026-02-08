// Package inspector provides standalone UI Inspector for TUI applications
//
// The standalone Inspector operates as an independent overlay interface,
// similar to browser DevTools, rather than being embedded in the application UI.
//
// Usage:
//   inspector := NewStandaloneInspector()
//   inspector.Enable()
//   inspector.AttachToApp(rootVNode)
//
// Features:
//   - F12: Toggle inspector overlay
//   - Independent rendering layer
//   - No modification to application UI tree
//   - Real-time performance and diagnostics
package inspector

import (
	"fmt"
	"os"
	"sync"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/components/display"
	"github.com/wwsheng009/mint/components/navigation"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

// StandaloneInspector operates as an independent overlay
// Similar to Chrome DevTools or browser developer tools
type StandaloneInspector struct {
	mu sync.RWMutex

	// State
	enabled    bool
	visible    bool
	activeTab  InspectorTab

	// Data sources
	treeView   *TreeView
	treeViewComponent *display.TreeView // New: TreeView component with navigation
	perf       *PerformanceAnalyzer
	diagnostics *LayoutDiagnostics
	editor     *PropertyEditor

	// VNode tracking
	appRoot    rtui.VNode
	selectedVNode rtui.VNode
	selectedPath  string

	// Overlay configuration
	overlayWidth  int
	overlayHeight int
	position      OverlayPosition

	// Floating position (for dragging)
	floatX      int  // X position when in floating mode
	floatY      int  // Y position when in floating mode
	isDragging  bool // Currently being dragged
	dragStartX  int  // Mouse X when drag started
	dragStartY  int  // Mouse Y when drag started
	floatStartX int  // Panel X when drag started
	floatStartY int  // Panel Y when drag started

	// Tree scroll state
	treeScrollOffset int  // Vertical scroll offset for tree view (in items)
	treeLines        []string // Cached tree lines for virtual scrolling
	treeTotalLines   int      // Total number of tree lines (for scroll bounds)

	// Key debug info (for displaying what keys are being pressed)
	lastKey       string  // Last key name received
	lastAlt       bool    // Last Alt modifier state
	lastCtrl      bool    // Last Ctrl modifier state
	lastShift     bool    // Last Shift modifier state
	showKeyDebug  bool    // Show key debug info in UI
}

// InspectorTab represents different inspector panels
type InspectorTab int

const (
	TabElements InspectorTab = iota
	TabConsole
	TabPerformance
	TabDiagnostics
	TabNetwork
)

// OverlayPosition defines where inspector overlay appears
type OverlayPosition int

const (
	PositionRight OverlayPosition = iota
	PositionBottom
	PositionLeft
	PositionFloating
)

// Tab names
var tabNames = map[InspectorTab]string{
	TabElements:     "Elements",
	TabConsole:      "Console",
	TabPerformance:  "Performance",
	TabDiagnostics:  "Diagnostics",
	TabNetwork:      "Network",
}

// NewStandaloneInspector creates a new standalone inspector instance
func NewStandaloneInspector() *StandaloneInspector {
	return &StandaloneInspector{
		enabled:       false,
		visible:       false,
		activeTab:     TabElements,
		treeView:      NewTreeView(),
		perf:          NewPerformanceAnalyzer(),
		diagnostics:   NewLayoutDiagnostics(),
		editor:        NewPropertyEditor(),
		overlayWidth:  80,
		overlayHeight: 25,
		position:      PositionFloating, // Change to floating by default
		// Floating position (top-right corner)
		floatX:        80,  // Default X position
		floatY:        5,   // Default Y position
		isDragging:    false,
	}
}

// Enable enables the inspector (F12 will work)
func (si *StandaloneInspector) Enable() {
	si.mu.Lock()
	defer si.mu.Unlock()

	si.enabled = true
	si.perf.Enable()

	if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
		fmt.Fprintf(os.Stderr, "[Inspector] Enabled (F12 to toggle)\n")
	}
}

// Disable disables the inspector
func (si *StandaloneInspector) Disable() {
	si.mu.Lock()
	defer si.mu.Unlock()

	si.enabled = false
	si.visible = false
	si.perf.Disable()

	if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
		fmt.Fprintf(os.Stderr, "[Inspector] Disabled\n")
	}
}

// IsEnabled returns whether inspector is enabled
func (si *StandaloneInspector) IsEnabled() bool {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.enabled
}

// ToggleVisibility toggles inspector overlay visibility
// Called by F12 key handler
func (si *StandaloneInspector) ToggleVisibility() {
	si.mu.Lock()
	defer si.mu.Unlock()

	if !si.enabled {
		return
	}

	si.visible = !si.visible

	if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
		state := "hidden"
		if si.visible {
			state = "visible"
		}
		fmt.Fprintf(os.Stderr, "[Inspector] Toggled: %s\n", state)
	}
}

// IsVisible returns whether inspector overlay is currently visible
func (si *StandaloneInspector) IsVisible() bool {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.visible
}

// AttachToApp attaches inspector to application root
// This should be called each frame with the current VNode tree
func (si *StandaloneInspector) AttachToApp(root rtui.VNode) {
	si.mu.Lock()
	defer si.mu.Unlock()

	si.appRoot = root

	// Update tree view
	si.treeView.SetRoot(root)

	// Run diagnostics if visible
	if si.visible {
		si.diagnostics.Analyze(root)
	}
}

// SelectElement selects an element by path
func (si *StandaloneInspector) SelectElement(path string) {
	si.mu.Lock()
	defer si.mu.Unlock()

	si.selectedPath = path
	if si.treeView != nil {
		node := si.treeView.FindNodeByPath(path)
		if node != nil {
			si.selectedVNode = node.VNode
		}
	}
}

// GetSelectedVNode returns the currently selected VNode
func (si *StandaloneInspector) GetSelectedVNode() rtui.VNode {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.selectedVNode
}

// SetActiveTab switches to a different inspector tab
func (si *StandaloneInspector) SetActiveTab(tab InspectorTab) {
	si.mu.Lock()
	defer si.mu.Unlock()
	si.activeTab = tab
}

// GetActiveTab returns the currently active tab
func (si *StandaloneInspector) GetActiveTab() InspectorTab {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.activeTab
}

// RenderOverlay renders the inspector overlay as a VNode
// This is used in framework-level overlay mode (LayerInspector)
func (si *StandaloneInspector) RenderOverlay() rtui.VNode {
	si.mu.RLock()
	defer si.mu.RUnlock()

	if !si.visible {
		return nil
	}

	// Build overlay content
	content := si.buildOverlayContent()

	// Mark as Inspector layer
	content.SetLayer(rtui.LayerInspector)

	return content
}

// buildOverlayContent builds the overlay UI using Tab and ScrollView components
// Returns a modern inspector panel with reusable components
func (si *StandaloneInspector) buildOverlayContent() rtui.VNode {
	// Create title bar
	titleBarParts := []rtui.VNode{
		app.NewTextBuilder("╔═ INSPECTOR ═╗").
			Style(style.NewStyle().Bold(true).Foreground(style.Blue).Background(style.Yellow).Reverse(true)).
			Build(),
		app.NewTextBuilder("F12:关闭 | Alt+H/J/K/L:移动 | Ctrl+D:按键调试").
			Style(style.NewStyle().Foreground(style.White).Background(style.Blue).Reverse(true)).
			Build(),
	}

	// Add key debug info (always visible when Inspector is shown)
	// Shows what key was last pressed
	modifiers := ""
	if si.lastAlt {
		modifiers += "Alt+"
	}
	if si.lastCtrl {
		modifiers += "Ctrl+"
	}
	if si.lastShift {
		modifiers += "Shift+"
	}
	if modifiers == "" {
		modifiers = "无"
	}

	keyInfo := fmt.Sprintf("🔍 Last key: '%s' (%s)", si.lastKey, modifiers)
	titleBarParts = append(titleBarParts,
		app.NewTextBuilder(keyInfo).
			Style(style.NewStyle().Foreground(style.Yellow).Background(style.Blue).Reverse(true)).
			Build(),
	)

	titleBar := rtui.VStack(titleBarParts...)

	// Create tabs using Tab component
	tabItems := []*navigation.TabItem{
		{ID: "elements", Label: "Elements", Content: si.buildElementsTabContent()},
		{ID: "console", Label: "Console", Content: si.buildConsoleTabContent()},
		{ID: "performance", Label: "Performance", Content: si.buildPerformanceTabContent()},
		{ID: "diagnostics", Label: "Diagnostics", Content: si.buildDiagnosticsTabContent()},
		{ID: "network", Label: "Network", Content: si.buildNetworkTabContent()},
	}

	// Build tabs with Tab component using Builder pattern
	tabsBuilder := navigation.TabsBuilder()
	for _, tab := range tabItems {
		tabsBuilder.AddTab(tab.ID, tab.Label)
		tabsBuilder.Content(tab.ID, tab.Content)
	}
	tabsBuilder.ActiveTab(int(si.activeTab))
	tabsComponent := tabsBuilder.Build()

	// Combine title bar, tabs, and separator
	content := rtui.VStack(
		titleBar,
		tabsComponent,
		ui.Text("─"), // Separator
	)

	// Wrap in bordered box
	panel := rtui.Bordered().
		Style(string(theme.Border())).
		Child(content).
		Width(si.overlayWidth).
		Height(si.overlayHeight).
		Build()

	// Set background on panel
	panel.SetStyle(style.NewStyle().Background(style.Blue))

	return panel
}

// buildOverlayContainer creates the overlay container
func (si *StandaloneInspector) buildOverlayContainer(content rtui.VNode) rtui.VNode {
	// Tab bar
	tabBar := si.buildTabBar()

	// Main content area
	mainContent := rtui.VStack(
		tabBar,
		ui.Text("─"), // Separator
		content,
	)

	// Wrap in bordered box with title
	overlay := rtui.Bordered().
		Style(string(theme.Border())).
		Child(mainContent).
		Width(si.overlayWidth).
		Height(si.overlayHeight).
		Build()

	return overlay
}

// buildTabBar creates the tab selection bar
func (si *StandaloneInspector) buildTabBar() rtui.VNode {
	var tabs []rtui.VNode

	allTabs := []InspectorTab{
		TabElements,
		TabConsole,
		TabPerformance,
		TabDiagnostics,
		TabNetwork,
	}

	for _, tab := range allTabs {
		tabName := tabNames[tab]

		// Highlight active tab
		if tab == si.activeTab {
			tabs = append(tabs,
				app.NewTextBuilder(fmt.Sprintf("[%s]", tabName)).
					Style(style.FgBgBold(style.Yellow, style.Blue)).
					Build(),
			)
		} else {
			tabs = append(tabs,
				app.NewTextBuilder(fmt.Sprintf(" %s ", tabName)).
					Style(style.Foreground(style.White)).
					Build(),
			)
		}
	}

	return rtui.HStack(tabs...)
}

// buildActiveTabContent builds content for the active tab
func (si *StandaloneInspector) buildActiveTabContent() rtui.VNode {
	switch si.activeTab {
	case TabElements:
		return si.buildElementsTabContent()
	case TabConsole:
		return si.buildConsoleTabContent()
	case TabPerformance:
		return si.buildPerformanceTabContent()
	case TabDiagnostics:
		return si.buildDiagnosticsTabContent()
	case TabNetwork:
		return si.buildNetworkTabContent()
	default:
		return ui.Text("Tab not implemented")
	}
}

// buildElementsTabContent builds content for Elements tab (used by Tab component)
func (si *StandaloneInspector) buildElementsTabContent() rtui.VNode {
	// Tree statistics
	stats := si.treeView.GetTreeStats()

	header := rtui.VStack(
		app.NewTextBuilder("📦 Layout Tree").
			Style(style.FgBold(style.Green)).
			Build(),
		app.NewTextBuilder(fmt.Sprintf("Nodes: %d | Depth: %d | Leaves: %d",
			stats.TotalNodes, stats.MaxDepth, stats.LeafNodes)).
			Style(style.Foreground(style.White)).
			Build(),
		app.NewTextBuilder("").
			Build(),
	)

	// Selected element info
	var selectedInfo rtui.VNode
	if si.selectedVNode != nil {
		props := si.selectedVNode.Props()
		infoText := fmt.Sprintf("Selected: %s", si.selectedVNode.Type())
		if props != nil {
			if key, ok := props["key"]; ok && key != "" {
				infoText += fmt.Sprintf(" (key: %v)", key)
			}
		}

		selectedInfo = rtui.VStack(
			app.NewTextBuilder("─").
				Style(style.Foreground(theme.Muted())).
				Build(),
			app.NewTextBuilder(infoText).
				Style(style.FgBold(style.Yellow)).
				Build(),
			app.NewTextBuilder(fmt.Sprintf("Path: %s", si.selectedPath)).
				Style(style.Foreground(style.Cyan)).
				Build(),
			app.NewTextBuilder("").
				Build(),
		)
	}

	// Tree visualization - use TreeView component with navigation
	allLines, totalLines := si.treeView.GetTreeLines()
	si.treeTotalLines = totalLines

	treeViewHeight := si.overlayHeight - 14 // Approximate available space for tree

	// Create or update TreeView component with navigation support
	if si.treeViewComponent == nil {
		si.treeViewComponent = display.NewTreeView().
			FromLines(allLines).
			ExpandLevel(1).
			ShowIcons(true).
			Compact(false).
			Build().(*display.TreeView)
	} else {
		// Update existing component with new lines
		updated := display.NewTreeView().
			FromLines(allLines).
			ExpandLevel(1).
			ShowIcons(true).
			Compact(false).
			Build().(*display.TreeView)

		// Preserve navigation state
		updated.SetFocusIndex(si.treeViewComponent.GetFocusIndex())
		updated.SetScrollOffset(si.treeScrollOffset)
		updated.SetViewportHeight(treeViewHeight)
		si.treeViewComponent = updated
	}

	// Sync scroll offset
	si.treeViewComponent.SetViewportHeight(treeViewHeight)
	si.treeViewComponent.SetScrollOffset(si.treeScrollOffset)

	// Get visible lines from TreeView component
	scrollOffset := si.treeViewComponent.GetScrollOffset()
	startLine := scrollOffset
	endLine := startLine + treeViewHeight
	if endLine > len(allLines) {
		endLine = len(allLines)
	}

	// Create Text nodes for visible lines
	var lineNodes []ui.VNode
	focusIndex := si.treeViewComponent.GetFocusIndex()
	selectedIdx := si.treeViewComponent.GetSelectedLine().NodeID

	for i := startLine; i < endLine; i++ {
		line := allLines[i]

		// Highlight selected line
		if i == selectedIdx {
			lineNodes = append(lineNodes, app.NewTextBuilder(line).
				Style(style.NewStyle().Reverse(true)).
				Build())
		} else if i == focusIndex {
			// Highlight focused line
			lineNodes = append(lineNodes, app.NewTextBuilder(line).
				Style(style.NewStyle().Bold(true).Foreground(style.Yellow)).
				Build())
		} else {
			lineNodes = append(lineNodes, ui.Text(line))
		}
	}

	// Display visible lines in VStack
	treePreview := ui.VStackBuilder(lineNodes...).
		Width(si.overlayWidth - 4).
		Build()

	// Instructions
	instructions := rtui.VStack(
		app.NewTextBuilder("─").
			Style(style.Foreground(theme.Muted())).
			Build(),
		app.NewTextBuilder("Instructions:").
			Style(style.FgBold(style.Cyan)).
			Build(),
		app.NewTextBuilder("  ↑↓: Navigate | Enter: Inspect").
			Style(style.Foreground(theme.Muted())).
			Build(),
		app.NewTextBuilder("  E: Expand/Collapse").
			Style(style.Foreground(theme.Muted())).
			Build(),
		app.NewTextBuilder("  PgUp/PgDn: Scroll tree").
			Style(style.Foreground(theme.Muted())).
			Build(),
		app.NewTextBuilder("  Home/End: Top/Bottom").
			Style(style.Foreground(theme.Muted())).
			Build(),
	)

	return rtui.VStack(
		header,
		selectedInfo,
		treePreview,
		instructions,
	)
}

// buildConsoleTabContent builds content for Console tab
func (si *StandaloneInspector) buildConsoleTabContent() rtui.VNode {
	return rtui.VStack(
		app.NewTextBuilder("💻 Console").
			Style(style.FgBold(style.Green)).
			Build(),
		app.NewTextBuilder("").
			Build(),
		app.NewTextBuilder("Console messages will appear here").
			Style(style.Foreground(theme.Muted())).
			Build(),
		app.NewTextBuilder("Use Inspector.Log() to output messages").
			Style(style.Foreground(theme.Muted())).
			Build(),
	)
}

// buildPerformanceTabContent builds content for Performance tab
func (si *StandaloneInspector) buildPerformanceTabContent() rtui.VNode {
	metrics := si.perf.GetMetrics()

	// Build metrics display
	metricsText := fmt.Sprintf(
		"Frame Count: %d\n"+
			"FPS: %.1f\n"+
			"Avg Render Time: %s\n"+
			"Memory: %s\n"+
			"GC Count: %d\n"+
			"Last GC: %s",
		metrics.FrameCount,
		metrics.FPS,
		metrics.AvgRenderTime,
		formatBytes(metrics.LastHeapAlloc),
		metrics.NumGC,
		metrics.LastGCTime,
	)

	return rtui.VStack(
		app.NewTextBuilder("⚡ Performance").
			Style(style.FgBold(style.Green)).
			Build(),
		app.NewTextBuilder("").
			Build(),
		app.NewTextBuilder(metricsText).
			Style(style.Foreground(style.White)).
			Build(),
		app.NewTextBuilder("").
			Build(),
		app.NewTextBuilder("─").
			Style(style.Foreground(theme.Muted())).
			Build(),
		app.NewTextBuilder("Real-time Performance Metrics").
			Style(style.FgBold(style.Cyan)).
			Build(),
		app.NewTextBuilder("Data updates every frame").
			Style(style.Foreground(theme.Muted())).
			Build(),
	)
}

// buildDiagnosticsTabContent builds content for Diagnostics tab
func (si *StandaloneInspector) buildDiagnosticsTabContent() rtui.VNode {
	problems := si.diagnostics.GetProblems()
	counts := si.diagnostics.CountBySeverity()

	// Build summary
	summaryText := fmt.Sprintf(
		"Total Problems: %d\n"+
			"Critical: %d\n"+
			"Errors: %d\n"+
			"Warnings: %d\n"+
			"Info: %d",
		len(problems),
		counts[SeverityCritical],
		counts[SeverityError],
		counts[SeverityWarning],
		counts[SeverityInfo],
	)

	return rtui.VStack(
		app.NewTextBuilder("🔧 Diagnostics").
			Style(style.FgBold(style.Green)).
			Build(),
		app.NewTextBuilder("").
			Build(),
		app.NewTextBuilder(summaryText).
			Style(style.Foreground(style.White)).
			Build(),
		app.NewTextBuilder("").
			Build(),
		app.NewTextBuilder("─").
			Style(style.Foreground(theme.Muted())).
			Build(),
		app.NewTextBuilder("Layout Problems").
			Style(style.FgBold(style.Cyan)).
			Build(),
		app.NewTextBuilder("No problems detected").
			Style(style.Foreground(style.Green)).
			Build(),
	)
}

// buildNetworkTabContent builds content for Network tab
func (si *StandaloneInspector) buildNetworkTabContent() rtui.VNode {
	return rtui.VStack(
		app.NewTextBuilder("🌐 Network").
			Style(style.FgBold(style.Green)).
			Build(),
		app.NewTextBuilder("").
			Build(),
		app.NewTextBuilder("Network activity will appear here").
			Style(style.Foreground(theme.Muted())).
			Build(),
		app.NewTextBuilder("HTTP requests, WebSocket messages").
			Style(style.Foreground(theme.Muted())).
			Build(),
	)
}

// buildConsoleTab builds the Console tab
func (si *StandaloneInspector) buildConsoleTab() rtui.VNode {
	return rtui.VStack(
		app.NewTextBuilder("💻 Console").
			Style(style.FgBold(style.Green)).
			Build(),
		app.NewTextBuilder("").
			Build(),
		app.NewTextBuilder("Console messages will appear here").
			Style(style.Foreground(theme.Muted())).
			Build(),
		app.NewTextBuilder("Use Inspector.Log() to output messages").
			Style(style.Foreground(theme.Muted())).
			Build(),
	)
}

// buildPerformanceTab builds the Performance tab
func (si *StandaloneInspector) buildPerformanceTab() rtui.VNode {
	metrics := si.perf.GetMetrics()

	// Build metrics display
	metricsText := fmt.Sprintf(
		"Frame Count: %d\n"+
		"FPS: %.1f\n"+
		"Avg Render Time: %s\n"+
		"Memory: %s\n"+
		"GC Count: %d\n"+
		"Last GC: %s",
		metrics.FrameCount,
		metrics.FPS,
		metrics.AvgRenderTime,
		formatBytes(metrics.LastHeapAlloc),
		metrics.NumGC,
		metrics.LastGCTime,
	)

	return rtui.VStack(
		app.NewTextBuilder("⚡ Performance").
			Style(style.FgBold(style.Green)).
			Build(),
		app.NewTextBuilder("").
			Build(),
		app.NewTextBuilder(metricsText).
			Style(style.Foreground(style.White)).
			Build(),
		app.NewTextBuilder("").
			Build(),
		app.NewTextBuilder("─").
			Style(style.Foreground(theme.Muted())).
			Build(),
		app.NewTextBuilder("Real-time Performance Metrics").
			Style(style.FgBold(style.Cyan)).
			Build(),
		app.NewTextBuilder("Data updates every frame").
			Style(style.Foreground(theme.Muted())).
			Build(),
	)
}

// buildDiagnosticsTab builds the Diagnostics tab
func (si *StandaloneInspector) buildDiagnosticsTab() rtui.VNode {
	problems := si.diagnostics.GetProblems()
	counts := si.diagnostics.CountBySeverity()

	// Build summary
	summaryText := fmt.Sprintf(
		"Total Problems: %d\n"+
		"Critical: %d\n"+
		"Errors: %d\n"+
		"Warnings: %d\n"+
		"Info: %d",
		len(problems),
		counts[SeverityCritical],
		counts[SeverityError],
		counts[SeverityWarning],
		counts[SeverityInfo],
	)

	var problemList rtui.VNode
	if len(problems) > 0 {
		var items []rtui.VNode
		for i, p := range problems {
			if i >= 10 { // Show first 10
				items = append(items,
					app.NewTextBuilder(fmt.Sprintf("... and %d more", len(problems)-10)).
						Style(style.Foreground(theme.Muted())).
						Build(),
				)
				break
			}

			severityStr := "INFO"
			severityStyle := style.Foreground
			switch p.Severity {
			case SeverityWarning:
				severityStr = "WARN"
				severityStyle = style.FgBold
			case SeverityError:
				severityStr = "ERR"
				severityStyle = style.FgBold
			case SeverityCritical:
				severityStr = "CRIT"
				severityStyle = style.FgBold
			}

			items = append(items,
				rtui.VStack(
					app.NewTextBuilder(fmt.Sprintf("[%s] %s", severityStr, p.Type)).
						Style(severityStyle(style.Yellow)).
						Build(),
					app.NewTextBuilder(fmt.Sprintf("  %s", p.Message)).
						Style(style.Foreground(style.White)).
						Build(),
					app.NewTextBuilder(fmt.Sprintf("  → %s", p.Suggestion)).
						Style(style.Foreground(style.Cyan)).
						Build(),
					app.NewTextBuilder("").
						Build(),
				),
			)
		}
		problemList = rtui.VStack(items...)
	} else {
		problemList = app.NewTextBuilder("✓ No layout problems detected").
			Style(style.FgBold(style.Green)).
			Build()
	}

	return rtui.VStack(
		app.NewTextBuilder("🔍 Diagnostics").
			Style(style.FgBold(style.Green)).
			Build(),
		app.NewTextBuilder("").
			Build(),
		app.NewTextBuilder(summaryText).
			Style(style.Foreground(style.White)).
			Build(),
		app.NewTextBuilder("").
			Build(),
		app.NewTextBuilder("─").
			Style(style.Foreground(theme.Muted())).
			Build(),
		app.NewTextBuilder("Layout Problems:").
			Style(style.FgBold(style.Cyan)).
			Build(),
		app.NewTextBuilder("").
			Build(),
		problemList,
	)
}

// buildNetworkTab builds the Network tab
func (si *StandaloneInspector) buildNetworkTab() rtui.VNode {
	return rtui.VStack(
		app.NewTextBuilder("🌐 Network").
			Style(style.FgBold(style.Green)).
			Build(),
		app.NewTextBuilder("").
			Build(),
		app.NewTextBuilder("Network activity monitoring").
			Style(style.Foreground(theme.Muted())).
			Build(),
		app.NewTextBuilder("(Not yet implemented for TUI)").
			Style(style.Foreground(theme.Muted())).
			Build(),
	)
}

// ExportReport generates and returns a report in the specified format
// Note: Standalone inspector uses simple text format for now
func (si *StandaloneInspector) ExportReport() string {
	si.mu.RLock()
	defer si.mu.RUnlock()

	metrics := si.perf.GetMetrics()
	problems := si.diagnostics.GetProblems()
	stats := si.treeView.GetTreeStats()

	report := fmt.Sprintf(
		"=== Standalone Inspector Report ===\n\n"+
		"Performance:\n"+
		"  FPS: %.1f\n"+
		"  Memory: %s\n"+
		"  Frames: %d\n"+
		"  GC Count: %d\n\n"+
		"Layout Tree:\n"+
		"  Nodes: %d\n"+
		"  Depth: %d\n"+
		"  Leaves: %d\n\n"+
		"Diagnostics:\n"+
		"  Problems: %d\n",
		metrics.FPS,
		formatBytes(metrics.LastHeapAlloc),
		metrics.FrameCount,
		metrics.NumGC,
		stats.TotalNodes,
		stats.MaxDepth,
		stats.LeafNodes,
		len(problems),
	)

	return report
}

// GetTreeView returns the tree view instance
func (si *StandaloneInspector) GetTreeView() *TreeView {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.treeView
}

// GetPerformance returns the performance analyzer
func (si *StandaloneInspector) GetPerformance() *PerformanceAnalyzer {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.perf
}

// GetDiagnostics returns the diagnostics instance
func (si *StandaloneInspector) GetDiagnostics() *LayoutDiagnostics {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.diagnostics
}

// GetPropertyEditor returns the property editor
func (si *StandaloneInspector) GetPropertyEditor() *PropertyEditor {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.editor
}

// GetTreeViewComponent returns the TreeView display component (for testing)
func (si *StandaloneInspector) GetTreeViewComponent() *display.TreeView {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.treeViewComponent
}

// =============================================================================
// Tree Scroll Control
// =============================================================================

// ScrollTreeBy scrolls the tree view by the given delta
func (si *StandaloneInspector) ScrollTreeBy(delta int) {
	si.mu.Lock()
	defer si.mu.Unlock()

	newOffset := si.treeScrollOffset + delta

	// Clamp to valid range
	// Available height for tree is approximately overlayHeight - 14
	treeViewHeight := si.overlayHeight - 14
	maxOffset := len(si.treeLines) - treeViewHeight

	if maxOffset < 0 {
		maxOffset = 0
	}

	// Clamp scroll offset
	if newOffset < 0 {
		newOffset = 0
	}
	if newOffset > maxOffset {
		newOffset = maxOffset
	}

	si.treeScrollOffset = newOffset
}

// ScrollTreeTo scrolls the tree view to an absolute position
func (si *StandaloneInspector) ScrollTreeTo(offset int) {
	si.mu.Lock()
	defer si.mu.Unlock()

	// Clamp to valid range
	treeViewHeight := si.overlayHeight - 14
	maxOffset := len(si.treeLines) - treeViewHeight

	if maxOffset < 0 {
		maxOffset = 0
	}

	// Clamp scroll offset
	if offset < 0 {
		offset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}

	si.treeScrollOffset = offset
}

// ScrollTreeTop scrolls to the top of the tree
func (si *StandaloneInspector) ScrollTreeTop() {
	si.mu.Lock()
	defer si.mu.Unlock()
	si.treeScrollOffset = 0
}

// ScrollTreeBottom scrolls to the bottom of the tree
func (si *StandaloneInspector) ScrollTreeBottom() {
	si.mu.Lock()
	defer si.mu.Unlock()

	treeViewHeight := si.overlayHeight - 14
	maxOffset := len(si.treeLines) - treeViewHeight

	if maxOffset < 0 {
		si.treeScrollOffset = 0
	} else {
		si.treeScrollOffset = maxOffset
	}
}

// PageUpTree scrolls up by one page
func (si *StandaloneInspector) PageUpTree() {
	treeViewHeight := si.overlayHeight - 14
	si.ScrollTreeBy(-treeViewHeight)
}

// PageDownTree scrolls down by one page
func (si *StandaloneInspector) PageDownTree() {
	treeViewHeight := si.overlayHeight - 14
	si.ScrollTreeBy(treeViewHeight)
}

// CanScrollTreeUp returns true if can scroll up
func (si *StandaloneInspector) CanScrollTreeUp() bool {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.treeScrollOffset > 0
}

// CanScrollTreeDown returns true if can scroll down
func (si *StandaloneInspector) CanScrollTreeDown() bool {
	si.mu.RLock()
	defer si.mu.RUnlock()

	treeViewHeight := si.overlayHeight - 14
	maxOffset := len(si.treeLines) - treeViewHeight

	return si.treeScrollOffset < maxOffset && maxOffset > 0
}

// GetTreeScrollPosition returns current scroll offset
func (si *StandaloneInspector) GetTreeScrollPosition() int {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.treeScrollOffset
}

// SetOverlaySize sets the overlay dimensions
func (si *StandaloneInspector) SetOverlaySize(width, height int) {
	si.mu.Lock()
	defer si.mu.Unlock()
	si.overlayWidth = width
	si.overlayHeight = height
}

// SetPosition sets the overlay position
func (si *StandaloneInspector) SetPosition(pos OverlayPosition) {
	si.mu.Lock()
	defer si.mu.Unlock()
	si.position = pos
}

// StartFrame marks the beginning of a frame (for performance tracking)
func (si *StandaloneInspector) StartFrame() {
	si.perf.StartFrame()
}

// EndFrame marks the end of a frame (for performance tracking)
func (si *StandaloneInspector) EndFrame() {
	si.perf.EndFrame()
}

// Log adds a message to the inspector console
func (si *StandaloneInspector) Log(message string) {
	// TODO: Implement console logging
	if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
		fmt.Fprintf(os.Stderr, "[Inspector] %s\n", message)
	}
}

// =============================================================================
// Position Management
// =============================================================================

// GetPosition returns the current floating position
func (si *StandaloneInspector) GetPosition() (x, y int) {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.floatX, si.floatY
}

// SetPosition sets the floating position
func (si *StandaloneInspector) SetFloatingPosition(x, y int) {
	si.mu.Lock()
	defer si.mu.Unlock()
	si.floatX = x
	si.floatY = y
	if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
		fmt.Fprintf(os.Stderr, "[Inspector] Position set to (%d, %d)\n", x, y)
	}
}

// Move moves the inspector by the specified offset
func (si *StandaloneInspector) Move(dx, dy int) {
	si.mu.Lock()
	defer si.mu.Unlock()
	si.floatX += dx
	si.floatY += dy
	// Ensure position is non-negative
	if si.floatX < 0 {
		si.floatX = 0
	}
	if si.floatY < 0 {
		si.floatY = 0
	}
	if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
		fmt.Fprintf(os.Stderr, "[Inspector] Moved by (%d, %d) to (%d, %d)\n", dx, dy, si.floatX, si.floatY)
	}
}

// =============================================================================
// Keyboard Event Handling
// =============================================================================

// HandleKeyEvent processes keyboard events for the inspector
// Returns true if the event was handled
func (si *StandaloneInspector) HandleKeyEvent(key string, alt bool, ctrl bool, shift bool) bool {
	si.mu.Lock()
	defer si.mu.Unlock()

	if !si.visible || !si.enabled {
		return false
	}

	// Store key info for debug display
	si.lastKey = key
	si.lastAlt = alt
	si.lastCtrl = ctrl
	si.lastShift = shift

	// Debug output for key detection
	if os.Getenv("TUI_DEBUG_UI") == "true" || os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
		modifiers := ""
		if alt { modifiers += "Alt+" }
		if ctrl { modifiers += "Ctrl+" }
		if shift { modifiers += "Shift+" }
		if modifiers == "" { modifiers = "none" }
		fmt.Fprintf(os.Stderr, "[Inspector] Key received: key='%s' modifiers=%s showKeyDebug=%v\n",
			key, modifiers, si.showKeyDebug)
	}

	// Debug mode: toggle with Ctrl+D (when Inspector is visible)
	if key == "d" && ctrl {
		si.showKeyDebug = !si.showKeyDebug
		fmt.Fprintf(os.Stderr, "[Inspector] showKeyDebug toggled to %v\n", si.showKeyDebug)
		return true
	}

	// Alt + Arrow keys to move the panel
	if alt {
		switch key {
		case "h", "left":
			si.floatX -= 2
			if si.floatX < 0 {
				si.floatX = 0
			}
			if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
				fmt.Fprintf(os.Stderr, "[Inspector] Moved left to x=%d\n", si.floatX)
			}
			return true
		case "l", "right":
			si.floatX += 2
			if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
				fmt.Fprintf(os.Stderr, "[Inspector] Moved right to x=%d\n", si.floatX)
			}
			return true
		case "k", "up":
			si.floatY -= 1
			if si.floatY < 0 {
				si.floatY = 0
			}
			if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
				fmt.Fprintf(os.Stderr, "[Inspector] Moved up to y=%d\n", si.floatY)
			}
			return true
		case "j", "down":
			si.floatY += 1
			if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
				fmt.Fprintf(os.Stderr, "[Inspector] Moved down to y=%d\n", si.floatY)
			}
			return true
		}
	}

	// Tab switching
	if key == "1" {
		si.activeTab = TabElements
		return true
	}
	if key == "2" {
		si.activeTab = TabConsole
		return true
	}
	if key == "3" {
		si.activeTab = TabPerformance
		return true
	}
	if key == "4" {
		si.activeTab = TabDiagnostics
		return true
	}
	if key == "5" {
		si.activeTab = TabNetwork
		return true
	}

	// Tab cycling - cycle through inspector tabs
	if key == "tab" {
		if shift {
			// Shift+Tab: cycle backward through tabs
			si.activeTab--
			if si.activeTab < TabElements {
				si.activeTab = TabNetwork
			}
		} else {
			// Tab (alone or with Ctrl/Alt): cycle forward through tabs
			si.activeTab++
			if si.activeTab > TabNetwork {
				si.activeTab = TabElements
			}
		}

		if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
			direction := "forward"
			if shift {
				direction = "backward"
			}
			fmt.Fprintf(os.Stderr, "[Inspector] Tab cycled %s to tab %d\n", direction, si.activeTab)
		}
		return true
	}

	// Tree scrolling and navigation - only when Elements tab is active
	if si.activeTab == TabElements {
		// If we have a TreeView component, delegate navigation to it
		if si.treeViewComponent != nil {
			treeViewHeight := si.overlayHeight - 14
			si.treeViewComponent.SetViewportHeight(treeViewHeight)

			// Map key strings to platform.SpecialKey and handle navigation
			var platformKey platform.SpecialKey
			var r rune
			handled := false

			switch key {
			case "up":
				platformKey = platform.KeyUp
				handled = si.treeViewComponent.HandleKey(platformKey, r)
			case "down":
				platformKey = platform.KeyDown
				handled = si.treeViewComponent.HandleKey(platformKey, r)
			case "pageup", "pgup": // Accept both for compatibility
				platformKey = platform.KeyPageUp
				handled = si.treeViewComponent.HandleKey(platformKey, r)
			case "pagedown", "pgdn": // Accept both for compatibility
				platformKey = platform.KeyPageDown
				handled = si.treeViewComponent.HandleKey(platformKey, r)
			case "home":
				platformKey = platform.KeyHome
				handled = si.treeViewComponent.HandleKey(platformKey, r)
			case "end":
				platformKey = platform.KeyEnd
				handled = si.treeViewComponent.HandleKey(platformKey, r)
			case "e":
				r = 'e'
				handled = si.treeViewComponent.HandleKey(platformKey, r)
			case "enter":
				platformKey = platform.KeyEnter
				handled = si.treeViewComponent.HandleKey(platformKey, r)
			}

			if handled {
				// Sync scroll offset back to Inspector
				si.treeScrollOffset = si.treeViewComponent.GetScrollOffset()
				if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
					focusedLine := si.treeViewComponent.GetFocusedLine()
					fmt.Fprintf(os.Stderr, "[Inspector] Tree navigation: focus=%d, scroll=%d, line=%q\n",
						si.treeViewComponent.GetFocusIndex(),
						si.treeScrollOffset,
						focusedLine.Content)
				}
				return true
			}
		}

		// Fallback to old scrolling behavior if TreeView not available
		treeViewHeight := si.overlayHeight - 14
		maxOffset := len(si.treeLines) - treeViewHeight
		if maxOffset < 0 {
			maxOffset = 0
		}

		switch key {
		case "pageup", "pgup": // Accept both for compatibility
			// Scroll up by one page
			si.treeScrollOffset -= treeViewHeight
			if si.treeScrollOffset < 0 {
				si.treeScrollOffset = 0
			}
			if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
				fmt.Fprintf(os.Stderr, "[Inspector] Tree scrolled up to offset %d\n", si.treeScrollOffset)
			}
			return true
		case "pagedown", "pgdn": // Accept both for compatibility
			// Scroll down by one page
			si.treeScrollOffset += treeViewHeight
			if si.treeScrollOffset > maxOffset {
				si.treeScrollOffset = maxOffset
			}
			if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
				fmt.Fprintf(os.Stderr, "[Inspector] Tree scrolled down to offset %d\n", si.treeScrollOffset)
			}
			return true
		case "home":
			// Scroll to top
			si.treeScrollOffset = 0
			if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
				fmt.Fprintf(os.Stderr, "[Inspector] Tree scrolled to top\n")
			}
			return true
		case "end":
			// Scroll to bottom
			si.treeScrollOffset = maxOffset
			if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
				fmt.Fprintf(os.Stderr, "[Inspector] Tree scrolled to bottom\n")
			}
			return true
		}
	}

	// When Inspector is visible, it's modal - capture ALL keyboard input
	// This prevents keys from falling through to the background app
	// (F12 and Ctrl+D are handled earlier by keyMap shortcuts)
	if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
		fmt.Fprintf(os.Stderr, "[Inspector] Visible mode: capturing key '%s' (alt=%v, ctrl=%v)\n",
			key, alt, ctrl)
	}
	return true
}
