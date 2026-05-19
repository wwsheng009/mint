package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/examples/component_fixtures"
	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime"
	compute_engine "github.com/wwsheng009/mint/runtime/compute"
)

// 本程序使用 ASCII 艺术可视化布局结果
// 显示每个节点的位置、尺寸和层级关系
func main() {
	fmt.Println("=== Layout ASCII Visualization ===")

	// 获取测试组件
	fixtures := component_fixtures.StandardFixtures()

	for _, fixture := range fixtures {
		fmt.Printf("========================================\n")
		fmt.Printf("Fixture: %s\n", fixture.Name)
		fmt.Printf("Description: %s\n", fixture.Description)
		fmt.Printf("========================================\n\n")

		// 构建VNode
		vnode := fixture.Build()
		if vnode == nil {
			fmt.Printf("❌ Failed to build VNode\n\n")
			continue
		}

		// 创建Fiber
		fiber := reconciler.CreateFiberFromVNode(vnode)
		if fiber == nil {
			fmt.Printf("❌ Failed to create Fiber\n\n")
			continue
		}

		// 执行布局
		engine := compute_engine.NewEngine()
		constraints := runtime.BoxConstraints{
			MinWidth:  80,
			MaxWidth:  80,
			MinHeight: 24,
			MaxHeight: 24,
		}

		layout, err := engine.Layout(vnode, fiber, constraints)
		if err != nil {
			fmt.Printf("❌ Layout failed: %v\n\n", err)
			continue
		}

		if layout == nil || layout.Root == nil {
			fmt.Printf("❌ Layout result is nil\n\n")
			continue
		}

		// 可视化布局
		visualizeLayout(fixture.Name, layout)
		fmt.Println()
	}

	fmt.Println("=== Visualization Complete ===")
}

// visualizeLayout 使用ASCII艺术可视化布局
func visualizeLayout(name string, layout *compute_engine.ComputedLayout) {
	if layout == nil || layout.Root == nil {
		return
	}

	// 1. 显示布局统计
	printLayoutStats(layout)

	// 2. 显示层级树视图
	fmt.Println("\n--- Layout Tree (ASCII View) ---")
	printLayoutTree(layout.Root, 0)

	// 3. 显示可视化网格
	fmt.Println("\n--- Layout Grid Visualization ---")
	printLayoutGrid(layout.Root, 80, 24)

	// 4. 显示节点详情
	fmt.Println("\n--- Node Details ---")
	printNodeDetails(layout.Root, 0)
}

// printLayoutStats 显示布局统计
func printLayoutStats(layout *compute_engine.ComputedLayout) {
	root := layout.Root
	totalNodes := countNodes(root)
	totalDepth := maxDepth(root)

	fmt.Println("Layout Statistics:")
	fmt.Printf("  Root Size: %dx%d\n", root.Box.Width, root.Box.Height)
	fmt.Printf("  Root Position: (%d, %d)\n", root.Box.X, root.Box.Y)
	fmt.Printf("  Total Nodes: %d\n", totalNodes)
	fmt.Printf("  Max Depth: %d\n", totalDepth)
}

// printLayoutTree 递归打印布局树
func printLayoutTree(box *compute_engine.ComputedBox, depth int) {
	if box == nil {
		return
	}

	// 缩进
	indent := strings.Repeat("  ", depth)

	// 节点信息
	nodeType := getNodeType(box)
	nodeID := box.VNode.Type().String()
	
	// 打印节点
	fmt.Printf("%s├─ [%s] %s\n", indent, nodeType, nodeID)
	fmt.Printf("%s│  Size: %dx%d\n", indent, box.Box.Width, box.Box.Height)
	fmt.Printf("%s│  Position: (%d, %d)\n", indent, box.Box.X, box.Box.Y)

	// 打印子节点
	if len(box.Children) > 0 {
		for i, child := range box.Children {
			if i == len(box.Children)-1 {
				printLayoutTreeLast(child, depth+1, true)
			} else {
				printLayoutTree(child, depth+1)
			}
		}
	} else {
		fmt.Printf("%s│  Children: 0 (leaf node)\n", indent)
	}
}

// printLayoutTreeLast 打印最后一个子节点（使用└而不是├）
func printLayoutTreeLast(box *compute_engine.ComputedBox, depth int, isLast bool) {
	if box == nil {
		return
	}

	indent := strings.Repeat("  ", depth)
	prefix := "└─"
	if isLast {
		prefix = "└─"
	} else {
		prefix = "├─"
	}

	nodeType := getNodeType(box)
	nodeID := box.VNode.Type().String()
	
	fmt.Printf("%s%s [%s] %s\n", indent, prefix, nodeType, nodeID)
	fmt.Printf("%s  │  Size: %dx%d\n", indent, box.Box.Width, box.Box.Height)
	fmt.Printf("%s  │  Position: (%d, %d)\n", indent, box.Box.X, box.Box.Y)

	if len(box.Children) > 0 {
		for i, child := range box.Children {
			if i == len(box.Children)-1 {
				printLayoutTreeLast(child, depth+1, true)
			} else {
				printLayoutTree(child, depth+1)
			}
		}
	} else {
		fmt.Printf("%s  │  Children: 0 (leaf node)\n", indent)
	}
}

