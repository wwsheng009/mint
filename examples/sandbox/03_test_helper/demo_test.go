// 03_test_helper/demo_test.go
// TestHelper 链式 API 演示测试

package main

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/ui"
)

// TestHelperBasic 演示基本的链式调用
func TestHelperBasic(t *testing.T) {
	t.Log("=== TestHelper 基本链式调用演示 ===")

	testApp, err := ui.RunTestWithSandbox(FormApp,
		ui.WithWidth(40),
		ui.WithHeight(16),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()
	helper := sb.Helper()

	// 链式调用：输入用户名
	t.Log("\n--- 输入用户名 ---")
	helper.
		Type("alice").
		Press(platform.KeyTab).
		Process()

	time.Sleep(100 * time.Millisecond)

	// 验证输入
	rendered := testApp.GetRenderString()
	if contains(rendered, "alice") {
		t.Log("✅ 用户名输入成功")
	}
}

// TestHelperFormSubmit 演示表单提交
func TestHelperFormSubmit(t *testing.T) {
	t.Log("=== 表单提交演示 ===")

	testApp, err := ui.RunTestWithSandbox(FormApp,
		ui.WithWidth(40),
		ui.WithHeight(16),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()
	time.Sleep(50 * time.Millisecond)

	// 填写表单
	t.Log("\n--- 填写表单 ---")
	sb.Helper().
		Type("bob").
		Press(platform.KeyTab).  // 切换到密码框
		Type("secret123").
		Press(platform.KeyTab).  // 切换到 Submit 按钮
		Press(platform.KeyEnter). // 提交
		Process()

	time.Sleep(100 * time.Millisecond)

	// 验证提交结果
	rendered := testApp.GetRenderString()
	if contains(rendered, "Welcome, bob!") {
		t.Log("✅ 表单提交成功")
	} else {
		t.Logf("渲染结果:\n%s", rendered)
	}
}

// TestHelperWait 演示等待功能
func TestHelperWait(t *testing.T) {
	t.Log("=== 等待功能演示 ===")

	testApp, err := ui.RunTestWithSandbox(FormApp,
		ui.WithWidth(40),
		ui.WithHeight(16),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()

	// 使用 Wait 在操作之间延迟
	t.Log("\n--- 带延迟的操作序列 ---")
	sb.Helper().
		Type("test").
		Wait(50 * time.Millisecond).
		Press(platform.KeyTab).
		Wait(50 * time.Millisecond).
		Type("password").
		Process()

	time.Sleep(100 * time.Millisecond)
	t.Log("✅ 带延迟的操作完成")
}

// TestHelperClear 演示清空表单
func TestHelperClear(t *testing.T) {
	t.Log("=== 表单清空演示 ===")

	testApp, err := ui.RunTestWithSandbox(FormApp,
		ui.WithWidth(40),
		ui.WithHeight(16),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()
	time.Sleep(50 * time.Millisecond)

	// 填写表单
	t.Log("\n--- 填写表单 ---")
	sb.Helper().
		Type("user1").
		Press(platform.KeyTab).
		Type("pass1").
		Press(platform.KeyTab).
		Press(platform.KeyEnter).
		Process()

	time.Sleep(100 * time.Millisecond)

	// 清空表单
	t.Log("\n--- 清空表单 ---")
	// 按 Tab 切换到 Clear 按钮
	sb.Helper().Tab().Process()
	time.Sleep(50 * time.Millisecond)
	// 按 Enter 点击 Clear
	sb.Helper().Press(platform.KeyEnter).Process()

	time.Sleep(100 * time.Millisecond)

	rendered := testApp.GetRenderString()
	if !contains(rendered, "user1") && !contains(rendered, "pass1") {
		t.Log("✅ 表单已清空")
	}
}

// TestHelperComplexSequence 演示复杂操作序列
func TestHelperComplexSequence(t *testing.T) {
	t.Log("=== 复杂操作序列演示 ===")

	testApp, err := ui.RunTestWithSandbox(FormApp,
		ui.WithWidth(40),
		ui.WithHeight(16),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()

	// 复杂的链式调用
	t.Log("\n--- 执行复杂序列 ---")
	result := sb.Helper().
		Type("charlie").
		Press(platform.KeyTab).
		Type("mypassword").
		Press(platform.KeyTab).
		Press(platform.KeyEnter).
		Wait(100 * time.Millisecond).
		AssertRender("Welcome, charlie!").
		Result()

	if result.OK() {
		t.Log("✅ 复杂序列执行成功")
	} else {
		t.Logf("错误: %v", result.Error())
	}
}

// TestHelperKeyboardShortcuts 演示键盘快捷键
func TestHelperKeyboardShortcuts(t *testing.T) {
	t.Log("=== 键盘快捷键演示 ===")

	testApp, err := ui.RunTestWithSandbox(FormApp,
		ui.WithWidth(40),
		ui.WithHeight(16),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()

	// 使用组合键
	t.Log("\n--- 使用快捷键 ---")
	sb.Helper().
		Type("dave").
		Press(platform.KeyTab).
		Type("pwd123").
		Process()

	time.Sleep(100 * time.Millisecond)

	// 测试 Tab 在按钮间切换
	t.Log("\n--- Tab 切换焦点 ---")
	sb.Helper().
		Tab().  // Submit -> Clear
		Tab().  // Clear -> Username
		Tab().  // Username -> Password
		Process()

	time.Sleep(50 * time.Millisecond)
	t.Log("✅ 快捷键操作完成")
}

// TestHelperTypeFast 演示快速输入
func TestHelperTypeFast(t *testing.T) {
	t.Log("=== 快速输入演示 ===")

	testApp, err := ui.RunTestWithSandbox(FormApp,
		ui.WithWidth(40),
		ui.WithHeight(16),
	)
	if err != nil {
		t.Fatalf("RunTestWithSandbox failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.GetSandbox()

	// TypeFast 快速输入（无延迟）
	t.Log("\n--- 快速输入 ---")
	start := time.Now()
	sb.Helper().
		TypeFast("quicktyping").
		Process()
	elapsed := time.Since(start)

	t.Logf("快速输入完成，耗时: %v", elapsed)

	time.Sleep(100 * time.Millisecond)

	rendered := testApp.GetRenderString()
	if contains(rendered, "quicktyping") {
		t.Log("✅ 快速输入成功")
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
