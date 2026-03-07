// Package main - Fiber reconciler 测试示例（Store 模式）
// 演示如何在 Fiber 模式下使用 Store + Reducer 架构
package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// AppState - 定义应用状态
// =============================================================================

type AppState struct {
	Count int
}

// =============================================================================
// Intent Types
// =============================================================================

type DecrementIntent struct{}
func (DecrementIntent) IntentType() string { return "Decrement" }
func (DecrementIntent) StayPressed() bool  { return true }

type IncrementIntent struct{}
func (IncrementIntent) IntentType() string { return "Increment" }
func (IncrementIntent) StayPressed() bool  { return true }

// =============================================================================
// Store 初始化
// =============================================================================

var fiberStore = store.NewStore(AppState{
	Count: 0,
})

// =============================================================================
// Reducer 注册
// =============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(DecrementIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Count--
			return s
		}).
		On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Count++
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), fiberStore)
}

// =============================================================================
// SimpleCounter - 计数器组件
// =============================================================================

func SimpleCounter() ui.VNode {
	// ✅ 使用 UseStoreSelector 订阅 count 字段
	count := ui.UseStoreSelector(
		fiberStore,
		func(s AppState) int { return s.Count },
	)

	return ui.VStack(
		ui.NewTextBuilder("Fiber Reconciler Test").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
			FgColor("green").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.NewButtonBuilder("  -  ").
				// ✅ 使用自定义 Intent - 由 Reducer 处理，无需类型断言
				OnPress(DecrementIntent{}).
				Build(),
			ui.Text("   "),
			ui.NewButtonBuilder("  +  ").
				OnPress(IncrementIntent{}).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("Fiber Mode: ENABLED").
			FgColor("yellow").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Tab/Arrows: focus | Enter/Space: click | q: quit").
			FgColor("bright-black").
			Build(),
	)
}

// =============================================================================
// Main
// =============================================================================

func main() {
	// Fiber is enabled via MINT_USE_FIBER environment variable
	// Run like: MINT_USE_FIBER=true go run examples/fiber/main.go
	// ✅ 无需 WithInit 或 GlobalState 注册
	err := ui.Run(SimpleCounter,
		ui.WithWidth(40),
		ui.WithHeight(14),
		ui.WithTitle("Fiber Test"),
	)
	if err != nil {
		panic(err)
	}
}
