// examples/fiber_counter/debug_test.go - 使用 Sandbox 观察内部状态定位问题
package main

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox"
	"github.com/wwsheng009/mint/ui"
)

// TestDebugContextInfo 测试获取内部 Context 信息
func TestDebugContextInfo(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true")

	t.Log("=== 测试 1: 获取初始 Context 信息 ===")

	testApp, err := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	// 获取内部状态
	ctxInfo := ui.DebugContextInfo()
	t.Logf("Context Info: %+v", ctxInfo)

	// 验证初始状态
	if hasCtx, ok := ctxInfo["hasContext"].(bool); ok && hasCtx {
		if componentID, ok := ctxInfo["componentID"].(string); ok {
			t.Logf("ComponentID: %s", componentID)
		}
		if hooksCount, ok := ctxInfo["hooksCount"].(int); ok {
			t.Logf("Hooks Count: %d", hooksCount)
			if hooksCount > 0 {
				t.Log("✅ Context 有 Hooks")
			}
		}
	}
}

// TestHooksPointerStability 测试 Hook 指针稳定性
func TestHooksPointerStability(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true")

	t.Log("=== 测试 2: Hook 指针稳定性 ===")

	// 收集多次渲染中的 hook 指针
	pointers := make([]string, 0, 5)
	var componentIDs []string

	// 设置一个渲染计数器
	renderCount := 0
	maxRenders := 3

	for i := 0; i < maxRenders; i++ {
		testApp, err := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
		if err != nil {
			t.Fatalf("TestRun failed: %v", err)
		}

		ctxInfo := ui.DebugContextInfo()
		if hasCtx, ok := ctxInfo["hasContext"].(bool); ok && hasCtx {
			if ptr, ok := ctxInfo["contextPointer"].(string); ok {
				pointers = append(pointers, ptr)
			}
			if compID, ok := ctxInfo["componentID"].(string); ok {
				componentIDs = append(componentIDs, compID)
			}
		}

		testApp.Close()
		renderCount++
	}

	t.Logf("完成了 %d 次渲染", renderCount)
	t.Logf("ComponentIDs: %v", componentIDs)
	t.Logf("Context Pointers: %v", pointers)

	// 检查 Context 是否被复用
	uniqueContexts := make(map[string]bool)
	for _, ptr := range pointers {
		uniqueContexts[ptr] = true
	}

	if len(uniqueContexts) == 1 {
		t.Log("✅ Context 被复用 - 所有渲染使用同一个 Context")
	} else {
		t.Logf("❌ Context 未复用 - 有 %%d 个不同的 Context 实例", len(uniqueContexts))
		t.Log("   这可能是问题的根源！每次渲染创建新 Context 导致 Hooks 丢失")
	}

	// 检查 ComponentID 是否一致
	uniqueCompIDs := make(map[string]bool)
	for _, id := range componentIDs {
		uniqueCompIDs[id] = true
	}

	if len(uniqueCompIDs) == 1 {
		t.Log("✅ ComponentID 一致")
	} else {
		t.Logf("❌ ComponentID 不一致: %v", uniqueCompIDs)
	}
}

// TestHookValueAcrossRenders 测试 Hook 值在多次渲染中的变化
func TestHookValueAcrossRenders(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true")

	t.Log("=== 测试 3: Hook 值在多次渲染中的变化 ===")

	values := make([]interface{}, 0)

	for i := 0; i < 3; i++ {
		testApp, err := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
		if err != nil {
			t.Fatalf("TestRun failed: %v", err)
		}

		ctxInfo := ui.DebugContextInfo()
		if hasCtx, ok := ctxInfo["hasContext"].(bool); ok && hasCtx {
			if hooks, ok := ctxInfo["hooks"].([]map[string]interface{}); ok && len(hooks) > 0 {
				if hookValue, ok := hooks[0]["value"]; ok {
					values = append(values, hookValue)
					t.Logf("渲染 #%d: Hook[0].Value = %v", i+1, hookValue)
				}
			}
		}

		testApp.Close()
	}

	// 检查值是否变化
	allZero := true
	for _, v := range values {
		if val, ok := v.(int); ok && val != 0 {
			allZero = false
		}
	}

	if allZero {
		t.Log("❌ 所有渲染中 Hook 值始终为 0 - setState 未生效")
	} else {
		t.Log("✅ Hook 值有变化 - setState 可能正常")
	}
}

