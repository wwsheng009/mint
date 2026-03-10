# Fiber-first 渲染流程分析

## 一、完整渲染流程

```
NewDeclarativeNodeFromFuncWithFiber(renderFn, fwApp)
    │
    ├── 1. 创建 Fiber Reconciler
    │       └── newFiberReconciler(fwApp, fn)
    │               ├── reconciler.NewReconciler(fwApp, fn)
    │               ├── SetFocusManager(focusMgr)
    │               └── SetRenderer(renderer)
    │
    ├── 2. 创建 FiberFocusManager
    │       └── rtui.NewFiberFocusManager()
    │
    ├── 3. 创建 PipelineRendererAdapter
    │       └── NewPipelineRendererAdapter()
    │
    └── 4. 初始化 Fiber-first 组件 (MINT_FIBER_FIRST=true)
            └── initFiberFirstPipeline()
                    ├── renderMode = RenderModeFiberFirst
                    ├── newLayoutEngine = NewNewLayoutEngineAdapter()
                    └── paintEngine = NewPaintEngine()

================================================================================

Paint() 被调用时:
    │
    └── DeclarativeNode.Paint(ctx, buf)
            │
            ├── 检查 renderMode
            │
            └── fiberFirstPaint(ctx, buf) (RenderModeFiberFirst)
                    │
                    ├── Phase 1: Reconcile (VNode → Fiber, VNode 丢弃)
                    │       │
                    │       └── n.reconciler.Render(ctx, nullBuf, renderFn)
                    │               │
                    │               └── Fiber 树构建/更新，VNode 被丢弃
                    │                       │
                    │                       └── getFiberRoot() → Fiber 根节点
                    │
                    ├── Phase 2: Layout (Fiber → LayoutBox)
                    │       │
                    │       ├── constraints := runtime.BoxConstraints{
                    │       │       MinWidth: 0,
                    │       │       MaxWidth: ctx.AvailableWidth,
                    │       │       MinHeight: 0,
                    │       │       MaxHeight: ctx.AvailableHeight,
                    │       │   }
                    │       │
                    │       └── n.newLayoutEngine.LayoutFiber(fiberRoot, constraints)
                    │               │
                    │               ├── NewFiberToNodeAdapterPure(fiberRoot)
                    │               │       └── Fiber → layout.Node 接口
                    │               │               │
                    │               │               ├── ID() → fiber.NodeIDStr()
                    │               │               ├── Measure(constraints) → fiber.Instance.Measure()
                    │               │               ├── Children() → fiber.Child → adapter
                    │               │               └── FlexStyle() → fiber.LayoutFlex...
                    │               │
                    │               └── layout.Engine.Layout(node, constraints)
                    │                       │
                    │                       ├── 递归 layoutNodeWithDepth()
                    │                       ├── fiber.Instance.Measure(constraints)
                    │                       └── 返回 layout.LayoutResult
                    │                               └── result.Root = *layout.LayoutBox
                    │
                    └── Phase 3: Paint (LayoutBox → PaintableBox → Buffer)
                            │
                            ├── NewFiberToPaintableConverter(fiberRoot)
                            │       └── buildFiberMap(fiberRoot)
                            │               ├── 索引所有 Fiber (DiffKey, Key, NodeID)
                            │               └── fiberMap[string]*Fiber
                            │
                            ├── converter.ConvertToLayout(layoutBoxRoot)
                            │       │
                            │       ├── 递归 Convert(layoutBox, parent)
                            │       │       │
                            │       │       ├── 创建 PaintableBox{
                            │       │       │       X, Y, Width, Height,
                            │       │       │       Layer, ZIndex,
                            │       │       │       Children: []*PaintableBox,
                            │       │       │   }
                            │       │       │
                            │       │       ├── findFiber(layoutBox.ID) → fiber
                            │       │       │
                            │       │       └── fillFromFiber(pbox, fiber)
                            │       │               │
                            │       │               ├── pbox.NodeID = fiber.NodeID
                            │       │               ├── pbox.Node = NewFiberPaintableNode(fiber)
                            │       │               ├── pbox.BorderStyle = ...
                            │       │               └── pbox.BorderColor = ...
                            │       │
                            │       └── 返回 paint.PaintableLayout
                            │               └── root = *paint.PaintableBox
                            │
                            └── n.paintEngine.PaintLayout(paintableLayout, buf)
                                    │
                                    ├── paintBox(pbox, buffer)
                                    │       │
                                    │       ├── fiber.Instance.SetBounds(x, y, w, h)
                                    │       │
                                    │       ├── pbox.Node.Paint(x, y) → []paint.DrawCmd
                                    │       │       │
                                    │       │       └── fiber.Instance.Paint(x, y)
                                    │       │               │
                                    │       │               └── 返回绘制命令
                                    │       │
                                    │       └── buffer.SetString/Draw(cmd)
                                    │
                                    └── 递归 paintBoxChildren(pbox, buffer)
```

