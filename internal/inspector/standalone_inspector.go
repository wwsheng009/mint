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

	frameworkevent "github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	componentlist "github.com/wwsheng009/mint/ui/components/list"
	componentpanel "github.com/wwsheng009/mint/ui/components/panel"
	componenttabs "github.com/wwsheng009/mint/ui/components/tabs"
	componenttreeview "github.com/wwsheng009/mint/ui/components/treeview"
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
	treeViewComponent *componenttreeview.VNode // TreeView component with navigation
	perf              *PerformanceAnalyzer
	diagnostics       *LayoutDiagnostics
	editor            *PropertyEditor

	// VNode tracking
	appRoot       ui.VNode
	selectedVNode ui.VNode
	selectedPath  string
	hoveredVNode  ui.VNode
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
	cachedOverlayContent ui.VNode // Cached overlay root for mouse events

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

	// HitTest data
	hitMapEntries []HitTestEntry // Cached hit test entries for display

	// Intent handlers initialization
	intentHandlersInitialized bool // Whether intent handlers have been registered
}

// HitTestEntry represents a single entry in the hit test display
type HitTestEntry struct {
	Type      string
	Bounds    string // Formatted bounds string
	ZOrder    int
	HitTest   string // Hit test result at current mouse position
	Clickable bool
	// Debug fields - additional info to distinguish widgets
	Tag   string // Component tag (button, text, box, etc.)
	Key   string // Component key (if available)
	Label string // Component label (text content, etc.)
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
	TabHitTest
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
	TabHitTest:     "HitTest",
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
		screenWidth:    80,                     // Default terminal width
		screenHeight:   25,                     // Default terminal height
	}
}

// Enable enables the inspector (F12 will work)
func (si *StandaloneInspector) Enable() {
	si.mu.Lock()
	defer si.mu.Unlock()

	si.enabled = true
	si.perf.Enable()

	// Initialize intent handlers (only once, only if app is initialized)
	if !si.intentHandlersInitialized {
		// Try to initialize intent handlers, but don't fail if app is not ready (e.g., in tests)
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Intent registration failed (e.g., app not initialized in tests)
					// This is OK - handlers will be registered during normal app startup
					log.InspectorLogger.Debug("Intent handler registration deferred: %v", r)
				}
			}()
			si.initIntentHandlers()
			si.intentHandlersInitialized = true
		}()
	}

	log.InspectorLogger.Debug("Enabled (F12 to toggle)")
}

// Disable disables the inspector
func (si *StandaloneInspector) Disable() {
	si.mu.Lock()
	defer si.mu.Unlock()

	si.enabled = false
	si.visible = false
	si.perf.Disable()

	log.InspectorLogger.Debug("Disabled")
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

	state := "hidden"
	if si.visible {
		state = "visible"
	}
	log.InspectorLogger.Debug("Toggled: %s", state)
}

// IsVisible returns whether inspector overlay is currently visible
func (si *StandaloneInspector) IsVisible() bool {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.visible
}

// AttachToApp attaches inspector to application root
// This should be called each frame with the current VNode tree
func (si *StandaloneInspector) AttachToApp(root ui.VNode) {
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
		log.InspectorLogger.Debug("AttachToApp: VNode unchanged, skipping SetRoot")
		return
	}

	// Update tree view (expensive recursive build) - only if VNode changed
	si.treeView.SetRoot(root)
	si.lastRootVNode = root

	log.InspectorLogger.Debug("AttachToApp: VNode changed, SetRoot called")

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
func (si *StandaloneInspector) GetSelectedVNode() ui.VNode {
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
func (si *StandaloneInspector) RenderOverlay() ui.VNode {
	si.mu.Lock()
	defer si.mu.Unlock()

	if !si.visible {
		return nil
	}

	// Build overlay content
	content := si.buildOverlayContent()

	// Mark as Inspector layer
	content.SetLayer(ui.LayerInspector)

	// Cache for event dispatching
	si.cachedOverlayContent = content

	return content
}

// RenderContent renders the inspector UI content without setting Layer
// This is used by the hook system. The hook will set LayerInspector
// and position the overlay, keeping this method purely focused on UI.
func (si *StandaloneInspector) RenderContent() ui.VNode {
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
func (si *StandaloneInspector) buildOverlayContent() ui.VNode {
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

	titleBar := ui.NewTextBuilder(headerText).
		Style(style.NewStyle().Foreground(style.White).Background(style.Blue).Bold(true)).
		Build()

	// Create tabs using Tab component
	tabItems := []componenttabs.TabItem{
		{ID: "elements", Label: "Elements(1)", Disabled: false},
		{ID: "console", Label: "Console(2)", Disabled: false},
		{ID: "performance", Label: "Performance(3)", Disabled: false},
		{ID: "diagnostics", Label: "Diagnostics(4)", Disabled: false},
		{ID: "layout", Label: "Layout(5)", Disabled: false},
		{ID: "network", Label: "Network(6)", Disabled: false},
		{ID: "screen", Label: "Screen(7)", Disabled: false},
		{ID: "hittest", Label: "HitTest(8)", Disabled: false},
	}

	// Build active tab content
	var activeContent ui.VNode
	switch si.activeTab {
	case TabElements:
		activeContent = si.buildElementsTabContent()
	case TabConsole:
		activeContent = si.buildConsoleTabContent()
	case TabPerformance:
		activeContent = si.buildPerformanceTabContent()
	case TabDiagnostics:
		activeContent = si.buildDiagnosticsTabContent()
	case TabLayout:
		activeContent = si.buildLayoutTabContent()
	case TabNetwork:
		activeContent = si.buildNetworkTabContent()
	case TabScreenInfo:
		activeContent = si.buildScreenInfoTabContent()
	case TabHitTest:
		activeContent = si.buildHitTestTabContent()
	default:
		activeContent = ui.Text("Tab not implemented")
	}

	// Build tabs component using Builder pattern
	tabsComponent := componenttabs.NewBuilder().
		Tabs(tabItems).
		Width(si.overlayWidth - 4).
		WrapTabs(true).
		TabGap(1).
		Build()

	// Build content using Stack component
	content := ui.NewVStack().
		SetChildren([]rtui.VNode{tabsComponent, ui.Text("─"), activeContent})

	// Use the new Panel component
	panel := componentpanel.NewBuilder().
		Header(titleBar).
		Content(content).
		Footer(ui.Text("─")).
		Width(si.overlayWidth).
		Height(si.overlayHeight).
		Rounded().
		Style(style.NewStyle().Background(style.Blue)).
		Build()

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
func (si *StandaloneInspector) buildOverlayContainer(content ui.VNode) ui.VNode {
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
	overlay := ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Border()).
		SetChildrenList([]ui.VNode{mainContent}).
		SetWidth(si.overlayWidth)

	return overlay
}

// buildTabBar creates the tab selection bar
func (si *StandaloneInspector) buildTabBar() ui.VNode {
	var tabs []ui.VNode

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
		{TabHitTest, "8", "HitTest"},
	}

	for _, item := range allTabs {
		// Highlight active tab
		if item.tab == si.activeTab {
			tabs = append(tabs,
				ui.NewTextBuilder(fmt.Sprintf("[%s]", item.name)).
					Style(style.FgBgBold(style.Yellow, style.Blue)).
					Build(),
			)
		} else {
			tabs = append(tabs,
				ui.NewTextBuilder(fmt.Sprintf(" %s(%s) ", item.name, item.key)).
					Style(style.Foreground(style.White)).
					Build(),
			)
		}
	}

	return rtui.HStack(tabs...)
}

// buildActiveTabContent builds content for the active tab
func (si *StandaloneInspector) buildActiveTabContent() ui.VNode {
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
	case TabHitTest:
		return si.buildHitTestTabContent()
	default:
		return ui.Text("Tab not implemented")
	}
}

