package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// PositionType Tests
// =============================================================================

func TestPositionType_String(t *testing.T) {
	tests := []struct {
		pt       PositionType
		expected string
	}{
		{PositionRelative, "relative"},
		{PositionAbsolute, "absolute"},
		{PositionFixed, "fixed"},
		{PositionType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.pt.String())
		})
	}
}

// =============================================================================
// Position Tests
// =============================================================================

func TestPosition_IsAbsolute(t *testing.T) {
	assert.True(t, Position{Type: PositionAbsolute}.IsAbsolute())
	assert.False(t, Position{Type: PositionRelative}.IsAbsolute())
}

func TestPosition_IsRelative(t *testing.T) {
	assert.True(t, Position{Type: PositionRelative}.IsRelative())
	assert.False(t, Position{Type: PositionAbsolute}.IsRelative())
}

func TestPosition_HasOffsets(t *testing.T) {
	top := 10
	left := 20
	right := 30
	bottom := 40

	pos := Position{
		Type:   PositionAbsolute,
		Top:    &top,
		Left:   &left,
		Right:  &right,
		Bottom: &bottom,
	}

	assert.True(t, pos.HasTop())
	assert.True(t, pos.HasLeft())
	assert.True(t, pos.HasRight())
	assert.True(t, pos.HasBottom())

	emptyPos := Position{Type: PositionAbsolute}
	assert.False(t, emptyPos.HasTop())
	assert.False(t, emptyPos.HasLeft())
	assert.False(t, emptyPos.HasRight())
	assert.False(t, emptyPos.HasBottom())
}

func TestNewRelativePosition(t *testing.T) {
	pos := NewRelativePosition()
	assert.Equal(t, PositionRelative, pos.Type)
	assert.Nil(t, pos.Top)
	assert.Nil(t, pos.Left)
	assert.Nil(t, pos.Right)
	assert.Nil(t, pos.Bottom)
}

func TestNewAbsolutePosition(t *testing.T) {
	pos := NewAbsolutePosition()
	assert.Equal(t, PositionAbsolute, pos.Type)
	assert.Nil(t, pos.Top)
	assert.Nil(t, pos.Left)
	assert.Nil(t, pos.Right)
	assert.Nil(t, pos.Bottom)
}

func TestNewAbsolutePositionWithOffsets(t *testing.T) {
	top, left, right, bottom := 10, 20, 30, 40
	pos := NewAbsolutePositionWithOffsets(&top, &left, &right, &bottom)

	assert.Equal(t, PositionAbsolute, pos.Type)
	assert.Equal(t, 10, *pos.Top)
	assert.Equal(t, 20, *pos.Left)
	assert.Equal(t, 30, *pos.Right)
	assert.Equal(t, 40, *pos.Bottom)
}

// =============================================================================
// CalculateAbsolutePosition Tests
// =============================================================================

func TestCalculateAbsolutePosition_TopLeft(t *testing.T) {
	top, left := 10, 20
	pos := NewAbsolutePositionWithOffsets(&top, &left, nil, nil)

	// Parent: 100x50, Node: 30x20
	x, y := CalculateAbsolutePosition(100, 50, 30, 20, pos)
	assert.Equal(t, 20, x) // left offset
	assert.Equal(t, 10, y) // top offset
}

func TestCalculateAbsolutePosition_BottomRight(t *testing.T) {
	right, bottom := 15, 25
	pos := NewAbsolutePositionWithOffsets(nil, nil, &right, &bottom)

	// Parent: 100x50, Node: 30x20
	// X = 100 - 15 - 30 = 55
	// Y = 50 - 25 - 20 = 5
	x, y := CalculateAbsolutePosition(100, 50, 30, 20, pos)
	assert.Equal(t, 55, x)
	assert.Equal(t, 5, y)
}

func TestCalculateAbsolutePosition_NoOffsets(t *testing.T) {
	pos := NewAbsolutePosition()

	// No offsets means 0,0
	x, y := CalculateAbsolutePosition(100, 50, 30, 20, pos)
	assert.Equal(t, 0, x)
	assert.Equal(t, 0, y)
}

func TestCalculateAbsolutePosition_MixedOffsets(t *testing.T) {
	top, right := 10, 15
	pos := NewAbsolutePositionWithOffsets(&top, nil, &right, nil)

	// Parent: 100x50, Node: 30x20
	// X = 100 - 15 - 30 = 55 (right takes precedence over left)
	// Y = 10 (top)
	x, y := CalculateAbsolutePosition(100, 50, 30, 20, pos)
	assert.Equal(t, 55, x)
	assert.Equal(t, 10, y)
}

// =============================================================================
// PositionedLayoutBox Tests
// =============================================================================

