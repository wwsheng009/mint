// Layer Z-Order Test Demo
// Layer Z-Order 测试 - 验证层级顺序和覆盖效果
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

// LayerZOrderTestApp 主应用 VNode
func LayerZOrderTestApp() ui.VNode {
	return ui.NewVStack().
		SetGap(0).
		SetChildrenList([]ui.VNode{
			// ┌─────────────────────────────────────────────────────────────┐
			// │  标题栏                                                       │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("╔════════════════════════════════════════════════════════╗").
				Foreground(theme.Primary()),
			newtext.New("║   Layer Z-Order Test Demo                            ║").
				Foreground(theme.Primary()).Bold(true),
			newtext.New("╚════════════════════════════════════════════════════════╝").
				Foreground(theme.Primary()),

			newtext.New(""),
			newtext.New("  测试验证：层级顺序和覆盖效果"),
			newtext.New("  按 ESC 或 Ctrl+C 退出程序"),
			newtext.New(""),

			// ┌─────────────────────────────────────────────────────────────┐
			// │  测试用例 1: 基础层级顺序                                      │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("═══ 测试用例 1: 基础层级顺序 ═══").
				Foreground(theme.Primary()).Bold(true),
			newtext.New(""),
			newtext.New("预期：从下到上依次显示，高层覆盖低层"),
			newtext.New(""),
			newtext.New("【第 1 层】Layer 0 (Base) - 应该位于最底层"),
			newtext.New("  期望：被上方元素覆盖"),
			tooltipcomp.NewToastBuilder("◆ Layer 0: Base").
				Layer(ui.LayerBase).
				Build(),

			newtext.New(""),

			newtext.New("【第 2 层】Layer 1 (Overlay) - 应该覆盖 Layer 0"),
			newtext.New("  期望：与 Layer 0 部分重叠时可见"),
			tooltipcomp.NewToastBuilder("◆ Layer 1: Overlay").
				Layer(ui.LayerOverlay).
				Info().
				Build(),

			newtext.New(""),

			newtext.New("【第 3 层】Layer 2 (Modal) - 应该覆盖 Layer 1"),
			newtext.New("  期望：比 Overlay 更靠上"),
			tooltipcomp.NewToastBuilder("◆ Layer 2: Modal").
				Layer(ui.LayerModal).
				Warning().
				Build(),

			newtext.New(""),

			newtext.New("【第 4 层】Layer 3 (Tooltip) - 应该覆盖 Layer 2"),
			newtext.New("  期望：比 Modal 更靠上"),
			tooltipcomp.NewToastBuilder("◆ Layer 3: Tooltip").
				Layer(ui.LayerTooltip).
				Build(),

			newtext.New(""),

			newtext.New("【第 5 层】Layer 4 (Inspector) - 应该位于最顶层"),
			newtext.New("  期望：覆盖所有下方元素"),
			tooltipcomp.NewToastBuilder("◆ Layer 4: Inspector (最高)").
				Layer(ui.LayerInspector).
				Success().
				Build(),

			newtext.New(""),

			// ┌─────────────────────────────────────────────────────────────┐
			// │  测试用例 2: 同位置覆盖测试                                    │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("═══ 测试用例 2: 同位置覆盖测试 ═══").
				Foreground(theme.Primary()).Bold(true),
			newtext.New(""),
			newtext.New("预期：相同位置的多层元素，顶层可见"),
			newtext.New(""),
			newtext.New("说明：下面 3 个 Toast 在接近的视觉位置，验证层级效果"),
			newtext.New(""),

			newtext.New("【同位置测试】Layer 0（底部）"),
			tooltipcomp.NewToastBuilder("⬇ Layer 0").
				Layer(ui.LayerBase).
				Build(),

			newtext.New(""),

			newtext.New("【同位置测试】Layer 2（中部）"),
			tooltipcomp.NewToastBuilder("⬇ Layer 2").
				Layer(ui.LayerModal).
				Build(),

			newtext.New(""),

			newtext.New("【同位置测试】Layer 4（顶部）"),
			tooltipcomp.NewToastBuilder("▲ Layer 4").
				Layer(ui.LayerInspector).
				Success().
				Build(),

			newtext.New(""),

			// ┌─────────────────────────────────────────────────────────────┐
			// │  测试用例 3: Modal 背景灰化测试                                │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("═══ 测试用例 3: Modal 背景灰化 ═══").
				Foreground(theme.Primary()).Bold(true),
			newtext.New(""),
			newtext.New("预期：Modal 下方的区域被灰色覆盖"),
			newtext.New(""),
			newtext.New("【Modal 测试】背景灰化区域"),
			tooltipcomp.NewToastBuilder("Modal 对话框 - 下方区域应被灰化").
				Layer(ui.LayerModal).
				Warning().
				Build(),

			newtext.New(""),
			newtext.New("  检查点："),
			newtext.New("    ✓ Modal 上方区域 - 应该灰色"),
			newtext.New("    ✓ Modal 左侧区域 - 应该灰色"),
			newtext.New("    ✓ Modal 右侧区域 - 应该灰色"),
			newtext.New("    ✓ Modal 下方区域 - 应该灰色"),
			newtext.New("    ✓ Modal 内部区域 - 保持原色"),
			newtext.New(""),

			// ┌─────────────────────────────────────────────────────────────┐
			// │  测试用例 4: 中文显示测试                                      │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("═══ 测试用例 4: 中文显示 ═══").
				Foreground(theme.Primary()).Bold(true),
			newtext.New(""),
			newtext.New("预期：Modal 区域中文字符正确显示"),
			newtext.New(""),

			newtext.New("【中文测试】Layer 2 中文字符"),
			tooltipcomp.NewToastBuilder("这是一个中文测试的 Modal 对话框").
				Layer(ui.LayerModal).
				Warning().
				Build(),

			newtext.New(""),

			newtext.New("  检查点："),
			newtext.New("    ✓ 中文字符宽度正确（每个字符占 2 列）"),
			newtext.New("    ✓ 中文字符不出现乱码或截断"),
			newtext.New("    ✓ Modal 背景灰化不破坏中文字符连续性"),
			newtext.New(""),
			newtext.New(""),

			// ┌─────────────────────────────────────────────────────────────┐
			// │  测试结果确认                                                 │
			// └─────────────────────────────────────────────────────────────┘
			newtext.New("═══ 测试结果确认 ═══").
				Foreground(theme.Primary()).Bold(true),
			newtext.New(""),
			newtext.New("请手动验证："),
			newtext.New(""),
			newtext.New("✅ 测试 1: 层级顺序"),
			newtext.New("   [ ] Layer 0 的内容在最底层"),
			newtext.New("   [ ] Layer 4 的内容在最顶层"),
			newtext.New("   [ ] 层级按 0→1→2→3→4 顺序排列"),
			newtext.New(""),
			newtext.New("✅ 测试 2: 同位置覆盖"),
			newtext.New("   [ ] 相同位置的多层元素，顶层可见"),
			newtext.New("   [ ] 低层元素被正确覆盖"),
			newtext.New(""),
			newtext.New("✅ 测试 3: Modal 背景灰化"),
			newtext.New("   [ ] Modal 四周区域正确灰化"),
			newtext.New("   [ ] Modal 内部内容保持原色"),
			newtext.New(""),
			newtext.New("✅ 测试 4: 中文显示"),
			newtext.New("   [ ] 中文在 Modal 中正确显示"),
			newtext.New("   [ ] 中文字符宽度正确（2 列）"),
			newtext.New("   [ ] Modal 背景灰化不破坏中文"),
			newtext.New(""),
			newtext.New("═══════════════════════════════════════════════════════════════"),
		})
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║   Layer Z-Order Test Demo                             ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")

	fwApp := framework.NewApp()
	fwApp.Resize(80, 60)

	declarativeNode := render.NewDeclarativeNodeFromFuncWithFiber(
		LayerZOrderTestApp,
		fwApp,
	)

	fwApp.SetRoot(declarativeNode)

	fmt.Println("")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("  本演示执行多项测试：")
	fmt.Println("    • 测试 1: 基础层级顺序验证")
	fmt.Println("    • 测试 2: 同位置覆盖效果")
	fmt.Println("    • 测试 3: Modal 背景灰化")
	fmt.Println("    • 测试 4: 中文字符显示")
	fmt.Println("")
	fmt.Println("  请在运行应用后，对照视觉输出确认各测试项：")
	fmt.Println("    ✓ 所有层级元素按正确顺序显示")
	fmt.Println("    ✓ Modal 背景灰化区域正确")
	fmt.Println("    ✓ 中文在 Modal 中正常显示")
	fmt.Println("")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("")
	fmt.Println("测试演示已启动... 按 ESC 或 Ctrl+C 退出")
	fmt.Println("")

	if err := fwApp.Run(); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
}