// buildElementsTabContent builds content for Elements tab (used by Tab component)
func (si *StandaloneInspector) buildElementsTabContent() ui.VNode {
	// Tree statistics
	stats := si.treeView.GetTreeStats()

	header := rtui.VStack(
		ui.NewTextBuilder("📦 Layout Tree").
			Style(style.FgBold(style.Green)).
			Build(),
		ui.NewTextBuilder(fmt.Sprintf("Nodes: %d | Depth: %d | Leaves: %d",
			stats.TotalNodes, stats.MaxDepth, stats.LeafNodes)).
			Style(style.Foreground(style.White)).
			Build(),
		ui.NewTextBuilder("").
			Build(),
	)

	// Tree visualization - use TreeView component with navigation
	// Optimize: Only regenerate lines if tree structure or expansion state changed
	currentChangeCount := si.treeView.GetChangeCount()
	if currentChangeCount != si.lastTreeChangeCount || len(si.treeLines) == 0 {
		si.treeLines, si.treeTotalLines = si.treeView.GetTreeLines()
		si.lastTreeChangeCount = currentChangeCount

		log.InspectorLogger.Debug("Tree lines regenerated (count: %d)", len(si.treeLines))
		// Log first few lines to debug
		for i := 0; i < min(5, len(si.treeLines)); i++ {
			log.InspectorLogger.Debug("Tree line %d: %q", i, si.treeLines[i])
		}
	}

	// Create TreeView component with navigation support
	// Store in treeViewComponent for potential future use
	si.treeViewComponent = componenttreeview.NewBuilder().
		FromLines(si.treeLines).
		ExpandLevel(-1). // Expand all, expansion controlled by si.treeView
		ShowIcons(true).
		Compact(false).
		BuildVNode()

	// Build selected/focused element info
	// Priority: show selected (Enter) if available, otherwise show focused (↑↓)
	var targetVNode ui.VNode
	var targetPath string
	var displayType string

	if si.selectedVNode != nil {
		// User pressed Enter to select
		targetVNode = si.selectedVNode
		targetPath = si.selectedPath
		displayType = "Selected"
	}

	// Create info display
	var selectedInfo ui.VNode
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
			ui.NewTextBuilder(separator).
				Style(style.Foreground(theme.Muted())).
				Build(),
			ui.NewTextBuilder(infoText).
				Style(style.FgBold(style.Yellow)).
				Build(),
			ui.NewTextBuilder(fmt.Sprintf("Path: %s", targetPath)).
				Style(style.Foreground(style.Cyan)).
				Build(),
			ui.NewTextBuilder("").
				Build(),
		)
	} else {
		// No selection or focus - show empty placeholder
		separator := strings.Repeat("─", si.overlayWidth-4)

		selectedInfo = rtui.VStack(
			ui.NewTextBuilder(separator).
				Style(style.Foreground(theme.Muted())).
				Build(),
			ui.NewTextBuilder("").
				Build(),
		)
	}

	// TreeView
	treePreview := componenttreeview.NewBuilder().
		FromLines(si.treeLines).
		ExpandLevel(-1).
		ShowIcons(true).
		Compact(false).
		Build()

	// Instructions
	separator := strings.Repeat("─", si.overlayWidth-4)
	instructions := rtui.VStack(
		ui.NewTextBuilder(separator).
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.NewTextBuilder("Instructions:").
			Style(style.FgBold(style.Cyan)).
			Build(),
		ui.NewTextBuilder("  ↑↓: Navigate | Enter: Inspect").
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.NewTextBuilder("  E: Expand/Collapse").
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.NewTextBuilder("  PgUp/PgDn: Scroll tree").
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.NewTextBuilder("  Home/End: Top/Bottom").
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
func (si *StandaloneInspector) buildConsoleTabContent() ui.VNode {
	return rtui.VStack(
		ui.NewTextBuilder("💻 Console").
			Style(style.FgBold(style.Green)).
			Build(),
		ui.NewTextBuilder("").
			Build(),
		ui.NewTextBuilder("Console messages will appear here").
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.NewTextBuilder("Use Inspector.Log() to output messages").
			Style(style.Foreground(theme.Muted())).
			Build(),
	)
}

// buildPerformanceTabContent builds content for Performance tab
func (si *StandaloneInspector) buildPerformanceTabContent() ui.VNode {
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
		ui.NewTextBuilder("⚡ Performance").
			Style(style.FgBold(style.Green)).
			Build(),
		ui.NewTextBuilder("").
			Build(),
		ui.NewTextBuilder(metricsText).
			Style(style.Foreground(style.White)).
			Build(),
		ui.NewTextBuilder("").
			Build(),
		ui.NewTextBuilder("─").
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.NewTextBuilder("Real-time Performance Metrics").
			Style(style.FgBold(style.Cyan)).
			Build(),
		ui.NewTextBuilder("Data updates every frame").
			Style(style.Foreground(theme.Muted())).
			Build(),
	)
}

