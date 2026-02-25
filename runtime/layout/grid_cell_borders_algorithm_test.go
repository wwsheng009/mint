package layout

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Grid Cell Borders Algorithm Verification Test
// =============================================================================

// TestGridCellPositionAlgorithm 验证 getCellPosition 计算的坐标
func TestGridCellPositionAlgorithm(t *testing.T) {
	style := DefaultGridStyle()
	style.Columns = []GridDimension{GridFixed(10), GridFixed(10), GridFixed(10)}
	style.Rows = []GridDimension{GridFixed(5), GridFixed(5)}
	style.ShowCellBorders = true
	style.ColumnGap = 0
	style.RowGap = 0

	grid := NewGridLayout("test-grid", style)
	grid.Measure(TightConstraints(100, 20))

	fmt.Println("========== 边框坐标计算 ==========")
	fmt.Println("垂直边框位置：")
	for col := 0; col <= 3; col++ {
		x := calculateBorderPositionX(grid, col)
		fmt.Printf("  边框[%d] = x=%d\n", col, x)
	}

	fmt.Println("\n========== Cell 位置计算 (getCellPosition) ==========")
	testCases := []struct {
		cell     string
		row, col int
		wantX, wantY int
	}{
		{"(0,0)", 0, 0, 1, 1},
		{"(0,1)", 0, 1, 12, 1},
		{"(0,2)", 0, 2, 23, 1},
		{"(1,0)", 1, 0, 1, 7},
		{"(1,1)", 1, 1, 12, 7},
		{"(1,2)", 1, 2, 23, 7},
	}

	for _, tc := range testCases {
		t.Run(tc.cell, func(t *testing.T) {
			x, y := grid.getCellPosition(tc.row, tc.col)
			fmt.Printf("Cell %s: 位置=(%d, %d), 期望=(%d, %d)\n", tc.cell, x, y, tc.wantX, tc.wantY)

			assert.Equal(t, tc.wantX, x, "Cell %s 的 X 坐标不正确", tc.cell)
			assert.Equal(t, tc.wantY, y, "Cell %s 的 Y 坐标不正确", tc.cell)

			// 验证 Cell 位置在正确的边框之间
			borderLeftX := calculateBorderPositionX(grid, tc.col)
			borderRightX := calculateBorderPositionX(grid, tc.col+1)

			assert.True(t, x > borderLeftX, "Cell %s 应该在边框[%d]右边", tc.cell, tc.col)
			assert.True(t, x < borderRightX, "Cell %s 应该在边框[%d]左边", tc.cell, tc.col+1)
		})
	}
}

// TestGridCellSizeAlgorithm 验证 getCellSize 计算的大小
func TestGridCellSizeAlgorithm(t *testing.T) {
	style := DefaultGridStyle()
	style.Columns = []GridDimension{GridFixed(10), GridFixed(10), GridFixed(10)}
	style.Rows = []GridDimension{GridFixed(5), GridFixed(5)}
	style.ShowCellBorders = true
	style.ColumnGap = 0
	style.RowGap = 0

	grid := NewGridLayout("test-grid", style)
	grid.Measure(TightConstraints(100, 20))

	fmt.Println("\n========== Cell 大小计算 (getCellSize) ==========")
	testCases := []struct {
		cell          string
		row, col      int
		rowSpan, colSpan int
		wantWidth, wantHeight int
	}{
		// 单个 cell，span=1
		{"(0,0)", 0, 0, 1, 1, 10, 5},  // 只包含内容，不加边框
		{"(0,1)", 0, 1, 1, 1, 10, 5},
		{"(0,2)", 0, 2, 1, 1, 10, 5},
		{"(1,0)", 1, 0, 1, 1, 10, 5},
		{"(1,1)", 1, 1, 1, 1, 10, 5},
		{"(1,2)", 1, 2, 1, 1, 10, 5},

		// 跨列 cell
		{"(0,0) span2", 0, 0, 1, 2, 20, 5},  // 2列内容宽度
		{"(0,1) span2", 0, 1, 1, 2, 20, 5},

		// 跨行 cell
		{"(0,0) spanRow2", 0, 0, 2, 1, 10, 10},  // 2行内容高度

		// 跨行跨列
		{"(0,0) span2x2", 0, 0, 2, 2, 20, 10},
	}

	for _, tc := range testCases {
		t.Run(tc.cell, func(t *testing.T) {
			w, h := grid.getCellSize(tc.row, tc.col, tc.rowSpan, tc.colSpan)
			fmt.Printf("Cell %s: 大小=(%d, %d), 期望=(%d, %d)\n", tc.cell, w, h, tc.wantWidth, tc.wantHeight)

			assert.Equal(t, tc.wantWidth, w, "Cell %s 的宽度不正确", tc.cell)
			assert.Equal(t, tc.wantHeight, h, "Cell %s 的高度不正确", tc.cell)

			// 关键验证：Cell 大小不应该包含边框
			// 内容宽度应该等于列宽总和，不加边框的1字符
			expectedContentWidth := 0
			for i := tc.col; i < tc.col+tc.colSpan && i < len(grid.colWidths); i++ {
				expectedContentWidth += grid.colWidths[i]
			}
			assert.Equal(t, expectedContentWidth, w, "Cell %s 宽度应该等于列宽总和（不含边框）", tc.cell)
		})
	}
}

