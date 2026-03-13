package spin

import (
	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for creating Spin VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Spin builder (spinning by default).
func NewBuilder() *Builder {
	return &Builder{node: New()}
}

// Key sets the key for diffing.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// Spinning sets whether the spinner is active.
func (b *Builder) Spinning(s bool) *Builder {
	b.node.SetSpinning(s)
	return b
}

// Tip sets the loading tip text shown next to the spinner.
func (b *Builder) Tip(tip string) *Builder {
	b.node.SetTip(tip)
	return b
}

// Size sets the spinner size.
func (b *Builder) Size(s Size) *Builder {
	b.node.SetSize(s)
	return b
}

// Small sets the spinner to small size.
func (b *Builder) Small() *Builder {
	b.node.Small()
	return b
}

// Default sets the spinner to default size.
func (b *Builder) Default() *Builder {
	b.node.Default()
	return b
}

// Large sets the spinner to large size.
func (b *Builder) Large() *Builder {
	b.node.Large()
	return b
}

// Delay sets the delay in milliseconds before the spinner appears.
func (b *Builder) Delay(ms int) *Builder {
	b.node.SetDelay(ms)
	return b
}

// Style sets additional style overrides.
func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetSpinStyle(s)
	return b
}

// Build returns the configured VNode.
func (b *Builder) Build() *VNode {
	return b.node
}