// buildDiagnosticsTabContent builds content for Diagnostics tab
func (si *StandaloneInspector) buildDiagnosticsTabContent() ui.VNode {
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
		ui.NewTextBuilder("🔧 Diagnostics").
			Style(style.FgBold(style.Green)).
			Build(),
		ui.NewTextBuilder("").
			Build(),
		ui.NewTextBuilder(summaryText).
			Style(style.Foreground(style.White)).
			Build(),
		ui.NewTextBuilder("").
			Build(),
		ui.NewTextBuilder("─").
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.NewTextBuilder("Layout Problems").
			Style(style.FgBold(style.Cyan)).
			Build(),
		ui.NewTextBuilder("No problems detected").
			Style(style.Foreground(style.Green)).
			Build(),
	)
}

// buildNetworkTabContent builds content for Network tab
func (si *StandaloneInspector) buildNetworkTabContent() ui.VNode {
	return rtui.VStack(
		ui.NewTextBuilder("🌐 Network").
			Style(style.FgBold(style.Green)).
			Build(),
		ui.NewTextBuilder("").
			Build(),
		ui.NewTextBuilder("Network activity will appear here").
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.NewTextBuilder("HTTP requests, WebSocket messages").
			Style(style.Foreground(theme.Muted())).
			Build(),
	)
}

