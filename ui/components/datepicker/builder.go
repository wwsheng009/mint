package datepicker

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for creating DatePicker VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new DatePicker builder.
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

func (b *Builder) Width(width int) *Builder {
	b.node.SetWidth(width)
	return b
}

func (b *Builder) Placeholder(placeholder string) *Builder {
	b.node.SetPlaceholder(placeholder)
	return b
}

func (b *Builder) Value(value string) *Builder {
	b.node.SetValue(value)
	return b
}

func (b *Builder) DefaultValue(value string) *Builder {
	b.node.SetDefaultValue(value)
	return b
}

func (b *Builder) Disabled(disabled bool) *Builder {
	b.node.SetDisabled(disabled)
	return b
}

func (b *Builder) PortalRoot(root string) *Builder {
	b.node.SetPopupPortalRoot(root)
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

func (b *Builder) OnChange(changeIntent intent.Intent) *Builder {
	b.node.SetChangeIntent(changeIntent)
	return b
}

func (b *Builder) ForField(binding intent.FieldBinding) *Builder {
	b.node.SetProps(rtui.Props{
		propChangeIntent: binding,
	})
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
