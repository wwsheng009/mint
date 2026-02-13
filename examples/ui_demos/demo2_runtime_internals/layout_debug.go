package main

import (
	"fmt"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	"github.com/wwsheng009/mint/runtime/debug"
	"github.com/wwsheng009/mint/runtime/ui"
)

func GetLayoutTree(root ui.VNode, constraints runtime.BoxConstraints) (*debug.LayoutTree, error) {
	fmt.Println("==============================================")
	fmt.Println("Demo2 布局调试信息")
	fmt.Println("==============================================")
	fmt.Println()

	// 创建布局引擎
	engine := compute.NewEngine()

	// 执行布局计算
	layout, err := engine.Layout(root, nil, constraints)
	if err != nil {
		return nil, err
	}

	// 提取布局信息
	tree := debug.GetLayoutTree(layout)

	return tree, nil
}

// TestControlPanelLayout builds a simplified ControlPanel for layout testing
// This avoids hooks which require the UI framework to be running
func TestControlPanelLayout() ui.VNode {
	// Create all buttons (simplified, without click handlers)
	allButtons := []ui.VNode{
		app.ButtonBuilder("[1] Event").FocusStyle(app.FocusStyleBracket).Build(),
		app.ButtonBuilder("[2]setState").FocusStyle(app.FocusStyleBracket).Build(),
		app.ButtonBuilder("[3]Scheduler").FocusStyle(app.FocusStyleBracket).Build(),
		app.ButtonBuilder("[4] Render").FocusStyle(app.FocusStyleBracket).Build(),
		app.ButtonBuilder("[5]Reconcile").FocusStyle(app.FocusStyleBracket).Build(),
		app.ButtonBuilder("[6] Layout").FocusStyle(app.FocusStyleBracket).Build(),
		app.ButtonBuilder("[7] Paint").FocusStyle(app.FocusStyleBracket).Build(),
		app.ButtonBuilder("[0] Idle").FocusStyle(app.FocusStyleBracket).Build(),
	}

	// Use Wrap component for automatic wrapping (same as demo2)
	wrappedButtons := app.WrapBuilder(allButtons...).
		Gap(1).
		RowGap(0).
		ScreenWidth(78).
		Align(ui.AlignCenter).
		FillWidth().
		Build()

	return ui.Bordered().
		Child(wrappedButtons).
		FillWidth().
		Build()
}

