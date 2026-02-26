# Intent System - 类型安全的声明式 Action 系统

## 概述

Intent 系统是一个类型安全的声明式 Action 系统，用于替代闭包模式的回调机制。它将组件的业务意图结构化，支持优先级调度和异步处理。

### 核心优势

- **类型安全**：泛型 Handler 保证编译期类型检查
- **无闭包**：组件只声明 Intent，不持有回调函数
- **优先级调度**：Intent → Priority → Lane 完整映射
- **Transition 支持**：异步操作可中断、可延迟
- **中间件**：支持日志、追踪等横切关注点
- **向后兼容**：通过 Bridge 与现有 Action 系统互操作

## 架构

```
┌─────────────────────────────────────────────────────────────┐
│                    Component (声明)                          │
│  Button("Open").Intent(OpenModal())                         │
└──────────────────────────┬──────────────────────────────────┘
                           │ Emit Intent
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    Intent Dispatcher                        │
│  - Resolve Priority (PriorityAware 接口)                    │
│  - Map to Lane (priority.DirtyLevel)                        │
│  - Queue Transition Intents (异步队列)                      │
└──────────────────────────┬──────────────────────────────────┘
                           │ Route to Handler
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    Typed Handler[T]                         │
│  - ctx.SetState("showModal", true)                          │
│  - ctx.ScheduleUpdate()                                     │
└──────────────────────────┬──────────────────────────────────┘
                           │ Mark Dirty
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    Scheduler                                │
│  - Fiber Update                                             │
│  - Render                                                   │
└─────────────────────────────────────────────────────────────┘
```

## 核心概念

### 1. Intent 接口

所有 Intent 必须实现 Intent 接口：

```go
type Intent interface {
    IntentType() string
}
```

### 2. PriorityAware 接口（可选）

如果 Intent 需要声明优先级，实现 PriorityAware 接口：

```go
type PriorityAware interface {
    Intent
    Priority() ActionPriority
}
```

### 3. TransitionIntent 接口（可选）

如果 Intent 支持异步处理，实现 TransitionIntent 接口：

```go
type TransitionIntent interface {
    Intent
    IsTransition() bool
}
```

## 优先级系统

### ActionPriority

```go
const (
    PriorityImmediate    // 紧急操作（焦点变更），同步处理
    PriorityUserBlocking // 用户操作（点击、输入），快速响应
    PriorityNormal       // 标准更新（默认）
    PriorityTransition   // 异步操作，可延迟、可中断
    PriorityIdle         // 后台任务，空闲时处理
)
```

### Priority → Lane 映射

| Priority            | Lane               | 说明           |
|---------------------|--------------------| ------------- |
| PriorityImmediate   | priority.DirtyHigh | 最高优先级     |
| PriorityUserBlocking| priority.DirtyHigh | 高优先级       |
| PriorityNormal      | priority.DirtyNormal | 正常优先级   |
| PriorityTransition  | priority.DirtyLow  | 低优先级       |
| PriorityIdle        | priority.DirtyLow  | 最低优先级     |

## 使用指南

### 1. 创建 Runtime

```go
import "github.com/wwsheng009/mint/runtime/intent"

// 创建完整的 Intent Runtime
rt := intent.NewRuntime()
```

### 2. 定义自定义 Intent

```go
// 定义 Intent 类型
type IncrementCounterIntent struct {
    Step int
}

// 实现 Intent 接口
func (IncrementCounterIntent) IntentType() string {
    return "IncrementCounter"
}

// 可选：实现 PriorityAware 接口
func (IncrementCounterIntent) Priority() intent.ActionPriority {
    return intent.PriorityUserBlocking
}
```

### 3. 注册 Handler

#### 方式一：类型安全注册（推荐）

```go
// 泛型注册，编译期类型检查
intent.RegisterTypedRuntime(rt, func(ctx *intent.ActionContext, i IncrementCounterIntent) intent.IntentResult {
    current, _ := ctx.GetState("counter")
    counter := 0
    if v, ok := current.(int); ok {
        counter = v
    }
    counter += i.Step
    ctx.SetState("counter", counter)
    return intent.HandledResult()
})
```

#### 方式二：通用注册

```go
rt.Register("IncrementCounter", intent.HandlerFunc(func(ctx *intent.ActionContext, i intent.Intent) intent.IntentResult {
    // 需要手动类型断言
    return intent.HandledResult()
}))
```

### 4. 发送 Intent

```go
// 直接发送
result := rt.Emit(IncrementCounterIntent{Step: 5})
if !result.Handled {
    // 处理未处理的情况
}

// 使用 Builder 模式
result := intent.NewBuilder(IncrementCounterIntent{Step: 5}).
    WithPriority(intent.PriorityImmediate).
    WithSource("counter-button").
    Dispatch(rt.Dispatcher)

// 使用 Emitter（组件内部）
emitter := intent.NewEmitter(rt.Dispatcher, "component-1")
result := emitter.Emit(IncrementCounterIntent{Step: 5})
```

### 5. 在组件中使用

```go
// ❌ 旧方式：闭包模式
button.OnClick(func() {
    counter += 1
})

// ✅ 新方式：Intent 模式
button.Intent(IncrementCounterIntent{Step: 1})
```

