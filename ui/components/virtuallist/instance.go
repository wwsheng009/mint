package virtuallist

import (
	"strings"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for virtual list components.
// It persists across renders and holds all state.
type Instance struct {
	// === Identification ===
	key string

	// === Props (from VNode, may change each render) ===
	items         []string
	itemCount     int
	itemHeight    int
	visibleCount  int
	height        int
	width         int
	listStyle     style.Style
	selectedStyle style.Style
	allowScroll   bool

	// === Runtime State ===
	scrollOffset  int    // Current scroll offset
	selectedIndex int    // Currently selected item
	bounds        [4]int // x, y, w, h
	dirty         bool
}

// Ensure Instance implements required interfaces
var (
	_ rtui.ComponentInstance     = (*Instance)(nil)
	_ rtui.PaintableInstance     = (*Instance)(nil)
	_ rtui.ActionHandlerInstance = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// NewInstance creates a new VirtualListInstance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:           getStringProp(props, "key", ""),
		items:         getItemsProp(props, []string{}),
		itemCount:     getIntProp(props, "itemCount", 0),
		itemHeight:    getIntProp(props, "itemHeight", 1),
		visibleCount:  getIntProp(props, "visibleCount", 10),
		height:        getIntProp(props, "height", 10),
		width:         getIntProp(props, "width", 40),
		listStyle:     getStyleProp(props, "listStyle"),
		selectedStyle: getStyleProp(props, "selectedStyle"),
		allowScroll:   getBoolProp(props, "allowScroll", true),
		scrollOffset:  getIntProp(props, "scrollOffset", 0),
		selectedIndex: getIntProp(props, "selectedIndex", -1),
		dirty:         true,
	}
	inst.normalizeItemCount()
	inst.clampOffset()
	inst.clampSelectedIndex()
	return inst
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

func (inst *Instance) Key() string           { return inst.key }
func (inst *Instance) SetKey(key string)     { inst.key = key }
func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *Instance) Destroy()              { inst.items = nil }
func (inst *Instance) OnMount()              { inst.dirty = true }
func (inst *Instance) OnUnmount()            {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldOffset := inst.scrollOffset
	oldSelected := inst.selectedIndex

	inst.items = getItemsProp(props, inst.items)
	inst.itemCount = getIntProp(props, "itemCount", inst.itemCount)
	inst.itemHeight = getIntProp(props, "itemHeight", inst.itemHeight)
	inst.visibleCount = getIntProp(props, "visibleCount", inst.visibleCount)
	inst.height = getIntProp(props, "height", inst.height)
	inst.width = getIntProp(props, "width", inst.width)
	inst.listStyle = getStyleProp(props, "listStyle")
	inst.selectedStyle = getStyleProp(props, "selectedStyle")
	inst.allowScroll = getBoolProp(props, "allowScroll", inst.allowScroll)
	inst.scrollOffset = getIntProp(props, "scrollOffset", inst.scrollOffset)
	inst.selectedIndex = getIntProp(props, "selectedIndex", inst.selectedIndex)

	inst.normalizeItemCount()
	// Clamp offset
	inst.clampOffset()

	// Clamp selected index
	inst.clampSelectedIndex()

	changed := oldOffset != inst.scrollOffset || oldSelected != inst.selectedIndex
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":           inst.key,
		"items":         inst.items,
		"itemCount":     inst.itemCount,
		"itemHeight":    inst.itemHeight,
		"visibleCount":  inst.visibleCount,
		"height":        inst.height,
		"width":         inst.width,
		"allowScroll":   inst.allowScroll,
		"scrollOffset":  inst.scrollOffset,
		"selectedIndex": inst.selectedIndex,
	}
}

func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }
func (inst *Instance) ClearDirty()                        { inst.dirty = false }

// =============================================================================
// Measurable Interface
// =============================================================================

