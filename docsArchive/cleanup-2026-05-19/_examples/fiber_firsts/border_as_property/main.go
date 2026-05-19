// 方案 A：边框作为容器属性演示
// 展示所有容器组件（Stack, Grid, Wrap, Absolute）原生支持边框，无需 Border 包装
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/examples/utils"
	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/grid"
	"github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/wrap"
)

// 辅助函数：Stack VNode
func borderStack(style, label string, children ...ui.VNode) ui.VNode {
	return ui.NewVStack().
		SetBorder(style, label).
		SetGap(0).
		SetChildrenList(children)
}

// 辅助函数：Grid VNode
func borderGrid(rows, cols int, style, label string, children ...ui.VNode) ui.VNode {
	return grid.New().
		Border(style, label).
		SetWidth(30)
}

// 辅助函数：Wrap VNode
func borderWrap(width int, style, label string, children ...ui.VNode) ui.VNode {
	return wrap.New().
		SetWidth(width).
		Border(style, label).
		SetGap(1).
		SetChildrenList(children)
}

// DemoApp 展示方案 A - 边框作为容器属性
func DemoApp() ui.VNode {
	return ui.NewVStack().
		SetGap(1).
		SetChildrenList([]ui.VNode{
			// =====================================================
			// 标题
			// =====================================================
			sectionTitle("==== 方案 A: 边框作为容器属性 ===="),
			highlight("所有容器组件原生支持边框，无需 Border 包装"),

			// =====================================================
			// Section 1: Stack 容器边框
			// =====================================================
			sectionTitle("==== 1. Stack 容器边框 ===="),

			subTitle("1.1 VStack 单线边框"),
			ui.NewVStack().
				SingleBorder("Stack Title").
				SetGap(0).
				SetChildrenList([]ui.VNode{
					text.New("Item 1"),
					text.New("Item 2"),
					text.New("Item 3"),
				}),

			subTitle("1.2 HStack 双线边框"),
			ui.NewHStack().
				DoubleBorder("Horizontal").
				SetWidth(25).
				SetGap(2).
				SetChildrenList([]ui.VNode{
					text.New("[A]"),
					text.New("[B]"),
					text.New("[C]"),
				}),

			subTitle("1.3 VStack 圆角边框"),
			borderStack("rounded", "标题",
				text.New("圆角边框容器"),
				text.New("内容"),
			),

			// =====================================================
			// Section 2: Grid 容器边框
			// =====================================================
			sectionTitle("==== 2. Grid 容器边框 ===="),

			subTitle("2.1 Grid 单线边框"),
			grid.New().
				SetColumns(grid.Flex{Factor: 1}, grid.Flex{Factor: 1}).
				SetRows(grid.Fixed(1), grid.Fixed(1)).
				SingleBorder("2x2 Grid").
				SetChildrenAuto([]ui.VNode{
					text.New("A1"), text.New("B1"),
					text.New("A2"), text.New("B2"),
				}),

			subTitle("2.2 Grid 双线边框"),
			grid.New().
				SetColumns(grid.Flex{Factor: 1}, grid.Flex{Factor: 1}, grid.Flex{Factor: 1}).
				SetRows(grid.Fixed(1), grid.Fixed(1)).
				DoubleBorder("3x2 Grid").
				SetChildrenAuto([]ui.VNode{
					text.New("(1,1)"), text.New("(1,2)"), text.New("(1,3)"),
					text.New("(2,1)"), text.New("(2,2)"), text.New("(2,3)"),
				}),

			// =====================================================
			// Section 3: Wrap 容器边框
			// =====================================================
			sectionTitle("==== 3. Wrap 容器边框 ===="),

			subTitle("3.1 Wrap 单线边框"),
			wrap.New().
				SetWidth(30).
				SingleBorder("Wrap").
				SetGap(1).
				SetChildrenList([]ui.VNode{
					text.New("Item1"), text.New("Item2"),
					text.New("Item3"), text.New("Item4"),
					text.New("Item5"), text.New("Item6"),
				}),

			// =====================================================
			// Section 4: 边框样式
			// =====================================================
			sectionTitle("==== 5.1 边框样式 ===="),

			subTitle("5.1 单线 (Single)"),
			borderStack("single", "",
				text.New("Single"),
				text.New("Line"),
				text.New("Border"),
			),

			subTitle("5.2 双线 (Double)"),
			borderStack("double", "",
				text.New("Double"),
				text.New("Line"),
			),

			subTitle("5.3 圆角 (Rounded)"),
			borderStack("rounded", "",
				text.New("Rounded"),
				text.New("Corners"),
			),

			subTitle("5.4 虚线 (Dashed)"),
			borderStack("dashed", "",
				text.New("Dashed"),
				text.New("Line"),
			),

			subTitle("5.5 无边框 (None)"),
			borderStack("none", "",
				text.New("No"),
				text.New("Border"),
			),

			// =====================================================
			// Section 6: 嵌套边框
			// =====================================================
			sectionTitle("==== 6. 嵌套边框 ===="),

			subTitle("6.1 Stack 嵌套"),
			ui.NewVStack().
				DoubleBorder("Outer").
				SetGap(0).
				SetChildrenList([]ui.VNode{
					ui.NewVStack().
						SingleBorder("Inner").
						SetGap(0).
						SetChildrenList([]ui.VNode{
							text.New("Nested"),
							text.New("Content"),
						}),
				}),
		})
}

// sectionTitle 创建章节标题
func sectionTitle(title string) ui.VNode {
	return text.New(title).
		Foreground(theme.Primary()).
		Bold(true)
}

// subTitle 创建子标题
func subTitle(title string) ui.VNode {
	return text.New("  " + title)
}

// highlight 创建高亮提示
func highlight(msg string) ui.VNode {
	return text.New("    >>> " + msg + " <<<").
		Foreground("yellow")
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("========================================")
	fmt.Println("  方案 A：边框作为容器属性演示        ")
	fmt.Println("  所有容器组件原生支持边框            ")
	fmt.Println("========================================")

	// 创建 framework 应用
	fwApp := framework.NewApp()

	// 创建带有 Fiber reconciler 的 DeclarativeNode
	node := render.NewDeclarativeNodeFromFuncWithFiber(DemoApp)
    node.SetApp(fwApp)
	node.SetRenderMode(render.RenderModeFiberFirst)

	fmt.Printf("\n配置:\n  渲染模式: %v\n", node.GetRenderMode())

	// 创建缓冲区 (80 x 100)
	buf := paint.NewBuffer(80, 100)
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 80, Height: 100},
		AvailableWidth:  80,
		AvailableHeight: 100,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Println("渲染具有原生边框支持的容器组件...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 80))

	// 渲染
	node.Paint(ctx, buf)

	// 输出结果并保存
	utils.PrintBuffer(buf, 80, 100)
	utils.SaveBuffer(buf, 80, 100)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("方案 A 边框特性总结:")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("  支持的容器:")
	fmt.Println("    - Stack (VStack/HStack)")
	fmt.Println("    - Grid")
	fmt.Println("    - Wrap")
	fmt.Println("    - Absolute")
	fmt.Println("    - Modal (默认 double)")
	fmt.Println("")
	fmt.Println("  边框样式:")
	fmt.Println("    - Single, Double, Rounded, Dashed, None")
	fmt.Println("")
	fmt.Println("  API:")
	fmt.Println("    - Border(style, label)")
	fmt.Println("    - SingleBorder(), DoubleBorder(), etc.")
	fmt.Println("")
	fmt.Println(strings.Repeat("=", 80))
}
