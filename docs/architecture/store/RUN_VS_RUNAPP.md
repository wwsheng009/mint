# ui.Run vs ui.RunApp[T] 详细对比

Mint UI 提供了两种应用启动方式，各有不同的使用场景和优势。

---

## 快速对比表

| 特性 | ui.Run | ui.RunApp[T] |
|------|--------|--------------|
| **API 签名** | `Run(app ComponentFunc)` | `RunApp[T](rt *AppRuntime[T])` |
| **状态管理** | 手动（全局 Store） | 自动（AppRuntime 内置） |
| **重新渲染** | 需要 `ctx.ScheduleUpdate()` | 自动订阅状态变化 |
| **类型安全** | 运行时检查 | 编译时检查（泛型） |
| **代码量** | 较多（初始化代码） | 较少（AppRuntime 封装） |
| **依赖注入** | 需要使用全局变量 | 通过 AppRuntime 参数传递 |
| **时间旅行调试** | 需要手动实现 | 内置支持 |
| **适用场景** | 老项目、灵活性要求高 | 新项目、标准 Redux 模式 |
| **学习曲线** | 较低（React-like） | 较高（Redux-like） |

---

## 核心区别

### 1. API 设计理念

#### ui.Run - React-like 风格
```go
// 传统的 React-like 方式
// 组件自己管理状态或从全局 Store 读取

func App() ui.VNode {
    count := useInt(0)  // 使用 Hooks

    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.Button("Increment", func() {
            count++  // 修改状态
        }),
    )
}

func main() {
    ui.Run(App)
}
```

#### ui.RunApp[T] - Redux-like 风格
```go
// Redux-like 方式
// 状态集中管理，单向数据流

type AppState struct {
    Count int
}

func AppView(state AppState) any {
    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", state.Count)),
        ui.Button("Increment", IncrementIntent{}),
    )
}

func main() {
    rt := statemachine.NewAppRuntime(AppState{}, AppView, AppReducer)
    ui.RunApp(rt,
        ui.WithInit(func() {
            reducerBuilder.RegisterToGlobal(rt.GetStore())
        }),
    )
}
```

---

### 2. 状态管理方式

#### ui.Run - 全局 Store（手动管理）

```go
// 1. 定义全局 Store
var appStore *store.Store[AppState]

// 2. 初始化函数
func initStore() {
    appStore = store.NewStore(AppState{})
    appReducer.RegisterToGlobal(appStore)
}

// 3. 组件读取状态
func App() ui.VNode {
    state := appStore.Get()  // ⚠️ 手动读取
    return renderUI(state)
}

// 4. 启动应用
func main() {
    ui.Run(App,
        ui.WithInit(initStore),  // ⚠️ 需要手动初始化
    )
}
```

**特点**:
- ❌ 需要手动创建全局变量
- ❌ 组件需要自己读取状态
- ❌ 初始化代码分离
- ✅ 可以在多个组件间共享 Store
- ✅ 更灵活的控制

#### ui.RunApp[T] - AppRuntime（自动管理）

```go
type AppState struct {
    Count int
}

func AppView(state AppState) any {
    return renderUI(state)  // ✅ 状态作为参数传入
}

func main() {
    rt := statemachine.NewAppRuntime(AppState{}, AppView, AppReducer)
    ui.RunApp(rt,
        ui.WithInit(func() {
            reducerBuilder.RegisterToGlobal(rt.GetStore())
        }),
    )
}
```

**特点**:
- ✅ 无需全局变量
- ✅ 状态自动注入到 View 函数
- ✅ 初始化代码集中
- ❌ 需要维护 AppRuntime
- ❌ 类型参数可能增加复杂度

---

### 3. 重新渲染机制

#### ui.Run - 需要手动触发

```go
// Reducer 中必须调用 ScheduleUpdate
var appReducer = reducer.NewBuilder[AppState]().
    On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
        s.Count++
        return s
    }).
    BuildAndRegister(registry, store)  // ✅ BuildAndRegister 自动调用 ScheduleUpdate
```

#### ui.RunApp[T] - 自动触发

