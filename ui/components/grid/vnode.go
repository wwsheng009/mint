// Package grid provides Fiber-first Grid layout component.
package grid

import (
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Grid Dimension Types (VNode layer - declarative only)
// =============================================================================

// Dimension represents a column or row dimension in the grid.
type Dimension interface {
	isGridDimension()
}

// Fixed creates a fixed-size dimension.
type Fixed int

func (f Fixed) isGridDimension() {}

// Flex creates a flexible dimension that takes remaining space.
type Flex struct {
	Factor int // Flex factor, defaults to 1
}

func (f Flex) isGridDimension() {}

// Auto creates a dimension that sizes to content.
type Auto struct{}

func (a Auto) isGridDimension() {}

// Min creates a dimension with minimum size.
type Min struct {
	Min     int
	Content Dimension
}

func (m Min) isGridDimension() {}

// Max creates a dimension with maximum size.
type Max struct {
	Max     int
	Content Dimension
}

func (m Max) isGridDimension() {}

// =============================================================================
// Grid Cell (VNode layer - uses rtui.VNode)
// =============================================================================

// Cell represents a cell in the grid.
type Cell struct {
	Child   rtui.VNode
	Row     int // 0-based
	Col     int // 0-based
	RowSpan int // default 1
	ColSpan int // default 1
}

// =============================================================================
// VNode - Description Only (No State, No Closures, No Paint)
// =============================================================================

// VNode is the grid layout component description.
// It contains ONLY declarative information - no state, no closures, no paint logic.
type VNode struct {
	*rtui.ElementVNode

	// === Identification ===
	key string

	// === Grid Definition ===
	columns []Dimension // column definitions
	rows    []Dimension // row definitions
	cells   []Cell      // grid cells

	// === Layout Props ===
	columnGap    int        // gap between columns
	rowGap       int        // gap between rows
	padding      [4]int     // top, right, bottom, left
	alignContent rtui.Align // alignment of the whole grid in container

	// === Sizing Props ===
	width  int // explicit width (0 = auto)
	height int // explicit height (0 = auto)
	flex   int // flex factor

	// === Border Props (方案 A - 边框作为容器属性) ===
	borderStyle  string // "none", "single", "double", "rounded", "dashed"
	borderLabel  string // Optional label displayed on top border

	// === ✨ Cell Borders Props (格子间边框) ===
	showCellBorders   bool   // 是否显示格子边框
	cellBorderStyle   string // 边框样式: "none", "single", "double", "light"
	cellBorderRounded bool   // cell 边框是否带圆角
	cellBorderColor   string // 边框颜色

	// === Style ===
	style style.Style
}

// Ensure VNode implements required interfaces
var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// =============================================================================
// Constructors
// =============================================================================

// New creates a new Grid VNode.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("grid"),
		columns:      []Dimension{Flex{Factor: 1}},
		rows:         []Dimension{Auto{}},
		cells:        make([]Cell, 0),
		columnGap:    0,
		rowGap:       0,
		padding:      [4]int{0, 0, 0, 0},
		alignContent: rtui.AlignStart,
		borderStyle:  "none", // 默认无边框
		borderLabel:  "",
		// ✨ Cell Borders 初始化
		showCellBorders:   false,
		cellBorderStyle:   "single",
		cellBorderRounded: false,
		cellBorderColor:   "",
	}
}

// =============================================================================
// rtui.VNode Interface Implementation
// =============================================================================

// Key returns the component key.
func (g *VNode) Key() string {
	return g.key
}

// SetKey sets the component key - returns VNode for chaining.
func (g *VNode) SetKey(key string) rtui.VNode {
	g.key = key
	return g
}

// Tag returns the tag name.
func (g *VNode) Tag() string {
	return "grid"
}

// Style returns the visual style.
func (g *VNode) Style() style.Style {
	return g.style
}

// SetStyle sets the visual style - returns VNode for chaining.
func (g *VNode) SetStyle(st style.Style) rtui.VNode {
	g.style = st
	return g
}

// Children returns child nodes (flattened from cells).
func (g *VNode) Children() []rtui.VNode {
	children := make([]rtui.VNode, len(g.cells))
	for i, cell := range g.cells {
		children[i] = cell.Child
	}
	return children
}

