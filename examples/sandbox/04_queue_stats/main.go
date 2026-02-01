// 04_queue_stats/main.go
// 队列统计与监控示例
//
// 演示如何使用 MockSandbox 的队列统计功能
// 监控事件队列的长度、内存使用和淘汰情况。

package main

import (
	"fmt"

	"github.com/wwsheng009/mint/ui"
)

// StatsApp 显示队列统计的应用
func StatsApp() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)
	events, setEvents, _ := ui.UseStateInt(0)
	memory, setMemory, _ := ui.UseStateInt(0)

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
		ui.ButtonBuilder("  [ + ]  ").
			OnClick(func() {
				setCount(count + 1)
				setEvents(events + 1)
				setMemory(memory + 128)
			}).
			Build(),
		ui.ButtonBuilder("  [ - ]  ").
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
