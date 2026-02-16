package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/examples/component_fixtures"
	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime"
	compute_engine "github.com/wwsheng009/mint/runtime/compute"
)

// 本程序显示布局数据并检测冲突
func main() {
	fmt.Println("=== Layout Data and Conflict Detection ===\n")

	fixtures := component_fixtures.StandardFixtures()

	for _, fixture := range fixtures {
		fmt.Printf("\n")
		fmt.Printf("Fixture: %s\n", fixture.Name)
		fmt.Printf("Description: %s\n", fixture.Description)
		fmt.Printf("====================================\n")

		// 构建VNode
		vnode := fixture.Build()
		if vnode == nil {
			fmt.Println("❌ Failed to build VNode")
			continue
		}

		// 创建Fiber
		fiber := reconciler.CreateFiberFromVNode(vnode)
		if fiber == nil {
			fmt.Println("❌ Failed to create Fiber")
			continue
		}

		// 执行布局
		engine := compute_engine.NewEngine()
		constraints := runtime.BoxConstraints{
			MinWidth:  80,
			MaxWidth: 80,
			MinHeight: 24,
			MaxHeight: 24,
		}

		layout, err := engine.Layout(vnode, fiber, constraints)
		if err != nil {
			fmt.Printf("❌ Layout failed: %v\n", err)
			continue
		}

		if layout == nil || layout.Root == nil {
			fmt.Println("❌ Layout result is nil")
			continue
		}

		// 显示布局数据
		displayLayoutData(layout)

		// 检测冲突
		conflicts := detectConflicts(layout, constraints)
		displayConflicts(fixture.Name, conflicts)
	}

	fmt.Println("\n=== All Tests Complete ===")
}

// displayLayoutData 显示布局数据
func displayLayoutData(layout *compute_engine.ComputedLayout) {
	if layout == nil || layout.Root == nil {
		return
	}

	fmt.Println("--- Layout Data ---")

	// 根节点信息
	root := layout.Root
	fmt.Printf("Root Node:\n")
	fmt.Printf("  ID: %s\n", root.VNode.Type().String())
	fmt.Printf("  Size: %dx%d\n", root.Box.Width, root.Box.Height)
	fmt.Printf("  Position: (%d,%d)\n", root.Box.X, root.Box.Y)
	fmt.Printf("  Bounds: (%d,%d,%d,%d)\n",
		root.Box.X, root.Box.Y,
		root.Box.X+root.Box.Width,
		root.Box.Y+root.Box.Height)

	// 统计信息
	stats := collectStatistics(layout.Root)
	fmt.Printf("\nStatistics:\n")
	fmt.Printf("  Total Nodes: %d\n", stats.totalNodes)
	fmt.Printf("  Max Depth: %d\n", stats.maxDepth)

	// 节点列表
	fmt.Println("\n--- All Nodes (%d) ---", len(stats.nodes))
	displayNodesWithTree(stats.nodes, 0)

	// ASCII可视化
	fmt.Println("\n--- ASCII Grid Visualization (80x24) ---")
	displayASCIGrid(stats.nodes, 80, 24)
}

// displayNodesWithTree 显示节点和树形结构
func displayNodesWithTree(nodes []NodeStats, depth int) {
	for _, node := range nodes {
		displayNodeWithIndent(node, depth, true)
	}
}

// displayNodeWithIndent 显示带缩进的节点信息
func displayNodeWithIndent(node NodeStats, depth int, isLast bool) {
	indent := strings.Repeat("  ", depth)
	prefix := "├─"
	if isLast {
		prefix = "└─"
	} else if depth == 0 {
		prefix = "┌─"
	}

	// 节点信息
	fmt.Printf("%s%s [%s] %s\n", indent, prefix, node.NodeType, node.NodeID)
	fmt.Printf("%s│  Size: %dx%d\n", indent, node.Width, node.Height)
	fmt.Printf("%s│  Position: (%d,%d)\n", indent, node.X, node.Y)
	fmt.Printf("%s│  Content: %s\n", indent, node.Content)

	// 显示子节点
	if len(node.Children) == 0 {
		fmt.Printf("%s│  Children: 0 (leaf node)\n", indent)
		return
	}

	// 显示子节点
	fmt.Printf("%s│  Children: %d\n", indent, len(node.Children))
	for i, child := range node.Children {
		displayNodeWithIndent(child, depth+1, i == len(node.Children)-1)
	}
}

