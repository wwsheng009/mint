package border

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Builder - Fluent API for creating Border VNode
// =============================================================================

// Builder provides a fluent API for building Border VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new border builder.
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

// Style sets the border style.
func (b *Builder) Style(s BorderStyle) *Builder {
	b.node.SetBorderStyle(s)
	return b
}

// Single sets border style to single line.
func (b *Builder) Single() *Builder {
	b.node.Single()
	return b
}

// Double sets border style to double line.
func (b *Builder) Double() *Builder {
	b.node.Double()
	return b
}

// Rounded sets border style to rounded corners.
func (b *Builder) Rounded() *Builder {
	b.node.Rounded()
	return b
}

// Dashed sets border style to dashed line.
func (b *Builder) Dashed() *Builder {
	b.node.Dashed()
	return b
}

// None removes the border.
func (b *Builder) None() *Builder {
	b.node.None()
	return b
}

// Color sets the border color.
func (b *Builder) Color(c string) *Builder {
	b.node.SetBorderColor(style.Color(c))
	return b
}

// Label sets the border label (displayed on top edge).
func (b *Builder) Label(label string) *Builder {
	b.node.SetBorderLabel(label)
	return b
}

// Width sets the content width (border adds borderWidth*2).
func (b *Builder) Width(w int) *Builder {
	b.node.SetWidth(w)
	return b
}

// Height sets the content height (border adds borderWidth*2).
func (b *Builder) Height(h int) *Builder {
	b.node.SetHeight(h)
	return b
}

// Flex sets the flex factor.
func (b *Builder) Flex(f int) *Builder {
	b.node.SetFlex(f)
	return b
}

// Child sets the content inside the border.
func (b *Builder) Child(child rtui.VNode) *Builder {
	b.node.SetChild(child)
	return b
}

// StyleObj sets the visual style.
func (b *Builder) StyleObj(s style.Style) *Builder {
	b.node.SetStyleProps(s)
	return b
}

// Build returns the Border VNode.
func (b *Builder) Build() rtui.VNode {
	return b.node
}

// BuildInstance returns the Border VNode as a ComponentInstance.
func (b *Builder) BuildInstance() rtui.ComponentInstance {
	return b.node.CreateInstance()
}

// =============================================================================
// Convenience Functions
// =============================================================================

// B creates a border container with default single-line style.
func B(child rtui.VNode) *VNode {
	return New().SetChild(child)
}

// Single creates a single-line border container.
func Single(child rtui.VNode) *VNode {
	return New().Single().SetChild(child)
}

// Double creates a double-line border container.
func Double(child rtui.VNode) *VNode {
	return New().Double().SetChild(child)
}

// Rounded creates a rounded-corners border container.
func Rounded(child rtui.VNode) *VNode {
	return New().Rounded().SetChild(child)
}

// Dashed creates a dashed-line border container.
func Dashed(child rtui.VNode) *VNode {
	return New().Dashed().SetChild(child)
}

// WithLabel creates a bordered container with a label on the top edge.
func WithLabel(label string, child rtui.VNode) *VNode {
	return New().Label(label).SetChild(child)
}

// WithColor creates a bordered container with a custom color.
func WithColor(color string, child rtui.VNode) *VNode {
	return New().Color(color).SetChild(child)
}

// =============================================================================
// Size Helpers
// =============================================================================

// Box calculates the total size including border.
// Returns (totalWidth, totalHeight) given content dimensions.
func Box(contentWidth, contentHeight int, s BorderStyle) (int, int) {
	bw := GetBorderWidth(s)
	return contentWidth + bw*2, contentHeight + bw*2
}

// Content calculates the available content size given total dimensions.
// Returns (contentWidth, contentHeight) given total dimensions.
func Content(totalWidth, totalHeight int, s BorderStyle) (int, int) {
	bw := GetBorderWidth(s)
	return totalWidth - bw*2, totalHeight - bw*2
}

// Padding returns the horizontal and vertical padding added by the border.
// Returns (horizontalPadding, verticalPadding).
func Padding(s BorderStyle) (int, int) {
	bw := GetBorderWidth(s)
	return bw * 2, bw * 2
}

// Offset returns the x,y offset for content inside the border.
func Offset(s BorderStyle) (int, int) {
	bw := GetBorderWidth(s)
	return bw, bw
}
