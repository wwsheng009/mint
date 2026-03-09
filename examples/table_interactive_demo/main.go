package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	tablecomp "github.com/wwsheng009/mint/ui/components/table"
)

type IncidentRecord struct {
	ID      int
	Service string
	Region  string
	Owner   string
	Status  string
	P95MS   int
	ErrorPC float64
	Updated string
}

type AppState struct {
	SearchText    string
	StatusFilter  string
	RegionFilter  string
	SelectedRow   string
	CheckedRows   string
	SelectionMode string
	PageSize      int
	ShowBorder    bool
	ShowFooter    bool
	ShowScrollbar bool
	CurrentPage   int
	PageCount     int
	SortColumn    int
	SortDesc      bool
	FilteredRows  int
	LastAction    string
}

type ToggleBorderIntent struct{}
type ToggleFooterIntent struct{}
type ToggleScrollbarIntent struct{}
type ClearFiltersIntent struct{}
type ResetDemoIntent struct{}
type ClearCheckedIntent struct{}
type SetSelectionModeIntent struct {
	Mode string
}
type StepPageIntent struct {
	Delta int
}
type SetPageIntent struct {
	Page int
}

type AdjustPageSizeIntent struct {
	Delta int
}

func (ToggleBorderIntent) IntentType() string     { return "TableDemoToggleBorder" }
func (ToggleFooterIntent) IntentType() string     { return "TableDemoToggleFooter" }
func (ToggleScrollbarIntent) IntentType() string  { return "TableDemoToggleScrollbar" }
func (ClearFiltersIntent) IntentType() string     { return "TableDemoClearFilters" }
func (ResetDemoIntent) IntentType() string        { return "TableDemoReset" }
func (ClearCheckedIntent) IntentType() string     { return "TableDemoClearChecked" }
func (SetSelectionModeIntent) IntentType() string { return "TableDemoSetSelectionMode" }
func (StepPageIntent) IntentType() string         { return "TableDemoStepPage" }
func (SetPageIntent) IntentType() string          { return "TableDemoSetPage" }
func (AdjustPageSizeIntent) IntentType() string   { return "TableDemoAdjustPageSize" }
func (ToggleBorderIntent) StayPressed() bool      { return true }
func (ToggleFooterIntent) StayPressed() bool      { return true }
func (ToggleScrollbarIntent) StayPressed() bool   { return true }
func (ClearFiltersIntent) StayPressed() bool      { return true }
func (ResetDemoIntent) StayPressed() bool         { return true }
func (ClearCheckedIntent) StayPressed() bool      { return true }
func (SetSelectionModeIntent) StayPressed() bool  { return true }
func (StepPageIntent) StayPressed() bool          { return true }
func (SetPageIntent) StayPressed() bool           { return true }
func (AdjustPageSizeIntent) StayPressed() bool    { return true }

var demoRecords = generateRecords()
var demoRows = buildRows(demoRecords)
var demoColumns = []tablecomp.TableColumn{
	{Title: "ID", Width: 6, Sortable: true, Align: rtui.AlignEnd},
	{Title: "Service", Width: 16, Sortable: true},
	{Title: "Region", Width: 8, Sortable: true},
	{Title: "Owner", Width: 10, Sortable: true},
	{Title: "Status", Width: 10, Sortable: true},
	{Title: "P95(ms)", Width: 8, Sortable: true, Align: rtui.AlignEnd},
	{Title: "Err%", Width: 6, Sortable: true, Align: rtui.AlignEnd},
	{Title: "Updated", Width: 10, Sortable: true},
}

var demoStore = store.NewStore(newInitialState())

