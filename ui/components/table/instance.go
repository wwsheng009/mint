package table

import (
	"github.com/wwsheng009/mint/ui/components/internal/proputil"
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
	scrollutil "github.com/wwsheng009/mint/ui/components/internal/scroll"
)

type rowView struct {
	sourceIndex int
	cells       []string
}

type tableView struct {
	rows          []rowView
	pageRows      []rowView
	totalCount    int
	filteredCount int
	currentPage   int
	pageCount     int
	start         int
	end           int
}

const (
	scrollbarReservedWidth = 2
	selectionColumnWidth   = 3
)

// Instance is the runtime entity for table components.
type Instance struct {
	key         string
	componentID string

	columns        []TableColumn
	rows           [][]string
	emptyText      string
	headerStyle    style.Style
	tableStyle     style.Style
	selectedStyle  style.Style
	borderStyle    style.Style
	statusStyle    style.Style
	filterStyle    style.Style
	scrollbarStyle style.Style
	gap            int
	showBorder     bool
	showFooter     bool
	showScrollbar  bool
	pageSize       int
	searchQuery    string
	filters        map[int]string

	currentPage              int
	currentPageControlled    bool
	sortColumn               int
	sortDescending           bool
	sortControlled           bool
	selectedIndex            int
	selectedIndexControlled  bool
	selectionIntent          intent.Intent
	selectionIntentField     intent.FieldIntent
	selectionMode            SelectionMode
	checkedIndices           []int
	checkedIndicesControlled bool
	lastPropCurrentPage      int
	lastPropSortColumn       int
	lastPropSortDescending   bool

	changeIntent      intent.Intent
	changeIntentField intent.FieldIntent
	pageIntentField   intent.FieldIntent

	pendingCurrentPage    int
	hasPendingCurrentPage bool
	pendingSortColumn     int
	pendingSortDescending bool
	hasPendingSort        bool

	parent        rtui.ComponentInstance
	intentEmitter func(intent.Intent)
	focused       bool
	bounds        [4]int
	dirty         bool
}

var (
	_ rtui.ComponentInstance     = (*Instance)(nil)
	_ rtui.PaintableInstance     = (*Instance)(nil)
	_ rtui.FocusableInstance     = (*Instance)(nil)
	_ rtui.ActionHandlerInstance = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:                      proputil.GetString(props, "key", ""),
		componentID:              proputil.GetString(props, "componentID", ""),
		columns:                  getColumnsProp(props, []TableColumn{}),
		rows:                     getRowsProp(props, [][]string{}),
		emptyText:                proputil.GetString(props, "emptyText", "(empty)"),
		headerStyle:              proputil.GetStyle(props, "headerStyle", style.Style{}),
		tableStyle:               proputil.GetStyle(props, "tableStyle", style.Style{}),
		selectedStyle:            proputil.GetStyle(props, "selectedStyle", style.Style{}.Reverse(true)),
		borderStyle:              proputil.GetStyle(props, "borderStyle", style.Style{}.Foreground(style.BrightBlack)),
		statusStyle:              proputil.GetStyle(props, "statusStyle", style.Style{}.Foreground(style.BrightBlack)),
		filterStyle:              proputil.GetStyle(props, "filterStyle", style.Style{}.Foreground(style.BrightBlack)),
		scrollbarStyle:           proputil.GetStyle(props, "scrollbarStyle", style.Style{}.Foreground(style.BrightBlack)),
		gap:                      maxInt(0, proputil.GetInt(props, "gap", 0)),
		showBorder:               proputil.GetBool(props, "showBorder", false),
		showFooter:               proputil.GetBool(props, "showFooter", true),
		showScrollbar:            proputil.GetBool(props, "showScrollbar", true),
		pageSize:                 maxInt(0, proputil.GetInt(props, "pageSize", 0)),
		searchQuery:              proputil.GetString(props, "searchQuery", ""),
		filters:                  getFiltersProp(props, map[int]string{}),
		currentPage:              maxInt(0, proputil.GetInt(props, "currentPage", 0)),
		currentPageControlled:    proputil.GetBool(props, "currentPageControlled", false),
		sortColumn:               proputil.GetInt(props, "sortColumn", -1),
		sortDescending:           proputil.GetBool(props, "sortDescending", false),
		sortControlled:           proputil.GetBool(props, "sortControlled", false),
		selectedIndex:            proputil.GetInt(props, "selectedIndex", -1),
		selectedIndexControlled:  proputil.GetBool(props, "selectedIndexControlled", false),
		selectionIntent:          proputil.GetIntent(props, "selectionIntent", nil),
		selectionIntentField:     getFieldIntentProp(props, "selectionIntent"),
		selectionMode:            getSelectionModeProp(props, "selectionMode", SelectionNone),
		checkedIndices:           getIntsProp(props, "checkedIndices", nil),
		checkedIndicesControlled: proputil.GetBool(props, "checkedIndicesControlled", false),
		lastPropCurrentPage:      maxInt(0, proputil.GetInt(props, "currentPage", 0)),
		lastPropSortColumn:       proputil.GetInt(props, "sortColumn", -1),
		lastPropSortDescending:   proputil.GetBool(props, "sortDescending", false),
		changeIntent:             proputil.GetIntent(props, "changeIntent", nil),
		changeIntentField:        getFieldIntentProp(props, "changeIntent"),
		pageIntentField:          getFieldIntentProp(props, "pageIntent"),
		dirty:                    true,
	}
	inst.normalizeViewState(false)
	inst.normalizeCheckedIndices()
	return inst
}

