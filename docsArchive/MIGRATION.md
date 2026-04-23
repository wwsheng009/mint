# 从闭包模式迁移到纯状态模式

本文档帮助您将基于闭包的状态管理迁移到 Fiber-first 的纯状态模式。

## 目录

- [迁移概述](#迁移概述)
- [状态访问模式对比](#状态访问模式对比)
- [逐步迁移指南](#逐步迁移指南)
- [常见迁移场景](#常见迁移场景)
- [向后兼容性](#向后兼容性)

---

## 迁移概述

### 为什么迁移？

| 方面 | 闭包模式 (旧) | 纯状态模式 (新) |
|------|---------------|----------------|
| **状态存储** | 回调闭包捕获变量 | `ComponentContext.GlobalState` |
| **状态更新** | 直接修改闭包变量 | `ctx.SetState(key, value)` |
| **状态读取** | 从闭包变量读取 | `ctx.GetIntState(key, default)` |
| **组件复用** | 难以实例化 | `InstanceManager` 自动复用 |
| **副作用管理** | 手动清理 | `useEffect`生命周期自动管理 |
| **可测试性** | 低 | 高 |

### 核心变化

```go
// ❌ 旧模式：基于闭包
func NewApp() *app.App {
    step := 1                      // 闭包变量
    agreed := false                // 闭包变量
    fields := make(map[int]map[string]string)

    return app.New(app.WithInit(...)).
        OnIntent(func(i UpdateStepIntent) {
            step = i.Step  // 直接修改闭包变量
        })

    // 问题：状态在多个地方，难以追踪和调试
}

// ✅ 新模式：纯状态
func App() ui.VNode {
    ctx := rtui.GetCurrentContext()
    step := ctx.GetIntState("step", 1)      // 从 GlobalState 读取
    agreed := ctx.GetBoolState("agreed", false)

    return ui.VStack(...).OnPress(UpdateStepIntent{...})

    // 优势：状态集中的 ComponentContext.GlobalState 中
}
```

---

## 状态访问模式对比

### 1. 读取状态

```go
// ❌ 旧模式：闭包变量
step := step  // 从闭包读取

// ✅ 新模式：从 GlobalState 读取
ctx := rtui.GetCurrentContext()
step := ctx.GetIntState("step", 1)
// 或者使用更清晰的方法
step := ctx.GetGlobalInt("step", 1)
```

### 2. 更新状态

```go
// ❌ 旧模式：直接修改闭包变量
step = newStep

// ✅ 新模式：通过 Intent Handler
ui.RegisterIntent(func(ctx *intent.ActionContext, i UpdateStepIntent) intent.IntentResult {
    ctx.SetState("step", i.Step)  // 更新 GlobalState
    return intent.HandledResult()
})

// 发射 Intent
app.Button("Next").
    OnPress(UpdateStepIntent{Step: step + 1}).  // 触发 Intent
    Build()
```

### 3. 嵌套状态

```go
// ❌ 旧模式：嵌套 map
type FormData map[string]string
var fields = map[int]FormData{
    1: {"username": "", "email": ""},
    2: {"address": "", "phone": ""},
}

// 访问：fields[step]["username"]

// ✅ 新模式：扁平化键
ctx.SetState("username", "john")
ctx.SetState("email", "john@example.com")

// 或者带前缀的键
ctx.SetState("step1:username", "john")
username := ctx.GetStringState("step1:username", "")
```

---

## 逐步迁移指南

### 第 1 步：识别闭包变量

找到所有在闭包中捕获的状态变量。

```go
// 示例：检查闭包变量
func NewApp() *app.App {
    var step int               // ← 闭包变量
    var agreed bool            // ← 闭包变量
    var fields map[int]map[string]string  // ← 闭包变量

    // ...
}
```

### 第 2 步：定义状态键

为每个闭包变量定义一个全局状态键。

```go
// 状态键的定义（常量，便于管理）
const (
    StateStep    = "step"
    StateAgreed  = "agreed"
    StateField   = "field"
)
```

### 第 3 步：编写 Intent Handler

为每个状态更新定义 Intent 和 Handler。

```go
// Intent 定义
type IntentFactory struct{}

func (IntentFactory) UpdateStep(step int) UpdateStepIntent {
    return UpdateStepIntent{Step: step}
}

func (IntentFactory) UpdateAgreed(agreed bool) UpdateAgreedIntent {
    return UpdateAgreedIntent{Agreed: agreed}
}

// Intent Handler 注册
ui.RegisterIntent(func(ctx *intent.ActionContext, i UpdateStepIntent) intent.IntentResult {
    ctx.SetState(StateStep, i.Step)
    return intent.HandledResult()
})

ui.RegisterIntent(func(ctx *intent.ActionContext, i UpdateAgreedIntent) intent.IntentResult {
    ctx.SetState(StateAgreed, i.Agreed)
    return intent.HandledResult()
})
```

### 第 4 步：重构组件函数

将旧组件函数转换为使用 `Context.State` 的新版本。

```go
// ❌ 旧组件
func renderStep1() ui.VNode {
    username, _ := fields[1]["username"]  // 闭包变量
    email, _ := fields[1]["email"]

    return ui.VStack(
        ui.Text("Username:"),
        app.NewTextBuilder().Value(username).OnInput(func(v string) {
            fields[1]["username"] = v  // 闭包更新
        }).Build(),
        // ...
    )
}

// ✅ 新组件
func Step1() ui.VNode {
    ctx := rtui.GetCurrentContext()

    username := ctx.GetStringState(StateField+":username", "")
    email := ctx.GetStringState(StateField+":email", "")

    return ui.VStack(
        ui.Text("Username:"),
        app.NewTextBuilder().Value(username).
            OnInput(func(v string) {
                // 发射 Intent 更新
                EmitIntent(UpdateFieldIntent{
                    Field: "username",
                    Value: v,
                })
            }).
            Build(),
        // ...
    )
}
```

### 第 5 步：更新 App 函数

确保 `App()` 从 `ComponentContext` 读取状态。

```go
// ❌ 旧 App
func NewApp() *app.App {
    var step int
    // ...

    appBuilder := app.New(app.WithInit(...)).
        OnIntent(func(i intents.UpdateStepInt) intents.Intent {
            step = i.Step
            rtui.Repaint()
            return intents.Handled
        }).
        ComponentTree(app.UI{
            UI: func() ui.BoundComponent { return ComponentTree(step) },
        })

    return appBuilder
}

// ✅ 新 App
func App() ui.VNode {
    ctx := rtui.GetCurrentContext()
    step := ctx.GetIntState(StateStep, 1)

    return ui.VStack(
        StepSelector(step),
        StepContent(step),
        Actions(step),
    )
}
```

### 第 6 步：移除状态初始化的 WithInit 回调

旧模式中需要在 `WithInit` 中初始化状态变量，新模式中不需要。

```go
// ❌ 旧模式：WithInit 中初始化
app.New(app.WithInit(func(ctx ui.Context) {
    step = 1
    agreed = false
    fields = map[int]map[string]string{
        1: {"username": "", "email": ""},
        2: {"address": "", "phone": ""},
    }
}))

// ✅ 新模式：无需显式初始化（GlobalState 默认为空 map）
// 也可以在 App() 中设置初始值
func App() ui.VNode {
    ctx := rtui.GetCurrentContext()

    // 首次渲染时设置默认值
    if ctx.GetIntState("inited", 0) == 0 {
        ctx.SetState("inited", 1)
        ctx.SetState(StateStep, 1)
        ctx.SetState(StateAgreed, false)
    }

    // ...
}
```

---

## 常见迁移场景

### 场景 1：简单计数器

```go
// ❌ 旧模式
var count int

appButton := app.Button(fmt.Sprintf("Count: %d", count)).
    OnClick(func() {
        count++
        rtui.Repaint()
    }).
    Build()

// ✅ 新模式
type IncrementIntent struct{}
func (IncrementIntent) IntentType() string { return "Increment" }

ui.RegisterIntent(func(ctx *intent.ActionContext, _ IncrementIntent) intent.IntentResult {
    currentCount := ctx.GetIntState("count", 0)
    ctx.SetState("count", currentCount+1)
    return intent.HandledResult()
})

func Counter() ui.VNode {
    ctx := rtui.GetCurrentContext()
    count := ctx.GetIntState("count", 0)

    return app.Button(fmt.Sprintf("Count: %d", count)).
        OnPress(IncrementIntent{}).
        Build()
}
```

### 场景 2：表单字段

```go
// ❌ 旧模式
var fields map[string]string

appNewTextBuilder().Value(value).
    OnInput(func(newValue string) {
        fields[fieldName] = newValue
        rtui.Repaint()
    }).
    Build()

// ✅ 新模式
type FieldUpdateIntent struct {
    Field string
    Value string
}
func (FieldUpdateIntent) IntentType() string { return "FieldUpdate" }

func InputField(fieldName string) ui.VNode {
    ctx := rtui.GetCurrentContext()
    value := ctx.GetStringState(fieldName, "")

    return appNewTextBuilder().Value(value).
        OnInput(func(newValue string) {
            EmitIntent(FieldUpdateIntent{Field: fieldName, Value: newValue})
        }).
        Build()
}
```

### 场景 3：状态同步

```go
// ❌ 旧模式：多个相关状态
var currentStep int
var maxStep int
var canProceed bool

// 手动同步
canProceed = currentStep < maxStep

// ✅ 新模式：从单个状态派生
func App() ui.VNode {
    ctx := rtui.GetCurrentContext()
    step := ctx.GetIntState(StateStep, 1)

    const maxStep = 5
    canProceed := step < maxStep  // ✅ 自动计算，无需同步

    return ui.VStack(...)
}
```

### 场景 4：列表管理

```go
// ❌ 旧模式
var items []Item

appButton("Add Item").OnClick(func() {
    items = append(items, Item{})
    rtui.Repaint()
})

// ✅ 新模式
type AddItemIntent struct {}
func (AddItemIntent) IntentType() string { return "AddItem" }

ui.RegisterIntent(func(ctx *intent.ActionContext, _ AddItemIntent) intent.IntentResult {
    var currentItems []Item
    if items, exists := ctx.GetState("items"); exists {
        currentItems = items.([]Item)
    }
    ctx.SetState("items", append(currentItems, Item{}))
    return intent.HandledResult()
})
```

---

## 向后兼容性

### 状态字段名称兼容

为了支持从旧代码平滑迁移，`ComponentContext` 保留了向后兼容的方法：

```go
// 这些方法仍然有效（别名）
ctx.SetState(key, value)      // 实际更新 GlobalState
ctx.GetState(key)             // 实际从 GlobalState 读取
ctx.GetIntState(key, default) // 实际从 GlobalState 读取
```

### 新的语义化方法

新代码建议使用更清晰的方法：

```go
// 推荐使用：语义更清晰
ctx.SetGlobalState(key, value)           // 设置全局状态
ctx.GetGlobalState(key, default)         // 获取全局状态
ctx.GetGlobalInt(key, default)           // 获取 int 类型的全局状态
ctx.GetGlobalString(key, default)        // 获取 string 类型的全局状态
ctx.GetGlobalBool(key, false)            // 获取 bool 类型的全局状态
```

### 批量更新自动优化

新实现自动支持批量更新：

```go
// ✅ 多次 SetState 会被自动批处理
ctx.SetState("field1", "value1")
ctx.SetState("field2", "value2")
ctx.SetState("field3", "value3")
// 这些更新会被合并，只触发一次重新渲染
```

---

## 迁移检查清单

在完成迁移后，请确保：

- [ ] 所有闭包变量已移除
- [ ] 每个状态更新都有对应的 Intent 和 Handler
- [ ] 组件函数没有返回值（只是返回 VNode）
- [ ] 所有状态读取通过 `ctx.GetxxxState()` 进行
- [ ] 没有 `WithInit` 中的状态初始化代码
- [ ] 使用 `ui.Run()` 而不是 `app.New()`
- [ ] 所有按钮的 `onClick/OnPress` 使用 Intent
- [ ] 嵌套状态已扁平化或使用前缀
- [ ] 运行测试确保行为一致

---

## 示例：完整迁移前后对比

### 迁移前（闭包模式）

```go
package main

import (
    "fmt"
    "github.com/wwsheng009/mint/app"
)

func NewApp() *app.App {
    var step int
    var agreed bool
    var fields map[int]map[string]string

    step = 1
    agreed = false
    fields = map[int]map[string]string{
        1: {"username": "", "email": ""},
        2: {"address": "", "phone": ""},
    }

    componentTree := func() ui.BoundComponent {
        return ComponentTree(step, agreed, fields)
    }

    return app.New(app.WithInit(...)).
        OnIntent(func(i intents.UpdateStepInt) intents.Intent {
            step = i.Step
            rtui.Repaint()
            return intents.Handled
        }).
        OnIntent(func(i intents.UpdateFieldInt) intents.Intent {
            fields[step][i.Field] = i.Value
            rtui.Repaint()
            return intents.Handled
        }).
        UI(app.UI{
            UI: func() ui.BoundComponent { return componentTree() },
        })
}

func ComponentTree(step int, agreed bool, fields map[int]map[string]string) ui.BoundComponent {
    return &ui.VStack{
        Children: []ui.BoundComponent{
            StepIndicator(step),
            StepContent(step, fields),
            Actions(step, agreed),
        },
    }
}
```

### 迁移后（纯状态模式）

```go
package main

import (
    "github.com/wwsheng009/mint/app"
    "github.com/wwsheng009/mint/runtime/intent"
    rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Intent 定义
type UpdateStepIntent struct {
    Step int
}
func (UpdateStepIntent) IntentType() string { return "UpdateStep" }

type UpdateFieldIntent struct {
    Field string
    Value string
}
func (UpdateFieldIntent) IntentType() string { return "UpdateField" }

// 主应用
func App() ui.VNode {
    ctx := rtui.GetCurrentContext()
    step := ctx.GetIntState("step", 1)
    agreed := ctx.GetBoolState("agreed", false)

    return ui.VStack(
        StepIndicator(step),
        StepContent(step),
        Actions(step, agreed),
    )
}

func main() {
    ui.Run(App,
        ui.WithInit(func() {
            // 注册 Intent Handler
            ui.RegisterIntent(func(ctx *intent.ActionContext, i UpdateStepIntent) intent.IntentResult {
                ctx.SetState("step", i.Step)
                return intent.HandledResult()
            })

            ui.RegisterIntent(func(ctx *intent.ActionContext, i UpdateFieldIntent) intent.IntentResult {
                ctx.SetState(i.Field, i.Value)
                return intent.HandledResult()
            })
        }),
    )
}
```

---

## 常见问题

### Q: 我应该在什么时候迁移？

A: 如果您的项目满足以下条件，建议迁移：
- 需要跨多个组件共享状态
- 遇到组件复用问题（如多个对话框实例）
- 需要更好的测试覆盖率
- 闭包状态管理变得复杂

### Q: 迁移会影响现有功能吗？

A: 不会。两种模式可以共存。您可以逐步迁移，不必一次性重写全部代码。

### Q: 如何处理复杂的嵌套数据结构？

A: 有两种方式：
1. **扁平化**：使用前缀键（如 `"user:username"`, `"user:email"`）
2. **嵌入 JSON**：将嵌套数据序列化为 JSON 字符串存储

```go
// 方式 1：扁平化键
ctx.SetState("user:username", "john")
ctx.SetState("user:email", "john@example.com")

// 方式 2：JSON 嵌入
type User struct {
    Username string `json:"username"`
    Email    string `json:"email"`
}
user := User{Username: "john", Email: "john@example.com"}
data, _ := json.Marshal(user)
ctx.SetState("user", data)
```

### Q: 如何撤销迁移？

A: 只需将代码恢复到之前的版本即可。两种模式完全独立，没有强制依赖。

---

## 总结

迁移到纯状态模式的主要优势：

1. ✅ **更好的组件复用**：`InstanceManager` 自动管理实例
2. ✅ **更容易测试**：状态与逻辑分离
3. ✅ **更清晰的代码**：统一的状态访问 API
4. ✅ **更好的性能**：批量更新自动优化
5. ✅ **更强的类型安全**：Intent 类型系统

建议：
- 从小的功能模块开始迁移
- 保持旧代码和新代码并存
- 使用 IDE 的重命名工具辅助
- 充分测试确保行为一致

### 相关文档

- [FIBER_STATE_ARCHITECTURE.md](./FIBER_STATE_ARCHITECTURE.md) - 核心架构
- [BEST_PRACTICES.md](./BEST_PRACTICES.md) - 最佳实践
- [PERFORMANCE.md](./PERFORMANCE.md) - 性能优化