// SetChildren sets child nodes - returns VNode for chaining.
func (g *VNode) SetChildren(children []rtui.VNode) rtui.VNode {
	// Auto-position children in row-major order
	g.cells = make([]Cell, len(children))
	for i, child := range children {
		col := i % max(1, len(g.columns))
		row := i / max(1, len(g.columns))
		g.cells[i] = Cell{
			Child:   child,
			Row:     row,
			Col:     col,
			RowSpan: 1,
			ColSpan: 1,
		}
	}
	return g
}

// GetLayer returns the rendering layer.
func (g *VNode) GetLayer() rtui.Layer {
	return rtui.LayerBase
}

// SetLayer sets the rendering layer - returns VNode for chaining.
func (g *VNode) SetLayer(l rtui.Layer) rtui.VNode {
	return g
}

// Props returns the node properties.
func (g *VNode) Props() rtui.Props {
	return rtui.Props{
		"key":          g.key,
		"columns":      g.columns,
		"rows":         g.rows,
		"cells":        g.cells,
		"columnGap":    g.columnGap,
		"rowGap":       g.rowGap,
		"padding":      g.padding,
		"alignContent": g.alignContent,
		"width":        g.width,
		"height":       g.height,
		"flex":         g.flex,
		"borderStyle":  g.borderStyle,  // ✨ 边框样式
		"label":        g.borderLabel,  // ✨ 边框标签
		// ✨ Cell Borders 属性
		"showCellBorders":   g.showCellBorders,
		"cellBorderStyle":   g.cellBorderStyle,
		"cellBorderRounded": g.cellBorderRounded,
		"cellBorderColor":   g.cellBorderColor,
	}
}

// SetProps sets the node properties - returns VNode for chaining.
func (g *VNode) SetProps(p rtui.Props) rtui.VNode {
	if v, ok := p["key"].(string); ok {
		g.key = v
	}
	if v, ok := p["columns"].([]Dimension); ok {
		g.columns = v
	}
	if v, ok := p["rows"].([]Dimension); ok {
		g.rows = v
	}
	if v, ok := p["cells"].([]Cell); ok {
		g.cells = v
	}
	if v, ok := p["columnGap"].(int); ok {
		g.columnGap = v
	}
	if v, ok := p["rowGap"].(int); ok {
		g.rowGap = v
	}
	if v, ok := p["padding"].([4]int); ok {
		g.padding = v
	}
	if v, ok := p["alignContent"].(rtui.Align); ok {
		g.alignContent = v
	}
	if v, ok := p["width"].(int); ok {
		g.width = v
	}
	if v, ok := p["height"].(int); ok {
		g.height = v
	}
	if v, ok := p["flex"].(int); ok {
		g.flex = v
	}
	// ✨ 边框属性
	if v, ok := p["borderStyle"].(string); ok {
		g.borderStyle = v
	}
	if v, ok := p["label"].(string); ok {
		g.borderLabel = v
	}
	// ✨ Cell Borders 属性
	if v, ok := p["showCellBorders"].(bool); ok {
		g.showCellBorders = v
	}
	if v, ok := p["cellBorderStyle"].(string); ok {
		g.cellBorderStyle = v
	}
	if v, ok := p["cellBorderRounded"].(bool); ok {
		g.cellBorderRounded = v
	}
	if v, ok := p["cellBorderColor"].(string); ok {
		g.cellBorderColor = v
	}
	return g
}

// =============================================================================
// InstanceFactory Implementation
// =============================================================================

// CreateInstance creates a new GridInstance from this VNode description.
func (g *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(rtui.Props{
		"key":          g.key,
		"columns":      g.columns,
		"rows":         g.rows,
		"cells":        g.cells,
		"columnGap":    g.columnGap,
		"rowGap":       g.rowGap,
		"padding":      g.padding,
		"alignContent": g.alignContent,
		"width":        g.width,
		"height":       g.height,
		"flex":         g.flex,
		"borderStyle":  g.borderStyle,  // ✨ 边框样式
		"label":        g.borderLabel,  // ✨ 边框标签
		// ✨ Cell Borders 属性
		"showCellBorders":   g.showCellBorders,
		"cellBorderStyle":   g.cellBorderStyle,
		"cellBorderRounded": g.cellBorderRounded,
		"cellBorderColor":   g.cellBorderColor,
		"style":        g.style,
	})
}

