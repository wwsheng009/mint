// examples/fiber_counter/main.go - Fiber 模式计数器示例（带调试）
package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/ui"
)

// DebugCounter 是带调试输出的计数器组件
func DebugCounter() ui.VNode {
	count, setCount, _, hookIndex := ui.UseStateIntWithDebug(0)

	// Debug: log what value we got
	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "[DebugCounter] Using count=%d, hookIndex=%d\n", count, hookIndex)
	}

	// Create the count text with logging
	countTextStr := fmt.Sprintf("Count: %d (hookIndex=%d)", count, hookIndex)
	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "[DebugCounter] Creating TextVNode with content: %s\n", countTextStr)
	}
	countText := ui.NewTextBuilder(countTextStr).
		FgColor("green").
		Build()

	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "[DebugCounter] Created TextVNode ptr=%p, content=%s\n", countText, countText.(*ui.TextVNode).Content())
	}

	return ui.VStack(
		ui.NewTextBuilder("=== Fiber Counter (Debug Mode) ===").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.Text(""),
		countText,
		ui.Text(""),
		ui.HStack(
			ui.ButtonBuilder("  -  ").
				OnClick(func() {
					fmt.Fprintf(os.Stderr, "[DEBUG] onClick: decrement called, current count=%d\n", count)
					setCount(func(c int) int {
						newVal := c - 1
						fmt.Fprintf(os.Stderr, "[DEBUG] setState: %d -> %d\n", c, newVal)
						return newVal
					})
				}).
				Build(),
			ui.Text("   "),
			ui.ButtonBuilder("  +  ").
				OnClick(func() {
					fmt.Fprintf(os.Stderr, "[DEBUG] onClick: increment called, current count=%d\n", count)
					setCount(func(c int) int {
						newVal := c + 1
						fmt.Fprintf(os.Stderr, "[DEBUG] setState: %d -> %d\n", c, newVal)
						return newVal
					})
				}).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("Tab: focus | Enter: click | q: quit").
			FgColor("bright-black").
			Build(),
	)
}

// SimpleCounter 简化的计数器（用于对比）
func SimpleCounter() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)

	return ui.VStack(
		ui.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
			FgColor("green").
			Build(),
		ui.HStack(
			ui.ButtonBuilder(" - ").
				OnClick(func() { setCount(func(c int) int { return c - 1 }) }).
				Build(),
			ui.Text(" "),
			ui.ButtonBuilder(" + ").
				OnClick(func() { setCount(func(c int) int { return c + 1 }) }).
				Build(),
		),
	)
}

func main() {
	// 检查环境变量
	useFiber := os.Getenv("MINT_USE_FIBER") == "true"
	debugUI := os.Getenv("TUI_DEBUG_UI") == "true"

	fmt.Fprintf(os.Stderr, "=== Fiber Counter Debug Info ===\n")
	fmt.Fprintf(os.Stderr, "MINT_USE_FIBER: %v\n", useFiber)
	fmt.Fprintf(os.Stderr, "TUI_DEBUG_UI: %v\n", debugUI)
	fmt.Fprintf(os.Stderr, "==============================\n\n")

	if useFiber {
		fmt.Fprintf(os.Stderr, "Running in FIBER mode\n")
	} else {
		fmt.Fprintf(os.Stderr, "Running in LEGACY mode\n")
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
