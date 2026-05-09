# 事件架构现状与长期优化方案

本文档记录 Mint 当前事件/输入路径，以及仍未完成的长期目标。早期版本把很多内容写成未来计划；截至当前源码，Msg、HitMap、TargetFiber、Wheel Delta、MouseMove 合并等能力已经部分落地，但完整捕获/目标/冒泡和 Cmd 体系仍是目标。

## 当前已落地的主路径

当前主路径更接近：

```text
platform.RawInput
  -> framework/event.Pump
  -> runtime/msg.Msg
  -> framework.App.processMsg
  -> runtime/input.InputTracker
  -> runtime/interaction.InteractionContext
  -> runtime/action.InputProcessor
  -> runtime/action.Router / ActionBridge / ScopeDispatcher
  -> Fiber target instance or global handler
  -> Intent / Store / component state update
  -> schedule update / render
```

关键源码入口：

- `../../framework/event/pump.go`: 从平台输入读取 `platform.RawInput` 并转换为 `runtime/msg.Msg`。
- `../../framework/app.go`: 主循环、事件 drain、mouse move coalescing、`processMsg`。
- `../../framework/app_interaction.go`: `Msg` 到 `InputSnapshot`、HitMap 命中和 InteractionContext 集成。
- `../../runtime/msg`: Key/Mouse/Resize 等消息类型。
- `../../runtime/input`: 输入快照、tracker、keymap、mouse tracker。
- `../../runtime/action`: Action、InputProcessor、Router、ScopeDispatcher、middleware。
- `../../runtime/intent`: 语义 Intent 和类型化 dispatch。

## 当前已完成能力

### Msg 层

`framework/event.Pump` 当前输出 `runtime/msg.Msg`，而不是直接输出旧 `framework/event.Event`。旧 Event 适配桥仍存在，见 `framework/event/msg_adapter.go`。

### HitMap 命中

Render 后 `framework.App` 会从 root 的 `GetHitMap()` 取得命中图，并传给 `Pump`。鼠标消息可带：

- `TargetID`
- `TargetFiber`
- `LocalX`
- `LocalY`
- `TargetBounds`

这让很多鼠标交互不再依赖组件自己维护的旧 bounds。

### Wheel Delta

`MouseMsg` 已包含滚轮方向 delta。上滚和下滚可以在组件侧按 delta 处理。

### MouseMove 合并

`framework.App` 主循环 drain pending events 时会合并高频 mouse move，只保留最后一个 move，同时保留 press/release/wheel 和 key events。

### Action 主路径

当前 `processMsg` 优先走 `InputProcessor -> Action` 路径。鼠标 Action 可通过 `TargetFiber` 分发；键盘 Action 优先发给当前 focus Fiber；否则进入 `ActionRouter`。

### InteractionContext

`App` 已内置 `InputTracker` 与 `InteractionContext`，用于 pressed/click/cancel/reset 等输入状态推断和实例注册。

## 当前仍是兼容或历史路径的部分

- `framework/event.Event` 和 `MouseEvent` 仍存在，主要用于兼容旧组件和旧事件处理。
- `DeclarativeNode.HandleEvent()` 仍保留 VNode tree fallback，会调用旧式 `frameworkevent.Component.HandleEvent()`。
- `MsgToEvent` 是迁移期适配器，不应作为新架构主路径。

## 尚未完成的长期目标

### 阶段化分发

目标仍包括捕获、目标、冒泡三阶段，并支持：

- `StopPropagation`
- `PreventDefault`
- listener priority
- overlay/inspector/global shortcut 明确优先级

当前源码里还没有完整 DOM-style 阶段模型。

### 组件 Update API

目标中的 `Update(Msg) (Cmd, bool)` 尚不是当前组件主接口。当前组件运行期主接口是 `HandleAction(*action.Action) bool` 和各类 instance capability。

### Cmd 体系

标准化 Cmd，例如 `After(d)`、`Tick(d)`、`Batch`、`IO`，仍是长期目标。当前 tick、异步和组件内部行为仍由现有 runtime/app/instance 机制处理。

### 测试按节点注入

E2E 已具备 locator、hit assertion、trace 等能力，但 `InjectMouseByID(nodeID, action)` 这类统一 API 仍不应在文档里写成已完成，除非源码已提供同名能力。

## 长期目标形态

理想状态：

```text
Raw input / async result
  -> Msg
  -> HitMap path resolution
  -> capture phase
  -> target phase
  -> bubble phase
  -> Action / Intent / Update
  -> Cmd batch
  -> render scheduling
```

## 后续建议

1. 继续减少 VNode event fallback，把鼠标和键盘交互统一到 HitMap/TargetFiber/ActionBridge。
2. 给 `runtime/action` 和 `framework/event` 的边界补充更清晰的开发指南。
3. 为 E2E 提供稳定的按 component id / target id 注入辅助。
4. 若引入 Cmd，先从 tick/timeout/async task 这类易验证场景开始。
5. 将旧 `frameworkevent.Component.HandleEvent()` 文档明确标注为兼容路径。
