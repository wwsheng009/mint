// Layout Diagnostic Tool
// A comprehensive tool to analyze constraint propagation through the VNode tree
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/components/display"
	"github.com/wwsheng009/mint/components/navigation"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

type DiagnosticResult struct {
	Indent      string
	Depth       int
	NodeType    string
	Tag         string
	Constraints string
	Size        string
	Props       map[string]interface{}
	HasMeasure  bool
	MeasuredSize string
	Children    int
	Propagated  string
	Issues      []string
}

type LayoutDiagnostic struct {
	engine      *compute.Engine
	results     []*DiagnosticResult
	maxDepth    int
	showDetails bool
}

func NewLayoutDiagnostic() *LayoutDiagnostic {
	return &LayoutDiagnostic{
		engine:      compute.NewEngine(),
		maxDepth:    20,
		showDetails: true,
	}
}

func (ld *LayoutDiagnostic) AnalyzeVNode(vnode rtui.VNode, parentConstraints runtime.BoxConstraints) []*DiagnosticResult {
	ld.results = []*DiagnosticResult{}
	ld.analyzeNode(vnode, parentConstraints, 0, "")
	return ld.results
}

func (ld *LayoutDiagnostic) analyzeNode(node rtui.VNode, constraints runtime.BoxConstraints, depth int, path string) {
	if depth > ld.maxDepth {
		return
	}

	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

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

		// Check Props for explicit width/height (only if we have a measured size)
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
		// Determine constraints to pass to child
		childConstraints := ld.determineChildConstraints(node, child, constraints, depth, i, len(children))
		childPath := fmt.Sprintf("%s[%d]", path, i)
		ld.analyzeNode(child, childConstraints, depth+1, childPath)
	}
}

