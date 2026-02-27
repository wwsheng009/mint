// Test Fiber mode
package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/app"
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

	// Register intent handlers
	ui.On(DecrementIntent{}, func() {
		setCount(func(c int) int { return c - 1 })
	})
	ui.On(IncrementIntent{}, func() {
		setCount(func(c int) int { return c + 1 })
	})

	return ui.VStack(
		ui.Text("Fiber Mode Test"),
		ui.Text(fmt.Sprintf("Count: %d", count)),
		app.HStack(
			app.ButtonBuilder("[-]").
				OnPress(DecrementIntent{}).
				Build(),
			app.ButtonBuilder("[+]").
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
