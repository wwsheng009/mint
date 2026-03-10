# 渲染架构清理总结

## 概述

本文档总结 Mint 框架渲染架构的清理工作，重点说明如何移除对 `runtime/compute` 包的依赖，统一使用 `layout.Engine + PaintableBox` 架构。

**日期:** 2026-02-21
**目标:** 移除 `runtime/compute` 依赖，统一到 `layout.Engine + PaintableBox` 架构
**状态:** ✅ 主要清理工作完成

---

## 一、清理工作概览

### 1.1 清理的文件

| 文件 | 清理内容 | 状态 |
|-----|---------|------|
| `internal/render/paint_engine.go` | 移除 Legacy Paint 方法、layerMgr、compute 导入 | ✅ 完成 |
| `internal/render/declarative_node.go` | 移除 layerMgr 字段和 SetLayerManager | ✅ 完成 |
| `internal/render/pipeline_renderer.go` | 移除 layerMgr 和 layerEvents 字段 | ✅ 完成 |
| `internal/render/rendering_pipeline.go` | 移除 layerMgr 字段和 GetLayerMgr | ✅ 完成 |
| `internal/render/fiber_adapter.go` | 支持新架构优先，保留 Legacy 回退 | ✅ 完成 |
| `docs/fiber/FIBER_FIRST_RENDER_FLOW.md` | 更新渲染流程文档 | ✅ 完成 |
| `docs/fiber/FIBER_RENDER_PATH_ANALYSIS.md` | 更新渲染路径分析文档 | ✅ 完成 |

### 1.2 保留的兼容性代码

| 代码 | 位置 | 原因 |
|-----|------|------|
| `compute.Engine` | `runtime/compute` | Legacy 渲染路径向后兼容 |
| `compute.ComputedBox` | `runtime/compute` | Legacy 渲染路径向后兼容 |
| `legacyPaint()` | `internal/render/declarative_node.go` | 向后兼容旧代码 |
| `LayoutSwitcher` | `internal/render/layout_switcher.go` | 管理引擎切换 |
| `fiber_adapter.go` 中的 compute 回退 | `internal/render/fiber_adapter.go` | 兼容性回退 |

---

## 二、清理详情

### 2.1 paint_engine.go 清理

#### 已移除的方法

```go
// ✅ 已移除 - Legacy Paint 方法
func (e *PaintEngine) Paint(layout ComputedLayout, buffer *Buffer)
func (e *PaintEngine) paintNode(node ComputedNode, buffer *Buffer, x, y int)
func (e *PaintEngine) paintText(node ComputedNode, buffer *Buffer, x, y int)
func (e *PaintEngine) paintElement(node ComputedNode, buffer *Buffer, x, y int)
func (e *PaintEngine) paintContainerBackground(node ComputedNode, buffer *Buffer, x, y int)
func (e *PaintEngine) paintChildren(node ComputedNode, buffer *Buffer, x, y int)
func (e *PaintEngine) paintBordered(node ComputedNode, buffer *Buffer, x, y int)
func (e *PaintEngine) paintTable(table ComputedNode, buffer *Buffer, x, y int)
func (e *PaintEngine) paintModalBackdrop(backdrop ComputedNode, buffer *Buffer, x, y int)
func (e *PaintEngine) PaintLayers(layers []ComputedLayer, buffer *Buffer)
func (e *PaintEngine) PaintRenderPlanes(planes []ComputedLayer, buffer *Buffer)

// ✅ 已移除 - 字段
parentBackgroundLegacy style.Color
```

#### 已移除的导入

```go
// ✅ 已移除
"mint/internal/runtime/compute"
"mint/internal/runtime/layer"
```

#### 保留的新 API

```go
// ✅ 保留 - 新架构 API
func (e *PaintEngine) PaintLayout(paintLayout *PaintableLayout, buffer *Buffer)
func (e *PaintEngine) PaintPaintableLayouts(layouts PaintableLayouts, buffer *Buffer)
func (e *PaintEngine) PaintPaintablePlanes(planes *PaintablePlanes, buffer *Buffer)
```

### 2.2 fiber_adapter.go 新架构支持

#### 新架构优先

