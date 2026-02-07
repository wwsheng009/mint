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
			app.ButtonBuilder("Left").
				PaddingH(1, 2).
				Flex(1).
				SetTextAlign(rtui.AlignStart).
				Build(),
			app.ButtonBuilder("Center").
				PaddingH(1, 1).
				Flex(1).
				SetTextAlign(rtui.AlignCenter).
				Build(),
			app.ButtonBuilder("Right").
				PaddingH(2, 1).
				Flex(1).
				SetTextAlign(rtui.AlignEnd).
				Build(),
		).
			Gap(1).
			Build(),
		ui.Text("│"),
		ui.Text("│ Expected: [  Left  ]    [ Center ]    [  Right  ]"),
		ui.Text("│          (more pad)    (even pad)    (more pad)"),
		ui.Text("│"),
		ui.Text("│ Example 2: Button with padding and centering"),
		ui.Text("│"),
		app.ButtonBuilder("Padded & Centered").
			PaddingAll(2).
			Flex(1).
			SetTextAlign(rtui.AlignCenter).
			Build(),
		ui.Text("│"),
		ui.Text("└────────────────────────────────────────────────────────────┘"),
	).
		Gap(0).
		Build()
}
