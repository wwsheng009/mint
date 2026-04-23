// 04_queue_stats/main.go
// 队列统计与监控示例 (Store 模式)
//
// 演示如何使用 MockSandbox 的队列统计功能
// 监控事件队列的长度、内存使用和淘汰情况。

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
	Count  int // 计数器值
	Events int // 事件数量
	Memory int // 内存使用（字节）
}

// ============================================================================
// Intent Types
// ============================================================================

type IncrementStatsIntent struct{}
func (IncrementStatsIntent) IntentType() string { return "IncrementStats" }
func (IncrementStatsIntent) StayPressed() bool  { return true }

type DecrementStatsIntent struct{}
func (DecrementStatsIntent) IntentType() string { return "DecrementStats" }
func (DecrementStatsIntent) StayPressed() bool  { return true }

// ============================================================================
// Store 初始化
// ============================================================================

var queueStatsStore = store.NewStore(AppState{
	Count:  0,
	Events: 0,
	Memory: 0,
})

// ============================================================================
// Reducer 注册
// ============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(IncrementStatsIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Count++
			s.Events++
			s.Memory += 128
			return s
		}).
		On(DecrementStatsIntent{}, func(s AppState, i intent.Intent) AppState {
			if s.Count > 0 {
				s.Count--
			}
			if s.Events > 0 {
				s.Events--
			}
			if s.Memory >= 128 {
				s.Memory -= 128
			} else {
				s.Memory = 0
			}
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), queueStatsStore)
}

// ============================================================================
// StatsApp - 显示队列统计的应用
// ============================================================================

func StatsApp() ui.VNode {
	// ✅ 订阅存储的状态
	count := ui.UseStoreSelector(queueStatsStore, func(s AppState) int { return s.Count })
	events := ui.UseStoreSelector(queueStatsStore, func(s AppState) int { return s.Events })
	memory := ui.UseStoreSelector(queueStatsStore, func(s AppState) int { return s.Memory })

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
		ui.NewButtonBuilder("  [ + ]  ").
			OnPress(IncrementStatsIntent{}).
			Build(),
		ui.NewButtonBuilder("  [ - ]  ").
			OnPress(DecrementStatsIntent{}).
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

// ============================================================================
// Main
// ============================================================================

func main() {
	err := ui.Run(StatsApp,
		ui.WithWidth(40),
		ui.WithHeight(18),
		ui.WithTitle("Queue Stats Demo (Store 模式)"),
	)
	if err != nil {
		panic(err)
	}
}
