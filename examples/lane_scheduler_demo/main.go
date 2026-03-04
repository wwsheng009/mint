// Package main demonstrates Lane Scheduler integration with ui.Run.
package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime/scheduler"
	"github.com/wwsheng009/mint/ui"
)

// Intent types
type IncrementIntent struct{}
type DecrementIntent struct{}
type BackgroundTaskIntent struct{}

func (IncrementIntent) IntentType() string { return "Increment" }
func (DecrementIntent) IntentType() string { return "Decrement" }
func (BackgroundTaskIntent) IntentType() string { return "BackgroundTask" }

func main() {
	fmt.Println("=== Lane Scheduler Demo ===")
	fmt.Println()
	fmt.Println("This demo shows how to use Lane Scheduler with ui.Run().")
	fmt.Println("User input (buttons) uses high priority (InputLane).")
	fmt.Println("Background tasks use low priority (IdleLane).")
	fmt.Println()

	// Register intent handlers
	ui.On(IncrementIntent{}, func(ctx *intent.ActionContext) {
		// High-priority user input
		if fn, ok := ctx.GetState("setCount"); ok {
			if setter, ok := fn.(func(func(int) int)); ok {
				setter(func(c int) int { return c + 1 })
			}
		}
	})

	ui.On(DecrementIntent{}, func(ctx *intent.ActionContext) {
		// High-priority user input
		if fn, ok := ctx.GetState("setCount"); ok {
			if setter, ok := fn.(func(func(int) int)); ok {
				setter(func(c int) int { return c - 1 })
			}
		}
	})

	ui.On(BackgroundTaskIntent{}, func(ctx *intent.ActionContext) {
		// Low-priority background task
		// This demonstrates scheduling work at different priorities
		if rtui.HasGlobalFiberScheduler() {
			rtui.ScheduleIdle(func() {
				fmt.Println("Background task executed at idle priority")
				// Perform non-critical work
			})
		} else {
			fmt.Println("Lane scheduler not enabled, executing immediately")
		}
	})

	// Run with Lane Scheduler enabled
	err := ui.Run(App,
		ui.WithLaneScheduler(),
		ui.WithWidth(60),
		ui.WithHeight(20),
		ui.WithTitle("Lane Scheduler Demo"),
	)
	if err != nil {
		fmt.Println("Error:", err)
	}
}

// App is the main application component
func App() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)

	// Store setter in GlobalState for handler access
	ctx := ui.GetCurrentContext()
	if ctx != nil {
		ctx.GlobalState["setCount"] = setCount
	}

	// Check if scheduler is enabled
	schedulerEnabled := rtui.HasGlobalFiberScheduler()

	return ui.VStack(
		ui.NewTextBuilder("Lane Scheduler Demo").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
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
			ui.Text("  "),
			ui.NewButtonBuilder(" BG ").
				OnPress(BackgroundTaskIntent{}).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("Scheduler: %v", schedulerEnabled)).
			FgColor("yellow").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Buttons: -/+ = InputLane | BG = IdleLane").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Press q to quit").
			FgColor("bright-black").
			Build(),
	)
}

// init registers the lane info
func init() {
	// Print lane priorities for demonstration
	fmt.Println("Lane Priorities (highest to lowest):")
	fmt.Printf("  SyncLane:       %s (priority: %d)\n", scheduler.SyncLane, scheduler.SyncLane.Priority())
	fmt.Printf("  InputLane:      %s (priority: %d)\n", scheduler.InputLane, scheduler.InputLane.Priority())
	fmt.Printf("  DefaultLane:    %s (priority: %d)\n", scheduler.DefaultLane, scheduler.DefaultLane.Priority())
	fmt.Printf("  TransitionLane: %s (priority: %d)\n", scheduler.TransitionLane, scheduler.TransitionLane.Priority())
	fmt.Printf("  IdleLane:       %s (priority: %d)\n", scheduler.IdleLane, scheduler.IdleLane.Priority())
	fmt.Println()
}
