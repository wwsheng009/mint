package main

import (
	ui "github.com/wwsheng009/mint/ui"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/app"
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
		ui.Text("Expected above: Buttons evenly distributed, text aligned as per label"),
		ui.Text(""),
		ui.Text("Test 2: Button with padding and centering"),
		ui.Text("─────────────────────────────────────────────────────────────────────"),
		ui.Text(""),
		app.ButtonBuilder("Padded & Centered Button").
			PaddingAll(2).
			Flex(1).
			SetTextAlign(rtui.AlignCenter).
			Build(),
		ui.Text(""),
		ui.Text("Expected above: Button fills width, text centered"),
		ui.Text(""),
		ui.Text("Test 3: Buttons with vertical margin"),
		ui.Text("─────────────────────────────────────────────────────────────────────"),
		app.ButtonBuilder("Button 1").
			MarginV(0, 1).
			Build(),
		app.ButtonBuilder("Button 2").
			MarginV(0, 1).
			Build(),
		ui.Text(""),
		ui.Text("Expected above: Buttons stacked with gap of 1 row"),
		ui.Text(""),
	).
		Gap(0).
		Build()
}
