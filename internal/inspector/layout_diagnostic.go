// Package inspector provides layout diagnostic capabilities for the UI Inspector
package inspector

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// DiagnosticResult represents the diagnostic information for a single node
type DiagnosticResult struct {
	Indent       string
	Depth        int
	NodeType     string
	Tag          string
	Constraints  string
	Size         string
	Props        map[string]interface{}
	HasMeasure   bool
	MeasuredSize string
	Children     int
	Propagated   string
	Issues       []string
	VisibleLines int // For TreeView: number of visible lines vs total lines
	TotalLines   int // For TreeView: total lines
}

// LayoutDiagnostic provides comprehensive layout analysis for VNode trees
type LayoutDiagnostic struct {
	engine      *compute.Engine
	results     []*DiagnosticResult
	maxDepth    int
	showDetails bool
}

// NewLayoutDiagnostic creates a new layout diagnostic instance
func NewLayoutDiagnostic() *LayoutDiagnostic {
	return &LayoutDiagnostic{
		engine:      compute.NewEngine(),
		maxDepth:    20,
		showDetails: true,
	}
}

// AnalyzeVNode analyzes a VNode tree and returns diagnostic results
func (ld *LayoutDiagnostic) AnalyzeVNode(vnode rtui.VNode, parentConstraints runtime.BoxConstraints) []*DiagnosticResult {
	ld.results = []*DiagnosticResult{}
	ld.analyzeNode(vnode, parentConstraints, 0, "")
	return ld.results
}

// AnalyzeSelectedNode analyzes a single selected node with its constraint chain
func (ld *LayoutDiagnostic) AnalyzeSelectedNode(vnode rtui.VNode, constraints runtime.BoxConstraints) *DiagnosticResult {
	if vnode == nil {
		return nil
	}

	result := &DiagnosticResult{
		Depth:       0,
		NodeType:    vnode.Type().String(),
		Constraints: ld.formatConstraints(constraints),
		Props:       vnode.Props(),
	}

	// Get Tag if available
	if tagger, ok := vnode.(interface{ Tag() string }); ok {
		result.Tag = tagger.Tag()
	}

	// Check if node implements Measure
	if measurable, ok := vnode.(interface{ Measure(runtime.BoxConstraints) runtime.Size }); ok {
		result.HasMeasure = true
		size := measurable.Measure(constraints)
		result.MeasuredSize = fmt.Sprintf("%dx%d", size.Width, size.Height)
		result.Size = fmt.Sprintf("%dx%d", size.Width, size.Height)

		// Check for issues
		ld.checkForIssues(vnode, constraints, size, result)

		// Check Props for explicit width/height
		if result.Props != nil {
			if h, ok := result.Props["height"].(int); ok && h > 0 {
				if size.Height != h {
					result.Issues = append(result.Issues,
						fmt.Sprintf("⚠️  Has Height(%d) prop but measured size is %dx%d", h, size.Width, size.Height))
				} else {
					result.Propagated = fmt.Sprintf("Height(%d) → size", h)
				}
			}
			if w, ok := result.Props["width"].(int); ok && w > 0 {
				if size.Width != w {
					result.Issues = append(result.Issues,
						fmt.Sprintf("⚠️  Has Width(%d) prop but measured size is %dx%d", w, size.Width, size.Height))
				}
			}
		}

		// Check for TreeView specific info
		if result.Tag == "treeview" {
			children := vnode.Children()
			result.Children = len(children)
			if len(children) > 0 {
				// The first child is the VStack containing the actual lines
				vstackChildren := children[0].Children()
				result.VisibleLines = len(vstackChildren)
				// Try to get total lines from props
				if result.Props != nil {
					if totalLines, ok := result.Props["totalLines"].(int); ok {
						result.TotalLines = totalLines
					}
				}
			}
		}
	} else {
		result.MeasuredSize = "N/A (no Measure method)"
		result.Issues = append(result.Issues, "⚠️  Node does not implement Measure()")
	}

	// Count children
	children := vnode.Children()
	result.Children = len(children)

	return result
}

// AnalyzeComputedLayout analyzes a computed layout tree
func (ld *LayoutDiagnostic) AnalyzeComputedLayout(box *compute.ComputedBox, constraints runtime.BoxConstraints) []*DiagnosticResult {
	ld.results = []*DiagnosticResult{}
	ld.analyzeComputedBox(box, constraints, 0, "")
	return ld.results
}