// =============================================================================
// Builder Methods - Fluent API
// =============================================================================

// SetColumns sets the column definitions.
func (g *VNode) SetColumns(cols ...Dimension) *VNode {
	g.columns = cols
	return g
}

// SetRows sets the row definitions.
func (g *VNode) SetRows(rows ...Dimension) *VNode {
	g.rows = rows
	return g
}

// AddCell adds a cell at the specified position.
func (g *VNode) AddCell(row, col int, child rtui.VNode) *VNode {
	g.cells = append(g.cells, Cell{
		Child:   child,
		Row:     row,
		Col:     col,
		RowSpan: 1,
		ColSpan: 1,
	})
	return g
}

// AddCellSpan adds a cell that spans multiple rows and columns.
func (g *VNode) AddCellSpan(row, col, rowSpan, colSpan int, child rtui.VNode) *VNode {
	g.cells = append(g.cells, Cell{
		Child:   child,
		Row:     row,
		Col:     col,
		RowSpan: rowSpan,
		ColSpan: colSpan,
	})
	return g
}

// SetGap sets the gap between columns and rows.
func (g *VNode) SetGap(columnGap, rowGap int) *VNode {
	g.columnGap = columnGap
	g.rowGap = rowGap
	return g
}

// SetPadding sets the padding (top, right, bottom, left).
func (g *VNode) SetPadding(top, right, bottom, left int) *VNode {
	g.padding = [4]int{top, right, bottom, left}
	return g
}

// SetAlignContent sets the grid alignment in container.
func (g *VNode) SetAlignContent(a rtui.Align) *VNode {
	g.alignContent = a
	return g
}

// SetWidth sets the explicit width.
func (g *VNode) SetWidth(width int) *VNode {
	g.width = width
	return g
}

// SetHeight sets the explicit height.
func (g *VNode) SetHeight(height int) *VNode {
	g.height = height
	return g
}

// SetFlex sets the flex factor.
func (g *VNode) SetFlex(flex int) *VNode {
	g.flex = flex
	return g
}

// SetCells sets all cells.
func (g *VNode) SetCells(cells []Cell) *VNode {
	g.cells = cells
	return g
}

// SetChildrenAuto positions children automatically in row-major order.
func (g *VNode) SetChildrenAuto(children []rtui.VNode) *VNode {
	g.cells = make([]Cell, len(children))
	numCols := max(1, len(g.columns))
	for i, child := range children {
		g.cells[i] = Cell{
			Child:   child,
			Row:     i / numCols,
			Col:     i % numCols,
			RowSpan: 1,
			ColSpan: 1,
		}
	}
	return g
}

// =============================================================================
// Props Accessors
// =============================================================================

// Columns returns the column definitions.
func (g *VNode) Columns() []Dimension {
	return g.columns
}

// Rows returns the row definitions.
func (g *VNode) Rows() []Dimension {
	return g.rows
}

// Cells returns the grid cells.
func (g *VNode) Cells() []Cell {
	return g.cells
}

// ColumnGap returns the gap between columns.
func (g *VNode) ColumnGap() int {
	return g.columnGap
}

// RowGap returns the gap between rows.
func (g *VNode) RowGap() int {
	return g.rowGap
}

// Padding returns the padding [top, right, bottom, left].
func (g *VNode) Padding() [4]int {
	return g.padding
}

// AlignContent returns the grid alignment.
func (g *VNode) AlignContent() rtui.Align {
	return g.alignContent
}

// Width returns the explicit width.
func (g *VNode) Width() int {
	return g.width
}

// Height returns the explicit height.
func (g *VNode) Height() int {
	return g.height
}

// Flex returns the flex factor.
func (g *VNode) Flex() int {
	return g.flex
}

// =============================================================================
// Layout Info (for flex layout)
// =============================================================================