func init() {
	reducer.NewBuilder[AppState]().
		On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
			fieldChange, ok := i.(intent.FieldChangeIntent)
			if !ok {
				return s
			}
			switch fieldChange.Field {
			case "searchText":
				s.SearchText = fieldChange.Value
				s.SelectedRow = ""
				s.CurrentPage = 0
				s = recomputeDerivedState(s)
				s.LastAction = fmt.Sprintf("Search = %q", strings.TrimSpace(s.SearchText))
			case "statusFilter":
				s.StatusFilter = fieldChange.Value
				s.SelectedRow = ""
				s.CurrentPage = 0
				s = recomputeDerivedState(s)
				s.LastAction = fmt.Sprintf("Status filter = %q", strings.TrimSpace(s.StatusFilter))
			case "regionFilter":
				s.RegionFilter = fieldChange.Value
				s.SelectedRow = ""
				s.CurrentPage = 0
				s = recomputeDerivedState(s)
				s.LastAction = fmt.Sprintf("Region filter = %q", strings.TrimSpace(s.RegionFilter))
			case "checkedRows":
				s.CheckedRows = normalizeCheckedRows(fieldChange.Value, s.SelectionMode)
				s.LastAction = fmt.Sprintf("CheckedRows = %s", checkedSummary(checkedIndices(s.CheckedRows)))
			case "currentPage":
				if page, err := strconv.Atoi(strings.TrimSpace(fieldChange.Value)); err == nil {
					s.CurrentPage = clampInt(page, 0, maxInt(0, s.PageCount-1))
				}
			}
			return s
		}).
		On(tablecomp.StateChangeIntent{}, func(s AppState, i intent.Intent) AppState {
			change, ok := i.(tablecomp.StateChangeIntent)
			if !ok || change.ComponentID != "ops.table" {
				return s
			}
			s.CurrentPage = change.CurrentPage
			s.PageCount = maxInt(1, change.PageCount)
			s.SortColumn = change.SortColumn
			s.SortDesc = change.SortDescending
			s.FilteredRows = change.FilteredRows
			s.SelectedRow = normalizeSelectedRow(strconv.Itoa(change.SelectedSourceIndex))
			switch {
			case change.SelectedSourceIndex >= 0:
				record := demoRecords[change.SelectedSourceIndex]
				s.LastAction = fmt.Sprintf("Page %d/%d · %s · Selected #%d %s",
					change.CurrentPage+1,
					maxInt(1, change.PageCount),
					sortSummary(change.SortColumn, change.SortDescending),
					record.ID,
					record.Service,
				)
			default:
				s.LastAction = fmt.Sprintf("Page %d/%d · %s",
					change.CurrentPage+1,
					maxInt(1, change.PageCount),
					sortSummary(change.SortColumn, change.SortDescending),
				)
			}
			return s
		}).
		On(ToggleBorderIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ShowBorder = !s.ShowBorder
			s.LastAction = fmt.Sprintf("Border = %t", s.ShowBorder)
			return s
		}).
		On(ToggleFooterIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ShowFooter = !s.ShowFooter
			s.LastAction = fmt.Sprintf("Footer = %t", s.ShowFooter)
			return s
		}).
		On(ToggleScrollbarIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ShowScrollbar = !s.ShowScrollbar
			s.LastAction = fmt.Sprintf("Scrollbar = %t", s.ShowScrollbar)
			return s
		}).
		On(SetSelectionModeIntent{}, func(s AppState, i intent.Intent) AppState {
			modeIntent, ok := i.(SetSelectionModeIntent)
			if !ok {
				return s
			}
			s.SelectionMode = normalizeSelectionMode(modeIntent.Mode)
			s.CheckedRows = normalizeCheckedRows(s.CheckedRows, s.SelectionMode)
			s.LastAction = fmt.Sprintf("SelectionMode = %s", strings.ToUpper(s.SelectionMode))
			return s
		}).
		On(ClearCheckedIntent{}, func(s AppState, i intent.Intent) AppState {
			s.CheckedRows = ""
			s.LastAction = "Cleared checked rows"
			return s
		}).
		On(StepPageIntent{}, func(s AppState, i intent.Intent) AppState {
			step, ok := i.(StepPageIntent)
			if !ok {
				return s
			}
			s.CurrentPage = clampInt(s.CurrentPage+step.Delta, 0, maxInt(0, s.PageCount-1))
			s.LastAction = fmt.Sprintf("Page %d/%d", s.CurrentPage+1, maxInt(1, s.PageCount))
			return s
		}).
		On(SetPageIntent{}, func(s AppState, i intent.Intent) AppState {
			setPage, ok := i.(SetPageIntent)
			if !ok {
				return s
			}
			s.CurrentPage = clampInt(setPage.Page, 0, maxInt(0, s.PageCount-1))
			s.LastAction = fmt.Sprintf("Page %d/%d", s.CurrentPage+1, maxInt(1, s.PageCount))
			return s
		}).
		On(ClearFiltersIntent{}, func(s AppState, i intent.Intent) AppState {
			s.SearchText = ""
			s.StatusFilter = ""
			s.RegionFilter = ""
			s.SelectedRow = ""
			s.CurrentPage = 0
			s = recomputeDerivedState(s)
			s.LastAction = "Cleared search and filters"
			return s
		}).
		On(AdjustPageSizeIntent{}, func(s AppState, i intent.Intent) AppState {
			adjust, ok := i.(AdjustPageSizeIntent)
			if !ok {
				return s
			}
			s.PageSize = clampInt(s.PageSize+adjust.Delta, 5, 18)
			s.SelectedRow = ""
			s.CurrentPage = 0
			s = recomputeDerivedState(s)
			s.LastAction = fmt.Sprintf("PageSize = %d", s.PageSize)
			return s
		}).
		On(ResetDemoIntent{}, func(s AppState, i intent.Intent) AppState {
			return newInitialState()
		}).
		BuildAndRegister(intent.DefaultRegistry(), demoStore)
}

