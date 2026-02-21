# Fiber 渲染路径分析报告

## 概述

本文档分析 Mint 框架中 Fiber 的渲染路径，明确最新的渲染流程以及需要清理的遗留代码。

---

## 一、最新渲染路径流程图 (NewDeclarativeNodeFromFuncWithFiber)

### 1.1 入口点

```
ui.Run() / Demo main()
    │
    ▼
NewDeclarativeNodeFromFuncWithFiber(renderFn, fwApp)
    │
    ├── 创建 fiberReconcilerAdapter (reconciler)
    ├── 创建 FiberFocusManager
    ├── 创建 PipelineRendererAdapter
    ├── 调用 reconciler.SetRenderer(renderer)  [Phase 8: NodeID传播]
    └── initFiberFirstPipeline()  [如果 MINT_FIBER_FIRST=true]
```

### 1.2 完整流程图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          DeclarativeNode.Paint()                             │
│                     (declarative_node.go:308-344)                            │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                    ┌───────────────┼───────────────┐
                    │               │               │
            fiberFirstEnabled?  RenderModeBoth?   Default
            (MINT_FIBER_FIRST)                    (Legacy)
                    │               │               │
                    ▼               ▼               ▼
        ┌───────────────┐   ┌───────────────┐   ┌───────────────┐
        │fiberFirstPaint│   │ comparePaint  │   │ legacyPaint   │
        │  (推荐路径)    │   │ (测试对比模式) │   │  (兼容路径)    │
        └───────────────┘   └───────────────┘   └───────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                   fiberFirstPaint() - Fiber-first 三阶段渲染                  │
│                     (declarative_node.go:346-454)                            │
└─────────────────────────────────────────────────────────────────────────────┘
                    │
    ┌───────────────┼───────────────┐
    │               │               │
    ▼               ▼               ▼
┌────────┐    ┌──────────┐    ┌──────────┐
│ Phase 1│    │ Phase 2  │    │ Phase 3  │
│Reconcile│   │ Layout   │    │  Paint   │
└────────┘    └──────────┘    └──────────┘
    │               │               │
    ▼               ▼               ▼

┌─────────────────────────────────────────────────────────────────────────────┐
│ Phase 1: Reconciliation (VNode → Fiber, VNode 丢弃)                          │
├─────────────────────────────────────────────────────────────────────────────┤
│  reconciler.Render()                                                         │
│      │                                                                       │
│      ├── beginWork(): 创建/复用 Fiber 节点                                   │
│      │       └── 从 VNode 提取 Props/Style/Instance                          │
│      │       └── 调用 InstanceFactory 创建持久化组件实例                       │
│      │                                                                       │
│      └── completeWork(): 完成 Fiber 树构建                                    │
│              └── 设置 Fiber.NodeID (稳定运行时标识)                           │
│              └── Fiber.Instance 持有运行时实体                                │
│              └── VNode 在此阶段后不再使用                                     │
│                                                                              │
│  关键文件: internal/reconciler/fiber_sync.go, runtime/ui/fiber.go            │
└─────────────────────────────────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ Phase 2: Fiber-based Layout (Fiber → LayoutBox)                             │
├─────────────────────────────────────────────────────────────────────────────┤
│  newLayoutEngine.LayoutFiber(fiberRoot, constraints)                        │
│      │                                                                       │
│      ├── FiberToNodeAdapterPure 将 Fiber 适配为 layout.Node                 │
│      │       └── 从 Fiber.Instance/Style/Props 获取尺寸                      │
│      │       └── 不访问 VNode (Fiber-first 架构)                             │
│      │                                                                       │
│      └── layout.Engine.Layout()                                             │
│              └── 递归测量和布局                                               │
│              └── 输出 layout.LayoutResult (含 LayoutBox 树和 HitMap)         │
│                                                                              │
│  关键文件: runtime/layout/engine.go, internal/render/fiber_adapter.go       │
└─────────────────────────────────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ Phase 3: Paint (LayoutBox → PaintableLayout → Buffer)                       │
├─────────────────────────────────────────────────────────────────────────────┤
│  FiberToPaintableConverter(fiberRoot).ConvertToLayout(layoutBoxRoot)        │
│      │                                                                       │
│      ├── 从 Fiber.Instance 获取 PaintableInstance                           │
│      │       └── Instance.Paint() 绘制到 PaintableBox                        │
│      │                                                                       │
│      └── PaintEngine.PaintLayout(paintableLayout, buffer)                   │
│              └── 遍历 PaintableBox 树                                        │
│              └── 调用 PaintableBox.Paint() 写入 buffer                       │
│              └── 构建 HitMap 用于事件路由                                     │
│                                                                              │
│  关键文件: runtime/paint/engine.go, internal/render/fiber_to_paintable.go   │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 1.3 关键数据结构

