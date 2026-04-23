# Store + Reducer 模式指南

**版本**: v1.0
**创建日期**: 2026-03-04
**状态**: 活跃

---

## 概述

Store + Reducer 是一种状态管理模式，实现了单向数据流：

```
Intent → Dispatcher → Reducer → Store → View
```

### 核心原则

| 原则 | 描述 |
|------|------|
| **单一真相源** | 所有状态存储在一个 Store 中 |
| **状态只读** | 状态只能通过 Reducer 修改 |
| **纯函数修改** | Reducer 是纯函数，无副作用 |

---

## 核心组件

### 1. Store[T] - 状态容器

```go
// runtime/store/store.go

type Store[T any] struct {
    state     T
    listeners []func(T)
}

// 创建 Store
store := store.NewStore(AppState{Count: 0})

// 读取状态
state := store.Get()

// 更新状态（通常由 Reducer 调用）
store.Set(newState)

// 订阅状态变化
unsubscribe := store.Subscribe(func(state AppState) {
    // 状态变化时回调
    render(state)
})
```

### 2. Reducer[T] - 状态变换函数

```go
// runtime/reducer/reducer.go

type Reducer[T any] func(state T, i intent.Intent) T

// 定义 Reducer
func AppReducer(state AppState, i intent.Intent) AppState {
    switch intent := i.(type) {
    case IncrementIntent:
        state.Count++
    case DecrementIntent:
        state.Count--
    case SetNameIntent:
        state.Name = intent.Name
    }
    return state
}
```

### 3. AppRuntime[T] - 运行时

```go
// runtime/statemachine/runtime.go

// 创建运行时
rt := statemachine.NewAppRuntime(
    AppState{Count: 0},  // 初始状态
    AppView,              // 视图函数
    AppReducer,           // Reducer
)

// 获取状态
state := rt.GetState()

// 发射 Intent
rt.Dispatch(IncrementIntent{})

// 订阅状态变化
rt.Subscribe(func(state AppState) {
    render(state)
})
```

---

## 完整示例

### 定义状态

```go
package main

type AppState struct {
    // 计数器
    Count int

    // 表单
    Username     string
    Email        string
    UsernameErr  string
    EmailErr     string

    // UI 状态
    ActiveTab string
    Loading   bool
}
```

### 定义 Intent

```go
// 计数器 Intent
type IncrementIntent struct{}
func (IncrementIntent) IntentType() string { return "Increment" }

type DecrementIntent struct{}
func (DecrementIntent) IntentType() string { return "Decrement" }

// 表单 Intent
type FieldChangeIntent struct {
    Field string
    Value string
}
func (FieldChangeIntent) IntentType() string { return "FieldChange" }

type SubmitIntent struct{}
func (SubmitIntent) IntentType() string { return "Submit" }

type ResetIntent struct{}
func (ResetIntent) IntentType() string { return "Reset" }
```

### 定义 Reducer

```go
import "github.com/wwsheng009/mint/runtime/reducer"

// 方式 1: 直接定义函数
var AppReducer reducer.Reducer[AppState] = func(state AppState, i intent.Intent) AppState {
    switch intent := i.(type) {
    case IncrementIntent:
        state.Count++
    case DecrementIntent:
        state.Count--
    case FieldChangeIntent:
        switch intent.Field {
        case "username":
            state.Username = intent.Value
            // 实时验证
            if len(state.Username) < 3 {
                state.UsernameErr = "用户名至少3字符"
            } else {
                state.UsernameErr = ""
            }
        case "email":
            state.Email = intent.Value
        }
    case SubmitIntent:
        // 提交时验证
        if len(state.Username) < 3 {
            state.UsernameErr = "用户名至少3字符"
        }
        if !isValidEmail(state.Email) {
            state.EmailErr = "邮箱格式错误"
        }
    case ResetIntent:
        return AppState{} // 重置为初始状态
    }
    return state
}

// 方式 2: 使用 Builder
var AppReducer = reducer.NewBuilder[AppState]().
    On(IncrementIntent{}, func(s AppState, _ intent.Intent) AppState {
        s.Count++
        return s
    }).
    On(DecrementIntent{}, func(s AppState, _ intent.Intent) AppState {
        s.Count--
        return s
    }).
    On(FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
        intent := i.(FieldChangeIntent)
        // ... 处理字段变更
        return s
    }).
    Build()
```

### 定义视图函数

