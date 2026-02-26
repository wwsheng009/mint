// examples/fiber_counter/main_one_space.go - Fiber 模式计数器示例（单空格测试版）
//
// 使用 GlobalState + 内置 Intent 模式（与 examples/counter 相同）
// 仅使用单个空格（会触发按钮重复问题）
package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// 主组件
// =============================================================================

func SimpleCounter() ui.VNode {
	// 获取当前组件上下文
	ctx := ui.GetCurrentContext()

	// ✅ 正确方式：直接从 GlobalState 读取值（不使用 UseStateInt）
	// GlobalState 由 Intent Handler 更新
	count := ctx.GetIntState("count", 0)

	// 检查是否使用 Fiber 模式
	isFiber := os.Getenv("MINT_USE_FIBER") == "true"

	return ui.VStack(
		app.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
			FgColor("green").
			Build(),
		ui.HStack(
			app.ButtonBuilder(" - ").
				// ✅ 使用内置 Intent - 会自动注册 handler 处理
				OnPress(intent.Decrement("count", 1)).
				Build(),
			ui.Text(" "),  // ⚠️ 单个空格 - 会触发按钮重复问题
			app.ButtonBuilder(" + ").
				OnPress(intent.Increment("count", 1)).
				Build(),

		),
		app.Text(""),
		app.NewTextBuilder(fmt.Sprintf("[Fiber: %v] Using GlobalState + Built-in Intent", isFiber)).
			FgColor("bright-black").
			Build(),
	)
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	// 不需要自定义 WithInit - 内置 Intent 已经自动注册
	err := ui.Run(SimpleCounter,
		ui.WithWidth(40),
		ui.WithHeight(10),
		ui.WithTitle("Fiber Counter (One-Space Test)"),
	)

	if err != nil {
		panic(err)
	}
}
