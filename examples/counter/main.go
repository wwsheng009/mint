package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

// Counter is a dynamic counter component using ComponentContext and Intent
func Counter() ui.VNode {
	// Get current context (initialized by Fiber-first framework during render)
	ctx := ui.GetCurrentContext()

	// Read count from GlobalState (incremented by IncrementIntent)
	// Using default value of 0 for first render
	count := ctx.GetIntState("count", 0)

	// Check if running in Fiber mode
	isFiber := os.Getenv("MINT_USE_FIBER") == "true"
	fiberStr := "OFF"
	if isFiber {
		fiberStr = "ON"
	}

	log.TempLogger.Debug("[Counter] Render: count=%d, Fiber=%s", count, fiberStr)

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
		app.NewTextBuilder(fmt.Sprintf("[Fiber: %s] Using GlobalState + IncrementIntent", fiberStr)).
			FgColor("yellow").
			Build(),
		app.Text(""),
		app.HStack(
			// Decrement button using Intent (Fiber-first)
			// Intent created with fresh parameter values at render time
			app.ButtonBuilder("  -  ").
				OnPress(intent.Decrement("count", 1)).
				Build(),
			app.Text("   "),
			// Increment button using Intent (Fiber-first)
			// Intent created with fresh parameter values at render time
			app.ButtonBuilder("  +  ").
				OnPress(intent.Increment("count", 1)).
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
