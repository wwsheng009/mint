# Fiber-first 渲染路径完整分析

## 概述

`fiberFirstPaint` 是 Mint TUI 框架的核心渲染路径，采用 **Fiber-first 架构**，实现了完整的三阶段渲染流水线：

```
Phase 1: VNode → Fiber (VNode 丢弃)
Phase 2: Fiber → LayoutBox (布局计算)
Phase 3: LayoutBox → PaintableBox → Buffer (绘制)
```

**位置**: `internal/render/declarative_node.go:412`

---

## Phase 1: Fiber Reconciliation

### 代码位置
`declarative_node.go:414-433`

### 流程

```go
// Phase 1: Fiber Reconciliation
n.reconciler.Render(component.PaintContext{
    Bounds: paint.Rect{X: 0, Y: 0, Width: 1, Height: 1},
}, nullBuf, n.renderFn)
```

### 关键操作

1. **创建/复用 Fiber 节点**
   - `beginWork()` 调用 `InstanceFactory` 创建 ComponentInstance
   - 根据 DiffKey 判断是否可以复用现有 Fiber

2. **完成 Fiber 树构建**
   - `completeWork()` 设置 Fiber 字段：
     - `Fiber.NodeID` - 稳定的运行时 ID (uint64)
     - `Fiber.Instance` - Component 实例，跨渲染持久化
     - `Fiber.Layer` - 渲染层级 (Base=0, Overlay=1, Modal=2, Tooltip=3, Inspector=4)

3. **VNode 丢弃**
   - Phase 1 结束后，VNode 立即释放
   - 后续阶段仅使用 Fiber 数据

### Phase 1.5: FocusManager 同步

```go
if n.fwApp != nil && n.focusMgr != nil {
    appSetter.SetFocusManagerFromDeclarativeNode(n.focusMgr)
}
```

- 将 FiberFocusManager 同步到 framework.App
- 用于 Tab 导航和事件路由

---

## Phase 2: Fiber-based Layout

### 代码位置
`declarative_node.go:444-473`

### 流程

```go
layoutResult, err := n.newLayoutEngine.LayoutFiber(fiberRoot, constraints)
```

### 详细步骤

#### Step 1: 创建 Fiber 适配器

```go
// layout_switcher.go:77
node := NewFiberToNodeAdapterPure(fiber)
```

- `FiberToNodeAdapterPure` 实现 `layout.Node` 接口
- 关键方法：
  - `Measure(constraints) → Size`
  - `Children() → []Node`
  - `FlexStyle() → FlexStyle`
  - `GridStyle() → GridStyle`
  - `WrapStyle() → WrapStyle`
  - `AbsoluteStyle() → AbsoluteStyle`

#### Step 2: 布局引擎计算

```go
// layout_switcher.go:84
result := a.engine.Layout(node, layoutConstraints)
```

**实际调用**: `runtime/layout/types.go:414-440`

```go
func (e *Engine) Layout(root Node, constraints Constraints) *LayoutResult {
    // 1. 检查缓存
    if cached := e.cache.Get(root, constraints); cached != nil {
        return cached
    }

    // 2. 递归布局
    box := e.layoutNode(root, constraints, 0, 0)
    result.Root = box
    result.Boxes = e.collectBoxes(box)

    // 3. 构建命中映射表
    result.HitMap.BuildFromLayoutBox(box)

    // 4. 存入缓存
    e.cache.Put(root, constraints, result)
    return result
}
```

#### Step 3: 坐标计算方式（父子累积）

```go
// types.go:499 + 620-637
func (e *Engine) layoutNodeWithDepth(node Node, constraints Constraints, x, y int, ...) *LayoutBox {
    // 父节点传入自己的 x, y
    // 子节点相对于父节点定位

    ...
    borderOffsetX, borderOffsetY := nodeBorder.ContentOffset()

    // 示例: 默认垂直布局
    childX := x + borderOffsetX
    childY := y + borderOffsetY

    for _, child := range node.Children() {
        childBox := e.layoutNodeWithDepth(child, childConstraints, childX, childY, ...)
        box.Children = append(box.Children, childBox)
        childY += childBox.Height
    }
}
```

**关键点**:
- 坐标是相对于父节点的
- 通过父节点的 `x + offsetX` 传递给子节点
- **Layer 属性不影响布局坐标计算**（单树共享架构）

#### Step 4: Layer 属性提取

