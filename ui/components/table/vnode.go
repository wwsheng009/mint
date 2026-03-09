// Package table provides Fiber-first Table data display component.
package table

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Types
// =============================================================================

// TableColumn represents a column in a table
type TableColumn struct {
	Title    string
	Width    int
	Sortable bool
	Align    rtui.Align
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
	scrollbarStyle style.Style

	// === Layout Props ===
	gap           int
	showBorder    bool
	showFooter    bool
	showScrollbar bool
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

	changeIntent      intent.Intent
	changeIntentField intent.FieldIntent
	pageIntentField   intent.FieldIntent
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
		scrollbarStyle:    style.Style{}.Foreground(style.BrightBlack),
		gap:               0,
		showBorder:        false,
		showFooter:        true,
		showScrollbar:     true,
		pageSize:          0,
		searchQuery:       "",
		filters:           map[int]string{},
		currentPage:       0,
		sortColumn:        -1,
		selectedIndex:     -1,
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
	return rtui.Props{
		"key":                     v.key,
		"componentID":             v.componentID,
		"columns":                 v.columns,
		"rows":                    v.rows,
		"emptyText":               v.emptyText,
		"headerStyle":             v.headerStyle,
		"tableStyle":              v.tableStyle,
		"selectedStyle":           v.selectedStyle,
		"borderStyle":             v.borderStyle,
		"statusStyle":             v.statusStyle,
		"scrollbarStyle":          v.scrollbarStyle,
		"gap":                     v.gap,
		"showBorder":              v.showBorder,
		"showFooter":              v.showFooter,
		"showScrollbar":           v.showScrollbar,
		"pageSize":                v.pageSize,
		"searchQuery":             v.searchQuery,
		"filters":                 v.filters,
		"currentPage":             v.currentPage,
		"currentPageControlled":   v.currentPageControlled,
		"sortColumn":              v.sortColumn,
		"sortDescending":          v.sortDescending,
		"sortControlled":          v.sortControlled,
		"selectedIndex":           v.selectedIndex,
		"selectedIndexControlled": v.selectedIndexControlled,
		"changeIntent":            v.changeIntent,
		"changeIntentField":       v.changeIntentField,
		"pageIntentField":         v.pageIntentField,
	}
}

func (v *VNode) SetProps(p rtui.Props) rtui.VNode {
	if val, ok := p["key"].(string); ok {
		v.key = val
	}
	if val, ok := p["componentID"].(string); ok {
		v.componentID = val
	}
	if val, ok := p["columns"].([]TableColumn); ok {
		v.columns = val
	}
	if val, ok := p["rows"].([][]string); ok {
		v.rows = val
	}
	if val, ok := p["emptyText"].(string); ok {
		v.emptyText = val
	}
	if val, ok := p["headerStyle"].(style.Style); ok {
		v.headerStyle = val
	}
	if val, ok := p["tableStyle"].(style.Style); ok {
		v.tableStyle = val
	}
	if val, ok := p["selectedStyle"].(style.Style); ok {
		v.selectedStyle = val
	}
	if val, ok := p["borderStyle"].(style.Style); ok {
		v.borderStyle = val
	}
	if val, ok := p["statusStyle"].(style.Style); ok {
		v.statusStyle = val
	}
	if val, ok := p["scrollbarStyle"].(style.Style); ok {
		v.scrollbarStyle = val
	}
	if val, ok := p["gap"].(int); ok {
		v.gap = val
	}
	if val, ok := p["showBorder"].(bool); ok {
		v.showBorder = val
	}
	if val, ok := p["showFooter"].(bool); ok {
		v.showFooter = val
	}
	if val, ok := p["showScrollbar"].(bool); ok {
		v.showScrollbar = val
	}
	if val, ok := p["pageSize"].(int); ok {
		v.pageSize = val
	}
	if val, ok := p["searchQuery"].(string); ok {
		v.searchQuery = val
	}
	if val, ok := p["filters"].(map[int]string); ok {
		v.filters = cloneFilters(val)
	}
	if val, ok := p["currentPage"].(int); ok {
		v.currentPage = val
	}
	if val, ok := p["currentPageControlled"].(bool); ok {
		v.currentPageControlled = val
	}
	if val, ok := p["sortColumn"].(int); ok {
		v.sortColumn = val
	}
	if val, ok := p["sortDescending"].(bool); ok {
		v.sortDescending = val
	}
	if val, ok := p["sortControlled"].(bool); ok {
		v.sortControlled = val
	}
	if val, ok := p["selectedIndex"].(int); ok {
		v.selectedIndex = val
	}
	if val, ok := p["selectedIndexControlled"].(bool); ok {
		v.selectedIndexControlled = val
	}
	if val, ok := p["changeIntent"].(intent.Intent); ok {
		v.changeIntent = val
	}
	if val, ok := p["changeIntentField"].(intent.FieldIntent); ok {
		v.changeIntentField = val
	}
	if val, ok := p["pageIntentField"].(intent.FieldIntent); ok {
		v.pageIntentField = val
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
