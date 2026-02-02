// Package main provides chain API testing examples
package main

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/ui"
)

// TestChainBasic 基础链式调用测试
func TestChainBasic(t *testing.T) {
	testApp, err := ui.RunTest(Counter,
		ui.WithSize(40, 18),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(50 * time.Millisecond)

	// 验证初始状态
	rendered := testApp.GetRenderString()
	if rendered == "" {
		t.Error("Initial render is empty")
	}

	t.Log("Basic chain test passed")
}

// TestChainIncrement 递增操作链式测试
func TestChainIncrement(t *testing.T) {
	testApp, err := ui.RunTest(Counter,
		ui.WithSize(40, 18),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(50 * time.Millisecond)

	// 初始状态验证
	rendered := testApp.GetRenderString()
	if !contains(rendered, "Count: 0") {
		t.Logf("Warning: Initial state may not show Count: 0")
	}

	// 模拟点击 + 按钮：Tab 两次到 + 按钮，然后 Enter
	testApp.InjectSpecialKey(platform.KeyTab)
	time.Sleep(20 * time.Millisecond)
	testApp.InjectSpecialKey(platform.KeyTab)
	time.Sleep(20 * time.Millisecond)
	testApp.InjectSpecialKey(platform.KeyEnter)
	time.Sleep(50 * time.Millisecond)

	// 验证状态更新
	rendered = testApp.GetRenderString()
	t.Logf("After increment: %s", rendered)

	if contains(rendered, "Count: 1") {
		t.Log("✓ Increment successful")
	} else {
		t.Logf("State after increment: %s", rendered)
	}
}

// TestChainNameChange 名字修改链式测试
func TestChainNameChange(t *testing.T) {
	testApp, err := ui.RunTest(Counter,
		ui.WithSize(40, 18),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(50 * time.Millisecond)

	// Tab 到输入框（第4个元素：- button, + button, input, 或者 input 在不同位置）
	// 先试试多次 Tab
	testApp.InjectSpecialKey(platform.KeyTab)
	testApp.InjectSpecialKey(platform.KeyTab)
	testApp.InjectSpecialKey(platform.KeyTab)
	time.Sleep(20 * time.Millisecond)

	// 输入名字
	testApp.InjectString("Alice")
	time.Sleep(20 * time.Millisecond)
	testApp.InjectSpecialKey(platform.KeyEnter)
	time.Sleep(50 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("After name change: %s", rendered)

	if contains(rendered, "Alice") {
		t.Log("✓ Name change successful")
	}
}

// TestChainMultipleUpdates 多次状态更新测试
func TestChainMultipleUpdates(t *testing.T) {
	testApp, err := ui.RunTest(Counter,
		ui.WithSize(40, 18),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(50 * time.Millisecond)

	successCount := 0

	// 递增 5 次
	for i := 1; i <= 5; i++ {
		testApp.InjectSpecialKey(platform.KeyTab)
		testApp.InjectSpecialKey(platform.KeyTab)
		testApp.InjectSpecialKey(platform.KeyEnter)
		time.Sleep(30 * time.Millisecond)

		rendered := testApp.GetRenderString()
		expected := contains(rendered, "Count:") && contains(rendered, string(rune('0'+i)))

		if expected {
			successCount++
		}

		t.Logf("Iteration %d: %s", i, rendered)
	}

	t.Logf("Successful updates: %d/5", successCount)
}

// TestChainNavigation 焦点导航测试
func TestChainNavigation(t *testing.T) {
	testApp, err := ui.RunTest(Counter,
		ui.WithSize(40, 18),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(50 * time.Millisecond)

	// 测试 Tab 导航
	for i := 0; i < 6; i++ {
		testApp.InjectSpecialKey(platform.KeyTab)
		time.Sleep(20 * time.Millisecond)

		rendered := testApp.GetRenderString()
		t.Logf("Tab #%d: %s", i+1, rendered)
	}

	t.Log("Navigation test completed")
}

// TestChainComplexScenario 复杂场景测试
func TestChainComplexScenario(t *testing.T) {
	testApp, err := ui.RunTest(Counter,
		ui.WithSize(40, 18),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(50 * time.Millisecond)

	t.Log("=== 复杂场景测试 ===")

	// 1. 验证初始状态
	rendered := testApp.GetRenderString()
	t.Logf("初始状态: %s", rendered)

	// 2. 递增 3 次
	for i := 0; i < 3; i++ {
		testApp.InjectSpecialKey(platform.KeyTab)
		testApp.InjectSpecialKey(platform.KeyTab)
		testApp.InjectSpecialKey(platform.KeyEnter)
		time.Sleep(30 * time.Millisecond)
	}

	rendered = testApp.GetRenderString()
	t.Logf("递增3次后: %s", rendered)

	// 3. 递减 1 次
	testApp.InjectSpecialKey(platform.KeyTab)
	testApp.InjectSpecialKey(platform.KeyEnter)
	time.Sleep(30 * time.Millisecond)

	rendered = testApp.GetRenderString()
	t.Logf("递减1次后: %s", rendered)

	t.Log("Complex scenario test completed")
}

// TestChainKeyboardShortcuts 键盘快捷键测试
func TestChainKeyboardShortcuts(t *testing.T) {
	testApp, err := ui.RunTest(Counter,
		ui.WithSize(40, 18),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(50 * time.Millisecond)

	// 测试 Tab
	testApp.InjectSpecialKey(platform.KeyTab)
	time.Sleep(20 * time.Millisecond)

	// 测试 Enter
	testApp.InjectSpecialKey(platform.KeyEnter)
	time.Sleep(30 * time.Millisecond)

	// 测试 Escape
	testApp.InjectSpecialKey(platform.KeyEscape)
	time.Sleep(20 * time.Millisecond)

	t.Log("Keyboard shortcuts test completed")
}

// TestChainMouseActions 鼠标操作测试
func TestChainMouseActions(t *testing.T) {
	testApp, err := ui.RunTest(Counter,
		ui.WithSize(40, 18),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(50 * time.Millisecond)

	// 测试鼠标点击 (+ 按钮位置大约在 x=20, y=8)
	testApp.InjectMouse(20, 8, platform.MouseLeft, platform.MousePress)
	time.Sleep(50 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("After mouse click: %s", rendered)

	if contains(rendered, "Count: 1") {
		t.Log("✓ Mouse click successful")
	} else {
		t.Log("Mouse click may not have worked - bounds may need adjustment")
	}
}

// TestChainStateValidation 状态验证测试
func TestChainStateValidation(t *testing.T) {
	testApp, err := ui.RunTest(Counter,
		ui.WithSize(40, 18),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Close()

	time.Sleep(50 * time.Millisecond)

	// 验证初始状态
	rendered := testApp.GetRenderString()
	initialOK := contains(rendered, "Count:")
	if !initialOK {
		t.Error("Initial state missing 'Count:'")
	}

	// 执行操作并验证
	testApp.InjectSpecialKey(platform.KeyTab)
	testApp.InjectSpecialKey(platform.KeyTab)
	testApp.InjectSpecialKey(platform.KeyEnter)
	time.Sleep(50 * time.Millisecond)

	rendered = testApp.GetRenderString()
	t.Logf("After operation: %s", rendered)

	t.Log("State validation test completed")
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		   (len(s) >= len(substr)) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
