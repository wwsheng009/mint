// 01_event_recording/main.go
// 事件录制与回放示例
//
// 演示如何使用 EventRecorder 录制用户操作序列，
// 然后通过 ReplaySandbox 回放这些操作。

package main

import (
	"fmt"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

// Intent Types
type DecrementEventRecordingIntent struct{}
func (DecrementEventRecordingIntent) IntentType() string { return "Decrement" }
func (DecrementEventRecordingIntent) StayPressed() bool  { return true }

type IncrementEventRecordingIntent struct{}
func (IncrementEventRecordingIntent) IntentType() string { return "Increment" }
func (IncrementEventRecordingIntent) StayPressed() bool  { return true }

// SimpleCounter 简单的计数器应用
func SimpleCounter() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)

	// Register intent handlers
	ui.On(DecrementEventRecordingIntent{}, func() {
		setCount(func(c int) int { return c - 1 })
	})
	ui.On(IncrementEventRecordingIntent{}, func() {
		setCount(func(c int) int { return c + 1 })
	})

	return ui.VStack(
		app.NewTextBuilder("╔══════════════════════════════╗").
			FgColor("cyan").
			Build(),
		app.NewTextBuilder("║   Event Recording Demo       ║").
			FgColor("cyan").
			Build(),
		app.NewTextBuilder("╚══════════════════════════════╝").
			FgColor("cyan").
			Build(),
		ui.Text(""),
		app.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
			FgColor("green").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.HStack(
			app.ButtonBuilder("  [ - ]  ").
				OnPress(DecrementEventRecordingIntent{}).
				Build(),
			ui.Text(" "),
			app.ButtonBuilder("  [ + ]  ").
				OnPress(IncrementEventRecordingIntent{}).
				Build(),
		),
		ui.Text(""),
		app.NewTextBuilder("──────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		app.NewTextBuilder("This demo records your actions").
			FgColor("bright-black").
			Build(),
		app.NewTextBuilder("and can replay them later.").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.HStack(
			app.NewTextBuilder("Tab: focus").
				FgColor("bright-black").
				Build(),
			ui.Text("  "),
			app.NewTextBuilder("Enter: click").
				FgColor("bright-black").
				Build(),
			ui.Text("  "),
			app.NewTextBuilder("q: quit").
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
