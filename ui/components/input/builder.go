package input

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for building Input VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Input builder.
func NewBuilder() *Builder {
	return &Builder{
		node: New(),
	}
}

// Placeholder sets the placeholder text.
func (b *Builder) Placeholder(text string) *Builder {
	b.node.SetPlaceholder(text)
	return b
}

// Value sets the initial value.
func (b *Builder) Value(value string) *Builder {
	b.node.SetValue(value)
	return b
}

// Key sets the key for diffing.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// Type sets the input type.
func (b *Builder) Type(t Type) *Builder {
	b.node.SetType(t)
	return b
}

// Password sets the input type to password.
func (b *Builder) Password() *Builder {
	b.node.SetPassword()
	return b
}

// MaxLen sets the maximum length.
func (b *Builder) MaxLen(len int) *Builder {
	b.node.SetMaxLen(len)
	return b
}

// Disabled sets the disabled state.
func (b *Builder) Disabled(v bool) *Builder {
	b.node.SetDisabled(v)
	return b
}

// ReadOnly sets the read-only state.
func (b *Builder) ReadOnly(v bool) *Builder {
	b.node.SetReadOnly(v)
	return b
}

// Width sets the explicit width.
func (b *Builder) Width(w int) *Builder {
	b.node.SetWidth(w)
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

// OnChange sets the change intent.
func (b *Builder) OnChange(changeIntent intent.Intent) *Builder {
	b.node.SetChangeIntent(changeIntent)
	return b
}

// OnSubmit sets the submit intent.
func (b *Builder) OnSubmit(submitIntent intent.Intent) *Builder {
	b.node.SetSubmitIntent(submitIntent)
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