```go
func (a *FiberToNodeAdapter) GetPosition() (int, int) {
    // ✅ 优先检查新架构 layout.LayoutBox
    if a.fiber != nil && a.fiber.Layout != nil {
        if layoutBox, ok := a.fiber.Layout.(*layout.LayoutBox); ok {
            if layoutBox != nil {
                return layoutBox.X, layoutBox.Y
            }
        }
    }

    // ⚠️ 兼容性回退到旧架构 compute.ComputedBox
    if a.fiber != nil && a.fiber.Computed != nil {
        if computedBox, ok := a.fiber.Computed.(*compute.ComputedBox); ok {
            if computedBox != nil {
                return computedBox.X, computedBox.Y
            }
        }
    }

    if boundsSetter, ok := a.fiber.Instance.(BoundsGetter); ok {
        return boundsSetter.GetBounds()
    }
    return 0, 0
}
```

同样的模式应用于：
- `SetPosition(x, y int)`
- `GetSize() (int, int)`
- `SetSize(w, h int)`

---

## 三、新旧架构对比

### 3.1 渲染管线对比

| 阶段 | Legacy 架构 | Fiber-first 架构 |
|-----|------------|------------------|
| **协调** | VNode 树保留 | Fiber 树 + VNode 丢弃 |
| **布局引擎** | `compute.Engine` | `layout.Engine` |
| **Layout 结果** | `compute.ComputedBox` | `layout.LayoutBox` |
| **Paint 引擎** | `Paint(ComputedLayout)` | `PaintLayout(PaintableLayout)` |
| **组件实例** | ❌ 无 | ✅ `ComponentInstance` |
| **回调钩子** | ❌ 无 | ✅ Hook 系统 |

### 3.2 数据类型对比

| 数据类型 | Legacy 类型 | 新架构类型 | 说明 |
|---------|-----------|-----------|------|
| 布局节点 | - | `layout.Node` | 统一接口 |
| 布局结果 | `compute.ComputedBox` | `layout.LayoutBox` | 纯布局数据 |
| 可绘制节点 | - | `paint.PaintableNode` | 统一接口 |
| 可绘制盒子 | `compute.ComputedLayout` | `paint.PaintableBox` | 纯绘制数据 |
| 组件实例 | ❌ 不支持 | `ComponentInstance` | 持久化实体 |

### 3.3 方法签名对比

#### Legacy Paint (已移除)

```go
// ❌ 已移除
func Paint(ComputedLayout, *Buffer)
func PaintLayers([]ComputedLayer, *Buffer)
type ComputedLayout interface {
    Root() ComputedNode
    Nodes() []ComputedNode
    Hitmap() computed.HitMap
}
```

#### New Paint (保留)

```go
// ✅ 保留
func PaintLayout(*PaintableLayout, *Buffer)
func PaintPaintableLayouts(PaintableLayouts, *Buffer)
type PaintableLayout struct {
    Root   *PaintableBox
    Planes *PaintablePlanes
}
type PaintableBox struct {
    X, Y, Width, Height int
    Node                PaintableNode
    Layer               int
    ZIndex              int
    Children            []*PaintableBox
}
```

---

## 四、Fiber-first 渲染流程

### 4.1 完整流程

