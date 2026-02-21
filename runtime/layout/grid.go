// Package layout provides grid layout types
package layout

// =============================================================================
// Grid Dimension Types
// =============================================================================

// GridDimension represents a column or row dimension in the grid.
type GridDimension interface {
	isGridDimension()
}

// GridFixed creates a fixed-size dimension.
type GridFixed int

func (f GridFixed) isGridDimension() {}

// GridFlex creates a flexible dimension that takes remaining space.
type GridFlex struct {
	Factor int // Flex factor, defaults to 1
}

func (f GridFlex) isGridDimension() {}

// GridAuto creates a dimension that sizes to content.
type GridAuto struct{}

func (a GridAuto) isGridDimension() {}

// GridMin creates a dimension with minimum size.
type GridMin struct {
	Min     int
	Content GridDimension
}

func (m GridMin) isGridDimension() {}

// GridMax creates a dimension with maximum size.
type GridMax struct {
	Max     int
	Content GridDimension
}

func (m GridMax) isGridDimension() {}

// =============================================================================
// Grid Cell
// =============================================================================

// GridCell represents a cell in the grid.
type GridCell struct {
	Child   Node // The child node (implements layout.Node)
	Row     int  // 0-based
	Col     int  // 0-based
	RowSpan int  // default 1
	ColSpan int  // default 1
}

// =============================================================================
// Grid Style
// =============================================================================

// GridStyle defines grid layout properties
type GridStyle struct {
	// Columns defines column dimensions
	Columns []GridDimension

	// Rows defines row dimensions
	Rows []GridDimension

	// Cells defines grid cells with positions
	Cells []GridCell

	// ColumnGap gap between columns
	ColumnGap int

	// RowGap gap between rows
	RowGap int

	// Padding inner spacing
	Padding Padding

	// Width explicit width (0 = auto)
	Width int

	// Height explicit height (0 = auto)
	Height int
}

// GridStyleProvider defines the interface for grid containers
type GridStyleProvider interface {
	GetGridStyle() *GridStyle
}

// DefaultGridStyle creates default grid style
func DefaultGridStyle() *GridStyle {
	return &GridStyle{
		Columns:    []GridDimension{GridFlex{Factor: 1}},
		Rows:       []GridDimension{GridAuto{}},
		ColumnGap:  0,
		RowGap:     0,
		Padding:    Padding{},
	}
}

// =============================================================================
// GridLayout
// =============================================================================

// GridLayout implements grid layout algorithm
type GridLayout struct {
	id       string
	style    *GridStyle
	children []Node
	size     Size
	position Point

	// Calculated sizes
	colWidths  []int
	rowHeights []int
}

// NewGridLayout creates a new grid layout
func NewGridLayout(id string, style *GridStyle) *GridLayout {
	if style == nil {
		style = DefaultGridStyle()
	}
	return &GridLayout{
		id:    id,
		style: style,
	}
}

// SetChildren sets the children nodes
func (g *GridLayout) SetChildren(children []Node) {
	g.children = children
}

// ID returns the node identifier
func (g *GridLayout) ID() string {
	return g.id
}

// Type returns the node type
func (g *GridLayout) Type() string {
	return "grid"
}

// Children returns child nodes
func (g *GridLayout) Children() []Node {
	return g.children
}

// GetPosition returns the current position
func (g *GridLayout) GetPosition() (int, int) {
	return g.position.X, g.position.Y
}

// SetPosition sets the position
func (g *GridLayout) SetPosition(x, y int) {
	g.position.X = x
	g.position.Y = y
}

// GetSize returns the current size
func (g *GridLayout) GetSize() (int, int) {
	return g.size.Width, g.size.Height
}

// SetSize sets the size
func (g *GridLayout) SetSize(width, height int) {
	g.size.Width = width
	g.size.Height = height
}

// GetWidth returns the width
func (g *GridLayout) GetWidth() int {
	return g.size.Width
}

// GetHeight returns the height
func (g *GridLayout) GetHeight() int {
	return g.size.Height
}

// GetGridStyle returns the grid style
func (g *GridLayout) GetGridStyle() *GridStyle {
	return g.style
}

// Measure measures the grid size given constraints
func (g *GridLayout) Measure(constraints Constraints) Size {
	numCols := len(g.style.Columns)
	if numCols == 0 {
		numCols = 1
	}

	numRows := g.calculateRowCount()

	// Calculate column widths
	g.colWidths = g.calculateColumnWidths(constraints.MaxWidth - g.style.Padding.Left - g.style.Padding.Right)

	// Calculate row heights
	g.rowHeights = g.calculateRowHeights(constraints.MaxHeight - g.style.Padding.Top - g.style.Padding.Bottom, numCols)

	// Calculate total size
	totalW := g.style.Padding.Left + g.style.Padding.Right
	totalH := g.style.Padding.Top + g.style.Padding.Bottom

	for _, w := range g.colWidths {
		totalW += w
	}
	for _, h := range g.rowHeights {
		totalH += h
	}

	// Add gaps
	if numCols > 1 {
		totalW += g.style.ColumnGap * (numCols - 1)
	}
	if numRows > 1 {
		totalH += g.style.RowGap * (numRows - 1)
	}

	// Apply explicit dimensions
	if g.style.Width > 0 {
		totalW = g.style.Width
	}
	if g.style.Height > 0 {
		totalH = g.style.Height
	}

	// Apply constraints
	totalW = constraints.ConstrainWidth(totalW)
	totalH = constraints.ConstrainHeight(totalH)

	return Size{Width: totalW, Height: totalH}
}

