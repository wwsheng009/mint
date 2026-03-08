package list

import (
	"strings"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	scrollutil "github.com/wwsheng009/mint/ui/components/internal/scroll"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for List components
type Instance struct {
	// === Identification ===
	key string

	// === Props (from VNode, may change each render) ===
	header                 string
	rows                   []string
	emptyText              string
	maxRows                int
	showBorder             bool
	showSeparator          bool
	separatorChar          rune
	headerStyle            style.Style
	rowStyle               style.Style
	rowStyleFn             func(int, string) style.Style
	selectedStyle          style.Style
	borderStyle            style.Style
	showScrollbar          bool
	scrollbarStyle         style.Style
	scrollOffset           int
	scrollOffsetControlled bool
	selectedIndex          int
	viewportHeight         int
	allowScroll            bool

	// === Runtime State ===
	bounds [4]int // x, y, w, h
	dirty  bool
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

// NewInstance creates a new ListInstance from props
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:                    getStringProp(props, "key", ""),
		header:                 getStringProp(props, "header", ""),
		rows:                   getStringsProp(props, []string{}),
		emptyText:              getStringProp(props, "emptyText", "(empty)"),
		maxRows:                getIntProp(props, "maxRows", 0),
		showBorder:             getBoolProp(props, "showBorder", true),
		showSeparator:          getBoolProp(props, "showSeparator", true),
		separatorChar:          getRuneProp(props, "separatorChar", '─'),
		headerStyle:            getStyleProp(props, "headerStyle"),
		rowStyle:               getStyleProp(props, "rowStyle"),
		selectedStyle:          getStyleProp(props, "selectedStyle"),
		borderStyle:            getStyleProp(props, "borderStyle"),
		showScrollbar:          getBoolProp(props, "showScrollbar", true),
		scrollbarStyle:         getStyleProp(props, "scrollbarStyle"),
		scrollOffset:           getIntProp(props, "scrollOffset", 0),
		scrollOffsetControlled: getBoolProp(props, "scrollOffsetControlled", false),
		selectedIndex:          getIntProp(props, "selectedIndex", -1),
		viewportHeight:         getIntProp(props, "viewportHeight", 10),
		allowScroll:            getBoolProp(props, "allowScroll", true),
		dirty:                  true,
	}

	// Extract rowStyleFn if provided in props
	if fn, ok := props["rowStyleFn"].(func(int, string) style.Style); ok {
		inst.rowStyleFn = fn
	}

	return inst
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

func (inst *Instance) Key() string           { return inst.key }
func (inst *Instance) SetKey(key string)     { inst.key = key }
func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *Instance) Destroy()              { inst.rows = nil }
func (inst *Instance) OnMount()              { inst.dirty = true }
func (inst *Instance) OnUnmount()            {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldSelected := inst.selectedIndex
	oldScroll := inst.scrollOffset
	oldScrollControlled := inst.scrollOffsetControlled
	oldShowScrollbar := inst.showScrollbar
	oldScrollbarStyle := inst.scrollbarStyle

	inst.header = getStringProp(props, "header", inst.header)
	inst.rows = getStringsProp(props, inst.rows)
	inst.emptyText = getStringProp(props, "emptyText", inst.emptyText)
	inst.maxRows = getIntProp(props, "maxRows", inst.maxRows)
	inst.showBorder = getBoolProp(props, "showBorder", inst.showBorder)
	inst.showSeparator = getBoolProp(props, "showSeparator", inst.showSeparator)
	inst.separatorChar = getRuneProp(props, "separatorChar", inst.separatorChar)
	inst.headerStyle = getStyleProp(props, "headerStyle")
	inst.rowStyle = getStyleProp(props, "rowStyle")
	inst.selectedStyle = getStyleProp(props, "selectedStyle")
	inst.borderStyle = getStyleProp(props, "borderStyle")
	inst.showScrollbar = getBoolProp(props, "showScrollbar", inst.showScrollbar)
	inst.scrollbarStyle = getStyleProp(props, "scrollbarStyle")
	if controlled, ok := props["scrollOffsetControlled"].(bool); ok {
		inst.scrollOffsetControlled = controlled
	}
	if inst.scrollOffsetControlled {
		inst.scrollOffset = getIntProp(props, "scrollOffset", inst.scrollOffset)
	} else if offset, ok := props["scrollOffset"].(int); ok {
		inst.scrollOffset = offset
	}
	inst.selectedIndex = getIntProp(props, "selectedIndex", inst.selectedIndex)
	inst.viewportHeight = getIntProp(props, "viewportHeight", inst.viewportHeight)
	inst.allowScroll = getBoolProp(props, "allowScroll", inst.allowScroll)

	// Update rowStyleFn if provided in props
	if fn, ok := props["rowStyleFn"].(func(int, string) style.Style); ok {
		inst.rowStyleFn = fn
	}

	inst.clampScroll()

	changed := oldSelected != inst.selectedIndex ||
		oldScroll != inst.scrollOffset ||
		oldScrollControlled != inst.scrollOffsetControlled ||
		oldShowScrollbar != inst.showScrollbar ||
		oldScrollbarStyle != inst.scrollbarStyle
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":                    inst.key,
		"header":                 inst.header,
		"rows":                   inst.rows,
		"emptyText":              inst.emptyText,
		"showScrollbar":          inst.showScrollbar,
		"scrollbarStyle":         inst.scrollbarStyle,
		"scrollOffsetControlled": inst.scrollOffsetControlled,
		"scrollOffset":           inst.scrollOffset,
		"selectedIndex":          inst.selectedIndex,
		"viewportHeight":         inst.viewportHeight,
		"allowScroll":            inst.allowScroll,
	}
}

