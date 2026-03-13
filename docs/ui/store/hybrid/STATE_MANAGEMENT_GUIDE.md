# Mint UI 状态管理指南

**创建时间**: 2026-03-07
**版本**: v1.0
**适用版本**: Mint UI v0.11+

---

## ⚠️ 重要提示：GlobalState 已弃用

**状态**: `ComponentContext.GlobalState` 及相关方法已标记为 **Deprecated**

本指南介绍的混合模式（UseStoreField/UseStoreSelector）是为了帮助平滑过渡到 Store + Reducer 架构。对于新代码，强烈建议直接使用 **Store + Reducer** 架构。

**详细说明**: [GlobalState 弃用公告](/docsArchive/GLOBALSTATE_DEPRECATION.md) | [迁移指南](../guides/MIGRATION_GUIDE.md)

---

## 目录

- [概述](#概述)
- [状态管理方案对比](#状态管理方案对比)
- [选择合适的状态管理方案](#选择合适的状态管理方案)
- [详细使用指南](#详细使用指南)
- [迁移路径](#迁移路径)
- [最佳实践](#最佳实践)

---

## 概述

Mint UI 提供了三种状态管理方案，以满足不同场景的需求：

| 方案 | 适用场景 | 复杂度 | 推荐指数 |
|------|---------|--------|---------|
| **useState** | 组件内部局部状态 | ⭐ 简单 | ⭐⭐⭐ |
| **UseStoreField** | 从 Store 订阅特定字段（混合模式） | ⭐⭐ 中等 | ⭐⭐⭐⭐⭐ |
| **Store + Reducer** | 应用全局状态管理 | ⭐⭐⭐ 复杂 | ⭐⭐⭐⭐ |

---

## 状态管理方案对比

### 1. useState Hook - 局部状态

**特点**：
- React 风格的简单 API
- 组件内部状态，不共享
- 自动触发重渲染
- 可能有闭包引用问题（已修复）

**适用场景**：
- ✅ 简单的 UI 状态（展开/折叠、选中/未选中）
- ✅ 不需要跨组件共享的状态
- ✅ 快速原型开发

**不适用场景**：
- ❌ 需要跨组件通信
- ❌ 复杂的业务逻辑
- ❌ 需要持久化的状态

**示例**：

```go
func ExpanderComponent() ui.VNode {
    // 简单的展开/折叠状态
    expanded, setExpanded := ui.UseStateBool(false)

    return ui.VStack(
        ui.NewButtonBuilder(fmt.Sprintf("%s", map[bool]string{
            true:  "▼",
            false: "▶",
        }[expanded])).
            OnPress(func() {
                setExpanded(!expanded)
            }).
            Build(),

        // 条件渲染
        func() ui.VNode {
            if expanded {
                return ui.NewTextBuilder("展开的内容").Build()
            }
            return ui.VNode{}
        }(),
    )
}
```

---

### 2. UseStoreField - 混合模式（推荐）⭐

**特点**：
- Hook 风格的简单 API
- 使用 Store 架构作为数据源
- 自动订阅和重渲染
- 类型安全，无闭包问题
- 渐进式迁移路径

**适用场景**：
- ✅ 表单字段状态管理
- ✅ 需要跨组件共享的字段
- ✅ 从 useState 迁移到 Store 的过渡
- ✅ 需要简单 API 和强大架构并存

**示例**：

```go
// 定义应用状态
type AppState struct {
    Username string
    Email    string
    Count    int
    Agreed   bool
}

// 全局 Store
var appStore = store.NewStore(AppState{
    Username: "",
    Email:    "",
    Count:    0,
    Agreed:   false,
})

// Reducer
var appReducer = reducer.BindField(reducer.NewBuilder[AppState]()).
    BindStringField("username", func(s *AppState, v string) {
        s.Username = v
    }).
    BindStringField("email", func(s *AppState, v string) {
        s.Email = v
    }).
    BindIntField("count", func(s *AppState, v int) {
        s.Count = v
    }).
    BindBoolField("agreed", func(s *AppState, v bool) {
        s.Agreed = v
    }).
    GetBuilder().
    BuildAndRegister(intent.DefaultRegistry(), appStore)

// 组件中使用
func LoginForm() ui.VNode {
    // 使用 UseStoreField - Hook 风格 API
    username, setUsername := ui.UseStoreField(
        appStore,
        func(s AppState) string { return s.Username },
        func(s AppState, v string) AppState { s.Username = v; return s },
    )

    email, setEmail := ui.UseStoreField(
        appStore,
        func(s AppState) string { return s.Email },
        func(s AppState, v string) AppState { s.Email = v; return s },
    )

    agreed, setAgreed := ui.UseStoreField(
        appStore,
        func(s AppState) bool { return s.Agreed },
        func(s AppState, v bool) AppState { s.Agreed = v; return s },
    )

    return ui.VStack(
        ui.NewInputBuilder().
            Value(username).
            OnChange(setUsername).
            Placeholder("Username").
            Build(),

        ui.NewInputBuilder().
            Value(email).
            OnChange(setEmail).
            Placeholder("Email").
            Build(),

        ui.NewCheckboxBuilder().
            Checked(agreed).
            OnToggle(func(_ ui.VNode) {
                setAgreed(!agreed)
            }).
            Label("I agree").
            Build(),
    )
}
```

---

### 3. Store + Reducer - 全局架构

**特点**：
- 单一数据源
- 声明式状态更新
- 类型安全
- 支持中间件、时间旅行等高级功能
- 完整的架构规范

**适用场景**：
- ✅ 大型应用的状态管理
- ✅ 复杂的业务逻辑
- ✅ 需要时间旅行调试
- ✅ 需要全局审计日志
- ✅ 多人协作项目

**示例**：

```go
// 定义应用状态
type AppState struct {
    Count    int
    Username string
    Items    []Item
}

// 定义 Intents
type IncrementIntent struct {
    Amount int
}

func (IncrementIntent) IntentType() string {
    return "Increment"
}

type DecrementIntent struct {
    Amount int
}

func (DecrementIntent) IntentType() string {
    return "Decrement"
}

type AddItemIntent struct {
    Item Item
}

func (AddItemIntent) IntentType() string {
    return "AddItem"
}

// 创建 Store
var appStore = store.NewStore(AppState{})

// 创建 Reducer
var appReducer = reducer.NewBuilder[AppState]().
    On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
        s.Count += 1
        return s
    }).
    On(DecrementIntent{}, func(s AppState, i intent.Intent) AppState {
        s.Count -= 1
        return s
    }).
    On(AddItemIntent{}, func(s AppState, i intent.Intent) AppState {
        if addItem, ok := i.(AddItemIntent); ok {
            s.Items = append(s.Items, addItem.Item)
        }
        return s
    }).
    BuildAndRegister(intent.DefaultRegistry(), appStore)

// 组件中使用
func CounterComponent() ui.VNode {
    // 直接读取 Store
    state := appStore.Get()

    return ui.VStack(
        ui.NewTextBuilder(fmt.Sprintf("Count: %d", state.Count)).Build(),
        ui.NewButtonBuilder("Increment").
            OnPress(IncrementIntent{Amount: 1}).
            Build(),
        ui.NewButtonBuilder("Decrement").
            OnPress(DecrementIntent{Amount: 1}).
            Build(),
    )
}
```

---

## 选择合适的状态管理方案

### 决策树

```
是否需要跨组件共享？
│
├─ 否 → useState
│   └─ 简单的局部状态
│
└─ 是 → 是否需要复杂的业务逻辑？
    │
    ├─ 否 → UseStoreField
    │   └─ 表单字段、简单共享状态
    │
    └─ 是 → Store + Reducer
        └─ 完整的应用状态管理
```

### 快速参考表

| 场景 | 推荐方案 | 原因 |
|------|---------|------|
| 展开折叠菜单 | useState | 简单，不需要共享 |
| 模态框打开/关闭 | useState | UI 临时状态 |
| 表单输入 | UseStoreField | 需要跨组件验证 |
| 用户登录状态 | Store + Reducer | 全局、持久化 |
| 购物车 | Store + Reducer | 复杂业务逻辑 |
| 首偏好设置 | UseStoreField | 需要共享但简单 |
| 列表筛选 | UseStoreSelector | 派生状态 |
| 实时数据 | Store + Reducer | 统一数据流 |

---

## 详细使用指南

### UseStoreField 详解

#### 函数签名

```go
func UseStoreField[S any, T any](
    store *store.Store[S],
    selector func(S) T,      // 从 State 读取字段
    setter func(S, T) S,      // 更新字段并返回新 State
) (T, func(T))              // 返回：当前值和设置函数
```

#### 参数说明

1. **store**: Store 实例
2. **selector**: 从 State 中提取特定字段的函数
3. **setter**: 更新字段并返回新 State 的函数

#### 返回值

1. **T**: 当前字段值
2. **func(T)**: 设置字段的函数（自动触发 Store 更新）

#### 完整示例

```go
package main

import (
    "fmt"
    "github.com/wwsheng009/mint/runtime/store"
    "github.com/wwsheng009/mint/runtime/reducer"
    "github.com/wwsheng009/mint/runtime/intent"
    "github.com/wwsheng009/mint/ui"
)

type AppState struct {
    Username     string
    Email        string
    Password     string
    RememberMe   bool
}

var appStore = store.NewStore(AppState{})

// 设置 Reducer
var appReducer = reducer.BindField(reducer.NewBuilder[AppState]()).
    BindStringField("username", func(s *AppState, v string) { s.Username = v }).
    BindStringField("email", func(s *AppState, v string) { s.Email = v }).
    BindStringField("password", func(s *AppState, v string) { s.Password = v }).
    BindBoolField("rememberMe", func(s *AppState, v bool) { s.RememberMe = v }).
    GetBuilder().
    BuildAndRegister(intent.DefaultRegistry(), appStore)

func RegisterForm() ui.VNode {
    username, setUsername := ui.UseStoreField(
        appStore,
        func(s AppState) string { return s.Username },
        func(s AppState, v string) AppState { s.Username = v; return s },
    )

    email, setEmail := ui.UseStoreField(
        appStore,
        func(s AppState) string { return s.Email },
        func(s AppState, v string) AppState { s.Email = v; return s },
    )

    password, setPassword := ui.UseStoreField(
        appStore,
        func(s AppState) string { return s.Password },
        func(s AppState, v string) AppState { s.Password = v; return s },
    )

    rememberMe, setRememberMe := ui.UseStoreField(
        appStore,
        func(s AppState) bool { return s.RememberMe },
        func(s AppState, v bool) AppState { s.RememberMe = v; return s },
    )

    return ui.VStack(
        ui.NewTextBuilder("用户注册").Build(),

        ui.NewInputBuilder().
            Value(username).
            OnChange(setUsername).
            Placeholder("用户名").
            Build(),

        ui.NewInputBuilder().
            Value(email).
            OnChange(setEmail).
            Placeholder("邮箱").
            Build(),

        ui.NewInputBuilder().
            Value(password).
            OnChange(setPassword).
            Placeholder("密码").
            Password(true).
            Build(),

        ui.NewCheckboxBuilder().
            Checked(rememberMe).
            OnToggle(func() { setRememberMe(!rememberMe) }).
            Label("记住我").
            Build(),
    )
}
```

---

### UseStoreFieldFunctional 详解

#### 函数签名

```go
func UseStoreFieldFunctional[S any, T any](
    store *store.Store[S],
    selector func(S) T,
    setter func(S, T) S,
) (T, func(interface{}))
```

#### 用途

`UseStoreFieldFunctional` 与 `UseStoreField` 类似，但支持**功能性更新**。

功能更新（Functional Update）允许基于当前值计算新值，这在以下场景非常有用：

- 计数器：`setCount(func(old) old + 1)`
- 切换布尔值：`setToggle(func(old) !old)`
- 追加列表：`setList(func(old) append(old, newItem))`

#### 支持的更新方式

**1. 直接设置值**

```go
setCount(10)
```

**2. 函数式更新**

```go
setCount(func(old int) int {
    return old + 1
})
```

#### 完整示例

```go
package main

type AppState struct {
    Count int
}

var appStore = store.NewStore(AppState{})

func Counter() ui.VNode {
    // 使用 UseStoreFieldFunctional
    count, setCount := ui.UseStoreFieldFunctional(
        appStore,
        func(s AppState) int { return s.Count },
        func(s AppState, v int) AppState { s.Count = v; return s },
    )

    return ui.VStack(
        ui.NewTextBuilder(fmt.Sprintf("计数: %d", count)).Build(),

        // 直接设置值
        ui.NewButtonBuilder("设置为 10").
            OnPress(func() { setCount(10) }).
            Build(),

        // 函数式更新 - 加 1
        ui.NewButtonBuilder("+1").
            OnPress(func() {
                setCount(func(old int) int {
                    return old + 1
                })
            }).
            Build(),

        // 函数式更新 - 减 1
        ui.NewButtonBuilder("-1").
            OnPress(func() {
                setCount(func(old int) int {
                    return old - 1
                })
            }).
            Build(),

        // 函数式更新 - 翻倍
        ui.NewButtonBuilder("翻倍").
            OnPress(func() {
                setCount(func(old int) int {
                    return old * 2
                })
            }).
            Build(),
    )
}
```

#### 什么时候使用功能性更新？

| 场景 | 推荐 | 原因 |
|------|-----|------|
| 设置固定值 | 直接设置 | 代码更清晰 |
| 基于当前值计算 | 函数式更新 | 避免闭包捕获旧值 |
| 多步聚合更新 | 函数式更新 | 原子操作，避免竞态 |

#### 优势

1. **避免闭包捕获旧值**
   ```go
   // ❌ 可能捕获旧值
   count, setCount := ui.UseStoreField(appStore, ...)
   // 点击多次，每次都+1，但可能捕获的是同一个 count
   ui.Button("加1").OnClick(func() {
       setCount(count + 1)  // count 可能是旧的
   })

   // ✅ 函数式更新，总是获取最新值
   count, setCount := ui.UseStoreFieldFunctional(appStore, ...)
   ui.Button("加1").OnClick(func() {
       setCount(func(old int) int {
           return old + 1  // old 是最新的
       })
   })
   ```

2. **更易理解**
   ```go
   // 清晰地表达"增加 1"的意图
   setCount(func(old int) int { return old + 1 })
   ```

3. **支持多步原子操作**
   ```go
   // 原子地实现"如果 > 10 则重置为 0"
   setCount(func(old int) int {
       if old > 10 {
           return 0
       }
       return old
   })
   ```

---

### UseStoreSelector 详解

#### 函数签名

```go
func UseStoreSelector[S any, T any](
    store *store.Store[S],
    selector func(S) T,      // 从 State 派生值
) T                         // 返回：派生值
```

#### 用途

- 计算派生状态（derived state）
- 过滤、映射集合数据
- 聚合计算

#### 示例

```go
type AppState struct {
    Items     []Item
    Filter    string
}

var appStore = store.NewStore(AppState{})

func ShoppingList() ui.VNode {
    // 直接 Store 访问
    filter := ui.UseStoreSelector(
        appStore,
        func(s AppState) string { return s.Filter },
    )

    // 派生状态：过滤后的列表
    filteredItems := ui.UseStoreSelector(
        appStore,
        func(s AppState) []Item {
            if s.Filter == "" {
                return s.Items
            }
            filtered := make([]Item, 0)
            for _, item := range s.Items {
                if strings.Contains(item.Name, s.Filter) {
                    filtered = append(filtered, item)
                }
            }
            return filtered
        },
    )

    // 派生状态：总数量
    itemCount := ui.UseStoreSelector(
        appStore,
        func(s AppState) int {
            count := 0
            for _, item := range s.Items {
                count += item.Quantity
            }
            return count
        },
    )

    return ui.VStack(
        ui.NewInputBuilder().
            Value(filter).
            OnChange(func(v string) {
                appStore.Update(func(s AppState) AppState {
                    s.Filter = v
                    return s
                })
            }).
            Placeholder("筛选").
            Build(),

        ui.NewTextBuilder(fmt.Sprintf("共 %d 件商品", itemCount)).Build(),

        // 渲染过滤后的列表
        ui.VStack(func() []ui.VNode {
            nodes := make([]ui.VNode, len(filteredItems))
            for i, item := range filteredItems {
                nodes[i] = ui.NewTextBuilder(
                    fmt.Sprintf("- %s x%d", item.Name, item.Quantity),
                ).Build()
            }
            return nodes
        }()...),
    )
}
```

#### 性能优化：深度比较（Deep Comparison）

`UseStoreSelector` 使用深度比较算法来确定派生值是否发生变化。这确保了：

1. **切片和 Map 的正确比较**：即使切片/Map 的底层数组地址不同，只要内容相同，就不会触发不必要的重新渲染。

2. **复杂的嵌套数据结构**：自动递归比较嵌套的 struct、slice、map、pointer 等类型。

3. **避免不必要的渲染**：只有当派生值真正发生变化时，组件才会重新渲染。

**示例**：

```go
// Store 中的 Items 字段被更新为新切片（内容相同）
itemsCopy := []string{"a", "b", "c"}
store.Update(func(s AppState) AppState {
    s.Items = itemsCopy  // 新 slice，内容相同
    return s
})

// UseStoreSelector 会检测到相同内容，不会触发重新渲染
items := ui.UseStoreSelector(
    store,
    func(s AppState) []string { return s.Items },
)
```

**注意**：
- 深度比较使用 Go 反射机制，对于非常深或非常大的数据结构可能会有性能开销
- 对于大型数据集，建议使用 UseStoreField 订阅特定字段，而不是直接订阅整个切片

---

### UseStoreComputed 详解

#### 函数签名

```go
func UseStoreComputed[T any](
    computed *store.Computed[T, T],
) T
```

#### 用途

- 订阅 Store 的计算值
- 自动缓存和更新

#### 示例

```go
type AppState struct {
    Items []Item
}

var appStore = store.NewStore(AppState{})

// 创建计算值（组件外部）
var totalPrice = store.NewComputed(appStore, func(s AppState) float64 {
    total := 0.0
    for _, item := range s.Items {
        total += item.Price * float64(item.Quantity)
    }
    return total
})

var itemCount = store.NewComputed(appStore, func(s AppState) int {
    count := 0
    for _, item := range s.Items {
        count += item.Quantity
    }
    return count
})

func ShoppingList() ui.VNode {
    // 订阅计算值
    total := ui.UseStoreComputed(totalPrice)
    count := ui.UseStoreComputed(itemCount)

    return ui.VStack(
        ui.NewTextBuilder(fmt.Sprintf("共 %d 件商品", count)).Build(),
        ui.NewTextBuilder(fmt.Sprintf("总价: %.2f", total)).Build(),
    )
}
```

---

## 迁移路径

### 从 useState 迁移到 UseStoreField

#### 步骤 1: 定义 AppState

```go
// 之前：分散的 useState
func MyComponent() ui.VNode {
    username, setUsername := ui.UseStateString("")
    email, setEmail := ui.UseStateString("")
    // ...
}

// 之后：统一的 AppState
type AppState struct {
    Username string
    Email    string
}

var appStore = store.NewStore(AppState{})
```

#### 步骤 2: 替换 useState 为 UseStoreField

```go
// 之前
func MyComponent() ui.VNode {
    username, setUsername := ui.UseStateString("")
    email, setEmail := ui.UseStateString("")
    // ...
}

// 之后
func MyComponent() ui.VNode {
    username, setUsername := ui.UseStoreField(
        appStore,
        func(s AppState) string { return s.Username },
        func(s AppState, v string) AppState { s.Username = v; return s },
    )

    email, setEmail := ui.UseStoreField(
        appStore,
        func(s AppState) string { return s.Email },
        func(s AppState, v string) AppState { s.Email = v; return s },
    )
    // ...
}
```

#### 步骤 3: 设置 Reducer

```go
var appReducer = reducer.BindField(reducer.NewBuilder[AppState]()).
    BindStringField("username", func(s *AppState, v string) {
        s.Username = v
    }).
    BindStringField("email", func(s *AppState, v string) {
        s.Email = v
    }).
    GetBuilder().
    BuildAndRegister(intent.DefaultRegistry(), appStore)
```

---

### 从 MVP 模式迁移到混合模式

#### 之前的 MVP 模式（需要 GlobalState 同步）：

```go
// ❌ 旧代码：需要手动同步
func MyForm() ui.VNode {
    username, setUsername := ui.UseStateString("")
    email, setEmail := ui.UseStateString("")

    // 手动同步到 GlobalState
    ctx := ui.GetCurrentContext()
    if ctx != nil {
        ctx.GlobalState["setUsername"] = setUsername
        ctx.GlobalState["setEmail"] = setEmail
    }

    return ui.VStack(
        ui.NewInputBuilder().
            ForField(intent.ForField("username")).
            Value(username).
            Build(),
        ui.NewInputBuilder().
            ForField(intent.ForField("email")).
            Value(email).
            Build(),
    )
}

// 还需要手动注册 Handler
ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
    switch i.Field {
    case "username":
        if fn, ok := ctx.GetState("setUsername"); ok {
            if setter, ok := fn.(func(string)); ok {
                setter(i.Value)
            }
        }
    case "email":
        // ...
    }
    return intent.HandledResult()
})
```

#### 新的混合模式（自动同步）：

```go
// ✅ 新代码：自动订阅和更新
func MyForm() ui.VNode {
    username, setUsername := ui.UseStoreField(
        appStore,
        func(s AppState) string { return s.Username },
        func(s AppState, v string) AppState { s.Username = v; return s },
    )

    email, setEmail := ui.UseStoreField(
        appStore,
        func(s AppState) string { return s.Email },
        func(s AppState, v string) AppState { s.Email = v; return s },
    )

    return ui.VStack(
        ui.NewInputBuilder().
            Value(username).
            OnChange(setUsername).
            Placeholder("Username").
            Build(),
        ui.NewInputBuilder().
            Value(email).
            OnChange(setEmail).
            Placeholder("Email").
            Build(),
    )
}

// Reducer 自动处理，无需手动注册 Handler
```

---

## 最佳实践

### 1. 选择合适的状态管理方案

```go
// ✅ 好：局部状态用 useState
func Expander() ui.VNode {
    expanded, setExpanded := ui.UseStateBool(false)
    // ...
}

// ✅ 好：共享状态用 UseStoreField
func LoginForm() ui.VNode {
    username, setUsername := ui.UseStoreField(...)
    // ...
}

// ✅ 好：全局状态用 Store + Reducer
func App() ui.VNode {
    state := appStore.Get()
    // ...
}
```

### 2. 避免过度订阅

```go
// ❌ 不好：订阅整个 State
state := appStore.Get()

// ✅ 好：只订阅需要的字段
username := ui.UseStoreField(appStore,
    func(s AppState) string { return s.Username },
    func(s AppState, v string) AppState { s.Username = v; return s },
)
```

### 3. 使用派生状态避免重复计算

```go
// ❌ 不好：在组件中重复计算
func MyComponent() ui.VNode {
    items := ui.UseStoreField(...)
    totalPrice := 0.0
    for _, item := range items {
        totalPrice += item.Price
    }
    // ...
}

// ✅ 好：使用 Computed 或 Selector
var totalPrice = store.NewComputed(appStore, func(s AppState) float64 {
    total := 0.0
    for _, item := range s.Items {
        total += item.Price
    }
    return total
})
```

### 4. 函数式更新避免闭包陷阱

```go
// ⚠️ 可能有问题：闭包捕获旧值
func Counter() ui.VNode {
    count, setCount := ui.UseStoreField(appStore, ...)
    // 多次点击可能使用相同的 count 值
    ui.Button("加1").OnPress(func() {
        setCount(count + 1)
    })
}

// ✅ 好：使用函数式更新
func Counter() ui.VNode {
    count, setCount := ui.UseStoreFieldFunctional(appStore, ...)
    // 每次更新都基于最新值
    ui.Button("加1").OnPress(func() {
        setCount(func(old int) int {
            return old + 1
        })
    })
}
```

### 5. 选择合适的 Hook

| Hook | 用途 | 适用场景 |
|------|------|---------|
| `ui.UseStoreField` | 订阅单个字段 | 表单输入、简单字段 |
| `ui.UseStoreFieldFunctional` | 订阅字段（支持函数式更新） | 计数器、切换、增量更新 |
| `ui.UseStoreSelector` | 订阅派生状态 | 过滤、映射、聚合 |
| `ui.UseStoreComputed` | 订阅预计算的派生状态 | 高开销聚合、多个组件共享 |
| `ui.UseStore` | 直接访问 Store | 高级用法、特殊需求 |

### 6. 结构化 AppState

```go
// ✅ 好：按功能分组
type AppState struct {
    User    UserState
    Cart    CartState
    Filters FilterState
}

type UserState struct {
    Username string
    Email    string
}

type CartState struct {
    Items []CartItem
}

type FilterState struct {
    Query   string
    Filters map[string]bool
}

// 使用
username := ui.UseStoreField(appStore,
    func(s AppState) string { return s.User.Username },
    func(s AppState, v string) AppState {
        s.User.Username = v
        return s
    },
)
```

---

## 附录：快速参考

### Hook 对比表

| Hook | 返回值 | 订阅方式 | 功能 |
|------|--------|---------|------|
| `useState` | `(val, setFunc)` | 否 | 局部状态 |
| `UseStoreField` | `(val, setFunc)` | Store 字段 | 字段订阅 |
| `UseStoreFieldFunctional` | `(val, setFunc)` | Store 字段 | 字段订阅 + 函数式更新 |
| `UseStoreSelector` | `val` | Store 派生 | 派生状态 |
| `UseStoreComputed` | `val` | 计算值 | 缓存派生状态 |
| `UseStore` | `*Store[T]` | 无 | 直接访问 |

### API 索引

#### Store Hooks

```go
// 订阅 Store 字段
val, setVal := ui.UseStoreField(
    store,
    func(s S) T { return s.Field },
    func(s S, v T) S { s.Field = v; return s },
)

// 订阅 Store 字段（支持函数式更新）
val, setVal := ui.UseStoreFieldFunctional(
    store,
    func(s S) T { return s.Field },
    func(s S, v T) S { s.Field = v; return s },
)

// 使用：setVal(newValue) 或 setVal(func(old T) T { return old + 1 })
```

---

## 常见问题

### Q1: 当应该使用 useState vs UseStoreField？

**A**:
- **useState**: 组件内部、临时的 UI 状态
- **UseStoreField**: 需要跨组件共享、持久化的业务状态

---

### Q2: UseStoreField 和 Store + Reducer 可以混用吗？

**A**: 可以！这是混合模式的主要优势。例如：
- 简单的 UI 状态用 `useState`
- 表单字段用 `UseStoreField`
- 复杂的业务逻辑用 `Store + Reducer`

---

### Q3: UseStoreField 会导致性能问题吗？

**A**: 不会超过原始的 Store 订阅。实际上：
- 只订阅特定字段，避免整个 State 更新
- 使用 `Memo` hook 避免不必要的重渲染
- 可以使用 `UseStoreSelector` 进行更细粒度的控制

```go
// 优化：使用 memo 避免不必要的重渲染
memoizedComponent := ui.Memo func() ui.VNode {
    username, _ := ui.UseStoreField(...)
    return ui.NewInputBuilder().Value(username).Build()
}
```

---

### Q4: 如何从 GlobalState 迁移？

**A**:
1. 定义 AppState 结构
2. 创建 Store
3. 将 `ctx.GlobalState["key"]` 替换为 `UseStoreField`
4. 设置 Reducer 替代手动 Handler

```go
// 之前
username, setUsername := ui.UseStateString("")
ctx.GlobalState["username"] = setUsername

// 之后
username, setUsername := ui.UseStoreField(...)
```

---

### Q5: UseStoreField 支持函数式更新吗？

**A**: 当前不支持。如果需要函数式更新，有两种选择：

**选择 1**: 直接获取旧值计算新值

```go
func increment() {
    oldCount := appStore.Get().Count
    setCount(oldCount + 1)
}
```

**选择 2**: 使用 `appStore.Update`

```go
func increment() {
    appStore.Update(func(s AppState) AppState {
        s.Count++
        return s
    })
}
```

---

## 总结

| 方案 | 适用场景 | API 复杂度 | 推荐指数 |
|------|---------|-----------|---------|
| **useState** | 局部 UI 状态 | ⭐ 简单 | ⭐⭐⭐ |
| **UseStoreField** | 混合模式（推荐） | ⭐⭐ 中等 | ⭐⭐⭐⭐⭐ |
| **Store + Reducer** | 全局架构 | ⭐⭐⭐ 复杂 | ⭐⭐⭐⭐ |

**核心建议**：
1. 新项目优先使用 **UseStoreField**（混合模式）
2. 组件内部 UI 状态可以保留 `useState`
3. 大型应用、复杂业务使用 **Store + Reducer**
4. 渐进式迁移：从 `UseStoreField` 开始，逐步迁移到完整架构

---

## 相关文档

- [Store + Reducer API 参考](../api/API_REFERENCE.md)
- [MVP 迁移指南](/docsArchive/architecture/mvp/MVP_MIGRATION_GUIDE.md)
- [Field Binding 优化指南](../optimization/FIELD_BINDING_OPTIMIZATION.md)
- [混合模式示例](/examples/store_mixed_demo)

---

**文档创建**: 2026-03-07
**状态**: 完成 ✅
