package filterbar

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for constructing FilterBar VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a FilterBar builder.
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

func (b *Builder) Fields(fields []Field) *Builder {
	b.node.SetFields(fields)
	return b
}

func (b *Builder) Field(field Field) *Builder {
	b.node.AddField(field)
	return b
}

func (b *Builder) Search(key, label, value string) *Builder {
	b.node.AddField(Search(key, label, value))
	return b
}

func (b *Builder) Text(key, label, value string) *Builder {
	b.node.AddField(Text(key, label, value))
	return b
}

func (b *Builder) Select(key, label string, options []Option) *Builder {
	b.node.AddField(Select(key, label, options))
	return b
}

func (b *Builder) Actions(actions []Action) *Builder {
	b.node.SetActions(actions)
	return b
}

func (b *Builder) Action(action Action) *Builder {
	b.node.AddAction(action)
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

func (b *Builder) RowGap(rowGap int) *Builder {
	b.node.SetRowGap(rowGap)
	return b
}

func (b *Builder) Wrap(wrap bool) *Builder {
	b.node.SetWrap(wrap)
	return b
}

func (b *Builder) LabelWidth(width int) *Builder {
	b.node.SetLabelWidth(width)
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetRootStyle(s)
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

// Of creates a FilterBar from fields.
func Of(fields []Field) rtui.VNode {
	return NewBuilder().Fields(fields).Build()
}
