# Mint UI Store + Reducer 架构

**版本**: v0.10
**状态**: 稳定可用 ✅

> 当前 API 提示：本文档组中有一部分页面是历史迁移和问题复盘资料。当前 Button 不再接受闭包回调，推荐 `ui.NewButtonBuilder("...").OnPress(SomeIntent{}).Build()`；当前字段组件推荐 `ui.NewInputBuilder().ForField(intent.BindField("field"))` 或 `intent.ForField(StateKey)`。旧 `ui.Button(label, func)`、`OnClick(func)`、`OnPress(func)` 只应作为迁移对照阅读。

---

## 概述

Store + Reducer 架构是 Mint UI 的推荐状态管理模式，实现了**单向数据流**和**单一真相源**模式：

```
Intent → Dispatcher → Reducer → Store → View
```

### 核心原则

| 原则 | 说明 |
|------|------|
| **单一真相源** | 所有状态存储在一个 Store 中 |
| **状态只读** | 状态只能通过 Reducer 修改 |
| **无副作用** | Reducer 是纯函数，无副作用 |

---

## 快速开始

### 1. 定义 State

```go
type AppState struct {
    Count    int
    Username string
    Email    string
}

var appStore *store.Store[AppState]

func initStore() {
    appStore = store.NewStore(AppState{
        Count:    0,
        Username: "",
        Email:    "",
    })
}
```

### 2. 定义 Reducer

**方案 1: 基础方式（使用 On）**

```go
import "github.com/wwsheng009/mint/runtime/reducer"
import "github.com/wwsheng009/mint/runtime/intent"
import "github.com/wwsheng009/mint/runtime/store"

var appReducer = reducer.NewBuilder[AppState]().
    On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
        s.Count++
        return s
    }).
    On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
        s.Username = i.(intent.FieldChangeIntent).Value
        return s
    })
```

**方案 2: 优化方式（使用 FieldMap，推荐）**

```go
import "strconv"
import "github.com/wwsheng009/mint/runtime/reducer"
import "github.com/wwsheng009/mint/runtime/intent"
import "github.com/wwsheng009/mint/runtime/store"

// 使用 FieldMap 消除 switch-case 硬编码
var appReducer = reducer.BindField(reducer.NewBuilder[AppState]()).
    BindFieldMap(map[string]func(AppState, string) AppState{
        // 所有字段集中定义，单一处理器
        "Username": func(s AppState, val string) AppState {
            s.Username = val
            return s
        },
        "Email": func(s AppState, val string) AppState {
            s.Email = val
            return s
        },
        "Age": func(s AppState, val string) AppState {
            if v, err := strconv.Atoi(val); err == nil {
                s.Age = v
            }
            return s
        },
        "Agreed": func(s AppState, val string) AppState {
            s.Agreed = val == "true"
            return s
        },
    }).
    GetBuilder().
    On(SubmitIntent{}, func(s AppState, i intent.Intent) AppState {
        // 提交逻辑
        return s
    })
```

**详细优化指南**: [FIELD_BINDING_OPTIMIZATION.md](../optimization/FIELD_BINDING_OPTIMIZATION.md)

### 3. 注册 Handlers

```go
// 在 main 或 WithInit 中注册
appReducer.RegisterToGlobal(appStore)
```

### 4. 组件渲染

```go
func App() ui.VNode {
    state := appStore.Get()
    
    return ui.VStack(
        ui.NewButtonBuilder("+").
            OnPress(IncrementIntent{}).
            Build(),
        ui.NewTextBuilder(fmt.Sprintf("Count: %d", state.Count)).Build(),
    )
}
```

---

## 核心组件详情

### Store[T] - 状态容器

**位置**: `runtime/store/store.go`

| API | 说明 |
|------|------|
| `NewStore[T](initial T)` | 初始化 Store |
| `Get() T` | 读取状态 |
| `Set(next T)` | 更新状态 |
| `Subscribe(callback func(T)) func()` | 订阅变化 |
| `Update(fn func(T) T)` | 函数式更新 |
| `ListenerCount() int` | 查询订阅者数量 |

**高级功能**:
- `Selector[T,R]` - 选择器支持
- `Computed[T,R]` - 计算值，自动缓存
- `Computed[T,R].Get()` - 获取缓存值

### Reducer[T] - 状态函数器

**位置**: `runtime/reducer/reducer.go`

| API | 说明 |
|------|------|
| `New[T](fn func(T, Intent) T)` | 创建 Reducer |
| `NewBuilder[T]()` | 创建 Builder |
| `On(Intent, Handler)` | 注册 handler |
| `OnTyped(Intent, Handler)` | 类型安全注册 |
| `OnField(string, Handler)` | 字段绑定（已废弃） |

**FieldBinding API**（推荐）:
| API | 说明 |
|------|------|
| `BindField(Builder[T])` | 创建 FieldBinder |
| `BindFieldMap(Map[T])` | 使用映射表绑定字段 |
| `BindStringField(name, Handler)` | 绑定字符串字段 |
| `BindIntField(name, Handler)` | 绑定整型字段（自动转换）|
| `BindBoolField(name, Handler)` | 绑定布尔字段（自动转换）|
| `UpdateStringFieldIfChanged(state, current, next, update)` | 字符串字段未变化时 no-op，变化时执行更新逻辑，适合筛选 scope 变更后重置分页或 selection |

