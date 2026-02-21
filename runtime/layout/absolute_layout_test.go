package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Mock AbsoluteStyleProvider Node
// =============================================================================

// MockAbsoluteNode is a mock node that implements AbsoluteStyleProvider
type MockAbsoluteNode struct {
	*MockNode
	absStyle *AbsoluteStyle
	border   Border
}

// NewMockAbsoluteNode creates a mock absolute positioned container
func NewMockAbsoluteNode(id string, containerWidth, containerHeight int) *MockAbsoluteNode {
	return &MockAbsoluteNode{
		MockNode: NewMockNode(id, containerWidth, containerHeight),
		absStyle: NewAbsoluteStyle(),
		border:   NewBorder(BorderNone),
	}
}

// GetAbsoluteStyle implements AbsoluteStyleProvider
func (m *MockAbsoluteNode) GetAbsoluteStyle() *AbsoluteStyle {
	return m.absStyle
}

// SetChildren sets the children of this node
func (m *MockAbsoluteNode) SetChildren(children []Node) {
	m.children = children
}

// SetPositionStyle sets the absolute positioning style
func (m *MockAbsoluteNode) SetPositionStyle(left, top, right, bottom PositionValue, anchor Anchor) {
	m.absStyle.Left = left
	m.absStyle.Top = top
	m.absStyle.Right = right
	m.absStyle.Bottom = bottom
	m.absStyle.Anchor = anchor
}

// GetBorder implements Bordered interface
func (m *MockAbsoluteNode) GetBorder() Border {
	return m.border
}

// SetBorder sets the border style
func (m *MockAbsoluteNode) SetBorder(style BorderStyle) {
	m.border = NewBorder(style)
}

// =============================================================================
// Engine.Layout with AbsoluteStyleProvider Tests
// =============================================================================

func TestEngine_AbsoluteLayout_TopLeft(t *testing.T) {
	// Container: 60x10
	// Child: 10x1 positioned at (0,0)
	container := NewMockAbsoluteNode("container", 60, 10)
	child := NewMockMeasurableNode("child", 10, 1)
	container.SetChildren([]Node{child})
	container.SetPositionStyle(AbsolutePos(0), AbsolutePos(0), nil, nil, AnchorTopLeft)

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.NotNil(t, result.Root)
	assert.Equal(t, 60, result.Root.Width)
	assert.Equal(t, 10, result.Root.Height)
	assert.Len(t, result.Root.Children, 1)

	// Child should be at (0,0)
	childBox := result.Root.Children[0]
	assert.Equal(t, 0, childBox.X)
	assert.Equal(t, 0, childBox.Y)
}

func TestEngine_AbsoluteLayout_TopRight(t *testing.T) {
	// Container: 60x10
	// Child: 10x1 positioned at top-right
	container := NewMockAbsoluteNode("container", 60, 10)
	child := NewMockMeasurableNode("child", 10, 1)
	container.SetChildren([]Node{child})
	container.SetPositionStyle(nil, AbsolutePos(0), AbsolutePos(0), nil, AnchorTopLeft)

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)

	// Child should be at (50,0) = 60 - 10 - 0
	childBox := result.Root.Children[0]
	assert.Equal(t, 50, childBox.X, "Child should be at right edge minus child width")
	assert.Equal(t, 0, childBox.Y)
}

func TestEngine_AbsoluteLayout_BottomLeft(t *testing.T) {
	// Container: 60x10
	// Child: 10x1 positioned at bottom-left
	container := NewMockAbsoluteNode("container", 60, 10)
	child := NewMockMeasurableNode("child", 10, 1)
	container.SetChildren([]Node{child})
	container.SetPositionStyle(AbsolutePos(0), nil, nil, AbsolutePos(0), AnchorTopLeft)

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)

	// Child should be at (0,9) = 10 - 1 - 0
	childBox := result.Root.Children[0]
	assert.Equal(t, 0, childBox.X)
	assert.Equal(t, 9, childBox.Y, "Child should be at bottom edge minus child height")
}

