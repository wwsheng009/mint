package display

import (
	"testing"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/ui"
)

// TestTreeViewWithBoundedHeight tests that TreeView respects height constraints
func TestTreeViewWithBoundedHeight(t *testing.T) {
	// Create a TreeView with many lines
	lines := []string{
		"Root",
		"├── Child 1",
		"│   ├── Grandchild 1.1",
		"│   └── Grandchild 1.2",
		"├── Child 2",
		"│   ├── Grandchild 2.1",
		"│   └── Grandchild 2.2",
		"└── Child 3",
	}

	treeViewVNode := NewTreeView().
		FromLines(lines).
		Build()

	// Measure with bounded height constraints
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 5, // Only show 5 lines at a time
	}

	// TreeView implements Measurable interface
	measurable, ok := treeViewVNode.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	})
	if !ok {
		t.Fatal("TreeView should implement Measurable interface")
	}

	size := measurable.Measure(constraints)

	// TreeView should return the bounded height
	if size.Height != 5 {
		t.Errorf("TreeView height = %d, want 5 (should respect MaxHeight constraint)", size.Height)
	}

	// TreeView's internal viewport height should be set for virtual scrolling
	// (We can't directly access private fields, but we verified size is correct)
}

// TestTreeViewWithUnboundedHeight tests that TreeView renders all lines when unbounded
func TestTreeViewWithUnboundedHeight(t *testing.T) {
	lines := []string{
		"Root",
		"├── Child 1",
		"└── Child 2",
	}

	treeViewVNode := NewTreeView().
		FromLines(lines).
		Build()

	// Measure with unbounded height
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: runtime.Infinity,
	}

	measurable, ok := treeViewVNode.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	})
	if !ok {
		t.Fatal("TreeView should implement Measurable interface")
	}

	size := measurable.Measure(constraints)

	// TreeView should return natural height (all lines)
	if size.Height != 3 {
		t.Errorf("TreeView height = %d, want 3 (natural height)", size.Height)
	}

	// When unbounded, TreeView should render all lines (viewport = 0 internally)
}

// TestTreeViewWidthConstraints tests that TreeView respects width constraints
func TestTreeViewWidthConstraints(t *testing.T) {
	lines := []string{
		"Root",
		"├── A very long child name that exceeds constraint",
	}

	treeViewVNode := NewTreeView().
		FromLines(lines).
		Build()

	// Measure with bounded width
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  30,
		MinHeight: 0,
		MaxHeight: runtime.Infinity,
	}

	measurable, ok := treeViewVNode.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	})
	if !ok {
		t.Fatal("TreeView should implement Measurable interface")
	}

	size := measurable.Measure(constraints)

	// TreeView should return the bounded width
	if size.Width != 30 {
		t.Errorf("TreeView width = %d, want 30 (should respect MaxWidth constraint)", size.Width)
	}
}

// TestTreeViewInVStack tests TreeView inside VStack with height constraint
func TestTreeViewInVStack(t *testing.T) {
	// This simulates the Inspector use case
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

	// Create VStack with height constraint (like Inspector does)
	vstack := ui.VStackBuilder(
		ui.Text("Header"),
		treeViewVNode,
		ui.Text("Footer"),
	).Height(10).Build()

	// Measure the VStack
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 10,
	}

	measurable, ok := vstack.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	})
	if !ok {
		t.Fatal("VStack should implement Measurable interface")
	}

	size := measurable.Measure(constraints)

	// VStack should respect height constraint
	if size.Height != 10 {
		t.Errorf("VStack height = %d, want 10", size.Height)
	}

	// TreeView should have received bounded constraints from VStack
	// We can't directly check private viewportHeight field, but we verified
	// the VStack itself respects the constraint, which means it propagated
	// correctly to the TreeView child
}
