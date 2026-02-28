你这个问题本质上不是“能不能监听 release”，而是 **TUI 输入模型本身就不保证有 release 语义** ——这点和 GUI 完全不同。

先把结论说清楚：

👉 **大多数 Go TUI 库（比如 `tcell` / `bubbletea`）确实“弱化甚至没有可靠的 mouse release 事件”**
👉 按钮的 press/release 状态，必须用“状态机 + 推断”来做，而不是依赖底层事件完整性

下面我给你拆透（结合你现在在做的 TUI 架构，其实这里是关键设计点）。

---

# 一、为什么 TUI 很难有“release”？

终端协议（ANSI / xterm mouse）本质是这样的：

### 1. 不同模式行为不一致

终端鼠标支持几种模式：

* X10（只支持按下）
* VT200（部分支持）
* SGR（1006，现代终端）

👉 **只有 SGR 模式才“理论上”有 release**

---

### 2. 实际问题（重点）

即使是 SGR：

* release 可能丢失（终端 / ssh / tmux）
* 拖出窗口 → 不会有 release
* 终端切换 → 状态丢失
* Windows 兼容更差

👉 结论：

> **release 事件不可靠，不能作为核心逻辑依赖**

---

# 二、正确做法：按钮状态机（核心）

你不能写：

```go
if mouseDown { pressed = true }
if mouseUp   { pressed = false }
```

👉 这是 GUI 思维，在 TUI 会翻车。

---

## 正确模型（推荐）

定义按钮状态：

```go
type ButtonState int

const (
    Idle ButtonState = iota
    Hover
    Pressed
)
```

---

## 关键：用“推断”代替 release

### 状态转移逻辑：

```text
Idle
 ├─ mouse over → Hover
 └─ mouse down inside → Pressed

Pressed
 ├─ mouse move outside → Idle（取消）
 ├─ next frame no button → Idle（推断 release）
 └─ key/enter → trigger

Hover
 ├─ mouse down → Pressed
 └─ mouse leave → Idle
```

---

# 三、在 tcell 中怎么实现？

`tcell.EventMouse` 提供：

```go
btn := ev.Buttons()
```

关键点：

👉 `Buttons()` 是“当前状态”，不是事件！

---

## 正确处理方式（重要）

你要做：

```go
prev := lastButtons
curr := ev.Buttons()

if prev == 0 && curr != 0 {
    // press
}

if prev != 0 && curr == 0 {
    // release（推断）
}
```

👉 **这才是标准写法**

---

## 示例（简化版）

```go
type MouseTracker struct {
    last tcell.ButtonMask
}

func (m *MouseTracker) Update(ev *tcell.EventMouse) (pressed, released bool) {
    curr := ev.Buttons()

    if m.last == 0 && curr != 0 {
        pressed = true
    }
    if m.last != 0 && curr == 0 {
        released = true
    }

    m.last = curr
    return
}
```

---

# 四、真正工程级方案（推荐你用这个）

你现在做 TUI 架构，建议直接上：

## 👉 “交互状态机 + 帧驱动模型”（类似 React Fiber）

核心思想：

👉 **输入不是事件流，而是“状态快照流”**

---

## 1. 定义全局 Input Snapshot

```go
type InputState struct {
    MouseX, MouseY int
    Buttons        tcell.ButtonMask
}
```

---

## 2. 每帧计算

```go
prev := lastState
curr := newState

pressed  := prev.Buttons == 0 && curr.Buttons != 0
released := prev.Buttons != 0 && curr.Buttons == 0
```

---

## 3. Button 自己做判定

```go
func (b *Button) Update(input InputState) {
    inside := b.Contains(input.MouseX, input.MouseY)

    switch b.state {
    case Idle:
        if inside && input.Buttons != 0 {
            b.state = Pressed
        }

    case Pressed:
        if input.Buttons == 0 {
            if inside {
                b.OnClick()
            }
            b.state = Idle
        }
    }
}
```

---

# 五、关键设计建议（你现在这个阶段很重要）

### ❌ 不要做

* 依赖 mouse up 事件
* 依赖 click 事件（终端没有）
* 事件驱动按钮

---

### ✅ 要做

#### 1. 输入状态统一（你之前说的 Intent 系统可以接这里）

👉 Input → Intent → Component

---

#### 2. 按钮必须“自管理状态”

不要：

```go
onClick()
```

而是：

```go
button.Update(input)
```

