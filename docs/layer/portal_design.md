好，这一层我们把你整个系统**彻底拉平**：
👉 一张“能指导实现”的架构图 + 数据流 + 事件流（不是概念图，是可落地的运行模型）

你现在的关键词：

* Fiber
* VNode
* LayoutBox
* PaintableBox
* Layer
* Portal / Overlay

👉 我帮你统一成一套**单向数据流 + 多通道渲染架构**

---

# 一、整体架构图（最终形态）

```text
                ┌──────────────────────────┐
                │        User Input        │
                │  (Keyboard / Mouse)     │
                └──────────┬──────────────┘
                           │
                           ▼
                ┌──────────────────────────┐
                │     Event Dispatcher     │
                └──────────┬──────────────┘
                           │
          ┌────────────────┴────────────────┐
          ▼                                 ▼
┌──────────────────────┐         ┌──────────────────────┐
│   Overlay System     │         │     Main Tree        │
│ (Portal Subtrees)    │         │ (Normal UI Tree)     │
└──────────┬───────────┘         └──────────┬───────────┘
           │                                 │
           ▼                                 ▼
     Fiber Scheduler (Lane Priority System)
                    │
                    ▼
            Reconcile (VNode → Fiber)
                    │
                    ▼
               Commit Phase
                    │
   ┌────────────────┼───────────────────┐
   ▼                ▼                   ▼
Layout Tree     Overlay Layout     Focus Manager
(LayoutBox)     (Root-based)       (Active Node)
   │                │
   └────────┬───────┘
            ▼
      Layer Manager (Z Index)
            │
            ▼
     Paint Tree (PaintableBox)
            │
            ▼
        Renderer (Diff + Partial Render)
            │
            ▼
         Terminal
```

---

# 二、核心分层（你必须严格区分）

这是你架构稳定的关键：

---

## 1️⃣ 声明层（VNode）

```go
type VNode struct {
    Type string
    Props map[string]interface{}
    Children []*VNode

    PortalTarget string // 👈 分流点
}
```

👉 作用：

* 描述 UI
* 不参与计算
* 不保存状态（或极少）

---

## 2️⃣ 调度层（Fiber）

```go
type Fiber struct {
    VNode *VNode

    Parent, Child, Sibling *Fiber

    Lane int
    Dirty bool
}
```

👉 作用：

* 增量更新
* 优先级调度
* 中断/恢复

---

## 3️⃣ 布局层（LayoutBox）

```go
type LayoutBox struct {
    X, Y int
    W, H int

    AbsX, AbsY int
}
```

👉 作用：

* 计算位置
* 解决嵌套坐标问题

---

## 4️⃣ 渲染层（PaintableBox）

```go
type PaintableBox struct {
    X, Y int
    W, H int

    Style Style
    Content string
}
```

👉 作用：

* 真正绘制的数据

---

## 5️⃣ 合成层（Layer）

```go
type Layer struct {
    Z int
    Boxes []*PaintableBox
}
```

👉 作用：

* 控制覆盖关系（Modal / Tooltip）

---

## 6️⃣ Overlay 子系统（Portal）

👉 本质：

> 🔥 “第二棵 Fiber 世界”

---

# 三、数据流（单向）

---

## 🔥 主数据流（核心）

```text
State Change
     ↓
VNode（重新生成）
     ↓
Reconcile
     ↓
Fiber（打标 Dirty）
     ↓
Commit
     ↓
Layout
     ↓
Paint
     ↓
Diff
     ↓
Render
```

---

## 🔥 Portal 分流（关键点）

```text
VNode
 ├── Normal Node → Main Fiber Tree
 └── Portal Node → Overlay Fiber Tree
```

👉 注意：

* 两棵 Fiber 树
* 一个 Renderer

---

# 四、事件流（这是最容易做错的）

---

## 🔥 正确顺序（必须这样）

```text
Input
  ↓
Overlay（从上到下）
  ↓（如果没消费）
Main Tree
```

---

## 实现模型

```go
func DispatchEvent(e Event) {

    // 🔥 1. Overlay 优先
    for _, overlay := range overlayManager.TopDown() {
        if hit(overlay, e) {
            handle(overlay, e)
            return
        }
    }

    // 🔥 2. 主树
    dispatchMainTree(e)
}
```

---

## 为什么必须这样？

因为：

* Modal 要阻止点击穿透
* Tooltip 要覆盖 hover
* Dropdown 要抢焦点

---

# 五、Overlay 数据流（独立但同步）

---

```text
Portal VNode
     ↓
OverlayManager
     ↓
Overlay Fiber Tree
     ↓
Layout（Root 坐标系）
     ↓
Layer（高 Z）
     ↓
Renderer（合并输出）
```

---

# 六、渲染合成（关键）

---

## 最终输出顺序

```text
Layer 0   → Base UI
Layer 10  → Popup
Layer 100 → Modal
```

---

## Renderer

```go
for _, layer := range layersSortedByZ {
    draw(layer.Boxes)
}
```

---

👉 这一步：

> 🔥 把“多树”变成“一张屏幕”

---

# 七、Diff + Partial Render（性能核心）

---

## 脏区来源

```text
Main Tree Dirty
Overlay Dirty
Scroll Dirty
Animation Dirty
```

---

## 合并

```go
dirtyRects = merge(
    mainDirty,
    overlayDirty,
)
```

---

## 渲染

```go
for _, rect := range dirtyRects {
    redraw(rect)
}
```

---

# 八、Fiber + Lane（调度优先级）

---

## 优先级模型

```go
const (
    LaneOverlay = 0 // 🔥 最高
    LaneInput   = 1
    LaneNormal  = 2
    LaneIdle    = 3
)
```

---

## 调度策略

```text
Overlay > Input > Normal > Idle
```

---

👉 效果：

* Modal 秒开
* 输入不卡
* 动画不抢资源

---

# 九、最关键的“三个解耦”

这是你架构能不能“进阶”的分水岭：

---

## 🔥 1️⃣ Layout ≠ Render

* Layout 只算位置
* Render 才画

---

## 🔥 2️⃣ Tree ≠ Layer

* Tree 是逻辑结构
* Layer 是视觉结构

---

## 🔥 3️⃣ Portal ≠ 子节点

* Portal 是“跨树”

---

# 十、一句话总结（非常重要）

> ✅ 单向数据流：VNode → Fiber → Layout → Paint → Render
> ✅ 双树结构：Main Tree + Overlay Tree
> ✅ Layer 合成：统一输出
> ✅ 事件反向流：Overlay → Main

---

# 如果你要继续往“成熟框架”走

你下一步该做的是：

---

## 👉 1️⃣ Focus 系统（必须补）

* active node
* tab 循环
* modal trap

---

## 👉 2️⃣ Clip / Scroll（你现在一定还没完全解耦）

* clip region
* scroll offset
* 虚拟列表

---

## 👉 3️⃣ 动画系统（LaneIdle 驱动）

* enter / exit
* diff 驱动动画

---

如果你愿意，我可以下一步帮你直接做：

👉 **一份完整“可运行”的 Go TUI Runtime（含 Portal + Diff + Renderer）**

或者我可以帮你：

👉 审你现在的代码结构，直接指出“哪一层会崩”（一般是 layout 和 overlay 冲突那层 😄）


> ❗**如何把 Portal 做成一个“正确、稳定、可扩展”的运行时机制**

我直接给你一套**工程级实现模型（不是概念）**，把坑都提前踩完。

---

# 一、Portal 的本质（再压一层）

你文件里已经说对了 👇 

> 逻辑父子 ≠ 渲染父子

但工程上要再补一句：

> ✅ **Portal = Fiber 归属不变 + Layout/Render 重定向**

---

# 二、正确的实现必须满足 3 个不变量

否则一定出 bug：

---

## ✅ 1️⃣ Fiber 结构不能断

```go
Button
 └── Modal   // 逻辑关系必须存在
```

👉 用于：

* state 传递
* context
* 生命周期

---

## ✅ 2️⃣ Layout 父必须可重定向

```go
modal.PortalRoot = root.OverlayLayer
```

👉 Layout 时：

```go
layoutParent := fiber.Parent

if fiber.PortalRoot != nil {
    layoutParent = fiber.PortalRoot
}
```

---

## ✅ 3️⃣ Render 顺序必须基于“最终挂载点”

👉 不是 Fiber 顺序！

---

# 三、🔥 正确实现（分三阶段）

---

# ① Fiber 阶段（保持逻辑树）