func (ld *LayoutDiagnostic) determineChildConstraints(parent, child rtui.VNode, parentConstraints runtime.BoxConstraints, depth int, childIndex, totalChildren int) runtime.BoxConstraints {
	// This simulates how different container types pass constraints to children
	// For now, just propagate parent constraints (can be enhanced later)
	return parentConstraints
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

func (ld *LayoutDiagnostic) formatConstraints(constraints runtime.BoxConstraints) string {
	return fmt.Sprintf("W[%d:%d] H[%d:%d]",
		constraints.MinWidth, constraints.MaxWidth,
		constraints.MinHeight, constraints.MaxHeight)
}

func (ld *LayoutDiagnostic) PrintResults() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("LAYOUT DIAGNOSTIC REPORT")
	fmt.Println(strings.Repeat("=", 80))

	for _, result := range ld.results {
		fmt.Printf("\n%s[%d] %s", result.Indent, result.Depth, result.NodeType)
		if result.Tag != "" {
			fmt.Printf(" (tag: %s)", result.Tag)
		}
		fmt.Println()

		fmt.Printf("%s  Constraints: %s\n", result.Indent, result.Constraints)

		if result.HasMeasure {
			fmt.Printf("%s  Measured:   %s\n", result.Indent, result.MeasuredSize)
		} else {
			fmt.Printf("%s  Measured:   %s\n", result.Indent, result.MeasuredSize)
		}

		if result.Propagated != "" {
			fmt.Printf("%s  Propagated: %s\n", result.Indent, result.Propagated)
		}

		if result.Props != nil && len(result.Props) > 0 {
			fmt.Printf("%s  Props: ", result.Indent)
			first := true
			for k, v := range result.Props {
				if !first {
					fmt.Print(", ")
				}
				fmt.Printf("%s=%v", k, v)
				first = false
			}
			fmt.Println()
		}

		if result.Children > 0 {
			fmt.Printf("%s  Children: %d\n", result.Indent, result.Children)
		}

		if len(result.Issues) > 0 {
			fmt.Printf("%s  Issues:\n", result.Indent)
			for _, issue := range result.Issues {
				fmt.Printf("%s    %s\n", result.Indent, issue)
			}
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	ld.printSummary()
}

func (ld *LayoutDiagnostic) printSummary() {
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

	fmt.Printf("Summary:\n")
	fmt.Printf("  Total nodes analyzed: %d\n", totalNodes)
	fmt.Printf("  Nodes with Measure(): %d\n", nodesWithMeasure)
	fmt.Printf("  Nodes with issues: %d\n", nodesWithIssues)
	fmt.Printf("  Nodes with Height prop: %d\n", nodesWithHeightProp)
	fmt.Printf("  Nodes with Width prop: %d\n", nodesWithWidthProp)

	if nodesWithIssues > 0 {
		fmt.Printf("\n⚠️  Found %d nodes with constraint issues\n", nodesWithIssues)
	} else {
		fmt.Printf("\n✅ No constraint issues found!\n")
	}
}

func main() {
	os.Setenv("TUI_DEBUG_LAYOUT", "false")
	os.Setenv("TUI_DEBUG_INSPECTOR", "false")

	fmt.Println("Layout Diagnostic Tool")
	fmt.Println("===================")

	// Test 1: Simple VStack with height constraint
	fmt.Println("\n[Test 1] VStack with Height(10) prop")
	fmt.Println("---------------------------------------")
	vstack := ui.VStackBuilder(
		ui.Text("Line 1"),
		ui.Text("Line 2"),
		ui.Text("Line 3"),
		ui.Text("Line 4"),
		ui.Text("Line 5"),
	).Height(10).Build()

	diagnostic := NewLayoutDiagnostic()
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 10,
	}

	_ = diagnostic.AnalyzeVNode(vstack, constraints)
	diagnostic.PrintResults()

	// Test 2: TreeView with bounded height
	fmt.Println("\n[Test 2] TreeView with bounded height constraint")
	fmt.Println("----------------------------------------------")
	lines := []string{
		"Root",
		"├── Child 1",
		"│   ├── Grandchild 1.1",
		"│   └── Grandchild 1.2",
		"├── Child 2",
		"├── Child 3",
		"├── Child 4",
		"├── Child 5",
		"├── Child 6",
		"├── Child 7",
		"└── Child 8",
	}

	treeView := display.NewTreeView().
		FromLines(lines).
		ExpandLevel(1).
		Build()

	diagnostic2 := NewLayoutDiagnostic()
	constraints2 := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth: 80,
		MinHeight: 0,
		MaxHeight: 8,
	}

	_ = diagnostic2.AnalyzeVNode(treeView, constraints2)
	diagnostic2.PrintResults()

	// Test 3: Tabs with height constraint
	fmt.Println("\n[Test 3] Tabs with Height(15) prop")
	fmt.Println("------------------------------------")
	tabs := navigation.TabsBuilder().
		AddTab("tab1", "Tab 1").
		Content("tab1", ui.VStack(
			ui.Text("Content Line 1"),
			ui.Text("Content Line 2"),
			ui.Text("Content Line 3"),
			ui.Text("Content Line 4"),
			ui.Text("Content Line 5"),
			ui.Text("Content Line 6"),
			ui.Text("Content Line 7"),
			ui.Text("Content Line 8"),
			ui.Text("Content Line 9"),
			ui.Text("Content Line 10"),
		)).
		AddTab("tab2", "Tab 2").
		Content("tab2", ui.Text("Short content")).
		Height(15).
		Build()

	diagnostic3 := NewLayoutDiagnostic()
	constraints3 := runtime.BoxConstraints{
		MinWidth:   0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 15,
	}

	_ = diagnostic3.AnalyzeVNode(tabs, constraints3)
	diagnostic3.PrintResults()

	// Test 4: Simulated Inspector structure
	fmt.Println("\n[Test 4] Simulated Inspector Structure")
	fmt.Println("------------------------------------")
	inspVStack := ui.VStackBuilder(
		ui.Text("📦 Layout Tree"),
		ui.Text("Nodes: 32 | Depth: 4"),
		ui.Text(""),
		ui.Text("────────────────────────"),
		ui.Text("Focused: Element"),
		ui.Text("Path: vstack"),
		ui.Text(""),
		// Simulate TreeView with many lines
		ui.Text("Root"),
		ui.Text("├── Child 1"),
		ui.Text("│   ├── Grandchild 1.1"),
		ui.Text("│   └── Grandchild 1.2"),
		ui.Text("├── Child 2"),
		ui.Text("├── Child 3"),
		ui.Text("├── Child 4"),
		ui.Text("├── Child 5"),
		ui.Text("├── Child 6"),
		ui.Text("├── Child 7"),
		ui.Text("├── Child 8"),
		ui.Text("└── Child 9"),
		ui.Text(""),
		ui.Text("Instructions: Use arrow keys to navigate"),
	).
		Width(76).
		Height(20).
		Build()

	diagnostic4 := NewLayoutDiagnostic()
	constraints4 := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth: 76,
		MinHeight: 0,
		MaxHeight: 20,
	}

	_ = diagnostic4.AnalyzeVNode(inspVStack, constraints4)
	diagnostic4.PrintResults()

	// Test 5: Inspector-like VStack with TreeView (the exact scenario)
	fmt.Println("\n[Test 5] Inspector VStack with TreeView (Virtual Scrolling Test)")
	fmt.Println("---------------------------------------------------------------")

	// Create TreeView with many lines (like Inspector Elements tab)
	inspLines := make([]string, 34)
	for i := 0; i < 34; i++ {
		if i == 0 {
			inspLines[i] = "Root Element"
		} else {
			inspLines[i] = fmt.Sprintf("├── Child Element %d", i)
		}
	}

	inspTreeView := display.NewTreeView().
		FromLines(inspLines).
		ExpandLevel(1).
		ShowIcons(true).
		Build()

	// Wrap in VStack with Height constraint (like Inspector does)
	inspLikeVStack := ui.VStackBuilder(
		ui.Text("📦 Elements"),
		ui.Text(""),
		inspTreeView,
		ui.Text(""),
		ui.Text("Instructions: ↑↓ to navigate"),
	).
		Width(76).
		Height(20).
		Build()

	diagnostic5 := NewLayoutDiagnostic()
	constraints5 := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth: 76,
		MinHeight: 0,
		MaxHeight: 20,
	}

	_ = diagnostic5.AnalyzeVNode(inspLikeVStack, constraints5)
	diagnostic5.PrintResults()

	// Test 6: Verify TreeView virtual scrolling with Layout Engine
	fmt.Println("\n[Test 6] TreeView Virtual Scrolling with Layout Engine")
	fmt.Println("-------------------------------------------------------")

	engine := compute.NewEngine()
	// Phase 3: Pass nil for Fiber (non-Fiber mode, backward compatible)
	layout, err := engine.Layout(inspTreeView, nil, runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  76,
		MinHeight: 0,
		MaxHeight: 18, // Available height for TreeView
	})

	if err != nil {
		fmt.Printf("\n❌ Layout failed: %v\n", err)
	} else {
		fmt.Printf("\n✅ Layout complete: TreeView size = %dx%d\n", layout.Root.Box.Width, layout.Root.Box.Height)

		// Check how many children are rendered
		tvChildren := layout.Root.GetVNode().Children()
		fmt.Printf("   Total lines in TreeView: %d\n", len(inspLines))
		fmt.Printf("   Direct children of TreeView: %d\n", len(tvChildren))

		// The TreeView has 1 child (the VStack), so check that
		if len(tvChildren) > 0 {
			vstackChildren := tvChildren[0].Children()
			fmt.Printf("   Children in TreeView's VStack: %d\n", len(vstackChildren))

			if len(vstackChildren) <= 20 {
				fmt.Printf("   ✅ Virtual scrolling WORKING! Only rendering %d visible lines\n", len(vstackChildren))
			} else {
				fmt.Printf("   ❌ Virtual scrolling NOT working! Rendering all %d lines instead of ~20\n", len(vstackChildren))
			}
		}
	}

	// Test 7: UpdateLines() preserves virtual scrolling
	fmt.Println("\n[Test 7] TreeView UpdateLines() Virtual Scrolling Preservation")
	fmt.Println("----------------------------------------------------------------")

	// Get the TreeView as *display.TreeView to call UpdateLines
	tvPtr, ok := inspTreeView.(*display.TreeView)
	if !ok {
		fmt.Printf("❌ Could not convert to *display.TreeView\n")
	} else {
		// Update with new lines (more than before)
		newInspLines := make([]string, 50)
		for i := 0; i < 50; i++ {
			if i == 0 {
				newInspLines[i] = "Updated Root Element"
			} else {
				newInspLines[i] = fmt.Sprintf("├── Updated Child %d", i)
			}
		}

		fmt.Printf("Before UpdateLines():\n")
		fmt.Printf("  Total lines: %d\n", len(inspLines))
		fmt.Printf("  viewportHeight: (need to check via reflection)\n")
		tvChildren1 := tvPtr.Children()
		fmt.Printf("  Children rendered: %d\n", len(tvChildren1))

		// Call UpdateLines() - this should preserve viewportHeight
		tvPtr.UpdateLines(newInspLines)

		fmt.Printf("\nAfter UpdateLines():\n")
		fmt.Printf("  Total lines: %d\n", len(newInspLines))

		// Measure again with same constraints
		engine2 := compute.NewEngine()
		// Phase 3: Pass nil for Fiber (non-Fiber mode, backward compatible)
		layout2, err := engine2.Layout(tvPtr, nil, runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth:  76,
			MinHeight: 0,
			MaxHeight: 18,
		})

		if err != nil {
			fmt.Printf("  ❌ Layout failed after UpdateLines(): %v\n", err)
		} else {
			fmt.Printf("  ✅ Layout complete: %dx%d\n", layout2.Root.Box.Width, layout2.Root.Box.Height)

			tvChildren2 := tvPtr.Children()
			fmt.Printf("  Children rendered after UpdateLines(): %d\n", len(tvChildren2))

			// Check the VStack child
			if len(tvChildren2) > 0 {
				vstackChildren2 := tvChildren2[0].Children()
				fmt.Printf("  Children in VStack: %d\n", len(vstackChildren2))

				if len(vstackChildren2) <= 20 {
					fmt.Printf("  ✅ Virtual scrolling PRESERVED after UpdateLines()!\n")
				} else {
					fmt.Printf("  ❌ Virtual scrolling BROKEN after UpdateLines()! Rendering %d lines\n", len(vstackChildren2))
				}
			}
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("DIAGNOSTIC COMPLETE")
	fmt.Println(strings.Repeat("=", 80))
}

