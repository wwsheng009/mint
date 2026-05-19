# 使用 Sandbox 调试 Fiber 集成指南

> 针对问题: [Fiber架构集成状态与问题分析](/docsArchive/issue/2026-02-01-fiber-integration-status.md)

## 概述

本文档展示如何使用 `sandbox` 模块来测试和调试 Fiber 集成问题。

## 问题回顾

### 当前问题
1. **状态更新后UI不刷新** - `setState`被调用，但`useState`返回旧值
2. **按钮点击可能无响应** - `HandleEvent`可能未被调用

### 调试目标
- 验证 `setState` 是否正确更新 hook.Value
- 验证重新渲染时 `useState` 是否返回新值
- 验证事件是否正确路由到组件

## 快速开始

### 1. 创建测试环境

```go
import (
    "testing"
    "github.com/wwsheng009/mint/ui"
)

func TestCounterWithSandbox(t *testing.T) {
    // 创建测试应用
    testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 12))
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    sb := testApp.Sandbox()
}
```

### 2. 验证初始渲染

```go
func TestInitialRender(t *testing.T) {
    testApp, _ := ui.TestRun(Counter, ui.TestWithSize(40, 12))
    defer testApp.Close()

    sb := testApp.Sandbox()

    // 获取渲染输出
    rendered := sb.RenderString()

    // 验证初始计数
    if !strings.Contains(rendered, "Count: 0") {
        t.Errorf("Expected 'Count: 0', got: %s", rendered)
    }
}
```

### 3. 测试状态更新

```go
func TestStateUpdate(t *testing.T) {
    testApp, _ := ui.TestRun(Counter, ui.TestWithSize(40, 12))
    defer testApp.Close()

    sb := testApp.Sandbox()

    // 模拟点击 + 按钮
    // Tab 移动焦点，Enter 触发点击
    sb.Helper().
        Tab().
        Tab().
        Press(platform.KeyEnter).
        Process()

    // 验证状态更新
    if err := sb.AssertRender("Count: 1"); err != nil {
        t.Errorf("State not updated: %v", err)
    }
}
```

## 调试 useState 问题

### 问题: setState 调用但 useState 返回旧值

创建一个专门的测试来验证 hooks 状态：

```go
func TestUseStateUpdate(t *testing.T) {
    testApp, _ := ui.TestRun(Counter, ui.TestWithSize(40, 12))
    defer testApp.Close()

    sb := testApp.Sandbox()

    // 1. 创建初始快照
    snap1, _ := sb.Snapshot(sandbox.SnapshotFull, "initial")
    t.Logf("Initial snapshot: %s", snap1.Metadata.ID)

    // 2. 触发状态更新
    sb.Helper().
        Tab().Tab().
        Press(platform.KeyEnter).
        Process()

    // 3. 获取渲染输出
    rendered := sb.RenderString()
    t.Logf("After click:\n%s", rendered)

    // 4. 验证状态
    if strings.Contains(rendered, "Count: 1") {
        t.Log("✅ useState 返回新值 - 状态更新正常")
    } else if strings.Contains(rendered, "Count: 0") {
        t.Log("❌ useState 返回旧值 - setState 未生效")

        // 调试: 检查 hooks
        sb.Helper().RenderString()
    } else {
        t.Log("❓ 未找到计数器输出")
    }
}
```

## 调试事件处理

### 问题: HandleEvent 未被调用

```go
func TestEventHandling(t *testing.T) {
    testApp, _ := ui.TestRun(Counter, ui.TestWithSize(40, 12))
    defer testApp.Close()

    sb := testApp.Sandbox()

    // 设置事件处理器来捕获调用
    eventCaptured := false
    sb.SetEventHandler(func(event platform.RawInput) error {
        t.Logf("Event received: Type=%d, Key=%c",
            event.Type, event.Key)
        eventCaptured = true
        return nil
    })

    // 注入事件
    sb.InjectKey('a')
    sb.ProcessEvents()

    if !eventCaptured {
        t.Error("❌ EventHandler 未被调用")
    } else {
        t.Log("✅ EventHandler 正常工作")
    }
}
```

## 使用快照进行状态对比

```go
func TestSnapshotComparison(t *testing.T) {
    testApp, _ := ui.TestRun(Counter, ui.TestWithSize(40, 12))
    defer testApp.Close()

    sb := testApp.Sandbox()

    // 初始状态快照
    initialSnap, _ := sb.Snapshot(sandbox.SnapshotFull, "initial")

    // 执行操作
    sb.Helper().
        Tab().Tab().
        Press(platform.KeyEnter).
        Process()

    // 操作后快照
    afterSnap, _ := sb.Snapshot(sandbox.SnapshotFull, "after-click")

    // 对比
    t.Logf("Initial events: %d", len(initialSnap.Events))
    t.Logf("After events: %d", len(afterSnap.Events))

    // 恢复并验证
    sb.Restore(initialSnap)
    restored := sb.RenderString()

    if strings.Contains(restored, "Count: 0") {
        t.Log("✅ 快照恢复成功")
    }
}
```

## 链式测试 API

Sandbox 提供了链式 API 用于编写更清晰的测试：

```go
func TestChainAPI(t *testing.T) {
    testApp, _ := ui.TestRun(Counter, ui.TestWithSize(40, 12))
    defer testApp.Close()

    result := testApp.Helper().
        // 初始验证
        AssertRender("Count: 0").
        AssertRender("Mint UI Counter Demo").

        // 移动焦点
        Tab().
        Process().
        AssertRender("-").

        // 点击 + 按钮
        Tab().
        Press(platform.KeyEnter).
        Process().
        AssertRender("Count: 1").

        // 再次点击
        Press(platform.KeyEnter).
        Process().
        AssertRender("Count: 2").

        // 获取结果
        Result()

    if !result.OK() {
        t.Errorf("Test failed: %v", result.Error())
    }
}
```