| 数据结构 | 阶段 | 生命周期 | 说明 |
|---------|------|---------|------|
| `VNode` | Reconcile | 临时 | 声明式描述，创建 Fiber 后丢弃 |
| `Fiber` | 全程 | 持久 | 树结构和调度，跨渲染保持 |
| `Fiber.Instance` | 全程 | 持久 | 运行时实体 (Button, Text 等) |
| `layout.LayoutBox` | Layout | 临时 | 布局结果 |
| `paint.PaintableBox` | Paint | 临时 | 绘制数据 |

---

## 二、Legacy VNode 路径流程图

### 2.1 入口点

```
NewDeclarativeNodeFromFunc(fn)  [已弃用]
    │
    ├── 不使用 Fiber reconciler
    ├── renderer = NewPipelineRendererAdapter()
    └── useFiber = false
```

### 2.2 流程图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          legacyPaint()                                       │
│                     (declarative_node.go:524-618)                            │
└─────────────────────────────────────────────────────────────────────────────┘
                    │
    ┌───────────────┼───────────────┐
    │               │               │
    ▼               ▼               ▼
renderWithFiberContext  nonFiberRender  applyFocusState
(有 reconciler)        (无 reconciler)   (焦点状态)
    │               │               │
    └───────────────┴───────────────┘
                    │
                    ▼
        ┌───────────────────────┐
        │   PipelineRenderer    │
        │   .RenderWithConstraints()│
        └───────────────────────┘
                    │
                    ▼
        ┌───────────────────────┐
        │   RenderingPipeline   │
        │   .Render() 或 .RenderLayers() │
        └───────────────────────┘
                    │
    ┌───────────────┼───────────────┐
    │               │               │
    ▼               ▼               ▼
LayoutSwitcher   compute.Engine   renderLegacy
(切换引擎)        (旧布局引擎)      (最终回退)
    │               │
    └───────────────┘
                    │
                    ▼
        ┌───────────────────────┐
        │     PaintEngine       │
        │   .Paint()            │
        └───────────────────────┘
```

---

## 三、遗留代码分析

### 3.1 需要清理的组件

| 组件 | 文件位置 | 状态 | 建议 |
|-----|---------|------|------|
| `LayoutSwitcher` | internal/render/layout_switcher.go | 兼容性 | 保留，但标记为 deprecated |
| `ComputeEngineAdapter` | internal/render/layout_switcher.go | 兼容性 | 保留，仅 legacy 路径使用 |
| `compute.Engine` | runtime/compute/engine.go | 旧引擎 | 保留，仅 legacy 路径使用 |
| `ComputedBox/ComputedLayout` | runtime/compute/*.go | 旧数据结构 | 保留，仅 legacy 路径使用 |
| `renderLegacy` | internal/render/rendering_pipeline.go | 回退 | 保留作为安全回退 |
| `legacyPaint` | internal/render/declarative_node.go | 兼容性 | 保留，标记为 deprecated |

### 3.2 旧路径的使用场景

```
Legacy 路径在以下情况下会被使用:

1. 未设置 MINT_USE_FIBER=true
   └── NewDeclarativeNodeFromFunc() 被调用

2. 未设置 MINT_FIBER_FIRST=true
   └── NewDeclarativeNodeFromFuncWithFiber() 但 fiberFirstEnabled=false

3. fiberFirstPaint 失败回退
   └── 布局或绘制失败时回退到 legacyPaint

4. 测试对比模式 (RenderModeBoth)
   └── comparePaint 同时运行两条路径
```

### 3.3 代码依赖关系

```
                    ┌────────────────────┐
                    │  DeclarativeNode   │
                    └─────────┬──────────┘
                              │
            ┌─────────────────┼─────────────────┐
            │                 │                 │
            ▼                 ▼                 ▼
    ┌───────────────┐ ┌───────────────┐ ┌───────────────┐
    │fiberFirstPaint│ │  legacyPaint  │ │ comparePaint  │
    └───────┬───────┘ └───────┬───────┘ └───────────────┘
            │                 │
            ▼                 ▼
    ┌───────────────┐ ┌───────────────┐
    │NewLayoutEngine│ │LayoutSwitcher │
    │   Adapter     │ │               │
    └───────┬───────┘ └───────┬───────┘
            │                 │
            ▼                 ├─────────────────┐
    ┌───────────────┐         │                 │
    │ layout.Engine │         ▼                 ▼
    │(runtime/layout)│ ┌───────────────┐ ┌───────────────┐
    └───────────────┘ │compute.Engine │ │NewLayoutEngine│
                      │(runtime/compute)│ │   Adapter     │
                      └───────────────┘ └───────────────┘
