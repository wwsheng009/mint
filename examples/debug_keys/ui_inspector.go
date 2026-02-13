package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/internal/reconciler"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// KeyInspectorUI - UI Inspector，在界面中显示所有层次的 KEY 信息
func main() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🔑 UI Key Inspector - 交互式调试工具")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("\n这个工具会在 UI 中以 inspector 的形式显示所有 nodes 的 KEY 信息")
	fmt.Println("按 'q' 退出应用\n")

	// 状态管理
	var modalOpen bool
	var overlayVisible bool
	var inspectorEnabled bool = true
	var showKeys bool = true
	var showPaths bool = true
	var showLayers bool = true

	ui.Run(func() ui.VNode {
		// Base layer content
		baseContent := app.VStack(
			app.NewTextBuilder("🔑 UI Key Inspector 演示").Bold(true).FgColor("cyan").Build(),
			app.Text(""),
			app.NewTextBuilder("点击按钮打开不同的 layer，观察 Inspector 中的 KEY 变化").FgColor("gray").Build(),
			app.Text(""),
			app.ButtonBuilder("打开 Modal").
				OnClick(func() {
					modalOpen = true
				}).
				Build(),
			app.ButtonBuilder("显示 Overlay").
				OnClick(func() {
					overlayVisible = true
				}).
				Build(),
			app.ButtonBuilder("切换 Inspector").
				OnClick(func() {
					inspectorEnabled = !inspectorEnabled
				}).
				Build(),
			app.ButtonBuilder("关闭所有 Layers").
				OnClick(func() {
					modalOpen = false
					overlayVisible = false
				}).
				Build(),
			app.Text(""),
			app.NewTextBuilder("─────────────────────────────────").FgColor("gray").Build(),
			app.NewTextBuilder("显示选项:").FgColor("cyan").Build(),
			app.Text(""),
			buildCheckbox("显示 Keys", showKeys, func() {
				showKeys = !showKeys
			}),
			buildCheckbox("显示 Paths", showPaths, func() {
				showPaths = !showPaths
			}),
			buildCheckbox("显示 Layer 标记", showLayers, func() {
				showLayers = !showLayers
			}),
		)

		var modalContent ui.VNode
		if modalOpen {
			modalContent = app.VStack(
				app.NewTextBuilder("这是 Modal (LayerModal)").FgColor("cyan").Build(),
				app.NewTextBuilder("观察 Inspector 中这个节点的 KEY").FgColor("gray").Build(),
				app.Text(""),
				app.ButtonBuilder("Modal 内部的按钮").
					OnClick(func() {
						fmt.Println("Modal button clicked!")
					}).
					Build(),
				app.ButtonBuilder("关闭 Modal").
					OnClick(func() {
						modalOpen = false
					}).
					Build(),
			)
		}

		var overlayContent ui.VNode
		if overlayVisible {
			overlayContent = app.VStack(
				app.NewTextBuilder("这是 Overlay (LayerOverlay)").FgColor("yellow").Build(),
				app.NewTextBuilder("观察 Inspector 中这个节点的 KEY").FgColor("gray").Build(),
				app.Text(""),
				app.ButtonBuilder("Overlay 按钮").
					OnClick(func() {
						fmt.Println("Overlay button clicked!")
					}).
					Build(),
			)
		}

		// 创建 Inspector overlay
		var inspectorContent ui.VNode
		if inspectorEnabled {
			inspectorContent = createInspectorOverlay(showKeys, showPaths, showLayers)
		}

		// 组合所有 children
		children := []ui.VNode{baseContent}
		if modalContent != nil {
			children = append(children, modalContent)
		}
		if overlayContent != nil {
			children = append(children, overlayContent)
		}
		if inspectorContent != nil {
			children = append(children, inspectorContent)
		}

		return app.VStack(children...)
	},
		ui.WithWidth(80),
		ui.WithHeight(40),
		ui.WithTitle("UI Key Inspector"),
	)
}

// buildCheckbox 创建一个简单的 checkbox (使用 text + button 模拟)
func buildCheckbox(label string, checked bool, onClick func()) ui.VNode {
	var status string
	if checked {
		status = "[X] "
	} else {
		status = "[ ] "
	}

	return app.HStack(
		app.NewTextBuilder(status+label).Build(),
		app.ButtonBuilder("切换").
			OnClick(onClick).
			Build(),
	)
}

