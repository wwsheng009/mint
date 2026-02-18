package compute

import (
	"testing"

	"github.com/wwsheng009/mint/runtime"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestFiberOnlyLayoutDebug diagnoses the Fiber-only layout issue
func TestFiberOnlyLayoutDebug(t *testing.T) {
	// Create a simple VStack with text children (similar to modal content)
	text1 := rtui.NewElement("text")
	text1.SetProps(rtui.Props{"content": "=== MODAL START ==="})

	text2 := rtui.NewElement("text")
	text2.SetProps(rtui.Props{"content": "*** Are you sure? ***"})

	vstack := rtui.NewElement("vstack")
	vstack.SetProps(rtui.Props{
		"direction": "column",
		"gap":       0,
	})
	vstack.SetChildren([]rtui.VNode{text1, text2})

	// Create bordered container
	bordered := rtui.NewElement("bordered")
	bordered.SetProps(rtui.Props{"width": 40})
	bordered.SetChildren([]rtui.VNode{vstack})
	bordered.SetLayer(rtui.LayerModal)

	// Create Fiber tree
	fiber := rtui.CreateFiberFromVNode(bordered)

	// Debug: print Fiber tree
	t.Log("=== Fiber Tree ===")
	printFiberTree(t, fiber, 0)
	
	// Debug: test MeasureChild for text nodes
	t.Log("=== Testing MeasureChild ===")
	engine := NewEngine()
	
	vstackFiber := fiber.Child // bordered -> vstack
	text1Fiber := vstackFiber.Child // vstack -> text1
	text2Fiber := text1Fiber.Sibling
	
	constraints := runtime.BoxConstraints{MaxWidth: 80, MaxHeight: 24}
	
	// Measure text nodes
	size1 := engine.MeasureChild(text1Fiber, constraints)
	t.Logf("text1 MeasureChild: %dx%d", size1.Width, size1.Height)
	
	size2 := engine.MeasureChild(text2Fiber, constraints)
	t.Logf("text2 MeasureChild: %dx%d", size2.Width, size2.Height)
	
	// Measure vstack
	vstackConstraints := runtime.BoxConstraints{MaxWidth: 38, MaxHeight: 22} // bordered inner constraints
	vstackSize := engine.MeasureChild(vstackFiber, vstackConstraints)
	t.Logf("vstack MeasureChild: %dx%d", vstackSize.Width, vstackSize.Height)

	// Build ComputedBox using BuildComputedBoxFiberOnly
	layoutConstraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 24,
	}

	layout, err := engine.BuildComputedBoxFiberOnly(fiber, layoutConstraints)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	t.Log("=== ComputedBox Tree ===")
	printBoxTree(t, layout.Root, 0)

	// Verify bordered box has correct size
	if layout.Root.Box.Width == 0 {
		t.Error("Bordered box width is 0, expected > 0")
	}
	if layout.Root.Box.Height == 0 {
		t.Error("Bordered box height is 0, expected > 0")
	}
	if layout.Root.Box.Height < 4 {
		t.Errorf("Bordered box height is %d, expected >= 4 (2 texts + 2 borders)", layout.Root.Box.Height)
	}

	// Verify children exist
	if len(layout.Root.Children) == 0 {
		t.Error("Bordered box has no children")
	}
}

func printFiberTree(t *testing.T, f *rtui.Fiber, depth int) {
	if f == nil {
		return
	}
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	t.Logf("%sFiber: NodeID=%d Tag=%s Layer=%d Props=%v MemoizedState=%v",
		indent, f.NodeID, f.Tag, f.Layer, f.Props, f.MemoizedState)

	for child := f.Child; child != nil; child = child.Sibling {
		printFiberTree(t, child, depth+1)
	}
}

func printBoxTree(t *testing.T, box *ComputedBox, depth int) {
	if box == nil {
		return
	}
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	t.Logf("%sBox: NodeID=%d Layer=%d Size=%dx%d Pos=(%d,%d)",
		indent, box.NodeID, box.Layer, box.Box.Width, box.Box.Height, box.Box.X, box.Box.Y)

	for _, child := range box.Children {
		printBoxTree(t, child, depth+1)
	}
}