```go
// ui.RunApp 内部自动订阅状态变化
rt := statemachine.NewAppRuntime(AppState{}, AppView, AppReducer)
ui.RunApp(rt)  // ✅ 自动订阅，状态变化时重新渲染
```

**内部实现**:
```go
// ui.RunApp 的简化实现
rt.OnStateChange(func(newState T) {
    // 触发 Fiber 重新渲染
    declarativeRoot.GetReconciler().ScheduleUpdate()
})
```

---

### 4. 代码结构对比

#### ui.Run 典型项目结构

```
├── main.go
├── store/
│   ├── store.go          // 全局 Store 定义
│   └── reducers.go       // Reducer 定义
├── components/
│   ├── app.go            // 主组件
│   ├── counter.go        // 计数器组件
│   └── form.go           // 表单组件
└── utils/
    └── helpers.go
```

#### ui.RunApp[T] 典型项目结构

```
├── main.go
├── state/
│   ├── app_state.go      // AppState 定义
│   ├── app_view.go       // AppView 函数
│   └── app_reducer.go    // AppReducer 定义
└── runtime/
    └── init.go           // AppRuntime 初始化
```

---

## 详细对比

### 优点对比

#### ui.Run 的优点
1. ✅ **灵活性高** - 适合各种状态管理策略
2. ✅ **React-like** - 如果熟悉 React Hooks，容易上手
3. ✅ **全局共享** - 多个组件可以共享一个 Store
4. ✅ **渐进迁移** - 可以逐步从 UseState 迁移
5. ✅ **无泛型约束** - API 更简单

#### ui.RunApp[T] 的优点
1. ✅ **类型安全** - 泛型提供编译时类型检查
2. ✅ **自动封装** - AppRuntime 封装了所有状态管理
3. ✅ **Redux 模式** - 符合 Redux 最佳实践
4. ✅ **时间旅行** - 内置历史记录和撤销功能
5. ✅ **简化初始化** - 启动代码更简洁

### 缺点对比

#### ui.Run 的缺点
1. ❌ **手动管理** - 需要手动初始化和读取状态
2. ❌ **全局变量** - 依赖全局变量，测试困难
3. ❌ **手动触发** - 需要确保 Reducer 调用 ScheduleUpdate

#### ui.RunApp[T] 的缺点
1. ❌ **需要注册** - 必须使用 RegisterToGlobal 注册 handlers
2. ❌ **类型参数** - 泛型增加 API 复杂度
3. ❌ **视图约束** - AppView 必须返回 `any` 类型

---

## 使用场景

### 选择 ui.Run 的情况

1. **从 React 迁移的项目**
   - 熟悉 React Hooks 模式
   - 需要保持原有代码结构

2. **需要多个独立 Store 的应用**
   ```go
   var userStore *store.Store[UserState]
   var uiStore *store.Store[UIState]
   var dataStore *store.Store[DataState]

   func App() ui.VNode {
       user := userStore.Get()
       ui := uiStore.Get()
       // ...
   }
   ```

3. **需要灵活控制初始化顺序**
   ```go
   func main() {
       initDatabase()
       initAuthService()
       initStore()      // 按特定顺序初始化
       ui.Run(App)
   }
   ```

4. **使用 UseState 的旧项目**
   - 正在逐步迁移到 Store + Reducer
   - 需要新旧代码共存

### 选择 ui.RunApp[T] 的情况

1. **全新项目**
   - 无历史包袱
   - 采用最佳实践

2. **需要时间旅行调试**
   ```go
   rt := statemachine.NewAppRuntime(
       AppState{},
       AppView,
       AppReducer,
       statemachine.WithMaxHistory(100),
   )

   // 可以跳转到任意历史状态
   rt.JumpTo(50)
   rt.Undo()
   ```

3. **需要类型安全的状态管理**
   ```go
   // 编译期检查类型错误
   type AppState struct {
       Username string
       Count int
   }

   func AppView(state AppState) any {
       // ✅ 编译时检查，避免拼写错误
       fmt.Println(state.Username)
   }
   ```

