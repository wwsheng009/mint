package main

import (
	ui "github.com/wwsheng009/mint/ui"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/app"
)

func main() {
	err := ui.Run(VisualDemo,
		ui.WithWidth(80),
		ui.WithHeight(15),
		ui.WithTitle("Box Model Visual Demo"),
	)
	if err != nil {
		panic(err)
	}
}

func VisualDemo() ui.VNode {
	return ui.VStackBuilder(
		ui.Text("┌─ Box Model Visual Test ─────────────────────────────────────┐"),
		ui.Text("│"),
		ui.Text("│ Example 1: Three flex buttons with different alignments"),
		ui.Text("│"),
		ui.HStackBuilder(
			ui.Flex(app.ButtonBuilder("Left").
				PaddingH(1, 2).
				TextAlign(rtui.AlignStart).
				Build(), 1),
			ui.Flex(app.ButtonBuilder("Center").
				PaddingH(1, 1).
				TextAlign(rtui.AlignCenter).
				Build(), 1),
			ui.Flex(app.ButtonBuilder("Right").
				PaddingH(2, 1).
				TextAlign(rtui.AlignEnd).
				Build(), 1),
		).
			Gap(1).
			// ⭐ 不再需要显式设置 Width，flex 容器会自动填充父容器宽度
			Build(),
		ui.Text("│"),
		ui.Text("│ Expected: [  Left  ]    [ Center ]    [  Right  ]"),
		ui.Text("│          (more pad)    (even pad)    (more pad)"),
		ui.Text("│"),
		ui.Text("│ Example 2: Button with padding and centering"),
		ui.Text("│"),
		ui.HStackBuilder(
			ui.Flex(app.ButtonBuilder("Padded & Centered").
				PaddingAll(2).
				TextAlign(rtui.AlignCenter).
				Build(), 1),
		).
			// ⭐ 不再需要显式设置 Width，flex 容器会自动填充父容器宽度
			Build(),
		ui.Text("│"),
		ui.Text("└────────────────────────────────────────────────────────────┘"),
	).
		Gap(0).
		Build()
}