// createInspectorOverlay 创建 Inspector overlay，显示 KEY 信息
func createInspectorOverlay(showKeys, showPaths, showLayers bool) ui.VNode {
	// TODO: 这里应该获取当前的 VNode 和 Fiber 树
	// 由于当前架构限制，我们创建一个示例演示

	lines := []string{
		"══════════════════════════════════════════",
		"🔑 KEY INSPECTOR",
		"══════════════════════════════════════════",
		"",
		"示例 KEY 信息:",
		"│─ vstack",
		"  │─ text",
		"  │─ button",
		"  │─ vstack [MODAL]",
		"    │─ text",
		"    │─ button",
		"",
		fmt.Sprintf("显示 Keys: %v", showKeys),
		fmt.Sprintf("显示 Paths: %v", showPaths),
		fmt.Sprintf("显示 Layers: %v", showLayers),
		"",
		"注意: 这是一个演示 overlay",
		"要获取真实数据，需要访问",
		"reconciler 的内部状态",
		"══════════════════════════════════════════",
	}

	textNodes := make([]ui.VNode, len(lines))
	for i, line := range lines {
		textNodes[i] = app.Text(line)
	}

	return app.VStack(textNodes...)
}

// 以下是命令行版本的调试 API，可以独立使用
type DebugKeyInspector struct {
	MaxDepth   int
	ShowKeys   bool
	ShowPaths  bool
	ShowLayers bool
}

func NewDebugKeyInspector() *DebugKeyInspector {
	return &DebugKeyInspector{
		MaxDepth:   20,
		ShowKeys:   true,
		ShowPaths:  true,
		ShowLayers: true,
	}
}

// InspectVNodes 显示 VNode 树的所有 Key 信息
func (dki *DebugKeyInspector) InspectVNodes(vnode rtui.VNode) {
	fmt.Println("\n" + strings.Repeat("═", 80))
	fmt.Println("🔑 VNODE TREE - KEY INFORMATION")
	fmt.Println(strings.Repeat("═", 80))

	if vnode == nil {
		fmt.Println("❌ VNode is nil")
		return
	}

	count := dki.walkVNode(vnode, 0)
	fmt.Printf("\n✅ Total VNodes: %d\n\n", count)
}

func (dki *DebugKeyInspector) walkVNode(vnode rtui.VNode, depth int) int {
	if depth > dki.MaxDepth || vnode == nil {
		return 0
	}

	indent := strings.Repeat("  ", depth)

	var typeName, key string
	var layer rtui.Layer

	switch n := vnode.(type) {
	case *rtui.ElementVNode:
		typeName = n.Tag()
		key = n.Key()
		layer = n.GetLayer()
	case *rtui.TextVNode:
		typeName = "text"
		key = n.Key()
		layer = n.GetLayer()
	case *rtui.FragmentVNode:
		typeName = "fragment"
		key = n.Key()
	case *rtui.ComponentVNode:
		if nameable, ok := vnode.(interface{ Name() string }); ok {
			typeName = nameable.Name()
		} else {
			typeName = "component"
		}
		key = n.Key()
	}

	info := fmt.Sprintf("%s│─ %s", indent, typeName)

	if dki.ShowLayers && layer != rtui.LayerBase {
		info += fmt.Sprintf(" [%s]", getLayerName(layer))
	}

	if dki.ShowKeys {
		if key != "" {
			info += fmt.Sprintf(" \033[33mKey:%q\033[0m", key)
		} else {
			info += " \033[90m⚠️无Key\033[0m"
		}
	}

	fmt.Println(info)

	count := 1
	children := vnode.Children()
	for _, child := range children {
		count += dki.walkVNode(child, depth+1)
	}

	return count
}

// InspectFibers 显示 Fiber 树的所有 Path 和 Key 信息
func (dki *DebugKeyInspector) InspectFibers(fiber *reconciler.Fiber) {
	fmt.Println("\n" + strings.Repeat("═", 80))
	fmt.Println("🌳 FIBER TREE - PATH & KEY INFORMATION")
	fmt.Println(strings.Repeat("═", 80))

	if fiber == nil {
		fmt.Println("❌ Fiber is nil")
		return
	}

	count := dki.walkFiber(fiber, 0)
	fmt.Printf("\n✅ Total Fibers: %d\n\n", count)
}

