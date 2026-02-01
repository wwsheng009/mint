package data

import (
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// VirtualListVNode represents a virtualized list component
// Only renders visible items for better performance with large datasets
type VirtualListVNode struct {
	*ui.ElementVNode
	items         []interface{}           // All items in the list
	itemCount     int                     // Number of items (for lazy loading)
	renderItem    func(interface{}) ui.VNode // Function to render each item
	itemHeight    int                     // Height of each item (0 = variable)
	visibleCount  int                     // Number of visible items
	scrollOffset  int                     // Current scroll offset
	height        int                     // Total height of the list
	allowScroll   bool                    // Whether scrolling is enabled
	onScrollEnd   func()                  // Callback when scrolled to end
	onItemSelect  func(int)               // Callback when item is selected
	selectedIndex int                     // Currently selected item
	primaryKey    func(interface{}) string // Function to get stable key for item
}

// NewVirtualList creates a new virtual list
func NewVirtualList() *VirtualListVNode {
	return &VirtualListVNode{
		ElementVNode:   ui.NewElement("virtuallist"),
		items:          make([]interface{}, 0),
		itemCount:      0,
		renderItem:     nil,
		itemHeight:     1,
		visibleCount:   10,
		scrollOffset:   0,
		height:         10,
		allowScroll:    true,
		onScrollEnd:    nil,
		onItemSelect:   nil,
		selectedIndex:  -1,
		primaryKey:     nil,
	}
}

// VirtualList creates a new virtual list node
func VirtualList(items []interface{}, renderItem func(interface{}) ui.VNode) ui.VNode {
	return &VirtualListVNode{
		ElementVNode:  ui.NewElement("virtuallist"),
		items:        items,
		itemCount:    len(items),
		renderItem:   renderItem,
		itemHeight:   1,
		visibleCount: 10,
		scrollOffset: 0,
		height:       10,
		allowScroll:  true,
		selectedIndex: -1,
	}
}

// Builder pattern
type VirtualListBuilderType struct {
	node *VirtualListVNode
}

// VirtualListBuilder creates a new virtual list builder
func VirtualListBuilder() *VirtualListBuilderType {
	return &VirtualListBuilderType{node: NewVirtualList()}
}

// Items sets the list items
func (b *VirtualListBuilderType) Items(items []interface{}) *VirtualListBuilderType {
	b.node.items = items
	b.node.itemCount = len(items)
	return b
}

// ItemCount sets the item count (for lazy loading)
func (b *VirtualListBuilderType) ItemCount(count int) *VirtualListBuilderType {
	b.node.itemCount = count
	return b
}

// RenderItem sets the render function for each item
func (b *VirtualListBuilderType) RenderItem(fn func(interface{}) ui.VNode) *VirtualListBuilderType {
	b.node.renderItem = fn
	return b
}

// ItemHeight sets the height of each item
func (b *VirtualListBuilderType) ItemHeight(height int) *VirtualListBuilderType {
	b.node.itemHeight = height
	return b
}

// VisibleCount sets the number of visible items
func (b *VirtualListBuilderType) VisibleCount(count int) *VirtualListBuilderType {
	b.node.visibleCount = count
	return b
}

// Height sets the total height of the list
func (b *VirtualListBuilderType) Height(height int) *VirtualListBuilderType {
	b.node.height = height
	return b
}

// ScrollOffset sets the initial scroll offset
func (b *VirtualListBuilderType) ScrollOffset(offset int) *VirtualListBuilderType {
	b.node.scrollOffset = offset
	return b
}

// AllowScroll enables/disables scrolling
func (b *VirtualListBuilderType) AllowScroll(allow bool) *VirtualListBuilderType {
	b.node.allowScroll = allow
	return b
}

// OnScrollEnd sets callback for when list is scrolled to end
func (b *VirtualListBuilderType) OnScrollEnd(fn func()) *VirtualListBuilderType {
	b.node.onScrollEnd = fn
	return b
}

// OnItemSelect sets callback for when an item is selected
func (b *VirtualListBuilderType) OnItemSelect(fn func(int)) *VirtualListBuilderType {
	b.node.onItemSelect = fn
	return b
}

// SelectedIndex sets the initially selected item
func (b *VirtualListBuilderType) SelectedIndex(index int) *VirtualListBuilderType {
	b.node.selectedIndex = index
	return b
}

// PrimaryKey sets the function to get stable key for each item
func (b *VirtualListBuilderType) PrimaryKey(fn func(interface{}) string) *VirtualListBuilderType {
	b.node.primaryKey = fn
	return b
}

// Width sets the width
func (b *VirtualListBuilderType) Width(w int) *VirtualListBuilderType {
	b.node.SetProp("width", w)
	return b
}

// Style sets the visual style
func (b *VirtualListBuilderType) Style(s style.Style) *VirtualListBuilderType {
	b.node.SetStyle(s)
	return b
}

// FgColor sets the foreground color
func (b *VirtualListBuilderType) FgColor(c interface{}) *VirtualListBuilderType {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s.FG = style.Color(colorStr)
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s.FG = color
		b.node.SetStyle(s)
	}
	return b
}

