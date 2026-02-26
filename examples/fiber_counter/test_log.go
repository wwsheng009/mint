package main

import (
	"fmt"
	"os"

	log "github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/app"
)

// SimpleCounter with debug logging
func SimpleCounter() ui.VNode {
	ctx := ui.GetCurrentContext()

	log.UILogger.Debug("[SimpleCounter] Component called, ctx=%p", ctx)

	// GlobalState 读取
	count := ctx.GetIntState("count", 0)
	log.UILogger.Debug("[SimpleCounter] Count=%d read from GlobalState", count)

	// 检查 fiber 模式
	isFiber := os.Getenv("MINT_USE_FIBER") == "true"
	log.UILogger.Debug("[SimpleCounter] Fiber mode: %v", isFiber)

	// 构建 UI
	log.UILogger.Debug("[SimpleCounter] Building VNode tree...")
	result := ui.VStack(
		app.NewTextBuilder(fmt.Sprintf("Count: %d", count)).
			FgColor("green").
			Build(),
		ui.HStack(
			app.ButtonBuilder(" - ").
				OnPress(intent.Decrement("count", 1)).
				Build(),
			ui.Text(" "),
			ui.Text(" "),
			app.ButtonBuilder(" + ").
				OnPress(intent.Increment("count", 1)).
				Build(),
		),
		app.Text(""),
		app.NewTextBuilder(fmt.Sprintf("[Fiber: %v] Using GlobalState + Built-in Intent", isFiber)).
			FgColor("bright-black").
			Build(),
	)

	log.UILogger.Debug("[SimpleCounter] VNode tree built successfully")
	return result
}

func main() {
	os.Setenv("TUI_DEBUG_ALL", "true")
	os.Setenv("MINT_USE_FIBER", "true")

	log.UILogger.Info("[main] Starting fiber_counter with TUI_DEBUG_ALL=true")
	log.UILogger.Info("[main] MINT_USE_FIBER=%s", os.Getenv("MINT_USE_FIBER"))

	err := ui.Run(SimpleCounter,
		ui.WithWidth(40),
		ui.WithHeight(10),
		ui.WithTitle("Fiber Counter (Debug)"),
	)

	if err != nil {
		log.UILogger.Error("[main] Error: %v", err)
		panic(err)
	}

	log.UILogger.Info("[main] Exiting normally")
}