func main() {
	err := ui.Run(App,
		ui.WithWidth(122),
		ui.WithHeight(38),
		ui.WithTitle("Interactive Table Demo"),
		ui.WithPluginSetup(func(app *framework.App) {
			app.OnKeyCombo("f2", func() { ui.EmitIntentGlobal(ToggleBorderIntent{}) })
			app.OnKeyCombo("f3", func() { ui.EmitIntentGlobal(ToggleFooterIntent{}) })
			app.OnKeyCombo("f4", func() { ui.EmitIntentGlobal(ToggleScrollbarIntent{}) })
			app.OnKeyCombo("f5", func() { ui.EmitIntentGlobal(ClearFiltersIntent{}) })
			app.OnKeyCombo("f6", func() { ui.EmitIntentGlobal(AdjustPageSizeIntent{Delta: -1}) })
			app.OnKeyCombo("f7", func() { ui.EmitIntentGlobal(AdjustPageSizeIntent{Delta: 1}) })
			app.OnKeyCombo("f8", func() { ui.EmitIntentGlobal(SetSelectionModeIntent{Mode: "none"}) })
			app.OnKeyCombo("f9", func() { ui.EmitIntentGlobal(SetSelectionModeIntent{Mode: "single"}) })
			app.OnKeyCombo("f10", func() { ui.EmitIntentGlobal(SetSelectionModeIntent{Mode: "multi"}) })
			app.OnKeyCombo("f11", func() { ui.EmitIntentGlobal(ClearCheckedIntent{}) })
			app.OnKeyCombo("f12", func() { ui.EmitIntentGlobal(ResetDemoIntent{}) })
		}),
	)
	if err != nil {
		panic(err)
	}
}

func App() ui.VNode {
	state := demoStore.Get()
	filteredCount := filteredRecordCount(state.SearchText, state.StatusFilter, state.RegionFilter)
	selectedRecord, hasSelected := selectedRecord(state.SelectedRow)

	return ui.NewVStack().
		SetGap(1).
		SetChildrenList([]ui.VNode{
			headerPanel(state, filteredCount),
			controlsPanel(state),
			ui.HStackBuilder(
				ui.Flex(tablePane(state), 3),
				ui.Flex(sidebar(state, filteredCount, selectedRecord, hasSelected), 2),
			).Gap(1).Stretch().Build(),
		})
}

