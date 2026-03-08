// Demo 2: Runtime Internals Visualization with UI Inspector
//
// This demo visualizes the internal runtime scheduling flow from setState
// to terminal buffer output - the "Total Assembly Diagram" of the engine.
//
// Pipeline:
//   Event → setState → Scheduler → Render → Reconcile → Layout
//   → Layer Merge → Paint → Buffer Diff → Terminal Output
//
// This version integrates the full UI Inspector for debugging and analysis.
//
// Based on: framework/docs/ui/demo/demo2_inside.md
//
// Architecture: Store + Reducer + Custom Intent (Single Source of Truth)

package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/inspector"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// AppState (Single Source of Truth)
// =============================================================================

// AppState represents the demo state.
type AppState struct {
	CurrentPhase  string
	EventCount    int
	RenderCount   int
	BufferUpdates int
	ShowInspector bool
}

// =============================================================================
// Custom Intent Types
// =============================================================================

// SetPhaseIntent sets the current pipeline phase.
type SetPhaseIntent struct {
	Phase string
}

func (SetPhaseIntent) IntentType() string { return "SetPhase" }
func (SetPhaseIntent) StayPressed() bool  { return true }

// SetPhaseWithRenderCountIntent sets phase and increments render count.
type SetPhaseWithRenderCountIntent struct {
	Phase string
}

func (SetPhaseWithRenderCountIntent) IntentType() string { return "SetPhaseWithRenderCount" }
func (SetPhaseWithRenderCountIntent) StayPressed() bool  { return true }

// SetPhaseWithBufferCountIntent sets phase and increments buffer count.
type SetPhaseWithBufferCountIntent struct {
	Phase string
}

func (SetPhaseWithBufferCountIntent) IntentType() string { return "SetPhaseWithBufferCount" }
func (SetPhaseWithBufferCountIntent) StayPressed() bool  { return true }

// SetPhaseWithEventCountIntent sets phase and increments event count.
type SetPhaseWithEventCountIntent struct {
	Phase string
}

func (SetPhaseWithEventCountIntent) IntentType() string { return "SetPhaseWithEventCount" }
func (SetPhaseWithEventCountIntent) StayPressed() bool  { return true }

// ToggleInspectorIntent toggles the inspector visibility.
type ToggleInspectorIntent struct{}

func (ToggleInspectorIntent) IntentType() string { return "ToggleInspector" }
func (ToggleInspectorIntent) StayPressed() bool  { return true }

// =============================================================================
// Reducer (Pure Function)
// =============================================================================

// appReducer handles all state transitions.
var appReducer = reducer.NewBuilder[AppState]()

// Initialize the reducer.
func init() {
	// Handle SetPhaseIntent
	appReducer.On(SetPhaseIntent{}, func(s AppState, i intent.Intent) AppState {
		spi := i.(SetPhaseIntent)
		s.CurrentPhase = spi.Phase
		return s
	})

	// Handle SetPhaseWithEventCountIntent
	appReducer.On(SetPhaseWithEventCountIntent{}, func(s AppState, i intent.Intent) AppState {
		pi := i.(SetPhaseWithEventCountIntent)
		s.CurrentPhase = pi.Phase
		s.EventCount++
		return s
	})

	// Handle SetPhaseWithRenderCountIntent
	appReducer.On(SetPhaseWithRenderCountIntent{}, func(s AppState, i intent.Intent) AppState {
		pi := i.(SetPhaseWithRenderCountIntent)
		s.CurrentPhase = pi.Phase
		s.RenderCount++
		return s
	})

	// Handle SetPhaseWithBufferCountIntent
	appReducer.On(SetPhaseWithBufferCountIntent{}, func(s AppState, i intent.Intent) AppState {
		pi := i.(SetPhaseWithBufferCountIntent)
		s.CurrentPhase = pi.Phase
		s.BufferUpdates++
		return s
	})

	// Handle ToggleInspectorIntent
	appReducer.On(ToggleInspectorIntent{}, func(s AppState, i intent.Intent) AppState {
		s.ShowInspector = !s.ShowInspector
		// Toggle global inspector
		if s.ShowInspector {
			globalInspector.Enable()
		} else {
			globalInspector.Disable()
		}
		return s
	})
}