// TestStateUpdateWithDirectInjection 直接测试状态更新
func TestStateUpdateWithDirectInjection(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true")

	t.Log("=== 测试 4: 直接调用 setState 后观察状态 ===")

	testApp, err := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()

	// 获取初始状态
	initialInfo := ui.DebugContextInfo()
	t.Logf("初始状态: %+v", initialInfo)

	// 尝试模拟状态更新
	// 这需要访问组件实例来调用 setState
	// 让我们通过点击按钮来触发

	t.Log("模拟点击 + 按钮...")

	// 使用 Tab 移动焦点，Enter 点击
	sb.Helper().
		Tab().Tab().
		Press(platform.KeyEnter).
		Process()

	// 等待渲染完成
	// 获取更新后的状态
	updatedInfo := ui.DebugContextInfo()
	t.Logf("更新后状态: %+v", updatedInfo)

	// 比较差异
	if initialHooks, ok := initialInfo["hooks"].([]map[string]interface{}); ok {
		if updatedHooks, ok := updatedInfo["hooks"].([]map[string]interface{}); ok {
			if len(initialHooks) > 0 && len(updatedHooks) > 0 {
				initialVal := initialHooks[0]["value"]
				updatedVal := updatedHooks[0]["value"]

				t.Logf("Hook 值变化: %v -> %v", initialVal, updatedVal)

				if initialVal == updatedVal && initialVal == 0 {
					t.Log("❌ Hook 值未变化 - setState 未生效或 useState 未读取新值")
				} else {
					t.Log("✅ Hook 值已变化")
				}
			}
		}
	}

	// 检查渲染输出
	rendered := sb.RenderString()
	if strings.Contains(rendered, "Count: 1") {
		t.Log("✅ 渲染输出显示 Count: 1")
	} else if strings.Contains(rendered, "Count: 0") {
		t.Log("❌ 渲染输出仍显示 Count: 0")
	} else {
		t.Logf("⚠️  渲染输出: %s", rendered)
	}
}

// TestEventHandlerInjection 测试事件处理器是否被调用
func TestEventHandlerInjection(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true")

	t.Log("=== 测试 5: 事件处理器是否被调用 ===")

	testApp, err := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()

	// 收集调试日志
	logs := []string{}
	ui.SetDebugLogger(func(msg string) {
		logs = append(logs, msg)
		t.Logf("[DebugLog] %s", msg)
	})

	// 获取初始状态
	initialLogs := len(logs)
	t.Logf("初始日志数: %d", initialLogs)

	// 尝试触发状态更新
	sb.Helper().Tab().Tab().Press(platform.KeyEnter).Process()

	afterLogs := len(logs)
	t.Logf("操作后日志数: %d", afterLogs)

	// 查找关键日志
	onClickFound := false
	setStateFound := false
	getValueFound := false

	for _, log := range logs {
		if strings.Contains(log, "onClick") {
			onClickFound = true
		}
		if strings.Contains(log, "setState") {
			setStateFound = true
		}
		if strings.Contains(log, "getValue") {
			getValueFound = true
		}
	}

	t.Logf("onClick 调用: %v", onClickFound)
	t.Logf("setState 调用: %v", setStateFound)
	t.Logf("getValue 调用: %v", getValueFound)

	if !onClickFound {
		t.Log("❌ onClick 未被调用 - 事件路由可能有问题")
	} else {
		t.Log("✅ onClick 被调用")
	}

	if !setStateFound && onClickFound {
		t.Log("❌ onClick 被调用但 setState 未被调用 - 问题在 setState")
	} else if setStateFound && !getValueFound {
		t.Log("❌ setState 被调用但 getValue 未被调用 - useState 有问题")
	}
}

// TestContextRecreation 测试 Context 是否被重新创建
func TestContextRecreation(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")

	t.Log("=== 测试 6: Context 重新创建问题 ===")

	contextPointers := make(map[string]int)

	// 多次渲染
	for i := 0; i < 5; i++ {
		testApp, err := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
		if err != nil {
			t.Fatalf("TestRun failed: %v", err)
		}

		ctxInfo := ui.DebugContextInfo()
		if ptr, ok := ctxInfo["contextPointer"].(string); ok {
			contextPointers[ptr]++
		}

		testApp.Close()
	}

	t.Logf("Context 指针统计:")
	for ptr, count := range contextPointers {
		t.Logf("  %s: %d 次使用", ptr, count)
	}

	if len(contextPointers) == 1 {
		t.Log("✅ 使用同一个 Context - 正常")
	} else {
		t.Logf("❌ 使用 %d 个不同的 Context - Context 在每次渲染时被重新创建", len(contextPointers))
		t.Log("   这会导致 Hooks 状态丢失")
	}
}

