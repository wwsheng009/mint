package display

import (
	"testing"

	"github.com/wwsheng009/mint/runtime"
)

// TestTreeViewUpdateLinesPreservesViewportHeight tests that UpdateLines preserves viewportHeight
func TestTreeViewUpdateLinesPreservesViewportHeight(t *testing.T) {
	lines := []string{
		"Root",
		"├── Child 1",
		"│   ├── Grandchild 1.1",
		"│   └── Grandchild 1.2",
		"├── Child 2",
		"└── Child 3",
	}

	treeViewVNode := NewTreeView().
		FromLines(lines).
		Build()

	treeView, ok := treeViewVNode.(*TreeView)
	if !ok {
		t.Fatal("Build() should return *TreeView")
	}

	// First measurement - should set viewportHeight
	constraints1 := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 5,
	}

	measurable, ok := treeViewVNode.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	})
	if !ok {
		t.Fatal("TreeView should implement Measurable interface")
	}

	size1 := measurable.Measure(constraints1)
	t.Logf("First measurement: size=%dx%d, viewportHeight should be 5", size1.Width, size1.Height)

	// Check that viewportHeight was set
	if treeView.viewportHeight != 5 {
		t.Errorf("After first measurement, viewportHeight = %d, want 5", treeView.viewportHeight)
	}

	// Update lines - this should preserve viewportHeight
	newLines := []string{
		"Root (Updated)",
		"├── Child 1 (Updated)",
		"└── Child 2 (Updated)",
	}

	treeView.UpdateLines(newLines)

	// Verify viewportHeight is still 5 after UpdateLines
	if treeView.viewportHeight != 5 {
		t.Errorf("After UpdateLines, viewportHeight = %d, want 5 (should be preserved)", treeView.viewportHeight)
	}

	// Second measurement with different constraints
	constraints2 := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 8,
	}

	size2 := measurable.Measure(constraints2)
	t.Logf("Second measurement: size=%dx%d, viewportHeight should be 8", size2.Width, size2.Height)

	// Check that viewportHeight was updated
	if treeView.viewportHeight != 8 {
		t.Errorf("After second measurement, viewportHeight = %d, want 8", treeView.viewportHeight)
	}

	// Update lines again
	treeView.UpdateLines(lines)

	// Verify viewportHeight is still 8 after UpdateLines
	if treeView.viewportHeight != 8 {
		t.Errorf("After second UpdateLines, viewportHeight = %d, want 8 (should be preserved)", treeView.viewportHeight)
	}

	t.Logf("✓ UpdateLines correctly preserves viewportHeight")
}
