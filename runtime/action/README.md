# Action System (V3)

Action 处理和执行系统。

## 职责

- **Action 类型定义**：语义化的事件类型（导航、编辑、表单等）
- **Action 分发**：将 Action 路由到正确的处理器
- **Action 处理**：组件实现 Target 接口响应 Action
- **复合 Action**：支持批量并发/顺序执行多个 Action
- **Action Logging**：可记录和调试 Action 分发历史

## 纯 Go 约束

此目录必须保持纯 Go 实现，不能依赖：
- Bubble Tea
- DSL 解析器
- 具体组件
- lipgloss

## 核心概念

### Action 是语义化事件

Action 不是原始按键，而是**语义化的用户意图**：

```go
// 平台产生原始输入
RawInput: KeyEvent{Key: KeyTab}

// Runtime 转换为语义化 Action
Action: Action{Type: ActionNavigateNext}

// 组件只处理 Action，不关心原始输入
Component.HandleAction(ActionNavigateNext)
```

**设计原则**：所有状态变化必须能追溯到 Action，便于调试、测试和回放。

### Action 类型

Action 类型分为以下类别：

| 类别 | 示例 Action |
|------|------------|
| **导航** | `NavigateNext`, `NavigatePrev`, `NavigateUp`, `NavigateDown`, `NavigatePageUp` |
| **编辑** | `InputChar`, `Backspace`, `DeleteChar`, `CursorLeft`, `CursorRight`, `SelectAll` |
| **表单** | `Submit`, `Cancel`, `Validate`, `Reset`, `Clear` |
| **选择** | `SelectItem`, `DeselectItem`, `ToggleSelect`, `SelectRange` |
| **鼠标** | `MouseClick`, `MouseDoubleClick`, `MouseDrag`, `MouseWheel` |
| **视图** | `Scroll`, `ScrollUp`, `ScrollDown`, `ZoomIn`, `ZoomOut` |
| **窗口** | `Quit`, `Close`, `Maximize`, `Fullscreen` |
| **系统** | `Copy`, `Cut`, `Paste`, `Undo`, `Redo`, `Search` |
| **AI** | `AIInspect`, `AIFind`, `AIQuery`, `AIDispatch` |

完整定义见 `action.go`。

### Action 结构

```go
type Action struct {
    Type      ActionType   // Action 类型
    Payload   interface{}  // 携带的数据（字符、文本、方向等）
    Source    string       // 来源组件 ID
    Target    string       // 目标组件 ID
    Timestamp time.Time    // 时间戳
}
```

使用链式 API 构建 Action：

```go
action := action.NewAction(action.ActionInputChar).
    WithPayload('A').
    WithSource("editor").
    WithTarget("input-field")
```

### Dispatcher（分发器）

`Dispatcher` 负责将 Action 路由到处理器：

```go
type Dispatcher struct {
    targets        map[string]Target          // 注册的目标组件
    globalHandlers map[ActionType][]Handler   // 全局处理器
    defaultHandler Handler                    // 默认处理器
    log            bool                       // 日志开关
    logEntries     []LogEntry                 // 日志历史
}
```

**分发顺序**：

1. 全局处理器（`Subscribe` 订阅）
2. 指定目标（`Action.Target`）
3. 默认处理器（如果没有被处理）

```go
// dispatch 流程
if dispatcher.dispatchGlobal(action) {
    // 全局处理器处理了
} else if action.Target != "" {
    return dispatcher.dispatchToTarget(action, action.Target)
} else if defaultHandler != nil {
    return defaultHandler(action)
}
```

### Target（处理目标）

组件通过实现 `Target` 接口接收 Action：

```go
// 在 target.go 中定义
type Target interface {
    ID() string
    HandleAction(a *Action) bool  // 返回 true 表示已处理
}
```

### 复合 Action

支持批量执行多个 Action：

| 模式 | 说明 | 实现方式 |
|------|------|----------|
| `Concurrent` | 并发执行 | `Batch()` 或 `NewCompositeAction(Concurrent)` |
| `Sequential` | 顺序执行 | `Sequence()` 或 `NewCompositeAction(Sequential)` |

支持高级功能：
- **Worker Pool**：限制并发数
- **Retry**：失败重试
- **Timeout**：超时控制
- **Fallback**：主操作失败时执行备用操作

## 使用示例

### 基本 Action 创建

