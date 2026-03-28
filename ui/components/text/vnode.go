// Package text provides a Fiber-first Text component.
// Text displays text with optional styling and word wrapping.
package text

import (
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Prop Keys
// =============================================================================

// Prop key constants — shared by VNode and Instance to avoid magic strings.
const (
	propContent = "content"
	propKey = "key"
	propMaxWidth = "maxWidth"
	propPadding = "padding"
	propStyle = "style"
	propTextAlign = "textAlign"
	propWrap = "wrap"
)

// =============================================================================
// VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the text component description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Content ===
	content string

	// === Visual Props ===
	style style.Style

	// === Layout Props ===
	padding   [4]int // top, right, bottom, left
	textAlign rtui.Align
	maxWidth  int  // optional max width for wrapping (deprecated: use Wrap() instead)
	wrap      bool // enable word wrapping

	// === Box Model (via interface) ===
	rtui.BoxModelMixin
}

// Ensure VNode implements required interfaces
var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
	_ rtui.BoxModel        = (*VNode)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// New creates a new Text VNode with the given content.
func New(content string) *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("text"),
		content:      content,
		style:        style.Style{},
		padding:      [4]int{},
		textAlign:    rtui.AlignStart,
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

// Type returns VNodeText to identify this as a text node.
func (t *VNode) Type() rtui.VNodeType {
	return rtui.VNodeText
}

// Key returns the component key.
func (t *VNode) Key() string {
	return t.key
}

// SetKey sets the component key - returns VNode for chaining.
func (t *VNode) SetKey(key string) rtui.VNode {
	t.key = key
	return t
}

// Tag returns the tag name.
func (t *VNode) Tag() string {
	return "text"
}

// Style returns the visual style.
func (t *VNode) Style() style.Style {
	return t.style
}

// SetStyle sets the visual style - returns VNode for chaining.
func (t *VNode) SetStyle(s style.Style) rtui.VNode {
	t.style = s
	return t
}

// Children returns child nodes (text has no children).
func (t *VNode) Children() []rtui.VNode {
	return nil
}

// SetChildren is a no-op for text - returns VNode for chaining.
func (t *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	// Text has no children
	return t
}

// GetLayer returns the rendering layer.
func (t *VNode) GetLayer() rtui.Layer {
	return rtui.LayerBase
}

// SetLayer sets the rendering layer - returns VNode for chaining.
func (t *VNode) SetLayer(l rtui.Layer) rtui.VNode {
	return t
}

// Props returns the node properties.
func (t *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:       t.key,
		propContent:   t.content,
		propStyle:     t.style,
		propPadding:   t.padding,
		propTextAlign: t.textAlign,
		propMaxWidth:  t.maxWidth,
		propWrap:      t.wrap,
	}
}

// SetProps sets the node properties - returns VNode for chaining.
func (t *VNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p[propKey].(string); ok {
		t.key = v
	}
	if v, ok := p[propContent].(string); ok {
		t.content = v
	}
	if v, ok := p[propStyle].(style.Style); ok {
		t.style = v
	}
	if v, ok := p[propPadding].([4]int); ok {
		t.padding = v
	}
	if v, ok := p[propTextAlign].(rtui.Align); ok {
		t.textAlign = v
	}
	if v, ok := p[propMaxWidth].(int); ok {
		t.maxWidth = v
	}
	if v, ok := p[propWrap].(bool); ok {
		t.wrap = v
	}
	return t
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

// CreateInstance creates a new TextInstance from this VNode description.
func (t *VNode) CreateInstance() rtui.ComponentInstance {
	props := rtui.Props{
		propKey:       t.key,
		propContent:   t.content,
		propStyle:     t.style,
		propPadding:   t.Padding(),
		propTextAlign: t.TextAlign(),
		propMaxWidth:  t.maxWidth,
		propWrap:      t.wrap,
	}
	return NewInstance(props)
}

// =============================================================================
// Builder Methods - Fluent API (return *VNode for chaining)
// =============================================================================

// SetContent sets the text content.
func (t *VNode) SetContent(content string) *VNode {
	t.content = content
	return t
}

// SetStyleProps sets the visual style.
func (t *VNode) SetStyleProps(s style.Style) *VNode {
	t.style = s
	return t
}

// SetPaddingProps sets the padding (top, right, bottom, left).
func (t *VNode) SetPaddingProps(top, right, bottom, left int) *VNode {
	t.BoxModelMixin.SetPadding(top, right, bottom, left)
	return t
}

// SetTextAlignProps sets the text alignment.
func (t *VNode) SetTextAlignProps(align rtui.Align) *VNode {
	t.BoxModelMixin.SetTextAlign(align)
	return t
}

// SetMaxWidth sets the maximum width.
func (t *VNode) SetMaxWidth(maxWidth int) *VNode {
	t.maxWidth = maxWidth
	return t
}

// SetWrap enables or disables word wrapping.
func (t *VNode) SetWrap(wrap bool) *VNode {
	t.wrap = wrap
	return t
}

// =============================================================================
// Style Builder Methods
// =============================================================================

// Foreground sets the foreground color.
func (t *VNode) Foreground(fg style.Color) *VNode {
	t.style = t.style.Foreground(fg)
	return t
}

// Background sets the background color.
func (t *VNode) Background(bg style.Color) *VNode {
	t.style = t.style.Background(bg)
	return t
}

// Bold sets the bold attribute.
func (t *VNode) Bold(bold bool) *VNode {
	t.style = t.style.Bold(bold)
	return t
}

// Italic sets the italic attribute.
func (t *VNode) Italic(italic bool) *VNode {
	t.style = t.style.Italic(italic)
	return t
}

// Underline sets the underline attribute.
func (t *VNode) Underline(underline bool) *VNode {
	t.style = t.style.Underline(underline)
	return t
}

// =============================================================================
// Props Accessors (for Instance creation)
// =============================================================================

// Content returns the text content.
func (t *VNode) Content() string {
	return t.content
}

// MaxWidth returns the maximum width.
func (t *VNode) MaxWidth() int {
	return t.maxWidth
}

// =============================================================================
// layout.BoxModelProvider Implementation
// =============================================================================

// GetBoxModel returns the box model for the Text VNode.
// Implements layout.BoxModelProvider for unified padding/border handling.
// Note: Text uses BoxModelMixin for padding/margin, and has no border.
func (t *VNode) GetBoxModel() layout.BoxModel {
	return layout.BoxModel{
		Padding: layout.Padding{
			Left:   t.BoxModelMixin.Padding()[3],
			Right:  t.BoxModelMixin.Padding()[1],
			Top:    t.BoxModelMixin.Padding()[0],
			Bottom: t.BoxModelMixin.Padding()[2],
		},
		Margin: layout.Margin{
			Left:   t.BoxModelMixin.Margin()[3],
			Right:  t.BoxModelMixin.Margin()[1],
			Top:    t.BoxModelMixin.Margin()[0],
			Bottom: t.BoxModelMixin.Margin()[2],
		},
		// Text typically doesn't have a border
		Border: layout.Border{Style: layout.BorderNone},
	}
}
