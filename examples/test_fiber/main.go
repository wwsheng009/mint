// Test Fiber mode
package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

// Intent Types
type DecrementIntent struct{}

func (DecrementIntent) IntentType() string { return "Decrement" }
func (DecrementIntent) StayPressed() bool  { return true }

type IncrementIntent struct{}

func (IncrementIntent) IntentType() string { return "Increment" }
func (IncrementIntent) StayPressed() bool  { return true }

func SimpleApp() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)

	// Save setter to context for handler access
	ctx := ui.GetCurrentContext()
	if ctx != nil {
		ctx.SetState("setCount", setCount)
	}

	// Register intent handlers
	ui.On(DecrementIntent{}, func(ctx *intent.ActionContext) {
		if fn, ok := ctx.GetState("setCount"); ok {
			if setter, ok := fn.(func(func(int) int)); ok {
				setter(func(c int) int { return c - 1 })
			}
		}
	})
	ui.On(IncrementIntent{}, func(ctx *intent.ActionContext) {
		if fn, ok := ctx.GetState("setCount"); ok {
			if setter, ok := fn.(func(func(int) int)); ok {
				setter(func(c int) int { return c + 1 })
			}
		}
	})

	return ui.VStack(
		ui.Text("Fiber Mode Test"),
		ui.Text(fmt.Sprintf("Count: %d", count)),
		ui.HStack(
			ui.NewButtonBuilder("[-]").
				OnPress(DecrementIntent{}).
				Build(),
			ui.NewButtonBuilder("[+]").
				OnPress(IncrementIntent{}).
				Build(),
		),
	)
}

func main() {
	// Enable Fiber mode explicitly
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("TUI_DEBUG_UI", "true")

	fmt.Println("Starting with Fiber mode...")
	fmt.Println("MINT_USE_FIBER =", os.Getenv("MINT_USE_FIBER"))

	err := ui.Run(SimpleApp,
		ui.WithWidth(40),
		ui.WithHeight(10),
		ui.WithTitle("Fiber Test"),
	)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
