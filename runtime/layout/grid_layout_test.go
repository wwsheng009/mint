package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Mock GridStyleProvider Node
// =============================================================================

// MockGridNode is a mock node that implements GridStyleProvider
type MockGridNode struct {
	*MockNode
	gridStyle *GridStyle
	border    Border
}

// NewMockGridNode creates a mock grid container
func NewMockGridNode(id string, width, height int) *MockGridNode {
	return &MockGridNode{
		MockNode:  NewMockNode(id, width, height),
		gridStyle: DefaultGridStyle(),
		border:    NewBorder(BorderNone),
	}
}

// GetGridStyle implements GridStyleProvider
func (m *MockGridNode) GetGridStyle() *GridStyle {
	return m.gridStyle
}

// SetGridColumns sets column definitions
func (m *MockGridNode) SetGridColumns(cols []GridDimension) {
	m.gridStyle.Columns = cols
}

// SetGridRows sets row definitions
func (m *MockGridNode) SetGridRows(rows []GridDimension) {
	m.gridStyle.Rows = rows
}

// SetGridCells sets grid cells
func (m *MockGridNode) SetGridCells(cells []GridCell) {
	m.gridStyle.Cells = cells
}

// SetGridChildren sets children (alternative to cells)
func (m *MockGridNode) SetGridChildren(children []Node) {
	m.children = children
}

// SetGridGap sets column and row gaps
func (m *MockGridNode) SetGridGap(colGap, rowGap int) {
	m.gridStyle.ColumnGap = colGap
	m.gridStyle.RowGap = rowGap
}

// GetBorder implements Bordered interface
func (m *MockGridNode) GetBorder() Border {
	return m.border
}

// SetBorder sets the border style
func (m *MockGridNode) SetBorder(style BorderStyle) {
	m.border = NewBorder(style)
}

// =============================================================================
// GridDimension Tests
// =============================================================================

func TestGridFixed(t *testing.T) {
	fixed := GridFixed(100)
	assert.NotNil(t, fixed)
}

func TestGridFlex(t *testing.T) {
	flex := GridFlex{Factor: 1}
	assert.Equal(t, 1, flex.Factor)
}

func TestGridAuto(t *testing.T) {
	auto := GridAuto{}
	assert.NotNil(t, auto)
}

func TestGridMin(t *testing.T) {
	min := GridMin{Min: 50, Content: GridAuto{}}
	assert.Equal(t, 50, min.Min)
}

func TestGridMax(t *testing.T) {
	max := GridMax{Max: 200, Content: GridFlex{Factor: 1}}
	assert.Equal(t, 200, max.Max)
}

// =============================================================================
// GridStyle Tests
// =============================================================================

func TestDefaultGridStyle(t *testing.T) {
	style := DefaultGridStyle()
	assert.NotNil(t, style)
	assert.Len(t, style.Columns, 1)
	assert.Len(t, style.Rows, 1)
	assert.Equal(t, 0, style.ColumnGap)
	assert.Equal(t, 0, style.RowGap)
}

func TestGridStyle_WithFixedColumns(t *testing.T) {
	style := &GridStyle{
		Columns: []GridDimension{
			GridFixed(20),
			GridFixed(30),
			GridFixed(20),
		},
		Rows: []GridDimension{
			GridAuto{},
		},
	}

	assert.Len(t, style.Columns, 3)
	assert.Len(t, style.Rows, 1)
}

func TestGridStyle_WithFlexColumns(t *testing.T) {
	style := &GridStyle{
		Columns: []GridDimension{
			GridFlex{Factor: 1},
			GridFlex{Factor: 2},
			GridFlex{Factor: 1},
		},
		Rows: []GridDimension{
			GridAuto{},
		},
	}

	assert.Len(t, style.Columns, 3)
}

// =============================================================================
// Engine.Layout with GridStyleProvider Tests
// =============================================================================

