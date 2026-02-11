// Package inspector provides standalone UI Inspector for TUI applications
//
// The standalone Inspector operates as an independent overlay interface,
// similar to browser DevTools, rather than being embedded in the application UI.
//
// Usage:
//
//	inspector := NewStandaloneInspector()
//	inspector.Enable()
//	inspector.AttachToApp(rootVNode)
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
	"strings"
	"sync"
	"time"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/components/container"
	"github.com/wwsheng009/mint/components/display"
	"github.com/wwsheng009/mint/components/navigation"
	"github.com/wwsheng009/mint/framework/action"
	frameworkevent "github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime"
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
	enabled   bool
	visible   bool
	activeTab InspectorTab

	// Data sources
	treeView          *TreeView
	treeViewComponent *display.TreeView // New: TreeView component with navigation
	perf              *PerformanceAnalyzer
	diagnostics       *LayoutDiagnostics
	editor            *PropertyEditor

	// VNode tracking
	appRoot       rtui.VNode
	selectedVNode rtui.VNode
	selectedPath  string
	hoveredVNode  rtui.VNode
	hoveredPath   string

	// Overlay configuration
	overlayWidth  int
	overlayHeight int
	position      OverlayPosition

	// Mouse tracking
	lastMouseX       int
	lastMouseY       int
	lastMouseButton  frameworkevent.MouseButton
	lastMouseEvent   frameworkevent.EventType
	lastMouseHandled bool

	// Floating position (for dragging)
	floatX      int  // X position when in floating mode
	floatY      int  // Y position when in floating mode
	isDragging  bool // Currently being dragged
	dragStartX  int  // Mouse X when drag started
	dragStartY  int  // Mouse Y when drag started
	floatStartX int  // Panel X when drag started
	floatStartY int  // Panel Y when drag started

	// Tree scroll state
	treeScrollOffset int      // Vertical scroll offset for tree view (in items)
	treeLines        []string // Cached tree lines for virtual scrolling
	treeTotalLines   int      // Total number of tree lines (for scroll bounds)

	// Cached overlay content for event dispatching
	cachedOverlayContent rtui.VNode // Cached overlay root for mouse events

	// Key debug info (for displaying what keys are being pressed)
	lastKey      string // Last key name received
	lastAlt      bool   // Last Alt modifier state
	lastCtrl     bool   // Last Ctrl modifier state
	lastShift    bool   // Last Shift modifier state
	showKeyDebug bool   // Show key debug info in UI

	// Update throttling
	lastUpdate     time.Time
	updateInterval time.Duration

	// VNode change tracking (to avoid unnecessary tree rebuilds)
	lastRootVNode       ui.VNode // Last root VNode that was attached
	lastTreeChangeCount int64    // Track tree view changes to avoid regenerating lines

	// Screen information
	screenWidth  int // Screen width (terminal columns)
	screenHeight int // Screen height (terminal rows)
}

// InspectorTab represents different inspector panels
type InspectorTab int

