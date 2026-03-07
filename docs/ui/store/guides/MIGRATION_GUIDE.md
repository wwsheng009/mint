# UseState 迁移指南

**版本**: v1.0
**创建时间**: 2026-03-04
**状态**: 当前架构 → Store + Reducer

---

## ⚠️ 重要提示：GlobalState 已弃用

**状态**: `ComponentContext.GlobalState` 及相关方法已标记为 **Deprecated**

请优先迁移到 **Store + Reducer** 架构，该架构提供：
- ✅ 类型安全的状态管理
- ✅ 无类型断言的代码
- ✅ 单一数据源原则
- ✅ 更清晰的代码结构

**详细说明**: [GlobalState 弃用公告](../GLOBALSTATE_DEPRECATION.md) | [混合模式指南](../hybrid/STATE_MANAGEMENT_GUIDE.md)

---

## 概述

本指南帮助你将使用 `UseState` 的代码迁移到推荐的 **Store + Reducer 架构**。

---

## 为什么迁移？

| UseState 架构 | Store + Reducer 架构 |
|----------------|---------------------|
| ❌ 复杂的 setter marshaling | ✅ 简单的状态读取 |
| ❌ 需要类型断言 | ✅ 无类型断言 |
| ❌ 时序依赖（WithInit） | ✅ 无时序依赖 |
| ❌ 多源状态（组件 + GlobalState） | ✅ 单一状态源（Store） |
| ❌ 难以测试（闭包依赖） | ✅ 易于测试（纯函数） |

---

## 快速对比

### 使用 UseState（旧方式）❌

```go
func App() ui.VNode {
    // 1. 创建状态
    username, setUsername := ui.UseStateString("")
    email, setEmail := ui.UseStateString("")
    
    // 2. 保存 setter 到 GlobalState
    ctx := ui.GetCurrentContext()
    if ctx != nil {
        ctx.GlobalState["username-setter"] = setUsername
        ctx.GlobalState["email-setter"] = setEmail
    }

    // 3. 注册 handler
    return ui.WithInit(func() {
        ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
            switch i.Field {
            case "username":
                if fn, ok := ctx.GetState("username-setter"); ok {
                    if setter, ok := fn.(func(string)); ok {
                        setter(i.Value)
                    }
                }
            // ...
            }
            return intent.HandledResult()
        })
    },
        // UI 构建...
        ui.NewInputBuilder().ForField(intent.BindField("username")).Value(username).Build(),
    )
}
```

**问题**：
- 需要 5+ 步
- 需要类型断言
- setter marshaling 复杂
- 时序依赖（WithInit）

---

### 使用 Store + Reducer（新方式）✅

```go
// 1. 定义 State
type AppState struct {
    Username string
    Email    string
    ClickCount int
}

// 2. 创建全局 Store
var appStore *store.Store[AppState]

func initStore() {
    appStore = store.NewStore(AppState{
        Username: "",
        Email: "",
        ClickCount: 0,
    })
}

// 3. 定义 Reducer
var appReducer = reducer.NewBuilder[AppState]().
    On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
        if fieldChange, ok := i.(intent.FieldChangeIntent); ok {
            switch fieldChange.Field {
            case "username":
                s.Username = fieldChange.Value
            case "email":
                s.Email = fieldChange.Value
            }
        }
        return s
    }).
    On(ClickButtonIntent{}, func(s AppState, i intent.Intent) AppState {
        s.ClickCount++
        return s
    })

// 4. 注册 handlers
appReducer.RegisterToGlobal(appStore)

// 5. 视图读取 Store
func App() ui.VNode {
    state := appStore.Get()  // 每次渲染时获取最新值
    
    return ui.VStack(
        ui.NewInputBuilder().
            ForField(intent.BindField("username")).  // 自动发射 FieldChangeIntent
            Value(state.Username).                       // 显示 Store 中的值
            Placeholder("Username").
            Build(),
    )
}
```

**优点**：
- 只需要 3 步
- 无类型断言
- 无 setter marshaling
- 无时序依赖

---

## 详细迁移步骤

### 步骤 1: 定义 Application State

将所有组件状态合并到一个 `AppState` 结构体中：

```go
// Before: 多个独立的 UseState
username, setUsername := ui.UseStateString("")
email, setEmail := ui.UseStateString("")
count, setCount, _ := ui.UseStateInt(0)
checked, setChecked := ui.UseStateBool(false)

// After: 单一 State 结构
type AppState struct {
    Username  string
    Email     string
    Count     int
    Checked   bool
    // ... 其他状态
}
```

---

### 步骤 2: 创建全局 Store

```go
import "github.com/wwsheng009/mint/runtime/store"

var appStore *store.Store[AppState]

func initStore() {
    appStore = store.NewStore(AppState{
        Username: "",
        Email:    "",
        Count:    0,
        Checked:  false,
    })
}
```

**注意**: Store 应该是全局单例，在 `init()` 函数中创建一次。

---

### 步骤 3: 定义 Reducer

使用 `ReducerBuilder` 创建状态转换逻辑：

