package list

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/form"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
	scrollutil "github.com/wwsheng009/mint/ui/components/internal/scroll"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for List components
type Instance struct {
	// === Identification ===
	key         string
	componentID string

	// === Props (from VNode, may change each render) ===
	header                   string
	items                    []RowItem
	rows                     []string
	emptyText                string
	maxRows                  int
	showBorder               bool
	showSeparator            bool
	separatorChar            rune
	headerStyle              style.Style
	rowStyle                 style.Style
	rowStyleFn               func(int, string) style.Style
	matchStyle               style.Style
	selectedStyle            style.Style
	borderStyle              style.Style
	showScrollbar            bool
	scrollbarStyle           style.Style
	changeIntent             intent.Intent
	changeIntentField        intent.FieldIntent
	selectionIntent          intent.Intent
	selectionIntentField     intent.FieldIntent
	selectionMode            SelectionMode
	searchQuery              string
	searchFn                 func(string, string) bool
	showSearchStats          bool
	searchStatsStyle         style.Style
	scrollOffset             int
	scrollOffsetControlled   bool
	selectedIndex            int
	selectedIndexControlled  bool
	checkedIndices           []int
	checkedIndicesControlled bool
	viewportHeight           int
	formID                   string
	allowScroll              bool

	// === Runtime State ===
	parent                    rtui.ComponentInstance
	focused                   bool
	bounds                    [4]int // x, y, w, h
	dirty                     bool
	scrollOffsetInitialized   bool
	selectedIndexInitialized  bool
	checkedIndicesInitialized bool
	pendingScrollOffset       int
	hasPendingScrollOffset    bool
	pendingSelectedIndex      int
	hasPendingSelectedIndex   bool
	pendingCheckedIndices     []int
	hasPendingCheckedIndices  bool
	lastPropScrollOffset      int
	lastPropSelectedIndex     int
	lastPropCheckedIndices    []int
	lastSearchQuery           string
	lastSearchTotal           int
	lastSearchSelected        int
	intentEmitter             func(intent.Intent)
	pendingEnsureVisible      bool
}