// GetLayoutInfo returns layout information for the layout engine.
func (g *VNode) GetLayoutInfo() rtui.LayoutInfo {
	return rtui.LayoutInfo{
		Flex: g.flex,
	}
}

// MeasureConstraints returns the constraints for Measure.
func (g *VNode) MeasureConstraints(c layout.Constraints) layout.Size {
	inst := g.CreateInstance()
	if measurable, ok := inst.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		return measurable.Measure(c)
	}
	return layout.Size{Width: c.MinWidth, Height: c.MinHeight}
}

// =============================================================================
// ✨ Border Builder Methods (方案 A - 边框作为容器属性)
// =============================================================================

// Border sets the border style and label.
func (g *VNode) Border(style string, label string) *VNode {
	g.borderStyle = style
	g.borderLabel = label
	return g
}

// Bordered sets border with specified style (no label).
func (g *VNode) Bordered(style string) *VNode {
	return g.Border(style, "")
}

// NoBorder removes border.
func (g *VNode) NoBorder() *VNode {
	return g.Border("none", "")
}

// SingleBorder sets single line border with optional label.
func (g *VNode) SingleBorder(label ...string) *VNode {
	lbl := ""
	if len(label) > 0 {
		lbl = label[0]
	}
	return g.Border("single", lbl)
}

// DoubleBorder sets double line border with optional label.
func (g *VNode) DoubleBorder(label ...string) *VNode {
	lbl := ""
	if len(label) > 0 {
		lbl = label[0]
	}
	return g.Border("double", lbl)
}

// RoundedBorder sets rounded border with optional label.
func (g *VNode) RoundedBorder(label ...string) *VNode {
	lbl := ""
	if len(label) > 0 {
		lbl = label[0]
	}
	return g.Border("rounded", lbl)
}

// DashedBorder sets dashed border with optional label.
func (g *VNode) DashedBorder(label ...string) *VNode {
	lbl := ""
	if len(label) > 0 {
		lbl = label[0]
	}
	return g.Border("dashed", lbl)
}

// BorderLabel sets only the border label (keeps current style).
func (g *VNode) BorderLabel(label string) *VNode {
	g.borderLabel = label
	return g
}

// =============================================================================
// ✨ Cell Borders Builder Methods (格子间边框)
// =============================================================================

// SetShowCellBorders sets whether to show cell borders.
func (g *VNode) SetShowCellBorders(show bool) *VNode {
	g.showCellBorders = show
	return g
}

// SetCellBorderStyle sets the cell border style.
func (g *VNode) SetCellBorderStyle(style string) *VNode {
	g.cellBorderStyle = style
	return g
}

// SetCellBorderRounded sets whether cell borders should be rounded.
func (g *VNode) SetCellBorderRounded(rounded bool) *VNode {
	g.cellBorderRounded = rounded
	return g
}

// SetCellBorderColor sets the cell border color.
func (g *VNode) SetCellBorderColor(color string) *VNode {
	g.cellBorderColor = color
	return g
}

// ShowCellBorders shows cell borders (single style by default).
func (g *VNode) ShowCellBorders() *VNode {
	return g.SetShowCellBorders(true)
}

// HideCellBorders hides cell borders.
func (g *VNode) HideCellBorders() *VNode {
	return g.SetShowCellBorders(false)
}

// SingleCellBorders shows cell borders with single line style.
func (g *VNode) SingleCellBorders() *VNode {
	return g.SetShowCellBorders(true).SetCellBorderStyle("single")
}

// DoubleCellBorders shows cell borders with double line style.
func (g *VNode) DoubleCellBorders() *VNode {
	return g.SetShowCellBorders(true).SetCellBorderStyle("double")
}

// LightCellBorders shows cell borders with light line style.
func (g *VNode) LightCellBorders() *VNode {
	return g.SetShowCellBorders(true).SetCellBorderStyle("light")
}

// RoundedCellBorders shows cell borders with rounded corners.
func (g *VNode) RoundedCellBorders() *VNode {
	return g.SetShowCellBorders(true).SetCellBorderRounded(true)
}

// =============================================================================
// Helper Functions
// =============================================================================

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
