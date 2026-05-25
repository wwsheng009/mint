// Package table provides Fiber-first Table data display component.
package table

import (
	"strings"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Prop Keys
// =============================================================================

// Prop key constants — shared by VNode and Instance to avoid magic strings.
const (
	propBorderStyle              = "borderStyle"
	propChangeIntent             = "changeIntent"
	propChangeIntentField        = "changeIntentField"
	propCheckedIndices           = "checkedIndices"
	propCheckedIndicesControlled = "checkedIndicesControlled"
	propColumns                  = "columns"
	propComponentID              = "componentID"
	propCurrentPage              = "currentPage"
	propCurrentPageControlled    = "currentPageControlled"
	propEmptyText                = "emptyText"
	propExpandIntent             = "expandIntent"
	propExpandIntentField        = "expandIntentField"
	propExpandedContent          = "expandedContent"
	propExpandedControlled       = "expandedControlled"
	propExpandedIndices          = "expandedIndices"
	propFilterStyle              = "filterStyle"
	propFilters                  = "filters"
	propGap                      = "gap"
	propHeaderStyle              = "headerStyle"
	propKey                      = "key"
	propPageIntentField          = "pageIntentField"
	propPageSize                 = "pageSize"
	propRows                     = "rows"
	propScrollbarStyle           = "scrollbarStyle"
	propSearchQuery              = "searchQuery"
	propSelectedIndex            = "selectedIndex"
	propSelectedIndexControlled  = "selectedIndexControlled"
	propSelectedStyle            = "selectedStyle"
	propSelectionIntent          = "selectionIntent"
	propSelectionIntentField     = "selectionIntentField"
	propSelectionMode            = "selectionMode"
	propShowBorder               = "showBorder"
	propShowFooter               = "showFooter"
	propShowScrollbar            = "showScrollbar"
	propSortColumn               = "sortColumn"
	propSortControlled           = "sortControlled"
	propSortDescending           = "sortDescending"
	propStatusStyle              = "statusStyle"
	propStatusText               = "statusText"
	propTableStyle               = "tableStyle"
	propTreeParents              = "treeParents"
)

// =============================================================================
// Types
// =============================================================================

// TableColumn represents a column in a table
type TableColumn struct {
	Title        string
	Width        int
	WidthPercent int
	MinWidth     int
	MaxWidth     int
	FixedLeft    bool
	FixedRight   bool
	Sortable     bool
	Align        rtui.Align
}

// =============================================================================
// VNode - Pure Description (No State, No Closures, No Paint)
// =============================================================================

// VNode is the table component description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key         string
	componentID string

	// === Table Props ===
	columns        []TableColumn
	rows           [][]string
	emptyText      string
	headerStyle    style.Style
	tableStyle     style.Style
	selectedStyle  style.Style
	borderStyle    style.Style
	statusStyle    style.Style
	statusText     string
	filterStyle    style.Style
	scrollbarStyle style.Style

	// === Layout Props ===
	gap             int
	showBorder      bool
	showFooter      bool
	showScrollbar   bool
	pageSize        int
	searchQuery     string
	filters         map[int]string
	expandedContent map[int]string
	treeParents     map[int]int

	currentPage              int
	currentPageControlled    bool
	expandedIndices          []int
	expandedControlled       bool
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

	changeIntent      intent.Intent
	changeIntentField intent.FieldIntent
	pageIntentField   intent.FieldIntent
	expandIntent      intent.Intent
	expandIntentField intent.FieldIntent
}

