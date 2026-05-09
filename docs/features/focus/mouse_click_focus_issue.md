# 鼠标点击与焦点管理

本文记录鼠标点击焦点问题的当前状态。早期问题是“点击按钮触发动作，但焦点不跟随鼠标目标”。当前源码已经在 Fiber-first action 路径中处理该问题。

## 当前结论

- 鼠标点击可聚焦组件时，App 会尝试把焦点切到命中的 `TargetFiber` 或其最近的可聚焦祖先。
- Button 不再依赖旧 `onClick func()` 或 VNode 内部 `hasFocus` 字段；当前 Button 运行期实现位于 `ui/components/button/instance.go`，通过 `control.FocusableBehavior`、`control.PressableBehavior` 和 `intent` 处理状态与动作。
- 当前主路径不要求组件自己在鼠标点击时调用 `RequestFocus()`。

## 当前主路径

```text
terminal mouse input
  -> framework/event.Pump
  -> HitMap lookup
  -> runtime/msg.MouseMsg{
       TargetID,
       TargetFiber,
       TargetBounds,
       LocalX,
       LocalY,
     }
  -> framework.App.processMsg
  -> runtime/action.InputProcessor
  -> ActionClick
  -> nearestFocusableFiber(TargetFiber)
  -> FiberFocusManager.SetFocusByID(...)
  -> ActionBridge.DispatchFromFiber(...)
```

## 关键源码

| 功能 | 当前源码 |
|---|---|
| MouseMsg 结构 | `runtime/msg/mouse_msg.go` |
| HitMap 命中与 TargetFiber 填充 | `framework/event/pump.go`、`internal/render/layout_switcher.go` |
| MouseMsg 到 ActionClick | `runtime/action/processor.go` |
| 鼠标焦点切换 | `framework/app.go` 的 `processMsg` |
| 最近可聚焦 Fiber 查找 | `framework/app.go` 的 `nearestFocusableFiber` 相关逻辑 |
| Button 运行期焦点状态 | `ui/components/button/instance.go` |
| Button 声明 props | `ui/components/button/vnode.go` |

## 焦点切换位置

在 `framework.App.processMsg` 的鼠标目标路由阶段，当前源码会先处理鼠标捕获，再根据命中的 Fiber 转移焦点：

```text
if MouseMsg has TargetFiber:
  if action is ActionClick:
    focusFiber := nearestFocusableFiber(targetFiber)
    focusManager.SetFocusByID(focusFiber.NodeID)
  dispatch action to target fiber
```

这样焦点跟随鼠标点击，同时 action 仍然分发给被命中的组件。

## 与 TargetBounds 的关系

鼠标命中依赖 HitMap，而不是组件本地估算 bounds：

- `TargetID` 标识命中的节点。
- `TargetFiber` 让 ActionBridge 可以直接从 Fiber 目标分发。
- `TargetBounds` 是经过布局和变换后的最终屏幕边界。
- `LocalX` / `LocalY` 是相对目标组件的坐标。

因此 Modal、Portal、绝对定位、层级覆盖等场景应优先依赖 HitMap / TargetBounds，而不是在组件中手写 `ContainsPoint`。

## 兼容路径

`internal/render/declarative_node_event.go` 中仍保留 `handleMouseFocus()`，用于 framework event 兼容路径。它会遍历 legacy focusable 节点并根据 bounds 切换焦点。

这不是当前主路径。新问题排查应优先看 `framework.App.processMsg`、`MouseMsg.TargetFiber` 和 HitMap。

## 排查步骤

1. 确认 `MouseMsg.TargetFiber` 不为空。
2. 确认目标 Fiber 或其祖先的 `Instance` 实现 `rtui.FocusableInstance`。
3. 确认实例未 disabled。
4. 确认 `FiberFocusManager.SetFocusByID` 能找到对应 NodeID。
5. 确认 action 没有被 middleware 提前 stop。

调试命令：

```bash
TUI_LOG_OUTPUT=console TUI_DEBUG_HITMAP=true TUI_DEBUG_PUMP=true TUI_DEBUG_FOCUS=true TUI_DEBUG_RENDER=true go run ./examples/modal
```

## 测试建议

```bash
go test ./framework -run Mouse -count=1
go test ./ui/e2e -run Focus -count=1
go test ./ui/e2e -run Button -count=1
```

## 维护约束

- 不要再建议修改历史路径 `components/button/button.go`。
- 不要把旧 `ButtonVNode.HandleEvent`、`onClick func()`、`b.hasFocus` 写作当前 Button 模型。
- 不要把 `RequestFocus()` 作为当前推荐方案；当前推荐是在 App action 路由层根据 `TargetFiber` 做焦点转移。
