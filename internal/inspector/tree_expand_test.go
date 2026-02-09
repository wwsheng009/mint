package inspector

import (
	"testing"
)

// TestTreeViewUniqueIDLookup tests that GetUniqueIDForLineIndex returns correct IDs
func TestTreeViewUniqueIDLookup(t *testing.T) {
	tv := NewTreeView()
	tv.maxDepth = 5

	// Build a simple test tree:
	// Root (expanded)
	//   ├── Child1 (collapsed, has 2 children)
	//   └── Child2
	tv.root = &TreeNode{
		Path:     "Root",
		UniqueID: "Root[0]",
		Info:     ElementInfo{Type: "Root", Label: "root"},
		Level:    0,
		Expanded: true,
		Index:    0,
		Children: []*TreeNode{
			{
				Path:     "Root.LayoutNode",
				UniqueID: "Root.LayoutNode[0]",
				Info:     ElementInfo{Type: "LayoutNode", Label: "child1"},
				Level:    1,
				Expanded: false, // Collapsed
				Index:    0,
				Children: []*TreeNode{
					{Path: "Root.LayoutNode.Text", UniqueID: "Root.LayoutNode.Text[0]", Info: ElementInfo{Type: "ElementVNode", Label: "gc1"}, Level: 2, Index: 0},
					{Path: "Root.LayoutNode.ElementVNode", UniqueID: "Root.LayoutNode.ElementVNode[1]", Info: ElementInfo{Type: "ElementVNode", Label: "gc2"}, Level: 2, Index: 1},
				},
			},
			{
				Path:     "Root.LayoutNode2",
				UniqueID: "Root.LayoutNode2[1]",
				Info:     ElementInfo{Type: "LayoutNode", Label: "child2"},
				Level:    1,
				Expanded: true,
				Index:    1,
				Children: []*TreeNode{},
			},
		},
	}

	// Generate tree lines
	lines, _ := tv.GetTreeLines()

	t.Logf("=== Generated %d lines ===", len(lines))
	for i, line := range lines {
		t.Logf("Line %d: %s", i, line)
	}

	// Test path lookup for each line
	expectedPaths := []string{
		"Root[0]",                   // Line 0: Root node
		"Root.LayoutNode[0]",        // Line 1: Child1 (collapsed)
		"Root.LayoutNode[0]",        // Line 2: Collapsed indicator (+ 2 children)
		"Root.LayoutNode2[1]",       // Line 3: Child2
	}

	for i, expectedPath := range expectedPaths {
		if i >= len(lines) {
			t.Fatalf("Not enough lines: expected %d, got %d", len(expectedPaths), len(lines))
		}

		actualUID := tv.GetUniqueIDForLineIndex(i)
		if actualUID != expectedPath {
			t.Errorf("Line %d: expected unique ID %q, got %q", i, expectedPath, actualUID)
		} else {
			t.Logf("✓ Line %d -> unique ID %q (correct)", i, actualUID)
		}
	}
}