// =============================================================================
// Store (Single State Source)
// =============================================================================

// appStore holds the demo state.
var appStore = store.NewStore(AppState{
	CurrentPhase:  "idle",
	EventCount:    0,
	RenderCount:   0,
	BufferUpdates: 0,
	ShowInspector: inspectorEnabled,
})

// =============================================================================
// Global inspector instance
// =============================================================================

var globalInspector *inspector.Inspector
var globalPerf *inspector.PerformanceAnalyzer
var globalDiagnostics *inspector.LayoutDiagnostics
var globalTreeView *inspector.TreeView
var globalExporter *inspector.ReportExporter
var globalEditor *inspector.PropertyEditor
var inspectorEnabled bool

func main() {
	// Initialize inspector components
	globalInspector = inspector.NewInspector()
	globalPerf = inspector.NewPerformanceAnalyzer()
	globalDiagnostics = inspector.NewLayoutDiagnostics()
	globalTreeView = inspector.NewTreeView()
	globalExporter = inspector.NewReportExporter(globalInspector)
	globalEditor = inspector.NewPropertyEditor()

	// Enable inspector from environment
	if os.Getenv("TUI_INSPECTOR") == "true" {
		inspectorEnabled = true
		globalInspector.Enable()
		globalPerf.Enable()
		fmt.Println("UI Inspector enabled - Press F12 to toggle")
	}

	// Check if layout debug mode is enabled
	if os.Getenv("TUI_UI_DEBUG_LAYOUT") == "true" || os.Getenv("TUI_LAYOUT_DEBUG") == "true" {
		TestLayoutInfo()
		return
	}

	// Initialize theme
	_ = theme.SetTheme("nord")

	// Register reducer handlers to store
	appReducer.RegisterToGlobal(appStore)

	err := ui.Run(RuntimeDemoWithInspector,
		ui.WithWidth(120), // Wider to accommodate inspector
		ui.WithHeight(40),
		ui.WithTitle("Mint TUI - Runtime Internals + Inspector"),
	)
	if err != nil {
		panic(err)
	}
}

// RuntimeDemoWithInspector wraps the demo with inspector integration
func RuntimeDemoWithInspector() ui.VNode {
	// Get current state snapshot from Store
	state := appStore.Get()

	// Track render performance
	globalPerf.StartFrame()
	defer globalPerf.EndFrame()

	// Build main UI
	mainContent := ui.VStack(
		HeaderPanel(),
		PipelineVisualization(state.CurrentPhase),
		StatisticsPanel(state.EventCount, state.RenderCount, state.BufferUpdates),
		ControlPanel(),
		ExplanationPanel(state.CurrentPhase),
	)

	// If inspector is enabled, show it alongside main content
	if state.ShowInspector {
		// Perform layout diagnostics
		rootVNode := mainContent
		globalDiagnostics.Analyze(rootVNode)

		// Update tree view
		globalTreeView.SetRoot(rootVNode)

		// Build inspector panel
		inspectorPanel := buildInspectorPanel(state.CurrentPhase, state.EventCount, state.RenderCount, state.BufferUpdates)

		// Show main content and inspector side by side
		// Use HStack to place inspector on the right side
		return ui.HStack(
			mainContent,
			ui.Text("│"), // Separator
			inspectorPanel,
		)
	}

	return mainContent
}

