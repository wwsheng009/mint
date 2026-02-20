package stack

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Builder - Fluent API for creating Stack VNode
// =============================================================================

// Builder provides a fluent API for building Stack VNodes.
type Builder struct {
	node *VNode
}

// NewBuilder creates a new Stack builder.
func NewBuilder(dir Direction) *Builder {
	return &Builder{
		node: New(dir),
	}
}

// Key sets the key for diffing.
func (b *Builder) Key(key string) *Builder {
	b.node.SetKey(key)
	return b
}

// Direction sets the layout direction.
func (b *Builder) Direction(dir Direction) *Builder {
	b.node.SetDirection(dir)
	return b
}

// Align sets the main axis alignment.
func (b *Builder) Align(a Align) *Builder {
	b.node.SetAlign(a)
	return b
}

// CrossAlign sets the cross axis alignment.
func (b *Builder) CrossAlign(a Align) *Builder {
	b.node.SetCrossAlign(a)
	return b
}

// Gap sets the spacing between children.
func (b *Builder) Gap(gap int) *Builder {
	b.node.SetGap(gap)
	return b
}

// Padding sets the padding (top, right, bottom, left).
func (b *Builder) Padding(top, right, bottom, left int) *Builder {
	b.node.SetPadding(top, right, bottom, left)
	return b
}

// Stretch makes children stretch to fill cross axis.
func (b *Builder) Stretch() *Builder {
	b.node.Stretch()
	return b
}

// Width sets the explicit width.
func (b *Builder) Width(w int) *Builder {
	b.node.SetWidth(w)
	return b
}

// Height sets the explicit height.
func (b *Builder) Height(h int) *Builder {
	b.node.SetHeight(h)
	return b
}

// Flex sets the flex factor.
func (b *Builder) Flex(f int) *Builder {
	b.node.SetFlex(f)
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

// Center centers children on main axis.
func (b *Builder) Center() *Builder {
	b.node.Center()
	return b
}

// CenterCross centers children on cross axis.
func (b *Builder) CenterCross() *Builder {
	b.node.CenterCross()
	return b
}

// Build returns the Stack VNode.
func (b *Builder) Build() rtui.VNode {
	return b.node
}

// BuildInstance returns the Stack VNode as a ComponentInstance.
func (b *Builder) BuildInstance() rtui.ComponentInstance {
	return b.node.CreateInstance()
}

// =============================================================================
// HStack Builder
// =============================================================================

// HStackBuilder provides a fluent API for building HStack VNodes.
type HStackBuilder struct {
	*Builder
}

// NewHStackBuilder creates a new HStack builder.
func NewHStackBuilder() *HStackBuilder {
	return &HStackBuilder{
		Builder: NewBuilder(Row),
	}
}

// Build returns the HStack VNode.
func (b *HStackBuilder) Build() *VNode {
	return b.node
}

// =============================================================================
// VStack Builder
// =============================================================================

// VStackBuilder provides a fluent API for building VStack VNodes.
type VStackBuilder struct {
	*Builder
}

// NewVStackBuilder creates a new VStack builder.
func NewVStackBuilder() *VStackBuilder {
	return &VStackBuilder{
		Builder: NewBuilder(Column),
	}
}

// Build returns the VStack VNode.
func (b *VStackBuilder) Build() *VNode {
	return b.node
}

// =============================================================================
// Convenience Functions
// =============================================================================

// H creates a horizontal stack (HStack).
func H(children ...rtui.VNode) *VNode {
	return NewHStack().SetChildrenList(children)
}

// V creates a vertical stack (VStack).
func V(children ...rtui.VNode) *VNode {
	return NewVStack().SetChildrenList(children)
}

// HBox creates an HStack with children.
func HBox(children ...rtui.VNode) rtui.VNode {
	return H(children...)
}

// VBox creates a VStack with children.
func VBox(children ...rtui.VNode) rtui.VNode {
	return V(children...)
}

// RowStack creates a row layout with gap.
func RowStack(gap int, children ...rtui.VNode) rtui.VNode {
	return NewHStack().SetGap(gap).SetChildrenList(children)
}

// ColStack creates a column layout with gap.
func ColStack(gap int, children ...rtui.VNode) rtui.VNode {
	return NewVStack().SetGap(gap).SetChildrenList(children)
}

// Spacer creates a flexible spacer.
func Spacer(flex int) *spacerVNode {
	return &spacerVNode{flex: flex}
}

// spacerVNode is a simple spacer component.
type spacerVNode struct {
	flex int
}

func (s *spacerVNode) Type() rtui.VNodeType                           { return rtui.VNodeElement }
func (s *spacerVNode) Key() string                                    { return "" }
func (s *spacerVNode) SetKey(string) rtui.VNode                       { return s }
func (s *spacerVNode) Tag() string                                    { return "spacer" }
func (s *spacerVNode) Style() style.Style                             { return style.Style{} }
func (s *spacerVNode) SetStyle(style.Style) rtui.VNode                { return s }
func (s *spacerVNode) Children() []rtui.VNode                         { return nil }
func (s *spacerVNode) SetChildren([]rtui.VNode) rtui.VNode            { return s }
func (s *spacerVNode) Props() rtui.Props                              { return rtui.Props{"flex": s.flex} }
func (s *spacerVNode) SetProps(rtui.Props) rtui.VNode                 { return s }
func (s *spacerVNode) GetLayer() rtui.Layer                           { return rtui.LayerBase }
func (s *spacerVNode) SetLayer(rtui.Layer) rtui.VNode                 { return s }
func (s *spacerVNode) CreateInstance() rtui.ComponentInstance         { return nil }
func (s *spacerVNode) GetLayoutInfo() rtui.LayoutInfo {
	return rtui.LayoutInfo{Flex: s.flex}
}
