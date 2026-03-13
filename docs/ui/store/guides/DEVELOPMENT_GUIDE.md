# Store + Reducer 开发指南

**版本**: v0.11
**最后更新**: 2026-03-05

---

## 概述

本文档指导如何在 Mint UI 中使用 Store + Reducer 架构构建应用程序。

### 架构图

```
┌───────────────────────────────────────────────────────────┐
│                      视图层 (View)                          │
│  ui.OnPress() → Instance → Intent                          │
└─────────────────┬─────────────────────────────────────────┘
                  │
                  ▼
┌───────────────────────────────────────────────────────────┐
│              调度器层 (Registry/Dispatcher)                │
│  intent.DefaultRegistry.Register()                          │
└─────────────────┬─────────────────────────────────────────┘
                  │
                  ▼
┌───────────────────────────────────────────────────────────┐
│              Reducer 层 (Pure Function)                      │
│  Reducer[T](state, intent) → newState                       │
└─────────────────┬─────────────────────────────────────────┘
                  │
                  ▼
┌───────────────────────────────────────────────────────────┐
│              Store 层 (Single Source of Truth)              │
│  Store[T].Set(newState) → Notify subscribers                │
└───────────────────────────────────────────────────────────┘
```

---

## 快速开始

### 1. 定义 AppState

```go
// ✅ 推荐：扁平结构，包含所有状态
type AppState struct {
    // 表单字段
    Username string
    Email    string
    Password string
    Age      int

    // UI 状态
    IsLoading bool
    ShowModal bool
    Step      int
}
```

### 2. 定义 Intent

```go
// 自定义 Intent
type IncrementIntent struct {
    Amount int
}
func (IncrementIntent) IntentType() string { return "Increment" }

// 系统内置 Intent
// - intent.FieldChangeIntent - 字段变更
// - intent.SubmitIntent - 提交
// - ...
```

### 3. 定义 Reducer

**方案 1: 基础方式**

```go
var appReducer = reducer.NewBuilder[AppState]().
    On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
        s.Count++
        return s
    }).
    On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
        fci := i.(intent.FieldChangeIntent)
        switch fci.Field {
        case "username":
            s.Username = fci.Value
        case "password":
            s.Password = fci.Value
        }
        return s
    })
```

**方案 2: 优化方式（使用 FieldMap，推荐）**

```go
import "strconv"
import "github.com/wwsheng009/mint/runtime/reducer"

// 使用 FieldMap 消除 switch-case 硬编码
var appReducer = reducer.BindField(reducer.NewBuilder[AppState]()).
    BindFieldMap(map[string]func(AppState, string) AppState{
        // 所有字段集中定义，单一处理器
        "username": func(s AppState, val string) AppState {
            s.Username = val
            return s
        },
        "password": func(s AppState, val string) AppState {
            s.Password = val
            return s
        },
        "age": func(s AppState, val string) AppState {
            if v, err := strconv.Atoi(val); err == nil {
                s.Age = v
            }
            return s
        },
    }).
    GetBuilder().
    On(SubmitIntent{}, func(s AppState, i intent.Intent) AppState {
        // 提交逻辑
        return s
    })
```

**详细 FieldBinding 优化指南**: [FIELD_BINDING_OPTIMIZATION.md](../optimization/FIELD_BINDING_OPTIMIZATION.md)

### 4. 创建 Store

```go
var appStore = store.NewStore(AppState{
    Username: "",
    Email:    "",
    Age:      0,
})
```

### 5. 注册 Handlers

```go
func main() {
    // 自动注册所有 handlers
    appReducer.RegisterToGlobal(appStore)

    // 运行应用
    ui.Run(App, ui.WithTitle("App"))
}
```

### 6. 构建视图

```go
func App() ui.VNode {
    // 每次渲染从 Store 读取最新状态
    state := appStore.Get()

    return ui.VStack(
        ui.NewTextBuilder(state.Username).Build(),
        ui.NewInputBuilder().
            ForField(intent.BindField("username")).
            Value(state.Username).
            Build(),
    )
}
```

---

## 最佳实践

### 1. AppState 设计

**✅ 推荐：扁平结构**

```go
type AppState struct {
    // 表单字段
    Username string
    Email    string

    // UI 状态
    IsLoading bool
}
```

**❌ 避免：深层嵌套**

```go
type AppState struct {
    Form FormData  // 嵌套结构
    UI   UIState   // 嵌套结构
}
```

### 2. Reducer 设计

