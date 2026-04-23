// 05_injection_strategy/main.go
// 事件注入策略示例 (Store 模式)
//
// 演示不同的事件注入策略：
// - InjectProhibited: 禁止注入（生产环境）
// - InjectAllowed: 允许注入（测试环境）
// - InjectRecorded: 仅录制注入

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
	Count    int    // 计数器值
	Strategy string // 注入策略
}

// ============================================================================
// Intent Types
// ============================================================================

type IncrementStrategyIntent struct{}
func (IncrementStrategyIntent) IntentType() string { return "Increment" }
func (IncrementStrategyIntent) StayPressed() bool  { return true }

type DecrementStrategyIntent struct{}
func (DecrementStrategyIntent) IntentType() string { return "Decrement" }
func (DecrementStrategyIntent) StayPressed() bool  { return true }

// ============================================================================
// Store 初始化
// ============================================================================

var injectionStrategyStore = store.NewStore(AppState{
	Count:    0,
	Strategy: "Allowed",
})

// ============================================================================
// Reducer 注册
// ============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(IncrementStrategyIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Count++
			return s
		}).
		On(DecrementStrategyIntent{}, func(s AppState, i intent.Intent) AppState {
			if s.Count > 0 {
				s.Count--
			}
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), injectionStrategyStore)
}

// ============================================================================
// StrategyApp - 演示注入策略的应用
// ============================================================================

func StrategyApp() ui.VNode {
	// ✅ 订阅存储的状态
	count := ui.UseStoreSelector(injectionStrategyStore, func(s AppState) int { return s.Count })
	strategy := ui.UseStoreSelector(injectionStrategyStore, func(s AppState) string { return s.Strategy })

	return ui.VStack(
		ui.NewTextBuilder("╔══════════════════════════════╗").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("║   Injection Strategy Demo     ║").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("╚══════════════════════════════╝").
			FgColor("cyan").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("Strategy: "),
			ui.NewTextBuilder(strategy).
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
		ui.NewButtonBuilder("  [ + ]  ").
			OnPress(IncrementStrategyIntent{}).
			Build(),
		ui.NewButtonBuilder("  [ - ]  ").
			OnPress(DecrementStrategyIntent{}).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("──────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Strategies:").
			FgColor("bright-black").
			Build(),
		ui.Text("  • Prohibited - 生产环境"),
		ui.Text("  • Allowed - 测试环境"),
		ui.Text("  • Recorded - 仅录制"),
	)
}

// ============================================================================
// Main
// ============================================================================

func main() {
	err := ui.Run(StrategyApp,
		ui.WithWidth(40),
		ui.WithHeight(18),
		ui.WithTitle("Injection Strategy Demo (Store 模式)"),
	)
	if err != nil {
		panic(err)
	}
}