func (inst *Instance) Key() string           { return inst.key }
func (inst *Instance) SetKey(key string)     { inst.key = key }
func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *Instance) Destroy()              { inst.columns = nil; inst.rows = nil; inst.filters = nil }
func (inst *Instance) OnMount()              { inst.dirty = true }
func (inst *Instance) OnUnmount()            {}
func (inst *Instance) Parent() interface{}   { return inst.parent }
func (inst *Instance) SetParent(parent rtui.ComponentInstance) {
	inst.parent = parent
}

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldColumns := append([]TableColumn(nil), inst.columns...)
	oldRows := cloneRows(inst.rows)
	oldEmptyText := inst.emptyText
	oldHeaderStyle := inst.headerStyle
	oldTableStyle := inst.tableStyle
	oldSelectedStyle := inst.selectedStyle
	oldBorderStyle := inst.borderStyle
	oldStatusStyle := inst.statusStyle
	oldFilterStyle := inst.filterStyle
	oldScrollbarStyle := inst.scrollbarStyle
	oldGap := inst.gap
	oldShowBorder := inst.showBorder
	oldShowFooter := inst.showFooter
	oldShowScrollbar := inst.showScrollbar
	oldPageSize := inst.pageSize
	oldSearchQuery := inst.searchQuery
	oldFilters := cloneFilters(inst.filters)
	oldCurrentPage := inst.currentPage
	oldCurrentPageControlled := inst.currentPageControlled
	oldSortColumn := inst.sortColumn
	oldSortDescending := inst.sortDescending
	oldSortControlled := inst.sortControlled
	oldSelectedIndex := inst.selectedIndex
	oldSelectedIndexControlled := inst.selectedIndexControlled
	oldSelectionIntent := inst.selectionIntent
	oldSelectionMode := inst.selectionMode
	oldCheckedIndices := append([]int(nil), inst.checkedIndices...)
	oldCheckedIndicesControlled := inst.checkedIndicesControlled
	oldPropCurrentPage := inst.lastPropCurrentPage
	oldPropSortColumn := inst.lastPropSortColumn
	oldPropSortDescending := inst.lastPropSortDescending
	oldComponentID := inst.componentID
	oldChangeIntent := inst.changeIntent
	oldChangeIntentField := inst.changeIntentField
	oldPageIntentField := inst.pageIntentField

	inst.componentID = proputil.GetString(props, "componentID", inst.componentID)
	inst.columns = getColumnsProp(props, inst.columns)
	inst.rows = getRowsProp(props, inst.rows)
	inst.emptyText = proputil.GetString(props, "emptyText", inst.emptyText)
	inst.headerStyle = proputil.GetStyle(props, "headerStyle", inst.headerStyle)
	inst.tableStyle = proputil.GetStyle(props, "tableStyle", inst.tableStyle)
	inst.selectedStyle = proputil.GetStyle(props, "selectedStyle", inst.selectedStyle)
	inst.borderStyle = proputil.GetStyle(props, "borderStyle", inst.borderStyle)
	inst.statusStyle = proputil.GetStyle(props, "statusStyle", inst.statusStyle)
	inst.filterStyle = proputil.GetStyle(props, "filterStyle", inst.filterStyle)
	inst.scrollbarStyle = proputil.GetStyle(props, "scrollbarStyle", inst.scrollbarStyle)
	inst.gap = maxInt(0, proputil.GetInt(props, "gap", inst.gap))
	inst.showBorder = proputil.GetBool(props, "showBorder", inst.showBorder)
	inst.showFooter = proputil.GetBool(props, "showFooter", inst.showFooter)
	inst.showScrollbar = proputil.GetBool(props, "showScrollbar", inst.showScrollbar)
	inst.pageSize = maxInt(0, proputil.GetInt(props, "pageSize", inst.pageSize))
	inst.searchQuery = proputil.GetString(props, "searchQuery", inst.searchQuery)
	inst.filters = getFiltersProp(props, inst.filters)
	inst.changeIntent = proputil.GetIntent(props, "changeIntent", nil)
	inst.changeIntentField = getFieldIntentProp(props, "changeIntent")
	inst.pageIntentField = getFieldIntentProp(props, "pageIntent")
	inst.selectionIntent = proputil.GetIntent(props, "selectionIntent", nil)
	inst.selectionIntentField = getFieldIntentProp(props, "selectionIntent")
	inst.selectionMode = getSelectionModeProp(props, "selectionMode", inst.selectionMode)

	if controlled, ok := props[propCurrentPageControlled].(bool); ok {
		inst.currentPageControlled = controlled
	}
	if inst.currentPageControlled {
		inst.currentPage = maxInt(0, proputil.GetInt(props, "currentPage", inst.currentPage))
		inst.lastPropCurrentPage = inst.currentPage
	}

	if controlled, ok := props[propSortControlled].(bool); ok {
		inst.sortControlled = controlled
	}
	if inst.sortControlled {
		inst.sortColumn = proputil.GetInt(props, "sortColumn", inst.sortColumn)
		inst.sortDescending = proputil.GetBool(props, "sortDescending", inst.sortDescending)
		inst.lastPropSortColumn = inst.sortColumn
		inst.lastPropSortDescending = inst.sortDescending
	}

	if controlled, ok := props[propSelectedIndexControlled].(bool); ok {
		inst.selectedIndexControlled = controlled
	}
	if inst.selectedIndexControlled {
		inst.selectedIndex = proputil.GetInt(props, "selectedIndex", inst.selectedIndex)
	}
	if controlled, ok := props[propCheckedIndicesControlled].(bool); ok {
		inst.checkedIndicesControlled = controlled
	}
	if inst.checkedIndicesControlled {
		inst.checkedIndices = getIntsProp(props, "checkedIndices", inst.checkedIndices)
	} else if checkedIndices, ok := props[propCheckedIndices].([]int); ok {
		inst.checkedIndices = append([]int(nil), checkedIndices...)
	}

	inst.reconcilePendingState(oldPropCurrentPage, oldPropSortColumn, oldPropSortDescending)

	resetPage := oldSearchQuery != inst.searchQuery || !equalFilters(oldFilters, inst.filters) || oldPageSize != inst.pageSize
	inst.normalizeViewState(resetPage)
	inst.normalizeCheckedIndices()

	changed := !columnsEqual(oldColumns, inst.columns) ||
		!rowsEqual(oldRows, inst.rows) ||
		oldEmptyText != inst.emptyText ||
		oldHeaderStyle != inst.headerStyle ||
		oldTableStyle != inst.tableStyle ||
		oldSelectedStyle != inst.selectedStyle ||
		oldBorderStyle != inst.borderStyle ||
		oldStatusStyle != inst.statusStyle ||
		oldFilterStyle != inst.filterStyle ||
		oldScrollbarStyle != inst.scrollbarStyle ||
		oldGap != inst.gap ||
		oldShowBorder != inst.showBorder ||
		oldShowFooter != inst.showFooter ||
		oldShowScrollbar != inst.showScrollbar ||
		oldPageSize != inst.pageSize ||
		oldSearchQuery != inst.searchQuery ||
		!equalFilters(oldFilters, inst.filters) ||
		oldCurrentPage != inst.currentPage ||
		oldCurrentPageControlled != inst.currentPageControlled ||
		oldSortColumn != inst.sortColumn ||
		oldSortDescending != inst.sortDescending ||
		oldSortControlled != inst.sortControlled ||
		oldSelectedIndex != inst.selectedIndex ||
		oldSelectedIndexControlled != inst.selectedIndexControlled ||
		!sameIntent(oldSelectionIntent, inst.selectionIntent) ||
		oldSelectionMode != inst.selectionMode ||
		oldCheckedIndicesControlled != inst.checkedIndicesControlled ||
		!equalInts(oldCheckedIndices, inst.checkedIndices) ||
		oldPropCurrentPage != inst.lastPropCurrentPage ||
		oldPropSortColumn != inst.lastPropSortColumn ||
		oldPropSortDescending != inst.lastPropSortDescending ||
		oldComponentID != inst.componentID ||
		!sameIntent(oldChangeIntent, inst.changeIntent) ||
		!sameFieldIntent(oldChangeIntentField, inst.changeIntentField) ||
		!sameFieldIntent(oldPageIntentField, inst.pageIntentField)
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	props := rtui.Props{
		propKey:                     inst.key,
		propColumns:                 inst.columns,
		propRows:                    inst.rows,
		propEmptyText:               inst.emptyText,
		propHeaderStyle:             inst.headerStyle,
		propTableStyle:              inst.tableStyle,
		propSelectedStyle:           inst.selectedStyle,
		propBorderStyle:             inst.borderStyle,
		propStatusStyle:             inst.statusStyle,
		propFilterStyle:             inst.filterStyle,
		propScrollbarStyle:          inst.scrollbarStyle,
		propGap:                     inst.gap,
		propShowBorder:              inst.showBorder,
		propShowFooter:              inst.showFooter,
		propShowScrollbar:           inst.showScrollbar,
		propPageSize:                inst.pageSize,
		propSearchQuery:             inst.searchQuery,
		propFilters:                 cloneFilters(inst.filters),
		propCurrentPage:             inst.currentPage,
		propCurrentPageControlled:   inst.currentPageControlled,
		propSortColumn:              inst.sortColumn,
		propSortDescending:          inst.sortDescending,
		propSortControlled:          inst.sortControlled,
		propSelectedIndex:           inst.selectedIndex,
		propSelectedIndexControlled: inst.selectedIndexControlled,
		propSelectionIntent:         inst.selectionIntent,
		propSelectionMode:           inst.selectionMode,
		propComponentID:             inst.componentID,
		propChangeIntent:            inst.changeIntent,
		propChangeIntentField:       inst.changeIntentField,
		propPageIntentField:         inst.pageIntentField,
	}
	if inst.selectionIntentField != nil {
		props[propSelectionIntentField] = inst.selectionIntentField
	}
	if inst.checkedIndicesControlled {
		props[propCheckedIndicesControlled] = true
		props[propCheckedIndices] = append([]int(nil), inst.checkedIndices...)
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

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if len(inst.columns) == 0 {
		return layout.Size{}
	}

	view := inst.processedView()
	_, contentWidth := inst.calculateColumnWidths(view.rows, 0)
	if inst.shouldShowFooter(view) {
		contentWidth = maxInt(contentWidth, paint.StringWidth(inst.statusText(view)))
	}
	innerWidth := inst.innerWidth(contentWidth, view, len(view.pageRows))
	width := innerWidth
	if inst.showBorder {
		width += 4
	}

	height := inst.lineCountForView(view)
	width = constraints.ConstrainWidth(width)
	height = constraints.ConstrainHeight(height)
	return layout.Size{Width: width, Height: height}
}

