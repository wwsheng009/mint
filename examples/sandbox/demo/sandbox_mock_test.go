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
		// 先重置焦点（Escape）
		if err := testApp.InjectSpecialKey(platform.KeyEscape); err != nil {
			t.Errorf("Failed to inject Escape: %v", err)
		}
		time.Sleep(50 * time.Millisecond)

		// 使用 Tab 导航到 "+" 按钮并按 Enter 点击
		// 第一个 Tab: 聚焦到 [-] 按钮
		if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
			t.Errorf("Failed to inject Tab: %v", err)
		}
		time.Sleep(50 * time.Millisecond)

		// 第二个 Tab: 聚焦到 [+] 按钮
		if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
			t.Errorf("Failed to inject Tab: %v", err)
		}
		time.Sleep(50 * time.Millisecond)

		// 按 Enter 点击
		if err := testApp.InjectSpecialKey(platform.KeyEnter); err != nil {
			t.Errorf("Failed to inject Enter: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
		testApp.GetFrameworkApp().ForceRenderNow()
		time.Sleep(50 * time.Millisecond)

		// 检查渲染
		rendered := testApp.GetRenderString()
		t.Logf("After increment:\n%s", rendered)

		if err := testApp.AssertRender("Count: 1"); err != nil {
			t.Errorf("Increment failed: %v", err)
		}
	})

	// 测试减少计数
	t.Run("DecrementCount", func(t *testing.T) {
		// 导航到 "-" 按钮并点击
		if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
			t.Errorf("Failed to inject Tab: %v", err)
		}
		time.Sleep(50 * time.Millisecond)

		if err := testApp.InjectSpecialKey(platform.KeyEnter); err != nil {
			t.Errorf("Failed to inject Enter: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
		testApp.GetFrameworkApp().ForceRenderNow()
		time.Sleep(50 * time.Millisecond)

		// 检查渲染
		rendered := testApp.GetRenderString()
		t.Logf("After decrement:\n%s", rendered)

		if err := testApp.AssertRender("Count: -1"); err != nil {
			t.Errorf("Decrement failed: %v", err)
		}
	})

	// 测试连续增加
	t.Run("MultipleIncrements", func(t *testing.T) {
		// 连续点击 5 次 "+"
		for i := 0; i < 5; i++ {
			// Tab 到 "+" 按钮（需要 2 次 Tab：- -> +）
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
			time.Sleep(100 * time.Millisecond)
			testApp.GetFrameworkApp().ForceRenderNow()
			time.Sleep(50 * time.Millisecond)

			// 检查渲染
			rendered := testApp.GetRenderString()
			t.Logf("After increment %d:\n%s", i+1, rendered)

			// 注意：由于我们从 -1 开始，所以第一次 + 后是 0
			if err := testApp.AssertRender("Count: 2"); err != nil && i == 3 {
				t.Errorf("Increment %d failed: %v", i+1, err)
			}
		}

		// 最终验证
		if err := testApp.AssertRender("Count: 4"); err != nil {
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

	// 导航到输入框（需要多次 Tab）
	for i := 0; i < 4; i++ {
		if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
			t.Errorf("Failed to inject Tab %d: %v", i, err)
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
	time.Sleep(100 * time.Millisecond)
	testApp.GetFrameworkApp().ForceRenderNow()
	time.Sleep(50 * time.Millisecond)

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

	t.Logf("Buttons count: %d", len(root.GetButtons()))
	t.Logf("Inputs count: %d", len(root.GetInputs()))
	t.Logf("Focused index: %d, type: %d", root.GetFocusedIndex(), root.GetFocusedType())

	// 检查按钮存在
	buttons := root.GetButtons()
	if len(buttons) != 2 {
		t.Errorf("Expected 2 buttons, got %d", len(buttons))
	}

	// 检查输入框存在
	inputs := root.GetInputs()
	if len(inputs) != 1 {
		t.Errorf("Expected 1 input, got %d", len(inputs))
	}
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

	root := testApp.GetDeclarativeRoot()
	buttons := root.GetButtons()

	if len(buttons) == 0 {
		t.Fatal("No buttons found")
	}

	t.Logf("Found %d buttons", len(buttons))

	// 获取第二个按钮（"+" 按钮）的边界
	if len(buttons) > 1 {
		bounds := buttons[1].Bounds()
		t.Logf("Button [+] bounds: x=%d, y=%d, w=%d, h=%d",
			bounds[0], bounds[1], bounds[2], bounds[3])

		// 点击按钮中心
		clickX := bounds[0] + bounds[2]/2
		clickY := bounds[1] + bounds[3]/2
		t.Logf("Clicking at x=%d, y=%d", clickX, clickY)

		if err := testApp.InjectMouse(clickX, clickY, platform.MouseLeft, platform.MousePress); err != nil {
			t.Errorf("Failed to inject mouse click: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
		testApp.GetFrameworkApp().ForceRenderNow()
		time.Sleep(50 * time.Millisecond)

		// 检查渲染
		rendered := testApp.GetRenderString()
		t.Logf("After mouse click:\n%s", rendered)

		if err := testApp.AssertRender("Count: 1"); err != nil {
			t.Errorf("Counter not incremented after mouse click: %v", err)
		}
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

	// 1. 修改名字
	t.Log("=== Step 1: Change name ===")
	for i := 0; i < 4; i++ {
		if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
			t.Errorf("Failed to inject Tab: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	name := "Bob"
	for _, ch := range name {
		if err := testApp.InjectKey(ch); err != nil {
			t.Errorf("Failed to inject key '%c': %v", ch, err)
		}
		time.Sleep(50 * time.Millisecond)
	}
	time.Sleep(100 * time.Millisecond)
	testApp.GetFrameworkApp().ForceRenderNow()
	time.Sleep(50 * time.Millisecond)

	if err := testApp.AssertRender("Hello, Bob"); err != nil {
		t.Errorf("Name change failed: %v", err)
	}

	// 2. 增加计数
	t.Log("=== Step 2: Increment count ===")
	for i := 0; i < 2; i++ {
		if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
			t.Errorf("Failed to inject Tab: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	if err := testApp.InjectSpecialKey(platform.KeyEnter); err != nil {
		t.Errorf("Failed to inject Enter: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	testApp.GetFrameworkApp().ForceRenderNow()
	time.Sleep(50 * time.Millisecond)

	if err := testApp.AssertRender("Count: 1"); err != nil {
		t.Errorf("Increment failed: %v", err)
	}

	// 3. 减少计数
	t.Log("=== Step 3: Decrement count ===")
	if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
		t.Errorf("Failed to inject Tab: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	if err := testApp.InjectSpecialKey(platform.KeyEnter); err != nil {
		t.Errorf("Failed to inject Enter: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	testApp.GetFrameworkApp().ForceRenderNow()
	time.Sleep(50 * time.Millisecond)

	if err := testApp.AssertRender("Count: 0"); err != nil {
		t.Errorf("Decrement failed: %v", err)
	}

	rendered := testApp.GetRenderString()
	t.Logf("Final render:\n%s", rendered)
}
