package table

import (
	"strings"
	"unicode/utf8"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for table components.
// It persists across renders and holds all state.
type Instance struct {
	// === Identification ===
	key string

	// === Props (from VNode, may change each render) ===
	columns     []TableColumn
	rows        [][]string
	headerStyle style.Style
	tableStyle  style.Style
	gap         int

	// === Runtime State ===
	bounds [4]int // x, y, w, h
	dirty  bool
}

// Ensure Instance implements required interfaces
var (
	_ rtui.ComponentInstance     = (*Instance)(nil)
	_ rtui.PaintableInstance     = (*Instance)(nil)
	_ rtui.ActionHandlerInstance = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// NewInstance creates a new TableInstance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:         getStringProp(props, "key", ""),
		columns:     getColumnsProp(props, []TableColumn{}),
		rows:        getRowsProp(props, [][]string{}),
		headerStyle: getStyleProp(props, "headerStyle"),
		tableStyle:  getStyleProp(props, "tableStyle"),
		gap:         getIntProp(props, "gap", 0),
		dirty:       true,
	}
	return inst
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

func (inst *Instance) Key() string           { return inst.key }
func (inst *Instance) SetKey(key string)     { inst.key = key }
func (inst *Instance) Init(props rtui.Props) { inst.SetProps(props) }
func (inst *Instance) Destroy()             { inst.columns = nil; inst.rows = nil }
func (inst *Instance) OnMount()             { inst.dirty = true }
func (inst *Instance) OnUnmount()           {}

func (inst *Instance) SetProps(props rtui.Props) bool {
	oldColumns := inst.columns

	inst.columns = getColumnsProp(props, inst.columns)
	inst.rows = getRowsProp(props, inst.rows)
	inst.headerStyle = getStyleProp(props, "headerStyle")
	inst.tableStyle = getStyleProp(props, "tableStyle")
	inst.gap = getIntProp(props, "gap", inst.gap)

	// Check if data changed
	changed := !columnsEqual(oldColumns, inst.columns)
	if changed {
		inst.dirty = true
	}
	return changed
}

func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":      inst.key,
		"columns":  inst.columns,
		"rows":     inst.rows,
		"gap":      inst.gap,
	}
}

func (inst *Instance) MarkDirty()    { inst.dirty = true }
func (inst *Instance) IsDirty() bool { return inst.dirty }
func (inst *Instance) GetContext() *rtui.ComponentContext { return nil }
func (inst *Instance) ClearDirty()   { inst.dirty = false }

// =============================================================================
// Measurable Interface
// =============================================================================

// Measure implements layout measurement.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if len(inst.columns) == 0 {
		return layout.Size{Width: 0, Height: 0}
	}

	// Calculate width based on columns
	totalWidth := 0
	for _, col := range inst.columns {
		if col.Width > 0 {
			totalWidth += col.Width
		} else {
			totalWidth += utf8.RuneCountInString(col.Title) + 2 // +2 for padding
		}
	}

	// Add separator width
	if len(inst.columns) > 0 {
		totalWidth += len(inst.columns) - 1 // +1 for each " | " separator
	}

	// Height: header (1) + separator (1) + rows (N) + gap
	height := 2 + len(inst.rows)
	if inst.gap > 0 {
		height += inst.gap
	}

	// Apply constraints
	if totalWidth < constraints.MinWidth {
		totalWidth = constraints.MinWidth
	}
	if totalWidth > constraints.MaxWidth && constraints.MaxWidth > 0 {
		totalWidth = constraints.MaxWidth
	}
	if height < constraints.MinHeight {
		height = constraints.MinHeight
	}
	if height > constraints.MaxHeight && constraints.MaxHeight > 0 {
		height = constraints.MaxHeight
	}

	return layout.Size{Width: totalWidth, Height: height}
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint implements drawing logic for the table.
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	if len(inst.columns) == 0 {
		return nil
	}

	var cmds []paint.DrawCmd
	currentY := y

	// Build header row
	var headers []string
	for _, col := range inst.columns {
		headers = append(headers, col.Title)
	}
	headerLine := strings.Join(headers, " | ")
	cmds = append(cmds, paint.NewTextCmd(x, currentY, headerLine, inst.headerStyle))
	currentY++

	// Draw separator line
	separator := strings.Repeat("─", len(headerLine))
	cmds = append(cmds, paint.NewTextCmd(x, currentY, separator, inst.tableStyle))
	currentY++

	// Add gap if specified
	if inst.gap > 0 {
		currentY += inst.gap
	}

	// Draw data rows
	for i, row := range inst.rows {
		if i >= len(row) {
			continue
		}
		rowLine := strings.Join(row, " | ")
		cmds = append(cmds, paint.NewTextCmd(x, currentY, rowLine, inst.tableStyle))
		currentY++
	}

	return cmds
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
// ActionHandlerInstance Interface (Table is display-only)
// =============================================================================

func (inst *Instance) CanHandleAction(actionType string) bool {
	// Table is display-only, but can handle navigation concepts
	return actionType == "navigate_up" ||
		actionType == "navigate_down" ||
		actionType == "navigate_home" ||
		actionType == "navigate_end"
}

func (inst *Instance) HandleAction(actionType string, payload interface{}) bool {
	// Table is primarily display-only
	// Actions are handled for potential future row selection
	return true
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

func getIntProp(props rtui.Props, key string, def int) int {
	if v, ok := props[key]; ok {
		if i, ok := v.(int); ok {
			return i
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

func getColumnsProp(props rtui.Props, def []TableColumn) []TableColumn {
	v, ok := props["columns"]
	if !ok {
		return def
	}
	if cols, ok := v.([]TableColumn); ok {
		return cols
	}
	return def
}

func getRowsProp(props rtui.Props, def [][]string) [][]string {
	v, ok := props["rows"]
	if !ok {
		return def
	}
	if rows, ok := v.([][]string); ok {
		return rows
	}
	return def
}

// columnsEqual compares two column slices
func columnsEqual(a, b []TableColumn) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Title != b[i].Title || a[i].Width != b[i].Width {
			return false
		}
	}
	return true
}
