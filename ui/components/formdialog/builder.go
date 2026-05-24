package formdialog

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/form"
)

// Builder provides a fluent API for constructing FormDialog VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a FormDialog builder.
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

func (b *Builder) Description(description string) *Builder {
	b.node.SetDescription(description)
	return b
}

func (b *Builder) Open(open bool) *Builder {
	b.node.SetOpen(open)
	return b
}

func (b *Builder) Opened() *Builder {
	return b.Open(true)
}

func (b *Builder) Closed() *Builder {
	return b.Open(false)
}

func (b *Builder) Width(width int) *Builder {
	b.node.SetWidth(width)
	return b
}

func (b *Builder) Height(height int) *Builder {
	b.node.SetHeight(height)
	return b
}

func (b *Builder) Size(width, height int) *Builder {
	b.node.SetWidth(width)
	b.node.SetHeight(height)
	return b
}

func (b *Builder) FormID(formID string) *Builder {
	b.node.SetFormID(formID)
	return b
}

func (b *Builder) Layout(layout form.FormLayout) *Builder {
	b.node.SetLayout(layout)
	return b
}

func (b *Builder) Values(values map[string]interface{}) *Builder {
	b.node.SetValues(values)
	return b
}

func (b *Builder) Value(field string, value interface{}) *Builder {
	b.node.SetValue(field, value)
	return b
}

func (b *Builder) ValidateAll(validate bool) *Builder {
	b.node.SetValidateAll(validate)
	return b
}

func (b *Builder) Child(child rtui.VNode) *Builder {
	b.node.AddChild(child)
	return b
}

func (b *Builder) Children(children ...rtui.VNode) *Builder {
	b.node.AddChildren(children...)
	return b
}

func (b *Builder) SubmitText(text string) *Builder {
	b.node.SetSubmitText(text)
	return b
}

func (b *Builder) CancelText(text string) *Builder {
	b.node.SetCancelText(text)
	return b
}

func (b *Builder) SubmitVariant(variant button.Variant) *Builder {
	b.node.SetSubmitVariant(variant)
	return b
}

func (b *Builder) SubmitDisabled(disabled bool) *Builder {
	b.node.SetSubmitDisabled(disabled)
	return b
}

func (b *Builder) DisabledReason(reason string) *Builder {
	b.node.SetDisabledReason(reason)
	return b
}

func (b *Builder) OnSubmit(i intent.Intent) *Builder {
	b.node.SetSubmitIntent(i)
	return b
}

func (b *Builder) OnCancel(i intent.Intent) *Builder {
	b.node.SetCancelIntent(i)
	return b
}

func (b *Builder) OnClose(i intent.Intent) *Builder {
	b.node.SetCloseIntent(i)
	return b
}

func (b *Builder) Closeable(closeable bool) *Builder {
	b.node.SetCloseable(closeable)
	return b
}

func (b *Builder) CloseOnEsc(closeOnEsc bool) *Builder {
	b.node.SetCloseOnEsc(closeOnEsc)
	return b
}

func (b *Builder) CloseOnBackdrop(closeOnBackdrop bool) *Builder {
	b.node.SetCloseOnBackdrop(closeOnBackdrop)
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