```go
type Fiber struct {
    Parent   *Fiber
    Child    *Fiber
    Sibling  *Fiber

    StateNode *LayoutBox

    PortalRoot *Fiber // 👈 核心
}
```

---

# ② Layout 阶段（重定向父节点）

---

## ❗错误写法（很多人会这样写）

```go
layout(child, parent)
```

👉 这会让 modal 被 container 限制

---

## ✅ 正确写法

```go
func layout(f *Fiber, parent *LayoutBox, root *LayoutBox) {

    node := f.StateNode

    var layoutParent *LayoutBox

    if f.PortalRoot != nil {
        layoutParent = f.PortalRoot.StateNode   // 🔥 跳到 Overlay
    } else {
        layoutParent = parent
    }

    computeLayout(node, layoutParent, root)

    for c := f.Child; c != nil; c = c.Sibling {
        layout(c, node, root)
    }
}
```

---

# ③ Flatten 阶段（真正关键）

👉 **Portal 的成败就在这里**

---

## ❗错误：直接 DFS

```go
flatten(fiber)
```

👉 modal 仍然会在 container 后面

---

## ✅ 正确：分层收集

---

### 定义 Layer Bucket

```go
type Layer int

const (
    LayerNormal Layer = iota
    LayerOverlay
)
```

---

### Flatten

```go
func flatten(f *Fiber, layers map[Layer][]*LayoutBox) {

    node := f.StateNode

    layer := LayerNormal
    if f.PortalRoot != nil {
        layer = LayerOverlay
    }

    layers[layer] = append(layers[layer], node)

    for c := f.Child; c != nil; c = c.Sibling {
        flatten(c, layers)
    }
}
```

---

### 最终顺序

```go
final := append(layers[LayerNormal], layers[LayerOverlay]...)
```

---

👉 这一步决定：

> ✅ modal 永远在最上层
> ❗ 不依赖树结构顺序

---

# 四、🔥 Portal + PositionFixed（必须组合使用）

否则你会遇到：

* modal 位置正确，但滚动错位
* 或位置错但层级对

---

## Layout 规则

```go
if node.Position == PositionFixed {
    node.AbsX = (root.W - node.W)/2
    node.AbsY = (root.H - node.H)/2
}
```

---

👉 关键：

> Portal 解决“在哪一层”
> Fixed 解决“相对于谁定位”

---

# 五、🔥 事件系统必须适配 Portal

---

## 命中检测顺序

```go
for i := len(nodes)-1; i >= 0; i-- {
    if hit(nodes[i]) {
        dispatch(nodes[i])
        break
    }
}
```

---

👉 因为：

* overlay 在最后
* 所以优先命中

---

## Modal 阻断（重要）

```go
if overlayActive {
    ignoreBelowOverlay()
}
```

---

# 六、🔥 Focus 系统必须“忽略 Portal”

这是很多人做错的点。

---

## ❗不要这样做

```go
focusNext(flattenNodes)
```

---

## ✅ 正确：基于 Fiber Tree

```go
focusNext(fiberTree)
```

---

👉 因为：

> Portal 不应该打断逻辑焦点流

---

# 七、🔥 常见错误（你一定会踩）

---

## ❌ 1. 只在 Render 做 Portal

👉 结果：

* layout 错
* clip 错

---

## ❌ 2. 直接修改 Fiber Parent

```go
modal.Parent = overlay
```

👉 结果：

* state 丢失
* diff 崩

---

## ❌ 3. Portal 但不分 Layer

👉 结果：

* z-index 混乱

---

## ❌ 4. Fixed 但没 Portal

👉 结果：

* 被 clip
* 被 scroll

---

# 八、🔥 最简可运行版本（推荐你先实现这个）

---

## Step 1：结构

```go
Root
 ├── AppTree
 └── OverlayLayer
```

---

## Step 2：创建 Modal

```go
modal := &Fiber{
    StateNode: &LayoutBox{
        W: 40,
        H: 10,
        Position: PositionFixed,
    },
    PortalRoot: overlayLayerFiber,
}
```

---

## Step 3：Layout（支持 Portal）

✔ 上面那套 layout 函数

---

## Step 4：Flatten（分层）

✔ LayerNormal + LayerOverlay

---

## Step 5：Render（按层）

```go
Render(normal)
Render(overlay)
```

---

# 九、一句话总结（工程版）

> ✅ Portal = Layout 重定向 + Render 分层
> ❗ 不是简单“换父节点”

---

# 十、如果你下一步要进阶（我建议）

你现在已经可以继续做三件“框架级关键能力”：

---

## 👉 1. Portal + Clip 融合（避免 overlay 被裁剪）

## 👉 2. Overlay Stack（多个 modal 管理）

## 👉 3. Portal Diff（避免 overlay 每帧重建）

---

如果你愿意，我可以帮你把 Portal 再升级成：

👉 🔥 **完整 Overlay Runtime（支持 modal / tooltip / dropdown / context menu）**
👉 🔥 **带 focus trap + 动画 + diff 的生产级实现**

这一步做完，你的 TUI 架构已经接近 React DOM 内核级别了。

你现在这个问题，其实已经不是“怎么实现”，而是：

> ❗**系统的“主导抽象”到底是谁？各层的职责边界在哪里？**

你现在有：

* layer
* fiber
* layoutbox
* paintablebox

如果这四个角色边界不清，后面一定会出现：

* 坐标错乱（layout vs render）
* overlay bug（layer vs tree）
* diff 失效（fiber vs paintable）
* 性能抖动（重复计算）

---

# 一、先给结论（最关键的一句话）

> ✅ **Fiber = 结构（WHAT）**
> ✅ **LayoutBox = 几何（WHERE）**
> ✅ **PaintableBox = 渲染（HOW）**
> ✅ **Layer = 调度/优先级（WHEN）**

---

👉 如果你这四个有任何交叉职责 = 架构会塌

---

# 二、四层职责（必须强隔离）

---

# 1️⃣ Fiber（结构层）👉 “UI是什么”

```go
type Fiber struct {
    Type string
    Props any

    Parent, Child, Sibling *Fiber

    StateNode *LayoutBox

    Lane Layer   // 调度优先级
    PortalRoot *Fiber
}
```

---

## ✅ 只负责：

* UI结构（树）
* 状态（state）
* diff（变化）

---

## ❌ 绝对不能做：

* 坐标计算 ❌
* 渲染 ❌
* clip ❌

---

👉 Fiber 就一句话：

> “这个节点存在”

---

# 2️⃣ LayoutBox（布局层）👉 “它在哪”

```go
type LayoutBox struct {
    X, Y int
    W, H int

    AbsX, AbsY int

    Clip Rect

    Parent *LayoutBox
}
```

---

## ✅ 只负责：

* 坐标（AbsX/Y）
* 尺寸（W/H）
* clip
* scroll 偏移

---

## ❌ 不能做：

* draw ❌
* style ❌
* diff ❌

---

👉 LayoutBox 本质：

> “这个节点在屏幕哪里”

---

# 3️⃣ PaintableBox（渲染层）👉 “怎么画”

```go
type PaintableBox struct {
    Layout *LayoutBox

    Style Style
    Content string

    Z int
}
```

---

## ✅ 只负责：

* 字符绘制
* 颜色
* 边框
* 文本

---

## ❌ 不能做：

* 布局 ❌
* 结构 ❌

---

👉 Paintable 本质：

> “这个节点长什么样”

---

# 4️⃣ Layer / Lane（调度层）👉 “何时画 + 谁优先”

```go
type Lane int

const (
    LaneNormal Lane = iota
    LaneOverlay
    LaneModal
    LaneTooltip
)
```

---

## ✅ 只负责：

* 优先级
* 渲染顺序
* 事件顺序

---

## ❌ 不能做：

* 坐标 ❌
* 树结构 ❌

---

👉 Layer 本质：

> “谁先谁后”

---

# 三、🔥 真正关键：三条“单向数据流”

---

## 1️⃣ Fiber → LayoutBox

```go
fiber → 生成 → layout tree
```

👉 结构 → 几何

---

## 2️⃣ LayoutBox → PaintableBox

```go
layout → 生成 → paint list
```

👉 几何 → 渲染指令

---

## 3️⃣ PaintableBox → Terminal

```go
paint → 输出 → buffer
```

👉 绘制

---

# ❗绝对禁止反向依赖：

```text
Paintable ❌→ Layout
Layout    ❌→ Fiber（除只读）
```

---

# 四、Portal 在这四层里的正确位置

