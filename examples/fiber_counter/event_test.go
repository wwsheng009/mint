// examples/fiber_counter/event_test.go - 事件注入测试
package main

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/ui"
)

// TestTabEnterButtonClick 测试 Tab + Enter 触发按钮点击
func TestTabEnterButtonClick(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true")

	t.Log("=== 测试 Tab + Enter 触发按钮点击 ===")

	testApp, err := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	// 检查初始状态
	ctx := testApp.GetContext()
	t.Logf("初始: ComponentID=%s, Hooks=%d", ctx.ComponentID, len(ctx.Hooks))
	if len(ctx.Hooks) > 0 {
		t.Logf("  Hooks[0].Value = %v", ctx.Hooks[0].Value)
	}

	// 检查收集的按钮
	buttons := testApp.GetButtons()
	t.Logf("收集了 %d 个按钮", len(buttons))
	for i, btn := range buttons {
		t.Logf("  按钮 %d: label=%s", i, btn.Label())
	}

	sb := testApp.Sandbox()

	// 按 Tab 切换到第二个按钮 (+)
	t.Log("按 Tab 键...")
	sb.Helper().Tab().Process()

	// 按 Enter 触发点击
	t.Log("按 Enter 键...")
	sb.Helper().Press(platform.KeyEnter).Process()

	// 检查状态是否更新
	ctx = testApp.GetContext()
	t.Logf("点击后: ComponentID=%s, Hooks=%d", ctx.ComponentID, len(ctx.Hooks))
	if len(ctx.Hooks) > 0 {
		t.Logf("  Hooks[0].Value = %v", ctx.Hooks[0].Value)

		if ctx.Hooks[0].Value == 1 {
			t.Log("✅ 点击成功！计数器已更新为 1")
		} else {
			t.Logf("❌ 点击失败。计数器仍为 %v", ctx.Hooks[0].Value)
		}
	}
}

// TestDirectButtonClick 测试直接触发按钮点击
func TestDirectButtonClick(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true")

	t.Log("=== 测试直接触发按钮点击 ===")

	testApp, err := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	// 获取初始状态
	ctx := testApp.GetContext()
	initialValue := 0
	if len(ctx.Hooks) > 0 {
		initialValue = ctx.Hooks[0].Value.(int)
	}
	t.Logf("初始值: %d", initialValue)

	// 直接触发第二个按钮 (+) 的点击
	t.Log("触发按钮 1 (+) 的点击...")
	testApp.TriggerButtonClick(1)

	// 检查状态
	ctx = testApp.GetContext()
	if len(ctx.Hooks) > 0 {
		newValue := ctx.Hooks[0].Value.(int)
		t.Logf("点击后值: %d", newValue)

		if newValue == initialValue+1 {
			t.Logf("✅ 点击成功！值从 %d 变为 %d", initialValue, newValue)
		} else {
			t.Logf("❌ 点击失败。值从 %d 变为 %d", initialValue, newValue)
		}
	}
}

// TestMultipleButton 测试多次点击
func TestMultipleButton(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true")

	t.Log("=== 测试多次按钮点击 ===")

	testApp, err := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	// 多次点击 + 按钮
	for i := 0; i < 5; i++ {
		testApp.TriggerButtonClick(1) // 点击 + 按钮
		ctx := testApp.GetContext()
		if len(ctx.Hooks) > 0 {
			value := ctx.Hooks[0].Value.(int)
			t.Logf("点击 %d: 值 = %d", i+1, value)
		}
	}

	// 最终验证
	ctx := testApp.GetContext()
	if len(ctx.Hooks) > 0 {
		value := ctx.Hooks[0].Value.(int)
		if value == 5 {
			t.Log("✅ 多次点击成功！最终值为 5")
		} else {
			t.Logf("❌ 多次点击失败。最终值为 %d，期望 5", value)
		}
	}
}

// TestButtonLabels 验证按钮标签和顺序
func TestButtonLabels(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")

	testApp, err := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	buttons := testApp.GetButtons()
	t.Logf("收集了 %d 个按钮", len(buttons))

	if len(buttons) < 2 {
		t.Fatalf("期望至少 2 个按钮，得到 %d", len(buttons))
	}

	// 验证按钮顺序
	label0 := buttons[0].Label()
	label1 := buttons[1].Label()

	t.Logf("按钮 0: label='%s'", label0)
	t.Logf("按钮 1: label='%s'", label1)

	// 根据实现，按钮顺序应该是 [" - ", "  +  "] 或类似
	// 主要验证我们能正确收集按钮
	if strings.Contains(label0, "-") || strings.Contains(label0, "+") {
		t.Log("✅ 按钮标签包含预期的符号")
	}
}

// TestEventSequence 测试事件序列
func TestEventSequence(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true")

	t.Log("=== 测试事件序列 (Tab -> Tab -> Enter) ===")

	testApp, err := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()

	// 初始值
	ctx := testApp.GetContext()
	t.Logf("初始: Hooks[0].Value = %v", ctx.Hooks[0].Value)

	// 事件序列: Tab (到 -), Tab (到 +), Enter (触发 +)
	sb.Helper().Tab().Process()           // 切换到第一个按钮
	sb.Helper().Tab().Process()           // 切换到第二个按钮 (+)
	sb.Helper().Press(platform.KeyEnter).Process() // 触发点击

	// 检查结果
	ctx = testApp.GetContext()
	t.Logf("最终: Hooks[0].Value = %v", ctx.Hooks[0].Value)

	if ctx.Hooks[0].Value == 1 {
		t.Log("✅ 事件序列成功！计数器已更新")
	} else {
		t.Log("❌ 事件序列失败")
	}
}
