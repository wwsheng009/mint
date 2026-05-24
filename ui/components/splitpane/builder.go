package splitpane

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for constructing SplitPane VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a SplitPane builder.
func NewBuilder() *Builder {
	return &Builder{node: New()}
}

func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

func (b *Builder) SetID(id string) *Builder {
	b.node.SetID(id)
	return b
}

func (b *Builder) Direction(direction Direction) *Builder {
	b.node.SetDirection(direction)
	return b
}

func (b *Builder) Horizontal() *Builder {
	b.node.SetDirection(DirectionHorizontal)
	b.node.SetSeparatorGlyph("│")
	return b
}

func (b *Builder) Vertical() *Builder {
	b.node.SetDirection(DirectionVertical)
	b.node.SetSeparatorGlyph("─")
	return b
}

func (b *Builder) Primary(primary rtui.VNode) *Builder {
	b.node.SetPrimary(primary)
	return b
}

func (b *Builder) Secondary(secondary rtui.VNode) *Builder {
	b.node.SetSecondary(secondary)
	return b
}

func (b *Builder) Panes(primary, secondary rtui.VNode) *Builder {
	b.node.SetPrimary(primary)
	b.node.SetSecondary(secondary)
	return b
}

// PrimarySize sets the fixed primary width for horizontal panes or height for vertical panes.
func (b *Builder) PrimarySize(size int) *Builder {
	b.node.SetPrimarySize(size)
	return b
}

// SecondarySize sets the fixed secondary width for horizontal panes or height for vertical panes.
func (b *Builder) SecondarySize(size int) *Builder {
	b.node.SetSecondarySize(size)
	return b
}

func (b *Builder) PrimaryFlex(flex int) *Builder {
	b.node.SetPrimaryFlex(flex)
	return b
}

func (b *Builder) SecondaryFlex(flex int) *Builder {
	b.node.SetSecondaryFlex(flex)
	return b
}

func (b *Builder) Gap(gap int) *Builder {
	b.node.SetGap(gap)
	return b
}

func (b *Builder) Separator(enabled bool) *Builder {
	b.node.SetSeparator(enabled)
	return b
}

func (b *Builder) SeparatorGlyph(glyph string) *Builder {
	b.node.SetSeparatorGlyph(glyph)
	return b
}

func (b *Builder) Width(width int) *Builder {
	b.node.SetWidth(width)
	return b
}

func (b *Builder) Height(height int) *Builder {
	b.node.SetHeight(height)
	return b
}

func (b *Builder) Size(width, height int) *Builder {
	b.node.SetWidth(width)
	b.node.SetHeight(height)
	return b
}

func (b *Builder) Align(align rtui.Align) *Builder {
	b.node.SetAlign(align)
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetRootStyle(s)
	return b
}

func (b *Builder) SeparatorStyle(s style.Style) *Builder {
	b.node.SetSeparatorStyle(s)
	return b
}

func (b *Builder) SeparatorColor(color style.Color) *Builder {
	s := b.node.separatorStyle
	s.FG = color
	b.node.SetSeparatorStyle(s)
	return b
}

func (b *Builder) Build() rtui.VNode {
	return b.node
}

func (b *Builder) BuildVNode() *VNode {
	return b.node
}

func (b *Builder) BuildInstance() *Instance {
	return NewInstance(b.node.Props())
}

// Of creates a horizontal SplitPane from two panes.
func Of(primary, secondary rtui.VNode) rtui.VNode {
	return NewBuilder().Panes(primary, secondary).Build()
}

// Horizontal creates a left/right SplitPane.
func Horizontal(primary, secondary rtui.VNode) rtui.VNode {
	return NewBuilder().Horizontal().Panes(primary, secondary).Build()
}

// Vertical creates a top/bottom SplitPane.
func Vertical(primary, secondary rtui.VNode) rtui.VNode {
	return NewBuilder().Vertical().Panes(primary, secondary).Build()
}
