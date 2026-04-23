# Fiber 问题复现与 Sandbox 测试指南

## 问题确认

通过运行 `examples/fiber_counter`，我们已确认了 Fiber 集成问题：

### 观察到的问题

1. **useState 始终返回初始值**
   ```
   [DEBUG] UseStateIntWithDebug: initial=0, hookIndex=0
   ```
   每次重新渲染都显示 `initial=0`，说明 useState 没有返回更新后的值。

2. **onClick 事件未触发**
   - 没有看到 `[DEBUG] onClick` 输出
   - 没有看到 `[DEBUG] setState` 输出
   - 按钮点击没有触发任何状态更新

## 使用 Sandbox 测试这个问题

### 1. 基础渲染测试

```go
func TestFiberCounterRender(t *testing.T) {
    t.Setenv("MINT_USE_FIBER", "true")
    t.Setenv("TUI_DEBUG_UI", "true")

    testApp, _ := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
    defer testApp.Close()

    sb := testApp.Sandbox()
    rendered := sb.RenderString()

    // 验证初始渲染
    if !strings.Contains(rendered, "Count: 0") {
        t.Errorf("Expected 'Count: 0', got: %s", rendered)
    }
}
```

### 2. 状态更新测试（核心测试）

```go
func TestFiberCounterIncrement(t *testing.T) {
    t.Setenv("MINT_USE_FIBER", "true")
    t.Setenv("TUI_DEBUG_UI", "true")

    testApp, _ := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
    defer testApp.Close()

    sb := testApp.Sandbox()

    // 记录初始状态
    initial := sb.RenderString()
    t.Logf("Initial:\n%s", initial)

    // 模拟点击 + 按钮
    sb.Helper().
        Tab().Tab().
        Press(platform.KeyEnter).
        Process()

    // 检查状态是否更新
    after := sb.RenderString()
    t.Logf("After click:\n%s", after)

    if strings.Contains(after, "Count: 1") {
        t.Log("✅ 状态更新成功")
    } else if strings.Contains(after, "Count: 0") {
        t.Log("❌ 状态未更新 - 问题已复现")
    }
}
```

### 3. 事件处理测试

```go
func TestFiberEventHandler(t *testing.T) {
    t.Setenv("MINT_USE_FIBER", "true")

    testApp, _ := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
    defer testApp.Close()

    sb := testApp.Sandbox()

    // 设置事件处理器
    onClickCalled := false
    sb.SetEventHandler(func(event platform.RawInput) error {
        if event.Type == platform.InputMouse && event.MouseAction == platform.MousePress {
            onClickCalled = true
            t.Log("✅ 鼠标点击事件被接收")
        }
        return nil
    })

    // 注入点击事件
    sb.InjectMouse(5, 4, platform.MouseLeft, platform.MousePress)
    sb.ProcessEvents()

    if !onClickCalled {
        t.Error("❌ 点击事件未触发")
    }
}
```

### 4. 快照对比测试

```go
func TestFiberSnapshotComparison(t *testing.T) {
    t.Setenv("MINT_USE_FIBER", "true")

    testApp, _ := ui.TestRun(DebugCounter, ui.TestWithSize(40, 12))
    defer testApp.Close()

    sb := testApp.Sandbox()

    // 创建初始快照
    snap1, _ := sb.Snapshot(sandbox.SnapshotFull, "initial")

    // 执行操作
    sb.Helper().Tab().Tab().Press(platform.KeyEnter).Process()

    // 创建操作后快照
    snap2, _ := sb.Snapshot(sandbox.SnapshotFull, "after-click")

    // 对比差异
    t.Logf("Initial events: %d", len(snap1.Events))
    t.Logf("After events: %d", len(snap2.Events))

    // 恢复初始状态
    sb.Restore(snap1)
    restored := sb.RenderString()

    if strings.Contains(restored, "Count: 0") {
        t.Log("✅ 快照恢复正常")
    }
}
```

## 运行测试

```bash
# 运行 Fiber 模式程序（手动测试）
MINT_USE_FIBER=true TUI_DEBUG_UI=true go run ./examples/fiber_counter/main.go

# 运行自动化测试
go test ./examples/fiber_counter/ -v -run TestFiber

# 运行特定测试
go test ./examples/counter/ -v -run TestFiberCounterIncrement
```

## 调试步骤

### 步骤 1: 确认 useState 行为

运行测试并观察 `getOrCreateHook` 的输出：
- 如果总是显示 `EXISTS, hookIndex=0, value=0`
- 说明 hook 被正确创建和存储
- 但返回的值始终是初始值

### 步骤 2: 确认 setState 是否被调用

在 DebugCounter 中，我们添加了详细的日志：
```go
OnClick(func() {
    fmt.Fprintf(os.Stderr, "[DEBUG] onClick: increment called, current count=%d\n", count)
    setCount(func(c int) int { ... })
})
```

如果没有看到此输出，说明 onClick 根本没有被调用。

### 步骤 3: 检查事件路由

使用 Sandbox 的事件注入功能：
```go
sb.InjectMouse(5, 4, platform.MouseLeft, platform.MousePress)
sb.ProcessEvents()
```

然后检查 EventHandler 是否被触发。

## 问题定位

基于当前观察，问题可能在于：

1. **事件路由问题**
   - HandleEvent 在 Fiber 模式下可能没有正确路由到按钮
   - 焦点系统可能没有正确工作

2. **Context 持久化问题**
   - 组件重新渲染时，Context 被重新创建
   - Hooks 存储在新的 Context 中，而不是复用旧的

3. **Hook 指针失效**
   - Hooks 切片扩容导致指针失效
   - onClick 中捕获的 hook 指针指向旧内存

## 下一步调试

1. 在 `ui/interaction_state.go` 中添加调试日志
2. 在 `internal/render/declarative_node_event.go` 中添加调试日志
3. 在 `internal/reconciler/reconciler.go` 的 `prepareFreshStack` 中添加调试日志
4. 使用 Sandbox 的快照功能对比每次渲染前后的差异

## 相关文件

- `examples/fiber_counter/main.go` - 调试版本的计数器
- `ui/hooks.go` - UseStateIntWithDebug 函数
- `examples/counter/fiber_test.go` - Sandbox 测试用例
- `docs/sandbox/fiber_debug/FIBER_TESTING_GUIDE.md` - 完整测试指南