// displayASCIGrid 显示80x24的ASCII网格
func displayASCIGrid(nodes []NodeStats, width, height int) {
	// 创建空网格
	grid := make([][]rune, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]rune, width)
		for x := 0; x < width; x++ {
			grid[y][x] = ' '
		}
	}

	// 填充节点到网格
	fillGridWithNodes(nodes, grid, 0, 0, height, 0)

	// 打印网格
	fmt.Print("  ┌")
	for i := 0; i < width; i++ {
		fmt.Print("─")
	}
	fmt.Println("┐")

	for y := 0; y < height; y++ {
		fmt.Print("  │")
		for x := 0; x < width; x++ {
			fmt.Printf("%c", grid[y][x])
		}
		fmt.Println("│")
	}

	// 底部边框
	fmt.Print("  └")
	for i := 0; i < width; i++ {
		fmt.Print("─")
	}
	fmt.Println("┘")
}

// fillGridWithNodes 将节点填入网格（绘制父容器边框，子组件在内部）
// 采用分层绘制：先绘制叶子节点内容，再绘制容器边框（边框叠加在内容上）
func fillGridWithNodes(nodes []NodeStats, grid [][]rune, offsetX, offsetY int, maxHeight int, depth int) {
	for _, node := range nodes {
		if node.Width == 0 || node.Height == 0 {
			continue
		}

		// 计算节点在网格中的位置
		x := node.X + offsetX
		y := node.Y + offsetY

		if len(node.Children) > 0 {
			// 父容器：先递归绘制子节点，再绘制边框（边框叠加在子节点上）
			fillGridWithNodes(node.Children, grid, offsetX, offsetY, maxHeight, depth+1)
			// 绘制边框（会覆盖部分子节点内容，但保留可见性）
			drawBorder(grid, x, y, node.Width, node.Height, maxHeight, depth)
		} else {
			// 叶子节点：填充内容区域
			marker := getNodeType(node.NodeType)
			fillContentArea(grid, x, y, node.Width, node.Height, maxHeight, marker)
		}
	}
}

// fillContentArea 填充内容区域
func fillContentArea(grid [][]rune, x, y, width, height, maxHeight int, marker rune) {
	maxX := len(grid[0])
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			gridY := y + dy
			gridX := x + dx
			if gridY >= 0 && gridY < maxHeight && gridX >= 0 && gridX < maxX {
				grid[gridY][gridX] = marker
			}
		}
	}
}

// drawBorder 绘制边框（使用Unicode box-drawing字符，depth用于不同层级样式）
// 只在空位置绘制边框，保留已有内容（子节点）
func drawBorder(grid [][]rune, x, y, width, height, maxHeight int, depth int) {
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

	maxX := len(grid[0])

	// 辅助函数：只在空位置绘制
	drawIfEmpty := func(gridY, gridX int, char rune) {
		if gridY >= 0 && gridY < maxHeight && gridX >= 0 && gridX < maxX {
			if grid[gridY][gridX] == ' ' || grid[gridY][gridX] == 0 {
				grid[gridY][gridX] = char
			}
		}
	}

	if width < 2 || height < 2 {
		// 太小的区域用简单填充（但不覆盖已有内容）
		for dy := 0; dy < height; dy++ {
			for dx := 0; dx < width; dx++ {
				drawIfEmpty(y+dy, x+dx, '+')
			}
		}
		return
	}

	// 绘制四个角
	drawIfEmpty(y, x, cornerH)
	drawIfEmpty(y, x+width-1, cornerR)
	drawIfEmpty(y+height-1, x, cornerL)
	drawIfEmpty(y+height-1, x+width-1, cornerU)

	// 绘制上边框
	for dx := 1; dx < width-1; dx++ {
		drawIfEmpty(y, x+dx, horiz)
	}

	// 绘制下边框
	for dx := 1; dx < width-1; dx++ {
		drawIfEmpty(y+height-1, x+dx, horiz)
	}

	// 绘制左边框
	for dy := 1; dy < height-1; dy++ {
		drawIfEmpty(y+dy, x, vert)
	}

	// 绘制右边框
	for dy := 1; dy < height-1; dy++ {
		drawIfEmpty(y+dy, x+width-1, vert)
	}
}

// getNodeType 根据节点类型返回标记字符
func getNodeType(nodeType string) rune {
	switch nodeType {
	case "Element", "element":
		return 'E'
	case "Text", "text":
		return 'T'
	case "VStack", "vstack":
		return 'V'
	case "HStack", "hstack":
		return 'H'
	case "Flex", "flex":
		return 'F'
	case "Bordered", "border":
		return '#'
	default:
		return '·'
	}
}

