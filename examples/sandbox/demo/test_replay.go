package main

import (
	"fmt"
	"log"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox/real"
	"github.com/wwsheng009/mint/sandbox/replay"
)

// ReplayExample 演示事件回放功能
func ReplayExample() {
	fmt.Println("=== Event Replay Example ===")
	fmt.Println()

	// 1. 首先录制一些事件
	fmt.Println("Step 1: Recording events...")
	events := recordSampleEvents()
	fmt.Printf("Recorded %d events\n", len(events))
	fmt.Println()

	// 2. 创建回放沙箱
	fmt.Println("Step 2: Creating replay sandbox...")
	replaySB := replay.New(events, 40, 18)
	replaySB.Initialize(nil)

	fmt.Println("Replay sandbox created with events")
	fmt.Println()

	// 4. 回放事件
	fmt.Println("Step 3: Replaying events...")
	replayAllEvents(replaySB)

	// 5. 单步调试示例
	fmt.Println("\nStep 4: Step-by-step debugging...")
	stepByStepReplay(replaySB)
}

// recordSampleEvents 录制示例事件
func recordSampleEvents() []platform.RawInput {
	// 创建临时沙箱来录制
	sb, err := real.New(40, 18)
	if err != nil {
		log.Fatal(err)
	}
	defer sb.Close()

	sb.Initialize(nil)
	sb.Run()

	// 模拟一些事件（实际中这是用户输入）
	// 这里我们直接构建事件

	events := []platform.RawInput{}

	// 模拟用户点击 "+" 按钮 3 次
	for i := 0; i < 3; i++ {
		// Tab 到 "+"
		events = append(events, buildTabEvent())
		events = append(events, buildTabEvent())

		// Enter 点击
		events = append(events, buildEnterEvent())
	}

	// 模拟用户输入名字
	events = append(events, buildTabEvent())
	events = append(events, buildTabEvent())
	events = append(events, buildTabEvent())

	// 输入 "Alice"
	name := "Alice"
	for _, r := range name {
		events = append(events, buildKeyEvent(r))
	}

	return events
}

// replayAllEvents 回放所有事件
func replayAllEvents(replaySB *replay.ReplaySandbox) {
	count := 0

	for {
		event, err := replaySB.Step()
		if err != nil {
			if count > 0 {
				fmt.Printf("\nReplayed %d events\n", count)
			}
			break
		}

		count++
		fmt.Printf("Event %d: Type=%d", count, event.Type)

		switch event.Type {
		case platform.InputKeyPress:
			if event.Special != 0 {
				fmt.Printf(", Special=%d", event.Special)
			} else {
				fmt.Printf(", Key='%c'", event.Key)
			}
		case platform.InputResize:
			fmt.Printf(", Size=%dx%d", event.Width, event.Height)
		}

		fmt.Println()

		// 在实际应用中，这里会将事件发送到 UI 引擎
		// 例如：ui.DispatchEvent(event)
	}

	if count > 0 {
		fmt.Printf("\nReplayed %d events\n", count)
	}
}

// stepByStepReplay 单步调试示例
func stepByStepReplay(replaySB *replay.ReplaySandbox) {
	fmt.Println("Step-by-step debugging...")
	fmt.Println("Press Enter to see next event...")

	for {
		// 获取当前事件但不移动索引
		event, err := replaySB.Step()
		if err != nil {
			break
		}

		fmt.Printf("Event: ")
		switch event.Type {
		case platform.InputKeyPress:
			fmt.Printf("KeyPress")
			if event.Special != 0 {
				fmt.Printf(" (Special=%d)", event.Special)
			} else {
				fmt.Printf(" (Key='%c')", event.Key)
			}
		case platform.InputResize:
			fmt.Printf("Resize %dx%d", event.Width, event.Height)
		default:
			fmt.Printf("Type=%d", event.Type)
		}

		// 等待用户确认
		var input string
		fmt.Scanln(&input)
	}

	fmt.Println("Debugging complete.")
}

// SpeedControlExample 变速回放示例
func SpeedControlExample() {
	fmt.Println("=== Speed Control Example ===")
	fmt.Println()

	events := recordSampleEvents()
	replaySB := replay.New(events, 40, 18)

	// 1. 正常速度
	fmt.Println("1. Playing at normal speed (1.0x)...")
	replaySB.SetSpeed(1.0)
	fmt.Printf("Speed set to %.1fx\n", replaySB.GetSpeed())
	playAllEvents(replaySB)

	// 2. 快速播放
	fmt.Println("\n2. Playing at 2x speed...")
	replaySB.SetSpeed(2.0)
	fmt.Printf("Speed set to %.1fx\n", replaySB.GetSpeed())
	playAllEvents(replaySB)

	// 3. 慢速播放
	fmt.Println("\n3. Playing at 0.5x speed...")
	replaySB.SetSpeed(0.5)
	fmt.Printf("Speed set to %.1fx\n", replaySB.GetSpeed())
	playAllEvents(replaySB)
}

// playAllEvents 播放所有事件
func playAllEvents(replaySB *replay.ReplaySandbox) {
	for {
		_, err := replaySB.Step()
		if err != nil {
			break
		}
		// 在实际实现中，这里会根据速度调整延迟
	}
}

// NavigationExample 导航示例
func NavigationExample() {
	fmt.Println("=== Navigation Example ===")
	fmt.Println()

	events := recordSampleEvents()
	replaySB := replay.New(events, 40, 18)

	fmt.Println("Step by step navigation...")
	for i := 0; i < 5; i++ {
		event, err := replaySB.Step()
		if err != nil {
			break
		}
		fmt.Printf("Event %d: Type=%d\n", i+1, event.Type)
	}

	// 回退
	fmt.Println("\nStepping back...")
	for i := 0; i < 2; i++ {
		event, err := replaySB.StepBack()
		if err != nil {
			break
		}
		fmt.Printf("Back to: Type=%d\n", event.Type)
	}
}

// SnapshotWithReplay 快照与回放结合
func SnapshotWithReplay() {
	fmt.Println("=== Snapshot & Replay Example ===")
	fmt.Println()

	fmt.Println("Note: Replay sandbox does not support snapshot operations directly.")
	fmt.Println("Snapshots are available in Mock and Real sandboxes.")
	fmt.Println()
	fmt.Println("For snapshot examples, please refer to test_snapshot.go")
}

// 辅助函数：构建按键事件
func buildKeyEvent(key rune) platform.RawInput {
	return platform.RawInput{
		Type:      platform.InputKeyPress,
		Key:       key,
		Timestamp: getCurrentTimestamp(),
	}
}

// 辅助函数：构建 Tab 事件
func buildTabEvent() platform.RawInput {
	return platform.RawInput{
		Type:      platform.InputKeyPress,
		Special:   platform.KeyTab,
		Timestamp: getCurrentTimestamp(),
	}
}

// 辅助函数：构建 Enter 事件
func buildEnterEvent() platform.RawInput {
	return platform.RawInput{
		Type:      platform.InputKeyPress,
		Special:   platform.KeyEnter,
		Timestamp: getCurrentTimestamp(),
	}
}

// 辅助函数：获取当前时间戳
func getCurrentTimestamp() time.Time {
	// 简化版本，实际中应使用 time.Now()
	return time.Unix(1234567890, 0)
}
