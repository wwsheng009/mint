// Package main is a simple test for Fiber reconciler
package main

import (
	"fmt"

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

// SimpleCounter is a counter component for Fiber testing
func SimpleCounter() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)

	// 将 setter 保存到 GlobalState，供 handler 从 ActionContext 读取
	ctx := ui.GetCurrentContext()
	if ctx != nil {
		ctx.GlobalState["setCount"] = setCount
	}

	// Register intent handlers using ui.On
	ui.On(DecrementIntent{}, func(actx *intent.ActionContext) {
		if fn, ok := actx.GetState("setCount"); ok {
			if setter, ok := fn.(func(func(int) int)); ok {
				setter(func(c int) int { return c - 1 })
			}
		}
	})
	ui.On(IncrementIntent{}, func(actx *intent.ActionContext) {
		if fn, ok := actx.GetState("setCount"); ok {
			if setter, ok := fn.(func(func(int) int)); ok {
				setter(func(c int) int { return c + 1 })
			}
		}
	})

	return ui.VStack(
		ui.NewTextBuilder("Fiber Reconciler Test").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
			FgColor("green").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.NewButtonBuilder("  -  ").
				OnPress(DecrementIntent{}).
				Build(),
			ui.Text("   "),
			ui.NewButtonBuilder("  +  ").
				OnPress(IncrementIntent{}).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("Fiber Mode: ENABLED").
			FgColor("yellow").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Tab/Arrows: focus | Enter/Space: click | q: quit").
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
