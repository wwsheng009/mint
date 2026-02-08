package inspector

import (
	"fmt"
	"strings"
	"time"
)

// ReportFormat represents the format of exported reports
type ReportFormat int

const (
	FormatText ReportFormat = iota // Plain text
	FormatMarkdown                 // Markdown
	FormatJSON                     // JSON
)

// ReportExporter exports inspection reports
type ReportExporter struct {
	inspector    *Inspector
	treeView     *TreeView
	diagnostics  *LayoutDiagnostics
	performance  *PerformanceAnalyzer
}

// NewReportExporter creates a new report exporter
func NewReportExporter(inspector *Inspector) *ReportExporter {
	return &ReportExporter{
		inspector:   inspector,
		treeView:    NewTreeView(),
		diagnostics: NewLayoutDiagnostics(),
		performance: NewPerformanceAnalyzer(),
	}
}

// SetTreeView sets the tree view to include in reports
func (re *ReportExporter) SetTreeView(tv *TreeView) {
	re.treeView = tv
}

// SetDiagnostics sets the diagnostics to include in reports
func (re *ReportExporter) SetDiagnostics(ld *LayoutDiagnostics) {
	re.diagnostics = ld
}

// SetPerformance sets the performance analyzer to include in reports
func (re *ReportExporter) SetPerformance(pa *PerformanceAnalyzer) {
	re.performance = pa
}

// Export generates a complete inspection report
func (re *ReportExporter) Export(format ReportFormat) string {
	switch format {
	case FormatMarkdown:
		return re.exportMarkdown()
	case FormatJSON:
		return re.exportJSON()
	case FormatText:
		return re.exportText()
	default:
		return re.exportText()
	}
}

// exportText exports report as plain text
func (re *ReportExporter) exportText() string {
	var sections []string

	// Header
	sections = append(sections, re.header())

	// Selected Element Info
	if re.inspector.GetSelectedVNode() != nil {
		sections = append(sections, re.selectedElementSection())
	}

	// Tree View
	if re.treeView != nil && re.treeView.root != nil {
		sections = append(sections, re.treeViewSection())
	}

	// Diagnostics
	if re.diagnostics != nil && len(re.diagnostics.GetProblems()) > 0 {
		sections = append(sections, re.diagnosticsSection())
	}

	// Performance
	if re.performance != nil && re.performance.GetMetrics().FrameCount > 0 {
		sections = append(sections, re.performanceSection())
	}

	// Footer
	sections = append(sections, re.footer())

	return strings.Join(sections, "\n\n")
}