## 二、数据类型流转

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Fiber-first 数据流                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Phase 1: Reconcile                                                         │
│  ┌─────────┐   reconciler.Render()      ┌─────────┐                        │
│  │ VNode   │                         ╱│  Fiber  │◀── 持久化，跨渲染保持      │
│  │ (临时)  │  ╻                     ╱ └─────────┘                        │
│  └─────────┘  │                    │                                        │
│               │ CreateFiber()      │    InstanceManager                    │
│               │                    │    InstanceFactory                    │
│               │                    ╱                                        │
│            ╱ ╲                   ╱                                         │
│           ╱   ╲                 ╱                                          │
│        Fiber树构建              VNode被丢弃                                  │
│           ▼                                                               │
│  Phase 2: Layout                                                           │
│                             ┌──────────────┐                                │
│                             │   Fiber      │                                │
│                             │  ┌────────┐  │                                │
│                             │  │NodeID  │  │                                │
│                             │  │Tag     │  │                                │
│                             │  │Props   │  │                                │
│                             │  │Style   │  │                                │
│                             │  │Instance│──┼──▶ ComponentInstance          │
│                             │  └────────┘  │       │                        │
│                             └──────┬───────┘       │                        │
│                                    │               │                        │
│                    NewFiberToNodeAdapterPure       │                        │
│                                    │               │                        │
│                                    ▼               │                        │
│                             ┌──────────────┐       │                        │
│                             │ layout.Node  │ ◀───┐ │                        │
│                             │  (interface) │     │ │ 测量                   │
│                             │   Implement  │─────┘                        │
│                             └──────┬───────┘                                │
│                                    │                                        │
│                         layout.Engine                                        │
│                         ┌───────┴───────┐                                  │
│                         │   Layout()    │                                  │
│                         │ 递归测量和布局 │                                  │
│                         │ Apply() 设置   │                                  │
│                         │ 位置和尺寸     │                                  │
│                         └───────┬───────┘                                  │
│                                    │                                        │
│                                    ▼                                        │
│                             ┌──────────────┐                                │
│                             │ layout.Box   │ ◀── 纯布局结果               │
│                             │  ┌────────┐  │                                │
│                             │  │X, Y    │  │                                │
│                             │  │Width   │  │                                │
│                             │  │Height  │  │                                │
│                             │  │Layer   │  │                                │
│                             │  │ZIndex  │  │                                │
│                             │  │Border  │  │                                │
│                             │  │Children│  │                                │
│                             │  └────────┘  │                                │
│                             │  ❌ 无 Fiber │                                │
│                             │  ❌ 无 VNode │                                │
│                             └──────┬───────┘                                │
│                                    │                                        │
│  Phase 3: Paint                    │                                        │
│                                    │                                        │
│                    FiberToPaintableConverter                                 │
│                    buildFiberMap(fiberRoot)                                 │
│                    索引: {DiffKey, Key, NodeID} → Fiber                     │
│                                    │                                        │
│                                    ▼                                        │
│                             ┌──────────────┐                                │
│                             │PaintableBox  │ ◀── 纯绘制数据               │
│                             │  ┌────────┐  │                                │
│                             │  │Node    │──┼──▶ FiberPaintableNode        │
│                             │  │        │  │       │                       │
│                             │  │X, Y    │  │       ├── Tag(), Style()      │
│                             │  │Width   │  │       ├── SetBounds()        │
│                             │  │Height  │  │       └── Paint()            │
│                             │  │NodeID  │  │               │               │
│                             │  │Layer   │  │               ▼               │
│                             │  │ZIndex  │  │        Instance.Paint()       │
│                             │  │Children│  │               │               │
│                             │  │Border  │  │               ▼               │
│                             │  └────────┘  │        []DrawCmd             │
│                             └──────────────┘                                │
│                                    │                                        │
│                                    ▼                                        │
│                             PaintEngine                                     │
│                             paintBox()                                       │
│                             │                                               │
│                             ├── SetBounds()                                 │
│                             ├── Paint() → DrawCmd                           │
│                             └── buffer.SetString/Draw()                     │
│                                    │                                        │
│                                    ▼                                        │
│                             ┌──────────────┐                                │
│                             │   Buffer     │ ◀── 最终输出                 │
│                             │ Cells[y][x] │                                │
│                             └──────────────┘                                │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 三、关键适配器

