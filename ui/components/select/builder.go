package selectcomp

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for building Select VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Select builder.
func NewBuilder() *Builder {
	return &Builder{
		node: New(),
	}
}

// Options sets the options list.
func (b *Builder) Options(opts []Option) *Builder {
	b.node.SetOptions(opts)
	return b
}

// AddOption adds a single option.
func (b *Builder) AddOption(value, label string) *Builder {
	b.node.AddOption(value, label)
	return b
}

// Key sets the key for diffing.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// Selected sets the selected index.
func (b *Builder) Selected(idx int) *Builder {
	b.node.SetSelectedIndex(idx)
	return b
}

// Disabled sets the disabled state.
func (b *Builder) Disabled(v bool) *Builder {
	b.node.SetDisabled(v)
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

// OnChange sets the change intent.
func (b *Builder) OnChange(changeIntent intent.Intent) *Builder {
	b.node.SetChangeIntent(changeIntent)
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

// Select creates a new Select VNode (for backward compatibility).
// This matches the old form.Select() API.
func Select() *VNode {
	return New()
}

// NewSelect creates a new Select VNode (alias for New, for backward compatibility).
// This matches the old form.NewSelect() API.
func NewSelect() *VNode {
	return New()
}

// SelectBuilder is an alias for Builder (for backward compatibility).
// This matches the old form.SelectBuilder type.
type SelectBuilder = Builder

// SelectOption is an alias for Option (for backward compatibility).
// This matches the old form.SelectOption type.
type SelectOption = Option
