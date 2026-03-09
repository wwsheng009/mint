package list

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/form"
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
	changeIntent           intent.Intent
	changeIntentField      intent.FieldIntent
	scrollOffset           int
	scrollOffsetControlled bool
	selectedIndex          int
	viewportHeight         int
	formID                 string
	allowScroll            bool

	// === Runtime State ===
	parent        rtui.ComponentInstance
	focused       bool
	bounds        [4]int // x, y, w, h
	dirty         bool
	intentEmitter func(intent.Intent)
}

// Ensure Instance implements required interfaces
var (
	_ rtui.ComponentInstance     = (*Instance)(nil)
	_ rtui.PaintableInstance     = (*Instance)(nil)
	_ rtui.FocusableInstance     = (*Instance)(nil)
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
		changeIntent:           getIntentProp(props, "changeIntent"),
		changeIntentField:      getChangeIntentFieldProp(props, "changeIntent"),
		scrollOffset:           getIntProp(props, "scrollOffset", 0),
		scrollOffsetControlled: getBoolProp(props, "scrollOffsetControlled", false),
		selectedIndex:          getIntProp(props, "selectedIndex", -1),
		viewportHeight:         getIntProp(props, "viewportHeight", 10),
		formID:                 getStringProp(props, "formID", ""),
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
func (inst *Instance) Parent() interface{}   { return inst.parent }
func (inst *Instance) SetParent(parent rtui.ComponentInstance) {
	inst.parent = parent
}

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldHeader := inst.header
	oldRows := append([]string(nil), inst.rows...)
	oldEmptyText := inst.emptyText
	oldMaxRows := inst.maxRows
	oldShowBorder := inst.showBorder
	oldShowSeparator := inst.showSeparator
	oldSeparatorChar := inst.separatorChar
	oldHeaderStyle := inst.headerStyle
	oldRowStyle := inst.rowStyle
	oldSelectedStyle := inst.selectedStyle
	oldBorderStyle := inst.borderStyle
	oldSelected := inst.selectedIndex
	oldScroll := inst.scrollOffset
	oldScrollControlled := inst.scrollOffsetControlled
	oldShowScrollbar := inst.showScrollbar
	oldScrollbarStyle := inst.scrollbarStyle
	oldChangeIntent := inst.changeIntent
	oldFormID := inst.formID
	oldViewportHeight := inst.viewportHeight
	oldAllowScroll := inst.allowScroll
	oldRowStyleFn := inst.rowStyleFn

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
	inst.changeIntent = getIntentProp(props, "changeIntent")
	inst.changeIntentField = getChangeIntentFieldProp(props, "changeIntent")
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
	inst.formID = getStringProp(props, "formID", inst.formID)
	inst.allowScroll = getBoolProp(props, "allowScroll", inst.allowScroll)

	// Update rowStyleFn if provided in props
	if fn, ok := props["rowStyleFn"].(func(int, string) style.Style); ok {
		inst.rowStyleFn = fn
	} else if _, exists := props["rowStyleFn"]; exists {
		inst.rowStyleFn = nil
	}

	inst.clampSelectedIndex()
	inst.clampScroll()

	changed := oldHeader != inst.header ||
		!equalStrings(oldRows, inst.rows) ||
		oldEmptyText != inst.emptyText ||
		oldMaxRows != inst.maxRows ||
		oldShowBorder != inst.showBorder ||
		oldShowSeparator != inst.showSeparator ||
		oldSeparatorChar != inst.separatorChar ||
		oldHeaderStyle != inst.headerStyle ||
		oldRowStyle != inst.rowStyle ||
		oldSelectedStyle != inst.selectedStyle ||
		oldBorderStyle != inst.borderStyle ||
		oldSelected != inst.selectedIndex ||
		oldScroll != inst.scrollOffset ||
		oldScrollControlled != inst.scrollOffsetControlled ||
		oldShowScrollbar != inst.showScrollbar ||
		oldScrollbarStyle != inst.scrollbarStyle ||
		!sameIntent(oldChangeIntent, inst.changeIntent) ||
		oldFormID != inst.formID ||
		oldViewportHeight != inst.viewportHeight ||
		oldAllowScroll != inst.allowScroll ||
		!sameRowStyleFn(oldRowStyleFn, inst.rowStyleFn)
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
		"changeIntent":           inst.changeIntent,
		"scrollOffsetControlled": inst.scrollOffsetControlled,
		"scrollOffset":           inst.scrollOffset,
		"selectedIndex":          inst.selectedIndex,
		"viewportHeight":         inst.viewportHeight,
		"formID":                 inst.formID,
		"allowScroll":            inst.allowScroll,
	}
}

