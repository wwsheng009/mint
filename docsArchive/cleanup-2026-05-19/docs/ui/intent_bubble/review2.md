你这个问题其实已经 **超出了 OptionGroup bug 修复层面**，而是触及 **Mint UI Runtime 的核心架构问题**：

> **Fiber-first 架构中缺少正式的“父子组件关系模型”和“组件通信机制”。**

当前 OptionGroup 只是 **第一个暴露这个问题的组件**。
如果不从架构上解决，将来 **Tabs / Form / Menu / Tree / Portal / Dialog / Layout / DragDrop** 全部都会重复踩坑。

你现在文档里提出的 **Closure 方案** 只是一个 **局部 workaround**，虽然简单，但**不能作为 Mint 的长期架构方案**。 

下面我从 **系统架构师角度**给你一个 **完整的 Mint 组件通信架构升级方案**。

---

# 一、问题的根本原因（Mint 架构层）

Mint 当前核心结构：

```
VNode (描述)
   ↓
Fiber (调度)
   ↓
Instance (持久组件)
```

架构特点：

| 层        | 生命周期         |
| -------- | ------------ |
| VNode    | 每次 render 创建 |
| Fiber    | 每次 render 创建 |
| Instance | 持久           |

这是典型 **Fiber-first runtime**。 

但有三个结构性缺陷：

### 1 缺少组件父子关系模型

Fiber 只有：

```
fiber.return   (parent)
fiber.sibling
```

但 **Instance 没有 parent/children**。

导致：

```
Instance A
   ?
Instance B
```

组件层 **没有树结构**。

---

### 2 VNode 生命周期过短

VNode 在 Fiber 创建后就丢弃：

```
VNode -> Fiber -> Instance
        ↑
VNode 已经消失
```

所以：

```
无法后期修改 children props
```

---

### 3 Instance 之间没有通信通道

现在组件通信只有三种方式：

```
1 props
2 closure
3 store
```

但缺少：

```
context
event bubble
parent lookup
```

所以 OptionGroup 才会出现：

```
子组件无法调用父组件
```

---

# 二、正确的 UI Runtime 架构（参考 React / Flutter / SwiftUI）

现代 UI Runtime 都有 **三套机制**：

| 机制                | 用途   |
| ----------------- | ---- |
| Parent/Child Tree | 结构   |
| Context           | 依赖注入 |
| Event Bubble      | 组件通信 |

React / Flutter / SwiftUI 都是这样。

Mint 目前 **三者都没有完整实现**。

---

# 三、Mint 架构升级目标

建议把 Mint 升级为 **三层通信模型**：

```
Component Communication Model

1️⃣ Parent Tree (结构关系)
2️⃣ Context (依赖注入)
3️⃣ Intent/Event (行为通信)
```

这样：

```
组件结构
      ↓
Context
      ↓
Intent / Event
```

---

# 四、第一层：建立 Instance 组件树（最关键）

这是 **必须做的架构升级**。

现在：

```
Fiber tree
```

应该增加：

```
Instance tree
```

设计：

```go
type ComponentInstance interface {

    Parent() ComponentInstance
    Children() []ComponentInstance

    AddChild(ComponentInstance)
}
```

Fiber 创建 Instance 时建立关系：

```
CreateFiber()

Parent Fiber
    ↓
CreateInstance()
    ↓
AddChild()
```

实现：

```go
type BaseInstance struct {
    parent   ComponentInstance
    children []ComponentInstance
}

func (b *BaseInstance) Parent() ComponentInstance {
    return b.parent
}

func (b *BaseInstance) AddChild(c ComponentInstance) {
    b.children = append(b.children, c)
}
```

Fiber 创建时：

```go
func mountFiber(f *Fiber) {

    inst := f.vnode.CreateInstance()

    if f.return != nil {
        parentInst := f.return.instance
        parentInst.AddChild(inst)
    }
}
```

得到：

```
Instance Tree

OptionGroupInstance
    ├ OptionInstance
    ├ OptionInstance
    └ OptionInstance
```

这是 **整个系统的基础能力**。

---

# 五、第二层：Context 机制（解决组合组件）

类似 React Context。

用于：

```
Form
OptionGroup
Tabs
Menu
List
Table
Theme
Router
```

设计：

```
Provider
Consumer
```

Context API：

```go
type ContextKey string

func Provide(ctx ContextKey, value any)

func UseContext(ctx ContextKey) any
```

Fiber 在运行时维护：

```
Context Stack
```

示例：

OptionGroup：

```
Provide(optionGroupContext, inst)
```

Option：

