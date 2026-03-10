# GlobalState 弃用公告

**状态**: ⚠️ **Deprecated** - 将于 v1.0 版本移除
**公告日期**: 2026-03-07
**迁移目标**: v1.0

---

## 概述

`ComponentContext.GlobalState` 及相关方法（`GetState`, `SetState`, `GetGlobalState` 等）已正式标记为 **Deprecated**。建议所有新代码使用 **Store + Reducer** 架构，现有代码应尽快迁移。

---

## 为什么弃用 GlobalState？

GlobalState 存在以下问题：

1. **类型不安全**
   ```go
   // ❌ 需要类型断言，容易出错
   setter, ok := ctx.GetState("usernameSetter")
   if setter != nil {
       setter.(func(string))("new value")  // 危险的类型断言
   }
   ```

2. **字符串键无法编译时检查**
   ```go
   // ❌ 拼写错误不会在编译时发现
   ctx.SetState("usernam", "john")  // 拼写错误！
   ```

3. **三重状态系统导致混乱**
   - 组件局部状态（UseState）
   - 全局状态（GlobalState）
   - setter 闭包

4. **单一数据源原则被破坏**
   - 状态分散在多个地方
   - 难以追踪数据流

---

## 推荐的迁移方案

### 方案 1: 使用 Store + Reducer（推荐用于应用级状态）

```go
// ✅ 新方式：Store + Reducer
type AppState struct {
    Username string
    Counter  int
}

var appStore = store.NewStore(AppState{})

// 组件中使用
username, setUsername := ui.UseStoreField(
    appStore,
    func(s AppState) string { return s.Username },
    func(s AppState, v string) AppState { s.Username = v; return s },
)

// Intent Handler 直接操作 Store
func handleUsernameChange(ctx context.ActionContext) {
    appStore.Update(func(s AppState) AppState {
        s.Username = "new value"
        return s
    })
}
```

### 方案 2: 使用 UseStoreField/UseStoreSelector（推荐用于组件级状态）

```go
// ✅ 类型安全的 Field 订阅
value, setValue := ui.UseStoreField(
    store,
    selector,
    reducer,
)

// ✅ 支持函数式更新
setValue(func(old int) int {
    return old + 1  // increment
})
```

---

## 迁移对照表

| GlobalState (旧) | Store / Hooks (新) |
|-----------------|---------------------|
| `ctx.SetState("key", value)` | `store.Update(func(s) S { s.Field = v; return s })` |
| `ctx.GetState("key")` | `Store.Get().Field` |
| `ctx.GetIntState("key", 0)` | `UseStoreField(store, intSelector, intReducer)` |
| `ctx.GetStringState("key", "")` | `UseStoreField(store, stringSelector, stringReducer)` |
| `ctx.GetBoolState("key", false)` | `UseStoreField(store, boolSelector, boolReducer)` |

---

## 分步迁移指南

### 第 1 步：定义 Store State

```go
// 从多个 GlobalState 迁移到单一 State
type AppState struct {
    InputText  string  // 原 ctx.GetState("inputText")
    Submitted  bool    // 原 ctx.GetState("submitted")
    Count      int     // 原 ctx.GetIntState("count", 0)
}

var appStore = store.NewStore(AppState{
    InputText: "",
    Submitted: false,
    Count:     0,
})
```

### 第 2 步：替换 GetState 为 UseStoreField

```go
// ❌ 旧代码
inputValue, setInputValue := ui.UseStateString("")
inputSetterKey := intent.NewKey("inputSetter")

ui.WithOnMount(func(actx context.ActionContext) {
    ctx := GetContext(actx)
    ctx.SetState(inputSetterKey.String(), setInputValue)
})

// ✅ 新代码
inputValue, setInputValue := ui.UseStoreField(
    appStore,
    func(s AppState) string { return s.InputText },
    func(s AppState, v string) AppState { s.InputText = v; return s },
)
```

### 第 3 步：替换 Intent Handler 中的 GetState/类型断言

```go
// ❌ 旧代码
func handleChange(actx context.ActionContext) {
    ctx := GetContext(actx)
    if setter, ok := ctx.GetState("inputSetter"); ok {
        setter.(func(string))(newValue)  // 类型断言！
    }
}

// ✅ 新代码
func handleChange(actx context.ActionContext) {
    appStore.Update(func(s AppState) AppState {
        s.InputText = "new value"
        return s
    })
}
```

### 第 4 步：移除 OnMount 中的 setter 注册

```go
// ❌ 旧代码
ui.WithOnMount(func(actx context.ActionContext) {
    ctx := GetContext(actx)
    ctx.SetState("usernameSetter", setUsername)
    ctx.SetState("emailSetter", setEmail)
    // ... 更多 setter 注册
})

// ✅ 新代码：无需 OnMount 注册，直接使用 Store
```

---

## 迁移示例

### 完整迁移示例

