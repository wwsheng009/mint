package popconfirm

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
)

// Builder provides a fluent API for creating Popconfirm VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Popconfirm builder.
func NewBuilder(child rtui.VNode) *Builder {
	return &Builder{node: New(child)}
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

func (b *Builder) AnchorID(id string) *Builder {
	b.node.SetAnchorID(id)
	return b
}

func (b *Builder) Child(child rtui.VNode) *Builder {
	b.node.SetChild(child)
	return b
}

func (b *Builder) Title(title string) *Builder {
	b.node.SetTitle(title)
	return b
}

func (b *Builder) Description(description string) *Builder {
	b.node.SetDescription(description)
	return b
}

func (b *Builder) Placement(placement Placement) *Builder {
	b.node.SetPlacement(placement)
	return b
}

func (b *Builder) Trigger(trigger TriggerMode) *Builder {
	b.node.SetTrigger(trigger)
	return b
}

func (b *Builder) Open(open bool) *Builder {
	b.node.SetOpen(open)
	return b
}

func (b *Builder) InitialOpen(open bool) *Builder {
	b.node.SetInitialOpen(open)
	return b
}

func (b *Builder) Disabled(disabled bool) *Builder {
	b.node.SetDisabled(disabled)
	return b
}

func (b *Builder) ShowArrow(show bool) *Builder {
	b.node.SetShowArrow(show)
	return b
}

func (b *Builder) ShowCancel(show bool) *Builder {
	b.node.SetShowCancel(show)
	return b
}

func (b *Builder) GapRows(rows int) *Builder {
	b.node.SetGapRows(rows)
	return b
}

func (b *Builder) MaxWidth(width int) *Builder {
	b.node.SetMaxWidth(width)
	return b
}

func (b *Builder) OkText(text string) *Builder {
	b.node.SetOkText(text)
	return b
}

func (b *Builder) CancelText(text string) *Builder {
	b.node.SetCancelText(text)
	return b
}

func (b *Builder) OkVariant(variant button.Variant) *Builder {
	b.node.SetOkVariant(variant)
	return b
}

func (b *Builder) CancelVariant(variant button.Variant) *Builder {
	b.node.SetCancelVariant(variant)
	return b
}

func (b *Builder) FooterLayout(layout FooterLayout) *Builder {
	b.node.SetFooterLayout(layout)
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

func (b *Builder) OverlayStyle(s style.Style) *Builder {
	b.node.SetOverlayStyle(s)
	return b
}

func (b *Builder) TitleStyle(s style.Style) *Builder {
	b.node.SetTitleStyle(s)
	return b
}

func (b *Builder) TextStyle(s style.Style) *Builder {
	b.node.SetTextStyle(s)
	return b
}

func (b *Builder) OkButtonStyle(s style.Style) *Builder {
	b.node.SetOkButtonStyle(s)
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

func (b *Builder) OnChange(i intent.Intent) *Builder {
	b.node.SetChangeIntent(i)
	return b
}

func (b *Builder) OpenForField(binding intent.FieldBinding) *Builder {
	b.node.SetChangeIntentField(binding)
	return b
}

func (b *Builder) Build() rtui.VNode {
	return b.node
}

func (b *Builder) BuildVNode() *VNode {
	return b.node
}
