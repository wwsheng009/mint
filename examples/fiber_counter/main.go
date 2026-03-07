// examples/fiber_counter/main.go - Fiber 模式计数器示例（Store 模式）
//
// 采用 Store + Reducer 架构，提供类型安全的状态管理
//
// 三种状态管理模式：
//   1. useState (局部状态) - 适用于简单的组件内部状态
//   2. UseStoreField (订阅字段) - 适用于从 Store 订阅特定字段
//   3. Store + Reducer (应用级状态) - 本示例使用
//
// 详细说明请参考:
//   - Store 架构: docs/ui/store/README.md
//   - 迁移指南: docs/ui/store/guides/MIGRATION_GUIDE.md

package main

import (
	"fmt"
	"os"

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
// 自定义 Intent 类型
// =============================================================================

type IncrementIntent struct{}
func (IncrementIntent) IntentType() string { return "Increment" }
func (IncrementIntent) StayPressed() bool  { return true }

type DecrementIntent struct{}
func (DecrementIntent) IntentType() string { return "Decrement" }
func (DecrementIntent) StayPressed() bool  { return true }

// =============================================================================
// Store 初始化
// =============================================================================

var counterStore = store.NewStore(AppState{
	Count: 0,
})

// =============================================================================
// Reducer 注册
// =============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Count++
			return s
		}).
		On(DecrementIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Count--
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), counterStore)
}

// =============================================================================
// 主组件
// =============================================================================

func SimpleCounter() ui.VNode {
	// ✅ 从 Store 订阅count字段 - 类型安全，自动订阅更新
	count := ui.UseStoreSelector(
		counterStore,
		func(s AppState) int { return s.Count },
	)

	// 检查是否使用 Fiber 模式
	isFiber := os.Getenv("MINT_USE_FIBER") == "true"

	return ui.VStack(
		ui.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
			FgColor("green").
			Build(),
		ui.HStack(
			ui.NewButtonBuilder(" - ").
				// ✅ 使用自定义 Intent - 由 Reducer 处理
				OnPress(DecrementIntent{}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" + ").
				OnPress(IncrementIntent{}).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("[Fiber: %v] Store + Reducer 模式", isFiber)).
			FgColor("bright-black").
			Build(),
	)
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	// ✅ 无需 WithInit 或 GlobalState 注册
	// Store 在包级别初始化，所有组件共享

	err := ui.Run(SimpleCounter,
		ui.WithWidth(40),
		ui.WithHeight(10),
		ui.WithTitle("Fiber Counter - Store 模式"),
	)

	if err != nil {
		panic(err)
	}
}
