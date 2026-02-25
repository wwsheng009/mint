package grid

import (
	"fmt"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
)

// =============================================================================
// Cell Borders 常量定义
// =============================================================================

const (
	CellBorderStyleNone   = "none"
	CellBorderStyleSingle = "single"
	CellBorderStyleDouble = "double"
	CellBorderStyleLight  = "light"
)

// BorderChars 边框字符集
type BorderChars struct {
	horizontal, vertical             string // 水平和垂直线
	topLeft, topRight, bottomLeft, bottomRight string // 四个角
	topCross, bottomCross, leftCross, rightCross string // 四个交点
	cross string // 中心交点
}

// 边框字符集定义
var cellBorderChars = map[string]BorderChars{
	"single": {
		horizontal:   "─",
		vertical:     "│",
		topLeft:      "┌",
		topRight:     "┐",
		bottomLeft:   "└",
		bottomRight:  "┘",
		topCross:     "┬",
		bottomCross:  "┴",
		leftCross:    "├",
		rightCross:   "┤",
		cross:        "┼",
	},
	"double": {
		horizontal:   "═",
		vertical:     "║",
		topLeft:      "╔",
		topRight:     "╗",
		bottomLeft:   "╚",
		bottomRight:  "╝",
		topCross:     "╦",
		bottomCross:  "╩",
		leftCross:    "╠",
		rightCross:   "╣",
		cross:        "╬",
	},
	"light": {
		horizontal:   "─",
		vertical:     "│",
		topLeft:      "┌",
		topRight:     "┐",
		bottomLeft:   "└",
		bottomRight:  "┘",
		topCross:     "┬",
		bottomCross:  "┴",
		leftCross:    "├",
		rightCross:   "┤",
		cross:        "┼",
	},
}

// 圆角边框字符
var roundedBorderChars = BorderChars{
	horizontal:   "─",
	vertical:     "│",
	topLeft:      "╭",
	topRight:     "╮",
	bottomLeft:   "╰",
	bottomRight:  "╯",
	topCross:     "┬",
	bottomCross:  "┴",
	leftCross:    "├",
	rightCross:   "┤",
	cross:        "┼",
}

// =============================================================================
// Cell Borders 绘制方法
// =============================================================================