```
┌─────────────────────────────────────────────────────────────────────┐
│                    Fiber-first 渲染流程                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  1. 初始化                                                          │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ NewDeclarativeNodeFromFuncWithFiber(renderFn, fwApp)       │   │
│  │   • fiberReconciler = reconciler.NewReconciler(fwApp, fn)  │   │
│  │   • fiberFocusManager = NewFiberFocusManager()              │   │
│  │   • renderer = NewPipelineRendererAdapter()                 │   │
│  │   • initFiberFirstPipeline()  (MINT_FIBER_FIRST=true)      │   │
│  │       ├─ renderMode = RenderModeFiberFirst                  │   │
│  │       ├─ newLayoutEngine = NewNewLayoutEngineAdapter()      │   │
│  │       └─ paintEngine = NewPaintEngine()                     │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                      │
│  2. Paint() 调用                                                   │
│                              ▼                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ DeclarativeNode.Paint(ctx, buf)                            │   │
│  │   └─ fiberFirstPaint(ctx, buf)                             │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                      │
│  3. Phase 1: Reconcile (VNode → Fiber, VNode 丢弃)                 │
│                              ▼                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ reconciler.Render(ctx, nullBuf, renderFn)                   │   │
│  │   • beginWork(): 创建/复用 Fiber 节点                       │   │
│  │   • InstanceFactory: 创建 ComponentInstance                 │   │
│  │   • completeWork(): 完成 Fiber 树构建                        │   │
│  │   • fiberRoot := getFiberRoot()                             │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                      │
│  4. Phase 2: Layout (Fiber → LayoutBox)                           │
│                              ▼                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ newLayoutEngine.LayoutFiber(fiberRoot, constraints)         │   │
│  │   • node := NewFiberToNodeAdapterPure(fiberRoot)            │   │
│  │   • result := layout.Engine.Layout(node, constraints)       │   │
│  │   • layoutBoxRoot := result.Root (*layout.LayoutBox)        │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                      │
│  5. Phase 3: Paint (LayoutBox → PaintableBox → Buffer)           │
│                              ▼                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ converter := NewFiberToPaintableConverter(fiberRoot)        │   │
│  │ paintableLayout := converter.ConvertToLayout(layoutBoxRoot) │   │
│  │ paintEngine.PaintLayout(paintableLayout, buf)               │   │
│  │   output: buffer                                            │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 4.2 适配器链

```
Fiber → FiberToNodeAdapterPure → layout.Node → layout.Engine
                                              ↓
                                        layout.LayoutBox
                                              ↓
                                 FiberToPaintableConverter
                                              ↓
                                          paint.PaintableBox
                                              ↓
                                         PaintEngine
                                              ↓
                                            Buffer
```

---

## 五、关键组件

### 5.1 核心接口

#### layout.Node

```go
type Node interface {
    ID() string
    Measure(Constraints) Size
    Children() []Node
    FlexStyle() FlexStyle
    GetPosition() (int, int)
    SetPosition(x, y int)
    GetSize() (int, int)
    SetSize(w, h int)
}
```

#### paint.PaintableNode

```go
type PaintableNode interface {
    Tag() string
    Style() style.Style
    SetStyle(s style.Style)
    TextContent() string
    NodeType() NodeType
    Paint(x, y int) []DrawCmd
}
```

#### ComponentInstance

```go
type ComponentInstance interface {
    Measure(layout.Constraints) layout.Size
    Paint(x, y int) []paint.DrawCmd
    SetBounds(x, y, w, h int)
    HandleEvent(Event) bool // 可选
}
```

### 5.2 核心适配器

| 适配器 | 实现 | 文件 |
|--------|------|------|
| `FiberToNodeAdapter` | 实现 `layout.Node` | `internal/render/fiber_adapter.go` |
| `FiberPaintableNode` | 实现 `paint.PaintableNode` | `internal/render/converter.go` |
| `FiberToPaintableConverter` | 转换 `LayoutBox → PaintableBox` | `internal/render/converter.go` |
| `NewLayoutEngineAdapter` | 适配 `layout.Engine` | `internal/render/layout_switcher.go` |

---

## 六、测试验证

### 6.1 测试状态

| 测试 | 描述 | 状态 |
|-----|------|------|
| `TestDeclarativeNode_Paint_Fiber` | Fiber 渲染基本测试 | ✅ 通过 |
| `TestFullRenderingPipeline_NewButton` | 完整渲染管线测试 | ✅ 通过 |
| `TestPaintEngine_PaintLayout` | PaintEngine.PaintLayout 测试 | ✅ 通过 |
| `text_demo` | 实际应用运行 | ✅ 通过 |

### 6.2 编译状态

```bash
✅ go build ./...  # 所有包编译通过
```

---

## 七、环境变量

| 环境变量 | 默认值 | 推荐值 | 说明 |
|---------|--------|--------|------|
| `MINT_USE_FIBER` | `true` | `true` | 启用 Fiber 协调器 |
| `MINT_FIBER_FIRST` | `false` | `true` | 启用 Fiber-first 渲染路径 |
| `MINT_DEBUG_TEST` | `false` | - | 调试日志 |
| `MINT_WARN_LEGACY` | `false` | - | 废弃警告 |

### 推荐配置

```go
// 新代码推荐
os.Setenv("MINT_USE_FIBER", "true")
os.Setenv("MINT_FIBER_FIRST", "true")