const (
	TabElements InspectorTab = iota
	TabConsole
	TabPerformance
	TabDiagnostics
	TabLayout
	TabNetwork
	TabScreenInfo
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
	TabElements:    "Elements",
	TabConsole:     "Console",
	TabPerformance: "Performance",
	TabDiagnostics: "Diagnostics",
	TabLayout:      "Layout",
	TabNetwork:     "Network",
	TabScreenInfo:  "Screen",
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
		// Floating position (left side, visible in 80-column terminal)
		floatX:         0, // Default X position (left edge)
		floatY:         0, // Default Y position (top edge)
		isDragging:     false,
		updateInterval: 200 * time.Millisecond, // Throttle updates (5 FPS for tree/diagnostics)
		screenWidth:    80, // Default terminal width
		screenHeight:   25, // Default terminal height
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

	// Throttle heavy updates to avoid excessive tree traversal
	if time.Since(si.lastUpdate) < si.updateInterval {
		return
	}
	si.lastUpdate = time.Now()

	// Check if VNode has actually changed (by pointer comparison)
	// This avoids expensive tree rebuilding when the same VNode is passed multiple times
	if si.lastRootVNode == root {
		if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
			fmt.Fprintf(os.Stderr, "[Inspector] AttachToApp: VNode unchanged, skipping SetRoot\n")
		}
		return
	}

	// Update tree view (expensive recursive build) - only if VNode changed
	si.treeView.SetRoot(root)
	si.lastRootVNode = root

	if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
		fmt.Fprintf(os.Stderr, "[Inspector] AttachToApp: VNode changed, SetRoot called\n")
	}

	// Run diagnostics if visible AND on diagnostics tab
	// Analyze is very expensive (full tree traversal + rule checking)
	if si.visible && si.activeTab == TabDiagnostics {
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
//
// Deprecated: Use RenderContent() instead. This method will be removed
// once the hook system is fully integrated.
func (si *StandaloneInspector) RenderOverlay() rtui.VNode {
	si.mu.Lock()
	defer si.mu.Unlock()

	if !si.visible {
		return nil
	}

	// Build overlay content
	content := si.buildOverlayContent()

	// Mark as Inspector layer
	content.SetLayer(rtui.LayerInspector)

	// Cache for event dispatching
	si.cachedOverlayContent = content

	return content
}

// RenderContent renders the inspector UI content without setting Layer
// This is used by the hook system. The hook will set LayerInspector
// and position the overlay, keeping this method purely focused on UI.
func (si *StandaloneInspector) RenderContent() rtui.VNode {
	si.mu.RLock()
	defer si.mu.RUnlock()

	if !si.visible {
		return nil
	}

	// Build overlay content (UI only, no Layer set)
	return si.buildOverlayContent()
}

// buildOverlayContent builds the overlay UI using Tab and ScrollView components
// Returns a modern inspector panel with reusable components
func (si *StandaloneInspector) buildOverlayContent() rtui.VNode {
	// Build combined header line: shortcuts + last key
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

	keyInfo := fmt.Sprintf("🔍 '%s' (%s)", si.lastKey, modifiers)
	mouseInfo := fmt.Sprintf("🖱 %d,%d %s", si.lastMouseX, si.lastMouseY, formatMouseButton(si.lastMouseEvent, si.lastMouseButton))
	hoverInfo := fmt.Sprintf("Hover: %s", si.formatHovered())
	headerText := fmt.Sprintf("F12:关闭 | Alt+J/K/L:移动 | Ctrl+D:调试 | %s | %s | %s", keyInfo, mouseInfo, hoverInfo)

	titleBar := app.NewTextBuilder(headerText).
		Style(style.NewStyle().Foreground(style.White).Background(style.Blue).Bold(true)).
		Build()

	// Create tabs using Tab component
	tabItems := []*navigation.TabItem{
		{ID: "elements", Label: "Elements(1)", Content: si.buildElementsTabContent()},
		{ID: "console", Label: "Console(2)", Content: si.buildConsoleTabContent()},
		{ID: "performance", Label: "Performance(3)", Content: si.buildPerformanceTabContent()},
		{ID: "diagnostics", Label: "Diagnostics(4)", Content: si.buildDiagnosticsTabContent()},
		{ID: "layout", Label: "Layout(5)", Content: si.buildLayoutTabContent()},
		{ID: "network", Label: "Network(6)", Content: si.buildNetworkTabContent()},
		{ID: "screen", Label: "Screen(7)", Content: si.buildScreenInfoTabContent()},
	}

	// Build tabs with Tab component using Builder pattern
	tabsBuilder := navigation.TabsBuilder()
	for _, tab := range tabItems {
		tabsBuilder.AddTab(tab.ID, tab.Label)
		tabsBuilder.Content(tab.ID, tab.Content)
	}
	tabsBuilder.ActiveTab(int(si.activeTab))
	tabsBuilder.OnChange(func(id string) {
		// Map id back to enum index
		for idx, item := range tabItems {
			if item.ID == id {
				si.activeTab = InspectorTab(idx)
				break
			}
		}
	})

	// No manual height calculation needed!
	// Panel + Flex layout handles this automatically.
	tabsComponent := tabsBuilder.Build()

	// Use the new Panel component
	panel := container.PanelBuilder().
		Header(titleBar).
		Content(tabsComponent).
		Footer(ui.Text("─")). // Bottom separator
		Width(si.overlayWidth).
		Height(si.overlayHeight).
		Build()

	// Set background on panel
	// Note: We need to cast to Stylable or set prop if Panel doesn't expose it directly via builder
	// Panel wraps BorderedNode, which supports SetStyle
	if stylable, ok := panel.(interface{ SetStyle(style.Style) }); ok {
		stylable.SetStyle(style.NewStyle().Background(style.Blue))
	}

	return panel
}

// formatMouseButton returns human-readable mouse state.
func formatMouseButton(eventType frameworkevent.EventType, btn frameworkevent.MouseButton) string {
	var btnStr string
	switch btn {
	case frameworkevent.MouseLeft:
		btnStr = "Left"
	case frameworkevent.MouseMiddle:
		btnStr = "Middle"
	case frameworkevent.MouseRight:
		btnStr = "Right"
	default:
		btnStr = "None"
	}

	var evt string
	switch eventType {
	case frameworkevent.EventMousePress:
		evt = "Press"
	case frameworkevent.EventMouseRelease:
		evt = "Release"
	case frameworkevent.EventMouseMove:
		evt = "Move"
	case frameworkevent.EventMouseWheel:
		evt = "Wheel"
	default:
		evt = "None"
	}

	return fmt.Sprintf("%s/%s", btnStr, evt)
}

// buildOverlayContainer creates the overlay container
func (si *StandaloneInspector) buildOverlayContainer(content rtui.VNode) rtui.VNode {
	// Tab bar
	tabBar := si.buildTabBar()

	// Main content area
	// The VStack has bounded height which allows flex children to properly size
	mainContent := ui.VStackBuilder(
		tabBar,
		ui.Text("─"), // Separator
		content,
	).
		Width(si.overlayWidth).
		Height(si.overlayHeight).
		Build()

	// Wrap in bordered box (bordered wrapper doesn't need height constraint
	// since mainContent already has it)
	overlay := rtui.Bordered().
		Style(string(theme.Border())).
		Child(mainContent).
		Width(si.overlayWidth).
		Build()

	return overlay
}

// buildTabBar creates the tab selection bar
func (si *StandaloneInspector) buildTabBar() rtui.VNode {
	var tabs []rtui.VNode

	allTabs := []struct {
		tab  InspectorTab
		key  string
		name string
	}{
		{TabElements, "1", "Elements"},
		{TabConsole, "2", "Console"},
		{TabPerformance, "3", "Performance"},
		{TabDiagnostics, "4", "Diagnostics"},
		{TabLayout, "5", "Layout"},
		{TabNetwork, "6", "Network"},
		{TabScreenInfo, "7", "Screen"},
	}

	for _, item := range allTabs {
		// Highlight active tab
		if item.tab == si.activeTab {
			tabs = append(tabs,
				app.NewTextBuilder(fmt.Sprintf("[%s]", item.name)).
					Style(style.FgBgBold(style.Yellow, style.Blue)).
					Build(),
			)
		} else {
			tabs = append(tabs,
				app.NewTextBuilder(fmt.Sprintf(" %s(%s) ", item.name, item.key)).
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
	case TabLayout:
		return si.buildLayoutTabContent()
	case TabNetwork:
		return si.buildNetworkTabContent()
	case TabScreenInfo:
		return si.buildScreenInfoTabContent()
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

	// Check for expand/collapse state change BEFORE building tree
	if si.treeViewComponent != nil && si.treeViewComponent.ExpandStateChanged() {
		lineIndex := si.treeViewComponent.GetExpandStateLineIndex()
		// Get uniqueID for this line index
		// Use cached lines for lookup
		if lineIndex >= 0 && lineIndex < len(si.treeLines) {
			uniqueID := si.treeView.GetUniqueIDForLineIndex(lineIndex)
			if uniqueID != "" {
				si.treeView.ToggleNode(uniqueID)
				if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
					fmt.Fprintf(os.Stderr, "[Inspector] Toggled node: %s (line %d)\n", uniqueID, lineIndex)
				}
			}
		}
		// Clear the flag
		si.treeViewComponent.ClearExpandStateChanged()
	}

	// Tree visualization - use TreeView component with navigation
	// Optimize: Only regenerate lines if tree structure or expansion state changed
	currentChangeCount := si.treeView.GetChangeCount()
	if currentChangeCount != si.lastTreeChangeCount || len(si.treeLines) == 0 {
		si.treeLines, si.treeTotalLines = si.treeView.GetTreeLines()
		si.lastTreeChangeCount = currentChangeCount

		if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
			fmt.Fprintf(os.Stderr, "[Inspector] Tree lines regenerated (count: %d)\n", len(si.treeLines))
			// Log first few lines to debug
			for i := 0; i < min(5, len(si.treeLines)); i++ {
				fmt.Fprintf(os.Stderr, "[Inspector] Tree line %d: %q\n", i, si.treeLines[i])
			}
		}
	}

	// Create or update TreeView component with navigation support
	if si.treeViewComponent == nil {
		si.treeViewComponent = display.NewTreeView().
			FromLines(si.treeLines).
			ExpandLevel(1).
			ShowIcons(true).
			Compact(false).
			Build().(*display.TreeView)
	} else {
		// Update existing TreeView with new lines WITHOUT creating a new instance
		// This preserves the viewportHeight that was set by the layout engine
		if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
			fmt.Fprintf(os.Stderr, "[Inspector] Updating TreeViewComponent with %d lines\n", len(si.treeLines))
		}
		si.treeViewComponent.UpdateLines(si.treeLines)
	}
	// Get focused line index to display above tree
	focusIndex := si.treeViewComponent.GetFocusIndex()

	// Build selected/focused element info
	// Priority: show selected (Enter) if available, otherwise show focused (↑↓)
	var targetVNode rtui.VNode
	var targetPath string
	var displayType string

	if si.selectedVNode != nil {
		// User pressed Enter to select
		targetVNode = si.selectedVNode
		targetPath = si.selectedPath
		displayType = "Selected"
	} else if focusIndex >= 0 {
		// Show focused item - use focusIndex to find node by index
		// The tree lines match 1-to-1 with the flattened tree nodes
		flatNodes := si.treeView.GetFlatList()
		if focusIndex < len(flatNodes) {
			node := flatNodes[focusIndex]
			targetVNode = node.VNode
			targetPath = node.Path
			displayType = "Focused"
		}
	}

	// Create info display
	var selectedInfo rtui.VNode
	if targetVNode != nil {
		props := targetVNode.Props()
		infoText := fmt.Sprintf("%s: %s", displayType, targetVNode.Type())
		if props != nil {
			if key, ok := props["key"]; ok && key != "" {
				infoText += fmt.Sprintf(" (key: %v)", key)
			}
		}

		// Create separator line that fills the width
		separator := strings.Repeat("─", si.overlayWidth-4)

		selectedInfo = rtui.VStack(
			app.NewTextBuilder(separator).
				Style(style.Foreground(theme.Muted())).
				Build(),
			app.NewTextBuilder(infoText).
				Style(style.FgBold(style.Yellow)).
				Build(),
			app.NewTextBuilder(fmt.Sprintf("Path: %s", targetPath)).
				Style(style.Foreground(style.Cyan)).
				Build(),
			app.NewTextBuilder("").
				Build(),
		)
	} else {
		// No selection or focus - show empty placeholder
		separator := strings.Repeat("─", si.overlayWidth-4)

		selectedInfo = rtui.VStack(
			app.NewTextBuilder(separator).
				Style(style.Foreground(theme.Muted())).
				Build(),
			app.NewTextBuilder("").
				Build(),
		)
	}

	// TreeView implements Measurable interface and will receive bounded constraints from parent
	// Set Flex prop directly on TreeView so it can grow to fill available space
	si.treeViewComponent.SetProp("flex", 1)
	si.treeViewComponent.SetProp("treeView", true) // hint for tests/diagnostics
	treePreview := si.treeViewComponent            // Directly use TreeView, no wrapper!

	// Instructions
	separator := strings.Repeat("─", si.overlayWidth-4)
	instructions := rtui.VStack(
		app.NewTextBuilder(separator).
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

	// Build the content VStack
	// Using Flex(1) makes the VStack expand to fill the Tabs content area
	// The Tabs component will pass the available height constraint to this VStack
	// The VStack will then distribute this height to its children (specifically TreeView with flex=1)
	return ui.VStackBuilder(
		header,
		selectedInfo,
		treePreview, // TreeView will receive bounded height constraints
		instructions,
	).
		Width(si.overlayWidth - 4).
		Height(si.overlayHeight - 4).
		Flex(1). // Expand to fill Tab content area
		Build()
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

// buildScreenInfoTabContent builds content for Screen Info tab
// Shows real-time screen size, mouse position, and button detection
func (si *StandaloneInspector) buildScreenInfoTabContent() rtui.VNode {
	// Build screen size info
	screenInfo := fmt.Sprintf("Terminal: %d cols × %d rows", si.screenWidth, si.screenHeight)

	// Build overlay info
	overlayInfo := fmt.Sprintf("Overlay: %d×%d at (%d,%d)", si.overlayWidth, si.overlayHeight, si.floatX, si.floatY)

	// Build mouse info
	mouseInfo := fmt.Sprintf("Mouse: (%d, %d)  %s  %s",
		si.lastMouseX, si.lastMouseY,
		formatMouseEventType(si.lastMouseEvent),
		formatMouseButtonName(si.lastMouseButton))

	// Check if mouse is in overlay
	minX, minY := si.floatX, si.floatY
	maxX, maxY := si.floatX+si.overlayWidth, si.floatY+si.overlayHeight
	mouseInOverlay := si.lastMouseX >= minX && si.lastMouseX < maxX && si.lastMouseY >= minY && si.lastMouseY < maxY

	// Build overlay bounds info
	boundsInfo := fmt.Sprintf("Bounds: (%d,%d) to (%d,%d)", minX, minY, maxX, maxY)
	insideInfo := fmt.Sprintf("Mouse in overlay: %s", formatBool(mouseInOverlay, "Yes", "No"))

	// Build element detection info
	var hoverInfo, selInfo string
	if si.hoveredVNode != nil {
		h := si.formatHovered()
		if len(h) > 50 {
			h = h[:47] + "..."
		}
		hoverInfo = fmt.Sprintf("Hovered: %s", h)
	} else {
		hoverInfo = "Hovered: None"
	}

	if si.selectedVNode != nil {
		selInfo = fmt.Sprintf("Selected: %s (%s)", si.selectedVNode.Type().String(), si.selectedPath)
		if len(selInfo) > 50 {
			selInfo = fmt.Sprintf("Selected: %s", si.selectedVNode.Type().String())
		}
	} else {
		selInfo = "Selected: None"
	}

	// Build button detection
	buttonInfo := si.buildButtonDetectionInfo()

	// Build inspector state
	stateInfo := fmt.Sprintf("Inspector: %s | %s | Tab: %s",
		formatBool(si.enabled, "Enabled", "Disabled"),
		formatBool(si.visible, "Visible", "Hidden"),
		tabNames[si.activeTab])

	// Build all lines compactly but readable
	return rtui.VStack(
		app.NewTextBuilder("📺 Screen Information").
			Style(style.FgBold(style.Green)).
			Build(),
		ui.Text("─"),

		// Screen & Overlay (2 lines)
		app.NewTextBuilder(screenInfo).
			Style(style.Foreground(style.White)).
			Build(),
		app.NewTextBuilder(overlayInfo).
			Style(style.Foreground(style.White)).
			Build(),

		ui.Text(""),

		// Mouse (2 lines)
		app.NewTextBuilder(mouseInfo).
			Style(style.Foreground(style.Cyan)).
			Build(),
		app.NewTextBuilder(boundsInfo + "  " + insideInfo).
			Style(func() style.Style {
				if mouseInOverlay {
					return style.Foreground(style.Green)
				}
				return style.Foreground(style.Red)
			}()).
			Build(),

		ui.Text(""),

		// Elements (2 lines)
		app.NewTextBuilder(hoverInfo).
			Style(style.Foreground(style.Yellow)).
			Build(),
		app.NewTextBuilder(selInfo).
			Style(style.Foreground(style.Green)).
			Build(),

		ui.Text(""),

		// Button & State (2 lines)
		app.NewTextBuilder(buttonInfo).
			Style(style.Foreground(style.White)).
			Build(),
		app.NewTextBuilder(stateInfo).
			Style(style.Foreground(style.Magenta)).
			Build(),

		ui.Text(""),

		// Instructions
		app.NewTextBuilder("Alt+H/J/K/L: Move overlay | Real-time updates").
			Style(style.Foreground(theme.Muted())).
			Build(),
	)
}

// buildButtonDetectionInfo returns info about buttons under mouse
func (si *StandaloneInspector) buildButtonDetectionInfo() string {
	if si.hoveredVNode == nil {
		return "No element under mouse"
	}

	// Check if hovered element is a button
	if tagger, ok := si.hoveredVNode.(interface{ Tag() string }); ok {
		tag := tagger.Tag()
		if tag == "button" {
			return fmt.Sprintf("✓ Button detected: %s", tag)
		}
	}

	// Check for onClick handler
	props := si.hoveredVNode.Props()
	if props != nil {
		if _, hasOnClick := props["onClick"]; hasOnClick {
			return "✓ Clickable element (has onClick)"
		}
	}

	typeName := si.hoveredVNode.Type().String()
	return fmt.Sprintf("Element: %s (not a button)", typeName)
}

// formatBool returns a formatted string based on boolean value
func formatBool(cond bool, trueStr, falseStr string) string {
	if cond {
		return trueStr
	}
	return falseStr
}

// formatMouseEventType returns human-readable mouse event type
func formatMouseEventType(eventType frameworkevent.EventType) string {
	switch eventType {
	case frameworkevent.EventMousePress:
		return "Press"
	case frameworkevent.EventMouseRelease:
		return "Release"
	case frameworkevent.EventMouseMove:
		return "Move"
	case frameworkevent.EventMouseWheel:
		return "Wheel"
	default:
		return "None"
	}
}

// formatMouseButtonName returns human-readable mouse button name
func formatMouseButtonName(btn frameworkevent.MouseButton) string {
	switch btn {
	case frameworkevent.MouseLeft:
		return "Left"
	case frameworkevent.MouseMiddle:
		return "Middle"
	case frameworkevent.MouseRight:
		return "Right"
	default:
		return "None"
	}
}

// buildLayoutTabContent builds content for Layout Diagnostics tab
func (si *StandaloneInspector) buildLayoutTabContent() rtui.VNode {
	// Create diagnostic instance
	diagnostic := NewLayoutDiagnostic()

	// Analyze the selected node if available, otherwise analyze the entire app root
	var targetVNode rtui.VNode
	var displayText string

	if si.selectedVNode != nil {
		targetVNode = si.selectedVNode
		displayText = fmt.Sprintf("Selected: %s", si.selectedVNode.Type().String())
	} else if si.appRoot != nil {
		targetVNode = si.appRoot
		displayText = "Analyzing entire app root"
	} else {
		return rtui.VStack(
			app.NewTextBuilder("📐 Layout Diagnostics").
				Style(style.FgBold(style.Green)).
				Build(),
			app.NewTextBuilder("").
				Build(),
			app.NewTextBuilder("No VNode to analyze").
				Style(style.Foreground(theme.Muted())).
				Build(),
			app.NewTextBuilder("Select a node in Elements tab first").
				Style(style.Foreground(theme.Muted())).
				Build(),
		)
	}

	// Analyze the target node
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  si.overlayWidth - 4,
		MinHeight: 0,
		MaxHeight: si.overlayHeight - 10,
	}

	result := diagnostic.AnalyzeSelectedNode(targetVNode, constraints)
	formattedResult := diagnostic.FormatSingleResult(result)

	// Split into lines for display
	lines := strings.Split(formattedResult, "\n")

	// Limit lines to fit in overlay
	maxLines := si.overlayHeight - 8
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, "... (truncated)")
	}

	// Convert lines to VNodes
	contentNodes := make([]rtui.VNode, 0, len(lines))
	for _, line := range lines {
		// Skip empty lines for cleaner display
		if line == "" {
			continue
		}

		// Colorize based on content
		var text rtui.VNode
		if strings.Contains(line, "✅") {
			text = app.NewTextBuilder(line).
				Style(style.FgBold(style.Green)).
				Build()
		} else if strings.Contains(line, "❌") {
			text = app.NewTextBuilder(line).
				Style(style.FgBold(style.Red)).
				Build()
		} else if strings.Contains(line, "⚠️") {
			text = app.NewTextBuilder(line).
				Style(style.FgBold(style.Yellow)).
				Build()
		} else if strings.Contains(line, "SELECTED NODE") || strings.Contains(line, "═") {
			text = app.NewTextBuilder(line).
				Style(style.FgBold(style.Cyan)).
				Build()
		} else if strings.Contains(line, "Constraints:") || strings.Contains(line, "Measured:") {
			text = app.NewTextBuilder(line).
				Style(style.Foreground(style.White)).
				Build()
		} else if strings.Contains(line, "Issues:") {
			text = app.NewTextBuilder(line).
				Style(style.FgBold(style.Yellow)).
				Build()
		} else {
			text = app.NewTextBuilder(line).
				Style(style.Foreground(style.White)).
				Build()
		}
		contentNodes = append(contentNodes, text)
	}

	// Build the VStack with header and content
	headerNodes := []rtui.VNode{
		app.NewTextBuilder("📐 Layout Diagnostics").
			Style(style.FgBold(style.Green)).
			Build(),
		app.NewTextBuilder(displayText).
			Style(style.Foreground(style.Cyan)).
			Build(),
		app.NewTextBuilder("").
			Build(),
		app.NewTextBuilder("─").
			Style(style.Foreground(theme.Muted())).
			Build(),
	}

	// Combine header and content
	allNodes := append(headerNodes, contentNodes...)

	return rtui.VStack(allNodes...)
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

// SetScreenSize sets the screen (terminal) size
func (si *StandaloneInspector) SetScreenSize(width, height int) {
	si.mu.Lock()
	defer si.mu.Unlock()
	si.screenWidth = width
	si.screenHeight = height
}

// GetScreenSize returns the screen (terminal) size
func (si *StandaloneInspector) GetScreenSize() (width, height int) {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.screenWidth, si.screenHeight
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
		if alt {
			modifiers += "Alt+"
		}
		if ctrl {
			modifiers += "Ctrl+"
		}
		if shift {
			modifiers += "Shift+"
		}
		if modifiers == "" {
			modifiers = "none"
		}
		fmt.Fprintf(os.Stderr, "[Inspector] Key received: key='%s' modifiers=%s showKeyDebug=%v\n",
			key, modifiers, si.showKeyDebug)
	}

	// Debug mode: toggle with Ctrl+D (when Inspector is visible)
	if key == "d" && ctrl {
		si.showKeyDebug = !si.showKeyDebug
		fmt.Fprintf(os.Stderr, "[Inspector] showKeyDebug toggled to %v\n", si.showKeyDebug)
		return true
	}

	// Layout Dump with Ctrl+L
	if key == "l" && ctrl {
		if si.appRoot != nil {
			analyzer := NewLayoutAnalyzer()
			snapshot := analyzer.Capture(si.appRoot, 0)
			treeStr := analyzer.FormatTree(snapshot)
			_ = os.WriteFile("layout_dump.txt", []byte(treeStr), 0644)
			fmt.Fprintf(os.Stderr, "[Inspector] Layout dump saved to layout_dump.txt\n")
		}
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
		if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
			fmt.Fprintf(os.Stderr, "[Inspector] Switched to Elements tab (key=1)\n")
		}
		return true
	}
	if key == "2" {
		si.activeTab = TabConsole
		if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
			fmt.Fprintf(os.Stderr, "[Inspector] Switched to Console tab (key=2)\n")
		}
		return true
	}
	if key == "3" {
		si.activeTab = TabPerformance
		if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
			fmt.Fprintf(os.Stderr, "[Inspector] Switched to Performance tab (key=3)\n")
		}
		return true
	}
	if key == "4" {
		si.activeTab = TabDiagnostics
		if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
			fmt.Fprintf(os.Stderr, "[Inspector] Switched to Diagnostics tab (key=4)\n")
		}
		return true
	}
	if key == "5" {
		si.activeTab = TabLayout
		if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
			fmt.Fprintf(os.Stderr, "[Inspector] Switched to Layout tab (key=5)\n")
		}
		return true
	}
	if key == "6" {
		si.activeTab = TabNetwork
		if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
			fmt.Fprintf(os.Stderr, "[Inspector] Switched to Network tab (key=6)\n")
		}
		return true
	}
	if key == "7" {
		si.activeTab = TabScreenInfo
		if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
			fmt.Fprintf(os.Stderr, "[Inspector] Switched to Screen Info tab (key=7)\n")
		}
		return true
	}

	// Tab cycling - cycle through inspector tabs
	if key == "tab" {
		if shift {
			// Shift+Tab: cycle backward through tabs
			si.activeTab--
			if si.activeTab < TabElements {
				si.activeTab = TabScreenInfo
			}
		} else {
			// Tab (alone or with Ctrl/Alt): cycle forward through tabs
			si.activeTab++
			if si.activeTab > TabScreenInfo {
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
			// Calculate approximate tree view height for paging
			// Overhead: Border(2) + Title(3) + Sep(1) + TabBar(1) + Header(3) + SelectedInfo(4) + Instructions(6) = 20
			treeViewHeight := si.overlayHeight - 20
			if treeViewHeight < 1 {
				treeViewHeight = 1
			}
			// Note: We don't strictly need to set this here as Measure() will update it,
			// but setting it helps with accurate Paging calculations before next render
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

				// Update selected node info when Enter is pressed
				if handled {
					selectedLine := si.treeViewComponent.GetSelectedLine()
					if selectedLine.NodeID >= 0 && selectedLine.Path != "" {
						// Find the node by path and update selection
						node := si.treeView.FindNodeByPath(selectedLine.Path)
						if node != nil {
							si.selectedVNode = node.VNode
							si.selectedPath = selectedLine.Path
							if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
								fmt.Fprintf(os.Stderr, "[Inspector] Selected node: %s (path: %s)\n",
									node.Info.Type, selectedLine.Path)
							}
						}
					}
				}
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
		// Overhead: Border(2) + Title(3) + Sep(1) + TabBar(1) + Header(3) + SelectedInfo(4) + Instructions(6) = 20
		treeViewHeight := si.overlayHeight - 20
		if treeViewHeight < 1 {
			treeViewHeight = 1
		}
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

// HandleMouseEvent processes mouse events for the inspector (position & button tracking).
// Returns true if the event should be captured by the inspector.
func (si *StandaloneInspector) HandleMouseEvent(eventType frameworkevent.EventType, ev *frameworkevent.MouseEvent) bool {
	si.mu.Lock()
	defer si.mu.Unlock()

	if !si.enabled {
		return false
	}

	// Track latest position and button for display
	si.lastMouseX = ev.X
	si.lastMouseY = ev.Y
	si.lastMouseButton = ev.Button
	si.lastMouseEvent = eventType

	// Debug logging for all mouse events
	if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
		fmt.Fprintf(os.Stderr, "[Inspector] HandleMouseEvent: type=%v, pos=(%d,%d), button=%d, visible=%v\n",
			eventType, ev.X, ev.Y, ev.Button, si.visible)
	}

	// Update hovered VNode by hit testing current app root
	if si.appRoot != nil {
		if node := findVNodeAtRecursive(si.appRoot, ev.X, ev.Y, 0); node != nil {
			si.hoveredVNode = node
			si.hoveredPath = si.lookupPathForVNode(node)
		} else {
			si.hoveredVNode = nil
			si.hoveredPath = ""
		}
	}

	// If inspector is visible and mouse is over its floating area, capture click to prevent app interference.
	if si.visible {
		// Rough bounding box based on overlay size and floating position.
		minX, minY := si.floatX, si.floatY
		maxX := si.floatX + si.overlayWidth
		maxY := si.floatY + si.overlayHeight

		if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
			fmt.Fprintf(os.Stderr, "[Inspector] Overlay bounds: minX=%d, minY=%d, maxX=%d, maxY=%d\n",
				minX, minY, maxX, maxY)
			fmt.Fprintf(os.Stderr, "[Inspector] Mouse pos: (%d,%d), in overlay: %v\n",
				ev.X, ev.Y, ev.X >= minX && ev.X < maxX && ev.Y >= minY && ev.Y < maxY)
		}

		if ev.X >= minX && ev.X < maxX && ev.Y >= minY && ev.Y < maxY {
			// Convert to overlay coordinates
			localX := ev.X - minX
			localY := ev.Y - minY

			// Try to deliver event to components in the overlay
			// Create MouseEvent with local coordinates for the overlay
			localEv := &frameworkevent.MouseEvent{
				BaseEvent: frameworkevent.NewBaseEvent(eventType),
				X:         localX, // Overlay-local coordinates
				Y:         localY,
				LocalX:    localX, // Same as X for overlay-root coordinates
				LocalY:    localY,
				Button:    ev.Button,
			}

			// Try to deliver to cached overlay content
			if si.cachedOverlayContent != nil {
				// The overlay content is a Panel that contains Tabs
				// Try to deliver the event to it
				if component, ok := si.cachedOverlayContent.(frameworkevent.Component); ok {
					handled := component.HandleEvent(localEv)
					if handled {
						if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
							fmt.Fprintf(os.Stderr, "[Inspector] Event handled by overlay component\n")
						}
						return true
					}
				}
			}

			// Fallback: manual handling for tab bar
			handled := si.handleOverlayMouse(localX, localY, eventType, ev.Button)
			if handled && os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
				fmt.Fprintf(os.Stderr, "[Inspector] Event handled by manual fallback\n")
			}
			return handled
		}
	}

	// Do not block application mouse handling.
	return false
}