// buildScreenInfoTabContent builds content for Screen Info tab
// Shows real-time screen size, mouse position, and button detection
func (si *StandaloneInspector) buildScreenInfoTabContent() ui.VNode {
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
		ui.NewTextBuilder("📺 Screen Information").
			Style(style.FgBold(style.Green)).
			Build(),
		ui.Text("─"),

		// Screen & Overlay (2 lines)
		ui.NewTextBuilder(screenInfo).
			Style(style.Foreground(style.White)).
			Build(),
		ui.NewTextBuilder(overlayInfo).
			Style(style.Foreground(style.White)).
			Build(),

		ui.Text(""),

		// Mouse (2 lines)
		ui.NewTextBuilder(mouseInfo).
			Style(style.Foreground(style.Cyan)).
			Build(),
		ui.NewTextBuilder(boundsInfo+"  "+insideInfo).
			Style(func() style.Style {
				if mouseInOverlay {
					return style.Foreground(style.Green)
				}
				return style.Foreground(style.Red)
			}()).
			Build(),

		ui.Text(""),

		// Elements (2 lines)
		ui.NewTextBuilder(hoverInfo).
			Style(style.Foreground(style.Yellow)).
			Build(),
		ui.NewTextBuilder(selInfo).
			Style(style.Foreground(style.Green)).
			Build(),

		ui.Text(""),

		// Button & State (2 lines)
		ui.NewTextBuilder(buttonInfo).
			Style(style.Foreground(style.White)).
			Build(),
		ui.NewTextBuilder(stateInfo).
			Style(style.Foreground(style.Magenta)).
			Build(),

		ui.Text(""),

		// Instructions
		ui.NewTextBuilder("Alt+H/J/K/L: Move overlay | Real-time updates").
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

// buildHitTestTabContent builds content for HitTest tab using ListVNode component
func (si *StandaloneInspector) buildHitTestTabContent() ui.VNode {
	// Update hit test entries from current app root
	entries := si.updateHitTestEntries()

	// Prepare data for ListVNode
	// The first row will be the column header
	rows := make([]string, 0, len(entries)+1)

	// Add column header as first data row
	colHeader := fmt.Sprintf("%-3s %-25s %-10s %-2s %-2s", "Z", "Node (Type/Tag/Key/Label)", "Bounds", "H", "C")
	rows = append(rows, colHeader)

	// Add entry rows (in reverse order - highest Z first)
	// Limit to 12 entries to demonstrate overflow handling
	maxDisplayEntries := 12
	displayed := 0
	for i := range entries {
		if displayed >= maxDisplayEntries {
			break
		}

		idx := len(entries) - 1 - i
		e := entries[idx]

		hitMark := "·"
		if e.HitTest == "YES" {
			hitMark = "✓"
		}
		clickMark := "·"
		if e.Clickable {
			clickMark = "Y"
		}

		line := fmt.Sprintf("%-3d %-25s %-10s %-2s",
			e.ZOrder, formatNodeInfo(e.Type, e.Tag, e.Key, e.Label), hitMark, clickMark)
		rows = append(rows, line)
		displayed++
	}

	// Add overflow indicator if there are more entries
	if len(entries) > maxDisplayEntries {
		overflowText := fmt.Sprintf("... (%d more entries)", len(entries)-maxDisplayEntries)
		rows = append(rows, overflowText)
	}

	// Build the list using ListVNode
	list := componentlist.NewBuilder().
		Header("🎯 Hit Test Data").
		Rows(rows).
		HeaderStyle(style.Style{}.Bold(true).Foreground(style.Color("green"))).
		RowStyle(style.Style{}).
		EmptyText("(no entries)").
		ShowSeparator(true).
		MaxRows(17). // Header + separator + colHeader + 12 entries + overflow
		Key("inspector-hittest-list").
		Build()

	// Create summary line (separate from the list)
	hoveredStr := si.formatHovered()
	summaryText := fmt.Sprintf("Entries:%d  Mouse:(%d,%d)  %s",
		len(entries), si.lastMouseX, si.lastMouseY, hoveredStr)

	// Wrap both summary and list in VStack
	return rtui.VStack(
		ui.NewTextBuilder(summaryText).
			Style(style.Foreground(style.White)).
			Build(),
		list,
	)
}

// updateHitTestEntries updates the hit test entries from the current app root
// NOTE: Caller must already hold si.mu lock (called from buildHitTestTabContent
// -> buildOverlayContent -> RenderOverlay/RenderContent which hold the lock)
// updateHitTestEntries collects and returns hit test entries from current app root
// NOTE: Returns entries slice (does not modify si.hitMapEntries directly)
// This avoidsktlock issues when called from RenderContent which holds RLock
func (si *StandaloneInspector) updateHitTestEntries() []HitTestEntry {
	var entries []HitTestEntry

	if si.appRoot == nil {
		log.InspectorLogger.Debug("updateHitTestEntries: appRoot is nil")
		return entries
	}

	log.InspectorLogger.Debug("updateHitTestEntries: appRoot type=%d, nil=%v",
		si.appRoot.Type(), si.appRoot == nil)

	// Traverse the VNode tree to collect hit test information
	si.collectHitTestEntries(si.appRoot, 0, 0, 0, &entries)

	log.InspectorLogger.Debug("update: collected %d entries", len(entries))
	return entries
}

// collectHitTestEntries recursively collects hit test entries from a VNode tree
// entries is an accumulator slice to append collected entries to
func (si *StandaloneInspector) collectHitTestEntries(node ui.VNode, x, y, zOrder int, entries *[]HitTestEntry) {
	if node == nil {
		return
	}

	// Debug logging
	log.InspectorLogger.Debug("collect: node type=%d, zOrder=%d", node.Type(), zOrder)

	// Get node bounds if available
	var bounds string
	var clickable bool

	// Try to get bounds from various sources
	// Signature 1: GetBounds() [4]int (array return) - used by ElementVNode
	if boundsProvider, ok := node.(interface{ GetBounds() [4]int }); ok {
		arr := boundsProvider.GetBounds()
		bx, by, bw, bh := arr[0], arr[1], arr[2], arr[3]
		log.InspectorLogger.Debug("  [4]int GetBounds: %v", arr)
		// Only consider nodes with positive size
		if bw > 0 && bh > 0 {
			bounds = fmt.Sprintf("%d,%d %dx%d", bx, by, bw, bh)
		}
	} else if boundsProvider, ok := node.(interface{ GetBounds() (int, int, int, int) }); ok {
		// Signature 2: GetBounds() (int, int, int, int) (tuple return)
		bx, by, bw, bh := boundsProvider.GetBounds()
		log.InspectorLogger.Debug("  (int,int,int,int) GetBounds: %d,%d,%d,%d", bx, by, bw, bh)
		// Only consider nodes with positive size
		if bw > 0 && bh > 0 {
			bounds = fmt.Sprintf("%d,%d %dx%d", bx, by, bw, bh)
		}
	} else {
		// No GetBounds method
		log.InspectorLogger.Debug("  No GetBounds method on node type=%d", node.Type())
	}

	// Skip nodes without valid bounds (but still recurse into children)
	if bounds == "" {
		// Still recurse into children to collect their info
		if node.Children() != nil {
			for _, child := range node.Children() {
				si.collectHitTestEntries(child, x, y, zOrder+1, entries)
			}
		}
	}

	// Parse bounds for hit testing
	bx, by, bw, bh := 0, 0, 0, 0
	fmt.Sscanf(bounds, "%d,%d %dx%d", &bx, &by, &bw, &bh)

	// Check if mouse is in bounds
	mouseInBounds := si.lastMouseX >= bx && si.lastMouseX < bx+bw &&
		si.lastMouseY >= by && si.lastMouseY < by+bh

	// Check if clickable (has onClick handler)
	props := node.Props()
	if props != nil {
		if _, hasOnClick := props["onClick"]; hasOnClick {
			clickable = true
		}
	}

	// Also check if it's a button
	if tagger, ok := node.(interface{ Tag() string }); ok {
		if tagger.Tag() == "button" {
			clickable = true
		}
	}

	hitTest := "NO"
	if mouseInBounds {
		hitTest = "YES"
	}

	// Collect extra debug info: tag, key, label
	var tag string
	var key string
	var label string

	// Get tag if available
	if tagger, ok := node.(interface{ Tag() string }); ok {
		tag = tagger.Tag()
	}

	// Get key from Props (not from VNode.Key() which may be overridden by Fiber)
	if props != nil {
		// Try _componentKey first
		if k, ok := props["_componentKey"].(string); ok {
			key = k
		}
		// If no _componentKey, try "key" prop
		if key == "" {
			if k, ok := props["key"].(string); ok {
				key = k
			}
		}
	}
	// If no key in props, fall back to VNode.Key() method
	if key == "" {
		if keyer, ok := node.(interface{ Key() string }); ok {
			key = keyer.Key()
		}
	}
	// Get label/text content from various sources
	// Try Label() method first (for buttons)
	if labeler, ok := node.(interface{ Label() string }); ok {
		label = labeler.Label()
	}
	// Try Content() method (for TextVNode)
	if label == "" {
		if contenter, ok := node.(interface{ Content() string }); ok {
			label = contenter.Content()
		}
	}
	// Fall back to GetTextContent utility
	if label == "" {
		label = rtui.GetTextContent(node)
	}
	// Truncate long labels
	if len(label) > 12 {
		label = label[:9] + "..."
	}

	entry := HitTestEntry{
		Type:      node.Type().String(),
		Bounds:    bounds,
		ZOrder:    zOrder,
		HitTest:   hitTest,
		Clickable: clickable,
		Tag:       tag,
		Key:       key,
		Label:     label,
	}
	*entries = append(*entries, entry)

	// Recurse into children
	if node.Children() != nil {
		for _, child := range node.Children() {
			si.collectHitTestEntries(child, x, y, zOrder+1, entries)
		}
	}
}

// buildLayoutTabContent builds content for Layout Diagnostics tab
func (si *StandaloneInspector) buildLayoutTabContent() ui.VNode {
	// Create diagnostic instance
	diagnostic := NewLayoutDiagnostic()

	// Analyze the selected node if available, otherwise analyze the entire app root
	var targetVNode ui.VNode
	var displayText string

	if si.selectedVNode != nil {
		targetVNode = si.selectedVNode
		displayText = fmt.Sprintf("Selected: %s", si.selectedVNode.Type().String())
	} else if si.appRoot != nil {
		targetVNode = si.appRoot
		displayText = "Analyzing entire app root"
	} else {
		return rtui.VStack(
			ui.NewTextBuilder("📐 Layout Diagnostics").
				Style(style.FgBold(style.Green)).
				Build(),
			ui.NewTextBuilder("").
				Build(),
			ui.NewTextBuilder("No VNode to analyze").
				Style(style.Foreground(theme.Muted())).
				Build(),
			ui.NewTextBuilder("Select a node in Elements tab first").
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
	lines := strings.Split(formattedResult, "")

	// Limit lines to fit in overlay
	maxLines := si.overlayHeight - 8
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, "... (truncated)")
	}

	// Convert lines to VNodes
	contentNodes := make([]ui.VNode, 0, len(lines))
	for _, line := range lines {
		// Skip empty lines for cleaner display
		if line == "" {
			continue
		}

		// Colorize based on content
		var text ui.VNode
		if strings.Contains(line, "✅") {
			text = ui.NewTextBuilder(line).
				Style(style.FgBold(style.Green)).
				Build()
		} else if strings.Contains(line, "❌") {
			text = ui.NewTextBuilder(line).
				Style(style.FgBold(style.Red)).
				Build()
		} else if strings.Contains(line, "⚠️") {
			text = ui.NewTextBuilder(line).
				Style(style.FgBold(style.Yellow)).
				Build()
		} else if strings.Contains(line, "SELECTED NODE") || strings.Contains(line, "═") {
			text = ui.NewTextBuilder(line).
				Style(style.FgBold(style.Cyan)).
				Build()
		} else if strings.Contains(line, "Constraints:") || strings.Contains(line, "Measured:") {
			text = ui.NewTextBuilder(line).
				Style(style.Foreground(style.White)).
				Build()
		} else if strings.Contains(line, "Issues:") {
			text = ui.NewTextBuilder(line).
				Style(style.FgBold(style.Yellow)).
				Build()
		} else {
			text = ui.NewTextBuilder(line).
				Style(style.Foreground(style.White)).
				Build()
		}
		contentNodes = append(contentNodes, text)
	}

	// Build the VStack with header and content
	headerNodes := []ui.VNode{
		ui.NewTextBuilder("📐 Layout Diagnostics").
			Style(style.FgBold(style.Green)).
			Build(),
		ui.NewTextBuilder(displayText).
			Style(style.Foreground(style.Cyan)).
			Build(),
		ui.NewTextBuilder("").
			Build(),
		ui.NewTextBuilder("─").
			Style(style.Foreground(theme.Muted())).
			Build(),
	}

	// Combine header and content
	allNodes := append(headerNodes, contentNodes...)

	return rtui.VStack(allNodes...)
}

// buildConsoleTab builds the Console tab
func (si *StandaloneInspector) buildConsoleTab() ui.VNode {
	return rtui.VStack(
		ui.NewTextBuilder("💻 Console").
			Style(style.FgBold(style.Green)).
			Build(),
		ui.NewTextBuilder("").
			Build(),
		ui.NewTextBuilder("Console messages will appear here").
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.NewTextBuilder("Use Inspector.Log() to output messages").
			Style(style.Foreground(theme.Muted())).
			Build(),
	)
}

// buildPerformanceTab builds the Performance tab
func (si *StandaloneInspector) buildPerformanceTab() ui.VNode {
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
		ui.NewTextBuilder("⚡ Performance").
			Style(style.FgBold(style.Green)).
			Build(),
		ui.NewTextBuilder("").
			Build(),
		ui.NewTextBuilder(metricsText).
			Style(style.Foreground(style.White)).
			Build(),
		ui.NewTextBuilder("").
			Build(),
		ui.NewTextBuilder("─").
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.NewTextBuilder("Real-time Performance Metrics").
			Style(style.FgBold(style.Cyan)).
			Build(),
		ui.NewTextBuilder("Data updates every frame").
			Style(style.Foreground(theme.Muted())).
			Build(),
	)
}

