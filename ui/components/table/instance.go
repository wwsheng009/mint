package table

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/wwsheng009/mint/runtime/action"
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

// Instance is the runtime entity for table components.
type Instance struct {
	key string

	columns       []TableColumn
	rows          [][]string
	emptyText     string
	headerStyle   style.Style
	tableStyle    style.Style
	selectedStyle style.Style
	borderStyle   style.Style
	statusStyle   style.Style
	gap           int
	showBorder    bool
	showFooter    bool
	pageSize      int
	searchQuery   string
	filters       map[int]string

	currentPage             int
	currentPageControlled   bool
	sortColumn              int
	sortDescending          bool
	sortControlled          bool
	selectedIndex           int
	selectedIndexControlled bool

	focused bool
	bounds  [4]int
	dirty   bool
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
		key:                     getStringProp(props, "key", ""),
		columns:                 getColumnsProp(props, []TableColumn{}),
		rows:                    getRowsProp(props, [][]string{}),
		emptyText:               getStringProp(props, "emptyText", "(empty)"),
		headerStyle:             getStyleProp(props, "headerStyle"),
		tableStyle:              getStyleProp(props, "tableStyle"),
		selectedStyle:           getStylePropWithDefault(props, "selectedStyle", style.Style{}.Reverse(true)),
		borderStyle:             getStylePropWithDefault(props, "borderStyle", style.Style{}.Foreground(style.BrightBlack)),
		statusStyle:             getStylePropWithDefault(props, "statusStyle", style.Style{}.Foreground(style.BrightBlack)),
		gap:                     maxInt(0, getIntProp(props, "gap", 0)),
		showBorder:              getBoolProp(props, "showBorder", false),
		showFooter:              getBoolProp(props, "showFooter", true),
		pageSize:                maxInt(0, getIntProp(props, "pageSize", 0)),
		searchQuery:             getStringProp(props, "searchQuery", ""),
		filters:                 getFiltersProp(props, map[int]string{}),
		currentPage:             maxInt(0, getIntProp(props, "currentPage", 0)),
		currentPageControlled:   getBoolProp(props, "currentPageControlled", false),
		sortColumn:              getIntProp(props, "sortColumn", -1),
		sortDescending:          getBoolProp(props, "sortDescending", false),
		sortControlled:          getBoolProp(props, "sortControlled", false),
		selectedIndex:           getIntProp(props, "selectedIndex", -1),
		selectedIndexControlled: getBoolProp(props, "selectedIndexControlled", false),
		dirty:                   true,
	}
	inst.normalizeViewState(false)
	return inst
}