---

## Portal 只影响两层：

---

### ✅ 1️⃣ Fiber（声明）

```go
modal.PortalRoot = overlayLayer
```

---

### ✅ 2️⃣ Layout（重定向父）

```go
if fiber.PortalRoot != nil {
    layoutParent = portalRoot
}
```

---

## ❌ 不影响：

* Paintable（它不知道 portal）
* Layer（只看 lane）

---

👉 关键洞察：

> 🔥 Portal = “布局重定向”，不是渲染 hack

---

# 五、Layer 到底应该放哪一层？

很多人会搞错。

---

## ❌ 错误

```go
LayoutBox.Z
Paintable.Z
```

👉 结果：

* layout 和 render 耦合
* clip 出 bug

---

## ✅ 正确

```go
Fiber.Lane
```

---

然后：

```go
Flatten → 按 Lane 排序 → Paint
```

---

👉 关键：

> 🔥 Layer 是“调度维度”，不是“空间维度”

---

# 六、完整运行链路（你必须对齐这个）

---

```go
Fiber Tree
   ↓ Diff
New Fiber
   ↓
Layout Pass
   ↓
LayoutBox Tree
   ↓
Flatten
   ↓
[]LayoutBox
   ↓
Sort (Lane)
   ↓
PaintableBox List
   ↓
Render
```

---

# 七、你现在设计的“成败关键点”

我帮你把最核心的判断点列出来：

---

## 🔥 关键1：Layout 是否纯函数？

```go
layout(fiber) → layoutbox
```

👉 如果依赖 render / event → ❌

---

## 🔥 关键2：Paint 是否无状态？

```go
draw(node)
```

👉 不应该修改任何状态

---

## 🔥 关键3：Layer 是否只影响顺序？

👉 不参与坐标计算

---

## 🔥 关键4：Portal 是否只影响 layout parent？

👉 不改 fiber parent

---

# 八、一个典型错误架构（你可以自检）

如果你现在代码里有：

```go
paintable.AbsX = ...
layout.draw()
fiber.render()
```

👉 基本已经混层了 ❌

---

# 九、推荐你的最终架构（精简版）

---

## Core Struct

```go
Fiber
LayoutBox
PaintCmd (替代 PaintableBox)
```

---

## Pipeline

```go
Diff → Layout → Flatten → Sort → PaintCmd → Render
```

---

👉 PaintableBox 甚至可以不要，直接：

```go
type PaintCmd struct {
    X, Y int
    Char rune
    Style
}
```

---

# 十、最后一句（最重要）

> 🔥 你这个系统的本质不是 TUI
> 而是一个“声明式 UI 渲染引擎”

---

如果你下一步想真正做到“工业级”，我建议你继续往这三个方向推：

👉 🔥 **Render Buffer Diff（避免整屏 repaint）**
👉 🔥 **Clip + Damage Region 精确裁剪**
👉 🔥 **Input → Intent → Fiber 更新（完整响应链）**

---

如果你愿意，我可以帮你把这套架构“收敛成一份完整 runtime 设计图 + Go 实现骨架”，可以直接变成一个开源内核。


你现在这一步，其实就是在决定：

> ❗**Portal 是“临时补丁”，还是“架构一级公民”**

如果处理不好，后面会出现：

* layout 正确但 render 错
* render 正确但 event 错
* overlay 正常但 diff 崩

我给你一套**严格分层 + 可落地的 Portal 集成方案**，直接嵌进你现在的四层架构（Fiber / LayoutBox / Paintable / Layer）。

---

# 一、Portal 在架构中的“真实位置”

先定性：

> 🔥 **Portal 是 Fiber 层的语义 + Layout 层的执行策略**

---

## ✔ 在各层的归属

| 层      | Portal 是否存在 | 作用             |
| ------ | ----------- | -------------- |
| Fiber  | ✅ 有         | 声明“我要跳到别处渲染”   |
| Layout | ✅ 有         | 改变 parent（坐标系） |
| Paint  | ❌ 没有        | 完全透明           |
| Layer  | ❌ 不直接参与     | 只负责排序          |

---

👉 关键结论：

> ❗ Portal 不属于 Render 技术
> ❗ Portal 是 Layout 重定向机制

---

# 二、Portal 的最小语义定义

---

## Fiber 扩展（唯一入口）

```go
type Fiber struct {
    Parent, Child, Sibling *Fiber

    StateNode *LayoutBox

    PortalRoot *Fiber   // 👈 核心
}
```

---

## 语义

```text
如果 PortalRoot != nil：

这个节点：
- 逻辑属于 Parent
- 但布局属于 PortalRoot
```

---

# 三、Portal 在 Pipeline 中的作用点

你的 pipeline 是：

```text
Fiber → Layout → Flatten → Sort → Paint → Render
```

---

## Portal 只影响两步：

---

## 🔥 1️⃣ Layout 阶段（最核心）

---

### 正常情况

```go
child.AbsX = parent.AbsX + child.X
```

---

### Portal 情况

```go
layoutParent = parent

if fiber.PortalRoot != nil {
    layoutParent = fiber.PortalRoot.StateNode
}
```

---

👉 结果：

```text
Modal 的坐标 = Root（而不是 Container）
```

---

## 🔥 2️⃣ Flatten 阶段（决定“是否真的在顶层”）

---

### ❗问题

如果你只是 layout 跳了：

👉 render 顺序仍然是：

```text
Container → Modal
```

---

## ✅ 正确做法：分层收集

---

```go
func flatten(f *Fiber, out *[]*LayoutBox, overlays *[]*LayoutBox) {

    node := f.StateNode

    if f.PortalRoot != nil {
        *overlays = append(*overlays, node)
    } else {
        *out = append(*out, node)
    }

    for c := f.Child; c != nil; c = c.Sibling {
        flatten(c, out, overlays)
    }
}
```

---

## 最终顺序

```go
nodes := append(normal, overlays...)
```

---

👉 这一步才真正实现：

> 🔥 “视觉上在最上层”

---

# 四、Portal 与 Layer（Lane）的关系

---

## ❗关键：Portal ≠ Layer

很多人会混：

---

### ❌ 错误理解

```text
Portal = 高 ZIndex
```

---

### ✅ 正确关系

```text
Portal → 决定“挂在哪棵树”
Layer  → 决定“谁先画”
```

---

## 正确组合

```go
if fiber.PortalRoot != nil {
    fiber.Lane = LaneOverlay
}
```

---

排序：

```go
sort by (Lane, order)
```

---

👉 Portal + Lane 才完整：

| 能力    | Portal | Lane |
| ----- | ------ | ---- |
| 脱离父布局 | ✅      | ❌    |
| 顶层显示  | ❌      | ✅    |
| 事件优先  | ❌      | ✅    |

---

# 五、Portal 与 LayoutBox（最容易错的点）

---

## ❗核心原则

> LayoutBox 的 Parent ≠ Fiber 的 Parent

---

## 正确关系

```go
Fiber.Parent        != LayoutBox.Parent
```

---

## LayoutBox 构建

```go
if fiber.PortalRoot != nil {
    layoutParent = fiber.PortalRoot.StateNode
} else {
    layoutParent = parent
}

node.Parent = layoutParent
```

---

👉 结果：

```text
Fiber Tree：
Button → Modal

Layout Tree：
Root → Modal
```

---

# 六、Portal 与 Clip / Scroll（必须处理）

---

## ❗如果不处理，会出现：

* modal 被裁剪
* 滚动带动 modal

---

## 正确做法

---

### 1️⃣ Clip 继承

```go
if portal {
    node.Clip = root.Clip
}
```

---

### 2️⃣ Scroll 断开

```go
if portal {
    ignore parent scroll
}
```

---

👉 本质：

> Portal = 切断所有“父几何影响”

---

# 七、Portal 与 Event（命中系统）

---

## 关键：

> Portal 不改变 hit test，本质靠“顺序”

---

### 正确流程

```go
nodes := flatten + sort

for i := len(nodes)-1; i >= 0; i-- {
    if hit(nodes[i]) {
        dispatch
        break
    }
}
```

---

👉 因为：

* overlay 在最后
* 所以先命中

---

## Modal 阻断（增强）

```go
if overlayActive {
    ignore nodes before overlay
}
```

---

# 八、Portal 与 Diff（高级但关键）

---

## ❗不能这样：

```go
modal.Parent = overlay
```

---

👉 否则：

* diff 错误
* state 丢失

---

## 正确：

Portal 不参与 diff