func (ld *LayoutDiagnostic) AnalyzeComputedLayout(box *compute.ComputedBox, constraints runtime.BoxConstraints) []*DiagnosticResult {
	ld.results = []*DiagnosticResult{}
	ld.analyzeComputedBox(box, constraints, 0, "")
	ld.PrintResults()
	return ld.results
}

func (ld *LayoutDiagnostic) analyzeComputedBox(box *compute.ComputedBox, constraints runtime.BoxConstraints, depth int, path string) {
	if depth > ld.maxDepth {
		return
	}

	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

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
		// Determine constraints for child
		childConstraints := ld.determineChildConstraintsFromLayout(box, child, i)
		childPath := fmt.Sprintf("%s[%d]", path, i)
		ld.analyzeComputedBox(child, childConstraints, depth+1, childPath)
	}
}

func (ld *LayoutDiagnostic) determineChildConstraintsFromLayout(parent *compute.ComputedBox, child *compute.ComputedBox, childIndex int) runtime.BoxConstraints {
	// Try to infer what constraints were passed to the child
	// This is heuristic but can help identify issues

	// For now, just return the parent's constraints
	// A more sophisticated approach would track the actual constraints used
	return runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth: runtime.Infinity,
		MinHeight: 0,
		MaxHeight: runtime.Infinity,
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
				// This is a TreeView
				if actualSize.Height > 20 {
					result.Issues = append(result.Issues,
						fmt.Sprintf("ℹ️  TreeView is large (%dx%d) - should use virtual scrolling", actualSize.Width, actualSize.Height))
				}
			}
		}
	}

	// Check for suspicious VStack behavior
	if result.Tag == "vstack" && result.Children > 3 {
		// Check children sizes
		totalChildrenHeight := 0
		for _, child := range box.Children {
			totalChildrenHeight += child.Box.Height
		}

		if actualSize.Height > 0 && totalChildrenHeight > actualSize.Height && actualSize.Height < 100 {
			// Children don't fit - should have virtual scrolling or clipping
			result.Issues = append(result.Issues,
				fmt.Sprintf("⚠️  VStack content (%d lines) exceeds container height (%d) - virtual scrolling needed?",
					totalChildrenHeight, actualSize.Height))
		}
	}

	// Check if VStack children have bounded height
	if result.Tag == "vstack" && len(box.Children) > 0 {
		for _, child := range box.Children {
			// Check if child has TreeView
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
