// Package main demonstrates Fiber architecture basics (Store 模式)
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
func (DecrementIntent) IntentType() string   { return "Decrement" }
func (DecrementIntent) StayPressed() bool    { return true }

type IncrementIntent struct{}
func (IncrementIntent) IntentType() string   { return "Increment" }
func (IncrementIntent) StayPressed() bool    { return true }

// =============================================================================
// Store 初始化
// =============================================================================

var demoStore = store.NewStore(AppState{
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
		BuildAndRegister(intent.DefaultRegistry(), demoStore)
}

// =============================================================================
// Demo Component
// =============================================================================

func DemoComponent() ui.VNode {
	// ✅ 使用 UseStoreSelector 订阅 count 字段
	count := ui.UseStoreSelector(
		demoStore,
		func(s AppState) int { return s.Count },
	)

	return ui.VStack(
		ui.NewTextBuilder("Fiber Architecture Demo").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Count: "+fmt.Sprintf("%d", count)).
			FgColor("green").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.NewButtonBuilder(" - ").
				// ✅ 使用自定义 Intent - 由 Reducer 处理
				OnPress(DecrementIntent{}).
				Build(),
			ui.Text("  "),
			ui.NewButtonBuilder(" + ").
				OnPress(IncrementIntent{}).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("Fiber enables:").
			FgColor("yellow").
			Build(),
		ui.NewTextBuilder("  • Interruptible rendering").
			FgColor("bright-black").
			Build(),
		ui.NewTextBuilder("  • Priority scheduling").
			FgColor("bright-black").
			Build(),
		ui.NewTextBuilder("  • Concurrent updates").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Press q to quit").
			FgColor("bright-black").
			Build(),
	)
}

// =============================================================================
// Main
// =============================================================================

func main() {
	fmt.Println("=== Fiber Architecture Demo ===")
	fmt.Println()
	fmt.Println("This demo shows the core concepts of Fiber architecture:")
	fmt.Println("1. FiberNode structure")
	fmt.Println("2. Virtual DOM diffing")
	fmt.Println("3. Reconciliation process")
	fmt.Println()
	fmt.Println("Key concepts:")
	fmt.Println("- Each component has a corresponding FiberNode")
	fmt.Println("- FiberNode stores hooks, state, and VNode")
	fmt.Println("- Reconciliation updates the tree efficiently")
	fmt.Println()
	fmt.Println("See the following files for implementation:")
	fmt.Println("  - runtime/ui/fiber.go: FiberNode definition")
	fmt.Println("  - runtime/ui/fiber_vnode.go: VNode diffing")
	fmt.Println("  - runtime/ui/hooks.go: Hooks implementation")
	fmt.Println()

	// Run the demo app
	runDemoApp()
}

func runDemoApp() {
	fmt.Println("--- Running Demo App ---")
	fmt.Println()

	// ✅ 无需 WithInit 或 GlobalState 注册
	err := ui.Run(DemoComponent,
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Fiber Demo"),
	)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
