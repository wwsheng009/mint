package timeline

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for creating Timeline VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Timeline builder.
func NewBuilder() *Builder {
	return &Builder{node: New(nil)}
}

func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

func (b *Builder) SetID(id string) *Builder {
	b.node.SetID(id)
	return b
}

func (b *Builder) Items(items []Item) *Builder {
	b.node.SetItems(items)
	return b
}

func (b *Builder) Item(item Item) *Builder {
	b.node.AddItem(item)
	return b
}

func (b *Builder) Pending(pending string) *Builder {
	b.node.SetPending(pending)
	return b
}

func (b *Builder) Reverse(reverse bool) *Builder {
	b.node.SetReverse(reverse)
	return b
}

func (b *Builder) Width(width int) *Builder {
	b.node.SetWidth(width)
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

func (b *Builder) LabelStyle(s style.Style) *Builder {
	b.node.SetLabelStyle(s)
	return b
}

func (b *Builder) ContentStyle(s style.Style) *Builder {
	b.node.SetContentStyle(s)
	return b
}

func (b *Builder) PendingStyle(s style.Style) *Builder {
	b.node.SetPendingStyle(s)
	return b
}

func (b *Builder) LineStyle(s style.Style) *Builder {
	b.node.SetLineStyle(s)
	return b
}

func (b *Builder) Build() rtui.VNode {
	return b.node
}

func (b *Builder) BuildVNode() *VNode {
	return b.node
}

func Of(items []Item) *VNode {
	return New(items)
}
