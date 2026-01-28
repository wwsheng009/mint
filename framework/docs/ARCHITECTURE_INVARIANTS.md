# Architecture Invariants (V3)

> **版本**: V3
> **核心原则**: 这些不变量是框架长期演进的"护栏"
> **重要性**: 违反这些规则的代码不应该被接受

## 概述

架构不变量是框架长期演进的"护栏"。它们不是可选的最佳实践，而是必须遵守的规则。违反不变量的代码在 code review 中应该被拒绝。

---

## 🔒 不变量 1：Runtime 永远不知道"组件是什么"

### 规则

`runtime` 包**绝不导入** `framework` 或任何组件类型。

### 允许

```go
// ✅ Runtime 只知道：
- Node / Tree（抽象节点）
- Layout（布局计算）
- Focus Graph（焦点图）
- CellBuffer（虚拟画布）
- Dirty Region（脏区域）
- Action（语义指令）
- State Snapshot（状态快照）
```

### 禁止

```go
// ❌ Runtime 绝不导入：
import "github.com/wwsheng009/mint/framework"
import "github.com/wwsheng009/mint/framework/component"
import "github.com/wwsheng009/mint/framework/event"

// ❌ Runtime 绝不知道：
- Button
- Input
- Table
- Component 接口（任何 framework 定义的接口）
```

### 检查方法

```bash
# 检查 Runtime 是否违反不变量
grep -r "framework" tui/runtime/
grep -r "component\." tui/runtime/
# 应该返回空结果（除了注释和测试）
```

### 违反后果

一旦 Runtime 知道组件类型：
- Runtime 无法被其他框架复用
- 无法进行 headless 测试
- Runtime 和 Framework 变成耦合体
- 违反分层架构原则

---

## 🔒 不变量 2：所有 UI 行为必须能被 Replay

### 规则

任何 UI 状态变化必须能追溯到一次 `Dispatch(Action)` 调用。

### 允许

```go
// ✅ 正确：状态变化通过 Action
runtime.Dispatch(Action{
    Type: ActionInputText,
    Payload: "hello",
})
// → State Update → Dirty → Render
```

### 禁止

```go
// ❌ 错误：直接修改状态
input.value = "world"  // 绕过 Action

// ❌ 错误：偷偷改字段
component.text = "new text"  // 无法 replay

// ❌ 错误：时间相关行为不可控
time.Sleep(time.Second)  // 无法记录/重放
```

### 检查方法

```go
// 每个 Action 应该能记录并重放
type ActionLog struct {
    Timestamp time.Time
    Action    Action
    BeforeState StateSnapshot
    AfterState  StateSnapshot
}

// 测试：任何状态变化都应该能从 Action Log 重放
func TestStateReproducibility(t *testing.T) {
    // 记录 Action 序列
    log := captureActionLog(func() {
        // 执行一些操作
    })

    // 重放应该得到相同的状态
    finalState := replayActions(log)
    assert.Equal(t, currentState, finalState)
}
```

### 违反后果

一旦允许绕过 Action：
- UI 自动化测试变得不可能
- AI 无法精确控制 UI
- Bug 无法复现
- 无法录制/回放用户操作
- 时间旅行调试成为不可能

---

## 🔒 不变量 3：Render 永远是幂等的

### 规则

给定相同的 State、Layout、Time，Render **必须**画出完全一样的结果。

### 允许

```go
// ✅ 正确：Render 只依赖显式状态
func (t *Text) Paint(ctx PaintContext, buf *CellBuffer) {
    content := t.state.Content  // 来自 state
    style := t.state.Style     // 来自 state
    // ...
}
```

### 禁止

```go
// ❌ 错误：Render 读取外部状态
func (t *Text) Paint(ctx PaintContext, buf *CellBuffer) {
    if globalConfig.DarkMode {  // 读取全局状态
        // ...
    }
}

// ❌ 错误：Render 有随机性
func (b *Button) Paint(ctx PaintContext, buf *CellBuffer) {
    if rand.Float64() < 0.5 {  // 非确定性
        // ...
    }
}

// ❌ 错误：Render 修改状态
func (c *Component) Paint(ctx PaintContext, buf *CellBuffer) {
    c.counter++  // 副作用！
}
```

### 检查方法

