// Package main demonstrates Lane Scheduler integration with ui.Run (Store 模式).
package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime/scheduler"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// ============================================================================
// AppState - 定义应用状态
// ============================================================================

type AppState struct {
	Count int // 计数器值
}

// ============================================================================
// Intent Types
// ============================================================================

type IncrementIntent struct{}
type DecrementIntent struct{}
type BackgroundTaskIntent struct{}

func (IncrementIntent) IntentType() string      { return "Increment" }
func (DecrementIntent) IntentType() string      { return "Decrement" }
func (BackgroundTaskIntent) IntentType() string { return "BackgroundTask" }

// ============================================================================
// Store 初始化
// ============================================================================

var laneSchedulerStore = store.NewStore(AppState{
	Count: 0,
})

// ============================================================================
// Reducer 注册
// ============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Count++
			return s
		}).
		On(DecrementIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Count--
			return s
		}).
		On(BackgroundTaskIntent{}, func(s AppState, i intent.Intent) AppState {
			// Background task intent - just a trigger, doesn't modify state
			fmt.Println("Background task scheduled at IdleLane...")
			if rtui.HasGlobalFiberScheduler() {
				rtui.ScheduleIdle(func() {
					fmt.Println("Background task executed at idle priority")
				})
				// Flush to execute scheduled idle work
				rtui.FlushScheduler()
			} else {
				fmt.Println("Lane scheduler not enabled, executing immediately")
			}
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), laneSchedulerStore)

	// Print lane priorities for demonstration
	fmt.Println("Lane Priorities (highest to lowest):")
	fmt.Printf("  SyncLane:       %s (priority: %d)\n", scheduler.SyncLane, scheduler.SyncLane.Priority())
	fmt.Printf("  InputLane:      %s (priority: %d)\n", scheduler.InputLane, scheduler.InputLane.Priority())
	fmt.Printf("  DefaultLane:    %s (priority: %d)\n", scheduler.DefaultLane, scheduler.DefaultLane.Priority())
	fmt.Printf("  TransitionLane: %s (priority: %d)\n", scheduler.TransitionLane, scheduler.TransitionLane.Priority())
	fmt.Printf("  IdleLane:       %s (priority: %d)\n", scheduler.IdleLane, scheduler.IdleLane.Priority())
	fmt.Println()

	fmt.Println("=== Lane Scheduler Demo ===")
	fmt.Println()
	fmt.Println("This demo shows how to use Lane Scheduler with ui.Run().")
	fmt.Println("User input (buttons) uses high priority (InputLane).")
	fmt.Println("Background tasks use low priority (IdleLane).")
	fmt.Println()
}

// ============================================================================
// Main
// ============================================================================

func main() {
	// Run with Lane Scheduler enabled
	err := ui.Run(App,
		ui.WithLaneScheduler(),
		ui.WithWidth(60),
		ui.WithHeight(20),
		ui.WithTitle("Lane Scheduler Demo (Store 模式)"),
	)
	if err != nil {
		fmt.Println("Error:", err)
	}
}

// ============================================================================
// App - 主应用组件
// ============================================================================

func App() ui.VNode {
	// ✅ 订阅存储的状态
	count := ui.UseStoreSelector(laneSchedulerStore, func(s AppState) int { return s.Count })

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