// collectStatistics 收集统计信息并构建树结构
func collectStatistics(root *compute_engine.ComputedBox) LayoutStatistics {
	stats := LayoutStatistics{
		totalNodes: 0,
		maxDepth:   0,
		nodes:      []NodeStats{},
	}

	var traverse func(box *compute_engine.ComputedBox, depth int) NodeStats
	traverse = func(box *compute_engine.ComputedBox, depth int) NodeStats {
		if box == nil {
			return NodeStats{}
		}

		stats.totalNodes++
		if depth > stats.maxDepth {
			stats.maxDepth = depth
		}

		// 先递归处理子节点，构建子节点列表
		children := make([]NodeStats, 0, len(box.Children))
		for _, child := range box.Children {
			childStats := traverse(child, depth+1)
			children = append(children, childStats)
		}

		// 创建当前节点（包含子节点）
		node := NodeStats{
			NodeType:  box.VNode.Type().String(),
			NodeID:    box.VNode.Type().String() + "@" + fmt.Sprintf("%d,%d", box.Box.X, box.Box.Y),
			X:         box.Box.X,
			Y:         box.Box.Y,
			Width:     box.Box.Width,
			Height:    box.Box.Height,
			MinX:      box.Box.X,
			MinY:      box.Box.Y,
			MaxX:      box.Box.X + box.Box.Width,
			MaxY:      box.Box.Y + box.Box.Height,
			Content:   getNodeContent(box),
			Children:  children,
			Depth:     depth,
		}

		return node
	}

	// 构建树结构，根节点放在nodes列表中
	if root != nil {
		rootNode := traverse(root, 0)
		stats.nodes = []NodeStats{rootNode}
	}

	return stats
}

// collectAllNodes 收集所有节点到列表
func collectAllNodes(root *compute_engine.ComputedBox) []NodeStats {
	var nodes []NodeStats

	var traverse func(box *compute_engine.ComputedBox)
	traverse = func(box *compute_engine.ComputedBox) {
		if box == nil {
			return
		}
		nodes = append(nodes, NodeStats{
			NodeType: box.VNode.Type().String(),
			NodeID:    box.VNode.Type().String() + "@" + fmt.Sprintf("%d,%d", box.Box.X, box.Box.Y),
			X:          box.Box.X,
			Y:          box.Box.Y,
			Width:       box.Box.Width,
			Height:      box.Box.Height,
			Content:     getNodeContent(box),
			Children:    []NodeStats{},
			Depth:      0,
		})
		for _, child := range box.Children {
			traverse(child)
		}
	}

	traverse(root)
	return nodes
}

// getNodeContent 获取节点的内容文本
func getNodeContent(box *compute_engine.ComputedBox) string {
	if box == nil || box.VNode == nil {
		return ""
	}

	nodeType := box.VNode.Type().String()

	// 尝试获取文本内容
	if textGetter, ok := box.VNode.(interface{ GetText() string }); ok {
		return textGetter.GetText()
	}

	// 尝试获取标签
	if labelGetter, ok := box.VNode.(interface{ Label() string }); ok {
		return labelGetter.Label()
	}

	// 根据类型返回示例内容
	switch nodeType {
	case "text":
		return "Text content"
	case "element":
		return "Element node"
	case "vstack":
		return "Vertical stack"
	case "hstack":
		return "Horizontal stack"
	case "flex":
		return "Flex container"
	case "border":
		return "Bordered container"
	default:
		return "Unknown node"
	}
}

// detectConflicts 检测布局冲突
func detectConflicts(layout *compute_engine.ComputedLayout, constraints runtime.BoxConstraints) []Conflict {
	if layout == nil || layout.Root == nil {
		return nil
	}

	var conflicts []Conflict
	nodes := collectAllNodes(layout.Root)

	// 1. 检测重叠冲突
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			node1 := nodes[i]
			node2 := nodes[j]

			// 跳过自身和父子关系
			if node1.NodeID == node2.NodeID {
				continue
			}
			if isDescendantOf(node1, node2) || isDescendantOf(node2, node1) {
				continue
			}

			// 检查重叠
			if checkOverlap(node1, node2) {
				conflict := Conflict{
					Type:     "Overlap",
					Nodes:     []string{node1.NodeID, node2.NodeID},
					Node1:     node1,
					Node2:     node2,
					Position:  fmt.Sprintf("(%d,%d)", node1.X, node1.Y),
					Details:   fmt.Sprintf("%s overlaps with %s",
						node1.NodeType, node1.Content, node2.NodeType, node2.Content),
					Severity:  "Warning",
				}
				conflicts = append(conflicts, conflict)
			}
		}
	}

	// 2. 检测边界冲突
	for _, node := range nodes {
		boundaryConflicts := checkBoundaryViolation(node, constraints)
		conflicts = append(conflicts, boundaryConflicts...)
	}

	// 3. 检测约束违反
	for _, node := range nodes {
		constraintConflicts := checkConstraintViolation(node, constraints)
		conflicts = append(conflicts, constraintConflicts...)
	}

	return conflicts
}