// TestDirectStateMutation 测试直接修改 Hook 值
func TestDirectStateMutation(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")

	t.Log("=== 测试 7: 直接修改 Hook 值 ===")

	testApp, err := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	// 获取当前 Context
	ctx := ui.GetCurrentContext()
	if ctx == nil {
		t.Fatal("无法获取当前 Context")
	}

	t.Logf("初始 Context: ComponentID=%s, Hooks=%d", ctx.ComponentID, len(ctx.Hooks))

	// 直接修改 Hook 值
	if len(ctx.Hooks) > 0 {
		oldValue := ctx.Hooks[0].Value
		t.Logf("修改前: Hooks[0].Value = %v", oldValue)

		// 直接设置新值
		ctx.Hooks[0].Value = 999

		newValue := ctx.Hooks[0].Value
		t.Logf("修改后: Hooks[0].Value = %v", newValue)

		// 获取渲染输出
		sb := testApp.Sandbox()
		rendered := sb.RenderString()

		if strings.Contains(rendered, "999") {
			t.Log("✅ 直接修改 Hook 值后渲染显示新值")
			t.Log("   这证明问题不在渲染层，而在 setState 的传播")
		} else if strings.Contains(rendered, "0") {
			t.Log("❌ 直接修改 Hook 值后仍显示 0")
			t.Log("   这说明渲染使用了缓存的 VNode 或 useState 没有读取 Hook 值")
		}
	}
}

// TestSnapshotWithInternalState 测试快照是否保存内部状态
func TestSnapshotWithInternalState(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")

	t.Log("=== 测试 8: 快照与内部状态 ===")

	testApp, err := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()

	// 创建包含内部状态的快照
	snap1, err := sb.Snapshot(sandbox.SnapshotFull, "initial")
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}

	// 记录内部状态
	ctxInfo1 := ui.DebugContextInfo()
	t.Logf("快照 1 - ContextInfo: %+v", ctxInfo1)

	// 尝试修改状态
	sb.Helper().Tab().Tab().Press(platform.KeyEnter).Process()

	ctxInfo2 := ui.DebugContextInfo()
	t.Logf("操作后 - ContextInfo: %+v", ctxInfo2)

	// 恢复快照
	sb.Restore(snap1)

	ctxInfo3 := ui.DebugContextInfo()
	t.Logf("恢复后 - ContextInfo: %+v", ctxInfo3)

	// 验证
	if ctxInfo1["contextPointer"] == ctxInfo3["contextPointer"] {
		t.Log("✅ 恢复后 Context 指针一致")
	}
}

// TestComprehensiveDebug 综合调试测试
func TestComprehensiveDebug(t *testing.T) {
	t.Setenv("MINT_USE_FIBER", "true")
	t.Setenv("TUI_DEBUG_UI", "true")

	t.Log("╔════════════════════════════════════════════════════════╗")
	t.Log("║         Fiber 问题综合调试测试                          ║")
	t.Log("╚════════════════════════════════════════════════════════╝")

	// 启用详细日志
	ui.SetDebugLogger(func(msg string) {
		t.Logf("[LOG] %s", msg)
	})

	testApp, err := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
	if err != nil {
		t.Fatalf("TestRun failed: %v", err)
	}
	defer testApp.Close()

	sb := testApp.Sandbox()

	// 1. 初始状态分析
	t.Log("\n--- 阶段 1: 初始状态分析 ---")

	ctxInfo := ui.DebugContextInfo()
	dumpContextState(t, ctxInfo)

	initialRender := sb.RenderString()
	t.Logf("初始渲染输出:\n%s", initialRender)

	// 2. 执行操作
	t.Log("\n--- 阶段 2: 执行点击操作 ---")

	t.Log("操作前: HookIndex=%d", ctxInfo["hookIndex"])
	sb.Helper().Tab().Tab().Press(platform.KeyEnter).Process()

	afterCtxInfo := ui.DebugContextInfo()
	t.Log("操作后: HookIndex=%d", afterCtxInfo["hookIndex"])

	// 3. 对比差异
	t.Log("\n--- 阶段 3: 状态对比 ---")

	compareContextStates(t, ctxInfo, afterCtxInfo)

	// 4. 验证渲染输出
	t.Log("\n--- 阶段 4: 渲染输出验证 ---")

	afterRender := sb.RenderString()
	t.Logf("操作后渲染输出:\n%s", afterRender)

	// 检查关键指标
	checkSuccessIndicators(t, initialRender, afterRender, ctxInfo, afterCtxInfo)
}