```
group := UseContext(optionGroupContext)
group.SelectOption(...)
```

这样：

```
Option 不需要 props
Option 不需要 closure
Option 不需要 store
```

组件自然通信。

---

# 六、第三层：Intent / Event 系统

你现在已经有：

```
Intent DSL
Store + Reducer
```

这是非常正确的方向。

但应该把它升级为：

```
UI Event System
```

设计：

```
Component Event Bubble
```

类似 DOM。

Example：

```
Option Click
   ↓
Emit Intent
   ↓
Bubble
   ↓
OptionGroup
```

API：

```go
func (c *Component) Emit(intent Intent)
```

运行时：

```
child -> parent -> parent
```

实现：

```go
func bubble(inst ComponentInstance, intent Intent) {

    for inst != nil {

        if inst.HandleIntent(intent) {
            return
        }

        inst = inst.Parent()
    }
}
```

OptionGroup：

```
HandleIntent(OptionSelectedIntent)
```

---

# 七、最终通信模型

Mint 最终通信模型：

```
                 ┌───────────────┐
                 │    Store      │
                 └──────┬────────┘
                        │
                 Intent / Action
                        │
                        ▼
┌─────────────── Component Tree ───────────────┐
│                                              │
│  Parent Instance                             │
│        ▲                                     │
│        │ Event Bubble                        │
│        │                                     │
│   Child Instance                             │
│                                              │
└──────────────────────────────────────────────┘
                        ▲
                        │
                    Context
```

组件通信：

| 类型            | 用途   |
| ------------- | ---- |
| Parent lookup | 组合组件 |
| Context       | 依赖注入 |
| Intent        | 事件通信 |
| Store         | 全局状态 |

---

# 八、OptionGroup 在新架构中的实现

Option：

```
OptionInstance
    ↓
UseContext(OptionGroupContext)
    ↓
Emit Intent
```

代码：

```go
func (o *OptionInstance) OnClick() {

    group := runtime.UseContext(optionGroupContext)

    runtime.Emit(
        OptionSelectIntent{
            Value: o.value,
        },
    )
}
```

OptionGroup：

```go
func (g *OptionGroupInstance) HandleIntent(i Intent) bool {

    switch v := i.(type) {

    case OptionSelectIntent:

        g.Select(v.Value)
        return true
    }

    return false
}
```

不需要：

```
closure
props callback
global registry
```

组件 **完全解耦**。

---

# 九、Portal / Layout / Overlay 也会一起解决

有 Instance Tree 后：

```
Portal
Overlay
Dialog
Tooltip
Drag
```

都可以实现。

因为：

```
组件逻辑树 ≠ 渲染树
```

Instance Tree = 逻辑树
Layout Tree = 渲染树

这是 React / Flutter 的标准架构。

---

# 十、最终 Mint Runtime 架构

推荐最终架构：

```
Mint Runtime

VNode (描述)
     ↓
Fiber (调度)
     ↓
Instance (组件)

Instance Tree
     ↓
Context System
     ↓
Intent / Event System
     ↓
Store / Reducer
```

能力：

```
✔ 组合组件
✔ 父子通信
✔ Context
✔ Event Bubble
✔ Portal
✔ 高级组件
```

---

# 十一、建议实施路线（现实可落地）

不要一次性大改。

分三步：

### Phase 1（2天）

建立：

```
Instance Tree
```

新增：

```
Parent()
Children()
```

---

### Phase 2（3天）

实现：

```
Context
```

---

### Phase 3（3天）

实现：

```
Intent Bubble
```

---

# 十二、结论（架构审查结果）

OptionGroup 问题本质是：

> **Mint 缺少组件通信基础设施**

最优方案不是 closure，而是：

```
1 Instance Tree
2 Context
3 Intent Bubble
```

Closure 只是 **临时修复方案**。 

---

如果你愿意，我可以 **继续帮你做一件非常有价值的事**：

我可以直接给你设计 **Mint Runtime 2.0 完整架构**（非常关键），包括：

* Fiber Runtime 完整结构
* Instance Tree
* Context 实现
* Intent Bubble
* Portal 设计
* Layout Tree
* 并发调度 Lane
* TUI 输入事件统一模型

并且给你 **完整 Go 代码框架（可直接落地）**。
这一套做完，Mint 架构会直接提升 **一个数量级**。

下面给你一套 **Mint Runtime 2.0 的完整架构设计 + 可落地 Go 框架**。
目标不是修补 OptionGroup，而是 **建立完整 UI Runtime 基础设施**，保证：

* Fiber-first
* 高性能 TUI
* 声明式组件
* 可组合组件
* 并发调度
* 可扩展 UI