// checkBoundaryViolation 检查边界违反
func checkBoundaryViolation(node NodeStats, constraints runtime.BoxConstraints) []Conflict {
	var conflicts []Conflict

	maxW := constraints.MaxWidth
	maxH := constraints.MaxHeight

	// 检查是否超出右边界
	if maxW < runtime.Infinity && node.X+node.Width > maxW {
		conflicts = append(conflicts, Conflict{
			Type:     "Boundary",
			Nodes:     []string{node.NodeID},
			Position:  fmt.Sprintf("Right: %d", maxW),
			Details:   fmt.Sprintf("%s exceeds max width boundary (x=%d, width=%d)",
				node.NodeType, node.X+node.Width, maxW),
			Severity: "Warning",
		})
	}

	// 检查是否超出底部边界
	if maxH < runtime.Infinity && node.Y+node.Height > maxH {
		conflicts = append(conflicts, Conflict{
			Type:     "Boundary",
			Nodes:     []string{node.NodeID},
			Position: fmt.Sprintf("Bottom: %d", maxH),
			Details:   fmt.Sprintf("%s exceeds max height boundary (y=%d, height=%d)",
				node.NodeType, node.Y+node.Height, maxH),
			Severity: "Warning",
		})
	}

	// 检查是否超出左边界
	if node.X < 0 {
		conflicts = append(conflicts, Conflict{
			Type:     "Boundary",
			Nodes:     []string{node.NodeID},
			Position: fmt.Sprintf("Left: 0"),
			Details:   fmt.Sprintf("%s has negative X coordinate (x=%d)",
				node.NodeType, node.X),
			Severity:  "Info",
		})
	}

	// 检查是否超出顶部边界
	if node.Y < 0 {
		conflicts = append(conflicts, Conflict{
			Type:     "Boundary",
			Nodes:     []string{node.NodeID},
			Position: fmt.Sprintf("Top: 0"),
			Details:   fmt.Sprintf("%s has negative Y coordinate (y=%d)",
				node.NodeType, node.Y),
			Severity:  "Info",
		})
	}

	return conflicts
}

// checkConstraintViolation 检查约束违反
func checkConstraintViolation(node NodeStats, constraints runtime.BoxConstraints) []Conflict {
	var conflicts []Conflict

	// 检查最小宽度
	if constraints.MinWidth > 0 && node.Width < constraints.MinWidth {
		conflicts = append(conflicts, Conflict{
			Type:     "ConstraintViolation",
			Nodes:     []string{node.NodeID},
			Position: fmt.Sprintf("Width: %d", node.Width),
			Details:   fmt.Sprintf("%s violates minimum width constraint (%d < %d)",
				node.NodeType, node.Width, constraints.MinWidth),
			Severity: "Info",
		})
	}

	// 检查最大宽度
	if constraints.MaxWidth < runtime.Infinity && node.Width > constraints.MaxWidth {
		conflicts = append(conflicts, Conflict{
			Type:     "ConstraintViolation",
			Nodes:     []string{node.NodeID},
			Position: fmt.Sprintf("Width: %d", node.Width),
			Details:   fmt.Sprintf("%s exceeds maximum width constraint (%d > %d)",
				node.NodeType, node.Width, constraints.MaxWidth),
			Severity:  "Info",
		})
	}

	// 检查最小高度
	if constraints.MinHeight > 0 && node.Height < constraints.MinHeight {
		conflicts = append(conflicts, Conflict{
			Type:     "ConstraintViolation",
			Nodes:     []string{node.NodeID},
			Position: fmt.Sprintf("Height: %d", node.Height),
			Details: fmt.Sprintf("%s violates minimum height constraint (%d < %d)",
				node.NodeType, node.Height, constraints.MinHeight),
			Severity: "Info",
		})
	}

	// 检查最大高度
	if constraints.MaxHeight < runtime.Infinity && node.Height > constraints.MaxHeight {
		conflicts = append(conflicts, Conflict{
			Type:     "ConstraintViolation",
			Nodes:     []string{node.NodeID},
			Position: fmt.Sprintf("Height: %d", node.Height),
			Details: fmt.Sprintf("%s exceeds maximum height constraint (%d > %d)",
				node.NodeType, node.Height, constraints.MaxHeight),
			Severity: "Info",
		})
	}

	return conflicts
}