```go
// types.go:576-580
layer := GetLayerFromNode(node)
zIndex := GetZIndexFromNode(node)

box := &LayoutBox{
    ID:       node.ID(),
    X:        x,
    Y:        y,
    Width:    width,
    Height:   height,
    Layer:    layer,  // ✨ Propagated
    ZIndex:   zIndex, // ✨ Propagated
    ...
}
```

#### 回调: 布局容器类型支持

| 容器类型 | 接口 | 实现文件 |
|---------|------|---------|
| Flex     | `FlexStyleProvider` | `runtime/layout/flex.go` |
| Grid     | `GridStyleProvider` | `runtime/layout/grid.go` |
| Wrap     | `WrapStyleProvider` | `runtime/layout/wrap.go` |
| Absolute | `AbsoluteStyleProvider` | `runtime/layout/absolute.go` |

#### 注释: 方案A - 单树共享布局

```go
// declarative_node.go:444
// 方案A: 单树共享布局 - 所有layer在一个布局树中计算
// 移除了LayerManager的坐标归一化，保留原始坐标用于正确渲染
```

**设计原则**:
- ✅ 所有层的节点在同一个 LayoutBox 树中
- ✅ Layer 属性仅用于控制渲染顺序 (Z-order) 和 HitTest 优先级
- ✅ 不需要为每个 layer 创建独立的布局树
- ✅ 坐标保持父子累积的原始值

### Phase 2 输出

```go
// *layout.LayoutResult {
type LayoutResult struct {
    Root     *LayoutBox        // 布局树根节点
    Boxes    []LayoutBox       // 所有盒子扁平化列表
    HitMap   *layout.HitMap    // 事件路由映射表
    Dirty    bool              // 脏标记
    // ⚠️ LayerManager 字段已存在但未使用
    LayerManager *LayerManager // 当前未在 fiberFirstPaint 路径中使用
}
```

---

## Phase 3: Paint 绘制

### 代码位置
`declarative_node.go:486-520`

### 流程

#### Step 1: LayoutBox → PaintableBox 转换

```go
converter := NewFiberToPaintableConverter(fiberRoot)
paintableLayout := converter.ConvertToLayout(innerResult.Root)
```

**转换器**: `internal/render/converter.go`

```go
type FiberToPaintableConverter struct {
    fiberMap map[string]*reconciler.Fiber  // 快速查找 Fiber
}

func (c *FiberToPaintableConverter) Convert(lbox *LayoutBox, parent *PaintableBox) *PaintableBox {
    pbox := &paint.PaintableBox{
        X:        lbox.X,      // 直接复制坐标
        Y:        lbox.Y,
        Width:    lbox.Width,
        Height:   lbox.Height,
        Layer:    convertLayoutLayerToInt(lbox.Layer),  // ✨ Layer 属性传播
        ZIndex:   lbox.ZIndex,                         // ✨ ZIndex 属性传播
        Parent:   parent,
        ...
    }

    // 查找对应 Fiber，填充绘制需要的数据
    if fiber := c.findFiber(lbox.ID); fiber != nil {
        c.fillFromFiber(pbox, fiber)
    }

    return pbox
}
```

#### Step 2: 构建 PaintablePlanes（按分层组织）

```go
// declarative_node.go:506-517
planes := paint.NewPaintablePlanes()

var walkPaintable func(box *paint.PaintableBox)
walkPaintable = func(box *paint.PaintableBox) {
    if box == nil {
        return
    }
    planes.AddToLayer(paint.RenderLayer(box.Layer), box)  // ✨ 按 layer 添加
    for _, child := range box.Children {
        walkPaintable(child)
    }
}
walkPaintable(paintableLayout.Root)
```

**PaintablePlanes 结构**:
```go
type PaintablePlanes struct {
    layers map[int][]*PaintableBox  // 按 layer ID 存储盒子
}

func (p *PaintablePlanes) AddToLayer(layer RenderLayer, box *PaintableBox) {
    p.layers[int(layer)] = append(p.layers[int(layer)], box)
}
```

#### Step 3: 绘制到 Buffer

```go
// declarative_node.go:520
if err := n.paintEngine.PaintPaintablePlanes(planes, buf); err != nil {
    n.legacyPaint(ctx, buf)
    return
}
```

**绘制引擎**:
- 按 Z-order 递减顺序绘制各层（Tooltip → Inspector → ... → Base）
- 每层内部按 ZIndex 递减顺序绘制
- 实现：`runtime/paint/paint_engine.go:PaintPaintablePlanes()`