// exportMarkdown exports report as markdown
func (re *ReportExporter) exportMarkdown() string {
	var md strings.Builder

	// Header
	md.WriteString("# UI Inspector Report\n\n")
	md.WriteString(fmt.Sprintf("**Generated**: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// Table of Contents
	md.WriteString("## Table of Contents\n\n")
	md.WriteString("- [Selected Element](#selected-element)\n")
	if re.treeView != nil && re.treeView.root != nil {
		md.WriteString("- [Layout Tree](#layout-tree)\n")
	}
	if re.diagnostics != nil && len(re.diagnostics.GetProblems()) > 0 {
		md.WriteString("- [Layout Problems](#layout-problems)\n")
	}
	if re.performance != nil && re.performance.GetMetrics().FrameCount > 0 {
		md.WriteString("- [Performance](#performance)\n")
	}
	md.WriteString("\n")

	// Selected Element
	if re.inspector.GetSelectedVNode() != nil {
		md.WriteString("## Selected Element\n\n")
		md.WriteString("```\n")
		selected := re.inspector.GetSelectedInfo()
		sidebar := NewSidebar()
		md.WriteString(sidebar.GetCopyableText(selected))
		md.WriteString("\n```\n\n")
	}

	// Tree View
	if re.treeView != nil && re.treeView.root != nil {
		md.WriteString("## Layout Tree\n\n")
		md.WriteString("```\n")
		md.WriteString(re.treeView.FormatTree())
		md.WriteString("\n```\n\n")
	}

	// Diagnostics
	if re.diagnostics != nil && len(re.diagnostics.GetProblems()) > 0 {
		md.WriteString("## Layout Problems\n\n")
		problems := re.diagnostics.GetProblems()

		// Summary table
		counts := re.diagnostics.CountBySeverity()
		md.WriteString("| Severity | Count |\n")
		md.WriteString("|----------|-------|\n")
		md.WriteString(fmt.Sprintf("| Critical | %d |\n", counts[SeverityCritical]))
		md.WriteString(fmt.Sprintf("| Error | %d |\n", counts[SeverityError]))
		md.WriteString(fmt.Sprintf("| Warning | %d |\n", counts[SeverityWarning]))
		md.WriteString(fmt.Sprintf("| Info | %d |\n", counts[SeverityInfo]))
		md.WriteString("\n")

		// Problem details
		md.WriteString("### Details\n\n")
		for i, problem := range problems {
			if i >= 50 { // Limit
				break
			}
			md.WriteString(fmt.Sprintf("#### %d. %s\n\n", i+1, problem.Type))
			md.WriteString(fmt.Sprintf("- **Severity**: %s\n", severityToString(problem.Severity)))
			md.WriteString(fmt.Sprintf("- **Location**: `%s`\n", problem.Location))
			md.WriteString(fmt.Sprintf("- **Message**: %s\n", problem.Message))
			if problem.Suggestion != "" {
				md.WriteString(fmt.Sprintf("- **Suggestion**: %s\n", problem.Suggestion))
			}
			md.WriteString("\n")
		}
	}

	// Performance
	if re.performance != nil && re.performance.GetMetrics().FrameCount > 0 {
		md.WriteString("## Performance\n\n")
		metrics := re.performance.GetMetrics()
		md.WriteString(fmt.Sprintf("- **Total Frames**: %d\n", metrics.FrameCount))
		md.WriteString(fmt.Sprintf("- **FPS**: %.1f\n", metrics.FPS))
		md.WriteString(fmt.Sprintf("- **Avg Render Time**: %s\n", formatDuration(metrics.AvgRenderTime)))
		md.WriteString(fmt.Sprintf("- **Memory**: %s\n", formatBytes(metrics.LastHeapAlloc)))
		md.WriteString(fmt.Sprintf("- **GC Count**: %d\n", metrics.NumGC))
		md.WriteString("\n")
	}

	// Footer
	md.WriteString("---\n")
	md.WriteString("*Generated by Mint TUI UI Inspector*\n")

	return md.String()
}

// exportJSON exports report as JSON
func (re *ReportExporter) exportJSON() string {
	// Simplified JSON export (for now)
	var json strings.Builder

	json.WriteString("{\n")
	json.WriteString(fmt.Sprintf("  \"generated\": \"%s\",\n", time.Now().Format(time.RFC3339)))

	// Selected element
	if re.inspector.GetSelectedVNode() != nil {
		selected := re.inspector.GetSelectedInfo()
		json.WriteString("  \"selectedElement\": {\n")
		json.WriteString(fmt.Sprintf("    \"type\": \"%s\",\n", selected.Type))
		json.WriteString(fmt.Sprintf("    \"label\": \"%s\",\n", selected.Label))
		json.WriteString(fmt.Sprintf("    \"position\": {\"x\": %d, \"y\": %d},\n", selected.Position.X, selected.Position.Y))
		json.WriteString(fmt.Sprintf("    \"size\": {\"width\": %d, \"height\": %d}\n", selected.Size.Width, selected.Size.Height))
		json.WriteString("  },\n")
	}

	// Tree stats
	if re.treeView != nil && re.treeView.root != nil {
		stats := re.treeView.GetTreeStats()
		json.WriteString("  \"treeStats\": {\n")
		json.WriteString(fmt.Sprintf("    \"totalNodes\": %d,\n", stats.TotalNodes))
		json.WriteString(fmt.Sprintf("    \"leafNodes\": %d,\n", stats.LeafNodes))
		json.WriteString(fmt.Sprintf("    \"parentNodes\": %d,\n", stats.ParentNodes))
		json.WriteString(fmt.Sprintf("    \"maxDepth\": %d\n", stats.MaxDepth))
		json.WriteString("  },\n")
	}

	// Problems
	if re.diagnostics != nil && len(re.diagnostics.GetProblems()) > 0 {
		problems := re.diagnostics.GetProblems()
		counts := re.diagnostics.CountBySeverity()
		json.WriteString("  \"problems\": {\n")
		json.WriteString(fmt.Sprintf("    \"total\": %d,\n", len(problems)))
		json.WriteString(fmt.Sprintf("    \"critical\": %d,\n", counts[SeverityCritical]))
		json.WriteString(fmt.Sprintf("    \"error\": %d,\n", counts[SeverityError]))
		json.WriteString(fmt.Sprintf("    \"warning\": %d,\n", counts[SeverityWarning]))
		json.WriteString(fmt.Sprintf("    \"info\": %d\n", counts[SeverityInfo]))
		json.WriteString("  },\n")
	}

	// Performance
	if re.performance != nil && re.performance.GetMetrics().FrameCount > 0 {
		metrics := re.performance.GetMetrics()
		json.WriteString("  \"performance\": {\n")
		json.WriteString(fmt.Sprintf("    \"frames\": %d,\n", metrics.FrameCount))
		json.WriteString(fmt.Sprintf("    \"fps\": %.1f,\n", metrics.FPS))
		json.WriteString(fmt.Sprintf("    \"avgRenderTime\": \"%s\",\n", formatDuration(metrics.AvgRenderTime)))
		json.WriteString(fmt.Sprintf("    \"memory\": \"%s\",\n", formatBytes(metrics.LastHeapAlloc)))
		json.WriteString(fmt.Sprintf("    \"gcCount\": %d\n", metrics.NumGC))
		json.WriteString("  }\n")
	} else {
		// Remove trailing comma
		jsonStr := json.String()
		if strings.HasSuffix(jsonStr, ",\n") {
			return jsonStr[:len(jsonStr)-2] + "\n"
		}
	}

	json.WriteString("}\n")
	return json.String()
}

// Helper methods

func (re *ReportExporter) header() string {
	lines := []string{
		"╔════════════════════════════════════════════════════════════╗",
		"║           UI INSPECTOR REPORT                              ║",
		"╠════════════════════════════════════════════════════════════╣",
		fmt.Sprintf("║ Generated: %-45s ║", time.Now().Format("2006-01-02 15:04:05")),
		"╚════════════════════════════════════════════════════════════╝",
	}
	return strings.Join(lines, "\n")
}

func (re *ReportExporter) footer() string {
	return "\n---\nGenerated by Mint TUI UI Inspector"
}

func (re *ReportExporter) selectedElementSection() string {
	lines := []string{
		"┌─ Selected Element ─────────────────────────────────────┐",
		"│                                                        │",
	}

	selected := re.inspector.GetSelectedInfo()
	lines = append(lines, fmt.Sprintf("│ Type: %-50s │", selected.Type))
	lines = append(lines, fmt.Sprintf("│ Label: %-49s │", selected.Label))
	lines = append(lines, fmt.Sprintf("│ Position: (%d, %d)%36s │",
		selected.Position.X, selected.Position.Y, ""))
	lines = append(lines, fmt.Sprintf("│ Size: %dx%d%43s │",
		selected.Size.Width, selected.Size.Height, ""))

	lines = append(lines, "└────────────────────────────────────────────────────────┘")
	return strings.Join(lines, "\n")
}

func (re *ReportExporter) treeViewSection() string {
	stats := re.treeView.GetTreeStats()

	lines := []string{
		"┌─ Layout Tree Summary ──────────────────────────────────┐",
		fmt.Sprintf("│ Total Nodes: %-42d │", stats.TotalNodes),
		fmt.Sprintf("│ Leaf Nodes: %-42d │", stats.LeafNodes),
		fmt.Sprintf("│ Parent Nodes: %-40d │", stats.ParentNodes),
		fmt.Sprintf("│ Max Depth: %-43d │", stats.MaxDepth),
		"└────────────────────────────────────────────────────────┘",
	}

	return strings.Join(lines, "\n")
}

func (re *ReportExporter) diagnosticsSection() string {
	problems := re.diagnostics.GetProblems()
	counts := re.diagnostics.CountBySeverity()

	lines := []string{
		"┌─ Layout Problems ──────────────────────────────────────┐",
		fmt.Sprintf("│ Total: %-47d │", len(problems)),
		fmt.Sprintf("│ Critical: %-43d │", counts[SeverityCritical]),
		fmt.Sprintf("│ Errors: %-45d │", counts[SeverityError]),
		fmt.Sprintf("│ Warnings: %-43d │", counts[SeverityWarning]),
		fmt.Sprintf("│ Info: %-46d │", counts[SeverityInfo]),
		"└────────────────────────────────────────────────────────┘",
	}

	return strings.Join(lines, "\n")
}

func (re *ReportExporter) performanceSection() string {
	metrics := re.performance.GetMetrics()

	lines := []string{
		"┌─ Performance ──────────────────────────────────────────┐",
		fmt.Sprintf("│ Total Frames: %-41d │", metrics.FrameCount),
		fmt.Sprintf("│ FPS: %-49.1f │", metrics.FPS),
		fmt.Sprintf("│ Avg Render: %-42s │", formatDuration(metrics.AvgRenderTime)),
		fmt.Sprintf("│ Memory: %-45s │", formatBytes(metrics.LastHeapAlloc)),
		fmt.Sprintf("│ GC Count: %-43d │", metrics.NumGC),
		"└────────────────────────────────────────────────────────┘",
	}

	return strings.Join(lines, "\n")
}

// ExportToFile writes report to a file (simplified - just returns content)
func (re *ReportExporter) ExportToFile(format ReportFormat, filename string) error {
	content := re.Export(format)
	// In a real implementation, this would write to file
	// For now, just return the content
	_ = filename
	_ = content
	return nil
}

// QuickSummary generates a quick summary report
func (re *ReportExporter) QuickSummary() string {
	var summary strings.Builder

	summary.WriteString("=== UI Inspector Quick Summary ===\n\n")

	// Selected element
	if re.inspector.GetSelectedVNode() != nil {
		selected := re.inspector.GetSelectedInfo()
		summary.WriteString(fmt.Sprintf("Selected: %s (%s)\n", selected.Type, selected.Label))
	}

	// Tree stats
	if re.treeView != nil && re.treeView.root != nil {
		stats := re.treeView.GetTreeStats()
		summary.WriteString(fmt.Sprintf("Tree: %d nodes, depth %d\n", stats.TotalNodes, stats.MaxDepth))
	}

	// Problems
	if re.diagnostics != nil && len(re.diagnostics.GetProblems()) > 0 {
		counts := re.diagnostics.CountBySeverity()
		summary.WriteString(fmt.Sprintf("Problems: %d errors, %d warnings\n",
			counts[SeverityError]+counts[SeverityCritical], counts[SeverityWarning]))
	}

	// Performance
	if re.performance != nil && re.performance.GetMetrics().FrameCount > 0 {
		metrics := re.performance.GetMetrics()
		summary.WriteString(fmt.Sprintf("Performance: %.1f FPS, %s memory\n",
			metrics.FPS, formatBytes(metrics.LastHeapAlloc)))
	}

	return summary.String()
}
