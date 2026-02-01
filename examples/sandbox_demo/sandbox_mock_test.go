package main

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox"
	"github.com/wwsheng009/mint/ui"
)

// TestCounterWithMock 使用 Mock 沙箱测试计数器应用
func TestCounterWithMock(t *testing.T) {
	// 1. 创建测试应用
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	// 2. 获取测试辅助器
	helper := testApp.Helper()

	// 3. 测试事件注入
	t.Run("EventInjection", func(t *testing.T) {
		result := helper.
			Type("hello").
			Tab().
			Enter().
			Process().
			Result()

		if !result.OK() {
			t.Errorf("Event injection failed: %v", result.Error())
		}
	})

	// 4. 测试增加计数
	t.Run("IncrementCount", func(t *testing.T) {
		// "+" 按钮是第二个按钮（索引 1）
		if err := testApp.ClickButton(1); err != nil {
			t.Errorf("ClickButton failed: %v", err)
		}

		// 检查渲染
		if err := testApp.Sandbox().AssertRender("Count: 1"); err != nil {
			t.Errorf("Increment failed: %v", err)
		}
	})

	// 5. 测试减少计数
	t.Run("DecrementCount", func(t *testing.T) {
		// "-" 按钮是第一个按钮（索引 0）
		if err := testApp.ClickButton(0); err != nil {
			t.Errorf("ClickButton failed: %v", err)
		}

		// 检查渲染
		if err := testApp.Sandbox().AssertRender("Count: -1"); err != nil {
			t.Errorf("Decrement failed: %v", err)
		}
	})

	// 6. 测试连续增加
	t.Run("MultipleIncrements", func(t *testing.T) {
		// 连续点击 5 次 "+"
		for i := 0; i < 5; i++ {
			if err := testApp.ClickButton(1); err != nil {
				t.Errorf("ClickButton failed: %v", err)
			}

			// 检查渲染
			expected := fmt.Sprintf("Count: %d", i+1)
			if err := testApp.Sandbox().AssertRender(expected); err != nil {
				t.Errorf("Increment %d failed: %v", i+1, err)
			}
		}

		// 最终验证
		if err := testApp.Sandbox().AssertRender("Count: 5"); err != nil {
			t.Errorf("Final count check failed: %v", err)
		}
	})
}

// TestCounterWithCustomConfig 使用自定义配置测试
func TestCounterWithCustomConfig(t *testing.T) {
	// 1. 创建自定义配置
	config := sandbox.DefaultConfig()
	config.Width = 50
	config.Height = 20
	config.Event.QueueMaxMemory = 50 * 1024 * 1024  // 50MB

	// 2. 使用自定义配置创建测试应用
	testApp, err := ui.TestRunWithConfig(Counter, config)
	if err != nil {
		t.Fatalf("TestRunWithConfig failed: %v", err)
	}
	defer testApp.Close()

	// 3. 获取沙箱并检查尺寸
	sb := testApp.Sandbox()
	width, height := sb.Size()

	if width != 50 || height != 20 {
		t.Errorf("Size mismatch: got %dx%d, want 50x20", width, height)
	}

	// 4. 验证应用正常运行
	helper := testApp.Helper()
	result := helper.
		Process().
		AssertRender("Sandbox Demo: Counter").
		Result()

	if !result.OK() {
		t.Errorf("App check failed: %v", result.Error())
	}
}

// TestDirectSandboxAccess 直接访问沙箱进行更精细的控制
func TestDirectSandboxAccess(t *testing.T) {
	// 1. 创建测试应用
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	// 2. 获取底层 Mock 沙箱
	sb := testApp.Sandbox()

	// 3. 获取渲染输出
	output := sb.RenderString()
	t.Logf("Initial output:\n%s", output)

	// 4. 检查初始状态
	if err := sb.AssertRender("Count: 0"); err != nil {
		t.Errorf("Initial state check failed: %v", err)
	}

	// 5. 直接注入事件（不使用辅助器）
	// Tab 到 "+" 按钮并点击
	sb.InjectSpecialKey(platform.KeyTab)
	sb.InjectSpecialKey(platform.KeyTab)
	sb.InjectSpecialKey(platform.KeyEnter)
	sb.ProcessEvents()

	// 6. 验证结果
	if err := sb.AssertRender("Count: 1"); err != nil {
		t.Errorf("Increment check failed: %v", err)
	}

	// 7. 检查队列统计
	stats := sb.QueueStats()
	t.Logf("Queue stats - Length: %d, Memory: %d bytes",
		stats.Length, stats.MemoryUsed)
}

// TestInputField 测试输入框功能
func TestInputField(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	helper := testApp.Helper()

	// 1. 导航到输入框（3 次 Tab）
	result := helper.
		Tab().Tab().Tab().
		Process().
		Result()

	if !result.OK() {
		t.Fatalf("Navigation failed: %v", result.Error())
	}

	// 2. 输入名字
	result = helper.
		Type("Alice").
		Process().
		AssertRender("Hello, Alice").
		Result()

	if !result.OK() {
		t.Errorf("Input failed: %v", result.Error())
	}
}

// TestMultipleActions 测试多个操作的组合
func TestMultipleActions(t *testing.T) {
	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 18))
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	helper := testApp.Helper()

	// 1. 完整的用户流程
	result := helper.
		// 修改名字
		Tab().Tab().Tab().
		Type("Bob").
		Process().
		AssertRender("Hello, Bob").

		// 增加计数 3 次
		Tab().Tab().Tab().Enter().  // 导航到 "+"
		Process().
		AssertRender("Count: 1").

		Tab().Tab().Tab().Enter().
		Process().
		AssertRender("Count: 2").

		Tab().Tab().Tab().Enter().
		Process().
		AssertRender("Count: 3").

		// 减少计数 1 次
		Tab().Tab().Tab().Tab().  // 导航到 "-"
		Enter().
		Process().
		AssertRender("Count: 2").
		Result()

	if !result.OK() {
		t.Errorf("Multiple actions failed: %v", result.Error())
	}
}
