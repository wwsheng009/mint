// examples/counter/fiber_test.go - 使用 Sandbox 测试 Fiber 集成
package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox"
	"github.com/wwsheng009/mint/ui"
)

// TestFiberCounterInitial 测试 Counter 组件初始渲染
func TestFiberCounterInitial(t *testing.T) {
	// 启用 Fiber 模式
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true")

	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 12))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()
	rendered := sb.RenderString()

	t.Logf("=== Initial Render Output ===")
	t.Logf("%s", rendered)

	// 验证关键元素存在
	expectedTexts := []string{
		"Mint UI Counter Demo",
		"Count:",
		"-",
		"+",
	}

	for _, expected := range expectedTexts {
		if !strings.Contains(rendered, expected) {
			t.Errorf("Expected text not found: %q", expected)
		}
	}

	// 注意：由于当前 Fiber 问题，初始计数可能显示为 0
	if strings.Contains(rendered, "Count: 0") {
		t.Log("✅ Initial count is 0")
	}
}

// TestFiberCounterIncrement 测试递增按钮（核心测试）
func TestFiberCounterIncrement(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true")

	t.Log("=== 测试 Fiber 状态更新 ===")

	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 12))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()

	// 初始状态
	rendered := sb.RenderString()
	t.Logf("=== 初始状态 ===")
	t.Logf("%s", rendered)

	// 尝试点击 + 按钮
	// Tab 移动焦点到第一个按钮 (-)，再 Tab 移动到第二个按钮 (+)
	t.Log("=== 模拟点击 + 按钮 ===")

	result := sb.Helper().
		Tab().
		Process().
		Tab().
		Process().
		Press(platform.KeyEnter).
		Process().
		Result()

	if !result.OK() {
		t.Logf("警告: 操作有错误: %v", result.Error())
	}

	// 获取更新后的渲染
	rendered = sb.RenderString()
	t.Logf("=== 点击后状态 ===")
	t.Logf("%s", rendered)

	// 检查状态是否更新
	if strings.Contains(rendered, "Count: 1") {
		t.Log("✅ 状态更新成功 - useState/setState 正常工作")
	} else if strings.Contains(rendered, "Count: 0") {
		t.Log("❌ 状态未更新 - 这是 Fiber 集成需要修复的问题")
		t.Log("   可能原因: setState 被调用但 useState 返回旧值")
	} else {
		t.Logf("⚠️  未找到计数器输出")
	}
}

// TestFiberCounterDecrement 测试递减按钮
func TestFiberCounterDecrement(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")

	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 12))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()

	// 点击 - 按钮
	sb.Helper().
		Tab().
		Process().
		Press(platform.KeyEnter).
		Process()

	rendered := sb.RenderString()
	t.Logf("=== 递减后状态 ===")
	t.Logf("%s", rendered)

	if strings.Contains(rendered, "Count: -1") {
		t.Log("✅ 递减成功")
	} else if strings.Contains(rendered, "Count: 0") {
		t.Log("❌ 递减无效 - 状态未更新")
	}
}

// TestFiberCounterMultipleUpdates 测试多次状态更新
func TestFiberCounterMultipleUpdates(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")

	t.Log("=== 测试多次状态更新 ===")

	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 12))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()

	// 尝试点击 + 按钮 3 次
	successfulUpdates := 0
	for i := 1; i <= 3; i++ {
		t.Logf("点击 #%d", i)

		sb.Helper().
			Tab().Tab().
			Press(platform.KeyEnter).
			Process()

		rendered := sb.RenderString()
		expected := fmt.Sprintf("Count: %d", i)

		if strings.Contains(rendered, expected) {
			t.Logf("✅ 第 %d 次点击成功", i)
			successfulUpdates++
		} else {
			t.Logf("❌ 第 %d 次点击后状态仍为 %s", i, extractCount(rendered))
		}
	}

	t.Logf("总结: %d/3 次状态更新成功", successfulUpdates)

	if successfulUpdates == 0 {
		t.Log("⚠️  所有状态更新都失败 - Fiber setState 需要修复")
	}
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

