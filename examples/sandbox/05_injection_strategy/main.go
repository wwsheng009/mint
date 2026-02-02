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

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

// StrategyApp 演示注入策略的应用
func StrategyApp() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)
	strategy, _ := ui.UseStateString("Allowed")

	return ui.VStack(
		app.NewTextBuilder("╔══════════════════════════════╗").
			FgColor("cyan").
			Build(),
		app.NewTextBuilder("║   Injection Strategy Demo     ║").
			FgColor("cyan").
			Build(),
		app.NewTextBuilder("╚══════════════════════════════╝").
			FgColor("cyan").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("Strategy: "),
			app.NewTextBuilder(strategy).
				FgColor("yellow").
				Bold(true).
				Build(),
		),
		ui.Text(""),
		app.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
			FgColor("green").
			Bold(true).
			Build(),
		ui.Text(""),
		app.ButtonBuilder("  [ + ]  ").
			OnClick(func() {
				setCount(count + 1)
			}).
			Build(),
		app.ButtonBuilder("  [ - ]  ").
			OnClick(func() {
				if count > 0 {
					setCount(count - 1)
				}
			}).
			Build(),
		ui.Text(""),
		app.NewTextBuilder("──────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		app.NewTextBuilder("Strategies:").
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
