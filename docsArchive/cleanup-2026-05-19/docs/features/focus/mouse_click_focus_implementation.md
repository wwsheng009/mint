# 鼠标点击切换焦点实现

本文记录当前源码中的鼠标点击焦点切换实现。早期实现文档曾聚焦 `DeclarativeNode.HandleEvent()` 中的 legacy event fallback；当前主实现已经迁移到 `framework.App.processMsg` 的 Fiber-first action 路径。

## 当前主实现

鼠标点击从终端输入到焦点切换的流程：

```text
RawInput(mouse)
  -> framework/event.Pump.convertToMouseMsg
  -> HitMap.HitTest
  -> MouseMsg.TargetFiber / TargetBounds / LocalX / LocalY
  -> framework.App.processMsg
  -> InputProcessor: MouseActionPress + MouseLeft -> ActionClick
  -> nearestFocusableFiber(TargetFiber)
  -> FiberFocusManager.SetFocusByID(NodeID)
  -> ActionBridge.DispatchFromFiber(TargetFiber, ActionClick, MouseMsg)
```

## 关键实现点

### 1. Pump 填充 TargetFiber

`framework/event/pump.go` 在转换鼠标输入时，会使用当前 HitMap 做命中测试，并把命中结果写入 `MouseMsg`：

```text
MouseMsg.TargetID
MouseMsg.TargetFiber
MouseMsg.TargetBounds
MouseMsg.LocalX / LocalY
```

这一步让后续 action 路由不需要重新遍历 VNode 树。

### 2. InputProcessor 生成语义 Action

`runtime/action/processor.go` 将左键 press 映射为 `ActionClick`：

```text
MouseActionPress + MouseLeft -> ActionClick
MouseActionRelease + MouseLeft -> ActionMouseRelease
MouseActionWheel -> ActionScroll
MouseActionMove -> ActionHover
```

### 3. App 处理焦点转移

`framework.App.processMsg` 在鼠标目标路由阶段：

- 对 press 建立 mouse capture。
- 对 release 应用并清理 capture。
- 当 action 是 `ActionClick` 时，查找 `TargetFiber` 最近的可聚焦 Fiber。
- 调用 `FiberFocusManager.SetFocusByID` 写入焦点状态。
- 继续通过 `ActionBridge.DispatchFromFiber` 把 action 分发给目标。

焦点切换和组件点击处理因此在同一条 action 路径中完成。

### 4. 组件实例接收焦点状态

可聚焦组件通过运行期实例实现 `rtui.FocusableInstance`：

```go
type FocusableInstance interface {
    ComponentInstance
    SetFocus(bool)
    HasFocus() bool
    IsDisabled() bool
}
```

Button 的当前实现位于 `ui/components/button/instance.go`。`SetFocus(true)` 会更新 `control.InteractionState.Focused` 并标记 dirty，后续 paint 使用该状态绘制焦点样式。

## legacy fallback

`internal/render/declarative_node_event.go` 仍包含：

- `DeclarativeNode.HandleEvent`
- `handleMouseFocus`
- `nodeWasClicked`

这些用于旧 framework event 路径。它们的存在不代表当前主输入路径仍依赖 VNode focusable 遍历。维护文档时应优先描述 `MouseMsg.TargetFiber -> App.processMsg -> FiberFocusManager`。

## 行为保证

当前实现目标：

- 鼠标点击可聚焦组件时焦点跟随目标。
- 点击非可聚焦子节点时，可向上找到最近可聚焦祖先。
- Modal / layer 场景下仍由 FocusManager 的 active layer 约束控制 Tab 范围。
- action 分发仍以实际命中的 `TargetFiber` 为目标，不因焦点转移而丢失鼠标目标。

## 调试

```bash
TUI_LOG_OUTPUT=console \
TUI_DEBUG_HITMAP=true \
TUI_DEBUG_PUMP=true \
TUI_DEBUG_FOCUS=true \
TUI_DEBUG_RENDER=true \
go run ./examples/modal
```

重点观察：

- HitTest 是否命中正确 bounds。
- `MouseMsg.TargetFiber` 是否存在。
- `ActionClick` 是否进入鼠标目标阶段。
- `FiberFocusManager.SetFocusByID` 是否成功。
- 目标组件的 `SetFocus(true)` 是否使实例 dirty。

## 测试建议

```bash
go test ./framework -run Mouse -count=1
go test ./ui/e2e -run Focus -count=1
go test ./ui/e2e -run Button -count=1
```

## 维护约束

- 新增组件若要参与焦点，应实现 `rtui.FocusableInstance`，而不是只在 VNode 上保存焦点字段。
- 鼠标目标应依赖 HitMap / `TargetFiber` / `TargetBounds`，不要在组件中复制一套全局命中路由。
- 历史 `DeclarativeNode.HandleEvent()` 文档只能作为兼容说明，不应作为当前架构主路径。
