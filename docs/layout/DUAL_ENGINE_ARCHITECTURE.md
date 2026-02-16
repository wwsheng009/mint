# Mint 双引擎渲染系统架构文档

## 概述

Mint 框架采用了双布局引擎架构，支持在稳定的 `runtime/compute` 引擎和新的 `runtime/layout` 引擎之间切换。这种设计允许渐进式迁移和功能验证。

## 架构图

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              用户应用层                                          │
│                                                                                 │
│   ui.Run(Hello) ──> framework.App ──> DeclarativeNode                          │
│                                                                   │             │
└───────────────────────────────────────────────────────────────────┼─────────────┘
                                                                    │
                                                                    ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           internal/render                                        │
│                                                                                 │
│   ┌─────────────────────────────────────────────────────────────────────────┐   │
│   │                    PipelineRendererAdapter                               │   │
│   │                                                                         │   │
│   │   ┌─────────────────────────────────────────────────────────────────┐   │   │
│   │   │                    RenderingPipeline                             │   │   │
│   │   │                                                                 │   │   │
│   │   │   ┌─────────────────────────────────────────────────────────┐   │   │   │
│   │   │   │           环境变量: MINT_LAYOUT_ENGINE                  │   │   │   │
│   │   │   │                                                         │   │   │   │
│   │   │   │   MINT_LAYOUT_ENGINE=compute (默认)                    │   │   │   │
│   │   │   │          │                                              │   │   │   │
│   │   │   │          ▼                                              │   │   │   │
│   │   │   │   ┌─────────────────┐                                   │   │   │   │
│   │   │   │   │ compute.Engine  │ ──────────────────────────┐      │   │   │   │
│   │   │   │   │   (稳定引擎)    │                           │      │   │   │   │
│   │   │   │   └─────────────────┘                           │      │   │   │   │
│   │   │   │                                                   │      │   │   │   │
│   │   │   │   MINT_LAYOUT_ENGINE=layout                      │      │   │   │   │
│   │   │   │          │                                       │      │   │   │   │
│   │   │   │          ▼                                       │      │   │   │   │
│   │   │   │   ┌─────────────────┐     ┌─────────────────┐   │      │   │   │   │
│   │   │   │   │ LayoutSwitcher  │────▶│ layout.Engine   │   │      │   │   │   │
│   │   │   │   │   (切换器)      │     │   (新引擎)      │   │      │   │   │   │
│   │   │   │   └─────────────────┘     └─────────────────┘   │      │   │   │   │
│   │   │   │          │                                       │      │   │   │   │
│   │   │   │          │  MINT_LAYOUT_ENGINE=both              │      │   │   │   │
│   │   │   │          │  (并行运行两个引擎比较)               │      │   │   │   │
│   │   │   │          ▼                                       ▼      │   │   │   │
│   │   │   │   ┌─────────────────┐     ┌─────────────────────────┐  │   │   │   │
│   │   │   │   │ ComputeEngine   │     │ NewLayoutEngineAdapter  │  │   │   │   │
│   │   │   │   │    Adapter      │     │        Adapter          │  │   │   │   │
│   │   │   │   └────────┬────────┘     └───────────┬─────────────┘  │   │   │   │
│   │   │   │            │                          │                │   │   │   │
│   │   │   └────────────┼──────────────────────────┼────────────────┘   │   │   │
│   │   │                 │                          │                    │   │   │
│   │   │                 ▼                          ▼                    │   │   │
│   │   │           LayoutResult (统一接口)                              │   │   │
│   │   │                 │                                               │   │   │
│   │   └─────────────────┼───────────────────────────────────────────────┘   │   │
│   │                     │                                                   │   │
│   │                     ▼                                                   │   │
│   │   ┌─────────────────────────────────────────────────────────────────┐   │   │
│   │   │                      PaintEngine                                │   │   │
│   │   │                   (绘制引擎)                                    │   │   │
│   │   └─────────────────────────────────────────────────────────────────┘   │   │
│   │                                                                         │   │
│   └─────────────────────────────────────────────────────────────────────────┘   │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           paint.Buffer                                          │
│                          (屏幕输出)                                              │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## 核心组件

### 1. RenderingPipeline (渲染管道)

位置: `internal/render/rendering_pipeline.go`

职责:
- 协调布局和绘制阶段
- 根据环境变量选择布局引擎
- 管理 HitMap 和 LayerManager

```go
type RenderingPipeline struct {
    layoutEngine *compute.Engine    // 旧引擎 (直接使用)
    switcher     *LayoutSwitcher    // 切换器 (支持新引擎)
    paintEngine  *PaintEngine       // 绘制引擎
    lastHitMap   *event.HitMap      // 事件命中测试
    layerMgr     *layer.Manager     // 图层管理
    useSwitcher  bool               // 是否使用切换器
}
```

### 2. LayoutSwitcher (布局引擎切换器)

