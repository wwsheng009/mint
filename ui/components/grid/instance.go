package grid

import (
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Instance - Runtime Entity
// =============================================================================

// Instance is the runtime entity for Grid components.
// It persists across renders and holds all state.
type Instance struct {
	// === Identification ===
	key string

	// === Grid Definition ===
	columns []Dimension
	rows    []Dimension
	cells   []Cell

	// === Layout Props ===
	columnGap    int
	rowGap       int
	padding      [4]int
	alignContent rtui.Align

	// === Sizing Props ===
	width  int
	height int
	flex   int

	// === Style ===
	instStyle style.Style

	// === Runtime State ===
	bounds     [4]int // x, y, w, h
	colWidths  []int  // calculated column widths
	rowHeights []int  // calculated row heights
	childBounds [][4]int
	dirty      bool
}

// Ensure Instance implements required interfaces
var (
	_ rtui.ComponentInstance = (*Instance)(nil)
	_ rtui.PaintableInstance = (*Instance)(nil)
	_ interface {
		Measure(layout.Constraints) layout.Size
	} = (*Instance)(nil)
)

// =============================================================================
// Constructor
// =============================================================================

// NewInstance creates a new GridInstance from props.
func NewInstance(props rtui.Props) *Instance {
	inst := &Instance{
		key:          getStringProp(props, "key", ""),
		columnGap:    getIntProp(props, "columnGap", 0),
		rowGap:       getIntProp(props, "rowGap", 0),
		padding:      getPaddingProp(props),
		alignContent: getAlignContentProp(props),
		width:        getIntProp(props, "width", 0),
		height:       getIntProp(props, "height", 0),
		flex:         getIntProp(props, "flex", 0),
		instStyle:    getStyleProp(props),
		dirty:        true,
	}

	// Parse columns
	if v, ok := props["columns"].([]Dimension); ok {
		inst.columns = v
	} else {
		inst.columns = []Dimension{Flex{Factor: 1}}
	}

	// Parse rows
	if v, ok := props["rows"].([]Dimension); ok {
		inst.rows = v
	} else {
		inst.rows = []Dimension{Auto{}}
	}

	// Parse cells
	if v, ok := props["cells"].([]Cell); ok {
		inst.cells = v
	}

	return inst
}

// =============================================================================
// ComponentInstance Interface
// =============================================================================

// Key implements ComponentInstance.
func (inst *Instance) Key() string {
	return inst.key
}

// SetKey implements ComponentInstance.
func (inst *Instance) SetKey(key string) {
	inst.key = key
}

// Init implements ComponentInstance.
func (inst *Instance) Init(props rtui.Props) {
	inst.SetProps(props)
}

// Destroy implements ComponentInstance.
func (inst *Instance) Destroy() {
	inst.cells = nil
	inst.colWidths = nil
	inst.rowHeights = nil
	inst.childBounds = nil
}

// OnMount implements ComponentInstance.
func (inst *Instance) OnMount() {}

// OnUnmount implements ComponentInstance.
func (inst *Instance) OnUnmount() {}

// SetProps implements ComponentInstance.
func (inst *Instance) SetProps(props rtui.Props) bool {
	oldColumns := inst.columns
	oldRows := inst.rows
	oldCells := inst.cells

	inst.key = getStringProp(props, "key", inst.key)
	inst.columnGap = getIntProp(props, "columnGap", inst.columnGap)
	inst.rowGap = getIntProp(props, "rowGap", inst.rowGap)
	inst.padding = getPaddingPropWithDefault(props, inst.padding)
	inst.alignContent = getAlignContentProp(props)
	inst.width = getIntProp(props, "width", inst.width)
	inst.height = getIntProp(props, "height", inst.height)
	inst.flex = getIntProp(props, "flex", inst.flex)
	inst.instStyle = getStyleProp(props)

	if v, ok := props["columns"].([]Dimension); ok {
		inst.columns = v
	}
	if v, ok := props["rows"].([]Dimension); ok {
		inst.rows = v
	}
	if v, ok := props["cells"].([]Cell); ok {
		inst.cells = v
	}

	changed := len(oldColumns) != len(inst.columns) ||
		len(oldRows) != len(inst.rows) ||
		len(oldCells) != len(inst.cells)

	if changed {
		inst.dirty = true
		inst.colWidths = nil
		inst.rowHeights = nil
		inst.childBounds = nil
	}
	return changed
}

// GetProps implements ComponentInstance.
func (inst *Instance) GetProps() rtui.Props {
	return rtui.Props{
		"key":          inst.key,
		"columns":      inst.columns,
		"rows":         inst.rows,
		"cells":        inst.cells,
		"columnGap":    inst.columnGap,
		"rowGap":       inst.rowGap,
		"padding":      inst.padding,
		"alignContent": inst.alignContent,
		"width":        inst.width,
		"height":       inst.height,
		"flex":         inst.flex,
	}
}

// MarkDirty implements ComponentInstance.
func (inst *Instance) MarkDirty() {
	inst.dirty = true
}

// IsDirty implements ComponentInstance.
func (inst *Instance) IsDirty() bool {
	return inst.dirty
}

// GetContext implements ComponentInstance (no hooks for Grid).
func (inst *Instance) GetContext() *rtui.ComponentContext {
	return nil
}

// =============================================================================
// Bounds Management
// =============================================================================

// GetBounds returns the layout bounds.
func (inst *Instance) GetBounds() (x, y, w, h int) {
	return inst.bounds[0], inst.bounds[1], inst.bounds[2], inst.bounds[3]
}

// SetBounds sets the layout bounds.
func (inst *Instance) SetBounds(x, y, w, h int) {
	inst.bounds = [4]int{x, y, w, h}
}

// SetChildBounds sets bounds for a specific child.
func (inst *Instance) SetChildBounds(index int, x, y, w, h int) {
	if index < 0 || index >= len(inst.cells) {
		return
	}
	if inst.childBounds == nil {
		inst.childBounds = make([][4]int, len(inst.cells))
	}
	inst.childBounds[index] = [4]int{x, y, w, h}
}

// GetChildBounds returns bounds for a specific child.
func (inst *Instance) GetChildBounds(index int) (x, y, w, h int) {
	if index < 0 || index >= len(inst.childBounds) {
		return 0, 0, 0, 0
	}
	b := inst.childBounds[index]
	return b[0], b[1], b[2], b[3]
}

// GetColumnWidths returns calculated column widths.
func (inst *Instance) GetColumnWidths() []int {
	return inst.colWidths
}

// GetRowHeights returns calculated row heights.
func (inst *Instance) GetRowHeights() []int {
	return inst.rowHeights
}

// =============================================================================
// Measurable Interface (Two-Pass Layout)
// =============================================================================

// Measure implements layout.Measurable interface.
// Calculates the grid's ideal size given the constraints.
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if inst == nil {
		return layout.Size{}
	}

	numCols := len(inst.columns)
	if numCols == 0 {
		numCols = 1
	}

	// Calculate actual number of rows needed based on cells
	actualNumRows := inst.calculateRowCount(numCols)

	// Calculate column widths and row heights
	inst.colWidths = inst.calculateColumnWidths(constraints.MaxWidth - inst.padding[1] - inst.padding[3])
	inst.rowHeights = inst.calculateRowHeights(constraints.MaxHeight - inst.padding[0] - inst.padding[2], numCols, actualNumRows)

	// Calculate total size
	totalW := inst.padding[1] + inst.padding[3]
	totalH := inst.padding[0] + inst.padding[2]

	for _, w := range inst.colWidths {
		totalW += w
	}
	for _, h := range inst.rowHeights {
		totalH += h
	}

	// Add gaps
	if numCols > 1 {
		totalW += inst.columnGap * (numCols - 1)
	}
	if actualNumRows > 1 {
		totalH += inst.rowGap * (actualNumRows - 1)
	}

	// Apply explicit dimensions
	if inst.width > 0 {
		totalW = inst.width
	}
	if inst.height > 0 {
		totalH = inst.height
	}

	// Apply constraints
	totalW = constraints.ConstrainWidth(totalW)
	totalH = constraints.ConstrainHeight(totalH)

	return layout.Size{Width: totalW, Height: totalH}
}

