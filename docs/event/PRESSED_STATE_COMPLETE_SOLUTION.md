# TUI Pressed 状态完整解决方案

## 文档信息

- **作者**: Qwen Code
- **日期**: 2026-02-28
- **相关文档**: docs/platform/key_release.md
- **状态**: 设计阶段

---

## 一、问题背景

### 1.1 现象

在 `examples/absolute/main.go` 中，按 Enter 键触发按钮后：
- 按钮的 `pressed` 状态变为 `true`
- Intent 正确触发（count 递增）
- 但 `pressed` 状态无法从 `true` 恢复为 `false`
- 按钮始终显示为"按下"状态

### 1.2 根本原因

**TUI 输入模型限制**：

1. **没有可靠的 key release 事件**
   - 终端协议（ANSI/xterm）的鼠标支持模式不一致
   - X10 模式只支持按下
   - 即使 SGR 模式有 release，也可能丢失（终端/ssh/tmux/终端切换）
   - Windows 兼容性更差

2. **依赖 release 事件是错误的**
   - GUI 思维：`mouseDown → pressed = true`, `mouseUp → pressed = false`
   - TUI 现实：release 事件可能永远不会到来

3. **现有方案的缺陷**

   | 方案 | 缺陷 |
   |-----|------|
   | `StayPressedIntent` | 需要外部逻辑重置，违反组件自治性 |
   | 焦点丢失推断 | 单焦点 UI 失效（焦点不会变化） |
   | `OnStateChange` 监听 | 无从区分"新键盘输入"和"状态更新" |

### 1.3 架构限制

当前架构流程：
```
vnode → fiber → layoutbox → paintable → action → intent
```

**关键限制**：
- `OnAction` 是 Action 唯一入口
- `OnStateChange` 不知道是哪个 Action 触发的
- 没有"帧级别"的输入状态跟踪
- 事件驱动模型（不可靠）
- 没有输入状态推断机制

---

## 二、设计原则（来自 key_release.md）

### 2.1 核心原则

👉 **输入不是事件流，而是"状态快照流"**

👉 **用"推断"代替 release 事件**

👉 **组件自管理状态（不依赖外部逻辑）**

### 2.2 输入驱动模式

#### ❌ 事件驱动（不可靠）

```go
// 错误的做法
case MouseActionPress:
    pressed = true
case MouseActionRelease:  // ❌ 可能永远不会触发
    pressed = false
```

#### ✅ 状态驱动（可靠）

```go
// 正确的做法
prev := lastState.Buttons
curr := currentState.Buttons

if prev == 0 && curr != 0 {
    // 推断：按下
    pressed = true
}
if prev != 0 && curr == 0 {
    // 推断：释放
    pressed = false
}
```

### 2.3 统一交互状态机

所有交互（Button/Drag/Select）使用统一的状态机模型：

```
InputState Snapshot (prev, curr)
        ↓
   推断 Intent (Press/Release/Move)
        ↓
   Interaction FSM (Idle/Hover/Pressed)
        ↓
   Component Update
```

---

## 三、架构设计

### 3.1 总体架构

```
[Platform Input] (tcell / runtime inputs)
        ↓
[Input Snapshot]   ← 帧级状态
        ↓
[InputTracker]     ← 推断边缘事件
        ↓
[IntentResolver]   ← 类型安全转换
        ↓
[InteractionFSM]   ← 统一状态机
        ↓
[Action Dispatcher] ← 现有 Action 系统
        ↓
[Component Instance] ← 按钮自更新状态
```

### 3.2 核心组件

#### 3.2.1 Input Snapshot（输入状态快照）

