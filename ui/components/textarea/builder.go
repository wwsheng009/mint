package textarea

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for building Textarea VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Textarea builder.
func NewBuilder() *Builder {
	return &Builder{node: New()}
}

func (b *Builder) Placeholder(text string) *Builder {
	b.node.SetPlaceholder(text)
	return b
}

func (b *Builder) Value(value string) *Builder {
	b.node.SetValue(value)
	return b
}

func (b *Builder) Rows(rows int) *Builder {
	b.node.SetRows(rows)
	return b
}

func (b *Builder) Cols(cols int) *Builder {
	b.node.SetCols(cols)
	return b
}

func (b *Builder) MaxLen(len int) *Builder {
	b.node.SetMaxLen(len)
	return b
}

func (b *Builder) Disabled(v bool) *Builder {
	b.node.SetDisabled(v)
	return b
}

func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

func (b *Builder) OnChange(i intent.Intent) *Builder {
	b.node.SetChangeIntent(i)
	return b
}

func (b *Builder) OnSubmit(i intent.Intent) *Builder {
	b.node.SetSubmitIntent(i)
	return b
}

func (b *Builder) Build() rtui.VNode {
	return b.node
}

func (b *Builder) BuildTyped() *VNode {
	return b.node
}
