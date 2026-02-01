// examples/fiber_counter/simple_test.go - 简化的调试测试
package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox"
	"github.com/wwsheng009/mint/ui"
)

// TestContextRecreation 测试 Context 是否被重新创建
func TestContextRecreation(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true")

	t.Log("=== 测试 Context 重新创建问题 ===")

	contextPointers := make(map[string]int)

	for i := 0; i < 3; i++ {
		testApp, err := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
		if err != nil {
			t.Fatalf("TestRun failed: %v", err)
		}

		// 获取当前 Context
		ctx := ui.GetCurrentContext()
		if ctx != nil {
			ptr := fmt.Sprintf("%p", ctx)
			contextPointers[ptr]++
			t.Logf("渲染 %d: Context 指针 = %s, ComponentID = %s, Hooks 数量 = %d",
				i+1, ptr, ctx.ComponentID, len(ctx.Hooks))

			// 打印 Hooks 状态
			for j, hook := range ctx.Hooks {
				t.Logf("  Hook[%d]: Type=%v, Value=%v", j, hook.Type, hook.Value)
			}
		}

		testApp.Close()
	}

	t.Logf("\n总结:")
	if len(contextPointers) == 1 {
		for ptr, count := range contextPointers {
			t.Logf("使用同一个 Context (指针: %s, 使用次数: %d)", ptr, count)
		}
	} else {
		t.Logf("使用 %d 个不同的 Context - Context 在每次渲染时被重新创建!", len(contextPointers))
		for ptr, count := range contextPointers {
			t.Logf("  Context 指针 %s: 使用 %d 次", ptr, count)
		}
	}
}

// TestHookValueTracking 测试 Hook 值的变化
func TestHookValueTracking(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true")

	t.Log("=== 测试 Hook 值追踪 ===")

	var hookValues []interface{}
	var contextPointers []string

	for i := 0; i < 3; i++ {
		testApp, err := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
		if err != nil {
			t.Fatalf("TestRun failed: %v", err)
		}

		ctx := ui.GetCurrentContext()
		if ctx != nil && len(ctx.Hooks) > 0 {
			hookValues = append(hookValues, ctx.Hooks[0].Value)
			contextPointers = append(contextPointers, fmt.Sprintf("%p", ctx))

			val := ctx.Hooks[0].Value
			t.Logf("渲染 %d: Hook[0].Value = %v, Hook 指针 = %p",
				i+1, val, &ctx.Hooks[0])
		}

		testApp.Close()
	}

	t.Logf("\nHook 值序列: %v", hookValues)

	// 检查是否有变化
	hasChange := false
	for i := 1; i < len(hookValues); i++ {
		if hookValues[i] != hookValues[0] {
			hasChange = true
		}
	}

	if hasChange {
		t.Log("✅ Hook 值有变化")
	} else {
		t.Log("❌ Hook 值始终为 0 - setState 未生效")
	}

	// 检查 Context
	if len(contextPointers) == 1 {
		t.Log("✅ 使用同一个 Context")
	} else {
		t.Logf("❌ 使用 %d 个不同 Context", len(contextPointers))
	}
}

// TestEventClickInjection 测试点击事件是否触发
func TestEventClickInjection(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true")

	t.Log("=== 测试点击事件注入 ===")

	testApp, err := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()

	// 设置调试日志
	logMessages := []string{}
	ui.SetDebugLogger(func(msg string) {
		logMessages = append(logMessages, msg)
	})

	// 获取初始渲染
	initialRender := sb.RenderString()
	t.Logf("初始渲染:\n%s", initialRender)

	// 清空日志（清除初始渲染的日志）
	logMessages = nil

	// 模拟点击 + 按钮
	t.Log("模拟点击 + 按钮...")
	sb.Helper().Tab().Tab().Press(platform.KeyEnter).Process()

	afterRender := sb.RenderString()
	t.Logf("点击后渲染:\n%s", afterRender)

	// 分析日志
	t.Logf("\n调试日志 (%d 条):", len(logMessages))
	onClickFound := false
	setStateFound := false
	useStateFound := false

	for _, msg := range logMessages {
		if strings.Contains(msg, "onClick") {
			onClickFound = true
			t.Logf("  [FOUND] %s", msg)
		}
		if strings.Contains(msg, "setState") {
			setStateFound = true
			t.Logf("  [FOUND] %s", msg)
		}
		if strings.Contains(msg, "UseStateInt") || strings.Contains(msg, "useState") {
			useStateFound = true
		}
	}

	t.Logf("\n结果:")
	if onClickFound {
		t.Log("✅ onClick 被调用")
	} else {
		t.Log("❌ onClick 未被调用 - 事件路由问题")
	}

	if setStateFound {
		t.Log("✅ setState 被调用")
	} else {
		t.Log("❌ setState 未被调用")
	}

	if useStateFound {
		t.Log("✅ useState 被调用")
	}

	// 检查最终状态
	if strings.Contains(afterRender, "Count: 1") {
		t.Log("✅ 渲染显示 Count: 1")
	} else if strings.Contains(afterRender, "Count: 0") {
		t.Log("❌ 渲染仍显示 Count: 0")
	}
}

// TestDirectContextMutation 测试直接修改 Context
func TestDirectContextMutation(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true")

	t.Log("=== 测试直接修改 Context ===")

	testApp, err := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	// 获取 Context
	ctx := ui.GetCurrentContext()
	if ctx == nil {
		t.Fatal("无法获取 Context")
	}

	t.Logf("修改前: ComponentID=%s, Hooks 数量=%d", ctx.ComponentID, len(ctx.Hooks))
	if len(ctx.Hooks) > 0 {
		t.Logf("  Hooks[0]: Type=%v, Value=%v, 指针=%p",
			ctx.Hooks[0].Type, ctx.Hooks[0].Value, &ctx.Hooks[0])
	}

	// 直接修改 Hook 值
	if len(ctx.Hooks) > 0 {
		oldVal := ctx.Hooks[0].Value
		ctx.Hooks[0].Value = 999
		t.Logf("直接修改: Hooks[0].Value: %v -> 999", oldVal)
	}

	// 重新渲染（触发 VNode 重建）
	sb := testApp.Sandbox()
	rendered := sb.RenderString()

	t.Logf("\n修改后渲染:\n%s", rendered)

	if strings.Contains(rendered, "999") {
		t.Log("✅ 直接修改 Hook 后渲染显示新值 - useState 会读取新值")
	} else {
		t.Log("❌ 直接修改 Hook 后仍显示旧值 - useState 使用缓存或未读取 Hook")
	}
}

// TestComprehensive 综合调试测试
func TestComprehensive(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true")

	t.Log("\n╔══════════════════════════════════════════════════╗")
	t.Log("║            Fiber 问题根因分析测试                      ║")
	t.Log("╚══════════════════════════════════════════════════╝")

	// 启用详细日志
	ui.SetDebugLogger(func(msg string) {
		t.Logf("[DBG] %s", msg)
	})

	// 1. 测试 Context 复用
	t.Run("Context复用", TestContextRecreation)

	// 2. 测试 Hook 值追踪
	t.Run("Hook值追踪", TestHookValueTracking)

	// 3. 测试点击事件
	t.Run("点击事件", TestEventClickInjection)

	// 4. 测试直接修改
	t.Run("直接修改Context", TestDirectContextMutation)
}