// checkOverlap 检查两个节点是否重叠
func checkOverlap(node1, node2 NodeStats) bool {
	x1, y1 := node1.X, node1.Y
	w1, h1 := node1.Width, node1.Height
	x2, y2 := node2.X, node2.Y
	w2, h2 := node2.Width, node2.Height

	// 检查x轴重叠
	overlapX := x1 < x2+w2 && x1+w1 > x2

	// 检查y轴重叠
	overlapY := y1 < y2+h2 && y1+h1 > y2

	return overlapX && overlapY
}

// isDescendantOf 检查node2是否是node1的子孙节点
func isDescendantOf(node1, node2 NodeStats) bool {
	// 检查node2是否在node1的子树中
	return isInSubtree(node1, node2)
}

// isInSubtree 检查node2是否在node1的子树中
func isInSubtree(parent, child NodeStats) bool {
	// 递归检查child是否在parent的子树中
	for _, c := range parent.Children {
		if c.NodeID == child.NodeID {
			return true
		}
		if isInSubtree(c, child) {
			return true
		}
	}

	return false
}

// displayConflicts 显示冲突信息
func displayConflicts(fixtureName string, conflicts []Conflict) {
	fmt.Println("--- Conflict Detection ---")

	if len(conflicts) == 0 {
		fmt.Println("✅ No conflicts detected")
		return
	}

	// 按严重性分组
	errors := []Conflict{}
	warnings := []Conflict{}
	infos := []Conflict{}

	for _, conflict := range conflicts {
		switch conflict.Severity {
		case "Error":
			errors = append(errors, conflict)
		case "Warning":
			warnings = append(warnings, conflict)
		case "Info":
			infos = append(infos, conflict)
		}
	}

	fmt.Printf("Total Conflicts: %d\n", len(conflicts))

	if len(errors) > 0 {
		fmt.Printf("\n❌ Errors (%d):\n", len(errors))
		for _, e := range errors {
			fmt.Printf("  [%s] %s\n", e.Type, e.Position)
			fmt.Printf("    %s\n", e.Details)
		}
	}

	if len(warnings) > 0 {
		fmt.Printf("\n⚠️  Warnings (%d):\n", len(warnings))
		for _, w := range warnings {
			fmt.Printf("  [%s] %s\n", w.Type, w.Position)
			fmt.Printf("    %s\n", w.Details)
		}
	}

	if len(infos) > 0 {
		fmt.Printf("\nℹ️  Infos (%d):\n", len(infos))
		for _, i := range infos {
			fmt.Printf("  [%s] %s\n", i.Type, i.Position)
			fmt.Printf("    %s\n", i.Details)
		}
	}

	// 按类型分组
	fmt.Println("\n--- Conflict Summary by Type ---")
	typeCounts := make(map[string]int)
	for _, c := range conflicts {
		typeCounts[c.Type]++
	}
	for ct, count := range typeCounts {
		fmt.Printf("  %s: %d\n", ct, count)
	}
}

	// NodeStats 节点数据结构
type NodeStats struct {
	NodeID   string    // 节点唯一ID
	NodeType string    // 节点类型（Element, Text, VStack等）
	X, Y   int       // 位置
	Width, Height int       // 尺寸
	MinX, MinY, MaxX, MaxY int  // 边界（最小/最大X/Y）
	Content string        // 内容文本
	Children []NodeStats // 子节点
	Depth   int          // 深度
}

// LayoutStatistics 布局统计
type LayoutStatistics struct {
	totalNodes int
	maxDepth   int
	nodes      []NodeStats
}

// Conflict 冲突信息
type Conflict struct {
	Type     string // 冲突类型
	Nodes     []string // 涉及的节点ID
	Node1     NodeStats // 节点1（用于重叠）
	Node2     NodeStats // 节点2（用于重叠）
	Position  string  // 冲突位置描述
	Details   string // 详细描述
	Severity  string // 严重性：Error/Warning/Info
}
