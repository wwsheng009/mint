# Pressed 状态无法从 true 变成 false 的问题分析

## 问题

修改 `OnStateChange` 后，pressed 状态无法从 true 变成 false。

## 根本原因

在 `OnAction` 中，对于 `ActionEnter`/`ActionSubmit`，代码总是重置 pressed 状态：

```go
if b.pressIntent != nil {
    inst.EmitIntent(b.pressIntent)
    
    // ❌ 问题：总是重置，没有检查 StayPressedIntent
    if act.Type == action.ActionEnter || act.Type == action.ActionSubmit {
        b.pressed = false
        state.Pressed = false
    }
}
```

**没有检查 `StayPressedIntent.StayPressed()` 的返回值**。

### 执行流程

```
1. 用户按 Enter 键
   → OnAction(ActionEnter)
   → b.pressed = true
   → Emit Intent
   → b.pressed = false (总是重置，不管 Intent.StayPressed())

2. 结果：pressed 状态被立即重置，即使 Intent.StayPressed() == true
```

## 设计意图

根据 `control/types.go` 中的 `StayPressedIntent` 接口设计：

```go
type StayPressedIntent interface {
    intent.Intent
    StayPressed() bool
}
```

- `StayPressed() == true`: 保持 pressed 状态（用于导航、状态更新等）
- `StayPressed() == false`: 立即重置 pressed 状态（用于 Quit、Delete 等）

## 修复方案

### 修改 OnAction 中的重置逻辑

```go
func (b *PressableBehavior) OnAction(inst Instance, act *action.Action) bool {
    state := inst.GetState()
    if state.Disabled {
        return false
    }

    switch act.Type {
    case action.ActionPress, action.ActionClick, action.ActionEnter,
        action.ActionSubmit, action.ActionMousePress:
        if !b.pressed {
            b.pressed = true
            state.Pressed = true

            if b.pressIntent != nil {
                inst.EmitIntent(b.pressIntent)

                // For keyboard (ActionEnter, ActionSubmit), check if we should reset pressed state
                if act.Type == action.ActionEnter || act.Type == action.ActionSubmit {
                    // ✅ 修复：检查 StayPressedIntent
                    if stayPressed, ok := b.pressIntent.(StayPressedIntent); ok && stayPressed.StayPressed() {
                        // Keep pressed state
                    } else {
                        b.pressed = false
                        state.Pressed = false
                    }
                }
                // For mouse (ActionMousePress), don't reset (will wait for release)
            } else {
                // No intent: always reset pressed state immediately
                if act.Type == action.ActionEnter || act.Type == action.ActionSubmit {
                    b.pressed = false
                    state.Pressed = false
                }
            }

            inst.MarkDirty()
        }
        return true

    case action.ActionRelease, action.ActionPressEnd, action.ActionMouseRelease:
        if b.pressed {
            b.pressed = false
            state.Pressed = false
            inst.MarkDirty()
        }
        return true
    }
    return false
}
```

### 行为总结

| Intent 类型 | Action | StayPressed() | 结果 pressed 状态 |
|-----------|--------|---------------|------------------|
| `StayPressedIntent` (StayPressed=true) | ActionEnter | true | **保持 true** |
| `StayPressedIntent` (StayPressed=false) | ActionEnter | false | **变为 false** |
| 普通 | ActionEnter | N/A | **变为 false** |
| 任意 | ActionMousePress | N/A | **保持 true** (等待 release) |
| 任意 | ActionMouseRelease | N/A | **变为 false** |

## 测试验证

修复后，`TestPressableBehavior_StayPressedIntent_Keep` 应该通过：

```go
func TestPressableBehavior_StayPressedIntent_Keep(t *testing.T) {
    inst := newMockInstance()
    
    b := &PressableBehavior{
        pressIntent: KeepPressedIntent{},  // StayPressed() == true
    }
    
    act := action.NewAction(action.ActionEnter)
    handled := b.OnAction(inst, act)
    
    // ✅ 预期：b.pressed == true (保持)
    if !b.pressed {
        t.Errorf("Expected pressed=true for StayPressedIntent, got false")
    }
}
```

## 其他影响

### 1. 终端 UI 的行为

- **Enter/Submit**: 按 Enter → pressed = true → Emit Intent → (根据 StayPressed() 重置或保持)
- **Space**: 按 Space → pressed = true → Emit Intent → (无 StayPressed 检查，保持？需要确认)
- **MousePress**: 鼠标按下 → pressed = true → Emit Intent → 保持 true
- **MouseRelease**: 鼠标释放 → pressed = false

### 2. 渐进式导航

对于 `StayPressed() == true` 的 Intent（如导航）：
```
按 Enter → pressed = true (保持视觉反馈)
→ 用户看到按钮按下
→ 导航切换后，需要手动重置？
```

这需要在导航切换逻辑中添加 `ActionRelease` 或类似的操作来重置状态。

## 相关代码

- `ui/components/control/types.go`: `PressableBehavior.OnAction` (第 285-324 行)
- `ui/components/control/types.go`: `StayPressedIntent` 定义 (第 14-40 行)
- `ui/components/control/stay_pressed_test.go`: 测试