func (inst *Instance) Key() string           { return inst.key }
func (inst *Instance) SetKey(key string)     { inst.key = key }
func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *Instance) Destroy()              { inst.columns = nil; inst.rows = nil; inst.filters = nil }
func (inst *Instance) OnMount()              { inst.dirty = true }
func (inst *Instance) OnUnmount()            {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldColumns := append([]TableColumn(nil), inst.columns...)
	oldRows := cloneRows(inst.rows)
	oldEmptyText := inst.emptyText
	oldHeaderStyle := inst.headerStyle
	oldTableStyle := inst.tableStyle
	oldSelectedStyle := inst.selectedStyle
	oldBorderStyle := inst.borderStyle
	oldStatusStyle := inst.statusStyle
	oldGap := inst.gap
	oldShowBorder := inst.showBorder
	oldShowFooter := inst.showFooter
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

	inst.columns = getColumnsProp(props, inst.columns)
	inst.rows = getRowsProp(props, inst.rows)
	inst.emptyText = getStringProp(props, "emptyText", inst.emptyText)
	inst.headerStyle = getStylePropWithDefault(props, "headerStyle", inst.headerStyle)
	inst.tableStyle = getStylePropWithDefault(props, "tableStyle", inst.tableStyle)
	inst.selectedStyle = getStylePropWithDefault(props, "selectedStyle", inst.selectedStyle)
	inst.borderStyle = getStylePropWithDefault(props, "borderStyle", inst.borderStyle)
	inst.statusStyle = getStylePropWithDefault(props, "statusStyle", inst.statusStyle)
	inst.gap = maxInt(0, getIntProp(props, "gap", inst.gap))
	inst.showBorder = getBoolProp(props, "showBorder", inst.showBorder)
	inst.showFooter = getBoolProp(props, "showFooter", inst.showFooter)
	inst.pageSize = maxInt(0, getIntProp(props, "pageSize", inst.pageSize))
	inst.searchQuery = getStringProp(props, "searchQuery", inst.searchQuery)
	inst.filters = getFiltersProp(props, inst.filters)

	if controlled, ok := props["currentPageControlled"].(bool); ok {
		inst.currentPageControlled = controlled
	}
	if inst.currentPageControlled {
		inst.currentPage = maxInt(0, getIntProp(props, "currentPage", inst.currentPage))
	}

	if controlled, ok := props["sortControlled"].(bool); ok {
		inst.sortControlled = controlled
	}
	if inst.sortControlled {
		inst.sortColumn = getIntProp(props, "sortColumn", inst.sortColumn)
		inst.sortDescending = getBoolProp(props, "sortDescending", inst.sortDescending)
	}

	if controlled, ok := props["selectedIndexControlled"].(bool); ok {
		inst.selectedIndexControlled = controlled
	}
	if inst.selectedIndexControlled {
		inst.selectedIndex = getIntProp(props, "selectedIndex", inst.selectedIndex)
	}

	resetPage := oldSearchQuery != inst.searchQuery || !equalFilters(oldFilters, inst.filters)
	inst.normalizeViewState(resetPage)

	changed := !columnsEqual(oldColumns, inst.columns) ||
		!rowsEqual(oldRows, inst.rows) ||
		oldEmptyText != inst.emptyText ||
		oldHeaderStyle != inst.headerStyle ||
		oldTableStyle != inst.tableStyle ||
		oldSelectedStyle != inst.selectedStyle ||
		oldBorderStyle != inst.borderStyle ||
		oldStatusStyle != inst.statusStyle ||
		oldGap != inst.gap ||
		oldShowBorder != inst.showBorder ||
		oldShowFooter != inst.showFooter ||
		oldPageSize != inst.pageSize ||
		oldSearchQuery != inst.searchQuery ||
		!equalFilters(oldFilters, inst.filters) ||
		oldCurrentPage != inst.currentPage ||
		oldCurrentPageControlled != inst.currentPageControlled ||
		oldSortColumn != inst.sortColumn ||
		oldSortDescending != inst.sortDescending ||
		oldSortControlled != inst.sortControlled ||
		oldSelectedIndex != inst.selectedIndex ||
		oldSelectedIndexControlled != inst.selectedIndexControlled
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":                     inst.key,
		"columns":                 inst.columns,
		"rows":                    inst.rows,
		"emptyText":               inst.emptyText,
		"headerStyle":             inst.headerStyle,
		"tableStyle":              inst.tableStyle,
		"selectedStyle":           inst.selectedStyle,
		"borderStyle":             inst.borderStyle,
		"statusStyle":             inst.statusStyle,
		"gap":                     inst.gap,
		"showBorder":              inst.showBorder,
		"showFooter":              inst.showFooter,
		"pageSize":                inst.pageSize,
		"searchQuery":             inst.searchQuery,
		"filters":                 cloneFilters(inst.filters),
		"currentPage":             inst.currentPage,
		"currentPageControlled":   inst.currentPageControlled,
		"sortColumn":              inst.sortColumn,
		"sortDescending":          inst.sortDescending,
		"sortControlled":          inst.sortControlled,
		"selectedIndex":           inst.selectedIndex,
		"selectedIndexControlled": inst.selectedIndexControlled,
	}
}

func (inst *Instance) MarkDirty()                         { inst.dirty = true }
func (inst *Instance) IsDirty() bool                      { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }
func (inst *Instance) ClearDirty()                        { inst.dirty = false }

