package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// TableLayout Tests
// =============================================================================

func TestNewTableLayout(t *testing.T) {
	rows := [][]Node{
		{NewMockNode("r0c0", 10, 5), NewMockNode("r0c1", 20, 5)},
		{NewMockNode("r1c0", 10, 3), NewMockNode("r1c1", 15, 3)},
	}
	table := NewTableLayout("test", rows)

	assert.Equal(t, "test", table.ID())
	assert.Equal(t, "table", table.Type())
	assert.Len(t, table.Rows(), 2)
}

func TestTableLayout_Children(t *testing.T) {
	rows := [][]Node{
		{NewMockNode("r0c0", 10, 5), NewMockNode("r0c1", 20, 5)},
		{NewMockNode("r1c0", 10, 3), NewMockNode("r1c1", 15, 3)},
	}
	table := NewTableLayout("test", rows)

	children := table.Children()
	assert.Len(t, children, 4)
}

func TestTableLayout_EmptyTable(t *testing.T) {
	table := NewTableLayout("empty", nil)

	w, h := table.GetSize()
	assert.Equal(t, 0, w)
	assert.Equal(t, 0, h)

	children := table.Children()
	assert.Nil(t, children)
}

func TestTableLayout_ColumnWidths(t *testing.T) {
	rows := [][]Node{
		{NewMockNode("r0c0", 10, 5), NewMockNode("r0c1", 20, 5)},
		{NewMockNode("r1c0", 15, 3), NewMockNode("r1c1", 12, 3)},
	}
	table := NewTableLayout("test", rows)

	// Column widths should be max of each column
	// Col 0: max(10, 15) = 15
	// Col 1: max(20, 12) = 20
	colWidths := table.ColumnWidths()
	assert.Len(t, colWidths, 2)
	assert.Equal(t, 15, colWidths[0])
	assert.Equal(t, 20, colWidths[1])
}

func TestTableLayout_RowHeights(t *testing.T) {
	rows := [][]Node{
		{NewMockNode("r0c0", 10, 5), NewMockNode("r0c1", 20, 7)},
		{NewMockNode("r1c0", 15, 3), NewMockNode("r1c1", 12, 4)},
	}
	table := NewTableLayout("test", rows)

	// Row heights should be max of each row
	// Row 0: max(5, 7) = 7
	// Row 1: max(3, 4) = 4
	rowHeights := table.RowHeights()
	assert.Len(t, rowHeights, 2)
	assert.Equal(t, 7, rowHeights[0])
	assert.Equal(t, 4, rowHeights[1])
}

func TestTableLayout_GetSize(t *testing.T) {
	rows := [][]Node{
		{NewMockNode("r0c0", 10, 5), NewMockNode("r0c1", 20, 5)},
		{NewMockNode("r1c0", 15, 3), NewMockNode("r1c1", 12, 3)},
	}
	table := NewTableLayout("test", rows)

	w, h := table.GetSize()
	// Width: 15 + 20 = 35
	// Height: 5 + 3 = 8 (using max heights from each row)
	assert.Equal(t, 35, w)
	assert.Equal(t, 8, h)
}

func TestTableLayout_CellPosition(t *testing.T) {
	rows := [][]Node{
		{NewMockNode("r0c0", 10, 5), NewMockNode("r0c1", 20, 5)},
		{NewMockNode("r1c0", 15, 3), NewMockNode("r1c1", 12, 3)},
	}
	table := NewTableLayout("test", rows)

	// Cell (0,0) should be at (0, 0)
	x, y := table.CellPosition(0, 0)
	assert.Equal(t, 0, x)
	assert.Equal(t, 0, y)

	// Cell (0,1) should be at (15, 0)
	x, y = table.CellPosition(0, 1)
	assert.Equal(t, 15, x)
	assert.Equal(t, 0, y)

	// Cell (1,0) should be at (0, 5)
	x, y = table.CellPosition(1, 0)
	assert.Equal(t, 0, x)
	assert.Equal(t, 5, y)

	// Cell (1,1) should be at (15, 5)
	x, y = table.CellPosition(1, 1)
	assert.Equal(t, 15, x)
	assert.Equal(t, 5, y)
}