---

#### 3. 支持“取消点击”

GUI 逻辑：

> 按下 → 拖走 → 松开 = 不触发

你必须实现：

```go
if pressed && !inside {
    cancel
}
```

---

# 六、如果你用的是 Bubble Tea

类似逻辑：

```go
case tea.MouseMsg:
    if msg.Type == tea.MouseLeftDown {
        pressed = true
    }
```

👉 但：

* `MouseLeftUp` 不稳定
* 推荐仍然做状态缓存

---

# 七、一句话总结

👉 **TUI 没有“可靠的 release”，只有“按钮状态变化”**

所以你要设计的是：

> **Input Snapshot → 状态机 → 推断 Press / Release → 组件自处理**

---

# 如果你要进阶（我强烈建议）

你现在这个问题，其实已经进入：

👉 **“交互系统设计层”而不是“事件处理层”**

我可以帮你往下推三步（这对你现在做 Fiber/TUI 很关键）：

1. 🔥 **统一 Input → Intent DSL（类型安全）**
2. 🔥 **Button / Drag / Select 的通用交互状态机**
3. 🔥 **支持 Lane 优先级（输入比渲染优先）**

你可以直接说一句：

👉「把 Button + Drag 做成统一交互模型」

我可以帮你做一版完整可落地的设计（含 Go 代码骨架）。
很好，这一步已经从“怎么处理事件”进入到**交互运行时设计**了。你这个方向是对的，而且可以直接做成你 TUI 的核心竞争力。

我不给你讲概念，直接给你一套**可落地的架构（Go 代码骨架 + 设计约束）**，重点解决三件事：

> ✅ Input → Intent（类型安全）
> ✅ Button / Drag / Select 统一状态机
> ✅ Lane（输入优先于渲染，可中断）

---

# 一、总体架构（你要实现的是这个）

```text
[ tcell / 输入源 ]
        ↓
[ Input Snapshot ]   ←（帧级状态）
        ↓
[ Intent Resolver ]  ←（类型安全转换）
        ↓
[ Interaction FSM ]  ←（Button / Drag / Select）
        ↓
[ Update Queue (Lane) ]
        ↓
[ Render / Commit ]
```

👉 核心思想：

> **事件 → 状态 → 意图 → 状态机 → 调度**

---

# 二、Input → Intent（类型安全 DSL）

你之前提到“避免字符串 field”，这里直接给你强类型设计。

---

## 1. Input Snapshot（统一输入层）

```go
type InputState struct {
    MouseX, MouseY int
    Buttons        uint8
    Keys           map[Key]bool

    // 辅助（由系统计算）
    PrevButtons uint8
}
```

---

## 2. Intent 定义（核心 DSL）

不要用 string，直接用类型：

```go
type Intent interface {
    isIntent()
}

type (
    IntentPress struct {
        X, Y int
        Button uint8
    }

    IntentRelease struct {
        X, Y int
        Button uint8
    }

    IntentMove struct {
        X, Y int
    }

    IntentKey struct {
        Key Key
    }
)

func (IntentPress) isIntent()   {}
func (IntentRelease) isIntent() {}
func (IntentMove) isIntent()    {}
func (IntentKey) isIntent()     {}
```

---

## 3. Resolver（输入 → 意图）

```go
func ResolveIntent(prev, curr InputState) []Intent {
    var intents []Intent

    // Move
    if prev.MouseX != curr.MouseX || prev.MouseY != curr.MouseY {
        intents = append(intents, IntentMove{curr.MouseX, curr.MouseY})
    }

    // Press
    if prev.Buttons == 0 && curr.Buttons != 0 {
        intents = append(intents, IntentPress{
            X: curr.MouseX, Y: curr.MouseY, Button: curr.Buttons,
        })
    }

    // Release（推断）
    if prev.Buttons != 0 && curr.Buttons == 0 {
        intents = append(intents, IntentRelease{
            X: curr.MouseX, Y: curr.MouseY, Button: prev.Buttons,
        })
    }

    return intents
}
```

👉 到这里，你已经**彻底摆脱 tcell 的不稳定事件流**

---

# 三、统一交互状态机（重点）

你不要为 Button / Drag / Select 分别写逻辑。

👉 正确做法：**一个统一 FSM + 不同策略**

---

## 1. 通用状态定义

```go
type InteractionState int

const (
    StateIdle InteractionState = iota
    StateHover
    StatePressed
    StateDragging
    StateSelecting
)
```