// Ensure Instance implements required interfaces
var (
	_ rtui.ComponentInstance     = (*Instance)(nil)
	_ rtui.PaintableInstance     = (*Instance)(nil)
	_ rtui.FocusableInstance     = (*Instance)(nil)
	_ rtui.ActionHandlerInstance = (*Instance)(nil)
	_ intent.IntentHandler       = (*Instance)(nil)
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
		key:                      proputil.GetString(props, "key", ""),
		componentID:              proputil.GetString(props, "componentID", ""),
		header:                   proputil.GetString(props, "header", ""),
		emptyText:                proputil.GetString(props, "emptyText", "(empty)"),
		maxRows:                  proputil.GetInt(props, "maxRows", 0),
		showBorder:               proputil.GetBool(props, "showBorder", true),
		showSeparator:            proputil.GetBool(props, "showSeparator", true),
		separatorChar:            getRuneProp(props, "separatorChar", '─'),
		headerStyle:              proputil.GetStyle(props, "headerStyle", style.Style{}),
		rowStyle:                 proputil.GetStyle(props, "rowStyle", style.Style{}),
		matchStyle:               proputil.GetStyle(props, "matchStyle", style.Style{}),
		selectedStyle:            proputil.GetStyle(props, "selectedStyle", style.Style{}),
		borderStyle:              proputil.GetStyle(props, "borderStyle", style.Style{}),
		showScrollbar:            proputil.GetBool(props, "showScrollbar", true),
		scrollbarStyle:           proputil.GetStyle(props, "scrollbarStyle", style.Style{}),
		changeIntent:             proputil.GetIntent(props, "changeIntent", nil),
		changeIntentField:        getChangeIntentFieldProp(props, "changeIntent"),
		selectionIntent:          proputil.GetIntent(props, "selectionIntent", nil),
		selectionIntentField:     getChangeIntentFieldProp(props, "selectionIntent"),
		selectionMode:            getSelectionModeProp(props, "selectionMode", SelectionNone),
		searchQuery:              proputil.GetString(props, "searchQuery", ""),
		searchFn:                 getSearchFn(props),
		showSearchStats:          proputil.GetBool(props, "showSearchStats", false),
		searchStatsStyle:         proputil.GetStyle(props, "searchStatsStyle", style.Style{}),
		scrollOffset:             proputil.GetInt(props, "scrollOffset", 0),
		scrollOffsetControlled:   proputil.GetBool(props, "scrollOffsetControlled", false),
		selectedIndex:            proputil.GetInt(props, "selectedIndex", -1),
		selectedIndexControlled:  proputil.GetBool(props, "selectedIndexControlled", false),
		checkedIndices:           getIntsProp(props, "checkedIndices", nil),
		checkedIndicesControlled: proputil.GetBool(props, "checkedIndicesControlled", false),
		viewportHeight:           proputil.GetInt(props, "viewportHeight", 10),
		formID:                   proputil.GetString(props, "formID", ""),
		allowScroll:              proputil.GetBool(props, "allowScroll", true),
		lastPropScrollOffset:     proputil.GetInt(props, "scrollOffset", 0),
		lastPropSelectedIndex:    proputil.GetInt(props, "selectedIndex", -1),
		lastPropCheckedIndices:   getIntsProp(props, "checkedIndices", nil),
		dirty:                    true,
	}
	inst.items, inst.rows = getListRowsAndItems(props, []RowItem{}, []string{})

	// Extract rowStyleFn if provided in props
	if fn, ok := props[propRowStyleFn].(func(int, string) style.Style); ok {
		inst.rowStyleFn = fn
	}
	inst.scrollOffsetInitialized = inst.scrollOffsetControlled || hasProp(props, "scrollOffset")
	inst.selectedIndexInitialized = inst.selectedIndexControlled || hasProp(props, "selectedIndex")
	inst.checkedIndicesInitialized = inst.checkedIndicesControlled || hasProp(props, "checkedIndices")
	inst.clampSelectedIndex()
	inst.normalizeCheckedIndices()
	inst.pendingEnsureVisible = true
	inst.normalizeSelectionAndScroll()

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
	oldComponentID := inst.componentID
	oldHeader := inst.header
	oldItems := cloneItems(inst.items)
	oldRows := append([]string(nil), inst.rows...)
	oldEmptyText := inst.emptyText
	oldMaxRows := inst.maxRows
	oldShowBorder := inst.showBorder
	oldShowSeparator := inst.showSeparator
	oldSeparatorChar := inst.separatorChar
	oldHeaderStyle := inst.headerStyle
	oldRowStyle := inst.rowStyle
	oldMatchStyle := inst.matchStyle
	oldSelectedStyle := inst.selectedStyle
	oldBorderStyle := inst.borderStyle
	oldSelected := inst.selectedIndex
	oldScroll := inst.scrollOffset
	oldScrollControlled := inst.scrollOffsetControlled
	oldSelectedControlled := inst.selectedIndexControlled
	oldPropScroll := inst.lastPropScrollOffset
	oldPropSelected := inst.lastPropSelectedIndex
	oldPropChecked := append([]int(nil), inst.lastPropCheckedIndices...)
	oldShowScrollbar := inst.showScrollbar
	oldScrollbarStyle := inst.scrollbarStyle
	oldChangeIntent := inst.changeIntent
	oldSelectionIntent := inst.selectionIntent
	oldSelectionMode := inst.selectionMode
	oldSearchQuery := inst.searchQuery
	oldSearchFn := inst.searchFn
	oldShowSearchStats := inst.showSearchStats
	oldSearchStatsStyle := inst.searchStatsStyle
	oldCheckedIndices := append([]int(nil), inst.checkedIndices...)
	oldCheckedIndicesControlled := inst.checkedIndicesControlled
	oldFormID := inst.formID
	oldViewportHeight := inst.viewportHeight
	oldAllowScroll := inst.allowScroll
	oldRowStyleFn := inst.rowStyleFn

	inst.componentID = proputil.GetString(props, "componentID", inst.componentID)
	inst.header = proputil.GetString(props, "header", inst.header)
	inst.items, inst.rows = getListRowsAndItems(props, inst.items, inst.rows)
	inst.emptyText = proputil.GetString(props, "emptyText", inst.emptyText)
	inst.maxRows = proputil.GetInt(props, "maxRows", inst.maxRows)
	inst.showBorder = proputil.GetBool(props, "showBorder", inst.showBorder)
	inst.showSeparator = proputil.GetBool(props, "showSeparator", inst.showSeparator)
	inst.separatorChar = getRuneProp(props, "separatorChar", inst.separatorChar)
	inst.headerStyle = proputil.GetStyle(props, "headerStyle", style.Style{})
	inst.rowStyle = proputil.GetStyle(props, "rowStyle", style.Style{})
	inst.matchStyle = proputil.GetStyle(props, "matchStyle", style.Style{})
	inst.selectedStyle = proputil.GetStyle(props, "selectedStyle", style.Style{})
	inst.borderStyle = proputil.GetStyle(props, "borderStyle", style.Style{})
	inst.showScrollbar = proputil.GetBool(props, "showScrollbar", inst.showScrollbar)
	inst.scrollbarStyle = proputil.GetStyle(props, "scrollbarStyle", style.Style{})
	inst.changeIntent = proputil.GetIntent(props, "changeIntent", nil)
	inst.changeIntentField = getChangeIntentFieldProp(props, "changeIntent")
	inst.selectionIntent = proputil.GetIntent(props, "selectionIntent", nil)
	inst.selectionIntentField = getChangeIntentFieldProp(props, "selectionIntent")
	inst.selectionMode = getSelectionModeProp(props, "selectionMode", inst.selectionMode)
	inst.searchQuery = proputil.GetString(props, "searchQuery", inst.searchQuery)
	inst.searchFn = getSearchFnOrCurrent(props, inst.searchFn)
	inst.showSearchStats = proputil.GetBool(props, "showSearchStats", inst.showSearchStats)
	inst.searchStatsStyle = proputil.GetStyle(props, "searchStatsStyle", style.Style{})
	if controlled, ok := props[propScrollOffsetControlled].(bool); ok {
		inst.scrollOffsetControlled = controlled
	}
	if inst.scrollOffsetControlled {
		nextScroll := proputil.GetInt(props, "scrollOffset", inst.scrollOffset)
		inst.lastPropScrollOffset = nextScroll
		if inst.hasPendingScrollOffset {
			if nextScroll == inst.pendingScrollOffset || nextScroll != oldPropScroll {
				inst.scrollOffset = nextScroll
				inst.hasPendingScrollOffset = false
			} else {
				inst.scrollOffset = inst.pendingScrollOffset
			}
		} else {
			inst.scrollOffset = nextScroll
		}
		inst.scrollOffsetInitialized = true
	} else if offset, ok := props[propScrollOffset].(int); ok {
		if !inst.scrollOffsetInitialized {
			inst.scrollOffset = offset
			inst.scrollOffsetInitialized = true
		}
		inst.lastPropScrollOffset = inst.scrollOffset
		inst.hasPendingScrollOffset = false
	}
	if !inst.scrollOffsetControlled {
		inst.lastPropScrollOffset = inst.scrollOffset
		inst.hasPendingScrollOffset = false
	}
	if controlled, ok := props[propSelectedIndexControlled].(bool); ok {
		inst.selectedIndexControlled = controlled
	}
	if inst.selectedIndexControlled {
		nextSelected := proputil.GetInt(props, "selectedIndex", inst.selectedIndex)
		inst.lastPropSelectedIndex = nextSelected
		if inst.hasPendingSelectedIndex {
			if nextSelected == inst.pendingSelectedIndex || nextSelected != oldPropSelected {
				inst.selectedIndex = nextSelected
				inst.hasPendingSelectedIndex = false
			} else {
				inst.selectedIndex = inst.pendingSelectedIndex
			}
		} else {
			inst.selectedIndex = nextSelected
		}
		inst.selectedIndexInitialized = true
	} else if selectedIndex, ok := props[propSelectedIndex].(int); ok {
		if !inst.selectedIndexInitialized {
			inst.selectedIndex = selectedIndex
			inst.selectedIndexInitialized = true
		}
		inst.lastPropSelectedIndex = inst.selectedIndex
		inst.hasPendingSelectedIndex = false
	}
	if !inst.selectedIndexControlled {
		inst.lastPropSelectedIndex = inst.selectedIndex
		inst.hasPendingSelectedIndex = false
	}
	if inst.selectedIndex != oldSelected {
		inst.pendingEnsureVisible = true
	}
	if controlled, ok := props[propCheckedIndicesControlled].(bool); ok {
		inst.checkedIndicesControlled = controlled
	}
	if inst.checkedIndicesControlled {
		nextChecked := getIntsProp(props, "checkedIndices", inst.checkedIndices)
		inst.lastPropCheckedIndices = append([]int(nil), nextChecked...)
		if inst.hasPendingCheckedIndices {
			if equalInts(nextChecked, inst.pendingCheckedIndices) || !equalInts(nextChecked, oldPropChecked) {
				inst.checkedIndices = nextChecked
				inst.hasPendingCheckedIndices = false
			} else {
				inst.checkedIndices = append([]int(nil), inst.pendingCheckedIndices...)
			}
		} else {
			inst.checkedIndices = nextChecked
		}
		inst.checkedIndicesInitialized = true
	} else if checkedIndices, ok := props[propCheckedIndices].([]int); ok {
		if !inst.checkedIndicesInitialized {
			inst.checkedIndices = append([]int(nil), checkedIndices...)
			inst.checkedIndicesInitialized = true
		}
		inst.lastPropCheckedIndices = append([]int(nil), inst.checkedIndices...)
		inst.hasPendingCheckedIndices = false
	}
	if !inst.checkedIndicesControlled {
		inst.lastPropCheckedIndices = append([]int(nil), inst.checkedIndices...)
		inst.hasPendingCheckedIndices = false
	}
	inst.viewportHeight = proputil.GetInt(props, "viewportHeight", inst.viewportHeight)
	inst.formID = proputil.GetString(props, "formID", inst.formID)
	inst.allowScroll = proputil.GetBool(props, "allowScroll", inst.allowScroll)

	// Update rowStyleFn if provided in props
	if fn, ok := props[propRowStyleFn].(func(int, string) style.Style); ok {
		inst.rowStyleFn = fn
	} else if _, exists := props[propRowStyleFn]; exists {
		inst.rowStyleFn = nil
	}

	inst.clampSelectedIndex()
	inst.normalizeCheckedIndices()
	inst.normalizeSelectionAndScroll()
	inst.syncPendingControlledState()

	changed := oldHeader != inst.header ||
		oldComponentID != inst.componentID ||
		!equalItems(oldItems, inst.items) ||
		!equalStrings(oldRows, inst.rows) ||
		oldEmptyText != inst.emptyText ||
		oldMaxRows != inst.maxRows ||
		oldShowBorder != inst.showBorder ||
		oldShowSeparator != inst.showSeparator ||
		oldSeparatorChar != inst.separatorChar ||
		oldHeaderStyle != inst.headerStyle ||
		oldRowStyle != inst.rowStyle ||
		oldMatchStyle != inst.matchStyle ||
		oldSelectedStyle != inst.selectedStyle ||
		oldBorderStyle != inst.borderStyle ||
		oldSelected != inst.selectedIndex ||
		oldScroll != inst.scrollOffset ||
		oldScrollControlled != inst.scrollOffsetControlled ||
		oldSelectedControlled != inst.selectedIndexControlled ||
		oldShowScrollbar != inst.showScrollbar ||
		oldScrollbarStyle != inst.scrollbarStyle ||
		!sameIntent(oldChangeIntent, inst.changeIntent) ||
		!sameIntent(oldSelectionIntent, inst.selectionIntent) ||
		oldSelectionMode != inst.selectionMode ||
		oldSearchQuery != inst.searchQuery ||
		!sameSearchFn(oldSearchFn, inst.searchFn) ||
		oldShowSearchStats != inst.showSearchStats ||
		oldSearchStatsStyle != inst.searchStatsStyle ||
		oldCheckedIndicesControlled != inst.checkedIndicesControlled ||
		!equalInts(oldCheckedIndices, inst.checkedIndices) ||
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
	props := rtui.Props{
		propKey:                      inst.key,
		propComponentID:              inst.componentID,
		propHeader:                   inst.header,
		propItems:                    cloneItems(inst.items),
		propRows:                     inst.rows,
		propEmptyText:                inst.emptyText,
		propShowScrollbar:            inst.showScrollbar,
		propScrollbarStyle:           inst.scrollbarStyle,
		propChangeIntent:             inst.changeIntent,
		propSelectionIntent:          inst.selectionIntent,
		propSelectionMode:            inst.selectionMode,
		propMatchStyle:               inst.matchStyle,
		propSearchQuery:              inst.searchQuery,
		propShowSearchStats:          inst.showSearchStats,
		propSearchStatsStyle:         inst.searchStatsStyle,
		propScrollOffsetControlled:   inst.scrollOffsetControlled,
		propScrollOffset:             inst.scrollOffset,
		propSelectedIndex:            inst.selectedIndex,
		propSelectedIndexControlled:  inst.selectedIndexControlled,
		propCheckedIndices:           append([]int(nil), inst.checkedIndices...),
		propCheckedIndicesControlled: inst.checkedIndicesControlled,
		propViewportHeight:           inst.viewportHeight,
		propFormID:                   inst.formID,
		propAllowScroll:              inst.allowScroll,
	}
	if inst.searchFn != nil {
		props[propSearchFn] = inst.searchFn
	}
	return props
}

