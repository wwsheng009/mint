# Store + Reducer 开发指南

**版本**: v1.0  
**创建时间**: 2026-03-04  
**适用版本**: Mint UI v0.10+

---

## 目录

- [一、架构概览](#一架构概览)
- [二、快速开始](#二快速开始)
- [三、State 设计](#三state-设计)
- [四、Reducer 设计](#四reducer-设计)
- [五、组件集成](#五组件集成)
- [六、数据流](#六数据流)
- [七、最佳实践](#七最佳实践)
- [八、常见模式](#八常见模式)
- [九、调试](#九调试)
- [十、性能优化](#十性能优化)

---

## 一、架构概览

### 1.1 核心原则

| 原则 | 说明 |
|------|------|
| **Single Source of Truth** | 所有状态存储在一个 `Store[T]` |
| **State Immutability** | 状态不可变，修改总是返回新的 State |
| **Pure Reducers** | Reducer 是纯函数，无副作用 |
| **Unidirectional Data Flow** | Intent → Reducer → Store → View |

### 1.2 架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Mint UI 架构                                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────┐   Intent   ┌─────────────┐   Reducer    ┌──────────┐ │
│  │   Component │ ────┬──────►  │ Dispatcher │ ──────┬──────► │  State   │ │
│  └─────────────┘   │         └─────────────┘       │         └──────────┘ │
│                    │                                    │            ▲                │
│                    │                                    │            │                │
│                    │           ┌─────────────┐            │            │                │
│                    └───────────│    Store   │────────────│            │                │
│                                │   Store<T>│             │            │                │
│                                └─────────────┘             │            │                │
│                                         │                       │            ▼                │
│                                         │                       │    ┌──────────┐                 │
│                                         └───────────────────────────────┤   View    │                 │
│                                                                 │  (VNode)  │                 │
│                                                                 └──────────┘                 │
│                                                                             │
└─────────────────────────────────────────────────────────────────────┘
```

### 1.3 核心组件

| 组件 | 职责 | 位置 |
|------|------|------|
| **Store[T]** | 单一状态源，管理应用状态 | `runtime/store/store.go` |
| **Reducer[T]** | 纯函数，转换状态 | `runtime/reducer/reducer.go` |
| **ViewFunction[T]** | 纯函数，渲染状态到 VNode | 用户代码 |
| **Dispatcher** | 分发 Intent 到 Reducer | `runtime/intent/` |

---

## 二、快速开始

### 2.1 创建你的第一个 Store + Reducer 应用

```go
package main

import (
    "github.com/wwsheng009/mint/runtime/intent"
    "github.com/wwshengeng009/mint/runtime/reducer"
    "github.com/wwsheng009/mint/runtime/store"
    "github.com/wwsheng009/mint/ui"
    buttonComp "github.com/wwsheng009/mint/ui/components/button"
)

// 1. 定义 State
type AppState struct {
    Count int
}

// 2. 定义 Intent
type IncrementIntent struct{}

func (IncrementIntent) IntentType() string { return "Increment" }
func (IncrementIntent) StayPressed() bool  { return true }

// 3. 创建全局 Store
var appStore *store.Store[AppState]

func initStore() {
    appStore = store.NewStore(AppState{Count: 0})
}

// 4. 定义 Reducer
var appReducer = reducer.NewBuilder[AppState]().
    On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
        s.Count++
        return s
    })

// 5. 注册 Handlers
func registerHandlers() {
    appReducer.RegisterToGlobal(appStore)
}

// 6. 视图函数
func App() ui.VNode {
    state := appStore.Get()  // 从 Store 读取状态
    
    return ui.VStack(
        ui.NewTextBuilder(fmt.Sprintf("Count: %d", state.Count)).Build(),
        buttonComp.NewBuilder("+").
            OnPress(IncrementIntent{}).
            Build(),
    )
}

// 7. 主函数
func main() {
    initStore()
    
    err := ui.Run(App,
        ui.WithWidth(40),
        ui.WithHeight(10),
        ui.WithInit(registerHandlers),
    )
    if err != nil {
        panic(err)
    }
}
```

---

## 三、State 设计

### 3.1 原则

| 原则 | 说明 | 示例 |
|------|------|------|
| **扁平优先** | 尽量保持 State 扁平 | ✅ `Count int` |
| | 避免深层嵌套 | ❌ `Form.Form.Data.Field` |
| **类型安全** | 使用具体类型 | ✅ `Count int` |
| | 避免使用 interface{} | ❌ `Value interface{}` |
| **简单类型** | 优先使用基础类型 | ✅ `Username string` |
| | 考虑序列化成本 | ⚠️ `Data struct{...}` |
| **只读视图** | State 绝不修改 | 使用 `appStore.Get()` |

### 3.2 好的 State 设计示例

```go
// ✅ 好的设计：扁平、类型安全
type AppState struct {
    // 计数器
    Count int

    // 表单字段
    Username string
    Email    string

    // Checkbox (使用 string 统一存储)
    AgreeTerms string

    // UI 状态
    ActiveTab int
    IsLoading bool

    // 列表
    Items []string
}

// ❌ 避免的设计：深层嵌套
type BadAppState struct {
    Data struct {  // 嵌套结构
        Form struct {
            Username string
        }
        UI struct {
            Count struct {
                Value int
            }
        }
    }
}
```

### 3.3 State 初始化

```go
// 默认值
err := ui.Run(App, ui.WithInit(initStore))

func initStore() {
    appStore.Set(AppState{
        Count: 0,
        Username: "",
        Email: "",
        ActiveTab: 0,
        IsLoading: false,
        Items: []string{},
    })
}
```

---

## 四、Reducer 设计

### 4.1 Reducer 原则

| 原则 | 说明 | 示例 |
|------|------|------|
| **Pure Function** | 无副作用，不修改输入 | `s.Count++` ✅ |
| **Immutable** | 返回新的 State，不修改输入 | `NewState(s)` ✅ |
| **Default Case** | 处理未知的 Intent | `return s` |
| **小 Function** | 每个 Intent 单独处理 | `Case Increment` ✅ |

### 4.2 基础 Reducer

```go
type IncrementIntent struct{}

func (IncrementIntent) IntentType() string { return "Increment" }
func (IncrementIntent) StayPressed() bool  { return true }

var appReducer = reducer.NewBuilder[AppState]().
    On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
        s.Count++
        return s
    })
```

### 4.3 FieldChangeIntent Reducer

```go
var appReducer = reducer.NewBuilder[AppState]().
    On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
        if fieldChange, ok := i.(intent.FieldChangeIntent); ok {
            switch fieldChange.Field {
            case "username":
                s.Username = fieldChange.Value
            case "email":
                s.Email = fieldChange.Value
            case "agree":  // Checkbox 存储为 "true"/"false"
                s.AgreeTerms = fieldChange.Value
            }
        }
        return s
    })
```

### 4.4 复杂逻辑 Reducer

```go
var appReducer = reducer.NewBuilder[AppState]().
    On(SubmitFormIntent{}, func(s AppState, i intent.Intent) AppState {
        // 验证所有字段
        isValid := true

        if len(s.Username) < 3 {
            s.UsernameErr = "用户名至少3个字符"
            isValid = false
        }

        if !strings.Contains(s.Email, "@") {
            s.EmailErr = "请输入有效邮箱"
            isValid = false
        }

        if isValid {
            // 表单验证通过
            s.Submitted = true
        }

        return s
    })
```

---

## 五、组件集成

### 5.1 Input 组件

```go
// Input 组件自动发射 FieldChangeIntent
inputComp.NewBuilder().
    ForField(intent.BindField("username")).  // 绑定到 State 字段
    Value(state.Username).                      // 显示值
    Placeholder("Username").
    Build()
```

### 5.2 Checkbox 组件

```go
// Checkbox 存储布尔值为 "true"/"false" 字符串
checkboxComp.NewBuilder().
    Label("Remember me").
    ForField(intent.BindField("agree")).
    Checked(state.Agree == "true").  // 字符串转布尔显示
    Build()
```

### 5.3 Button 组件

```go
// 使用自定义 Intent
buttonComp.NewBuilder("Submit").
    OnPress(SubmitFormIntent{}).
    Build()
```

### 5.4 Select 组件

```go
// Select 组件也可以使用 ForField
selectComp.NewBuilder().
    ForField(intent.BindField("country")).
    Value(state.Country).
    Options([]selectComp.Option{...}).
    Build()
```

---

## 六、数据流

### 6.1 用户输入数据流

```
用户输入 'a'
    ↓
Input Instance 缓冲: inst.value = "a"
    ↓
ForField 自动发射: FieldChangeIntent{Field: "username", Value: "a"}
    ↓
Dispatcher → Handler (BuildAndRegister 自动注册)
    ↓
Reducer 处理: s.Username = "a"
    ↓
Store 更新: store.Set(newState)
    ↓
组件重新渲染: state := appStore.Get() → 输入框显示 "a"
```

### 6.2 按钮点击数据流

```
用户按 ENTER
    ↓
Button Instance 发射: OnPress 意图
    ↓
Dispatcher → Handler
    ↓
Reducer 处理: s.ClickCount++ → newState
    ↓
Store 更新: store.Set(newState)
    ↓
组件重新渲染: 显示新的 ClickCount
```

---

## 七、最佳实践

### 7.1 State 设计最佳实践

| ✅ 推荐 | ❌ 避免 |
|--------|--------|
| **扁平结构**: `Count int, Username string` | **深层嵌套**: `Data.Form.Input.Value` |
| **具体类型**: `Count int` | **避免 interface{}** |
| **字符串表示**: `Checked string` | **JSON 序列化** |
| **字段分组**: 前缀分组 | **无组织** |

### 7.2 Reducer 设计最佳实践

| ✅ 推荐 | ❌ 避免 |
|--------|--------|
| **Switch Intent Type**: `switch i.IntentType()` | **类型 switch** |
| **类型断言保护**: `if typed, ok := i.(Type); ok` | **直接断言** |
| **返回新 State**: `return s` 或 `return newState` | **修改输入 State** |
| **Default case**: 未处理 Intent 返回原 State | | **Panic** |

### 7.3 组件集成最佳实践

| ✅ 推荐 | ❌ 避免 |
|--------|--------|
| **ForField**: 绑定到 State 字段 | **直接传递 setter** |
| **Store.Get()**: 每次渲染时读取 | **缓存 State** |
| **自定义 Intent**: 业务逻辑 | **通用 Intent** |
| **自动注册**: BuildAndRegister | **手动 RegisterIntent** |

---

## 八、常见模式

### 模式 1: 表单验证

```go
type SubmitFormIntent struct{}

var appReducer = reducer.NewBuilder[AppState]().
    On(SubmitFormIntent{}, func(s AppState, i intent.Intent) AppState {
        // 验证所有字段
        if len(s.Username) < 3 {
            s.UsernameErr = "用户名至少3个字符"
        }
        if len(s.Email) < 5 {
            s.EmailErr = "邮箱至少5个字符"
        }
        return s
    }).
    On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
        // 实时验证
        if len(s.Username) < 3 {
            s.UsernameErr = "用户名至少3个字符"
        } else {
            s.UsernameErr = ""
        }
        return s
    })
```

### 模式 2: 异步数据加载

```go
type LoadDataIntent struct{}

var appReducer = reducer.NewBuilder[AppState]().
    On(LoadDataIntent{}, func(s AppState, i intent.Intent) AppState {
        s.IsLoading = true
        s.Data = nil
        return s
        // 异步加载在侧 effects 中
    })

// 在 effect 中处理异步
ui.UseEffect(func() func() CleanupFunc {
    go func() {
        data := fetchData()
        appStore.Set(func(s AppState) AppState {
            s.IsLoading = false
            s.Data = data
            return s
        })
    }, nil)
```

### 模式 3: 条件渲染

```go
func App() ui.VNode {
    state := appStore.Get()
    
    if state.IsLoading {
        return ui.Text("Loading...")
    }
    
    if state.Error != "" {
        return ui.NewTextBuilder("Error: " + state.Error).
            FgColor("red").
            Build()
    }
    
    // 正常渲染
    return renderForm(state)
}
```

---

## 九、调试

### 9.1 启用调试日志

```go
import "github.com/wwsheng009/mint/internal/log"

func main() {
    log.UILogger.Enable(true)  // 启用日志
    
    err := ui.Run(App)
    if err != nil {
        panic(err)
    }
}
```

### 9.2 检查 State 变化

```go
// 在 View 中打印 State（调试用）
func App() ui.VNode {
    state := appStore.Get()
    fmt.Printf("State: %+v\n", state)
    
    return ui.VStack(...)
}
```

### 9.3 暂停 Store 订阅

```go
// 订阅 State 变化
_ = appStore.Subscribe(func(oldState, newState AppState) {
    fmt.Printf("State changed: %v -> %v\n", oldState, newState)
})
```

---

## 十、性能优化

### 10.1 避免 Store 过度订阅

```go
// ❌ 错误：每次渲染都订阅
func BadComponent() ui.VNode {
    unsubscribe := appStore.Subscribe(func(state AppState) {
        // 每次渲染都会创建新的订阅，导致内存泄漏
        render(state)
    })
    return ui.Text("...")
}

// ✅ 正确：全局订阅一次
func goodComponent() ui.VNode {
    state := appStore.Get()
    return render(state)
}
```

### 10.2 使用 Selector 避免重复计算

```go
// ❌ 每次都计算
func App() ui.VNode {
    state := appStore.Get()
    username := state.Username
    usernameLength := len(username)  // 每次渲染都计算
    
    return ui.Text(fmt.Sprintf("Length: %d", usernameLength))
}

// ✅ 使用 Computed 缓存
var usernameLengthComputed store.Computed[AppState, int]

func initStore() {
    appStore := store.NewStore(AppState{Username: ""})
    usernameLengthComputed = store.NewComputed(appStore,
        func(s AppState) int {
            return len(s.Username)
        })
}

func App() ui.VNode {
    state := appStore.Get()
    length := usernameLengthComputed.Get()  // 自动缓存
    
    return ui.Text(fmt.Sprintf("Length: %d", length))
}
```

### 10.3 减少不必要的渲染

```go
// ❌ 每次状态变化都重新渲染整个应用
type AppState struct {
    Count int
    SubCount int
}

// ✅ 优化：只订阅需要的状态
func TopComponent() ui.VNode {
    appStore := store.NewStore(AppState{Count: 0, SubCount: 0})
    
    appStore.Subscribe(func(state AppState) {
        // 只在 Count 变化时重新渲染
        renderTop(state.Count)
    }, func(state AppState) AppState {
        // 只有 Count 变化时才通知
        return state.Count != state.SubCount
    })
}
```

---

## 总结

Store + Reducer 架构提供了：

| 优势 | 说明 |
|------|------|
| ✅ **状态清晰** | 单一状态源，状态变化可预测 |
| ✅ **易于测试** | Reducer 是纯函数，易于单元测试 |
| ✅ **易于调试** | 状态变化可追踪，支持时间旅行 |
| ✅ **易于扩展** | 通过 Reducer 组合和 Middleware 扩展 |
| ✅ **类型安全** | 编译期类型检查 |

**关键模式**：
1. State: 扁平、类型安全、可序列化
2. Reducer: 纯函数、不可变、处理所有 Intent
3. View: 从 Store 读取、无状态
4. Auto-Register: 使用 BuildAndRegister 自动注册 handlers

---

## 相关文档

- [迁移指南](MIGRATION_GUIDE.md) - 从 UseState 迁移
- [API 参考](API_REFERENCE.md) - Store 和 Reducer API
- [最佳实践](BEST_PRACTICES.md) - 更多最佳实践