```go
func (e *PaintEngine) PaintPaintablePlanes(planes *PaintablePlanes, buf *Buffer) error {
    // 从高 layer 到低 layer 绘制
    for layer := LayerMax; layer >= LayerBase; layer-- {
        boxes := planes.GetLayer(layer)
        // 应用 layer transforms (modal centering 等)
        for _, box := range boxes {
            // 绘制
        }
    }
}
```

### Phase 3 输出

```go
// 保存到 DeclarativeNode
n.mu.Lock()
n.lastPaintableRoot = paintableLayout.Root  // 用于 GetPaintableBoxes()
n.fiberLastHitMap = hitMap                  // 用于事件路由
n.mu.Unlock()
```

---

## 数据流图

```
┌─────────────────────────────────────────────────────────────────┐
│  用户代码: UI.render()(...props...)                             │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  Phase 1: Fiber Reconciliation (declarative_node.go:414)       │
├─────────────────────────────────────────────────────────────────┤
│  VNode (瞬态)                                                    │
│    ├── reconciler.Render(nullBuf, renderFn)                     │
│    │   ├── beginWork() → InstanceFactory.CreateInstance()      │
│    │   └── completeWork()                                      │
│    │       ├── fiber.NodeID = counter++                        │
│    │       ├── fiber.Instance = instance                       │
│    │       └── fiber.Layer = props["layer"]                    │
│    │                                                           │
│    └── ✅ VNode discarded                                       │
│                                                                │
│  Fiber (持久化)                                                 │
│    ├── NodeID: uint64                                          │
│    ├── Instance: ComponentInstance                             │
│    ├── Layer: Layer (Base/Overlay/Modal/Tooltip/Inspector)    │
│    ├── LayoutDirection, LayoutAlign, LayoutGap ...            │
│    └── BorderStyle, BorderLabel                               │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  Phase 1.5: FocusManager 同步 (declarative_node.go:436)        │
├─────────────────────────────────────────────────────────────────┤
│  fwApp.SetFocusManagerFromDeclarativeNode(n.focusMgr)          │
│  → 用于 Tab 导航和事件路由                                      │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  Phase 2: Fiber-based Layout (declarative_node.go:444)         │
├─────────────────────────────────────────────────────────────────┤
│  n.newLayoutEngine.LayoutFiber(fiberRoot, constraints)         │
│    │                                                           │
│    ├── FiberToNodeAdapterPure实现 layout.Node 接口             │
│    │   ├── Measure(constraints) → Size                        │
│    │   ├── Children() → []Node                                │
│    │   ├── FlexStyle() → FlexStyle                            │
│    │   ├── GridStyle() → GridStyle                            │
│    │   ├── WrapStyle() → WrapStyle                            │
│    │   └── AbsoluteStyle() → AbsoluteStyle                    │
│    │                                                           │
│    └── layout.Engine.Layout(node, constraints)                 │
│        │                                                       │
│        ├── layoutNodeWithDepth() 递归布局                      │
│        │   ├── GetLayerFromNode(node) → LayoutBox.Layer        │
│        │   ├── GetZIndexFromNode(node) → LayoutBox.ZIndex      │
│        │   └── 坐标: childX = x + offsetX, childY = y + ...    │
│        │                                                       │
│        └── ✅ 单树共享布局 - 所有layer一个树                    │
│           ↕                                                      │
│  LayoutBox (布局结果)                                            │
│    ├── Root *LayoutBox (布局树)                                │
│    ├── Boxes []LayoutBox (扁平化列表)                          │
│    ├── HitMap *layout.HitMap (事件路由)                        │
│    └── LayerManager *LayerManager (未使用)                      │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  Phase 3: Paint (declarative_node.go:486)                      │
├─────────────────────────────────────────────────────────────────┤
│  FiberToPaintableConverter.ConvertToLayout(layoutResult.Root) │
│    │                                                           │
│    ├── Convert(lbox, parent) → PaintableBox                    │
│    │   ├── X, Y, Width, Height 直接复制                       │
│    │   ├── Layer, ZIndex 直接复制                              │
│    │   └── 查找 Fiber 填充绘制数据                            │
│    │                                                           │
│    └── PaintableLayout                                         │
│        └── Root *PaintableBox (绘制树)                         │
│                                                                │
│  构建分层: planes.AddToLayer(paint.RenderLayer(box.Layer), ...)│
│    ├── Base layer (0)                                          │
│    ├── Overlay layer (1)                                       │
│    ├── Modal layer (2)                                         │
│    ├── Tooltip layer (3)                                       │
│    └── Inspector layer (4)                                     │
│                                                                │
│  paintEngine.PaintPaintablePlanes(planes, buf)                 │
│    │                                                           │
│    ├── 从高 Z 到低 Z 绘制各层                                  │
│    ├── 每层内部按 ZIndex 排序                                    │
│    └── 应用 layer transforms (modal centering)                 │
│        ↕                                                      │
│  Buffer (终端缓冲区)                                            │
└─────────────────────────────────────────────────────────────────┘
```