```go
// 测试幂等性
func TestRenderIdempotent(t *testing.T) {
    comp := NewComponent()
    ctx := NewPaintContext()
    buf1 := NewBuffer(80, 24)
    buf2 := NewBuffer(80, 24)

    comp.Paint(ctx, buf1)
    comp.Paint(ctx, buf2)

    assert.Equal(t, buf1, buf2)  // 必须相等
}

// 测试多次渲染结果一致
func TestRenderConsistent(t *testing.T) {
    comp := NewComponent()
    ctx := NewPaintContext()

    var results []string
    for i := 0; i < 10; i++ {
        buf := NewBuffer(80, 24)
        comp.Paint(ctx, buf)
        results = append(results, buf.String())
    }

    for i := 1; i < len(results); i++ {
        assert.Equal(t, results[0], results[i])
    }
}
```

### 违反后果

一旦 Render 不幂等：
- Diff 结果不可信
- Bug 无法复现
- 优化可能破坏功能
- AI 会产生不可预测的行为
- 测试变得不可靠

---

## 🔒 不变量 4：Component 不允许直接操作 Terminal

### 规则

所有输出必须通过：
```
Component → Painter → CellBuffer → Terminal
```

### 允许

```go
// ✅ 正确：通过 Paint 间接输出
func (t *Text) Paint(ctx PaintContext, buf *CellBuffer) {
    buf.SetCell(x, y, rune, style)  // 写 buffer，不直接输出
}
```

### 禁止

```go
// ❌ 错误：直接输出
import "github.com/wwsheng009/mint/framework/platform"

func (c *Component) SomeMethod() {
    platform.Terminal.WriteString("direct output")  // 越界！
}

// ❌ 错误：直接操作光标
func (i *Input) UpdateCursor() {
    terminal.MoveCursor(i.x, i.y)  // 绕过 framework
}
```

### 检查方法

```bash
# 检查 component 包是否导入 platform
grep -r "platform" tui/framework/component/
# 应该返回空结果（除了测试代码）
```

### 违反后果

一旦 Component 直接操作 Terminal：
- 无法实现双缓冲
- 无法实现 Diff 渲染
- 无法做单元测试
- 框架层被绕过
- 渲染顺序混乱

---

## 🔒 不变量 5：没有隐式全局状态

### 规则

所有状态必须：
1. 可枚举
2. 可快照
3. 可追踪变化

### 允许

```go
// ✅ 正确：显式状态
type AppState struct {
    Components map[string]ComponentState
    Focus      FocusPath
    Modals     []string
}

func (s *AppState) Snapshot() StateSnapshot {
    // 完整的状态快照
}
```

### 禁止

```go
// ❌ 错误：隐式全局变量
var currentFocus Component  // 谁知道当前焦点在哪？

var globalStyle Style      // 样式从哪来？

var isDirty bool           // 怎么知道谁 dirty？

// ❌ 错误：隐藏在闭包里
func makeHandler() func() {
    count := 0  // 外部无法访问的状态
    return func() {
        count++  // 隐式状态变化
    }
}
```

### 检查方法

```go
// 应该能获取完整状态快照
type StateSnapshot struct {
    Timestamp  time.Time
    FocusPath  []string
    Components map[string]ComponentState
    Modals     []ModalState
}

func (app *App) GetState() StateSnapshot {
    // 返回完整的、可枚举的状态
}

// 测试：所有状态都应该能被序列化
func TestStateSerializable(t *testing.T) {
    snapshot := app.GetState()
    data, err := json.Marshal(snapshot)
    assert.NoError(t, err)
    assert.NotEmpty(t, data)
}
```

### 违反后果

一旦有隐式全局状态：
- Debug 变成玄学
- AI 自动操作不可预测
- 状态无法恢复
- 无法实现时间旅行调试
- 测试覆盖率假象

---

## 🔒 不变量 6：Input ≠ Action

### 规则

- Platform 只产生 `RawInput`
- Runtime 负责转换 `RawInput → Action`
- Component 只处理 `Action`

### 允许

```go
// ✅ 正确：分层清晰
Platform (stdin)
    ↓ RawInput
Runtime (KeyMap)
    ↓ Action
Component (HandleAction)
    ↓ Handler
State Update
```

### 禁止

```go
// ❌ 错误：Component 处理原始按键
func (i *Input) HandleEvent(ev Event) bool {
    if key, ok := ev.(*KeyEvent); ok {
        if key.Key == 'a' {  // 直接判断按键
            // ...
        }
    }
}

// ❌ 错误：KeyMap 在 Framework 层
framework/keymap.go  // 应该在 Runtime
```

### 检查方法

