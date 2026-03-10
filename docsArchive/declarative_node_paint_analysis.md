# DeclarativeNode.Paint() 方法调用链路分析

本文档分析 `internal/render/declarative_node.go` 中 `Paint` 方法的完整调用处理链路与逻辑。

## 整体架构图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              App.render()                                    │
│                           (framework/app.go:1476)                           │
│                                                                              │
│   ctx := component.PaintContext{                                             │
│       AvailableWidth:  layoutWidth,                                          │
│       AvailableHeight: layoutHeight,                                         │
│   }                                                                          │
│   buf := a.renderer.GetBackBuffer()                                         │
│                                                                              │
│   paintable.Paint(ctx, buf)  ──────────────────────────────────────────────┐│
└─────────────────────────────────────────────────────────────────────────────┘│
                                                                               │
┌──────────────────────────────────────────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                    DeclarativeNode.Paint()                                   │
│              (internal/render/declarative_node.go:240)                       │
├─────────────────────────────────────────────────────────────────────────────┤
│ Phase 1: 获取 VNode 树                                                       │
│ ┌─────────────────────────────────────────────────────────────────────────┐ │
│ │ if useFiber && reconciler != nil:                                       │ │
│ │     n.root = n.renderWithFiberContext()  ────────────────────────────┐  │ │
│ │ else:                                                               │  │ │
│ │     n.root = n.nonFiberRender()                                     │  │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                              │
│ Phase 2: 应用焦点状态                                                        │
│ ┌─────────────────────────────────────────────────────────────────────────┐ │
│ │ n.applyFocusState()                                                    │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                              │
│ Phase 3: 渲染到 Buffer (UNIFIED RENDERING)                                  │
│ ┌─────────────────────────────────────────────────────────────────────────┐ │
│ │ if PipelineRendererAdapter:                                             │ │
│ │     pipeline.RenderWithConstraints(n.root, ctx.AvailableWidth,          │ │
│ │                                   ctx.AvailableHeight, buf)              │ │
│ │ else:                                                                   │ │
│ │     n.PaintVNode(n.root, ...)  // Legacy fallback                       │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 详细调用链