// calculateRowCount calculates the number of rows needed for the grid.
func (inst *Instance) calculateRowCount(numCols int) int {
	// If there are explicit cells, find max row
	if len(inst.cells) > 0 {
		maxRow := 0
		for _, cell := range inst.cells {
			endRow := cell.Row + cell.RowSpan
			if endRow > maxRow {
				maxRow = endRow
			}
		}
		return maxRow
	}

	// Use defined rows count if no cells
	if len(inst.rows) > 0 {
		return len(inst.rows)
	}

	return 1
}

// calculateColumnWidths calculates the actual width for each column.
func (inst *Instance) calculateColumnWidths(availableWidth int) []int {
	numCols := len(inst.columns)
	if numCols == 0 {
		return []int{availableWidth}
	}

	widths := make([]int, numCols)
	fixedWidth := 0
	flexCount := 0
	flexTotalFactor := 0

	// First pass: calculate fixed widths
	for i, col := range inst.columns {
		switch c := col.(type) {
		case Fixed:
			widths[i] = int(c)
			fixedWidth += widths[i]
		case Flex:
			flexCount++
			if c.Factor > 0 {
				flexTotalFactor += c.Factor
			} else {
				flexTotalFactor += 1
			}
		case Auto:
			// Auto columns get minimum width
			widths[i] = 10
			fixedWidth += widths[i]
		case Min:
			widths[i] = c.Min
			fixedWidth += widths[i]
		case Max:
			widths[i] = c.Max
			fixedWidth += widths[i]
		}
	}

	// Subtract column gaps from available width
	gapWidth := inst.columnGap * (numCols - 1)
	if gapWidth < 0 {
		gapWidth = 0
	}
	remainingWidth := availableWidth - fixedWidth - gapWidth
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	// Second pass: distribute remaining width to flex columns
	if flexCount > 0 && flexTotalFactor > 0 {
		for i, col := range inst.columns {
			if _, ok := col.(Flex); ok {
				factor := 1
				if f, ok := col.(Flex); ok && f.Factor > 0 {
					factor = f.Factor
				}
				widths[i] = (remainingWidth * factor) / flexTotalFactor
				if widths[i] < 1 {
					widths[i] = 1
				}
			}
		}
	}

	return widths
}