func (ld *LayoutDiagnostic) analyzeNode(node rtui.VNode, constraints runtime.BoxConstraints, depth int, path string) {
	if depth > ld.maxDepth {
		return
	}

	indent := strings.Repeat("  ", depth)

	result := &DiagnosticResult{
		Indent:      indent,
		Depth:       depth,
		NodeType:    node.Type().String(),
		Constraints: ld.formatConstraints(constraints),
		Props:       node.Props(),
	}

	// Get Tag if available
	if tagger, ok := node.(interface{ Tag() string }); ok {
		result.Tag = tagger.Tag()
	}

	// Check if node implements Measure
	if measurable, ok := node.(interface{ Measure(runtime.BoxConstraints) runtime.Size }); ok {
		result.HasMeasure = true
		size := measurable.Measure(constraints)
		result.MeasuredSize = fmt.Sprintf("%dx%d", size.Width, size.Height)
		result.Size = fmt.Sprintf("%dx%d", size.Width, size.Height)

		// Check for issues
		ld.checkForIssues(node, constraints, size, result)

		// Check Props for explicit width/height
		if result.Props != nil {
			if h, ok := result.Props["height"].(int); ok && h > 0 {
				if size.Height != h {
					result.Issues = append(result.Issues,
						fmt.Sprintf("⚠️  Has Height(%d) prop but measured size is %dx%d", h, size.Width, size.Height))
				} else {
					result.Propagated = fmt.Sprintf("Height(%d) → constraints", h)
				}
			}
			if w, ok := result.Props["width"].(int); ok && w > 0 {
				if size.Width != w {
					result.Issues = append(result.Issues,
						fmt.Sprintf("⚠️  Has Width(%d) prop but measured size is %dx%d", w, size.Width, size.Height))
				}
			}
		}

		// Check for TreeView specific info
		if result.Tag == "treeview" {
			children := node.Children()
			result.Children = len(children)
			if len(children) > 0 {
				vstackChildren := children[0].Children()
				result.VisibleLines = len(vstackChildren)
				if result.Props != nil {
					if totalLines, ok := result.Props["totalLines"].(int); ok {
						result.TotalLines = totalLines
					}
				}
			}
		}
	} else {
		result.MeasuredSize = "N/A (no Measure method)"
		result.Issues = append(result.Issues, "⚠️  Node does not implement Measure()")
	}

	// Count children
	children := node.Children()
	result.Children = len(children)

	ld.results = append(ld.results, result)

	// Recursively analyze children
	for i, child := range children {
		childConstraints := ld.determineChildConstraints(node, child, constraints)
		childPath := fmt.Sprintf("%s[%d]", path, i)
		ld.analyzeNode(child, childConstraints, depth+1, childPath)
	}
}

func (ld *LayoutDiagnostic) determineChildConstraints(parent, child rtui.VNode, parentConstraints runtime.BoxConstraints) runtime.BoxConstraints {
	// For now, just propagate parent constraints
	// This can be enhanced to simulate actual constraint propagation
	return parentConstraints
}

func (ld *LayoutDiagnostic) analyzeComputedBox(box *compute.ComputedBox, constraints runtime.BoxConstraints, depth int, path string) {
	if depth > ld.maxDepth {
		return
	}

	indent := strings.Repeat("  ", depth)

	result := &DiagnosticResult{
		Indent:      indent,
		Depth:       depth,
		NodeType:    box.GetVNode().Type().String(),
		Constraints: ld.formatConstraints(constraints),
		Props:       box.GetVNode().Props(),
		Size:        fmt.Sprintf("%dx%d", box.Box.Width, box.Box.Height),
		Children:    len(box.Children),
	}

	// Get Tag
	if tagger, ok := box.GetVNode().(interface{ Tag() string }); ok {
		result.Tag = tagger.Tag()
	}

	// Get actual size from layout
	actualSize := box.Box
	result.MeasuredSize = fmt.Sprintf("%dx%d (actual)", actualSize.Width, actualSize.Height)

	// Check for constraint violations
	ld.checkComputedIssues(box, constraints, result)

	ld.results = append(ld.results, result)

	// Analyze children
	for i, child := range box.Children {
		childConstraints := ld.determineChildConstraintsFromLayout(box, child)
		childPath := fmt.Sprintf("%s[%d]", path, i)
		ld.analyzeComputedBox(child, childConstraints, depth+1, childPath)
	}
}

func (ld *LayoutDiagnostic) determineChildConstraintsFromLayout(parent *compute.ComputedBox, child *compute.ComputedBox) runtime.BoxConstraints {
	// Try to infer what constraints were passed to the child
	return runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth: runtime.Infinity,
		MinHeight: 0,
		MaxHeight: runtime.Infinity,
	}
}

