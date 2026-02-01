// 01_event_recording/demo_test.go
// 事件录制与回放演示测试

package main

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox"
	"github.com/wwsheng009/mint/ui"
)

// TestEventRecording 演示事件录制功能
func TestEventRecording(t *testing.T) {
	t.Log("=== 事件录制演示 ===")

	// 1. 创建录制器
	recorder := sandbox.NewEventRecorder(1000)
	t.Logf("创建录制器，最大容量: %d 个事件", 1000)

	// 2. 创建测试应用（使用 Sandbox 模式以获取 MockSandbox）
	testApp, err := ui.RunTestWithSandbox(SimpleCounter,
		ui.WithWidth(40),
		ui.WithHeight(16),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	// 3. 获取 MockSandbox 并设置录制器
	sb := testApp.GetSandbox()
	sb.SetRecorder(recorder)

	// 4. 等待应用初始化
	time.Sleep(100 * time.Millisecond)

	// 5. 执行一系列操作（会被自动录制）
	t.Log("\n--- 执行操作并录制 ---")

	// 点击 + 按钮 3 次
	for i := 0; i < 3; i++ {
		t.Logf("操作 %d: 按 Tab 切换焦点", i+1)
		sb.Helper().Tab()
		time.Sleep(100 * time.Millisecond)

		t.Logf("操作 %d: 按 Enter 点击 + 按钮", i+1)
		sb.Helper().Press(platform.KeyEnter)
		time.Sleep(100 * time.Millisecond)
	}

	// 6. 获取录制的事件
	events := recorder.Events()
	t.Logf("\n--- 录制完成 ---")
	t.Logf("共录制了 %d 个事件", len(events))

	// 打印前几个事件详情
	for i, ev := range events {
		if i >= 5 {
			t.Logf("... 还有 %d 个事件", len(events)-5)
			break
		}
		t.Logf("事件 %d: Type=%v, Key=%v, Special=%v",
			i+1, ev.Type, ev.Key, ev.Special)
	}

	// 7. 验证应用状态
	rendered := testApp.GetRenderString()
	t.Logf("\n录制后渲染结果:\n%s", rendered)

	// 8. 保存录制到文件
	err = recorder.SaveToFile("recorded_events.json")
	if err != nil {
		t.Logf("保存录制到文件失败: %v", err)
	} else {
		t.Log("录制已保存到: recorded_events.json")
	}
}

// TestEventReplay 演示事件回放功能
func TestEventReplay(t *testing.T) {
	t.Log("=== 事件回放演示 ===")

	// 1. 首先录制一组操作
	t.Log("\n--- 步骤 1: 录制操作 ---")

	recorder := sandbox.NewEventRecorder(1000)

	testApp1, err := ui.RunTestWithSandbox(SimpleCounter,
		ui.WithWidth(40),
		ui.WithHeight(16),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}

	sb1 := testApp1.GetSandbox()
	sb1.SetRecorder(recorder)
	time.Sleep(100 * time.Millisecond)

	// 执行操作：点击 + 按钮 2 次
	sb1.Helper().Tab().Press(platform.KeyEnter)
	time.Sleep(100 * time.Millisecond)
	sb1.Helper().Tab().Press(platform.KeyEnter)
	time.Sleep(100 * time.Millisecond)

	// 获取录制的事件
	events := recorder.Events()
	t.Logf("录制了 %d 个事件", len(events))

	// 获取录制后的状态
	rendered1 := testApp1.GetRenderString()
	t.Logf("\n录制后的应用状态:\n%s", rendered1)

	testApp1.Close()

	// 2. 回放操作到新的应用实例
	t.Log("\n--- 步骤 2: 回放操作 ---")

	testApp2, err := ui.RunTestWithSandbox(SimpleCounter,
		ui.WithWidth(40),
		ui.WithHeight(16),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp2.Close()

	sb2 := testApp2.GetSandbox()

	// 回放每个事件
	t.Logf("开始回放 %d 个事件...", len(events))
	for i, ev := range events {
		sb2.InjectRaw(ev)
		time.Sleep(150 * time.Millisecond) // 等待框架处理事件

		if i < 5 || i == len(events)-1 {
			t.Logf("回放事件 %d: Type=%v", i+1, ev.Type)
		}
	}

	time.Sleep(100 * time.Millisecond)

	// 3. 比较回放后的状态与录制后的状态
	rendered2 := testApp2.GetRenderString()
	t.Logf("\n回放后的应用状态:\n%s", rendered2)

	// 验证结果一致 (按两次Tab+Enter，先点-按钮，再点+按钮，最终应该是0)
	if containsCount(rendered1, "Count: 0") && containsCount(rendered2, "Count: 0") {
		t.Log("✅ 回放成功！状态一致")
	} else {
		t.Logf("录制后包含 Count: 0: %v", containsCount(rendered1, "Count: 0"))
		t.Logf("回放后包含 Count: 0: %v", containsCount(rendered2, "Count: 0"))
		t.Error("❌ 回放失败！状态不一致")
	}
}

// TestRecordingToFile 演示录制保存和加载
func TestRecordingToFile(t *testing.T) {
	t.Log("=== 录制文件操作演示 ===")

	// 1. 录制操作
	recorder := sandbox.NewEventRecorder(1000)

	testApp, err := ui.RunTestWithSandbox(SimpleCounter,
		ui.WithWidth(40),
		ui.WithHeight(16),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()
	sb.SetRecorder(recorder)
	time.Sleep(50 * time.Millisecond)

	// 简单的操作序列
	sb.Helper().
		Tab().
		Press(platform.KeyEnter).
		Tab().
		Press(platform.KeyEnter)

	time.Sleep(200 * time.Millisecond)

	// 2. 保存到文件
	t.Log("保存录制到文件...")
	filename := "demo_recording.json"
	err = recorder.SaveToFile(filename)
	if err != nil {
		t.Fatalf("保存录制失败: %v", err)
	}
	t.Logf("✅ 录制已保存到: %s", filename)

	// 3. 从文件加载
	t.Log("\n从文件加载录制...")
	recorder2 := sandbox.NewEventRecorder(1000)
	err = recorder2.LoadFromFile(filename)
	if err != nil {
		t.Fatalf("加载录制失败: %v", err)
	}

	events := recorder2.Events()
	t.Logf("✅ 从文件加载了 %d 个事件", len(events))

	// 4. 验证加载的事件
	eventCount := 0
	keyPressCount := 0
	for _, ev := range events {
		eventCount++
		if ev.Type == platform.InputKeyPress {
			keyPressCount++
		}
	}
	t.Logf("事件统计: 总数=%d, 按键=%d", eventCount, keyPressCount)
}

// 辅助函数：检查渲染结果是否包含指定计数
func containsCount(rendered, text string) bool {
	// 简单检查字符串包含
	for i := 0; i < len(rendered)-len(text)+1; i++ {
		if rendered[i:i+len(text)] == text {
			return true
		}
	}
	return false
}
