// examples/fiber_counter/full_app_test.go - 完整应用测试（使用 RunTest 自动化）
package main

import (
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/ui"
)

// TestFullAppWithAutoRun 测试完整应用（自动化模式）
// 使用 RunTest 代替 Run，可以自动注入事件并验证结果
func TestFullAppWithAutoRun(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "false")

	// 使用 RunTest 创建可测试的应用
	testApp, err := ui.RunTest(DebugCounter,
		ui.WithWidth(40),
		ui.WithHeight(12),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	// 等待应用初始化
	time.Sleep(100 * time.Millisecond)

	// 获取初始渲染
	initialRender := testApp.GetRenderString()
	t.Logf("Initial render:\n%s", initialRender)

	if !strings.Contains(initialRender, "Count:") {
		t.Error("Initial render should contain 'Count:'")
	}
}

// TestFullAppIntegration 测试完整应用集成（使用 RunTest）
func TestFullAppIntegration(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "false")

	t.Log("=== 测试完整应用集成（使用 RunTest） ===")

	testApp, err := ui.RunTest(DebugCounter,
		ui.WithWidth(40),
		ui.WithHeight(12),
	)
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(100 * time.Millisecond)

	// 按 Tab 切换焦点
	if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
		t.Fatalf("InjectSpecialKey failed: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	// 按 Enter 点击
	if err := testApp.InjectSpecialKey(platform.KeyEnter); err != nil {
		t.Fatalf("InjectSpecialKey failed: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	// 检查结果
	rendered := testApp.GetRenderString()
	if strings.Contains(rendered, "Count: 1") || strings.Contains(rendered, "Count:1") {
		t.Log("✅ Full app integration test passed!")
	} else {
		t.Error("⚠️  Count should be 1 after click")
	}
}
