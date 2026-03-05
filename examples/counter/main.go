// Counter Demo using Store + Reducer architecture
//
// 以下三种 Intent 管理模式对比：
//   1. 组件级状态（旧）- UseState + Simple* Intent - 推荐用于组件内状态
//   2. 全局状态（旧）- GlobalState + runtime/intent 内置 Intent - 适用于跨组件共享
//   3. Store + Reducer（新）✅ - 单一状态源 + 纯函数 + 自动注册 - 推荐用于生产环境
//
// 架构优势：
//   - 单一状态源
//   - 纯函数 Reducer
//   - 编译期类型检查（无类型断言）
//   - 自动注册（无需手动注册 Handler）
//   - 数据流清晰（UI.Instance → Intent → Reducer → Store → VNode）
//
// 运行: go run main.go
//
// 详细说明请参考: docs/architecture/store/DEVELOPMENT_GUIDE.md
package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// AppState (Single Source of Truth)
// =============================================================================

// AppState represents the counter state.
type AppState struct {
	Count int
}

// =============================================================================
// Custom Intent Types (替代 runtime/intent 内置 Intent)
// =============================================================================

// IncrementIntent increments the count.
type IncrementIntent struct {
	Amount int
}

func (IncrementIntent) IntentType() string { return "Increment" }
func (IncrementIntent) StayPressed() bool  { return true }

// DecrementIntent decrements the count.
type DecrementIntent struct {
	Amount int
}

func (DecrementIntent) IntentType() string { return "Decrement" }
func (DecrementIntent) StayPressed() bool  { return true }

// =============================================================================
// Reducer (Pure Function)
// =============================================================================

// appReducer handles all state transitions for the counter.
var appReducer = reducer.NewBuilder[AppState]()

// Initialize the reducer.
func init() {
	// Handle IncrementIntent
	appReducer.On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
		ii := i.(IncrementIntent)
		s.Count += ii.Amount
		return s
	})

	// Handle DecrementIntent
	appReducer.On(DecrementIntent{}, func(s AppState, i intent.Intent) AppState {
		di := i.(DecrementIntent)
		s.Count -= di.Amount
		return s
	})
}

// =============================================================================
// Store (Single State Source)
// =============================================================================

// appStore holds the counter state.
var appStore = store.NewStore(AppState{
	Count: 0,
})

// =============================================================================
// Counter Component
// =============================================================================

func Counter() ui.VNode {
	// Get current state snapshot from Store
	state := appStore.Get()

	// Check if running in Fiber mode
	isFiber := os.Getenv("MINT_USE_FIBER") == "true"
	fiberStr := "OFF"
	if isFiber {
		fiberStr = "ON"
	}

	log.TempLogger.Debug("[Counter] Render: count=%d, Fiber=%s", state.Count, fiberStr)

	return ui.VStack(
		ui.NewTextBuilder("Mint UI Counter Demo").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("Count: %d", state.Count)).
			FgColor("green").
			Build(),
		ui.Text(""),
		// Debug info line
		ui.NewTextBuilder(fmt.Sprintf("[Fiber: %s] Store + Reducer + Custom Intent", fiberStr)).
			FgColor("yellow").
			Build(),
		ui.Text(""),
		ui.HStack(
			// Decrement button using custom DecrementIntent
			ui.NewButtonBuilder("  -  ").
				OnPress(DecrementIntent{Amount: 1}).
				Build(),
			ui.Text("   "),
			// Increment button using custom IncrementIntent
			ui.NewButtonBuilder("  +  ").
				OnPress(IncrementIntent{Amount: 1}).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("Focused button = BLUE background").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Tab/Arrows: focus | Enter/Space: click | q: quit").
			FgColor("bright-black").
			Build(),
	)
}

// =============================================================================
// Main Function
// =============================================================================

func main() {
	// Register all handlers automatically
	appReducer.RegisterToGlobal(appStore)

	err := ui.Run(Counter,
		ui.WithWidth(40),
		ui.WithHeight(12),
		ui.WithTitle("Counter Demo (Store+Reducer)"),
	)
	if err != nil {
		panic(err)
	}
}
