// Layer Rendering Demo
// Layer 渲染演示 - 展示多层级渲染机制
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

// LayerRenderApp 主应用 VNode
func LayerRenderApp() ui.VNode {
	return ui.NewVStack().
		SetGap(0).
		SetChildrenList([]ui.VNode{
			// ┌─────────────────────────────────────────────────────────────┐
			// │  标题栏                                                       │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("╔════════════════════════════════════════════════════════╗").
				Foreground(theme.Primary()),
			newtext.New("║   Layer Rendering Demo                               ║").
				Foreground(theme.Primary()).Bold(true),
			newtext.New("╚════════════════════════════════════════════════════════╝").
				Foreground(theme.Primary()),

			newtext.New(""),
			newtext.New("  展示 Fiber-first 多层级渲染机制"),
			newtext.New("  按 ESC 或 Ctrl+C 退出程序"),
			newtext.New(""),

			// ┌─────────────────────────────────────────────────────────────┐
			// │  多层级渲染演示                                               │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("═══ 多层级渲染演示 ═══").
				Foreground(theme.Primary()).Bold(true),
			newtext.New(""),

			newtext.New("渲染顺序（Z-Order）："),
			newtext.New("  1️⃣  Layer 0 (Base)     - 大尺寸，底层背景"),
			newtext.New("  2️⃣  Layer 1 (Overlay)  - 中等尺寸，标准内容"),
			newtext.New("  3️⃣  Layer 2 (Modal)    - 小尺寸，模态对话框"),
			newtext.New("  4️⃣  Layer 3 (Tooltip)  - 最小尺寸，悬浮提示"),
			newtext.New("  5️⃣  Layer 4 (Inspector)- 特殊标记，调试覆盖"),
			newtext.New(""),

			newtext.New("【Layer 0】Base 层 (大范围背景)"),
			tooltipcomp.NewToastBuilder("🔴 Layer 0: Base - 底层背景内容").
				Layer(ui.LayerBase).
				Build(),

			newtext.New(""),

			newtext.New("【Layer 1】Overlay 层 (普通内容)"),
			tooltipcomp.NewToastBuilder("🔵 Layer 1: Overlay - 标准通知").
				Layer(ui.LayerOverlay).
				Build(),

			newtext.New(""),

			newtext.New("【Layer 2】Modal 层 (模态框)"),
			tooltipcomp.NewToastBuilder("🟡 Layer 2: Modal - 模态对话框").
				Layer(ui.LayerModal).
				Warning().
				Build(),

			newtext.New(""),

			newtext.New("【Layer 3】Tooltip 层 (提示框)"),
			tooltipcomp.NewToastBuilder("🟣 Layer 3: Tooltip - 悬浮提示").
				Layer(ui.LayerTooltip).
				Build(),

			newtext.New(""),

			newtext.New("【Layer 4】Inspector 层 (调试层)"),
			tooltipcomp.NewToastBuilder("🟢 Layer 4: Inspector - 最高层级").
				Layer(ui.LayerInspector).
				Success().
				Build(),

			newtext.New(""),

			// ┌─────────────────────────────────────────────────────────────┐
			// │  渲染机制说明                                                 │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("═══ 渲染机制说明 ═══").
				Foreground(theme.Primary()).Bold(true),
			newtext.New(""),
			newtext.New("Fiber-first 渲染流程："),
			newtext.New(""),
			newtext.New("  1️⃣  ComputeLayout() - 计算各层布局树"),
			newtext.New("       • 每个层独立计算布局"),
			newtext.New("       • 支持不同约束条件"),
			newtext.New(""),
			newtext.New("  2️⃣  PaintPaintablePlanes() - 分层绘制"),
			newtext.New("       • 创建独立的 PaintableBuffer"),
			newtext.New("       • 按 renderOrder 顺序绘制"),
			newtext.New("       • Layer 0 → Layer 4"),
			newtext.New(""),
			newtext.New("  3️⃣  Buffer 合成"),
			newtext.New("       • 各层 buffer 按顺序叠加"),
			newtext.New("       • 高层覆盖低层内容"),
			newtext.New("       • 保持透明区域可见"),
			newtext.New(""),
			newtext.New("  4️⃣  Modal 背景灰化"),
			newtext.New("       • 检测 Modal 区域（Layer 2）"),
			newtext.New("       • 灰化下方区域（Layer 0,1）"),
			newtext.New("       • 灰化左右区域（同层级其他元素）"),
			newtext.New("       • 保留原有内容，仅改变颜色"),
			newtext.New(""),

			// ┌─────────────────────────────────────────────────────────────┐
			// │  关键技术点                                                   │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("═══ 关键技术点 ═══").
				Foreground(theme.Primary()).Bold(true),
			newtext.New(""),
			newtext.New("✅ 分层渲染："),
			newtext.New("   • 每层独立 buffer，避免相互干扰"),
			newtext.New("   • Layout 和 Paint 完全解耦"),
			newtext.New(""),
			newtext.New("✅ Z-Order 控制："),
			newtext.New("   • 通过 renderOrder 数组定义渲染顺序"),
			newtext.New("   • 确保高层级正确覆盖低层级"),
			newtext.New(""),
			newtext.New("✅ 优化策略："),
			newtext.New("   • 只渲染有变化的层（脏标记）"),
			newtext.New("   • 减少 buffer 复制和合成开销"),
			newtext.New(""),
			newtext.New("═══════════════════════════════════════════════════════════════"),
		})
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║   Layer Rendering Demo                                ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")

	fwApp := framework.NewApp()
	fwApp.Resize(80, 50)

	declarativeNode := render.NewDeclarativeNodeFromFuncWithFiber(LayerRenderApp)
	declarativeNode.SetApp(fwApp)

	fwApp.SetRoot(declarativeNode)

	fmt.Println("")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("  本演示展示：")
	fmt.Println("    • 5 个渲染层级（Layer 0-4）的正确渲染顺序")
	fmt.Println("    • Fiber-first 分层渲染机制")
	fmt.Println("    • 多个层级的布局和绘制流程")
	fmt.Println("")
	fmt.Println("  技术点：")
	fmt.Println("    • PaintPaintablePlanes() - 分层绘制")
	fmt.Println("    • 独立的 buffer 管理")
	fmt.Println("    • Modal 背景灰化机制")
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
