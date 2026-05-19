# Portal Design

## 一、整体架构

```
                User Input
                    │
                    ▼
            Event Dispatcher
                    │
          ┌────────┴────────┐
          ▼                 ▼
    Overlay System      Main Tree
         │                  │
         └────────┬─────────┘
                  │
      Fiber Scheduler (Lane Priority)
                  │
                  ▼
          Reconcile (VNode → Fiber)
                  │
                  ▼
             Commit Phase
                  │
    ┌─────────────┼──────────────┐
    ▼             ▼              ▼
 Layout Tree  Overlay Layout  Focus Manager
 (LayoutBox)   (Root-based)   (Active Node)
    │             │
    └──────┬──────┘
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

## 二、核心分层

### 1. 声明层（VNode）

```go
type VNode struct {
    Type string
    Props map[string]interface{}
    Children []*VNode
    PortalTarget string // 分流点
}
```

作用：描述 UI，不参与计算，不保存状态。

### 2. 调度层（Fiber）

```go
type Fiber struct {
    VNode *VNode
    Parent, Child, Sibling *Fiber
    Lane int
    Dirty bool
}
```

作用：增量更新，优先级调度，中断/恢复。

### 3. 布局层（LayoutBox）

```go
type LayoutBox struct {
    X, Y int
    W, H int
    AbsX, AbsY int
}
```

作用：计算位置，解决嵌套坐标问题。

### 4. 渲染层（PaintableBox）

```go
type PaintableBox struct {
    X, Y int
    W, H int
    Style Style
    Content string
}
```

作用：真正绘制的数据。

### 5. 合成层（Layer）

```go
type Layer struct {
    Z int
    Boxes []*PaintableBox
}
```

作用：控制覆盖关系。

### 6. Overlay 子系统

本质：第二棵 Fiber 世界。

## 三、数据流

### 主数据流

```
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

### Portal 分流

```
VNode
 ├── Normal Node → Main Fiber Tree
 └── Portal Node → Overlay Fiber Tree
```

两棵 Fiber 树，一个 Renderer。

## 四、事件流

### 正确顺序

```
Input
  ↓
Overlay（从上到下）
  ↓（如果没消费）
Main Tree
```

### 实现模型

```go
func DispatchEvent(e Event) {
    // 1. Overlay 优先
    for _, overlay := range overlayManager.TopDown() {
        if hit(overlay, e) {
            handle(overlay, e)
            return
        }
    }

    // 2. 主树
    dispatchMainTree(e)
}
```

### Modal 阻断

```go
if overlayActive {
    ignoreBelowOverlay()
}
```

## 五、Overlay 数据流

```
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

## 六、渲染合成

### 最终输出顺序

```
Layer 0   → Base UI
Layer 10  → Popup
Layer 100 → Modal
```

### Renderer

```go
for _, layer := range layersSortedByZ {
    draw(layer.Boxes)
}
```

## 七、Diff + Partial Render

### 脏区来源

```
Main Tree Dirty
Overlay Dirty
Scroll Dirty
Animation Dirty
```

### 合并

```go
dirtyRects = merge(
    mainDirty,
    overlayDirty,
)
```

### 渲染

```go
for _, rect := range dirtyRects {
    redraw(rect)
}
```

## 八、Fiber + Lane（调度优先级）

### 优先级模型

```go
const (
    LaneOverlay = 0 // 最高
    LaneInput   = 1
    LaneNormal  = 2
    LaneIdle    = 3
)
```

### 调度策略

```
Overlay > Input > Normal > Idle
```

## 九、三个关键解耦

### 1. Layout ≠ Render

Layout 只算位置，Render 才画。

### 2. Tree ≠ Layer

Tree 是逻辑结构，Layer 是视觉结构。

### 3. Portal ≠ 子节点

Portal 是跨树。

## 十、Portal 的本质

Portal = Fiber 归属不变 + Layout/Render 重定向

### 正确实现必须满足 3 个不变量

#### 1. Fiber 结构不能断

```
Button
 └── Modal   // 逻辑关系必须存在
```

用于：state 传递、context、生命周期。

#### 2. Layout 父必须可重定向

```go
modal.PortalRoot = root.OverlayLayer
```

Layout 时：

```go
layoutParent := fiber.Parent
if fiber.PortalRoot != nil {
    layoutParent = fiber.PortalRoot
}
```

#### 3. Render 顺序必须基于最终挂载点

不是 Fiber 顺序。

## 十一、Portal 实现

### Fiber 阶段（保持逻辑树）

```go
type Fiber struct {
    Parent   *Fiber
    Child    *Fiber
    Sibling  *Fiber
    StateNode *LayoutBox
    PortalRoot *Fiber // 核心
}
```

### Layout 阶段（重定向父节点）

```go
func layout(f *Fiber, parent *LayoutBox, root *LayoutBox) {
    node := f.StateNode
    var layoutParent *LayoutBox

    if f.PortalRoot != nil {
        layoutParent = f.PortalRoot.StateNode   // 跳到 Overlay
    } else {
        layoutParent = parent
    }

    computeLayout(node, layoutParent, root)

    for c := f.Child; c != nil; c = c.Sibling {
        layout(c, node, root)
    }
}
```

### Flatten 阶段（分层收集）

#### 定义 Layer Bucket

```go
type Layer int

