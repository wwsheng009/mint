# Fiber 统一架构重构方案

**版本**: 1.0
**日期**: 2026-02-14
**作者**: Crush AI
**状态**: 设计阶段

---

## 📋 目录

1. [引言](#引言)
2. [当前架构问题分析](#当前架构问题分析)
3. [新架构设计原则](#新架构设计原则)
4. [核心数据结构重构](#核心数据结构重构)
5. [完整数据流设计](#完整数据流设计)
6. [详细重构方案](#详细重构方案)
7. [破坏性变更清单](#破坏性变更清单)
8. [实施步骤](#实施步骤)
9. [测试和验证](#测试和验证)
10. [风险评估和缓解](#风险评估和缓解)
11. [迁移指南](#迁移指南)

---

## 引言

### 1.1 重构目标

本次重构的目标是根据设计文档 `diff_key.md` 和 `diff_layer.md`，对 Mint TUI Framework 进行**重大架构重组**，实现以下核心理念：

1. **Fiber 作为唯一运行时结构**：所有运行时状态、事件处理、布局计算都基于 Fiber，不再依赖 VNode
2. **Layer 作为渲染排序维度**：Layer 不参与结构变换，只控制 Z 轴渲染顺序
3. **废弃 StripLayers**：不再使用 VNode 克隆和树结构分离的方式实现多层渲染
4. **RenderPlane 投影机制**：使用渲染分桶而非树结构分离实现多层渲染

### 1.2 设计文档核心思想回顾

#### `diff_key.md` 核心思想

- **VNode 是声明层**：纯函数产物，描述 UI 结构，参与 sibling diff
- **Fiber 是协调层**：运行时实体，保存 NodeID 和 DiffKey，追踪状态
- **NodeID 不参与 diff**：NodeID 只是 Fiber 的身份标识，diff 基于 DiffKey
- **DiffKey vs NodeID 分离**：DiffKey 用于匹配，NodeID 用于运行时身份

#### `diff_layer.md` 核心思想

- **Layer 是渲染维度不是结构维度**：Layer 只控制绘制顺序，不改变树结构
- **废弃 StripLayers**：不再克隆 VNode 和分离层子树
- **RenderPlane 投影**：在单一 Fiber 树基础上，按 Layer 分桶进行渲染
- **HitTest 支持 Z-order**：从最高层往下检测，支持多层事件优先级

### 1.3 重构性质

**⚠️ 这是一个破坏性重构（Breaking Changes）：**

- 不兼容旧代码
- 需要大量重写现有实现
- 需要全面更新测试用例
- 可能影响所有依赖当前 Layer 系统的组件

---

## 当前架构问题分析

### 2.1 当前架构概览

当前系统采用的是混合架构：

```
VNode Tree (纯声明, 无 NodeID)
   ↓ Reconciler
Fiber Tree (包含 NodeID, DiffKey, 双缓冲)
   ↓ LayerManager
   ├─ StripLayers → BaseVNodeTree (clone VNode)  ❌ 问题1
   ├─ Modal Nodes → ModalVNodeTree                ❌ 问题2
   ├─ Overlay Nodes → OverlayVNodeTree            ❌ 问题3
   └─ Tooltip Nodes → TooltipVNodeTree            ❌ 问题4
   ↓ Layout Engine (多路径布局)
   ├─ Base Layout
   ├─ Modal Layout
   ├─ Overlay Layout
   └─ Tooltip Layout
   ↓ Merge HitMap
   ↓ Event Dispatch
```

### 2.2 核心问题识别

#### 问题 1：StripLayers 违背单一数据源原则

**位置**: `runtime/layer/collector.go:212-228`

```go
// ❌ 当前的实现
func (c *Collector) StripLayers(vnode rtui.VNode) rtui.VNode {
    if vnode.GetLayer() != rtui.LayerBase {
        return nil  // 直接丢弃
    }
    cloned := c.cloneWithoutLayers(vnode)  // ❌ 克隆 VNode
    return cloned
}

func (c *Collector) cloneWithoutLayers(vnode rtui.VNode) rtui.VNode {
    // ❌ 递归 clone 所有 VNode
    switch n := vnode.(type) {
    case *rtui.ElementVNode:
        cloned := rtui.NewElement(n.Tag())
        cloned.SetProps(n.Props().Clone())  // ❌ Clone Props
        cloned.SetStyle(n.Style())
        cloned.SetKey(n.Key())
        cloned.SetChildren(nonLayerChildren)
        return cloned
    }
}
```

**问题分析**：
- ❌ StripLayers 创建了 VNode 的克隆副本
- ❌ 这导致 VNode 树 ≠ Fiber 树
- ❌ identity 来源分裂（VNode 和 Fiber 各有一套）
- ❌ Key 同步困难（克隆的 VNode Key 可能不同）
- ❌ 增加内存开销（每个 render 都要 clone）
- ❌ 难以维持引用一致性

#### 问题 2：Layer 作为结构维度参与 Layout

**位置**: `runtime/layer/manager.go:46-103`

```go
// ❌ 当前的实现
func (m *Manager) CollectAndLayout(
    vnode rtui.VNode,
    fiber *reconciler.Fiber,
    constraints runtime.BoxConstraints,
    engine *compute.Engine,
) error {
    // 1. Collect layer nodes
    m.collector.Collect(vnode)

    // 2. ❌ Strip layer nodes from the main tree
    baseTree := m.collector.StripLayers(vnode)  // ❌ 创建独立树

    // 3. Layout the base layer
    baseLayout, err := engine.Layout(baseTree, fiber, constraints)

    // 4. ❌ Layout each collected layer separately
    for layer, nodes := range m.collector.GetLayers() {
        layerLayout, err := m.layoutLayer(node, layer, constraints, engine, fiber)
        m.layouts[layer] = layerLayout  // ❌ 多个独立 Layout
    }
}
```

**问题分析**：
- ❌ Layer 参与了结构变换（stripping）
- ❌ 每个 Layer 有独立的 Layout 树
- ❌ 布局上下文不一致（base 和 modal 的父子关系丢失）
- ❌ 增加复杂度（需要管理多个 Layout 树）

#### 问题 3：HitMap 基于多源数据构建

**位置**: `runtime/layer/manager.go:362-424`

```go
// ❌ 当前的实现
func (m *Manager) GetMergedHitMap() *event.HitMap {
    var entries []event.HitMapEntryInternal

    // ❌ 从多个独立的 Layout 构建和合并 HitMap
    renderOrder := []rtui.Layer{
        rtui.LayerBase,  // baseLayout
        rtui.LayerOverlay,  // overlayLayout
        rtui.LayerModal,  // modalLayout
        rtui.LayerTooltip,  // tooltipLayout
        rtui.LayerInspector,  // inspectorLayout
    }

    zOrder := 0
    for _, layer := range renderOrder {
        layout, ok := m.layouts[layer]
        if !ok || layout.HitMap == nil {
            continue
        }

        // ❌ 从多个独立的 HitMap 中提取和合并
        for _, entry := range layout.HitMap.AllEntries() {
            newEntry := event.HitMapEntryInternal{
                NodeID:  entry.NodeID,
                Node:    entry.Node,
                Bounds:  entry.Bounds,
                LocalXY: entry.LocalXY,
                ZOrder:  zOrder,  // ❌ 手动管理 Z-order
            }
            entries = append(entries, newEntry)
        }
        zOrder++
    }

    // ❌ sort by Z-order
    sort.Slice(entries, func(i, j int) bool {
        return entries[i].ZOrder < entries[j].ZOrder
    })

    return event.BuildHitMapFromEntries(entries)
}
```

**问题分析**：
- ❌ 需要从多个 Layout 构建多个 HitMap
- ❌ 需要合并多个 HitMap
- ❌ 需要手动管理 Z-order
- ❌ 容易出现同步问题（某个 Layer 的 HitMap 更新不及时）
- ❌ NodeID 可能重复（base 和 modal 的节点可能有相同 ID）

#### 问题 4：Fiber 树与渲染树不一致

**当前状态**：
- Fiber 树：完整的 VNode 树结构
- 渲染树：被 StripLayers 后的多个独立树

**不一致的表现**：
```go
// Fiber Tree (完整结构)
FiberRoot
 └─ App (NodeID: 1)
     ├─ Page (NodeID: 2)
     │   └─ Button (NodeID: 3)
     └─ Modal (NodeID: 4, LayerModal)  // ← Modal 在 Fiber 树中
         └─ ModalButton (NodeID: 5)

// Render Tree (被 Strip 后)
BaseTree
 └─ App (NodeID: ?)  // ← 没有明确的 NodeID 映射
     └─ Page (NodeID: ?)
         └─ Button (NodeID: ?)
// Modal 被移除，独立的 ModalTree
ModalTree
 └─ Modal (NodeID: ?)
     └─ ModalButton (NodeID: ?)
```

**问题分析**：
- ❌ Fiber 树的 NodeID 无法直接映射到渲染树
- ❌ Event Dispatch 需要在不同的树之间查找
- ❌ 难以追踪哪个 VNode 对应哪个 Fiber
- ❌ Debug 困难（Fiber 树结构与渲染树结构不同）

#### 问题 5：VNode 承担了过多职责

**位置**: `runtime/ui/vnode.go` 作为接口

```go
type VNode interface {
    Type() VNodeType
    Props() Props
    SetProps(p Props)
    Children() []VNode
    SetChildren(children []VNode)
    Key() string
    SetKey(key string)
    Style() style.Style
    SetStyle(s style.Style)
    Tag() string
    GetLayer() Layer          // ← VNode 知道 Layer 逻辑
    SetLayer(l Layer) VNode   // ← VNode 可以设置 Layer
}
```

**问题分析**：
- ❌ VNode 参与了运行时逻辑（Layer 管理）
- ❌ VNode 不应该是纯声明，它影响了运行时行为
- ❌ 违反关注点分离原则

---

## 新架构设计原则

### 3.1 核心原则

#### 原则 1：Fiber 是唯一运行时结构

> "Everything happens in Fiber. VNode is just declarative input."

**含义**：
- 所有运行时状态都保存在 Fiber
- 所有事件处理基于 Fiber.NodeID
- 所有布局计算基于 Fiber 树
- 所有渲染基于 Fiber 树

**禁止**：
- ❌ 直接遍历 VNode 树进行布局或渲染
- ❌ 在 VNode 上保存运行时状态
- ❌ 基于 VNode 构建 HitMap

#### 原则 2：Layer 是渲染排序维度

> "Layer controls drawing order, not tree structure."

**含义**：
- Layer 只是 Fiber 的一个属性
- Layer 不参与 diff 规则
- Layer 不影响 Fiber 树结构
- Layer 只在渲染阶段用于分桶

**禁止**：
- ❌ 根据 Layer 改变树结构
- ❌根据 Layer 修改 NodeID
- ❌ 根据 Layer clone VNode

#### 原则 3：单一 Fiber 树

> "There is only one tree - the Fiber tree."

**含义**：
- 系统中只有一棵完整的 Fiber 树
- 所有节点（包括 Modal、Overlay）都在这棵树中
- Layer 只是标记，不是独立子树

**禁止**：
- ❌ 生成多个独立的 Fiber 树
- ❌ 剥离子树到不同的 FiberRoot
- ❌ 创建临时的 VNode 树

#### 原则 4：VNode 纯声明

> "VNode describes what to render, not how to render."

**含义**：
- VNode 只描述 UI 结构
- VNode 不包含运行时信息
- VNode 每次 render 都是新的快照

**禁止**：
- ❌ 在 VNode 上保存 NodeID
- ❌ 在 VNode 上保存 Instance
- ❌ 在 VNode 上保存 Layout 结果

#### 原则 5：RenderPlane 投影而非 Strip

> "Project, don't strip."

**含义**：
- 使用 RenderPlane 进行渲染分桶
- 不修改树结构，只是创建投影视图
- HitMap 基于单一 Fiber 树构建

**禁止**：
- ❌ StripLayers
- ❌ clone VNode
- ❌ 物理分离子树

### 3.2 架构层次划分

```
┌─────────────────────────────────────────────────┐
│         用户代码 (User Code)                      │
│      组件、Hooks、事件处理                         │
└─────────────────────────────────────────────────┘
                      │
                      ↓ Render()
───────────────────────────────────────────────────
┌─────────────────────────────────────────────────┐
│         声明层 (Declarative Layer)                 │
│         VNode Tree                               │
│  - Type, Key, Props, Children                    │
│  - 纯声明，无运行时信息                           │
│  - 每次 Render 都是新的                          │
└─────────────────────────────────────────────────┘
                      │
                      ↓ Reconcile()
───────────────────────────────────────────────────
┌─────────────────────────────────────────────────┐
│         协调层 (Reconciliation Layer)             │
│         Fiber Tree                               │
│  - NodeID, DiffKey, Type                         │
│  - Parent, Child, Sibling                        │
│  - Props, MemoizedProps, MemoizedState           │
│  - Instance, UpdateQueue                         │
│  - Layer (声明值，不参与 diff)                    │
│  唯一运行时结构                                   │
└─────────────────────────────────────────────────┘
                      │
                      ↓ Layout()
───────────────────────────────────────────────────
┌─────────────────────────────────────────────────┐
│         布局层 (Layout Layer)                     │
│         ComputedBox Tree                          │
│  - Attached to Fiber                             │
│  - Box (X, Y, Width, Height)                     │
│  - Layer (从 Fiber 拷贝)                         │
│  - NodeID (从 Fiber 拷贝)                        │
└─────────────────────────────────────────────────┘
                      │
                      ↓ BuildRenderPlanes()
───────────────────────────────────────────────────
┌─────────────────────────────────────────────────┐
│         渲染投影层 (Render Projection Layer)       │
│         RenderPlanes (按 Layer 分桶)               │
│  - [LayerBase] []*ComputedBox                     │
│  - [LayerModal] []*ComputedBox                    │
│  - [LayerOverlay] []*ComputedBox                  │
│  - [LayerTooltip] []*ComputedBox                  │
│  - [LayerInspector] []*ComputedBox                │
│  只读投影，不修改 Fiber 树                        │
└─────────────────────────────────────────────────┘
                      │
                      ↓ Render()
───────────────────────────────────────────────────
┌─────────────────────────────────────────────────┐
│         渲染层 (Rendering Layer)                  │
│         Paint Buffer                              │
│  按 Layer 顺序绘制                                │
│  Base → Overlay → Modal → Tooltip → Inspector    │
└─────────────────────────────────────────────────┘
                      │
                      ↓ BuildHitMap()
───────────────────────────────────────────────────
┌─────────────────────────────────────────────────┐
│         事件层 (Event Layer)                      │
│         HitMap                                   │
│  - 基于 ComputedBox 构建                        │
│  - 包含 NodeID, Layer, ZOrder                    │
│  - HitTest 从高层往下检测                        │
└─────────────────────────────────────────────────┘
```

### 3.3 数据流向

```
User Code
   ↓
Render() → new VNode tree (每次都是新的)
   ↓
Reconcile(oldFiberTree, newVNodeTree)
   └─ 基于 DiffKey 匹配
   └─ 复用或创建 Fiber
   └─ 保留 NodeID（复用时）或分配新 NodeID（新建时）
   ↓
newFiberTree (单棵完整树)
   ├─ Normal nodes (LayerBase)
   ├─ Modal nodes (LayerModal)
   ├─ Overlay nodes (LayerOverlay)
   ├─ Tooltip nodes (LayerTooltip)
   └─ Inspector nodes (LayerInspector)
   ↓
Layout(newFiberTree)
   └─ 遍历整棵 Fiber 树
   └─ 为每个 Fiber 创建/更新 ComputedBox
   └─ 保存 Layout 结果到 Fiber.ComputedBox
   └─ 拷贝 Fiber.Layer 到 ComputedBox.Layer
   └─ 拷贝 Fiber.NodeID 到 ComputedBox.NodeID
   ↓
RenderPlanes (按 Layer 分桶)
   └─ 遍历所有 ComputedBox
   └─ 按 Layer 分桶
   └─ 不修改 ComputedBox，只创建视图
   ↓
Renderer
   └─ 按 Layer 顺序遍历 RenderPlanes
   └─ 绘制到 Buffer
   ↓
HitMap (基于所有 ComputedBox)
   └─ 包含 NodeID, Layer, ZOrder
   └─ HitTest 从高层往下检测
   ↓
Event Dispatch
   └─ 基于 NodeID 找到 Fiber
   └─ 基于 Fiber 找到 Instance
   └─ 触发事件处理器
```

---

## 核心数据结构重构

### 4.1 VNode（声明层）

**目标**：VNode 保持纯声明属性，移除运行时逻辑。

```go
// VNode 保持不变，但明确职责
type VNode interface {
    // === 基础信息 ===
    Type() VNodeType
    Tag() string

    // === 声明属性 ===
    Props() Props
    SetProps(p Props)
    Children() []VNode
    SetChildren(children []VNode)
    Key() string
    SetKey(key string)

    // === 样式系统 ===
    Style() style.Style
    SetStyle(s style.Style)

    // === ✨ Layer 声明保留 ===
    // VNode 仍然声明在哪个 Layer 渲染
    // 但这只是声明值，不参与 diff
    GetLayer() Layer
    SetLayer(l Layer) VNode

    // === ❌ 废弃的 API ===
    // 不再接受 SetBounds()，那是运行时信息
    // 不再接受 GetBounds()，那是 Layout 结果
}
```

**重要变更**：
- ✅ 保留 `GetLayer()` 和 `SetLayer()`
- ⚠️ 但要注意它们只是声明属性，不参与 diff
- ❌ 废弃任何运行时信息的保存（如 Bounds）

**默认值**：
```go
const DefaultLayer LayerBase

// 如果未设置 Layer，默认为 LayerBase
```

### 4.2 Fiber（协调层）

**目标**：Fiber 成为唯一运行时结构，包含所有运行时信息。

```go
// Fiber 完整定义
type Fiber struct {
    // ========== Identity ==========
    // ✨ NodeID 是 Fiber 的唯一身份标识
    //   - 只在创建新 Fiber 时分配
    //   - 复用 Fiber 时保持不变
    //   - 不参与 diff
    NodeID uint64

    // ✨ DiffKey 用于 sibling diff
    //   - 从 VNode.Key() 拷贝
    //   - 用于在 reconcileChildren 中匹配旧 Fiber
    //   - 如果 Key 为空，使用索引生成
    DiffKey string

    // Key 是 DiffKey 的别名（向后兼容）
    Key string

    // ========== Type ==========
    // VNode 类型（Element, Component, Text, Fragment 等）
    Type VNodeType
    Tag  string

    // ========== Tree Structure ==========
    // Fiber 树的连接关系
    Parent   *Fiber
    Child    *Fiber  // 第一个子节点
    Sibling  *Fiber  // 下一个兄弟节点

    // ========== Double Buffering ==========
    // 上一帧的 Fiber（用于 reconcile 复用判断）
    Alternate *Fiber

    // ========== Props & State ==========
    Props         Props         // 最新的 props（来自当前 VNode）
    MemoizedProps Props         // 上一次的 props（用于 shouldUpdate 判断）
    MemoizedState interface{}   // 组件的 state

    // ========== Update Queue ==========
    UpdateQueue *UpdateQueue

    // ========== Effect Flags ==========
    Flags        EffectFlag
    SubtreeFlags EffectFlag  // 子树 effect 聚合

    // ========== Priority ==========
    Lanes     Lane
    ChildLanes Lane

    // ========== Layer ==========
    // ✨ Layer 声明值（从 VNode 拷贝）
    //   - 默认值：LayerBase
    //   - 不参与 diff
    //   - 只在渲染阶段用于分桶
    Layer Layer

    // ========== Component Instance ==========
    // 组件实例（保存组件的 state 和生命周期）
    ComponentInstance ComponentInstance

    // ========== Computed Box ==========
    // ✨ ComputedBox 保存 Layout 结果
    //   - 在 Layout 阶段计算
    //   - 在 Render 阶段使用
    ComputedBox *compute.ComputedBox
}
```

**关键变更**：
- ✅ 新增 `ComputedBox *compute.ComputedBox` 字段
  - 用于保存 Layout 结果
  - 避免在 LayerManager 中维护独立的 Layout 树
- ✅ 新增 `Layer Layer` 字段
  - 从 VNode 拷贝
  - 默认值为 `LayerBase`
- ✅ 明确 `NodeID` 和 `DiffKey` 的职责分离

### 4.3 ComputedBox（布局层）

**目标**：ComputedBox 保存完整的布局信息和渲染所需的数据。

```go
// ComputedBox 保存单个 Fiber 的布局结果
type ComputedBox struct {
    // === Identity ===
    // ✨ NodeID 从 Fiber 拷贝
    NodeID uint64

    // === Box Geometry ===
    Box Box  // {X, Y, Width, Height}

    // === Layer ===
    // ✨ Layer 从 Fiber 拷贝
    //   - 用于 RenderPlanes 分桶
    //   - 用于 HitTest Z-order
    Layer rtui.Layer

    // === VNode Reference ===
    // ✨ 保存 VNode 引用（用于渲染）
    // 注意：这是引用，不是克隆
    VNode rtui.VNode

    // === Children ===
    // ✨ 子节点的 ComputedBox
    // 构成 ComputedBox 树
    Children []*ComputedBox

    // === Debug Info ===
    DebugPath string  // 用于调试
}
```

**关键变更**：
- ✅ 新增 `NodeID uint64` 字段
- ✅ 新增 `Layer rtui.Layer` 字段
- ✅ 新增 `Children []*ComputedBox` 字段
  - 构成 ComputedBox 树
  - 代替 LayerManager 中的独立 Layout 树

### 4.4 RenderPlane（渲染投影层）

**目标**：RenderPlane 是对 ComputedBox 树的只读分桶视图。

```go
// ✨ RenderPlane 是新的核心结构
// 用于替代 StripLayers，实现多层渲染

// RenderPlanes 按保存 ComputedBox
type RenderPlanes struct {
    // 按 Layer 分桶的 ComputedBox
    Planes map[rtui.Layer][]*compute.ComputedBox

    // 渲染顺序（从低到高）
    RenderOrder []rtui.Layer
}

// NewRenderPlanes 创建新的 RenderPlanes
func NewRenderPlanes() *RenderPlanes {
    return &RenderPlanes{
        Planes: make(map[rtui.Layer][]*compute.ComputedBox),
        RenderOrder: []rtui.Layer{
            rtui.LayerBase,
            rtui.LayerOverlay,
            rtui.LayerModal,
            rtui.LayerTooltip,
            rtui.LayerInspector,
        },
    }
}

// BuildFromFiber 从 Fiber 树构建 RenderPlanes
// 这是一个只读投影，不修改 Fiber 树
func (rp *RenderPlanes) BuildFromFiber(root *Fiber) {
    rp.Plane = make(map[rtui.Layer][]*compute.ComputedBox)

    // 遍历 Fiber 树，按 Layer 分桶
    rp.walkAndCollect(root)

    // 按 RenderOrder 排序每个 Plane 中的 ComputedBox
    rp.sortPlanes()
}

// walkAndCollect 递归遍历 Fiber 树，收集 ComputedBox
func (rp *RenderPlanes) walkAndCollect(fiber *Fiber) {
    if fiber == nil {
        return
    }

    // 如果有 ComputedBox，添加到对应的 Plane
    if fiber.ComputedBox != nil {
        layer := fiber.Layer
        if layer == rtui.LayerBase {
            layer = rtui.LayerBase  // 显式处理默认值
        }
        rp.Planes[layer] = append(rp.Planes[layer], fiber.ComputedBox)
    }

    // 递归处理子节点
    rp.walkAndCollect(fiber.Child)
    rp.walkAndCollect(fiber.Sibling)
}

// sortPlanes 对每个 Plane 中的 ComputedBox 按位置排序
// 确保渲染顺序一致
func (rp *RenderPlanes) sortPlanes() {
    for layer, boxes := range rp.Planes {
        sort.Slice(boxes, func(i, j int) bool {
            // 按 Y 排序，Y 相同按 X 排序
            if boxes[i].Box.Y != boxes[j].Box.Y {
                return boxes[i].Box.Y < boxes[j].Box.Y
            }
            return boxes[i].Box.X < boxes[j].Box.X
        })

        rp.Planes[layer] = boxes
    }
}

// GetPlane 获取指定 Layer 的所有 ComputedBox
func (rp *RenderPlanes) GetPlane(layer rtui.Layer) []*compute.ComputedBox {
    return rp.Planes[layer]
}

// Iterate 遍历所有 Layer 的所有 ComputedBox（按渲染顺序）
func (rp *RenderPlanes) Iterate(fn func(layer rtui.Layer, box *compute.ComputedBox) bool) {
    for _, layer := range rp.RenderOrder {
        for _, box := range rp.Planes[layer] {
            if !fn(layer, box) {
                return
            }
        }
    }
}
```

**关键特性**：
- ✅ 不修改 Fiber 树
- ✅ 不克隆 VNode
- ✅ 一次性分桶，O(n) 时间复杂度
- ✅ 支持按 Layer 遍历

### 4.5 HitMap（事件层）

**目标**：HitMap 基于单一 Fiber 树构建，支持从高层往下检测。

```go
// HitMap 保持不变，但需要确保支持 NodeID 和 Layer
type HitMap struct {
    entries []HitMapEntry
    root    layout.Node
    buildTime time.Time
}

type HitMapEntry struct {
    NodeID  uint64          // ✨ 从 ComputedBox.NodeID 拷贝
    Node    layout.Node
    Bounds  layout.Rect     // ✨ 从 ComputedBox.Box 拷贝
    LocalXY func(screenX, screenY int) (int, int)
    ZOrder  int             // ✨ 由 Layer 决定
    Layer   rtui.Layer      // ✨ 新增：Layer 信息
    Instance MsgHandler
}

// ✨ BuildHitMapFromFiber 从 Fiber 树构建 HitMap
// 新的 API，替代从多个 Layout 构建的方式
func BuildHitMapFromFiber(root *Fiber) *HitMap {
    if root == nil {
        return NewHitMap()
    }

    hm := &HitMap{
        entries:   make([]HitMapEntry, 0),
        buildTime: time.Now(),
    }

    // 遍历 Fiber 树，收集 ComputedBox
    hm.walkAndBuild(root, 0)

    // 按 Layer 和 Z-order 排序
    // 高层在前，确保 HitTest 优先命中高层
    hm.sortByLayerAndZOrder()

    return hm
}

// walkAndBuild 递归遍历 Fiber 树
func (hm *HitMap) walkAndBuild(fiber *Fiber, treeDepth int) {
    if fiber == nil {
        return
    }

    // 如果有 ComputedBox，创建 HitMapEntry
    if fiber.ComputedBox != nil {
        box := fiber.ComputedBox

        // ✨ 计算 Z-order
        // 层级越高，Z-order 越大
        zOrder := int(fiber.Layer) * 10000 + treeDepth
        // LayerBase (0) → zOrder: 0, 1, 2, ...
        // LayerModal (2) → zOrder: 20000, 20001, 20002, ...

        entry := HitMapEntry{
            NodeID: box.NodeID,
            Node:   rtui.AsLayoutNode(box.VNode),
            Bounds: layout.Rect{
                X:      box.Box.X,
                Y:      box.Box.Y,
                Width:  box.Box.Width,
                Height: box.Box.Height,
            },
            LocalXY: func(screenX, screenY int) (int, int) {
                return screenX - box.Box.X, screenY - box.Box.Y
            },
            ZOrder:  zOrder,
            Layer:   fiber.Layer,
            Instance: fiber.ComponentInstance,  // ✨ 添加 Instance 引用
        }

        hm.entries = append(hm.entries, entry)
    }

    // 递归处理子节点
    hm.walkAndBuild(fiber.Child, treeDepth+1)
    hm.walkAndBuild(fiber.Sibling, treeDepth)  // Sibling 在同一层级，treeDepth 不变
}

// sortByLayerAndZOrder 按 Layer 和 Z-order 排序
// 确保高层节点在前，HitTest 优先命中
func (hm *HitMap) sortByLayerAndZOrder() {
    sort.Slice(hm.entries, func(i, j int) bool {
        // 优先按 Layer 降序（高层在前）
        if hm.entries[i].Layer != hm.entries[j].Layer {
            return hm.entries[i].Layer > hm.entries[j].Layer
        }
        // Layer 相同，按 Z-order 降序
        return hm.entries[i].ZOrder > hm.entries[j].ZOrder
    })
}

// HitTest 保持不变，但现在是基于单一 Fiber 树
// 会优先命中高层节点
func (hm *HitMap) HitTest(x, y int) *HitMapEntry {
    // 从前向后遍历（已经按 Layer 降序排序，高层在前）
    for i := range hm.entries {
        entry := &hm.entries[i]
        if entry.Bounds.Contains(x, y) {
            return entry
        }
    }
    return nil
}
```

**关键变更**：
- ✅ 新增 `BuildHitMapFromFiber()` API
- ✅ 按 Layer 和 Z-order 排序
- ✅ HitTest 自动优先命中高层
- ✅ 不需要合并多个 HitMap

---

## 完整数据流设计

### 5.1 Render 流程

```text
┌────────────────────────────────────────────────────────────┐
│ Step 1: 用户触发重新渲染                                    │
│ ────────────────────────────────────────────────────────   │
│                                                              │
│  userComponent.SetState(newState)                           │
│    │                                                         │
│    ├─→ component.ScheduleUpdate(lane)                       │
│    └─→ reconciler.ScheduleUpdate(lane)                      │
│        │                                                     │
│        └─→ 将更新加入调度队列                                │
└────────────────────────────────────────────────────────────┘
                          ↓
┌────────────────────────────────────────────────────────────┐
│ Step 2: 调度器触发 Work Loop                                │
│ ────────────────────────────────────────────────────────   │
│                                                              │
│  scheduler.schedule(callback)                               │
│    │                                                         │
│    └─→ reconciler.workLoopSync()                            │
│        │                                                     │
│        └─→ 准备新的 Fiber 树                                │
└────────────────────────────────────────────────────────────┘
                          ↓
┌────────────────────────────────────────────────────────────┐
│ Step 3: Reconcile 阶段（Fiber 核心）                         │
│ ────────────────────────────────────────────────────────   │
│                                                              │
│  reconcile(oldFiberTree, newVNodeTree)                      │
│    │                                                         │
│    ├─→ reconcileChildren(parent, oldFiber, newVNodes)       │
│    │   │                                                     │
│    │   ├─→ 遍历 newVNodes                                   │
│    │   │   │                                                 │
│    │   │   ├─→ 根据 DiffKey 在 oldFiber 中匹配              │
│    │   │   │   │                                             │
│    │   │   │   ├─→ ✅ 匹配成功：复用 Fiber                   │
│    │   │   │   │       │                                    │
│    │   │   │   │       ├─→ fiber.NodeID 保持不变            │
│    │   │   │   │       ├─→ fiber.Type 更新                  │
│    │   │   │   │       ├─→ fiber.DiffKey 更新                │
│    │   │   │   │       ├─→ fiber.Props 更新                 │
│    │   │   │   │       ├─→ fiber.Layer 更新（从 VNode 拷贝） │
│    │   │   │   │       └─→ fiber.Alternate = oldFiber      │
│    │   │   │   │                                            │
│    │   │   │   └─→ ❌ 匹配失败：创建新 Fiber                 │
│    │   │   │       │                                        │
│    │   │   │       ├─→ fiber.NodeID = allocator.Next()      │
│    │   │   │       ├─→ fiber.Type = vnode.Type()            │
│    │   │   │       ├─→ fiber.DiffKey = vnode.Key()          │
│    │   │   │       ├─→ fiber.Props = vnode.Props()          │
│    │   │   │       ├─→ fiber.Layer = vnode.GetLayer()       │
│    │   │   │       └─→ fiber.Alternate = nil               │
│    │   │   │                                                │
│    │   │   └─→ 连接 Fiber 树结构                            │
│    │   │       │                                            │
│    │   │       ├─→ fiber.Parent = parent                    │
│    │   │       ├─→ fiber.Child, sibling 连接               │
│    │   │       └─→ fiber.Flags = Placement/Update          │
│    │   │                                                    │
│    │   └─→ 生成 newFiberTree                                │
│    │                                                         │
│    └─→ 返回 newFiberRoot                                    │
└────────────────────────────────────────────────────────────┘
                          ↓
                          newFiberTree（完整单一树）
                          ├─ Normal nodes (LayerBase)
                          ├─ Modal nodes (LayerModal)
                          ├─ Overlay nodes (LayerOverlay)
                          ├─ Tooltip nodes (LayerTooltip)
                          └─ Inspector nodes (LayerInspector)
                          ↓
┌────────────────────────────────────────────────────────────┐
│ Step 4: Layout 阶段（基于 Fiber）                            │
│ ────────────────────────────────────────────────────────   │
│                                                              │
│  Layout(newFiberTree, constraints)                         │
│    │                                                         │
│    ├─→ layoutFiber(fiber, constraints)                     │
│    │   │                                                     │
│    │   ├─→ 计算自身的尺寸                                   │
│    │   │   │                                                 │
│    │   │   ├─→ 根据约束计算                                 │
│    │   │   ├─→ 考虑 padding, margin                          │
│    │   │   └─→ 计算 width/height                           │
│    │   │                                                    │
│    │   ├─→ 更新 fiber.ComputedBox                           │
│    │   │   │                                                 │
│    │   │   ├─ computedBox.NodeID = fiber.NodeID            │
│    │   │   ├─ computedBox.Box = {X, Y, Width, Height}     │
│    │   │   ├─ computedBox.Layer = fiber.Layer             │
│    │   │   ├─ computedBox.VNode = fiber.VNode             │
│    │   │   └─ computedBox.Children = []                    │
│    │   │                                                    │
│    │   ├─→ 递归处理子节点                                   │
│    │   │   │                                                 │
│    │   │   └─→ layoutFiber(fiber.Child, constraints)       │
│    │   │                                                    │
│    │   └─→ 递归处理兄弟节点                                 │
│    │       │                                                 │
│    │       └─→ layoutFiber(fiber.Sibling, constraints)     │
│    │                                                         │
│    └─→ 生成 ComputedBox 树（附在 Fiber 树上）              │
└────────────────────────────────────────────────────────────┘
                          ↓
                          Fiber 树附带了 ComputedBox 树
                          ↓
┌────────────────────────────────────────────────────────────┐
│ Step 5: Build RenderPlanes 渲染投影层                        │
│ ────────────────────────────────────────────────────────   │
│                                                              │
│  renderPlanes := NewRenderPlanes()                          │
│  renderPlanes.BuildFromFiber(newFiberTree)                │
│    │                                                         │
│    ├─→ walkAndCollect(fiber)                               │
│    │   │                                                     │
│    │   ├─→ 如果 fiber.ComputedBox != nil                    │
│    │   │   │                                                 │
│    │   │   ├─→ layer := fiber.Layer                         │
│    │   │   ├─→ renderPlanes.Planes[layer].append(computedBox)
│    │   │   │                                                 │
│    │   │   └─→ （不修改 Fiber，只收集引用）                │
│    │   │                                                    │
│    │   ├─→ walkAndCollect(fiber.Child)                     │
│    │   └─→ walkAndCollect(fiber.Sibling)                   │
│    │                                                         │
│    ├─→ sortPlanes()                                         │
│    │   │                                                     │
│    │   └─→ 对每个 Plane 中的 ComputedBox 排序（按位置）     │
│    │                                                         │
│    └─→ 生成 RenderPlanes（分桶视图）                        │
│        │                                                     │
│        ├─ renderPlanes.Planes[LayerBase]    = [...]        │
│        ├─ renderPlanes.Planes[LayerOverlay] = [...]        │
│        ├─ renderPlanes.Planes[LayerModal]   = [...]        │
│        ├─ renderPlanes.Planes[LayerTooltip] = [...]        │
│        └─ renderPlanes.Planes[LayerInspector] = [...]     │
└────────────────────────────────────────────────────────────┘
                          ↓
┌────────────────────────────────────────────────────────────┐
│ Step 6: Render 阶段（按 Layer 顺序）                         │
│ ────────────────────────────────────────────────────────   │
│                                                              │
│  for _, layer := range renderPlanes.RenderOrder {          │
│      boxes := renderPlanes.Planes[layer]                    │
│      for _, box := range boxes {                            │
│          renderer.Render(box.VNode, box.Box.X, box.Box.Y, buffer) │
│      }                                                      │
│  }                                                           │
│                                                              │
│ 渲染顺序：                                                   │
│  1. LayerBase           (Z: 0)                              │
│  2. LayerOverlay        (Z: 1)                              │
│  3. LayerModal          (Z: 2)                              │
│  4. LayerTooltip        (Z: 3)                              │
│  5. LayerInspector      (Z: 4)                              │
└────────────────────────────────────────────────────────────┘
                          ↓
┌────────────────────────────────────────────────────────────┐
│ Step 7: Build HitMap 事件映射层                              │
│ ────────────────────────────────────────────────────────   │
│                                                              │
│  hitMap := event.BuildHitMapFromFiber(newFiberTree)       │
│    │                                                         │
│    ├─→ walkAndBuild(fiber, treeDepth)                      │
│    │   │                                                     │
│    │   ├─→ 如果 fiber.ComputedBox != nil                    │
│    │   │   │                                                 │
│    │   │   ├─→ zOrder := int(fiber.Layer) * 10000 + treeDepth │
│    │   │   │                                                 │
│    │   │   ├─→ entry := HitMapEntry{                        │
│    │   │   │         NodeID:  box.NodeID,                   │
│    │   │   │         Bounds:  box.Box,                      │
│    │   │   │         ZOrder:  zOrder,                       │
│    │   │   │         Layer:   fiber.Layer,                  │
│    │   │   │         Instance: fiber.ComponentInstance,     │
│    │   │   │     }                                          │
│    │   │   │                                                 │
│    │   │   └─→ hm.entries.append(entry)                    │
│    │   │                                                    │
│    │   ├─→ walkAndBuild(fiber.Child, treeDepth+1)          │
│    │   └─→ walkAndBuild(fiber.Sibling, treeDepth)          │
│    │                                                         │
│    ├─→ sortByLayerAndZOrder()                               │
│    │   │                                                     │
│    │   └─→ 按 Layer 降序，Z-order 降序排序                  │
│    │                                                         │
│    └─→ 生成 HitMap                                          │
└────────────────────────────────────────────────────────────┘
                          ↓
                          HitMap（全部节点，已按 Layer 排序）
                          ↓
┌────────────────────────────────────────────────────────────┐
│ Step 8: Event Dispatch 事件分发                             │
│ ────────────────────────────────────────────────────────   │
│                                                              │
│  用户点击 (x, y)                                            │
│    │                                                         │
│    ├─→ entry := hitMap.HitTest(x, y)                       │
│    │   │                                                     │
│    │   ├─→ 从前向后遍历 hitMap.entries                       │
│    │   │   │                                                 │
│    │   │   ├─→ 如果 entry.Bounds.Contains(x, y)             │
│    │   │   │       │                                         │
│    │   │   │       ├─→ ✅ 找到！返回 entry                  │
│    │   │   │       │   （因为是按 Layer 降序排序，优先命中高层）│
│    │   │   │       │                                         │
│    │   │   │       └─→ （如果命中了 Modal，不会再检测 Base）  │
│    │   │   │                                                 │
│    │   │   └─→ 继续检测下一项                               │
│    │   │                                                    │
│    │   └─→ 未命中返回 nil                                   │
│    │                                                         │
│    └─→ if entry != nil {                                   │
│            // 基于 NodeID 找到 Fiber                        │
│            fiber := runtime.FindFiberByNodeID(entry.NodeID) │
│        │                                                     │
│        // 基于 Fiber 找到 Instance                         │
│        instance := fiber.ComponentInstance                  │
│            │                                                 │
│            // 触发事件处理器                                │
│            instance.Handle(event.MouseEvent{X: x, Y: y})    │
│            │                                                 │
│            // 冒泡到父节点                                  │
│            for f := fiber.Parent; f != nil; f = f.Parent {  │
│                if f.ComponentInstance != nil {             │
│                    f.ComponentInstance.Handle(event.MouseEvent{X: x, Y: y}) │
│                }                                             │
│            }                                                 │
│        }                                                     │
└────────────────────────────────────────────────────────────┘
```

---

## 详细重构方案

### 6.1 Phase 1: 核心数据结构更新

**文件**: `runtime/ui/fiber.go`

#### 6.1.1 添加 Layer 字段

```go
type Fiber struct {
    // ... 现有字段 ...

    // ========== 新增字段 ==========
    // Layer 声明值（从 VNode 拷贝）
    // 默认值：LayerBase
    // 不参与 diff
    Layer rtui.Layer

    // ComputedBox 保存 Layout 结果
    ComputedBox *compute.ComputedBox
}
```

#### 6.1.2 更新 CreateFiber

```go
func CreateFiber(vnode rtui.VNode) *Fiber {
    fiber := &Fiber{
        NodeID:  0,  // 将由 reconciler 分配
        Type:    vnode.Type(),
        DiffKey: vnode.Key(),
        Key:     vnode.Key(),
        Props:   vnode.Props(),

        // ✨ 拷贝 Layer 声明值
        Layer: func() rtui.Layer {
            layer := vnode.GetLayer()
            if layer < rtui.LayerBase || layer > rtui.LayerInspector {
                return rtui.LayerBase
            }
            return layer
        }(),

        ComputedBox: nil,  // Layout 阶段赋值
    }

    // 设置 VNode 引用
    fiber.VNode = vnode

    return fiber
}
```

#### 6.1.3 更新 CloneFiber

```go
func CloneFiber(current *Fiber) *Fiber {
    clone := &Fiber{
        NodeID:        current.NodeID,     // ✅ 保持 NodeID 不变
        Type:          current.Type,
        DiffKey:       current.DiffKey,
        Key:           current.Key,
        Props:         current.MemoizedProps,  // 使用 MemoizedProps
        Layer:         current.Layer,           // ✨ 拷贝 Layer
        MemoizedProps: current.MemoizedProps,
        MemoizedState: current.MemoizedState,
        Alternate:     current,
        ComputedBox:   nil,  // Layout 阶段重新计算
    }

    clone.VNode = current.VNode  // 保持 VNode 引用

    clone.Flags = Update          // 标记为更新

    // 复制 UpdateQueue
    if current.UpdateQueue != nil {
        clone.UpdateQueue = current.UpdateQueue.Clone()
    }

    return clone
}
```

---

### 6.2 Phase 2: Reconciler 更新

**文件**: `internal/reconciler/reconciler.go`, `internal/reconciler/diff.go`

#### 6.2.1 reconcileChildren 保持不变（基于 DiffKey）

```go
// reconcileChildren 已经正确使用 DiffKey，无需修改
func reconcileChildren(returnFiber *Fiber, currentFirstChild *Fiber, newChildren []rtui.VNode, lanes Lane) *Fiber {
    // ✅ 基于 DiffKey 匹配，不基于 NodeID
    // 这符合 diff_key.md 的设计

    // ... 现有实现 ...
}
```

#### 6.2.2 shouldUpdate 保持不变

```go
// shouldUpdate 已经正确使用 DiffKey，无需修改
func shouldUpdate(current *Fiber, vnode rtui.VNode) bool {
    // ✅ 使用 DiffKey 比较，不使用 NodeID
    if current.DiffKey != vnode.Key() {
        return false
    }

    if current.Type != vnode.Type() {
        return false
    }

    return true
}
```

#### 6.2.3 更新 complete_work - 从 VNode 拷贝 Layer

**文件**: `internal/reconciler/complete_work.go`

```go
func CompleteWork(current, workInProgress *Fiber) *Fiber {
    // ... 现有逻辑 ...

    // ✨ 从 VNode 拷贝 Layer
    workInProgress.Layer = workInProgress.VNode.GetLayer()

    // ... 现有逻辑 ...

    return workInProgress
}
```

---

### 6.3 Phase 3: Layout Engine 更新

**文件**: `runtime/compute/engine.go`

#### 6.3.1 更新 Layout API

```go
// ✨ 新的 Layout API
// 直接基于 Fiber 树进行布局，不再需要多个 VNode 树

func (e *Engine) Layout(
    vnode rtui.VNode,
    fiber *reconciler.Fiber,
    constraints runtime.BoxConstraints,
) (*ComputedLayout, error) {
    e.mux.Lock()
    defer e.mux.Unlock()

    // 重置计算器状态
    e.nodeAllocator = 0
    e.stats = Statistics{}

    // ✨ 从 Fiber 而不是 VNode 开始布局
    // 这确保 Layout 结果直接附加到 Fiber 上
    rootBox := e.layoutFiber(fiber, constraints, 0)

    // 构建 ComputedLayout
    layout := &ComputedLayout{
        Root:   rootBox,
        Stats:  e.stats,
        HitMap: nil,  // 稍后单独构建
    }

    // ✨ 构建 HitMap（基于 Fiber）
    // 替代原来从多个 Layout 构建 HitMap 的方式
    layout.HitMap = e.buildHitMapFromFiber(fiber)

    return layout, nil
}
```

#### 6.3.2 新增 layoutFiber 方法

```go
// ✨ layoutFiber 基于 Fiber 进行布局
// 结果直接附加到 Fiber.ComputedBox

func (e *Engine) layoutFiber(
    fiber *reconciler.Fiber,
    constraints runtime.BoxConstraints,
    depth int,
) *ComputedBox {
    if fiber == nil {
        return nil
    }

    // 计算自身尺寸
    box := e.measureFiber(fiber, constraints)

    // ✨ 创建/更新 fiber.ComputedBox
    computedBox := &ComputedBox{
        NodeID: fiber.NodeID,
        Box:    box,
        Layer:  fiber.Layer,  // ✨ 拷贝 Layer
        VNode:  fiber.VNode,
        Children: []*ComputedBox{},
    }

    // ✨ 保存回 Fiber
    fiber.ComputedBox = computedBox

    // 递归处理子节点
    childConstraints := e.getChildConstraints(fiber, box)
    computedBox.Children = e.layoutFiberChildren(fiber.Child, childConstraints, depth+1)

    // 处理兄弟节点
    e.layoutFiberSibling(fiber.Sibling, constraints, depth)

    return computedBox
}

func (e *Engine) measureFiber(
    fiber *reconciler.Fiber,
    constraints runtime.BoxConstraints,
) Box {
    // ... 根据 fiber.VNode.Type() 和 Props 计算尺寸 ...

    // ⚠️ 注意：不再需要从 VNode.Bounds 读取
    // 因为 Bounds 是 Layout 结果，不是输入

    return Box{X: 0, Y: 0, Width: w, Height: h}
}

func (e *Engine) layoutFiberChildren(
    fiber *reconciler.Fiber,
    constraints runtime.BoxConstraints,
    depth int,
) []*ComputedBox {
    var children []*ComputedBox
    child := fiber

    for child != nil {
        box := e.layoutFiber(child, constraints, depth)
        if box != nil {
            children = append(children, box)
        }
        child = child.Sibling
    }

    return children
}

func (e *Engine) layoutFiberSibling(
    fiber *reconciler.Fiber,
    constraints runtime.BoxConstraints,
    depth int,
) {
    if fiber == nil {
        return
    }

    // Sibling 在同一层级，使用相同的 constraints
    e.layoutFiber(fiber, constraints, depth)
}
```

#### 6.3.3 新增 buildHitMapFromFiber

```go
// ✨ buildHitMapFromFiber 从 Fiber 树构建 HitMap
// 替代原来从多个 Layout 构建 HitMap 的方式

func (e *Engine) buildHitMapFromFiber(root *reconciler.Fiber) *event.HitMap {
    if root == nil {
        return event.NewHitMap()
    }

    var entries []event.HitMapEntryInternal

    // 遍历 Fiber 树，收集 ComputedBox
    e.walkFiberForHitMap(root, 0, &entries)

    // 构建 HitMap
    return event.BuildHitMapFromEntries(entries)
}

func (e *Engine) walkFiberForHitMap(
    fiber *reconciler.Fiber,
    treeDepth int,
    entries *[]event.HitMapEntryInternal,
) {
    if fiber == nil || fiber.ComputedBox == nil {
        return
    }

    box := fiber.ComputedBox

    // ✨ 计算 Z-order
    // 层级越高，Z-order 越大
    zOrder := int(fiber.Layer) * 10000 + treeDepth

    entry := event.HitMapEntryInternal{
        NodeID: box.NodeID,
        Node:   rtui.AsLayoutNode(box.VNode),
        Bounds: layout.Rect{
            X:      box.Box.X,
            Y:      box.Box.Y,
            Width:  box.Box.Width,
            Height: box.Box.Height,
        },
        LocalXY: func(screenX, screenY int) (int, int) {
            return screenX - box.Box.X, screenY - box.Box.Y
        },
        ZOrder:  zOrder,
        Layer:   fiber.Layer,
        Instance: fiber.ComponentInstance,
    }

    *entries = append(*entries, entry)

    // 递归处理子节点
    e.walkFiberForHitMap(fiber.Child, treeDepth+1, entries)
    e.walkFiberForHitMap(fiber.Sibling, treeDepth, entries)
}
```

---

### 6.4 Phase 4: 废弃 StripLayers，引入 RenderPlane

**文件**: `runtime/layer/manager.go`

#### 6.4.1 标记旧方法为 Deprecated

```go
// =============================================================================
// ❌ DEPRECATED: StripLayers 相关方法
// 在新架构中，这些方法不再使用，将被移除
// =============================================================================

// Deprecated: 使用 BuildRenderPlanes 替代
func (m *Manager) CollectAndLayout(...) error {
    // This method will be removed
    return errors.New("StripLayers is deprecated, use BuildRenderPlanes")
}

// Deprecated: 这个方法不再需要
func (c *Collector) StripLayers(vnode rtui.VNode) rtui.VNode {
    // This method will be removed
    return nil
}

// Deprecated: 这个方法不再需要
func (c *Collector) cloneWithoutLayers(vnode rtui.VNode) rtui.VNode {
    // This method will be removed
    return nil
}
```

#### 6.4.2 引入新的 RenderPlane

```go
// =============================================================================
// ✨ RenderPlane: 新的多层渲染机制
// =============================================================================

// RenderPlanes 按 Layer 分桶的 ComputedBox
type RenderPlanes struct {
    // 按 Layer 分桶的 ComputedBox
    Planes map[rtui.Layer][]*compute.ComputedBox

    // 渲染顺序（从低到高）
    RenderOrder []rtui.Layer
}

// NewRenderPlanes 创建新的 RenderPlanes
func NewRenderPlanes() *RenderPlanes {
    return &RenderPlanes{
        Planes: make(map[rtui.Layer][]*compute.ComputedBox),
        RenderOrder: []rtui.Layer{
            rtui.LayerBase,
            rtui.LayerOverlay,
            rtui.LayerModal,
            rtui.LayerTooltip,
            rtui.LayerInspector,
        },
    }
}

// BuildFromFiber 从 Fiber 树构建 RenderPlanes
// ✨ 这是新的核心 API
// 替代 StripLayers + 多路径 Layout
func (rp *RenderPlanes) BuildFromFiber(root *Fiber) {
    rp.Planes = make(map[rtui.Layer][]*compute.ComputedBox)

    // ✨ 遍历 Fiber 树，按 Layer 分桶
    rp.walkAndCollect(root)

    // ✨ 按位置排序每个 Plane
    rp.sortPlanes()
}

// walkAndCollect 递归遍历 Fiber 树，收集 ComputedBox
func (rp *RenderPlanes) walkAndCollect(fiber *Fiber) {
    if fiber == nil {
        return
    }

    // 如果有 ComputedBox，添加到对应的 Plane
    if fiber.ComputedBox != nil {
        layer := fiber.Layer
        if layer < rtui.LayerBase {
            layer = rtui.LayerBase
        }
        rp.Planes[layer] = append(rp.Planes[layer], fiber.ComputedBox)
    }

    // 递归处理子节点
    rp.walkAndCollect(fiber.Child)
    rp.walkAndCollect(fiber.Sibling)
}

// sortPlanes 对每个 Plane 中的 ComputedBox 按位置排序
func (rp *RenderPlanes) sortPlanes() {
    for layer, boxes := range rp.Planes {
        sort.Slice(boxes, func(i, j int) bool {
            // 按 Y 排序，Y 相同按 X 排序（确保渲染顺序一致）
            if boxes[i].Box.Y != boxes[j].Box.Y {
                return boxes[i].Box.Y < boxes[j].Box.Y
            }
            return boxes[i].Box.X < boxes[j].Box.X
        })
        rp.Planes[layer] = boxes
    }
}

// GetPlane 获取指定 Layer 的所有 ComputedBox
func (rp *RenderPlanes) GetPlane(layer rtui.Layer) []*compute.ComputedBox {
    return rp.Planes[layer]
}

// Iterate 遍历所有 Layer 的所有 ComputedBox（按渲染顺序）
func (rp *RenderPlanes) Iterate(fn func(layer rtui.Layer, box *compute.ComputedBox) bool) {
    for _, layer := range rp.RenderOrder {
        for _, box := range rp.Planes[layer] {
            if !fn(layer, box) {
                return
            }
        }
    }
}

// RenderPlanes 返回完整的分桶结果
func (rp *RenderPlanes) AllPlanes() map[rtui.Layer][]*compute.ComputedBox {
    return rp.Planes
}

// HasLayer 检查指定 Layer 是否有内容
func (rp *RenderPlanes) HasLayer(layer rtui.Layer) bool {
    return len(rp.Planes[layer]) > 0
}

// GetHighestLayer 返回最高非空 Layer
func (rp *RenderPlanes) GetHighestLayer() rtui.Layer {
    for i := len(rp.RenderOrder) - 1; i >= 0; i-- {
        layer := rp.RenderOrder[i]
        if len(rp.Planes[layer]) > 0 {
            return layer
        }
    }
    return rtui.LayerBase
}
```

#### 6.4.3 更新 LayerManager（新接口）

```go
// =============================================================================
// ✨ LayerManager: 新的 Layer 管理接口
// 基于 RenderPlane，废弃 StripLayers
// =============================================================================

type Manager struct {
    // ✨ 新字段：RenderPlanes
    renderPlanes *RenderPlanes

    // ❌ 旧字段：废弃
    // collector *Collector
    // layouts   LayerLayouts
}

// NewManager 创建新的 LayerManager
func NewManager() *Manager {
    return &Manager{
        renderPlanes: NewRenderPlanes(),
    }
}

// ✨ BuildRenderPlanes 从 Fiber 树构建 RenderPlanes
// 这是新的主要入口点
// 替代旧的 CollectAndLayout
func (m *Manager) BuildRenderPlanes(root *Fiber) *RenderPlanes {
    // 直接从 Fiber 树构建
    m.renderPlanes.BuildFromFiber(root)
    return m.renderPlanes
}

// GetRenderPlanes 获取 RenderPlanes
func (m *Manager) GetRenderPlanes() *RenderPlanes {
    return m.renderPlanes
}

// HasModal 检查是否有 Modal
func (m *Manager) HasModal() bool {
    return m.renderPlanes.HasLayer(rtui.LayerModal)
}

// GetHighestLayer 获取最高非空 Layer
func (m *Manager) GetHighestLayer() rtui.Layer {
    return m.renderPlanes.GetHighestLayer()
}

// ❌ 旧方法：废弃
// func (m *Manager) CollectAndLayout(...) error { ... }
// func (m *Manager) GetLayouts() LayerLayouts { ... }
// func (m *Manager) GetMergedHitMap() *event.HitMap { ... }
```

---

### 6.5 Phase 5: Render 更新

**文件**: `internal/render/vnode_renderer.go`

#### 6.5.1 按顺序渲染 RenderPlanes

```go
// ✨ 基于 RenderPlanes 渲染
// 按渲染顺序遍历所有 Layer

func (r *FiberRenderer) Render(
    vnode rtui.VNode,
    x, y int,
    buffer interface{},
) {
    // 1. 获取 Fiber 根节点
    fiberRoot := r.reconciler.GetFiberRoot()
    if fiberRoot == nil {
        return
    }

    // 2. 获取 LayerManager
    layerManager := r.app.GetLayerManager()
    if layerManager == nil {
        // 没有分层支持，直接渲染
        r.renderFiber(fiberRoot, x, y, buffer)
        return
    }

    // 3. 构建 RenderPlanes
    renderPlanes := layerManager.BuildRenderPlanes(fiberRoot)

    // 4. 按 Layer 顺序渲染
    // 渲染顺序：Base → Overlay → Modal → Tooltip → Inspector
    for _, layer := range renderPlanes.RenderOrder {
        boxes := renderPlanes.GetPlane(layer)

        for _, box := range boxes {
            // 渲染这个 box
            r.renderComputedBox(box, buffer)
        }
    }
}

func (r *FiberRenderer) renderComputedBox(
    box *compute.ComputedBox,
    buffer interface{},
) {
    if box == nil || box.VNode == nil {
        return
    }

    // 基于 VNode 渲染
    r.renderVNode(box.VNode, box.Box.X, box.Box.Y, buffer)

    // 递归渲染子节点
    for _, child := range box.Children {
        r.renderComputedBox(child, buffer)
    }
}
```

---

### 6.6 Phase 6: Event 更新

**文件**: `runtime/event/hitmap.go`

#### 6.6.1 添加新的 API

```go
// ✨ BuildHitMapFromFiber 从 Fiber 树构建 HitMap
// 这是新的主要入口点
// 替代旧的从多个 Layout 构建 HitMap 的方式

func BuildHitMapFromFiber(root *Fiber) *HitMap {
    if root == nil {
        return NewHitMap()
    }

    hm := &HitMap{
        entries:   make([]HitMapEntry, 0),
        buildTime: time.Now(),
    }

    // 遍历 Fiber 树，收集 ComputedBox
    hm.walkAndBuild(root, 0)

    // 按 Layer 和 Z-order 排序（高层优先）
    hm.sortByLayerAndZOrder()

    return hm
}

// walkAndBuild 递归遍历 Fiber 树
func (hm *HitMap) walkAndBuild(fiber *Fiber, treeDepth int) {
    if fiber == nil || fiber.ComputedBox == nil {
        return
    }

    box := fiber.ComputedBox

    // 计算 Z-order
    // 层级越高，Z-order 越大
    zOrder := int(fiber.Layer) * 10000 + treeDepth

    entry := HitMapEntry{
        NodeID:  box.NodeID,
        Node:    rtui.AsLayoutNode(box.VNode),
        Bounds: layout.Rect{
            X:      box.Box.X,
            Y:      box.Box.Y,
            Width:  box.Box.Width,
            Height: box.Box.Height,
        },
        LocalXY: func(screenX, screenY int) (int, int) {
            return screenX - box.Box.X, screenY - box.Box.Y
        },
        ZOrder:   zOrder,
        Layer:    fiber.Layer,
        Instance: fiber.ComponentInstance,
    }

    hm.entries = append(hm.entries, entry)

    // 递归处理子节点
    hm.walkAndBuild(fiber.Child, treeDepth+1)
    hm.walkAndBuild(fiber.Sibling, treeDepth)
}

// sortByLayerAndZOrder 按 Layer 和 Z-order 排序
// 确保 HitTest 优先命中高层
func (hm *HitMap) sortByLayerAndZOrder() {
    sort.Slice(hm.entries, func(i, j int) bool {
        // 优先按 Layer 降序（高层在前）
        if hm.entries[i].Layer != hm.entries[j].Layer {
            return hm.entries[i].Layer > hm.entries[j].Layer
        }
        // Layer 相同，按 Z-order 降序
        return hm.entries[i].ZOrder > hm.entries[j].ZOrder
    })
}
```

#### 6.6.2 HitTest 保持不变（但行为改变）

```go
// HitTest 保持不变，但因为 entries 已经按 Layer 排序
// 所以会自动优先命中高层

func (hm *HitMap) HitTest(x, y int) *HitMapEntry {
    // 从前向后遍历（已经按 Layer 降序排序，高层在前）
    for i := range hm.entries {
        entry := &hm.entries[i]
        if entry.Bounds.Contains(x, y) {
            return entry
        }
    }
    return nil
}
```

---

### 6.7 Phase 7: App 更新

**文件**: `framework/app.go`

#### 6.7.1 更新 Render 流程

```go
func (a *App) render(ctx component.PaintContext) {
    // ... 现有逻辑 ...

    // ✨ Reconcile 生成新的 Fiber 树
    newFiberRoot := a.reconciler.workLoopSync()

    // ✨ 从 Fiber 树进行 Layout
    // Layout 结果直接附加到 Fiber.ComputedBox
    layout, err := a.engine.Layout(a.rootComponent, newFiberRoot, a.constraints)
    if err != nil {
        log.RenderLogger.Error("[render] Layout error: %v", err)
        return
    }

    // ✨ 构建 RenderPlanes（基于 Fiber 树）
    // 替代旧的 CollectAndLayout
    layerManager := a.GetLayerManager()
    renderPlanes := layerManager.BuildRenderPlanes(newFiberRoot)

    // ✨ 渲染（按 Layer 顺序）
    a.renderer.RenderWithRenderPlanes(renderPlanes, ctx.Buffer)

    // ✨ 构建 HitMap（基于 Fiber 树）
    // 替代旧的 GetMergedHitMap
    hitMap := event.BuildHitMapFromFiber(newFiberRoot)
    a.hitMap = hitMap

    // ... 现有逻辑 ...
}
```

---

## 破坏性变更清单

### 7.1 API 变更

#### 废弃的 API

```go
// ❌ 废弃：runtime/layer/manager.go

// StripLayers 不再使用
func (m *Manager) CollectAndLayout(...) error

// ❌ 废弃：runtime/layer/collector.go

// 这个方法不再需要
func (c *Collector) StripLayers(vnode rtui.VNode) rtui.VNode

// 这方法不再需要
func (c *Collector) cloneWithoutLayers(vnode rtui.VNode) rtui.VNode

// ❌ 废弃：runtime/layer/manager.go

// 这个方法不再需要（HitMap 现在基于单一 Fiber 树）
func (m *Manager) GetMergedHitMap() *event.HitMap

// ❌ 废弃：runtime/layer/manager.go

// 这个方法不再需要（RenderPlanes 替代）
func (m *Manager) GetLayouts() LayerLayouts

// ❌ 废弃：runtime/ui/vnode.go

// 如果 VNode 有 SetBounds() 方法，需要废弃
func (v VNode) SetBounds(x, y, width, height int)

// 如果 VNode 有 GetBounds() 方法，需要废弃
func (v VNode) GetBounds() [4]int
```

#### 新增的 API

```go
// ✨ 新增：runtime/layer/manager.go

// 从 Fiber 树构建 RenderPlanes
func (m *Manager) BuildRenderPlanes(root *Fiber) *RenderPlanes

// 获取 RenderPlanes
func (m *Manager) GetRenderPlanes() *RenderPlanes

// ✨ 新增：runtime/layer/manager.go

// RenderPlanes 类型
type RenderPlanes struct {
    Planes      map[rtui.Layer][]*compute.ComputedBox
    RenderOrder []rtui.Layer
}

// 从 Fiber 树构建
func (rp *RenderPlanes) BuildFromFiber(root *Fiber)

// 按 Layer 获取
func (rp *RenderPlanes) GetPlane(layer rtui.Layer) []*compute.ComputedBox

// 按渲染顺序遍历
func (rp *RenderPlanes) Iterate(fn func(layer rtui.Layer, box *compute.ComputedBox) bool)

// ✨ 新增：runtime/compute/engine.go

// 基于 Fiber 树进行 Layout
func (e *Engine) Layout(
    vnode rtui.VNode,
    fiber *Fiber,
    constraints BoxConstraints,
) (*ComputedLayout, error)

// ✨ 新增：runtime/event/hitmap.go

// 从 Fiber 树构建 HitMap
func BuildHitMapFromFiber(root *Fiber) *HitMap

// ✨ 新增：runtime/ui/fiber.go

// Fiber 新增字段
type Fiber struct {
    Layer       rtui.Layer  // Layer 声明值
    ComputedBox *compute.ComputedBox  // Layout 结果
}
```

### 7.2 行为变更

#### 变更 1：Modal 布局不再独立

**旧行为**：
```go
// Modal 有独立的 Layout 树
modalLayout := engine.Layout(modalVNode, modalConstraints)

// Modal 布局不受父组件约束影响
// 因为它被 Strip 到了独立的树
```

**新行为**：
```go
// Modal 在 Fiber 树中，与父组件在同一棵树
// Layout 时考虑完整父子关系
layout := engine.Layout(rootVNode, rootFiber, constraints)

// Modal 布局仍然可以独立调整位置
// 通过 layoutLayer() 的 transform 逻辑
```

**影响**：
- Modal 布局可能稍有变化（因为约束传递方式不同）
- Modal 居中逻辑需要调整（从 transform 方式改为显式设置位置）

#### 变更 2：HitMap 基于 Fiber 树

**旧行为**：
```go
// HitMap 从多个独立的 Layout 合并
baseHitMap := engine.Layout(baseTree).HitMap
modalHitMap := engine.Layout(modalTree).HitMap
mergedHitMap := layerManager.Merge(baseHitMap, modalHitMap)
```

**新行为**：
```go
// HitMap 直接从单一 Fiber 树构建
hitMap := event.BuildHitMapFromFiber(fiberRoot)
```

**影响**：
- HitTest 行为保持一致，但实现更简单
- 不需要手动管理 Z-order
- 不需要合并多个 HitMap

#### 变更 3：Debug 输出不同

**旧行为**：
```
=== LayerManager Layouts ===
Base Layout:
  - Button {X:10, Y:10, W:100, H:30}
Modal Layout:
  - ModalBox {X:100, Y:50, W:200, H:150}
```

**新行为**：
```
=== RenderPlanes ===
LayerBase:
  - Button {NodeID:3, X:10, Y:10, W:100, H:30}
LayerModal:
  - ModalBox {NodeID:5, X:100, Y:50, W:200, H:150}
```

**影响**：
- Debug 输出结构不同
- 所有节点都有 NodeID 标识

### 7.3 兼容性影响

#### 不兼容的用例

**用例 1：手动 StripLayers**
```go
// ❌ 旧代码（不再工作）
baseTree := layerManager.CollectAndLayout(vnode, fiber, constraints)
modalNodes := layerManager.GetModalNodes()

// ✅ 新代码
renderPlanes := layerManager.BuildRenderPlanes(fiberRoot)
modalBoxes := renderPlanes.GetPlane(rtui.LayerModal)
```

**用例 2：访问独立的 Layout 树**
```go
// ❌ 旧代码（不再工作）
baseLayout := layerManager.GetLayout(rtui.LayerBase)

// ✨ 新代码
baseBoxes := renderPlanes.GetPlane(rtui.LayerBase)
// 或者直接访问 Fiber.ComputedBox
```

**用例 3：获取 VNode.Bounds**
```go
// ❌ 旧代码（不再工作）
bounds := vnode.GetBounds()

// ✨ 新代码
// Bounds 是 Layout 结果，在 ComputedBox 中
fiber := runtime.FindFiberByVNode(vnode)
if fiber != nil && fiber.ComputedBox != nil {
    bounds := fiber.ComputedBox.Box
}
```

#### 需要更新的组件

**受影响的组件**：
- `runtime/layer/collector.go` - 重写或废弃
- `runtime/layer/manager.go` - 大量重构
- `internal/reconciler/reconciler.go` - 小幅修改
- `runtime/compute/engine.go` - 新增基于 Fiber 的 Layout
- `runtime/event/hitmap.go` - 新增 BuildHitMapFromFiber
- `internal/render/vnode_renderer.go` - 基于 RenderPlanes 渲染
- `framework/app.go` - 更新 Render 流程

**不受影响的组件**：
- `runtime/ui/vnode.go` - 结构基本不变（废弃 SetBounds 和 GetBounds）
- `runtime/ui/fiber.go` - 新增字段，但现有字段不变
- `internal/reconciler/diff.go` - 算法不变（基于 DiffKey）

---

## 实施步骤

### 8.1 阶段划分

#### Phase 1: 基础设施（1-2周）

**目标**：添加必要的字段和方法，不破坏现有功能

**任务**：
1. ✅ Fiber 新增 `Layer` 和 `ComputedBox` 字段
2. ✅ ComputedBox 新增 `NodeID`, `Layer`, `Children` 字段
3. ✅ Reconciler 从 VNode 拷贝 Layer 到 Fiber
4. ✅ 保持现有 Reconciler 算法不变（基于 DiffKey）
5. ✅ 添加 BuildHitMapFromFiber() API（与现有 API 并存）

**验收标准**：
- [ ] 所有现有测试通过
- [ ] Fiber.Layer 正确从 VNode.GetLayer() 拷贝
- [ ] BuildHitMapFromFiber() 与现有 HitMap 行为一致

#### Phase 2: Layout 重构（1-2周）

**目标**：Layout 基于 Fiber 而不是 VNode

**任务**：
1. ✅ Engine.Layout() 修改为基于 Fiber
2. ✅ 新增 layoutFiber() 方法
3. ✅ Layout 结果附加到 Fiber.ComputedBox
4. ✅ 废弃独立的 Layer Layout 路径
5. ✅ 更新所有 Layout 调用点

**验收标准**：
- [ ] Layout 结果正确（与 Phase 1 相同）
- [ ] Fiber.ComputedBox 正确填充
- [ ] Modal、Overlay 布局仍然正确

#### Phase 3: RenderPlane 引入（1周）

**目标**：引入 RenderPlane，不破坏现有功能

**任务**：
1. ✅ 添加 RenderPlanes 类型
2. ✅ 实现 BuildFromFiber() 方法
3. ✅ 在 LayerManager 中添加 BuildRenderPlanes() API
4. ✅ 与现有 CollectAndLayout() 并存

**验收标准**：
- [ ] RenderPlanes 正确分桶
- [ ] 现有 CollectAndLayout() 仍然工作
- [ ] 测试可以同时使用新旧两种方式

#### Phase 4: 废弃 StripLayers（1周）

**目标**：移除 StripLayers 相关代码

**任务**：
1. ✅ 标记 StripLayers 为 Deprecated
2. ✅ 标记 cloneWithoutLayers 为 Deprecated
3. ✅ 标记 CollectAndLayout 为 Deprecated
4. ✅ 移除所有 StripLayers 调用点

**验收标准**：
- [ ] 没有 StripLayers 调用
- [ ] 没有克隆 VNode
- [ ] 所有功能仍然正常

#### Phase 5: Render 更新（1周）

**目标**：Render 基于 RenderPlanes

**任务**：
1. ✅ 更新 FiberRenderer.Render()
2. ✅ 按 Layer 顺序渲染 RenderPlanes
3. ✅ 更新所有 Render 调用点

**验收标准**：
- [ ] 渲染顺序正确（Base → Overlay → Modal）
- [ ] Modal、Overlay、Tooltip 正确显示
- [ ] 无渲染错误

#### Phase 6: HitMap 更新（1周）

**目标**：HitMap 基于单一 Fiber 树

**任务**：
1. ✅ 移除 GetMergedHitMap() 方法
2. ✅ 所有 HitTest 使用 BuildHitMapFromFiber()
3. ✅ 验证 HitTest 正确性
4. ✅ 验证模态框事件处理

**验收标准**：
- [ ] HitTest 正确命中高层节点（Modal 优先）
- [ ] 事件冒泡正确
- [ ] Modal 点击处理正确

#### Phase 7: 清理和优化（1周）

**目标**：清理废弃代码，优化性能

**任务**：
1. ✅ 移除所有 Deprecated 代码
2. ✅ 清理 unused imports
3. ✅ 优化 RenderPlanes 性能
4. ✅ 添加更多单元测试
5. ✅ 更新文档

**验收标准**：
- [ ] 代码无 Deprecated 标记
- [ ] 测试覆盖率 > 80%
- [ ] 文档完整

#### Phase 8: 综合测试（1周）

**目标**：全面测试，确保功能完整

**任务**：
1. ✅ 运行所有现有测试
2. ✅ 添加集成测试
3. ✅ 性能测试
4. ✅ 用户场景测试

**验收标准**：
- [ ] 所有测试通过
- [ ] 无性能退化
- [ ] 用户场景验证通过

### 8.2 详细任务分解

#### Phase 1 第1周：数据结构更新

```bash
 Day 1-2: Fiber 结构更新
   - [ ] 添加 Fiber.Layer 字段
   - [ ] 添加 Fiber.ComputedBox 字段
   - [ ] 更新 CreateFiber()
   - [ ] 更新 CloneFiber()
   - [ ] 添加单元测试

 Day 3-4: Reconciler 更新
   - [ ] complete_work() 拷贝 Layer
   - [ ] 添加单元测试
   - [ ] 集成测试

 Day 5: HitMap API 添加
   - [ ] 添加 BuildHitMapFromFiber() API
   - [ ] 添加单元测试
   - [ ] 与现有 API 对比测试

 Day 6-7: 测试和修复
   - [ ] 运行所有测试
   - [ ] 修复问题
   - [ ] Code Review
```

#### Phase 2 第1周：Layout 重构

```bash
 Day 1-2: Engine.Layout() 更新
   - [ ] 修改 Layout() 签名
   - [ ] 添加 layoutFiber() 方法
   - [ ] 添加 buildHitMapFromFiber() 方法

 Day 3-4: 集成和测试
   - [ ] 更新所有 Layout 调用点
   - [ ] 添加集成测试
   - [ ] 验证布局正确性

 Day 5-7: 测试和修复
   - [ ] 运行所有测试
   - [ ] 修复 Layout 问题
   - [ ] Code Review
```

#### Phase 3 第1周：RenderPlane 引入

```bash
 Day 1-2: RenderPlanes 类型
   - [ ] 添加 RenderPlanes 类型
   - [ ] 实现 BuildFromFiber()
   - [ ] 实现 sortPlanes()
   - [ ] 添加单元测试

 Day 3-4: LayerManager 更新
   - [ ] 添加 BuildRenderPlanes() API
   - [ ] 保持 CollectAndLayout() 可用
   - [ ] 添加集成测试

 Day 5-7: 测试和验证
   - [ ] 验证 RenderPlanes 正确性
   - [ ] 对比新旧两种方式
   - [ ] Code Review
```

#### Phase 4 第1周：废弃 StripLayers

```bash
 Day 1-2: 标记 Deprecated
   - [ ] 标记 StripLayers 为 Deprecated
   - [ ] 标记 cloneWithoutLayers 为 Deprecated
   - [ ] 标记 CollectAndLayout 为 Deprecated

 Day 3-4: 移除调用点
   - [ ] 查找所有 StripLayers 调用
   - [ ] 替换为 BuildRenderPlanes()
   - [ ] 验证功能不变

 Day 5-7: 测试和修复
   - [ ] 运行所有测试
   - [ ] 修复问题
   - [ ] Code Review
```

#### Phase 5 第1周：Render 更新

```bash
 Day 1-2: FiberRenderer 更新
   - [ ] 修改 Render() 方法
   - [ ] 添加 renderComputedBox() 方法
   - [ ] 基于 RenderPlanes 渲染

 Day 3-4: 集成和测试
   - [ ] 更新所有 Render 调用点
   - [ ] 添加渲染测试
   - [ ] 验证渲染顺序

 Day 5-7: 测试和修复
   - [ ] 运行所有测试
   - [ ] 修复渲染问题
   - [ ] Code Review
```

#### Phase 6 第1周：HitMap 更新

```bash
 Day 1-2: 移除旧 API
   - [ ] 移除 GetMergedHitMap()
   - [ ] 更新所有 HitMap 调用点
   - [ ] 使用 BuildHitMapFromFiber()

 Day 3-4: HitTest 验证
   - [ ] 验证 HitTest 正确性
   - [ ] 验证 Z-order 排序
   - [ ] 验证 Modal 优先命中

 Day 5-7: 测试和修复
   - [ ] 运行所有测试
   - [ ] 修复事件问题
   - [ ] Code Review
```

#### Phase 7 第1周：清理和优化

```bash
 Day 1-2: 清理废弃代码
   - [ ] 移除所有 Deprecated 代码
   - [ ] 清理 unused imports
   - [ ] 移除无用的注释

 Day 3-4: 优化
   - [ ] 优化 RenderPlanes 性能
   - [ ] 优化 HitMap 性能
   - [ ] 添加性能基准测试

 Day 5-7: 文档和测试
   - [ ] 更新文档
   - [ ] 添加更多单元测试
   - [ ] Code Review
```

#### Phase 8 第1周：综合测试

```bash
 Day 1-2: 现有测试
   - [ ] 运行所有现有测试
   - [ ] 修复测试失败
   - [ ] 确保无回归

 Day 3-4: 新测试
   - [ ] 添加集成测试
   - [ ] 添加端到端测试
   - [ ] 添加用户场景测试

 Day 5-7: 性能和验证
   - [ ] 性能测试
   - [ ] 内存泄漏测试
   - [ ] 最终验证
```

---

## 测试和验证

### 9.1 单元测试

#### Fiber 测试

```go
// runtime/ui/fiber_test.go

func TestFiberLayerCopyFromVNode(t *testing.T) {
    vnode := rtui.NewElement("div")
    vnode.SetLayer(rtui.LayerModal)

    fiber := CreateFiber(vnode)

    assert.Equal(t, rtui.LayerModal, fiber.Layer)
}

func TestFiberComputedBoxNotNilAfterLayout(t *testing.T) {
    // Setup
    vnode := rtui.NewElement("div")
    fiber := CreateFiber(vnode)
    engine := compute.NewEngine()

    // Layout
    _, err := engine.Layout(vnode, fiber, runtime.BoxConstraints{
        MinWidth: 0,
        MaxWidth: 800,
        MinHeight: 0,
        MaxHeight: 600,
    })

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, fiber.ComputedBox)
    assert.Equal(t, fiber.NodeID, fiber.ComputedBox.NodeID)
    assert.Equal(t, fiber.Layer, fiber.ComputedBox.Layer)
}
```

#### Reconciler 测试

```go
// internal/reconciler/reconciler_test.go

func TestReconcilerLayerPreserved(t *testing.T) {
    // Setup
    reconciler := NewReconciler(nil)
    oldVNode := rtui.NewElement("div")
    oldVNode.SetLayer(rtui.LayerModal)
    oldFiber := CreateFiber(oldVNode)

    newVNode := rtui.NewElement("div")
    newVNode.SetLayer(rtui.LayerModal)

    // Reconcile
    newFiber := reconciler.reconcile(oldFiber, newVNode)

    // Assert
    assert.Equal(t, oldFiber.NodeID, newFiber.NodeID)
    assert.Equal(t, rtui.LayerModal, newFiber.Layer)
}
```

#### RenderPlanes 测试

```go
// runtime/layer/renderplanes_test.go

func TestRenderPlanesBuildFromFiber(t *testing.T) {
    // Setup
    fiberRoot := &Fiber{
        NodeID: 1,
        Layer:  rtui.LayerBase,
    }
    fiberRoot.Child = &Fiber{
        NodeID: 2,
        Layer:  rtui.LayerModal,
        Parent: fiberRoot,
        ComputedBox: &compute.ComputedBox{
            NodeID: 2,
            Box:    runtime.Box{X: 10, Y: 10, Width: 100, Height: 100},
            Layer:  rtui.LayerModal,
        },
    }

    // Build
    rp := NewRenderPlanes()
    rp.BuildFromFiber(fiberRoot)

    // Assert
    assert.NotNil(t, rp.Planes[rtui.LayerBase])
    assert.NotNil(t, rp.Planes[rtui.LayerModal])

    baseBoxes := rp.GetPlane(rtui.LayerBase)
    modalBoxes := rp.GetPlane(rtui.LayerModal)

    assert.Equal(t, 1, len(modalBoxes))
    assert.Equal(t, uint64(2), modalBoxes[0].NodeID)
}
```

#### HitMap 测试

```go
// runtime/event/hitmap_fiber_test.go

func TestBuildHitMapFromFiber(t *testing.T) {
    // Setup
    fiberRoot := &Fiber{
        NodeID: 1,
        Layer:  rtui.LayerBase,
        ComputedBox: &compute.ComputedBox{
            NodeID: 1,
            Box:    runtime.Box{X: 0, Y: 0, Width: 100, Height: 100},
            Layer:  rtui.LayerBase,
        },
    }

    modalFiber := &Fiber{
        NodeID: 2,
        Layer:  rtui.LayerModal,
        Parent: fiberRoot,
        ComputedBox: &compute.ComputedBox{
            NodeID: 2,
            Box:    runtime.Box{X: 50, Y: 50, Width: 100, Height: 100},
            Layer:  rtui.LayerModal,
        },
    }

    // Build
    hitMap := event.BuildHitMapFromFiber(fiberRoot)

    // HitTest at (60, 60) - Should hit modal first
    entry := hitMap.HitTest(60, 60)

    // Assert
    assert.NotNil(t, entry)
    assert.Equal(t, uint64(2), entry.NodeID)  // Modal
    assert.Equal(t, rtui.LayerModal, entry.Layer)
}
```

### 9.2 集成测试

```go
// integration/layer_render_test.go

func TestLayerRenderIntegration(t *testing.T) {
    // Setup App with Modal
    app := setupAppWithModal()

    // Render
    ctx := component.PaintContext{
        Width:  800,
        Height: 600,
    }
    buffer := paint.NewBuffer(800, 600)
    app.Render(ctx, buffer)

    // Verify RenderPlanes
    layerManager := app.GetLayerManager()
    renderPlanes := layerManager.GetRenderPlanes()

    assert.True(t, renderPlanes.HasLayer(rtui.LayerBase))
    assert.True(t, renderPlanes.HasLayer(rtui.LayerModal))

    // Verify HitMap
    hitMap := app.GetHitMap()
    assert.NotNil(t, hitMap)
    assert.NotZero(t, hitMap.Size())

    // Verify Modal hit first
    modalEntry := hitMap.HitTest(400, 300)  // Center (modal is centered)
    assert.NotNil(t, modalEntry)
    assert.Equal(t, rtui.LayerModal, modalEntry.Layer)
}
```

### 9.3 性能测试

```go
// performance/renderplanes_bench_test.go

func BenchmarkRenderPlanesBuildFromFiber(b *testing.B) {
    // Setup large Fiber tree
    fiberRoot := createLargeFiberTree(1000)  // 1000 nodes

    rp := NewRenderPlanes()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        rp.BuildFromFiber(fiberRoot)
    }
}

func BenchmarkBuildHitMapFromFiber(b *testing.B) {
    // Setup large Fiber tree
    fiberRoot := createLargeFiberTree(1000)  // 1000 nodes

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        event.BuildHitMapFromFiber(fiberRoot)
    }
}
```

### 9.4 用户场景测试

#### 场景 1：Modal 点击

```go
func TestUserScenario_ModalClick(t *testing.T) {
    // Setup app with modal
    app := setupApp("examples/modal")

    // Simulate click on modal button
    app.InjectMouseClick(450, 350)  // Click on modal's OK button

    // Verify button triggered
    assert.True(t, app.GetButtonClicked())
    assert.True(t, app.ModalClosed())
}

func TestUserScenario_ModalBackgroundClick(t *testing.T) {
    // Setup app with modal
    app := setupApp("examples/modal")

    // Simulate click outside modal
    app.InjectMouseClick(50, 50)  // Click on background

    // Verify modal not closed (modal blocks background clicks)
    assert.True(t, app.GetModalOpen())
}
```

#### 场景 2：Multi-layer HitTest

```go
func TestUserScenario_MultiLayerHitTest(t *testing.T) {
    // Setup app with multiple layers
    // - Base: Button
    // - Overlay: Tooltip
    // - Modal: Modal

    app := setupApp("examples/multilayer")

    // Click on tooltip (highest layer)
    tooltipEntry := app.HitTest(200, 100)
    assert.Equal(t, rtui.LayerTooltip, tooltipEntry.Layer)

    // Click on modal (second highest)
    modalEntry := app.HitTest(400, 300)
    assert.Equal(t, rtui.LayerModal, modalEntry.Layer)

    // Click on button (base layer, no overlay/modal at position)
    buttonEntry := app.HitTest(50, 50)
    assert.Equal(t, rtui.LayerBase, buttonEntry.Layer)
}
```

---

## 风险评估和缓解

### 10.1 风险识别

#### 风险 1：Layout 行为改变

**描述**：Layout 从多路径改为单路径，可能导致布局结果不同。

**影响**：中等
- Modal 位置可能稍有变化
- Overlay 位置可能调整
- 用户界面布局改变

**缓解措施**：
- ✅ Phase 2 时进行详细的 Layout 对比测试
- ✅ 保持 Modal 居中逻辑（transform 方式改为 explicit）
- ✅ 提供迁移指南
- ✅ 预留时间进行 UI 调整

#### 风险 2：性能退化

**描述**：Single pass Layout 可能有性能问题。

**影响**：低
- O(n) 时间复杂度（StripLayers 也是 O(n)）
- 单棵树遍历比多棵树快
- RenderPlanes 构建是 O(n)

**缓解措施**：
- ✅ Phase 7 时进行性能基准测试
- ✅ 优化 RenderPlanes 构建逻辑
- ✅ 添加性能监控

#### 风险 3：事件处理错误

**描述**：HitMap 基于 Fiber，可能导致事件分发出错。

**影响**：高
- Modal 点击不响应
- Tooltip 不显示
- 错误的组件响应事件

**缓解措施**：
- ✅ Phase 6 时进行详细的事件测试
- ✅ 验证 HitTest Z-order 排序
- ✅ 验证事件冒泡逻辑
- ✅ 添加详细的事件调试日志

#### 风险 4：破坏性变更太多

**描述**：大量 API 变更，可能导致迁移困难。

**影响**：中等
- 组件开发者需要更新代码
- 第三方库可能需要适配
- 学习曲线陡峭

**缓解措施**：
- ✅ 提供详细的迁移指南
- ✅ 提供兼容层（可选）
- ✅ 逐步废弃（Deprecated 标记）
- ✅ 完善的文档和示例

#### 风险 5：测试覆盖不足

**描述**：重构涉及大量代码，可能遗漏边缘情况。

**影响**：中等
- 边缘情况出现 bug
- 用户报告未预期行为

**缓解措施**：
- ✅ Phase 7 时添加大量单元测试
- ✅ Phase 8 时添加集成测试
- ✅ Phase 8 时添加用户场景测试
- ✅ 保持现有测试不失败

### 10.2 回滚计划

如果重构遇到重大问题，需要回滚：

#### 回滚触发条件

- [ ] Phase 2 后：Layout 错误 > 5%
- [ ] Phase 5 后：渲染错误 > 10%
- [ ] Phase 6 后：事件错误 > 5%
- [ ] Phase 8 后：性能退化 > 20%

#### 回滚步骤

1. 创建回滚分支
```bash
git checkout -b rollback-fiber-refactor
git revert <last-commit-before-refactor>
```

2. 恢复旧代码
```bash
# 恢复 StripLayers 相关代码
git checkout <commit-before-phase-4> -- runtime/layer/
```

3. 验证旧功能正常
```bash
go test ./...
go run examples/modal/main.go
```

4. 通知团队
- 发送邮件通知
- 更新 issue 状态
- 提供临时解决方案

---

## 迁移指南

### 11.1 组件开发者迁移

#### 迁移步骤 1：更新 API 调用

```go
// ❌ 旧代码
func MyComponent() rtui.VNode {
    return rtui.NewElement("div")
}

func (a *App) renderOverlay() {
    // 使用 StripLayers
    baseTree := layerManager.CollectAndLayout(vnode, fiber, constraints)
    modalNodes := layerManager.GetModalNodes()
}

// ✅ 新代码
func MyComponent() rtui.VNode {
    return rtui.NewElement("div")
}

func (a *App) renderOverlay() {
    // 使用 RenderPlanes
    fiberRoot := a.reconciler.GetFiberRoot()
    renderPlanes := layerManager.BuildRenderPlanes(fiberRoot)
    modalBoxes := renderPlanes.GetPlane(rtui.LayerModal)
}
```

#### 迁移步骤 2：移除 VNode.Bounds

```go
// ❌ 旧代码
func (b *Button) GetBounds() [4]int {
    return []int{b.x, b.y, b.width, b.height}
}

// ✨ 新代码
// Bounds 已经由 Layout 计算，不要在组件中设置
// 如果需要访问 Bounds：
func (b *Button) GetBoundsFromRuntime() [4]int {
    fiber := runtime.FindFiberByVNode(b.VNode)
    if fiber != nil && fiber.ComputedBox != nil {
        return []int{
            fiber.ComputedBox.Box.X,
            fiber.ComputedBox.Box.Y,
            fiber.ComputedBox.Box.Width,
            fiber.ComputedBox.Box.Height,
        }
    }
    return []int{0, 0, 0, 0}
}
```

#### 迁移步骤 3：更新 HitTest 使用

```go
// ❌ 旧代码
func (a *App) handleMouseClick(x, y int) {
    hitMap := a.layerManager.GetMergedHitMap()
    entry := hitMap.HitTest(x, y)
    if entry != nil {
        // 处理点击
    }
}

// ✨ 新代码
func (a *App) handleMouseClick(x, y int) {
    // HitMap 已经在 render() 时构建
    hitMap := a.GetHitMap()
    entry := hitMap.HitTest(x, y)
    if entry != nil {
        // 处理点击
    }
}
```

### 11.2 第三方库迁移

如果你的库依赖 Mint 的内部 API：

#### 依赖 StripLayers

```go
// ❌ 旧代码
func MyLayoutEngine(vnode rtui.VNode) rtui.VNode {
    collector := layer.NewCollector()
    collector.Collect(vnode)
    return collector.StripLayers(vnode)
}

// ✨ 新代码
func MyLayoutEngine(fiberRoot *Fiber) *RenderPlanes {
    renderPlanes := NewRenderPlanes()
    renderPlanes.BuildFromFiber(fiberRoot)
    return renderPlanes
}
```

#### 依赖独立的 HitMap

```go
// ❌ 旧代码
func MyInspector(vnode rtui.VNode) *event.HitMap {
    engine := compute.NewEngine()
    layout := engine.Layout(vnode, constraints)
    return layout.HitMap
}

// ✨ 新代码
func MyInspector(fiberRoot *Fiber) *event.HitMap {
    return event.BuildHitMapFromFiber(fiberRoot)
}
```

### 11.3 常见问题

#### Q1: Modal 位置不对

**原因**：Modal 布局约束改变

**解决方案**：
```go
// 检查 modal 的 constraints
modalConstraints := runtime.BoxConstraints{
    MinWidth:  0,
    MaxWidth:  maxWidth,
    MinHeight: 0,
    MaxHeight: maxHeight,
}

// 确保使用正确的 constraints
layout, err := engine.Layout(modalVNode, modalFiber, modalConstraints)
```

#### Q2: Event 不响应

**原因**：HitMap 未正确构建或 Layer 排序错误

**解决方案**：
```go
// 确保 HitMap 基于 Fiber
hitMap := event.BuildHitMapFromFiber(fiberRoot)

// 启用 Debug 日志
log.HitMapLogger.Enabled = true

// 检查 HitMap entries
for _, entry := range hitMap.AllEntries() {
    log.HitMapLogger.Debug("[HitMap] ID=%d, Layer=%d, Bounds=%v",
        entry.NodeID, entry.Layer, entry.Bounds)
}
```

#### Q3: Tooltip 不显示

**原因**：Tooltip Layer Z-order 较低，被遮挡

**解决方案**：
```go
// 确保 Tooltip 使用正确的 Layer
tooltip.SetLayer(rtui.LayerTooltip)

// 检查 Tooltip box.Size > 0
if fiber.ComputedBox.Box.Width > 0 && fiber.ComputedBox.Box.Height > 0 {
    // Tooltip 可见
}
```

---

## 附录

### A. 参考文档

1. `docs/render/fiber/diff_key.md` - DiffKey vs NodeID 设计
2. `docs/render/fiber/diff_layer.md` - Layer 作为渲染维度设计
3. `AGENTS.md` - Agent 指南
4. `docs/layer_system_analysis.md` - Layer 系统分析

### B. 术语表

| 术语 | 定义 |
|------|------|
| VNode | 虚拟节点，声明式 UI 描述 |
| Fiber | 工作单元，运行时实例，保存 NodeID |
| NodeID | 运行时唯一身份标识（uint64） |
| DiffKey | 用于 sibling diff 的 key（string） |
| Layer | 渲染层（Base, Overlay, Modal, Tooltip, Inspector） |
| RenderPlane | 渲染投影层，按 Layer 分桶 |
| ComputedBox | 布局计算结果 |
| HitMap | 事件命中映射表 |
| StripLayers | ❌ 废弃的多层渲染机制 |
| RenderPlanes | ✨ 新的多层渲染机制 |

### C. 相关 Issue

- 链接到相关的 GitHub Issues 或设计文档
- 记录重构相关的讨论和决策

---

## 总结

本次重构将 Mint TUI Framework 从混合架构转变为统一的 Fiber 架构：

### 核心变更

1. ✅ **Fiber 作为唯一运行时结构**
   - 所有状态保存在 Fiber
   - VNode 纯声明，不保存运行时信息

2. ✅ **Layer 作为渲染维度**
   - Layer 不参与结构变换
   - 只在渲染阶段用于分桶

3. ✅ **废弃 StripLayers**
   - 不再克隆 VNode
   - 不再生成独立子树

4. ✅ **RenderPlane 投影**
   - 对 Fiber 树的只读分桶
   - 不修改树结构

5. ✅ **单一 HitMap**
   - 基于单一 Fiber 树构建
   - 自动按 Layer 排序

### 下一步

- 开始实施 Phase 1（基础设施）
- 定期更新进度报告
- 及时处理阻塞问题
- 保持代码质量和测试覆盖率

---

**文档版本**: 1.0
**最后更新**: 2026-02-14
**审核状态**: 待审核
