// Package main demonstrates Fiber architecture basics.
package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	fmt.Println("=== Fiber Architecture Demo ===")
	fmt.Println()
	fmt.Println("This demo shows the core concepts of Fiber architecture:")
	fmt.Println("1. FiberNode structure")
	fmt.Println("2. Virtual DOM diffing")
	fmt.Println("3. Reconciliation process")
	fmt.Println()
	fmt.Println("Key concepts:")
	fmt.Println("- Each component has a corresponding FiberNode")
	fmt.Println("- FiberNode stores hooks, state, and VNode")
	fmt.Println("- Reconciliation updates the tree efficiently")
	fmt.Println()
	fmt.Println("See the following files for implementation:")
	fmt.Println("  - runtime/ui/fiber.go: FiberNode definition")
	fmt.Println("  - runtime/ui/fiber_vnode.go: VNode diffing")
	fmt.Println("  - runtime/ui/hooks.go: Hooks implementation")
	fmt.Println()

	// Run the demo app
	runDemoApp()
}

func runDemoApp() {
	fmt.Println("--- Running Demo App ---")
	fmt.Println()

	err := ui.Run(DemoComponent,
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Fiber Demo"),
	)
	if err != nil {
		fmt.Println("Error:", err)
	}
}

// DemoComponent demonstrates basic Fiber rendering
func DemoComponent() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)

	// Store setter in GlobalState for handler access
	ctx := ui.GetCurrentContext()
	if ctx != nil {
		ctx.GlobalState["setCount"] = setCount
	}

	return ui.VStack(
		ui.NewTextBuilder("Fiber Architecture Demo").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Count: "+fmt.Sprintf("%d", count)).
			FgColor("green").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.NewButtonBuilder(" - ").
				OnPress(DecrementIntent{}).
				Build(),
			ui.Text("  "),
			ui.NewButtonBuilder(" + ").
				OnPress(IncrementIntent{}).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("Fiber enables:").
			FgColor("yellow").
			Build(),
		ui.NewTextBuilder("  • Interruptible rendering").
			FgColor("bright-black").
			Build(),
		ui.NewTextBuilder("  • Priority scheduling").
			FgColor("bright-black").
			Build(),
		ui.NewTextBuilder("  • Concurrent updates").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Press q to quit").
			FgColor("bright-black").
			Build(),
	)
}

// Intent types
type DecrementIntent struct{}

func (DecrementIntent) IntentType() string { return "Decrement" }
func (DecrementIntent) StayPressed() bool   { return true }

type IncrementIntent struct{}

func (IncrementIntent) IntentType() string { return "Increment" }
func (IncrementIntent) StayPressed() bool   { return true }

func init() {
	// Register intent handlers using the new On API
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
}
