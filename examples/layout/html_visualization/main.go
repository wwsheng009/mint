package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/panel"
	"github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/layout/visualizer"
)

// 演示: HTML 可视化输出
func main() {
	fmt.Println("=== HTML Layout Visualization Demo ===\n")

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

	// 生成完整的 HTML 文件
	htmlOutput := vis.PrintHTML()

	// 保存到文件
	err := os.WriteFile("simple_layout.html", []byte(htmlOutput), 0644)
	if err != nil {
		fmt.Printf("Error writing HTML file: %v\n", err)
		return
	}

	fmt.Println("✓ 已生成 simple_layout.html 文件")
	fmt.Println("  在浏览器中打开此文件查看可视化布局\n")

	// 示例 2: 带有多个嵌套组件的复杂布局
	fmt.Println("示例 2: 复杂的嵌套布局")
	fmt.Println("---")

	complexLayout := panel.NewBuilder().
		Title("Dashboard").
		OuterSize(70, 30).
		Content(
			ui.NewVStack().
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
	htmlOutput2 := vis2.PrintHTML()

	err = os.WriteFile("complex_layout.html", []byte(htmlOutput2), 0644)
	if err != nil {
		fmt.Printf("Error writing HTML file: %v\n", err)
		return
	}

	fmt.Println("✓ 已生成 complex_layout.html 文件")
	fmt.Println("  在浏览器中打开此文件查看可视化布局\n")

	// 示例 3: 行内 HTML 格式 (适合嵌入其他页面)
	fmt.Println("示例 3: 行内 HTML 输出")
	fmt.Println("---")
	inlineHTML := vis.PrintHTMLOneline()
	fmt.Println("行内 HTML (前 200 字符):")
	fmt.Println(inlineHTML[:min(200, len(inlineHTML))] + "...\n")

	// 示例 4: 节点索引
	fmt.Println("示例 4: 节点索引")
	fmt.Println("---")
	indexHTML := vis.PrintHTMLIndex()
	err = os.WriteFile("node_index.html", []byte(indexHTML), 0644)
	if err != nil {
		fmt.Printf("Error writing HTML file: %v\n", err)
		return
	}

	fmt.Println("✓ 已生成 node_index.html 文件\n")

	// 示例 5: 检测布局问题
	fmt.Println("示例 5: 布局问题检测")
	fmt.Println("---")

	// 创建一个超出约束的布局
	problematicLayout := panel.NewBuilder().
		Title("Oversized Panel").
		OuterSize(100, 50). // 超出约束
		Content(text.New("This content might exceed constraints")).
		Build()

	strictConstraints := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  50,  // 宽度限制
		MinHeight: 0,
		MaxHeight: 20,  // 高度限制
	}

	vis3 := visualizer.VisualizeVNode(problematicLayout, strictConstraints)
	htmlOutput3 := vis3.PrintHTML()

	err = os.WriteFile("problems_layout.html", []byte(htmlOutput3), 0644)
	if err != nil {
		fmt.Printf("Error writing HTML file: %v\n", err)
		return
	}

	fmt.Println("✓ 已生成 problems_layout.html 文件")
	fmt.Println("  此文件展示了布局问题的可视化（超出约束）\n")

	// 打印检测到的问题
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
	fmt.Println("\n提示: 生成的 HTML 文件包含:")
	fmt.Println("  - 交互式节点导航")
	fmt.Println("  - 约束详情显示")
	fmt.Println("  - 布局问题高亮")
	fmt.Println("  - 响应式布局")
	fmt.Println("  - 节点索引导航")
}