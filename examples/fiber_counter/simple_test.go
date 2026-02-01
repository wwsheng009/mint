// examples/fiber_counter/simple_test.go - 简化的调试测试
package main

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/ui"
)

// TestSimpleContextRecreation 简单测试 Context 是否被重新创建
// 与 debug_test.go 中的测试不同，这里使用 TestApp.GetContext()
func TestSimpleContextRecreation(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true")

	t.Log("=== 测试 Context 重新创建问题 (使用 TestApp.GetContext) ===")

	contextPointers := make(map[string]int)

	for i := 0; i < 3; i++ {
		testApp, err := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
		if err != nil {
			t.Fatalf("TestRun failed: %v", err)
		}

		// 使用 TestApp.GetContext() 获取 Context
		ctx := testApp.GetContext()
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
			t.Logf("❌ 使用同一个 Context (指针: %s, 使用次数: %d)", ptr, count)
			t.Log("   这是问题！每次 TestRun 应该创建新的 Context")
		}
	} else {
		t.Logf("✅ 使用 %d 个不同的 Context - 每次测试创建独立 Context", len(contextPointers))
		for ptr, count := range contextPointers {
			t.Logf("  Context 指针 %s: 使用 %d 次", ptr, count)
		}
	}
}

// TestHookValueInSameApp 测试同一 TestApp 中多次渲染的 Hook 值变化
func TestHookValueInSameApp(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true")

	t.Log("=== 测试同一 TestApp 中多次渲染的 Hook 值 ===")

	testApp, err := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	var hookValues []interface{}
	contextPtr := ""

	// 多次渲染同一个 TestApp
	for i := 0; i < 3; i++ {
		testApp.Render()

		ctx := testApp.GetContext()
		if ctx != nil {
			if contextPtr == "" {
				contextPtr = fmt.Sprintf("%p", ctx)
			}
			if len(ctx.Hooks) > 0 {
				hookValues = append(hookValues, ctx.Hooks[0].Value)
				t.Logf("渲染 %d: Hook[0].Value = %v, Context=%s",
					i+1, ctx.Hooks[0].Value, fmt.Sprintf("%p", ctx))
			}
		}
	}

	t.Logf("\nHook 值序列: %v", hookValues)
	t.Logf("Context 指针: %s", contextPtr)

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
	ctx := testApp.GetContext()
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

	// 重新渲染
	testApp.Render()

	// 检查修改后的值
	ctx = testApp.GetContext()
	if ctx != nil && len(ctx.Hooks) > 0 {
		newVal := ctx.Hooks[0].Value
		t.Logf("重新渲染后: Hooks[0].Value = %v", newVal)

		if newVal == 999 {
			t.Log("❌ 直接修改后值仍是 999 - useState 重新初始化了 Hook")
		} else if newVal == 0 {
			t.Log("✅ useState 重新初始化为 0 - 预期行为")
		}
	}
}

// TestComprehensiveSimple 综合调试测试（简化版）
func TestComprehensiveSimple(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true")

	t.Log("\n╔══════════════════════════════════════════════════╗")
	t.Log("║         Fiber 问题根因分析测试（简化版）              ║")
	t.Log("╚══════════════════════════════════════════════════╝")

	// 1. 测试 Context 复用
	t.Run("Context复用", TestSimpleContextRecreation)

	// 2. 测试 Hook 值追踪
	t.Run("Hook值追踪", TestHookValueInSameApp)

	// 3. 测试直接修改
	t.Run("直接修改Context", TestDirectContextMutation)
}
