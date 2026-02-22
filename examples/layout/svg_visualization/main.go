package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/panel"
	"github.com/wwsheng009/mint/ui/components/stack"
	"github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/layout/visualizer"
)

// 演示: SVG 可视化输出
func main() {
	fmt.Println("=== SVG Layout Visualization Demo ===\n")

	// 示例 1: 简单的 Panel 布局
	fmt.Println("示例 1: 简单的 Panel 布局")
	fmt.Println("---")
	simplePanel := panel.NewBuilder().
		Title("Settings").
		OuterSize(40, 15).
		Content(text.New("Settings go here")).
		Build()

	constraints := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 24,
	}

	vis := visualizer.VisualizeVNode(simplePanel, constraints)

	// 生成 SVG 文件（详细版）
	svgOutput := vis.PrintSVG()

	err := os.WriteFile("simple_layout.svg", []byte(svgOutput), 0644)
	if err != nil {
		fmt.Printf("Error writing SVG file: %v\n", err)
		return
	}

	fmt.Println("✓ 已生成 simple_layout.svg 文件")
	fmt.Println("  在浏览器或 SVG 查看器中打开此文件\n")

	// 示例 2: 复杂的嵌套布局
	fmt.Println("示例 2: 复杂的嵌套布局")
	fmt.Println("---")

	complexLayout := panel.NewBuilder().
		Title("Dashboard").
		OuterSize(70, 30).
		Content(
			stack.NewVStack().
				SetChildren([]ui.VNode{
					panel.NewBuilder().
						Title("Statistics").
						OuterSize(60, 8).
						Content(text.New("Users: 100 | Sales: $5K | Orders: 50")).
						Build(),
					panel.NewBuilder().
						Title("Recent Activity").
						OuterSize(60, 10).
						Content(text.New("Latest updates and user activities go here")).
						Build(),
					panel.NewBuilder().
						Title("Quick Actions").
						OuterSize(60, 6).
						Content(text.New("Create | Edit | Delete")).
						Build(),
				}),
		).
		Build()

	vis2 := visualizer.VisualizeVNode(complexLayout, constraints)
	svgOutput2 := vis2.PrintSVG()

	err = os.WriteFile("complex_layout.svg", []byte(svgOutput2), 0644)
	if err != nil {
		fmt.Printf("Error writing SVG file: %v\n", err)
		return
	}

	fmt.Println("✓ 已生成 complex_layout.svg 文件\n")

	// 示例 3: 简化版 SVG（圆形节点）
	fmt.Println("示例 3: 简化版 SVG（圆形节点）")
	fmt.Println("---")
	simpleSVG := vis.PrintSVGSimple()

	err = os.WriteFile("simple_layout_circle.svg", []byte(simpleSVG), 0644)
	if err != nil {
		fmt.Printf("Error writing SVG file: %v\n", err)
		return
	}

	fmt.Println("✓ 已生成 simple_layout_circle.svg 文件\n")

	// 示例 4: Tree map 风格
	fmt.Println("示例 4: Tree Map 风格可视化")
	fmt.Println("---")
	treemapSVG := vis.PrintSVGTreeMap()

	err = os.WriteFile("treemap_layout.svg", []byte(treemapSVG), 0644)
	if err != nil {
		fmt.Printf("Error writing SVG file: %v\n", err)
		return
	}

	fmt.Println("✓ 已生成 treemap_layout.svg 文件\n")

	// 示例 5: 行业级测试：大组件树
	fmt.Println("示例 5: 大型组件树")
	fmt.Println("---")

	massiveLayout := panel.NewBuilder().
		Title("Massive Layout").
		OuterSize(90, 50).
		Content(
			stack.NewVStack().
				SetChildren([]ui.VNode{
					panel.NewBuilder().Title("Section 1").OuterSize(80, 8).
						Content(text.New("Content 1")).Build(),
					panel.NewBuilder().Title("Section 2").OuterSize(80, 8).
						Content(text.New("Content 2")).Build(),
					panel.NewBuilder().Title("Section 3").OuterSize(80, 8).
						Content(text.New("Content 3")).Build(),
					panel.NewBuilder().Title("Section 4").OuterSize(80, 8).
						Content(text.New("Content 4")).Build(),
					panel.NewBuilder().Title("Section 5").OuterSize(80, 8).
						Content(text.New("Content 5")).Build(),
				}),
		).
		Build()

	vis3 := visualizer.VisualizeVNode(massiveLayout, layout.Constraints{
		MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 60,
	})
	massiveSVG := vis3.PrintSVG()

	err = os.WriteFile("massive_layout.svg", []byte(massiveSVG), 0644)
	if err != nil {
		fmt.Printf("Error writing SVG file: %v\n", err)
		return
	}

	fmt.Println("✓ 已生成 massive_layout.svg 文件\n")

	// 打印统计
	fmt.Println("=== 生成统计 ===")
	fmt.Println("  大型布局树已成功渲染为 SVG")

	// 打印问题
	problems := vis3.FindProblems()
	if len(problems) > 0 {
		fmt.Printf("⚠️  检测到 %d 个布局问题:\n", len(problems))
		for i, problem := range problems {
			fmt.Printf("  %d. %s\n", i+1, problem)
		}
	} else {
		fmt.Println("✓ 未检测到布局问题")
	}

	fmt.Println("\n=== Demo 完成 ===")
	fmt.Println("\n提示: 生成的 SVG 文件包含:")
	fmt.Println("  - 详细版 SVG: 显示完整的节点信息、约束、连接线")
	fmt.Println("  - 简化版 SVG: 使用圆形节点，适合概览")
	fmt.Println("  - Tree Map: 使用嵌套矩形，适合快速查看节点分布")
	fmt.Println("  - 所有 SVG 都可以在浏览器中直接打开查看")
	fmt.Println("  - SVG 文件可以导入到其他设计工具中进一步编辑")
}