func TestLayoutInfo() {
	fmt.Println("==============================================")
	fmt.Println("Demo2 布局调试信息")
	fmt.Println("==============================================")
	fmt.Println()

	// 获取简化的 ControlPanel（不使用 hooks）
	root := TestControlPanelLayout()

	// 设置约束 (与 demo2 相同)
	constraints := runtime.NewBoxConstraints(0, 100, 0, 35)

	// 获取布局树
	tree, err := GetLayoutTree(root, constraints)
	if err != nil {
		fmt.Printf("❌ 布局计算失败: %v\n", err)
		return
	}

	// 1. 打印完整的布局树
	fmt.Println("📋 完整布局树:")
	fmt.Println(debug.FormatLayoutTree(tree))

	// 2. 查找所有按钮
	fmt.Println("\n🔘 所有按钮:")
	fmt.Println("────────────────────────────────────────────────────────────────")
	buttons := debug.FindComponentsByType(tree, "button")
	fmt.Printf("找到 %d 个按钮:\n\n", len(buttons))

	for i, btn := range buttons {
		fmt.Printf("%d. %s\n", i+1, btn.Label)
		fmt.Printf("   路径: %s\n", btn.Path)
		fmt.Printf("   位置: (%d, %d)\n", btn.X, btn.Y)
		fmt.Printf("   尺寸: %dx%d\n", btn.Width, btn.Height)

		if btn.Flex > 0 {
			fmt.Printf("   Flex: %d ✅\n", btn.Flex)
		} else {
			fmt.Printf("   Flex: 无 ❌\n")
		}

		if btn.Padding != [4]int{} {
			fmt.Printf("   Padding: [top=%d, right=%d, bottom=%d, left=%d]\n",
				btn.Padding[0], btn.Padding[1], btn.Padding[2], btn.Padding[3])
		}
		fmt.Println()
	}

	// 3. 查找所有布局容器
	fmt.Println("📦 布局容器 (HStack/VStack):")
	fmt.Println("────────────────────────────────────────────────────────────────")
	hstacks := debug.FindComponentsByType(tree, "hstack")
	vstacks := debug.FindComponentsByType(tree, "vstack")

	fmt.Printf("找到 %d 个 HStack:\n", len(hstacks))
	for i, hs := range hstacks {
		fmt.Printf("  %d. 路径=%s\n", i+1, hs.Path)
		fmt.Printf("     位置: (%d, %d)\n", hs.X, hs.Y)
		fmt.Printf("     尺寸: %dx%d\n", hs.Width, hs.Height)
		fmt.Printf("     Gap: %d\n", hs.Gap)
		if hs.Align != "" {
			fmt.Printf("     Align: %s\n", hs.Align)
		}
		fmt.Printf("     子组件: %d 个\n", len(hs.Children))
		fmt.Println()
	}

	fmt.Printf("找到 %d 个 VStack:\n", len(vstacks))
	for i, vs := range vstacks {
		fmt.Printf("  %d. 路径=%s\n", i+1, vs.Path)
		fmt.Printf("     位置: (%d, %d)\n", vs.X, vs.Y)
		fmt.Printf("     尺寸: %dx%d\n", vs.Width, vs.Height)
		fmt.Printf("     Gap: %d\n", vs.Gap)
		fmt.Printf("     子组件: %d 个\n", len(vs.Children))
		fmt.Println()
	}

	// 4. 统计信息
	fmt.Println("\n📊 统计信息:")
	fmt.Println("────────────────────────────────────────────────────────────────")

	typeCount := make(map[string]int)
	countComponents(&tree.Root, typeCount)

	fmt.Println("组件类型统计:")
	for t, count := range typeCount {
		fmt.Printf("  %s: %d\n", t, count)
	}

	// 5. 检查按钮分布
	fmt.Println("\n🔍 ControlPanel 按钮分布分析:")
	fmt.Println("────────────────────────────────────────────────────────────────")

	// 找到包含最多按钮的 HStack
	var maxButtonHStack debug.LayoutInfo
	maxButtonCount := 0

	for _, hs := range hstacks {
		buttonCount := 0
		for _, child := range hs.Children {
			if child.Type == "button" {
				buttonCount++
			}
		}
		if buttonCount > maxButtonCount && buttonCount > 1 {
			maxButtonCount = buttonCount
			maxButtonHStack = hs
		}
	}

	if maxButtonHStack.Type != "" {
		fmt.Printf("找到包含 %d 个按钮的 HStack:\n", maxButtonCount)
		fmt.Printf("  路径: %s\n", maxButtonHStack.Path)
		fmt.Printf("  位置: (%d, %d)\n", maxButtonHStack.X, maxButtonHStack.Y)
		fmt.Printf("  尺寸: %dx%d\n", maxButtonHStack.Width, maxButtonHStack.Height)
		fmt.Printf("  Gap: %d\n", maxButtonHStack.Gap)
		fmt.Println()

		// 分析每个按钮的宽度
		totalWidth := 0
		widths := make([]int, 0, maxButtonCount)

		for i, child := range maxButtonHStack.Children {
			if child.Type == "button" {
				totalWidth += child.Width
				widths = append(widths, child.Width)
				fmt.Printf("  按钮 %d (%s): 位置=(%d, %d), 尺寸=%dx%d, Flex=%d\n",
					i+1, child.Label, child.X, child.Y, child.Width, child.Height, child.Flex)
			}
		}

		fmt.Printf("\n  按钮总宽度: %d\n", totalWidth)
		if maxButtonCount > 0 {
			avgWidth := totalWidth / maxButtonCount
			minWidth := widths[0]
			maxWidth := widths[0]

			for _, w := range widths {
				if w < minWidth {
					minWidth = w
				}
				if w > maxWidth {
					maxWidth = w
				}
			}

			fmt.Printf("  平均宽度: %d\n", avgWidth)
			fmt.Printf("  宽度范围: %d - %d (差异: %d)\n", minWidth, maxWidth, maxWidth-minWidth)

			// 检查是否均匀分布
			if maxWidth-minWidth <= 1 {
				fmt.Println("  ✅ 按钮均匀分布 (宽度差异 ≤ 1)")
			} else {
				fmt.Printf("  ⚠️ 按钮宽度不均匀 (差异: %d)\n", maxWidth-minWidth)
			}
		}
	}
}

func countComponents(info *debug.LayoutInfo, typeCount map[string]int) {
	typeCount[info.Type]++
	if info.Tag != "" {
		typeCount[info.Tag]++
	}
	for _, child := range info.Children {
		countComponents(&child, typeCount)
	}
}
