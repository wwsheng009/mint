package ui

import "github.com/wwsheng009/mint/runtime/style"

// TextVNode represents a text node with content
type TextVNode struct {
	content string
	key     string
	props   Props
	style   style.Style
}

// NewText creates a new text VNode
func NewText(content string) *TextVNode {
	return &TextVNode{
		content: content,
		props:   make(Props),
		style:   style.Style{},
	}
}

// Type implements VNode
func (t *TextVNode) Type() VNodeType {
	return VNodeText
}

// Props implements VNode
func (t *TextVNode) Props() Props {
	return t.props
}

// SetProps implements VNode
func (t *TextVNode) SetProps(p Props) {
	t.props = p
}

// Children implements VNode (text nodes have no children)
func (t *TextVNode) Children() []VNode {
	return nil
}

// SetChildren implements VNode (text nodes have no children)
func (t *TextVNode) SetChildren(children []VNode) {
	// Text nodes don't have children
}

// Key implements VNode
func (t *TextVNode) Key() string {
	return t.key
}

// SetKey implements VNode
func (t *TextVNode) SetKey(key string) {
	t.key = key
}

// Style implements VNode
func (t *TextVNode) Style() style.Style {
	return t.style
}

// SetStyle implements VNode
func (t *TextVNode) SetStyle(s style.Style) {
	t.style = s
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

// Text creates a new text node (simple version)
func Text(content string) VNode {
	return NewText(content)
}

// NewTextBuilder creates a text builder for chained calls
func NewTextBuilder(content string) *TextBuilder {
	return &TextBuilder{
		node: NewText(content),
	}
}

// TextBuilder provides fluent API for building text nodes
type TextBuilder struct {
	node *TextVNode
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

// Build returns the VNode
func (b *TextBuilder) Build() VNode {
	return b.node
}
