package inspector

import (
	"fmt"
	"os"
	"testing"

	"github.com/wwsheng009/mint/components/display"
	"github.com/wwsheng009/mint/components/navigation"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	"github.com/wwsheng009/mint/ui"
)

// TestInspectorLayoutConstraintChain tests the full constraint chain in Inspector
func TestInspectorLayoutConstraintChain(t *testing.T) {
	os.Setenv("TUI_DEBUG_LAYOUT", "true")

	// Create Inspector with same dimensions as demo2
	insp := NewStandaloneInspector()
	insp.overlayWidth = 80
	insp.overlayHeight = 25

	// Build the actual Inspector overlay content
	overlayContent := insp.buildOverlayContent()

	fmt.Printf("[TEST] ============================================\n")
	fmt.Printf("[TEST] Testing Inspector Layout Constraint Chain\n")
	fmt.Printf("[TEST] ============================================\n\n")

	// Create layout engine
	engine := compute.NewEngine()

	// Measure with overlay constraints
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 25,
	}

	layout, err := engine.Layout(overlayContent, nil, constraints)
	if err != nil {
		t.Fatalf("Layout failed: %v", err)
	}

	fmt.Printf("\n[TEST] Layout complete: root size = %dx%d\n\n",
		layout.Root.Box.Width, layout.Root.Box.Height)

	// Traverse layout tree to find TreeView and check its constraints
	fmt.Printf("[TEST] Traversing layout tree to find TreeView...\n")
	findAndCheckTreeView(layout.Root, 0, 25, t)
}

// findAndCheckTreeView recursively finds TreeView and verifies it received bounded constraints
func findAndCheckTreeView(box *compute.ComputedBox, depth int, maxHeight int, t *testing.T) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	// Check if this node is a TreeView
	isTreeView := false
	if box.VNode.Type().String() == "element" {
		// TreeView is an element, check by type assertion or props
		if tv, ok := box.VNode.(*display.TreeView); ok {
			isTreeView = true
			fmt.Printf("[TEST]%s✓ FOUND TreeView at depth %d\n", indent, depth)
			fmt.Printf("[TEST]%s  Size: %dx%d\n", indent, box.Box.Width, box.Box.Height)
			fmt.Printf("[TEST]%s  Position: (%d, %d)\n", indent, box.Box.X, box.Box.Y)

			// TreeView should be constrained
			if box.Box.Height > maxHeight {
				t.Errorf("[TEST]%s  ✗ FAIL: TreeView height %d exceeds max height %d",
					indent, box.Box.Height, maxHeight)
			} else {
				fmt.Printf("[TEST]%s  ✓ PASS: TreeView height %d within bounds\n",
					indent, box.Box.Height)
			}

			// Check if TreeView has bounded height (should use virtual scrolling)
			// TreeView internal state is private, but we can infer from size
			if box.Box.Height < maxHeight && box.Box.Height > 0 {
				fmt.Printf("[TEST]%s  ✓ TreeView appears to be using virtual scrolling (height=%d < %d)\n",
					indent, box.Box.Height, maxHeight)
			}
			_ = tv // Avoid unused variable warning
		}
	}

	if !isTreeView {
		// Log non-treeview nodes at shallow depth
		if depth < 5 {
			fmt.Printf("[TEST]%sNode: type=%-15s size=%dx%d pos=(%d,%d)\n",
				indent, box.VNode.Type(), box.Box.Width, box.Box.Height, box.Box.X, box.Box.Y)
		}
	}

	// Recursively check children
	for _, child := range box.Children {
		// Adjust max height for children based on container
		childMaxHeight := maxHeight
		if box.Box.Height > 0 && box.Box.Height < childMaxHeight {
			childMaxHeight = box.Box.Height
		}

		findAndCheckTreeView(child, depth+1, childMaxHeight, t)
	}
}

