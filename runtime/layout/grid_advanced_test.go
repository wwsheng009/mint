package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Grid 高级功能测试 - RowSpan 和 ColSpan
// =============================================================================

func TestGrid_RowSpan_Basic(t *testing.T) {
	// 单元格跨越2行
	grid := NewMockGridNode("grid", 100, 100)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})

	// 第一个单元格跨越2行
	cell1 := NewMockMeasurableNode("cell1", 50, 100)
	cell2 := NewMockMeasurableNode("cell2", 50, 50)

	grid.SetGridCells([]GridCell{
		{Child: cell1, Row: 0, Col: 0, RowSpan: 2, ColSpan: 1},
		{Child: cell2, Row: 0, Col: 1, RowSpan: 1, ColSpan: 1},
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 2)
}

func TestGrid_ColSpan_Basic(t *testing.T) {
	// 单元格跨越2列
	grid := NewMockGridNode("grid", 100, 100)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})

	// 第一个单元格跨越2列
	cell1 := NewMockMeasurableNode("cell1", 100, 50)
	cell2 := NewMockMeasurableNode("cell2", 50, 50)

	grid.SetGridCells([]GridCell{
		{Child: cell1, Row: 0, Col: 0, RowSpan: 1, ColSpan: 2},
		{Child: cell2, Row: 1, Col: 0, RowSpan: 1, ColSpan: 1},
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 2)
}

func TestGrid_RowSpan_ColSpan_Combined(t *testing.T) {
	// 单元格同时跨越多行多列
	grid := NewMockGridNode("grid", 100, 100)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})

	// 第一个单元格跨越2行2列
	cell1 := NewMockMeasurableNode("cell1", 100, 100)

	grid.SetGridCells([]GridCell{
		{Child: cell1, Row: 0, Col: 0, RowSpan: 2, ColSpan: 2},
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 1)
}

func TestGrid_Span_Complex(t *testing.T) {
	// 复杂的 span 组合
	// 3x3 网格
	grid := NewMockGridNode("grid", 300, 300)
	grid.SetGridColumns([]GridDimension{GridFixed(100), GridFixed(100), GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(100), GridFixed(100), GridFixed(100)})

	cells := []GridCell{
		{Child: NewMockMeasurableNode("header", 300, 100), Row: 0, Col: 0, RowSpan: 1, ColSpan: 3}, // 顶行跨越3列
		{Child: NewMockMeasurableNode("sidebar", 100, 200), Row: 1, Col: 0, RowSpan: 2, ColSpan: 1}, // 左侧跨越2行
		{Child: NewMockMeasurableNode("main", 200, 200), Row: 1, Col: 1, RowSpan: 2, ColSpan: 2},    // 主区域跨越2行2列
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 3)
}

func TestGrid_Span_ExceedsBounds(t *testing.T) {
	// Span 超出网格边界
	grid := NewMockGridNode("grid", 100, 100)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50)})

	// RowSpan 超出行数
	cell := NewMockMeasurableNode("cell", 50, 100)
	grid.SetGridCells([]GridCell{
		{Child: cell, Row: 0, Col: 0, RowSpan: 5, ColSpan: 1}, // 只有1行但跨越5行
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	// 应该优雅处理
	assert.NotNil(t, result)
}

func TestGrid_Span_Zero(t *testing.T) {
	// RowSpan/ColSpan 为 0 (应该视为1)
	grid := NewMockGridNode("grid", 100, 100)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50)})

	cell := NewMockMeasurableNode("cell", 50, 50)
	grid.SetGridCells([]GridCell{
		{Child: cell, Row: 0, Col: 0, RowSpan: 0, ColSpan: 0},
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestGrid_Span_Negative(t *testing.T) {
	// 负数的 RowSpan/ColSpan
	grid := NewMockGridNode("grid", 100, 100)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50)})

	cell := NewMockMeasurableNode("cell", 50, 50)
	grid.SetGridCells([]GridCell{
		{Child: cell, Row: 0, Col: 0, RowSpan: -1, ColSpan: -1},
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
}

// =============================================================================
// Grid Flex 尺寸测试
// =============================================================================

func TestGrid_FlexColumn_EqualDistribution(t *testing.T) {
	// 等比例 flex 分布
	grid := NewMockGridNode("grid", 100, 50)
	grid.SetGridColumns([]GridDimension{
		GridFlex{Factor: 1},
		GridFlex{Factor: 1},
		GridFlex{Factor: 1},
	})
	grid.SetGridRows([]GridDimension{GridFixed(50)})

	cells := make([]GridCell, 3)
	for i := range cells {
		cells[i] = GridCell{
			Child: NewMockMeasurableNode("cell", 0, 50),
			Row:   0,
			Col:   i,
		}
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, NewConstraints(100, 100, 50, 50))

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 3)
}