const (
    LayerNormal Layer = iota
    LayerOverlay
)
```

#### Flatten

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

#### 最终顺序

```go
final := append(layers[LayerNormal], layers[LayerOverlay]...)
```

## 十二、Portal + PositionFixed

### Layout 规则

```go
if node.Position == PositionFixed {
    node.AbsX = (root.W - node.W)/2
    node.AbsY = (root.H - node.H)/2
}
```

Portal 解决"在哪一层"，Fixed 解决"相对于谁定位"。

## 十三、事件系统适配 Portal

### 命中检测顺序

```go
for i := len(nodes)-1; i >= 0; i-- {
    if hit(nodes[i]) {
        dispatch(nodes[i])
        break
    }
}
```

overlay 在最后，优先命中。

### Focus 系统

基于 Fiber Tree，而不是 flattenNodes。

```go
focusNext(fiberTree)
```

Portal 不打断逻辑焦点流。

## 十四、常见错误

### 1. 只在 Render 做 Portal

导致：layout 错、clip 错。

### 2. 直接修改 Fiber Parent

```go
modal.Parent = overlay
```

导致：state 丢失、diff 崩。

### 3. Portal 但不分 Layer

导致：z-index 混乱。

### 4. Fixed 但没 Portal

导致：被 clip、被 scroll。

## 十五、整体架构总结

四个角色职责：

- Fiber = 结构（WHAT）
- LayoutBox = 几何（WHERE）
- PaintableBox = 渲染（HOW）
- Layer = 调度/优先级（WHEN）

### 三条单向数据流

#### 1. Fiber → LayoutBox

```
fiber → 生成 → layout tree
```

#### 2. LayoutBox → PaintableBox

```
layout → 生成 → paint list
```

#### 3. PaintableBox → Terminal

```
paint → 输出 → buffer
```

### 禁止反向依赖

```
Paintable → Layout
Layout → Fiber（除只读）
```

## 十六、Portal 在架构中的位置

Portal 是 Fiber 层的语义 + Layout 层的执行策略。

| 层      | Portal 是否存在 | 作用             |
| ------ | -------------- | -------------- |
| Fiber  | 有             | 声明跳到别处渲染    |
| Layout | 有             | 改变 parent（坐标系） |
| Paint  | 没有            | 完全透明           |
| Layer  | 不直接参与         | 只负责排序          |

### Portal 最小语义定义

```go
type Fiber struct {
    Parent, Child, Sibling *Fiber
    StateNode *LayoutBox
    PortalRoot *Fiber
}
```

如果 PortalRoot != nil：
- 逻辑属于 Parent
- 但布局属于 PortalRoot

## 十七、Portal 与 Layer 的关系

Portal ≠ Layer

错误理解：
```
Portal = 高 ZIndex
```

正确关系：
```
Portal → 决定"挂在哪棵树"
Layer  → 决定"谁先画"
```

### 正确组合

```go
if fiber.PortalRoot != nil {
    fiber.Lane = LaneOverlay
}
```

排序：
```go
sort by (Lane, order)
```

| 能力        | Portal | Layer |
| --------- | ------ | ----- |
| 脱离父布局     | ✅      | ❌     |
| 顶层显示      | ❌      | ✅     |
| 事件优先      | ❌      | ✅     |
| 渲染顺序      | ❌      | ✅     |

## 十八、Portal 与 LayoutBox

核心原则：
```
LayoutBox 的 Parent ≠ Fiber 的 Parent
```

```go
if fiber.PortalRoot != nil {
    layoutParent = fiber.PortalRoot.StateNode
} else {
    layoutParent = parent
}
node.Parent = layoutParent
```

结果：
```
Fiber Tree:
Button → Modal

Layout Tree:
Root → Modal
```

## 十九、Portal 与 Clip / Scroll

处理方式：

### Clip 继承

```go
if portal {
    node.Clip = root.Clip
}
```

### Scroll 断开

```go
if portal {
    ignore parent scroll
}
```

Portal = 切断所有父几何影响。

## 二十、Portal 与 Diff

不能这样：
```go
modal.Parent = overlay
```

否则：diff 错误、state 丢失

正确：
```
Diff 只基于 Fiber Tree
```

Portal 只是渲染策略。

## 二十一、完整集成图

```
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

