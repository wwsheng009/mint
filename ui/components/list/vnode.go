package list

import (
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Prop Keys
// =============================================================================

// Prop key constants — shared by VNode and Instance to avoid magic strings.
const (
	propAllowScroll = "allowScroll"
	propBorderStyle = "borderStyle"
	propChangeIntent = "changeIntent"
	propCheckedIndices = "checkedIndices"
	propCheckedIndicesControlled = "checkedIndicesControlled"
	propComponentID = "componentID"
	propEmptyText = "emptyText"
	propFormID = "formID"
	propHeader = "header"
	propHeaderStyle = "headerStyle"
	propKey = "key"
	propMatchStyle = "matchStyle"
	propMaxRows = "maxRows"
	propRowStyle = "rowStyle"
	propRowStyleFn = "rowStyleFn"
	propRows = "rows"
	propScrollOffset = "scrollOffset"
	propScrollOffsetControlled = "scrollOffsetControlled"
	propScrollbarStyle = "scrollbarStyle"
	propSearchFn = "searchFn"
	propSearchQuery = "searchQuery"
	propSearchStatsStyle = "searchStatsStyle"
	propSelectedIndex = "selectedIndex"
	propSelectedIndexControlled = "selectedIndexControlled"
	propSelectedStyle = "selectedStyle"
	propSelectionIntent = "selectionIntent"
	propSelectionMode = "selectionMode"
	propSeparatorChar = "separatorChar"
	propShowBorder = "showBorder"
	propShowScrollbar = "showScrollbar"
	propShowSearchStats = "showSearchStats"
	propShowSeparator = "showSeparator"
	propViewportHeight = "viewportHeight"
)

// =============================================================================
// VNode - Pure Description
// =============================================================================

// VNode is the pure description for the List component
// Contains only declarative properties - no state, no closures, no Paint
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key         string
	componentID string

	// === Visual Properties ===
	header     string   // Optional column header text
	rows       []string // Pre-formatted data rows
	emptyText  string   // Text shown when rows is empty
	maxRows    int      // Maximum visible rows (0 = unlimited)
	showBorder bool     // Show border around the list

	// === Separator ===
	showSeparator bool // Show separator line between header and rows
	separatorChar rune // Separator character (default '─')

	// === Styles ===
	headerStyle      style.Style                   // Style for the header row
	rowStyle         style.Style                   // Default style for data rows
	rowStyleFn       func(int, string) style.Style // Dynamic style function per row
	matchStyle       style.Style                   // Style for matched search rows
	selectedStyle    style.Style                   // Style for the selected row
	borderStyle      style.Style                   // Style for the border
	showScrollbar    bool
	scrollbarStyle   style.Style
	changeIntent     intent.Intent
	selectionIntent  intent.Intent
	selectionMode    SelectionMode
	searchQuery      string
	searchFn         func(string, string) bool
	showSearchStats  bool
	searchStatsStyle style.Style

	// === State Properties (declarative initial state) ===
	scrollOffset             int  // Initial scroll offset
	scrollOffsetControlled   bool // Whether scrollOffset is externally controlled
	selectedIndex            int  // Currently selected row index
	selectedIndexControlled  bool // Whether selectedIndex is externally controlled
	checkedIndices           []int
	checkedIndicesControlled bool
	viewportHeight           int // Visible height for scrolling
	formID                   string

	// === Interaction ===
	allowScroll bool // Whether scrolling is enabled
}

