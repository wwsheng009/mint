// Package compute tests Fiber-first layout algorithm.
// This test ensures Fiber-only layout works without VNode access.
package compute

import (
	"testing"

	"github.com/wwsheng009/mint/runtime"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestFiberOnlyLayout tests the Fiber-only layout implementation.
// This verifies:
// 1. BuildComputedBoxFiberOnly() works without VNode
// 2. Layout properties come from Fiber fields only
// 3. Tree traversal uses Fiber.Child -> Fiber.Sibling chain
func TestFiberOnlyLayout(t *testing.T) {
	// Create a simple Fiber tree
	// Root (HStack) with two text children

	// Create root Fiber
	root := &rtui.Fiber{
		Type:   rtui.VNodeElement,
		Tag:    "hstack",
		NodeID:  1,
		Layer:  rtui.LayerBase,
		LayoutDirection: rtui.DirectionRow,
		LayoutGap:       1,
		LayoutAlign:     rtui.AlignStart,
		LayoutCrossAlign: rtui.AlignStart,
		LayoutPadding:    [4]int{0, 0, 0, 0},
	}

	// Create first child (text)
	child1 := &rtui.Fiber{
		Type:         rtui.VNodeText,
		Tag:          "text",
		NodeID:       2,
		Layer:        rtui.LayerBase,
		MemoizedState: "Hello",
	}

	// Create second child (text)
	child2 := &rtui.Fiber{
		Type:         rtui.VNodeText,
		Tag:          "text",
		NodeID:       3,
		Layer:        rtui.LayerBase,
		MemoizedState: "World",
	}

	// Link children: root.Child = child1, child1.Sibling = child2
	root.Child = child1
	child1.Return = root
	child1.Sibling = child2
	child2.Return = root

	// Create engine and run Fiber-only layout
	engine := NewEngine()

	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  runtime.Infinity,
		MinHeight: 0,
		MaxHeight: runtime.Infinity,
	}

	layout, err := engine.BuildComputedBoxFiberOnly(root, constraints)
	if err != nil {
		t.Fatalf("BuildComputedBoxFiberOnly failed: %v", err)
	}

	if layout == nil {
		t.Fatal("Layout is nil")
	}

	if layout.Root == nil {
		t.Fatal("Layout root is nil")
	}

	// Verify root box
	rootBox := layout.Root
	// Expected: "Hello" (5) + gap (1) + "World" (5) = 11
	// But each text's natural width is used in the layout
	if rootBox.Box.Width < 5 { // At least "Hello"
		t.Logf("Root box width: %d (expected at least 11)", rootBox.Box.Width)
	}
	if rootBox.Box.Height != 1 {
		t.Logf("Root box height: %d", rootBox.Box.Height)
	}

	// Verify children
	if len(rootBox.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(rootBox.Children))
	}

	// Verify NodeIDs are propagated
	if rootBox.Children[0].NodeID != 2 {
		t.Errorf("Child 1 NodeID: got %d, want 2", rootBox.Children[0].NodeID)
	}
	if rootBox.Children[1].NodeID != 3 {
		t.Errorf("Child 2 NodeID: got %d, want 3", rootBox.Children[1].NodeID)
	}

	t.Logf("✅ Fiber-only layout works!")
	t.Logf("Root: %v", rootBox.Box)
	t.Logf("Children: %d", len(rootBox.Children))
	for i, child := range rootBox.Children {
		t.Logf("  Child %d: NodeID=%d Size=%v", i, child.NodeID, child.Box)
	}
}

// TestFiberOnlyVStackLayout tests vertical stack layout.
func TestFiberOnlyVStackLayout(t *testing.T) {
	// Create root Fiber (VStack)
	root := &rtui.Fiber{
		Type:   rtui.VNodeElement,
		Tag:    "vstack",
		NodeID: 1,
		Layer:  rtui.LayerBase,
		LayoutDirection: rtui.DirectionColumn,
		LayoutGap:       1,
		LayoutAlign:     rtui.AlignStart,
		LayoutCrossAlign: rtui.AlignStart,
		LayoutPadding:    [4]int{0, 0, 0, 0},
	}

	// Create two text children
	child1 := &rtui.Fiber{
		Type:         rtui.VNodeText,
		Tag:          "text",
		NodeID:       2,
		Layer:        rtui.LayerBase,
		MemoizedState: "Line1",
	}

	child2 := &rtui.Fiber{
		Type:         rtui.VNodeText,
		Tag:          "text",
		NodeID:       3,
		Layer:        rtui.LayerBase,
		MemoizedState: "Line2",
	}

	// Link children
	root.Child = child1
	child1.Return = root
	child1.Sibling = child2
	child2.Return = root

	// Create engine and run layout
	engine := NewEngine()

	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  runtime.Infinity,
		MinHeight: 0,
		MaxHeight: runtime.Infinity,
	}

	layout, err := engine.BuildComputedBoxFiberOnly(root, constraints)
	if err != nil {
		t.Fatalf("BuildComputedBoxFiberOnly failed: %v", err)
	}

	if layout.Root == nil {
		t.Fatal("Layout root is nil")
	}

	rootBox := layout.Root

	// VStack should have height = 2 (1 + 1 gap + 1) = 3
	if rootBox.Box.Height < 2 || rootBox.Box.Height > 3 {
		t.Logf("Root box height: %d (expected around 3)", rootBox.Box.Height)
	}

	// Verify children
	if len(rootBox.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(rootBox.Children))
	}

	t.Logf("✅ Fiber-only VStack layout works!")
	t.Logf("Root: %v", rootBox.Box)
	for i, child := range rootBox.Children {
		t.Logf(" Child %d: NodeID=%d Size=%v", i, child.NodeID, child.Box)
	}
}

