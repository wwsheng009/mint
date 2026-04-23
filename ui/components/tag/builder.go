package tag

import (
	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for creating Tag VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Tag builder with the given text.
func NewBuilder(text string) *Builder {
	return &Builder{node: New(text)}
}

// Key sets the key for diffing.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// Text sets the tag text.
func (b *Builder) Text(text string) *Builder {
	b.node.SetText(text)
	return b
}

// Color sets the tag color.
func (b *Builder) Color(c TagColor) *Builder {
	b.node.SetColor(c)
	return b
}

// Closable sets whether the tag can be closed.
func (b *Builder) Closable(closable bool) *Builder {
	b.node.SetClosable(closable)
	return b
}

// Icon sets the icon prefix.
func (b *Builder) Icon(icon string) *Builder {
	b.node.SetIcon(icon)
	return b
}

// CloseIntent sets the intent to emit when the tag is closed.
func (b *Builder) CloseIntent(ci interface{}) *Builder {
	b.node.SetCloseIntent(ci)
	return b
}

// Style sets the visual style.
func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetTagStyle(s)
	return b
}

// Default sets the tag color to default.
func (b *Builder) Default() *Builder {
	b.node.Default()
	return b
}

// Primary sets the tag color to primary.
func (b *Builder) Primary() *Builder {
	b.node.Primary()
	return b
}

// Success sets the tag color to success.
func (b *Builder) Success() *Builder {
	b.node.Success()
	return b
}

// Warning sets the tag color to warning.
func (b *Builder) Warning() *Builder {
	b.node.Warning()
	return b
}

// Error sets the tag color to error.
func (b *Builder) Error() *Builder {
	b.node.Error()
	return b
}

// Processing sets the tag color to processing.
func (b *Builder) Processing() *Builder {
	b.node.Processing()
	return b
}

// Build returns the constructed Tag VNode.
func (b *Builder) Build() *VNode {
	return b.node
}

// Tag is a shortcut constructor: tag.Tag("text").
func Tag(text string) *VNode {
	return New(text)
}