// buildDiagnosticsTab builds the Diagnostics tab
func (si *StandaloneInspector) buildDiagnosticsTab() ui.VNode {
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

	var problemList ui.VNode
	if len(problems) > 0 {
		var items []ui.VNode
		for i, p := range problems {
			if i >= 10 { // Show first 10
				items = append(items,
					ui.NewTextBuilder(fmt.Sprintf("... and %d more", len(problems)-10)).
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
					ui.NewTextBuilder(fmt.Sprintf("[%s] %s", severityStr, p.Type)).
						Style(severityStyle(style.Yellow)).
						Build(),
					ui.NewTextBuilder(fmt.Sprintf("  %s", p.Message)).
						Style(style.Foreground(style.White)).
						Build(),
					ui.NewTextBuilder(fmt.Sprintf("  → %s", p.Suggestion)).
						Style(style.Foreground(style.Cyan)).
						Build(),
					ui.NewTextBuilder("").
						Build(),
				),
			)
		}
		problemList = rtui.VStack(items...)
	} else {
		problemList = ui.NewTextBuilder("✓ No layout problems detected").
			Style(style.FgBold(style.Green)).
			Build()
	}

	return rtui.VStack(
		ui.NewTextBuilder("🔍 Diagnostics").
			Style(style.FgBold(style.Green)).
			Build(),
		ui.NewTextBuilder("").
			Build(),
		ui.NewTextBuilder(summaryText).
			Style(style.Foreground(style.White)).
			Build(),
		ui.NewTextBuilder("").
			Build(),
		ui.NewTextBuilder("─").
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.NewTextBuilder("Layout Problems:").
			Style(style.FgBold(style.Cyan)).
			Build(),
		ui.NewTextBuilder("").
			Build(),
		problemList,
	)
}

