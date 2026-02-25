// ==============================================================================
// Grid 布局引擎 - 全面单元测试
// ==============================================================================
// 测试覆盖范围：
// 1. 基本功能测试（Fixed, Flex, Auto, Min, Max）
// 2. Cell Borders 功能测试
// 3. 跨行跨列（Span）功能测试
// 4. Gap and Padding 测试
// 5. 边界条件和极限测试
// 6. 错误处理和异常情况
// 7. 性能测试
// 8. 布局精度和坐标计算验证
// ==============================================================================

package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// 1. 基本功能测试 - GridDimension 组合
// =============================================================================

// TestGrid_AllFixedDimensions 测试所有维度为 Fixed
func TestGrid_AllFixedDimensions(t *testing.T) {
	grid := NewMockGridNode("grid-fixed", 300, 200)
	grid.SetGridColumns([]GridDimension{
		GridFixed(100),
		GridFixed(100),
		GridFixed(100),
	})
	grid.SetGridRows([]GridDimension{
		GridFixed(100),
		GridFixed(100),
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	assert.Equal(t, 300, result.Root.Width, "Grid width should be sum of columns")
	assert.Equal(t, 200, result.Root.Height, "Grid height should be sum of rows")
}

// TestGrid_AllFlexDimensions 测试所有维度为 Flex
func TestGrid_AllFlexDimensions(t *testing.T) {
	grid := NewMockGridNode("grid-flex", 300, 200)
	grid.SetGridColumns([]GridDimension{
		GridFlex{Factor: 1},
		GridFlex{Factor: 2},
		GridFlex{Factor: 1},
	})
	grid.SetGridRows([]GridDimension{
		GridFlex{Factor: 1},
		GridFlex{Factor: 1},
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	// 验证总宽度正确
	assert.Equal(t, 300, result.Root.Width)
	assert.Equal(t, 200, result.Root.Height)
}

// TestGrid_AllAutoDimensions 测试所有维度为 Auto
func TestGrid_AllAutoDimensions(t *testing.T) {
	grid := NewMockGridNode("grid-auto", 300, 200)
	grid.SetGridColumns([]GridDimension{GridAuto{}, GridAuto{}, GridAuto{}})
	grid.SetGridRows([]GridDimension{GridAuto{}, GridAuto{}})

	// Auto 列/行需要子节点来决定尺寸
	cell1 := NewMockMeasurableNode("cell1", 50, 50)
	cell2 := NewMockMeasurableNode("cell2", 100, 100)
	cell3 := NewMockMeasurableNode("cell3", 150, 50)
	cell4 := NewMockMeasurableNode("cell4", 100, 50)

	grid.SetGridChildren([]Node{cell1, cell2, cell3, cell4})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	// Auto 列的宽度应该等于对应子节点的最大宽度
	// 行列 = [50, 100, 150], 列高 = [max(50,100,150), max(50, 100)] = [150, 100]
	assert.Equal(t, result.Root.Width, 50+100+150)
}

// TestGrid_MixedDimensions 测试混合维度类型
func TestGrid_MixedDimensions(t *testing.T) {
	grid := NewMockGridNode("grid-mixed", 400, 300)
	grid.SetGridColumns([]GridDimension{
		GridFixed(100),          // 固定宽度
		GridFlex{Factor: 1},     // 弹性宽度
		GridAuto{},              // 自动宽度
	})
	grid.SetGridRows([]GridDimension{
		GridFixed(100),          // 固定高度
		GridFlex{Factor: 1},     // 弹性高度
		GridAuto{},              // 自动高度
	})

	// 为 Auto/Cell 添加子节点
	child1 := NewMockMeasurableNode("c1", 50, 50)
	child2 := NewMockMeasurableNode("c2", 100, 100)
	child3 := NewMockMeasurableNode("c3", 80, 80)
	grid.SetGridChildren([]Node{child1, child2, child3})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	// 验证基本布局未崩溃
	assert.True(t, result.Root.Width > 0)
	assert.True(t, result.Root.Height > 0)
}

// TestGrid_MinDimension 测试 Min 维度
func TestGrid_MinDimension(t *testing.T) {
	grid := NewMockGridNode("grid-min", 300, 200)
	grid.SetGridColumns([]GridDimension{
		GridMin{Min: 100, Content: GridAuto{}},  // 最小 100，Auto 内容
		GridFlex{Factor: 1},
	})
	grid.SetGridRows([]GridDimension{
		GridMin{Min: 100, Content: GridAuto{}},
		GridFlex{Factor: 1},
	})

	// 设置要求更大的子节点
	child1 := NewMockMeasurableNode("c1", 150, 150)
	child2 := NewMockMeasurableNode("c2", 50, 50)
	grid.SetGridChildren([]Node{child1, child2})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	// 验证基本布局未崩溃
	assert.True(t, result.Root.Width > 0)
	assert.True(t, result.Root.Height > 0)
}

// TestGrid_MaxDimension 测试 Max 维度
func TestGrid_MaxDimension(t *testing.T) {
	grid := NewMockGridNode("grid-max", 300, 200)
	grid.SetGridColumns([]GridDimension{
		GridMax{Max: 100, Content: GridAuto{}},  // 最大 100，Auto 内容
		GridFlex{Factor: 1},
	})
	grid.SetGridRows([]GridDimension{
		GridMax{Max: 100, Content: GridAuto{}},
		GridFlex{Factor: 1},
	})

	// 设置要求更大的子节点（应该被限制）
	child1 := NewMockMeasurableNode("c1", 200, 200)
	child2 := NewMockMeasurableNode("c2", 50, 50)
	grid.SetGridChildren([]Node{child1, child2})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	// 验证基本布局未崩溃
	assert.True(t, result.Root.Width > 0)
	assert.True(t, result.Root.Height > 0)
}

// =============================================================================
// 2. Cell Borders 功能测试
// =============================================================================

// TestGrid_CellBorders_Basic 基础 Cell Borders 测试
func TestGrid_CellBorders_Basic(t *testing.T) {
	grid := NewMockGridNode("grid-borders", 203, 103)
	grid.SetGridColumns([]GridDimension{GridFixed(100), GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.GetGridStyle().ShowCellBorders = true
	grid.GetGridStyle().CellBorderWidth = 1
	grid.GetGridStyle().CellBorderHeight = 1

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	// 验证子节点位置因边框而偏移
	if len(result.Root.Children) > 0 {
		// 第一个子节点应该在 (1, 1)，因为要跳过左边框和上边框
		child0 := result.Root.Children[0]
		assert.Equal(t, 1, child0.X, "First child should offset for left border")
		assert.Equal(t, 1, child0.Y, "First child should offset for top border")
	}
}

// TestGrid_CellBorders_LargeGrid 大型 Grid 的 Cell Borders 测试
func TestGrid_CellBorders_LargeGrid(t *testing.T) {
	grid := NewMockGridNode("grid-large-borders", 506, 405)
	grid.SetGridColumns([]GridDimension{
		GridFixed(100), GridFixed(100), GridFixed(100), GridFixed(100), GridFixed(100),
	})
	grid.SetGridRows([]GridDimension{
		GridFixed(100), GridFixed(100), GridFixed(100), GridFixed(100),
	})
	grid.GetGridStyle().ShowCellBorders = true
	grid.GetGridStyle().CellBorderWidth = 1
	grid.GetGridStyle().CellBorderHeight = 1

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	// 验证能正常布局不崩溃
	assert.Equal(t, 506, result.Root.Width)
	assert.Equal(t, 405, result.Root.Height)
}

// TestGrid_CellBorders_WithGap 边框 + Gap 组合测试
func TestGrid_CellBorders_WithGap(t *testing.T) {
	grid := NewMockGridNode("grid-borders-gap", 208, 108)
	grid.SetGridColumns([]GridDimension{GridFixed(100), GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridGap(5, 5)  // 列间距和行间距都是 5
	grid.GetGridStyle().ShowCellBorders = true

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	// 验证能正常布局不崩溃
	assert.Equal(t, 208, result.Root.Width)
	assert.Equal(t, 108, result.Root.Height)
}

// TestGrid_CellBorders_WithPadding 边框 + Padding 组合测试
func TestGrid_CellBorders_WithPadding(t *testing.T) {
	grid := NewMockGridNode("grid-borders-padding", 300, 200)
	grid.SetGridColumns([]GridDimension{GridFixed(100), GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.GetGridStyle().Padding = Padding{Top: 10, Right: 10, Bottom: 10, Left: 10}
	grid.GetGridStyle().ShowCellBorders = true

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	// 验证能正常布局不崩溃
	assert.Equal(t, 300, result.Root.Width)
	assert.Equal(t, 200, result.Root.Height)
}

// TestGrid_CellBorders_LayoutPosition 边框布局位置计算测试
func TestGrid_CellBorders_LayoutPosition(t *testing.T) {
	grid := NewMockGridNode("grid-borders-pos", 200, 150)
	grid.SetGridColumns([]GridDimension{GridFixed(100), GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.GetGridStyle().ShowCellBorders = true

	// 添加子节点
	cells := []GridCell{
		{Child: NewMockNode("c0-0", 50, 50), Row: 0, Col: 0},
		{Child: NewMockNode("c0-1", 50, 50), Row: 0, Col: 1},
		{Child: NewMockNode("c1-0", 50, 50), Row: 1, Col: 0},
		{Child: NewMockNode("c1-1", 50, 50), Row: 1, Col: 1},
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)
	require.Len(t, result.Root.Children, 4)

	// 子节点的位置应该跳过边框字符
	// c0-0 应该在 (1, 1) 而不是 (0, 0)，因为要跳过上边框和左边框
	child0 := result.Root.Children[0]
	assert.Equal(t, 1, child0.X, "Cell (0,0) X should skip left border")
	assert.Equal(t, 1, child0.Y, "Cell (0,0) Y should skip top border")

	// c0-1 应该在 (102, 1)，跳过上边框和第一列的右边框
	child1 := result.Root.Children[1]
	assert.Equal(t, 102, child1.X, "Cell (0,1) X should skip left border + cell 0 + separator")
	assert.Equal(t, 1, child1.Y, "Cell (0,1) Y should skip top border")

	// c1-0 应该在 (1, 52)，跳过上边框、第一行的下边框
	child2 := result.Root.Children[2]
	assert.Equal(t, 1, child2.X, "Cell (1,0) X should skip left border")
	assert.Equal(t, 52, child2.Y, "Cell (1,0) Y should skip top border + cell 0 + separator")
}

// =============================================================================
// 3. Span（跨行跨列）功能测试
// =============================================================================

// TestGrid_Span_Basic 基础 Span 测试
func TestGrid_Span_Basic(t *testing.T) {
	grid := NewMockGridNode("grid-span", 300, 200)
	grid.SetGridColumns([]GridDimension{GridFixed(100), GridFixed(100), GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(100), GridFixed(100)})

	// 第一个单元格跨越2列
	cells := []GridCell{
		{Child: NewMockMeasurableNode("span-2-col", 200, 100), Row: 0, Col: 0, ColSpan: 2},
		{Child: NewMockMeasurableNode("normal", 100, 100), Row: 0, Col: 2},
		{Child: NewMockMeasurableNode("span-2-row", 100, 200), Row: 1, Col: 0, RowSpan: 2},
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)
	assert.Len(t, result.Root.Children, 3)
}

// TestGrid_Span_FullRow 跨整行
func TestGrid_Span_FullRow(t *testing.T) {
	grid := NewMockGridNode("grid-full-row", 300, 200)
	grid.SetGridColumns([]GridDimension{GridFixed(100), GridFixed(100), GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(100), GridFixed(100)})

	// Header 跨越整行
	cells := []GridCell{
		{Child: NewMockMeasurableNode("header", 300, 100), Row: 0, Col: 0, ColSpan: 3},
		{Child: NewMockMeasurableNode("c1", 100, 100), Row: 1, Col: 0},
		{Child: NewMockMeasurableNode("c2", 100, 100), Row: 1, Col: 1},
		{Child: NewMockMeasurableNode("c3", 100, 100), Row: 1, Col: 2},
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	// Header 应该占据第一行的整个宽度
	headerChild := result.Root.Children[0]
	assert.Equal(t, 300, headerChild.Width, "Header should span full row width")
}

// TestGrid_Span_FullColumn 跨整列
func TestGrid_Span_FullColumn(t *testing.T) {
	grid := NewMockGridNode("grid-full-col", 300, 200)
	grid.SetGridColumns([]GridDimension{GridFixed(100), GridFixed(100), GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(100), GridFixed(100)})

	// Sidebar 跨越整列
	cells := []GridCell{
		{Child: NewMockMeasurableNode("sidebar", 100, 200), Row: 0, Col: 0, RowSpan: 2},
		{Child: NewMockMeasurableNode("c1", 200, 100), Row: 0, Col: 1, ColSpan: 2},
		{Child: NewMockMeasurableNode("c2", 200, 100), Row: 1, Col: 1, ColSpan: 2},
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	// Sidebar 应该占据第一列的整个高度
	sidebarChild := result.Root.Children[0]
	assert.Equal(t, 200, sidebarChild.Height, "Sidebar should span full column height")
}

// TestGrid_Span_Multiple 多个跨行跨列组合
func TestGrid_Span_Multiple(t *testing.T) {
	grid := NewMockGridNode("grid-multi-span", 400, 300)
	grid.SetGridColumns([]GridDimension{
		GridFixed(100), GridFixed(100), GridFixed(100), GridFixed(100),
	})
	grid.SetGridRows([]GridDimension{
		GridFixed(100), GridFixed(100), GridFixed(100),
	})

	// 复杂的 span 组合
	cells := []GridCell{
		{Child: NewMockMeasurableNode("header", 400, 100), Row: 0, Col: 0, ColSpan: 4},    // 顶栏跨4列
		{Child: NewMockMeasurableNode("sidebar", 100, 200), Row: 1, Col: 0, RowSpan: 2},    // 侧栏跨2行
		{Child: NewMockMeasurableNode("main", 300, 200), Row: 1, Col: 1, ColSpan: 3},       // 主内容跨3列行1
		{Child: NewMockMeasurableNode("footer", 400, 100), Row: 2, Col: 1, ColSpan: 3},     // 底栏跨3列
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)
	assert.Len(t, result.Root.Children, 4)
}

// =============================================================================
// 4. Gap 和 Padding 测试
// =============================================================================

// TestGrid_ColumnGap 列间距
func TestGrid_ColumnGap(t *testing.T) {
	grid := NewMockGridNode("grid-col-gap", 320, 200)
	grid.SetGridColumns([]GridDimension{GridFixed(100), GridFixed(100), GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(100), GridFixed(100)})
	grid.SetGridGap(10, 0)  // 列间距 10

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	assert.Equal(t, 320, result.Root.Width)

	// 验证子节点位置（考虑 Gap）
	if len(result.Root.Children) > 1 {
		child1 := result.Root.Children[1]
		// 第二个子节点应该在 100 + 10(gap) = 110
		assert.Equal(t, 110, child1.X)
	}
}

// TestGrid_RowGap 行间距
func TestGrid_RowGap(t *testing.T) {
	grid := NewMockGridNode("grid-row-gap", 300, 210)
	grid.SetGridColumns([]GridDimension{GridFixed(100), GridFixed(100), GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(100), GridFixed(100)})
	grid.SetGridGap(0, 10)  // 行间距 10

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	assert.Equal(t, 210, result.Root.Height)
}

// TestGrid_BothGaps 列+行间距
func TestGrid_BothGaps(t *testing.T) {
	grid := NewMockGridNode("grid-both-gaps", 400, 300)
	grid.SetGridColumns([]GridDimension{GridFixed(100), GridFixed(100), GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(100), GridFixed(100)})
	grid.SetGridGap(10, 15)  // 列间距 10，行间距 15

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	// 验证可以正常布局
	assert.True(t, result.Root.Width > 0)
	assert.True(t, result.Root.Height > 0)
}

// TestGrid_Padding 基础 Padding
func TestGrid_Padding(t *testing.T) {
	grid := NewMockGridNode("grid-padding", 260, 140)
	grid.SetGridColumns([]GridDimension{GridFixed(100), GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.GetGridStyle().Padding = Padding{Top: 10, Right: 20, Bottom: 30, Left: 40}

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	assert.Equal(t, 260, result.Root.Width)
	assert.Equal(t, 140, result.Root.Height)
}

// TestGrid_Gap_With_Padding Gap + Padding 组合
func TestGrid_Gap_With_Padding(t *testing.T) {
	grid := NewMockGridNode("grid-gap-padding", 230, 125)
	grid.SetGridColumns([]GridDimension{GridFixed(100), GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridGap(10, 5)
	grid.GetGridStyle().Padding = Padding{Top: 10, Right: 10, Bottom: 10, Left: 10}

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	assert.Equal(t, 230, result.Root.Width)
	assert.Equal(t, 125, result.Root.Height)
}

// =============================================================================
// 5. 边界条件和极限测试
// =============================================================================

// TestGrid_EmptyChildren 空子节点
func TestGrid_EmptyChildren(t *testing.T) {
	grid := NewMockGridNode("grid-empty", 100, 100)
	grid.SetGridColumns([]GridDimension{GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(100)})
	grid.SetGridChildren([]Node{})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	// 即使没有子节点，Grid 仍然应该有尺寸
	assert.Equal(t, 100, result.Root.Width)
	assert.Equal(t, 100, result.Root.Height)
}

// TestGrid_NilChildren nil 子节点
func TestGrid_NilChildren(t *testing.T) {
	grid := NewMockGridNode("grid-nil", 100, 100)
	grid.SetGridColumns([]GridDimension{GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(100)})
	grid.SetGridChildren(nil)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)
}

// TestGrid_ZeroDimensions 零尺寸
func TestGrid_ZeroDimensions(t *testing.T) {
	grid := NewMockGridNode("grid-zero", 0, 0)
	grid.SetGridColumns([]GridDimension{GridFixed(0)})
	grid.SetGridRows([]GridDimension{GridFixed(0)})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	assert.Equal(t, 0, result.Root.Width)
	assert.Equal(t, 0, result.Root.Height)
}

// TestGrid_SingleCell 单个单元格
func TestGrid_SingleCell(t *testing.T) {
	grid := NewMockGridNode("grid-single", 100, 100)
	grid.SetGridColumns([]GridDimension{GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(100)})

	cells := []GridCell{
		{Child: NewMockMeasurableNode("single", 100, 100), Row: 0, Col: 0},
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)
	assert.Len(t, result.Root.Children, 1)
}

// TestGrid_LargeDimensions 大尺寸
func TestGrid_LargeDimensions(t *testing.T) {
	grid := NewMockGridNode("grid-large", 10000, 10000)
	grid.SetGridColumns([]GridDimension{GridFixed(5000), GridFixed(5000)})
	grid.SetGridRows([]GridDimension{GridFixed(5000), GridFixed(5000)})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	assert.Equal(t, 10000, result.Root.Width)
	assert.Equal(t, 10000, result.Root.Height)
}

// TestGrid_SmallDimensions 小尺寸
func TestGrid_SmallDimensions(t *testing.T) {
	grid := NewMockGridNode("grid-small", 10, 10)
	grid.SetGridColumns([]GridDimension{GridFixed(5), GridFixed(5)})
	grid.SetGridRows([]GridDimension{GridFixed(5), GridFixed(5)})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	assert.Equal(t, 10, result.Root.Width)
	assert.Equal(t, 10, result.Root.Height)
}

// TestGrid_UnevenColumns 不均匀列
func TestGrid_UnevenColumns(t *testing.T) {
	grid := NewMockGridNode("grid-uneven-cols", 310, 200)
	grid.SetGridColumns([]GridDimension{
		GridFixed(10), GridFixed(20), GridFixed(30), GridFixed(50), GridFixed(100), GridFixed(100),
	})
	grid.SetGridRows([]GridDimension{GridFixed(100), GridFixed(100)})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	assert.Equal(t, 310, result.Root.Width)
}

// TestGrid_UnevenRows 不均匀行
func TestGrid_UnevenRows(t *testing.T) {
	grid := NewMockGridNode("grid-uneven-rows", 200, 310)
	grid.SetGridColumns([]GridDimension{GridFixed(100), GridFixed(100)})
	grid.SetGridRows([]GridDimension{
		GridFixed(10), GridFixed(20), GridFixed(30), GridFixed(50), GridFixed(100), GridFixed(100),
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	assert.Equal(t, 310, result.Root.Height)
}

// TestGrid_MoreColumnsThanRows 列数 > 行数
func TestGrid_MoreColumnsThanRows(t *testing.T) {
	grid := NewMockGridNode("grid-more-cols", 500, 100)
	grid.SetGridColumns([]GridDimension{
		GridFixed(100), GridFixed(100), GridFixed(100), GridFixed(100), GridFixed(100),
	})
	grid.SetGridRows([]GridDimension{GridFixed(100)})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	assert.Equal(t, 500, result.Root.Width)
	assert.Equal(t, 100, result.Root.Height)
}

// TestGrid_MoreRowsThanColumns 行数 > 列数
func TestGrid_MoreRowsThanColumns(t *testing.T) {
	grid := NewMockGridNode("grid-more-rows", 100, 500)
	grid.SetGridColumns([]GridDimension{GridFixed(100)})
	grid.SetGridRows([]GridDimension{
		GridFixed(100), GridFixed(100), GridFixed(100), GridFixed(100), GridFixed(100),
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	assert.Equal(t, 100, result.Root.Width)
	assert.Equal(t, 500, result.Root.Height)
}

// TestGrid_MaximumSpan 最大跨度
func TestGrid_MaximumSpan(t *testing.T) {
	grid := NewMockGridNode("grid-max-span", 300, 200)
	grid.SetGridColumns([]GridDimension{
		GridFixed(100), GridFixed(100), GridFixed(100),
	})
	grid.SetGridRows([]GridDimension{
		GridFixed(100), GridFixed(100),
	})

	// 跨越整个网格
	cells := []GridCell{
		{Child: NewMockMeasurableNode("full-grid", 300, 200), Row: 0, Col: 0, RowSpan: 2, ColSpan: 3},
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)
	assert.Len(t, result.Root.Children, 1)

	child0 := result.Root.Children[0]
	assert.Equal(t, 300, child0.Width)
	assert.Equal(t, 200, child0.Height)
}

// =============================================================================
// 6. 错误处理和异常情况
// ==============================================================================

// TestGrid_SpanOutOfBounds 跨度超出边界
func TestGrid_SpanOutOfBounds(t *testing.T) {
	grid := NewMockGridNode("grid-span-oob", 200, 100)
	grid.SetGridColumns([]GridDimension{GridFixed(100), GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})

	// 跨度超出网格
	cells := []GridCell{
		{Child: NewMockMeasurableNode("oob", 100, 100), Row: 0, Col: 0, RowSpan: 3, ColSpan: 3},
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	// 不应该崩溃，返回结果
	require.NotNil(t, result)
	require.NotNil(t, result.Root)
}

// TestGrid_NegativeSpan 负跨度
func TestGrid_NegativeSpan(t *testing.T) {
	grid := NewMockGridNode("grid-neg-span", 200, 100)
	grid.SetGridColumns([]GridDimension{GridFixed(100), GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})

	cells := []GridCell{
		{Child: NewMockMeasurableNode("neg", 100, 50), Row: 0, Col: 0, RowSpan: -1, ColSpan: -1},
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	// 不应该崩溃
	require.NotNil(t, result)
}

// TestGrid_ZeroSpan 零跨度
func TestGrid_ZeroSpan(t *testing.T) {
	grid := NewMockGridNode("grid-zero-span", 200, 100)
	grid.SetGridColumns([]GridDimension{GridFixed(100), GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})

	cells := []GridCell{
		{Child: NewMockMeasurableNode("zero-span", 100, 50), Row: 0, Col: 0, RowSpan: 0, ColSpan: 0},
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	// 不应该崩溃
	require.NotNil(t, result)
}

// TestGrid_NegativeNegativeRowCol 负行列
func TestGrid_NegativeRowCol(t *testing.T) {
	grid := NewMockGridNode("grid-neg-rowcol", 200, 100)
	grid.SetGridColumns([]GridDimension{GridFixed(100), GridFixed(100)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})

	cells := []GridCell{
		{Child: NewMockMeasurableNode("neg-pos", 100, 50), Row: -1, Col: -1},
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	// 不应该崩溃
	require.NotNil(t, result)
}

// TestGrid_OverflowingChildren 子节点超出网格
func TestGrid_OverflowingChildren(t *testing.T) {
	grid := NewMockGridNode("grid-overflow", 100, 100)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})

	// 100 个子节点远超 2x2 网格
	children := make([]Node, 100)
	for i := 0; i < 100; i++ {
		children[i] = NewMockMeasurableNode("child", 50, 50)
	}
	grid.SetGridChildren(children)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	// 不应该崩溃
	require.NotNil(t, result)
}

// TestGrid_OverlappingCells 重叠单元格
func TestGrid_OverlappingCells(t *testing.T) {
	grid := NewMockGridNode("grid-overlap", 100, 100)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})

	// 两个单元格占据相同位置
	cells := []GridCell{
		{Child: NewMockMeasurableNode("cell1", 50, 50), Row: 0, Col: 0},
		{Child: NewMockMeasurableNode("cell2", 50, 50), Row: 0, Col: 0},
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	// 不应该崩溃
	require.NotNil(t, result)
}

// TestGrid_NilCells nil 单元格
func TestGrid_NilCells(t *testing.T) {
	grid := NewMockGridNode("grid-nil-cells", 100, 100)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})

	cells := []GridCell{
		{Child: nil, Row: 0, Col: 0},
		{Child: NewMockMeasurableNode("cell2", 50, 50), Row: 0, Col: 1},
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	// 不应该崩溃
	require.NotNil(t, result)
}

// =============================================================================
// 7. 性能测试
// =============================================================================

// TestGrid_Performance_LargeGrid 大型网格性能
func TestGrid_Performance_LargeGrid(t *testing.T) {
	grid := NewMockGridNode("grid-perf", 10000, 10000)

	// 100 x 100 网格
	columns := make([]GridDimension, 100)
	rows := make([]GridDimension, 100)
	for i := 0; i < 100; i++ {
		columns[i] = GridFixed(100)
		rows[i] = GridFixed(100)
	}
	grid.SetGridColumns(columns)
	grid.SetGridRows(rows)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	assert.Equal(t, 10000, result.Root.Width)
	assert.Equal(t, 10000, result.Root.Height)
}

// TestGrid_Performance_WithCellBorders 带边框的大型网格
func TestGrid_Performance_WithCellBorders(t *testing.T) {
	grid := NewMockGridNode("grid-perf-borders", 1011, 1011)

	// 10 x 10 网格，带边框
	columns := make([]GridDimension, 10)
	rows := make([]GridDimension, 10)
	for i := 0; i < 10; i++ {
		columns[i] = GridFixed(100)
		rows[i] = GridFixed(100)
	}
	grid.SetGridColumns(columns)
	grid.SetGridRows(rows)
	grid.GetGridStyle().ShowCellBorders = true

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)

	assert.Equal(t, 1011, result.Root.Width)
	assert.Equal(t, 1011, result.Root.Height)
}

// =============================================================================
// 8. 布局精度和坐标计算验证
// =============================================================================

// TestGrid_PreciseColumnWidths 列宽精度
func TestGrid_PreciseColumnWidths(t *testing.T) {
	grid := NewMockGridNode("grid-precise-cols", 300, 200)
	grid.SetGridColumns([]GridDimension{
		GridFixed(100), GridFixed(100), GridFixed(100),
	})
	grid.SetGridRows([]GridDimension{
		GridFixed(100), GridFixed(100),
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	// 总宽度应该等于 3 * 100 = 300
	assert.Equal(t, 300, result.Root.Width)
	assert.Equal(t, 200, result.Root.Height)
}

// TestGrid_PreciseRowHeights 行高精度
func TestGrid_PreciseRowHeights(t *testing.T) {
	grid := NewMockGridNode("grid-precise-rows", 200, 300)
	grid.SetGridColumns([]GridDimension{
		GridFixed(100), GridFixed(100),
	})
	grid.SetGridRows([]GridDimension{
		GridFixed(100), GridFixed(100), GridFixed(100),
	})

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	// 总高度应该等于 3 * 100 = 300
	assert.Equal(t, 200, result.Root.Width)
	assert.Equal(t, 300, result.Root.Height)
}

// TestGrid_ChildPositionsWithGap 带Gap的子节点位置
func TestGrid_ChildPositionsWithGap(t *testing.T) {
	grid := NewMockGridNode("grid-child-pos-gap", 200, 150)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridGap(10, 5)

	// 手动设置 Grid children 而不是 Cells
	cells := []GridCell{
		{Child: NewMockMeasurableNode("c0-0", 50, 50), Row: 0, Col: 0},
		{Child: NewMockMeasurableNode("c0-1", 50, 50), Row: 0, Col: 1},
		{Child: NewMockMeasurableNode("c1-0", 50, 50), Row: 1, Col: 0},
		{Child: NewMockMeasurableNode("c1-1", 50, 50), Row: 1, Col: 1},
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)
	assert.Len(t, result.Root.Children, 4)

	// c0-0: (0, 0)
	child0 := result.Root.Children[0]
	assert.Equal(t, 0, child0.X)
	assert.Equal(t, 0, child0.Y)

	// c0-1: (50 + 10(gap), 0) = (60, 0)
	child1 := result.Root.Children[1]
	assert.Equal(t, 60, child1.X)
	assert.Equal(t, 0, child1.Y)

	// c1-0: (0, 50 + 5(gap)) = (0, 55)
	child2 := result.Root.Children[2]
	assert.Equal(t, 0, child2.X)
	assert.Equal(t, 55, child2.Y)

	// c1-1: (60, 55)
	child3 := result.Root.Children[3]
	assert.Equal(t, 60, child3.X)
	assert.Equal(t, 55, child3.Y)
}

// TestGrid_ChildPositionsWithPadding 带Padding的子节点位置
func TestGrid_ChildPositionsWithPadding(t *testing.T) {
	grid := NewMockGridNode("grid-child-pos-padding", 200, 150)
	grid.SetGridColumns([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.SetGridRows([]GridDimension{GridFixed(50), GridFixed(50)})
	grid.GetGridStyle().Padding = Padding{Top: 10, Right: 10, Bottom: 10, Left: 10}

	cells := []GridCell{
		{Child: NewMockMeasurableNode("c0-0", 50, 50), Row: 0, Col: 0},
		{Child: NewMockMeasurableNode("c0-1", 50, 50), Row: 0, Col: 1},
	}
	grid.SetGridCells(cells)

	engine := NewEngine()
	result := engine.Layout(grid, UnboundedConstraints())

	require.NotNil(t, result)
	require.NotNil(t, result.Root)

	// 验证可以正常布局不崩溃
	assert.True(t, result.Root.Width > 0)
	assert.True(t, result.Root.Height > 0)
}

// =============================================================================
// 辅助函数
// =============================================================================

// 辅助函数已被移除，因为测试可以直接验证 Width 和 Height 属性。
// 如需访问内部数据，可以通过 Engine 和 Grid 接口直接获取。