// buildInspectorPanel creates the inspector sidebar panel
func buildInspectorPanel(currentPhase string, eventCount, renderCount, bufferUpdates int) ui.VNode {
	var inspectorSections []ui.VNode

	// Header
	inspectorSections = append(inspectorSections,
		ui.NewTextBuilder("╔═ UI INSPECTOR ═╗").
			Style(style.FgBold(style.Yellow)).
			Build(),
	)

	// Current Phase Section
	inspectorSections = append(inspectorSections,
		ui.NewTextBuilder("┌─ Current Phase ─").
			Style(style.FgBold(style.Cyan)).
			Build(),
		ui.NewTextBuilder(fmt.Sprintf("Phase: %s", currentPhase)).
			Style(style.Foreground(style.White)).
			Build(),
		ui.NewTextBuilder("").
			Style(style.Foreground(theme.Muted())).
			Build(),
	)

	// Performance Section
	inspectorSections = append(inspectorSections, buildPerformanceSection())

	// Statistics Section
	inspectorSections = append(inspectorSections, buildStatisticsSection(eventCount, renderCount, bufferUpdates))

	// Diagnostics Section
	inspectorSections = append(inspectorSections, buildDiagnosticsSection())

	// Tree View Section (abbreviated)
	inspectorSections = append(inspectorSections, buildTreeViewSection())

	// Selected Element Section
	if globalInspector.GetSelectedVNode() != nil {
		inspectorSections = append(inspectorSections, buildSelectedElementSection())
	}

	// Instructions
	inspectorSections = append(inspectorSections,
		ui.NewTextBuilder("─────────────────").
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.NewTextBuilder("[I]: Toggle").
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.NewTextBuilder("Tab: Navigate").
			Style(style.Foreground(theme.Muted())).
			Build(),
	)

	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Info()).
		SetChildrenList([]ui.VNode{ui.VStack(inspectorSections...)})
}

// buildPerformanceSection creates performance metrics display
func buildPerformanceSection() ui.VNode {
	metrics := globalPerf.GetMetrics()

	var perfText string
	if metrics.FrameCount > 0 {
		perfText = fmt.Sprintf("FPS: %.1f | Mem: %s",
			metrics.FPS,
			formatBytes(metrics.LastHeapAlloc),
		)
	} else {
		perfText = "Performance: collecting..."
	}

	return ui.VStack(
		ui.NewTextBuilder("┌─ Performance ─").
			Style(style.FgBold(style.Cyan)).
			Build(),
		ui.NewTextBuilder(perfText).
			Style(style.Foreground(theme.Text())).
			Build(),
	)
}

// buildDiagnosticsSection creates diagnostics display
func buildDiagnosticsSection() ui.VNode {
	problems := globalDiagnostics.GetProblems()
	counts := globalDiagnostics.CountBySeverity()

	diagText := fmt.Sprintf("Problems: %d ERR, %d WARN",
		counts[inspector.SeverityError]+counts[inspector.SeverityCritical],
		counts[inspector.SeverityWarning])

	if len(problems) == 0 {
		diagText = "✓ No layout problems"
	}

	return ui.VStack(
		ui.NewTextBuilder("┌─ Diagnostics ──").
			Style(style.FgBold(style.Cyan)).
			Build(),
		ui.NewTextBuilder(diagText).
			Style(style.Foreground(theme.Text())).
			Build(),
	)
}

// buildTreeViewSection creates tree view summary
func buildTreeViewSection() ui.VNode {
	stats := globalTreeView.GetTreeStats()

	treeText := fmt.Sprintf("Nodes: %d | Depth: %d",
		stats.TotalNodes,
		stats.MaxDepth)

	return ui.VStack(
		ui.NewTextBuilder("┌─ Layout Tree ────").
			Style(style.FgBold(style.Cyan)).
			Build(),
		ui.NewTextBuilder(treeText).
			Style(style.Foreground(theme.Text())).
			Build(),
	)
}

// buildSelectedElementSection creates selected element info
func buildSelectedElementSection() ui.VNode {
	selected := globalInspector.GetSelectedInfo()

	elemText := fmt.Sprintf("%s (%s)",
		selected.Type,
		selected.Label)

	if selected.Label == "" {
		elemText = selected.Type
	}

	return ui.VStack(
		ui.NewTextBuilder("┌─ Selected ──────").
			Style(style.FgBold(style.Cyan)).
			Build(),
		ui.NewTextBuilder(elemText).
			Style(style.Foreground(theme.Text())).
			Build(),
		ui.NewTextBuilder(fmt.Sprintf("Pos: (%d, %d) | Size: %dx%d",
			selected.Position.X,
			selected.Position.Y,
			selected.Size.Width,
			selected.Size.Height)).
			Style(style.Foreground(theme.Muted())).
			Build(),
	)
}

