package layout

import (
	"fmt"

	"github.com/wwsheng009/mint/framework/action"
	ui "github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/runtime/style"
)

// Interface implementation assertions
var _ action.ActionTarget = (*VirtualList)(nil)
var _ action.ScrollableActionTarget = (*VirtualList)(nil)

// VirtualList is a virtualized list component that only renders visible items
// This is similar to react-window or react-virtualized in web development
//
// Virtual scrolling improves performance by only rendering items that are
// currently visible in the viewport, rather than the entire list.
//
// Usage:
//   virtualList := layout.NewVirtualList(itemCount, renderItemFunc).
//       ItemHeight(1).      // Height of each item in lines
//       ViewportHeight(20). // Number of visible lines
//       ScrollOffset(5).    // Current scroll position
//       Build()
type VirtualList struct {
	*ui.ElementVNode

	itemCount      int       // Total number of items
	itemHeight     int       // Height of each item (in lines)
	viewportHeight int       // Viewport height (in lines)
	scrollOffset   int       // Current scroll offset (in items)
	renderItem     func(int) ui.VNode // Function to render an item
}

// VirtualListBuilder builds virtualized lists
type VirtualListBuilder struct {
	node           *VirtualList
	itemCount      int
	itemHeight     int
	viewportHeight int
	scrollOffset   int
	renderItem     func(int) ui.VNode
	style          style.Style
}

// NewVirtualList creates a new virtual list builder
//
// Parameters:
//   - itemCount: Total number of items in the list
//   - renderItem: Function that renders an item given its index
//
// Example:
//   virtualList := layout.NewVirtualList(1000, func(i int) ui.VNode {
//       return ui.Text(fmt.Sprintf("Item %d", i))
//   }).
//       ItemHeight(1).
//       ViewportHeight(20).
//       Build()
func NewVirtualList(itemCount int, renderItem func(int) ui.VNode) *VirtualListBuilder {
	return &VirtualListBuilder{
		itemCount:      itemCount,
		itemHeight:     1, // Default: each item is 1 line tall
		viewportHeight: 20, // Default: show 20 items
		scrollOffset:   0,
		renderItem:     renderItem,
		node: &VirtualList{
			ElementVNode: ui.NewElement("virtual-list"),
		},
	}
}

// ItemHeight sets the height of each item in lines
func (b *VirtualListBuilder) ItemHeight(height int) *VirtualListBuilder {
	b.itemHeight = height
	return b
}

// ViewportHeight sets the viewport height in lines
func (b *VirtualListBuilder) ViewportHeight(height int) *VirtualListBuilder {
	b.viewportHeight = height
	return b
}

// ScrollOffset sets the current scroll position (in items)
func (b *VirtualListBuilder) ScrollOffset(offset int) *VirtualListBuilder {
	b.scrollOffset = offset
	return b
}

// Style sets the visual style for all items
func (b *VirtualListBuilder) Style(s style.Style) *VirtualListBuilder {
	b.style = s
	return b
}

// Key sets key for diffing
func (b *VirtualListBuilder) Key(key string) *VirtualListBuilder {
	b.node.SetKey(key)
	return b
}

// Build creates the virtual list VNode
func (b *VirtualListBuilder) Build() ui.VNode {
	// Clamp scroll offset to valid range
	maxOffset := b.itemCount - (b.viewportHeight / b.itemHeight)
	if maxOffset < 0 {
		maxOffset = 0
	}
	if b.scrollOffset < 0 {
		b.scrollOffset = 0
	}
	if b.scrollOffset > maxOffset {
		b.scrollOffset = maxOffset
	}

	// Store configuration in node
	b.node.itemCount = b.itemCount
	b.node.itemHeight = b.itemHeight
	b.node.viewportHeight = b.viewportHeight
	b.node.scrollOffset = b.scrollOffset
	b.node.renderItem = b.renderItem

	// Calculate which items are visible
	itemsPerPage := b.viewportHeight / b.itemHeight
	if itemsPerPage < 1 {
		itemsPerPage = 1
	}

	startIndex := b.scrollOffset
	endIndex := startIndex + itemsPerPage
	if endIndex > b.itemCount {
		endIndex = b.itemCount
	}

	// Build visible items
	var children []ui.VNode
	for i := startIndex; i < endIndex; i++ {
		item := b.renderItem(i)
		// Apply style if set
		if b.style.FG != "" || b.style.BG != "" {
			item.SetStyle(b.style)
		}
		children = append(children, item)
	}

	// Create VStack with visible items
	result := ui.VStackBuilder(children...).
		Height(b.viewportHeight).
		Build()

	return result
}

