package transfer

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for creating Transfer VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Transfer builder.
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

func (b *Builder) Items(items []Item) *Builder {
	b.node.SetItems(items)
	return b
}

func (b *Builder) AddItem(item Item) *Builder {
	b.node.AddItem(item)
	return b
}

func (b *Builder) Titles(source, target string) *Builder {
	b.node.SetTitles(source, target)
	return b
}

func (b *Builder) Operations(toTarget, toSource string) *Builder {
	b.node.SetOperations(toTarget, toSource)
	return b
}

func (b *Builder) Searchable(searchable bool) *Builder {
	b.node.SetSearchable(searchable)
	return b
}

func (b *Builder) SearchPlaceholders(source, target string) *Builder {
	b.node.SetSearchPlaceholders(source, target)
	return b
}

func (b *Builder) SearchValues(source, target string) *Builder {
	b.node.SetSearchValues(source, target)
	return b
}

func (b *Builder) InitialSearchValues(source, target string) *Builder {
	b.node.SetInitialSearchValues(source, target)
	return b
}

func (b *Builder) TargetKeys(keys []string) *Builder {
	b.node.SetTargetKeys(keys)
	return b
}

func (b *Builder) InitialTargetKeys(keys []string) *Builder {
	b.node.SetInitialTargetKeys(keys)
	return b
}

func (b *Builder) ListWidth(width int) *Builder {
	b.node.SetListWidth(width)
	return b
}

func (b *Builder) ListHeight(height int) *Builder {
	b.node.SetListHeight(height)
	return b
}

func (b *Builder) Width(width int) *Builder {
	b.node.SetWidth(width)
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyleProps(s)
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

// Of creates a Transfer with the given items.
func Of(items []Item) rtui.VNode {
	return NewBuilder().Items(items).Build()
}
