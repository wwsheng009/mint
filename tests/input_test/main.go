package main

import (
	"fmt"
	"log"
	"time"

	"github.com/wwsheng009/mint/devtools/standalone"
	"github.com/wwsheng009/mint/runtime/platform"
)

func main() {
	// 初始化 Logger
	logger, err := standalone.NewLogger(nil)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	fmt.Printf("[DevTools] Logger started: %s\n", logger.GetPath())
	fmt.Println("=== Input Reader Test ===")
	fmt.Println("Press keys (ESC to exit)...")
	fmt.Println()

	// 创建输入读取器
	reader, err := platform.NewInputReader()
	if err != nil {
		log.Fatalf("Failed to create input reader: %v", err)
	}

	// 创建输入通道
	inputs := make(chan platform.RawInput, 100)

	// 启动输入读取器
	if err := reader.Start(inputs); err != nil {
		log.Fatalf("Failed to start input reader: %v", err)
	}
	defer reader.Stop()

	logger.LogMessage("Input reader started")

	fmt.Println("Waiting for input...")
	fmt.Println("------------------------------")

	eventCount := 0
	timeout := time.After(60 * time.Second)

	for {
		select {
		case <-timeout:
			fmt.Println("\n[TIMEOUT] No input received for 60 seconds")
			logger.LogMessage("TIMEOUT: No input received")
			return

		case raw, ok := <-inputs:
			if !ok {
				fmt.Println("\n[INFO] Input channel closed")
				return
			}

			eventCount++

			// 打印详细信息
			fmt.Printf("\n[Event #%d]\n", eventCount)
			fmt.Printf("  Type:        %d\n", raw.Type)
			fmt.Printf("  Key:         %c (%d)\n", raw.Key, raw.Key)
			fmt.Printf("  Special:     %d\n", raw.Special)
			fmt.Printf("  Modifiers:   %d\n", raw.Modifiers)
			fmt.Printf("  MouseX:      %d\n", raw.MouseX)
			fmt.Printf("  MouseY:      %d\n", raw.MouseY)
			fmt.Printf("  MouseButton: %d\n", raw.MouseButton)
			fmt.Printf("  MouseAction: %d\n", raw.MouseAction)
			fmt.Printf("  Timestamp:   %v\n", raw.Timestamp)
			fmt.Println("------------------------------")

			// 记录到日志
			logger.LogCustom("input", "system", map[string]interface{}{
				"event_num":    eventCount,
				"type":         raw.Type,
				"key":          fmt.Sprintf("%c", raw.Key),
				"special":      raw.Special,
				"modifiers":    raw.Modifiers,
				"mouse_x":      raw.MouseX,
				"mouse_y":      raw.MouseY,
				"mouse_button": raw.MouseButton,
				"mouse_action": raw.MouseAction,
			})

			// ESC 键退出
			if raw.Special == platform.KeyEscape {
				fmt.Println("\n[ESC] Exiting...")
				logger.LogMessage("ESC pressed, exiting")
				return
			}
		}
	}
}