## 内存安全测试

验证有界队列是否正常工作：

```go
func TestMemorySafety(t *testing.T) {
    config := sandbox.MockConfig()
    config.Event.QueueMaxSize = 100  // 限制队列大小

    testApp, _ := ui.TestRunWithConfig(Counter, config)
    defer testApp.Close()

    sb := testApp.Sandbox().(*mock.MockSandbox)

    // 注入大量事件
    for i := 0; i < 1000; i++ {
        sb.InjectKey('a')
    }

    // 检查队列统计
    stats := sb.QueueStats()
    t.Logf("Queue stats: Length=%d, Memory=%d, Evicted=%d",
        stats.Length, stats.MemoryUsed, stats.EvictCount)

    if stats.Length <= 100 {
        t.Log("✅ 队列大小限制正常工作")
    }
}
```

## 完整测试示例

```go
// examples/counter/fiber_test.go
package main

import (
    "strings"
    "testing"

    "github.com/wwsheng009/mint/runtime/platform"
    "github.com/wwsheng009/mint/ui"
)

// TestFiberCounterInitial 渲染测试
func TestFiberCounterInitial(t *testing.T) {
    testApp, err := ui.TestRun(Counter, ui.TestWithSize(40, 12))
    if err != nil {
        t.Fatal(err)
    }
    defer testApp.Close()

    sb := testApp.Sandbox()
    rendered := sb.RenderString()

    // 验证关键元素
    checks := []string{
        "Mint UI Counter Demo",
        "Count: 0",
        "-",
        "+",
    }

    for _, check := range checks {
        if !strings.Contains(rendered, check) {
            t.Errorf("Missing expected text: %q", check)
        }
    }
}

// TestFiberCounterIncrement 测试递增按钮
func TestFiberCounterIncrement(t *testing.T) {
    testApp, _ := ui.TestRun(Counter, ui.TestWithSize(40, 12))
    defer testApp.Close()

    sb := testApp.Sandbox()

    // 点击 + 按钮两次
    for i := 1; i <= 2; i++ {
        sb.Helper().
            Tab().Tab().
            Press(platform.KeyEnter).
            Process()

        expected := fmt.Sprintf("Count: %d", i)
        if err := sb.AssertRender(expected); err != nil {
            t.Errorf("After %d clicks: %v", i, err)
        }
    }
}

// TestFiberCounterDecrement 测试递减按钮
func TestFiberCounterDecrement(t *testing.T) {
    testApp, _ := ui.TestRun(Counter, ui.TestWithSize(40, 12))
    defer testApp.Close()

    sb := testApp.Sandbox()

    // 点击 - 按钮
    sb.Helper().
        Tab().
        Press(platform.KeyEnter).
        Process()

    if err := sb.AssertRender("Count: -1"); err != nil {
        t.Errorf("Decrement failed: %v", err)
    }
}

// TestFiberCounterSnapshot 快照测试
func TestFiberCounterSnapshot(t *testing.T) {
    testApp, _ := ui.TestRun(Counter, ui.TestWithSize(40, 12))
    defer testApp.Close()

    sb := testApp.Sandbox()

    // 保存初始状态
    snap, _ := sb.Sandbox().Snapshot(sandbox.SnapshotStandard, "initial")

    // 修改状态
    sb.Helper().Tab().Tab().Press(platform.KeyEnter).Process()

    // 验证状态已改变
    if err := sb.AssertRender("Count: 1"); err == nil {
        t.Log("State changed after click")
    }

    // 恢复快照
    sb.Restore(snap)

    // 验证恢复到初始状态
    if err := sb.AssertRender("Count: 0"); err != nil {
        t.Errorf("Restore failed: %v", err)
    }
}
```

## 运行测试

```bash
# 运行所有测试
go test ./examples/counter/

# 运行特定测试
go test ./examples/counter/ -run TestFiberCounter

# 启用调试输出
TUI_DEBUG_UI=true go test ./examples/counter/ -v

# 运行所有 sandbox 测试
go test ./sandbox/...
```

## 调试技巧

### 1. 打印渲染输出

```go
t.Log("=== Rendered Output ===")
t.Log(sb.RenderString())
```

### 2. 检查队列状态

```go
stats := sb.(*mock.MockSandbox).QueueStats()
t.Logf("Queue: %d events, %d bytes evicted",
    stats.Length, stats.EvictCount)
```

### 3. 列出快照

```go
snapshots := sb.ListSnapshots()
for _, snap := range snapshots {
    t.Logf("Snapshot: %s (Level=%d)", snap.ID, snap.Level)
}
```

### 4. 录制事件序列

```go
config := sandbox.MockConfig()
config.Event.RecordEnabled = true
config.Event.RecordMaxLen = 100

testApp, _ := ui.TestRunWithConfig(Counter, config)

// ... 执行操作 ...

// 获取录制的事件
events := sb.(*mock.MockSandbox).recorder.Events()
t.Logf("Recorded %d events", len(events))
```

## 下一步

1. 运行上述测试，确认问题是否可重现
2. 使用快照功能对比状态更新前后的差异
3. 添加调试日志到 `useState` 和 `setState`
4. 使用 Sandbox 的事件注入功能模拟各种交互场景
