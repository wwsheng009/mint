# ui.RunApp[T] 使用指南

`ui.RunApp[T]` 是 Mint UI 的 Store + Reducer 架构的推荐入口点，提供了更简洁的 API 和自动化的状态管理。

---

## 快速开始

### 最小示例

```go
package main

import (
	"fmt"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/statemachine"
	"github.com/wwsheng009/mint/ui"
)

// 1. 定义应用状态
type AppState struct {
	Count int
}

// 2. 定义 Intent
type IncrementIntent struct{}

func (IncrementIntent) IntentType() string { return "Increment" }

// 3. 定义纯函数视图
func AppView(state AppState) any {
	return ui.VStack(
		ui.Text(fmt.Sprintf("Count: %d", state.Count)),
		ui.NewButtonBuilder("+").OnPress(IncrementIntent{}).Build(),
	)
}

// 4. 构建 Reducer Builder
var appReducerBuilder = reducer.NewBuilder[AppState]().
	On(IncrementIntent{}, func(s AppState, i intent.Intent) AppState {
		s.Count++
		return s
	})

// 5. 主函数
func main() {
	rt := statemachine.NewAppRuntime(
		AppState{},
		AppView,
		appReducerBuilder.Build(),
	)

	ui.RunApp(rt,
		ui.WithInit(func() {
			// ⚠️ 必须注册 Intent handlers！
			appReducerBuilder.RegisterToGlobal(rt.GetStore())
		}),
	)
}
```

---

## 重要：Intent Handlers 注册

### 问题：为什么需要注册？

`ui.RunApp[T]` 创建了自己的 Intent Runtime，但不会自动注册 Reducer 的 handlers。如果你不注册，点击按钮时会看到错误：

```
[Intent] No handler for intent type=Increment
```

### 解决方案 1：使用 RegisterToGlobal（推荐）

```go
func main() {
	rt := statemachine.NewAppRuntime(AppState{}, AppView, appReducerBuilder.Build())

	ui.RunApp(rt,
		ui.WithInit(func() {
			// 连接 Reducer 和 AppRuntime 的 Store
			appReducerBuilder.RegisterToGlobal(rt.GetStore())
		}),
	)
}
```

**优点**：
- ✅ 简洁，一行代码
- ✅ 自动处理 Dispatcher → Reducer → Store → UI 流程
- ✅ 支持 `ctx.ScheduleUpdate()` 自动重新渲染

### 解决方案 2：手动注册 Handlers（高级）

如果你需要更多控制，可以手动注册 handlers：

```go
ui.RunApp(rt,
	ui.WithInit(func() {
		intent.RegisterTyped(func(ctx *intent.ActionContext, i IncrementIntent) intent.IntentResult {
			// 手动调用 Reducer
			currentState := rt.GetState()
			newState := appReducerBuilder.Reduce(currentState, i)
			rt.Dispatch(newState) // 或者直接: rt.GetStore().Set(newState)

			// 手动触发重新渲染
			ctx.ScheduleUpdate()

			return intent.HandledResult()
		})
	}),
)
```

**何时使用**：
- 需要自定义 Intent 处理逻辑
- 需要添加副作用（如日志、API 调用）
- 需要条件性处理

---

## 完整架构流程

```
用户点击按钮
    ↓
按钮发出 Intent (IncrementIntent)
    ↓
Intent Runtime 查找 Handler
    ↓
Handler 调用 Reducer (通过 RegisterToGlobal 注册)
    ↓
Reducer 计算新状态 (纯函数)
    ↓
Store 更新状态 (Store.Set)
    ↓
AppRuntime 订阅到状态变化
    ↓
触发 Fiber 重新渲染 (自动)
```

---

## 对比：两种使用模式

### 模式 A: ui.Run + 全局 Store（传统）

```go
var appStore *store.Store[AppState]

func initStore() {
	appStore = store.NewStore(AppState{})
	appReducer.RegisterToGlobal(appStore)
}

func App() ui.VNode {
	state := appStore.Get()
	return renderUI(state)
}

func main() {
	ui.Run(App,
		ui.WithInit(initStore),
	)
}
```

**适用场景**：
- 需要跨多个组件共享全局 Store
- 需要完全控制初始化过程
- 从 UseState 迁移时的中间步骤

### 模式 B: ui.RunApp + AppRuntime（推荐）