func TestPositionedLayoutBox_IsAbsolute(t *testing.T) {
	relativeBox := &PositionedLayoutBox{
		LayoutBox:    &LayoutBox{ID: "rel"},
		PositionType: PositionRelative,
	}
	assert.False(t, relativeBox.IsAbsolute())

	absoluteBox := &PositionedLayoutBox{
		LayoutBox:    &LayoutBox{ID: "abs"},
		PositionType: PositionAbsolute,
	}
	assert.True(t, absoluteBox.IsAbsolute())
}

func TestPositionedLayoutBox_GetEffectivePosition(t *testing.T) {
	// Relative box uses LayoutBox X,Y
	relativeBox := &PositionedLayoutBox{
		LayoutBox:    &LayoutBox{ID: "rel", X: 10, Y: 20},
		PositionType: PositionRelative,
	}
	x, y := relativeBox.GetEffectivePosition()
	assert.Equal(t, 10, x)
	assert.Equal(t, 20, y)

	// Absolute box uses AbsoluteX,AbsoluteY
	absoluteBox := &PositionedLayoutBox{
		LayoutBox:    &LayoutBox{ID: "abs", X: 10, Y: 20},
		PositionType: PositionAbsolute,
		AbsoluteX:    50,
		AbsoluteY:    60,
	}
	x, y = absoluteBox.GetEffectivePosition()
	assert.Equal(t, 50, x)
	assert.Equal(t, 60, y)
}

// =============================================================================
// Positionable Interface Tests
// =============================================================================

// mockPositionableNode implements Positionable interface
type mockPositionableNode struct {
	*MockNode
	position Position
}

func newMockPositionableNode(id string, width, height int, position Position) *mockPositionableNode {
	return &mockPositionableNode{
		MockNode: NewMockNode(id, width, height),
		position: position,
	}
}

func (n *mockPositionableNode) GetPositionType() Position {
	return n.position
}

func TestPositionable_Interface(t *testing.T) {
	top, left := 10, 20
	pos := NewAbsolutePositionWithOffsets(&top, &left, nil, nil)
	node := newMockPositionableNode("test", 100, 50, pos)

	// Test interface assertion
	var _ Positionable = node

	// Test GetPositionType
	result := node.GetPositionType()
	assert.True(t, result.IsAbsolute())
	assert.True(t, result.HasTop())
	assert.True(t, result.HasLeft())
}

func TestIsPositionAbsolute(t *testing.T) {
	absoluteNode := newMockPositionableNode("abs", 100, 50, NewAbsolutePosition())
	relativeNode := newMockPositionableNode("rel", 100, 50, NewRelativePosition())
	regularNode := NewMockNode("regular", 100, 50) // Does not implement Positionable

	assert.True(t, IsPositionAbsolute(absoluteNode))
	assert.False(t, IsPositionAbsolute(relativeNode))
	assert.False(t, IsPositionAbsolute(regularNode))
}

func TestGetPositionFromNode(t *testing.T) {
	pos := NewAbsolutePosition()
	positionableNode := newMockPositionableNode("test", 100, 50, pos)
	regularNode := NewMockNode("regular", 100, 50)

	// Positionable node returns its position
	result := GetPositionFromNode(positionableNode)
	assert.Equal(t, PositionAbsolute, result.Type)

	// Regular node returns relative position
	result = GetPositionFromNode(regularNode)
	assert.Equal(t, PositionRelative, result.Type)

	// Nil node returns relative position
	result = GetPositionFromNode(nil)
	assert.Equal(t, PositionRelative, result.Type)
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestPosition_ZeroOffsets(t *testing.T) {
	zero := 0
	pos := NewAbsolutePositionWithOffsets(&zero, &zero, nil, nil)

	assert.True(t, pos.HasTop())
	assert.True(t, pos.HasLeft())
	assert.Equal(t, 0, *pos.Top)
	assert.Equal(t, 0, *pos.Left)
}

func TestPosition_NegativeOffsets(t *testing.T) {
	neg := -10
	pos := NewAbsolutePositionWithOffsets(&neg, nil, nil, nil)

	// Negative offsets are allowed (pull element in opposite direction)
	assert.True(t, pos.HasTop())
	assert.Equal(t, -10, *pos.Top)
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkCalculateAbsolutePosition(b *testing.B) {
	top, left := 10, 20
	pos := NewAbsolutePositionWithOffsets(&top, &left, nil, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateAbsolutePosition(100, 50, 30, 20, pos)
	}
}

func BenchmarkIsPositionAbsolute(b *testing.B) {
	node := newMockPositionableNode("test", 100, 50, NewAbsolutePosition())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsPositionAbsolute(node)
	}
}

func BenchmarkGetPositionFromNode(b *testing.B) {
	node := newMockPositionableNode("test", 100, 50, NewAbsolutePosition())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetPositionFromNode(node)
	}
}