4. **简单应用**
   - 状态集中在一个 Store
   - 不需要复杂的状态管理策略

---

## 性能对比

| 指标 | ui.Run | ui.RunApp[T] | 说明 |
|------|--------|--------------|------|
| **编译速度** | ⚡ 快 | 🐢 稍慢 | 泛型增加编译时 |
| **运行时性能** | ⚡ 快 | ⚡ 快 | 无显著差异 |
| **内存占用** | 低 | 低 | 基本相同 |
| **首次渲染** | 快 | 快 | 无显著差异 |
| **重新渲染** | 依赖实现 | 优化 | RunApp 自动优化 |

---

## 代码量对比

### 示例：简单的计数器

#### ui.Run 版本
```go
var countStore *store.Store[int]

func initStore() {
    countStore = store.NewStore(0)
    reducer.NewBuilder[int]().
        On(IncrementIntent{}, func(s int, i intent.Intent) int {
            return s + 1
        }).
        RegisterToGlobal(countStore)
}

func App() ui.VNode {
    count := countStore.Get()
    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.Button("Increment", IncrementIntent{}),
    )
}

func main() {
    ui.Run(App, ui.WithInit(initStore))
}
```
**代码行数**: ~25 行

#### ui.RunApp[T] 版本
```go
type AppState struct {
    Count int
}

func AppView(state AppState) any {
    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", state.Count)),
        ui.Button("Increment", IncrementIntent{}),
    )
}

var appReducer = reducer.NewBuilder[AppState]().
    On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
        s.Count++
        return s
    })

func main() {
    rt := statemachine.NewAppRuntime(AppState{}, AppView, appReducer.Build())
    ui.RunApp(rt,
        ui.WithInit(func() {
            appReducer.RegisterToGlobal(rt.GetStore())
        }),
    )
}
```
**代码行数**: ~22 行

**结论**: 代码量基本相同

---

## 推荐

### 原则

1. **新项目 → ui.RunApp[T]**
   - 采用最佳实践
   - 更好的类型安全
   - 内置时间旅行调试

2. **从 UseState 迁移 → ui.Run**
   - 更平滑的迁移路径
   - 可以逐步引入 Store + Reducer

3. **灵活需求 → ui.Run**
   - 多个独立 Store
   - 复杂的初始化逻辑
   - 需要与遗留代码集成

### 决策流程

```
开始
  │
  ├─ 是否新项目？
  │   ├─ 是 → ui.RunApp[T]
  │   └─ 否 → 继续
  │
  ├─ 是否需要多个独立 Store？
  │   ├─ 是 → ui.Run
  │   └─ 否 → 继续
  │
  ├─ 是否需要时间旅行调试？
  │   ├─ 是 → ui.RunApp[T]
  │   └─ 否 → 继续
  │
  ├─ 是否正在从 UseState 迁移？
  │   ├─ 是 → ui.Run
  │   └─ 否 → 继续
  │
  └─ → ui.RunApp[T] (默认选择)
```

---

## 总结

| 方面 | ui.Run | ui.RunApp[T] |
|------|--------|--------------|
| **设计哲学** | React-like | Redux-like |
| **状态管理** | 手动 | 自动 |
| **类型安全** | 运行时 | 编译时 |
| **学习曲线** | 低 | 中 |
| **代码量** | 相同 | 相同 |
| **推荐场景** | 老项目、灵活需求 | 新项目、标准模式 |

**最终建议**:
- **新项目**: 使用 `ui.RunApp[T]`
- **老项目**: 根据需求选择，或逐步迁移到 `ui.RunApp[T]`
- **复杂项目**: 可以混合使用两者

---

## 参考资料

- **RunApp 指南**: `docs/architecture/store/RUNAPP_GUIDE.md`
- **Store + Reducer 指南**: `docs/architecture/store/STORE_REDUCER_GUIDE.md`
- **实现总结**: `docs/architecture/store/IMPLEMENTATION_SUMMARY.md`
- **示例代码**:
  - `examples/runapp_demo/main.go` - RunApp 示例
  - `examples/store_reducer_demo/main.go` - Store + Reducer 示例