func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }
func (inst *Instance) ClearDirty()                        { inst.dirty = false }
func (inst *Instance) SetIntentEmitter(fn func(intent.Intent)) {
	inst.intentEmitter = fn
}

func (inst *Instance) recordPendingScroll() {
	if !inst.scrollOffsetControlled {
		return
	}
	inst.pendingScrollOffset = inst.scrollOffset
	inst.hasPendingScrollOffset = true
}

func (inst *Instance) recordPendingSelected() {
	if !inst.selectedIndexControlled {
		return
	}
	inst.pendingSelectedIndex = inst.selectedIndex
	inst.hasPendingSelectedIndex = true
}

func (inst *Instance) recordPendingChecked() {
	if !inst.checkedIndicesControlled {
		return
	}
	inst.pendingCheckedIndices = append([]int(nil), inst.checkedIndices...)
	inst.hasPendingCheckedIndices = true
}

func (inst *Instance) syncPendingControlledState() {
	if inst.hasPendingScrollOffset {
		inst.pendingScrollOffset = inst.scrollOffset
	}
	if inst.hasPendingSelectedIndex {
		inst.pendingSelectedIndex = inst.selectedIndex
	}
	if inst.hasPendingCheckedIndices {
		inst.pendingCheckedIndices = append([]int(nil), inst.checkedIndices...)
	}
}

func (inst *Instance) emitLocalIntent(i intent.Intent) {
	if i == nil {
		return
	}
	intent.Emit(inst, i)
}

func (inst *Instance) emitRowSelect(rowIndex int) {
	if rowIndex < 0 || rowIndex >= len(inst.rows) {
		return
	}
	if inst.componentID != "" {
		inst.emitLocalIntent(RowSelectWithID(inst.componentID, rowIndex, inst.rows[rowIndex]))
		return
	}
	inst.emitLocalIntent(RowSelect(rowIndex, inst.rows[rowIndex]))
}

func (inst *Instance) emitNavigation(direction string, fromIndex, toIndex int) {
	if inst.componentID != "" {
		inst.emitLocalIntent(NavigationWithID(inst.componentID, direction, fromIndex, toIndex))
		return
	}
	inst.emitLocalIntent(Navigation(direction, fromIndex, toIndex))
}

func (inst *Instance) emitScrollIntent(delta int) {
	if delta == 0 {
		return
	}
	viewSize := inst.effectiveVisibleHeight()
	contentSize := len(inst.rows)
	if inst.componentID != "" {
		inst.emitLocalIntent(ScrollWithID(inst.componentID, inst.scrollOffset, delta, viewSize, contentSize))
		return
	}
	inst.emitLocalIntent(Scroll(inst.scrollOffset, delta, viewSize, contentSize))
}

