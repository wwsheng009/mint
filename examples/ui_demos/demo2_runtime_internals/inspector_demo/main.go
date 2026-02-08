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

package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/inspector"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// Global inspector instance
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
	currentPhase, setCurrentPhase := ui.UseStateString("idle")
	eventCount, setEventCount, _ := ui.UseStateInt(0)
	renderCount, setRenderCount, _ := ui.UseStateInt(0)
	bufferUpdates, setBufferUpdates, _ := ui.UseStateInt(0)
	showInspector, setShowInspector := ui.UseStateBool(inspectorEnabled)

	// Track render performance
	globalPerf.StartFrame()
	defer globalPerf.EndFrame()

	// Build main UI
	mainContent := ui.VStack(
		HeaderPanel(),
		PipelineVisualization(currentPhase),
		StatisticsPanel(eventCount, renderCount, bufferUpdates),
		ControlPanel(setCurrentPhase, setEventCount, setRenderCount, setBufferUpdates, setShowInspector),
		ExplanationPanel(currentPhase),
	)

	// If inspector is enabled, show it alongside main content
	if showInspector {
		// Perform layout diagnostics
		rootVNode := mainContent
		globalDiagnostics.Analyze(rootVNode)

		// Update tree view
		globalTreeView.SetRoot(rootVNode)

		// Build inspector panel
		inspectorPanel := buildInspectorPanel(currentPhase, eventCount, renderCount, bufferUpdates)

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
		app.NewTextBuilder("╔═ UI INSPECTOR ═╗").
			Style(style.FgBold(style.Yellow)).
			Build(),
	)

	// Current Phase Section
	inspectorSections = append(inspectorSections,
		app.NewTextBuilder("┌─ Current Phase ─").
			Style(style.FgBold(style.Cyan)).
			Build(),
		app.NewTextBuilder(fmt.Sprintf("Phase: %s", currentPhase)).
			Style(style.Foreground(style.White)).
			Build(),
		app.NewTextBuilder("").
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
		app.NewTextBuilder("─────────────────").
			Style(style.Foreground(theme.Muted())).
			Build(),
		app.NewTextBuilder("[I]: Toggle").
			Style(style.Foreground(theme.Muted())).
			Build(),
		app.NewTextBuilder("Tab: Navigate").
			Style(style.Foreground(theme.Muted())).
			Build(),
	)

	return ui.Bordered().
		Style(string(theme.Info())).
		Child(ui.VStack(inspectorSections...)).
		FillHeight().
		Width(50). // Set fixed width for inspector panel
		Build()
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
		app.NewTextBuilder("┌─ Performance ─").
			Style(style.FgBold(style.Cyan)).
			Build(),
		app.NewTextBuilder(perfText).
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
		app.NewTextBuilder("┌─ Diagnostics ──").
			Style(style.FgBold(style.Cyan)).
			Build(),
		app.NewTextBuilder(diagText).
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
		app.NewTextBuilder("┌─ Layout Tree ────").
			Style(style.FgBold(style.Cyan)).
			Build(),
		app.NewTextBuilder(treeText).
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
		app.NewTextBuilder("┌─ Selected ──────").
			Style(style.FgBold(style.Cyan)).
			Build(),
		app.NewTextBuilder(elemText).
			Style(style.Foreground(theme.Text())).
			Build(),
		app.NewTextBuilder(fmt.Sprintf("Pos: (%d, %d) | Size: %dx%d",
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
		app.NewTextBuilder("┌─ Statistics ────").
			Style(style.FgBold(style.Cyan)).
			Build(),
		app.NewTextBuilder(fmt.Sprintf("Events: %d", eventCount)).
			Style(style.Foreground(style.White)).
			Build(),
		app.NewTextBuilder(fmt.Sprintf("Renders: %d", renderCount)).
			Style(style.Foreground(style.White)).
			Build(),
		app.NewTextBuilder(fmt.Sprintf("Buffers: %d", bufferUpdates)).
			Style(style.Foreground(style.White)).
			Build(),
		app.NewTextBuilder("").
			Style(style.Foreground(theme.Muted())).
			Build(),
	)
}