这套设计参考：

* **React Fiber**
* **Flutter**
* **SwiftUI**

但针对 **Go + TUI runtime** 做了大量简化和优化。

---

# Mint Runtime 2.0 总体架构

核心结构：

```
VNode (声明式 UI)
      │
      ▼
Fiber Runtime (调度)
      │
      ▼
Component Instance (组件实例)
      │
      ▼
Instance Tree (逻辑组件树)
      │
      ├── Context System
      ├── Intent Bubble
      ├── Layout Tree
      └── Portal Layer
```

系统运行：

```
Input
  ↓
Intent
  ↓
Reducer / Store
  ↓
Render (VNode)
  ↓
Fiber Diff
  ↓
Instance Update
  ↓
Layout
  ↓
Paint
```

---

# 一、Fiber Runtime 完整结构

Fiber 是 Mint 的 **核心调度引擎**。

### Fiber Node

```go
package fiber

type Fiber struct {

    // 树结构
    Parent   *Fiber
    Child    *Fiber
    Sibling  *Fiber

    // 当前组件
    VNode    VNode
    Instance ComponentInstance

    // Diff
    Alternate *Fiber

    // 调度
    Lane      Lane

    // flags
    Flags     Flags
}
```

---

### Fiber Flags

```go
type Flags uint8

const (
    FlagPlacement Flags = 1 << iota
    FlagUpdate
    FlagDeletion
)
```

---

### Fiber 调度入口

```go
func ScheduleUpdate(root *Fiber, lane Lane) {

    root.Lane = lane

    workLoop(root)
}
```

---

### Fiber Work Loop

```go
func workLoop(root *Fiber) {

    work := root

    for work != nil {

        work = performUnitOfWork(work)
    }

    commitRoot(root)
}
```

---

### Fiber UnitOfWork

```go
func performUnitOfWork(f *Fiber) *Fiber {

    reconcile(f)

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

# 二、Instance Tree（组件树）

这是 **Mint 2.0 最关键升级**。

组件必须有逻辑树。

---

### BaseInstance

```go
package runtime

type BaseInstance struct {

    parent   ComponentInstance
    children []ComponentInstance
}

func (b *BaseInstance) Parent() ComponentInstance {
    return b.parent
}

func (b *BaseInstance) Children() []ComponentInstance {
    return b.children
}

func (b *BaseInstance) addChild(c ComponentInstance) {
    b.children = append(b.children, c)
}
```

---

### ComponentInstance 接口

```go
type ComponentInstance interface {

    Parent() ComponentInstance
    Children() []ComponentInstance

    OnMount()
    OnUpdate()
    OnUnmount()

    HandleIntent(Intent) bool

    Layout(LayoutContext)
    Paint(PaintContext)
}
```

---

### Fiber 创建 Instance

```go
func mountInstance(f *Fiber) {

    inst := f.VNode.CreateInstance()

    f.Instance = inst

    if f.Parent != nil && f.Parent.Instance != nil {

        parent := f.Parent.Instance

        parent.(*BaseInstance).addChild(inst)
    }
}
```

---

# 三、Context System

用于组件依赖注入。

类似 **React Context**。

---

### ContextKey

```go
type ContextKey string
```

---

### Context Provider

```go
type ContextProvider struct {
    values map[ContextKey]any
}
```

---

### Fiber Context Stack

```go
type FiberContext struct {

    stack []map[ContextKey]any
}
```

---

### Provide

```go
func Provide(key ContextKey, value any) {

    current := runtime.currentFiber

    ctx := current.context

    ctx[key] = value
}
```

---

### UseContext

```go
func UseContext(key ContextKey) any {

    f := runtime.currentFiber

    for f != nil {

        if v, ok := f.context[key]; ok {
            return v
        }

        f = f.Parent
    }

    return nil
}
```

---

### 示例：OptionGroup

```go
Provide(optionGroupContext, inst)
```

Option：

```go
group := UseContext(optionGroupContext).(*OptionGroupInstance)
```

---

# 四、Intent Bubble（组件事件系统）

组件之间通信。

类似 DOM event。

---

### Intent

```go
type Intent interface {
    Type() string
}
```

---

### Emit

```go
func Emit(inst ComponentInstance, intent Intent) {

    bubble(inst, intent)
}
```

---

### Bubble

```go
func bubble(inst ComponentInstance, intent Intent) {

    for inst != nil {

        handled := inst.HandleIntent(intent)

        if handled {
            return
        }

        inst = inst.Parent()
    }
}
```

---

### 示例

Option 点击：

```go
Emit(o, OptionSelectIntent{
    Value: o.value,
})
```

OptionGroup：

```go
func (g *OptionGroupInstance) HandleIntent(i Intent) bool {

    switch v := i.(type) {

    case OptionSelectIntent:

        g.Select(v.Value)
        return true
    }

    return false
}
```

---

# 五、Portal 设计

Portal 用于：

```
Dialog
Tooltip
Overlay
Dropdown
```

逻辑树 ≠ 渲染树。

---

### Portal VNode

```go
type PortalVNode struct {

    Target string
    Child  VNode
}
```

---

### Portal Instance

```go
type PortalInstance struct {

    BaseInstance

    target string
}
```

---

### Portal Render

Portal 不在当前 layout tree：

```
App
 ├ Layout
 └ PortalRoot
       └ Dialog
