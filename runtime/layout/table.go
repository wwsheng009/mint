// Package layout provides table layout types
package layout

// =============================================================================
// Table Types
// =============================================================================

// TableLayout represents a table with rows and columns
type TableLayout struct {
	id       string
	rows     [][]Node
	colWidths []int  // Cached column widths
	rowHeights []int // Cached row heights
}

// NewTableLayout creates a new table layout
func NewTableLayout(id string, rows [][]Node) *TableLayout {
	return &TableLayout{
		id:    id,
		rows:  rows,
	}
}

// ID returns the node identifier
func (t *TableLayout) ID() string {
	return t.id
}

// Type returns the node type
func (t *TableLayout) Type() string {
	return "table"
}

// Children returns all cells as a flat list
func (t *TableLayout) Children() []Node {
	var children []Node
	for _, row := range t.rows {
		children = append(children, row...)
	}
	return children
}

// Rows returns the table rows
func (t *TableLayout) Rows() [][]Node {
	return t.rows
}

// GetPosition returns the current position
func (t *TableLayout) GetPosition() (x, y int) {
	return 0, 0
}

// SetPosition sets the position
func (t *TableLayout) SetPosition(x, y int) {
	// Position is handled by parent layout
}

// GetSize returns the current size
func (t *TableLayout) GetSize() (width, height int) {
	if len(t.colWidths) == 0 || len(t.rowHeights) == 0 {
		t.calculateSizes()
	}

	for _, w := range t.colWidths {
		width += w
	}
	for _, h := range t.rowHeights {
		height += h
	}
	return width, height
}

// SetSize sets the size
func (t *TableLayout) SetSize(width, height int) {
	// Size is calculated during layout
}

// GetWidth returns the width
func (t *TableLayout) GetWidth() int {
	w, _ := t.GetSize()
	return w
}

// GetHeight returns the height
func (t *TableLayout) GetHeight() int {
	_, h := t.GetSize()
	return h
}

// ColumnWidths returns the calculated column widths
func (t *TableLayout) ColumnWidths() []int {
	if len(t.colWidths) == 0 {
		t.calculateSizes()
	}
	return t.colWidths
}

// RowHeights returns the calculated row heights
func (t *TableLayout) RowHeights() []int {
	if len(t.rowHeights) == 0 {
		t.calculateSizes()
	}
	return t.rowHeights
}

// calculateSizes calculates column widths and row heights
func (t *TableLayout) calculateSizes() {
	if len(t.rows) == 0 {
		return
	}

	// Determine number of columns (max row length)
	numCols := 0
	for _, row := range t.rows {
		if len(row) > numCols {
			numCols = len(row)
		}
	}

	// Initialize column widths
	t.colWidths = make([]int, numCols)
	t.rowHeights = make([]int, len(t.rows))

	// First pass: find max width for each column and max height for each row
	for rowIdx, row := range t.rows {
		maxRowHeight := 0
		for colIdx, cell := range row {
			if cell == nil {
				continue
			}

			// Get cell size
			w, h := cell.GetSize()

			// Update column width (max of all cells in column)
			if w > t.colWidths[colIdx] {
				t.colWidths[colIdx] = w
			}

			// Update row height (max of all cells in row)
			if h > maxRowHeight {
				maxRowHeight = h
			}
		}
		t.rowHeights[rowIdx] = maxRowHeight
	}
}

// CellPosition returns the x, y position of a cell
func (t *TableLayout) CellPosition(row, col int) (x, y int) {
	if len(t.colWidths) == 0 {
		t.calculateSizes()
	}

	// Sum widths of columns before this one
	for i := 0; i < col && i < len(t.colWidths); i++ {
		x += t.colWidths[i]
	}

	// Sum heights of rows before this one
	for i := 0; i < row && i < len(t.rowHeights); i++ {
		y += t.rowHeights[i]
	}

	return x, y
}

// CellSize returns the width, height allocated to a cell
func (t *TableLayout) CellSize(row, col int) (width, height int) {
	if len(t.colWidths) == 0 {
		t.calculateSizes()
	}

	if col < len(t.colWidths) {
		width = t.colWidths[col]
	}
	if row < len(t.rowHeights) {
		height = t.rowHeights[row]
	}

	return width, height
}

// Measure measures the table size given constraints
func (t *TableLayout) Measure(constraints Constraints) Size {
	t.calculateSizes()

	width := 0
	height := 0

	for _, w := range t.colWidths {
		width += w
	}
	for _, h := range t.rowHeights {
		height += h
	}

	// Apply constraints
	width = constraints.ConstrainWidth(width)
	height = constraints.ConstrainHeight(height)

	return Size{Width: width, Height: height}
}

