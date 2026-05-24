package confirmdialog

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
)

// Builder provides a fluent API for constructing ConfirmDialog VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a ConfirmDialog builder.
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

func (b *Builder) Message(message string) *Builder {
	b.node.SetMessage(message)
	return b
}

func (b *Builder) Warning(warning string) *Builder {
	b.node.SetWarning(warning)
	return b
}

func (b *Builder) Open(open bool) *Builder {
	b.node.SetOpen(open)
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

func (b *Builder) TargetItems(items []TargetItem) *Builder {
	b.node.SetTargetItems(items)
	return b
}

func (b *Builder) Target(item TargetItem) *Builder {
	b.node.AddTargetItem(item)
	return b
}

func (b *Builder) ReasonLabel(label string) *Builder {
	b.node.SetReasonLabel(label)
	return b
}

func (b *Builder) ReasonValue(value string) *Builder {
	b.node.SetReasonValue(value)
	return b
}

func (b *Builder) ReasonField(field string) *Builder {
	b.node.SetReasonField(field)
	return b
}

func (b *Builder) ReasonPlaceholder(placeholder string) *Builder {
	b.node.SetReasonPlaceholder(placeholder)
	return b
}

func (b *Builder) ReasonRequired(required bool) *Builder {
	b.node.SetReasonRequired(required)
	return b
}

func (b *Builder) ConfirmText(text string) *Builder {
	b.node.SetConfirmText(text)
	return b
}

func (b *Builder) CancelText(text string) *Builder {
	b.node.SetCancelText(text)
	return b
}

func (b *Builder) ConfirmVariant(variant button.Variant) *Builder {
	b.node.SetConfirmVariant(variant)
	return b
}

func (b *Builder) DisableConfirm(disabled bool) *Builder {
	b.node.SetDisableConfirm(disabled)
	return b
}

func (b *Builder) DisabledReason(reason string) *Builder {
	b.node.SetDisabledReason(reason)
	return b
}

func (b *Builder) OnConfirm(i intent.Intent) *Builder {
	b.node.SetConfirmIntent(i)
	return b
}

func (b *Builder) OnCancel(i intent.Intent) *Builder {
	b.node.SetCancelIntent(i)
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