// HeaderPanel shows the title with border
func HeaderPanel() ui.VNode {
	headerContent := ui.HStackBuilder(
		app.NewTextBuilder("Runtime Scheduling Pipeline Visualization").
			Style(style.FgBold(theme.Text())).
			Build(),
	).
		Gap(0).
		Align(ui.AlignCenter).
		Build()

	return ui.Bordered().
		Style(string(theme.Primary())).
		Child(headerContent).
		FillWidth().
		Build()
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

	return ui.Bordered().
		Style(string(theme.Border())).
		Child(
			ui.VStack(
				buildPipelineLine(phases, activeIndex),
				ui.Text(""),
				buildPipelineArrows(phases, activeIndex),
			),
		).
		Build()
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

	return app.NewTextBuilder(result).
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

	return app.NewTextBuilder(result).
		Style(style.Foreground(theme.Muted())).
		Build()
}

// StatisticsPanel shows runtime statistics
func StatisticsPanel(eventCount, renderCount, bufferUpdates int) ui.VNode {
	content := ui.HStack(
		app.NewTextBuilder("Events:").
			Style(style.Foreground(theme.Text())).
			Build(),
		app.NewTextBuilder(fmt.Sprintf("%6d", eventCount)).
			Style(style.FgBgBold(theme.Error(), theme.BG())).
			Build(),
		app.NewTextBuilder("  Renders:").
			Style(style.Foreground(theme.Text())).
			Build(),
		app.NewTextBuilder(fmt.Sprintf("%6d", renderCount)).
			Style(style.FgBgBold(theme.Info(), theme.BG())).
			Build(),
		app.NewTextBuilder("  Buffers:").
			Style(style.Foreground(theme.Text())).
			Build(),
		app.NewTextBuilder(fmt.Sprintf("%6d", bufferUpdates)).
			Style(style.FgBgBold(theme.Success(), theme.BG())).
			Build(),
	)

	return ui.Bordered().
		Style(string(theme.Border())).
		Child(content).
		Build()
}

// ControlPanel provides buttons to trigger each phase
func ControlPanel(
	setCurrentPhase func(string),
	setEventCount func(interface{}),
	setRenderCount func(interface{}),
	setBufferUpdates func(interface{}),
	setShowInspector func(bool),
) ui.VNode {
	allButtons := []ui.VNode{
		app.ButtonBuilder("[1] Event").
			Variant(app.ButtonVariantDanger).
			OnClick(func() {
				setCurrentPhase("Event")
				setEventCount(func(c int) int { return c + 1 })
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[2]setState").
			Variant(app.ButtonVariantSecondary).
			OnClick(func() {
				setCurrentPhase("setState")
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[3]Scheduler").
			Variant(app.ButtonVariantSuccess).
			OnClick(func() {
				setCurrentPhase("Scheduler")
				setRenderCount(func(c int) int { return c + 1 })
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[4] Render").
			Variant(app.ButtonVariantPrimary).
			OnClick(func() {
				setCurrentPhase("Render")
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[5]Reconcile").
			OnClick(func() {
				setCurrentPhase("Reconcile")
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[6] Layout").
			OnClick(func() {
				setCurrentPhase("Layout")
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[7] Paint").
			OnClick(func() {
				setCurrentPhase("Paint")
				setBufferUpdates(func(c int) int { return c + 1 })
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[0] Idle").
			OnClick(func() {
				setCurrentPhase("idle")
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("[I] Toggle Inspector").
			Variant(app.ButtonVariantSecondary).
			OnClick(func() {
				// Toggle inspector state
				newState := !inspectorEnabled
				inspectorEnabled = newState

				// Update global inspector
				if newState {
					globalInspector.Enable()
				} else {
					globalInspector.Disable()
				}

				// Update UI state to trigger re-render
				setShowInspector(newState)
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
	}

	wrappedButtons := app.WrapBuilder(allButtons...).
		Gap(1).
		RowGap(0).
		ScreenWidth(78).
		Align(ui.AlignCenter).
		FillWidth().
		Build()

	return ui.Bordered().
		Style(string(theme.Border())).
		Child(wrappedButtons).
		FillWidth().
		Build()
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

	content := app.NewTextBuilder(text).
		Style(style.Foreground(theme.Text())).
		Build()

	return ui.Bordered().
		Style(string(theme.Border())).
		Child(content).
		Build()
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