func headerPanel(state AppState, filteredCount int) ui.VNode {
	return ui.NewVStack().
		SingleBorder("Interactive Table").
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder("搜索 / 列过滤 / 表头排序 / 键盘翻页 / 行选中 / checkbox 单选多选 / 状态同步").Bold(true).FgColor("bright-cyan").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Rows=%d  Filtered=%d  Page=%d/%d  PageSize=%d  Sort=%s",
				len(demoRecords),
				filteredCount,
				state.CurrentPage+1,
				maxInt(1, state.PageCount),
				state.PageSize,
				sortSummary(state.SortColumn, state.SortDesc))).
				FgColor("bright-white").
				Build(),
			ui.NewTextBuilder(fmt.Sprintf("Search=%q  Status=%q  Region=%q  Border=%t  Footer=%t",
				strings.TrimSpace(state.SearchText),
				strings.TrimSpace(state.StatusFilter),
				strings.TrimSpace(state.RegionFilter),
				state.ShowBorder,
				state.ShowFooter)).
				FgColor("yellow").
				Build(),
			ui.NewTextBuilder(fmt.Sprintf("SelectionMode=%s  Checked=%s  Scrollbar=%t  LastAction=%s",
				strings.ToUpper(state.SelectionMode),
				checkedSummary(checkedIndices(state.CheckedRows)),
				state.ShowScrollbar,
				state.LastAction)).
				FgColor("bright-black").
				Build(),
		})
}

func controlsPanel(state AppState) ui.VNode {
	return ui.NewVStack().
		SingleBorder("Controls").
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.HStackBuilder(
				ui.Flex(inputBlock("Search", state.SearchText, "searchText", "service / owner / status / region / id", 30), 2),
				ui.Flex(inputBlock("Status Filter", state.StatusFilter, "statusFilter", "open / ack / closed / triage", 24), 1),
				ui.Flex(inputBlock("Region Filter", state.RegionFilter, "regionFilter", "us-east / eu-west / ap-south", 24), 1),
			).Gap(1).Stretch().Build(),
			paginationControls(state),
			selectionControls(state),
			ui.NewTextBuilder("Tab 在三个 input 和 table 间切换；点击表头排序；↑↓ 选中，PageUp/PageDown 或 ←/→ 翻页").FgColor("bright-black").Build(),
		})
}

func paginationControls(state AppState) ui.VNode {
	return ui.HStackBuilder(
		ui.NewTextBuilder("Page").FgColor("cyan").Build(),
		ui.NewButtonBuilder("First").OnPress(SetPageIntent{Page: 0}).Build(),
		ui.NewButtonBuilder("Prev").OnPress(StepPageIntent{Delta: -1}).Build(),
		ui.NewTextBuilder(fmt.Sprintf("%d/%d", state.CurrentPage+1, maxInt(1, state.PageCount))).FgColor("bright-white").Build(),
		ui.NewButtonBuilder("Next").OnPress(StepPageIntent{Delta: 1}).Build(),
		ui.NewButtonBuilder("Last").OnPress(SetPageIntent{Page: maxInt(0, state.PageCount-1)}).Build(),
		ui.NewTextBuilder("受控分页：按钮与 table 内部翻页共享同一份状态").FgColor("bright-black").Build(),
	).Gap(1).Build()
}

func selectionControls(state AppState) ui.VNode {
	return ui.HStackBuilder(
		ui.NewTextBuilder("Selection").FgColor("cyan").Build(),
		ui.NewButtonBuilder("None").OnPress(SetSelectionModeIntent{Mode: "none"}).Build(),
		ui.NewButtonBuilder("Single").OnPress(SetSelectionModeIntent{Mode: "single"}).Build(),
		ui.NewButtonBuilder("Multi").OnPress(SetSelectionModeIntent{Mode: "multi"}).Build(),
		ui.NewButtonBuilder("Clear Checked").OnPress(ClearCheckedIntent{}).Build(),
		ui.NewTextBuilder(fmt.Sprintf("Mode=%s Checked=%s", strings.ToUpper(state.SelectionMode), checkedSummary(checkedIndices(state.CheckedRows)))).FgColor("bright-black").Build(),
	).Gap(1).Build()
}

func inputBlock(label, value, field, placeholder string, width int) ui.VNode {
	return ui.NewVStack().
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder(label).FgColor("cyan").Build(),
			ui.NewInputBuilder().
				Placeholder(placeholder).
				Value(value).
				Width(width).
				ForField(intent.BindField(field)).
				Build(),
		})
}

