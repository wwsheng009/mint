package panel

import (
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for constructing Panel VNodes.
type Builder struct {
	vnode *VNode
}

// NewBuilder creates a new Panel builder.
func NewBuilder() *Builder {
	return &Builder{
		vnode: New(),
	}
}

// Key sets the component key.
func (b *Builder) Key(key string) *Builder {
	b.vnode.key = key
	return b
}

// SetID sets the business identifier for positioning and Portal anchoring.
// This is separate from Key() which is used for list diffing.
func (b *Builder) SetID(id string) *Builder {
	b.vnode.SetID(id)
	return b
}

// Title sets the panel title.
func (b *Builder) Title(title string) *Builder {
	b.vnode.SetTitle(title)
	return b
}

// Header sets the header component.
func (b *Builder) Header(header rtui.VNode) *Builder {
	b.vnode.SetHeader(header)
	return b
}

// Content sets the main content component.
func (b *Builder) Content(content rtui.VNode) *Builder {
	b.vnode.SetContent(content)
	return b
}

// Child sets the main content (alias for Content).
func (b *Builder) Child(child rtui.VNode) *Builder {
	return b.Content(child)
}

// Footer sets the footer component.
func (b *Builder) Footer(footer rtui.VNode) *Builder {
	b.vnode.SetFooter(footer)
	return b
}

// Width sets the width.
func (b *Builder) Width(w int) *Builder {
	b.vnode.SetWidth(w)
	return b
}

// Height sets the height.
func (b *Builder) Height(h int) *Builder {
	b.vnode.SetHeight(h)
	return b
}

// Size sets both width and height.
func (b *Builder) Size(w, h int) *Builder {
	return b.Width(w).Height(h)
}

// Flex sets the flex factor.
func (b *Builder) Flex(f int) *Builder {
	b.vnode.SetFlex(f)
	return b
}

// Padding sets the inner padding.
func (b *Builder) Padding(p int) *Builder {
	b.vnode.SetPadding(p)
	return b
}

// Style sets the style.
func (b *Builder) Style(s style.Style) *Builder {
	b.vnode.SetStyle(s)
	return b
}

// BorderStyle sets the border style.
func (b *Builder) BorderStyle(s layout.BorderStyle) *Builder {
	b.vnode.SetBorderStyle(s)
	return b
}

// BorderColor sets the border color.
func (b *Builder) BorderColor(c style.Color) *Builder {
	b.vnode.SetBorderColor(c)
	return b
}

// Color sets the border color (alias for BorderColor).
func (b *Builder) Color(c style.Color) *Builder {
	return b.BorderColor(c)
}

// BorderLabel sets the border label.
func (b *Builder) BorderLabel(l string) *Builder {
	b.vnode.SetBorderLabel(l)
	return b
}

// Label sets the border label (alias for BorderLabel).
func (b *Builder) Label(l string) *Builder {
	return b.BorderLabel(l)
}

// Rounded sets rounded border style.
func (b *Builder) Rounded() *Builder {
	b.vnode.Rounded()
	return b
}

// Double sets double border style.
func (b *Builder) Double() *Builder {
	b.vnode.Double()
	return b
}

// Single sets single border style.
func (b *Builder) Single() *Builder {
	b.vnode.Single()
	return b
}

// NoBorder removes the border.
func (b *Builder) NoBorder() *Builder {
	b.vnode.NoBorder()
	return b
}

// Build returns the configured VNode.
func (b *Builder) Build() rtui.VNode {
	return b.vnode
}

// BuildVNode returns the configured *VNode.
func (b *Builder) BuildVNode() *VNode {
	return b.vnode
}

// =============================================================================
// Convenience Functions
// =============================================================================

// Of creates a Panel with the given content.
func Of(content rtui.VNode) rtui.VNode {
	return NewBuilder().Content(content).Build()
}

// OfSize creates a Panel with explicit dimensions.
func OfSize(content rtui.VNode, width, height int) rtui.VNode {
	return NewBuilder().Content(content).Width(width).Height(height).Build()
}

// Titled creates a Panel with a title.
func Titled(title string, content rtui.VNode) rtui.VNode {
	return NewBuilder().Title(title).Content(content).Rounded().Build()
}

// Bordered creates a Panel with border.
func Bordered(content rtui.VNode, width, height int) rtui.VNode {
	return NewBuilder().Content(content).Width(width).Height(height).Rounded().Build()
}

// =============================================================================
// Fluent Global Functions
// =============================================================================

// Panel creates a new Panel builder.
func Panel() *Builder {
	return NewBuilder()
}

// WithTitle creates a titled Panel builder.
func WithTitle(title string) *Builder {
	return NewBuilder().Title(title).Rounded()
}

// HeaderFooter creates a Panel with header and footer.
func HeaderFooter(header, content, footer rtui.VNode) rtui.VNode {
	return NewBuilder().Header(header).Content(content).Footer(footer).Rounded().Build()
}