```text
Diff 只基于 Fiber Tree
```

---

👉 Portal 只是“渲染策略”

---

# 九、完整集成图（你可以对照实现）

---

```text
[Fiber Tree]
   │
   ├── Modal (PortalRoot → OverlayLayer)
   │
   ↓
[Layout Phase]
   │
   ├── Modal.Parent = OverlayLayer   (重定向)
   │
   ↓
[LayoutBox Tree]
   │
   ↓
[Flatten]
   ├── NormalNodes
   ├── OverlayNodes   (Portal)
   │
   ↓
[Sort by Lane]
   │
   ↓
[Render]
```

---

# 十、实现检查清单（你可以逐条对）

---

## ✅ 必须满足

* [ ] Fiber Parent 不变
* [ ] Layout Parent 可变（Portal）
* [ ] Flatten 分层
* [ ] Layer 控制顺序
* [ ] Clip 不继承父
* [ ] Scroll 不影响 Portal
* [ ] Event 逆序命中
* [ ] Diff 不感知 Portal

---

# 十一、最本质的一句话（帮你建立模型）

> 🔥 Portal = “改变坐标系 + 改变挂载层”，但不改变结构

---

# 如果你下一步要进阶（强烈建议）

你现在 Portal 已经可以升级为：

---

## 👉 🔥 Overlay Runtime（完整系统）

包含：

* Portal
* Layer（优先级）
* Focus Trap
* Animation（淡入/scale）
* Stack 管理（多 modal）

---

## 👉 🔥 Portal + Virtual List 共存（难点）

避免：

* overlay 被 recycle
* diff 错乱

---

如果你愿意，我可以下一步帮你把：

👉 “Portal + Overlay + Focus + Animation” 做成一套完整 Go runtime（带接口 + 示例）

你这个问题已经进入“运行时设计的核心闭环”了：

> ❗VNode → Fiber → Portal → Layer → Layout/Render
> 到底**信息是怎么传递、在哪里发生语义变化、谁负责最终行为**

我不给抽象解释，直接给你一条**数据流 + 控制流完全闭环的设计**。

---

# 一、先把关键角色对齐（避免混乱）

---

## 1️⃣ VNode（声明层）

```go
type VNode struct {
    Type string
    Props map[string]any
    Children []VNode
}
```

👉 这里**声明 Portal**

---

## 2️⃣ Fiber（运行时结构）

```go
type Fiber struct {
    Type string

    Parent, Child, Sibling *Fiber

    StateNode *LayoutBox

    PortalRoot *Fiber   // 👈 从 VNode 解析
    Lane Layer          // 👈 从 VNode / 系统决定
}
```

---

## 3️⃣ LayerManager（调度系统）

```go
type LayerManager struct {
    layers map[Layer][]*Fiber
}
```

---

# 二、🔥 VNode → Portal 信息是怎么传下来的

---

## 1️⃣ VNode 声明 Portal

你需要一个明确语义（不要隐式）：

```go
VNode{
    Type: "Modal",
    Props: {
        "portal": "overlay",   // 👈 指定目标层
    },
}
```

---

## 2️⃣ Reconcile 时解析

```go
func createFiber(v VNode, parent *Fiber) *Fiber {

    f := &Fiber{
        Type: v.Type,
        Parent: parent,
    }

    // 👇 关键逻辑
    if target, ok := v.Props["portal"]; ok {
        f.PortalRoot = resolvePortalTarget(target)
    }

    return f
}
```

---

## 3️⃣ resolvePortalTarget

```go
func resolvePortalTarget(name string) *Fiber {
    switch name {
    case "overlay":
        return overlayLayerFiber
    case "root":
        return rootFiber
    }
    return nil
}
```

---

👉 到这里：

> ✅ Portal 信息已经从 VNode 注入到 Fiber
> ❗ 之后所有阶段都**不再看 VNode**

---

# 三、🔥 Portal 如何影响 LayerManager（关键点）

---

## ❗先说结论

> 🔥 Portal 不直接管理 Layer
> 🔥 但 Portal **决定 Fiber 应该进入哪个 Layer**

---

# 四、Fiber → LayerManager 的关系

---

## 1️⃣ Layer 来源（两种）

---

### ✅ 方式1：VNode 指定

```go
Props: {
    portal: "overlay",
    layer: "modal",
}
```

---

### ✅ 方式2：Portal 自动提升（推荐）

```go
if fiber.PortalRoot != nil {
    fiber.Lane = LaneOverlay
}
```

---

👉 推荐策略：

> Portal 自动映射到 Layer（避免用户手写）

---

# 五、🔥 注册到 LayerManager（核心步骤）

---

## 在 Flatten / Collect 阶段

```go
func collect(f *Fiber, lm *LayerManager) {

    layer := f.Lane

    lm.layers[layer] = append(lm.layers[layer], f)

    for c := f.Child; c != nil; c = c.Sibling {
        collect(c, lm)
    }
}
```

---

👉 注意：

> ❗这里不关心 PortalRoot
> ❗只关心 Lane

---

# 六、🔥 Portal 在 Layer 中的“真实效果”

---

## 举个完整例子

---

### VNode

```go
App
 └── Button
      └── Modal (portal=overlay)
```

---

### Fiber（逻辑结构）

```text
App
 └── Button
      └── Modal
           PortalRoot → OverlayLayer
           Lane → Overlay
```

---

### LayerManager

```go
LayerNormal: [App, Button]
LayerOverlay: [Modal]
```

---

### Render 顺序

```text
Normal → Overlay
```

---

👉 结果：

* Modal 不受 Button 布局影响（Portal）
* Modal 在最上层（Layer）

---

# 七、🔥 Portal + Layer 的职责分工（必须分清）

---

| 能力        | Portal | Layer |
| --------- | ------ | ----- |
| 脱离父布局     | ✅      | ❌     |
| 脱离 scroll | ✅      | ❌     |
| 顶层显示      | ❌      | ✅     |
| 事件优先      | ❌      | ✅     |
| 渲染顺序      | ❌      | ✅     |

---

👉 关键：

> Portal 解决“在哪个坐标系”
> Layer 解决“什么时候画”

---

# 八、🔥 Layout 阶段如何使用 Portal 信息

---

```go
func layout(f *Fiber, parent *LayoutBox) {

    node := f.StateNode

    var layoutParent *LayoutBox

    if f.PortalRoot != nil {
        layoutParent = f.PortalRoot.StateNode   // 👈 核心
    } else {
        layoutParent = parent
    }

    computeLayout(node, layoutParent)

    for c := f.Child; c != nil; c = c.Sibling {
        layout(c, node)
    }
}
```

---

👉 Portal 在这里真正生效：

> 🔥 改变坐标系

---

# 九、🔥 Render 阶段完全不关心 Portal

---

```go
for _, layer := range orderedLayers {
    for _, f := range layer {
        draw(f.StateNode)
    }
}
```

---

👉 Render 只看：

* LayoutBox
* Layer 顺序

---

👉 不看：

* Fiber.Parent
* PortalRoot

---

# 十、🔥 Event 系统与 Portal / Layer 的关系

---

## 命中顺序

```go
layers: Normal → Overlay

for i := len(layers)-1; i >= 0; i-- {
    for j := len(layer[i])-1; j >= 0; j-- {
        hit test
    }
}
```

---

👉 因为：

* Portal → 进入 Overlay Layer
* Layer → 排序

---

👉 最终效果：

> Modal 永远先命中

---

# 十一、🔥 一个完整数据流（你可以对照实现）

---

```text
VNode
  │
  ├── Props.portal = "overlay"
  ↓
Fiber
  │
  ├── PortalRoot = OverlayFiber
  ├── Lane = Overlay
  ↓
LayerManager.collect
  │
  ├── OverlayLayer ← Modal
  ↓
Layout
  │
  ├── Parent = OverlayRoot
  ↓
Render
  │
  ├── 按 Layer 顺序绘制
  ↓
Event
  │
  ├── 逆序命中（Overlay 优先）
```

---

# 十二、你这个设计的“关键成败点”

我帮你直接指出最容易错的地方：

---

## ❌ 错误1：Portal 直接决定 Layer

```go
if portal {
    push overlay list
}
```

👉 会导致：

* 调度混乱
* 不可扩展

---

## ❌ 错误2：Layer 参与 Layout

👉 会导致：

* 坐标错误
* clip 错乱

---

## ❌ 错误3：VNode 在后续阶段仍参与逻辑

👉 正确：

