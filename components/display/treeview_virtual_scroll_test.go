package display

import (
	"fmt"
	"os"
	"testing"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
)

// TestTreeViewChildrenConstraints tests that TreeView's internal VStack
// receives proper height constraints
func TestTreeViewChildrenConstraints(t *testing.T) {
	os.Setenv("TUI_DEBUG_LAYOUT", "true")

	// Create TreeView with many lines
	lines := make([]string, 30)
	for i := 0; i < 30; i++ {
		lines[i] = fmt.Sprintf("Line %d", i)
	}

	treeView := NewTreeView().
		FromLines(lines).
		Build()

	// Measure TreeView with bounded height
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 10, // Only show 10 lines
	}

	engine := compute.NewEngine()
	layout, err := engine.Layout(treeView, nil, constraints)
	if err != nil {
		t.Fatalf("Layout failed: %v", err)
	}

	fmt.Printf("\n[TEST] TreeView layout: %dx%d (constraint MaxHeight=10)\n",
		layout.Root.Box.Width, layout.Root.Box.Height)

	// TreeView should respect the constraint
	if layout.Root.Box.Height > 10 {
		t.Errorf("TreeView height %d exceeds constraint 10", layout.Root.Box.Height)
	}

	// Check children
	fmt.Printf("[TEST] TreeView has %d children:\n", len(layout.Root.Children))
	for i, child := range layout.Root.Children {
		fmt.Printf("[TEST]   Child %d: type=%-15s size=%dx%d\n",
			i, child.VNode.Type(), child.Box.Width, child.Box.Height)

		// The child is a VStack containing all lines
		// It should NOT exceed the parent's height
		if child.Box.Height > 10 {
			t.Errorf("Child %d (VStack with lines) height %d exceeds parent constraint 10",
				i, child.Box.Height)
		}
	}

	// Also check what TreeView.Measure() returns directly
	measurable, ok := treeView.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	})
	if ok {
		size := measurable.Measure(constraints)
		fmt.Printf("[TEST] TreeView.Measure() directly: %dx%d\n", size.Width, size.Height)

		// TreeView.Measure() should return the bounded height
		if size.Height != 10 {
			t.Errorf("TreeView.Measure() height = %d, want 10", size.Height)
		}
	}
}

// TestTreeViewVirtualScrolling verifies that TreeView uses virtual scrolling
func TestTreeViewVirtualScrolling(t *testing.T) {
	os.Setenv("TUI_DEBUG_INSPECTOR", "true")

	// Create TreeView with 100 lines
	lines := make([]string, 100)
	for i := 0; i < 100; i++ {
		lines[i] = fmt.Sprintf("Node %d with some description text", i)
	}

	treeView := NewTreeView().
		FromLines(lines).
		Build()

	// Measure with small viewport
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 10,
	}

	measurable, ok := treeView.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	})
	if !ok {
		t.Fatal("TreeView should implement Measurable")
	}

	size := measurable.Measure(constraints)
	fmt.Printf("\n[TEST] TreeView with 100 lines, viewport height=10\n")
	fmt.Printf("[TEST] Measured size: %dx%d\n", size.Width, size.Height)

	// TreeView should return viewport height
	if size.Height != 10 {
		t.Errorf("TreeView height = %d, want 10", size.Height)
	}

	// Check children count - should be limited by viewport
	children := treeView.Children()
	fmt.Printf("[TEST] TreeView children count: %d (should be ~10 for virtual scrolling)\n", len(children))

	// With virtual scrolling, children should be limited to viewport height
	if len(children) > 15 { // Allow some margin
		t.Errorf("TreeView has %d children, expected ~10 (virtual scrolling not working)", len(children))
	} else if len(children) <= 10 {
		fmt.Printf("[TEST] ✓ Virtual scrolling appears to be working (only %d children for 100 lines)\n", len(children))
	}
}

