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

	// ✨ Cell Borders 格子边框
	ShowCellBorders   bool   // 是否显示格子边框
	CellBorderWidth   int    // 边框宽度（每条线 1 字符）
	CellBorderHeight  int    // 边框高度（每条线 1 字符）
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

	// ✨ Cell Borders: 计算边框占用的空间
	// 边框每条线占 1 字符
	// Grid 宽/高度包含边框：左边框(1) + 中间(numCols-1) + 右边框(1) = numCols + 1
	// 水平边框: 上(1) + 中间(numRows-1) + 下(1) = numRows + 1
	// 垂直边框: 左(1) + 中间(numCols-1) + 右(1) = numCols + 1
	cellBorderWidth := 0
	cellBorderHeight := 0
	if g.style.ShowCellBorders {
		cellBorderWidth = numCols + 1
		cellBorderHeight = numRows + 1
	}

	// ✨ 计算可用宽度（减去 padding 和边框）
	availableWidth := constraints.MaxWidth - g.style.Padding.Left - g.style.Padding.Right - cellBorderWidth
	if availableWidth < 0 {
		availableWidth = 0
	}

	// ✨ 计算可用高度（减去 padding 和边框）
	availableHeight := constraints.MaxHeight - g.style.Padding.Top - g.style.Padding.Bottom - cellBorderHeight
	if availableHeight < 0 {
		availableHeight = 0
	}

	// Calculate column widths and row heights（计算的是纯内容宽度）
	g.colWidths = g.calculateColumnWidths(availableWidth)
	g.rowHeights = g.calculateRowHeights(availableHeight, numCols)

	// Calculate total size
	totalW := g.style.Padding.Left + g.style.Padding.Right + cellBorderWidth
	totalH := g.style.Padding.Top + g.style.Padding.Bottom + cellBorderHeight

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

	// ✨ Check if explicit rows are defined
	if len(g.style.Rows) > 0 {
		return len(g.style.Rows)
	}

	// Auto-calculate from children and columns
	numCols := len(g.style.Columns)
	if numCols == 0 {
		numCols = 1
	}
	numChildren := len(g.children)
	result := (numChildren + numCols - 1) / numCols
	if result < 1 {
		result = 1
	}
	return result
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

	// ✨ FIX: 处理无限可用宽度的情况
	// 如果 availableWidth 是 MaxInt（无限宽度），flex 列应该只分配最小宽度
	// 而不是使用 (remainingWidth * factor) / flexTotalFactor 产生不合理的大值
	if flexCount > 0 && flexTotalFactor > 0 {
		for i, col := range g.style.Columns {
			if _, ok := col.(GridFlex); ok {
				// 检测可用宽度是否无限
				isInfinite := availableWidth >= MaxInt || availableWidth > 1000000
				if isInfinite {
					// 无限宽度：使用最小宽度 1（让后续 Measure 计算自然尺寸）
					widths[i] = 1
				} else {
					// 有限宽度：按 flex factor 分配剩余空间
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

	// ✨ FIX: 处理无限可用高度的情况
	// 如果 availableHeight 是 MaxInt（无限高度），flex 行应该只分配最小高度
	// 而不是使用 (remainingHeight * factor) / flexTotalFactor 产生不合理的大值
	if flexCount > 0 && flexTotalFactor > 0 {
		for i := 0; i < numRows && i < len(rows); i++ {
			if _, ok := rows[i].(GridFlex); ok {
				// 检测可用高度是否无限
				isInfinite := availableHeight >= MaxInt || availableHeight > 1000000
				if isInfinite {
					// 无限高度：使用最小高度 1（让后续 Measure 计算自然尺寸）
					heights[i] = 1
				} else {
					// 有限高度：按 flex factor 分配剩余空间
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

			contentX, contentY := g.getCellPosition(cell.Row, cell.Col)
			w, h := g.getCellSize(cell.Row, cell.Col, cell.RowSpan, cell.ColSpan)

			// Add padding to get absolute position (like flex.go and wrap.go)
			x := g.style.Padding.Left + contentX
			y := g.style.Padding.Top + contentY

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

			contentX, contentY := g.getCellPosition(row, col)
			w, h := g.getCellSize(row, col, 1, 1)

			// Add padding to get absolute position (like flex.go and wrap.go)
			x := g.style.Padding.Left + contentX
			y := g.style.Padding.Top + contentY

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

// getCellPosition returns the x, y position of a cell relative to content area
// ✨ 返回内容区域的起始位置（子节点的渲染起始位置）
// 如果有边框，返回边框内内容区域的起始位置
func (g *GridLayout) getCellPosition(row, col int) (x, y int) {
	// ✨ 相对于内容空间起点（不包含 padding）
	x = 0
	y = 0

	// ✨ 添加 cell 边框的偏移（每条边框占 1 字符，位于格子之间）
	// 格子内容需要跳过上边框和左边框
	//
	// 边框计算方式：
	// - colWidths 和 rowHeights 是纯内容宽度/高度（不包含边框）
	// - 每条边框占 1 字符
	// - 垂直边框位于列之间：第 0 条在 x=0，第 1 条在 x=colWidths[0]+1，第 2 条在 x=colWidths[0]+1+colWidths[1]+1...
	// - 水平边框位于行之间：第 0 条在 y=0，第 1 条在 y=rowHeights[0]+1，第 2 条在 y=rowHeights[0]+1+rowHeights[1]+1...
	//
	if g.style.ShowCellBorders {
		// 第 col 条垂直线的位置 = 前面内容宽度累积 + 前面边框数量
		// 内容起始位置 = 第 col 条垂直线位置 + 1（垂直线占 1 个字符）
		for c := 0; c < col; c++ {
			x += g.colWidths[c] + 1  // 内容宽度 + 右边框
			if c < col-1 {
				x += g.style.ColumnGap  // 列间距在格子之间
			}
		}
		// 上边框位于 x=0，所以内容从 x+1 开始（跳过上边框）
		x += 1
	} else {
		// Sum widths of columns before this one (无边框时)
		for i := 0; i < col && i < len(g.colWidths); i++ {
			x += g.colWidths[i]
		}

		// Add gaps between columns
		if col > 0 {
			x += g.style.ColumnGap * col
		}
	}

	// Y 坐标处理（与 X 对称）
	if g.style.ShowCellBorders {
		// 第 row 条水平线的位置 = 前面内容高度累积 + 前面边框高度
		// 内容起始位置 = 第 row 条水平线位置 + 1（水平线占 1 个字符）
		for r := 0; r < row; r++ {
			y += g.rowHeights[r] + 1  // 内容高度 + 下边框
			if r < row-1 {
				y += g.style.RowGap  // 行间距在格子之间
			}
		}
		// 上边框位于 y=0，所以内容从 y+1 开始（跳过上边框）
		y += 1
	} else {
		// Sum heights of rows before this one (无边框时)
		for i := 0; i < row && i < len(g.rowHeights); i++ {
			y += g.rowHeights[i]
		}

		// Add gaps between rows
		if row > 0 {
			y += g.style.RowGap * row
		}
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

	// ✨ Cell 边框不包含在 getCellSize 中
	// getCellSize 返回的是**纯内容区域**的大小
	// 边框由 Paint 绘制，不影响内容区域的大小
	//
	// 关键：
	// - getCellPosition 返回的是内容区域的起始位置（跳过左边框+上边框）
	// - getCellSize 返回的是内容区域的大小（不包含边框）
	// - 边框由 Grid.Paint() 单独绘制
	//
	// 示例：对于 cell (0,0)
	//   - getCellPosition 返回 (1, 1)（跳过第0条左边框和上边框）
	//   - getCellSize 返回 (10, 5)（纯内容大小）
	//   - Cell 绘制范围：[1, 10] 在边框 [0, 11] 之间

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
