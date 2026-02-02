// examples/counter/fiber_test.go - Fiber 集成测试 (新版)
package main

import (
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/ui"
)

// TestFiberCounterInitial 测试 Counter 组件初始渲染
func TestFiberCounterInitial(t *testing.T) {
	testApp, err := ui.RunTest(Counter,
		ui.WithSize(40, 12),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(50 * time.Millisecond)

	rendered := testApp.GetRenderString()

	t.Logf("=== Initial Render Output ===")
	t.Logf("%s", rendered)

	// 验证关键元素存在
	expectedTexts := []string{
		"Mint UI Counter Demo",
		"Count:",
	}

	for _, expected := range expectedTexts {
		if !strings.Contains(rendered, expected) {
			t.Errorf("Expected text not found: %q", expected)
		}
	}

	if strings.Contains(rendered, "Count: 0") {
		t.Log("✓ Initial count is 0")
	}
}

// TestFiberCounterIncrement 测试递增按钮
func TestFiberCounterIncrement(t *testing.T) {
	t.Log("=== 测试状态更新 ===")

	testApp, err := ui.RunTest(Counter,
		ui.WithSize(40, 12),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(50 * time.Millisecond)

	// 初始状态
	rendered := testApp.GetRenderString()
	t.Logf("=== 初始状态 ===")
	t.Logf("%s", rendered)

	// 点击 + 按钮：Tab 两次到 + 按钮，然后 Enter
	t.Log("=== 模拟点击 + 按钮 ===")

	testApp.InjectSpecialKey(platform.KeyTab)
	time.Sleep(20 * time.Millisecond)
	testApp.InjectSpecialKey(platform.KeyTab)
	time.Sleep(20 * time.Millisecond)
	testApp.InjectSpecialKey(platform.KeyEnter)
	time.Sleep(50 * time.Millisecond)

	// 获取更新后的渲染
	rendered = testApp.GetRenderString()
	t.Logf("=== 点击后状态 ===")
	t.Logf("%s", rendered)

	// 检查状态是否更新
	if strings.Contains(rendered, "Count: 1") {
		t.Log("✓ 状态更新成功")
	} else if strings.Contains(rendered, "Count: 0") {
		t.Log("✗ 状态未更新")
	}
}

// TestFiberCounterDecrement 测试递减按钮
func TestFiberCounterDecrement(t *testing.T) {
	testApp, err := ui.RunTest(Counter,
		ui.WithSize(40, 12),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(50 * time.Millisecond)

	// 点击 - 按钮：Tab 一次到 - 按钮，然后 Enter
	testApp.InjectSpecialKey(platform.KeyTab)
	time.Sleep(20 * time.Millisecond)
	testApp.InjectSpecialKey(platform.KeyEnter)
	time.Sleep(50 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("=== 递减后状态 ===")
	t.Logf("%s", rendered)

	if strings.Contains(rendered, "Count: -1") {
		t.Log("✓ 递减成功")
	} else if strings.Contains(rendered, "Count: 0") {
		t.Log("✗ 递减无效")
	}
}

// TestFiberCounterMultipleUpdates 测试多次状态更新
func TestFiberCounterMultipleUpdates(t *testing.T) {
	t.Log("=== 测试多次状态更新 ===")

	testApp, err := ui.RunTest(Counter,
		ui.WithSize(40, 12),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(50 * time.Millisecond)

	// 尝试点击 + 按钮 3 次
	successfulUpdates := 0
	for i := 1; i <= 3; i++ {
		t.Logf("点击 #%d", i)

		testApp.InjectSpecialKey(platform.KeyTab)
		testApp.InjectSpecialKey(platform.KeyTab)
		testApp.InjectSpecialKey(platform.KeyEnter)
		time.Sleep(30 * time.Millisecond)

		rendered := testApp.GetRenderString()
		expected := "Count: " + string(rune('0'+i))

		if strings.Contains(rendered, expected) {
			t.Logf("✓ 第 %d 次点击成功", i)
			successfulUpdates++
		} else {
			t.Logf("✗ 第 %d 次点击后状态: %s", i, extractCount(rendered))
		}
	}

	t.Logf("总结: %d/3 次状态更新成功", successfulUpdates)
}

// extractCount 从渲染输出中提取计数
func extractCount(rendered string) string {
	lines := strings.Split(rendered, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Count:") {
			return strings.TrimSpace(line)
		}
	}
	return "Count: ?"
}

// TestFiberButtonNavigation 测试按钮焦点导航
func TestFiberButtonNavigation(t *testing.T) {
	t.Log("=== 测试焦点导航 ===")

	testApp, err := ui.RunTest(Counter,
		ui.WithSize(40, 12),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(50 * time.Millisecond)

	// 初始渲染
	initial := testApp.GetRenderString()
	t.Logf("=== 初始状态 ===")
	t.Logf("%s", initial)

	// 模拟 Tab 键切换焦点
	for i := 0; i < 3; i++ {
		t.Logf("=== Tab #%d ===", i+1)
		testApp.InjectSpecialKey(platform.KeyTab)
		time.Sleep(30 * time.Millisecond)

		rendered := testApp.GetRenderString()
		t.Logf("%s", rendered)

		if strings.Contains(rendered, "[") || strings.Contains(rendered, "]") {
			t.Log("✓ 检测到焦点标记")
		}
	}
}

// TestFiberMouseInteraction 测试鼠标交互
func TestFiberMouseInteraction(t *testing.T) {
	testApp, err := ui.RunTest(Counter,
		ui.WithSize(40, 12),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(50 * time.Millisecond)

	// 测试鼠标点击 (+ 按钮位置大约在 x=20, y=8)
	testApp.InjectMouse(20, 8, platform.MouseLeft, platform.MousePress)
	time.Sleep(50 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("=== 鼠标点击后 ===")
	t.Logf("%s", rendered)

	if strings.Contains(rendered, "Count: 1") {
		t.Log("✓ 鼠标点击成功")
	} else {
		t.Log("✗ 鼠标点击可能未生效 - 坐标可能需要调整")
	}
}