func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	if len(inst.columns) == 0 {
		return nil
	}

	view := inst.processedView()
	maxInnerWidth := inst.maxInnerPaintWidth()
	maxContentWidth := maxInnerWidth
	if maxInnerWidth > 0 && inst.showScrollbar {
		maxContentWidth = maxInt(1, maxInnerWidth-scrollbarReservedWidth)
	}
	widths, contentWidth := inst.calculateColumnWidths(view.rows, maxContentWidth)
	if contentWidth < 1 {
		contentWidth = 1
	}

	showFooter := inst.shouldShowFooter(view)
	statusText := inst.statusText(view)
	if showFooter {
		contentWidth = maxInt(contentWidth, paint.StringWidth(statusText))
		if maxContentWidth > 0 {
			contentWidth = minInt(contentWidth, maxContentWidth)
		}
	}
	renderedRows := view.pageRows
	if limit := inst.maxRenderableRows(showFooter); limit >= 0 && len(renderedRows) > limit {
		renderedRows = renderedRows[:limit]
	}
	showScrollbar := inst.shouldPaintScrollbar(view, len(renderedRows))
	innerWidth := inst.innerWidth(contentWidth, view, len(renderedRows))

	type lineSpec struct {
		text  string
		style style.Style
	}

	innerLines := make([]lineSpec, 0, inst.lineCountForView(view))
	innerLines = append(innerLines, lineSpec{
		text:  inst.composeInnerLine(inst.buildHeaderLine(widths, view), contentWidth, showScrollbar),
		style: inst.headerStyle,
	})
	if inst.hasActiveColumnFilters() {
		innerLines = append(innerLines, lineSpec{
			text:  inst.composeInnerLine(inst.buildFilterLine(widths), contentWidth, showScrollbar),
			style: inst.filterStyle,
		})
	}
	innerLines = append(innerLines, lineSpec{text: inst.composeInnerLine(strings.Repeat("─", maxInt(1, contentWidth)), contentWidth, showScrollbar), style: inst.borderStyle})
	for i := 0; i < inst.gap; i++ {
		innerLines = append(innerLines, lineSpec{text: inst.composeInnerLine("", contentWidth, showScrollbar), style: inst.tableStyle})
	}

	if len(renderedRows) == 0 {
		innerLines = append(innerLines, lineSpec{
			text:  inst.composeInnerLine(inst.emptyLineText(), contentWidth, showScrollbar),
			style: inst.tableStyle,
		})
	} else {
		for rowOffset, row := range renderedRows {
			absoluteIndex := view.start + rowOffset
			innerLines = append(innerLines, lineSpec{
				text:  inst.composeInnerLine(inst.buildDataLine(row, widths), contentWidth, showScrollbar),
				style: inst.rowStyleFor(absoluteIndex),
			})
		}
	}

	if showFooter {
		innerLines = append(innerLines, lineSpec{
			text:  inst.composeInnerLine(statusText, contentWidth, showScrollbar),
			style: inst.statusStyle,
		})
	}

	lines := make([]lineSpec, 0, len(innerLines)+2)
	if inst.showBorder {
		borderStyle := inst.borderStyle
		if inst.focused {
			borderStyle = borderStyle.Bold(true)
		}
		lines = append(lines, lineSpec{
			text:  "┌" + strings.Repeat("─", innerWidth+2) + "┐",
			style: borderStyle,
		})
		for _, line := range innerLines {
			lines = append(lines, lineSpec{
				text:  "│ " + padRightToWidth(truncateText(line.text, innerWidth), innerWidth) + " │",
				style: line.style,
			})
		}
		lines = append(lines, lineSpec{
			text:  "└" + strings.Repeat("─", innerWidth+2) + "┘",
			style: borderStyle,
		})
	} else {
		lines = innerLines
	}

	if maxHeight := inst.maxPaintHeight(); maxHeight > 0 && len(lines) > maxHeight {
		lines = lines[:maxHeight]
	}

	cmds := make([]paint.DrawCmd, 0, len(lines))
	for lineIndex, line := range lines {
		cmds = append(cmds, paint.NewTextCmd(x, y+lineIndex, line.text, line.style))
	}

	if showScrollbar {
		viewport := scrollutil.NewVerticalViewport(view.filteredCount, maxInt(1, len(renderedRows)), view.start)
		scrollbarX := x + contentWidth + 1
		scrollbarY := y + 2 + inst.gap
		if inst.showBorder {
			scrollbarX += 2
			scrollbarY++
		}
		scrollbarStyle := inst.scrollbarStyle
		if scrollbarStyle.FG == "" {
			scrollbarStyle = scrollbarStyle.Foreground(inst.borderStyle.FG)
		}
		cmds = append(cmds, scrollutil.DrawVerticalScrollbar(
			scrollbarX,
			scrollbarY,
			maxInt(1, len(renderedRows)),
			viewport,
			scrollbarStyle,
			scrollutil.DefaultVerticalScrollbarConfig(),
		)...)
	}
	return cmds
}

