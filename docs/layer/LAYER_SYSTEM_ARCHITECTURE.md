# Layer 系统架构说明

**Mint TUI Fiber-First 架构中的渲染层级系统**

---

## 📋 文档导航

| 文档 | 说明 |
|------|------|
| **本文档** | Layer 系统高层架构概览 |
| `FIBER_FIRST_LAYER_SYSTEM.md` | Fiber-First Layer 系统完整技术细节 |
| `TW_RENDERING_SYSTEMS_EXPLAINED.md` | ~~历史分析（已过时）~~ |

---

## 🎯 核心概念

### 单一渲染架构

Mint TUI 使用**统一的渲染架构**，不是两套并存的系统：

```
framework/App
    ↓
DeclarativeNode (Paintable 接口)
    ↓
PipelineRenderer
    ↓
    ├─→ hasLayerNodes() 自动检测
    │   ├─ Fiber 树检查 (优先)
    │   └─ VNode 树检查 (回退)
    │
    └─→ RenderingPipeline
            ├─→ RenderLayers() (多层级)
            └─→ Render() (单层级)
```

### Fiber-First 架构

| 层级 | 说明 |
|------|------|
| **VNode** | 临时描述结构，每帧重建 |
| **Fiber** | 持久化节点，存储状态跨帧保持 |
| **Layer** | 存储在 Fiber.Layer 字段中 |

---

## 🏗️ 架构组件

### 1. Layer 枚举

**位置**: `runtime/types/layer.go`

```go
type Layer int

const (
    LayerBase      Layer = iota  // 0
    LayerOverlay                 // 1
    LayerModal                   // 2
    LayerTooltip                 // 3
    LayerInspector               // 4
)
```

**特性**:
- `String()` - 字符串表示
- `ZIndex()` - z-index 值
- `IsValid()` - 有效性检查
- `IsModal()` - 模态层判断
- `IsOverlay()` - 覆盖层判断

### 2. VNode 接口

**位置**: `runtime/ui/vnode.go`

```go
type VNode interface {
    // ...
    GetLayer() Layer
    SetLayer(Layer) VNode
}
```

### 3. Fiber 结构

**位置**: `runtime/ui/fiber.go`

```go
type Fiber struct {
    // ...
    NodeID uint64
    Layer  Layer  // ← 持久化的 Layer 状态

    // 布局结果 (缓存)
    ComputedBox interface{}
}
```

### 4. 渲染流程

#### PipelineRenderer (Layer 检测)

**位置**: `internal/render/pipeline_renderer.go`

```go
func (r *PipelineRenderer) Render(vnode rtui.VNode, x, y int, buffer interface{}) error {
    // 应用 VNode 钩子
    vnode = r.hooks.ApplyVNodeHooks(vnode)

    // 检测是否有 Layer 节点
    hasLayers := r.hasLayerNodes(vnode)

    if hasLayers {
        // 多层级渲染
        log.RenderLogger.Debug("Using RenderLayers for multi-layer rendering")
        err = r.pipeline.RenderLayers(vnode, r.fiber, constraints, buf)
    } else {
        // 单层级渲染
        log.RenderLogger.Debug("Using standard Render")
        err = r.pipeline.Render(vnode, r.fiber, constraints, buf)
    }

    return err
}
```

#### hasLayerNodes() (检测逻辑)

```go
func (r *PipelineRenderer) hasLayerNodes(vnode rtui.VNode) bool {
    if vnode == nil {
        return false
    }

    // Fiber 树检查 (优先 - 更准确)
    if r.fiber != nil {
        return r.hasLayerNodesFromFiber(r.fiber)
    }

    // VNode 树检查 (回退 - 兼容非 Fiber 模式)
    return r.hasLayerNodesFromVNode(vnode)
}

func (r *PipelineRenderer) hasLayerNodesFromFiber(fiber *rtui.Fiber) bool {
    if fiber == nil {
        return false
    }

    // 检查此节点
    layer := fiber.Layer
    if layer != rtui.LayerBase && layer.IsValid() {
        log.HitMapLogger.Debug("[hasLayerNodes] ✅ Found layer node: Layer=%d", layer)
        return true
    }

    // 递归检查子节点 (Child → Sibling)
    for child := fiber.Child; child != nil; child = child.Sibling {
        if r.hasLayerNodesFromFiber(child) {
            return true
        }
    }

    return false
}
```

