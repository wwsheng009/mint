package ui

import (
	"testing"
)

// TestVStack tests VStack layout
func TestVStack(t *testing.T) {
	child1 := Text("Child 1")
	child2 := Text("Child 2")
	child3 := Text("Child 3")

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
	if layout.Direction() != DirectionColumn {
		t.Errorf("VStack direction = %v, want %v", layout.Direction(), DirectionColumn)
	}

	// Check children
	children := layout.Children()
	if len(children) != 3 {
		t.Errorf("len(Children()) = %v, want 3", len(children))
	}
}

// TestHStack tests HStack layout
func TestHStack(t *testing.T) {
	child1 := Text("Child 1")
	child2 := Text("Child 2")

	hstack := HStack(child1, child2)

	if _, ok := hstack.(*LayoutNode); !ok {
		t.Error("HStack() should return *LayoutNode")
	}

	layout := hstack.(*LayoutNode)
	if layout.Direction() != DirectionRow {
		t.Errorf("HStack direction = %v, want %v", layout.Direction(), DirectionRow)
	}
}

// TestLayoutBuilder tests layout builder pattern
func TestLayoutBuilder(t *testing.T) {
	// Note: The builder is in runtime/ui package and not directly accessible
	// We test the public API instead by creating layouts with VStack/HStack
	child1 := Text("Child 1")
	child2 := Text("Child 2")

	vstack := VStack(child1, child2)
	if vstack == nil {
		t.Error("VStack() returned nil")
	}

	// Verify children were added correctly
	children := vstack.Children()
	if len(children) != 2 {
		t.Errorf("len(Children()) = %v, want 2", len(children))
	}
}

// TestNestedLayouts tests nested layout structures
func TestNestedLayouts(t *testing.T) {
	inner1 := VStack(Text("A"), Text("B"))
	inner2 := HStack(Text("C"), Text("D"))
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
	child1 := Text("Child 1")
	child2 := Text("Child 2")
	child3 := Text("Child 3")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = VStack(child1, child2, child3)
	}
}

// BenchmarkHStack benchmarks HStack creation
func BenchmarkHStack(b *testing.B) {
	child1 := Text("Child 1")
	child2 := Text("Child 2")

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
			HStack(Text("A"), Text("B")),
			HStack(Text("C"), Text("D")),
		)
	}
}
