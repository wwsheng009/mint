package main

import (
	ui "github.com/wwsheng009/mint/ui"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/app"
)

func main() {
	err := ui.Run(TestLayout,
		ui.WithWidth(80),
		ui.WithHeight(20),
		ui.WithTitle("Box Model Test"),
	)
	if err != nil {
		panic(err)
	}
}

func TestLayout() ui.VNode {
	return ui.VStackBuilder(
		ui.Text("Box Model Layout Test"),
		ui.Text("─────────────────────"),
		ui.Text(""),
		ui.Text("1. Three flex buttons:"),
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
		ui.Text(""),
		ui.Text("2. Text with padding:"),
		ui.PaddingAll(ui.Text("Padded Text"), 2),
		ui.Text(""),
		ui.Text("3. Button with margins:"),
		app.ButtonBuilder("Btn1").
			MarginV(0, 1).
			Build(),
		app.ButtonBuilder("Btn2").
			MarginV(0, 1).
			Build(),
		ui.Text(""),
		ui.Text("4. Spacious button:"),
		app.ButtonBuilder("Spacious").
			PaddingAll(1).
			MarginV(0, 0).
			Flex(1).
			Build(),
	).
		Gap(0).
		Build()
}
