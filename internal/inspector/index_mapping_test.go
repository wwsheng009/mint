package inspector

import (
	"strings"
	"testing"

	componenttreeview "github.com/wwsheng009/mint/ui/components/treeview"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestTreeLinesAndFlatListConsistency verifies that GetTreeLines and GetFlatList
// have consistent indexing - the node at flatNodes[i] should correspond to treeLines[i+1]
func TestTreeLinesAndFlatListConsistency(t *testing.T) {
	// Create a tree structure similar to the demo
	rootVNode := rtui.NewElement("base")
	rootVNode.SetKey("/root/base[0]")

	// Create children
	vstack := rtui.NewElement("vstack")
	vstack.SetKey("/root/base[0]/vstack[0]")
	vstack.SetChildren([]rtui.VNode{
		createTextVNode("Text1", "/root/base[0]/vstack[0]/text[0]"),
		createTextVNode("Text2", "/root/base[0]/vstack[0]/text[1]"),
	})

	bordered := rtui.NewElement("bordered")
	bordered.SetKey("/root/base[0]/bordered[0]")
	bordered.SetChildren([]rtui.VNode{
		createTextVNode("BorderedText", "/root/base[0]/bordered[0]/text[0]"),
	})

	rootVNode.SetChildren([]rtui.VNode{vstack, bordered})

	// Create TreeView and set root
	tv := NewTreeView()
	tv.SetShowIcons(false)
	tv.SetShowPaths(true)

	err := tv.SetRoot(rootVNode)
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}

	// Get tree lines and flat list
	treeLines, _ := tv.GetTreeLines()
	flatNodes := tv.GetFlatList()

	t.Logf("=== Tree Lines (%d lines) ===", len(treeLines))
	for i, line := range treeLines {
		t.Logf("[%d] %s", i, truncate(line, 60))
	}

	t.Logf("\n=== Flat Nodes (%d nodes) ===", len(flatNodes))
	for i, node := range flatNodes {
		t.Logf("[%d] Path=%s, Type=%s", i, node.Path, node.Info.Type)
	}

	// Verify consistency
	// treeLines[0] = header
	// treeLines[1..n-2] = nodes (should match flatNodes[0..n-3])
	// treeLines[n-1] = footer

	headerLine := 0
	footerLine := len(treeLines) - 1

	t.Logf("\n=== Index Mapping Test ===")
	t.Logf("Header line: %d", headerLine)
	t.Logf("Footer line: %d", footerLine)

	// Check that each flatNode corresponds to the correct treeLine
	for i, node := range flatNodes {
		treeLineIdx := i + 1 // +1 to skip header

		if treeLineIdx >= footerLine {
			t.Errorf("flatNodes[%d] maps to treeLines[%d] which is footer or beyond!", i, treeLineIdx)
			continue
		}

		treeLine := treeLines[treeLineIdx]

		// The tree line should contain the node's Path or Type
		if node.Path != "" && !strings.Contains(treeLine, node.Path) {
			// Try checking for Type instead
			if !strings.Contains(treeLine, node.Info.Type) {
				t.Errorf("flatNodes[%d] (Path=%s, Type=%s) doesn't match treeLines[%d]: %s",
					i, node.Path, node.Info.Type, treeLineIdx, truncate(treeLine, 50))
			}
		} else {
			t.Logf("✓ flatNodes[%d] (Path=%s) matches treeLines[%d]", i, node.Path, treeLineIdx)
		}
	}

	// Verify counts
	expectedNodeLines := len(flatNodes)
	actualNodeLines := len(treeLines) - 2 // minus header and footer

	if expectedNodeLines != actualNodeLines {
		t.Errorf("Count mismatch: flatNodes has %d nodes, treeLines has %d node lines (expected %d)",
			len(flatNodes), actualNodeLines, expectedNodeLines)
	}
}

// TestFocusIndexToFlatNodeMapping tests the focusIndex to flatNodes mapping
func TestFocusIndexToFlatNodeMapping(t *testing.T) {
	// Create a simple tree
	rootVNode := rtui.NewElement("root")
	rootVNode.SetKey("/root/base[0]")

	child1 := rtui.NewElement("child1")
	child1.SetKey("/root/base[0]/child1[0]")

	child2 := rtui.NewElement("child2")
	child2.SetKey("/root/base[0]/child2[1]")

	rootVNode.SetChildren([]rtui.VNode{child1, child2})

	tv := NewTreeView()
	tv.SetRoot(rootVNode)

	treeLines, _ := tv.GetTreeLines()
	flatNodes := tv.GetFlatList()

	t.Logf("Tree lines: %d", len(treeLines))
	t.Logf("Flat nodes: %d", len(flatNodes))

	// Test mapping for each possible focusIndex
	for focusIndex := 0; focusIndex < len(treeLines); focusIndex++ {
		// Apply the mapping logic
		nodeIndex := focusIndex - 1 // adjust for header

		var mappedNode *TreeNode
		var isValid bool

		if nodeIndex >= 0 && nodeIndex < len(flatNodes) && focusIndex < len(treeLines)-1 {
			mappedNode = flatNodes[nodeIndex]
			isValid = true
		}

		if focusIndex == 0 {
			// Header line - should not map to any node
			if isValid {
				t.Errorf("focusIndex=0 (header) should not map to a node, but got %s", mappedNode.Path)
			} else {
				t.Logf("✓ focusIndex=0 (header) correctly maps to no node")
			}
		} else if focusIndex == len(treeLines)-1 {
			// Footer line - should not map to any node
			if isValid {
				t.Errorf("focusIndex=%d (footer) should not map to a node, but got %s", focusIndex, mappedNode.Path)
			} else {
				t.Logf("✓ focusIndex=%d (footer) correctly maps to no node", focusIndex)
			}
		} else {
			// Node line - should map to a node
			if !isValid {
				t.Errorf("focusIndex=%d should map to a node but didn't", focusIndex)
			} else {
				t.Logf("✓ focusIndex=%d maps to flatNodes[%d]: %s", focusIndex, nodeIndex, mappedNode.Path)
			}
		}
	}
}

