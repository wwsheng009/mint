// 02_snapshot/main.go
// 快照系统示例
//
// 演示如何使用 SnapshotManager 保存和恢复应用状态。
// 支持三种快照级别：Minimal、Standard、Full。

package main

import (
	"fmt"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

// StatefulApp 有状态的应用，用于演示快照功能
func StatefulApp() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)
	text, setText := ui.UseStateString("")
	mode, setMode, _ := ui.UseStateInt(0)

	// 模式显示
	modeNames := []string{"正常", "编辑", "查看"}
	modeName := modeNames[mode%3]

	return ui.VStack(
		app.NewTextBuilder("╔══════════════════════════════╗").
			FgColor("cyan").
			Build(),
		app.NewTextBuilder("║     Snapshot Demo             ║").
			FgColor("cyan").
			Build(),
		app.NewTextBuilder("╚══════════════════════════════╝").
			FgColor("cyan").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("Mode: "),
			app.NewTextBuilder(modeName).
				FgColor("yellow").
				Bold(true).
				Build(),
		),
		ui.Text(""),
		app.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
			FgColor("green").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("Text: "),
			app.NewTextBuilder(text).
				FgColor("magenta").
				Build(),
		),
		ui.Text(""),
		app.ButtonBuilder("  [ + ]  ").
			OnClick(func() {
				setCount(count + 1)
			}).
			Build(),
		app.ButtonBuilder("  [ - ]  ").
			OnClick(func() {
				setCount(count - 1)
			}).
			Build(),
		app.ButtonBuilder("  [ Mode ]  ").
			OnClick(func() {
				setMode(mode + 1)
			}).
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("Input: "),
			app.InputBuilder().
				Value(text).
				Placeholder("Type something...").
				MaxLength(20).
				OnChange(setText).
				Build(),
		),
		ui.Text(""),
		app.NewTextBuilder("──────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		app.NewTextBuilder("This app demonstrates snapshots.").
			FgColor("bright-black").
			Build(),
		app.NewTextBuilder("Run tests to see save/restore.").
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
