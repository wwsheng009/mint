# Store Hybrid 状态管理指南

本文是混合状态管理模式的兼容说明。当前项目推荐优先阅读并使用上一级文档：

- [../README.md](../README.md)

旧版本文档曾在示例中使用 `OnPress(func(){...})`、`OnClick(func(){...})`。当前 Button API 的主路径是 `OnPress(intent.Intent)`，因此不要再把闭包回调写作当前推荐用法。

## 当前推荐顺序

| 场景 | 推荐方案 |
|---|---|
| 应用级状态 | `runtime/store` + reducer + intent |
| 多组件共享字段 | `UseStoreSelector` / `UseStoreField` |
| 简单局部过渡状态 | Hooks，作为兼容/过渡能力使用 |
| 按钮、选择、表单提交 | `intent.Intent` + handler/reducer |

## 混合模式是什么

混合模式指的是：

```text
Component
  -> UseStoreSelector / UseStoreField 订阅状态
  -> OnPress(intent) 发出意图
  -> intent handler / reducer 更新 Store
  -> Store 通知组件重新 render
```

它适合从旧 Hooks 或 GlobalState 迁移到 Store/Reducer 的中间阶段。

## 当前示例

```go
type CounterState struct {
    Count int
}

type IncrementIntent struct{}

func (IncrementIntent) IntentType() string { return "Increment" }
```

组件订阅 store 并发出 intent：

```go
func CounterView(counterStore *store.Store[CounterState]) ui.VNode {
    count := ui.UseStoreSelector(counterStore, func(s CounterState) int {
        return s.Count
    })

    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.NewButtonBuilder("加1").
            OnPress(IncrementIntent{}).
            Build(),
    )
}
```

在应用初始化处注册 handler：

```go
ui.RegisterIntent(func(ctx *intent.ActionContext, i IncrementIntent) intent.IntentResult {
    current := counterStore.Get()
    current.Count++
    counterStore.Set(current)
    return intent.HandledResult()
})
```

> 以上示例展示形态，实际代码应根据当前 store 类型、生命周期和 handler 注册位置组织。

## UseStoreSelector

`UseStoreSelector` 用于订阅派生值：

```go
name := ui.UseStoreSelector(appStore, func(s AppState) string {
    return s.Profile.Name
})
```

只有 selector 结果变化时，组件才需要更新，适合从全局状态中取局部视图数据。

## UseStoreField

`UseStoreField` 用于读写单一字段。它适合简单表单字段迁移，但复杂业务仍建议通过 reducer 处理。

```go
value, setValue := ui.UseStoreField(
    appStore,
    func(s AppState) string { return s.Name },
    func(s *AppState, value string) { s.Name = value },
)
```

如果更新依赖当前值，可使用 functional variant：

```go
count, setCount := ui.UseStoreFieldFunctional(
    appStore,
    func(s AppState) int { return s.Count },
    func(s *AppState, value int) { s.Count = value },
)

setCount(func(old int) int {
    return old + 1
})
```

## Hooks 的位置

`UseStateInt`、`UseStateString`、`UseStateBool` 仍可用于局部状态和旧代码迁移，但源码注释已将部分 hook 标记为 deprecated / transitional。新业务状态优先用 Store/Reducer。

```go
count, setCount, _ := ui.UseStateInt(0)
setCount(func(old int) int { return old + 1 })
```

当状态需要跨组件共享、参与测试回放、由 intent 统一驱动，或要避免闭包捕获旧值时，应迁移到 Store/Reducer。

## 迁移检查

- 是否还存在 `OnClick(func(){...})`。
- 是否还存在 `OnPress(func(){...})`。
- Button 是否改为 `OnPress(intent)`。
- 状态更新是否进入 store/reducer 或明确保持为局部 Hook。
- selector 是否只返回组件实际需要的数据。
- handler 注册是否避免重复注册和闭包旧值问题。

## 测试建议

```bash
go test ./ui -run Store -count=1
go test ./runtime/store ./runtime/reducer ./runtime/intent -count=1
go test ./ui/e2e -run "Input|Checkbox|List|Tabs|Slider" -count=1
```

更多当前说明见 [../README.md](../README.md)。