```

---

## 四、两条保留的渲染路径

### 4.1 路径 A: Fiber-first (推荐)

**启用条件:**
```go
os.Setenv("MINT_USE_FIBER", "true")
os.Setenv("MINT_FIBER_FIRST", "true")

node := render.NewDeclarativeNodeFromFuncWithFiber(renderFn, fwApp)
node.SetRenderMode(render.RenderModeFiberFirst)
```

**流程:**
```
VNode → Fiber Reconcile → Fiber Tree → layout.Engine (Fiber-based) → PaintableLayout → Buffer
        (VNode 丢弃)      (持久化)
```

**优势:**
- 组件实例持久化 (不重复创建)
- Hook 状态正确保持
- 更好的性能 (增量布局/绘制)
- 简化的数据流

### 4.2 路径 B: Legacy VNode (兼容)

**启用条件:**
```go
// 方式 1: 不设置环境变量
node := render.NewDeclarativeNodeFromFunc(renderFn)

// 方式 2: 设置 MINT_USE_FIBER=false
node := render.NewDeclarativeNodeFromFuncWithFiber(renderFn, fwApp)
// 但 fiberFirstEnabled = false
```

**流程:**
```
VNode → PipelineRenderer → LayoutSwitcher → compute.Engine/layout.Engine → PaintEngine → Buffer
        (VNode 保持)        (可切换引擎)
```

**用途:**
- 向后兼容旧代码
- 调试/测试对比
- 安全回退

---

## 五、清理建议

### 5.1 保留的代码

| 代码 | 原因 |
|-----|------|
| `fiberFirstPaint` | 主渲染路径 |
| `legacyPaint` | 兼容性回退 (已标记 deprecated) |
| `LayoutSwitcher` | legacy 路径需要 (已标记 deprecated) |
| `compute.Engine` | legacy 路径需要 |
| `NewDeclarativeNodeFromFunc` | 标记 deprecated，但保留 |
| `RenderModeBoth/comparePaint` | 测试对比 (已标记 deprecated) |

### 5.2 已标记 Deprecated 的代码 (2024-02)

以下代码已添加 `// Deprecated:` 注释，建议迁移到 Fiber-first 架构：

| 文件 | 代码 | 替代方案 |
|-----|------|---------|
| `layout_switcher.go` | `LayoutEngineType` | 不再需要 |
| `layout_switcher.go` | `LayoutSwitcher` | `NewLayoutEngineAdapter` |
| `layout_switcher.go` | `ComputeEngineAdapter` | `NewLayoutEngineAdapter` |
| `layout_switcher.go` | `ParallelRenderingPipeline` | `RenderingPipeline` |
| `declarative_node.go` | `legacyPaint()` | `fiberFirstPaint()` |
| `declarative_node.go` | `comparePaint()` | `fiberFirstPaint()` |
| `declarative_node.go` | `layoutSwitcher` 字段 | `newLayoutEngine` |
| `rendering_pipeline.go` | `switcher` 字段 | 直接使用 `layout.Engine` |
| `fiber_adapter.go` | `VNodeToNodeAdapter` | `FiberToNodeAdapterPure` |
| `fiber_adapter.go` | `FlexLayoutAdapter` | `FiberToNodeAdapterPure` |
| `vnode_renderer.go` | `NonFiberRenderer` | `FiberRenderer` |

### 5.3 可移除的代码 (长期)

| 代码 | 替代方案 | 时机 |
|-----|---------|------|
| `compute.Engine` | `layout.Engine` | 所有组件迁移完成 |
| `ComputedBox` | `LayoutBox + PaintableBox` | 所有组件迁移完成 |
| `LayoutSwitcher` | 直接使用 `layout.Engine` | 不再需要引擎切换 |
| `legacyPaint` | `fiberFirstPaint` | 确认 fiber 路径稳定 |

### 5.4 Deprecated 注释示例

