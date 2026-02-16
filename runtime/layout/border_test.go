package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// BorderStyle Tests
// =============================================================================

func TestBorderStyle_String(t *testing.T) {
	tests := []struct {
		style    BorderStyle
		expected string
	}{
		{BorderNone, "none"},
		{BorderSingle, "single"},
		{BorderDouble, "double"},
		{BorderRounded, "rounded"},
		{BorderDashed, "dashed"},
		{BorderStyle(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.style.String())
		})
	}
}

func TestBorderStyle_HasBorder(t *testing.T) {
	assert.False(t, BorderNone.HasBorder())
	assert.True(t, BorderSingle.HasBorder())
	assert.True(t, BorderDouble.HasBorder())
	assert.True(t, BorderRounded.HasBorder())
	assert.True(t, BorderDashed.HasBorder())
}

// =============================================================================
// Border Tests
// =============================================================================

func TestNewBorder(t *testing.T) {
	tests := []struct {
		name         string
		style        BorderStyle
		expectWidth  int
		expectBorder bool
	}{
		{"none", BorderNone, 0, false},
		{"single", BorderSingle, 1, true},
		{"double", BorderDouble, 2, true},
		{"rounded", BorderRounded, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			border := NewBorder(tt.style)
			assert.Equal(t, tt.style, border.Style)
			assert.Equal(t, tt.expectWidth, border.Width)
			assert.Equal(t, tt.expectBorder, border.HasBorder())
		})
	}
}

func TestNewBorderWithLabel(t *testing.T) {
	border := NewBorderWithLabel(BorderSingle, "Test Label")
	assert.Equal(t, BorderSingle, border.Style)
	assert.Equal(t, "Test Label", border.Label)
	assert.True(t, border.HasBorder())
}

func TestBorder_HasBorder(t *testing.T) {
	assert.False(t, Border{Style: BorderNone}.HasBorder())
	assert.True(t, Border{Style: BorderSingle}.HasBorder())
}

func TestBorder_Padding(t *testing.T) {
	tests := []struct {
		name              string
		border            Border
		expectHorizontal  int
		expectVertical    int
	}{
		{
			name:             "none",
			border:           Border{Style: BorderNone, Width: 1},
			expectHorizontal: 0,
			expectVertical:   0,
		},
		{
			name:             "single width 1",
			border:           Border{Style: BorderSingle, Width: 1},
			expectHorizontal: 2,
			expectVertical:   2,
		},
		{
			name:             "double width 2",
			border:           Border{Style: BorderDouble, Width: 2},
			expectHorizontal: 4,
			expectVertical:   4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectHorizontal, tt.border.HorizontalPadding())
			assert.Equal(t, tt.expectVertical, tt.border.VerticalPadding())
		})
	}
}

func TestBorder_ContentOffset(t *testing.T) {
	// No border
	none := Border{Style: BorderNone}
	x, y := none.ContentOffset()
	assert.Equal(t, 0, x)
	assert.Equal(t, 0, y)

	// With border
	single := Border{Style: BorderSingle, Width: 1}
	x, y = single.ContentOffset()
	assert.Equal(t, 1, x)
	assert.Equal(t, 1, y)

	double := Border{Style: BorderDouble, Width: 2}
	x, y = double.ContentOffset()
	assert.Equal(t, 2, x)
	assert.Equal(t, 2, y)
}

// =============================================================================
// BorderedNode Tests
// =============================================================================

func TestNewBorderedNode(t *testing.T) {
	child := NewMockNode("child", 50, 30)
	border := NewBorder(BorderSingle)
	node := NewBorderedNode("bordered", child, border)

	assert.Equal(t, "bordered", node.ID())
	assert.Equal(t, "bordered", node.Type()) // Wait, Type() returns "bordered"
	assert.Equal(t, border, node.GetBorder())
	assert.Equal(t, child, node.GetChild())
}