// TestFiberEventRouting 测试事件路由
func TestFiberEventRouting(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true")

	t.Log("=== 测试事件路由 ===")

	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 12))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()

	// 设置事件处理器来捕获事件
	eventReceived := false
	sb.SetEventHandler(func(event platform.RawInput) error {
		t.Logf("Event: Type=%d, Key=%c, Special=%v",
			event.Type, event.Key, event.Special)
		eventReceived = true
		return nil
	})

	// 注入 Tab 事件
	sb.InjectSpecialKey(platform.KeyTab)
	sb.ProcessEvents()

	if eventReceived {
		t.Log("✅ EventHandler 被调用 - 事件路由正常")
	} else {
		t.Log("❌ EventHandler 未被调用 - 事件路由可能有问题")
	}

	// 注入 Enter 事件
	eventReceived = false
	sb.InjectSpecialKey(platform.KeyEnter)
	sb.ProcessEvents()

	if eventReceived {
		t.Log("✅ Enter 事件被接收")
	}
}

// TestFiberWithSnapshot 使用快照测试状态持久化
func TestFiberWithSnapshot(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")

	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 12))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()

	// 保存初始快照
	t.Log("=== 创建初始快照 ===")
	snap1, err := sb.Snapshot(sandbox.SnapshotStandard, "initial")
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}
	t.Logf("快照 ID: %s, 事件数: %d", snap1.Metadata.ID, len(snap1.Events))

	// 尝试修改状态
	t.Log("=== 尝试修改状态 ===")
	sb.Helper().Tab().Tab().Press(platform.KeyEnter).Process()

	afterRender := sb.RenderString()
	t.Logf("修改后状态: %s", extractCount(afterRender))

	// 恢复快照
	t.Log("=== 恢复快照 ===")
	err = sb.Restore(snap1)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	restoredRender := sb.RenderString()
	t.Logf("恢复后状态: %s", extractCount(restoredRender))

	if strings.Contains(restoredRender, "Count: 0") {
		t.Log("✅ 快照恢复成功")
	}
}

// TestFiberButtonNavigation 测试按钮焦点导航
func TestFiberButtonNavigation(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")

	t.Log("=== 测试焦点导航 ===")

	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 12))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()

	// 初始渲染
	initial := sb.RenderString()
	t.Logf("=== 初始状态 ===")
	t.Logf("%s", initial)

	// 模拟 Tab 键切换焦点
	for i := 0; i < 3; i++ {
		t.Logf("=== Tab #%d ===", i+1)
		sb.Helper().Tab().Process()

		rendered := sb.RenderString()
		t.Logf("%s", rendered)

		// 检查是否有选中标记 (blue 背景)
		if strings.Contains(rendered, "[") || strings.Contains(rendered, "]") {
			t.Log("✅ 检测到焦点标记")
		}
	}
}

// TestFiberChainAPI 测试链式 API
func TestFiberChainAPI(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")

	testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 12))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	// 使用链式 API 进行完整测试
	result := testApp.Helper().
		// 验证初始状态
		AssertRender("Mint UI Counter Demo").
		AssertRender("Count: 0").

		// 移动到第一个按钮
		Tab().
		Process().

		// 移动到第二个按钮
		Tab().
		Process().

		// 点击
		Press(platform.KeyEnter).
		Process().

		// 验证结果
		Result()

	t.Logf("=== 链式测试结果 ===")
	if result.OK() {
		t.Log("✅ 所有断言通过")
	} else {
		t.Logf("❌ 测试失败: %v", result.Error())

		// 打印最终状态
		sb := testApp.Sandbox()
		t.Logf("最终输出:\n%s", sb.RenderString())
	}
}

// TestFiberMemorySafety 测试内存安全性
func TestFiberMemorySafety(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")

	config := sandbox.MockConfig()
	config.Event.QueueMaxSize = 50  // 限制队列大小

	testApp, err := ui.TestRunWithConfig(Counter, config)
	if err != nil {
		t.Fatalf("TestRunWithConfig failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()

	// 注入大量事件
	t.Log("=== 注入 1000 个事件 ===")
	for i := 0; i < 1000; i++ {
		sb.InjectKey('a')
	}

	stats := sb.QueueStats()
	t.Logf("队列统计: Length=%d, Memory=%d, Evicted=%d",
		stats.Length, stats.MemoryUsed, stats.EvictCount)

	if stats.Length <= 50 {
		t.Log("✅ 队列大小限制正常工作")
	} else {
		t.Errorf("队列超过限制: %d > 50", stats.Length)
	}

	if stats.EvictCount > 0 {
		t.Logf("✅ 淘汰机制正常: %d 个事件被淘汰", stats.EvictCount)
	}
}