---

## 2. 通用上下文

```go
type InteractionContext struct {
    HotID    int // 当前 hover
    ActiveID int // 当前按下的组件

    StartX, StartY int
}
```

---

## 3. 核心：统一 Update

```go
func UpdateInteraction(
    ctx *InteractionContext,
    intents []Intent,
    hitTest func(x, y int) int,
) {
    for _, intent := range intents {
        switch e := intent.(type) {

        case IntentMove:
            id := hitTest(e.X, e.Y)
            ctx.HotID = id

            if ctx.ActiveID != 0 {
                // 判断是否进入 Drag
                if abs(e.X-ctx.StartX) > 1 || abs(e.Y-ctx.StartY) > 1 {
                    // 进入 Dragging
                }
            }

        case IntentPress:
            id := hitTest(e.X, e.Y)
            if id != 0 {
                ctx.ActiveID = id
                ctx.StartX = e.X
                ctx.StartY = e.Y
            }

        case IntentRelease:
            if ctx.ActiveID != 0 {
                if ctx.ActiveID == ctx.HotID {
                    // Click
                    DispatchClick(ctx.ActiveID)
                }
                ctx.ActiveID = 0
            }
        }
    }
}
```

---

## 4. 三种交互统一出来了

### Button

```text
Press → Release（同一组件） = Click
```

---

### Drag

```text
Press → Move（超过阈值） → Dragging
```

---

### Select（框选）

```text
Press → Move → Release（范围）
```

---

👉 关键 insight：

> **Button / Drag / Select 本质是同一个状态机的不同分支**

---

# 四、Lane 调度模型（输入优先）

你之前提 Fiber，这里给你一个“简化但够用”的版本。

---

## 1. Lane 定义

```go
type Lane int

const (
    LaneInput Lane = iota
    LaneUpdate
    LaneRender
)
```

---

## 2. Update Task

```go
type Task struct {
    Lane Lane
    Run  func()
}
```

---

## 3. 调度器

```go
type Scheduler struct {
    queues map[Lane][]Task
}

func (s *Scheduler) Schedule(t Task) {
    s.queues[t.Lane] = append(s.queues[t.Lane], t)
}
```

---

## 4. 执行策略（关键）

```go
func (s *Scheduler) Flush() {
    // 1. 输入优先
    for _, t := range s.queues[LaneInput] {
        t.Run()
    }

    // 2. 状态更新
    for _, t := range s.queues[LaneUpdate] {
        t.Run()
    }

    // 3. 渲染（可中断）
    for _, t := range s.queues[LaneRender] {
        t.Run()
    }

    s.clear()
}
```

---

## 5. 中断渲染（关键点）

在 render 中：

```go
func render() {
    for i := 0; i < len(nodes); i++ {
        if hasPendingInput() {
            return // 中断
        }
        draw(nodes[i])
    }
}
```

---

# 五、把三者串起来（完整运行流）

```go
loop {
    input := pollInput()

    scheduler.Schedule(Task{
        Lane: LaneInput,
        Run: func() {
            intents := ResolveIntent(prev, input)

            UpdateInteraction(ctx, intents, hitTest)
        },
    })

    scheduler.Schedule(Task{
        Lane: LaneRender,
        Run: func() {
            render()
        },
    })

    scheduler.Flush()

    prev = input
}
```

---

# 六、你这个架构的“质变点”

做到这里，你的 TUI 已经不是普通 TUI 了，而是：

### ✅ 类似 IMGUI + React Fiber 混合模型

你获得：

* 不依赖终端事件可靠性
* 所有交互统一抽象
* 可扩展（双击、长按、拖拽）
* 可中断渲染（性能关键）
* 完全可测试（Input → Intent 可回放）

---

# 七、可以继续进化的方向（下一步建议）

如果你要再往上走一层，我建议你做这三个升级：

---

## 🔥 1. Intent → Action（语义层）

```go
type Action interface {}

type ActionClick struct{ ID int }
type ActionDrag struct{ ID int; DX, DY int }
```

👉 Intent 是“物理”，Action 是“语义”

---

## 🔥 2. Fiber Node 挂 Interaction

```go
type Node struct {
    ID int
    Bounds Rect

    OnClick func()
    OnDrag  func(dx, dy int)
}
```

---

## 🔥 3. Lane 细化（React 那套）

