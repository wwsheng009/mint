// 02_snapshot/main.go
// 快照系统示例
//
// 演示如何使用 SnapshotManager 保存和恢复应用状态。
// 支持三种快照级别：Minimal、Standard、Full。

package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

// Intent Types
type IncrementSnapshotIntent struct{}
func (IncrementSnapshotIntent) IntentType() string { return "Increment" }
func (IncrementSnapshotIntent) StayPressed() bool  { return true }

type DecrementSnapshotIntent struct{}
func (DecrementSnapshotIntent) IntentType() string { return "Decrement" }
func (DecrementSnapshotIntent) StayPressed() bool  { return true }

type ToggleModeIntent struct{}
func (ToggleModeIntent) IntentType() string { return "ToggleMode" }
func (ToggleModeIntent) StayPressed() bool  { return true }

// StatefulApp 有状态的应用，用于演示快照功能
func StatefulApp() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)
	text, _ := ui.UseStateString("")
	mode, setMode, _ := ui.UseStateInt(0)

	// 将 setter 保存到 GlobalState，供 handler 从 ActionContext 读取
	ctx := ui.GetCurrentContext()
	if ctx != nil {
		ctx.GlobalState["count"] = count
		ctx.GlobalState["setCount"] = setCount
		ctx.GlobalState["mode"] = mode
		ctx.GlobalState["setMode"] = setMode
	}

	// Register intent handlers
	ui.On(IncrementSnapshotIntent{}, func(actx *intent.ActionContext) {
		currentCount := actx.GetIntState("count", 0)
		if fn, ok := actx.GetState("setCount"); ok {
			if setter, ok := fn.(func(int)); ok {
				setter(currentCount + 1)
			}
		}
	})
	ui.On(DecrementSnapshotIntent{}, func(actx *intent.ActionContext) {
		currentCount := actx.GetIntState("count", 0)
		if fn, ok := actx.GetState("setCount"); ok {
			if setter, ok := fn.(func(int)); ok {
				setter(currentCount - 1)
			}
		}
	})
	ui.On(ToggleModeIntent{}, func(actx *intent.ActionContext) {
		currentMode := actx.GetIntState("mode", 0)
		if fn, ok := actx.GetState("setMode"); ok {
			if setter, ok := fn.(func(int)); ok {
				setter(currentMode + 1)
			}
		}
	})

	// 模式显示
	modeNames := []string{"正常", "编辑", "查看"}
	modeName := modeNames[mode%3]

	return ui.VStack(
		ui.NewTextBuilder("╔══════════════════════════════╗").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("║     Snapshot Demo             ║").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("╚══════════════════════════════╝").
			FgColor("cyan").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("Mode: "),
			ui.NewTextBuilder(modeName).
				FgColor("yellow").
				Bold(true).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
			FgColor("green").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("Text: "),
			ui.NewTextBuilder(text).
				FgColor("magenta").
				Build(),
		),
		ui.Text(""),
		ui.NewButtonBuilder("  [ + ]  ").
			OnPress(IncrementSnapshotIntent{}).
			Build(),
		ui.NewButtonBuilder("  [ - ]  ").
			OnPress(DecrementSnapshotIntent{}).
			Build(),
		ui.NewButtonBuilder("  [ Mode ]  ").
			OnPress(ToggleModeIntent{}).
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("Input: "),
			ui.NewInputBuilder().
				Value(text).
				Placeholder("Type something...").
				Build(), // TODO: integrate with FieldChangeIntent
		),
		ui.Text(""),
		ui.NewTextBuilder("──────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("This app demonstrates snapshots.").
			FgColor("bright-black").
			Build(),
		ui.NewTextBuilder("Run tests to see save/restore.").
			FgColor("bright-black").
			Build(),
	)
}

func main() {
	err := ui.Run(StatefulApp,
		ui.WithWidth(40),
		ui.WithHeight(20),
		ui.WithTitle("Snapshot Demo"),
	)
	if err != nil {
		panic(err)
	}
}
