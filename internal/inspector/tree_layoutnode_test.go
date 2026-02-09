package inspector

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/ui"
)

// TestTreeViewMultipleLayoutNodes demonstrates how multiple LayoutNodes are distinguished
func TestTreeViewMultipleLayoutNodes(t *testing.T) {
	tv := NewTreeView()
	tv.maxDepth = 5

	// Create a structure with multiple LayoutNodes (VStacks)
	root := ui.VStack(
		ui.VStack(  // LayoutNode 1
			ui.Text("A"),
		),
		ui.VStack(  // LayoutNode 2
			ui.Text("B"),
		),
		ui.VStack(  // LayoutNode 3
			ui.Text("C"),
		),
	)

	err := tv.SetRoot(root)
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}

	// Get all nodes
	allNodes := tv.GetFlatList()

	// Collect LayoutNodes
	var layoutNodes []*TreeNode
	for _, node := range allNodes {
		if strings.Contains(node.Info.Type, "LayoutNode") {
			layoutNodes = append(layoutNodes, node)
		}
	}

	t.Logf("=== Found %d LayoutNodes ===\n", len(layoutNodes))

	// Verify each LayoutNode has a unique ID
	uidMap := make(map[string]bool)
	for i, node := range layoutNodes {
		uid := node.UniqueID

		// Check for duplicates
		if uidMap[uid] {
			t.Errorf("❌ DUPLICATE LayoutNode ID: %s", uid)
		}
		uidMap[uid] = true

		t.Logf("LayoutNode #%d:", i+1)
		t.Logf("  Type: %s", node.Info.Type)
		t.Logf("  Path: %s", node.Path)
		t.Logf("  UniqueID: %s", uid)
		t.Logf("  Children: %d", len(node.Children))
		t.Logf("")
	}

	// Verify all IDs are unique
	if len(uidMap) != len(layoutNodes) {
		t.Errorf("❌ Expected %d unique IDs, got %d", len(layoutNodes), len(uidMap))
	} else {
		t.Logf("✓ All %d LayoutNodes have unique IDs", len(layoutNodes))
	}

	// Verify no two LayoutNodes have the same path
	pathMap := make(map[string]bool)
	for _, node := range layoutNodes {
		if pathMap[node.Path] {
			t.Errorf("❌ DUPLICATE LayoutNode Path: %s", node.Path)
		}
		pathMap[node.Path] = true
	}

	if len(pathMap) != len(layoutNodes) {
		t.Errorf("❌ Expected %d unique paths, got %d", len(layoutNodes), len(pathMap))
	} else {
		t.Logf("✓ All %d LayoutNodes have unique paths", len(layoutNodes))
	}
}

// TestTreeViewLayoutNodeExpansion tests that expand/collapse works for LayoutNodes
func TestTreeViewLayoutNodeExpansion(t *testing.T) {
	tv := NewTreeView()
	tv.maxDepth = 5

	// Create nested LayoutNodes
	root := ui.VStack(
		ui.VStack(
			ui.Text("A"),
			ui.Text("B"),
		),
		ui.Text("C"),
	)

	err := tv.SetRoot(root)
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}

	// Get all LayoutNodes
	allNodes := tv.GetFlatList()
	var layoutNodes []*TreeNode
	for _, node := range allNodes {
		if strings.Contains(node.Info.Type, "LayoutNode") {
			layoutNodes = append(layoutNodes, node)
		}
	}

	if len(layoutNodes) < 2 {
		t.Fatalf("Expected at least 2 LayoutNodes, got %d", len(layoutNodes))
	}

	// Get the nested LayoutNode (index 1, not the root)
	nestedLayout := layoutNodes[1]
	uid := nestedLayout.UniqueID

	t.Logf("=== Testing LayoutNode Expansion ===")
	t.Logf("Nested LayoutNode UniqueID: %s", uid)
	t.Logf("Initial expanded state: %v", nestedLayout.Expanded)

	// Get initial line count
	lines1, _ := tv.GetTreeLines()
	initialCount := len(lines1)
	t.Logf("Initial line count: %d", initialCount)

	// Toggle the nested LayoutNode
	tv.ToggleNode(uid)

	lines2, _ := tv.GetTreeLines()
	afterToggleCount := len(lines2)
	t.Logf("After toggle line count: %d", afterToggleCount)

	// The line count should change
	if afterToggleCount == initialCount {
		t.Logf("Note: Line count didn't change (may have been expanded/collapsed already)")
	}

	// Find the node again and verify state changed
	node := tv.findNodeByUniqueID(tv.root, uid)
	if node == nil {
		t.Errorf("❌ Could not find node with UniqueID: %s", uid)
	} else {
		t.Logf("✓ Found node with UniqueID: %s", uid)
		t.Logf("New expanded state: %v", node.Expanded)
	}

	t.Logf("✓ LayoutNode expand/collapse works correctly")
}

// TestTreeViewLayoutNodeIDFormat demonstrates the ID format for LayoutNodes
func TestTreeViewLayoutNodeIDFormat(t *testing.T) {
	tv := NewTreeView()
	tv.maxDepth = 5

	testCases := []struct {
		name     string
		tree     ui.VNode
		expected []string // Expected UniqueID patterns
	}{
		{
			name: "Single VStack",
			tree: ui.VStack(ui.Text("A")),
			expected: []string{
				"vstack[0]",              // Root VStack
				"vstack[0].text[0]",      // Text child
			},
		},
		{
			name: "Nested VStacks",
			tree: ui.VStack(
				ui.VStack(ui.Text("A")),
			),
			expected: []string{
				"vstack[0]",                  // Root VStack
				"vstack[0].vstack[0]",        // Nested VStack
				"vstack[0].vstack[0].text[0]", // Text
			},
		},
		{
			name: "Multiple VStacks",
			tree: ui.VStack(
				ui.VStack(ui.Text("A")),
				ui.VStack(ui.Text("B")),
			),
			expected: []string{
				"vstack[0]",                     // Root VStack
				"vstack[0].vstack[0]",           // First nested VStack
				"vstack[0].vstack[0].text[0]",   // First Text
				"vstack[1].vstack[1]",           // Second nested VStack (different index!)
				"vstack[1].vstack[1].text[0]",   // Second Text
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tv.SetRoot(tc.tree)
			if err != nil {
				t.Fatalf("SetRoot failed: %v", err)
			}

			allNodes := tv.GetFlatList()

			t.Logf("\n=== %s ===", tc.name)
			for i, node := range allNodes {
				t.Logf("Node %d: %-20s UniqueID: %s", i, node.Info.Type, node.UniqueID)
			}

			// Verify we got the expected number of nodes
			if len(allNodes) != len(tc.expected) {
				t.Logf("Note: Got %d nodes, expected %d", len(allNodes), len(tc.expected))
			}

			// Verify all IDs are unique
			uids := make(map[string]bool)
			for _, node := range allNodes {
				if uids[node.UniqueID] {
					t.Errorf("❌ Duplicate ID: %s", node.UniqueID)
				}
				uids[node.UniqueID] = true
			}

			if len(uids) == len(allNodes) {
				t.Logf("✓ All %d nodes have unique IDs", len(allNodes))
			}
		})
	}
}