func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }
func (inst *Instance) ClearDirty()                        { inst.dirty = false }
func (inst *Instance) SetIntentEmitter(fn func(intent.Intent)) {
	inst.intentEmitter = fn
}

// =============================================================================
// Measurable Interface
// =============================================================================

// Measure implements layout measurement
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	width := inst.calculateWidth()
	height := inst.calculateHeight()

	width = constraints.ConstrainWidth(width)
	height = constraints.ConstrainHeight(height)

	return layout.Size{Width: width, Height: height}
}

// calculateHeight calculates the total height of the list
func (inst *Instance) calculateHeight() int {
	dataHeight := 1
	if len(inst.rows) > 0 {
		dataHeight = min(len(inst.rows), inst.visibleHeight())
	}
	if inst.header != "" {
		dataHeight++
		if inst.showSeparator && len(inst.rows) > 0 {
			dataHeight++
		}
	}
	if inst.showBorder {
		dataHeight += 2
	}
	return max(1, dataHeight)
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
	width := inst.effectivePaintWidth()
	borderStyle := inst.borderStyle
	if inst.focused {
		borderStyle = borderStyle.Bold(true)
	}

	// Draw top border
	topBorder := "┌" + strings.Repeat("─", width-2) + "┐"
	cmds = append(cmds, paint.NewTextCmd(x, y, topBorder, borderStyle))

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
				scrollbarStyle = scrollbarStyle.Foreground(borderStyle.FG)
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
	cmds = append(cmds, paint.NewTextCmd(x, currentY, bottomBorder, borderStyle))

	return cmds
}

// paintWithoutBorder paints the list without a border
func (inst *Instance) paintWithoutBorder(x, y int) []paint.DrawCmd {
	cmds := []paint.DrawCmd{}
	currentY := y
	width := inst.effectivePaintWidth()

	// Draw header if present
	if inst.header != "" {
		cmds = append(cmds, paint.NewTextCmd(x, currentY, inst.truncateText(inst.header, width), inst.headerStyle))
		currentY++
	}

	// Draw separator if present
	if inst.showSeparator && inst.header != "" && len(inst.rows) > 0 {
		sepLine := strings.Repeat(string(inst.separatorChar), max(1, width))
		cmds = append(cmds, paint.NewTextCmd(x, currentY, sepLine, inst.rowStyle))
		currentY++
	}

	// Draw rows
	if len(inst.rows) == 0 {
		cmds = append(cmds, paint.NewTextCmd(x, currentY, inst.truncateText(inst.emptyText, width), inst.rowStyle))
		currentY++
	} else {
		viewport := inst.dataViewport()
		startRow, endRow := viewport.VisibleRange()
		for rowIndex := startRow; rowIndex < endRow; rowIndex++ {
			rowText := inst.rows[rowIndex]
			rowStyle := inst.rowStyleFor(rowIndex, rowText)
			cmds = append(cmds, paint.NewTextCmd(x, currentY, inst.truncateText(rowText, width), rowStyle))
			currentY++
		}
	}

	return cmds
}

// =============================================================================
// ActionHandlerInstance Interface
// =============================================================================

