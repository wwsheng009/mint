package scrollview

import (
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder creates ScrollView VNodes with a fluent API.
type Builder struct {
	vnode *VNode
}

// NewBuilder creates a new ScrollView builder.
func NewBuilder() *Builder {
	return &Builder{
		vnode: New(),
	}
}

// Child sets the child content.
func (b *Builder) Child(child rtui.VNode) *Builder {
	b.vnode.child = child
	return b
}

// Width sets the viewport width.
func (b *Builder) Width(width int) *Builder {
	b.vnode.width = width
	return b
}

// Height sets the viewport height.
func (b *Builder) Height(height int) *Builder {
	b.vnode.height = height
	return b
}

// ScrollOffset sets the initial scroll position in uncontrolled mode.
func (b *Builder) ScrollOffset(offset int) *Builder {
	b.vnode.SetInitialScrollOffset(offset)
	return b
}

// ScrollOffsetControlled sets the scroll position in controlled mode.
func (b *Builder) ScrollOffsetControlled(offset int) *Builder {
	b.vnode.SetScrollOffsetControlled(offset)
	return b
}

// InitialScrollOffset sets the initial scroll position in uncontrolled mode.
func (b *Builder) InitialScrollOffset(offset int) *Builder {
	b.vnode.SetInitialScrollOffset(offset)
	return b
}

// ShowBorder enables border display.
func (b *Builder) ShowBorder(show bool) *Builder {
	b.vnode.showBorder = show
	return b
}

// ShowIndicator enables scroll indicator display.
func (b *Builder) ShowIndicator(show bool) *Builder {
	b.vnode.showIndicator = show
	return b
}

// Border adds a border to the scrollview.
func (b *Builder) Border() *Builder {
	b.vnode.showBorder = true
	return b
}

// NoBorder removes the border.
func (b *Builder) NoBorder() *Builder {
	b.vnode.showBorder = false
	return b
}

// Style sets the visual style.
func (b *Builder) Style(s style.Style) *Builder {
	b.vnode.style = s
	return b
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

// Of creates a ScrollView with the given child content.
func Of(child rtui.VNode) rtui.VNode {
	return NewBuilder().Child(child).Build()
}

// OfSize creates a ScrollView with explicit dimensions.
func OfSize(child rtui.VNode, width, height int) rtui.VNode {
	return NewBuilder().Child(child).Width(width).Height(height).Build()
}

// Bordered creates a ScrollView with border.
func Bordered(child rtui.VNode, width, height int) rtui.VNode {
	return NewBuilder().Child(child).Width(width).Height(height).Border().Build()
}

// =============================================================================
// Fluent Global Functions (matching existing patterns)
// =============================================================================

// ScrollView creates a new ScrollView builder.
func ScrollView() *Builder {
	return NewBuilder()
}

// Scroll creates a ScrollView with child content (convenience).
func Scroll(child rtui.VNode) rtui.VNode {
	return Of(child)
}

// ScrollSize creates a ScrollView with explicit size (convenience).
func ScrollSize(child rtui.VNode, width, height int) rtui.VNode {
	return OfSize(child, width, height)
}

// =============================================================================
// VNode SetChild Wrapper
// =============================================================================

// SetChild sets the child content (implements VNode interface pattern).
func (v *VNode) SetChildNode(child rtui.VNode) *VNode {
	v.child = child
	return v
}

// =============================================================================
// Layout Support
// =============================================================================

// GetScrollInfo returns scroll information for layout integration.
func (v *VNode) GetScrollInfo() (offset, totalLines, viewportH int) {
	// This will be populated from Instance during layout
	return v.scrollOffset, 0, v.height
}

// GetContentSize returns the content size for layout.
func (v *VNode) GetContentSize() layout.Size {
	// Default size, will be updated during Measure
	return layout.Size{
		Width:  v.width,
		Height: v.height,
	}
}