func TestEngine_AbsoluteLayout_BottomRight(t *testing.T) {
	// Container: 60x10
	// Child: 10x1 positioned at bottom-right
	container := NewMockAbsoluteNode("container", 60, 10)
	child := NewMockMeasurableNode("child", 10, 1)
	container.SetChildren([]Node{child})
	container.SetPositionStyle(nil, nil, AbsolutePos(0), AbsolutePos(0), AnchorTopLeft)

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)

	// Child should be at (50,9) = (60-10-0, 10-1-0)
	childBox := result.Root.Children[0]
	assert.Equal(t, 50, childBox.X)
	assert.Equal(t, 9, childBox.Y)
}

func TestEngine_AbsoluteLayout_Center(t *testing.T) {
	// Container: 60x10
	// Child: 10x1 positioned at center using anchor
	container := NewMockAbsoluteNode("container", 60, 10)
	child := NewMockMeasurableNode("child", 10, 1)
	container.SetChildren([]Node{child})
	// Position at 50%,50% with Center anchor
	container.SetPositionStyle(RelativePos(50), RelativePos(50), nil, nil, AnchorCenter)

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)

	// Child should be centered: (30-5, 5-0) = (25, 5)
	// Note: 50% of 10 = 5, 5 - 1/2 = 5 - 0 = 5 (integer division for child height)
	childBox := result.Root.Children[0]
	assert.Equal(t, 25, childBox.X, "Child X should be centered (50% of 60 minus half width)")
	assert.Equal(t, 5, childBox.Y, "Child Y should be centered (50% of 10 minus half height, with integer division)")
}

func TestEngine_AbsoluteLayout_Percentage(t *testing.T) {
	// Container: 100x100
	// Child: 20x10 positioned at 25%,25%
	container := NewMockAbsoluteNode("container", 100, 100)
	child := NewMockMeasurableNode("child", 20, 10)
	container.SetChildren([]Node{child})
	container.SetPositionStyle(RelativePos(25), RelativePos(25), nil, nil, AnchorTopLeft)

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)

	// Child should be at (25,25) = 25% of 100
	childBox := result.Root.Children[0]
	assert.Equal(t, 25, childBox.X)
	assert.Equal(t, 25, childBox.Y)
}

func TestEngine_AbsoluteLayout_MultipleChildren(t *testing.T) {
	// Container with multiple absolute positioned children
	container := NewMockAbsoluteNode("container", 100, 100)

	child1 := NewMockMeasurableNode("child1", 10, 1)
	child2 := NewMockMeasurableNode("child2", 10, 1)
	child3 := NewMockMeasurableNode("child3", 10, 1)

	// Child 1: top-left
	// Child 2: center
	// Child 3: bottom-right
	container.SetChildren([]Node{child1, child2, child3})

	// Note: All children share the same container's absStyle
	// This test verifies the layout engine can handle multiple children
	container.SetPositionStyle(AbsolutePos(0), AbsolutePos(0), nil, nil, AnchorTopLeft)

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 3)
}

func TestEngine_AbsoluteLayout_WithBorder(t *testing.T) {
	// Test that absolute layout respects border offset
	container := NewMockAbsoluteNode("container", 60, 10)
	containerWithBorder := NewMockBorderedNode(container, BorderSingle, "container")
	child := NewMockMeasurableNode("child", 10, 1)
	container.SetChildren([]Node{child})
	container.SetPositionStyle(AbsolutePos(0), AbsolutePos(0), nil, nil, AnchorTopLeft)

	engine := NewEngine()
	result := engine.Layout(containerWithBorder, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)

	// With border, child should be offset by border offset (1,1)
	childBox := result.Root.Children[0]
	// The absolute layout should add borderOffset to the calculated position
	assert.Equal(t, 1, childBox.X, "Child X should include border offset")
	assert.Equal(t, 1, childBox.Y, "Child Y should include border offset")
}

// MockBorderedNode wraps a node with border support
type MockBorderedNode struct {
	Node
	border Border
	id     string
}

func NewMockBorderedNode(inner Node, style BorderStyle, id string) *MockBorderedNode {
	return &MockBorderedNode{
		Node:   inner,
		border: NewBorder(style),
		id:     id,
	}
}

func (m *MockBorderedNode) GetBorder() Border {
	return m.border
}

func (m *MockBorderedNode) ID() string {
	return m.id
}

