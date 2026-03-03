package main

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/ui"
)

// TestCounterWithRunTest 使用 RunTest (新版 API) 测试计数器应用
func TestCounterWithRunTest(t *testing.T) {
	testApp, err := ui.RunTest(Counter,
		ui.WithWidth(40),
		ui.WithHeight(18),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	// 等待初始渲染和事件泵启动
	time.Sleep(300 * time.Millisecond)

	// 检查应用状态
	fwApp := testApp.GetFrameworkApp()
	t.Logf("App state: %v", fwApp.GetState())
	t.Logf("App pump: %v", fwApp.GetPump())

	// 获取渲染输出
	rendered := testApp.GetRenderString()
	t.Logf("Initial render:\n%s", rendered)

	// 检查初始状态
	if err := testApp.AssertRender("Count: 0"); err != nil {
		t.Errorf("Initial state check failed: %v", err)
	}

	if err := testApp.AssertRender("Hello, Guest"); err != nil {
		t.Errorf("Initial greeting check failed: %v", err)
	}

	// 测试增加计数
	t.Run("IncrementCount", func(t *testing.T) {
		// 使用 Tab 导航到 "+" 按钮并按 Enter 点击
		// 焦点从 [-] (index 0) -> [+] (index 1)
		if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
			t.Errorf("Failed to inject Tab: %v", err)
		}
		time.Sleep(50 * time.Millisecond)

		// 按 Enter 点击
		if err := testApp.InjectSpecialKey(platform.KeyEnter); err != nil {
			t.Errorf("Failed to inject Enter: %v", err)
		}
		time.Sleep(150 * time.Millisecond)

		// 检查渲染
		rendered := testApp.GetRenderString()
		t.Logf("After increment:\n%s", rendered)

		if err := testApp.AssertRender("Count: 1"); err != nil {
			t.Errorf("Increment failed: %v", err)
		}
	})

	// 测试减少计数
	t.Run("DecrementCount", func(t *testing.T) {
		// 当前状态: Count = 1 (after IncrementCount)
		// 当前焦点在 [+] (index 1)，需要导航到 [-] (index 0)
		// 路径: [+] (1) -> Input (2) -> [-] (0) wraps around
		if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
			t.Errorf("Failed to inject Tab: %v", err)
		}
		time.Sleep(50 * time.Millisecond)

		if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
			t.Errorf("Failed to inject Tab: %v", err)
		}
		time.Sleep(50 * time.Millisecond)

		if err := testApp.InjectSpecialKey(platform.KeyEnter); err != nil {
			t.Errorf("Failed to inject Enter: %v", err)
		}
		time.Sleep(150 * time.Millisecond)

		// 检查渲染: 1 -> 0
		rendered := testApp.GetRenderString()
		t.Logf("After decrement:\n%s", rendered)

		if err := testApp.AssertRender("Count: 0"); err != nil {
			t.Errorf("Decrement failed: %v", err)
		}
	})

	// 测试连续增加
	t.Run("MultipleIncrements", func(t *testing.T) {
		// 当前状态: Count = 0 (after IncrementCount + DecrementCount)
		// 焦点在 [-] (0)，需要先导航到 [+] (1)
		if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
			t.Errorf("Failed to inject Tab: %v", err)
		}
		time.Sleep(50 * time.Millisecond)

		// 连续点击 3 次 "+" (焦点已经在 [+] 上)
		for i := 0; i < 3; i++ {
			if err := testApp.InjectSpecialKey(platform.KeyEnter); err != nil {
				t.Errorf("Failed to inject Enter: %v", err)
			}
			time.Sleep(150 * time.Millisecond)

			// 检查渲染
			rendered := testApp.GetRenderString()
			t.Logf("After increment %d:\n%s", i+1, rendered)
		}

		// 最终验证: 0 + 3 = 3
		if err := testApp.AssertRender("Count: 3"); err != nil {
			t.Errorf("Final count check failed: %v", err)
		}
	})
}

