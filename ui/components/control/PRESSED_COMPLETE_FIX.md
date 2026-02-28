# Pressed 状态问题完整分析 & 修复方案

## 问题根因分析

### 1. 当前架构流程

```
vnode -> fiber -> layoutbox -> paintable -> action -> intent
```

### 2. 问题定位

#### 问题点 1: MouseActionRelease 未被转换为 Action

**文件**: `runtime/action/processor.go`

```go
func (p *InputProcessor) processMouseMsg(mouseMsg *runtimemsg.MouseMsg) *Action {
    switch mouseMsg.Action {
    case runtimemsg.MouseActionPress:
        return NewAction(ActionClick)... // ✅ 有处理
        
    case runtimemsg.MouseActionWheel:
        return NewAction(ActionScroll)... // ✅ 有处理
        
    case runtimemsg.MouseActionMove:
        return NewAction(ActionHover)... // ✅ 有处理
        
    // ❌ MouseActionRelease 没有处理！直接返回 nil
    }
    return nil
}
```

**后果**:
- `MouseActionRelease` 消息被忽略，永远不会转换为 `ActionMouseRelease`
- PressableBehavior 的 `OnAction` 中 `case action.ActionRelease` 永远不会触发
- 使用鼠标点击时，pressed 状态在 `ActionClick` 处理后没有重置的机会

#### 问题点 2: 键盘 Enter 按下后没有对应 release

**根本设计问题**:
- 终端 UI (TUI) 没有可靠的 key release 事件
- 不能依赖 `ActionEnter` 或 `ActionSubmit` 后自动重置 pressed 状态
- 需要使用"状态机 + 推断"模式

#### 问题点 3: PressableBehavior 的设计问题

当前实现：
```go
case action.ActionEnter, action.ActionSubmit:
    b.pressed = false  // ❌ 立即重置
```

这个设计的问题：
- 用户看不到视觉反馈（按钮瞬间按下又弹起）
- 不符合 TUI 的交互模型（需要保持状态直到下次操作）

---

## 解决方案（基于 key_release.md 设计思想）

根据 `docs/platform/key_release.md` 的建议，实施以下架构：

### 设计原则（来自 key_release.md）

```
Input Snapshot → Intent Resolver → Interaction FSM → Action → Intent
```

👉 **输入不是事件流，而是"状态快照流"**
👉 **用"推断"代替 release 事件**

### 关键设计点

#### 1. Input Snapshot（输入状态快照）

```go
type InputState struct {
    MouseX, MouseY  int
    MouseButton     uint8      // 按钮状态（位掩码）
    KeyState        map[bool]  // 键盘状态（可选）

    // 推断出的边缘事件（由 Resolver 计算）
    Pressed  bool  // prev.Button == 0 && curr.Button != 0
    Released bool  // prev.Button != 0 && curr.Button == 0
}
```

#### 2. 推断逻辑（核心算法）

```go
type MouseTracker struct {
    lastButton uint8
}

func (m *MouseTracker) Update(currButton uint8) (pressed, released bool) {
    if m.lastButton == 0 && currButton != 0 {
        pressed = true  // 推断: 按下
    }
    if m.lastButton != 0 && currButton == 0 {
        released = true // 推断: 释放
    }
    m.lastButton = currButton
    return
}
```

#### 3. 统一状态机（Interaction FSM）

```text
Idle
 ├─ mouse over → Hover
 └─ mouse down inside → Pressed

Pressed
 ├─ mouse move outside → Idle（取消）
 ├─ button released → Click（如果还在组件内）
 └─ released outside → Idle（取消）

Hover
 ├─ mouse down → Pressed
 └─ mouse leave → Idle
```

---

## 具体实施方案（Phase-by-Phase）

### Phase 1: 修复 InputProcessor（立即修复）

**目标**: 让 `MouseActionRelease` 能够被正确转换为 Action

**文件**: `runtime/action/processor.go`

```go
func (p *InputProcessor) processMouseMsg(mouseMsg *runtimemsg.MouseMsg) *Action {
    switch mouseMsg.Action {
    case runtimemsg.MouseActionPress:
        if mouseMsg.Button == runtimemsg.MouseLeft {
            act := NewAction(ActionClick).
                WithSource("mouse").
                WithPayload(mouseMsg)
            if mouseMsg.TargetID != 0 {
                act.WithTargetID(mouseMsg.TargetID)
            }
            return act
        }
        // ... 其他按钮同理
        
    case runtimemsg.MouseActionRelease:
        // ✅ 添加 Release 处理
        if mouseMsg.Button == runtimemsg.MouseLeft {
            act := NewAction(ActionMouseRelease).
                WithSource("mouse").
                WithPayload(mouseMsg)
            if mouseMsg.TargetID != 0 {
                act.WithTargetID(mouseMsg.TargetID)
            }
            return act
        }
        // ... 其他按钮的 release
        
    case runtimemsg.MouseActionWheel:
        // ...
        
    case runtimemsg.MouseActionMove:
        // ...
    }
    
    return nil
}
```

**影响**: 修复后，PressableBehavior 的 `case action.ActionRelease` 会正确处理鼠标释放

---

### Phase 2: 移除 StayPressedIntent 机制（重构）

**问题分析**:
- `StayPressedIntent` 是临时 hack 方案
- 依赖外部逻辑来重置状态
- 不符合状态机模式

**方案**: 使用"焦点变化推断"代替

