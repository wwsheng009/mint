package descriptions

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for creating Descriptions VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Descriptions builder.
func NewBuilder() *Builder {
	return &Builder{node: New(nil)}
}

// Key sets the diff key.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// SetID sets the business identifier.
func (b *Builder) SetID(id string) *Builder {
	b.node.SetID(id)
	return b
}

// Title sets the optional title.
func (b *Builder) Title(title string) *Builder {
	b.node.SetTitle(title)
	return b
}

// Extra sets the optional header-side node.
func (b *Builder) Extra(extra rtui.VNode) *Builder {
	b.node.SetExtra(extra)
	return b
}

// Items replaces all description items.
func (b *Builder) Items(items []Item) *Builder {
	b.node.SetItems(items)
	return b
}

// Item appends a description item.
func (b *Builder) Item(item Item) *Builder {
	b.node.AddItem(item)
	return b
}

// Column sets the target column count.
func (b *Builder) Column(column int) *Builder {
	b.node.SetColumn(column)
	return b
}

// Bordered toggles bordered mode.
func (b *Builder) Bordered(bordered bool) *Builder {
	b.node.SetBordered(bordered)
	return b
}

// Colon toggles label colon rendering.
func (b *Builder) Colon(colon bool) *Builder {
	b.node.SetColon(colon)
	return b
}

// Layout sets the item layout mode.
func (b *Builder) Layout(layout Layout) *Builder {
	b.node.SetLayout(layout)
	return b
}

// Horizontal renders items inline.
func (b *Builder) Horizontal() *Builder {
	b.node.SetLayout(LayoutHorizontal)
	return b
}

// Vertical renders labels above content.
func (b *Builder) Vertical() *Builder {
	b.node.SetLayout(LayoutVertical)
	return b
}

// Width sets a preferred width.
func (b *Builder) Width(width int) *Builder {
	b.node.SetWidth(width)
	return b
}

// Style sets the root style.
func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

// TitleStyle sets the title style.
func (b *Builder) TitleStyle(s style.Style) *Builder {
	b.node.SetTitleStyle(s)
	return b
}

// LabelStyle sets the default label style.
func (b *Builder) LabelStyle(s style.Style) *Builder {
	b.node.SetLabelStyle(s)
	return b
}

// ContentStyle sets the default content style.
func (b *Builder) ContentStyle(s style.Style) *Builder {
	b.node.SetContentStyle(s)
	return b
}

// Build returns the configured VNode.
func (b *Builder) Build() rtui.VNode {
	return b.node
}

// BuildVNode returns the concrete VNode.
func (b *Builder) BuildVNode() *VNode {
	return b.node
}

// Of creates a descriptions vnode from items.
func Of(items []Item) *VNode {
	return New(items)
}