// Measure implements layout measurement.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	width := inst.width
	height := inst.height

	// Apply constraints
	if width < constraints.MinWidth {
		width = constraints.MinWidth
	}
	if width > constraints.MaxWidth && constraints.MaxWidth > 0 {
		width = constraints.MaxWidth
	}
	if height < constraints.MinHeight {
		height = constraints.MinHeight
	}
	if height > constraints.MaxHeight && constraints.MaxHeight > 0 {
		height = constraints.MaxHeight
	}

	return layout.Size{Width: width, Height: height}
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint implements drawing logic for the virtual list.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	var cmds []paint.DrawCmd
	listStyle := inst.listStyle
	width := inst.paintWidth()
	height := inst.paintHeight()
	contentWidth := maxInt(0, width-4)

	// Get visible range
	start, end := inst.getVisibleRange()

	// Draw top border
	topBorder := "┌" + strings.Repeat("─", maxInt(0, width-2)) + "┐"
	cmds = append(cmds, paint.NewTextCmd(x, y, topBorder, listStyle))

	// Draw visible items
	for i := start; i < end; i++ {
		rowY := y + 1 + (i - start)
		if rowY >= y+height-1 {
			break
		}

		// Get item text
		itemText := ""
		if i < len(inst.items) {
			itemText = inst.items[i]
		}

		itemText = inst.truncateText(itemText, contentWidth)
		paddingWidth := maxInt(0, contentWidth-paint.StringWidth(itemText))
		itemText = "│ " + itemText + strings.Repeat(" ", paddingWidth) + " │"

		// Highlight selected item
		itemStyle := listStyle
		if i == inst.selectedIndex {
			itemStyle = inst.selectedStyle
		}

		cmds = append(cmds, paint.NewTextCmd(x, rowY, itemText, itemStyle))
	}

	// Fill remaining space with empty rows
	visibleItemCount := end - start
	for i := visibleItemCount; i < height-2; i++ {
		rowY := y + 1 + i
		emptyRow := "│" + strings.Repeat(" ", maxInt(0, width-2)) + "│"
		cmds = append(cmds, paint.NewTextCmd(x, rowY, emptyRow, listStyle))
	}

	// Draw bottom border
	bottomBorder := "└" + strings.Repeat("─", maxInt(0, width-2)) + "┘"
	cmds = append(cmds, paint.NewTextCmd(x, y+height-1, bottomBorder, listStyle))

	return cmds
}

// =============================================================================
// ActionHandlerInstance Interface
// =============================================================================

func (inst *Instance) HandleAction(act *action.Action) bool {
	if !inst.allowScroll {
		return false
	}

	switch act.Type {
	case action.ActionNavigateUp:
		if inst.selectedIndex > 0 {
			return inst.navigateUp()
		}
		return false
	case action.ActionNavigateDown:
		if inst.selectedIndex < inst.itemCount-1 {
			return inst.navigateDown()
		}
		return false
	case action.ActionNavigateHome:
		if inst.scrollOffset > 0 {
			return inst.scrollTop()
		}
		return false
	case action.ActionNavigateEnd:
		if !inst.isAtEnd() {
			return inst.scrollBottom()
		}
		return false
	case action.ActionNavigatePageUp:
		if inst.canScrollUp() {
			return inst.pageUp()
		}
		return false
	case action.ActionNavigatePageDown:
		if inst.canScrollDown() {
			return inst.pageDown()
		}
		return false
	case action.ActionSelect:
		if inst.selectedIndex >= 0 {
			return true
		}
		return inst.selectItem()
	}
	return false
}

// Scroll methods
func (inst *Instance) handleScroll(payload interface{}) bool {
	if delta, ok := payload.(int); ok {
		inst.scrollBy(delta)

		// Emit scroll end intent if at bottom
		if inst.isAtEnd() {
			// Should emit intent, but for simplicity skip for now
		}
		return true
	}
	return false
}

func (inst *Instance) navigateUp() bool {
	if inst.selectedIndex > 0 {
		inst.selectedIndex--
		inst.dirty = true
		return true
	}
	return false
}

func (inst *Instance) navigateDown() bool {
	if inst.selectedIndex < inst.itemCount-1 {
		inst.selectedIndex++
		inst.dirty = true
		return true
	}
	return false
}

func (inst *Instance) scrollTop() bool {
	inst.scrollOffset = 0
	inst.dirty = true
	return true
}

func (inst *Instance) scrollBottom() bool {
	inst.scrollOffset = inst.getMaxOffset()
	inst.dirty = true
	return true
}

func (inst *Instance) pageUp() bool {
	inst.scrollBy(-inst.visibleCount)
	return true
}

func (inst *Instance) pageDown() bool {
	inst.scrollBy(inst.visibleCount)
	return true
}

func (inst *Instance) selectItem() bool {
	if inst.itemCount > 0 && inst.selectedIndex < 0 {
		inst.selectedIndex = 0
		inst.dirty = true
		return true
	}
	return false
}

// =============================================================================
// Virtual Scrolling Methods
// =============================================================================

func (inst *Instance) scrollBy(delta int) int {
	inst.scrollOffset += delta
	inst.clampOffset()
	return inst.scrollOffset
}

func (inst *Instance) scrollTo(offset int) int {
	inst.scrollOffset = offset
	inst.clampOffset()
	return inst.scrollOffset
}