func TestTableLayout_CellSize(t *testing.T) {
	rows := [][]Node{
		{NewMockNode("r0c0", 10, 5), NewMockNode("r0c1", 20, 5)},
		{NewMockNode("r1c0", 15, 3), NewMockNode("r1c1", 12, 3)},
	}
	table := NewTableLayout("test", rows)

	// Cell in column 0 should have width 15 (max of column)
	// Cell in row 0 should have height 5 (max of row)
	w, h := table.CellSize(0, 0)
	assert.Equal(t, 15, w)
	assert.Equal(t, 5, h)

	// Cell in column 1 should have width 20 (max of column)
	w, h = table.CellSize(0, 1)
	assert.Equal(t, 20, w)
	assert.Equal(t, 5, h)
}

func TestTableLayout_RaggedRows(t *testing.T) {
	// Table with different row lengths
	rows := [][]Node{
		{NewMockNode("r0c0", 10, 5), NewMockNode("r0c1", 20, 5), NewMockNode("r0c2", 5, 5)},
		{NewMockNode("r1c0", 15, 3)}, // Shorter row
	}
	table := NewTableLayout("test", rows)

	// Should handle ragged rows
	colWidths := table.ColumnWidths()
	assert.Len(t, colWidths, 3) // Max row length
	assert.Equal(t, 15, colWidths[0])
	assert.Equal(t, 20, colWidths[1])
	assert.Equal(t, 5, colWidths[2])
}

func TestTableLayout_Measure(t *testing.T) {
	rows := [][]Node{
		{NewMockNode("r0c0", 10, 5), NewMockNode("r0c1", 20, 5)},
	}
	table := NewTableLayout("test", rows)

	// Unbounded constraints
	constraints := UnboundedConstraints()
	size := table.Measure(constraints)
	assert.Equal(t, 30, size.Width)  // 10 + 20
	assert.Equal(t, 5, size.Height)

	// Tight constraints
	tightConstraints := TightConstraints(20, 10)
	size = table.Measure(tightConstraints)
	assert.Equal(t, 20, size.Width)
	assert.Equal(t, 10, size.Height)
}

// =============================================================================
// TableRow Tests
// =============================================================================

func TestNewTableRow(t *testing.T) {
	cells := []Node{NewMockNode("c0", 10, 5), NewMockNode("c1", 20, 5)}
	row := NewTableRow("row", cells)

	assert.Equal(t, "row", row.ID())
	assert.Equal(t, "tableRow", row.Type())
	assert.Len(t, row.Cells(), 2)
}

func TestTableRow_Children(t *testing.T) {
	cells := []Node{NewMockNode("c0", 10, 5), NewMockNode("c1", 20, 5)}
	row := NewTableRow("row", cells)

	children := row.Children()
	assert.Len(t, children, 2)
}

func TestTableRow_GetSize(t *testing.T) {
	cells := []Node{NewMockNode("c0", 10, 5), NewMockNode("c1", 20, 8)}
	row := NewTableRow("row", cells)

	w, h := row.GetSize()
	assert.Equal(t, 30, w) // 10 + 20
	assert.Equal(t, 8, h)  // max(5, 8)
}

// =============================================================================
// TableCell Tests
// =============================================================================

func TestNewTableCell(t *testing.T) {
	content := NewMockNode("content", 10, 5)
	cell := NewTableCell("cell", content)

	assert.Equal(t, "cell", cell.ID())
	assert.Equal(t, "tableCell", cell.Type())
	assert.Equal(t, content, cell.Content())
	assert.Equal(t, 1, cell.ColSpan())
	assert.Equal(t, 1, cell.RowSpan())
}

func TestNewTableCellWithSpan(t *testing.T) {
	content := NewMockNode("content", 10, 5)
	cell := NewTableCellWithSpan("cell", content, 2, 3)

	assert.Equal(t, 2, cell.ColSpan())
	assert.Equal(t, 3, cell.RowSpan())
}

func TestTableCell_Children(t *testing.T) {
	content := NewMockNode("content", 10, 5)
	cell := NewTableCell("cell", content)

	children := cell.Children()
	assert.Len(t, children, 1)
	assert.Equal(t, content, children[0])

	// Nil content
	nilCell := NewTableCell("nil", nil)
	assert.Nil(t, nilCell.Children())
}