func TestGrid_FlexColumn_UnequalDistribution(t *testing.T) {
	// 不等比例 flex 分布 (1:2:1)
	grid := NewMockGridNode("grid", 100, 50)
	grid.SetGridColumns([]GridDimension{
		GridFlex{Factor: 1},
		GridFlex{Factor: 2},
		GridFlex{Factor: 1},
	})
	grid.SetGridRows([]GridDimension{GridFixed(50)})

	cells := []GridCell{
		{Child: NewMockMeasurableNode("c1", 0, 50), Row: 0, Col: 0}, // 应该是25宽
		{Child: NewMockMeasurableNode("c2", 0, 50), Row: 0, Col: 1}, // 应该是50宽
		{Child: NewMockMeasurableNode("c3", 0, 50), Row: 0, Col: 2}, // 应该是25宽
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, NewConstraints(100, 100, 50, 50))

	assert.NotNil(t, result)
}

func TestGrid_FlexRow_EqualDistribution(t *testing.T) {
	// 行等比例 flex 分布
	grid := NewMockGridNode("grid", 100, 100)
	grid.SetGridColumns([]GridDimension{GridFixed(100)})
	grid.SetGridRows([]GridDimension{
		GridFlex{Factor: 1},
		GridFlex{Factor: 1},
	})

	cells := []GridCell{
		{Child: NewMockMeasurableNode("r1", 100, 0), Row: 0, Col: 0},
		{Child: NewMockMeasurableNode("r2", 100, 0), Row: 1, Col: 0},
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, NewConstraints(100, 100, 100, 100))

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 2)
}

func TestGrid_MixedFixedFlex(t *testing.T) {
	// 混合 Fixed 和 Flex
	grid := NewMockGridNode("grid", 200, 100)
	grid.SetGridColumns([]GridDimension{
		GridFixed(50),       // 固定50
		GridFlex{Factor: 1}, // flex 剩余空间
		GridFixed(30),       // 固定30
	})
	grid.SetGridRows([]GridDimension{GridFixed(100)})

	cells := []GridCell{
		{Child: NewMockMeasurableNode("c1", 50, 100), Row: 0, Col: 0},
		{Child: NewMockMeasurableNode("c2", 120, 100), Row: 0, Col: 1}, // 200-50-30=120
		{Child: NewMockMeasurableNode("c3", 30, 100), Row: 0, Col: 2},
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, NewConstraints(200, 200, 100, 100))

	assert.NotNil(t, result)
}

func TestGrid_FlexWithGaps(t *testing.T) {
	// Flex 列 + 间隙
	grid := NewMockGridNode("grid", 100, 100)
	grid.SetGridColumns([]GridDimension{
		GridFlex{Factor: 1},
		GridFlex{Factor: 1},
	})
	grid.SetGridRows([]GridDimension{GridFixed(100)})
	grid.SetGridGap(5, 0) // 列间隙5px

	cells := []GridCell{
		{Child: NewMockMeasurableNode("c1", 0, 100), Row: 0, Col: 0},
		{Child: NewMockMeasurableNode("c2", 0, 100), Row: 0, Col: 1},
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, NewConstraints(100, 100, 100, 100))

	assert.NotNil(t, result)
	// 总宽度100，间隙5，两列应该是 (100-5)/2 = 47.5
}

func TestGrid_AllFlex(t *testing.T) {
	// 全部是 Flex，无固定尺寸
	grid := NewMockGridNode("grid", 0, 0)
	grid.SetGridColumns([]GridDimension{
		GridFlex{Factor: 1},
		GridFlex{Factor: 1},
		GridFlex{Factor: 1},
	})
	grid.SetGridRows([]GridDimension{
		GridFlex{Factor: 1},
		GridFlex{Factor: 1},
	})

	cells := make([]GridCell, 0, 6)
	for r := 0; r < 2; r++ {
		for c := 0; c < 3; c++ {
			cells = append(cells, GridCell{
				Child: NewMockMeasurableNode("cell", 0, 0),
				Row:   r,
				Col:   c,
			})
		}
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, NewConstraints(300, 300, 200, 200))

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 6)
}