| 适配器 | 位置 | 输入 | 输出 | 状态 |
|--------|------|------|------|------|
| **FiberToNodeAdapter** | `internal/render/fiber_adapter.go` | `*Fiber` | `layout.Node` | ✅ 完整 |
| **FiberPaintableNode** | `internal/render/converter.go` | `*Fiber` | `paint.PaintableNode` | ✅ 完整 |
| **FiberToPaintableConverter** | `internal/render/converter.go` | `*LayoutBox` + Fiber 索引 | `*PaintableBox` | ✅ 完整 |
| **NewLayoutEngineAdapter** | `internal/render/layout_switcher.go` | `*Fiber` + Constraints | `LayoutResult` | ✅ 完整 |
| **PipelineRendererAdapter** | `internal/render/pipeline_renderer.go` | VNode + Renderer | — | ✅ 完整 |
| **VNodeToNodeAdapter** | `internal/render/fiber_adapter.go` | VNode | `layout.Node` | ⚠️ Deprecated |
| **PaintEngine** | `internal/render/paint_engine.go` | `*PaintableLayout` | Buffer 绘制 | ✅ 新API |

## 四、各层接口实现

### 1. layout.Node (FiberToNodeAdapter 实现)

```go
type FiberToNodeAdapter struct {
    fiber *reconciler.Fiber
}

// 布局测量
func (a *FiberToNodeAdapter) Measure(layout.Constraints) layout.Size {
    // 1. 优先使用 Fiber.Instance (ComponentInstance)
    if measurable, ok := a.fiber.Instance.(Measurable); ok {
        return measurable.Measure(constraints)
    }

    // 2. 回退到 Style 尺寸
    if a.fiber.Style.Width > 0 || a.fiber.Style.Height > 0 {
        return layout.Size{...}
    }

    return layout.Size{Width: 0, Height: 0}
}

// 子节点
func (a *FiberToNodeAdapter) Children() []layout.Node {
    children := []layout.Node{}
    for child := a.fiber.Child; child != nil; child = child.Sibling {
        children = append(children, NewFiberToNodeAdapterPure(child))
    }
    return children
}

// Flex 布局属性
func (a *FiberToNodeAdapter) FlexStyle() layout.FlexStyle {
    return layout.FlexStyle{
        Direction: a.fiber.LayoutDirection,
        Align:     a.fiber.LayoutAlign,
        CrossAlign: a.fiber.LayoutCrossAlign,
        Gap:       a.fiber.LayoutGap,
        Padding:   a.fiber.LayoutPadding,
        Flex:      a.fiber.LayoutFlex,
    }
}

// 位置和尺寸
func (a *FiberToNodeAdapter) ID() string
func (a *FiberToNodeAdapter) GetPosition() (int, int)
func (a *FiberToNodeAdapter) SetPosition(x, y int)
func (a *FiberToNodeAdapter) GetSize() (int, int)
func (a *FiberToNodeAdapter) SetSize(w, h int)
```

### 2. paint.PaintableNode (FiberPaintableNode 实现)

```go
type FiberPaintableNode struct {
    fiber *reconciler.Fiber
}

// 绘制命令
func (n *FiberPaintableNode) Paint(x, y int) []paint.DrawCmd {
    if n.fiber == nil {
        return nil
    }

    // Primary Path: Fiber.Instance.Paint() (Fiber-first)
    if n.fiber.Instance != nil {
        if inst, ok := n.fiber.Instance.(PaintableInstance); ok {
            return inst.Paint(x, y)
        }
    }

    // Fallback: PaintRegistry (简单无状态组件)
    if fn := rtui.GetPaint(n.fiber.Tag); fn != nil {
        return fn(n.fiber.Props, n.fiber.Style, x, y)
    }

    return nil
}

// 其他接口
func (n *FiberPaintableNode) Tag() string
func (n *FiberPaintableNode) Style() style.Style
func (n *FiberPaintableNode) SetStyle(s style.Style)
func (n *FiberPaintableNode) TextContent() string
func (n *FiberPaintableNode) NodeType() paint.NodeType

// 布局边界设置
func (n *FiberPaintableNode) SetBounds(x, y, w, h int) {
    if n.fiber == nil || n.fiber.Instance == nil {
        return
    }
    if boundsSetter, ok := n.fiber.Instance.(SetBounds); ok {
        boundsSetter.SetBounds(x, y, w, h)
    }
}
```

### 3. ComponentInstance (由各组件实现)