// TestGridCellDrawRange 验证 Cell 绘制范围不覆盖边框
func TestGridCellDrawRange(t *testing.T) {
	style := DefaultGridStyle()
	style.Columns = []GridDimension{GridFixed(10), GridFixed(10), GridFixed(10)}
	style.Rows = []GridDimension{GridFixed(5), GridFixed(5)}
	style.ShowCellBorders = true
	style.ColumnGap = 0
	style.RowGap = 0

	grid := NewGridLayout("test-grid", style)
	grid.Measure(TightConstraints(100, 20))

	fmt.Println("\n========== Cell 绘制范围验证 ==========")
	fmt.Println("边框位置:  x=0, 11, 22, 33")
	fmt.Println("预期 Cell 绘制范围（不覆盖边框）：")
	fmt.Println("  Cell(0,0): [1, 10]")
	fmt.Println("  Cell(0,1): [12, 21]")
	fmt.Println("  Cell(0,2): [23, 32]")
	fmt.Println()

	for row := 0; row < 2; row++ {
		for col := 0; col < 3; col++ {
			cellName := fmt.Sprintf("(%d,%d)", row, col)

			// 获取 Cell 位置
			x, y := grid.getCellPosition(row, col)

			// 获取 Cell 大小
			w, h := grid.getCellSize(row, col, 1, 1)

			// 计算 Cell 绘制范围（最后绘制的字符位置）
			endX := x + w - 1
			endY := y + h - 1

			fmt.Printf("Cell %s: 起始=(%d,%d), 大小=(%d,%d), 结束=(%d,%d)\n",
				cellName, x, y, w, h, endX, endY)

			// 验证：Cell 范围不应该覆盖任何边框
			borderLeftX := calculateBorderPositionX(grid, col)
			borderRightX := calculateBorderPositionX(grid, col+1)
			borderTopY := calculateBorderPositionY(grid, row)
			borderBottomY := calculateBorderPositionY(grid, row+1)

			// X 轴验证：Cell 应该在左右边框之间
			assert.True(t, x > borderLeftX, "Cell %s X起始 (%d) 应该 > 左边框X (%d)", cellName, x, borderLeftX)
			assert.True(t, endX < borderRightX, "Cell %s X结束 (%d) 应该 < 右边框X (%d)", cellName, endX, borderRightX)

			// Y 轴验证：Cell 应该在上下边框之间
			assert.True(t, y > borderTopY, "Cell %s Y起始 (%d) 应该 > 上边框Y (%d)", cellName, y, borderTopY)
			assert.True(t, endY < borderBottomY, "Cell %s Y结束 (%d) 应该 < 下边框Y (%d)", cellName, endY, borderBottomY)

			// 边距验证：Cell 应该与边框保持 1 字符的距离
			assert.Equal(t, borderLeftX + 1, x, "Cell %s X起始应该是 左边框+1", cellName)
			assert.Equal(t, borderTopY + 1, y, "Cell %s Y起始应该是 上边框+1", cellName)
			assert.Equal(t, borderRightX - 1, endX, "Cell %s X结束应该是 右边框-1", cellName)
			assert.Equal(t, borderBottomY - 1, endY, "Cell %s Y结束应该是 下边框-1", cellName)
		}
	}
}

// 计算垂直边框的 X 坐标（模拟 GenCellBorderDrawCmds 的逻辑）
func calculateBorderPositionX(grid *GridLayout, col int) int {
	x := 0
	for c := 0; c < col; c++ {
		x += grid.colWidths[c] + 1
	}
	return x
}

// 计算水平边框的 Y 坐标（模拟 GenCellBorderDrawCmds 的逻辑）
func calculateBorderPositionY(grid *GridLayout, row int) int {
	y := 0
	for r := 0; r < row; r++ {
		y += grid.rowHeights[r] + 1
	}
	return y
}
