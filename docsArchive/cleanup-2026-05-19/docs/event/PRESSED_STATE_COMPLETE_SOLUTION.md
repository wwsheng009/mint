# TUI Pressed 状态解决方案现状

## 文档信息

- 原始设计日期: 2026-02-28
- 当前状态: 已部分实现并接入 `framework.App`
- 相关源码: `runtime/input`、`runtime/interaction`、`framework/app.go`、`framework/app_interaction.go`

## 背景

终端输入不能可靠依赖 key release / mouse release。不同终端、tmux、ssh、平台输入层对 release 支持不一致，因此 pressed 状态不能只按 GUI 模型实现：

```text
press -> pressed = true
release -> pressed = false
```

Mint 当前采用更可靠的输入状态推断模型：

```text
runtime/msg.Msg
  -> InputSnapshot
  -> InputTracker
  -> inferred InputIntent
  -> InteractionContext
  -> component instance reset/click/cancel behavior
```

## 当前源码实现

### App 字段

`framework.App` 当前持有：

- `inputTracker *input.InputTracker`
- `interactionCtx *interaction.InteractionContext`

这些字段在 `framework.NewApp()` 和 `framework.NewAppWithSource()` 中初始化。

### processMsg 集成顺序

当前 `framework.App.processMsg` 的相关顺序是：

1. 处理全局快捷键。
2. AppSelection 模式下 selection adapter 先消费鼠标/键盘选择。
3. 将 `runtime/msg.Msg` 转换为 `input.InputSnapshot`。
4. 调用 `inputTracker.Update(snapshot)` 推断输入意图。
5. 调用 `interactionCtx.Update(intents, a.hitTest)` 更新交互上下文。
6. 再进入 `InputProcessor.ProcessMsg` 转 Action。
7. 通过 FocusManager、ActionBridge、ActionRouter 分发。

### HitTest

`framework/app_interaction.go` 提供 HitMap 相关命中逻辑。鼠标位置可解析到 NodeID / Fiber target，用于 InteractionContext 判断 active/hot/click/cancel。

### 组件实例注册

`framework.App` 在 render 后刷新可交互实例，把支持相关能力的 instance 注册到 `InteractionContext`。这允许 InteractionContext 对 pressed reset、click、cancel 做统一协调。

## 与原始设计相比的差异

原始设计中部分伪代码已经过时：

- `NewApp(root component.Node)` 与当前源码不符；当前是 `framework.NewApp()` 后通过 `SetRoot(...)` 设置 root。
- 当前输入快照使用 `runtime/msg` 里的 key/mouse 类型，不是文档早期伪代码里的 `runtimeplatform.MouseAction` 组合。
- 当前 Action 系统已经是主路径，legacy event path 只是兼容路径。
- 当前已有帧级输入状态跟踪，不应再写“没有输入状态推断机制”。

## 已实现

- `InputSnapshot`
- `InputTracker`
- `InteractionContext`
- `framework.App` 集成 input tracker 和 interaction context
- Msg 到 InputSnapshot 转换
- HitMap 命中回调
- render 后交互实例刷新
- pressed/click/cancel/reset 基础机制

## 仍需持续验证

- 所有 pressable 组件是否都正确实现 reset/cancel 行为。
- 鼠标拖出、release 丢失、焦点切换、新键盘输入等边界是否都有 E2E 覆盖。
- 组件自定义 pressed 逻辑是否仍绕过统一 `InteractionContext`。
- 文档和示例是否仍引用 `StayPressedIntent` 作为主要方案。

## 验证建议

推荐按层运行：

```bash
go test ./runtime/input ./runtime/interaction -count=1
go test ./framework -run Interaction -count=1
go test ./ui/components/control ./ui/components/button -count=1
go test ./ui/e2e -run Button -count=1
```

全量命令：

```bash
go test ./... -count=1
```

在资源受限环境下，全量命令可能受 Go 工具链并发编译内存影响，应优先使用分层命令定位行为问题。

## 设计原则保留

仍应遵循：

- 输入应被看成状态快照流，而不是可靠 release 事件流。
- pressed 复位应通过推断和统一交互上下文完成。
- 组件实例负责自身视觉状态，但不要把全局交互推断分散在各组件里。
- 新组件应优先接入 `runtime/action`、`runtime/interaction` 和 instance capability。

## 相关文档

- [platform/key_release.md](../platform/key_release.md)
- [long_term_event_architecture.md](long_term_event_architecture.md)
- [../architecture/README.md](../architecture/README.md)
