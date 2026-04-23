package empty

import (
	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for creating Empty VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Empty builder with default description.
func NewBuilder() *Builder {
	return &Builder{node: New()}
}

// Key sets the key for diffing.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// Description sets the description text shown below the image.
func (b *Builder) Description(desc string) *Builder {
	b.node.SetDescription(desc)
	return b
}

// Image sets a custom ASCII art / icon string.
func (b *Builder) Image(img string) *Builder {
	b.node.SetImage(img)
	return b
}

// Style applies a custom style to the empty component.
func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyle(s)
	return b
}

// Build returns the constructed VNode.
func (b *Builder) Build() *VNode {
	return b.node
}