---

## 关键设计决策

### 1. 单树共享布局 vs 多树独立布局

| 方案 | 描述 | 状态 |
|-----|------|-----|
| 方案A | 单树共享布局 - 所有layer在一个布局树中计算 | ✅ 已实现 |
| 方案B | 多树独立布局 - 每个独立的布局树 | ❌ 未实现 |

**当前实现**: 方案A

**优势**:
- 简化架构，减少数据结构
- Layer 作为属性而非结构维度
- 坐标计算统一，易于维护

**参考**: `docs/layer/layer_design.md`, `LAYER_SYSTEM_ARCHITECTURE.md`

### 2. Layer 属性的作用

| 作用 | 说明 |
|-----|------|
| 渲染顺序 | 控制各层绘制的前后顺序 |
| HitTest 优先级 | 事件路由时优先命中高 Z 节点 |
| ❌ 不用于坐标计算 | Layer 不影响 LayoutBox 的 X/Y 坐标 |

### 3. 坐标系统

- **相对坐标**: LayoutBox 的 X/Y 是相对于父节点的
- **父子累积**: 子节点坐标 = `parentX + offsetX`
- **全局坐标**: 当前缺失 `AbsX/AbsY` 字段（计划添加）

---

## 实现状态 vs 设计意图对比

### ✅ 已实现（符合设计意图）

1. **Fiber-first 架构**
   - VNode 在 Phase 1 后丢弃 ✅
   - Fiber 属性完整传播到 LayoutBox 和 PaintableBox ✅

2. **Layer 链式传播**
   - Fiber.Layer → LayoutBox.Layer → PaintableBox.Layer ✅

3. **单树共享布局**
   - 所有层在同一棵布局树中计算 ✅
   - Layer 属性用于 Z-order 排序 ✅

4. **分层绘制**
   - PaintablePlanes 按层组织 PaintableBox ✅
   - 绘制引擎按 Z-order 绘制 ✅

5. **事件路由**
   - HitMap 从 LayoutBox 构建 ✅
   - 支持 TargetFiber 丰富属性 ✅
   - 从高 Z 到低 Z 路由事件 ✅

### ⚠️ 部分实现（需要优化）

1. **Modal 居中**
   - 当前位置: `internal/render/rendering_pipeline.go:applyLayerTransformsToPaintable()` (Paint 阶段)
   - 设计位置: 应该在 Layout 阶段
   - 建议修复: 在 `layoutNodeWithDepth()` 或 `LayerManager.ApplyLayerTransforms()` 中实现

2. **LayerManager 状态**
   - LayerManager 已定义，但在 `fiberFirstPaint` 路径中未被调用
   - 注释说明已移除坐标归一化逻辑

### ❌ 未实现（计划中的功能）

1. **AbsX/AbsY 全局坐标**
   - LayoutBox 缺少 AbsX/AbsY 字段
   - 影响: 需要向上遍历父链计算全局坐标
   - 计划: 在布局阶段预计算全局坐标

2. **Position:Fixed 支持**
   - Position:Fixed 类型已定义 (`runtime/layout/position.go`)
   - Fiber 结构缺少 `Position` 和 `Anchor` 字段
   - 布局引擎未处理 Position:Fixed 的特殊逻辑（相对于视口定位）

3. **Portal 系统**
   - 设计文档中提到的 Fiber.PortalRoot 未实现
   - 需要在 completeWork 中设置 PortalRoot
   - 需要 OverlayManager 管理跨树挂载的节点

---

## 接口和类型关系

### Fiber → LayoutBox 转换

**适配器**: `FiberToNodeAdapterPure` (`internal/render/fiber_adapter.go`)