func tablePane(state AppState) ui.VNode {
	filters := map[int]string{
		2: state.RegionFilter,
		4: state.StatusFilter,
	}

	builder := tablecomp.NewBuilder().
		ComponentID("ops.table").
		Columns(demoColumns).
		Rows(demoRows).
		SearchQuery(state.SearchText).
		Filters(filters).
		PageSize(state.PageSize).
		CurrentPage(state.CurrentPage).
		SortBy(state.SortColumn, state.SortDesc).
		CheckedIndices(checkedIndices(state.CheckedRows)...).
		EmptyText("No incidents match the current search and filters").
		ShowBorder(state.ShowBorder).
		ShowFooter(state.ShowFooter).
		ShowScrollbar(state.ShowScrollbar).
		HeaderStyle(style.Style{}.Foreground(style.BrightCyan).Bold(true)).
		TableStyle(style.Style{}.Foreground(style.BrightWhite)).
		BorderStyle(style.Style{}.Foreground(style.BrightBlack)).
		StatusStyle(style.Style{}.Foreground(style.BrightBlack)).
		ScrollbarStyle(style.Style{}.Foreground(style.BrightBlack)).
		SelectedStyle(style.Style{}.Foreground(style.Black).Background(style.BrightCyan).Bold(true)).
		SelectionForField(intent.BindField("checkedRows")).
		PageForField(intent.BindField("currentPage"))

	switch normalizeSelectionMode(state.SelectionMode) {
	case "single":
		builder.SingleSelect()
	case "multi":
		builder.MultiSelect()
	}

	return builder.Build()
}

func sidebar(state AppState, filteredCount int, record IncidentRecord, hasSelected bool) ui.VNode {
	return ui.NewVStack().
		SetGap(1).
		SetChildrenList([]ui.VNode{
			statePanel(state, filteredCount),
			detailPanel(record, hasSelected),
			helpPanel(),
		})
}

func statePanel(state AppState, filteredCount int) ui.VNode {
	return ui.NewVStack().
		SingleBorder("Table State").
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder(fmt.Sprintf("CurrentPage: %d / %d", state.CurrentPage+1, maxInt(1, state.PageCount))).FgColor("bright-white").Build(),
			ui.NewTextBuilder(fmt.Sprintf("FilteredRows: %d / %d", filteredCount, len(demoRecords))).FgColor("bright-white").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Sort: %s", sortSummary(state.SortColumn, state.SortDesc))).FgColor("yellow").Build(),
			ui.NewTextBuilder(fmt.Sprintf("SelectedRow: %s", displaySelectedRow(state.SelectedRow))).FgColor("cyan").Build(),
			ui.NewTextBuilder(fmt.Sprintf("SelectionMode: %s", strings.ToUpper(state.SelectionMode))).FgColor("cyan").Build(),
			ui.NewTextBuilder(fmt.Sprintf("CheckedRows: %s", checkedSummary(checkedIndices(state.CheckedRows)))).FgColor("bright-white").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Scrollbar: %t", state.ShowScrollbar)).FgColor("bright-black").Build(),
			ui.NewTextBuilder("排序与分页现在走受控模式；table 交互会把目标状态同步回 store").FgColor("bright-black").Build(),
		})
}

func detailPanel(record IncidentRecord, hasSelected bool) ui.VNode {
	children := []ui.VNode{
		ui.NewTextBuilder("表格保持原始数据源索引；详情面板直接回看原始记录").FgColor("green").Build(),
	}
	if !hasSelected {
		children = append(children, ui.NewTextBuilder("No row selected").FgColor("bright-black").Build())
	} else {
		children = append(children,
			ui.NewTextBuilder(fmt.Sprintf("ID: #%d", record.ID)).FgColor("yellow").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Service: %s", record.Service)).FgColor("bright-white").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Region: %s", record.Region)).FgColor("bright-white").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Owner: %s", record.Owner)).FgColor("bright-white").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Status: %s", record.Status)).FgColor(statusColor(record.Status)).Build(),
			ui.NewTextBuilder(fmt.Sprintf("P95: %d ms", record.P95MS)).FgColor("bright-white").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Error: %.1f%%", record.ErrorPC)).FgColor("bright-white").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Updated: %s", record.Updated)).FgColor("bright-black").Build(),
		)
	}
	return ui.NewVStack().
		SingleBorder("Selection Details").
		SetGap(0).
		SetChildrenList(children)
}