## 内置 Intent 类型

### 导航 Intent

```go
// 导航到路径
intent.Navigate("/dashboard")
intent.NavigateWithParams("/users", map[string]interface{}{"id": 123})
```

### 状态 Intent

```go
// 设置状态
intent.SetState("key", value)

// 切换布尔状态
intent.Toggle("isExpanded")
```

### UI Intent

```go
// 打开/关闭 Modal
intent.OpenModal("settings-modal")
intent.OpenModalWithData("modal-1", data)
intent.CloseModal("settings-modal")
intent.CloseModalWithResult("modal-1", result)

// 工具提示
intent.ShowTooltip("Help text", "target-id")
intent.HideTooltip()
```

### 焦点 Intent

```go
// 聚焦/失焦
intent.Focus("username-input")
intent.Blur("username-input")
```

### 数据 Intent（Transition）

```go
// 异步加载数据
intent.LoadData("/api/users", "users")

// 刷新数据
intent.Refresh([]string{"users", "posts"})
```

### 表单 Intent

```go
// 表单提交
intent.SubmitForm("login-form", formData)

// 表单验证
intent.ValidateForm("login-form")
```

### 动作 Intent

```go
// 点击
intent.Click("button-1")

// 按压（Enter键等）
intent.Press("submit-button")
```

## Transition（异步）支持

### 定义 Transition Intent

```go
type LoadDataIntent struct {
    URL  string
    Key  string
}

func (LoadDataIntent) IntentType() string {
    return "LoadData"
}

// 实现 TransitionIntent 接口
func (LoadDataIntent) IsTransition() bool {
    return true
}
```

### 处理 Transition Intent

Transition Intent 会被自动加入队列，由 Scheduler 在空闲时处理：

```go
// 发送 Transition Intent（自动入队）
result := rt.Emit(intent.LoadData("/api/data", "dataKey"))
// result.Async == true

// Scheduler 处理队列
rt.Dispatcher.ProcessQueue(10 * time.Millisecond)
```

### Transition 包装器

可以将任意 Intent 包装为 Transition：

```go
original := SaveDataIntent{Data: "test"}
wrapped := intent.Transition(original)
// wrapped.IsTransition() == true
```

## 中间件

中间件用于处理横切关注点（日志、追踪、错误处理等）：

### 添加中间件

```go
// 日志中间件
rt.Registry.Use(func(next intent.Handler) intent.Handler {
    return intent.HandlerFunc(func(ctx *intent.ActionContext, i intent.Intent) intent.IntentResult {
        start := time.Now()
        log.Printf("[Intent] Before: %s", i.IntentType())
        
        result := next.Handle(ctx, i)
        
        log.Printf("[Intent] After: %s (handled: %v, duration: %v)", 
            i.IntentType(), result.Handled, time.Since(start))
        return result
    })
})
```

### 错误处理中间件

```go
rt.Registry.Use(func(next intent.Handler) intent.Handler {
    return intent.HandlerFunc(func(ctx *intent.ActionContext, i intent.Intent) intent.IntentResult {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("[Intent] Panic in %s: %v", i.IntentType(), r)
            }
        }()
        return next.Handle(ctx, i)
    })
})
```

## ActionContext

ActionContext 提供了 Handler 执行时的上下文：

```go
type ActionContext struct {
    context.Context    // Go 标准上下文
    Source      string // 发送源
    Timestamp   time.Time // 发送时间
    // ...
}

// 方法
ctx.SetState(key, value)      // 设置状态
ctx.GetState(key)             // 获取状态
ctx.ScheduleUpdate()          // 调度更新
```

## IntentResult

Handler 返回 IntentResult 表示处理结果：

```go
type IntentResult struct {
    Handled bool          // 是否已处理
    Error   error         // 错误信息
    Async   bool          // 是否异步
    Done    chan struct{} // 异步完成信号
}

// 构造函数
intent.HandledResult()                    // 成功处理
intent.ErrorResult(err)                   // 处理失败
intent.AsyncResult(done)                  // 异步处理
```

## 与现有 Action 系统集成

### Action → Intent 转换

```go
import "github.com/wwsheng009/mint/runtime/action"

bridge := intent.NewActionBridge(dispatcher, registry)

// 从 Action 创建 Intent
a := &action.Action{Type: action.ActionClick, Payload: data}
intent := intent.IntentFromAction(a)

// 从 Action 派发
result := bridge.DispatchFromAction(a)
```

### Intent → Action 转换

```go
// 将 Intent 转换为 Action
i := intent.OpenModal("settings")
a := intent.ActionFromIntent(i)
```

## 完整示例

### 示例 1：计数器应用