// handleOverlayMouse processes clicks within the overlay (e.g., tabs, treeview).
func (si *StandaloneInspector) handleOverlayMouse(localX, localY int, eventType frameworkevent.EventType, btn frameworkevent.MouseButton) bool {
	// Only handle left button presses
	if eventType != frameworkevent.EventMousePress || btn != frameworkevent.MouseLeft {
		return false
	}

	// Tab bar location calculation:
	// Panel structure:
	//   Row 0: Border top line
	//   Row 1: Header (titleBar) - may wrap to multiple lines
	//   Row 1+headerLines: Separator line
	//   Row 2+headerLines: Tab bar (first line of Tabs component)
	//   Row 3+headerLines: Tab content
	//
	// For a simple title bar that doesn't wrap, tab bar is at row 2
	// But to be safe, let's check rows 1-3 for the tab bar pattern

	// Check if we're in a potential tab bar row (rows 1-3)
	if localY >= 1 && localY <= 3 {
		// Build the tab labels
		labels := []string{
			tabNames[TabElements] + "(1)",
			tabNames[TabConsole] + "(2)",
			tabNames[TabPerformance] + "(3)",
			tabNames[TabDiagnostics] + "(4)",
			tabNames[TabLayout] + "(5)",
			tabNames[TabNetwork] + "(6)",
		}

		// Build the same string as the Tabs component would render:
		// Active tab: "[Label]"  Inactive: " Label "  Separator: " | "
		cursor := 0
		for idx, label := range labels {
			var width int
			if InspectorTab(idx) == si.activeTab {
				// Active tab: [label]
				width = len(label) + 2 // [ ]
				if localX >= cursor && localX < cursor+width {
					return false // clicking active tab does nothing, consume event
				}
				cursor += width
			} else {
				// Inactive tab: " label " (with spaces)
				width = len(label) + 2 // leading/trailing spaces
				if localX >= cursor && localX < cursor+width {
					si.activeTab = InspectorTab(idx)
					if os.Getenv("TUI_INSPECTOR_VERBOSE") == "true" {
						fmt.Fprintf(os.Stderr, "[Inspector] Tab clicked: %s (row %d, col %d)\n", label, localY, localX)
					}
					return true // tab switched, consume event
				}
				cursor += width
			}
			// separator " | " except after last tab
			if idx < len(labels)-1 {
				cursor += 3
			}
		}
	}

	// Try to handle click through handleOverlayClick (includes TreeView, etc.)
	if si.handleOverlayClick(localX, localY) {
		return true
	}

	return false // not handled, let event pass through
}

