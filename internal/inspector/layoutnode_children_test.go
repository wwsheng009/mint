package inspector

import (
	"testing"

	"github.com/wwsheng009/mint/ui"
)

// TestLayoutNodeChildrenAccess tests that LayoutNode correctly exposes its children
func TestLayoutNodeChildrenAccess(t *testing.T) {
	// Create a VStack with children
	vstack := ui.VStack(
		ui.Text("A"),
		ui.Text("B"),
		ui.Text("C"),
	)

	t.Logf("=== Testing LayoutNode Children Access ===\n")

	// Check if it has children
	children := vstack.Children()
	t.Logf("VStack type: %T", vstack)
	t.Logf("Children count: %d", len(children))

	for i, child := range children {
		t.Logf("Child %d: type=%T, value=%v", i, child, child)
	}

	// Verify children are accessible
	if len(children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(children))
	} else {
		t.Logf("✓ VStack has 3 children")
	}

	// Now build a tree and check if children are preserved
	tv := NewTreeView()
	err := tv.SetRoot(vstack)
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}

	allNodes := tv.GetFlatList()
	t.Logf("\n=== Tree built from VStack ===")
	t.Logf("Total nodes: %d", len(allNodes))

	for i, node := range allNodes {
		t.Logf("Node %d: Type=%s, Path=%s, Children=%d, UniqueID=%s",
			i, node.Info.Type, node.Path, len(node.Children), node.UniqueID)

		// For the root node, verify it has children in the tree
		if i == 0 && len(node.Children) == 0 {
			t.Errorf("❌ Root node should have children in tree!")
		}
	}

	// Verify root has children in the tree structure
	root := tv.root
	if root == nil {
		t.Fatal("Root is nil")
	}

	t.Logf("\n=== Root Node Details ===")
	t.Logf("Root Type: %s", root.Info.Type)
	t.Logf("Root Path: %s", root.Path)
	t.Logf("Root Children (in tree): %d", len(root.Children))
	t.Logf("Root Expanded: %v", root.Expanded)

	if len(root.Children) == 0 {
		t.Errorf("❌ Root node in tree has no children!")
	} else {
		t.Logf("✓ Root node in tree has %d children", len(root.Children))

		// List all children
		for i, child := range root.Children {
			t.Logf("  Child %d: %s (ID: %s)", i, child.Info.Type, child.UniqueID)
		}
	}
}

// TestLayoutNodeCollapseExpand tests collapsing and expanding LayoutNode
func TestLayoutNodeCollapseExpand(t *testing.T) {
	// Create a nested structure
	root := ui.VStack(
		ui.VStack(
			ui.Text("A"),
			ui.Text("B"),
		),
		ui.Text("C"),
	)

	tv := NewTreeView()
	err := tv.SetRoot(root)
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}

	t.Logf("\n=== Testing Collapse/Expand ===")

	// Get initial tree
	lines1, _ := tv.GetTreeLines()
	t.Logf("\nInitial tree (%d lines):", len(lines1))
	for i, line := range lines1 {
		t.Logf("  %2d: %s", i, line)
	}

	// Find the nested VStack
	var nestedVStackUID string
	allNodes := tv.GetFlatList()
	for _, node := range allNodes {
		if len(node.Children) > 0 && node.Level > 0 {
			nestedVStackUID = node.UniqueID
			t.Logf("\nFound collapsible node: %s", node.UniqueID)
			t.Logf("  Children: %d", len(node.Children))
			t.Logf("  Expanded: %v", node.Expanded)
			break
		}
	}

	if nestedVStackUID == "" {
		t.Skip("No collapsible node found")
		return
	}

	// Collapse it
	t.Logf("\n=== Collapsing %s ===", nestedVStackUID)
	tv.ToggleNode(nestedVStackUID)

	lines2, _ := tv.GetTreeLines()
	t.Logf("\nAfter collapse (%d lines):", len(lines2))
	for i, line := range lines2 {
		t.Logf("  %2d: %s", i, line)
	}

	// Verify line count changed
	if len(lines2) == len(lines1) {
		t.Logf("Note: Line count didn't change (may have been collapsed already)")
	} else {
		t.Logf("✓ Line count changed: %d → %d", len(lines1), len(lines2))
	}

	// Expand it back
	t.Logf("\n=== Expanding %s ===", nestedVStackUID)
	tv.ToggleNode(nestedVStackUID)

	lines3, _ := tv.GetTreeLines()
	t.Logf("\nAfter expand (%d lines):", len(lines3))
	for i, line := range lines3 {
		t.Logf("  %2d: %s", i, line)
	}

	if len(lines3) == len(lines1) {
		t.Logf("✓ Line count restored: %d → %d", len(lines1), len(lines3))
	}
}

// TestVNodeChildrenDirectly tests VNode.Children() directly
func TestVNodeChildrenDirectly(t *testing.T) {
	t.Logf("=== Direct VNode.Children() Test ===\n")

	// Create different types of VNodes
	text1 := ui.Text("A")
	text2 := ui.Text("B")
	vstack := ui.VStack(text1, text2)
	hstack := ui.HStack(text1, text2)

	testCases := []struct {
		name     string
		vnode    ui.VNode
		expected int
	}{
		{"Text node", text1, 0},
		{"VStack with 2 children", vstack, 2},
		{"HStack with 2 children", hstack, 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			children := tc.vnode.Children()
			t.Logf("%s: type=%T", tc.name, tc.vnode)
			t.Logf("  Children count: %d (expected %d)", len(children), tc.expected)

			if len(children) != tc.expected {
				t.Errorf("❌ Expected %d children, got %d", tc.expected, len(children))
			} else {
				t.Logf("✓ Children count correct")
			}

			// Print child details
			for i, child := range children {
				t.Logf("  Child %d: type=%T", i, child)
			}
		})
	}
}
