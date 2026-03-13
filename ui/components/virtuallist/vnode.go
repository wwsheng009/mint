// Package virtuallist provides Fiber-first VirtualList component.
package virtuallist

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Prop Keys
// =============================================================================

// Prop key constants — shared by VNode and Instance to avoid magic strings.
const (
	propAllowScroll = "allowScroll"
	propHeight = "height"
	propItemCount = "itemCount"
	propItemHeight = "itemHeight"
	propItems = "items"
	propKey = "key"
	propListStyle = "listStyle"
	propScrollOffset = "scrollOffset"
	propSelectedIndex = "selectedIndex"
	propSelectedStyle = "selectedStyle"
	propVisibleCount = "visibleCount"
	propWidth = "width"
)

// =============================================================================
// VNode - Pure Description (No State, No Closures, No Paint)
// =============================================================================

// VNode is the virtual list component description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
// For Fiber-first, item rendering is handled via callbacks.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === List Props ===
	items       []string // Simple string items (for serialization)
	itemCount   int      // Number of items (for lazy loading)
	itemHeight  int      // Height of each item
	visibleCount int      // Number of visible items
	height      int      // Total height of the list
	width       int      // Width of the list

	// === Style ===
	listStyle      style.Style
	selectedStyle  style.Style

	// === Constraints ===
	allowScroll bool

	// === Initial State ===
	scrollOffset    int
	selectedIndex  int
}

// Ensure VNode implements required interfaces
var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// New creates a new virtual list VNode.
func New() *VNode {
	return &VNode{
		ElementVNode:    rtui.NewElement("virtuallist"),
		items:          []string{},
		itemCount:      0,
		itemHeight:     1,
		visibleCount:   10,
		scrollOffset:   0,
		height:         10,
		width:          40,
		allowScroll:    true,
		selectedIndex:  -1,
		listStyle:      style.Style{},
		selectedStyle:  style.Style{}.Bold(true),
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

func (v *VNode) Key() string           { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode { v.key = key; return v }
func (v *VNode) Tag() string           { return "virtuallist" }
func (v *VNode) Type() rtui.VNodeType  { return rtui.VNodeElement }

func (v *VNode) Children() []rtui.VNode {
	// VirtualList is a leaf component - no children
	return []rtui.VNode{}
}

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	// VirtualList manages its own rendering
	return v
}

func (v *VNode) GetLayer() rtui.Layer   { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Style() style.Style    { return v.listStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode { v.listStyle = s; return v }

func (v *VNode) TextContent() string   { return "" }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propKey:            v.key,
		propItems:          v.items,
		propItemCount:      v.itemCount,
		propItemHeight:     v.itemHeight,
		propVisibleCount:   v.visibleCount,
		propHeight:         v.height,
		propWidth:          v.width,
		propAllowScroll:    v.allowScroll,
		propScrollOffset:    v.scrollOffset,
		propSelectedIndex:  v.selectedIndex,
		propListStyle:      v.listStyle,
		propSelectedStyle:  v.selectedStyle,
	}
}

func (v *VNode) SetProps(p rtui.Props) rtui.VNode {
	if val, ok := p[propKey].(string); ok {
		v.key = val
	}
	if val, ok := p[propItems].([]string); ok {
		v.items = val
	}
	if val, ok := p[propItemCount].(int); ok {
		v.itemCount = val
	}
	if val, ok := p[propItemHeight].(int); ok {
		v.itemHeight = val
	}
	if val, ok := p[propVisibleCount].(int); ok {
		v.visibleCount = val
	}
	if val, ok := p[propHeight].(int); ok {
		v.height = val
	}
	if val, ok := p[propWidth].(int); ok {
		v.width = val
	}
	if val, ok := p[propAllowScroll].(bool); ok {
		v.allowScroll = val
	}
	if val, ok := p[propScrollOffset].(int); ok {
		v.scrollOffset = val
	}
	if val, ok := p[propSelectedIndex].(int); ok {
		v.selectedIndex = val
	}
	if val, ok := p[propListStyle].(style.Style); ok {
		v.listStyle = val
	}
	if val, ok := p[propSelectedStyle].(style.Style); ok {
		v.selectedStyle = val
	}
	return v
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(rtui.Props{
		propKey:            v.key,
		propItems:          v.items,
		propItemCount:      v.itemCount,
		propItemHeight:     v.itemHeight,
		propVisibleCount:   v.visibleCount,
		propHeight:         v.height,
		propWidth:          v.width,
		propAllowScroll:    v.allowScroll,
		propScrollOffset:    v.scrollOffset,
		propSelectedIndex:  v.selectedIndex,
		propListStyle:      v.listStyle,
		propSelectedStyle:  v.selectedStyle,
	})
}

// =============================================================================
// Fluent Setters
// =============================================================================

func (v *VNode) SetItems(items []string) *VNode   { v.items = items; return v }
func (v *VNode) SetItemCount(count int) *VNode     { v.itemCount = count; return v }
func (v *VNode) SetItemHeight(h int) *VNode        { v.itemHeight = h; return v }
func (v *VNode) SetVisibleCount(count int) *VNode  { v.visibleCount = count; return v }
func (v *VNode) SetHeight(h int) *VNode           { v.height = h; return v }
func (v *VNode) SetWidth(w int) *VNode            { v.width = w; return v }
func (v *VNode) SetAllowScroll(allow bool) *VNode { v.allowScroll = allow; return v }
func (v *VNode) SetScrollOffset(offset int) *VNode   { v.scrollOffset = offset; return v }
func (v *VNode) SetSelectedIndex(index int) *VNode { v.selectedIndex = index; return v }

func (v *VNode) SetListStyle(s style.Style) *VNode     { v.listStyle = s; return v }
func (v *VNode) SetSelectedStyle(s style.Style) *VNode { v.selectedStyle = s; return v }

func (v *VNode) SetSize(w, h int) *VNode            { return v.SetWidth(w).SetHeight(h) }
func (v *VNode) SetViewport(w, visible int) *VNode  { return v.SetWidth(w).SetVisibleCount(visible) }

func (v *VNode) AddItem(item string) *VNode {
	v.items = append(v.items, item)
	v.itemCount = len(v.items)
	return v
}

// =============================================================================
// Accessors
// =============================================================================

func (v *VNode) Items() []string        { return v.items }
func (v *VNode) ItemCount() int           { return v.itemCount }
func (v *VNode) ItemHeight() int          { return v.itemHeight }
func (v *VNode) VisibleCount() int         { return v.visibleCount }
func (v *VNode) Height() int              { return v.height }
func (v *VNode) Width() int               { return v.width }
func (v *VNode) AllowScroll() bool         { return v.allowScroll }
func (v *VNode) ScrollOffset() int         { return v.scrollOffset }
func (v *VNode) SelectedIndex() int        { return v.selectedIndex }
func (v *VNode) ListStyle() style.Style     { return v.listStyle }
func (v *VNode) SelectedStyle() style.Style { return v.selectedStyle }