```go
import "github.com/wwsheng009/mint/runtime/action"

// 创建简单 Action
navigateNext := action.NewAction(action.ActionNavigateNext)

// 创建带数据的 Action
inputChar := action.NewAction(action.ActionInputChar).
    WithPayload('A').
    WithSource("keyboard")

// 创建带目标的 Action
submit := action.NewAction(action.ActionSubmit).
    WithTarget("form1")
```

### 使用 Dispatcher

```go
// 创建分发器
d := action.NewDispatcher()

// 注册目标组件
d.Register(myComponent)

// 订阅全局处理器（例如：快捷键）
d.Subscribe(action.ActionQuit, func(a *action.Action) bool {
    // 全局处理退出
    return true // 表示已处理
})

// 分发 Action
d.Dispatch(submit)
```

### 实现 Target 接口

```go
type MyButton struct {
    id    string
    label string
    // ...
}

func (b *MyButton) ID() string {
    return b.id
}

func (b *MyButton) HandleAction(a *action.Action) bool {
    switch a.Type {
    case action.ActionMouseClick:
        b.onClick()
        return true
    case action.ActionSubmit:
        b.onSubmit()
        return true
    }
    return false // 未处理
}
```

### 复合 Action：并发执行

```go
// 并发执行多个 Action
batch := action.Batch(
    action.ActionFunc(func(ctx context.Context) action.ActionResult {
        return action.OKAction
    }),
    action.ActionFunc(func(ctx context.Context) action.ActionResult {
        return action.MessageAction("完成")
    }),
)

result := batch.Execute(context.Background())
```

### 复合 Action：顺序执行

```go
// 顺序执行（前一个失败会中断）
sequence := action.Sequence(
    action.ActionFunc(loadData),
    action.ActionFunc(validateData),
    action.ActionFunc(saveData),
)

result := sequence.Execute(context.Background())
``}
```

### 带回调的复合 Action

```go
action.BatchWithCallback(func(results []action.ActionResult) {
    for _, r := range results {
        fmt.Println(r.Message)
    }
}, action1, action2, action3)
```

### 重试和超时

```go
// 重试 3 次，每次间隔 1 秒
retryAction := action.NewRetryAction(
    action.ActionFunc(fetchData),
    3,                    // maxRetries
    time.Second,           // delay
)

// 超时控制
timeoutAction := action.NewTimeoutAction(
    action.ActionFunc(longRunningTask),
    5 * time.Second,
)
```

### Worker Pool（限制并发）

```go
// 创建工作池（最多 5 个并发）
pool := action.NewWorkerPool(5)
pool.Start()
defer pool.Stop()

// 提交任务
for i := 0; i < 100; i++ {
    task := action.ActionFunc(func(ctx context.Context) action.ActionResult {
        return processItem(i)
    })
    pool.Submit(task)
}
```

### 限制并发的并行执行

```go
// 限制并发数为 3
parallel := action.ParallelWithLimit(3,
    action.ActionFunc(task1),
    action.ActionFunc(task2),
    action.ActionFunc(task3),
    action.ActionFunc(task4),
    // ...
)
```

### Action Logging

```go
// 启用日志
d.EnableLog(true)

// 分发后查看日志
d.PrintLog()

// 获取日志条目
entries := d.GetLog()
for _, entry := range entries {
    fmt.Printf("%s -> %v\n", entry.Action, entry.Handled)
}

// 清空日志
d.ClearLog()
```

## 核心类型

| 类型 | 说明 |
|------|------|
| `Action` | Action 主体 |
| `ActionType` | Action 类型枚举（string） |
| `Dispatcher` | Action 分发器 |
| `Handler` | Action 处理器函数类型：`func(*Action) bool` |
| `Target` | Action 处理目标接口 |
| `CompositeAction` | 复合 Action（并发/顺序） |
| `ActionResult` | Action 执行结果 |
| `ActionHandler` | 可执行的 Action 接口 |
| `Mode` | 执行模式（并发/顺序） |
| `WorkerPool` | 工作池 |
| `RetryAction` | 带重试的 Action |
| `TimeoutAction` | 带超时的 Action |
| `FallbackAction` | 带回退的 Action |

## 文件结构

| 文件 | 说明 |
|------|------|
| `action.go` | Action 类型和常量定义（V3 架构） |
| `dispatcher.go` | Dispatcher 分发器实现 |
| `target.go` | Target 接口定义 |
| `composite.go` | 复合 Action 系统（并发/顺序执行） |
| `errors.go` | 错误类型定义（如果有） |
| `integration_test.go` | 集成测试 |

