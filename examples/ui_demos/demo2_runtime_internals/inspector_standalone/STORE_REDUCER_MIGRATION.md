# Store + Reducer 架构迁移总结

## 文件
`E:\projects\yao\wwsheng009\mint\examples\ui_demos\demo2_runtime_internals\inspector_standalone\main.go`

## 迁移时间
2026年3月8日

## 架构变更

### 1. 定义 AppState 结构（单一数据源）
```go
// AppState represents the demo state.
type AppState struct {
	CurrentPhase  string
	EventCount    int
	RenderCount   int
	BufferUpdates int
}
```

### 2. 定义自定义 Intent 类型
```go
type SetPhaseIntent struct {
	Phase string
}

type SetPhaseWithRenderCountIntent struct {
	Phase string
}

type SetPhaseWithBufferCountIntent struct {
	Phase string
}

type SetPhaseWithEventCountIntent struct {
	Phase string
}

type ToggleInspectorIntent struct{}
```

### 3. 创建 appReducer 纯函数
```go
var appReducer = reducer.NewBuilder[AppState]()

func init() {
	appReducer.On(SetPhaseIntent{}, func(s AppState, i intent.Intent) AppState {
		spi := i.(SetPhaseIntent)
		s.CurrentPhase = spi.Phase
		return s
	})

	appReducer.On(SetPhaseWithEventCountIntent{}, func(s AppState, i intent.Intent) AppState {
		pi := i.(SetPhaseWithEventCountIntent)
		s.CurrentPhase = pi.Phase
		s.EventCount++
		return s
	})

	appReducer.On(SetPhaseWithRenderCountIntent{}, func(s AppState, i intent.Intent) AppState {
		pi := i.(SetPhaseWithRenderCountIntent)
		s.CurrentPhase = pi.Phase
		s.RenderCount++
		return s
	})

	appReducer.On(SetPhaseWithBufferCountIntent{}, func(s AppState, i intent.Intent) AppState {
		pi := i.(SetPhaseWithBufferCountIntent)
		s.CurrentPhase = pi.Phase
		s.BufferUpdates++
		return s
	})

	appReducer.On(ToggleInspectorIntent{}, func(s AppState, i intent.Intent) AppState {
		globalInspector.ToggleVisibility()
		return s
	})
}
```

### 4. 创建 appStore 作为单一状态源
```go
var appStore = store.NewStore(AppState{
	CurrentPhase:  "idle",
	EventCount:    0,
	RenderCount:   0,
	BufferUpdates: 0,
})
```

### 5. 在 main 函数中注册 reducer 到全局
```go
// Register reducer handlers to store
appReducer.RegisterToGlobal(appStore)
```

### 6. 替换 UseState 为 Store.Get()
**之前：**
```go
func RuntimeDemoStandalone() ui.VNode {
	currentPhase, setCurrentPhase := ui.UseStateString("idle")
	eventCount, setEventCount, _ := ui.UseStateInt(0)
	renderCount, setRenderCount, _ := ui.UseStateInt(0)
	bufferUpdates, setBufferUpdates, _ := ui.UseStateInt(0)
	// ...
}
```

**之后：**
```go
func RuntimeDemoStandalone() ui.VNode {
	state := appStore.Get()
	// ...
}
```

### 7. 移除 ui.On() 闭包监听器
**之前：**
```go
func ControlPanel(...) ui.VNode {
	ui.On(InspectorActionIntent{Action: "event"}, func(actx *intent.ActionContext) {
		if setCurrentPhase != nil {
			setCurrentPhase("Event")
		}
		if setEventCount != nil {
			setEventCount(func(c int) int { return c + 1 })
		}
	})
	// ...
}
```

**之后：**
不需要 ui.On() 监听器，因为 reducer 已全局注册。

### 8. 将按钮 OnPress 改为直接使用 Intent 类型
**之前：**
```go
ui.NewButtonBuilder("[1] Event").
	Variant(ui.ButtonVariantDanger).
	OnPress(InspectorActionIntent{Action: "event"}).
	FocusStyle(ui.FocusStyleBracket).
	Build()
```

**之后：**
```go
ui.NewButtonBuilder("[1] Event").
	Variant(ui.ButtonVariantDanger).
	OnPress(SetPhaseWithEventCountIntent{Phase: "Event"}).
	FocusStyle(ui.FocusStyleBracket).
	Build()
```

### 9. 简化函数签名
**之前：**
```go
func buildDemoContent(
	currentPhase string,
	eventCount, renderCount, bufferUpdates int,
	setCurrentPhase func(string),
	setEventCount, setRenderCount, setBufferUpdates func(interface{}),
) ui.VNode {
	// ...
}
```

**之后：**
```go
func buildDemoContent(state AppState) ui.VNode {
	// ...
}
```

### 10. 调用更新
**之前：**
```go
buildDemoContent(
	currentPhase, eventCount, renderCount, bufferUpdates,
	setCurrentPhase, setEventCount, setRenderCount, setBufferUpdates,
)
```

**之后：**
```go
buildDemoContent(state)
```

## 迁移验证

- ✅ 编译成功（无错误）
- ✅ 保持原有功能不变
- ✅ 遵循 Store + Reducer + Custom Intent 架构模式
- ✅ 与参考文件（inspector_demo/main.go, inspector_overlay/main.go）架构一致

## 架构优势

1. **单一数据源**：所有状态集中在 Store 中管理
2. **类型安全**：自定义 Intent 类型提供编译时检查
3. **可预测性**：Reducer 纯函数确保状态转换可预测
4. **可测试性**：Reducer 函数易于单元测试
5. **关注点分离**：状态逻辑与 UI 组件解耦
6. **集中管理**：所有状态转换在 reducer 中统一处理