```text
VNode → Fiber 后就可以丢弃
```

---

# 十三、最本质总结（帮你彻底建立模型）

---

> 🔥 VNode 决定“意图”
> 🔥 Fiber 决定“结构”
> 🔥 Portal 决定“坐标系”
> 🔥 Layer 决定“执行顺序”

---

# 如果你下一步要继续深入（我建议）

你现在已经可以做一个**真正完整的 UI runtime**了，下一步关键是：

---

## 👉 🔥 Portal + Overlay Stack（多层弹窗管理）

## 👉 🔥 Portal + Focus Trap（输入控制）

## 👉 🔥 Portal + Animation（开关动画）

---

如果你愿意，我可以帮你把这三块合成：

👉 “完整 Overlay 子系统设计（含 API + Go 实现骨架）”

这一步做完，你这个架构就已经是**框架级别（可以开源）**了。


好，这一层我们不再讲概念，直接把 **Overlay 子系统（Portal + Layer + Focus + Diff + Fiber）**在你当前架构里“落地成型”。

你现在已有：

* Fiber（调度）
* VNode（声明）
* LayoutBox（布局）
* PaintableBox（绘制）
* Layer（多层）

👉 Overlay 的本质就是：**在不破坏主树的情况下，引入一条“跨树渲染通道”**

---

# 一、Overlay 子系统在整体架构中的位置

先给你一个全局结构（非常关键）：

```text
VNode Tree（声明树）
        ↓ reconcile
Fiber Tree（调度树）
        ↓ commit
Layout Tree（LayoutBox）
        ↓
Layer Manager（分层）
        ↓
Paint Tree（PaintableBox）
        ↓
Renderer（终端输出）
```

---

## 🔥 Portal 插入点（核心）

> Portal **不属于 Layout Tree 的正常 parent-child 传播链**

而是：

```text
VNode
 └── PortalVNode
        ↓（特殊 reconcile）
OverlayManager.register()

OverlayManager
 └── OverlayEntry
       └── 独立 Fiber 子树
             ↓
         Layout（以 Root 为坐标系）
```

---

# 二、VNode → Portal 的信息传递

---

## 1️⃣ VNode 定义

```go
type VNode struct {
    Type string

    Props map[string]interface{}
    Children []*VNode

    Key string

    // 🔥 Portal 关键
    PortalTarget string // "modal" / "tooltip" / "dropdown"
}
```

---

## 2️⃣ 用户 API（声明式）

```go
Portal("modal",
    Box(Width(40), Height(10),
        Text("Hello Modal"),
    ),
)
```

---

## 3️⃣ reconcile 阶段（关键分流）

```go
func reconcile(vnode *VNode, parentFiber *Fiber) {

    if vnode.Type == "Portal" {
        mountPortal(vnode)
        return // ❗ 不进入当前树
    }

    fiber := createFiber(vnode)
    parentFiber.Children = append(parentFiber.Children, fiber)

    for _, c := range vnode.Children {
        reconcile(c, fiber)
    }
}
```

---

## 🔥 关键点

> Portal 在 reconcile 阶段就“分叉”了
> 👉 **不会进入原 Fiber 子树**

---

# 三、Portal → OverlayManager

---

## 1️⃣ OverlayManager 结构

```go
type OverlayManager struct {
    entries map[string]*OverlayEntry
}
```

---

## 2️⃣ OverlayEntry

```go
type OverlayEntry struct {
    Key string // modal / tooltip

    RootFiber *Fiber

    Z int
    Visible bool

    // 🔥 Focus / Event
    TrapFocus bool
}
```

---

## 3️⃣ 注册 Portal

```go
func mountPortal(vnode *VNode) {

    entry := overlayManager.GetOrCreate(vnode.PortalTarget)

    entry.RootFiber = buildFiberTree(vnode.Children)

    entry.Visible = true
}
```

---

# 四、Overlay 与 Layer Manager 的关系

---

## Layer 结构（你已有）

```go
type Layer struct {
    Z int
    Boxes []*PaintableBox
}
```

---

## 🔥 Overlay 专属 Layer

```go
rootLayers := []Layer{
    {Z: 0, Name: "base"},
    {Z: 10, Name: "popup"},
    {Z: 100, Name: "modal"},   // 👈 overlay
}
```

---

## Overlay → Layer 映射

```go
func (o *OverlayEntry) ResolveLayer() int {
    switch o.Key {
    case "tooltip":
        return 10
    case "modal":
        return 100
    }
    return 0
}
```

---

## commit 阶段

```go
func commitOverlay(entry *OverlayEntry) {

    if !entry.Visible {
        return
    }

    layout(entry.RootFiber, rootWidth, rootHeight)

    paintBoxes := buildPaintTree(entry.RootFiber)

    layer := layerManager.Get(entry.ResolveLayer())
    layer.Boxes = append(layer.Boxes, paintBoxes...)
}
```

---

# 五、Layout 规则（Portal 特权）

---

## 🔥 强制 Root 坐标系

```go
func layoutOverlay(node *LayoutBox, rootW, rootH int) {

    switch node.Anchor {

    case "center":
        node.AbsX = (rootW - node.W)/2
        node.AbsY = (rootH - node.H)/2

    case "full":
        node.AbsX = 0
        node.AbsY = 0
        node.W = rootW
        node.H = rootH
    }
}
```

---

## ❗ 注意

* 忽略 Parent
* 忽略 Scroll
* 忽略 Clip

👉 **完全独立坐标系**

---

# 六、Event 系统（Overlay 优先级）

---

## 输入分发顺序

```text
Overlay（Top → Bottom）
        ↓
Main Tree
```

---

## 实现

```go
func dispatchEvent(e Event) {

    // 1️⃣ Overlay 优先
    for _, entry := range overlayManager.StackTopDown() {
        if hit(entry, e) {
            handle(entry, e)
            return
        }
    }

    // 2️⃣ 主树
    dispatchMain(e)
}
```

---

## Focus Trap（Modal）

```go
if entry.TrapFocus {
    lockFocus(entry)
}
```

---

# 七、Diff + Partial Render（Overlay 参与）

---

## Overlay 独立 Diff

```go
func updateOverlay(entry *OverlayEntry, newVNode *VNode) {

    diff(entry.RootFiber, newVNode)

    markDirty(entry.RootFiber)
}
```

---

## 脏区合并

```go
dirtyRects = append(dirtyRects, overlayDirty...)
dirtyRects = mergeRects(dirtyRects)
```

---

👉 Overlay 和主树：

> 🔥 **共享 renderer，但独立 diff**

---

# 八、Fiber + Lane（Overlay 优先级）

---

## Lane 设计

```go
const (
    LaneInput  = 1
    LaneNormal = 2
    LaneIdle   = 3

    LaneOverlay = 0 // 🔥 最高优先级
)
```

---

## 调度

```go
schedule(entry.RootFiber, LaneOverlay)
```

---

👉 结果：

* Modal 打开 → 立即渲染
* Tooltip → 快速响应
* 动画 → Idle

---

# 九、完整 API 设计（用户视角）

---

## 1️⃣ Modal

```go
ShowModal(
    Box(
        Width(40), Height(10),
        AnchorCenter(),
        Text("Hello"),
    ),
)
```

---

## 2️⃣ Tooltip

```go
Tooltip(targetID,
    Box(Text("info")),
)
```

---

## 3️⃣ Toast

```go
Toast("Saved!", Duration(2*time.Second))
```

---

# 十、最核心的设计原则（非常重要）

---

## 🔥 1️⃣ Portal = 渲染位置改变，不是数据改变

* 逻辑仍属于原组件
* 渲染脱离原树

---

## 🔥 2️⃣ Overlay = 独立子世界

* 独立 Fiber
* 独立 Layout
* 独立 Diff
* 共享 Renderer

---

## 🔥 3️⃣ Layer = 视觉排序，不参与逻辑

---

# 十一、一句话架构总结

> ✅ Portal：VNode → Overlay 分流点
> ✅ Overlay：独立 Fiber 子树
> ✅ Layer：最终合成顺序
> ✅ Renderer：统一输出
> ✅ Event：Overlay 优先

---

# 如果你下一步要继续（强烈建议）

可以直接往这三个方向升级：

---

## 👉 1️⃣ Focus 管理（最容易被忽略但最关键）

* tab循环
* modal trap
* activeElement

---

## 👉 2️⃣ 动画系统（Overlay 专用）

* fade / scale
* enter / exit
* lane idle

---