// formatHovered returns a human-readable name of the currently hovered control.
func (si *StandaloneInspector) formatHovered() string {
	if si.hoveredVNode == nil {
		return "无"
	}

	var name string
	if tagger, ok := si.hoveredVNode.(interface{ Tag() string }); ok {
		name = tagger.Tag()
	} else {
		name = fmt.Sprintf("%T", si.hoveredVNode)
	}

	if si.hoveredPath != "" {
		return fmt.Sprintf("%s (%s)", name, si.hoveredPath)
	}
	return name
}

// lookupPathForVNode tries to find the path of a VNode in the current tree.
func (si *StandaloneInspector) lookupPathForVNode(vnode rtui.VNode) string {
	if si.treeView == nil {
		return ""
	}
	flat := si.treeView.GetFlatList()
	for _, n := range flat {
		if n.VNode == vnode {
			return n.Path
		}
	}
	return ""
}

// ============================================================================
// CaptureActionHandler Implementation - Phase 3-5
// ============================================================================

// HandleCapture implements action.CaptureActionHandler interface.
// The Inspector captures overlay-related actions before they reach the application.
//
// Priority: High (100) - Inspector should intercept overlay actions before application
func (si *StandaloneInspector) HandleCapture(act *action.Action, target *runtime.LayoutNode) bool {
	si.mu.Lock()
	defer si.mu.Unlock()

	// Only capture when inspector is enabled and visible
	if !si.enabled || !si.visible {
		return false
	}

	// Capture mouse-related actions
	switch act.Type {
	case action.ActionClick, action.ActionHover:
		return si.handleCaptureMouseAction(act)

	case action.ActionInspect:
		// Toggle inspector visibility
		si.visible = !si.visible
		return true // Stop propagation

	default:
		// Don't capture other actions
		return false
	}
}

