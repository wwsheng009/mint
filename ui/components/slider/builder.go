package slider

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for building Slider VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Slider builder.
func NewBuilder() *Builder {
	return &Builder{node: New()}
}

// Label sets the descriptive label.
func (b *Builder) Label(text string) *Builder {
	b.node.SetLabel(text)
	return b
}

// Key sets the key for diffing.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// SetID sets the business identifier.
func (b *Builder) SetID(id string) *Builder {
	b.node.SetID(id)
	return b
}

// Value sets the current value.
func (b *Builder) Value(v int) *Builder {
	b.node.SetValue(v)
	return b
}

// Min sets the minimum value.
func (b *Builder) Min(v int) *Builder {
	b.node.SetMin(v)
	return b
}

// Max sets the maximum value.
func (b *Builder) Max(v int) *Builder {
	b.node.SetMax(v)
	return b
}

// Step sets the step increment.
func (b *Builder) Step(v int) *Builder {
	b.node.SetStep(v)
	return b
}

// Width sets the track width in characters.
func (b *Builder) Width(v int) *Builder {
	b.node.SetWidth(v)
	return b
}

// Disabled sets the disabled state.
func (b *Builder) Disabled(v bool) *Builder {
	b.node.SetDisabled(v)
	return b
}

// ShowValue controls whether the numeric value is shown.
func (b *Builder) ShowValue(v bool) *Builder {
	b.node.SetShowValue(v)
	return b
}

// Style sets the visual style.
func (b *Builder) Style(st style.Style) *Builder {
	b.node.SetStyleProps(st)
	return b
}

// OnChange sets the intent to emit on value change.
func (b *Builder) OnChange(i intent.Intent) *Builder {
	b.node.SetIntent(i)
	return b
}

// ForField binds the slider value to a FieldChangeIntent.
func (b *Builder) ForField(binding intent.FieldBinding) *Builder {
	b.node.SetIntent(binding)
	return b
}

// ForForm sets the form ID for Form integration.
func (b *Builder) ForForm(binding intent.FormBinding) *Builder {
	b.node.SetFormID(binding.GetFormID())
	return b
}

// Build returns the VNode.
func (b *Builder) Build() rtui.VNode {
	return b.node
}

// BuildTyped returns the typed VNode.
func (b *Builder) BuildTyped() *VNode {
	return b.node
}

// BuildInstance returns the runtime instance.
func (b *Builder) BuildInstance() *Instance {
	return NewInstance(b.node.Props())
}

// Slider creates a new Slider VNode.
func Slider() *VNode {
	return New()
}

// NewSlider creates a new Slider VNode.
func NewSlider() *VNode {
	return New()
}
