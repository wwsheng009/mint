package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/runtime/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/panel"
	"github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/layout/visualizer"
)

// 演示 1: 基础 Panel 可视化 - 盒模型样式 (Chrome DevTools 风格)
func example1() {
	fmt.Println("=== 示例 1: 基础 Panel 盒模型可视化 ===")

	// 创建一个简单的 Panel
	p := panel.NewBuilder().
		Title("Settings").
		OuterSize(40, 15).
		Content(text.New("Settings go here")).
		Build()

	// 定义约束
	constraints := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 24,
	}

	// 从 VNode 自动创建可视化
	vis := visualizer.VisualizeVNode(p, constraints)

	// 打印盒模型 (Chrome DevTools 风格)
	fmt.Println(vis.PrintBoxModel())
}

// 演示 1b: 树形图样式
func example1b() {
	fmt.Println("=== 示例 1b: 树形图样式 ===")

	// 创建一个简单的 Panel
	p := panel.NewBuilder().
		Title("Settings").
		OuterSize(40, 15).
		Content(text.New("Settings go here")).
		Build()

	// 定义约束
	constraints := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 24,
	}

	// 从 VNode 自动创建可视化
	vis := visualizer.VisualizeVNode(p, constraints)

	// 打印树形图
	fmt.Println(vis.PrintTree())
}

// 演示 1c: 网格可视化
func example1c() {
	fmt.Println("=== 示例 1c: 网格可视化 ===")

	// 创建一个简单的 Panel
	p := panel.NewBuilder().
		Title("Settings").
		OuterSize(30, 12).
		Content(text.New("Settings go here")).
		Build()

	// 定义约束
	constraints := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  50,
		MinHeight: 0,
		MaxHeight: 20,
	}

	// 从 VNode 自动创建可视化
	vis := visualizer.VisualizeVNode(p, constraints)

	// 打印网格可视化
	fmt.Println(vis.PrintGrid())
}

// 演示 2: 手动构建可视化
func example2() {
	fmt.Println("=== 示例 2: 手动构建可视化 ===")

	vis := visualizer.NewVisualizer()

	// 添加根节点
	vis.AddNode(
		"root",
		"panel",
		layout.Rect{X: 0, Y: 0, Width: 40, Height: 15},
		layout.Constraints{MinWidth: 0, MaxWidth: 80, MinHeight: 0, MaxHeight: 24},
		layout.Constraints{},
		layout.Size{Width: 40, Height: 15},
		"",
	)

	// 添加子节点
	vis.AddNode(
		"child1",
		"text",
		layout.Rect{X: 2, Y: 2, Width: 38, Height: 5},
		layout.Constraints{MinWidth: 0, MaxWidth: 78, MinHeight: 0, MaxHeight: 22},
		layout.Constraints{},
		layout.Size{Width: 38, Height: 5},
		"root",
	)

	vis.AddNode(
		"child2",
		"button",
		layout.Rect{X: 15, Y: 10, Width: 10, Height: 3},
		layout.Constraints{MinWidth: 0, MaxWidth: 78, MinHeight: 0, MaxHeight: 22},
		layout.Constraints{},
		layout.Size{Width: 10, Height: 3},
		"root",
	)

	// 添加自定义属性
	vis.SetNodeProperty("root", "name", "MainPanel")
	vis.SetNodeProperty("root", "priority", "high")

	// 打印布局树
	fmt.Println(vis.PrintTree())
}

// 演示 3: 检测布局问题
func example3() {
	fmt.Println("=== 示例 3: 检测布局问题 ===")

	vis := visualizer.NewVisualizer()

	// 添加一个超出约束的节点（模拟问题）
	vis.AddNode(
		"problematic",
		"panel",
		layout.Rect{},
		layout.Constraints{MinWidth: 0, MaxWidth: 30, MinHeight: 0, MaxHeight: 10},  // 严格约束
		layout.Constraints{},
		layout.Size{Width: 40, Height: 15},  // 尺寸超出约束
		"",
	)

	// 查找问题
	problems := vis.FindProblems()

	if len(problems) == 0 {
		fmt.Println("✅ 未发现布局问题")
	} else {
		fmt.Printf("⚠️  发现 %d 个布局问题：\n\n", len(problems))
		for i, problem := range problems {
			fmt.Printf("%d. %s\n", i+1, problem)
		}
	}
}

