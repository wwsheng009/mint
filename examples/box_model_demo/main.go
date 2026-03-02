package main

import (
	ui "github.com/wwsheng009/mint/ui"
)

func main() {
	err := ui.Run(BoxModelDemo,
		ui.WithWidth(80),
		ui.WithHeight(20),
		ui.WithTitle("Box Model Demo"),
	)
	if err != nil {
		panic(err)
	}
}

func BoxModelDemo() ui.VNode {
	return ui.VStackBuilder(
		ui.Text("Universal Box Model Demo"),
		ui.Text("─────────────────────────"),

		// Example 1: Button with padding
		ui.Text("1. Button with PaddingAll(2):"),
		ui.NewButtonBuilder("Btn").
			PaddingAll(2).  // Universal method - works on any component
			Build(),

		ui.Text(""),

		// Example 2: Button with different horizontal padding
		ui.Text("2. Button with PaddingH(3, 3):"),
		ui.NewButtonBuilder("Btn").
			PaddingH(3, 3).
			Build(),

		ui.Text(""),

		// Example 3: Multiple buttons using VStack with Gap instead of margin
		ui.Text("3. Buttons with VStack Gap(1):"),
		ui.VStackBuilder(
			ui.NewButtonBuilder("Btn1").Build(),
			ui.NewButtonBuilder("Btn2").Build(),
			ui.NewButtonBuilder("Btn3").Build(),
		).
			Gap(1).
			Build(),

		ui.Text(""),

		// Example 4: Text with padding
		ui.Text("4. Text with PaddingAll(1):"),
		ui.PaddingAll(ui.Text("Text with padding"), 1),

		ui.Text(""),

		// Example 5: Buttons with different padding
		ui.Text("5. Buttons with PaddingLeft(2) vs PaddingRight(2):"),
		ui.HStackBuilder(
			ui.NewButtonBuilder("Left").
				PaddingH(0, 2).  // Right padding
				Build(),
			ui.NewButtonBuilder("Center").
				PaddingH(1, 1).  // Even padding
				Build(),
			ui.NewButtonBuilder("Right").
				PaddingH(2, 0).  // Left padding
				Build(),
		).
			Gap(1).
			Build(),
	).
		Gap(0).
	Build()
}