// printLayoutGrid 打印布局网格可视化
func printLayoutGrid(root *compute_engine.ComputedBox, width, height int) {
	// 创建空网格
	grid := make([][]rune, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]rune, width)
		for x := 0; x < width; x++ {
			grid[y][x] = ' '
		}
	}

	// 填充节点到网格
	fillGridWithNodes(root, grid, 0, 0, 0)

	// 打印网格
	printGrid(grid)
}

// fillGridWithNodes 递归填充节点到网格
// 采用分层绘制：先绘制叶子节点内容，再绘制容器边框
func fillGridWithNodes(box *compute_engine.ComputedBox, grid [][]rune, offsetX, offsetY int, depth int) {
	if box == nil || box.Box.Width == 0 || box.Box.Height == 0 {
		return
	}

	// 计算节点在网格中的位置
	x := box.Box.X + offsetX
	y := box.Box.Y + offsetY
	w := box.Box.Width
	h := box.Box.Height

	if len(box.Children) > 0 {
		// 父容器：先递归绘制子节点，再绘制边框
		for _, child := range box.Children {
			fillGridWithNodes(child, grid, offsetX, offsetY, depth+1)
		}
		// 绘制边框（只在空位置绘制，保留子节点内容）
		drawBoxBorder(grid, x, y, w, h, depth)
	} else {
		// 叶子节点：填充内容区域
		marker := getBoxMarker(box)
		for dy := 0; dy < h && y+dy < len(grid); dy++ {
			for dx := 0; dx < w && x+dx < len(grid[y+dy]); dx++ {
				grid[y+dy][x+dx] = marker
			}
		}
	}
}

// drawBoxBorder 绘制容器边框（只在空位置绘制，保留已有内容）
func drawBoxBorder(grid [][]rune, x, y, w, h int, depth int) {
	// 根据深度选择边框字符
	var cornerH, cornerL, cornerR, cornerU, horiz, vert rune
	switch depth % 3 {
	case 0:
		cornerH, cornerL, cornerR, cornerU = '┌', '└', '┐', '┘'
		horiz, vert = '─', '│'
	case 1:
		cornerH, cornerL, cornerR, cornerU = '╭', '╰', '╮', '╯'
		horiz, vert = '─', '│'
	case 2:
		cornerH, cornerL, cornerR, cornerU = '╔', '╚', '╗', '╝'
		horiz, vert = '═', '║'
	}

	maxY := len(grid)
	maxX := len(grid[0])

	// 辅助函数：只在空位置绘制
	drawIfEmpty := func(gridY, gridX int, char rune) {
		if gridY >= 0 && gridY < maxY && gridX >= 0 && gridX < maxX {
			if grid[gridY][gridX] == ' ' || grid[gridY][gridX] == 0 {
				grid[gridY][gridX] = char
			}
		}
	}

	if w < 2 || h < 2 {
		// 太小的区域用简单填充（但不覆盖已有内容）
		for dy := 0; dy < h; dy++ {
			for dx := 0; dx < w; dx++ {
				drawIfEmpty(y+dy, x+dx, '+')
			}
		}
		return
	}

	// 绘制四个角
	drawIfEmpty(y, x, cornerH)
	drawIfEmpty(y, x+w-1, cornerR)
	drawIfEmpty(y+h-1, x, cornerL)
	drawIfEmpty(y+h-1, x+w-1, cornerU)

	// 绘制上边框
	for dx := 1; dx < w-1; dx++ {
		drawIfEmpty(y, x+dx, horiz)
	}

	// 绘制下边框
	for dx := 1; dx < w-1; dx++ {
		drawIfEmpty(y+h-1, x+dx, horiz)
	}

	// 绘制左边框
	for dy := 1; dy < h-1; dy++ {
		drawIfEmpty(y+dy, x, vert)
	}

	// 绘制右边框
	for dy := 1; dy < h-1; dy++ {
		drawIfEmpty(y+dy, x+w-1, vert)
	}
}

