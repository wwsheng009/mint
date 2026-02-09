package ui

import (
	"testing"

	"github.com/wwsheng009/mint/runtime"
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

// TestVStackPropagatesHeightConstraints tests that VStack propagates height constraints to children
func TestVStackPropagatesHeightConstraints(t *testing.T) {
	// Create a VStack with explicit height using builder
	vstack := VStackBuilder(
		Text("Line 1"),
		Text("Line 2"),
		Text("Line 3"),
	).Height(10).Build()

	// Measure with bounded constraints
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 10,
	}

	// The VStack should implement Measurable interface
	measurable, ok := vstack.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	})
	if !ok {
		t.Fatal("VStack should implement Measurable interface")
	}

	size := measurable.Measure(constraints)

	// VStack should respect the height constraint
	if size.Height != 10 {
		t.Errorf("VStack height = %d, want 10 (should respect height constraint)", size.Height)
	}
}

// TestVStackWithNonFlexChildrenRespectsHeight tests VStack constrains non-flex children
func TestVStackWithNonFlexChildrenRespectsHeight(t *testing.T) {
	// Create a VStack with Height prop that's smaller than natural height
	// Natural height would be 3 lines, but we constrain to 2
	vstack := VStackBuilder(
		Text("Line 1"),
		Text("Line 2"),
		Text("Line 3"),
	).Height(2).Build()

	// The VStack has Height(2) prop, which should constrain children
	measurable, ok := vstack.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	})
	if !ok {
		t.Fatal("VStack should implement Measurable interface")
	}

	// Measure with unbounded constraints - the Height(2) prop should still apply
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: runtime.Infinity,
	}

	size := measurable.Measure(constraints)

	// VStack should return height=2 due to Height prop
	if size.Height != 2 {
		t.Errorf("VStack height = %d, want 2 (Height prop should constrain)", size.Height)
	}
}

// TestHStackPropagatesHeightConstraints tests that HStack propagates height constraints to children
func TestHStackPropagatesHeightConstraints(t *testing.T) {
	// Create an HStack with explicit height
	hstack := HStackBuilder(
		Text("A"),
		Text("B"),
	).Height(5).Build()

	measurable, ok := hstack.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	})
	if !ok {
		t.Fatal("HStack should implement Measurable interface")
	}

	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  runtime.Infinity,
		MinHeight: 0,
		MaxHeight: 5,
	}

	size := measurable.Measure(constraints)

	// HStack should respect the height constraint
	if size.Height != 5 {
		t.Errorf("HStack height = %d, want 5", size.Height)
	}
}

// TestNestedVStackPropagatesConstraints tests nested VStack constraint propagation
func TestNestedVStackPropagatesConstraints(t *testing.T) {
	// Outer VStack with height constraint
	outer := VStackBuilder(
		// Inner VStack without explicit height
		VStackBuilder(
			Text("A"),
			Text("B"),
		).Build(),
	).Height(8).Build()

	measurable, ok := outer.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	})
	if !ok {
		t.Fatal("VStack should implement Measurable interface")
	}

	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 8,
	}

	size := measurable.Measure(constraints)

	// Outer VStack should respect height constraint
	if size.Height != 8 {
		t.Errorf("Outer VStack height = %d, want 8", size.Height)
	}
}

// TestVStackWidthConstraints tests VStack width constraint propagation
func TestVStackWidthConstraints(t *testing.T) {
	// Create VStack with width constraint
	vstack := VStackBuilder(
		Text("Short"),
		Text("Very long text that should be constrained"),
	).Width(20).Build()

	measurable, ok := vstack.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	})
	if !ok {
		t.Fatal("VStack should implement Measurable interface")
	}

	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  20,
		MinHeight: 0,
		MaxHeight: runtime.Infinity,
	}

	size := measurable.Measure(constraints)

	// VStack should respect width constraint
	if size.Width != 20 {
		t.Errorf("VStack width = %d, want 20", size.Width)
	}
}

