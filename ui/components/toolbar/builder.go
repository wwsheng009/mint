package toolbar

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for constructing Toolbar VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a Toolbar builder.
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

func (b *Builder) Title(title string) *Builder {
	b.node.SetTitle(title)
	return b
}

func (b *Builder) TitleWidth(width int) *Builder {
	b.node.SetTitleWidth(width)
	return b
}

func (b *Builder) Width(width int) *Builder {
	b.node.SetWidth(width)
	return b
}

func (b *Builder) Gap(gap int) *Builder {
	b.node.SetGap(gap)
	return b
}

func (b *Builder) Dense(dense bool) *Builder {
	b.node.SetDense(dense)
	return b
}

func (b *Builder) Separator(separator string) *Builder {
	b.node.SetSeparator(separator)
	return b
}

func (b *Builder) UseStatusBar(use bool) *Builder {
	b.node.SetUseStatusBar(use)
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetRootStyle(s)
	return b
}

func (b *Builder) LeftItems(items []Item) *Builder {
	b.node.SetLeftItems(items)
	return b
}

func (b *Builder) Left(item Item) *Builder {
	b.node.AddLeftItem(item)
	return b
}

func (b *Builder) CenterItems(items []Item) *Builder {
	b.node.SetCenterItems(items)
	return b
}

func (b *Builder) Center(item Item) *Builder {
	b.node.AddCenterItem(item)
	return b
}

func (b *Builder) RightItems(items []Item) *Builder {
	b.node.SetRightItems(items)
	return b
}

func (b *Builder) Right(item Item) *Builder {
	b.node.AddRightItem(item)
	return b
}

func (b *Builder) Action(item Item) *Builder {
	b.node.AddRightItem(item)
	return b
}

func (b *Builder) Actions(items []Item) *Builder {
	b.node.SetRightItems(items)
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

// Of creates a Toolbar from left, center, and right items.
func Of(left, center, right []Item) rtui.VNode {
	return NewBuilder().LeftItems(left).CenterItems(center).RightItems(right).Build()
}
