package popover

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for creating Popover VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Popover builder.
func NewBuilder(child rtui.VNode) *Builder {
	return &Builder{node: New(child)}
}

// Key sets the diff key.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// SetID sets the business identifier.
func (b *Builder) SetID(id string) *Builder {
	b.node.SetID(id)
	return b
}

// ComponentID sets the local intent routing ID.
func (b *Builder) ComponentID(id string) *Builder {
	b.node.SetComponentID(id)
	return b
}

// Child sets the anchor child.
func (b *Builder) Child(child rtui.VNode) *Builder {
	b.node.SetChild(child)
	return b
}

// Title sets the popover title.
func (b *Builder) Title(title string) *Builder {
	b.node.SetTitle(title)
	return b
}

// Body sets the body text.
func (b *Builder) Body(body string) *Builder {
	b.node.SetBody(body)
	return b
}

// Placement sets the placement mode.
func (b *Builder) Placement(placement Placement) *Builder {
	b.node.SetPlacement(placement)
	return b
}

// Trigger sets the trigger mode.
func (b *Builder) Trigger(trigger TriggerMode) *Builder {
	b.node.SetTrigger(trigger)
	return b
}

// Open sets the controlled open state.
func (b *Builder) Open(open bool) *Builder {
	b.node.SetOpen(open)
	return b
}

// InitialOpen sets the uncontrolled initial open state.
func (b *Builder) InitialOpen(open bool) *Builder {
	b.node.SetInitialOpen(open)
	return b
}

// Disabled toggles disabled state.
func (b *Builder) Disabled(disabled bool) *Builder {
	b.node.SetDisabled(disabled)
	return b
}

// ShowArrow toggles arrow rendering.
func (b *Builder) ShowArrow(show bool) *Builder {
	b.node.SetShowArrow(show)
	return b
}

// GapRows sets the vertical gap rows.
func (b *Builder) GapRows(rows int) *Builder {
	b.node.SetGapRows(rows)
	return b
}

// MaxWidth sets the maximum wrapped content width.
func (b *Builder) MaxWidth(width int) *Builder {
	b.node.SetMaxWidth(width)
	return b
}

// Style sets the root style.
func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

// OverlayStyle sets the overlay fill style.
func (b *Builder) OverlayStyle(s style.Style) *Builder {
	b.node.SetOverlayStyle(s)
	return b
}

// BorderStyle sets the border style.
func (b *Builder) BorderStyle(s style.Style) *Builder {
	b.node.SetBorderStyle(s)
	return b
}

// ShadowStyle sets the shadow style.
func (b *Builder) ShadowStyle(s style.Style) *Builder {
	b.node.SetShadowStyle(s)
	return b
}

// TitleStyle sets the title style.
func (b *Builder) TitleStyle(s style.Style) *Builder {
	b.node.SetTitleStyle(s)
	return b
}

// BodyStyle sets the body style.
func (b *Builder) BodyStyle(s style.Style) *Builder {
	b.node.SetBodyStyle(s)
	return b
}

// OnChange sets the optional custom change intent.
func (b *Builder) OnChange(i intent.Intent) *Builder {
	b.node.SetChangeIntent(i)
	return b
}

// OpenForField binds open state to a field.
func (b *Builder) OpenForField(binding intent.FieldBinding) *Builder {
	b.node.SetChangeIntentField(binding)
	return b
}

// Build returns the configured VNode.
func (b *Builder) Build() rtui.VNode {
	return b.node
}

// BuildVNode returns the concrete VNode.
func (b *Builder) BuildVNode() *VNode {
	return b.node
}