```go
LaneDiscreteInput   // click
LaneContinuousInput // drag
LaneTransition      // 动画
```

---

# 最后一句话总结

👉 你现在要做的不是“处理按钮”，而是：

> **构建一个“交互运行时（Interaction Runtime）”**

---

如果你愿意，我可以下一步直接帮你做：

👉 **“Fiber Node + Interaction + Diff + O(n) 更新”的完整运行时设计（Go版）**

这一步会直接让你的 TUI 架构达到“框架级”。
好，这一层我们直接做“框架级运行时”。我会给你一套**能落地的最小正确实现（MVP Runtime）**，同时把扩展点（Fiber / Lane / Interaction）全部预留好。

目标：

> ✅ Fiber Node（可中断遍历）
> ✅ Interaction（统一交互）
> ✅ O(n) Diff（无 key & 有 key 两种）
> ✅ Scheduler（Lane 优先级）
> ✅ Commit（最小更新）

---

# 一、整体结构（你最终要跑的是这个）

```text
App(State)
   ↓
Render() → Virtual Tree
   ↓
Reconcile（Diff）
   ↓
Fiber Tree（可中断）
   ↓
Commit（最小变更）
   ↓
Screen
```

---

# 二、核心数据结构（最关键）

## 1. Element（声明式）

```go
type Element struct {
    Type  string
    Key   string
    Props Props
    Kids  []Element
}

type Props struct {
    X, Y, W, H int

    Text string

    OnClick func()
}
```

---

## 2. Fiber Node（运行时节点）

```go
type Fiber struct {
    // 树结构
    Parent  *Fiber
    Child   *Fiber
    Sibling *Fiber

    // 对应 element
    Type string
    Key  string
    Props Props

    // 状态
    StateNode any // 比如缓存 layout / text buffer

    // diff
    Alternate *Fiber
    EffectTag EffectTag
}
```

---

## 3. EffectTag（最小更新）

```go
type EffectTag int

const (
    Noop EffectTag = iota
    Placement
    Update
    Deletion
)
```

---

# 三、O(n) Diff（核心算法）

你必须实现两种模式：

---

## 1️⃣ 无 Key（简单快速）

```go
func reconcileChildren(wip *Fiber, elements []Element) {
    var prevSibling *Fiber
    oldFiber := wip.Alternate.Child

    for i, el := range elements {
        var newFiber *Fiber

        sameType := oldFiber != nil && el.Type == oldFiber.Type

        if sameType {
            newFiber = &Fiber{
                Type: el.Type,
                Props: el.Props,
                Alternate: oldFiber,
                EffectTag: Update,
            }
        } else {
            newFiber = &Fiber{
                Type: el.Type,
                Props: el.Props,
                EffectTag: Placement,
            }
        }

        if oldFiber != nil {
            oldFiber = oldFiber.Sibling
        }

        if i == 0 {
            wip.Child = newFiber
        } else {
            prevSibling.Sibling = newFiber
        }

        prevSibling = newFiber
    }
}
```

👉 复杂度：**O(n)**

---

## 2️⃣ 有 Key（可复用）

```go
func reconcileChildrenWithKey(wip *Fiber, elements []Element) {
    oldMap := map[string]*Fiber{}
    old := wip.Alternate.Child

    for old != nil {
        oldMap[old.Key] = old
        old = old.Sibling
    }

    var prev *Fiber

    for i, el := range elements {
        oldFiber := oldMap[el.Key]

        var newFiber *Fiber

        if oldFiber != nil && oldFiber.Type == el.Type {
            newFiber = &Fiber{
                Type: el.Type,
                Key: el.Key,
                Props: el.Props,
                Alternate: oldFiber,
                EffectTag: Update,
            }
            delete(oldMap, el.Key)
        } else {
            newFiber = &Fiber{
                Type: el.Type,
                Key: el.Key,
                Props: el.Props,
                EffectTag: Placement,
            }
        }

        if i == 0 {
            wip.Child = newFiber
        } else {
            prev.Sibling = newFiber
        }

        prev = newFiber
    }

    // 剩余的 old → 删除
    for _, f := range oldMap {
        f.EffectTag = Deletion
        deletions = append(deletions, f)
    }
}
```

---

# 四、Fiber 调度（可中断）

---

## 1. 全局状态

```go
var nextUnitOfWork *Fiber
var wipRoot *Fiber
var currentRoot *Fiber
var deletions []*Fiber
```