// TestTabsWithNestedVStackConstraints tests Tabs containing VStack with height constraint
func TestTabsWithNestedVStackConstraints(t *testing.T) {
	os.Setenv("TUI_DEBUG_LAYOUT", "true")

	// Simulate Inspector's Elements tab structure
	// Create a VStack with content that exceeds height constraint (like Inspector does)
	innerVStack := ui.VStackBuilder(
		ui.Text("📦 Layout Tree"),
		ui.Text("Nodes: 32 | Depth: 4 | Leaves: 20"),
		ui.Text(""),
		ui.Text("────────────────────────────────────"),
		ui.Text("Focused: Element"),
		ui.Text("Path: vstack"),
		ui.Text(""),
		// TreeView would go here - simulate with many Text nodes
		ui.Text("> Root"),
		ui.Text("  ├── Child 1"),
		ui.Text("  │   ├── Grandchild 1.1"),
		ui.Text("  │   └── Grandchild 1.2"),
		ui.Text("  ├── Child 2"),
		ui.Text("  └── Child 3"),
		ui.Text(""),
		ui.Text("Instructions: Use arrow keys to navigate"),
	).
		Width(76).
		Height(20).  // This is what Inspector does
		Build()

	// Verify innerVStack has Height prop
	props := innerVStack.Props()
	heightProp, hasHeight := props["height"].(int)
	fmt.Printf("[TEST] Inner VStack has Height prop: %v, value: %d\n", hasHeight, heightProp)

	if !hasHeight || heightProp != 20 {
		t.Errorf("Inner VStack should have Height(20) prop, got: %d", heightProp)
	}

	// Create Tabs with this VStack as content
	tabs := navigation.TabsBuilder().
		AddTab("elements", "Elements").
		Content("elements", innerVStack).
		AddTab("console", "Console").
		Content("console", ui.Text("Console content")).
		Height(21).  // This is what Inspector does
		Build()

	// Create layout engine
	engine := compute.NewEngine()

	// Layout the tabs with parent constraint
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 21,
	}

	layout, err := engine.Layout(tabs, nil, constraints)
	if err != nil {
		t.Fatalf("Layout failed: %v", err)
	}

	fmt.Printf("\n[TEST] Tabs layout size: %dx%d (constraint MaxHeight=21)\n",
		layout.Root.Box.Width, layout.Root.Box.Height)

	// Tabs should fit within constraint
	if layout.Root.Box.Height > 21 {
		t.Errorf("Tabs height %d exceeds constraint 21", layout.Root.Box.Height)
	}

	// Find the inner VStack in the layout
	fmt.Printf("[TEST] Searching for inner VStack in Tabs layout...\n")
	findVStackInLayout(layout.Root, 0, 21, t)
}

// findVStackInLayout finds VStack nodes and checks their constraints
func findVStackInLayout(box *compute.ComputedBox, depth int, maxHeight int, t *testing.T) bool {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	// Check if this is a VStack
	if box.VNode.Type().String() == "vstack" || box.VNode.Type().String() == "element" {
		props := box.VNode.Props()
		if props != nil {
			if h, ok := props["height"].(int); ok {
				fmt.Printf("[TEST]%s✓ Found VStack with Height(%d) prop at depth %d\n",
					indent, h, depth)
				fmt.Printf("[TEST]%s  Actual size: %dx%d\n",
					indent, box.Box.Width, box.Box.Height)

				// VStack should respect its Height prop
				if box.Box.Height != h {
					t.Errorf("[TEST]%s  ✗ FAIL: VStack has Height(%d) prop but measured height is %d",
						indent, h, box.Box.Height)
				} else {
					fmt.Printf("[TEST]%s  ✓ PASS: VStack height matches Height prop\n", indent)
				}

				return true
			}
		}
	}

	// Recursively check children
	for _, child := range box.Children {
		if findVStackInLayout(child, depth+1, maxHeight, t) {
			return true
		}
	}

	return false
}

