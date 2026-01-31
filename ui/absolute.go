package ui

import (
	"github.com/wwsheng009/mint/runtime/style"
)

// Position represents a position value
type Position interface {
	isPosition()
}

// AbsolutePosition is a fixed position in cells
type AbsolutePosition int

func (a AbsolutePosition) isPosition() {}

// RelativePosition is a percentage (0-100)
type RelativePosition int

func (r RelativePosition) isPosition() {}

// Anchor represents anchor points for positioning
type Anchor int

const (
	AnchorTopLeft Anchor = iota
	AnchorTop
	AnchorTopRight
	AnchorLeft
	AnchorCenter
	AnchorRight
	AnchorBottomLeft
	AnchorBottom
	AnchorBottomRight
)

// AbsoluteVNode represents an absolutely positioned element
type AbsoluteVNode struct {
	*ElementVNode
	child     VNode
	left      Position
	top       Position
	right     Position
	bottom    Position
	anchor    Anchor
	zIndex    int
	width     int  // 0 = auto
	height    int  // 0 = auto
}

// NewAbsolute creates a new absolute positioned node
func NewAbsolute(child VNode) *AbsoluteVNode {
	return &AbsoluteVNode{
		ElementVNode: NewElement("absolute"),
		child:        child,
		left:         nil,
		top:          nil,
		right:        nil,
		bottom:       nil,
		anchor:       AnchorTopLeft,
		zIndex:       0,
		width:        0,
		height:       0,
	}
}

// Absolute creates an absolutely positioned node
func Absolute(child VNode) VNode {
	return NewAbsolute(child)
}

// Builder pattern
type AbsoluteBuilderType struct {
	node *AbsoluteVNode
}

// AbsoluteBuilder creates a new absolute builder
func AbsoluteBuilder(child VNode) *AbsoluteBuilderType {
	return &AbsoluteBuilderType{
		node: NewAbsolute(child),
	}
}

// Left sets the left position
func (b *AbsoluteBuilderType) Left(pos Position) *AbsoluteBuilderType {
	b.node.left = pos
	return b
}

// Top sets the top position
func (b *AbsoluteBuilderType) Top(pos Position) *AbsoluteBuilderType {
	b.node.top = pos
	return b
}

// Right sets the right position
func (b *AbsoluteBuilderType) Right(pos Position) *AbsoluteBuilderType {
	b.node.right = pos
	return b
}

// Bottom sets the bottom position
func (b *AbsoluteBuilderType) Bottom(pos Position) *AbsoluteBuilderType {
	b.node.bottom = pos
	return b
}

// Anchor sets the anchor point
func (b *AbsoluteBuilderType) Anchor(a Anchor) *AbsoluteBuilderType {
	b.node.anchor = a
	return b
}

// ZIndex sets the z-index (stacking order)
func (b *AbsoluteBuilderType) ZIndex(z int) *AbsoluteBuilderType {
	b.node.zIndex = z
	return b
}

// Width sets the width
func (b *AbsoluteBuilderType) Width(w int) *AbsoluteBuilderType {
	b.node.width = w
	return b
}

// Height sets the height
func (b *AbsoluteBuilderType) Height(h int) *AbsoluteBuilderType {
	b.node.height = h
	return b
}

// Style sets the visual style
func (b *AbsoluteBuilderType) Style(s style.Style) *AbsoluteBuilderType {
	b.node.SetStyle(s)
	return b
}

// FgColor sets the foreground color
func (b *AbsoluteBuilderType) FgColor(c interface{}) *AbsoluteBuilderType {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s.FG = style.Color(colorStr)
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s.FG = color
		b.node.SetStyle(s)
	}
	return b
}

// BgColor sets the background color
func (b *AbsoluteBuilderType) BgColor(c interface{}) *AbsoluteBuilderType {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s.BG = style.Color(colorStr)
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s.BG = color
		b.node.SetStyle(s)
	}
	return b
}

// Key sets the key for diffing
func (b *AbsoluteBuilderType) Key(key string) *AbsoluteBuilderType {
	b.node.SetKey(key)
	return b
}

// Build returns the absolute VNode
func (b *AbsoluteBuilderType) Build() VNode {
	return b.node
}

// Getters
func (a *AbsoluteVNode) Child() VNode        { return a.child }
func (a *AbsoluteVNode) Left() Position      { return a.left }
func (a *AbsoluteVNode) Top() Position       { return a.top }
func (a *AbsoluteVNode) Right() Position     { return a.right }
func (a *AbsoluteVNode) Bottom() Position    { return a.bottom }
func (a *AbsoluteVNode) Anchor() Anchor      { return a.anchor }
func (a *AbsoluteVNode) ZIndex() int         { return a.zIndex }
func (a *AbsoluteVNode) AbsWidth() int       { return a.width }
func (a *AbsoluteVNode) AbsHeight() int      { return a.height }

// Setters
func (a *AbsoluteVNode) SetChild(child VNode)      { a.child = child }
func (a *AbsoluteVNode) SetLeft(pos Position)      { a.left = pos }
func (a *AbsoluteVNode) SetTop(pos Position)       { a.top = pos }
func (a *AbsoluteVNode) SetRight(pos Position)     { a.right = pos }
func (a *AbsoluteVNode) SetBottom(pos Position)    { a.bottom = pos }
func (a *AbsoluteVNode) SetAnchor(anchor Anchor)   { a.anchor = anchor }
func (a *AbsoluteVNode) SetZIndex(z int)           { a.zIndex = z }
func (a *AbsoluteVNode) SetAbsWidth(w int)         { a.width = w }
func (a *AbsoluteVNode) SetAbsHeight(h int)        { a.height = h }

