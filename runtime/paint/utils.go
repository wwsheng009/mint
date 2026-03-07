package paint

import (
	"fmt"
	"strings"
)

// BufferCoordinatesString 返回 Buffer 的坐标网格视图字符串
// 上下显示 X 坐标（竖向显示），每列都有标注
// 左侧显示 Y 坐标
// 每个单元格固定 2 字符格式，保持对齐
func BufferCoordinatesString(buf *Buffer, width, height int) string {
	var sb strings.Builder
	maxY := height

	// Print X coordinate header vertically (for every column)
	// Each cell = 2 chars, so mark every 2 character positions
	sb.WriteString("     ")
	for x := 0; x < width; x++ {
		// Tens digit (skip if single digit)
		sb.WriteString(fmt.Sprintf("%d ", x/10))
	}
	sb.WriteString("\n")
	sb.WriteString("     ")
	for x := 0; x < width; x++ {
		// Ones digit
		sb.WriteString(fmt.Sprintf("%d ", x%10))
	}
	sb.WriteString("\n")
	sb.WriteString("     ")
	for x := 0; x < width; x++ {
		sb.WriteString("│ ")
	}
	sb.WriteString("\n\n")

	for y := 0; y < maxY; y++ {
		sb.WriteString(fmt.Sprintf("Y%02d: ", y))

		// Print each cell in this row
		for x := 0; x < width; x++ {
			cell := buf.GetContent(x, y)
			if cell.Cluster == "" || cell.Cluster == " " {
				sb.WriteString(". ") // 2 chars for empty cell
			} else {
				// Unified format - always 2 chars
				runes := []rune(cell.Cluster)
				if len(runes) > 0 {
					sb.WriteString(fmt.Sprintf("%c ", runes[0])) // char + space = 2 chars
				} else {
					sb.WriteString(". ")
				}
			}
		}
		sb.WriteString("\n")
	}

	// Print X coordinate footer vertically
	sb.WriteString("\n")
	sb.WriteString("     ")
	for x := 0; x < width; x++ {
		sb.WriteString("│ ")
	}
	sb.WriteString("\n")
	sb.WriteString("     ")
	for x := 0; x < width; x++ {
		// Tens digit
		sb.WriteString(fmt.Sprintf("%d ", x/10))
	}

	sb.WriteString("\n")
	sb.WriteString("     ")
	for x := 0; x < width; x++ {
		// Ones digit
		sb.WriteString(fmt.Sprintf("%d ", x%10))
	}

	return sb.String()
}
