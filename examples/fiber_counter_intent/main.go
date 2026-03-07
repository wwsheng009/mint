package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// AppState - 定义应用状态（三个计数器）
// =============================================================================

type AppState struct {
	Counter1 int // 用于方案1
	Counter2 int // 用于方案2
	Counter3 int // 用于方案3
}

// =============================================================================
// Intent Types - 演示不同的 Intent 定义方式
// =============================================================================

// 方案1: 简单的自定义 Intent（无额外字段）
type Counter1Increment struct{}
func (Counter1Increment) IntentType() string      { return "Counter1Increment" }
func (Counter1Increment) StayPressed() bool       { return true }

type Counter1Decrement struct{}
func (Counter1Decrement) IntentType() string      { return "Counter1Decrement" }
func (Counter1Decrement) StayPressed() bool       { return true }

// 方案2: 需要参数的 Intent（用于通用递增/递减）
// 实际上在 Store 模式下，每个计数器有独立的 Store 实例
type Counter2Increment struct{}
func (Counter2Increment) IntentType() string      { return "Counter2Increment" }
func (Counter2Increment) StayPressed() bool       { return true }

type Counter2Decrement struct{}
func (Counter2Decrement) IntentType() string      { return "Counter2Decrement" }
func (Counter2Decrement) StayPressed() bool       { return true }

// 方案3: 带参数的自定义 Intent（步长+10）
type Counter3Increment struct {
	Step int
}
func (Counter3Increment) IntentType() string      { return "Counter3Increment" }
func (Counter3Increment) StayPressed() bool       { return true }

// =============================================================================
// Store 初始化
// =============================================================================

var appStore = store.NewStore(AppState{
	Counter1: 0,
	Counter2: 0,
	Counter3: 0,
})

// =============================================================================
// Reducer 注册
// =============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		// 方案1: Counter1 的递增/递减
		On(Counter1Increment{}, func(s AppState, i intent.Intent) AppState {
			s.Counter1++
			return s
		}).
		On(Counter1Decrement{}, func(s AppState, i intent.Intent) AppState {
			s.Counter1--
			return s
		}).
		// 方案2: Counter2 的递增/递增
		On(Counter2Increment{}, func(s AppState, i intent.Intent) AppState {
			s.Counter2++
			return s
		}).
		On(Counter2Decrement{}, func(s AppState, i intent.Intent) AppState {
			s.Counter2--
			return s
		}).
		// 方案3: Counter3 的跨步递增
		On(Counter3Increment{}, func(s AppState, i intent.Intent) AppState {
			// ✅ 无需类型断言，直接读取 Intent 字段
			if inc, ok := i.(Counter3Increment); ok {
				s.Counter3 += inc.Step
			}
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), appStore)
}

// =============================================================================
// 方案1: 使用简单的自定义 Intent + UseStoreSelector
// =============================================================================

func CounterWithSimpleIntent() ui.VNode {
	// ✅ 订阅 Counter1 字段
	count := ui.UseStoreSelector(
		appStore,
		func(s AppState) int { return s.Counter1 },
	)

	return ui.VStack(
		ui.NewTextBuilder(fmt.Sprintf("Count: %d", count)).FgColor("green").Build(),
		ui.HStack(
			ui.NewButtonBuilder(" - ").OnPress(Counter1Decrement{}).Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" + ").OnPress(Counter1Increment{}).Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("[方式1: 简单的自定义 Intent]").FgColor("cyan").Build(),
		ui.Text("  类型安全的 Store 订阅"),
	)
}

// =============================================================================
// 方案2: 使用独立 Store 的新方式（推荐）
// =============================================================================

// 定义方案2的独立状态
type Counter2State struct {
	Count int
}

type Counter2IncrementIntent struct{}
func (Counter2IncrementIntent) IntentType() string      { return "Counter2IncrementIntent" }
func (Counter2IncrementIntent) StayPressed() bool       { return true }

type Counter2DecrementIntent struct{}
func (Counter2DecrementIntent) IntentType() string      { return "Counter2DecrementIntent" }
func (Counter2DecrementIntent) StayPressed() bool       { return true }

var counter2Store = store.NewStore(Counter2State{Count: 0})

func init() {
	// 注册 Counter2 的 Reducer
	reducer.NewBuilder[Counter2State]().
		On(Counter2IncrementIntent{}, func(s Counter2State, i intent.Intent) Counter2State {
			s.Count++
			return s
		}).
		On(Counter2DecrementIntent{}, func(s Counter2State, i intent.Intent) Counter2State {
			s.Count--
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), counter2Store)
}

func CounterWithStore() ui.VNode {
	// ✅ 订阅 Counter2 的 Count 字段
	count := ui.UseStoreSelector(
		counter2Store,
		func(s Counter2State) int { return s.Count },
	)

	return ui.VStack(
		ui.NewTextBuilder(fmt.Sprintf("Count: %d", count)).FgColor("yellow").Build(),
		ui.HStack(
			ui.NewButtonBuilder(" - ").OnPress(Counter2DecrementIntent{}).Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" + ").OnPress(Counter2IncrementIntent{}).Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("[方式2: 独立 Store + Reducer]").FgColor("cyan").Build(),
		ui.Text("  跨组件共享状态"),
	)
}

// =============================================================================
// 方案3: 带参数的自定义 Intent
// =============================================================================

func CounterWithCustomIntent() ui.VNode {
	// ✅ 订阅 Counter3 字段
	count := ui.UseStoreSelector(
		appStore,
		func(s AppState) int { return s.Counter3 },
	)

	return ui.VStack(
		ui.NewTextBuilder(fmt.Sprintf("Count: %d", count)).FgColor("magenta").Build(),
		ui.HStack(
			ui.NewButtonBuilder(" +10 ").OnPress(Counter3Increment{Step: 10}).Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("[方式3: 带参数的自定义 Intent]").FgColor("cyan").Build(),
		ui.Text("  无需类型断言，直接访问字段"),
	)
}

// =============================================================================
// 主界面：展示三种方案
// =============================================================================

func SimpleCounter() ui.VNode {
	return ui.VStack(
		ui.NewTextBuilder("Mint UI - Intent 管理模式").FgColor("bright-cyan").Build(),
		ui.Text(""),
		ui.NewTextBuilder("【方案1】简单 Intent + Store").FgColor("white").Build(),
		CounterWithSimpleIntent(),
		ui.Text(""),
		ui.NewTextBuilder("【方案2】独立 Store（推荐跨组件）").FgColor("white").Build(),
		CounterWithStore(),
		ui.Text(""),
		ui.NewTextBuilder("【方案3】带参数的 Intent").FgColor("white").Build(),
		CounterWithCustomIntent(),
	)
}

func main() {
	err := ui.Run(SimpleCounter,
		ui.WithWidth(40),
		ui.WithHeight(30),
		ui.WithTitle("Intent 管理模式对比（Store 模式）"),
	)
	if err != nil {
		panic(err)
	}
}