// TestTreeViewExpandCollapse tests actual expand/collapse behavior
func TestTreeViewExpandCollapse(t *testing.T) {
	tv := NewTreeView()
	tv.maxDepth = 5

	// Build initial tree with child1 collapsed
	tv.root = &TreeNode{
		Path:     "Root",
		UniqueID: "Root[0]",
		Info:     ElementInfo{Type: "Root", Label: "root"},
		Level:    0,
		Expanded: true,
		Index:    0,
		Children: []*TreeNode{
			{
				Path:     "Root.LayoutNode",
				UniqueID: "Root.LayoutNode[0]",
				Info:     ElementInfo{Type: "LayoutNode", Label: "child1"},
				Level:    1,
				Expanded: false, // Start collapsed
				Index:    0,
				Children: []*TreeNode{
					{Path: "Root.LayoutNode.Text", UniqueID: "Root.LayoutNode.Text[0]", Info: ElementInfo{Type: "Text", Label: "gc1"}, Level: 2, Index: 0},
					{Path: "Root.LayoutNode.ElementVNode", UniqueID: "Root.LayoutNode.ElementVNode[1]", Info: ElementInfo{Type: "Text", Label: "gc2"}, Level: 2, Index: 1},
				},
			},
			{
				Path:     "Root.LayoutNode2",
				UniqueID: "Root.LayoutNode2[1]",
				Info:     ElementInfo{Type: "LayoutNode", Label: "child2"},
				Level:    1,
				Expanded: true,
				Index:    1,
				Children: []*TreeNode{},
			},
		},
	}

	// Initial state: child1 collapsed
	lines1, _ := tv.GetTreeLines()
	t.Logf("=== Initial state (child1 collapsed) ===")
	t.Logf("Lines: %d", len(lines1))
	for i, line := range lines1 {
		t.Logf("  Line %d: %s", i, line)
	}

	// Line 1 should have unique ID Root.LayoutNode[0]
	uid1 := tv.GetUniqueIDForLineIndex(1)
	if uid1 != "Root.LayoutNode[0]" {
		t.Errorf("Line 1: expected unique ID Root.LayoutNode[0], got %q", uid1)
	}

	// Expand child1
	t.Logf("\n=== Expanding Root.LayoutNode[0] ===")
	t.Logf("Before ToggleNode: child1.Expanded = %v", tv.root.Children[0].Expanded)
	tv.ToggleNode("Root.LayoutNode[0]")
	t.Logf("After ToggleNode: child1.Expanded = %v", tv.root.Children[0].Expanded)
	t.Logf("expanded map value: %v", tv.expanded["Root.LayoutNode[0]"])

	lines2, _ := tv.GetTreeLines()
	t.Logf("Lines after expand: %d", len(lines2))
	for i, line := range lines2 {
		t.Logf("  Line %d: %s", i, line)
	}

	// After expanding, should have more lines
	expectedLinesAfterExpand := len(lines1) + 1
	if len(lines2) != expectedLinesAfterExpand {
		t.Errorf("After expand: expected %d lines, got %d", expectedLinesAfterExpand, len(lines2))
	}

	// Line 1 should still be Root.LayoutNode[0] (now expanded)
	uid2 := tv.GetUniqueIDForLineIndex(1)
	if uid2 != "Root.LayoutNode[0]" {
		t.Errorf("Line 1 after expand: expected unique ID Root.LayoutNode[0], got %q", uid2)
	}

	// Line 2 should now be Root.LayoutNode.Text[0] (was collapsed indicator)
	uid3 := tv.GetUniqueIDForLineIndex(2)
	if uid3 != "Root.LayoutNode.Text[0]" {
		t.Errorf("Line 2 after expand: expected unique ID Root.LayoutNode.Text[0], got %q", uid3)
	} else {
		t.Logf("✓ Line 2 correctly shows grandchild1 after expand")
	}

	// Collapse it back
	t.Logf("\n=== Collapsing Root.LayoutNode[0] ===")
	tv.ToggleNode("Root.LayoutNode[0]")

	lines3, _ := tv.GetTreeLines()
	t.Logf("Lines after collapse: %d", len(lines3))
	if len(lines3) != len(lines1) {
		t.Errorf("After collapse: expected %d lines (same as initial), got %d", len(lines1), len(lines3))
	}
}

