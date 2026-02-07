// Mint TUI Flex Examples
// 展示如何像 CSS Flexbox 一样使用 Mint TUI 布局系统

package main

import (
	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// 示例 1: 基本 Flex 容器
// =============================================================================

// CSS: display: flex; flex-direction: row;
func Example1_HStack() ui.VNode {
	return ui.HStack(
		app.NewTextBuilder("Item 1").Build(),
		app.NewTextBuilder("Item 2").Build(),
		app.NewTextBuilder("Item 3").Build(),
	)
}

// CSS: display: flex; flex-direction: column;
func Example1_VStack() ui.VNode {
	return ui.VStack(
		app.NewTextBuilder("Item 1").Build(),
		app.NewTextBuilder("Item 2").Build(),
		app.NewTextBuilder("Item 3").Build(),
	)
}

// =============================================================================
// 示例 2: 主轴对齐 (justify-content → .Align())
// =============================================================================

// CSS: justify-content: flex-start;
func Example2_AlignStart() ui.VNode {
	return ui.HStackBuilder(
		app.NewTextBuilder("Item 1").Build(),
		app.NewTextBuilder("Item 2").Build(),
		app.NewTextBuilder("Item 3").Build(),
	).
		Align(ui.AlignStart). // 相当于 justify-content: flex-start
		Build()
}

// CSS: justify-content: center;
func Example2_AlignCenter() ui.VNode {
	return ui.HStackBuilder(
		app.NewTextBuilder("Item 1").Build(),
		app.NewTextBuilder("Item 2").Build(),
		app.NewTextBuilder("Item 3").Build(),
	).
		Align(ui.AlignCenter). // 相当于 justify-content: center
		Build()
}

// CSS: justify-content: flex-end;
func Example2_AlignEnd() ui.VNode {
	return ui.HStackBuilder(
		app.NewTextBuilder("Item 1").Build(),
		app.NewTextBuilder("Item 2").Build(),
		app.NewTextBuilder("Item 3").Build(),
	).
		Align(ui.AlignEnd). // 相当于 justify-content: flex-end
		Build()
}

// CSS: justify-content: space-between;
func Example2_SpaceBetween() ui.VNode {
	return ui.HStackBuilder(
		app.NewTextBuilder("Item 1").Build(),
		app.NewTextBuilder("Item 2").Build(),
		app.NewTextBuilder("Item 3").Build(),
	).
		Align(ui.AlignSpaceBetween). // 相当于 justify-content: space-between
		Build()
}

// CSS: justify-content: space-around;
func Example2_SpaceAround() ui.VNode {
	return ui.HStackBuilder(
		app.NewTextBuilder("Item 1").Build(),
		app.NewTextBuilder("Item 2").Build(),
		app.NewTextBuilder("Item 3").Build(),
	).
		Align(ui.AlignSpaceAround). // 相当于 justify-content: space-around
		Build()
}

// =============================================================================
// 示例 3: 交叉轴对齐 (align-items → .AlignCross())
// =============================================================================

// CSS: align-items: center;
func Example3_CrossAlignCenter() ui.VNode {
	return ui.HStackBuilder(
		app.NewTextBuilder("Short").Build(),
		app.NewTextBuilder("Medium Text").Build(),
		app.NewTextBuilder("Very Long Text Here").Build(),
	).
		Align(ui.AlignStart).
		AlignCross(ui.AlignCenter). // 相当于 align-items: center
		Build()
}

// =============================================================================
// 示例 4: 间距 (gap)
// =============================================================================

// CSS: gap: 8px;
func Example4_Gap() ui.VNode {
	return ui.HStackBuilder(
		app.NewTextBuilder("Item 1").Build(),
		app.NewTextBuilder("Item 2").Build(),
		app.NewTextBuilder("Item 3").Build(),
	).
		Gap(2). // 2 个空格的间距
		Build()
}

// =============================================================================
// 示例 5: Flex Grow (拉伸因子)
// =============================================================================

// CSS: .left { width: 200px; } .right { width: 200px; } .center { flex-grow: 1; }
func Example5_FlexGrow() ui.VNode {
	return ui.HStack(
		ui.Box().Width(20).Build(),              // 左侧固定
		ui.Box().Flex(1).Build(),                // 中间自适应 (flex-grow: 1)
		ui.Box().Width(20).Build(),              // 右侧固定
	)
}

// CSS: flex-grow 比例
func Example5_FlexRatio() ui.VNode {
	return ui.HStack(
		ui.Box().Flex(1).Build(), // 1 份
		ui.Box().Flex(2).Build(), // 2 份
		ui.Box().Flex(3).Build(), // 3 份
	)
}

// =============================================================================
// 示例 6: 拉伸到父容器 (width/height: 100%)
// =============================================================================

// CSS: width: 100%;
func Example6_FillWidth() ui.VNode {
	return ui.VStack(
		ui.Bordered().
			Child(app.NewTextBuilder("I stretch to full width").Build()).
			FillWidth(). // width: 100%
			Build(),
		ui.Box().Build(), // 保持原宽
	)
}

// CSS: height: 100%;
func Example6_FillHeight() ui.VNode {
	return ui.HStack(
		ui.Bordered().
			Child(app.NewTextBuilder("I stretch to full height").Build()).
			FillHeight(). // height: 100%
			Build(),
		ui.Box().Build(), // 保持原高
	)
}

// =============================================================================
// 示例 7: 完全居中 (水平和垂直)
// =============================================================================

// CSS:
// .container {
//   display: flex;
//   justify-content: center;
//   align-items: center;
//   width: 100%;
//   height: 100vh;
// }
func Example7_PerfectCenter() ui.VNode {
	return ui.HStackBuilder(
		app.NewTextBuilder("Perfectly Centered").Build(),
	).
		Align(ui.AlignCenter).      // 水平居中 (justify-content: center)
		AlignCross(ui.AlignCenter). // 垂直居中 (align-items: center)
		FillWidth().                // width: 100%
		FillHeight().               // height: 100vh
		Build()
}

// =============================================================================
// 示例 8: Padding (内边距)
// =============================================================================

// CSS: padding: 10px 20px;
func Example8_Padding() ui.VNode {
	return ui.HStackBuilder(
		app.NewTextBuilder("Content").Build(),
	).
		Padding(1, 2, 1, 2). // 上, 右, 下, 左 (单位：行/字符)
		Build()
}

// =============================================================================
// 示例 9: 嵌套布局 (复杂布局)
// =============================================================================

// CSS:
// .app {
//   display: flex;
//   flex-direction: column;
//   height: 100vh;
// }
// .header { flex: 0 0 60px; }
// .main { flex: 1; }
// .footer { flex: 0 0 40px; }
func Example9_NestedLayout() ui.VNode {
	return ui.VStack(
		// Header: 固定高度
		ui.Bordered().
			Child(app.NewTextBuilder("Header").Build()).
			Height(3). // 3 行高
			Build(),

		// Main: 填充剩余空间
		ui.VStack(
			app.NewTextBuilder("Main Content").Build(),
			app.NewTextBuilder("More Content").Build(),
		).
			Flex(1). // flex: 1 (填充剩余垂直空间)
			Build(),

		// Footer: 固定高度
		ui.Bordered().
			Child(app.NewTextBuilder("Footer").Build()).
			Height(2). // 2 行高
			Build(),
	).
		Gap(0).
		FillHeight(). // height: 100vh
		Build()
}

// =============================================================================
// 示例 10: 混合对齐 (space-between + center)
// =============================================================================

// CSS:
// .container {
//   display: flex;
//   justify-content: space-between;
//   align-items: center;
// }
func Example10_MixedAlignment() ui.VNode {
	return ui.HStackBuilder(
		app.NewTextBuilder("Left").Build(),
		app.NewTextBuilder("Right").Build(),
	).
		Align(ui.AlignSpaceBetween). // 两端对齐
		AlignCross(ui.AlignCenter).  // 垂直居中
		FillWidth().                  // 填充宽度
		Build()
}

// =============================================================================
// 示例 11: Stretch vs FillWidth
// =============================================================================

// CSS (模拟):
// .container-stretch { align-items: stretch; } // 所有子元素拉伸
// .single-fill { width: 100%; }               // 单个元素拉伸
func Example11_StretchVsFill() ui.VNode {
	// 方式 1: 使用 Stretch() - 所有子元素都横向拉伸
	container1 := ui.VStackBuilder(
		ui.Bordered().Child(app.NewTextBuilder("All items stretched").Build()).Build(),
		ui.Bordered().Child(app.NewTextBuilder("Including this one").Build()).Build(),
		ui.Bordered().Child(app.NewTextBuilder("And this one").Build()).Build(),
	).
		Stretch(). // 相当于 align-items: stretch (所有子元素拉伸宽度)
		Build()

	// 方式 2: 使用 FillWidth() - 只让特定元素拉伸
	container2 := ui.VStack(
		ui.Bordered().Child(app.NewTextBuilder("Not stretched").Build()).Build(),
		ui.Bordered().
			Child(app.NewTextBuilder("Only this stretched").Build()).
			FillWidth(). // 相当于 width: 100% (只这个元素拉伸)
			Build(),
		ui.Bordered().Child(app.NewTextBuilder("Not stretched").Build()).Build(),
	)

	return ui.VStack(container1, container2).Gap(1).Build()
}

// =============================================================================
// 示例 12: 实际应用 - 响应式头部 (demo2)
// =============================================================================

// 这是 demo2_runtime_internals 中的实际使用
func Example12_HeaderPanel() ui.VNode {
	titleContent := ui.HStackBuilder(
		app.NewTextBuilder("Runtime Scheduling Pipeline Visualization").
			Build(),
	).
		Gap(0).
		Align(ui.AlignCenter). // 标题居中
		Build()

	return ui.Bordered().
		Child(titleContent).
		FillWidth(). // 标题横向拉伸到整个屏幕
		Build()
}

// =============================================================================
// 示例 13: 实际应用 - 表单布局
// =============================================================================

func Example13_FormLayout() ui.VNode {
	return ui.VStack(
		// 表单标题
		ui.Bordered().
			Child(app.NewTextBuilder("User Information").Build()).
			FillWidth().
			Build(),

		// 表单字段
		ui.VStack(
			ui.HStack(
				app.NewTextBuilder("Name:").Build(),
				ui.Box().Width(30).Build(), // Input 占位符
			).Build(),
			ui.HStack(
				app.NewTextBuilder("Email:").Build(),
				ui.Box().Width(30).Build(), // Input 占位符
			).Build(),
		).
			Gap(1).
			Build(),

		// 按钮
		ui.HStack(
			app.ButtonBuilder("Submit").Build(),
			app.ButtonBuilder("Cancel").Build(),
		).
			Gap(2).
			Align(ui.AlignCenter).
			Build(),
	).
		Gap(1).
		Build()
}

// =============================================================================
// 示例 14: 实际应用 - 卡片网格 (模拟)
// =============================================================================

func Example14_CardGrid() ui.VNode {
	// 第一行
	row1 := ui.HStack(
		createCard("Card 1"),
		createCard("Card 2"),
		createCard("Card 3"),
	).
		Gap(1).
		Build()

	// 第二行
	row2 := ui.HStack(
		createCard("Card 4"),
		createCard("Card 5"),
		createCard("Card 6"),
	).
		Gap(1).
		Build()

	return ui.VStack(row1, row2).
		Gap(1).
		Build()
}

func createCard(title string) ui.VNode {
	return ui.Bordered().
		Child(app.NewTextBuilder(title).Build()).
		Width(30).  // 固定宽度
		Height(10). // 固定高度
		Build()
}

// =============================================================================
// 示例 15: 进度条 (比例布局)
// =============================================================================

func Example15_ProgressBar() ui.VNode {
	progress := 75 // 75%

	bar := ui.HStack(
		ui.Box().
			Width(progress). // 已完成部分
			Build(),
		ui.Box().
			Flex(1). // 未完成部分 (填充剩余空间)
			Build(),
	)

	return ui.Bordered().
		Child(bar).
		FillWidth().
		Build()
}
