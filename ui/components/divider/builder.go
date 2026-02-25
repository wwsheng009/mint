package divider

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Builder - Fluent API for creating Divider VNode
// =============================================================================

// Builder provides a fluent API for building Divider VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Divider builder.
func NewBuilder() *Builder {
	return &Builder{
		node: New(),
	}
}

// Label sets the centered label text.
func (b *Builder) Label(label string) *Builder {
	b.node.SetLabel(label)
	return b
}

// Text sets the centered label text (alias for Label).
func (b *Builder) Text(text string) *Builder {
	return b.Label(text)
}

// Key sets the key for diffing.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// Style sets the visual style.
func (b *Builder) Style(s style.Style) *Builder {
	b.node.SetStyleProps(s)
	return b
}

// DividerStyle sets the line style.
func (b *Builder) DividerStyle(s Style) *Builder {
	b.node.SetDividerStyle(s)
	return b
}

// Thickness sets the line thickness.
func (b *Builder) Thickness(thickness int) *Builder {
	b.node.SetThickness(thickness)
	return b
}

// Orientation sets the divider direction.
func (b *Builder) Orientation(o Orientation) *Builder {
	b.node.SetOrientation(o)
	return b
}

// Horizontal sets orientation to horizontal.
func (b *Builder) Horizontal() *Builder {
	b.node.Horizontal()
	return b
}

// Vertical sets orientation to vertical.
func (b *Builder) Vertical() *Builder {
	b.node.Vertical()
	return b
}

// Solid sets style to solid line.
func (b *Builder) Solid() *Builder {
	b.node.Solid()
	return b
}

// Dashed sets style to dashed line.
func (b *Builder) Dashed() *Builder {
	b.node.Dashed()
	return b
}

// Dotted sets style to dotted line.
func (b *Builder) Dotted() *Builder {
	b.node.Dotted()
	return b
}

// Double sets style to double line.
func (b *Builder) Double() *Builder {
	b.node.Double()
	return b
}

// FillWidth sets whether to fill available width.
func (b *Builder) FillWidth(fill bool) *Builder {
	b.node.SetFillWidth(fill)
	return b
}

// FgColor sets the foreground color.
func (b *Builder) FgColor(c interface{}) *Builder {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s.FG = style.Color(colorStr)
		b.node.SetStyleProps(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s.FG = color
		b.node.SetStyleProps(s)
	}
	return b
}

// Build returns the Divider VNode.
func (b *Builder) Build() rtui.VNode {
	return b.node
}

// BuildInstance returns the Divider VNode as a ComponentInstance.
func (b *Builder) BuildInstance() rtui.ComponentInstance {
	return b.node.CreateInstance()
}

// =============================================================================
// Convenience Functions
// =============================================================================

// D creates a simple divider.
func D() rtui.VNode {
	return New()
}

// H creates a horizontal divider with optional label.
func H(label ...string) rtui.VNode {
	d := New().Horizontal()
	if len(label) > 0 {
		d.SetLabel(label[0])
	}
	return d
}

// V creates a vertical divider.
func V() rtui.VNode {
	return New().Vertical()
}

// WithLabel creates a divider with a centered label.
func WithLabel(label string) rtui.VNode {
	return New().SetLabel(label)
}

// Section creates a section divider with label.
func Section(title string) rtui.VNode {
	return New().SetLabel(" " + title + " ").Double()
}

// =============================================================================
// Backward Compatibility - Aliases for old API (basic package)
// =============================================================================

// Divider creates a new Divider VNode (for backward compatibility).
// This matches the old basic.Divider() API.
func Divider() *VNode {
	return New()
}

// NewDivider creates a new Divider VNode (alias for New, for backward compatibility).
// This matches the old basic.NewDivider() API.
func NewDivider() *VNode {
	return New()
}

// DividerBuilder is an alias for Builder (for backward compatibility).
// This matches the old basic.DividerBuilder type.
type DividerBuilder = Builder

// DividerStyle type and constants (for backward compatibility with basic package)
type DividerStyle = Style

const (
	DividerSolid   = StyleSolid
	DividerDashed  = StyleDashed
	DividerDotted  = StyleDotted
)
