# 焦点管理功能文档

本目录记录 Mint 当前焦点管理行为、已完成实现和历史问题分析。

## 当前能力

- Tab 键导航。
- Shift+Tab 反向导航。
- 鼠标点击切换焦点。
- Modal / Layer 场景下的焦点范围控制。
- Fiber-first 焦点收集和实例焦点状态写入。

## 当前实现模型

当前焦点主模型是 Fiber-first：

```text
Fiber tree
  -> FiberFocusManager collects focusable Fibers
  -> active layer / modal focus scope filtering
  -> current focused Fiber
  -> ComponentInstance.SetFocus(bool)
```

焦点状态属于运行期实例，而不是 VNode。

## 关键源码位置

| 功能 | 源码 |
|---|---|
| Fiber focus manager | `../../../runtime/ui/fiber_focus_manager.go` |
| Focus manager compatibility helpers | `../../../runtime/ui/focus_manager.go` |
| DeclarativeNode event/focus handling | `../../../internal/render/declarative_node_event.go` |
| App key/action routing | `../../../framework/app.go` |
| Runtime focus package | `../../../runtime/focus` |

重点方法：

- `runtime/ui.FiberFocusManager.HandleEvent`
- `runtime/ui.CollectFocusable`
- `runtime/ui.CollectFocusableInLayer`
- `internal/render.DeclarativeNode.handleMouseFocus`
- `framework.App.GetFocusManager`
- `framework.App.SetFocusManagerFromDeclarativeNode`

## 事件处理概览

当前事件路径不是单一旧 `HandleEvent` 路径。简化理解：

```text
platform input
  -> runtime/msg
  -> framework.App.processMsg
  -> global shortcuts / selection mode
  -> InputTracker / InteractionContext
  -> InputProcessor -> Action
  -> navigation action handled by FocusManager
  -> mouse action routed by TargetFiber when available
  -> keyboard action routed to focused Fiber when available
  -> ActionRouter fallback
```

`DeclarativeNode.HandleEvent()` 仍保留兼容路径，处理部分 focus 和 legacy event fallback，但新文档应优先描述 `Msg -> Action -> Fiber/FocusManager` 主路径。

## 可聚焦组件

可聚焦能力由组件实例提供。常见可聚焦组件包括：

- `ui/components/button`
- `ui/components/input`
- `ui/components/textarea`
- `ui/components/select`
- `ui/components/checkbox`
- `ui/components/radio`
- `ui/components/switch`
- `ui/components/tabs`
- `ui/components/menu`
- `ui/components/list`
- `ui/components/table`
- `ui/components/treeview`

具体以组件 instance 是否实现 focus 相关能力为准。

## Modal 和 Layer 焦点

当 Modal / active layer 存在时，FocusManager 会按 layer 范围收集可聚焦 Fiber，避免 Tab 跳到背景内容。

相关机制：

- Fiber 节点带 layer 信息。
- Portal-aware layout 和 Layer render path 保留 overlay/modal 结构。
- FocusManager 按 active layer 做焦点 scope。

## 调试

启用焦点和 UI 日志：

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG_FOCUS=true TUI_DEBUG_UI=true go run ./examples/ui_demos/demo1_full_featured
```

排查鼠标点击焦点：

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG_HITMAP=true TUI_DEBUG_PUMP=true go run ./examples/modal
```

运行相关测试：

```bash
go test ./runtime/focus ./runtime/ui -run Focus -count=1
go test ./ui/e2e -run Focus -count=1
go test ./examples/ui_demos/demo1_full_featured -count=1
```

## 文档列表

- [mouse_click_focus_issue.md](mouse_click_focus_issue.md): 鼠标点击焦点问题分析。
- [mouse_click_focus_implementation.md](mouse_click_focus_implementation.md): 鼠标点击焦点切换实现记录。
- [tab_key_focus_implementation.md](tab_key_focus_implementation.md): Tab 键焦点切换实现记录。

这些文档包含历史行号和旧路径时，应以当前源码为准。

## 维护注意

- 不要再把 `components/button/button.go` 这类旧路径写作当前实现路径；当前组件在 `ui/components/<name>/`。
- 不要把 `FocusableVNode` 描述为主模型；当前主模型是 Fiber + ComponentInstance。
- 若涉及鼠标命中，应同时参考 HitMap / TargetBounds 文档。

## 相关文档

- [../../event/long_term_event_architecture.md](../../event/long_term_event_architecture.md)
- [../../howto/migrate-to-targetbounds.md](../../howto/migrate-to-targetbounds.md)
- [../../architecture/README.md](../../architecture/README.md)