```go
import "github.com/wwsheng009/mint/ui"

func AppView(state AppState) ui.VNode {
    return ui.VStack(
        // 计数器
        ui.Text(fmt.Sprintf("Count: %d", state.Count)),
        ui.HStack(
            ui.NewButtonBuilder("-").OnPress(DecrementIntent{}).Build(),
            ui.NewButtonBuilder("+").OnPress(IncrementIntent{}).Build(),
        ),

        ui.Text(""),

        // 表单
        ui.Text("Username:"),
        ui.Input().Value(state.Username).OnChange(FieldChangeIntent{Field: "username"}),
        ui.Text(state.UsernameErr).FgColor("red"),

        ui.Text("Email:"),
        ui.Input().Value(state.Email).OnChange(FieldChangeIntent{Field: "email"}),
        ui.Text(state.EmailErr).FgColor("red"),

        ui.HStack(
            ui.NewButtonBuilder("Submit").OnPress(SubmitIntent{}).Build(),
            ui.NewButtonBuilder("Reset").OnPress(ResetIntent{}).Build(),
        ),
    )
}
```

### 初始化运行时

```go
import (
    "github.com/wwsheng009/mint/runtime/statemachine"
    "github.com/wwsheng009/mint/ui"
)

func main() {
    // 创建运行时
    rt := statemachine.NewAppRuntime(
        AppState{Count: 0},
        AppView,
        AppReducer,
        statemachine.WithMaxHistory(50), // 启用时间旅行调试
    )

    // 订阅状态变化
    rt.Subscribe(func(state AppState) {
        // 触发重新渲染
        renderApp(state)
    })

    // 运行应用
    ui.RunApp(rt)
}
```

---

## 高级用法

### 组合 Reducer

```go
// 多个 Reducer 组合
var CombinedReducer = reducer.Compose(
    CounterReducer,
    FormReducer,
    UIReducer,
)
```

### 中间件

```go
// 日志中间件
func LoggingMiddleware[T any](next reducer.Reducer[T]) reducer.Reducer[T] {
    return func(state T, i intent.Intent) T {
        fmt.Printf("Intent: %s\n", i.IntentType())
        newState := next(state, i)
        fmt.Printf("State changed: %+v\n", newState)
        return newState
    }
}

// 应用中间件
var AppReducer = reducer.WithMiddleware(
    BaseReducer,
    LoggingMiddleware[AppState],
)
```

### 时间旅行调试

```go
rt := statemachine.NewAppRuntime(
    AppState{},
    AppView,
    AppReducer,
    statemachine.WithMaxHistory(100),
)

// 获取历史记录
history := rt.History()

// 撤销
rt.Undo()

// 跳转到特定状态
rt.JumpTo(5)
```

### Computed Values

```go
import "github.com/wwsheng009/mint/runtime/store"

// 创建计算值
usernameLength := store.NewComputed(store, func(s AppState) int {
    return len(s.Username)
})

// 获取计算值
length := usernameLength.Get()
```

---

## 迁移指南

### 从 UseState 迁移

**旧代码 (UseState)**:
```go
func Counter() ui.VNode {
    count, setCount, _ := ui.UseStateInt(0)

    ctx := ui.GetCurrentContext()
    if ctx != nil {
        ctx.SetState("setCount", setCount)
    }

    ui.On(IncrementIntent{}, func(ctx *intent.ActionContext) {
        // 从 context 读取 setter
    })

    return ui.Text(fmt.Sprintf("Count: %d", count))
}
```

**新代码 (Store + Reducer)**:
```go
// 定义状态
type AppState struct {
    Count int
}

// 定义 Reducer
var AppReducer = reducer.NewBuilder[AppState]().
    On(IncrementIntent{}, func(s AppState, _ intent.Intent) AppState {
        s.Count++
        return s
    }).
    Build()

// 视图函数
func AppView(state AppState) ui.VNode {
    return ui.Text(fmt.Sprintf("Count: %d", state.Count))
}
```

---

## 最佳实践

### 1. 状态设计

```go
// ✅ 扁平化状态
type AppState struct {
    User    User
    Posts   []Post
    UI      UIState
}

// ❌ 避免嵌套过深
type AppState struct {
    Data struct {
        User struct {
            Profile struct {
                Name string
            }
        }
    }
}
```

### 2. Reducer 纯函数

```go
// ✅ 纯函数 - 无副作用
func AppReducer(state AppState, i intent.Intent) AppState {
    newState := state // 复制
    newState.Count++
    return newState
}

// ❌ 避免副作用
func AppReducer(state AppState, i intent.Intent) AppState {
    saveToDatabase(state) // 不要这样做！
    return state
}
```

### 3. Intent 命名

```go
// ✅ 使用动词 + 名词
type IncrementCounterIntent struct{}
type SubmitFormIntent struct{}
type LoadUserDataIntent struct{}

// ❌ 避免模糊命名
type UpdateIntent struct{}  // 更新什么？
type ChangeIntent struct{}  // 改变什么？
```

---

## 相关文档

- [INTENT_HANDLER_MIGRATION.md](../migration/INTENT_HANDLER_MIGRATION.md) - Intent Handler 迁移指南
- [REFACTOR_PLAN.md](/docsArchive/REFACTOR_PLAN.md) - 完整重构计划
- [store.md](../README.md) - Store 详细设计

---

**最后更新**: 2026-03-04