// =============================================================================
// AbsoluteStyle.CalculatePosition Tests (Comprehensive)
// =============================================================================

func TestAbsoluteStyle_CalculatePosition_AllCorners(t *testing.T) {
	containerWidth := 100
	containerHeight := 50
	childWidth := 20
	childHeight := 10

	tests := []struct {
		name      string
		left      PositionValue
		top       PositionValue
		right     PositionValue
		bottom    PositionValue
		anchor    Anchor
		expectedX int
		expectedY int
	}{
		{
			name:      "TopLeft at origin",
			left:      AbsolutePos(0),
			top:       AbsolutePos(0),
			anchor:    AnchorTopLeft,
			expectedX: 0,
			expectedY: 0,
		},
		{
			name:      "TopLeft with offset",
			left:      AbsolutePos(10),
			top:       AbsolutePos(5),
			anchor:    AnchorTopLeft,
			expectedX: 10,
			expectedY: 5,
		},
		{
			name:      "TopRight with right=0",
			right:     AbsolutePos(0),
			top:       AbsolutePos(0),
			anchor:    AnchorTopLeft,
			expectedX: 80, // 100 - 20 - 0
			expectedY: 0,
		},
		{
			name:      "BottomLeft with bottom=0",
			left:      AbsolutePos(0),
			bottom:    AbsolutePos(0),
			anchor:    AnchorTopLeft,
			expectedX: 0,
			expectedY: 40, // 50 - 10 - 0
		},
		{
			name:      "BottomRight corner",
			right:     AbsolutePos(0),
			bottom:    AbsolutePos(0),
			anchor:    AnchorTopLeft,
			expectedX: 80,
			expectedY: 40,
		},
		{
			name:      "Center anchor",
			left:      RelativePos(50),
			top:       RelativePos(50),
			anchor:    AnchorCenter,
			expectedX: 40, // 50 - 20/2 = 40
			expectedY: 20, // 25 - 10/2 = 20
		},
		{
			name:      "Percentage position 25%",
			left:      RelativePos(25),
			top:       RelativePos(25),
			anchor:    AnchorTopLeft,
			expectedX: 25,
			expectedY: 12, // 50 * 25 / 100 = 12.5 -> 12
		},
		{
			name:      "Percentage position 75%",
			left:      RelativePos(75),
			top:       RelativePos(75),
			anchor:    AnchorTopLeft,
			expectedX: 75,
			expectedY: 37, // 50 * 75 / 100 = 37.5 -> 37
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := NewAbsoluteStyle()
			style.Left = tt.left
			style.Top = tt.top
			style.Right = tt.right
			style.Bottom = tt.bottom
			style.Anchor = tt.anchor

			x, y := style.CalculatePosition(containerWidth, containerHeight, childWidth, childHeight)
			assert.Equal(t, tt.expectedX, x, "X position mismatch")
			assert.Equal(t, tt.expectedY, y, "Y position mismatch")
		})
	}
}