func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }
func (inst *Instance) ClearDirty()                        { inst.dirty = false }

// =============================================================================
// Measurable Interface
// =============================================================================

// Measure implements layout measurement
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	width := 60 // Default width
	height := inst.calculateHeight()

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

// calculateHeight calculates the total height of the list
func (inst *Instance) calculateHeight() int {
	dataHeight := 0

	if inst.maxRows > 0 {
		dataHeight = inst.maxRows
	} else if len(inst.rows) < inst.viewportHeight {
		dataHeight = len(inst.rows)
	} else {
		dataHeight = inst.viewportHeight
	}

	// Add header and separator if present
	if inst.header != "" {
		dataHeight++ // Header row
		if inst.showSeparator && len(inst.rows) > 0 {
			dataHeight++ // Separator line
		}
	}

	// Add border
	if inst.showBorder {
		dataHeight += 2 // Top and bottom border
	}

	// Empty text
	if len(inst.rows) == 0 && inst.emptyText != "" {
		dataHeight = 3
	}

	return dataHeight
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint implements drawing logic for the list
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	if inst.showBorder {
		// Draw with border
		return inst.paintWithBorder(x, y)
	}

	// Draw without border
	return inst.paintWithoutBorder(x, y)
}

// paintWithBorder paints the list with a border
func (inst *Instance) paintWithBorder(x, y int) []paint.DrawCmd {
	var cmds []paint.DrawCmd

	// Calculate width
	width := inst.calculateWidth()

	// Draw top border
	topBorder := "┌" + strings.Repeat("─", width-2) + "┐"
	cmds = append(cmds, paint.NewTextCmd(x, y, topBorder, inst.borderStyle))

	currentY := y + 1

	// Draw header if present
	if inst.header != "" {
		headerLine := "│ " + inst.truncateText(inst.header, width-4) + " │"
		cmds = append(cmds, paint.NewTextCmd(x, currentY, headerLine, inst.headerStyle))
		currentY++
	}

	// Draw separator if present
	if inst.showSeparator && inst.header != "" && len(inst.rows) > 0 {
		sepLine := "│" + strings.Repeat(string(inst.separatorChar), width-2) + "│"
		cmds = append(cmds, paint.NewTextCmd(x, currentY, sepLine, inst.rowStyle))
		currentY++
	}

	dataStartY := currentY
	visibleHeight := inst.visibleHeight()

	// Draw rows
	if len(inst.rows) == 0 {
		// Show empty text
		emptyLine := "│ " + inst.truncateText(inst.emptyText, width-4) + " │"
		cmds = append(cmds, paint.NewTextCmd(x, currentY, emptyLine, inst.rowStyle))
		currentY++
	} else {
		viewport := inst.dataViewport()
		startRow, endRow := viewport.VisibleRange()
		for rowIndex := startRow; rowIndex < endRow && currentY < y+inst.calculateHeight()-1; rowIndex++ {
			rowText := inst.rows[rowIndex]
			rowStyle := inst.rowStyleFor(rowIndex, rowText)
			truncated := inst.truncateText(rowText, width-4)
			rowLine := "│ " + truncated + " │"
			cmds = append(cmds, paint.NewTextCmd(x, currentY, rowLine, rowStyle))
			currentY++
		}

		if inst.showScrollbar {
			scrollbarStyle := inst.scrollbarStyle
			if scrollbarStyle.FG == "" {
				scrollbarStyle = scrollbarStyle.Foreground(inst.borderStyle.FG)
			}
			cmds = append(cmds, scrollutil.DrawVerticalScrollbar(
				x+width-1,
				dataStartY,
				visibleHeight,
				viewport,
				scrollbarStyle,
				scrollutil.DefaultVerticalScrollbarConfig(),
			)...)
		}
	}

	// Draw bottom border
	bottomBorder := "└" + strings.Repeat("─", width-2) + "┘"
	cmds = append(cmds, paint.NewTextCmd(x, currentY, bottomBorder, inst.borderStyle))

	return cmds
}