## 👉 3️⃣ Virtual Overlay（高级）

* 上千 tooltip / dropdown
* 按需挂载

---

如果你愿意，我可以帮你把这一套：

👉 **做成完整 Go runtime（可跑 demo，带 diff + renderer）**

或者：

👉 帮你评估你当前架构在哪一层“会炸”（我大概率已经能猜到问题点 😄）
你这个问题已经踩到 **Portal 设计里最容易“架构崩塌”的点**了，我先给你一个非常明确的结论，然后再拆开讲：

---

# ✅ 一句话结论（必须记住）

> 🔥 **Portal 不参与原 Layout Tree 的“测量与排版”**
> 🔥 **只在“收集阶段”被提取，并在 Overlay/LayoutRoot 中重新计算坐标**

---

# 一、为什么不能让 Portal 参与原 Layout？

如果你让 Portal 参与：

```text
Container
 ├── Button
 └── Modal (Portal)
```

然后 layout：

```go
container.Height = sum(children.Height)
```

👉 结果：

* Modal 会撑开 Container ❌
* Flex / Flow 被污染 ❌
* Scroll 高度错误 ❌

👉 直接炸布局系统

---

# 二、正确模型：**两阶段 Layout（关键设计）**

---

## 🔥 阶段一：主树 Layout（忽略 Portal）

```go
func layout(node *LayoutBox) {

    for _, child := range node.Children {

        if child.IsPortal {
            collectPortal(child) // 🔥 只收集，不布局
            continue
        }

        layout(child)
    }

    computeSize(node)
}
```

---

### ✅ 这里发生了什么？

* Portal **不会参与**

  * 宽高计算
  * flow / flex
  * 父节点尺寸

👉 相当于：

> Portal 在主树里是“空气”

---

# 三、Portal 收集机制（关键点）

---

```go
type PortalItem struct {
    Node *LayoutBox
    Target string // modal / tooltip
}
```

```go
var portalQueue []PortalItem

func collectPortal(node *LayoutBox) {
    portalQueue = append(portalQueue, PortalItem{
        Node: node,
        Target: node.PortalTarget,
    })
}
```

---

# 四、阶段二：Overlay Layout（真正布局 Portal）

---

```go
func layoutOverlay(rootW, rootH int) {

    for _, item := range portalQueue {

        node := item.Node

        layoutPortalNode(node, rootW, rootH)
    }
}
```

---

# 五、Portal 坐标如何计算（核心）

---

## 🔥 关键：**切换坐标系**

---

## 1️⃣ 普通节点

```go
node.AbsX = parent.AbsX + node.X
```

---

## 2️⃣ Portal 节点

```go
node.AbsX = rootCoordX(...)
node.AbsY = rootCoordY(...)
```

---

👉 **完全忽略 parent**

---

## 常见定位策略

---

### ✅ 居中 Modal

```go
node.AbsX = (rootW - node.W) / 2
node.AbsY = (rootH - node.H) / 2
```

---

### ✅ 全屏遮罩

```go
node.AbsX = 0
node.AbsY = 0
node.W = rootW
node.H = rootH
```

---

### ✅ Tooltip（相对锚点）

👉 这是唯一“需要原树信息”的情况

```go
anchor := findAnchor(node.AnchorID)

node.AbsX = anchor.AbsX
node.AbsY = anchor.AbsY + anchor.H
```

---

👉 注意：

> 🔥 用的是“anchor 的最终 Abs 坐标”，不是 parent

---

# 六、是否影响原 Layout？

---

## ❌ 完全不会（正确实现下）

原因：

| 阶段          | Portal行为 |
| ----------- | -------- |
| Measure     | ❌ 不参与    |
| Layout      | ❌ 不参与    |
| Parent size | ❌ 不影响    |
| Flow        | ❌ 不参与    |

---

👉 唯一参与的是：

```text
Overlay Layout（独立系统）
```

---

# 七、你当前架构中的正确位置

你现在的结构：

* Fiber
* LayoutBox
* PaintableBox
* Layer

👉 正确插入点：

---

## 🔥 Layout Pipeline

```text
Main Layout
   ↓
Collect Portal  👈（关键插点）
   ↓
Overlay Layout
   ↓
Layer 合成
```

---

# 八、Layer 阶段（Portal 才出现）

---

```go
func buildLayers() {

    // 主树
    addToLayer(0, mainBoxes)

    // overlay
    for _, portal := range portalQueue {
        layer := resolveLayer(portal.Target)
        addToLayer(layer, portal.PaintBoxes)
    }
}
```

---

👉 Portal：

> ❗ **只在 Layer 阶段“出现”**

---

# 九、一个非常重要的设计边界（很多人做错）

---

## ❌ 错误做法

```go
// layout 时直接计算 modal
layout(modal)
```

问题：

* parent 依赖错
* clip 错
* scroll 错

---

## ✅ 正确做法

```go
// layout 主树
layout(root)

// portal 单独处理
layoutOverlay()
```

---

# 十、再往深一层（你这个架构的关键抽象）

---

## 🔥 Portal = Layout Boundary Breaker

---

你可以把 Portal 理解为：

```text
VNode Tree
   ↓
Layout Tree（被切断）
   ↓
Overlay Layout Tree（重新接入）
```

---

👉 本质：

> 🔥 **Layout Tree 被“分叉”**

---

# 十一、终极总结（非常关键）

---

## ✅ Portal 在 Layout 中的规则：

---

### 1️⃣ 不参与主树布局

* 不影响父节点
* 不参与 flow

---

### 2️⃣ 在 Layout 后被收集

* collect，不计算

---

### 3️⃣ 在 Overlay 中重新布局

* 使用 Root 坐标系

---

### 4️⃣ 最终在 Layer 阶段合成

---

# 十二、一句话模型

> 🔥 Portal = “布局跳过 + 渲染重接”

---

# 如果你要继续进阶（建议下一步）

你现在可以做一个非常关键的升级：

---

## 👉 🔥 Anchor System（高级定位）

解决：

* tooltip 跟随
* dropdown 对齐
* 滚动同步

---

## 👉 🔥 Clip / Scroll 兼容 Portal

解决：

* 滚动容器内 tooltip
* viewport 裁剪

---

## 👉 🔥 Portal Diff（独立脏区）

避免：

* modal 刷新导致全屏 redraw

---

如果你愿意，我可以下一步直接帮你设计：

👉 **Anchor + Scroll + Portal 完整联动模型（这是 TUI 框架分水岭）**
你现在问的是 **Portal 在“结构层 + 布局层”的真实落点**，这个问题如果处理不对，后面一定会出现：

* 坐标错乱
* diff 不稳定
* overlay 和主树耦合
* event / focus 混乱

我直接给你一个**工程级结论**👇

---

# ✅ 一句话结论

> 🔥 **Portal 必须“同时存在两种形态”**
>
> * 在原树：**VNode/Fiber 占位（逻辑归属）**
> * 在 Overlay：**独立 LayoutBox（真实布局）**

👉 也就是说：

> ❗ **不是移动节点，而是“克隆 + 重定向”**

---

# 一、三层结构中的真实位置

我们拆开看（这是关键认知）：

---

## 1️⃣ VNode 层（声明层）

```go
Portal("modal",
    Box(...)
)
```

👉 仍然在原树：

```text
App
 └── Page
      └── Button
           └── PortalVNode
```

---

## 2️⃣ Fiber 层（调度层）

👉 Portal Fiber **也在原树**

```text
ButtonFiber
   └── PortalFiber   👈 存在
```

但：

```go
portalFiber.Flags |= FiberPortal
```

---

## 🔥 关键：Fiber 不负责布局位置

👉 所以这里还没问题

---

## 3️⃣ Layout 层（关键分叉）

👉 从这里开始 **分裂**

---

# 二、Portal 在 Layout 中的正确处理

---

## 🔥 核心策略：**“不创建子 LayoutBox”**

---

## ❌ 错误做法

```go
parent.Children = append(parent.Children, portalLayoutBox)
```

👉 这会导致：

* 污染 flow ❌
* 影响尺寸 ❌

---

## ✅ 正确做法：**两步走**

---

## 1️⃣ 主树 Layout 时

```go
func layout(node *Fiber, parentLayout *LayoutBox) {

    if node.IsPortal {
        collectPortal(node) // 🔥 只收集
        return
    }

    layoutBox := createLayoutBox(node)

    parentLayout.Children = append(parentLayout.Children, layoutBox)

    for _, child := range node.Children {
        layout(child, layoutBox)
    }
}
```

