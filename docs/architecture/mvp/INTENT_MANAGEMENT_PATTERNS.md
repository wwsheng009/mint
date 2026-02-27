# Intent 管理模式指南

**生成时间**: 2026-02-27
**最后更新**: 2026-02-27

---

## 目录

- [一、概述](#一概述)
- [二、Intent 类型架构](#二intent-类型架构)
- [三、三种管理模式对比](#三三种管理模式对比)
- [四、方案详情与示例](#四方案详情与示例)
  - [方案 1: 组件级状态（ui.On + Simple* Intent）](#方案-1-组件级状态ui-on--simple-intent)
  - [方案 2: 全局状态（runtime/intent 内置）](#方案-2-全局状态runtimeintent-内置)
  - [方案 3: 自定义 Intent（灵活复杂场景）](#方案-3-自定义-intent灵活复杂场景)
- [五、常见问题与陷阱](#五常见问题与陷阱)
- [六、最佳实践](#六最佳实践)
- [七、迁移指南](#七迁移指南)

---

## 一、概述

在 Mint UI 中，Intent 是处理用户交互和状态变更的核心机制。根据不同的场景，我们有三种主要的 Intent 管理模式：

| 模式 | 适用场景 | 数据源 | 代码复杂度 |
|------|---------|--------|-----------|
| **组件级状态** | 单组件内部状态 | `UseStateInt/UseStateBool` | 简单 |
| **全局状态** | 跨组件共享状态 | `GlobalState` | 最简单 |
| **自定义 Intent** | 复杂业务逻辑、参数传递 | `UseState + 自定义类型` | 中等 |

---

## 二、Intent 类型架构

### 2.1 Intent 类型层次

```
runtime/intent/builtin.go          # 系统内置（已注册 handler）
├── IncrementIntent {Key, Delta}   # 全局状态递增
├── ToggleIntent {Key}             # 全局状态切换
└── ...

ui/intent.go                       # UI 层通用（无 handler）
├── SimpleIncrementIntent {}       # 简单递增（组件状态）
├── SimpleDecrementIntent {}       # 简单递减（组件状态）
├── SimpleToggleIntent {}          # 简单切换（组件状态）
└── ...

[User Custom]                      # 用户自定义（灵活扩展）
├── CustomIncrement {Step}         # 带参数的递增
├── OpenModal {ModalID, Data}      # 模态框打开
└── ...
```

### 2.2 命名规范

| 前缀 | 含义 | 示例 |
|------|------|------|
| `Increment*` | 全局状态，携带 Key | `IncrementIntent{Key: "count"}` |
| `Simple*` | 组件状态，无参数 | `SimpleIncrementIntent{}` |
| 无前缀 | 用户自定义，灵活定义 | `CustomIncrement{Step: 10}` |

---

## 三、三种管理模式对比

### 3.1 快速对比表

| 维度 | 方案 1 | 方案 2 | 方案 3 |
|------|--------|--------|--------|
| **API** | `ui.On + Simple*` | `intent.Increment` | `ui.On + 自定义` |
| **状态位置** | Hooks (`UseStateInt`) | GlobalState (`ctx.GetState`) | Hooks (`UseStateInt`) |
| **闭包捕获** | ✅ 自然访问 | ❌ 需要函数式 setState | ✅ 自然访问 |
| **跨组件共享** | ❌ 不支持 | ✅ 支持 | ❌ 需要全局传递 |
| **参数传递** | ❌ 无参数 | ✅ 通过 Key/Delta | ✅ 自定义结构 |
| **Intent 冲突** | ✅ 无冲突 | ❌ Key 冲突风险 | ✅ IntentType 唯一 |
| **代码行数** | ~10 行 | ~5 行 | ~15 行 |

### 3.2 决策树

```
需要跨组件共享状态？
├─ 是 → 使用方案 2：runtime/intent 内置函数
└─ 否 → 需要参数传递？
    ├─ 是 → 使用方案 3：自定义 Intent
    └─ 否 → 使用方案 1：ui.On + Simple* Intent
```

---

## 四、方案详情与示例

### 方案 1: 组件级状态（ui.On + Simple* Intent）

#### 适用场景
- ✅ 单组件内部的状态管理
- ✅ 简单交互（递增、递减、切换等）
- ✅ 不需要跨组件共享状态
- ✅ 快速原型开发

#### 架构图

```
┌─────────────────────────────────────┐
│         组件渲染循环                 │
├─────────────────────────────────────┤
│  UseStateInt(0) → count, setCount  │
│  UseStateInt(1) → _, _, _          │
│                                     │
│  ui.On(SimpleIncrementIntent){      │
│    setCount(func(c) int {           │
│      return c + 1                  │
│    })                               │
│  }                                  │
└─────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────┐
│    事件触发：Button.OnPress         │
└─────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────┐
│  Intent Runtime                    │
│  → 查找 SimpleIncrementIntent handler│
│  → 执行闭包：调用 setCount           │
└─────────────────────────────────────┘
```

#### 核心实现

```go
// ui/intent.go

// Simple* Intent 类型（无参数，避免与 runtime/intent 冲突）
type SimpleIncrementIntent struct{}
func (SimpleIncrementIntent) IntentType() string { return "SimpleIncrement" }
func (SimpleIncrementIntent) StayPressed() bool  { return true }

// On 注册 handler（使用 sync.Map 避免重复注册）
var registeredHandlers sync.Map

func On[T interface{ IntentType() string; StayPressed() bool }](
    intentType T,
    handler func(),
) {
    key := intentType.IntentType()
    if _, loaded := registeredHandlers.LoadOrStore(key, true); loaded {
        return
    }
    rtui.RegisterIntent(func(ctx *intent.ActionContext, i T) intent.IntentResult {
        handler()
        return intent.HandledResult()
    })
}
```

#### 使用示例

```go
package main

import (
    "fmt"
    "github.com/wwsheng009/mint/app"
    "github.com/wwsheng009/mint/ui"
)

func Counter() ui.VNode {
    // 1. 定义状态（使用 hooks）
    _, setCount, _ := ui.UseStateInt(0)

    // 2. 注册 handler（闭包捕获 setCount）
    ui.On(ui.SimpleIncrementIntent{}, func() {
        // ✅ 函数形式的 setCount，避免闭包捕获旧值
        setCount(func(c int) int { return c + 1 })
    })

    ui.On(ui.SimpleDecrementIntent{}, func() {
        setCount(func(c int) int { return c - 1 })
    })

    // 3. 读取当前值（hooks 顺序必须保持一致）
    count, _, _ := ui.UseStateInt(0)

    // 4. 返回 VNode（绑定 Intent）
    return ui.VStack(
        app.NewTextBuilder(fmt.Sprintf("Count: %d", count)).Build(),
        ui.HStack(
            app.ButtonBuilder(" - ").OnPress(ui.SimpleDecrementIntent{}).Build(),
            app.ButtonBuilder(" + ").OnPress(ui.SimpleIncrementIntent{}).Build(),
        ),
    )
}
```

#### 陷阱：闭包捕获旧值

```go
// ❌ 错误示例：闭包捕获首次渲染时的旧值（0）
ui.On(ui.SimpleIncrementIntent{}, func() {
    count, _, _ := ui.UseStateInt(0)
    setCount(count + 1)  // ❌ 永远是 0 + 1 = 1
})

// ✅ 正确示例：使用函数形式
ui.On(ui.SimpleIncrementIntent{}, func() {
    setCount(func(c int) int { return c + 1 })  // ✅ 获取最新值
})
```

---

### 方案 2: 全局状态（runtime/intent 内置）

#### 适用场景
- ✅ 多个组件需要访问同一状态
- ✅ 导航状态、模态框状态等全局 UI 状态
- ✅ 全局配置、主题切换等
- ❌ 不适合大量局部状态（命名冲突风险）

#### 架构图

```
┌─────────────────────────────────────┐
│    组件 A                            │
│  ctx.GetIntState("count", 0)        │
│  Button → OnPress(Incr("count",1))  │
└─────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────┐
│    GlobalState (单例)               │
│  ["count"] = 0                      │
└─────────────────────────────────────┘
              ▲
              │
┌─────────────────────────────────────┐
│    组件 B                            │
│  ctx.GetIntState("count", 0)        │
│  Text → 显示同一计数                │
└─────────────────────────────────────┘
```

#### 核心实现

```go
// runtime/intent/builtin_handlers.go

// IncrementIntent 和内置 handler
type IncrementIntent struct {
    Key   string
    Delta int
}

func handleIncrement(ctx *ActionContext, i IncrementIntent) IntentResult {
    current, ok := ctx.GetState(i.Key)
    if !ok {
        ctx.SetState(i.Key, i.Delta)
    } else {
        switch v := current.(type) {
        case int:
            ctx.SetState(i.Key, v+i.Delta)
        case float64:
            ctx.SetState(i.Key, v+float64(i.Delta))
        }
    }
    ctx.ScheduleUpdate()
    return HandledResult()
}

// 辅助函数
func Increment(key string, delta int) IncrementIntent {
    return IncrementIntent{Key: key, Delta: delta}
}
```

#### 使用示例

```go
package main

import (
    "fmt"
    "github.com/wwsheng009/mint/app"
    "github.com/wwsheng009/mint/runtime/intent"
    "github.com/wwsheng009/mint/ui"
)

func GlobalCounter() ui.VNode {
    ctx := ui.GetCurrentContext()
    count := ctx.GetIntState("global_count", 0)

    return ui.VStack(
        app.NewTextBuilder(fmt.Sprintf("Global Count: %d", count)).Build(),
        ui.HStack(
            app.ButtonBuilder(" - ").
                OnPress(intent.Decrement("global_count", 1)).
                Build(),
            app.ButtonBuilder(" + ").
                OnPress(intent.Increment("global_count", 1)).
                Build(),
        ),
    )
}

// 另一个组件可以共享同一个计数
func AnotherComponent() ui.VNode {
    ctx := ui.GetCurrentContext()
    count := ctx.GetIntState("global_count", 0)
    return ui.Text(fmt.Sprintf("Shared: %d", count))
}
```

#### 陷阱：Key 命名冲突

```go
// ❌ 错误示例：多个组件使用相同的 Key
func ComponentA() {
    // Key: "count" -> 0
    ctx.GetIntState("count", 0)
}
func ComponentB() {
    // Key: "count" -> 冲突！
    ctx.GetIntState("count", 0)
}

// ✅ 正确示例：使用组件前缀唯一标识
func ComponentA() {
    ctx.GetIntState("counter_A_count", 0)
}
func ComponentB() {
    ctx.GetIntState("counter_B_count", 0)
}
```

---

### 方案 3: 自定义 Intent（灵活复杂场景）

#### 适用场景
- ✅ 需要传递参数（步长、选项、索引等）
- ✅ 复杂业务逻辑（计算、验证、API 调用）
- ✅ 需要类型安全的 Intent（结构体字段）
- ✅ 模块化组件设计

#### 架构图

```
┌─────────────────────────────────────┐
│      自定义 Intent 定义             │
│                                     │
│  type StepIncrement struct {        │
│      Step      int                  │
│      Min, Max  int                  │
│  }                                  │
│  func (s) IntentType() string      │
└─────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────┐
│      组件注册 handler               │
│                                     │
│  ui.On(StepIncrement{Step: 10}){   │
│    setCount(func(c) int {          │
│      newVal := c + 10              │
│      if newVal > max { return max } │
│      return newVal                 │
│    })                               │
│  }                                  │
└─────────────────────────────────────┘
```

#### 使用示例

```go
package main

import (
    "fmt"
    "github.com/wwsheng009/mint/app"
    "github.com/wwsheng009/mint/ui"
)

// ===== 自定义 Intent 定义 =====

// StepIncrement 带步长和边界检查的递增 Intent
type StepIncrement struct {
    Step int
}

func (StepIncrement) IntentType() string { return "StepIncrement" }
func (StepIncrement) StayPressed() bool  { return true }

// ===== 组件实现 =====

func RangeCounter() ui.VNode {
    const Min, Max = 0, 100

    _, setCount, _ := ui.UseStateInt(0)

    // 注册不同步长的 handler
    ui.On(StepIncrement{Step: 1}, func() {
        setCount(func(c int) int {
            new := c + 1
            if new > Max {
                return Max
            }
            return new
        })
    })

    ui.On(StepIncrement{Step: 10}, func() {
        setCount(func(c int) int {
            new := c + 10
            if new > Max {
                return Max
            }
            return new
        })
    })

    count, _, _ := ui.UseStateInt(0)

    return ui.VStack(
        app.NewTextBuilder(fmt.Sprintf("Count: %d [%d-%d]", count, Min, Max)).Build(),
        ui.HStack(
            app.ButtonBuilder(" +1 ").OnPress(StepIncrement{Step: 1}).Build(),
            app.ButtonBuilder(" +10 ").OnPress(StepIncrement{Step: 10}).Build(),
        ),
    )
}
```

#### 高级示例：带参数的类型安全 Intent

```go
// === Intent 定义 ===
type AddTaskIntent struct {
    Title       string
    Priority    int  // 0-3: Low, Medium, High, Urgent
    AssignTo    string
}

func (AddTaskIntent) IntentType() string { return "AddTask" }
func (AddTaskIntent) StayPressed() bool  { return true }

// === 组件 ===
func TaskBoard() ui.VNode {
    // 使用 interface{} 允许任意类型的任务列表
    _, setTasks, _ := ui.UseState([]interface{}{})

    ui.On(AddTaskIntent{}, func(intent AddTaskIntent) {
        setTasks(func(tasks []interface{}) []interface{} {
            return append(tasks, intent)  // 将 intent 本身作为 task 存储
        })
    })

    return ui.VStack(
        app.ButtonBuilder("Add Task").
            OnPress(AddTaskIntent{
                Title: "Review PR", Priority: 2, AssignTo: "@self",
            }).
            Build(),
    )
}
```

---

## 五、常见问题与陷阱

### 5.1 闭包捕获陷阱（循环变量）

#### 问题
```go
// ❌ 错误：循环中捕获相同的变量引用
for i := 0; i < 5; i++ {
    ui.On(ui.SimpleIncrementIntent{}, func() {
        fmt.Println(i)  // ❌ 永远打印 4
    })
}
```

**说明**：这与 5.5 节的问题是不同的。5.1 是循环变量捕获问题，5.5 是 RegisterIntent 在组件内注册导致的 handler 引用过期问题。

#### 解决方案
```go
// ✅ 方案 1:使用循环变量
for i := 0; i < 5; i++ {
    index := i  // 捕获局部副本
    ui.On(ui.SimpleIncrementIntent{}, func() {
        fmt.Println(index)
    })
}

// ✅ 方案 2:使用函数形式
for i := 0; i < 5; i++ {
    ui.On(ui.SimpleIncrementIntent{}, func(idx int) func() {
        return func() { fmt.Println(idx) }
    }(i))
}
```

### 5.2 Hooks 顺序陷阱

#### 问题
```go
// ❌ 错误：hooks 调用顺序不一致
func Counter() ui.VNode {
    _, setCount, _ := ui.UseStateInt(0)
    ui.On(...)
    count, _, _ := ui.UseStateInt(0)  // ❌ 第二次调用，破坏顺序
    ...
}
```

#### 解决方案
```go
// ✅ 正确：hooks 所有调用在顶部，顺序一致
func Counter() ui.VNode {
    _, setCount, _ := ui.UseStateInt(0)
    _, _, _ = ui.UseStateInt(0)  // 占位，保持顺序
    ui.On(...)
    count, _, _ := ui.UseStateInt(0)  // 第三次调用
    ...
}
```

### 5.3 IntentType 冲突

#### 问题
```go
// ❌ 两个不同 Intent 使用相同的 IntentType()
type AIncrement struct{}
func (AIncrement) IntentType() string { return "Increment" }

type BIncrement struct{}
func (BIncrement) IntentType() string { return "Increment" }
// 冲突！后注册的会覆盖先注册的
```

#### 解决方案
```go
// ✅ 使用唯一的前缀
type AIncrement struct{}
func (AIncrement) IntentType() string { return "AIncrement" }

type BIncrement struct{}
func (BIncrement) IntentType() string { return "BIncrement" }
```

### 5.4 GlobalState 污染

#### 问题
```go
// ❌ 使用 GlobalState 存储运行时数据
func Component() {
    ctx := ui.GetCurrentContext()
    ctx.GlobalState["_temp"] = 123  // ❌ 污染全局状态
}
```

#### 解决方案
```go
// ✅ 使用 Hooks 状态：UseStateInt / UseStateBool
func Component() {
    _, setValue, _ := ui.UseStateInt(0)
    setValue(123)  // ✅ 组件级状态
}
```

### 5.5 RegisterIntent 闭包引用失效

#### 问题
```go
// ❌ 错误：在组件函数内注册 RegisterIntent
func App() ui.VNode {
    username, setUsername := ui.UseStateString("")  // 每次渲染都是新变量

    // 问题：handler 持有的是首次渲染时的闭包引用
    ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
        setUsername(i.Value)  // 闭包引用旧的 setter，状态更新后 UI 不刷新
        return intent.HandledResult()
    })

    return app.InputBuilder().ForField(intent.BindField("username")).Value(username).Build()
}
```

**为什么失效**：
1. `RegisterIntent` 注册的 handler 持有的是**首次渲染时**的 `setUsername` 闭包引用
2. 后续渲染时 `setUsername` 是新的引用
3. 用户输入触发的是旧 handler，调用旧 setter，状态可能更新但 UI 不会正确刷新

#### 解决方案

**方案 1：WithInit + GlobalState 动态获取（推荐 FieldChangeIntent）**

```go
// ✅ 在 WithInit 中注册，通过 GlobalState 动态获取 setter
func main() {
    ui.Run(App, ui.WithInit(func() {
        // 只注册一次，持久处理器
        ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
            switch i.Field {
            case "username":
                setUsername, _ := ctx.GetState("usernameSetter")
                if fn, ok := setUsername.(func(string)); ok {
                    fn(i.Value)  // 动态获取最新 setter
                }
            }
            return intent.HandledResult()
        })
    }))
}

func App() ui.VNode {
    username, setUsername := ui.UseStateString("")
    // 渲染时更新 GlobalState 中的 setter 引用
    ctx := ui.GetCurrentContext()
    if ctx != nil {
        ctx.GlobalState["usernameSetter"] = setUsername
    }
    return app.InputBuilder().ForField(intent.BindField("username")).Value(username).Build()
}
```

**方案 2：使用 ui.On（有去重机制，适合自定义业务 Intent）**

```go
// ✅ ui.On 有去重机制（sync.Map），多次渲染只注册一次
func App() ui.VNode {
    username, setUsername := ui.UseStateString("")

    // ui.On 可以安全地在组件内使用，不会重复注册
    ui.On(CustomUpdateIntent{}, func() {
        setUsername("new value")
    })

    return app.InputBuilder().ForField(intent.BindField("username")).Value(username).Build()
}
```

**最佳实践**：

| 场景 | 注册位置 |
|------|---------|
| `FieldChangeIntent`（表单字段变更） | `WithInit` 中注册 |
| 自定义业务 Intent（组件级状态） | `ui.On` 在组件内注册 |
| `FieldChangeIntent`（自定义 Intent 替代） | `ui.On` 在组件内注册 |

---

## 六、最佳实践

### 6.1 状态管理最佳实践

```go
// 决策树：选择合适的状态管理方式
func ManageState() {
    // Q1: 状态是否需要跨组件共享？
    if isShared {
        // 使用全局状态
        ctx := ui.GetCurrentContext()
        value := ctx.GetState("shared_key", defaultValue)
        return
    }

    // Q2: 状态是否需要复杂逻辑？
    if hasComplexLogic {
        // 使用 UseReducer（如果有的话）或自定义 Hook
        return
    }

    // 默认：使用 Hooks
    value, setValue, _ := ui.UseStateInt(0)
}
```

### 6.2 Intent 命名最佳实践

```go
// ✅ 好的命名：清晰表达意图
type ToggleSidebarIntent struct{}
type ChangeThemeIntent struct{ Theme string }
type SaveFormDataIntent struct{ FormID string }

// ❌ 避免的命名：过于通用或模糊
type ActionIntent struct{}        // 太通用
type IncrementIntent struct{}   // 可能与内置冲突（用 SimpleIncrement）
```

### 6.3 代码组织最佳实践

```go
// 推荐的项目结构
components/
├── counter/
│   ├── counter.go          // 组件定义
│   ├── intents.go          // 自定义 Intent 类型
│   └── handlers.go         // Handler 注册（可选）
└── todo/
    ├── todo.go
    ├── intents.go
    └── handlers.go

// 单文件示例（小组件）
// counter.go
package counter

import "github.com/wwsheng009/mint/ui"

// ==================== Intent 定义 ====================

type IncrementIntent struct{}
func (Increment IntentType) string { return "CounterIncrement" }
func (Increment) StayPressed() bool { return true }

// ==================== 组件定义 ====================

func Counter() ui.VNode {
    _, setCount, _ := ui.UseStateInt(0)
    ui.On(IncrementIntent{}, func() {
        setCount(func(c int) int { return c + 1 })
    })
    count, _, _ := ui.UseStateInt(0)
    return ui.Text(fmt.Sprintf("Count: %d", count))
}
```

---

## 七、迁移指南

### 7.1 从全局状态迁移到组件状态

#### 旧代码（全局状态）
```go
func OldCounter() ui.VNode {
    ctx := ui.GetCurrentContext()
    count := ctx.GetIntState("counter", 0)

    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        app.ButtonBuilder(" + ").
            OnPress(intent.Increment("counter", 1)).
            Build(),
    )
}
```

#### 新代码（组件状态）
```go
func NewCounter() ui.VNode {
    _, setCount, _ := ui.UseStateInt(0)

    ui.On(ui.SimpleIncrementIntent{}, func() {
        setCount(func(c int) int { return c + 1 })
    })

    count, _, _ := ui.UseStateInt(0)

    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        app.ButtonBuilder(" + ").
            OnPress(ui.SimpleIncrementIntent{}).
            Build(),
    )
}
```

### 7.2 从自定义 Handler 迁移到 ui.On

#### 旧代码（手动 RegisterIntent）
```go
func init() {
    intent.RegisterTypedRuntime(intentRuntime, func(ctx *intent.ActionContext, i MyIncrement) intent.IntentResult {
        // 如何访问 setCount?
        // 问题：无法访问组件局部变量
        ctx.SetState("count", i.Amount)
        return intent.HandledResult()
    })
}

type MyIncrement struct { Amount int }
func (MyIncrement) IntentType() string { return "CustomIncrement" }

func OldCounter() ui.VNode {
    ctx := ui.GetCurrentContext()
    count := ctx.GetIntState("count", 0)
    ...
}
```

#### 新代码（ui.On）
```go
func NewCounter() ui.VNode {
    _, setCount, _ := ui.UseStateInt(0)

    ui.On(MyIncrement{Amount: 1}, func(i MyIncrement) {
        setCount(func(c int) int { return c + i.Amount })
    })

    count, _, _ := ui.UseStateInt(0)
    ...
}
```

---

## 八、参考资源

### 相关文档

- [INTENT_DATA_FLOW_ANALYSIS.md](./INTENT_DATA_FLOW_ANALYSIS.md) - Intent 数据流分析
- [MVP_MIGRATION_GUIDE.md](./MVP_MIGRATION_GUIDE.md) - MVP 迁移指南
- [COMPONENT_INTENT_REVIEW.md](./COMPONENT_INTENT_REVIEW.md) - 组件 Intent 检查报告

### 示例代码

- `examples/fiber_counter_test/main.go` - 三种管理模式的完整演示
- `ui/intent.go` - 通用 Intent 和 On 函数实现
- `runtime/intent/builtin.go` - 内置 Intent 类型定义
- `runtime/intent/builtin_handlers.go` - 内置 handler 实现

---

## 附录：快速参考

### Intent API 快速参考

| 位置 | API | 用途 |
|------|-----|------|
| `ui/intent.go` | `ui.On[T](intent, handler)` | 注册组件级 Intent handler |
| `ui/intent.go` | `ui.SimpleIncrementIntent{}` | 通用递增 Intent |
| `ui/intent.go` | `ui.SimpleDecrementIntent{}` | 通用递减 Intent |
| `ui/intent.go` | `ui.SimpleToggleIntent{}` | 通用切换 Intent |
| `runtime/intent/builtin.go` | `intent.Increment(key, delta)` | 全局状态递增 |
| `runtime/intent/builtin.go` | `intent.Decrement(key, delta)` | 全局状态递减 |
| `runtime/intent/builtin.go` | `intent.Toggle(key)` | 全局状态切换 |

### Hooks API 快速参考

| Hook | 返回值 | 用途 |
|------|-------|------|
| `UseStateInt(init)` | `(int, func(func(int)int), func())` | 整数状态 |
| `UseStateBool(init)` | `(bool, func(func(bool)bool), func())` | 布尔状态 |
| `UseEffect(callback, deps)` | - | 副作用 |
| `UseMemo(compute, deps)` | `interface{}` | 缓存计算 |
| `UseRef(init)` | `Ref` | 可变引用 |
| `GetCurrentContext()` | `*ComponentContext` | 获取上下文 |

---

**文档维护**：如有疑问或发现错误，请提交 Issue 或 PR。
