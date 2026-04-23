// 02_snapshot/main.go
// 快照系统示例 (Store 模式)
//
// 演示如何使用 SnapshotManager 保存和恢复应用状态。
// 支持三种快照级别：Minimal、Standard、Full。

package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// ============================================================================
// AppState - 定义应用状态
// ============================================================================

type AppState struct {
	Count int    // 计数器值
	Text  string // 文本内容
	Mode  int    // 模式索引
}

// ============================================================================
// Intent Types
// ============================================================================

type IncrementSnapshotIntent struct{}
func (IncrementSnapshotIntent) IntentType() string { return "Increment" }
func (IncrementSnapshotIntent) StayPressed() bool  { return true }

type DecrementSnapshotIntent struct{}
func (DecrementSnapshotIntent) IntentType() string { return "Decrement" }
func (DecrementSnapshotIntent) StayPressed() bool  { return true }

type ToggleModeIntent struct{}
func (ToggleModeIntent) IntentType() string { return "ToggleMode" }
func (ToggleModeIntent) StayPressed() bool  { return true }

type SetSnapshotTextIntent struct {
	Text string
}
func (SetSnapshotTextIntent) IntentType() string { return "SetSnapshotText" }
func (SetSnapshotTextIntent) StayPressed() bool  { return false }

// ============================================================================
// Store 初始化
// ============================================================================

var snapshotStore = store.NewStore(AppState{
	Count: 0,
	Text:  "",
	Mode:  0,
})

// ============================================================================
// Reducer 注册
// ============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(IncrementSnapshotIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Count++
			return s
		}).
		On(DecrementSnapshotIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Count--
			return s
		}).
		On(ToggleModeIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Mode = (s.Mode + 1) % 3
			return s
		}).
		On(SetSnapshotTextIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Text = i.(SetSnapshotTextIntent).Text
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), snapshotStore)
}

// ============================================================================
// StatefulApp - 有状态的应用
// ============================================================================

func StatefulApp() ui.VNode {
	// ✅ 订阅存储的状态
	count := ui.UseStoreSelector(snapshotStore, func(s AppState) int { return s.Count })
	text := ui.UseStoreSelector(snapshotStore, func(s AppState) string { return s.Text })
	mode := ui.UseStoreSelector(snapshotStore, func(s AppState) int { return s.Mode })

	// 模式显示
	modeNames := []string{"正常", "编辑", "查看"}
	modeName := modeNames[mode]

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

// ============================================================================
// Main
// ============================================================================

func main() {
	err := ui.Run(StatefulApp,
		ui.WithWidth(40),
		ui.WithHeight(20),
		ui.WithTitle("Snapshot Demo (Store 模式)"),
	)
	if err != nil {
		panic(err)
	}
}