// buildNetworkTab builds the Network tab
func (si *StandaloneInspector) buildNetworkTab() ui.VNode {
	return rtui.VStack(
		ui.NewTextBuilder("🌐 Network").
			Style(style.FgBold(style.Green)).
			Build(),
		ui.NewTextBuilder("").
			Build(),
		ui.NewTextBuilder("Network activity monitoring").
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.NewTextBuilder("(Not yet implemented for TUI)").
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
func (si *StandaloneInspector) GetTreeViewComponent() *componenttreeview.VNode {
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
	log.InspectorLogger.Debug("%s", message)
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
	log.InspectorLogger.Debug("Position set to (%d, %d)", x, y)
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
	log.InspectorLogger.Debug("Moved by (%d, %d) to (%d, %d)", dx, dy, si.floatX, si.floatY)
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
	log.InspectorLogger.Debug("Key received: key='%s' modifiers=%s showKeyDebug=%v",
		key, modifiers, si.showKeyDebug)

	// Debug mode: toggle with Ctrl+D (when Inspector is visible)
	if key == "d" && ctrl {
		si.showKeyDebug = !si.showKeyDebug
		log.InspectorLogger.Debug("showKeyDebug toggled to %v", si.showKeyDebug)
		return true
	}

	// Layout Dump with Ctrl+L
	if key == "l" && ctrl {
		if si.appRoot != nil {
			analyzer := NewLayoutAnalyzer()
			snapshot := analyzer.Capture(si.appRoot, 0)
			treeStr := analyzer.FormatTree(snapshot)
			_ = os.WriteFile("layout_dump.txt", []byte(treeStr), 0644)
			log.InspectorLogger.Debug("Layout dump saved to layout_dump.txt")
		}
		return true
	}

	// Alt + Arrow keys to move the panel (use Intent pattern)
	if alt {
		si.mu.Unlock() // Unlock before emitting intent to avoid deadlock
		switch key {
		case "h", "left":
			go ui.EmitIntentGlobal(MoveInspectorIntent{Direction: "left"})
		case "l", "right":
			go ui.EmitIntentGlobal(MoveInspectorIntent{Direction: "right"})
		case "k", "up":
			go ui.EmitIntentGlobal(MoveInspectorIntent{Direction: "up"})
		case "j", "down":
			go ui.EmitIntentGlobal(MoveInspectorIntent{Direction: "down"})
		}
		return true
	}


	// Tab switching using Intent pattern (optimized with map)
	tabKeyMap := map[string]InspectorTab{
		"1": TabElements,
		"2": TabConsole,
		"3": TabPerformance,
		"4": TabDiagnostics,
		"5": TabLayout,
		"6": TabNetwork,
		"7": TabScreenInfo,
		"8": TabHitTest,
	}

	if tab, ok := tabKeyMap[key]; ok {
		si.mu.Unlock() // Unlock before emitting intent
		go ui.EmitIntentGlobal(SwitchTabIntent{Tab: tab})
		return true
	}

	// Tab cycling using Intent pattern
	if key == "tab" {
		si.mu.Unlock() // Unlock before emitting intent
		currentTab := si.activeTab
		shiftedTab := func() InspectorTab {
			if shift {
				currentTab--
				if currentTab < TabElements {
					currentTab = TabHitTest
				}
			} else {
				currentTab++
				if currentTab > TabHitTest {
					currentTab = TabElements
				}
			}
			return currentTab
		}()
		go ui.EmitIntentGlobal(SwitchTabIntent{Tab: shiftedTab})
		return true
	}

	// Tree scrolling and navigation using Intent pattern (only when Elements tab is active)
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

			// Handle tree navigation using Intent pattern
			var dir string
			switch key {
			case "up":
				dir = "up"
			case "down":
				dir = "down"
			case "pageup", "pgup":
				dir = "pageup"
			case "pagedown", "pgdn":
				dir = "pagedown"
			case "home":
				dir = "home"
			case "end":
				dir = "end"
			case "e", "enter":
				dir = "enter"
			}

			if dir != "" {
				si.mu.Unlock() // Unlock before emitting intent
				go ui.EmitIntentGlobal(TreeNavIntent{Direction: dir})
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
		case "pageup", "pgup":
			si.treeScrollOffset -= treeViewHeight
			if si.treeScrollOffset < 0 {
				si.treeScrollOffset = 0
			}
			log.InspectorLogger.Debug("Tree scrolled up to offset %d", si.treeScrollOffset)
			return true
		case "pagedown", "pgdn":
			si.treeScrollOffset += treeViewHeight
			if si.treeScrollOffset > maxOffset {
				si.treeScrollOffset = maxOffset
			}
			log.InspectorLogger.Debug("Tree scrolled down to offset %d", si.treeScrollOffset)
			return true
		case "home":
			si.treeScrollOffset = 0
			log.InspectorLogger.Debug("Tree scrolled to top")
			return true
		case "end":
			si.treeScrollOffset = maxOffset
			log.InspectorLogger.Debug("Tree scrolled to bottom")
			return true
		}

		return true
	}

	// When Inspector is visible, it's modal - capture ALL keyboard input
	// This prevents keys from falling through to the background app
	// (F12 and Ctrl+D are handled earlier by keyMap shortcuts)
	log.InspectorLogger.Debug("Visible mode: capturing key '%s' (alt=%v, ctrl=%v)",
		key, alt, ctrl)
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
	log.InspectorLogger.Debug("HandleMouseEvent: type=%v, pos=(%d,%d), button=%d, visible=%v",
		eventType, ev.X, ev.Y, ev.Button, si.visible)

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

		log.InspectorLogger.Debug("Overlay bounds: minX=%d, minY=%d, maxX=%d, maxY=%d",
			minX, minY, maxX, maxY)
		log.InspectorLogger.Debug("Mouse pos: (%d,%d), in overlay: %v",
			ev.X, ev.Y, ev.X >= minX && ev.X < maxX && ev.Y >= minY && ev.Y < maxY)

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
						log.InspectorLogger.Debug("Event handled by overlay component")
						return true
					}
				}
			}

			// Fallback: manual handling for tab bar
			handled := si.handleOverlayMouse(localX, localY, eventType, ev.Button)
			if handled {
				log.InspectorLogger.Debug("Event handled by manual fallback")
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
					log.InspectorLogger.Debug("Tab clicked: %s (row %d, col %d)", label, localY, localX)
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

// formatNodeInfo formats node information for display
// Parameters are organized as: (id, tag, key, label) to match call signature
func formatNodeInfo(nodeType, tag, key, label string) string {
	info := ""

	// Debug logging to trace the issue
	log.InspectorLogger.Debug("formatNodeInfo: type=%q tag=%q key=%q label=%q", nodeType, tag, key, label)

	// Use id as node type (Element, Text, etc.) - this is what we have from HitTestEntry.NodeID
	if nodeType != "" {
		info += nodeType
	}

	// Add tag in parentheses if available
	if tag != "" && tag != nodeType {
		if info != "" {
			info += "/"
		}
		info += "(" + tag + ")"
	}

	// Add key in brackets if available
	if key != "" {
		if info != "" {
			info += " "
		}
		// Truncate long keys
		if len(key) > 8 {
			key = key[:8] + "~"
		}
		info += "[" + key + "]"
	}

	// Add label in quotes if available (优先显示label)
	if label != "" {
		if info != "" {
			info += " "
		}
		// Truncate long labels
		if len(label) > 15 {
			label = label[:15] + "..."
		}
		info += "'" + label + "'"
	}

	// Fallback if empty
	if info == "" {
		info = "Unknown"
	}

	return info
}

// lookupPathForVNode tries to find the path of a VNode in the current tree.
func (si *StandaloneInspector) lookupPathForVNode(vnode ui.VNode) string {
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

// =============================================================================
// Custom Intent Types for MVP Pattern
// =============================================================================

// SwitchTabIntent represents an intent to switch to a specific inspector tab
type SwitchTabIntent struct {
	Tab InspectorTab
}

func (SwitchTabIntent) IntentType() string { return "SwitchTab" }
func (SwitchTabIntent) StayPressed() bool  { return true }

// TreeNavIntent represents an intent to navigate the tree view
type TreeNavIntent struct {
	Direction string // "up", "down", "pageup", "pagedown", "home", "end", "enter"
}

func (TreeNavIntent) IntentType() string { return "TreeNav" }
func (TreeNavIntent) StayPressed() bool  { return true }

// ToggleInspectorIntent represents an intent to toggle the inspector visibility
type ToggleInspectorIntent struct{}

func (ToggleInspectorIntent) IntentType() string { return "ToggleInspector" }
func (ToggleInspectorIntent) StayPressed() bool  { return true }

// MoveInspectorIntent represents an intent to move the floating inspector
type MoveInspectorIntent struct {
	Direction string // "up", "down", "left", "right"
}

func (MoveInspectorIntent) IntentType() string { return "MoveInspector" }
func (MoveInspectorIntent) StayPressed() bool  { return true }

// navigateTree navigates the tree view by the specified direction
// delta: -1 for up, 1 for down
func (si *StandaloneInspector) navigateTree(delta int) {
	if si.treeViewComponent == nil || si.treeLines == nil {
		return
	}

	// Calculate the current visible lines
	treeViewHeight := si.overlayHeight - 20
	if treeViewHeight < 1 {
		treeViewHeight = 1
	}

	// Update scroll offset
	si.treeScrollOffset += delta

	// Ensure scroll offset stays within bounds
	maxOffset := len(si.treeLines) - treeViewHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if si.treeScrollOffset < 0 {
		si.treeScrollOffset = 0
	}
	if si.treeScrollOffset > maxOffset {
		si.treeScrollOffset = maxOffset
	}
}

// =============================================================================
// Intent Handler Initialization (MVP Pattern)
// =============================================================================

// initIntentHandlers registers all intent handlers for the inspector
// This follows the MVP pattern where Intent carries data and handlers update state
func (si *StandaloneInspector) initIntentHandlers() {
	// Register SwitchTab handler
	rtui.RegisterIntent(func(ctx *intent.ActionContext, i SwitchTabIntent) intent.IntentResult {
		si.mu.Lock()
		defer si.mu.Unlock()

		// Ensure the tab is valid
		if i.Tab >= TabElements && i.Tab <= TabHitTest {
			si.activeTab = i.Tab
			log.InspectorLogger.Debug("Switched to tab %d via SwitchTabIntent", i.Tab)
		}
		return intent.HandledResult()
	})

	// Register TreeNav handler
	rtui.RegisterIntent(func(ctx *intent.ActionContext, i TreeNavIntent) intent.IntentResult {
		si.mu.Lock()
		defer si.mu.Unlock()

		// Only handle navigation when Elements tab is active
		if si.activeTab != TabElements {
			return intent.HandledResult()
		}

		switch i.Direction {
		case "up":
			si.navigateTree(-1)
			log.InspectorLogger.Debug("Tree: navigated up (Intent), scroll offset=%d", si.treeScrollOffset)
		case "down":
			si.navigateTree(1)
			log.InspectorLogger.Debug("Tree: navigated down (Intent), scroll offset=%d", si.treeScrollOffset)
		case "pageup":
			treeViewHeight := si.overlayHeight - 20
			if treeViewHeight < 1 {
				treeViewHeight = 1
			}
			si.treeScrollOffset -= treeViewHeight
			if si.treeScrollOffset < 0 {
				si.treeScrollOffset = 0
			}
			log.InspectorLogger.Debug("Tree: scrolled page up (Intent), scroll offset=%d", si.treeScrollOffset)
		case "pagedown":
			treeViewHeight := si.overlayHeight - 20
			if treeViewHeight < 1 {
				treeViewHeight = 1
			}
			maxOffset := len(si.treeLines) - treeViewHeight
			if maxOffset < 0 {
				maxOffset = 0
			}
			si.treeScrollOffset += treeViewHeight
			if si.treeScrollOffset > maxOffset {
				si.treeScrollOffset = maxOffset
			}
			log.InspectorLogger.Debug("Tree: scrolled page down (Intent), scroll offset=%d", si.treeScrollOffset)
		case "home":
			si.treeScrollOffset = 0
			log.InspectorLogger.Debug("Tree: scrolled to top (Intent)")
		case "end":
			treeViewHeight := si.overlayHeight - 20
			if treeViewHeight < 1 {
				treeViewHeight = 1
			}
			maxOffset := len(si.treeLines) - treeViewHeight
			if maxOffset < 0 {
				maxOffset = 0
			}
			si.treeScrollOffset = maxOffset
			log.InspectorLogger.Debug("Tree: scrolled to bottom (Intent)")
		case "enter":
			log.InspectorLogger.Debug("Tree: enter key pressed on line at offset %d (Intent)", si.treeScrollOffset)
		}
		return intent.HandledResult()
	})

	// Register ToggleInspector handler (no parameters, can use ui.On)
	ui.On(ToggleInspectorIntent{}, func() {
		si.ToggleVisibility()
	})

	// Register MoveInspector handler
	rtui.RegisterIntent(func(ctx *intent.ActionContext, i MoveInspectorIntent) intent.IntentResult {
		si.mu.Lock()
		defer si.mu.Unlock()

		switch i.Direction {
		case "up":
			si.floatY -= 1
			if si.floatY < 0 {
				si.floatY = 0
			}
			log.InspectorLogger.Debug("Moved inspector up to y=%d (Intent)", si.floatY)
		case "down":
			si.floatY += 1
			log.InspectorLogger.Debug("Moved inspector down to y=%d (Intent)", si.floatY)
		case "left":
			si.floatX -= 2
			if si.floatX < 0 {
				si.floatX = 0
			}
			log.InspectorLogger.Debug("Moved inspector left to x=%d (Intent)", si.floatX)
		case "right":
			si.floatX += 2
			log.InspectorLogger.Debug("Moved inspector right to x=%d (Intent)", si.floatX)
		}
		return intent.HandledResult()
	})

	log.InspectorLogger.Debug("Inspector intent handlers initialized")
}

