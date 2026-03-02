package main

import (
	"fmt"

	ui "github.com/wwsheng009/mint/ui"
)

func main() {
	err := ui.Run(WrapDemo,
		ui.WithWidth(80),
		ui.WithHeight(20),
		ui.WithTitle("Wrap Component Demo"),
	)
	if err != nil {
		panic(err)
	}
}

func WrapDemo() ui.VNode {
	// Create 10 buttons to demonstrate automatic wrapping
	var buttons []ui.VNode
	for i := 1; i <= 10; i++ {
		label := fmt.Sprintf("[%d]", i)
		buttons = append(buttons,
			ui.NewButtonBuilder(label).
				FocusStyle(ui.FocusStyleBracket).
				Build(),
		)
	}

	return ui.VStack(
		ui.Text("Wrap Component Demo"),
		ui.Text("─────────────────"),
		ui.Text(""),
		ui.Text("10 buttons with automatic wrapping:"),
		ui.Text(""),
		ui.NewWrapBuilder().Children(buttons...).
			Gap(1).
			RowGap(0).
			Width(76). // 80 - border padding
			Align(ui.AlignStart).
			Build(),
		ui.Text(""),
		ui.Text("─ Usage ───────────────────────────────────"),
		ui.Text("WrapBuilder(children...).Gap(1).ScreenWidth(76)"),
		ui.Text(""),
	)
}
