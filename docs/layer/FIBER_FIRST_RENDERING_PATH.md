# fiberFirstPaint() 渲染路径技术分析

**NewDeclarativeNodeFromFuncWithFiber 中的多层 Layer 管理完整解析**

---

## 📋 目录

1. [概述](#概述)
2. [初始化流程](#初始化流程)
3. [Phase 1: Fiber Reconciliation](#phase-1-fiber-reconciliation)
4. [Phase 2: Layout (Fiber → LayoutBox)](#phase-2-layout-fiber--layoutbox)
5. [Phase 3: Paint (LayoutBox → Buffer)](#phase-3-paint-layoutbox--buffer)
6. [Layer 传播机制](#layer-传播机制)
7. [关键代码路径](#关键代码路径)
8. [数据流转图](#数据流转图)

---

## 概述

### 渲染流程概览

```
NewDeclarativeNodeFromFuncWithFiber(renderFn, fwApp)
    ↓
initFiberFirstPipeline()
    ├─ newLayoutEngine = NewNewLayoutEngineAdapter()
    └─ paintEngine = NewPaintEngine()
    ↓
Paint(ctx, buf) [每帧执行]
    ↓
fiberFirstPaint(ctx, buf)
    ├─ Phase 1: Fiber Reconciliation (VNode → Fiber)
    ├─ Phase 2: Layout (Fiber → LayoutBox)
    └─ Phase 3: Paint (LayoutBox → Buffer)
```

### 三阶段渲染架构

| 阶段 | 输入 | 输出 | 核心组件 |
|------|------|------|---------|
| **Phase 1** | VNode 函数 (`renderFn`) | Fiber 树 | `reconciler` |
| **Phase 2** | Fiber 树 | LayoutBox 树 | `NewLayoutEngineAdapter` |
| **Phase 3** | LayoutBox 树 | Buffer | `PaintEngine` + `PaintablePlanes` |

---

## 初始化流程

### NewDeclarativeNodeFromFuncWithFiber 构造函数

**文件**: `internal/render/declarative_node.go`

```go
// NewDeclarativeNodeFromFuncWithFiber 创建一个 Fiber-first DeclarativeNode
func NewDeclarativeNodeFromFuncWithFiber(
    renderFn func() rtui.VNode,
    fwApp framework.ISystem,
) *DeclarativeNode {
    // 步骤 1: 创建基本结构
    n := &DeclarativeNode{
        renderFn:           renderFn,
        fwApp:              fwApp,
        reconciler:         reconciler.NewReconciler(),
        renderMode:         RenderModeFiberFirst,    // ← Fiber-first 模式
        fiberFirstEnabled:  true,                    // ← 自动启用
    }

    // 步骤 2: 初始化 Fiber-first pipeline 组件
    n.initFiberFirstPipeline()

    // 步骤 3: 启动帧调度器
    n.scheduler = NewFrameScheduler(n)

    return n
}
```

### initFiberFirstPipeline 组件初始化

```go
// initFiberFirstPipeline 初始化 Fiber-first 渲染所需的所有组件
func (n *DeclarativeNode) initFiberFirstPipeline() {
    // Layout Engine - 使用 runtime/layout
    if n.newLayoutEngine == nil {
        n.newLayoutEngine = NewNewLayoutEngineAdapter()
        log.RenderLogger.Debug("✅ NewLayoutEngineAdapter initialized")
    }

    // Paint Engine - 支持 PaintPaintablePlanes
    if n.paintEngine == nil {
        n.paintEngine = NewPaintEngine()
        log.RenderLogger.Debug("✅ PaintEngine initialized")
    }

    // Note: FiberToPaintableConverter 是每帧创建的 (在 fiberFirstPaint 中)
}
```

---

## Phase 1: Fiber Reconciliation

### 执行位置

**文件**: `internal/render/declarative_node.go:fiberFirstPaint()`

```go
// ========================================
// Phase 1: Fiber Reconciliation
// ========================================
// The reconciler updates the Fiber tree. VNode is discarded after this.
// Use a minimal buffer for reconciliation (actual painting happens later).

nullBuf := paint.NewBuffer(1, 1)  // ← 最小 buffer，仅用于 reconciliation
n.reconciler.Render(component.PaintContext{
    Bounds: paint.Rect{X: 0, Y: 0, Width: 1, Height: 1},
}, nullBuf, n.renderFn)

// 获取 Fiber 根节点
fiberRoot := n.getFiberRoot()
```

### Reconciler 工作流程

```go
// reconciler.Render() 内部流程:
1. 调用 renderFn() 生成 VNode 树
2. 对比新旧 VNode，更新 Fiber 树
   - beginWork(): 创建 Fiber 节点
   - completeWork(): 完成 Fiber 节点
3. VNode 被丢弃，Fiber 跨帧持久化
```

### Fiber 节点结构

**文件**: `runtime/ui/fiber.go`

```go
type Fiber struct {
    // === 基础信息 ===
    Type  string    // 节点类型 (element, text, component等)
    Tag   string    // HTML 标签名 (对于 element)
    Props rtui.Props // 属性

    // === Fiber 树结构 ===
    Child   *Fiber
    Sibling *Fiber
    Return  *Fiber

    // === 状态管理 ===
    NodeID          uint64          // 唯一标识符
    State           interface{}     // 组件状态
    Hooks           []Hook          // Hook 列表
    Ref             *Ref

    // === Layer 信息 (多层渲染的核心) ===
    Layer rtui.Layer  // ← 持久化的 Layer 状态

    // === 性能优化 ===
    EffectTag   EffectTag       // 副作用标记
    Alternate   *Fiber          // 双缓冲 Fiber
    Refs        map[string]*Ref // 引用映射
}
```

### Fiber 创建时的 Layer 初始化

**文件**: `runtime/ui/fiber_util.go`

```go
func NewFiber(
    vnodeType string,
    tag string,
    props rtui.Props,
    children []rtui.VNode,
) *Fiber {
    return &Fiber{
        Type:    vnodeType,
        Tag:     tag,
        Props:   props,
        NodeID:  generateNodeID(),
        Layer:   props.GetLayer(),  // ← 从 VNode.GetLayer() 获取初始 Layer

        // ... 其他字段初始化
        Child:   nil,
        Sibling: nil,
    }
}
```

### Fiber 树的 Layer 传播

```
VNode 树 (每帧创建)
├─ HStack (LayerBase)
│  ├─ Button (LayerBase)
│  └─ Text (LayerBase)
└─ Tooltip (LayerTooltip)  ← VNode.GetLayer() 返回 LayerTooltip
              ↓
              ↓ NewFiber()
              ↓
Fiber 树 (跨帧持久化)
├─ Fiber #1 (LayerBase)
│  ├─ Fiber #2 (LayerBase)
│  └─ Fiber #3 (LayerBase)
└─ Fiber #4 (LayerTooltip)  ← Fiber.Layer = VNode.GetLayer()
```

---

## Phase 2: Layout (Fiber → LayoutBox)

### 执行位置

**文件**: `internal/render/declarative_node.go:fiberFirstPaint()`

```go
// ========================================
// Phase 2: Fiber-based Layout
// ========================================
// Use the new layout engine directly (runtime/layout), bypassing LayoutSwitcher

constraints := runtime.BoxConstraints{
    MinWidth:  0,
    MaxWidth:  ctx.AvailableWidth,
    MinHeight: 0,
    MaxHeight: ctx.AvailableHeight,
}

// 确保新的 layout engine 已初始化
if n.newLayoutEngine == nil {
    n.newLayoutEngine = NewNewLayoutEngineAdapter()
}

// 执行布局
layoutResult, err := n.newLayoutEngine.LayoutFiber(fiberRoot, constraints)
if err != nil {
    log.RenderLogger.Debug("Layout FAILED: %v, falling back to legacy", err)
    n.legacyPaint(ctx, buf)
    return
}
```

### LayoutFiber 实现

**文件**: `internal/render/layout_switcher.go`

```go
// LayoutFiber 使用 Fiber-only adapter 执行布局
func (a *NewLayoutEngineAdapter) LayoutFiber(
    fiber *reconciler.Fiber,
    constraints runtime.BoxConstraints,
) (LayoutResult, error) {
    // 步骤 1: 创建 Fiber-only adapter
    node := NewFiberToNodeAdapterPure(fiber)

    // 步骤 2: 转换 constraints
    layoutConstraints := layout.Constraints{
        MinWidth:  constraints.MinWidth,
        MaxWidth:  constraints.MaxWidth,
        MinHeight: constraints.MinHeight,
        MaxHeight: constraints.MaxHeight,
    }

    // 步骤 3: 执行布局
    result := a.engine.Layout(node, layoutConstraints)

    // 步骤 4: 返回适配后的结果
    return &newLayoutResultAdapter{result: result, fiberRoot: fiber}, nil
}
```

### FiberToNodeAdapterPure Layer 映射

**文件**: `internal/render/fiber_adapter.go`

```go
// FiberToNodeAdapterPure 实现 layout.Node 接口
type FiberToNodeAdapter struct {
    fiber *reconciler.Fiber
}

// 实现 layout.Node 接口的 GetLayer()
// 将 Fiber.Layer 映射到 layout.Layer
func (a *FiberToNodeAdapter) GetLayer() layout.Layer {
    if a.fiber == nil {
        return layout.LayerBase
    }
    // ← 零拷贝映射: types.Layer 与 layout.Layer 是相同类型
    return layout.Layer(a.fiber.Layer)
}
```

### 类型统一: layout.Layer

**文件**: `runtime/layout/types.go`

```go
// Layer 来自 types.Layer (类型别名)
type Layer = types.Layer

// types.Layer 定义在 runtime/types/layer.go
type Layer int

const (
    LayerBase      Layer = iota  // 0
    LayerOverlay                 // 1
    LayerModal                   // 2
    LayerTooltip                 // 3
    LayerInspector               // 4
)
```

### LayoutBox 结构

**文件**: `runtime/layout/types.go`

```go
type LayoutBox struct {
    ID string

    // === 位置和尺寸 ===
    X, Y          int          // 相对于父节点的位置
    Width, Height int          // 尺寸

    // === 对齐 ===
    Baseline int              // 基线 (用于文本对齐)

    // === Layer 信息 (从 Fiber 继承) ===
    Layer layout.Layer        // ← 从 Fiber.Layer 映射而来

    // === 层内排序 ===
    ZIndex int                // Z-Index (在同一个 Layer 内排序)

    // === 边框 (可选) ===
    Border Border

    // === 子节点 ===
    Children []*LayoutBox     // 子节点布局结果
}
```

---

## Phase 3: Paint (LayoutBox → Buffer)

### 执行位置

**文件**: `internal/render/declarative_node.go:fiberFirstPaint()`

```go
// ========================================
// Phase 3: Paint using PaintableLayout
// ========================================

// 步骤 3.1: 从 layoutResult 获取 LayoutBox 根节点
var layoutBoxRoot *layout.LayoutBox
if adapter, ok := layoutResult.(*newLayoutResultAdapter); ok {
    layoutResultInner := adapter.GetLayoutResult()
    if layoutResultInner != nil {
        layoutBoxRoot = layoutResultInner.Root
    }
}

if layoutBoxRoot != nil {
    // 步骤 3.2: 转换为 PaintableLayout
    //     使用 Fiber 数据填充 PaintableBox 的 PaintableNode
    converter := NewFiberToPaintableConverter(fiberRoot)
    paintableLayout := converter.ConvertToLayout(layoutBoxRoot)

    if paintableLayout != nil && paintableLayout.Root != nil {
        // 步骤 3.3: 构建 PaintablePlanes (按 Layer 分组)
        planes := paint.NewPaintablePlanes()
        var buildPlanes func(box *paint.PaintableBox)
        buildPlanes = func(box *paint.PaintableBox) {
            if box == nil {
                return
            }
            // 添加到对应层级的 plane
            planes.AddToLayer(paint.RenderLayer(box.Layer), box)

            // 递归处理子节点
            for _, child := range box.Children {
                buildPlanes(child)
            }
        }
        buildPlanes(paintableLayout.Root)

        // 步骤 3.4: 使用 PaintablePlanes 绘制 (正确的 Z-Order)
        if err := n.paintEngine.PaintPaintablePlanes(planes, buf); err != nil {
            log.RenderLogger.Debug("PaintPaintablePlanes FAILED: %v, falling back", err)
            n.legacyPaint(ctx, buf)
            return
        }

        // 步骤 3.5: 保存 HitMap 用于事件路由
        if hitMap := layoutResult.GetHitMap(); hitMap != nil {
            n.fiberLastHitMap = hitMap
            log.RenderLogger.Debug("✅ Saved HitMap with %d entries", hitMap.Size())
        }

        log.RenderLogger.Debug("✅ PaintPaintablePlanes complete")
        return
    }
}

// Fallback: 使用旧的渲染路径
n.legacyPaint(ctx, buf)
```

### Step 3.1: LayoutBox → PaintableBox 转换

**文件**: `internal/render/converter.go`

```go
// FiberToPaintableConverter 将 LayoutBox 转换为 PaintableBox
type FiberToPaintableConverter struct {
    fiberRoot *reconciler.Fiber
    fiberMap  map[string]*reconciler.Fiber  // ID -> Fiber
}

// Convert 执行递归转换
func (c *FiberToPaintableConverter) Convert(
    lbox *layout.LayoutBox,
    parent *paint.PaintableBox,
) *paint.PaintableBox {
    if lbox == nil {
        return nil
    }

    // 创建 PaintableBox 并复制基本信息
    pbox := &paint.PaintableBox{
        X:        lbox.X,
        Y:        lbox.Y,
        Width:    lbox.Width,
        Height:   lbox.Height,
        Layer:    convertLayoutLayerToInt(lbox.Layer),  // ← Layer 转换
        ZIndex:   lbox.ZIndex,
        Parent:   parent,
        Children: make([]*paint.PaintableBox, 0, len(lbox.Children)),
    }

    // 查找匹配的 Fiber 并填充 paint-specific 数据
    if fiber := c.findFiber(lbox.ID); fiber != nil {
        c.fillFromFiber(pbox, fiber)
    }

    // 递归转换子节点
    for _, childLBox := range lbox.Children {
        childPBox := c.Convert(childLBox, pbox)
        if childPBox != nil {
            pbox.Children = append(pbox.Children, childPBox)
        }
    }

    return pbox
}

// layout.Layer → int 转换
func convertLayoutLayerToInt(l layout.Layer) int {
    // layout.Layer 是 types.Layer 的别名，types.Layer 是 int 类型
    // 因此可以直接转换
    return int(l)
}
```

### PaintableBox 结构

**文件**: `runtime/paint/paintable_box.go`

```go
type PaintableBox struct {
    // === 几何信息 ===
    X, Y, Width, Height int

    // === Layer 信息 (从 LayoutBox 继承) ===
    Layer int          // ← int(types.Layer)

    // === 层内排序 ===
    ZIndex int

    // === 绘制节点 ===
    Node PaintableNode  // FiberPaintableNode

    // === 树结构 ===
    Parent   *PaintableBox
    Children []*PaintableBox
}
```

### Step 3.2: 构建 PaintablePlanes

**文件**: `runtime/paint/paintable_planes.go`

```go
type PaintablePlanes struct {
    // planes 存储每层的 PaintableBox 集合
    // LayerBase(0) < LayerOverlay(1) < LayerModal(2) < LayerTooltip(3) < LayerInspector(4)
    planes map[RenderLayer][]*PaintableBox

    // renderOrder 存储渲染顺序（从低层到高层）
    renderOrder []RenderLayer
}

func NewPaintablePlanes() *PaintablePlanes {
    return &PaintablePlanes{
        planes: make(map[RenderLayer][]*PaintableBox),
        renderOrder: []RenderLayer{
            RenderLayerBase,      // 0
            RenderLayerOverlay,   // 1
            RenderLayerModal,     // 2
            RenderLayerTooltip,   // 3
            RenderLayerInspector, // 4
        },
    }
}

// AddToLayer 添加一个 PaintableBox 到指定层
func (pp *PaintablePlanes) AddToLayer(layer RenderLayer, box *paint.PaintableBox) {
    if box == nil {
        return
    }

    _, ok := pp.planes[layer]
    if !ok {
        pp.planes[layer] = make([]*PaintableBox, 0)
    }
    pp.planes[layer] = append(pp.planes[layer], box)
}

// GetRenderOrder 返回渲染顺序
func (pp *PaintablePlanes) GetRenderOrder() []RenderLayer {
    return pp.renderOrder
}

// GetLayer 返回指定层的所有 box
func (pp *PaintablePlanes) GetLayer(layer RenderLayer) []*PaintableBox {
    return pp.planes[layer]
}
```

### Step 3.3: 按 Layer 顺序渲染

**文件**: `internal/render/paint_engine.go`

```go
// PaintPaintablePlanes 按 renderOrder 顺序绘制所有层
func (e *PaintEngine) PaintPaintablePlanes(
    planes *paint.PaintablePlanes,
    buffer *paint.Buffer,
) error {
    for _, layer := range planes.GetRenderOrder() {
        boxes := planes.GetLayer(layer)
        if len(boxes) == 0 {
            continue
        }

        log.RenderLogger.Debug("Painting layer %s: %d boxes", layer.String(), len(boxes))

        // 按 Y 坐标排序 (优化渲染顺序)
        sort.Slice(boxes, func(i, j int) bool {
            return boxes[i].Y < boxes[j].Y
        })

        // 绘制该层的所有 box
        for _, box := range boxes {
            layout := paint.NewPaintableLayout(box)
            if err := e.PaintLayout(layout, buffer); err != nil {
                return fmt.Errorf("error painting box in layer %s: %w", layer.String(), err)
            }
        }

        // Modal 层特殊处理：绘制背景遮罩
        if layer == paint.RenderLayerModal && len(boxes) > 0 {
            e.paintModalBackdropBox(boxes[0], buffer)
            log.RenderLogger.Debug("Painted modal backdrop")
        }
    }

    return nil
}

// paintModalBackdropBox 绘制 Modal 背景遮罩
// 灰化非 Modal 区域 (dimming effect)
func (e *PaintEngine) paintModalBackdropBox(
    modalBox *paint.PaintableBox,
    buffer *paint.Buffer,
) {
    // 使用半透明背景或其他视觉效果来遮罩非 Modal 区域
    // 具体实现取决于视觉设计
}
```

### RenderLayer 类型定义

**文件**: `runtime/paint/paintable_planes.go`

```go
// RenderLayer 是 types.Layer 的别名 (向后兼容)
type RenderLayer = types.Layer

// String 方法用于日志输出
func (l RenderLayer) String() string {
    switch l {
    case LayerBase:
        return "base"
    case LayerOverlay:
        return "overlay"
    case LayerModal:
        return "modal"
    case LayerTooltip:
        return "tooltip"
    case LayerInspector:
        return "inspector"
    default:
        return "unknown"
    }
}
```

---

## Layer 传播机制

### 完整的数据流

```
┌──────────────────────────────────────────────────────────────────────────┐
│                           VNode (临时)                                   │
│   VNode.GetLayer() → rtui.Layer (types.Layer)                            │
│   - Component 每帧调用 renderFn() 返回新的 VNode 树                       │
└────────────────────┬─────────────────────────────────────────────────────┘
                     │
                     │ NewFiber()
                     │ Fiber.Layer = vnode.GetLayer()
                     ↓
┌──────────────────────────────────────────────────────────────────────────┐
│                           Fiber (持久)                                    │
│   fiber.Layer: rtui.Layer (types.Layer)                                  │
│   - 跨帧持久化存储                                                       │
│   - VNode 被丢弃后仍然保持                                               │
└────────────────────┬─────────────────────────────────────────────────────┘
                     │
                     │ FiberToNodeAdapterPure.GetLayer()
                     │ return layout.Layer(a.fiber.Layer)
                     ↓
┌──────────────────────────────────────────────────────────────────────────┐
│                       LayoutBox (布局结果)                                │
│   layout.Layer: types.Layer (类型别名)                                   │
│   - layout.Layer == types.Layer (零拷贝)                                 │
│   - 从 Fiber.Layer 直接映射                                              │
└────────────────────┬─────────────────────────────────────────────────────┘
                     │
                     │ convertLayoutLayerToInt()
                     │ return int(l)
                     ↓
┌──────────────────────────────────────────────────────────────────────────┐
│                     PaintableBox (渲染数据)                               │
│   paintableBox.Layer: int                                                │
│   - int(types.Layer)                                                     │
│   - 从 LayoutBox.Layer 转换而来                                          │
└────────────────────┬─────────────────────────────────────────────────────┘
                     │
                     │ planes.AddToLayer()
                     │ planes[RenderLayer(box.Layer)] = append(planes[...], box)
                     ↓
┌──────────────────────────────────────────────────────────────────────────┐
│                    PaintablePlanes (分层容器)                             │
│   planes map[int][]*PaintableBox:                                       │
│   ┌──────────────────────────────────────────────────┐                  │
│   │ planes[0]: [box1, box2, ...]  ← LayerBase       │                  │
│   │ planes[1]: [box3, box4, ...]  ← LayerOverlay    │                  │
│   │ planes[2]: [box5, ...]        ← LayerModal      │                  │
│   │ planes[3]: [box6, ...]        ← LayerTooltip    │                  │
│   │ planes[4]: [box7, ...]        ← LayerInspector  │                  │
│   └──────────────────────────────────────────────────┘                  │
│   renderOrder: [0, 1, 2, 3, 4]                                            │
└────────────────────┬─────────────────────────────────────────────────────┘
                     │
                     │ PaintEngine.PaintPaintablePlanes()
                     │ for _, layer := range planes.GetRenderOrder() { ... }
                     ↓
┌──────────────────────────────────────────────────────────────────────────┐
│                           Buffer (最终渲染)                               │
│   渲染顺序: 0 → 1 → 2 → 3 → 4                                            │
│   高层会覆盖低层的像素                                                   │
│   Modal 层额外绘制背景遮罩 (灰化非 Modal 区域)                             │
└──────────────────────────────────────────────────────────────────────────┘
```

### 类型统一体系

```go
// ==================== 统一类型定义 ====================

// 1. runtime/types/layer.go - 源类型定义
package types
type Layer int

const (
    LayerBase      Layer = iota  // 0
    LayerOverlay                 // 1
    LayerModal                   // 2
    LayerTooltip                 // 3
    LayerInspector               // 4
)


// 2. runtime/ui/fiber.go - 使用类型别名
package ui
type Layer = types.Layer

// Fiber 结构
type Fiber struct {
    Layer types.Layer  // ← 持久存储
}


// 3. runtime/layout/types.go - 使用类型别名
package layout
type Layer = types.Layer

// LayoutBox 结构
type LayoutBox struct {
    Layer layout.Layer  // ← 与 types.Layer 是同一类型
}


// 4. runtime/paint/paintable_planes.go - 使用类型别名
package paint
type RenderLayer = types.Layer

// PaintableBox 结构
type PaintableBox struct {
    Layer int  // ← int(types.Layer)
}


// 5. 统一类型验证
// 由于是类型别名，以下比较都是合法的:
var l types.Layer = types.LayerBase
var fl ui.Layer = ui.LayerBase       // ui.Layer == types.Layer
var ll layout.Layer = layout.LayerBase // layout.Layer == types.Layer
var pl paint.RenderLayer = paint.RenderLayerBase // paint.RenderLayer == types.Layer

// 比较示例:
if l == fl && l == ll && l == int(pl) {
    // 都等于 LayerBase (0)
}
```

### 零拷贝传递

```go
// FiberToNodeAdapterPure.GetLayer() 实现
func (a *FiberToNodeAdapter) GetLayer() layout.Layer {
    if a.fiber == nil {
        return layout.LayerBase
    }
    // 零拷贝: layout.Layer 是 types.Layer 的别名
    // types.Layer == ui.Layer == layout.Layer == paint.RenderLayer
    // 这是一个简单的类型转换，不涉及任何数据拷贝
    return layout.Layer(a.fiber.Layer)
}

// convertLayoutLayerToInt() 实现
func convertLayoutLayerToInt(l layout.Layer) int {
    // types.Layer 是 int 类型的子集
    // 这是一个简单的类型转换，成本极低
    return int(l)
}
```

---

## 关键代码路径

### 三阶段关键函数调用

#### Phase 1: Reconciliation

```
fiberFirstPaint()
    ↓
reconciler.Render(nullBuf, renderFn)
    ↓
renderFn()  // 用户提供的 VNode 生成函数
    ↓
reconciler.beginWork()  // 创建/更新 Fiber 节点
    ↓
NewFiber(vnodeType, tag, props, children)
    └─ Fiber.Layer = vnode.GetLayer()  // ← Layer 初始化
    ↓
reconciler.completeWork()  // 完成 Fiber 节点
```

#### Phase 2: Layout

```
fiberFirstPaint()
    ↓
newLayoutEngine.LayoutFiber(fiberRoot, constraints)
    ↓
NewFiberToNodeAdapterPure(fiber)
    └─ GetLayer() → layout.Layer(fiber.Layer)  // ← Layer 映射
    ↓
a.engine.Layout(node, layoutConstraints)
    ↓
layout.BoxTreeBuilder.Build()  // 递归构建 LayoutBox 树
    └─ layoutBox.Layer = node.GetLayer()  // ← Layer 传递
```

#### Phase 3: Paint

```
fiberFirstPaint()
    ↓
NewFiberToPaintableConverter(fiberRoot)
    ↓
converter.ConvertToLayout(layoutBoxRoot)
    ↓
converter.Convert(lbox, parent)
    └─ pbox.Layer = convertLayoutLayerToInt(lbox.Layer)  // ← Layer 转换
    ↓ (递归遍历所有子节点)
buildPlanes(paintableLayout.Root)
    ↓
planes.AddToLayer(paint.RenderLayer(box.Layer), box)
    ↓
paintEngine.PaintPaintablePlanes(planes, buf)
    ↓
for _, layer := range planes.GetRenderOrder()  // ← Layer 渲染顺序
    for _, box := range planes.GetLayer(layer)
        e.PaintLayout(box, buf)
```

### 关键文件清单

| 文件 | 核心函数/类型 | 职责 |
|-----|-------------|------|
| `internal/render/declarative_node.go` | `NewDeclarativeNodeFromFuncWithFiber()` | 构造函数，初始化组件 |
| `internal/render/declarative_node.go` | `initFiberFirstPipeline()` | 初始化 layout engine 和 paint engine |
| `internal/render/declarative_node.go` | `fiberFirstPaint()` | 三阶段渲染的主函数 |
| `runtime/ui/fiber_util.go` | `NewFiber()` | 创建 Fiber 节点，初始化 Layer |
| `runtime/ui/fiber.go` | `Fiber` 结构 | Fiber 节点定义，包含 Layer 字段 |
| `internal/render/layout_switcher.go` | `NewLayoutEngineAdapter.LayoutFiber()` | Phase 2: 布局执行 |
| `internal/render/fiber_adapter.go` | `FiberToNodeAdapterPure` | Fiber → LayoutBox 的 Layer 映射 |
| `runtime/layout/types.go` | `LayoutBox` 结构 | 布局结果，包含 Layer 字段 |
| `internal/render/converter.go` | `FiberToPaintableConverter` | LayoutBox → PaintableBox 的 Layer 转换 |
| `runtime/paint/paintable_planes.go` | `PaintablePlanes` | 分层容器，管理多层级渲染 |
| `internal/render/paint_engine.go` | `PaintEngine.PaintPaintablePlanes()` | Phase 3: 按层顺序绘制 |

---

## 数据流转图

### Layer 数据结构转换

```
           VNode (runtime/ui)
              │
              ├─ GetLayer()
              │   返回: rtui.Layer
              │   实际类型: types.Layer
              │
              ▼
         Fiber (runtime/ui)
              │
              ├─ Fiber.Layer
              │   类型: types.Layer
              │   存储: 持久化字段
              │
              ▼
    FiberToNodeAdapterPure (internal/render)
              │
              ├─ GetLayer()
              │   返回: layout.Layer
              │   映射: layout.Layer = types.Layer (零拷贝)
              │
              ▼
      LayoutBox (runtime/layout)
              │
              ├─ LayoutBox.Layer
              │   类型: layout.Layer (types.Layer 别名)
              │   存储: 字段
              │
              ▼
    FiberToPaintableConverter (internal/render)
              │
              ├─ Convert()
              │   映射: PaintableBox.Layer = int(LayoutBox.Layer)
              │
              ▼
    PaintableBox (runtime/paint)
              │
              ├─ PaintableBox.Layer
              │   类型: int
              │   存储: 字段
              │
              ▼
    PaintablePlanes (runtime/paint)
              │
              ├─ AddToLayer()
              │   分组: planes[RenderLayer(box.Layer)] = [...]
              │
              ▼
    PaintEngine (internal/render)
              │
              ├─ PaintPaintablePlanes()
              │   渲染: 按 renderOrder [0,1,2,3,4] 顺序绘制
              │
              ▼
          Buffer (最终渲染)
```

### 时间线

```
T=0:   NewDeclarativeNodeFromFuncWithFiber()
           ├─ 创建 reconciler
           ├─ newLayoutEngine = NewNewLayoutEngineAdapter()
           └─ paintEngine = NewPaintEngine()

T=1:   用户请求渲染 (每帧)

T=1+ε: fiberFirstPaint() 开始

T=1+ε1: Phase 1: Fiber Reconciliation
           ├─ reconciler.Render(nullBuf, renderFn)
           ├─ VNode 树被创建
           ├─ Fiber 树被更新 (fiber.Layer = VNode.GetLayer())
           └─ VNode 被丢弃

T=1+ε2: Phase 2: Layout
           ├─ newLayoutEngine.LayoutFiber(fiberRoot, constraints)
           ├─ FiberToNodeAdapterPure(fiber)
           ├─ layoutEngine.Layout(node, constraints)
           └─ LayoutBox 树被创建 (layoutBox.Layer = fiber.Layer)

T=1+ε3: Phase 3: Paint
           ├─ FiberToPaintableConverter.ConvertToLayout()
           │   └─ PaintableBox 树被创建 (pbox.Layer = int(lbox.Layer))
           ├─ buildPlanes() 遍历 PaintableBox 树
           │   └─ PaintablePlanes 被构建 (按 Layer 分组)
           └─ paintEngine.PaintPaintablePlanes(planes, buf)
               └─ 按 renderOrder 顺序绘制各个 Layer

T=Δ:   fiberFirstPaint() 完成
           ├─ Buffer 已更新
           └─ HitMap 已保存
```

---

## 总结

### fiberFirstPaint() 渲染路径特点

1. **清晰的阶段划分**
   - Phase 1: Reconciliation - Fiber 树构建
   - Phase 2: Layout - LayoutBox 树构建
   - Phase 3: Paint - Buffer 渲染

2. **零拷贝的 Layer 传递**
   - 所有包使用 `types.Layer` 统一类型
   - 类型别名实现，无需数据拷贝

3. **持久化的 Fiber 存储**
   - VNode 每帧创建后被丢弃
   - Fiber 跨帧持久化存储 Layer 信息

4. **自动的多层级渲染**
   - PaintablePlanes 按 Layer 分组
   - PaintEngine 按固定的 renderOrder 渲染

5. **Modal 特殊处理**
   - 自动绘制背景遮罩
   - 灰化非 Modal 区域

---

**文档版本**: 1.0
**创建日期**: 2026-03-01
**作者**: AI 分析
**状态**: ✅ 技术分析完成
