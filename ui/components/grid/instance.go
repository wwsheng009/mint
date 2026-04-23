package grid

import (
	"fmt"
	"github.com/wwsheng009/mint/ui/components/internal/proputil"

	"github.com/wwsheng009/mint/internal/log"
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
	key     string
	fiberId int // Fiber node ID for tracing

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

	// ✨ Border Props (方案 A - 边框作为容器属性) ===
	borderStyle string // "none", "single", "double", "rounded", "dashed"
	borderLabel string // Optional label displayed on top border

	// === ✨ Cell Borders Props (格子间边框) ===
	showCellBorders   bool   // 是否显示格子边框
	cellBorderStyle   string // 边框样式: "none", "single", "double", "light"
	cellBorderRounded bool   // cell 边框是否带圆角
	cellBorderColor   string // 边框颜色

	// === Style ===
	instStyle style.Style

	// === Runtime State ===
	bounds      [4]int // x, y, w, h
	colWidths   []int  // calculated column widths
	rowHeights  []int  // calculated row heights
	childBounds [][4]int
	dirty       bool
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
		key:          proputil.GetString(props, "key", ""),
		columnGap:    proputil.GetInt(props, "columnGap", 0),
		rowGap:       proputil.GetInt(props, "rowGap", 0),
		padding:      getPaddingProp(props),
		alignContent: getAlignContentProp(props),
		width:        proputil.GetInt(props, "width", 0),
		height:       proputil.GetInt(props, "height", 0),
		flex:         proputil.GetInt(props, "flex", 0),
		instStyle:    proputil.GetStyle(props, "style", style.Style{}),
		dirty:        true,
		borderStyle:  proputil.GetString(props, "borderStyle", "none"), // ✨ 边框样式
		borderLabel:  proputil.GetString(props, "label", ""),           // ✨ 边框标签
		// ✨ Cell Borders 初始化
		showCellBorders:   proputil.GetBool(props, "showCellBorders", false),
		cellBorderStyle:   proputil.GetString(props, "cellBorderStyle", "single"),
		cellBorderRounded: proputil.GetBool(props, "cellBorderRounded", false),
		cellBorderColor:   proputil.GetString(props, "cellBorderColor", ""),
	}

	// Parse columns
	if v, ok := props[propColumns].([]Dimension); ok {
		inst.columns = v
	} else {
		inst.columns = []Dimension{Flex{Factor: 1}}
	}

	// Parse rows
	if v, ok := props[propRows].([]Dimension); ok {
		inst.rows = v
	} else {
		inst.rows = []Dimension{Auto{}}
	}

	// Parse cells
	if v, ok := props[propCells].([]Cell); ok {
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

	inst.key = proputil.GetString(props, "key", inst.key)
	inst.columnGap = proputil.GetInt(props, "columnGap", inst.columnGap)
	inst.rowGap = proputil.GetInt(props, "rowGap", inst.rowGap)
	inst.padding = getPaddingPropWithDefault(props, inst.padding)
	inst.alignContent = getAlignContentProp(props)
	inst.width = proputil.GetInt(props, "width", inst.width)
	inst.height = proputil.GetInt(props, "height", inst.height)
	inst.flex = proputil.GetInt(props, "flex", inst.flex)
	inst.instStyle = proputil.GetStyle(props, "style", style.Style{})

	// ✨ 容器边框属性
	if v, ok := props[propBorderStyle].(string); ok {
		inst.borderStyle = v
	}
	if v, ok := props[propLabel].(string); ok {
		inst.borderLabel = v
	}

	// ✨ Cell Borders 属性
	if v, ok := props[propShowCellBorders].(bool); ok {
		inst.showCellBorders = v
	}
	if v, ok := props[propCellBorderStyle].(string); ok {
		inst.cellBorderStyle = v
	}
	if v, ok := props[propCellBorderRounded].(bool); ok {
		inst.cellBorderRounded = v
	}
	if v, ok := props[propCellBorderColor].(string); ok {
		inst.cellBorderColor = v
	}

	if v, ok := props[propColumns].([]Dimension); ok {
		inst.columns = v
	}
	if v, ok := props[propRows].([]Dimension); ok {
		inst.rows = v
	}
	if v, ok := props[propCells].([]Cell); ok {
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
		propKey:          inst.key,
		propColumns:      inst.columns,
		propRows:         inst.rows,
		propCells:        inst.cells,
		propColumnGap:    inst.columnGap,
		propRowGap:       inst.rowGap,
		propPadding:      inst.padding,
		propAlignContent: inst.alignContent,
		propWidth:        inst.width,
		propHeight:       inst.height,
		propFlex:         inst.flex,
		// ✨ 容器边框属性
		propBorderStyle: inst.borderStyle,
		propLabel:       inst.borderLabel,
		// ✨ Cell Borders 属性
		propShowCellBorders:   inst.showCellBorders,
		propCellBorderStyle:   inst.cellBorderStyle,
		propCellBorderRounded: inst.cellBorderRounded,
		propCellBorderColor:   inst.cellBorderColor,
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

	// ✨ 当 Layout Engine 设置 bounds 后，根据实际分配的 box 高度重新计算 rowHeights
	// 这样可以确保边框绘制在正确的位置，不会超出 box 范围
	if len(inst.columns) == 0 {
		return
	}

	numCols := len(inst.columns)
	numRows := inst.calculateRowCount(numCols)

	// ✨ Cell Borders: 计算边框占用高度
	cellBorderHeight := 0
	if inst.showCellBorders {
		cellBorderHeight = numRows + 1 // 上边框 + 中间分隔 + 下边框 = numRows + 1
	}

	// ✨ 计算实际可用的内容高度
	availableH := h - inst.padding[0] - inst.padding[2] - cellBorderHeight
	if availableH < 0 {
		availableH = 0
	}

	// ✨ 重新计算 rowHeights 以适应实际分配的 box 高度
	oldRowHeights := make([]int, len(inst.rowHeights))
	copy(oldRowHeights, inst.rowHeights)
	inst.rowHeights = inst.calculateRowHeights(availableH, numCols, numRows)

	// ✨ 重新计算 colWidths 以适应实际分配的 box 宽度
	// Cell Borders: 垂直边框占用 numCols + 1 个字符
	cellBorderWidth := 0
	if inst.showCellBorders {
		cellBorderWidth = numCols + 1
	}
	availableW := w - inst.padding[3] - inst.padding[1] - cellBorderWidth
	if availableW < 0 {
		availableW = 0
	}
	oldColWidths := make([]int, len(inst.colWidths))
	copy(oldColWidths, inst.colWidths)
	inst.colWidths = inst.calculateColumnWidths(availableW)

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
// ✨ 使用 runtime/layout.Grid 进行布局计算。
func (inst *Instance) Measure(constraints layout.Constraints) layout.Size {
	if inst == nil {
		return layout.Size{}
	}

	// === 追踪集成：记录入口 ===
	gridId := fmt.Sprintf("grid-%s", inst.key)
	if layout.IsTracerEnabled() {
		layout.TraceMeasuring(
			"parent",
			gridId,
			fmt.Sprintf("root/grids/%s/measure", inst.key),
			constraints,
			layout.Constraints{},
			layout.Size{},
			"Grid.Measure entrance",
		)

		// 记录 Grid 配置
		if len(inst.columns) > 0 || len(inst.rows) > 0 {
			layout.TraceMeasuring(
				gridId,
				"grid-config",
				fmt.Sprintf("root/grids/%s/config", inst.key),
				constraints,
				layout.Constraints{},
				layout.Size{},
				fmt.Sprintf(
					"Config: cols=%d, rows=%d, cells=%d, gaps=%dx%d, padding=[%d,%d,%d,%d], width=%d, height=%d, flex=%d, showCellBorders=%v",
					len(inst.columns), len(inst.rows), len(inst.cells),
					inst.columnGap, inst.rowGap,
					inst.padding[0], inst.padding[1], inst.padding[2], inst.padding[3],
					inst.width, inst.height, inst.flex,
					inst.showCellBorders,
				),
			)
		}
	}

	// ✨ 使用 runtime/layout.Grid 进行布局计算
	gridStyle := inst.GetGridStyle()
	gridLayout := layout.NewGridLayout("ui-grid", gridStyle)

	// 调用 layout.Grid 的 Measure 方法
	size := gridLayout.Measure(constraints)

	// ✨ 从 layout.Grid 获取计算结果
	inst.colWidths = gridLayout.GetColumnWidths()
	inst.rowHeights = gridLayout.GetRowHeights()

	// === 追踪集成：记录列宽和行高 ===
	if layout.IsTracerEnabled() && len(inst.colWidths) > 0 {
		// 记录每列的计算结果
		for colIdx, colWidth := range inst.colWidths {
			colDescription := "Unknown"
			if colIdx < len(inst.columns) {
				colDescription = formatDimensionDescription(inst.columns[colIdx])
			}
			layout.TraceMeasuring(
				gridId,
				fmt.Sprintf("col-%d", colIdx),
				fmt.Sprintf("root/grids/%s/col-%d", inst.key, colIdx),
				layout.Constraints{MinWidth: colWidth, MaxWidth: colWidth},
				layout.Constraints{MinWidth: colWidth, MaxWidth: colWidth},
				layout.Size{Width: colWidth, Height: 0},
				fmt.Sprintf("Column %d: %s, width=%d", colIdx, colDescription, colWidth),
			)
		}

		// 记录每行的计算结果
		for rowIdx, rowHeight := range inst.rowHeights {
			rowDesc := "Auto"
			if rowIdx < len(inst.rows) {
				rowDesc = formatDimensionDescription(inst.rows[rowIdx])
			}
			layout.TraceMeasuring(
				gridId,
				fmt.Sprintf("row-%d", rowIdx),
				fmt.Sprintf("root/grids/%s/row-%d", inst.key, rowIdx),
				layout.Constraints{MinHeight: rowHeight, MaxHeight: rowHeight},
				layout.Constraints{MinHeight: rowHeight, MaxHeight: rowHeight},
				layout.Size{Width: 0, Height: rowHeight},
				fmt.Sprintf("Row %d: %s, height=%d", rowIdx, rowDesc, rowHeight),
			)
		}

		// 记录汇总信息
		contentW := sumWithGaps(inst.colWidths, 0)
		contentH := sumWithGaps(inst.rowHeights, 0)
		layout.TraceMeasuring(
			gridId,
			"grid-summary",
			fmt.Sprintf("root/grids/%s/summary", inst.key),
			layout.Constraints{MinWidth: contentW, MaxWidth: contentW, MinHeight: contentH, MaxHeight: contentH},
			layout.Constraints{MinWidth: size.Width, MaxWidth: size.Width, MinHeight: size.Height, MaxHeight: size.Height},
			size,
			fmt.Sprintf(
				"Summary: content=%dx%d, result=%dx%d, cols=%d, rows=%d, colGap=%d, rowGap=%d",
				contentW, contentH, size.Width, size.Height,
				len(inst.colWidths), len(inst.rowHeights),
				inst.columnGap, inst.rowGap,
			),
		)
	}

	// === 追踪集成：记录出口 ===
	if layout.IsTracerEnabled() {
		outputConstraints := layout.Constraints{
			MinWidth:  size.Width,
			MaxWidth:  size.Width,
			MinHeight: size.Height,
			MaxHeight: size.Height,
		}
		layout.TraceMeasuring(
			gridId,
			"parent",
			fmt.Sprintf("root/grids/%s/measure", inst.key),
			constraints,
			outputConstraints,
			size,
			fmt.Sprintf(
				"Grid.Measure complete: %dx%d (cols+gaps=%d, rows+gaps=%d, padding=[%d,%d,%d,%d], borders=%v)",
				size.Width, size.Height,
				sumWithGaps(inst.colWidths, inst.columnGap),
				sumWithGaps(inst.rowHeights, inst.rowGap),
				inst.padding[0], inst.padding[1], inst.padding[2], inst.padding[3],
				inst.showCellBorders,
			),
		)
	}
	return size
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
	autoCount := 0 // ✨ 统计 Auto 行的数量

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
			// ✨ Auto rows: use the height from previous Measure calculation
			// If inst.rowHeights is available (from Measure phase), use it as reference
			// Otherwise, use minimum height 1
			if i < len(inst.rowHeights) && inst.rowHeights[i] > 0 {
				heights[i] = inst.rowHeights[i]
			} else {
				heights[i] = 1
			}
			fixedHeight += heights[i]
			autoCount++
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
	// ✨ Note: Auto rows keep their minimum height (1 line), they don't expand
	//      This prevents Auto rows from taking up too much space
	// ✨ FIX: Auto rows should also expand proportionally to fill available space
	//      When availableHeight is larger than the original measured heights,
	//      Auto rows should be scaled up proportionally
	if flexCount > 0 && flexTotalFactor > 0 {
		// Distribute to flex rows
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

	// ✨ Third pass: scale Auto rows proportionally if there's remaining space
	// This allows Auto rows to use available space beyond their measured minimum
	// ✨ DEBUG
	log.RenderLogger.Debug("[DEBUG SETBOUNDS] Before Third pass: autoCount=%d, remainingHeight=%d\n",
		autoCount, remainingHeight)

	// ✨ FIX: Auto rows may also need to be scaled down when content > available space
	// or scaled up when content < available space
	if autoCount > 0 {
		// Calculate current total auto height
		currentAutoHeight := 0
		autoIndices := []int{}
		for i := 0; i < numRows; i++ {
			if _, ok := rows[i].(Auto); ok && i < len(inst.rowHeights) && inst.rowHeights[i] > 0 {
				currentAutoHeight += heights[i]
				autoIndices = append(autoIndices, i)
			}
		}

		// Distribute remaining height proportionally to Auto rows
		if currentAutoHeight > 0 && len(autoIndices) > 0 {
			distributed := 0
			for idx, i := range autoIndices {
				// Proportional distribution based on original height
				proportion := float64(heights[i]) / float64(currentAutoHeight)
				extraHeight := int(proportion * float64(remainingHeight))
				heights[i] += extraHeight
				distributed += extraHeight
				// Add remainder to first Auto row
				if idx == len(autoIndices)-1 {
					remainder := remainingHeight - distributed
					heights[i] += remainder
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
// 绘制格子边框（如果启用）
func (inst *Instance) Paint(x, y int) []paint.DrawCmd {
	// ✨ 绘制格子边框
	cmds := inst.GenCellBorderDrawCmds(x, y)
	return cmds
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

func getPaddingProp(props rtui.Props) [4]int {
	if v, ok := props[propPadding]; ok {
		if p, ok := v.([4]int); ok {
			return p
		}
	}
	return [4]int{0, 0, 0, 0}
}

func getPaddingPropWithDefault(props rtui.Props, def [4]int) [4]int {
	if v, ok := props[propPadding]; ok {
		if p, ok := v.([4]int); ok {
			return p
		}
	}
	return def
}

func getAlignContentProp(props rtui.Props) rtui.Align {
	if v, ok := props[propAlignContent]; ok {
		if a, ok := v.(rtui.Align); ok {
			return a
		}
	}
	return rtui.AlignStart
}

// ✨ Cell Borders 辅助函数

// =============================================================================
// ✨ GridStyleProvider Interface Implementation
// =============================================================================

// GetGridStyle implements layout.GridStyleProvider interface
// 返回 grid 的样式信息，包括 cell borders 配置
func (inst *Instance) GetGridStyle() *layout.GridStyle {
	// 转换 columns
	gridCols := make([]layout.GridDimension, len(inst.columns))
	for i, col := range inst.columns {
		switch c := col.(type) {
		case Fixed:
			gridCols[i] = layout.GridFixed(c)
		case Flex:
			gridCols[i] = layout.GridFlex{Factor: c.Factor}
		case Auto:
			gridCols[i] = layout.GridAuto{}
		case Min:
			gridCols[i] = layout.GridMin{Min: c.Min, Content: layout.GridAuto{}}
		case Max:
			gridCols[i] = layout.GridMax{Max: c.Max, Content: layout.GridAuto{}}
		}
	}

	// 转换 rows
	gridRows := make([]layout.GridDimension, len(inst.rows))
	for i, row := range inst.rows {
		switch r := row.(type) {
		case Fixed:
			gridRows[i] = layout.GridFixed(r)
		case Flex:
			gridRows[i] = layout.GridFlex{Factor: r.Factor}
		case Auto:
			gridRows[i] = layout.GridAuto{}
		case Min:
			gridRows[i] = layout.GridMin{Min: r.Min, Content: layout.GridAuto{}}
		case Max:
			gridRows[i] = layout.GridMax{Max: r.Max, Content: layout.GridAuto{}}
		}
	}

	// ✨ 不再返回 Cells，因为：
	// 1. Cells 的 Child 需要填入 layout.Node，但 GetGridStyle() 无法访问 layout.Node
	// 2. layout.Node 由 layout.Node.Children() 通过 grid.SetChildren() 设置
	// 3. LayoutChildren 会自动使用 g.children 进行布局（row-major 顺序）
	// 4. 这样可以确保子节点的位置通过 gridLayout.LayoutChildren() 正确计算

	return &layout.GridStyle{
		Columns:          gridCols,
		Rows:             gridRows,
		Cells:            nil, // 不返回 Cells，使用自动布局
		ColumnGap:        inst.columnGap,
		RowGap:           inst.rowGap,
		Padding:          layout.Padding{Top: inst.padding[0], Right: inst.padding[1], Bottom: inst.padding[2], Left: inst.padding[3]},
		Width:            inst.width,
		Height:           inst.height,
		ShowCellBorders:  inst.showCellBorders,
		CellBorderWidth:  1,
		CellBorderHeight: 1,
	}
}

// =============================================================================
// Helper Functions for Tracing
// =============================================================================

// formatDimensionDescription 格式化维度描述（用于追踪）
func formatDimensionDescription(d Dimension) string {
	switch v := d.(type) {
	case Fixed:
		return fmt.Sprintf("Fixed(%d)", int(v))
	case Flex:
		return fmt.Sprintf("Flex(factor=%d)", v.Factor)
	case Auto:
		return "Auto"
	case Min:
		return fmt.Sprintf("Min(%d, %s)", v.Min, formatDimensionDescription(v.Content))
	case Max:
		return fmt.Sprintf("Max(%d, %s)", v.Max, formatDimensionDescription(v.Content))
	default:
		return "Unknown"
	}
}

// sumWithGaps 计算数组之和加上间隙总和
func sumWithGaps(arr []int, gap int) int {
	if len(arr) == 0 {
		return 0
	}
	total := 0
	for _, v := range arr {
		total += v
	}
	if len(arr) > 1 && gap > 0 {
		total += gap * (len(arr) - 1)
	}
	return total
}