位置: `internal/render/layout_switcher.go`

职责:
- 管理两个布局引擎的实例
- 根据配置选择活动引擎
- 支持并行运行两个引擎进行比较

```go
type LayoutSwitcher struct {
    activeType       LayoutEngineType  // 当前活动引擎
    computeEngine    *ComputeEngineAdapter
    newEngine        *NewLayoutEngineAdapter
    compareResults   bool              // 是否比较结果
    tolerancePercent float64           // 容差百分比
}
```

### 3. 引擎适配器

#### ComputeEngineAdapter

位置: `internal/render/layout_switcher.go`

将 `compute.Engine` 适配到 `LayoutEngine` 接口:

```go
type ComputeEngineAdapter struct {
    engine *compute.Engine
}
```

#### NewLayoutEngineAdapter

位置: `internal/render/layout_switcher.go`

将 `layout.Engine` 适配到 `LayoutEngine` 接口:

```go
type NewLayoutEngineAdapter struct {
    engine *layout.Engine
}
```

### 4. FiberToNodeAdapter (Fiber 转换器)

位置: `internal/render/fiber_adapter.go`

将 Fiber 树转换为 `layout.Node` 接口，使新引擎能够处理 Fiber 树:

```go
type FiberToNodeAdapter struct {
    fiber    *reconciler.Fiber
    vnode    rtui.VNode
    children []layout.Node
}
```

## 引擎比较

### compute.Engine (旧引擎)

位置: `runtime/compute/engine.go`

特点:
- ✅ 生产稳定
- ✅ 完整功能支持
- ✅ 与 PaintEngine 紧密集成
- ⚠️ 代码耦合度较高

### layout.Engine (新引擎)

位置: `runtime/layout/types.go`

特点:
- ✅ 模块化设计
- ✅ 循环引用检测
- ✅ 深度限制 (500层)
- ✅ 扩展功能 (Border, Table, Layer, Position)
- ✅ 89.1% 测试覆盖率
- ⚠️ 需要适配器与 PaintEngine 集成

## 使用方式

### 方式一: 环境变量 (推荐)

```bash
# 使用旧引擎 (默认)
MINT_LAYOUT_ENGINE=compute go run ./examples/hello

# 使用新引擎
MINT_LAYOUT_ENGINE=layout go run ./examples/hello

# 并行比较两个引擎
MINT_LAYOUT_ENGINE=both go run ./examples/hello
```

### 方式二: 代码中切换

```go
// 创建 Pipeline
pipeline := render.NewRenderingPipeline()

// 获取切换器并设置引擎类型
switcher := pipeline.GetSwitcher()
if switcher != nil {
    switcher.SetEngineType(render.LayoutEngineNew)
}

// 或使用 ParallelRenderingPipeline
parallelPipeline := render.NewParallelRenderingPipeline()
parallelPipeline.SetLayoutEngineType(render.LayoutEngineNew)
```

## 新引擎功能矩阵

| 功能 | compute.Engine | layout.Engine | 说明 |
|------|---------------|---------------|------|
| **基础布局** ||||
| Flexbox | ✅ | ✅ | 弹性盒子布局 |
| Flex Direction | ✅ | ✅ | Row/Column/RowReverse/ColumnReverse |
| Flex Wrap | ✅ | ✅ | Wrap/NoWrap/WrapReverse |
| **Flex 属性** ||||
| Flex Grow | ✅ | ✅ | 弹性增长 |
| Flex Shrink | ✅ | ✅ | 弹性收缩 |
| Flex Basis | ✅ | ✅ | 基础尺寸 |
| **对齐** ||||
| Main Axis | ✅ | ✅ | Start/End/Center/SpaceBetween/SpaceAround/SpaceEvenly |
| Cross Axis | ✅ | ✅ | Start/End/Center |
| Stretch | ✅ | ✅ | 交叉轴拉伸 |
| Baseline | ❌ | ✅ | 基线对齐 (新引擎独有) |
| **间距** ||||
| Gap | ✅ | ✅ | 子元素间距 |
| Padding | ✅ | ✅ | 内边距 |
| Margin | ✅ | ✅ | 外边距 |
| **扩展功能** ||||
| Border | ❌ | ✅ | 边框容器 |
| Table | ❌ | ✅ | 表格布局 |
| Layer | ❌ | ✅ | 图层系统 (Modal/Tooltip) |
| Position | ✅ | ✅ | 绝对定位 |
| **安全特性** ||||
| 循环检测 | ❌ | ✅ | 防止无限递归 |
| 深度限制 | ❌ | ✅ | 最大 500 层 |
| Nil 处理 | ⚠️ | ✅ | 安全处理 nil 子节点 |
| **性能** ||||
| 缓存 | ✅ | ✅ | 布局结果缓存 |
| 增量更新 | ✅ | ✅ | 脏标记追踪 |

## 渲染流程