---

## 2. 工作循环（关键）

```go
func workLoop() {
    for nextUnitOfWork != nil {
        nextUnitOfWork = performUnitOfWork(nextUnitOfWork)

        if shouldYield() {
            return
        }
    }

    if wipRoot != nil {
        commitRoot()
    }
}
```

---

## 3. 单步执行（Fiber 核心）

```go
func performUnitOfWork(f *Fiber) *Fiber {
    // 1. 处理当前节点
    if f.Alternate == nil {
        // mount
    } else {
        // update
    }

    reconcileChildren(f, f.Props.Children)

    // 2. 返回下一个工作节点（DFS）
    if f.Child != nil {
        return f.Child
    }

    for f != nil {
        if f.Sibling != nil {
            return f.Sibling
        }
        f = f.Parent
    }

    return nil
}
```

---

# 五、Commit 阶段（最小更新）

---

```go
func commitRoot() {
    for _, f := range deletions {
        commitDeletion(f)
    }

    commitWork(wipRoot.Child)

    currentRoot = wipRoot
    wipRoot = nil
}
```

---

## commitWork

```go
func commitWork(f *Fiber) {
    if f == nil {
        return
    }

    parent := findParent(f)

    switch f.EffectTag {
    case Placement:
        draw(f)
    case Update:
        update(f)
    case Deletion:
        remove(f)
    }

    commitWork(f.Child)
    commitWork(f.Sibling)
}
```

---

# 六、Interaction 接入 Fiber（关键融合点）

---

## 1. 命中检测（HitTest）

```go
func hitTest(f *Fiber, x, y int) *Fiber {
    if f == nil {
        return nil
    }

    if inside(f.Props, x, y) {
        return f
    }

    if res := hitTest(f.Child, x, y); res != nil {
        return res
    }

    return hitTest(f.Sibling, x, y)
}
```

---

## 2. Interaction Context

```go
type Interaction struct {
    Hot    *Fiber
    Active *Fiber
}
```

---

## 3. 绑定事件

```go
func dispatchClick(f *Fiber) {
    if f.Props.OnClick != nil {
        f.Props.OnClick()
    }
}
```

---

## 4. 更新交互

```go
func updateInteraction(ctx *Interaction, intent Intent, root *Fiber) {
    switch e := intent.(type) {

    case IntentMove:
        ctx.Hot = hitTest(root, e.X, e.Y)

    case IntentPress:
        ctx.Active = hitTest(root, e.X, e.Y)

    case IntentRelease:
        target := hitTest(root, e.X, e.Y)
        if ctx.Active != nil && ctx.Active == target {
            dispatchClick(target)
        }
        ctx.Active = nil
    }
}
```

---

# 七、Scheduler（Lane）

---

```go
type Lane int

const (
    LaneInput Lane = iota
    LaneRender
)
```

---

## 调度

```go
func scheduleUpdate(root *Fiber) {
    wipRoot = &Fiber{
        Type: root.Type,
        Props: root.Props,
        Alternate: currentRoot,
    }

    nextUnitOfWork = wipRoot
}
```

---

## 主循环

```go
for {
    input := pollInput()

    // 1. 输入优先
    intents := ResolveIntent(prev, input)
    for _, i := range intents {
        updateInteraction(ctx, i, currentRoot)
    }

    // 2. 渲染（可中断）
    workLoop()

    prev = input
}
```

---

# 八、这个架构的关键价值

你现在拥有：

---

## ✅ 1. 真正的 O(n) UI 更新

* 不重绘全屏
* 只更新变化节点

---

## ✅ 2. 可中断渲染（TUI 非常关键）

* 输入不会卡
* 大列表不卡顿

---

## ✅ 3. 统一交互系统

* Button / Drag / Select 一套 FSM
* 不依赖终端事件

---

## ✅ 4. 可扩展能力

你可以很自然加：

* 双击（时间窗口）
* 长按（timer）
* 拖拽排序
* 虚拟列表

---

# 九、你下一步应该做的（非常关键）

我建议你按这个顺序落地：

---

## Step 1（必须）

👉 跑通最小版本：

* Element → Fiber
* reconcileChildren（无 key）
* commit（打印）

---

## Step 2

👉 接入 Interaction

---

## Step 3

👉 加 Lane + 中断

---

## Step 4（进阶）

👉 Text diff / buffer diff（减少 terminal flush）

---