package basic

import (
	"unicode/utf8"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// TextVNode represents a text node with content
// This is the concrete implementation of ui.VNode for text
type TextVNode struct {
	content string
	key     string
	props   ui.Props
	style   style.Style
}

// NewText creates a new text VNode
func NewText(content string) *TextVNode {
	return &TextVNode{
		content: content,
		props:   make(ui.Props),
		style:   style.Style{},
	}
}

// Type implements ui.VNode
func (t *TextVNode) Type() ui.VNodeType {
	return ui.VNodeText
}

// Props implements ui.VNode
func (t *TextVNode) Props() ui.Props {
	return t.props
}

// SetProps implements ui.VNode
func (t *TextVNode) SetProps(p ui.Props) {
	t.props = p
}

// Children implements ui.VNode (text nodes have no children)
func (t *TextVNode) Children() []ui.VNode {
	return nil
}

// SetChildren implements ui.VNode (text nodes have no children)
func (t *TextVNode) SetChildren(children []ui.VNode) {
	// Text nodes don't have children
}

// Key implements ui.VNode
func (t *TextVNode) Key() string {
	return t.key
}

// SetKey implements ui.VNode
func (t *TextVNode) SetKey(key string) {
	t.key = key
}

// Style implements ui.VNode
func (t *TextVNode) Style() style.Style {
	return t.style
}

// SetStyle implements ui.VNode
func (t *TextVNode) SetStyle(s style.Style) {
	t.style = s
}

// Tag implements ui.VNode - returns "text"
func (t *TextVNode) Tag() string {
	return "text"
}

// Content returns the text content
func (t *TextVNode) Content() string {
	return t.content
}

// SetContent sets the text content
func (t *TextVNode) SetContent(content string) *TextVNode {
	t.content = content
	return t
}

// =============================================================================
// Measurable Interface Implementation
// =============================================================================

// Measure implements runtime.Measurable interface
// Calculates the size of the text based on content and constraints
func (t *TextVNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
	if t == nil {
		return runtime.Size{Width: 0, Height: 0}
	}

	// Calculate text width (number of visible characters)
	content := t.content
	if content == "" {
		content = " " // Empty text still has minimal width
	}

	// Simple width calculation: count rune width
	width := utf8.RuneCountInString(content)

	// Height is always 1 for single-line text
	height := 1

	// Apply constraints
	if width < constraints.MinWidth {
		width = constraints.MinWidth
	}
	if width > constraints.MaxWidth && constraints.MaxWidth > 0 {
		width = constraints.MaxWidth
	}
	if height < constraints.MinHeight {
		height = constraints.MinHeight
	}
	if height > constraints.MaxHeight && constraints.MaxHeight > 0 {
		height = constraints.MaxHeight
	}

	// Apply explicit style dimensions if set
	if t.style.Width > 0 {
		width = t.style.Width
	}
	if t.style.Height > 0 {
		height = t.style.Height
	}

	return runtime.Size{Width: width, Height: height}
}

// =============================================================================
// Paintable Interface Implementation
// =============================================================================

// Paint implements paint.Paintable interface
// Generates draw commands for rendering this text component
func (t *TextVNode) Paint(x, y int) []paint.DrawCmd {
	if t == nil {
		return nil
	}

	return []paint.DrawCmd{
		paint.NewTextCmd(x, y, t.content, t.style),
	}
}

// Text creates a new text node (simple version)
func Text(content string) ui.VNode {
	return NewText(content)
}

// TextBuilder provides fluent API for building text nodes
type TextBuilder struct {
	node *TextVNode
}

// NewTextBuilder creates a text builder for chained calls
func NewTextBuilder(content string) *TextBuilder {
	return &TextBuilder{
		node: NewText(content),
	}
}

// Content sets the text content
func (b *TextBuilder) Content(content string) *TextBuilder {
	b.node.SetContent(content)
	return b
}

// Key sets the key for diffing
func (b *TextBuilder) Key(key string) *TextBuilder {
	b.node.SetKey(key)
	return b
}

// Style sets the visual style
func (b *TextBuilder) Style(s style.Style) *TextBuilder {
	b.node.SetStyle(s)
	return b
}

// FgColor sets the foreground color
// Accepts either style.Color (string) or a string
func (b *TextBuilder) FgColor(c interface{}) *TextBuilder {
	if colorStr, ok := c.(string); ok {
		b.node.style.FG = style.Color(colorStr)
	} else if color, ok := c.(style.Color); ok {
		b.node.style.FG = color
	}
	return b
}

// BgColor sets the background color
func (b *TextBuilder) BgColor(c interface{}) *TextBuilder {
	if colorStr, ok := c.(string); ok {
		b.node.style.BG = style.Color(colorStr)
	} else if color, ok := c.(style.Color); ok {
		b.node.style.BG = color
	}
	return b
}

// Bold sets the bold attribute
func (b *TextBuilder) Bold(v bool) *TextBuilder {
	b.node.style = b.node.style.Bold(v)
	return b
}

// Italic sets the italic attribute
func (b *TextBuilder) Italic(v bool) *TextBuilder {
	b.node.style = b.node.style.Italic(v)
	return b
}

// Underline sets the underline attribute
func (b *TextBuilder) Underline(v bool) *TextBuilder {
	b.node.style = b.node.style.Underline(v)
	return b
}

// Build returns the ui.VNode
func (b *TextBuilder) Build() ui.VNode {
	return b.node
}