```go
// runtime/input/snapshot.go
package input

import (
    runtimeaction "github.com/wwsheng009/mint/runtime/action"
    runtimeplatform "github.com/wwsheng009/mint/runtime/platform"
)

// InputSnapshot 是输入状态的快照
// 包含鼠标、键盘等所有输入的当前状态
type InputSnapshot struct {
    // 鼠标状态
    MouseX       int
    MouseY       int
    MouseButton  runtimeplatform.MouseButton
    MouseAction  runtimeplatform.MouseAction
    
    // 键盘状态
    KeyboardKey  rune                    // 可打印字符
    SpecialKey   runtimeplatform.SpecialKey // 特殊键
    Modifiers    runtimeplatform.Modifier
    
    // 时间戳（用于事件顺序）
    Timestamp int64
}

// Clone 创建快照的副本
func (s *InputSnapshot) Clone() *InputSnapshot {
    return &InputSnapshot{
        MouseX:      s.MouseX,
        MouseY:      s.MouseY,
        MouseButton: s.MouseButton,
        MouseAction: s.MouseAction,
        KeyboardKey: s.KeyboardKey,
        SpecialKey:  s.SpecialKey,
        Modifiers:   s.Modifiers,
        Timestamp:   s.Timestamp,
    }
}
```

#### 3.2.2 InputTracker（输入跟踪器）

```go
// runtime/input/tracker.go
package input

import (
    log "github.com/wwsheng009/mint/runtime/logger"
    runtimeaction "github.com/wwsheng009/mint/runtime/action"
)

// InputTracker 跟踪输入状态变化，推断边缘事件
//
// 根据 docs/platform/key_release.md 的设计原则：
// - 不依赖底层事件完整性
// - 通过比较前一帧和当前帧状态推断 Press/Release
type InputTracker struct {
    // 上一帧状态
    lastSnapshot *InputSnapshot
    
    // 边缘事件推断结果
    lastIntent   InputIntent
}

// InputIntent 表示输入意图（状态变化）
type InputIntent interface {
    isInputIntent()
}

// 按钮按下意图
type InputPressIntent struct {
    X, Y     int
    Button   runtimeplatform.MouseButton
    Source   string // "mouse" | "keyboard"
}

func (InputPressIntent) isInputIntent() {}

// 按钮释放意图
type InputReleaseIntent struct {
    X, Y     int
    Button   runtimeplatform.MouseButton
    Source   string // "mouse" | "keyboard"
}

func (InputReleaseIntent) isInputIntent() {}

// 鼠标移动意图
type InputMoveIntent struct {
    X, Y int
}

func (InputMoveIntent) isInputIntent() {}

// 键盘输入意图
type InputKeyboardIntent struct {
    Key    rune
    Special runtimeplatform.SpecialKey
    Mod    runtimeplatform.Modifier
}

func (InputKeyboardIntent) isInputIntent() {}

// NewInputTracker 创建新的输入跟踪器
func NewInputTracker() *InputTracker {
    return &InputTracker{
        lastSnapshot: &InputSnapshot{},
        lastIntent:   nil,
    }
}

// Update 更新输入状态并返回推断的意图
//
// 核心推断逻辑：
// - prev.Button == 0 && curr.Button != 0 → 推断 Press
// - prev.Button != 0 && curr.Button == 0 → 推断 Release
// - prev.{X,Y} != curr.{X,Y} → 推断 Move
func (t *InputTracker) Update(snapshot *InputSnapshot) []InputIntent {
    if snapshot == nil {
        return nil
    }
    
    var intents []InputIntent
    
    // 鼠标状态推断
    t.inferMouseState(t.lastSnapshot, snapshot, &intents)
    
    // 键盘状态推断
    t.inferKeyboardState(t.lastSnapshot, snapshot, &intents)
    
    // 更新状态
    t.lastSnapshot = snapshot.Clone()
    t.lastIntent = nil
    
    return intents
}

// inferMouseState 推断鼠标状态变化
func (t *InputTracker) inferMouseState(prev, curr *InputSnapshot, intents *[]InputIntent) {
    // 推断 Press：之前无按钮，现在有按钮
    if prev.MouseButton == runtimeplatform.MouseNone && curr.MouseButton != runtimeplatform.MouseNone {
        *intents = append(*intents, InputPressIntent{
            X:       curr.MouseX,
            Y:       curr.MouseY,
            Button:  curr.MouseButton,
            Source:  "mouse",
        })
    }
    
    // 推断 Release：之前有按钮，现在无按钮
    if prev.MouseButton != runtimeplatform.MouseNone && curr.MouseButton == runtimeplatform.MouseNone {
        *intents = append(*intents, InputReleaseIntent{
            X:       curr.MouseX,
            Y:       curr.MouseY,
            Button:  prev.MouseButton,
            Source:  "mouse",
        })
    }
    
    // 推断 Move：位置变化
    if prev.MouseX != curr.MouseX || prev.MouseY != curr.MouseY {
        *intents = append(*intents, InputMoveIntent{
            X: curr.MouseX,
            Y: curr.MouseY,
        })
    }
}

// inferKeyboardState 推断键盘状态变化
func (t *InputTracker) inferKeyboardState(prev, curr *InputSnapshot, intents *[]InputIntent) {
    // 推断键盘输入：有新按键
    if (curr.KeyboardKey != 0 || curr.SpecialKey != runtimeplatform.KeyUnknown) {
        *intents = append(*intents, InputKeyboardIntent{
            Key:     curr.KeyboardKey,
            Special: curr.SpecialKey,
            Mod:     curr.Modifiers,
        })
    }
}
```

