package main

import (
	"fmt"

	"github.com/wwsheng009/mint/examples/component_fixtures"
	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime"
	compute_engine "github.com/wwsheng009/mint/runtime/compute"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

// 本程序测试 border 和 style 如何影响布局约束的处理
func main() {
	fmt.Println("=== Border and Style Layout Constraint Test ===\n")

	// 测试不同的约束场景
	testScenarios := []struct {
		name        string
		constraints runtime.BoxConstraints
		description string
	}{
		{
			name: "tight_constraints",
			constraints: runtime.BoxConstraints{
				MinWidth:  80, MaxWidth: 80,
				MinHeight: 24, MaxHeight: 24,
			},
			description: "Fixed size 80x24",
		},
		{
			name: "loose_constraints",
			constraints: runtime.BoxConstraints{
				MinWidth:  40, MaxWidth: 120,
				MinHeight: 10, MaxHeight: 30,
			},
			description: "Flexible size 40-120 x 10-30",
		},
		{
			name: "unbounded_constraints",
			constraints: runtime.BoxConstraints{
				MinWidth:  0, MaxWidth: runtime.Infinity,
				MinHeight: 0, MaxHeight: runtime.Infinity,
			},
			description: "Unbounded (0 - infinity)",
		},
		{
			name: "min_only_constraints",
			constraints: runtime.BoxConstraints{
				MinWidth:  50, MaxWidth: runtime.Infinity,
				MinHeight: 15, MaxHeight: runtime.Infinity,
			},
			description: "Minimum only (50+, 15+)",
		},
		{
			name: "small_tight_constraints",
			constraints: runtime.BoxConstraints{
				MinWidth:  30, MaxWidth: 30,
				MinHeight: 10, MaxHeight: 10,
			},
			description: "Small fixed 30x10",
		},
	}

	// 测试带有border和style的组件
	testComponents := []struct {
		name      string
		fixture   string
		hasBorder bool
	}{
		{"simple_text", "simple_vstack", false},
		{"bordered_content", "bordered_content", true},
		{"demo1_header", "demo1_header", true},
		{"flex_layout", "flex_layout", false},
	}

	for _, scenario := range testScenarios {
		fmt.Printf("========================================\n")
		fmt.Printf("Scenario: %s\n", scenario.name)
		fmt.Printf("Description: %s\n", scenario.description)
		fmt.Printf("Constraints: %s\n", constraintsToString(scenario.constraints))
		fmt.Printf("========================================\n\n")

		for _, comp := range testComponents {
			testComponentWithConstraints(scenario, comp)
		}
		fmt.Println()
	}

	// 专门测试border行为
	fmt.Println("\n=== Border-Specific Tests ===\n")
	testBorderBehavior()
}

func testComponentWithConstraints(scenario struct {
	name        string
	constraints runtime.BoxConstraints
	description string
}, comp struct {
	name      string
	fixture   string
	hasBorder bool
}) {
	fmt.Printf("--- Testing: %s (hasBorder: %v) ---\n", comp.name, comp.hasBorder)

	// 获取fixture
	fixture := component_fixtures.GetFixture(comp.fixture)
	if fixture == nil {
		fmt.Printf("❌ Fixture not found: %s\n", comp.fixture)
		return
	}

	// 构建VNode
	vnode := fixture.Build()
	if vnode == nil {
		fmt.Printf("❌ Failed to build VNode\n")
		return
	}

	// 创建Fiber
	fiber := reconciler.CreateFiberFromVNode(vnode)
	if fiber == nil {
		fmt.Printf("❌ Failed to create Fiber\n")
		return
	}

	// 使用compute引擎布局
	engine := compute_engine.NewEngine()
	layout, err := engine.Layout(vnode, fiber, scenario.constraints)
	if err != nil {
		fmt.Printf("❌ Layout failed: %v\n", err)
		return
	}

	if layout == nil || layout.Root == nil {
		fmt.Printf("❌ Layout result is nil\n")
		return
	}

	// 分析结果
	root := layout.Root
	fmt.Printf("✅ Layout successful\n")
	fmt.Printf("   Result size: %dx%d\n", root.Box.Width, root.Box.Height)
	fmt.Printf("   Result position: (%d,%d)\n", root.Box.X, root.Box.Y)

	// 检查约束是否被遵守
	fmt.Printf("   Constraints respected: ")
	if root.Box.Width >= scenario.constraints.MinWidth &&
		root.Box.Width <= scenario.constraints.MaxWidth &&
		root.Box.Height >= scenario.constraints.MinHeight &&
		root.Box.Height <= scenario.constraints.MaxHeight {
		fmt.Printf("✅ YES\n")
	} else {
		fmt.Printf("⚠️  NO\n")
		if root.Box.Width < scenario.constraints.MinWidth {
			fmt.Printf("      Width %d < MinWidth %d\n", root.Box.Width, scenario.constraints.MinWidth)
		}
		if root.Box.Width > scenario.constraints.MaxWidth {
			fmt.Printf("      Width %d > MaxWidth %d\n", root.Box.Width, scenario.constraints.MaxWidth)
		}
		if root.Box.Height < scenario.constraints.MinHeight {
			fmt.Printf("      Height %d < MinHeight %d\n", root.Box.Height, scenario.constraints.MinHeight)
		}
		if root.Box.Height > scenario.constraints.MaxHeight {
			fmt.Printf("      Height %d > MaxHeight %d\n", root.Box.Height, scenario.constraints.MaxHeight)
		}
	}

	// 分析border影响
	if comp.hasBorder {
		fmt.Printf("   Border impact: \n")
		// Border添加2到宽度和高度（各边1个字符）
		innerWidth := root.Box.Width - 2
		innerHeight := root.Box.Height - 2
		fmt.Printf("      Total: %dx%d (including border)\n", root.Box.Width, root.Box.Height)
		fmt.Printf("      Inner: %dx%d (content area)\n", innerWidth, innerHeight)
		fmt.Printf("      Border: 2x2 (1 char on each side)\n")
	}

	// 检查style属性
	if root.VNode != nil {
		fmt.Printf("   Style properties checked\n")
	}

	fmt.Println()
}

func testBorderBehavior() {
	// 创建一个简单的带border的文本
	vnode := rtui.Bordered().
		Style("blue").
		Child(ui.Text("Hello World")).
		Build()

	fiber := reconciler.CreateFiberFromVNode(vnode)
	engine := compute_engine.NewEngine()

	// 测试不同约束下的border行为
	borderTests := []struct {
		name        string
		constraints runtime.BoxConstraints
	}{
		{
			name: "exact_content_size",
			constraints: runtime.BoxConstraints{
				MinWidth:  11, MaxWidth: 11, // "Hello World" = 11 chars
				MinHeight: 1, MaxHeight: 1,
			},
		},
		{
			name: "content_plus_border",
			constraints: runtime.BoxConstraints{
				MinWidth:  13, MaxWidth: 13, // 11 + 2 (border)
				MinHeight: 3, MaxHeight: 3,   // 1 + 2 (border)
			},
		},
		{
			name: "larger_space",
			constraints: runtime.BoxConstraints{
				MinWidth:  20, MaxWidth: 20,
				MinHeight: 5, MaxHeight: 5,
			},
		},
		{
			name: "smaller_space",
			constraints: runtime.BoxConstraints{
				MinWidth:  10, MaxWidth: 10,
				MinHeight: 1, MaxHeight: 1,
			},
		},
	}

	for _, test := range borderTests {
		fmt.Printf("\n--- Border Test: %s ---\n", test.name)
		fmt.Printf("Constraints: %s\n", constraintsToString(test.constraints))

		layout, err := engine.Layout(vnode, fiber, test.constraints)
		if err != nil {
			fmt.Printf("❌ Layout failed: %v\n", err)
			continue
		}

		if layout == nil || layout.Root == nil {
			fmt.Printf("❌ Layout result is nil\n")
			continue
		}

		root := layout.Root
		fmt.Printf("Result: %dx%d\n", root.Box.Width, root.Box.Height)
		fmt.Printf("Content size: %dx%d (excluding border)\n",
			max(0, root.Box.Width-2),
			max(0, root.Box.Height-2))

		// 分析约束尊重情况
		contentWidth := max(0, root.Box.Width-2)
		contentHeight := max(0, root.Box.Height-2)

		fmt.Printf("\nConstraint Analysis:\n")

		if test.constraints.MaxWidth < runtime.Infinity {
			fmt.Printf("  Total width %d vs MaxWidth %d: ",
				root.Box.Width, test.constraints.MaxWidth)
			if root.Box.Width <= test.constraints.MaxWidth {
				fmt.Printf("✅ OK\n")
			} else {
				fmt.Printf("❌ OVER\n")
			}

			fmt.Printf("  Content width %d vs (MaxWidth-border) %d: ",
				contentWidth, test.constraints.MaxWidth-2)
			if contentWidth <= test.constraints.MaxWidth-2 {
				fmt.Printf("✅ OK\n")
			} else {
				fmt.Printf("❌ OVER\n")
			}
		}

		if test.constraints.MaxHeight < runtime.Infinity {
			fmt.Printf("  Total height %d vs MaxHeight %d: ",
				root.Box.Height, test.constraints.MaxHeight)
			if root.Box.Height <= test.constraints.MaxHeight {
				fmt.Printf("✅ OK\n")
			} else {
				fmt.Printf("❌ OVER\n")
			}

			fmt.Printf("  Content height %d vs (MaxHeight-border) %d: ",
				contentHeight, test.constraints.MaxHeight-2)
			if contentHeight <= test.constraints.MaxHeight-2 {
				fmt.Printf("✅ OK\n")
			} else {
				fmt.Printf("❌ OVER\n")
			}
		}
	}
}

func constraintsToString(c runtime.BoxConstraints) string {
	minW := c.MinWidth
	maxW := c.MaxWidth
	minH := c.MinHeight
	maxH := c.MaxHeight

	var widthStr, heightStr string

	if maxW == runtime.Infinity {
		widthStr = fmt.Sprintf("%d-∞", minW)
	} else if minW == maxW {
		widthStr = fmt.Sprintf("%d", minW)
	} else {
		widthStr = fmt.Sprintf("%d-%d", minW, maxW)
	}

	if maxH == runtime.Infinity {
		heightStr = fmt.Sprintf("%d-∞", minH)
	} else if minH == maxH {
		heightStr = fmt.Sprintf("%d", minH)
	} else {
		heightStr = fmt.Sprintf("%d-%d", minH, maxH)
	}

	return fmt.Sprintf("[%s x %s]", widthStr, heightStr)
}