```go
// PressableBehavior.OnStateChange
func (b *PressableBehavior) OnStateChange(inst Instance, oldState, newState InteractionState) {
    // ❌ 旧方案：Pressed 状态不同步
    // ✅ 新方案：焦点丢失时重置 Pressed 状态

    // 当焦点丢失时，重置 pressed 状态
    if oldState.Focused && !newState.Focused && b.pressed {
        b.pressed = false
        instState := inst.GetState()
        instState.Pressed = false
        inst.MarkDirty()
    }
}
```

**设计理由**:
- 在 GUI 中，焦点丢失通常会取消操作
- 用户按 Enter → pressed = true → 意图发送 → 下次焦点切换 → 焦点丢失事件触发 → pressed = false
- 这个模式更符合用户心智模型

---

### Phase 3: 实施完整的状态跟踪（长期方案）

根据 `key_release.md` 完整实施：

#### 3.1 创建 InputState Tracker

**文件**: `runtime/input/state_tracker.go`

```go
package input

import (
    runtimemsg "github.com/wwsheng009/mint/runtime/msg"
)

// InputState 输入状态快照
type InputState struct {
    MouseX, MouseY  int
    MouseButton     uint8
}

// InputTracker 输入状态跟踪器
type InputTracker struct {
    lastState *InputState
}

func NewInputTracker() *InputTracker {
    return &InputTracker{
        lastState: &InputState{},
    }
}

// Update 更新状态并返回推断的事件
func (t *InputTracker) Update(msg runtimemsg.Msg) (hover, pressed, released bool) {
    currState := t.currentStateFromMsg(msg)
    
    // 推断边缘事件
    if t.lastState.MouseButton == 0 && currState.MouseButton != 0 {
        pressed = true
    }
    if t.lastState.MouseButton != 0 && currState.MouseButton == 0 {
        released = true
    }
    if t.lastState.MouseX != currState.MouseX || t.lastState.MouseY != currState.MouseY {
        hover = true
    }
    
    t.lastState = currState
    return
}

func (t *InputTracker) currentStateFromMsg(msg runtimemsg.Msg) *InputState {
    if mouseMsg, ok := msg.(*runtimemsg.MouseMsg); ok {
        return &InputState{
            MouseX:    mouseMsg.X,
            MouseY:    mouseMsg.Y,
            MouseButton: t.buttonToMask(mouseMsg.Button, mouseMsg.Action),
        }
    }
    return t.lastState
}

func (t *InputTracker) buttonToMask(button runtimemsg.MouseButton, action runtimemsg.MouseAction) uint8 {
    if action == runtimemsg.MouseActionRelease {
        return 0
    }
    switch button {
    case runtimemsg.MouseLeft:
        return 1
    case runtimemsg.MouseRight:
        return 2
    case runtimemsg.MouseMiddle:
        return 4
    }
    return 0
}
```

#### 3.2 集成到 Framework/App

**文件**: `framework/app.go`

```go
type App struct {
    // ... 现有字段

    inputTracker *input.InputTracker  // ✅ 新增
}

func (a *App) processMsg(msg runtimemsg.Msg) {
    // 推断边缘事件
    hover, pressed, released := a.inputTracker.Update(msg)
    
    // 转换为 Action
    act := a.inputProcessor.ProcessMsg(msg)
    
    // 补充 Release Action（原始消息可能没有）
    if act == nil && released {
        act = action.NewAction(action.ActionMouseRelease).WithPayload(msg)
    }
    
    // ... 后续处理不变
}
```

---

## 修复优先级

### Priority 1: 立即修复

**文件**: `runtime/action/processor.go`
**任务**: 添加 `MouseActionRelease` 处理
**时间**: 5 分钟
**效果**: 鼠标点击后 pressed 状态能够正确重置

### Priority 2: 短期重构

**文件**: `ui/components/control/types.go`
**任务**: 修改 `OnStateChange`，在焦点丢失时重置 pressed 状态
**时间**: 10 分钟
**效果**: 键盘 Enter 操作后的 pressed 状态能够在焦点切换时重置

### Priority 3: 长期架构

**文件**: 新建 `runtime/input/state_tracker.go`
**任务**: 实施完整的 InputState Tracker
**时间**: 1-2 小时
**效果**: 完整的状态跟踪和推断机制，符合 `key_release.md` 的设计原则

---

## 测试验证

### 测试场景 1: 鼠标点击

1. 点击按钮 → `ActionClick` → pressed = true
2. 释放鼠标 → `ActionMouseRelease` → pressed = false

### 测试场景 2: 键盘 Enter

1. Focus 在按钮上 → pressed = false
2. 按 Enter → `ActionEnter` → pressed = true → 意图发送
3. Tab 切换焦点 → 焦点丢失 → `OnStateChange` 触发 → pressed = false

### 测试场景 3: 鼠标拖出

1. 按下按钮（在组件内）→ pressed = true
2. 拖出组件 → 鼠标离开事件 → pressed = false（取消操作）
3. 释放鼠标 → 无触发（因为已取消）

---

## 总结

| 方案 | 阶段 | 优点 | 缺点 | 推荐度 |
|-----|-----|------|------|--------|
| 修复 InputProcessor | Priority 1 | 立即解决鼠标问题 | 键盘问题仍存在 | ⭐⭐⭐⭐⭐ |
| 焦点推断 | Priority 2 | 符合用户心智模型 | 依赖焦点切换 | ⭐⭐⭐⭐ |
| StayPressedIntent | 当前方案 | 简单 | hack，不可靠 | ⭐⭐ |
| InputState Tracker | Priority 3 | 完整方案，可扩展 | 工作量大 | ⭐⭐⭐⭐⭐ |

**建议执行顺序**: Priority 1 → Priority 2 → Priority 3

这样可以：
1. 立即解决最常见的问题（鼠标点击）
2. 短期内改善用户体验
3. 长期建立健壮的架构