```

---

# 六、Layout Tree

Layout Tree 独立于 Instance Tree。

类似 Flutter RenderObject。

---

### LayoutNode

```go
type LayoutNode struct {

    Instance ComponentInstance

    Parent   *LayoutNode
    Children []*LayoutNode

    Bounds Rect
}
```

---

### LayoutContext

```go
type LayoutContext struct {

    Width  int
    Height int
}
```

---

### Layout

```go
func layoutNode(n *LayoutNode, ctx LayoutContext) {

    n.Instance.Layout(ctx)

    for _, c := range n.Children {

        layoutNode(c, ctx)
    }
}
```

---

# 七、并发调度 Lane

类似 React 18。

---

### Lane

```go
type Lane uint8

const (

    SyncLane Lane = 1 << iota
    InputLane
    RenderLane
    IdleLane
)
```

---

### 调度优先级

```
Input > Sync > Render > Idle
```

---

### Scheduler

```go
func scheduleUpdate(root *Fiber, lane Lane) {

    root.pendingLanes |= lane

    ensureRootScheduled(root)
}
```

---

### 抢占调度

```go
func performWork() {

    for {

        lane := getHighestPriorityLane()

        workLoop(lane)
    }
}
```

---

# 八、TUI 输入事件统一模型

统一：

```
Keyboard
Mouse
Focus
Clipboard
```

---

### InputEvent

```go
type InputEvent interface {

    Type() EventType
}
```

---

### EventType

```go
type EventType int

const (

    EventKey EventType = iota
    EventMouse
    EventFocus
)
```

---

### KeyEvent

```go
type KeyEvent struct {

    Key   string
    Ctrl  bool
    Alt   bool
    Shift bool
}
```

---

### MouseEvent

```go
type MouseEvent struct {

    X int
    Y int

    Button int
}
```

---

### Input → Intent

Input 不直接控制 UI。

转换为 Intent。

```
KeyPress
   ↓
Action
   ↓
Intent
   ↓
Store
```

示例：

```go
func handleKey(e KeyEvent) {

    switch e.Key {

    case "enter":

        EmitFocusIntent()
    }
}
```

---

# 九、最终 Mint Runtime 结构

目录建议：

```
runtime/

  fiber/
    fiber.go
    scheduler.go
    lanes.go

  instance/
    instance.go
    tree.go

  context/
    context.go

  intent/
    intent.go
    bubble.go

  layout/
    layout.go
    flex.go
    stack.go

  portal/
    portal.go

  input/
    event.go
    keyboard.go
    mouse.go

  paint/
    draw.go
```

---

# 十、运行流程（完整）

```
Input Event
     ↓
Intent
     ↓
Reducer
     ↓
Render VNode
     ↓
Fiber Diff
     ↓
Instance Update
     ↓
Layout Tree
     ↓
Paint
```

---

# 十一、Mint 2.0 能力

升级后 Mint 将支持：

```
✔ Fiber并发调度
✔ 声明式组件
✔ Context
✔ Intent Bubble
✔ Portal
✔ 高性能 Layout
✔ TUI 统一输入
✔ 复杂组件组合
```

OptionGroup / Form / Menu / Tabs / Tree / Dialog
都能优雅实现。

---

如果你愿意，我可以继续帮你 **把这套架构升级到真正工业级**：

我可以直接给你：

1️⃣ **完整 Fiber Diff 算法（O(n) key diff）**
2️⃣ **Flexbox Layout Engine（TUI版 Yoga）**
3️⃣ **FocusManager 完整设计**
4️⃣ **HitMap 鼠标命中系统**
5️⃣ **Portal Overlay Layer**
6️⃣ **完整 Go Runtime Skeleton（3000+ 行）**

这样 **Mint 可以直接成为 TUI 领域类似 React 的 runtime**。