func (dki *DebugKeyInspector) walkFiber(fiber *reconciler.Fiber, depth int) int {
	if depth > dki.MaxDepth || fiber == nil {
		return 0
	}

	indent := strings.Repeat("  ", depth)

	var typeName string
	if fiber.Type == rtui.VNodeElement {
		typeName = fiber.Tag
	} else if fiber.Type == rtui.VNodeComponent {
		if nameable, ok := fiber.VNode.(interface{ Name() string }); ok {
			typeName = nameable.Name()
		} else {
			typeName = "component"
		}
	} else if fiber.Type == rtui.VNodeText {
		typeName = "text"
	} else if fiber.Type == rtui.VNodeFragment {
		typeName = "fragment"
	} else {
		typeName = fmt.Sprintf("type(%d)", fiber.Type)
	}

	var layer rtui.Layer
	if fiber.VNode != nil {
		layer = fiber.VNode.GetLayer()
	}

	info := fmt.Sprintf("%s│─ %s", indent, typeName)

	if dki.ShowLayers && layer != rtui.LayerBase {
		info += fmt.Sprintf(" [%s]", getLayerName(layer))
	}

	if dki.ShowKeys && fiber.Key != "" {
		info += fmt.Sprintf(" \033[33mKey:%q\033[0m", fiber.Key)
	}

	if dki.ShowPaths && fiber.Path != "" {
		info += fmt.Sprintf(" \033[36mPath:%q\033[0m", fiber.Path)
	}

	if fiber.PathSegment != "" {
		info += fmt.Sprintf(" \033[35mSegment:%q\033[0m", fiber.PathSegment)
	}

	if fiber.Key != "" && fiber.Path != "" && fiber.Key != fiber.Path {
		info += " \033[31m⚠️Key≠Path\033[0m"
	}

	fmt.Println(info)

	count := 1
	child := fiber.Child
	for child != nil {
		count += dki.walkFiber(child, depth+1)
		child = child.Sibling
	}

	return count
}

// CompareTrees 比较 VNode 和 Fiber 树的 Key 一致性
func (dki *DebugKeyInspector) CompareTrees(vnode rtui.VNode, fiber *reconciler.Fiber) {
	fmt.Println("\n" + strings.Repeat("═", 80))
	fmt.Println("🔍 VNode vs Fiber - KEY CONSISTENCY CHECK")
	fmt.Println(strings.Repeat("═", 80))

	if vnode == nil || fiber == nil {
		fmt.Println("❌ Cannot compare: VNode or Fiber is nil")
		return
	}

	vnodeMap := make(map[string]int)
	fiberMap := make(map[string]int)

	dki.collectVNodeKeys(vnode, vnodeMap)
	dki.collectFiberKeys(fiber, fiberMap)

	fmt.Println("\n📊 Key Distribution:")
	fmt.Printf("  VNode unique keys: %d\n", len(vnodeMap))
	fmt.Printf("  Fiber unique keys: %d\n", len(fiberMap))

	fmt.Println("\n⚠️  Inconsistencies:")

	mismatches := 0
	for key, vnodeCount := range vnodeMap {
		fiberCount, exists := fiberMap[key]
		if !exists {
			fmt.Printf("  ❌ VNode Key %q exists %d times but not in Fibers\n", key, vnodeCount)
			mismatches++
		} else if vnodeCount != fiberCount {
			fmt.Printf("  ⚠️  Key %q: VNode=%d, Fiber=%d\n", key, vnodeCount, fiberCount)
			mismatches++
		}
	}

	for key, fiberCount := range fiberMap {
		if _, exists := vnodeMap[key]; !exists {
			fmt.Printf("  ❌ Fiber Key %q exists %d times but not in VNodes\n", key, fiberCount)
			mismatches++
		}
	}

	if mismatches == 0 {
		fmt.Println("  ✅ All keys are consistent!")
	} else {
		fmt.Printf("  ❌ Found %d inconsistencies\n", mismatches)
	}

	fmt.Println()
}

func (dki *DebugKeyInspector) collectVNodeKeys(vnode rtui.VNode, keyMap map[string]int) {
	if vnode == nil {
		return
	}

	key := vnode.Key()
	if key != "" {
		keyMap[key]++
	}

	children := vnode.Children()
	for _, child := range children {
		dki.collectVNodeKeys(child, keyMap)
	}
}

func (dki *DebugKeyInspector) collectFiberKeys(fiber *reconciler.Fiber, keyMap map[string]int) {
	if fiber == nil {
		return
	}

	if fiber.Key != "" {
		keyMap[fiber.Key]++
	}

	child := fiber.Child
	for child != nil {
		dki.collectFiberKeys(child, keyMap)
		child = child.Sibling
	}
}

