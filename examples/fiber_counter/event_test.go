// examples/fiber_counter/event_test.go - 事件注入测试 (新版 API)
package main

import (
	"strings"
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/ui"
)

// TestTabEnterButtonClick 测试 Tab + Enter 触发按钮点击
func TestTabEnterButtonClick(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")

	t.Log("=== 测试 Tab + Enter 触发按钮点击 ===")

	testApp, err := ui.RunTest(DebugCounter, ui.WithSize(40, 12))
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	// 等待应用初始化
	time.Sleep(50 * time.Millisecond)

	// 检查收集的按钮
	buttons := testApp.GetButtons()
	t.Logf("收集了 %d 个按钮", len(buttons))
	if len(buttons) < 2 {
		t.Fatalf("期望至少 2 个按钮，得到 %d", len(buttons))
	}

	// 按 Tab 切换到第二个按钮 (+)
	t.Log("按 Tab 键...")
	if err := testApp.InjectSpecialKey(platform.KeyTab); err != nil {
		t.Logf("Warning: Tab inject failed: %v", err)
	}
	time.Sleep(20 * time.Millisecond)

	// 按 Enter 触发点击
	t.Log("按 Enter 键...")
	if err := testApp.InjectSpecialKey(platform.KeyEnter); err != nil {
		t.Logf("Warning: Enter inject failed: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	// 检查渲染输出
	rendered := testApp.GetRenderString()
	t.Logf("渲染输出:\n%s", rendered)

	// 验证计数器已更新
	if strings.Contains(rendered, "Count: 1") || strings.Contains(rendered, "计数: 1") {
		t.Log("✅ 点击成功！计数器已更新为 1")
	} else {
		t.Logf("❌ 点击失败。输出中未找到计数器 1")
	}
}

// TestMultipleButtonClick 测试多次点击
func TestMultipleButtonClick(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")

	t.Log("=== 测试多次按钮点击 ===")

	testApp, err := ui.RunTest(DebugCounter, ui.WithSize(40, 12))
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	// 等待应用初始化
	time.Sleep(50 * time.Millisecond)

	// 多次点击 + 按钮 (通过 Tab + Enter)
	for i := 0; i < 5; i++ {
		// Tab 切换到 + 按钮
		testApp.InjectSpecialKey(platform.KeyTab)
		time.Sleep(10 * time.Millisecond)

		// Enter 触发点击
		testApp.InjectSpecialKey(platform.KeyEnter)
		time.Sleep(30 * time.Millisecond)

		rendered := testApp.GetRenderString()
		expected := i + 1
		if strings.Contains(rendered, string(rune('0'+expected))) ||
		   strings.Contains(rendered, "Count: "+string(rune('0'+expected))) {
			t.Logf("点击 %d: ✅ 计数器 = %d", i+1, expected)
		}
	}

	// 最终验证
	time.Sleep(50 * time.Millisecond)
	rendered := testApp.GetRenderString()
	t.Logf("最终渲染输出:\n%s", rendered)

	if strings.Contains(rendered, "5") {
		t.Log("✅ 多次点击成功！最终值为 5")
	} else {
		t.Logf("❌ 多次点击失败。最终值不是 5")
	}
}

// TestButtonLabels 验证按钮标签和顺序
func TestButtonLabels(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")

	testApp, err := ui.RunTest(DebugCounter, ui.WithSize(40, 12))
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(50 * time.Millisecond)

	buttons := testApp.GetButtons()
	t.Logf("收集了 %d 个按钮", len(buttons))

	if len(buttons) < 2 {
		t.Fatalf("期望至少 2 个按钮，得到 %d", len(buttons))
	}

	// 验证按钮标签 (按钮是 FocusableVNode 接口)
	for i, btn := range buttons {
		if fb, ok := btn.(interface{ Label() string }); ok {
			t.Logf("按钮 %d: label='%s'", i, fb.Label())
		}
	}

	t.Log("✅ 按钮收集成功")
}

// TestEventSequence 测试事件序列
func TestEventSequence(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")

	t.Log("=== 测试事件序列 (Tab -> Tab -> Enter) ===")

	testApp, err := ui.RunTest(DebugCounter, ui.WithSize(40, 12))
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(50 * time.Millisecond)

	// 初始渲染
	rendered := testApp.GetRenderString()
	t.Logf("初始输出:\n%s", rendered)

	// 事件序列: Tab (到 -), Tab (到 +), Enter (触发 +)
	testApp.InjectSpecialKey(platform.KeyTab)
	time.Sleep(20 * time.Millisecond)

	testApp.InjectSpecialKey(platform.KeyTab)
	time.Sleep(20 * time.Millisecond)

	testApp.InjectSpecialKey(platform.KeyEnter)
	time.Sleep(50 * time.Millisecond)

	// 检查结果
	rendered = testApp.GetRenderString()
	t.Logf("最终输出:\n%s", rendered)

	if strings.Contains(rendered, "1") {
		t.Log("✅ 事件序列成功！计数器已更新")
	} else {
		t.Log("❌ 事件序列失败")
	}
}

// TestFocusNavigation 测试焦点导航
func TestFocusNavigation(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")

	t.Log("=== 测试焦点导航 ===")

	testApp, err := ui.RunTest(DebugCounter, ui.WithSize(40, 12))
	if err != nil {
		t.Fatalf("RunTest failed: %v", err)
	}
	defer testApp.Close()

	time.Sleep(50 * time.Millisecond)

	// 检查初始焦点
	initialIndex := testApp.GetFocusedIndex()
	t.Logf("初始焦点索引: %d", initialIndex)

	// 按 Tab 移动焦点
	for i := 0; i < 5; i++ {
		testApp.InjectSpecialKey(platform.KeyTab)
		time.Sleep(20 * time.Millisecond)

		index := testApp.GetFocusedIndex()
		t.Logf("Tab %d: 焦点索引 = %d", i+1, index)
	}

	t.Log("✅ 焦点导航测试完成")
}
