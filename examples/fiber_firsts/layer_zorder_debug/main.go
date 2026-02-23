// Layer Z-Order Debug Demo
// Layer Z-Order 调试 - 详细输出层级信息用于调试
package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/ui"
	newstack "github.com/wwsheng009/mint/ui/components/stack"
	newtext "github.com/wwsheng009/mint/ui/components/text"
	tooltipcomp "github.com/wwsheng009/mint/ui/components/tooltip"
)

// LayerZOrderDebugApp 主应用 VNode
func LayerZOrderDebugApp() ui.VNode {
	return newstack.NewVStack().
		SetGap(0).
		SetChildrenList([]ui.VNode{
			// ┌─────────────────────────────────────────────────────────────┐
			// │  标题栏                                                       │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("╔════════════════════════════════════════════════════════╗").
				Foreground(theme.Primary()),
			newtext.New("║   Layer Z-Order Debug Demo                           ║").
				Foreground(theme.Primary()).Bold(true),
			newtext.New("╚════════════════════════════════════════════════════════╝").
				Foreground(theme.Primary()),

			newtext.New(""),
			newtext.New("  调试用：详细输出层级信息和渲染过程"),
			newtext.New("  按 ESC 或 Ctrl+C 退出程序"),
			newtext.New(""),

			// ┌─────────────────────────────────────────────────────────────┐
			// │  层级定义                                                     │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("═══ Layer 定义 (ui/layer.go) ═══").
				Foreground(theme.Primary()).Bold(true),
			newtext.New(""),
			newtext.New("  const ("),
			newtext.New("      LayerBase       Layer = 0  // 主要内容层"),
			newtext.New("      LayerOverlay    Layer = 1  // 标准通知层"),
			newtext.New("      LayerModal      Layer = 2  // 模态对话框层"),
			newtext.New("      LayerTooltip    Layer = 3  // 悬浮提示层"),
			newtext.New("      LayerInspector  Layer = 4  // 调试覆盖层"),
			newtext.New("  )"),
			newtext.New(""),
			newtext.New("  renderOrder = [LayerBase, LayerOverlay, LayerModal, LayerTooltip, LayerInspector]"),
			newtext.New(""),
			newtext.New("  ┌────────────────────────────────────────────────────────────┐"),
			newtext.New("  │  渲染顺序                                                     │"),
			newtext.New("  │  ↓     先绘制 Layer 0                                     │"),
			newtext.New("  │  ↓ 然后  Layer 1                                         │"),
			newtext.New("  │  ↓ 然后  Layer 2                                         │"),
			newtext.New("  │  ↓ 然后  Layer 3                                         │"),
			newtext.New("  │  ↓ 最后  Layer 4 (覆盖所有)                              │"),
			newtext.New("  └────────────────────────────────────────────────────────────┘"),
			newtext.New(""),

			// ┌─────────────────────────────────────────────────────────────┐
			// │  调试输出区域                                                 │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("═══ 调试验证区域 ═══").
				Foreground(theme.Primary()).Bold(true),
			newtext.New(""),

			newtext.New("【Layer 0】Base - 底层内容"),
			newtext.New("  • 应该：最先绘制，位于最底层"),
			newtext.New("  • 验证：被其他层元素覆盖"),
			tooltipcomp.NewToastBuilder("Layer 0: Base (底层)").
				Layer(ui.LayerBase).
				Build(),

			newtext.New(""),

			newtext.New("【Layer 2】Modal - 模态层"),
			newtext.New("  • 应该：显示在 Base 和 Overlay 之上"),
			newtext.New("  • 验证：下方内容被灰色背景覆盖"),
			newtext.New("  • 机制：paintModalBackdropBox() 灰化"),
			tooltipcomp.NewToastBuilder("Layer 2: Modal (模态)").
				Layer(ui.LayerModal).
				Warning().
				Build(),

			newtext.New(""),

			newtext.New("【Layer 4】Inspector - 最高层"),
			newtext.New("  • 应该：最后绘制，覆盖所有内容"),
			newtext.New("  • 验证：位于最顶层，完全可见"),
			tooltipcomp.NewToastBuilder("Layer 4: Inspector (最高)").
				Layer(ui.LayerInspector).
				Success().
				Build(),

			newtext.New(""),

			// ┌─────────────────────────────────────────────────────────────┐
			// │  调试检查点                                                   │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("═══ 调试检查点 ═══").
				Foreground(theme.Primary()).Bold(true),
			newtext.New(""),
			newtext.New("✅ 检查点 1: 渲染顺序"),
			newtext.New("   检查 PaintPaintablePlanes() 中的循环顺序"),
			newtext.New("   for _, layer := range []rtui.Layer{...}"),
			newtext.New(""),
			newtext.New("✅ 检查点 2: Buffer 合成"),
			newtext.New("   检查 buffer.DrawToMainBuffer() 是否按正确顺序"),
			newtext.New("   低层 buffer 先绘制，高层 buffer 后绘制"),
			newtext.New(""),
			newtext.New("✅ 检查点 3: Modal 背景灰化"),
			newtext.New("   检查 paintModalBackdropBox() 的逻辑"),
			newtext.New("   • 正确计算 modal 的 bounds (x, y, w, h)"),
			newtext.New("   • 灰化位置：上方、下方、左侧、右侧"),
			newtext.New("   • 跳过延续单元格（中文处理）"),
			newtext.New(""),
			newtext.New("✅ 检查点 4: 中文字符宽度"),
			newtext.New("   检查 runtime/paint/width.go 中的 EastAsianWidth"),
			newtext.New("   • 应该 = true 才能正确计算中文宽度"),
			newtext.New("   • 中文宽度 = 2，ASCII 宽度 = 1"),
			newtext.New(""),

			// ┌─────────────────────────────────────────────────────────────┐
			// │  常见问题排查                                                 │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("═══ 常见问题排查 ═══").
				Foreground(theme.Primary()).Bold(true),
			newtext.New(""),
			newtext.New("❌ 问题 1: Modal 不显示"),
			newtext.New("   检查: Modal 层的 buffer 是否有内容"),
			newtext.New("   检查: Modal 的 computeLayout() 是否返回有效的 bounds"),
			newtext.New(""),
			newtext.New("❌ 问题 2: 背景不灰化"),
			newtext.New("   检查: modalBox 是否为 nil"),
			newtext.New("   检查: modalY, modalX, modalHeight, modalWidth 值"),
			newtext.New(""),
			newtext.New("❌ 问题 3: 中文字符显示异常"),
			newtext.New("   检查: EastAsianWidth 是否 = true"),
			newtext.New("   检查: paintModalBackdropBox 是否跳过 continuation"),
			newtext.New(""),
			newtext.New("❌ 问题 4: 层级错乱"),
			newtext.New("   检查: renderOrder 数组的顺序"),
			newtext.New("   检查: 各层 VNode.GetLayer() 返回值"),
			newtext.New(""),
			newtext.New("═══════════════════════════════════════════════════════════════"),
		})
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")
	os.Setenv("MINT_DEBUG_LAYER", "true")

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║   Layer Z-Order Debug Demo                            ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")

	fwApp := framework.NewApp()
	fwApp.Resize(80, 55)

	declarativeNode := render.NewDeclarativeNodeFromFuncWithFiber(
		LayerZOrderDebugApp,
		fwApp,
	)

	fwApp.SetRoot(declarativeNode)

	fmt.Println("")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("  本演示用于调试 Layer 相关问题：")
	fmt.Println("    • 层级渲染顺序（Z-Order）")
	fmt.Println("    • Buffer 合成机制")
	fmt.Println("    • Modal 背景灰化")
	fmt.Println("    • 中文字符处理")
	fmt.Println("")
	fmt.Println("  环境变量：")
	fmt.Println("    MINT_DEBUG_LAYER=true - 启用层级调试输出")
	fmt.Println("")
	fmt.Println("  检查点：")
	fmt.Println("    1. 验证 Layer 0 的内容被正确覆盖")
	fmt.Println("    2. 验证 Modal 背景区域正确灰化")
	fmt.Println("    3. 验证 Layer 4 的内容在最顶层")
	fmt.Println("")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("")
	fmt.Println("调试演示已启动... 按 ESC 或 Ctrl+C 退出")
	fmt.Println("")

	if err := fwApp.Run(); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
}