```go
import "github.com/wwsheng009/mint/runtime/reducer"
import "github.com/wwsheng009/mint/runtime/intent"

var appReducer = reducer.NewBuilder[AppState]().
    // 处理按钮点击
    On(ClickButtonIntent{}, func(s AppState, i intent.Intent) AppState {
        s.Count++
        return s
    }).
    // 处理输入框字段变更（FieldChangeIntent 自动发射）
    On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
        if fieldChange, ok := i.(intent.FieldChangeIntent); ok {
            switch fieldChange.Field {
            case "username":
                s.Username = fieldChange.Value
            case "email":
                s.Email = fieldChange.Value
            }
        }
        return s
    })
```

**关键点**：
- Reducer 是纯函数，无副作用
- Reducer 直接修改 State 的副本
- 使用 `s.Field = value` 更新字段，不需要类型断言

---

### 步骤 3.1: 优化 FieldBinding（推荐）

使用 **FieldBinding API** 可以消除 switch-case 硬编码，让代码更简洁：

```go
import "strconv"
import "github.com/wwsheng009/mint/runtime/reducer"
import "github.com/wwsheng009/mint/runtime/intent"

// 使用 FieldMap 替代 switch-case
var appReducer = reducer.BindField(reducer.NewBuilder[AppState]()).
    BindFieldMap(map[string]func(AppState, string) AppState{
        // 所有字段集中定义，单一处理器
        "username": func(s AppState, val string) AppState {
            s.Username = val
            return s
        },
        "email": func(s AppState, val string) AppState {
            s.Email = val
            return s
        },
        "count": func(s AppState, val string) AppState {
            if v, err := strconv.Atoi(val); err == nil {
                s.Count = v
            }
            return s
        },
        "checked": func(s AppState, val string) AppState {
            s.Checked = val == "true"
            return s
        },
    }).
    GetBuilder().
    // 自定义 Intent
    On(ClickButtonIntent{}, func(s AppState, i intent.Intent) AppState {
        s.Count++
        return s
    })
```

**优势**：
| 传统方式 | FieldBinding 方式 |
|----------|------------------|
| switch-case 硬编码 | 字段映射表 |
| 类型断言 | 泛型类型安全 |
| 分散的字段逻辑 | 集中的字段定义 |
| 需要手动类型转换 | 自动类型转换（BindIntField、BindBoolField） |

**详细优化指南**: [FIELD_BINDING_OPTIMIZATION.md](./FIELD_BINDING_OPTIMIZATION.md)

---

### 步骤 4: 注册 Handlers

使用 `BuildAndRegister` 自动注册 handlers：

```go
func main() {
    initStore()  // 初始化 Store
    
    // 注册 handlers（在 Run 之前）
    appReducer.RegisterToGlobal(appStore)
    
    // 启动应用
    err := ui.Run(App, ui.WithWidth(60), ui.WithHeight(30))
    if err != nil {
        panic(err)
    }
}
```

**替代方案**：如果需要在 `WithInit` 中注册其他东西，可以这样做：

```go
err := ui.Run(App,
    ui.WithWidth(60),
    ui.WithHeight(30),
    ui.WithInit(func() {
        // 注册 Store handlers
        appReducer.RegisterToGlobal(appStore)
        // 注册其他 handlers
        ui.RegisterIntent(...)
    }),
)
```

---

### 步骤 5: 重写组件视图

从 `UseState` 迁移到 `Store.Get()`：

```go
// Before: 使用 UseState
func App() ui.VNode {
    username, _ := ui.UseStateString("")
    email, _ := ui.UseStateString("")
    
    return ui.NewInputBuilder().
        ForField(intent.BindField("username")).
        Value(username).
        Build()
}

// After: 使用 Store
func App() ui.VNode {
    state := appStore.Get()  // 从 Store 读取最新状态
    
    return ui.NewInputBuilder().
        ForField(intent.BindField("username")).
        Value(state.Username).  // 直接使用 State 中的字段
        Build()
}
```

**关键变化**：
- 移除 `UseState...` 调用
- 移除 setter 定义
- 移除 `GlobalState["setter"]` 保存
- 改用 `appStore.Get()` 读取状态

---

## 核心模式对比

### Pattern 1: 输入框字段

**Before (UseState)**:
```go
username, setUsername := ui.UseStateString("")

ui.WithInit(func() {
    ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
        if fn, ok := ctx.GetState("username-setter"); ok {
            if setter, ok := fn.(func(string)); ok {
                setter(i.Value)
            }
        }
        return intent.HandledResult()
    })
}, ...)

ui.NewInputBuilder().
    ForField(intent.BindField("username")).
    Value(username).
    Build()
```

**After (Store + Reducer)**:
```go
// State 定义
type AppState struct {
    Username string
}

// Reducer
appReducer := reducer.NewBuilder[AppState]().
    On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
        s.Username = i.(intent.FieldChangeIntent).Value
        return s
    }).RegisterToGlobal(appStore)

// 组件
state := appStore.Get()
ui.NewInputBuilder().
    ForField(intent.BindField("username")).  // 自动发射 FieldChangeIntent
    Value(state.Username).                      // 从 Store 读取
    Build()
```

