package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox"
	"github.com/wwsheng009/mint/sandbox/real"
)

// RealSandboxExample 演示如何使用真实沙箱运行应用
func RealSandboxExample() {
	// 1. 创建真实沙箱
	sb, err := real.New(40, 18)
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}
	defer sb.Close()

	// 2. 获取事件通道，可以监听事件
	// 注意：这里需要与实际的 UI 引擎集成
	go func() {
		for event := range sb.Events() {
			// 在实际应用中，这里会将事件发送到 UI 引擎
			// 例如：ui.DispatchEvent(event)
			fmt.Printf("[Event] %+v\n", event)
		}
	}()

	// 3. 初始化并运行
	if err := sb.Initialize(nil); err != nil {
		log.Fatalf("Initialize failed: %v", err)
	}

	if err := sb.Run(); err != nil {
		log.Fatalf("Run failed: %v", err)
	}

	// 4. 等待一段时间后停止
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Sandbox running. Press Ctrl+C to exit...")
	fmt.Println("Type keys to see events being recorded...")

	// 5. 等待退出
	<-sigCh

	// 6. 获取录制的事件
	events := sb.RecordedEvents()
	fmt.Printf("\nRecorded %d events\n", len(events))

	// 7. 可以将事件保存到文件供回放使用
	if len(events) > 0 {
		saveEvents(events, "recorded_events.json")
	}
}

// saveEvents 保存事件到文件（简化版）
func saveEvents(events []platform.RawInput, filename string) {
	fmt.Printf("Events saved to %s (not actually implemented)\n", filename)
	// 实际实现中，应该使用 JSON 序列化
}

// ManualInteractionExample 手动交互示例
// 这个例子演示如何将现有应用与真实沙箱集成
func ManualInteractionExample() {
	fmt.Println("Running counter with real sandbox...")
	fmt.Println("You can interact with the application normally.")
	fmt.Println("All events will be recorded for later replay.")

	// 注意：实际运行需要与 ui.Run() 集成
	// 这里演示如何使用真实沙箱捕获事件

	sb, err := real.New(80, 24)
	if err != nil {
		log.Fatal(err)
	}
	defer sb.Close()

	sb.Initialize(nil)
	sb.Run()

	// 在这里，你可以：
	// 1. 将事件转发到 ui.Run() 的输入
	// 2. 或者将 ui.Run() 的输入重定向到沙箱
	// 3. 等待用户交互

	// 示例：等待一段时间
	fmt.Println("Recording events for 10 seconds...")
	time.Sleep(10 * time.Second)

	// 获取录制的事件
	events := sb.RecordedEvents()
	fmt.Printf("Captured %d user events\n", len(events))

	// 分析事件
	analyzeEvents(events)
}

// analyzeEvents 分析录制的事件
func analyzeEvents(events []platform.RawInput) {
	fmt.Println("\n=== Event Analysis ===")

	keyPressCount := 0
	mouseClickCount := 0
	resizeCount := 0

	for _, event := range events {
		switch event.Type {
		case platform.InputKeyPress:
			keyPressCount++
		case platform.InputMouse:
			if event.MouseAction == platform.MousePress {
				mouseClickCount++
			}
		case platform.InputResize:
			resizeCount++
		}
	}

	fmt.Printf("Key presses: %d\n", keyPressCount)
	fmt.Printf("Mouse clicks: %d\n", mouseClickCount)
	fmt.Printf("Window resizes: %d\n", resizeCount)
}

// WithSnapshotRecording 使用快照记录
func WithSnapshotRecording() {
	sb, err := real.New(80, 24)
	if err != nil {
		log.Fatal(err)
	}
	defer sb.Close()

	sb.Initialize(nil)
	sb.Run()

	fmt.Println("Creating snapshots during interaction...")

	// 创建初始快照
	snap1, _ := sb.Snapshot(sandbox.SnapshotStandard, "initial")
	fmt.Printf("Created snapshot: %s\n", snap1.Metadata.ID)

	// 模拟一些交互（实际中这是用户输入）
	fmt.Println("Press Enter to create another snapshot...")
	// 等待用户输入...

	// 创建第二个快照
	snap2, _ := sb.Snapshot(sandbox.SnapshotStandard, "after-interaction")
	fmt.Printf("Created snapshot: %s\n", snap2.Metadata.ID)

	// 列出所有快照
	snapshots := sb.ListSnapshots()
	fmt.Printf("\nTotal snapshots: %d\n", len(snapshots))
	for _, meta := range snapshots {
		fmt.Printf("- ID: %s, Tags: %v\n", meta.ID, meta.Tags)
	}
}

// InteractiveRecording 交互式录制示例
func InteractiveRecording() {
	sb, err := real.New(80, 24)
	if err != nil {
		log.Fatal(err)
	}
	defer sb.Close()

	sb.Initialize(nil)
	sb.Run()

	fmt.Println("=== Interactive Recording ===")
	fmt.Println("Interact with your application...")
	fmt.Println("Events will be recorded.")
	fmt.Println("Press 'q' to stop recording.")

	// 监控事件（简化版）
	eventCount := 0

	// 监控事件通道（简化版）
	go func() {
		for event := range sb.Events() {
			eventCount++

			// 检查是否按下 'q'
			if event.Type == platform.InputKeyPress && event.Key == 'q' {
				fmt.Printf("\nStopping recording after %d events.\n", eventCount)
				// 这里应该有停止逻辑
			}
		}
	}()

	// 在实际应用中，这里会将事件发送到 UI 引擎
	// ui.Run(app, ui.WithInputSource(sb.Events()))

	// 等待用户完成
	fmt.Println("Recording session complete.")
}
