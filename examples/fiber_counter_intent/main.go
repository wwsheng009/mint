package main

import (
	"fmt"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// 方案 1: 使用 ui.On + ui 内置通用 Intent（推荐组件内状态）
// =============================================================================
// 这种方式适用于组件内部的局部状态，状态与组件绑定，通过闭包自然访问
// 无需定义 Intent 类型，系统提供通用的 SimpleIncrementIntent/SimpleDecrementIntent

func CounterWithHooks() ui.VNode {
	// 使用 hooks 状态管理
	count, setCount, _ := ui.UseStateInt(0)

	// 注册通用 Intent 处理器（无参数，行为由闭包决定）
	ui.On(ui.SimpleIncrementIntent{}, func() {
		// ✅ 使用函数形式的 setCount 避免闭包捕获旧值
		setCount(func(c int) int { return c + 1 })
	})

	ui.On(ui.SimpleDecrementIntent{}, func() {
		setCount(func(c int) int { return c - 1 })
	})

	return ui.VStack(
		app.NewTextBuilder(fmt.Sprintf("Count: %d", count)).FgColor("green").Build(),
		ui.HStack(
			app.ButtonBuilder(" - ").OnPress(ui.SimpleDecrementIntent{}).Build(),
			ui.Text(" "),
			app.ButtonBuilder(" + ").OnPress(ui.SimpleIncrementIntent{}).Build(),
		),
		app.Text(""),
		app.NewTextBuilder("[方式1: ui.On + UseState]").FgColor("cyan").Build(),
		app.Text("  组件内状态，闭包自然访问"),
	)
}

// =============================================================================
// 方案 2: 使用 runtime/intent 内置函数（推荐全局状态）
// =============================================================================
// 这种方式适用于跨组件共享的全局状态，通过 Key 标识
// 系统内置 handler，无需自定义

func CounterWithGlobalState() ui.VNode {
	ctx := ui.GetCurrentContext()

	// 从全局状态读取计数
	count := ctx.GetIntState("counter2", 0)

	return ui.VStack(
		app.NewTextBuilder(fmt.Sprintf("Count: %d", count)).FgColor("yellow").Build(),
		ui.HStack(
			app.ButtonBuilder(" - ").OnPress(intent.Decrement("counter2", 1)).Build(),
			ui.Text(" "),
			app.ButtonBuilder(" + ").OnPress(intent.Increment("counter2", 1)).Build(),
		),
		app.Text(""),
		app.NewTextBuilder("[方式2: intent.Increment]").FgColor("cyan").Build(),
		app.Text("  全局状态，跨组件共享"),
	)
}

// =============================================================================
// 方案 3: 自定义 Intent 类型 + ui.On（推荐复杂场景）
// =============================================================================
// 这种方式适用于需要传递参数或特定业务逻辑的场景

// CustomIncrement 自定义带步长的递增 Intent
type CustomIncrement struct {
	Step int
}

func (CustomIncrement) IntentType() string { return "CustomIncrement" }
func (CustomIncrement) StayPressed() bool  { return true }

func CounterWithCustomIntent() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)

	// 注册自定义 Intent 处理器
	ui.On(CustomIncrement{Step: 10}, func() {
		setCount(func(c int) int { return c + 10 })
	})
	return ui.VStack(
		app.NewTextBuilder(fmt.Sprintf("Count: %d", count)).FgColor("magenta").Build(),
		ui.HStack(
			app.ButtonBuilder(" +10 ").OnPress(CustomIncrement{Step: 10}).Build(),
		),
		app.Text(""),
		app.NewTextBuilder("[方式3: 自定义 Intent]").FgColor("cyan").Build(),
		app.Text("  支持参数传递"),
	)
}

// =============================================================================
// 主界面：展示三种方案
// =============================================================================

func SimpleCounter() ui.VNode {
	return ui.VStack(
		app.NewTextBuilder("Mint UI - Intent 管理模式").FgColor("bright-cyan").Build(),
		app.Text(""),
		app.NewTextBuilder("【方案1】组件级状态").FgColor("white").Build(),
		CounterWithHooks(),
		app.Text(""),
		app.NewTextBuilder("【方案2】全局状态").FgColor("white").Build(),
		CounterWithGlobalState(),
		app.Text(""),
		app.NewTextBuilder("【方案3】自定义 Intent").FgColor("white").Build(),
		CounterWithCustomIntent(),
	)
}

func main() {
	err := ui.Run(SimpleCounter,
		ui.WithWidth(40),
		ui.WithHeight(30),
		ui.WithTitle("Intent 管理模式对比"),
	)
	if err != nil {
		panic(err)
	}
}