// 入口点
func main() {
    app := MyApp()
    ui.Run(app)
}

// 测试入口
func TestMyComponent(t *testing.T) {
    app := MyTestApp()
    ta, _ := ui.RunTest(app, ui.Width(80), ui.Height(24))
    // ... 测试逻辑
}
```

---

## 八、遗留代码说明

### 8.1 计划保留的代码

| 代码 | 原因 | 建议 |
|-----|------|------|
| `compute.Engine` | Legacy 路径兼容 | 标记 deprecated |
| `compute.ComputedBox` | Legacy 路径兼容 | 标记 deprecated |
| `legacyPaint()` | 向后兼容旧应用 | 标记 deprecated |
| `LayoutSwitcher` | 管理引擎切换 | 暂时保留 |

### 8.2 建议移除的代码 (长期)

| 代码 | 替代方案 | 时机 |
|-----|---------|------|
| `compute.Engine` | `layout.Engine` | 所有组件迁移完成 |
| `computed.ComputedBox` | `layout.LayoutBox + PaintableBox` | 所有组件迁移完成 |
| `LayoutSwitcher` | 直接使用 `layout.Engine` | 不再需要引擎切换 |
| `legacyPaint()` | `fiberFirstPaint()` | 确认 fiber 路径稳定 |

---

## 九、迁移指南

### 9.1 从 Legacy 迁移到 Fiber-first

#### 步骤 1: 更新入口代码

```go
// 旧代码
package main

import (
    "mint/ui"
    "mint/render"
)

func main() {
    app := MyApp()
    ui.Run(app)
}

// 使用 NewDeclarativeNodeFromFunc (旧方式)
func NewMyComponent() component.Node {
    return render.NewDeclarativeNodeFromFunc(func() rtui.VNode {
        // ...
    })
}
```

```go
// 新代码
package main

import (
    "mint/ui"
    "mint/framework"
    "mint/render"
)

func main() {
    os.Setenv("MINT_FIBER_FIRST", "true")  // 可选
    app := MyApp()
    ui.Run(app)
}

// 使用 NewDeclarativeNodeFromFuncWithFiber (新方式)
func NewMyComponent(fwApp *framework.App) component.Node {
    return render.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode {
        // ...
    }, fwApp)
}
```

#### 步骤 2: 确保组件实现接口

```go
type MyComponentInstance struct {
    Props      interface{}
    LayoutFlex layout.FlexStyle
    // ...
}

// 实现 ComponentInstance
func (c *MyComponentInstance) Measure(constraints layout.Constraints) layout.Size {
    // 计算尺寸
    return layout.Size{Width: width, Height: height}
}

func (c *MyComponentInstance) Paint(x, y int) []paint.DrawCmd {
    // 返回绘制命令
    return []paint.DrawCmd{...}
}