func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if len(inst.columns) == 0 {
		return layout.Size{}
	}

	view := inst.processedView()
	_, innerWidth := inst.calculateColumnWidths(view.rows, 0)
	if inst.shouldShowFooter(view) {
		innerWidth = maxInt(innerWidth, paint.StringWidth(inst.statusText(view)))
	}
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
	widths, innerWidth := inst.calculateColumnWidths(view.rows, maxInnerWidth)
	if innerWidth < 1 {
		innerWidth = 1
	}

	showFooter := inst.shouldShowFooter(view)
	statusText := inst.statusText(view)
	if showFooter {
		innerWidth = maxInt(innerWidth, paint.StringWidth(statusText))
		if maxInnerWidth > 0 {
			innerWidth = minInt(innerWidth, maxInnerWidth)
		}
	}
	renderedRows := view.pageRows
	if limit := inst.maxRenderableRows(showFooter); limit >= 0 && len(renderedRows) > limit {
		renderedRows = renderedRows[:limit]
	}

	type lineSpec struct {
		text  string
		style style.Style
	}

	innerLines := make([]lineSpec, 0, inst.lineCountForView(view))
	innerLines = append(innerLines, lineSpec{
		text:  padRightToWidth(truncateText(inst.buildHeaderLine(widths), innerWidth), innerWidth),
		style: inst.headerStyle,
	})
	innerLines = append(innerLines, lineSpec{text: strings.Repeat("─", maxInt(1, innerWidth)), style: inst.borderStyle})
	for i := 0; i < inst.gap; i++ {
		innerLines = append(innerLines, lineSpec{text: strings.Repeat(" ", maxInt(1, innerWidth)), style: inst.tableStyle})
	}

	if len(renderedRows) == 0 {
		innerLines = append(innerLines, lineSpec{
			text:  padRightToWidth(truncateText(inst.emptyText, innerWidth), innerWidth),
			style: inst.tableStyle,
		})
	} else {
		for rowOffset, row := range renderedRows {
			absoluteIndex := view.start + rowOffset
			innerLines = append(innerLines, lineSpec{
				text:  padRightToWidth(truncateText(inst.buildDataLine(row.cells, widths), innerWidth), innerWidth),
				style: inst.rowStyleFor(absoluteIndex),
			})
		}
	}

	if showFooter {
		innerLines = append(innerLines, lineSpec{
			text:  padRightToWidth(truncateText(statusText, innerWidth), innerWidth),
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
			return inst.selectIndex(0)
		}
		return len(inst.filteredSortedRows()) > 0
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
	page := clampInt(inst.currentPage, 0, maxInt(0, pageCount-1))
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

	if inst.sortColumn >= 0 && inst.sortColumn < len(inst.columns) {
		sortColumn := inst.sortColumn
		descending := inst.sortDescending
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

	separatorWidth := maxInt(0, len(inst.columns)-1) * 3
	if maxInnerWidth > 0 && totalContentWidth+separatorWidth > maxInnerWidth {
		widths = shrinkWidthsToFit(widths, maxInnerWidth-separatorWidth)
		totalContentWidth = 0
		for _, width := range widths {
			totalContentWidth += width
		}
	}

	return widths, totalContentWidth + separatorWidth
}

func (inst *Instance) lineCountForView(view tableView) int {
	if len(inst.columns) == 0 {
		return 0
	}
	lines := 2 + inst.gap
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

func (inst *Instance) buildHeaderLine(widths []int) string {
	cells := make([]string, len(inst.columns))
	for columnIndex := range inst.columns {
		cells[columnIndex] = formatCell(inst.headerLabel(columnIndex), widths[columnIndex])
	}
	return strings.Join(cells, " │ ")
}

func (inst *Instance) buildDataLine(cells []string, widths []int) string {
	formatted := make([]string, len(widths))
	for columnIndex := range widths {
		cell := ""
		if columnIndex < len(cells) {
			cell = cells[columnIndex]
		}
		formatted[columnIndex] = formatCell(cell, widths[columnIndex])
	}
	return strings.Join(formatted, " │ ")
}

func (inst *Instance) headerLabel(columnIndex int) string {
	label := inst.columns[columnIndex].Title
	if inst.sortColumn == columnIndex {
		if inst.sortDescending {
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
	return inst.pageSize > 0 ||
		inst.searchQuery != "" ||
		len(inst.filters) > 0 ||
		(inst.sortColumn >= 0 && inst.sortColumn < len(inst.columns)) ||
		view.filteredCount != view.totalCount
}

func (inst *Instance) statusText(view tableView) string {
	parts := []string{fmt.Sprintf("Rows %d/%d", view.filteredCount, view.totalCount)}
	if view.pageCount > 1 {
		parts = append(parts, fmt.Sprintf("Page %d/%d", view.currentPage+1, view.pageCount))
	}
	if inst.sortColumn >= 0 && inst.sortColumn < len(inst.columns) {
		direction := "↑"
		if inst.sortDescending {
			direction = "↓"
		}
		parts = append(parts, fmt.Sprintf("Sort %s %s", inst.columns[inst.sortColumn].Title, direction))
	}
	if query := strings.TrimSpace(inst.searchQuery); query != "" {
		parts = append(parts, fmt.Sprintf("Search %q", query))
	}
	if len(inst.filters) > 0 {
		parts = append(parts, fmt.Sprintf("Filters %d", len(inst.filters)))
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
	if !inst.selectedIndexControlled {
		inst.selectIndex(rowIndex)
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

func (inst *Instance) columnAtLocalX(localX int, widths []int) (int, bool) {
	if inst.showBorder {
		if localX < 2 {
			return -1, false
		}
		localX -= 2
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

func (inst *Instance) toggleSort(columnIndex int) bool {
	if inst.sortControlled || columnIndex < 0 || columnIndex >= len(inst.columns) {
		return false
	}
	if !inst.columns[columnIndex].Sortable {
		return false
	}

	if inst.sortColumn == columnIndex {
		inst.sortDescending = !inst.sortDescending
	} else {
		inst.sortColumn = columnIndex
		inst.sortDescending = false
	}
	if !inst.currentPageControlled {
		inst.currentPage = 0
	}
	if !inst.selectedIndexControlled {
		inst.selectedIndex = -1
	}
	inst.normalizeViewState(true)
	inst.dirty = true
	return true
}

func (inst *Instance) selectIndex(index int) bool {
	rows := inst.filteredSortedRows()
	if len(rows) == 0 || inst.selectedIndexControlled {
		return false
	}

	clamped := clampInt(index, 0, len(rows)-1)
	if inst.selectedIndex == clamped && (inst.pageSize <= 0 || inst.currentPage == clamped/maxInt(1, inst.pageSize)) {
		return false
	}

	inst.selectedIndex = clamped
	if inst.pageSize > 0 && !inst.currentPageControlled {
		inst.currentPage = clamped / inst.pageSize
	}
	inst.dirty = true
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
	if inst.pageSize <= 0 || inst.currentPageControlled {
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

	inst.currentPage = nextPage
	if !inst.selectedIndexControlled {
		start := nextPage * inst.pageSize
		rowOffset := 0
		if inst.selectedIndex >= 0 {
			rowOffset = inst.selectedIndex % inst.pageSize
		}
		target := minInt(start+rowOffset, len(view.rows)-1)
		if target < start {
			target = start
		}
		inst.selectedIndex = target
	}
	inst.dirty = true
	return true
}

func getStringProp(props rtui.Props, key, def string) string {
	if value, ok := props[key]; ok {
		if text, ok := value.(string); ok {
			return text
		}
	}
	return def
}

func getIntProp(props rtui.Props, key string, def int) int {
	if value, ok := props[key]; ok {
		if number, ok := value.(int); ok {
			return number
		}
	}
	return def
}

func getBoolProp(props rtui.Props, key string, def bool) bool {
	if value, ok := props[key]; ok {
		if flag, ok := value.(bool); ok {
			return flag
		}
	}
	return def
}

func getStyleProp(props rtui.Props, key string) style.Style {
	value, ok := props[key]
	if !ok {
		return style.Style{}
	}
	if result, ok := value.(style.Style); ok {
		return result
	}
	return style.Style{}
}

func getStylePropWithDefault(props rtui.Props, key string, def style.Style) style.Style {
	value, ok := props[key]
	if !ok {
		return def
	}
	if result, ok := value.(style.Style); ok {
		return result
	}
	return def
}

func getColumnsProp(props rtui.Props, def []TableColumn) []TableColumn {
	value, ok := props["columns"]
	if !ok {
		return def
	}
	if columns, ok := value.([]TableColumn); ok {
		return columns
	}
	return def
}

func getRowsProp(props rtui.Props, def [][]string) [][]string {
	value, ok := props["rows"]
	if !ok {
		return def
	}
	if rows, ok := value.([][]string); ok {
		return rows
	}
	return def
}

func getFiltersProp(props rtui.Props, def map[int]string) map[int]string {
	value, ok := props["filters"]
	if !ok {
		return cloneFilters(def)
	}
	if filters, ok := value.(map[int]string); ok {
		return cloneFilters(filters)
	}
	return cloneFilters(def)
}

func columnsEqual(left, right []TableColumn) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].Title != right[index].Title ||
			left[index].Width != right[index].Width ||
			left[index].Sortable != right[index].Sortable {
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

func formatCell(text string, width int) string {
	return padRightToWidth(truncateText(text, width), width)
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
