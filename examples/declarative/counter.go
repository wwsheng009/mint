package main

import (
	"fmt"

	"github.com/wwsheng009/mint/ui"
)

// Counter is a dynamic counter component using useState
func Counter() ui.VNode {
	count, setCount, getCount := ui.UseStateInt(0)

	// Use getCount() in handlers to get the latest value
	decrement := func() {
		setCount(getCount() - 1)
	}
	increment := func() {
		setCount(getCount() + 1)
	}

	return ui.VStack(
		ui.NewTextBuilder("Mint UI Counter Demo").
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
				OnClick(decrement).
				Build(),
			ui.Text("   "),
			ui.ButtonBuilder("  +  ").
				OnClick(increment).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("Tab/Arrows: focus | Enter/Space: click | q: quit").
			FgColor("bright-black").
			Build(),
	)
}

func main() {
	err := ui.Run(Counter,
		ui.WithWidth(40),
		ui.WithHeight(12),
		ui.WithTitle("Counter Demo"),
	)
	if err != nil {
		panic(err)
	}
}
