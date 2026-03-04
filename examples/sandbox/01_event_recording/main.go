// 01_event_recording/main.go
// 事件录制与回放示例
//
// 演示如何使用 EventRecorder 录制用户操作序列，
// 然后通过 ReplaySandbox 回放这些操作。

package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
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

	// 将 setter 保存到 GlobalState，供 handler 从 ActionContext 读取
	ctx := ui.GetCurrentContext()
	if ctx != nil {
		ctx.GlobalState["setCount"] = setCount
	}

	// Register intent handlers
	ui.On(DecrementEventRecordingIntent{}, func(actx *intent.ActionContext) {
		if fn, ok := actx.GetState("setCount"); ok {
			if setter, ok := fn.(func(func(int) int)); ok {
				setter(func(c int) int { return c - 1 })
			}
		}
	})
	ui.On(IncrementEventRecordingIntent{}, func(actx *intent.ActionContext) {
		if fn, ok := actx.GetState("setCount"); ok {
			if setter, ok := fn.(func(func(int) int)); ok {
				setter(func(c int) int { return c + 1 })
			}
		}
	})

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
			ui.NewButtonBuilder("  [ - ]  ").
				OnPress(DecrementEventRecordingIntent{}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("  [ + ]  ").
				OnPress(IncrementEventRecordingIntent{}).
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
