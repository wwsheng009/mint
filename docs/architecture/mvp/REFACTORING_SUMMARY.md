# Mint UI 架构重构方案总结

**分支**: refactor/store-based-architecture  
**创建时间**: 2026-03-04  
**状态**: 分析完成

---

## 一、当前状态

### 1.1 已存在的 Store + Reducer 架构

**组件完整度**: 100%

| 组件 | 位置 | 状态 |
|------|------|------|
| **Store[T]** | `runtime/store/store.go` | ✅ 完整实现 |
| **Reducer[T]** | `runtime/reducer/reducer.go` | ✅ 完整实现 |
| **AppRuntime[T]** | `runtime/statemachine/runtime.go` | ✅ 完整实现 |
| **BuildAndRegister** | `runtime/reducer/reducer.go` | ✅ 完整实现 |
| **示例** | `examples/store_reducer_demo/main.go` | ✅ 可运行 |

### 1.2 store_reducer_demo 的正确实现

```go
// 1. 定义 State
type AppState struct {
    Count   int
    Username string
    Email    string
}

// 2. 创建全局 Store
appStore := store.NewStore(AppState{Count: 0, Username: "", Email: ""})

// 3. 定义 Reducer
reducerBuilder := reducer.NewBuilder[AppState]().
    On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
        s.Count++
        return s
    }).
    On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
        // FieldChangeIntent 已经由 Instance 自动发射
        // Reducer 直接更新 State
        if fieldChange, ok := i.(intent.FieldChangeIntent); ok {
            s.Username = fieldChange.Value
        }
        return s
    })

// 4. 注册 handlers（自动注册到 registry）
reducerBuilder.RegisterToGlobal(appStore)

// 5. 视图从 Store 读取
func App() ui.VNode {
    state := appStore.Get()  // 每次获取最新状态
    return ui.NewButtonBuilder("+").
        OnPress(IncrementIntent{}).
        Build()
}
```

**关键点**：
- Store 是**全局单例**，所有组件都从同一个 Store 读取
- BuildAndRegister 自动注册 handlers，不需要手动 WithInit
- FieldChangeIntent 由 Instance 通过 ForField 自动发射
- Handlers 会自动调用 ScheduleUpdate() 触发重新渲染

---

## 二、架构差异对比

### 2.1 错误架构 (focus_switching_demo)

```
❌ 当前错误实现
├─ UseState + GlobalState + Instance (三重状态系统)
├─Setter → GlobalState → Handler → 手动类型断言
├─WithInit 中手动注册 handlers
└─状态同步需要 5 步 + 类型断言
```

### 2.2 正确架构 (store_reducer_demo)

```
✅ 目标架构（已实现）
├─ Store (单一状态源)
├─Reducer (纯函数状态转换)
├─BuildAndRegister 自动注册
└─视图从 Store 读取
```

---

## 三、重构目标

### 3.1 统一架构

让所有应用都使用 Store + Reducer 架构：

```
┌─────────────────────────────────────────────────────────────┐
│                     目标架构                                 │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  [全局] Store[T] ← 单一真相源                               │
│                                                             │
│  ↓                                                          │
│  ViewFunction[T] (纯函数)                                   │
│    ├─ state := store.Get()                                 │
│    ├─ return ui.VNode(...)                                  │
│    └─ ForField 自动绑定 FieldChangeIntent                    │
│                                                             │
│  ↓                                                          │
│  Reducer[T] (纯函数)                                       │
│    ├─ 接收 Intent                                           │
│    ├─ 计算新 State                                         │
│    └─ 返回新 State                                         │
│                                                             │
│  ↓                                                          │
│  BuildAndRegister (自动注册)                               │
│    ├─ Intent → Action                                      │
│    ├─ Reducer → New State                                  │
│    ├─ Store.Update()                                       │
│    └─ ScheduleUpdate()                                     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 简化的用户代码

**重构前**（使用 UseState）：
```go
// 需要 5 步 + 类型断言
value, setValue := ui.UseStateString("")
ctx.GlobalState["field-setter"] = setValue
ui.WithInit(func() {
    ui.RegisterIntent(func(ctx *intent.ActionContext, i intent.FieldChangeIntent) intent.IntentResult {
        if fn, ok := ctx.GetState("field-setter"); ok {
            if setter, ok := fn.(func(string)); ok {  // 类型断言
                setter(i.Value)
            }
        }
        return intent.HandledResult()
    })
}, ...)
input.ForField(intent.BindField("field")).Value(value).Build()
```

**重构后**（使用 Store + Reducer）：
```go
// 只需要 3 步，无类型断言
var appStore *store.Store[AppState]

// 注册 handlers（一次性）
reducerBuilder := reducer.NewBuilder[AppState]().
    On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
        s.Field = i.(intent.FieldChangeIntent).Value
        return s
    }).RegisterToGlobal(appStore)

