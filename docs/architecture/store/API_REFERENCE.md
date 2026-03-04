# API 参考 - Store + Reducer

**版本**: v1.0  
**创建时间**: 2026-03-04

---

## 目录

- [状态管理 API](#状态管理-api)
- [Intent API](#intent-api)
- [组件 API](#组件-api)
- [已废弃 API](#已废弃-api)
- [完整示例](#完整示例)

---

## 状态管理 API

### Store[T]

**导入**: `github.com/wwsheng009/mint/runtime/store`

#### NewStore

```go
func NewStore[T any](initial T) *Store[T]
```

创建一个新的 Store 初始状态。

**参数**:
- `initial T`: 初始状态值

**返回**:
- `*Store[T]`: Store 实例

**示例**:
```go
type AppState struct {
    Count int
    Username string
}

appStore := store.NewStore(AppState{
    Count: 0,
    Username: "",
})
```

---

#### Get

```go
func (s *Store[T]) Get() T
```

获取 Store 中的当前状态。这是读取状态的唯一方式。

**返回**:
- `T`: 当前状态值

**示例**:
```go
state := appStore.Get()
fmt.Printf("Count: %d", state.Count)
```

---

#### Set

```go
func (s *Store[T]) Set(next T)
```

更新 Store 中的状态。这会通知所有订阅者。通常由 Reducer 调用。

**参数**:
- `next T`: 新的状态值

**示例**:
```go
appStore.Set(AppState{
    Count: 1,
    Username: "john",
})

// 通常由 Reducer 使用
appReducer := reducer.NewBuilder[AppState]().
    On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
        s.Count++
        return s  // 返回状态副本，Store 内部会调用 Set()
    })
```

---

#### Update

```go
func (s *Store[T]) Update(fn func(T) T)
```

使用函数更新状态。这是原子操作，避免并发问题。

**参数**:
- `fn func(T) T`: 更新函数，接收当前状态，返回新状态

**示例**:
```go
appStore.Update(func(state AppState) AppState {
    state.Count++
    return state
})
```

---

#### Subscribe

```go
func (s *Store[T]) Subscribe(callback func(T)) func()
```

订阅状态变化。当状态改变时，回调会被调用。

**参数**:
- `callback func(T)`: 状态变化回调函数

**返回**:
- `func()`: 取消订阅函数

**示例**:
```go
unsubscribe := appStore.Subscribe(func(oldState, newState AppState) {
    fmt.Printf("State changed: %+v -> %+v", oldState, newState)
})

// 取消订阅
unsubscribe()
```

---

## Intent API

### FieldChangeIntent

**用途**: 表单字段变更（UI 组件自动发射）

**定义**:
```go
type FieldChangeIntent struct {
    Field string  // 字段名
    Value string  // 字段值
}

func (FieldChangeIntent) IntentType() string { return "FieldChange" }
```

**Handler 位置**: `intent.HandleFieldChange`

**使用方式**: 组件通过 `ForField()` 自动发射

---

### SetStateIntent

**用途**: 设置全局状态值（直接访问 GlobalState）

**定义**:
```go
type SetStateIntent struct {
    Key   string
    Value interface{}
}

func (SetStateIntent) IntentType() string { return "SetState" }
```

**Handler 位置**: `intent.HandleSetState`

**注意**: 在 Store + Reducer 架构中不推荐使用，推荐使用 FieldChangeIntent

---

### ToggleIntent

**用途**: 切换布尔状态

**定义**:
```go
type ToggleIntent struct {
    Key string
}

func (ToggleIntent) IntentType() string { return "Toggle" }
```

**Handler 位置**: `intent.HandleToggle`

**使用方式**: 组件通过 `OnToggle()` 或手动发射

---

### IncrementIntent

**用途**: 递增数值状态

**定义**:
```go
type IncrementIntent struct {
    Key   string
    Delta int
}

func (IncrementIntent) IntentType() string { return "Increment" }
```

**Handler 位置**: `intent.HandleIncrement`

**使用方式**: 组件通过发射意图或手动发射

---

## 组件 API

### ForField

**适用组件**: Input, Textarea, Checkbox, Select, Tabs

**用途**: 绑定到 State 字段，并自动发射 `FieldChangeIntent`

**签名**:
```go
// Input
inputComp.NewBuilder().ForField(intent.BindField("field")).Value(state.Field).Build()

// Checkbox
checkboxComp.NewBuilder().ForField(intent.BindField("field")).Checked(state.Field == "true").Build()

// Select
selectComp.NewBuilder().ForField(intent.BindField("field")).Value(state.Field).Build()
```

**示例**:
```go
inputComp.NewBuilder().
    ForField(intent.BindField("username")).
    Value(state.Username).
    Placeholder("Username").
    Build()
```

---

### OnPress

**适用组件**: Button

**用途**: 设置按钮按压时发射的 Intent

**签名**:
```go
buttonComp.NewBuilder("Button").
    OnPress(CustomIntent{}).
    Build()
```

**示例**:
```go
type ClickButtonIntent struct{}

func (ClickButtonIntent) IntentType() string { return "ClickButton" }

buttonComp.NewBuilder("Click Me").
    OnPress(CustomButtonIntent{}).
    Build()
```

---

### OnToggle

**适用组件**: Checkbox

**用途**: 设置 checkbox 切换时发射的 Intent

**签名**:
```go
checkboxComp.NewBuilder().
    Label("Remember me").
    OnToggle(CustomToggleIntent{}).
    Build()
```

**示例**:
```go
checkboxComp.NewBuilder().
    Label("Remember me").
    ForField(intent.BindField("remember")).  // 使用 ForField 代替
    Checked(state.Remember == "true").
    Build()
```

---

## 已废弃 API

### UseState 系列

**状态**: `DEPRECATED`

**位置**: `ui/hooks.go`

**API**:
```go
// ❌ 已废弃：使用 UseState
username, setUsername := ui.UseStateString("")
email, setEmail := ui.UseStateString("")
count, setCount, getter := ui.UseStateInt(0)
checked, setChecked := ui.UseStateBool(false)
```

**替代方案**: Store + Reducer

```go
// ✅ 推荐：使用 Store + Reducer
type AppState struct {
    Username string
    Email    string
    Count    int
    Checked  string // "true"/"false"
}

// 定义 Reducer
var appReducer = reducer.NewBuilder[AppState]().
    On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
        s.Username = fieldChange.Value  // 直接更新，无类型断言
        return s
    }).RegisterToGlobal(appStore)

// 组件中使用
state := appStore.Get()
inputComp.NewBuilder().ForField(intent.BindField("username")).Value(state.Username).Build()
```

---

### RegisterIntent + GlobalState 手动注册

**状态**: `DEPRECATED`（在 Store + Reducer 架构下不推荐）

**位置**: `hooks.go` (GlobalState)

**API**:
```go
// ❌ 已废弃：手动注册 + 类型断言
ui.WithInit(func() {
    ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
        if fn, ok := ctx.GetState("setter"); ok {
            if setter, ok := fn.(func(string)); ok {  // 类型断言
                setter(i.Value)
            }
        }
        return intent.HandledResult()
    })
}, ...)
```

**替代方案**: BuildAndRegister 自动注册

```go
// ✅ 推荐：自动注册（无类型断言）
var appReducer = reducer.NewBuilder[AppState]().
    On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
        s.Field = i.(intent.FieldChangeIntent).Value
        return s
    })

// 在 main 或 init 中一次性注册
appReducer.RegisterToGlobal(appStore)
```

---

### SetChangeIntent (旧 API)

**状态**: `DEPRECATED`

**位置**: `ui/components/*/vnode.go` (组件级 setter)

**API**:
```go
inputComp.New().SetChangeIntent(intent.SetState("field", "value"))
```

**问题**: `SetStateIntent` 的值是静态的，不会随用户输入变化

**替代方案**: ForField + FieldChangeIntent

```go
// ✅ 推荐：使用 ForField
inputComp.NewBuilder().
    ForField(intent.BindField("field")).
    Value(state.Field).
    Build()
```

---

## 完整示例

### 示例：完整的 Store + Reducer 应用

```go
package main

import (
    "fmt"

    "github.com/wwsheng009/mint/runtime/intent"
    "github.com/wwsheng009/mint/runtime/reducer"
    "github.com/wwsheng009/mint/runtime/store"
    "github.com/wwsheng009/mint/ui"
    buttonComp "github.com/wwsheng009/mint/ui/components/button"
    inputComp "github.com/wwsheng009/mint/ui/components/input"
)

// State
type AppState struct {
    Count    int
    Username string
    Email    string
    Checked  string
}

// Intents
type IncrementIntent struct{}

func (IncrementIntent) IntentType() string { return "Increment" }
func (IncrementIntent) StayPressed() bool   { return true }

// Global Store
var appStore *store.Store[AppState]

func initStore() {
    appStore = store.NewStore(AppState{
        Count:    0,
        Username: "",
        Email:    "",
        Checked:  "false",
    })
}

// Reducer
var appReducer = reducer.NewBuilder[AppState]().
    On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
        s.Count++
        return s
    }).
    On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
        if fieldChange, ok := i.(intent.FieldChangeIntent); ok {
            switch fieldChange.Field {
            case "username":
                s.Username = fieldChange.Value
            case "email":
                s.Email = fieldChange.Value
            case "checked":
                s.Checked = fieldChange.Value
            }
        }
        return s
    })

// View
func App() ui.VNode {
    state := appStore.Get()
    
    return ui.VStack(
        ui.NewTextBuilder(fmt.Sprintf("Count: %d", state.Count)).Build(),
        buttonComp.NewBuilder("+").
            OnPress(IncrementIntent{}).
            Build(),
        
        ui.VStack(
            ui.NewTextBuilder("Username").Build(),
            inputComp.NewBuilder().
                ForField(intent.BindField("username")).
                Value(state.Username).
                Placeholder("Username").
                Build(),
        ),
        
        ui.VStack(
            ui.NewTextBuilder("Email").Build(),
            inputComp.NewBuilder().
                ForField(intent.BindField("email")).
                Value(state.Email).
                Placeholder("Email").
                Build(),
        ),
        
        ui.VStack(
            ui.NewTextBuilder("Checkbox").Build(),
            checkboxComp.NewBuilder().
                Label("Remember me").
                ForField(intent.BindField("checked")).
                Checked(state.Checked == "true").
                Build(),
        ),
    )
}

func main() {
    initStore()
    appReducer.RegisterToGlobal(appStore)
    
    err := ui.Run(App, ui.WithWidth(50), ui.WithHeight(20))
    if err != nil {
        panic(err)
    }
}
```

---

## API 快速索引

| 功能 | 已废弃 | 新 API | 类别 |
|------|-------|-------|------|
| 状态管理 | `UseState...` | `Store[T]` | 核心 |
| | `GlobalState["setter"]` | `Store.Get()` | 核心 |
| 状态更新 | `setter(value)` | `Store.Set(newState)` | 核心 |
| 状态订阅 | ❌ 无 | `Store.Subscribe(callback)` | 核心 |
| 字段绑定 | `SetChangeIntent(SetState)` | `ForField(FieldBinding)` | 组件 |
| 按钮意图 | `Click` | 自定义 Intent | Intent |
| 字段变更 | `FieldChangeIntent` | `FieldChangeIntent` | Intent |

---

## 版本兼容性

### Mint UI v0.9 使用 UseState 的迁移路径

| API | v0.9 | v0.10+ |
|-----|------|--------|
| `UseStateString` | ✅ 支持 | ⚠️ Deprecated |
| `UseStateInt` | ✅ 支持 | ⚠️ Deprecated |
| `UseStateBool` | ✅ 支持 | ⚠️ Deprecated |
| `GlobalState["setter"]` | ✅ 支持 | ⚠️ 不推荐 |
| `ui.RegisterIntent` | ✅ 支持 | ⚠️ 使用 BuildAndRegister |

### Store + Reducer API

| API | v0.9 | v0.10+ |
|-----|------|--------|
| `store.NewStore` | ✅ 支持 | ✅ 推荐 |
| `store.Store.Get()` | ✅ 支持 | ✅ 推荐 |
| `reducer.NewBuilder[State]` | ✅ 支持 | ✅ 推荐 |
| `BuildAndRegister` | ✅ 支持 | ✅ 推荐 |
| `ForField(Intent.BindField)` | ✅ 支持 | ✅ 推荐 |

---

## 参见指南

- [迁移指南](MIGRATION_GUIDE.md) - 从 UseState 迁移到 Store + Reducer
- [开发指南](DEVELOPMENT_GUIDE.md) - Store + Reducer 开发最佳实践
- [示例代码](../../../examples/store_reducer_demo/main.go) - 完整示例参考
