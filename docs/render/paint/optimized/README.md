# Fiber-First 优化渲染管线文档

## 文档概述

本目录包含基于 Fiber-first 架构的优化渲染管线设计文档，旨在消除 VNode 运行时依赖，简化渲染流程，提高系统性能。

## 当前系统问题

### 核心问题
1. **双重依赖**：Layout 和 Paint 都依赖 VNode
2. **三层结构混乱**：VNode + Fiber + ComputedBox 职责不清
3. **渲染管线复杂**：VNode 在每次 Paint 时重新创建

### 架构对比

```
旧架构 (3层):
VNode (声明) → Fiber (结构) → ComputedBox (布局结果)
     ↑              ↑              ↑
     └──────────────┴──────────────┘
         三层相互依赖，生命周期混乱

新架构 (2层):
VNode (临时) → Fiber (持久) + paint.PaintableBox (实例)
                    ↓
              LayoutResult
                    ↓
                Buffer
```

## 优化后的架构

### 核心原则

> **Fiber 是唯一的运行时实体**
> **VNode 只在 Reconcile 阶段存在**
> **paint.PaintableBox 是渲染单元**

### 渲染流程

```
用户代码 → renderFn() → VNode (临时)
                          ↓
                   Reconciler (协调)
                          ↓
                   Fiber 树 (持久化)
                    ↓         ↓
              Instance    Instance → paint.PaintableBox
                          ↓
                   Layout Engine
                          ↓
              LayoutResult (paint.PaintableBox + Box)
                          ↓
                   Paint Engine
                          ↓
                   Buffer (屏幕缓冲)
```

## 文档列表

### 1. [FIBER_FIRST_RENDER_PIPELINE.md](./FIBER_FIRST_RENDER_PIPELINE.md)
**核心架构文档**

- 当前系统问题分析
- Fiber-First 优化目标
- 优化后的渲染流程
- 核心架构变更
- 实施方案 (5个Phase)
- 迁移路径
- 性能对比

**适合**: 架构师、技术负责人、需要了解整体设计的人员

### 2. [IMPLEMENTATION_GUIDE.md](./IMPLEMENTATION_GUIDE.md)
**实施指南文档**

- 核心接口定义
- 代码实现示例
- 组件迁移方法
- 测试策略
- 性能监控
- 调试技巧

**适合**: 开发人员、需要具体实施的人员

## 关键优化点

### 1. Fiber 结构变更

```go
// 旧结构
type Fiber struct {
    VNode     VNode           // ❌
    LayoutBox ComputedBox     // ❌
}

// 新结构
type Fiber struct {
    Instance  paint.PaintableBox  // ✅
    Style     Style               // ✅
    Props     MemoizedProps       // ✅
}
```

### 2. Layout 引擎变更

```go
// 旧实现 - 依赖 VNode
func Layout(vnode VNode, fiber *Fiber) ComputedBox

// 新实现 - 纯 Fiber 驱动
func LayoutFiber(root *Fiber, constraints Constraints) *LayoutResult
```

### 3. Paint 引擎变更

```go
// 旧实现 - 依赖 VNode
func Paint(vnode VNode, computedBox ComputedBox, buf *Buffer)

// 新实现 - 纯 paint.PaintableBox 驱动
func PaintLayout(layoutResult *LayoutResult, buf *paint.Buffer)
```


## 多层渲染支持

### Layer 类型

Fiber-First 架构支持 8 种渲染层，按优先级从低到高：

| Layer | 用途 | Z-Index 范围 |
|-------|------|-------------|
| LayerBase | 默认层 | 0-999 |
| LayerDropdown | 下拉菜单 | 1000-1999 |
| LayerSticky | 粘性定位 | 2000-2999 |
| LayerFixed | 固定定位 | 3000-3999 |
| LayerModalBackdrop | 模态背景 | 4000-4999 |
| LayerModal | 模态对话框 | 5000-5999 |
| LayerPopover | 弹出框 | 6000-6999 |
| LayerTooltip | 工具提示 | 7000-7999 |

### 渲染方式

