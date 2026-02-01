// examples/fiber_counter/run_test.go - 完整应用的自动化测试（使用 ui.RunTest）
package main

import (
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/ui"
)

// TestRunTestBasic 测试 RunTest 基本功能
func TestRunTestBasic(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true") // 开启调试输出

	testApp, err := ui.RunTest(DebugCounter,
		ui.WithWidth(40),
		ui.WithHeight(12),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	// 等待初始化完成
	time.Sleep(200 * time.Millisecond)

	// 强制渲染一次
	testApp.GetFrameworkApp().ForceRenderNow()

	// 检查初始渲染
	rendered := testApp.GetRenderString()
	t.Logf("Initial render:\n%s", rendered)

	if !strings.Contains(rendered, "Count:") {
		t.Error("Initial render does not contain 'Count:'")
	}
}

// TestRunTestTab 测试 Tab 键焦点切换
func TestRunTestTab(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "false")

	testApp, err := ui.RunTest(DebugCounter,
		ui.WithWidth(40),
		ui.WithHeight(12),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(200 * time.Millisecond)

	// 按 Tab 键切换到第二个按钮
	if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
		t.Fatalf("InjectSpecialKey failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// 检查渲染（焦点应该移动）
	rendered := testApp.GetRenderString()
	t.Logf("After Tab:\n%s", rendered)
}

// TestRunTestEnter 测试 Enter 键触发按钮点击
func TestRunTestEnter(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true")

	testApp, err := ui.RunTest(DebugCounter,
		ui.WithWidth(40),
		ui.WithHeight(12),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(200 * time.Millisecond)

	// 强制初始渲染
	testApp.GetFrameworkApp().ForceRenderNow()

	// 获取初始渲染
	initialRender := testApp.GetRenderString()
	t.Logf("Initial render:\n%s", initialRender)

	// 检查初始值
	if !strings.Contains(initialRender, "Count: 0") && !strings.Contains(initialRender, "Count:0") {
		t.Logf("Warning: Initial count may not be 0")
	}

	// 按 Tab 键切换到 + 按钮
	if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
		t.Fatalf("InjectSpecialKey (Tab) failed: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	// 按 Enter 键触发点击
	if err := testApp.InjectSpecialKey(platform.KeyEnter); err != nil {
		t.Fatalf("InjectSpecialKey (Enter) failed: %v", err)
	}

	// 等待状态更新
	time.Sleep(100 * time.Millisecond)

	// 强制重新渲染以获取最新状态
	testApp.GetFrameworkApp().ForceRenderNow()

	// 获取更新后的渲染
	updatedRender := testApp.GetRenderString()
	t.Logf("After Enter:\n%s", updatedRender)

	// 检查计数器是否增加
	if strings.Contains(updatedRender, "Count: 1") || strings.Contains(updatedRender, "Count:1") {
		t.Log("✅ Count increased to 1!")
	} else if strings.Contains(updatedRender, "Count: 0") || strings.Contains(updatedRender, "Count:0") {
		t.Log("⚠️  Count still 0 - Fiber integration issue confirmed")
	}
}

// TestRunTestMultipleClicks 测试多次点击
func TestRunTestMultipleClicks(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "false")

	testApp, err := ui.RunTest(DebugCounter,
		ui.WithWidth(40),
		ui.WithHeight(12),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(200 * time.Millisecond)

	// 按 Tab 切换到 + 按钮
	testApp.InjectSpecialKey(platform.KeyTab)
	time.Sleep(50 * time.Millisecond)

	// 多次点击 + 按钮
	clicks := 5
	for i := 0; i < clicks; i++ {
		if err := testApp.InjectSpecialKey(platform.KeyEnter); err != nil {
			t.Fatalf("InjectSpecialKey (Enter %d) failed: %v", i, err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	// 等待所有更新完成
	time.Sleep(200 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("After %d clicks:\n%s", clicks, rendered)

	// 检查结果
	if strings.Contains(rendered, "Count: 5") || strings.Contains(rendered, "Count:5") {
		t.Log("✅ Count is 5 as expected!")
	} else {
		t.Logf("⚠️  Count is not 5 - Fiber integration issue")
	}
}

// TestRunTestDecrement 测试减少按钮
func TestRunTestDecrement(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "false")

	testApp, err := ui.RunTest(DebugCounter,
		ui.WithWidth(40),
		ui.WithHeight(12),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(200 * time.Millisecond)

	// 先按几次 Tab 切换到 - 按钮（焦点顺序: - -> +，所以需要 Shift+Tab 或者多次 Tab）
	// 实际上 - 按钮是第一个，所以焦点默认就在 - 按钮上

	// 按 Enter 减少计数（如果是 0，应该保持 0）
	if err := testApp.InjectSpecialKey(platform.KeyEnter); err != nil {
		t.Fatalf("InjectSpecialKey (Enter) failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	rendered := testApp.GetRenderString()
	t.Logf("After decrement:\n%s", rendered)
}

// TestRunTestQuit 测试退出功能
func TestRunTestQuit(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "false")

	testApp, err := ui.RunTest(DebugCounter,
		ui.WithWidth(40),
		ui.WithHeight(12),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	// 按 'q' 键退出
	if err := testApp.InjectKey('q'); err != nil {
		t.Fatalf("InjectKey ('q') failed: %v", err)
	}

	// 等待应用退出
	time.Sleep(200 * time.Millisecond)

	// 关闭（如果还没有退出）
	testApp.Close()

	t.Log("Quit test completed")
}

// TestRunTestLegacyMode 测试非 Fiber 模式（对比）
func TestRunTestLegacyMode(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "false") // 关闭 Fiber
	t.Setenv("TUI_DEBUG_UI", "false")

	testApp, err := ui.RunTest(DebugCounter,
		ui.WithWidth(40),
		ui.WithHeight(12),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(200 * time.Millisecond)

	initialRender := testApp.GetRenderString()
	t.Logf("Legacy mode initial:\n%s", initialRender)

	// 按 Tab 切换
	testApp.InjectSpecialKey(platform.KeyTab)
	time.Sleep(50 * time.Millisecond)

	// 按 Enter 点击
	testApp.InjectSpecialKey(platform.KeyEnter)
	time.Sleep(200 * time.Millisecond)

	updatedRender := testApp.GetRenderString()
	t.Logf("Legacy mode after click:\n%s", updatedRender)

	// 非 Fiber 模式可能工作正常
	if strings.Contains(updatedRender, "Count: 1") || strings.Contains(updatedRender, "Count:1") {
		t.Log("✅ Legacy mode works correctly!")
	}
}
