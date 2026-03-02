// examples/fiber_counter/main.go - Fiber 模式计数器示例（修复版）
//
// 采用【模式2：全局状态】runtime/intent 内置 Intent
// 适用于跨组件共享的状态，使用 Key 标识状态位置
//
// 三种 Intent 管理模式：
//   1. 组件级状态 - ui.On + UseState + Simple* Intent（推荐组件内状态）
//   2. 全局状态 - runtime/intent 内置函数（本示例）
//   3. 自定义 Intent - 自定义类型 + ui.On（复杂场景）
//
// 详细说明请参考: docs/architecture/mvp/INTENT_MANAGEMENT_PATTERNS.md

package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// 主组件
// =============================================================================

func SimpleCounter() ui.VNode {
	// 获取当前组件上下文
	ctx := ui.GetCurrentContext()

	// 从 GlobalState 读取计数（通过 Key "count" 标识）
	// Intent 内置 handler 会自动处理 Increment/Decrement 操作
	count := ctx.GetIntState("count", 0)

	// 检查是否使用 Fiber 模式
	isFiber := os.Getenv("MINT_USE_FIBER") == "true"

	return ui.VStack(
		ui.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
			FgColor("green").
			Build(),
		ui.HStack(
			ui.NewButtonBuilder(" - ").
				// ✅ 使用内置 Intent - 会自动注册 handler 处理
				OnPress(intent.Decrement("count", 1)).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" + ").
				OnPress(intent.Increment("count", 1)).
				Build(),

		),
		ui.Text(""),
		ui.NewTextBuilder(fmt.Sprintf("[Fiber: %v] 全局状态模式", isFiber)).
			FgColor("bright-black").
			Build(),
	)
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	// 内置 Intent 已自动注册 handler，无需 WithInit
	err := ui.Run(SimpleCounter,
		ui.WithWidth(40),
		ui.WithHeight(10),
		ui.WithTitle("Fiber Counter - GlobalState 模式"),
	)

	if err != nil {
		panic(err)
	}
}
