// Package main is a simple test for Fiber reconciler
package main

import (
	"fmt"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

// SimpleCounter is a counter component for Fiber testing
func SimpleCounter() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)

	return ui.VStack(
		app.NewTextBuilder("Fiber Reconciler Test").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.Text(""),
		app.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
			FgColor("green").
			Build(),
		ui.Text(""),
		ui.HStack(
			app.ButtonBuilder("  -  ").
				OnClick(func() {
					setCount(func(c int) int { return c - 1 })
				}).
				Build(),
			ui.Text("   "),
			app.ButtonBuilder("  +  ").
				OnClick(func() {
					setCount(func(c int) int { return c + 1 })
				}).
				Build(),
		),
		ui.Text(""),
		app.NewTextBuilder("Fiber Mode: ENABLED").
			FgColor("yellow").
			Build(),
		ui.Text(""),
		app.NewTextBuilder("Tab/Arrows: focus | Enter/Space: click | q: quit").
			FgColor("bright-black").
			Build(),
	)
}

func main() {
	// Fiber is enabled via MINT_USE_FIBER environment variable
	// Run like: MINT_USE_FIBER=true go run examples/fiber/main.go
	err := ui.Run(SimpleCounter,
		ui.WithWidth(40),
		ui.WithHeight(14),
		ui.WithTitle("Fiber Test"),
	)
	if err != nil {
		panic(err)
	}
}