```go
package main

import (
    "fmt"
    "github.com/wwsheng009/mint/runtime/intent"
)

// 定义 Intent
type IncrementIntent struct {
    Step int
}

func (IncrementIntent) IntentType() string { return "Increment" }
func (IncrementIntent) Priority() intent.ActionPriority {
    return intent.PriorityUserBlocking
}

func main() {
    // 创建 Runtime
    rt := intent.NewRuntime()

    // 注册 Handler
    intent.RegisterTypedRuntime(rt, func(ctx *intent.ActionContext, i IncrementIntent) intent.IntentResult {
        current, _ := ctx.GetState("count")
        count := 0
        if v, ok := current.(int); ok {
            count = v
        }
        count += i.Step
        ctx.SetState("count", count)
        fmt.Printf("Count: %d\n", count)
        return intent.HandledResult()
    })

    // 发送 Intent
    rt.Emit(IncrementIntent{Step: 1})
    rt.Emit(IncrementIntent{Step: 5})
}

// Output:
// Count: 1
// Count: 6
```

### 示例 2：Modal 管理

```go
package main

import (
    "fmt"
    "github.com/wwsheng009/mint/runtime/intent"
)

func main() {
    rt := intent.NewRuntime()

    // 注册 Modal 打开 Handler
    intent.RegisterTypedRuntime(rt, func(ctx *intent.ActionContext, i intent.OpenModalIntent) intent.IntentResult {
        fmt.Printf("Opening modal: %s\n", i.ModalID)
        ctx.SetState("activeModal", i.ModalID)
        ctx.ScheduleUpdate()
        return intent.HandledResult()
    })

    // 注册 Modal 关闭 Handler
    intent.RegisterTypedRuntime(rt, func(ctx *intent.ActionContext, i intent.CloseModalIntent) intent.IntentResult {
        fmt.Printf("Closing modal: %s\n", i.ModalID)
        ctx.SetState("activeModal", nil)
        ctx.ScheduleUpdate()
        return intent.HandledResult()
    })

    // 使用
    rt.Emit(intent.OpenModal("settings"))
    rt.Emit(intent.CloseModal("settings"))
}

// Output:
// Opening modal: settings
// Closing modal: settings
```

### 示例 3：异步数据加载

```go
package main

import (
    "fmt"
    "time"
    "github.com/wwsheng009/mint/runtime/intent"
)

func main() {
    rt := intent.NewRuntime()

    // 注册异步加载 Handler
    intent.RegisterTypedRuntime(rt, func(ctx *intent.ActionContext, i intent.LoadDataIntent) intent.IntentResult {
        fmt.Printf("Loading data from: %s\n", i.URL)
        // 模拟异步操作
        ctx.SetState(i.Key, map[string]interface{}{"loaded": true})
        return intent.HandledResult()
    })

    // 发送 Transition Intent
    result := rt.Emit(intent.LoadData("/api/users", "users"))
    fmt.Printf("Async: %v\n", result.Async)

    // 处理队列
    processed := rt.Dispatcher.ProcessQueue(10 * time.Millisecond)
    fmt.Printf("Processed: %d\n", processed)
}

// Output:
// Async: true
// Processed: 1
```

## 设计原则

### 1. Intent 是声明，不是行为

```
❌ 闭包模式：组件持有行为
button.OnClick(func() { showModal() })

✅ Intent 模式：组件声明意图
button.Intent(OpenModal())
```

### 2. Intent 可调度

- 不同优先级的 Intent 会被 Scheduler 按优先级处理
- Transition Intent 可以被高优先级任务打断
- 支持时间分片和增量渲染

### 3. Intent 可追踪

- 所有 Intent 都有明确的类型标识
- Dispatcher 记录完整的派发日志
- 便于 DevTools 追踪和调试

## API 参考

### 类型

| 类型 | 说明 |
|------|------|
| Intent | 基础 Intent 接口 |
| PriorityAware | 带优先级的 Intent |
| TransitionIntent | 异步 Intent |
| ActionPriority | 优先级枚举 |
| ActionContext | Handler 上下文 |
| IntentResult | 处理结果 |
| Registry | Handler 注册表 |
| Dispatcher | Intent 分发器 |
| Runtime | 完整运行时 |
| Emitter | 组件发送器 |
| Builder | Intent 构建器 |

### 函数

| 函数 | 说明 |
|------|------|
| NewRuntime() | 创建运行时 |
| RegisterTyped[T]() | 类型安全注册 |
| HandledResult() | 创建成功结果 |
| ErrorResult(err) | 创建错误结果 |
| AsyncResult(ch) | 创建异步结果 |
| Transition(i) | 包装为 Transition |
| WithPriority(i, p) | 覆盖优先级 |

## 迁移指南

### 从闭包模式迁移

1. **识别闭包使用**

```go
// 旧代码
button.OnClick(func() {
    setShowModal(true)
})
```

2. **定义 Intent**

```go
type ShowModalIntent struct{}

func (ShowModalIntent) IntentType() string { return "ShowModal" }
```

3. **注册 Handler**

```go
intent.RegisterTypedRuntime(rt, func(ctx *intent.ActionContext, i ShowModalIntent) intent.IntentResult {
    ctx.SetState("showModal", true)
    return intent.HandledResult()
})
```

4. **更新组件**

```go
// 新代码
button.Intent(ShowModalIntent{})
```

## 参考资料

- [Fiber-first Architecture](../../docs/fiber/fiber_first/)
- [Action Runtime](../action/)
- [Scheduler](../scheduler/)
- [Priority System](../priority/)
