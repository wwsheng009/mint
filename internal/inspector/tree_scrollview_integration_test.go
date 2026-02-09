package inspector

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/ui"
)

// TestTreeViewWithScrollView verifies that TreeView works correctly when wrapped in ScrollView
func TestTreeViewWithScrollView(t *testing.T) {
	t.Log("\n=== Testing TreeView + ScrollView Integration ===\n")

	inspector := NewStandaloneInspector()
	inspector.Enable()

	// Create a tree with 20 nodes
	var children []ui.VNode
	for i := 0; i < 20; i++ {
		children = append(children, ui.Text(fmt.Sprintf("Node %d", i+1)))
	}
	root := ui.VStack(children...)

	inspector.AttachToApp(root)

	// Get tree lines (before ScrollView wrapping)
	treeView := inspector.GetTreeView()
	lines, totalLines := treeView.GetTreeLines()

	t.Logf("✅ Tree has %d lines, %d total nodes", len(lines), totalLines)

	// Tree may have more nodes than content items due to wrapper nodes
	// (VStack wrapper, LayoutNode wrappers, etc.)
	if totalLines < 20 {
		t.Errorf("Expected at least 20 total nodes, got %d", totalLines)
	} else {
		t.Logf("✅ Tree has %d nodes (>= 20 content items)", totalLines)
	}

	if len(lines) == 0 {
		t.Error("❌ Tree lines should not be empty")
	} else {
		t.Logf("✅ Tree has %d lines", len(lines))
	}

	// Check that the lines contain expected content
	firstLine := lines[0]
	if len(firstLine) == 0 {
		t.Error("❌ First line should not be empty")
	} else {
		t.Logf("✅ First line: %s", firstLine)
	}

	// Verify that ScrollView wrapper doesn't affect tree extraction
	elementsTab := inspector.buildElementsTabContent()
	if elementsTab == nil {
		t.Fatal("❌ Elements tab should not be nil")
	}
	t.Log("✅ Elements tab created successfully with ScrollView wrapper")

	// Verify tree still has correct content after wrapping
	_, totalLines2 := treeView.GetTreeLines()
	if totalLines2 != totalLines {
		t.Errorf("❌ Tree total changed after wrapping: %d -> %d", totalLines, totalLines2)
	} else {
		t.Log("✅ Tree content unchanged after ScrollView wrapping")
	}

	t.Log("\n=== Test Complete ===")
}
