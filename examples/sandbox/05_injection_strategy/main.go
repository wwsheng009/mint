// 05_injection_strategy/main.go
// 事件注入策略示例
//
// 演示不同的事件注入策略：
// - InjectProhibited: 禁止注入（生产环境）
// - InjectAllowed: 允许注入（测试环境）
// - InjectRecorded: 仅录制注入

package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

// Intent Types
type IncrementStrategyIntent struct{}
func (IncrementStrategyIntent) IntentType() string { return "Increment" }
func (IncrementStrategyIntent) StayPressed() bool  { return true }

type DecrementStrategyIntent struct{}
func (DecrementStrategyIntent) IntentType() string { return "Decrement" }
func (DecrementStrategyIntent) StayPressed() bool  { return true }

// StrategyApp 演示注入策略的应用
func StrategyApp() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)
	strategy, _ := ui.UseStateString("Allowed")

	// 将 setter 保存到 GlobalState，供 handler 从 ActionContext 读取
	ctx := ui.GetCurrentContext()
	if ctx != nil {
		ctx.GlobalState["count"] = count
		ctx.GlobalState["setCount"] = setCount
	}

	// Register intent handlers
	ui.On(IncrementStrategyIntent{}, func(actx *intent.ActionContext) {
		currentCount := actx.GetIntState("count", 0)
		if fn, ok := actx.GetState("setCount"); ok {
			if setter, ok := fn.(func(int)); ok {
				setter(currentCount + 1)
			}
		}
	})
	ui.On(DecrementStrategyIntent{}, func(actx *intent.ActionContext) {
		currentCount := actx.GetIntState("count", 0)
		if currentCount > 0 {
			if fn, ok := actx.GetState("setCount"); ok {
				if setter, ok := fn.(func(int)); ok {
					setter(currentCount - 1)
				}
			}
		}
	})

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

func main() {
	err := ui.Run(StrategyApp,
		ui.WithWidth(40),
		ui.WithHeight(18),
		ui.WithTitle("Injection Strategy Demo"),
	)
	if err != nil {
		panic(err)
	}
}
