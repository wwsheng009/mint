package switchcomp

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for building Switch VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Switch builder.
func NewBuilder() *Builder {
	return &Builder{
		node: New(""),
	}
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

// SetID sets the business identifier for positioning and Portal anchoring.
func (b *Builder) SetID(id string) *Builder {
	b.node.SetID(id)
	return b
}

// Disabled sets the disabled state.
func (b *Builder) Disabled(v bool) *Builder {
	b.node.SetDisabled(v)
	return b
}

// Checked sets the checked state.
func (b *Builder) Checked(v bool) *Builder {
	b.node.SetChecked(v)
	return b
}

// CheckedLabel sets the checked-state track label.
func (b *Builder) CheckedLabel(label string) *Builder {
	b.node.SetCheckedLabel(label)
	return b
}

// UncheckedLabel sets the unchecked-state track label.
func (b *Builder) UncheckedLabel(label string) *Builder {
	b.node.SetUncheckedLabel(label)
	return b
}

// Labels sets both track labels.
func (b *Builder) Labels(checkedLabel, uncheckedLabel string) *Builder {
	b.node.SetLabels(checkedLabel, uncheckedLabel)
	return b
}

// Style sets the visual style.
func (b *Builder) Style(st style.Style) *Builder {
	b.node.SetStyle(st)
	return b
}

// FgColor sets the foreground color.
func (b *Builder) FgColor(c interface{}) *Builder {
	if colorStr, ok := c.(string); ok {
		st := b.node.Style()
		st = st.Foreground(style.Color(colorStr))
		b.node.SetStyle(st)
	} else if color, ok := c.(style.Color); ok {
		st := b.node.Style()
		st = st.Foreground(color)
		b.node.SetStyle(st)
	}
	return b
}

// BgColor sets the background color.
func (b *Builder) BgColor(c interface{}) *Builder {
	if colorStr, ok := c.(string); ok {
		st := b.node.Style()
		st = st.Background(style.Color(colorStr))
		b.node.SetStyle(st)
	} else if color, ok := c.(style.Color); ok {
		st := b.node.Style()
		st = st.Background(color)
		b.node.SetStyle(st)
	}
	return b
}

// Bold sets the bold attribute.
func (b *Builder) Bold(v bool) *Builder {
	st := b.node.Style()
	st = st.Bold(v)
	b.node.SetStyle(st)
	return b
}

// OnToggle sets the toggle intent.
func (b *Builder) OnToggle(toggleIntent intent.Intent) *Builder {
	b.node.SetIntent(toggleIntent)
	return b
}

// ForField binds the switch state to FieldChangeIntent.
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

// Switch creates a new Switch VNode.
func Switch(label string) *VNode {
	return New(label)
}

// NewSwitch creates a new Switch VNode.
func NewSwitch(label string) *VNode {
	return New(label)
}
