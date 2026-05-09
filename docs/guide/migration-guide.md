# 迁移指南

本文面向从旧 Mint API、旧 VNode 状态模型或 closure callback 模型迁移到当前源码结构的代码。当前推荐路径是：

```text
ui.Run / ui.RunApp
  -> ui package shortcuts/builders
  -> ui/components/<name> VNode + Instance
  -> runtime/intent + runtime/action
  -> runtime/store / reducer for application state
```

## 当前迁移原则

1. 组件创建优先使用 `ui` 包。
2. 组件实现放在 `ui/components/<component>/`，采用 `VNode + Instance + Builder`。
3. 运行期状态属于 `ComponentInstance`、store 或 reducer，不应放在 VNode 字段里。
4. 用户交互优先使用 `intent.Intent`、`runtime/action`、store/reducer，不推荐在 VNode 中保存闭包。
5. `UseState*`、`UseHoverState` 等 Hooks 仍存在，但源码注释已将部分 Hook 标为兼容/过渡能力；新应用状态优先使用 store/reducer。

## API 对照

| 旧写法 | 当前写法 |
|---|---|
| `app.Text("hello")` | `ui.Text("hello")` |
| `app.NewTextBuilder("hello")` | `ui.NewTextBuilder("hello")` |
| `app.ButtonBuilder("Save")` | `ui.NewButtonBuilder("Save")` |
| `app.InputBuilder()` | `ui.NewInputBuilder()` |
| `ui.Button("Save", fn)` | `ui.NewButtonBuilder("Save").OnPress(intent).Build()` |
| `button.OnClick(func(){ ... })` | `button.OnPress(intent.SomeIntent(...))` |
| VNode 字段保存 focus/hover/pressed | `ComponentInstance` + behavior/state |
| 组件外手写鼠标命中 | HitMap / `MouseMsg.TargetFiber` / `TargetBounds` |

`app.Run` / `app.Quit` 作为兼容入口仍存在，但组件 API 不应再从 `app` 包推导。

## Button 迁移

旧闭包写法：

```go
// 旧：不要在新文档和新代码中继续推荐
button.OnClick(func() {
    save()
})
```

当前 intent 写法：

```go
type SaveIntent struct{}

func (SaveIntent) IntentType() string { return "Save" }

func SaveButton() ui.VNode {
    return ui.NewButtonBuilder("Save").
        OnPress(SaveIntent{}).
        Build()
}
```

在应用启动或模块初始化阶段注册 handler：

```go
ui.RegisterIntent(func(ctx *intent.ActionContext, i SaveIntent) intent.IntentResult {
    // 更新 store、调用 service 或触发 reducer
    return intent.HandledResult()
})
```

对于通用动作，也可以使用内置 intent：

```go
ui.NewButtonBuilder("Open Settings").
    OnPress(intent.OpenModal("settings")).
    Build()
```

## 状态迁移

### 旧模式：局部变量或 VNode 字段

```go
func Counter() ui.VNode {
    count := 0
    return ui.Text(fmt.Sprintf("Count: %d", count))
}
```

这类状态会随着 render 重建而丢失。

### 过渡模式：Hooks

```go
func Counter() ui.VNode {
    count, setCount, _ := ui.UseStateInt(0)

    ui.On(ui.SimpleIncrementIntent{}, func(ctx *intent.ActionContext) {
        setCount(func(old int) int { return old + 1 })
    })

    return ui.VStack(
        ui.Text(fmt.Sprintf("Count: %d", count)),
        ui.NewButtonBuilder("+").
            OnPress(ui.SimpleIncrementIntent{}).
            Build(),
    )
}
```

Hooks 适合小范围兼容和局部状态，但新业务状态建议迁移到 store/reducer。

### 推荐模式：Store / Reducer / Intent

推荐将应用状态放入 `runtime/store`，用 intent 或 reducer 更新。详见：

- [../ui/store/README.md](../ui/store/README.md)
- [component-development-guide.md](component-development-guide.md)

典型结构：

```text
Component emits Intent
  -> handler / reducer
  -> Store.Set(...)
  -> subscribed components rerender
```

## 组件实现迁移

每个新组件应尽量采用当前组件目录模式：

```text
ui/components/<name>/
  vnode.go       # 纯声明 props
  instance.go    # 运行期状态、Measure、Paint、Action/Focus
  builder.go     # fluent API
  <name>_test.go
  README.md
```

VNode 要保持轻量：

- 保存 props。
- 不保存 hover/focus/pressed 等运行期状态。
- 不保存不可序列化的业务闭包，除非该组件已有明确兼容理由。
- 通过 `CreateInstance()` 创建 `ComponentInstance`。

Instance 负责：

- `SetProps` 同步声明属性。
- 持有运行期状态。
- 可选实现 `rtui.PaintableInstance`。
- 可选实现 `rtui.FocusableInstance`。
- 可选实现 `rtui.ActionHandlerInstance`。
- 可选实现 `Measure(layout.Constraints)`。

## 事件迁移

旧事件分发常以 VNode `HandleEvent` 或组件私有 `ContainsPoint` 为中心。当前主路径是：

```text
framework/event.Pump
  -> runtime/msg.KeyMsg / MouseMsg
  -> runtime/action.InputProcessor
  -> framework.App.processMsg
  -> FiberFocusManager / ActionBridge
```

鼠标事件应依赖：

- `MouseMsg.TargetID`
- `MouseMsg.TargetFiber`
- `MouseMsg.TargetBounds`
- `MouseMsg.LocalX` / `LocalY`

相关文档：

- [../howto/migrate-to-targetbounds.md](../howto/migrate-to-targetbounds.md)
- [../event/long_term_event_architecture.md](../event/long_term_event_architecture.md)

## 测试迁移

旧测试可以逐步迁移到：

```go
app, err := ui.RunTestWithSandbox(App, ui.WithSize(80, 24))
if err != nil {
    t.Fatal(err)
}
```

Sandbox 组件兼容性见：

- [../sandbox/SANDBOX_COMPONENTS_COMPATIBILITY.md](../sandbox/SANDBOX_COMPONENTS_COMPATIBILITY.md)
- [../testing/e2e/README.md](../testing/e2e/README.md)

建议分层执行：

```bash
go test ./ui/components/<component> -count=1
go test ./ui/e2e -run <ComponentOrFeature> -count=1
go test ./examples/sandbox/... -count=1
```

## 迁移检查清单

- 组件 API 是否使用 `ui.NewXBuilder` / `ui.X(...)`。
- Button 是否使用 `OnPress(intent)`，而不是 `OnClick(func())`。
- 运行期状态是否移出 VNode。
- 可聚焦组件是否实现 `rtui.FocusableInstance`。
- 鼠标命中是否依赖 HitMap / TargetBounds。
- 应用状态是否进入 store/reducer，或至少明确标注 Hooks 只是过渡。
- 文档中的源码路径是否指向 `ui/components/<name>/`、`framework/app.go`、`runtime/msg`、`runtime/action` 等当前文件。