```go
func AppView(state AppState) any {
	return renderUI(state)
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

**适用场景**：
- 新项目或已使用 AppRuntime 的项目
- 依赖自动重新渲染
- 需要时间旅行调试

---

## 常见错误

### 错误 1: No handler for intent type

**错误信息**：
```
[Intent] No handler for intent type=IncrementIntent: no handler registered
```

**原因**：
忘记注册 Intent handlers。

**解决**：
```go
ui.RunApp(rt,
	ui.WithInit(func() {
		reducerBuilder.RegisterToGlobal(rt.GetStore()) // ← 添加这行
	}),
)
```

### 错误 2: Type mismatch

**错误信息**：
```
cannot use AppView (type AppState -> ui.VNode) as type AppState -> any
```

**原因**：
`AppView` 的返回类型必须是 `any`，而不是 `ui.VNode`。

**解决**：
```go
// ❌ 错误
func AppView(state AppState) ui.VNode { ... }

// ✅ 正确
func AppView(state AppState) any {
	return renderAppView(state) // 内部函数返回 ui.VNode
}
```

### 错误 3: Store not exposed

**错误**：
```go
appReducerBuilder.RegisterToGlobal(rt.GetStore())
// undefined: rt.GetStore
```

**原因**：
使用旧版本的 Mint UI，`AppRuntime` 没有 `GetStore()` 方法。

**解决**：
更新到最新版本，或使用全局 Store 模式。

---

## 高级用法

### 1. 时间旅行调试

```go
type UndoIntent struct{}

func (UndoIntent) IntentType() string { return "Undo" }

rt := statemachine.NewAppRuntime(
	AppState{},
	AppView,
	AppReducer,
	statemachine.WithMaxHistory(100), // 记录最近 100 个状态
)

ui.RunApp(rt,
	ui.WithInit(func() {
		intent.RegisterTypedGlobally(func(ctx *intent.ActionContext, i UndoIntent) intent.IntentResult {
			_ = rt.Undo()
			return intent.HandledResult()
		})
	}),
)
```

在 UI 中提供撤销按钮时，按钮仍然只声明 Intent：

```go
ui.NewButtonBuilder("Undo").OnPress(UndoIntent{}).Build()
```

### 2. 调试 Intent 流程

使用 `LoggingMiddleware` 记录所有 Intent 处理：

```go
var AppReducer = reducer.WithMiddleware(
	reducerBuilder.Build(),
	reducer.LoggingMiddleware(func(state AppState, i intent.Intent, newState AppState) {
		fmt.Printf("Intent: %v, State: %v -> %v\n", i, state, newState)
	}),
)
```

### 3. 多个 Reducer 组合

```go
counterReducer := reducer.NewBuilder[AppState]().On(IncrementIntent{...}).Build()
formReducer := reducer.NewBuilder[AppState]().On(SubmitFormIntent{...}).Build()

appReducer := reducer.Compose(counterReducer, formReducer)
```

---

## 性能优化

### 1. 使用 Selector 避免不必要的渲染

```go
type AppState struct {
	Count int
	HeavyData []int // 大数据集
}

// ❌ 每次渲染都会访问 HeavyData
func AppView(state AppState) any {
	return ui.VStack(
		ui.Text(strconv.Itoa(state.Count)),
		computeFromHeavyData(state.HeavyData), // 每次都计算
	)
}

// ✅ 使用 Selector 缓存结果
var countDataSelector = store.NewComputed(
	func(state AppState) string {
		return computeFromHeavyData(state.HeavyData)
	},
)

func AppView(state AppState) any {
	// 只在 Count 变化时重新渲染
	return ui.VStack(
		ui.Text(strconv.Itoa(state.Count)),
		ui.Text(countDataSelector.Get()),
	)
}
```

### 2. Lane 调度优化

```go
ui.RunApp(rt,
	ui.WithLaneScheduler(),
	ui.WithDefaultLane(2), // Input lane，高优先级
)
```

---

## 迁移指南

### 从 ui.Run 迁移到 ui.RunApp

**前 (ui.Run)**：
```go
var appStore *store.Store[AppState]

func initStore() {
	appStore = store.NewStore(AppState{})
	appReducer.RegisterToGlobal(appStore)
}

func App() ui.VNode {
	state := appStore.Get()
	return renderUI(state)
}

func main() {
	ui.Run(App,
		ui.WithInit(initStore),
	)
}
```

**后 (ui.RunApp)**：
```go
func AppView(state AppState) any {
	return renderUI(state)
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

## 参考资料

- **API 文档**: `docs/ui/store/api/API_REFERENCE.md`
- **完整示例**: `examples/runapp_demo/main.go`
- **Store + Reducer 指南**: `docs/ui/store/guides/STORE_REDUCER_GUIDE.md`
- **迁移指南**: `docs/ui/store/guides/MIGRATION_GUIDE.md`

---

**🎯 总结**：`ui.RunApp[T]` 是最现代的开发方式，但不要忘记使用 `ui.WithInit` 和 `RegisterToGlobal` 注册 Intent handlers！