func TestTableCell_GetSize(t *testing.T) {
	content := NewMockNode("content", 10, 5)
	cell := NewTableCell("cell", content)

	w, h := cell.GetSize()
	assert.Equal(t, 10, w)
	assert.Equal(t, 5, h)

	// Nil content
	nilCell := NewTableCell("nil", nil)
	w, h = nilCell.GetSize()
	assert.Equal(t, 0, w)
	assert.Equal(t, 0, h)
}

func TestTableCell_Align(t *testing.T) {
	content := NewMockNode("content", 10, 5)
	cell := NewTableCell("cell", content)

	assert.Equal(t, AlignLeft, cell.Align()) // Default

	cell.SetAlign(AlignCenter)
	assert.Equal(t, AlignCenter, cell.Align())

	cell.SetAlign(AlignRight)
	assert.Equal(t, AlignRight, cell.Align())
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestIsTable(t *testing.T) {
	table := NewTableLayout("test", nil)
	regular := NewMockNode("regular", 10, 5)

	assert.True(t, isTable(table))
	assert.False(t, isTable(regular))
}

func TestIsTableRow(t *testing.T) {
	row := NewTableRow("row", nil)
	regular := NewMockNode("regular", 10, 5)

	assert.True(t, isTableRow(row))
	assert.False(t, isTableRow(regular))
}

func TestIsTableCell(t *testing.T) {
	cell := NewTableCell("cell", nil)
	regular := NewMockNode("regular", 10, 5)

	assert.True(t, isTableCell(cell))
	assert.False(t, isTableCell(regular))
}

func TestGetTableFromNode(t *testing.T) {
	table := NewTableLayout("test", nil)
	regular := NewMockNode("regular", 10, 5)

	result := GetTableFromNode(table)
	assert.NotNil(t, result)

	result = GetTableFromNode(regular)
	assert.Nil(t, result)
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestTableLayout_NilCells(t *testing.T) {
	rows := [][]Node{
		{nil, NewMockNode("r0c1", 20, 5)},
		{NewMockNode("r1c0", 15, 3), nil},
	}
	table := NewTableLayout("test", rows)

	// Should handle nil cells gracefully
	colWidths := table.ColumnWidths()
	assert.Len(t, colWidths, 2)
	assert.Equal(t, 15, colWidths[0])
	assert.Equal(t, 20, colWidths[1])
}

func TestTableLayout_SingleCell(t *testing.T) {
	rows := [][]Node{
		{NewMockNode("r0c0", 10, 5)},
	}
	table := NewTableLayout("test", rows)

	w, h := table.GetSize()
	assert.Equal(t, 10, w)
	assert.Equal(t, 5, h)
}

func TestTableLayout_LargeTable(t *testing.T) {
	// Create a 10x10 table
	rows := make([][]Node, 10)
	for i := range rows {
		rows[i] = make([]Node, 10)
		for j := range rows[i] {
			rows[i][j] = NewMockNode("cell", 5, 2)
		}
	}
	table := NewTableLayout("large", rows)

	w, h := table.GetSize()
	assert.Equal(t, 50, w) // 10 * 5
	assert.Equal(t, 20, h) // 10 * 2
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkTableLayout_GetSize(b *testing.B) {
	rows := [][]Node{
		{NewMockNode("r0c0", 10, 5), NewMockNode("r0c1", 20, 5)},
		{NewMockNode("r1c0", 15, 3), NewMockNode("r1c1", 12, 3)},
	}
	table := NewTableLayout("test", rows)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		table.GetSize()
	}
}

func BenchmarkTableLayout_CellPosition(b *testing.B) {
	rows := [][]Node{
		{NewMockNode("r0c0", 10, 5), NewMockNode("r0c1", 20, 5)},
		{NewMockNode("r1c0", 15, 3), NewMockNode("r1c1", 12, 3)},
	}
	table := NewTableLayout("test", rows)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		table.CellPosition(1, 1)
	}
}

func BenchmarkTableLayout_Measure(b *testing.B) {
	rows := [][]Node{
		{NewMockNode("r0c0", 10, 5), NewMockNode("r0c1", 20, 5)},
		{NewMockNode("r1c0", 15, 3), NewMockNode("r1c1", 12, 3)},
	}
	table := NewTableLayout("test", rows)
	constraints := UnboundedConstraints()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		table.Measure(constraints)
	}
}