func (ld *LayoutDiagnostic) checkForIssues(node rtui.VNode, constraints runtime.BoxConstraints, size runtime.Size, result *DiagnosticResult) {
	// Check if bounded constraints are respected
	if constraints.HasBoundedHeight() && size.Height > constraints.MaxHeight {
		result.Issues = append(result.Issues,
			fmt.Sprintf("❌ Height %d exceeds MaxHeight %d", size.Height, constraints.MaxHeight))
	}
	if constraints.HasBoundedWidth() && size.Width > constraints.MaxWidth {
		result.Issues = append(result.Issues,
			fmt.Sprintf("❌ Width %d exceeds MaxWidth %d", size.Width, constraints.MaxWidth))
	}

	// Check for suspicious patterns
	tag := result.Tag
	nodeType := result.NodeType

	// VStack/HStack with many children but unbounded constraints
	if (tag == "vstack" || tag == "hstack") && result.Children > 5 {
		if !constraints.HasBoundedHeight() && tag == "vstack" {
			result.Issues = append(result.Issues,
				"⚠️  VStack with many children but no bounded height constraint")
		}
		if !constraints.HasBoundedWidth() && tag == "hstack" {
			result.Issues = append(result.Issues,
				"⚠️  HStack with many children but no bounded width constraint")
		}
	}

	// TextVNode or Element with large content
	if nodeType == "text" || nodeType == "element" {
		if size.Height > 20 {
			result.Issues = append(result.Issues,
				fmt.Sprintf("ℹ️  Large content: %dx%d (may need virtual scrolling)", size.Width, size.Height))
		}
	}
}

func (ld *LayoutDiagnostic) checkComputedIssues(box *compute.ComputedBox, constraints runtime.BoxConstraints, result *DiagnosticResult) {
	actualSize := box.Box

	// Check size constraint violations
	if constraints.HasBoundedWidth() && actualSize.Width > constraints.MaxWidth {
		result.Issues = append(result.Issues,
			fmt.Sprintf("❌ Width %d exceeds MaxWidth %d", actualSize.Width, constraints.MaxWidth))
	}
	if constraints.HasBoundedHeight() && actualSize.Height > constraints.MaxHeight {
		result.Issues = append(result.Issues,
			fmt.Sprintf("❌ Height %d exceeds MaxHeight %d", actualSize.Height, constraints.MaxHeight))
	}

	// Check if node has explicit size props but doesn't match
	if result.Props != nil {
		if h, ok := result.Props["height"].(int); ok && h > 0 {
			if actualSize.Height != h {
				result.Issues = append(result.Issues,
					fmt.Sprintf("⚠️  Has Height(%d) prop but actual size is %dx%d (prop not applied?)", h, actualSize.Width, actualSize.Height))
			} else {
				result.Propagated = fmt.Sprintf("Height(%d) → size", h)
			}
		}
		if w, ok := result.Props["width"].(int); ok && w > 0 {
			if actualSize.Width != w {
				result.Issues = append(result.Issues,
					fmt.Sprintf("⚠️  Has Width(%d) prop but actual size is %dx%d (prop not applied?)", w, actualSize.Width, actualSize.Height))
			}
		}
	}

	// Check for TreeView specific issues
	if result.NodeType == "element" {
		if result.Props != nil {
			if _, hasTreeView := result.Props["treeView"]; hasTreeView {
				if actualSize.Height > 20 {
					result.Issues = append(result.Issues,
						fmt.Sprintf("ℹ️  TreeView is large (%dx%d) - should use virtual scrolling", actualSize.Width, actualSize.Height))
				}
			}
		}
	}

	// Check for suspicious VStack behavior
	if result.Tag == "vstack" && result.Children > 3 {
		totalChildrenHeight := 0
		for _, child := range box.Children {
			totalChildrenHeight += child.Box.Height
		}

		if actualSize.Height > 0 && totalChildrenHeight > actualSize.Height && actualSize.Height < 100 {
			result.Issues = append(result.Issues,
				fmt.Sprintf("⚠️  VStack content (%d lines) exceeds container height (%d) - virtual scrolling needed?",
					totalChildrenHeight, actualSize.Height))
		}
	}

	// Check if VStack children have bounded height
	if result.Tag == "vstack" && len(box.Children) > 0 {
		for _, child := range box.Children {
			if childProps := child.GetVNode().Props(); childProps != nil {
				if _, hasTreeView := childProps["treeView"]; hasTreeView {
					childHeight := child.Box.Height
					if childHeight > 15 {
						result.Issues = append(result.Issues,
							fmt.Sprintf("⚠️  VStack contains TreeView with height %d (should use virtual scrolling)", childHeight))
					}
				}
			}
		}
	}
}

func (ld *LayoutDiagnostic) formatConstraints(constraints runtime.BoxConstraints) string {
	return fmt.Sprintf("W[%d:%d] H[%d:%d]",
		constraints.MinWidth, constraints.MaxWidth,
		constraints.MinHeight, constraints.MaxHeight)
}

