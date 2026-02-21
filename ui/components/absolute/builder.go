package absolute

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for creating Absolute VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Absolute builder.
func NewBuilder(child rtui.VNode) *Builder {
	return &Builder{
		node: New(child),
	}
}

// Key sets the component key.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// Left sets the left position.
func (b *Builder) Left(pos PositionValue) *Builder {
	b.node.SetLeft(pos)
	return b
}

// Top sets the top position.
func (b *Builder) Top(pos PositionValue) *Builder {
	b.node.SetTop(pos)
	return b
}

// Right sets the right position.
func (b *Builder) Right(pos PositionValue) *Builder {
	b.node.SetRight(pos)
	return b
}

// Bottom sets the bottom position.
func (b *Builder) Bottom(pos PositionValue) *Builder {
	b.node.SetBottom(pos)
	return b
}

// Anchor sets the anchor point.
func (b *Builder) Anchor(a Anchor) *Builder {
	b.node.SetAnchor(a)
	return b
}

// ZIndex sets the z-index (stacking order).
func (b *Builder) ZIndex(z int) *Builder {
	b.node.SetZIndex(z)
	return b
}

// Width sets the width.
func (b *Builder) Width(w int) *Builder {
	b.node.SetWidth(w)
	return b
}

// Height sets the height.
func (b *Builder) Height(h int) *Builder {
	b.node.SetHeight(h)
	return b
}

// Style sets the visual style.
func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

// FgColor sets the foreground color.
func (b *Builder) FgColor(c interface{}) *Builder {
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

// BgColor sets the background color.
func (b *Builder) BgColor(c interface{}) *Builder {
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

// Flex sets the flex factor.
func (b *Builder) Flex(flex int) *Builder {
	b.node.SetFlex(flex)
	return b
}

// Position sets left and top positions.
func (b *Builder) Position(left, top PositionValue) *Builder {
	b.node.Position(left, top)
	return b
}

// Size sets width and height.
func (b *Builder) Size(width, height int) *Builder {
	b.node.Size(width, height)
	return b
}

// Build returns the Absolute VNode.
func (b *Builder) Build() rtui.VNode {
	return b.node
}

// =============================================================================
// Convenience Functions
// =============================================================================

// TopLeft positions at top-left corner.
func TopLeft(child rtui.VNode) rtui.VNode {
	return NewBuilder(child).
		Left(AbsolutePos(0)).
		Top(AbsolutePos(0)).
		Build()
}

// TopRight positions at top-right corner.
func TopRight(child rtui.VNode) rtui.VNode {
	return NewBuilder(child).
		Right(AbsolutePos(0)).
		Top(AbsolutePos(0)).
		Anchor(AnchorTopRight).
		Build()
}

// BottomLeft positions at bottom-left corner.
func BottomLeft(child rtui.VNode) rtui.VNode {
	return NewBuilder(child).
		Left(AbsolutePos(0)).
		Bottom(AbsolutePos(0)).
		Anchor(AnchorBottomLeft).
		Build()
}

// BottomRight positions at bottom-right corner.
func BottomRight(child rtui.VNode) rtui.VNode {
	return NewBuilder(child).
		Right(AbsolutePos(0)).
		Bottom(AbsolutePos(0)).
		Anchor(AnchorBottomRight).
		Build()
}

// Center positions at center.
func Center(child rtui.VNode) rtui.VNode {
	return NewBuilder(child).
		Left(RelativePos(50)).
		Top(RelativePos(50)).
		Anchor(AnchorCenter).
		Build()
}

// At positions at specific coordinates.
func At(child rtui.VNode, x, y int) rtui.VNode {
	return NewBuilder(child).
		Left(AbsolutePos(x)).
		Top(AbsolutePos(y)).
		Build()
}

// AtPercent positions at percentage coordinates.
func AtPercent(child rtui.VNode, xPercent, yPercent int) rtui.VNode {
	return NewBuilder(child).
		Left(RelativePos(xPercent)).
		Top(RelativePos(yPercent)).
		Build()
}