**✅ 推荐：纯函数**

```go
appReducer.On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
    s.Count++  // 直接修改返回新状态
    return s
})
```

**❌ 避免：副作用**

```go
appReducer.On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
    fmt.Println("Logging...")  // 副作用
    s.Count++
    appStore.Set(s)           // 直接调用 Store（错误！）
    return s
})
```

### 3. 字段绑定

**✅ 推荐：使用 FieldMap**

```go
var appReducer = reducer.BindField(reducer.NewBuilder[AppState]()).
    BindFieldMap(map[string]func(AppState, string) AppState{
        "username": func(s AppState, val string) AppState {
            s.Username = val
            return s
        },
        "email": func(s AppState, val string) AppState {
            s.Email = val
            return s
        },
    })
```

### 4. 组件设计

**✅ 推荐：每次渲染读取最新状态**

```go
func App() ui.VNode {
    state := appStore.Get()  // 每次渲染读取最新状态
    return ui.VStack(...)
}
```

**❌ 避免：捕获闭包**

```go
func App() ui.VNode {
    state := appStore.Get()
    stateCopy := state  // 复制状态（错误！）
    return ui.VStack(func() ui.VNode {
        // 使用过时的 stateCopy
    })
}
```

---

## 常见问题

### Q1: 如何处理异步操作？

**方案 1: 使用状态机模式**

```go
type AppState struct {
    Loading  bool
    Error    string
    Data     string
}

type FetchDataIntent struct{}
type FetchSuccessIntent struct {
    Data string
}
type FetchErrorIntent struct {
    Error string
}

// Reducer
appReducer.On(FetchDataIntent{}, func(s AppState, i intent.Intent) AppState {
    s.Loading = true
    return s
})

// 使用 goroutine
func fetchData() {
    go func() {
        data, err := api.Fetch()
        if err != nil {
            ui.Dispatch(FetchErrorIntent{Error: err.Error()})
        } else {
            ui.Dispatch(FetchSuccessIntent{Data: data})
        }
    }()
}
```

### Q2: 如何处理表单验证？

**方案 1: 在 AppState 中存储错误**

```go
type AppState struct {
    Username string
    Email    string
    UsernameErr string
    EmailErr    string
}

// Reducer
appReducer.On(SubmitIntent{}, func(s AppState, i intent.Intent) AppState {
    // 验证逻辑
    if s.Username == "" {
        s.UsernameErr = "Username is required"
    } else {
        s.UsernameErr = ""
    }

    if s.Email == "" {
        s.EmailErr = "Email is required"
    } else {
        s.EmailErr = ""
    }

    return s
})

// 组件
usernameInput := ui.NewInputBuilder().
    ForField(intent.BindField("username")).
    Value(state.Username).
    Build()

if state.UsernameErr != "" {
    ui.NewTextBuilder(state.UsernameErr).
        FgColor("red").
        Build()
}
```

### Q3: 如何优化性能？

**方案 1: 使用 Compute 和 Selector**

```go
// 选择器 - 计算派生状态
type AppState struct {
    Items []string
}

// 计算值 - 自动缓存
itemCount := appStore.Compute(func(s AppState) int {
    return len(s.Items)
})

// 使用
count := itemCount.Get()  // 自动缓存，重复调用不重复计算
```

**方案 2: 避免不必要的渲染**

```go
// ✅ 使用 computed 避免重复计算
var usernameLength = appStore.Compute(func(s AppState) int {
    return len(s.Username)
})

// ❌ ❌ 避免在渲染时重复计算
func App() ui.VNode {
    state := appStore.Get()

    length := len(state.Username)  // 每次渲染都计算
    // ...
}
```

### Q4: 如何调试？

**方案 1: 使用中间件和日志**

```go
// 日志中间件
appReducer := reducer.WithMiddleware(
    reducer.NewBuilder[AppState](),
    reducer.LoggingMiddleware[AppState](func(state AppState, i intent.Intent, newState AppState) {
        log.Printf("Intent: %v\nState: %+v → %+v\n", i, state, newState)
    }),
).Build()
```

**方案 2: 使用 AppRuntime 时间旅行**

```go
runtime := NewAppRuntime(AppState{}, App, appReducer)

// 跳转到历史状态
runtime.JumpTo(0)

// 撤销
runtime.Undo()

// 查看历史
history := runtime.History()
```

---

## 迁移到 Store + Reducer

### 从 UseState 迁移

**旧方式（UseState）**

