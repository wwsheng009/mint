# Fiber Integration Bug 修复报告

**日期**: 2025年
**问题**: setState 更新后 UI 不刷新 (计数器始终显示 0)
**状态**: ✅ 已修复

---

## 问题概述

### 症状
- Fiber 模式下，点击按钮后 `setState` 被调用
- 状态值已更新 (hook.Value = 1)
- 但 UI 仍显示旧值 (Count: 0)

### 初始假设
1. onClick 事件未触发 ❌
2. setState 未正确更新 hook ❌
3. 组件未重新渲染 ❌
4. VNode 缓存/复用问题 ✅ **(真正原因)**

---

## 调试过程

### 第一步：验证事件流程

使用 `RunTest` + `InjectSpecialKey` 模拟用户输入：

```go
testApp, err := ui.RunTest(DebugCounter, ui.WithWidth(40), ui.WithHeight(12))
testApp.InjectSpecialKey(platform.KeyTab)  // 切换焦点
testApp.InjectSpecialKey(platform.KeyEnter) // 触发点击
```

**发现**: 事件正确到达 `HandleEvent`，onClick 被调用

### 第二步：添加调试日志

在关键位置添加日志追踪：

```go
// hooks.go - useState
func useState(initial interface{}) (interface{}, func(interface{})) {
    ...
    if os.Getenv("TUI_DEBUG_UI") == "true" {
        fmt.Fprintf(os.Stderr, "useState: componentID=%s, hookIndex=%d, value=%v\n",
            ctx.ComponentID, hookIndex, currentValue)
    }
    ...
}

// main.go - DebugCounter
func DebugCounter() ui.VNode {
    count, setCount, _, hookIndex := ui.UseStateIntWithDebug(0)

    if os.Getenv("TUI_DEBUG_UI") == "true" {
        fmt.Fprintf(os.Stderr, "[DebugCounter] Using count=%d, hookIndex=%d\n", count, hookIndex)
    }

    countTextStr := fmt.Sprintf("Count: %d (hookIndex=%d)", count, hookIndex)
    if os.Getenv("TUI_DEBUG_UI") == "true" {
        fmt.Fprintf(os.Stderr, "[DebugCounter] Creating TextVNode with content: %s\n", countTextStr)
    }
    ...
}
```

### 第三步：日志分析

**关键日志输出**:
```
[DebugCounter] Using count=1, hookIndex=0                          ← 组件获取到正确值
[DebugCounter] Created TextVNode ptr=0xc0000bdb80, content=Count: 1 ← 创建了正确的 VNode
renderVNodeFiber: type=*ui.TextVNode, node=&{content:Count: 0...}   ← 但渲染的是旧 VNode!
```

**结论**: 组件函数正确创建了新 VNode (`Count: 1`)，但 Fiber 渲染的是旧 VNode (`Count: 0`)

### 第四步：定位 Fiber 协调器问题

检查 `reconciler.go` 的渲染流程：

```go
func (r *Reconciler) Render(...) {
    r.prepareFreshStack(renderFunc)  // 创建 workInProgress 树
    r.workLoopSync()                 // 处理 workInProgress
    r.CommitRoot()                   // 提交变更
}

func (r *Reconciler) CommitRoot() {
    r.renderFiberToBuffer(r.root, 0, 0, r.buffer)  // ← 问题：使用的是 r.root!
}
```

**问题**: `workLoopSync()` 处理完 `workInProgress` 树后，直接设置为 `nil`，没有与 `r.root` 交换！

### 第五步：修复

```go
func (r *Reconciler) workLoopSync() {
    ...
    r.performUnitOfWork(r.workInProgress)

    // 关键修复：交换 workInProgress 和 root 树 (双缓冲)
    r.root = r.workInProgress
    r.workInProgress = nil
}
```

---

## 如何使用 Sandbox 进行调试

### 1. RunTest API - 创建可测试的应用

```go
testApp, err := ui.RunTest(ComponentFunc,
    ui.WithWidth(40),
    ui.WithHeight(12),
)
defer testApp.Close()
```

### 2. 事件注入

```go
// 注入特殊键
testApp.InjectSpecialKey(platform.KeyTab)
testApp.InjectSpecialKey(platform.KeyEnter)

// 注入字符键
testApp.InjectKey('q')

// 注入鼠标事件 (如果需要)
```

### 3. 获取渲染结果

```go
// 获取渲染后的字符串
rendered := testApp.GetRenderString()

// 获取 buffer
buf := testApp.GetBuffer()

// 强制立即渲染
testApp.GetFrameworkApp().ForceRenderNow()
```

### 4. 调试技巧

#### 技巧 1: 环境变量控制日志
```go
t.Setenv("TUI_DEBUG_UI", "true")
```

#### 技巧 2: 在组件中添加日志
```go
func MyComponent() ui.VNode {
    count, setCount := ui.UseStateInt(0)

    if os.Getenv("TUI_DEBUG_UI") == "true" {
        fmt.Fprintf(os.Stderr, "[Component] count=%d\n", count)
    }
    ...
}
```

#### 技巧 3: 追踪 VNode 生命周期
```go
// 在创建时记录
countText := ui.NewTextBuilder(fmt.Sprintf("Count: %d", count)).Build()
if os.Getenv("TUI_DEBUG_UI") == "true" {
    fmt.Fprintf(os.Stderr, "[Create] TextVNode ptr=%p, content=%s\n",
        countText, countText.(*ui.TextVNode).Content())
}
```

---

## 架构理解

### Fiber 双缓冲机制

```
┌─────────────────────────────────────────────────────┐
│  Render Phase (workLoopSync)                        │
│                                                     │
│  current tree          workInProgress tree          │
│  ┌──────┐              ┌──────┐                     │
│  │ Fiber│ ───clone──→ │ Fiber│                     │
│  └──────┘              └──────┘                     │
│     ↑                       ↓                       │
│     └──────────swap──────────┘                     │
│           (修复点: r.root = r.workInProgress)      │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│  Commit Phase (CommitRoot)                          │
│                                                     │
│  使用 r.root 渲染到 buffer                          │
│  (现在 r.root 是更新后的树)                         │
└─────────────────────────────────────────────────────┘
```

### VNode 更新流程

```
1. setState() → scheduleRender() → MarkDirty()
2. 重新渲染 → 调用组件函数 → 创建新的 VNode
3. reconciler.diff() → cloneExistingFiber() → 更新 fiber.VNode
4. workLoopSync() → 交换 workInProgress ↔ root
5. CommitRoot() → 使用 root 树渲染
```

---

## 测试结果

### 修复前
```
Count: 0 → 点击 → Count: 0  ❌ (不更新)
```

### 修复后
```
Count: 0 → 点击 → Count: 1  ✅ (正确更新)
```

### 测试通过率
- `run_test.go`: 7/7 ✅
- `full_app_test.go`: 2/2 ✅
- 总计: 9/9 ✅

---

## 经验总结

1. **系统性调试**: 从事件流程 → 状态更新 → 组件渲染 → VNode 创建 → Fiber 协调，逐层追踪
2. **日志的重要性**: 在关键路径添加日志是定位问题的最快方法
3. **理解架构**: 深入理解 Fiber 双缓冲机制才能找到根本原因
4. **测试驱动**: 使用 RunTest 可以自动化测试，不需要手动交互

---

## 相关文件

- `ui/reconciler.go` - Fiber 协调器 (修复位置)
- `ui/hooks.go` - useState 实现
- `ui/app.go` - RunTest API
- `examples/fiber_counter/run_test.go` - 自动化测试
- `examples/fiber_counter/main.go` - 示例组件
