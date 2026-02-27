package main

import (
	"fmt"

	"github.com/wwsheng009/mint/app"
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
			app.ButtonBuilder(label).
				FocusStyle(app.FocusStyleBracket).
				Build(),
		)
	}

	return ui.VStack(
		app.Text("Wrap Component Demo"),
		app.Text("─────────────────"),
		app.Text(""),
		app.Text("10 buttons with automatic wrapping:"),
		app.Text(""),
		app.WrapBuilder(buttons...).
			Gap(1).
			RowGap(0).
			Width(76). // 80 - border padding
			Align(ui.AlignStart).
			Build(),
		app.Text(""),
		app.Text("─ Usage ───────────────────────────────────"),
		app.Text("WrapBuilder(children...).Gap(1).ScreenWidth(76)"),
		app.Text(""),
	)
}
