// UI Inspector Feature Demo
//
// This program demonstrates the UI Inspector features:
// 1. Performance Analysis
// 2. Layout Diagnostics
// 3. Tree View
// 4. Property Editing
// 5. Report Export

package main

import (
	"fmt"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/inspector"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	fmt.Println("=== UI Inspector Feature Demo ===")
	fmt.Println("")
	fmt.Println("This demo showcases all UI Inspector features:")
	fmt.Println("1. Performance Analysis (FPS, memory, render time)")
	fmt.Println("2. Layout Diagnostics (problems, warnings, errors)")
	fmt.Println("3. Tree View (VNode tree structure)")
	fmt.Println("4. Selected Element Info (position, size, properties)")
	fmt.Println("5. Report Export (Text, Markdown, JSON)")
	fmt.Println("")
	fmt.Println("Press Ctrl+C to exit")
	fmt.Println("")

	// Initialize inspector components
	insp := inspector.NewInspector()
	perf := inspector.NewPerformanceAnalyzer()
	diagnostics := inspector.NewLayoutDiagnostics()
	treeView := inspector.NewTreeView()
	exporter := inspector.NewReportExporter(insp)

	// Enable all components
	insp.Enable()
	perf.Enable()

	// Initialize theme
	_ = theme.SetTheme("nord")

	// Run the demo
	err := ui.Run(func() ui.VNode {
		return InspectorDemo(insp, perf, diagnostics, treeView, exporter)
	},
		ui.WithWidth(100),
		ui.WithHeight(30),
		ui.WithTitle("UI Inspector Demo"),
	)

	if err != nil {
		panic(err)
	}

	// Export report on exit
	fmt.Println("\n=== Generating Final Report ===")
	markdownReport := exporter.Export(inspector.FormatMarkdown)
	fmt.Println(markdownReport)
}

// InspectorDemo creates the demo UI
func InspectorDemo(
	insp *inspector.Inspector,
	perf *inspector.PerformanceAnalyzer,
	diagnostics *inspector.LayoutDiagnostics,
	treeView *inspector.TreeView,
	exporter *inspector.ReportExporter,
) ui.VNode {
	// Track performance
	perf.StartFrame()
	defer perf.EndFrame()

	// Build demo content
	content := BuildDemoContent()

	// Analyze layout
	diagnostics.Analyze(content)

	// Update tree view
	treeView.SetRoot(content)

	// Build UI
	return ui.VStack(
		BuildHeader(),
		BuildPerformancePanel(perf),
		BuildDiagnosticsPanel(diagnostics),
		BuildTreePanel(treeView),
		BuildDemoContentPanel(content),
		BuildInstructions(),
	)
}

// BuildHeader creates the header
func BuildHeader() ui.VNode {
	header := ui.NewTextBuilder("╔══════════════════════════════════════════════════╗").
		Style(style.FgBold(style.Cyan)).
		Build()

	subheader := ui.NewTextBuilder("║     UI Inspector Feature Demonstration              ║").
		Style(style.FgBold(style.Yellow)).
		Build()

	border := ui.NewTextBuilder("╚══════════════════════════════════════════════════╝").
		Style(style.FgBold(style.Cyan)).
		Build()

	return ui.VStack(header, subheader, border)
}

// BuildPerformancePanel creates performance display
func BuildPerformancePanel(perf *inspector.PerformanceAnalyzer) ui.VNode {
	metrics := perf.GetMetrics()

	var fpsText, memText string
	if metrics.FrameCount > 0 {
		fpsText = fmt.Sprintf("FPS: %.1f", metrics.FPS)
		memText = fmt.Sprintf("Memory: %s", formatBytes(metrics.LastHeapAlloc))
	} else {
		fpsText = "FPS: collecting..."
		memText = "Memory: collecting..."
	}

	label := ui.NewTextBuilder("┌─ Performance ──────────────────────────────────").
		Style(style.FgBold(style.Green)).
		Build()

	fps := ui.NewTextBuilder(fpsText).
		Style(style.Foreground(style.White)).
		Build()

	mem := ui.NewTextBuilder(memText).
		Style(style.Foreground(style.White)).
		Build()

	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Border()).
		SetChildrenList([]ui.VNode{ui.VStack(label, fps, mem)})
}

// BuildDiagnosticsPanel creates diagnostics display
func BuildDiagnosticsPanel(diagnostics *inspector.LayoutDiagnostics) ui.VNode {
	problems := diagnostics.GetProblems()
	counts := diagnostics.CountBySeverity()

	var diagText string
	if len(problems) == 0 {
		diagText = "✓ No layout problems detected"
	} else {
		diagText = fmt.Sprintf("Found %d problems: %d Critical, %d Errors, %d Warnings",
			len(problems),
			counts[inspector.SeverityCritical],
			counts[inspector.SeverityError],
			counts[inspector.SeverityWarning])
	}

	label := ui.NewTextBuilder("┌─ Layout Diagnostics ───────────────────────────").
		Style(style.FgBold(style.Magenta)).
		Build()

	text := ui.NewTextBuilder(diagText).
		Style(style.Foreground(style.White)).
		Build()

	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Border()).
		SetChildrenList([]ui.VNode{ui.VStack(label, text)})
}

// BuildTreePanel creates tree view display
func BuildTreePanel(treeView *inspector.TreeView) ui.VNode {
	stats := treeView.GetTreeStats()

	treeText := fmt.Sprintf("Total Nodes: %d | Max Depth: %d | Leaf Nodes: %d",
		stats.TotalNodes,
		stats.MaxDepth,
		stats.LeafNodes)

	label := ui.NewTextBuilder("┌─ Layout Tree ────────────────────────────────────").
		Style(style.FgBold(style.Cyan)).
		Build()

	text := ui.NewTextBuilder(treeText).
		Style(style.Foreground(style.White)).
		Build()

	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Border()).
		SetChildrenList([]ui.VNode{ui.VStack(label, text)})
}

// BuildDemoContentPanel shows the actual demo content
func BuildDemoContentPanel(content ui.VNode) ui.VNode {
	label := ui.NewTextBuilder("┌─ Demo Content ───────────────────────────────────").
		Style(style.FgBold(style.Yellow)).
		Build()

	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Border()).
		SetChildrenList([]ui.VNode{ui.VStack(label, content)})
}

// BuildDemoContent creates the demo VNode tree
func BuildDemoContent() ui.VNode {
	button1 := ui.NewButtonBuilder("Button 1").
		Variant(ui.ButtonVariantPrimary).
		Build()

	button2 := ui.NewButtonBuilder("Button 2").
		Variant(ui.ButtonVariantSuccess).
		Build()

	text1 := ui.Text("Hello, UI Inspector!")
	text2 := ui.Text("This is a demonstration.")

	// Create nested structure
	row := ui.HStack(button1, button2)
	column := ui.VStack(text1, text2, row)

	// Wrap in bordered container
	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Info()).
		SetChildrenList([]ui.VNode{column})
}

// BuildInstructions shows usage instructions
func BuildInstructions() ui.VNode {
	instructions := []string{
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
		"Instructions:",
		"  • Observe real-time performance metrics above",
		"  • Layout diagnostics run automatically each frame",
		"  • Tree view shows the VNode structure",
		"  • Press Ctrl+C to exit and see the final report",
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
	}

	var nodes []ui.VNode
	for _, text := range instructions {
		nodes = append(nodes, ui.NewTextBuilder(text).
			Style(style.Foreground(theme.Muted())).
			Build())
	}

	return ui.VStack(nodes...)
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
