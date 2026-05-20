// Package utils provides buffer observation helpers for ignored diagnostic tools.
package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/runtime/paint"
)

// PrintBuffer 将 Buffer 内容打印到终端
// width/height: Buffer 的尺寸
// 注意：此函数会正确处理宽字符（如中文）的延续单元格
func PrintBuffer(buf *paint.Buffer, width, height int) {
	fmt.Printf("┌%s┐\n", strings.Repeat("─", width))
	for y := 0; y < height; y++ {
		var line strings.Builder
		hasContent := false
		for x := 0; x < width; x++ {
			cell := buf.GetContent(x, y)
			// 跳过宽字符的延续单元格
			if cell.IsContinuation {
				continue
			}
			if cell.Cluster != "" && cell.Cluster != " " {
				line.WriteString(cell.Cluster)
				hasContent = true
			} else {
				line.WriteString(" ")
			}
		}
		if hasContent {
			trimmed := strings.TrimRight(line.String(), " ")
			fmt.Printf("|%-*s|\n", width, trimmed)
		}
	}
	fmt.Printf("└%s┘\n", strings.Repeat("─", width))
}

// SaveBuffer 将 Buffer 内容保存到文件 output.txt
// width/height: Buffer 的尺寸
// 注意：此函数会正确处理宽字符（如中文）的延续单元格
func SaveBuffer(buf *paint.Buffer, width, height int) {
	file, _ := os.Create("output.txt")
	defer file.Close()

	for y := 0; y < height; y++ {
		var line strings.Builder
		hasContent := false
		for x := 0; x < width; x++ {
			cell := buf.GetContent(x, y)
			// 跳过宽字符的延续单元格
			if cell.IsContinuation {
				continue
			}
			if cell.Cluster != "" && cell.Cluster != " " {
				line.WriteString(cell.Cluster)
				hasContent = true
			} else {
				line.WriteString(" ")
			}
		}
		if hasContent {
			file.WriteString(strings.TrimRight(line.String(), " ") + "\n")
		}
	}
	fmt.Println("📄 输出已保存到: output.txt")
}

// PrintBufferWithFilename 将 Buffer 内容打印到终端并保存到指定文件
// width/height: Buffer 的尺寸
// filename: 输出文件名（为空时不保存）
func PrintBufferWithFilename(buf *paint.Buffer, width, height int, filename string) {
	PrintBuffer(buf, width, height)

	if filename != "" {
		file, _ := os.Create(filename)
		defer file.Close()

		for y := 0; y < height; y++ {
			var line strings.Builder
			hasContent := false
			for x := 0; x < width; x++ {
				cell := buf.GetContent(x, y)
				if cell.IsContinuation {
					continue
				}
				if cell.Cluster != "" && cell.Cluster != " " {
					line.WriteString(cell.Cluster)
					hasContent = true
				} else {
					line.WriteString(" ")
				}
			}
			if hasContent {
				file.WriteString(strings.TrimRight(line.String(), " ") + "\n")
			}
		}
		fmt.Printf("📄 输出已保存到: %s\n", filename)
	}
}

// PrintBufferCoordinates 打印 Buffer 的坐标网格视图
// 上下显示 X 坐标（竖向显示），每列都有标注
// 左侧显示 Y 坐标
// 每个单元格固定 2 字符格式，保持对齐
func PrintBufferCoordinates(buf *paint.Buffer, width, height int) {
	fmt.Print(paint.BufferCoordinatesString(buf, width, height))
}
