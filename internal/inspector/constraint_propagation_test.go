package inspector

import (
	"fmt"
	"os"
	"testing"

	"github.com/wwsheng009/mint/components/navigation"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

// TestInspectorElementsTabVStackConstraints tests that the Elements tab VStack
// properly propagates height constraints to the TreeView
func TestInspectorElementsTabVStackConstraints(t *testing.T) {
	// Enable debug logging
	os.Setenv("TUI_DEBUG_LAYOUT", "true")

	// Create Inspector
	insp := NewStandaloneInspector()
	insp.overlayWidth = 80
	insp.overlayHeight = 25

	// Build Elements tab content (this contains the TreeView)
	elementsContent := insp.buildElementsTabContent()
	if elementsContent == nil {
		t.Fatal("buildElementsTabContent() returned nil")
	}

	fmt.Printf("[TEST] ElementsContent type: %s\n", elementsContent.Type())

	// The Elements tab content is a VStack with explicit height
	// Let's verify it has the height prop set
	props := elementsContent.Props()
	if props == nil {
		t.Fatal("Elements content has no props")
	}

	heightProp, hasHeightProp := props["height"].(int)
	fmt.Printf("[TEST] Elements content has height prop: %v, value: %d\n", hasHeightProp, heightProp)

	if !hasHeightProp || heightProp == 0 {
		t.Errorf("Elements content should have Height prop set, got: %d", heightProp)
	}

	// Now measure the VStack with constraints
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  76, // overlayWidth - 4
		MinHeight: 0,
		MaxHeight: 20, // availableHeight = 25 - 5 (titleBar + tabBar + separator)
	}

	// The VStack should implement Measurable
	measurable, ok := elementsContent.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	})
	if !ok {
		t.Fatal("Elements VStack should implement Measurable interface")
	}

	size := measurable.Measure(constraints)
	fmt.Printf("[TEST] Measured size: %dx%d (constraints: MaxWidth=%d, MaxHeight=%d)\n",
		size.Width, size.Height, constraints.MaxWidth, constraints.MaxHeight)

	// VStack should respect the height constraint
	if size.Height != 20 {
		t.Errorf("VStack height = %d, want 20 (should respect MaxHeight constraint)", size.Height)
	}

	// Now check if the TreeView child received bounded constraints
	// Find the TreeView in the VStack children
	children := elementsContent.Children()
	fmt.Printf("[TEST] VStack has %d children\n", len(children))

	var treeViewVNode rtui.VNode
	for i, child := range children {
		fmt.Printf("[TEST] Child %d: type=%s\n", i, child.Type())
		if child.Type().String() == "element" {
			// Check if it's a TreeView by checking for specific methods or props
			if _, hasTreeView := child.Props()["treeView"]; hasTreeView {
				treeViewVNode = child
				fmt.Printf("[TEST] Found TreeView at child %d\n", i)
				break
			}
		}
	}

	if treeViewVNode == nil {
		t.Skip("TreeView not found in VStack children (may need different detection method)")
		return
	}

	// Verify TreeView was measured with bounded constraints
	if measurableTV, ok := treeViewVNode.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	}); ok {
		// Create constraints simulating what VStack would pass to TreeView
		// TreeView is one of several children, so it gets a portion of the height
		treeViewConstraints := runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth:  76,
			MinHeight: 0,
			MaxHeight: 15, // Approximate: 20 total - header - other elements
		}

		tvSize := measurableTV.Measure(treeViewConstraints)
		fmt.Printf("[TEST] TreeView measured size: %dx%d (constraints MaxHeight=%d)\n",
			tvSize.Width, tvSize.Height, treeViewConstraints.MaxHeight)

		// TreeView should respect the height constraint
		if tvSize.Height != treeViewConstraints.MaxHeight {
			t.Errorf("TreeView height = %d, want %d (should respect MaxHeight)",
				tvSize.Height, treeViewConstraints.MaxHeight)
		}
	}
}

// TestInspectorLayoutEngineIntegration tests the full layout engine pipeline
func TestInspectorLayoutEngineIntegration(t *testing.T) {
	// Enable debug logging
	os.Setenv("TUI_DEBUG_LAYOUT", "true")

	// Create Inspector
	insp := NewStandaloneInspector()
	insp.overlayWidth = 80
	insp.overlayHeight = 25
	insp.floatX = 0
	insp.floatY = 0

	// Build the overlay content
	overlayContent := insp.buildOverlayContent()

	// Create layout engine
	engine := compute.NewEngine()

	// Measure the overlay
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

	fmt.Printf("[TEST] Overlay layout: root size = %dx%d\n",
		layout.Root.Box.Width, layout.Root.Box.Height)

	// The overlay should fit within constraints
	if layout.Root.Box.Width > 80 {
		t.Errorf("Overlay width %d exceeds constraint MaxWidth 80", layout.Root.Box.Width)
	}
	if layout.Root.Box.Height > 25 {
		t.Errorf("Overlay height %d exceeds constraint MaxHeight 25", layout.Root.Box.Height)
	}

	// Find the TreeView in the computed layout
	findTreeViewInLayout(layout.Root, 0, t)
}

