package progress

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Builder provides a fluent API for building Progress VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Progress builder.
func NewBuilder() *Builder {
	return &Builder{node: New()}
}

func (b *Builder) Value(v int) *Builder {
	b.node.SetValue(v)
	return b
}

func (b *Builder) Max(v int) *Builder {
	b.node.SetMax(v)
	return b
}

func (b *Builder) Label(label string) *Builder {
	b.node.SetLabel(label)
	return b
}

func (b *Builder) Width(w int) *Builder {
	b.node.SetWidth(w)
	return b
}

func (b *Builder) ShowPercent(v bool) *Builder {
	b.node.SetShowPercent(v)
	return b
}

func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

func (b *Builder) Build() rtui.VNode {
	return b.node
}

func (b *Builder) BuildTyped() *VNode {
	return b.node
}

// =============================================================================
// Backward Compatibility - Aliases for old API
// =============================================================================

// Progress creates a new Progress VNode (for backward compatibility).
// This matches the old feedback.Progress() API.
func Progress() *VNode {
	return New()
}

// NewProgress creates a new Progress VNode (alias for New, for backward compatibility).
// This matches the old feedback.NewProgress() API.
func NewProgress() *VNode {
	return New()
}

// ProgressBuilder is an alias for Builder (for backward compatibility).
// This matches the old feedback.ProgressBuilder type.
type ProgressBuilder = Builder