// buildStatisticsSection creates statistics display
func buildStatisticsSection(eventCount, renderCount, bufferUpdates int) ui.VNode {
	return ui.VStack(
		ui.NewTextBuilder("┌─ Statistics ────").
			Style(style.FgBold(style.Cyan)).
			Build(),
		ui.NewTextBuilder(fmt.Sprintf("Events: %d", eventCount)).
			Style(style.Foreground(style.White)).
			Build(),
		ui.NewTextBuilder(fmt.Sprintf("Renders: %d", renderCount)).
			Style(style.Foreground(style.White)).
			Build(),
		ui.NewTextBuilder(fmt.Sprintf("Buffers: %d", bufferUpdates)).
			Style(style.Foreground(style.White)).
			Build(),
		ui.NewTextBuilder("").
			Style(style.Foreground(theme.Muted())).
			Build(),
	)
}

// HeaderPanel shows the title with border
func HeaderPanel() ui.VNode {
	headerContent := ui.HStackBuilder(
		ui.NewTextBuilder("Runtime Scheduling Pipeline Visualization").
			Style(style.FgBold(theme.Text())).
			Build(),
	).
		Gap(0).
		Align(ui.AlignCenter).
		Build()

	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Primary()).
		SetChildrenList([]ui.VNode{headerContent})
}

// PipelineVisualization shows the runtime pipeline flow
func PipelineVisualization(currentPhase string) ui.VNode {
	phases := []struct {
		name     string
		color    string
		position int
	}{
		{"Event", "red", 0},
		{"setState", "yellow", 15},
		{"Scheduler", "green", 30},
		{"Render", "blue", 45},
		{"Reconcile", "magenta", 60},
		{"Layout", "cyan", 75},
		{"Paint", "white", 90},
	}

	activeIndex := -1
	for i, p := range phases {
		if p.name == currentPhase {
			activeIndex = i
			break
		}
	}

	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Border()).
		SetChildrenList([]ui.VNode{
			ui.VStack(
				buildPipelineLine(phases, activeIndex),
				ui.Text(""),
				buildPipelineArrows(phases, activeIndex),
			),
		})
}

// buildPipelineLine creates the phase boxes
func buildPipelineLine(phases []struct{ name string; color string; position int }, activeIndex int) ui.VNode {
	var result string
	for _, p := range phases {
		spaces := p.position - len(result)
		for i := 0; i < spaces; i++ {
			result += " "
		}
		result += "[" + p.name + "]"
	}

	return ui.NewTextBuilder(result).
		Style(style.Foreground(theme.Muted())).
		Build()
}

// buildPipelineArrows creates the flow arrows
func buildPipelineArrows(phases []struct{ name string; color string; position int }, activeIndex int) ui.VNode {
	var result string
	for i := range phases {
		if i > 0 {
			result += "       "
		}
		if i < len(phases)-1 {
			result += "  ↓  "
		}
	}

	return ui.NewTextBuilder(result).
		Style(style.Foreground(theme.Muted())).
		Build()
}

// StatisticsPanel shows runtime statistics
func StatisticsPanel(eventCount, renderCount, bufferUpdates int) ui.VNode {
	content := ui.HStack(
		ui.NewTextBuilder("Events:").
			Style(style.Foreground(theme.Text())).
			Build(),
		ui.NewTextBuilder(fmt.Sprintf("%6d", eventCount)).
			Style(style.FgBgBold(theme.Error(), theme.BG())).
			Build(),
		ui.NewTextBuilder("  Renders:").
			Style(style.Foreground(theme.Text())).
			Build(),
		ui.NewTextBuilder(fmt.Sprintf("%6d", renderCount)).
			Style(style.FgBgBold(theme.Info(), theme.BG())).
			Build(),
		ui.NewTextBuilder("  Buffers:").
			Style(style.Foreground(theme.Text())).
			Build(),
		ui.NewTextBuilder(fmt.Sprintf("%6d", bufferUpdates)).
			Style(style.FgBgBold(theme.Success(), theme.BG())).
			Build(),
	)

	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Border()).
		SetChildrenList([]ui.VNode{content})
}

