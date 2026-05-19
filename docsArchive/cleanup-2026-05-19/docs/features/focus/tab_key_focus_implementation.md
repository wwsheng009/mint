# Tab 键焦点切换实现

本文记录 Mint 当前源码中的 Tab / Shift+Tab 焦点导航路径。早期文档曾把 `runtime/ui.VNodeFocusManager` 和 `DeclarativeNode.HandleEvent()` 写作主路径；这些代码仍有兼容价值，但当前应用主路径应以 `runtime/msg -> runtime/action -> framework.App -> FiberFocusManager` 为准。

## 当前主路径

```text
terminal input
  -> framework/event.Pump
  -> runtime/msg.KeyMsg
  -> framework.App.processMsg
  -> runtime/action.InputProcessor
  -> ActionNavigateNext / ActionNavigatePrev
  -> framework.App.handleNavigationAction
  -> runtime/ui.FiberFocusManager.FocusNext / FocusPrev
  -> focused Fiber.Instance.SetFocus(bool)
```

## 关键源码

| 功能 | 当前源码 |
|---|---|
| KeyMsg 到 Action 的转换 | `runtime/action/processor.go` |
| 默认 Tab keymap | `runtime/action/keymap.go` |
| App 消息入口 | `framework/app.go` 的 `processMsg` |
| 导航 action 处理 | `framework/app.go` 的 `handleNavigationAction` |
| Fiber 焦点管理 | `runtime/ui/fiber_focus_manager.go` |
| Fiber 焦点收集 | `runtime/ui/fiber_focus_manager.go` 的 `CollectFromFiber` |
| 焦点状态写入 | `runtime/ui/fiber_focus_manager.go` 的 `applyFocusState` |

## Action 映射

`runtime/action.InputProcessor` 会把 Tab 转换为导航 action：

```go
case runtimeplatform.KeyTab:
    if keyMsg.Mod.Shift {
        return NewActionFromKey(ActionNavigatePrev, "keyboard")
    }
    return NewActionFromKey(ActionNavigateNext, "keyboard")
```

`framework.App.handleNavigationAction` 再把 action 交给焦点管理器：

```go
switch act.Type {
case action.ActionNavigateNext:
    handled = a.focusManager.FocusNext()
case action.ActionNavigatePrev:
    handled = a.focusManager.FocusPrev()
case action.ActionNavigateHome:
    handled = a.focusManager.FocusFirst()
case action.ActionNavigateEnd:
    handled = a.focusManager.FocusLast()
}
```

方向键当前不会被 `FiberFocusManager` 当作全局焦点移动处理；在文本输入组件获得焦点时，App 会把部分导航 action 重映射为 cursor action。

## FiberFocusManager 行为

当前焦点列表来自 Fiber 树，而不是 VNode 树：

```text
FiberFocusManager.CollectFromFiber(root)
  -> collectFocusableFibers
  -> 检查 fiber.Instance 是否实现 FocusableInstance
  -> 排除 disabled 实例
```

焦点切换时，管理器会调用组件运行期实例：

```text
old Fiber.Instance.SetFocus(false)
new Fiber.Instance.SetFocus(true)
```

也就是说，焦点状态属于 `ComponentInstance`，不是 VNode。以 Button 为例，当前实现位于 `ui/components/button/instance.go`，实例实现 `rtui.FocusableInstance`，并在 `SetFocus(bool)` 中更新 `control.InteractionState.Focused`。

## Layer / Modal 范围

`FiberFocusManager` 维护每个可聚焦 Fiber 的 effective layer。当 active layer 大于 `LayerBase` 时，`FocusNext()` / `FocusPrev()` 只在该 layer 内循环，用于 Modal 等 overlay 场景的焦点范围控制。

```text
activeLayer = LayerModal
  -> findNextInLayer(...)
  -> 只聚焦 modal layer 中的 focusable fiber
```

## 兼容路径

`runtime/ui/focus_manager.go` 中的 `VNodeFocusManager` 和 `internal/render/declarative_node_event.go` 中的 `HandleEvent` 仍保留兼容逻辑：

- `VNodeFocusManager.HandleEvent` 可处理 framework event 风格的 Tab。
- `DeclarativeNode.HandleEvent` 可在 legacy event fallback 中调用 focus manager。

新文档和新实现不应把这条路径写作当前主模型。

## 调试

启用焦点与 action 路由日志：

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG_FOCUS=true TUI_DEBUG_UI=true TUI_DEBUG_RENDER=true go run ./examples/ui_demos/demo1_full_featured
```

建议测试：

```bash
go test ./runtime/ui -run FiberFocus -count=1
go test ./ui/e2e -run Focus -count=1
go test ./framework -run ProcessMsg -count=1
```

## 维护约束

- 不要把 `FocusableVNode` 描述为当前焦点主模型。
- 不要把 `components/button/button.go` 这类历史路径写成当前组件路径；当前 Button 在 `ui/components/button/`。
- 不要把 `DeclarativeNode.HandleEvent()` 描述为所有输入的主入口；当前主入口是 `framework.App.processMsg`。
