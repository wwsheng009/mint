package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

// Counter is a dynamic counter component using ComponentContext and Intent
// 采用【模式2：全局状态】runtime/intent 内置 Intent
// 适用于跨组件共享的状态，使用 Key 标识状态位置
//
// 三种 Intent 管理模式：
//   1. 组件级状态 - ui.On + UseState + Simple* Intent（推荐组件内状态）
//   2. 全局状态 - runtime/intent 内置函数（本示例）
//   3. 自定义 Intent - 自定义类型 + ui.On（复杂场景）
//
// 详细说明请参考: docs/architecture/mvp/INTENT_MANAGEMENT_PATTERNS.md
func Counter() ui.VNode {
	// Get current context (initialized by Fiber-first framework during render)
	ctx := ui.GetCurrentContext()

	// Read count from GlobalState (incremented by IncrementIntent)
	// 使用全局状态，通过 Key "count" 标识
	// IncrementIntent 内置 handler 会自动处理
	count := ctx.GetIntState("count", 0)

	// Check if running in Fiber mode
	isFiber := os.Getenv("MINT_USE_FIBER") == "true"
	fiberStr := "OFF"
	if isFiber {
		fiberStr = "ON"
	}

	log.TempLogger.Debug("[Counter] Render: count=%d, Fiber=%s", count, fiberStr)

	return ui.VStack(
		ui.NewTextBuilder("Mint UI Counter Demo").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
			FgColor("green").
			Build(),
		ui.Text(""),
		// Debug info line
		ui.NewTextBuilder(fmt.Sprintf("[Fiber: %s] Using GlobalState + IncrementIntent", fiberStr)).
			FgColor("yellow").
			Build(),
		ui.Text(""),
		ui.HStack(
			// Decrement button using Intent (Fiber-first)
			// Intent created with fresh parameter values at render time
			ui.NewButtonBuilder("  -  ").
				OnPress(intent.Decrement("count", 1)).
				Build(),
			ui.Text("   "),
			// Increment button using Intent (Fiber-first)
			// Intent created with fresh parameter values at render time
			ui.NewButtonBuilder("  +  ").
				OnPress(intent.Increment("count", 1)).
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

func main() {
	err := ui.Run(Counter,
		ui.WithWidth(40),
		ui.WithHeight(12),
		ui.WithTitle("Counter Demo"),
	)
	if err != nil {
		panic(err)
	}
}
