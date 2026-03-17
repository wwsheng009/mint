package pagination

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
)

// Builder provides a fluent API for creating Pagination VNodes.
type Builder struct {
	node *VNode
}

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

func (b *Builder) Total(total int) *Builder {
	b.node.SetTotal(total)
	return b
}

func (b *Builder) PageSize(pageSize int) *Builder {
	b.node.SetPageSize(pageSize)
	return b
}

func (b *Builder) CurrentPage(page int) *Builder {
	b.node.SetCurrentPage(page)
	return b
}

func (b *Builder) MaxButtons(maxButtons int) *Builder {
	b.node.SetMaxButtons(maxButtons)
	return b
}

func (b *Builder) ShowTotal(show bool) *Builder {
	b.node.SetShowTotal(show)
	return b
}

func (b *Builder) Disabled(disabled bool) *Builder {
	b.node.SetDisabled(disabled)
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

func (b *Builder) SelectedStyle(s style.Style) *Builder {
	b.node.SetSelectedStyle(s)
	return b
}

func (b *Builder) DisabledStyle(s style.Style) *Builder {
	b.node.SetDisabledStyle(s)
	return b
}

func (b *Builder) OnPageChange(i intent.Intent) *Builder {
	b.node.SetPageIntent(i)
	return b
}

func (b *Builder) PageForField(binding intent.FieldBinding) *Builder {
	b.node.SetPageIntentField(binding)
	return b
}

func (b *Builder) Build() *VNode {
	return b.node
}