func TestBorderedNode_Children(t *testing.T) {
	child := NewMockNode("child", 50, 30)
	border := NewBorder(BorderSingle)
	node := NewBorderedNode("bordered", child, border)

	children := node.Children()
	assert.Len(t, children, 1)
	assert.Equal(t, child, children[0])

	// Nil child
	nilNode := NewBorderedNode("nil", nil, border)
	assert.Nil(t, nilNode.Children())
}

func TestBorderedNode_GetSize(t *testing.T) {
	child := NewMockNode("child", 50, 30)

	// No border
	noneNode := NewBorderedNode("none", child, Border{Style: BorderNone})
	w, h := noneNode.GetSize()
	assert.Equal(t, 50, w)
	assert.Equal(t, 30, h)

	// Single border (adds 2 to each dimension)
	singleNode := NewBorderedNode("single", child, NewBorder(BorderSingle))
	w, h = singleNode.GetSize()
	assert.Equal(t, 52, w) // 50 + 2
	assert.Equal(t, 32, h) // 30 + 2

	// Double border (adds 4 to each dimension)
	doubleNode := NewBorderedNode("double", child, NewBorder(BorderDouble))
	w, h = doubleNode.GetSize()
	assert.Equal(t, 54, w) // 50 + 4
	assert.Equal(t, 34, h) // 30 + 4
}

func TestBorderedNode_MeasureInner(t *testing.T) {
	child := NewMockNode("child", 50, 30)
	border := NewBorder(BorderSingle)
	node := NewBorderedNode("bordered", child, border)

	// Constraints 100x50, border takes 2x2, inner should be 98x48
	constraints := Constraints{MinWidth: 100, MaxWidth: 100, MinHeight: 50, MaxHeight: 50}
	inner := node.MeasureInner(constraints)

	assert.Equal(t, 98, inner.MaxWidth)
	assert.Equal(t, 48, inner.MaxHeight)
}