---

### Pattern 2: 计数器

**Before (UseState + Functional Update)**:
```go
clickCount, setClickCount, _ := ui.UseStateInt(0)

ui.WithInit(func() {
    ctx.GlobalState["setClickCount"] = setClickCount
    ui.RegisterIntent(func(ctx *intent.ActionContext, i ClickButtonIntent) intent.IntentResult {
        if fn, ok := ctx.GetState("setClickCount"); ok {
            if setter, ok := fn.(func(interface{})); ok {
                setter(func(c int) int { return c + 1 })
            }
        }
        return intent.HandledResult()
    })
}, ...)

ui.NewButtonBuilder("+").OnPress(ClickButtonIntent{}).Build()
```

**After (Store + Reducer)**:
```go
// State 定义
type AppState struct {
    ClickCount int
}

// Reducer
appReducer := reducer.NewBuilder[AppState]().
    On(ClickButtonIntent{}, func(s AppState, i intent.Intent) AppState {
        s.ClickCount++
        return s
    }).RegisterToGlobal(appStore)

// 组件
state := appStore.Get()
ui.NewButtonBuilder(fmt.Sprintf("Count: %d", state.ClickCount)).
    OnPress(ClickButtonIntent{}).
    Build()
```

---

### Pattern 3: Checkbox

**Before (UseState)**:
```go
checked, setChecked := ui.UseStateBool(false)

ui.WithInit(func() {
    ctx.GlobalState["checked-setter"] = setChecked
    ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
        if i.Field == "checked" {
            if fn, ok := ctx.GetState("checked-setter"); ok {
                value := i.Value == "true"
                if setter, ok := fn.(func(bool)); ok {
                    setter(value)
                }
            }
        }
        return intent.HandledResult()
    })
}, ...)

checkbox.NewBuilder().
    Label("Option").
    ForField(intent.BindField("checked")).
    Checked(checked).
    Build()
```

**After (Store + Reducer)**:
```go
// State 定义
type AppState struct {
    Checked string  // 存储为 "true"/"false"
}

// Reducer
appReducer := reducer.NewBuilder[AppState]().
    On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
        if fieldChange, ok := i.(intent.FieldChangeIntent); ok {
            if fieldChange.Field == "checked" {
                s.Checked = fieldChange.Value
            }
        }
        return s
    }).RegisterToGlobal(appStore)

// 组件
state := appStore.Get()
checkbox.NewBuilder().
    Label("Option").
    ForField(intent.BindField("checked")).
    Checked(state.Checked == "true").  // 字符串转布尔
    Build()
```

---

## 常见问题

### Q1: State 在组件之间如何共享？

**A**: Store 是全局单例，所有组件都从同一个 Store 读取：

```go
var appStore *store.Store[AppState]

func ComponentA() ui.VNode {
    state := appStore.Get()
    return ui.Text(state.Username)
}

func ComponentB() ui.VNode {
    state := appStore.Get()
    return ui.Text(state.Email)
}
```

### Q2: 如何处理多个页面/应用的状态？

**A**: 使用不同的 Store 实例：

```go
// 不同应用使用不同 Store
var app1Store = store.NewStore(App1State{})
var app2Store = store.NewStore(App2State{})
```

### Q3: 如何重置状态？

**A**: 直接调用 `Store.Set()`:

```go
appStore.Set(AppState{
    Username: "",
    Email: "",
    Count: 0,
})
```

### Q4: 如何处理复杂的嵌套对象？

**A**: 在 State 中使用嵌套结构：

```go
type AppState struct {
    Form FormData
    Theme ThemeConfig
}

type FormData struct {
    Username string
    Email    string
}

// Reducer
On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
    s.Form.Username = i.(intent.FieldChangeIntent).Value
    return s
})
```

---

## 迁移检查清单

在完成迁移后，确认以下事项：

- [ ] 所有的 `UseState...` 调用已移除
- [ ] 所有 setter 定义已移除
- [ ] 所有 `GlobalState["setter"]` 保存已移除
- [ ] 所有类型断言已移除
- [ ] Reducer 已定义并使用 `BuildAndRegister` 注册
- [ ] 所有使用 `Value()` 的地方改为从 Store 读取
- [ ] 所有组件都能正确显示 Store 中的状态

---

## 示例代码

### 完整示例

参见以下示例了解完整的 Store + Reducer 实现：

- **focus_switching_demo/main.go** - 已迁移到 Store + Reducer
- **store_reducer_demo/main.go** - Store + Reducer 参考实现

---

## 相关文档

- [Store 架构指南](STORE_ARCHITECTURE.md)
- [Reducer 指南](REDUCER_GUIDE.md)
- [最佳实践](BEST_PRACTICES.md)
- [API 参考](API_REFERENCE.md)

---

## 获取帮助

如果有问题，请参考：

1. [focus_switching_demo](../../../examples/fiber_firsts/focus_switching_demo/) - 最新迁移的示例
2. [store_reducer_demo](../../../examples/store_reducer_demo/) - 完整的 Store + Reducer 示例
3. [常见问题](FAQ.md) - 迁移过程中的常见问题