// ControlPanel provides buttons to trigger each phase
func ControlPanel() ui.VNode {
	allButtons := []ui.VNode{
		ui.NewButtonBuilder("[1] Event").
			Variant(ui.ButtonVariantDanger).
			OnPress(SetPhaseWithEventCountIntent{Phase: "Event"}).
			FocusStyle(ui.FocusStyleBracket).
			Build(),
		ui.NewButtonBuilder("[2]setState").
			Variant(ui.ButtonVariantSecondary).
			OnPress(SetPhaseIntent{Phase: "setState"}).
			FocusStyle(ui.FocusStyleBracket).
			Build(),
		ui.NewButtonBuilder("[3]Scheduler").
			Variant(ui.ButtonVariantSuccess).
			OnPress(SetPhaseWithRenderCountIntent{Phase: "Scheduler"}).
			FocusStyle(ui.FocusStyleBracket).
			Build(),
		ui.NewButtonBuilder("[4] Render").
			Variant(ui.ButtonVariantPrimary).
			OnPress(SetPhaseIntent{Phase: "Render"}).
			FocusStyle(ui.FocusStyleBracket).
			Build(),
		ui.NewButtonBuilder("[5]Reconcile").
			OnPress(SetPhaseIntent{Phase: "Reconcile"}).
			FocusStyle(ui.FocusStyleBracket).
			Build(),
		ui.NewButtonBuilder("[6] Layout").
			OnPress(SetPhaseIntent{Phase: "Layout"}).
			FocusStyle(ui.FocusStyleBracket).
			Build(),
		ui.NewButtonBuilder("[7] Paint").
			OnPress(SetPhaseWithBufferCountIntent{Phase: "Paint"}).
			FocusStyle(ui.FocusStyleBracket).
			Build(),
		ui.NewButtonBuilder("[0] Idle").
			OnPress(SetPhaseIntent{Phase: "idle"}).
			FocusStyle(ui.FocusStyleBracket).
			Build(),
		ui.NewButtonBuilder("[I] Toggle Inspector").
			Variant(ui.ButtonVariantSecondary).
			OnPress(ToggleInspectorIntent{}).
			FocusStyle(ui.FocusStyleBracket).
			Build(),
	}

	wrappedButtons := ui.NewWrapBuilder().Children(allButtons...).
		Gap(1).
		RowGap(0).
		Width(78).
		Align(ui.AlignCenter).
		FillWidth().
		Build()

	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Border()).
		SetChildrenList([]ui.VNode{wrappedButtons})
}

// ExplanationPanel shows detailed explanation of each phase
func ExplanationPanel(currentPhase string) ui.VNode {
	explanations := map[string]string{
		"idle":      "System idle, waiting for events...",
		"Event":     "Event captured from terminal, queued for processing.",
		"setState":  "State changes queued, components marked dirty for re-render.",
		"Scheduler": "Batches dirty components, schedules rendering with time-slicing.",
		"Render":    "Component functions called to generate VNode trees from state.",
		"Reconcile": "Diff algorithm compares old/new VNode trees, computes minimal changes.",
		"Layout":    "Constraint-based layout computes position (x,y) and size (w,h).",
		"Paint":     "Nodes render to back buffer. Only dirty regions are updated.",
	}

	text := explanations[currentPhase]
	if text == "" {
		text = "Select a phase to see detailed explanation..."
	}

	content := ui.NewTextBuilder(text).
		Style(style.Foreground(theme.Text())).
		Build()

	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Border()).
		SetChildrenList([]ui.VNode{content})
}

// Helper function to format bytes
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

// TestLayoutInfo runs layout info tests
func TestLayoutInfo() {
	fmt.Println("=== Layout Info Test ===")
	// Implementation would go here
}
