package text

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Builder - Fluent API for creating Text VNode
// =============================================================================

// Builder provides a fluent API for building Text VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Text builder.
func NewBuilder(content string) *Builder {
	return &Builder{
		node: New(content),
	}
}

// Content sets the text content.
func (b *Builder) Content(content string) *Builder {
	b.node.SetContent(content)
	return b
}

// Key sets the key for diffing.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// SetID sets the business identifier for positioning and Portal anchoring.
// This is separate from Key() which is used for list diffing.
func (b *Builder) SetID(id string) *Builder {
	b.node.SetID(id)
	return b
}

// Style sets the visual style.
func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyleProps(s)
	return b
}

// FgColor sets the foreground color.
func (b *Builder) FgColor(c interface{}) *Builder {
	if colorStr, ok := c.(string); ok {
		b.node.style.FG = style.Color(colorStr)
	} else if color, ok := c.(style.Color); ok {
		b.node.style.FG = color
	}
	return b
}

// BgColor sets the background color.
func (b *Builder) BgColor(c interface{}) *Builder {
	if colorStr, ok := c.(string); ok {
		b.node.style.BG = style.Color(colorStr)
	} else if color, ok := c.(style.Color); ok {
		b.node.style.BG = color
	}
	return b
}

// Bold sets the bold attribute.
func (b *Builder) Bold(v bool) *Builder {
	b.node.Bold(v)
	return b
}

// Italic sets the italic attribute.
func (b *Builder) Italic(v bool) *Builder {
	b.node.Italic(v)
	return b
}

// Underline sets the underline attribute.
func (b *Builder) Underline(v bool) *Builder {
	b.node.Underline(v)
	return b
}

// Padding sets the padding (top, right, bottom, left).
func (b *Builder) Padding(top, right, bottom, left int) *Builder {
	b.node.SetPaddingProps(top, right, bottom, left)
	return b
}

// TextAlign sets the text alignment.
func (b *Builder) TextAlign(align rtui.Align) *Builder {
	b.node.SetTextAlignProps(align)
	return b
}

// MaxWidth sets the maximum width.
func (b *Builder) MaxWidth(maxWidth int) *Builder {
	b.node.SetMaxWidth(maxWidth)
	return b
}

// Build returns the Text VNode.
func (b *Builder) Build() rtui.VNode {
	return b.node
}

// BuildInstance returns the Text VNode as a ComponentInstance.
// This is useful for direct instance creation without going through Fiber.
func (b *Builder) BuildInstance() rtui.ComponentInstance {
	return b.node.CreateInstance()
}

// =============================================================================
// Convenience Functions
// =============================================================================

// T creates a simple text node with the given content.
// This is a shorthand for New(content).
func T(content string) rtui.VNode {
	return New(content)
}

// Styled creates a text node with content and style.
func Styled(content string, s style.Style) rtui.VNode {
	return New(content).SetStyleProps(s)
}

// Bold creates a bold text node.
func Bold(content string) rtui.VNode {
	return New(content).Bold(true)
}

// Colored creates a text node with foreground color.
func Colored(content string, fg style.Color) rtui.VNode {
	return New(content).Foreground(fg)
}

// =============================================================================
// Backward Compatibility - Aliases for old API
// =============================================================================

// Text creates a simple text node (alias for T, for backward compatibility).
// This matches the old basic.Text() API.
func Text(content string) rtui.VNode {
	return New(content)
}

// NewText creates a new Text VNode (alias for New, for backward compatibility).
// This matches the old basic.NewText() API.
func NewText(content string) *VNode {
	return New(content)
}
