// Interactive Layer Visualization Demo
// 交互式 Layer 可视化演示 - 使用 framework.App 和 ui/components
package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/ui"
	newtext "github.com/wwsheng009/mint/ui/components/text"
	tooltipcomponent "github.com/wwsheng009/mint/ui/components/tooltip"
)

// LayerVisualApp 主应用 VNode
func LayerVisualApp() ui.VNode {
	return ui.NewVStack().
		SetGap(0).
		SetChildrenList([]ui.VNode{
			// ┌─────────────────────────────────────────────────────────────┐
			// │  标题栏                                                       │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("╔════════════════════════════════════════════════════════╗").
				Foreground(theme.Primary()).Bold(true),
			newtext.New("║   Interactive Layer Z-Order Visualization            ║").
				Foreground(theme.Primary()).Bold(true),
			newtext.New("╚════════════════════════════════════════════════════════╝").
				Foreground(theme.Primary()).Bold(true),

			newtext.New(""),
			newtext.New("  按按钮触发不同 Layer 的 Toast 通知"),
			newtext.New("  展示 Fiber-first 多层渲染效果（从 Layer 0 到 4）"),
			newtext.New("  按 ESC 或 Ctrl+C 退出程序"),
			newtext.New(""),

			// ┌─────────────────────────────────────────────────────────────┐
			// │  Toast 显示区域（固定位置，方便观察层级效果）                 │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("═══ Toast 显示区域 (不同层级的 Toast) ═══").
				Foreground(theme.Primary()).Bold(true),
			newtext.New(""),

			// 为了展示 Z-Order 覆盖效果，我们将不同 Layer 的 Toast 放在相似的视觉位置
			newtext.New("【Layer 0】Base 层 - 主要内容："),
			tooltipcomponent.NewToastBuilder("🔴 Layer 0: (最宽，底层背景)").
				Layer(ui.LayerBase).
				Build(),

			newtext.New(""),
			newtext.New("【Layer 2】Modal 层 - 模态对话框："),
			tooltipcomponent.NewToastBuilder("🟡 Layer 2: (中等宽度)").
				Layer(ui.LayerModal).
				Warning().
				Build(),

			newtext.New(""),
			newtext.New("【Layer 4】Inspector 层 - 最高层级："),
			tooltipcomponent.NewToastBuilder("🟢 Layer 4: (最高，覆盖所有)").
				Layer(ui.LayerInspector).
				Success().
				Build(),

			newtext.New(""),

			// ┌─────────────────────────────────────────────────────────────┐
			// │  更多 Toast 示例                                            │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New(""),
			newtext.New("【Layer 1】Overlay 层 - 标准通知："),
			tooltipcomponent.NewToastBuilder("🔵 Layer 1: (默认 Toast 层)").
				Layer(ui.LayerOverlay).
				Info().
				Build(),

			newtext.New(""),
			newtext.New("【Layer 3】Tooltip 层 - 悬浮提示："),
			tooltipcomponent.NewToastBuilder("🟣 Layer 3: (Tooltip 层)").
				Layer(ui.LayerTooltip).
				Build(),

			newtext.New(""),

			// ┌─────────────────────────────────────────────────────────────┐
			// │  Layer 渲染顺序说明                                           │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("═══ Layer 渲染顺序 (Z-Order) ═══").
				Foreground(theme.Primary()).Bold(true),
			newtext.New(""),
			newtext.New("  渲染顺序（先绘制底层，后绘制顶层）："),
			newtext.New(""),
			newtext.New("    1️⃣  Layer 0: Base     (主要内容层)"),
			newtext.New("    2️⃣  Layer 1: Overlay  (标准通知层)"),
			newtext.New("    3️⃣  Layer 2: Modal    (模态对话框层)"),
			newtext.New("    4️⃣  Layer 3: Tooltip  (悬浮提示层)"),
			newtext.New("    5️⃣  Layer 4: Inspector(调试/覆盖层)"),
			newtext.New(""),
			newtext.New("  ✅ 高层级后绘制，覆盖低层级内容"),
			newtext.New("  ✅ 同位置的多层元素，顶层可见"),
			newtext.New(""),

			// ┌─────────────────────────────────────────────────────────────┐
			// │  底部说明                                                     │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("═══════════════════════════════════════════════════════════════"),
			newtext.New(""),
			newtext.New("  技术实现："),
			newtext.New("    • framework.App - 完整的事件循环和渲染系统"),
			newtext.New("    • SetRoot() - 自动启用 Fiber-first 渲染模式"),
			newtext.New("    • fiberFirstPaint() - 使用 PaintPaintablePlanes()"),
			newtext.New("    • 按 renderOrder= [0,1,2,3,4] 顺序绘制各层级"),
			newtext.New(""),
			newtext.New("  验证要点："),
			newtext.New("    ✓ 不同层级的 Toast 按顺序显示"),
			newtext.New("    ✓ 颜色区分不同层级（红/黄/绿/蓝/紫）"),
			newtext.New("    ✓ Toast 自动定位避免重叠（但层级信息已正确传递）"),
			newtext.New(""),
			newtext.New("═══════════════════════════════════════════════════════════════"),
		})
}

func main() {
	// 环境变量设置（可选，SetRoot 已自动处理）
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║   Interactive Layer Z-Order Visualization              ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")

	// 创建 framework App
	fwApp := framework.NewApp()
	fwApp.Resize(80, 40)

	// 创建 DeclarativeNode 使用 Fiber-first 渲染
	// 在交互模式下展示不同 Layer 的 Toast
	declarativeNode := render.NewDeclarativeNodeFromFuncWithFiber(
		LayerVisualApp,
		fwApp,
	)

	// 设置 Root（自动启用 Fiber-first 模式）
	fwApp.SetRoot(declarativeNode)

	fmt.Println("")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("  应用特点：")
	fmt.Println("    • 使用 framework.NewApp() 实现完整的事件循环")
	fmt.Println("    • 自动启用 Fiber-first 多层渲染路径")
	fmt.Println("    • 展示 5 个不同 Layer 的 Toast 组件")
	fmt.Println("")
	fmt.Println("  Layer 说明：")
	fmt.Println("    Layer 0 (Base):     🔴 红色 - 主要内容层")
	fmt.Println("    Layer 1 (Overlay):  🔵 蓝色 - 标准通知层")
	fmt.Println("    Layer 2 (Modal):    🟡 黄色 - 模态对话框层")
	fmt.Println("    Layer 3 (Tooltip):  🟣 紫色 - 悬浮提示层")
	fmt.Println("    Layer 4 (Inspector):🟢 绿色 - 调试覆盖层（最高）")
	fmt.Println("")
	fmt.Println("  注意：")
	fmt.Println("    • Toast 组件会自动定位，避免视觉重叠")
	fmt.Println("    • 但渲染顺序正确：从 Layer 0 到 4 依次绘制")
	fmt.Println("    • 如果元素在同一位置，高层级会覆盖低层级")
	fmt.Println("")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("")
	fmt.Println("正在启动应用... 按 ESC 或 Ctrl+C 退出")
	fmt.Println("")

	// 运行应用（交互模式）
	if err := fwApp.Run(); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
}