#### RenderingPipeline.RenderLayers()

**位置**: `internal/render/rendering_pipeline.go`

```go
func (p *RenderingPipeline) RenderLayers(
    vnode rtui.VNode,
    fiber *reconciler.Fiber,
    constraints runtime.BoxConstraints,
    buffer *paint.Buffer,
) error {
    // 1. 选择适配器
    var node layout.Node
    var converter PaintableConverter

    if fiber != nil {
        // Fiber-first 路径
        node = NewFiberToNodeAdapterPure(fiber)
        converter = NewFiberToPaintableConverter(fiber)
    } else {
        // VNode 路径 (回退)
        node = NewVNodeToNodeAdapter(vnode)
        converter = NewVNodeToPaintableConverter(vnode)
    }

    // 2. 执行布局
    result := p.layoutEngine.Layout(node, layoutConstraints)

    // 3. 转换为 PaintableLayout
    paintableLayout := converter.ConvertToLayout(result.Root)

    // 4. 应用 Layer 变换 (Modal 居中等)
    p.applyLayerTransformsToPaintable(paintableLayout.Root, layoutConstraints)

    // 5. 构建 PaintablePlanes (按层级分组)
    paintablePlanes := p.buildPaintablePlanes(paintableLayout.Root)

    // 6. 绘制
    if err := p.paintEngine.PaintPaintablePlanes(paintablePlanes, buffer); err != nil {
        return err
    }

    // 7. 构建 HitMap (事件路由)
    p.lastHitMap = p.buildHitMapFromPaintablePlanes(paintablePlanes)

    return nil
}
```

### 5. PaintablePlanes (层级平面)

**位置**: `runtime/paint/paintable_planes.go`

```go
type PaintablePlanes struct {
    planes map[int]*PaintablePlane  // key: layer index
}

type PaintablePlane struct {
    layer int
    boxes []*PaintableBox
}
```

**构建过程**:
```
PaintableBox 树
    ↓
buildPaintablePlanes()
    ↓
按 Layer 字段分组
    ↓
PaintablePlanes
    ├─ planes[0] (base):   [...]
    ├─ planes[1] (overlay): [...]
    ├─ planes[2] (modal):   [...]
    ├─ planes[3] (tooltip): [...]
    └─ planes[4] (inspector): [...]
```

**render 顺序** (低 → 高):
```
LayerBase (0) → LayerOverlay (1) → LayerModal (2) → LayerTooltip (3) → LayerInspector (4)
```

**HitMap 构建顺序** (高 → 低):
```
LayerInspector (4) → LayerTooltip (3) → LayerModal (2) → LayerOverlay (1) → LayerBase (0)
```

---

## 🔗 Layer 传播路径

### VNode → Fiber 创建

**位置**: `runtime/ui/fiber_util.go:197`

```go
func NewFiber(...) *Fiber {
    return &Fiber{
        Type:    vnodeType,
        Tag:     tag,
        Props:   props,
        NodeID:  generateNodeID(),
        Layer:   vnode.GetLayer(),  // ← 初始化 Layer
        // ...
    }
}
```

### FiberVNode 的 Layer 访问

**位置**: `runtime/ui/fiber_vnode.go`

```go
type FiberVNode struct {
    fiber *Fiber
}

func (f *FiberVNode) GetLayer() Layer {
    if f.fiber == nil {
        return LayerBase
    }
    return f.fiber.Layer  // ← 从 Fiber 读取
}

func (f *FiberVNode) SetLayer(l Layer) VNode {
    if f.fiber != nil {
        f.fiber.Layer = l  // ← 直接修改 Fiber.Layer
    }
    return f
}
```

---

## 🎨 组件 Layer API

### Tooltip 组件

**位置**: `ui/components/tooltip/vnode.go`

```go
type VNode struct {
    content   rtui.VNode
    text      string
    layer     rtui.Layer  // ← 持久化字段
}

func (t *VNode) GetLayer() rtui.Layer {
    return t.layer
}

func (t *VNode) SetLayer(l rtui.Layer) rtui.VNode {
    t.layer = l
    return t
}

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
    layer    rtui.Layer

// Layer() 通用方法
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
    TooltipLayer().
    Build()

tooltip.NewBuilder(content, "Important").
    ModalLayer().
    Build()
```

### Toast 组件 (类似)

