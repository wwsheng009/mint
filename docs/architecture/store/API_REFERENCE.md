# Store + Reducer API 参考

**版本**: v0.11
**最后更新**: 2026-03-05

---

## 目录

- [Store API](#store-api)
- [Reducer API](#reducer-api)
- [FieldBinding API](#fieldbinding-api)
- [Intent API](#intent-api)
- [AppRuntime API](#appruntime-api)

---

## Store API

**位置**: `github.com/wwsheng009/mint/runtime/store`

### 创建 Store

```go
func NewStore[T any](initial T) *Store[T]
```

**参数**:
- `initial`: 初始状态

**返回值**: `*Store[T]` - Store 实例

**示例**:
```go
appStore := store.NewStore(AppState{
    Count: 0,
})
```

---

### 读取状态

```go
func (s *Store[T]) Get() T
```

**返回值**: `T` - 当前状态

**示例**:
```go
state := appStore.Get()
fmt.Printf("Count: %d\n", state.Count)
```

---

### 更新状态

```go
func (s *Store[T]) Set(next T)
func (s *Store[T]) Update(fn func(T) T)
```

**参数**:
- `Set(next)`: 直接设置新状态
- `Update(fn)`: 基于当前状态计算新状态

**示例**:
```go
// 方式 1: 直接设置
appStore.Set(AppState{Count: 10})

// 方式 2: 函数式更新
appStore.Update(func(state AppState) AppState {
    state.Count++
    return state
})
```

---

### 订阅状态变化

```go
func (s *Store[T]) Subscribe(callback func(T)) func()
```

**参数**:
- `callback`: 状态变化回调

**返回值**: `func()` - 取消订阅函数

**示例**:
```go
unsubscribe := appStore.Subscribe(func(state AppState) {
    fmt.Printf("State changed: %+v\n", state)
})

// 取消订阅
unsubscribe()
```

---

### 选择器

```go
func (s *Store[T]) Select[R any](selector func(T) R) *Computed[R]
```

**参数**:
- `selector`: 选择器函数，从 `T` 中提取 `R`

**返回值**: `*Computed[R]` - 计算值

**示例**:
```go
totalPrice := appStore.Select(func(state AppState) float64 {
    total := 0.0
    for _, item := range state.Items {
        total += item.Price
    }
    return total
})

price := totalPrice.Get()
```

---

### 计算值

```go
func (s *Store[T]) Compute[R any](compute func(T) R) *Computed[R]
```

**参数**:
- `compute`: 计算函数，自动缓存

**返回值**: `*Computed[R]` - 计算值

**示例**:
```go
itemCount := appStore.Compute(func(state AppState) int {
    return len(state.Items)
})

// 自动缓存，重复调用不重复计算
count1 := itemCount.Get()
count2 := itemCount.Get()  // 使用缓存
```

---

### Computed API

```go
type Computed[R] struct {
    Get() R
    Invalidate()
}
```

**方法**:
- `Get() R`: 获取值
- `Invalidate()`: 清除缓存，下次调用重新计算

**示例**:
```go
itemCount := appStore.Compute(func(state AppState) int {
    return len(state.Items)
})

itemCount.Invalidate()  // 清除缓存
value := itemCount.Get()  // 重新计算
```

---

### 订阅者数量

```go
func (s *Store[T]) ListenerCount() int
```

**返回值**: `int` - 当前订阅者数量

**示例**:
```go
count := appStore.ListenerCount()
fmt.Printf("Subscribers: %d\n", count)
```

---

## Reducer API

**位置**: `github.com/wwsheng009/mint/runtime/reducer/generic_reducer.go`

### 创建 Reducer

```go
func New[T](fn func(T, intent.Intent) T) *Reducer[T]
func NewBuilder[T]() *Builder[T]
```

**参数**:
- `fn`: Reducer 函数
- `NewBuilder()`: 使用 Builder 模式

**返回值**: `*Reducer[T]` 或 `*Builder[T]`

**示例**:
```go
// 方式 1: 直接创建
appReducer := reducer.New(func(state AppState, i intent.Intent) AppState {
    // 处理逻辑
    return state
})

// 方式 2: Builder 模式（推荐）
appReducer := reducer.NewBuilder[AppState]()
```

---

### 注册 Handler

```go
func (b *Builder[T]) On(intentType Intent, handler Handler[T]) *Builder[T]
```

**参数**:
- `intentType`: Intent 类型
- `handler`: 处理函数 `func(T, intent.Intent) T`

**返回值**: `*Builder[T]` - 支持链式调用

**示例**:
```go
appReducer := reducer.NewBuilder[AppState]().
    On(IncrementIntent{}, func(state AppState, i intent.Intent) AppState {
        state.Count++
        return state
    }).
    On(DecrementIntent{}, func(state AppState, i intent.Intent) AppState {
        state.Count--
        return state
    })
```

---

### 字段绑定（已废弃）

```go
func (b *Builder[T]) OnField(fieldName string, handler FieldHandler[T]) *Builder[T]
```

**状态**: 已废弃，推荐使用 FieldBinding API

---

### 构建 Reducer

```go
func (b *Builder[T]) Build() *Reducer[T]
```

**返回值**: `*Reducer[T]` - Reducer 实例

**示例**:
```go
appReducer := reducer.NewBuilder[AppState]().
    On(IncrementIntent{}, func(state AppState, i intent.Intent) AppState {
        state.Count++
        return state
    }).
    Build()
```

---

### 注册到 Global Registry

```go
func (r *Reducer[T]) RegisterToGlobal(stores ...*store.Store[T])
```

**参数**:
- `stores`: 一个或多个 Store

**示例**:
```go
appReducer.RegisterToGlobal(appStore)
```

---

### BuildAndRegister

```go
func (b *Builder[T]) BuildAndRegister(registry *intent.GlobalIntentRegistry, stores ...*store.Store[T]) *Reducer[T]
```

**参数**:
- `registry`: 全局 Intent Registry
- `stores`: 一个或多个 Store

**返回值**: `*Reducer[T]` - Reducer 实例

**示例**:
```go
appReducer := reducer.NewBuilder[AppState]().
    On(IncrementIntent{}, func(...) ...).
    BuildAndRegister(intent.DefaultRegistry, appStore)
```

---

## FieldBinding API

**位置**: `github.com/wwsheng009/mint/runtime/reducer/field_mapping.go`

### 创建 FieldBinder

```go
func BindField[T any](builder *Builder[T]) *FieldBinder[T]
```

**参数**:
- `builder`: Builder 实例

**返回值**: `*FieldBinder[T]` - FieldBinder 实例

**示例**:
```go
fieldBinder := reducer.BindField(reducer.NewBuilder[AppState]())
```

---

### BindFieldMap

```go
func (fb *FieldBinder[T]) BindFieldMap(fieldMap FieldMap[T]) *FieldBinder[T]
```

**类型定义**:
```go
type FieldMap[T any] = map[string]func(T, string) T
```

**参数**:
- `fieldMap`: 字段映射表（字段名 → 更新函数）

**返回值**: `*FieldBinder[T]` - 支持链式调用

**示例**:
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
        "age": func(s AppState, val string) AppState {
            if v, err := strconv.Atoi(val); err == nil {
                s.Age = v
            }
            return s
        },
    })
```

---

### 类型化绑定

#### BindStringField

```go
func (fb *FieldBinder[T]) BindStringField(fieldName string, setter func(*T, string)) *FieldBinder[T]
```

**参数**:
- `fieldName`: 字段名
- `setter`: 字段设置函数

**示例**:
```go
var appReducer = reducer.BindField(reducer.NewBuilder[AppState]()).
    BindStringField("username", func(s *AppState, val string) {
        s.Username = val
    }).
    BindStringField("email", func(s *AppState, val string) {
        s.Email = val
    })
```

---

#### BindIntField

```go
func (fb *FieldBinder[T]) BindIntField(fieldName string, setter func(*T, int)) *FieldBinder[T]
```

**参数**:
- `fieldName`: 字段名
- `setter`: 类型安全的字段设置函数

**示例**:
```go
var appReducer = reducer.BindField(reducer.NewBuilder[AppState]()).
    BindIntField("age", func(s *AppState, val int) {
        s.Age = val
    }).
    BindIntField("count", func(s *AppState, val int) {
        s.Count = val
    })
```

---

#### BindBoolField

```go
func (fb *FieldBinder[T]) BindBoolField(fieldName string, setter func(*T, bool)) *FieldBinder[T]
```

**参数**:
- `fieldName`: 字段名
- `setter`: 类型安全的字段设置函数

**示例**:
```go
var appReducer = reducer.BindField(reducer.NewBuilder[AppState]()).
    BindBoolField("agreed", func(s *AppState, val bool) {
        s.Agreed = val
    }).
    BindBoolField("enabled", func(s *AppState, val bool) {
        s.Enabled = val
    })
```

---

### GetBuilder

```go
func (fb *FieldBinder[T]) GetBuilder() *Builder[T]
```

**返回值**: `*Builder[T]` - Builder 实例

**示例**:
```go
var appReducer = reducer.BindField(reducer.NewBuilder[AppState]()).
    BindFieldMap(map[string]func(AppState, string) AppState{
        "username": func(s AppState, val string) AppState {
            s.Username = val
            return s
        },
    }).
    GetBuilder().
    On(SubmitIntent{}, func(s AppState, i intent.Intent) AppState {
        return s
    })
```

---

### Build

```go
func (fb *FieldBinder[T]) Build() *Reducer[T]
```

**返回值**: `*Reducer[T]` - Reducer 实例

**示例**:
```go
appReducer := reducer.BindField(reducer.NewBuilder[AppState]()).
    BindStringField("username", func(s *AppState, val string) {
        s.Username = val
    }).
    Build()
```

---

### 组合示例

```go
// 方式 1: FieldMap + 自定义 Intent
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
        // 提交逻辑
        return s
    })

// 方式 2: 类型化绑定
var appReducer2 = reducer.BindField(reducer.NewBuilder[AppState]()).
    BindStringField("username", func(s *AppState, val string) {
        s.Username = val
    }).
    BindIntField("age", func(s *AppState, val int) {
        s.Age = val
    }).
    BindBoolField("agreed", func(s *AppState, val bool) {
        s.Agreed = val
    }).
    GetBuilder().
    On(SubmitIntent{}, func(s AppState, i intent.Intent) AppState {
        return s
    })
```

**详细优化指南**: [FIELD_BINDING_OPTIMIZATION.md](./FIELD_BINDING_OPTIMIZATION.md)

---

## Intent API

**位置**: `github.com/wwsheng009/mint/runtime/intent`

### Intent 接口

```go
type Intent interface {
    IntentType() string
}
```

### FieldChangeIntent

```go
type FieldChangeIntent struct {
    Field string
    Value string
}

func FieldChangeIntent.IntentType() string {
    return "FieldChange"
}
```

**属性**:
- `Field`: 字段名
- `Value`: 字段值

---

### ForField API

```go
func BindField(field string) FieldChangeIntent
```

**参数**:
- `field`: 字段名

**返回值**: `FieldChangeIntent`

**使用**:
```go
ui.NewInputBuilder().
    ForField(intent.BindField("username")).
    Value(state.Username).
    Build()
```

---

### 自定义 Intent

```go
type IncrementIntent struct {
    Amount int
}

func (IncrementIntent) IntentType() string {
    return "Increment"
}

type SubmitIntent struct{}

func (SubmitIntent) IntentType() string {
    return "Submit"
}
```

---

## AppRuntime API

**位置**: `github.com/wwsheng009/mint/runtime/statemachine/runtime.go`

### 创建 AppRuntime

```go
func NewAppRuntime[T any](initial T, view View[T], reducer *Reducer[T]) *AppRuntime[T]
```

**参数**:
- `initial`: 初始状态
- `view`: 视图函数
- `reducer`: Reducer 实例

**返回值**: `*AppRuntime[T]` - AppRuntime 实例

**示例**:
```go
runtime := NewAppRuntime[AppState](
    AppState{Count: 0},
    App,
    appReducer,
)
```

---

### GetState

```go
func (r *AppRuntime[T]) GetState() T
```

**返回值**: `T` - 当前状态

---

### Dispatch

```go
func (r *AppRuntime[T]) Dispatch(i Intent)
```

**参数**:
- `i`: Intent

---

### Subscribe

```go
func (r *AppRuntime[T]) Subscribe(callback func(T)) func()
```

**参数**:
- `callback`: 状态变化回调

**返回值**: `func()` - 取消订阅函数

---

### 时间旅行

```go
func (r *AppRuntime[T]) JumpTo(index int)
func (r *AppRuntime[T]) Undo()
func (r *AppRuntime[T]) History() []T
func (r *AppRuntime[T]) WithMaxHistory(n int) *AppRuntime[T]
```

**方法**:
- `JumpTo(index)`: 跳转到历史状态
- `Undo()`: 撤销到上一个状态
- `History()`: 完整历史记录
- `WithMaxHistory(n)`: 配置历史大小

---

### View

```go
func (r *AppRuntime[T]) View() ui.VNode
```

**返回值**: `ui.VNode` - 视图节点

---

### RunApp

```go
func (r *AppRuntime[T]) RunApp(opts ...ui.AppOption) error
```

**参数**:
- `opts`: 应用选项

**返回值**: `error` - 错误信息

---

## 使用示例

### 完整示例：字段绑定优化

```go
package main

import (
    "strconv"
    "github.com/wwsheng009/mint/runtime/intent"
    "github.com/wwsheng009/mint/runtime/reducer"
    "github.com/wwsheng009/mint/runtime/store"
    "github.com/wwsheng009/mint/ui"
)

// State
type AppState struct {
    Username string
    Email    string
    Age      int
    Agreed   bool
}

// Intent
type SubmitIntent struct{}

// Store
var appStore = store.NewStore(AppState{})

// Reducer: 使用 FieldMap 方式
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
        "age": func(s AppState, val string) AppState {
            if v, err := strconv.Atoi(val); err == nil {
                s.Age = v
            }
            return s
        },
        "agreed": func(s AppState, val string) AppState {
            s.Agreed = val == "true"
            return s
        },
    }).
    GetBuilder().
    On(SubmitIntent{}, func(s AppState, i intent.Intent) AppState {
        fmt.Printf("Submit: %+v\n", s)
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
        ui.NewInputBuilder().
            ForField(intent.BindField("age")).
            Value(strconv.Itoa(state.Age)).
            Placeholder("Age").
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

### 选择器和计算值示例

```go
func App() ui.VNode {
    // 计算 - 自动缓存
    itemCount := appStore.Compute(func(state AppState) int {
        return len(state.Items)
    })

    // 选择器 - 派生状态
    totalPrice := appStore.Select(func(state AppState) float64 {
        total := 0.0
        for _, item := range state.Items {
            total += item.Price
        }
        return total
    })

    return ui.VStack(
        ui.NewTextBuilder(fmt.Sprintf("Items: %d", itemCount.Get())).Build(),
        ui.NewTextBuilder(fmt.Sprintf("Total: %.2f", totalPrice.Get())).Build(),
    )
}
```

---

## 总结

### 主要 API

| 组件 | 主要 API |
|------|---------|
| Store | `NewStore`, `Get`, `Set`, `Update`, `Subscribe`, `Compute`, `Select` |
| Reducer | `NewBuilder`, `On`, `Build`, `RegisterToGlobal` |
| FieldBinding | `BindField`, `BindFieldMap`, `BindStringField`, `BindIntField`, `BindBoolField` |
| Intent | `FieldChangeIntent`, `BindField` |
| AppRuntime | `NewAppRuntime`, `Dispatch`, `JumpTo`, `Undo`, `History` |

### 快速参考

```go
// Store
appStore = store.NewStore(AppState{})
state = appStore.Get()
appStore.Set(newState)
appStore.Update(func(s AppState) AppState { ... })

// Reducer
appReducer = reducer.NewBuilder[AppState]().
    On(Intent{}, func(...) ...).
    On(intent.FieldChangeIntent{}, func(...) ...).
    Build()

// FieldBinding（推荐）
appReducer = reducer.BindField(reducer.NewBuilder[AppState]()).
    BindFieldMap(map[string]func(AppState, string) AppState{
        "field": func(s, val) AppState { ... },
    }).
    GetBuilder().
    Build()

// 注册
appReducer.RegisterToGlobal(appStore)

// 组件
ui.NewInputBuilder().
    ForField(intent.BindField("field")).
    Value(state.Field).
    Build()
```

---

**文档创建**: 2026-03-05
**状态**: 完成 ✅
