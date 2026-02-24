package tabs

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for constructing Tabs VNodes.
type Builder struct {
	vnode *VNode
}

// NewBuilder creates a new Tabs builder.
func NewBuilder() *Builder {
	return &Builder{
		vnode: New(),
	}
}

// Key sets the component key.
func (b *Builder) Key(key string) *Builder {
	b.vnode.SetKey(key)
	return b
}

// Tabs sets the tab items.
func (b *Builder) Tabs(tabs []TabItem) *Builder {
	b.vnode.SetTabs(tabs)
	return b
}

// AddTab adds a single tab.
func (b *Builder) AddTab(id, label string) *Builder {
	b.vnode.AddTab(id, label)
	return b
}

// AddTabWithOptions adds a tab with disabled option.
func (b *Builder) AddTabWithOptions(id, label string, disabled bool) *Builder {
	b.vnode.AddTabWithOptions(id, label, disabled)
	return b
}

// Position sets the tab position.
func (b *Builder) Position(pos TabPosition) *Builder {
	b.vnode.SetPosition(pos)
	return b
}

// WrapTabs enables/disables tab wrapping.
func (b *Builder) WrapTabs(wrap bool) *Builder {
	b.vnode.SetWrapTabs(wrap)
	return b
}

// TabGap sets the gap between tabs when wrapping.
func (b *Builder) TabGap(gap int) *Builder {
	b.vnode.SetTabGap(gap)
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

// Flex sets the flex factor.
func (b *Builder) Flex(f int) *Builder {
	b.vnode.SetFlex(f)
	return b
}

// Size sets both width and height.
func (b *Builder) Size(w, h int) *Builder {
	b.vnode.Size(w, h)
	return b
}

// Style sets the tab style.
func (b *Builder) Style(s style.Style) *Builder {
	b.vnode.SetTabStyle(s)
	return b
}

// ActiveTabStyle sets the active tab style.
func (b *Builder) ActiveTabStyle(s style.Style) *Builder {
	b.vnode.SetActiveTabStyle(s)
	return b
}

// OnChange sets the change intent.
func (b *Builder) OnChange(intent intent.Intent) *Builder {
	b.vnode.SetIntent(intent)
	return b
}

// Tab style convenience methods for position
func (b *Builder) Top() *Builder    { b.vnode.Top(); return b }
func (b *Builder) Bottom() *Builder { b.vnode.Bottom(); return b }
func (b *Builder) Left() *Builder   { b.vnode.Left(); return b }
func (b *Builder) Right() *Builder  { b.vnode.Right(); return b }

// Build returns the configured VNode.
func (b *Builder) Build() rtui.VNode {
	return b.vnode
}

// BuildVNode returns the configured *VNode.
func (b *Builder) BuildVNode() *VNode {
	return b.vnode
}

// BuildInstance directly creates an Instance.
func (b *Builder) BuildInstance() *Instance {
	return NewInstance(b.vnode.Props())
}

// =============================================================================
// Convenience Functions
// =============================================================================

// Of creates a Tabs component with the given tabs.
func Of(tabs []TabItem) rtui.VNode {
	return NewBuilder().Tabs(tabs).Build()
}

// OfWidth creates a Tabs component with explicit width.
func OfWidth(tabs []TabItem, width int) rtui.VNode {
	return NewBuilder().Tabs(tabs).Width(width).Build()
}

// OfSize creates a Tabs component with explicit dimensions.
func OfSize(tabs []TabItem, width, height int) rtui.VNode {
	return NewBuilder().Tabs(tabs).Size(width, height).Build()
}

// =============================================================================
// Fluent Global Functions
// =============================================================================

// Tabs creates a new Tabs builder.
func Tabs() *Builder {
	return NewBuilder()
}

// WithPosition creates a Tabs builder with position set.
func WithPosition(pos TabPosition) *Builder {
	return NewBuilder().Position(pos)
}

// TopPosition creates a Tabs builder with top position.
func TopPosition() *Builder {
	return NewBuilder().Top()
}

// BottomPosition creates a Tabs builder with bottom position.
func BottomPosition() *Builder {
	return NewBuilder().Bottom()
}

// LeftPosition creates a Tabs builder with left position.
func LeftPosition() *Builder {
	return NewBuilder().Left()
}

// RightPosition creates a Tabs builder with right position.
func RightPosition() *Builder {
	return NewBuilder().Right()
}