#### 3.2.3 InteractionFSM（统一交互状态机）

```go
// runtime/interaction/fsm.go
package interaction

import (
    "github.com/wwsheng009/mint/runtime/input"
)

// InteractionState 交互状态枚举
type InteractionState int

const (
    StateIdle     InteractionState = iota // 空闲
    StateHover                             // 悬停
    StatePressed                           // 按下
    StateDragging                          // 拖拽中
    StateSelecting                         // 选择中
)

// InteractionContext 交互上下文
//
// 跟踪全局的交互状态：
// - HotID: 当前鼠标悬停的组件
// - ActiveID: 当前按下的组件
type InteractionContext struct {
    HotID    int  // 当前 hover 的组件 ID
    ActiveID int  // 当前按下的组件 ID
    
    // 按下时的起始位置（用于拖拽判断）
    StartX, StartY int
    
    // 当前状态
    State InteractionState
    
    // 组件注册表（ID → Instance）
    Instances map[int]interface{}
}

// NewInteractionContext 创建新的交互上下文
func NewInteractionContext() *InteractionContext {
    return &InteractionContext{
        Instances: make(map[int]interface{}),
    }
}

// RegisterInstance 注册交互组件
func (c *InteractionContext) RegisterInstance(id int, inst interface{}) {
    if c.Instances == nil {
        c.Instances = make(map[int]interface{})
    }
    c.Instances[id] = inst
}

// UnregisterInstance 注销交互组件
func (c *InteractionContext) UnregisterInstance(id int) {
    delete(c.Instances, id)
}

// Update 更新交互状态
//
// 核心逻辑：
// - InputMove: 更新 HotID
// - InputPress: 设置 ActiveID
// - InputRelease: 处理 Click/Cancel
func (c *InteractionContext) Update(intents []input.InputIntent, hitTest func(int, int) int) {
    for _, intent := range intents {
        switch e := intent.(type) {
        case input.InputMoveIntent:
            c.handleMove(e.X, e.Y, hitTest)
            
        case input.InputPressIntent:
            c.handlePress(e.X, e.Y, hitTest, e.Source)
            
        case input.InputReleaseIntent:
            c.handleRelease(e.X, e.Y, hitTest, e.Source)
            
        case input.InputKeyboardIntent:
            c.handleKeyboard(e.Key, e.Special, e.Mod)
        }
    }
}

// handleMove 处理鼠标移动
func (c *InteractionContext) handleMove(x, y int, hitTest func(int, int) int) {
    id := hitTest(x, y)
    c.HotID = id
    
    // 检查是否进入拖拽状态
    if c.ActiveID != 0 {
        if abs(x-c.StartX) > DragThreshold || abs(y-c.StartY) > DragThreshold {
            c.State = StateDragging
        }
    }
}

// handlePress 处理按下事件
func (c *InteractionContext) handlePress(x, y int, hitTest func(int, int) int, source string) {
    id := hitTest(x, y)
    
    if id != 0 {
        c.ActiveID = id
        c.StartX = x
        c.StartY = y
        
        if source == "mouse" {
            c.State = StatePressed
        }
        // Keyboard press uses different logic (handled elsewhere)
    }
}

// handleRelease 处理释放事件
func (c *InteractionContext) handleRelease(x, y int, hitTest func(int, int) int, source string) {
    if c.ActiveID != 0 {
        targetID := hitTest(x, y)
        
        // 鼠标释放
        if source == "mouse" {
            if c.ActiveID == targetID {
                // Click：在同一组件内按下并释放
                c.dispatchClick(c.ActiveID)
            } else {
                // Cancel：拖出后释放
                c.dispatchCancel(c.ActiveID)
            }
        }
        
        c.ActiveID = 0
        if c.State == StateDragging {
            c.State = StateIdle
        }
    }
}

// handleKeyboard 处理键盘输入
func (c *InteractionContext) handleKeyboard(key rune, special runtimeplatform.SpecialKey, mod runtimeplatform.Modifier) {
    // 遵循 docs/platform/key_release.md 的设计：
    // 新的键盘输入应该重置所有交互状态
    c.resetAllPressedStates()
}

// dispatchClick 分发 click 事件
func (c *InteractionContext) dispatchClick(id int) {
    if inst, ok := c.Instances[id].(ClickHandler); ok {
        inst.OnClick()
    }
}

// dispatchCancel 分发 cancel 事件
func (c *InteractionContext) dispatchCancel(id int) {
    if inst, ok := c.Instances[id].(CancelHandler); ok {
        inst.OnCancel()
    }
}

// resetAllPressedStates 重置所有组件的 pressed 状态
func (c *InteractionContext) resetAllPressedStates() {
    for _, inst := range c.Instances {
        if handler, ok := inst.(PressedResetHandler); ok {
            handler.ResetPressed()
        }
    }
}

// 交互处理接口
type ClickHandler interface {
    OnClick()
}

type CancelHandler interface {
    OnCancel()
}

type PressedResetHandler interface {
    ResetPressed()
}

// 常量
const DragThreshold = 3 // 拖拽判定阈值（像素）

func abs(n int) int {
    if n < 0 {
        return -n
    }
    return n
}
```