```go
// ToastVNode 同样支持 GetLayer()/SetLayer()
type ToastVNode struct {
    text      string
    toastType ToastType
    layer     rtui.Layer  // ← 持久化
}

// ToastBuilder 同样支持 SetRenderLayer() 等
type ToastBuilder struct {
    text      string
    toastType ToastType
    layer     rtui.Layer  // ← 持久化
}

func (b *ToastBuilder) SetRenderLayer() *ToastBuilder {
    b.layer = rtui.LayerOverlay
    return b
}
```

---

## 🚀 使用示例

### 示例 1: Modal 对话框

```go
func (s *AppState) App() ui.VNode {
    content := ui.HStack(
        ui.NewText("Welcome"),
        UI.NewButton("Open Modal", func() {
            s.showModal = true
        }),
    )

    if s.showModal {
        modalContent := modal.New(
            ui.NewText("Are you sure?"),
            modal.WithActions(...),
        )
        modalContent.SetLayer(ui.LayerModal)

        return ui.VStack(content, modalContent)
    }

    return content
}
```

### 示例 2: 多个 Tooltip 不同层级

```go
func MultiTooltipExample() ui.VNode {
    tooltip1 := tooltip.NewBuilder(btn1, "Normal").
        TooltipLayer().Build()

    tooltip2 := tooltip.NewBuilder(btn2, "IMPORTANT!").
        ModalLayer().Build()

    tooltip3 := tooltip.NewBuilder(btn3, "DEBUG: info").
        InspectorLayer().Build()

    return ui.VStack(tooltip1, tooltip2, tooltip3)
}
```

### 示例 3: 动态 Layer

```go
type DynamicLayer struct {
    layer rtui.Layer
}

func (d *DynamicLayer) Render() ui.VNode {
    content := ui.NewText("Content")
    content.SetLayer(d.layer)

    return ui.HStack(
        content,
        ui.NewButton("Change", func() {
            d.layer = rtui.LayerOverlay
        }),
    )
}
```

---

## 🐛 调试与验证

### 启用日志

```bash
export TUI_LAYER_DEBUG=true
export TUI_DEBUG_RENDER=true
export TUI_DEBUG_HITMAP=true
```

### 预期日志输出

```
[PipelineRenderer] Buffer size: 120x40
[PipelineRenderer] hasLayers=true
[PipelineRenderer] Using RenderLayers for multi-layer rendering
[RenderingPipeline] RenderLayers started
[RenderingPipeline] Layout complete, root size=120x40
[RenderingPipeline] Apply layer transforms
[RenderingPipeline] PaintablePlanes: 5 planes, 42 boxes
[PaintEngine] Painting 5 layers (low → high)
[RenderingPipeline] RenderLayers complete
```

### 手动验证 Layer 检测

```go
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

---

## ❓ 常见问题

### Q1: 为什么我的 Tooltip 不显示在最上层？

**检查清单**:
1. Layer 是否为 `LayerTooltip`?
2. 是否有 `LayerInspector` 元素遮挡?
3. 检查日志: `[hasLayerNodes] Found layer node`

### Q2: Modal 不居中？

**确保**:
1. Modal 在 `LayerModal` 层
2. 容器有明确尺寸（非 Infinity）
3. Modal 内容不自适应

### Q3: 如何动态改变 Layer?

```go
fiberVNode.SetLayer(rtui.LayerOverlay)
// Fiber.Layer 直接修改，下一帧渲染生效
```

### Q4: 为什么 hasLayerNodesFromFiber 检测失败?

可能原因:
1. Fiber 为 nil
2. 所有节点都是 LayerBase
3. 检查递归是否完整

---

## 📊 性能考虑

| 操作 | 开销 |
|------|------|
| `hasLayerNodes()` | < 1ms |
| `RenderLayers()` vs `Render()` | +5-15% |
| PaintablePlanes 构建 | +2-5% |
| HitMap 构建 | +3-8% |

**优化建议**:
- 避免过度使用 Layer
- 限制 Modal 数量
- 利用 Fiber 缓存
- 生产环境禁用日志

---

## 📚 相关文档

- `docs/layer/FIBER_FIRST_LAYER_SYSTEM.md` - Fiber-First Layer 系统完整技术细节
- `ui/components/tooltip/layer_demo.go` - Tooltip Layer 使用示例
- `internal/render/` - 源码实现

---

**文档版本**: 2.0
**最后更新**: 2026-02-23
**状态**: ✅ Fiber-First 架构