// =============================================================================
// VirtualList Instance Methods
// =============================================================================

// ScrollBy scrolls by the given delta (in items)
func (vl *VirtualList) ScrollBy(delta int) int {
	itemsPerPage := vl.viewportHeight / vl.itemHeight
	if itemsPerPage < 1 {
		itemsPerPage = 1
	}

	maxOffset := vl.itemCount - itemsPerPage
	if maxOffset < 0 {
		maxOffset = 0
	}

	newOffset := vl.scrollOffset + delta
	if newOffset < 0 {
		newOffset = 0
	}
	if newOffset > maxOffset {
		newOffset = maxOffset
	}

	vl.scrollOffset = newOffset
	return newOffset
}

// ScrollTo scrolls to an absolute position (in items)
func (vl *VirtualList) ScrollTo(offset int) int {
	itemsPerPage := vl.viewportHeight / vl.itemHeight
	if itemsPerPage < 1 {
		itemsPerPage = 1
	}

	maxOffset := vl.itemCount - itemsPerPage
	if maxOffset < 0 {
		maxOffset = 0
	}

	if offset < 0 {
		offset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}

	vl.scrollOffset = offset
	return offset
}

// ScrollTop scrolls to the top
func (vl *VirtualList) ScrollTop() {
	vl.scrollOffset = 0
}

// ScrollBottom scrolls to the bottom
func (vl *VirtualList) ScrollBottom() {
	itemsPerPage := vl.viewportHeight / vl.itemHeight
	if itemsPerPage < 1 {
		itemsPerPage = 1
	}

	maxOffset := vl.itemCount - itemsPerPage
	if maxOffset < 0 {
		vl.scrollOffset = 0
	} else {
		vl.scrollOffset = maxOffset
	}
}

// PageUp scrolls up by one page
func (vl *VirtualList) PageUp() int {
	itemsPerPage := vl.viewportHeight / vl.itemHeight
	if itemsPerPage < 1 {
		itemsPerPage = 1
	}
	return vl.ScrollBy(-itemsPerPage)
}

// PageDown scrolls down by one page
func (vl *VirtualList) PageDown() int {
	itemsPerPage := vl.viewportHeight / vl.itemHeight
	if itemsPerPage < 1 {
		itemsPerPage = 1
	}
	return vl.ScrollBy(itemsPerPage)
}

// CanScrollUp returns true if can scroll up
func (vl *VirtualList) CanScrollUp() bool {
	return vl.scrollOffset > 0
}

// CanScrollDown returns true if can scroll down
func (vl *VirtualList) CanScrollDown() bool {
	itemsPerPage := vl.viewportHeight / vl.itemHeight
	if itemsPerPage < 1 {
		itemsPerPage = 1
	}
	maxOffset := vl.itemCount - itemsPerPage
	return vl.scrollOffset < maxOffset && maxOffset > 0
}

// GetScrollOffset returns current scroll offset (in items)
func (vl *VirtualList) GetScrollOffset() int {
	return vl.scrollOffset
}

// GetItemCount returns total number of items
func (vl *VirtualList) GetItemCount() int {
	return vl.itemCount
}

// GetViewportSize returns viewport height (in lines)
func (vl *VirtualList) GetViewportSize() int {
	return vl.viewportHeight
}

// IsScrollable returns true if content is larger than viewport
func (vl *VirtualList) IsScrollable() bool {
	itemsPerPage := vl.viewportHeight / vl.itemHeight
	return vl.itemCount > itemsPerPage
}

// GetVisibleRange returns the start and end indices of visible items
func (vl *VirtualList) GetVisibleRange() (start, end int) {
	itemsPerPage := vl.viewportHeight / vl.itemHeight
	if itemsPerPage < 1 {
		itemsPerPage = 1
	}

	start = vl.scrollOffset
	end = start + itemsPerPage
	if end > vl.itemCount {
		end = vl.itemCount
	}

	return start, end
}

