package main

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
	ui "github.com/wwsheng009/mint/ui"
)

func main() {
	err := ui.Run(Demo,
		ui.WithWidth(80),
		ui.WithHeight(25),
		ui.WithTitle("Box Model - Final Test"),
	)
	if err != nil {
		panic(err)
	}
}

func Demo() ui.VNode {
	return ui.VStackBuilder(
		ui.Text("╔══════════════════════════════════════════════════════════════════════╗"),
		ui.Text("║                     Box Model Layout Test                            ║"),
		ui.Text("╚══════════════════════════════════════════════════════════════════════╝"),
		ui.Text(""),
		ui.Text("Test 1: Three flex buttons with different alignments"),
		ui.Text("─────────────────────────────────────────────────────────────────────"),
		ui.Text(""),
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
		ui.Text("Expected above: Buttons evenly distributed, text aligned as per label"),
		ui.Text(""),
		ui.Text("Test 2: Button with padding and centering"),
		ui.Text("─────────────────────────────────────────────────────────────────────"),
		ui.Text(""),
		ui.NewButtonBuilder("Padded & Centered Button").
			PaddingAll(2).
			Flex(1).
			TextAlign(rtui.AlignCenter).
			Build(),
		ui.Text(""),
		ui.Text("Expected above: Button fills width, text centered"),
		ui.Text(""),
		ui.Text("Test 3: Buttons with vertical margin"),
		ui.Text("─────────────────────────────────────────────────────────────────────"),
		ui.NewButtonBuilder("Button 1").
			MarginV(0, 1).
			Build(),
		ui.NewButtonBuilder("Button 2").
			MarginV(0, 1).
			Build(),
		ui.Text(""),
		ui.Text("Expected above: Buttons stacked with gap of 1 row"),
		ui.Text(""),
	).
		Gap(0).
		Build()
}