func TestBorderedNode_MeasureOuter(t *testing.T) {
	child := NewMockNode("child", 50, 30)
	border := NewBorder(BorderSingle)
	node := NewBorderedNode("bordered", child, border)

	// Inner 50x30, border adds 2x2, outer should be 52x32
	w, h := node.MeasureOuter(50, 30)
	assert.Equal(t, 52, w)
	assert.Equal(t, 32, h)

	// No border
	noneNode := NewBorderedNode("none", child, Border{Style: BorderNone})
	w, h = noneNode.MeasureOuter(50, 30)
	assert.Equal(t, 50, w)
	assert.Equal(t, 30, h)
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestIsBordered(t *testing.T) {
	border := NewBorder(BorderSingle)
	borderedNode := NewBorderedNode("test", NewMockNode("child", 10, 10), border)
	regularNode := NewMockNode("regular", 10, 10)

	assert.True(t, isBordered(borderedNode))
	assert.False(t, isBordered(regularNode))
}

func TestGetBorderFromNode(t *testing.T) {
	border := NewBorder(BorderSingle)
	borderedNode := NewBorderedNode("test", NewMockNode("child", 10, 10), border)
	regularNode := NewMockNode("regular", 10, 10)

	// Bordered node returns its border
	result := GetBorderFromNode(borderedNode)
	assert.Equal(t, BorderSingle, result.Style)

	// Regular node returns no border
	result = GetBorderFromNode(regularNode)
	assert.Equal(t, BorderNone, result.Style)

	// Nil node returns no border
	result = GetBorderFromNode(nil)
	assert.Equal(t, BorderNone, result.Style)
}

func TestHasBorder(t *testing.T) {
	border := NewBorder(BorderSingle)
	borderedNode := NewBorderedNode("test", NewMockNode("child", 10, 10), border)
	regularNode := NewMockNode("regular", 10, 10)

	assert.True(t, HasBorder(borderedNode))
	assert.False(t, HasBorder(regularNode))
}

func TestCalculateBorderConstraints(t *testing.T) {
	constraints := Constraints{MinWidth: 100, MaxWidth: 100, MinHeight: 50, MaxHeight: 50}

	// No border
	result := CalculateBorderConstraints(constraints, Border{Style: BorderNone})
	assert.Equal(t, 100, result.MaxWidth)
	assert.Equal(t, 50, result.MaxHeight)

	// Single border (width 1)
	result = CalculateBorderConstraints(constraints, NewBorder(BorderSingle))
	assert.Equal(t, 98, result.MaxWidth)
	assert.Equal(t, 48, result.MaxHeight)

	// Double border (width 2)
	result = CalculateBorderConstraints(constraints, NewBorder(BorderDouble))
	assert.Equal(t, 96, result.MaxWidth)
	assert.Equal(t, 46, result.MaxHeight)
}

func TestCalculateBorderBoxSize(t *testing.T) {
	// No border
	w, h := CalculateBorderBoxSize(50, 30, Border{Style: BorderNone})
	assert.Equal(t, 50, w)
	assert.Equal(t, 30, h)

	// Single border
	w, h = CalculateBorderBoxSize(50, 30, NewBorder(BorderSingle))
	assert.Equal(t, 52, w)
	assert.Equal(t, 32, h)

	// Double border
	w, h = CalculateBorderBoxSize(50, 30, NewBorder(BorderDouble))
	assert.Equal(t, 54, w)
	assert.Equal(t, 34, h)
}

// =============================================================================
// Bordered Interface Tests
// =============================================================================

func TestBordered_Interface(t *testing.T) {
	border := NewBorderWithLabel(BorderSingle, "Test")
	node := NewBorderedNode("test", NewMockNode("child", 10, 10), border)

	// Test interface assertion
	var _ Bordered = node

	// Test GetBorder
	result := node.GetBorder()
	assert.Equal(t, BorderSingle, result.Style)
	assert.Equal(t, "Test", result.Label)
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestBorder_ZeroWidth(t *testing.T) {
	border := Border{Style: BorderSingle, Width: 0}
	// Width 0 with style should still work
	assert.True(t, border.HasBorder())
	assert.Equal(t, 0, border.HorizontalPadding())
	assert.Equal(t, 0, border.VerticalPadding())
}

func TestBorder_NegativeWidth(t *testing.T) {
	border := Border{Style: BorderSingle, Width: -1}
	// Implementation doesn't prevent negative, but padding calc should work
	assert.True(t, border.HasBorder())
	assert.Equal(t, -2, border.HorizontalPadding())
}

func TestBorderedNode_NilChild(t *testing.T) {
	border := NewBorder(BorderSingle)
	node := NewBorderedNode("test", nil, border)

	w, h := node.GetSize()
	assert.Equal(t, 0, w)
	assert.Equal(t, 0, h)
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkBorder_Padding(b *testing.B) {
	border := NewBorder(BorderSingle)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = border.HorizontalPadding()
		_ = border.VerticalPadding()
	}
}

func BenchmarkBorderedNode_MeasureInner(b *testing.B) {
	child := NewMockNode("child", 50, 30)
	border := NewBorder(BorderSingle)
	node := NewBorderedNode("test", child, border)
	constraints := Constraints{MinWidth: 100, MaxWidth: 100, MinHeight: 50, MaxHeight: 50}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = node.MeasureInner(constraints)
	}
}

func BenchmarkCalculateBorderConstraints(b *testing.B) {
	constraints := Constraints{MinWidth: 100, MaxWidth: 100, MinHeight: 50, MaxHeight: 50}
	border := NewBorder(BorderSingle)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateBorderConstraints(constraints, border)
	}
}

func BenchmarkGetBorderFromNode(b *testing.B) {
	border := NewBorder(BorderSingle)
	node := NewBorderedNode("test", NewMockNode("child", 10, 10), border)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetBorderFromNode(node)
	}
}
