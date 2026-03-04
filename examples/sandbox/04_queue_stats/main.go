// 04_queue_stats/main.go
// 队列统计与监控示例
//
// 演示如何使用 MockSandbox 的队列统计功能
// 监控事件队列的长度、内存使用和淘汰情况。

package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

// Intent Types
type IncrementStatsIntent struct{}
func (IncrementStatsIntent) IntentType() string { return "IncrementStats" }
func (IncrementStatsIntent) StayPressed() bool  { return true }

type DecrementStatsIntent struct{}
func (DecrementStatsIntent) IntentType() string { return "DecrementStats" }
func (DecrementStatsIntent) StayPressed() bool  { return true }

// StatsApp 显示队列统计的应用
func StatsApp() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)
	events, setEvents, _ := ui.UseStateInt(0)
	memory, setMemory, _ := ui.UseStateInt(0)

	// 将 setter 保存到 GlobalState，供 handler 从 ActionContext 读取
	ctx := ui.GetCurrentContext()
	if ctx != nil {
		ctx.GlobalState["count"] = count
		ctx.GlobalState["setCount"] = setCount
		ctx.GlobalState["setEvents"] = setEvents
		ctx.GlobalState["events"] = events
		ctx.GlobalState["setMemory"] = setMemory
		ctx.GlobalState["memory"] = memory
	}

	// Register intent handlers
	ui.On(IncrementStatsIntent{}, func(actx *intent.ActionContext) {
		currentCount := actx.GetIntState("count", 0)
		currentEvents := actx.GetIntState("events", 0)
		currentMemory := actx.GetIntState("memory", 0)
		
		if fn, ok := actx.GetState("setCount"); ok {
			if setter, ok := fn.(func(int)); ok {
				setter(currentCount + 1)
			}
		}
		if fn, ok := actx.GetState("setEvents"); ok {
			if setter, ok := fn.(func(int)); ok {
				setter(currentEvents + 1)
			}
		}
		if fn, ok := actx.GetState("setMemory"); ok {
			if setter, ok := fn.(func(int)); ok {
				setter(currentMemory + 128)
			}
		}
	})
	ui.On(DecrementStatsIntent{}, func(actx *intent.ActionContext) {
		currentCount := actx.GetIntState("count", 0)
		currentEvents := actx.GetIntState("events", 0)
		currentMemory := actx.GetIntState("memory", 0)
		
		if currentCount > 0 {
			if fn, ok := actx.GetState("setCount"); ok {
				if setter, ok := fn.(func(int)); ok {
					setter(currentCount - 1)
				}
			}
		}
		if currentEvents > 0 {
			if fn, ok := actx.GetState("setEvents"); ok {
				if setter, ok := fn.(func(int)); ok {
					setter(currentEvents - 1)
				}
			}
		}
		if currentMemory > 0 {
			if fn, ok := actx.GetState("setMemory"); ok {
				if setter, ok := fn.(func(int)); ok {
					setter(currentMemory - 128)
				}
			}
		}
	})

	return ui.VStack(
		ui.NewTextBuilder("╔══════════════════════════════╗").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("║    Queue Stats Demo           ║").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("╚══════════════════════════════╝").
			FgColor("cyan").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("Count: "),
			ui.NewTextBuilder(fmt.Sprintf("%d", count)).
				FgColor("green").
				Bold(true).
				Build(),
		),
		ui.Text(""),
		ui.HStack(
			ui.Text("Events: "),
			ui.NewTextBuilder(fmt.Sprintf("%d", events)).
				FgColor("yellow").
				Build(),
		),
		ui.Text(""),
		ui.HStack(
			ui.Text("Memory: "),
			ui.NewTextBuilder(fmt.Sprintf("%d bytes", memory)).
				FgColor("magenta").
				Build(),
		),
		ui.Text(""),
		ui.NewButtonBuilder("  [ + ]  ").
			OnPress(IncrementStatsIntent{}).
			Build(),
		ui.NewButtonBuilder("  [ - ]  ").
			OnPress(DecrementStatsIntent{}).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("──────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("This demo shows queue stats.").
			FgColor("bright-black").
			Build(),
		ui.NewTextBuilder("Run tests to see monitoring.").
			FgColor("bright-black").
			Build(),
	)
}

func main() {
	err := ui.Run(StatsApp,
		ui.WithWidth(40),
		ui.WithHeight(18),
		ui.WithTitle("Queue Stats Demo"),
	)
	if err != nil {
		panic(err)
	}
}
