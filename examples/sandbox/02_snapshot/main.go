// 02_snapshot/main.go
// 快照系统示例
//
// 演示如何使用 SnapshotManager 保存和恢复应用状态。
// 支持三种快照级别：Minimal、Standard、Full。

package main

import (
	"fmt"

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
		ui.ButtonBuilder("  [ + ]  ").
			OnClick(func() {
				setCount(count + 1)
			}).
			Build(),
		ui.ButtonBuilder("  [ - ]  ").
			OnClick(func() {
				setCount(count - 1)
			}).
			Build(),
		ui.ButtonBuilder("  [ Mode ]  ").
			OnClick(func() {
				setMode(mode + 1)
			}).
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("Input: "),
			ui.InputBuilder().
				Value(text).
				Placeholder("Type something...").
				MaxLength(20).
				OnChange(setText).
				Build(),
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