```go
type ButtonInstance struct { ... }

// 测量
func (b *ButtonInstance) Measure(layout.Constraints) layout.Size {
    // 计算按钮尺寸：内容 + padding + border
    textWidth := len(b.text)
    width := textWidth + b.paddingLeft + b.paddingRight + 2 // 左右边框
    height := 1 + b.paddingTop + b.paddingBottom + 2 // 上下边框
    return layout.Size{Width: width, Height: height}
}

// 绘制
func (b *ButtonInstance) Paint(x, y int) []paint.DrawCmd {
    return []paint.DrawCmd{
        border.Draw(x, y, ...),    // 边框
        text.Draw(x+1, y+1, ...),  // 文本
    }
}

// 设置边界
func (b *ButtonInstance) SetBounds(x, y, w, h int) {
    b.x, b.y = x, y
    b.width, b.height = w, h
}
```

## 五、渲染模式对比

| 特性 | Legacy 渲染 | Fiber-first 渲染 |
|------|------------|------------------|
| **环境变量** | `MINT_FIBER_FIRST=false` | `MINT_FIBER_FIRST=true` |
| **渲染函数** | `legacyPaint()` | `fiberFirstPaint()` |
| **VNode 生命周期** | 保留到 Paint 阶段 | Reconcile 后立即丢弃 |
| **布局引擎** | `compute.Engine` | `layout.Engine` |
| **Layout 结果** | `compute.ComputedBox` | `layout.LayoutBox` |
| **Paint 引擎** | `Paint(ComputedLayout)` | `PaintLayout(PaintableLayout)` |
| **ComponentInstance** | 不可用 | ✅ 可用 |
| **事件处理** | VNode 回调 | Fiber.Instance 事件 |

## 六、入口点

| 入口点 | 函数 | 位置 | 渲染模式 |
|--------|------|------|----------|
| `ui.Run()` | `Run(app)` | `ui/app.go` | Legacy (需设置环境变量) |
| `ui.RunTest()` | `RunTest(app)` | `ui/test.go` | Fiber-first |
| `NewDeclarativeNodeFromFuncWithFiber()` | `declarative_node.go` | `internal/render` | Fiber-first |

### 使用示例

```go
// 方式 1: 环境变量控制
os.Setenv("MINT_FIBER_FIRST", "true")
ui.Run(app)

// 方式 2: 直接设置模式
node := render.NewDeclarativeNodeFromFuncWithFiber(app, fwApp)
node.SetRenderMode(render.RenderModeFiberFirst)

// 方式 3: 测试中使用
ta, err := ui.RunTest(app, ui.Width(80), ui.Height(24))
```

## 七、环境变量

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `MINT_USE_FIBER` | `true` | 启用 Fiber 协调器 |
| `MINT_FIBER_FIRST` | `false` | 启用 Fiber-first 渲染路径 |
| `MINT_DEBUG_TEST` | `false` | 启用调试日志 |
| `MINT_WARN_LEGACY` | `false` | 启用废弃警告 |

## 八、已完成的清理

### ✅ 移除的 Legacy 依赖

- ✅ `paint_engine.go`: 移除 `Paint(ComputedLayout)` 及相关方法
- ✅ `paint_engine.go`: 移除 `PaintLayers` 和 `PaintRenderPlanes`
- ✅ `paint_engine.go`: 移除 `parentBackgroundLegacy` 字段
- ✅ `paint_engine.go`: 移除 `compute` 包导入
- ✅ `declarative_node.go`: 移除 `layerMgr` 字段和 `SetLayerManager()`
- ✅ `pipeline_renderer.go`: 移除 `layerMgr` 和 `layerEvents` 字段
- ✅ `rendering_pipeline.go`: 移除 `layerMgr` 字段和 `GetLayerManager()`

### ✅ 保留的新 API

- ✅ `PaintEngine.PaintLayout(*PaintableLayout)`
- ✅ `PaintEngine.PaintPaintableLayouts(PaintableLayouts)`
- ✅ `PaintEngine.PaintPaintablePlanes(*PaintablePlanes)`

### ✅ 兼容性支持

- ✅ `fiber_adapter.go`: 同时支持 `layout.LayoutBox` 和 `compute.ComputedBox`
- ✅ `fiber_adapter.go`: 优先新架构，回退旧架构

## 九、架构设计原则

1. **LayoutBox** - 纯布局数据，无 Fiber/VNode 依赖
2. **PaintableBox** - 纯绘制数据，通过 FiberPaintableNode 访问 Fiber
3. **解耦** - Layout 和 Paint 通过适配器访问 Fiber，不直接依赖
4. **Fiber-first** - VNode 仅用于创建 Fiber，Reconcile 后立即丢弃
5. **持久化** - Fiber 和 ComponentInstance 跨渲染保持状态

## 十、测试验证

✅ `TestDeclarativeNode_Paint_Fiber` - Fiber 渲染测试
✅ `TestFullRenderingPipeline_NewButton` - 完整渲染管线测试
✅ `TestPaintEngine_PaintLayout` - PaintEngine 测试
✅ `text_demo` - 实际应用运行
