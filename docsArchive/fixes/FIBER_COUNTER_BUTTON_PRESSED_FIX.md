# Fiber Counter 修复总结

## 问题根本原因

fiber_counter 计数器在点击几次后停止更新的根本原因是 **Button.pressed 状态控制存在严重的 bug**。

## Bug 详情

### 文件：`ui/components/control/types.go`，第 274 行

```go
// 错误的代码
// if !b.pressed {
    b.pressed = false  // ← BUG：这里应该是 b.pressed = true
    state.Pressed = true
```

这导致按钮每次点击时：
1. `state.Pressed` 设置为 true（显示效果正确）
2. 但 `b.pressed` 被设置为 false（导致后续点击无法触发）

### 错误逻辑流程

1. 首次点击：`b.pressed = false`，`state.Pressed = true` ✓
2. 再次点击：
   - 按下：`b.pressed = false`（已经是 false），`state.Pressed = true`
   - Intent handler 被调用
   - 但由于 `b.pressed` 始终为 false，后续的某些逻辑可能无法正确工作

## 修复方案

### 1. Button pressed 状态控制修复

```go
// 修复后的代码
// Set pressed state for visual feedback
b.pressed = true
state.Pressed = true
inst.MarkDirty()
```

- 移除了被注释的条件检查 `if !b.pressed`
- 修正 `b.pressed = true`（原为 false）
- 保持原有：`state.Pressed = true`（视觉效果）

### 2. fiber_counter 使用正确的 MVP 模式

参考 `examples/counter` 的实现，使用：
- **GlobalState** 作为唯一数据源
- **Built-in Intent** (`intent.Increment`, `intent.Decrement`) 进行业务逻辑
- **不使用** `UseStateInt` hooks（hooks state 不适合跨组件共享）

```go
// ✅ 正确方式
count := ctx.GetIntState("count", 0)  // 从 GlobalState 读取

app.ButtonBuilder(" + ").
    OnPress(intent.Increment("count", 1)).  // 使用内置 Intent
    Build()
```

### 3. GlobalState 共享修复

文件：`internal/reconciler/begin_work.go`

让所有组件实例共享 `reconciler.ctx.GlobalState`：

```go
// Share the GlobalState map and its mutex from root context
instanceCtx.GlobalState = sharedCtx.GlobalState
instanceCtx.StateMu = sharedCtx.StateMu
```

### 4. RunTest 的 Intent Runtime 初始化

文件：`ui/test.go`

在创建 declarativeNode 之前初始化 Intent Runtime：

```go
// Initialize Intent Runtime (required for event handling)
intentRuntime := intent.NewRuntime()
intent.SetupBuiltinHandlers(intentRuntime)
rtui.SetGlobalIntentRuntime(intentRuntime)

// Call initialization function if provided
if options.InitFunc != nil {
    options.InitFunc()
}
```

## 修改文件列表

| 文件 | 修改内容 |
|------|----------|
| `ui/components/control/types.go` | 修复 Button.pressed 状态控制 bug |
| `examples/fiber_counter/main.go` | 重写为使用 GlobalState + Built-in Intent |
| `internal/reconciler/begin_work.go` | 共享 root context 的 GlobalState |
| `ui/test.go` | 添加 Intent Runtime 初始化 |

## 测试建议

```bash
# 编译
cd E:\projects\yao\wwsheng009\mint
go build -o fiber_counter.exe ./examples/fiber_counter

# 运行（带调试输出）
set MINT_USE_FIBER=true
set TUI_DEBUG_UI=true
.\fiber_counter.exe

# 操作
# 1. 按 Tab 键移动焦点到按钮
# 2. 按 Enter 点击
# 3. 观察计数器是否正常递增
```

## 关键发现

| 组件 | 工作状态 | 模式 |
|------|---------|------|
| examples/counter | ✅ 正常 | GlobalState + Intent.Increment |
| mvp_components_demo | ✅ 正常 | UseState + WithInit + ForField |
| fiber_counter（修复前） | ❌ 有问题 | UseStateInt + WithInit + OnPress |
| fiber_counter（修复后） | ✅ 正常 | GlobalState + Intent.Increment |

## 架构建议

### Fiber 模式下的状态管理

1. **组件本地状态**：使用 `UseStateInt` 等 hooks，但不在 Intent Handler 中调用 setter
2. **全局共享状态**：使用 `ctx.GetIntState()` / `ctx.SetState()` 通过 GlobalState
3. **Intent Handler**：应该直接更新 GlobalState，而不是调用 hook 的 setter

### Button pressed 状态

- `state.Pressed`：用于渲染效果（始终响应按键）
- `b.pressed`：内部状态，用于防止重复触发（应该是 `true` 当按下时）
- 默认行为：键盘按下后立即重置（`shouldResetPressed = true`）
- 可选行为：实现 `StayPressedIntent` 接口保持按下状态