**旧代码（GlobalState）**：
```go
func CounterDemo() ui.VNode {
    count, setCount := ui.UseStateInt(0)

    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.Button("+", ui.WithOnClick(func(actx context.ActionContext) {
            ctx := GetContext(actx)
            if fn, ok := ctx.GetState("setCount"); ok {
                fn.(func(int))(count + 1)  // ❌ 类型断言！
            }
        })),
        // ❌ 还需要 OnMount 注册 setter
        ui.WithOnMount(func(actx context.ActionContext) {
            ctx := GetContext(actx)
            ctx.SetState("setCount", setCount)
        }),
    )
}
```

**新代码（Store + Reducer）**：
```go
type AppState struct {
    Count int
}

var counterStore = store.NewStore(AppState{Count: 0})

func CounterDemo() ui.VNode {
    count, setCount := ui.UseStoreField(
        counterStore,
        func(s AppState) int { return s.Count },
        func(s AppState, v int) AppState { s.Count = v; return s },
    )

    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.Button("+", ui.WithOnClick(func(actx context.ActionContext) {
            // ✅ 类型安全，无需闭包
            counterStore.Update(func(s AppState) AppState {
                s.Count++
                return s
            })
        })),
        // ✅ 无需 OnMount
    )
}
```

---

## 已废弃的 API 列表

以下 API 已标记为 `Deprecated`，将在 v1.0 中移除：

### 字段
- `ComponentContext.GlobalState` - 使用 `store.Store` 代替

### 方法
- `ComponentContext.GetState()` - 使用 `UseStoreField()` 或 `Store.Get()` 代替
- `ComponentContext.SetState()` - 使用 `Store.Update()` 代替
- `ComponentContext.GetStringState()` - 使用 `UseStoreField()` 代替
- `ComponentContext.GetIntState()` - 使用 `UseStoreField()` 代替
- `ComponentContext.GetBoolState()` - 使用 `UseStoreField()` 代替
- `ComponentContext.GetGlobalState()` - 使用 `UseStoreField()` 代替
- `ComponentContext.SetGlobalState()` - 使用 `Store.Update()` 代替
- `ComponentContext.GetGlobalString()` - 使用 `UseStoreField()` 代替
- `ComponentContext.GetGlobalInt()` - 使用 `UseStoreField()` 代替
- `ComponentContext.GetGlobalBool()` - 使用 `UseStoreField()` 代替

---

## 迁移进度

已有以下示例完成迁移：

| 示例 | 状态 | 代码变化 |
|------|------|---------|
| focus_switching_demo | ✅ 完成 | 220 → 170 (-23%) |
| validation_demo | ✅ 完成 | 382 → 382 (0%, 改进) |
| mvp_form_demo | ✅ 完成 | 271 → 271 (0%, 改进) |
| mvp_components_demo | ✅ 完成 | 380 → 380 (0%, 改进) |
| typesafe_form_demo | ✅ 完成 | 197 → 172 (-13%) |
| ant_design_demo | ✅ 完成 | 429 → 293 (-32%) |
| checkbox demo | ✅ 完成 | 124 → 64 (-48%) |
| absolute demo | ✅ 完成 | 94 → 84 (-11%) |
| counter demo | ✅ 完成 | 91 → 101 (+11%) |

---

## 常见问题 (FAQ)

### Q1: GlobalState 什么时候会被移除？
**A**: 计划在 v1.0 版本中移除。在此之前，它会保持向后兼容性。

### Q2: 我的新项目应该使用什么状态管理？
**A**:
- **应用级全局状态**: 使用 `Store + Reducer`
- **组件级状态**: 使用 `UseStoreField` / `UseStoreSelector`
- **简单组件**: 也可以使用 `useState` (不建议跨组件共享)

### Q3: 如何处理需要跨组件共享的状态？
**A**: 创建一个全局 Store 实例，在需要使用的地方通过 `UseStoreField` 订阅：

```go
// main.go - 初始化
var appStore = store.NewStore(AppState{})

// component_a.go - 使用
value, setValue := ui.UseStoreField(appStore, ...)

// component_b.go - 使用
value, setValue := ui.UseStoreField(appStore, ...)
```

### Q4: 迁移后代码量会增加吗？
**A**: 不一定。根据迁移数据，大多数示例代码量持平或减少：
- 50% 示例代码量减少（最多减少 48%）
- 50% 示例代码量持平
- 只有极少数示例代码量略微增加（但代码更清晰）

### Q5: GlobalState 还能在 v0.x 版本中使用吗？
**A**: 可以。在 v1.0 之前的所有版本中，GlobalState 会保持向后兼容。但强烈建议新代码避免使用，现有代码尽快迁移。

---

## 相关文档

- [Store + Reducer 架构指南](../store/README.md)
- [混合模式状态管理指南](../store/hybrid/STATE_MANAGEMENT_GUIDE.md)
- [迁移指南：从 GlobalState 到 Store](../store/guides/MIGRATION_GUIDE.md)
- [迁移进度](../store/status/MIGRATION_PROGRESS.md)

---

## 获取帮助

如果您在迁移过程中遇到问题：

1. 查阅 [迁移指南](../store/guides/MIGRATION_GUIDE.md)
2. 参考已迁移的示例（见"迁移进度"部分）
3. 提交 Issue 寻求帮助

---

**最后更新**: 2026-03-07