// =============================================================================
// Grid 间隙测试
// =============================================================================

func TestGrid_Gap_Equal(t *testing.T) {
	// 等间隙
	grid := NewMockGridNode("grid", 110, 110)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridGap(10, 10) // 行间隙10，列间隙10

	cells := make([]GridCell, 0, 4)
	for r := 0; r < 2; r++ {
		for c := 0; c < 2; c++ {
			cells = append(cells, GridCell{
				Child: NewMockMeasurableNode("cell", 50, 50),
				Row:   r,
				Col:   c,
			})
		}
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestGrid_Gap_Asymmetric(t *testing.T) {
	// 非对称间隙
	grid := NewMockGridNode("grid", 120, 105)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridGap(5, 20) // 行间隙5，列间隙20

	cells := make([]GridCell, 0, 4)
	for r := 0; r < 2; r++ {
		for c := 0; c < 2; c++ {
			cells = append(cells, GridCell{
				Child: NewMockMeasurableNode("cell", 50, 50),
				Row:   r,
				Col:   c,
			})
		}
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestGrid_Gap_Zero(t *testing.T) {
	// 零间隙
	grid := NewMockGridNode("grid", 100, 100)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridGap(0, 0)

	cells := make([]GridCell, 0, 4)
	for r := 0; r < 2; r++ {
		for c := 0; c < 2; c++ {
			cells = append(cells, GridCell{
				Child: NewMockMeasurableNode("cell", 50, 50),
				Row:   r,
				Col:   c,
			})
		}
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestGrid_Gap_Large(t *testing.T) {
	// 大间隙
	grid := NewMockGridNode("grid", 200, 200)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridGap(100, 100) // 间隙100

	cells := make([]GridCell, 0, 4)
	for r := 0; r < 2; r++ {
		for c := 0; c < 2; c++ {
			cells = append(cells, GridCell{
				Child: NewMockMeasurableNode("cell", 50, 50),
				Row:   r,
				Col:   c,
			})
		}
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestGrid_Gap_ExceedsSize(t *testing.T) {
	// 间隙超过网格尺寸
	grid := NewMockGridNode("grid", 50, 50)
	grid.SetGridColumns([]GridDimension{GridFixed(20), GridFixed(20)})
	grid.SetGridRows([]GridDimension{GridFixed(20), GridFixed(20)})
	grid.SetGridGap(100, 100) // 间隙100 > 网格50

	cells := make([]GridCell, 0, 4)
	for r := 0; r < 2; r++ {
		for c := 0; c < 2; c++ {
			cells = append(cells, GridCell{
				Child: NewMockMeasurableNode("cell", 20, 20),
				Row:   r,
				Col:   c,
			})
		}
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	// 应该优雅处理
	assert.NotNil(t, result)
}

// =============================================================================
// Grid + Border 组合测试
// =============================================================================

func TestGrid_WithBorder(t *testing.T) {
	// Grid 容器有 Border
	grid := NewMockGridNode("grid", 102, 52)
	grid.SetBorder(BorderSingle)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50)})

	cells := []GridCell{
		{Child: NewMockMeasurableNode("c1", 50, 50), Row: 0, Col: 0},
		{Child: NewMockMeasurableNode("c2", 50, 50), Row: 0, Col: 1},
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
}

func TestGrid_ChildrenWithBorder(t *testing.T) {
	// Grid 子单元格有 Border
	grid := NewMockGridNode("grid", 104, 104)
	grid.SetGridColumns([]GridDimension{GridFixed(52), GridFixed(52)})
	grid.SetGridRows([]GridDimension{GridFixed(52), GridFixed(52)})

	cells := make([]GridCell, 0, 4)
	for r := 0; r < 2; r++ {
		for c := 0; c < 2; c++ {
			border := NewMockCompositeNode("border", 52, 52)
			border.SetBorder(BorderSingle)

			content := NewMockMeasurableNode("content", 50, 50)
			border.SetChildren([]Node{content})

			cells = append(cells, GridCell{
				Child: border,
				Row:   r,
				Col:   c,
			})
		}
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 4)
}

func TestGrid_BorderAndBorder(t *testing.T) {
	// Grid 容器和子单元格都有 Border
	grid := NewMockGridNode("grid", 106, 106)
	grid.SetBorder(BorderSingle)
	grid.SetGridColumns([]GridDimension{GridFixed(52), GridFixed(52)})
	grid.SetGridRows([]GridDimension{GridFixed(52), GridFixed(52)})

	cells := make([]GridCell, 0, 4)
	for r := 0; r < 2; r++ {
		for c := 0; c < 2; c++ {
			border := NewMockCompositeNode("border", 52, 52)
			border.SetBorder(BorderSingle)

			content := NewMockMeasurableNode("content", 50, 50)
			border.SetChildren([]Node{content})

			cells = append(cells, GridCell{
				Child: border,
				Row:   r,
				Col:   c,
			})
		}
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
}

// =============================================================================
// Grid 对齐测试
// =============================================================================

func TestGrid_Alignment_Default(t *testing.T) {
	// 默认对齐（左上）
	grid := NewMockGridNode("grid", 100, 100)
	grid.SetGridColumns([]GridDimension{GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(100)})

	// 子节点小于单元格
	cell := NewMockMeasurableNode("cell", 50, 50)
	grid.SetGridCells([]GridCell{{Child: cell, Row: 0, Col: 0}})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
}

// =============================================================================
// Grid 空单元格测试
// =============================================================================

func TestGrid_EmptyCells(t *testing.T) {
	// 部分单元格为空
	grid := NewMockGridNode("grid", 100, 100)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})

	// 只填充部分单元格
	cells := []GridCell{
		{Child: NewMockMeasurableNode("c1", 50, 50), Row: 0, Col: 0},
		// (0, 1) 空
		// (1, 0) 空
		{Child: NewMockMeasurableNode("c4", 50, 50), Row: 1, Col: 1},
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 2)
}

func TestGrid_AllEmptyCells(t *testing.T) {
	// 所有单元格都为空
	grid := NewMockGridNode("grid", 100, 100)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})

	grid.SetGridCells(nil)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	assert.NotNil(t, result)
	assert.Len(t, result.Root.Children, 0)
}

// =============================================================================
// 基准测试
// =============================================================================

func BenchmarkGrid_RowSpan_4x4(b *testing.B) {
	grid := NewMockGridNode("grid", 400, 400)
	grid.SetGridColumns([]GridDimension{
		GridFixed(100), GridFixed(100), GridFixed(100), GridFixed(100),
	})
	grid.SetGridRows([]GridDimension{
		GridFixed(100), GridFixed(100), GridFixed(100), GridFixed(100),
	})

	cells := []GridCell{
		{Child: NewMockMeasurableNode("span", 400, 100), Row: 0, Col: 0, RowSpan: 1, ColSpan: 4},
		{Child: NewMockMeasurableNode("c1", 100, 300), Row: 1, Col: 0, RowSpan: 3, ColSpan: 1},
		{Child: NewMockMeasurableNode("c2", 300, 300), Row: 1, Col: 1, RowSpan: 3, ColSpan: 3},
	}
	grid.SetGridCells(cells)

	engine := NewEngine()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Layout(grid, UnboundedConstraints())
	}
}

func BenchmarkGrid_Flex_10Columns(b *testing.B) {
	grid := NewMockGridNode("grid", 1000, 50)

	cols := make([]GridDimension, 10)
	for i := range cols {
		cols[i] = GridFlex{Factor: 1}
	}
	grid.SetGridColumns(cols)
	grid.SetGridRows([]GridDimension{GridFixed(50)})

	cells := make([]GridCell, 10)
	for i := range cells {
		cells[i] = GridCell{
			Child: NewMockMeasurableNode("c", 0, 50),
			Row:   0,
			Col:   i,
		}
	}
	grid.SetGridCells(cells)

	engine := NewEngine()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Layout(grid, NewConstraints(1000, 1000, 50, 50))
	}
}

func BenchmarkGrid_WithGaps_5x5(b *testing.B) {
	grid := NewMockGridNode("grid", 500, 500)
	grid.SetGridGap(5, 5)

	cols := make([]GridDimension, 5)
	rows := make([]GridDimension, 5)
	for i := range cols {
		cols[i] = GridFixed(100)
		rows[i] = GridFixed(100)
	}
	grid.SetGridColumns(cols)
	grid.SetGridRows(rows)

	cells := make([]GridCell, 0, 25)
	for r := 0; r < 5; r++ {
		for c := 0; c < 5; c++ {
			cells = append(cells, GridCell{
				Child: NewMockMeasurableNode("c", 100, 100),
				Row:   r,
				Col:   c,
			})
		}
	}
	grid.SetGridCells(cells)

	engine := NewEngine()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Layout(grid, UnboundedConstraints())
	}
}