## 二十二、实现检查清单

必须满足：

- [ ] Fiber Parent 不变
- [ ] Layout Parent 可变
- [ ] Flatten 分层
- [ ] Layer 控制顺序
- [ ] Clip 不继承父
- [ ] Scroll 不影响 Portal
- [ ] Event 逆序命中
- [ ] Diff 不感知 Portal

## 二十三、VNode → Portal 信息传递

### VNode 声明 Portal

```go
VNode{
    Type: "Modal",
    Props: {
        "portal": "overlay",
    },
}
```

### Reconcile 时解析

```go
func createFiber(v VNode, parent *Fiber) *Fiber {
    f := &Fiber{
        Type: v.Type,
        Parent: parent,
    }

    if target, ok := v.Props["portal"]; ok {
        f.PortalRoot = resolvePortalTarget(target)
    }

    return f
}
```

### resolvePortalTarget

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

Portal 信息已经从 VNode 注入到 Fiber。

## 二十四、Portal → LayerManager

### Layer 来源两种方式

方式1：VNode 指定
```go
Props: {
    portal: "overlay",
    layer: "modal",
}
```

方式2：Portal 自动提升（推荐）
```go
if fiber.PortalRoot != nil {
    fiber.Lane = LaneOverlay
}
```

推荐策略：Portal 自动映射到 Layer。

### 注册到 LayerManager

```go
func collect(f *Fiber, lm *LayerManager) {
    layer := f.Lane
    lm.layers[layer] = append(lm.layers[layer], f)

    for c := f.Child; c != nil; c = c.Sibling {
        collect(c, lm)
    }
}
```

只关心 Lane，不关心 PortalRoot。

## 二十五、完整示例

### VNode
```
App
 └── Button
      └── Modal (portal=overlay)
```

### Fiber（逻辑结构）
```
App
 └── Button
      └── Modal
           PortalRoot → OverlayLayer
           Lane → Overlay
```

### LayerManager
```
LayerNormal: [App, Button]
LayerOverlay: [Modal]
```

### Render 顺序
```
Normal → Overlay
```

结果：Modal 不受 Button 布局影响，Modal 在最上层。

## 二十六、Portal 子系统设计

### OverlayManager 结构

```go
type OverlayManager struct {
    entries map[string]*OverlayEntry
}
```

### OverlayEntry

```go
type OverlayEntry struct {
    Key string // modal / tooltip
    RootFiber *Fiber
    Z int
    Visible bool
    TrapFocus bool
}
```

### 注册 Portal

```go
func mountPortal(vnode *VNode) {
    entry := overlayManager.GetOrCreate(vnode.PortalTarget)
    entry.RootFiber = buildFiberTree(vnode.Children)
    entry.Visible = true
}
```

## 二十七、Overlay Layout 规则

### 强制 Root 坐标系

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

忽略 Parent、Scroll、Clip，完全独立坐标系。

## 二十八、Portal 不参与原 Layout

### 两阶段 Layout

阶段一：主树 Layout（忽略 Portal）

```go
func layout(node *LayoutBox) {
    for _, child := range node.Children {
        if child.IsPortal {
            collectPortal(child) // 只收集，不布局
            continue
        }
        layout(child)
    }
    computeSize(node)
}
```

阶段二：Overlay Layout

```go
func layoutOverlay(rootW, rootH int) {
    for _, item := range portalQueue {
        layoutPortalNode(item.Node, rootW, rootH)
    }
}
```

### Portal 收集机制

```go
type PortalItem struct {
    Node *LayoutBox
    Target string // modal / tooltip
}

var portalQueue []PortalItem

func collectPortal(node *LayoutBox) {
    portalQueue = append(portalQueue, PortalItem{
        Node: node,
        Target: node.PortalTarget,
    })
}
```

### Portal 坐标计算

普通节点：
```go
node.AbsX = parent.AbsX + node.X
```

Portal 节点：
```go
node.AbsX = rootCoordX(...)
node.AbsY = rootCoordY(...)
```

常见定位策略：

居中 Modal：
```go
node.AbsX = (rootW - node.W) / 2
node.AbsY = (rootH - node.H) / 2
```

全屏遮罩：
```go
node.AbsX = 0
node.AbsY = 0
node.W = rootW
node.H = rootH
```

Tooltip（相对锚点）：
```go
anchor := findAnchor(node.AnchorID)
node.AbsX = anchor.AbsX
node.AbsY = anchor.AbsY + anchor.H
```

## 二十九、Portal 在结构中的行为

### 双重身份

