package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/ui/layout/visualizer"
)

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║  Mint 布局可视化 - Buffer 方式演示                       ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()

	// 示例 1: 简单的嵌套布局
	fmt.Println("═════════════════════════════════════════════════════")
	fmt.Println("示例 1: 简单的嵌套布局")
	fmt.Println("═════════════════════════════════════════════════════")

	vis1 := visualizer.NewVisualizer()
	vis1.AddNode(
		"panel_root",
		"panel",
		layout.Rect{X: 0, Y: 0, Width: 50, Height: 20},
		layout.Constraints{MinWidth: 0, MaxWidth: 80, MinHeight: 0, MaxHeight: 30},
		layout.Constraints{MinWidth: 0, MaxWidth: 78, MinHeight: 0, MaxHeight: 28},
		layout.Size{Width: 50, Height: 20},
		"",
	)
	vis1.AddNode(
		"border_child",
		"bordered",
		layout.Rect{X: 1, Y: 1, Width: 48, Height: 18},
		layout.Constraints{MinWidth: 0, MaxWidth: 78, MinHeight: 0, MaxHeight: 28},
		layout.Constraints{MinWidth: 0, MaxWidth: 76, MinHeight: 0, MaxHeight: 26},
		layout.Size{Width: 48, Height: 18},
		"panel_root",
	)
	vis1.AddNode(
		"text_content",
		"text",
		layout.Rect{X: 2, Y: 2, Width: 46, Height: 16},
		layout.Constraints{MinWidth: 0, MaxWidth: 76, MinHeight: 0, MaxHeight: 26},
		layout.Constraints{MinWidth: 0, MaxWidth: 74, MinHeight: 0, MaxHeight: 24},
		layout.Size{Width: 46, Height: 16},
		"border_child",
	)

	fmt.Println(vis1.PrintBoxModel())

	// 示例 2: 复杂的 Stack 嵌套布局
	fmt.Println("═════════════════════════════════════════════════════")
	fmt.Println("示例 2: Stack 嵌套布局 (多个子元素)")
	fmt.Println("═════════════════════════════════════════════════════")

	vis2 := visualizer.NewVisualizer()
	vis2.AddNode(
		"panel_complex",
		"panel",
		layout.Rect{X: 0, Y: 0, Width: 50, Height: 15},
		layout.Constraints{MinWidth: 0, MaxWidth: 80, MinHeight: 0, MaxHeight: 30},
		layout.Constraints{MinWidth: 0, MaxWidth: 78, MinHeight: 0, MaxHeight: 28},
		layout.Size{Width: 50, Height: 15},
		"",
	)
	vis2.AddNode(
		"vstack_main",
		"vstack",
		layout.Rect{X: 2, Y: 2, Width: 46, Height: 11},
		layout.Constraints{MinWidth: 0, MaxWidth: 78, MinHeight: 0, MaxHeight: 28},
		layout.Constraints{MinWidth: 0, MaxWidth: 76, MinHeight: 0, MaxHeight: 26},
		layout.Size{Width: 46, Height: 11},
		"panel_complex",
	)
	vis2.AddNode(
		"element1",
		"text",
		layout.Rect{X: 4, Y: 4, Width: 42, Height: 2},
		layout.Constraints{MinWidth: 0, MaxWidth: 76, MinHeight: 0, MaxHeight: 26},
		layout.Constraints{MinWidth: 0, MaxWidth: 74, MinHeight: 0, MaxHeight: 24},
		layout.Size{Width: 42, Height: 2},
		"vstack_main",
	)
	vis2.AddNode(
		"element2",
		"button",
		layout.Rect{X: 4, Y: 6, Width: 15, Height: 2},
		layout.Constraints{MinWidth: 0, MaxWidth: 76, MinHeight: 0, MaxHeight: 26},
		layout.Constraints{MinWidth: 0, MaxWidth: 74, MinHeight: 0, MaxHeight: 24},
		layout.Size{Width: 15, Height: 2},
		"vstack_main",
	)
	vis2.AddNode(
		"element3",
		"text",
		layout.Rect{X: 4, Y: 8, Width: 42, Height: 3},
		layout.Constraints{MinWidth: 0, MaxWidth: 76, MinHeight: 0, MaxHeight: 26},
		layout.Constraints{MinWidth: 0, MaxWidth: 74, MinHeight: 0, MaxHeight: 24},
		layout.Size{Width: 42, Height: 3},
		"vstack_main",
	)

	fmt.Println(vis2.PrintBoxModel())

	// 验证边框完整性
	fmt.Println("═════════════════════════════════════════════════════")
	fmt.Println("验证边框完整性")
	fmt.Println("═════════════════════════════════════════════════════")

	output := vis2.PrintBoxModel()
	lines := strings.Split(output, "\n")

	panelRightBorderCount := 0
	vstackRightBorderCount := 0

	for _, line := range lines {
		// 统计最右侧的 │ 数量
		if len(line) > 50 && strings.Contains(line, "│") {
			runes := []rune(line)
			if len(runes) > 0 && runes[len(runes)-1] == '│' {
				panelRightBorderCount++
			}
		}

		// 统计第二层（vstack）的右边框
		if strings.HasPrefix(line, "│ ┌") || strings.HasPrefix(line, "│ │ ") || strings.HasPrefix(line, "│ ├") {
			runes := []rune(line)
			if len(runes) > 50 && runes[len(runes)-1] == '│' {
				vstackRightBorderCount++
			}
		}
	}

	fmt.Printf("✅ Panel 右边框 │ 总数: %d\n", panelRightBorderCount)
	fmt.Printf("✅ VStack 右边框 │ 总数: %d\n", vstackRightBorderCount)

	if panelRightBorderCount > 5 && vstackRightBorderCount > 5 {
		fmt.Println("✅ 边框完整性验证通过！")
	} else {
		fmt.Println("❌ 边框可能不完整")
	}

	fmt.Println("\n═════════════════════════════════════════════════════")
	fmt.Println("✅ Buffer 方式演示完成")
	fmt.Println("═════════════════════════════════════════════════════")
}