func (c *MyComponentInstance) SetBounds(x, y, w, h int) {
    // 设置边界
    c.bounds.X, c.bounds.Y = x, y
    c.bounds.Width, c.bounds.Height = w, h
}
```

#### 步骤 3: 注册组件

```go
runtime ui.RegisterComponent("MyComponent", func() ComponentInstance {
    return &MyComponentInstance{}
})
```

### 9.2 组件迁移检查清单

- [ ] 组件实现 `ComponentInstance` 接口
- [ ] 实现 `Measure()` 方法
- [ ] 实现 `Paint()` 方法返回 `[]paint.DrawCmd`
- [ ] 实现 `SetBounds()` 方法
- [ ] 使用 `NewDeclarativeNodeFromFuncWithFiber` 创建节点
- [ ] 设置 `MINT_FIBER_FIRST=true` 或使用 `ui.RunTest()`
- [ ] 更新组件注册方式

---

## 十、总结

### 10.1 完成的工作

✅ **Phase 1-2: 清理 layer.Manager**
- 移除 `rendering_pipeline.go` 中的 `layerMgr` 字段
- 移除 `pipeline_renderer.go` 中的 `layerMgr` 和 `layerEvents` 字段
- 移除 `declarative_node.go` 中的 `layerMgr` 字段和 `SetLayerManager()` 方法
- 移除 `layer` 包导入

✅ **Phase 3: 清理 Legacy Paint 方法**
- 移除 `paint_engine.go` 中的所有 Legacy Paint 方法
- 移除 `Paint(ComputedLayout)`, `PaintLayers()`, `PaintRenderPlanes()`
- 移除 `parentBackgroundLegacy` 字段
- 移除 `compute` 和 `layer` 包导入
- 保留新 API: `PaintLayout()`, `PaintPaintableLayouts()`, `PaintPaintablePlanes()`

✅ **Phase 4: 新架构支持**
- `fiber_adapter.go` 优先支持 `layout.LayoutBox`
- 保留 `compute.ComputedBox` 作为Legacy回退

✅ **Phase 5: 文档更新**
- 更新 `FIBER_FIRST_RENDER_FLOW.md`
- 更新 `FIBER_RENDER_PATH_ANALYSIS.md`
- 创建本文档 `RENDER_ARCHITECTURE_CLEANUP_SUMMARY.md`

### 10.2 架构状态

```
当前架构状态:
├─ 主渲染路径 (Fiber-first)
│  ├─ VNode → Fiber Reconcile
│  ├─ layout.Engine 布局
│  ├─ PaintableLayout 绘制
│  └─ ✅ 完全实现，测试通过
│
├─ 兼容路径 (Legacy)
│  ├─ VNode 保持到 Paint 阶段
│  ├─ compute.Engine 布局
│  ├─ 旧 Paint API
│  └─ ⚠️ 向后兼容

├─ 兼容性支持
│  └─ fiber_adapter.go: 同时支持新旧架构
│
└─ 文档
   ├─ FIBER_FIRST_RENDER_FLOW.md (已更新)
   ├─ FIBER_RENDER_PATH_ANALYSIS.md (已更新)
   └─ RENDER_ARCHITECTURE_CLEANUP_SUMMARY.md (本文档)
```

### 10.3 编译和测试状态

```
✅ go build ./...  - 编译通过
✅ TestDeclarativeNode_Paint_Fiber - 通过
✅ TestFullRenderingPipeline_NewButton - 通过
✅ TestPaintEngine_PaintLayout - 通过
✅ text_demo - 运行成功
```

### 10.4 下一步

1. 继续检查其他可能的 Legacy 依赖
2. 更新组件文档
3. 运行完整测试套件
4. 逐步迁移所有组件到新架构

---

## 附录: 文件变更列表

### 修改的文件

```
internal/render/paint_engine.go
  - 移除 Legacy Paint 方法 (23处)
  - 移除 parentBackgroundLegacy 字段
  - 移除 compute 和 layer 包导入

internal/render/declarative_node.go
  - 移除 layerMgr 字段
  - 移除 SetLayerManager() 方法
  - 移除 RuntimeGetLayerMgr() 方法

internal/render/pipeline_renderer.go
  - 移除 layerMgr 字段
  - 移除 layerEvents 字段

internal/render/rendering_pipeline.go
  - 移除 layerMgr 字段
  - 移除 GetLayerMgr() 方法
  - 移除 layer 包导入

internal/render/fiber_adapter.go
  - GetPosition() 优先支持 layout.LayoutBox
  - SetPosition() 优先支持 layout.LayoutBox
  - GetSize() 优先支持 layout.LayoutBox
  - SetSize() 优先支持 layout.LayoutBox

docs/fiber/FIBER_FIRST_RENDER_FLOW.md
  - 更新渲染流程图
  - 添加完整接口说明
  - 添加环境变量说明
  - 添加测试验证部分

docs/fiber/FIBER_RENDER_PATH_ANALYSIS.md
  - 更新架构状态
  - 添加清理详情
  - 更新数据流转图
  - 添加迁移指南

docs/fiber/RENDER_ARCHITECTURE_CLEANUP_SUMMARY.md
  - (本文档) 完整总结清理工作
```