// Ensure VNode implements required interfaces
var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// New creates a new List VNode
func New() *VNode {
	return &VNode{
		ElementVNode:    rtui.NewElement("list"),
		key:             "",
		componentID:     "",
		header:          "",
		rows:            []string{},
		emptyText:       "(empty)",
		maxRows:         0,
		showBorder:      true,
		showSeparator:   true,
		separatorChar:   '─',
		headerStyle:     style.Style{}.Bold(true),
		rowStyle:        style.Style{},
		matchStyle:      style.Style{},
		selectedStyle:   style.Style{BG: style.Blue, FG: style.White},
		borderStyle:     style.Style{FG: style.White},
		showScrollbar:   true,
		changeIntent:    nil,
		selectionIntent: nil,
		selectionMode:   SelectionNone,
		searchQuery:     "",
		showSearchStats: false,
		scrollOffset:    0,
		selectedIndex:   -1,
		checkedIndices:  nil,
		viewportHeight:  10,
		formID:          "",
		allowScroll:     true,
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

func (v *VNode) Key() string                                  { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode                 { v.key = key; return v }
func (v *VNode) Tag() string                                  { return "list" }
func (v *VNode) Style() style.Style                           { return v.rowStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode            { v.rowStyle = s; return v }
func (v *VNode) Children() []rtui.VNode                       { return nil }
func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }
func (v *VNode) GetLayer() rtui.Layer                         { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode             { return v }

func (v *VNode) Props() rtui.Props {
	props := rtui.Props{
		propKey:                     v.key,
		propComponentID:             v.componentID,
		propHeader:                  v.header,
		propRows:                    v.rows,
		propEmptyText:               v.emptyText,
		propMaxRows:                 v.maxRows,
		propShowBorder:              v.showBorder,
		propShowSeparator:           v.showSeparator,
		propSeparatorChar:           v.separatorChar,
		propHeaderStyle:             v.headerStyle,
		propRowStyle:                v.rowStyle,
		propMatchStyle:              v.matchStyle,
		propSelectedStyle:           v.selectedStyle,
		propBorderStyle:             v.borderStyle,
		propShowScrollbar:           v.showScrollbar,
		propScrollbarStyle:          v.scrollbarStyle,
		propChangeIntent:            v.changeIntent,
		propSelectionIntent:         v.selectionIntent,
		propSelectionMode:           v.selectionMode,
		propSearchQuery:             v.searchQuery,
		propShowSearchStats:         v.showSearchStats,
		propSearchStatsStyle:        v.searchStatsStyle,
		propScrollOffsetControlled:  v.scrollOffsetControlled,
		propScrollOffset:            v.scrollOffset,
		propSelectedIndex:           v.selectedIndex,
		propSelectedIndexControlled: v.selectedIndexControlled,
		propCheckedIndices:          append([]int(nil), v.checkedIndices...),
		propViewportHeight:          v.viewportHeight,
		propFormID:                  v.formID,
		propAllowScroll:             v.allowScroll,
	}
	props[propCheckedIndicesControlled] = v.checkedIndicesControlled

	// Add rowStyleFn if it's set (functions can be stored in Props)
	if v.rowStyleFn != nil {
		props[propRowStyleFn] = v.rowStyleFn
	}
	if v.searchFn != nil {
		props[propSearchFn] = v.searchFn
	}

	return props
}

func (v *VNode) SetProps(p rtui.Props) rtui.VNode {
	if key, ok := p[propKey].(string); ok {
		v.key = key
	}
	if componentID, ok := p[propComponentID].(string); ok {
		v.componentID = componentID
	}
	if header, ok := p[propHeader].(string); ok {
		v.header = header
	}
	if rows, ok := p[propRows].([]string); ok {
		v.rows = rows
	}
	if emptyText, ok := p[propEmptyText].(string); ok {
		v.emptyText = emptyText
	}
	if maxRows, ok := p[propMaxRows].(int); ok {
		v.maxRows = maxRows
	}
	if showBorder, ok := p[propShowBorder].(bool); ok {
		v.showBorder = showBorder
	}
	if showSeparator, ok := p[propShowSeparator].(bool); ok {
		v.showSeparator = showSeparator
	}
	if separatorChar, ok := p[propSeparatorChar].(rune); ok {
		v.separatorChar = separatorChar
	}
	if headerStyle, ok := p[propHeaderStyle].(style.Style); ok {
		v.headerStyle = headerStyle
	}
	if rowStyle, ok := p[propRowStyle].(style.Style); ok {
		v.rowStyle = rowStyle
	}
	if matchStyle, ok := p[propMatchStyle].(style.Style); ok {
		v.matchStyle = matchStyle
	}
	if selectedStyle, ok := p[propSelectedStyle].(style.Style); ok {
		v.selectedStyle = selectedStyle
	}
	if borderStyle, ok := p[propBorderStyle].(style.Style); ok {
		v.borderStyle = borderStyle
	}
	if showScrollbar, ok := p[propShowScrollbar].(bool); ok {
		v.showScrollbar = showScrollbar
	}
	if scrollbarStyle, ok := p[propScrollbarStyle].(style.Style); ok {
		v.scrollbarStyle = scrollbarStyle
	}
	if changeIntent, ok := p[propChangeIntent].(intent.Intent); ok {
		v.changeIntent = changeIntent
	}
	if selectionIntent, ok := p[propSelectionIntent].(intent.Intent); ok {
		v.selectionIntent = selectionIntent
	}
	if selectionMode, ok := p[propSelectionMode].(SelectionMode); ok {
		v.selectionMode = selectionMode
	}
	if searchQuery, ok := p[propSearchQuery].(string); ok {
		v.searchQuery = searchQuery
	}
	if searchFn, ok := p[propSearchFn].(func(string, string) bool); ok {
		v.searchFn = searchFn
	}
	if showSearchStats, ok := p[propShowSearchStats].(bool); ok {
		v.showSearchStats = showSearchStats
	}
	if searchStatsStyle, ok := p[propSearchStatsStyle].(style.Style); ok {
		v.searchStatsStyle = searchStatsStyle
	}
	if scrollOffset, ok := p[propScrollOffset].(int); ok {
		v.scrollOffset = scrollOffset
	}
	if controlled, ok := p[propScrollOffsetControlled].(bool); ok {
		v.scrollOffsetControlled = controlled
	}
	if selectedIndex, ok := p[propSelectedIndex].(int); ok {
		v.selectedIndex = selectedIndex
	}
	if controlled, ok := p[propSelectedIndexControlled].(bool); ok {
		v.selectedIndexControlled = controlled
	}
	if checkedIndices, ok := p[propCheckedIndices].([]int); ok {
		v.checkedIndices = append([]int(nil), checkedIndices...)
	}
	if controlled, ok := p[propCheckedIndicesControlled].(bool); ok {
		v.checkedIndicesControlled = controlled
	}
	if viewportHeight, ok := p[propViewportHeight].(int); ok {
		v.viewportHeight = viewportHeight
	}
	if formID, ok := p[propFormID].(string); ok {
		v.formID = formID
	}
	if allowScroll, ok := p[propAllowScroll].(bool); ok {
		v.allowScroll = allowScroll
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
// Setter Methods (Fluent API)
// =============================================================================

func (v *VNode) SetComponentID(id string) *VNode                       { v.componentID = id; return v }
func (v *VNode) SetHeader(header string) *VNode                        { v.header = header; return v }
func (v *VNode) SetRows(rows []string) *VNode                          { v.rows = rows; return v }
func (v *VNode) SetEmptyText(text string) *VNode                       { v.emptyText = text; return v }
func (v *VNode) SetMaxRows(n int) *VNode                               { v.maxRows = n; return v }
func (v *VNode) SetShowBorder(show bool) *VNode                        { v.showBorder = show; return v }
func (v *VNode) SetShowSeparator(show bool) *VNode                     { v.showSeparator = show; return v }
func (v *VNode) SetSeparatorChar(ch rune) *VNode                       { v.separatorChar = ch; return v }
func (v *VNode) SetHeaderStyle(s style.Style) *VNode                   { v.headerStyle = s; return v }
func (v *VNode) SetRowStyle(s style.Style) *VNode                      { v.rowStyle = s; return v }
func (v *VNode) SetRowStyleFn(fn func(int, string) style.Style) *VNode { v.rowStyleFn = fn; return v }
func (v *VNode) SetMatchStyle(s style.Style) *VNode                    { v.matchStyle = s; return v }
func (v *VNode) SetSelectedStyle(s style.Style) *VNode                 { v.selectedStyle = s; return v }
func (v *VNode) SetBorderStyle(s style.Style) *VNode                   { v.borderStyle = s; return v }
func (v *VNode) SetShowScrollbar(show bool) *VNode                     { v.showScrollbar = show; return v }
func (v *VNode) SetScrollbarStyle(s style.Style) *VNode                { v.scrollbarStyle = s; return v }
func (v *VNode) SetChangeIntent(changeIntent intent.Intent) *VNode {
	v.changeIntent = changeIntent
	return v
}
func (v *VNode) SetSelectionIntent(selectionIntent intent.Intent) *VNode {
	v.selectionIntent = selectionIntent
	return v
}
func (v *VNode) SetSelectionMode(mode SelectionMode) *VNode { v.selectionMode = mode; return v }
func (v *VNode) SetSearchQuery(query string) *VNode         { v.searchQuery = query; return v }
func (v *VNode) SetSearchFn(fn func(string, string) bool) *VNode {
	v.searchFn = fn
	return v
}
func (v *VNode) SetShowSearchStats(show bool) *VNode { v.showSearchStats = show; return v }
func (v *VNode) SetSearchStatsStyle(s style.Style) *VNode {
	v.searchStatsStyle = s
	return v
}
func (v *VNode) SetScrollOffset(offset int) *VNode {
	v.scrollOffset = offset
	v.scrollOffsetControlled = true
	return v
}
func (v *VNode) SetScrollOffsetControlled(offset int) *VNode {
	v.scrollOffset = offset
	v.scrollOffsetControlled = true
	return v
}
func (v *VNode) SetInitialScrollOffset(offset int) *VNode {
	v.scrollOffset = offset
	v.scrollOffsetControlled = false
	return v
}
func (v *VNode) SetSelectedIndex(index int) *VNode {
	v.selectedIndex = index
	v.selectedIndexControlled = true
	return v
}
func (v *VNode) SetSelectedIndexControlled(index int) *VNode {
	v.selectedIndex = index
	v.selectedIndexControlled = true
	return v
}
func (v *VNode) SetInitialSelectedIndex(index int) *VNode {
	v.selectedIndex = index
	v.selectedIndexControlled = false
	return v
}
func (v *VNode) SetCheckedIndices(indices []int) *VNode {
	v.checkedIndices = append([]int(nil), indices...)
	v.checkedIndicesControlled = true
	return v
}
func (v *VNode) SetInitialCheckedIndices(indices []int) *VNode {
	v.checkedIndices = append([]int(nil), indices...)
	v.checkedIndicesControlled = false
	return v
}
func (v *VNode) SetViewportHeight(height int) *VNode { v.viewportHeight = height; return v }
func (v *VNode) SetFormID(formID string) *VNode      { v.formID = formID; return v }
func (v *VNode) SetAllowScroll(allow bool) *VNode    { v.allowScroll = allow; return v }

// =============================================================================
// Convenience Methods
// =============================================================================

// AddRow adds a single row to the list
func (v *VNode) AddRow(row string) *VNode {
	v.rows = append(v.rows, row)
	return v
}

// AddRows adds multiple rows to the list
func (v *VNode) AddRows(rows ...string) *VNode {
	v.rows = append(v.rows, rows...)
	return v
}

// ============================================================================
// Getter Methods for Testing/Demo
// ============================================================================

func (v *VNode) Header() string         { return v.header }
func (v *VNode) Rows() []string         { return v.rows }
func (v *VNode) RowCount() int          { return len(v.rows) }
func (v *VNode) GetComponentID() string { return v.componentID }
func (v *VNode) GetSelectedIndex() int  { return v.selectedIndex }

// Measure creates a temporary instance and measures it with the given constraints.
func (v *VNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
	inst := v.CreateInstance().(*Instance)

	// Convert runtime.BoxConstraints to layout.Constraints
	layoutConstraints := layout.Constraints{
		MinWidth:  constraints.MinWidth,
		MaxWidth:  constraints.MaxWidth,
		MinHeight: constraints.MinHeight,
		MaxHeight: constraints.MaxHeight,
	}

	layoutSize := inst.Measure(layoutConstraints)

	// Convert layout.Size to runtime.Size (they have the same fields)
	return runtime.Size{
		Width:  layoutSize.Width,
		Height: layoutSize.Height,
	}
}

// Paint creates a temporary instance and returns its paint commands.
func (v *VNode) Paint(x, y int) []paint.DrawCmd {
	inst := v.CreateInstance().(*Instance)
	return inst.Paint(x, y)
}
