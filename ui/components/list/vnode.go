package list

import (
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// VNode - Pure Description
// =============================================================================

// VNode is the pure description for the List component
// Contains only declarative properties - no state, no closures, no Paint
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Visual Properties ===
	header      string   // Optional column header text
	rows        []string // Pre-formatted data rows
	emptyText   string   // Text shown when rows is empty
	maxRows     int      // Maximum visible rows (0 = unlimited)
	showBorder  bool     // Show border around the list

	// === Separator ===
	showSeparator bool  // Show separator line between header and rows
	separatorChar rune // Separator character (default '─')

	// === Styles ===
	headerStyle   style.Style // Style for the header row
	rowStyle      style.Style // Default style for data rows
	rowStyleFn    func(int, string) style.Style // Dynamic style function per row
	selectedStyle style.Style // Style for the selected row
	borderStyle   style.Style // Style for the border

	// === State Properties (declarative initial state) ===
	scrollOffset   int // Initial scroll offset
	selectedIndex  int // Currently selected row index
	viewportHeight int // Visible height for scrolling

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
		ElementVNode:   rtui.NewElement("list"),
		key:            "",
		header:         "",
		rows:           []string{},
		emptyText:      "(empty)",
		maxRows:        0,
		showBorder:     true,
		showSeparator:  true,
		separatorChar:  '─',
		headerStyle:    style.Style{}.Bold(true),
		rowStyle:       style.Style{},
		selectedStyle:  style.Style{BG: style.Blue, FG: style.White},
		borderStyle:    style.Style{FG: style.White},
		scrollOffset:   0,
		selectedIndex:  -1,
		viewportHeight: 10,
		allowScroll:    true,
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

func (v *VNode) Key() string                   { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode  { v.key = key; return v }
func (v *VNode) Tag() string                   { return "list" }
func (v *VNode) Style() style.Style            { return v.rowStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode { v.rowStyle = s; return v }
func (v *VNode) Children() []rtui.VNode        { return nil }
func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }
func (v *VNode) GetLayer() rtui.Layer          { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
	props := rtui.Props{
		"key":            v.key,
		"header":         v.header,
		"rows":           v.rows,
		"emptyText":      v.emptyText,
		"maxRows":        v.maxRows,
		"showBorder":     v.showBorder,
		"showSeparator":  v.showSeparator,
		"separatorChar":  v.separatorChar,
		"headerStyle":    v.headerStyle,
		"rowStyle":       v.rowStyle,
		"selectedStyle":  v.selectedStyle,
		"borderStyle":    v.borderStyle,
		"scrollOffset":   v.scrollOffset,
		"selectedIndex":  v.selectedIndex,
		"viewportHeight": v.viewportHeight,
		"allowScroll":    v.allowScroll,
	}

	// Add rowStyleFn if it's set (functions can be stored in Props)
	if v.rowStyleFn != nil {
		props["rowStyleFn"] = v.rowStyleFn
	}

	return props
}

func (v *VNode) SetProps(p rtui.Props) rtui.VNode {
	if key, ok := p["key"].(string); ok {
		v.key = key
	}
	if header, ok := p["header"].(string); ok {
		v.header = header
	}
	if rows, ok := p["rows"].([]string); ok {
		v.rows = rows
	}
	if emptyText, ok := p["emptyText"].(string); ok {
		v.emptyText = emptyText
	}
	if maxRows, ok := p["maxRows"].(int); ok {
		v.maxRows = maxRows
	}
	if showBorder, ok := p["showBorder"].(bool); ok {
		v.showBorder = showBorder
	}
	if showSeparator, ok := p["showSeparator"].(bool); ok {
		v.showSeparator = showSeparator
	}
	if separatorChar, ok := p["separatorChar"].(rune); ok {
		v.separatorChar = separatorChar
	}
	if headerStyle, ok := p["headerStyle"].(style.Style); ok {
		v.headerStyle = headerStyle
	}
	if rowStyle, ok := p["rowStyle"].(style.Style); ok {
		v.rowStyle = rowStyle
	}
	if selectedStyle, ok := p["selectedStyle"].(style.Style); ok {
		v.selectedStyle = selectedStyle
	}
	if borderStyle, ok := p["borderStyle"].(style.Style); ok {
		v.borderStyle = borderStyle
	}
	if scrollOffset, ok := p["scrollOffset"].(int); ok {
		v.scrollOffset = scrollOffset
	}
	if selectedIndex, ok := p["selectedIndex"].(int); ok {
		v.selectedIndex = selectedIndex
	}
	if viewportHeight, ok := p["viewportHeight"].(int); ok {
		v.viewportHeight = viewportHeight
	}
	if allowScroll, ok := p["allowScroll"].(bool); ok {
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

func (v *VNode) SetHeader(header string) *VNode     { v.header = header; return v }
func (v *VNode) SetRows(rows []string) *VNode      { v.rows = rows; return v }
func (v *VNode) SetEmptyText(text string) *VNode  { v.emptyText = text; return v }
func (v *VNode) SetMaxRows(n int) *VNode          { v.maxRows = n; return v }
func (v *VNode) SetShowBorder(show bool) *VNode  { v.showBorder = show; return v }
func (v *VNode) SetShowSeparator(show bool) *VNode { v.showSeparator = show; return v }
func (v *VNode) SetSeparatorChar(ch rune) *VNode { v.separatorChar = ch; return v }
func (v *VNode) SetHeaderStyle(s style.Style) *VNode { v.headerStyle = s; return v }
func (v *VNode) SetRowStyle(s style.Style) *VNode    { v.rowStyle = s; return v }
func (v *VNode) SetRowStyleFn(fn func(int, string) style.Style) *VNode { v.rowStyleFn = fn; return v }
func (v *VNode) SetSelectedStyle(s style.Style) *VNode { v.selectedStyle = s; return v }
func (v *VNode) SetBorderStyle(s style.Style) *VNode   { v.borderStyle = s; return v }
func (v *VNode) SetScrollOffset(offset int) *VNode  { v.scrollOffset = offset; return v }
func (v *VNode) SetSelectedIndex(index int) *VNode { v.selectedIndex = index; return v }
func (v *VNode) SetViewportHeight(height int) *VNode { v.viewportHeight = height; return v }
func (v *VNode) SetAllowScroll(allow bool) *VNode  { v.allowScroll = allow; return v }

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

func (v *VNode) Header() string { return v.header }
func (v *VNode) Rows() []string { return v.rows }
func (v *VNode) RowCount() int { return len(v.rows) }

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