## 最佳实践

### 1. 语义化 Action 命名

```go
// 推荐（语义化）
ActionNavigateNext
ActionNavigatePrev
ActionSubmit

// 不推荐（平台相关）
ActionTabKey     // ❌ 应该是 ActionNavigateNext
ActionEnterKey   // ❌ 应该是 ActionSubmit
```

### 2. Action 应该是可序列化的

用于时间旅行调试和回放：

```go
type Action struct {
    Type      ActionType   // ✅ 可序列化
    Payload   interface{}  // ⚠️  建议使用简单类型
    Timestamp time.Time    // ✅ 可序列化
}

// 推荐：Payload 使用简单类型
WithPayload('A')        // rune
WithPayload("hello")    // string
WithPayload(42)         // int

// 不推荐：Payload 使用复杂类型
WithPayload(myStruct{}) // 难以序列化
```

### 3. 使用 Composite Action 提高性能

```go
// 推荐：并发执行独立任务
action.Batch(task1, task2, task3).Execute(ctx)

// 推荐：顺序执行依赖任务
action.Sequence(step1, step2, step3).Execute(ctx)
```

### 4. 限制并发数避免资源耗尽

```go
// 大量任务时使用 WorkerPool
pool := action.NewWorkerPool(10)
for i := 0; i < 1000; i++ {
    pool.Submit(action.ActionFunc(processItem))
}
pool.Stop()
```

### 5. 错误处理和恢复

```go
// 使用 Fallback 提供备用方案
fallback := action.NewFallbackAction(
    action.ActionFunc(primaryAction),
    action.ActionFunc(fallbackAction),
)

// 使用 Retry 处理临时错误
retry := action.NewRetryAction(
    action.ActionFunc(networkCall),
    3, time.Second,
)
```

### 6. Action 分发最佳实践

```go
// 1. 注册组件
d.Register(button)

// 2. 订阅全局快捷键
quitUnsub := d.Subscribe(action.ActionQuit, handleQuit)
defer quitUnsub() // 清理订阅

// 3. 设置默认处理器（调试）
d.SetDefaultHandler(func(a *action.Action) bool {
    fmt.Printf("Unhandled action: %s\n", a.Type)
    return false
})

// 4. 启用日志（开发环境）
d.EnableLog(debugMode)
```

## 与其他模块的集成

### Focus System

```go
// 将焦点导航映射为 Action
switch key {
case key.Tab:
    d.Dispatch(action.NewAction(action.ActionNavigateNext).
        WithTarget(focusID))
case key.ShiftTab:
    d.Dispatch(action.NewAction(action.ActionNavigatePrev).
        WithTarget(focusID))
}
```

### Platform（输入）

```go
// Platform 转换原始输入为 Action
func ConvertKeyEvent(key Key) *action.Action {
    switch key {
    case KeyEnter:
        return action.NewAction(action.ActionSubmit)
    case KeyEsc:
        return action.NewAction(action.ActionCancel)
    default:
        if key.IsChar() {
            return action.NewAction(action.ActionInputChar).
                WithPayload(key.Rune)
        }
    }
    return nil
}
```

### DevTools（时间旅行）

```go
// 记录 Action 用于重放
func (d *Dispatcher) DispatchWithLog(a *Action) bool {
    result := d.Dispatch(a)
    devtools.RecordAction(a, result)
    return result
}
```

## 常见问题

### Q: Action 没有被处理？

检查：
1. 目标组件是否注册：`d.Register(component)`
2. 是否被全局处理器拦截：检查 `Subscribe` 的返回值
3. 是否设置了 Target：`a.WithTarget(id)`

### Q: 复合 Action 失败了？

使用回调查看详细结果：

```go
action.BatchWithCallback(func(results []action.ActionResult) {
    for i, r := range results {
        if !r.OK {
            fmt.Printf("Action[%d] failed: %v\n", i, r.Error)
        }
    }
}, actions...)
```

### Q: 如何取消正在执行的 Composite Action？

```go
composite := action.Batch(action1, action2, action3)

// 在另一个 goroutine 中取消
go func() {
    time.Sleep(1 * time.Second)
    composite.Cancel()
}()

result := composite.Execute(ctx)
```