func helpPanel() ui.VNode {
	return ui.NewVStack().
		SingleBorder("Shortcuts").
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder("Tab: search/status/region/table 之间切换").FgColor("bright-white").Build(),
			ui.NewTextBuilder("Table: 点击表头排序；第三次点击清除排序；↑↓ 选中；Enter/Space 勾选；PageUp/PageDown 或 ←/→ 翻页").FgColor("bright-white").Build(),
			ui.NewTextBuilder("Multi 模式下点击表头 checkbox 可对当前过滤结果全选/反选").FgColor("bright-white").Build(),
			ui.NewTextBuilder("F2 border  F3 footer  F4 scrollbar  F5 clear filters").FgColor("bright-black").Build(),
			ui.NewTextBuilder("F6/F7 page size -/+  F8 none  F9 single  F10 multi").FgColor("bright-black").Build(),
			ui.NewTextBuilder("F11 clear checked  F12 reset  Buttons: First / Prev / Next / Last").FgColor("bright-black").Build(),
		})
}

func newInitialState() AppState {
	initial := AppState{
		SearchText:    "",
		StatusFilter:  "",
		RegionFilter:  "",
		SelectedRow:   "",
		CheckedRows:   "2,7,18",
		SelectionMode: "multi",
		PageSize:      10,
		ShowBorder:    true,
		ShowFooter:    true,
		ShowScrollbar: true,
		CurrentPage:   0,
		PageCount:     1,
		SortColumn:    -1,
		SortDesc:      false,
		FilteredRows:  len(demoRecords),
		LastAction:    "Ready",
	}
	return recomputeDerivedState(initial)
}

func recomputeDerivedState(state AppState) AppState {
	state.FilteredRows = filteredRecordCount(state.SearchText, state.StatusFilter, state.RegionFilter)
	state.PageCount = pageCount(state.FilteredRows, state.PageSize)
	state.SelectionMode = normalizeSelectionMode(state.SelectionMode)
	state.CheckedRows = normalizeCheckedRows(state.CheckedRows, state.SelectionMode)
	if state.CurrentPage >= state.PageCount {
		state.CurrentPage = 0
	}
	if state.PageCount < 1 {
		state.PageCount = 1
	}
	return state
}

func filteredRecordCount(search, statusFilter, regionFilter string) int {
	count := 0
	for _, record := range demoRecords {
		if matchesRecord(record, search, statusFilter, regionFilter) {
			count++
		}
	}
	return count
}

func matchesRecord(record IncidentRecord, search, statusFilter, regionFilter string) bool {
	statusNeedle := strings.ToLower(strings.TrimSpace(statusFilter))
	if statusNeedle != "" && !strings.Contains(strings.ToLower(record.Status), statusNeedle) {
		return false
	}

	regionNeedle := strings.ToLower(strings.TrimSpace(regionFilter))
	if regionNeedle != "" && !strings.Contains(strings.ToLower(record.Region), regionNeedle) {
		return false
	}

	searchNeedle := strings.ToLower(strings.TrimSpace(search))
	if searchNeedle == "" {
		return true
	}

	return strings.Contains(strings.ToLower(strconv.Itoa(record.ID)), searchNeedle) ||
		strings.Contains(strings.ToLower(record.Service), searchNeedle) ||
		strings.Contains(strings.ToLower(record.Region), searchNeedle) ||
		strings.Contains(strings.ToLower(record.Owner), searchNeedle) ||
		strings.Contains(strings.ToLower(record.Status), searchNeedle) ||
		strings.Contains(strings.ToLower(record.Updated), searchNeedle) ||
		strings.Contains(strings.ToLower(strconv.Itoa(record.P95MS)), searchNeedle) ||
		strings.Contains(strings.ToLower(fmt.Sprintf("%.1f", record.ErrorPC)), searchNeedle)
}