// =============================================================================
// TableRow
// =============================================================================

// TableRow represents a single row in a table
type TableRow struct {
	id       string
	cells    []Node
	height   int
}

// NewTableRow creates a new table row
func NewTableRow(id string, cells []Node) *TableRow {
	return &TableRow{
		id:     id,
		cells:  cells,
	}
}

// ID returns the node identifier
func (r *TableRow) ID() string {
	return r.id
}

// Type returns the node type
func (r *TableRow) Type() string {
	return "tableRow"
}

// Children returns the cell nodes
func (r *TableRow) Children() []Node {
	return r.cells
}

// Cells returns the cells
func (r *TableRow) Cells() []Node {
	return r.cells
}

// GetPosition returns the current position
func (r *TableRow) GetPosition() (x, y int) {
	return 0, 0
}

// SetPosition sets the position
func (r *TableRow) SetPosition(x, y int) {
}

// GetSize returns the current size
func (r *TableRow) GetSize() (width, height int) {
	for _, cell := range r.cells {
		if cell != nil {
			w, h := cell.GetSize()
			width += w
			if h > height {
				height = h
			}
		}
	}
	return width, height
}

// SetSize sets the size
func (r *TableRow) SetSize(width, height int) {
}

// GetWidth returns the width
func (r *TableRow) GetWidth() int {
	w, _ := r.GetSize()
	return w
}

// GetHeight returns the height
func (r *TableRow) GetHeight() int {
	_, h := r.GetSize()
	return h
}

// =============================================================================
// TableCell
// =============================================================================

// TableCell represents a single cell in a table
type TableCell struct {
	id        string
	content   Node
	colSpan   int
	rowSpan   int
	align     Align
}

// Align represents cell content alignment
type Align int

const (
	// AlignLeft aligns content to the left
	AlignLeft Align = iota
	// AlignCenter centers content
	AlignCenter
	// AlignRight aligns content to the right
	AlignRight
)

// NewTableCell creates a new table cell
func NewTableCell(id string, content Node) *TableCell {
	return &TableCell{
		id:      id,
		content: content,
		colSpan: 1,
		rowSpan: 1,
	}
}

// NewTableCellWithSpan creates a new table cell with column and row span
func NewTableCellWithSpan(id string, content Node, colSpan, rowSpan int) *TableCell {
	return &TableCell{
		id:      id,
		content: content,
		colSpan: colSpan,
		rowSpan: rowSpan,
	}
}

// ID returns the node identifier
func (c *TableCell) ID() string {
	return c.id
}

// Type returns the node type
func (c *TableCell) Type() string {
	return "tableCell"
}

// Children returns the content node
func (c *TableCell) Children() []Node {
	if c.content == nil {
		return nil
	}
	return []Node{c.content}
}

// Content returns the cell content
func (c *TableCell) Content() Node {
	return c.content
}

// ColSpan returns the column span
func (c *TableCell) ColSpan() int {
	return c.colSpan
}

// RowSpan returns the row span
func (c *TableCell) RowSpan() int {
	return c.rowSpan
}

// Align returns the content alignment
func (c *TableCell) Align() Align {
	return c.align
}

// SetAlign sets the content alignment
func (c *TableCell) SetAlign(align Align) {
	c.align = align
}

// GetPosition returns the current position
func (c *TableCell) GetPosition() (x, y int) {
	return 0, 0
}

// SetPosition sets the position
func (c *TableCell) SetPosition(x, y int) {
}

// GetSize returns the current size
func (c *TableCell) GetSize() (width, height int) {
	if c.content == nil {
		return 0, 0
	}
	return c.content.GetSize()
}

// SetSize sets the size
func (c *TableCell) SetSize(width, height int) {
}

// GetWidth returns the width
func (c *TableCell) GetWidth() int {
	w, _ := c.GetSize()
	return w
}

// GetHeight returns the height
func (c *TableCell) GetHeight() int {
	_, h := c.GetSize()
	return h
}

// =============================================================================
// Helper Functions
// =============================================================================

// isTable checks if a node is a TableLayout
func isTable(node Node) bool {
	_, ok := node.(*TableLayout)
	return ok
}

// isTableRow checks if a node is a TableRow
func isTableRow(node Node) bool {
	_, ok := node.(*TableRow)
	return ok
}

// isTableCell checks if a node is a TableCell
func isTableCell(node Node) bool {
	_, ok := node.(*TableCell)
	return ok
}

// GetTableFromNode extracts table layout if node is a table
func GetTableFromNode(node Node) *TableLayout {
	if t, ok := node.(*TableLayout); ok {
		return t
	}
	return nil
}