func TestAbsoluteStyle_AllAnchors(t *testing.T) {
	containerWidth := 100
	containerHeight := 100
	childWidth := 20
	childHeight := 10

	// Position at (50, 50) and test all anchors
	tests := []struct {
		name      string
		anchor    Anchor
		expectedX int
		expectedY int
	}{
		{"AnchorTopLeft", AnchorTopLeft, 50, 50},
		{"AnchorTop", AnchorTop, 40, 50},     // 50 - 20/2 = 40
		{"AnchorTopRight", AnchorTopRight, 30, 50}, // 50 - 20 = 30
		{"AnchorLeft", AnchorLeft, 50, 45},   // 50 - 10/2 = 45
		{"AnchorCenter", AnchorCenter, 40, 45}, // 50-10, 50-5
		{"AnchorRight", AnchorRight, 30, 45}, // 50-20, 50-5
		{"AnchorBottomLeft", AnchorBottomLeft, 50, 40}, // 50 - 10 = 40
		{"AnchorBottom", AnchorBottom, 40, 40}, // 50-10, 50-10
		{"AnchorBottomRight", AnchorBottomRight, 30, 40}, // 50-20, 50-10
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := NewAbsoluteStyle()
			style.Left = AbsolutePos(50)
			style.Top = AbsolutePos(50)
			style.Anchor = tt.anchor

			x, y := style.CalculatePosition(containerWidth, containerHeight, childWidth, childHeight)
			assert.Equal(t, tt.expectedX, x, "X position mismatch for %s", tt.name)
			assert.Equal(t, tt.expectedY, y, "Y position mismatch for %s", tt.name)
		})
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestEngine_AbsoluteLayout_NestedInVStack(t *testing.T) {
	// VStack containing multiple absolute positioned containers
	vstack := NewMockFlexNode("vstack", FlexColumn)
	vstack.SetSize(60, 30)

	// First absolute container
	abs1 := NewMockAbsoluteNode("abs1", 60, 10)
	child1 := NewMockMeasurableNode("child1", 10, 1)
	abs1.SetChildren([]Node{child1})
	abs1.SetPositionStyle(AbsolutePos(0), AbsolutePos(0), nil, nil, AnchorTopLeft)

	// Second absolute container
	abs2 := NewMockAbsoluteNode("abs2", 60, 10)
	child2 := NewMockMeasurableNode("child2", 10, 1)
	abs2.SetChildren([]Node{child2})
	abs2.SetPositionStyle(nil, nil, AbsolutePos(0), AbsolutePos(0), AnchorTopLeft)

	vstack.SetChildren([]Node{abs1, abs2})

	engine := NewEngine()
	result := engine.Layout(vstack, NewConstraints(60, 60, 0, 100))

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 2)

	// VStack children should be arranged vertically
	assert.Equal(t, 0, result.Root.Children[0].Y)
	assert.Greater(t, result.Root.Children[1].Y, result.Root.Children[0].Y)
}

// MockFlexNode for flex layout testing
type MockFlexNode struct {
	*MockNode
	flexStyle *FlexStyle
}

func NewMockFlexNode(id string, direction FlexDirection) *MockFlexNode {
	return &MockFlexNode{
		MockNode: NewMockNode(id, 0, 0),
		flexStyle: &FlexStyle{
			Direction: direction,
			MainAxis:  MainStart,
			CrossAxis: CrossStart,
			Gap:       0,
		},
	}
}

func (m *MockFlexNode) GetFlexStyle() *FlexStyle {
	return m.flexStyle
}

func (m *MockFlexNode) SetChildren(children []Node) {
	m.children = children
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestEngine_AbsoluteLayout_EmptyContainer(t *testing.T) {
	container := NewMockAbsoluteNode("container", 60, 10)
	container.SetChildren(nil)

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 0)
}

func TestEngine_AbsoluteLayout_NoStyle(t *testing.T) {
	// Container without AbsoluteStyle should use default layout
	container := NewMockNode("container", 60, 10)
	child := NewMockMeasurableNode("child", 10, 1)
	container.children = []Node{child}

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	// Default layout should stack children vertically
	assert.Len(t, result.Root.Children, 1)
}

func TestEngine_AbsoluteLayout_ZeroSizeChild(t *testing.T) {
	container := NewMockAbsoluteNode("container", 60, 10)
	child := NewMockMeasurableNode("child", 0, 0)
	container.SetChildren([]Node{child})
	container.SetPositionStyle(AbsolutePos(10), AbsolutePos(5), nil, nil, AnchorTopLeft)

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)
	// Child should still be positioned
	assert.Equal(t, 10, result.Root.Children[0].X)
	assert.Equal(t, 5, result.Root.Children[0].Y)
}

func TestEngine_AbsoluteLayout_ChildLargerThanContainer(t *testing.T) {
	// Child larger than container
	container := NewMockAbsoluteNode("container", 20, 10)
	child := NewMockMeasurableNode("child", 30, 15)
	container.SetChildren([]Node{child})
	container.SetPositionStyle(AbsolutePos(0), AbsolutePos(0), nil, nil, AnchorTopLeft)

	engine := NewEngine()
	result := engine.Layout(container, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)
	// Child should still be positioned at origin
	assert.Equal(t, 0, result.Root.Children[0].X)
	assert.Equal(t, 0, result.Root.Children[0].Y)
}
