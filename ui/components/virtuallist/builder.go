package virtuallist

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Builder - Fluent API
// =============================================================================

// Builder provides a fluent API for constructing VirtualList VNodes.
type Builder struct {
	vnode *VNode
}

// NewBuilder creates a new VirtualList builder.
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

// Items sets the list items.
func (b *Builder) Items(items []string) *Builder {
	b.vnode.SetItems(items)
	return b
}

// ItemCount sets the item count.
func (b *Builder) ItemCount(count int) *Builder {
	b.vnode.SetItemCount(count)
	return b
}

// ItemHeight sets the height of each item.
func (b *Builder) ItemHeight(height int) *Builder {
	b.vnode.SetItemHeight(height)
	return b
}

// VisibleCount sets the number of visible items.
func (b *Builder) VisibleCount(count int) *Builder {
	b.vnode.SetVisibleCount(count)
	return b
}

// Height sets the total height of the list.
func (b *Builder) Height(height int) *Builder {
	b.vnode.SetHeight(height)
	return b
}

// Width sets the total width of the list.
func (b *Builder) Width(width int) *Builder {
	b.vnode.SetWidth(width)
	return b
}

// Size sets both width and height.
func (b *Builder) Size(w, h int) *Builder {
	b.vnode.SetWidth(w)
	b.vnode.SetHeight(h)
	return b
}

// ScrollOffset sets the current scroll offset.
func (b *Builder) ScrollOffset(offset int) *Builder {
	b.vnode.SetScrollOffset(offset)
	return b
}

// SelectedIndex sets the currently selected item index.
func (b *Builder) SelectedIndex(index int) *Builder {
	b.vnode.SetSelectedIndex(index)
	return b
}

// AllowScroll enables/disables scrolling.
func (b *Builder) AllowScroll(allow bool) *Builder {
	b.vnode.SetAllowScroll(allow)
	return b
}

// ListStyle sets the style for list items.
func (b *Builder) ListStyle(s style.Style) *Builder {
	b.vnode.SetListStyle(s)
	return b
}

// SelectedStyle sets the style for the selected item.
func (b *Builder) SelectedStyle(s style.Style) *Builder {
	b.vnode.SetSelectedStyle(s)
	return b
}

// FgColor sets the foreground color for list items.
func (b *Builder) FgColor(c style.Color) *Builder {
	s := b.vnode.listStyle
	s.FG = c
	b.vnode.SetListStyle(s)
	return b
}

// BgColor sets the background color for list items.
func (b *Builder) BgColor(c style.Color) *Builder {
	s := b.vnode.listStyle
	s.BG = c
	b.vnode.SetListStyle(s)
	return b
}

// SelectedFgColor sets the foreground color for the selected item.
func (b *Builder) SelectedFgColor(c style.Color) *Builder {
	s := b.vnode.selectedStyle
	s.FG = c
	b.vnode.SetSelectedStyle(s)
	return b
}

// SelectedBgColor sets the background color for the selected item.
func (b *Builder) SelectedBgColor(c style.Color) *Builder {
	s := b.vnode.selectedStyle
	s.BG = c
	b.vnode.SetSelectedStyle(s)
	return b
}

// AddItem adds a single item to the list.
func (b *Builder) AddItem(item string) *Builder {
	b.vnode.AddItem(item)
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

// BuildInstance directly creates an Instance.
func (b *Builder) BuildInstance() *Instance {
	return NewInstance(b.vnode.Props())
}

// =============================================================================
// Convenience Functions
// =============================================================================

// Of creates a VirtualList with the given items.
func Of(items []string) rtui.VNode {
	return NewBuilder().Items(items).Build()
}

// OfSize creates a VirtualList with explicit dimensions.
func OfSize(items []string, width, height int) rtui.VNode {
	return NewBuilder().Items(items).Width(width).Height(height).Build()
}

// OfItems creates a VirtualList with custom item count and visible count.
func OfItems(items []string, itemCount, visibleCount int) rtui.VNode {
	return NewBuilder().Items(items).ItemCount(itemCount).VisibleCount(visibleCount).Build()
}

// =============================================================================
// Fluent Global Functions
// =============================================================================

// VirtualList creates a new VirtualList builder.
func VirtualList() *Builder {
	return NewBuilder()
}

// WithItems creates a builder with pre-configured items.
func WithItems(items []string) *Builder {
	return NewBuilder().Items(items)
}