// paintWithoutBorder paints the list without a border
func (inst *Instance) paintWithoutBorder(x, y int) []paint.DrawCmd {
	cmds := []paint.DrawCmd{}
	currentY := y

	// Draw header if present
	if inst.header != "" {
		cmds = append(cmds, paint.NewTextCmd(x, currentY, inst.header, inst.headerStyle))
		currentY++
	}

	// Draw separator if present
	if inst.showSeparator && inst.header != "" && len(inst.rows) > 0 {
		sepLine := strings.Repeat(string(inst.separatorChar), 50)
		cmds = append(cmds, paint.NewTextCmd(x, currentY, sepLine, inst.rowStyle))
		currentY++
	}

	// Draw rows
	if len(inst.rows) == 0 {
		cmds = append(cmds, paint.NewTextCmd(x, currentY, inst.emptyText, inst.rowStyle))
		currentY++
	} else {
		viewport := inst.dataViewport()
		startRow, endRow := viewport.VisibleRange()
		for rowIndex := startRow; rowIndex < endRow; rowIndex++ {
			rowText := inst.rows[rowIndex]
			rowStyle := inst.rowStyleFor(rowIndex, rowText)
			cmds = append(cmds, paint.NewTextCmd(x, currentY, rowText, rowStyle))
			currentY++
		}
	}

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
	case action.ActionScroll:
		if delta, ok := scrollutil.DeltaFromAction(act); ok {
			return inst.scrollBy(delta)
		}
		return false
	case action.ActionNavigateUp:
		if inst.selectedIndex > 0 {
			return inst.navigateUp()
		}
		return false
	case action.ActionNavigateDown:
		if inst.selectedIndex < len(inst.rows)-1 {
			return inst.navigateDown()
		}
		return false
	case action.ActionNavigateHome:
		if inst.scrollOffset > 0 {
			return inst.navigateHome()
		}
		return false
	case action.ActionNavigateEnd:
		if inst.scrollOffset < inst.dataViewport().MaxOffset() {
			return inst.navigateEnd()
		}
		return false
	case action.ActionNavigatePageUp:
		if inst.scrollOffset > 0 {
			return inst.pageUp()
		}
		return false
	case action.ActionNavigatePageDown:
		if inst.scrollOffset < inst.dataViewport().MaxOffset() {
			return inst.pageDown()
		}
		return false
	case action.ActionSelect:
		if inst.selectedIndex >= 0 {
			return true
		}
		return false
	}
	return false
}

// =============================================================================
// Navigation Methods
// =============================================================================

func (inst *Instance) navigateUp() bool {
	if inst.selectedIndex > 0 {
		inst.selectedIndex--
		inst.dirty = true
		inst.ensureSelectedRowVisible()
		return true
	}
	return false
}

func (inst *Instance) navigateDown() bool {
	if inst.selectedIndex < len(inst.rows)-1 {
		inst.selectedIndex++
		inst.dirty = true
		inst.ensureSelectedRowVisible()
		return true
	}
	return false
}

func (inst *Instance) navigateHome() bool {
	inst.scrollOffset = 0
	if len(inst.rows) > 0 {
		inst.selectedIndex = 0
	} else {
		inst.selectedIndex = -1
	}
	inst.dirty = true
	return true
}

func (inst *Instance) navigateEnd() bool {
	if len(inst.rows) > 0 {
		viewport := inst.dataViewport()
		inst.scrollOffset = viewport.MaxOffset()
		inst.selectedIndex = len(inst.rows) - 1
	} else {
		inst.scrollOffset = 0
		inst.selectedIndex = -1
	}
	inst.dirty = true
	return true
}

func (inst *Instance) pageUp() bool {
	viewport := inst.dataViewport()
	viewport.PageUp()
	inst.scrollOffset = viewport.Offset
	inst.selectedIndex = max(0, inst.selectedIndex-viewport.ViewSize)
	inst.dirty = true
	return true
}