#### 3.2.4 PressableBehaviorIntegration（集成到现有的 PressableBehavior）

```go
// ui/components/control/types.go
package control

import (
    "github.com/wwsheng009/mint/runtime/interaction"
)

// PressableBehavior 集成交互状态机
//
// 实现接口：
// - ClickHandler: 处理 click 事件
// - CancelHandler: 处理 cancel 事件（拖出后取消）
// - PressedResetHandler: 处理 pressed 状态重置
func (b *PressableBehavior) OnClick() {
    // Click 事件由 InteractionFSM 分发
    // Pressed 状态已经通过 InputTracker 推断的 Release 事件更新
}

func (b *PressableBehavior) OnCancel() {
    // Cancel 事件：拖出后取消
    b.pressed = false
    // Instance state 更新由 InteractionFSM 触发
}

func (b *PressableBehavior) ResetPressed() {
    // 新键盘输入时重置 pressed 状态
    if b.pressed {
        b.pressed = false
        // Instance state 更新需要通过 Instance 接口
        // 这里需要传递 Instance 参数或保存引用
    }
}
```

---

## 四、集成到现有架构

### 4.1 Framework/App 集成

```go
// framework/app.go
package framework

import (
    "github.com/wwsheng009/mint/runtime/input"
    "github.com/wwsheng009/mint/runtime/interaction"
)

type App struct {
    // ... 现有字段
    
    // 新增：输入跟踪和交互系统
    inputTracker      *input.InputTracker
    interactionCtx    *interaction.InteractionContext
}

func NewApp(root component.Node) *App {
    app := &App{
        // ... 现有初始化
        inputTracker:         input.NewInputTracker(),
        interactionCtx:       interaction.NewInteractionContext(),
    }
    return app
}

func (a *App) Run() error {
    // ... 现有初始化
    
    // 注册交互组件
    a.registerInteractionComponents(a.root)
    
    // 主循环
    for {
        // 1. 获取输入
        msg := a.Platform.Poll()
        
        // 2. 转换为 InputSnapshot
        snapshot := a.msgToSnapshot(msg)
        
        // 3. 推断边缘事件
        intents := a.inputTracker.Update(snapshot)
        
        // 4. 更新交互状态机
        a.interactionCtx.Update(intents, a.hitTest)
        
        // 5. 现有的 Msg → Action 处理（保持兼容）
        act := a.inputProcessor.ProcessMsg(msg)
        if act != nil {
            a.actionDispatcher.Dispatch(act)
        }
        
        // 6. 渲染
        a.render()
        
        // 7. 屏幕刷新
        a.Flush()
    }
}

// msgToSnapshot 将 Msg 转换为 InputSnapshot
func (a *App) msgToSnapshot(msg runtimemsg.Msg) *input.InputSnapshot {
    snapshot := &input.InputSnapshot{
        Timestamp: time.Now().UnixNano(),
    }
    
    switch m := msg.(type) {
    case *runtimemsg.MouseMsg:
        snapshot.MouseX = m.X
        snapshot.MouseY = m.Y
        snapshot.MouseButton = runtimeplatform.MouseButton(m.Button)
        snapshot.MouseAction = runtimeplatform.MouseAction(m.Action)
        
    case *runtimemsg.KeyMsg:
        snapshot.KeyboardKey = m.Rune
        snapshot.SpecialKey = m.Special
        snapshot.Modifiers = runtimeplatform.Modifier{
            Alt:   m.Mod.Alt,
            Ctrl:  m.Mod.Ctrl,
            Shift: m.Mod.Shift,
        }
    }
    
    return snapshot
}

// hitTest 命中测试（查找鼠标位置所在的组件）
func (a *App) hitTest(x, y int) int {
    // 实现命中测试逻辑
    // 遍历 Fiber 树，查找包含 (x,y) 的组件
    return a.hitTestFiber(a.rootFiber, x, y)
}

func (a *App) hitTestFiber(fiber *rtui.Fiber, x, y int) int {
    if fiber == nil {
        return 0
    }
    
    // 检查当前位置是否在组件边界内
    bounds := a.getComponentBounds(fiber)
    if x >= bounds.X && x < bounds.X+bounds.Width &&
       y >= bounds.Y && y < bounds.Y+bounds.Height {
        return int(fiber.NodeID)
    }
    
    // 递归检查子节点
    return a.hitTestFiber(fiber.Child, x, y)
}
```

