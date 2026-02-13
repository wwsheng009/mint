// examples/fiber_counter/main.go - Fiber 模式计数器示例（带调试）
package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/ui"
)

// DebugCounter 是带调试输出的计数器组件
func DebugCounter() ui.VNode {
	count, setCount, _, hookIndex := ui.UseStateIntWithDebug(0)

	// Debug: log what value we got
	if log.UILogger.Enabled() {
		log.UILogger.Debug( "[DebugCounter] Using count=%d, hookIndex=%d\n", count, hookIndex)
	}

	// Create the count text with logging
	countTextStr := fmt.Sprintf("Count: %d (hookIndex=%d)", count, hookIndex)
	if log.UILogger.Enabled() {
		log.UILogger.Debug("[DebugCounter] Creating TextVNode with content: %s\n", countTextStr)
	}
	countText := app.NewTextBuilder(countTextStr).
		FgColor("green").
		Build()

	if log.UILogger.Enabled() {
		// Use Props() to get content without type assertion
		content := ""
		if countText.Props() != nil {
			content = countText.Props().GetString("content")
		}
		log.UILogger.Debug("[DebugCounter] Created TextVNode ptr=%p, content=%s\n", countText, content)
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
				OnClick(func() {
					log.UILogger.Debug("[DEBUG] onClick: decrement called, current count=%d\n", count)
					setCount(func(c int) int {
						newVal := c - 1
						log.UILogger.Debug("[DEBUG] setState: %d -> %d\n", c, newVal)
						return newVal
					})
				}).
				Build(),
			ui.Text("   "),
			app.ButtonBuilder("  +  ").
				OnClick(func() {
					log.UILogger.Debug("[DEBUG] onClick: increment called, current count=%d\n", count)
					setCount(func(c int) int {
						newVal := c + 1
						log.UILogger.Debug("[DEBUG] setState: %d -> %d\n", c, newVal)
						return newVal
					})
				}).
				Build(),
		),
		ui.Text(""),
		app.NewTextBuilder("Tab: focus | Enter: click | q: quit").
			FgColor("bright-black").
			Build(),
	)
}

// SimpleCounter 简化的计数器（用于对比）
func SimpleCounter() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)

	return ui.VStack(
		app.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
			FgColor("green").
			Build(),
		ui.HStack(
			app.ButtonBuilder(" - ").
				OnClick(func() { setCount(func(c int) int { return c - 1 }) }).
				Build(),
			ui.Text(" "),
			app.ButtonBuilder(" + ").
				OnClick(func() { setCount(func(c int) int { return c + 1 }) }).
				Build(),
		),
	)
}

func main() {
	// 检查环境变量
	useFiber := os.Getenv("MINT_USE_FIBER") == "true"
	debugUI := log.UILogger.Enabled()

	log.UILogger.Debug("=== Fiber Counter Debug Info ===\n")
	log.UILogger.Debug("MINT_USE_FIBER: %v\n", useFiber)
	log.UILogger.Debug("TUI_DEBUG_UI: %v\n", debugUI)
	log.UILogger.Debug("==============================\n\n")

	if useFiber {
		log.UILogger.Debug("Running in FIBER mode\n")
	} else {
		log.UILogger.Debug("Running in LEGACY mode\n")
	}

	var app ui.ComponentFunc
	if debugUI {
		app = DebugCounter
	} else {
		app = SimpleCounter
	}

	err := ui.Run(app,
		ui.WithWidth(40),
		ui.WithHeight(10),
		ui.WithTitle("Fiber Counter"),
	)
	if err != nil {
		panic(err)
	}
}
