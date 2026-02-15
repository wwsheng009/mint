// Package compute provides integration tests for Fiber-first layout
package compute

import (
	"testing"

	"github.com/wwsheng009/mint/runtime"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Fiber-First Layout Integration Tests
// =============================================================================

// TestFiberVsVNodeLayout compares Fiber-only layout with VNode layout
// to ensure both paths produce identical results
func TestFiberVsVNodeLayout(t *testing.T) {
	engine := NewEngine()

	// Create a simple VNode tree using basic elements
	// VStack with 3 text children
	vnodeTree := rtui.Element("vstack").
		Children(
			rtui.Element("text").Prop("content", "A").Build(),
			rtui.Element("text").Prop("content", "B").Build(),
			rtui.Element("text").Prop("content", "C").Build(),
		).Build()

	// Create Fiber tree from VNode tree
	fiberTree := rtui.CreateFiberFromVNode(vnodeTree)

	constraints := runtime.BoxConstraints{
		MinWidth:  0, MinHeight: 0,
		MaxWidth: 100, MaxHeight: 100,
	}

	// Layout using VNode path (existing)
	vnodeLayout, err := engine.Layout(vnodeTree, fiberTree, constraints)
	if err != nil {
		t.Fatalf("VNode layout failed: %v", err)
	}

	// Layout using Fiber-only path (new)
	fiberLayout, err := engine.LayoutFiber(fiberTree, constraints)
	if err != nil {
		t.Fatalf("Fiber layout failed: %v", err)
	}

	// Compare root box dimensions
	if vnodeLayout.Root.Box.Width != fiberLayout.Root.Box.Width {
		t.Errorf("Width mismatch: VNode=%d, Fiber=%d",
			vnodeLayout.Root.Box.Width, fiberLayout.Root.Box.Width)
	}
	if vnodeLayout.Root.Box.Height != fiberLayout.Root.Box.Height {
		t.Errorf("Height mismatch: VNode=%d, Fiber=%d",
			vnodeLayout.Root.Box.Height, fiberLayout.Root.Box.Height)
	}

	// Compare children count
	vnodeChildren := countDescendantsBox(vnodeLayout.Root)
	fiberChildren := countDescendantsBox(fiberLayout.Root)

	if vnodeChildren != fiberChildren {
		t.Errorf("Children count mismatch: VNode=%d, Fiber=%d",
			vnodeChildren, fiberChildren)
	}

	t.Logf("✅ Fiber-first layout matches VNode layout")
	t.Logf("   Dimensions: %dx%d", vnodeLayout.Root.Box.Width, vnodeLayout.Root.Box.Height)
	t.Logf("   Total boxes: %d", vnodeChildren+1)
}

// TestFiberVsVNodeComplexLayout tests a more complex layout structure
func TestFiberVsVNodeComplexLayout(t *testing.T) {
	engine := NewEngine()

	// Create a complex nested layout
	vnodeTree := rtui.Element("vstack").
		Children(
			rtui.Element("text").Prop("content", "Header").Build(),
			rtui.Element("hstack").
				Children(
					rtui.Element("text").Prop("content", "L").Build(),
					rtui.Element("text").Prop("content", "R").Build(),
				).Build(),
			rtui.Element("text").Prop("content", "Footer").Build(),
		).Build()

	fiberTree := rtui.CreateFiberFromVNode(vnodeTree)

	constraints := runtime.BoxConstraints{
		MinWidth:  0, MinHeight: 0,
		MaxWidth: 80, MaxHeight: 100,
	}

	vnodeLayout, err := engine.Layout(vnodeTree, fiberTree, constraints)
	if err != nil {
		t.Fatalf("VNode layout failed: %v", err)
	}

	fiberLayout, err := engine.LayoutFiber(fiberTree, constraints)
	if err != nil {
		t.Fatalf("Fiber layout failed: %v", err)
	}

	// Compare root dimensions
	if vnodeLayout.Root.Box.Width != fiberLayout.Root.Box.Width {
		t.Errorf("Complex layout width mismatch: VNode=%d, Fiber=%d",
			vnodeLayout.Root.Box.Width, fiberLayout.Root.Box.Width)
	}
	if vnodeLayout.Root.Box.Height != fiberLayout.Root.Box.Height {
		t.Errorf("Complex layout height mismatch: VNode=%d, Fiber=%d",
			vnodeLayout.Root.Box.Height, fiberLayout.Root.Box.Height)
	}

	totalVNodeBoxes := countDescendantsBox(vnodeLayout.Root)
	totalFiberBoxes := countDescendantsBox(fiberLayout.Root)

	if totalVNodeBoxes != totalFiberBoxes {
		t.Errorf("Complex layout children count mismatch: VNode=%d, Fiber=%d",
			totalVNodeBoxes, totalFiberBoxes)
	}

	t.Logf("✅ Complex Fiber-first layout matches VNode layout")
	t.Logf("   Dimensions: %dx%d", vnodeLayout.Root.Box.Width, vnodeLayout.Root.Box.Height)
	t.Logf("   Total boxes: %d", totalVNodeBoxes+1)
}

// TestFiberVsVNodeHStackLayout tests horizontal layout
func TestFiberVsVNodeHStackLayout(t *testing.T) {
	engine := NewEngine()

	// HStack with multiple children
	vnodeTree := rtui.Element("hstack").
		Children(
			rtui.Element("text").Prop("content", "A").Build(),
			rtui.Element("text").Prop("content", "B").Build(),
			rtui.Element("text").Prop("content", "C").Build(),
			rtui.Element("text").Prop("content", "D").Build(),
		).Build()

	fiberTree := rtui.CreateFiberFromVNode(vnodeTree)

	constraints := runtime.BoxConstraints{
		MinWidth:  0, MinHeight: 0,
		MaxWidth: 100, MaxHeight: 50,
	}

	vnodeLayout, err := engine.Layout(vnodeTree, fiberTree, constraints)
	if err != nil {
		t.Fatalf("VNode layout failed: %v", err)
	}

	fiberLayout, err := engine.LayoutFiber(fiberTree, constraints)
	if err != nil {
		t.Fatalf("Fiber layout failed: %v", err)
	}

	// For HStack, width should be sum of children + gaps
	// Both paths should produce same result
	if vnodeLayout.Root.Box.Width != fiberLayout.Root.Box.Width {
		t.Errorf("HStack width mismatch: VNode=%d, Fiber=%d",
			vnodeLayout.Root.Box.Width, fiberLayout.Root.Box.Width)
	}

	t.Logf("✅ HStack Fiber-first layout matches VNode layout")
	t.Logf("   Dimensions: %dx%d", vnodeLayout.Root.Box.Width, vnodeLayout.Root.Box.Height)
}

// =============================================================================
// Helper Functions
// =============================================================================

// countDescendants counts all descendant boxes
func countDescendants(box *ComputedLayout) int {
	if box == nil || box.Root == nil {
		return 0
	}
	return countDescendantsBox(box.Root) - 1 // Exclude root
}

// countDescendantsBox recursively counts boxes in tree
func countDescendantsBox(box *ComputedBox) int {
	if box == nil {
		return 0
	}
	count := 1
	for _, child := range box.Children {
		count += countDescendantsBox(child)
	}
	return count
}