// 视图从 Store 读取
state := appStore.Get()
input.ForField(intent.BindField("field")).Value(state.Field).Build()
```

---

## 四、重构方案

### 方案 A：渐进式迁移（推荐）

1. **保留两套 API**
   - 保留 `UseState` API（向后兼容）
   - 继续完善 `Store + Reducer` API

2. **文档引导**
   - 在文档中明确说明 Store + Reducer 是推荐方式
   - 提供迁移指南

3. **示例更新**
   - store_reducer_demo 已经是正确实现
   - 将 focus_switching_demo 重写为 Store + Reducer 模式

### 方案 B：完全重构

1. **移除 UseState**
   - 需要改动大量现有代码
   - 破坏性大，风险高

2. **添加 RunApp 入口**
   - 创建新的 `ui.RunApp[T](AppRuntime[T])` 函数
   - 让用户选择使用新 API

---

## 五、实施计划

### Phase 1: 重写 focus_switching_demo

```go
// 文件: examples/fiber_firsts/focus_switching_demo/main.go

type AppState struct {
    Input1      string
    Input2      string
    ClickCount  int
    Checked1    bool
    Checked2    bool
}

var appStore *store.Store[AppState]

func main() {
    appStore = store.NewStore(AppState{
        Input1: "", Input2: "", ClickCount: 0, Checked1: false, Checked2: false,
    })

    // 注册 Reducer handlers
    reducer.NewBuilder[AppState]().
        On(intent.FieldChangeIntent{}, handleFieldChange).
        On(ClickButtonIntent{}, handleClick).
        RegisterToGlobal(appStore)

    err := ui.Run(App, ui.WithWidth(60), ui.WithHeight(35), ui.WithTitle("Focus Switching Demo"))
    if err != nil {
        panic(err)
    }
}

func App() ui.VNode {
    state := appStore.Get()
    // 构建 UI...
}
```

### Phase 2: 更新文档

1. 添加 `USING_STORE_REDUCER.md` - Store + Reducer 使用指南
2. 更新 `MVP_MIGRATION_GUIDE.md` - 说明 Store + Reducer 是 MVP 模式的完整实现
3. 更新 `README.md` - 推荐使用 Store + Reducer

### Phase 3: 可选：添加 RunApp 入口

```go
// 文件: ui/app.go

func RunApp[T any](config RuntimeConfig[T]) error {
    // 创建 AppRuntime
    rt := statemachine.NewAppRuntime(config.InitialState, config.View, config.Reducer)
    
    // 调用 ui.Run
    return ui.Run(func() ui.VNode {
        return rt.View().(ui.VNode)
    }, config.Options...)
}
```

---

## 六、风险评估

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|---------|
| 破坏现有应用 | 高 | 低 | 保留 UseState API，渐进式迁移 |
| Store 单例限制 | 中 | 中 | 可以创建多个 Store 实例 |
| 性能下降 | 低 | 低 | Store 订阅机制已优化 |

---

## 七、决策建议

### 推荐路径

1. **短期**（本次重构）：
   - 将 focus_switching_demo 重写为 Store + Reducer 模式
   - 更新文档，说明 Store + Reducer 是推荐方式

2. **中期**：
   - 提供更好的 Store + Reducer 文档和示例
   - 迁移更多的示例

3. **长期**：
   - 考虑添加 RunApp 入口
   - 逐步推广 Store + Reducer 最佳实践

### 关键原则

> **Store + Reducer 架构已经完整实现** ✓
> 
> 需要做的是：
> 1. **更新示例代码**，统一使用正确架构
> 2. **更新文档**，明确推荐路径
> 3. **可选**：添加新的高层 API

---

## 八、总结

### 关键发现

1. **Store + Reducer 架构已完整实现**：
   - `runtime/store/store.go` ✅
   - `runtime/reducer/reducer.go` ✅  
   - `runtime/statemachine/runtime.go` ✅
   - `buildAndRegister` 自动注册 ✅

2. **store_reducer_demo 是参考实现** ✅
   - 全局 Store
   - Reducer handlers 自动注册
   - 视图从 Store 读取

3. **focus_switching_demo 使用错误架构** ❌
   - UseState + GlobalState + Instance
   - 手动注册 handlers
   - 需要重构为 Store + Reducer

4. **不需要重新设计和实现 Store + Reducer**
   - 它们已经完整且正确
   - 需要的是：更新示例代码 + 更新文档

---

## 后续行动

1. ✅ 立即：重构 focus_switching_demo 为 Store + Reducer 模式
2. ✅ 同步：更新 focus_switching_demo 的 README
3. ✅ 同步：检查其他示例是否需要更新
4. ✅ 可选：添加 Store + Reducer 迁移指南