### 4.2 迁移路径

#### Phase 1: 基础设施（1天）

1. 实现 `InputSnapshot`
2. 实现 `InputTracker`
3. 添加单元测试

#### Phase 2: 交互状态机（1天）

1. 实现 `InteractionContext` 和 `InteractionFSM`
2. 实现命中测试机制
3. 添加集成测试

#### Phase 3: 组件集成（1天）

1. 修改 `PressableBehavior` 接口
2. 集成到 `framework/App`
3. 更新现有组件（Button, Checkbox 等）

#### Phase 4: 验证和调优（1天）

1. 测试所有交互场景
2. 性能调优
3. 文档更新

---

## 五、验证测试

### 5.1 测试场景

| 场景 | 输入 | 预期结果 |
|-----|------|---------|
| 鼠标点击 | MousePress → MouseRelease | pressed: true → false (释放) |
| 鼠标拖出 | MousePress → MouseMove(外部) → MouseRelease | pressed: true → false (取消) |
| 键盘 Enter | ActionEnter | pressed: true, 保持 |
| 键盘新输入 | ActionEnter → 其他键盘输入 | pressed: true → false |
| 焦点切换 | ActionEnter → Tab | pressed: true → false |
| 单焦点 UI | ActionEnter（只有一个焦点） | pressed: true, 保持直到新输入 |