func (inst *Instance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

func (inst *Instance) HandleAction(act *action.Action) bool {
	if act == nil || len(inst.columns) == 0 {
		return false
	}

	switch act.Type {
	case action.ActionClick:
		return inst.handleClick(act)
	case action.ActionNavigateUp:
		return inst.navigateUp()
	case action.ActionNavigateDown:
		return inst.navigateDown()
	case action.ActionNavigateHome:
		return inst.navigateHome()
	case action.ActionNavigateEnd:
		return inst.navigateEnd()
	case action.ActionNavigatePageUp, action.ActionNavigateLeft:
		return inst.movePage(-1)
	case action.ActionNavigatePageDown, action.ActionNavigateRight:
		return inst.movePage(1)
	case action.ActionScroll:
		delta, ok := scrollutil.DeltaFromAction(act)
		if !ok || delta == 0 {
			return false
		}
		if delta > 0 {
			return inst.movePage(1)
		}
		return inst.movePage(-1)
	case action.ActionSelect, action.ActionEnter:
		if inst.selectedIndex < 0 {
			if !inst.selectIndex(0) {
				return false
			}
		}
		if inst.selectionMode != SelectionNone {
			return inst.handleActivateSelection()
		}
		return len(inst.filteredSortedRows()) > 0
	case action.ActionSelectAll:
		return inst.selectAllFilteredRows()
	case action.ActionClear:
		return inst.clearCheckedSelection()
	}
	return false
}

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

func (inst *Instance) normalizeViewState(resetPage bool) {
	if inst.pageSize < 0 {
		inst.pageSize = 0
	}
	if inst.sortColumn >= len(inst.columns) {
		inst.sortColumn = -1
	}
	if inst.hasPendingSort && inst.pendingSortColumn >= len(inst.columns) {
		inst.pendingSortColumn = -1
		inst.pendingSortDescending = false
	}
	if inst.currentPage < 0 {
		inst.currentPage = 0
	}
	if inst.selectedIndex < -1 {
		inst.selectedIndex = -1
	}
	if resetPage && !inst.currentPageControlled {
		inst.currentPage = 0
	}
	if resetPage && !inst.selectedIndexControlled {
		inst.selectedIndex = -1
	}

	rows := inst.filteredSortedRows()
	if len(rows) == 0 {
		if !inst.currentPageControlled {
			inst.currentPage = 0
		}
		if !inst.selectedIndexControlled {
			inst.selectedIndex = -1
		}
		return
	}

	if inst.selectedIndex >= len(rows) {
		inst.selectedIndex = len(rows) - 1
	}
	if inst.pageSize <= 0 {
		if !inst.currentPageControlled {
			inst.currentPage = 0
		}
		return
	}

	pageCount := pageCountFor(len(rows), inst.pageSize)
	if inst.currentPage >= pageCount {
		inst.currentPage = pageCount - 1
	}
	if inst.currentPage < 0 {
		inst.currentPage = 0
	}
	if !inst.currentPageControlled && inst.selectedIndex >= 0 {
		inst.currentPage = inst.selectedIndex / inst.pageSize
	}
}

func (inst *Instance) processedView() tableView {
	rows := inst.filteredSortedRows()
	view := tableView{
		rows:          rows,
		pageRows:      rows,
		totalCount:    len(inst.rows),
		filteredCount: len(rows),
		currentPage:   0,
		pageCount:     1,
		start:         0,
		end:           len(rows),
	}
	if inst.pageSize <= 0 {
		return view
	}

	pageCount := pageCountFor(len(rows), inst.pageSize)
	page := clampInt(inst.effectiveCurrentPage(), 0, maxInt(0, pageCount-1))
	start := page * inst.pageSize
	end := minInt(start+inst.pageSize, len(rows))
	if start > len(rows) {
		start = len(rows)
	}
	if end < start {
		end = start
	}
	view.currentPage = page
	view.pageCount = pageCount
	view.start = start
	view.end = end
	view.pageRows = rows[start:end]
	return view
}

func (inst *Instance) filteredSortedRows() []rowView {
	if len(inst.columns) == 0 {
		return nil
	}

	searchTerm := strings.ToLower(strings.TrimSpace(inst.searchQuery))
	rows := make([]rowView, 0, len(inst.rows))
	for rowIndex, row := range inst.rows {
		cells := inst.normalizeRow(row)
		if !inst.matchesFilters(cells) {
			continue
		}
		if searchTerm != "" && !inst.matchesSearch(cells, searchTerm) {
			continue
		}
		rows = append(rows, rowView{
			sourceIndex: rowIndex,
			cells:       cells,
		})
	}

	sortColumn, descending := inst.effectiveSortState()
	if sortColumn >= 0 && sortColumn < len(inst.columns) {
		sort.SliceStable(rows, func(i, j int) bool {
			comparison := compareCellValues(rows[i].cells[sortColumn], rows[j].cells[sortColumn])
			if comparison == 0 {
				return rows[i].sourceIndex < rows[j].sourceIndex
			}
			if descending {
				return comparison > 0
			}
			return comparison < 0
		})
	}
	return rows
}

func (inst *Instance) normalizeRow(row []string) []string {
	cells := make([]string, len(inst.columns))
	for index := range inst.columns {
		if index < len(row) {
			cells[index] = row[index]
		}
	}
	return cells
}

func (inst *Instance) matchesFilters(cells []string) bool {
	for columnIndex, rawFilter := range inst.filters {
		if columnIndex < 0 || columnIndex >= len(cells) {
			continue
		}
		filter := strings.ToLower(strings.TrimSpace(rawFilter))
		if filter == "" {
			continue
		}
		if !strings.Contains(strings.ToLower(cells[columnIndex]), filter) {
			return false
		}
	}
	return true
}

func (inst *Instance) matchesSearch(cells []string, searchTerm string) bool {
	for _, cell := range cells {
		if strings.Contains(strings.ToLower(cell), searchTerm) {
			return true
		}
	}
	return false
}

func (inst *Instance) calculateColumnWidths(rows []rowView, maxInnerWidth int) ([]int, int) {
	widths := make([]int, len(inst.columns))
	totalContentWidth := 0
	for columnIndex, column := range inst.columns {
		width := column.Width
		if width <= 0 {
			width = paint.StringWidth(inst.headerLabel(columnIndex))
			for _, row := range rows {
				if columnIndex < len(row.cells) {
					width = maxInt(width, paint.StringWidth(row.cells[columnIndex]))
				}
			}
			width = maxInt(3, width)
		}
		width = maxInt(1, width)
		widths[columnIndex] = width
		totalContentWidth += width
	}

	separatorCount := maxInt(0, len(inst.columns)-1)
	prefixWidth := inst.selectionPrefixWidth()
	if inst.selectionMode != SelectionNone {
		separatorCount++
	}
	separatorWidth := separatorCount * 3
	if maxInnerWidth > 0 && totalContentWidth+separatorWidth+prefixWidth > maxInnerWidth {
		widths = shrinkWidthsToFit(widths, maxInnerWidth-separatorWidth-prefixWidth)
		totalContentWidth = 0
		for _, width := range widths {
			totalContentWidth += width
		}
	}

	return widths, totalContentWidth + separatorWidth + prefixWidth
}

func (inst *Instance) lineCountForView(view tableView) int {
	if len(inst.columns) == 0 {
		return 0
	}
	lines := 2 + inst.gap
	if inst.hasActiveColumnFilters() {
		lines++
	}
	if len(view.pageRows) == 0 {
		lines++
	} else {
		lines += len(view.pageRows)
	}
	if inst.shouldShowFooter(view) {
		lines++
	}
	if inst.showBorder {
		lines += 2
	}
	return lines
}

func (inst *Instance) selectionPrefixWidth() int {
	if inst.selectionMode == SelectionNone {
		return 0
	}
	return selectionColumnWidth
}

func (inst *Instance) emptyLineText() string {
	if inst.selectionMode == SelectionNone {
		return inst.emptyText
	}
	return formatCell("", selectionColumnWidth, rtui.AlignStart) + " │ " + inst.emptyText
}

func (inst *Instance) hasActiveColumnFilters() bool {
	for _, rawFilter := range inst.filters {
		if strings.TrimSpace(rawFilter) != "" {
			return true
		}
	}
	return false
}

func (inst *Instance) selectionHeaderMarker(view tableView) string {
	if inst.selectionMode != SelectionMultiple {
		return "Sel"
	}
	total := len(view.rows)
	if total == 0 {
		return "[ ]"
	}
	checked := 0
	for _, row := range view.rows {
		if inst.isChecked(row.sourceIndex) {
			checked++
		}
	}
	switch {
	case checked == 0:
		return "[ ]"
	case checked == total:
		return "[x]"
	default:
		return "[-]"
	}
}

func (inst *Instance) buildFilterLine(widths []int) string {
	cells := make([]string, 0, len(inst.columns)+1)
	if inst.selectionMode != SelectionNone {
		cells = append(cells, formatCell("", selectionColumnWidth, rtui.AlignStart))
	}
	for columnIndex := range inst.columns {
		label := ""
		if rawFilter, ok := inst.filters[columnIndex]; ok {
			filter := strings.TrimSpace(rawFilter)
			if filter != "" {
				label = "~" + filter
			}
		}
		cells = append(cells, formatCell(label, widths[columnIndex], inst.columns[columnIndex].Align))
	}
	return strings.Join(cells, " │ ")
}

func (inst *Instance) shouldPaintScrollbar(view tableView, visibleRows int) bool {
	if !inst.showScrollbar {
		return false
	}
	return view.filteredCount > maxInt(1, visibleRows)
}

func (inst *Instance) innerWidth(contentWidth int, view tableView, visibleRows int) int {
	if contentWidth < 1 {
		contentWidth = 1
	}
	if inst.shouldPaintScrollbar(view, visibleRows) {
		return contentWidth + scrollbarReservedWidth
	}
	return contentWidth
}

func (inst *Instance) composeInnerLine(text string, contentWidth int, withScrollbar bool) string {
	line := padRightToWidth(truncateText(text, contentWidth), contentWidth)
	if withScrollbar {
		line += strings.Repeat(" ", scrollbarReservedWidth)
	}
	return line
}

func (inst *Instance) maxInnerPaintWidth() int {
	if inst.bounds[2] <= 0 {
		return 0
	}
	if inst.showBorder {
		return maxInt(1, inst.bounds[2]-4)
	}
	return maxInt(1, inst.bounds[2])
}

func (inst *Instance) maxPaintHeight() int {
	if inst.bounds[3] <= 0 {
		return 0
	}
	return maxInt(1, inst.bounds[3])
}

func (inst *Instance) maxRenderableRows(showFooter bool) int {
	if inst.bounds[3] <= 0 {
		return -1
	}
	chrome := 2 + inst.gap
	if inst.hasActiveColumnFilters() {
		chrome++
	}
	if showFooter {
		chrome++
	}
	if inst.showBorder {
		chrome += 2
	}
	available := inst.bounds[3] - chrome
	if available < 0 {
		return 0
	}
	return available
}

func (inst *Instance) buildHeaderLine(widths []int, view tableView) string {
	cells := make([]string, 0, len(inst.columns)+1)
	if inst.selectionMode != SelectionNone {
		cells = append(cells, formatCell(inst.selectionHeaderMarker(view), selectionColumnWidth, rtui.AlignStart))
	}
	for columnIndex := range inst.columns {
		cells = append(cells, formatCell(inst.headerLabel(columnIndex), widths[columnIndex], inst.columns[columnIndex].Align))
	}
	return strings.Join(cells, " │ ")
}

func (inst *Instance) buildDataLine(row rowView, widths []int) string {
	formatted := make([]string, 0, len(widths)+1)
	if inst.selectionMode != SelectionNone {
		formatted = append(formatted, formatCell(inst.selectionMarker(row.sourceIndex), selectionColumnWidth, rtui.AlignStart))
	}
	for columnIndex := range widths {
		cell := ""
		if columnIndex < len(row.cells) {
			cell = row.cells[columnIndex]
		}
		formatted = append(formatted, formatCell(cell, widths[columnIndex], inst.columns[columnIndex].Align))
	}
	return strings.Join(formatted, " │ ")
}

func (inst *Instance) headerLabel(columnIndex int) string {
	label := inst.columns[columnIndex].Title
	sortColumn, sortDescending := inst.effectiveSortState()
	if sortColumn == columnIndex {
		if sortDescending {
			return label + " ↓"
		}
		return label + " ↑"
	}
	return label
}

func (inst *Instance) rowStyleFor(absoluteIndex int) style.Style {
	if absoluteIndex == inst.selectedIndex {
		return inst.selectedStyle
	}
	return inst.tableStyle
}

func (inst *Instance) shouldShowFooter(view tableView) bool {
	if !inst.showFooter {
		return false
	}
	sortColumn, _ := inst.effectiveSortState()
	return inst.pageSize > 0 ||
		inst.searchQuery != "" ||
		len(inst.filters) > 0 ||
		inst.selectionMode != SelectionNone ||
		(sortColumn >= 0 && sortColumn < len(inst.columns)) ||
		view.filteredCount != view.totalCount
}

func (inst *Instance) statusText(view tableView) string {
	parts := []string{fmt.Sprintf("Rows %d/%d", view.filteredCount, view.totalCount)}
	if view.pageCount > 1 {
		parts = append(parts, fmt.Sprintf("Page %d/%d", view.currentPage+1, view.pageCount))
	}
	sortColumn, sortDescending := inst.effectiveSortState()
	if sortColumn >= 0 && sortColumn < len(inst.columns) {
		direction := "↑"
		if sortDescending {
			direction = "↓"
		}
		parts = append(parts, fmt.Sprintf("Sort %s %s", inst.columns[sortColumn].Title, direction))
	}
	if query := strings.TrimSpace(inst.searchQuery); query != "" {
		parts = append(parts, fmt.Sprintf("Search %q", query))
	}
	if len(inst.filters) > 0 {
		parts = append(parts, fmt.Sprintf("Filters %d", len(inst.filters)))
	}
	if inst.selectionMode != SelectionNone {
		parts = append(parts, fmt.Sprintf("Checked %d", len(inst.checkedIndices)))
	}
	return strings.Join(parts, " · ")
}

func (inst *Instance) handleClick(act *action.Action) bool {
	mouseMsg, ok := act.Payload.(*runtimemsg.MouseMsg)
	if !ok || mouseMsg == nil {
		return false
	}

	view := inst.processedView()
	widths, _ := inst.calculateColumnWidths(view.rows, inst.maxInnerPaintWidth())
	if mouseMsg.LocalY == inst.headerLocalY() {
		if inst.selectionHeaderHit(mouseMsg.LocalX) {
			return inst.selectAllFilteredRows()
		}
		columnIndex, hit := inst.columnAtLocalX(mouseMsg.LocalX, widths)
		if !hit {
			return false
		}
		if inst.toggleSort(columnIndex) {
			return true
		}
		return inst.columns[columnIndex].Sortable
	}

	rowIndex, hit := inst.rowIndexAtLocalY(mouseMsg.LocalY, view)
	if !hit {
		return false
	}
	inst.selectIndex(rowIndex)
	if inst.selectionMode != SelectionNone && rowIndex >= 0 && rowIndex < len(view.rows) {
		inst.applySelectionAtSourceIndex(view.rows[rowIndex].sourceIndex)
	}
	return true
}

func (inst *Instance) headerLocalY() int {
	if inst.showBorder {
		return 1
	}
	return 0
}

func (inst *Instance) dataStartLocalY() int {
	start := 2 + inst.gap
	if inst.hasActiveColumnFilters() {
		start++
	}
	if inst.showBorder {
		start++
	}
	return start
}

func (inst *Instance) rowIndexAtLocalY(localY int, view tableView) (int, bool) {
	relative := localY - inst.dataStartLocalY()
	if relative < 0 || relative >= len(view.pageRows) {
		return -1, false
	}
	return view.start + relative, true
}

func (inst *Instance) selectionHeaderHit(localX int) bool {
	if inst.selectionMode != SelectionMultiple {
		return false
	}
	contentX, ok := inst.contentLocalX(localX)
	if !ok {
		return false
	}
	return contentX >= 0 && contentX < selectionColumnWidth
}

func (inst *Instance) effectiveCurrentPage() int {
	if inst.currentPageControlled && inst.hasPendingCurrentPage {
		return inst.pendingCurrentPage
	}
	return inst.currentPage
}

func (inst *Instance) effectiveSortState() (int, bool) {
	if inst.sortControlled && inst.hasPendingSort {
		return inst.pendingSortColumn, inst.pendingSortDescending
	}
	return inst.sortColumn, inst.sortDescending
}

func (inst *Instance) reconcilePendingState(oldCurrentPage, oldSortColumn int, oldSortDescending bool) {
	if inst.currentPageControlled && inst.hasPendingCurrentPage {
		if inst.currentPage == inst.pendingCurrentPage || inst.currentPage != oldCurrentPage {
			inst.hasPendingCurrentPage = false
		}
	}
	if inst.sortControlled && inst.hasPendingSort {
		if (inst.sortColumn == inst.pendingSortColumn && inst.sortDescending == inst.pendingSortDescending) ||
			inst.sortColumn != oldSortColumn || inst.sortDescending != oldSortDescending {
			inst.hasPendingSort = false
		}
	}
}

func (inst *Instance) columnAtLocalX(localX int, widths []int) (int, bool) {
	contentX, ok := inst.contentLocalX(localX)
	if !ok {
		return -1, false
	}
	localX = contentX

	if inst.showBorder {
	}

	if inst.selectionMode != SelectionNone {
		if localX < selectionColumnWidth {
			return -1, false
		}
		localX -= selectionColumnWidth
		if localX < 3 {
			return -1, false
		}
		localX -= 3
	}

	position := 0
	for columnIndex, width := range widths {
		if localX >= position && localX < position+width {
			return columnIndex, true
		}
		position += width
		if columnIndex < len(widths)-1 {
			if localX < position+3 {
				return -1, false
			}
			position += 3
		}
	}
	return -1, false
}

func (inst *Instance) contentLocalX(localX int) (int, bool) {
	if inst.showBorder {
		if localX < 2 {
			return -1, false
		}
		localX -= 2
	}
	return localX, true
}

func (inst *Instance) nextSortState(columnIndex int) (int, bool) {
	currentSortColumn, currentSortDescending := inst.effectiveSortState()
	if currentSortColumn != columnIndex {
		return columnIndex, false
	}
	if !currentSortDescending {
		return columnIndex, true
	}
	return -1, false
}

func (inst *Instance) toggleSort(columnIndex int) bool {
	if columnIndex < 0 || columnIndex >= len(inst.columns) {
		return false
	}
	if !inst.columns[columnIndex].Sortable {
		return false
	}

	nextSortColumn, nextSortDescending := inst.nextSortState(columnIndex)
	inst.sortColumn = nextSortColumn
	inst.sortDescending = nextSortDescending
	inst.currentPage = 0
	inst.selectedIndex = -1
	if inst.sortControlled {
		inst.pendingSortColumn = nextSortColumn
		inst.pendingSortDescending = nextSortDescending
		inst.hasPendingSort = true
	}
	if inst.currentPageControlled {
		inst.pendingCurrentPage = 0
		inst.hasPendingCurrentPage = true
	}
	inst.normalizeViewState(true)
	inst.dirty = true
	if inst.sortControlled || inst.currentPageControlled || inst.selectedIndexControlled {
		inst.emitStateSnapshot(inst.selectedIndex, inst.currentPage, inst.sortColumn, inst.sortDescending)
		return true
	}
	inst.emitStateChanged()
	return true
}

func (inst *Instance) selectIndex(index int) bool {
	rows := inst.filteredSortedRows()
	if len(rows) == 0 {
		return false
	}

	clamped := clampInt(index, 0, len(rows)-1)
	targetPage := inst.currentPage
	if inst.pageSize > 0 {
		targetPage = clamped / maxInt(1, inst.pageSize)
	}
	if inst.selectedIndex == clamped && (inst.pageSize <= 0 || inst.currentPage == targetPage) {
		return false
	}

	inst.selectedIndex = clamped
	if inst.pageSize > 0 {
		inst.currentPage = targetPage
	}
	if inst.currentPageControlled {
		inst.pendingCurrentPage = targetPage
		inst.hasPendingCurrentPage = true
	}
	if inst.selectedIndexControlled || inst.currentPageControlled {
		inst.dirty = true
		inst.emitStateSnapshot(clamped, targetPage, inst.sortColumn, inst.sortDescending)
		return true
	}
	inst.dirty = true
	inst.emitStateChanged()
	return true
}

func (inst *Instance) navigateUp() bool {
	rows := inst.filteredSortedRows()
	if len(rows) == 0 {
		return false
	}
	if inst.selectedIndex < 0 {
		return inst.selectIndex(0)
	}
	return inst.selectIndex(inst.selectedIndex - 1)
}

func (inst *Instance) navigateDown() bool {
	rows := inst.filteredSortedRows()
	if len(rows) == 0 {
		return false
	}
	if inst.selectedIndex < 0 {
		return inst.selectIndex(0)
	}
	return inst.selectIndex(inst.selectedIndex + 1)
}

func (inst *Instance) navigateHome() bool {
	return inst.selectIndex(0)
}

func (inst *Instance) navigateEnd() bool {
	rows := inst.filteredSortedRows()
	if len(rows) == 0 {
		return false
	}
	return inst.selectIndex(len(rows) - 1)
}

func (inst *Instance) movePage(delta int) bool {
	if inst.pageSize <= 0 {
		return false
	}

	view := inst.processedView()
	if view.pageCount <= 1 {
		return false
	}

	nextPage := clampInt(view.currentPage+delta, 0, view.pageCount-1)
	if nextPage == view.currentPage {
		return false
	}

	requestedSelectedIndex := inst.selectedIndex
	start := nextPage * inst.pageSize
	rowOffset := 0
	if inst.selectedIndex >= 0 {
		rowOffset = inst.selectedIndex % inst.pageSize
	}
	target := minInt(start+rowOffset, len(view.rows)-1)
	if target < start {
		target = start
	}
	requestedSelectedIndex = target
	inst.selectedIndex = target
	inst.currentPage = nextPage
	if inst.currentPageControlled {
		inst.pendingCurrentPage = nextPage
		inst.hasPendingCurrentPage = true
	}
	if inst.currentPageControlled {
		inst.dirty = true
		inst.emitStateSnapshot(requestedSelectedIndex, nextPage, inst.sortColumn, inst.sortDescending)
		return true
	}
	inst.dirty = true
	inst.emitStateChanged()
	return true
}

func (inst *Instance) emitStateChanged() {
	inst.emitStateSnapshot(inst.selectedIndex, inst.currentPage, inst.sortColumn, inst.sortDescending)
}

func (inst *Instance) emitStateSnapshot(selectedIndex, currentPage, sortColumn int, sortDescending bool) {
	if inst.intentEmitter == nil {
		return
	}

	rows := inst.filteredSortedRows()
	pageCount := pageCountFor(len(rows), inst.pageSize)
	if pageCount < 1 {
		pageCount = 1
	}
	clampedPage := clampInt(currentPage, 0, pageCount-1)
	visibleRows := visibleRowsForPage(len(rows), inst.pageSize, clampedPage)
	selectedSourceIndex := inst.selectedSourceIndexFor(rows, selectedIndex)
	if inst.componentID != "" {
		inst.intentEmitter(StateChange(
			inst.componentID,
			selectedIndex,
			selectedSourceIndex,
			clampedPage,
			pageCount,
			inst.pageSize,
			sortColumn,
			sortDescending,
			inst.searchQuery,
			inst.filters,
			visibleRows,
			len(rows),
			len(inst.rows),
		))
	}

	if inst.changeIntentField != nil {
		inst.intentEmitter(intent.FieldChangeIntent{
			Field: inst.changeIntentField.GetField(),
			Value: strconv.Itoa(selectedSourceIndex),
		})
	} else if inst.changeIntent != nil {
		inst.intentEmitter(inst.changeIntent)
	}
	if inst.pageIntentField != nil {
		inst.intentEmitter(intent.FieldChangeIntent{
			Field: inst.pageIntentField.GetField(),
			Value: strconv.Itoa(clampedPage),
		})
	}
}

func (inst *Instance) handleActivateSelection() bool {
	view := inst.processedView()
	if inst.selectedIndex < 0 || inst.selectedIndex >= len(view.rows) {
		return false
	}
	return inst.applySelectionAtSourceIndex(view.rows[inst.selectedIndex].sourceIndex)
}

func (inst *Instance) selectAllFilteredRows() bool {
	if inst.selectionMode != SelectionMultiple {
		return false
	}
	view := inst.processedView()
	if len(view.rows) == 0 {
		return false
	}

	allChecked := true
	for _, row := range view.rows {
		if !inst.isChecked(row.sourceIndex) {
			allChecked = false
			break
		}
	}

	if allChecked {
		filteredSet := make(map[int]struct{}, len(view.rows))
		for _, row := range view.rows {
			filteredSet[row.sourceIndex] = struct{}{}
		}
		next := inst.checkedIndices[:0]
		for _, checkedIndex := range inst.checkedIndices {
			if _, exists := filteredSet[checkedIndex]; !exists {
				next = append(next, checkedIndex)
			}
		}
		inst.checkedIndices = next
	} else {
		seen := make(map[int]struct{}, len(inst.checkedIndices)+len(view.rows))
		next := make([]int, 0, len(inst.checkedIndices)+len(view.rows))
		for _, checkedIndex := range inst.checkedIndices {
			if _, exists := seen[checkedIndex]; exists {
				continue
			}
			seen[checkedIndex] = struct{}{}
			next = append(next, checkedIndex)
		}
		for _, row := range view.rows {
			if _, exists := seen[row.sourceIndex]; exists {
				continue
			}
			seen[row.sourceIndex] = struct{}{}
			next = append(next, row.sourceIndex)
		}
		inst.checkedIndices = next
	}

	inst.normalizeCheckedIndices()
	inst.dirty = true
	inst.emitCheckedSelectionChanged()
	return true
}

func (inst *Instance) clearCheckedSelection() bool {
	if len(inst.checkedIndices) == 0 {
		return false
	}
	inst.checkedIndices = nil
	inst.dirty = true
	inst.emitCheckedSelectionChanged()
	return true
}

func (inst *Instance) applySelectionAtSourceIndex(sourceIndex int) bool {
	if inst.selectionMode == SelectionNone || sourceIndex < 0 || sourceIndex >= len(inst.rows) {
		return false
	}

	changed := false
	switch inst.selectionMode {
	case SelectionSingle:
		changed = !equalInts(inst.checkedIndices, []int{sourceIndex})
		inst.checkedIndices = []int{sourceIndex}
	case SelectionMultiple:
		if inst.isChecked(sourceIndex) {
			inst.checkedIndices = removeInt(inst.checkedIndices, sourceIndex)
		} else {
			inst.checkedIndices = append(inst.checkedIndices, sourceIndex)
		}
		sort.Ints(inst.checkedIndices)
		changed = true
	}
	inst.normalizeCheckedIndices()
	if changed {
		inst.dirty = true
		inst.emitCheckedSelectionChanged()
	}
	return true
}

func (inst *Instance) emitCheckedSelectionChanged() {
	if inst.intentEmitter == nil {
		return
	}
	value := inst.checkedSelectionValue()
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

func (inst *Instance) selectionMarker(sourceIndex int) string {
	if inst.isChecked(sourceIndex) {
		return "[x]"
	}
	return "[ ]"
}

func (inst *Instance) isChecked(sourceIndex int) bool {
	for _, checkedIndex := range inst.checkedIndices {
		if checkedIndex == sourceIndex {
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

func (inst *Instance) selectedSourceIndex(rows []rowView) int {
	return inst.selectedSourceIndexFor(rows, inst.selectedIndex)
}

func (inst *Instance) selectedSourceIndexFor(rows []rowView, selectedIndex int) int {
	if selectedIndex < 0 || selectedIndex >= len(rows) {
		return -1
	}
	return rows[selectedIndex].sourceIndex
}

func (inst *Instance) GetSelectedIndex() int {
	return inst.selectedIndex
}

func (inst *Instance) GetCheckedIndices() []int {
	return append([]int(nil), inst.checkedIndices...)
}

func (inst *Instance) GetSelectedSourceIndex() int {
	return inst.selectedSourceIndex(inst.filteredSortedRows())
}

func (inst *Instance) GetSelectedRow() ([]string, bool) {
	rows := inst.filteredSortedRows()
	if inst.selectedIndex < 0 || inst.selectedIndex >= len(rows) {
		return nil, false
	}
	row := append([]string(nil), rows[inst.selectedIndex].cells...)
	return row, true
}

func (inst *Instance) SelectIndexForAI(index int) bool {
	return inst.selectIndex(index)
}

func (inst *Instance) ToggleSelectionAtSourceIndex(index int) bool {
	return inst.applySelectionAtSourceIndex(index)
}

func getColumnsProp(props rtui.Props, def []TableColumn) []TableColumn {
	value, ok := props[propColumns]
	if !ok {
		return def
	}
	if columns, ok := value.([]TableColumn); ok {
		return columns
	}
	return def
}

func getRowsProp(props rtui.Props, def [][]string) [][]string {
	value, ok := props[propRows]
	if !ok {
		return def
	}
	if rows, ok := value.([][]string); ok {
		return rows
	}
	return def
}

func getFiltersProp(props rtui.Props, def map[int]string) map[int]string {
	value, ok := props[propFilters]
	if !ok {
		return cloneFilters(def)
	}
	if filters, ok := value.(map[int]string); ok {
		return cloneFilters(filters)
	}
	return cloneFilters(def)
}

func getIntsProp(props rtui.Props, key string, def []int) []int {
	if value, ok := props[key]; ok {
		if ints, ok := value.([]int); ok {
			return append([]int(nil), ints...)
		}
	}
	return append([]int(nil), def...)
}

func getSelectionModeProp(props rtui.Props, key string, def SelectionMode) SelectionMode {
	if value, ok := props[key]; ok {
		if mode, ok := value.(SelectionMode); ok {
			return mode
		}
	}
	return def
}

func getFieldIntentProp(props rtui.Props, key string) intent.FieldIntent {
	if value, ok := props[key]; ok {
		if result, ok := value.(intent.FieldIntent); ok {
			return result
		}
	}
	if value, ok := props[key+"Field"]; ok {
		if result, ok := value.(intent.FieldIntent); ok {
			return result
		}
	}
	return nil
}

func columnsEqual(left, right []TableColumn) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].Title != right[index].Title ||
			left[index].Width != right[index].Width ||
			left[index].Sortable != right[index].Sortable ||
			left[index].Align != right[index].Align {
			return false
		}
	}
	return true
}

func rowsEqual(left, right [][]string) bool {
	if len(left) != len(right) {
		return false
	}
	for rowIndex := range left {
		if len(left[rowIndex]) != len(right[rowIndex]) {
			return false
		}
		for columnIndex := range left[rowIndex] {
			if left[rowIndex][columnIndex] != right[rowIndex][columnIndex] {
				return false
			}
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

func equalFilters(left, right map[int]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, leftValue := range left {
		if rightValue, ok := right[key]; !ok || rightValue != leftValue {
			return false
		}
	}
	return true
}

func cloneRows(rows [][]string) [][]string {
	cloned := make([][]string, len(rows))
	for rowIndex, row := range rows {
		cloned[rowIndex] = append([]string(nil), row...)
	}
	return cloned
}

func removeInt(values []int, target int) []int {
	filtered := values[:0]
	for _, value := range values {
		if value != target {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func sameIntent(left, right intent.Intent) bool {
	return reflect.DeepEqual(left, right)
}

func sameFieldIntent(left, right intent.FieldIntent) bool {
	return reflect.DeepEqual(left, right)
}

func formatCell(text string, width int, align rtui.Align) string {
	trimmed := truncateText(text, width)
	switch align {
	case rtui.AlignCenter:
		return padCenterToWidth(trimmed, width)
	case rtui.AlignEnd:
		return padLeftToWidth(trimmed, width)
	default:
		return padRightToWidth(trimmed, width)
	}
}

func truncateText(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if paint.StringWidth(text) <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return trimToWidth(text, maxWidth)
	}

	runes := []rune(text)
	for end := len(runes); end > 0; end-- {
		candidate := string(runes[:end])
		if paint.StringWidth(candidate)+3 <= maxWidth {
			return candidate + "..."
		}
	}
	return trimToWidth(text, maxWidth)
}

func trimToWidth(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	var builder strings.Builder
	currentWidth := 0
	for _, character := range text {
		charWidth := paint.RuneWidth(character)
		if currentWidth+charWidth > maxWidth {
			break
		}
		builder.WriteRune(character)
		currentWidth += charWidth
	}
	return builder.String()
}

func padRightToWidth(text string, width int) string {
	textWidth := paint.StringWidth(text)
	if textWidth >= width {
		return trimToWidth(text, width)
	}
	return text + strings.Repeat(" ", width-textWidth)
}

func padLeftToWidth(text string, width int) string {
	textWidth := paint.StringWidth(text)
	if textWidth >= width {
		return trimToWidth(text, width)
	}
	return strings.Repeat(" ", width-textWidth) + text
}

func padCenterToWidth(text string, width int) string {
	textWidth := paint.StringWidth(text)
	if textWidth >= width {
		return trimToWidth(text, width)
	}
	left := (width - textWidth) / 2
	right := width - textWidth - left
	return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
}

func shrinkWidthsToFit(widths []int, target int) []int {
	if target <= 0 {
		shrunk := make([]int, len(widths))
		for index := range shrunk {
			shrunk[index] = 1
		}
		return shrunk
	}

	shrunk := append([]int(nil), widths...)
	current := 0
	for _, width := range shrunk {
		current += width
	}
	for current > target {
		largestIndex := -1
		for index, width := range shrunk {
			if width <= 1 {
				continue
			}
			if largestIndex == -1 || width > shrunk[largestIndex] {
				largestIndex = index
			}
		}
		if largestIndex == -1 {
			break
		}
		shrunk[largestIndex]--
		current--
	}
	return shrunk
}

func compareCellValues(left, right string) int {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)

	leftNumber, leftErr := strconv.ParseFloat(left, 64)
	rightNumber, rightErr := strconv.ParseFloat(right, 64)
	if leftErr == nil && rightErr == nil {
		switch {
		case leftNumber < rightNumber:
			return -1
		case leftNumber > rightNumber:
			return 1
		default:
			return 0
		}
	}

	leftLower := strings.ToLower(left)
	rightLower := strings.ToLower(right)
	switch {
	case leftLower < rightLower:
		return -1
	case leftLower > rightLower:
		return 1
	default:
		return 0
	}
}

func pageCountFor(totalRows, pageSize int) int {
	if pageSize <= 0 {
		return 1
	}
	if totalRows <= 0 {
		return 1
	}
	return (totalRows + pageSize - 1) / pageSize
}

func visibleRowsForPage(totalRows, pageSize, currentPage int) int {
	if totalRows <= 0 {
		return 0
	}
	if pageSize <= 0 {
		return totalRows
	}
	pageCount := pageCountFor(totalRows, pageSize)
	if pageCount < 1 {
		pageCount = 1
	}
	page := clampInt(currentPage, 0, pageCount-1)
	start := page * pageSize
	end := minInt(start+pageSize, totalRows)
	if end < start {
		return 0
	}
	return end - start
}

func clampInt(value, minimum, maximum int) int {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