// PrintStatistics 打印统计信息
func (dki *DebugKeyInspector) PrintStatistics(vnode rtui.VNode, fiber *reconciler.Fiber) {
	fmt.Println("\n" + strings.Repeat("═", 80))
	fmt.Println("📈 STATISTICS")
	fmt.Println(strings.Repeat("═", 80))

	if vnode != nil {
		vnodeStats := dki.analyzeVNodes(vnode)
		fmt.Println("\n📊 VNode Tree:")
		fmt.Printf("  Total nodes: %d\n", vnodeStats.total)
		fmt.Printf("  With key: %d (%.1f%%)\n", vnodeStats.withKey,
			float64(vnodeStats.withKey)/float64(vnodeStats.total)*100)
		fmt.Printf("  Without key: %d (%.1f%%)\n", vnodeStats.withoutKey,
			float64(vnodeStats.withoutKey)/float64(vnodeStats.total)*100)
		if vnodeStats.layerNodes > 0 {
			fmt.Printf("  Layer nodes: %d\n", vnodeStats.layerNodes)
		}
	}

	if fiber != nil {
		fiberStats := dki.analyzeFibers(fiber)
		fmt.Println("\n🌳 Fiber Tree:")
		fmt.Printf("  Total nodes: %d\n", fiberStats.total)
		fmt.Printf("  With key: %d (%.1f%%)\n", fiberStats.withKey,
			float64(fiberStats.withKey)/float64(fiberStats.total)*100)
		fmt.Printf("  Without key: %d (%.1f%%)\n", fiberStats.withoutKey,
			float64(fiberStats.withoutKey)/float64(fiberStats.total)*100)
		fmt.Printf("  With path: %d (%.1f%%)\n", fiberStats.withPath,
			float64(fiberStats.withPath)/float64(fiberStats.total)*100)
		if fiberStats.layerNodes > 0 {
			fmt.Printf("  Layer nodes: %d\n", fiberStats.layerNodes)
		}
	}

	fmt.Println()
}

type VNodeAnalysis struct {
	total      int
	withKey    int
	withoutKey int
	layerNodes int
}

func (dki *DebugKeyInspector) analyzeVNodes(vnode rtui.VNode) *VNodeAnalysis {
	stats := &VNodeAnalysis{}
	dki.walkVNodeAnalysis(vnode, stats)
	return stats
}

func (dki *DebugKeyInspector) walkVNodeAnalysis(vnode rtui.VNode, stats *VNodeAnalysis) {
	if vnode == nil {
		return
	}

	stats.total++
	if vnode.Key() != "" {
		stats.withKey++
	} else {
		stats.withoutKey++
	}
	if vnode.GetLayer() != rtui.LayerBase {
		stats.layerNodes++
	}

	children := vnode.Children()
	for _, child := range children {
		dki.walkVNodeAnalysis(child, stats)
	}
}

type FiberAnalysis struct {
	total      int
	withKey    int
	withoutKey int
	withPath   int
	layerNodes int
}

func (dki *DebugKeyInspector) analyzeFibers(fiber *reconciler.Fiber) *FiberAnalysis {
	stats := &FiberAnalysis{}
	dki.walkFiberAnalysis(fiber, stats)
	return stats
}

func (dki *DebugKeyInspector) walkFiberAnalysis(fiber *reconciler.Fiber, stats *FiberAnalysis) {
	if fiber == nil {
		return
	}

	stats.total++
	if fiber.Key != "" {
		stats.withKey++
	} else {
		stats.withoutKey++
	}
	if fiber.Path != "" {
		stats.withPath++
	}
	if fiber.VNode != nil && fiber.VNode.GetLayer() != rtui.LayerBase {
		stats.layerNodes++
	}

	child := fiber.Child
	for child != nil {
		dki.walkFiberAnalysis(child, stats)
		child = child.Sibling
	}
}

func getLayerName(layer rtui.Layer) string {
	switch layer {
	case rtui.LayerBase:
		return "BASE"
	case rtui.LayerOverlay:
		return "OVERLAY"
	case rtui.LayerModal:
		return "MODAL"
	case rtui.LayerTooltip:
		return "TOOLTIP"
	case rtui.LayerInspector:
		return "INSPECTOR"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", layer)
	}
}
