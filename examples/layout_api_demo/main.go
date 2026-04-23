package main

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
	ui "github.com/wwsheng009/mint/ui"
)

func main() {
	err := ui.Run(LayoutAPIDemo,
		ui.WithWidth(80),
		ui.WithHeight(40),
		ui.WithTitle("Layout API Demo"),
	)
	if err != nil {
		panic(err)
	}
}

// LayoutAPIDemo demonstrates padding, margin, flex, and gap layout APIs.
func LayoutAPIDemo() ui.VNode {
	return ui.VStackBuilder(
		ui.Text("╔══════════════════════════════════════════════════════════════════════╗"),
		ui.Text("║                       Layout API Demo                               ║"),
		ui.Text("╚══════════════════════════════════════════════════════════════════════╝"),
		ui.Text(""),

		// ── Section 1: Padding ──────────────────────────────────────────────────
		ui.Text("── 1. Padding ──────────────────────────────────────────────────────────"),
		ui.Text(""),
		ui.Text("PaddingAll(2):"),
		ui.NewButtonBuilder("Btn").
			PaddingAll(2).
			Build(),
		ui.Text(""),
		ui.Text("PaddingH(3, 3):"),
		ui.NewButtonBuilder("Btn").
			PaddingH(3, 3).
			Build(),
		ui.Text(""),
		ui.Text("PaddingAll(1) on Text:"),
		ui.PaddingAll(ui.Text("Text with padding"), 1),
		ui.Text(""),
		ui.Text("PaddingLeft(2) / PaddingH(1,1) / PaddingRight(2):"),
		ui.HStackBuilder(
			ui.NewButtonBuilder("Left").
				PaddingH(0, 2).
				Build(),
			ui.NewButtonBuilder("Center").
				PaddingH(1, 1).
				Build(),
			ui.NewButtonBuilder("Right").
				PaddingH(2, 0).
				Build(),
		).
			Gap(1).
			Build(),
		ui.Text(""),

		// ── Section 2: Margin ───────────────────────────────────────────────────
		ui.Text("── 2. Margin ───────────────────────────────────────────────────────────"),
		ui.Text(""),
		ui.Text("MarginV(top, bottom):"),
		ui.NewButtonBuilder("Btn1 MarginV(1,0)").
			MarginV(1, 0).
			Build(),
		ui.NewButtonBuilder("Btn2 MarginV(0,1)").
			MarginV(0, 1).
			Build(),
		ui.Text(""),
		ui.Text("MarginH(left, right) in HStack:"),
		ui.HStackBuilder(
			ui.NewButtonBuilder("Left").
				MarginH(1, 0).
				Build(),
			ui.NewButtonBuilder("Center").
				Build(),
			ui.NewButtonBuilder("Right").
				MarginH(0, 1).
				Build(),
		).
			Gap(1).
			Build(),
		ui.Text(""),
		ui.Text("Margin(top, right, bottom, left):"),
		ui.HStackBuilder(
			ui.NewButtonBuilder("top=1").
				Margin(1, 0, 0, 0).
				Build(),
			ui.NewButtonBuilder("right=1").
				Margin(0, 1, 0, 0).
				Build(),
			ui.NewButtonBuilder("bottom=1").
				Margin(0, 0, 1, 0).
				Build(),
			ui.NewButtonBuilder("left=1").
				Margin(0, 0, 0, 1).
				Build(),
		).
			Gap(1).
			Build(),
		ui.Text(""),
		ui.Text("MarginAll(1) and MarginAll(2):"),
		ui.HStackBuilder(
			ui.NewButtonBuilder("MarginAll(1)").
				MarginAll(1).
				Build(),
			ui.NewButtonBuilder("MarginAll(2)").
				MarginAll(2).
				Build(),
		).
			Build(),
		ui.Text(""),

		// ── Section 3: Flex + TextAlign ─────────────────────────────────────────
		ui.Text("── 3. Flex + TextAlign ─────────────────────────────────────────────────"),
		ui.Text(""),
		ui.Text("Flex(1) with AlignStart / AlignCenter / AlignEnd:"),
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
		ui.Text("Flex(1) full-width centered button:"),
		ui.NewButtonBuilder("Padded & Centered Button").
			PaddingAll(2).
			Flex(1).
			TextAlign(rtui.AlignCenter).
			Build(),
		ui.Text(""),

		// ── Section 4: Gap ──────────────────────────────────────────────────────
		ui.Text("── 4. Gap ──────────────────────────────────────────────────────────────"),
		ui.Text(""),
		ui.Text("VStack Gap(1) between buttons:"),
		ui.VStackBuilder(
			ui.NewButtonBuilder("Btn1").Build(),
			ui.NewButtonBuilder("Btn2").Build(),
			ui.NewButtonBuilder("Btn3").Build(),
		).
			Gap(1).
			Build(),
		ui.Text(""),
		ui.Text("MarginV(0,1) on each button (same visual result):"),
		ui.NewButtonBuilder("Button 1").
			MarginV(0, 1).
			Build(),
		ui.NewButtonBuilder("Button 2").
			MarginV(0, 1).
			Build(),
		ui.Text(""),
	).
		Gap(0).
		Build()
}