```
App.render()
    │
    ├── paintable.Paint(ctx, buf)
    │       │
    │       ▼
    │   DeclarativeNode.Paint(ctx, buf)  [declarative_node.go:240]
    │       │
    │       ├─── Phase 1: 获取 VNode 树
    │       │    │
    │       │    ├── [Fiber 模式]
    │       │    │   renderWithFiberContext()  [declarative_node.go:353]
    │       │    │       │
    │       │    │       ├── n.reconciler.Render(ctx, buf, renderFn)
    │       │    │       │       │
    │       │    │       │       └── fiberReconcilerAdapter.Render()
    │       │    │       │               │
    │       │    │       │               └── reconciler.Reconciler.Render()
    │       │    │       │                       (执行 Fiber 协调，调用 hooks)
    │       │    │       │
    │       │    │       └── 捕获 renderFn() 返回的 VNode 树
    │       │    │
    │       │    └── [非 Fiber 模式]
    │       │        nonFiberRender()  [declarative_node.go:395]
    │       │            │
    │       │            ├── n.renderFn()  // 调用组件渲染函数
    │       │            └── expandComponents(n.root)  // 展开 ComponentVNode
    │       │
    │       ├─── Phase 2: 应用焦点状态
    │       │    │
    │       │    └── applyFocusState()  [declarative_node.go:424]
    │       │            │
    │       │            ├── Fiber 模式：同步 FocusManager 到 App
    │       │            └── 非 Fiber 模式：收集 Focusable 节点，设置焦点
    │       │
    │       └─── Phase 3: 渲染到 Buffer
    │            │
    │            ├── [PipelineRendererAdapter 路径] ✅ 推荐
    │            │   PipelineRendererAdapter.GetPipeline()
    │            │       │
    │            │       └── PipelineRenderer.GetPipeline()
    │            │               │
    │            │               └── RenderingPipeline.RenderWithConstraints()
    │            │                       [rendering_pipeline.go:229]
    │            │                       │
    │            │                       ├── hasLayers = hasLayerNodes(vnode)
    │            │                       │
    │            │                       ├── [有 Layer 节点]
    │            │                       │   RenderLayers(vnode, fiber, constraints, buf)
    │            │                       │   [rendering_pipeline.go:358]
    │            │                       │       │
    │            │                       │       ├── LayoutSwitcher.Layout()
    │            │                       │       │       │
    │            │                       │       │       ├── runtime/layout 引擎
    │            │                       │       │       └── compute.Engine (旧引擎)
    │            │                       │       │
    │            │                       │       ├── Build RenderPlanes
    │            │                       │       │
    │            │                       │       ├── applyModalCenteringToRenderPlanes()
    │            │                       │       │
    │            │                       │       └── PaintEngine.PaintRenderPlanes()
    │            │                       │
    │            │                       └── [无 Layer 节点]
    │            │                           Render(vnode, fiber, constraints, buf)
    │            │                           [rendering_pipeline.go:70]
    │            │                               │
    │            │                               ├── LayoutSwitcher.Layout()
    │            │                               │       │
    │            │                               │       └── 返回 LayoutResult
    │            │                               │
    │            │                               ├── [newLayoutResultAdapter 路径]
    │            │                               │   convertLayoutBoxToPaintableLayout()
    │            │                               │   FiberToPaintableConverter.ConvertToLayout()
    │            │                               │   PaintEngine.PaintLayout()
    │            │                               │
    │            │                               └── [computeLayoutResultAdapter 路径]
    │            │                                   PaintEngine.Paint(computedLayout, buf)
    │            │
    │            └── [Legacy 路径] (fallback)
    │                PaintVNode(n.root, ctx.Bounds.X, ctx.Bounds.Y, buf)
    │                [declarative_node.go:499]
    │                    │
    │                    ├── Paintable 接口检查 → paintable.Paint(x, y)
    │                    │
    │                    ├── VNode 类型处理：
    │                    │   ├── VNodeText → paintText()
    │                    │   ├── VNodeElement → paintElement()
    │                    │   ├── VNodeFragment → paintChildren()
    │                    │   └── VNodeComponent (已展开)
    │                    │
    │                    ├── 特殊元素处理：
    │                    │   ├── BorderedNode → paintBordered()
    │                    │   └── Table → paintTable()
    │                    │
    │                    └── 递归绘制子节点
```

## 关键组件职责

| 组件 | 文件位置 | 职责 |
|------|----------|------|
| **DeclarativeNode** | `internal/render/declarative_node.go` | 桥接 VNode 树与 Framework，管理渲染模式 |
| **PipelineRendererAdapter** | `internal/render/vnode_renderer.go` | 适配 PipelineRenderer 到 VNodeRenderer 接口 |
| **PipelineRenderer** | `internal/render/pipeline_renderer.go` | 协调 Layout/Paint 两阶段渲染 |
| **RenderingPipeline** | `internal/render/rendering_pipeline.go` | 核心渲染管线，支持多 Layer 渲染 |
| **LayoutSwitcher** | `internal/render/layout_switcher.go` | 切换布局引擎（compute.Engine 或 runtime/layout） |
| **PaintEngine** | `internal/render/paint_engine.go` | 执行绘制命令到 Buffer |
| **fiberReconcilerAdapter** | `internal/render/declarative_node.go:1740` | 适配 Fiber Reconciler |

## 渲染模式对比

### Fiber 模式 (推荐)

```
┌──────────────────────────────────────────────────────────────────┐
│                       Fiber 模式 (推荐)                          │
├──────────────────────────────────────────────────────────────────┤
│ 1. renderWithFiberContext()                                      │
│    - 调用 reconciler.Render() 触发 Fiber 协调                    │
│    - Hooks 在 Fiber context 中执行                               │
│    - 捕获 renderFn() 返回的 VNode 树                             │
│                                                                  │
│ 2. applyFocusState()                                             │
│    - Focus 由 FiberFocusManager 管理                             │
│    - 同步 FocusManager 到 framework.App                          │
│                                                                  │
│ 3. PipelineRendererAdapter.RenderWithConstraints()               │
│    - 使用 Layout/Paint 分离架构                                   │
│    - 支持 Modal centering 等高级布局                              │
└──────────────────────────────────────────────────────────────────┘
```

