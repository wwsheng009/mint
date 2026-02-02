package layout

// =============================================================================
// Grid Layout Component
// =============================================================================

import (
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// GridDimension represents a column or row dimension
type GridDimension interface {
	isGridDimension()
}

// Fixed creates a fixed dimension
type Fixed int

func (f Fixed) isGridDimension() {}

// Flex creates a flexible dimension that takes remaining space
type Flex struct {
	Factor int // Flex factor, defaults to 1
}

func (f Flex) isGridDimension() {}

// Auto creates a dimension that sizes to content
type Auto struct{}

func (a Auto) isGridDimension() {}

// Min creates a dimension with minimum size
type Min struct {
	Min     int
	Content GridDimension
}

func (m Min) isGridDimension() {}

// Max creates a dimension with maximum size
type Max struct {
	Max     int
	Content GridDimension
}

func (m Max) isGridDimension() {}

// GridCell represents a cell in the grid
type GridCell struct {
	Child   ui.VNode
	Row     int // 0-based
	Col     int // 0-based
	RowSpan int
	ColSpan int
}

// GridVNode represents a grid layout container
type GridVNode struct {
	*ui.ElementVNode
	columns      []GridDimension
	rows         []GridDimension
	cells        []GridCell
	gap          [2]int // [columnGap, rowGap]
	padding      [4]int // top, right, bottom, left
	alignContent ui.Align // Alignment of the whole grid in the container
}

// NewGrid creates a new grid
func NewGrid() *GridVNode {
	return &GridVNode{
		ElementVNode: ui.NewElement("grid"),
		columns:      []GridDimension{Flex{Factor: 1}},
		rows:         []GridDimension{Auto{}},
		cells:        make([]GridCell, 0),
		gap:          [2]int{0, 0},
		padding:      [4]int{0, 0, 0, 0},
		alignContent: ui.AlignStart,
	}
}

// Grid creates a new grid node
func Grid(columns, rows []GridDimension, cells ...ui.VNode) ui.VNode {
	grid := NewGrid()
	grid.columns = columns
	grid.rows = rows

	// Add cells - if cells are provided without position info,
	// position them automatically
	for i, cell := range cells {
		col := i % len(columns)
		row := i / len(columns)
		grid.cells = append(grid.cells, GridCell{
			Child:   cell,
			Row:     row,
			Col:     col,
			RowSpan: 1,
			ColSpan: 1,
		})
	}

	return grid
}

// Builder pattern
type GridBuilderType struct {
	grid     *GridVNode
	cellDefs []GridCell
}

// GridBuilder creates a new grid builder
func GridBuilder() *GridBuilderType {
	return &GridBuilderType{
		grid:     NewGrid(),
		cellDefs: make([]GridCell, 0),
	}
}

// Columns sets the column definitions
func (b *GridBuilderType) Columns(dims ...GridDimension) *GridBuilderType {
	b.grid.columns = dims
	return b
}

// Rows sets the row definitions
func (b *GridBuilderType) Rows(dims ...GridDimension) *GridBuilderType {
	b.grid.rows = dims
	return b
}

// Cell adds a cell at the specified position
func (b *GridBuilderType) Cell(row, col int, child ui.VNode) *GridBuilderType {
	b.cellDefs = append(b.cellDefs, GridCell{
		Child:   child,
		Row:     row,
		Col:     col,
		RowSpan: 1,
		ColSpan: 1,
	})
	return b
}

// CellSpan adds a cell that spans multiple rows and columns
func (b *GridBuilderType) CellSpan(row, col, rowSpan, colSpan int, child ui.VNode) *GridBuilderType {
	b.cellDefs = append(b.cellDefs, GridCell{
		Child:   child,
		Row:     row,
		Col:     col,
		RowSpan: rowSpan,
		ColSpan: colSpan,
	})
	return b
}

// Gap sets the gap between columns and rows
func (b *GridBuilderType) Gap(columnGap, rowGap int) *GridBuilderType {
	b.grid.gap = [2]int{columnGap, rowGap}
	return b
}

// Padding sets the padding
func (b *GridBuilderType) Padding(top, right, bottom, left int) *GridBuilderType {
	b.grid.padding = [4]int{top, right, bottom, left}
	return b
}

// AlignContent sets the grid alignment
func (b *GridBuilderType) AlignContent(a ui.Align) *GridBuilderType {
	b.grid.alignContent = a
	return b
}

// Width sets the width
func (b *GridBuilderType) Width(n int) *GridBuilderType {
	b.grid.SetProp("width", n)
	return b
}

// Height sets the height
func (b *GridBuilderType) Height(n int) *GridBuilderType {
	b.grid.SetProp("height", n)
	return b
}

// Style sets the visual style
func (b *GridBuilderType) Style(s style.Style) *GridBuilderType {
	b.grid.SetStyle(s)
	return b
}

// FgColor sets the foreground color
func (b *GridBuilderType) FgColor(c interface{}) *GridBuilderType {
	if colorStr, ok := c.(string); ok {
		s := b.grid.Style()
		s.FG = style.Color(colorStr)
		b.grid.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.grid.Style()
		s.FG = color
		b.grid.SetStyle(s)
	}
	return b
}

// BgColor sets the background color
func (b *GridBuilderType) BgColor(c interface{}) *GridBuilderType {
	if colorStr, ok := c.(string); ok {
		s := b.grid.Style()
		s.BG = style.Color(colorStr)
		b.grid.SetStyle(s)
	} else if color, ok := c.(style.Color); ok {
		s := b.grid.Style()
		s.BG = color
		b.grid.SetStyle(s)
	}
	return b
}

// Key sets the key for diffing
func (b *GridBuilderType) Key(key string) *GridBuilderType {
	b.grid.SetKey(key)
	return b
}

// Build returns the grid ui.VNode
func (b *GridBuilderType) Build() ui.VNode {
	b.grid.cells = b.cellDefs
	return b.grid
}

// Getters
func (g *GridVNode) Columns() []GridDimension { return g.columns }
func (g *GridVNode) Rows() []GridDimension    { return g.rows }
func (g *GridVNode) Cells() []GridCell       { return g.cells }
func (g *GridVNode) Gap() [2]int             { return g.gap }
func (g *GridVNode) Padding() [4]int         { return g.padding }
func (g *GridVNode) AlignContent() ui.Align   { return g.alignContent }

// Setters
func (g *GridVNode) SetColumns(cols []GridDimension) { g.columns = cols }
func (g *GridVNode) SetRows(rows []GridDimension)    { g.rows = rows }
func (g *GridVNode) SetCells(cells []GridCell)       { g.cells = cells }
func (g *GridVNode) SetGap(colGap, rowGap int)       { g.gap = [2]int{colGap, rowGap} }
func (g *GridVNode) SetPadding(top, right, bottom, left int) {
	g.padding = [4]int{top, right, bottom, left}
}
func (g *GridVNode) SetAlignContent(a ui.Align) { g.alignContent = a }

// CalculateColumnWidths calculates the actual width for each column
func (g *GridVNode) CalculateColumnWidths(totalWidth int) []int {
	numCols := len(g.columns)
	if numCols == 0 {
		return []int{totalWidth}
	}

	widths := make([]int, numCols)
	fixedWidth := 0
	flexCount := 0
	flexTotalFactor := 0

	// First pass: calculate fixed widths
	for i, col := range g.columns {
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
			// Auto columns get minimum width, will be adjusted if there's extra space
			widths[i] = 10 // minimum
			fixedWidth += widths[i]
		case Min:
			widths[i] = c.Min
			fixedWidth += widths[i]
		case Max:
			widths[i] = c.Max
			fixedWidth += widths[i]
		}
	}

	// Calculate remaining width for flex columns
	remainingWidth := totalWidth - fixedWidth - g.padding[1] - g.padding[3]
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	// Second pass: distribute remaining width to flex columns
	if flexCount > 0 && flexTotalFactor > 0 {
		for i, col := range g.columns {
			if _, ok := col.(Flex); ok {
				// Find the flex factor
				var factor int
				if f, ok := col.(Flex); ok && f.Factor > 0 {
					factor = f.Factor
				} else {
					factor = 1
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

// CalculateRowHeights calculates the actual height for each row
func (g *GridVNode) CalculateRowHeights(totalHeight int) []int {
	numRows := len(g.rows)
	if numRows == 0 {
		return []int{totalHeight}
	}

	heights := make([]int, numRows)
	fixedHeight := 0
	flexCount := 0
	flexTotalFactor := 0

	// First pass: calculate fixed heights
	for i, row := range g.rows {
		switch r := row.(type) {
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
			// Auto rows get minimum height
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

	// Calculate remaining height for flex rows
	remainingHeight := totalHeight - fixedHeight - g.padding[0] - g.padding[2]
	if remainingHeight < 0 {
		remainingHeight = 0
	}

	// Second pass: distribute remaining height to flex rows
	if flexCount > 0 && flexTotalFactor > 0 {
		for i, row := range g.rows {
			if _, ok := row.(Flex); ok {
				var factor int
				if f, ok := row.(Flex); ok && f.Factor > 0 {
					factor = f.Factor
				} else {
					factor = 1
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
// Measurable & Paintable Interface Implementation
// =============================================================================

// Measure implements runtime.Measurable interface
func (g *GridVNode) Measure(constraints runtime.BoxConstraints) runtime.Size {
	if g == nil {
		return runtime.Size{Width: 0, Height: 0}
	}

	// Default size for grid container
	width := 80
	height := 24

	// Check if explicit dimensions are set
	elemStyle := g.Style()
	if elemStyle.Width > 0 {
		width = elemStyle.Width
	}
	if elemStyle.Height > 0 {
		height = elemStyle.Height
	}

	// Apply constraints
	if width < constraints.MinWidth {
		width = constraints.MinWidth
	}
	if width > constraints.MaxWidth && constraints.MaxWidth > 0 {
		width = constraints.MaxWidth
	}
	if height < constraints.MinHeight {
		height = constraints.MinHeight
	}
	if height > constraints.MaxHeight && constraints.MaxHeight > 0 {
		height = constraints.MaxHeight
	}

	return runtime.Size{Width: width, Height: height}
}

// Paint implements paint.Paintable interface
// Grid is primarily a layout container - actual rendering is handled by the layout engine
func (g *GridVNode) Paint(x, y int) []paint.DrawCmd {
	if g == nil {
		return nil
	}

	// Grid container itself doesn't have visual representation
	// The layout engine will position and render children
	// Return empty command set - children will be painted by the reconciler
	return []paint.DrawCmd{}
}