// dumpContextState 打印 Context 状态详情
func dumpContextState(t *testing.T, info map[string]interface{}) {
	t.Log("Context 状态:")

	if hasCtx, ok := info["hasContext"].(bool); ok {
		t.Logf("  有 Context: %v", hasCtx)
	}

	if compID, ok := info["componentID"].(string); ok {
		t.Logf("  ComponentID: %s", compID)
	}

	if hookIndex, ok := info["hookIndex"].(int); ok {
		t.Logf("  HookIndex: %d", hookIndex)
	}

	if hooks, ok := info["hooks"].([]map[string]interface{}); ok {
		t.Logf("  Hooks 数量: %d", len(hooks))
		for i, hook := range hooks {
			if typ, ok := hook["type"].(string); ok {
				if val, ok := hook["value"].(int); ok {
					t.Logf("    [%d] %s = %d", i, typ, val)
				}
			}
		}
	}
}

// compareContextStates 对比两次 Context 状态
func compareContextStates(t *testing.T, before, after map[string]interface{}) {
	t.Log("Context 变化:")

	// 检查 Context 指针
	beforePtr, _ := before["contextPointer"].(string)
	afterPtr, _ := after["contextPointer"].(string)

	if beforePtr == afterPtr {
		t.Log("  ✅ Context 指针相同 - Context 被复用")
	} else {
		t.Logf("  ❌ Context 指针不同!")
		t.Logf("     Before: %s", beforePtr)
		t.Logf("     After:  %s", afterPtr)
	}

	// 检查 Hook 值
	beforeHooks, _ := before["hooks"].([]map[string]interface{})
	afterHooks, _ := after["hooks"].([]map[string]interface{})

	if len(beforeHooks) > 0 && len(afterHooks) > 0 {
		beforeVal := beforeHooks[0]["value"]
		afterVal := afterHooks[0]["value"]

		if beforeVal == afterVal {
			t.Logf("  ❌ Hook 值未变化: %v", beforeVal)
		} else {
			t.Logf("  ✅ Hook 值已变化: %v -> %v", beforeVal, afterVal)
		}
	}
}

// checkSuccessIndicators 检查成功指标
func checkSuccessIndicators(t *testing.T, initialRender, afterRender string,
	beforeCtx, afterCtx map[string]interface{}) {

	success := true

	// 检查 1: Context 复用
	if beforeCtx["contextPointer"] == afterCtx["contextPointer"] {
		t.Log("✅ 指标 1: Context 复用 - 通过")
	} else {
		t.Log("❌ 指标 1: Context 复用 - 失败 (每次渲染创建新 Context)")
		success = false
	}

	// 检查 2: 状态更新
	beforeHooks, _ := beforeCtx["hooks"].([]map[string]interface{})
	afterHooks, _ := afterCtx["hooks"].([]map[string]interface{})

	if len(beforeHooks) > 0 && len(afterHooks) > 0 {
		if beforeHooks[0]["value"] != afterHooks[0]["value"] {
			t.Log("✅ 指标 2: Hook 值变化 - 通过")
		} else {
			t.Log("❌ 指标 2: Hook 值变化 - 失败 (setState 未生效)")
			success = false
		}
	}

	// 检查 3: 渲染输出
	if strings.Contains(afterRender, "Count: 1") {
		t.Log("✅ 指标 3: 渲染输出更新 - 通过")
	} else if strings.Contains(afterRender, "Count: 0") {
		t.Log("❌ 指标 3: 渲染输出更新 - 失败 (仍显示 0)")
		success = false
	}

	// 总结
	t.Log("\n═══════════════════════════════════════════════════════════")
	if success {
		t.Log("总体: ✅ 所有关键指标通过 - Fiber 状态管理正常")
	} else {
		t.Log("总体: ❌ 存在问题 - 需要进一步调试")
		t.Log("\n可能的原因:")
		t.Log("1. Context 在每次渲染时被重新创建")
		t.Log("2. Hooks 存储在新 Context 中，导致读取旧值")
		t.Log("3. onClick 事件未被正确路由")
		t.Log("4. useState 没有读取 Hook 的最新值")
	}
}
