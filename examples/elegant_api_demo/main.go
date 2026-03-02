package main

import (
	rtui "github.com/wwsheng009/mint/runtime/ui"
	ui "github.com/wwsheng009/mint/ui"
)

func main() {
	err := ui.Run(ElegantAPIDemo,
		ui.WithWidth(80),
		ui.WithHeight(25),
		ui.WithTitle("Elegant API Demo"),
	)
	if err != nil {
		panic(err)
	}
}

func ElegantAPIDemo() ui.VNode {
	return ui.VStackBuilder(
		ui.Text("✨ Elegant VNode Builder API Demo"),
		ui.Text("────────────────────────────────"),

		// Example 1: Flex buttons with no SetProp
		ui.Text("1. Flex buttons (no SetProp):"),
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

		// Example 2: Padding works on Text
		ui.Text("2. Text with PaddingAll(2):"),
		ui.PaddingAll(ui.Text("Padded Text"), 2),

		ui.Text(""),

		// Example 3: MarginV (垂直 margin - top/bottom)
		ui.Text("3. MarginV(top, bottom):"),
		ui.Text("   Btn1: MarginV(1, 0) → top=1, bottom=0"),
		ui.NewButtonBuilder("Btn1").
			MarginV(1, 0).
			Build(),
		ui.Text("   Btn2: MarginV(0, 1) → top=0, bottom=1"),
		ui.NewButtonBuilder("Btn2").
			MarginV(0, 1).
			Build(),
		ui.Text("   Btn3: MarginV(1, 1) → top=1, bottom=1"),
		ui.NewButtonBuilder("Btn3").
			MarginV(1, 1).
			Build(),

		ui.Text(""),

		// Example 4: MarginH (水平 margin - left/right)
		ui.Text("4. MarginH(left, right) in HStack:"),
		ui.HStackBuilder(
			ui.NewButtonBuilder("Left").
				MarginH(1, 0).  // left=1, right=0
				Build(),
			ui.NewButtonBuilder("Center").
				MarginH(0, 0).  // no margin
				Build(),
			ui.NewButtonBuilder("Right").
				MarginH(0, 1).  // left=0, right=1
				Build(),
		).
			Gap(1).
			Build(),

		ui.Text(""),

		// Example 5: Margin (四个方向)
		ui.Text("5. Margin(top, right, bottom, left):"),
		ui.HStackBuilder(
			ui.NewButtonBuilder("TL").
				Margin(1, 0, 0, 0).  // top=1
				Build(),
			ui.NewButtonBuilder("TR").
				Margin(0, 1, 0, 0).  // right=1
				Build(),
			ui.NewButtonBuilder("BL").
				Margin(0, 0, 1, 0).  // bottom=1
				Build(),
			ui.NewButtonBuilder("BR").
				Margin(0, 0, 0, 1).  // left=1
				Build(),
		).
			Gap(1).
			Build(),

		ui.Text(""),

		// Example 6: MarginAll (所有方向相同)
		ui.Text("6. MarginAll(value) - same on all sides:"),
		ui.NewButtonBuilder("Margin1").
			MarginAll(1).
			Build(),
		ui.NewButtonBuilder("Margin2").
			MarginAll(2).
			Build(),

		ui.Text(""),

		// Example 7: Combined padding + margin + flex
		ui.Text("7. Combined: Padding(1) + MarginV(0,0) + Flex(1):"),
		ui.NewButtonBuilder("Spacious").
			PaddingAll(1).
			MarginV(0, 0).
			Flex(1).
			Build(),

		ui.Text(""),

		// Example 8: CSS-like API
		ui.Text("8. CSS-like comparison:"),
		ui.Text("   CSS:     { padding: 2px; margin: 1px; flex: 1; }"),
		ui.Text("   Mint:    .PaddingAll(2).MarginAll(1).Flex(1).Build()"),
		ui.NewButtonBuilder("Click Me").
			PaddingAll(2).
			MarginAll(1).
			Flex(1).
			Build(),
	).
		Gap(0).
		Build()
}