// TestInspectorElementsTabDirectly tests the Elements tab content directly
func TestInspectorElementsTabDirectly(t *testing.T) {
	os.Setenv("TUI_DEBUG_LAYOUT", "true")

	insp := NewStandaloneInspector()
	insp.overlayWidth = 80
	insp.overlayHeight = 25

	// Build just the Elements tab content
	elementsContent := insp.buildElementsTabContent()

	fmt.Printf("\n[TEST] Testing Elements tab content directly\n")
	fmt.Printf("[TEST] Content type: %s\n", elementsContent.Type())

	// Check Height prop
	props := elementsContent.Props()
	heightProp, hasHeight := props["height"].(int)
	fmt.Printf("[TEST] Elements content has Height prop: %v, value: %d\n", hasHeight, heightProp)

	if !hasHeight {
		t.Error("Elements content should have Height prop")
	}

	// Measure with constraints simulating what Tabs would pass
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  76,
		MinHeight: 0,
		MaxHeight: 20,  // availableHeight from Inspector
	}

	// Measure using compute engine
	engine := compute.NewEngine()
	layout, err := engine.Layout(elementsContent, nil, constraints)
	if err != nil {
		t.Fatalf("Layout failed: %v", err)
	}

	fmt.Printf("[TEST] Elements content measured size: %dx%d\n",
		layout.Root.Box.Width, layout.Root.Box.Height)

	// Should respect the constraint
	if layout.Root.Box.Height > 20 {
		t.Errorf("Elements content height %d exceeds constraint 20", layout.Root.Box.Height)
	}

	// Traverse to check children
	fmt.Printf("[TEST] Traversing Elements content children:\n")
	for i, child := range layout.Root.Children {
		fmt.Printf("[TEST]   Child %d: type=%-15s size=%dx%d\n",
			i, child.VNode.Type(), child.Box.Width, child.Box.Height)

		// No child should exceed parent height
		if child.Box.Height > 20 {
			t.Errorf("Child %d height %d exceeds parent height 20", i, child.Box.Height)
		}
	}
}

// TestTreeViewInConstrainedVStack tests TreeView inside VStack with height constraint
func TestTreeViewInConstrainedVStack(t *testing.T) {
	os.Setenv("TUI_DEBUG_LAYOUT", "true")

	// Create TreeView with many lines
	treeLines := []string{
		"Root",
		"├── Child 1",
		"│   ├── Grandchild 1.1",
		"│   └── Grandchild 1.2",
		"├── Child 2",
		"│   ├── Grandchild 2.1",
		"│   └── Grandchild 2.2",
		"├── Child 3",
		"├── Child 4",
		"├── Child 5",
		"├── Child 6",
		"├── Child 7",
		"├── Child 8",
		"├── Child 9",
		"└── Child 10",
	}

	treeView := display.NewTreeView().
		FromLines(treeLines).
		Build()

	// Put TreeView in VStack with height constraint (like Inspector does)
	vstack := ui.VStackBuilder(
		ui.Text("📦 Layout Tree"),
		ui.Text("Nodes: 11 | Depth: 3"),
		ui.Text(""),
		ui.Text("────────────────────────"),
		treeView,  // TreeView should receive bounded height constraint
		ui.Text(""),
		ui.Text("Instructions: Navigate with arrow keys"),
	).
		Width(76).
		Height(20).  // This is the key - VStack has explicit height
		Build()

	// Layout with engine
	engine := compute.NewEngine()
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  76,
		MinHeight: 0,
		MaxHeight: 20,
	}

	layout, err := engine.Layout(vstack, nil, constraints)
	if err != nil {
		t.Fatalf("Layout failed: %v", err)
	}

	fmt.Printf("\n[TEST] VStack layout: %dx%d (constraint MaxHeight=20)\n",
		layout.Root.Box.Width, layout.Root.Box.Height)

	if layout.Root.Box.Height != 20 {
		t.Errorf("VStack height = %d, want 20", layout.Root.Box.Height)
	}

	// Find TreeView and check its size
	fmt.Printf("[TEST] Checking TreeView size...\n")
	treeViewFound := false
	for i, child := range layout.Root.Children {
		if child.VNode.Type().String() == "element" {
			if _, ok := child.VNode.(*display.TreeView); ok {
				treeViewFound = true
				fmt.Printf("[TEST]   Child %d (TreeView): size=%dx%d\n",
					i, child.Box.Width, child.Box.Height)

				// TreeView should be constrained by VStack
				// It has ~18 lines available: 20 total - header(2) - text(1) - instructions(1)
				if child.Box.Height > 20 {
					t.Errorf("TreeView height %d exceeds VStack height 20", child.Box.Height)
				}

				// TreeView should be using virtual scrolling
				// (height should be significantly less than total lines)
				if child.Box.Height >= 11 {
					t.Logf("TreeView height %d suggests virtual scrolling is NOT working (total lines: 11)",
						child.Box.Height)
				} else {
					t.Logf("TreeView height %d suggests virtual scrolling IS working", child.Box.Height)
				}
			}
		}
	}

	if !treeViewFound {
		t.Error("TreeView not found in VStack children")
	}
}