// GenCellBorderDrawCmds 生成格子边框的绘制命令
// 返回 []paint.DrawCmd 用于绘制
func (inst *Instance) GenCellBorderDrawCmds(originX, originY int) []paint.DrawCmd {
	if inst == nil || !inst.showCellBorders {
		return nil
	}

	numCols := len(inst.colWidths)
	if numCols == 0 {
		return nil
	}
	numRows := len(inst.rowHeights)
	if numRows == 0 {
		return nil
	}

	cmds := []paint.DrawCmd{}

	// 获取边框字符集
	chars := cellBorderChars[inst.cellBorderStyle]
	if inst.cellBorderRounded && inst.cellBorderStyle == "single" {
		chars = roundedBorderChars
	}

	// 准备样式
	borderStyle := style.Style{}
	if inst.cellBorderColor != "" {
		borderStyle.FG = style.Color(inst.cellBorderColor)
	}

	// 计算内容区域起始位置（在 padding 和容器边框之后）
	contentX := originX + inst.padding[3]
	contentY := originY + inst.padding[0]

	// 绘制交点和角，以及完整水平线
	for row := 0; row <= numRows; row++ {
		// 计算当前行的 y 坐标
		y := contentY
		for r := 0; r < row; r++ {
			y += inst.rowHeights[r] + 1 // +1 是上边框宽度
			if r < row-1 {
				y += inst.rowGap
			}
		}

		// 绘制每个格子水平线的起始和交点
		for col := 0; col <= numCols; col++ {
			// 计算当前位置的 x 坐标
			x := contentX
			for c := 0; c < col; c++ {
				x += inst.colWidths[c] + 1 // +1 是左边框宽度
				if c < col-1 {
					x += inst.columnGap
				}
			}

			// 确定使用哪个字符（交点和角）
			var char string
			if row == 0 && col == 0 {
				char = chars.topLeft
			} else if row == 0 && col == numCols {
				char = chars.topRight
			} else if row == numRows && col == 0 {
				char = chars.bottomLeft
			} else if row == numRows && col == numCols {
				char = chars.bottomRight
			} else if row == 0 {
				char = chars.topCross
			} else if row == numRows {
				char = chars.bottomCross
			} else if col == 0 {
				char = chars.leftCross
			} else if col == numCols {
				char = chars.rightCross
			} else {
				char = chars.cross
			}

			// ✨ DEBUG: 打印四个角和底部边框字符
			if row == numRows {
				fmt.Printf("[DEBUG BORDERS] Bottom border char at(%d, %d): %s (row=%d, col=%d)\n", x, originY+y, char, row, col)
			}

			// 添加绘制命令（交点）
			if len(char) > 0 {
				cmds = append(cmds, paint.DrawCmd{
					X:     x,
					Y:     y,
					Text:  char, // 直接使用完整字符串
					Style: borderStyle,
				})
			}
		}

		// 绘制水平线（不包括交点）
		horizontalX := contentX
		for col := 0; col < numCols; col++ {
			// 绘制该格子的水平线内容
			for dx := 0; dx < inst.colWidths[col]; dx++ {
				if len(chars.horizontal) > 0 {
					// ✨ DEBUG: 打印底部边框水平线
					if row == numRows && col == 0 && dx == 0 {
						fmt.Printf("[DEBUG BORDERS] Bottom horizontal starting at(%d, %d)\n", horizontalX + 1 + dx, originY + y)
					}

					cmds = append(cmds, paint.DrawCmd{
						X:     horizontalX + 1 + dx,
						Y:     y,
						Text:  chars.horizontal, // 直接使用完整字符串
						Style: borderStyle,
					})
				}
			}
			// 移动到下一个格子位置（包括右边框）
			horizontalX += inst.colWidths[col] + 1 + inst.columnGap
		}
	}

	// 绘制垂直线（不包括交点）
	for col := 0; col <= numCols; col++ {
		// ✨ 垂直线 x 坐标 = contentX + sum(colWidths[0..col-1]) + col
		// 每个格子都有右边框(1)，最后一条边框是 Grid 右边界
		x := contentX
		for c := 0; c < col; c++ {
			x += inst.colWidths[c] + 1  // 格子内容宽度 + 右边框宽度(1)
			if c < col-1 {
				x += inst.columnGap
			}
		}

		// ✨ DEBUG: 打印垂直线位置
		fmt.Printf("[DEBUG BORDERS] Vertical line %d: x=%d (relative to originX=%d)\n", col, x-contentX, contentX)

		// 垂直线 y 坐标从 contentY 开始（在边框上），内容区域是 contentY + 1
		verticalY := contentY
		for row := 0; row < numRows; row++ {
			// 绘制该格子的垂直线内容（从 contentY + 1 到 contentY + 1 + rowHeights[row] - 1）
			for dy := 0; dy < inst.rowHeights[row]; dy++ {
				if len(chars.vertical) > 0 {
					cmds = append(cmds, paint.DrawCmd{
						X:     x,
						Y:     verticalY + 1 + dy,  // 从上边框之后开始绘制
						Text:  chars.vertical,      // 直接使用完整字符串
						Style: borderStyle,
					})
				}
			}
			// 移动到下一行（最后一行不加边框高度，因为底边框在那一行底部）
			if row == numRows-1 {
				verticalY += inst.rowHeights[row]  // 最后一行只加内容高度
			} else {
				verticalY += inst.rowHeights[row] + 1 + inst.rowGap  // 其他行：内容 + 边框 + 间距
			}
		}
	}

	return cmds
}