---

👉 结论：

> ❗ Portal **不会生成主树 LayoutBox**

---

# 三、Portal LayoutBox 从哪里来？

---

## ✅ 答案：**在 Overlay 阶段“重新创建一棵 Layout 子树”**

---

## Portal Entry

```go
type OverlayEntry struct {
    FiberRoot *Fiber
    LayoutRoot *LayoutBox
}
```

---

## 构建 Overlay Layout

```go
func buildOverlayLayout(entry *OverlayEntry, rootW, rootH int) {

    entry.LayoutRoot = buildLayoutTree(entry.FiberRoot, nil)

    layoutOverlay(entry.LayoutRoot, rootW, rootH)
}
```

---

👉 注意：

> 🔥 **这是“第二棵 Layout Tree”**

---

# 四、Portal 的布局信息从哪里来？

---

## 来源：VNode / Props

```go
Box(
    Width(40),
    Height(10),
    AnchorCenter(),
)
```

---

## 转换：

```go
type LayoutBox struct {
    W, H int

    Anchor AnchorType
    Position PositionType
}
```

---

## 在 Overlay Layout 使用

```go
func layoutOverlay(node *LayoutBox, rootW, rootH int) {

    switch node.Anchor {

    case AnchorCenter:
        node.AbsX = (rootW - node.W)/2
        node.AbsY = (rootH - node.H)/2
    }
}
```

---

👉 **完全不依赖 parent**

---

# 五、关键问题：Portal 是“引用”还是“复制”？

---

## 🔥 正确答案：**Fiber 共享，Layout 独立**

---

| 层级           | 是否共享 |
| ------------ | ---- |
| VNode        | ✅    |
| Fiber        | ✅    |
| LayoutBox    | ❌    |
| PaintableBox | ❌    |

---

👉 解释：

* Fiber → 状态 / diff
* Layout → 位置（不同世界）

---

# 六、完整生命周期（非常关键）

---

## 1️⃣ Render 阶段

```text
VNode → Fiber
```

---

## 2️⃣ Layout 阶段

```text
Main Tree:
    Portal → collect（跳过）

Overlay:
    Portal → buildLayoutTree（新建）
```

---

## 3️⃣ Paint 阶段

```text
Main Paint
Overlay Paint
```

---

## 4️⃣ 合成

```text
Layer:
    base
    popup
    modal
```

---

# 七、Portal 的“真实存在形式”

---

你可以这样理解：

---

## 🔥 Portal 是一个“跨树引用”

```text
Main Fiber Tree
     │
     ├── PortalFiber
            │
            └──────► Overlay Fiber Subtree
                          │
                          └── Layout Tree（独立）
```

---

👉 注意：

> ❗ Layout 是“断开的”
> ❗ Fiber 是“连续的”

---

# 八、一个常见坑（你很可能会踩）

---

## ❌ 用同一个 LayoutBox

```go
portalNode.Layout = existingLayout
```

结果：

* 坐标污染
* 重复绘制
* diff 错乱

---

## ✅ 正确

```go
new LayoutBox for overlay
```

---

# 九、简化实现版本（如果你想先跑起来）

---

## 可以先这样做：

👉 不建第二棵 Fiber Tree，只建 Layout Tree

```go
portalQueue []Fiber
```

```go
for _, f := range portalQueue {
    layoutOverlay(buildLayoutTree(f))
}
```

---

👉 这是 MVP，可行

---

# 十、最终总结（非常关键）

---

## Portal 在你的架构中：

---

### ✅ 在 Fiber 中：

* 存在（逻辑归属）
* 不移动

---

### ❌ 在主 Layout 中：

* 不存在
* 不参与计算

---

### ✅ 在 Overlay Layout 中：

* 重新生成 LayoutBox
* 使用 Root 坐标系

---

### ✅ 在 Layer 中：

* 作为高 Z 层绘制

---

# 一句话本质

> 🔥 Portal = “Fiber 不动，Layout 重建”

---

# 如果你要继续进阶（下一步建议）

你现在已经到了一个很关键的阶段，可以继续升级：

---

## 👉 1️⃣ Anchor + Portal（难点）

* tooltip 跟随元素
* resize / scroll 同步

---

## 👉 2️⃣ Portal Diff（性能）

* overlay 局部刷新
* 不影响主树

---

## 👉 3️⃣ Focus 系统（必须补）

* modal trap
* tab navigation

---

如果你愿意，我可以下一步帮你：

👉 画出 **Portal + Anchor + Scroll 三者联动的完整模型（这是最难的一层）**

这个问题你已经问到**实现层的分水岭**了：
👉 Portal 到底是“容器节点”还是“特殊标记”？
👉 在 flex / stack 里到底算不算子元素？

我直接给你**工程级结论**👇

---

# ✅ 一句话结论（非常关键）

> 🔥 **Portal 在结构上是“容器”，在布局上是“透明的（不参与）”**

---

# 一、Portal 的“双重身份”

你必须接受这个“看起来矛盾但非常关键”的设计：

---

## 1️⃣ 在声明 / Fiber 中

👉 Portal 是一个**容器节点**

```go
Portal("modal",
    Box(...),
)
```

```text
Flex
 ├── Item1
 ├── Portal   👈 是子节点（逻辑上）
 └── Item2
```

---

## 2️⃣ 在 Layout 中

👉 Portal 是**不存在的（透明）**

```text
Flex Layout:
 ├── Item1
 └── Item2
```

👉 Portal **被提取走了**

---

# 二、为什么必须这样设计？

如果你不这么做：

---

## ❌ Portal 参与 flex

```go
flexWidth = sum(children.width)
```

👉 结果：

* modal 占空间 ❌
* 布局错位 ❌

---

## ❌ Portal 不在树里

👉 结果：

* state 断裂 ❌
* diff 不稳定 ❌
* 生命周期丢失 ❌

---

## ✅ 正确答案

> 🔥 **逻辑归属在原树，布局归属在 Overlay**

---

# 三、Portal 在 flex / stack 中的真实行为

---

## 示例

```go
Flex(
    Box(W:10),
    Portal("modal",
        Box(W:40, H:10),
    ),
    Box(W:20),
)
```

---

## Layout 实际计算

```text
Flex children（参与布局）:
 ├── Box(10)
 └── Box(20)
```

👉 Portal 完全不参与：

```go
flexWidth = 10 + 20
```

---

## Portal 去哪里了？

👉 被“收集”：

```go
portalQueue = [
    Portal(Box(40x10))
]
```

---

# 四、Portal 是否需要“容器 LayoutBox”？

---

## ❌ 不需要（在主树）

```go
// 不要这样
parent.Children = append(parent.Children, portalLayoutBox)
```

---

## ✅ 需要（在 Overlay）

```go
overlayLayoutRoot = LayoutBox{
    Children: buildFromPortalChildren(...)
}
```

---

👉 也就是说：

| 位置      | 是否有 LayoutBox |
| ------- | ------------- |
| 主树      | ❌ 没有          |
| Overlay | ✅ 有           |

---

# 五、Portal 的正确处理流程（你必须这样实现）

---

## 🔥 Step 1：Layout 主树

```go
for _, child := range node.Children {

    if child.IsPortal {
        portalQueue = append(portalQueue, child)
        continue // ❗跳过
    }

    layout(child)
}
```

---

## 🔥 Step 2：Overlay Layout

```go
for _, portal := range portalQueue {

    layoutTree := buildLayoutTree(portal.Children)

    layoutOverlay(layoutTree)
}
```

---

👉 注意：

> ❗ 是 portal.Children，不是 portal 本身

---

# 六、Portal 本身要不要生成 Box？

---

## 🔥 关键点

> ❗ **Portal 自己通常“不渲染”，只是一个逻辑容器**

---

## 类似于：

```text
React.Fragment / Portal
```

---

## 所以：

```go
Portal(
    Box(...)
)
```

👉 实际 Layout：

```text
Overlay:
 └── Box(...)
```

---

# 七、一个非常容易踩的坑

---

## ❌ 把 Portal 当成普通容器

```go
layout(portal)
layout(portal.Children)
```

👉 结果：

* parent 参与 ❌
* 坐标错 ❌
* clip 错 ❌

---

## ✅ 正确

```go
// 主树
skip portal

// overlay
layout(portal.Children)
```

---

# 八、Portal + Flex 的“正确心智模型”

---

## 🔥 Portal = “布局空洞（hole）”

---

