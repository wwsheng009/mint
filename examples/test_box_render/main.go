package main

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
	ui "github.com/wwsheng009/mint/ui"
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
			ui.NewButtonBuilder("Left").
				PaddingH(1, 2).
				Flex(1).
				TextAlign(rtui.AlignStart).
				Build(),
			ui.NewButtonBuilder("Center").
				PaddingH(1, 1).
				Flex(1).
				TextAlign(rtui.AlignCenter).
				Build(),
			ui.NewButtonBuilder("Right").
				PaddingH(2, 1).
				Flex(1).
				TextAlign(rtui.AlignEnd).
				Build(),
		).
			Gap(1).
			Build(),
		ui.Text(""),
		ui.Text("2. Text with padding:"),
		ui.PaddingAll(ui.Text("Padded Text"), 2),
		ui.Text(""),
		ui.Text("3. Button with margins:"),
		ui.NewButtonBuilder("Btn1").
			MarginV(0, 1).
			Build(),
		ui.NewButtonBuilder("Btn2").
			MarginV(0, 1).
			Build(),
		ui.Text(""),
		ui.Text("4. Spacious button:"),
		ui.NewButtonBuilder("Spacious").
			PaddingAll(1).
			MarginV(0, 0).
			Flex(1).
			Build(),
	).
		Gap(0).
		Build()
}