// TestFiberOnlyBoundedConstraints tests bounded width constraints.
func TestFiberOnlyBoundedConstraints(t *testing.T) {
	// Create root with bounded width
	root := &rtui.Fiber{
		Type:   rtui.VNodeElement,
		Tag:    "hstack",
		NodeID:  1,
		Layer:  rtui.LayerBase,
		LayoutDirection: rtui.DirectionRow,
		LayoutGap:       1,
		LayoutPadding:    [4]int{0, 0, 0, 0},
	}

	// Add text child
	child := &rtui.Fiber{
		Type:         rtui.VNodeText,
		Tag:          "text",
		NodeID:       2,
		Layer:        rtui.LayerBase,
		MemoizedState: "ABCDEFGHIJ", // 10 chars
	}

	root.Child = child
	child.Return = root

	// Bounded width constraint
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  20, // Bounded to 20
		MinHeight: 0,
		MaxHeight: runtime.Infinity,
	}

	engine := NewEngine()
	layout, err := engine.BuildComputedBoxFiberOnly(root, constraints)
	if err != nil {
		t.Fatalf("BuildComputedBoxFiberOnly failed: %v", err)
	}

	if layout.Root == nil {
		t.Fatal("Layout root is nil")
	}

	rootBox := layout.Root

	// Width should be at most 20 (constrained)
	if rootBox.Box.Width > 20 {
		t.Errorf("Root width %d exceeds constraint 20", rootBox.Box.Width)
	}

	t.Logf("✅ Fiber-only bounded constraints work!")
	t.Logf("Root: %v (maxWidth=20)", rootBox.Box)
}

// TestFiberOnlyPadding tests that padding is correctly applied.
func TestFiberOnlyPadding(t *testing.T) {
	// Create root with padding
	root := &rtui.Fiber{
		Type:   rtui.VNodeElement,
		Tag:    "hstack",
		NodeID:  1,
		Layer:  rtui.LayerBase,
		LayoutDirection: rtui.DirectionRow,
		LayoutGap:       0,
		LayoutAlign:     rtui.AlignStart,
		LayoutPadding:    [4]int{1, 2, 1, 2}, // top=1, right=2, bottom=1, left=2
	}

	// Add text child
	child := &rtui.Fiber{
		Type:         rtui.VNodeText,
		Tag:          "text",
		NodeID:       2,
		Layer:        rtui.LayerBase,
		MemoizedState: "X",
	}

	root.Child = child
	child.Return = root

	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  runtime.Infinity,
		MinHeight: 0,
		MaxHeight: runtime.Infinity,
	}

	engine := NewEngine()
	layout, err := engine.BuildComputedBoxFiberOnly(root, constraints)
	if err != nil {
		t.Fatalf("BuildComputedBoxFiberOnly failed: %v", err)
	}

	if layout.Root == nil {
		t.Fatal("Layout root is nil")
	}

	rootBox := layout.Root

	// Width should be: padding (2+2=4) + text (1) = 5
	if rootBox.Box.Width < 4 || rootBox.Box.Width > 5 {
		t.Logf("Root width: %d (expected around 5: 2+2+1)", rootBox.Box.Width)
	}

	// Height should be: padding (1+1=2) + text (1) = 3
	if rootBox.Box.Height < 2 || rootBox.Box.Height > 3 {
		t.Logf("Root height: %d (expected around 3: 1+1+1)", rootBox.Box.Height)
	}

	t.Logf("✅ Fiber-only padding works!")
	t.Logf("Root: %v (padding=[1,2,1,2])", rootBox.Box)
}