// printGrid 打印网格
func printGrid(grid [][]rune) {
	if len(grid) == 0 {
		return
	}

	// 打印顶部边框
	fmt.Print("  ┌")
	for x := 0; x < len(grid[0]); x++ {
		fmt.Print("─")
	}
	fmt.Println("┐")

	// 打印每一行
	for y := 0; y < len(grid); y++ {
		fmt.Printf("  │")
		for x := 0; x < len(grid[y]); x++ {
			fmt.Printf("%c", grid[y][x])
		}
		fmt.Println("│")
	}

	// 打印底部边框
	fmt.Print("  └")
	for x := 0; x < len(grid[0]); x++ {
		fmt.Print("─")
	}
	fmt.Println("┘")
}

// printNodeDetails 打印节点详细信息
func printNodeDetails(box *compute_engine.ComputedBox, depth int) {
	if box == nil {
		return
	}

	indent := strings.Repeat("  ", depth)

	fmt.Printf("%s[Node #%d]\n", indent, depth+1)
	fmt.Printf("%sType: %s\n", indent, box.VNode.Type().String())
	fmt.Printf("%sSize: %dx%d\n", indent, box.Box.Width, box.Box.Height)
	fmt.Printf("%sPosition: (%d, %d)\n", indent, box.Box.X, box.Box.Y)
	fmt.Printf("%sBounds: (%d, %d, %d, %d)\n", indent,
		box.Box.X, box.Box.Y, box.Box.X+box.Box.Width, box.Box.Y+box.Box.Height)
	fmt.Printf("%sChildren: %d\n", indent, len(box.Children))

	// 递归打印子节点
	for _, child := range box.Children {
		printNodeDetails(child, depth+1)
	}
}

// getNodeType 获取节点类型
func getNodeType(box *compute_engine.ComputedBox) string {
	if box == nil {
		return "Unknown"
	}

	typeStr := box.VNode.Type().String()
	
	// 简化类型名称
	switch {
	case strings.Contains(typeStr, "border"):
		return "Border"
	case strings.Contains(typeStr, "vstack"):
		return "VStack"
	case strings.Contains(typeStr, "hstack"):
		return "HStack"
	case strings.Contains(typeStr, "flex"):
		return "Flex"
	case strings.Contains(typeStr, "text"):
		return "Text"
	case strings.Contains(typeStr, "element"):
		return "Element"
	default:
		return typeStr
	}
}

// getBoxMarker 获取节点的网格标记
func getBoxMarker(box *compute_engine.ComputedBox) rune {
	if box == nil {
		return ' '
	}

	nodeType := getNodeType(box)
	
	// 为不同类型使用不同字符
	switch nodeType {
	case "Border":
		return '#'
	case "VStack":
		return 'V'
	case "HStack":
		return 'H'
	case "Flex":
		return 'F'
	case "Text":
		return 'T'
	case "Element":
		return 'E'
	default:
		return '·'
	}
}

// countNodes 递归统计节点数
func countNodes(box *compute_engine.ComputedBox) int {
	if box == nil {
		return 0
	}
	count := 1
	for _, child := range box.Children {
		count += countNodes(child)
	}
	return count
}

// maxDepth 计算最大深度
func maxDepth(box *compute_engine.ComputedBox) int {
	if box == nil {
		return 0
	}
	if len(box.Children) == 0 {
		return 1
	}
	maxChildDepth := 0
	for _, child := range box.Children {
		depth := maxDepth(child)
		if depth > maxChildDepth {
			maxChildDepth = depth
		}
	}
	return maxChildDepth + 1
}

// printBox 打印单个box的ASCII表示
func printBox(box *compute_engine.ComputedBox, label string) {
	if box == nil {
		return
	}

	w := box.Box.Width
	h := box.Box.Height

	// 顶部边框
	fmt.Printf("┌")
	for i := 0; i < w; i++ {
		if i == 0 || i == w-1 {
			fmt.Printf("─")
		} else if i == w/2 && len(label) > 0 {
			// 在中间显示标签
			labelLen := len(label)
			if labelLen <= w-2 {
				spaces := (w - 2 - labelLen) / 2
				fmt.Printf("%s%s%s", strings.Repeat(" ", spaces), label, strings.Repeat(" ", w-2-spaces-labelLen))
				for j := i + 1; j < w-1; j++ {
					fmt.Printf(" ")
				}
				continue
			}
		}
		fmt.Printf("─")
	}
	fmt.Println("┐")

	// 内容区域
	for i := 0; i < h; i++ {
		fmt.Printf("│")
		for j := 0; j < w; j++ {
			fmt.Printf(" ")
		}
		fmt.Println("│")
	}

	// 底部边框
	fmt.Printf("└")
	for i := 0; i < w; i++ {
		fmt.Printf("─")
	}
	fmt.Println("┘")

	// 尺寸信息
	fmt.Printf("  %dx%d @ (%d,%d)\n", w, h, box.Box.X, box.Box.Y)
}
