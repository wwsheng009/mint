package optiongroup

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for building OptionGroup VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new OptionGroup builder.
func NewBuilder(options []Option) *Builder {
	return &Builder{
		node: New(options),
	}
}

// Label sets the component label text.
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

// Mode sets the selection mode (single or multiple).
func (b *Builder) Mode(mode SelectMode) *Builder {
	b.node.SetMode(mode)
	return b
}

// Selected sets the selected value (for ModeSingle).
func (b *Builder) Selected(selected string) *Builder {
	b.node.SetSelected(selected)
	return b
}

// Selecteds sets the selected values (for ModeMultiple).
func (b *Builder) Selecteds(selecteds []string) *Builder {
	b.node.SetSelecteds(selecteds)
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

// Orientation sets the layout orientation.
func (b *Builder) Orientation(orientation Orientation) *Builder {
	b.node.SetOrientation(orientation)
	return b
}

// Spacing sets the gap between options.
func (b *Builder) Spacing(spacing int) *Builder {
	b.node.SetSpacing(spacing)
	return b
}

// OnSelect sets the select intent (replaces closure).
func (b *Builder) OnSelect(selectIntent intent.Intent) *Builder {
	b.node.SetIntent(selectIntent)
	return b
}

// ForField sets the field binding for MVP data flow.
// This creates a FieldBinding intent that will be used by the Instance
// to emit FieldChangeIntent when the user selects an option.
//
// For ModeSingle: emits FieldChangeIntent with the selected value
// For ModeMultiple: emits FieldChangeIntent with comma-separated values
//
// Example (ModeSingle):
//
//	optiongroup.NewBuilder(options).
//		Mode(ModeSingle).
//		ForField(intent.BindField("city"))
//
// Example (ModeMultiple):
//
//	optiongroup.NewBuilder(options).
//		Mode(ModeMultiple).
//		ForField(intent.BindField("interests"))
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
// Convenience Methods
// =============================================================================

// Single sets mode to single-select (radio behavior).
func (b *Builder) Single() *Builder {
	b.node.SetMode(ModeSingle)
	return b
}

// Multiple sets mode to multi-select (checkbox behavior).
func (b *Builder) Multiple() *Builder {
	b.node.SetMode(ModeMultiple)
	return b
}

// Vertical sets orientation to vertical.
func (b *Builder) Vertical() *Builder {
	b.node.SetOrientation(OrientationVertical)
	return b
}

// Horizontal sets orientation to horizontal.
func (b *Builder) Horizontal() *Builder {
	b.node.SetOrientation(OrientationHorizontal)
	return b
}