// 演示 4: 约束传播链
func example4() {
	fmt.Println("=== 示例 4: 约束传播链 ===")

	vis := visualizer.NewVisualizer()

	// 构建一个简单的层级
	vis.AddNode(
		"parent",
		"panel",
		layout.Rect{},
		layout.Constraints{MinWidth: 0, MaxWidth: 80, MinHeight: 0, MaxHeight: 24},
		layout.Constraints{MinWidth: 0, MaxWidth: 78, MinHeight: 0, MaxHeight: 22},
		layout.Size{Width: 40, Height: 15},
		"",
	)

	vis.AddNode(
		"border",
		"border",
		layout.Rect{},
		layout.Constraints{MinWidth: 0, MaxWidth: 78, MinHeight: 0, MaxHeight: 22},
		layout.Constraints{MinWidth: 0, MaxWidth: 76, MinHeight: 0, MaxHeight: 20},
		layout.Size{Width: 38, Height: 13},
		"parent",
	)

	vis.AddNode(
		"content",
		"text",
		layout.Rect{},
		layout.Constraints{MinWidth: 0, MaxWidth: 76, MinHeight: 0, MaxHeight: 20},
		layout.Constraints{},
		layout.Size{Width: 38, Height: 10},
		"border",
	)

	// 打印从根到内容的约束传播链
	fmt.Println(vis.PrintConstraintChain("content"))
}

// 演示 5: 布局摘要
func example5() {
	fmt.Println("=== 示例 5: 布局摘要 ===")

	vis := visualizer.NewVisualizer()

	// 构建一个小布局树
	vis.AddNode(
		"root",
		"panel",
		layout.Rect{X: 0, Y: 0, Width: 50, Height: 20},
		layout.Constraints{MinWidth: 0, MaxWidth: 80, MinHeight: 0, MaxHeight: 24},
		layout.Constraints{},
		layout.Size{Width: 50, Height: 20},
		"",
	)

	vis.AddNode("title", "text", layout.Rect{},
		layout.Constraints{}, layout.Constraints{}, layout.Size{Width: 50, Height: 3}, "root",
	)

	vis.AddNode("content", "text", layout.Rect{},
		layout.Constraints{}, layout.Constraints{}, layout.Size{Width: 50, Height: 15}, "root",
	)

	vis.AddNode("footer", "text", layout.Rect{},
		layout.Constraints{}, layout.Constraints{}, layout.Size{Width: 50, Height: 2}, "root",
	)

	// 打印摘要
	fmt.Println(vis.PrintSummary())
}

// 演示 6: 复杂布局（带子布局）
func example6() {
	fmt.Println("=== 示例 6: 复杂布局 ===")

	// 创建带有子布局的 Panel
	mainPanel := panel.NewBuilder().
		Title("Main Application").
		OuterSize(60, 20).
		Content(
			ui.NewVStack().
				SetChildren([]rtui.VNode{
					panel.NewBuilder().
						Title("Section 1").
						OuterSize(50, 6).
						Content(text.New("Section 1 content")).
						Build(),
					panel.NewBuilder().
						Title("Section 2").
						OuterSize(50, 6).
						Content(text.New("Section 2 content")).
						Build(),
				}),
		).
		Build()

	// 从 VNode 创建可视化
	constraints := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 30,
	}

	vis := visualizer.VisualizeVNode(mainPanel, constraints)

	// 打印摘要
	fmt.Println(vis.PrintSummary())

	// 检查问题
	problems := vis.FindProblems()
	if len(problems) > 0 {
		fmt.Println("⚠️  发现布局问题：")
		for _, problem := range problems {
			fmt.Printf("  - %s\n", problem)
		}
	}
}

// 演示 7: 对比不同约束
func example7() {
	fmt.Println("=== 示例 7: 对比不同约束 ===")

	// 创建一个简单的 Panel
	p := panel.NewBuilder().
		Title("Responsive Panel").
		Content(text.New("This content should adapt to the constraints")).
		Build()

	// 约束 1: 小尺寸
	constraints1 := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  40,
		MinHeight: 0,
		MaxHeight: 10,
	}

	// 约束 2: 大尺寸
	constraints2 := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 20,
	}

	vis1 := visualizer.VisualizeVNode(p, constraints1)
	vis2 := visualizer.VisualizeVNode(p, constraints2)

	fmt.Println("--- 小尺寸约束 ---")
	fmt.Println(vis1.PrintSummary())

	fmt.Println("--- 大尺寸约束 ---")
	fmt.Println(vis2.PrintSummary())
}

func main() {
	// 运行所有示例
	example1()
	example2()
	example3()
	example4()
	example5()
	example6()
	example7()

	// 环境变量控制详细输出
	if os.Getenv("DEBUG_LAYOUT") == "true" {
		fmt.Println("=== 详细输出已启用 ===")
	}
}