// Priority returns the priority for the capture handler.
// Inspector uses high priority to ensure it can intercept overlay events.
func (si *StandaloneInspector) Priority() int {
	return 100 // High priority - Inspector overlay should be handled first
}

// handleCaptureMouseAction handles mouse-related actions in the capture phase.
// Returns true if the action was captured and should stop propagation.
func (si *StandaloneInspector) handleCaptureMouseAction(act *action.Action) bool {
	// Extract mouse coordinates from payload
	x, y, ok := act.GetPayloadPoint()
	if !ok {
		return false
	}

	// Check if mouse is over the inspector overlay
	minX, minY := si.floatX, si.floatY
	maxX := si.floatX + si.overlayWidth
	maxY := si.floatY + si.overlayHeight

	if x >= minX && x < maxX && y >= minY && y < maxY {
		// Mouse is over inspector overlay
		localX := x - minX
		localY := y - minY

		// Handle the action within the overlay
		if act.Type == action.ActionClick {
			return si.handleOverlayClick(localX, localY)
		}
		// For hover, we track but don't stop propagation
		// so the app can also see hover events
	}

	return false
}

// handleOverlayClick handles click actions within the overlay.
// Returns true if the click was handled by the overlay.
func (si *StandaloneInspector) handleOverlayClick(localX, localY int) bool {
	// Check if click is in the tab bar area
	// Panel structure:
	//   Row 0: Border top line
	//   Row 1: Header (titleBar)
	//   Row 2: Separator
	//   Row 3: Tab bar
	//   Row 4+: Tab content

	// Calculate header height
	headerHeight := 2 // Title bar + separator

	// Tab bar is at row 3 (after header)
	tabBarY := headerHeight
	tabBarHeight := 1

	// Check if click is in tab bar
	if localY >= tabBarY && localY < tabBarY+tabBarHeight {
		return si.handleTabBarClick(localX)
	}

	// TreeView clicks are now handled by the component system:
	// Panel → Tabs → ActiveTabContent → TreeView
	// Each component manages its own hit testing with proper bounds from the layout engine.
	return false
}

// handleTabBarClick handles clicks on the tab bar.
// Returns true if the click was handled.
func (si *StandaloneInspector) handleTabBarClick(localX int) bool {
	// Tab layout: [Elements] [Layout] [Network] [Perf] [Diag] [Console]
	// Each tab is roughly 8 characters wide with spacing

	tabs := []InspectorTab{
		TabElements,
		TabLayout,
		TabNetwork,
		TabPerformance,
		TabDiagnostics,
		TabConsole,
	}

	tabWidth := 10 // Approximate width of each tab including spacing
	xOffset := 2   // Left padding

	for i, tab := range tabs {
		tabStart := xOffset + i*tabWidth
		tabEnd := tabStart + tabWidth - 2

		if localX >= tabStart && localX < tabEnd {
			// Clicked on this tab
			si.activeTab = tab
			return true // Handled
		}
	}

	return false
}