在声明/Fiber中：是容器节点
```
Flex
 ├── Item1
 ├── Portal   👈 是子节点（逻辑上）
 └── Item2
```

在 Layout 中：是不存在的（透明）
```
Flex Layout:
 ├── Item1
 └── Item2
```

### Portal 在 flex / stack 中

示例：
```go
Flex(
    Box(W:10),
    Portal("modal",
        Box(W:40, H:10),
    ),
    Box(W:20),
)
```

Layout 实际计算：
```
Flex children（参与布局）:
 ├── Box(10)
 └── Box(20)
```

flexWidth = 10 + 20

Portal 被收集到 portalQueue。

### Portal 是否需要容器 LayoutBox

在主树：不需要
```go
// 不要这样
parent.Children = append(parent.Children, portalLayoutBox)
```

在 Overlay：需要
```go
overlayLayoutRoot = LayoutBox{
    Children: buildFromPortalChildren(...)
}
```

### Portal 本身是否渲染

Portal 自己通常不渲染，只是一个逻辑容器。

类似于：
```
React.Fragment / Portal
```

实际 Layout：
```
Overlay:
 └── Box(...)
```

### Portal 的正确处理流程

Step 1：Layout 主树
```go
for _, child := range node.Children {
    if child.IsPortal {
        portalQueue = append(portalQueue, child)
        continue // 跳过
    }
    layout(child)
}
```

Step 2：Overlay Layout
```go
for _, portal := range portalQueue {
    layoutTree := buildLayoutTree(portal.Children)
    layoutOverlay(layoutTree)
}
```

是 portal.Children，不是 portal 本身。

## 三十、多 Portal 管理

### OverlayManager

```go
type OverlayManager struct {
    entries []*OverlayEntry // 有序列表
}
```

### OverlayEntry

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

### 多个 Portal 的来源

同一棵树中：
```go
Flex(
    Portal("tooltip", ...),
    Portal("modal", ...),
)
```

不同组件触发：
```go
Button → Portal(modal)
Input  → Portal(tooltip)
Toast  → Portal(notification)
```

### 布局策略

每个 Portal 独立 layout，共享同一个 Root 坐标系。

```go
func layoutOverlays(rootW, rootH int) {
    for _, entry := range overlayManager.entries {
        if !entry.Visible {
            continue
        }

        // 构建独立 Layout Tree
        entry.LayoutRoot = buildLayoutTree(entry.FiberRoot)

        // 计算布局（root 坐标系）
        layoutOverlay(entry.LayoutRoot, rootW, rootH)
    }
}
```

### 多种定位模式

Modal（全局居中）：
```go
node.AbsX = (rootW - node.W)/2
node.AbsY = (rootH - node.H)/2
```

Tooltip（锚点）：
```go
node.AbsX = anchor.AbsX
node.AbsY = anchor.AbsY + anchor.H
```

Toast（堆叠布局）：
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

### Z-index 排序

```go
sort.Slice(entries, func(i, j int) bool {
    return entries[i].Z < entries[j].Z
})
```

常见设计：
```
Base UI   → 0
Dropdown  → 10
Tooltip   → 20
Modal     → 100
Toast     → 200
```

### Layer 合成

```go
func buildLayers() {
    // 主树
    addToLayer(0, mainBoxes)

    // Overlay（按 Z）
    for _, entry := range sortedEntries {
        layer := layerManager.Get(entry.Z)
        layer.Boxes = append(layer.Boxes, entry.PaintBoxes...)
    }
}
```

### 事件系统（多 Portal 冲突解决）

事件优先级：
```
Top Overlay
   ↓
...
   ↓
Bottom Overlay
   ↓
Main Tree
```

```go
for i := len(entries)-1; i >= 0; i-- {
    e := entries[i]
    if hit(e, event) {
        handle(e)
        return
    }
}
```

从上往下（Z 最大优先）。

### 性能关键

每个 Portal：
- 独立 dirty region
- 合并渲染

```go
dirtyRects = append(dirtyRects, entryDirty...)
```

---

## 总结

### Portal 的本质

Portal = 改变坐标系 + 改变挂载层，但不改变结构

Portal = Fiber 不动，Layout 重建

Portal = "布局跳过 + 渲染重接"

Portal = "结构存在 + 布局消失 + Overlay 重生"

### 关键原则

1. Portal 在 Fiber 中：存在（逻辑归属），不移动
2. Portal 在主 Layout 中：不存在，不参与计算
3. Portal 在 Overlay Layout 中：重新生成 LayoutBox，使用 Root 坐标系
4. Portal 在 Layer 中：作为高 Z 层绘制

### 多 Portal 模型

多 Portal = 多棵独立 Layout Tree + 单一 Layer 合成 + Z 排序 + 事件反向分发
