package ui

import (
	"testing"
)

// TestVStack tests VStack layout
func TestVStack(t *testing.T) {
	child1 := NewText("Child 1")
	child2 := NewText("Child 2")
	child3 := NewText("Child 3")

	vstack := VStack(child1, child2, child3)

	// Should be a VNode
	if vstack == nil {
		t.Fatal("VStack() returned nil")
	}

	// Check it's a LayoutNode
	if _, ok := vstack.(*LayoutNode); !ok {
		t.Error("VStack() should return *LayoutNode")
	}

	layout := vstack.(*LayoutNode)
	if layout.direction != DirectionColumn {
		t.Errorf("VStack direction = %v, want %v", layout.direction, DirectionColumn)
	}

	// Check children
	children := layout.Children()
	if len(children) != 3 {
		t.Errorf("len(Children()) = %v, want 3", len(children))
	}
}

// TestHStack tests HStack layout
func TestHStack(t *testing.T) {
	child1 := NewText("Child 1")
	child2 := NewText("Child 2")

	hstack := HStack(child1, child2)

	if _, ok := hstack.(*LayoutNode); !ok {
		t.Error("HStack() should return *LayoutNode")
	}

	layout := hstack.(*LayoutNode)
	if layout.direction != DirectionRow {
		t.Errorf("HStack direction = %v, want %v", layout.direction, DirectionRow)
	}
}

// TestLayoutBuilder tests layout builder pattern
func TestLayoutBuilder(t *testing.T) {
	child1 := NewText("Child 1")
	child2 := NewText("Child 2")

	builder := &LayoutBuilder{
		node: &LayoutNode{
			ElementVNode: NewElement("vstack"),
			direction:    DirectionColumn,
		},
		children: []VNode{child1, child2},
	}

	// Test Gap
	result := builder.Gap(2)
	if result.node.gap != 2 {
		t.Errorf("Gap(2) = %v, want 2", result.node.gap)
	}

	// Test Padding
	result = builder.Padding(1, 2, 3, 4)
	padding := result.node.Padding()
	if len(padding) != 4 {
		t.Errorf("Padding length = %v, want 4", len(padding))
	}
	if padding[0] != 1 || padding[1] != 2 || padding[2] != 3 || padding[3] != 4 {
		t.Errorf("Padding = %v, want [1, 2, 3, 4]", padding)
	}

	// Test Build
	vnode := builder.Build()
	if vnode == nil {
		t.Error("Build() returned nil")
	}
}

// TestLayoutGap tests layout gap
func TestLayoutGap(t *testing.T) {
	layout := &LayoutNode{
		ElementVNode: NewElement("vstack"),
		direction:    DirectionColumn,
		gap:          5,
	}

	if layout.Gap() != 5 {
		t.Errorf("Gap() = %v, want 5", layout.Gap())
	}
}

// TestLayoutPadding tests layout padding
func TestLayoutPadding(t *testing.T) {
	layout := &LayoutNode{
		ElementVNode: NewElement("vstack"),
		direction:    DirectionColumn,
		padding:      [4]int{1, 2, 3, 4},
	}

	padding := layout.Padding()
	if padding[0] != 1 || padding[1] != 2 || padding[2] != 3 || padding[3] != 4 {
		t.Errorf("Padding() = %v, want [1, 2, 3, 4]", padding)
	}
}

// TestNestedLayouts tests nested layout structures
func TestNestedLayouts(t *testing.T) {
	inner1 := VStack(NewText("A"), NewText("B"))
	inner2 := HStack(NewText("C"), NewText("D"))
	outer := VStack(inner1, inner2)

	if outer == nil {
		t.Fatal("VStack() returned nil")
	}

	children := outer.Children()
	if len(children) != 2 {
		t.Errorf("len(Children()) = %v, want 2", len(children))
	}

	// First child should be a VStack
	if _, ok := children[0].(*LayoutNode); !ok {
		t.Error("First child should be *LayoutNode")
	}

	// Second child should be an HStack
	if _, ok := children[1].(*LayoutNode); !ok {
		t.Error("Second child should be *LayoutNode")
	}
}

// TestLayoutWithEmptyChildren tests layout with no children
func TestLayoutWithEmptyChildren(t *testing.T) {
	vstack := VStack()

	if vstack == nil {
		t.Fatal("VStack() returned nil")
	}

	children := vstack.Children()
	if len(children) != 0 {
		t.Errorf("len(Children()) = %v, want 0", len(children))
	}
}

// BenchmarkVStack benchmarks VStack creation
func BenchmarkVStack(b *testing.B) {
	child1 := NewText("Child 1")
	child2 := NewText("Child 2")
	child3 := NewText("Child 3")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = VStack(child1, child2, child3)
	}
}

// BenchmarkHStack benchmarks HStack creation
func BenchmarkHStack(b *testing.B) {
	child1 := NewText("Child 1")
	child2 := NewText("Child 2")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = HStack(child1, child2)
	}
}

// BenchmarkNestedLayouts benchmarks nested layout creation
func BenchmarkNestedLayouts(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = VStack(
			HStack(NewText("A"), NewText("B")),
			HStack(NewText("C"), NewText("D")),
		)
	}
}
