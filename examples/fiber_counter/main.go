// examples/fiber_counter/main.go - Fiber 模式计数器示例（带调试）
//
// 改进：使用简洁的 ui.On() API，避免手动保存 setter 到 GlobalState
//
package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// 简单的 Intent 定义（仅用于类型标识）
// =============================================================================

type IncrementIntent struct{}
func (IncrementIntent) IntentType() string { return "Increment" }
func (IncrementIntent) StayPressed() bool  { return true }

type DecrementIntent struct{}
func (DecrementIntent) IntentType() string { return "Decrement" }
func (DecrementIntent) StayPressed() bool  { return true }

// =============================================================================
// 辅助函数：On() - 简洁的 Intent 处理器注册
// =============================================================================

// On 注册一个意图处理器，使用闭包直接访问局部变量
// 使用标签确保只注册一次（避免每次渲染都注册）
func On[T interface{ IntentType() string; StayPressed() bool }](intentType T, handler func()) {
	ctx := ui.GetCurrentContext()
	if ctx == nil {
		return
	}

	// 使用标签确保只注册一次
	registryKey := "__on_handler_registered_" + intentType.IntentType()
	if _, exists := ctx.GlobalState[registryKey]; exists {
		return
	}

	ctx.GlobalState[registryKey] = true
	ui.RegisterIntent(func(ctx *intent.ActionContext, i T) intent.IntentResult {
		handler()
		return intent.HandledResult()
	})
}

// =============================================================================
// 组件
// =============================================================================

// SimpleCounter 简化的计数器
func SimpleCounter() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)

	// ✅ 简洁：直接使用 On 注册处理器，无需手动保存 setter
	On(IncrementIntent{}, func() {
		setCount(count + 1)
	})

	On(DecrementIntent{}, func() {
		setCount(count - 1)
	})

	return ui.VStack(
		app.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
			FgColor("green").
			Build(),
		ui.HStack(
			app.ButtonBuilder(" - ").
				OnPress(DecrementIntent{}).
				Build(),
			ui.Text(" "),
			app.ButtonBuilder(" + ").
				OnPress(IncrementIntent{}).
				Build(),
		),
	)
}

// DebugCounter 是带调试输出的计数器组件
func DebugCounter() ui.VNode {
	count, setCount, _, hookIndex := ui.UseStateIntWithDebug(0)

	if log.UILogger.Enabled() {
		log.UILogger.Debug("[DebugCounter] Using count=%d, hookIndex=%d", count, hookIndex)
	}

	// ✅ 使用 On 注册处理器
	On(IncrementIntent{}, func() {
		log.UILogger.Debug("[DebugCounter] Incrementing from %d", count)
		setCount(count + 1)
	})

	On(DecrementIntent{}, func() {
		log.UILogger.Debug("[DebugCounter] Decrementing from %d", count)
		setCount(count - 1)
	})

	// Create the count text with logging
	countTextStr := fmt.Sprintf("Count: %d (hookIndex=%d)", count, hookIndex)
	log.UILogger.Debug("[DebugCounter] Creating TextVNode with content: %s", countTextStr)

	countText := app.NewTextBuilder(countTextStr).
		FgColor("green").
		Build()

	if log.UILogger.Enabled() {
		content := ""
		if countText.Props() != nil {
			content = countText.Props().GetString("content")
		}
		log.UILogger.Debug("[DebugCounter] Created TextVNode ptr=%p, content=%s", countText, content)
	}

	return ui.VStack(
		app.NewTextBuilder("=== Fiber Counter (Debug Mode) ===").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.Text(""),
		countText,
		ui.Text(""),
		ui.HStack(
			app.ButtonBuilder("  -  ").
				OnPress(DecrementIntent{}).
				Build(),
			ui.Text("   "),
			app.ButtonBuilder("  +  ").
				OnPress(IncrementIntent{}).
				Build(),
		),
		ui.Text(""),
		app.NewTextBuilder("Tab: focus | Enter: click | q: quit").
			FgColor("bright-black").
			Build(),
	)
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	// 检查环境变量
	useFiber := os.Getenv("MINT_USE_FIBER") == "true"
	debugUI := log.UILogger.Enabled()

	log.UILogger.Debug("=== Fiber Counter Debug Info ===")
	log.UILogger.Debug("MINT_USE_FIBER: %v", useFiber)
	log.UILogger.Debug("TUI_DEBUG_UI: %v", debugUI)
	log.UILogger.Debug("==============================")

	if useFiber {
		log.UILogger.Debug("Running in FIBER mode")
	} else {
		log.UILogger.Debug("Running in LEGACY mode")
	}

	var app ui.ComponentFunc
	if debugUI {
		app = DebugCounter
	} else {
		app = SimpleCounter
	}

	// ✅ 不需要 WithInit，意图处理器在组件内部通过 On() 注册
	err := ui.Run(app,
		ui.WithWidth(40),
		ui.WithHeight(10),
		ui.WithTitle("Fiber Counter (Improved API)"),
	)

	if err != nil {
		panic(err)
	}
}