```go
func App() ui.VNode {
    username, setUsername := ui.UseStateString("")
    email, setEmail := ui.UseStateString("")

    // setter 手动注册
    ctx := ui.GetCurrentContext()
    ctx.GlobalState["usernameSetter"] = setUsername
    ctx.GlobalState["emailSetter"] = setEmail

    // Handler 手动注册
    ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
        switch i.Field {
        case "username":
            setting, _ := ctx.GetState("usernameSetter")
            if setter, ok := setting.(func(string)); ok {
                setter(i.Value)
            }
        // ...
        }
    })

    return ui.VStack(...)
}
```

**新方式（Store + Reducer）**

```go
type AppState struct {
    Username string
    Email    string
}

var appStore = store.NewStore(AppState{})
var appReducer = reducer.BindField(reducer.NewBuilder[AppState]()).
    BindFieldMap(map[string]func(AppState, string) AppState{
        "username": func(s AppState, val string) AppState {
            s.Username = val
            return s
        },
        "email": func(s AppState, val string) AppState {
            s.Email = val
            return s
        },
    })

func App() ui.VNode {
    state := appStore.Get()
    return ui.VStack(
        ui.NewInputBuilder().
            ForField(intent.BindField("username")).
            Value(state.Username).
            Build(),
    )
}
```

**详细迁移指南**: [MIGRATION_GUIDE.md](./MIGRATION_GUIDE.md)

---

## 代码示例

### 完整示例：简单的表单

```go
package main

import (
    "github.com/wwsheng009/mint/runtime/intent"
    "github.com/wwsheng009/mint/runtime/reducer"
    "github.com/wwsheng009/mint/runtime/store"
    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/ui/components/input"
)

// State
type AppState struct {
    Username string
    Email    string
}

// Intent
type SubmitIntent struct{}

// Store
var appStore = store.NewStore(AppState{})

// Reducer (使用 FieldMap)
var appReducer = reducer.BindField(reducer.NewBuilder[AppState]()).
    BindFieldMap(map[string]func(AppState, string) AppState{
        "username": func(s AppState, val string) AppState {
            s.Username = val
            return s
        },
        "email": func(s AppState, val string) AppState {
            s.Email = val
            return s
        },
    }).
    GetBuilder().
    On(SubmitIntent{}, func(s AppState, i intent.Intent) AppState {
        fmt.Printf("Submit: %s, %s\n", s.Username, s.Email)
        return s
    })

// App
func App() ui.VNode {
    state := appStore.Get()

    return ui.VStack(
        ui.NewInputBuilder().
            ForField(intent.BindField("username")).
            Value(state.Username).
            Placeholder("Username").
            Build(),
        ui.NewInputBuilder().
            ForField(intent.BindField("email")).
            Value(state.Email).
            Placeholder("Email").
            Build(),
        ui.NewButtonBuilder("Submit").
            OnPress(SubmitIntent{}).
            Build(),
    )
}

// Main
func main() {
    appReducer.RegisterToGlobal(appStore)
    ui.Run(App, ui.WithTitle("App"))
}
```

---

## 总结

### Store + Reducer 优势

| 优势 | 说明 |
|------|------|
| 单一状态源 | 所有状态存储在一个 Store 中 |
| 单向数据流 | Intent → Reducer → Store → View |
| 类型安全 | 泛型支持，编译期类型检查 |
| 易测试 | 纯函数，易于单元测试 |
| 易调试 | 中间件、时间旅行等调试工具 |
| 可扩展 | 中间件、Plugin 支持 |

### 最佳实践

1. ✅ 使用 **FieldMap** 处理字段变更
2. ✅ **纯函数** Reducer
3. ✅ **扁平结构** AppState
4. ✅ **每次渲染** 读取最新状态
5. ✅ **自动注册** handlers

### 相关文档

- **API 参考**: [API_REFERENCE.md](../api/API_REFERENCE.md)
- **迁移指南**: [MIGRATION_GUIDE.md](./MIGRATION_GUIDE.md)
- **状态评估**: [CURRENT_STATUS.md](/docsArchive/status/CURRENT_STATUS.md)
- **字段绑定优化**: [FIELD_BINDING_OPTIMIZATION.md](../optimization/FIELD_BINDING_OPTIMIZATION.md)
- **迁移进度**: [MIGRATION_PROGRESS.md](/docsArchive/status/MIGRATION_PROGRESS.md)

---

**文档创建**: 2026-03-05
**状态**: 完成 ✅
