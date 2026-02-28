# Fiber-First 模式下的 Layer 系统完整说明

**Mint TUI Fiber-First 架构中多层级渲染的完整实现**

---

## 📋 目录

1. [架构概述](#架构概述)
2. [Layer 系统基础](#layer-系统基础)
3. [Fiber-First 渲染流程](#fiber-first-渲染流程)
4. [Layer 传播机制](#layer-传播机制)
5. [多层级渲染的实现](#多层级渲染的实现)
6. [组件 Layer API](#组件-layer-api)
7. [使用示例](#使用示例)
8. [调试与验证](#调试与验证)
9. [性能考虑](#性能考虑)
10. [常见问题](#常见问题)

---

## 架构概述

### 渲染流程概览

```
NewDeclarativeNodeFromFuncWithFiber(fn, fwApp)
    ↓
    ├─ initFiberFirstPipeline()  // 初始化组件
    │   ├─ newLayoutEngine = NewLayoutEngineAdapter()
    │   ├─ paintEngine = NewPaintEngine()
    │   └─ converter = NewFiberToPaintableConverter() (渲染时创建)
    │
    ↓
Paint(ctx, buf) [每次重绘调用]
    ↓fiberFirstPaint(ctx, buf)
    ├─ Phase 1: Fiber Reconciliation (VNode → Fiber)
    │       └─ reconciler.Render(nullBuf, n.renderFn)
    │
    ├─ Phase 2: Layout (Fiber → LayoutBox)
    │       └─ newLayoutEngine.LayoutFiber(fiberRoot, constraints)
    │           └─ FiberToNodeAdapterPure(fiber)
    │               └─ GetLayer() → layout.Layer
    │
    └─ Phase 3: Paint (LayoutBox → PaintableBox → Buffer)
        ├─ FiberToPaintableConverter.ConvertToLayout()
        │       └─ layout.Layer → paintableBox.Layer (int)
        │
        ├─ buildPlanes() 遍历 PaintableBox 树
        │       └─ planes.AddToLayer(RenderLayer(box.Layer), box)
        │
        └─ paintEngine.PaintPaintablePlanes(planes, buf)
            └─ 按 renderOrder [0,1,2,3,4] 依次渲染各层
```

### Fiber-First 核心要点

| 概念 | 旧架构 (VNode-first) | 新架构 (Fiber-first) |
|------|--------------------|---------------------|
| **状态存储** | VNode 临时存储 | Fiber 持久存储 |
| **Layer 传播** | VNode.GetLayer() | Fiber.Layer 字段 |
| **布局结果** | 每次重新计算 | 缓存在 Fiber.ComputedBox |
| **检测方式** | 优先 VNode 树 | 优先 Fiber 树 |

---

## Layer 系统基础

### Layer 枚举定义

**位置**: `runtime/types/layer.go`

```go
type Layer int

const (
    LayerBase      Layer = iota  // 0: 基础层
    LayerOverlay               // 1: 覆盖层 (下拉菜单、弹出框)
    LayerModal                 // 2: 模态层 (模态对话框、紧急提示)
    LayerTooltip               // 3: 提示层 (Tooltip、帮助信息)
    LayerInspector             // 4: 检查器层 (UI 调试覆盖层)
)
```

### Layer 特性

```go
func (l Layer) String() string          // "base", "overlay", ...
func (l Layer) ZIndex() int             // 返回 z-index 值
func (l Layer) IsValid() bool           // 检查是否有效
func (l Layer) IsModal() bool           // 是否是模态层
func (l Layer) IsOverlay() bool         // 是否是覆盖层
```

### Z-Order 计算方式

在 HitMap 中，Z-order 计算公式为：
```
zOrder = int(Layer) * 10000 + Depth
```

这确保了所有较高层级的节点都在所有较低层级节点之上，无论深度如何。

---

## Fiber-First 渲染流程

### 完整流程图

```
┌─────────────────────────────────────────────────────────────┐
│  1. 用户调用 ui.Run(app_func)                              │
│     ↓                                                        │
│  2. framework/App.Run() 初始化                             │
│     ↓                                                        │
│  3. internal/render 创建 DeclarativeNode                    │
│     ↓                                                        │
│  4. 每帧渲染: DeclarativeNode.Paint()                      │
│     ↓                                                        │
│  5. 调度器: scheduleRender()                               │
│     ├─ 创建/更新 Fiber 树 (Reconciler)                    │
│     └─ 执行 beginWork() / completeWork()                  │
│     ↓                                                        │
│  6. PipelineRenderer.Render()                              │
│     ├─ 获取 Fiber 树指针                                   │
│     ├─ hasLayerNodes(fiber) 检测                          │
│     └─ 调用 Render() 或 RenderLayers()                    │
│     ↓                                                        │
│  7. RenderingPipeline.RenderLayers()                       │
│     ├─ layout.Engine.Layout(fiber)                        │
│     ├─ applyLayerTransforms() (Modal 居中)                 │
│     ├─ buildPaintablePlanes() (按层级分组)                │
│     └─ PaintEngine.PaintPaintablePlanes()                 │
│     ↓                                                        │
│  8. buildHitMapFromPaintablePlanes() (事件路由)           │
│     ↓                                                        │
│  9. 输出到终端                                           │
└─────────────────────────────────────────────────────────────┘
```

### Phase 1: Fiber Reconciliation

**位置**: `internal/render/declarative_node.go:fiberFirstPaint()`

```go
// Phase 1: Fiber Reconciliation (VNode → Fiber)
// The reconciler updates the Fiber tree, VNode is discarded after this
// Use a minimal buffer for reconciliation (actual painting happens later)
nullBuf := paint.NewBuffer(1, 1)
n.reconciler.Render(component.PaintContext{
    Bounds: paint.Rect{X: 0, Y: 0, Width: 1, Height: 1},
}, nullBuf, n.renderFn)

// Get the Fiber root from reconciler
fiberRoot := n.getFiberRoot()
```

**Fiber 节点的 Layer 字段初始化** (`runtime/ui/fiber_util.go`):

```go
func NewFiber(...) *Fiber {
    return &Fiber{
        Type:    vnodeType,
        Tag:     tag,
        Props:   props,
        NodeID:  generateNodeID(),
        Layer:   vnode.GetLayer(),  // ← 从 VNode 获取初始 Layer
        // ... 其他字段
    }
}
```

**Fiber.Layer 类型定义**:

```go
// runtime/ui/fiber.go
type Fiber struct {
    // ...
    NodeID uint64
    Layer  rtui.Layer  // ← 持久化的 Layer 状态
    // ...
}

// Layer 是 types.Layer 的别名
type Layer = types.Layer
```

### Phase 2: Layout (Fiber → LayoutBox)

**位置**: `internal/render/declarative_node.go:fiberFirstPaint()`

```go
// Phase 2: Fiber-based Layout
// Use the new layout engine directly (runtime/layout), bypassing LayoutSwitcher
constraints := runtime.BoxConstraints{
    MinWidth:  0,
    MaxWidth:  ctx.AvailableWidth,
    MinHeight: 0,
    MaxHeight: ctx.AvailableHeight,
}

// Ensure the new layout engine is initialized
if n.newLayoutEngine == nil {
    n.newLayoutEngine = NewNewLayoutEngineAdapter()
}

// Perform layout using the new runtime/layout engine
layoutResult, err := n.newLayoutEngine.LayoutFiber(fiberRoot, constraints)
if err != nil {
    log.RenderLogger.Debug("Layout FAILED: %v, falling back to legacy", err)
    n.legacyPaint(ctx, buf)
    return
}
```

**LayoutFiber 实现** (`internal/render/layout_switcher.go`):

```go
func (a *NewLayoutEngineAdapter) LayoutFiber(
    fiber *reconciler.Fiber,
    constraints runtime.BoxConstraints,
) (LayoutResult, error) {
    // Use Fiber-only adapter (no VNode dependency)
    node := NewFiberToNodeAdapterPure(fiber)

    // Convert constraints
    layoutConstraints := layout.Constraints{
        MinWidth:  constraints.MinWidth,
        MaxWidth:  constraints.MaxWidth,
        MinHeight: constraints.MinHeight,
        MaxHeight: constraints.MaxHeight,
    }

    // Perform layout
    result := a.engine.Layout(node, layoutConstraints)

    return &newLayoutResultAdapter{result: result, fiberRoot: fiber}, nil
}
```

**Fiber → LayoutBox 的 Layer 传递** (`internal/render/fiber_adapter.go`):

```go
// FiberToNodeAdapterPure 实现 layout.Node 接口
func (a *FiberToNodeAdapter) GetLayer() layout.Layer {
    if a.fiber == nil {
        return layout.LayerBase
    }
    return layout.Layer(a.fiber.Layer)  // ← Fiber.Layer → layout.Layer
}
```

**LayoutBox 结构** (`runtime/layout/types.go`):

```go
type LayoutBox struct {
    ID string

    // X, Y 位置（相对于父节点）
    X, Y int

    // Width, Height 尺寸
    Width, Height int

    // Baseline 基线（用于文本对齐）
    Baseline int

    // Layer 渲染层（用于多层渲染）
    Layer layout.Layer  // ← 从 Fiber 继承

    // ZIndex 层内排序索引
    ZIndex int

    // Border 边框信息（如果有）
    Border Border

    // Children 子节点布局结果
    Children []*LayoutBox
}
```

### Phase 3: Paint (LayoutBox → PaintableBox → Buffer)

**位置**: `internal/render/declarative_node.go:fiberFirstPaint()`

```go
// Phase 3: Paint using PaintableLayout
// Convert LayoutResult to PaintableLayout and use PaintEngine
if layoutResult != nil {
    // Get LayoutBox from adapter
    var layoutBoxRoot *layout.LayoutBox
    if adapter, ok := layoutResult.(*newLayoutResultAdapter); ok {
        layoutResultInner := adapter.GetLayoutResult()
        if layoutResultInner != nil {
            layoutBoxRoot = layoutResultInner.Root
        }
    }

    if layoutBoxRoot != nil {
        // Convert LayoutBox to PaintableLayout using Fiber data
        converter := NewFiberToPaintableConverter(fiberRoot)
        paintableLayout := converter.ConvertToLayout(layoutBoxRoot)

        if paintableLayout != nil && paintableLayout.Root != nil {
            // Build PaintablePlanes for multi-layer rendering
            planes := paint.NewPaintablePlanes()
            var buildPlanes func(box *paint.PaintableBox)
            buildPlanes = func(box *paint.PaintableBox) {
                if box == nil {
                    return
                }
                planes.AddToLayer(paint.RenderLayer(box.Layer), box)
                for _, child := range box.Children {
                    buildPlanes(child)
                }
            }
            buildPlanes(paintableLayout.Root)

            // Paint using PaintablePlanes for proper layer Z-Ordering
            if err := n.paintEngine.PaintPaintablePlanes(planes, buf); err != nil {
                log.RenderLogger.Debug("PaintPaintablePlanes FAILED: %v, falling back", err)
                n.legacyPaint(ctx, buf)
                return
            }

            // Save HitMap for event routing
            if hitMap := layoutResult.GetHitMap(); hitMap != nil {
                n.fiberLastHitMap = hitMap
                log.RenderLogger.Debug("✅ Saved HitMap with %d entries", hitMap.Size())
            }

            log.RenderLogger.Debug("✅ PaintPaintablePlanes complete")
            return
        }
    }
}

n.legacyPaint(ctx, buf)  // Fallback
```

#### Step 3.1: LayoutBox → PaintableBox 转换

**FiberToPaintableConverter.Convert()** (`internal/render/converter.go`):

```go
func (c *FiberToPaintableConverter) Convert(
    lbox *layout.LayoutBox,
    parent *paint.PaintableBox,
) *paint.PaintableBox {
    if lbox == nil {
        return nil
    }

    pbox := &paint.PaintableBox{
        X:        lbox.X,
        Y:        lbox.Y,
        Width:    lbox.Width,
        Height:   lbox.Height,
        Layer:    convertLayoutLayerToInt(lbox.Layer),  // layout.Layer → int
        ZIndex:   lbox.ZIndex,
        Parent:   parent,
        Children: make([]*paint.PaintableBox, 0, len(lbox.Children)),
    }

    // Find matching Fiber and fill paint-specific data
    if fiber := c.findFiber(lbox.ID); fiber != nil {
        c.fillFromFiber(pbox, fiber)
    }

    // Recursively convert children
    for _, childLBox := range lbox.Children {
        childPBox := c.Convert(childLBox, pbox)
        if childPBox != nil {
            pbox.Children = append(pbox.Children, childPBox)
        }
    }

    return pbox
}

func convertLayoutLayerToInt(l layout.Layer) int {
    return int(l)  // 由于 layout.Layer 是 types.Layer 的别名，直接返回即可
}
```

#### Step 3.2: 构建 PaintablePlanes

**PaintablePlanes 结构** (`runtime/paint/paintable_planes.go`):

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
            RenderLayerBase,
            RenderLayerOverlay,
            RenderLayerModal,
            RenderLayerTooltip,
            RenderLayerInspector,
        },
    }
}

// AddToLayer 添加一个 PaintableBox 到指定层
func (pp *PaintablePlanes) AddToLayer(layer RenderLayer, box *paint.PaintableBox) {
    if box == nil { return }

    _, ok := pp.planes[layer]
    if !ok {
        pp.planes[layer] = make([]*PaintableBox, 0)
    }
    pp.planes[layer] = append(pp.planes[layer], box)
}
```

#### Step 3.3: 按 Layer 顺序渲染

**PaintEngine.PaintPaintablePlanes()** (`internal/render/paint_engine.go`):

```go
func (e *PaintEngine) PaintPaintablePlanes(
    planes *paint.PaintablePlanes,
    buffer *paint.Buffer,
) error {
    for _, layer := range planes.GetRenderOrder() {
        boxes := planes.GetLayer(layer)
        if len(boxes) == 0 { continue }

        for _, box := range boxes {
            layout := paint.NewPaintableLayout(box)
            if err := e.PaintLayout(layout, buffer); err != nil {
                return fmt.Errorf("error painting box in layer %s: %w", layer.String(), err)
            }
        }

        // Modal 层特殊处理：绘制背景遮罩
        if layer == paint.RenderLayerModal && len(boxes) > 0 {
            e.paintModalBackdropBox(boxes[0], buffer)
        }
    }
    return nil
}
```

---

## Layer 传播机制

### VNode → Fiber 的传播

```
┌─────────────────┐
│   VNode         │  创建时调用 GetLayer()
│   GetLayer()    │ ──────────────────┐
└─────────────────┘                    │
                                      ↓
                            ┌─────────────────────┐
                            │     Fiber            │
                            │     Layer 字段       │
                            └─────────────────────┘
```

### 传播时机

1. **Fiber 创建时** (`runtime/ui/fiber_util.go:197`):
   ```go
   Layer: vnode.GetLayer(),  // 初始 Layer
   ```

2. **Fiber 更新时** (`reconciler.diff.go`):
   ```go
   // 如果 Layer 改变，需要更新 Fiber.Layer
   if newLayer != oldLayer {
       workInProgressFiber.Layer = newLayer
   }
   ```

3. **子节点继承**:
   - 子节点如果未显式设置 Layer，继承父节点的 Layer
   - 这在 `NewFiber()` 中通过 `vnode.GetLayer()` 自动处理

### FiberVNode 的 Layer 访问

```go
type FiberVNode struct {
    fiber *Fiber
}

func (f *FiberVNode) GetLayer() Layer {
    if f.fiber == nil {
        return LayerBase
    }
    return f.fiber.Layer  // 从 Fiber 读取
}

func (f *FiberVNode) SetLayer(l Layer) VNode {
    if f.fiber != nil {
        f.fiber.Layer = l  // 直接修改 Fiber.Layer
    }
    return f
}
```

---

## 多层级渲染的实现

### PaintablePlanes 结构

**位置**: `runtime/paint/paintable_planes.go`

```go
type PaintablePlanes struct {
    planes map[int]*PaintablePlane  // key: layer index
}

type PaintablePlane struct {
    layer   int
    boxes   []*PaintableBox
}
```

### 层级构建过程

**位置**: `internal/render/rendering_pipeline.go:buildPaintablePlanes()`

```go
func (p *RenderingPipeline) buildPaintablePlanes(root *paint.PaintableBox) *paint.PaintablePlanes {
    planes := paint.NewPaintablePlanes()

    var walk func(box *paint.PaintableBox)
    walk = func(box *paint.PaintableBox) {
        if box == nil {
            return
        }

        // 添加到对应层级的 plane
        planes.AddBox(box)

        // 递归处理子节点
        for _, child := range box.Children {
            walk(child)
        }
    }

    walk(root)
    return planes
}
```

### 层级绘制顺序

**低层级 → 高层级** (渲染时):

```
LayerBase (0)      → 先绘制（底层）
LayerOverlay (1)   →
LayerModal (2)     → 后绘制（覆盖上层）
LayerTooltip (3)   →
LayerInspector (4) → 最后绘制（顶层）
```

### HitMap 构建顺序

**高层级 → 低层级** (点击测试时):

```
LayerInspector (4) → 先检测（优先）
LayerTooltip (3)   →
LayerModal (2)     →
LayerOverlay (1)   →
LayerBase (0)      → 最后检测
```

### Modal 特殊处理

**位置**: `internal/render/rendering_pipeline.go:centerPaintableModalBox()`

```go
func (p *RenderingPipeline) centerPaintableModalBox(
    box *paint.PaintableBox,
    constraints layout.Constraints,
) {
    modalWidth := box.Width
    modalHeight := box.Height
    containerWidth := constraints.MaxWidth
    containerHeight := constraints.MaxHeight

    // 计算居中偏移
    offsetX := (containerWidth - modalWidth) / 2
    offsetY := (containerHeight - modalHeight) / 2

    // 偏移整个 Modal 树
    p.shiftPaintableBoxTree(box, offsetX, offsetY)
}
```

---

## 组件 Layer API

### VNode 接口

**位置**: `runtime/ui/vnode.go`

```go
type VNode interface {
    // ...
    GetLayer() Layer
    SetLayer(Layer) VNode
}
```

### 实现 GetLayer/SetLayer 的组件

| 组件 | 默认层 | 支持方法 |
|------|--------|---------|
| `ElementVNode` | LayerBase | `GetLayer()`, `SetLayer()` |
| `TextVNode` | LayerBase | `GetLayer()` (不可设置) |
| `ComponentVNode` | LayerBase | `GetLayer()`, `SetLayer()` |
| `MemoVNode` | LayerBase | `GetLayer()`, `SetLayer()` |
| `LayoutVNode` (HStack, VStack 等) | LayerBase | `GetLayer()`, `SetLayer()` |
| `FiberVNode` | LayerBase | `GetLayer()`, `SetLayer()` |

### Tooltip 组件示例

**位置**: `ui/components/tooltip/vnode.go`

```go
type VNode struct {
    content   rtui.VNode
    text      string
    position  Position
    tooltip   *rtText.Text
    show      bool
    layer     rtui.Layer  // ← 新增持久化 Layer 字段
}

// GetLayer 返回当前层级
func (t *VNode) GetLayer() rtui.Layer {
    return t.layer
}

// SetLayer 设置层级
func (t *VNode) SetLayer(l rtui.Layer) rtui.VNode {
    t.layer = l
    return t
}

// New() 创建新 Tooltip
func New(content rtui.VNode, text string) *VNode {
    return &VNode{
        content: content,
        text:    text,
        layer:   rtui.LayerTooltip,  // ← 默认 Tooltip 层
    }
}
```

### Tooltip Builder API

**位置**: `ui/components/tooltip/builder.go`

```go
type Builder struct {
    content  rtui.VNode
    text     string
    position Position
    layer    rtui.Layer  // ← 新增
}

func Layer(l rtui.Layer) func(*Builder) {
    return func(b *Builder) {
        b.layer = l
    }
}

// 便捷方法
func BaseLayer() func(*Builder)       { return Layer(rtui.LayerBase) }
func OverlayLayer() func(*Builder)    { return Layer(rtui.LayerOverlay) }
func ModalLayer() func(*Builder)      { return Layer(rtui.LayerModal) }
func TooltipLayer() func(*Builder)    { return Layer(rtui.LayerTooltip) }
func InspectorLayer() func(*Builder)  { return Layer(rtui.LayerInspector) }

// 使用示例
tooltip.NewBuilder(content, "Help info").
    Layer(rtui.LayerTooltip).   // 明确指定
    Build()

tooltip.NewBuilder(content, "Important").
    ModalLayer().               // 便捷方法
    Build()
```

---

## 使用示例

### 示例 1: Modal 对话框

```go
package main

import (
    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/ui/components/button"
    "github.com/wwsheng009/mint/ui/components/modal"
)

type AppState struct {
    showModal bool
}

func (s *AppState) App() ui.VNode {
    // 背景内容 (Base 层)
    content := ui.HStack(
        ui.NewText("Welcome to the app"),
        UI.NewButton("Open Modal", func() {
            s.showModal = true
        }),
    )

    // Modal (Modal 层)
    if s.showModal {
        modalContent := modal.New(
            ui.NewText("Are you sure?"),
            modal.WithActions(
                button.New("Cancel", func() { s.showModal = false }),
                button.New("Confirm", func() { s.showModal = false }),
            ),
        )
        modalContent.SetLayer(ui.LayerModal)  // ← 显式设置为 Modal 层

        return ui.VStack(content, modalContent)
    }

    return content
}
```

### 示例 2: 多个 Tooltip 不同层级

```go
package main

import (
    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/ui/components/tooltip"
)

func MultiTooltipExample() ui.VNode {
    button1 := ui.NewButton("Normal Tip")
    tooltip1 := tooltip.NewBuilder(button1, "This is a normal tooltip").
        TooltipLayer().  // Tooltip 层 (3)
        Build()

    button2 := ui.NewButton("Important Notice")
    tooltip2 := tooltip.NewBuilder(button2, "WARN: This action is irreversible!").
        ModalLayer().    // Modal 层 (2) - 更显眼
        Build()

    button3 := ui.NewButton("Debug Info")
    tooltip3 := tooltip.NewBuilder(button3, "DEBUG: Handler ID=12345").
        InspectorLayer(). // Inspector 层 (4) - 调试用
        Build()

    return ui.VStack(tooltip1, tooltip2, tooltip3)
}
```

### 示例 3: Toast 通知不同优先级

```go
package main

import (
    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/ui/components/tooltip"
)

type NotificationManager struct {
    toastManager *tooltip.ToastManager
}

func (nm *NotificationManager) ShowInfo(msg string) {
    nm.toastManager.Add(
        tooltip.NewToastBuilder(msg).Info().OverlayLayer().Build(),
    )
}

func (nm *NotificationManager) ShowError(msg string) {
    nm.toastManager.Add(
        tooltip.NewToastBuilder(msg).Error().ModalLayer().Build(),
    )
}

func (nm *NotificationManager) ShowDebug(msg string) {
    nm.toastManager.Add(
        tooltip.NewToastBuilder(msg).InspectorLayer().Build(),
    )
}
```

### 示例 4: 自定义 Inspector 覆盖层

```go
package main

import (
    "github.com/wwsheng009/mint/ui"
    "github.com/wwsheng009/mint/internal/inspector"
)

var globalInspector *inspector.StandaloneInspector

func AppWithInspector() ui.VNode {
    appContent := buildAppContent()

    // Inspector 覆盖层
    if globalInspector.IsVisible() {
        inspectorOverlay := globalInspector.RenderOverlay()
        inspectorOverlay.SetLayer(ui.LayerInspector)

        return ui.VStack(appContent, inspectorOverlay)
    }

    return appContent
}
```

---

## 调试与验证

### 启用日志输出

```bash
# Layer 系统
export TUI_LAYER_DEBUG=true

# 渲染流程
export TUI_DEBUG_RENDER=true

# HitMap
export TUI_DEBUG_HITMAP=true

# 布局
export TUI_DEBUG_LAYOUT=true
```

### 预期日志输出

```
[PipelineRenderer] Buffer size: 120x40
[PipelineRenderer] hasLayers=true
[PipelineRenderer] Using RenderLayers for multi-layer rendering
[RenderingPipeline] RenderLayers started (layout.Engine path)
[RenderingPipeline] Layout complete, root size=120x40
[RenderingPipeline] Apply layer transforms (Modal centering)
[RenderingPipeline] PaintablePlanes: 5 planes, 42 boxes
[PaintEngine] Painting 5 layers (low → high)
[PaintEngine] Layer 0 (base): 25 boxes
[PaintEngine] Layer 1 (overlay): 8 boxes
[PaintEngine] Layer 2 (modal): 3 boxes
[PaintEngine] Layer 3 (tooltip): 5 boxes
[PaintEngine] Layer 4 (inspector): 1 box
[RenderingPipeline] RenderLayers complete
```

### 验证 Layer 检测

```go
// 在应用代码中手动验证
func debugLayerDetection(root ui.VNode) {
    fiber := root.(*rtui.FiberVNode).Fiber()
    hasLayers := checkLayerNodes(fiber)
    log.Printf("Has layer nodes: %v\n", hasLayers)
}

func checkLayerNodes(fiber *rtui.Fiber) bool {
    if fiber == nil {
        return false
    }
    if fiber.Layer != rtui.LayerBase && fiber.Layer.IsValid() {
        log.Printf("Found layer node: ID=%d, Layer=%v\n", fiber.NodeID, fiber.Layer)
        return true
    }
    for child := fiber.Child; child != nil; child = fiber.Sibling {
        if checkLayerNodes(child) {
            return true
        }
    }
    return false
}
```

### 可视化 Layer 布局

使用 SVG 布局可视化工具:

```bash
cd examples/svg_layout_visualization
go run main.go
```

可生成展示不同层级布局关系的 SVG 文件。

---

## 性能考虑

### 性能开销

| 操作 | 开销 | 说明 |
|------|------|------|
| `hasLayerNodes()` | < 1ms | 递归遍历 Fiber 树 |
| `RenderLayers()` vs `Render()` | +5-15% | 仅在有 Layer 时 |
| PaintablePlanes 构建 | +2-5% | 按层级分组 |
| HitMap 构建 | +3-8% | Z-order 排序 |

### 优化建议

1. **避免过度使用 Layer**
   - 只在需要时设置 Layer
   - 大多数组件应使用默认的 LayerBase

2. **限制 Modal 数量**
   - 同时显示的 Modal 不应超过 1-2 个
   - 避免 Modal 内嵌

3. **利用 Fiber 缓存**
   - Fiber.ComputedBox 缓存布局结果
   - 避免每帧重新计算

4. **调试时禁用日志**
   ```go
   // 生产环境禁用调试日志
   log.LayerLogger.Disable()
   log.RenderLogger.Disable()
   log.PipelineLogger.Disable()
   ```

### 性能测量

```go
import (
    "time"
    "github.com/wwsheng009/mint/internal/log"
)

func measureRenderTime() {
    start := time.Now()

    // 渲染操作
    renderer.Render(vnode, 0, 0, buffer)

    elapsed := time.Since(start)
    log.PipelineLogger.Debug("Render took %v\n", elapsed)
}
```

---

## 常见问题

### Q1: 为什么我的 Tooltip 不显示在最上层？

**A:** 检查以下几点:

1. **Layer 是否正确设置?**
   ```go
   t := tooltip.New(content, "info")
   t.SetLayer(rtui.LayerTooltip)  // 确保使用 Tooltip 层
   ```

2. **Fiber 树中的 Layer 是否传播?**
   - 使用 `FiberVNode` 而不是普通 `VNode`
   - 检查日志: `[hasLayerNodes] Found layer node`

3. **是否有更高层的元素遮挡?**
   - Inspector 层 (4) > Tooltip 层 (3)
   - 检查 Inspector 是否已启用

### Q2: 为什么 Modal 不居中？

**A:** Modal 居中逻辑 (`centerPaintableModalBox`) 需要:

1. **Modal 在 `LayerModal` 层**
   ```go
   modal.SetLayer(rtui.LayerModal)
   ```

2. **容器有明确的尺寸**
   - 使用 `NewBoxConstraints(minW, maxW, minH, maxH)`
   - 确保 maxW/maxH 不是 Infinity

3. **Modal 内容不自适应**
   - Modal 内容应有固定或最大尺寸

### Q3: 如何动态改变组件的 Layer？

**A:** 使用 SetLayer() 并触发重绘:

```go
type DynamicLayerComponent struct {
    currentLayer rtui.Layer
}

func (c *DynamicLayerComponent) Render() ui.VNode {
    content := ui.NewText("Content")
    content.SetLayer(c.currentLayer)

    return ui.HStack(
        content,
        ui.NewButton("Change Layer", func() {
            c.currentLayer = rtui.LayerOverlay
            // 触发重绘 (由框架自动处理)
        }),
    )
}
```

### Q4: Fiber-first 模式下的 VNode Layer 设置会被保留吗？

**A:** 会，但需要注意:

1. **VNode 的 Layer 在 Fiber 创建时复制**:
   ```go
   fiber.Layer = vnode.GetLayer()  // 初始化时
   ```

2. **修改 Fiber.Layer 后，VNode 的 Layer 不会同步更新**:
   ```go
   fiberVNode.SetLayer(rtui.LayerModal)  // 修改 Fiber.Layer
   // fiberVNode.fiber.Layer = rtui.LayerModal
   ```

3. **推荐使用 FiberVNode.SetLayer()**:
   - 这会直接修改 `Fiber.Layer`
   - 下一帧渲染时会使用新的 Layer

### Q5: 如何检查某个 Fiber 在哪个 Layer？

**A:**:

```go
func getFiberLayer(fiber *rtui.Fiber) rtui.Layer {
    if fiber == nil {
        return rtui.LayerBase
    }
    return fiber.Layer
}

// 调试输出
fiber := root.(*rtui.FiberVNode).Fiber()
layer := getFiberLayer(fiber)
fmt.Printf("Fiber Layer: %v (%d)\n", layer, layer.ZIndex())
```

### Q6: 为什么 hasLayerNodesFromFiber 检测失败？

**A:** 可能原因:

1. **Fiber 为 nil**:
   - 检查 `PipelineRenderer.fiber` 是否已设置

2. **所有节点都是 LayerBase**:
   ```go
   if layer != rtui.LayerBase && layer.IsValid() {
       return true  // 需要非 Base 层
   }
   ```

3. **递归检查子节点**:
   - 确保遍历完整树: `Child → Sibling`

---

## 相关文档

### 实现文档

- `internal/render/pipeline_renderer.go` - PipelineRenderer 实现
- `internal/render/rendering_pipeline.go` - RenderingPipeline 实现
- `runtime/ui/fiber.go` - Fiber 结构定义
- `runtime/ui/fiber_util.go` - Fiber 创建工具
- `runtime/types/layer.go` - Layer 类型定义

### 用户文档

- `ui/components/tooltip/layer_demo.go` - Tooltip Layer 使用示例
- `docs/layer/LAYER_SYSTEM_ARCHITECTURE.md` - 架构说明 (已更新)
- `docs/layer/LAYER_SYSTEM_IMPLEMENTATION_SUMMARY.md` - 实施总结 (已更新)

### 历史文档

- `docs/layer/TWO_RENDERING_SYSTEMS_EXPLAINED.md` - **已过时** (基于错误假设)

---

## 总结

### Fiber-First Layer 系统核心要点

1. **Layer 存储在 Fiber 节点中**
   - 持久化存储，跨帧保持
   - 从 VNode 初始化: `Fiber.Layer = VNode.GetLayer()`
   - 直接访问: `fiber.Layer` (types.Layer 类型)

2. **完整的 Fiber-first 渲染流程**
   - **Phase 1**: Fiber Reconciliation (VNode → Fiber，VNode 被丢弃)
   - **Phase 2**: Layout (Fiber → LayoutBox，通过 FiberToNodeAdapterPure)
   - **Phase 3**: Paint (LayoutBox → PaintableBox → Buffer，通过 PaintPaintablePlanes)

3. **Layer 传播路径 (零拷贝传递)**
   ```
   VNode.GetLayer()
       ↓
   Fiber.Layer (持久化存储)
       ↓
   FiberToNodeAdapterPure.GetLayer() → layout.Layer
       ↓
   LayoutBox.Layer
       ↓
   PaintableBox.Layer (int)
       ↓
   PaintablePlanes 分组
       ↓
   PaintEngine.PaintPaintablePlanes() 按层序渲染
   ```

4. **统一类型体系**
   - 所有包使用 `runtime/types.Layer` 统一类型
   - `layout.Layer` 和 `paint.RenderLayer` 都是类型别名
   - 零拷贝传递，无需转换

5. **层级绘制顺序**
   - 渲染: Base (0) → Overlay (1) → Modal (2) → Tooltip (3) → Inspector (4)
   - HitMap: Inspector (4) → Tooltip (3) → Modal (2) → Overlay (1) → Base (0)

6. **Modal 特殊处理**
   - 自动绘制背景遮罩（灰化非 Modal 区域）
   - 由 `paintModalBackdropBox()` 处理

7. **组件 Layer API**
   - `GetLayer()` / `SetLayer()`
   - Builder 便捷方法: `BaseLayer()`, `OverlayLayer()`, `ModalLayer()`, `TooltipLayer()`, `InspectorLayer()`

---

### fiberFirstPaint() 关键代码位置

| 步骤 | 位置 | 说明 |
|------|------|------|
| Phase 1: Reconciliation | `declarative_node.go:fiberFirstPaint()` | `reconciler.Render(nullBuf, fn)` |
| Phase 2: Layout | `layout_switcher.go:LayoutFiber()` | `NewLayoutEngineAdapter.LayoutFiber()` |
| Fiber→Node Adapter | `fiber_adapter.go:FiberToNodeAdapterPure` | `GetLayer()` 实现 |
| Phase 3.1: Layout→Paintable | `converter.go:FiberToPaintableConverter` | `Convert()` 方法 |
| Phase 3.2: Build Planes | `declarative_node.go:buildPlanes()` | 遍历树并分组 |
| Phase 3.3: Paint Planes | `paint_engine.go:PaintPaintablePlanes()` | 按 renderOrder 渲染 |

---

**文档版本**: 2.0
**最后更新**: 2026-03-01
**状态**: ✅ 当前实现 - Fiber-first 渲染路径分析
**关键更新**: 添加了 fiberFirstPaint() 三阶段详细流程分析
