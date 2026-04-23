package anchor

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for creating Anchor VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Anchor builder.
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

func (b *Builder) ComponentID(id string) *Builder {
	b.node.SetComponentID(id)
	return b
}

func (b *Builder) Title(title string) *Builder {
	b.node.SetTitle(title)
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

func (b *Builder) ActiveKey(key string) *Builder {
	b.node.SetActiveKey(key)
	return b
}

func (b *Builder) InitialActiveKey(key string) *Builder {
	b.node.SetInitialActiveKey(key)
	return b
}

func (b *Builder) ViewportHeight(height int) *Builder {
	b.node.SetViewportHeight(height)
	return b
}

func (b *Builder) Width(width int) *Builder {
	b.node.SetWidth(width)
	return b
}

func (b *Builder) ShowBorder(show bool) *Builder {
	b.node.SetShowBorder(show)
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyleProps(s)
	return b
}

func (b *Builder) CurrentStyle(s style.Style) *Builder {
	b.node.SetCurrentStyle(s)
	return b
}

func (b *Builder) OnChange(changeIntent intent.Intent) *Builder {
	b.node.SetChangeIntent(changeIntent)
	return b
}

func (b *Builder) ForField(binding intent.FieldBinding) *Builder {
	b.node.SetChangeIntent(binding)
	return b
}

func (b *Builder) ForForm(binding intent.FormBinding) *Builder {
	b.node.SetFormID(binding.GetFormID())
	return b
}

func (b *Builder) Build() rtui.VNode {
	return b.node
}

func (b *Builder) BuildVNode() *VNode {
	return b.node
}

// Of creates an Anchor with the given items.
func Of(items []Item) rtui.VNode {
	return NewBuilder().Items(items).Build()
}