// calculateRowCount calculates the number of rows needed
func (g *GridLayout) calculateRowCount() int {
	if len(g.style.Cells) > 0 {
		maxRow := 0
		for _, cell := range g.style.Cells {
			endRow := cell.Row + cell.RowSpan
			if endRow > maxRow {
				maxRow = endRow
			}
		}
		return maxRow
	}

	// Auto-calculate from children and columns
	numCols := len(g.style.Columns)
	if numCols == 0 {
		numCols = 1
	}
	numChildren := len(g.children)
	return (numChildren + numCols - 1) / numCols
}

// calculateColumnWidths calculates the actual width for each column
func (g *GridLayout) calculateColumnWidths(availableWidth int) []int {
	numCols := len(g.style.Columns)
	if numCols == 0 {
		return []int{availableWidth}
	}

	widths := make([]int, numCols)
	fixedWidth := 0
	flexCount := 0
	flexTotalFactor := 0

	// First pass: calculate fixed widths
	for i, col := range g.style.Columns {
		switch c := col.(type) {
		case GridFixed:
			widths[i] = int(c)
			fixedWidth += widths[i]
		case GridFlex:
			flexCount++
			if c.Factor > 0 {
				flexTotalFactor += c.Factor
			} else {
				flexTotalFactor += 1
			}
		case GridAuto:
			// Auto columns get minimum width, will be expanded later
			widths[i] = 0
		case GridMin:
			widths[i] = c.Min
			fixedWidth += widths[i]
		case GridMax:
			widths[i] = c.Max
			fixedWidth += widths[i]
		}
	}

	// Subtract column gaps from available width
	gapWidth := g.style.ColumnGap * (numCols - 1)
	if gapWidth < 0 {
		gapWidth = 0
	}
	remainingWidth := availableWidth - fixedWidth - gapWidth
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	// Second pass: distribute remaining width to flex columns
	if flexCount > 0 && flexTotalFactor > 0 {
		for i, col := range g.style.Columns {
			if _, ok := col.(GridFlex); ok {
				factor := 1
				if f, ok := col.(GridFlex); ok && f.Factor > 0 {
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

// calculateRowHeights calculates the actual height for each row
func (g *GridLayout) calculateRowHeights(availableHeight, numCols int) []int {
	numRows := g.calculateRowCount()
	if numRows == 0 {
		numRows = 1
	}

	// Ensure we have row definitions for all rows
	rows := g.style.Rows
	if len(rows) < numRows {
		// Expand rows to match numRows
		expanded := make([]GridDimension, numRows)
		copy(expanded, rows)
		for i := len(rows); i < numRows; i++ {
			expanded[i] = GridAuto{} // Default to Auto for undefined rows
		}
		rows = expanded
	}

	heights := make([]int, numRows)
	fixedHeight := 0
	flexCount := 0
	flexTotalFactor := 0

	// First pass: calculate fixed heights and measure auto rows
	for i := 0; i < numRows && i < len(rows); i++ {
		switch r := rows[i].(type) {
		case GridFixed:
			heights[i] = int(r)
			fixedHeight += heights[i]
		case GridFlex:
			flexCount++
			if r.Factor > 0 {
				flexTotalFactor += r.Factor
			} else {
				flexTotalFactor += 1
			}
		case GridAuto:
			// Measure content height for this row
			heights[i] = g.measureRowHeight(i, numCols)
			fixedHeight += heights[i]
		case GridMin:
			heights[i] = r.Min
			fixedHeight += heights[i]
		case GridMax:
			heights[i] = r.Max
			fixedHeight += heights[i]
		}
	}

	// Subtract row gaps from available height
	gapHeight := g.style.RowGap * (numRows - 1)
	if gapHeight < 0 {
		gapHeight = 0
	}
	remainingHeight := availableHeight - fixedHeight - gapHeight
	if remainingHeight < 0 {
		remainingHeight = 0
	}

	// Second pass: distribute remaining height to flex rows
	if flexCount > 0 && flexTotalFactor > 0 {
		for i := 0; i < numRows && i < len(rows); i++ {
			if _, ok := rows[i].(GridFlex); ok {
				factor := 1
				if f, ok := rows[i].(GridFlex); ok && f.Factor > 0 {
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

// measureRowHeight measures the maximum height of children in a row
func (g *GridLayout) measureRowHeight(row, numCols int) int {
	maxHeight := 1 // minimum row height

	// Check explicit cells
	for _, cell := range g.style.Cells {
		if cell.Row == row && cell.Child != nil {
			rowSpan := cell.RowSpan
			if rowSpan <= 0 {
				rowSpan = 1
			}
			if measurable, ok := cell.Child.(Measurable); ok {
				size := measurable.Measure(Constraints{
					MaxWidth:  MaxInt,
					MaxHeight: MaxInt,
				})
				h := size.Height / rowSpan
				if h > maxHeight {
					maxHeight = h
				}
			} else {
				_, h := cell.Child.GetSize()
				h = h / rowSpan
				if h > maxHeight {
					maxHeight = h
				}
			}
		}
	}

	// Check auto-positioned children
	if len(g.style.Cells) == 0 && len(g.children) > 0 {
		for i, child := range g.children {
			if child == nil {
				continue
			}
			childRow := i / numCols
			if childRow == row {
				if measurable, ok := child.(Measurable); ok {
					size := measurable.Measure(Constraints{
						MaxWidth:  MaxInt,
						MaxHeight: MaxInt,
					})
					if size.Height > maxHeight {
						maxHeight = size.Height
					}
				} else {
					_, h := child.GetSize()
					if h > maxHeight {
						maxHeight = h
					}
				}
			}
		}
	}

	return maxHeight
}

// LayoutChildren lays out all children and returns their LayoutBoxes
func (g *GridLayout) LayoutChildren(width, height int) []LayoutBox {
	if len(g.style.Cells) == 0 && len(g.children) == 0 {
		return nil
	}

	numCols := len(g.style.Columns)
	if numCols == 0 {
		numCols = 1
	}

	// Recalculate sizes with actual constraints
	constraints := TightConstraints(width, height)
	g.Measure(constraints)

	boxes := make([]LayoutBox, 0)

	// Layout explicit cells
	if len(g.style.Cells) > 0 {
		for _, cell := range g.style.Cells {
			if cell.Child == nil {
				continue
			}

			x, y := g.getCellPosition(cell.Row, cell.Col)
			w, h := g.getCellSize(cell.Row, cell.Col, cell.RowSpan, cell.ColSpan)

			box := LayoutBox{
				ID:     cell.Child.ID(),
				X:      x,
				Y:      y,
				Width:  w,
				Height: h,
			}
			boxes = append(boxes, box)

			cell.Child.SetPosition(x, y)
			cell.Child.SetSize(w, h)
		}
	} else {
		// Auto-position children in row-major order
		for i, child := range g.children {
			if child == nil {
				continue
			}

			row := i / numCols
			col := i % numCols

			x, y := g.getCellPosition(row, col)
			w, h := g.getCellSize(row, col, 1, 1)

			box := LayoutBox{
				ID:     child.ID(),
				X:      x,
				Y:      y,
				Width:  w,
				Height: h,
			}
			boxes = append(boxes, box)

			child.SetPosition(x, y)
			child.SetSize(w, h)
		}
	}

	return boxes
}

// getCellPosition returns the x, y position of a cell
func (g *GridLayout) getCellPosition(row, col int) (x, y int) {
	// Add padding
	x = g.style.Padding.Left
	y = g.style.Padding.Top

	// Sum widths of columns before this one
	for i := 0; i < col && i < len(g.colWidths); i++ {
		x += g.colWidths[i]
	}

	// Sum heights of rows before this one
	for i := 0; i < row && i < len(g.rowHeights); i++ {
		y += g.rowHeights[i]
	}

	// Add gaps
	if col > 0 {
		x += g.style.ColumnGap * col
	}
	if row > 0 {
		y += g.style.RowGap * row
	}

	return x, y
}

// getCellSize returns the width, height allocated to a cell (with span)
func (g *GridLayout) getCellSize(row, col, rowSpan, colSpan int) (width, height int) {
	// Sum column widths for the span
	for i := col; i < col+colSpan && i < len(g.colWidths); i++ {
		width += g.colWidths[i]
	}

	// Sum row heights for the span
	for i := row; i < row+rowSpan && i < len(g.rowHeights); i++ {
		height += g.rowHeights[i]
	}

	// Add internal gaps for spans
	if colSpan > 1 {
		width += g.style.ColumnGap * (colSpan - 1)
	}
	if rowSpan > 1 {
		height += g.style.RowGap * (rowSpan - 1)
	}

	return width, height
}

// GetColumnWidths returns calculated column widths
func (g *GridLayout) GetColumnWidths() []int {
	return g.colWidths
}

// GetRowHeights returns calculated row heights
func (g *GridLayout) GetRowHeights() []int {
	return g.rowHeights
}