```go
// Component 应该只处理 Action
type ActionHandler interface {
    HandleAction(a Action) bool
}

// 不应该直接处理 KeyEvent
func TestComponentNoKeyEvent(t *testing.T) {
    comp := NewInput()

    // 应该没有 HandleKeyEvent 方法
    assert.Nil(t, comp.HandleKeyEvent)

    // 应该有 HandleAction 方法
    assert.Implements(t, (*ActionHandler)(nil), comp)
}
```

### 违反后果

一旦 Component 处理原始输入：
- 无法支持不同键盘布局
- 无法实现自定义快捷键
- AI 无法抽象操作
- 国际化支持困难
- 输入法支持复杂

---

## 🔒 不变量 7：DSL/Spec 是一等公民

### 规则

Builder API 只是语法糖，`ComponentSpec` 才是主入口。

### 允许

```go
// ✅ 正确：Spec 是主要形式
type ComponentSpec struct {
    Type   string
    Props  map[string]interface{}
    Style  string
    Events map[string]string
}

func LoadSpec(spec ComponentSpec) Component {
    // 从 spec 创建组件
}

// Builder API 只是语法糖
func NewText(content string) *Text {
    return LoadSpec(ComponentSpec{
        Type: "text",
        Props: map[string]interface{}{"content": content},
    }).(*Text)
}
```

### 禁止

```go
// ❌ 错误：Builder API 是唯一入口
// 如果无法用 Builder 创建，就无法创建组件

// ❌ 错误：Spec 和 Builder 不等价
// 有些功能只能用 Builder，有些只能用 Spec
```

### 检查方法

```bash
# 应该能从 JSON/YAML 创建任何组件
cat component.json | jq .
# → 生成等价的组件树

# 测试：Spec 和 Builder 等价
func TestSpecBuilderEquivalence(t *testing.T) {
    // 从 Builder 创建
    comp1 := NewText("hello")

    // 从 Spec 创建
    comp2 := LoadSpec(ComponentSpec{
        Type: "text",
        Props: map[string]interface{}{"content": "hello"},
    })

    // 应该等价
    assert.Equal(t, comp1.GetState(), comp2.GetState())
}
```

### 违反后果

一旦 Spec 不是一等公民：
- AI 生成 UI 困难
- DSL 变成第二套 API
- 动态加载组件困难
- 配置文件无法驱动 UI

---

## 🔒 不变量 8：事件流必须有明确阶段

### 规则

每个 UI 事件必须经过：
1. **Capture Phase**: Root → Target
2. **Target Phase**: 目标组件
3. **Bubble Phase**: Target → Root

### 允许

```go
// ✅ 正确：明确的事件阶段
type EventPhase int

const (
    PhaseCapture EventPhase = iota
    PhaseTarget
    PhaseBubble
)

func (r *Router) Dispatch(ev Event) {
    // 1. Capture
    for _, handler := range r.captureHandlers {
        handler.HandleEvent(ev, PhaseCapture)
        if ev.Stopped() { return }
    }

    // 2. Target
    if target := ev.Target(); target != nil {
        target.HandleEvent(ev, PhaseTarget)
    }

    // 3. Bubble
    for _, parent := range r.getParentChain(ev.Target()) {
        parent.HandleEvent(ev, PhaseBubble)
        if ev.Stopped() { return }
    }
}
```

### 禁止

```go
// ❌ 错误：不分阶段直接派发
func (r *Router) Dispatch(ev Event) {
    target.HandleEvent(ev)  // 谁先谁后？顺序不确定
}

// ❌ 错误：只有冒泡，没有捕获
```

### 检查方法

```go
// 应该能明确知道事件在哪个阶段
type EventContext struct {
    Phase    EventPhase
    Current  Component
    Target   Component
}

// 测试：事件阶段验证
func TestEventPhases(t *testing.T) {
    root := NewContainer()
    child := NewButton()
    root.Add(child)

    phases := []EventPhase{}

    // 订阅所有阶段
    root.OnEvent(func(ev Event) {
        phases = append(phases, ev.Phase)
    })

    // 触发事件
    child.Click()

    // 验证阶段顺序
    assert.Equal(t, []EventPhase{PhaseCapture, PhaseTarget, PhaseBubble}, phases)
}
```

### 违反后果

一旦事件阶段不明确：
- 事件顺序变成"玄学"
- 不同组件作者写出不兼容逻辑
- 全局拦截无法实现
- Modal/Dialog 行为不一致

---

