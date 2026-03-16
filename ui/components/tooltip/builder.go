package tooltip

import (
	"time"

	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Tooltip Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for creating Tooltip VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Tooltip builder.
func NewBuilder(content rtui.VNode, text string) *Builder {
	return &Builder{
		node: New(content, text),
	}
}

// Key sets the key for diffing.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// SetID sets the business identifier for positioning and Portal anchoring.
// This is separate from Key() which is used for list diffing.
func (b *Builder) SetID(id string) *Builder {
	b.node.SetID(id)
	return b
}

// Text sets the tooltip text.
func (b *Builder) Text(text string) *Builder {
	b.node.SetText(text)
	return b
}

// Position sets the tooltip position.
func (b *Builder) Position(position Position) *Builder {
	b.node.SetPosition(position)
	return b
}

// Top sets position to top.
func (b *Builder) Top() *Builder {
	return b.Position(PositionTop)
}

// TopLeft sets position to top-left.
func (b *Builder) TopLeft() *Builder {
	return b.Position(PositionTopLeft)
}

// TopRight sets position to top-right.
func (b *Builder) TopRight() *Builder {
	return b.Position(PositionTopRight)
}

// Bottom sets position to bottom.
func (b *Builder) Bottom() *Builder {
	return b.Position(PositionBottom)
}

// BottomLeft sets position to bottom-left.
func (b *Builder) BottomLeft() *Builder {
	return b.Position(PositionBottomLeft)
}

// BottomRight sets position to bottom-right.
func (b *Builder) BottomRight() *Builder {
	return b.Position(PositionBottomRight)
}

// Left sets position to left.
func (b *Builder) Left() *Builder {
	return b.Position(PositionLeft)
}

// LeftTop sets position to left-top.
func (b *Builder) LeftTop() *Builder {
	return b.Position(PositionLeftTop)
}

// LeftBottom sets position to left-bottom.
func (b *Builder) LeftBottom() *Builder {
	return b.Position(PositionLeftBottom)
}

// Right sets position to right.
func (b *Builder) Right() *Builder {
	return b.Position(PositionRight)
}

// RightTop sets position to right-top.
func (b *Builder) RightTop() *Builder {
	return b.Position(PositionRightTop)
}

// RightBottom sets position to right-bottom.
func (b *Builder) RightBottom() *Builder {
	return b.Position(PositionRightBottom)
}

// Auto sets position to auto.
func (b *Builder) Auto() *Builder {
	return b.Position(PositionAuto)
}

// Delay sets the delay before showing the tooltip.
func (b *Builder) Delay(delay time.Duration) *Builder {
	b.node.SetDelay(delay)
	return b
}

// Style sets the visual style.
func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyleProps(s)
	return b
}

// FgColor sets the foreground color.
func (b *Builder) FgColor(c interface{}) *Builder {
	s := b.node.Style()
	switch v := c.(type) {
	case string:
		s.FG = style.Color(v)
	case style.Color:
		s.FG = v
	}
	b.node.SetStyleProps(s)
	return b
}

// BgColor sets the background color.
func (b *Builder) BgColor(c interface{}) *Builder {
	s := b.node.Style()
	switch v := c.(type) {
	case string:
		s.BG = style.Color(v)
	case style.Color:
		s.BG = v
	}
	b.node.SetStyleProps(s)
	return b
}

// Build returns the VNode as rtui.VNode interface.
func (b *Builder) Build() rtui.VNode {
	return b.node
}

// Layer sets the rendering layer for the tooltip.
func (b *Builder) Layer(l rtui.Layer) *Builder {
	b.node.SetLayer(l)
	return b
}

// SetRenderLayer is an alias for Layer for backward compatibility.
func (b *Builder) SetRenderLayer(l rtui.Layer) *Builder {
	return b.Layer(l)
}

// BaseLayer sets the tooltip to the base layer.
func (b *Builder) BaseLayer() *Builder {
	return b.Layer(rtui.LayerBase)
}

// OverlayLayer sets the tooltip to the overlay layer.
func (b *Builder) OverlayLayer() *Builder {
	return b.Layer(rtui.LayerOverlay)
}

// ModalLayer sets the tooltip to the modal layer.
func (b *Builder) ModalLayer() *Builder {
	return b.Layer(rtui.LayerModal)
}

// TooltipLayer sets the tooltip to the tooltip layer (default).
func (b *Builder) TooltipLayer() *Builder {
	return b.Layer(rtui.LayerTooltip)
}

// InspectorLayer sets the tooltip to the inspector layer.
func (b *Builder) InspectorLayer() *Builder {
	return b.Layer(rtui.LayerInspector)
}

// BuildInstance creates and returns the Instance directly.
func (b *Builder) BuildInstance() *Instance {
	return NewInstance(b.node.Props())
}

// =============================================================================
// Convenience Functions
// =============================================================================

// T creates a new Tooltip builder.
func T(content rtui.VNode, text string) *Builder {
	return NewBuilder(content, text)
}

// Tooltip creates a new Tooltip VNode directly.
func Tooltip(content rtui.VNode, text string) *VNode {
	return New(content, text)
}

// TooltipBuilder is an alias for Builder (for backward compatibility).
type TooltipBuilder = Builder