// GetScrollPercent returns scroll position as percentage (0-100)
func (vl *VirtualList) GetScrollPercent() int {
	itemsPerPage := vl.viewportHeight / vl.itemHeight
	if itemsPerPage < 1 {
		itemsPerPage = 1
	}

	maxOffset := vl.itemCount - itemsPerPage
	if maxOffset <= 0 {
		return 100 // Fully visible
	}

	percent := (vl.scrollOffset * 100) / maxOffset
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	return percent
}

// GetScrollIndicator returns a string representation of scroll position
func (vl *VirtualList) GetScrollIndicator() string {
	if !vl.IsScrollable() {
		return ""
	}

	percent := vl.GetScrollPercent()

	// Choose appropriate arrow
	arrow := "↕"
	if percent == 0 {
		arrow = "▼"
	} else if percent >= 100 {
		arrow = "▲"
	}

	return fmt.Sprintf("[%d%% %s]", percent, arrow)
}

// ============================================================================
// ActionTarget 接口实现
// ============================================================================

// HandleAction implements ActionTarget interface
func (vl *VirtualList) HandleAction(act *action.Action) bool {
	if act == nil {
		return false
	}

	switch act.Type {
	case action.ActionScroll:
		// Scroll by delta (from Payload)
		if delta, ok := act.GetPayloadInt(); ok {
			vl.ScrollBy(delta)
			return true
		}

	case action.ActionNavigateUp:
		// Scroll up by line
		vl.ScrollBy(-1)
		return true

	case action.ActionNavigateDown:
		// Scroll down by line
		vl.ScrollBy(1)
		return true

	case action.ActionNavigatePageUp:
		// Page up
		vl.PageUp()
		return true

	case action.ActionNavigatePageDown:
		// Page down
		vl.PageDown()
		return true

	case action.ActionNavigateHome:
		// Scroll to top
		vl.ScrollTop()
		return true

	case action.ActionNavigateEnd:
		// Scroll to bottom
		vl.ScrollBottom()
		return true
	}

	return false
}

// GetSupportedActions implements ActionTarget interface
func (vl *VirtualList) GetSupportedActions() []action.ActionType {
	return []action.ActionType{
		action.ActionScroll,
		action.ActionNavigateUp,
		action.ActionNavigateDown,
		action.ActionNavigatePageUp,
		action.ActionNavigatePageDown,
		action.ActionNavigateHome,
		action.ActionNavigateEnd,
	}
}

// CanHandleAction implements ActionTarget interface
func (vl *VirtualList) CanHandleAction(act *action.Action) bool {
	if act == nil {
		return false
	}

	switch act.Type {
	case action.ActionScroll:
		return vl.IsScrollable()

	case action.ActionNavigateUp:
		return vl.CanScrollUp()

	case action.ActionNavigateDown:
		return vl.CanScrollDown()

	case action.ActionNavigatePageUp:
		return vl.CanScrollUp()

	case action.ActionNavigatePageDown:
		return vl.CanScrollDown()

	case action.ActionNavigateHome:
		return vl.scrollOffset > 0

	case action.ActionNavigateEnd:
		itemsPerPage := vl.viewportHeight / vl.itemHeight
		if itemsPerPage < 1 {
			itemsPerPage = 1
		}
		maxOffset := vl.itemCount - itemsPerPage
		return vl.scrollOffset < maxOffset
	}

	return false
}

// ============================================================================
// ScrollableActionTarget 接口实现
// ============================================================================

// CanScroll implements ScrollableActionTarget interface
// delta > 0 表示向上滚动，delta < 0 表示向下滚动
func (vl *VirtualList) CanScroll(delta int) bool {
	if delta > 0 {
		// 向上滚动（内容向下移动）
		return vl.scrollOffset > 0
	} else if delta < 0 {
		// 向下滚动（内容向上移动）
		return vl.CanScrollDown()
	}
	return vl.IsScrollable()
}

// Scroll implements ScrollableActionTarget interface
// delta > 0 表示向上滚动，delta < 0 表示向下滚动
func (vl *VirtualList) Scroll(delta int) bool {
	newOffset := vl.ScrollBy(delta)
	return newOffset != vl.scrollOffset
}

// GetScrollPosition implements ScrollableActionTarget interface
// 返回 (当前位置, 总范围, 可见范围)
func (vl *VirtualList) GetScrollPosition() (current, total, visible int) {
	itemsPerPage := vl.viewportHeight / vl.itemHeight
	if itemsPerPage < 1 {
		itemsPerPage = 1
	}
	return vl.scrollOffset, vl.itemCount, itemsPerPage
}
