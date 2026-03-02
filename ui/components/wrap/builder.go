package wrap

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Builder - Fluent API for creating Wrap VNode
// =============================================================================

// Builder provides a fluent API for building Wrap VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Wrap builder.
func NewBuilder() *Builder {
	return &Builder{
		node: New(),
	}
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

// Gap sets spacing between items in the same row.
func (b *Builder) Gap(n int) *Builder {
	b.node.SetGap(n)
	return b
}

// RowGap sets spacing between rows (0 = use gap value).
func (b *Builder) RowGap(n int) *Builder {
	b.node.SetRowGap(n)
	return b
}

// Align sets main-axis alignment for each row.
func (b *Builder) Align(a Align) *Builder {
	b.node.SetAlign(a)
	return b
}

// Width sets container width for wrap calculation.
func (b *Builder) Width(w int) *Builder {
	b.node.SetWidth(w)
	return b
}

// Padding sets the padding (top, right, bottom, left).
func (b *Builder) Padding(top, right, bottom, left int) *Builder {
	b.node.SetPadding(top, right, bottom, left)
	return b
}

// FillWidth makes each row stretch to fill the container width.
func (b *Builder) FillWidth() *Builder {
	b.node.SetFillWidth(true)
	return b
}

// FillHeight makes the wrap container stretch to fill parent's height.
func (b *Builder) FillHeight() *Builder {
	b.node.SetFillHeight(true)
	return b
}

// Children sets the children.
func (b *Builder) Children(children ...rtui.VNode) *Builder {
	b.node.SetChildrenList(children)
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

// BgColor sets the background color.
func (b *Builder) BgColor(c interface{}) *Builder {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s.BG = style.Color(colorStr)
		b.node.SetStyleProps(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s.BG = color
		b.node.SetStyleProps(s)
	}
	return b
}

// Center sets alignment to center.
func (b *Builder) Center() *Builder {
	b.node.SetAlign(AlignCenter)
	return b
}

// End sets alignment to end.
func (b *Builder) End() *Builder {
	b.node.SetAlign(AlignEnd)
	return b
}

// Build returns the Wrap VNode.
func (b *Builder) Build() rtui.VNode {
	return b.node
}

// BuildInstance returns the Wrap VNode as a ComponentInstance.
func (b *Builder) BuildInstance() rtui.ComponentInstance {
	return b.node.CreateInstance()
}

// =============================================================================
// Convenience Functions
// =============================================================================

// W creates a new Wrap builder.
func W() *Builder {
	return NewBuilder()
}

// Wrap creates a new Wrap VNode directly with children.
func Wrap(children ...rtui.VNode) *VNode {
	return New().SetChildrenList(children)
}

// WrapWithWidth creates a Wrap with specified width and children.
func WrapWithWidth(width int, children ...rtui.VNode) rtui.VNode {
	return New().SetWidth(width).SetChildrenList(children)
}

// WrapWithGap creates a Wrap with specified gap and children.
func WrapWithGap(gap int, children ...rtui.VNode) rtui.VNode {
	return New().SetGap(gap).SetChildrenList(children)
}

// WrapConfig creates a Wrap with full configuration.
func WrapConfig(width, gap int, align Align, children ...rtui.VNode) rtui.VNode {
	return New().
		SetWidth(width).
		SetGap(gap).
		SetAlign(align).
		SetChildrenList(children)
}
