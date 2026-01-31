package main

import (
	"fmt"

	"github.com/wwsheng009/mint/ui"
)

func RefDemo() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)

	return ui.VStack(
		ui.NewTextBuilder("State Demo").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
			FgColor("green").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.ButtonBuilder("  -  ").
				OnClick(func() {
					setCount(func(c int) int { return c - 1 })
				}).
				Build(),
			ui.Text("   "),
			ui.ButtonBuilder("  +  ").
				OnClick(func() {
					setCount(func(c int) int { return c + 1 })
				}).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("Tab: focus | Enter: click | q: quit").
			FgColor("bright-black").
			Build(),
	)
}

func main() {
	err := ui.Run(RefDemo,
		ui.WithWidth(40),
		ui.WithHeight(12),
		ui.WithTitle("State Demo"),
	)
	if err != nil {
		panic(err)
	}
}