func (inst *Instance) HandleAction(act *action.Action) bool {
	if act == nil {
		return false
	}

	switch act.Type {
	case action.ActionScroll:
		if !inst.allowScroll {
			return false
		}
		if delta, ok := scrollutil.DeltaFromAction(act); ok {
			return inst.scrollBy(delta)
		}
		return false
	case action.ActionClick:
		return inst.handleClick(act)
	case action.ActionNavigateUp:
		if len(inst.rows) == 0 {
			return false
		}
		return inst.navigateUp()
	case action.ActionNavigateDown:
		if len(inst.rows) == 0 {
			return false
		}
		return inst.navigateDown()
	case action.ActionNavigateHome:
		if len(inst.rows) == 0 {
			return false
		}
		return inst.navigateHome()
	case action.ActionNavigateEnd:
		if len(inst.rows) == 0 {
			return false
		}
		return inst.navigateEnd()
	case action.ActionNavigatePageUp:
		if len(inst.rows) == 0 {
			return false
		}
		return inst.pageUp()
	case action.ActionNavigatePageDown:
		if len(inst.rows) == 0 {
			return false
		}
		return inst.pageDown()
	case action.ActionSelect, action.ActionEnter:
		if inst.selectedIndex >= 0 && inst.selectedIndex < len(inst.rows) {
			inst.emitSelectionChanged()
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
	if len(inst.rows) == 0 {
		return false
	}
	if inst.selectedIndex < 0 {
		return inst.selectIndex(0, true)
	}
	return inst.selectIndex(inst.selectedIndex-1, true)
}

func (inst *Instance) navigateDown() bool {
	if len(inst.rows) == 0 {
		return false
	}
	if inst.selectedIndex < 0 {
		return inst.selectIndex(0, true)
	}
	return inst.selectIndex(inst.selectedIndex+1, true)
}

func (inst *Instance) navigateHome() bool {
	if len(inst.rows) == 0 {
		return false
	}
	if inst.selectedIndex == 0 && inst.scrollOffset == 0 {
		return false
	}
	inst.scrollOffset = 0
	inst.selectedIndex = 0
	inst.dirty = true
	inst.emitSelectionChanged()
	return true
}

func (inst *Instance) navigateEnd() bool {
	if len(inst.rows) == 0 {
		return false
	}
	lastIndex := len(inst.rows) - 1
	maxOffset := inst.dataViewport().MaxOffset()
	if inst.selectedIndex == lastIndex && inst.scrollOffset == maxOffset {
		return false
	}
	inst.scrollOffset = maxOffset
	inst.selectedIndex = lastIndex
	inst.dirty = true
	inst.emitSelectionChanged()
	return true
}

func (inst *Instance) pageUp() bool {
	viewport := inst.dataViewport()
	if viewport.Offset == 0 && inst.selectedIndex <= 0 {
		return false
	}
	viewport.PageUp()
	inst.scrollOffset = viewport.Offset
	if inst.selectedIndex < 0 {
		inst.selectedIndex = 0
	} else {
		inst.selectedIndex = max(0, inst.selectedIndex-viewport.ViewSize)
	}
	inst.dirty = true
	inst.emitSelectionChanged()
	return true
}

func (inst *Instance) pageDown() bool {
	viewport := inst.dataViewport()
	lastIndex := len(inst.rows) - 1
	if viewport.Offset == viewport.MaxOffset() && inst.selectedIndex >= lastIndex {
		return false
	}
	viewport.PageDown()
	inst.scrollOffset = viewport.Offset
	if inst.selectedIndex < 0 {
		inst.selectedIndex = min(lastIndex, max(0, viewport.ViewSize-1))
	} else {
		inst.selectedIndex = min(lastIndex, inst.selectedIndex+viewport.ViewSize)
	}
	inst.dirty = true
	inst.emitSelectionChanged()
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

func (inst *Instance) clampSelectedIndex() {
	if len(inst.rows) == 0 {
		inst.selectedIndex = -1
		return
	}
	if inst.selectedIndex >= len(inst.rows) {
		inst.selectedIndex = len(inst.rows) - 1
	}
	if inst.selectedIndex < -1 {
		inst.selectedIndex = -1
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

func (inst *Instance) effectiveVisibleHeight() int {
	height := inst.visibleHeight()
	if inst.bounds[3] <= 0 {
		return max(1, height)
	}
	available := inst.bounds[3] - inst.chromeHeight()
	if available < 1 {
		return 1
	}
	return min(height, available)
}

func (inst *Instance) chromeHeight() int {
	chrome := 0
	if inst.showBorder {
		chrome += 2
	}
	if inst.header != "" {
		chrome++
		if inst.showSeparator && len(inst.rows) > 0 {
			chrome++
		}
	}
	return chrome
}

func (inst *Instance) dataViewport() scrollutil.VerticalViewport {
	return scrollutil.NewVerticalViewport(len(inst.rows), inst.effectiveVisibleHeight(), inst.scrollOffset)
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

func (inst *Instance) effectivePaintWidth() int {
	if inst.bounds[2] > 0 {
		return max(4, inst.bounds[2])
	}
	return max(4, inst.calculateWidth())
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

func (inst *Instance) handleClick(act *action.Action) bool {
	mouseMsg, ok := act.Payload.(*runtimemsg.MouseMsg)
	if !ok || mouseMsg == nil {
		return false
	}
	rowIndex, ok := inst.rowIndexAtLocalY(mouseMsg.LocalY)
	if !ok {
		return false
	}
	if rowIndex == inst.selectedIndex {
		inst.emitSelectionChanged()
		return true
	}
	return inst.selectIndex(rowIndex, true)
}

func (inst *Instance) rowIndexAtLocalY(localY int) (int, bool) {
	if len(inst.rows) == 0 {
		return -1, false
	}
	dataStart := 0
	if inst.showBorder {
		dataStart++
	}
	if inst.header != "" {
		dataStart++
		if inst.showSeparator && len(inst.rows) > 0 {
			dataStart++
		}
	}
	relative := localY - dataStart
	if relative < 0 || relative >= inst.effectiveVisibleHeight() {
		return -1, false
	}
	viewport := inst.dataViewport()
	startRow, endRow := viewport.VisibleRange()
	rowIndex := startRow + relative
	if rowIndex < startRow || rowIndex >= endRow || rowIndex >= len(inst.rows) {
		return -1, false
	}
	return rowIndex, true
}

func (inst *Instance) selectIndex(index int, emit bool) bool {
	if len(inst.rows) == 0 {
		return false
	}
	clamped := max(0, min(len(inst.rows)-1, index))
	changed := inst.selectedIndex != clamped
	inst.selectedIndex = clamped
	inst.ensureSelectedRowVisible()
	if changed {
		inst.dirty = true
		if emit {
			inst.emitSelectionChanged()
		}
	}
	return changed
}

func (inst *Instance) emitSelectionChanged() {
	if inst.intentEmitter == nil {
		return
	}
	value := fmt.Sprintf("%d", inst.selectedIndex)
	if inst.formID != "" {
		if inst.changeIntentField != nil {
			intent.Emit(inst, form.FieldChange(inst.formID, inst.changeIntentField.GetField(), value, true))
		}
		return
	}
	if inst.changeIntentField != nil {
		inst.intentEmitter(intent.FieldChangeIntent{
			Field: inst.changeIntentField.GetField(),
			Value: value,
		})
		return
	}
	if inst.changeIntent != nil {
		inst.intentEmitter(inst.changeIntent)
	}
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
// FocusableInstance Interface
// =============================================================================

func (inst *Instance) SetFocus(focused bool) {
	if inst.focused == focused {
		return
	}
	inst.focused = focused
	inst.dirty = true
}

func (inst *Instance) HasFocus() bool { return inst.focused }
func (inst *Instance) IsDisabled() bool {
	return false
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

func getIntentProp(props rtui.Props, key string) intent.Intent {
	if v, ok := props[key]; ok {
		if i, ok := v.(intent.Intent); ok {
			return i
		}
	}
	return nil
}

func getChangeIntentFieldProp(props rtui.Props, key string) intent.FieldIntent {
	if v, ok := props[key]; ok {
		if i, ok := v.(intent.FieldIntent); ok {
			return i
		}
	}
	return nil
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func sameRowStyleFn(left, right func(int, string) style.Style) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return reflect.ValueOf(left).Pointer() == reflect.ValueOf(right).Pointer()
}

func sameIntent(left, right intent.Intent) bool {
	return reflect.DeepEqual(left, right)
}