// Ensure VNode implements required interfaces
var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// New creates a new table VNode.
func New() *VNode {
	return &VNode{
		ElementVNode:      rtui.NewElement("table"),
		columns:           []TableColumn{},
		rows:              [][]string{},
		emptyText:         "(empty)",
		headerStyle:       style.Style{}.Bold(true),
		tableStyle:        style.Style{},
		selectedStyle:     style.Style{}.Reverse(true),
		borderStyle:       style.Style{}.Foreground(style.BrightBlack),
		statusStyle:       style.Style{}.Foreground(style.BrightBlack),
		filterStyle:       style.Style{}.Foreground(style.BrightBlack),
		scrollbarStyle:    style.Style{}.Foreground(style.BrightBlack),
		gap:               0,
		showBorder:        false,
		showFooter:        true,
		showScrollbar:     true,
		pageSize:          0,
		searchQuery:       "",
		filters:           map[int]string{},
		expandedContent:   map[int]string{},
		treeParents:       map[int]int{},
		expandedIndices:   nil,
		currentPage:       0,
		sortColumn:        -1,
		selectedIndex:     -1,
		selectionMode:     SelectionNone,
		checkedIndices:    nil,
		changeIntentField: nil,
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

func (v *VNode) Key() string                  { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode { v.key = key; return v }
func (v *VNode) Tag() string                  { return "table" }
func (v *VNode) Type() rtui.VNodeType         { return rtui.VNodeElement }

func (v *VNode) Children() []rtui.VNode {
	// Table is a leaf component - no children
	return []rtui.VNode{}
}

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	// Table is display-only, has no children
	return v
}

func (v *VNode) GetLayer() rtui.Layer             { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Style() style.Style                { return v.tableStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode { v.tableStyle = s; return v }

func (v *VNode) TextContent() string { return "" }

func (v *VNode) Props() rtui.Props {
	props := rtui.Props{
		propKey:                     v.key,
		propComponentID:             v.componentID,
		propColumns:                 v.columns,
		propRows:                    v.rows,
		propEmptyText:               v.emptyText,
		propHeaderStyle:             v.headerStyle,
		propTableStyle:              v.tableStyle,
		propSelectedStyle:           v.selectedStyle,
		propBorderStyle:             v.borderStyle,
		propStatusStyle:             v.statusStyle,
		propStatusText:              v.statusText,
		propFilterStyle:             v.filterStyle,
		propScrollbarStyle:          v.scrollbarStyle,
		propGap:                     v.gap,
		propShowBorder:              v.showBorder,
		propShowFooter:              v.showFooter,
		propShowScrollbar:           v.showScrollbar,
		propPageSize:                v.pageSize,
		propSearchQuery:             v.searchQuery,
		propFilters:                 v.filters,
		propExpandedContent:         cloneFilters(v.expandedContent),
		propTreeParents:             cloneIntMap(v.treeParents),
		propCurrentPage:             v.currentPage,
		propCurrentPageControlled:   v.currentPageControlled,
		propSortColumn:              v.sortColumn,
		propSortDescending:          v.sortDescending,
		propSortControlled:          v.sortControlled,
		propSelectedIndex:           v.selectedIndex,
		propSelectedIndexControlled: v.selectedIndexControlled,
		propSelectionIntent:         v.selectionIntent,
		propSelectionMode:           v.selectionMode,
		propChangeIntent:            v.changeIntent,
		propChangeIntentField:       v.changeIntentField,
		propPageIntentField:         v.pageIntentField,
		propExpandIntent:            v.expandIntent,
		propExpandIntentField:       v.expandIntentField,
	}
	if v.selectionIntentField != nil {
		props[propSelectionIntentField] = v.selectionIntentField
	}
	if v.expandedControlled {
		props[propExpandedControlled] = true
		props[propExpandedIndices] = append([]int(nil), v.expandedIndices...)
	}
	if v.checkedIndicesControlled {
		props[propCheckedIndicesControlled] = true
		props[propCheckedIndices] = append([]int(nil), v.checkedIndices...)
	}
	return props
}

func (v *VNode) SetProps(p rtui.Props) rtui.VNode {
	if val, ok := p[propKey].(string); ok {
		v.key = val
	}
	if val, ok := p[propComponentID].(string); ok {
		v.componentID = val
	}
	if val, ok := p[propColumns].([]TableColumn); ok {
		v.columns = val
	}
	if val, ok := p[propRows].([][]string); ok {
		v.rows = val
	}
	if val, ok := p[propEmptyText].(string); ok {
		v.emptyText = val
	}
	if val, ok := p[propHeaderStyle].(style.Style); ok {
		v.headerStyle = val
	}
	if val, ok := p[propTableStyle].(style.Style); ok {
		v.tableStyle = val
	}
	if val, ok := p[propSelectedStyle].(style.Style); ok {
		v.selectedStyle = val
	}
	if val, ok := p[propBorderStyle].(style.Style); ok {
		v.borderStyle = val
	}
	if val, ok := p[propStatusStyle].(style.Style); ok {
		v.statusStyle = val
	}
	if val, ok := p[propStatusText].(string); ok {
		v.statusText = val
	}
	if val, ok := p[propFilterStyle].(style.Style); ok {
		v.filterStyle = val
	}
	if val, ok := p[propScrollbarStyle].(style.Style); ok {
		v.scrollbarStyle = val
	}
	if val, ok := p[propGap].(int); ok {
		v.gap = val
	}
	if val, ok := p[propShowBorder].(bool); ok {
		v.showBorder = val
	}
	if val, ok := p[propShowFooter].(bool); ok {
		v.showFooter = val
	}
	if val, ok := p[propShowScrollbar].(bool); ok {
		v.showScrollbar = val
	}
	if val, ok := p[propPageSize].(int); ok {
		v.pageSize = val
	}
	if val, ok := p[propSearchQuery].(string); ok {
		v.searchQuery = val
	}
	if val, ok := p[propFilters].(map[int]string); ok {
		v.filters = cloneFilters(val)
	}
	if val, ok := p[propExpandedContent].(map[int]string); ok {
		v.expandedContent = cloneFilters(val)
	}
	if val, ok := p[propTreeParents].(map[int]int); ok {
		v.treeParents = cloneIntMap(val)
	}
	if val, ok := p[propCurrentPage].(int); ok {
		v.currentPage = val
	}
	if val, ok := p[propCurrentPageControlled].(bool); ok {
		v.currentPageControlled = val
	}
	if val, ok := p[propExpandedIndices].([]int); ok {
		v.expandedIndices = append([]int(nil), val...)
		v.expandedControlled = true
	}
	if val, ok := p[propExpandedControlled].(bool); ok {
		v.expandedControlled = val
	}
	if val, ok := p[propSortColumn].(int); ok {
		v.sortColumn = val
	}
	if val, ok := p[propSortDescending].(bool); ok {
		v.sortDescending = val
	}
	if val, ok := p[propSortControlled].(bool); ok {
		v.sortControlled = val
	}
	if val, ok := p[propSelectedIndex].(int); ok {
		v.selectedIndex = val
	}
	if val, ok := p[propSelectedIndexControlled].(bool); ok {
		v.selectedIndexControlled = val
	}
	if val, ok := p[propSelectionIntent].(intent.Intent); ok {
		v.selectionIntent = val
	}
	if val, ok := p[propSelectionIntentField].(intent.FieldIntent); ok {
		v.selectionIntentField = val
	}
	if val, ok := p[propSelectionMode].(SelectionMode); ok {
		v.selectionMode = val
	}
	if val, ok := p[propCheckedIndices].([]int); ok {
		v.checkedIndices = append([]int(nil), val...)
		v.checkedIndicesControlled = true
	}
	if val, ok := p[propCheckedIndicesControlled].(bool); ok {
		v.checkedIndicesControlled = val
	}
	if val, ok := p[propChangeIntent].(intent.Intent); ok {
		v.changeIntent = val
	}
	if val, ok := p[propChangeIntentField].(intent.FieldIntent); ok {
		v.changeIntentField = val
	}
	if val, ok := p[propPageIntentField].(intent.FieldIntent); ok {
		v.pageIntentField = val
	}
	if val, ok := p[propExpandIntent].(intent.Intent); ok {
		v.expandIntent = val
	}
	if val, ok := p[propExpandIntentField].(intent.FieldIntent); ok {
		v.expandIntentField = val
	}
	return v
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

// =============================================================================
// Fluent Setters
// =============================================================================

func (v *VNode) SetColumns(cols []TableColumn) *VNode  { v.columns = cols; return v }
func (v *VNode) SetComponentID(id string) *VNode       { v.componentID = id; return v }
func (v *VNode) SetRows(rows [][]string) *VNode        { v.rows = rows; return v }
func (v *VNode) SetEmptyText(text string) *VNode       { v.emptyText = text; return v }
func (v *VNode) SetHeaderStyle(s style.Style) *VNode   { v.headerStyle = s; return v }
func (v *VNode) SetGap(gap int) *VNode                 { v.gap = gap; return v }
func (v *VNode) SetShowBorder(show bool) *VNode        { v.showBorder = show; return v }
func (v *VNode) SetShowFooter(show bool) *VNode        { v.showFooter = show; return v }
func (v *VNode) SetShowScrollbar(show bool) *VNode     { v.showScrollbar = show; return v }
func (v *VNode) SetBorderStyle(s style.Style) *VNode   { v.borderStyle = s; return v }
func (v *VNode) SetSelectedStyle(s style.Style) *VNode { v.selectedStyle = s; return v }
func (v *VNode) SetStatusStyle(s style.Style) *VNode   { v.statusStyle = s; return v }
func (v *VNode) SetStatusText(text string) *VNode      { v.statusText = strings.TrimSpace(text); return v }
func (v *VNode) SetFilterStyle(s style.Style) *VNode   { v.filterStyle = s; return v }
func (v *VNode) SetScrollbarStyle(s style.Style) *VNode {
	v.scrollbarStyle = s
	return v
}
func (v *VNode) SetPageSize(pageSize int) *VNode    { v.pageSize = pageSize; return v }
func (v *VNode) SetSearchQuery(query string) *VNode { v.searchQuery = query; return v }
func (v *VNode) SetFilters(filters map[int]string) *VNode {
	v.filters = cloneFilters(filters)
	return v
}
func (v *VNode) SetExpandedContent(content map[int]string) *VNode {
	v.expandedContent = cloneFilters(content)
	return v
}
func (v *VNode) SetTreeParents(parents map[int]int) *VNode {
	v.treeParents = cloneIntMap(parents)
	return v
}
func (v *VNode) SetTreeParent(index int, parentIndex int) *VNode {
	if v.treeParents == nil {
		v.treeParents = make(map[int]int)
	}
	v.treeParents[index] = parentIndex
	return v
}
func (v *VNode) SetExpandedRow(index int, content string) *VNode {
	if v.expandedContent == nil {
		v.expandedContent = make(map[int]string)
	}
	if strings.TrimSpace(content) == "" {
		delete(v.expandedContent, index)
	} else {
		v.expandedContent[index] = content
	}
	return v
}
func (v *VNode) SetExpandedIndices(indices []int) *VNode {
	v.expandedIndices = append([]int(nil), indices...)
	v.expandedControlled = true
	return v
}
func (v *VNode) SetFilter(columnIndex int, value string) *VNode {
	if v.filters == nil {
		v.filters = make(map[int]string)
	}
	if value == "" {
		delete(v.filters, columnIndex)
	} else {
		v.filters[columnIndex] = value
	}
	return v
}
func (v *VNode) SetCurrentPage(page int) *VNode {
	v.currentPage = page
	v.currentPageControlled = true
	return v
}
func (v *VNode) SetSortBy(columnIndex int, descending bool) *VNode {
	v.sortColumn = columnIndex
	v.sortDescending = descending
	v.sortControlled = true
	return v
}
func (v *VNode) SetSelectedIndex(index int) *VNode {
	v.selectedIndex = index
	v.selectedIndexControlled = true
	return v
}
func (v *VNode) SetSelectionIntent(i intent.Intent) *VNode {
	v.selectionIntent = i
	return v
}
func (v *VNode) SetSelectionFieldIntent(i intent.FieldIntent) *VNode {
	v.selectionIntentField = i
	return v
}
func (v *VNode) SetSelectionMode(mode SelectionMode) *VNode {
	v.selectionMode = mode
	return v
}
func (v *VNode) SetCheckedIndices(indices []int) *VNode {
	v.checkedIndices = append([]int(nil), indices...)
	v.checkedIndicesControlled = true
	return v
}
func (v *VNode) SetIntent(i intent.Intent) *VNode {
	v.changeIntent = i
	return v
}
func (v *VNode) SetFieldIntent(i intent.FieldIntent) *VNode {
	v.changeIntentField = i
	return v
}
func (v *VNode) SetPageFieldIntent(i intent.FieldIntent) *VNode {
	v.pageIntentField = i
	return v
}
func (v *VNode) SetExpandIntent(i intent.Intent) *VNode {
	v.expandIntent = i
	return v
}
func (v *VNode) SetExpandFieldIntent(i intent.FieldIntent) *VNode {
	v.expandIntentField = i
	return v
}

func (v *VNode) AddRow(row ...string) *VNode {
	v.rows = append(v.rows, row)
	return v
}

// =============================================================================
// Accessors
// =============================================================================

func (v *VNode) Columns() []TableColumn   { return v.columns }
func (v *VNode) Rows() [][]string         { return v.rows }
func (v *VNode) EmptyText() string        { return v.emptyText }
func (v *VNode) HeaderStyle() style.Style { return v.headerStyle }
func (v *VNode) TableStyle() style.Style  { return v.tableStyle }
func (v *VNode) Gap() int                 { return v.gap }

func cloneFilters(filters map[int]string) map[int]string {
	if len(filters) == 0 {
		return map[int]string{}
	}
	cloned := make(map[int]string, len(filters))
	for key, value := range filters {
		cloned[key] = value
	}
	return cloned
}

func cloneIntMap(values map[int]int) map[int]int {
	if len(values) == 0 {
		return map[int]int{}
	}
	cloned := make(map[int]int, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}
