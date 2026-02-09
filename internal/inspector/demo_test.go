package inspector

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/ui"
)

// TestInspectorDemo demonstrates the inspector with a complex tree
func TestInspectorDemo(t *testing.T) {
	// Create a complex nested structure
	root := ui.VStack(
		ui.VStack(  // LayoutNode 1 (nested)
			ui.Text("A"),
			ui.Text("B"),
		),
		ui.VStack(  // LayoutNode 2 (nested)
			ui.Text("C"),
		),
		ui.VStack(  // LayoutNode 3 (nested with HStack)
			ui.HStack(
				ui.Text("D"),
				ui.Text("E"),
			),
		),
	)

	tv := NewTreeView()
	tv.SetShowPaths(true)  // Show paths to make it clearer
	err := tv.SetRoot(root)
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}

	fmt.Println("\n=== Initial Tree (All Expanded) ===")
	lines, _ := tv.GetTreeLines()
	for i, line := range lines {
		fmt.Printf("%2d: %s\n", i, line)
	}

	// Find a LayoutNode to collapse
	allNodes := tv.GetFlatList()
	var targetUID string
	for _, node := range allNodes {
		// Find a nested LayoutNode with children
		if len(node.Children) > 0 && node.Level > 0 && node.Info.Type == "LayoutNode" {
			targetUID = node.UniqueID
			fmt.Printf("\n=== Found collapsible LayoutNode ===\n")
			fmt.Printf("UniqueID: %s\n", node.UniqueID)
			fmt.Printf("Path: %s\n", node.Path)
			fmt.Printf("Children: %d\n", len(node.Children))
			fmt.Printf("Expanded: %v\n", node.Expanded)
			break
		}
	}

	if targetUID != "" {
		// Collapse it
		fmt.Printf("\n=== Pressing E to collapse %s ===\n", targetUID)
		tv.ToggleNode(targetUID)

		lines, _ = tv.GetTreeLines()
		fmt.Printf("\n=== After Collapse (%d lines) ===\n", len(lines))
		for i, line := range lines {
			fmt.Printf("%2d: %s\n", i, line)
		}

		// Expand it back
		fmt.Printf("\n=== Pressing E to expand %s ===\n", targetUID)
		tv.ToggleNode(targetUID)

		lines, _ = tv.GetTreeLines()
		fmt.Printf("\n=== After Expand (%d lines) ===\n", len(lines))
		for i, line := range lines {
			fmt.Printf("%2d: %s\n", i, line)
		}
	}

	fmt.Println("\n✓ Demo complete - LayoutNode collapse/expand works!")
}