```
1. 用户调用 ui.Run(ComponentFunc)
        │
        ▼
2. framework.App 创建 DeclarativeNode
        │
        ▼
3. DeclarativeNode.Paint() 被调用
        │
        ▼
4. PipelineRendererAdapter.Render()
        │
        ▼
5. RenderingPipeline.Render()
   │
   ├─ 检查 MINT_LAYOUT_ENGINE 环境变量
   │
   ├─ if MINT_LAYOUT_ENGINE == "layout":
   │       LayoutSwitcher.Layout() → NewLayoutEngineAdapter.Layout()
   │
   ├─ elif MINT_LAYOUT_ENGINE == "both":
   │       LayoutSwitcher.Layout() → 并行运行两个引擎
   │
   └─ else (默认):
           compute.Engine.Layout()
        │
        ▼
6. PaintEngine.Paint(computedLayout, buffer)
        │
        ▼
7. 输出到 paint.Buffer (屏幕)
```

## 统一接口

### LayoutResult

```go
type LayoutResult interface {
    GetRootBox() PaintableBox     // 获取根布局盒子
    GetHitMap() *event.HitMap     // 获取命中测试地图
    GetRenderPlanes() *layer.RenderPlanes  // 获取渲染平面
}
```

### PaintableBox

```go
type PaintableBox interface {
    GetBounds() (x, y, width, height int)  // 获取边界
    GetChildren() []PaintableBox           // 获取子盒子
}
```

### LayoutEngine

```go
type LayoutEngine interface {
    Layout(vnode rtui.VNode, fiber *reconciler.Fiber, 
           constraints runtime.BoxConstraints) (LayoutResult, error)
    LayoutFiber(fiber *reconciler.Fiber, 
               constraints runtime.BoxConstraints) (LayoutResult, error)
    GetType() LayoutEngineType
    GetStats() CacheStats
    ClearCache()
    SetDebug(debug bool)
}
```

## 文件结构

```
internal/render/
├── rendering_pipeline.go    # 主渲染管道
├── layout_switcher.go       # 引擎切换器
├── fiber_adapter.go         # Fiber 到 layout.Node 适配器
├── paint_engine.go          # 绘制引擎
├── pipeline_renderer.go     # Pipeline 渲染器
├── declarative_node.go      # 声明式节点
└── component.go             # 组件渲染

runtime/layout/
├── types.go                 # 核心类型和引擎
├── flex.go                  # Flexbox 布局
├── constraints.go           # 约束系统
├── cache.go                 # 布局缓存
├── dirty.go                 # 脏标记追踪
├── border.go                # 边框容器
├── table.go                 # 表格布局
├── layer.go                 # 图层系统
├── position.go              # 绝对定位
├── hitmap.go                # 命中测试
├── validator.go             # 布局验证
└── *_test.go                # 测试文件

runtime/compute/
├── engine.go                # 计算引擎
├── box.go                   # 布局盒子
├── constraints.go           # 约束
└── cache.go                 # 缓存
```

## 测试覆盖

| 包 | 覆盖率 |
|---|--------|
| runtime/layout | 89.1% |
| internal/render | 85.3% |

## 最佳实践

### 1. 开发阶段

使用新引擎进行开发和测试:

```bash
export MINT_LAYOUT_ENGINE=layout
go run ./examples/your_app
```

### 2. 生产环境

默认使用稳定引擎:

```bash
# 不设置环境变量，或显式设置
export MINT_LAYOUT_ENGINE=compute
```

### 3. 回归测试

使用并行模式比较两个引擎的输出:

```bash
export MINT_LAYOUT_ENGINE=both
go test ./...
```

### 4. 性能测试

```bash
# 旧引擎
MINT_LAYOUT_ENGINE=compute go test -bench=. ./runtime/layout

# 新引擎
MINT_LAYOUT_ENGINE=layout go test -bench=. ./runtime/layout
```

## 迁移指南

### 从 compute.Engine 迁移到 layout.Engine

1. **接口兼容**: 两个引擎都实现 `LayoutEngine` 接口
2. **结果转换**: `LayoutResult` 接口统一了输出格式
3. **渐进迁移**: 通过环境变量切换，无需修改代码

### 注意事项

1. 新引擎的 Baseline 对齐是独有功能
2. 新引擎的循环检测可能导致不同的布局结果 (循环被截断)
3. 深度限制 (500层) 可能影响极深的组件树

## 未来计划

1. **完全迁移**: 当新引擎足够稳定后，成为默认引擎
2. **性能优化**: 进一步优化缓存策略
3. **功能扩展**: 
   - Grid 布局
   - 更多对齐选项
   - 自适应布局

## 参考文档

- [Flexbox 规范](https://css-tricks.com/snippets/css/a-guide-to-flexbox/)
- [布局系统设计](./LAYOUT_COMPLETION_PLAN.md)
- [性能测试报告](./ENGINE_COMPARISON_REPORT.md)
