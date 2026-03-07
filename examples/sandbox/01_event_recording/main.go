// 01_event_recording/main.go
// 事件录制与回放示例 (Store 模式)
//
// 演示如何使用 EventRecorder 录制用户操作序列，
// 然后通过 ReplaySandbox 回放这些操作。

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
	Count int // 计数器值
}

// ============================================================================
// Intent Types
// ============================================================================

type DecrementEventRecordingIntent struct{}
func (DecrementEventRecordingIntent) IntentType() string { return "Decrement" }
func (DecrementEventRecordingIntent) StayPressed() bool  { return true }

type IncrementEventRecordingIntent struct{}
func (IncrementEventRecordingIntent) IntentType() string { return "Increment" }
func (IncrementEventRecordingIntent) StayPressed() bool  { return true }

// ============================================================================
// Store 初始化
// ============================================================================

var eventRecordingStore = store.NewStore(AppState{
	Count: 0,
})

// ============================================================================
// Reducer 注册
// ============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(DecrementEventRecordingIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Count--
			if s.Count < 0 {
				s.Count = 0
			}
			return s
		}).
		On(IncrementEventRecordingIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Count++
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), eventRecordingStore)
}

// ============================================================================
// SimpleCounter - 简单的计数器应用
// ============================================================================

func SimpleCounter() ui.VNode {
	// ✅ 订阅存储的状态
	count := ui.UseStoreSelector(eventRecordingStore, func(s AppState) int { return s.Count })

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

// ============================================================================
// Main
// ============================================================================

func main() {
	err := ui.Run(SimpleCounter,
		ui.WithWidth(40),
		ui.WithHeight(16),
		ui.WithTitle("Event Recording Demo (Store 模式)"),
	)
	if err != nil {
		panic(err)
	}
}
