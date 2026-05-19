package debug_keys

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/internal/reconciler"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// DebugKeyInspector 调试工具，用于显示所有层次的 KEY 信息
// 使用方法：
//
//	inspector := DebugKeyInspector{}
//	inspector.InspectVNodes(rootVNode)
//	inspector.InspectFibers(rootFiber)
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

	// 获取节点信息
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

	// 构建显示信息
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

	// 递归遍历子节点
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
		if fiber.ComponentName != "" {
			typeName = fiber.ComponentName
		} else if fiber.Tag != "" {
			typeName = fiber.Tag
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

	layer := fiber.Layer

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

	// 检查 Key 和 Path 是否匹配（重要！）
	if fiber.Key != "" && fiber.Path != "" && fiber.Key != fiber.Path {
		info += " \033[31m⚠️Key≠Path\033[0m"
	}

	fmt.Println(info)

	// 递归遍历 Fiber 树
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

	// 收集 VNode keys
	dki.collectVNodeKeys(vnode, vnodeMap)

	// 收收 Fiber keys
	dki.collectFiberKeys(fiber, fiberMap)

	fmt.Println("\n📊 Key Distribution:")
	fmt.Printf("  VNode unique keys: %d\n", len(vnodeMap))
	fmt.Printf("  Fiber unique keys: %d\n", len(fiberMap))

	// 找出不一致的 keys
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
	if fiber.Layer != rtui.LayerBase {
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