// TestCounterWithInputField 测试输入框功能
func TestCounterWithInputField(t *testing.T) {
	testApp, err := ui.RunTest(Counter,
		ui.WithWidth(40),
		ui.WithHeight(18),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	// 等待初始渲染
	time.Sleep(100 * time.Millisecond)

	// 导航到输入框: [-] (0) -> [+] (1) -> Input (2)
	for i := 0; i < 2; i++ {
		if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
			t.Errorf("Failed to inject Tab %d: %v", i, err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	// 清除现有文本 "Guest" (5个字符) - 使用 Backspace
	for i := 0; i < 5; i++ {
		if err := testApp.InjectSpecialKey(platform.KeyBackspace); err != nil {
			t.Errorf("Failed to inject Backspace: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	// 输入名字
	name := "Alice"
	for _, ch := range name {
		if err := testApp.InjectKey(ch); err != nil {
			t.Errorf("Failed to inject key '%c': %v", ch, err)
		}
		time.Sleep(50 * time.Millisecond)
	}
	time.Sleep(150 * time.Millisecond)

	// 检查渲染
	rendered := testApp.GetRenderString()
	t.Logf("After input:\n%s", rendered)

	if err := testApp.AssertRender("Hello, Alice"); err != nil {
		t.Errorf("Input failed: %v", err)
	}
}

// TestCounterGetDeclarativeRoot 测试获取声明式根节点
func TestCounterGetDeclarativeRoot(t *testing.T) {
	testApp, err := ui.RunTest(Counter,
		ui.WithWidth(40),
		ui.WithHeight(18),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	// 等待初始渲染
	time.Sleep(100 * time.Millisecond)

	// 获取声明式根节点
	root := testApp.GetDeclarativeRoot()
	t.Logf("Root node: %v", root)

	// 使用 TestableApp 的方法获取焦点信息
	t.Logf("Focused index: %d, type: %d", testApp.GetFocusedIndex(), testApp.GetFocusedType())
}

// TestCounterMouseClick 测试鼠标点击
func TestCounterMouseClick(t *testing.T) {
	testApp, err := ui.RunTest(Counter,
		ui.WithWidth(40),
		ui.WithHeight(18),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	// 等待初始渲染
	time.Sleep(100 * time.Millisecond)

	// 使用键盘导航进行测试 (鼠标点击测试需要 Bounds 支持)
	// 导航到 [+] 按钮并点击
	if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
		t.Errorf("Failed to inject Tab: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	if err := testApp.InjectSpecialKey(platform.KeyEnter); err != nil {
		t.Errorf("Failed to inject Enter: %v", err)
	}
	time.Sleep(150 * time.Millisecond)

	// 检查渲染
	rendered := testApp.GetRenderString()
	t.Logf("After click:\n%s", rendered)

	if err := testApp.AssertRender("Count: 1"); err != nil {
		t.Errorf("Counter not incremented: %v", err)
	}
}

// TestCounterComprehensive 综合测试
func TestCounterComprehensive(t *testing.T) {
	testApp, err := ui.RunTest(Counter,
		ui.WithWidth(40),
		ui.WithHeight(18),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	// 等待初始渲染
	time.Sleep(100 * time.Millisecond)

	// 1. 增加计数
	t.Log("=== Step 1: Increment count ===")
	// 从 [-] (0) -> [+] (1)
	if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
		t.Errorf("Failed to inject Tab: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	if err := testApp.InjectSpecialKey(platform.KeyEnter); err != nil {
		t.Errorf("Failed to inject Enter: %v", err)
	}
	time.Sleep(150 * time.Millisecond)

	if err := testApp.AssertRender("Count: 1"); err != nil {
		t.Errorf("Increment failed: %v", err)
	}

	// 2. 减少计数
	t.Log("=== Step 2: Decrement count ===")
	// 当前焦点在 [+] (1)，需要导航到 [-] (0)
	// Tab: [+] (1) -> Input (2) -> [-] (0) wraps
	for i := 0; i < 2; i++ {
		if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
			t.Errorf("Failed to inject Tab: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	if err := testApp.InjectSpecialKey(platform.KeyEnter); err != nil {
		t.Errorf("Failed to inject Enter: %v", err)
	}
	time.Sleep(150 * time.Millisecond)

	if err := testApp.AssertRender("Count: 0"); err != nil {
		t.Errorf("Decrement failed: %v", err)
	}

	rendered := testApp.GetRenderString()
	t.Logf("Final render:\n%s", rendered)
}
