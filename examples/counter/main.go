package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

// Counter is a dynamic counter component using useState
func Counter() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)

	// Check if running in Fiber mode
	isFiber := os.Getenv("MINT_USE_FIBER") == "true"
	fiberStr := "OFF"
	if isFiber {
		fiberStr = "ON"
	}

	return ui.VStack(
		app.NewTextBuilder("Mint UI Counter Demo").
			FgColor("cyan").
			Bold(true).
			Build(),
		app.Text(""),
		app.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
			FgColor("green").
			Build(),
		app.Text(""),
		// Debug info line
		app.NewTextBuilder(fmt.Sprintf("[Fiber: %s] Press Tab to test focus navigation", fiberStr)).
			FgColor("yellow").
			Build(),
		app.Text(""),
		app.HStack(
			app.ButtonBuilder("  -  ").
				OnClick(func() {
					setCount(func(c int) int { return c - 1 })
				}).
				Build(),
			app.Text("   "),
			app.ButtonBuilder("  +  ").
				OnClick(func() {
					setCount(func(c int) int { return c + 1 })
				}).
				Build(),
		),
		app.Text(""),
		app.NewTextBuilder("Focused button = BLUE background").
			FgColor("bright-black").
			Build(),
		app.Text(""),
		app.NewTextBuilder("Tab/Arrows: focus | Enter/Space: click | q: quit").
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
