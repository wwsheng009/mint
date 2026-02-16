package main

import (
	"fmt"

	"github.com/wwsheng009/mint/examples/component_fixtures"
	rtlayout "github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// 此程序使用 component_fixtures 中的测试数据对 runtime/layout 包进行集成测试
// 验证布局引擎与真实UI组件的兼容性
func main() {
	fmt.Println("=== Layout Engine Integration Test with Component Fixtures ===\n")

	// 创建布局引擎
	engine := rtlayout.NewEngine()
	constraints := rtlayout.NewConstraints(80, 80, 24, 24)

	// 获取所有预定义组件
	fixtures := component_fixtures.StandardFixtures()

	fmt.Printf("Testing %d component fixtures...\n\n", len(fixtures))

	// 测试每个组件
	for _, fixture := range fixtures {
		fmt.Printf("--- Testing: %s ---\n", fixture.Name)
		fmt.Printf("Description: %s\n", fixture.Description)

		// 构建VNode
		vnode := fixture.Build()
		if vnode == nil {
			fmt.Printf("❌ FAILED: VNode build failed\n\n")
			continue
		}

		// 转换为layout.Node
		layoutNode := rtui.AsLayoutNode(vnode)

		// 执行布局
		result := engine.Layout(layoutNode, constraints)
		if result == nil || result.Root == nil {
			fmt.Printf("❌ FAILED: Layout calculation failed\n\n")
			continue
		}

		// 显示结果
		fmt.Printf("✅ SUCCESS\n")
		fmt.Printf("   Root size: %dx%d\n", result.Root.Width, result.Root.Height)
		fmt.Printf("   Position: (%d, %d)\n", result.Root.X, result.Root.Y)
		fmt.Printf("   Total boxes: %d\n", len(result.Boxes))

		fmt.Println()
	}

	// 测试自定义配置的Demo1应用
	fmt.Println("--- Testing Custom Configured Demo1 App ---")
	customVNode := component_fixtures.BuildDemo1App(
		component_fixtures.WithCount(100),
		component_fixtures.WithInput("test input"),
		component_fixtures.WithItems([]string{"A", "B", "C", "D", "E"}),
		component_fixtures.WithSize(120, 40),
	)

	customLayoutNode := rtui.AsLayoutNode(customVNode)
	customConstraints := rtlayout.NewConstraints(120, 120, 40, 40)
	customResult := engine.Layout(customLayoutNode, customConstraints)

	if customResult != nil && customResult.Root != nil {
		fmt.Printf("✅ Custom app layout successful\n")
		fmt.Printf("   Size: %dx%d\n", customResult.Root.Width, customResult.Root.Height)
		fmt.Printf("   Boxes: %d\n", len(customResult.Boxes))
	} else {
		fmt.Printf("❌ Custom app layout failed\n")
	}

	// 显示缓存统计
	stats := engine.GetStats()
	fmt.Printf("\n=== Cache Statistics ===\n")
	fmt.Printf("Cache Hits: %d\n", stats.CacheHits)
	fmt.Printf("Cache Misses: %d\n", stats.CacheMisses)
	total := stats.CacheHits + stats.CacheMisses
	if total > 0 {
		hitRate := float64(stats.CacheHits) / float64(total) * 100
		fmt.Printf("Hit Rate: %.2f%%\n", hitRate)
	}

	// 测试缓存一致性
	fmt.Println("\n=== Testing Cache Consistency ===")
	testFixture := component_fixtures.GetFixture("demo1_full_app")
	if testFixture != nil {
		vnode1 := testFixture.Build()
		layoutNode1 := rtui.AsLayoutNode(vnode1)

		result1 := engine.Layout(layoutNode1, constraints)
		result2 := engine.Layout(layoutNode1, constraints) // 应该命中缓存

		if result1 != nil && result2 != nil {
			if result1.Root.Width == result2.Root.Width &&
				result1.Root.Height == result2.Root.Height {
				fmt.Println("✅ Cache consistency verified - results are identical")
			} else {
				fmt.Println("❌ Cache consistency failed - results differ")
			}
		}

		// 测试缓存失效
		engine.Invalidate()
		result3 := engine.Layout(layoutNode1, constraints)

		if result3 != nil && result3.Root.Width == result1.Root.Width {
			fmt.Println("✅ Cache invalidation works - results still consistent")
		}
	}

	// 测试不同约束
	fmt.Println("\n=== Testing Different Constraints ===")
	constraintTests := []struct {
		name        string
		constraints rtlayout.Constraints
	}{
		{"Unbounded", rtlayout.UnboundedConstraints()},
		{"Tight 80x24", rtlayout.TightConstraints(80, 24)},
		{"Loose 10x10", rtlayout.LooseConstraints(10, 10)},
		{"Bounded 0-100x0-50", rtlayout.NewConstraints(0, 100, 0, 50)},
	}

	simpleFixture := component_fixtures.GetFixture("simple_vstack")
	if simpleFixture != nil {
		simpleVNode := simpleFixture.Build()
		simpleLayoutNode := rtui.AsLayoutNode(simpleVNode)

		for _, ct := range constraintTests {
			ctResult := engine.Layout(simpleLayoutNode, ct.constraints)
			if ctResult != nil && ctResult.Root != nil {
				fmt.Printf("✅ %s: %dx%d\n",
					ct.name,
					ctResult.Root.Width,
					ctResult.Root.Height)
			} else {
				fmt.Printf("❌ %s: FAILED\n", ct.name)
			}
		}
	}

	fmt.Println("\n=== All Tests Completed ===")
}
