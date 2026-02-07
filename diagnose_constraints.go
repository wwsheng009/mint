package main

import (
	"fmt"
	"github.com/wwsheng009/mint/app"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	fmt.Println("=== 诊断约束传递 ===\n")

	// 重建 elegant_api_demo 的结构
	text1 := ui.Text("✨ Elegant VNode Builder API Demo")
	text2 := ui.Text("────────────────────────────────")
	text3 := ui.Text("1. Flex buttons (no SetProp needed):")

	hstack := ui.HStackBuilder(
		app.ButtonBuilder("Left").PaddingH(1, 2).Flex(1).SetTextAlign(rtui.AlignStart).Build(),
		app.ButtonBuilder("Center").PaddingH(1, 1).Flex(1).SetTextAlign(rtui.AlignCenter).Build(),
		app.ButtonBuilder("Right").PaddingH(2, 1).Flex(1).SetTextAlign(rtui.AlignEnd).Build(),
	).Gap(1).Build()

	vstack := ui.VStackBuilder(text1, text2, text3, hstack).Gap(0).Build()

	// 测试 1: VStack 在 80 宽度容器中
	fmt.Println("Test 1: Measure VStack with 80-width constraint")
	fmt.Println("─────────────────────────────────────────────────")
	size1 := measureVNode(vstack, 0, 80)
	fmt.Printf("VStack size: Width=%d, Height=%d\n", size1.Width, size1.Height)
	fmt.Printf("Expected: Width ~80 (should fill container)\n")
	fmt.Println()

	// 测试 2: HStack 在 80 宽度容器中
	fmt.Println("Test 2: Measure HStack with 80-width constraint")
	fmt.Println("─────────────────────────────────────────────────")
	size2 := measureVNode(hstack, 0, 80)
	fmt.Printf("HStack size: Width=%d, Height=%d\n", size2.Width, size2.Height)
	fmt.Printf("Expected: Width ~77 (80 - padding)\n")
	fmt.Println()

	// 测试 3: HStack 无约束
	fmt.Println("Test 3: Measure HStack with NO constraint")
	fmt.Println("─────────────────────────────────────────────────")
	size3 := measureVNode(hstack, 0, runtime.Infinity)
	fmt.Printf("HStack size: Width=%d, Height=%d\n", size3.Width, size3.Height)
	fmt.Printf("Expected: Width ~32 (natural width of buttons)\n")
	fmt.Println()

	// 测试 4: 单个按钮在 26 宽度约束下
	fmt.Println("Test 4: Measure individual button with 26-width constraint")
	fmt.Println("─────────────────────────────────────────────────")
	btn := app.ButtonBuilder("Left").PaddingH(1, 2).Flex(1).SetTextAlign(rtui.AlignStart).Build()
	size4 := measureVNode(btn, 26, 26)
	fmt.Printf("Button size: Width=%d, Height=%d\n", size4.Width, size4.Height)
	fmt.Printf("Expected: Width=26 (stretched to fit)\n")
	fmt.Println()

	// 结论
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("结论:")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	if size2.Width > 70 {
		fmt.Println("✅ HStack 正确响应 80 宽度约束")
		fmt.Println("   → Flex 子元素应该被拉伸")
	} else {
		fmt.Println("❌ HStack 没有正确响应 80 宽度约束")
		fmt.Println("   → Flex 子元素保持自然宽度")
		fmt.Println()
		fmt.Println("可能原因:")
		fmt.Println("1. 布局引擎没有正确传递约束给 HStack")
		fmt.Println("2. HStack 的 Measure 实现有 bug")
		fmt.Println("3. 实际运行时的约束与我们测试的不同")
	}
}

func measureVNode(vnode ui.VNode, minWidth, maxWidth int) runtime.Size {
	if measurable, ok := vnode.(interface {
		Measure(constraints runtime.BoxConstraints) runtime.Size
	}); ok {
		return measurable.Measure(runtime.BoxConstraints{
			MinWidth:  minWidth,
			MaxWidth:  maxWidth,
			MinHeight: 0,
			MaxHeight: 1000,
		})
	}
	return runtime.Size{Width: 0, Height: 0}
}