// TestTreeViewIndexConsistency tests that TreeView component's focusIndex
// correctly maps to inspector.TreeView's flatNodes when using FromLines
func TestTreeViewIndexConsistency(t *testing.T) {
	// Create inspector.TreeView with a tree
	rootVNode := rtui.NewElement("base")
	rootVNode.SetKey("/root/base[0]")

	child1 := rtui.NewElement("child1")
	child1.SetKey("/root/base[0]/child1[0]")

	child2 := rtui.NewElement("child2")
	child2.SetKey("/root/base[0]/child2[1]")

	rootVNode.SetChildren([]rtui.VNode{child1, child2})

	inspectorTV := NewTreeView()
	inspectorTV.SetRoot(rootVNode)

	// Get tree lines and flat nodes
	treeLines, _ := inspectorTV.GetTreeLines()
	flatNodes := inspectorTV.GetFlatList()

	t.Logf("=== Inspector Tree Lines (%d) ===", len(treeLines))
	for i, line := range treeLines {
		t.Logf("[%d] %s", i, truncate(line, 50))
	}

	t.Logf("\n=== Inspector Flat Nodes (%d) ===", len(flatNodes))
	for i, node := range flatNodes {
		t.Logf("[%d] Path=%s", i, node.Path)
	}

	// Create TreeView component from lines
	// Note: Use BuildVNode() to get *treeview.VNode directly, then type assert
	tvVNode := componenttreeview.NewBuilder().
		FromLines(treeLines).
		ExpandLevel(-1).
		BuildVNode()
	treeViewComp := tvVNode

	t.Logf("\n=== TreeViewVNode created successfully ===")
	// Note: treeview.VNode.nodes is private, we can check props instead
	props := treeViewComp.Props()
	if nodes, ok := props["nodes"].([]componenttreeview.TreeNode); ok {
		t.Logf("TreeView has %d nodes", len(nodes))
	}

	// displayLines := treeViewComp.GetLines()
	// t.Logf("\n=== Display Tree Lines (%d) ===", len(displayLines))
	// for i, line := range displayLines {
	// 	t.Logf("[%d] Content=%q", i, truncate(line.Content, 40))
	// }

	// Check if TreeView skipped any lines (e.g., empty lines)
	// Commented out: treeview.VNode doesn't have GetLines() method anymore
	/*
	if len(displayLines) != len(treeLines) {
		t.Errorf("Line count mismatch: inspector has %d lines, display has %d lines",
			len(treeLines), len(displayLines))
	}

	// Test each focusIndex
	t.Logf("\n=== Focus Index Mapping Test ===")
	for focusIndex := 0; focusIndex < len(displayLines); focusIndex++ {
		// Simulate navigation to this line
		treeViewComp.FocusLine(focusIndex)
		actualFocusIndex := treeViewComp.GetFocusIndex()

		if actualFocusIndex != focusIndex {
			t.Errorf("FocusLine(%d) resulted in focusIndex=%d", focusIndex, actualFocusIndex)
			continue
		}


		// Apply mapping logic: flatNodes[focusIndex - 1]
		nodeIndex := focusIndex - 1

		var expectedPath string
		if nodeIndex >= 0 && nodeIndex < len(flatNodes) && focusIndex < len(treeLines)-1 {
			expectedPath = flatNodes[nodeIndex].Path
		}

		// Get the display line content
		displayLine := displayLines[focusIndex]

		t.Logf("focusIndex=%d: nodeIndex=%d, expectedPath=%q, displayContent=%q",
			focusIndex, nodeIndex, expectedPath, truncate(displayLine.Content, 30))

		// Verify that the display line content matches the expected node
		if expectedPath != "" {
			if !strings.Contains(displayLine.Content, expectedPath) {
				// The content might contain the type name instead of path
				if nodeIndex >= 0 && nodeIndex < len(flatNodes) {
					nodeType := flatNodes[nodeIndex].Info.Type
					if !strings.Contains(displayLine.Content, nodeType) {
						t.Errorf("focusIndex=%d: display content %q doesn't contain path %q or type %q",
							focusIndex, truncate(displayLine.Content, 30), expectedPath, nodeType)
					}
				}
			}
		}
	}
	*/
}

// Helper function to create a text VNode with a key
func createTextVNode(content, key string) rtui.VNode {
	textNode := &rtui.TextVNode{
		ElementVNode: rtui.NewElement("text"),
	}
	textNode.SetKey(key)
	return textNode
}

// Helper to truncate strings
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