// BgColor sets the background color
func (b *VirtualListBuilderType) BgColor(c interface{}) *VirtualListBuilderType {
	if colorStr, ok := c.(string); ok {
		s := b.node.Style()
		s.BG = style.Color(colorStr)
		b.node.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.node.Style()
		s.BG = color
		b.node.SetStyle(s)
	}
	return b
}

// Key sets the key for diffing
func (b *VirtualListBuilderType) Key(key string) *VirtualListBuilderType {
	b.node.SetKey(key)
	return b
}

// Build returns the virtual list ui.VNode
func (b *VirtualListBuilderType) Build() ui.VNode {
	return b.node
}

// Getters
func (v *VirtualListVNode) Items() []interface{}          { return v.items }
func (v *VirtualListVNode) ItemCount() int                 { return v.itemCount }
func (v *VirtualListVNode) RenderItem() func(interface{}) ui.VNode { return v.renderItem }
func (v *VirtualListVNode) ItemHeight() int               { return v.itemHeight }
func (v *VirtualListVNode) VisibleCount() int             { return v.visibleCount }
func (v *VirtualListVNode) ScrollOffset() int             { return v.scrollOffset }
func (v *VirtualListVNode) ListHeight() int               { return v.height }
func (v *VirtualListVNode) AllowScroll() bool             { return v.allowScroll }
func (v *VirtualListVNode) OnScrollEnd() func()           { return v.onScrollEnd }
func (v *VirtualListVNode) OnItemSelect() func(int)       { return v.onItemSelect }
func (v *VirtualListVNode) SelectedIndex() int            { return v.selectedIndex }
func (v *VirtualListVNode) PrimaryKeyFunc() func(interface{}) string { return v.primaryKey }

// Setters
func (v *VirtualListVNode) SetItems(items []interface{})       { v.items = items }
func (v *VirtualListVNode) SetItemCount(count int)             { v.itemCount = count }
func (v *VirtualListVNode) SetRenderItem(fn func(interface{}) ui.VNode) { v.renderItem = fn }
func (v *VirtualListVNode) SetItemHeight(h int)                { v.itemHeight = h }
func (v *VirtualListVNode) SetVisibleCount(count int)           { v.visibleCount = count }
func (v *VirtualListVNode) SetScrollOffset(offset int)          { v.scrollOffset = offset }
func (v *VirtualListVNode) SetListHeight(h int)                 { v.height = h }
func (v *VirtualListVNode) SetAllowScroll(allow bool)           { v.allowScroll = allow }
func (v *VirtualListVNode) SetOnScrollEnd(fn func())            { v.onScrollEnd = fn }
func (v *VirtualListVNode) SetOnItemSelect(fn func(int))        { v.onItemSelect = fn }
func (v *VirtualListVNode) SetSelectedIndex(index int)         { v.selectedIndex = index }
func (v *VirtualListVNode) SetPrimaryKey(fn func(interface{}) string) { v.primaryKey = fn }

// ScrollBy scrolls the list by the given amount
func (v *VirtualListVNode) ScrollBy(delta int) int {
	newOffset := v.scrollOffset + delta
	if newOffset < 0 {
		newOffset = 0
	}
	maxOffset := v.itemCount - v.visibleCount
	if maxOffset < 0 {
		maxOffset = 0
	}
	if newOffset > maxOffset {
		newOffset = maxOffset
	}
	v.scrollOffset = newOffset
	return newOffset
}

// ScrollTo scrolls to a specific position
func (v *VirtualListVNode) ScrollTo(offset int) int {
	if offset < 0 {
		offset = 0
	}
	maxOffset := v.itemCount - v.visibleCount
	if maxOffset < 0 {
		maxOffset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}
	v.scrollOffset = offset
	return offset
}

// ScrollToItem scrolls to make a specific item visible
func (v *VirtualListVNode) ScrollToItem(index int) int {
	if index < 0 {
		index = 0
	}
	if index >= v.itemCount {
		index = v.itemCount - 1
	}
	// Position the item in the middle of the visible area
	targetOffset := index - v.visibleCount/2
	return v.ScrollTo(targetOffset)
}

// GetVisibleRange returns the start and end index of visible items
func (v *VirtualListVNode) GetVisibleRange() (start, end int) {
	start = v.scrollOffset
	end = start + v.visibleCount
	if end > v.itemCount {
		end = v.itemCount
	}
	return start, end
}

// IsItemAtEnd checks if scrolled to the end
func (v *VirtualListVNode) IsItemAtEnd() bool {
	return v.scrollOffset >= v.itemCount-v.visibleCount || v.itemCount <= v.visibleCount
}

// GetItem returns the item at the given index
func (v *VirtualListVNode) GetItem(index int) interface{} {
	if index >= 0 && index < len(v.items) {
		return v.items[index]
	}
	return nil
}