```go
// LayoutSwitcher manages switching between layout engines
//
// Deprecated: Use NewLayoutEngineAdapter directly with Fiber-first architecture.
// The engine switching is no longer needed as Fiber-first always uses runtime/layout.Engine.
//
// Migration: Replace LayoutSwitcher.Layout() with NewLayoutEngineAdapter.LayoutFiber()
type LayoutSwitcher struct { ... }

// legacyPaint is the original VNode-based rendering path
//
// Deprecated: Use fiberFirstPaint with Fiber-first architecture instead.
// This method is kept for backward compatibility but should not be used in new code.
func (n *DeclarativeNode) legacyPaint(ctx component.PaintContext, buf *paint.Buffer) { ... }

// VNodeToNodeAdapter wraps a VNode tree to implement layout.Node interface
//
// Deprecated: Use FiberToNodeAdapterPure with Fiber-first architecture instead.
// In Fiber-first architecture, VNode is discarded after Fiber creation.
type VNodeToNodeAdapter struct { ... }
```

---

## 六、环境变量配置

| 环境变量 | 值 | 效果 |
|---------|---|------|
| `MINT_USE_FIBER` | `true` | 启用 Fiber reconciler |
| `MINT_FIBER_FIRST` | `true` | 启用 Fiber-first 渲染路径 |
| `MINT_LAYOUT_ENGINE` | `compute` | 使用旧布局引擎 (legacy) |
| `MINT_LAYOUT_ENGINE` | `layout` | 使用新布局引擎 (Fiber-first) |
| `MINT_LAYOUT_ENGINE` | `both` | 并行运行两个引擎对比 |
| `MINT_DEBUG_TEST` | `true` | 启用调试输出 |
| `MINT_USE_LEGACY_RENDERER` | `true` | 使用旧渲染器 (调试用) |

---

## 七、迁移指南

### 7.1 从 Legacy 迁移到 Fiber-first

**步骤 1:** 更新入口代码
```go
// 旧代码
node := render.NewDeclarativeNodeFromFunc(app)

// 新代码
fwApp := framework.NewApp()
node := render.NewDeclarativeNodeFromFuncWithFiber(app, fwApp)
node.SetRenderMode(render.RenderModeFiberFirst)
```

**步骤 2:** 确保组件实现正确接口
```go
// 组件需要实现
type ComponentInstance interface {
    GetSize() (int, int)
    Paint(buf *paint.Buffer, x, y int)
    Measure(constraints layout.Constraints) layout.Size
}
```

**步骤 3:** 设置环境变量
```bash
export MINT_USE_FIBER=true
export MINT_FIBER_FIRST=true
```

### 7.2 已迁移的组件

| 组件 | 位置 | 状态 |
|-----|------|------|
| Button | ui/components/button | ✅ 已迁移 |
| Text | runtime/ui | ✅ 已迁移 |
| TextArea | runtime/ui | ✅ 已迁移 |
| Input | runtime/ui | ✅ 已迁移 |
| Checkbox | runtime/ui | ✅ 已迁移 |
| Select | runtime/ui | ✅ 已迁移 |
| Progress | runtime/ui | ✅ 已迁移 |
| Stack (HStack/VStack) | runtime/ui | ✅ 已迁移 |
| Grid | runtime/ui | ✅ 已迁移 |
| Border | runtime/ui | ✅ 已迁移 |
| Absolute | runtime/ui | ✅ 已迁移 |

---

## 八、总结

### 当前架构

```
                    ┌─────────────────────────────────────┐
                    │        DeclarativeNode.Paint()      │
                    └─────────────────┬───────────────────┘
                                      │
                 ┌────────────────────┼────────────────────┐
                 │                    │                    │
                 ▼                    ▼                    ▼
         Fiber-first            Legacy VNode         Compare Mode
         (推荐路径)              (兼容路径)           (测试对比)
                 │                    │                    │
                 ▼                    ▼                    ▼
         layout.Engine          LayoutSwitcher       两者并行
         (Fiber-based)          (compute/layout)
```

### 最终目标架构

```
                    ┌─────────────────────────────────────┐
                    │        DeclarativeNode.Paint()      │
                    └─────────────────┬───────────────────┘
                                      │
                 ┌────────────────────┼────────────────────┐
                 │                    │                    │
                 ▼                    ▼                    ▼
         Fiber-first            Legacy VNode         Compare Mode
         (主路径)               (deprecated)         (可选)
                 │                    │                    │
                 ▼                    ▼                    ▼
         layout.Engine          compute.Engine       两者对比
```

保留两条路径可以确保平滑过渡，同时为未来的完全迁移预留空间。
