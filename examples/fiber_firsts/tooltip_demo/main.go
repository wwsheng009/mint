// Tooltip Component Demo
// Tooltip 组件演示 - 展示各种 Tooltip 和 Toast 功能
package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/ui"
	newtext "github.com/wwsheng009/mint/ui/components/text"
	tooltipcomp "github.com/wwsheng009/mint/ui/components/tooltip"
)

// TooltipDemoApp 主应用 VNode
func TooltipDemoApp() ui.VNode {
	return ui.NewVStack().
		SetGap(0).
		SetChildrenList([]ui.VNode{
			// ┌─────────────────────────────────────────────────────────────┐
			// │  标题栏                                                       │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("╔════════════════════════════════════════════════════════╗").
				Foreground(theme.Primary()),
			newtext.New("║   Tooltip & Toast Component Demo                     ║").
				Foreground(theme.Primary()).Bold(true),
			newtext.New("╚════════════════════════════════════════════════════════╝").
				Foreground(theme.Primary()),

			newtext.New(""),
			newtext.New("  本演示展示 Tooltip 组件的各种功能和样式"),
			newtext.New("  按 ESC 或 Ctrl+C 退出程序"),
			newtext.New(""),

			// ┌─────────────────────────────────────────────────────────────┐
			// │  Toast 示例                                                  │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("═══ Toast 组件示例 ═══").
				Foreground(theme.Primary()).Bold(true),
			newtext.New(""),

			newtext.New("【默认 Toast】"),
			tooltipcomp.NewToastBuilder("这是一个默认 Toast 通知").
				Info().
				Build(),

			newtext.New(""),
			newtext.New("【成功 Toast】"),
			tooltipcomp.NewToastBuilder("操作成功完成！").
				Success().
				Build(),

			newtext.New(""),
			newtext.New("【警告 Toast】"),
			tooltipcomp.NewToastBuilder("请注意：此操作不可撤销").
				Warning().
				Build(),

			newtext.New(""),
			newtext.New("【错误 Toast】"),
			tooltipcomp.NewToastBuilder("发生错误：请重试").
				Error().
				Build(),

			newtext.New(""),

			// ┌─────────────────────────────────────────────────────────────┐
			// │  不同层级的 Toast                                            │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("═══ 不同层级的 Toast ═══").
				Foreground(theme.Primary()).Bold(true),
			newtext.New(""),

			newtext.New("【Modal 层 Toast】"),
			tooltipcomp.NewToastBuilder("模态层通知（黄色）").
				Layer(ui.LayerModal).
				Warning().
				Build(),

			newtext.New(""),
			newtext.New("【Tooltip 层 Toast】"),
			tooltipcomp.NewToastBuilder("提示层通知（紫色）").
				Layer(ui.LayerTooltip).
				Build(),

			newtext.New(""),
			newtext.New("【Inspector 层 Toast】"),
			tooltipcomp.NewToastBuilder("最高层级通知（绿色）").
				Layer(ui.LayerInspector).
				Success().
				Build(),

			newtext.New(""),

			// ┌─────────────────────────────────────────────────────────────┐
			// │  底部说明                                                     │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("═══════════════════════════════════════════════════════════════"),
			newtext.New(""),
			newtext.New("  Toast 组件特性："),
			newtext.New("    • 自适应宽度，自动换行"),
			newtext.New("    • 语义化样式：Info/Success/Warning/Error"),
			newtext.New("    • 支持 5 个渲染层级（Layer 0-4）"),
			newtext.New("    • 临时消息显示，适合通知和提示"),
			newtext.New(""),
			newtext.New("  层级说明："),
			newtext.New("    Layer 0 (Base):     主要内容层"),
			newtext.New("    Layer 1 (Overlay):  默认 Toast 层"),
			newtext.New("    Layer 2 (Modal):    模态对话框层"),
			newtext.New("    Layer 3 (Tooltip):  悬浮提示层"),
			newtext.New("    Layer 4 (Inspector): 调试覆盖层"),
			newtext.New(""),
			newtext.New("═══════════════════════════════════════════════════════════════"),
		})
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║   Tooltip & Toast Component Demo                      ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")

	fwApp := framework.NewApp()
	fwApp.Resize(80, 40)

	declarativeNode := render.NewDeclarativeNodeFromFuncWithFiber(
		TooltipDemoApp,
		fwApp,
	)

	fwApp.SetRoot(declarativeNode)

	fmt.Println("")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("  本演示展示：")
	fmt.Println("    • 不同样式的 Toast（Info/Success/Warning/Error）")
	fmt.Println("    • 不同层级的 Toast 渲染效果")
	fmt.Println("    • Fiber-first 多层渲染机制")
	fmt.Println("")
	fmt.Println("  特性：")
	fmt.Println("    • Toast 自动定位，避免视觉重叠")
	fmt.Println("    • 语义化颜色区分不同类型的通知")
	fmt.Println("    • 支持层级控制，用于不同场景的提示")
	fmt.Println("")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("")
	fmt.Println("演示已启动... 按 ESC 或 Ctrl+C 退出")
	fmt.Println("")

	if err := fwApp.Run(); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
}
