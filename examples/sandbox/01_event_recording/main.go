// 01_event_recording/main.go
// 事件录制与回放示例
//
// 演示如何使用 EventRecorder 录制用户操作序列，
// 然后通过 ReplaySandbox 回放这些操作。

package main

import (
	"fmt"

	"github.com/wwsheng009/mint/ui"
)

// SimpleCounter 简单的计数器应用
func SimpleCounter() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)

	return ui.VStack(
		ui.NewTextBuilder("╔══════════════════════════════╗").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("║   Event Recording Demo       ║").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("╚══════════════════════════════╝").
			FgColor("cyan").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
			FgColor("green").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.ButtonBuilder("  [ - ]  ").
				OnClick(func() {
					setCount(func(c int) int { return c - 1 })
				}).
				Build(),
			ui.Text(" "),
			ui.ButtonBuilder("  [ + ]  ").
				OnClick(func() {
					setCount(func(c int) int { return c + 1 })
				}).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("──────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("This demo records your actions").
			FgColor("bright-black").
			Build(),
		ui.NewTextBuilder("and can replay them later.").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.NewTextBuilder("Tab: focus").
				FgColor("bright-black").
				Build(),
			ui.Text("  "),
			ui.NewTextBuilder("Enter: click").
				FgColor("bright-black").
				Build(),
			ui.Text("  "),
			ui.NewTextBuilder("q: quit").
				FgColor("bright-black").
				Build(),
		),
	)
}

func main() {
	err := ui.Run(SimpleCounter,
		ui.WithWidth(40),
		ui.WithHeight(16),
		ui.WithTitle("Event Recording Demo"),
	)
	if err != nil {
		panic(err)
	}
}