// findTreeViewInLayout recursively searches for TreeView in computed layout
func findTreeViewInLayout(box *compute.ComputedBox, depth int, t *testing.T) bool {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	fmt.Printf("[TEST]%sBox: type=%s size=%dx%d pos=(%d,%d)\n",
		indent, box.GetVNode().Type(), box.Box.Width, box.Box.Height, box.Box.X, box.Box.Y)

	// Check if this is a TreeView by type or props
	if box.GetVNode().Type().String() == "element" {
		props := box.GetVNode().Props()
		if props != nil {
			if _, hasTreeView := props["treeView"]; hasTreeView {
				fmt.Printf("[TEST]%s→ Found TreeView! size=%dx%d\n", indent, box.Box.Width, box.Box.Height)

				// TreeView should have bounded size
				if box.Box.Height > 20 {
					t.Errorf("TreeView height %d is too large (should be constrained by parent)", box.Box.Height)
				}
				return true
			}
		}
	}

	// Search children
	for _, child := range box.Children {
		if findTreeViewInLayout(child, depth+1, t) {
			return true
		}
	}

	return false
}

// TestVStackPropagatesConstraintsToTreeView tests a simplified VStack+TreeView scenario
func TestVStackPropagatesConstraintsToTreeView(t *testing.T) {
	// Use a simple VStack with Text instead of TreeView for this test
	// (TreeView requires the display package which creates import cycles in tests)
	vstack := ui.VStackBuilder(
		ui.Text("Header"),
		ui.Text("Tree: Root"), // Placeholder for TreeView
		ui.Text("Tree: Child 1"),
		ui.Text("Tree: Child 2"),
		ui.Text("Tree: Child 3"),
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

	// VStack should respect the height constraint
	if size.Height != 10 {
		t.Errorf("VStack height = %d, want 10", size.Height)
	}

	fmt.Printf("[TEST] VStack with Height(10) prop measured as: %dx%d\n", size.Width, size.Height)

	// Now test with layout engine
	engine := compute.NewEngine()
	layout, err := engine.Layout(vstack, nil, constraints)
	if err != nil {
		t.Fatalf("Layout failed: %v", err)
	}

	fmt.Printf("[TEST] Layout engine result: %dx%d\n", layout.Root.Box.Width, layout.Root.Box.Height)

	if layout.Root.Box.Height != 10 {
		t.Errorf("Layout height = %d, want 10", layout.Root.Box.Height)
	}

	// Check that children are properly constrained
	fmt.Printf("[TEST] VStack has %d children in layout:\n", len(layout.Root.Children))
	for i, child := range layout.Root.Children {
		fmt.Printf("[TEST]   Child %d: size=%dx%d\n", i, child.Box.Width, child.Box.Height)

		// No single child should be taller than the container
		if child.Box.Height > 10 {
			t.Errorf("Child %d height %d exceeds container height 10", i, child.Box.Height)
		}
	}
}

// TestTabsInVStackConstraints tests Tabs inside VStack with height constraint
func TestTabsInVStackConstraints(t *testing.T) {
	// Create Tabs with content that exceeds constraint
	tabs := navigation.TabsBuilder().
		AddTab("tab1", "Tab 1").
		Content("tab1", ui.VStack(
			ui.Text("Line 1"),
			ui.Text("Line 2"),
			ui.Text("Line 3"),
			ui.Text("Line 4"),
			ui.Text("Line 5"),
			ui.Text("Line 6"),
			ui.Text("Line 7"),
			ui.Text("Line 8"),
			ui.Text("Line 9"),
			ui.Text("Line 10"),
		)).
		AddTab("tab2", "Tab 2").
		Content("tab2", ui.Text("Content 2")).
		Build()

	// Put Tabs inside VStack with height constraint
	vstack := ui.VStackBuilder(
		ui.Text("Header"),
		tabs,
		ui.Text("Footer"),
	).Height(15).Build()

	// Measure with layout engine
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 15,
	}

	engine := compute.NewEngine()
	layout, err := engine.Layout(vstack, nil, constraints)
	if err != nil {
		t.Fatalf("Layout failed: %v", err)
	}

	fmt.Printf("[TEST] VStack+Tabs layout: %dx%d\n", layout.Root.Box.Width, layout.Root.Box.Height)

	// VStack should fit in constraint
	if layout.Root.Box.Height > 15 {
		t.Errorf("VStack height %d exceeds constraint 15", layout.Root.Box.Height)
	}

	// Find Tabs in layout
	for i, child := range layout.Root.Children {
		if child.GetVNode().Type().String() == "tabs" {
			fmt.Printf("[TEST] Found Tabs at child %d: size=%dx%d\n", i, child.Box.Width, child.Box.Height)

			// Tabs should be constrained by VStack
			if child.Box.Height > 15 {
				t.Errorf("Tabs height %d exceeds parent VStack height 15", child.Box.Height)
			}

			// Tabs has tab bar (1) + content, so should be at least 2
			if child.Box.Height < 2 {
				t.Errorf("Tabs height %d is too small (should be at least 2)", child.Box.Height)
			}
		}
	}
}