func TestEngine_GridLayout_SingleCell(t *testing.T) {
	// Grid with single cell
	grid := NewMockGridNode("grid", 60, 10)
	grid.SetGridColumns([]GridDimension{GridFixed(60)})
	grid.SetGridRows([]GridDimension{GridFixed(10)})

	child := NewMockMeasurableNode("child", 60, 10)
	grid.SetGridCells([]GridCell{
		{Child: child, Row: 0, Col: 0},
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.NotNil(t, result.Root)
	assert.Len(t, result.Root.Children, 1)
}

func TestEngine_GridLayout_TwoColumns(t *testing.T) {
	// Grid with 2 columns: 30px each
	grid := NewMockGridNode("grid", 60, 10)
	grid.SetGridColumns([]GridDimension{
		GridFixed(30),
		GridFixed(30),
	})
	grid.SetGridRows([]GridDimension{GridFixed(10)})

	child1 := NewMockMeasurableNode("child1", 30, 10)
	child2 := NewMockMeasurableNode("child2", 30, 10)
	grid.SetGridCells([]GridCell{
		{Child: child1, Row: 0, Col: 0},
		{Child: child2, Row: 0, Col: 1},
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 2)

	// Children should be laid out in columns
	box1 := result.Root.Children[0]
	box2 := result.Root.Children[1]

	assert.Equal(t, 0, box1.X, "First child should start at X=0")
	assert.GreaterOrEqual(t, box2.X, box1.X+box1.Width, "Second child should be to the right of first")
}

func TestEngine_GridLayout_TwoRows(t *testing.T) {
	// Grid with 2 rows
	grid := NewMockGridNode("grid", 60, 20)
	grid.SetGridColumns([]GridDimension{GridFixed(60)})
	grid.SetGridRows([]GridDimension{
		GridFixed(10),
		GridFixed(10),
	})

	child1 := NewMockMeasurableNode("child1", 60, 10)
	child2 := NewMockMeasurableNode("child2", 60, 10)
	grid.SetGridCells([]GridCell{
		{Child: child1, Row: 0, Col: 0},
		{Child: child2, Row: 1, Col: 0},
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 2)

	// Children should be laid out in rows
	box1 := result.Root.Children[0]
	box2 := result.Root.Children[1]

	assert.Equal(t, 0, box1.Y, "First row should start at Y=0")
	assert.GreaterOrEqual(t, box2.Y, box1.Y+box1.Height, "Second row should be below first")
}

func TestEngine_GridLayout_2x2(t *testing.T) {
	// 2x2 grid
	grid := NewMockGridNode("grid", 60, 20)
	grid.SetGridColumns([]GridDimension{
		GridFixed(30),
		GridFixed(30),
	})
	grid.SetGridRows([]GridDimension{
		GridFixed(10),
		GridFixed(10),
	})

	child00 := NewMockMeasurableNode("child00", 30, 10)
	child01 := NewMockMeasurableNode("child01", 30, 10)
	child10 := NewMockMeasurableNode("child10", 30, 10)
	child11 := NewMockMeasurableNode("child11", 30, 10)

	grid.SetGridCells([]GridCell{
		{Child: child00, Row: 0, Col: 0},
		{Child: child01, Row: 0, Col: 1},
		{Child: child10, Row: 1, Col: 0},
		{Child: child11, Row: 1, Col: 1},
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 4)

	// Verify 2x2 layout positions
	// Row 0: (0,0) and (30,0)
	// Row 1: (0,10) and (30,10)
	boxes := result.Root.Children
	assert.Equal(t, 0, boxes[0].X)
	assert.Equal(t, 0, boxes[0].Y)
	assert.Equal(t, 30, boxes[1].X)
	assert.Equal(t, 0, boxes[1].Y)
	assert.Equal(t, 0, boxes[2].X)
	assert.Equal(t, 10, boxes[2].Y)
	assert.Equal(t, 30, boxes[3].X)
	assert.Equal(t, 10, boxes[3].Y)
}

func TestEngine_GridLayout_WithGap(t *testing.T) {
	// Grid with gap between cells
	grid := NewMockGridNode("grid", 62, 20)
	grid.SetGridColumns([]GridDimension{
		GridFixed(30),
		GridFixed(30),
	})
	grid.SetGridRows([]GridDimension{GridFixed(10)})
	grid.SetGridGap(2, 0) // 2px column gap

	child1 := NewMockMeasurableNode("child1", 30, 10)
	child2 := NewMockMeasurableNode("child2", 30, 10)
	grid.SetGridCells([]GridCell{
		{Child: child1, Row: 0, Col: 0},
		{Child: child2, Row: 0, Col: 1},
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 2)

	// With 2px gap, second child should be at X=32 (30+2)
	if len(result.Root.Children) >= 2 {
		assert.Equal(t, 0, result.Root.Children[0].X)
		assert.Equal(t, 32, result.Root.Children[1].X, "Second child should be at X=32 (30 + 2 gap)")
	}
}

func TestEngine_GridLayout_FlexColumns(t *testing.T) {
	// Grid with flex columns (1:2:1 ratio)
	grid := NewMockGridNode("grid", 100, 10)
	grid.SetGridColumns([]GridDimension{
		GridFlex{Factor: 1},
		GridFlex{Factor: 2},
		GridFlex{Factor: 1},
	})
	grid.SetGridRows([]GridDimension{GridFixed(10)})

	child1 := NewMockMeasurableNode("child1", 25, 10)
	child2 := NewMockMeasurableNode("child2", 50, 10)
	child3 := NewMockMeasurableNode("child3", 25, 10)

	grid.SetGridCells([]GridCell{
		{Child: child1, Row: 0, Col: 0},
		{Child: child2, Row: 0, Col: 1},
		{Child: child3, Row: 0, Col: 2},
	})

	engine := NewEngine()
	result := engine.Layout(grid, NewConstraints(100, 100, 10, 10))

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 3)

	// Verify flex distribution: 25, 50, 25
	assert.Equal(t, 0, result.Root.Children[0].X)
	assert.Equal(t, 25, result.Root.Children[1].X)
	assert.Equal(t, 75, result.Root.Children[2].X)
}

func TestEngine_GridLayout_AutoRows(t *testing.T) {
	// Grid with auto-sized rows (based on content)
	grid := NewMockGridNode("grid", 60, 30)
	grid.SetGridColumns([]GridDimension{GridFixed(60)})
	grid.SetGridRows([]GridDimension{
		GridAuto{},
		GridAuto{},
	})

	child1 := NewMockMeasurableNode("child1", 60, 5)
	child2 := NewMockMeasurableNode("child2", 60, 8)

	grid.SetGridCells([]GridCell{
		{Child: child1, Row: 0, Col: 0},
		{Child: child2, Row: 1, Col: 0},
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 2)
}

func TestEngine_GridLayout_MixedDimensions(t *testing.T) {
	// Grid with mixed column dimensions
	grid := NewMockGridNode("grid", 100, 10)
	grid.SetGridColumns([]GridDimension{
		GridFixed(20),
		GridFlex{Factor: 1},
		GridFixed(20),
	})
	grid.SetGridRows([]GridDimension{GridFixed(10)})

	child1 := NewMockMeasurableNode("child1", 20, 10)
	child2 := NewMockMeasurableNode("child2", 60, 10)
	child3 := NewMockMeasurableNode("child3", 20, 10)

	grid.SetGridCells([]GridCell{
		{Child: child1, Row: 0, Col: 0},
		{Child: child2, Row: 0, Col: 1},
		{Child: child3, Row: 0, Col: 2},
	})

	engine := NewEngine()
	result := engine.Layout(grid, NewConstraints(100, 100, 10, 10))

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 3)
}

func TestEngine_GridLayout_CellSpan(t *testing.T) {
	// Grid with col span
	grid := NewMockGridNode("grid", 60, 20)
	grid.SetGridColumns([]GridDimension{
		GridFixed(20),
		GridFixed(20),
		GridFixed(20),
	})
	grid.SetGridRows([]GridDimension{
		GridFixed(10),
		GridFixed(10),
	})

	// Child spanning 2 columns
	child1 := NewMockMeasurableNode("child1", 40, 10)
	grid.SetGridCells([]GridCell{
		{Child: child1, Row: 0, Col: 0, ColSpan: 2},
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestEngine_GridLayout_Empty(t *testing.T) {
	grid := NewMockGridNode("grid", 60, 10)
	grid.SetGridCells(nil)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 0)
}

func TestEngine_GridLayout_NoDimensions(t *testing.T) {
	grid := NewMockGridNode("grid", 60, 10)
	grid.SetGridColumns(nil)
	grid.SetGridRows(nil)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestEngine_GridLayout_ZeroSizeChild(t *testing.T) {
	grid := NewMockGridNode("grid", 60, 10)
	grid.SetGridColumns([]GridDimension{GridFixed(30)})
	grid.SetGridRows([]GridDimension{GridFixed(10)})

	child := NewMockMeasurableNode("child", 0, 0)
	grid.SetGridCells([]GridCell{
		{Child: child, Row: 0, Col: 0},
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)
}

func TestEngine_GridLayout_OutOfBounds(t *testing.T) {
	// Child with row/col beyond defined dimensions
	grid := NewMockGridNode("grid", 60, 10)
	grid.SetGridColumns([]GridDimension{GridFixed(30)})
	grid.SetGridRows([]GridDimension{GridFixed(10)})

	child := NewMockMeasurableNode("child", 30, 10)
	grid.SetGridCells([]GridCell{
		{Child: child, Row: 5, Col: 5}, // Out of bounds
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	// Should handle gracefully
	assert.NotNil(t, result)
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestEngine_GridLayout_NestedInVStack(t *testing.T) {
	// VStack containing grid
	vstack := NewMockFlexNode("vstack", FlexColumn)
	vstack.SetSize(60, 30)

	grid := NewMockGridNode("grid", 60, 20)
	grid.SetGridColumns([]GridDimension{GridFixed(30), GridFixed(30)})
	grid.SetGridRows([]GridDimension{GridFixed(10), GridFixed(10)})

	child := NewMockMeasurableNode("child", 30, 10)
	grid.SetGridCells([]GridCell{{Child: child, Row: 0, Col: 0}})

	vstack.SetChildren([]Node{grid})

	engine := NewEngine()
	result := engine.Layout(vstack, NewConstraints(60, 60, 0, 100))

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)
}

func TestEngine_GridLayout_WithAbsoluteChild(t *testing.T) {
	// Grid containing absolute positioned child
	grid := NewMockGridNode("grid", 60, 20)
	grid.SetGridColumns([]GridDimension{GridFixed(60)})
	grid.SetGridRows([]GridDimension{GridFixed(20)})

	// Absolute container inside grid cell
	abs := NewMockAbsoluteNode("abs", 60, 20)
	absChild := NewMockMeasurableNode("absChild", 10, 1)
	abs.SetChildren([]Node{absChild})
	abs.SetPositionStyle(AbsolutePos(0), AbsolutePos(0), nil, nil, AnchorTopLeft)

	grid.SetGridCells([]GridCell{
		{Child: abs, Row: 0, Col: 0},
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)

	// Grid child should be at origin
	assert.Equal(t, 0, result.Root.Children[0].X)
	assert.Equal(t, 0, result.Root.Children[0].Y)
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkEngine_GridLayout_2x2(b *testing.B) {
	grid := NewMockGridNode("grid", 60, 20)
	grid.SetGridColumns([]GridDimension{GridFixed(30), GridFixed(30)})
	grid.SetGridRows([]GridDimension{GridFixed(10), GridFixed(10)})

	grid.SetGridCells([]GridCell{
		{Child: NewMockMeasurableNode("c00", 30, 10), Row: 0, Col: 0},
		{Child: NewMockMeasurableNode("c01", 30, 10), Row: 0, Col: 1},
		{Child: NewMockMeasurableNode("c10", 30, 10), Row: 1, Col: 0},
		{Child: NewMockMeasurableNode("c11", 30, 10), Row: 1, Col: 1},
	})

	engine := NewEngine()
	constraints := UnboundedConstraints()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Layout(grid, constraints)
	}
}

func BenchmarkEngine_GridLayout_Flex(b *testing.B) {
	grid := NewMockGridNode("grid", 100, 10)
	grid.SetGridColumns([]GridDimension{
		GridFlex{Factor: 1},
		GridFlex{Factor: 2},
		GridFlex{Factor: 1},
	})
	grid.SetGridRows([]GridDimension{GridFixed(10)})

	grid.SetGridCells([]GridCell{
		{Child: NewMockMeasurableNode("c0", 25, 10), Row: 0, Col: 0},
		{Child: NewMockMeasurableNode("c1", 50, 10), Row: 0, Col: 1},
		{Child: NewMockMeasurableNode("c2", 25, 10), Row: 0, Col: 2},
	})

	engine := NewEngine()
	constraints := NewConstraints(100, 100, 10, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Layout(grid, constraints)
	}
}