func (inst *Instance) emitSearchStats() {
	query := strings.TrimSpace(inst.searchQuery)
	total, selected := inst.matchStats()
	if inst.lastSearchQuery == query && inst.lastSearchTotal == total && inst.lastSearchSelected == selected {
		return
	}
	inst.lastSearchQuery = query
	inst.lastSearchTotal = total
	inst.lastSearchSelected = selected
	if inst.componentID != "" {
		inst.emitLocalIntent(SearchStatsWithID(inst.componentID, query, total, selected))
		return
	}
	inst.emitLocalIntent(SearchStats(query, total, selected))
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
	visibleRows := inst.visibleRowIndices()
	dataHeight := 1
	if len(visibleRows) > 0 {
		dataHeight = min(len(visibleRows), inst.visibleHeight())
	}
	dataHeight += inst.statsHeight()
	if inst.header != "" {
		dataHeight++
		if inst.showSeparator && len(visibleRows) > 0 {
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
	visibleRows := inst.visibleRowIndices()
	borderStyle := inst.borderStyle
	if inst.focused {
		borderStyle = borderStyle.Bold(true)
	}

	// Draw top border
	topBorder := "┌" + strings.Repeat("─", width-2) + "┐"
	cmds = append(cmds, paint.NewTextCmd(x, y, topBorder, borderStyle))

	currentY := y + 1

	if inst.showSearchStats {
		statsLine := "│ " + inst.truncateText(inst.searchStatsLine(), width-4) + " │"
		statsStyle := inst.searchStatsStyle
		if statsStyle == (style.Style{}) {
			statsStyle = inst.rowStyle
		}
		cmds = append(cmds, paint.NewTextCmd(x, currentY, statsLine, statsStyle))
		currentY++
	}

	// Draw header if present
	if inst.header != "" {
		headerLine := "│ " + inst.truncateText(inst.headerDisplayText(), width-4) + " │"
		cmds = append(cmds, paint.NewTextCmd(x, currentY, headerLine, inst.headerStyle))
		currentY++
	}

	// Draw separator if present
	if inst.showSeparator && inst.header != "" && len(visibleRows) > 0 {
		sepLine := "│" + strings.Repeat(string(inst.separatorChar), width-2) + "│"
		cmds = append(cmds, paint.NewTextCmd(x, currentY, sepLine, inst.rowStyle))
		currentY++
	}

	dataStartY := currentY
	viewport := inst.dataViewportFor(len(visibleRows))
	visibleHeight := viewport.ViewSize

	// Draw rows
	if len(visibleRows) == 0 {
		// Show empty text
		emptyLine := "│ " + inst.truncateText(inst.emptyDisplayText(), width-4) + " │"
		cmds = append(cmds, paint.NewTextCmd(x, currentY, emptyLine, inst.rowStyle))
		currentY++
	} else {
		startRow, endRow := viewport.VisibleRange()
		for visibleIndex := startRow; visibleIndex < endRow && currentY < y+inst.calculateHeight()-1; visibleIndex++ {
			rowIndex := visibleRows[visibleIndex]
			rowText := inst.rows[rowIndex]
			rowStyle := inst.rowStyleFor(rowIndex, rowText, inst.searchActive())
			truncated := inst.truncateText(inst.rowDisplayText(rowIndex, rowText), width-4)
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
	visibleRows := inst.visibleRowIndices()

	if inst.showSearchStats {
		statsStyle := inst.searchStatsStyle
		if statsStyle == (style.Style{}) {
			statsStyle = inst.rowStyle
		}
		cmds = append(cmds, paint.NewTextCmd(x, currentY, inst.truncateText(inst.searchStatsLine(), width), statsStyle))
		currentY++
	}

	// Draw header if present
	if inst.header != "" {
		cmds = append(cmds, paint.NewTextCmd(x, currentY, inst.truncateText(inst.headerDisplayText(), width), inst.headerStyle))
		currentY++
	}

	// Draw separator if present
	if inst.showSeparator && inst.header != "" && len(visibleRows) > 0 {
		sepLine := strings.Repeat(string(inst.separatorChar), max(1, width))
		cmds = append(cmds, paint.NewTextCmd(x, currentY, sepLine, inst.rowStyle))
		currentY++
	}

	// Draw rows
	if len(visibleRows) == 0 {
		cmds = append(cmds, paint.NewTextCmd(x, currentY, inst.truncateText(inst.emptyDisplayText(), width), inst.rowStyle))
		currentY++
	} else {
		viewport := inst.dataViewportFor(len(visibleRows))
		startRow, endRow := viewport.VisibleRange()
		for visibleIndex := startRow; visibleIndex < endRow; visibleIndex++ {
			rowIndex := visibleRows[visibleIndex]
			rowText := inst.rows[rowIndex]
			rowStyle := inst.rowStyleFor(rowIndex, rowText, inst.searchActive())
			cmds = append(cmds, paint.NewTextCmd(x, currentY, inst.truncateText(inst.rowDisplayText(rowIndex, rowText), width), rowStyle))
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
	case action.ActionSearch:
		if direction, ok := act.Payload.(string); ok {
			switch strings.ToLower(strings.TrimSpace(direction)) {
			case "next":
				return inst.navigateMatch(1)
			case "prev":
				return inst.navigateMatch(-1)
			}
		}
		return false
	case action.ActionSelect, action.ActionEnter:
		return inst.handleActivate()
	}
	return false
}

// HandleIntent implements intent.IntentHandler to allow external control of the list.
func (inst *Instance) HandleIntent(i intent.Intent) bool {
	if !intent.ShouldHandleIntentWithID(inst.componentID, i) {
		return false
	}

	switch v := i.(type) {
	case SelectNextIntent:
		return inst.navigateDown()
	case SelectPrevIntent:
		return inst.navigateUp()
	case SelectByIndexIntent:
		if v.Index == -1 {
			return inst.clearSelection(true)
		}
		if v.Index < 0 || v.Index >= len(inst.rows) {
			return false
		}
		return inst.selectIndex(v.Index, true, true)
	case ClearSelectionIntent:
		return inst.clearSelection(true)
	case ScrollToIntent:
		return inst.scrollTo(v.Offset, true)
	case ScrollByIntent:
		return inst.scrollBy(v.Delta)
	case SearchNextIntent:
		return inst.navigateMatch(1)
	case SearchPrevIntent:
		return inst.navigateMatch(-1)
	case ToggleCheckedIntent:
		return inst.toggleCheckedAtIndex(v.Index, true)
	case ClearCheckedIntent:
		return inst.clearCheckedSelection(true)
	}

	return false
}

// =============================================================================
// Navigation Methods
// =============================================================================

func (inst *Instance) navigateUp() bool {
	visibleRows := inst.visibleRowIndices()
	if len(visibleRows) == 0 {
		return false
	}
	oldSelected := inst.selectedIndex
	currentVisible := inst.visibleRowPosition(visibleRows, inst.selectedIndex)
	if currentVisible <= 0 {
		if !inst.selectIndex(visibleRows[0], true, true) {
			return false
		}
		inst.emitNavigation("up", oldSelected, inst.selectedIndex)
		return true
	}
	if !inst.selectIndex(visibleRows[currentVisible-1], true, true) {
		return false
	}
	inst.emitNavigation("up", oldSelected, inst.selectedIndex)
	return true
}

func (inst *Instance) navigateDown() bool {
	visibleRows := inst.visibleRowIndices()
	if len(visibleRows) == 0 {
		return false
	}
	oldSelected := inst.selectedIndex
	currentVisible := inst.visibleRowPosition(visibleRows, inst.selectedIndex)
	if currentVisible < 0 {
		if !inst.selectIndex(visibleRows[0], true, true) {
			return false
		}
		inst.emitNavigation("down", oldSelected, inst.selectedIndex)
		return true
	}
	targetVisible := min(len(visibleRows)-1, currentVisible+1)
	if !inst.selectIndex(visibleRows[targetVisible], true, true) {
		return false
	}
	inst.emitNavigation("down", oldSelected, inst.selectedIndex)
	return true
}

func (inst *Instance) navigateHome() bool {
	visibleRows := inst.visibleRowIndices()
	if len(visibleRows) == 0 {
		return false
	}
	targetIndex := visibleRows[0]
	if inst.selectedIndex == targetIndex && inst.scrollOffset == 0 {
		return false
	}
	oldSelected := inst.selectedIndex
	oldScroll := inst.scrollOffset
	inst.scrollOffset = 0
	inst.selectedIndex = targetIndex
	inst.recordPendingScroll()
	inst.recordPendingSelected()
	inst.dirty = true
	inst.emitSelectionChanged()
	inst.emitStateChanged()
	inst.emitRowSelect(inst.selectedIndex)
	inst.emitNavigation("home", oldSelected, inst.selectedIndex)
	inst.emitScrollIntent(inst.scrollOffset - oldScroll)
	inst.emitSearchStats()
	return true
}

func (inst *Instance) navigateEnd() bool {
	visibleRows := inst.visibleRowIndices()
	if len(visibleRows) == 0 {
		return false
	}
	lastIndex := visibleRows[len(visibleRows)-1]
	maxOffset := inst.dataViewportFor(len(visibleRows)).MaxOffset()
	if inst.selectedIndex == lastIndex && inst.scrollOffset == maxOffset {
		return false
	}
	oldSelected := inst.selectedIndex
	oldScroll := inst.scrollOffset
	inst.scrollOffset = maxOffset
	inst.selectedIndex = lastIndex
	inst.recordPendingScroll()
	inst.recordPendingSelected()
	inst.dirty = true
	inst.emitSelectionChanged()
	inst.emitStateChanged()
	inst.emitRowSelect(inst.selectedIndex)
	inst.emitNavigation("end", oldSelected, inst.selectedIndex)
	inst.emitScrollIntent(inst.scrollOffset - oldScroll)
	inst.emitSearchStats()
	return true
}

func (inst *Instance) pageUp() bool {
	visibleRows := inst.visibleRowIndices()
	if len(visibleRows) == 0 {
		return false
	}
	viewport := inst.dataViewportFor(len(visibleRows))
	currentVisible := inst.visibleRowPosition(visibleRows, inst.selectedIndex)
	if viewport.Offset == 0 && currentVisible <= 0 {
		return false
	}
	oldSelected := inst.selectedIndex
	oldScroll := inst.scrollOffset
	viewport.PageUp()
	inst.scrollOffset = viewport.Offset
	if currentVisible < 0 {
		inst.selectedIndex = visibleRows[0]
	} else {
		targetVisible := max(0, currentVisible-viewport.ViewSize)
		inst.selectedIndex = visibleRows[targetVisible]
	}
	inst.recordPendingScroll()
	inst.recordPendingSelected()
	inst.dirty = true
	inst.emitSelectionChanged()
	inst.emitStateChanged()
	inst.emitRowSelect(inst.selectedIndex)
	inst.emitNavigation("pageup", oldSelected, inst.selectedIndex)
	inst.emitScrollIntent(inst.scrollOffset - oldScroll)
	inst.emitSearchStats()
	return true
}

func (inst *Instance) pageDown() bool {
	visibleRows := inst.visibleRowIndices()
	if len(visibleRows) == 0 {
		return false
	}
	viewport := inst.dataViewportFor(len(visibleRows))
	lastVisible := len(visibleRows) - 1
	currentVisible := inst.visibleRowPosition(visibleRows, inst.selectedIndex)
	if viewport.Offset == viewport.MaxOffset() && currentVisible >= lastVisible {
		return false
	}
	oldSelected := inst.selectedIndex
	oldScroll := inst.scrollOffset
	viewport.PageDown()
	inst.scrollOffset = viewport.Offset
	if currentVisible < 0 {
		targetVisible := min(lastVisible, max(0, viewport.ViewSize-1))
		inst.selectedIndex = visibleRows[targetVisible]
	} else {
		targetVisible := min(lastVisible, currentVisible+viewport.ViewSize)
		inst.selectedIndex = visibleRows[targetVisible]
	}
	inst.recordPendingScroll()
	inst.recordPendingSelected()
	inst.dirty = true
	inst.emitSelectionChanged()
	inst.emitStateChanged()
	inst.emitRowSelect(inst.selectedIndex)
	inst.emitNavigation("pagedown", oldSelected, inst.selectedIndex)
	inst.emitScrollIntent(inst.scrollOffset - oldScroll)
	inst.emitSearchStats()
	return true
}

func (inst *Instance) navigateMatch(direction int) bool {
	if direction == 0 || !inst.searchActive() {
		return false
	}
	visibleRows := inst.visibleRowIndices()
	if len(visibleRows) == 0 {
		return false
	}
	oldSelected := inst.selectedIndex
	currentVisible := inst.visibleRowPosition(visibleRows, inst.selectedIndex)
	if currentVisible < 0 {
		if direction > 0 {
			if !inst.selectIndex(visibleRows[0], true, true) {
				return false
			}
			inst.emitNavigation("searchnext", oldSelected, inst.selectedIndex)
			return true
		}
		if !inst.selectIndex(visibleRows[len(visibleRows)-1], true, true) {
			return false
		}
		inst.emitNavigation("searchprev", oldSelected, inst.selectedIndex)
		return true
	}
	nextVisible := currentVisible + direction
	for nextVisible < 0 {
		nextVisible += len(visibleRows)
	}
	nextVisible %= len(visibleRows)
	if !inst.selectIndex(visibleRows[nextVisible], true, true) {
		return false
	}
	if direction > 0 {
		inst.emitNavigation("searchnext", oldSelected, inst.selectedIndex)
	} else {
		inst.emitNavigation("searchprev", oldSelected, inst.selectedIndex)
	}
	return true
}

// =============================================================================
// Helper Methods
// =============================================================================

// ensureSelectedRowVisible ensures the selected row is visible
func (inst *Instance) ensureSelectedRowVisible() {
	visibleRows := inst.visibleRowIndices()
	if len(visibleRows) == 0 {
		return
	}
	visibleIndex := inst.visibleRowPosition(visibleRows, inst.selectedIndex)
	if visibleIndex < 0 {
		return
	}
	viewport := inst.dataViewportFor(len(visibleRows))
	if viewport.EnsureVisible(visibleIndex) {
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

func (inst *Instance) normalizeSelectionAndScroll() {
	visibleRows := inst.visibleRowIndices()
	searchActive := inst.searchActive()
	if len(visibleRows) == 0 {
		if searchActive && !inst.selectedIndexControlled {
			inst.selectedIndex = -1
		}
		inst.scrollOffset = 0
		inst.emitSearchStats()
		return
	}

	visibleIndex := inst.visibleRowPosition(visibleRows, inst.selectedIndex)
	if searchActive && visibleIndex < 0 && !inst.selectedIndexControlled {
		inst.selectedIndex = visibleRows[0]
		visibleIndex = 0
	}

	viewport := inst.dataViewportFor(len(visibleRows))
	if visibleIndex >= 0 && inst.pendingEnsureVisible {
		viewport.EnsureVisible(visibleIndex)
		inst.pendingEnsureVisible = false
	}
	inst.scrollOffset = viewport.Offset

	inst.emitSearchStats()
}

func (inst *Instance) scrollBy(delta int) bool {
	visibleRows := inst.visibleRowIndices()
	viewport := inst.dataViewportFor(len(visibleRows))
	oldOffset := inst.scrollOffset
	if !viewport.ScrollBy(delta) {
		return false
	}
	inst.scrollOffset = viewport.Offset
	inst.recordPendingScroll()
	inst.dirty = true

	// Clamp selectedIndex into the new visible viewport range, mirroring keyboard scroll behaviour.
	// Only clamp when there is an active selection (selectedIndex >= 0) and it drifts out of view.
	oldSelected := inst.selectedIndex
	if len(visibleRows) > 0 && inst.selectedIndex >= 0 {
		startRow, endRow := viewport.VisibleRange()
		currentVisible := inst.visibleRowPosition(visibleRows, inst.selectedIndex)
		if currentVisible >= 0 && currentVisible < startRow {
			inst.selectedIndex = visibleRows[startRow]
		} else if currentVisible >= endRow {
			inst.selectedIndex = visibleRows[endRow-1]
		}
		if inst.selectedIndex != oldSelected {
			inst.recordPendingSelected()
			inst.emitSelectionChanged()
			inst.emitRowSelect(inst.selectedIndex)
			direction := "scroll-down"
			if delta < 0 {
				direction = "scroll-up"
			}
			inst.emitNavigation(direction, oldSelected, inst.selectedIndex)
		}
	}

	inst.emitStateChanged()
	inst.emitScrollIntent(inst.scrollOffset - oldOffset)
	inst.emitSearchStats()
	return true
}

func (inst *Instance) scrollTo(offset int, emitState bool) bool {
	viewport := inst.dataViewportFor(len(inst.visibleRowIndices()))
	oldOffset := inst.scrollOffset
	if !viewport.ScrollTo(offset) {
		return false
	}
	inst.scrollOffset = viewport.Offset
	inst.recordPendingScroll()
	inst.dirty = true
	if emitState {
		inst.emitStateChanged()
	}
	inst.emitScrollIntent(inst.scrollOffset - oldOffset)
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
	chrome += inst.statsHeight()
	if inst.showBorder {
		chrome += 2
	}
	if inst.header != "" {
		chrome++
		if inst.showSeparator && len(inst.visibleRowIndices()) > 0 {
			chrome++
		}
	}
	return chrome
}

func (inst *Instance) statsHeight() int {
	if inst.showSearchStats {
		return 1
	}
	return 0
}

func (inst *Instance) dataViewportFor(contentSize int) scrollutil.VerticalViewport {
	return scrollutil.NewVerticalViewport(contentSize, inst.effectiveVisibleHeight(), inst.scrollOffset)
}

func (inst *Instance) rowStyleFor(rowIndex int, rowText string, matched bool) style.Style {
	if rowIndex == inst.selectedIndex {
		return inst.selectedStyle
	}
	if matched && inst.matchStyle != (style.Style{}) {
		return inst.matchStyle
	}
	if inst.rowStyleFn != nil {
		return inst.rowStyleFn(rowIndex, rowText)
	}
	return inst.rowStyle
}

func (inst *Instance) searchActive() bool {
	return strings.TrimSpace(inst.searchQuery) != ""
}

func (inst *Instance) rowMatches(rowText, query string) bool {
	if query == "" {
		return true
	}
	if inst.searchFn != nil {
		return inst.searchFn(rowText, query)
	}
	return strings.Contains(strings.ToLower(rowText), strings.ToLower(query))
}

func (inst *Instance) visibleRowIndices() []int {
	query := strings.TrimSpace(inst.searchQuery)
	visible := make([]int, 0, len(inst.rows))
	for rowIndex, rowText := range inst.rows {
		if inst.rowMatches(rowText, query) {
			visible = append(visible, rowIndex)
		}
	}
	return visible
}

func (inst *Instance) visibleRowPosition(visibleRows []int, rowIndex int) int {
	for visibleIndex, sourceIndex := range visibleRows {
		if sourceIndex == rowIndex {
			return visibleIndex
		}
	}
	return -1
}

func (inst *Instance) matchStats() (total int, selected int) {
	if !inst.searchActive() {
		return 0, 0
	}
	visibleRows := inst.visibleRowIndices()
	total = len(visibleRows)
	if total == 0 {
		return 0, 0
	}
	position := inst.visibleRowPosition(visibleRows, inst.selectedIndex)
	if position >= 0 {
		selected = position + 1
	}
	return total, selected
}

func (inst *Instance) searchStatsLine() string {
	query := strings.TrimSpace(inst.searchQuery)
	total, selected := inst.matchStats()
	if query == "" {
		return "Search: --"
	}
	return fmt.Sprintf("Search: %q %d/%d", query, selected, total)
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
	visibleRows := inst.visibleRowIndices()

	// Check header width (using display width, not byte length)
	headerWidth := paint.StringWidth(inst.headerDisplayText())
	if headerWidth+4 > maxWidth {
		maxWidth = headerWidth + 4
	}

	if inst.showSearchStats {
		statsWidth := paint.StringWidth(inst.searchStatsLine())
		if statsWidth+4 > maxWidth {
			maxWidth = statsWidth + 4
		}
	}

	// Check rows width (using display width, not byte length)
	for _, rowIndex := range visibleRows {
		row := inst.rows[rowIndex]
		rowWidth := paint.StringWidth(inst.rowDisplayText(rowIndex, row))
		if rowWidth+4 > maxWidth {
			maxWidth = rowWidth + 4
		}
	}

	return maxWidth
}

func (inst *Instance) handleClick(act *action.Action) bool {
	mouseMsg, ok := act.Payload.(*runtimemsg.MouseMsg)
	if !ok || mouseMsg == nil {
		return inst.handleActivate()
	}
	rowIndex, ok := inst.rowIndexAtLocalY(mouseMsg.LocalY)
	if !ok {
		return false
	}
	selectedChanged := false
	if rowIndex != inst.selectedIndex {
		selectedChanged = inst.selectIndex(rowIndex, true, inst.selectionMode == SelectionNone)
	} else if inst.selectionMode == SelectionNone {
		inst.emitSelectionChanged()
		inst.emitStateChanged()
	}
	if inst.selectionMode != SelectionNone {
		beforeChecked := inst.GetCheckedIndices()
		handled := inst.applySelectionAtIndex(rowIndex, false)
		if handled && (selectedChanged || !equalInts(beforeChecked, inst.checkedIndices)) {
			inst.emitStateChanged()
		}
	}
	return true
}

func (inst *Instance) rowIndexAtLocalY(localY int) (int, bool) {
	visibleRows := inst.visibleRowIndices()
	if len(visibleRows) == 0 {
		return -1, false
	}
	dataStart := 0
	if inst.showBorder {
		dataStart++
	}
	dataStart += inst.statsHeight()
	if inst.header != "" {
		dataStart++
		if inst.showSeparator && len(visibleRows) > 0 {
			dataStart++
		}
	}
	relative := localY - dataStart
	if relative < 0 || relative >= inst.effectiveVisibleHeight() {
		return -1, false
	}
	viewport := inst.dataViewportFor(len(visibleRows))
	startRow, endRow := viewport.VisibleRange()
	visibleIndex := startRow + relative
	if visibleIndex < startRow || visibleIndex >= endRow || visibleIndex >= len(visibleRows) {
		return -1, false
	}
	return visibleRows[visibleIndex], true
}

func (inst *Instance) selectIndex(index int, emitSelection bool, emitState bool) bool {
	if len(inst.rows) == 0 {
		return false
	}
	clamped := max(0, min(len(inst.rows)-1, index))
	changed := inst.selectedIndex != clamped
	oldScroll := inst.scrollOffset
	inst.selectedIndex = clamped
	inst.ensureSelectedRowVisible()
	if changed {
		inst.recordPendingSelected()
		inst.recordPendingScroll()
		inst.dirty = true
		if emitSelection {
			inst.emitSelectionChanged()
		}
		if emitState {
			inst.emitStateChanged()
		}
		inst.emitRowSelect(inst.selectedIndex)
		inst.emitScrollIntent(inst.scrollOffset - oldScroll)
		inst.emitSearchStats()
	}
	return changed
}

func (inst *Instance) clearSelection(emitState bool) bool {
	if inst.selectedIndex == -1 {
		return false
	}
	inst.selectedIndex = -1
	inst.recordPendingSelected()
	inst.dirty = true
	inst.emitSelectionChanged()
	if emitState {
		inst.emitStateChanged()
	}
	inst.emitSearchStats()
	return true
}

func (inst *Instance) handleActivate() bool {
	visibleRows := inst.visibleRowIndices()
	if len(visibleRows) == 0 {
		return false
	}
	if inst.selectionMode == SelectionNone {
		if inst.selectedIndex < 0 || inst.visibleRowPosition(visibleRows, inst.selectedIndex) < 0 {
			return inst.selectIndex(visibleRows[0], true, true)
		}
		if inst.selectedIndex >= len(inst.rows) {
			return false
		}
		inst.emitSelectionChanged()
		inst.emitStateChanged()
		return true
	}

	selectedChanged := false
	if inst.selectedIndex < 0 || inst.visibleRowPosition(visibleRows, inst.selectedIndex) < 0 {
		selectedChanged = inst.selectIndex(visibleRows[0], true, false)
	}
	if inst.selectedIndex < 0 || inst.selectedIndex >= len(inst.rows) {
		return false
	}
	beforeChecked := inst.GetCheckedIndices()
	handled := inst.applySelectionAtIndex(inst.selectedIndex, false)
	if handled && (selectedChanged || !equalInts(beforeChecked, inst.checkedIndices)) {
		inst.emitStateChanged()
	}
	return handled || selectedChanged
}

func (inst *Instance) toggleCheckedAtIndex(index int, emitState bool) bool {
	if index < 0 || index >= len(inst.rows) || inst.selectionMode == SelectionNone {
		return false
	}
	selectedChanged := false
	if inst.selectedIndex != index {
		selectedChanged = inst.selectIndex(index, true, false)
	}
	beforeChecked := inst.GetCheckedIndices()
	handled := inst.applySelectionAtIndex(index, false)
	if handled && emitState && (selectedChanged || !equalInts(beforeChecked, inst.checkedIndices)) {
		inst.emitStateChanged()
	}
	return handled || selectedChanged
}

func (inst *Instance) clearCheckedSelection(emitState bool) bool {
	if len(inst.checkedIndices) == 0 {
		return false
	}
	inst.checkedIndices = nil
	inst.recordPendingChecked()
	inst.dirty = true
	inst.emitCheckedSelectionChanged()
	if emitState {
		inst.emitStateChanged()
	}
	return true
}

func (inst *Instance) applySelectionAtIndex(index int, emitState bool) bool {
	if index < 0 || index >= len(inst.rows) || inst.selectionMode == SelectionNone {
		return false
	}
	changed := false
	switch inst.selectionMode {
	case SelectionSingle:
		changed = !equalInts(inst.checkedIndices, []int{index})
		inst.checkedIndices = []int{index}
	case SelectionMultiple:
		if inst.isChecked(index) {
			inst.checkedIndices = removeInt(inst.checkedIndices, index)
		} else {
			inst.checkedIndices = append(inst.checkedIndices, index)
		}
		sort.Ints(inst.checkedIndices)
		changed = true
	}
	inst.normalizeCheckedIndices()
	if changed {
		inst.recordPendingChecked()
		inst.dirty = true
		inst.emitCheckedSelectionChanged()
		if emitState {
			inst.emitStateChanged()
		}
	}
	return true
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

func (inst *Instance) emitCheckedSelectionChanged() {
	if inst.componentID != "" {
		inst.emitLocalIntent(SelectionChangeWithID(inst.componentID, inst.selectionMode, inst.checkedIndices, inst.checkedRowsText()))
	} else {
		inst.emitLocalIntent(SelectionChange(inst.selectionMode, inst.checkedIndices, inst.checkedRowsText()))
	}
	if inst.intentEmitter == nil {
		return
	}
	value := inst.checkedSelectionValue()
	if inst.formID != "" {
		if inst.selectionIntentField != nil {
			intent.Emit(inst, form.FieldChange(inst.formID, inst.selectionIntentField.GetField(), value, true))
		}
		return
	}
	if inst.selectionIntentField != nil {
		inst.intentEmitter(intent.FieldChangeIntent{
			Field: inst.selectionIntentField.GetField(),
			Value: value,
		})
		return
	}
	if inst.selectionIntent != nil {
		inst.intentEmitter(inst.selectionIntent)
	}
}

func (inst *Instance) emitStateChanged() {
	if inst.intentEmitter == nil || inst.componentID == "" {
		return
	}
	visibleRows := inst.visibleRowCount()
	inst.intentEmitter(StateChange(
		inst.componentID,
		inst.selectedIndex,
		inst.selectedRowText(),
		inst.scrollOffset,
		inst.effectiveVisibleHeight(),
		visibleRows,
		len(inst.rows),
		inst.selectionMode,
		inst.checkedIndices,
		inst.checkedRowsText(),
	))
}

func (inst *Instance) checkedSelectionValue() string {
	if len(inst.checkedIndices) == 0 {
		return ""
	}
	parts := make([]string, len(inst.checkedIndices))
	for index, value := range inst.checkedIndices {
		parts[index] = strconv.Itoa(value)
	}
	if inst.selectionMode == SelectionSingle {
		return parts[0]
	}
	return strings.Join(parts, ",")
}

func (inst *Instance) headerDisplayText() string {
	if inst.selectionMode == SelectionNone {
		return inst.header
	}
	return strings.Repeat(" ", inst.selectionMarkerWidth()) + inst.header
}

func (inst *Instance) emptyDisplayText() string {
	text := inst.emptyText
	if inst.searchActive() {
		text = "(no matches)"
	}
	if inst.selectionMode == SelectionNone {
		return text
	}
	return strings.Repeat(" ", inst.selectionMarkerWidth()) + text
}

func (inst *Instance) rowDisplayText(rowIndex int, rowText string) string {
	if inst.selectionMode == SelectionNone {
		return rowText
	}
	return inst.selectionMarker(rowIndex) + rowText
}

func (inst *Instance) selectionMarkerWidth() int {
	if inst.selectionMode == SelectionNone {
		return 0
	}
	return 4
}

func (inst *Instance) selectionMarker(rowIndex int) string {
	if inst.isChecked(rowIndex) {
		return "[x] "
	}
	return "[ ] "
}

func (inst *Instance) isChecked(rowIndex int) bool {
	for _, checkedIndex := range inst.checkedIndices {
		if checkedIndex == rowIndex {
			return true
		}
	}
	return false
}

func (inst *Instance) normalizeCheckedIndices() {
	if inst.selectionMode == SelectionNone || len(inst.rows) == 0 {
		inst.checkedIndices = nil
		return
	}
	normalized := make([]int, 0, len(inst.checkedIndices))
	seen := make(map[int]struct{}, len(inst.checkedIndices))
	for _, checkedIndex := range inst.checkedIndices {
		if checkedIndex < 0 || checkedIndex >= len(inst.rows) {
			continue
		}
		if _, exists := seen[checkedIndex]; exists {
			continue
		}
		seen[checkedIndex] = struct{}{}
		normalized = append(normalized, checkedIndex)
	}
	sort.Ints(normalized)
	if inst.selectionMode == SelectionSingle && len(normalized) > 1 {
		normalized = normalized[:1]
	}
	inst.checkedIndices = normalized
}

func (inst *Instance) selectedRowText() string {
	if inst.selectedIndex < 0 || inst.selectedIndex >= len(inst.rows) {
		return ""
	}
	return inst.rows[inst.selectedIndex]
}

func (inst *Instance) checkedRowsText() []string {
	if len(inst.checkedIndices) == 0 {
		return nil
	}
	rows := make([]string, 0, len(inst.checkedIndices))
	for _, checkedIndex := range inst.checkedIndices {
		if checkedIndex < 0 || checkedIndex >= len(inst.rows) {
			continue
		}
		rows = append(rows, inst.rows[checkedIndex])
	}
	return rows
}

func (inst *Instance) visibleRowCount() int {
	visibleRows := inst.visibleRowIndices()
	if len(visibleRows) == 0 {
		return 0
	}
	startRow, endRow := inst.dataViewportFor(len(visibleRows)).VisibleRange()
	if endRow < startRow {
		return 0
	}
	return endRow - startRow
}

// =============================================================================
// Getters
// =============================================================================

func (inst *Instance) GetScrollOffset() int   { return inst.scrollOffset }
func (inst *Instance) GetSelectedIndex() int  { return inst.selectedIndex }
func (inst *Instance) GetViewportHeight() int { return inst.viewportHeight }
func (inst *Instance) GetComponentID() string { return inst.componentID }
func (inst *Instance) GetItems() []RowItem    { return cloneItems(inst.items) }
func (inst *Instance) GetRows() []string      { return append([]string(nil), inst.rows...) }
func (inst *Instance) GetCheckedIndices() []int {
	return append([]int(nil), inst.checkedIndices...)
}
func (inst *Instance) GetSelectionMode() SelectionMode { return inst.selectionMode }
func (inst *Instance) GetSelectedRow() (string, bool) {
	if inst.selectedIndex < 0 || inst.selectedIndex >= len(inst.rows) {
		return "", false
	}
	return inst.rows[inst.selectedIndex], true
}
func (inst *Instance) SelectIndex(index int) bool {
	return inst.selectIndex(index, true, true)
}
func (inst *Instance) ToggleSelectionAt(index int) bool {
	return inst.applySelectionAtIndex(index, true)
}

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

func getItemsProp(props rtui.Props, def []RowItem) []RowItem {
	if value, ok := props[propItems]; ok {
		if items, ok := value.([]RowItem); ok {
			return cloneItems(items)
		}
	}
	return cloneItems(def)
}

func getStringsProp(props rtui.Props, def []string) []string {
	v, ok := props[propRows]
	if !ok {
		return def
	}
	if rows, ok := v.([]string); ok {
		return rows
	}
	return def
}

func getListRowsAndItems(props rtui.Props, currentItems []RowItem, currentRows []string) ([]RowItem, []string) {
	if _, ok := props[propItems]; ok {
		return normalizeItemsAndRows(getItemsProp(props, currentItems), nil)
	}
	if _, ok := props[propRows]; ok {
		return normalizeItemsAndRows(nil, getStringsProp(props, currentRows))
	}
	currentItems = cloneItems(currentItems)
	currentRows = append([]string(nil), currentRows...)
	return currentItems, currentRows
}

func getIntsProp(props rtui.Props, key string, def []int) []int {
	if value, ok := props[key]; ok {
		if ints, ok := value.([]int); ok {
			return append([]int(nil), ints...)
		}
	}
	return append([]int(nil), def...)
}

func getRuneProp(props rtui.Props, key string, def rune) rune {
	if v, ok := props[key]; ok {
		if r, ok := v.(rune); ok {
			return r
		}
	}
	return def
}

func getChangeIntentFieldProp(props rtui.Props, key string) intent.FieldIntent {
	if v, ok := props[key]; ok {
		if i, ok := v.(intent.FieldIntent); ok {
			return i
		}
	}
	return nil
}

func getSearchFn(props rtui.Props) func(string, string) bool {
	if value, ok := props[propSearchFn]; ok {
		if fn, ok := value.(func(string, string) bool); ok {
			return fn
		}
	}
	return nil
}

func getSearchFnOrCurrent(props rtui.Props, current func(string, string) bool) func(string, string) bool {
	if value, ok := props[propSearchFn]; ok {
		if fn, ok := value.(func(string, string) bool); ok {
			return fn
		}
		return nil
	}
	return current
}

func getSelectionModeProp(props rtui.Props, key string, def SelectionMode) SelectionMode {
	if value, ok := props[key]; ok {
		if mode, ok := value.(SelectionMode); ok {
			return mode
		}
	}
	return def
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

func equalInts(left, right []int) bool {
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

func removeInt(values []int, target int) []int {
	for index, value := range values {
		if value == target {
			return append(append([]int(nil), values[:index]...), values[index+1:]...)
		}
	}
	return append([]int(nil), values...)
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

func sameSearchFn(left, right func(string, string) bool) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return reflect.ValueOf(left).Pointer() == reflect.ValueOf(right).Pointer()
}

func hasProp(props rtui.Props, key string) bool {
	_, ok := props[key]
	return ok
}
