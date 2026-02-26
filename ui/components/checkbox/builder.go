package checkbox

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for building Checkbox VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Checkbox builder.
func NewBuilder() *Builder {
	return &Builder{
		node: New(""),
	}
}

// Label sets the label text.
func (b *Builder) Label(text string) *Builder {
	b.node.SetLabel(text)
	return b
}

// Key sets the key for diffing.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
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

// Style sets the visual style.
func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

// FgColor sets the foreground color.
func (b *Builder) FgColor(c interface{}) *Builder {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s = s.Foreground(style.Color(colorStr))
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s = s.Foreground(color)
		b.node.SetStyle(s)
	}
	return b
}

// BgColor sets the background color.
func (b *Builder) BgColor(c interface{}) *Builder {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s = s.Background(style.Color(colorStr))
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s = s.Background(color)
		b.node.SetStyle(s)
	}
	return b
}

// Bold sets the bold attribute.
func (b *Builder) Bold(v bool) *Builder {
	s := b.node.Style()
	s = s.Bold(v)
	b.node.SetStyle(s)
	return b
}

// OnToggle sets the toggle intent (replaces OnChange closure).
func (b *Builder) OnToggle(toggleIntent intent.Intent) *Builder {
	b.node.SetIntent(toggleIntent)
	return b
}

// ForField sets the field binding for MVP data flow.
// This creates a FieldBinding intent that will be used by the Instance
// to emit FieldChangeIntent when the user toggles the checkbox.
//
// Example:
//
//	checkbox.NewBuilder().ForField(intent.BindField("acceptTerms"))
func (b *Builder) ForField(binding intent.FieldBinding) *Builder {
	b.node.SetIntent(binding)
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

// =============================================================================
// Backward Compatibility - Aliases for old API
// =============================================================================

// Checkbox creates a new Checkbox VNode (for backward compatibility).
// This matches the old form.Checkbox() API.
func Checkbox(label string) *VNode {
	return New(label)
}

// NewCheckbox creates a new Checkbox VNode (alias for New, for backward compatibility).
// This matches the old form.NewCheckbox() API.
func NewCheckbox(label string) *VNode {
	return New(label)
}

// CheckboxBuilder is an alias for Builder (for backward compatibility).
// This matches the old form.CheckboxBuilder type.
type CheckboxBuilder = Builder