func selectedRecord(raw string) (IncidentRecord, bool) {
	index, ok := selectedIndex(raw)
	if !ok || index < 0 || index >= len(demoRecords) {
		return IncidentRecord{}, false
	}
	return demoRecords[index], true
}

func selectedIndex(raw string) (int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return -1, false
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 || value >= len(demoRecords) {
		return -1, false
	}
	return value, true
}

func normalizeSelectedRow(raw string) string {
	value, ok := selectedIndex(raw)
	if !ok {
		return ""
	}
	return strconv.Itoa(value)
}

func displaySelectedRow(raw string) string {
	if value, ok := selectedIndex(raw); ok {
		return strconv.Itoa(value)
	}
	return "(none)"
}

func checkedIndices(raw string) []int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	indices := make([]int, 0, len(parts))
	seen := map[int]struct{}{}
	for _, part := range parts {
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || value < 0 || value >= len(demoRecords) {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		indices = append(indices, value)
	}
	return indices
}

func checkedSummary(indices []int) string {
	if len(indices) == 0 {
		return "[]"
	}
	parts := make([]string, len(indices))
	for index, value := range indices {
		parts[index] = strconv.Itoa(value)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func normalizeSelectionMode(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "single":
		return "single"
	case "multi", "multiple":
		return "multi"
	default:
		return "none"
	}
}

func normalizeCheckedRows(raw, mode string) string {
	indices := checkedIndices(raw)
	switch normalizeSelectionMode(mode) {
	case "none":
		return ""
	case "single":
		if len(indices) == 0 {
			return ""
		}
		return strconv.Itoa(indices[0])
	default:
		if len(indices) == 0 {
			return ""
		}
		parts := make([]string, len(indices))
		for index, value := range indices {
			parts[index] = strconv.Itoa(value)
		}
		return strings.Join(parts, ",")
	}
}

func sortSummary(columnIndex int, descending bool) string {
	if columnIndex < 0 || columnIndex >= len(demoColumns) {
		return "none"
	}
	direction := "asc"
	if descending {
		direction = "desc"
	}
	return fmt.Sprintf("%s (%s)", demoColumns[columnIndex].Title, direction)
}

func statusColor(status string) style.Color {
	switch strings.ToLower(status) {
	case "open":
		return style.BrightRed
	case "triage":
		return style.Yellow
	case "mitigated":
		return style.BrightCyan
	default:
		return style.BrightGreen
	}
}

func buildRows(records []IncidentRecord) [][]string {
	rows := make([][]string, len(records))
	for index, record := range records {
		rows[index] = []string{
			strconv.Itoa(record.ID),
			record.Service,
			record.Region,
			record.Owner,
			record.Status,
			strconv.Itoa(record.P95MS),
			fmt.Sprintf("%.1f", record.ErrorPC),
			record.Updated,
		}
	}
	return rows
}

func generateRecords() []IncidentRecord {
	services := []string{"gateway", "billing", "search", "worker", "notifier", "identity", "catalog", "reporting"}
	regions := []string{"us-east", "us-west", "eu-west", "ap-south"}
	owners := []string{"team-core", "team-pay", "team-growth", "team-ops", "team-data"}
	statuses := []string{"open", "triage", "mitigated", "closed"}

	records := make([]IncidentRecord, 96)
	for index := range records {
		records[index] = IncidentRecord{
			ID:      1000 + index,
			Service: services[index%len(services)],
			Region:  regions[(index/2)%len(regions)],
			Owner:   owners[(index/3)%len(owners)],
			Status:  statuses[(index/5)%len(statuses)],
			P95MS:   45 + (index*19)%420,
			ErrorPC: float64((index*7)%125) / 10.0,
			Updated: fmt.Sprintf("%02dh ago", index%24),
		}
	}
	return records
}

func pageCount(totalRows, pageSize int) int {
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

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