### 5.2 关键测试用例

```go
// runtime/input/tracker_test.go
func TestInputTracker_MousePressRelease(t *testing.T) {
    tracker := NewInputTracker()
    
    // 按下
    pressSnap := &InputSnapshot{
        MouseX:       10,
        MouseY:       10,
        MouseButton:  MouseLeft,
        MouseAction:  MouseActionPress,
    }
    intents := tracker.Update(pressSnap)
    
    if len(intents) != 1 {
        t.Errorf("Expected 1 intent, got %d", len(intents))
    }
    if _, ok := intents[0].(InputPressIntent); !ok {
        t.Error("Expected InputPressIntent")
    }
    
    // 释放
    releaseSnap := &InputSnapshot{
        MouseX:       10,
        MouseY:       10,
        MouseButton:  MouseNone,
        MouseAction:  MouseActionRelease,
    }
    intents = tracker.Update(releaseSnap)
    
    if len(intents) != 1 {
        t.Errorf("Expected 1 intent, got %d", len(intents))
    }
    if _, ok := intents[0].(InputReleaseIntent); !ok {
        t.Error("Expected InputReleaseIntent")
    }
}

func TestInteractionContext_KeyboardReset(t *testing.T) {
    ctx := NewInteractionContext()
    
    // 注册测试组件
    testInst := &TestPressableInstance{}
    ctx.RegisterInstance(1, testInst)
    
    // 按下
    ctx.Update([]InputIntent{
        InputPressIntent{X: 10, Y: 10, Button: MouseNone, Source: "keyboard"},
    }, nil)
    
    testInst.Pressed = true
    
    // 新键盘输入
    ctx.Update([]InputIntent{
        InputKeyboardIntent{Key: 'a'},
    }, nil)
    
    // 验证 pressed 状态被重置
    if testInst.Pressed {
        t.Error("Expected pressed to be reset on keyboard input")
    }
}
```

---

## 六、优势总结

### 6.1 符合设计原则

✅ **输入是状态快照流**（不是事件流）
✅ **用推断代替 release 事件**（不依赖不可靠的事件）
✅ **组件自管理状态**（不依赖外部逻辑）
✅ **统一交互模型**（Button/Drag/Select 共享 FSM）

### 6.2 支持的交互

- ✅ Button 点击（鼠标 + 键盘）
- ✅ 鼠标拖拽
- ✅ 框选
- ✅ 长按（可扩展）
- ✅ 双击（可扩展）

### 6.3 架构价值

- ✅ 可测试（Input → Intent 可回放）
- ✅ 可中断渲染（输入优先）
- ✅ 组件解耦（通过 Intent 通信）
- ✅ 类型安全（强类型 Intent）

---

## 七、相关文档

- [docs/platform/key_release.md](../../platform/key_release.md) - 设计原则
- [docs/platform/FIBER_ARCHITECTURE.md](../../platform/FIBER_ARCHITECTURE.md) - Fiber 架构
- [docs/fiber/fiber_first/fiber_intent.md](../../fiber/fiber_first/fiber_intent.md) - Intent 系统

---

## 八、实施检查清单

- [ ] Phase 1: InputSnapshot + InputTracker
- [ ] Phase 2: InteractionFSM
- [ ] Phase 3: 组件集成
- [ ] Phase 4: 验证测试
- [ ] 文档更新
- [ ] 代码审查