## V3 新增不变量

### 🔒 不变量 9：动画采用按需 Tick

### 规则

动画定时器只在有活动动画时运行。

### 允许

```go
// ✅ 正确：按需 Tick
func (mgr *Manager) AddAnimation(anim Animation) {
    mgr.animations = append(mgr.animations, anim)
    if len(mgr.animations) == 1 {
        mgr.ticker.Start()  // 第一个动画时才启动
    }
}

func (mgr *Manager) RemoveAnimation(id string) {
    // 移除后如果没动画了，停止 ticker
    if len(mgr.animations) == 0 {
        mgr.ticker.Stop()
    }
}
```

### 禁止

```go
// ❌ 错误：全局 Tick
func (app *App) Run() {
    ticker := time.NewTicker(16ms) // 始终运行
    for {
        select {
        case <-ticker.C:
            app.Update()  // 即使没有动画也在运行
        }
    }
}
```

---

### 🔒 不变量 10：状态必须是显式的

### 规则

所有状态必须可以通过 `StateSnapshot` 完整枚举。

### 允许

```go
// ✅ 正确：状态集中管理
type ComponentState struct {
    ID    string
    Type  string
    Props map[string]interface{}
    State map[string]interface{}
    Rect  Rect
}

func (c *Component) ExportState() ComponentState {
    return ComponentState{
        ID:    c.id,
        State: c.stateHolder.GetState(),
        Props: c.stateHolder.GetProps(),
    }
}
```

### 禁止

```go
// ❌ 错误：状态分散在闭包中
func makeCounter() func() int {
    count := 0  // 隐藏状态
    return func() int {
        count++
        return count
    }
}
```

---

## 不变量优先级

### 必须锁死（违反即拒绝）

1. Runtime 不知道组件
2. 不绕过 Action 修改状态
3. Component 不直接操作 Terminal
4. Input 与 Action 分离

### 架构级红线（违反需要架构评审）

5. Render 幂等性
6. 无隐式全局状态
7. 明确事件流阶段
8. 状态可枚举

### 设计原则（违反需要充分理由）

9. 按需 Tick（动画）
10. Spec 是一等公民

---

## 检查清单

在提交代码前，请确认：

- [ ] Runtime 没有导入 framework
- [ ] 所有状态变化通过 Action
- [ ] Render 函数无副作用
- [ ] Component 没有直接输出到 Terminal
- [ ] 所有状态可枚举、可快照
- [ ] Component 只处理 Action，不处理 RawInput
- [ ] 事件经过 Capture/Target/Bubble 三个阶段
- [ ] Builder API 和 Spec 等价
- [ ] 动画只在需要时 Tick
- [ ] 没有隐式状态（闭包、全局变量）

---

## 违反不变量的后果

| 严重程度 | 后果 |
|---------|------|
| 轻微 | 警告，要求修复 |
| 中等 | 拒绝合并，要求重构 |
| 严重 | 阻止发布，要求架构重新评审 |

---

## 检查工具

### 自动化检查

```bash
# scripts/check-invariants.sh

#!/bin/bash
set -e

echo "Checking invariants..."

# 1. Runtime 不导入 framework
if grep -r "framework" tui/runtime/ --include="*.go" | grep -v "_test.go" | grep -v "//"; then
    echo "❌ Runtime imports framework!"
    exit 1
fi

# 2. Component 不导入 platform
if grep -r "platform" tui/framework/component/ --include="*.go" | grep -v "_test.go" | grep -v "//"; then
    echo "❌ Component imports platform!"
    exit 1
fi

# 3. 检查全局变量
if grep -r "^var [A-Z]" tui/framework/component/ --include="*.go" | grep -v "_test.go"; then
    echo "❌ Global variables found!"
    exit 1
fi

echo "✅ All invariants satisfied!"
```

---

> **记住：这些不变量不是限制创造力，而是保护架构长期健康的护栏。**
> **当你觉得某个不变量阻碍了你，请先提出架构变更申请，而不是直接绕过。**

## 相关文档

- [ARCHITECTURE.md](./ARCHITECTURE.md) - 架构概览
- [BOUNDARIES.md](./BOUNDARIES.md) - 边界定义
- [ACTION_SYSTEM.md](./ACTION_SYSTEM.md) - Action 系统
- [STATE_MANAGEMENT.md](./STATE_MANAGEMENT.md) - 状态管理
- [AI_INTEGRATION.md](./AI_INTEGRATION.md) - AI 集成
