// 04_queue_stats/main.go
// 队列统计与监控示例
//
// 演示如何使用 MockSandbox 的队列统计功能
// 监控事件队列的长度、内存使用和淘汰情况。

package main

import (
	"fmt"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

// StatsApp 显示队列统计的应用
func StatsApp() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)
	events, setEvents, _ := ui.UseStateInt(0)
	memory, setMemory, _ := ui.UseStateInt(0)

	return ui.VStack(
		app.NewTextBuilder("╔══════════════════════════════╗").
			FgColor("cyan").
			Build(),
		app.NewTextBuilder("║    Queue Stats Demo           ║").
			FgColor("cyan").
			Build(),
		app.NewTextBuilder("╚══════════════════════════════╝").
			FgColor("cyan").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("Count: "),
			app.NewTextBuilder(fmt.Sprintf("%d", count)).
				FgColor("green").
				Bold(true).
				Build(),
		),
		ui.Text(""),
		ui.HStack(
			ui.Text("Events: "),
			app.NewTextBuilder(fmt.Sprintf("%d", events)).
				FgColor("yellow").
				Build(),
		),
		ui.Text(""),
		ui.HStack(
			ui.Text("Memory: "),
			app.NewTextBuilder(fmt.Sprintf("%d bytes", memory)).
				FgColor("magenta").
				Build(),
		),
		ui.Text(""),
		app.ButtonBuilder("  [ + ]  ").
			OnClick(func() {
				setCount(count + 1)
				setEvents(events + 1)
				setMemory(memory + 128)
			}).
			Build(),
		app.ButtonBuilder("  [ - ]  ").
			OnClick(func() {
				if count > 0 {
					setCount(count - 1)
				}
				if events > 0 {
					setEvents(events - 1)
				}
				if memory > 0 {
					setMemory(memory - 128)
				}
			}).
			Build(),
		ui.Text(""),
		app.NewTextBuilder("──────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		app.NewTextBuilder("This demo shows queue stats.").
			FgColor("bright-black").
			Build(),
		app.NewTextBuilder("Run tests to see monitoring.").
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
