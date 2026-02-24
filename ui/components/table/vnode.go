// Package table provides Fiber-first Table data display component.
package table

import (
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Types
// =============================================================================

// TableColumn represents a column in a table
type TableColumn struct {
	Title string
	Width int
}

// =============================================================================
// VNode - Pure Description (No State, No Closures, No Paint)
// =============================================================================

// VNode is the table component description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Table Props ===
	columns     []TableColumn
	rows        [][]string
	headerStyle style.Style
	tableStyle  style.Style

	// === Layout Props ===
	gap int // Gap between header and rows
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
		ElementVNode: rtui.NewElement("table"),
		columns:      []TableColumn{},
		rows:         [][]string{},
		headerStyle:  style.Style{}.Bold(true),
		tableStyle:   style.Style{},
		gap:          0,
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

func (v *VNode) Key() string           { return v.key }
func (v *VNode) SetKey(key string) rtui.VNode { v.key = key; return v }
func (v *VNode) Tag() string           { return "table" }
func (v *VNode) Type() rtui.VNodeType  { return rtui.VNodeElement }

func (v *VNode) Children() []rtui.VNode {
	// Table is a leaf component - no children
	return []rtui.VNode{}
}

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	// Table is display-only, has no children
	return v
}

func (v *VNode) GetLayer() rtui.Layer   { return rtui.LayerBase }
func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Style() style.Style    { return v.tableStyle }
func (v *VNode) SetStyle(s style.Style) rtui.VNode { v.tableStyle = s; return v }

func (v *VNode) TextContent() string   { return "" }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		"key":         v.key,
		"columns":     v.columns,
		"rows":        v.rows,
		"headerStyle": v.headerStyle,
		"tableStyle":  v.tableStyle,
		"gap":         v.gap,
	}
}

func (v *VNode) SetProps(p rtui.Props) rtui.VNode {
	if val, ok := p["key"].(string); ok {
		v.key = val
	}
	if val, ok := p["columns"].([]TableColumn); ok {
		v.columns = val
	}
	if val, ok := p["rows"].([][]string); ok {
		v.rows = val
	}
	if val, ok := p["headerStyle"].(style.Style); ok {
		v.headerStyle = val
	}
	if val, ok := p["tableStyle"].(style.Style); ok {
		v.tableStyle = val
	}
	if val, ok := p["gap"].(int); ok {
		v.gap = val
	}
	return v
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(rtui.Props{
		"key":         v.key,
		"columns":     v.columns,
		"rows":        v.rows,
		"headerStyle": v.headerStyle,
		"tableStyle":  v.tableStyle,
		"gap":         v.gap,
	})
}

// =============================================================================
// Fluent Setters
// =============================================================================

func (v *VNode) SetColumns(cols []TableColumn) *VNode  { v.columns = cols; return v }
func (v *VNode) SetRows(rows [][]string) *VNode         { v.rows = rows; return v }
func (v *VNode) SetHeaderStyle(s style.Style) *VNode   { v.headerStyle = s; return v }
func (v *VNode) SetGap(gap int) *VNode                  { v.gap = gap; return v }

func (v *VNode) AddRow(row ...string) *VNode {
	v.rows = append(v.rows, row)
	return v
}

// =============================================================================
// Accessors
// =============================================================================

func (v *VNode) Columns() []TableColumn  { return v.columns }
func (v *VNode) Rows() [][]string      { return v.rows }
func (v *VNode) HeaderStyle() style.Style { return v.headerStyle }
func (v *VNode) TableStyle() style.Style  { return v.tableStyle }
func (v *VNode) Gap() int              { return v.gap }