**详细文档**: [FIELD_BINDING_OPTIMIZATION.md](../optimization/FIELD_BINDING_OPTIMIZATION.md)

**自动注册**:
```go
// 方式 1: 自动注册所有 handlers
appReducer.RegisterToGlobal(appStore)  // 每个状态类型的 handler 自动注册

// 方式 2: 手动注册到 registry
appReducer.BuildAndRegister(registry, appStore)
```

### AppRuntime[T] - 运行时（可选）

**位置**: `runtime/statemachine/runtime.go`

| API | 说明 |
|------|------|
| `NewAppRuntime(initial, View, Reducer)` | 创建运行时 |
| `GetState() T` | 获取状态 |
| `Dispatch(Intent)` | 分发 intent |
| `Subscribe(callback)` | 订阅状态变化 |
| `View()` | 获取视图 |
| `JumpTo(index)` | 时间旅行跳转 |
| `Undo()` | 撤销到上一个状态 |
| `History()` | 完整历史记录 |
| `WithMaxHistory(n)` | 配置历史大小 |

---

## 使用指南

### 基础示例

详细的完整示例请参考：

- **示例**: `examples/store_reducer_demo/main.go`
- **迁移指南**: `MIGRATION_GUIDE.md` - 如何将 UseState 迁移到 Store + Reducer
- **开发指南**: `DEVELOPMENT_GUIDE.md` - 如何设计和构建应用
- **API 参考**: `API_REFERENCE.md` - 完整 API 文档

---

## 当前状态

- ✅ **核心组件完整度**: Store[T], Reducer[T], FieldBinding API 完整
- ✅ **示例代码有效**: 9 个示例已迁移并运行
  - focus_switching_demo, validation_demo, mvp_form_demo, mvp_components_demo
  - typesafe_form_demo, ant_design_demo, checkbox demo, absolute demo, counter demo
- ✅ **代码简化**: 平均减少 25% 代码量
- ✅ **文档完善度**: 迁移指南、开发指南、API 参考、字段绑定优化指南
- ✅ **FieldBinding 优化**: 使用 FieldMap 消除 switch-case 硬编码
- ⚠️ **需要更新**: 文档中提到但未实现的 RunApp 入口

---

## 相关文档

| 文档 | 说明 |
|------|------|
| **API** | `docs/ui/store/api/API_REFERENCE.md` - API 参考 |
| **迁移** | `docs/ui/store/guides/MIGRATION_GUIDE.md` - 迁移指南 |
| **开发** | `docs/ui/store/guides/DEVELOPMENT_GUIDE.md` - 开发指南 |
| **状态** | `docsArchive/status/CURRENT_STATUS.md` - 状态评估 |
| **优化** | `docs/ui/store/optimization/FIELD_BINDING_OPTIMIZATION.md` - 字段绑定优化 |
| **进度** | `docsArchive/status/MIGRATION_PROGRESS.md` - 迁移进度 |

---

## 版本兼容性

| Mint UI v0.9 | Mint UI v0.10+ | 变化点 |
|---------------|----------------------|----------|
| UseState | ✅ 支持 | ⚠️ Deprecated |
| GlobalState | ✅ 支持 | ⚠️ 不推荐 |
| BuildAndRegister | ✅ 支持 | ✅ 推荐 |
| ForField API | ✅ 支持 | ✅ 推荐 |
| ForField API | ✅ 支持 | ✅ 推荐 |

---

## 示例代码

### 基础应用示例

```go
package main

import (
    "github.com/wwsheng009/mint/runtime/intent"
    "github.com/wwsheng009/mint/runtime/reducer"
    "github.com/wwsheng009/mint/runtime/store"
    "github.com/wwsheng009/mint/ui"
    buttonComp "github.com/wwsheng009/mint/ui/components/button"
)

// State
type AppState struct {
    Count int
}

// Intent
type IncrementIntent struct{}
func (IncrementIntent) IntentType() string { return "Increment" }

// Store
var appStore *store.Store[AppState]

// Reducer
var appReducer = reducer.NewBuilder[AppState]().
    On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
        s.Count++
        return s
    })

// View
func App() ui.VNode {
    state := appStore.Get()
    fmt.Printf("State in main: %+v\n", state)
    return ui.VStack(
        ui.NewTextBuilder(fmt.Sprintf("Count: %d", state.Count)).Bold(true).Build(),
        buttonComp.NewButton("Increment").OnPress(IncrementIntent{}).Build(),
    )
}

func main() {
    appStore = store.NewStore(AppState{Count: 0})
    appReducer.RegisterToGlobal(appStore)
    
    err := ui.Run(App)
    if err != nil {
        panic(err)
    }
}
```

---

## 运行应用

```go
// 在 main.go
ui.Run(App, ui.WithTitle("App"))
```

---

## 调试提示

- 使用 `appStore.Get()` 查看当前状态
- 使用 `appStore.Set(newState)` 更新状态（通常由 Reducer 调用）
- 使用 `appStore.Subscribe()` 监听状态变化