// CalculatePosition calculates the actual x, y position based on container size
func (a *AbsoluteVNode) CalculatePosition(containerWidth, containerHeight int) (int, int) {
	x := 0
	y := 0

	// Calculate X position
	if a.left != nil {
		switch pos := a.left.(type) {
		case AbsolutePosition:
			x = int(pos)
		case RelativePosition:
			x = containerWidth * int(pos) / 100
		}
	} else if a.right != nil {
		switch pos := a.right.(type) {
		case AbsolutePosition:
			x = containerWidth - int(pos)
		case RelativePosition:
			x = containerWidth - (containerWidth * int(pos) / 100)
		}
	}

	// Calculate Y position
	if a.top != nil {
		switch pos := a.top.(type) {
		case AbsolutePosition:
			y = int(pos)
		case RelativePosition:
			y = containerHeight * int(pos) / 100
		}
	} else if a.bottom != nil {
		switch pos := a.bottom.(type) {
		case AbsolutePosition:
			y = containerHeight - int(pos)
		case RelativePosition:
			y = containerHeight - (containerHeight * int(pos) / 100)
		}
	}

	// Adjust based on anchor
	childWidth := a.width
	if childWidth == 0 && a.child != nil {
		childWidth = 20 // Default, will be measured during render
	}

	childHeight := a.height
	if childHeight == 0 && a.child != nil {
		childHeight = 1 // Default, will be measured during render
	}

	switch a.anchor {
	case AnchorTop, AnchorTopLeft:
		// No adjustment needed for top-left anchors
	case AnchorTopRight:
		x = x - childWidth
	case AnchorLeft:
		// No adjustment needed for left anchor
	case AnchorCenter:
		x = x - childWidth/2
		y = y - childHeight/2
	case AnchorRight:
		x = x - childWidth
	case AnchorBottom:
		y = y - childHeight
	case AnchorBottomLeft:
		y = y - childHeight
	case AnchorBottomRight:
		x = x - childWidth
		y = y - childHeight
	}

	return x, y
}

// Stack creates a stacking context for absolute positioning
func Stack(children ...VNode) VNode {
	node := NewElement("stack")
	node.SetChildren(children)
	return node
}

// StackBuilder provides fluent API for building stacks
type StackBuilder struct {
	node *ElementVNode
}

// StackBuilder creates a new stack builder
func StackBuilderFunc() *StackBuilder {
	return &StackBuilder{
		node: NewElement("stack"),
	}
}

// Width sets the width
func (b *StackBuilder) Width(w int) *StackBuilder {
	b.node.SetProp("width", w)
	return b
}

// Height sets the height
func (b *StackBuilder) Height(h int) *StackBuilder {
	b.node.SetProp("height", h)
	return b
}

// Child adds a child
func (b *StackBuilder) Child(child VNode) *StackBuilder {
	b.node.SetChildren([]VNode{child})
	return b
}

// Children adds multiple children
func (b *StackBuilder) Children(children ...VNode) *StackBuilder {
	b.node.SetChildren(children)
	return b
}

// Style sets the visual style
func (b *StackBuilder) Style(s style.Style) *StackBuilder {
	b.node.SetStyle(s)
	return b
}

// FgColor sets the foreground color
func (b *StackBuilder) FgColor(c interface{}) *StackBuilder {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s.FG = style.Color(colorStr)
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s.FG = color
		b.node.SetStyle(s)
	}
	return b
}

// BgColor sets the background color
func (b *StackBuilder) BgColor(c interface{}) *StackBuilder {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s.BG = style.Color(colorStr)
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s.BG = color
		b.node.SetStyle(s)
	}
	return b
}

// Key sets the key for diffing
func (b *StackBuilder) Key(key string) *StackBuilder {
	b.node.SetKey(key)
	return b
}

// Build returns the stack VNode
func (b *StackBuilder) Build() VNode {
	return b.node
}

// =============================================================================
// Convenience functions for common positioning patterns
// =============================================================================

// TopLeft positions at top-left
func TopLeft(child VNode) VNode {
	return AbsoluteBuilder(child).
		Left(AbsolutePosition(0)).
		Top(AbsolutePosition(0)).
		Build()
}

// TopRight positions at top-right
func TopRight(child VNode) VNode {
	return AbsoluteBuilder(child).
		Right(AbsolutePosition(0)).
		Top(AbsolutePosition(0)).
		Anchor(AnchorTopRight).
		Build()
}

// BottomLeft positions at bottom-left
func BottomLeft(child VNode) VNode {
	return AbsoluteBuilder(child).
		Left(AbsolutePosition(0)).
		Bottom(AbsolutePosition(0)).
		Anchor(AnchorBottomLeft).
		Build()
}

// BottomRight positions at bottom-right
func BottomRight(child VNode) VNode {
	return AbsoluteBuilder(child).
		Right(AbsolutePosition(0)).
		Bottom(AbsolutePosition(0)).
		Anchor(AnchorBottomRight).
		Build()
}

// Center positions at center
func Center(child VNode) VNode {
	return AbsoluteBuilder(child).
		Left(RelativePosition(50)).
		Top(RelativePosition(50)).
		Anchor(AnchorCenter).
		Build()
}

// Badge creates a small badge positioned at top-right
func Badge(text string, color string) VNode {
	return TopRight(
		NewTextBuilder(text).
			FgColor(color).
			Bold(true).
			Build(),
	)
}