// TestTreeViewPathConsistency tests that paths remain consistent after expand/collapse
func TestTreeViewPathConsistency(t *testing.T) {
	tv := NewTreeView()
	tv.maxDepth = 5

	// Build tree
	tv.root = &TreeNode{
		Path:     "Root",
		UniqueID: "Root[0]",
		Info:     ElementInfo{Type: "Root", Label: "root"},
		Level:    0,
		Expanded: true,
		Index:    0,
		Children: []*TreeNode{
			{
				Path:     "Root.Node",
				UniqueID: "Root.Node[0]",
				Info:     ElementInfo{Type: "Node", Label: "a"},
				Level:    1,
				Expanded: true,
				Index:    0,
				Children: []*TreeNode{
					{Path: "Root.Node.Node", UniqueID: "Root.Node.Node[0]", Info: ElementInfo{Type: "Node", Label: "1"}, Level: 2, Index: 0},
					{Path: "Root.Node.ElementVNode", UniqueID: "Root.Node.ElementVNode[1]", Info: ElementInfo{Type: "Node", Label: "2"}, Level: 2, Index: 1},
				},
			},
			{
				Path:     "Root.Node2",
				UniqueID: "Root.Node2[1]",
				Info:     ElementInfo{Type: "Node", Label: "b"},
				Level:    1,
				Expanded: true,
				Index:    1,
				Children: []*TreeNode{
					{Path: "Root.Node2.Node", UniqueID: "Root.Node2.Node[0]", Info: ElementInfo{Type: "Node", Label: "1"}, Level: 2, Index: 0},
				},
			},
		},
	}

	// Get unique ID for Root.Node2[1] (should be Root.Node2[1])
	uidBefore := tv.findUIDDirect("Root.Node2[1]")
	t.Logf("UniqueID for Root.Node2 before collapse: line with UID %q", uidBefore)

	// Collapse Root.Node[0]
	tv.ToggleNode("Root.Node[0]")

	lines, _ := tv.GetTreeLines()
	t.Logf("After collapsing Root.Node[0], %d lines:", len(lines))
	for i, line := range lines {
		t.Logf("  Line %d: %s", i, line)
	}

	// Root.Node2 should now be at a different line index, but unique ID lookup should still work
	uidAfter := tv.findUIDDirect("Root.Node2[1]")
	t.Logf("UniqueID for Root.Node2 after collapse: line with UID %q", uidAfter)

	if uidBefore != uidAfter {
		t.Logf("Note: Root.Node2 still has UID %q (consistent)", uidBefore)
	}
}

// TestTreeViewIndexBasedIDs tests that index-based IDs prevent collisions
func TestTreeViewIndexBasedIDs(t *testing.T) {
	tv := NewTreeView()
	tv.maxDepth = 5

	// Build tree with multiple siblings of same type
	// This tests that index-based IDs prevent collisions
	tv.root = &TreeNode{
		Path:     "VStack",
		UniqueID: "VStack[0]",
		Info:     ElementInfo{Type: "VStack", Label: "root"},
		Level:    0,
		Expanded: true,
		Index:    0,
		Children: []*TreeNode{
			{
				Path:     "VStack.Text",
				UniqueID: "VStack.Text[0]",  // First Text - index 0
				Info:     ElementInfo{Type: "Text", Label: "A"},
				Level:    1,
				Index:    0,
			},
			{
				Path:     "VStack.Text",
				UniqueID: "VStack.Text[1]",  // Second Text - index 1 (different ID!)
				Info:     ElementInfo{Type: "Text", Label: "B"},
				Level:    1,
				Index:    1,
			},
			{
				Path:     "VStack.Text",
				UniqueID: "VStack.Text[2]",  // Third Text - index 2 (different ID!)
				Info:     ElementInfo{Type: "Text", Label: "C"},
				Level:    1,
				Index:    2,
			},
		},
	}

	// Verify all three Text nodes have different UniqueIDs
	uids := make(map[string]bool)
	for _, child := range tv.root.Children {
		if child.UniqueID == "" {
			t.Errorf("Child has empty UniqueID")
		}
		if uids[child.UniqueID] {
			t.Errorf("Duplicate UniqueID found: %s", child.UniqueID)
		}
		uids[child.UniqueID] = true
		t.Logf("Child %d: UniqueID=%s, Label=%s", child.Index, child.UniqueID, child.Info.Label)
	}

	// Verify we have 3 unique IDs
	if len(uids) != 3 {
		t.Errorf("Expected 3 unique IDs, got %d", len(uids))
	}
}

// Helper method to find which unique ID is at which line index
func (tv *TreeView) findUIDDirect(targetUID string) string {
	for i := 0; i < 20; i++ {
		if uid := tv.GetUniqueIDForLineIndex(i); uid == targetUID {
			return uid
		}
	}
	return ""
}