```text
Flex:
 ├── Item1
 ├── [Portal]   👈 空洞
 └── Item2
```

---

👉 Layout 视角：

```text
Item1 + Item2
```

---

👉 渲染视角：

```text
Main Layer: Item1, Item2
Overlay:    Modal
```

---

# 九、进阶一点（你现在这个架构必须考虑）

---

## Portal 可能需要“占位”吗？

👉 有两种模式：

---

### ✅ 模式1：完全不占位（推荐）

```go
Portal(...) // 不影响 layout
```

---

### ⚠️ 模式2：占位 Portal（极少数）

比如：

* dropdown anchor
* inline popup

👉 这时候：

```go
Portal + Placeholder
```

---

# 十、总结（非常关键）

---

## 🔥 Portal 在你的系统中应该是：

---

### 1️⃣ 声明层

* 是容器（有 children）

---

### 2️⃣ Fiber 层

* 是节点（参与 diff）

---

### 3️⃣ Layout 层（主树）

* ❌ 不存在
* ❌ 不参与 flex / stack

---

### 4️⃣ Layout 层（Overlay）

* ✅ 重新生成子树
* ✅ 独立布局

---

### 5️⃣ 渲染层

* 在高 Z Layer

---

# 一句话本质

> 🔥 Portal = “结构存在 + 布局消失 + Overlay 重生”

---

# 如果你下一步要继续（非常建议）

你现在可以进入真正难点了：

---

## 👉 🔥 Portal + Anchor（核心难题）

* dropdown 跟随按钮
* tooltip 自动避让边界

---

## 👉 🔥 Portal + Scroll

* 滚动容器内定位
* viewport clipping

---

## 👉 🔥 Portal + Diff

* overlay 局部刷新
* 不影响主树

---

如果你愿意，我可以下一步帮你直接设计：

👉 **Anchor + Scroll + Portal 三者统一坐标系统（这是整个 TUI 最难的一层）**


很好，这一步你已经进入 **Overlay 系统的“多实例调度”问题**了。
如果这里设计不好，会直接导致：

* modal / tooltip 互相覆盖错乱
* z-index 失控
* 重绘顺序不稳定
* 性能爆炸（每个 portal 都全屏 repaint）

我给你一个**可以直接落地的完整模型**👇

---

# ✅ 一句话结论

> 🔥 **多个 Portal = 多个 Overlay Entry（独立布局） + 统一 Layer 合成（排序绘制）**

---

# 一、核心结构（必须引入）

---

## 🔥 OverlayManager（核心调度中心）

```go
type OverlayManager struct {
    entries []*OverlayEntry // 🔥 有序列表（很关键）
}
```

---

## 🔥 OverlayEntry（每个 Portal 一个）

```go
type OverlayEntry struct {
    ID string

    FiberRoot *Fiber

    LayoutRoot *LayoutBox
    PaintBoxes []*PaintableBox

    Z int
    Visible bool

    Anchor *LayoutBox // 可选（tooltip）
}
```

---

# 二、多个 Portal 的来源

---

## 1️⃣ 同一棵树中

```go
Flex(
    Portal("tooltip", ...),
    Portal("modal", ...),
)
```

---

## 2️⃣ 不同组件触发

```go
Button → Portal(modal)
Input  → Portal(tooltip)
Toast  → Portal(notification)
```

---

👉 最终统一变成：

```go
overlayManager.entries = [
    tooltip,
    modal,
    toast,
]
```

---

# 三、布局策略（关键）

---

# 🔥 核心原则

> ❗ 每个 Portal **独立 layout**
> ❗ 但 **共享同一个 Root 坐标系**

---

## 布局流程

```go
func layoutOverlays(rootW, rootH int) {

    for _, entry := range overlayManager.entries {

        if !entry.Visible {
            continue
        }

        // 1️⃣ 构建独立 Layout Tree
        entry.LayoutRoot = buildLayoutTree(entry.FiberRoot)

        // 2️⃣ 计算布局（🔥 root 坐标系）
        layoutOverlay(entry.LayoutRoot, rootW, rootH)
    }
}
```

---

# 四、多个 Portal 如何定位？

---

## 🔥 三种典型模式

---

## 1️⃣ Modal（全局居中）

```go
node.AbsX = (rootW - node.W)/2
node.AbsY = (rootH - node.H)/2
```

👉 完全独立，不关心其他 portal

---

## 2️⃣ Tooltip（锚点）

```go
node.AbsX = anchor.AbsX
node.AbsY = anchor.AbsY + anchor.H
```

👉 依赖主树节点

---

## 3️⃣ Toast（堆叠布局）

👉 这里是重点👇

---

### 🔥 多 Portal 之间“相互影响”的唯一场景

```go
Toast1 (top=0)
Toast2 (top=+1)
Toast3 (top=+2)
```

---

## 实现

```go
func layoutToastStack(entries []*OverlayEntry) {

    y := 0

    for _, e := range entries {

        e.LayoutRoot.AbsX = rootW - e.LayoutRoot.W
        e.LayoutRoot.AbsY = y

        y += e.LayoutRoot.H + 1
    }
}
```

---

👉 注意：

> 🔥 这是“Overlay 内部布局”，不是主 Layout

---

# 五、Z-index 排序（决定谁盖谁）

---

## 🔥 排序规则

```go
sort.Slice(entries, func(i, j int) bool {
    return entries[i].Z < entries[j].Z
})
```

---

## 常见设计

| 类型       | Z   |
| -------- | --- |
| Base UI  | 0   |
| Dropdown | 10  |
| Tooltip  | 20  |
| Modal    | 100 |
| Toast    | 200 |

---

---

# 六、Layer 合成（最终绘制）

---

```go
func buildLayers() {

    // 1️⃣ 主树
    addToLayer(0, mainBoxes)

    // 2️⃣ Overlay（按 Z）
    for _, entry := range sortedEntries {

        layer := layerManager.Get(entry.Z)

        layer.Boxes = append(layer.Boxes, entry.PaintBoxes...)
    }
}
```

---

👉 关键点：

> 🔥 所有 Portal 最终“合流”到 Layer

---

# 七、事件系统（多 Portal 冲突解决）

---

## 🔥 事件优先级

```text
Top Overlay
   ↓
...
   ↓
Bottom Overlay
   ↓
Main Tree
```

---

## 实现

```go
for i := len(entries)-1; i >= 0; i-- {

    e := entries[i]

    if hit(e, event) {
        handle(e)
        return
    }
}
```

---

👉 从上往下（Z 最大优先）

---

# 八、性能关键点（非常重要）

---

## ❌ 错误做法

```go
// 每个 portal 都全屏 redraw
redraw(screen)
```

---

## ✅ 正确

```go
dirtyRects = append(dirtyRects, entryDirty...)
```

---

👉 每个 Portal：

* 独立 dirty region
* 合并渲染

---

# 九、你这个架构的“正确抽象”

---

## 🔥 统一模型

```text
Portal
   ↓
OverlayEntry（1:1）
   ↓
独立 Layout
   ↓
独立 Paint
   ↓
Layer 合成
```

---

# 十、最关键的设计边界

---

## ❗ Portal 之间默认“互不影响”

---

只有三种情况例外：

---

### 1️⃣ 同组 Overlay（如 Toast）

👉 需要 stack layout

---

### 2️⃣ Anchor 冲突

👉 tooltip 避让

---

### 3️⃣ Focus 管理

👉 modal 阻断下层

---

---

# 十一、推荐你实现的最小版本（MVP）

---

## 🔥 必须实现

```go
OverlayManager {
    entries []
}
```

```go
for each portal:
    collect
    build layout
    paint
    merge layer
```

---

## ❗ 暂时不用做

* overlay 之间碰撞检测
* 智能避让
* 动态 reflow

---

# 十二、一句话总结

> 🔥 多 Portal = 多棵独立 Layout Tree + 单一 Layer 合成 + Z 排序 + 事件反向分发

---

# 如果你下一步继续（我强烈建议）

你现在可以进入真正“高级 TUI”的阶段了：

---

## 👉 🔥 Overlay Group System

* toast stack
* dropdown group
* modal stack

---

## 👉 🔥 Anchor + Scroll 联动（最难）

* tooltip 跟随滚动
* clipping 修正

---

## 👉 🔥 Overlay Diff（性能天花板）

* 局部 repaint
* 避免全屏刷新

---

如果你愿意，我可以下一步直接帮你设计：

👉 **Overlay + Anchor + Scroll 的统一坐标系统（工业级实现）**