```go
// fiber_adapter.go
type FiberToNodeAdapterPure struct {
    fiber *reconciler.Fiber
}

// 实现 layout.Node 接口
func (a *FiberToNodeAdapterPure) Measure(constraints Constraints) Size { ... }
func (a *FiberToNodeAdapterPure) Children() []Node { ... }
func (a *FiberToNodeAdapterPure) FlexStyle() *style.FlexStyle { ... }
func (a *FiberToNodeAdapterPure) Layer() Layer { return a.fiber.Layer }
func (a *FiberToNodeAdapterPure) ZIndex() int { ... }
```

### LayoutBox → PaintableBox 转换

**转换器**: `FiberToPaintableConverter` (`internal/render/converter.go`)

```go
type FiberToPaintableConverter struct {
    fiberMap map[string]*reconciler.Fiber
}

func (c *FiberToPaintableConverter) Convert(lbox *LayoutBox, parent *PaintableBox) *PaintableBox {
    pbox := &paint.PaintableBox{
        X:        lbox.X,
        Y:        lbox.Y,
        Width:    lbox.Width,
        Height:   lbox.Height,
        Layer:    int(lbox.Layer),
        ZIndex:   lbox.ZIndex,
        ...
    }

    // 查找对应的 Fiber，填充 Node 引用
    if fiber := c.findFiber(lbox.ID); fiber != nil {
        pbox.Node = NewFiberPaintableNode(fiber)
    }

    return pbox
}
```

---

## 性能优化

1. **布局缓存**
   - `layout.Engine` 使用 `cache.Get/Put` 缓存布局结果
   - 基于 Node 和 constraints 的组合键

2. **HitMap 构建**
   - 在布局阶段构建，避免在绘制阶段重复计算
   - 用于高效的事件路由

---

## 相关文档

- `docs/layer/layer_design.md` - Layer 系统设计原则
- `docs/layer/LAYER_SYSTEM_ARCHITECTURE.md` - Fiber-first Layer 架构文档
- `docs/layer/POSITIONING.md` - Modal 定位实现指南
- `runtime/layout/README.md` - 布局引擎架构
- `internal/reconciler/reconcile_engine.md` - Reconciler 设计文档

---

## 示例: 完整渲染流程

```go
// 用户代码
func MyRender() UI {
    return VBox(
        Text("Hello"),

        // Modal 在 Overlay layer
        Modal(
            Text("Alert"),
            Layer(LayerOverlay),
        ),
    )
}

// Phase 1: Fiber Reconciliation
// reconciler.Render(nullBuf, renderFn)
// → Fiber tree created:
//    - Root.NodeID = 1, Layer = 0 (Base)
//    - VBox.NodeID = 2, Layer = 0
//    - Text("Hello").NodeID = 3, Layer = 0
//    - Modal.NodeID = 4, Layer = 1 (Overlay)
//    - Text("Alert").NodeID = 5, Layer = 1

// Phase 2: Fiber-based Layout
// LayoutFiber(fiberRoot, constraints)
// → LayoutBox tree:
//    - Root[0,0,80,24, Layer=0]
//      ├── VBox[0,0,80,20, Layer=0]
//      │   └── Text[0,0,5,1, Layer=0]
//      └── Modal[0,0,20,5, Layer=1]  // ⚠️ 应该居中，但当前未实现
//          └── Text[0,0,5,1, Layer=1]

// Phase 3: Paint
// PaintablePlanes:
//    - Layer 1 (Overlay): [Modal(0,0,20,5), Text(0,0,5,1)]
//    - Layer 0 (Base): [Root(0,0,80,24), VBox(0,0,80,20), Text(0,0,5,1)]
// PaintPaintablePlanes() 会按 Z-order 绘制
```

---

## 总结

`fiberFirstPaint` 是一个设计良好、实现完整的 Fiber-first 渲染流水线：

✅ **核心特性**:
- 三阶段清晰的职责分离
- 单树共享布局架构
- Layer 属性正确传播到绘制阶段
- 高效的缓存和事件路由机制

⚠️ **需要改进**:
- 将 Modal 居中逻辑从 Paint 阶段移到 Layout 阶段
- 添加 AbsX/AbsY 全局坐标支持

❌ **计划中功能**:
- Position:Fixed 完整支持
- Portal 跨树挂载系统

---

**文档版本**: v1.0
**更新日期**: 2026-03-01
**作者**: Qwen Code