// calculateRowHeights calculates the actual height for each row.
// numCols is the number of columns in the grid.
// actualNumRows is the actual number of rows needed based on cells.
func (inst *Instance) calculateRowHeights(availableHeight int, numCols, actualNumRows int) []int {
	// Use actualNumRows if it's greater than defined rows
	numRows := actualNumRows
	if len(inst.rows) > numRows {
		numRows = len(inst.rows)
	}
	if numRows == 0 {
		return []int{availableHeight}
	}

	// Expand rows definition if needed
	rows := inst.rows
	if len(rows) < numRows {
		expanded := make([]Dimension, numRows)
		copy(expanded, rows)
		for i := len(rows); i < numRows; i++ {
			expanded[i] = Auto{} // Default to Auto for undefined rows
		}
		rows = expanded
	}

	heights := make([]int, numRows)
	fixedHeight := 0
	flexCount := 0
	flexTotalFactor := 0

	// First pass: calculate fixed heights
	for i := 0; i < numRows; i++ {
		switch r := rows[i].(type) {
		case Fixed:
			heights[i] = int(r)
			fixedHeight += heights[i]
		case Flex:
			flexCount++
			if r.Factor > 0 {
				flexTotalFactor += r.Factor
			} else {
				flexTotalFactor += 1
			}
		case Auto:
			// Auto rows get minimum height (1 line per row)
			heights[i] = 1
			fixedHeight += heights[i]
		case Min:
			heights[i] = r.Min
			fixedHeight += heights[i]
		case Max:
			heights[i] = r.Max
			fixedHeight += heights[i]
		}
	}

	// Subtract row gaps from available height
	gapHeight := inst.rowGap * (numRows - 1)
	if gapHeight < 0 {
		gapHeight = 0
	}
	remainingHeight := availableHeight - fixedHeight - gapHeight
	if remainingHeight < 0 {
		remainingHeight = 0
	}

	// Second pass: distribute remaining height to flex rows
	if flexCount > 0 && flexTotalFactor > 0 {
		for i := 0; i < numRows; i++ {
			if _, ok := rows[i].(Flex); ok {
				factor := 1
				if f, ok := rows[i].(Flex); ok && f.Factor > 0 {
					factor = f.Factor
				}
				heights[i] = (remainingHeight * factor) / flexTotalFactor
				if heights[i] < 1 {
					heights[i] = 1
				}
			}
		}
	}

	return heights
}

// =============================================================================
// PaintableInstance Interface
// =============================================================================

// Paint implements PaintableInstance.
// Grid 是纯布局容器，布局由 LayoutBox 处理，子元素渲染由 PaintEngine 处理。
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	// 纯布局容器没有自身需要绘制的内容
	return nil
}

// =============================================================================
// Style Management
// =============================================================================

// GetStyle returns the instance style.
func (inst *Instance) GetStyle() style.Style {
	return inst.instStyle
}

// SetStyle sets the instance style.
func (inst *Instance) SetStyle(s style.Style) {
	inst.instStyle = s
}

// ClearDirty clears the dirty flag.
func (inst *Instance) ClearDirty() {
	inst.dirty = false
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

func getStyleProp(props rtui.Props) style.Style {
	if v, ok := props["style"]; ok {
		if s, ok := v.(style.Style); ok {
			return s
		}
	}
	return style.Style{}
}

func getPaddingProp(props rtui.Props) [4]int {
	if v, ok := props["padding"]; ok {
		if p, ok := v.([4]int); ok {
			return p
		}
	}
	return [4]int{0, 0, 0, 0}
}

func getPaddingPropWithDefault(props rtui.Props, def [4]int) [4]int {
	if v, ok := props["padding"]; ok {
		if p, ok := v.([4]int); ok {
			return p
		}
	}
	return def
}

func getAlignContentProp(props rtui.Props) rtui.Align {
	if v, ok := props["alignContent"]; ok {
		if a, ok := v.(rtui.Align); ok {
			return a
		}
	}
	return rtui.AlignStart
}
