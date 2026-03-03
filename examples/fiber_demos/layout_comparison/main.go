package main

import (
	"fmt"

	"github.com/wwsheng009/mint/examples/component_fixtures"
	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime"
	compute_engine "github.com/wwsheng009/mint/runtime/compute"
	rtlayout "github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// 本程序比较 runtime/compute 和 runtime/layout 两个布局引擎
// 展示它们在处理相同VNode树时的异同
func main() {
	fmt.Println("=== Layout Engine Comparison: runtime/compute vs runtime/layout ===\n")


	// Initialize Intent Runtime (required for tests that call ui.On)
	intentRuntime := intent.NewRuntime()
	intent.SetupBuiltinHandlers(intentRuntime)
	rtui.SetGlobalIntentRuntime(intentRuntime)

	// 获取测试组件
	fixtures := component_fixtures.StandardFixtures()

	for _, fixture := range fixtures {
		fmt.Printf("\n========================================\n")
		fmt.Printf("Testing: %s\n", fixture.Name)
		fmt.Printf("Description: %s\n", fixture.Description)
		fmt.Printf("========================================\n\n")

		// 构建VNode
		vnode := fixture.Build()
		if vnode == nil {
			fmt.Printf("❌ Failed to build VNode\n")
			continue
		}

		// 使用 runtime/compute 布局
		computeResult, err := testComputeEngine(vnode)
		if err != nil {
			fmt.Printf("❌ runtime/compute failed: %v\n", err)
			continue
		}

		// 使用 runtime/layout 布局
		layoutResult, err := testLayoutEngine(vnode)
		if err != nil {
			fmt.Printf("❌ runtime/layout failed: %v\n", err)
			continue
		}

		// 比较结果
		compareResults(fixture.Name, computeResult, layoutResult)
	}

	fmt.Println("\n=== Comparison Complete ===")
}

// testComputeEngine 使用 runtime/compute 引擎进行布局
func testComputeEngine(vnode rtui.VNode) (*ComputeResult, error) {
	fmt.Println("--- runtime/compute Engine ---")

	// 创建Fiber
	fiber := reconciler.CreateFiberFromVNode(vnode)
	if fiber == nil {
		return nil, fmt.Errorf("failed to create fiber")
	}

	// 创建布局引擎
	engine := compute_engine.NewEngine()

	// 定义约束
	constraints := runtime.BoxConstraints{
		MinWidth:  80,
		MaxWidth:  80,
		MinHeight: 24,
		MaxHeight: 24,
	}

	// 执行布局
	layout, err := engine.Layout(vnode, fiber, constraints)
	if err != nil {
		return nil, fmt.Errorf("layout failed: %w", err)
	}
	if layout == nil {
		return nil, fmt.Errorf("layout returned nil")
	}

	// 统计信息
	nodeCount := component_fixtures.CountNodes(vnode)
	fiberCount := countFibers(fiber)
	boxCount := countComputeBoxes(layout)

	result := &ComputeResult{
		EngineName:   "runtime/compute",
		NodeCount:     nodeCount,
		FiberCount:    fiberCount,
		BoxCount:      boxCount,
		RootSize:      fmt.Sprintf("%dx%d", layout.Root.Box.Width, layout.Root.Box.Height),
		RootPosition:  fmt.Sprintf("(%d,%d)", layout.Root.Box.X, layout.Root.Box.Y),
		TotalBoxes:    boxCount,
	}

	fmt.Printf("✅ Layout successful\n")
	fmt.Printf("   VNode nodes: %d\n", nodeCount)
	fmt.Printf("   Fiber nodes: %d\n", fiberCount)
	fmt.Printf("   Layout boxes: %d\n", boxCount)
	fmt.Printf("   Root size: %s\n", result.RootSize)
	fmt.Printf("   Root position: %s\n", result.RootPosition)

	return result, nil
}

// testLayoutEngine 使用 runtime/layout 引擎进行布局
func testLayoutEngine(vnode rtui.VNode) (*LayoutResult, error) {
	fmt.Println("\n--- runtime/layout Engine ---")

	// 转换为layout.Node
	layoutNode := rtui.AsLayoutNode(vnode)
	if layoutNode == nil {
		return nil, fmt.Errorf("failed to convert to layout.Node")
	}

	// 创建布局引擎
	engine := rtlayout.NewEngine()

	// 定义约束
	constraints := rtlayout.NewConstraints(80, 80, 24, 24)

	// 执行布局
	result := engine.Layout(layoutNode, constraints)
	if result == nil {
		return nil, fmt.Errorf("layout returned nil")
	}

	// 统计信息
	nodeCount := component_fixtures.CountNodes(vnode)
	boxCount := len(result.Boxes)

	layoutResult := &LayoutResult{
		EngineName:   "runtime/layout",
		NodeCount:     nodeCount,
		BoxCount:      boxCount,
		RootSize:      fmt.Sprintf("%dx%d", result.Root.Width, result.Root.Height),
		RootPosition:  fmt.Sprintf("(%d,%d)", result.Root.X, result.Root.Y),
		TotalBoxes:    boxCount,
	}

	fmt.Printf("✅ Layout successful\n")
	fmt.Printf("   VNode nodes: %d\n", nodeCount)
	fmt.Printf("   Layout boxes: %d\n", boxCount)
	fmt.Printf("   Root size: %s\n", layoutResult.RootSize)
	fmt.Printf("   Root position: %s\n", layoutResult.RootPosition)

	return layoutResult, nil
}

// compareResults 比较两个引擎的布局结果
func compareResults(fixtureName string, compute *ComputeResult, layout *LayoutResult) {
	fmt.Println("\n--- Comparison ---")

	// 节点数比较
	fmt.Println("Node Count Comparison:")
	fmt.Printf("   VNode count: %d\n", compute.NodeCount)
	fmt.Printf("   runtime/compute boxes: %d\n", compute.BoxCount)
	fmt.Printf("   runtime/layout boxes: %d\n", layout.BoxCount)

	if compute.BoxCount == layout.BoxCount {
		fmt.Printf("   ✅ Box counts match\n")
	} else {
		fmt.Printf("   ⚠️  Box counts differ: %d vs %d\n", compute.BoxCount, layout.BoxCount)
	}

	// 尺寸比较
	fmt.Println("\nSize Comparison:")
	fmt.Printf("   runtime/compute root: %s\n", compute.RootSize)
	fmt.Printf("   runtime/layout root: %s\n", layout.RootSize)

	if compute.RootSize == layout.RootSize {
		fmt.Printf("   ✅ Root sizes match\n")
	} else {
		fmt.Printf("   ⚠️  Root sizes differ\n")
	}

	// 位置比较
	fmt.Println("\nPosition Comparison:")
	fmt.Printf("   runtime/compute root: %s\n", compute.RootPosition)
	fmt.Printf("   runtime/layout root: %s\n", layout.RootPosition)

	if compute.RootPosition == layout.RootPosition {
		fmt.Printf("   ✅ Root positions match\n")
	} else {
		fmt.Printf("   ⚠️  Root positions differ\n")
	}

	// 详细差异分析
	fmt.Println("\n--- Differences Analysis ---")
	analyzeDifferences(compute, layout)
}

// analyzeDifferences 分析两个引擎结果的差异
func analyzeDifferences(compute *ComputeResult, layout *LayoutResult) {
	// 分析尺寸差异
	if compute.RootSize != layout.RootSize {
		fmt.Println("⚠️  Root Size Difference:")
		fmt.Printf("   This is EXPECTED because:\n")
		fmt.Printf("   - runtime/compute uses Fiber system which actually measures VNode content\n")
		fmt.Printf("   - runtime/layout uses VNodeAdapter.GetSize() which returns GetBounds()\n")
		fmt.Printf("   - Un-layouted VNodes have bounds=[0,0,0,0]\n")
		fmt.Printf("   - runtime/compute sets bounds during layout, runtime/layout does not\n")
	}

	// 分析盒子数差异
	if compute.BoxCount != layout.BoxCount {
		fmt.Println("\n⚠️  Box Count Difference:")
		fmt.Printf("   Difference: %d boxes\n", compute.BoxCount-layout.BoxCount)
		fmt.Printf("   This might be due to:\n")
		fmt.Printf("   - Different node counting strategies\n")
		fmt.Printf("   - runtime/compute includes wrapper nodes\n")
		fmt.Printf("   - runtime/layout uses direct VNode tree traversal\n")
	}

	// 总结
	fmt.Println("\n--- Architecture Notes ---")
	fmt.Println("runtime/compute:")
	fmt.Println("  - Production layout engine")
	fmt.Println("  - Integrates with VNode/Fiber architecture")
	fmt.Println("  - Measures actual VNode content (text, components)")
	fmt.Println("  - Sets VNode bounds during layout")
	fmt.Println("  - Returns actual layout dimensions")
	fmt.Println("")
	fmt.Println("runtime/layout:")
	fmt.Println("  - Independent generic layout library")
	fmt.Println("  - Does not depend on VNode/Fiber")
	fmt.Println("  - Uses VNodeAdapter for compatibility")
	fmt.Println("  - GetSize() returns pre-layout bounds (0x0 for un-layouted VNodes)")
	fmt.Println("  - Can be used for layout-only computations")
}

// ComputeResult runtime/compute 引擎的布局结果
type ComputeResult struct {
	EngineName  string
	NodeCount    int
	FiberCount   int
	BoxCount     int
	RootSize     string
	RootPosition string
	TotalBoxes   int
}

// LayoutResult runtime/layout 引擎的布局结果
type LayoutResult struct {
	EngineName  string
	NodeCount    int
	BoxCount     int
	RootSize     string
	RootPosition string
	TotalBoxes   int
}

// countFibers 递归统计Fiber节点数
func countFibers(fiber *reconciler.Fiber) int {
	if fiber == nil {
		return 0
	}
	count := 1
	child := fiber.Child
	for child != nil {
		count += countFibers(child)
		child = child.Sibling
	}
	return count
}

// countComputeBoxes 递归统计runtime/compute的布局盒子数
func countComputeBoxes(layout *compute_engine.ComputedLayout) int {
	if layout == nil || layout.Root == nil {
		return 0
	}
	count := 0
	var traverse func(box *compute_engine.ComputedBox)
	traverse = func(box *compute_engine.ComputedBox) {
		if box == nil {
			return
		}
		count++
		for _, child := range box.Children {
			traverse(child)
		}
	}
	traverse(layout.Root)
	return count
}