func (inst *Instance) scrollToItem(index int) int {
	if index < 0 {
		index = 0
	}
	if index >= inst.itemCount {
		index = inst.itemCount - 1
	}
	// Position the item in the middle of the visible area
	targetOffset := index - inst.visibleCount/2
	return inst.scrollTo(targetOffset)
}

func (inst *Instance) clampOffset() {
	maxOffset := inst.getMaxOffset()
	if maxOffset < 0 {
		inst.scrollOffset = 0
	} else if inst.scrollOffset > maxOffset {
		inst.scrollOffset = maxOffset
	}
}

func (inst *Instance) clampSelectedIndex() {
	if inst.selectedIndex < 0 {
		inst.selectedIndex = -1
	} else if inst.selectedIndex >= inst.itemCount {
		inst.selectedIndex = inst.itemCount - 1
	}
}

func (inst *Instance) normalizeItemCount() {
	if inst.itemCount <= 0 || inst.itemCount < len(inst.items) {
		inst.itemCount = len(inst.items)
	}
}

func (inst *Instance) paintWidth() int {
	if inst.bounds[2] > 0 {
		return maxInt(4, inst.bounds[2])
	}
	return maxInt(4, inst.width)
}

func (inst *Instance) paintHeight() int {
	if inst.bounds[3] > 0 {
		return maxInt(2, inst.bounds[3])
	}
	return maxInt(2, inst.height)
}

func (inst *Instance) truncateText(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if paint.StringWidth(text) <= maxWidth {
		return text
	}
	if maxWidth <= 2 {
		return trimToWidth(text, maxWidth)
	}
	return trimToWidth(text, maxWidth-2) + ".."
}

func trimToWidth(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	var builder strings.Builder
	currentWidth := 0
	for _, r := range text {
		runeWidth := paint.RuneWidth(r)
		if currentWidth+runeWidth > maxWidth {
			break
		}
		builder.WriteRune(r)
		currentWidth += runeWidth
	}
	return builder.String()
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func (inst *Instance) getVisibleRange() (start, end int) {
	start = inst.scrollOffset
	end = start + inst.visibleCount
	if end > inst.itemCount {
		end = inst.itemCount
	}
	return start, end
}

func (inst *Instance) isAtEnd() bool {
	return inst.scrollOffset >= inst.itemCount-inst.visibleCount || inst.itemCount <= inst.visibleCount
}

func (inst *Instance) isScrollable() bool {
	return inst.itemCount > inst.visibleCount
}

func (inst *Instance) canScrollUp() bool {
	return inst.scrollOffset > 0
}

func (inst *Instance) canScrollDown() bool {
	return !inst.isAtEnd()
}

func (inst *Instance) getMaxOffset() int {
	maxOffset := inst.itemCount - inst.visibleCount
	if maxOffset < 0 {
		return 0
	}
	return maxOffset
}

// GetVisibleRange returns the start and end index of visible items
func (inst *Instance) GetVisibleRange() (start, end int) {
	return inst.getVisibleRange()
}

// IsItemAtEnd checks if scrolled to the end
func (inst *Instance) IsItemAtEnd() bool {
	return inst.isAtEnd()
}

// GetItem returns the item at the given index
func (inst *Instance) GetItem(index int) string {
	if index >= 0 && index < len(inst.items) {
		return inst.items[index]
	}
	return ""
}

// =============================================================================
// Getters
// =============================================================================

func (inst *Instance) GetOffset() int     { return inst.scrollOffset }
func (inst *Instance) ItemHeight() int    { return inst.itemHeight }
func (inst *Instance) VisibleCount() int  { return inst.visibleCount }
func (inst *Instance) ListHeight() int    { return inst.height }
func (inst *Instance) ListWidth() int     { return inst.width }
func (inst *Instance) SelectedIndex() int { return inst.selectedIndex }

// =============================================================================
// Bounds Support
// =============================================================================

func (inst *Instance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

// =============================================================================
// Prop Extraction Helpers
// =============================================================================

func getStringProp(props rtui.Props, key, def string) string {
	if v, ok := props[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func getBoolProp(props rtui.Props, key string, def bool) bool {
	if v, ok := props[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}

func getIntProp(props rtui.Props, key string, def int) int {
	if v, ok := props[key]; ok {
		if i, ok := v.(int); ok {
			return i
		}
	}
	return def
}

func getStyleProp(props rtui.Props, key string) style.Style {
	v, ok := props[key]
	if !ok {
		return style.Style{}
	}
	if s, ok := v.(style.Style); ok {
		return s
	}
	return style.Style{}
}

func getItemsProp(props rtui.Props, def []string) []string {
	v, ok := props["items"]
	if !ok {
		return def
	}
	if items, ok := v.([]string); ok {
		return items
	}
	return def
}