// FormatAsTree formats diagnostic results as a tree view
func (ld *LayoutDiagnostic) FormatAsTree() string {
	var sb strings.Builder

	sb.WriteString("\n" + strings.Repeat("═", 76) + "\n")
	sb.WriteString("📐 LAYOUT DIAGNOSTIC REPORT\n")
	sb.WriteString(strings.Repeat("═", 76) + "\n")

	for _, result := range ld.results {
		sb.WriteString(ld.formatResult(result))
	}

	sb.WriteString("\n" + strings.Repeat("═", 76) + "\n")
	ld.printSummaryToBuilder(&sb)
	sb.WriteString(strings.Repeat("═", 76) + "\n")

	return sb.String()
}

// FormatSingleResult formats a single diagnostic result for the selected node
func (ld *LayoutDiagnostic) FormatSingleResult(result *DiagnosticResult) string {
	if result == nil {
		return "No node selected\n"
	}

	var sb strings.Builder

	sb.WriteString("\n" + strings.Repeat("═", 76) + "\n")
	sb.WriteString("📐 SELECTED NODE LAYOUT INFO\n")
	sb.WriteString(strings.Repeat("═", 76) + "\n")

	sb.WriteString(ld.formatResult(result))

	if result.TotalLines > 0 && result.VisibleLines > 0 {
		sb.WriteString(fmt.Sprintf("  Virtual Scrolling: %d/%d lines rendered", result.VisibleLines, result.TotalLines))
		if result.VisibleLines < result.TotalLines {
			sb.WriteString(" ✅\n")
		} else {
			sb.WriteString(" ⚠️  (all lines rendered)\n")
		}
	}

	sb.WriteString(strings.Repeat("═", 76) + "\n")

	return sb.String()
}

func (ld *LayoutDiagnostic) formatResult(result *DiagnosticResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\n[%d] %s", result.Depth, result.NodeType))
	if result.Tag != "" {
		sb.WriteString(fmt.Sprintf(" (tag: %s)", result.Tag))
	}
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("  Constraints: %s\n", result.Constraints))

	if result.HasMeasure {
		sb.WriteString(fmt.Sprintf("  Measured:   %s\n", result.MeasuredSize))
	} else {
		sb.WriteString(fmt.Sprintf("  Measured:   %s\n", result.MeasuredSize))
	}

	if result.Propagated != "" {
		sb.WriteString(fmt.Sprintf("  Propagated: %s\n", result.Propagated))
	}

	if result.Props != nil && len(result.Props) > 0 {
		sb.WriteString("  Props: ")
		first := true
		for k, v := range result.Props {
			if !first {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%s=%v", k, v))
			first = false
		}
		sb.WriteString("\n")
	}

	if result.Children > 0 {
		sb.WriteString(fmt.Sprintf("  Children: %d\n", result.Children))
	}

	if len(result.Issues) > 0 {
		sb.WriteString("  Issues:\n")
		for _, issue := range result.Issues {
			sb.WriteString(fmt.Sprintf("    %s\n", issue))
		}
	}

	return sb.String()
}

func (ld *LayoutDiagnostic) printSummaryToBuilder(sb *strings.Builder) {
	totalNodes := len(ld.results)
	nodesWithMeasure := 0
	nodesWithIssues := 0
	nodesWithHeightProp := 0
	nodesWithWidthProp := 0

	for _, result := range ld.results {
		if result.HasMeasure {
			nodesWithMeasure++
		}
		if len(result.Issues) > 0 {
			nodesWithIssues++
		}
		if result.Props != nil {
			if _, ok := result.Props["height"].(int); ok {
				nodesWithHeightProp++
			}
			if _, ok := result.Props["width"].(int); ok {
				nodesWithWidthProp++
			}
		}
	}

	sb.WriteString("Summary:\n")
	sb.WriteString(fmt.Sprintf("  Total nodes analyzed: %d\n", totalNodes))
	sb.WriteString(fmt.Sprintf("  Nodes with Measure(): %d\n", nodesWithMeasure))
	sb.WriteString(fmt.Sprintf("  Nodes with issues: %d\n", nodesWithIssues))
	sb.WriteString(fmt.Sprintf("  Nodes with Height prop: %d\n", nodesWithHeightProp))
	sb.WriteString(fmt.Sprintf("  Nodes with Width prop: %d\n", nodesWithWidthProp))

	if nodesWithIssues > 0 {
		sb.WriteString(fmt.Sprintf("\n⚠️  Found %d nodes with constraint issues\n", nodesWithIssues))
	} else {
		sb.WriteString("\n✅ No constraint issues found!\n")
	}
}

// GetSummary returns a summary of diagnostic results
func (ld *LayoutDiagnostic) GetSummary() (totalNodes, nodesWithIssues int) {
	totalNodes = len(ld.results)
	nodesWithIssues = 0

	for _, result := range ld.results {
		if len(result.Issues) > 0 {
			nodesWithIssues++
		}
	}

	return
}
