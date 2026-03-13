# FocusManager 与 Button Pressed 状态冲突修复

## 问题描述

焦点管理器设置焦点时，导致 button 的 pressed 状态更新异常，可能造成按钮一直处于按下状态。

## 根本原因

### PressableBehavior.OnStateChange 的不合理实现

**修改前**：
```go
func (b *PressableBehavior) OnStateChange(inst Instance, oldState, newState InteractionState) {
    b.pressed = newState.Pressed
}
```

**问题**：
1. 这个实现假设 `inst.state.Pressed` 是唯一的状态源
2. 当焦点变化时，`OnStateChange` 会从 `inst.state.Pressed` 读取可能的过时值
3. 导致 `b.pressed` 被错误设置

### 问题流程

```
1. 用户点击按钮 (ActionMousePress)
   → b.pressed = true
   → inst.state.Pressed = true (OnAction 直接修改)

2. Intent 触发重新渲染
   → Reconciler 更新 fiber 树

3. FiberFocusManager.SetFocusByIndex 设置焦点
   → button.SetFocus(true)
   → inst.state.Focused = true
   → inst.behaviors.OnStateChange(inst, oldState, inst.state)

4. PressableBehavior.OnStateChange
   → b.pressed = newState.Pressed (读取到过时值 true)
   → 结果：b.pressed 被错误地设置为 true
```

**即使焦点变化，b.pressed 仍然保持 true**，这会导致按钮看起来一直处于按下状态。

## 设计问题

### 状态责任划分

```
InteractionState (inst.state):
  - Focused: 由 FocusableBehavior 与 FocusManager 协作管理
  - Pressed: 由 PressableBehavior 管理
  - Hovered: 由 HoverableBehavior 管理
  - Disabled: 由 DisableableBehavior 管理

当前设计的问题：
- PressableBehavior.OnAction 直接修改 inst.state.Pressed
- PressableBehavior.OnStateChange 从 inst.state.Pressed 同步回 b.pressed
- 这造成了状态的双向同步循环，容易出错
```

### 为什么焦点的变化不应该影响 Pressed 状态

`PressableBehavior.OnStateChange` 的语义问题：
- 当 `inst.state.Focused` 变化时，焦点管理器调用 `OnStateChange`
- 但 `Pressed` 状态应该是独立的，与焦点变化无关
- 同步 `Pressed` 状态会导致焦点管理器按下逻辑污染了交互状态

## 修复方案

### 修改 PressableBehavior.OnStateChange

**修改后**：
```go
// OnStateChange handles state changes.
// Note: PressableBehavior's pressed state is managed entirely by OnAction.
// We do NOT synchronize from newState.Pressed because:
// 1. b.pressed is the behavior's internal state (truth source)
// 2. Synchronizing from newState.Pressed can cause issues when focus changes trigger OnStateChange
// 3. Each behavior (Focusable, Hoverable, Disabled) manages its own state independently
func (b *PressableBehavior) OnStateChange(inst Instance, oldState, newState InteractionState) {
    // Pressed state is NOT synchronized here.
    // b.pressed is managed by OnAction only.
    // This prevents conflicts when the FocusManager changes the Focused state
    // and OnStateChange is called with stale Pressed values from inst.state.
}
```

### 修改原理

1. **b.pressed 是 PressableBehavior 的内部状态（truth source）**
   - 完全由 `OnAction` 管理
   - 不依赖 `inst.state.Pressed`

2. **移除从 inst.state.Pressed 的同步**
   - 避免焦点变化导致 pressed 状态同步问题
   - 解除状态双向同步的循环

3. **每个 Behavior 独立管理自己的状态**
   - FocusableBehavior: Focused
   - HoverableBehavior: Hovered
   - PressableBehavior: Pressed
   - DisableableBehavior: Disabled

## 改进建议

### 改进 1：考虑移除 inst.state.Pressed

既然 `b.pressed` 是 pressed 状态的唯一来源，可以考虑：

1. **不修改 inst.state.Pressed**，只在 Behavior 内部维护状态
2. 或者明确 `inst.state.Pressed` 是**只读的**，用于样式解析

### 改进 2：ResolveStyle 使用 Behavior 状态

```go
func ResolveStyle(baseStyle style.Style, state InteractionState, behaviors []Behavior) style.Style {
    if state.Disabled {
        return s
    }

    // 检查各个 Behavior 的状态（而不是只看 inst.state）
    for _, b := range behaviors {
        if pb, ok := b.(*PressableBehavior); ok && pb.IsPressed() {
            // Pressed state (从 Behavior 读取)
            return s
        }
    }

    if state.Focused {
        return s
    }

    if state.Hovered {
        return s
    }

    return s
}
```

## 相关代码

- `ui/components/control/types.go:318-331`: PressableBehavior.OnStateChange
- `ui/components/control/types.go:225-265`: PressableBehavior.OnAction
- `ui/components/button/instance.go:361-375`: button.SetFocus
- `runtime/ui/fiber_focus_manager.go:456-481`: FiberFocusManager.applyFocusState

## 测试状态

修复前的测试失败：`TestPressableBehavior_StayPressedIntent_Keep` - 这个测试失败在修改之前就存在，不是由此修复导致的。

修复后的编译状态：
```bash
cd E:\projects\yao\wwsheng009\mint
go build examples/absolute/main.go
# ✓ 编译成功
```

## 验证

1. 编译 absolute 示例
2. 点击 "Messages" 按钮
3. 多次点击验证 count 递增是否正常
4. 确认按钮没有始终保持 pressed 视觉状态