### 非 Fiber 模式 (Legacy)

```
┌──────────────────────────────────────────────────────────────────┐
│                      非 Fiber 模式 (Legacy)                      │
├──────────────────────────────────────────────────────────────────┤
│ 1. nonFiberRender()                                              │
│    - 直接调用 renderFn()                                         │
│    - 使用 ComponentContext 管理 hooks                            │
│    - expandComponents() 展开 ComponentVNode                      │
│                                                                  │
│ 2. applyFocusState()                                             │
│    - 从 VNode 树收集 Focusable 节点                              │
│    - 通过 VNodeFocusManager 管理焦点                             │
│                                                                  │
│ 3. PaintVNode() 或 PipelineRenderer                              │
│    - Legacy 路径直接递归绘制                                     │
└──────────────────────────────────────────────────────────────────┘
```

## 数据流

```
用户代码 ──> renderFn() ──> VNode 树
                               │
                               ▼
                    Fiber Reconciler (协调 + Diff)
                               │
                               ▼
                       Fiber 树 + 更新的 VNode
                               │
                               ▼
                    Layout 阶段 (compute.Engine 或 runtime/layout)
                               │
                               ▼
                    ComputedBox / LayoutBox 树
                               │
                               ▼
                    Paint 阶段 (PaintEngine)
                               │
                               ▼
                       paint.Buffer (屏幕缓冲)
                               │
                               ▼
                    renderer.Render() (diff + ANSI)
                               │
                               ▼
                          终端输出
```

## 核心源码位置

### Paint 方法入口

```go
// declarative_node.go:240
func (n *DeclarativeNode) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    n.mu.Lock()
    defer n.mu.Unlock()

    // Phase 1: Get the VNode tree
    if n.useFiber && n.reconciler != nil {
        n.root = n.renderWithFiberContext()
    } else {
        n.root = n.nonFiberRender()
    }

    if n.root == nil {
        return
    }

    // Phase 2: Apply focus state
    n.applyFocusState()

    // Phase 3: UNIFIED RENDERING - use PipelineRenderer with constraint-based layout
    if n.renderer != nil {
        if adapter, ok := n.renderer.(*PipelineRendererAdapter); ok {
            pipeline := adapter.GetPipeline()
            if err := pipeline.RenderWithConstraints(n.root, ctx.AvailableWidth, ctx.AvailableHeight, buf); err != nil {
                // Fallback to legacy rendering if pipeline fails
                n.PaintVNode(n.root, ctx.Bounds.X, ctx.Bounds.Y, buf)
            }
        } else {
            n.renderer.Render(n.root, ctx.Bounds.X, ctx.Bounds.Y, buf)
        }
    } else {
        // Fallback to legacy painting
        n.PaintVNode(n.root, ctx.Bounds.X, ctx.Bounds.Y, buf)
    }
}
```

### Fiber 模式渲染

```go
// declarative_node.go:353
func (n *DeclarativeNode) renderWithFiberContext() rtui.VNode {
    if n.renderFn == nil {
        return n.root
    }

    var capturedVNode rtui.VNode
    nullBuf := paint.NewBuffer(1, 1)
    n.reconciler.Render(component.PaintContext{
        Bounds: paint.Rect{X: 0, Y: 0, Width: 1, Height: 1},
    }, nullBuf, func() rtui.VNode {
        vnode := n.renderFn()
        capturedVNode = vnode
        return vnode
    })

    return capturedVNode
}
```

## 相关文档

- [Pipeline Renderer 架构](../pipeline_renderer.md)
- [Layout Engine 对比](../layout_engine_comparison.md)
- [Fiber Reconciler 设计](../../fiber/fiber_reconciler.md)