```go
// 单层渲染 (性能更优)
layoutResult := layoutEngine.LayoutFiber(root, constraints)
paintEngine.PaintSingleLayer(layoutResult, buf)

// 多层渲染 (支持 Modal、Dropdown 等)
layeredResult := layoutEngine.LayoutFiberLayered(root, constraints)
paintEngine.PaintMultiLayer(layeredResult, buf)
```

## 接口集成设计

### FiberToNodeAdapter（适配器位置）

> **重要架构约束**：适配器位于 `internal/render/fiber_adapter.go`，而非 `runtime/layout`
> 
> `runtime/layout` 是纯布局引擎，不依赖 Fiber/VNode，只定义抽象接口。

通过适配器模式将 Fiber 与 runtime/layout 解耦：

```
Fiber (internal/render)
    ↓
FiberToNodeAdapter (internal/render/fiber_adapter.go)
    ↓ 实现
layout.Node (runtime/layout 接口)
layout.Layered
layout.Measurable
layout.Dirtyable
```

### 接口映射

| Fiber 字段 | layout 接口 | 用途 |
|-----------|------------|------|
| Style.Layer | GetLayer() | 渲染层 |
| Style.ZIndex | GetZIndex() | 层内排序 |
| Instance.GetSize() | Measure() | 尺寸测量 |
| Flags | IsLayoutDirty() | 增量布局 |

## 性能提升预期

| 指标 | 旧架构 | 新架构 | 提升 |
|------|--------|--------|------|
| 内存占用 | VNode + Fiber + ComputedBox | Fiber + paint.PaintableBox | ~30% ↓ |
| 渲染时间 | VNode 创建 + Layout + Paint | Layout + Paint | ~20% ↓ |
| GC 压力 | 每帧创建 VNode | 仅更新 Fiber | ~40% ↓ |
| 并发能力 | ❌ | ✅ | ∞ |

## 实施计划

### 总时间：3-4周

1. **Phase 1**: Fiber 结构优化 (2-3天)
2. **Phase 2**: Layout 引擎优化 (3-5天)
3. **Phase 3**: Paint 引擎优化 (3-5天)
4. **Phase 4**: 渲染管线集成 (5-7天)
5. **Phase 5**: 组件迁移 (7-10天)

### 里程碑

- **Week 1**: 完成 Phase 1-2，核心架构就位
- **Week 2**: 完成 Phase 3-4，渲染管线可用
- **Week 3**: 完成 Phase 5，所有组件迁移
- **Week 4**: 测试、优化、上线

## 风险控制

### 高风险点
1. 组件迁移工作量
2. 布局兼容性
3. 性能回退

### 缓解措施
1. 渐进式迁移 - 保持双轨运行
2. 充分测试 - 每个阶段完整测试
3. 性能监控 - 实时监控关键指标
4. 回滚机制 - 随时可切换回旧路径

## 成功标准

### 技术标准
- [ ] VNode 在 commit 后被完全丢弃
- [ ] Layout 只读 Fiber
- [ ] Paint 只用 paint.PaintableBox
- [ ] 所有组件实现 paint.PaintableBox
- [ ] 性能提升 > 15%

### 功能标准
- [ ] 所有现有功能正常
- [ ] 所有测试通过
- [ ] 示例应用正常运行
- [ ] 无内存泄漏

## 快速开始

1. **了解架构**: 阅读 [FIBER_FIRST_RENDER_PIPELINE.md](./FIBER_FIRST_RENDER_PIPELINE.md)
2. **开始实施**: 参考 [IMPLEMENTATION_GUIDE.md](./IMPLEMENTATION_GUIDE.md)
3. **迁移组件**: 按照检查清单逐个迁移
4. **测试验证**: 确保功能和性能符合预期

## 相关文档

- [Fiber-First 架构文档](/docs/fiber/fiber_first/consolidated/FIBER_FIRST_ARCHITECTURE.md)
- [Fiber-First 快速参考](/docs/fiber/fiber_first/consolidated/FIBER_FIRST_QUICK_REFERENCE.md)
- [当前渲染流程分析](/docsArchive/declarative_node_paint_analysis.md)

## 维护者

- 架构设计: Fiber-first 架构团队
- 文档更新: 随实施进展持续更新

---

**最后更新**: 2024年
**状态**: 设计完成，准备实施
