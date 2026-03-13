package rate

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for building Rate VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Rate builder.
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

// Count sets the total number of stars.
func (b *Builder) Count(v int) *Builder {
	b.node.SetCount(v)
	return b
}

// AllowClear controls whether max rating can be cleared back to 0 by toggling.
func (b *Builder) AllowClear(v bool) *Builder {
	b.node.SetAllowClear(v)
	return b
}

// Character sets the filled character used for active stars.
func (b *Builder) Character(ch string) *Builder {
	b.node.SetCharacter(ch)
	return b
}

// Disabled sets the disabled state.
func (b *Builder) Disabled(v bool) *Builder {
	b.node.SetDisabled(v)
	return b
}

// ShowValue controls whether the x/y value text is shown.
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

// ForField binds the rating value to a FieldChangeIntent.
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

// Rate creates a new Rate VNode.
func Rate() *VNode {
	return New()
}

// NewRate creates a new Rate VNode.
func NewRate() *VNode {
	return New()
}