func (inst *Instance) pageDown() bool {
	viewport := inst.dataViewport()
	viewport.PageDown()
	inst.scrollOffset = viewport.Offset
	inst.selectedIndex = min(len(inst.rows)-1, inst.selectedIndex+viewport.ViewSize)
	inst.dirty = true
	return true
}

// =============================================================================
// Helper Methods
// =============================================================================

// ensureSelectedRowVisible ensures the selected row is visible
func (inst *Instance) ensureSelectedRowVisible() {
	if inst.selectedIndex < 0 || inst.selectedIndex >= len(inst.rows) {
		return
	}

	viewport := inst.dataViewport()
	if viewport.EnsureVisible(inst.selectedIndex) {
		inst.scrollOffset = viewport.Offset
	}
}

// clampScroll ensures scroll offset is valid
func (inst *Instance) clampScroll() {
	inst.scrollOffset = inst.dataViewport().Offset
}

func (inst *Instance) scrollBy(delta int) bool {
	viewport := inst.dataViewport()
	if !viewport.ScrollBy(delta) {
		return false
	}
	inst.scrollOffset = viewport.Offset
	inst.dirty = true
	return true
}

func (inst *Instance) visibleHeight() int {
	if inst.maxRows > 0 {
		if inst.maxRows < 1 {
			return 1
		}
		return inst.maxRows
	}
	if inst.viewportHeight < 1 {
		return 1
	}
	return inst.viewportHeight
}

func (inst *Instance) dataViewport() scrollutil.VerticalViewport {
	return scrollutil.NewVerticalViewport(len(inst.rows), inst.visibleHeight(), inst.scrollOffset)
}

func (inst *Instance) rowStyleFor(rowIndex int, rowText string) style.Style {
	if rowIndex == inst.selectedIndex {
		return inst.selectedStyle
	}
	if inst.rowStyleFn != nil {
		return inst.rowStyleFn(rowIndex, rowText)
	}
	return inst.rowStyle
}

// truncateText truncates text to fit within max width
// Uses rune count and StringWidth to properly handle Unicode characters
func (inst *Instance) truncateText(text string, maxWidth int) string {
	// Check display width, not byte length
	textWidth := paint.StringWidth(text)
	if textWidth <= maxWidth {
		return text
	}

	// Truncate and add ellipsis (ellipsis takes 3 display cells)
	if maxWidth > 3 {
		// Use rune-based truncation to avoid cutting UTF-8 multibyte characters
		runes := []rune(text)

		// Keep truncating until we fit within maxWidth - ellipsis width
		for i := len(runes); i > 0; i-- {
			candidateText := string(runes[:i])
			if paint.StringWidth(candidateText)+3 <= maxWidth {
				return candidateText + "..."
			}
		}
	}

	// Fallback: truncate to fit maxWidth
	runes := []rune(text)
	result := ""
	currentWidth := 0
	for _, r := range runes {
		rWidth := paint.RuneWidth(r)
		if currentWidth+rWidth > maxWidth {
			break
		}
		result += string(r)
		currentWidth += rWidth
	}
	return result
}

// calculateWidth calculates the maximum width needed
// Uses StringWidth to properly handle Unicode characters
func (inst *Instance) calculateWidth() int {
	maxWidth := 40 // Minimum width

	// Check header width (using display width, not byte length)
	headerWidth := paint.StringWidth(inst.header)
	if headerWidth+4 > maxWidth {
		maxWidth = headerWidth + 4
	}

	// Check rows width (using display width, not byte length)
	for _, row := range inst.rows {
		rowWidth := paint.StringWidth(row)
		if rowWidth+4 > maxWidth {
			maxWidth = rowWidth + 4
		}
	}

	return maxWidth
}

// =============================================================================
// Getters
// =============================================================================

func (inst *Instance) GetScrollOffset() int   { return inst.scrollOffset }
func (inst *Instance) GetSelectedIndex() int  { return inst.selectedIndex }
func (inst *Instance) GetViewportHeight() int { return inst.viewportHeight }

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

func getStringsProp(props rtui.Props, def []string) []string {
	v, ok := props["rows"]
	if !ok {
		return def
	}
	if rows, ok := v.([]string); ok {
		return rows
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

func getRuneProp(props rtui.Props, key string, def rune) rune {
	if v, ok := props[key]; ok {
		if r, ok := v.(rune); ok {
			return r
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
